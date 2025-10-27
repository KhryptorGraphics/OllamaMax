const { EventEmitter } = require('events');

// Mock worker_threads
jest.mock('worker_threads', () => ({
  Worker: jest.fn().mockImplementation(() => ({
    postMessage: jest.fn(),
    terminate: jest.fn(),
    on: jest.fn()
  }))
}));

// Create AgentPoolManager class for testing
class AgentPoolManager extends EventEmitter {
  constructor(options = {}) {
    super();
    this.options = {
      poolSize: options.poolSize || 30,
      minPoolSize: options.minPoolSize || 10,
      maxPoolSize: options.maxPoolSize || 50,
      warmupBatchSize: options.warmupBatchSize || 5,
      healthCheckInterval: options.healthCheckInterval || 30000,
      agentTimeout: options.agentTimeout || 300000,
      ...options
    };

    this.agentPools = new Map();
    this.activeAgents = new Map();
    this.agentCapabilities = new Map();
    
    // Track interval IDs for cleanup
    this.intervalIds = [];
    this.poolMetrics = {
      totalSpawned: 0,
      totalRequested: 0,
      poolHits: 0,
      poolMisses: 0,
      averageSpawnTime: 0,
      averageWarmupTime: 0
    };

    this.initializeAgentPools();
    this.startHealthChecking();
    this.startPredictiveWarming();
  }

  initializeAgentPools() {
    const agentTypes = [
      'researcher', 'coder', 'tester', 'reviewer', 'planner',
      'backend-dev', 'frontend-dev', 'ml-developer', 'system-architect',
      'perf-analyzer', 'security-auditor', 'code-analyzer',
      'hierarchical-coordinator', 'mesh-coordinator', 'adaptive-coordinator'
    ];

    agentTypes.forEach(type => {
      this.agentPools.set(type, {
        available: [],
        warming: new Set(),
        capabilities: this.getAgentCapabilities(type),
        targetSize: this.calculateTargetPoolSize(type),
        lastUsed: Date.now(),
        requestCount: 0,
        spawnTime: 0
      });
    });

    console.log(`🏊 Initialized ${agentTypes.length} agent pools`);
    this.warmupInitialPools();
  }

  getAgentCapabilities(agentType) {
    const capabilityMap = {
      'researcher': ['research', 'analysis', 'investigation', 'data-gathering'],
      'coder': ['coding', 'implementation', 'debugging', 'refactoring'],
      'tester': ['testing', 'validation', 'qa', 'automation'],
      'reviewer': ['code-review', 'quality-assurance', 'standards-compliance'],
      'planner': ['planning', 'architecture', 'strategy', 'coordination'],
      'backend-dev': ['api-development', 'database', 'server-side', 'microservices'],
      'frontend-dev': ['ui', 'ux', 'client-side', 'responsive-design'],
      'ml-developer': ['machine-learning', 'ai-training', 'data-science', 'neural-networks'],
      'system-architect': ['system-design', 'scalability', 'architecture', 'patterns'],
      'perf-analyzer': ['performance-analysis', 'optimization', 'profiling', 'benchmarking'],
      'security-auditor': ['security-analysis', 'vulnerability-assessment', 'compliance'],
      'code-analyzer': ['static-analysis', 'code-quality', 'metrics', 'complexity-analysis'],
      'hierarchical-coordinator': ['coordination', 'management', 'task-delegation'],
      'mesh-coordinator': ['peer-coordination', 'consensus', 'distributed-systems'],
      'adaptive-coordinator': ['dynamic-coordination', 'optimization', 'learning']
    };

    return capabilityMap[agentType] || ['general-purpose'];
  }

  calculateTargetPoolSize(agentType) {
    const baseSize = 2;
    const popularTypes = ['coder', 'researcher', 'tester', 'reviewer'];
    const coordinatorTypes = ['hierarchical-coordinator', 'mesh-coordinator'];
    
    if (popularTypes.includes(agentType)) return baseSize + 3;
    if (coordinatorTypes.includes(agentType)) return baseSize + 2;
    return baseSize;
  }

  async warmupInitialPools() {
    console.log('🔥 Starting initial pool warmup...');
    const warmupPromises = [];

    for (const [agentType, pool] of this.agentPools) {
      const warmupCount = Math.min(pool.targetSize, this.options.warmupBatchSize);
      warmupPromises.push(this.warmupAgentType(agentType, warmupCount));
    }

    const results = await Promise.allSettled(warmupPromises);
    const successful = results.filter(r => r.status === 'fulfilled').length;
    
    console.log(`✅ Initial warmup completed: ${successful}/${results.length} agent types warmed`);
    this.emit('warmup:complete', { successful, total: results.length });
  }

