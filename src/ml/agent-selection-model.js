#!/usr/bin/env node

/**
 * Random Forest Agent Selection Model
 * ML-based intelligent agent selection using historical performance data
 */

const { RandomForestRegressor } = require('ml-random-forest');
const Redis = require('ioredis');
const { performance } = require('perf_hooks');

class AgentSelectionModel {
  constructor(config = {}) {
    this.config = {
      nEstimators: config.nEstimators || 100,
      maxDepth: config.maxDepth || 10,
      minSamplesLeaf: config.minSamplesLeaf || 3,
      maxFeatures: config.maxFeatures || 'sqrt',
      redisNodes: config.redisNodes || [
        { host: 'redis-cluster-0.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-1.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-2.redis-cluster-service.ollamamax-redis', port: 6379 }
      ],
      redisPassword: config.redisPassword || 'ollama_redis_pass',
      modelUpdateInterval: config.modelUpdateInterval || 300000, // 5 minutes
      ...config
    };

    this.model = null;
    this.featureNames = [
      'agent_historical_success_rate',
      'agent_avg_execution_time', 
      'agent_current_load',
      'task_complexity_score',
      'task_priority_level',
      'time_of_day_factor',
      'agent_specialization_match',
      'resource_availability',
      'context_similarity',
      'recent_performance_trend'
    ];

    this.redis = null;
    this.isTraining = false;
    this.lastTrainingTime = 0;
    this.modelAccuracy = 0;
    
    this.initializeRedis();
    this.startModelUpdateScheduler();
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
      console.log('✅ Agent Selection Model connected to Redis cluster');
    } catch (error) {
      console.error('❌ Redis cluster connection failed:', error.message);
      throw error;
    }
  }

  async extractFeatures(taskRequest, availableAgents) {
    const features = [];
    
    for (const agent of availableAgents) {
      const agentFeatures = await this.getAgentFeatures(agent, taskRequest);
      features.push(agentFeatures);
    }
    
    return features;
  }

  async getAgentFeatures(agentId, taskRequest) {
    const [
      agentStats,
      currentLoad,
      specializationMatch,
      resourceInfo,
      recentTrend
    ] = await Promise.all([
      this.getAgentStats(agentId),
      this.getCurrentLoad(agentId),
      this.calculateSpecializationMatch(agentId, taskRequest),
      this.getResourceAvailability(agentId),
      this.getRecentPerformanceTrend(agentId)
    ]);

    const taskComplexity = this.calculateTaskComplexity(taskRequest);
    const timeOfDayFactor = this.getTimeOfDayFactor();
    const contextSimilarity = await this.calculateContextSimilarity(agentId, taskRequest);

    return [
      agentStats.successRate || 0.5,           // agent_historical_success_rate
      agentStats.avgExecutionTime || 5000,     // agent_avg_execution_time
      currentLoad || 0,                        // agent_current_load
      taskComplexity,                          // task_complexity_score
      taskRequest.priority || 5,               // task_priority_level
      timeOfDayFactor,                         // time_of_day_factor
      specializationMatch,                     // agent_specialization_match
      resourceInfo.availability || 0.5,       // resource_availability
      contextSimilarity,                       // context_similarity
      recentTrend || 0                         // recent_performance_trend
    ];
  }

  async getAgentStats(agentId) {
    try {
      const stats = await this.redis.hgetall(`agent:${agentId}:aggregated`);
      return {
        successRate: parseFloat(stats.successRate) || 0.5,
        avgExecutionTime: parseFloat(stats.avgDuration) || 5000,
        totalTasks: parseInt(stats.totalTasks) || 0,
        errorRate: parseFloat(stats.errorRate) || 0.05
      };
    } catch (error) {
      console.error(`Error fetching agent stats for ${agentId}:`, error.message);
      return { successRate: 0.5, avgExecutionTime: 5000, totalTasks: 0, errorRate: 0.05 };
    }
  }

  async getCurrentLoad(agentId) {
    try {
      const load = await this.redis.get(`agent:${agentId}:current_load`);
      return parseFloat(load) || 0;
    } catch (error) {
      return 0;
    }
  }

  calculateSpecializationMatch(agentId, taskRequest) {
    // Simplified specialization matching based on agent type and task requirements
    const agentType = this.getAgentType(agentId);
    const taskType = taskRequest.type || 'general';
    
    const matchMap = {
      'coder': { 'implementation': 0.9, 'coding': 0.9, 'debugging': 0.8, 'general': 0.6 },
      'reviewer': { 'review': 0.9, 'quality': 0.8, 'security': 0.7, 'general': 0.5 },
      'tester': { 'testing': 0.9, 'validation': 0.8, 'qa': 0.8, 'general': 0.5 },
      'researcher': { 'analysis': 0.9, 'research': 0.9, 'investigation': 0.8, 'general': 0.6 },
      'planner': { 'planning': 0.9, 'strategy': 0.8, 'coordination': 0.7, 'general': 0.6 }
    };

    return matchMap[agentType]?.[taskType] || 0.5;
  }

  getAgentType(agentId) {
    // Extract agent type from agent ID or use registry lookup
    if (agentId.includes('coder')) return 'coder';
    if (agentId.includes('reviewer')) return 'reviewer';
    if (agentId.includes('tester')) return 'tester';
    if (agentId.includes('researcher')) return 'researcher';
    if (agentId.includes('planner')) return 'planner';
    return 'general';
  }

  async getResourceAvailability(agentId) {
    try {
      const resources = await this.redis.hgetall(`agent:${agentId}:resources`);
      const cpuUsage = parseFloat(resources.cpu) || 0;
      const memoryUsage = parseFloat(resources.memory) || 0;
      
      // Calculate availability as inverse of resource usage
      return Math.max(0, 1 - Math.max(cpuUsage / 100, memoryUsage / 100));
    } catch (error) {
      return 0.5; // Default availability
    }
  }

  async getRecentPerformanceTrend(agentId) {
    try {
      const recentMetrics = await this.redis.lrange(`agent:${agentId}:recent_performance`, 0, 9);
      if (recentMetrics.length < 2) return 0;
      
      const scores = recentMetrics.map(m => {
        const metric = JSON.parse(m);
        return metric.success ? 1 : 0;
      });
      
      // Calculate trend (positive = improving, negative = declining)
      const recent = scores.slice(0, Math.floor(scores.length / 2));
      const older = scores.slice(Math.floor(scores.length / 2));
      
      const recentAvg = recent.reduce((a, b) => a + b, 0) / recent.length;
      const olderAvg = older.reduce((a, b) => a + b, 0) / older.length;
      
      return recentAvg - olderAvg;
    } catch (error) {
      return 0;
    }
  }

  calculateTaskComplexity(taskRequest) {
    let complexity = 1; // Base complexity

    // Factor in task description length
    const descLength = (taskRequest.description || '').length;
    complexity += Math.min(descLength / 1000, 3);

    // Factor in estimated resources
    if (taskRequest.expectedDuration > 300000) complexity += 2; // > 5 minutes
    if (taskRequest.files && taskRequest.files.length > 5) complexity += 1.5;
    if (taskRequest.dependencies && taskRequest.dependencies.length > 3) complexity += 1;

    // Factor in task type
    const complexTasks = ['architecture', 'system-design', 'optimization', 'migration'];
    if (complexTasks.includes(taskRequest.type)) complexity += 2;

    return Math.min(complexity, 10); // Cap at 10
  }

  getTimeOfDayFactor() {
    const hour = new Date().getHours();
    
    // Peak performance hours (9 AM - 5 PM): 1.0
    // Off hours: 0.8
    if (hour >= 9 && hour <= 17) {
      return 1.0;
    } else {
      return 0.8;
    }
  }

  async calculateContextSimilarity(agentId, taskRequest) {
    try {
      // Get agent's recent task contexts
      const recentContexts = await this.redis.lrange(`agent:${agentId}:recent_contexts`, 0, 4);
      if (recentContexts.length === 0) return 0;

      const currentContext = this.extractTaskContext(taskRequest);
      let maxSimilarity = 0;

      for (const contextStr of recentContexts) {
        const context = JSON.parse(contextStr);
        const similarity = this.calculateCosineSimilarity(currentContext, context);
        maxSimilarity = Math.max(maxSimilarity, similarity);
      }

      return maxSimilarity;
    } catch (error) {
      return 0;
    }
  }

  extractTaskContext(taskRequest) {
    // Simple context extraction based on keywords
    const text = `${taskRequest.type || ''} ${taskRequest.description || ''} ${taskRequest.tags?.join(' ') || ''}`.toLowerCase();
    const keywords = ['api', 'database', 'frontend', 'backend', 'test', 'security', 'performance', 'ui', 'algorithm', 'data'];
    
    return keywords.map(keyword => text.includes(keyword) ? 1 : 0);
  }

  calculateCosineSimilarity(vecA, vecB) {
    if (vecA.length !== vecB.length) return 0;
    
    let dotProduct = 0;
    let normA = 0;
    let normB = 0;
    
    for (let i = 0; i < vecA.length; i++) {
      dotProduct += vecA[i] * vecB[i];
      normA += vecA[i] * vecA[i];
      normB += vecB[i] * vecB[i];
    }
    
    if (normA === 0 || normB === 0) return 0;
    
    return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB));
  }

  async collectTrainingData() {
    try {
      const trainingData = [];
      const labels = [];

      // Get historical task assignments and outcomes
      const taskKeys = await this.redis.keys('task:*:assignment');
      
      for (const key of taskKeys.slice(-1000)) { // Limit to recent 1000 tasks
        const assignment = await this.redis.hgetall(key);
        if (!assignment.agentId || !assignment.outcome) continue;

        const taskId = key.split(':')[1];
        const taskRequest = await this.redis.hgetall(`task:${taskId}:request`);
        if (!taskRequest.description) continue;

        const features = await this.getAgentFeatures(assignment.agentId, taskRequest);
        const outcome = parseFloat(assignment.outcome); // Success score (0-1)

        trainingData.push(features);
        labels.push(outcome);
      }

      console.log(`📊 Collected ${trainingData.length} training samples`);
      return { features: trainingData, labels };
    } catch (error) {
      console.error('❌ Error collecting training data:', error.message);
      return { features: [], labels: [] };
    }
  }

  async trainModel() {
    if (this.isTraining) {
      console.log('🔄 Model training already in progress');
      return;
    }

    this.isTraining = true;
    const startTime = performance.now();

    try {
      console.log('🚀 Starting Random Forest model training...');
      
      const { features, labels } = await this.collectTrainingData();
      
      if (features.length < 50) {
        console.log('⚠️  Insufficient training data, using default model');
        this.model = new RandomForestRegressor({
          nEstimators: this.config.nEstimators,
          maxDepth: this.config.maxDepth,
          minSamplesLeaf: this.config.minSamplesLeaf,
          maxFeatures: this.config.maxFeatures,
          seed: 42
        });
        return;
      }

      // Split data for validation
      const splitIndex = Math.floor(features.length * 0.8);
      const trainFeatures = features.slice(0, splitIndex);
      const trainLabels = labels.slice(0, splitIndex);
      const testFeatures = features.slice(splitIndex);
      const testLabels = labels.slice(splitIndex);

      // Train the model
      this.model = new RandomForestRegressor({
        nEstimators: this.config.nEstimators,
        maxDepth: this.config.maxDepth,
        minSamplesLeaf: this.config.minSamplesLeaf,
        maxFeatures: this.config.maxFeatures,
        seed: 42
      });

      this.model.fit(trainFeatures, trainLabels);

      // Validate model
      const predictions = this.model.predict(testFeatures);
      this.modelAccuracy = this.calculateAccuracy(predictions, testLabels);

      const trainingTime = performance.now() - startTime;
      this.lastTrainingTime = Date.now();

      console.log(`✅ Model training completed:`);
      console.log(`   📈 Accuracy: ${(this.modelAccuracy * 100).toFixed(2)}%`);
      console.log(`   🕐 Training time: ${trainingTime.toFixed(0)}ms`);
      console.log(`   📊 Training samples: ${trainFeatures.length}`);
      console.log(`   🧪 Test samples: ${testFeatures.length}`);

      // Store model metadata
      await this.redis.hset('ml:agent_selection_model', {
        accuracy: this.modelAccuracy,
        trainingTime: trainingTime,
        trainingSamples: trainFeatures.length,
        lastTrained: this.lastTrainingTime,
        version: '1.0'
      });

    } catch (error) {
      console.error('❌ Model training failed:', error.message);
      throw error;
    } finally {
      this.isTraining = false;
    }
  }

  calculateAccuracy(predictions, actual) {
    let correct = 0;
    const threshold = 0.1; // Within 10% considered correct
    
    for (let i = 0; i < predictions.length; i++) {
      if (Math.abs(predictions[i] - actual[i]) <= threshold) {
        correct++;
      }
    }
    
    return correct / predictions.length;
  }

  async selectBestAgent(taskRequest, availableAgents) {
    if (!this.model) {
      console.log('⚠️  Model not trained, using fallback selection');
      return this.fallbackSelection(taskRequest, availableAgents);
    }

    try {
      const features = await this.extractFeatures(taskRequest, availableAgents);
      const predictions = this.model.predict(features);

      // Combine predictions with agent IDs and sort by predicted performance
      const agentPredictions = availableAgents.map((agentId, index) => ({
        agentId,
        predictedPerformance: predictions[index],
        features: features[index]
      }));

      agentPredictions.sort((a, b) => b.predictedPerformance - a.predictedPerformance);

      const selectedAgent = agentPredictions[0];
      
      console.log(`🎯 Selected agent: ${selectedAgent.agentId}`);
      console.log(`   📈 Predicted performance: ${(selectedAgent.predictedPerformance * 100).toFixed(1)}%`);
      
      return {
        agentId: selectedAgent.agentId,
        confidence: selectedAgent.predictedPerformance,
        alternatives: agentPredictions.slice(1, 3),
        selectionReason: 'ml_model_prediction',
        modelAccuracy: this.modelAccuracy
      };

    } catch (error) {
      console.error('❌ ML selection failed:', error.message);
      return this.fallbackSelection(taskRequest, availableAgents);
    }
  }

  fallbackSelection(taskRequest, availableAgents) {
    // Simple fallback based on specialization match
    let bestAgent = availableAgents[0];
    let bestScore = 0;

    for (const agentId of availableAgents) {
      const score = this.calculateSpecializationMatch(agentId, taskRequest);
      if (score > bestScore) {
        bestScore = score;
        bestAgent = agentId;
      }
    }

    return {
      agentId: bestAgent,
      confidence: bestScore,
      alternatives: [],
      selectionReason: 'fallback_specialization_match',
      modelAccuracy: 0
    };
  }

  startModelUpdateScheduler() {
    setInterval(async () => {
      try {
        await this.trainModel();
      } catch (error) {
        console.error('❌ Scheduled model update failed:', error.message);
      }
    }, this.config.modelUpdateInterval);

    console.log(`🕐 Model update scheduler started (${this.config.modelUpdateInterval / 1000}s interval)`);
  }

  async getModelStatus() {
    const status = {
      modelTrained: !!this.model,
      accuracy: this.modelAccuracy,
      lastTrainingTime: this.lastTrainingTime,
      isTraining: this.isTraining,
      features: this.featureNames,
      config: this.config
    };

    try {
      const redisStats = await this.redis.hgetall('ml:agent_selection_model');
      status.persistedStats = redisStats;
    } catch (error) {
      status.persistedStats = {};
    }

    return status;
  }

  async shutdown() {
    console.log('🔄 Shutting down Agent Selection Model...');
    
    if (this.redis) {
      await this.redis.disconnect();
    }
    
    console.log('✅ Agent Selection Model shutdown complete');
  }
}

module.exports = AgentSelectionModel;

// Start model if run directly
if (require.main === module) {
  const model = new AgentSelectionModel();
  
  // Initialize and train model
  model.trainModel().catch(console.error);
  
  // Graceful shutdown
  process.on('SIGTERM', async () => {
    await model.shutdown();
    process.exit(0);
  });
  
  process.on('SIGINT', async () => {
    await model.shutdown();
    process.exit(0);
  });
}