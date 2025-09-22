/**
 * Dynamic Topology Optimization System for Advanced Swarm Orchestration
 * Implements adaptive topology switching, pattern recognition, and performance optimization
 * 
 * Topology patterns:
 * - Hierarchical: Tree-like structure with clear command chains
 * - Mesh: Fully connected network for maximum communication
 * - Ring: Circular communication pattern for sequential processing
 * - Star: Central coordinator with spoke agents
 * - Hybrid: Adaptive combination of multiple patterns
 */

const EventEmitter = require('events');
const Redis = require('ioredis');

class TopologyOptimizer extends EventEmitter {
  constructor(options = {}) {
    super();
    
    this.redis = new Redis({
      host: process.env.REDIS_HOST || 'redis-cluster-0.redis-cluster-service.ollamamax-redis',
      port: process.env.REDIS_PORT || 6379,
      password: process.env.REDIS_PASSWORD || 'ollama_redis_pass',
      retryDelayOnFailure: 1000,
      maxRetriesPerRequest: 3
    });

    // Topology optimization configuration
    this.config = {
      optimizationInterval: options.optimizationInterval || 60000, // 1 minute
      performanceWindowSize: options.performanceWindowSize || 100,
      topologyChangeThreshold: options.topologyChangeThreshold || 0.15,
      adaptationRate: options.adaptationRate || 0.1,
      maxTopologyChanges: options.maxTopologyChanges || 5,
      ...options
    };

    // Topology patterns with their characteristics
    this.topologyPatterns = {
      hierarchical: {
        name: 'Hierarchical Tree',
        description: 'Tree-like structure with clear command chains',
        characteristics: {
          communicationEfficiency: 0.7,
          scalability: 0.9,
          faultTolerance: 0.6,
          coordinationOverhead: 0.4,
          taskDelegation: 0.95,
          loadDistribution: 0.7
        },
        optimalFor: ['complex_planning', 'multi_phase_tasks', 'clear_dependencies'],
        constraints: { minAgents: 3, maxLevels: 5, spanOfControl: 7 }
      },
      mesh: {
        name: 'Full Mesh Network',
        description: 'Fully connected network for maximum communication',
        characteristics: {
          communicationEfficiency: 0.95,
          scalability: 0.5,
          faultTolerance: 0.95,
          coordinationOverhead: 0.8,
          taskDelegation: 0.6,
          loadDistribution: 0.9
        },
        optimalFor: ['collaborative_tasks', 'consensus_required', 'high_reliability'],
        constraints: { minAgents: 2, maxAgents: 15, communicationCost: 'high' }
      },
      ring: {
        name: 'Ring Topology',
        description: 'Circular communication pattern for sequential processing',
        characteristics: {
          communicationEfficiency: 0.8,
          scalability: 0.8,
          faultTolerance: 0.7,
          coordinationOverhead: 0.3,
          taskDelegation: 0.8,
          loadDistribution: 0.6
        },
        optimalFor: ['sequential_processing', 'pipeline_tasks', 'ordered_execution'],
        constraints: { minAgents: 3, maxAgents: 50, messageDelay: 'moderate' }
      },
      star: {
        name: 'Star Topology',
        description: 'Central coordinator with spoke agents',
        characteristics: {
          communicationEfficiency: 0.9,
          scalability: 0.7,
          faultTolerance: 0.5,
          coordinationOverhead: 0.5,
          taskDelegation: 0.9,
          loadDistribution: 0.8
        },
        optimalFor: ['centralized_control', 'fast_coordination', 'simple_tasks'],
        constraints: { minAgents: 2, maxAgents: 20, centralBottleneck: true }
      },
      hybrid: {
        name: 'Adaptive Hybrid',
        description: 'Dynamic combination of multiple patterns',
        characteristics: {
          communicationEfficiency: 0.85,
          scalability: 0.85,
          faultTolerance: 0.8,
          coordinationOverhead: 0.6,
          taskDelegation: 0.85,
          loadDistribution: 0.85
        },
        optimalFor: ['complex_scenarios', 'changing_requirements', 'mixed_workloads'],
        constraints: { minAgents: 5, adaptationComplexity: 'high' }
      }
    };

    // Current topology state
    this.currentTopology = {
      pattern: 'star', // Default pattern
      configuration: {},
      agents: new Map(),
      connections: new Map(),
      performance: {
        communicationLatency: 0,
        throughput: 0,
        reliability: 0,
        efficiency: 0
      },
      lastOptimized: Date.now()
    };

    // Performance tracking
    this.performanceHistory = [];
    this.topologyHistory = [];
    this.optimizationMetrics = {
      totalOptimizations: 0,
      successfulChanges: 0,
      performanceImprovements: 0,
      averageOptimizationTime: 0
    };

    this.initializeOptimizer();
  }

