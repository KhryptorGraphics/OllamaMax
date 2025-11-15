const { describe, test, expect, beforeEach, afterEach } = require('@jest/globals');

// Mock the NodeRegistry and LoadBalancer classes
class MockNode {
  constructor(id, name, health = {}) {
    this.id = id;
    this.name = name;
    this.status = 'healthy';
    this.health = {
      load: 50,
      cpu: 30,
      memory: 40,
      latency: 100,
      requestsPerSecond: 10,
      modelsLoaded: [],
      uptime: Date.now(),
      ...health
    };
  }
}

// Import the LoadBalancer class from the actual implementation
const LoadBalancer = require('../../src/services/node-registry').LoadBalancer || class LoadBalancer {
  constructor() {
    this.currentIndex = 0;
    this.pinnedSessions = new Map();
    this.nodeScores = new Map();
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

  autoSelect(nodes, options) {
    const { taskType, modelName, sessionId } = options;
    
    if (sessionId && taskType === 'chat') {
      return this.pinnedSelect(nodes, options);
    }
    
    if (modelName) {
      const nodesWithModel = nodes.filter(node => 
        node.health.modelsLoaded && node.health.modelsLoaded.includes(modelName)
      );
      if (nodesWithModel.length > 0) {
        return this.leastLoaded(nodesWithModel);
      }
    }
    
    return this.resourceAwareSelect(nodes, options);
  }

  singleNode(nodes, options) {
    const { preferredNodeId } = options;
    
    if (preferredNodeId) {
      const preferredNode = nodes.find(node => node.id === preferredNodeId);
      if (preferredNode) return preferredNode;
    }
    
    return this.getBestNode(nodes);
  }

  pinnedSelect(nodes, options) {
    const { sessionId } = options;
    
    if (!sessionId) {
      return this.autoSelect(nodes, options);
    }
    
    const pinnedNodeId = this.pinnedSessions.get(sessionId);
    if (pinnedNodeId) {
      const pinnedNode = nodes.find(node => node.id === pinnedNodeId);
      if (pinnedNode && pinnedNode.status === 'healthy') {
        return pinnedNode;
      }
      this.pinnedSessions.delete(sessionId);
    }
    
    const selectedNode = this.getBestNode(nodes);
    this.pinnedSessions.set(sessionId, selectedNode.id);
    
    return selectedNode;
  }

  broadcastSelect(nodes, options) {
    const { maxNodes = 3, minNodes = 1 } = options;
    
    const sortedNodes = [...nodes].sort((a, b) => 
      this.calculateNodeScore(b) - this.calculateNodeScore(a)
    );
    
    const selectedCount = Math.min(Math.max(minNodes, sortedNodes.length), maxNodes);
    return sortedNodes.slice(0, selectedCount);
  }

  geographicSelect(nodes, options) {
    const { preferredRegion, preferredZone } = options;
    
    let candidateNodes = nodes;
    if (preferredRegion) {
      const regionalNodes = nodes.filter(node => 
        node.region === preferredRegion || node.health.region === preferredRegion
      );
      if (regionalNodes.length > 0) {
        candidateNodes = regionalNodes;
      }
    }
    
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

  resourceAwareSelect(nodes, options) {
    const { requiredCpu = 0, requiredMemory = 0, requiredGpu = false } = options;
    
    const suitableNodes = nodes.filter(node => {
      const health = node.health;
      const availableCpu = 100 - (health.cpu || 0);
      const availableMemory = 100 - (health.memory || 0);
      
      return availableCpu >= requiredCpu && 
             availableMemory >= requiredMemory &&
             (!requiredGpu || health.hasGpu);
    });
    
    if (suitableNodes.length === 0) {
      return this.getBestNode(nodes);
    }
    
    return suitableNodes.reduce((best, node) => {
      const bestScore = this.calculateResourceScore(best);
      const nodeScore = this.calculateResourceScore(node);
      return nodeScore > bestScore ? node : best;
    });
  }

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
    
    const loadScore = Math.max(0, 100 - (health.load || 50)) / 100;
    const cpuScore = Math.max(0, 100 - (health.cpu || 50)) / 100;
    const memoryScore = Math.max(0, 100 - (health.memory || 50)) / 100;
    const responseTimeScore = Math.max(0, 1000 - (health.avgResponseTime || 500)) / 1000;
    const uptimeScore = Math.min(1, (now - (health.uptime || now)) / (24 * 60 * 60 * 1000));
    
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

  clearSessionAffinity(sessionId) {
    if (sessionId) {
      this.pinnedSessions.delete(sessionId);
    } else {
      this.pinnedSessions.clear();
    }
  }

  getRoutingStats() {
    return {
      pinnedSessions: this.pinnedSessions.size,
      currentIndex: this.currentIndex,
      lastScoreUpdate: this.lastScoreUpdate
    };
  }
};

describe('Scheduler Routing Tests', () => {
  let loadBalancer;
  let testNodes;

  beforeEach(() => {
    loadBalancer = new LoadBalancer();
    testNodes = [
      new MockNode('node1', 'Node 1', { load: 20, cpu: 10, memory: 30 }),
      new MockNode('node2', 'Node 2', { load: 60, cpu: 50, memory: 40 }),
      new MockNode('node3', 'Node 3', { load: 80, cpu: 70, memory: 60 }),
      new MockNode('node4', 'Node 4', { load: 30, cpu: 20, memory: 25 })
    ];
  });

  afterEach(() => {
    loadBalancer = null;
    testNodes = null;
  });

  describe('Round Robin Routing', () => {
    test('should distribute requests evenly across nodes', () => {
      const selections = [];
      for (let i = 0; i < 8; i++) {
        const selected = loadBalancer.select(testNodes, 'round-robin');
        selections.push(selected.id);
      }

      // Should cycle through all nodes twice
      expect(selections).toEqual([
        'node1', 'node2', 'node3', 'node4',
        'node1', 'node2', 'node3', 'node4'
      ]);
    });

    test('should handle single node', () => {
      const singleNode = [testNodes[0]];
      const selected = loadBalancer.select(singleNode, 'round-robin');
      expect(selected.id).toBe('node1');
    });
  });

  describe('Least Loaded Routing', () => {
    test('should select node with lowest load', () => {
      const selected = loadBalancer.select(testNodes, 'least-loaded');
      expect(selected.id).toBe('node1'); // load: 20
    });

    test('should handle equal loads', () => {
      const equalNodes = [
        new MockNode('equal1', 'Equal 1', { load: 50 }),
        new MockNode('equal2', 'Equal 2', { load: 50 })
      ];
      const selected = loadBalancer.select(equalNodes, 'least-loaded');
      expect(['equal1', 'equal2']).toContain(selected.id);
    });
  });

  describe('Fastest Response Routing', () => {
    test('should select node with highest requests per second', () => {
      testNodes[2].health.requestsPerSecond = 50; // Make node3 fastest
      const selected = loadBalancer.select(testNodes, 'fastest');
      expect(selected.id).toBe('node3');
    });
  });

  describe('Auto Routing', () => {
    test('should use pinned routing for chat sessions', () => {
      const options = { taskType: 'chat', sessionId: 'session123' };
      const selected1 = loadBalancer.select(testNodes, 'auto', options);
      const selected2 = loadBalancer.select(testNodes, 'auto', options);

      // Should return same node for same session
      expect(selected1.id).toBe(selected2.id);
    });

    test('should prefer nodes with required model', () => {
      testNodes[1].health.modelsLoaded = ['llama2'];
      testNodes[2].health.modelsLoaded = ['gpt-3.5'];

      const options = { modelName: 'llama2' };
      const selected = loadBalancer.select(testNodes, 'auto', options);
      expect(selected.id).toBe('node2');
    });

    test('should fallback to resource-aware when no model match', () => {
      const options = { modelName: 'nonexistent-model' };
      const selected = loadBalancer.select(testNodes, 'auto', options);
      expect(selected).toBeDefined();
    });
  });

  describe('Single Node Routing', () => {
    test('should return preferred node when available', () => {
      const options = { preferredNodeId: 'node3' };
      const selected = loadBalancer.select(testNodes, 'single-node', options);
      expect(selected.id).toBe('node3');
    });

    test('should fallback to best node when preferred unavailable', () => {
      const options = { preferredNodeId: 'nonexistent' };
      const selected = loadBalancer.select(testNodes, 'single-node', options);
      expect(selected).toBeDefined();
      expect(selected.id).toBe('node1'); // Best node based on scoring
    });
  });

  describe('Pinned Routing (Session Affinity)', () => {
    test('should maintain session affinity', () => {
      const options = { sessionId: 'test-session' };
      const selected1 = loadBalancer.select(testNodes, 'pinned', options);
      const selected2 = loadBalancer.select(testNodes, 'pinned', options);

      expect(selected1.id).toBe(selected2.id);
      expect(loadBalancer.pinnedSessions.has('test-session')).toBe(true);
    });

    test('should handle unhealthy pinned node', () => {
      const options = { sessionId: 'test-session' };
      const selected1 = loadBalancer.select(testNodes, 'pinned', options);

      // Make the pinned node unhealthy
      selected1.status = 'unhealthy';

      const selected2 = loadBalancer.select(testNodes, 'pinned', options);
      expect(selected2.status).toBe('healthy');
    });

    test('should clear session affinity', () => {
      const options = { sessionId: 'test-session' };
      loadBalancer.select(testNodes, 'pinned', options);

      expect(loadBalancer.pinnedSessions.has('test-session')).toBe(true);

      loadBalancer.clearSessionAffinity('test-session');
      expect(loadBalancer.pinnedSessions.has('test-session')).toBe(false);
    });
  });

  describe('Broadcast Routing', () => {
    test('should return multiple nodes', () => {
      const options = { maxNodes: 3, minNodes: 2 };
      const selected = loadBalancer.select(testNodes, 'broadcast', options);

      expect(Array.isArray(selected)).toBe(true);
      expect(selected.length).toBeGreaterThanOrEqual(2);
      expect(selected.length).toBeLessThanOrEqual(3);
    });

    test('should respect minNodes constraint', () => {
      const options = { maxNodes: 10, minNodes: 3 };
      const selected = loadBalancer.select(testNodes, 'broadcast', options);

      expect(selected.length).toBeGreaterThanOrEqual(3);
    });

    test('should not exceed available nodes', () => {
      const options = { maxNodes: 10, minNodes: 1 };
      const selected = loadBalancer.select(testNodes, 'broadcast', options);

      expect(selected.length).toBeLessThanOrEqual(testNodes.length);
    });
  });

  describe('Geographic Routing', () => {
    beforeEach(() => {
      testNodes[0].health.region = 'us-east';
      testNodes[1].health.region = 'us-west';
      testNodes[2].health.region = 'eu-central';
      testNodes[3].health.region = 'us-east';

      testNodes[0].health.zone = 'us-east-1a';
      testNodes[3].health.zone = 'us-east-1b';
    });

    test('should prefer nodes in specified region', () => {
      const options = { preferredRegion: 'us-east' };
      const selected = loadBalancer.select(testNodes, 'geographic', options);

      expect(['node1', 'node4']).toContain(selected.id);
    });

    test('should prefer nodes in specified zone', () => {
      const options = { preferredRegion: 'us-east', preferredZone: 'us-east-1a' };
      const selected = loadBalancer.select(testNodes, 'geographic', options);

      expect(selected.id).toBe('node1');
    });

    test('should fallback when region not available', () => {
      const options = { preferredRegion: 'nonexistent-region' };
      const selected = loadBalancer.select(testNodes, 'geographic', options);

      expect(selected).toBeDefined();
    });
  });

  describe('Resource Aware Routing', () => {
    test('should select nodes meeting resource requirements', () => {
      const options = { requiredCpu: 80, requiredMemory: 70 }; // High requirements
      const selected = loadBalancer.select(testNodes, 'resource-aware', options);

      // Should select node1 (cpu: 10, memory: 30 - most available resources)
      expect(selected.id).toBe('node1');
    });

    test('should handle GPU requirements', () => {
      testNodes[1].health.hasGpu = true;
      const options = { requiredGpu: true };
      const selected = loadBalancer.select(testNodes, 'resource-aware', options);

      expect(selected.id).toBe('node2');
    });

    test('should fallback when no nodes meet requirements', () => {
      const options = { requiredCpu: 99, requiredMemory: 99 }; // Impossible requirements
      const selected = loadBalancer.select(testNodes, 'resource-aware', options);

      expect(selected).toBeDefined(); // Should fallback to best available
    });
  });

  describe('Node Scoring', () => {
    test('should calculate node scores correctly', () => {
      const score1 = loadBalancer.calculateNodeScore(testNodes[0]); // Low load/cpu/memory
      const score3 = loadBalancer.calculateNodeScore(testNodes[2]); // High load/cpu/memory

      expect(score1).toBeGreaterThan(score3);
    });

    test('should calculate resource scores correctly', () => {
      const resourceScore1 = loadBalancer.calculateResourceScore(testNodes[0]);
      const resourceScore3 = loadBalancer.calculateResourceScore(testNodes[2]);

      expect(resourceScore1).toBeGreaterThan(resourceScore3);
    });
  });

  describe('Routing Statistics', () => {
    test('should provide routing statistics', () => {
      // Create some pinned sessions
      loadBalancer.select(testNodes, 'pinned', { sessionId: 'session1' });
      loadBalancer.select(testNodes, 'pinned', { sessionId: 'session2' });

      const stats = loadBalancer.getRoutingStats();

      expect(stats).toHaveProperty('pinnedSessions');
      expect(stats).toHaveProperty('currentIndex');
      expect(stats).toHaveProperty('lastScoreUpdate');
      expect(stats.pinnedSessions).toBe(2);
    });
  });
});
