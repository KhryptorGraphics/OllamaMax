/**
 * Advanced Swarm Intelligence Patterns System
 * Implements nature-inspired collective intelligence behaviors including
 * ant colony optimization, particle swarm optimization, bee algorithm,
 * flocking behaviors, and emergent decision-making patterns
 * 
 * Patterns:
 * - Ant Colony Optimization (ACO): Path finding and resource optimization
 * - Particle Swarm Optimization (PSO): Solution space exploration
 * - Artificial Bee Colony (ABC): Foraging and resource allocation
 * - Boids Flocking: Coordinated movement and formation
 * - Firefly Algorithm: Synchronization and optimization
 * - Wolf Pack: Hierarchical hunting strategies
 */

const EventEmitter = require('events');
const Redis = require('ioredis');

class SwarmIntelligencePatterns extends EventEmitter {
  constructor(options = {}) {
    super();
    
    this.redis = new Redis({
      host: process.env.REDIS_HOST || 'redis-cluster-0.redis-cluster-service.ollamamax-redis',
      port: process.env.REDIS_PORT || 6379,
      password: process.env.REDIS_PASSWORD || 'ollama_redis_pass',
      retryDelayOnFailure: 1000,
      maxRetriesPerRequest: 3
    });

    // Swarm intelligence configuration
    this.config = {
      maxAgents: options.maxAgents || 50,
      updateInterval: options.updateInterval || 10000, // 10 seconds
      convergenceThreshold: options.convergenceThreshold || 0.01,
      explorationRate: options.explorationRate || 0.3,
      exploitationRate: options.exploitationRate || 0.7,
      ...options
    };

    // Pattern-specific configurations
    this.patterns = {
      aco: {
        name: 'Ant Colony Optimization',
        enabled: true,
        pheromoneDecay: 0.1,
        pheromoneIntensity: 1.0,
        alpha: 1.0, // Pheromone importance
        beta: 2.0,  // Heuristic importance
        maxAnts: 20,
        pathMemory: 100,
        convergenceThreshold: 0.95
      },
      pso: {
        name: 'Particle Swarm Optimization',
        enabled: true,
        inertia: 0.9,
        personalBest: 2.0,
        globalBest: 2.0,
        maxParticles: 30,
        maxVelocity: 10.0,
        dimensions: 10,
        convergenceThreshold: 0.001
      },
      abc: {
        name: 'Artificial Bee Colony',
        enabled: true,
        scoutBeeRatio: 0.1,
        onlookerBeeRatio: 0.5,
        employedBeeRatio: 0.4,
        maxTrials: 50,
        maxBees: 25,
        danceThreshold: 0.8
      },
      boids: {
        name: 'Boids Flocking',
        enabled: true,
        separationRadius: 2.0,
        alignmentRadius: 5.0,
        cohesionRadius: 8.0,
        separationWeight: 1.5,
        alignmentWeight: 1.0,
        cohesionWeight: 1.0,
        maxSpeed: 5.0,
        maxForce: 0.3
      },
      firefly: {
        name: 'Firefly Algorithm',
        enabled: true,
        attractiveness: 1.0,
        lightAbsorption: 0.1,
        randomization: 0.2,
        maxFireflies: 25,
        brightnessThreshold: 0.9
      },
      wolfpack: {
        name: 'Wolf Pack Algorithm',
        enabled: true,
        alphaRatio: 0.1,   // 10% alphas
        betaRatio: 0.2,    // 20% betas
        omegaRatio: 0.7,   // 70% omegas
        huntingRadius: 10.0,
        packSize: 20,
        cooperationFactor: 0.8
      }
    };

    // Current swarm state
    this.swarmState = {
      agents: new Map(),
      activePatterns: new Set(),
      globalBest: null,
      convergenceMetrics: {},
      environmentMap: new Map(),
      resourceMap: new Map(),
      coordinationState: 'idle'
    };

    // Pattern-specific state
    this.patternState = {
      aco: {
        pheromoneMap: new Map(),
        antColonies: new Map(),
        bestPaths: new Map(),
        pathQuality: new Map()
      },
      pso: {
        particles: new Map(),
        globalBestPosition: null,
        globalBestFitness: -Infinity,
        swarmCenter: null,
        velocities: new Map()
      },
      abc: {
        foodSources: new Map(),
        employedBees: new Map(),
        onlookerBees: new Map(),
        scoutBees: new Map(),
        danceFloor: new Map()
      },
      boids: {
        flock: new Map(),
        neighbors: new Map(),
        formations: new Map(),
        targets: new Map()
      },
      firefly: {
        fireflies: new Map(),
        brightness: new Map(),
        attractions: new Map(),
        synchronization: 0
      },
      wolfpack: {
        alphas: new Map(),
        betas: new Map(),
        omegas: new Map(),
        prey: new Map(),
        hunting: new Map()
      }
    };

    this.initializeSwarmPatterns();
  }

  async initializeSwarmPatterns() {
    try {
      // Load historical swarm data
      await this.loadSwarmHistory();
      
      // Initialize enabled patterns
      await this.initializePatterns();
      
      // Start pattern execution loops
      this.startPatternUpdates();
      
      console.log('Swarm intelligence patterns initialized successfully');
      this.emit('patterns_initialized', {
        enabledPatterns: Array.from(this.swarmState.activePatterns),
        maxAgents: this.config.maxAgents,
        updateInterval: this.config.updateInterval
      });
    } catch (error) {
      console.error('Failed to initialize swarm patterns:', error);
      throw error;
    }
  }

