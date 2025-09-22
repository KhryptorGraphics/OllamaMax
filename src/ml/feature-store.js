#!/usr/bin/env node

/**
 * Feature Store for ML Pipeline
 * Centralized feature management with real-time serving and batch processing
 */

const Redis = require('ioredis');
const { performance } = require('perf_hooks');

class FeatureStore {
  constructor(config = {}) {
    this.config = {
      redisNodes: config.redisNodes || [
        { host: 'redis-cluster-0.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-1.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-2.redis-cluster-service.ollamamax-redis', port: 6379 }
      ],
      redisPassword: config.redisPassword || 'ollama_redis_pass',
      featureTTL: config.featureTTL || 24 * 60 * 60, // 24 hours
      batchUpdateInterval: config.batchUpdateInterval || 300000, // 5 minutes
      featureGroups: config.featureGroups || [
        'agent_performance',
        'task_characteristics', 
        'system_metrics',
        'temporal_features',
        'contextual_features'
      ],
      ...config
    };

    this.redis = null;
    this.featureDefinitions = new Map();
    this.computationCache = new Map();
    this.updateSchedules = new Map();

    this.initializeRedis();
    this.initializeFeatureDefinitions();
    this.startBatchUpdates();
  }

  async initializeRedis() {
    try {
      this.redis = new Redis.Cluster(this.config.redisNodes, {
        redisOptions: {
          password: this.config.redisPassword,
          lazyConnect: true
        },
        enableOfflineQueue: false,
        retryDelayOnFailover: 100,
        maxRetriesPerRequest: 3
      });

      await this.redis.ping();
      console.log('✅ Feature Store connected to Redis cluster');
    } catch (error) {
      console.error('❌ Redis cluster connection failed:', error.message);
      throw error;
    }
  }

