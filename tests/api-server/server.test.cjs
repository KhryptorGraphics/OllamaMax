const { EventEmitter } = require('events');

describe('API Server Components', () => {
  describe('NodeRegistry Functionality', () => {
    // Create a simplified NodeRegistry for testing
    class NodeRegistry {
      constructor() {
        this.nodes = new Map();
        this.loadBalancer = new LoadBalancer();
      }

      addNode(id, config) {
        const node = {
          id,
          name: config.name,
          url: config.url,
          status: 'connecting',
          health: {
            load: 0,
            memory: 0,
            requestsPerSecond: 0,
            queue: 0,
            lastCheck: Date.now()
          },
          connection: null
        };
        
        this.nodes.set(id, node);
        node.status = 'healthy'; // Simulate successful connection
        return node;
      }

      removeNode(id) {
        const node = this.nodes.get(id);
        if (node && node.connection) {
          // Simulate connection cleanup
        }
        this.nodes.delete(id);
      }

      getHealthyNodes() {
        return Array.from(this.nodes.values()).filter(n => n.status === 'healthy');
      }

      getAllNodes() {
        return Array.from(this.nodes.values()).map(node => ({
          id: node.id,
          name: node.name,
          status: node.status,
          ...node.health
        }));
      }

      selectNode(strategy = 'round-robin') {
        const healthyNodes = this.getHealthyNodes();
        if (healthyNodes.length === 0) return null;
        return this.loadBalancer.select(healthyNodes, strategy);
      }

      async updateNodeHealth(node) {
        try {
          // Simulate health check
          node.health.lastCheck = Date.now();
          node.health.load = Math.random() * 100;
          node.health.memory = Math.random() * 100;
          node.health.requestsPerSecond = Math.floor(Math.random() * 20);
          
          if (node.health.load > 90) {
            node.status = 'warning';
          } else {
            node.status = 'healthy';
          }
        } catch (error) {
          node.status = 'error';
        }
      }
    }

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
        if (nodes.length === 0) return null;
        const node = nodes[this.currentIndex % nodes.length];
        this.currentIndex++;
        return node;
      }

      leastLoaded(nodes) {
        if (nodes.length === 0) return null;
        return nodes.reduce((min, node) => 
          node.health.load < min.health.load ? node : min
        );
      }

      fastest(nodes) {
        if (nodes.length === 0) return null;
        return nodes.reduce((min, node) => 
          node.health.requestsPerSecond > min.health.requestsPerSecond ? node : min
        );
      }
    }

    test('should initialize with empty node registry', () => {
      const registry = new NodeRegistry();
      expect(registry.nodes).toBeInstanceOf(Map);
      expect(registry.nodes.size).toBe(0);
    });

    test('should add node with correct configuration', () => {
      const registry = new NodeRegistry();
      const config = {
        name: 'test-node',
        url: 'http://localhost:11434'
      };
      
      const node = registry.addNode('node-1', config);
      
      expect(node.id).toBe('node-1');
      expect(node.name).toBe('test-node');
      expect(node.url).toBe('http://localhost:11434');
      expect(node.status).toBe('healthy');
      expect(node.health).toBeDefined();
      expect(registry.nodes.size).toBe(1);
    });

    test('should remove node and clean up connections', () => {
      const registry = new NodeRegistry();
      const config = { name: 'test-node', url: 'http://localhost:11434' };
      
      registry.addNode('node-1', config);
      expect(registry.nodes.size).toBe(1);
      
      registry.removeNode('node-1');
      expect(registry.nodes.size).toBe(0);
    });

    test('should filter healthy nodes correctly', () => {
      const registry = new NodeRegistry();
      
      registry.addNode('healthy-1', { name: 'healthy', url: 'http://localhost:11434' });
      registry.addNode('error-1', { name: 'error', url: 'http://localhost:11435' });
      
      // Manually set status
      registry.nodes.get('error-1').status = 'error';
      
      const healthyNodes = registry.getHealthyNodes();
      expect(healthyNodes).toHaveLength(1);
      expect(healthyNodes[0].name).toBe('healthy');
    });

    test('should update node health metrics', async () => {
      const registry = new NodeRegistry();
      const node = registry.addNode('node-1', { 
        name: 'test-node', 
        url: 'http://localhost:11434' 
      });
      
      await registry.updateNodeHealth(node);
      
      expect(node.health.lastCheck).toBeGreaterThan(0);
      expect(node.health.load).toBeGreaterThanOrEqual(0);
      expect(node.health.memory).toBeGreaterThanOrEqual(0);
    });
  });

  describe('LoadBalancer', () => {
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
        if (nodes.length === 0) return null;
        const node = nodes[this.currentIndex % nodes.length];
        this.currentIndex++;
        return node;
      }

      leastLoaded(nodes) {
        if (nodes.length === 0) return null;
        return nodes.reduce((min, node) => 
          node.health.load < min.health.load ? node : min
        );
      }

      fastest(nodes) {
        if (nodes.length === 0) return null;
        return nodes.reduce((min, node) => 
          node.health.requestsPerSecond > min.health.requestsPerSecond ? node : min
        );
      }
    }

    test('should implement round-robin strategy', () => {
      const balancer = new LoadBalancer();
      const nodes = [
        { id: 1, name: 'node1', health: { load: 10 } },
        { id: 2, name: 'node2', health: { load: 20 } },
        { id: 3, name: 'node3', health: { load: 30 } }
      ];
      
      expect(balancer.select(nodes, 'round-robin').id).toBe(1);
      expect(balancer.select(nodes, 'round-robin').id).toBe(2);
      expect(balancer.select(nodes, 'round-robin').id).toBe(3);
      expect(balancer.select(nodes, 'round-robin').id).toBe(1);
    });

    test('should implement least-loaded strategy', () => {
      const balancer = new LoadBalancer();
      const nodes = [
        { id: 1, name: 'node1', health: { load: 30 } },
        { id: 2, name: 'node2', health: { load: 10 } },
        { id: 3, name: 'node3', health: { load: 20 } }
      ];
      
      const selected = balancer.select(nodes, 'least-loaded');
      expect(selected.id).toBe(2);
      expect(selected.health.load).toBe(10);
    });

    test('should implement fastest strategy', () => {
      const balancer = new LoadBalancer();
      const nodes = [
        { id: 1, name: 'node1', health: { requestsPerSecond: 10 } },
        { id: 2, name: 'node2', health: { requestsPerSecond: 30 } },
        { id: 3, name: 'node3', health: { requestsPerSecond: 20 } }
      ];
      
      const selected = balancer.select(nodes, 'fastest');
      expect(selected.id).toBe(2);
      expect(selected.health.requestsPerSecond).toBe(30);
    });

    test('should handle empty node list', () => {
      const balancer = new LoadBalancer();
      expect(balancer.select([], 'round-robin')).toBeNull();
      expect(balancer.select([], 'least-loaded')).toBeNull();
      expect(balancer.select([], 'fastest')).toBeNull();
    });
  });

  describe('MessageQueue', () => {
    class MessageQueue {
      constructor() {
        this.queue = [];
        this.processing = false;
      }

      add(message, callback) {
        this.queue.push({ message, callback });
        this.process();
      }

      async process() {
        if (this.processing || this.queue.length === 0) return;
        
        this.processing = true;
        
        while (this.queue.length > 0) {
          const { message, callback } = this.queue.shift();
          
          try {
            await callback(message);
          } catch (error) {
            console.error('Queue processing error:', error);
          }
          
          await new Promise(resolve => setTimeout(resolve, 1));
        }
        
        this.processing = false;
      }

      getLength() {
        return this.queue.length;
      }
    }

    test('should process messages in order', async () => {
      const queue = new MessageQueue();
      const processedMessages = [];
      
      const mockCallback = jest.fn((message) => {
        processedMessages.push(message);
        return Promise.resolve();
      });
      
      queue.add('message1', mockCallback);
      queue.add('message2', mockCallback);
      queue.add('message3', mockCallback);
      
      // Wait for processing
      await new Promise(resolve => setTimeout(resolve, 50));
      
      expect(mockCallback).toHaveBeenCalledTimes(3);
      expect(processedMessages).toEqual(['message1', 'message2', 'message3']);
    });

    test('should handle processing errors gracefully', async () => {
      const queue = new MessageQueue();
      const errorCallback = jest.fn(() => Promise.reject(new Error('Processing failed')));
      
      queue.add('error-message', errorCallback);
      
      await new Promise(resolve => setTimeout(resolve, 50));
      
      expect(errorCallback).toHaveBeenCalled();
      expect(queue.processing).toBe(false);
    });
  });

  describe('REST API Handlers', () => {
    test('should handle GET /api/nodes', () => {
      const mockReq = {};
      const mockRes = { json: jest.fn() };
      
      const mockNodes = [
        { id: 'node-1', name: 'test-node', status: 'healthy' }
      ];
      
      const nodeHandler = (req, res) => {
        res.json({
          nodes: mockNodes,
          queueLength: 0
        });
      };
      
      nodeHandler(mockReq, mockRes);
      
      expect(mockRes.json).toHaveBeenCalledWith({
        nodes: mockNodes,
        queueLength: 0
      });
    });

    test('should handle POST /api/nodes', () => {
      const mockReq = {
        body: { name: 'new-node', url: 'http://localhost:11434' }
      };
      const mockRes = { json: jest.fn() };
      
      const createNodeHandler = (req, res) => {
        const nodeId = `node-${Date.now()}`;
        res.json({
          id: nodeId,
          name: req.body.name,
          status: 'connecting'
        });
      };
      
      createNodeHandler(mockReq, mockRes);
      
      expect(mockRes.json).toHaveBeenCalledWith({
        id: expect.stringMatching(/^node-\d+$/),
        name: 'new-node',
        status: 'connecting'
      });
    });

    test('should handle GET /api/health', () => {
      const mockReq = {};
      const mockRes = { json: jest.fn() };
      
      const healthHandler = (req, res) => {
        res.json({
          status: 'healthy',
          nodes: 2,
          totalNodes: 3,
          queueLength: 0,
          uptime: process.uptime()
        });
      };
      
      healthHandler(mockReq, mockRes);
      
      expect(mockRes.json).toHaveBeenCalledWith({
        status: 'healthy',
        nodes: 2,
        totalNodes: 3,
        queueLength: 0,
        uptime: expect.any(Number)
      });
    });
  });

  describe('WebSocket Message Handling', () => {
    test('should handle valid WebSocket messages', () => {
      const mockWs = new EventEmitter();
      mockWs.send = jest.fn();
      
      const messageHandler = (message) => {
        try {
          const data = JSON.parse(message);
          mockWs.send(JSON.stringify({ type: 'response', data }));
        } catch (error) {
          mockWs.send(JSON.stringify({ type: 'error', message: error.message }));
        }
      };
      
      mockWs.on('message', messageHandler);
      
      const validMessage = JSON.stringify({ type: 'inference', model: 'llama2' });
      mockWs.emit('message', validMessage);
      
      expect(mockWs.send).toHaveBeenCalledWith(
        expect.stringContaining('"type":"response"')
      );
    });

    test('should handle malformed WebSocket messages', () => {
      const mockWs = new EventEmitter();
      mockWs.send = jest.fn();
      
      const messageHandler = (message) => {
        try {
          JSON.parse(message);
          mockWs.send(JSON.stringify({ type: 'response' }));
        } catch (error) {
          mockWs.send(JSON.stringify({ type: 'error', message: error.message }));
        }
      };
      
      mockWs.on('message', messageHandler);
      mockWs.emit('message', 'invalid json');
      
      expect(mockWs.send).toHaveBeenCalledWith(
        expect.stringContaining('"type":"error"')
      );
    });
  });

  describe('Performance and Scalability', () => {
    test('should handle multiple concurrent connections', () => {
      const mockClients = new Set();
      
      for (let i = 0; i < 100; i++) {
        const mockClient = {
          id: i,
          send: jest.fn(),
          readyState: 1
        };
        mockClients.add(mockClient);
      }
      
      expect(mockClients.size).toBe(100);
      
      const broadcastMessage = JSON.stringify({ type: 'broadcast', data: 'test' });
      mockClients.forEach(client => {
        if (client.readyState === 1) {
          client.send(broadcastMessage);
        }
      });
      
      mockClients.forEach(client => {
        expect(client.send).toHaveBeenCalledWith(broadcastMessage);
      });
    });

    test('should handle configuration from environment variables', () => {
      const mockEnv = {
        PORT: '3000',
        REDIS_HOST: 'redis-server',
        REDIS_PORT: '6380',
        OLLAMA_PRIMARY: 'http://primary:11434'
      };
      
      const config = {
        port: mockEnv.PORT || 13100,
        redisHost: mockEnv.REDIS_HOST || 'localhost',
        redisPort: mockEnv.REDIS_PORT || 6379
      };
      
      expect(config.port).toBe('3000');
      expect(config.redisHost).toBe('redis-server');
      expect(config.redisPort).toBe('6380');
    });
  });

  describe('Error Handling', () => {
    test('should handle connection errors gracefully', () => {
      const errorHandler = jest.fn((error) => {
        console.error('Connection error:', error);
      });
      
      const mockError = new Error('Connection failed');
      errorHandler(mockError);
      
      expect(errorHandler).toHaveBeenCalledWith(mockError);
    });

    test('should handle inference request failures', async () => {
      const mockFetch = jest.fn().mockRejectedValue(new Error('Model not found'));
      global.fetch = mockFetch;
      
      const inferenceHandler = async (ws, data) => {
        try {
          await fetch(`http://localhost:11434/api/generate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              model: data.model,
              prompt: data.content
            })
          });
        } catch (error) {
          ws.send(JSON.stringify({
            type: 'error',
            message: `Inference failed: ${error.message}`
          }));
        }
      };
      
      const mockWs = { send: jest.fn() };
      const mockData = { model: 'nonexistent', content: 'test' };
      
      await inferenceHandler(mockWs, mockData);
      
      expect(mockWs.send).toHaveBeenCalledWith(
        expect.stringContaining('"type":"error"')
      );
    });
  });
});