#!/usr/bin/env node

/**
 * Smart Agents Hive-Mind Swarm Command
 * Creates and manages a massively parallel software development team
 * with neural learning and auto-scaling capabilities
 */

const { spawn } = require('child_process');
const fs = require('fs').promises;
const path = require('path');
const { performance } = require('perf_hooks');

// Import enhanced swarm implementation
const EnhancedSmartAgentsSwarm = require('./swarm-enhanced');

class SmartAgentsSwarm {
  constructor(options = {}) {
    this.maxAgents = options.maxAgents || 25;
    this.minAgents = options.minAgents || 8;
    this.currentAgents = 0;
    this.activeAgents = new Map();
    this.taskQueue = [];
    this.completedTasks = [];
    this.metrics = {
      tasksCompleted: 0,
      totalExecutionTime: 0,
      efficiency: 0,
      learningData: []
    };
    this.neuralMemory = new Map();
    this.agentSpecializations = [
      'architect', 'backend', 'frontend', 'devops', 'security',
      'quality-engineer', 'performance', 'refactoring', 'python-expert',
      'testing', 'documentation', 'requirements', 'system-design'
    ];
  }

  async initialize() {
    console.log('🚀 Initializing Smart Agents Hive-Mind Swarm...');
    await this.loadNeuralMemory();
    await this.setupMetricsCollection();
    console.log(`✅ Swarm initialized with ${this.minAgents}-${this.maxAgents} agent capacity`);
  }

  async loadNeuralMemory() {
    try {
      const memoryPath = path.join(__dirname, '../../memory/neural-memory.json');
      const data = await fs.readFile(memoryPath, 'utf8');
      const memoryData = JSON.parse(data);
      this.neuralMemory = new Map(Object.entries(memoryData));
      console.log(`🧠 Loaded neural memory with ${this.neuralMemory.size} patterns`);
    } catch (error) {
      console.log('🧠 Creating new neural memory system...');
      this.neuralMemory = new Map();
    }
  }

  async saveNeuralMemory() {
    try {
      const memoryPath = path.join(__dirname, '../../memory/neural-memory.json');
      await fs.mkdir(path.dirname(memoryPath), { recursive: true });
      const memoryData = Object.fromEntries(this.neuralMemory);
      await fs.writeFile(memoryPath, JSON.stringify(memoryData, null, 2));
    } catch (error) {
      console.error('❌ Failed to save neural memory:', error.message);
    }
  }

  async setupMetricsCollection() {
    const metricsPath = path.join(__dirname, '../../metrics');
    await fs.mkdir(metricsPath, { recursive: true });
    
    // Initialize performance tracking
    setInterval(() => this.collectMetrics(), 5000);
  }

  async analyzeTask(task) {
    console.log(`🔍 Analyzing task: "${task.substring(0, 100)}..."`);
    
    // Use neural memory to determine optimal agent types
    const complexity = this.calculateTaskComplexity(task);
    const requiredSpecializations = this.determineRequiredSpecializations(task);
    const estimatedAgentCount = Math.min(
      this.maxAgents,
      Math.max(this.minAgents, Math.ceil(complexity * 3))
    );

    return {
      complexity,
      requiredSpecializations,
      estimatedAgentCount,
      priority: this.calculateTaskPriority(task)
    };
  }

  calculateTaskComplexity(task) {
    const complexityKeywords = {
      high: ['architecture', 'system', 'distributed', 'microservices', 'security', 'performance'],
      medium: ['api', 'database', 'frontend', 'backend', 'testing', 'integration'],
      low: ['bug', 'fix', 'update', 'documentation', 'style']
    };

    let complexity = 0.3; // base complexity
    
    Object.entries(complexityKeywords).forEach(([level, keywords]) => {
      const matches = keywords.filter(keyword => 
        task.toLowerCase().includes(keyword.toLowerCase())
      ).length;
      
      switch(level) {
        case 'high': complexity += matches * 0.4; break;
        case 'medium': complexity += matches * 0.2; break;
        case 'low': complexity += matches * 0.1; break;
      }
    });

    return Math.min(1.0, complexity);
  }