  async loadSwarmHistory() {
    const historyData = await this.redis.get('swarm:patterns:history');
    if (historyData) {
      try {
        const history = JSON.parse(historyData);
        this.swarmState.convergenceMetrics = history.convergenceMetrics || {};
        this.swarmState.globalBest = history.globalBest || null;
      } catch (e) {
        console.warn('Failed to load swarm history:', e);
      }
    }
  }

  async initializePatterns() {
    // Initialize each enabled pattern
    for (const [patternName, config] of Object.entries(this.patterns)) {
      if (config.enabled) {
        await this.initializePattern(patternName);
        this.swarmState.activePatterns.add(patternName);
      }
    }
  }

  async initializePattern(patternName) {
    switch (patternName) {
      case 'aco':
        await this.initializeACO();
        break;
      case 'pso':
        await this.initializePSO();
        break;
      case 'abc':
        await this.initializeABC();
        break;
      case 'boids':
        await this.initializeBoids();
        break;
      case 'firefly':
        await this.initializeFirefly();
        break;
      case 'wolfpack':
        await this.initializeWolfPack();
        break;
      default:
        console.warn(`Unknown pattern: ${patternName}`);
    }
  }

  /**
   * Ant Colony Optimization (ACO) Implementation
   */
  async initializeACO() {
    const acoConfig = this.patterns.aco;
    const acoState = this.patternState.aco;
    
    // Initialize ant colonies
    for (let i = 0; i < acoConfig.maxAnts; i++) {
      const antId = `ant_${i}`;
      const ant = {
        id: antId,
        position: this.generateRandomPosition(),
        path: [],
        pathLength: 0,
        pheromoneDeposit: 0,
        currentTarget: null,
        status: 'exploring'
      };
      
      acoState.antColonies.set(antId, ant);
    }
    
    console.log(`Initialized ACO with ${acoConfig.maxAnts} ants`);
  }

  async executeACO() {
    const acoConfig = this.patterns.aco;
    const acoState = this.patternState.aco;
    
    // Update all ants
    for (const [antId, ant] of acoState.antColonies) {
      await this.updateAnt(ant);
    }
    
    // Update pheromone trails
    await this.updatePheromones();
    
    // Find best paths
    await this.evaluateAntPaths();
    
    // Check convergence
    const convergence = this.checkACOConvergence();
    
    return {
      pattern: 'aco',
      convergence,
      bestPath: this.getBestPath(),
      activeAnts: acoState.antColonies.size,
      pheromoneTrails: acoState.pheromoneMap.size
    };
  }

  async updateAnt(ant) {
    const acoConfig = this.patterns.aco;
    
    // Select next position based on pheromone and heuristic information
    const nextPosition = await this.selectNextPosition(ant);
    
    if (nextPosition) {
      // Move ant to new position
      const previousPosition = ant.position;
      ant.position = nextPosition;
      ant.path.push(nextPosition);
      ant.pathLength += this.calculateDistance(previousPosition, nextPosition);
      
      // Update pheromone deposit based on path quality
      ant.pheromoneDeposit = 1.0 / (1.0 + ant.pathLength);
      
      // Check if reached target or found solution
      if (this.isTargetReached(ant, nextPosition)) {
        ant.status = 'returning';
        await this.depositPheromones(ant);
      }
    } else {
      // No valid move, start over
      ant.position = this.generateRandomPosition();
      ant.path = [];
      ant.pathLength = 0;
      ant.status = 'exploring';
    }
  }

  async selectNextPosition(ant) {
    const acoConfig = this.patterns.aco;
    const currentPos = ant.position;
    const availablePositions = await this.getAvailablePositions(currentPos, ant.path);
    
    if (availablePositions.length === 0) {
      return null;
    }
    
    // Calculate probabilities for each available position
    const probabilities = [];
    let totalProbability = 0;
    
    for (const position of availablePositions) {
      const pheromone = this.getPheromoneLevel(currentPos, position);
      const heuristic = this.calculateHeuristicValue(currentPos, position);
      
      const probability = Math.pow(pheromone, acoConfig.alpha) * Math.pow(heuristic, acoConfig.beta);
      probabilities.push(probability);
      totalProbability += probability;
    }
    
    // Roulette wheel selection
    const random = Math.random() * totalProbability;
    let cumulativeProbability = 0;
    
    for (let i = 0; i < availablePositions.length; i++) {
      cumulativeProbability += probabilities[i];
      if (random <= cumulativeProbability) {
        return availablePositions[i];
      }
    }
    
    return availablePositions[availablePositions.length - 1];
  }

  async updatePheromones() {
    const acoConfig = this.patterns.aco;
    const acoState = this.patternState.aco;
    
    // Evaporate existing pheromones
    for (const [key, value] of acoState.pheromoneMap) {
      const newValue = value * (1 - acoConfig.pheromoneDecay);
      if (newValue < 0.01) {
        acoState.pheromoneMap.delete(key);
      } else {
        acoState.pheromoneMap.set(key, newValue);
      }
    }
    
    // Deposit new pheromones from ants
    for (const [antId, ant] of acoState.antColonies) {
      if (ant.status === 'returning' && ant.path.length > 1) {
        await this.depositPheromones(ant);
      }
    }
  }

  async depositPheromones(ant) {
    const acoConfig = this.patterns.aco;
    const acoState = this.patternState.aco;
    
    // Deposit pheromones along the ant's path
    for (let i = 0; i < ant.path.length - 1; i++) {
      const from = ant.path[i];
      const to = ant.path[i + 1];
      const key = this.getEdgeKey(from, to);
      
      const currentPheromone = acoState.pheromoneMap.get(key) || 0;
      const newPheromone = currentPheromone + ant.pheromoneDeposit * acoConfig.pheromoneIntensity;
      
      acoState.pheromoneMap.set(key, newPheromone);
    }
  }

