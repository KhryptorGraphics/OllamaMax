#!/usr/bin/env node

/**
 * Smart Agents Test Runner
 * Validates the swarm system functionality
 */

const EnhancedSmartAgentsSwarm = require('./swarm-enhanced');
const path = require('path');

class SwarmTestRunner {
  constructor() {
    this.tests = [];
    this.results = [];
  }

  addTest(name, description, testFn) {
    this.tests.push({
      name,
      description,
      testFn,
      status: 'pending'
    });
  }

  async runAllTests() {
    console.log('🧪 Running Smart Agents Swarm Tests...\n');

    for (const test of this.tests) {
      console.log(`🔬 Running: ${test.name}`);
      console.log(`   ${test.description}`);
      
      const startTime = Date.now();
      
      try {
        await test.testFn();
        const duration = Date.now() - startTime;
        
        test.status = 'passed';
        test.duration = duration;
        
        console.log(`   ✅ PASSED (${duration}ms)\n`);
      } catch (error) {
        const duration = Date.now() - startTime;
        
        test.status = 'failed';
        test.error = error.message;
        test.duration = duration;
        
        console.log(`   ❌ FAILED (${duration}ms)`);
        console.log(`   Error: ${error.message}\n`);
      }
      
      this.results.push(test);
    }

    this.printTestSummary();
  }

  printTestSummary() {
    const passed = this.results.filter(t => t.status === 'passed').length;
    const failed = this.results.filter(t => t.status === 'failed').length;
    const total = this.results.length;

    console.log('📊 Test Summary:');
    console.log(`   Total: ${total}`);
    console.log(`   Passed: ${passed} ✅`);
    console.log(`   Failed: ${failed} ${failed > 0 ? '❌' : ''}`);
    console.log(`   Success Rate: ${((passed / total) * 100).toFixed(1)}%`);

    if (failed > 0) {
      console.log('\n❌ Failed Tests:');
      this.results
        .filter(t => t.status === 'failed')
        .forEach(test => {
          console.log(`   - ${test.name}: ${test.error}`);
        });
    }
  }
}

// Test implementations
async function testSwarmInitialization() {
  const swarm = new EnhancedSmartAgentsSwarm();
  await swarm.initializeSwarm();
  
  if (!swarm.neuralLearning) {
    throw new Error('Neural learning system not initialized');
  }
  
  if (!swarm.agentSelector) {
    throw new Error('Agent selector not initialized');
  }
  
  if (swarm.maxAgents !== 25) {
    throw new Error('Max agents not set correctly');
  }
}

async function testTaskAnalysis() {
  const swarm = new EnhancedSmartAgentsSwarm();
  await swarm.initializeSwarm();
  
  const analysis = await swarm.analyzeTask('build a secure API with authentication');
  
  if (!analysis.complexity || analysis.complexity <= 0) {
    throw new Error('Task complexity not calculated');
  }
  
  if (!analysis.requiredSpecializations || analysis.requiredSpecializations.length === 0) {
    throw new Error('Required specializations not determined');
  }
  
  if (!analysis.estimatedAgentCount || analysis.estimatedAgentCount < 1) {
    throw new Error('Agent count not estimated');
  }
}

async function testAgentSelection() {
  const swarm = new EnhancedSmartAgentsSwarm();
  await swarm.initializeSwarm();
  
  const taskAnalysis = {
    complexity: 0.7,
    requiredSpecializations: ['backend-architect', 'security-engineer'],
    estimatedAgentCount: 5,
    taskType: 'development',
    priority: 8
  };
  
  const selectedAgents = swarm.agentSelector.selectAgents(taskAnalysis);
  
  if (!selectedAgents || selectedAgents.length === 0) {
    throw new Error('No agents selected');
  }
  
  if (!selectedAgents.some(agent => agent.specialization === 'general-purpose')) {
    throw new Error('General purpose agent not included');
  }
}

