/**
 * Multi-Objective Optimization Engine for Advanced Swarm Orchestration
 * Implements NSGA-II and MOEA/D algorithms for Pareto-optimal agent allocation
 * 
 * Optimization objectives:
 * - Performance: Minimize execution time and maximize throughput
 * - Cost: Minimize resource usage and operational costs
 * - Quality: Maximize success rate and minimize error rate
 * - Reliability: Maximize availability and fault tolerance
 */

const EventEmitter = require('events');
const Redis = require('ioredis');

class MultiObjectiveOptimizer extends EventEmitter {
  constructor(options = {}) {
    super();
    
    this.redis = new Redis({
      host: process.env.REDIS_HOST || 'redis-cluster-0.redis-cluster-service.ollamamax-redis',
      port: process.env.REDIS_PORT || 6379,
      password: process.env.REDIS_PASSWORD || 'ollama_redis_pass',
      retryDelayOnFailure: 1000,
      maxRetriesPerRequest: 3
    });

    // Algorithm configuration
    this.config = {
      populationSize: options.populationSize || 100,
      maxGenerations: options.maxGenerations || 200,
      crossoverRate: options.crossoverRate || 0.9,
      mutationRate: options.mutationRate || 0.1,
      archiveSize: options.archiveSize || 100,
      ...options
    };

    // Optimization objectives with weights and constraints
    this.objectives = {
      performance: {
        name: 'Performance Optimization',
        weight: 0.25,
        minimize: true, // Minimize execution time
        metrics: ['avg_execution_time', 'throughput', 'queue_length'],
        constraints: { max_execution_time: 30000, min_throughput: 10 }
      },
      cost: {
        name: 'Resource Cost Optimization',
        weight: 0.25,
        minimize: true, // Minimize resource costs
        metrics: ['cpu_usage', 'memory_usage', 'network_io', 'storage_cost'],
        constraints: { max_cpu: 0.8, max_memory: 0.85, max_cost_per_hour: 50 }
      },
      quality: {
        name: 'Quality Maximization',
        weight: 0.3,
        minimize: false, // Maximize quality scores
        metrics: ['success_rate', 'error_rate', 'validation_score'],
        constraints: { min_success_rate: 0.85, max_error_rate: 0.05 }
      },
      reliability: {
        name: 'Reliability & Fault Tolerance',
        weight: 0.2,
        minimize: false, // Maximize reliability
        metrics: ['availability', 'mtbf', 'recovery_time', 'redundancy_factor'],
        constraints: { min_availability: 0.99, max_recovery_time: 5000 }
      }
    };

    // Pareto front archive for non-dominated solutions
    this.paretoArchive = [];
    this.currentGeneration = 0;
    this.convergenceHistory = [];
    
    this.initializeOptimizer();
  }

  async initializeOptimizer() {
    try {
      // Load historical optimization data
      const historicalData = await this.redis.get('swarm:optimization:history');
      if (historicalData) {
        const data = JSON.parse(historicalData);
        this.convergenceHistory = data.convergenceHistory || [];
        this.paretoArchive = data.paretoArchive || [];
      }

      // Initialize population with diverse solutions
      await this.initializePopulation();
      
      console.log('Multi-objective optimizer initialized successfully');
      this.emit('initialized', { 
        populationSize: this.config.populationSize,
        objectives: Object.keys(this.objectives).length,
        archiveSize: this.paretoArchive.length
      });
    } catch (error) {
      console.error('Failed to initialize multi-objective optimizer:', error);
      throw error;
    }
  }