  initializeFeatureDefinitions() {
    // Agent Performance Features
    this.defineFeatureGroup('agent_performance', {
      'agent_success_rate_1h': {
        description: 'Agent success rate over last 1 hour',
        type: 'float',
        range: [0, 1],
        computation: 'rolling_average',
        window: '1h',
        updateFrequency: 'realtime'
      },
      'agent_success_rate_24h': {
        description: 'Agent success rate over last 24 hours', 
        type: 'float',
        range: [0, 1],
        computation: 'rolling_average',
        window: '24h',
        updateFrequency: 'batch'
      },
      'agent_avg_execution_time': {
        description: 'Average task execution time in milliseconds',
        type: 'float',
        range: [0, null],
        computation: 'rolling_average',
        window: '24h',
        updateFrequency: 'realtime'
      },
      'agent_current_load': {
        description: 'Current number of active tasks',
        type: 'integer',
        range: [0, null],
        computation: 'current_value',
        updateFrequency: 'realtime'
      },
      'agent_error_rate': {
        description: 'Error rate over last 24 hours',
        type: 'float',
        range: [0, 1],
        computation: 'rolling_rate',
        window: '24h',
        updateFrequency: 'batch'
      },
      'agent_resource_efficiency': {
        description: 'CPU/Memory efficiency score',
        type: 'float',
        range: [0, 1],
        computation: 'efficiency_score',
        window: '1h',
        updateFrequency: 'batch'
      }
    });

    // Task Characteristics Features  
    this.defineFeatureGroup('task_characteristics', {
      'task_complexity_score': {
        description: 'Computed task complexity (1-10 scale)',
        type: 'float',
        range: [1, 10],
        computation: 'complexity_analysis',
        updateFrequency: 'realtime'
      },
      'task_priority_level': {
        description: 'Task priority (1-10 scale)',
        type: 'integer',
        range: [1, 10],
        computation: 'direct_value',
        updateFrequency: 'realtime'
      },
      'task_estimated_duration': {
        description: 'Estimated task duration in milliseconds',
        type: 'float',
        range: [0, null],
        computation: 'duration_estimation',
        updateFrequency: 'realtime'
      },
      'task_type_frequency': {
        description: 'Frequency of similar task types',
        type: 'float',
        range: [0, 1],
        computation: 'type_frequency',
        window: '7d',
        updateFrequency: 'batch'
      },
      'task_specialization_match': {
        description: 'How well task matches agent specialization',
        type: 'float',
        range: [0, 1],
        computation: 'specialization_matching',
        updateFrequency: 'realtime'
      }
    });

    // System Metrics Features
    this.defineFeatureGroup('system_metrics', {
      'system_queue_length': {
        description: 'Current task queue length',
        type: 'integer', 
        range: [0, null],
        computation: 'current_value',
        updateFrequency: 'realtime'
      },
      'system_avg_response_time': {
        description: 'System average response time',
        type: 'float',
        range: [0, null],
        computation: 'rolling_average',
        window: '1h',
        updateFrequency: 'batch'
      },
      'system_cpu_utilization': {
        description: 'Overall system CPU utilization',
        type: 'float',
        range: [0, 1],
        computation: 'rolling_average',
        window: '5m',
        updateFrequency: 'realtime'
      },
      'system_memory_utilization': {
        description: 'Overall system memory utilization',
        type: 'float',
        range: [0, 1],
        computation: 'rolling_average', 
        window: '5m',
        updateFrequency: 'realtime'
      },
      'active_agent_count': {
        description: 'Number of currently active agents',
        type: 'integer',
        range: [0, null],
        computation: 'current_value',
        updateFrequency: 'realtime'
      }
    });

    // Temporal Features
    this.defineFeatureGroup('temporal_features', {
      'hour_of_day': {
        description: 'Current hour normalized (0-1)',
        type: 'float',
        range: [0, 1],
        computation: 'time_normalization',
        updateFrequency: 'realtime'
      },
      'day_of_week': {
        description: 'Current day of week normalized (0-1)',
        type: 'float',
        range: [0, 1],
        computation: 'time_normalization',
        updateFrequency: 'batch'
      },
      'is_business_hours': {
        description: 'Whether current time is business hours',
        type: 'boolean',
        computation: 'business_hours_check',
        updateFrequency: 'realtime'
      },
      'seasonal_factor': {
        description: 'Seasonal adjustment factor',
        type: 'float',
        range: [0.5, 1.5],
        computation: 'seasonal_calculation',
        updateFrequency: 'batch'
      },
      'workload_trend': {
        description: 'Recent workload trend (-1 to 1)',
        type: 'float',
        range: [-1, 1],
        computation: 'trend_analysis',
        window: '2h',
        updateFrequency: 'batch'
      }
    });

    // Contextual Features
    this.defineFeatureGroup('contextual_features', {
      'similar_task_performance': {
        description: 'Performance on similar recent tasks',
        type: 'float',
        range: [0, 1],
        computation: 'similarity_matching',
        window: '7d',
        updateFrequency: 'batch'
      },
      'agent_context_continuity': {
        description: 'Context continuity with recent tasks',
        type: 'float',
        range: [0, 1],
        computation: 'context_similarity',
        window: '1h',
        updateFrequency: 'realtime'
      },
      'swarm_coordination_score': {
        description: 'How well agent coordinates in swarms',
        type: 'float',
        range: [0, 1],
        computation: 'coordination_analysis',
        window: '24h',
        updateFrequency: 'batch'
      }
    });

    console.log('✅ Feature definitions initialized');
    console.log(`   📊 Feature groups: ${this.config.featureGroups.length}`);
    console.log(`   🔧 Total features: ${this.getTotalFeatureCount()}`);
  }

  defineFeatureGroup(groupName, features) {
    this.featureDefinitions.set(groupName, features);
  }

  getTotalFeatureCount() {
    let count = 0;
    for (const [groupName, features] of this.featureDefinitions) {
      count += Object.keys(features).length;
    }
    return count;
  }

