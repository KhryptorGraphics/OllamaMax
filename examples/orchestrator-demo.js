#!/usr/bin/env node

/**
 * SPARC Orchestrator Demo
 * Demonstrates various orchestration patterns and capabilities
 */

import SPARCOrchestrator from '../src/sparc-orchestrator.js';
import { OrchestrationPatterns, selectPattern } from '../src/orchestration-patterns.js';

// Color output for better visibility
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  dim: '\x1b[2m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m'
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function logSection(title) {
  console.log('\n' + '='.repeat(60));
  log(`  ${title}`, 'bright');
  console.log('='.repeat(60) + '\n');
}

// Demo scenarios
const demos = {
  /**
   * Demo 1: Simple Feature Development
   */
  async simpleFeature(orchestrator) {
    logSection('Demo 1: Simple Feature Development');
    
    const task = 'Build a user registration form with email validation';
    log(`Task: ${task}`, 'cyan');
    
    // Use domain strategy for feature development
    const results = await orchestrator.coordinateTask(task, {
      strategy: 'domain',
      parallel: false
    });
    
    log('\n✅ Feature Development Complete!', 'green');
    log(`Duration: ${results.duration}ms`, 'dim');
    log(`Subtasks completed: ${results.subtasks.length}`, 'dim');
    
    // Show insights
    if (results.summary?.insights?.length > 0) {
      log('\n📊 Insights:', 'yellow');
      results.summary.insights.forEach(insight => 
        console.log(`  • ${insight}`)
      );
    }
    
    return results;
  },

  /**
   * Demo 2: Parallel Analysis
   */
  async parallelAnalysis(orchestrator) {
    logSection('Demo 2: Parallel System Analysis');
    
    const task = 'Analyze system performance, security vulnerabilities, and code quality';
    log(`Task: ${task}`, 'cyan');
    
    // Use parallel strategy for independent analyses
    const results = await orchestrator.coordinateTask(task, {
      strategy: 'parallel',
      parallel: true
    });
    
    log('\n✅ Analysis Complete!', 'green');
    log(`Parallel execution saved time`, 'green');
    log(`Duration: ${results.duration}ms`, 'dim');
    
    // Monitor progress
    const status = await orchestrator.monitorProgress();
    log(`\n📈 Status:`, 'yellow');
    log(`  Agents used: ${status.agents.length}`, 'dim');
    log(`  Memory entries: ${status.memory.entries}`, 'dim');
    
    return results;
  },

  /**
   * Demo 3: SPARC Sequential Pipeline
   */
  async sparcPipeline(orchestrator) {
    logSection('Demo 3: SPARC Methodology Pipeline');
    
    const task = 'Implement binary search algorithm using SPARC methodology';
    log(`Task: ${task}`, 'cyan');
    
    // Use sequential strategy for SPARC pipeline
    const results = await orchestrator.coordinateTask(task, {
      strategy: 'sequential',
      parallel: false
    });
    
    log('\n✅ SPARC Pipeline Complete!', 'green');
    
    // Show each phase result
    const phases = ['Specification', 'Pseudocode', 'Architecture', 'Refinement', 'Completion'];
    results.subtasks.forEach((subtask, idx) => {
      log(`  ${phases[idx]}: ${subtask.result?.success ? '✓' : '✗'}`, 
          subtask.result?.success ? 'green' : 'red');
    });
    
    return results;
  },

  /**
   * Demo 4: Adaptive Strategy
   */
  async adaptiveStrategy(orchestrator) {
    logSection('Demo 4: Adaptive Strategy Selection');
    
    const tasks = [
      'Fix typo in README',
      'Add logging to user service',
      'Refactor entire authentication system for better security and performance'
    ];
    
    for (const task of tasks) {
      const complexity = orchestrator.assessComplexity(task);
      log(`\nTask: ${task}`, 'cyan');
      log(`Complexity: ${(complexity * 100).toFixed(0)}%`, 'yellow');
      
      const results = await orchestrator.coordinateTask(task, {
        strategy: 'adaptive'
      });
      
      log(`Strategy selected: ${complexity < 0.3 ? 'parallel' : complexity < 0.7 ? 'domain' : 'sequential'}`, 'magenta');
      log(`Agents used: ${results.subtasks.length}`, 'dim');
    }
  },

  /**
   * Demo 5: Consensus Building
   */
  async consensusBuilding(orchestrator) {
    logSection('Demo 5: Consensus Building for Code Review');
    
    const task = 'Review pull request for security, performance, and code quality';
    log(`Task: ${task}`, 'cyan');
    
    // Use consensus pattern
    const pattern = OrchestrationPatterns.consensus;
    const results = await pattern.execute(task, orchestrator);
    
    if (results.consensus) {
      log('\n✅ Consensus Achieved!', 'green');
      log(`Approval after ${results.rounds} round(s)`, 'dim');
    } else {
      log('\n❌ No Consensus', 'red');
      log(`Failed after ${results.rounds} rounds`, 'dim');
    }
    
    return results;
  },

  /**
   * Demo 6: Event-Driven Workflow
   */
  async eventDriven(orchestrator) {
    logSection('Demo 6: Event-Driven Development Workflow');
    
    const task = 'Develop feature with automatic testing and deployment';
    log(`Task: ${task}`, 'cyan');
    
    // Track events
    const events = [];
    
    // Custom event handler
    orchestrator.on('subtask-completed', ({ agent, subtask }) => {
      events.push(`${agent.type} completed`);
      log(`  🔔 Event: ${agent.type} completed`, 'yellow');
    });
    
    // Use event-driven pattern
    const pattern = OrchestrationPatterns.eventDriven;
    const results = await pattern.execute(task, orchestrator);
    
    log('\n✅ Event-Driven Workflow Complete!', 'green');
    log(`Total events processed: ${events.length}`, 'dim');
    
    return results;
  },

  /**
   * Demo 7: MapReduce for Data Processing
   */
  async mapReduce(orchestrator) {
    logSection('Demo 7: MapReduce Pattern for Large Data');
    
    const task = 'Process and analyze user activity logs for patterns and anomalies';
    log(`Task: ${task}`, 'cyan');
    
    // Use MapReduce pattern
    const pattern = OrchestrationPatterns.mapReduce;
    const results = await pattern.execute(task, orchestrator);
    
    log('\n✅ MapReduce Complete!', 'green');
    log(`Map phase: 5 parallel mappers`, 'dim');
    log(`Reduce phase: 2 reducers`, 'dim');
    log(`Total processing time: ${Date.now()}ms`, 'dim');
    
    return results;
  },

  /**
   * Demo 8: Memory Sharing
   */
  async memorySharing(orchestrator) {
    logSection('Demo 8: Inter-Agent Memory Sharing');
    
    // Share context between agents
    await orchestrator.shareMemory('project-context', {
      framework: 'React',
      database: 'PostgreSQL',
      authentication: 'JWT'
    });
    
    log('Shared project context in memory', 'green');
    
    const task = 'Build API endpoint using shared context';
    log(`Task: ${task}`, 'cyan');
    
    // Agents will use shared memory
    const results = await orchestrator.coordinateTask(task, {
      strategy: 'domain'
    });
    
    // Retrieve shared results
    const sharedResults = await orchestrator.getMemory('researcher-result');
    log('\n📝 Shared Memory Contents:', 'yellow');
    if (sharedResults) {
      console.log(JSON.stringify(sharedResults, null, 2));
    }
    
    return results;
  },

  /**
   * Demo 9: Progress Monitoring
   */
  async progressMonitoring(orchestrator) {
    logSection('Demo 9: Real-time Progress Monitoring');
    
    const task = 'Perform comprehensive system audit and optimization';
    log(`Task: ${task}`, 'cyan');
    
    // Start long-running task
    const taskPromise = orchestrator.coordinateTask(task, {
      strategy: 'sequential',
      timeout: 30000
    });
    
    // Monitor progress every 500ms
    const interval = setInterval(async () => {
      const status = await orchestrator.monitorProgress();
      
      // Clear line and update
      process.stdout.write('\r' + ' '.repeat(80) + '\r');
      process.stdout.write(
        `Progress: ${status.agents.filter(a => a.status === 'working').length} agents working, ` +
        `${status.tasks[0]?.subtasksCompleted || 0}/${status.tasks[0]?.subtasksTotal || 0} subtasks`
      );
    }, 500);
    
    const results = await taskPromise;
    clearInterval(interval);
    
    log('\n\n✅ Monitoring Complete!', 'green');
    log(`Final metrics:`, 'yellow');
    console.log(JSON.stringify(results.metrics, null, 2));
    
    return results;
  },

  /**
   * Demo 10: Pattern Auto-Selection
   */
  async patternAutoSelection(orchestrator) {
    logSection('Demo 10: Intelligent Pattern Selection');
    
    const tasks = [
      'Review and approve deployment to production',
      'Process 1TB of log files for analysis',
      'Handle user registration when form is submitted',
      'Optimize database queries across all services',
      'Build responsive dashboard with real-time updates'
    ];
    
    for (const task of tasks) {
      log(`\nTask: ${task}`, 'cyan');
      
      // Auto-select pattern
      const pattern = selectPattern(task);
      const patternName = Object.keys(OrchestrationPatterns)
        .find(key => OrchestrationPatterns[key] === pattern);
      
      log(`Selected Pattern: ${patternName}`, 'magenta');
      
      // Show pattern structure
      if (pattern.structure) {
        log('Pattern Structure:', 'yellow');
        console.log(JSON.stringify(pattern.structure, null, 2));
      }
    }
  }
};