  /**
   * Particle Swarm Optimization (PSO) Implementation
   */
  async initializePSO() {
    const psoConfig = this.patterns.pso;
    const psoState = this.patternState.pso;
    
    // Initialize particles
    for (let i = 0; i < psoConfig.maxParticles; i++) {
      const particleId = `particle_${i}`;
      const particle = {
        id: particleId,
        position: this.generateRandomVector(psoConfig.dimensions),
        velocity: this.generateRandomVector(psoConfig.dimensions, -1, 1),
        bestPosition: null,
        bestFitness: -Infinity,
        fitness: 0
      };
      
      // Initial fitness evaluation
      particle.fitness = await this.evaluateParticleFitness(particle);
      particle.bestPosition = [...particle.position];
      particle.bestFitness = particle.fitness;
      
      // Update global best
      if (particle.fitness > psoState.globalBestFitness) {
        psoState.globalBestFitness = particle.fitness;
        psoState.globalBestPosition = [...particle.position];
      }
      
      psoState.particles.set(particleId, particle);
    }
    
    console.log(`Initialized PSO with ${psoConfig.maxParticles} particles`);
  }

  async executePSO() {
    const psoConfig = this.patterns.pso;
    const psoState = this.patternState.pso;
    
    let improved = false;
    
    // Update all particles
    for (const [particleId, particle] of psoState.particles) {
      // Update velocity
      await this.updateParticleVelocity(particle);
      
      // Update position
      this.updateParticlePosition(particle);
      
      // Evaluate fitness
      particle.fitness = await this.evaluateParticleFitness(particle);
      
      // Update personal best
      if (particle.fitness > particle.bestFitness) {
        particle.bestFitness = particle.fitness;
        particle.bestPosition = [...particle.position];
        
        // Update global best
        if (particle.fitness > psoState.globalBestFitness) {
          psoState.globalBestFitness = particle.fitness;
          psoState.globalBestPosition = [...particle.position];
          improved = true;
        }
      }
    }
    
    // Update swarm center
    psoState.swarmCenter = this.calculateSwarmCenter();
    
    return {
      pattern: 'pso',
      improved,
      globalBestFitness: psoState.globalBestFitness,
      swarmDiversity: this.calculateSwarmDiversity(),
      convergence: this.checkPSOConvergence()
    };
  }

  async updateParticleVelocity(particle) {
    const psoConfig = this.patterns.pso;
    const psoState = this.patternState.pso;
    
    for (let d = 0; d < particle.position.length; d++) {
      const inertiaComponent = psoConfig.inertia * particle.velocity[d];
      const personalComponent = psoConfig.personalBest * Math.random() * 
        (particle.bestPosition[d] - particle.position[d]);
      const globalComponent = psoConfig.globalBest * Math.random() * 
        (psoState.globalBestPosition[d] - particle.position[d]);
      
      particle.velocity[d] = inertiaComponent + personalComponent + globalComponent;
      
      // Clamp velocity
      particle.velocity[d] = Math.max(-psoConfig.maxVelocity, 
        Math.min(psoConfig.maxVelocity, particle.velocity[d]));
    }
  }

  updateParticlePosition(particle) {
    for (let d = 0; d < particle.position.length; d++) {
      particle.position[d] += particle.velocity[d];
      
      // Boundary handling (reflect)
      if (particle.position[d] < -10) {
        particle.position[d] = -10;
        particle.velocity[d] *= -0.5;
      } else if (particle.position[d] > 10) {
        particle.position[d] = 10;
        particle.velocity[d] *= -0.5;
      }
    }
  }

  /**
   * Artificial Bee Colony (ABC) Implementation
   */
  async initializeABC() {
    const abcConfig = this.patterns.abc;
    const abcState = this.patternState.abc;
    
    const employedCount = Math.floor(abcConfig.maxBees * abcConfig.employedBeeRatio);
    const onlookerCount = Math.floor(abcConfig.maxBees * abcConfig.onlookerBeeRatio);
    const scoutCount = abcConfig.maxBees - employedCount - onlookerCount;
    
    // Initialize food sources
    for (let i = 0; i < employedCount; i++) {
      const sourceId = `source_${i}`;
      const foodSource = {
        id: sourceId,
        position: this.generateRandomVector(10),
        quality: 0,
        trials: 0,
        employedBee: `employed_${i}`
      };
      
      foodSource.quality = await this.evaluateFoodSource(foodSource);
      abcState.foodSources.set(sourceId, foodSource);
    }
    
    // Initialize employed bees
    for (let i = 0; i < employedCount; i++) {
      const beeId = `employed_${i}`;
      const bee = {
        id: beeId,
        type: 'employed',
        assignedSource: `source_${i}`,
        danceValue: 0
      };
      
      abcState.employedBees.set(beeId, bee);
    }
    
    // Initialize onlooker bees
    for (let i = 0; i < onlookerCount; i++) {
      const beeId = `onlooker_${i}`;
      const bee = {
        id: beeId,
        type: 'onlooker',
        selectedSource: null,
        watchingDance: true
      };
      
      abcState.onlookerBees.set(beeId, bee);
    }
    
    // Initialize scout bees
    for (let i = 0; i < scoutCount; i++) {
      const beeId = `scout_${i}`;
      const bee = {
        id: beeId,
        type: 'scout',
        exploring: true
      };
      
      abcState.scoutBees.set(beeId, bee);
    }
    
    console.log(`Initialized ABC with ${abcConfig.maxBees} bees (${employedCount} employed, ${onlookerCount} onlooker, ${scoutCount} scout)`);
  }

