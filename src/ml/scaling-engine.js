#!/usr/bin/env node

/**
 * Predictive Scaling Engine
 * Intelligent agent scaling using ML predictions and real-time optimization
 */

const Redis = require('ioredis');
const { performance } = require('perf_hooks');
const AgentSelectionModel = require('./agent-selection-model');
const PredictiveScalingSystem = require('./predictive-scaling');
const ABTestingFramework = require('./ab-testing-framework');
const FeatureStore = require('./feature-store');

class PredictiveScalingEngine {
  constructor(config = {}) {
    this.config = {
      redisNodes: config.redisNodes || [
        { host: 'redis-cluster-0.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-1.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-2.redis-cluster-service.ollamamax-redis', port: 6379 }
      ],
      redisPassword: config.redisPassword || 'ollama_redis_pass',
      decisionInterval: config.decisionInterval || 30000, // 30 seconds
      scalingCooldown: config.scalingCooldown || 300000, // 5 minutes
      agentPool: config.agentPool || {
        minAgents: 3,
        maxAgents: 50,
        scaleUpStep: 2,
        scaleDownStep: 1,
        emergencyThreshold: 100, // Queue length that triggers emergency scaling
        panicThreshold: 200
      },
      performanceThresholds: config.performanceThresholds || {
        maxResponseTime: 30000, // 30 seconds
        maxQueueLength: 50,
        maxCpuUtilization: 0.8,
        minSuccessRate: 0.85
      },
      scalingStrategies: config.scalingStrategies || [
        'predictive',      // ML-based prediction
        'reactive',        // Traditional threshold-based
        'hybrid',          // Combination of both
        'ab_test'         // A/B test different strategies
      ],
      ...config
    };

    this.redis = null;
    this.agentSelectionModel = null;
    this.predictiveScalingSystem = null;
    this.abTesting = null;
    this.featureStore = null;
    
    this.scalingHistory = [];
    this.lastScalingAction = 0;
    this.currentStrategy = 'hybrid';
    this.performanceMetrics = {
      accuracy: 0,
      responseTime: 0,
      costEfficiency: 0
    };

    this.initializeComponents();
    this.startScalingEngine();
  }

  async initializeComponents() {
    console.log('🚀 Initializing Predictive Scaling Engine...');

    try {
      // Initialize Redis
      this.redis = new Redis.Cluster(this.config.redisNodes, {
        redisOptions: {
          password: this.config.redisPassword,
          lazyConnect: true
        },
        enableOfflineQueue: false,
        retryDelayOnFailover: 100,
        maxRetriesPerRequest: 3
      });

      await this.redis.ping();
      console.log('✅ Connected to Redis cluster');

      // Initialize ML components
      this.featureStore = new FeatureStore({
        redisNodes: this.config.redisNodes,
        redisPassword: this.config.redisPassword
      });

      this.agentSelectionModel = new AgentSelectionModel({
        redisNodes: this.config.redisNodes,
        redisPassword: this.config.redisPassword
      });

      this.predictiveScalingSystem = new PredictiveScalingSystem({
        redisNodes: this.config.redisNodes,
        redisPassword: this.config.redisPassword,
        minAgents: this.config.agentPool.minAgents,
        maxAgents: this.config.agentPool.maxAgents
      });

      this.abTesting = new ABTestingFramework({
        redisNodes: this.config.redisNodes,
        redisPassword: this.config.redisPassword
      });

      console.log('✅ All components initialized');

      // Start A/B test for scaling strategies
      await this.initializeScalingStrategyTest();

    } catch (error) {
      console.error('❌ Initialization failed:', error.message);
      throw error;
    }
  }

  async initializeScalingStrategyTest() {
    try {
      // Create A/B test for scaling strategies
      const strategyTest = await this.abTesting.createTest({
        name: 'Scaling Strategy Comparison',
        description: 'Compare predictive vs hybrid scaling strategies',
        hypothesis: 'ML-based predictive scaling will improve efficiency by 20%',
        control: {
          name: 'Hybrid Scaling',
          strategy: 'hybrid',
          config: { predictiveWeight: 0.7, reactiveWeight: 0.3 }
        },
        treatment: {
          name: 'Pure Predictive Scaling',
          strategy: 'predictive',
          config: { confidenceThreshold: 0.8 }
        },
        targetMetrics: ['response_time', 'resource_utilization', 'success_rate'],
        duration: 7 * 24 * 60 * 60 * 1000, // 7 days
        minSampleSize: 200
      });

      console.log(`🧪 Scaling strategy A/B test initialized: ${strategyTest.id}`);
      this.strategyTestId = strategyTest.id;

    } catch (error) {
      console.error('❌ A/B test initialization failed:', error.message);
    }
  }

