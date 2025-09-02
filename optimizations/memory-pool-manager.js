#!/usr/bin/env node

/**
 * Memory Pool Manager for OllamaMax
 * Implements object pooling to reduce memory allocation overhead
 */

class MemoryPoolManager {
  constructor() {
    this.pools = new Map();
    this.metrics = {
      total_objects_created: 0,
      total_objects_reused: 0,
      pools_created: 0,
      memory_saved_mb: 0,
      gc_pressure_reduction: 0
    };
    
    this.initializeCommonPools();
  }

  /**
   * Initialize object pools for common objects
   */
  initializeCommonPools() {
    // Agent instance pool
    this.createPool('agent_instances', {
      factory: () => ({
        id: null,
        type: null,
        status: 'idle',
        memory: new Map(),
        taskQueue: [],
        performance: { requests: 0, avgTime: 0 },
        reset() {
          this.id = null;
          this.type = null;
          this.status = 'idle';
          this.memory.clear();
          this.taskQueue.length = 0;
          this.performance = { requests: 0, avgTime: 0 };
        }
      }),
      maxSize: 20,
      preAllocate: 5
    });

    // Message object pool
    this.createPool('message_objects', {
      factory: () => ({
        id: null,
        type: null,
        content: null,
        metadata: {},
        timestamp: null,
        sender: null,
        reset() {
          this.id = null;
          this.type = null;
          this.content = null;
          this.metadata = {};
          this.timestamp = null;
          this.sender = null;
        }
      }),
      maxSize: 100,
      preAllocate: 20
    });

    // WebSocket frame pool
    this.createPool('websocket_frames', {
      factory: () => ({
        opcode: null,
        payload: null,
        masked: false,
        fin: true,
        buffer: Buffer.alloc(1024), // Pre-allocated buffer
        reset() {
          this.opcode = null;
          this.payload = null;
          this.masked = false;
          this.fin = true;
          this.buffer.fill(0);
        }
      }),
      maxSize: 50,
      preAllocate: 10
    });

    // HTTP response pool
    this.createPool('http_responses', {
      factory: () => ({
        statusCode: null,
        headers: {},
        body: null,
        metadata: {},
        reset() {
          this.statusCode = null;
          this.headers = {};
          this.body = null;
          this.metadata = {};
        }
      }),
      maxSize: 30,
      preAllocate: 10
    });

    // Task execution context pool
    this.createPool('task_contexts', {
      factory: () => ({
        taskId: null,
        agentId: null,
        startTime: null,
        endTime: null,
        result: null,
        errors: [],
        performance: {},
        reset() {
          this.taskId = null;
          this.agentId = null;
          this.startTime = null;
          this.endTime = null;
          this.result = null;
          this.errors.length = 0;
          this.performance = {};
        }
      }),
      maxSize: 40,
      preAllocate: 8
    });

    console.log(`✅ Initialized ${this.pools.size} object pools`);
  }

  /**
   * Create a new object pool
   */
  createPool(name, config) {
    const pool = {
      name,
      objects: [],
      available: [],
      inUse: [],
      factory: config.factory,
      maxSize: config.maxSize || 50,
      created: 0,
      reused: 0,
      peakUsage: 0
    };

    // Pre-allocate objects
    const preAllocate = config.preAllocate || Math.min(5, pool.maxSize);
    for (let i = 0; i < preAllocate; i++) {
      const obj = pool.factory();
      pool.objects.push(obj);
      pool.available.push(obj);
      pool.created++;
    }

    this.pools.set(name, pool);
    this.metrics.pools_created++;
    
    console.log(`   Created pool '${name}': ${preAllocate} pre-allocated objects`);
  }

  /**
   * Get object from pool
   */
  getObject(poolName) {
    const pool = this.pools.get(poolName);
    if (!pool) {
      throw new Error(`Pool '${poolName}' not found`);
    }

    let obj;

    // Try to reuse available object
    if (pool.available.length > 0) {
      obj = pool.available.pop();
      pool.reused++;
      this.metrics.total_objects_reused++;
    } else if (pool.objects.length < pool.maxSize) {
      // Create new object if under limit
      obj = pool.factory();
      pool.objects.push(obj);
      pool.created++;
      this.metrics.total_objects_created++;
    } else {
      // Pool exhausted - force creation (should be rare)
      obj = pool.factory();
      this.metrics.total_objects_created++;
      console.warn(`⚠️ Pool '${poolName}' exhausted, creating object outside pool`);
    }

    pool.inUse.push(obj);
    pool.peakUsage = Math.max(pool.peakUsage, pool.inUse.length);

    return obj;
  }

