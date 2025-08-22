import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import SPARCOrchestrator from '../src/sparc-orchestrator.js';
import { OrchestrationPatterns, selectPattern } from '../src/orchestration-patterns.js';

describe('SPARC Orchestrator', () => {
  let orchestrator;
  
  beforeEach(() => {
    orchestrator = new SPARCOrchestrator();
    // Mock executeCommand to avoid actual subprocess calls
    orchestrator.executeCommand = jest.fn().mockResolvedValue({
      success: true,
      output: 'mocked output'
    });
  });
  
  afterEach(async () => {
    await orchestrator.cleanup();
  });
  
  describe('Initialization', () => {
    it('should initialize orchestrator environment', async () => {
      const result = await orchestrator.initialize();
      
      expect(result).toBe(true);
      expect(orchestrator.executeCommand).toHaveBeenCalledWith(
        expect.stringContaining('swarm init')
      );
    });
    
    it('should emit initialized event', async () => {
      const listener = jest.fn();
      orchestrator.on('initialized', listener);
      
      await orchestrator.initialize();
      
      expect(listener).toHaveBeenCalled();
    });
  });
  
  describe('Task Decomposition', () => {
    it('should decompose task using domain strategy', () => {
      const task = 'Build user authentication system';
      const subtasks = orchestrator.decomposeTask(task, 'domain');
      
      expect(subtasks).toHaveLength(5);
      expect(subtasks[0].agent).toBe('researcher');
      expect(subtasks[1].agent).toBe('architect');
      expect(subtasks[2].agent).toBe('coder');
      expect(subtasks[3].agent).toBe('tester');
      expect(subtasks[4].agent).toBe('reviewer');
    });
    
    it('should decompose task using parallel strategy', () => {
      const task = 'Analyze system performance';
      const subtasks = orchestrator.decomposeTask(task, 'parallel');
      
      expect(subtasks).toHaveLength(3);
      expect(subtasks.every(st => st.parallel)).toBe(true);
    });
    
    it('should decompose task using sequential strategy', () => {
      const task = 'Implement SPARC methodology';
      const subtasks = orchestrator.decomposeTask(task, 'sequential');
      
      expect(subtasks).toHaveLength(5);
      expect(subtasks[0].sequence).toBe(1);
      expect(subtasks[4].sequence).toBe(5);
    });
    
    it('should use adaptive strategy based on complexity', () => {
      const simpleTask = 'Fix typo';
      const complexTask = 'Refactor system architecture for optimization';
      
      orchestrator.assessComplexity = jest.fn()
        .mockReturnValueOnce(0.2)  // Simple
        .mockReturnValueOnce(0.8); // Complex
      
      const simpleSubtasks = orchestrator.decomposeTask(simpleTask, 'adaptive');
      const complexSubtasks = orchestrator.decomposeTask(complexTask, 'adaptive');
      
      expect(simpleSubtasks.length).toBeLessThan(complexSubtasks.length);
    });
  });
  
  describe('Agent Management', () => {
    it('should spawn agent with correct properties', async () => {
      const agent = await orchestrator.spawnAgent('researcher', ['analysis', 'documentation']);
      
      expect(agent.type).toBe('researcher');
      expect(agent.capabilities).toEqual(['analysis', 'documentation']);
      expect(agent.status).toBe('idle');
      expect(agent.id).toMatch(/researcher-\d+/);
    });
    
    it('should track spawned agents', async () => {
      const agent1 = await orchestrator.spawnAgent('coder');
      const agent2 = await orchestrator.spawnAgent('tester');
      
      expect(orchestrator.agents.size).toBe(2);
      expect(orchestrator.agents.has(agent1.id)).toBe(true);
      expect(orchestrator.agents.has(agent2.id)).toBe(true);
    });
    
    it('should emit agent-spawned event', async () => {
      const listener = jest.fn();
      orchestrator.on('agent-spawned', listener);
      
      const agent = await orchestrator.spawnAgent('reviewer');
      
      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({ type: 'reviewer' })
      );
    });
  });
  
  describe('Task Coordination', () => {
    it('should coordinate task with default options', async () => {
      await orchestrator.initialize();
      
      const results = await orchestrator.coordinateTask('Test task');
      
      expect(results).toBeDefined();
      expect(results.taskId).toMatch(/task-\d+/);
      expect(results.task).toBe('Test task');
    });
    
    it('should execute subtasks in parallel when specified', async () => {
      await orchestrator.initialize();
      
      const startTime = Date.now();
      await orchestrator.coordinateTask('Parallel task', {
        strategy: 'parallel',
        parallel: true
      });
      const duration = Date.now() - startTime;
      
      // Parallel execution should be faster
      expect(duration).toBeLessThan(1000);
    });
    
    it('should execute subtasks sequentially when required', async () => {
      await orchestrator.initialize();
      
      const results = await orchestrator.coordinateTask('Sequential task', {
        strategy: 'sequential',
        parallel: false
      });
      
      expect(results).toBeDefined();
      expect(orchestrator.memory.size).toBeGreaterThan(0);
    });
  });
  
  describe('Memory Management', () => {
    it('should share memory between agents', async () => {
      await orchestrator.shareMemory('test-key', { data: 'test value' });
      
      expect(orchestrator.memory.has('test-key')).toBe(true);
      expect(orchestrator.memory.get('test-key').value).toEqual({ data: 'test value' });
    });
    
    it('should retrieve memory with TTL check', async () => {
      await orchestrator.shareMemory('ttl-test', 'value');
      
      const value = await orchestrator.getMemory('ttl-test');
      expect(value).toBe('value');
      
      // Simulate expired TTL
      const memData = orchestrator.memory.get('ttl-test');
      memData.timestamp = Date.now() - 4000000; // Expired
      
      const expiredValue = await orchestrator.getMemory('ttl-test');
      expect(expiredValue).toEqual({ success: true, output: 'mocked output' });
    });
  });
  
  describe('Progress Monitoring', () => {
    it('should return current progress status', async () => {
      await orchestrator.spawnAgent('coder');
      await orchestrator.spawnAgent('tester');
      
      const status = await orchestrator.monitorProgress();
      
      expect(status.agents).toHaveLength(2);
      expect(status.tasks).toHaveLength(0);
      expect(status.memory.entries).toBe(0);
    });
    
    it('should track task progress', async () => {
      await orchestrator.initialize();
      
      // Start a task
      const taskPromise = orchestrator.coordinateTask('Monitor this task');
      
      // Check progress while running
      await new Promise(resolve => setTimeout(resolve, 100));
      const status = await orchestrator.monitorProgress();
      
      expect(status.tasks.length).toBeGreaterThan(0);
      
      await taskPromise;
    });
  });
  
  describe('Result Aggregation', () => {
    it('should synthesize results correctly', () => {
      const results = [
        { success: true, insights: ['insight1'], recommendations: ['rec1'] },
        { success: false, error: 'test error' },
        { success: true, insights: ['insight2'], nextSteps: ['step1'] }
      ];
      
      const synthesis = orchestrator.synthesizeResults(results);
      
      expect(synthesis.successCount).toBe(2);
      expect(synthesis.failureCount).toBe(1);
      expect(synthesis.insights).toEqual(['insight1', 'insight2']);
      expect(synthesis.recommendations).toEqual(['rec1']);
      expect(synthesis.nextSteps).toEqual(['step1']);
    });
    
    it('should calculate metrics', async () => {
      const taskData = {
        duration: 5000,
        subtasks: [1, 2, 3],
        results: [
          { success: true },
          { success: true },
          { success: false }
        ]
      };
      
      const metrics = await orchestrator.calculateMetrics(taskData);
      
      expect(metrics.totalDuration).toBe(5000);
      expect(metrics.averageSubtaskDuration).toBeCloseTo(1666.67, 1);
      expect(metrics.successRate).toBeCloseTo(66.67, 1);
    });
  });
  
  describe('Complexity Assessment', () => {
    it('should assess simple task complexity', () => {
      const complexity = orchestrator.assessComplexity('Fix a typo');
      expect(complexity).toBeLessThan(0.3);
    });
    
    it('should assess moderate task complexity', () => {
      const complexity = orchestrator.assessComplexity('Implement user authentication with JWT tokens');
      expect(complexity).toBeGreaterThan(0.3);
      expect(complexity).toBeLessThan(0.7);
    });
    
    it('should assess complex task complexity', () => {
      const complexity = orchestrator.assessComplexity(
        'Refactor the entire system architecture to improve performance and security while maintaining backward compatibility'
      );
      expect(complexity).toBeGreaterThan(0.7);
    });
  });
});

