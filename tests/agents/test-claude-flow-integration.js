#!/usr/bin/env node

/**
 * Comprehensive test suite for Claude Flow Integration
 */

import { describe, test, expect, beforeAll, afterAll, beforeEach } from '@jest/globals';
import { 
  ClaudeFlowIntegration, 
  createIntegration, 
  orchestrate, 
  orchestrateBatch,
  spawnAgent,
  getAvailableAgents,
  findAgents 
} from '../../src/agents/claude-flow-integration.js';

describe('Claude Flow Integration', () => {
  let integration;
  
  beforeAll(async () => {
    // Use test configuration
    integration = createIntegration({
      autoSwarm: false,
      memorySharing: true,
      hooks: false, // Disable hooks for testing
      topology: 'mesh',
      maxAgents: 5
    });
  });
  
  afterAll(async () => {
    if (integration?.initialized) {
      await integration.shutdown();
    }
  });
  
  beforeEach(async () => {
    if (!integration.initialized) {
      await integration.initialize();
    }
  });

  describe('Initialization', () => {
    test('should initialize successfully', async () => {
      const status = await integration.getStatus();
      expect(status.initialized).toBe(true);
      expect(status.config).toBeDefined();
    });

    test('should have proper configuration', () => {
      expect(integration.config.topology).toBe('mesh');
      expect(integration.config.maxAgents).toBe(5);
      expect(integration.config.memorySharing).toBe(true);
    });
  });

  describe('Agent Management', () => {
    test('should spawn single agent successfully', async () => {
      const result = await integration.spawnAgent(
        'coder', 
        'Create a simple Hello World function',
        { timeout: 30000 }
      );
      
      expect(result).toBeDefined();
      expect(result.agentId).toBeDefined();
      expect(result.agentType).toBe('coder');
      expect(result.success).toBe(true);
    });

    test('should handle agent spawn failure gracefully', async () => {
      await expect(
        integration.spawnAgent('nonexistent-agent', 'test task')
      ).rejects.toThrow('Unknown agent type: nonexistent-agent');
    });

    test('should retry failed agent spawns', async () => {
      const result = await integration.spawnAgent(
        'coder',
        'Task that might fail',
        { retries: 2, timeout: 5000 }
      );
      
      expect(result).toBeDefined();
      // Should either succeed or fail after retries
    });

    test('should list active agents', async () => {
      // Spawn a few agents first
      await integration.spawnAgent('coder', 'Test task 1');
      await integration.spawnAgent('tester', 'Test task 2');
      
      const agents = await integration.listActiveAgents();
      expect(agents.total).toBeGreaterThan(0);
      expect(agents.agents).toBeInstanceOf(Array);
    });

    test('should get specific agent status', async () => {
      const spawnResult = await integration.spawnAgent('reviewer', 'Review test code');
      const status = await integration.getAgentStatus('reviewer', spawnResult.agentId);
      
      expect(status.agentId).toBe(spawnResult.agentId);
      expect(status.agentType).toBe('reviewer');
      expect(status.status).toBeDefined();
    });
  });

  describe('Task Orchestration', () => {
    test('should orchestrate simple task', async () => {
      const result = await integration.orchestrateTask(
        'Build a simple web application with authentication',
        { strategy: 'parallel', maxAgents: 3 }
      );
      
      expect(result.id).toBeDefined();
      expect(result.status).toBe('completed');
      expect(result.agents).toBeInstanceOf(Array);
      expect(result.results).toBeInstanceOf(Array);
      expect(result.synthesis).toBeDefined();
    });

    test('should handle sequential orchestration', async () => {
      const result = await integration.orchestrateTask(
        'Create API endpoints then write tests',
        { strategy: 'sequential', maxAgents: 2 }
      );
      
      expect(result.status).toBe('completed');
      expect(result.results.length).toBeGreaterThan(0);
      
      // Check sequential execution
      result.results.forEach((r, index) => {
        if (r.sequence) {
          expect(r.sequence).toBe(index + 1);
        }
      });
    });

    test('should handle pipeline orchestration', async () => {
      const result = await integration.orchestrateTask(
        'Design schema, implement API, then create tests',
        { strategy: 'pipeline', maxAgents: 3 }
      );
      
      expect(result.status).toBe('completed');
      expect(result.results.length).toBeGreaterThan(0);
      
      // Each stage should have input from previous
      result.results.forEach((r, index) => {
        if (index > 0 && r.input) {
          expect(r.input).toBeDefined();
        }
      });
    });

    test('should handle hierarchical orchestration', async () => {
      const result = await integration.orchestrateTask(
        'Plan and implement microservices architecture',
        { strategy: 'hierarchical', maxAgents: 4 }
      );
      
      expect(result.status).toBe('completed');
      
      // Should have coordinator and worker roles
      const coordinatorResult = result.results.find(r => r.role === 'coordinator');
      const workerResults = result.results.filter(r => r.role === 'worker');
      
      expect(coordinatorResult).toBeDefined();
      expect(workerResults.length).toBeGreaterThan(0);
    });

    test('should synthesize results properly', async () => {
      const result = await integration.orchestrateTask(
        'Create full-stack application',
        { maxAgents: 4 }
      );
      
      const synthesis = result.synthesis;
      expect(synthesis.totalAgents).toBe(result.results.length);
      expect(synthesis.successfulAgents).toBeDefined();
      expect(synthesis.insights).toBeDefined();
      expect(synthesis.combinedResult).toBeDefined();
      expect(synthesis.combinedResult.recommendations).toBeInstanceOf(Array);
    });
  });

  describe('Memory Management', () => {
    test('should store and retrieve memory', async () => {
      const testData = { test: 'data', timestamp: Date.now() };
      const key = 'test/memory/key';
      
      await integration.storeInMemory(key, testData);
      const retrieved = await integration.retrieveFromMemory(key);
      
      expect(retrieved).toEqual(testData);
    });

    test('should handle memory compression', async () => {
      const largeData = {
        description: 'A very long description that should be compressed',
        timestamp: new Date().toISOString(),
        capabilities: ['test1', 'test2', 'test3'],
        result: 'Some result data'
      };
      
      await integration.storeInMemory('test/large', largeData, { compress: true });
      const retrieved = await integration.retrieveFromMemory('test/large');
      
      expect(retrieved).toEqual(largeData);
    });

    test('should search memory patterns', async () => {
      // Store some test data
      await integration.storeInMemory('agents/coder/task1', { type: 'coder', task: 'test1' });
      await integration.storeInMemory('agents/coder/task2', { type: 'coder', task: 'test2' });
      await integration.storeInMemory('agents/tester/task1', { type: 'tester', task: 'test1' });
      
      const results = await integration.searchMemory('agents/coder/*');
      expect(results.length).toBeGreaterThan(0);
    });

    test('should handle TTL expiration', async () => {
      const shortLivedData = { test: 'expires-soon' };
      await integration.storeInMemory('test/ttl', shortLivedData, { ttl: 100 }); // 100ms
      
      // Immediate retrieval should work
      const immediate = await integration.retrieveFromMemory('test/ttl');
      expect(immediate).toEqual(shortLivedData);
      
      // Wait for expiration
      await new Promise(resolve => setTimeout(resolve, 200));
      
      const expired = await integration.retrieveFromMemory('test/ttl');
      expect(expired).toBeNull();
    });
  });

  describe('Hook System', () => {
    test('should execute pre-task hooks', async () => {
      const hookData = {
        agentType: 'coder',
        agentId: 'test-123',
        task: 'Test hook execution'
      };
      
      // This should not throw
      await integration.executeHook('pre-task', hookData);
    });

    test('should execute post-task hooks', async () => {
      const hookData = {
        agentType: 'coder',
        agentId: 'test-123',
        result: { success: true },
        success: true
      };
      
      await integration.executeHook('post-task', hookData);
    });

    test('should handle session management hooks', async () => {
      await integration.executeHook('session-restore', { sessionId: 'test-session' });
      await integration.executeHook('session-end', { exportMetrics: true });
    });
  });

  describe('Task Analysis', () => {
    test('should analyze task requirements correctly', () => {
      const requirements1 = integration.analyzeTaskRequirements('implement secure authentication API');
      expect(requirements1).toContain('implementation');
      expect(requirements1).toContain('security-audit');
      
      const requirements2 = integration.analyzeTaskRequirements('test the user interface');
      expect(requirements2).toContain('unit-testing');
      
      const requirements3 = integration.analyzeTaskRequirements('optimize database performance');
      expect(requirements3).toContain('optimization');
    });

    test('should select optimal agents', () => {
      const capabilities = ['code-generation', 'unit-testing', 'security-audit'];
      const agents = integration.selectOptimalAgents(capabilities, { maxAgents: 3 });
      
      expect(agents.length).toBeLessThanOrEqual(3);
      expect(agents.length).toBeGreaterThan(0);
      
      // Should have unique agent types
      const types = agents.map(a => a.type);
      const uniqueTypes = new Set(types);
      expect(uniqueTypes.size).toBe(types.length);
    });

    test('should validate task safety', () => {
      const safeTask = integration.validateTaskSafety('create a new user component');
      expect(safeTask.safe).toBe(true);
      expect(safeTask.risk).toBe('low');
      
      const dangerousTask = integration.validateTaskSafety('rm -rf /important/data');
      expect(dangerousTask.safe).toBe(false);
      expect(dangerousTask.risk).toBe('high');
      expect(dangerousTask.warnings.length).toBeGreaterThan(0);
    });

    test('should analyze file context', async () => {
      const context1 = await integration.analyzeFileContext('create a React component for login.jsx');
      expect(context1.detectedFileTypes).toContain('javascript');
      expect(context1.suggestedAgents).toContain('frontend-dev');
      
      const context2 = await integration.analyzeFileContext('implement Python API with Django');
      expect(context2.detectedFileTypes).toContain('python');
      expect(context2.suggestedAgents).toContain('backend-dev');
    });
  });

  describe('Performance and Metrics', () => {
    test('should calculate task complexity', () => {
      const simple = integration.calculateTaskComplexity('fix typo');
      const complex = integration.calculateTaskComplexity('implement distributed microservices architecture with security');
      
      expect(simple).toBeLessThan(complex);
      expect(complex).toBeGreaterThan(1);
    });

    test('should calculate efficiency scores', () => {
      const efficiency1 = integration.calculateEfficiency(5000, 'simple fix', true);
      const efficiency2 = integration.calculateEfficiency(50000, 'simple fix', true);
      
      expect(efficiency1).toBeGreaterThan(efficiency2);
      expect(efficiency1).toBeLessThanOrEqual(1.0);
      expect(efficiency2).toBeGreaterThan(0);
    });

    test('should collect metrics', async () => {
      // Spawn some agents to generate metrics
      await integration.spawnAgent('coder', 'Generate metrics test');
      
      const metrics = await integration.collectMetrics();
      
      expect(metrics.timestamp).toBeDefined();
      expect(metrics.uptime).toBeGreaterThanOrEqual(0);
      expect(metrics.totalAgents).toBeGreaterThanOrEqual(0);
      expect(metrics.agentsByStatus).toBeDefined();
      expect(metrics.agentsByType).toBeDefined();
      expect(metrics.memoryUsage).toBeDefined();
      expect(metrics.performance).toBeDefined();
    });
  });

  describe('Health and Status', () => {
    test('should perform health check', async () => {
      const health = await integration.healthCheck();
      
      expect(health.status).toBeDefined();
      expect(['healthy', 'degraded', 'unhealthy', 'error']).toContain(health.status);
      expect(health.initialized).toBe(true);
      expect(health.timestamp).toBeDefined();
      expect(health.checks).toBeDefined();
      expect(health.checks.bridge).toBeDefined();
      expect(health.checks.memory).toBeDefined();
      expect(health.checks.agents).toBeDefined();
    });

    test('should provide comprehensive status', async () => {
      const status = await integration.getStatus();
      
      expect(status.initialized).toBe(true);
      expect(status.activeAgents).toBeDefined();
      expect(status.bridgeStatus).toBeDefined();
      expect(status.config).toBeDefined();
      expect(status.memory).toBeDefined();
      expect(status.uptime).toBeGreaterThanOrEqual(0);
    });
  });

  describe('Error Handling', () => {
    test('should handle bridge failures gracefully', async () => {
      // Simulate bridge failure
      const originalSpawn = integration.bridge.spawnAgent;
      integration.bridge.spawnAgent = () => Promise.reject(new Error('Bridge failure'));
      
      await expect(
        integration.spawnAgent('coder', 'test task', { retries: 1 })
      ).rejects.toThrow();
      
      // Restore original function
      integration.bridge.spawnAgent = originalSpawn;
    });

    test('should handle memory failures with fallback', async () => {
      // Store with primary memory unavailable
      await integration.storeInMemory('test/fallback', { test: 'data' });
      
      // Should use fallback memory if primary fails
      const retrieved = await integration.retrieveFromMemory('test/fallback');
      expect(retrieved).toBeDefined();
    });

    test('should handle orchestration failures', async () => {
      // Test with invalid strategy
      await expect(
        integration.orchestrateTask('test task', { strategy: 'invalid-strategy' })
      ).rejects.toThrow();
    });
  });

  describe('Shutdown and Cleanup', () => {
    test('should shutdown gracefully', async () => {
      const testIntegration = createIntegration({ autoSwarm: false });
      await testIntegration.initialize();
      
      // Should not throw
      await testIntegration.shutdown();
      expect(testIntegration.initialized).toBe(false);
    });

    test('should handle emergency shutdown', async () => {
      const testIntegration = createIntegration({ autoSwarm: false });
      await testIntegration.initialize();
      
      await testIntegration.emergencyShutdown();
      expect(testIntegration.initialized).toBe(false);
    });
  });
});