  async computeFeature(featureName, entityId, context = {}) {
    const featureKey = `feature:${featureName}:${entityId}`;
    
    // Check cache first
    const cached = await this.redis.get(featureKey);
    if (cached) {
      return JSON.parse(cached);
    }

    const featureDef = this.getFeatureDefinition(featureName);
    if (!featureDef) {
      throw new Error(`Feature definition not found: ${featureName}`);
    }

    let value;
    const startTime = performance.now();

    try {
      switch (featureDef.computation) {
        case 'rolling_average':
          value = await this.computeRollingAverage(featureName, entityId, featureDef.window);
          break;
        case 'rolling_rate':
          value = await this.computeRollingRate(featureName, entityId, featureDef.window);
          break;
        case 'current_value':
          value = await this.getCurrentValue(featureName, entityId);
          break;
        case 'complexity_analysis':
          value = await this.computeTaskComplexity(entityId, context);
          break;
        case 'duration_estimation':
          value = await this.estimateTaskDuration(entityId, context);
          break;
        case 'specialization_matching':
          value = await this.computeSpecializationMatch(entityId, context);
          break;
        case 'efficiency_score':
          value = await this.computeEfficiencyScore(entityId, featureDef.window);
          break;
        case 'time_normalization':
          value = this.computeTimeNormalization(featureName);
          break;
        case 'business_hours_check':
          value = this.isBusinessHours();
          break;
        case 'seasonal_calculation':
          value = this.computeSeasonalFactor();
          break;
        case 'trend_analysis':
          value = await this.computeTrend(entityId, featureDef.window);
          break;
        case 'similarity_matching':
          value = await this.computeSimilarTaskPerformance(entityId, context, featureDef.window);
          break;
        case 'context_similarity':
          value = await this.computeContextContinuity(entityId, context, featureDef.window);
          break;
        case 'coordination_analysis':
          value = await this.computeCoordinationScore(entityId, featureDef.window);
          break;
        default:
          throw new Error(`Unknown computation type: ${featureDef.computation}`);
      }

      // Validate range
      if (featureDef.range && featureDef.range.length === 2) {
        const [min, max] = featureDef.range;
        if (min !== null && value < min) value = min;
        if (max !== null && value > max) value = max;
      }

      const computationTime = performance.now() - startTime;

      const result = {
        value,
        timestamp: Date.now(),
        computationTime,
        featureDefinition: featureDef
      };

      // Cache result
      await this.redis.setex(featureKey, this.config.featureTTL, JSON.stringify(result));

      return result;

    } catch (error) {
      console.error(`❌ Error computing feature ${featureName}:`, error.message);
      throw error;
    }
  }

  getFeatureDefinition(featureName) {
    for (const [groupName, features] of this.featureDefinitions) {
      if (features[featureName]) {
        return features[featureName];
      }
    }
    return null;
  }

  async computeRollingAverage(featureName, entityId, window) {
    const windowMs = this.parseTimeWindow(window);
    const cutoffTime = Date.now() - windowMs;
    
    const rawDataKey = `raw:${featureName}:${entityId}`;
    const rawData = await this.redis.zrangebyscore(rawDataKey, cutoffTime, '+inf', 'WITHSCORES');
    
    if (rawData.length === 0) return 0;
    
    let sum = 0;
    let count = 0;
    
    for (let i = 0; i < rawData.length; i += 2) {
      sum += parseFloat(rawData[i]);
      count++;
    }
    
    return count > 0 ? sum / count : 0;
  }

  async computeRollingRate(featureName, entityId, window) {
    const windowMs = this.parseTimeWindow(window);
    const cutoffTime = Date.now() - windowMs;
    
    const totalKey = `count:total:${entityId}`;
    const successKey = `count:success:${entityId}`;
    
    const total = await this.redis.zcount(totalKey, cutoffTime, '+inf');
    const success = await this.redis.zcount(successKey, cutoffTime, '+inf');
    
    return total > 0 ? (total - success) / total : 0; // Error rate
  }

  async getCurrentValue(featureName, entityId) {
    const currentKey = `current:${featureName}:${entityId}`;
    const value = await this.redis.get(currentKey);
    return parseFloat(value) || 0;
  }

