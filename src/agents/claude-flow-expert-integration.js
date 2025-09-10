#!/usr/bin/env node

/**
 * Claude Flow Expert Integration Layer
 * Seamlessly integrates expert agents with npx claude-flow@alpha operations
 */

import { exec, spawn } from 'child_process';
import { promisify } from 'util';
import fs from 'fs/promises';
import path from 'path';
import { EXPERT_AGENT_REGISTRY, getExpertAgent, getExpertsForCommand } from './expert-agent-registry.js';
import { ClaudeFlowBridge } from './claude-flow-bridge.js';

const execAsync = promisify(exec);

/**
 * Claude Flow Expert Command Router
 * Maps claude-flow commands to expert agents automatically
 */
export class ClaudeFlowExpertRouter {
  constructor(options = {}) {
    this.bridge = new ClaudeFlowBridge(options);
    this.commandMap = this.buildCommandMap();
    this.activeExperts = new Map();
    this.memoryPath = path.join(process.cwd(), '.claude-flow', 'memory');
    this.metricsPath = path.join(process.cwd(), '.claude-flow', 'metrics');
  }

  /**
   * Build command to expert mapping
   */
  buildCommandMap() {
    return {
      // SPARC Commands
      'sparc run architect': ['sparc-architect'],
      'sparc run analyzer': ['sparc-analyzer'],
      'sparc run coder': ['sparc-coder'],
      'sparc run debugger': ['sparc-debugger'],
      'sparc run optimizer': ['sparc-optimizer'],
      'sparc run documenter': ['sparc-documenter'],
      'sparc run tester': ['sparc-tester'],
      'sparc run reviewer': ['sparc-reviewer'],
      'sparc run orchestrator': ['sparc-orchestrator'],
      'sparc pipeline': ['sparc-orchestrator', 'sparc-coder', 'sparc-tester'],
      'sparc tdd': ['sparc-tester', 'sparc-coder', 'sparc-reviewer'],
      'sparc concurrent': ['sparc-orchestrator', 'smart-agent'],
      
      // GitHub Commands
      'github pr-manager': ['github-pr-analyzer'],
      'github issue-tracker': ['github-issue-orchestrator'],
      'github workflow-automation': ['github-actions-architect'],
      'github release-manager': ['github-release-coordinator'],
      'github security-scan': ['github-security-auditor'],
      'github swarm-pr': ['github-pr-analyzer', 'sparc-reviewer'],
      'github swarm-issue': ['github-issue-orchestrator', 'sparc-orchestrator'],
      'github code-review': ['sparc-reviewer', 'github-pr-analyzer'],
      
      // Swarm Commands
      'swarm init --topology hierarchical': ['swarm-queen'],
      'swarm init --topology mesh': ['swarm-mesh-node'],
      'swarm adapt': ['swarm-adaptive-controller'],
      'swarm scale': ['swarm-queen', 'swarm-adaptive-controller'],
      
      // Analysis Commands
      'analysis bottleneck-detect': ['performance-profiler', 'sparc-analyzer'],
      'analysis performance-report': ['performance-profiler', 'cache-optimizer'],
      'analysis security-audit': ['security-penetration-tester', 'security-cryptographer'],
      'analysis vulnerability-scan': ['github-security-auditor', 'security-penetration-tester'],
      
      // Neural Commands
      'neural train': ['neural-architect', 'ml-feature-engineer'],
      'neural patterns': ['neural-architect', 'sparc-analyzer'],
      'neural benchmark': ['performance-profiler', 'neural-architect'],
      
      // Infrastructure Commands
      'k8s deploy': ['kubernetes-orchestrator'],
      'terraform plan': ['terraform-infrastructure-coder'],
      'terraform apply': ['terraform-infrastructure-coder'],
      'infra provision': ['terraform-infrastructure-coder', 'kubernetes-orchestrator'],
      
      // Data Commands
      'data pipeline-create': ['data-pipeline-architect'],
      'data warehouse-design': ['data-warehouse-specialist'],
      'data quality-check': ['data-pipeline-architect', 'sparc-tester'],
      
      // Frontend Commands
      'react optimize': ['react-optimization-expert'],
      'css architect': ['css-architecture-specialist'],
      'frontend performance': ['react-optimization-expert', 'performance-profiler'],
      'design-system create': ['css-architecture-specialist', 'sparc-documenter'],
      
      // Backend Commands
      'graphql schema-design': ['graphql-api-architect'],
      'microservices design': ['microservices-architect'],
      'api create --type graphql': ['graphql-api-architect'],
      'service-mesh deploy': ['microservices-architect', 'kubernetes-orchestrator'],
      
      // Blockchain Commands
      'blockchain audit': ['smart-contract-auditor'],
      'defi design-protocol': ['defi-protocol-architect'],
      'smart-contract security': ['smart-contract-auditor', 'security-cryptographer'],
      
      // Mobile Commands
      'mobile optimize': ['react-native-performance'],
      'flutter architect': ['flutter-architect'],
      'react-native profile': ['react-native-performance', 'performance-profiler'],
      
      // Testing Commands
      'test e2e-create': ['e2e-automation-architect'],
      'test performance': ['performance-test-engineer'],
      'test visual-regression': ['e2e-automation-architect'],
      'benchmark load': ['performance-test-engineer', 'performance-profiler'],
      
      // Compliance Commands
      'compliance audit': ['fintech-compliance-expert'],
      'healthcare hipaa-audit': ['healthcare-hipaa-specialist'],
      'fintech kyc-implement': ['fintech-compliance-expert', 'security-cryptographer']
    };
  }

