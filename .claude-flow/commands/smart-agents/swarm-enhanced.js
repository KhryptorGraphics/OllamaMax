#!/usr/bin/env node

/**
 * Enhanced Smart Agents Swarm with Complete Integration
 * Combines all components for a fully functional hive-mind system
 */

const { spawn } = require('child_process');
const fs = require('fs').promises;
const path = require('path');
const { performance } = require('perf_hooks');

// Import specialized components
const NeuralLearningSystem = require('../../agents/neural-learning');
const { AgentSpecializations, AgentSelector } = require('../../agents/agent-specializations');
const ClaudeAgentIntegration = require('./claude-integration');
const SPARCIntegration = require('./sparc-integration');

class EnhancedSmartAgentsSwarm {
  constructor(options = {}) {
    // Core configuration
    this.maxAgents = options.maxAgents || 25;
    this.minAgents = options.minAgents || 8;
    this.currentAgents = 0;
    
    // Agent management
    this.activeAgents = new Map();
    this.taskQueue = [];
    this.completedTasks = [];
    
    // Specialized systems
    this.neuralLearning = new NeuralLearningSystem(path.join(__dirname, '../../memory'));
    this.agentSelector = new AgentSelector();
    this.claudeIntegration = new ClaudeAgentIntegration(this);
    this.sparcIntegration = new SPARCIntegration(this);
    
    // Performance metrics
    this.metrics = {
      tasksCompleted: 0,
      totalExecutionTime: 0,
      efficiency: 0,
      swarmHealth: 100,
      neuralLearnings: 0,
      adaptations: 0
    };

    // Auto-scaling configuration
    this.scalingConfig = {
      scaleUpThreshold: 0.8,    // Scale up when efficiency drops below 80%
      scaleDownThreshold: 0.3,  // Scale down when load is below 30%
      evaluationInterval: 10000, // Check every 10 seconds
      cooldownPeriod: 30000,    // Wait 30s between scaling operations
      lastScalingAction: 0
    };

    this.initializeSwarm();
  }

  async initializeSwarm() {
    console.log('🚀 Initializing Enhanced Smart Agents Swarm...');
    
    // Initialize all subsystems
    await this.neuralLearning.initializeLearningSystem();
    await this.setupMetricsCollection();
    await this.setupAutoScaling();
    
    console.log(`✅ Enhanced swarm initialized with ${this.minAgents}-${this.maxAgents} agent capacity`);
    console.log(`🧠 Neural learning: ${this.neuralLearning.learningData.size} patterns loaded`);
    console.log(`🎯 Agent specializations: ${Object.keys(AgentSpecializations).length} types available`);
  }

  /**
   * Enhanced task analysis with neural learning integration
   */
  async analyzeTask(task, context = {}) {
    console.log(`🔍 Enhanced task analysis: "${task.substring(0, 100)}..."`);
    
    // Basic complexity analysis
    const baseComplexity = this.calculateTaskComplexity(task);
    
    // Neural learning enhancement
    const neuralInsights = await this.getNeuralInsights(task, baseComplexity);
    
    // Final complexity with learning adjustments
    const adjustedComplexity = Math.min(1.0, baseComplexity * neuralInsights.complexityMultiplier);
    
    // Determine optimal agent configuration
    const requiredSpecializations = this.determineRequiredSpecializations(task, neuralInsights);
    const estimatedAgentCount = this.calculateOptimalAgentCount(adjustedComplexity, requiredSpecializations);
    
    // Task type classification
    const taskType = this.classifyTaskType(task);
    
    return {
      complexity: adjustedComplexity,
      baseComplexity,
      neuralInsights,
      requiredSpecializations,
      estimatedAgentCount,
      taskType,
      priority: this.calculateTaskPriority(task, neuralInsights),
      recommendations: this.generateTaskRecommendations(task, neuralInsights)
    };
  }

