const EventEmitter = require('events');

class HealthMonitor extends EventEmitter {
  constructor(nodeRegistry, config = {}) {
    super();
    this.nodeRegistry = nodeRegistry;
    this.config = {
      checkInterval: config.checkInterval || 30000, // 30 seconds
      healthTimeout: config.healthTimeout || 5000,   // 5 seconds
      maxRetries: config.maxRetries || 3,
      ...config
    };
    
    this.healthHistory = new Map(); // nodeId -> health history
    this.systemMetrics = {
      startTime: Date.now(),
      totalRequests: 0,
      successfulRequests: 0,
      failedRequests: 0,
      averageLatency: 0,
      lastUpdate: Date.now()
    };
    
    this.isRunning = false;
    this.healthCheckInterval = null;
  }

  start() {
    if (this.isRunning) return;
    
    this.isRunning = true;
    this.healthCheckInterval = setInterval(() => {
      this.performHealthChecks();
    }, this.config.checkInterval);
    
    console.log('Health monitor started');
    this.emit('started');
  }

  stop() {
    if (!this.isRunning) return;
    
    this.isRunning = false;
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
      this.healthCheckInterval = null;
    }
    
    console.log('Health monitor stopped');
    this.emit('stopped');
  }

  async performHealthChecks() {
    const nodes = this.nodeRegistry.getNodes();
    const healthPromises = nodes.map(node => this.checkNodeHealth(node));
    
    try {
      const results = await Promise.allSettled(healthPromises);
      this.processHealthResults(nodes, results);
    } catch (error) {
      console.error('Health check batch failed:', error);
    }
  }

  async checkNodeHealth(node) {
    const startTime = Date.now();
    const healthData = {
      nodeId: node.id,
      nodeName: node.name,
      timestamp: startTime,
      status: 'unknown',
      latency: 0,
      error: null,
      metrics: {}
    };

    try {
      if (node.mock) {
        // Mock node - simulate health check
        healthData.status = 'healthy';
        healthData.latency = Math.floor(Math.random() * 50) + 10;
        healthData.metrics = this.generateMockMetrics(node);
      } else {
        // Real node - perform actual health check
        const response = await this.performRealHealthCheck(node);
        healthData.status = response.status;
        healthData.latency = Date.now() - startTime;
        healthData.metrics = response.metrics || {};
      }
    } catch (error) {
      healthData.status = 'unhealthy';
      healthData.latency = Date.now() - startTime;
      healthData.error = error.message;
    }

    // Update node health in registry
    this.updateNodeHealth(node, healthData);
    
    // Store in history
    this.storeHealthHistory(node.id, healthData);
    
    return healthData;
  }

  generateMockMetrics(node) {
    const baseLoad = node.health?.load || 0;
    const baseCpu = node.health?.cpu || 0;
    const baseMemory = node.health?.memory || 0;
    
    return {
      cpu: Math.max(0, Math.min(100, baseCpu + (Math.random() - 0.5) * 10)),
      memory: Math.max(0, Math.min(100, baseMemory + (Math.random() - 0.5) * 5)),
      load: Math.max(0, Math.min(100, baseLoad + (Math.random() - 0.5) * 15)),
      disk: Math.random() * 80,
      network: {
        bytesIn: Math.floor(Math.random() * 1000000),
        bytesOut: Math.floor(Math.random() * 1000000),
        packetsIn: Math.floor(Math.random() * 10000),
        packetsOut: Math.floor(Math.random() * 10000)
      },
      uptime: Date.now() - (node.health?.uptime || Date.now()),
      requestsPerSecond: Math.floor(Math.random() * 100),
      errorRate: Math.random() * 0.05, // 0-5% error rate
      responseTime: Math.floor(Math.random() * 500) + 50
    };
  }

  async performRealHealthCheck(node) {
    const fetch = require('node-fetch');
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.config.healthTimeout);
    
    try {
      const response = await fetch(`${node.url}/api/tags`, {
        method: 'GET',
        signal: controller.signal,
        headers: { 'Accept': 'application/json' }
      });
      
      clearTimeout(timeout);
      
      if (response.ok) {
        const data = await response.json();
        return {
          status: 'healthy',
          metrics: {
            modelsLoaded: data.models?.length || 0,
            responseTime: response.headers.get('x-response-time') || 0
          }
        };
      } else {
        return {
          status: 'unhealthy',
          error: `HTTP ${response.status}`
        };
      }
    } catch (error) {
      clearTimeout(timeout);
      throw error;
    }
  }

  updateNodeHealth(node, healthData) {
    // Update node health metrics
    node.health = {
      ...node.health,
      ...healthData.metrics,
      lastCheck: healthData.timestamp,
      latency: healthData.latency,
      status: healthData.status
    };
    
    // Update node status
    node.status = healthData.status;
    
    // Emit health change event
    this.emit('healthUpdate', {
      nodeId: node.id,
      nodeName: node.name,
      status: healthData.status,
      metrics: healthData.metrics,
      latency: healthData.latency
    });
  }

  storeHealthHistory(nodeId, healthData) {
    if (!this.healthHistory.has(nodeId)) {
      this.healthHistory.set(nodeId, []);
    }

    const history = this.healthHistory.get(nodeId);
    history.push(healthData);

    // Keep only last 100 entries
    if (history.length > 100) {
      history.shift();
    }
  }

  processHealthResults(nodes, results) {
    let healthyCount = 0;
    let unhealthyCount = 0;
    let totalLatency = 0;
    let validLatencyCount = 0;

    results.forEach((result, index) => {
      if (result.status === 'fulfilled') {
        const healthData = result.value;
        if (healthData.status === 'healthy') {
          healthyCount++;
        } else {
          unhealthyCount++;
        }

        if (healthData.latency > 0) {
          totalLatency += healthData.latency;
          validLatencyCount++;
        }
      } else {
        unhealthyCount++;
        console.error(`Health check failed for node ${nodes[index]?.name}:`, result.reason);
      }
    });

    // Update system metrics
    this.systemMetrics.lastUpdate = Date.now();
    if (validLatencyCount > 0) {
      this.systemMetrics.averageLatency = totalLatency / validLatencyCount;
    }

    // Emit cluster health summary
    this.emit('clusterHealth', {
      totalNodes: nodes.length,
      healthyNodes: healthyCount,
      unhealthyNodes: unhealthyCount,
      averageLatency: this.systemMetrics.averageLatency,
      timestamp: Date.now()
    });
  }

  getNodeHealthHistory(nodeId, limit = 50) {
    const history = this.healthHistory.get(nodeId) || [];
    return history.slice(-limit);
  }

  getClusterHealthSummary() {
    const nodes = this.nodeRegistry.getAllNodes();
    const summary = {
      totalNodes: nodes.length,
      healthyNodes: 0,
      unhealthyNodes: 0,
      unknownNodes: 0,
      averageLatency: 0,
      totalLatency: 0,
      nodeDetails: []
    };

    let totalLatency = 0;
    let latencyCount = 0;

    nodes.forEach(node => {
      const nodeDetail = {
        id: node.id,
        name: node.name,
        status: node.status,
        lastCheck: node.lastCheck,
        latency: node.latency || 0,
        cpu: node.cpu || 0,
        memory: node.memory || 0,
        load: node.load || 0
      };

      switch (node.status) {
        case 'healthy':
          summary.healthyNodes++;
          break;
        case 'unhealthy':
        case 'error':
          summary.unhealthyNodes++;
          break;
        default:
          summary.unknownNodes++;
      }

      if (node.latency && node.latency > 0) {
        totalLatency += node.latency;
        latencyCount++;
      }

      summary.nodeDetails.push(nodeDetail);
    });

    if (latencyCount > 0) {
      summary.averageLatency = totalLatency / latencyCount;
    }

    return summary;
  }

  getSystemMetrics() {
    return {
      ...this.systemMetrics,
      uptime: Date.now() - this.systemMetrics.startTime,
      healthCheckInterval: this.config.checkInterval,
      isRunning: this.isRunning
    };
  }

  recordRequest(success = true, latency = 0) {
    this.systemMetrics.totalRequests++;
    if (success) {
      this.systemMetrics.successfulRequests++;
    } else {
      this.systemMetrics.failedRequests++;
    }

    // Update average latency (simple moving average)
    if (latency > 0) {
      const alpha = 0.1; // Smoothing factor
      this.systemMetrics.averageLatency =
        (alpha * latency) + ((1 - alpha) * this.systemMetrics.averageLatency);
    }
  }

  getPrometheusMetrics() {
    const nodes = this.nodeRegistry.getAllNodes();
    const metrics = [];

    // System-level metrics
    metrics.push(`# HELP ollama_nodes_total Total number of nodes`);
    metrics.push(`# TYPE ollama_nodes_total gauge`);
    metrics.push(`ollama_nodes_total ${nodes.length}`);

    metrics.push(`# HELP ollama_nodes_healthy Number of healthy nodes`);
    metrics.push(`# TYPE ollama_nodes_healthy gauge`);
    const healthyNodes = nodes.filter(n => n.status === 'healthy').length;
    metrics.push(`ollama_nodes_healthy ${healthyNodes}`);

    metrics.push(`# HELP ollama_requests_total Total number of requests`);
    metrics.push(`# TYPE ollama_requests_total counter`);
    metrics.push(`ollama_requests_total ${this.systemMetrics.totalRequests}`);

    metrics.push(`# HELP ollama_requests_successful_total Total number of successful requests`);
    metrics.push(`# TYPE ollama_requests_successful_total counter`);
    metrics.push(`ollama_requests_successful_total ${this.systemMetrics.successfulRequests}`);

    metrics.push(`# HELP ollama_average_latency_seconds Average request latency in seconds`);
    metrics.push(`# TYPE ollama_average_latency_seconds gauge`);
    metrics.push(`ollama_average_latency_seconds ${this.systemMetrics.averageLatency / 1000}`);

    // Node-level metrics
    nodes.forEach(node => {
      const labels = `{node_id="${node.id}",node_name="${node.name}"}`;

      metrics.push(`# HELP ollama_node_cpu_usage CPU usage percentage`);
      metrics.push(`# TYPE ollama_node_cpu_usage gauge`);
      metrics.push(`ollama_node_cpu_usage${labels} ${node.cpu || 0}`);

      metrics.push(`# HELP ollama_node_memory_usage Memory usage percentage`);
      metrics.push(`# TYPE ollama_node_memory_usage gauge`);
      metrics.push(`ollama_node_memory_usage${labels} ${node.memory || 0}`);

      metrics.push(`# HELP ollama_node_load Load percentage`);
      metrics.push(`# TYPE ollama_node_load gauge`);
      metrics.push(`ollama_node_load${labels} ${node.load || 0}`);

      metrics.push(`# HELP ollama_node_latency_seconds Node response latency in seconds`);
      metrics.push(`# TYPE ollama_node_latency_seconds gauge`);
      metrics.push(`ollama_node_latency_seconds${labels} ${(node.latency || 0) / 1000}`);

      const statusValue = node.status === 'healthy' ? 1 : 0;
      metrics.push(`# HELP ollama_node_healthy Node health status (1=healthy, 0=unhealthy)`);
      metrics.push(`# TYPE ollama_node_healthy gauge`);
      metrics.push(`ollama_node_healthy${labels} ${statusValue}`);
    });

    return metrics.join('\n');
  }
}

module.exports = HealthMonitor;