  /**
   * Initialize expert router
   */
  async initialize() {
    await this.bridge.initialize();
    await this.loadExpertMemory();
    await this.setupMetricsCollection();
    console.log('✅ Claude Flow Expert Router initialized');
  }

  /**
   * Load expert memory patterns
   */
  async loadExpertMemory() {
    try {
      await fs.mkdir(this.memoryPath, { recursive: true });
      const memoryFile = path.join(this.memoryPath, 'expert-patterns.json');
      
      if (await this.fileExists(memoryFile)) {
        const data = await fs.readFile(memoryFile, 'utf-8');
        const patterns = JSON.parse(data);
        console.log(`🧠 Loaded ${Object.keys(patterns).length} expert patterns`);
      }
    } catch (error) {
      console.warn('Creating new expert memory:', error.message);
    }
  }

  /**
   * Setup metrics collection
   */
  async setupMetricsCollection() {
    try {
      await fs.mkdir(this.metricsPath, { recursive: true });
    } catch (error) {
      console.warn('Metrics setup warning:', error.message);
    }
  }

  /**
   * Route claude-flow command to expert agents
   */
  async routeCommand(command, args = {}) {
    // Find matching experts for command
    const expertTypes = this.commandMap[command] || this.findBestExperts(command);
    
    if (expertTypes.length === 0) {
      console.warn(`No experts found for command: ${command}`);
      return this.fallbackToGeneralAgent(command, args);
    }

    // Spawn expert agents
    const experts = await this.spawnExperts(expertTypes, command, args);
    
    // Coordinate expert execution
    return await this.coordinateExperts(experts, command, args);
  }

  /**
   * Find best experts for unmatched command
   */
  findBestExperts(command) {
    const experts = getExpertsForCommand(command);
    return experts.slice(0, 3).map(e => e.type);
  }

  /**
   * Spawn expert agents
   */
  async spawnExperts(expertTypes, command, args) {
    const experts = [];
    
    for (const expertType of expertTypes) {
      const expert = getExpertAgent(expertType);
      
      if (expert) {
        const spawnedExpert = await this.spawnExpert(expert, expertType, command, args);
        experts.push(spawnedExpert);
      }
    }
    
    return experts;
  }