  async initializeOptimizer() {
    try {
      // Load historical topology data
      const historicalData = await this.redis.get('swarm:topology:history');
      if (historicalData) {
        const data = JSON.parse(historicalData);
        this.topologyHistory = data.topologyHistory || [];
        this.optimizationMetrics = data.optimizationMetrics || this.optimizationMetrics;
      }

      // Start optimization loop
      this.optimizationTimer = setInterval(() => {
        this.optimizeTopology().catch(console.error);
      }, this.config.optimizationInterval);

      console.log('Topology optimizer initialized successfully');
      this.emit('initialized', {
        patterns: Object.keys(this.topologyPatterns).length,
        currentPattern: this.currentTopology.pattern,
        optimizationInterval: this.config.optimizationInterval
      });
    } catch (error) {
      console.error('Failed to initialize topology optimizer:', error);
      throw error;
    }
  }

  /**
   * Main topology optimization algorithm
   */
  async optimizeTopology() {
    const startTime = Date.now();
    
    try {
      // Get current swarm state
      const swarmState = await this.getCurrentSwarmState();
      if (!swarmState || swarmState.agents.length < 2) {
        return; // Need at least 2 agents for topology optimization
      }

      // Analyze current performance
      const currentPerformance = await this.analyzeCurrentPerformance(swarmState);
      this.performanceHistory.push({
        timestamp: Date.now(),
        pattern: this.currentTopology.pattern,
        ...currentPerformance
      });

      // Keep performance history within window size
      if (this.performanceHistory.length > this.config.performanceWindowSize) {
        this.performanceHistory.shift();
      }

      // Evaluate potential topology patterns
      const patternEvaluations = await this.evaluateAllPatterns(swarmState);
      
      // Select optimal pattern
      const optimalPattern = this.selectOptimalPattern(patternEvaluations, currentPerformance);
      
      // Check if topology change is beneficial
      if (this.shouldChangeTopology(optimalPattern, currentPerformance)) {
        await this.transitionToTopology(optimalPattern.pattern, swarmState);
        this.optimizationMetrics.successfulChanges++;
        
        if (optimalPattern.expectedPerformance > currentPerformance.overallScore) {
          this.optimizationMetrics.performanceImprovements++;
        }
      }

      // Update metrics
      this.optimizationMetrics.totalOptimizations++;
      const optimizationTime = Date.now() - startTime;
      this.optimizationMetrics.averageOptimizationTime = 
        (this.optimizationMetrics.averageOptimizationTime * (this.optimizationMetrics.totalOptimizations - 1) + optimizationTime) / 
        this.optimizationMetrics.totalOptimizations;

      // Store optimization results
      await this.storeOptimizationResults({
        timestamp: Date.now(),
        currentPattern: this.currentTopology.pattern,
        evaluations: patternEvaluations,
        selectedPattern: optimalPattern,
        performanceGain: optimalPattern.expectedPerformance - currentPerformance.overallScore,
        optimizationTime
      });

      this.emit('optimization_complete', {
        pattern: this.currentTopology.pattern,
        performance: currentPerformance,
        optimizationTime
      });

    } catch (error) {
      console.error('Topology optimization failed:', error);
      this.emit('optimization_error', error);
    }
  }

  async getCurrentSwarmState() {
    try {
      // Get active agents from Redis
      const agentKeys = await this.redis.keys('agent:*:status');
      const agents = [];
      
      for (const key of agentKeys) {
        const agentData = await this.redis.get(key);
        if (agentData) {
          const agent = JSON.parse(agentData);
          if (agent.status === 'active') {
            agents.push(agent);
          }
        }
      }

      // Get current tasks
      const taskKeys = await this.redis.keys('task:*:status');
      const tasks = [];
      
      for (const key of taskKeys) {
        const taskData = await this.redis.get(key);
        if (taskData) {
          const task = JSON.parse(taskData);
          if (['pending', 'in_progress'].includes(task.status)) {
            tasks.push(task);
          }
        }
      }

      // Get communication patterns
      const communicationData = await this.redis.get('swarm:communication:patterns');
      const communicationPatterns = communicationData ? JSON.parse(communicationData) : {};

      return {
        agents,
        tasks,
        communicationPatterns,
        timestamp: Date.now(),
        currentTopology: this.currentTopology.pattern
      };
    } catch (error) {
      console.error('Failed to get swarm state:', error);
      return null;
    }
  }

