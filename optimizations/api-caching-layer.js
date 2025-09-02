#!/usr/bin/env node

/**
 * Redis-Based API Caching Layer for OllamaMax
 * Implements intelligent caching with TTL and invalidation strategies
 */

const Redis = require('redis');
const crypto = require('crypto');

class APICachingLayer {
  constructor(options = {}) {
    this.redis = null;
    this.config = {
      host: options.host || 'localhost',
      port: options.port || 13101,
      password: options.password || 'ollama_redis_pass',
      keyPrefix: 'ollama:cache:',
      defaultTTL: options.defaultTTL || 30, // 30 seconds
      maxMemory: options.maxMemory || '100mb'
    };
    
    this.cacheRules = new Map([
      ['/api/health', { ttl: 10, tags: ['health'] }],
      ['/api/nodes', { ttl: 30, tags: ['nodes', 'status'] }],
      ['/api/nodes/detailed', { ttl: 60, tags: ['nodes', 'detailed'] }],
      ['/api/models', { ttl: 300, tags: ['models', 'static'] }], // 5 minutes for models
      ['/api/nodes/health', { ttl: 15, tags: ['health', 'nodes'] }]
    ]);
    
    this.metrics = {
      hits: 0,
      misses: 0,
      sets: 0,
      invalidations: 0,
      errors: 0,
      avgResponseTime: 0,
      cacheEfficiency: 0
    };
  }

  /**
   * Initialize Redis connection and caching layer
   */
  async initialize() {
    console.log('🚀 Initializing API Caching Layer...');
    
    try {
      this.redis = Redis.createClient({
        host: this.config.host,
        port: this.config.port,
        password: this.config.password,
        retry_strategy: (options) => {
          if (options.error && options.error.code === 'ECONNREFUSED') {
            return new Error('Redis server connection refused');
          }
          if (options.total_retry_time > 1000 * 60 * 60) {
            return new Error('Retry time exhausted');
          }
          return Math.min(options.attempt * 100, 3000);
        }
      });

      await this.redis.connect();
      
      // Configure Redis for optimal caching
      await this.redis.configSet('maxmemory', this.config.maxMemory);
      await this.redis.configSet('maxmemory-policy', 'allkeys-lru');
      
      console.log('✅ Redis caching layer initialized');
      
      // Set up cache warming
      await this.warmCache();
      
    } catch (error) {
      console.error('❌ Failed to initialize caching layer:', error.message);
      throw error;
    }
  }

  /**
   * Middleware function for Express/HTTP servers
   */
  cacheMiddleware() {
    return async (req, res, next) => {
      // Only cache GET requests
      if (req.method !== 'GET') {
        return next();
      }

      const cacheKey = this.generateCacheKey(req);
      const startTime = Date.now();

      try {
        // Try to get from cache
        const cachedResponse = await this.get(cacheKey);
        
        if (cachedResponse) {
          // Cache hit
          this.metrics.hits++;
          this.updateResponseTimeMetric(Date.now() - startTime);
          
          const response = JSON.parse(cachedResponse);
          res.set(response.headers);
          res.status(response.statusCode).json(response.data);
          
          console.log(`💾 Cache HIT: ${req.path} (${Date.now() - startTime}ms)`);
          return;
        }

        // Cache miss - continue to actual handler
        this.metrics.misses++;
        
        // Intercept response to cache it
        const originalSend = res.send;
        const originalJson = res.json;
        
        res.send = (data) => {
          this.cacheResponse(cacheKey, req.path, res.statusCode, res.getHeaders(), data);
          return originalSend.call(res, data);
        };
        
        res.json = (data) => {
          this.cacheResponse(cacheKey, req.path, res.statusCode, res.getHeaders(), data);
          return originalJson.call(res, data);
        };
        
        console.log(`🔍 Cache MISS: ${req.path}`);
        next();
        
      } catch (error) {
        this.metrics.errors++;
        console.error(`Cache error for ${req.path}:`, error.message);
        next();
      }
    };
  }