  async executeABC() {
    const abcConfig = this.patterns.abc;
    const abcState = this.patternState.abc;
    
    // Employed bee phase
    await this.executeEmployedBeePhase();
    
    // Onlooker bee phase
    await this.executeOnlookerBeePhase();
    
    // Scout bee phase
    await this.executeScoutBeePhase();
    
    // Update dance floor
    await this.updateDanceFloor();
    
    // Find best food source
    const bestSource = this.getBestFoodSource();
    
    return {
      pattern: 'abc',
      bestQuality: bestSource?.quality || 0,
      activeSources: abcState.foodSources.size,
      danceActivity: abcState.danceFloor.size,
      convergence: this.checkABCConvergence()
    };
  }

  async executeEmployedBeePhase() {
    const abcState = this.patternState.abc;
    
    for (const [beeId, bee] of abcState.employedBees) {
      const source = abcState.foodSources.get(bee.assignedSource);
      if (!source) continue;
      
      // Generate new candidate solution
      const candidatePosition = await this.generateCandidatePosition(source.position);
      const candidateSource = {
        id: `candidate_${Date.now()}`,
        position: candidatePosition,
        quality: 0,
        trials: 0
      };
      
      candidateSource.quality = await this.evaluateFoodSource(candidateSource);
      
      // Greedy selection
      if (candidateSource.quality > source.quality) {
        source.position = candidateSource.position;
        source.quality = candidateSource.quality;
        source.trials = 0;
        
        // Calculate dance value
        bee.danceValue = source.quality;
      } else {
        source.trials++;
      }
    }
  }

  async executeOnlookerBeePhase() {
    const abcConfig = this.patterns.abc;
    const abcState = this.patternState.abc;
    
    // Calculate selection probabilities
    const probabilities = this.calculateSourceProbabilities();
    
    for (const [beeId, bee] of abcState.onlookerBees) {
      // Select food source based on dance
      const selectedSourceId = this.selectSourceByProbability(probabilities);
      const source = abcState.foodSources.get(selectedSourceId);
      
      if (source) {
        bee.selectedSource = selectedSourceId;
        
        // Generate new candidate solution
        const candidatePosition = await this.generateCandidatePosition(source.position);
        const candidateSource = {
          id: `candidate_${Date.now()}`,
          position: candidatePosition,
          quality: 0,
          trials: 0
        };
        
        candidateSource.quality = await this.evaluateFoodSource(candidateSource);
        
        // Greedy selection
        if (candidateSource.quality > source.quality) {
          source.position = candidateSource.position;
          source.quality = candidateSource.quality;
          source.trials = 0;
        } else {
          source.trials++;
        }
      }
    }
  }

  async executeScoutBeePhase() {
    const abcConfig = this.patterns.abc;
    const abcState = this.patternState.abc;
    
    // Find exhausted food sources
    const exhaustedSources = [];
    for (const [sourceId, source] of abcState.foodSources) {
      if (source.trials >= abcConfig.maxTrials) {
        exhaustedSources.push(sourceId);
      }
    }
    
    // Replace exhausted sources with new random ones
    for (const sourceId of exhaustedSources) {
      const newSource = {
        id: sourceId,
        position: this.generateRandomVector(10),
        quality: 0,
        trials: 0
      };
      
      newSource.quality = await this.evaluateFoodSource(newSource);
      abcState.foodSources.set(sourceId, newSource);
    }
  }

  /**
   * Boids Flocking Implementation
   */
  async initializeBoids() {
    const boidsConfig = this.patterns.boids;
    const boidsState = this.patternState.boids;
    
    // Initialize boid agents
    for (let i = 0; i < 20; i++) {
      const boidId = `boid_${i}`;
      const boid = {
        id: boidId,
        position: this.generateRandomVector(3), // 3D position
        velocity: this.generateRandomVector(3, -1, 1),
        acceleration: [0, 0, 0],
        maxSpeed: boidsConfig.maxSpeed,
        maxForce: boidsConfig.maxForce,
        neighbors: new Set()
      };
      
      boidsState.flock.set(boidId, boid);
    }
    
    console.log(`Initialized Boids with ${boidsState.flock.size} agents`);
  }

  async executeBoids() {
    const boidsConfig = this.patterns.boids;
    const boidsState = this.patternState.boids;
    
    // Update neighbors for all boids
    this.updateBoidsNeighbors();
    
    // Apply flocking rules to all boids
    for (const [boidId, boid] of boidsState.flock) {
      // Reset acceleration
      boid.acceleration = [0, 0, 0];
      
      // Apply flocking forces
      const separation = this.calculateSeparation(boid);
      const alignment = this.calculateAlignment(boid);
      const cohesion = this.calculateCohesion(boid);
      
      // Weight and apply forces
      this.addForce(boid.acceleration, separation, boidsConfig.separationWeight);
      this.addForce(boid.acceleration, alignment, boidsConfig.alignmentWeight);
      this.addForce(boid.acceleration, cohesion, boidsConfig.cohesionWeight);
      
      // Update velocity and position
      this.updateBoidVelocity(boid);
      this.updateBoidPosition(boid);
    }
    
    // Calculate flock metrics
    const flockCenter = this.calculateFlockCenter();
    const flockCohesion = this.calculateFlockCohesion();
    
    return {
      pattern: 'boids',
      flockSize: boidsState.flock.size,
      flockCenter,
      cohesion: flockCohesion,
      formations: boidsState.formations.size
    };
  }