  async analyzeCurrentPerformance(swarmState) {
    const analysis = {
      communicationLatency: 0,
      throughput: 0,
      reliability: 0,
      efficiency: 0,
      overallScore: 0
    };

    try {
      // Analyze communication latency
      analysis.communicationLatency = await this.calculateCommunicationLatency(swarmState);
      
      // Analyze throughput
      analysis.throughput = await this.calculateThroughput(swarmState);
      
      // Analyze reliability
      analysis.reliability = await this.calculateReliability(swarmState);
      
      // Analyze efficiency
      analysis.efficiency = await this.calculateEfficiency(swarmState);
      
      // Calculate overall performance score
      analysis.overallScore = (
        (1 - analysis.communicationLatency) * 0.25 + // Lower latency is better
        analysis.throughput * 0.3 +
        analysis.reliability * 0.25 +
        analysis.efficiency * 0.2
      );

      return analysis;
    } catch (error) {
      console.error('Performance analysis failed:', error);
      return analysis;
    }
  }

  async calculateCommunicationLatency(swarmState) {
    // Simulate communication latency based on topology pattern and agent count
    const baseLatency = 50; // Base latency in ms
    const agentCount = swarmState.agents.length;
    
    const topologyMultipliers = {
      hierarchical: 1.2 + Math.log2(agentCount) * 0.1,
      mesh: 1.0 + agentCount * 0.02,
      ring: 1.1 + agentCount * 0.05,
      star: 1.0 + agentCount * 0.01,
      hybrid: 1.1 + agentCount * 0.015
    };

    const multiplier = topologyMultipliers[this.currentTopology.pattern] || 1.0;
    const latency = baseLatency * multiplier;
    
    // Normalize to 0-1 scale (higher values = worse performance)
    return Math.min(1.0, latency / 500);
  }

  async calculateThroughput(swarmState) {
    const completedTasks = swarmState.tasks.filter(task => task.status === 'completed').length;
    const totalTasks = swarmState.tasks.length || 1;
    const agentUtilization = swarmState.agents.filter(agent => agent.status === 'busy').length / swarmState.agents.length;
    
    // Topology efficiency factors
    const topologyEfficiency = this.topologyPatterns[this.currentTopology.pattern]?.characteristics.taskDelegation || 0.5;
    
    const throughput = (completedTasks / totalTasks) * agentUtilization * topologyEfficiency;
    
    return Math.min(1.0, throughput);
  }

  async calculateReliability(swarmState) {
    const activeAgents = swarmState.agents.filter(agent => agent.status === 'active').length;
    const totalAgents = swarmState.agents.length || 1;
    const agentReliability = activeAgents / totalAgents;
    
    // Topology fault tolerance factor
    const faultTolerance = this.topologyPatterns[this.currentTopology.pattern]?.characteristics.faultTolerance || 0.5;
    
    return agentReliability * faultTolerance;
  }

  async calculateEfficiency(swarmState) {
    // Resource utilization efficiency
    const avgCpuUsage = swarmState.agents.reduce((sum, agent) => sum + (agent.resources?.cpu || 0), 0) / swarmState.agents.length;
    const avgMemoryUsage = swarmState.agents.reduce((sum, agent) => sum + (agent.resources?.memory || 0), 0) / swarmState.agents.length;
    
    const resourceEfficiency = 1 - Math.abs((avgCpuUsage + avgMemoryUsage) / 2 - 0.7); // Optimal around 70%
    
    // Communication overhead factor
    const overhead = this.topologyPatterns[this.currentTopology.pattern]?.characteristics.coordinationOverhead || 0.5;
    const communicationEfficiency = 1 - overhead;
    
    return (resourceEfficiency * 0.6 + communicationEfficiency * 0.4);
  }

  async evaluateAllPatterns(swarmState) {
    const evaluations = {};
    
    for (const [patternName, patternConfig] of Object.entries(this.topologyPatterns)) {
      // Check if pattern is feasible for current swarm
      if (this.isPatternFeasible(patternName, swarmState)) {
        evaluations[patternName] = await this.evaluatePattern(patternName, swarmState);
      }
    }
    
    return evaluations;
  }

