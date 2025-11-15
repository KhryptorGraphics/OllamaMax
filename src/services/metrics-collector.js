const EventEmitter = require('events');

class MetricsCollector extends EventEmitter {
  constructor(config = {}) {
    super();
    this.config = {
      collectionInterval: config.collectionInterval || 15000, // 15 seconds
      retentionPeriod: config.retentionPeriod || 24 * 60 * 60 * 1000, // 24 hours
      maxDataPoints: config.maxDataPoints || 1000,
      ...config
    };

    this.metrics = {
      // Request metrics
      requests: {
        total: 0,
        successful: 0,
        failed: 0,
        byModel: new Map(),
        byNode: new Map(),
        latencyHistogram: []
      },
      
      // Node metrics
      nodes: {
        total: 0,
        healthy: 0,
        unhealthy: 0,
        byStatus: new Map(),
        resourceUsage: new Map()
      },
      
      // Model metrics
      models: {
        total: 0,
        loaded: 0,
        replicas: 0,
        byModel: new Map(),
        transferMetrics: new Map()
      },
      
      // System metrics
      system: {
        uptime: Date.now(),
        memoryUsage: process.memoryUsage(),
        cpuUsage: process.cpuUsage(),
        eventLoopDelay: 0
      },
      
      // Time series data
      timeSeries: {
        requests: [],
        latency: [],
        nodeHealth: [],
        resourceUsage: []
      }
    };

    this.isCollecting = false;
    this.collectionInterval = null;
  }

  start() {
    if (this.isCollecting) return;
    
    this.isCollecting = true;
    this.collectionInterval = setInterval(() => {
      this.collectMetrics();
    }, this.config.collectionInterval);
    
    console.log('Metrics collector started');
    this.emit('started');
  }

  stop() {
    if (!this.isCollecting) return;
    
    this.isCollecting = false;
    if (this.collectionInterval) {
      clearInterval(this.collectionInterval);
      this.collectionInterval = null;
    }
    
    console.log('Metrics collector stopped');
    this.emit('stopped');
  }

  collectMetrics() {
    const timestamp = Date.now();
    
    // Collect system metrics
    this.collectSystemMetrics(timestamp);
    
    // Clean up old data
    this.cleanupOldData(timestamp);
    
    this.emit('metricsCollected', {
      timestamp,
      metrics: this.getSnapshot()
    });
  }

  collectSystemMetrics(timestamp) {
    // Update system metrics
    this.metrics.system.memoryUsage = process.memoryUsage();
    this.metrics.system.cpuUsage = process.cpuUsage();
    
    // Measure event loop delay
    const start = process.hrtime.bigint();
    setImmediate(() => {
      const delay = Number(process.hrtime.bigint() - start) / 1000000; // Convert to ms
      this.metrics.system.eventLoopDelay = delay;
    });
  }

  recordRequest(data) {
    const { modelName, nodeId, success, latency, error } = data;
    
    // Update request counters
    this.metrics.requests.total++;
    if (success) {
      this.metrics.requests.successful++;
    } else {
      this.metrics.requests.failed++;
    }
    
    // Update by model
    if (modelName) {
      if (!this.metrics.requests.byModel.has(modelName)) {
        this.metrics.requests.byModel.set(modelName, { total: 0, successful: 0, failed: 0 });
      }
      const modelStats = this.metrics.requests.byModel.get(modelName);
      modelStats.total++;
      if (success) {
        modelStats.successful++;
      } else {
        modelStats.failed++;
      }
    }
    
    // Update by node
    if (nodeId) {
      if (!this.metrics.requests.byNode.has(nodeId)) {
        this.metrics.requests.byNode.set(nodeId, { total: 0, successful: 0, failed: 0 });
      }
      const nodeStats = this.metrics.requests.byNode.get(nodeId);
      nodeStats.total++;
      if (success) {
        nodeStats.successful++;
      } else {
        nodeStats.failed++;
      }
    }
    
    // Record latency
    if (latency && latency > 0) {
      this.metrics.requests.latencyHistogram.push({
        timestamp: Date.now(),
        latency,
        modelName,
        nodeId
      });
      
      // Keep histogram size manageable
      if (this.metrics.requests.latencyHistogram.length > this.config.maxDataPoints) {
        this.metrics.requests.latencyHistogram.shift();
      }
    }
    
    // Add to time series
    this.metrics.timeSeries.requests.push({
      timestamp: Date.now(),
      total: this.metrics.requests.total,
      successful: this.metrics.requests.successful,
      failed: this.metrics.requests.failed
    });
    
    this.emit('requestRecorded', data);
  }

