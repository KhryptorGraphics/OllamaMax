/**
 * Swarm Coordination Algorithm Optimizer
 * ML-powered optimization of swarm coordination algorithms
 */

const Redis = require('ioredis');
const RandomForest = require('ml-random-forest').RandomForestClassifier;

class SwarmAlgorithmOptimizer {
  constructor() {
    this.redis = new Redis.Cluster([
      { host: 'localhost', port: 7000 },
      { host: 'localhost', port: 7001 },
      { host: 'localhost', port: 7002 }
    ], {
      redisOptions: { password: process.env.REDIS_PASSWORD, connectTimeout: 10000 }
    });

    this.algorithms = ['queen', 'mesh', 'adaptive', 'ring', 'star'];
    this.classifier = null;
    this.performanceHistory = new Map();
  }

  async trackAlgorithmPerformance(algorithm, taskType, metrics) {
    const key = `${algorithm}:${taskType}`;

    if (!this.performanceHistory.has(key)) {
      this.performanceHistory.set(key, []);
    }

    this.performanceHistory.get(key).push({
      algorithm,
      taskType,
      efficiency: metrics.efficiency || 0,
      latency: metrics.latency || 0,
      successRate: metrics.successRate || 0,
      resourceUtil: metrics.resourceUtil || 0,
      timestamp: Date.now()
    });

    // Store in Redis
    await this.redis.lpush(`swarm:algorithm:performance:${algorithm}`, JSON.stringify({
      taskType,
      metrics,
      timestamp: Date.now()
    }));
    await this.redis.ltrim(`swarm:algorithm:performance:${algorithm}`, 0, 999);
  }

  async trainAlgorithmClassifier() {
    console.log('[Swarm Algorithm Optimizer] Training algorithm classifier...');

    const trainingData = [];
    const labels = [];

    for (const [key, history] of this.performanceHistory.entries()) {
      const [algorithm, taskType] = key.split(':');

      for (const entry of history) {
        trainingData.push([
          this.taskTypeToFeature(taskType),
          entry.efficiency,
          entry.latency / 1000, // Normalize
          entry.successRate,
          entry.resourceUtil
        ]);
        labels.push(this.algorithms.indexOf(algorithm));
      }
    }

    if (trainingData.length < 10) {
      return { success: false, error: 'Insufficient training data' };
    }

    this.classifier = new RandomForest({ nEstimators: 50, maxDepth: 10 });
    this.classifier.train(trainingData, labels);

    await this.redis.set('ml:swarm_algorithm_classifier', JSON.stringify({
      trained_at: Date.now(),
      samples: trainingData.length
    }));

    return { success: true, samples: trainingData.length };
  }

  taskTypeToFeature(taskType) {
    const types = ['code', 'research', 'test', 'review', 'deploy'];
    return types.indexOf(taskType) >= 0 ? types.indexOf(taskType) / types.length : 0.5;
  }

  async predictOptimalAlgorithm(taskContext) {
    if (!this.classifier) {
      return { algorithm: 'queen', confidence: 0, reason: 'No classifier available' };
    }

    const features = [
      this.taskTypeToFeature(taskContext.type),
      taskContext.complexity || 0.5,
      taskContext.expectedLatency || 0.5,
      taskContext.priorityLevel || 0.5,
      taskContext.resourceBudget || 0.5
    ];

    const prediction = this.classifier.predict([features])[0];
    const algorithm = this.algorithms[prediction] || 'queen';

    return {
      algorithm,
      confidence: 0.75,
      alternativeAlgorithms: this.algorithms.filter(a => a !== algorithm).slice(0, 2)
    };
  }

  async compareAlgorithms(taskType) {
    const comparison = {};

    for (const algorithm of this.algorithms) {
      const key = `swarm:algorithm:performance:${algorithm}`;
      const data = await this.redis.lrange(key, 0, 99);

      if (data.length === 0) {
        comparison[algorithm] = { avgEfficiency: 0, avgLatency: 0, sampleCount: 0 };
        continue;
      }

      const parsed = data.map(d => JSON.parse(d));
      const metrics = parsed.filter(p => p.taskType === taskType);

      if (metrics.length > 0) {
        comparison[algorithm] = {
          avgEfficiency: metrics.reduce((sum, m) => sum + (m.metrics.efficiency || 0), 0) / metrics.length,
          avgLatency: metrics.reduce((sum, m) => sum + (m.metrics.latency || 0), 0) / metrics.length,
          sampleCount: metrics.length
        };
      } else {
        comparison[algorithm] = { avgEfficiency: 0, avgLatency: 0, sampleCount: 0 };
      }
    }

    return comparison;
  }

  async close() {
    if (this.redis) {
      await this.redis.quit();
    }
  }
}

// CLI support
if (require.main === module) {
  const optimizer = new SwarmAlgorithmOptimizer();
  const command = process.argv[2];

  if (command === 'train') {
    optimizer.trainAlgorithmClassifier().then(() => optimizer.close());
  } else if (command === 'predict') {
    const taskContext = { type: 'code', complexity: 0.7, priorityLevel: 0.8 };
    optimizer.predictOptimalAlgorithm(taskContext).then(result => {
      console.log(JSON.stringify(result, null, 2));
      optimizer.close();
    });
  }
}

module.exports = SwarmAlgorithmOptimizer;