  startScalingEngine() {
    // Main scaling decision loop
    setInterval(async () => {
      await this.makeScalingDecision();
    }, this.config.decisionInterval);

    // Performance monitoring
    setInterval(async () => {
      await this.monitorPerformance();
    }, 60000); // Every minute

    // Strategy optimization
    setInterval(async () => {
      await this.optimizeStrategy();
    }, 600000); // Every 10 minutes

    console.log(`🕐 Scaling engine started (${this.config.decisionInterval / 1000}s decision interval)`);
  }

  async makeScalingDecision() {
    try {
      const startTime = performance.now();

      // Step 1: Collect current state
      const currentState = await this.collectSystemState();
      
      // Step 2: Get A/B test assignment for this decision
      const strategyAssignment = await this.abTesting.assignVariant(
        this.strategyTestId,
        `decision_${Date.now()}`,
        'system'
      );

      const strategy = strategyAssignment ? strategyAssignment.strategy : this.currentStrategy;

      // Step 3: Make scaling prediction based on strategy
      const scalingRecommendation = await this.generateScalingRecommendation(currentState, strategy);
      
      // Step 4: Execute scaling decision
      const scalingResult = await this.executeScalingDecision(scalingRecommendation);

      // Step 5: Record metrics for A/B testing
      if (strategyAssignment) {
        setTimeout(async () => {
          const resultMetrics = await this.collectResultMetrics(scalingResult);
          await this.abTesting.recordResult(this.strategyTestId, strategyAssignment.taskId, resultMetrics);
        }, 120000); // Record results after 2 minutes
      }

      const decisionTime = performance.now() - startTime;
      
      console.log(`⚖️  Scaling decision completed in ${decisionTime.toFixed(2)}ms`);
      console.log(`   📊 Strategy: ${strategy}`);
      console.log(`   🎯 Recommendation: ${scalingRecommendation.action} (${scalingRecommendation.targetAgents} agents)`);
      console.log(`   🔄 Result: ${scalingResult.executed ? 'executed' : 'skipped'}`);

    } catch (error) {
      console.error('❌ Scaling decision failed:', error.message);
    }
  }

  async collectSystemState() {
    const [
      activeAgents,
      queueLength,
      avgResponseTime,
      systemMetrics,
      temporalFeatures
    ] = await Promise.all([
      this.getActiveAgentCount(),
      this.getQueueLength(),
      this.getAverageResponseTime(),
      this.getSystemMetrics(),
      this.getTemporalFeatures()
    ]);

    return {
      timestamp: Date.now(),
      activeAgents,
      queueLength,
      avgResponseTime,
      ...systemMetrics,
      ...temporalFeatures
    };
  }

  async getActiveAgentCount() {
    try {
      const agentKeys = await this.redis.keys('agent:*:status');
      let activeCount = 0;
      
      for (const key of agentKeys) {
        const status = await this.redis.get(key);
        if (status === 'active' || status === 'busy') activeCount++;
      }
      
      return activeCount;
    } catch (error) {
      return this.config.agentPool.minAgents;
    }
  }

  async getQueueLength() {
    try {
      const queueLength = await this.redis.llen('swarm:task_queue');
      return queueLength || 0;
    } catch (error) {
      return 0;
    }
  }

  async getAverageResponseTime() {
    try {
      const responseTime = await this.redis.get('metrics:avg_response_time');
      return parseFloat(responseTime) || 0;
    } catch (error) {
      return 0;
    }
  }

  async getSystemMetrics() {
    try {
      const globalStats = await this.redis.hgetall('global:stats:realtime');
      return {
        cpuUtilization: parseFloat(globalStats.avgCpu) || 0,
        memoryUtilization: parseFloat(globalStats.avgMemory) || 0,
        successRate: parseFloat(globalStats.successRate) || 0.5,
        errorRate: 1 - (parseFloat(globalStats.successRate) || 0.5)
      };
    } catch (error) {
      return {
        cpuUtilization: 0.5,
        memoryUtilization: 0.5,
        successRate: 0.5,
        errorRate: 0.5
      };
    }
  }