  /**
   * Get neural learning insights for task optimization
   */
  async getNeuralInsights(task, baseComplexity) {
    const insights = {
      complexityMultiplier: 1.0,
      recommendedSpecializations: [],
      successProbability: 0.7,
      estimatedExecutionTime: 5000,
      riskFactors: [],
      optimizations: []
    };

    // Analyze against learned patterns
    const taskPatterns = this.extractTaskPatterns(task);
    
    for (const pattern of taskPatterns) {
      const patternKey = `success-${pattern}`;
      const learningData = this.neuralLearning.learningData.get(patternKey);
      
      if (learningData && learningData.confidence > 0.6) {
        insights.successProbability = Math.max(insights.successProbability, learningData.successRate);
        insights.estimatedExecutionTime = Math.min(insights.estimatedExecutionTime, learningData.avgExecutionTime);
        
        // Adjust complexity based on learning
        if (learningData.successRate > 0.8) {
          insights.complexityMultiplier *= 0.9; // Reduce complexity for well-learned patterns
        } else if (learningData.successRate < 0.5) {
          insights.complexityMultiplier *= 1.2; // Increase complexity for problematic patterns
          insights.riskFactors.push(`Pattern ${pattern} has low success rate`);
        }
      }
    }

    // Get optimization recommendations
    insights.optimizations = await this.neuralLearning.getLearningRecommendations();
    
    return insights;
  }

  /**
   * Extract meaningful patterns from task description
   */
  extractTaskPatterns(task) {
    const patterns = [];
    const taskLower = task.toLowerCase();
    
    // Technical patterns
    const techPatterns = {
      'api-development': ['api', 'endpoint', 'rest', 'graphql'],
      'database-work': ['database', 'sql', 'query', 'schema'],
      'frontend-development': ['ui', 'frontend', 'react', 'vue', 'angular'],
      'backend-development': ['backend', 'server', 'microservice'],
      'security-implementation': ['security', 'auth', 'authentication', 'authorization'],
      'performance-optimization': ['performance', 'optimize', 'fast', 'speed'],
      'testing-implementation': ['test', 'testing', 'quality', 'validation'],
      'deployment-automation': ['deploy', 'ci', 'cd', 'docker', 'kubernetes']
    };

    Object.entries(techPatterns).forEach(([pattern, keywords]) => {
      if (keywords.some(keyword => taskLower.includes(keyword))) {
        patterns.push(pattern);
      }
    });

    return patterns;
  }

  /**
   * Calculate optimal agent count with neural learning
   */
  calculateOptimalAgentCount(complexity, specializations) {
    // Base calculation
    let baseCount = Math.ceil(complexity * 10 + specializations.length);
    
    // Neural learning adjustments
    const learningInsights = this.agentSelector.getLearningInsights();
    
    // Adjust based on historical performance
    if (learningInsights.mostSuccessfulCombinations.length > 0) {
      const avgSuccessfulSize = learningInsights.mostSuccessfulCombinations
        .reduce((sum, combo) => sum + combo.agentCount, 0) / learningInsights.mostSuccessfulCombinations.length;
      
      baseCount = Math.round((baseCount + avgSuccessfulSize) / 2);
    }

    // Ensure within bounds
    return Math.max(this.minAgents, Math.min(this.maxAgents, baseCount));
  }

  /**
   * Enhanced agent spawning with specialization optimization
   */
  async spawnOptimalAgentTeam(taskAnalysis) {
    console.log(`🤖 Spawning optimal agent team for complexity ${(taskAnalysis.complexity * 100).toFixed(1)}%`);
    
    // Select optimal agent configuration
    const selectedAgents = this.agentSelector.selectAgents(taskAnalysis, {
      maxAgents: taskAnalysis.estimatedAgentCount
    });

    console.log(`🎯 Selected ${selectedAgents.length} specialized agents:`);
    selectedAgents.forEach(agent => {
      console.log(`   - ${agent.specialization} (priority: ${agent.priority}, role: ${agent.role})`);
    });

    // Spawn agents in parallel
    const spawnPromises = selectedAgents.map(async (agentConfig) => {
      return await this.spawnSpecializedAgent(agentConfig, taskAnalysis);
    });

    try {
      const spawnedAgents = await Promise.all(spawnPromises);
      console.log(`✅ Successfully spawned ${spawnedAgents.length} agents in parallel`);
      
      return spawnedAgents;
    } catch (error) {
      console.error('❌ Failed to spawn agent team:', error.message);
      throw error;
    }
  }

