/**
 * Cross-Agent Learning System for Advanced Swarm Intelligence
 * Implements distributed learning, knowledge sharing, skill transfer,
 * collective memory, and adaptive expertise development across agents
 * 
 * Learning Mechanisms:
 * - Distributed Reinforcement Learning with shared experience replay
 * - Knowledge Graph construction and reasoning
 * - Skill Transfer Learning between specialized agents  
 * - Collective Memory with semantic indexing
 * - Meta-Learning for rapid adaptation
 * - Peer-to-Peer Teaching and mentoring
 */

const EventEmitter = require('events');
const Redis = require('ioredis');

class CrossAgentLearningSystem extends EventEmitter {
  constructor(options = {}) {
    super();
    
    this.redis = new Redis({
      host: process.env.REDIS_HOST || 'redis-cluster-0.redis-cluster-service.ollamamax-redis',
      port: process.env.REDIS_PORT || 6379,
      password: process.env.REDIS_PASSWORD || 'ollama_redis_pass',
      retryDelayOnFailure: 1000,
      maxRetriesPerRequest: 3
    });

    // Learning system configuration
    this.config = {
      maxAgents: options.maxAgents || 50,
      learningRate: options.learningRate || 0.01,
      experienceReplaySize: options.experienceReplaySize || 10000,
      knowledgeGraphSize: options.knowledgeGraphSize || 100000,
      skillTransferThreshold: options.skillTransferThreshold || 0.8,
      collectiveMemorySize: options.collectiveMemorySize || 50000,
      teachingInterval: options.teachingInterval || 60000, // 1 minute
      metaLearningInterval: options.metaLearningInterval || 300000, // 5 minutes
      ...options
    };

    // Learning agents registry
    this.agents = new Map(); // agentId -> Agent Learning Profile

    // Distributed Learning Components
    this.distributedRL = {
      experienceReplay: [],           // Shared experience buffer
      globalModel: null,              // Global Q-network or policy
      agentModels: new Map(),        // Agent-specific models
      actionSpace: [],               // Available actions
      stateSpace: [],                // State representation
      rewardSignals: new Map(),      // Reward tracking
      explorationRate: 0.1           // Epsilon for exploration
    };

    // Knowledge Graph System
    this.knowledgeGraph = {
      nodes: new Map(),              // Concepts, facts, procedures
      edges: new Map(),              // Relationships between concepts
      semanticIndex: new Map(),      // Semantic similarity index
      ontologies: new Map(),         // Domain-specific ontologies
      reasoningEngine: null,         // Inference engine
      queryCache: new Map()          // Cached query results
    };

    // Skill Transfer System
    this.skillTransfer = {
      skillLibrary: new Map(),       // Repository of learned skills
      transferMatrix: new Map(),     // Skill compatibility matrix
      mentorAgents: new Set(),       // Agents capable of teaching
      learnerAgents: new Set(),      // Agents seeking to learn
      transferHistory: [],           // Record of successful transfers
      expertiseMap: new Map()        // Agent expertise mapping
    };

    // Collective Memory System
    this.collectiveMemory = {
      episodicMemory: [],            // Sequence of experiences
      semanticMemory: new Map(),     // Factual knowledge
      proceduralMemory: new Map(),   // How-to knowledge
      workingMemory: new Map(),      // Current context
      memoryIndex: new Map(),        // Fast retrieval index
      forgettingCurve: 0.1           // Memory decay rate
    };

    // Meta-Learning System
    this.metaLearning = {
      learningStrategies: new Map(), // Different learning approaches
      adaptationHistory: [],         // Record of adaptations
      performanceMetrics: new Map(), // Learning performance tracking
      optimizationTargets: [],       // What to optimize
      transferLearningRules: new Map(), // Rules for transfer
      curriculumLearning: []         // Structured learning progression
    };

    // Peer Teaching System
    this.peerTeaching = {
      teacherStudentPairs: new Map(), // Current teaching relationships
      teachingMaterials: new Map(),   // Lessons and examples
      assessments: new Map(),         // Learning assessments
      teachingMethods: [],            // Different teaching strategies
      feedbackSystems: new Map(),     // Performance feedback
      graduationCriteria: new Map()   // When to complete learning
    };

    this.initializeLearningSystem();
  }

  async initializeLearningSystem() {
    try {
      // Load persistent learning data
      await this.loadLearningHistory();
      
      // Initialize learning components
      await this.initializeDistributedRL();
      await this.initializeKnowledgeGraph();
      await this.initializeSkillTransfer();
      await this.initializeCollectiveMemory();
      await this.initializeMetaLearning();
      await this.initializePeerTeaching();
      
      // Start learning processes
      this.startLearningCycles();
      
      console.log('Cross-agent learning system initialized successfully');
      this.emit('learning_initialized', {
        components: ['distributedRL', 'knowledgeGraph', 'skillTransfer', 'collectiveMemory', 'metaLearning', 'peerTeaching'],
        maxAgents: this.config.maxAgents,
        learningRate: this.config.learningRate
      });
    } catch (error) {
      console.error('Failed to initialize learning system:', error);
      throw error;
    }
  }

  async loadLearningHistory() {
    const historyData = await this.redis.get('swarm:learning:history');
    if (historyData) {
      try {
        const history = JSON.parse(historyData);
        this.distributedRL.experienceReplay = history.experienceReplay || [];
        this.knowledgeGraph.nodes = new Map(history.knowledgeNodes || []);
        this.skillTransfer.skillLibrary = new Map(history.skillLibrary || []);
        this.collectiveMemory.semanticMemory = new Map(history.semanticMemory || []);
      } catch (e) {
        console.warn('Failed to load learning history:', e);
      }
    }
  }

  /**
   * Agent Registration and Profile Management
   */
  async registerAgent(agentInfo) {
    const agentId = agentInfo.id;
    
    const learningProfile = {
      id: agentId,
      type: agentInfo.type || 'worker',
      capabilities: agentInfo.capabilities || [],
      specializations: agentInfo.specializations || [],
      learningHistory: [],
      skillLevels: new Map(),
      teachingAbilities: new Map(),
      learningPreferences: {
        preferredStyle: agentInfo.learningStyle || 'experiential',
        learningRate: agentInfo.learningRate || this.config.learningRate,
        curiosity: agentInfo.curiosity || 0.5,
        exploration: agentInfo.exploration || 0.3
      },
      performance: {
        accuracy: 0.5,
        speed: 0.5,
        efficiency: 0.5,
        adaptability: 0.5
      },
      relationships: {
        mentors: new Set(),
        students: new Set(),
        peers: new Set(),
        collaborators: new Set()
      },
      created: Date.now(),
      lastActive: Date.now()
    };
    
    this.agents.set(agentId, learningProfile);
    
    // Initialize agent in learning systems
    await this.initializeAgentInSystems(agentId, learningProfile);
    
    this.emit('agent_registered', { agentId, profile: learningProfile });
    console.log(`Registered agent for learning: ${agentId}`);
    
    return learningProfile;
  }

