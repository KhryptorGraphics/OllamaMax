/**
 * SPARC Orchestration Patterns
 * Common patterns for multi-agent coordination
 */

export const OrchestrationPatterns = {
  /**
   * Hierarchical Coordination Pattern
   * Master agent delegates to specialized workers
   */
  hierarchical: {
    structure: {
      master: 'orchestrator',
      workers: ['researcher', 'coder', 'tester', 'reviewer'],
      communication: 'top-down'
    },
    
    async execute(task, orchestrator) {
      // Master analyzes and decomposes
      const plan = await orchestrator.spawnAgent('planner');
      const subtasks = await orchestrator.decomposeTask(task, 'domain');
      
      // Delegate to workers
      const workers = await Promise.all(
        subtasks.map(st => orchestrator.spawnAgent(st.agent))
      );
      
      // Master coordinates execution
      const results = [];
      for (const [idx, subtask] of subtasks.entries()) {
        const result = await orchestrator.executeAgentTask(workers[idx], subtask);
        results.push(result);
        
        // Master reviews each result
        await orchestrator.shareMemory(`review-${idx}`, result);
      }
      
      return orchestrator.aggregateResults(results);
    }
  },

  /**
   * Parallel Pipeline Pattern
   * Multiple parallel streams with synchronization points
   */
  parallelPipeline: {
    structure: {
      streams: [
        ['analyzer', 'designer', 'implementer'],
        ['researcher', 'documenter', 'validator'],
        ['security', 'performance', 'optimizer']
      ],
      syncPoints: [0.33, 0.66, 1.0]
    },
    
    async execute(task, orchestrator) {
      const streams = this.structure.streams;
      
      // Execute streams in parallel
      const streamPromises = streams.map(async (stream) => {
        const results = [];
        
        for (const agentType of stream) {
          const agent = await orchestrator.spawnAgent(agentType);
          const subtask = { agent: agentType, work: `${agentType} processing: ${task}` };
          const result = await orchestrator.executeAgentTask(agent, subtask);
          results.push(result);
          
          // Share intermediate results
          await orchestrator.shareMemory(`${agentType}-output`, result);
        }
        
        return results;
      });
      
      const allResults = await Promise.all(streamPromises);
      
      // Synchronize at checkpoints
      return orchestrator.synthesizeResults(allResults.flat());
    }
  },

  /**
   * Event-Driven Pattern
   * Agents react to events and trigger other agents
   */
  eventDriven: {
    structure: {
      events: ['task-created', 'code-changed', 'test-failed', 'review-needed'],
      handlers: {
        'task-created': ['analyzer', 'planner'],
        'code-changed': ['tester', 'reviewer'],
        'test-failed': ['debugger', 'fixer'],
        'review-needed': ['reviewer', 'optimizer']
      }
    },
    
    async execute(task, orchestrator) {
      const eventQueue = [];
      const results = [];
      
      // Initial event
      eventQueue.push({ type: 'task-created', data: task });
      
      while (eventQueue.length > 0) {
        const event = eventQueue.shift();
        const handlers = this.structure.handlers[event.type] || [];
        
        // Spawn handlers for event
        const handlerPromises = handlers.map(async (agentType) => {
          const agent = await orchestrator.spawnAgent(agentType);
          const subtask = { 
            agent: agentType, 
            work: `Handle ${event.type}: ${event.data}` 
          };
          
          const result = await orchestrator.executeAgentTask(agent, subtask);
          
          // Generate new events based on results
          if (result.success && agentType === 'coder') {
            eventQueue.push({ type: 'code-changed', data: result.output });
          } else if (!result.success && agentType === 'tester') {
            eventQueue.push({ type: 'test-failed', data: result.error });
          }
          
          return result;
        });
        
        const handlerResults = await Promise.all(handlerPromises);
        results.push(...handlerResults);
      }
      
      return orchestrator.aggregateResults(results);
    }
  },

  /**
   * Adaptive Strategy Pattern
   * Dynamically adjusts strategy based on task complexity
   */
  adaptive: {
    structure: {
      strategies: {
        simple: ['coder', 'tester'],
        moderate: ['researcher', 'architect', 'coder', 'tester'],
        complex: ['analyzer', 'architect', 'designer', 'coder', 'tester', 'optimizer', 'reviewer']
      },
      thresholds: {
        simple: 0.3,
        moderate: 0.7,
        complex: 1.0
      }
    },
    
    async execute(task, orchestrator) {
      // Assess complexity
      const complexity = orchestrator.assessComplexity(task);
      
      // Select strategy
      let strategy = 'simple';
      if (complexity > this.structure.thresholds.moderate) {
        strategy = 'complex';
      } else if (complexity > this.structure.thresholds.simple) {
        strategy = 'moderate';
      }
      
      // Execute with selected strategy
      const agentTypes = this.structure.strategies[strategy];
      const agents = await Promise.all(
        agentTypes.map(type => orchestrator.spawnAgent(type))
      );
      
      // Adaptive execution
      const results = [];
      for (const [idx, agent] of agents.entries()) {
        const subtask = {
          agent: agent.type,
          work: `${agent.type} handling ${strategy} task: ${task}`
        };
        
        // Pass context from previous agent
        if (idx > 0) {
          const context = await orchestrator.getMemory(`agent-${idx-1}-output`);
          subtask.context = context;
        }
        
        const result = await orchestrator.executeAgentTask(agent, subtask);
        await orchestrator.shareMemory(`agent-${idx}-output`, result);
        results.push(result);
      }
      
      return orchestrator.aggregateResults(results);
    }
  },

  /**
   * Consensus Building Pattern
   * Multiple agents work together to reach consensus
   */
  consensus: {
    structure: {
      validators: ['reviewer', 'tester', 'security'],
      threshold: 0.66, // 2/3 majority
      maxRounds: 3
    },
    
    async execute(task, orchestrator) {
      const validators = await Promise.all(
        this.structure.validators.map(type => orchestrator.spawnAgent(type))
      );
      
      let consensus = false;
      let round = 0;
      let finalResult = null;
      
      while (!consensus && round < this.structure.maxRounds) {
        round++;
        
        // Each validator evaluates
        const evaluations = await Promise.all(
          validators.map(async (validator) => {
            const subtask = {
              agent: validator.type,
              work: `Evaluate round ${round}: ${task}`
            };
            
            return orchestrator.executeAgentTask(validator, subtask);
          })
        );
        
        // Check for consensus
        const approvals = evaluations.filter(e => e.success).length;
        const approvalRate = approvals / validators.length;
        
        if (approvalRate >= this.structure.threshold) {
          consensus = true;
          finalResult = orchestrator.synthesizeResults(evaluations);
        } else {
          // Share feedback for next round
          await orchestrator.shareMemory(`round-${round}-feedback`, evaluations);
        }
      }
      
      return finalResult || { consensus: false, rounds: round };
    }
  },

  /**
   * MapReduce Pattern
   * Distribute work across multiple agents then aggregate
   */
  mapReduce: {
    structure: {
      mappers: 5,
      reducers: 2,
      partitioner: 'hash'
    },
    
    async execute(task, orchestrator) {
      // Split task into chunks
      const chunks = this.partitionTask(task, this.structure.mappers);
      
      // Map phase - parallel processing
      const mappers = await Promise.all(
        Array(this.structure.mappers).fill(null).map(() => 
          orchestrator.spawnAgent('processor')
        )
      );
      
      const mapResults = await Promise.all(
        chunks.map(async (chunk, idx) => {
          const subtask = {
            agent: mappers[idx].type,
            work: `Process chunk ${idx}: ${chunk}`
          };
          
          return orchestrator.executeAgentTask(mappers[idx], subtask);
        })
      );
      
      // Shuffle and sort
      const shuffled = this.shuffle(mapResults);
      
      // Reduce phase - aggregate results
      const reducers = await Promise.all(
        Array(this.structure.reducers).fill(null).map(() => 
          orchestrator.spawnAgent('aggregator')
        )
      );
      
      const reduceResults = await Promise.all(
        shuffled.map(async (data, idx) => {
          const reducerIdx = idx % this.structure.reducers;
          const subtask = {
            agent: reducers[reducerIdx].type,
            work: `Reduce partition ${idx}: ${data.length} items`
          };
          
          return orchestrator.executeAgentTask(reducers[reducerIdx], subtask);
        })
      );
      
      // Final aggregation
      return orchestrator.synthesizeResults(reduceResults);
    },
    
    partitionTask(task, numPartitions) {
      // Simple string splitting for demo
      const words = task.split(' ');
      const chunkSize = Math.ceil(words.length / numPartitions);
      const chunks = [];
      
      for (let i = 0; i < numPartitions; i++) {
        const start = i * chunkSize;
        const end = Math.min(start + chunkSize, words.length);
        chunks.push(words.slice(start, end).join(' '));
      }
      
      return chunks;
    },
    
    shuffle(mapResults) {
      // Group by key (simplified)
      const groups = {};
      
      mapResults.forEach(result => {
        const key = result.output?.key || 'default';
        if (!groups[key]) groups[key] = [];
        groups[key].push(result);
      });
      
      return Object.values(groups);
    }
  }
};