  isPatternFeasible(patternName, swarmState) {
    const pattern = this.topologyPatterns[patternName];
    const agentCount = swarmState.agents.length;
    
    // Check minimum agent requirements
    if (pattern.constraints.minAgents && agentCount < pattern.constraints.minAgents) {
      return false;
    }
    
    // Check maximum agent limits
    if (pattern.constraints.maxAgents && agentCount > pattern.constraints.maxAgents) {
      return false;
    }
    
    // Additional feasibility checks based on task types
    const taskTypes = swarmState.tasks.map(task => task.type || 'general');
    const uniqueTaskTypes = [...new Set(taskTypes)];
    
    // Mesh topology might not be efficient for large numbers of agents
    if (patternName === 'mesh' && agentCount > 10) {
      return false;
    }
    
    return true;
  }

  async evaluatePattern(patternName, swarmState) {
    const pattern = this.topologyPatterns[patternName];
    const agentCount = swarmState.agents.length;
    const taskCount = swarmState.tasks.length;
    
    // Predict performance metrics for this pattern
    const predictedPerformance = {
      communicationLatency: await this.predictCommunicationLatency(patternName, swarmState),
      throughput: await this.predictThroughput(patternName, swarmState),
      reliability: await this.predictReliability(patternName, swarmState),
      efficiency: await this.predictEfficiency(patternName, swarmState)
    };
    
    // Calculate expected overall performance
    const expectedPerformance = (
      (1 - predictedPerformance.communicationLatency) * 0.25 +
      predictedPerformance.throughput * 0.3 +
      predictedPerformance.reliability * 0.25 +
      predictedPerformance.efficiency * 0.2
    );
    
    // Calculate transition cost
    const transitionCost = this.calculateTransitionCost(patternName);
    
    // Calculate pattern suitability for current workload
    const workloadSuitability = this.calculateWorkloadSuitability(patternName, swarmState);
    
    return {
      pattern: patternName,
      expectedPerformance,
      transitionCost,
      workloadSuitability,
      predictedMetrics: predictedPerformance,
      feasible: this.isPatternFeasible(patternName, swarmState),
      score: expectedPerformance * workloadSuitability - transitionCost * 0.1
    };
  }

  async predictCommunicationLatency(patternName, swarmState) {
    const characteristics = this.topologyPatterns[patternName].characteristics;
    const agentCount = swarmState.agents.length;
    
    // Base latency calculation
    const baseLatency = 1 - characteristics.communicationEfficiency;
    const scalingFactor = Math.log(agentCount) / 10; // Logarithmic scaling
    
    return Math.min(1.0, baseLatency + scalingFactor);
  }

  async predictThroughput(patternName, swarmState) {
    const characteristics = this.topologyPatterns[patternName].characteristics;
    const agentCount = swarmState.agents.length;
    const taskCount = swarmState.tasks.length;
    
    // Throughput based on delegation efficiency and load distribution
    const baseThroughput = (characteristics.taskDelegation + characteristics.loadDistribution) / 2;
    
    // Scaling efficiency (some patterns scale better than others)
    const scalingEfficiency = characteristics.scalability * Math.min(1.0, 10 / agentCount);
    
    return Math.min(1.0, baseThroughput * scalingEfficiency);
  }

  async predictReliability(patternName, swarmState) {
    const characteristics = this.topologyPatterns[patternName].characteristics;
    const agentCount = swarmState.agents.length;
    
    // Base reliability from pattern characteristics
    const baseReliability = characteristics.faultTolerance;
    
    // Redundancy factor (more agents can provide redundancy)
    const redundancyFactor = Math.min(0.2, agentCount / 50);
    
    return Math.min(1.0, baseReliability + redundancyFactor);
  }

  async predictEfficiency(patternName, swarmState) {
    const characteristics = this.topologyPatterns[patternName].characteristics;
    
    // Efficiency based on coordination overhead
    const coordinationEfficiency = 1 - characteristics.coordinationOverhead;
    const loadDistributionEfficiency = characteristics.loadDistribution;
    
    return (coordinationEfficiency + loadDistributionEfficiency) / 2;
  }

