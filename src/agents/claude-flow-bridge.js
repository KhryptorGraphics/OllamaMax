#!/usr/bin/env node

/**
 * Claude Flow Bridge - Agent Activation Bridge
 * 
 * Comprehensive agent activation bridge that maps claude-flow agent requests 
 * to Claude Code Task tool with proper coordination and error handling.
 * 
 * Features:
 * - Intelligent agent spawning system
 * - Hook integration for pre/post task coordination
 * - Agent capability matching and tool selection
 * - Fallback mechanisms for unsupported agent types
 * - Swarm coordination initialization
 * - Agent lifecycle and cleanup management
 * - Production-ready error handling and logging
 */

import { AGENT_REGISTRY, getAgent, getAgentsByCapability, mapToClaudeCodeType } from './agent-registry.js';
import { execSync } from 'child_process';
import { writeFileSync, readFileSync, existsSync, mkdirSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/**
 * Bridge Configuration
 */
const BRIDGE_CONFIG = {
  maxAgents: 10,
  timeoutMs: 300000, // 5 minutes
  retryAttempts: 3,
  retryDelayMs: 1000,
  logLevel: process.env.NODE_ENV === 'production' ? 'info' : 'debug',
  hookTimeout: 30000, // 30 seconds for hooks
  swarmInitTimeout: 60000, // 60 seconds for swarm init
  memoryPath: join(process.cwd(), 'memory', 'bridge-state.json'),
  logPath: join(process.cwd(), 'logs', 'claude-flow-bridge.log')
};

/**
 * Agent State Management
 */
class AgentStateManager {
  constructor() {
    this.activeAgents = new Map();
    this.swarmState = null;
    this.executionHistory = [];
    this.ensureDirectories();
    this.loadState();
  }

  ensureDirectories() {
    const dirs = [
      dirname(BRIDGE_CONFIG.memoryPath),
      dirname(BRIDGE_CONFIG.logPath)
    ];
    
    dirs.forEach(dir => {
      if (!existsSync(dir)) {
        mkdirSync(dir, { recursive: true });
      }
    });
  }

  loadState() {
    try {
      if (existsSync(BRIDGE_CONFIG.memoryPath)) {
        const state = JSON.parse(readFileSync(BRIDGE_CONFIG.memoryPath, 'utf8'));
        this.activeAgents = new Map(state.activeAgents || []);
        this.swarmState = state.swarmState || null;
        this.executionHistory = state.executionHistory || [];
      }
    } catch (error) {
      Logger.warn('Failed to load bridge state:', error.message);
    }
  }

  saveState() {
    try {
      const state = {
        activeAgents: Array.from(this.activeAgents.entries()),
        swarmState: this.swarmState,
        executionHistory: this.executionHistory.slice(-100), // Keep last 100 entries
        timestamp: new Date().toISOString()
      };
      writeFileSync(BRIDGE_CONFIG.memoryPath, JSON.stringify(state, null, 2));
    } catch (error) {
      Logger.error('Failed to save bridge state:', error.message);
    }
  }

  addAgent(agentId, config) {
    this.activeAgents.set(agentId, {
      ...config,
      createdAt: new Date().toISOString(),
      status: 'spawning'
    });
    this.saveState();
  }

  updateAgentStatus(agentId, status, metadata = {}) {
    const agent = this.activeAgents.get(agentId);
    if (agent) {
      agent.status = status;
      agent.lastUpdate = new Date().toISOString();
      agent.metadata = { ...agent.metadata, ...metadata };
      this.saveState();
    }
  }

  removeAgent(agentId) {
    const agent = this.activeAgents.get(agentId);
    if (agent) {
      this.executionHistory.push({
        agentId,
        config: agent,
        completedAt: new Date().toISOString()
      });
      this.activeAgents.delete(agentId);
      this.saveState();
    }
  }

  getActiveAgents() {
    return Array.from(this.activeAgents.values());
  }

  setSwarmState(state) {
    this.swarmState = state;
    this.saveState();
  }
}

/**
 * Enhanced Logger with multiple levels and file output
 */
class Logger {
  static levels = { debug: 0, info: 1, warn: 2, error: 3 };
  static currentLevel = Logger.levels[BRIDGE_CONFIG.logLevel] || 1;

  static log(level, ...args) {
    if (Logger.levels[level] >= Logger.currentLevel) {
      const timestamp = new Date().toISOString();
      const message = `[${timestamp}] [${level.toUpperCase()}] ${args.join(' ')}`;
      console.log(message);
      
      try {
        const logEntry = message + '\n';
        require('fs').appendFileSync(BRIDGE_CONFIG.logPath, logEntry);
      } catch (error) {
        // Fail silently for file logging
      }
    }
  }

  static debug(...args) { Logger.log('debug', ...args); }
  static info(...args) { Logger.log('info', ...args); }
  static warn(...args) { Logger.log('warn', ...args); }
  static error(...args) { Logger.log('error', ...args); }
}

/**
 * Hook Integration Manager
 */
class HookManager {
  static async executeHook(hookType, agentType, taskDescription, metadata = {}) {
    const startTime = Date.now();
    Logger.debug(`Executing ${hookType} hook for agent: ${agentType}`);

    try {
      const agent = getAgent(agentType);
      const hookCommand = agent.hooks?.[hookType];
      
      if (!hookCommand) {
        Logger.debug(`No ${hookType} hook defined for agent: ${agentType}`);
        return { success: true, skipped: true };
      }

      const sessionId = metadata.sessionId || `swarm-${Date.now()}`;
      const enhancedCommand = hookCommand
        .replace('--description "[task]"', `--description "${taskDescription}"`)
        .replace('--session-id "swarm-[id]"', `--session-id "${sessionId}"`)
        .replace('--agent coder', `--agent ${agentType}`);

      const result = execSync(enhancedCommand, {
        timeout: BRIDGE_CONFIG.hookTimeout,
        encoding: 'utf8',
        stdio: ['pipe', 'pipe', 'pipe']
      });

      const duration = Date.now() - startTime;
      Logger.info(`${hookType} hook completed for ${agentType} in ${duration}ms`);
      
      return {
        success: true,
        output: result.trim(),
        duration,
        command: enhancedCommand
      };

    } catch (error) {
      const duration = Date.now() - startTime;
      Logger.error(`${hookType} hook failed for ${agentType} after ${duration}ms:`, error.message);
      
      return {
        success: false,
        error: error.message,
        duration,
        recoverable: error.code !== 'ETIMEDOUT'
      };
    }
  }

  static async executePreTask(agentType, taskDescription, metadata = {}) {
    const results = await Promise.all([
      HookManager.executeHook('pre', agentType, taskDescription, metadata),
      HookManager.executeSessionRestore(metadata.sessionId)
    ]);

    return {
      preTask: results[0],
      sessionRestore: results[1],
      success: results.every(r => r.success)
    };
  }

  static async executePostTask(agentType, taskDescription, metadata = {}) {
    const results = await Promise.all([
      HookManager.executeHook('post', agentType, taskDescription, metadata),
      HookManager.executeSessionEnd(metadata.sessionId)
    ]);

    return {
      postTask: results[0],
      sessionEnd: results[1],
      success: results.every(r => r.success)
    };
  }

  static async executeSessionRestore(sessionId) {
    if (!sessionId) return { success: true, skipped: true };

    try {
      const command = `npx claude-flow@alpha hooks session-restore --session-id "${sessionId}"`;
      const result = execSync(command, {
        timeout: BRIDGE_CONFIG.hookTimeout,
        encoding: 'utf8'
      });

      Logger.debug(`Session restore completed for: ${sessionId}`);
      return { success: true, output: result.trim() };
    } catch (error) {
      Logger.warn(`Session restore failed for ${sessionId}:`, error.message);
      return { success: false, error: error.message, recoverable: true };
    }
  }

  static async executeSessionEnd(sessionId) {
    if (!sessionId) return { success: true, skipped: true };

    try {
      const command = `npx claude-flow@alpha hooks session-end --export-metrics true --session-id "${sessionId}"`;
      const result = execSync(command, {
        timeout: BRIDGE_CONFIG.hookTimeout,
        encoding: 'utf8'
      });

      Logger.debug(`Session end completed for: ${sessionId}`);
      return { success: true, output: result.trim() };
    } catch (error) {
      Logger.warn(`Session end failed for ${sessionId}:`, error.message);
      return { success: false, error: error.message, recoverable: true };
    }
  }
}

/**
 * Swarm Coordination Manager
 */
class SwarmCoordinator {
  constructor(stateManager) {
    this.stateManager = stateManager;
  }

  async initializeSwarm(topology = 'mesh', maxAgents = BRIDGE_CONFIG.maxAgents) {
    Logger.info(`Initializing swarm with topology: ${topology}, maxAgents: ${maxAgents}`);

    try {
      const swarmConfig = {
        topology,
        maxAgents,
        strategy: 'adaptive',
        createdAt: new Date().toISOString()
      };

      // Store swarm state before MCP call
      this.stateManager.setSwarmState(swarmConfig);

      // Optional: Initialize MCP swarm if available
      try {
        const mcpCommand = `npx claude-flow@alpha mcp swarm_init --topology ${topology} --maxAgents ${maxAgents}`;
        const result = execSync(mcpCommand, {
          timeout: BRIDGE_CONFIG.swarmInitTimeout,
          encoding: 'utf8'
        });
        
        swarmConfig.mcpInitialized = true;
        swarmConfig.mcpOutput = result.trim();
        Logger.info('MCP swarm initialization successful');
      } catch (mcpError) {
        Logger.warn('MCP swarm initialization failed, proceeding without MCP:', mcpError.message);
        swarmConfig.mcpInitialized = false;
        swarmConfig.mcpError = mcpError.message;
      }

      this.stateManager.setSwarmState(swarmConfig);
      return swarmConfig;

    } catch (error) {
      Logger.error('Swarm initialization failed:', error.message);
      throw new SwarmInitializationError(`Failed to initialize swarm: ${error.message}`);
    }
  }

  async spawnMCPAgent(agentType) {
    try {
      const command = `npx claude-flow@alpha mcp agent_spawn --type ${agentType}`;
      const result = execSync(command, {
        timeout: BRIDGE_CONFIG.hookTimeout,
        encoding: 'utf8'
      });
      
      Logger.debug(`MCP agent spawned: ${agentType}`);
      return { success: true, output: result.trim() };
    } catch (error) {
      Logger.warn(`MCP agent spawn failed for ${agentType}:`, error.message);
      return { success: false, error: error.message };
    }
  }
}

/**
 * Agent Capability Matcher
 */
class CapabilityMatcher {
  static analyzeTask(taskDescription) {
    const analysis = {
      capabilities: [],
      tools: [],
      complexity: 'medium',
      domains: [],
      keywords: []
    };

    const keywords = taskDescription.toLowerCase();
    
    // Capability detection
    if (keywords.includes('test') || keywords.includes('testing')) {
      analysis.capabilities.push('unit-testing', 'integration-testing');
      analysis.tools.push('Bash', 'Write');
    }
    
    if (keywords.includes('api') || keywords.includes('backend')) {
      analysis.capabilities.push('api-design', 'database-design');
      analysis.domains.push('backend');
    }
    
    if (keywords.includes('ui') || keywords.includes('frontend') || keywords.includes('component')) {
      analysis.capabilities.push('ui-development', 'component-design');
      analysis.domains.push('frontend');
    }
    
    if (keywords.includes('security') || keywords.includes('auth')) {
      analysis.capabilities.push('security-audit', 'authentication');
      analysis.domains.push('security');
    }
    
    if (keywords.includes('performance') || keywords.includes('optimize')) {
      analysis.capabilities.push('optimization', 'profiling');
      analysis.domains.push('performance');
    }
    
    if (keywords.includes('architecture') || keywords.includes('design')) {
      analysis.capabilities.push('system-design', 'pattern-application');
      analysis.domains.push('architecture');
    }

    // Complexity analysis
    if (keywords.includes('comprehensive') || keywords.includes('complex') || keywords.includes('enterprise')) {
      analysis.complexity = 'high';
    } else if (keywords.includes('simple') || keywords.includes('basic') || keywords.includes('quick')) {
      analysis.complexity = 'low';
    }

    // Extract keywords
    const taskKeywords = keywords.split(/\s+/).filter(word => word.length > 3);
    analysis.keywords = [...new Set(taskKeywords)];

    return analysis;
  }

  static matchAgents(taskAnalysis, maxAgents = 3) {
    const scoredAgents = [];

    for (const [agentType, agent] of Object.entries(AGENT_REGISTRY)) {
      let score = 0;
      
      // Capability matching (40% weight)
      const capabilityMatch = taskAnalysis.capabilities.filter(cap => 
        agent.capabilities.some(agentCap => 
          agentCap.includes(cap) || cap.includes(agentCap)
        )
      ).length;
      score += (capabilityMatch / Math.max(taskAnalysis.capabilities.length, 1)) * 40;
      
      // Tool matching (30% weight)
      const toolMatch = taskAnalysis.tools.filter(tool => 
        agent.tools.includes(tool)
      ).length;
      score += (toolMatch / Math.max(taskAnalysis.tools.length, 1)) * 30;
      
      // Domain matching (20% weight)
      const domainMatch = taskAnalysis.domains.some(domain => 
        agentType.includes(domain) || 
        agent.description.toLowerCase().includes(domain) ||
        agent.specializations && Object.values(agent.specializations).flat().some(spec =>
          typeof spec === 'string' && spec.includes(domain)
        )
      );
      score += domainMatch ? 20 : 0;
      
      // Keyword matching (10% weight)
      const keywordMatch = taskAnalysis.keywords.filter(keyword =>
        agent.description.toLowerCase().includes(keyword) ||
        agentType.includes(keyword)
      ).length;
      score += (keywordMatch / Math.max(taskAnalysis.keywords.length, 1)) * 10;

      if (score > 0) {
        scoredAgents.push({ agentType, agent, score });
      }
    }

    // Sort by score and return top matches
    return scoredAgents
      .sort((a, b) => b.score - a.score)
      .slice(0, maxAgents);
  }

  static getFallbackAgent(taskDescription) {
    const analysis = CapabilityMatcher.analyzeTask(taskDescription);
    
    // Default fallbacks based on task type
    if (analysis.domains.includes('backend')) return 'backend-dev';
    if (analysis.domains.includes('frontend')) return 'coder';
    if (analysis.domains.includes('testing')) return 'tester';
    if (analysis.domains.includes('security')) return 'security-manager';
    if (analysis.domains.includes('architecture')) return 'system-architect';
    
    return 'coder'; // Ultimate fallback
  }
}

/**
 * Custom Error Classes
 */
class BridgeError extends Error {
  constructor(message, code = 'BRIDGE_ERROR') {
    super(message);
    this.name = 'BridgeError';
    this.code = code;
  }
}

class SwarmInitializationError extends BridgeError {
  constructor(message) {
    super(message, 'SWARM_INIT_ERROR');
    this.name = 'SwarmInitializationError';
  }
}

class AgentSpawnError extends BridgeError {
  constructor(message, agentType) {
    super(message, 'AGENT_SPAWN_ERROR');
    this.name = 'AgentSpawnError';
    this.agentType = agentType;
  }
}

/**
 * Main Bridge Class
 */
export class ClaudeFlowBridge {
  constructor(options = {}) {
    this.config = { ...BRIDGE_CONFIG, ...options };
    this.stateManager = new AgentStateManager();
    this.swarmCoordinator = new SwarmCoordinator(this.stateManager);
    this.initialized = false;
    
    Logger.info('Claude Flow Bridge initialized');
  }

  /**
   * Initialize the bridge and optionally set up swarm coordination
   */
  async initialize(swarmConfig = null) {
    try {
      Logger.info('Initializing Claude Flow Bridge...');
      
      if (swarmConfig) {
        await this.swarmCoordinator.initializeSwarm(
          swarmConfig.topology || 'mesh',
          swarmConfig.maxAgents || this.config.maxAgents
        );
      }
      
      this.initialized = true;
      Logger.info('Bridge initialization completed successfully');
      return true;
    } catch (error) {
      Logger.error('Bridge initialization failed:', error.message);
      throw error;
    }
  }

  /**
   * Spawn a single agent with full lifecycle management
   */
  async spawnAgent(agentType, taskDescription, options = {}) {
    const agentId = `agent-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    try {
      Logger.info(`Spawning agent: ${agentType} (${agentId})`);
      
      // Get agent configuration
      const agent = getAgent(agentType);
      const claudeCodeType = mapToClaudeCodeType(agentType);
      
      // Register agent in state manager
      this.stateManager.addAgent(agentId, {
        agentType,
        agent,
        claudeCodeType,
        taskDescription,
        options,
        metadata: {
          sessionId: options.sessionId || `session-${Date.now()}`,
          priority: options.priority || 'medium',
          timeout: options.timeout || this.config.timeoutMs
        }
      });
      
      // Execute pre-task hooks
      const preTaskResult = await HookManager.executePreTask(
        agentType,
        taskDescription,
        this.stateManager.activeAgents.get(agentId).metadata
      );
      
      this.stateManager.updateAgentStatus(agentId, 'hooks-pre-completed', {
        preTaskResult
      });
      
      // Optional MCP agent spawn
      if (options.useMCP !== false) {
        await this.swarmCoordinator.spawnMCPAgent(agentType);
      }
      
      // Prepare Claude Code Task call
      const taskConfig = this.buildTaskConfig(agent, taskDescription, options);
      this.stateManager.updateAgentStatus(agentId, 'ready-for-execution', {
        taskConfig
      });
      
      Logger.info(`Agent ${agentType} (${agentId}) ready for Claude Code Task execution`);
      
      return {
        agentId,
        agentType,
        claudeCodeType,
        taskConfig,
        preTaskResult,
        status: 'ready',
        execute: () => this.executeAgent(agentId),
        cleanup: () => this.cleanupAgent(agentId)
      };
      
    } catch (error) {
      Logger.error(`Failed to spawn agent ${agentType}:`, error.message);
      this.stateManager.updateAgentStatus(agentId, 'failed', { error: error.message });
      throw new AgentSpawnError(`Failed to spawn agent ${agentType}: ${error.message}`, agentType);
    }
  }

  /**
   * Spawn multiple agents in parallel
   */
  async spawnAgents(requests, options = {}) {
    Logger.info(`Spawning ${requests.length} agents in parallel`);
    
    try {
      const spawnPromises = requests.map(async (request) => {
        try {
          return await this.spawnAgent(
            request.agentType,
            request.taskDescription,
            { ...options, ...request.options }
          );
        } catch (error) {
          Logger.error(`Failed to spawn agent ${request.agentType}:`, error.message);
          return { error: error.message, agentType: request.agentType };
        }
      });
      
      const results = await Promise.allSettled(spawnPromises);
      
      const successful = results
        .filter(r => r.status === 'fulfilled' && !r.value.error)
        .map(r => r.value);
      
      const failed = results
        .filter(r => r.status === 'rejected' || (r.status === 'fulfilled' && r.value.error))
        .map(r => r.status === 'rejected' ? r.reason : r.value);
      
      Logger.info(`Agent spawning completed: ${successful.length} successful, ${failed.length} failed`);
      
      return { successful, failed, total: requests.length };
      
    } catch (error) {
      Logger.error('Batch agent spawning failed:', error.message);
      throw error;
    }
  }

  /**
   * Intelligent agent selection based on task analysis
   */
  async selectAndSpawnAgents(taskDescription, maxAgents = 3, options = {}) {
    Logger.info(`Analyzing task and selecting optimal agents: "${taskDescription}"`);
    
    try {
      const taskAnalysis = CapabilityMatcher.analyzeTask(taskDescription);
      const matchedAgents = CapabilityMatcher.matchAgents(taskAnalysis, maxAgents);
      
      if (matchedAgents.length === 0) {
        Logger.warn('No matching agents found, using fallback');
        const fallbackAgent = CapabilityMatcher.getFallbackAgent(taskDescription);
        matchedAgents.push({
          agentType: fallbackAgent,
          agent: getAgent(fallbackAgent),
          score: 0
        });
      }
      
      Logger.debug(`Selected agents: ${matchedAgents.map(a => `${a.agentType}(${a.score.toFixed(1)})`).join(', ')}`);
      
      const requests = matchedAgents.map(match => ({
        agentType: match.agentType,
        taskDescription: `${taskDescription}\n\nSpecific focus: ${match.agent.description}`,
        options: { 
          ...options,
          score: match.score,
          taskAnalysis
        }
      }));
      
      return await this.spawnAgents(requests, options);
      
    } catch (error) {
      Logger.error('Intelligent agent selection failed:', error.message);
      throw error;
    }
  }

  /**
   * Execute an agent (placeholder for Claude Code Task integration)
   */
  async executeAgent(agentId) {
    const agentState = this.stateManager.activeAgents.get(agentId);
    if (!agentState) {
      throw new BridgeError(`Agent ${agentId} not found`);
    }
    
    try {
      Logger.info(`Executing agent: ${agentState.agentType} (${agentId})`);
      this.stateManager.updateAgentStatus(agentId, 'executing');
      
      // This would be integrated with Claude Code's Task tool
      // Task(agentState.agent.name, agentState.taskDescription, agentState.claudeCodeType)
      
      // Simulated execution for now
      const executionResult = {
        agentId,
        agentType: agentState.agentType,
        status: 'completed',
        startTime: new Date().toISOString(),
        duration: 1000, // Simulated
        output: 'Task completed successfully'
      };
      
      this.stateManager.updateAgentStatus(agentId, 'completed', { executionResult });
      
      // Execute post-task hooks
      const postTaskResult = await HookManager.executePostTask(
        agentState.agentType,
        agentState.taskDescription,
        agentState.metadata
      );
      
      this.stateManager.updateAgentStatus(agentId, 'hooks-post-completed', {
        postTaskResult
      });
      
      Logger.info(`Agent execution completed: ${agentState.agentType} (${agentId})`);
      return { ...executionResult, postTaskResult };
      
    } catch (error) {
      Logger.error(`Agent execution failed for ${agentId}:`, error.message);
      this.stateManager.updateAgentStatus(agentId, 'failed', { error: error.message });
      throw error;
    }
  }

  /**
   * Cleanup agent resources
   */
  async cleanupAgent(agentId) {
    Logger.debug(`Cleaning up agent: ${agentId}`);
    
    try {
      const agentState = this.stateManager.activeAgents.get(agentId);
      if (agentState) {
        // Execute final cleanup hooks if needed
        await HookManager.executeSessionEnd(agentState.metadata.sessionId);
      }
      
      this.stateManager.removeAgent(agentId);
      Logger.debug(`Agent cleanup completed: ${agentId}`);
      
    } catch (error) {
      Logger.warn(`Agent cleanup failed for ${agentId}:`, error.message);
    }
  }

  /**
   * Build task configuration for Claude Code Task tool
   */
  buildTaskConfig(agent, taskDescription, options = {}) {
    const enhancedDescription = `
${taskDescription}

Agent: ${agent.name} (${agent.description})
Capabilities: ${agent.capabilities.join(', ')}
Tools: ${agent.tools.join(', ')}

Coordination Instructions:
1. Execute pre-task hooks for coordination
2. ${taskDescription}
3. Store progress in swarm memory using: npx claude-flow@alpha hooks post-edit --file "[file]" --memory-key "swarm/${agent.name}/[step]"
4. Notify other agents of progress: npx claude-flow@alpha hooks notify --message "[what was done]"
5. Execute post-task hooks for cleanup

Priority: ${options.priority || 'medium'}
Timeout: ${options.timeout || this.config.timeoutMs}ms
`.trim();

    return {
      name: agent.name,
      description: enhancedDescription,
      type: agent.type || 'general-purpose',
      tools: agent.tools,
      specializations: agent.specializations,
      hooks: agent.hooks,
      metadata: {
        originalAgentType: agent,
        taskAnalysis: options.taskAnalysis,
        score: options.score,
        sessionId: options.sessionId
      }
    };
  }

  /**
   * Shutdown bridge and cleanup resources
   */
  async shutdown() {
    try {
      // Save state before shutdown
      await this.stateManager.saveState();
      
      // Clear active agents
      this.stateManager.activeAgents.clear();
      
      // Reset initialization
      this.initialized = false;
      
      if (this.logger) {
        this.logger.info('Bridge shutdown complete');
      }
    } catch (error) {
      if (this.logger) {
        this.logger.error('Error during bridge shutdown:', error);
      }
      throw error;
    }
  }

  /**
   * Get bridge status and statistics
   */
  getStatus() {
    const activeAgents = this.stateManager.getActiveAgents();
    const swarmState = this.stateManager.swarmState;
    
    return {
      initialized: this.initialized,
      swarm: swarmState,
      agents: {
        active: activeAgents.length,
        total: this.stateManager.executionHistory.length + activeAgents.length,
        byStatus: activeAgents.reduce((acc, agent) => {
          acc[agent.status] = (acc[agent.status] || 0) + 1;
          return acc;
        }, {})
      },
      performance: {
        averageSpawnTime: this.calculateAverageSpawnTime(),
        successRate: this.calculateSuccessRate()
      },
      lastActivity: this.stateManager.executionHistory.slice(-5)
    };
  }

  calculateAverageSpawnTime() {
    const completed = this.stateManager.executionHistory.slice(-10);
    if (completed.length === 0) return 0;
    
    const times = completed
      .filter(h => h.config.metadata?.executionResult?.duration)
      .map(h => h.config.metadata.executionResult.duration);
    
    return times.length > 0 ? times.reduce((a, b) => a + b, 0) / times.length : 0;
  }

  calculateSuccessRate() {
    const recent = this.stateManager.executionHistory.slice(-20);
    if (recent.length === 0) return 100;
    
    const successful = recent.filter(h => h.config.status === 'completed').length;
    return (successful / recent.length) * 100;
  }

  /**
   * Cleanup all resources
   */
  async cleanup() {
    Logger.info('Cleaning up Claude Flow Bridge...');
    
    try {
      const activeAgents = Array.from(this.stateManager.activeAgents.keys());
      await Promise.all(activeAgents.map(agentId => this.cleanupAgent(agentId)));
      
      this.stateManager.saveState();
      Logger.info('Bridge cleanup completed');
      
    } catch (error) {
      Logger.error('Bridge cleanup failed:', error.message);
      throw error;
    }
  }
}

/**
 * Factory Functions
 */
export function createBridge(options = {}) {
  return new ClaudeFlowBridge(options);
}

export async function quickSpawn(agentType, taskDescription, options = {}) {
  const bridge = createBridge();
  await bridge.initialize();
  return await bridge.spawnAgent(agentType, taskDescription, options);
}

export async function intelligentSpawn(taskDescription, maxAgents = 3, options = {}) {
  const bridge = createBridge();
  await bridge.initialize();
  return await bridge.selectAndSpawnAgents(taskDescription, maxAgents, options);
}

/**
 * CLI Integration (if called directly)
 */
if (import.meta.url === `file://${process.argv[1]}`) {
  const command = process.argv[2];
  const args = process.argv.slice(3);
  
  switch (command) {
    case 'spawn':
      if (args.length < 2) {
        console.error('Usage: claude-flow-bridge.js spawn <agent-type> <task-description>');
        process.exit(1);
      }
      quickSpawn(args[0], args.slice(1).join(' '))
        .then(result => {
          console.log('Agent spawned successfully:', result.agentId);
          console.log('Execute with:', `agent.execute()`);
        })
        .catch(error => {
          console.error('Spawn failed:', error.message);
          process.exit(1);
        });
      break;
      
    case 'intelligent':
      if (args.length < 1) {
        console.error('Usage: claude-flow-bridge.js intelligent <task-description>');
        process.exit(1);
      }
      intelligentSpawn(args.join(' '))
        .then(result => {
          console.log(`Spawned ${result.successful.length} agents successfully`);
          result.successful.forEach(agent => {
            console.log(`- ${agent.agentType} (${agent.agentId})`);
          });
        })
        .catch(error => {
          console.error('Intelligent spawn failed:', error.message);
          process.exit(1);
        });
      break;
      
    case 'status':
      const bridge = createBridge();
      const status = bridge.getStatus();
      console.log('Bridge Status:', JSON.stringify(status, null, 2));
      break;
      
    default:
      console.log(`
Claude Flow Bridge v1.0.0

Usage:
  claude-flow-bridge.js spawn <agent-type> <task-description>
  claude-flow-bridge.js intelligent <task-description>
  claude-flow-bridge.js status

Examples:
  claude-flow-bridge.js spawn coder "Implement REST API endpoints"
  claude-flow-bridge.js intelligent "Build a secure authentication system"
  claude-flow-bridge.js status
      `);
  }
}

export default ClaudeFlowBridge;