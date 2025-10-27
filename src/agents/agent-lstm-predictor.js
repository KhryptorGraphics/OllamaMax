/**
 * Agent-Specific LSTM Load Predictor
 *
 * LSTM model for predicting individual agent load and performance.
 * Optimized for per-agent predictions with 30-step sequences.
 *
 * References:
 * - src/ml/predictive-scaling.js
 * - src/ml/feature-store.js
 */

const tf = require('@tensorflow/tfjs-node');
const Redis = require('ioredis');

class AgentLSTMPredictor {
  constructor() {
    // Redis cluster configuration - use environment variables for flexibility
    const redisNodes = process.env.REDIS_NODES
      ? JSON.parse(process.env.REDIS_NODES)
      : [
          { host: 'redis-cluster-0.redis-cluster-service.ollamamax-redis', port: 6379 },
          { host: 'redis-cluster-1.redis-cluster-service.ollamamax-redis', port: 6379 },
          { host: 'redis-cluster-2.redis-cluster-service.ollamamax-redis', port: 6379 }
        ];

    this.redis = new Redis.Cluster(redisNodes, {
      redisOptions: {
        password: process.env.REDIS_PASSWORD || 'ollama_redis_pass',
        connectTimeout: 10000
      }
    });

    // Model configuration
    this.sequenceLength = 30; // 30 minutes of agent activity
    this.predictionHorizon = 15; // Predict next 15 minutes
    this.features = [
      'agent_active_tasks',
      'agent_cpu_usage',
      'agent_memory_usage',
      'agent_success_rate',
      'agent_response_time',
      'task_complexity_avg',
      'hour_of_day'
    ];
    this.featureCount = this.features.length;

    // Per-agent model cache
    this.models = new Map(); // agentId -> model
    this.modelVersions = new Map(); // agentId -> version
    this.predictionCache = new Map(); // agentId -> { prediction, timestamp }
    this.cacheTTL = 60000; // 1 minute cache TTL
  }

  /**
   * Create LSTM model for per-agent predictions
   * 2-layer LSTM (64 + 32 units) optimized for agent load
   */
  createModel() {
    const model = tf.sequential();

    // First LSTM layer (64 units)
    model.add(tf.layers.lstm({
      units: 64,
      returnSequences: true,
      inputShape: [this.sequenceLength, this.featureCount]
    }));
    model.add(tf.layers.dropout({ rate: 0.2 }));

    // Second LSTM layer (32 units)
    model.add(tf.layers.lstm({
      units: 32,
      returnSequences: false
    }));
    model.add(tf.layers.dropout({ rate: 0.2 }));

    // Dense output layer (predict task count)
    model.add(tf.layers.dense({ units: 1, activation: 'relu' }));

    model.compile({
      optimizer: tf.train.adam(0.001),
      loss: 'meanSquaredError',
      metrics: ['mae']
    });

    return model;
  }

  /**
   * Collect training data for specific agent
   *
   * @param {string} agentId - Agent identifier
   * @returns {Object} Training data { sequences, targets }
   */
  async collectTrainingData(agentId) {
    try {
      // Get agent historical metrics from Redis
      const metricsKey = `agent:${agentId}:metrics:timeseries`;
      const metricsData = await this.redis.lrange(metricsKey, 0, -1);

      if (metricsData.length < this.sequenceLength + 1) {
        console.log(`[Agent LSTM] Insufficient data for agent ${agentId}: ${metricsData.length} points`);
        return null;
      }

      const metrics = metricsData.map(m => JSON.parse(m));

      // Build sequences
      const sequences = [];
      const targets = [];

      for (let i = 0; i <= metrics.length - this.sequenceLength - 1; i++) {
        const sequence = metrics.slice(i, i + this.sequenceLength);
        const target = metrics[i + this.sequenceLength];

        // Extract features for sequence
        const sequenceFeatures = sequence.map(m => [
          m.active_tasks || 0,
          m.cpu_usage || 0,
          m.memory_usage || 0,
          m.success_rate || 0,
          m.response_time || 0,
          m.task_complexity_avg || 0.5,
          new Date(m.timestamp).getHours() / 24 // Normalize hour
        ]);

        sequences.push(sequenceFeatures);
        targets.push(target.active_tasks || 0);
      }

      return { sequences, targets };
    } catch (error) {
      console.error(`[Agent LSTM] Error collecting training data for ${agentId}:`, error);
      return null;
    }
  }

