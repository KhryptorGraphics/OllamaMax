#!/usr/bin/env node

/**
 * LSTM Predictive Scaling System
 * Deep learning-based workload prediction and automated agent scaling
 */

const tf = require('@tensorflow/tfjs-node');
const Redis = require('ioredis');
const { performance } = require('perf_hooks');

class PredictiveScalingSystem {
  constructor(config = {}) {
    this.config = {
      sequenceLength: config.sequenceLength || 60, // 60 time steps
      predictionHorizon: config.predictionHorizon || 15, // 15 minutes ahead
      features: config.features || [
        'active_agents',
        'queue_length', 
        'avg_response_time',
        'cpu_utilization',
        'memory_utilization',
        'task_complexity_avg',
        'success_rate',
        'hour_of_day',
        'day_of_week',
        'seasonal_factor'
      ],
      batchSize: config.batchSize || 32,
      epochs: config.epochs || 100,
      learningRate: config.learningRate || 0.001,
      validationSplit: config.validationSplit || 0.2,
      redisNodes: config.redisNodes || [
        { host: 'redis-cluster-0.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-1.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-2.redis-cluster-service.ollamamax-redis', port: 6379 }
      ],
      redisPassword: config.redisPassword || 'ollama_redis_pass',
      modelUpdateInterval: config.modelUpdateInterval || 1800000, // 30 minutes
      scaleUpThreshold: config.scaleUpThreshold || 0.8,
      scaleDownThreshold: config.scaleDownThreshold || 0.3,
      minAgents: config.minAgents || 5,
      maxAgents: config.maxAgents || 50,
      ...config
    };

    this.model = null;
    this.scaler = null;
    this.redis = null;
    this.isTraining = false;
    this.lastTrainingTime = 0;
    this.modelAccuracy = 0;
    this.currentSequence = [];
    this.scalingHistory = [];

    this.initializeRedis();
    this.startDataCollection();
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
      console.log('✅ Predictive Scaling System connected to Redis cluster');
    } catch (error) {
      console.error('❌ Redis cluster connection failed:', error.message);
      throw error;
    }
  }

  async collectCurrentMetrics() {
    try {
      const now = Date.now();
      const hour = new Date(now).getHours();
      const dayOfWeek = new Date(now).getDay();
      
      const [
        activeAgents,
        queueLength,
        avgResponseTime,
        cpuUtil,
        memUtil,
        taskComplexity,
        successRate
      ] = await Promise.all([
        this.getActiveAgentsCount(),
        this.getQueueLength(),
        this.getAverageResponseTime(),
        this.getCPUUtilization(),
        this.getMemoryUtilization(),
        this.getTaskComplexityAverage(),
        this.getSuccessRate()
      ]);

      const seasonalFactor = this.calculateSeasonalFactor(hour, dayOfWeek);

      return {
        timestamp: now,
        active_agents: activeAgents,
        queue_length: queueLength,
        avg_response_time: avgResponseTime,
        cpu_utilization: cpuUtil,
        memory_utilization: memUtil,
        task_complexity_avg: taskComplexity,
        success_rate: successRate,
        hour_of_day: hour / 24, // Normalize to 0-1
        day_of_week: dayOfWeek / 7, // Normalize to 0-1
        seasonal_factor: seasonalFactor
      };
    } catch (error) {
      console.error('❌ Error collecting metrics:', error.message);
      return null;
    }
  }

  async getActiveAgentsCount() {
    try {
      const agentKeys = await this.redis.keys('agent:*:status');
      let activeCount = 0;
      
      for (const key of agentKeys) {
        const status = await this.redis.get(key);
        if (status === 'active') activeCount++;
      }
      
      return activeCount;
    } catch (error) {
      return 0;
    }
  }

  async getQueueLength() {
    try {
      const queueLength = await this.redis.llen('swarm:task_queue');
      return queueLength || 0;
    } catch (error) {
      return 0;
    }
  }

  async getAverageResponseTime() {
    try {
      const responseTime = await this.redis.get('metrics:avg_response_time');
      return parseFloat(responseTime) || 0;
    } catch (error) {
      return 0;
    }
  }

  async getCPUUtilization() {
    try {
      const globalStats = await this.redis.hgetall('global:stats:realtime');
      return parseFloat(globalStats.avgCpu) || 0;
    } catch (error) {
      return 0;
    }
  }

  async getMemoryUtilization() {
    try {
      const globalStats = await this.redis.hgetall('global:stats:realtime');
      return parseFloat(globalStats.avgMemory) || 0;
    } catch (error) {
      return 0;
    }
  }

  async getTaskComplexityAverage() {
    try {
      const complexity = await this.redis.get('metrics:avg_task_complexity');
      return parseFloat(complexity) || 1;
    } catch (error) {
      return 1;
    }
  }

  async getSuccessRate() {
    try {
      const globalStats = await this.redis.hgetall('global:stats:realtime');
      return parseFloat(globalStats.successRate) || 0.5;
    } catch (error) {
      return 0.5;
    }
  }

  calculateSeasonalFactor(hour, dayOfWeek) {
    // Business hours boost
    const businessHoursFactor = (hour >= 9 && hour <= 17) ? 1.2 : 0.8;
    
    // Weekday vs weekend
    const weekdayFactor = (dayOfWeek >= 1 && dayOfWeek <= 5) ? 1.1 : 0.9;
    
    return businessHoursFactor * weekdayFactor;
  }

  startDataCollection() {
    // Collect metrics every minute
    setInterval(async () => {
      const metrics = await this.collectCurrentMetrics();
      if (metrics) {
        // Add to current sequence
        this.currentSequence.push(metrics);
        
        // Maintain sequence length
        if (this.currentSequence.length > this.config.sequenceLength) {
          this.currentSequence.shift();
        }
        
        // Store in Redis for training data
        await this.redis.lpush('ml:scaling_metrics', JSON.stringify(metrics));
        await this.redis.ltrim('ml:scaling_metrics', 0, 10000); // Keep last 10k samples
        
        // Make prediction if model is ready and we have enough data
        if (this.model && this.currentSequence.length === this.config.sequenceLength) {
          await this.makePredictionAndScale();
        }
      }
    }, 60000); // 1 minute

    console.log('🕐 Data collection started (1-minute intervals)');
  }

  async collectTrainingData() {
    try {
      console.log('📊 Collecting training data...');
      
      const metricsData = await this.redis.lrange('ml:scaling_metrics', 0, -1);
      
      if (metricsData.length < this.config.sequenceLength + this.config.predictionHorizon) {
        console.log('⚠️  Insufficient training data');
        return { sequences: [], targets: [] };
      }

      const metrics = metricsData.map(data => JSON.parse(data)).reverse(); // Oldest first
      
      const sequences = [];
      const targets = [];

      // Create sequences and targets
      for (let i = 0; i < metrics.length - this.config.sequenceLength - this.config.predictionHorizon; i++) {
        const sequence = metrics.slice(i, i + this.config.sequenceLength);
        const target = metrics[i + this.config.sequenceLength + this.config.predictionHorizon];
        
        // Convert to feature vectors
        const sequenceFeatures = sequence.map(m => this.extractFeatureVector(m));
        const targetFeatures = this.extractTargetVector(target);
        
        sequences.push(sequenceFeatures);
        targets.push(targetFeatures);
      }

      console.log(`📈 Collected ${sequences.length} training sequences`);
      return { sequences, targets };
    } catch (error) {
      console.error('❌ Error collecting training data:', error.message);
      return { sequences: [], targets: [] };
    }
  }

  extractFeatureVector(metrics) {
    return this.config.features.map(feature => {
      const value = metrics[feature];
      return typeof value === 'number' ? value : 0;
    });
  }

  extractTargetVector(metrics) {
    // Predict key metrics that drive scaling decisions
    return [
      metrics.active_agents,
      metrics.queue_length,
      metrics.avg_response_time,
      metrics.cpu_utilization
    ];
  }

  async createModel() {
    const inputShape = [this.config.sequenceLength, this.config.features.length];
    const outputSize = 4; // Predict 4 target metrics

    const model = tf.sequential({
      layers: [
        // First LSTM layer
        tf.layers.lstm({
          units: 128,
          returnSequences: true,
          inputShape: inputShape,
          dropout: 0.2,
          recurrentDropout: 0.2
        }),
        
        // Second LSTM layer
        tf.layers.lstm({
          units: 64,
          returnSequences: false,
          dropout: 0.2,
          recurrentDropout: 0.2
        }),
        
        // Dense layers
        tf.layers.dense({
          units: 32,
          activation: 'relu'
        }),
        tf.layers.dropout({ rate: 0.3 }),
        
        tf.layers.dense({
          units: 16,
          activation: 'relu'
        }),
        
        // Output layer
        tf.layers.dense({
          units: outputSize,
          activation: 'linear' // Regression output
        })
      ]
    });

    model.compile({
      optimizer: tf.train.adam(this.config.learningRate),
      loss: 'meanSquaredError',
      metrics: ['mae']
    });

    return model;
  }

  normalizeData(data) {
    const tensor = tf.tensor(data);
    const mean = tensor.mean(0);
    const std = tensor.std(0);
    
    // Avoid division by zero
    const normalizedStd = std.add(tf.scalar(1e-8));
    const normalized = tensor.sub(mean).div(normalizedStd);
    
    return {
      normalized: normalized,
      mean: mean,
      std: normalizedStd
    };
  }

  async trainModel() {
    if (this.isTraining) {
      console.log('🔄 Model training already in progress');
      return;
    }

    this.isTraining = true;
    const startTime = performance.now();

    try {
      console.log('🚀 Starting LSTM model training...');
      
      const { sequences, targets } = await this.collectTrainingData();
      
      if (sequences.length < 100) {
        console.log('⚠️  Insufficient training data, creating default model');
        this.model = await this.createModel();
        return;
      }

      // Normalize data
      const { normalized: normalizedSequences, mean: seqMean, std: seqStd } = this.normalizeData(sequences);
      const { normalized: normalizedTargets, mean: targMean, std: targStd } = this.normalizeData(targets);

      // Store scalers for prediction
      this.scaler = {
        sequences: { mean: seqMean, std: seqStd },
        targets: { mean: targMean, std: targStd }
      };

      // Convert to tensors
      const xTrain = normalizedSequences;
      const yTrain = normalizedTargets;

      // Create and train model
      this.model = await this.createModel();
      
      const history = await this.model.fit(xTrain, yTrain, {
        epochs: this.config.epochs,
        batchSize: this.config.batchSize,
        validationSplit: this.config.validationSplit,
        callbacks: {
          onEpochEnd: (epoch, logs) => {
            if (epoch % 10 === 0) {
              console.log(`Epoch ${epoch}: loss = ${logs.loss.toFixed(4)}, val_loss = ${logs.val_loss.toFixed(4)}`);
            }
          }
        }
      });

      const trainingTime = performance.now() - startTime;
      this.lastTrainingTime = Date.now();
      
      const finalLoss = history.history.val_loss[history.history.val_loss.length - 1];
      this.modelAccuracy = 1 - Math.min(finalLoss, 1); // Convert loss to accuracy estimate

      console.log(`✅ LSTM model training completed:`);
      console.log(`   📈 Final validation loss: ${finalLoss.toFixed(4)}`);
      console.log(`   🎯 Estimated accuracy: ${(this.modelAccuracy * 100).toFixed(2)}%`);
      console.log(`   🕐 Training time: ${(trainingTime / 1000).toFixed(1)}s`);
      console.log(`   📊 Training samples: ${sequences.length}`);

      // Store model metadata
      await this.redis.hset('ml:predictive_scaling_model', {
        accuracy: this.modelAccuracy,
        trainingTime: trainingTime,
        trainingSamples: sequences.length,
        lastTrained: this.lastTrainingTime,
        finalLoss: finalLoss,
        version: '1.0'
      });

      // Cleanup tensors
      normalizedSequences.dispose();
      normalizedTargets.dispose();

    } catch (error) {
      console.error('❌ LSTM model training failed:', error.message);
      throw error;
    } finally {
      this.isTraining = false;
    }
  }

  async makePrediction(sequence) {
    if (!this.model || !this.scaler) {
      console.log('⚠️  Model or scaler not ready');
      return null;
    }

    try {
      // Convert sequence to feature vectors
      const features = sequence.map(m => this.extractFeatureVector(m));
      
      // Normalize input
      const inputTensor = tf.tensor([features]);
      const normalizedInput = inputTensor.sub(this.scaler.sequences.mean).div(this.scaler.sequences.std);
      
      // Make prediction
      const prediction = this.model.predict(normalizedInput);
      
      // Denormalize output
      const denormalized = prediction.mul(this.scaler.targets.std).add(this.scaler.targets.mean);
      const result = await denormalized.data();
      
      // Cleanup tensors
      inputTensor.dispose();
      normalizedInput.dispose();
      prediction.dispose();
      denormalized.dispose();

      return {
        predicted_active_agents: Math.max(0, result[0]),
        predicted_queue_length: Math.max(0, result[1]),
        predicted_avg_response_time: Math.max(0, result[2]),
        predicted_cpu_utilization: Math.max(0, Math.min(1, result[3])),
        confidence: this.modelAccuracy,
        timestamp: Date.now()
      };
    } catch (error) {
      console.error('❌ Prediction failed:', error.message);
      return null;
    }
  }

  async makePredictionAndScale() {
    try {
      const prediction = await this.makePrediction(this.currentSequence);
      if (!prediction) return;

      console.log(`🔮 Prediction: agents=${prediction.predicted_active_agents.toFixed(1)}, queue=${prediction.predicted_queue_length.toFixed(1)}, response=${prediction.predicted_avg_response_time.toFixed(1)}ms`);
      
      // Calculate required agent count based on predictions
      const requiredAgents = this.calculateRequiredAgents(prediction);
      const currentAgents = await this.getActiveAgentsCount();
      
      // Make scaling decision
      const scalingDecision = this.makeScalingDecision(currentAgents, requiredAgents, prediction);
      
      if (scalingDecision.action !== 'none') {
        await this.executeScaling(scalingDecision);
      }

      // Store prediction and decision
      await this.redis.lpush('ml:scaling_predictions', JSON.stringify({
        prediction,
        currentAgents,
        requiredAgents,
        scalingDecision,
        timestamp: Date.now()
      }));
      await this.redis.ltrim('ml:scaling_predictions', 0, 1000);

    } catch (error) {
      console.error('❌ Prediction and scaling failed:', error.message);
    }
  }

  calculateRequiredAgents(prediction) {
    // Multi-factor agent requirement calculation
    
    // Base requirement from queue length and response time
    let requiredAgents = Math.max(
      prediction.predicted_queue_length * 0.5, // One agent per 2 tasks in queue
      prediction.predicted_avg_response_time / 1000 // One agent per second of response time
    );

    // Adjust for CPU utilization
    if (prediction.predicted_cpu_utilization > 0.8) {
      requiredAgents *= 1.5; // Scale up if high CPU predicted
    }

    // Apply bounds
    requiredAgents = Math.max(this.config.minAgents, Math.min(this.config.maxAgents, Math.ceil(requiredAgents)));
    
    return requiredAgents;
  }

  makeScalingDecision(currentAgents, requiredAgents, prediction) {
    const difference = requiredAgents - currentAgents;
    const relativeChange = Math.abs(difference) / currentAgents;

    // Scaling thresholds
    if (difference > 0 && relativeChange > 0.2) {
      // Scale up if we need 20% more agents
      return {
        action: 'scale_up',
        targetCount: requiredAgents,
        reason: `Predicted load increase: queue=${prediction.predicted_queue_length.toFixed(1)}, response_time=${prediction.predicted_avg_response_time.toFixed(1)}ms`,
        confidence: prediction.confidence
      };
    } else if (difference < -2 && relativeChange > 0.3) {
      // Scale down if we have 30% too many agents (minimum decrease of 2)
      return {
        action: 'scale_down',
        targetCount: requiredAgents,
        reason: `Predicted load decrease: lower queue and response time expected`,
        confidence: prediction.confidence
      };
    }

    return {
      action: 'none',
      targetCount: currentAgents,
      reason: 'Current capacity sufficient for predicted load',
      confidence: prediction.confidence
    };
  }

  async executeScaling(scalingDecision) {
    try {
      console.log(`⚖️  Executing scaling: ${scalingDecision.action} to ${scalingDecision.targetCount} agents`);
      console.log(`   📋 Reason: ${scalingDecision.reason}`);
      console.log(`   🎯 Confidence: ${(scalingDecision.confidence * 100).toFixed(1)}%`);

      // Store scaling action
      const scalingAction = {
        action: scalingDecision.action,
        targetCount: scalingDecision.targetCount,
        reason: scalingDecision.reason,
        confidence: scalingDecision.confidence,
        timestamp: Date.now()
      };

      await this.redis.lpush('ml:scaling_actions', JSON.stringify(scalingAction));
      await this.redis.ltrim('ml:scaling_actions', 0, 500);

      // Execute scaling via swarm management
      await this.redis.publish('swarm:scaling_command', JSON.stringify({
        command: scalingDecision.action,
        targetAgentCount: scalingDecision.targetCount,
        reason: scalingDecision.reason,
        source: 'predictive_scaling'
      }));

      this.scalingHistory.push(scalingAction);
      
      console.log(`✅ Scaling command sent: ${scalingDecision.action} to ${scalingDecision.targetCount} agents`);

    } catch (error) {
      console.error('❌ Scaling execution failed:', error.message);
    }
  }

  startModelUpdateScheduler() {
    setInterval(async () => {
      try {
        await this.trainModel();
      } catch (error) {
        console.error('❌ Scheduled model update failed:', error.message);
      }
    }, this.config.modelUpdateInterval);

    console.log(`🕐 Model update scheduler started (${this.config.modelUpdateInterval / 60000}min interval)`);
  }

  async getSystemStatus() {
    const status = {
      modelTrained: !!this.model,
      scalerReady: !!this.scaler,
      accuracy: this.modelAccuracy,
      lastTrainingTime: this.lastTrainingTime,
      isTraining: this.isTraining,
      currentSequenceLength: this.currentSequence.length,
      requiredSequenceLength: this.config.sequenceLength,
      scalingHistory: this.scalingHistory.slice(-5), // Last 5 scaling actions
      config: {
        sequenceLength: this.config.sequenceLength,
        predictionHorizon: this.config.predictionHorizon,
        features: this.config.features,
        modelUpdateInterval: this.config.modelUpdateInterval
      }
    };

    try {
      const redisStats = await this.redis.hgetall('ml:predictive_scaling_model');
      status.persistedStats = redisStats;
      
      const recentPredictions = await this.redis.lrange('ml:scaling_predictions', 0, 4);
      status.recentPredictions = recentPredictions.map(p => JSON.parse(p));
      
      const recentActions = await this.redis.lrange('ml:scaling_actions', 0, 4);
      status.recentActions = recentActions.map(a => JSON.parse(a));
      
    } catch (error) {
      status.persistedStats = {};
      status.recentPredictions = [];
      status.recentActions = [];
    }

    return status;
  }

  async shutdown() {
    console.log('🔄 Shutting down Predictive Scaling System...');
    
    if (this.model) {
      this.model.dispose();
    }
    
    if (this.scaler) {
      this.scaler.sequences.mean.dispose();
      this.scaler.sequences.std.dispose();
      this.scaler.targets.mean.dispose();
      this.scaler.targets.std.dispose();
    }
    
    if (this.redis) {
      await this.redis.disconnect();
    }
    
    console.log('✅ Predictive Scaling System shutdown complete');
  }
}

module.exports = PredictiveScalingSystem;

// Start system if run directly
if (require.main === module) {
  const system = new PredictiveScalingSystem();
  
  // Initialize and train model
  system.trainModel().catch(console.error);
  
  // Graceful shutdown
  process.on('SIGTERM', async () => {
    await system.shutdown();
    process.exit(0);
  });
  
  process.on('SIGINT', async () => {
    await system.shutdown();
    process.exit(0);
  });
}