  /**
   * Generate cache key for request
   */
  generateCacheKey(req) {
    const keyData = {
      path: req.path,
      query: req.query,
      headers: {
        'user-agent': req.get('user-agent'),
        'accept': req.get('accept')
      }
    };
    
    const hash = crypto.createHash('sha256')
      .update(JSON.stringify(keyData))
      .digest('hex');
    
    return `${this.config.keyPrefix}${hash}`;
  }

  /**
   * Cache API response
   */
  async cacheResponse(cacheKey, path, statusCode, headers, data) {
    try {
      const cacheRule = this.getCacheRule(path);
      if (!cacheRule) return; // Not cacheable
      
      const response = {
        statusCode,
        headers,
        data,
        cached: Date.now(),
        path
      };
      
      await this.set(cacheKey, JSON.stringify(response), cacheRule.ttl);
      
      // Tag for easy invalidation
      if (cacheRule.tags) {
        for (const tag of cacheRule.tags) {
          await this.tagCache(tag, cacheKey);
        }
      }
      
      this.metrics.sets++;
      console.log(`💾 Cached: ${path} (TTL: ${cacheRule.ttl}s)`);
      
    } catch (error) {
      this.metrics.errors++;
      console.error(`Failed to cache response for ${path}:`, error.message);
    }
  }

  /**
   * Get cache rule for path
   */
  getCacheRule(path) {
    for (const [pattern, rule] of this.cacheRules.entries()) {
      if (path.startsWith(pattern) || path.match(new RegExp(pattern))) {
        return rule;
      }
    }
    return null;
  }

  /**
   * Get value from cache
   */
  async get(key) {
    try {
      return await this.redis.get(key);
    } catch (error) {
      this.metrics.errors++;
      return null;
    }
  }

  /**
   * Set value in cache with TTL
   */
  async set(key, value, ttl = null) {
    try {
      const expiration = ttl || this.config.defaultTTL;
      await this.redis.setEx(key, expiration, value);
      return true;
    } catch (error) {
      this.metrics.errors++;
      return false;
    }
  }

  /**
   * Tag cache entries for group invalidation
   */
  async tagCache(tag, cacheKey) {
    const tagKey = `${this.config.keyPrefix}tag:${tag}`;
    await this.redis.sAdd(tagKey, cacheKey);
    await this.redis.expire(tagKey, 3600); // Tags expire in 1 hour
  }

  /**
   * Invalidate cache by tags
   */
  async invalidateByTag(tag) {
    try {
      const tagKey = `${this.config.keyPrefix}tag:${tag}`;
      const cacheKeys = await this.redis.sMembers(tagKey);
      
      if (cacheKeys.length > 0) {
        await this.redis.del(cacheKeys);
        await this.redis.del(tagKey);
        this.metrics.invalidations += cacheKeys.length;
        console.log(`🗑️ Invalidated ${cacheKeys.length} cache entries for tag: ${tag}`);
      }
      
    } catch (error) {
      this.metrics.errors++;
      console.error(`Failed to invalidate cache tag ${tag}:`, error.message);
    }
  }

  /**
   * Warm cache with frequently accessed endpoints
   */
  async warmCache() {
    console.log('🔥 Warming API cache...');
    
    const warmupEndpoints = [
      { path: '/api/health', delay: 0 },
      { path: '/api/nodes', delay: 1000 },
      { path: '/api/models', delay: 2000 }
    ];
    
    for (const endpoint of warmupEndpoints) {
      setTimeout(async () => {
        try {
          await this.preloadEndpoint(endpoint.path);
        } catch (error) {
          console.warn(`Cache warmup failed for ${endpoint.path}:`, error.message);
        }
      }, endpoint.delay);
    }
  }

  async preloadEndpoint(path) {
    // This would normally make an actual API call
    // For now, simulate successful warmup
    const mockResponse = {
      statusCode: 200,
      headers: { 'content-type': 'application/json' },
      data: { status: 'warmed', path },
      cached: Date.now(),
      path
    };
    
    const cacheKey = this.generateCacheKey({ path, query: {}, get: () => null });
    await this.set(cacheKey, JSON.stringify(mockResponse), 30);
    
    console.log(`🔥 Cache warmed: ${path}`);
  }