  /**
   * Spawn a specialized agent with enhanced configuration
   */
  async spawnSpecializedAgent(agentConfig, taskAnalysis) {
    const agentId = `agent-${Date.now()}-${agentConfig.specialization}-${Math.random().toString(36).substr(2, 6)}`;
    
    const agentData = {
      id: agentId,
      specialization: agentConfig.specialization,
      priority: agentConfig.priority,
      role: agentConfig.role,
      status: 'spawning',
      spawnTime: Date.now(),
      taskAnalysis,
      configuration: AgentSpecializations[agentConfig.specialization]
    };

    console.log(`🤖 Spawning ${agentConfig.specialization} agent (${agentId})`);
    
    this.activeAgents.set(agentId, agentData);
    this.currentAgents++;

    try {
      // Execute agent using Claude integration
      const result = await this.claudeIntegration.executeAgent(
        agentConfig.specialization,
        {
          task: taskAnalysis.task || 'Enhanced agent execution',
          complexity: taskAnalysis.complexity,
          priority: taskAnalysis.priority,
          taskType: taskAnalysis.taskType
        },
        {
          totalAgents: this.currentAgents,
          parallelMode: true,
          learningEnabled: true
        }
      );

      // Update agent status
      agentData.status = result.success ? 'completed' : 'failed';
      agentData.result = result;
      agentData.completionTime = Date.now();

      // Process learning data
      await this.neuralLearning.processLearningData({
        agentId,
        specialization: agentConfig.specialization,
        taskData: taskAnalysis,
        success: result.success,
        executionTime: result.executionTime,
        output: result.output,
        learningData: result.learningData
      });

      return result;
    } catch (error) {
      agentData.status = 'failed';
      agentData.error = error.message;
      console.error(`❌ Agent ${agentId} execution failed:`, error.message);
      throw error;
    }
  }

  /**
   * Auto-scaling system with neural learning
   */
  async setupAutoScaling() {
    setInterval(() => this.evaluateAutoScaling(), this.scalingConfig.evaluationInterval);
  }

  async evaluateAutoScaling() {
    const now = Date.now();
    
    // Check cooldown period
    if (now - this.scalingConfig.lastScalingAction < this.scalingConfig.cooldownPeriod) {
      return;
    }

    // Calculate current metrics
    const workloadRatio = this.taskQueue.length / Math.max(this.currentAgents, 1);
    const efficiency = this.calculateSwarmEfficiency();
    const neuralRecommendations = this.neuralLearning.getLearningRecommendations();

    // Neural learning insights for scaling
    const scalingInsights = this.analyzeScalingPatterns();

    // Scale up conditions
    if ((workloadRatio > 2 || efficiency < this.scalingConfig.scaleUpThreshold) && 
        this.currentAgents < this.maxAgents) {
      await this.intelligentScaleUp(scalingInsights);
    }
    // Scale down conditions
    else if (workloadRatio < this.scalingConfig.scaleDownThreshold && 
             efficiency > 0.8 && 
             this.currentAgents > this.minAgents) {
      await this.intelligentScaleDown(scalingInsights);
    }
  }

  /**
   * Analyze scaling patterns from neural learning
   */
  analyzeScalingPatterns() {
    const insights = {
      optimalSize: this.minAgents,
      recommendedSpecializations: [],
      scalingTrend: 'stable',
      efficiency: 0.7
    };

    // Analyze historical scaling success
    const learningInsights = this.agentSelector.getLearningInsights();
    
    if (learningInsights.performancePatterns.size > 0) {
      let totalAgents = 0;
      let totalTasks = 0;
      
      learningInsights.performancePatterns.forEach((pattern) => {
        totalAgents += pattern.frequency;
        totalTasks += pattern.frequency;
      });
      
      insights.optimalSize = Math.round(totalAgents / Math.max(totalTasks, 1));
    }

    return insights;
  }