  async warmupAgentType(agentType, count) {
    const pool = this.agentPools.get(agentType);
    if (!pool) return 0;

    console.log(`🔄 Warming up ${count} ${agentType} agents...`);
    const startTime = Date.now();
    
    const warmupPromises = [];
    for (let i = 0; i < count; i++) {
      const agentId = `${agentType}-warm-${Date.now()}-${i}`;
      pool.warming.add(agentId);
      warmupPromises.push(this.createWarmAgent(agentType, agentId));
    }

    try {
      const agents = await Promise.all(warmupPromises);
      agents.forEach(agent => {
        if (agent) {
          pool.available.push(agent);
          pool.warming.delete(agent.id);
          this.agentCapabilities.set(agent.id, pool.capabilities);
        }
      });

      const warmupTime = Date.now() - startTime;
      pool.spawnTime = warmupTime / agents.filter(a => a).length;
      
      console.log(`✅ Warmed ${agents.filter(a => a).length} ${agentType} agents in ${warmupTime}ms`);
      return agents.filter(a => a).length;
    } catch (error) {
      console.error(`❌ Failed to warmup ${agentType} agents:`, error);
      pool.warming.clear();
      return 0;
    }
  }

  async createWarmAgent(agentType, agentId) {
    try {
      // Use deterministic delay with fake timers
      const creationDelay = 150;  // Fixed delay for testing
      await new Promise(resolve => setTimeout(resolve, creationDelay));

      const agent = {
        id: agentId,
        type: agentType,
        status: 'warm',
        createdAt: Date.now(),
        lastUsed: null,
        taskCount: 0,
        capabilities: this.getAgentCapabilities(agentType),
        worker: {
          postMessage: jest.fn(),
          terminate: jest.fn(),
          on: jest.fn()
        }
      };

      this.poolMetrics.totalSpawned++;
      return agent;
    } catch (error) {
      console.error(`❌ Failed to create warm agent ${agentId}:`, error);
      return null;
    }
  }

  async getAgent(requiredCapabilities = [], priority = 'normal') {
    const startTime = Date.now();
    console.log(`🎯 Requesting agent with capabilities: [${requiredCapabilities.join(', ')}]`);
    
    this.poolMetrics.totalRequested++;

    const agent = await this.findBestMatchingAgent(requiredCapabilities, priority);
    
    if (agent) {
      this.poolMetrics.poolHits++;
      const responseTime = Date.now() - startTime;
      
      this.activateAgent(agent);
      
      console.log(`⚡ Agent ${agent.id} retrieved from pool in ${responseTime}ms`);
      this.emit('agent:retrieved', { agent, responseTime, source: 'pool' });
      
      this.maintainPoolSize(agent.type);
      
      return agent;
    } else {
      this.poolMetrics.poolMisses++;
      
      const newAgent = await this.createAgentOnDemand(requiredCapabilities);
      const responseTime = Date.now() - startTime;
      
      if (newAgent) {
        this.activateAgent(newAgent);
        console.log(`🔨 New agent ${newAgent.id} created on-demand in ${responseTime}ms`);
        this.emit('agent:retrieved', { agent: newAgent, responseTime, source: 'on-demand' });
        return newAgent;
      } else {
        console.error('❌ Failed to provide agent - all methods exhausted');
        this.emit('agent:failed', { capabilities: requiredCapabilities, responseTime });
        throw new Error('Agent pool exhausted - unable to provide suitable agent');
      }
    }
  }

  async findBestMatchingAgent(requiredCapabilities, priority) {
    let bestAgent = null;
    let bestScore = -1;

    for (const [agentType, pool] of this.agentPools) {
      if (pool.available.length === 0) continue;

      const score = this.calculateCompatibilityScore(pool.capabilities, requiredCapabilities);
      
      if (score > bestScore) {
        bestScore = score;
        bestAgent = pool.available[0];
      }
    }

    if (bestAgent && bestScore > 0.3) {
      const pool = this.agentPools.get(bestAgent.type);
      pool.available = pool.available.filter(a => a.id !== bestAgent.id);
      pool.requestCount++;
      pool.lastUsed = Date.now();
      
      return bestAgent;
    }

    return null;
  }