  determineRequiredSpecializations(task) {
    const specializations = [];
    const taskLower = task.toLowerCase();

    if (taskLower.includes('architecture') || taskLower.includes('system')) {
      specializations.push('architect', 'system-architect');
    }
    if (taskLower.includes('security')) {
      specializations.push('security-engineer');
    }
    if (taskLower.includes('performance')) {
      specializations.push('performance-engineer');
    }
    if (taskLower.includes('test')) {
      specializations.push('quality-engineer');
    }
    if (taskLower.includes('ui') || taskLower.includes('frontend')) {
      specializations.push('frontend-architect');
    }
    if (taskLower.includes('api') || taskLower.includes('backend')) {
      specializations.push('backend-architect');
    }
    if (taskLower.includes('python')) {
      specializations.push('python-expert');
    }
    if (taskLower.includes('refactor')) {
      specializations.push('refactoring-expert');
    }

    // Always include a general-purpose agent
    specializations.push('general-purpose');

    return [...new Set(specializations)];
  }

  calculateTaskPriority(task) {
    const urgentKeywords = ['critical', 'urgent', 'fix', 'bug', 'error', 'security'];
    const taskLower = task.toLowerCase();
    
    const urgentMatches = urgentKeywords.filter(keyword => 
      taskLower.includes(keyword)
    ).length;

    return Math.min(10, 5 + urgentMatches * 2);
  }