  /**
   * Intelligent scale up with specialization selection
   */
  async intelligentScaleUp(insights) {
    const additionalAgents = Math.min(3, this.maxAgents - this.currentAgents);
    console.log(`📈 Intelligent scale up: Adding ${additionalAgents} agents`);
    
    // Select most needed specializations
    const neededSpecializations = this.selectScalingSpecializations(additionalAgents, 'up');
    
    for (const specialization of neededSpecializations) {
      await this.spawnSpecializedAgent(
        { specialization, priority: 7, role: 'scaler' },
        { 
          task: 'Support swarm workload',
          complexity: 0.5,
          priority: 5,
          taskType: 'scaling'
        }
      );
    }

    this.scalingConfig.lastScalingAction = Date.now();
    this.metrics.adaptations++;
  }

  /**
   * Intelligent scale down with performance preservation
   */
  async intelligentScaleDown(insights) {
    const agentsToRemove = Math.min(2, this.currentAgents - this.minAgents);
    console.log(`📉 Intelligent scale down: Removing ${agentsToRemove} agents`);
    
    // Select least effective agents for removal
    const sortedAgents = Array.from(this.activeAgents.values())
      .filter(agent => agent.status === 'idle' || agent.status === 'completed')
      .sort((a, b) => {
        const aEffectiveness = this.calculateAgentEffectiveness(a);
        const bEffectiveness = this.calculateAgentEffectiveness(b);
        return aEffectiveness - bEffectiveness;
      });

    for (let i = 0; i < Math.min(agentsToRemove, sortedAgents.length); i++) {
      const agent = sortedAgents[i];
      this.activeAgents.delete(agent.id);
      this.currentAgents--;
      console.log(`🔻 Removed agent ${agent.id} (${agent.specialization})`);
    }

    this.scalingConfig.lastScalingAction = Date.now();
    this.metrics.adaptations++;
  }

  /**
   * Select specializations for scaling operations
   */
  selectScalingSpecializations(count, direction) {
    const learningInsights = this.agentSelector.getLearningInsights();
    const specializations = [];

    if (direction === 'up') {
      // Add most effective specializations
      const effective = Array.from(learningInsights.specializationEfficiency.entries())
        .filter(([spec, data]) => data.successfulTasks / data.totalTasks > 0.7)
        .sort((a, b) => (b[1].successfulTasks / b[1].totalTasks) - (a[1].successfulTasks / a[1].totalTasks))
        .slice(0, count);
      
      specializations.push(...effective.map(([spec]) => spec));
    }

    // Fill with general-purpose if needed
    while (specializations.length < count) {
      specializations.push('general-purpose');
    }

    return specializations;
  }

  /**
   * Calculate agent effectiveness for scaling decisions
   */
  calculateAgentEffectiveness(agent) {
    if (!agent.result) return 0;
    
    const successScore = agent.result.success ? 1 : 0;
    const speedScore = agent.result.executionTime ? Math.max(0, 1 - (agent.result.executionTime / 10000)) : 0.5;
    
    return (successScore * 0.7 + speedScore * 0.3);
  }

  /**
   * Enhanced metrics collection with neural learning
   */
  async setupMetricsCollection() {
    setInterval(() => this.collectEnhancedMetrics(), 5000);
  }