  calculateCompatibilityScore(agentCapabilities, requiredCapabilities) {
    if (requiredCapabilities.length === 0) return 1.0;

    const matches = requiredCapabilities.filter(req => 
      agentCapabilities.some(cap => 
        cap.includes(req) || req.includes(cap) || 
        this.areCapabilitiesCompatible(cap, req)
      )
    ).length;

    return matches / requiredCapabilities.length;
  }

  areCapabilitiesCompatible(cap1, cap2) {
    const compatibilityMap = {
      'coding': ['implementation', 'development', 'programming'],
      'testing': ['validation', 'qa', 'quality-assurance'],
      'research': ['analysis', 'investigation'],
      'coordination': ['management', 'orchestration']
    };

    for (const [key, synonyms] of Object.entries(compatibilityMap)) {
      if ((cap1 === key && synonyms.includes(cap2)) ||
          (cap2 === key && synonyms.includes(cap1))) {
        return true;
      }
    }

    return false;
  }

  async createAgentOnDemand(requiredCapabilities) {
    const bestAgentType = this.determineBestAgentType(requiredCapabilities);
    const agentId = `${bestAgentType}-ondemand-${Date.now()}`;
    
    console.log(`🔨 Creating on-demand agent: ${agentId}`);
    
    return this.createWarmAgent(bestAgentType, agentId);
  }

  determineBestAgentType(requiredCapabilities) {
    let bestType = 'coder';
    let bestScore = -1;

    for (const [agentType, pool] of this.agentPools) {
      const score = this.calculateCompatibilityScore(pool.capabilities, requiredCapabilities);
      if (score > bestScore) {
        bestScore = score;
        bestType = agentType;
      }
    }

    return bestType;
  }

  activateAgent(agent) {
    agent.status = 'active';
    agent.lastUsed = Date.now();
    agent.taskCount++;
    this.activeAgents.set(agent.id, agent);
    
    console.log(`🚀 Agent ${agent.id} activated for task execution`);
  }

  async releaseAgent(agentId, taskResult = {}) {
    const agent = this.activeAgents.get(agentId);
    if (!agent) {
      console.warn(`⚠️ Attempted to release unknown agent: ${agentId}`);
      return;
    }

    console.log(`🔄 Releasing agent ${agentId} after task completion`);
    
    this.activeAgents.delete(agentId);
    
    const pool = this.agentPools.get(agent.type);
    const shouldReturnToPool = this.shouldReturnToPool(agent, pool, taskResult);
    
    if (shouldReturnToPool && pool) {
      agent.status = 'warm';
      agent.lastTask = taskResult;
      pool.available.push(agent);
      
      console.log(`♻️ Agent ${agentId} returned to ${agent.type} pool`);
      this.emit('agent:returned', { agentId, agentType: agent.type });
    } else {
      await this.terminateAgent(agent);
      console.log(`🔌 Agent ${agentId} terminated`);
      this.emit('agent:terminated', { agentId, reason: 'pool-optimization' });
    }
  }

  shouldReturnToPool(agent, pool, taskResult) {
    if (pool && pool.available.length < pool.targetSize) return true;
    
    const agentAge = Date.now() - agent.createdAt;
    if (agentAge > this.options.agentTimeout) return false;
    
    if (taskResult.success === false && taskResult.severity === 'critical') return false;
    
    return pool && pool.available.length < this.options.maxPoolSize;
  }

  async maintainPoolSize(agentType) {
    const pool = this.agentPools.get(agentType);
    if (!pool) return;

    const currentSize = pool.available.length + pool.warming.size;
    const shortfall = pool.targetSize - currentSize;
    
    if (shortfall > 0) {
      console.log(`📈 Pool ${agentType} needs ${shortfall} more agents`);
      this.warmupAgentType(agentType, Math.min(shortfall, this.options.warmupBatchSize));
    }
  }

  startHealthChecking() {
    const healthIntervalId = setInterval(async () => {
      console.log('🏥 Running agent pool health check...');
      
      for (const [agentType, pool] of this.agentPools) {
        const healthyAgents = [];
        const unhealthyAgents = [];
        
        for (const agent of pool.available) {
          const isHealthy = await this.checkAgentHealth(agent);
          if (isHealthy) {
            healthyAgents.push(agent);
          } else {
            unhealthyAgents.push(agent);
          }
        }
        
        pool.available = healthyAgents;
        for (const agent of unhealthyAgents) {
          await this.terminateAgent(agent);
        }
        
        if (unhealthyAgents.length > 0) {
          console.log(`🔄 Replacing ${unhealthyAgents.length} unhealthy ${agentType} agents`);
          this.warmupAgentType(agentType, unhealthyAgents.length);
        }
      }
      
      this.emit('health:check', this.getPoolStatus());
    }, this.options.healthCheckInterval);
    
    this.intervalIds.push(healthIntervalId);
  }