  updateBoidsNeighbors() {
    const boidsState = this.patternState.boids;
    const boidsConfig = this.patterns.boids;
    
    // Clear existing neighbors
    for (const [boidId, boid] of boidsState.flock) {
      boid.neighbors.clear();
    }
    
    // Find neighbors for each boid
    for (const [boidId1, boid1] of boidsState.flock) {
      for (const [boidId2, boid2] of boidsState.flock) {
        if (boidId1 !== boidId2) {
          const distance = this.calculateVectorDistance(boid1.position, boid2.position);
          if (distance < boidsConfig.cohesionRadius) {
            boid1.neighbors.add(boidId2);
          }
        }
      }
    }
  }

  calculateSeparation(boid) {
    const boidsState = this.patternState.boids;
    const boidsConfig = this.patterns.boids;
    const steer = [0, 0, 0];
    let count = 0;
    
    for (const neighborId of boid.neighbors) {
      const neighbor = boidsState.flock.get(neighborId);
      if (!neighbor) continue;
      
      const distance = this.calculateVectorDistance(boid.position, neighbor.position);
      if (distance > 0 && distance < boidsConfig.separationRadius) {
        // Calculate vector pointing away from neighbor
        const diff = this.subtractVectors(boid.position, neighbor.position);
        this.normalizeVector(diff);
        this.scaleVector(diff, 1 / distance); // Weight by distance
        
        this.addVectors(steer, diff);
        count++;
      }
    }
    
    if (count > 0) {
      this.scaleVector(steer, 1 / count);
      this.normalizeVector(steer);
      this.scaleVector(steer, boid.maxSpeed);
      this.subtractVectors(steer, boid.velocity);
      this.limitVector(steer, boid.maxForce);
    }
    
    return steer;
  }

  calculateAlignment(boid) {
    const boidsState = this.patternState.boids;
    const boidsConfig = this.patterns.boids;
    const sum = [0, 0, 0];
    let count = 0;
    
    for (const neighborId of boid.neighbors) {
      const neighbor = boidsState.flock.get(neighborId);
      if (!neighbor) continue;
      
      const distance = this.calculateVectorDistance(boid.position, neighbor.position);
      if (distance > 0 && distance < boidsConfig.alignmentRadius) {
        this.addVectors(sum, neighbor.velocity);
        count++;
      }
    }
    
    if (count > 0) {
      this.scaleVector(sum, 1 / count);
      this.normalizeVector(sum);
      this.scaleVector(sum, boid.maxSpeed);
      
      const steer = this.subtractVectors(sum, boid.velocity);
      this.limitVector(steer, boid.maxForce);
      return steer;
    }
    
    return [0, 0, 0];
  }

  calculateCohesion(boid) {
    const boidsState = this.patternState.boids;
    const boidsConfig = this.patterns.boids;
    const sum = [0, 0, 0];
    let count = 0;
    
    for (const neighborId of boid.neighbors) {
      const neighbor = boidsState.flock.get(neighborId);
      if (!neighbor) continue;
      
      const distance = this.calculateVectorDistance(boid.position, neighbor.position);
      if (distance > 0 && distance < boidsConfig.cohesionRadius) {
        this.addVectors(sum, neighbor.position);
        count++;
      }
    }
    
    if (count > 0) {
      this.scaleVector(sum, 1 / count);
      return this.seek(boid, sum);
    }
    
    return [0, 0, 0];
  }

  seek(boid, target) {
    const desired = this.subtractVectors(target, boid.position);
    this.normalizeVector(desired);
    this.scaleVector(desired, boid.maxSpeed);
    
    const steer = this.subtractVectors(desired, boid.velocity);
    this.limitVector(steer, boid.maxForce);
    return steer;
  }

  /**
   * Pattern Execution and Management
   */
  startPatternUpdates() {
    this.patternTimer = setInterval(async () => {
      try {
        await this.updateAllPatterns();
      } catch (error) {
        console.error('Pattern update error:', error);
      }
    }, this.config.updateInterval);
  }

  async updateAllPatterns() {
    const results = {};
    
    // Execute each active pattern
    for (const patternName of this.swarmState.activePatterns) {
      try {
        results[patternName] = await this.executePattern(patternName);
      } catch (error) {
        console.error(`Pattern ${patternName} execution failed:`, error);
      }
    }
    
    // Update global swarm state
    await this.updateGlobalSwarmState(results);
    
    // Check for pattern interactions
    await this.handlePatternInteractions(results);
    
    // Store results
    await this.storePatternResults(results);
    
    this.emit('patterns_updated', {
      results,
      timestamp: Date.now(),
      activePatterns: Array.from(this.swarmState.activePatterns)
    });
  }

  async executePattern(patternName) {
    switch (patternName) {
      case 'aco':
        return await this.executeACO();
      case 'pso':
        return await this.executePSO();
      case 'abc':
        return await this.executeABC();
      case 'boids':
        return await this.executeBoids();
      case 'firefly':
        return await this.executeFirefly();
      case 'wolfpack':
        return await this.executeWolfPack();
      default:
        throw new Error(`Unknown pattern: ${patternName}`);
    }
  }

  async updateGlobalSwarmState(results) {
    // Update global best solution across all patterns
    for (const [patternName, result] of Object.entries(results)) {
      if (result.bestFitness !== undefined) {
        if (!this.swarmState.globalBest || result.bestFitness > this.swarmState.globalBest.fitness) {
          this.swarmState.globalBest = {
            pattern: patternName,
            fitness: result.bestFitness,
            solution: result.bestSolution || result.globalBestPosition,
            timestamp: Date.now()
          };
        }
      }
    }
    
    // Update convergence metrics
    for (const [patternName, result] of Object.entries(results)) {
      if (result.convergence !== undefined) {
        this.swarmState.convergenceMetrics[patternName] = result.convergence;
      }
    }
  }

  async handlePatternInteractions(results) {
    // Cross-pattern information sharing
    await this.shareInformationBetweenPatterns(results);
    
    // Dynamic pattern activation/deactivation
    await this.adaptivePatternManagement(results);
    
    // Hybrid optimization
    await this.hybridOptimization(results);
  }

