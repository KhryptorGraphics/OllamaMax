/**
 * Ollama Node Connector
 * Discovers, registers, and manages real Ollama nodes
 */

const fetch = require('node-fetch');
const EventEmitter = require('events');

class OllamaConnector extends EventEmitter {
  constructor(nodeRegistry) {
    super();
    this.nodeRegistry = nodeRegistry;
    this.discoveryInterval = null;
    this.healthCheckInterval = null;
    this.config = {
      discoveryIntervalMs: 30000, // 30 seconds
      healthCheckIntervalMs: 10000, // 10 seconds
      nodeTimeout: 5000, // 5 seconds
      maxRetries: 3
    };
  }

  /**
   * Start automatic node discovery and health monitoring
   */
  start() {
    console.log('🔍 Starting Ollama node discovery...');
    
    // Initial discovery
    this.discoverNodes();
    
    // Periodic discovery
    this.discoveryInterval = setInterval(() => {
      this.discoverNodes();
    }, this.config.discoveryIntervalMs);
    
    // Periodic health checks
    this.healthCheckInterval = setInterval(() => {
      this.performHealthChecks();
    }, this.config.healthCheckIntervalMs);
    
    console.log('✓ Ollama connector started');
  }

  /**
   * Stop the connector
   */
  stop() {
    if (this.discoveryInterval) {
      clearInterval(this.discoveryInterval);
      this.discoveryInterval = null;
    }
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
      this.healthCheckInterval = null;
    }
    console.log('✓ Ollama connector stopped');
  }

  /**
   * Discover Ollama nodes from environment and configuration
   */
  async discoverNodes() {
    const nodeUrls = this.getNodeUrlsFromConfig();
    
    for (const url of nodeUrls) {
      try {
        await this.registerNode(url);
      } catch (error) {
        console.error(`Failed to register node ${url}:`, error.message);
      }
    }
  }

  /**
   * Get node URLs from environment variables
   */
  getNodeUrlsFromConfig() {
    const urls = [];
    
    // Check OLLAMA_NODES environment variable (comma-separated)
    if (process.env.OLLAMA_NODES) {
      const nodes = process.env.OLLAMA_NODES.split(',').map(n => n.trim());
      urls.push(...nodes);
    }
    
    // Check individual OLLAMA_NODE_* variables
    for (let i = 1; i <= 10; i++) {
      const nodeUrl = process.env[`OLLAMA_NODE_${i}`];
      if (nodeUrl) {
        urls.push(nodeUrl.trim());
      }
    }
    
    // Default local node if none configured
    if (urls.length === 0 && process.env.ENABLE_LOCAL_OLLAMA === 'true') {
      urls.push('http://localhost:11434');
    }
    
    return [...new Set(urls)]; // Remove duplicates
  }

  /**
   * Register a new Ollama node
   */
  async registerNode(url) {
    try {
      // Check if node is already registered
      const existingNodes = this.nodeRegistry.getNodes();
      const exists = existingNodes.some(n => n.url === url);
      
      if (exists) {
        return; // Already registered
      }

      // Verify node is accessible
      const info = await this.getNodeInfo(url);
      
      if (!info) {
        throw new Error('Failed to get node info');
      }

      // Get available models
      const models = await this.getNodeModels(url);

      // Register with node registry
      const nodeId = `ollama-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      
      this.nodeRegistry.addNode({
        id: nodeId,
        name: `Ollama Node (${new URL(url).hostname})`,
        url,
        status: 'healthy',
        mock: false,
        version: info.version || 'unknown',
        models: models.map(m => m.name),
        capabilities: {
          inference: true,
          embeddings: true,
          streaming: true
        },
        metrics: {
          load: 0,
          memory: 0,
          requestsPerSecond: 0,
          queue: 0
        }
      });

      console.log(`✓ Registered Ollama node: ${url}`);
      this.emit('node-registered', { url, nodeId, models: models.length });
      
    } catch (error) {
      console.error(`Failed to register node ${url}:`, error.message);
      throw error;
    }
  }

  /**
   * Get node information
   */
  async getNodeInfo(url) {
    try {
      const response = await fetch(`${url}/api/version`, {
        timeout: this.config.nodeTimeout
      });

      if (!response.ok) {
        return null;
      }

      return await response.json();
    } catch (error) {
      return null;
    }
  }

  /**
   * Get available models on a node
   */
  async getNodeModels(url) {
    try {
      const response = await fetch(`${url}/api/tags`, {
        timeout: this.config.nodeTimeout
      });

      if (!response.ok) {
        return [];
      }

      const data = await response.json();
      return data.models || [];
    } catch (error) {
      return [];
    }
  }

  /**
   * Perform health checks on all registered nodes
   */
  async performHealthChecks() {
    const nodes = this.nodeRegistry.getNodes();
    const realNodes = nodes.filter(n => !n.mock);

    for (const node of realNodes) {
      try {
        const isHealthy = await this.checkNodeHealth(node.url);

        if (isHealthy) {
          this.nodeRegistry.updateNodeStatus(node.id, 'healthy');
          this.emit('node-healthy', { nodeId: node.id, url: node.url });
        } else {
          this.nodeRegistry.updateNodeStatus(node.id, 'unhealthy');
          this.emit('node-unhealthy', { nodeId: node.id, url: node.url });
        }
      } catch (error) {
        this.nodeRegistry.updateNodeStatus(node.id, 'error');
        this.emit('node-error', { nodeId: node.id, url: node.url, error: error.message });
      }
    }
  }

  /**
   * Check if a node is healthy
   */
  async checkNodeHealth(url) {
    try {
      const response = await fetch(`${url}/api/version`, {
        timeout: this.config.nodeTimeout
      });

      return response.ok;
    } catch (error) {
      return false;
    }
  }

  /**
   * Remove a node
   */
  removeNode(nodeId) {
    this.nodeRegistry.removeNode(nodeId);
    this.emit('node-removed', { nodeId });
    console.log(`✓ Removed node: ${nodeId}`);
  }

  /**
   * Get connector statistics
   */
  getStats() {
    const nodes = this.nodeRegistry.getNodes();
    const realNodes = nodes.filter(n => !n.mock);

    return {
      totalNodes: nodes.length,
      realNodes: realNodes.length,
      mockNodes: nodes.length - realNodes.length,
      healthyNodes: realNodes.filter(n => n.status === 'healthy').length,
      unhealthyNodes: realNodes.filter(n => n.status !== 'healthy').length,
      discoveryInterval: this.config.discoveryIntervalMs,
      healthCheckInterval: this.config.healthCheckIntervalMs
    };
  }

  /**
   * Test connection to a node
   */
  async testConnection(url) {
    try {
      const info = await this.getNodeInfo(url);
      if (!info) {
        return { success: false, error: 'Failed to connect to node' };
      }

      const models = await this.getNodeModels(url);

      return {
        success: true,
        info,
        models: models.length,
        modelList: models.map(m => m.name)
      };
    } catch (error) {
      return {
        success: false,
        error: error.message
      };
    }
  }
}

module.exports = OllamaConnector;