  /**
   * Update response time metrics
   */
  updateResponseTimeMetric(responseTime) {
    this.metrics.avgResponseTime = (this.metrics.avgResponseTime + responseTime) / 2;
  }

  /**
   * Get cache performance metrics
   */
  getMetrics() {
    const totalRequests = this.metrics.hits + this.metrics.misses;
    this.metrics.cacheEfficiency = totalRequests > 0 ? 
      (this.metrics.hits / totalRequests * 100) : 0;
    
    return {
      ...this.metrics,
      hitRate: this.metrics.cacheEfficiency.toFixed(2) + '%',
      totalRequests,
      errorRate: (this.metrics.errors / Math.max(1, totalRequests) * 100).toFixed(2) + '%'
    };
  }

  /**
   * Generate cache optimization report
   */
  async generateOptimizationReport() {
    const metrics = this.getMetrics();
    const redisInfo = await this.redis.info('memory');
    
    return {
      timestamp: new Date().toISOString(),
      cache_performance: metrics,
      redis_memory: this.parseRedisMemoryInfo(redisInfo),
      optimization_impact: {
        estimated_latency_reduction: `${Math.min(80, metrics.cacheEfficiency * 0.8).toFixed(1)}%`,
        server_load_reduction: `${Math.min(70, metrics.cacheEfficiency * 0.7).toFixed(1)}%`,
        bandwidth_savings: `${Math.min(60, metrics.cacheEfficiency * 0.6).toFixed(1)}%`
      },
      recommendations: this.generateCacheRecommendations(metrics)
    };
  }

  parseRedisMemoryInfo(info) {
    const lines = info.split('\r\n');
    const memInfo = {};
    
    for (const line of lines) {
      const [key, value] = line.split(':');
      if (key && value) {
        memInfo[key] = value;
      }
    }
    
    return {
      used_memory: memInfo.used_memory,
      used_memory_human: memInfo.used_memory_human,
      mem_fragmentation_ratio: memInfo.mem_fragmentation_ratio
    };
  }

  generateCacheRecommendations(metrics) {
    const recommendations = [];
    
    if (metrics.cacheEfficiency < 50) {
      recommendations.push({
        priority: 'high',
        issue: 'Low cache hit rate',
        action: 'Increase TTL for stable endpoints or improve cache warming',
        impact: 'High - Significant performance improvement potential'
      });
    }
    
    if (metrics.errorRate > 5) {
      recommendations.push({
        priority: 'medium',
        issue: 'High cache error rate',
        action: 'Investigate Redis connectivity and error handling',
        impact: 'Medium - Cache reliability improvement'
      });
    }
    
    if (metrics.avgResponseTime > 20) {
      recommendations.push({
        priority: 'medium',
        issue: 'Slow cache operations',
        action: 'Optimize Redis configuration or use local cache fallback',
        impact: 'Medium - Faster cache operations'
      });
    }
    
    return recommendations;
  }

  /**
   * Cleanup and disconnect
   */
  async cleanup() {
    if (this.redis) {
      await this.redis.disconnect();
      console.log('🔌 Redis connection closed');
    }
  }
}

module.exports = { APICachingLayer };

// CLI execution and testing
if (require.main === module) {
  const cache = new APICachingLayer();
  
  (async () => {
    try {
      await cache.initialize();
      
      // Simulate cache operations
      console.log('\n🧪 Testing cache operations...');
      
      // Test cache set/get
      await cache.set('test:key1', JSON.stringify({ test: 'data' }), 60);
      const result = await cache.get('test:key1');
      console.log('✅ Cache set/get test:', result ? 'PASSED' : 'FAILED');
      
      // Test tag invalidation
      await cache.tagCache('test_tag', 'test:key1');
      await cache.invalidateByTag('test_tag');
      const afterInvalidation = await cache.get('test:key1');
      console.log('✅ Tag invalidation test:', !afterInvalidation ? 'PASSED' : 'FAILED');
      
      // Generate report
      console.log('\n📊 Cache Performance Report:');
      const report = await cache.generateOptimizationReport();
      console.log(JSON.stringify(report, null, 2));
      
    } catch (error) {
      console.error('Cache testing failed:', error.message);
    } finally {
      await cache.cleanup();
    }
  })();
}