  calculateTransitionCost(targetPattern) {
    if (targetPattern === this.currentTopology.pattern) {
      return 0; // No transition needed
    }
    
    // Base transition costs (0-1 scale)
    const transitionCosts = {
      'hierarchical': { 'mesh': 0.8, 'ring': 0.6, 'star': 0.4, 'hybrid': 0.7 },
      'mesh': { 'hierarchical': 0.8, 'ring': 0.9, 'star': 0.7, 'hybrid': 0.6 },
      'ring': { 'hierarchical': 0.6, 'mesh': 0.9, 'star': 0.5, 'hybrid': 0.7 },
      'star': { 'hierarchical': 0.4, 'mesh': 0.7, 'ring': 0.5, 'hybrid': 0.5 },
      'hybrid': { 'hierarchical': 0.7, 'mesh': 0.6, 'ring': 0.7, 'star': 0.5 }
    };
    
    return transitionCosts[this.currentTopology.pattern]?.[targetPattern] || 0.5;
  }

  calculateWorkloadSuitability(patternName, swarmState) {
    const pattern = this.topologyPatterns[patternName];
    const optimalFor = pattern.optimalFor;
    
    // Analyze task types and characteristics
    const taskTypes = swarmState.tasks.map(task => task.type || 'general');
    const taskCharacteristics = this.analyzeTaskCharacteristics(swarmState.tasks);
    
    let suitabilityScore = 0.5; // Base suitability
    
    // Check for optimal workload patterns
    if (optimalFor.includes('complex_planning') && taskCharacteristics.complexity > 0.7) {
      suitabilityScore += 0.3;
    }
    
    if (optimalFor.includes('collaborative_tasks') && taskCharacteristics.collaboration > 0.6) {
      suitabilityScore += 0.3;
    }
    
    if (optimalFor.includes('sequential_processing') && taskCharacteristics.sequential > 0.7) {
      suitabilityScore += 0.3;
    }
    
    if (optimalFor.includes('centralized_control') && taskCharacteristics.centralization > 0.6) {
      suitabilityScore += 0.3;
    }
    
    return Math.min(1.0, suitabilityScore);
  }

  analyzeTaskCharacteristics(tasks) {
    if (!tasks || tasks.length === 0) {
      return { complexity: 0.5, collaboration: 0.5, sequential: 0.5, centralization: 0.5 };
    }
    
    const characteristics = {
      complexity: 0,
      collaboration: 0,
      sequential: 0,
      centralization: 0
    };
    
    for (const task of tasks) {
      // Analyze task complexity
      const steps = task.steps?.length || 1;
      const dependencies = task.dependencies?.length || 0;
      characteristics.complexity += Math.min(1.0, (steps + dependencies) / 10);
      
      // Analyze collaboration requirements
      const requiredAgents = task.requiredAgents || 1;
      characteristics.collaboration += Math.min(1.0, requiredAgents / 5);
      
      // Analyze sequential nature
      const hasOrder = task.sequential || (dependencies > 0);
      characteristics.sequential += hasOrder ? 1 : 0;
      
      // Analyze centralization needs
      const needsCoordination = task.coordination || (requiredAgents > 1);
      characteristics.centralization += needsCoordination ? 1 : 0;
    }
    
    // Normalize by task count
    const taskCount = tasks.length;
    characteristics.complexity /= taskCount;
    characteristics.collaboration /= taskCount;
    characteristics.sequential /= taskCount;
    characteristics.centralization /= taskCount;
    
    return characteristics;
  }

  selectOptimalPattern(evaluations, currentPerformance) {
    let bestPattern = null;
    let bestScore = -Infinity;
    
    for (const [patternName, evaluation] of Object.entries(evaluations)) {
      if (evaluation.feasible && evaluation.score > bestScore) {
        bestScore = evaluation.score;
        bestPattern = evaluation;
      }
    }
    
    // If no better pattern found, stick with current
    if (!bestPattern || bestScore <= currentPerformance.overallScore + this.config.topologyChangeThreshold) {
      bestPattern = {
        pattern: this.currentTopology.pattern,
        expectedPerformance: currentPerformance.overallScore,
        transitionCost: 0,
        workloadSuitability: 1,
        score: currentPerformance.overallScore
      };
    }
    
    return bestPattern;
  }

