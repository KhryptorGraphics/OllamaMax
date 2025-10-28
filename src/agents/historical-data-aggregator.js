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

    // Safety cap on total processed records
    const maxTotalRecords = parseInt(process.env.MAX_AGGREGATED_RECORDS) || 10000;
    let totalProcessed = 0;

    for (const pattern of this.dataSources.redis) {
      // Use SCAN instead of KEYS for non-blocking iteration
      let cursor = '0';
      let keys = [];

      do {
        const [newCursor, scannedKeys] = await this.redis.scan(
          cursor,
          'MATCH', pattern,
          'COUNT', 500
        );
        cursor = newCursor;
        keys.push(...scannedKeys);

        // Safety check to avoid excessive memory
        if (keys.length >= maxTotalRecords || totalProcessed >= maxTotalRecords) {
          console.warn(`[Data Aggregator] Reached max records limit (${maxTotalRecords}), stopping scan`);
          break;
        }
      } while (cursor !== '0');

      for (const key of keys) {
        if (totalProcessed >= maxTotalRecords) break;
        const type = key.split(':')[0];

        // Detect key type and fetch accordingly
        const keyType = await this.redis.type(key);
        let value = null;

        try {
          switch (keyType) {
            case 'string':
              value = await this.redis.get(key);
              if (value) {
                const parsed = JSON.parse(value);
                if (parsed.timestamp >= startTime && parsed.timestamp <= endTime) {
                  data[type === 'agent' ? 'agents' : type === 'task' ? 'tasks' : type === 'swarm' ? 'swarms' : 'ml'].push({
                    ...parsed,
                    _key: key,
                    _type: 'string'
                  });
                }
              }
              break;

            case 'hash':
              const hashData = await this.redis.hgetall(key);
              if (hashData && Object.keys(hashData).length > 0) {
                // Normalize hash to standard format
                const normalized = {
                  ...hashData,
                  timestamp: parseInt(hashData.timestamp) || Date.now(),
                  _key: key,
                  _type: 'hash'
                };
                if (normalized.timestamp >= startTime && normalized.timestamp <= endTime) {
                  data[type === 'agent' ? 'agents' : type === 'task' ? 'tasks' : type === 'swarm' ? 'swarms' : 'ml'].push(normalized);
                }
              }
              break;

            case 'list':
              const listData = await this.redis.lrange(key, 0, 99); // Last 100 entries
              for (const item of listData) {
                try {
                  const parsed = JSON.parse(item);
                  if (parsed.timestamp >= startTime && parsed.timestamp <= endTime) {
                    data[type === 'agent' ? 'agents' : type === 'task' ? 'tasks' : type === 'swarm' ? 'swarms' : 'ml'].push({
                      ...parsed,
                      _key: key,
                      _type: 'list'
                    });
                  }
                } catch (e) {
                  // Skip invalid JSON
                }
              }
              break;

            case 'zset':
              const zsetData = await this.redis.zrange(key, 0, 99, 'WITHSCORES');
              for (let i = 0; i < zsetData.length; i += 2) {
                try {
                  const member = JSON.parse(zsetData[i]);
                  const score = parseFloat(zsetData[i + 1]);
                  if (score >= startTime && score <= endTime) {
                    data[type === 'agent' ? 'agents' : type === 'task' ? 'tasks' : type === 'swarm' ? 'swarms' : 'ml'].push({
                      ...member,
                      timestamp: score,
                      _key: key,
                      _type: 'zset'
                    });
                  }
                } catch (e) {
                  // Skip invalid JSON
                }
              }
              break;

            default:
              // Skip unsupported types (set, etc.)
              break;
          }
        } catch (error) {
          console.warn(`[Data Aggregator] Error processing key ${key}:`, error.message);
        }

        totalProcessed++;
      }
    }

    console.log(`[Data Aggregator] Processed ${totalProcessed} total records`);
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

  /**
   * Map categorical fields to numeric encodings
   */
  encodeCategorical(value, fieldName) {
    const encodings = {
      complexity: { 'low': 0, 'medium': 1, 'high': 2 },
      task_type: { 'general': 0, 'code': 1, 'research': 2, 'testing': 3, 'deployment': 4 },
      specialization: { 'general': 0, 'coder': 1, 'researcher': 2, 'tester': 3, 'devops': 4 },
      agent_type: { 'general-purpose': 0, 'coder': 1, 'researcher': 2, 'tester': 3 }
    };

    if (encodings[fieldName] && encodings[fieldName][value] !== undefined) {
      return encodings[fieldName][value];
    }

    // For unknown categorical values, use one-hot encoding (simplified)
    return typeof value === 'string' ? value.length % 10 : value;
  }

  /**
   * Normalize numeric values to 0-1 range
   */
  normalize(value, field) {
    const ranges = {
      duration: [0, 10000],
      load: [0, 10],
      success_rate: [0, 1],
      execution_time: [0, 10000],
      cpu_usage: [0, 1],
      memory_usage: [0, 1]
    };

    if (ranges[field]) {
      const [min, max] = ranges[field];
      return Math.max(0, Math.min(1, (value - min) / (max - min)));
    }

    return value;
  }

  extractFeatures(dataPoint, modelType) {
    if (modelType === 'agent_selection') {
      const metrics = dataPoint.metrics || {};

      // Handle categorical fields with encoding
      const complexity = this.encodeCategorical(metrics.complexity || 'medium', 'complexity') / 2;
      const taskType = this.encodeCategorical(metrics.task_type || 'general', 'task_type') / 4;
      const specialization = this.encodeCategorical(metrics.specialization || 'general', 'specialization') / 4;

      // Normalize numeric fields
      const agentLoad = this.normalize(metrics.agent_load || 0, 'load');
      const successRate = this.normalize(metrics.success_rate || 0.5, 'success_rate');
      const avgDuration = this.normalize(metrics.avg_duration || 1000, 'duration');

      return [complexity, taskType, specialization, agentLoad, successRate, avgDuration];
    }
    return [0, 0, 0, 0, 0, 0];
  }

  extractLabel(dataPoint, modelType) {
    if (modelType === 'agent_selection') {
      // Binary label for success
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
