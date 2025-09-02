#!/usr/bin/env node

/**
 * Smart Load Balancer for OllamaMax Distributed System
 * Implements weighted round-robin with health scoring and circuit breakers
 */

const http = require('http');
const { performance } = require('perf_hooks');

class SmartLoadBalancer {
  constructor(nodes = []) {
    this.nodes = new Map();
    this.strategy = 'weighted_round_robin';
    this.roundRobinIndex = 0;
    
    // Circuit breaker configuration
    this.circuitBreaker = {
      failureThreshold: 5,
      recoveryTimeout: 60000,
      halfOpenMaxCalls: 3,
      openCircuits: new Map() // nodeId -> { openedAt, failures, halfOpenCalls }
    };
    
    // Health check configuration
    this.healthCheck = {
      interval: 15000, // 15 seconds
      timeout: 5000,   // 5 seconds
      endpoint: '/api/version'
    };
    
    this.metrics = {
      totalRequests: 0,
      successfulRequests: 0,
      failedRequests: 0,
      avgResponseTime: 0,
      circuitBreakerTrips: 0,
      loadDistribution: new Map()
    };
    
    // Initialize nodes
    this.initializeNodes(nodes);
  }

  /**
   * Initialize node tracking
   */
  initializeNodes(nodeConfigs) {
    const defaultNodes = [
      { id: 'primary', url: 'http://localhost:13000', weight: 1.0, capacity: 100 },
      { id: 'worker-2', url: 'http://localhost:13001', weight: 1.0, capacity: 100 },
      { id: 'worker-3', url: 'http://localhost:13002', weight: 1.0, capacity: 100 }
    ];
    
    const nodes = nodeConfigs.length > 0 ? nodeConfigs : defaultNodes;
    
    for (const nodeConfig of nodes) {
      this.nodes.set(nodeConfig.id, {
        ...nodeConfig,
        health: 1.0,
        responseTime: 0,
        activeConnections: 0,
        requestCount: 0,
        successCount: 0,
        failureCount: 0,
        lastHealthCheck: Date.now(),
        cpuUsage: 0,
        memoryUsage: 0,
        dynamicWeight: nodeConfig.weight
      });
      
      this.metrics.loadDistribution.set(nodeConfig.id, 0);
    }
    
    console.log(`🔧 Initialized load balancer with ${this.nodes.size} nodes`);
    
    // Start health monitoring
    this.startHealthMonitoring();
  }

  /**
   * Select best node using smart algorithm
   */
  async selectNode(requestContext = {}) {
    const availableNodes = this.getAvailableNodes();
    
    if (availableNodes.length === 0) {
      throw new Error('No healthy nodes available');
    }
    
    let selectedNode;
    
    switch (this.strategy) {
      case 'weighted_round_robin':
        selectedNode = this.selectWeightedRoundRobin(availableNodes);
        break;
      case 'least_connections':
        selectedNode = this.selectLeastConnections(availableNodes);
        break;
      case 'fastest_response':
        selectedNode = this.selectFastestResponse(availableNodes);
        break;
      case 'resource_aware':
        selectedNode = this.selectResourceAware(availableNodes);
        break;
      default:
        selectedNode = availableNodes[0];
    }
    
    // Update metrics
    this.metrics.totalRequests++;
    this.metrics.loadDistribution.set(
      selectedNode.id,
      this.metrics.loadDistribution.get(selectedNode.id) + 1
    );
    
    selectedNode.activeConnections++;
    selectedNode.requestCount++;
    
    return selectedNode;
  }

  /**
   * Get nodes not in circuit breaker open state
   */
  getAvailableNodes() {
    const available = [];
    
    for (const [nodeId, node] of this.nodes.entries()) {
      const circuit = this.circuitBreaker.openCircuits.get(nodeId);
      
      if (!circuit) {
        // Normal operation
        available.push(node);
      } else if (circuit.state === 'half-open') {
        // Allow limited requests in half-open state
        if (circuit.halfOpenCalls < this.circuitBreaker.halfOpenMaxCalls) {
          available.push(node);
        }
      } else if (Date.now() - circuit.openedAt > this.circuitBreaker.recoveryTimeout) {
        // Transition to half-open state
        circuit.state = 'half-open';
        circuit.halfOpenCalls = 0;
        available.push(node);
        console.log(`🔄 Circuit breaker for ${nodeId} moved to half-open`);
      }
    }
    
    return available.filter(node => node.health > 0.3); // Minimum health threshold
  }

