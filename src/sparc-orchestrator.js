#!/usr/bin/env node

/**
 * SPARC Orchestrator - Multi-Agent Task Coordination System
 * 
 * Core capabilities:
 * - Task decomposition into agent-specific work
 * - Parallel and sequential execution patterns
 * - Memory sharing between agents
 * - Progress tracking and monitoring
 * - Result synthesis and aggregation
 */

import { spawn } from 'child_process';
import { promises as fs } from 'fs';
import path from 'path';
import { EventEmitter } from 'events';

class SPARCOrchestrator extends EventEmitter {
  constructor() {
    super();
    this.agents = new Map();
    this.tasks = new Map();
    this.memory = new Map();
    this.results = new Map();
    this.metricsPath = path.join(process.cwd(), '.claude-flow', 'metrics');
  }

  /**
   * Initialize orchestrator environment
   */
  async initialize() {
    // Ensure metrics directory exists
    await fs.mkdir(this.metricsPath, { recursive: true });
    
    // Initialize swarm
    await this.executeCommand('npx claude-flow@alpha swarm init --topology mesh --max-agents 10');
    
    // Set up event handlers
    this.setupEventHandlers();
    
    this.emit('initialized');
    return true;
  }

  /**
   * Decompose task into agent-specific work
   */
  decomposeTask(task, strategy = 'balanced') {
    const subtasks = [];
    
    switch (strategy) {
      case 'domain':
        // Decompose by domain expertise
        subtasks.push(
          { agent: 'researcher', work: `Research requirements for: ${task}` },
          { agent: 'architect', work: `Design architecture for: ${task}` },
          { agent: 'coder', work: `Implement solution for: ${task}` },
          { agent: 'tester', work: `Test implementation of: ${task}` },
          { agent: 'reviewer', work: `Review and optimize: ${task}` }
        );
        break;
        
      case 'parallel':
        // Decompose for parallel execution
        subtasks.push(
          { agent: 'analyzer', work: `Analyze components for: ${task}`, parallel: true },
          { agent: 'designer', work: `Design interfaces for: ${task}`, parallel: true },
          { agent: 'documenter', work: `Document specifications for: ${task}`, parallel: true }
        );
        break;
        
      case 'sequential':
        // Decompose for sequential pipeline
        subtasks.push(
          { agent: 'specification', work: `Define specs for: ${task}`, sequence: 1 },
          { agent: 'pseudocode', work: `Create pseudocode for: ${task}`, sequence: 2 },
          { agent: 'architecture', work: `Design architecture for: ${task}`, sequence: 3 },
          { agent: 'refinement', work: `Refine implementation of: ${task}`, sequence: 4 },
          { agent: 'completion', work: `Complete and integrate: ${task}`, sequence: 5 }
        );
        break;
        
      case 'adaptive':
        // Analyze task complexity and adapt
        const complexity = this.assessComplexity(task);
        if (complexity > 0.7) {
          return this.decomposeTask(task, 'sequential');
        } else if (complexity > 0.4) {
          return this.decomposeTask(task, 'domain');
        } else {
          return this.decomposeTask(task, 'parallel');
        }
        
      default:
        // Balanced distribution
        const agentTypes = ['researcher', 'coder', 'tester', 'reviewer'];
        agentTypes.forEach(agent => {
          subtasks.push({ agent, work: `Handle ${agent} tasks for: ${task}` });
        });
    }
    
    return subtasks;
  }

  /**
   * Spawn agent with specific capabilities
   */
  async spawnAgent(type, capabilities = []) {
    const agentId = `${type}-${Date.now()}`;
    
    // Spawn via claude-flow
    await this.executeCommand(
      `npx claude-flow@alpha agent spawn --type ${type} --name ${agentId}`
    );
    
    const agent = {
      id: agentId,
      type,
      capabilities,
      status: 'idle',
      tasks: [],
      results: []
    };
    
    this.agents.set(agentId, agent);
    this.emit('agent-spawned', agent);
    
    return agent;
  }

  /**
   * Coordinate task execution
   */
  async coordinateTask(task, options = {}) {
    const {
      strategy = 'adaptive',
      parallel = true,
      timeout = 300000,
      retryOnFailure = true
    } = options;
    
    const taskId = `task-${Date.now()}`;
    const subtasks = this.decomposeTask(task, strategy);
    
    // Store task metadata
    this.tasks.set(taskId, {
      id: taskId,
      task,
      subtasks,
      status: 'pending',
      startTime: Date.now(),
      results: []
    });
    
    // Spawn required agents
    const requiredAgents = [...new Set(subtasks.map(st => st.agent))];
    const agents = await Promise.all(
      requiredAgents.map(type => this.spawnAgent(type))
    );
    
    // Execute subtasks
    if (parallel && subtasks.every(st => !st.sequence)) {
      // Parallel execution
      await this.executeParallel(taskId, subtasks, agents);
    } else {
      // Sequential or mixed execution
      await this.executeSequential(taskId, subtasks, agents);
    }
    
    // Aggregate results
    const results = await this.aggregateResults(taskId);
    
    // Update task status
    const taskData = this.tasks.get(taskId);
    taskData.status = 'completed';
    taskData.endTime = Date.now();
    taskData.duration = taskData.endTime - taskData.startTime;
    
    this.emit('task-completed', { taskId, results });
    
    return results;
  }