  async shareInformationBetweenPatterns(results) {
    // Share best solutions between PSO and ABC
    if (results.pso && results.abc && this.swarmState.activePatterns.has('pso') && this.swarmState.activePatterns.has('abc')) {
      const psoState = this.patternState.pso;
      const abcState = this.patternState.abc;
      
      // Introduce best PSO position as new food source in ABC
      if (psoState.globalBestPosition) {
        const newSource = {
          id: `pso_import_${Date.now()}`,
          position: [...psoState.globalBestPosition],
          quality: psoState.globalBestFitness,
          trials: 0
        };
        
        abcState.foodSources.set(newSource.id, newSource);
      }
    }
    
    // Share pheromone information from ACO to other patterns
    if (results.aco && this.swarmState.activePatterns.has('aco')) {
      const bestPath = this.getBestPath();
      if (bestPath) {
        // Influence PSO particles towards best ACO path
        await this.biasPatternTowardsPath('pso', bestPath);
      }
    }
  }

  async adaptivePatternManagement(results) {
    const convergenceThreshold = 0.95;
    
    for (const [patternName, result] of Object.entries(results)) {
      if (result.convergence > convergenceThreshold) {
        // Pattern has converged, consider switching or hybridizing
        console.log(`Pattern ${patternName} converged, considering adaptation`);
        
        // Activate complementary pattern
        const complementaryPattern = this.getComplementaryPattern(patternName);
        if (complementaryPattern && !this.swarmState.activePatterns.has(complementaryPattern)) {
          await this.activatePattern(complementaryPattern);
        }
      }
    }
  }

  async hybridOptimization(results) {
    // Create hybrid solutions by combining insights from multiple patterns
    const hybridSolutions = [];
    
    if (results.pso && results.abc) {
      // Create hybrid solution between PSO and ABC
      const psoSolution = this.patternState.pso.globalBestPosition;
      const abcSolution = this.getBestFoodSource()?.position;
      
      if (psoSolution && abcSolution) {
        const hybrid = this.blendSolutions(psoSolution, abcSolution, 0.5);
        hybridSolutions.push({
          type: 'pso_abc_hybrid',
          solution: hybrid,
          fitness: await this.evaluateHybridSolution(hybrid)
        });
      }
    }
    
    // Evaluate and potentially adopt hybrid solutions
    for (const hybrid of hybridSolutions) {
      if (!this.swarmState.globalBest || hybrid.fitness > this.swarmState.globalBest.fitness) {
        this.swarmState.globalBest = {
          pattern: hybrid.type,
          fitness: hybrid.fitness,
          solution: hybrid.solution,
          timestamp: Date.now()
        };
      }
    }
  }

  /**
   * Utility Methods for Pattern Operations
   */
  generateRandomPosition() {
    return {
      x: Math.random() * 100,
      y: Math.random() * 100
    };
  }

  generateRandomVector(dimensions, min = 0, max = 10) {
    const vector = [];
    for (let i = 0; i < dimensions; i++) {
      vector.push(Math.random() * (max - min) + min);
    }
    return vector;
  }

  calculateDistance(pos1, pos2) {
    const dx = pos1.x - pos2.x;
    const dy = pos1.y - pos2.y;
    return Math.sqrt(dx * dx + dy * dy);
  }

  calculateVectorDistance(vec1, vec2) {
    let sum = 0;
    for (let i = 0; i < vec1.length; i++) {
      const diff = vec1[i] - vec2[i];
      sum += diff * diff;
    }
    return Math.sqrt(sum);
  }

  // Vector operations
  addVectors(a, b) {
    for (let i = 0; i < a.length; i++) {
      a[i] += b[i];
    }
    return a;
  }

  subtractVectors(a, b) {
    const result = [];
    for (let i = 0; i < a.length; i++) {
      result[i] = a[i] - b[i];
    }
    return result;
  }

  scaleVector(vector, scalar) {
    for (let i = 0; i < vector.length; i++) {
      vector[i] *= scalar;
    }
    return vector;
  }

  normalizeVector(vector) {
    const magnitude = Math.sqrt(vector.reduce((sum, component) => sum + component * component, 0));
    if (magnitude > 0) {
      for (let i = 0; i < vector.length; i++) {
        vector[i] /= magnitude;
      }
    }
    return vector;
  }

  limitVector(vector, maxMagnitude) {
    const magnitude = Math.sqrt(vector.reduce((sum, component) => sum + component * component, 0));
    if (magnitude > maxMagnitude) {
      this.normalizeVector(vector);
      this.scaleVector(vector, maxMagnitude);
    }
    return vector;
  }

  addForce(acceleration, force, weight) {
    for (let i = 0; i < acceleration.length; i++) {
      acceleration[i] += force[i] * weight;
    }
  }

  // Placeholder fitness functions - would be replaced with actual problem-specific functions
  async evaluateParticleFitness(particle) {
    // Sphere function as example
    return particle.position.reduce((sum, x) => sum - x * x, 0);
  }

  async evaluateFoodSource(source) {
    // Rastrigin function as example
    const A = 10;
    const n = source.position.length;
    return -A * n - source.position.reduce((sum, x) => sum + (x * x - A * Math.cos(2 * Math.PI * x)), 0);
  }

  async evaluateHybridSolution(solution) {
    // Simple evaluation for hybrid solutions
    return solution.reduce((sum, x) => sum - Math.abs(x), 0);
  }

  // Pattern-specific helper methods
  getPheromoneLevel(from, to) {
    const key = this.getEdgeKey(from, to);
    return this.patternState.aco.pheromoneMap.get(key) || 0.1;
  }

