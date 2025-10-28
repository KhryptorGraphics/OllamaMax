/**
 * Unified Neural Training Orchestrator
 * Coordinates all neural training activities across ML models, swarm systems, and Claude-Flow
 */

const AgentLSTMPredictor = require('./agent-lstm-predictor.js');
const NeuralPatternTrainer = require('./neural-pattern-trainer.js');
const HistoricalDataAggregator = require('./historical-data-aggregator.js');
const Redis = require('ioredis');

class UnifiedNeuralOrchestrator {
  constructor() {
    // Read Redis cluster nodes from environment or use defaults
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
        connectTimeout: 10000,
        readyCheck: (callback) => callback()
      },
      enableOfflineQueue: false,
      retryDelayOnFailover: 100,
      maxRetriesPerRequest: 3
    });

    this.components = {
      lstmPredictor: new AgentLSTMPredictor(),
      patternTrainer: new NeuralPatternTrainer(),
      dataAggregator: new HistoricalDataAggregator()
    };

    this.trainingSchedule = {
      agent_lstm: 2 * 3600000,      // 2 hours
      pattern_training: 4 * 3600000, // 4 hours
      full_pipeline: 6 * 3600000     // 6 hours
    };

    this.trainingJobs = new Map();
  }

  async orchestrateFullTraining() {
    console.log('[Neural Orchestrator] Starting full training pipeline...');

    // Verify Redis connectivity
    try {
      await this.redis.ping();
      console.log('✅ Neural Orchestrator connected to Redis cluster');
    } catch (error) {
      console.error('❌ Redis connection failed:', error.message);
      throw error;
    }

    const startTime = Date.now();

    try {
      // Phase 1: Aggregate historical data
      console.log('[Phase 1] Aggregating historical data...');
      await this.components.dataAggregator.syncClaudeFlowMemoryToRedis();
      const data = await this.components.dataAggregator.aggregateHistoricalData(
        Date.now() - 24 * 3600000,
        Date.now()
      );

      // Phase 2: Train agent LSTM models
      console.log('[Phase 2] Training agent LSTM models...');
      const agentIds = await this.getActiveAgents();
      const lstmResults = [];
      for (const agentId of agentIds.slice(0, 10)) { // Limit to 10 agents
        const result = await this.components.lstmPredictor.trainAgentModel(agentId);
        lstmResults.push({ agentId, result });
      }

      // Phase 3: Train neural pattern recognition
      console.log('[Phase 3] Training neural pattern recognition...');
      const patterns = await this.components.patternTrainer.collectPatterns();
      const patternResult = await this.components.patternTrainer.trainOnPatterns(patterns);

      // Phase 4: Validate models
      console.log('[Phase 4] Validating models...');
      const validation = await this.validateAllModels();

      const duration = Date.now() - startTime;

      const result = {
        success: true,
        duration,
        components: {
          dataAggregation: { records: data.length },
          agentLSTM: { trained: lstmResults.filter(r => r.result.success).length, total: lstmResults.length },
          patternTraining: patternResult
        },
        validation,
        timestamp: Date.now()
      };

      await this.redis.setex('neural:training:last_result', 3600, JSON.stringify(result));

      console.log(`[Neural Orchestrator] Training completed in ${(duration / 1000).toFixed(1)}s`);

      return result;
    } catch (error) {
      console.error('[Neural Orchestrator] Training failed:', error);
      return { success: false, error: error.message };
    }
  }

  async trainSpecificModel(modelName) {
    console.log(`[Neural Orchestrator] Training ${modelName}...`);

    switch (modelName) {
      case 'agent-lstm':
        const agentIds = await this.getActiveAgents();
        return await this.components.lstmPredictor.trainAgentModel(agentIds[0]);

      case 'pattern-trainer':
        const patterns = await this.components.patternTrainer.collectPatterns();
        return await this.components.patternTrainer.trainOnPatterns(patterns);

      default:
        return { success: false, error: `Unknown model: ${modelName}` };
    }
  }

  async getTrainingStatus() {
    const lastResult = await this.redis.get('neural:training:last_result');

    if (!lastResult) {
      return { status: 'never_trained', jobs: [] };
    }

    const parsed = JSON.parse(lastResult);

    return {
      status: parsed.success ? 'healthy' : 'failed',
      lastTraining: new Date(parsed.timestamp),
      duration: parsed.duration,
      components: parsed.components,
      jobs: Array.from(this.trainingJobs.entries()).map(([name, job]) => ({
        name,
        status: job.status,
        startTime: job.startTime
      }))
    };
  }

  async validateAllModels() {
    const validation = {
      agent_lstm: { valid: false, accuracy: 0 },
      pattern_trainer: { valid: false, accuracy: 0 }
    };

    try {
      // Validate LSTM predictor
      const agentIds = await this.getActiveAgents();
      if (agentIds.length > 0) {
        const prediction = await this.components.lstmPredictor.predictAgentLoad(agentIds[0]);
        validation.agent_lstm = {
          valid: prediction.confidence > 0.5,
          accuracy: prediction.confidence
        };
      }

      // Validate pattern trainer
      const testPattern = { agentType: 'coder', taskType: 'code-generation', complexity: 'medium', teamSize: 3 };
      const patternPred = await this.components.patternTrainer.predictPatternSuccess(testPattern);
      validation.pattern_trainer = {
        valid: patternPred.confidence > 0.5,
        accuracy: patternPred.confidence
      };

    } catch (error) {
      console.error('[Neural Orchestrator] Validation error:', error);
    }

    return validation;
  }

  async getActiveAgents() {
    const keys = await this.redis.keys('agent:*:status');
    return keys.map(k => k.split(':')[1]).filter(id => !id.includes(':'));
  }

  async close() {
    await this.components.lstmPredictor.close();
    await this.components.patternTrainer.close();
    await this.components.dataAggregator.close();
    if (this.redis) {
      await this.redis.quit();
    }
  }
}

// CLI support
if (require.main === module) {
  const orchestrator = new UnifiedNeuralOrchestrator();
  const command = process.argv[2];

  if (command === 'train') {
    orchestrator.orchestrateFullTraining().then(() => orchestrator.close());
  } else if (command === 'status') {
    orchestrator.getTrainingStatus().then(status => {
      console.log(JSON.stringify(status, null, 2));
      orchestrator.close();
    });
  } else if (command === 'metrics') {
    orchestrator.getTrainingStatus().then(status => {
      console.log('Training Metrics:');
      console.log(`  Status: ${status.status}`);
      console.log(`  Last Training: ${status.lastTraining}`);
      console.log(`  Duration: ${(status.duration / 1000).toFixed(1)}s`);
      orchestrator.close();
    });
  } else {
    console.log('Usage: node unified-neural-orchestrator.js [train|status|metrics]');
    process.exit(1);
  }
}

module.exports = UnifiedNeuralOrchestrator;
