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
const AgentPerformanceForecaster = require('../../src/agents/agent-performance-forecaster');
const AgentLSTMPredictor = require('../../src/agents/agent-lstm-predictor');
const PredictiveScalingSystem = require('../../src/ml/predictive-scaling');
const Redis = require('ioredis');

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
    this.performanceForecaster = new AgentPerformanceForecaster();
    this.lstmPredictor = new AgentLSTMPredictor();
    this.predictiveScaling = new PredictiveScalingSystem();

    // Redis for ML pub/sub integration
    const redisNodes = process.env.REDIS_NODES
      ? JSON.parse(process.env.REDIS_NODES)
      : [
          { host: 'redis-cluster-0.redis-cluster-service.ollamamax-redis', port: 6379 },
          { host: 'redis-cluster-1.redis-cluster-service.ollamamax-redis', port: 6379 },
          { host: 'redis-cluster-2.redis-cluster-service.ollamamax-redis', port: 6379 }
        ];

    this.redis = new Redis.Cluster(redisNodes, {
      redisOptions: {
        password: process.env.REDIS_PASSWORD || 'ollama_redis_pass',
        connectTimeout: 10000
      }
    });

    // Subscribe to predictive scaling commands
    this.redis.subscribe('swarm:scaling_command', (err) => {
      if (err) {
        console.error('❌ Failed to subscribe to scaling commands:', err);
      } else {
        console.log('✅ Subscribed to swarm:scaling_command channel');
      }
    });

    this.redis.on('message', (channel, message) => {
      if (channel === 'swarm:scaling_command') {
        this.handleScalingCommand(JSON.parse(message));
      }
    });
    
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
    await this.performanceForecaster.initialize();
    await this.setupMetricsCollection();
    await this.setupAutoScaling();

    console.log(`✅ Enhanced swarm initialized with ${this.minAgents}-${this.maxAgents} agent capacity`);
    console.log(`🧠 Neural learning: ${this.neuralLearning.learningData.size} patterns loaded`);
    console.log(`🎯 Agent specializations: ${Object.keys(AgentSpecializations).length} types available`);
    console.log(`📊 ML forecaster and LSTM predictor initialized`);
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

    // Use ML forecaster to rank agents by predicted performance
    const availableAgents = Array.from(this.activeAgents.keys());
    const rankedAgents = await this.performanceForecaster.rankAgentsByPredictedPerformance(
      { type: taskAnalysis.taskType, complexity: taskAnalysis.complexity },
      availableAgents
    );

    // Select optimal agent configuration using both traditional and ML-based selection
    const selectedAgents = this.agentSelector.selectAgents(taskAnalysis, {
      maxAgents: taskAnalysis.estimatedAgentCount,
      mlRankings: rankedAgents
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

    // Get LSTM load forecasts for scaling decisions
    const loadForecasts = await this.getLSTMLoadForecasts();

    // Get Random Forest agent selection predictions
    const rfPredictions = await this.getRFPredictions();

    // Get predictive scaling system recommendations
    const predictiveScalingStatus = await this.predictiveScaling.getSystemStatus();
    const mlPredictions = predictiveScalingStatus.recentPredictions?.[0] || null;

    // Neural learning insights for scaling
    const scalingInsights = this.analyzeScalingPatterns();

    // Combine all ML forecasts for enhanced decision making
    const combinedForecast = this.combineMLForecasts({
      lstm: loadForecasts,
      randomForest: rfPredictions,
      predictiveScaling: mlPredictions,
      neuralLearning: neuralRecommendations
    });

    // Enhanced scaling logic with ML integration
    const scalingDecision = this.makeMLEnhancedScalingDecision({
      workloadRatio,
      efficiency,
      combinedForecast,
      scalingInsights,
      currentAgents: this.currentAgents
    });

    // Execute scaling based on ML-enhanced decision
    if (scalingDecision.action === 'scale_up' && this.currentAgents < this.maxAgents) {
      await this.intelligentScaleUp(scalingInsights, combinedForecast);
    } else if (scalingDecision.action === 'scale_down' && this.currentAgents > this.minAgents) {
      await this.intelligentScaleDown(scalingInsights, combinedForecast);
    }

    // Log ML-enhanced decision
    if (scalingDecision.action !== 'none') {
      console.log(`🤖 ML-enhanced scaling decision: ${scalingDecision.action}`);
      console.log(`   📊 Combined forecast confidence: ${(scalingDecision.confidence * 100).toFixed(1)}%`);
      console.log(`   🎯 Reason: ${scalingDecision.reason}`);
    }
  }

  /**
   * Combine ML forecasts from multiple models (LSTM, RF, Predictive Scaling)
   */
  combineMLForecasts(forecasts) {
    const combined = {
      predictedLoad: 0,
      predictedAgents: 0,
      predictedQueueLength: 0,
      predictedResponseTime: 0,
      confidence: 0,
      models: []
    };

    let totalWeight = 0;

    // LSTM predictions (short-term 5-15 min, weight 0.35)
    if (forecasts.lstm && forecasts.lstm.length > 0) {
      const lstm = forecasts.lstm[0];
      combined.predictedLoad += lstm.load * 0.35;
      combined.confidence += lstm.confidence * 0.35;
      combined.models.push('LSTM');
      totalWeight += 0.35;
    }

    // Random Forest predictions (medium-term 1-4 hours, weight 0.30)
    if (forecasts.randomForest && forecasts.randomForest.successRate) {
      combined.predictedLoad += forecasts.randomForest.successRate * 0.30;
      combined.confidence += (forecasts.randomForest.confidence || 0.8) * 0.30;
      combined.models.push('RandomForest');
      totalWeight += 0.30;
    }

    // Predictive Scaling predictions (deep learning LSTM, weight 0.35)
    if (forecasts.predictiveScaling && forecasts.predictiveScaling.prediction) {
      const pred = forecasts.predictiveScaling.prediction;
      combined.predictedAgents = pred.predicted_active_agents || 0;
      combined.predictedQueueLength = pred.predicted_queue_length || 0;
      combined.predictedResponseTime = pred.predicted_avg_response_time || 0;
      combined.confidence += (pred.confidence || 0.7) * 0.35;
      combined.models.push('PredictiveScaling');
      totalWeight += 0.35;
    }

    // Normalize confidence if we have partial forecasts
    if (totalWeight > 0 && totalWeight < 1) {
      combined.confidence = combined.confidence / totalWeight;
    }

    return combined;
  }

  /**
   * Make ML-enhanced scaling decision
   */
  makeMLEnhancedScalingDecision({ workloadRatio, efficiency, combinedForecast, scalingInsights, currentAgents }) {
    const decision = {
      action: 'none',
      reason: 'Optimal capacity',
      confidence: combinedForecast.confidence
    };

    // Use predicted agent count from ML models
    const predictedOptimalAgents = combinedForecast.predictedAgents ||
                                   Math.ceil(combinedForecast.predictedQueueLength * 0.5) ||
                                   currentAgents;

    // Scale up if ML predicts we need more agents
    if (predictedOptimalAgents > currentAgents * 1.2 ||
        combinedForecast.predictedQueueLength > 10 ||
        combinedForecast.predictedResponseTime > 5000) {
      decision.action = 'scale_up';
      decision.reason = `ML forecast: need ${Math.round(predictedOptimalAgents)} agents (predicted queue: ${Math.round(combinedForecast.predictedQueueLength)}, response time: ${Math.round(combinedForecast.predictedResponseTime)}ms)`;
    }
    // Scale down if ML predicts we have too many agents
    else if (predictedOptimalAgents < currentAgents * 0.7 &&
             combinedForecast.predictedQueueLength < 3 &&
             currentAgents > this.minAgents) {
      decision.action = 'scale_down';
      decision.reason = `ML forecast: optimal ${Math.round(predictedOptimalAgents)} agents (low predicted load)`;
    }
    // Traditional heuristics as fallback
    else if (workloadRatio > 2 || efficiency < this.scalingConfig.scaleUpThreshold) {
      decision.action = 'scale_up';
      decision.reason = `Traditional heuristics: workload ratio ${workloadRatio.toFixed(2)}, efficiency ${(efficiency * 100).toFixed(1)}%`;
      decision.confidence = 0.6; // Lower confidence for non-ML decisions
    }
    else if (workloadRatio < this.scalingConfig.scaleDownThreshold && efficiency > 0.8) {
      decision.action = 'scale_down';
      decision.reason = `Traditional heuristics: low workload ${workloadRatio.toFixed(2)}, high efficiency ${(efficiency * 100).toFixed(1)}%`;
      decision.confidence = 0.6;
    }

    return decision;
  }

  /**
   * Get Random Forest predictions from agent selection model
   */
  async getRFPredictions() {
    try {
      // Create a dummy task request for RF prediction
      const taskRequest = {
        type: 'general',
        complexity: 'medium',
        priority: 5
      };

      // Get available agent IDs
      const availableAgents = Array.from(this.activeAgents.keys());

      if (availableAgents.length === 0) {
        return { successRate: 0.5, confidence: 0 };
      }

      // Get RF prediction for agent performance
      const prediction = await this.performanceForecaster.predictAgentPerformance(
        availableAgents[0],
        taskRequest
      );

      return {
        successRate: prediction.prediction.successRate,
        confidence: prediction.confidence,
        estimatedDuration: prediction.prediction.estimatedDuration
      };
    } catch (error) {
      console.error('❌ RF prediction error:', error.message);
      return { successRate: 0.5, confidence: 0 };
    }
  }

  /**
   * Handle scaling commands from predictive scaling system
   */
  async handleScalingCommand(command) {
    try {
      console.log(`📨 Received scaling command from ${command.source}: ${command.command} to ${command.targetAgentCount} agents`);

      if (command.command === 'scale_up') {
        const additionalAgents = Math.min(
          command.targetAgentCount - this.currentAgents,
          this.maxAgents - this.currentAgents
        );

        if (additionalAgents > 0) {
          for (let i = 0; i < additionalAgents; i++) {
            await this.spawnSpecializedAgent(
              { specialization: 'general-purpose', priority: 6, role: 'ml-scaled' },
              { task: command.reason, complexity: 0.5, priority: 5, taskType: 'ml-scaling' }
            );
          }
          console.log(`✅ ML-driven scale up: Added ${additionalAgents} agents (${command.reason})`);
        }
      } else if (command.command === 'scale_down') {
        const agentsToRemove = Math.min(
          this.currentAgents - command.targetAgentCount,
          this.currentAgents - this.minAgents
        );

        if (agentsToRemove > 0) {
          const sortedAgents = Array.from(this.activeAgents.values())
            .filter(agent => agent.status === 'idle' || agent.status === 'completed')
            .sort((a, b) => this.calculateAgentEffectiveness(a) - this.calculateAgentEffectiveness(b))
            .slice(0, agentsToRemove);

          for (const agent of sortedAgents) {
            this.activeAgents.delete(agent.id);
            this.currentAgents--;
          }
          console.log(`✅ ML-driven scale down: Removed ${agentsToRemove} agents (${command.reason})`);
        }
      }

      this.scalingConfig.lastScalingAction = Date.now();
      this.metrics.adaptations++;
    } catch (error) {
      console.error('❌ Error handling scaling command:', error.message);
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

    // Get ML metrics from forecaster
    const mlMetrics = await this.getMLMetrics();

    // Get predictive scaling status
    const predictiveScalingStatus = await this.predictiveScaling.getSystemStatus();

    // ML adoption and effectiveness metrics
    const mlAdoptionMetrics = {
      modelsActive: {
        lstm: !!this.lstmPredictor,
        randomForest: !!this.performanceForecaster,
        predictiveScaling: predictiveScalingStatus.modelTrained,
        neuralLearning: learningReport.summary.totalPatterns > 0
      },
      modelAccuracy: {
        lstm: this.lstmPredictor.modelAccuracy || 0,
        randomForest: this.performanceForecaster.modelAccuracy || 0,
        predictiveScaling: predictiveScalingStatus.accuracy || 0
      },
      mlDrivenScalingActions: this.countMLDrivenActions(),
      combinedForecastConfidence: this.getAverageForecastConfidence(),
      predictionHorizon: {
        lstm_short: '5-15 minutes',
        randomForest_medium: '1-4 hours',
        predictiveScaling_adaptive: 'adaptive horizon'
      }
    };

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
      mlMetrics,
      mlAdoption: mlAdoptionMetrics,
      specializationDistribution: this.getSpecializationDistribution(),
      performanceInsights: this.getPerformanceInsights()
    };

    // Save enhanced metrics
    const metricsPath = path.join(__dirname, '../../metrics/enhanced-swarm-metrics.json');
    await fs.writeFile(metricsPath, JSON.stringify(this.metrics, null, 2));
  }

  /**
   * Count ML-driven scaling actions
   */
  countMLDrivenActions() {
    // Count actions triggered by predictive scaling system
    return this.scalingHistory?.filter(action => action.source === 'predictive_scaling')?.length || 0;
  }

  /**
   * Get average forecast confidence across all ML models
   */
  getAverageForecastConfidence() {
    const confidences = [];

    if (this.lstmPredictor.modelAccuracy) confidences.push(this.lstmPredictor.modelAccuracy);
    if (this.performanceForecaster.modelAccuracy) confidences.push(this.performanceForecaster.modelAccuracy);

    return confidences.length > 0
      ? confidences.reduce((sum, c) => sum + c, 0) / confidences.length
      : 0;
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

  /**
   * Get LSTM load forecasts for all active agents
   */
  async getLSTMLoadForecasts() {
    const forecasts = [];

    for (const agentId of this.activeAgents.keys()) {
      try {
        const forecast = await this.lstmPredictor.predictAgentLoad(agentId);
        forecasts.push({ agentId, ...forecast });
      } catch (error) {
        console.error(`Error forecasting load for agent ${agentId}:`, error.message);
      }
    }

    return forecasts;
  }

  /**
   * Get ML metrics from forecaster
   */
  async getMLMetrics() {
    try {
      const predictions = [];

      for (const [agentId, agentData] of this.activeAgents.entries()) {
        if (agentData.status === 'active' || agentData.status === 'idle') {
          const forecast = await this.performanceForecaster.getAgentLoadForecast(agentId, 'short');
          predictions.push(forecast);
        }
      }

      return {
        totalPredictions: predictions.length,
        avgConfidence: predictions.reduce((sum, p) => sum + p.confidence, 0) / Math.max(predictions.length, 1),
        predictions: predictions.slice(0, 5) // Top 5 for metrics
      };
    } catch (error) {
      console.error('Error collecting ML metrics:', error.message);
      return { totalPredictions: 0, avgConfidence: 0, predictions: [] };
    }
  }
}

module.exports = EnhancedSmartAgentsSwarm;