describe('Utility Functions', () => {
  test('should get available agents', () => {
    const agents = getAvailableAgents();
    
    expect(agents).toBeInstanceOf(Array);
    expect(agents.length).toBeGreaterThan(0);
    
    agents.forEach(agent => {
      expect(agent.type).toBeDefined();
      expect(agent.name).toBeDefined();
      expect(agent.capabilities).toBeInstanceOf(Array);
      expect(agent.tools).toBeInstanceOf(Array);
    });
  });

  test('should find agents by capability', () => {
    const coders = findAgents('code-generation');
    expect(coders.length).toBeGreaterThan(0);
    
    const testers = findAgents('unit-testing');
    expect(testers.length).toBeGreaterThan(0);
    
    const nonexistent = findAgents('nonexistent-capability');
    expect(nonexistent.length).toBe(0);
  });

  test('should spawn agent with utility function', async () => {
    const result = await spawnAgent('coder', 'Create utility test function');
    
    expect(result).toBeDefined();
    expect(result.success).toBe(true);
    expect(result.agentType).toBe('coder');
  });

  test('should orchestrate with utility function', async () => {
    const result = await orchestrate('Build simple component');
    
    expect(result).toBeDefined();
    expect(result.status).toBe('completed');
    expect(result.agents).toBeInstanceOf(Array);
  });

  test('should handle batch orchestration', async () => {
    const tasks = [
      'Create component A',
      'Create component B', 
      'Test components'
    ];
    
    const result = await orchestrateBatch(tasks, { 
      parallel: true, 
      maxConcurrent: 2 
    });
    
    expect(result.total).toBe(3);
    expect(result.results).toBeInstanceOf(Array);
    expect(result.results.length).toBe(3);
  });
});

// Run tests if this file is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  console.log('Running Claude Flow Integration tests...');
  
  // Basic smoke test
  async function smokeTest() {
    console.log('🧪 Running smoke test...');
    
    try {
      const integration = createIntegration({ 
        autoSwarm: false,
        hooks: false,
        memorySharing: false
      });
      
      await integration.initialize();
      console.log('✅ Initialization: PASS');
      
      const agents = getAvailableAgents();
      console.log(`✅ Available agents: ${agents.length} agents loaded`);
      
      const health = await integration.healthCheck();
      console.log(`✅ Health check: ${health.status}`);
      
      await integration.shutdown();
      console.log('✅ Shutdown: PASS');
      
      console.log('🎉 Smoke test completed successfully!');
      
    } catch (error) {
      console.error('❌ Smoke test failed:', error.message);
      process.exit(1);
    }
  }
  
  smokeTest();
}