  /**
   * NSGA-II Implementation for Multi-Objective Optimization
   */
  async optimizeWithNSGAII(swarmState, optimizationRequest) {
    const startTime = Date.now();
    
    try {
      // Extract current population or generate initial population
      let population = await this.getCurrentPopulation(swarmState);
      
      // Evolution loop
      for (let generation = 0; generation < this.config.maxGenerations; generation++) {
        this.currentGeneration = generation;
        
        // Evaluate objectives for each solution
        const evaluatedPopulation = await this.evaluatePopulation(population, swarmState);
        
        // Non-dominated sorting
        const fronts = this.nonDominatedSorting(evaluatedPopulation);
        
        // Crowding distance calculation
        fronts.forEach(front => {
          this.calculateCrowdingDistance(front);
        });
        
        // Selection for next generation
        population = this.environmentalSelection(fronts);
        
        // Update Pareto archive
        this.updateParetoArchive(fronts[0]); // First front contains non-dominated solutions
        
        // Genetic operators: crossover and mutation
        const offspring = await this.generateOffspring(population);
        population = population.concat(offspring);
        
        // Track convergence metrics
        const convergenceMetric = this.calculateConvergence(fronts[0]);
        this.convergenceHistory.push({
          generation,
          convergence: convergenceMetric,
          paretoSize: fronts[0].length,
          hypervolume: this.calculateHypervolume(fronts[0])
        });
        
        // Early termination if converged
        if (this.hasConverged(convergenceMetric)) {
          console.log(`NSGA-II converged at generation ${generation}`);
          break;
        }
        
        // Emit progress
        if (generation % 10 === 0) {
          this.emit('progress', {
            generation,
            paretoSize: fronts[0].length,
            convergence: convergenceMetric
          });
        }
      }
      
      // Select best solution from Pareto front
      const bestSolution = this.selectBestSolution(this.paretoArchive, optimizationRequest.preferences);
      
      const optimizationTime = Date.now() - startTime;
      
      // Store results
      await this.storeOptimizationResults({
        algorithm: 'NSGA-II',
        solution: bestSolution,
        paretoFront: this.paretoArchive.slice(0, 20), // Top 20 solutions
        generations: this.currentGeneration,
        convergenceHistory: this.convergenceHistory.slice(-50), // Last 50 generations
        optimizationTime,
        objectives: this.objectives
      });
      
      return bestSolution;
      
    } catch (error) {
      console.error('NSGA-II optimization failed:', error);
      throw error;
    }
  }

  /**
   * MOEA/D Implementation for Decomposition-based Optimization
   */
  async optimizeWithMOEAD(swarmState, optimizationRequest) {
    const startTime = Date.now();
    
    try {
      // Generate weight vectors for decomposition
      const weightVectors = this.generateWeightVectors();
      const neighborhoodSize = Math.floor(this.config.populationSize * 0.1);
      
      // Initialize population and neighborhoods
      let population = await this.getCurrentPopulation(swarmState);
      const neighborhoods = this.calculateNeighborhoods(weightVectors, neighborhoodSize);
      
      // Reference point for Tchebycheff approach
      let referencePoint = await this.calculateReferencePoint(population);
      
      for (let generation = 0; generation < this.config.maxGenerations; generation++) {
        this.currentGeneration = generation;
        
        for (let i = 0; i < population.length; i++) {
          // Select parents from neighborhood
          const parents = this.selectParentsFromNeighborhood(population, neighborhoods[i]);
          
          // Generate offspring
          const offspring = await this.crossoverAndMutation(parents);
          
          // Evaluate offspring
          const evaluatedOffspring = await this.evaluateSolution(offspring, swarmState);
          
          // Update reference point
          referencePoint = this.updateReferencePoint(referencePoint, evaluatedOffspring);
          
          // Update neighbors using Tchebycheff approach
          await this.updateNeighbors(population, evaluatedOffspring, neighborhoods[i], 
                                   weightVectors, referencePoint);
        }
        
        // Update external archive
        this.updateExternalArchive(population);
        
        // Track progress
        if (generation % 20 === 0) {
          const diversity = this.calculateDiversity(population);
          this.emit('progress', {
            generation,
            diversity,
            archiveSize: this.paretoArchive.length,
            referencePoint
          });
        }
      }
      
      // Select best solution
      const bestSolution = this.selectBestSolutionMOEAD(population, optimizationRequest.preferences);
      
      const optimizationTime = Date.now() - startTime;
      
      await this.storeOptimizationResults({
        algorithm: 'MOEA/D',
        solution: bestSolution,
        population: population.slice(0, 20),
        generations: this.currentGeneration,
        optimizationTime,
        diversity: this.calculateDiversity(population),
        objectives: this.objectives
      });
      
      return bestSolution;
      
    } catch (error) {
      console.error('MOEA/D optimization failed:', error);
      throw error;
    }
  }

