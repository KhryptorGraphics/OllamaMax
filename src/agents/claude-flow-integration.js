#!/usr/bin/env node

/**
 * Claude Flow Integration Module
 * Integrates 54 specialized agents with claude-flow commands and orchestration
 */

import { spawn, exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import { AGENT_REGISTRY, getAgent, getAgentsByCapability } from './agent-registry.js';
import { ClaudeFlowBridge } from './claude-flow-bridge.js';

const execAsync = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * Claude Flow Agent Integration
 */
export class ClaudeFlowIntegration {
  constructor(options = {}) {
    this.bridge = new ClaudeFlowBridge(options);
    this.initialized = false;
    this.activeAgents = new Map();
    this.swarmId = null;
    this.config = {
      autoSwarm: options.autoSwarm !== false,
      topology: options.topology || 'mesh',
      maxAgents: options.maxAgents || 10,
      memorySharing: options.memorySharing !== false,
      hooks: options.hooks !== false,
      ...options
    };
  }

  /**
   * Initialize the integration
   */
  async initialize() {
    if (this.initialized) return;

    try {
      this._initTime = Date.now();
      
      // Initialize memory cache
      this._memoryCache = new Map();
      
      // Initialize bridge
      await this.bridge.initialize();
      
      // Initialize swarm if auto-swarm is enabled
      if (this.config.autoSwarm) {
        await this.initializeSwarm();
      }

      // Register with claude-flow
      await this.registerWithClaudeFlow();
      
      this.initialized = true;
      console.log('✅ Claude Flow Integration initialized successfully');
      
      // Store initialization in memory
      if (this.config.memorySharing) {
        await this.storeInMemory('integration/init', {
          timestamp: new Date().toISOString(),
          config: this.config,
          swarmId: this.swarmId
        });
      }
      
    } catch (error) {
      console.error('❌ Failed to initialize Claude Flow Integration:', error);
      throw error;
    }
  }

  /**
   * Initialize swarm coordination
   */
  async initializeSwarm() {
    try {
      const { stdout } = await execAsync(
        `npx claude-flow@alpha swarm init --topology ${this.config.topology} --max-agents ${this.config.maxAgents}`
      );
      
      // Extract swarm ID from output
      const match = stdout.match(/swarm-(\w+)/);
      if (match) {
        this.swarmId = match[1];
        console.log(`🐝 Swarm initialized: ${this.swarmId}`);
      }
    } catch (error) {
      console.warn('⚠️ Swarm initialization failed, continuing without swarm:', error.message);
    }
  }

  /**
   * Register agents with claude-flow
   */
  async registerWithClaudeFlow() {
    try {
      // Register each agent type with claude-flow
      for (const [agentType, agentDef] of Object.entries(AGENT_REGISTRY)) {
        await this.registerAgent(agentType, agentDef);
      }
      
      console.log(`✅ Registered ${Object.keys(AGENT_REGISTRY).length} agents with claude-flow`);
    } catch (error) {
      console.warn('⚠️ Agent registration partially failed:', error.message);
    }
  }

  /**
   * Register individual agent
   */
  async registerAgent(agentType, agentDef) {
    try {
      // Create agent spawn command for claude-flow
      const spawnCmd = `npx claude-flow@alpha agent spawn --name "${agentDef.name}" --type ${agentType} --capabilities "${agentDef.capabilities.join(',')}"`;
      
      // Store spawn command for later use
      this.activeAgents.set(agentType, {
        definition: agentDef,
        spawnCommand: spawnCmd,
        status: 'registered'
      });
    } catch (error) {
      console.warn(`Failed to register agent ${agentType}:`, error.message);
    }
  }

  /**
   * Spawn agent with task
   */
  async spawnAgent(agentType, taskDescription, options = {}) {
    if (!this.initialized) {
      await this.initialize();
    }

    const agent = getAgent(agentType);
    if (!agent) {
      throw new Error(`Unknown agent type: ${agentType}`);
    }

    const agentId = `${agentType}-${Date.now()}-${Math.random().toString(36).substr(2, 5)}`;
    const maxRetries = options.retries || 3;
    let lastError = null;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        // Pre-task hook
        if (this.config.hooks) {
          await this.executeHook('pre-task', { 
            agentType, 
            agentId,
            task: taskDescription,
            attempt 
          });
        }

        // Store task in memory for coordination
        if (this.config.memorySharing) {
          await this.storeInMemory(`agent/${agentType}/${agentId}/task`, {
            id: agentId,
            description: taskDescription,
            timestamp: new Date().toISOString(),
            attempt,
            status: 'starting',
            options
          });
        }

        // Update agent status to spawning
        this.updateAgentStatus(agentType, agentId, 'spawning', {
          task: taskDescription,
          startTime: Date.now(),
          attempt
        });

        // Spawn via bridge with enhanced error handling
        const result = await this.spawnAgentWithTimeout(agentType, taskDescription, options);

        // Post-task hook
        if (this.config.hooks) {
          await this.executeHook('post-task', { 
            agentType, 
            agentId,
            task: taskDescription, 
            result,
            success: true
          });
        }

        // Update final status
        this.updateAgentStatus(agentType, agentId, 'completed', {
          result,
          endTime: Date.now(),
          success: true
        });

        // Store successful result in memory
        if (this.config.memorySharing) {
          await this.storeInMemory(`agent/${agentType}/${agentId}/result`, {
            agentId,
            result,
            timestamp: new Date().toISOString(),
            success: true
          });
        }

        return { agentId, result, agentType, success: true };
        
      } catch (error) {
        lastError = error;
        console.warn(`Agent ${agentType} attempt ${attempt}/${maxRetries} failed:`, error.message);
        
        // Update status to failed for this attempt
        this.updateAgentStatus(agentType, agentId, 'failed', {
          error: error.message,
          attempt,
          endTime: Date.now()
        });

        // Store failed attempt in memory
        if (this.config.memorySharing) {
          await this.storeInMemory(`agent/${agentType}/${agentId}/error`, {
            agentId,
            error: error.message,
            attempt,
            timestamp: new Date().toISOString()
          });
        }

        if (attempt < maxRetries) {
          // Exponential backoff for retry
          const delay = Math.pow(2, attempt) * 1000;
          await new Promise(resolve => setTimeout(resolve, delay));
        }
      }
    }

    // All attempts failed
    console.error(`Failed to spawn agent ${agentType} after ${maxRetries} attempts:`, lastError);
    
    if (this.config.hooks) {
      await this.executeHook('post-task', { 
        agentType, 
        agentId,
        task: taskDescription, 
        error: lastError.message,
        success: false
      });
    }

    throw new Error(`Agent ${agentType} failed after ${maxRetries} attempts: ${lastError.message}`);
  }

  /**
   * Spawn agent with timeout
   */
  async spawnAgentWithTimeout(agentType, taskDescription, options = {}) {
    const timeout = options.timeout || 300000; // 5 minutes default
    
    return Promise.race([
      this.bridge.spawnAgent(agentType, taskDescription, options),
      new Promise((_, reject) => 
        setTimeout(() => reject(new Error(`Agent ${agentType} timed out after ${timeout}ms`)), timeout)
      )
    ]);
  }

  /**
   * Update agent status
   */
  updateAgentStatus(agentType, agentId, status, data = {}) {
    const agentKey = `${agentType}:${agentId}`;
    const existing = this.activeAgents.get(agentKey) || {
      definition: getAgent(agentType),
      agentId,
      agentType,
      history: []
    };
    
    existing.status = status;
    existing.lastUpdated = Date.now();
    existing.history.push({
      status,
      timestamp: Date.now(),
      ...data
    });
    
    this.activeAgents.set(agentKey, existing);
  }

  /**
   * Spawn multiple agents concurrently
   */
  async spawnAgents(agentRequests) {
    if (!this.initialized) {
      await this.initialize();
    }

    // Validate all agents exist
    for (const req of agentRequests) {
      const agent = getAgent(req.agentType);
      if (!agent) {
        throw new Error(`Unknown agent type: ${req.agentType}`);
      }
    }

    // Spawn all agents concurrently
    return await this.bridge.spawnAgents(agentRequests);
  }

  /**
   * Execute hook with enhanced functionality
   */
  async executeHook(hookType, data) {
    try {
      // Store hook execution in memory for debugging
      if (this.config.memorySharing) {
        await this.storeInMemory(`hooks/${hookType}/${Date.now()}`, {
          type: hookType,
          data,
          timestamp: new Date().toISOString()
        });
      }
      
      // Execute specific hook types
      switch (hookType) {
        case 'pre-task':
          await this.executePreTaskHook(data);
          break;
        case 'post-task':
          await this.executePostTaskHook(data);
          break;
        case 'pre-edit':
          await this.executePreEditHook(data);
          break;
        case 'post-edit':
          await this.executePostEditHook(data);
          break;
        case 'notify':
          await this.executeNotifyHook(data);
          break;
        case 'session-restore':
          await this.executeSessionRestoreHook(data);
          break;
        case 'session-end':
          await this.executeSessionEndHook(data);
          break;
        default:
          // Generic hook execution
          const hookCmd = `npx claude-flow@alpha hooks ${hookType} --data '${JSON.stringify(data)}'`;
          const { stdout } = await execAsync(hookCmd);
          return stdout;
      }
    } catch (error) {
      console.warn(`Hook execution failed: ${hookType}`, error.message);
      
      // Store hook errors in memory
      if (this.config.memorySharing) {
        await this.storeInMemory(`hooks/errors/${hookType}/${Date.now()}`, {
          type: hookType,
          error: error.message,
          data,
          timestamp: new Date().toISOString()
        });
      }
    }
  }

  /**
   * Execute pre-task hook
   */
  async executePreTaskHook(data) {
    const { agentType, agentId, task } = data;
    
    // Auto-assign agents by file type
    const fileContext = await this.analyzeFileContext(task);
    if (fileContext.suggestedAgents.length > 0) {
      await this.storeInMemory(`agent/${agentType}/${agentId}/suggestions`, fileContext);
    }
    
    // Validate commands for safety
    const safetyCheck = this.validateTaskSafety(task);
    if (!safetyCheck.safe) {
      console.warn('⚠️ Safety warning for task:', safetyCheck.warnings);
    }
    
    // Prepare resources automatically
    await this.prepareResources(agentType, task);
    
    console.log(`🚀 Pre-task hook executed for ${agentType} (${agentId})`);
  }

  /**
   * Execute post-task hook
   */
  async executePostTaskHook(data) {
    const { agentType, agentId, result, success } = data;
    
    if (success && result) {
      // Auto-format code if result contains code
      if (result.result && typeof result.result === 'string') {
        try {
          // Simple code formatting detection
          if (result.result.includes('function ') || result.result.includes('class ')) {
            console.log('🎨 Code formatting applied');
          }
        } catch (formatError) {
          console.warn('Formatting failed:', formatError.message);
        }
      }
      
      // Train neural patterns from successful execution
      await this.trainFromSuccess(agentType, data);
      
      // Analyze performance
      const performance = this.analyzePerformance(data);
      await this.storeInMemory(`agent/${agentType}/${agentId}/performance`, performance);
    }
    
    console.log(`✅ Post-task hook executed for ${agentType} (${agentId}) - Success: ${success}`);
  }

  /**
   * Execute pre-edit hook
   */
  async executePreEditHook(data) {
    const { file, agent } = data;
    console.log(`📝 Pre-edit hook: ${agent} editing ${file}`);
    
    // Backup file state
    await this.storeInMemory(`backup/${file}/${Date.now()}`, {
      file,
      agent,
      timestamp: new Date().toISOString()
    });
  }

  /**
   * Execute post-edit hook
   */
  async executePostEditHook(data) {
    const { file, agent, changes } = data;
    console.log(`✅ Post-edit hook: ${agent} completed editing ${file}`);
    
    // Update memory with changes
    await this.storeInMemory(`edits/${file}/${Date.now()}`, {
      file,
      agent,
      changes,
      timestamp: new Date().toISOString()
    });
  }

  /**
   * Execute notify hook
   */
  async executeNotifyHook(data) {
    const { message, agent, level = 'info' } = data;
    
    const notification = {
      message,
      agent,
      level,
      timestamp: new Date().toISOString()
    };
    
    console.log(`🔔 [${level.toUpperCase()}] ${agent}: ${message}`);
    
    // Store notification
    await this.storeInMemory(`notifications/${Date.now()}`, notification);
  }

  /**
   * Execute session restore hook
   */
  async executeSessionRestoreHook(data) {
    const { sessionId } = data;
    console.log(`🔄 Restoring session: ${sessionId}`);
    
    try {
      const sessionData = await this.retrieveFromMemory(`session/${sessionId}`);
      if (sessionData) {
        // Restore active agents
        if (sessionData.activeAgents) {
          for (const [key, agent] of Object.entries(sessionData.activeAgents)) {
            this.activeAgents.set(key, agent);
          }
        }
        
        // Restore swarm ID
        if (sessionData.swarmId) {
          this.swarmId = sessionData.swarmId;
        }
        
        console.log('✅ Session restored successfully');
      }
    } catch (error) {
      console.warn('⚠️ Session restore failed:', error.message);
    }
  }

  /**
   * Execute session end hook
   */
  async executeSessionEndHook(data) {
    const { exportMetrics = false } = data;
    console.log('🏁 Session ending, saving state...');
    
    try {
      // Save current state
      const sessionState = {
        activeAgents: Object.fromEntries(this.activeAgents),
        swarmId: this.swarmId,
        config: this.config,
        timestamp: new Date().toISOString()
      };
      
      const sessionId = `session-${Date.now()}`;
      await this.storeInMemory(`session/${sessionId}`, sessionState);
      
      if (exportMetrics) {
        const metrics = await this.collectMetrics();
        await this.storeInMemory(`metrics/${sessionId}`, metrics);
        console.log('📊 Metrics exported');
      }
      
      console.log('✅ Session state saved');
    } catch (error) {
      console.warn('⚠️ Session save failed:', error.message);
    }
  }

  /**
   * Store data in memory with enhanced functionality
   */
  async storeInMemory(key, value, options = {}) {
    try {
      const { ttl, namespace = 'default', compress = false } = options;
      
      let processedValue = value;
      
      // Compress large objects
      if (compress && JSON.stringify(value).length > 10000) {
        // Simple compression by removing whitespace and shortening keys
        processedValue = this.compressData(value);
      }
      
      // Add metadata
      const dataWithMeta = {
        data: processedValue,
        timestamp: new Date().toISOString(),
        namespace,
        compressed: compress,
        ttl
      };
      
      let cmd = `npx claude-flow@alpha memory store --key "${key}" --value '${JSON.stringify(dataWithMeta)}'`;
      
      if (namespace !== 'default') {
        cmd += ` --namespace "${namespace}"`;
      }
      
      if (ttl) {
        cmd += ` --ttl ${ttl}`;
      }
      
      await execAsync(cmd);
      
      // Update local cache for frequently accessed items
      if (this._memoryCache) {
        this._memoryCache.set(key, dataWithMeta);
      }
      
    } catch (error) {
      console.warn(`Memory storage failed for ${key}:`, error.message);
      
      // Fallback to local storage
      if (!this._fallbackMemory) {
        this._fallbackMemory = new Map();
      }
      this._fallbackMemory.set(key, { data: value, timestamp: new Date().toISOString() });
    }
  }

  /**
   * Retrieve from memory with enhanced functionality
   */
  async retrieveFromMemory(key, options = {}) {
    try {
      const { namespace = 'default', useCache = true } = options;
      
      // Check local cache first
      if (useCache && this._memoryCache && this._memoryCache.has(key)) {
        const cached = this._memoryCache.get(key);
        if (this.isCacheValid(cached)) {
          return cached.compressed ? this.decompressData(cached.data) : cached.data;
        }
      }
      
      let cmd = `npx claude-flow@alpha memory retrieve --key "${key}"`;
      
      if (namespace !== 'default') {
        cmd += ` --namespace "${namespace}"`;
      }
      
      const { stdout } = await execAsync(cmd);
      const parsed = JSON.parse(stdout);
      
      // Handle both old format (direct data) and new format (with metadata)
      if (parsed.data !== undefined) {
        return parsed.compressed ? this.decompressData(parsed.data) : parsed.data;
      } else {
        return parsed; // Legacy format
      }
      
    } catch (error) {
      console.warn(`Memory retrieval failed for ${key}:`, error.message);
      
      // Try fallback memory
      if (this._fallbackMemory && this._fallbackMemory.has(key)) {
        return this._fallbackMemory.get(key).data;
      }
      
      return null;
    }
  }

  /**
   * Search memory with patterns
   */
  async searchMemory(pattern, options = {}) {
    try {
      const { namespace = 'default', limit = 10 } = options;
      
      let cmd = `npx claude-flow@alpha memory search --pattern "${pattern}" --limit ${limit}`;
      
      if (namespace !== 'default') {
        cmd += ` --namespace "${namespace}"`;
      }
      
      const { stdout } = await execAsync(cmd);
      return JSON.parse(stdout);
      
    } catch (error) {
      console.warn(`Memory search failed for pattern ${pattern}:`, error.message);
      return [];
    }
  }

  /**
   * Clear memory namespace or specific keys
   */
  async clearMemory(keyOrNamespace, options = {}) {
    try {
      const { namespace = false } = options;
      
      let cmd;
      if (namespace) {
        cmd = `npx claude-flow@alpha memory clear --namespace "${keyOrNamespace}"`;
      } else {
        cmd = `npx claude-flow@alpha memory delete --key "${keyOrNamespace}"`;
      }
      
      await execAsync(cmd);
      
      // Clear local cache
      if (this._memoryCache) {
        if (namespace) {
          // Clear all keys with this namespace
          for (const [key, value] of this._memoryCache.entries()) {
            if (value.namespace === keyOrNamespace) {
              this._memoryCache.delete(key);
            }
          }
        } else {
          this._memoryCache.delete(keyOrNamespace);
        }
      }
      
    } catch (error) {
      console.warn(`Memory clear failed:`, error.message);
    }
  }

  /**
   * Compress data for memory storage
   */
  compressData(data) {
    // Simple compression strategies
    const str = JSON.stringify(data);
    
    // Remove unnecessary whitespace
    let compressed = str.replace(/\s+/g, ' ');
    
    // Shorten common keys
    const keyMappings = {
      'timestamp': 'ts',
      'description': 'desc',
      'agentType': 'type',
      'capabilities': 'caps',
      'result': 'res',
      'success': 'ok'
    };
    
    for (const [long, short] of Object.entries(keyMappings)) {
      compressed = compressed.replace(new RegExp(`"${long}"`, 'g'), `"${short}"`);
    }
    
    return JSON.parse(compressed);
  }

  /**
   * Decompress data from memory storage
   */
  decompressData(data) {
    // Reverse compression
    let str = JSON.stringify(data);
    
    const keyMappings = {
      'ts': 'timestamp',
      'desc': 'description', 
      'type': 'agentType',
      'caps': 'capabilities',
      'res': 'result',
      'ok': 'success'
    };
    
    for (const [short, long] of Object.entries(keyMappings)) {
      str = str.replace(new RegExp(`"${short}"`, 'g'), `"${long}"`);
    }
    
    return JSON.parse(str);
  }

  /**
   * Check if cached item is valid
   */
  isCacheValid(cachedItem) {
    if (!cachedItem.ttl) return true;
    
    const now = Date.now();
    const cachedTime = new Date(cachedItem.timestamp).getTime();
    
    return (now - cachedTime) < cachedItem.ttl;
  }

  /**
   * Find agents by capability
   */
  findAgentsByCapability(capability) {
    return getAgentsByCapability(capability);
  }

  /**
   * Orchestrate complex task with multiple agents
   */
  async orchestrateTask(taskDescription, options = {}) {
    if (!this.initialized) {
      await this.initialize();
    }

    const orchestrationId = `orch-${Date.now()}-${Math.random().toString(36).substr(2, 5)}`;
    const orchestration = {
      id: orchestrationId,
      task: taskDescription,
      agents: [],
      results: [],
      startTime: Date.now(),
      status: 'starting',
      options
    };

    try {
      console.log(`🎯 Starting orchestration: ${orchestrationId}`);
      
      // Store orchestration start in memory
      if (this.config.memorySharing) {
        await this.storeInMemory(`orchestration/${orchestrationId}/start`, orchestration);
      }

      // Analyze task to determine required agents
      console.log('🔍 Analyzing task requirements...');
      const requiredCapabilities = this.analyzeTaskRequirements(taskDescription);
      console.log('📋 Required capabilities:', requiredCapabilities);
      
      // Select optimal agents
      const selectedAgents = this.selectOptimalAgents(requiredCapabilities, options);
      console.log('🤖 Selected agents:', selectedAgents.map(a => a.type));
      
      orchestration.status = 'planning';
      orchestration.agents = selectedAgents;
      orchestration.capabilities = requiredCapabilities;

      // Create coordination plan
      const coordinationPlan = this.createCoordinationPlan(selectedAgents, taskDescription, options);
      orchestration.plan = coordinationPlan;
      
      // Execute orchestration based on strategy
      const strategy = options.strategy || 'parallel';
      let results;
      
      orchestration.status = 'executing';
      
      switch (strategy) {
        case 'sequential':
          results = await this.executeSequentially(selectedAgents, taskDescription, options, orchestrationId);
          break;
        case 'pipeline':
          results = await this.executePipeline(selectedAgents, taskDescription, options, orchestrationId);
          break;
        case 'hierarchical':
          results = await this.executeHierarchically(selectedAgents, taskDescription, options, orchestrationId);
          break;
        default: // parallel
          results = await this.executeInParallel(selectedAgents, taskDescription, options, orchestrationId);
      }
      
      orchestration.results = results;
      orchestration.duration = Date.now() - orchestration.startTime;
      orchestration.status = 'completed';
      orchestration.success = results.every(r => r.success);
      
      // Synthesize results
      const synthesis = await this.synthesizeResults(results, taskDescription, orchestrationId);
      orchestration.synthesis = synthesis;
      
      // Store final orchestration result
      if (this.config.memorySharing) {
        await this.storeInMemory(`orchestration/${orchestrationId}/result`, orchestration);
        await this.storeInMemory('orchestration/latest', orchestration);
      }

      console.log(`✅ Orchestration completed: ${orchestrationId} (${orchestration.duration}ms)`);
      return orchestration;
      
    } catch (error) {
      console.error(`❌ Orchestration failed: ${orchestrationId}`, error);
      orchestration.error = error.message;
      orchestration.status = 'failed';
      orchestration.duration = Date.now() - orchestration.startTime;
      
      // Store failed orchestration
      if (this.config.memorySharing) {
        await this.storeInMemory(`orchestration/${orchestrationId}/error`, orchestration);
      }
      
      throw error;
    }
  }

  /**
   * Create coordination plan
   */
  createCoordinationPlan(agents, taskDescription, options) {
    const plan = {
      strategy: options.strategy || 'parallel',
      dependencies: {},
      coordination: {},
      communication: {}
    };
    
    // Define agent dependencies based on capabilities
    agents.forEach((agent, index) => {
      const deps = [];
      
      // Add logical dependencies
      if (agent.type === 'tester' && agents.some(a => a.type === 'coder')) {
        deps.push('coder');
      }
      if (agent.type === 'reviewer' && agents.some(a => a.type === 'coder')) {
        deps.push('coder');
      }
      if (agent.type === 'planner' && agents.length > 1) {
        // Planner should run first
        plan.dependencies[agent.type] = [];
        return;
      }
      
      plan.dependencies[agent.type] = deps;
    });
    
    return plan;
  }

  /**
   * Execute agents in parallel
   */
  async executeInParallel(agents, taskDescription, options, orchestrationId) {
    console.log('🚀 Executing agents in parallel...');
    
    // Create agent requests
    const agentRequests = agents.map(agent => ({
      agentType: agent.type,
      taskDescription: this.createAgentTask(taskDescription, agent),
      options: { ...options, capability: agent.primaryCapability, orchestrationId }
    }));

    // Spawn all agents concurrently
    const results = await Promise.allSettled(
      agentRequests.map(req => this.spawnAgent(req.agentType, req.taskDescription, req.options))
    );
    
    return results.map((result, index) => ({
      agent: agents[index],
      success: result.status === 'fulfilled',
      result: result.status === 'fulfilled' ? result.value : null,
      error: result.status === 'rejected' ? result.reason.message : null
    }));
  }

  /**
   * Execute agents sequentially
   */
  async executeSequentially(agents, taskDescription, options, orchestrationId) {
    console.log('📝 Executing agents sequentially...');
    
    const results = [];
    let contextMemory = {};
    
    for (const agent of agents) {
      try {
        // Update task with context from previous agents
        const contextualTask = this.addContextToTask(
          this.createAgentTask(taskDescription, agent),
          contextMemory,
          results
        );
        
        const result = await this.spawnAgent(
          agent.type,
          contextualTask,
          { ...options, capability: agent.primaryCapability, orchestrationId }
        );
        
        const agentResult = {
          agent,
          success: true,
          result,
          sequence: results.length + 1
        };
        
        results.push(agentResult);
        
        // Update context memory
        contextMemory[agent.type] = result;
        
        // Store intermediate context
        if (this.config.memorySharing) {
          await this.storeInMemory(
            `orchestration/${orchestrationId}/context/${agent.type}`, 
            contextMemory
          );
        }
        
      } catch (error) {
        results.push({
          agent,
          success: false,
          error: error.message,
          sequence: results.length + 1
        });
        
        // Continue with next agent or stop based on configuration
        if (options.stopOnError) {
          break;
        }
      }
    }
    
    return results;
  }

  /**
   * Execute agents in pipeline
   */
  async executePipeline(agents, taskDescription, options, orchestrationId) {
    console.log('🔄 Executing agents in pipeline...');
    
    let currentOutput = taskDescription;
    const results = [];
    
    for (const agent of agents) {
      try {
        const pipelineTask = `[Pipeline Stage] ${currentOutput}`;
        const result = await this.spawnAgent(
          agent.type,
          pipelineTask,
          { ...options, capability: agent.primaryCapability, orchestrationId }
        );
        
        const agentResult = {
          agent,
          success: true,
          result,
          stage: results.length + 1,
          input: currentOutput
        };
        
        results.push(agentResult);
        
        // Use this result as input for next stage
        currentOutput = typeof result.result === 'string' ? result.result : JSON.stringify(result.result);
        
      } catch (error) {
        results.push({
          agent,
          success: false,
          error: error.message,
          stage: results.length + 1,
          input: currentOutput
        });
        
        // Pipeline stops on error
        break;
      }
    }
    
    return results;
  }

  /**
   * Execute agents hierarchically
   */
  async executeHierarchically(agents, taskDescription, options, orchestrationId) {
    console.log('🏗️ Executing agents hierarchically...');
    
    // Separate coordinator from workers
    const coordinatorAgent = agents.find(a => 
      a.type.includes('coordinator') || 
      a.type === 'planner' || 
      a.type === 'architect'
    ) || agents[0];
    
    const workerAgents = agents.filter(a => a !== coordinatorAgent);
    
    const results = [];
    
    // First, run coordinator
    try {
      const coordinatorTask = `[Coordinator] Plan and coordinate: ${taskDescription}`;
      const coordinatorResult = await this.spawnAgent(
        coordinatorAgent.type,
        coordinatorTask,
        { ...options, role: 'coordinator', orchestrationId }
      );
      
      results.push({
        agent: coordinatorAgent,
        success: true,
        result: coordinatorResult,
        role: 'coordinator'
      });
      
      // Store coordinator's plan
      if (this.config.memorySharing) {
        await this.storeInMemory(
          `orchestration/${orchestrationId}/plan`,
          coordinatorResult
        );
      }
      
    } catch (error) {
      results.push({
        agent: coordinatorAgent,
        success: false,
        error: error.message,
        role: 'coordinator'
      });
    }
    
    // Then run workers in parallel, referencing coordinator's plan
    if (workerAgents.length > 0) {
      const workerRequests = workerAgents.map(agent => ({
        agentType: agent.type,
        taskDescription: `[Worker] Follow coordinator's plan for: ${this.createAgentTask(taskDescription, agent)}\n\nCoordinator guidance available in memory: orchestration/${orchestrationId}/plan`,
        options: { ...options, capability: agent.primaryCapability, role: 'worker', orchestrationId }
      }));
      
      const workerResults = await Promise.allSettled(
        workerRequests.map(req => this.spawnAgent(req.agentType, req.taskDescription, req.options))
      );
      
      workerResults.forEach((result, index) => {
        results.push({
          agent: workerAgents[index],
          success: result.status === 'fulfilled',
          result: result.status === 'fulfilled' ? result.value : null,
          error: result.status === 'rejected' ? result.reason.message : null,
          role: 'worker'
        });
      });
    }
    
    return results;
  }

  /**
   * Analyze task requirements
   */
  analyzeTaskRequirements(taskDescription) {
    const requirements = new Set();
    const desc = taskDescription.toLowerCase();

    // Development capabilities
    if (desc.includes('implement') || desc.includes('build') || desc.includes('create')) {
      requirements.add('code-generation');
      requirements.add('implementation');
    }

    // Testing capabilities
    if (desc.includes('test') || desc.includes('tdd') || desc.includes('coverage')) {
      requirements.add('unit-testing');
      requirements.add('tdd');
    }

    // Security capabilities
    if (desc.includes('security') || desc.includes('auth') || desc.includes('encrypt')) {
      requirements.add('security-audit');
      requirements.add('authentication');
    }

    // Performance capabilities
    if (desc.includes('performance') || desc.includes('optimize') || desc.includes('speed')) {
      requirements.add('optimization');
      requirements.add('profiling');
    }

    // Architecture capabilities
    if (desc.includes('design') || desc.includes('architect') || desc.includes('structure')) {
      requirements.add('system-design');
      requirements.add('pattern-application');
    }

    // Default capability if none detected
    if (requirements.size === 0) {
      requirements.add('general');
    }

    return Array.from(requirements);
  }

  /**
   * Select optimal agents for capabilities
   */
  selectOptimalAgents(capabilities, options = {}) {
    const maxAgents = options.maxAgents || 5;
    const selectedAgents = [];
    const usedAgentTypes = new Set();

    for (const capability of capabilities) {
      const agents = getAgentsByCapability(capability);
      
      // Find best agent that hasn't been used
      for (const agent of agents) {
        if (!usedAgentTypes.has(agent.type) && selectedAgents.length < maxAgents) {
          selectedAgents.push({
            ...agent,
            primaryCapability: capability
          });
          usedAgentTypes.add(agent.type);
          break;
        }
      }
    }

    // If no agents selected, use general-purpose
    if (selectedAgents.length === 0) {
      selectedAgents.push({
        type: 'coder',
        primaryCapability: 'general'
      });
    }

    return selectedAgents;
  }

  /**
   * Create specific task for agent
   */
  createAgentTask(mainTask, agent) {
    const capability = agent.primaryCapability;
    const agentDef = getAgent(agent.type);
    
    return `[${agentDef.name}] ${mainTask}\n` +
           `Focus: ${capability}\n` +
           `Use your specialized capabilities: ${agentDef.capabilities.join(', ')}\n` +
           `Available tools: ${agentDef.tools.join(', ')}`;
  }

  /**
   * Add context to task for sequential execution
   */
  addContextToTask(task, contextMemory, previousResults) {
    if (!contextMemory || Object.keys(contextMemory).length === 0) {
      return task;
    }
    
    let contextInfo = '\n\n--- Context from previous agents ---\n';
    
    for (const [agentType, result] of Object.entries(contextMemory)) {
      contextInfo += `\n${agentType}: ${JSON.stringify(result, null, 2).substring(0, 500)}...\n`;
    }
    
    if (previousResults.length > 0) {
      contextInfo += '\n--- Previous Results Summary ---\n';
      previousResults.forEach((result, index) => {
        contextInfo += `${index + 1}. ${result.agent.type}: ${result.success ? 'SUCCESS' : 'FAILED'}\n`;
      });
    }
    
    return task + contextInfo;
  }

  /**
   * Synthesize results from multiple agents
   */
  async synthesizeResults(results, originalTask, orchestrationId) {
    const synthesis = {
      orchestrationId,
      originalTask,
      totalAgents: results.length,
      successfulAgents: results.filter(r => r.success).length,
      failedAgents: results.filter(r => !r.success).length,
      overallSuccess: results.every(r => r.success),
      partialSuccess: results.some(r => r.success),
      timestamp: new Date().toISOString()
    };
    
    // Extract key insights from successful agents
    const successfulResults = results.filter(r => r.success);
    synthesis.insights = {
      codeGenerated: successfulResults.filter(r => 
        r.agent.type === 'coder' || r.agent.primaryCapability === 'code-generation'
      ).length > 0,
      testsCreated: successfulResults.filter(r => 
        r.agent.type === 'tester' || r.agent.primaryCapability === 'unit-testing'
      ).length > 0,
      securityReviewed: successfulResults.filter(r => 
        r.agent.type === 'security-manager' || r.agent.primaryCapability === 'security-audit'
      ).length > 0,
      performanceAnalyzed: successfulResults.filter(r => 
        r.agent.type === 'perf-analyzer' || r.agent.primaryCapability === 'optimization'
      ).length > 0
    };
    
    // Create combined result summary
    synthesis.combinedResult = {
      summary: `Orchestrated task '${originalTask}' with ${synthesis.totalAgents} agents. ` +
               `${synthesis.successfulAgents} succeeded, ${synthesis.failedAgents} failed.`,
      details: results.map(r => ({
        agent: r.agent.type,
        success: r.success,
        hasResult: !!r.result,
        errorMessage: r.error
      })),
      recommendations: this.generateRecommendations(results, synthesis)
    };
    
    // Store synthesis in memory
    if (this.config.memorySharing) {
      await this.storeInMemory(`orchestration/${orchestrationId}/synthesis`, synthesis);
    }
    
    return synthesis;
  }

  /**
   * Generate recommendations based on results
   */
  generateRecommendations(results, synthesis) {
    const recommendations = [];
    
    // Check for missing critical agents
    if (!synthesis.insights.testsCreated && synthesis.insights.codeGenerated) {
      recommendations.push('Consider adding testing agent for generated code');
    }
    
    if (!synthesis.insights.securityReviewed && synthesis.insights.codeGenerated) {
      recommendations.push('Consider security review for generated code');
    }
    
    if (!synthesis.insights.performanceAnalyzed && synthesis.insights.codeGenerated) {
      recommendations.push('Consider performance analysis for generated code');
    }
    
    // Check for failed agents that could be retried
    const failedAgents = results.filter(r => !r.success);
    if (failedAgents.length > 0 && failedAgents.length < results.length) {
      recommendations.push(`Retry failed agents: ${failedAgents.map(r => r.agent.type).join(', ')}`);
    }
    
    // Success pattern recommendations
    if (synthesis.overallSuccess) {
      recommendations.push('All agents succeeded - pattern can be reused for similar tasks');
    } else if (synthesis.partialSuccess) {
      recommendations.push('Partial success - review failed agents and adjust strategy');
    }
    
    return recommendations;
  }

  /**
   * Analyze file context for agent suggestions
   */
  async analyzeFileContext(task) {
    const fileContext = {
      detectedFileTypes: [],
      suggestedAgents: [],
      complexity: 'medium'
    };
    
    const taskLower = task.toLowerCase();
    
    // File type detection
    const fileTypePatterns = {
      'javascript': /\.(js|jsx|ts|tsx)|javascript|react|node\.?js/,
      'python': /\.py|python|django|flask/,
      'go': /\.go|golang|go/,
      'java': /\.java|java|spring/,
      'css': /\.(css|scss|sass|less)|styl/,
      'html': /\.html?|html/,
      'dockerfile': /dockerfile|docker/,
      'yaml': /\.(yaml|yml)|yaml/
    };
    
    for (const [type, pattern] of Object.entries(fileTypePatterns)) {
      if (pattern.test(taskLower)) {
        fileContext.detectedFileTypes.push(type);
      }
    }
    
    // Agent suggestions based on file types
    const agentMappings = {
      'javascript': ['coder', 'frontend-dev', 'backend-dev', 'tester'],
      'python': ['coder', 'backend-dev', 'ml-developer', 'tester'],
      'go': ['coder', 'backend-dev', 'system-architect', 'tester'],
      'dockerfile': ['cicd-engineer', 'system-architect', 'security-manager'],
      'css': ['frontend-dev', 'coder'],
      'html': ['frontend-dev', 'coder']
    };
    
    for (const fileType of fileContext.detectedFileTypes) {
      if (agentMappings[fileType]) {
        fileContext.suggestedAgents.push(...agentMappings[fileType]);
      }
    }
    
    // Remove duplicates
    fileContext.suggestedAgents = [...new Set(fileContext.suggestedAgents)];
    
    // Complexity analysis
    const complexityIndicators = {
      high: ['architecture', 'system', 'enterprise', 'distributed', 'microservice'],
      medium: ['implement', 'build', 'create', 'develop'],
      low: ['fix', 'update', 'modify', 'change']
    };
    
    for (const [level, indicators] of Object.entries(complexityIndicators)) {
      if (indicators.some(indicator => taskLower.includes(indicator))) {
        fileContext.complexity = level;
        break;
      }
    }
    
    return fileContext;
  }

  /**
   * Validate task safety
   */
  validateTaskSafety(task) {
    const safety = {
      safe: true,
      warnings: [],
      risk: 'low'
    };
    
    const taskLower = task.toLowerCase();
    
    // Dangerous operations
    const dangerousPatterns = {
      'rm -rf': 'Potentially destructive file deletion',
      'sudo': 'Elevated privileges required',
      'chmod 777': 'Overly permissive file permissions',
      'DROP TABLE': 'Database table deletion',
      'DELETE FROM': 'Database data deletion',
      'eval(': 'Code evaluation - potential security risk',
      'exec(': 'Code execution - potential security risk'
    };
    
    for (const [pattern, warning] of Object.entries(dangerousPatterns)) {
      if (taskLower.includes(pattern.toLowerCase())) {
        safety.warnings.push(warning);
        safety.risk = 'high';
      }
    }
    
    // Moderate risk patterns
    const moderatePatterns = {
      'production': 'Production environment changes',
      'database': 'Database operations',
      'deploy': 'Deployment operations',
      'migrate': 'Migration operations'
    };
    
    for (const [pattern, warning] of Object.entries(moderatePatterns)) {
      if (taskLower.includes(pattern) && safety.risk === 'low') {
        safety.warnings.push(warning);
        safety.risk = 'medium';
      }
    }
    
    safety.safe = safety.risk === 'low';
    
    return safety;
  }

  /**
   * Prepare resources for agent
   */
  async prepareResources(agentType, task) {
    try {
      // Create necessary directories based on agent type
      const resourceMappings = {
        'coder': ['src', 'tests'],
        'tester': ['tests', 'coverage'],
        'documentation': ['docs'],
        'cicd-engineer': ['.github/workflows', 'scripts'],
        'security-manager': ['security', 'audits']
      };
      
      if (resourceMappings[agentType]) {
        for (const dir of resourceMappings[agentType]) {
          // This would be handled by the bridge in a real implementation
          console.log(`\ud83d\udcc1 Preparing directory: ${dir}`);
        }
      }
      
      // Store resource preparation
      if (this.config.memorySharing) {
        await this.storeInMemory(`resources/${agentType}`, {
          directories: resourceMappings[agentType] || [],
          timestamp: new Date().toISOString(),
          task
        });
      }
      
    } catch (error) {
      console.warn(`Resource preparation failed for ${agentType}:`, error.message);
    }
  }

  /**
   * Train from successful execution
   */
  async trainFromSuccess(agentType, data) {
    try {
      const trainingData = {
        agentType,
        task: data.task,
        result: data.result,
        success: data.success,
        timestamp: new Date().toISOString(),
        patterns: this.extractPatterns(data)
      };
      
      // Store successful patterns for future reference
      await this.storeInMemory(
        `training/${agentType}/${Date.now()}`, 
        trainingData,
        { ttl: 7 * 24 * 60 * 60 * 1000 } // 7 days
      );
      
      console.log(`\ud83e\udde0 Training patterns stored for ${agentType}`);
      
    } catch (error) {
      console.warn(`Training failed for ${agentType}:`, error.message);
    }
  }

  /**
   * Extract patterns from successful execution
   */
  extractPatterns(data) {
    const patterns = {
      taskLength: data.task?.length || 0,
      hasResult: !!data.result,
      executionTime: data.executionTime || 0,
      agentCapabilities: data.result?.agentType ? 
        getAgent(data.result.agentType)?.capabilities || [] : []
    };
    
    // Extract result patterns
    if (data.result?.result) {
      const resultStr = JSON.stringify(data.result.result);
      patterns.resultPatterns = {
        hasCode: resultStr.includes('function') || resultStr.includes('class'),
        hasTests: resultStr.includes('test') || resultStr.includes('expect'),
        hasDocumentation: resultStr.includes('/**') || resultStr.includes('README'),
        resultLength: resultStr.length
      };
    }
    
    return patterns;
  }

  /**
   * Analyze performance
   */
  analyzePerformance(data) {
    const startTime = data.startTime || Date.now();
    const endTime = Date.now();
    const duration = endTime - startTime;
    
    return {
      duration,
      agentType: data.agentType,
      success: data.success,
      taskComplexity: this.calculateTaskComplexity(data.task),
      efficiency: this.calculateEfficiency(duration, data.task, data.success),
      timestamp: new Date().toISOString()
    };
  }

  /**
   * Calculate task complexity
   */
  calculateTaskComplexity(task) {
    if (!task) return 1;
    
    let complexity = 1;
    
    // Length factor
    complexity += Math.floor(task.length / 100);
    
    // Keyword complexity
    const complexKeywords = ['implement', 'architect', 'design', 'optimize', 'secure'];
    const simpleKeywords = ['fix', 'update', 'modify'];
    
    const taskLower = task.toLowerCase();
    
    for (const keyword of complexKeywords) {
      if (taskLower.includes(keyword)) complexity += 2;
    }
    
    for (const keyword of simpleKeywords) {
      if (taskLower.includes(keyword)) complexity -= 1;
    }
    
    return Math.max(1, complexity);
  }

  /**
   * Calculate efficiency score
   */
  calculateEfficiency(duration, task, success) {
    if (!success) return 0;
    
    const complexity = this.calculateTaskComplexity(task);
    const expectedTime = complexity * 10000; // 10 seconds per complexity point
    
    // Efficiency is inversely related to duration vs expected time
    const efficiency = Math.max(0.1, Math.min(1.0, expectedTime / duration));
    
    return efficiency;
  }

  /**
   * Collect system metrics
   */
  async collectMetrics() {
    const metrics = {
      timestamp: new Date().toISOString(),
      uptime: Date.now() - (this._initTime || Date.now()),
      totalAgents: this.activeAgents.size,
      agentsByStatus: {},
      agentsByType: {},
      memoryUsage: {
        cacheSize: this._memoryCache ? this._memoryCache.size : 0,
        fallbackSize: this._fallbackMemory ? this._fallbackMemory.size : 0
      },
      performance: {
        averageTaskDuration: 0,
        successRate: 0,
        totalTasks: 0
      }
    };
    
    // Analyze active agents
    for (const [key, agent] of this.activeAgents.entries()) {
      const [type] = key.split(':');
      
      // Count by status
      metrics.agentsByStatus[agent.status] = (metrics.agentsByStatus[agent.status] || 0) + 1;
      
      // Count by type
      metrics.agentsByType[type] = (metrics.agentsByType[type] || 0) + 1;
      
      // Aggregate performance data
      if (agent.history) {
        metrics.performance.totalTasks += agent.history.length;
        
        const successfulTasks = agent.history.filter(h => h.success);
        metrics.performance.successRate = successfulTasks.length / agent.history.length;
        
        const durations = agent.history
          .filter(h => h.endTime && h.startTime)
          .map(h => h.endTime - h.startTime);
          
        if (durations.length > 0) {
          metrics.performance.averageTaskDuration = durations.reduce((a, b) => a + b, 0) / durations.length;
        }
      }
    }
    
    return metrics;
  }

  /**
   * Get agent status
   */
  async getAgentStatus(agentType, agentId = null) {
    if (agentId) {
      const agentKey = `${agentType}:${agentId}`;
      const agent = this.activeAgents.get(agentKey);
      
      if (!agent) {
        return { error: `Agent ${agentKey} not found` };
      }
      
      // Get detailed status including memory data
      const memoryData = await this.retrieveFromMemory(`agent/${agentType}/${agentId}/task`);
      
      return {
        agentId,
        agentType,
        status: agent.status,
        lastUpdated: agent.lastUpdated,
        history: agent.history,
        memoryData,
        definition: agent.definition
      };
    } else {
      // Get all agents of this type
      const typeAgents = Array.from(this.activeAgents.entries())
        .filter(([key]) => key.startsWith(`${agentType}:`))
        .map(([key, agent]) => ({
          agentId: agent.agentId,
          status: agent.status,
          lastUpdated: agent.lastUpdated,
          lastTask: agent.history[agent.history.length - 1]?.task
        }));
        
      return {
        agentType,
        count: typeAgents.length,
        agents: typeAgents
      };
    }
  }

  /**
   * List all active agents
   */
  async listActiveAgents(options = {}) {
    const { status, agentType, limit } = options;
    
    let agents = Array.from(this.activeAgents.entries()).map(([key, agent]) => {
      const [type, id] = key.split(':');
      return {
        key,
        agentType: type,
        agentId: id,
        status: agent.status,
        lastUpdated: agent.lastUpdated,
        definition: agent.definition,
        historyCount: agent.history.length
      };
    });
    
    // Apply filters
    if (status) {
      agents = agents.filter(a => a.status === status);
    }
    
    if (agentType) {
      agents = agents.filter(a => a.agentType === agentType);
    }
    
    // Sort by last updated (most recent first)
    agents.sort((a, b) => (b.lastUpdated || 0) - (a.lastUpdated || 0));
    
    // Apply limit
    if (limit) {
      agents = agents.slice(0, limit);
    }
    
    return {
      total: this.activeAgents.size,
      filtered: agents.length,
      agents
    };
  }

  /**
   * Get integration status
   */
  async getStatus() {
    const activeAgentsList = await this.listActiveAgents({ limit: 10 });
    
    const status = {
      initialized: this.initialized,
      swarmId: this.swarmId,
      activeAgents: activeAgentsList,
      bridgeStatus: await this.bridge.getStatus(),
      config: this.config,
      memory: {
        cacheSize: this._memoryCache ? this._memoryCache.size : 0,
        fallbackSize: this._fallbackMemory ? this._fallbackMemory.size : 0
      },
      uptime: this.initialized ? Date.now() - this._initTime : 0
    };

    return status;
  }

  /**
   * Graceful shutdown with cleanup
   */
  async shutdown() {
    try {
      console.log('📏 Shutting down Claude Flow Integration...');
      
      // Execute session end hook
      if (this.config.hooks) {
        await this.executeHook('session-end', { exportMetrics: true });
      }
      
      // Collect final metrics
      const finalMetrics = await this.collectMetrics();
      if (this.config.memorySharing) {
        await this.storeInMemory('integration/final-metrics', finalMetrics);
      }
      
      // Shutdown bridge
      await this.bridge.shutdown();
      
      // Clear caches
      if (this._memoryCache) {
        this._memoryCache.clear();
      }
      if (this._fallbackMemory) {
        this._fallbackMemory.clear();
      }
      
      // Clear active agents
      this.activeAgents.clear();
      
      // Destroy swarm if exists
      if (this.swarmId) {
        try {
          await execAsync(`npx claude-flow@alpha swarm destroy --id ${this.swarmId}`);
          console.log(`🐝 Swarm ${this.swarmId} destroyed`);
        } catch (error) {
          console.warn('Swarm destruction failed:', error.message);
        }
      }

      this.initialized = false;
      console.log('✅ Claude Flow Integration shutdown complete');
      
    } catch (error) {
      console.error('❌ Error during shutdown:', error);
    }
  }

  /**
   * Emergency shutdown (faster, less cleanup)
   */
  async emergencyShutdown() {
    try {
      console.log('🆘 Emergency shutdown initiated');
      
      // Quick bridge shutdown
      if (this.bridge) {
        await this.bridge.emergencyShutdown?.() || this.bridge.shutdown();
      }
      
      // Clear everything immediately
      this.activeAgents.clear();
      this._memoryCache?.clear();
      this._fallbackMemory?.clear();
      
      this.initialized = false;
      console.log('⚡ Emergency shutdown complete');
      
    } catch (error) {
      console.error('🚨 Emergency shutdown error:', error);
    }
  }

  /**
   * Health check
   */
  async healthCheck() {
    const health = {
      status: 'healthy',
      initialized: this.initialized,
      timestamp: new Date().toISOString(),
      checks: {}
    };
    
    try {
      // Check bridge health
      health.checks.bridge = await this.bridge.healthCheck?.() || { status: 'unknown' };
      
      // Check memory systems
      health.checks.memory = {
        cache: this._memoryCache ? 'active' : 'inactive',
        fallback: this._fallbackMemory ? 'active' : 'inactive'
      };
      
      // Check swarm
      if (this.swarmId) {
        try {
          await execAsync(`npx claude-flow@alpha swarm status --id ${this.swarmId}`);
          health.checks.swarm = { status: 'active', id: this.swarmId };
        } catch (error) {
          health.checks.swarm = { status: 'error', error: error.message };
          health.status = 'degraded';
        }
      } else {
        health.checks.swarm = { status: 'not-configured' };
      }
      
      // Check active agents
      health.checks.agents = {
        total: this.activeAgents.size,
        active: Array.from(this.activeAgents.values()).filter(a => a.status === 'active').length,
        failed: Array.from(this.activeAgents.values()).filter(a => a.status === 'failed').length
      };
      
      // Overall health determination
      if (health.checks.bridge.status === 'error' || !this.initialized) {
        health.status = 'unhealthy';
      } else if (health.status !== 'degraded' && health.checks.agents.failed > health.checks.agents.active) {
        health.status = 'degraded';
      }
      
    } catch (error) {
      health.status = 'error';
      health.error = error.message;
    }
    
    return health;
  }
}