  calculateHeuristicValue(from, to) {
    const distance = this.calculateDistance(from, to);
    return distance > 0 ? 1.0 / distance : 1.0;
  }

  getEdgeKey(from, to) {
    return `${from.x},${from.y}->${to.x},${to.y}`;
  }

  async getAvailablePositions(currentPos, visitedPath) {
    // Generate available positions (simplified)
    const positions = [];
    for (let i = 0; i < 8; i++) {
      const angle = (i / 8) * 2 * Math.PI;
      const newPos = {
        x: currentPos.x + Math.cos(angle) * 10,
        y: currentPos.y + Math.sin(angle) * 10
      };
      
      // Check if not recently visited
      const recentlyVisited = visitedPath.slice(-5).some(pos => 
        this.calculateDistance(pos, newPos) < 5
      );
      
      if (!recentlyVisited) {
        positions.push(newPos);
      }
    }
    
    return positions;
  }

  isTargetReached(ant, position) {
    // Simple target check - would be problem-specific
    return ant.path.length > 20 || 
           (position.x > 80 && position.y > 80);
  }

  getBestPath() {
    const acoState = this.patternState.aco;
    let bestPath = null;
    let bestQuality = -Infinity;
    
    for (const [pathId, quality] of acoState.pathQuality) {
      if (quality > bestQuality) {
        bestQuality = quality;
        bestPath = acoState.bestPaths.get(pathId);
      }
    }
    
    return bestPath;
  }

  checkACOConvergence() {
    const acoState = this.patternState.aco;
    const bestPath = this.getBestPath();
    
    if (!bestPath) return 0;
    
    // Check if most ants are following similar paths
    let similarPaths = 0;
    for (const [antId, ant] of acoState.antColonies) {
      if (this.isPathSimilar(ant.path, bestPath)) {
        similarPaths++;
      }
    }
    
    return similarPaths / acoState.antColonies.size;
  }

  isPathSimilar(path1, path2, threshold = 0.8) {
    if (!path1 || !path2 || path1.length === 0 || path2.length === 0) return false;
    
    const minLength = Math.min(path1.length, path2.length);
    let similarPoints = 0;
    
    for (let i = 0; i < minLength; i++) {
      if (this.calculateDistance(path1[i], path2[i]) < 10) {
        similarPoints++;
      }
    }
    
    return (similarPoints / minLength) >= threshold;
  }

  calculateSwarmCenter() {
    const psoState = this.patternState.pso;
    const dimensions = psoState.globalBestPosition?.length || 0;
    const center = new Array(dimensions).fill(0);
    
    let count = 0;
    for (const [particleId, particle] of psoState.particles) {
      for (let d = 0; d < dimensions; d++) {
        center[d] += particle.position[d];
      }
      count++;
    }
    
    if (count > 0) {
      for (let d = 0; d < dimensions; d++) {
        center[d] /= count;
      }
    }
    
    return center;
  }

  calculateSwarmDiversity() {
    const psoState = this.patternState.pso;
    const center = psoState.swarmCenter;
    
    if (!center) return 0;
    
    let totalDistance = 0;
    let count = 0;
    
    for (const [particleId, particle] of psoState.particles) {
      totalDistance += this.calculateVectorDistance(particle.position, center);
      count++;
    }
    
    return count > 0 ? totalDistance / count : 0;
  }

  checkPSOConvergence() {
    const diversity = this.calculateSwarmDiversity();
    return diversity < this.config.convergenceThreshold ? 1 - diversity : 0;
  }

  calculateSourceProbabilities() {
    const abcState = this.patternState.abc;
    const probabilities = new Map();
    let totalQuality = 0;
    
    // Calculate total quality
    for (const [sourceId, source] of abcState.foodSources) {
      totalQuality += Math.max(0, source.quality);
    }
    
    // Calculate probabilities
    for (const [sourceId, source] of abcState.foodSources) {
      const probability = totalQuality > 0 ? Math.max(0, source.quality) / totalQuality : 1 / abcState.foodSources.size;
      probabilities.set(sourceId, probability);
    }
    
    return probabilities;
  }

  selectSourceByProbability(probabilities) {
    const random = Math.random();
    let cumulativeProbability = 0;
    
    for (const [sourceId, probability] of probabilities) {
      cumulativeProbability += probability;
      if (random <= cumulativeProbability) {
        return sourceId;
      }
    }
    
    // Fallback to first source
    return probabilities.keys().next().value;
  }

  async generateCandidatePosition(position) {
    const candidate = [...position];
    const dimension = Math.floor(Math.random() * candidate.length);
    const partner = this.generateRandomVector(candidate.length);
    
    // ABC candidate generation formula
    candidate[dimension] = position[dimension] + Math.random() * 2 - 1 * (position[dimension] - partner[dimension]);
    
    return candidate;
  }

  getBestFoodSource() {
    const abcState = this.patternState.abc;
    let bestSource = null;
    let bestQuality = -Infinity;
    
    for (const [sourceId, source] of abcState.foodSources) {
      if (source.quality > bestQuality) {
        bestQuality = source.quality;
        bestSource = source;
      }
    }
    
    return bestSource;
  }

  checkABCConvergence() {
    const bestSource = this.getBestFoodSource();
    const abcState = this.patternState.abc;
    
    if (!bestSource) return 0;
    
    // Check how many sources are near the best
    let nearBest = 0;
    for (const [sourceId, source] of abcState.foodSources) {
      const distance = this.calculateVectorDistance(source.position, bestSource.position);
      if (distance < 1.0) {
        nearBest++;
      }
    }
    
    return nearBest / abcState.foodSources.size;
  }

