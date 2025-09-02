// Event-Driven Coordination System for Real-Time Performance Optimization
const EventEmitter = require('events');
const { performance } = require('perf_hooks');

class EventDrivenCoordinationSystem extends EventEmitter {
  constructor(options = {}) {
    super();
    this.options = {
      maxListeners: options.maxListeners || 1000,
      batchSize: options.batchSize || 10,
      batchTimeout: options.batchTimeout || 50, // ms
      priorityLevels: options.priorityLevels || ['low', 'normal', 'high', 'critical'],
      deadLetterQueueSize: options.deadLetterQueueSize || 1000,
      retryAttempts: options.retryAttempts || 3,
      ...options
    };

    // Core coordination components
    this.eventBus = new EventEmitter();
    this.eventBus.setMaxListeners(this.options.maxListeners);
    
    // Event processing queues
    this.priorityQueues = new Map();
    this.batchProcessors = new Map();
    this.eventHandlers = new Map();
    this.eventMetrics = new Map();
    
    // Coordination state management
    this.coordinationState = new Map();
    this.agentRegistry = new Map();
    this.taskRegistry = new Map();
    this.subscriptions = new Map();
    
    // Performance tracking
    this.metrics = {
      eventsProcessed: 0,
      eventsDropped: 0,
      avgProcessingTime: 0,
      queueSizes: {},
      throughput: 0,
      errorRate: 0
    };

    this.initialize();
  }

  // Initialize the coordination system
  initialize() {
    console.log('🚀 Initializing Event-Driven Coordination System...');
    
    // Initialize priority queues
    this.options.priorityLevels.forEach(level => {
      this.priorityQueues.set(level, []);
      this.batchProcessors.set(level, null);
    });

    // Set up core event handlers
    this.setupCoreEventHandlers();
    
    // Start batch processing
    this.startBatchProcessing();
    
    // Start metrics collection
    this.startMetricsCollection();
    
    console.log('✅ Event-Driven Coordination System initialized');
    this.emit('system:ready');
  }

  // Set up core coordination event handlers
  setupCoreEventHandlers() {
    // Agent lifecycle events
    this.registerHandler('agent:spawn', this.handleAgentSpawn.bind(this));
    this.registerHandler('agent:ready', this.handleAgentReady.bind(this));
    this.registerHandler('agent:busy', this.handleAgentBusy.bind(this));
    this.registerHandler('agent:idle', this.handleAgentIdle.bind(this));
    this.registerHandler('agent:error', this.handleAgentError.bind(this));
    this.registerHandler('agent:terminate', this.handleAgentTerminate.bind(this));

    // Task coordination events
    this.registerHandler('task:create', this.handleTaskCreate.bind(this));
    this.registerHandler('task:assign', this.handleTaskAssign.bind(this));
    this.registerHandler('task:start', this.handleTaskStart.bind(this));
    this.registerHandler('task:progress', this.handleTaskProgress.bind(this));
    this.registerHandler('task:complete', this.handleTaskComplete.bind(this));
    this.registerHandler('task:error', this.handleTaskError.bind(this));

    // Coordination events
    this.registerHandler('swarm:init', this.handleSwarmInit.bind(this));
    this.registerHandler('swarm:scale', this.handleSwarmScale.bind(this));
    this.registerHandler('swarm:optimize', this.handleSwarmOptimize.bind(this));
    this.registerHandler('coordination:sync', this.handleCoordinationSync.bind(this));

    // System events
    this.registerHandler('system:resource-alert', this.handleResourceAlert.bind(this));
    this.registerHandler('system:performance-alert', this.handlePerformanceAlert.bind(this));
    this.registerHandler('system:health-check', this.handleHealthCheck.bind(this));

    console.log('📋 Core event handlers registered');
  }