  /**
   * Spawn individual expert
   */
  async spawnExpert(expert, expertType, command, args) {
    const taskDescription = this.buildExpertTask(expert, command, args);
    
    // Execute pre-hook if defined
    if (expert.hooks?.pre) {
      await this.executeHook(expert.hooks.pre);
    }

    // Store expert state
    this.activeExperts.set(expertType, {
      expert,
      command,
      startTime: Date.now(),
      status: 'active'
    });

    // Create task configuration
    const taskConfig = {
      name: expert.name,
      description: taskDescription,
      type: expert.type,
      tools: expert.tools,
      specializations: expert.specializations,
      metadata: {
        expertType,
        command,
        args
      }
    };

    return {
      expertType,
      expert,
      taskConfig,
      promise: this.executeExpertTask(taskConfig, expert)
    };
  }

  /**
   * Build expert-specific task description
   */
  buildExpertTask(expert, command, args) {
    const { task = '', options = {} } = args;
    
    return `
[${expert.name}]
Command: ${command}
Task: ${task}

As a ${expert.description}, execute this task using your specialized capabilities:
${expert.capabilities.join(', ')}

Available Tools: ${expert.tools.join(', ')}

Specializations to apply:
${JSON.stringify(expert.specializations, null, 2)}

Claude Flow Integration:
${expert.specializations?.claudeFlow?.join('\n') || 'Standard execution'}

Priority: ${options.priority || 'medium'}
Mode: Expert Execution
`.trim();
  }

  /**
   * Execute expert task
   */
  async executeExpertTask(taskConfig, expert) {
    try {
      // Use bridge to spawn agent
      const result = await this.bridge.spawnAgent(
        taskConfig.metadata.expertType,
        taskConfig.description,
        {
          ...taskConfig.metadata.args,
          expertMode: true
        }
      );

      // Execute post-hook if defined
      if (expert.hooks?.post) {
        await this.executeHook(expert.hooks.post);
      }

      // Store results in memory
      await this.storeExpertResults(taskConfig.metadata.expertType, result);

      return result;
    } catch (error) {
      console.error(`Expert execution failed for ${taskConfig.name}:`, error);
      throw error;
    }
  }

  /**
   * Coordinate multiple experts
   */
  async coordinateExperts(experts, command, args) {
    const coordination = {
      command,
      experts: experts.map(e => e.expertType),
      startTime: Date.now(),
      results: []
    };

    try {
      // Execute experts (parallel or sequential based on command)
      const isParallel = this.shouldRunParallel(command);
      
      if (isParallel) {
        // Parallel execution
        const results = await Promise.all(experts.map(e => e.promise));
        coordination.results = results;
      } else {
        // Sequential execution with data passing
        for (const expert of experts) {
          const result = await expert.promise;
          coordination.results.push(result);
          
          // Pass results to next expert via memory
          if (experts.indexOf(expert) < experts.length - 1) {
            await this.passDataToNextExpert(result, experts[experts.indexOf(expert) + 1]);
          }
        }
      }

      coordination.duration = Date.now() - coordination.startTime;
      coordination.status = 'success';
      
      // Store coordination results
      await this.storeCoordinationResults(coordination);
      
      return coordination;
    } catch (error) {
      coordination.status = 'failed';
      coordination.error = error.message;
      throw error;
    } finally {
      // Cleanup active experts
      experts.forEach(e => this.activeExperts.delete(e.expertType));
    }
  }

  /**
   * Determine if command should run experts in parallel
   */
  shouldRunParallel(command) {
    const parallelCommands = [
      'sparc concurrent',
      'github swarm-pr',
      'github swarm-issue',
      'analysis',
      'test'
    ];
    
    return parallelCommands.some(cmd => command.includes(cmd));
  }

  /**
   * Pass data between sequential experts
   */
  async passDataToNextExpert(result, nextExpert) {
    try {
      const memoryKey = `expert-handoff/${nextExpert.expertType}`;
      const memoryValue = {
        fromExpert: result.expertType,
        data: result,
        timestamp: new Date().toISOString()
      };
      
      await execAsync(
        `npx claude-flow@alpha memory store --key "${memoryKey}" --value '${JSON.stringify(memoryValue)}'`
      );
    } catch (error) {
      console.warn('Failed to pass data between experts:', error.message);
    }
  }