  /**
   * Weighted round-robin selection
   */
  selectWeightedRoundRobin(nodes) {
    // Calculate total weight
    const totalWeight = nodes.reduce((sum, node) => sum + node.dynamicWeight, 0);
    
    if (totalWeight === 0) {
      return nodes[this.roundRobinIndex % nodes.length];
    }
    
    // Select based on weight
    let weightSum = 0;
    const target = (this.roundRobinIndex % 100) / 100 * totalWeight;
    
    for (const node of nodes) {
      weightSum += node.dynamicWeight;
      if (weightSum >= target) {
        this.roundRobinIndex++;
        return node;
      }
    }
    
    return nodes[nodes.length - 1];
  }

  /**
   * Least connections selection
   */
  selectLeastConnections(nodes) {
    return nodes.reduce((least, current) => 
      current.activeConnections < least.activeConnections ? current : least
    );
  }

  /**
   * Fastest response time selection
   */
  selectFastestResponse(nodes) {
    return nodes.reduce((fastest, current) => 
      current.responseTime < fastest.responseTime ? current : fastest
    );
  }

  /**
   * Resource-aware selection (considers CPU, memory, and connections)
   */
  selectResourceAware(nodes) {
    return nodes.reduce((best, current) => {
      const currentScore = this.calculateResourceScore(current);
      const bestScore = this.calculateResourceScore(best);
      return currentScore > bestScore ? current : best;
    });
  }

  calculateResourceScore(node) {
    const healthFactor = node.health;
    const responseFactor = node.responseTime > 0 ? Math.max(0.1, 1000 / node.responseTime) : 1.0;
    const loadFactor = Math.max(0.1, 1.0 - (node.activeConnections / node.capacity));
    const cpuFactor = Math.max(0.1, 1.0 - (node.cpuUsage / 100));
    const memoryFactor = Math.max(0.1, 1.0 - (node.memoryUsage / 100));
    
    return healthFactor * responseFactor * loadFactor * cpuFactor * memoryFactor;
  }

  /**
   * Handle request completion
   */
  async onRequestComplete(nodeId, success, responseTime) {
    const node = this.nodes.get(nodeId);
    if (!node) return;
    
    node.activeConnections = Math.max(0, node.activeConnections - 1);
    node.responseTime = (node.responseTime + responseTime) / 2; // Moving average
    
    if (success) {
      node.successCount++;
      node.failureCount = Math.max(0, node.failureCount - 1);
      node.health = Math.min(1.0, node.health + 0.05);
      this.metrics.successfulRequests++;
      
      // Handle circuit breaker recovery
      const circuit = this.circuitBreaker.openCircuits.get(nodeId);
      if (circuit && circuit.state === 'half-open') {
        circuit.halfOpenCalls++;
        if (circuit.halfOpenCalls >= this.circuitBreaker.halfOpenMaxCalls) {
          // All half-open requests succeeded, close circuit
          this.circuitBreaker.openCircuits.delete(nodeId);
          console.log(`✅ Circuit breaker closed for ${nodeId}`);
        }
      }
      
    } else {
      node.failureCount++;
      node.health = Math.max(0.0, node.health - 0.1);
      this.metrics.failedRequests++;
      
      // Check if circuit breaker should trip
      if (node.failureCount >= this.circuitBreaker.failureThreshold) {
        this.tripCircuitBreaker(nodeId);
      }
    }
    
    // Update dynamic weight
    this.updateDynamicWeight(node);
    
    // Update average response time
    this.metrics.avgResponseTime = (this.metrics.avgResponseTime + responseTime) / 2;
  }

  tripCircuitBreaker(nodeId) {
    this.circuitBreaker.openCircuits.set(nodeId, {
      state: 'open',
      openedAt: Date.now(),
      failures: this.nodes.get(nodeId).failureCount,
      halfOpenCalls: 0
    });
    
    this.metrics.circuitBreakerTrips++;
    console.warn(`⚠️ Circuit breaker OPENED for ${nodeId}`);
  }

  updateDynamicWeight(node) {
    const baseWeight = node.weight;
    const healthFactor = node.health;
    const performanceFactor = node.responseTime > 0 ? 
      Math.max(0.1, 200 / node.responseTime) : 1.0;
    const loadFactor = Math.max(0.1, 1.0 - (node.activeConnections / node.capacity));
    
    node.dynamicWeight = baseWeight * healthFactor * performanceFactor * loadFactor;
  }

  /**
   * Start health monitoring
   */
  startHealthMonitoring() {
    console.log('🔍 Starting node health monitoring...');
    
    setInterval(async () => {
      await this.performHealthChecks();
    }, this.healthCheck.interval);
  }

  async performHealthChecks() {
    const healthPromises = [];
    
    for (const [nodeId, node] of this.nodes.entries()) {
      healthPromises.push(this.checkNodeHealth(nodeId, node));
    }
    
    await Promise.allSettled(healthPromises);
  }