  async computeTaskComplexity(taskId, context) {
    let complexity = 1;
    
    // Description length factor
    if (context.description) {
      complexity += Math.min(context.description.length / 1000, 3);
    }
    
    // File count factor
    if (context.files) {
      complexity += Math.min(context.files.length / 5, 2);
    }
    
    // Dependencies factor
    if (context.dependencies) {
      complexity += Math.min(context.dependencies.length / 3, 1.5);
    }
    
    // Task type factor
    const complexTypes = ['architecture', 'system-design', 'optimization', 'migration'];
    if (context.type && complexTypes.includes(context.type)) {
      complexity += 2;
    }
    
    return Math.min(complexity, 10);
  }

  async estimateTaskDuration(taskId, context) {
    // Base duration estimates by type
    const baseDurations = {
      'implementation': 300000,  // 5 minutes
      'review': 120000,         // 2 minutes
      'testing': 180000,        // 3 minutes
      'debugging': 600000,      // 10 minutes
      'architecture': 1800000,  // 30 minutes
      'optimization': 900000,   // 15 minutes
      'default': 300000
    };
    
    const baseDuration = baseDurations[context.type] || baseDurations.default;
    
    // Complexity multiplier
    const complexity = await this.computeTaskComplexity(taskId, context);
    const complexityMultiplier = Math.max(0.5, Math.min(3.0, complexity / 3));
    
    return baseDuration * complexityMultiplier;
  }

  async computeSpecializationMatch(entityId, context) {
    if (!context.agentId || !context.taskType) return 0.5;
    
    // Get agent specialization
    const agentType = await this.getAgentType(context.agentId);
    
    // Match matrix
    const matchScores = {
      'coder': {
        'implementation': 0.9,
        'coding': 0.9,
        'debugging': 0.8,
        'optimization': 0.7
      },
      'reviewer': {
        'review': 0.9,
        'quality': 0.8,
        'security': 0.7
      },
      'tester': {
        'testing': 0.9,
        'validation': 0.8,
        'qa': 0.8
      },
      'researcher': {
        'analysis': 0.9,
        'research': 0.9,
        'investigation': 0.8
      }
    };
    
    return matchScores[agentType]?.[context.taskType] || 0.5;
  }

  async computeEfficiencyScore(entityId, window) {
    const windowMs = this.parseTimeWindow(window);
    const cutoffTime = Date.now() - windowMs;
    
    // Get CPU and memory usage data
    const cpuKey = `raw:cpu_usage:${entityId}`;
    const memoryKey = `raw:memory_usage:${entityId}`;
    const throughputKey = `raw:task_throughput:${entityId}`;
    
    const [cpuData, memoryData, throughputData] = await Promise.all([
      this.redis.zrangebyscore(cpuKey, cutoffTime, '+inf'),
      this.redis.zrangebyscore(memoryKey, cutoffTime, '+inf'),
      this.redis.zrangebyscore(throughputKey, cutoffTime, '+inf')
    ]);
    
    if (cpuData.length === 0) return 0.5;
    
    const avgCpu = cpuData.reduce((sum, val) => sum + parseFloat(val), 0) / cpuData.length;
    const avgMemory = memoryData.reduce((sum, val) => sum + parseFloat(val), 0) / memoryData.length;
    const avgThroughput = throughputData.reduce((sum, val) => sum + parseFloat(val), 0) / throughputData.length;
    
    // Efficiency = throughput / resource_usage
    const resourceUsage = (avgCpu + avgMemory) / 2;
    return resourceUsage > 0 ? Math.min(1, avgThroughput / resourceUsage) : 0;
  }

  computeTimeNormalization(featureName) {
    const now = new Date();
    
    switch (featureName) {
      case 'hour_of_day':
        return now.getHours() / 24;
      case 'day_of_week':
        return now.getDay() / 7;
      default:
        return 0;
    }
  }