  recordNodeMetrics(nodeData) {
    const { nodeId, status, cpu, memory, load, latency } = nodeData;
    
    // Update node resource usage
    this.metrics.nodes.resourceUsage.set(nodeId, {
      cpu: cpu || 0,
      memory: memory || 0,
      load: load || 0,
      latency: latency || 0,
      timestamp: Date.now()
    });
    
    // Add to time series
    this.metrics.timeSeries.resourceUsage.push({
      timestamp: Date.now(),
      nodeId,
      cpu: cpu || 0,
      memory: memory || 0,
      load: load || 0,
      latency: latency || 0
    });
    
    this.emit('nodeMetricsRecorded', nodeData);
  }

  recordModelMetrics(modelData) {
    const { modelName, operation, nodeId, success, transferSize, duration } = modelData;
    
    // Update model metrics
    if (!this.metrics.models.byModel.has(modelName)) {
      this.metrics.models.byModel.set(modelName, {
        requests: 0,
        transfers: 0,
        replicas: 0,
        totalSize: 0
      });
    }
    
    const modelStats = this.metrics.models.byModel.get(modelName);
    
    switch (operation) {
      case 'inference':
        modelStats.requests++;
        break;
      case 'transfer':
        modelStats.transfers++;
        if (transferSize) {
          modelStats.totalSize += transferSize;
        }
        break;
      case 'replicate':
        if (success) {
          modelStats.replicas++;
        }
        break;
    }
    
    this.emit('modelMetricsRecorded', modelData);
  }

  cleanupOldData(currentTimestamp) {
    const cutoffTime = currentTimestamp - this.config.retentionPeriod;

    // Clean up time series data
    Object.keys(this.metrics.timeSeries).forEach(key => {
      this.metrics.timeSeries[key] = this.metrics.timeSeries[key].filter(
        item => item.timestamp > cutoffTime
      );
    });

    // Clean up latency histogram
    this.metrics.requests.latencyHistogram = this.metrics.requests.latencyHistogram.filter(
      item => item.timestamp > cutoffTime
    );
  }

  getSnapshot() {
    return {
      timestamp: Date.now(),
      requests: {
        total: this.metrics.requests.total,
        successful: this.metrics.requests.successful,
        failed: this.metrics.requests.failed,
        successRate: this.metrics.requests.total > 0
          ? this.metrics.requests.successful / this.metrics.requests.total
          : 0,
        byModel: Object.fromEntries(this.metrics.requests.byModel),
        byNode: Object.fromEntries(this.metrics.requests.byNode)
      },
      nodes: {
        total: this.metrics.nodes.total,
        healthy: this.metrics.nodes.healthy,
        unhealthy: this.metrics.nodes.unhealthy,
        resourceUsage: Object.fromEntries(this.metrics.nodes.resourceUsage)
      },
      models: {
        total: this.metrics.models.total,
        loaded: this.metrics.models.loaded,
        replicas: this.metrics.models.replicas,
        byModel: Object.fromEntries(this.metrics.models.byModel)
      },
      system: {
        ...this.metrics.system,
        uptime: Date.now() - this.metrics.system.uptime
      }
    };
  }

  getLatencyStats() {
    const latencies = this.metrics.requests.latencyHistogram.map(item => item.latency);

    if (latencies.length === 0) {
      return { count: 0, min: 0, max: 0, avg: 0, p50: 0, p95: 0, p99: 0 };
    }

    const sorted = latencies.sort((a, b) => a - b);
    const count = sorted.length;

    return {
      count,
      min: sorted[0],
      max: sorted[count - 1],
      avg: sorted.reduce((sum, val) => sum + val, 0) / count,
      p50: sorted[Math.floor(count * 0.5)],
      p95: sorted[Math.floor(count * 0.95)],
      p99: sorted[Math.floor(count * 0.99)]
    };
  }

  getTimeSeriesData(metric, timeRange = 3600000) { // Default 1 hour
    const cutoffTime = Date.now() - timeRange;

    if (!this.metrics.timeSeries[metric]) {
      return [];
    }

    return this.metrics.timeSeries[metric].filter(
      item => item.timestamp > cutoffTime
    );
  }