/**
 * Pattern selector based on task characteristics
 */
export function selectPattern(task, preferences = {}) {
  const characteristics = analyzeTask(task);
  
  if (preferences.pattern) {
    return OrchestrationPatterns[preferences.pattern];
  }
  
  // Auto-select based on characteristics
  if (characteristics.requiresConsensus) {
    return OrchestrationPatterns.consensus;
  } else if (characteristics.isDataIntensive) {
    return OrchestrationPatterns.mapReduce;
  } else if (characteristics.hasEvents) {
    return OrchestrationPatterns.eventDriven;
  } else if (characteristics.complexity > 0.7) {
    return OrchestrationPatterns.adaptive;
  } else if (characteristics.isParallelizable) {
    return OrchestrationPatterns.parallelPipeline;
  } else {
    return OrchestrationPatterns.hierarchical;
  }
}

/**
 * Analyze task characteristics
 */
function analyzeTask(task) {
  const taskLower = task.toLowerCase();
  
  return {
    complexity: assessComplexity(task),
    isParallelizable: taskLower.includes('parallel') || taskLower.includes('concurrent'),
    requiresConsensus: taskLower.includes('review') || taskLower.includes('approve'),
    hasEvents: taskLower.includes('trigger') || taskLower.includes('event'),
    isDataIntensive: taskLower.includes('process') || taskLower.includes('analyze'),
    domains: countDomains(task)
  };
}

function assessComplexity(task) {
  let score = 0;
  
  // Length factor
  if (task.length > 100) score += 0.2;
  if (task.length > 200) score += 0.2;
  
  // Keyword factors
  const complexKeywords = ['system', 'architecture', 'integration', 'optimization'];
  complexKeywords.forEach(keyword => {
    if (task.toLowerCase().includes(keyword)) score += 0.15;
  });
  
  return Math.min(score, 1.0);
}

function countDomains(task) {
  const domains = ['frontend', 'backend', 'database', 'api', 'security'];
  return domains.filter(d => task.toLowerCase().includes(d)).length;
}

export default OrchestrationPatterns;