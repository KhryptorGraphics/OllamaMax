#!/usr/bin/env node

/**
 * Agent Metrics Collection Pipeline
 * Collects and aggregates agent performance metrics for the enhanced system
 */

const Redis = require('ioredis');
const { performance } = require('perf_hooks');
const EventEmitter = require('events');
const fs = require('fs').promises;
const path = require('path');

class AgentMetricsCollector extends EventEmitter {
  constructor(config = {}) {
    super();
    
    // Redis cluster configuration
    this.redisCluster = new Redis.Cluster([
      { host: process.env.REDIS_NODE_1 || 'redis-cluster-0.redis-cluster-service.ollamamax-redis', port: 6379 },
      { host: process.env.REDIS_NODE_2 || 'redis-cluster-1.redis-cluster-service.ollamamax-redis', port: 6379 },
      { host: process.env.REDIS_NODE_3 || 'redis-cluster-2.redis-cluster-service.ollamamax-redis', port: 6379 },
      { host: process.env.REDIS_NODE_4 || 'redis-cluster-3.redis-cluster-service.ollamamax-redis', port: 6379 },
      { host: process.env.REDIS_NODE_5 || 'redis-cluster-4.redis-cluster-service.ollamamax-redis', port: 6379 },
      { host: process.env.REDIS_NODE_6 || 'redis-cluster-5.redis-cluster-service.ollamamax-redis', port: 6379 }
    ], {
      redisOptions: {
        password: process.env.REDIS_PASSWORD,
        retryDelayOnFailover: 100,
        maxRetriesPerRequest: 3
      },
      enableOfflineQueue: false,
      retryDelayOnClusterDown: 300,
      retryDelayOnFailover: 100,
      maxRetriesPerRequest: 3,
      scaleReads: 'slave'
    });

    // Configuration
    this.config = {
      batchSize: config.batchSize || 100,
      flushInterval: config.flushInterval || 5000,  // 5 seconds
      retentionPeriods: {
        realtime: 3600,      // 1 hour
        hourly: 86400,       // 24 hours  
        daily: 2592000,      // 30 days
        monthly: 31536000    // 1 year
      },
      ...config
    };

    // Metrics buffer
    this.metricsBuffer = [];
    this.aggregationBuffer = new Map();
    
    // Performance tracking
    this.collectorStats = {
      totalMetrics: 0,
      batchesProcessed: 0,
      errors: 0,
      avgProcessingTime: 0,
      lastFlush: Date.now()
    };

    // Initialize collector
    this.initialize();
  }

  async initialize() {
    console.log('🚀 Initializing Agent Metrics Collector...');
    
    try {
      // Test Redis cluster connection
      await this.testRedisConnection();
      
      // Setup periodic flush
      this.setupPeriodicFlush();
      
      // Setup aggregation pipeline
      this.setupAggregation();
      
      // Setup cleanup tasks
      this.setupCleanup();
      
      console.log('✅ Agent Metrics Collector initialized successfully');
      this.emit('initialized');
      
    } catch (error) {
      console.error('❌ Failed to initialize metrics collector:', error);
      this.emit('error', error);
    }
  }

  async testRedisConnection() {
    try {
      await this.redisCluster.ping();
      console.log('✅ Redis cluster connection successful');
      
      // Test cluster info
      const clusterInfo = await this.redisCluster.cluster('info');
      console.log(`📊 Redis cluster status: ${clusterInfo.includes('cluster_state:ok') ? 'OK' : 'DEGRADED'}`);
      
    } catch (error) {
      throw new Error(`Redis cluster connection failed: ${error.message}`);
    }
  }

  /**
   * Collect agent execution metrics
   */
  async collectAgentMetric(agentId, taskId, metrics) {
    const timestamp = Date.now();
    
    const metric = {
      agentId,
      taskId,
      timestamp,
      execution: {
        duration: metrics.duration || 0,
        success: metrics.success || false,
        errorType: metrics.errorType || null,
        retryCount: metrics.retryCount || 0
      },
      resources: {
        cpu: metrics.cpu || 0,
        memory: metrics.memory || 0,
        concurrent: metrics.concurrent || 0
      },
      quality: {
        successRate: metrics.successRate || 0,
        errorRate: metrics.errorRate || 0,
        performanceScore: metrics.performanceScore || 0
      },
      context: {
        topology: metrics.topology || 'unknown',
        swarmSize: metrics.swarmSize || 0,
        taskType: metrics.taskType || 'general'
      }
    };

    // Add to buffer
    this.metricsBuffer.push(metric);
    this.collectorStats.totalMetrics++;

    // Trigger immediate flush if buffer is full
    if (this.metricsBuffer.length >= this.config.batchSize) {
      await this.flushMetrics();
    }

    // Update aggregation data
    this.updateAggregation(metric);

    return metric;
  }