  /**
   * Solution representation and encoding
   */
  async getCurrentPopulation(swarmState) {
    const population = [];
    
    for (let i = 0; i < this.config.populationSize; i++) {
      const solution = {
        id: `sol_${Date.now()}_${i}`,
        genes: await this.generateRandomSolution(swarmState),
        objectives: {},
        rank: 0,
        crowdingDistance: 0,
        fitness: 0
      };
      
      population.push(solution);
    }
    
    return population;
  }

  async generateRandomSolution(swarmState) {
    const availableAgents = swarmState.agents || [];
    const pendingTasks = swarmState.tasks || [];
    
    return {
      // Agent allocation strategy
      allocation: {
        strategy: this.randomChoice(['round_robin', 'load_balanced', 'capability_matched', 'performance_optimized']),
        maxAgentsPerTask: Math.floor(Math.random() * 5) + 1,
        loadBalanceThreshold: Math.random() * 0.5 + 0.3
      },
      
      // Resource limits
      resources: {
        cpuLimit: Math.random() * 0.8 + 0.2,
        memoryLimit: Math.random() * 0.8 + 0.2,
        networkBandwidth: Math.random() * 1000 + 100,
        storageQuota: Math.random() * 10 + 1
      },
      
      // Scaling parameters
      scaling: {
        scaleUpThreshold: Math.random() * 0.3 + 0.7,
        scaleDownThreshold: Math.random() * 0.4 + 0.3,
        cooldownPeriod: Math.floor(Math.random() * 300) + 60,
        maxInstances: Math.floor(Math.random() * 20) + 5
      },
      
      // Quality parameters
      quality: {
        retryAttempts: Math.floor(Math.random() * 5) + 1,
        timeoutMs: Math.floor(Math.random() * 20000) + 5000,
        validationLevel: this.randomChoice(['basic', 'standard', 'strict']),
        errorThreshold: Math.random() * 0.1 + 0.01
      },
      
      // Topology configuration
      topology: {
        pattern: this.randomChoice(['hierarchical', 'mesh', 'ring', 'star', 'hybrid']),
        communicationProtocol: this.randomChoice(['http', 'websocket', 'grpc', 'message_queue']),
        redundancyLevel: Math.floor(Math.random() * 3) + 1
      }
    };
  }

  async evaluatePopulation(population, swarmState) {
    const evaluatedPopulation = [];
    
    // Evaluate solutions in parallel batches
    const batchSize = 10;
    for (let i = 0; i < population.length; i += batchSize) {
      const batch = population.slice(i, i + batchSize);
      const evaluatedBatch = await Promise.all(
        batch.map(solution => this.evaluateSolution(solution, swarmState))
      );
      evaluatedPopulation.push(...evaluatedBatch);
    }
    
    return evaluatedPopulation;
  }

  async evaluateSolution(solution, swarmState) {
    const objectives = {};
    
    // Performance evaluation
    objectives.performance = await this.evaluatePerformance(solution, swarmState);
    
    // Cost evaluation
    objectives.cost = await this.evaluateCost(solution, swarmState);
    
    // Quality evaluation
    objectives.quality = await this.evaluateQuality(solution, swarmState);
    
    // Reliability evaluation
    objectives.reliability = await this.evaluateReliability(solution, swarmState);
    
    // Check constraints
    const constraintViolations = this.checkConstraints(solution, objectives);
    
    return {
      ...solution,
      objectives,
      constraintViolations,
      feasible: constraintViolations === 0
    };
  }

  async evaluatePerformance(solution, swarmState) {
    // Simulate performance metrics based on solution parameters
    const baseExecutionTime = 1000; // Base execution time in ms
    const allocationEfficiency = this.calculateAllocationEfficiency(solution.genes.allocation);
    const resourceEfficiency = this.calculateResourceEfficiency(solution.genes.resources);
    
    const executionTime = baseExecutionTime * (1 - allocationEfficiency * 0.3) * (1 - resourceEfficiency * 0.2);
    const throughput = 100 / (executionTime / 1000); // Tasks per second
    const queueLength = Math.max(0, (swarmState.tasks?.length || 0) - throughput * 10);
    
    return {
      executionTime,
      throughput,
      queueLength,
      score: 1 / (1 + executionTime / 1000) // Normalized performance score
    };
  }

