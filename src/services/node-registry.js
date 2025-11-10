/**
 * Node Registry Service
 * Manages distributed Ollama nodes and their health status
 */

const fetch = require('node-fetch');

class LoadBalancer {
  constructor() {
    this.currentIndex = 0;
  }

  select(nodes, strategy) {
    switch(strategy) {
      case 'round-robin':
        return this.roundRobin(nodes);
      case 'least-loaded':
        return this.leastLoaded(nodes);
      case 'fastest':
        return this.fastest(nodes);
      default:
        return this.roundRobin(nodes);
    }
  }

  roundRobin(nodes) {
    const node = nodes[this.currentIndex % nodes.length];
    this.currentIndex++;
    return node;
  }

  leastLoaded(nodes) {
    return nodes.reduce((min, node) => 
      node.health.load < min.health.load ? node : min
    );
  }

  fastest(nodes) {
    return nodes.reduce((min, node) => 
      node.health.requestsPerSecond > min.health.requestsPerSecond ? node : min
    );
  }
}

class NodeRegistry {
  constructor() {
    this.nodes = new Map();
    this.loadBalancer = new LoadBalancer();
    this.healthCheckInterval = null;
    this.startHealthChecks();
  }

  addNode(idOrConfig, config) {
    // Support both signatures: addNode(id, config) and addNode(config)
    let id, nodeConfig;

    if (typeof idOrConfig === 'string') {
      // Old signature: addNode(id, config)
      id = idOrConfig;
      nodeConfig = config;
    } else {
      // New signature: addNode(config) where config includes id
      nodeConfig = idOrConfig;
      id = nodeConfig.id;
    }

    const node = {
      id,
      name: nodeConfig.name,
      url: nodeConfig.url,
      status: nodeConfig.status || (nodeConfig.mock ? 'healthy' : 'connecting'),
      mock: nodeConfig.mock || false,
      version: nodeConfig.version,
      models: nodeConfig.models || ['llama-3.2-1b', 'llama-3.2-3b'],
      capabilities: nodeConfig.capabilities || {
        inference: true,
        embeddings: true,
        streaming: true
      },
      health: {
        load: nodeConfig.metrics?.load || Math.random() * 50,
        memory: nodeConfig.metrics?.memory || Math.random() * 70,
        requestsPerSecond: nodeConfig.metrics?.requestsPerSecond || Math.floor(Math.random() * 100),
        queue: nodeConfig.metrics?.queue || 0,
        lastCheck: Date.now(),
        uptime: Date.now(),
        cpu: Math.random() * 60,
        disk: Math.random() * 40,
        avgResponseTime: Math.floor(Math.random() * 500) + 100,
        modelsLoaded: nodeConfig.models || ['llama-3.2-1b', 'llama-3.2-3b']
      },
      connection: null
    };

    this.nodes.set(id, node);

    if (!nodeConfig.mock) {
      this.connectToNode(node);
    }

    return node;
  }

  connectToNode(node) {
    // For real nodes, attempt to connect
    fetch(`${node.url}/api/tags`)
      .then(response => response.json())
      .then(data => {
        node.status = 'healthy';
        node.health.modelsLoaded = data.models?.map(m => m.name) || [];
        console.log(`Connected to node: ${node.name}`);
      })
      .catch(error => {
        console.error(`Failed to connect to node ${node.name}:`, error.message);
        node.status = 'error';
        // Retry after 5 seconds
        setTimeout(() => this.connectToNode(node), 5000);
      });
  }

  async updateNodeHealth(node) {
    if (node.mock) {
      // Update mock metrics with some variation
      node.health.load = Math.max(0, Math.min(100, node.health.load + (Math.random() - 0.5) * 10));
      node.health.memory = Math.max(0, Math.min(100, node.health.memory + (Math.random() - 0.5) * 5));
      node.health.cpu = Math.max(0, Math.min(100, node.health.cpu + (Math.random() - 0.5) * 15));
      node.health.requestsPerSecond = Math.max(0, node.health.requestsPerSecond + Math.floor((Math.random() - 0.5) * 20));
      node.health.avgResponseTime = Math.max(50, Math.min(1000, node.health.avgResponseTime + (Math.random() - 0.5) * 100));
      node.health.lastCheck = Date.now();
      return;
    }

    try {
      const response = await fetch(`${node.url}/api/tags`);
      if (response.ok) {
        const data = await response.json();
        node.status = 'healthy';
        node.health.modelsLoaded = data.models?.map(m => m.name) || [];
        node.health.lastCheck = Date.now();
      } else {
        node.status = 'warning';
      }
    } catch (error) {
      node.status = 'error';
      console.error(`Health check failed for ${node.name}:`, error.message);
    }
  }

  updateNodeStatus(id, status) {
    const node = this.nodes.get(id);
    if (node) {
      node.status = status;
      node.health.lastCheck = Date.now();
    }
  }

  removeNode(id) {
    const node = this.nodes.get(id);
    if (node) {
      if (node.connection) {
        node.connection.close();
      }
      this.nodes.delete(id);
    }
  }

  getHealthyNodes() {
    return Array.from(this.nodes.values()).filter(n => n.status === 'healthy');
  }

  getNodes() {
    // Alias for getAllNodes - returns raw node objects
    return Array.from(this.nodes.values());
  }

  getAllNodes() {
    return Array.from(this.nodes.values()).map(node => ({
      id: node.id,
      name: node.name,
      status: node.status,
      mock: node.mock,
      ...node.health
    }));
  }

  selectNode(strategy = 'round-robin') {
    const healthyNodes = this.getHealthyNodes();
    if (healthyNodes.length === 0) return null;
    
    return this.loadBalancer.select(healthyNodes, strategy);
  }

  startHealthChecks() {
    // Check health every 5 seconds
    this.healthCheckInterval = setInterval(() => {
      this.nodes.forEach(node => {
        this.updateNodeHealth(node);
      });
    }, 5000);
  }

  stopHealthChecks() {
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
    }
  }
}

module.exports = NodeRegistry;