  async checkNodeHealth(nodeId, node) {
    try {
      const healthStart = performance.now();
      const response = await this.makeHealthRequest(node.url);
      const responseTime = performance.now() - healthStart;
      
      if (response.success) {
        node.health = Math.min(1.0, node.health + 0.1);
        node.responseTime = responseTime;
        node.lastHealthCheck = Date.now();
        
        // Extract system metrics if available
        if (response.data && response.data.system) {
          node.cpuUsage = response.data.system.cpu || node.cpuUsage;
          node.memoryUsage = response.data.system.memory || node.memoryUsage;
        }
        
      } else {
        node.health = Math.max(0.0, node.health - 0.2);
        node.failureCount++;
      }
      
      this.updateDynamicWeight(node);
      
    } catch (error) {
      console.warn(`Health check failed for ${nodeId}:`, error.message);
      node.health = Math.max(0.0, node.health - 0.3);
      this.updateDynamicWeight(node);
    }
  }

  async makeHealthRequest(url) {
    return new Promise((resolve, reject) => {
      const healthUrl = `${url}${this.healthCheck.endpoint}`;
      const req = http.get(healthUrl, { timeout: this.healthCheck.timeout }, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
          resolve({
            success: res.statusCode === 200,
            statusCode: res.statusCode,
            data: data ? JSON.parse(data) : null
          });
        });
      });
      