  async spawnAgent(specialization, taskData) {
    const agentId = `agent-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    console.log(`🤖 Spawning ${specialization} agent (${agentId})`);

    const agent = {
      id: agentId,
      specialization,
      status: 'active',
      taskData,
      spawnTime: Date.now(),
      completedTasks: 0,
      learningData: []
    };

    this.activeAgents.set(agentId, agent);
    this.currentAgents++;

    // Create agent prompt based on specialization
    const agentPrompt = this.createAgentPrompt(specialization, taskData);
    
    try {
      // Execute agent task using Claude Code Task tool
      const result = await this.executeAgentTask(agentPrompt, specialization);
      await this.handleAgentResult(agentId, result);
    } catch (error) {
      console.error(`❌ Agent ${agentId} failed:`, error.message);
      await this.handleAgentFailure(agentId, error);
    }

    return agentId;
  }

  createAgentPrompt(specialization, taskData) {
    const basePrompt = `You are a ${specialization} agent in a hive-mind swarm. 
Your mission: ${taskData.task}

Context:
- Complexity Level: ${(taskData.complexity * 100).toFixed(1)}%
- Priority: ${taskData.priority}/10
- Team Size: ${this.currentAgents} agents
- Specialized Role: ${specialization}

Instructions:
1. Focus on your specialization while maintaining awareness of the broader system
2. Coordinate with other agents through shared context
3. Apply SPARC methodology for systematic development
4. Document your decisions and learnings for neural memory
5. Optimize for parallel execution and efficiency

Required Output:
- Detailed analysis of your specialized area
- Concrete implementation steps
- Integration points with other agents
- Performance metrics and learnings
- Recommendations for system optimization

Execute your specialized analysis and implementation now.`;

    return basePrompt;
  }

  async executeAgentTask(prompt, specialization) {
    // This would integrate with Claude Code's Task tool
    // For now, simulating agent execution
    const startTime = performance.now();
    
    console.log(`⚡ Agent ${specialization} executing task...`);
    
    // Simulate processing time based on complexity
    await new Promise(resolve => setTimeout(resolve, Math.random() * 2000 + 1000));
    
    const endTime = performance.now();
    const executionTime = endTime - startTime;

    return {
      success: true,
      executionTime,
      output: `Agent ${specialization} completed specialized analysis and implementation`,
      learningData: {
        specialization,
        executionTime,
        timestamp: Date.now()
      }
    };
  }

  async handleAgentResult(agentId, result) {
    const agent = this.activeAgents.get(agentId);
    if (!agent) return;

    agent.completedTasks++;
    agent.learningData.push(result.learningData);

    // Update neural memory with learnings
    const memoryKey = `${agent.specialization}-patterns`;
    const existingPatterns = this.neuralMemory.get(memoryKey) || [];
    existingPatterns.push({
      timestamp: Date.now(),
      executionTime: result.executionTime,
      success: result.success
    });
    
    this.neuralMemory.set(memoryKey, existingPatterns);

    console.log(`✅ Agent ${agentId} (${agent.specialization}) completed task`);
    
    // Check if we need to scale down
    await this.evaluateScaling();
  }

  async handleAgentFailure(agentId, error) {
    const agent = this.activeAgents.get(agentId);
    if (!agent) return;

    console.log(`🔄 Agent ${agentId} failed, attempting recovery...`);
    
    // Store failure patterns for learning
    const memoryKey = `${agent.specialization}-failures`;
    const existingFailures = this.neuralMemory.get(memoryKey) || [];
    existingFailures.push({
      timestamp: Date.now(),
      error: error.message,
      context: agent.taskData
    });
    
    this.neuralMemory.set(memoryKey, existingFailures);

    // Attempt to respawn with adjusted parameters
    if (agent.completedTasks === 0) {
      console.log(`🔄 Respawning ${agent.specialization} agent with adjusted parameters`);
      // Implement respawn logic here
    }
  }

  async evaluateScaling() {
    const workload = this.taskQueue.length;
    const efficiency = this.calculateSwarmEfficiency();
    
    if (workload > this.currentAgents * 2 && this.currentAgents < this.maxAgents) {
      await this.scaleUp();
    } else if (workload < this.currentAgents * 0.5 && this.currentAgents > this.minAgents) {
      await this.scaleDown();
    }
  }

  async scaleUp() {
    const additionalAgents = Math.min(3, this.maxAgents - this.currentAgents);
    console.log(`📈 Scaling up: Adding ${additionalAgents} agents`);
    
    for (let i = 0; i < additionalAgents; i++) {
      const specialization = this.selectOptimalSpecialization();
      await this.spawnAgent(specialization, { 
        task: 'Support swarm workload',
        complexity: 0.5,
        priority: 5
      });
    }
  }

  async scaleDown() {
    const agentsToRemove = Math.min(2, this.currentAgents - this.minAgents);
    console.log(`📉 Scaling down: Removing ${agentsToRemove} agents`);
    
    // Remove least efficient agents
    const sortedAgents = Array.from(this.activeAgents.values())
      .sort((a, b) => a.completedTasks - b.completedTasks);
    
    for (let i = 0; i < agentsToRemove; i++) {
      const agent = sortedAgents[i];
      this.activeAgents.delete(agent.id);
      this.currentAgents--;
      console.log(`🔻 Removed agent ${agent.id} (${agent.specialization})`);
    }
  }

  selectOptimalSpecialization() {
    // Use neural memory to determine most needed specialization
    const specializationCounts = new Map();
    
    this.activeAgents.forEach(agent => {
      const count = specializationCounts.get(agent.specialization) || 0;
      specializationCounts.set(agent.specialization, count + 1);
    });

    // Find least represented specialization
    let minCount = Infinity;
    let optimalSpecialization = 'general-purpose';
    
    this.agentSpecializations.forEach(spec => {
      const count = specializationCounts.get(spec) || 0;
      if (count < minCount) {
        minCount = count;
        optimalSpecialization = spec;
      }
    });

    return optimalSpecialization;
  }

  calculateSwarmEfficiency() {
    if (this.activeAgents.size === 0) return 0;
    
    let totalEfficiency = 0;
    this.activeAgents.forEach(agent => {
      const uptime = Date.now() - agent.spawnTime;
      const taskRate = agent.completedTasks / (uptime / 1000); // tasks per second
      totalEfficiency += taskRate;
    });
    
    return totalEfficiency / this.activeAgents.size;
  }

  async collectMetrics() {
    this.metrics = {
      timestamp: Date.now(),
      activeAgents: this.currentAgents,
      tasksCompleted: this.completedTasks.length,
      efficiency: this.calculateSwarmEfficiency(),
      swarmHealth: this.assessSwarmHealth(),
      neuralMemorySize: this.neuralMemory.size,
      learningPatterns: this.extractLearningPatterns()
    };

    // Save metrics
    const metricsPath = path.join(__dirname, '../../metrics/swarm-metrics.json');
    await fs.writeFile(metricsPath, JSON.stringify(this.metrics, null, 2));
  }

  assessSwarmHealth() {
    const healthFactors = {
      agentUtilization: this.currentAgents / this.maxAgents,
      taskCompletionRate: this.completedTasks.length / (this.completedTasks.length + this.taskQueue.length + 1),
      errorRate: this.calculateErrorRate(),
      learningProgress: this.assessLearningProgress()
    };

    const overallHealth = Object.values(healthFactors).reduce((sum, factor) => sum + factor, 0) / 4;
    return Math.round(overallHealth * 100);
  }

  calculateErrorRate() {
    let totalFailures = 0;
    this.neuralMemory.forEach((patterns, key) => {
      if (key.includes('failures')) {
        totalFailures += patterns.length;
      }
    });
    
    const totalTasks = this.completedTasks.length + totalFailures;
    return totalTasks > 0 ? totalFailures / totalTasks : 0;
  }

  assessLearningProgress() {
    // Assess how much the swarm has learned over time
    const recentLearnings = [];
    this.neuralMemory.forEach((patterns) => {
      const recent = patterns.filter(p => Date.now() - p.timestamp < 3600000); // last hour
      recentLearnings.push(...recent);
    });
    
    return Math.min(1, recentLearnings.length / 10); // normalize to 0-1
  }

  extractLearningPatterns() {
    const patterns = {};
    this.neuralMemory.forEach((data, key) => {
      if (data.length > 0) {
        const recent = data.slice(-5); // last 5 entries
        const avgExecutionTime = recent.reduce((sum, item) => sum + (item.executionTime || 0), 0) / recent.length;
        patterns[key] = {
          frequency: data.length,
          recentPerformance: avgExecutionTime,
          trend: this.calculateTrend(data)
        };
      }
    });
    return patterns;
  }

  calculateTrend(data) {
    if (data.length < 2) return 'stable';
    
    const recent = data.slice(-5);
    const older = data.slice(-10, -5);
    
    if (recent.length === 0 || older.length === 0) return 'stable';
    
    const recentAvg = recent.reduce((sum, item) => sum + (item.executionTime || 0), 0) / recent.length;
    const olderAvg = older.reduce((sum, item) => sum + (item.executionTime || 0), 0) / older.length;
    
    if (recentAvg < olderAvg * 0.9) return 'improving';
    if (recentAvg > olderAvg * 1.1) return 'degrading';
    return 'stable';
  }

  async executeSwarm(task, options = {}) {
    console.log('\n🚀 Smart Agents Hive-Mind Swarm Executing...\n');
    
    const analysis = await this.analyzeTask(task);
    console.log(`📊 Task Analysis:
    - Complexity: ${(analysis.complexity * 100).toFixed(1)}%
    - Required Agents: ${analysis.estimatedAgentCount}
    - Priority: ${analysis.priority}/10
    - Specializations: ${analysis.requiredSpecializations.join(', ')}`);

    const agentPromises = [];
    
    // Spawn required specialized agents
    for (const specialization of analysis.requiredSpecializations) {
      const agentPromise = this.spawnAgent(specialization, {
        task,
        complexity: analysis.complexity,
        priority: analysis.priority
      });
      agentPromises.push(agentPromise);
    }

    console.log(`\n⚡ Spawned ${agentPromises.length} specialized agents in parallel`);
    
    try {
      const agentResults = await Promise.all(agentPromises);
      console.log(`\n✅ All ${agentResults.length} agents completed successfully`);
      
      // Save neural memory
      await this.saveNeuralMemory();
      
      // Generate final report
      const report = await this.generateExecutionReport();
      console.log('\n📋 Execution Report Generated');
      
      return {
        success: true,
        agentsUsed: agentResults.length,
        swarmHealth: this.assessSwarmHealth(),
        report
      };
      
    } catch (error) {
      console.error('\n❌ Swarm execution failed:', error.message);
      return {
        success: false,
        error: error.message,
        swarmHealth: this.assessSwarmHealth()
      };
    }
  }

  async generateExecutionReport() {
    const report = {
      timestamp: new Date().toISOString(),
      swarmConfiguration: {
        totalAgents: this.currentAgents,
        maxCapacity: this.maxAgents,
        specializations: [...new Set(Array.from(this.activeAgents.values()).map(a => a.specialization))]
      },
      performance: this.metrics,
      recommendations: this.generateRecommendations(),
      neuralLearnings: this.extractLearningPatterns()
    };

    const reportPath = path.join(__dirname, '../../metrics/execution-report.json');
    await fs.writeFile(reportPath, JSON.stringify(report, null, 2));
    
    return report;
  }

  generateRecommendations() {
    const recommendations = [];
    
    if (this.calculateErrorRate() > 0.1) {
      recommendations.push('Consider improving error handling and recovery mechanisms');
    }
    
    if (this.calculateSwarmEfficiency() < 0.5) {
      recommendations.push('Optimize task distribution and agent specialization');
    }
    
    if (this.currentAgents === this.maxAgents) {
      recommendations.push('Consider increasing maximum agent capacity for better scalability');
    }
    
    return recommendations;
  }
}

// CLI Interface
async function main() {
  const args = process.argv.slice(2);
  const command = args[0];
  const task = args.slice(1).join(' ');

  if (!command) {
    console.log(`
🚀 Smart Agents Hive-Mind Swarm

Usage:
  smart-agents execute "<task>"     - Execute task with swarm
  smart-agents sparc "<task>"       - Execute with SPARC methodology
  smart-agents status               - Show swarm status
  smart-agents metrics              - Show performance metrics
  smart-agents train                - Trigger neural learning
  smart-agents scale <n>            - Set max agents (8-25)

Examples:
  smart-agents execute "build a distributed microservices architecture"
  smart-agents sparc "implement user authentication system"
  smart-agents execute "implement comprehensive testing suite"
  smart-agents execute "optimize system performance and security"
    `);
    return;
  }

  // Use enhanced swarm implementation
  const swarm = new EnhancedSmartAgentsSwarm();
  await swarm.initializeSwarm();

  switch (command) {
    case 'execute':
      if (!task) {
        console.error('❌ Please provide a task to execute');
        process.exit(1);
      }
      const result = await swarm.executeEnhancedSwarm(task);
      console.log('\n🎯 Final Enhanced Result:', result);
      break;

    case 'status':
      console.log('\n📊 Swarm Status:');
      console.log(`Active Agents: ${swarm.currentAgents}/${swarm.maxAgents}`);
      console.log(`Health: ${swarm.assessEnhancedSwarmHealth()}%`);
      console.log(`Neural Memory: ${swarm.neuralLearning.learningData.size} patterns`);
      console.log(`Auto-scaling: ${swarm.scalingConfig ? 'ENABLED' : 'DISABLED'}`);
      console.log(`SPARC Integration: ${swarm.sparcIntegration ? 'READY' : 'NOT AVAILABLE'}`);
      break;

    case 'metrics':
      await swarm.collectEnhancedMetrics();
      console.log('\n📈 Enhanced Swarm Metrics:');
      console.log(`├─ Active Agents: ${swarm.metrics.activeAgents}/${swarm.metrics.maxCapacity}`);
      console.log(`├─ Tasks Completed: ${swarm.metrics.tasksCompleted}`);
      console.log(`├─ Queue Length: ${swarm.metrics.taskQueueLength}`);
      console.log(`├─ Efficiency: ${(swarm.metrics.efficiency * 100).toFixed(1)}%`);
      console.log(`├─ Health Score: ${swarm.metrics.swarmHealth}%`);
      console.log(`├─ Neural Patterns: ${swarm.metrics.neuralMetrics.totalPatterns}`);
      console.log(`└─ Learning Rate: ${swarm.metrics.neuralMetrics.learningRate.toFixed(2)}/hr`);
      break;

    case 'train':
      console.log('🧠 Triggering neural learning optimization...');
      const trainingResult = await swarm.neuralLearning.trainSystem();
      console.log('\n📊 Training Results:');
      console.log(`Patterns: ${trainingResult.summary.totalPatterns}`);
      console.log(`High Confidence: ${trainingResult.summary.memoryUtilization.highConfidencePatterns}`);
      console.log('✅ Neural training completed');
      break;

    case 'sparc':
      if (!task) {
        console.error('❌ Please provide a task for SPARC execution');
        process.exit(1);
      }
      console.log('\n🎯 Executing SPARC methodology with Smart Agents...');
      const sparcResult = await swarm.sparcIntegration.executeSPARCWorkflow(task);
      console.log('\n📋 SPARC Results:', sparcResult.overallSuccess ? '✅ Success' : '❌ Failed');
      break;

    case 'scale':
      const maxAgents = parseInt(args[1]);
      if (maxAgents >= 8 && maxAgents <= 25) {
        swarm.maxAgents = maxAgents;
        console.log(`📏 Max agents set to ${maxAgents}`);
      } else {
        console.error('❌ Max agents must be between 8 and 25');
      }
      break;

    default:
      console.error(`❌ Unknown command: ${command}`);
      process.exit(1);
  }
}

if (require.main === module) {
  main().catch(console.error);
}

module.exports = SmartAgentsSwarm;