  /**
   * Execute tasks in parallel
   */
  async executeParallel(taskId, subtasks, agents) {
    const promises = subtasks.map(async (subtask) => {
      const agent = agents.find(a => a.type === subtask.agent);
      return this.executeAgentTask(agent, subtask);
    });
    
    const results = await Promise.all(promises);
    
    // Store results
    const taskData = this.tasks.get(taskId);
    taskData.results = results;
    
    return results;
  }

  /**
   * Execute tasks sequentially
   */
  async executeSequential(taskId, subtasks, agents) {
    const results = [];
    
    // Sort by sequence if defined
    const orderedSubtasks = subtasks.sort((a, b) => 
      (a.sequence || 0) - (b.sequence || 0)
    );
    
    for (const subtask of orderedSubtasks) {
      const agent = agents.find(a => a.type === subtask.agent);
      
      // Pass previous results to next agent via memory
      if (results.length > 0) {
        await this.shareMemory(
          `${taskId}-previous`,
          results[results.length - 1]
        );
      }
      
      const result = await this.executeAgentTask(agent, subtask);
      results.push(result);
    }
    
    // Store results
    const taskData = this.tasks.get(taskId);
    taskData.results = results;
    
    return results;
  }

  /**
   * Execute task with specific agent
   */
  async executeAgentTask(agent, subtask) {
    // Update agent status
    agent.status = 'working';
    agent.tasks.push(subtask);
    
    // Execute via claude-flow
    const command = `npx claude-flow@alpha task execute --agent ${agent.id} --task "${subtask.work}"`;
    const result = await this.executeCommand(command);
    
    // Store result
    agent.results.push(result);
    agent.status = 'idle';
    
    // Share result in memory
    await this.shareMemory(`${agent.id}-result`, result);
    
    this.emit('subtask-completed', { agent, subtask, result });
    
    return result;
  }

  /**
   * Share data in memory
   */
  async shareMemory(key, value) {
    this.memory.set(key, {
      value,
      timestamp: Date.now(),
      ttl: 3600000 // 1 hour TTL
    });
    
    // Also persist to claude-flow memory
    await this.executeCommand(
      `npx claude-flow@alpha memory store --key "${key}" --value '${JSON.stringify(value)}'`
    );
    
    this.emit('memory-shared', { key, value });
  }

  /**
   * Retrieve from memory
   */
  async getMemory(key) {
    const memoryData = this.memory.get(key);
    
    if (memoryData && Date.now() - memoryData.timestamp < memoryData.ttl) {
      return memoryData.value;
    }
    
    // Try claude-flow memory
    const result = await this.executeCommand(
      `npx claude-flow@alpha memory retrieve --key "${key}"`
    );
    
    return result;
  }

  /**
   * Monitor progress
   */
  async monitorProgress() {
    const status = {
      agents: Array.from(this.agents.values()).map(a => ({
        id: a.id,
        type: a.type,
        status: a.status,
        tasksCompleted: a.results.length
      })),
      tasks: Array.from(this.tasks.values()).map(t => ({
        id: t.id,
        status: t.status,
        subtasksTotal: t.subtasks.length,
        subtasksCompleted: t.results?.length || 0,
        duration: t.endTime ? t.endTime - t.startTime : Date.now() - t.startTime
      })),
      memory: {
        entries: this.memory.size,
        keys: Array.from(this.memory.keys())
      }
    };
    
    this.emit('progress-update', status);
    
    return status;
  }

  /**
   * Aggregate results from all agents
   */
  async aggregateResults(taskId) {
    const taskData = this.tasks.get(taskId);
    
    if (!taskData || !taskData.results) {
      return null;
    }
    
    const aggregated = {
      taskId,
      task: taskData.task,
      duration: taskData.duration,
      subtasks: taskData.subtasks.map((st, idx) => ({
        agent: st.agent,
        work: st.work,
        result: taskData.results[idx]
      })),
      summary: this.synthesizeResults(taskData.results),
      metrics: await this.calculateMetrics(taskData)
    };
    
    this.results.set(taskId, aggregated);
    
    return aggregated;
  }

  /**
   * Synthesize results into summary
   */
  synthesizeResults(results) {
    // Combine all results into coherent summary
    const synthesis = {
      successCount: results.filter(r => r?.success).length,
      failureCount: results.filter(r => !r?.success).length,
      insights: [],
      recommendations: [],
      nextSteps: []
    };
    
    // Extract insights from results
    results.forEach(result => {
      if (result?.insights) synthesis.insights.push(...result.insights);
      if (result?.recommendations) synthesis.recommendations.push(...result.recommendations);
      if (result?.nextSteps) synthesis.nextSteps.push(...result.nextSteps);
    });
    
    return synthesis;
  }