  async getTemporalFeatures() {
    const now = new Date();
    const hour = now.getHours();
    const dayOfWeek = now.getDay();
    
    return {
      hourOfDay: hour / 24,
      dayOfWeek: dayOfWeek / 7,
      isBusinessHours: (dayOfWeek >= 1 && dayOfWeek <= 5 && hour >= 9 && hour <= 17),
      seasonalFactor: this.calculateSeasonalFactor(hour, dayOfWeek)
    };
  }

  calculateSeasonalFactor(hour, dayOfWeek) {
    const businessHoursFactor = (hour >= 9 && hour <= 17) ? 1.2 : 0.8;
    const weekdayFactor = (dayOfWeek >= 1 && dayOfWeek <= 5) ? 1.1 : 0.9;
    return businessHoursFactor * weekdayFactor;
  }

  async generateScalingRecommendation(currentState, strategy) {
    let recommendation;

    switch (strategy) {
      case 'predictive':
        recommendation = await this.getPredictiveRecommendation(currentState);
        break;
      case 'reactive':
        recommendation = this.getReactiveRecommendation(currentState);
        break;
      case 'hybrid':
        recommendation = await this.getHybridRecommendation(currentState);
        break;
      default:
        recommendation = await this.getHybridRecommendation(currentState);
    }

    // Add safety checks
    recommendation = this.applySafetyChecks(recommendation, currentState);
    
    return recommendation;
  }

  async getPredictiveRecommendation(currentState) {
    try {
      // Use LSTM model to predict future workload
      const prediction = await this.predictiveScalingSystem.makePrediction([currentState]);
      
      if (!prediction) {
        return this.getReactiveRecommendation(currentState);
      }

      const recommendedAgents = Math.round(prediction.predicted_active_agents);
      
      return {
        strategy: 'predictive',
        action: this.determineScalingAction(currentState.activeAgents, recommendedAgents),
        currentAgents: currentState.activeAgents,
        targetAgents: recommendedAgents,
        confidence: prediction.confidence,
        reasoning: `ML prediction: ${recommendedAgents} agents needed`,
        metrics: prediction
      };

    } catch (error) {
      console.error('❌ Predictive recommendation failed:', error.message);
      return this.getReactiveRecommendation(currentState);
    }
  }

  getReactiveRecommendation(currentState) {
    const { activeAgents, queueLength, avgResponseTime, cpuUtilization } = currentState;
    let targetAgents = activeAgents;
    let reasoning = '';

    // Emergency scaling
    if (queueLength >= this.config.agentPool.emergencyThreshold) {
      targetAgents = Math.min(activeAgents + 5, this.config.agentPool.maxAgents);
      reasoning = `Emergency scaling: queue length ${queueLength}`;
    }
    // Panic scaling
    else if (queueLength >= this.config.agentPool.panicThreshold) {
      targetAgents = this.config.agentPool.maxAgents;
      reasoning = `Panic scaling: queue length ${queueLength}`;
    }
    // Performance-based scaling
    else if (avgResponseTime > this.config.performanceThresholds.maxResponseTime || 
             queueLength > this.config.performanceThresholds.maxQueueLength ||
             cpuUtilization > this.config.performanceThresholds.maxCpuUtilization) {
      targetAgents = Math.min(activeAgents + this.config.agentPool.scaleUpStep, this.config.agentPool.maxAgents);
      reasoning = `Performance threshold breach: response=${avgResponseTime}ms, queue=${queueLength}, cpu=${cpuUtilization}`;
    }
    // Scale down if under-utilized
    else if (queueLength === 0 && avgResponseTime < 1000 && cpuUtilization < 0.3 && activeAgents > this.config.agentPool.minAgents) {
      targetAgents = Math.max(activeAgents - this.config.agentPool.scaleDownStep, this.config.agentPool.minAgents);
      reasoning = `Scale down: low utilization`;
    }
    else {
      reasoning = 'No scaling needed';
    }

    return {
      strategy: 'reactive',
      action: this.determineScalingAction(activeAgents, targetAgents),
      currentAgents: activeAgents,
      targetAgents,
      confidence: 0.8,
      reasoning
    };
  }