  /**
   * Return object to pool
   */
  releaseObject(poolName, obj) {
    const pool = this.pools.get(poolName);
    if (!pool) return;

    // Remove from in-use list
    const inUseIndex = pool.inUse.indexOf(obj);
    if (inUseIndex > -1) {
      pool.inUse.splice(inUseIndex, 1);
    }

    // Reset object state
    if (obj.reset && typeof obj.reset === 'function') {
      obj.reset();
    }

    // Return to available pool if not over capacity
    if (pool.available.length < pool.maxSize / 2) {
      pool.available.push(obj);
    }
    // If pool has too many available objects, let this one be GC'd
  }

  /**
   * Get pool statistics
   */
  getPoolStats() {
    const stats = {};
    
    for (const [name, pool] of this.pools.entries()) {
      const reuseRate = pool.reused / (pool.created + pool.reused) * 100;
      const efficiency = pool.available.length / pool.objects.length * 100;
      
      stats[name] = {
        total_objects: pool.objects.length,
        available: pool.available.length,
        in_use: pool.inUse.length,
        reuse_rate: reuseRate.toFixed(2) + '%',
        efficiency: efficiency.toFixed(2) + '%',
        peak_usage: pool.peakUsage,
        created: pool.created,
        reused: pool.reused
      };
    }
    
    return stats;
  }

  /**
   * Calculate memory savings
   */
  calculateMemorySavings() {
    let totalSavings = 0;
    
    for (const [name, pool] of this.pools.entries()) {
      // Estimate object size (rough approximation)
      const estimatedObjectSize = this.estimateObjectSize(name);
      const objectsSaved = pool.reused;
      const memorySaved = (objectsSaved * estimatedObjectSize) / 1024 / 1024; // MB
      
      totalSavings += memorySaved;
    }
    
    this.metrics.memory_saved_mb = totalSavings;
    
    // Estimate GC pressure reduction
    const totalObjects = this.metrics.total_objects_created + this.metrics.total_objects_reused;
    const reuseRate = this.metrics.total_objects_reused / totalObjects * 100;
    this.metrics.gc_pressure_reduction = reuseRate;
    
    return {
      memory_saved_mb: totalSavings.toFixed(2),
      gc_pressure_reduction: reuseRate.toFixed(2) + '%',
      total_reuse_rate: reuseRate.toFixed(2) + '%'
    };
  }

  estimateObjectSize(poolName) {
    // Rough object size estimates in bytes
    const sizeEstimates = {
      'agent_instances': 2048,   // ~2KB per agent instance
      'message_objects': 512,    // ~512B per message
      'websocket_frames': 1024,  // ~1KB per frame
      'http_responses': 1024,    // ~1KB per response
      'task_contexts': 768       // ~768B per context
    };
    
    return sizeEstimates[poolName] || 512;
  }

  /**
   * Cleanup and shutdown
   */
  async cleanup() {
    console.log('🧹 Cleaning up memory pools...');
    
    for (const [name, pool] of this.pools.entries()) {
      pool.objects.length = 0;
      pool.available.length = 0;
      pool.inUse.length = 0;
    }
    
    this.pools.clear();
    
    // Force garbage collection if available
    if (global.gc) {
      global.gc();
      console.log('♻️ Forced garbage collection');
    }
  }

  /**
   * Generate performance report
   */
  generateReport() {
    const stats = this.getPoolStats();
    const savings = this.calculateMemorySavings();
    
    return {
      timestamp: new Date().toISOString(),
      pool_statistics: stats,
      memory_optimization: savings,
      overall_metrics: this.metrics,
      recommendations: [
        {
          category: 'pool_tuning',
          action: 'Adjust pool sizes based on usage patterns',
          priority: 'medium'
        },
        {
          category: 'gc_optimization', 
          action: 'Enable --expose-gc flag for manual GC control',
          priority: 'low'
        },
        {
          category: 'monitoring',
          action: 'Add real-time memory pool monitoring',
          priority: 'medium'
        }
      ]
    };
  }
}

// Export for use in other modules
module.exports = { MemoryPoolManager };

// CLI execution
if (require.main === module) {
  const poolManager = new MemoryPoolManager();
  
  // Simulate usage
  console.log('🧪 Testing memory pool performance...\n');
  
  // Test object lifecycle
  for (let i = 0; i < 100; i++) {
    const agent = poolManager.getObject('agent_instances');
    agent.id = `agent-${i}`;
    agent.type = 'test-agent';
    
    const message = poolManager.getObject('message_objects');
    message.id = `msg-${i}`;
    message.content = `Test message ${i}`;
    
    // Simulate some work
    setTimeout(() => {
      poolManager.releaseObject('agent_instances', agent);
      poolManager.releaseObject('message_objects', message);
    }, Math.random() * 1000);
  }
  
  // Generate report after test
  setTimeout(() => {
    console.log('\n📊 Memory Pool Performance Report:');
    const report = poolManager.generateReport();
    console.log(JSON.stringify(report, null, 2));
    
    poolManager.cleanup();
  }, 2000);
}