  async initializeAgentInSystems(agentId, profile) {
    // Initialize in distributed RL
    this.distributedRL.agentModels.set(agentId, {
      qTable: new Map(),
      policy: new Map(),
      experience: [],
      lastState: null,
      lastAction: null,
      cumulativeReward: 0
    });
    
    // Add to skill transfer system
    this.skillTransfer.expertiseMap.set(agentId, new Map());
    
    // Initialize memory contributions
    this.collectiveMemory.workingMemory.set(agentId, {
      currentContext: {},
      activeGoals: [],
      recentExperiences: []
    });
    
    // Set up teaching relationships if applicable
    if (profile.capabilities.includes('teaching') || profile.type === 'mentor') {
      this.skillTransfer.mentorAgents.add(agentId);
    } else {
      this.skillTransfer.learnerAgents.add(agentId);
    }
  }

  /**
   * Distributed Reinforcement Learning System
   */
  async initializeDistributedRL() {
    // Define action space (common actions across agents)
    this.distributedRL.actionSpace = [
      'explore_environment',
      'execute_task',
      'share_information',
      'request_help',
      'teach_skill',
      'learn_from_peer',
      'optimize_performance',
      'collaborate',
      'innovate',
      'consolidate_knowledge'
    ];
    
    // Define state space dimensions
    this.distributedRL.stateSpace = [
      'task_complexity',      // Current task difficulty
      'resource_availability', // Available computational resources
      'peer_proximity',       // Number of nearby agents
      'knowledge_confidence', // Confidence in current knowledge
      'performance_trend',    // Recent performance trajectory
      'exploration_level',    // Current exploration vs exploitation
      'collaboration_state',  // Active collaborations
      'learning_progress'     // Recent learning gains
    ];
    
    // Initialize global model (simplified Q-table approach)
    this.distributedRL.globalModel = {
      qTable: new Map(),
      updateCount: 0,
      lastUpdate: Date.now()
    };
    
    console.log('Distributed RL system initialized');
  }

  async executeDistributedRL() {
    const agents = Array.from(this.agents.keys());
    
    // Distribute learning updates across agents
    for (const agentId of agents) {
      await this.updateAgentRL(agentId);
    }
    
    // Aggregate learning from all agents
    await this.aggregateGlobalLearning();
    
    // Share updated model with all agents
    await this.distributeGlobalModel();
    
    return {
      activeAgents: agents.length,
      experienceReplaySize: this.distributedRL.experienceReplay.length,
      globalModelUpdates: this.distributedRL.globalModel.updateCount,
      averageReward: this.calculateAverageReward()
    };
  }

  async updateAgentRL(agentId) {
    const agent = this.agents.get(agentId);
    const agentModel = this.distributedRL.agentModels.get(agentId);
    
    if (!agent || !agentModel) return;
    
    // Get current state
    const currentState = await this.getAgentState(agentId);
    
    // Select action using epsilon-greedy policy
    const action = this.selectAction(agentId, currentState);
    
    // Execute action and observe reward
    const reward = await this.executeAction(agentId, action);
    
    // Store experience
    if (agentModel.lastState && agentModel.lastAction !== null) {
      const experience = {
        agentId,
        state: agentModel.lastState,
        action: agentModel.lastAction,
        reward,
        nextState: currentState,
        timestamp: Date.now()
      };
      
      // Add to agent's personal experience
      agentModel.experience.push(experience);
      if (agentModel.experience.length > 1000) {
        agentModel.experience.shift(); // Keep only recent experiences
      }
      
      // Add to shared experience replay
      this.distributedRL.experienceReplay.push(experience);
      if (this.distributedRL.experienceReplay.length > this.config.experienceReplaySize) {
        this.distributedRL.experienceReplay.shift();
      }
      
      // Update Q-value
      await this.updateQValue(agentId, experience);
    }
    
    // Update agent state
    agentModel.lastState = currentState;
    agentModel.lastAction = action;
    agentModel.cumulativeReward += reward;
  }

  async getAgentState(agentId) {
    const agent = this.agents.get(agentId);
    if (!agent) return [];
    
    // Create state vector based on state space
    const state = [];
    
    // Task complexity (0-1)
    state.push(await this.getTaskComplexity(agentId));
    
    // Resource availability (0-1)
    state.push(await this.getResourceAvailability(agentId));
    
    // Peer proximity (normalized)
    state.push(await this.getPeerProximity(agentId));
    
    // Knowledge confidence (0-1)
    state.push(await this.getKnowledgeConfidence(agentId));
    
    // Performance trend (-1 to 1)
    state.push(await this.getPerformanceTrend(agentId));
    
    // Exploration level (0-1)
    state.push(agent.learningPreferences.exploration);
    
    // Collaboration state (0-1)
    state.push(await this.getCollaborationState(agentId));
    
    // Learning progress (0-1)
    state.push(await this.getLearningProgress(agentId));
    
    return state;
  }

  selectAction(agentId, state) {
    const agentModel = this.distributedRL.agentModels.get(agentId);
    const agent = this.agents.get(agentId);
    
    // Epsilon-greedy action selection
    if (Math.random() < this.distributedRL.explorationRate * agent.learningPreferences.exploration) {
      // Explore: random action
      return Math.floor(Math.random() * this.distributedRL.actionSpace.length);
    } else {
      // Exploit: best known action
      const stateKey = this.encodeState(state);
      const qValues = agentModel.qTable.get(stateKey) || new Array(this.distributedRL.actionSpace.length).fill(0);
      
      // Select action with highest Q-value
      let bestAction = 0;
      let maxQ = qValues[0];
      
      for (let i = 1; i < qValues.length; i++) {
        if (qValues[i] > maxQ) {
          maxQ = qValues[i];
          bestAction = i;
        }
      }
      
      return bestAction;
    }
  }

  async executeAction(agentId, actionIndex) {
    const actionName = this.distributedRL.actionSpace[actionIndex];
    let reward = 0;
    
    switch (actionName) {
      case 'explore_environment':
        reward = await this.executeExploration(agentId);
        break;
      case 'execute_task':
        reward = await this.executeTask(agentId);
        break;
      case 'share_information':
        reward = await this.shareInformation(agentId);
        break;
      case 'request_help':
        reward = await this.requestHelp(agentId);
        break;
      case 'teach_skill':
        reward = await this.teachSkill(agentId);
        break;
      case 'learn_from_peer':
        reward = await this.learnFromPeer(agentId);
        break;
      case 'optimize_performance':
        reward = await this.optimizePerformance(agentId);
        break;
      case 'collaborate':
        reward = await this.initiateCollaboration(agentId);
        break;
      case 'innovate':
        reward = await this.attemptInnovation(agentId);
        break;
      case 'consolidate_knowledge':
        reward = await this.consolidateKnowledge(agentId);
        break;
      default:
        reward = 0;
    }
    
    // Record reward
    this.distributedRL.rewardSignals.set(`${agentId}_${Date.now()}`, reward);
    
    return reward;
  }