  /**
   * Collect swarm-level metrics
   */
  async collectSwarmMetric(metrics) {
    const timestamp = Date.now();
    
    const swarmMetric = {
      timestamp,
      topology: metrics.topology,
      activeAgents: metrics.activeAgents || 0,
      queueLength: metrics.queueLength || 0,
      utilization: metrics.utilization || 0,
      throughput: metrics.throughput || 0,
      avgResponseTime: metrics.avgResponseTime || 0,
      errorRate: metrics.errorRate || 0,
      scalingEvents: metrics.scalingEvents || 0
    };

    // Store swarm metrics with shorter TTL
    const key = `swarm:metrics:${timestamp}`;
    await this.redisCluster.hmset(key, swarmMetric);
    await this.redisCluster.expire(key, this.config.retentionPeriods.realtime);

    // Store in time series for trending
    await this.storeTimeSeries('swarm_utilization', metrics.utilization, timestamp);
    await this.storeTimeSeries('swarm_throughput', metrics.throughput, timestamp);
    await this.storeTimeSeries('swarm_response_time', metrics.avgResponseTime, timestamp);

    return swarmMetric;
  }

  /**
   * Update real-time aggregations
   */
  updateAggregation(metric) {
    const agentKey = `agent:${metric.agentId}`;
    
    if (!this.aggregationBuffer.has(agentKey)) {
      this.aggregationBuffer.set(agentKey, {
        agentId: metric.agentId,
        totalTasks: 0,
        successfulTasks: 0,
        totalDuration: 0,
        totalCpu: 0,
        totalMemory: 0,
        errors: [],
        lastUpdate: Date.now()
      });
    }

    const agg = this.aggregationBuffer.get(agentKey);
    agg.totalTasks++;
    agg.totalDuration += metric.execution.duration;
    agg.totalCpu += metric.resources.cpu;
    agg.totalMemory += metric.resources.memory;
    agg.lastUpdate = Date.now();

    if (metric.execution.success) {
      agg.successfulTasks++;
    } else {
      agg.errors.push({
        type: metric.execution.errorType,
        timestamp: metric.timestamp
      });
    }
  }

  /**
   * Flush metrics buffer to Redis
   */
  async flushMetrics() {
    if (this.metricsBuffer.length === 0) return;

    const startTime = performance.now();
    const batch = [...this.metricsBuffer];
    this.metricsBuffer = [];

    try {
      const pipeline = this.redisCluster.pipeline();

      // Store individual metrics
      for (const metric of batch) {
        const key = `agent:${metric.agentId}:metrics:${metric.timestamp}`;
        pipeline.hmset(key, {
          taskId: metric.taskId,
          timestamp: metric.timestamp,
          duration: metric.execution.duration,
          success: metric.execution.success ? 1 : 0,
          cpu: metric.resources.cpu,
          memory: metric.resources.memory,
          topology: metric.context.topology,
          taskType: metric.context.taskType
        });
        pipeline.expire(key, this.config.retentionPeriods.hourly);

        // Add to sorted sets for time-based queries
        pipeline.zadd(`agent:${metric.agentId}:timeline`, metric.timestamp, key);
        pipeline.zadd('global:metrics:timeline', metric.timestamp, key);
      }

      // Store aggregated data
      for (const [agentKey, agg] of this.aggregationBuffer) {
        const avgDuration = agg.totalTasks > 0 ? agg.totalDuration / agg.totalTasks : 0;
        const avgCpu = agg.totalTasks > 0 ? agg.totalCpu / agg.totalTasks : 0;
        const avgMemory = agg.totalTasks > 0 ? agg.totalMemory / agg.totalTasks : 0;
        const successRate = agg.totalTasks > 0 ? agg.successfulTasks / agg.totalTasks : 0;

        const aggKey = `${agentKey}:aggregated`;
        pipeline.hmset(aggKey, {
          totalTasks: agg.totalTasks,
          successfulTasks: agg.successfulTasks,
          avgDuration,
          avgCpu,
          avgMemory,
          successRate,
          errorCount: agg.errors.length,
          lastUpdate: agg.lastUpdate
        });
        pipeline.expire(aggKey, this.config.retentionPeriods.daily);
      }

      await pipeline.exec();

      // Update collector stats
      const processingTime = performance.now() - startTime;
      this.collectorStats.batchesProcessed++;
      this.collectorStats.avgProcessingTime = 
        (this.collectorStats.avgProcessingTime * (this.collectorStats.batchesProcessed - 1) + processingTime) 
        / this.collectorStats.batchesProcessed;
      this.collectorStats.lastFlush = Date.now();

      console.log(`📊 Flushed ${batch.length} metrics in ${processingTime.toFixed(2)}ms`);
      
      // Clear aggregation buffer
      this.aggregationBuffer.clear();

    } catch (error) {
      console.error('❌ Error flushing metrics:', error);
      this.collectorStats.errors++;
      
      // Return metrics to buffer for retry
      this.metricsBuffer.unshift(...batch);
      throw error;
    }
  }