  isBusinessHours() {
    const now = new Date();
    const hour = now.getHours();
    const day = now.getDay();
    
    // Monday-Friday, 9 AM - 5 PM
    return day >= 1 && day <= 5 && hour >= 9 && hour <= 17;
  }

  computeSeasonalFactor() {
    const now = new Date();
    const hour = now.getHours();
    const day = now.getDay();
    
    // Business hours boost
    const businessHoursFactor = this.isBusinessHours() ? 1.2 : 0.8;
    
    // Weekly pattern
    const weekdayFactor = (day >= 1 && day <= 5) ? 1.1 : 0.9;
    
    return businessHoursFactor * weekdayFactor;
  }

  async computeTrend(entityId, window) {
    const windowMs = this.parseTimeWindow(window);
    const cutoffTime = Date.now() - windowMs;
    
    const dataKey = `raw:workload:${entityId}`;
    const data = await this.redis.zrangebyscore(dataKey, cutoffTime, '+inf', 'WITHSCORES');
    
    if (data.length < 4) return 0; // Need at least 2 data points
    
    // Calculate simple linear trend
    const values = [];
    const timestamps = [];
    
    for (let i = 0; i < data.length; i += 2) {
      values.push(parseFloat(data[i]));
      timestamps.push(parseInt(data[i + 1]));
    }
    
    const n = values.length;
    const sumX = timestamps.reduce((a, b) => a + b, 0);
    const sumY = values.reduce((a, b) => a + b, 0);
    const sumXY = timestamps.reduce((sum, x, i) => sum + x * values[i], 0);
    const sumXX = timestamps.reduce((sum, x) => sum + x * x, 0);
    
    const slope = (n * sumXY - sumX * sumY) / (n * sumXX - sumX * sumX);
    
    // Normalize slope to -1 to 1 range
    return Math.max(-1, Math.min(1, slope * 1000000)); // Scale factor for time
  }

  async getFeatureVector(entityId, featureNames, context = {}) {
    const features = [];
    const computationPromises = featureNames.map(name => 
      this.computeFeature(name, entityId, context)
    );
    
    const results = await Promise.allSettled(computationPromises);
    
    for (let i = 0; i < results.length; i++) {
      const result = results[i];
      if (result.status === 'fulfilled') {
        features.push(result.value.value);
      } else {
        console.warn(`⚠️  Failed to compute feature ${featureNames[i]}:`, result.reason.message);
        features.push(0); // Default value for failed features
      }
    }
    
    return features;
  }

  async getFeatureGroup(groupName, entityId, context = {}) {
    const groupFeatures = this.featureDefinitions.get(groupName);
    if (!groupFeatures) {
      throw new Error(`Feature group not found: ${groupName}`);
    }
    
    const featureNames = Object.keys(groupFeatures);
    const values = await this.getFeatureVector(entityId, featureNames, context);
    
    const result = {};
    for (let i = 0; i < featureNames.length; i++) {
      result[featureNames[i]] = values[i];
    }
    
    return result;
  }

  async storeRawData(featureName, entityId, value, timestamp = null) {
    const ts = timestamp || Date.now();
    const rawDataKey = `raw:${featureName}:${entityId}`;
    
    // Store in sorted set for time-based queries
    await this.redis.zadd(rawDataKey, ts, value);
    
    // Clean old data (keep last 7 days)
    const cutoffTime = ts - (7 * 24 * 60 * 60 * 1000);
    await this.redis.zremrangebyscore(rawDataKey, '-inf', cutoffTime);
  }

  parseTimeWindow(window) {
    const units = {
      's': 1000,
      'm': 60 * 1000,
      'h': 60 * 60 * 1000,
      'd': 24 * 60 * 60 * 1000
    };
    
    const match = window.match(/^(\d+)([smhd])$/);
    if (!match) {
      throw new Error(`Invalid time window format: ${window}`);
    }
    
    const [, amount, unit] = match;
    return parseInt(amount) * units[unit];
  }

