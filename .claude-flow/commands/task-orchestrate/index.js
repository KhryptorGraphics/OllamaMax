#!/usr/bin/env node

/**
 * Task Orchestrate Command
 * Orchestrates complex tasks across the swarm
 */

const { spawn } = require('child_process');
const fs = require('fs').promises;
const path = require('path');
const { performance } = require('perf_hooks');

// Import task orchestration manager
const TaskOrchestrationManager = require('./task-orchestration-manager');

class TaskOrchestrateCLI {
  constructor() {
    this.orchestrationManager = new TaskOrchestrationManager();
    this.activeTasks = new Map();
  }

  async parseArguments(args) {
    const options = {
      task: null,
      strategy: 'balanced',
      priority: 'medium',
      help: false
    };

    for (let i = 0; i < args.length; i++) {
      const arg = args[i];
      
      switch (arg) {
        case '--task':
          options.task = args[++i];
          break;
        case '--strategy':
          options.strategy = args[++i];
          break;
        case '--priority':
          options.priority = args[++i];
          break;
        case '--help':
        case '-h':
          options.help = true;
          break;
      }
    }

    return options;
  }

  showHelp() {
    console.log(`
🎯 Task Orchestrate - Complex Task Management

Usage:
  task-orchestrate [options]

Options:
  --task <description>     Task description (required)
  --strategy <type>        Orchestration strategy (balanced, parallel, sequential, hierarchical)
  --priority <level>       Task priority (low, medium, high, critical)
  --help, -h              Show this help message

Strategies:
  balanced     - Distribute tasks evenly across available agents
  parallel     - Execute all subtasks simultaneously
  sequential   - Execute subtasks one after another
  hierarchical - Use tree structure with dependencies

Priority Levels:
  low          - Background tasks, non-urgent
  medium       - Standard priority (default)
  high         - Important tasks, faster execution
  critical     - Urgent tasks, maximum resources

Examples:
  # Basic task orchestration
  task-orchestrate --task "Implement user authentication"

  # High priority task with parallel execution
  task-orchestrate --task "Fix production bug" --priority critical --strategy parallel

  # Complex refactoring with hierarchical approach
  task-orchestrate --task "Refactor codebase" --strategy hierarchical --priority high

  # Sequential development workflow
  task-orchestrate --task "Build new feature" --strategy sequential --priority medium

Task Types Supported:
  • Development tasks (coding, testing, debugging)
  • Documentation tasks (writing, updating, reviewing)
  • Analysis tasks (code review, security audit, performance)
  • Maintenance tasks (refactoring, optimization, cleanup)
  • Integration tasks (API development, system integration)
    `);
  }

  async validateTask(task) {
    if (!task || task.trim().length === 0) {
      throw new Error('Task description is required. Use --task "description"');
    }

    if (task.length < 10) {
      throw new Error('Task description too brief. Please provide more details.');
    }

    if (task.length > 500) {
      throw new Error('Task description too long. Please be more concise.');
    }

    return true;
  }

  async validateStrategy(strategy) {
    const validStrategies = ['balanced', 'parallel', 'sequential', 'hierarchical'];
    if (!validStrategies.includes(strategy)) {
      throw new Error(`Invalid strategy: ${strategy}. Valid options: ${validStrategies.join(', ')}`);
    }
    return true;
  }

  async validatePriority(priority) {
    const validPriorities = ['low', 'medium', 'high', 'critical'];
    if (!validPriorities.includes(priority)) {
      throw new Error(`Invalid priority: ${priority}. Valid options: ${validPriorities.join(', ')}`);
    }
    return true;
  }

  async initializeOrchestration(options) {
    console.log('🚀 Initializing Task Orchestration...');
    console.log(`📋 Task: ${options.task}`);
    console.log(`⚙️  Strategy: ${options.strategy}`);
    console.log(`🎯 Priority: ${options.priority}`);

    const taskId = await this.orchestrationManager.createTask({
      description: options.task,
      strategy: options.strategy,
      priority: options.priority
    });

    this.activeTasks.set(taskId, {
      description: options.task,
      startTime: Date.now(),
      options
    });

    return taskId;
  }

  async executeOrchestration(taskId, options) {
    console.log('\n🔄 Executing Task Orchestration...');
    
    try {
      const result = await this.orchestrationManager.executeTask(taskId, {
        description: options.task,
        strategy: options.strategy,
        priority: options.priority
      });

      console.log('\n📊 Orchestration Results:');
      console.log(`✅ Success: ${result.success}`);
      console.log(`🤖 Agents Deployed: ${result.agentsDeployed}`);
      console.log(`📋 Subtasks Completed: ${result.subtasksCompleted}`);
      console.log(`⏱️  Execution Time: ${result.executionTime}ms`);
      console.log(`🎯 Strategy Used: ${result.strategyUsed}`);

      if (result.breakdown) {
        console.log('\n📝 Task Breakdown:');
        result.breakdown.forEach((subtask, index) => {
          console.log(`  ${index + 1}. ${subtask.description} (${subtask.status})`);
        });
      }

      if (result.recommendations) {
        console.log('\n💡 Recommendations:');
        result.recommendations.forEach(rec => {
          console.log(`  • ${rec}`);
        });
      }

      return result;
    } catch (error) {
      console.error('❌ Orchestration execution failed:', error.message);
      throw error;
    }
  }

  async run(args) {
    try {
      const options = await this.parseArguments(args);

      if (options.help) {
        this.showHelp();
        return;
      }

      await this.validateTask(options.task);
      await this.validateStrategy(options.strategy);
      await this.validatePriority(options.priority);
      
      const taskId = await this.initializeOrchestration(options);
      const result = await this.executeOrchestration(taskId, options);

      console.log('\n🎉 Task Orchestration completed successfully!');
      
      if (result.nextActions) {
        console.log('\n🔮 Suggested Next Actions:');
        result.nextActions.forEach(action => {
          console.log(`  • ${action}`);
        });
      }

      // Show performance metrics
      if (result.metrics) {
        console.log('\n📈 Performance Metrics:');
        console.log(`  • Efficiency: ${result.metrics.efficiency}%`);
        console.log(`  • Resource Utilization: ${result.metrics.resourceUtilization}%`);
        console.log(`  • Parallel Execution: ${result.metrics.parallelExecution}%`);
      }

    } catch (error) {
      console.error('❌ Error:', error.message);
      process.exit(1);
    }
  }
}

// CLI execution
if (require.main === module) {
  const cli = new TaskOrchestrateCLI();
  const args = process.argv.slice(2);
  cli.run(args);
}

module.exports = TaskOrchestrateCLI;