  /**
   * Store time series data for trending
   */
  async storeTimeSeries(metricName, value, timestamp = Date.now()) {
    const minute = Math.floor(timestamp / 60000) * 60000;
    const hour = Math.floor(timestamp / 3600000) * 3600000;
    const day = Math.floor(timestamp / 86400000) * 86400000;

    const pipeline = this.redisCluster.pipeline();

    // Store at different resolutions
    pipeline.zadd(`ts:${metricName}:1m`, timestamp, `${timestamp}:${value}`);
    pipeline.zadd(`ts:${metricName}:1h`, hour, `${hour}:${value}`);
    pipeline.zadd(`ts:${metricName}:1d`, day, `${day}:${value}`);

    // Set expiration for different resolutions
    pipeline.expire(`ts:${metricName}:1m`, this.config.retentionPeriods.realtime);
    pipeline.expire(`ts:${metricName}:1h`, this.config.retentionPeriods.hourly);
    pipeline.expire(`ts:${metricName}:1d`, this.config.retentionPeriods.daily);

    await pipeline.exec();
  }

  /**
   * Get agent performance metrics
   */
  async getAgentMetrics(agentId, timeRange = '1h') {
    const now = Date.now();
    const ranges = {
      '1h': 3600000,
      '6h': 21600000,
      '24h': 86400000,
      '7d': 604800000
    };
    const startTime = now - (ranges[timeRange] || ranges['1h']);

    // Get aggregated data
    const aggKey = `agent:${agentId}:aggregated`;
    const aggregated = await this.redisCluster.hgetall(aggKey);

    // Get recent metrics
    const timelineKey = `agent:${agentId}:timeline`;
    const recentKeys = await this.redisCluster.zrangebyscore(timelineKey, startTime, now);
    
    const recentMetrics = [];
    if (recentKeys.length > 0) {
      const pipeline = this.redisCluster.pipeline();
      recentKeys.forEach(key => pipeline.hgetall(key));
      const results = await pipeline.exec();
      
      results.forEach((result, index) => {
        if (result[1]) {
          recentMetrics.push({
            key: recentKeys[index],
            ...result[1]
          });
        }
      });
    }

    return {
      agentId,
      timeRange,
      aggregated: {
        totalTasks: parseInt(aggregated.totalTasks) || 0,
        successfulTasks: parseInt(aggregated.successfulTasks) || 0,
        avgDuration: parseFloat(aggregated.avgDuration) || 0,
        avgCpu: parseFloat(aggregated.avgCpu) || 0,
        avgMemory: parseFloat(aggregated.avgMemory) || 0,
        successRate: parseFloat(aggregated.successRate) || 0,
        errorCount: parseInt(aggregated.errorCount) || 0,
        lastUpdate: parseInt(aggregated.lastUpdate) || 0
      },
      recentMetrics,
      collectorStats: this.collectorStats
    };
  }

  /**
   * Get time series data
   */
  async getTimeSeries(metricName, resolution = '1m', timeRange = '1h') {
    const now = Date.now();
    const ranges = {
      '1h': 3600000,
      '6h': 21600000, 
      '24h': 86400000,
      '7d': 604800000
    };
    const startTime = now - (ranges[timeRange] || ranges['1h']);

    const key = `ts:${metricName}:${resolution}`;
    const data = await this.redisCluster.zrangebyscore(key, startTime, now, 'WITHSCORES');
    
    const timeSeries = [];
    for (let i = 0; i < data.length; i += 2) {
      const [timestamp, value] = data[i].split(':');
      timeSeries.push({
        timestamp: parseInt(timestamp),
        value: parseFloat(value),
        score: parseFloat(data[i + 1])
      });
    }

    return {
      metric: metricName,
      resolution,
      timeRange,
      data: timeSeries
    };
  }