  async updateQValue(agentId, experience) {
    const agentModel = this.distributedRL.agentModels.get(agentId);
    const agent = this.agents.get(agentId);
    
    const stateKey = this.encodeState(experience.state);
    const nextStateKey = this.encodeState(experience.nextState);
    
    // Initialize Q-values if not exist
    if (!agentModel.qTable.has(stateKey)) {
      agentModel.qTable.set(stateKey, new Array(this.distributedRL.actionSpace.length).fill(0));
    }
    if (!agentModel.qTable.has(nextStateKey)) {
      agentModel.qTable.set(nextStateKey, new Array(this.distributedRL.actionSpace.length).fill(0));
    }
    
    const qValues = agentModel.qTable.get(stateKey);
    const nextQValues = agentModel.qTable.get(nextStateKey);
    
    // Q-learning update: Q(s,a) = Q(s,a) + α[r + γ*max(Q(s',a')) - Q(s,a)]
    const learningRate = agent.learningPreferences.learningRate;
    const discountFactor = 0.9;
    
    const maxNextQ = Math.max(...nextQValues);
    const currentQ = qValues[experience.action];
    
    qValues[experience.action] = currentQ + learningRate * (
      experience.reward + discountFactor * maxNextQ - currentQ
    );
    
    agentModel.qTable.set(stateKey, qValues);
  }

  /**
   * Knowledge Graph System
   */
  async initializeKnowledgeGraph() {
    // Initialize core knowledge nodes
    const coreKnowledge = [
      { id: 'task_execution', type: 'procedure', domain: 'general' },
      { id: 'collaboration', type: 'concept', domain: 'social' },
      { id: 'problem_solving', type: 'skill', domain: 'cognitive' },
      { id: 'communication', type: 'skill', domain: 'social' },
      { id: 'learning', type: 'meta_skill', domain: 'cognitive' },
      { id: 'optimization', type: 'procedure', domain: 'technical' },
      { id: 'resource_management', type: 'skill', domain: 'operational' },
      { id: 'quality_assurance', type: 'procedure', domain: 'quality' }
    ];
    
    for (const knowledge of coreKnowledge) {
      this.knowledgeGraph.nodes.set(knowledge.id, {
        ...knowledge,
        confidence: 0.8,
        relevance: 1.0,
        usage_count: 0,
        contributors: new Set(),
        created: Date.now(),
        updated: Date.now()
      });
    }
    
    // Initialize relationships
    const relationships = [
      { from: 'problem_solving', to: 'task_execution', type: 'enables' },
      { from: 'collaboration', to: 'communication', type: 'requires' },
      { from: 'learning', to: 'problem_solving', type: 'enhances' },
      { from: 'optimization', to: 'resource_management', type: 'improves' },
      { from: 'quality_assurance', to: 'task_execution', type: 'validates' }
    ];
    
    for (const rel of relationships) {
      const edgeKey = `${rel.from}-${rel.type}-${rel.to}`;
      this.knowledgeGraph.edges.set(edgeKey, {
        from: rel.from,
        to: rel.to,
        type: rel.type,
        weight: 0.8,
        confidence: 0.9,
        created: Date.now()
      });
    }
    
    console.log('Knowledge graph initialized with core knowledge');
  }

  async addKnowledge(knowledge, contributorId) {
    const nodeId = knowledge.id || `knowledge_${Date.now()}`;
    
    const node = {
      id: nodeId,
      type: knowledge.type || 'concept',
      domain: knowledge.domain || 'general',
      content: knowledge.content || {},
      description: knowledge.description || '',
      confidence: knowledge.confidence || 0.7,
      relevance: knowledge.relevance || 0.8,
      usage_count: 0,
      contributors: new Set([contributorId]),
      created: Date.now(),
      updated: Date.now()
    };
    
    this.knowledgeGraph.nodes.set(nodeId, node);
    
    // Update semantic index
    await this.updateSemanticIndex(nodeId, node);
    
    // Create relationships if specified
    if (knowledge.relationships) {
      for (const rel of knowledge.relationships) {
        await this.addKnowledgeRelationship(nodeId, rel.target, rel.type, rel.weight);
      }
    }
    
    this.emit('knowledge_added', { nodeId, contributor: contributorId });
    
    return nodeId;
  }

  async addKnowledgeRelationship(fromId, toId, relationType, weight = 0.8) {
    const edgeKey = `${fromId}-${relationType}-${toId}`;
    
    const edge = {
      from: fromId,
      to: toId,
      type: relationType,
      weight,
      confidence: 0.8,
      usage_count: 0,
      created: Date.now()
    };
    
    this.knowledgeGraph.edges.set(edgeKey, edge);
    
    return edgeKey;
  }

  async queryKnowledge(query, agentId) {
    const queryKey = `${agentId}_${JSON.stringify(query)}`;
    
    // Check cache first
    if (this.knowledgeGraph.queryCache.has(queryKey)) {
      return this.knowledgeGraph.queryCache.get(queryKey);
    }
    
    const results = [];
    
    // Search by type
    if (query.type) {
      for (const [nodeId, node] of this.knowledgeGraph.nodes) {
        if (node.type === query.type) {
          results.push({ nodeId, node, relevance: node.relevance });
        }
      }
    }
    
    // Search by domain
    if (query.domain) {
      for (const [nodeId, node] of this.knowledgeGraph.nodes) {
        if (node.domain === query.domain) {
          results.push({ nodeId, node, relevance: node.relevance });
        }
      }
    }
    
    // Search by keywords
    if (query.keywords) {
      for (const [nodeId, node] of this.knowledgeGraph.nodes) {
        const relevance = this.calculateSemanticSimilarity(query.keywords, node);
        if (relevance > 0.3) {
          results.push({ nodeId, node, relevance });
        }
      }
    }
    
    // Sort by relevance
    results.sort((a, b) => b.relevance - a.relevance);
    
    // Cache results
    this.knowledgeGraph.queryCache.set(queryKey, results.slice(0, 10));
    
    return results.slice(0, 10);
  }

  /**
   * Skill Transfer System
   */
  async initializeSkillTransfer() {
    // Define skill categories
    const skillCategories = [
      'technical_skills',
      'problem_solving',
      'communication',
      'leadership',
      'creativity',
      'analysis',
      'optimization',
      'collaboration'
    ];
    
    // Initialize skill library with basic skills
    for (const category of skillCategories) {
      this.skillTransfer.skillLibrary.set(category, {
        category,
        skills: new Map(),
        prerequisites: [],
        difficulty: 0.5,
        transferability: 0.7,
        learningTime: 3600000, // 1 hour in ms
        created: Date.now()
      });
    }
    
    // Initialize transfer compatibility matrix
    this.initializeTransferMatrix();
    
    console.log('Skill transfer system initialized');
  }

  initializeTransferMatrix() {
    const skills = Array.from(this.skillTransfer.skillLibrary.keys());
    
    for (const skill1 of skills) {
      for (const skill2 of skills) {
        const compatibility = this.calculateSkillCompatibility(skill1, skill2);
        this.skillTransfer.transferMatrix.set(`${skill1}-${skill2}`, compatibility);
      }
    }
  }

  calculateSkillCompatibility(skill1, skill2) {
    // Simplified compatibility calculation
    if (skill1 === skill2) return 1.0;
    
    // Domain-based compatibility
    const relatedPairs = [
      ['technical_skills', 'optimization'],
      ['problem_solving', 'analysis'],
      ['communication', 'collaboration'],
      ['leadership', 'communication']
    ];
    
    for (const pair of relatedPairs) {
      if ((skill1 === pair[0] && skill2 === pair[1]) || (skill1 === pair[1] && skill2 === pair[0])) {
        return 0.8;
      }
    }
    
    return 0.3; // Base compatibility
  }