      req.on('error', reject);
      req.on('timeout', () => {
        req.destroy();
        reject(new Error('Health check timeout'));
      });
    });
  }

  /**
   * Get load balancer statistics
   */
  getLoadBalancerStats() {
    const nodeStats = {};
    
    for (const [nodeId, node] of this.nodes.entries()) {
      const circuit = this.circuitBreaker.openCircuits.get(nodeId);
      
      nodeStats[nodeId] = {
        health: (node.health * 100).toFixed(1) + '%',
        weight: node.dynamicWeight.toFixed(3),
        active_connections: node.activeConnections,
        request_count: node.requestCount,
        success_rate: node.requestCount > 0 ? 
          (node.successCount / node.requestCount * 100).toFixed(1) + '%' : 'N/A',
        avg_response_time: node.responseTime.toFixed(2) + 'ms',
        circuit_breaker: circuit ? circuit.state : 'closed',
        cpu_usage: node.cpuUsage.toFixed(1) + '%',
        memory_usage: node.memoryUsage.toFixed(1) + '%'
      };
    }
    
    const distributionStats = {};
    const totalDistributed = Array.from(this.metrics.loadDistribution.values())
      .reduce((sum, count) => sum + count, 0);
    
    for (const [nodeId, count] of this.metrics.loadDistribution.entries()) {
      distributionStats[nodeId] = totalDistributed > 0 ? 
        (count / totalDistributed * 100).toFixed(1) + '%' : '0%';
    }
    
    return {
      strategy: this.strategy,
      nodes: nodeStats,
      load_distribution: distributionStats,
      circuit_breaker_trips: this.metrics.circuitBreakerTrips,
      overall_success_rate: this.metrics.totalRequests > 0 ?
        (this.metrics.successfulRequests / this.metrics.totalRequests * 100).toFixed(1) + '%' : 'N/A',
      avg_response_time: this.metrics.avgResponseTime.toFixed(2) + 'ms'
    };
  }

  /**
   * Change load balancing strategy
   */
  setStrategy(strategy) {
    const validStrategies = [
      'weighted_round_robin',
      'least_connections', 
      'fastest_response',
      'resource_aware'
    ];
    
    if (validStrategies.includes(strategy)) {
      this.strategy = strategy;
      console.log(`⚖️ Load balancing strategy changed to: ${strategy}`);
    } else {
      throw new Error(`Invalid strategy: ${strategy}. Valid options: ${validStrategies.join(', ')}`);
    }
  }

  /**
   * Generate load balancing optimization report
   */
  generateOptimizationReport() {
    const stats = this.getLoadBalancerStats();
    
    return {
      timestamp: new Date().toISOString(),
      current_performance: stats,
      optimization_impact: {
        request_distribution_efficiency: this.calculateDistributionEfficiency(),
        failover_performance: this.calculateFailoverPerformance(),
        resource_utilization: this.calculateResourceUtilization()
      },
      recommendations: this.generateLoadBalancingRecommendations(stats)
    };
  }

  calculateDistributionEfficiency() {
    const distribution = Array.from(this.metrics.loadDistribution.values());
    if (distribution.length < 2) return 100;
    
    const avg = distribution.reduce((sum, val) => sum + val, 0) / distribution.length;
    const variance = distribution.reduce((sum, val) => sum + Math.pow(val - avg, 2), 0) / distribution.length;
    const stdDev = Math.sqrt(variance);
    
    // Lower standard deviation = better distribution
    const efficiency = Math.max(0, 100 - (stdDev / avg * 100));
    return efficiency.toFixed(1) + '%';
  }

  calculateFailoverPerformance() {
    const healthyNodes = Array.from(this.nodes.values()).filter(n => n.health > 0.7).length;
    const totalNodes = this.nodes.size;
    
    return {
      healthy_nodes: `${healthyNodes}/${totalNodes}`,
      redundancy_level: healthyNodes > 1 ? 'Good' : 'Limited',
      failover_capability: healthyNodes >= 2 ? 'High' : 'Low'
    };
  }

  calculateResourceUtilization() {
    let totalCpu = 0;
    let totalMemory = 0;
    let nodeCount = 0;
    
    for (const node of this.nodes.values()) {
      totalCpu += node.cpuUsage;
      totalMemory += node.memoryUsage;
      nodeCount++;
    }
    
    return {
      avg_cpu_usage: nodeCount > 0 ? (totalCpu / nodeCount).toFixed(1) + '%' : 'N/A',
      avg_memory_usage: nodeCount > 0 ? (totalMemory / nodeCount).toFixed(1) + '%' : 'N/A',
      utilization_efficiency: nodeCount > 0 ? 
        (100 - Math.max(totalCpu, totalMemory) / nodeCount).toFixed(1) + '%' : 'N/A'
    };
  }

  generateLoadBalancingRecommendations(stats) {
    const recommendations = [];
    
    // Check for uneven load distribution
    const distribution = Array.from(this.metrics.loadDistribution.values());
    const maxLoad = Math.max(...distribution);
    const minLoad = Math.min(...distribution);
    
    if (maxLoad > minLoad * 2) {
      recommendations.push({
        priority: 'high',
        category: 'load_distribution',
        issue: 'Uneven load distribution detected',
        action: 'Adjust node weights or switch to least_connections strategy',
        impact: 'Better resource utilization and improved response times'
      });
    }
    
    // Check circuit breaker frequency
    if (this.metrics.circuitBreakerTrips > 0) {
      recommendations.push({
        priority: 'medium',
        category: 'reliability',
        issue: `${this.metrics.circuitBreakerTrips} circuit breaker trips detected`,
        action: 'Investigate node stability and adjust failure thresholds',
        impact: 'Improved system reliability and reduced request failures'
      });
    }
    
    // Check overall success rate
    const successRate = this.metrics.totalRequests > 0 ?
      this.metrics.successfulRequests / this.metrics.totalRequests * 100 : 100;
    
    if (successRate < 95) {
      recommendations.push({
        priority: 'high',
        category: 'reliability',
        issue: `Success rate below target: ${successRate.toFixed(1)}%`,
        action: 'Improve node health monitoring and failover logic',
        impact: 'Higher request success rate and better user experience'
      });
    }
    
    return recommendations;
  }

  /**
   * Cleanup and shutdown
   */
  async cleanup() {
    console.log('🔌 Shutting down load balancer...');
    
    // Clear health check intervals
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
    }
    
    // Reset metrics
    this.metrics = {
      totalRequests: 0,
      successfulRequests: 0,
      failedRequests: 0,
      avgResponseTime: 0,
      circuitBreakerTrips: 0,
      loadDistribution: new Map()
    };
  }
}

module.exports = { SmartLoadBalancer };

// CLI execution for testing
if (require.main === module) {
  const loadBalancer = new SmartLoadBalancer();
  
  (async () => {
    console.log('🧪 Testing Smart Load Balancer...\n');
    
    // Simulate requests
    for (let i = 0; i < 20; i++) {
      try {
        const selectedNode = await loadBalancer.selectNode();
        console.log(`Request ${i + 1} → ${selectedNode.id} (weight: ${selectedNode.dynamicWeight.toFixed(3)})`);
        
        // Simulate request completion
        const responseTime = 50 + Math.random() * 100;
        const success = Math.random() > 0.1; // 90% success rate
        
        await loadBalancer.onRequestComplete(selectedNode.id, success, responseTime);
        
      } catch (error) {
        console.error(`Request ${i + 1} failed:`, error.message);
      }
      
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    
    // Display statistics
    console.log('\n📊 Load Balancer Statistics:');
    const stats = loadBalancer.getLoadBalancerStats();
    console.log(JSON.stringify(stats, null, 2));
    
    // Generate optimization report
    console.log('\n📋 Optimization Report:');
    const report = loadBalancer.generateOptimizationReport();
    console.log(JSON.stringify(report, null, 2));
    
    await loadBalancer.cleanup();
    
  })().catch(console.error);
}