  /**
   * Calculate performance metrics
   */
  async calculateMetrics(taskData) {
    const metrics = {
      totalDuration: taskData.duration,
      averageSubtaskDuration: taskData.duration / taskData.subtasks.length,
      parallelizationEfficiency: 0,
      successRate: 0,
      resourceUtilization: 0
    };
    
    if (taskData.results) {
      const successful = taskData.results.filter(r => r?.success).length;
      metrics.successRate = (successful / taskData.results.length) * 100;
    }
    
    // Save metrics
    await this.saveMetrics(metrics);
    
    return metrics;
  }

  /**
   * Save metrics to file
   */
  async saveMetrics(metrics) {
    const metricsFile = path.join(this.metricsPath, 'orchestrator-metrics.json');
    
    try {
      const existing = await fs.readFile(metricsFile, 'utf-8')
        .then(data => JSON.parse(data))
        .catch(() => []);
      
      existing.push({
        timestamp: Date.now(),
        ...metrics
      });
      
      await fs.writeFile(metricsFile, JSON.stringify(existing, null, 2));
    } catch (error) {
      console.error('Failed to save metrics:', error);
    }
  }

  /**
   * Assess task complexity
   */
  assessComplexity(task) {
    // Simple heuristic based on task description
    const complexityFactors = {
      length: task.length > 100 ? 0.2 : 0,
      keywords: 0,
      domains: 0
    };
    
    // Check for complexity keywords
    const complexKeywords = ['system', 'architecture', 'integration', 'optimization', 'refactor'];
    complexKeywords.forEach(keyword => {
      if (task.toLowerCase().includes(keyword)) {
        complexityFactors.keywords += 0.15;
      }
    });
    
    // Check for multiple domains
    const domains = ['frontend', 'backend', 'database', 'api', 'security', 'performance'];
    let domainCount = 0;
    domains.forEach(domain => {
      if (task.toLowerCase().includes(domain)) domainCount++;
    });
    complexityFactors.domains = Math.min(domainCount * 0.1, 0.3);
    
    return Math.min(
      complexityFactors.length + complexityFactors.keywords + complexityFactors.domains,
      1.0
    );
  }

  /**
   * Execute command via child process
   */
  async executeCommand(command) {
    return new Promise((resolve, reject) => {
      const [cmd, ...args] = command.split(' ');
      const process = spawn(cmd, args, { shell: true });
      
      let output = '';
      let error = '';
      
      process.stdout.on('data', (data) => {
        output += data.toString();
      });
      
      process.stderr.on('data', (data) => {
        error += data.toString();
      });
      
      process.on('close', (code) => {
        if (code === 0) {
          resolve({ success: true, output });
        } else {
          resolve({ success: false, error });
        }
      });
      
      process.on('error', (err) => {
        reject(err);
      });
    });
  }

  /**
   * Set up event handlers
   */
  setupEventHandlers() {
    this.on('initialized', () => {
      console.log('✅ SPARC Orchestrator initialized');
    });
    
    this.on('agent-spawned', (agent) => {
      console.log(`🤖 Agent spawned: ${agent.id} (${agent.type})`);
    });
    
    this.on('subtask-completed', ({ agent, subtask }) => {
      console.log(`✓ Subtask completed by ${agent.type}: ${subtask.work.substring(0, 50)}...`);
    });
    
    this.on('task-completed', ({ taskId, results }) => {
      console.log(`🎯 Task ${taskId} completed with ${results.subtasks.length} subtasks`);
    });
    
    this.on('progress-update', (status) => {
      console.log(`📊 Progress: ${status.tasks.length} tasks, ${status.agents.length} agents`);
    });
  }

  /**
   * Cleanup resources
   */
  async cleanup() {
    // Clear memory
    this.memory.clear();
    
    // Terminate agents
    for (const agent of this.agents.values()) {
      await this.executeCommand(`npx claude-flow@alpha agent terminate --id ${agent.id}`);
    }
    
    this.agents.clear();
    this.tasks.clear();
    this.results.clear();
    
    this.emit('cleanup-complete');
  }
}

// Export for use as module
export default SPARCOrchestrator;

// CLI interface
if (import.meta.url === `file://${process.argv[1]}`) {
  const orchestrator = new SPARCOrchestrator();
  
  async function main() {
    const task = process.argv[2] || 'coordinate feature development';
    const strategy = process.argv[3] || 'adaptive';
    
    console.log('🚀 SPARC Orchestrator Starting...');
    console.log(`📋 Task: ${task}`);
    console.log(`🎯 Strategy: ${strategy}`);
    
    try {
      await orchestrator.initialize();
      
      const results = await orchestrator.coordinateTask(task, {
        strategy,
        parallel: true
      });
      
      console.log('\n📊 Results:');
      console.log(JSON.stringify(results, null, 2));
      
      // Monitor progress
      const status = await orchestrator.monitorProgress();
      console.log('\n📈 Final Status:');
      console.log(JSON.stringify(status, null, 2));
      
    } catch (error) {
      console.error('❌ Error:', error);
    } finally {
      await orchestrator.cleanup();
    }
  }
  
  main();
}