  /**
   * Get system health metrics
   */
  async getSystemHealth() {
    const clusterInfo = await this.redisCluster.cluster('info');
    const clusterNodes = await this.redisCluster.cluster('nodes');
    
    // Parse cluster status
    const isHealthy = clusterInfo.includes('cluster_state:ok');
    const nodeCount = clusterNodes.split('\n').filter(line => line.trim()).length;
    
    return {
      redis: {
        healthy: isHealthy,
        nodes: nodeCount,
        info: clusterInfo
      },
      collector: this.collectorStats,
      bufferSize: this.metricsBuffer.length,
      aggregationSize: this.aggregationBuffer.size,
      uptime: Date.now() - this.collectorStats.lastFlush
    };
  }

  /**
   * Setup periodic flush
   */
  setupPeriodicFlush() {
    setInterval(async () => {
      try {
        await this.flushMetrics();
      } catch (error) {
        console.error('❌ Periodic flush error:', error);
      }
    }, this.config.flushInterval);
  }

  /**
   * Setup aggregation pipeline
   */
  setupAggregation() {
    // Real-time aggregation every 30 seconds
    setInterval(async () => {
      try {
        await this.computeRealTimeAggregations();
      } catch (error) {
        console.error('❌ Aggregation error:', error);
      }
    }, 30000);
  }

  /**
   * Compute real-time aggregations
   */
  async computeRealTimeAggregations() {
    const now = Date.now();
    const pipeline = this.redisCluster.pipeline();

    // Compute global metrics
    const globalKeys = await this.redisCluster.zrangebyscore('global:metrics:timeline', now - 300000, now);
    
    if (globalKeys.length > 0) {
      // Get all metrics for the last 5 minutes
      const metricsData = await Promise.all(
        globalKeys.map(key => this.redisCluster.hgetall(key))
      );

      const validMetrics = metricsData.filter(m => m && Object.keys(m).length > 0);
      
      if (validMetrics.length > 0) {
        const globalStats = {
          totalTasks: validMetrics.length,
          successfulTasks: validMetrics.filter(m => m.success === '1').length,
          avgDuration: validMetrics.reduce((sum, m) => sum + parseFloat(m.duration || 0), 0) / validMetrics.length,
          avgCpu: validMetrics.reduce((sum, m) => sum + parseFloat(m.cpu || 0), 0) / validMetrics.length,
          avgMemory: validMetrics.reduce((sum, m) => sum + parseFloat(m.memory || 0), 0) / validMetrics.length,
          timestamp: now
        };

        globalStats.successRate = globalStats.totalTasks > 0 ? globalStats.successfulTasks / globalStats.totalTasks : 0;

        pipeline.hmset('global:stats:realtime', globalStats);
        pipeline.expire('global:stats:realtime', 300);
      }
    }

    await pipeline.exec();
  }

  /**
   * Setup cleanup tasks
   */
  setupCleanup() {
    // Cleanup old data every hour
    setInterval(async () => {
      try {
        await this.cleanupOldData();
      } catch (error) {
        console.error('❌ Cleanup error:', error);
      }
    }, 3600000); // 1 hour
  }

  /**
   * Cleanup old data based on retention policies
   */
  async cleanupOldData() {
    const now = Date.now();
    
    // Cleanup timeline entries older than retention periods
    const timelineKeys = await this.redisCluster.keys('*:timeline');
    
    for (const key of timelineKeys) {
      const cutoff = now - this.config.retentionPeriods.daily;
      await this.redisCluster.zremrangebyscore(key, 0, cutoff);
    }

    console.log(`🧹 Cleaned up old data from ${timelineKeys.length} timeline keys`);
  }

  /**
   * Graceful shutdown
   */
  async shutdown() {
    console.log('🛑 Shutting down metrics collector...');
    
    // Flush remaining metrics
    if (this.metricsBuffer.length > 0) {
      await this.flushMetrics();
    }
    
    // Close Redis connections
    await this.redisCluster.disconnect();
    
    console.log('✅ Metrics collector shut down successfully');
  }
}

module.exports = AgentMetricsCollector;

// Export singleton instance for global use
module.exports.instance = new AgentMetricsCollector();

// Handle graceful shutdown
process.on('SIGTERM', async () => {
  await module.exports.instance.shutdown();
  process.exit(0);
});

process.on('SIGINT', async () => {
  await module.exports.instance.shutdown();
  process.exit(0);
});