#!/usr/bin/env node

/**
 * Validation script for Claude Flow Integration
 * Performs comprehensive validation of all implemented features
 */

import { createIntegration, getAvailableAgents } from '../src/agents/claude-flow-integration.js';
import fs from 'fs/promises';
import path from 'path';

const GREEN = '\x1b[32m';
const RED = '\x1b[31m';
const YELLOW = '\x1b[33m';
const BLUE = '\x1b[34m';
const RESET = '\x1b[0m';

class ValidationRunner {
  constructor() {
    this.integration = null;
    this.results = {
      passed: 0,
      failed: 0,
      warnings: 0,
      tests: []
    };
  }

  log(color, message) {
    console.log(`${color}${message}${RESET}`);
  }

  async test(name, testFn) {
    try {
      this.log(BLUE, `🧪 Testing: ${name}`);
      await testFn();
      this.log(GREEN, `✅ PASS: ${name}`);
      this.results.passed++;
      this.results.tests.push({ name, status: 'PASS' });
    } catch (error) {
      this.log(RED, `❌ FAIL: ${name} - ${error.message}`);
      this.results.failed++;
      this.results.tests.push({ name, status: 'FAIL', error: error.message });
    }
  }

  async warn(name, testFn) {
    try {
      await testFn();
    } catch (error) {
      this.log(YELLOW, `⚠️  WARN: ${name} - ${error.message}`);
      this.results.warnings++;
      this.results.tests.push({ name, status: 'WARN', error: error.message });
    }
  }

  async run() {
    this.log(BLUE, '🚀 Starting Claude Flow Integration Validation');
    console.log();

    try {
      // Initialize integration
      this.integration = createIntegration({
        autoSwarm: false,
        memorySharing: true,
        hooks: false, // Disable for testing
        topology: 'mesh',
        maxAgents: 3
      });

      await this.runCoreTests();
      await this.runAgentTests();
      await this.runOrchestrationTests();
      await this.runMemoryTests();
      await this.runUtilityTests();
      await this.runErrorHandlingTests();

    } finally {
      // Cleanup
      if (this.integration?.initialized) {
        await this.integration.shutdown();
      }
    }

    this.printSummary();
    return this.results.failed === 0;
  }

  async runCoreTests() {
    this.log(BLUE, '\n📋 Core Functionality Tests');
    
    await this.test('Integration Initialization', async () => {
      await this.integration.initialize();
      if (!this.integration.initialized) {
        throw new Error('Integration failed to initialize');
      }
    });

    await this.test('Configuration Validation', async () => {
      const config = this.integration.config;
      if (!config.topology || !config.maxAgents) {
        throw new Error('Invalid configuration');
      }
    });

    await this.test('Status Retrieval', async () => {
      const status = await this.integration.getStatus();
      if (!status.initialized || !status.config) {
        throw new Error('Invalid status response');
      }
    });

    await this.test('Health Check', async () => {
      const health = await this.integration.healthCheck();
      if (!health.status || !health.checks) {
        throw new Error('Invalid health check response');
      }
    });
  }

  async runAgentTests() {
    this.log(BLUE, '\n🤖 Agent Management Tests');

    await this.test('Available Agents List', async () => {
      const agents = getAvailableAgents();
      if (!Array.isArray(agents) || agents.length === 0) {
        throw new Error('No agents available');
      }
      
      // Verify we have key agent types
      const types = agents.map(a => a.type);
      const requiredTypes = ['coder', 'tester', 'reviewer', 'planner', 'researcher'];
      
      for (const type of requiredTypes) {
        if (!types.includes(type)) {
          throw new Error(`Missing required agent type: ${type}`);
        }
      }
    });

    await this.test('Single Agent Spawn', async () => {
      const result = await this.integration.spawnAgent(
        'coder',
        'Create a simple test function',
        { timeout: 30000 }
      );
      
      if (!result || !result.agentId || !result.agentType || result.success !== true) {
        throw new Error('Agent spawn failed or returned invalid result');
      }
    });

    await this.test('Agent Status Tracking', async () => {
      const spawnResult = await this.integration.spawnAgent('reviewer', 'Review test code');
      const status = await this.integration.getAgentStatus('reviewer', spawnResult.agentId);
      
      if (!status.agentId || status.agentType !== 'reviewer') {
        throw new Error('Agent status tracking failed');
      }
    });

    await this.test('Active Agents Listing', async () => {
      const agents = await this.integration.listActiveAgents();
      if (!agents.agents || !Array.isArray(agents.agents)) {
        throw new Error('Active agents listing failed');
      }
    });

    await this.warn('Agent Spawn with Invalid Type', async () => {
      try {
        await this.integration.spawnAgent('nonexistent-agent', 'test');
        throw new Error('Should have failed with invalid agent type');
      } catch (error) {
        if (!error.message.includes('Unknown agent type')) {
          throw error;
        }
        // Expected error
      }
    });
  }

