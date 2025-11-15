/**
 * Node Registry Service
 * Manages distributed Ollama nodes and their health status
 */

const fetch = require('node-fetch');

class LoadBalancer {
  constructor() {
    this.currentIndex = 0;
    this.pinnedSessions = new Map(); // sessionId -> nodeId
    this.nodeScores = new Map(); // nodeId -> score history
    this.lastScoreUpdate = Date.now();
  }

  select(nodes, strategy, options = {}) {
    switch(strategy) {
      case 'round-robin':
        return this.roundRobin(nodes);
      case 'least-loaded':
        return this.leastLoaded(nodes);
      case 'fastest':
        return this.fastest(nodes);
      case 'auto':
        return this.autoSelect(nodes, options);
      case 'single-node':
        return this.singleNode(nodes, options);
      case 'pinned':
        return this.pinnedSelect(nodes, options);
      case 'broadcast':
        return this.broadcastSelect(nodes, options);
      case 'geographic':
        return this.geographicSelect(nodes, options);
      case 'resource-aware':
        return this.resourceAwareSelect(nodes, options);
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

  // Auto mode: intelligently selects best strategy based on current conditions
  autoSelect(nodes, options) {
    const { taskType, modelName, sessionId } = options;

    // For chat sessions, prefer pinned routing for consistency
    if (sessionId && taskType === 'chat') {
      return this.pinnedSelect(nodes, options);
    }

    // For model-specific tasks, prefer nodes with the model loaded
    if (modelName) {
      const nodesWithModel = nodes.filter(node =>
        node.health.modelsLoaded && node.health.modelsLoaded.includes(modelName)
      );
      if (nodesWithModel.length > 0) {
        return this.leastLoaded(nodesWithModel);
      }
    }

    // Default to resource-aware selection
    return this.resourceAwareSelect(nodes, options);
  }

  // Single node mode: routes all requests to the best single node
  singleNode(nodes, options) {
    const { preferredNodeId } = options;

    if (preferredNodeId) {
      const preferredNode = nodes.find(node => node.id === preferredNodeId);
      if (preferredNode) return preferredNode;
    }

    // Select the best overall node based on composite score
    return this.getBestNode(nodes);
  }

  // Pinned mode: maintains session affinity
  pinnedSelect(nodes, options) {
    const { sessionId } = options;

    if (!sessionId) {
      return this.autoSelect(nodes, options);
    }

    // Check if we have a pinned node for this session
    const pinnedNodeId = this.pinnedSessions.get(sessionId);
    if (pinnedNodeId) {
      const pinnedNode = nodes.find(node => node.id === pinnedNodeId);
      if (pinnedNode && pinnedNode.status === 'healthy') {
        return pinnedNode;
      }
      // Pinned node is unavailable, remove from cache
      this.pinnedSessions.delete(sessionId);
    }

    // Select new node and pin it
    const selectedNode = this.getBestNode(nodes);
    this.pinnedSessions.set(sessionId, selectedNode.id);

    // Clean up old sessions (older than 1 hour)
    this.cleanupPinnedSessions();

    return selectedNode;
  }

  // Broadcast mode: returns multiple nodes for parallel processing
  broadcastSelect(nodes, options) {
    const { maxNodes = 3, minNodes = 1 } = options;

    // Sort nodes by performance score
    const sortedNodes = [...nodes].sort((a, b) =>
      this.calculateNodeScore(b) - this.calculateNodeScore(a)
    );

    const selectedCount = Math.min(Math.max(minNodes, sortedNodes.length), maxNodes);
    return sortedNodes.slice(0, selectedCount);
  }

  // Geographic selection: prefers nodes in the same region/zone
  geographicSelect(nodes, options) {
    const { preferredRegion, preferredZone } = options;

    // Filter by region if specified
    let candidateNodes = nodes;
    if (preferredRegion) {
      const regionalNodes = nodes.filter(node =>
        node.region === preferredRegion || node.health.region === preferredRegion
      );
      if (regionalNodes.length > 0) {
        candidateNodes = regionalNodes;
      }
    }

    // Further filter by zone if specified
    if (preferredZone) {
      const zonalNodes = candidateNodes.filter(node =>
        node.zone === preferredZone || node.health.zone === preferredZone
      );
      if (zonalNodes.length > 0) {
        candidateNodes = zonalNodes;
      }
    }

    return this.leastLoaded(candidateNodes);
  }

  // Resource-aware selection: considers CPU, memory, and current load
  resourceAwareSelect(nodes, options) {
    const { requiredCpu = 0, requiredMemory = 0, requiredGpu = false } = options;

    // Filter nodes that meet resource requirements
    const suitableNodes = nodes.filter(node => {
      const health = node.health;
      const availableCpu = 100 - (health.cpu || 0);
      const availableMemory = 100 - (health.memory || 0);

      return availableCpu >= requiredCpu &&
             availableMemory >= requiredMemory &&
             (!requiredGpu || health.hasGpu);
    });

    if (suitableNodes.length === 0) {
      // No nodes meet requirements, fall back to best available
      return this.getBestNode(nodes);
    }

    // Select node with best resource availability
    return suitableNodes.reduce((best, node) => {
      const bestScore = this.calculateResourceScore(best);
      const nodeScore = this.calculateResourceScore(node);
      return nodeScore > bestScore ? node : best;
    });
  }

  // Helper methods
  getBestNode(nodes) {
    return nodes.reduce((best, node) => {
      const bestScore = this.calculateNodeScore(best);
      const nodeScore = this.calculateNodeScore(node);
      return nodeScore > bestScore ? node : best;
    });
  }

  calculateNodeScore(node) {
    const health = node.health;
    const now = Date.now();

    // Base score factors
    const loadScore = Math.max(0, 100 - (health.load || 50)) / 100;
    const cpuScore = Math.max(0, 100 - (health.cpu || 50)) / 100;
    const memoryScore = Math.max(0, 100 - (health.memory || 50)) / 100;
    const responseTimeScore = Math.max(0, 1000 - (health.avgResponseTime || 500)) / 1000;
    const uptimeScore = Math.min(1, (now - (health.uptime || now)) / (24 * 60 * 60 * 1000));

    // Weighted composite score
    return (
      loadScore * 0.3 +
      cpuScore * 0.2 +
      memoryScore * 0.2 +
      responseTimeScore * 0.2 +
      uptimeScore * 0.1
    );
  }

  calculateResourceScore(node) {
    const health = node.health;
    const availableCpu = Math.max(0, 100 - (health.cpu || 0));
    const availableMemory = Math.max(0, 100 - (health.memory || 0));
    const loadFactor = Math.max(0, 100 - (health.load || 0));

    return (availableCpu + availableMemory + loadFactor) / 3;
  }

  cleanupPinnedSessions() {
    const now = Date.now();
    const oneHour = 60 * 60 * 1000;

    for (const [sessionId, nodeId] of this.pinnedSessions.entries()) {
      // This is a simple cleanup - in production, you'd track session timestamps
      if (Math.random() < 0.01) { // 1% chance to clean up on each call
        this.pinnedSessions.delete(sessionId);
      }
    }
  }

  // Get routing statistics
  getRoutingStats() {
    return {
      pinnedSessions: this.pinnedSessions.size,
      currentIndex: this.currentIndex,
      lastScoreUpdate: this.lastScoreUpdate
    };
  }

  // Clear session affinity
  clearSessionAffinity(sessionId) {
    if (sessionId) {
      this.pinnedSessions.delete(sessionId);
    } else {
      this.pinnedSessions.clear();
    }
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

  selectNode(strategy = 'round-robin', options = {}) {
    const healthyNodes = this.getHealthyNodes();
    if (healthyNodes.length === 0) return null;

    const result = this.loadBalancer.select(healthyNodes, strategy, options);

    // Handle broadcast mode returning multiple nodes
    if (Array.isArray(result) && strategy === 'broadcast') {
      return result;
    }

    return result;
  }

  // Enhanced node selection with detailed options
  selectNodeWithOptions(options = {}) {
    const {
      strategy = 'round-robin',
      sessionId,
      modelName,
      taskType,
      requiredCpu,
      requiredMemory,
      requiredGpu,
      preferredRegion,
      preferredZone,
      preferredNodeId,
      maxNodes,
      minNodes
    } = options;

    return this.selectNode(strategy, {
      sessionId,
      modelName,
      taskType,
      requiredCpu,
      requiredMemory,
      requiredGpu,
      preferredRegion,
      preferredZone,
      preferredNodeId,
      maxNodes,
      minNodes
    });
  }

  // Get routing statistics
  getRoutingStats() {
    return {
      totalNodes: this.nodes.size,
      healthyNodes: this.getHealthyNodes().length,
      loadBalancer: this.loadBalancer.getRoutingStats()
    };
  }

  // Clear session affinity for a specific session or all sessions
  clearSessionAffinity(sessionId = null) {
    this.loadBalancer.clearSessionAffinity(sessionId);
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

