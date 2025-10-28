/**
 * Agent Performance Forecasting System
 *
 * Comprehensive agent performance prediction using ensemble methods.
 * Combines LSTM, Random Forest, and neural learning for accurate forecasts.
 *
 * References:
 * - src/agents/agent-lstm-predictor.js
 * - src/ml/agent-selection-model.js
 * - .claude-flow/agents/neural-learning.js
 * - src/swarm/performance-optimizer.js
 */

const AgentLSTMPredictor = require('./agent-lstm-predictor.js');
const Redis = require('ioredis');

class AgentPerformanceForecaster {
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

    // Component instances
    this.lstmPredictor = new AgentLSTMPredictor();
    this.agentSelectionModel = null; // Lazy load
    this.neuralLearningSystem = null; // Lazy load

    // Ensemble weights
    this.ensembleWeights = {
      lstm: 0.4,        // Short-term predictions
      randomForest: 0.35, // Medium-term predictions
      neuralLearning: 0.25  // Pattern-based predictions
    };

    // Prediction cache
    this.predictionCache = new Map();
    this.cacheTTL = 60000; // 1 minute
  }

  /**
   * Initialize forecaster components
   */
  async initialize() {
    try {
      // Lazy load components
      const AgentSelectionModel = require('../ml/agent-selection-model.js');
      const NeuralLearningSystem = require('../../.claude-flow/agents/neural-learning.js');

      this.agentSelectionModel = new AgentSelectionModel();
      this.neuralLearningSystem = new NeuralLearningSystem();

      await this.agentSelectionModel.initialize();

      console.log('[Agent Forecaster] Initialized successfully');
    } catch (error) {
      console.error('[Agent Forecaster] Initialization error:', error);
      throw error;
    }
  }

  /**
   * Predict agent performance using ensemble methods
   *
   * @param {string} agentId - Agent identifier
   * @param {Object} taskRequest - Task request details
   * @returns {Object} Comprehensive performance forecast
   */
  async predictAgentPerformance(agentId, taskRequest) {
    try {
      // Check cache
      const cacheKey = `${agentId}:${JSON.stringify(taskRequest)}`;
      const cached = this.predictionCache.get(cacheKey);
      if (cached && Date.now() - cached.timestamp < this.cacheTTL) {
        return cached.prediction;
      }

      // Get predictions from each component
      const [lstmPred, rfPred, neuralPred] = await Promise.all([
        this.getLSTMPrediction(agentId, taskRequest),
        this.getRandomForestPrediction(agentId, taskRequest),
        this.getNeuralLearningPrediction(agentId, taskRequest)
      ]);

      // Ensemble prediction
      const ensemblePrediction = this.combinePresictions(lstmPred, rfPred, neuralPred);

      // Calculate confidence
      const confidence = this.calculateConfidence([lstmPred, rfPred, neuralPred]);

      // Get agent behavior insights
      const behaviorInsights = await this.getAgentBehaviorInsights(agentId);

      const forecast = {
        agentId,
        taskRequest,
        prediction: ensemblePrediction,
        confidence,
        components: {
          lstm: lstmPred,
          randomForest: rfPred,
          neuralLearning: neuralPred
        },
        behaviorInsights,
        timestamp: Date.now()
      };

      // Cache prediction
      this.predictionCache.set(cacheKey, { prediction: forecast, timestamp: Date.now() });

      // Store in Redis
      await this.redis.setex(
        `agent:${agentId}:predictions`,
        300, // 5 minute expiry
        JSON.stringify(forecast)
      );

      return forecast;

    } catch (error) {
      console.error('[Agent Forecaster] Prediction error:', error);
      return this.getDefaultPrediction(agentId, taskRequest);
    }
  }

  /**
   * Get LSTM-based prediction (short-term 5-15 min)
   */
  async getLSTMPrediction(agentId, taskRequest) {
    try {
      const [loadPred, availability, performance] = await Promise.all([
        this.lstmPredictor.predictAgentLoad(agentId),
        this.lstmPredictor.predictAgentAvailability(agentId),
        this.lstmPredictor.predictAgentPerformance(agentId, taskRequest.type)
      ]);

      return {
        successRate: performance.successRate,
        estimatedDuration: performance.estimatedDuration,
        predictedLoad: loadPred.load,
        availability: availability.probability,
        confidence: loadPred.confidence,
        horizon: 'short-term' // 5-15 min
      };
    } catch (error) {
      console.error('[Agent Forecaster] LSTM prediction error:', error);
      return { successRate: 0.5, estimatedDuration: 1000, confidence: 0 };
    }
  }

  /**
   * Get Random Forest prediction (medium-term 1-4 hours)
   */
  async getRandomForestPrediction(agentId, taskRequest) {
    try {
      if (!this.agentSelectionModel) {
        await this.initialize();
      }

      // Get available agents list (for model's selectBestAgent method)
      const availableAgents = [agentId];

      // Use the actual Random Forest model prediction
      const modelPrediction = await this.agentSelectionModel.selectBestAgent(taskRequest, availableAgents);

      if (modelPrediction && modelPrediction.agentId === agentId) {
        // Use model confidence as RF success rate
        const successRate = modelPrediction.confidence || 0.5;

        // Get agent stats for duration estimate
        const agentStats = await this.agentSelectionModel.getAgentStats(agentId);
        const baseDuration = agentStats.avgExecutionTime || 1000;

        // Adjust duration by task complexity
        const complexityFactors = {
          'low': 0.7,
          'medium': 1.0,
          'high': 1.5
        };
        const complexityFactor = complexityFactors[taskRequest.complexity] || 1.0;
        const estimatedDuration = Math.round(baseDuration * complexityFactor);

        return {
          successRate,
          estimatedDuration,
          confidence: modelPrediction.confidence || 0.8,
          horizon: 'medium-term' // 1-4 hours
        };
      }

      // Fallback to agent stats-based estimation
      const agentStats = await this.agentSelectionModel.getAgentStats(agentId);
      const successRate = agentStats.successRate || 0.5;
      const avgDuration = agentStats.avgExecutionTime || 1000;

      const complexityFactor = {
        'low': 0.7,
        'medium': 1.0,
        'high': 1.5
      }[taskRequest.complexity] || 1.0;

      return {
        successRate,
        estimatedDuration: Math.round(avgDuration * complexityFactor),
        confidence: 0.6, // Lower confidence for fallback
        horizon: 'medium-term'
      };
    } catch (error) {
      console.error('[Agent Forecaster] Random Forest prediction error:', error);
      return { successRate: 0.5, estimatedDuration: 1000, confidence: 0 };
    }
  }

  /**
   * Get Neural Learning prediction (pattern-based)
   */
  async getNeuralLearningPrediction(agentId, taskRequest) {
    try {
      if (!this.neuralLearningSystem) {
        await this.initialize();
      }

      // Get learned patterns for agent
      const patterns = await this.neuralLearningSystem.getAgentPatterns(agentId, taskRequest.type);

      if (!patterns || patterns.length === 0) {
        return { successRate: 0.5, estimatedDuration: 1000, confidence: 0 };
      }

      // Calculate success rate from patterns
      const successfulPatterns = patterns.filter(p => p.success).length;
      const successRate = successfulPatterns / patterns.length;

      // Calculate average duration
      const avgDuration = patterns.reduce((sum, p) => sum + (p.duration || 1000), 0) / patterns.length;

      return {
        successRate,
        estimatedDuration: Math.round(avgDuration),
        patternCount: patterns.length,
        confidence: Math.min(0.9, patterns.length / 10), // More patterns = higher confidence
        horizon: 'pattern-based'
      };
    } catch (error) {
      console.error('[Agent Forecaster] Neural Learning prediction error:', error);
      return { successRate: 0.5, estimatedDuration: 1000, confidence: 0 };
    }
  }

  /**
   * Combine predictions using weighted ensemble
   */
  combinePresictions(lstmPred, rfPred, neuralPred) {
    const successRate = (
      lstmPred.successRate * this.ensembleWeights.lstm +
      rfPred.successRate * this.ensembleWeights.randomForest +
      neuralPred.successRate * this.ensembleWeights.neuralLearning
    );

    const estimatedDuration = Math.round(
      lstmPred.estimatedDuration * this.ensembleWeights.lstm +
      rfPred.estimatedDuration * this.ensembleWeights.randomForest +
      neuralPred.estimatedDuration * this.ensembleWeights.neuralLearning
    );

    return { successRate, estimatedDuration };
  }

  /**
   * Calculate ensemble confidence
   */
  calculateConfidence(predictions) {
    // Weighted average of component confidences
    const totalConfidence = predictions.reduce((sum, pred, idx) => {
      const weight = Object.values(this.ensembleWeights)[idx];
      return sum + (pred.confidence || 0) * weight;
    }, 0);

    // Check prediction agreement (lower variance = higher confidence)
    const successRates = predictions.map(p => p.successRate);
    const variance = this.calculateVariance(successRates);
    const agreementFactor = Math.max(0, 1 - variance);

    return Math.min(0.95, totalConfidence * agreementFactor);
  }

  /**
   * Calculate variance of values
   */
  calculateVariance(values) {
    const mean = values.reduce((sum, val) => sum + val, 0) / values.length;
    const squaredDiffs = values.map(val => Math.pow(val - mean, 2));
    return squaredDiffs.reduce((sum, val) => sum + val, 0) / values.length;
  }

  /**
   * Get agent behavior insights
   */
  async getAgentBehaviorInsights(agentId) {
    try {
      const agentData = await this.redis.hgetall(`agent:${agentId}`);

      // Calculate fatigue (performance degradation)
      const activeTime = parseInt(agentData.active_time || 0);
      const fatigueLevel = Math.min(1, activeTime / (4 * 3600000)); // 4 hours = max fatigue

      // Calculate learning curve (improvement rate)
      const taskCount = parseInt(agentData.task_count || 0);
      const learningProgress = Math.min(1, taskCount / 100); // 100 tasks = fully learned

      // Optimal task types
      const specialization = agentData.specialization || 'general';

      return {
        fatigueLevel,
        learningProgress,
        optimalTaskTypes: [specialization],
        recommendedLoadLevel: fatigueLevel > 0.7 ? 'low' : 'normal'
      };
    } catch (error) {
      console.error('[Agent Forecaster] Behavior insights error:', error);
      return { fatigueLevel: 0, learningProgress: 0.5, optimalTaskTypes: ['general'] };
    }
  }

  /**
   * Rank agents by predicted performance
   *
   * @param {Object} taskRequest - Task request
   * @param {Array<string>} availableAgents - Available agent IDs
   * @returns {Array} Ranked agents with predictions
   */
  async rankAgentsByPredictedPerformance(taskRequest, availableAgents) {
    try {
      // Get predictions for all agents
      const predictions = await Promise.all(
        availableAgents.map(agentId => this.predictAgentPerformance(agentId, taskRequest))
      );

      // Rank by combined score (success rate * confidence / duration)
      const rankedAgents = predictions
        .map(pred => ({
          agentId: pred.agentId,
          score: (pred.prediction.successRate * pred.confidence) / (pred.prediction.estimatedDuration / 1000),
          prediction: pred
        }))
        .sort((a, b) => b.score - a.score);

      return rankedAgents;
    } catch (error) {
      console.error('[Agent Forecaster] Ranking error:', error);
      return availableAgents.map(agentId => ({ agentId, score: 0.5 }));
    }
  }

  /**
   * Get agent load forecast for time horizon
   *
   * @param {string} agentId - Agent identifier
   * @param {string} horizon - 'short' | 'medium' | 'long'
   * @returns {Object} Load forecast
   */
  async getAgentLoadForecast(agentId, horizon = 'short') {
    try {
      const loadPred = await this.lstmPredictor.predictAgentLoad(agentId);

      // Adjust for horizon
      const horizonFactors = {
        'short': 1.0,   // 5-15 min
        'medium': 1.2,  // 1-4 hours
        'long': 1.5     // 1-7 days
      };

      const adjustedLoad = Math.round(loadPred.load * (horizonFactors[horizon] || 1.0));

      return {
        agentId,
        horizon,
        predictedLoad: adjustedLoad,
        confidence: loadPred.confidence,
        timestamp: Date.now()
      };
    } catch (error) {
      console.error('[Agent Forecaster] Load forecast error:', error);
      return { agentId, horizon, predictedLoad: 0, confidence: 0 };
    }
  }

  /**
   * Get optimal agent for task
   *
   * @param {Object} taskRequest - Task request
   * @returns {Object} Best agent recommendation with alternatives
   */
  async getOptimalAgentForTask(taskRequest) {
    try {
      // Get all available agents
      const agentKeys = await this.redis.keys('agent:*:status');
      const availableAgents = agentKeys
        .map(key => key.split(':')[1])
        .filter(id => !id.includes(':'));

      // Rank agents
      const rankedAgents = await this.rankAgentsByPredictedPerformance(taskRequest, availableAgents);

      if (rankedAgents.length === 0) {
        return { primary: null, alternatives: [], reason: 'No agents available' };
      }

      return {
        primary: rankedAgents[0],
        alternatives: rankedAgents.slice(1, 4), // Top 3 alternatives
        totalCandidates: rankedAgents.length,
        timestamp: Date.now()
      };
    } catch (error) {
      console.error('[Agent Forecaster] Optimal agent selection error:', error);
      return { primary: null, alternatives: [], reason: error.message };
    }
  }

  /**
   * Get default prediction (fallback)
   */
  getDefaultPrediction(agentId, taskRequest) {
    return {
      agentId,
      taskRequest,
      prediction: { successRate: 0.5, estimatedDuration: 1000 },
      confidence: 0,
      components: {},
      behaviorInsights: { fatigueLevel: 0, learningProgress: 0.5, optimalTaskTypes: ['general'] },
      timestamp: Date.now()
    };
  }

  /**
   * Close connections
   */
  async close() {
    await this.lstmPredictor.close();
    if (this.redis) {
      await this.redis.quit();
    }
  }
}

// CLI support
if (require.main === module) {
  const forecaster = new AgentPerformanceForecaster();

  const command = process.argv[2];

  if (command === 'predict') {
    const agentId = process.argv[3];
    const taskType = process.argv[4] || 'general';

    if (!agentId) {
      console.error('Usage: node agent-performance-forecaster.js predict <agent-id> [task-type]');
      process.exit(1);
    }

    forecaster.initialize().then(() => {
      return forecaster.predictAgentPerformance(agentId, { type: taskType, complexity: 'medium' });
    }).then(result => {
      console.log(JSON.stringify(result, null, 2));
      forecaster.close();
    });
  } else {
    console.log('Usage: node agent-performance-forecaster.js predict <agent-id> [task-type]');
    process.exit(1);
  }
}

module.exports = AgentPerformanceForecaster;
