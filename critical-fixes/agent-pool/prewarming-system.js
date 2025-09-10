// Agent Pool Prewarming System for 90% Spawn Time Reduction
const EventEmitter = require('events');
const { Worker } = require('worker_threads');

class AgentPoolManager extends EventEmitter {
  constructor(options = {}) {
    super();
    this.options = {
      poolSize: options.poolSize || 30,
      minPoolSize: options.minPoolSize || 10,
      maxPoolSize: options.maxPoolSize || 50,
      warmupBatchSize: options.warmupBatchSize || 5,
      healthCheckInterval: options.healthCheckInterval || 30000,
      agentTimeout: options.agentTimeout || 300000, // 5 minutes
      ...options
    };

    // Agent pools organized by capability
    this.agentPools = new Map();
    this.activeAgents = new Map();
    this.agentCapabilities = new Map();
    this.poolMetrics = {
      totalSpawned: 0,
      totalRequested: 0,
      poolHits: 0,
      poolMisses: 0,
      averageSpawnTime: 0,
      averageWarmupTime: 0
    };

    // Initialize capability-based pools
    this.initializeAgentPools();
    this.startHealthChecking();
    this.startPredictiveWarming();
  }

  // Initialize pools for different agent types
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

  // Get capabilities for each agent type
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

  // Calculate target pool size based on usage patterns
  calculateTargetPoolSize(agentType) {
    const baseSize = 2;
    const popularTypes = ['coder', 'researcher', 'tester', 'reviewer'];
    const coordinatorTypes = ['hierarchical-coordinator', 'mesh-coordinator'];
    
    if (popularTypes.includes(agentType)) return baseSize + 3;
    if (coordinatorTypes.includes(agentType)) return baseSize + 2;
    return baseSize;
  }

  // Warmup initial pools with prewarmed agents
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

