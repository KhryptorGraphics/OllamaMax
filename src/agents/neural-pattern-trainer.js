/**
 * Neural Pattern Training Orchestrator
 * Trains neural networks on patterns extracted from Claude-Flow hooks
 */

const tf = require('@tensorflow/tfjs-node');
const Redis = require('ioredis');
const fs = require('fs').promises;
const path = require('path');

class NeuralPatternTrainer {
  constructor() {
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

    this.model = null;
    this.patternCategories = ['coordination', 'execution', 'communication', 'resource', 'optimization'];
    this.minConfidence = 0.7;
  }

  createModel() {
    const model = tf.sequential();

    model.add(tf.layers.dense({ units: 128, activation: 'relu', inputShape: [20] }));
    model.add(tf.layers.dropout({ rate: 0.3 }));
    model.add(tf.layers.dense({ units: 64, activation: 'relu' }));
    model.add(tf.layers.dropout({ rate: 0.2 }));
    model.add(tf.layers.dense({ units: 32, activation: 'relu' }));
    // Use linear activation for regression outputs (success_prob, duration, resources)
    model.add(tf.layers.dense({ units: 3, activation: 'linear' }));

    model.compile({
      optimizer: tf.train.adam(0.001),
      loss: 'meanSquaredError',
      metrics: ['mae']
    });

    return model;
  }

  async collectPatterns() {
    try {
      const memoryPath = path.join(process.cwd(), '.claude-flow/memory/performance-patterns.json');
      const fileData = await fs.readFile(memoryPath, 'utf-8');
      const patterns = JSON.parse(fileData);

      const redisPatterns = await this.redis.keys('patterns:*');
      for (const key of redisPatterns) {
        const data = await this.redis.lrange(key, 0, -1);
        patterns.push(...data.map(d => JSON.parse(d)));
      }

      return patterns;
    } catch (error) {
      console.error('[Neural Pattern Trainer] Pattern collection error:', error);
      return [];
    }
  }

  async trainOnPatterns(patterns) {
    if (patterns.length < 10) {
      return { success: false, error: 'Insufficient patterns for training' };
    }

    const features = patterns.map(p => this.extractFeatures(p));
    const labels = patterns.map(p => [
      p.success ? 1 : 0,
      (p.duration || 1000) / 10000,
      (p.resourceUsage || 0.5)
    ]);

    const xTrain = tf.tensor2d(features);
    const yTrain = tf.tensor2d(labels);

    if (!this.model) {
      this.model = this.createModel();
    }

    const history = await this.model.fit(xTrain, yTrain, {
      epochs: 30,
      batchSize: 16,
      validationSplit: 0.2,
      verbose: 0
    });

    xTrain.dispose();
    yTrain.dispose();

    await this.redis.set('ml:pattern_trainer:model', JSON.stringify({
      trained_at: Date.now(),
      patterns_count: patterns.length,
      final_loss: history.history.loss[history.history.loss.length - 1]
    }));

    return { success: true, patternsCount: patterns.length };
  }

  extractFeatures(pattern) {
    return [
      pattern.agentType === 'coder' ? 1 : 0,
      pattern.agentType === 'researcher' ? 1 : 0,
      pattern.taskType === 'code-generation' ? 1 : 0,
      pattern.complexity === 'high' ? 1 : 0,
      pattern.complexity === 'medium' ? 1 : 0,
      pattern.complexity === 'low' ? 1 : 0,
      pattern.teamSize || 1,
      pattern.parallelism ? 1 : 0,
      pattern.hasCoordination ? 1 : 0,
      pattern.usesCaching ? 1 : 0,
      pattern.hour / 24,
      pattern.dayOfWeek / 7,
      pattern.agentCount || 1,
      pattern.taskDuration / 10000,
      pattern.cpuUsage || 0.5,
      pattern.memoryUsage || 0.5,
      pattern.networkLatency || 0.5,
      pattern.errorRate || 0,
      pattern.retryCount || 0,
      pattern.priority || 0.5
    ];
  }

  async predictPatternSuccess(pattern) {
    if (!this.model) {
      return { success: 0.5, duration: 1000, confidence: 0 };
    }

    const features = tf.tensor2d([this.extractFeatures(pattern)]);
    const prediction = this.model.predict(features);
    const data = await prediction.data();

    features.dispose();
    prediction.dispose();

    return {
      success: data[0],
      duration: data[1] * 10000,
      resources: data[2],
      confidence: 0.8
    };
  }

  async recommendPatterns(taskContext) {
    const allPatterns = await this.collectPatterns();
    const predictions = await Promise.all(
      allPatterns.map(async p => ({
        pattern: p,
        prediction: await this.predictPatternSuccess(p)
      }))
    );

    return predictions
      .filter(p => p.prediction.success > this.minConfidence)
      .sort((a, b) => b.prediction.success - a.prediction.success)
      .slice(0, 5)
      .map(p => p.pattern);
  }

  async close() {
    if (this.model) {
      this.model.dispose();
    }
    if (this.redis) {
      await this.redis.quit();
    }
  }
}

if (require.main === module) {
  const trainer = new NeuralPatternTrainer();
  if (process.argv[2] === 'train') {
    trainer.collectPatterns().then(patterns => trainer.trainOnPatterns(patterns)).then(() => trainer.close());
  } else if (process.argv[2] === 'predict') {
    const testPattern = { agentType: 'coder', taskType: 'code-generation', complexity: 'medium', teamSize: 3 };
    trainer.predictPatternSuccess(testPattern).then(result => {
      console.log(JSON.stringify(result, null, 2));
      trainer.close();
    });
  }
}

module.exports = NeuralPatternTrainer;