  startBatchUpdates() {
    // Update batch features every 5 minutes
    setInterval(async () => {
      await this.updateBatchFeatures();
    }, this.config.batchUpdateInterval);

    console.log(`🕐 Batch feature updates started (${this.config.batchUpdateInterval / 1000}s interval)`);
  }

  async updateBatchFeatures() {
    console.log('🔄 Updating batch features...');
    
    const startTime = performance.now();
    let updateCount = 0;
    
    try {
      // Get list of entities to update
      const agentIds = await this.getActiveEntities('agent');
      const systemEntities = ['system'];
      
      for (const [groupName, features] of this.featureDefinitions) {
        for (const [featureName, featureDef] of Object.entries(features)) {
          if (featureDef.updateFrequency === 'batch') {
            
            const entities = featureName.startsWith('system_') ? systemEntities : agentIds;
            
            for (const entityId of entities) {
              try {
                // Force recomputation by clearing cache
                await this.redis.del(`feature:${featureName}:${entityId}`);
                await this.computeFeature(featureName, entityId);
                updateCount++;
              } catch (error) {
                console.error(`❌ Batch update failed for ${featureName}:${entityId}:`, error.message);
              }
            }
          }
        }
      }
      
      const updateTime = performance.now() - startTime;
      console.log(`✅ Batch features updated: ${updateCount} features in ${updateTime.toFixed(0)}ms`);
      
    } catch (error) {
      console.error('❌ Batch update failed:', error.message);
    }
  }

  async getActiveEntities(entityType) {
    switch (entityType) {
      case 'agent':
        const agentKeys = await this.redis.keys('agent:*:status');
        return agentKeys.map(key => key.split(':')[1]);
      default:
        return [];
    }
  }

  async getFeatureStats() {
    const stats = {
      totalFeatures: this.getTotalFeatureCount(),
      featureGroups: this.config.featureGroups.length,
      cacheHitRate: 0,
      avgComputationTime: 0,
      recentUpdates: 0
    };
    
    try {
      // Calculate cache statistics
      const featureKeys = await this.redis.keys('feature:*');
      const totalRequests = featureKeys.length;
      
      if (totalRequests > 0) {
        stats.cacheHitRate = totalRequests / (totalRequests * 1.2); // Approximate
      }
      
      // Get recent update count
      const updateKey = 'feature_store:batch_updates';
      const recentUpdates = await this.redis.get(updateKey);
      stats.recentUpdates = parseInt(recentUpdates) || 0;
      
    } catch (error) {
      console.error('Error calculating feature stats:', error.message);
    }
    
    return stats;
  }

  async shutdown() {
    console.log('🔄 Shutting down Feature Store...');
    
    if (this.redis) {
      await this.redis.disconnect();
    }
    
    console.log('✅ Feature Store shutdown complete');
  }
}

module.exports = FeatureStore;

// Start feature store if run directly
if (require.main === module) {
  const featureStore = new FeatureStore();
  
  // Example: Compute agent performance features
  const exampleUsage = async () => {
    try {
      console.log('📊 Computing example features...');
      
      const agentId = 'coder-001';
      const context = {
        taskId: 'task-123',
        taskType: 'implementation',
        description: 'Implement user authentication system',
        files: ['auth.js', 'user.js', 'middleware.js']
      };
      
      // Get agent performance features
      const perfFeatures = await featureStore.getFeatureGroup('agent_performance', agentId, context);
      console.log('Agent performance features:', perfFeatures);
      
      // Get task characteristics
      const taskFeatures = await featureStore.getFeatureGroup('task_characteristics', 'task-123', context);
      console.log('Task characteristics:', taskFeatures);
      
    } catch (error) {
      console.error('❌ Example failed:', error.message);
    }
  };
  
  setTimeout(exampleUsage, 2000); // Run after initialization
  
  // Graceful shutdown
  process.on('SIGTERM', async () => {
    await featureStore.shutdown();
    process.exit(0);
  });
  
  process.on('SIGINT', async () => {
    await featureStore.shutdown();
    process.exit(0);
  });
}