  // Register event handler with priority support
  registerHandler(eventType, handler, priority = 'normal', options = {}) {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, []);
    }

    const handlerInfo = {
      handler,
      priority,
      options,
      registeredAt: Date.now(),
      callCount: 0,
      totalExecutionTime: 0,
      errors: 0
    };

    this.eventHandlers.get(eventType).push(handlerInfo);
    
    // Sort handlers by priority
    this.eventHandlers.get(eventType).sort((a, b) => {
      const priorityOrder = { critical: 4, high: 3, normal: 2, low: 1 };
      return priorityOrder[b.priority] - priorityOrder[a.priority];
    });

    console.log(`📝 Registered handler for ${eventType} with ${priority} priority`);
  }

  // Emit event with priority and batching support
  async emitCoordinationEvent(eventType, data, priority = 'normal', options = {}) {
    const event = {
      id: `evt-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      type: eventType,
      data,
      priority,
      timestamp: Date.now(),
      source: options.source || 'unknown',
      correlation: options.correlation || null,
      retryCount: 0,
      ...options
    };

    console.log(`📨 Emitting event: ${eventType} (${priority}) [${event.id}]`);

    // Add to appropriate priority queue
    const queue = this.priorityQueues.get(priority);
    if (queue) {
      queue.push(event);
      
      // Trigger immediate processing for critical events
      if (priority === 'critical') {
        setImmediate(() => this.processPriorityQueue(priority));
      }
    } else {
      console.warn(`⚠️ Unknown priority level: ${priority}, using normal`);
      this.priorityQueues.get('normal').push(event);
    }

    // Update queue size metrics
    this.updateQueueMetrics();

    return event.id;
  }

  // Process priority queue with batching
  async processPriorityQueue(priority) {
    const queue = this.priorityQueues.get(priority);
    if (!queue || queue.length === 0) return;

    console.log(`⚡ Processing ${priority} queue: ${queue.length} events`);

    // Process events in batches
    const batchSize = priority === 'critical' ? 1 : this.options.batchSize;
    const batch = queue.splice(0, Math.min(batchSize, queue.length));

    await this.processBatch(batch);
  }

  // Process a batch of events
  async processBatch(events) {
    const batchStartTime = performance.now();
    const results = [];

    // Group events by type for efficient processing
    const eventGroups = new Map();
    events.forEach(event => {
      if (!eventGroups.has(event.type)) {
        eventGroups.set(event.type, []);
      }
      eventGroups.get(event.type).push(event);
    });

    // Process each event type group
    for (const [eventType, eventGroup] of eventGroups) {
      const groupResults = await this.processEventGroup(eventType, eventGroup);
      results.push(...groupResults);
    }

    const batchEndTime = performance.now();
    const batchProcessingTime = batchEndTime - batchStartTime;

    console.log(`✅ Processed batch of ${events.length} events in ${batchProcessingTime.toFixed(2)}ms`);
    
    // Update metrics
    this.metrics.eventsProcessed += events.length;
    this.updateProcessingTimeMetrics(batchProcessingTime);

    return results;
  }

  // Process events of the same type together
  async processEventGroup(eventType, events) {
    const handlers = this.eventHandlers.get(eventType) || [];
    if (handlers.length === 0) {
      console.warn(`⚠️ No handlers registered for event type: ${eventType}`);
      return events.map(e => ({ eventId: e.id, status: 'no-handler', error: 'No registered handlers' }));
    }

    console.log(`🔄 Processing ${events.length} ${eventType} events with ${handlers.length} handlers`);

    const results = [];
    
    // Process events through each handler
    for (const handlerInfo of handlers) {
      const handlerStartTime = performance.now();
      
      try {
        // Process all events of this type with this handler
        const handlerResults = await Promise.allSettled(
          events.map(event => this.executeHandler(handlerInfo, event))
        );

        handlerResults.forEach((result, index) => {
          const event = events[index];
          if (result.status === 'fulfilled') {
            results.push({
              eventId: event.id,
              eventType: event.type,
              status: 'success',
              result: result.value,
              handler: handlerInfo.handler.name || 'anonymous'
            });
          } else {
            results.push({
              eventId: event.id,
              eventType: event.type,
              status: 'error',
              error: result.reason,
              handler: handlerInfo.handler.name || 'anonymous'
            });
            handlerInfo.errors++;
            console.error(`❌ Handler error for ${event.type}:`, result.reason);
          }
        });

        const handlerEndTime = performance.now();
        handlerInfo.totalExecutionTime += (handlerEndTime - handlerStartTime);
        handlerInfo.callCount += events.length;

      } catch (error) {
        console.error(`❌ Fatal handler error for ${eventType}:`, error);
        handlerInfo.errors++;
        
        events.forEach(event => {
          results.push({
            eventId: event.id,
            eventType: event.type,
            status: 'fatal-error',
            error: error.message,
            handler: handlerInfo.handler.name || 'anonymous'
          });
        });
      }
    }

    return results;
  }

  // Execute individual handler with timeout and retry
  async executeHandler(handlerInfo, event) {
    const { handler, options } = handlerInfo;
    const timeout = options.timeout || 5000; // 5 second default timeout

    return new Promise(async (resolve, reject) => {
      const timeoutHandle = setTimeout(() => {
        reject(new Error(`Handler timeout after ${timeout}ms`));
      }, timeout);

      try {
        const result = await handler(event);
        clearTimeout(timeoutHandle);
        resolve(result);
      } catch (error) {
        clearTimeout(timeoutHandle);
        
        // Retry logic for retryable errors
        if (event.retryCount < this.options.retryAttempts && this.isRetryableError(error)) {
          event.retryCount++;
          console.warn(`🔄 Retrying event ${event.id} (attempt ${event.retryCount})`);
          
          // Add back to queue with exponential backoff
          setTimeout(() => {
            this.priorityQueues.get(event.priority).push(event);
          }, Math.pow(2, event.retryCount) * 1000);
          
          resolve({ status: 'retrying', attempt: event.retryCount });
        } else {
          reject(error);
        }
      }
    });
  }

  // Check if error is retryable
  isRetryableError(error) {
    const retryableErrors = [
      'TIMEOUT',
      'NETWORK_ERROR',
      'TEMPORARY_FAILURE',
      'RESOURCE_BUSY'
    ];
    
    return retryableErrors.some(type => 
      error.message.includes(type) || error.code === type
    );
  }

  // Start batch processing for all priority levels
  startBatchProcessing() {
    this.options.priorityLevels.forEach(priority => {
      const interval = this.getBatchInterval(priority);
      
      const processor = setInterval(async () => {
        await this.processPriorityQueue(priority);
      }, interval);
      
      this.batchProcessors.set(priority, processor);
      console.log(`⏱️ Started batch processor for ${priority} queue (${interval}ms interval)`);
    });
  }

  // Get batch processing interval based on priority
  getBatchInterval(priority) {
    const intervals = {
      critical: 10,   // 10ms - immediate processing
      high: 25,       // 25ms - very fast
      normal: 50,     // 50ms - standard
      low: 100        // 100ms - slower for low priority
    };
    
    return intervals[priority] || intervals.normal;
  }

  // Core event handlers implementation
  async handleAgentSpawn(event) {
    const { agentId, agentType, capabilities } = event.data;
    
    console.log(`🤖 Agent spawn: ${agentId} (${agentType})`);
    
    this.agentRegistry.set(agentId, {
      id: agentId,
      type: agentType,
      capabilities,
      status: 'spawning',
      spawnedAt: Date.now(),
      taskCount: 0,
      lastActivity: Date.now()
    });

    // Emit follow-up events
    await this.emitCoordinationEvent('agent:spawning', { agentId }, 'high');
    
    return { status: 'registered', agentId };
  }

  async handleAgentReady(event) {
    const { agentId } = event.data;
    const agent = this.agentRegistry.get(agentId);
    
    if (agent) {
      agent.status = 'ready';
      agent.readyAt = Date.now();
      console.log(`✅ Agent ready: ${agentId}`);
      
      // Trigger task assignment if tasks are waiting
      await this.emitCoordinationEvent('coordination:assign-tasks', { agentId }, 'normal');
    }
    
    return { status: 'ready', agentId };
  }

  async handleTaskCreate(event) {
    const { taskId, task, priority, requirements } = event.data;
    
    console.log(`📋 Task created: ${taskId} (${priority})`);
    
    this.taskRegistry.set(taskId, {
      id: taskId,
      task,
      priority,
      requirements,
      status: 'created',
      createdAt: Date.now(),
      assignedAgent: null,
      startedAt: null,
      completedAt: null
    });

    // Trigger task assignment
    await this.emitCoordinationEvent('task:assign', { taskId }, priority);
    
    return { status: 'created', taskId };
  }

  async handleTaskAssign(event) {
    const { taskId } = event.data;
    const task = this.taskRegistry.get(taskId);
    
    if (!task) {
      console.warn(`⚠️ Task not found for assignment: ${taskId}`);
      return { status: 'error', error: 'Task not found' };
    }

    // Find suitable agent
    const suitableAgent = this.findSuitableAgent(task.requirements);
    
    if (suitableAgent) {
      task.assignedAgent = suitableAgent.id;
      task.status = 'assigned';
      task.assignedAt = Date.now();
      
      suitableAgent.status = 'assigned';
      suitableAgent.currentTask = taskId;
      
      console.log(`🎯 Task assigned: ${taskId} -> Agent ${suitableAgent.id}`);
      
      // Emit task start event
      await this.emitCoordinationEvent('task:start', { taskId, agentId: suitableAgent.id }, task.priority);
      
      return { status: 'assigned', taskId, agentId: suitableAgent.id };
    } else {
      console.warn(`⚠️ No suitable agent found for task: ${taskId}`);
      
      // Re-queue for later assignment
      setTimeout(() => {
        this.emitCoordinationEvent('task:assign', { taskId }, 'normal');
      }, 5000);
      
      return { status: 'queued', message: 'No suitable agent available' };
    }
  }

  async handleSwarmInit(event) {
    const { topology, maxAgents, strategy } = event.data;
    
    console.log(`🐝 Initializing swarm: ${topology} topology, max ${maxAgents} agents`);
    
    const swarmId = `swarm-${Date.now()}`;
    this.coordinationState.set(swarmId, {
      topology,
      maxAgents,
      strategy,
      agents: [],
      createdAt: Date.now(),
      status: 'initializing'
    });

    return { status: 'initialized', swarmId };
  }

  async handleCoordinationSync(event) {
    const { swarmId } = event.data;
    
    console.log(`🔄 Synchronizing coordination for swarm: ${swarmId}`);
    
    const swarm = this.coordinationState.get(swarmId);
    if (swarm) {
      swarm.lastSync = Date.now();
      swarm.syncCount = (swarm.syncCount || 0) + 1;
      
      // Perform coordination optimizations
      await this.optimizeSwarmCoordination(swarm);
    }

    return { status: 'synchronized', swarmId };
  }

  // Agent matching algorithm
  findSuitableAgent(requirements) {
    const availableAgents = Array.from(this.agentRegistry.values())
      .filter(agent => agent.status === 'ready');

    if (availableAgents.length === 0) return null;

    // Score agents based on capability match
    const scoredAgents = availableAgents.map(agent => ({
      agent,
      score: this.calculateAgentScore(agent, requirements)
    }));

    // Return best scoring agent
    scoredAgents.sort((a, b) => b.score - a.score);
    return scoredAgents[0].score > 0 ? scoredAgents[0].agent : null;
  }

  // Calculate agent suitability score
  calculateAgentScore(agent, requirements) {
    if (!requirements || requirements.length === 0) return 1;
    
    const matches = requirements.filter(req => 
      agent.capabilities.includes(req)
    ).length;
    
    const baseScore = matches / requirements.length;
    const experienceBonus = Math.min(agent.taskCount * 0.1, 0.5);
    const recencyBonus = (Date.now() - agent.lastActivity) < 300000 ? 0.2 : 0;
    
    return baseScore + experienceBonus + recencyBonus;
  }

  // Swarm coordination optimization
  async optimizeSwarmCoordination(swarm) {
    // Implement intelligent load balancing
    const agents = swarm.agents;
    if (agents.length === 0) return;

    // Analyze workload distribution
    const workloadAnalysis = this.analyzeWorkloadDistribution(agents);
    
    // Apply optimization based on topology
    switch (swarm.topology) {
      case 'hierarchical':
        await this.optimizeHierarchicalTopology(swarm, workloadAnalysis);
        break;
      case 'mesh':
        await this.optimizeMeshTopology(swarm, workloadAnalysis);
        break;
      case 'adaptive':
        await this.optimizeAdaptiveTopology(swarm, workloadAnalysis);
        break;
    }
  }

  // Additional handler stubs (implement based on specific requirements)
  async handleAgentBusy(event) { return { status: 'handled' }; }
  async handleAgentIdle(event) { return { status: 'handled' }; }
  async handleAgentError(event) { return { status: 'handled' }; }
  async handleAgentTerminate(event) { return { status: 'handled' }; }
  async handleTaskStart(event) { return { status: 'handled' }; }
  async handleTaskProgress(event) { return { status: 'handled' }; }
  async handleTaskComplete(event) { return { status: 'handled' }; }
  async handleTaskError(event) { return { status: 'handled' }; }
  async handleSwarmScale(event) { return { status: 'handled' }; }
  async handleSwarmOptimize(event) { return { status: 'handled' }; }
  async handleResourceAlert(event) { return { status: 'handled' }; }
  async handlePerformanceAlert(event) { return { status: 'handled' }; }
  async handleHealthCheck(event) { return { status: 'handled' }; }

  // Metrics and monitoring
  startMetricsCollection() {
    setInterval(() => {
      this.updateQueueMetrics();
      this.calculateThroughput();
      this.emit('metrics:updated', this.getMetrics());
    }, 5000); // Every 5 seconds
  }

  updateQueueMetrics() {
    this.options.priorityLevels.forEach(level => {
      const queue = this.priorityQueues.get(level);
      this.metrics.queueSizes[level] = queue ? queue.length : 0;
    });
  }

  updateProcessingTimeMetrics(processingTime) {
    // Calculate weighted average
    const totalProcessed = this.metrics.eventsProcessed;
    this.metrics.avgProcessingTime = 
      (this.metrics.avgProcessingTime * (totalProcessed - 1) + processingTime) / totalProcessed;
  }

  calculateThroughput() {
    // Calculate events per second over last period
    const now = Date.now();
    if (!this.lastThroughputCheck) {
      this.lastThroughputCheck = now;
      this.lastEventCount = this.metrics.eventsProcessed;
      return;
    }

    const timeDiff = now - this.lastThroughputCheck;
    const eventDiff = this.metrics.eventsProcessed - this.lastEventCount;
    
    this.metrics.throughput = (eventDiff / timeDiff) * 1000; // events per second
    
    this.lastThroughputCheck = now;
    this.lastEventCount = this.metrics.eventsProcessed;
  }

  getMetrics() {
    return {
      ...this.metrics,
      queueSizes: { ...this.metrics.queueSizes },
      activeAgents: this.agentRegistry.size,
      activeTasks: this.taskRegistry.size,
      handlerStats: this.getHandlerStats(),
      systemUptime: Date.now() - (this.startTime || Date.now())
    };
  }

  getHandlerStats() {
    const stats = {};
    
    for (const [eventType, handlers] of this.eventHandlers) {
      stats[eventType] = handlers.map(h => ({
        priority: h.priority,
        callCount: h.callCount,
        avgExecutionTime: h.callCount > 0 ? h.totalExecutionTime / h.callCount : 0,
        errorRate: h.callCount > 0 ? (h.errors / h.callCount * 100).toFixed(2) + '%' : '0%'
      }));
    }
    
    return stats;
  }

  // Graceful shutdown
  async shutdown() {
    console.log('🛑 Shutting down Event-Driven Coordination System...');
    
    // Stop batch processors
    for (const processor of this.batchProcessors.values()) {
      clearInterval(processor);
    }
    
    // Process remaining events
    for (const priority of this.options.priorityLevels) {
      await this.processPriorityQueue(priority);
    }
    
    // Clear registries
    this.agentRegistry.clear();
    this.taskRegistry.clear();
    this.coordinationState.clear();
    
    console.log('✅ Event-Driven Coordination System shutdown complete');
  }
}

// Demo and testing
class CoordinationSystemDemo {
  static async runDemo() {
    console.log('🚀 Starting Event-Driven Coordination System Demo\n');
    
    const coordinator = new EventDrivenCoordinationSystem({
      batchSize: 5,
      batchTimeout: 100
    });

    // Wait for system to be ready
    await new Promise(resolve => {
      coordinator.once('system:ready', resolve);
    });

    console.log('✅ Coordination system ready\n');

    // Simulate coordination scenario
    console.log('📋 Simulating coordination workflow...');
    
    try {
      // 1. Initialize swarm
      await coordinator.emitCoordinationEvent('swarm:init', {
        topology: 'mesh',
        maxAgents: 10,
        strategy: 'balanced'
      }, 'high');

      // 2. Spawn agents
      const agentTypes = ['researcher', 'coder', 'tester', 'coordinator'];
      for (let i = 0; i < agentTypes.length; i++) {
        await coordinator.emitCoordinationEvent('agent:spawn', {
          agentId: `agent-${agentTypes[i]}-${i}`,
          agentType: agentTypes[i],
          capabilities: coordinator.getAgentCapabilities(agentTypes[i])
        }, 'normal');
      }

      // 3. Mark agents as ready
      for (let i = 0; i < agentTypes.length; i++) {
        setTimeout(() => {
          coordinator.emitCoordinationEvent('agent:ready', {
            agentId: `agent-${agentTypes[i]}-${i}`
          }, 'normal');
        }, 1000 * (i + 1));
      }

      // 4. Create tasks
      const tasks = [
        { id: 'task-research-1', task: 'Research AI trends', priority: 'high', requirements: ['research', 'analysis'] },
        { id: 'task-code-1', task: 'Implement API', priority: 'normal', requirements: ['coding', 'api-development'] },
        { id: 'task-test-1', task: 'Run test suite', priority: 'normal', requirements: ['testing', 'validation'] }
      ];

      setTimeout(() => {
        tasks.forEach((task, index) => {
          setTimeout(() => {
            coordinator.emitCoordinationEvent('task:create', task, task.priority);
          }, 500 * index);
        });
      }, 2000);

      // 5. Coordination sync
      setTimeout(() => {
        coordinator.emitCoordinationEvent('coordination:sync', {
          swarmId: 'swarm-123'
        }, 'normal');
      }, 5000);

      // Wait for processing and show results
      setTimeout(() => {
        console.log('\n📊 Final System Metrics:');
        console.table(coordinator.getMetrics());
        
        console.log('\n🤖 Agent Registry:');
        console.table(Array.from(coordinator.agentRegistry.values()));
        
        console.log('\n📋 Task Registry:');
        console.table(Array.from(coordinator.taskRegistry.values()));

        // Shutdown
        coordinator.shutdown();
      }, 8000);

    } catch (error) {
      console.error('❌ Demo failed:', error);
    }
  }

  // Helper method for demo
  static getAgentCapabilities(agentType) {
    const capabilities = {
      'researcher': ['research', 'analysis', 'investigation'],
      'coder': ['coding', 'implementation', 'debugging'],
      'tester': ['testing', 'validation', 'qa'],
      'coordinator': ['coordination', 'management', 'orchestration']
    };
    return capabilities[agentType] || ['general'];
  }
}

// Add capability helper to main class
EventDrivenCoordinationSystem.prototype.getAgentCapabilities = CoordinationSystemDemo.getAgentCapabilities;

module.exports = { EventDrivenCoordinationSystem, CoordinationSystemDemo };

// Run demo if executed directly
if (require.main === module) {
  CoordinationSystemDemo.runDemo().catch(console.error);
}