  async evaluateCost(solution, swarmState) {
    const resourceCosts = {
      cpu: solution.genes.resources.cpuLimit * 0.1, // Cost per CPU unit
      memory: solution.genes.resources.memoryLimit * 0.05, // Cost per memory unit
      network: solution.genes.resources.networkBandwidth * 0.001, // Cost per bandwidth unit
      storage: solution.genes.resources.storageQuota * 0.02 // Cost per storage unit
    };
    
    const totalResourceCost = Object.values(resourceCosts).reduce((sum, cost) => sum + cost, 0);
    const scalingCost = solution.genes.scaling.maxInstances * 0.5; // Cost for scaling capacity
    const redundancyCost = solution.genes.topology.redundancyLevel * 0.3; // Cost for redundancy
    
    const totalCost = totalResourceCost + scalingCost + redundancyCost;
    
    return {
      resourceCost: totalResourceCost,
      scalingCost,
      redundancyCost,
      totalCost,
      score: 1 / (1 + totalCost) // Normalized cost score (lower is better)
    };
  }

  async evaluateQuality(solution, swarmState) {
    const retryPenalty = Math.max(0, solution.genes.quality.retryAttempts - 3) * 0.05;
    const timeoutBonus = Math.min(0.2, solution.genes.quality.timeoutMs / 50000);
    const validationBonus = {
      'basic': 0,
      'standard': 0.1,
      'strict': 0.2
    }[solution.genes.quality.validationLevel];
    
    const baseQuality = 0.8;
    const qualityScore = Math.min(1.0, baseQuality - retryPenalty + timeoutBonus + validationBonus);
    
    const successRate = Math.max(0.7, qualityScore - solution.genes.quality.errorThreshold);
    const errorRate = Math.min(0.3, solution.genes.quality.errorThreshold * 2);
    
    return {
      successRate,
      errorRate,
      validationScore: qualityScore,
      score: qualityScore
    };
  }

  async evaluateReliability(solution, swarmState) {
    const redundancyBonus = solution.genes.topology.redundancyLevel * 0.1;
    const communicationReliability = {
      'http': 0.95,
      'websocket': 0.90,
      'grpc': 0.98,
      'message_queue': 0.99
    }[solution.genes.topology.communicationProtocol];
    
    const availability = Math.min(0.999, 0.95 + redundancyBonus);
    const mtbf = 10000 * (1 + redundancyBonus); // Mean time between failures
    const recoveryTime = Math.max(1000, 5000 * (1 - redundancyBonus));
    
    const reliabilityScore = availability * communicationReliability;
    
    return {
      availability,
      mtbf,
      recoveryTime,
      redundancyFactor: solution.genes.topology.redundancyLevel,
      score: reliabilityScore
    };
  }

  /**
   * Non-dominated sorting for NSGA-II
   */
  nonDominatedSorting(population) {
    const fronts = [[]];
    const dominationCount = new Array(population.length).fill(0);
    const dominatedSolutions = new Array(population.length).fill(null).map(() => []);
    
    // Calculate domination relationships
    for (let i = 0; i < population.length; i++) {
      for (let j = 0; j < population.length; j++) {
        if (i !== j) {
          if (this.dominates(population[i], population[j])) {
            dominatedSolutions[i].push(j);
          } else if (this.dominates(population[j], population[i])) {
            dominationCount[i]++;
          }
        }
      }
      
      if (dominationCount[i] === 0) {
        population[i].rank = 0;
        fronts[0].push(population[i]);
      }
    }
    
    // Build subsequent fronts
    let frontIndex = 0;
    while (fronts[frontIndex].length > 0) {
      const nextFront = [];
      
      for (const solution of fronts[frontIndex]) {
        const solutionIndex = population.indexOf(solution);
        
        for (const dominatedIndex of dominatedSolutions[solutionIndex]) {
          dominationCount[dominatedIndex]--;
          
          if (dominationCount[dominatedIndex] === 0) {
            population[dominatedIndex].rank = frontIndex + 1;
            nextFront.push(population[dominatedIndex]);
          }
        }
      }
      
      frontIndex++;
      fronts.push(nextFront);
    }
    
    return fronts.filter(front => front.length > 0);
  }

