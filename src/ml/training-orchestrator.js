#!/usr/bin/env node

/**
 * ML Model Training Orchestrator
 * Centralized training pipeline for all ML models with distributed training support
 */

const Redis = require('ioredis');
const { performance } = require('perf_hooks');
const AgentSelectionModel = require('./agent-selection-model');
const PredictiveScalingSystem = require('./predictive-scaling');
const FeatureStore = require('./feature-store');

class MLTrainingOrchestrator {
  constructor(config = {}) {
    this.config = {
      redisNodes: config.redisNodes || [
        { host: 'redis-cluster-0.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-1.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-2.redis-cluster-service.ollamamax-redis', port: 6379 }
      ],
      redisPassword: config.redisPassword || 'ollama_redis_pass',
      trainingSchedule: config.trainingSchedule || {
        agent_selection: '0 */6 * * *',    // Every 6 hours
        predictive_scaling: '0 */4 * * *', // Every 4 hours
        feature_engineering: '0 */2 * * *' // Every 2 hours
      },
      modelVersioning: config.modelVersioning || true,
      distributedTraining: config.distributedTraining || false,
      trainingThresholds: config.trainingThresholds || {
        minDataPoints: 1000,
        maxTrainingTime: 30 * 60 * 1000, // 30 minutes
        accuracyThreshold: 0.75
      },
      ...config
    };

    this.redis = null;
    this.models = new Map();
    this.featureStore = null;
    this.trainingQueue = [];
    this.activeTrainingJobs = new Map();
    this.trainingHistory = [];

    this.initializeRedis();
    this.initializeModels();
    this.startTrainingScheduler();
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
      console.log('✅ ML Training Orchestrator connected to Redis cluster');
    } catch (error) {
      console.error('❌ Redis cluster connection failed:', error.message);
      throw error;
    }
  }

  async initializeModels() {
    console.log('🤖 Initializing ML models...');

    // Initialize Feature Store
    this.featureStore = new FeatureStore({
      redisNodes: this.config.redisNodes,
      redisPassword: this.config.redisPassword
    });

    // Register models
    this.registerModel('agent_selection', {
      class: AgentSelectionModel,
      config: {
        nEstimators: 200,
        maxDepth: 15,
        minSamplesLeaf: 2,
        maxFeatures: 'sqrt'
      },
      features: [
        'agent_success_rate_24h',
        'agent_avg_execution_time', 
        'agent_current_load',
        'task_complexity_score',
        'task_priority_level',
        'agent_specialization_match',
        'system_cpu_utilization',
        'hour_of_day',
        'similar_task_performance',
        'agent_resource_efficiency'
      ],
      target: 'task_success_probability',
      retrainingInterval: 6 * 60 * 60 * 1000 // 6 hours
    });

    this.registerModel('predictive_scaling', {
      class: PredictiveScalingSystem,
      config: {
        sequenceLength: 120, // 2 hours of minute-level data
        predictionHorizon: 30, // 30 minutes ahead
        batchSize: 64,
        epochs: 150,
        learningRate: 0.0005
      },
      features: [
        'active_agent_count',
        'system_queue_length',
        'system_avg_response_time',
        'system_cpu_utilization',
        'system_memory_utilization',
        'task_complexity_score',
        'agent_success_rate_1h',
        'hour_of_day',
        'day_of_week',
        'seasonal_factor'
      ],
      target: 'required_agent_count',
      retrainingInterval: 4 * 60 * 60 * 1000 // 4 hours
    });

    console.log(`✅ Registered ${this.models.size} ML models`);
  }

  registerModel(modelName, modelConfig) {
    this.models.set(modelName, {
      name: modelName,
      ...modelConfig,
      instance: null,
      lastTrained: 0,
      version: '1.0.0',
      performance: {
        accuracy: 0,
        trainingTime: 0,
        dataPoints: 0
      }
    });
  }

  async scheduleTraining(modelName, priority = 'normal', force = false) {
    const modelConfig = this.models.get(modelName);
    if (!modelConfig) {
      throw new Error(`Model ${modelName} not found`);
    }

    // Check if retraining is needed
    const timeSinceLastTraining = Date.now() - modelConfig.lastTrained;
    if (!force && timeSinceLastTraining < modelConfig.retrainingInterval) {
      console.log(`⏭️  Model ${modelName} training skipped - too soon since last training`);
      return null;
    }

    const trainingJob = {
      id: `training_${modelName}_${Date.now()}`,
      modelName,
      priority,
      scheduledTime: Date.now(),
      status: 'queued',
      retries: 0,
      maxRetries: 3,
      force
    };

    // Add to queue with priority ordering
    if (priority === 'high') {
      this.trainingQueue.unshift(trainingJob);
    } else {
      this.trainingQueue.push(trainingJob);
    }

    await this.redis.lpush('ml:training_queue', JSON.stringify(trainingJob));
    
    console.log(`📅 Training scheduled: ${modelName} (${priority} priority)`);
    
    // Start processing if not already running
    this.processTrainingQueue();
    
    return trainingJob;
  }

  async processTrainingQueue() {
    if (this.trainingQueue.length === 0) return;

    // Process up to 2 concurrent training jobs
    const maxConcurrentJobs = 2;
    const runningJobs = this.activeTrainingJobs.size;

    if (runningJobs >= maxConcurrentJobs) {
      console.log(`⏸️  Training queue paused - ${runningJobs} jobs already running`);
      return;
    }

    const job = this.trainingQueue.shift();
    if (!job) return;

    try {
      await this.executeTrainingJob(job);
    } catch (error) {
      console.error(`❌ Training job failed: ${job.id}`, error.message);
      await this.handleTrainingFailure(job, error);
    }

    // Continue processing
    setTimeout(() => this.processTrainingQueue(), 1000);
  }

  async executeTrainingJob(job) {
    console.log(`🚀 Starting training job: ${job.id}`);
    
    job.status = 'running';
    job.startTime = Date.now();
    this.activeTrainingJobs.set(job.id, job);

    const modelConfig = this.models.get(job.modelName);
    const startTime = performance.now();

    try {
      // Step 1: Data validation
      await this.validateTrainingData(job.modelName);
      
      // Step 2: Feature preparation
      const features = await this.prepareFeatures(job.modelName);
      
      // Step 3: Model initialization
      if (!modelConfig.instance) {
        modelConfig.instance = new modelConfig.class({
          ...modelConfig.config,
          redisNodes: this.config.redisNodes,
          redisPassword: this.config.redisPassword
        });
      }

      // Step 4: Model training
      console.log(`🎯 Training ${job.modelName} model...`);
      await modelConfig.instance.trainModel();

      // Step 5: Model validation
      const validationResults = await this.validateModel(job.modelName);
      
      // Step 6: Model versioning
      if (this.config.modelVersioning) {
        await this.versionModel(job.modelName, validationResults);
      }

      const trainingTime = performance.now() - startTime;
      
      // Update model config
      modelConfig.lastTrained = Date.now();
      modelConfig.performance = {
        accuracy: validationResults.accuracy,
        trainingTime,
        dataPoints: validationResults.dataPoints,
        validationScore: validationResults.validationScore
      };

      // Update job status
      job.status = 'completed';
      job.completionTime = Date.now();
      job.duration = job.completionTime - job.startTime;
      job.results = validationResults;

      console.log(`✅ Training completed: ${job.modelName}`);
      console.log(`   🎯 Accuracy: ${(validationResults.accuracy * 100).toFixed(2)}%`);
      console.log(`   ⏱️  Duration: ${(trainingTime / 1000).toFixed(1)}s`);
      console.log(`   📊 Data points: ${validationResults.dataPoints}`);

      // Store training results
      await this.storeTrainingResults(job);
      
    } catch (error) {
      job.status = 'failed';
      job.error = error.message;
      job.completionTime = Date.now();
      throw error;
    } finally {
      this.activeTrainingJobs.delete(job.id);
      this.trainingHistory.push(job);
    }
  }

  async validateTrainingData(modelName) {
    const modelConfig = this.models.get(modelName);
    
    // Check data availability
    let dataCount = 0;
    
    switch (modelName) {
      case 'agent_selection':
        const taskKeys = await this.redis.keys('task:*:assignment');
        dataCount = taskKeys.length;
        break;
      case 'predictive_scaling':
        const metricsData = await this.redis.llen('ml:scaling_metrics');
        dataCount = metricsData;
        break;
    }

    if (dataCount < this.config.trainingThresholds.minDataPoints) {
      throw new Error(`Insufficient training data: ${dataCount} < ${this.config.trainingThresholds.minDataPoints}`);
    }

    console.log(`✅ Data validation passed: ${dataCount} data points`);
    return { dataCount };
  }

  async prepareFeatures(modelName) {
    const modelConfig = this.models.get(modelName);
    console.log(`🔧 Preparing features for ${modelName}...`);

    // Update batch features in feature store
    const featureUpdatePromises = modelConfig.features.map(featureName => 
      this.featureStore.updateBatchFeatures(featureName)
    );

    await Promise.allSettled(featureUpdatePromises);
    
    console.log(`✅ Features prepared: ${modelConfig.features.length} features`);
    return modelConfig.features;
  }

  async validateModel(modelName) {
    console.log(`🧪 Validating ${modelName} model...`);
    
    const modelConfig = this.models.get(modelName);
    const instance = modelConfig.instance;

    let validationResults;

    switch (modelName) {
      case 'agent_selection':
        validationResults = await this.validateAgentSelectionModel(instance);
        break;
      case 'predictive_scaling':
        validationResults = await this.validatePredictiveScalingModel(instance);
        break;
      default:
        validationResults = await this.genericModelValidation(instance);
    }

    if (validationResults.accuracy < this.config.trainingThresholds.accuracyThreshold) {
      console.warn(`⚠️  Model accuracy below threshold: ${validationResults.accuracy} < ${this.config.trainingThresholds.accuracyThreshold}`);
    }

    return validationResults;
  }

  async validateAgentSelectionModel(model) {
    // Test with recent task assignments
    const testAssignments = await this.getRecentTestData('agent_selection', 100);
    let correct = 0;
    
    for (const assignment of testAssignments) {
      const prediction = await model.selectBestAgent(assignment.taskRequest, assignment.availableAgents);
      if (prediction.agentId === assignment.actualAgent) {
        correct++;
      }
    }
    
    const accuracy = correct / testAssignments.length;
    
    return {
      accuracy,
      dataPoints: testAssignments.length,
      validationScore: accuracy,
      modelSpecific: {
        correctPredictions: correct,
        totalPredictions: testAssignments.length,
        averageConfidence: testAssignments.reduce((sum, t) => sum + (t.confidence || 0.5), 0) / testAssignments.length
      }
    };
  }

  async validatePredictiveScalingModel(model) {
    // Test prediction accuracy
    const testData = await this.getRecentTestData('predictive_scaling', 50);
    let totalError = 0;
    let validPredictions = 0;
    
    for (const dataPoint of testData) {
      const prediction = await model.makePrediction(dataPoint.sequence);
      if (prediction) {
        const actualAgents = dataPoint.actualMetrics.active_agents;
        const predictedAgents = prediction.predicted_active_agents;
        const error = Math.abs(actualAgents - predictedAgents) / Math.max(actualAgents, 1);
        totalError += error;
        validPredictions++;
      }
    }
    
    const meanError = validPredictions > 0 ? totalError / validPredictions : 1;
    const accuracy = Math.max(0, 1 - meanError);
    
    return {
      accuracy,
      dataPoints: testData.length,
      validationScore: accuracy,
      modelSpecific: {
        meanAbsoluteError: meanError,
        validPredictions,
        totalTestCases: testData.length
      }
    };
  }

  async genericModelValidation(model) {
    // Basic validation for unknown model types
    const status = await model.getModelStatus?.() || {};
    
    return {
      accuracy: status.accuracy || 0.5,
      dataPoints: status.trainingSamples || 0,
      validationScore: status.accuracy || 0.5,
      modelSpecific: status
    };
  }

  async getRecentTestData(modelType, limit) {
    switch (modelType) {
      case 'agent_selection':
        // Get recent task assignments for validation
        const assignmentKeys = await this.redis.keys('task:*:assignment');
        const recent = assignmentKeys.slice(-limit);
        const assignments = [];
        
        for (const key of recent) {
          const assignment = await this.redis.hgetall(key);
          if (assignment.agentId && assignment.outcome) {
            assignments.push({
              taskRequest: await this.redis.hgetall(key.replace(':assignment', ':request')),
              actualAgent: assignment.agentId,
              outcome: parseFloat(assignment.outcome),
              availableAgents: JSON.parse(assignment.availableAgents || '[]')
            });
          }
        }
        return assignments;
        
      case 'predictive_scaling':
        // Get recent scaling metrics sequences
        const metricsData = await this.redis.lrange('ml:scaling_metrics', 0, limit * 2);
        const sequences = [];
        
        for (let i = 0; i < metricsData.length - 60; i += 30) {
          const sequence = metricsData.slice(i, i + 60).map(d => JSON.parse(d));
          const actual = JSON.parse(metricsData[i + 60]);
          sequences.push({
            sequence,
            actualMetrics: actual
          });
        }
        return sequences;
        
      default:
        return [];
    }
  }

  async versionModel(modelName, validationResults) {
    const modelConfig = this.models.get(modelName);
    const currentVersion = modelConfig.version;
    
    // Increment version based on performance improvement
    const versionParts = currentVersion.split('.').map(Number);
    
    if (validationResults.accuracy > (modelConfig.performance.accuracy + 0.05)) {
      // Major improvement - increment minor version
      versionParts[1]++;
      versionParts[2] = 0;
    } else {
      // Minor improvement - increment patch version
      versionParts[2]++;
    }
    
    const newVersion = versionParts.join('.');
    modelConfig.version = newVersion;
    
    // Store model version metadata
    const versionData = {
      version: newVersion,
      modelName,
      timestamp: Date.now(),
      accuracy: validationResults.accuracy,
      dataPoints: validationResults.dataPoints,
      validationScore: validationResults.validationScore,
      previousVersion: currentVersion
    };
    
    await this.redis.hset(`ml:model_versions:${modelName}`, newVersion, JSON.stringify(versionData));
    await this.redis.set(`ml:model_current_version:${modelName}`, newVersion);
    
    console.log(`📦 Model versioned: ${modelName} v${currentVersion} → v${newVersion}`);
    
    return newVersion;
  }

  async storeTrainingResults(job) {
    const trainingResult = {
      jobId: job.id,
      modelName: job.modelName,
      startTime: job.startTime,
      completionTime: job.completionTime,
      duration: job.duration,
      status: job.status,
      results: job.results,
      error: job.error || null
    };
    
    await this.redis.lpush('ml:training_history', JSON.stringify(trainingResult));
    await this.redis.ltrim('ml:training_history', 0, 1000); // Keep last 1000 training jobs
    
    // Update model performance tracking
    await this.redis.hset(`ml:model_performance:${job.modelName}`, {
      lastTrained: job.completionTime,
      accuracy: job.results?.accuracy || 0,
      dataPoints: job.results?.dataPoints || 0,
      trainingTime: job.duration || 0
    });
  }

  async handleTrainingFailure(job, error) {
    console.error(`❌ Training failed for ${job.modelName}:`, error.message);
    
    job.retries++;
    
    if (job.retries < job.maxRetries) {
      console.log(`🔄 Retrying training job: ${job.id} (attempt ${job.retries + 1})`);
      
      // Add back to queue with delay
      setTimeout(() => {
        this.trainingQueue.unshift(job);
        this.processTrainingQueue();
      }, 60000); // 1 minute delay
    } else {
      console.error(`💥 Training job exhausted retries: ${job.id}`);
      job.status = 'failed';
      await this.storeTrainingResults(job);
    }
  }

  startTrainingScheduler() {
    // Check for scheduled training every 5 minutes
    setInterval(() => {
      this.checkScheduledTraining();
    }, 5 * 60 * 1000);

    // Immediate check for any pending training
    this.checkScheduledTraining();

    console.log('🕐 Training scheduler started');
  }

  async checkScheduledTraining() {
    for (const [modelName, modelConfig] of this.models) {
      const timeSinceLastTraining = Date.now() - modelConfig.lastTrained;
      
      if (timeSinceLastTraining >= modelConfig.retrainingInterval) {
        console.log(`⏰ Scheduled training triggered for ${modelName}`);
        await this.scheduleTraining(modelName, 'normal', false);
      }
    }
  }

  async getTrainingStatus() {
    const status = {
      registeredModels: this.models.size,
      activeJobs: this.activeTrainingJobs.size,
      queuedJobs: this.trainingQueue.length,
      completedJobs: this.trainingHistory.length,
      models: {},
      activeJobs: Array.from(this.activeTrainingJobs.values()),
      queuedJobs: this.trainingQueue,
      recentHistory: this.trainingHistory.slice(-10)
    };

    // Get model details
    for (const [modelName, modelConfig] of this.models) {
      status.models[modelName] = {
        version: modelConfig.version,
        lastTrained: modelConfig.lastTrained,
        performance: modelConfig.performance,
        features: modelConfig.features,
        nextTraining: modelConfig.lastTrained + modelConfig.retrainingInterval
      };
    }

    try {
      // Get persisted training history
      const recentTraining = await this.redis.lrange('ml:training_history', 0, 9);
      status.persistedHistory = recentTraining.map(h => JSON.parse(h));
    } catch (error) {
      status.persistedHistory = [];
    }

    return status;
  }

  async forceRetraining(modelName, priority = 'high') {
    console.log(`🔄 Force retraining requested for ${modelName}`);
    return await this.scheduleTraining(modelName, priority, true);
  }

  async shutdown() {
    console.log('🔄 Shutting down ML Training Orchestrator...');
    
    // Wait for active training jobs to complete
    if (this.activeTrainingJobs.size > 0) {
      console.log(`⏳ Waiting for ${this.activeTrainingJobs.size} training jobs to complete...`);
      
      const maxWaitTime = 5 * 60 * 1000; // 5 minutes
      const startWait = Date.now();
      
      while (this.activeTrainingJobs.size > 0 && (Date.now() - startWait) < maxWaitTime) {
        await new Promise(resolve => setTimeout(resolve, 5000)); // Check every 5 seconds
      }
    }
    
    // Shutdown model instances
    for (const [modelName, modelConfig] of this.models) {
      if (modelConfig.instance && typeof modelConfig.instance.shutdown === 'function') {
        await modelConfig.instance.shutdown();
      }
    }
    
    // Shutdown feature store
    if (this.featureStore) {
      await this.featureStore.shutdown();
    }
    
    if (this.redis) {
      await this.redis.disconnect();
    }
    
    console.log('✅ ML Training Orchestrator shutdown complete');
  }
}

module.exports = MLTrainingOrchestrator;

// Start orchestrator if run directly
if (require.main === module) {
  const orchestrator = new MLTrainingOrchestrator();
  
  // Schedule initial training for all models
  setTimeout(async () => {
    console.log('🚀 Scheduling initial training for all models...');
    
    try {
      await orchestrator.scheduleTraining('agent_selection', 'high', true);
      await orchestrator.scheduleTraining('predictive_scaling', 'high', true);
      
      console.log('✅ Initial training scheduled');
    } catch (error) {
      console.error('❌ Initial training scheduling failed:', error.message);
    }
  }, 5000);
  
  // Graceful shutdown
  process.on('SIGTERM', async () => {
    await orchestrator.shutdown();
    process.exit(0);
  });
  
  process.on('SIGINT', async () => {
    await orchestrator.shutdown();
    process.exit(0);
  });
}