  async collectEnhancedMetrics() {
    const baseMetrics = {
      timestamp: Date.now(),
      activeAgents: this.currentAgents,
      maxCapacity: this.maxAgents,
      tasksCompleted: this.completedTasks.length,
      taskQueueLength: this.taskQueue.length
    };

    // Neural learning metrics
    const learningReport = this.neuralLearning.generateLearningReport();
    
    // Enhanced swarm metrics
    this.metrics = {
      ...baseMetrics,
      efficiency: this.calculateSwarmEfficiency(),
      swarmHealth: this.assessEnhancedSwarmHealth(),
      neuralMetrics: {
        totalPatterns: learningReport.summary.totalPatterns,
        highConfidencePatterns: learningReport.summary.memoryUtilization.highConfidencePatterns,
        learningRate: learningReport.learningTrends.last24h?.learningRate || 0
      },
      specializationDistribution: this.getSpecializationDistribution(),
      performanceInsights: this.getPerformanceInsights()
    };

    // Save enhanced metrics
    const metricsPath = path.join(__dirname, '../../metrics/enhanced-swarm-metrics.json');
    await fs.writeFile(metricsPath, JSON.stringify(this.metrics, null, 2));
  }

  /**
   * Enhanced swarm health assessment
   */
  assessEnhancedSwarmHealth() {
    const factors = {
      agentUtilization: Math.min(1, this.currentAgents / (this.maxAgents * 0.8)),
      taskSuccess: this.calculateTaskSuccessRate(),
      neuralLearning: Math.min(1, this.neuralLearning.learningData.size / 100),
      efficiency: this.calculateSwarmEfficiency(),
      adaptability: Math.min(1, this.metrics.adaptations / 10)
    };

    const weights = {
      agentUtilization: 0.2,
      taskSuccess: 0.3,
      neuralLearning: 0.2,
      efficiency: 0.2,
      adaptability: 0.1
    };

    const overallHealth = Object.entries(factors).reduce((sum, [key, value]) => {
      return sum + (value * weights[key]);
    }, 0);

    return Math.round(overallHealth * 100);
  }

  /**
   * Get specialization distribution across active agents
   */
  getSpecializationDistribution() {
    const distribution = {};
    
    this.activeAgents.forEach(agent => {
      const spec = agent.specialization;
      distribution[spec] = (distribution[spec] || 0) + 1;
    });

    return distribution;
  }

  /**
   * Get performance insights from neural learning
   */
  getPerformanceInsights() {
    const learningInsights = this.agentSelector.getLearningInsights();
    const insights = {
      topPerformers: [],
      improvementOpportunities: [],
      emergingPatterns: []
    };

    // Top performing specializations
    if (learningInsights.specializationEfficiency.size > 0) {
      insights.topPerformers = Array.from(learningInsights.specializationEfficiency.entries())
        .filter(([, data]) => data.totalTasks > 5)
        .sort((a, b) => (b[1].successfulTasks / b[1].totalTasks) - (a[1].successfulTasks / a[1].totalTasks))
        .slice(0, 3)
        .map(([spec, data]) => ({
          specialization: spec,
          successRate: data.successfulTasks / data.totalTasks,
          totalTasks: data.totalTasks
        }));
    }

    return insights;
  }

  /**
   * Execute enhanced swarm with full integration
   */
  async executeEnhancedSwarm(task, options = {}) {
    console.log('\n🚀 Enhanced Smart Agents Swarm Executing...\n');
    
    // Enhanced task analysis
    const analysis = await this.analyzeTask(task, options);
    
    console.log(`📊 Enhanced Task Analysis:
    - Base Complexity: ${(analysis.baseComplexity * 100).toFixed(1)}%
    - Adjusted Complexity: ${(analysis.complexity * 100).toFixed(1)}%
    - Neural Success Probability: ${(analysis.neuralInsights.successProbability * 100).toFixed(1)}%
    - Optimal Agent Count: ${analysis.estimatedAgentCount}
    - Task Type: ${analysis.taskType}
    - Priority: ${analysis.priority}/10`);

    if (analysis.neuralInsights.riskFactors.length > 0) {
      console.log(`⚠️  Risk Factors Identified:`);
      analysis.neuralInsights.riskFactors.forEach(risk => console.log(`   - ${risk}`));
    }

    try {
      // Spawn optimal agent team
      const agentResults = await this.spawnOptimalAgentTeam(analysis);
      
      console.log(`\n✅ All ${agentResults.length} agents completed successfully`);
      
      // Update agent selector with results
      this.agentSelector.updateSelectionHistory(analysis, agentResults.map(r => ({
        specialization: r.specialization,
        success: r.success
      })), agentResults);

      // Generate comprehensive report
      const report = await this.generateEnhancedExecutionReport(analysis, agentResults);
      console.log('\n📋 Enhanced Execution Report Generated');

      return {
        success: true,
        agentsUsed: agentResults.length,
        swarmHealth: this.assessEnhancedSwarmHealth(),
        neuralLearnings: this.neuralLearning.learningData.size,
        analysis,
        report
      };
      
    } catch (error) {
      console.error('\n❌ Enhanced swarm execution failed:', error.message);
      return {
        success: false,
        error: error.message,
        swarmHealth: this.assessEnhancedSwarmHealth(),
        analysis
      };
    }
  }

