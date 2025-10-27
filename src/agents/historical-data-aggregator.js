/**
 * Historical Performance Data Aggregator
 * Aggregates data from Redis, Claude-Flow memory, and swarm metrics for ML training
 */

const Redis = require('ioredis');
const fs = require('fs').promises;
const path = require('path');

class HistoricalDataAggregator {
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

    this.dataSources = {
      redis: ['agent:*', 'task:*', 'swarm:*', 'ml:*'],
      files: [
        '.claude-flow/memory/neural-memory.json',
        '.claude-flow/memory/performance-patterns.json',
        '.claude-flow/memory/adaptation-rules.json',
        '.claude-flow/metrics/enhanced-swarm-metrics.json',
        '.claude-flow/metrics/agent-metrics.json',
        '.claude-flow/metrics/task-metrics.json'
      ]
    };
  }

  async aggregateHistoricalData(startTime, endTime) {
    console.log(`[Data Aggregator] Aggregating data from ${new Date(startTime)} to ${new Date(endTime)}`);

    const redisData = await this.collectRedisData(startTime, endTime);
    const fileData = await this.collectFileData();
    const mergedData = this.mergeData(redisData, fileData);

    await this.redis.set(
      'ml:aggregated_training_data',
      JSON.stringify({ data: mergedData, timestamp: Date.now() })
    );

    return mergedData;
  }

  async collectRedisData(startTime, endTime) {
    const data = { agents: [], tasks: [], swarms: [], ml: [] };

    for (const pattern of this.dataSources.redis) {
      const keys = await this.redis.keys(pattern);
      for (const key of keys.slice(0, 1000)) { // Limit to prevent overload
        const type = key.split(':')[0];
        const value = await this.redis.get(key);

        if (value) {
          try {
            const parsed = JSON.parse(value);
            if (parsed.timestamp >= startTime && parsed.timestamp <= endTime) {
              data[type === 'agent' ? 'agents' : type === 'task' ? 'tasks' : type === 'swarm' ? 'swarms' : 'ml'].push(parsed);
            }
          } catch (e) {
            // Skip non-JSON values
          }
        }
      }
    }

    return data;
  }

  async collectFileData() {
    const data = {};

    for (const filePath of this.dataSources.files) {
      try {
        const fullPath = path.join(process.cwd(), filePath);
        const content = await fs.readFile(fullPath, 'utf-8');
        const key = path.basename(filePath, '.json');
        data[key] = JSON.parse(content);
      } catch (error) {
        console.warn(`[Data Aggregator] Could not read ${filePath}:`, error.message);
      }
    }

    return data;
  }

  mergeData(redisData, fileData) {
    const merged = [];

    // Merge agent data
    for (const agent of redisData.agents) {
      merged.push({
        type: 'agent',
        id: agent.agent_id,
        timestamp: agent.timestamp,
        metrics: agent,
        source: 'redis'
      });
    }

    // Add file-based neural memory
    if (fileData['neural-memory']) {
      for (const memory of fileData['neural-memory']) {
        merged.push({
          type: 'neural_memory',
          timestamp: memory.timestamp || Date.now(),
          data: memory,
          source: 'file'
        });
      }
    }

    // Sort by timestamp
    return merged.sort((a, b) => a.timestamp - b.timestamp);
  }

  async getTrainingDataset(modelType, limit = 1000) {
    const aggregatedKey = 'ml:aggregated_training_data';
    const data = await this.redis.get(aggregatedKey);

    if (!data) {
      return { features: [], labels: [] };
    }

    const parsed = JSON.parse(data);
    const filtered = parsed.data
      .filter(d => this.isRelevantForModel(d, modelType))
      .slice(0, limit);

    return {
      features: filtered.map(d => this.extractFeatures(d, modelType)),
      labels: filtered.map(d => this.extractLabel(d, modelType))
    };
  }

  isRelevantForModel(dataPoint, modelType) {
    if (modelType === 'agent_selection') {
      return dataPoint.type === 'agent' || dataPoint.type === 'task';
    } else if (modelType === 'scaling') {
      return dataPoint.type === 'swarm' || dataPoint.type === 'agent';
    }
    return true;
  }

  extractFeatures(dataPoint, modelType) {
    if (modelType === 'agent_selection') {
      return [
        dataPoint.metrics?.complexity || 0.5,
        dataPoint.metrics?.agent_load || 0,
        dataPoint.metrics?.success_rate || 0.5,
        dataPoint.metrics?.avg_duration || 1000
      ];
    }
    return [0, 0, 0, 0];
  }

  extractLabel(dataPoint, modelType) {
    if (modelType === 'agent_selection') {
      return dataPoint.metrics?.success ? 1 : 0;
    }
    return 0;
  }

  async syncClaudeFlowMemoryToRedis() {
    console.log('[Data Aggregator] Syncing Claude-Flow memory to Redis...');

    const fileData = await this.collectFileData();

    for (const [key, value] of Object.entries(fileData)) {
      await this.redis.setex(
        `claude_flow:memory:${key}`,
        86400, // 24 hour expiry
        JSON.stringify(value)
      );
    }

    console.log('[Data Aggregator] Sync completed');
  }

  async getDataQualityReport() {
    const aggregatedKey = 'ml:aggregated_training_data';
    const data = await this.redis.get(aggregatedKey);

    if (!data) {
      return { quality: 'no_data', issues: [] };
    }

    const parsed = JSON.parse(data);
    const issues = [];

    // Check for missing timestamps
    const missingTimestamps = parsed.data.filter(d => !d.timestamp).length;
    if (missingTimestamps > 0) {
      issues.push(`${missingTimestamps} records missing timestamps`);
    }

    // Check data freshness
    const latestTimestamp = Math.max(...parsed.data.map(d => d.timestamp || 0));
    const ageHours = (Date.now() - latestTimestamp) / 3600000;
    if (ageHours > 24) {
      issues.push(`Data is ${ageHours.toFixed(1)} hours old`);
    }

    return {
      quality: issues.length === 0 ? 'good' : 'needs_attention',
      totalRecords: parsed.data.length,
      latestTimestamp: new Date(latestTimestamp),
      issues
    };
  }

  async close() {
    if (this.redis) {
      await this.redis.quit();
    }
  }
}

if (require.main === module) {
  const aggregator = new HistoricalDataAggregator();
  const command = process.argv[2];

  if (command === 'aggregate') {
    const startTime = Date.now() - 24 * 3600000; // Last 24 hours
    const endTime = Date.now();
    aggregator.aggregateHistoricalData(startTime, endTime).then(() => aggregator.close());
  } else if (command === 'sync') {
    aggregator.syncClaudeFlowMemoryToRedis().then(() => aggregator.close());
  } else if (command === 'validate') {
    aggregator.getDataQualityReport().then(report => {
      console.log(JSON.stringify(report, null, 2));
      aggregator.close();
    });
  }
}

module.exports = HistoricalDataAggregator;