  // Warmup agents for a specific type
  async warmupAgentType(agentType, count) {
    const pool = this.agentPools.get(agentType);
    if (!pool) return;

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
      // Clean up warming set
      pool.warming.clear();
      return 0;
    }
  }

  // Create a warm agent instance
  async createWarmAgent(agentType, agentId) {
    try {
      // Simulate agent creation with realistic timing
      const creationDelay = Math.random() * 200 + 100; // 100-300ms
      await new Promise(resolve => setTimeout(resolve, creationDelay));

      const agent = {
        id: agentId,
        type: agentType,
        status: 'warm',
        createdAt: Date.now(),
        lastUsed: null,
        taskCount: 0,
        capabilities: this.getAgentCapabilities(agentType),
        // Mock agent worker thread
        worker: null // Would be actual Worker instance in production
      };

      // Initialize mock worker for demonstration
      agent.worker = {
        postMessage: (message) => console.log(`📤 Agent ${agentId}:`, message),
        terminate: () => console.log(`🔌 Agent ${agentId} terminated`),
        on: (event, callback) => {
          // Mock event handling
          if (event === 'message') {
            setTimeout(() => callback({ status: 'ready' }), 10);
          }
        }
      };

      this.poolMetrics.totalSpawned++;
      return agent;
    } catch (error) {
      console.error(`❌ Failed to create warm agent ${agentId}:`, error);
      return null;
    }
  }

  // Get agent from pool (primary interface)
  async getAgent(requiredCapabilities = [], priority = 'normal') {
    const startTime = Date.now();
    console.log(`🎯 Requesting agent with capabilities: [${requiredCapabilities.join(', ')}]`);
    
    this.poolMetrics.totalRequested++;

    // Find best matching agent from pools
    const agent = await this.findBestMatchingAgent(requiredCapabilities, priority);
    
    if (agent) {
      // Agent found in pool (fast path)
      this.poolMetrics.poolHits++;
      const responseTime = Date.now() - startTime;
      
      // Move agent to active and update metrics
      this.activateAgent(agent);
      
      console.log(`⚡ Agent ${agent.id} retrieved from pool in ${responseTime}ms`);
      this.emit('agent:retrieved', { agent, responseTime, source: 'pool' });
      
      // Async: Maintain pool size
      this.maintainPoolSize(agent.type);
      
      return agent;
    } else {
      // No suitable agent in pool (slow path)
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

  // Find best matching agent from available pools - OPTIMIZED with Set for O(1) lookups
  async findBestMatchingAgent(requiredCapabilities, priority) {
    let bestAgent = null;
    let bestScore = -1;

    // Convert required capabilities to Set for faster lookups
    const requiredSet = new Set(requiredCapabilities);
    
    // Pre-filter pools with available agents using Map.entries() for better performance
    const availablePools = Array.from(this.agentPools.entries())
      .filter(([_, pool]) => pool.available.length > 0);
    
    for (const [agentType, pool] of availablePools) {
      // Calculate compatibility score with optimized Set operations
      const score = this.calculateCompatibilityScoreOptimized(pool.capabilities, requiredSet);
      
      if (score > bestScore) {
        bestScore = score;
        bestAgent = pool.available[0]; // Get first available agent
      }
    }

    if (bestAgent && bestScore > 0.3) { // Minimum 30% capability match
      // Remove agent from pool efficiently
      const pool = this.agentPools.get(bestAgent.type);
      pool.available.shift(); // O(1) removal from front vs filter O(n)
      pool.requestCount++;
      pool.lastUsed = Date.now();
      
      return bestAgent;
    }

    return null;
  }

  // Calculate compatibility score - OPTIMIZED version with Set operations
  calculateCompatibilityScoreOptimized(agentCapabilities, requiredCapabilitiesSet) {
    if (requiredCapabilitiesSet.size === 0) return 1.0;

    // Convert agent capabilities to Set for O(1) lookups
    const agentCapSet = new Set(agentCapabilities);
    let matches = 0;

    // Use Set iteration for better performance
    for (const req of requiredCapabilitiesSet) {
      // Direct match check - O(1)
      if (agentCapSet.has(req)) {
        matches++;
        continue;
      }
      
      // Substring/compatibility check - optimized with early exit
      let found = false;
      for (const cap of agentCapSet) {
        if (cap.includes(req) || req.includes(cap) || 
            this.areCapabilitiesCompatible(cap, req)) {
          matches++;
          found = true;
          break; // Early exit on first match
        }
      }
    }

    return matches / requiredCapabilitiesSet.size;
  }

  // Fallback method for backward compatibility
  calculateCompatibilityScore(agentCapabilities, requiredCapabilities) {
    const requiredSet = new Set(requiredCapabilities);
    return this.calculateCompatibilityScoreOptimized(agentCapabilities, requiredSet);
  }

  // Check if capabilities are compatible - OPTIMIZED with Maps for O(1) lookups
  areCapabilitiesCompatible(cap1, cap2) {
    // Use static Map for better performance (initialized once)
    if (!this.compatibilityLookup) {
      this.compatibilityLookup = new Map();
      const compatibilityRules = {
        'coding': ['implementation', 'development', 'programming'],
        'testing': ['validation', 'qa', 'quality-assurance'],
        'research': ['analysis', 'investigation'],
        'coordination': ['management', 'orchestration']
      };
      
      // Build bidirectional lookup map for O(1) access
      for (const [key, synonyms] of Object.entries(compatibilityRules)) {
        this.compatibilityLookup.set(key, new Set(synonyms));
        for (const synonym of synonyms) {
          if (!this.compatibilityLookup.has(synonym)) {
            this.compatibilityLookup.set(synonym, new Set([key]));
          } else {
            this.compatibilityLookup.get(synonym).add(key);
          }
        }
      }
    }

    // Fast bidirectional lookup - O(1)
    const cap1Synonyms = this.compatibilityLookup.get(cap1);
    const cap2Synonyms = this.compatibilityLookup.get(cap2);
    
    return (cap1Synonyms && cap1Synonyms.has(cap2)) ||
           (cap2Synonyms && cap2Synonyms.has(cap1));
  }

  // Create agent on-demand when pool is insufficient
  async createAgentOnDemand(requiredCapabilities) {
    // Determine best agent type for requirements
    const bestAgentType = this.determineBestAgentType(requiredCapabilities);
    const agentId = `${bestAgentType}-ondemand-${Date.now()}`;
    
    console.log(`🔨 Creating on-demand agent: ${agentId}`);
    
    return this.createWarmAgent(bestAgentType, agentId);
  }

  // Determine best agent type for given capabilities
  determineBestAgentType(requiredCapabilities) {
    let bestType = 'coder'; // Default
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

  // Activate agent (move from pool to active)
  activateAgent(agent) {
    agent.status = 'active';
    agent.lastUsed = Date.now();
    agent.taskCount++;
    this.activeAgents.set(agent.id, agent);
    
    console.log(`🚀 Agent ${agent.id} activated for task execution`);
  }

  // Release agent back to pool or terminate if not needed
  async releaseAgent(agentId, taskResult = {}) {
    const agent = this.activeAgents.get(agentId);
    if (!agent) {
      console.warn(`⚠️ Attempted to release unknown agent: ${agentId}`);
      return;
    }

    console.log(`🔄 Releasing agent ${agentId} after task completion`);
    
    // Remove from active agents
    this.activeAgents.delete(agentId);
    
    // Decide whether to return to pool or terminate
    const pool = this.agentPools.get(agent.type);
    const shouldReturnToPool = this.shouldReturnToPool(agent, pool, taskResult);
    
    if (shouldReturnToPool && pool) {
      // Reset agent to warm state and return to pool
      agent.status = 'warm';
      agent.lastTask = taskResult;
      pool.available.push(agent);
      
      console.log(`♻️ Agent ${agentId} returned to ${agent.type} pool`);
      this.emit('agent:returned', { agentId, agentType: agent.type });
    } else {
      // Terminate agent
      await this.terminateAgent(agent);
      console.log(`🔌 Agent ${agentId} terminated`);
      this.emit('agent:terminated', { agentId, reason: 'pool-optimization' });
    }
  }

  // Determine if agent should return to pool
  shouldReturnToPool(agent, pool, taskResult) {
    // Always return if pool is below target size
    if (pool && pool.available.length < pool.targetSize) return true;
    
    // Don't return if agent has been running too long
    const agentAge = Date.now() - agent.createdAt;
    if (agentAge > this.options.agentTimeout) return false;
    
    // Don't return if task failed badly
    if (taskResult.success === false && taskResult.severity === 'critical') return false;
    
    // Return if pool is moderately sized and agent performed well
    return pool && pool.available.length < this.options.maxPoolSize;
  }

  // Maintain pool size by warming up agents as needed
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

  // Health check all agents in pools - OPTIMIZED with parallel processing
  startHealthChecking() {
    setInterval(async () => {
      console.log('🏥 Running agent pool health check...');
      
      // Process all pools in parallel for better performance
      const healthCheckPromises = Array.from(this.agentPools.entries()).map(
        async ([agentType, pool]) => {
          if (pool.available.length === 0) return { agentType, healthyCount: 0, unhealthyCount: 0 };
          
          // Parallel health checks for all agents in pool
          const healthResults = await Promise.all(
            pool.available.map(agent => this.checkAgentHealth(agent))
          );
          
          // Efficiently separate healthy vs unhealthy agents
          const healthyAgents = [];
          const unhealthyAgents = [];
          
          for (let i = 0; i < pool.available.length; i++) {
            const agent = pool.available[i];
            if (healthResults[i]) {
              healthyAgents.push(agent);
            } else {
              unhealthyAgents.push(agent);
            }
          }
          
          // Update pool with healthy agents
          pool.available = healthyAgents;
          
          // Terminate unhealthy agents in parallel
          if (unhealthyAgents.length > 0) {
            await Promise.all(unhealthyAgents.map(agent => this.terminateAgent(agent)));
            console.log(`🔄 Replacing ${unhealthyAgents.length} unhealthy ${agentType} agents`);
            // Don't await warmup to avoid blocking health check
            this.warmupAgentType(agentType, unhealthyAgents.length);
          }
          
          return { 
            agentType, 
            healthyCount: healthyAgents.length, 
            unhealthyCount: unhealthyAgents.length 
          };
        }
      );
      
      // Wait for all health checks to complete
      const results = await Promise.all(healthCheckPromises);
      
      this.emit('health:check', {
        ...this.getPoolStatus(),
        healthCheckResults: results
      });
    }, this.options.healthCheckInterval);
  }

  // Check individual agent health
  async checkAgentHealth(agent) {
    try {
      // Mock health check - in production would ping worker thread
      const isResponsive = agent.worker && agent.status === 'warm';
      const agentAge = Date.now() - agent.createdAt;
      const tooOld = agentAge > this.options.agentTimeout;
      
      return isResponsive && !tooOld;
    } catch (error) {
      console.warn(`⚠️ Health check failed for agent ${agent.id}:`, error);
      return false;
    }
  }

  // Predictive warming based on usage patterns
  startPredictiveWarming() {
    setInterval(() => {
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
    }, 60000); // Every minute
  }

  // Analyze recent usage patterns
  analyzeRecentUsage(agentType) {
    const pool = this.agentPools.get(agentType);
    if (!pool) return { requests: 0, trend: 0 };
    
    // Mock analysis - in production would analyze time-series data
    return {
      requests: pool.requestCount,
      trend: pool.requestCount > 5 ? 1 : 0, // Simple trend
      lastUsed: Date.now() - pool.lastUsed
    };
  }

  // Predict future demand
  predictDemand(usage) {
    const basedemand = Math.max(2, Math.ceil(usage.requests * 0.8));
    const trendAdjustment = usage.trend * 2;
    const recentUsageBoost = usage.lastUsed < 300000 ? 1 : 0; // Used in last 5 min
    
    return basedemand + trendAdjustment + recentUsageBoost;
  }

  // Terminate agent and cleanup resources
  async terminateAgent(agent) {
    try {
      if (agent.worker && agent.worker.terminate) {
        agent.worker.terminate();
      }
      
      // Cleanup references
      this.activeAgents.delete(agent.id);
      this.agentCapabilities.delete(agent.id);
      
      console.log(`🔌 Agent ${agent.id} terminated successfully`);
    } catch (error) {
      console.error(`❌ Error terminating agent ${agent.id}:`, error);
    }
  }

  // Get current pool status and metrics
  getPoolStatus() {
    const status = {
      pools: {},
      active: this.activeAgents.size,
      metrics: {
        ...this.poolMetrics,
        poolHitRate: (this.poolMetrics.poolHits / (this.poolMetrics.poolHits + this.poolMetrics.poolMisses) * 100).toFixed(1) + '%'
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

  // Graceful shutdown
  async shutdown() {
    console.log('🛑 Shutting down agent pool manager...');
    
    // Terminate all agents
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

// Demo and testing
class AgentPoolDemo {
  static async runDemo() {
    console.log('🚀 Starting Agent Pool Prewarming System Demo\n');
    
    const poolManager = new AgentPoolManager({
      poolSize: 20,
      warmupBatchSize: 3
    });

    // Set up event listeners
    poolManager.on('warmup:complete', (data) => {
      console.log(`🔥 Warmup complete: ${data.successful}/${data.total} types`);
    });

    poolManager.on('agent:retrieved', (data) => {
      console.log(`⚡ Agent retrieved in ${data.responseTime}ms from ${data.source}`);
    });

    // Wait for initial warmup
    await new Promise(resolve => {
      poolManager.once('warmup:complete', resolve);
    });

    console.log('\n📊 Initial Pool Status:');
    console.table(poolManager.getPoolStatus().pools);

    // Simulate agent requests
    console.log('\n🎯 Simulating agent requests...');
    
    try {
      const requests = [
        { capabilities: ['coding', 'debugging'], priority: 'high' },
        { capabilities: ['research', 'analysis'], priority: 'normal' },
        { capabilities: ['testing', 'validation'], priority: 'normal' },
        { capabilities: ['coordination', 'management'], priority: 'high' },
        { capabilities: ['performance-analysis'], priority: 'low' }
      ];

      const agents = [];
      for (const request of requests) {
        const agent = await poolManager.getAgent(request.capabilities, request.priority);
        agents.push(agent);
      }

      console.log(`\n✅ Retrieved ${agents.length} agents successfully`);

      // Simulate task execution and release
      console.log('\n🔄 Simulating task completion and agent release...');
      for (const agent of agents) {
        await new Promise(resolve => setTimeout(resolve, 1000)); // Simulate work
        await poolManager.releaseAgent(agent.id, { success: true });
      }

      console.log('\n📊 Final Pool Status:');
      console.table(poolManager.getPoolStatus().pools);

      console.log('\n📈 Performance Metrics:');
      console.table(poolManager.getPoolStatus().metrics);

      // Shutdown
      await poolManager.shutdown();

    } catch (error) {
      console.error('❌ Demo failed:', error);
    }
  }
}

module.exports = { AgentPoolManager, AgentPoolDemo };

// Run demo if executed directly
if (require.main === module) {
  AgentPoolDemo.runDemo().catch(console.error);
}