  /**
   * Execute hook command
   */
  async executeHook(hookCommand) {
    try {
      await execAsync(hookCommand);
    } catch (error) {
      console.warn('Hook execution warning:', error.message);
    }
  }

  /**
   * Store expert results in memory
   */
  async storeExpertResults(expertType, result) {
    try {
      const memoryFile = path.join(this.memoryPath, `${expertType}-results.json`);
      const existing = await this.loadJsonFile(memoryFile) || [];
      
      existing.push({
        timestamp: new Date().toISOString(),
        result
      });
      
      // Keep only last 100 results
      const toStore = existing.slice(-100);
      await fs.writeFile(memoryFile, JSON.stringify(toStore, null, 2));
    } catch (error) {
      console.warn('Failed to store expert results:', error.message);
    }
  }

  /**
   * Store coordination results
   */
  async storeCoordinationResults(coordination) {
    try {
      const metricsFile = path.join(this.metricsPath, 'expert-coordination.json');
      const existing = await this.loadJsonFile(metricsFile) || [];
      
      existing.push({
        timestamp: new Date().toISOString(),
        ...coordination
      });
      
      // Keep only last 1000 coordinations
      const toStore = existing.slice(-1000);
      await fs.writeFile(metricsFile, JSON.stringify(toStore, null, 2));
    } catch (error) {
      console.warn('Failed to store coordination results:', error.message);
    }
  }

  /**
   * Fallback to general agent
   */
  async fallbackToGeneralAgent(command, args) {
    console.log('Falling back to general-purpose agent');
    return await this.bridge.spawnAgent('general-purpose', `Execute: ${command}`, args);
  }

  /**
   * Load JSON file safely
   */
  async loadJsonFile(filePath) {
    try {
      const data = await fs.readFile(filePath, 'utf-8');
      return JSON.parse(data);
    } catch {
      return null;
    }
  }

  /**
   * Check if file exists
   */
  async fileExists(filePath) {
    try {
      await fs.access(filePath);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get expert router status
   */
  async getStatus() {
    return {
      activeExperts: Array.from(this.activeExperts.entries()).map(([type, info]) => ({
        type,
        status: info.status,
        command: info.command,
        duration: Date.now() - info.startTime
      })),
      availableExperts: Object.keys(EXPERT_AGENT_REGISTRY).length,
      commandMappings: Object.keys(this.commandMap).length,
      bridgeStatus: await this.bridge.getStatus()
    };
  }

  /**
   * Shutdown router
   */
  async shutdown() {
    await this.bridge.shutdown();
    this.activeExperts.clear();
    console.log('✅ Expert router shutdown complete');
  }
}

/**
 * Create expert router instance
 */
export function createExpertRouter(options = {}) {
  return new ClaudeFlowExpertRouter(options);
}

/**
 * Execute claude-flow command with expert routing
 */
export async function executeWithExperts(command, args = {}) {
  const router = createExpertRouter();
  await router.initialize();
  
  try {
    const result = await router.routeCommand(command, args);
    return result;
  } finally {
    await router.shutdown();
  }
}

/**
 * CLI interface for testing
 */
if (import.meta.url === `file://${process.argv[1]}`) {
  const args = process.argv.slice(2);
  
  if (args.length === 0) {
    console.log(`
Claude Flow Expert Integration

Usage:
  node claude-flow-expert-integration.js <command> [args]

Examples:
  node claude-flow-expert-integration.js "sparc run architect" "design microservices"
  node claude-flow-expert-integration.js "github pr-manager" "review PR #123"
  node claude-flow-expert-integration.js "analysis security-audit" "scan for vulnerabilities"
    `);
    process.exit(0);
  }

  const command = args[0];
  const taskArgs = {
    task: args.slice(1).join(' '),
    options: {}
  };

  executeWithExperts(command, taskArgs)
    .then(result => {
      console.log('✅ Command executed successfully');
      console.log(JSON.stringify(result, null, 2));
    })
    .catch(error => {
      console.error('❌ Command failed:', error.message);
      process.exit(1);
    });
}