  async identifySkillTransferOpportunities() {
    const opportunities = [];
    
    // Find mentor-student pairs
    for (const mentorId of this.skillTransfer.mentorAgents) {
      const mentor = this.agents.get(mentorId);
      if (!mentor) continue;
      
      for (const learnerId of this.skillTransfer.learnerAgents) {
        if (learnerId === mentorId) continue;
        
        const learner = this.agents.get(learnerId);
        if (!learner) continue;
        
        // Find skills mentor can teach that learner needs
        const teachableSkills = await this.findTeachableSkills(mentorId, learnerId);
        
        if (teachableSkills.length > 0) {
          opportunities.push({
            mentor: mentorId,
            learner: learnerId,
            skills: teachableSkills,
            priority: this.calculateTransferPriority(mentor, learner, teachableSkills)
          });
        }
      }
    }
    
    // Sort by priority
    opportunities.sort((a, b) => b.priority - a.priority);
    
    return opportunities;
  }

  async findTeachableSkills(mentorId, learnerId) {
    const mentor = this.agents.get(mentorId);
    const learner = this.agents.get(learnerId);
    
    const teachableSkills = [];
    
    // Compare skill levels
    const mentorExpertise = this.skillTransfer.expertiseMap.get(mentorId) || new Map();
    const learnerExpertise = this.skillTransfer.expertiseMap.get(learnerId) || new Map();
    
    for (const [skill, mentorLevel] of mentorExpertise) {
      const learnerLevel = learnerExpertise.get(skill) || 0;
      
      // Mentor can teach if they're significantly more skilled
      if (mentorLevel >= this.config.skillTransferThreshold && mentorLevel - learnerLevel > 0.3) {
        teachableSkills.push({
          skill,
          mentorLevel,
          learnerLevel,
          transferDifficulty: this.calculateTransferDifficulty(skill, mentorLevel, learnerLevel)
        });
      }
    }
    
    return teachableSkills;
  }

  calculateTransferPriority(mentor, learner, skills) {
    let priority = 0;
    
    // Higher priority for more skills to transfer
    priority += skills.length * 0.2;
    
    // Higher priority for larger skill gaps
    const avgGap = skills.reduce((sum, s) => sum + (s.mentorLevel - s.learnerLevel), 0) / skills.length;
    priority += avgGap * 0.3;
    
    // Higher priority for compatible learning styles
    if (mentor.learningPreferences.preferredStyle === learner.learningPreferences.preferredStyle) {
      priority += 0.2;
    }
    
    // Higher priority for recent high performers
    priority += (mentor.performance.efficiency + learner.performance.adaptability) * 0.15;
    
    return priority;
  }