  /**
   * Generate comprehensive execution report
   */
  async generateEnhancedExecutionReport(analysis, results) {
    const learningReport = this.neuralLearning.generateLearningReport();
    
    const report = {
      timestamp: new Date().toISOString(),
      executionSummary: {
        taskAnalysis: analysis,
        agentsDeployed: results.length,
        successRate: results.filter(r => r.success).length / results.length,
        totalExecutionTime: results.reduce((sum, r) => sum + (r.executionTime || 0), 0),
        efficiency: this.calculateSwarmEfficiency()
      },
      swarmConfiguration: {
        totalAgents: this.currentAgents,
        maxCapacity: this.maxAgents,
        specializations: [...new Set(results.map(r => r.specialization))],
        scalingEvents: this.metrics.adaptations
      },
      neuralLearningInsights: {
        patternsLearned: learningReport.summary.totalPatterns,
        confidencePatterns: learningReport.summary.memoryUtilization.highConfidencePatterns,
        recommendations: learningReport.recommendations,
        topPatterns: learningReport.topPerformingPatterns
      },
      performanceMetrics: this.metrics,
      futureOptimizations: this.generateFutureOptimizations(analysis, results)
    };

    // Save report
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const reportPath = path.join(__dirname, '../../metrics', `enhanced-execution-report-${timestamp}.json`);
    await fs.writeFile(reportPath, JSON.stringify(report, null, 2));

    return report;
  }

  /**
   * Generate future optimization recommendations
   */
  generateFutureOptimizations(analysis, results) {
    const optimizations = [];

    // Analyze performance patterns
    const avgExecutionTime = results.reduce((sum, r) => sum + (r.executionTime || 0), 0) / results.length;
    
    if (avgExecutionTime > analysis.neuralInsights.estimatedExecutionTime * 1.5) {
      optimizations.push({
        type: 'performance',
        priority: 8,
        suggestion: 'Execution time exceeded neural predictions - review agent efficiency',
        targetArea: 'agent-optimization'
      });
    }

    // Neural learning recommendations
    const neuralRecommendations = this.neuralLearning.getLearningRecommendations();
    optimizations.push(...neuralRecommendations.map(rec => ({
      ...rec,
      source: 'neural-learning'
    })));

    return optimizations.sort((a, b) => b.priority - a.priority);
  }

  // Helper methods
  calculateSwarmEfficiency() {
    if (this.activeAgents.size === 0) return 0.5;
    
    let totalEfficiency = 0;
    let validAgents = 0;

    this.activeAgents.forEach(agent => {
      if (agent.result) {
        const successScore = agent.result.success ? 1 : 0;
        const speedScore = agent.result.executionTime ? 
          Math.max(0, 1 - (agent.result.executionTime / 10000)) : 0.5;
        
        totalEfficiency += (successScore * 0.7 + speedScore * 0.3);
        validAgents++;
      }
    });

    return validAgents > 0 ? totalEfficiency / validAgents : 0.5;
  }

  calculateTaskSuccessRate() {
    if (this.completedTasks.length === 0) return 0.7;
    
    const successfulTasks = this.completedTasks.filter(task => task.success).length;
    return successfulTasks / this.completedTasks.length;
  }