  getPrometheusMetrics() {
    const metrics = [];
    const timestamp = Date.now();

    // Request metrics
    metrics.push(`# HELP ollama_requests_total Total number of requests processed`);
    metrics.push(`# TYPE ollama_requests_total counter`);
    metrics.push(`ollama_requests_total ${this.metrics.requests.total}`);

    metrics.push(`# HELP ollama_requests_successful_total Total number of successful requests`);
    metrics.push(`# TYPE ollama_requests_successful_total counter`);
    metrics.push(`ollama_requests_successful_total ${this.metrics.requests.successful}`);

    metrics.push(`# HELP ollama_requests_failed_total Total number of failed requests`);
    metrics.push(`# TYPE ollama_requests_failed_total counter`);
    metrics.push(`ollama_requests_failed_total ${this.metrics.requests.failed}`);

    // Latency metrics
    const latencyStats = this.getLatencyStats();
    metrics.push(`# HELP ollama_request_latency_seconds Request latency statistics`);
    metrics.push(`# TYPE ollama_request_latency_seconds summary`);
    metrics.push(`ollama_request_latency_seconds{quantile="0.5"} ${latencyStats.p50 / 1000}`);
    metrics.push(`ollama_request_latency_seconds{quantile="0.95"} ${latencyStats.p95 / 1000}`);
    metrics.push(`ollama_request_latency_seconds{quantile="0.99"} ${latencyStats.p99 / 1000}`);
    metrics.push(`ollama_request_latency_seconds_sum ${(latencyStats.avg * latencyStats.count) / 1000}`);
    metrics.push(`ollama_request_latency_seconds_count ${latencyStats.count}`);

    // Model metrics
    for (const [modelName, stats] of this.metrics.models.byModel) {
      const labels = `{model="${modelName}"}`;
      metrics.push(`ollama_model_requests_total${labels} ${stats.requests}`);
      metrics.push(`ollama_model_transfers_total${labels} ${stats.transfers}`);
      metrics.push(`ollama_model_replicas${labels} ${stats.replicas}`);
    }

    // Node metrics
    for (const [nodeId, usage] of this.metrics.nodes.resourceUsage) {
      const labels = `{node_id="${nodeId}"}`;
      metrics.push(`ollama_node_cpu_usage${labels} ${usage.cpu}`);
      metrics.push(`ollama_node_memory_usage${labels} ${usage.memory}`);
      metrics.push(`ollama_node_load${labels} ${usage.load}`);
      metrics.push(`ollama_node_latency_seconds${labels} ${usage.latency / 1000}`);
    }

    // System metrics
    const memUsage = this.metrics.system.memoryUsage;
    metrics.push(`# HELP ollama_system_memory_usage_bytes System memory usage`);
    metrics.push(`# TYPE ollama_system_memory_usage_bytes gauge`);
    metrics.push(`ollama_system_memory_usage_bytes{type="rss"} ${memUsage.rss}`);
    metrics.push(`ollama_system_memory_usage_bytes{type="heapTotal"} ${memUsage.heapTotal}`);
    metrics.push(`ollama_system_memory_usage_bytes{type="heapUsed"} ${memUsage.heapUsed}`);

    metrics.push(`# HELP ollama_system_uptime_seconds System uptime in seconds`);
    metrics.push(`# TYPE ollama_system_uptime_seconds gauge`);
    metrics.push(`ollama_system_uptime_seconds ${(timestamp - this.metrics.system.uptime) / 1000}`);

    metrics.push(`# HELP ollama_system_event_loop_delay_seconds Event loop delay in seconds`);
    metrics.push(`# TYPE ollama_system_event_loop_delay_seconds gauge`);
    metrics.push(`ollama_system_event_loop_delay_seconds ${this.metrics.system.eventLoopDelay / 1000}`);

    return metrics.join('\n');
  }

  reset() {
    // Reset counters but keep configuration
    this.metrics.requests = {
      total: 0,
      successful: 0,
      failed: 0,
      byModel: new Map(),
      byNode: new Map(),
      latencyHistogram: []
    };

    this.metrics.timeSeries = {
      requests: [],
      latency: [],
      nodeHealth: [],
      resourceUsage: []
    };

    this.emit('metricsReset');
  }
}

module.exports = MetricsCollector;