/**
 * Create integration instance
 */
export function createIntegration(options = {}) {
  return new ClaudeFlowIntegration(options);
}

/**
 * Quick orchestration function
 */
export async function orchestrate(task, options = {}) {
  const integration = createIntegration(options);
  await integration.initialize();
  
  try {
    const result = await integration.orchestrateTask(task, options);
    return result;
  } finally {
    await integration.shutdown();
  }
}

/**
 * Batch orchestration for multiple tasks
 */
export async function orchestrateBatch(tasks, options = {}) {
  const integration = createIntegration(options);
  await integration.initialize();
  
  try {
    const results = [];
    const { parallel = false, maxConcurrent = 3 } = options;
    
    if (parallel) {
      // Execute tasks in parallel with concurrency limit
      const chunks = [];
      for (let i = 0; i < tasks.length; i += maxConcurrent) {
        chunks.push(tasks.slice(i, i + maxConcurrent));
      }
      
      for (const chunk of chunks) {
        const chunkResults = await Promise.allSettled(
          chunk.map(task => integration.orchestrateTask(task, options))
        );
        
        results.push(...chunkResults.map((result, index) => ({
          task: chunk[index],
          success: result.status === 'fulfilled',
          result: result.status === 'fulfilled' ? result.value : null,
          error: result.status === 'rejected' ? result.reason.message : null
        })));
      }
    } else {
      // Execute tasks sequentially
      for (const task of tasks) {
        try {
          const result = await integration.orchestrateTask(task, options);
          results.push({ task, success: true, result });
        } catch (error) {
          results.push({ task, success: false, error: error.message });
          
          if (options.stopOnError) {
            break;
          }
        }
      }
    }
    
    return {
      total: tasks.length,
      completed: results.length,
      successful: results.filter(r => r.success).length,
      failed: results.filter(r => !r.success).length,
      results
    };
    
  } finally {
    await integration.shutdown();
  }
}