  async getHybridRecommendation(currentState) {
    try {
      // Get both predictive and reactive recommendations
      const [predictive, reactive] = await Promise.all([
        this.getPredictiveRecommendation(currentState),
        Promise.resolve(this.getReactiveRecommendation(currentState))
      ]);

      // Weighted combination (configurable weights)
      const predictiveWeight = 0.7;
      const reactiveWeight = 0.3;

      const hybridTarget = Math.round(
        predictiveWeight * predictive.targetAgents + 
        reactiveWeight * reactive.targetAgents
      );

      // Use higher confidence recommendation in case of large disagreement
      const disagreementThreshold = 3;
      const disagreement = Math.abs(predictive.targetAgents - reactive.targetAgents);
      
      let finalTarget = hybridTarget;
      let confidence = (predictive.confidence * predictiveWeight + reactive.confidence * reactiveWeight);
      
      if (disagreement > disagreementThreshold) {
        const higherConfidence = predictive.confidence > reactive.confidence ? predictive : reactive;
        finalTarget = higherConfidence.targetAgents;
        confidence = higherConfidence.confidence;
      }

      return {
        strategy: 'hybrid',
        action: this.determineScalingAction(currentState.activeAgents, finalTarget),
        currentAgents: currentState.activeAgents,
        targetAgents: finalTarget,
        confidence,
        reasoning: `Hybrid: predictive=${predictive.targetAgents}, reactive=${reactive.targetAgents}, final=${finalTarget}`,
        components: { predictive, reactive },
        disagreement
      };

    } catch (error) {
      console.error('❌ Hybrid recommendation failed:', error.message);
      return this.getReactiveRecommendation(currentState);
    }
  }

  determineScalingAction(currentAgents, targetAgents) {
    if (targetAgents > currentAgents) {
      return 'scale_up';
    } else if (targetAgents < currentAgents) {
      return 'scale_down';
    } else {
      return 'maintain';
    }
  }

  applySafetyChecks(recommendation, currentState) {
    // Cooldown check
    const timeSinceLastScaling = Date.now() - this.lastScalingAction;
    if (timeSinceLastScaling < this.config.scalingCooldown && recommendation.action !== 'maintain') {
      return {
        ...recommendation,
        action: 'maintain',
        targetAgents: currentState.activeAgents,
        reasoning: `${recommendation.reasoning} - BLOCKED by cooldown (${Math.round((this.config.scalingCooldown - timeSinceLastScaling) / 1000)}s remaining)`
      };
    }

    // Bounds check
    if (recommendation.targetAgents < this.config.agentPool.minAgents) {
      recommendation.targetAgents = this.config.agentPool.minAgents;
    }
    if (recommendation.targetAgents > this.config.agentPool.maxAgents) {
      recommendation.targetAgents = this.config.agentPool.maxAgents;
    }

    // Emergency override
    if (currentState.queueLength >= this.config.agentPool.panicThreshold) {
      recommendation.targetAgents = this.config.agentPool.maxAgents;
      recommendation.action = 'scale_up';
      recommendation.reasoning += ' - EMERGENCY OVERRIDE';
    }

    return recommendation;
  }

  async executeScalingDecision(recommendation) {
    const result = {
      executed: false,
      reason: '',
      actualChange: 0,
      executionTime: 0
    };

    if (recommendation.action === 'maintain') {
      result.reason = 'No scaling action needed';
      return result;
    }

    const startTime = performance.now();

    try {
      console.log(`🎯 Executing scaling: ${recommendation.action} to ${recommendation.targetAgents} agents`);
      console.log(`   💭 Reasoning: ${recommendation.reasoning}`);

      // Execute scaling command
      const scalingCommand = {
        command: recommendation.action,
        targetAgentCount: recommendation.targetAgents,
        currentAgentCount: recommendation.currentAgents,
        strategy: recommendation.strategy,
        confidence: recommendation.confidence,
        reason: recommendation.reasoning,
        timestamp: Date.now()
      };

      await this.redis.publish('swarm:scaling_command', JSON.stringify(scalingCommand));
      
      // Store scaling action
      const scalingAction = {
        ...scalingCommand,
        executionTime: performance.now() - startTime
      };

      await this.redis.lpush('ml:scaling_actions', JSON.stringify(scalingAction));
      await this.redis.ltrim('ml:scaling_actions', 0, 1000);

      this.scalingHistory.push(scalingAction);
      this.lastScalingAction = Date.now();

      result.executed = true;
      result.reason = 'Scaling command sent successfully';
      result.actualChange = recommendation.targetAgents - recommendation.currentAgents;
      result.executionTime = performance.now() - startTime;

      console.log(`✅ Scaling executed: ${recommendation.action} (${result.executionTime.toFixed(2)}ms)`);

    } catch (error) {
      console.error('❌ Scaling execution failed:', error.message);
      result.reason = `Execution failed: ${error.message}`;
    }

    return result;
  }