  /**
   * Train LSTM model for specific agent
   *
   * @param {string} agentId - Agent identifier
   * @returns {Object} Training results
   */
  async trainAgentModel(agentId) {
    try {
      console.log(`[Agent LSTM] Training model for agent ${agentId}...`);

      // Collect training data
      const trainingData = await this.collectTrainingData(agentId);
      if (!trainingData) {
        return { success: false, error: 'Insufficient training data' };
      }

      const { sequences, targets } = trainingData;

      // Create tensors
      const xTrain = tf.tensor3d(sequences);
      const yTrain = tf.tensor2d(targets, [targets.length, 1]);

      // Create or get model
      let model = this.models.get(agentId);
      if (!model) {
        model = this.createModel();
        this.models.set(agentId, model);
      }

      // Train model
      const history = await model.fit(xTrain, yTrain, {
        epochs: 50,
        batchSize: 32,
        validationSplit: 0.2,
        verbose: 0,
        callbacks: {
          onEpochEnd: (epoch, logs) => {
            if (epoch % 10 === 0) {
              console.log(`[Agent LSTM] Epoch ${epoch}: loss=${logs.loss.toFixed(4)}, mae=${logs.mae.toFixed(4)}`);
            }
          }
        }
      });

      // Cleanup tensors
      xTrain.dispose();
      yTrain.dispose();

      // Update model version
      const version = (this.modelVersions.get(agentId) || 0) + 1;
      this.modelVersions.set(agentId, version);

      // Store model in Redis
      const modelKey = `ml:agent_lstm:${agentId}:model`;
      await this.redis.set(modelKey, JSON.stringify({
        version,
        trained_at: Date.now(),
        training_samples: sequences.length,
        final_loss: history.history.loss[history.history.loss.length - 1],
        final_mae: history.history.mae[history.history.mae.length - 1]
      }));

      console.log(`[Agent LSTM] Training completed for ${agentId}, version ${version}`);

      return {
        success: true,
        version,
        trainingSamples: sequences.length,
        finalLoss: history.history.loss[history.history.loss.length - 1],
        finalMAE: history.history.mae[history.history.mae.length - 1]
      };

    } catch (error) {
      console.error(`[Agent LSTM] Training error for ${agentId}:`, error);
      return { success: false, error: error.message };
    }
  }

  /**
   * Predict agent load for next 15 minutes
   *
   * @param {string} agentId - Agent identifier
   * @param {Array} currentSequence - Current 30-step sequence
   * @returns {Object} Prediction { load, confidence }
   */
  async predictAgentLoad(agentId, currentSequence = null) {
    try {
      // Check prediction cache
      const cached = this.predictionCache.get(agentId);
      if (cached && Date.now() - cached.timestamp < this.cacheTTL) {
        return cached.prediction;
      }

      // Get model
      let model = this.models.get(agentId);
      if (!model) {
        // Try to train model
        const trainResult = await this.trainAgentModel(agentId);
        if (!trainResult.success) {
          return { load: 0, confidence: 0, error: 'Model not available' };
        }
        model = this.models.get(agentId);
      }

      // Get current sequence if not provided
      if (!currentSequence) {
        const metricsKey = `agent:${agentId}:metrics:timeseries`;
        const metricsData = await this.redis.lrange(metricsKey, -this.sequenceLength, -1);

        if (metricsData.length < this.sequenceLength) {
          return { load: 0, confidence: 0, error: 'Insufficient recent data' };
        }

        currentSequence = metricsData.map(m => {
          const metrics = JSON.parse(m);
          return [
            metrics.active_tasks || 0,
            metrics.cpu_usage || 0,
            metrics.memory_usage || 0,
            metrics.success_rate || 0,
            metrics.response_time || 0,
            metrics.task_complexity_avg || 0.5,
            new Date(metrics.timestamp).getHours() / 24
          ];
        });
      }

      // Make prediction
      const inputTensor = tf.tensor3d([currentSequence]);
      const predictionTensor = model.predict(inputTensor);
      const prediction = await predictionTensor.data();

      inputTensor.dispose();
      predictionTensor.dispose();

      const predictedLoad = Math.max(0, Math.round(prediction[0]));
      const confidence = 0.85; // TODO: Calculate based on historical accuracy

      const result = { load: predictedLoad, confidence };

      // Cache prediction
      this.predictionCache.set(agentId, { prediction: result, timestamp: Date.now() });

      // Store prediction in Redis
      await this.redis.setex(
        `ml:agent_lstm:${agentId}:prediction`,
        300, // 5 minute expiry
        JSON.stringify({
          predicted_load: predictedLoad,
          confidence,
          timestamp: Date.now()
        })
      );

      return result;

    } catch (error) {
      console.error(`[Agent LSTM] Prediction error for ${agentId}:`, error);
      return { load: 0, confidence: 0, error: error.message };
    }
  }

  /**
   * Predict agent availability
   *
   * @param {string} agentId - Agent identifier
   * @returns {Object} { available: boolean, probability: number }
   */
  async predictAgentAvailability(agentId) {
    const loadPrediction = await this.predictAgentLoad(agentId);

    // Get agent capacity
    const agentData = await this.redis.hgetall(`agent:${agentId}`);
    const maxCapacity = parseInt(agentData.max_capacity || 5);

    const available = loadPrediction.load < maxCapacity;
    const probability = Math.max(0, Math.min(1, 1 - (loadPrediction.load / maxCapacity)));

    return { available, probability, predictedLoad: loadPrediction.load };
  }