  shouldChangeTopology(optimalPattern, currentPerformance) {
    // Don't change if it's the same pattern
    if (optimalPattern.pattern === this.currentTopology.pattern) {
      return false;
    }
    
    // Check if performance improvement justifies transition cost
    const performanceGain = optimalPattern.expectedPerformance - currentPerformance.overallScore;
    const minGainRequired = this.config.topologyChangeThreshold + optimalPattern.transitionCost;
    
    if (performanceGain < minGainRequired) {
      return false;
    }
    
    // Check if we've made too many recent changes
    const recentChanges = this.topologyHistory
      .filter(change => Date.now() - change.timestamp < 300000) // Last 5 minutes
      .length;
    
    if (recentChanges >= this.config.maxTopologyChanges) {
      return false;
    }
    
    return true;
  }

  async transitionToTopology(targetPattern, swarmState) {
    const startTime = Date.now();
    
    try {
      console.log(`Transitioning from ${this.currentTopology.pattern} to ${targetPattern}`);
      
      // Store current topology in history
      this.topologyHistory.push({
        timestamp: Date.now(),
        fromPattern: this.currentTopology.pattern,
        toPattern: targetPattern,
        reason: 'optimization',
        agentCount: swarmState.agents.length,
        taskCount: swarmState.tasks.length
      });
      
      // Update current topology
      const previousPattern = this.currentTopology.pattern;
      this.currentTopology.pattern = targetPattern;
      this.currentTopology.lastOptimized = Date.now();
      
      // Configure new topology
      await this.configureTopology(targetPattern, swarmState);
      
      // Update agent connections
      await this.updateAgentConnections(swarmState);
      
      // Store topology state
      await this.redis.setex('swarm:topology:current', 3600, JSON.stringify({
        pattern: targetPattern,
        timestamp: Date.now(),
        configuration: this.currentTopology.configuration,
        transitionTime: Date.now() - startTime
      }));
      
      this.emit('topology_changed', {
        from: previousPattern,
        to: targetPattern,
        transitionTime: Date.now() - startTime,
        agentCount: swarmState.agents.length
      });
      
      console.log(`Topology transition completed in ${Date.now() - startTime}ms`);
      
    } catch (error) {
      console.error('Topology transition failed:', error);
      throw error;
    }
  }

  async configureTopology(pattern, swarmState) {
    const agents = swarmState.agents;
    const config = {};
    
    switch (pattern) {
      case 'hierarchical':
        config.hierarchy = await this.buildHierarchy(agents);
        config.maxLevels = Math.min(5, Math.ceil(Math.log2(agents.length)));
        config.spanOfControl = Math.min(7, Math.ceil(agents.length / 3));
        break;
        
      case 'mesh':
        config.connections = await this.buildMeshConnections(agents);
        config.communicationProtocol = 'multicast';
        config.consensusThreshold = Math.ceil(agents.length / 2) + 1;
        break;
        
      case 'ring':
        config.ringOrder = await this.buildRingOrder(agents);
        config.direction = 'bidirectional';
        config.tokenPassing = true;
        break;
        
      case 'star':
        config.coordinator = await this.selectCoordinator(agents);
        config.spokes = agents.filter(agent => agent.id !== config.coordinator.id);
        config.backupCoordinator = await this.selectBackupCoordinator(agents);
        break;
        
      case 'hybrid':
        config.subTopologies = await this.buildHybridTopology(agents);
        config.bridgeAgents = await this.selectBridgeAgents(agents);
        config.adaptationRules = this.getHybridAdaptationRules();
        break;
    }
    
    this.currentTopology.configuration = config;
    return config;
  }

  async buildHierarchy(agents) {
    const hierarchy = { levels: [], totalLevels: 0 };
    const remainingAgents = [...agents];
    let currentLevel = 0;
    
    // Place root coordinator
    const root = this.selectBestCoordinator(remainingAgents);
    hierarchy.levels[0] = [root];
    remainingAgents.splice(remainingAgents.indexOf(root), 1);
    
    // Build subsequent levels
    while (remainingAgents.length > 0 && currentLevel < 4) {
      const parentLevel = hierarchy.levels[currentLevel];
      const nextLevel = [];
      const agentsPerParent = Math.ceil(remainingAgents.length / parentLevel.length);
      
      for (const parent of parentLevel) {
        const children = remainingAgents.splice(0, Math.min(agentsPerParent, remainingAgents.length));
        children.forEach(child => {
          child.parent = parent.id;
          child.level = currentLevel + 1;
        });
        nextLevel.push(...children);
        
        if (remainingAgents.length === 0) break;
      }
      
      hierarchy.levels[++currentLevel] = nextLevel;
    }
    
    hierarchy.totalLevels = currentLevel + 1;
    return hierarchy;
  }