// Main demo runner
async function runDemo() {
  const orchestrator = new SPARCOrchestrator();
  
  try {
    // Initialize
    log('\n🚀 Initializing SPARC Orchestrator...', 'bright');
    await orchestrator.initialize();
    log('✅ Orchestrator initialized\n', 'green');
    
    // Select demo
    const demoName = process.argv[2] || 'all';
    
    if (demoName === 'all') {
      // Run all demos
      for (const [name, demo] of Object.entries(demos)) {
        await demo(orchestrator);
        await new Promise(resolve => setTimeout(resolve, 1000)); // Pause between demos
      }
    } else if (demos[demoName]) {
      // Run specific demo
      await demos[demoName](orchestrator);
    } else {
      // List available demos
      log('Available demos:', 'yellow');
      Object.keys(demos).forEach(name => {
        console.log(`  • ${name}`);
      });
      log('\nUsage: node orchestrator-demo.js [demo-name|all]', 'dim');
    }
    
  } catch (error) {
    log(`\n❌ Error: ${error.message}`, 'red');
    console.error(error);
  } finally {
    // Cleanup
    log('\n🧹 Cleaning up...', 'dim');
    await orchestrator.cleanup();
    log('✅ Cleanup complete', 'green');
  }
}

// Run if executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  runDemo();
}

export { demos, runDemo };