async function testNeuralLearning() {
  const swarm = new EnhancedSmartAgentsSwarm();
  await swarm.initializeSwarm();
  
  // Simulate learning data
  const mockLearningData = {
    agentId: 'test-agent-123',
    specialization: 'backend-architect',
    taskData: { complexity: 0.5, priority: 5 },
    success: true,
    executionTime: 3000,
    output: 'Mock agent output',
    patterns: [{ pattern: 'api-design', confidence: 0.8 }]
  };
  
  await swarm.neuralLearning.processLearningData(mockLearningData);
  
  const report = swarm.neuralLearning.generateLearningReport();
  
  if (!report.summary) {
    throw new Error('Learning report not generated properly');
  }
}

async function testMetricsCollection() {
  const swarm = new EnhancedSmartAgentsSwarm();
  await swarm.initializeSwarm();
  
  // Wait for metrics collection
  await new Promise(resolve => setTimeout(resolve, 1000));
  
  if (!swarm.metrics) {
    throw new Error('Metrics not initialized');
  }
  
  if (typeof swarm.metrics.swarmHealth !== 'number') {
    throw new Error('Swarm health metric not calculated');
  }
}

async function testSPARCIntegration() {
  const swarm = new EnhancedSmartAgentsSwarm();
  await swarm.initializeSwarm();
  
  if (!swarm.sparcIntegration) {
    throw new Error('SPARC integration not initialized');
  }
  
  const sparcPhases = Object.keys(swarm.sparcIntegration.sparcPhases);
  
  if (sparcPhases.length !== 5) {
    throw new Error('SPARC phases not properly configured');
  }
  
  const expectedPhases = ['specification', 'pseudocode', 'architecture', 'refinement', 'completion'];
  for (const phase of expectedPhases) {
    if (!sparcPhases.includes(phase)) {
      throw new Error(`Missing SPARC phase: ${phase}`);
    }
  }
}

async function testConfigurationLoading() {
  const swarm = new EnhancedSmartAgentsSwarm({ maxAgents: 20, minAgents: 5 });
  
  if (swarm.maxAgents !== 20) {
    throw new Error('Custom max agents not set correctly');
  }
  
  if (swarm.minAgents !== 5) {
    throw new Error('Custom min agents not set correctly');
  }
}

async function testComplexityCalculation() {
  const swarm = new EnhancedSmartAgentsSwarm();
  await swarm.initializeSwarm();
  
  const simpleTask = 'fix a typo';
  const complexTask = 'build a distributed microservices architecture with security and performance optimization';
  
  const simpleComplexity = swarm.calculateTaskComplexity(simpleTask);
  const complexComplexity = swarm.calculateTaskComplexity(complexTask);
  
  if (complexComplexity <= simpleComplexity) {
    throw new Error('Complex task not rated higher complexity than simple task');
  }
}

// Main test execution
async function main() {
  const runner = new SwarmTestRunner();

  // Add all tests
  runner.addTest(
    'Swarm Initialization',
    'Test that swarm initializes with all required systems',
    testSwarmInitialization
  );

  runner.addTest(
    'Task Analysis',
    'Test task analysis and complexity calculation',
    testTaskAnalysis
  );

  runner.addTest(
    'Agent Selection',
    'Test agent selection algorithm',
    testAgentSelection
  );

  runner.addTest(
    'Neural Learning',
    'Test neural learning system functionality',
    testNeuralLearning
  );

  runner.addTest(
    'Metrics Collection',
    'Test metrics collection and health assessment',
    testMetricsCollection
  );

  runner.addTest(
    'SPARC Integration',
    'Test SPARC methodology integration',
    testSPARCIntegration
  );

  runner.addTest(
    'Configuration Loading',
    'Test custom configuration loading',
    testConfigurationLoading
  );

  runner.addTest(
    'Complexity Calculation',
    'Test task complexity calculation accuracy',
    testComplexityCalculation
  );

  // Run all tests
  await runner.runAllTests();

  // Exit with appropriate code
  const failed = runner.results.filter(t => t.status === 'failed').length;
  process.exit(failed > 0 ? 1 : 0);
}

if (require.main === module) {
  main().catch(error => {
    console.error('❌ Test runner failed:', error.message);
    process.exit(1);
  });
}

module.exports = SwarmTestRunner;