  async executeSkillTransfer(mentorId, learnerId, skill) {
    console.log(`Initiating skill transfer: ${skill} from ${mentorId} to ${learnerId}`);
    
    // Create teaching session
    const sessionId = `transfer_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const session = {
      id: sessionId,
      mentor: mentorId,
      learner: learnerId,
      skill,
      status: 'active',
      startTime: Date.now(),
      lessons: [],
      assessments: [],
      progress: 0,
      method: this.selectTeachingMethod(mentorId, learnerId, skill)
    };
    
    // Add to peer teaching system
    this.peerTeaching.teacherStudentPairs.set(sessionId, session);
    
    // Update agent relationships
    const mentor = this.agents.get(mentorId);
    const learner = this.agents.get(learnerId);
    
    mentor.relationships.students.add(learnerId);
    learner.relationships.mentors.add(mentorId);
    
    // Begin teaching process
    await this.conductTeachingSession(sessionId);
    
    return sessionId;
  }

  selectTeachingMethod(mentorId, learnerId, skill) {
    const mentor = this.agents.get(mentorId);
    const learner = this.agents.get(learnerId);
    
    const methods = ['demonstration', 'guided_practice', 'collaborative_learning', 'problem_based', 'scaffolded_instruction'];
    
    // Select based on learner preferences and skill type
    if (learner.learningPreferences.preferredStyle === 'visual') {
      return 'demonstration';
    } else if (learner.learningPreferences.preferredStyle === 'experiential') {
      return 'guided_practice';
    } else if (skill.includes('collaboration') || skill.includes('communication')) {
      return 'collaborative_learning';
    }
    
    return methods[Math.floor(Math.random() * methods.length)];
  }

  /**
   * Collective Memory System
   */
  async initializeCollectiveMemory() {
    // Initialize memory structures
    this.collectiveMemory.episodicMemory = [];
    this.collectiveMemory.semanticMemory = new Map();
    this.collectiveMemory.proceduralMemory = new Map();
    
    // Add basic procedural knowledge
    const basicProcedures = [
      {
        id: 'task_completion_process',
        steps: ['analyze_requirements', 'plan_approach', 'execute_solution', 'validate_results', 'document_learnings'],
        domain: 'general',
        effectiveness: 0.8
      },
      {
        id: 'collaboration_protocol',
        steps: ['establish_communication', 'define_roles', 'coordinate_actions', 'share_results', 'provide_feedback'],
        domain: 'social',
        effectiveness: 0.9
      },
      {
        id: 'learning_protocol',
        steps: ['identify_knowledge_gap', 'find_learning_resources', 'practice_skills', 'apply_knowledge', 'reflect_on_experience'],
        domain: 'learning',
        effectiveness: 0.85
      }
    ];
    
    for (const procedure of basicProcedures) {
      this.collectiveMemory.proceduralMemory.set(procedure.id, {
        ...procedure,
        usageCount: 0,
        successRate: 0.8,
        contributors: new Set(),
        created: Date.now(),
        updated: Date.now()
      });
    }
    
    console.log('Collective memory system initialized');
  }

  async addToCollectiveMemory(memoryType, content, contributorId) {
    const timestamp = Date.now();
    
    switch (memoryType) {
      case 'episodic':
        const episode = {
          id: `episode_${timestamp}`,
          content,
          contributor: contributorId,
          timestamp,
          context: await this.getCurrentContext(contributorId),
          emotional_valence: content.emotional_valence || 0,
          importance: content.importance || 0.5
        };
        
        this.collectiveMemory.episodicMemory.push(episode);
        
        // Maintain memory size limit
        if (this.collectiveMemory.episodicMemory.length > 10000) {
          this.collectiveMemory.episodicMemory.shift();
        }
        break;
        
      case 'semantic':
        const factId = content.id || `fact_${timestamp}`;
        this.collectiveMemory.semanticMemory.set(factId, {
          id: factId,
          fact: content.fact,
          domain: content.domain || 'general',
          confidence: content.confidence || 0.8,
          sources: new Set([contributorId]),
          created: timestamp,
          validated: false
        });
        break;
        
      case 'procedural':
        const procId = content.id || `procedure_${timestamp}`;
        this.collectiveMemory.proceduralMemory.set(procId, {
          id: procId,
          name: content.name,
          steps: content.steps,
          domain: content.domain || 'general',
          effectiveness: content.effectiveness || 0.5,
          usageCount: 0,
          successRate: content.successRate || 0.5,
          contributors: new Set([contributorId]),
          created: timestamp,
          updated: timestamp
        });
        break;
    }
    
    // Update memory index for fast retrieval
    await this.updateMemoryIndex(memoryType, content, contributorId);
    
    this.emit('memory_added', { type: memoryType, contributor: contributorId, timestamp });
  }

  async retrieveFromCollectiveMemory(query, agentId) {
    const results = {
      episodic: [],
      semantic: [],
      procedural: []
    };
    
    // Search episodic memory
    if (query.includeEpisodic !== false) {
      results.episodic = this.searchEpisodicMemory(query);
    }
    
    // Search semantic memory
    if (query.includeSemantic !== false) {
      results.semantic = this.searchSemanticMemory(query);
    }
    
    // Search procedural memory
    if (query.includeProcedural !== false) {
      results.procedural = this.searchProceduralMemory(query);
    }
    
    // Apply forgetting curve
    this.applyForgettingCurve(results);
    
    return results;
  }

  searchEpisodicMemory(query) {
    const results = [];
    const now = Date.now();
    
    for (const episode of this.collectiveMemory.episodicMemory) {
      let relevance = 0;
      
      // Temporal relevance (more recent = more relevant)
      const age = now - episode.timestamp;
      const temporalRelevance = Math.exp(-age / (24 * 3600 * 1000)); // 24-hour decay
      
      // Content relevance
      const contentRelevance = this.calculateContentSimilarity(query, episode.content);
      
      // Context relevance
      const contextRelevance = query.context ? 
        this.calculateContextSimilarity(query.context, episode.context) : 0.5;
      
      relevance = temporalRelevance * 0.3 + contentRelevance * 0.5 + contextRelevance * 0.2;
      
      if (relevance > 0.2) {
        results.push({ ...episode, relevance });
      }
    }
    
    return results.sort((a, b) => b.relevance - a.relevance).slice(0, 10);
  }

  searchSemanticMemory(query) {
    const results = [];
    
    for (const [factId, fact] of this.collectiveMemory.semanticMemory) {
      let relevance = fact.confidence * 0.3;
      
      // Domain matching
      if (query.domain && fact.domain === query.domain) {
        relevance += 0.3;
      }
      
      // Content similarity
      if (query.keywords) {
        relevance += this.calculateKeywordSimilarity(query.keywords, fact.fact) * 0.4;
      }
      
      if (relevance > 0.3) {
        results.push({ ...fact, relevance });
      }
    }
    
    return results.sort((a, b) => b.relevance - a.relevance).slice(0, 10);
  }

  searchProceduralMemory(query) {
    const results = [];
    
    for (const [procId, procedure] of this.collectiveMemory.proceduralMemory) {
      let relevance = procedure.effectiveness * 0.4 + procedure.successRate * 0.3;
      
      // Domain matching
      if (query.domain && procedure.domain === query.domain) {
        relevance += 0.3;
      }
      
      if (relevance > 0.3) {
        results.push({ ...procedure, relevance });
      }
    }
    
    return results.sort((a, b) => b.relevance - a.relevance).slice(0, 5);
  }

  /**
   * Learning Process Execution
   */
  startLearningCycles() {
    // Main learning update cycle
    this.learningTimer = setInterval(async () => {
      try {
        await this.executeLearningCycle();
      } catch (error) {
        console.error('Learning cycle error:', error);
      }
    }, this.config.teachingInterval);
    
    // Meta-learning cycle
    this.metaLearningTimer = setInterval(async () => {
      try {
        await this.executeMetaLearning();
      } catch (error) {
        console.error('Meta-learning error:', error);
      }
    }, this.config.metaLearningInterval);
  }

  async executeLearningCycle() {
    console.log('Executing learning cycle...');
    
    // 1. Distributed RL updates
    const rlResults = await this.executeDistributedRL();
    
    // 2. Knowledge sharing
    await this.facilitateKnowledgeSharing();
    
    // 3. Skill transfer opportunities
    const transferOpportunities = await this.identifySkillTransferOpportunities();
    
    // 4. Execute top transfer opportunities
    for (const opportunity of transferOpportunities.slice(0, 3)) {
      for (const skillTransfer of opportunity.skills.slice(0, 1)) {
        await this.executeSkillTransfer(opportunity.mentor, opportunity.learner, skillTransfer.skill);
      }
    }
    
    // 5. Update collective memory
    await this.updateCollectiveMemoryFromExperiences();
    
    // 6. Peer teaching sessions
    await this.conductActivePeerTeachingSessions();
    
    const results = {
      rlResults,
      knowledgeShared: await this.getKnowledgeSharingStats(),
      skillTransfers: transferOpportunities.length,
      activeTeachingSessions: this.peerTeaching.teacherStudentPairs.size,
      memorySize: {
        episodic: this.collectiveMemory.episodicMemory.length,
        semantic: this.collectiveMemory.semanticMemory.size,
        procedural: this.collectiveMemory.proceduralMemory.size
      }
    };
    
    await this.storeLearningResults(results);
    
    this.emit('learning_cycle_complete', results);
    
    return results;
  }

  async facilitateKnowledgeSharing() {
    const agents = Array.from(this.agents.keys());
    
    for (const agentId of agents) {
      // Check if agent has knowledge to share
      const agent = this.agents.get(agentId);
      if (!agent) continue;
      
      // Share recent experiences
      const recentExperiences = this.getRecentExperiences(agentId);
      for (const experience of recentExperiences.slice(0, 3)) {
        if (experience.shareWorthy) {
          await this.shareExperience(agentId, experience);
        }
      }
      
      // Share successful procedures
      const successfulProcedures = this.getSuccessfulProcedures(agentId);
      for (const procedure of successfulProcedures.slice(0, 2)) {
        await this.addToCollectiveMemory('procedural', procedure, agentId);
      }
    }
  }

  async updateCollectiveMemoryFromExperiences() {
    // Process recent RL experiences for memory formation
    const recentExperiences = this.distributedRL.experienceReplay.slice(-100);
    
    for (const experience of recentExperiences) {
      // Convert high-reward experiences to episodic memories
      if (experience.reward > 0.8) {
        await this.addToCollectiveMemory('episodic', {
          action: this.distributedRL.actionSpace[experience.action],
          state: experience.state,
          reward: experience.reward,
          success: true,
          context: experience.context,
          importance: experience.reward,
          shareWorthy: true
        }, experience.agentId);
      }
      
      // Extract patterns for semantic memory
      if (experience.reward > 0.6) {
        const pattern = this.extractPattern(experience);
        if (pattern) {
          await this.addToCollectiveMemory('semantic', {
            fact: pattern,
            domain: this.inferDomain(experience),
            confidence: experience.reward
          }, experience.agentId);
        }
      }
    }
  }

  async conductActivePeerTeachingSessions() {
    for (const [sessionId, session] of this.peerTeaching.teacherStudentPairs) {
      if (session.status === 'active') {
        await this.conductTeachingSession(sessionId);
      }
    }
  }

  async conductTeachingSession(sessionId) {
    const session = this.peerTeaching.teacherStudentPairs.get(sessionId);
    if (!session) return;
    
    const mentor = this.agents.get(session.mentor);
    const learner = this.agents.get(session.learner);
    
    if (!mentor || !learner) {
      session.status = 'failed';
      return;
    }
    
    // Create lesson based on teaching method
    const lesson = await this.createLesson(session);
    
    // Deliver lesson
    const deliveryResult = await this.deliverLesson(session, lesson);
    
    // Assess learning
    const assessment = await this.assessLearning(session, lesson);
    
    // Update progress
    session.progress = Math.min(1.0, session.progress + assessment.improvement);
    session.lessons.push(lesson);
    session.assessments.push(assessment);
    
    // Check for completion
    if (session.progress >= 0.9 || assessment.mastery) {
      session.status = 'completed';
      await this.graduateStudent(sessionId);
    } else if (Date.now() - session.startTime > 3600000) { // 1 hour timeout
      session.status = 'timeout';
    }
    
    console.log(`Teaching session ${sessionId}: progress ${session.progress.toFixed(2)}, status ${session.status}`);
  }

  async createLesson(session) {
    const mentor = this.agents.get(session.mentor);
    const learner = this.agents.get(session.learner);
    
    return {
      id: `lesson_${Date.now()}`,
      skill: session.skill,
      method: session.method,
      content: this.generateLessonContent(session.skill, session.method),
      exercises: this.generatePracticeExercises(session.skill),
      examples: await this.getRelevantExamples(session.skill),
      difficulty: this.calculateLessonDifficulty(learner, session.skill),
      timestamp: Date.now()
    };
  }

  async deliverLesson(session, lesson) {
    // Simulate lesson delivery based on method
    const deliveryEffectiveness = this.calculateDeliveryEffectiveness(session, lesson);
    
    return {
      effectiveness: deliveryEffectiveness,
      engagement: Math.random() * 0.4 + 0.6, // 0.6-1.0
      comprehension: deliveryEffectiveness * (Math.random() * 0.3 + 0.7),
      timestamp: Date.now()
    };
  }

  async assessLearning(session, lesson) {
    const learner = this.agents.get(session.learner);
    
    // Simulate assessment based on learner capabilities and lesson quality
    const baseScore = learner.performance.adaptability * 0.6 + lesson.content.quality * 0.4;
    const improvement = Math.min(0.3, baseScore * (Math.random() * 0.4 + 0.8));
    
    return {
      score: baseScore,
      improvement,
      mastery: baseScore > 0.85,
      strengths: this.identifyLearningStrengths(learner, lesson),
      weaknesses: this.identifyLearningWeaknesses(learner, lesson),
      timestamp: Date.now()
    };
  }

  // Utility methods and placeholders
  async getTaskComplexity(agentId) {
    // Placeholder - would analyze current task complexity
    return Math.random() * 0.5 + 0.3; // 0.3-0.8
  }

  async getResourceAvailability(agentId) {
    // Placeholder - would check available computational resources
    return Math.random() * 0.4 + 0.6; // 0.6-1.0
  }

  async getPeerProximity(agentId) {
    // Placeholder - would count nearby agents
    const nearbyAgents = Math.floor(Math.random() * 5);
    return Math.min(1.0, nearbyAgents / 5);
  }

  async getKnowledgeConfidence(agentId) {
    const agent = this.agents.get(agentId);
    return agent ? (agent.performance.accuracy + agent.performance.efficiency) / 2 : 0.5;
  }

  async getPerformanceTrend(agentId) {
    // Placeholder - would analyze recent performance history
    return (Math.random() - 0.5) * 2; // -1 to 1
  }

  async getCollaborationState(agentId) {
    const agent = this.agents.get(agentId);
    return agent ? agent.relationships.collaborators.size / 10 : 0;
  }

  async getLearningProgress(agentId) {
    const agent = this.agents.get(agentId);
    if (!agent) return 0;
    
    const recentLearning = agent.learningHistory.slice(-10);
    return recentLearning.length > 0 ? 
      recentLearning.reduce((sum, l) => sum + l.improvement, 0) / recentLearning.length : 0.5;
  }

  encodeState(state) {
    return state.map(s => Math.round(s * 10) / 10).join(',');
  }

  calculateAverageReward() {
    if (this.distributedRL.experienceReplay.length === 0) return 0;
    
    const recentRewards = this.distributedRL.experienceReplay.slice(-100).map(e => e.reward);
    return recentRewards.reduce((sum, r) => sum + r, 0) / recentRewards.length;
  }

  // Action execution placeholders
  async executeExploration(agentId) {
    return Math.random() * 0.4 + 0.1; // 0.1-0.5 reward for exploration
  }

  async executeTask(agentId) {
    const agent = this.agents.get(agentId);
    return agent ? agent.performance.efficiency * (Math.random() * 0.4 + 0.6) : 0.5;
  }

  async shareInformation(agentId) {
    const shared = Math.random() > 0.5;
    return shared ? 0.3 + Math.random() * 0.3 : 0.1;
  }

  async requestHelp(agentId) {
    const helpReceived = Math.random() > 0.3;
    return helpReceived ? 0.4 + Math.random() * 0.4 : 0;
  }

  async teachSkill(agentId) {
    if (this.skillTransfer.mentorAgents.has(agentId)) {
      return 0.5 + Math.random() * 0.4; // 0.5-0.9 reward for teaching
    }
    return 0.1;
  }

  async learnFromPeer(agentId) {
    return 0.3 + Math.random() * 0.5; // 0.3-0.8 reward for learning
  }

  async optimizePerformance(agentId) {
    const agent = this.agents.get(agentId);
    if (agent) {
      agent.performance.efficiency = Math.min(1.0, agent.performance.efficiency + 0.05);
      return 0.4 + Math.random() * 0.3;
    }
    return 0.2;
  }

  async initiateCollaboration(agentId) {
    const collaborationSuccessful = Math.random() > 0.4;
    return collaborationSuccessful ? 0.6 + Math.random() * 0.3 : 0.1;
  }

  async attemptInnovation(agentId) {
    const innovative = Math.random() > 0.7;
    return innovative ? 0.8 + Math.random() * 0.2 : 0.05;
  }

  async consolidateKnowledge(agentId) {
    return 0.3 + Math.random() * 0.2; // Steady reward for consolidation
  }

  async aggregateGlobalLearning() {
    // Aggregate Q-values from all agents
    const globalQTable = new Map();
    
    for (const [agentId, agentModel] of this.distributedRL.agentModels) {
      for (const [state, qValues] of agentModel.qTable) {
        if (!globalQTable.has(state)) {
          globalQTable.set(state, new Array(this.distributedRL.actionSpace.length).fill(0));
        }
        
        const globalQValues = globalQTable.get(state);
        for (let i = 0; i < qValues.length; i++) {
          globalQValues[i] += qValues[i] / this.distributedRL.agentModels.size;
        }
      }
    }
    
    this.distributedRL.globalModel.qTable = globalQTable;
    this.distributedRL.globalModel.updateCount++;
    this.distributedRL.globalModel.lastUpdate = Date.now();
  }

  async distributeGlobalModel() {
    // Share global knowledge back to individual agents
    for (const [agentId, agentModel] of this.distributedRL.agentModels) {
      // Blend global and local Q-values
      for (const [state, globalQValues] of this.distributedRL.globalModel.qTable) {
        if (agentModel.qTable.has(state)) {
          const localQValues = agentModel.qTable.get(state);
          const blendedQValues = [];
          
          for (let i = 0; i < globalQValues.length; i++) {
            // 70% local, 30% global blend
            blendedQValues[i] = localQValues[i] * 0.7 + globalQValues[i] * 0.3;
          }
          
          agentModel.qTable.set(state, blendedQValues);
        } else {
          // Adopt global Q-values for new states
          agentModel.qTable.set(state, [...globalQValues]);
        }
      }
    }
  }

  // Additional helper methods would be implemented here...
  // This includes semantic similarity calculations, memory indexing,
  // teaching content generation, assessment methods, etc.

  async updateSemanticIndex(nodeId, node) {
    // Placeholder for semantic indexing
    console.log(`Updated semantic index for ${nodeId}`);
  }

  calculateSemanticSimilarity(keywords, node) {
    // Simplified semantic similarity
    if (!keywords || !node.description) return 0;
    
    const nodeText = (node.description + ' ' + JSON.stringify(node.content)).toLowerCase();
    const keywordList = Array.isArray(keywords) ? keywords : [keywords];
    
    let matches = 0;
    for (const keyword of keywordList) {
      if (nodeText.includes(keyword.toLowerCase())) {
        matches++;
      }
    }
    
    return matches / keywordList.length;
  }

  async updateMemoryIndex(memoryType, content, contributorId) {
    // Placeholder for memory indexing
    console.log(`Updated memory index for ${memoryType}`);
  }

  async getCurrentContext(agentId) {
    return {
      timestamp: Date.now(),
      agentId,
      currentTasks: [],
      recentActions: []
    };
  }

  calculateContentSimilarity(query, content) {
    // Simplified content similarity
    return Math.random() * 0.8 + 0.1;
  }

  calculateContextSimilarity(context1, context2) {
    // Simplified context similarity
    return Math.random() * 0.6 + 0.2;
  }

  calculateKeywordSimilarity(keywords, text) {
    // Simplified keyword similarity
    return Math.random() * 0.7 + 0.1;
  }

  applyForgettingCurve(results) {
    const now = Date.now();
    const forgetRate = this.collectiveMemory.forgettingCurve;
    
    // Apply forgetting curve to episodic memories
    for (const episode of results.episodic) {
      const age = now - episode.timestamp;
      const forgettingFactor = Math.exp(-age * forgetRate / (24 * 3600 * 1000));
      episode.relevance *= forgettingFactor;
    }
  }

  getRecentExperiences(agentId) {
    // Placeholder - would return recent experiences for the agent
    return [
      { action: 'optimize_performance', reward: 0.8, shareWorthy: true },
      { action: 'collaborate', reward: 0.6, shareWorthy: true }
    ];
  }

  getSuccessfulProcedures(agentId) {
    // Placeholder - would return successful procedures discovered by agent
    return [
      {
        name: 'efficient_task_completion',
        steps: ['analyze', 'optimize', 'execute', 'validate'],
        domain: 'performance',
        effectiveness: 0.85,
        successRate: 0.9
      }
    ];
  }

  async shareExperience(agentId, experience) {
    await this.addToCollectiveMemory('episodic', experience, agentId);
  }

  extractPattern(experience) {
    // Placeholder for pattern extraction
    if (experience.reward > 0.7) {
      return `High reward achieved through ${this.distributedRL.actionSpace[experience.action]} in similar contexts`;
    }
    return null;
  }

  inferDomain(experience) {
    const action = this.distributedRL.actionSpace[experience.action];
    if (action.includes('collaboration') || action.includes('teach')) return 'social';
    if (action.includes('optimize') || action.includes('task')) return 'performance';
    if (action.includes('learn') || action.includes('innovation')) return 'cognitive';
    return 'general';
  }

  generateLessonContent(skill, method) {
    return {
      skill,
      method,
      quality: 0.7 + Math.random() * 0.3,
      examples: Math.floor(Math.random() * 3) + 2,
      exercises: Math.floor(Math.random() * 2) + 1
    };
  }

  generatePracticeExercises(skill) {
    return [
      { type: 'practice', difficulty: 0.5, description: `Practice ${skill} basics` },
      { type: 'application', difficulty: 0.7, description: `Apply ${skill} to scenario` }
    ];
  }

  async getRelevantExamples(skill) {
    return [
      { example: `Example 1 for ${skill}`, effectiveness: 0.8 },
      { example: `Example 2 for ${skill}`, effectiveness: 0.7 }
    ];
  }

  calculateLessonDifficulty(learner, skill) {
    const learnerLevel = this.skillTransfer.expertiseMap.get(learner.id)?.get(skill) || 0;
    return Math.max(0.3, Math.min(0.9, 0.5 + (0.5 - learnerLevel)));
  }

  calculateDeliveryEffectiveness(session, lesson) {
    const mentor = this.agents.get(session.mentor);
    const learner = this.agents.get(session.learner);
    
    return (mentor.performance.efficiency * 0.4 + 
            learner.performance.adaptability * 0.3 + 
            lesson.content.quality * 0.3);
  }

  calculateTransferDifficulty(skill, mentorLevel, learnerLevel) {
    return (mentorLevel - learnerLevel) * 0.5 + Math.random() * 0.3;
  }

  identifyLearningStrengths(learner, lesson) {
    return ['quick_understanding', 'good_retention'];
  }

  identifyLearningWeaknesses(learner, lesson) {
    return ['needs_more_practice', 'difficulty_with_advanced_concepts'];
  }

  async graduateStudent(sessionId) {
    const session = this.peerTeaching.teacherStudentPairs.get(sessionId);
    if (!session) return;
    
    const learner = this.agents.get(session.learner);
    if (learner) {
      // Update learner's skill level
      const currentLevel = this.skillTransfer.expertiseMap.get(session.learner)?.get(session.skill) || 0;
      const newLevel = Math.min(1.0, currentLevel + 0.3);
      
      if (!this.skillTransfer.expertiseMap.has(session.learner)) {
        this.skillTransfer.expertiseMap.set(session.learner, new Map());
      }
      this.skillTransfer.expertiseMap.get(session.learner).set(session.skill, newLevel);
      
      // Record successful transfer
      this.skillTransfer.transferHistory.push({
        mentor: session.mentor,
        learner: session.learner,
        skill: session.skill,
        completedAt: Date.now(),
        finalLevel: newLevel,
        sessionDuration: Date.now() - session.startTime
      });
    }
    
    console.log(`Student ${session.learner} graduated from ${session.skill} training`);
    this.emit('student_graduated', { sessionId, skill: session.skill, learner: session.learner });
  }

  async executeMetaLearning() {
    console.log('Executing meta-learning cycle...');
    
    // Analyze learning effectiveness across different strategies
    const strategyEffectiveness = await this.analyzeLearninnngStrategies();
    
    // Adapt learning parameters based on performance
    await this.adaptLearningParameters(strategyEffectiveness);
    
    // Update curriculum based on collective progress
    await this.updateCurriculumLearning();
    
    // Optimize transfer learning rules
    await this.optimizeTransferRules();
    
    const results = {
      strategiesAnalyzed: strategyEffectiveness.length,
      parametersAdapted: this.getAdaptedParametersCount(),
      curriculumUpdated: this.metaLearning.curriculumLearning.length,
      transferRulesOptimized: this.metaLearning.transferLearningRules.size
    };
    
    this.emit('meta_learning_complete', results);
    
    return results;
  }

  async analyzeLearninnngStrategies() {
    const strategies = [];
    
    // Analyze different teaching methods
    const teachingMethods = ['demonstration', 'guided_practice', 'collaborative_learning', 'problem_based'];
    for (const method of teachingMethods) {
      const effectiveness = this.calculateMethodEffectiveness(method);
      strategies.push({ method, effectiveness });
    }
    
    return strategies;
  }

  calculateMethodEffectiveness(method) {
    // Analyze completed teaching sessions using this method
    let totalEffectiveness = 0;
    let count = 0;
    
    for (const [sessionId, session] of this.peerTeaching.teacherStudentPairs) {
      if (session.method === method && session.status === 'completed') {
        totalEffectiveness += session.progress;
        count++;
      }
    }
    
    return count > 0 ? totalEffectiveness / count : 0.5;
  }

  async adaptLearningParameters(strategyEffectiveness) {
    // Adjust exploration rate based on learning success
    const avgReward = this.calculateAverageReward();
    if (avgReward > 0.7) {
      this.distributedRL.explorationRate = Math.max(0.05, this.distributedRL.explorationRate * 0.9);
    } else if (avgReward < 0.3) {
      this.distributedRL.explorationRate = Math.min(0.3, this.distributedRL.explorationRate * 1.1);
    }
    
    // Adjust agent learning rates based on individual performance
    for (const [agentId, agent] of this.agents) {
      const recentPerformance = this.calculateRecentPerformance(agentId);
      if (recentPerformance > 0.8) {
        agent.learningPreferences.learningRate *= 0.95; // Reduce learning rate for stable learners
      } else if (recentPerformance < 0.4) {
        agent.learningPreferences.learningRate *= 1.05; // Increase learning rate for struggling learners
      }
    }
  }

  calculateRecentPerformance(agentId) {
    const agent = this.agents.get(agentId);
    if (!agent) return 0.5;
    
    return (agent.performance.accuracy + agent.performance.efficiency + agent.performance.adaptability) / 3;
  }

  async updateCurriculumLearning() {
    // Update learning curriculum based on collective progress
    const curriculum = [];
    
    // Basic skills first
    curriculum.push({ skill: 'basic_task_execution', prerequisite: null, difficulty: 0.3 });
    curriculum.push({ skill: 'communication', prerequisite: 'basic_task_execution', difficulty: 0.4 });
    curriculum.push({ skill: 'collaboration', prerequisite: 'communication', difficulty: 0.6 });
    curriculum.push({ skill: 'optimization', prerequisite: 'basic_task_execution', difficulty: 0.7 });
    curriculum.push({ skill: 'innovation', prerequisite: 'optimization', difficulty: 0.9 });
    
    this.metaLearning.curriculumLearning = curriculum;
  }

  async optimizeTransferRules() {
    // Create rules based on successful transfers
    for (const transfer of this.skillTransfer.transferHistory.slice(-50)) {
      const rule = {
        fromSkill: transfer.skill,
        toAgent: transfer.learner,
        conditions: {
          minMentorLevel: 0.7,
          maxLearnerLevel: 0.5,
          compatibility: 0.6
        },
        effectiveness: transfer.finalLevel
      };
      
      this.metaLearning.transferLearningRules.set(`${transfer.skill}_${transfer.learner}`, rule);
    }
  }

  getAdaptedParametersCount() {
    return this.agents.size + 1; // All agent learning rates + global exploration rate
  }

  async getKnowledgeSharingStats() {
    return {
      knowledgeNodes: this.knowledgeGraph.nodes.size,
      relationships: this.knowledgeGraph.edges.size,
      recentShares: this.collectiveMemory.episodicMemory.filter(
        e => Date.now() - e.timestamp < 3600000
      ).length
    };
  }

  async storeLearningResults(results) {
    const timestamp = Date.now();
    const key = `swarm:learning:results:${timestamp}`;
    
    await this.redis.setex(key, 3600, JSON.stringify({
      timestamp,
      results,
      systemState: {
        agents: this.agents.size,
        knowledgeNodes: this.knowledgeGraph.nodes.size,
        skillTransfers: this.skillTransfer.transferHistory.length,
        memorySize: results.memorySize
      }
    }));
    
    // Update persistent learning history
    await this.redis.setex('swarm:learning:history', 3600 * 24, JSON.stringify({
      experienceReplay: this.distributedRL.experienceReplay.slice(-1000),
      knowledgeNodes: Array.from(this.knowledgeGraph.nodes.entries()),
      skillLibrary: Array.from(this.skillTransfer.skillLibrary.entries()),
      semanticMemory: Array.from(this.collectiveMemory.semanticMemory.entries()),
      lastUpdate: timestamp
    }));
  }

  // Status and monitoring
  async getStatus() {
    return {
      status: 'operational',
      registeredAgents: this.agents.size,
      learningComponents: {
        distributedRL: {
          experienceReplaySize: this.distributedRL.experienceReplay.length,
          agentModels: this.distributedRL.agentModels.size,
          globalModelUpdates: this.distributedRL.globalModel?.updateCount || 0
        },
        knowledgeGraph: {
          nodes: this.knowledgeGraph.nodes.size,
          edges: this.knowledgeGraph.edges.size,
          cachedQueries: this.knowledgeGraph.queryCache.size
        },
        skillTransfer: {
          mentors: this.skillTransfer.mentorAgents.size,
          learners: this.skillTransfer.learnerAgents.size,
          skillLibrary: this.skillTransfer.skillLibrary.size,
          transferHistory: this.skillTransfer.transferHistory.length
        },
        collectiveMemory: {
          episodic: this.collectiveMemory.episodicMemory.length,
          semantic: this.collectiveMemory.semanticMemory.size,
          procedural: this.collectiveMemory.proceduralMemory.size
        },
        peerTeaching: {
          activeSessions: this.peerTeaching.teacherStudentPairs.size,
          teachingMaterials: this.peerTeaching.teachingMaterials.size
        }
      },
      metaLearning: {
        strategies: this.metaLearning.learningStrategies.size,
        curriculum: this.metaLearning.curriculumLearning.length,
        transferRules: this.metaLearning.transferLearningRules.size
      },
      performance: {
        averageReward: this.calculateAverageReward(),
        explorationRate: this.distributedRL.explorationRate,
        knowledgeSharingRate: this.collectiveMemory.episodicMemory.length / Math.max(1, this.agents.size)
      }
    };
  }

  async cleanup() {
    if (this.learningTimer) {
      clearInterval(this.learningTimer);
    }
    
    if (this.metaLearningTimer) {
      clearInterval(this.metaLearningTimer);
    }
    
    if (this.redis) {
      await this.redis.quit();
    }
  }
}

module.exports = CrossAgentLearningSystem;