  /**
   * Check if solution1 dominates solution2
   */
  dominates(solution1, solution2) {
    let atLeastOneObjectiveBetter = false;
    
    for (const [objName, objConfig] of Object.entries(this.objectives)) {
      const value1 = solution1.objectives[objName].score;
      const value2 = solution2.objectives[objName].score;
      
      if (objConfig.minimize) {
        if (value1 > value2) return false; // solution1 is worse in this objective
        if (value1 < value2) atLeastOneObjectiveBetter = true;
      } else {
        if (value1 < value2) return false; // solution1 is worse in this objective
        if (value1 > value2) atLeastOneObjectiveBetter = true;
      }
    }
    
    return atLeastOneObjectiveBetter;
  }

  /**
   * Calculate crowding distance for diversity preservation
   */
  calculateCrowdingDistance(front) {
    if (front.length <= 2) {
      front.forEach(solution => solution.crowdingDistance = Infinity);
      return;
    }
    
    // Initialize crowding distances
    front.forEach(solution => solution.crowdingDistance = 0);
    
    // Calculate crowding distance for each objective
    for (const [objName, objConfig] of Object.entries(this.objectives)) {
      // Sort by objective value
      front.sort((a, b) => {
        const valueA = a.objectives[objName].score;
        const valueB = b.objectives[objName].score;
        return objConfig.minimize ? valueA - valueB : valueB - valueA;
      });
      
      // Set boundary solutions to infinite distance
      front[0].crowdingDistance = Infinity;
      front[front.length - 1].crowdingDistance = Infinity;
      
      // Calculate distances for intermediate solutions
      const maxValue = front[front.length - 1].objectives[objName].score;
      const minValue = front[0].objectives[objName].score;
      const range = maxValue - minValue;
      
      if (range > 0) {
        for (let i = 1; i < front.length - 1; i++) {
          const distance = (front[i + 1].objectives[objName].score - front[i - 1].objectives[objName].score) / range;
          front[i].crowdingDistance += distance;
        }
      }
    }
  }

  /**
   * Environmental selection for next generation
   */
  environmentalSelection(fronts) {
    const nextPopulation = [];
    let currentSize = 0;
    
    // Add complete fronts
    for (const front of fronts) {
      if (currentSize + front.length <= this.config.populationSize) {
        nextPopulation.push(...front);
        currentSize += front.length;
      } else {
        // Sort remaining front by crowding distance and select best
        front.sort((a, b) => b.crowdingDistance - a.crowdingDistance);
        const remaining = this.config.populationSize - currentSize;
        nextPopulation.push(...front.slice(0, remaining));
        break;
      }
    }
    
    return nextPopulation;
  }

  /**
   * Generate offspring through crossover and mutation
   */
  async generateOffspring(population) {
    const offspring = [];
    
    while (offspring.length < this.config.populationSize) {
      // Tournament selection
      const parent1 = this.tournamentSelection(population);
      const parent2 = this.tournamentSelection(population);
      
      // Crossover
      if (Math.random() < this.config.crossoverRate) {
        const [child1, child2] = await this.crossover(parent1, parent2);
        offspring.push(child1, child2);
      } else {
        offspring.push({ ...parent1 }, { ...parent2 });
      }
    }
    
    // Mutation
    for (const individual of offspring) {
      if (Math.random() < this.config.mutationRate) {
        await this.mutate(individual);
      }
    }
    
    return offspring.slice(0, this.config.populationSize);
  }

  tournamentSelection(population, tournamentSize = 3) {
    const tournament = [];
    
    for (let i = 0; i < tournamentSize; i++) {
      const randomIndex = Math.floor(Math.random() * population.length);
      tournament.push(population[randomIndex]);
    }
    
    // Select best based on rank and crowding distance
    tournament.sort((a, b) => {
      if (a.rank !== b.rank) {
        return a.rank - b.rank; // Lower rank is better
      }
      return b.crowdingDistance - a.crowdingDistance; // Higher crowding distance is better
    });
    
    return tournament[0];
  }