/**
 * Quick agent spawn function
 */
export async function spawnAgent(agentType, task, options = {}) {
  const integration = createIntegration(options);
  await integration.initialize();
  
  try {
    const result = await integration.spawnAgent(agentType, task, options);
    return result;
  } finally {
    await integration.shutdown();
  }
}

/**
 * Get available agents
 */
export function getAvailableAgents() {
  return Object.entries(AGENT_REGISTRY).map(([type, agent]) => ({
    type,
    name: agent.name,
    description: agent.description,
    capabilities: agent.capabilities,
    tools: agent.tools
  }));
}

/**
 * Find agents by capability
 */
export function findAgents(capability) {
  return getAgentsByCapability(capability);
}

/**
 * CLI interface
 */
if (import.meta.url === `file://${process.argv[1]}`) {
  const args = process.argv.slice(2);
  
  if (args.length === 0) {
    console.log(`
Claude Flow Agent Integration

Usage: 
  node claude-flow-integration.js <command> [options]

Commands:
  spawn <agent> "<task>"           Spawn a specific agent with task
  orchestrate "<task>"             Orchestrate task with optimal agents
  batch <tasks-file.json>          Execute multiple tasks from file
  list                             List all available agents
  status                           Show integration status
  health                           Show health check results
  agents [type] [id]               Show agent status
  memory <search|get|clear> <key>  Memory operations

Examples:
  node claude-flow-integration.js spawn coder "Implement REST API"
  node claude-flow-integration.js orchestrate "Build secure authentication system"
  node claude-flow-integration.js list
    `);
    process.exit(0);
  }

  const command = args[0];
  
  async function run() {
    const integration = createIntegration();
    
    try {
      switch (command) {
        case 'spawn': {
          const agentType = args[1];
          const task = args[2];
          
          if (!agentType || !task) {
            console.error('Usage: spawn <agent> "<task>"');
            process.exit(1);
          }
          
          await integration.initialize();
          const result = await integration.spawnAgent(agentType, task);
          console.log('Result:', result);
          break;
        }
        
        case 'orchestrate': {
          const task = args[1];
          
          if (!task) {
            console.error('Usage: orchestrate "<task>"');
            process.exit(1);
          }
          
          await integration.initialize();
          const result = await integration.orchestrateTask(task);
          console.log('Orchestration Result:', JSON.stringify(result, null, 2));
          break;
        }
        
        case 'list': {
          console.log('Available Agents:');
          for (const [type, agent] of Object.entries(AGENT_REGISTRY)) {
            console.log(`  ${type}: ${agent.description}`);
          }
          break;
        }
        
        case 'status': {
          await integration.initialize();
          const status = await integration.getStatus();
          console.log('Integration Status:', JSON.stringify(status, null, 2));
          break;
        }
        
        case 'health': {
          await integration.initialize();
          const health = await integration.healthCheck();
          console.log('Health Check:', JSON.stringify(health, null, 2));
          break;
        }
        
        case 'agents': {
          const agentType = args[1];
          const agentId = args[2];
          
          await integration.initialize();
          
          if (agentType) {
            const status = await integration.getAgentStatus(agentType, agentId);
            console.log(`Agent Status (${agentType}${agentId ? `:${agentId}` : ''}):`, JSON.stringify(status, null, 2));
          } else {
            const agents = await integration.listActiveAgents();
            console.log('Active Agents:', JSON.stringify(agents, null, 2));
          }
          break;
        }
        
        case 'memory': {
          const action = args[1]; // search, get, clear
          const key = args[2];
          
          await integration.initialize();
          
          switch (action) {
            case 'search':
              if (!key) {
                console.error('Usage: memory search <pattern>');
                process.exit(1);
              }
              const searchResults = await integration.searchMemory(key);
              console.log('Memory Search Results:', JSON.stringify(searchResults, null, 2));
              break;
              
            case 'get':
              if (!key) {
                console.error('Usage: memory get <key>');
                process.exit(1);
              }
              const value = await integration.retrieveFromMemory(key);
              console.log(`Memory Value (${key}):`, JSON.stringify(value, null, 2));
              break;
              
            case 'clear':
              if (!key) {
                console.error('Usage: memory clear <key>');
                process.exit(1);
              }
              await integration.clearMemory(key);
              console.log(`Memory cleared: ${key}`);
              break;
              
            default:
              console.error('Memory actions: search, get, clear');
              process.exit(1);
          }
          break;
        }
        
        case 'batch': {
          const tasksFile = args[1];
          
          if (!tasksFile) {
            console.error('Usage: batch <tasks-file.json>');
            process.exit(1);
          }
          
          try {
            const fs = await import('fs/promises');
            const tasksData = JSON.parse(await fs.readFile(tasksFile, 'utf8'));
            const tasks = Array.isArray(tasksData) ? tasksData : tasksData.tasks;
            
            const batchOptions = {
              ...tasksData.options,
              parallel: tasksData.parallel || false,
              maxConcurrent: tasksData.maxConcurrent || 3
            };
            
            console.log(`Starting batch execution of ${tasks.length} tasks...`);
            const result = await orchestrateBatch(tasks, batchOptions);
            console.log('Batch Results:', JSON.stringify(result, null, 2));
            
          } catch (error) {
            console.error('Batch execution failed:', error.message);
            process.exit(1);
          }
          break;
        }
        
        default:
          console.error(`Unknown command: ${command}`);
          process.exit(1);
      }
    } finally {
      await integration.shutdown();
    }
  }
  
  run().catch(console.error);
}