  async runOrchestrationTests() {
    this.log(BLUE, '\n🎯 Orchestration Tests');

    await this.test('Parallel Orchestration', async () => {
      const result = await this.integration.orchestrateTask(
        'Build a simple web application',
        { strategy: 'parallel', maxAgents: 3 }
      );
      
      if (!result.id || result.status !== 'completed' || !Array.isArray(result.agents)) {
        throw new Error('Parallel orchestration failed');
      }
    });

    await this.test('Sequential Orchestration', async () => {
      const result = await this.integration.orchestrateTask(
        'Design then implement then test',
        { strategy: 'sequential', maxAgents: 3 }
      );
      
      if (result.status !== 'completed' || !result.results) {
        throw new Error('Sequential orchestration failed');
      }
    });

    await this.test('Task Analysis', async () => {
      const requirements = this.integration.analyzeTaskRequirements('implement secure authentication API');
      if (!Array.isArray(requirements) || requirements.length === 0) {
        throw new Error('Task analysis failed');
      }
    });

    await this.test('Agent Selection', async () => {
      const agents = this.integration.selectOptimalAgents(['code-generation', 'unit-testing'], { maxAgents: 2 });
      if (!Array.isArray(agents) || agents.length === 0) {
        throw new Error('Agent selection failed');
      }
    });

    await this.test('Result Synthesis', async () => {
      const mockResults = [
        { agent: { type: 'coder' }, success: true, result: 'code created' },
        { agent: { type: 'tester' }, success: true, result: 'tests written' }
      ];
      
      const synthesis = await this.integration.synthesizeResults(
        mockResults, 
        'test task',
        'test-orch-123'
      );
      
      if (!synthesis.totalAgents || !synthesis.insights || !synthesis.combinedResult) {
        throw new Error('Result synthesis failed');
      }
    });
  }

  async runMemoryTests() {
    this.log(BLUE, '\n💾 Memory Management Tests');

    await this.test('Memory Store and Retrieve', async () => {
      const testData = { test: 'validation', timestamp: Date.now() };
      const key = 'validation/test/data';
      
      await this.integration.storeInMemory(key, testData);
      const retrieved = await this.integration.retrieveFromMemory(key);
      
      if (JSON.stringify(retrieved) !== JSON.stringify(testData)) {
        throw new Error('Memory store/retrieve failed');
      }
    });

    await this.test('Memory Compression', async () => {
      const largeData = {
        description: 'A very long description that should trigger compression logic',
        timestamp: new Date().toISOString(),
        capabilities: ['test1', 'test2', 'test3', 'test4', 'test5'],
        result: 'Some substantial result data that makes this object larger'
      };
      
      await this.integration.storeInMemory('validation/large', largeData, { compress: true });
      const retrieved = await this.integration.retrieveFromMemory('validation/large');
      
      if (JSON.stringify(retrieved) !== JSON.stringify(largeData)) {
        throw new Error('Memory compression failed');
      }
    });

    await this.test('Memory TTL', async () => {
      const shortData = { expires: 'soon' };
      await this.integration.storeInMemory('validation/ttl', shortData, { ttl: 100 });
      
      // Immediate retrieval should work
      const immediate = await this.integration.retrieveFromMemory('validation/ttl');
      if (JSON.stringify(immediate) !== JSON.stringify(shortData)) {
        throw new Error('TTL immediate retrieval failed');
      }
      
      // Wait for expiration
      await new Promise(resolve => setTimeout(resolve, 200));
      const expired = await this.integration.retrieveFromMemory('validation/ttl');
      if (expired !== null) {
        throw new Error('TTL expiration failed');
      }
    });

    await this.test('Memory Search', async () => {
      // Store test data
      await this.integration.storeInMemory('search/test1', { type: 'test' });
      await this.integration.storeInMemory('search/test2', { type: 'test' });
      
      const results = await this.integration.searchMemory('search/*');
      if (!Array.isArray(results)) {
        throw new Error('Memory search failed');
      }
    });
  }

  async runUtilityTests() {
    this.log(BLUE, '\n🔧 Utility Functions Tests');

    await this.test('File Context Analysis', async () => {
      const context = await this.integration.analyzeFileContext('create a React component for login.jsx');
      if (!context.detectedFileTypes || !context.suggestedAgents) {
        throw new Error('File context analysis failed');
      }
    });

    await this.test('Task Safety Validation', async () => {
      const safe = this.integration.validateTaskSafety('create a new component');
      const dangerous = this.integration.validateTaskSafety('rm -rf /important/data');
      
      if (!safe.safe || dangerous.safe) {
        throw new Error('Task safety validation failed');
      }
    });

    await this.test('Complexity Calculation', async () => {
      const simple = this.integration.calculateTaskComplexity('fix typo');
      const complex = this.integration.calculateTaskComplexity('implement distributed microservices architecture');
      
      if (simple >= complex) {
        throw new Error('Complexity calculation failed');
      }
    });

    await this.test('Efficiency Calculation', async () => {
      const efficiency = this.integration.calculateEfficiency(5000, 'simple task', true);
      if (efficiency <= 0 || efficiency > 1.0) {
        throw new Error('Efficiency calculation failed');
      }
    });

    await this.test('Metrics Collection', async () => {
      const metrics = await this.integration.collectMetrics();
      if (!metrics.timestamp || typeof metrics.totalAgents !== 'number') {
        throw new Error('Metrics collection failed');
      }
    });
  }