describe('Orchestration Patterns', () => {
  describe('Pattern Selection', () => {
    it('should select hierarchical pattern by default', () => {
      const pattern = selectPattern('Simple task');
      expect(pattern).toBe(OrchestrationPatterns.hierarchical);
    });
    
    it('should select consensus pattern for review tasks', () => {
      const pattern = selectPattern('Review and approve the new feature');
      expect(pattern).toBe(OrchestrationPatterns.consensus);
    });
    
    it('should select mapReduce for data-intensive tasks', () => {
      const pattern = selectPattern('Process and analyze large dataset');
      expect(pattern).toBe(OrchestrationPatterns.mapReduce);
    });
    
    it('should select eventDriven for event-based tasks', () => {
      const pattern = selectPattern('Trigger deployment when tests pass');
      expect(pattern).toBe(OrchestrationPatterns.eventDriven);
    });
    
    it('should select adaptive for complex tasks', () => {
      const pattern = selectPattern('Redesign system architecture for optimization and integration');
      expect(pattern).toBe(OrchestrationPatterns.adaptive);
    });
    
    it('should respect user preference', () => {
      const pattern = selectPattern('Any task', { pattern: 'consensus' });
      expect(pattern).toBe(OrchestrationPatterns.consensus);
    });
  });
  
  describe('MapReduce Pattern', () => {
    it('should partition task correctly', () => {
      const pattern = OrchestrationPatterns.mapReduce;
      const chunks = pattern.partitionTask('one two three four five six', 3);
      
      expect(chunks).toHaveLength(3);
      expect(chunks[0]).toBe('one two');
      expect(chunks[1]).toBe('three four');
      expect(chunks[2]).toBe('five six');
    });
    
    it('should shuffle results by key', () => {
      const pattern = OrchestrationPatterns.mapReduce;
      const mapResults = [
        { output: { key: 'a' }, data: 1 },
        { output: { key: 'b' }, data: 2 },
        { output: { key: 'a' }, data: 3 }
      ];
      
      const shuffled = pattern.shuffle(mapResults);
      
      expect(shuffled).toHaveLength(2);
      expect(shuffled[0]).toHaveLength(2); // Two items with key 'a'
      expect(shuffled[1]).toHaveLength(1); // One item with key 'b'
    });
  });
});