  async collectResultMetrics(scalingResult) {
    // Wait for system to stabilize after scaling
    await new Promise(resolve => setTimeout(resolve, 30000)); // 30 seconds

    const postScalingState = await this.collectSystemState();
    
    return {
      success: scalingResult.executed,
      executionTime: scalingResult.executionTime,
      resourceUtilization: (postScalingState.cpuUtilization + postScalingState.memoryUtilization) / 2,
      responseTime: postScalingState.avgResponseTime,
      queueLength: postScalingState.queueLength,
      userSatisfaction: postScalingState.successRate // Using success rate as proxy
    };
  }

  async monitorPerformance() {
    try {
      const currentState = await this.collectSystemState();
      
      // Calculate performance metrics
      const responseTimeScore = Math.max(0, 1 - (currentState.avgResponseTime / this.config.performanceThresholds.maxResponseTime));
      const queueScore = Math.max(0, 1 - (currentState.queueLength / this.config.performanceThresholds.maxQueueLength));
      const resourceScore = 1 - Math.max(currentState.cpuUtilization, currentState.memoryUtilization);
      
      this.performanceMetrics = {
        accuracy: (responseTimeScore + queueScore + currentState.successRate) / 3,
        responseTime: currentState.avgResponseTime,
        costEfficiency: resourceScore,
        timestamp: Date.now()
      };

      // Store performance metrics
      await this.redis.hset('ml:scaling_engine_performance', this.performanceMetrics);

    } catch (error) {
      console.error('❌ Performance monitoring failed:', error.message);
    }
  }

  async optimizeStrategy() {
    try {
      // Check A/B test results
      if (this.strategyTestId) {
        const testResults = await this.abTesting.getTestResults(this.strategyTestId);
        
        if (testResults && testResults.hasStatisticalSignificance) {
          console.log(`🎯 A/B test results: ${testResults.winningVariant} strategy is better`);
          
          if (testResults.winningVariant === 'treatment') {
            this.currentStrategy = 'predictive';
            console.log('📈 Switched to predictive strategy based on A/B test results');
          } else {
            this.currentStrategy = 'hybrid';
            console.log('📊 Keeping hybrid strategy based on A/B test results');
          }
        }
      }

    } catch (error) {
      console.error('❌ Strategy optimization failed:', error.message);
    }
  }

  async getEngineStatus() {
    const status = {
      currentStrategy: this.currentStrategy,
      lastScalingAction: this.lastScalingAction,
      scalingHistory: this.scalingHistory.slice(-10),
      performanceMetrics: this.performanceMetrics,
      config: this.config,
      components: {
        agentSelectionModel: !!this.agentSelectionModel,
        predictiveScalingSystem: !!this.predictiveScalingSystem,
        abTesting: !!this.abTesting,
        featureStore: !!this.featureStore
      }
    };

    try {
      const currentState = await this.collectSystemState();
      status.currentState = currentState;
      
      const recentActions = await this.redis.lrange('ml:scaling_actions', 0, 9);
      status.recentActions = recentActions.map(a => JSON.parse(a));
      
      if (this.strategyTestId) {
        const testStatus = await this.abTesting.getTestResults(this.strategyTestId);
        status.abTestStatus = testStatus;
      }
      
    } catch (error) {
      console.error('Error getting engine status:', error.message);
    }

    return status;
  }

  async shutdown() {
    console.log('🔄 Shutting down Predictive Scaling Engine...');
    
    const shutdownPromises = [
      this.agentSelectionModel?.shutdown(),
      this.predictiveScalingSystem?.shutdown(),
      this.abTesting?.shutdown(),
      this.featureStore?.shutdown()
    ].filter(Boolean);

    await Promise.all(shutdownPromises);
    
    if (this.redis) {
      await this.redis.disconnect();
    }
    
    console.log('✅ Predictive Scaling Engine shutdown complete');
  }
}

module.exports = PredictiveScalingEngine;

// Start engine if run directly
if (require.main === module) {
  const engine = new PredictiveScalingEngine();
  
  // Graceful shutdown
  process.on('SIGTERM', async () => {
    await engine.shutdown();
    process.exit(0);
  });
  
  process.on('SIGINT', async () => {
    await engine.shutdown();
    process.exit(0);
  });
}