  async crossover(parent1, parent2) {
    const child1 = { ...parent1 };
    const child2 = { ...parent2 };
    
    // Uniform crossover for allocation parameters
    if (Math.random() < 0.5) {
      [child1.genes.allocation, child2.genes.allocation] = [child2.genes.allocation, child1.genes.allocation];
    }
    
    // Blend crossover for continuous parameters
    const alpha = 0.5;
    for (const key of ['cpuLimit', 'memoryLimit', 'networkBandwidth', 'storageQuota']) {
      const value1 = parent1.genes.resources[key];
      const value2 = parent2.genes.resources[key];
      const diff = Math.abs(value1 - value2);
      
      const min = Math.min(value1, value2) - alpha * diff;
      const max = Math.max(value1, value2) + alpha * diff;
      
      child1.genes.resources[key] = Math.random() * (max - min) + min;
      child2.genes.resources[key] = Math.random() * (max - min) + min;
    }
    
    // Generate new IDs
    child1.id = `child_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    child2.id = `child_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    return [child1, child2];
  }

  async mutate(individual) {
    // Mutation probability for each gene
    const mutationProb = 0.1;
    
    // Mutate allocation strategy
    if (Math.random() < mutationProb) {
      individual.genes.allocation.strategy = this.randomChoice(['round_robin', 'load_balanced', 'capability_matched', 'performance_optimized']);
    }
    
    // Mutate resource limits with Gaussian noise
    for (const key of ['cpuLimit', 'memoryLimit']) {
      if (Math.random() < mutationProb) {
        const noise = (Math.random() - 0.5) * 0.2;
        individual.genes.resources[key] = Math.max(0.1, Math.min(1.0, individual.genes.resources[key] + noise));
      }
    }
    
    // Mutate scaling parameters
    if (Math.random() < mutationProb) {
      individual.genes.scaling.maxInstances = Math.max(1, individual.genes.scaling.maxInstances + Math.floor((Math.random() - 0.5) * 6));
    }
    
    // Mutate topology
    if (Math.random() < mutationProb) {
      individual.genes.topology.pattern = this.randomChoice(['hierarchical', 'mesh', 'ring', 'star', 'hybrid']);
    }
  }

  /**
   * Utility methods
   */
  randomChoice(array) {
    return array[Math.floor(Math.random() * array.length)];
  }

  calculateAllocationEfficiency(allocation) {
    const strategyScores = {
      'round_robin': 0.6,
      'load_balanced': 0.8,
      'capability_matched': 0.9,
      'performance_optimized': 0.95
    };
    return strategyScores[allocation.strategy] || 0.5;
  }

  calculateResourceEfficiency(resources) {
    // Efficiency based on resource utilization balance
    const utilization = (resources.cpuLimit + resources.memoryLimit) / 2;
    return Math.max(0, 1 - Math.abs(utilization - 0.7)); // Optimal around 70% utilization
  }

  checkConstraints(solution, objectives) {
    let violations = 0;
    
    for (const [objName, objConfig] of Object.entries(this.objectives)) {
      for (const [constraintName, constraintValue] of Object.entries(objConfig.constraints)) {
        const actualValue = objectives[objName][constraintName.replace('min_', '').replace('max_', '')];
        
        if (constraintName.startsWith('min_') && actualValue < constraintValue) {
          violations++;
        } else if (constraintName.startsWith('max_') && actualValue > constraintValue) {
          violations++;
        }
      }
    }
    
    return violations;
  }

  updateParetoArchive(newFront) {
    // Add new non-dominated solutions to archive
    for (const solution of newFront) {
      let dominated = false;
      const toRemove = [];
      
      // Check against existing archive solutions
      for (let i = 0; i < this.paretoArchive.length; i++) {
        const archiveSolution = this.paretoArchive[i];
        
        if (this.dominates(solution, archiveSolution)) {
          toRemove.push(i);
        } else if (this.dominates(archiveSolution, solution)) {
          dominated = true;
          break;
        }
      }
      
      // Remove dominated solutions from archive
      for (let i = toRemove.length - 1; i >= 0; i--) {
        this.paretoArchive.splice(toRemove[i], 1);
      }
      
      // Add solution if not dominated
      if (!dominated) {
        this.paretoArchive.push({ ...solution });
      }
    }
    
    // Limit archive size
    if (this.paretoArchive.length > this.config.archiveSize) {
      this.paretoArchive.sort((a, b) => b.crowdingDistance - a.crowdingDistance);
      this.paretoArchive = this.paretoArchive.slice(0, this.config.archiveSize);
    }
  }