  async updateDanceFloor() {
    const abcState = this.patternState.abc;
    const abcConfig = this.patterns.abc;
    
    abcState.danceFloor.clear();
    
    for (const [beeId, bee] of abcState.employedBees) {
      if (bee.danceValue >= abcConfig.danceThreshold) {
        abcState.danceFloor.set(beeId, {
          bee: beeId,
          source: bee.assignedSource,
          quality: bee.danceValue,
          dancers: 1
        });
      }
    }
  }

  calculateFlockCenter() {
    const boidsState = this.patternState.boids;
    const center = [0, 0, 0];
    let count = 0;
    
    for (const [boidId, boid] of boidsState.flock) {
      for (let i = 0; i < 3; i++) {
        center[i] += boid.position[i];
      }
      count++;
    }
    
    if (count > 0) {
      for (let i = 0; i < 3; i++) {
        center[i] /= count;
      }
    }
    
    return center;
  }

  calculateFlockCohesion() {
    const boidsState = this.patternState.boids;
    const center = this.calculateFlockCenter();
    let totalDistance = 0;
    let count = 0;
    
    for (const [boidId, boid] of boidsState.flock) {
      totalDistance += this.calculateVectorDistance(boid.position, center);
      count++;
    }
    
    return count > 0 ? 1 / (1 + totalDistance / count) : 0;
  }

  updateBoidVelocity(boid) {
    this.addVectors(boid.velocity, boid.acceleration);
    this.limitVector(boid.velocity, boid.maxSpeed);
  }

  updateBoidPosition(boid) {
    this.addVectors(boid.position, boid.velocity);
    
    // Boundary wrapping
    for (let i = 0; i < boid.position.length; i++) {
      if (boid.position[i] < -50) boid.position[i] = 50;
      if (boid.position[i] > 50) boid.position[i] = -50;
    }
  }

  // Pattern management methods
  getComplementaryPattern(patternName) {
    const complements = {
      'aco': 'pso',
      'pso': 'abc',
      'abc': 'aco',
      'boids': 'firefly',
      'firefly': 'wolfpack',
      'wolfpack': 'boids'
    };
    
    return complements[patternName];
  }

  async activatePattern(patternName) {
    if (this.patterns[patternName] && !this.swarmState.activePatterns.has(patternName)) {
      await this.initializePattern(patternName);
      this.swarmState.activePatterns.add(patternName);
      console.log(`Activated pattern: ${patternName}`);
    }
  }

  async biasPatternTowardsPath(patternName, path) {
    if (patternName === 'pso' && path) {
      const psoState = this.patternState.pso;
      
      // Influence some particles towards path positions
      let influenced = 0;
      for (const [particleId, particle] of psoState.particles) {
        if (Math.random() < 0.3 && influenced < 5) { // Influence 30% of particles, max 5
          const pathIndex = Math.floor(Math.random() * path.length);
          const pathPosition = path[pathIndex];
          
          // Convert path position to particle space
          if (pathPosition && pathPosition.x !== undefined && pathPosition.y !== undefined) {
            particle.position[0] = pathPosition.x / 10; // Scale to PSO space
            particle.position[1] = pathPosition.y / 10;
            influenced++;
          }
        }
      }
    }
  }

  blendSolutions(solution1, solution2, alpha) {
    const blended = [];
    for (let i = 0; i < Math.min(solution1.length, solution2.length); i++) {
      blended[i] = alpha * solution1[i] + (1 - alpha) * solution2[i];
    }
    return blended;
  }

  async storePatternResults(results) {
    const timestamp = Date.now();
    const key = `swarm:patterns:results:${timestamp}`;
    
    await this.redis.setex(key, 3600, JSON.stringify({
      timestamp,
      results,
      globalBest: this.swarmState.globalBest,
      convergenceMetrics: this.swarmState.convergenceMetrics,
      activePatterns: Array.from(this.swarmState.activePatterns)
    }));
    
    // Update history
    await this.redis.setex('swarm:patterns:history', 3600 * 24, JSON.stringify({
      convergenceMetrics: this.swarmState.convergenceMetrics,
      globalBest: this.swarmState.globalBest,
      lastUpdate: timestamp
    }));
  }

  // Placeholder implementations for Firefly and Wolf Pack patterns
  async initializeFirefly() {
    console.log('Firefly algorithm initialized (placeholder)');
  }

  async executeFirefly() {
    return {
      pattern: 'firefly',
      brightness: Math.random(),
      synchronization: Math.random(),
      convergence: Math.random()
    };
  }

  async initializeWolfPack() {
    console.log('Wolf Pack algorithm initialized (placeholder)');
  }

  async executeWolfPack() {
    return {
      pattern: 'wolfpack',
      packCohesion: Math.random(),
      huntingSuccess: Math.random(),
      convergence: Math.random()
    };
  }

  // Status and monitoring
  async getStatus() {
    return {
      status: 'operational',
      activePatterns: Array.from(this.swarmState.activePatterns),
      globalBest: this.swarmState.globalBest,
      convergenceMetrics: this.swarmState.convergenceMetrics,
      agentCounts: {
        total: this.swarmState.agents.size,
        aco: this.patternState.aco.antColonies.size,
        pso: this.patternState.pso.particles.size,
        abc: this.patternState.abc.foodSources.size,
        boids: this.patternState.boids.flock.size
      },
      coordinationState: this.swarmState.coordinationState,
      lastUpdate: Date.now()
    };
  }

  async cleanup() {
    if (this.patternTimer) {
      clearInterval(this.patternTimer);
    }
    
    if (this.redis) {
      await this.redis.quit();
    }
  }
}

module.exports = SwarmIntelligencePatterns;