  async runErrorHandlingTests() {
    this.log(BLUE, '\n🛡️ Error Handling Tests');

    await this.test('Invalid Agent Type Handling', async () => {
      try {
        await this.integration.spawnAgent('invalid-agent-type', 'test task');
        throw new Error('Should have thrown error for invalid agent');
      } catch (error) {
        if (!error.message.includes('Unknown agent type')) {
          throw new Error('Wrong error type for invalid agent');
        }
      }
    });

    await this.test('Memory Fallback', async () => {
      // Store in fallback memory when primary fails
      const fallbackKey = 'fallback/test';
      const fallbackData = { fallback: true };
      
      await this.integration.storeInMemory(fallbackKey, fallbackData);
      const retrieved = await this.integration.retrieveFromMemory(fallbackKey);
      
      // Should work even if primary memory has issues
      if (!retrieved) {
        throw new Error('Memory fallback failed');
      }
    });

    await this.test('Graceful Shutdown', async () => {
      const testIntegration = createIntegration({ autoSwarm: false });
      await testIntegration.initialize();
      
      // Should not throw
      await testIntegration.shutdown();
      if (testIntegration.initialized) {
        throw new Error('Shutdown failed to clear initialized flag');
      }
    });

    await this.test('Emergency Shutdown', async () => {
      const testIntegration = createIntegration({ autoSwarm: false });
      await testIntegration.initialize();
      
      await testIntegration.emergencyShutdown();
      if (testIntegration.initialized) {
        throw new Error('Emergency shutdown failed');
      }
    });
  }

  printSummary() {
    console.log('\n' + '='.repeat(60));
    this.log(BLUE, '📊 VALIDATION SUMMARY');
    console.log('='.repeat(60));
    
    this.log(GREEN, `✅ Passed: ${this.results.passed}`);
    this.log(RED, `❌ Failed: ${this.results.failed}`);
    this.log(YELLOW, `⚠️  Warnings: ${this.results.warnings}`);
    
    const total = this.results.passed + this.results.failed;
    const successRate = total > 0 ? (this.results.passed / total * 100).toFixed(1) : 0;
    
    console.log(`\nSuccess Rate: ${successRate}%`);
    console.log(`Total Tests: ${total + this.results.warnings}`);
    
    if (this.results.failed > 0) {
      this.log(RED, '\n❌ FAILED TESTS:');
      this.results.tests
        .filter(t => t.status === 'FAIL')
        .forEach(t => this.log(RED, `  • ${t.name}: ${t.error}`));
    }
    
    if (this.results.warnings > 0) {
      this.log(YELLOW, '\n⚠️  WARNINGS:');
      this.results.tests
        .filter(t => t.status === 'WARN')
        .forEach(t => this.log(YELLOW, `  • ${t.name}: ${t.error}`));
    }
    
    console.log('\n' + '='.repeat(60));
    
    if (this.results.failed === 0) {
      this.log(GREEN, '🎉 ALL CORE TESTS PASSED! Claude Flow Integration is ready for use.');
    } else {
      this.log(RED, '💥 SOME TESTS FAILED! Please review and fix issues before deployment.');
    }
  }
}

// Auto-generate validation report
async function generateReport(results) {
  const report = {
    timestamp: new Date().toISOString(),
    version: '1.0.0',
    environment: {
      node: process.version,
      platform: process.platform,
      arch: process.arch
    },
    summary: {
      total: results.passed + results.failed + results.warnings,
      passed: results.passed,
      failed: results.failed,
      warnings: results.warnings,
      successRate: results.passed + results.failed > 0 ? 
        (results.passed / (results.passed + results.failed) * 100).toFixed(1) : 0
    },
    tests: results.tests,
    recommendation: results.failed === 0 ? 
      'READY_FOR_PRODUCTION' : 
      'REQUIRES_FIXES_BEFORE_DEPLOYMENT'
  };
  
  try {
    const reportPath = path.join(process.cwd(), 'validation-report.json');
    await fs.writeFile(reportPath, JSON.stringify(report, null, 2));
    console.log(`📄 Validation report saved to: ${reportPath}`);
  } catch (error) {
    console.warn('⚠️  Could not save validation report:', error.message);
  }
}

// Run validation
async function main() {
  const runner = new ValidationRunner();
  
  try {
    const success = await runner.run();
    await generateReport(runner.results);
    
    process.exit(success ? 0 : 1);
    
  } catch (error) {
    console.error(`${RED}💥 Validation runner crashed: ${error.message}${RESET}`);
    console.error(error.stack);
    process.exit(2);
  }
}

// Run if executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}

export { ValidationRunner };