  selectBestSolution(paretoFront, preferences = {}) {
    if (paretoFront.length === 0) {
      return null;
    }
    
    // Default preferences favor balanced performance
    const defaultPreferences = {
      performance: 0.25,
      cost: 0.25,
      quality: 0.3,
      reliability: 0.2
    };
    
    const actualPreferences = { ...defaultPreferences, ...preferences };
    
    // Calculate weighted scores
    let bestSolution = null;
    let bestScore = -Infinity;
    
    for (const solution of paretoFront) {
      let weightedScore = 0;
      
      for (const [objName, weight] of Object.entries(actualPreferences)) {
        if (this.objectives[objName] && solution.objectives[objName]) {
          weightedScore += weight * solution.objectives[objName].score;
        }
      }
      
      if (weightedScore > bestScore) {
        bestScore = weightedScore;
        bestSolution = solution;
      }
    }
    
    return {
      solution: bestSolution,
      score: bestScore,
      paretoRank: bestSolution?.rank || 0,
      objectives: bestSolution?.objectives || {},
      configuration: bestSolution?.genes || {}
    };
  }

  async storeOptimizationResults(results) {
    const timestamp = new Date().toISOString();
    const key = `swarm:optimization:result:${timestamp}`;
    
    await this.redis.setex(key, 3600 * 24 * 7, JSON.stringify({
      timestamp,
      ...results
    }));
    
    // Update history
    await this.redis.setex('swarm:optimization:history', 3600 * 24 * 30, JSON.stringify({
      convergenceHistory: this.convergenceHistory,
      paretoArchive: this.paretoArchive,
      lastOptimization: timestamp
    }));
    
    console.log(`Optimization results stored: ${results.algorithm}, ${results.generations} generations, ${results.optimizationTime}ms`);
  }

  calculateConvergence(front) {
    if (front.length < 2) return 1.0;
    
    // Calculate average distance between consecutive solutions in objective space
    const distances = [];
    for (let i = 0; i < front.length - 1; i++) {
      let distance = 0;
      for (const objName of Object.keys(this.objectives)) {
        const diff = front[i].objectives[objName].score - front[i + 1].objectives[objName].score;
        distance += diff * diff;
      }
      distances.push(Math.sqrt(distance));
    }
    
    const avgDistance = distances.reduce((sum, d) => sum + d, 0) / distances.length;
    return Math.max(0, 1 - avgDistance); // Higher value means better convergence
  }

  hasConverged(convergenceMetric, threshold = 0.95) {
    if (this.convergenceHistory.length < 10) return false;
    
    const recentHistory = this.convergenceHistory.slice(-10);
    const avgConvergence = recentHistory.reduce((sum, h) => sum + h.convergence, 0) / recentHistory.length;
    
    return avgConvergence >= threshold;
  }

  calculateHypervolume(front, referencePoint = null) {
    if (!referencePoint) {
      referencePoint = {};
      for (const objName of Object.keys(this.objectives)) {
        referencePoint[objName] = 0;
      }
    }
    
    // Simplified hypervolume calculation for 2D case
    if (Object.keys(this.objectives).length === 2) {
      const objNames = Object.keys(this.objectives);
      const points = front.map(sol => [
        sol.objectives[objNames[0]].score,
        sol.objectives[objNames[1]].score
      ]);
      
      points.sort((a, b) => a[0] - b[0]);
      
      let hypervolume = 0;
      let prevY = referencePoint[objNames[1]];
      
      for (const point of points) {
        hypervolume += (point[1] - prevY) * (point[0] - referencePoint[objNames[0]]);
        prevY = Math.max(prevY, point[1]);
      }
      
      return hypervolume;
    }
    
    // For higher dimensions, return normalized volume approximation
    return front.length * Math.pow(0.8, Object.keys(this.objectives).length);
  }

  // Health check and status methods
  async getStatus() {
    return {
      status: 'operational',
      currentGeneration: this.currentGeneration,
      paretoArchiveSize: this.paretoArchive.length,
      objectives: Object.keys(this.objectives),
      convergenceHistory: this.convergenceHistory.slice(-5),
      configuration: {
        populationSize: this.config.populationSize,
        maxGenerations: this.config.maxGenerations,
        crossoverRate: this.config.crossoverRate,
        mutationRate: this.config.mutationRate
      }
    };
  }

  async cleanup() {
    if (this.redis) {
      await this.redis.quit();
    }
  }
}

module.exports = MultiObjectiveOptimizer;