  async checkAgentHealth(agent) {
    try {
      const isResponsive = agent.worker && agent.status === 'warm';
      const agentAge = Date.now() - agent.createdAt;
      const tooOld = agentAge > this.options.agentTimeout;
      
      return isResponsive && !tooOld;
    } catch (error) {
      console.warn(`⚠️ Health check failed for agent ${agent.id}:`, error);
      return false;
    }
  }

  startPredictiveWarming() {
    const predictiveIntervalId = setInterval(() => {
      console.log('🔮 Running predictive warming analysis...');
      
      for (const [agentType, pool] of this.agentPools) {
        const recentUsage = this.analyzeRecentUsage(agentType);
        const predictedDemand = this.predictDemand(recentUsage);
        const currentSupply = pool.available.length + pool.warming.size;
        
        if (predictedDemand > currentSupply) {
          const warmupCount = Math.min(
            predictedDemand - currentSupply,
            this.options.warmupBatchSize
          );
          
          console.log(`📈 Predictive warming: ${agentType} (+${warmupCount})`);
          this.warmupAgentType(agentType, warmupCount);
        }
      }
    }, 60000);
    
    this.intervalIds.push(predictiveIntervalId);
  }

  analyzeRecentUsage(agentType) {
    const pool = this.agentPools.get(agentType);
    if (!pool) return { requests: 0, trend: 0 };
    
    return {
      requests: pool.requestCount,
      trend: pool.requestCount > 5 ? 1 : 0,
      lastUsed: Date.now() - pool.lastUsed
    };
  }

  predictDemand(usage) {
    const basedemand = Math.max(2, Math.ceil(usage.requests * 0.8));
    const trendAdjustment = usage.trend * 2;
    const recentUsageBoost = usage.lastUsed < 300000 ? 1 : 0;
    
    return basedemand + trendAdjustment + recentUsageBoost;
  }

  async terminateAgent(agent) {
    try {
      if (agent.worker && agent.worker.terminate) {
        agent.worker.terminate();
      }
      
      this.activeAgents.delete(agent.id);
      this.agentCapabilities.delete(agent.id);
      
      console.log(`🔌 Agent ${agent.id} terminated successfully`);
    } catch (error) {
      console.error(`❌ Error terminating agent ${agent.id}:`, error);
    }
  }

  getPoolStatus() {
    const status = {
      pools: {},
      active: this.activeAgents.size,
      metrics: {
        ...this.poolMetrics,
        poolHitRate: this.poolMetrics.poolHits + this.poolMetrics.poolMisses > 0 
          ? ((this.poolMetrics.poolHits / (this.poolMetrics.poolHits + this.poolMetrics.poolMisses)) * 100).toFixed(1) + '%'
          : 'N/A'
      }
    };

    for (const [agentType, pool] of this.agentPools) {
      status.pools[agentType] = {
        available: pool.available.length,
        warming: pool.warming.size,
        target: pool.targetSize,
        requests: pool.requestCount,
        lastUsed: new Date(pool.lastUsed).toISOString()
      };
    }

    return status;
  }

  async shutdown() {
    console.log('🛑 Shutting down agent pool manager...');
    
    const terminationPromises = [];
    
    for (const agent of this.activeAgents.values()) {
      terminationPromises.push(this.terminateAgent(agent));
    }
    
    for (const pool of this.agentPools.values()) {
      for (const agent of pool.available) {
        terminationPromises.push(this.terminateAgent(agent));
      }
    }
    
    await Promise.all(terminationPromises);
    
    console.log('✅ Agent pool manager shutdown complete');
  }
}