  /**
   * Predict agent performance for task type
   *
   * @param {string} agentId - Agent identifier
   * @param {string} taskType - Task type
   * @returns {Object} { successRate: number, estimatedDuration: number }
   */
  async predictAgentPerformance(agentId, taskType) {
    try {
      // Get historical performance for this task type
      const performanceKey = `agent:${agentId}:performance:${taskType}`;
      const performanceData = await this.redis.hgetall(performanceKey);

      const successRate = parseFloat(performanceData.success_rate || 0.5);
      const avgDuration = parseInt(performanceData.avg_duration || 1000);

      // Adjust based on current load
      const loadPrediction = await this.predictAgentLoad(agentId);
      const loadFactor = 1 + (loadPrediction.load * 0.1); // 10% increase per task

      return {
        successRate,
        estimatedDuration: Math.round(avgDuration * loadFactor),
        confidence: loadPrediction.confidence
      };
    } catch (error) {
      console.error(`[Agent LSTM] Performance prediction error:`, error);
      return { successRate: 0.5, estimatedDuration: 1000, confidence: 0 };
    }
  }

  /**
   * Batch predictions for multiple agents
   *
   * @param {Array<string>} agentIds - Agent identifiers
   * @returns {Map} agentId -> prediction
   */
  async batchPredict(agentIds) {
    const predictions = new Map();

    // TODO: Optimize with actual batching
    for (const agentId of agentIds) {
      predictions.set(agentId, await this.predictAgentLoad(agentId));
    }

    return predictions;
  }

  /**
   * Test agent LSTM predictor
   */
  async test() {
    console.log('[Agent LSTM] Running tests...');

    try {
      // Create test agent data
      const testAgentId = 'test-agent-lstm';
      const metricsKey = `agent:${testAgentId}:metrics:timeseries`;

      // Generate synthetic time series (50 data points)
      const testData = [];
      for (let i = 0; i < 50; i++) {
        testData.push(JSON.stringify({
          active_tasks: Math.floor(Math.random() * 5),
          cpu_usage: Math.random() * 0.8,
          memory_usage: Math.random() * 0.7,
          success_rate: 0.8 + Math.random() * 0.2,
          response_time: 500 + Math.random() * 500,
          task_complexity_avg: Math.random(),
          timestamp: Date.now() - (50 - i) * 60000 // 1 minute intervals
        }));
      }

      // Store test data
      await this.redis.del(metricsKey);
      for (const data of testData) {
        await this.redis.rpush(metricsKey, data);
      }

      // Train model
      console.log('[Agent LSTM] Training test model...');
      const trainResult = await this.trainAgentModel(testAgentId);
      console.log('[Agent LSTM] Training result:', trainResult);

      if (trainResult.success) {
        // Test prediction
        console.log('[Agent LSTM] Testing prediction...');
        const prediction = await this.predictAgentLoad(testAgentId);
        console.log('[Agent LSTM] Prediction:', prediction);

        // Test availability
        const availability = await this.predictAgentAvailability(testAgentId);
        console.log('[Agent LSTM] Availability:', availability);

        // Test performance prediction
        const performance = await this.predictAgentPerformance(testAgentId, 'code-generation');
        console.log('[Agent LSTM] Performance:', performance);
      }

      // Cleanup
      await this.redis.del(metricsKey);
      this.models.delete(testAgentId);

      console.log('[Agent LSTM] Test completed successfully');
      return true;

    } catch (error) {
      console.error('[Agent LSTM] Test failed:', error);
      return false;
    }
  }

  /**
   * Close connections and cleanup
   */
  async close() {
    // Dispose all models
    for (const [agentId, model] of this.models.entries()) {
      model.dispose();
    }
    this.models.clear();

    if (this.redis) {
      await this.redis.quit();
    }
  }
}

// CLI support
if (require.main === module) {
  const predictor = new AgentLSTMPredictor();

  const command = process.argv[2];

  if (command === 'train') {
    const agentId = process.argv[3];
    if (!agentId) {
      console.error('Usage: node agent-lstm-predictor.js train <agent-id>');
      process.exit(1);
    }
    predictor.trainAgentModel(agentId).then(() => predictor.close());
  } else if (command === 'predict') {
    const agentId = process.argv[3];
    if (!agentId) {
      console.error('Usage: node agent-lstm-predictor.js predict <agent-id>');
      process.exit(1);
    }
    predictor.predictAgentLoad(agentId).then(result => {
      console.log(JSON.stringify(result, null, 2));
      predictor.close();
    });
  } else if (command === 'test') {
    predictor.test().then(success => {
      predictor.close();
      process.exit(success ? 0 : 1);
    });
  } else {
    console.log('Usage: node agent-lstm-predictor.js [train|predict|test] [agent-id]');
    process.exit(1);
  }
}

module.exports = AgentLSTMPredictor;