  async buildMeshConnections(agents) {
    const connections = new Map();
    
    // Create full mesh connections
    for (let i = 0; i < agents.length; i++) {
      const agent = agents[i];
      const agentConnections = [];
      
      for (let j = 0; j < agents.length; j++) {
        if (i !== j) {
          agentConnections.push({
            id: agents[j].id,
            weight: 1.0,
            latency: Math.random() * 50 + 10, // Simulated latency
            bandwidth: 1000 // Mbps
          });
        }
      }
      
      connections.set(agent.id, agentConnections);
    }
    
    return connections;
  }

  async buildRingOrder(agents) {
    // Sort agents by some criteria (e.g., processing power, load)
    const sortedAgents = [...agents].sort((a, b) => {
      const scoreA = (a.capabilities?.processing || 1) * (1 - (a.load || 0));
      const scoreB = (b.capabilities?.processing || 1) * (1 - (b.load || 0));
      return scoreB - scoreA;
    });
    
    return sortedAgents.map((agent, index) => ({
      agent: agent.id,
      position: index,
      next: sortedAgents[(index + 1) % sortedAgents.length].id,
      previous: sortedAgents[(index - 1 + sortedAgents.length) % sortedAgents.length].id
    }));
  }

  selectBestCoordinator(agents) {
    // Select coordinator based on capabilities and current load
    return agents.reduce((best, agent) => {
      const score = (agent.capabilities?.coordination || 1) * (1 - (agent.load || 0));
      const bestScore = (best.capabilities?.coordination || 1) * (1 - (best.load || 0));
      return score > bestScore ? agent : best;
    });
  }

  async selectCoordinator(agents) {
    return this.selectBestCoordinator(agents);
  }

  async selectBackupCoordinator(agents) {
    const coordinator = this.currentTopology.configuration.coordinator;
    const candidates = agents.filter(agent => agent.id !== coordinator.id);
    return this.selectBestCoordinator(candidates);
  }

  async buildHybridTopology(agents) {
    const subTopologies = [];
    const agentGroups = this.groupAgentsByCapability(agents);
    
    for (const [capability, groupAgents] of agentGroups) {
      if (groupAgents.length >= 2) {
        const optimalPattern = this.selectOptimalPatternForGroup(groupAgents);
        subTopologies.push({
          capability,
          agents: groupAgents.map(a => a.id),
          pattern: optimalPattern,
          coordinator: this.selectBestCoordinator(groupAgents).id
        });
      }
    }
    
    return subTopologies;
  }

  groupAgentsByCapability(agents) {
    const groups = new Map();
    
    for (const agent of agents) {
      const primaryCapability = agent.primaryCapability || 'general';
      if (!groups.has(primaryCapability)) {
        groups.set(primaryCapability, []);
      }
      groups.get(primaryCapability).push(agent);
    }
    
    return groups;
  }

  selectOptimalPatternForGroup(groupAgents) {
    const size = groupAgents.length;
    
    if (size <= 3) return 'star';
    if (size <= 6) return 'ring';
    if (size <= 10) return 'hierarchical';
    return 'mesh';
  }

  async selectBridgeAgents(agents) {
    // Select agents with high communication capabilities to bridge sub-topologies
    return agents
      .filter(agent => agent.capabilities?.communication > 0.7)
      .slice(0, Math.min(3, agents.length));
  }

  getHybridAdaptationRules() {
    return {
      loadBalancing: {
        threshold: 0.8,
        action: 'redistribute_tasks'
      },
      faultTolerance: {
        threshold: 0.5,
        action: 'activate_backup_topology'
      },
      performance: {
        threshold: 0.6,
        action: 'optimize_sub_topology'
      }
    };
  }

  async updateAgentConnections(swarmState) {
    const pattern = this.currentTopology.pattern;
    const config = this.currentTopology.configuration;
    
    // Update Redis with new connection information
    for (const agent of swarmState.agents) {
      const connections = this.getAgentConnections(agent.id, pattern, config);
      await this.redis.setex(`agent:${agent.id}:connections`, 3600, JSON.stringify(connections));
    }
    
    // Notify agents of topology change
    this.emit('connections_updated', {
      pattern,
      timestamp: Date.now(),
      agentCount: swarmState.agents.length
    });
  }