  calculateTaskComplexity(task) {
    const complexityKeywords = {
      high: ['architecture', 'system', 'distributed', 'microservices', 'security', 'performance', 'scalable'],
      medium: ['api', 'database', 'frontend', 'backend', 'testing', 'integration', 'deployment'],
      low: ['bug', 'fix', 'update', 'documentation', 'style', 'format']
    };

    let complexity = 0.3; // base complexity
    
    Object.entries(complexityKeywords).forEach(([level, keywords]) => {
      const matches = keywords.filter(keyword => 
        task.toLowerCase().includes(keyword.toLowerCase())
      ).length;
      
      switch(level) {
        case 'high': complexity += matches * 0.4; break;
        case 'medium': complexity += matches * 0.2; break;
        case 'low': complexity += matches * 0.1; break;
      }
    });

    return Math.min(1.0, complexity);
  }

  determineRequiredSpecializations(task, neuralInsights = {}) {
    const specializations = [];
    const taskLower = task.toLowerCase();

    // Core specialization detection
    const specializationMap = {
      'system-architect': ['architecture', 'system', 'distributed', 'scalable'],
      'backend-architect': ['backend', 'api', 'server', 'database'],
      'frontend-architect': ['frontend', 'ui', 'interface', 'react', 'vue'],
      'security-engineer': ['security', 'auth', 'authentication', 'secure'],
      'performance-engineer': ['performance', 'optimize', 'fast', 'speed'],
      'quality-engineer': ['test', 'quality', 'validation', 'QA'],
      'devops-architect': ['deploy', 'ci', 'cd', 'docker', 'kubernetes'],
      'python-expert': ['python', 'django', 'flask', 'fastapi']
    };

    Object.entries(specializationMap).forEach(([spec, keywords]) => {
      if (keywords.some(keyword => taskLower.includes(keyword))) {
        specializations.push(spec);
      }
    });

    // Neural learning recommendations
    if (neuralInsights.recommendedSpecializations) {
      specializations.push(...neuralInsights.recommendedSpecializations);
    }

    // Always include general-purpose for coordination
    specializations.push('general-purpose');

    return [...new Set(specializations)];
  }

  classifyTaskType(task) {
    const taskLower = task.toLowerCase();
    
    if (taskLower.includes('build') || taskLower.includes('create')) return 'development';
    if (taskLower.includes('fix') || taskLower.includes('bug')) return 'maintenance';
    if (taskLower.includes('optimize') || taskLower.includes('performance')) return 'optimization';
    if (taskLower.includes('deploy') || taskLower.includes('release')) return 'deployment';
    if (taskLower.includes('test') || taskLower.includes('quality')) return 'testing';
    if (taskLower.includes('design') || taskLower.includes('architecture')) return 'architecture';
    
    return 'general';
  }

  calculateTaskPriority(task, neuralInsights = {}) {
    const urgentKeywords = ['critical', 'urgent', 'fix', 'bug', 'error', 'security', 'production'];
    const taskLower = task.toLowerCase();
    
    let priority = 5; // base priority
    
    const urgentMatches = urgentKeywords.filter(keyword => 
      taskLower.includes(keyword)
    ).length;

    priority += urgentMatches * 2;

    // Neural learning adjustment
    if (neuralInsights.successProbability < 0.5) {
      priority += 2; // Increase priority for challenging tasks
    }

    return Math.min(10, Math.max(1, priority));
  }

  generateTaskRecommendations(task, neuralInsights) {
    const recommendations = [];

    if (neuralInsights.successProbability < 0.6) {
      recommendations.push('Consider breaking down this complex task into smaller components');
    }

    if (neuralInsights.riskFactors.length > 0) {
      recommendations.push('Review risk factors before execution');
    }

    if (neuralInsights.optimizations.length > 0) {
      recommendations.push('Apply neural learning optimizations for better results');
    }

    return recommendations;
  }
}

module.exports = EnhancedSmartAgentsSwarm;