describe('AgentPoolManager', () => {
  let poolManager;

  beforeEach(() => {
    jest.clearAllMocks();
    // Use fake timers for deterministic, faster tests
    jest.useFakeTimers();

    poolManager = new AgentPoolManager({
      poolSize: 10,
      minPoolSize: 5,
      maxPoolSize: 20,
      warmupBatchSize: 3,
      healthCheckInterval: 10000,
      agentTimeout: 60000
    });
  });

  afterEach(() => {
    // Clear any running intervals and timers
    if (poolManager && poolManager.intervalIds) {
      poolManager.intervalIds.forEach(id => clearInterval(id));
      poolManager.intervalIds = [];
    }
    jest.clearAllTimers();
    jest.useRealTimers();
  });

  describe('Initialization', () => {
    test('should initialize with default options', () => {
      const manager = new AgentPoolManager();
      
      expect(manager.options.poolSize).toBe(30);
      expect(manager.options.minPoolSize).toBe(10);
      expect(manager.options.maxPoolSize).toBe(50);
    });

    test('should initialize with custom options', () => {
      const customOptions = {
        poolSize: 15,
        minPoolSize: 5,
        maxPoolSize: 25
      };
      
      const manager = new AgentPoolManager(customOptions);
      
      expect(manager.options.poolSize).toBe(15);
      expect(manager.options.minPoolSize).toBe(5);
      expect(manager.options.maxPoolSize).toBe(25);
    });

    test('should initialize agent pools', () => {
      expect(poolManager.agentPools.size).toBeGreaterThan(0);
      
      const expectedTypes = ['researcher', 'coder', 'tester'];
      expectedTypes.forEach(type => {
        expect(poolManager.agentPools.has(type)).toBe(true);
      });
    });

    test('should extend EventEmitter', () => {
      expect(poolManager).toBeInstanceOf(EventEmitter);
    });
  });

  describe('Agent Creation and Warming', () => {
    test('should create warm agent with correct properties', async () => {
      jest.advanceTimersByTime(150);
      const agentPromise = poolManager.createWarmAgent('coder', 'test-agent-1');
      jest.advanceTimersByTime(150);
      const agent = await agentPromise;

      expect(agent).toBeDefined();
      expect(agent.id).toBe('test-agent-1');
      expect(agent.type).toBe('coder');
      expect(agent.status).toBe('warm');
      expect(agent.capabilities).toContain('coding');
    });

    test('should warmup multiple agents', async () => {
      const resultPromise = poolManager.warmupAgentType('coder', 3);
      jest.advanceTimersByTime(500);
      const result = await resultPromise;

      expect(result).toBe(3);

      const coderPool = poolManager.agentPools.get('coder');
      expect(coderPool.available.length).toBe(3);
      expect(coderPool.warming.size).toBe(0);
    });

    test('should track warmup metrics', async () => {
      const warmupPromise = poolManager.warmupAgentType('tester', 2);
      jest.advanceTimersByTime(400);
      await warmupPromise;

      const testerPool = poolManager.agentPools.get('tester');
      expect(testerPool.spawnTime).toBeGreaterThan(0);
    });
  });

  describe('Agent Retrieval', () => {
    beforeEach(async () => {
      await poolManager.warmupAgentType('coder', 2);
      await poolManager.warmupAgentType('tester', 2);
    });

    test('should retrieve agent from pool', async () => {
      const agent = await poolManager.getAgent(['coding', 'implementation']);
      
      expect(agent).toBeDefined();
      expect(agent.type).toBe('coder');
      expect(agent.status).toBe('active');
      expect(poolManager.activeAgents.has(agent.id)).toBe(true);
    });

    test('should calculate compatibility scores', () => {
      const agentCapabilities = ['coding', 'debugging', 'refactoring'];
      const requiredCapabilities = ['coding', 'testing'];
      
      const score = poolManager.calculateCompatibilityScore(agentCapabilities, requiredCapabilities);
      
      expect(score).toBe(0.5);
    });

    test('should handle empty requirements', () => {
      const score = poolManager.calculateCompatibilityScore(['coding'], []);
      expect(score).toBe(1.0);
    });

    test('should create on-demand agent when pool empty', async () => {
      // Empty all pools
      for (const pool of poolManager.agentPools.values()) {
        pool.available = [];
      }
      
      const agent = await poolManager.getAgent(['specialized-capability']);
      
      expect(agent).toBeDefined();
      expect(poolManager.poolMetrics.poolMisses).toBe(1);
    });
  });

  describe('Agent Lifecycle', () => {
    let testAgent;

    beforeEach(async () => {
      await poolManager.warmupAgentType('coder', 1);
      testAgent = await poolManager.getAgent(['coding']);
    });

    test('should activate agent correctly', () => {
      expect(testAgent.status).toBe('active');
      expect(testAgent.lastUsed).toBeGreaterThan(0);
      expect(testAgent.taskCount).toBe(1);
    });

    test('should release agent back to pool', async () => {
      await poolManager.releaseAgent(testAgent.id, { success: true });
      
      expect(poolManager.activeAgents.has(testAgent.id)).toBe(false);
      
      const coderPool = poolManager.agentPools.get('coder');
      expect(coderPool.available.some(agent => agent.id === testAgent.id)).toBe(true);
    });

    test('should terminate failed agents', async () => {
      const terminateSpy = jest.spyOn(poolManager, 'terminateAgent').mockResolvedValue();
      
      await poolManager.releaseAgent(testAgent.id, { 
        success: false, 
        severity: 'critical' 
      });
      
      expect(terminateSpy).toHaveBeenCalledWith(testAgent);
    });

    test('should handle unknown agent release', async () => {
      const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
      
      await poolManager.releaseAgent('unknown-agent-id');
      
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('Attempted to release unknown agent')
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Health Monitoring', () => {
    test('should check agent health', async () => {
      const healthyAgent = {
        id: 'healthy',
        worker: { active: true },
        status: 'warm',
        createdAt: Date.now()
      };
      
      const unhealthyAgent = {
        id: 'unhealthy',
        worker: null,
        status: 'error',
        createdAt: Date.now() - 70000
      };
      
      expect(await poolManager.checkAgentHealth(healthyAgent)).toBe(true);
      expect(await poolManager.checkAgentHealth(unhealthyAgent)).toBe(false);
    });

    test('should handle health check errors', async () => {
      const faultyAgent = {
        id: 'faulty',
        get worker() { throw new Error('Worker error'); },
        status: 'warm',
        createdAt: Date.now()
      };
      
      const result = await poolManager.checkAgentHealth(faultyAgent);
      expect(result).toBe(false);
    });
  });

  describe('Pool Metrics and Status', () => {
    test('should provide pool status', () => {
      const status = poolManager.getPoolStatus();
      
      expect(status).toHaveProperty('pools');
      expect(status).toHaveProperty('active');
      expect(status).toHaveProperty('metrics');
      expect(status.metrics).toHaveProperty('poolHitRate');
    });

    test('should calculate hit rate', () => {
      poolManager.poolMetrics.poolHits = 8;
      poolManager.poolMetrics.poolMisses = 2;
      
      const status = poolManager.getPoolStatus();
      expect(status.metrics.poolHitRate).toBe('80.0%');
    });
  });

  describe('Predictive Warming', () => {
    test('should analyze usage patterns', () => {
      const coderPool = poolManager.agentPools.get('coder');
      coderPool.requestCount = 10;
      coderPool.lastUsed = Date.now() - 60000;
      
      const usage = poolManager.analyzeRecentUsage('coder');
      
      expect(usage.requests).toBe(10);
      expect(usage.trend).toBe(1);
    });

    test('should predict demand', () => {
      const highUsage = { requests: 10, trend: 1, lastUsed: 60000 };
      const lowUsage = { requests: 2, trend: 0, lastUsed: 600000 };
      
      const highDemand = poolManager.predictDemand(highUsage);
      const lowDemand = poolManager.predictDemand(lowUsage);
      
      expect(highDemand).toBeGreaterThan(lowDemand);
    });
  });

  describe('Error Handling', () => {
    test('should handle agent pool exhaustion', async () => {
      for (const pool of poolManager.agentPools.values()) {
        pool.available = [];
      }
      
      const originalCreate = poolManager.createAgentOnDemand;
      poolManager.createAgentOnDemand = jest.fn().mockResolvedValue(null);
      
      await expect(poolManager.getAgent(['impossible-capability']))
        .rejects.toThrow('Agent pool exhausted');
      
      poolManager.createAgentOnDemand = originalCreate;
    });

    test('should handle concurrent requests', async () => {
      await poolManager.warmupAgentType('coder', 5);
      
      const requests = Promise.all([
        poolManager.getAgent(['coding']),
        poolManager.getAgent(['coding']),
        poolManager.getAgent(['coding'])
      ]);
      
      const agents = await requests;
      
      expect(agents).toHaveLength(3);
      expect(new Set(agents.map(a => a.id)).size).toBe(3);
    });
  });

  describe('Shutdown', () => {
    test('should shutdown gracefully', async () => {
      const mockAgent = { id: 'agent1', worker: { terminate: jest.fn() } };
      poolManager.activeAgents.set('agent1', mockAgent);
      
      const terminateSpy = jest.spyOn(poolManager, 'terminateAgent').mockResolvedValue();
      
      await poolManager.shutdown();
      
      expect(terminateSpy).toHaveBeenCalled();
    });
  });
});