  getAgentConnections(agentId, pattern, config) {
    switch (pattern) {
      case 'hierarchical':
        return this.getHierarchicalConnections(agentId, config);
      case 'mesh':
        return this.getMeshConnections(agentId, config);
      case 'ring':
        return this.getRingConnections(agentId, config);
      case 'star':
        return this.getStarConnections(agentId, config);
      case 'hybrid':
        return this.getHybridConnections(agentId, config);
      default:
        return [];
    }
  }

  getHierarchicalConnections(agentId, config) {
    const connections = [];
    
    // Find agent in hierarchy
    for (let level = 0; level < config.hierarchy.levels.length; level++) {
      const levelAgents = config.hierarchy.levels[level];
      const agent = levelAgents.find(a => a.id === agentId);
      
      if (agent) {
        // Connect to parent (if not root)
        if (level > 0 && agent.parent) {
          connections.push({ id: agent.parent, type: 'parent' });
        }
        
        // Connect to children (if any)
        if (level < config.hierarchy.levels.length - 1) {
          const childLevel = config.hierarchy.levels[level + 1];
          const children = childLevel.filter(child => child.parent === agentId);
          children.forEach(child => {
            connections.push({ id: child.id, type: 'child' });
          });
        }
        break;
      }
    }
    
    return connections;
  }

  getMeshConnections(agentId, config) {
    return config.connections.get(agentId) || [];
  }

  getRingConnections(agentId, config) {
    const agentInfo = config.ringOrder.find(info => info.agent === agentId);
    if (!agentInfo) return [];
    
    return [
      { id: agentInfo.next, type: 'next' },
      { id: agentInfo.previous, type: 'previous' }
    ];
  }

  getStarConnections(agentId, config) {
    if (agentId === config.coordinator.id) {
      // Coordinator connects to all spokes
      return config.spokes.map(spoke => ({ id: spoke.id, type: 'spoke' }));
    } else {
      // Spoke connects to coordinator
      return [{ id: config.coordinator.id, type: 'coordinator' }];
    }
  }

  getHybridConnections(agentId, config) {
    const connections = [];
    
    // Find agent's sub-topology
    for (const subTopo of config.subTopologies) {
      if (subTopo.agents.includes(agentId)) {
        // Connect within sub-topology
        const subConnections = subTopo.agents
          .filter(id => id !== agentId)
          .map(id => ({ id, type: 'sub_topology' }));
        connections.push(...subConnections);
        
        // If agent is bridge agent, connect to other sub-topologies
        if (config.bridgeAgents.some(bridge => bridge.id === agentId)) {
          for (const otherSubTopo of config.subTopologies) {
            if (otherSubTopo.capability !== subTopo.capability) {
              connections.push({ id: otherSubTopo.coordinator, type: 'bridge' });
            }
          }
        }
        break;
      }
    }
    
    return connections;
  }

  async storeOptimizationResults(results) {
    const key = `swarm:topology:optimization:${results.timestamp}`;
    
    await this.redis.setex(key, 3600 * 24, JSON.stringify(results));
    
    // Update history
    await this.redis.setex('swarm:topology:history', 3600 * 24 * 7, JSON.stringify({
      topologyHistory: this.topologyHistory.slice(-100), // Keep last 100 changes
      optimizationMetrics: this.optimizationMetrics,
      lastOptimization: results.timestamp
    }));
    
    console.log(`Topology optimization results stored: ${results.currentPattern} -> ${results.selectedPattern.pattern}`);
  }

  // Health check and status methods
  async getStatus() {
    const recentPerformance = this.performanceHistory.slice(-5);
    
    return {
      status: 'operational',
      currentTopology: {
        pattern: this.currentTopology.pattern,
        lastOptimized: this.currentTopology.lastOptimized,
        agentCount: this.currentTopology.agents.size
      },
      performanceHistory: recentPerformance,
      optimizationMetrics: this.optimizationMetrics,
      availablePatterns: Object.keys(this.topologyPatterns),
      optimizationInterval: this.config.optimizationInterval,
      lastOptimization: Math.max(...this.topologyHistory.map(h => h.timestamp), 0)
    };
  }

  async cleanup() {
    if (this.optimizationTimer) {
      clearInterval(this.optimizationTimer);
    }
    
    if (this.redis) {
      await this.redis.quit();
    }
  }
}

module.exports = TopologyOptimizer;