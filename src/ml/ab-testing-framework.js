#!/usr/bin/env node

/**
 * A/B Testing Framework for Agent Selection
 * Systematic comparison of agent selection strategies with statistical significance
 */

const Redis = require('ioredis');
const { performance } = require('perf_hooks');

class ABTestingFramework {
  constructor(config = {}) {
    this.config = {
      testDuration: config.testDuration || 7 * 24 * 60 * 60 * 1000, // 7 days
      minSampleSize: config.minSampleSize || 100,
      significanceLevel: config.significanceLevel || 0.05,
      powerLevel: config.powerLevel || 0.8,
      trafficSplit: config.trafficSplit || 0.5, // 50-50 split by default
      metrics: config.metrics || [
        'success_rate',
        'execution_time',
        'error_rate',
        'user_satisfaction',
        'resource_utilization'
      ],
      redisNodes: config.redisNodes || [
        { host: 'redis-cluster-0.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-1.redis-cluster-service.ollamamax-redis', port: 6379 },
        { host: 'redis-cluster-2.redis-cluster-service.ollamamax-redis', port: 6379 }
      ],
      redisPassword: config.redisPassword || 'ollama_redis_pass',
      ...config
    };

    this.redis = null;
    this.activeTests = new Map();
    this.testHistory = [];

    this.initializeRedis();
  }

  async initializeRedis() {
    try {
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
      console.log('✅ A/B Testing Framework connected to Redis cluster');
    } catch (error) {
      console.error('❌ Redis cluster connection failed:', error.message);
      throw error;
    }
  }

  async createTest(testConfig) {
    const testId = `ab_test_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const test = {
      id: testId,
      name: testConfig.name,
      description: testConfig.description,
      hypothesis: testConfig.hypothesis,
      startTime: Date.now(),
      endTime: Date.now() + (testConfig.duration || this.config.testDuration),
      status: 'active',
      trafficSplit: testConfig.trafficSplit || this.config.trafficSplit,
      variants: {
        control: {
          name: testConfig.control.name,
          strategy: testConfig.control.strategy,
          config: testConfig.control.config,
          samples: 0,
          metrics: {}
        },
        treatment: {
          name: testConfig.treatment.name,
          strategy: testConfig.treatment.strategy,
          config: testConfig.treatment.config,
          samples: 0,
          metrics: {}
        }
      },
      targetMetrics: testConfig.targetMetrics || this.config.metrics,
      minSampleSize: testConfig.minSampleSize || this.config.minSampleSize,
      significanceLevel: testConfig.significanceLevel || this.config.significanceLevel,
      results: null
    };

    // Store test configuration using SET with JSON string
    await this.redis.set(`ab_test:${testId}:config`, JSON.stringify(test));
    await this.redis.sadd('ab_tests:active', testId);
    
    this.activeTests.set(testId, test);
    
    console.log(`🧪 A/B Test created: ${testId}`);
    console.log(`   📋 Name: ${test.name}`);
    console.log(`   🎯 Hypothesis: ${test.hypothesis}`);
    console.log(`   ⏱️  Duration: ${Math.round(test.endTime - test.startTime) / (24 * 60 * 60 * 1000)} days`);
    console.log(`   🔀 Traffic split: ${(test.trafficSplit * 100).toFixed(1)}% / ${((1 - test.trafficSplit) * 100).toFixed(1)}%`);
    
    return test;
  }

  async assignVariant(testId, taskId, userId = null) {
    const test = this.activeTests.get(testId) || await this.loadTest(testId);
    if (!test || test.status !== 'active') {
      return null;
    }

    // Check if test has ended
    if (Date.now() > test.endTime) {
      await this.endTest(testId);
      return null;
    }

    // Deterministic assignment based on task/user ID
    const assignmentKey = userId || taskId;
    const hash = this.hashString(assignmentKey) % 100;
    const isControl = hash < (test.trafficSplit * 100);
    
    const variant = isControl ? 'control' : 'treatment';
    const assignment = {
      testId,
      taskId,
      userId,
      variant,
      assignmentTime: Date.now(),
      strategy: test.variants[variant].strategy,
      config: test.variants[variant].config
    };

    // Store assignment
    await this.redis.hset(`ab_test:${testId}:assignments`, taskId, JSON.stringify(assignment));
    
    // Increment sample count
    test.variants[variant].samples++;
    await this.redis.hincrby(`ab_test:${testId}:samples`, variant, 1);
    
    console.log(`🎲 Task ${taskId} assigned to ${variant} group (${test.variants[variant].strategy})`);
    
    return assignment;
  }

  hashString(str) {
    let hash = 0;
    if (str.length === 0) return hash;
    
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    
    return Math.abs(hash);
  }

  async recordResult(testId, taskId, metrics) {
    const test = this.activeTests.get(testId) || await this.loadTest(testId);
    if (!test) {
      console.error(`❌ Test ${testId} not found`);
      return;
    }

    // Get task assignment
    const assignmentData = await this.redis.hget(`ab_test:${testId}:assignments`, taskId);
    if (!assignmentData) {
      console.error(`❌ No assignment found for task ${taskId} in test ${testId}`);
      return;
    }

    const assignment = JSON.parse(assignmentData);
    const variant = assignment.variant;

    // Record result
    const result = {
      testId,
      taskId,
      variant,
      metrics,
      timestamp: Date.now(),
      executionTime: metrics.executionTime || 0,
      success: metrics.success || false,
      errorType: metrics.errorType || null
    };

    await this.redis.lpush(`ab_test:${testId}:results:${variant}`, JSON.stringify(result));
    
    // Update running metrics
    await this.updateRunningMetrics(testId, variant, metrics);
    
    // Check if we can analyze results
    if (test.variants.control.samples >= test.minSampleSize && 
        test.variants.treatment.samples >= test.minSampleSize) {
      await this.checkForSignificance(testId);
    }

    console.log(`📊 Result recorded for test ${testId}, task ${taskId} (${variant})`);
  }

  async updateRunningMetrics(testId, variant, metrics) {
    const key = `ab_test:${testId}:running_metrics:${variant}`;
    
    // Update counters and sums
    await this.redis.hincrby(key, 'total_tasks', 1);
    
    if (metrics.success) {
      await this.redis.hincrby(key, 'successful_tasks', 1);
    } else {
      await this.redis.hincrby(key, 'failed_tasks', 1);
    }
    
    if (metrics.executionTime) {
      await this.redis.hincrbyfloat(key, 'total_execution_time', metrics.executionTime);
    }
    
    if (metrics.resourceUtilization) {
      await this.redis.hincrbyfloat(key, 'total_resource_util', metrics.resourceUtilization);
    }
    
    // Calculate running averages
    const runningMetrics = await this.redis.hgetall(key);
    const totalTasks = parseInt(runningMetrics.total_tasks) || 1;
    
    const successRate = (parseInt(runningMetrics.successful_tasks) || 0) / totalTasks;
    const avgExecutionTime = (parseFloat(runningMetrics.total_execution_time) || 0) / totalTasks;
    const avgResourceUtil = (parseFloat(runningMetrics.total_resource_util) || 0) / totalTasks;
    
    await this.redis.hset(key, {
      success_rate: successRate,
      avg_execution_time: avgExecutionTime,
      avg_resource_utilization: avgResourceUtil,
      error_rate: 1 - successRate
    });
  }

  async checkForSignificance(testId) {
    try {
      const results = await this.analyzeTest(testId);
      
      if (results.hasStatisticalSignificance) {
        console.log(`🎯 Test ${testId} achieved statistical significance!`);
        console.log(`   📈 Winning variant: ${results.winningVariant}`);
        console.log(`   📊 Confidence: ${(results.confidence * 100).toFixed(2)}%`);
        
        // Auto-end test if configured
        if (this.config.autoEndOnSignificance) {
          await this.endTest(testId);
        }
      }
    } catch (error) {
      console.error(`❌ Error checking significance for test ${testId}:`, error.message);
    }
  }

  async analyzeTest(testId) {
    const test = this.activeTests.get(testId) || await this.loadTest(testId);
    if (!test) {
      throw new Error(`Test ${testId} not found`);
    }

    console.log(`📊 Analyzing test: ${test.name}`);

    // Get results for both variants
    const [controlResults, treatmentResults] = await Promise.all([
      this.getVariantResults(testId, 'control'),
      this.getVariantResults(testId, 'treatment')
    ]);

    if (controlResults.length === 0 || treatmentResults.length === 0) {
      return {
        testId,
        status: 'insufficient_data',
        hasStatisticalSignificance: false,
        message: 'Insufficient data for analysis'
      };
    }

    // Analyze each target metric
    const metricAnalysis = {};
    
    for (const metric of test.targetMetrics) {
      const analysis = await this.analyzeMetric(controlResults, treatmentResults, metric);
      metricAnalysis[metric] = analysis;
    }

    // Overall test analysis
    const primaryMetric = test.targetMetrics[0]; // Use first metric as primary
    const primaryAnalysis = metricAnalysis[primaryMetric];
    
    const analysis = {
      testId,
      testName: test.name,
      status: test.status,
      startTime: test.startTime,
      duration: Date.now() - test.startTime,
      variants: {
        control: {
          samples: controlResults.length,
          metrics: this.calculateVariantMetrics(controlResults)
        },
        treatment: {
          samples: treatmentResults.length,
          metrics: this.calculateVariantMetrics(treatmentResults)
        }
      },
      metricAnalysis,
      primaryMetric,
      hasStatisticalSignificance: primaryAnalysis.isSignificant,
      pValue: primaryAnalysis.pValue,
      confidence: 1 - primaryAnalysis.pValue,
      winningVariant: primaryAnalysis.isSignificant ? primaryAnalysis.winner : null,
      effect: primaryAnalysis.effect,
      recommendation: this.generateRecommendation(metricAnalysis, primaryAnalysis),
      timestamp: Date.now()
    };

    // Store analysis results using SET with JSON string
    await this.redis.set(`ab_test:${testId}:analysis`, JSON.stringify(analysis));
    
    return analysis;
  }

  async getVariantResults(testId, variant) {
    const resultsData = await this.redis.lrange(`ab_test:${testId}:results:${variant}`, 0, -1);
    return resultsData.map(data => JSON.parse(data));
  }

  calculateVariantMetrics(results) {
    const totalTasks = results.length;
    const successfulTasks = results.filter(r => r.metrics.success).length;
    const totalExecutionTime = results.reduce((sum, r) => sum + (r.metrics.executionTime || 0), 0);
    const totalResourceUtil = results.reduce((sum, r) => sum + (r.metrics.resourceUtilization || 0), 0);

    return {
      total_tasks: totalTasks,
      success_rate: successfulTasks / totalTasks,
      error_rate: 1 - (successfulTasks / totalTasks),
      avg_execution_time: totalExecutionTime / totalTasks,
      avg_resource_utilization: totalResourceUtil / totalTasks,
      p95_execution_time: this.calculatePercentile(results.map(r => r.metrics.executionTime || 0), 95),
      p99_execution_time: this.calculatePercentile(results.map(r => r.metrics.executionTime || 0), 99)
    };
  }

  calculatePercentile(values, percentile) {
    const sorted = values.sort((a, b) => a - b);
    const index = Math.ceil(sorted.length * (percentile / 100)) - 1;
    return sorted[index] || 0;
  }

  async analyzeMetric(controlResults, treatmentResults, metricName) {
    const controlValues = this.extractMetricValues(controlResults, metricName);
    const treatmentValues = this.extractMetricValues(treatmentResults, metricName);
    
    const controlMean = this.calculateMean(controlValues);
    const treatmentMean = this.calculateMean(treatmentValues);
    const controlStd = this.calculateStandardDeviation(controlValues, controlMean);
    const treatmentStd = this.calculateStandardDeviation(treatmentValues, treatmentMean);
    
    // Perform t-test
    const tTest = this.performTTest(controlValues, treatmentValues);
    
    // Calculate effect size (Cohen's d)
    const pooledStd = Math.sqrt(((controlValues.length - 1) * controlStd * controlStd + 
                                (treatmentValues.length - 1) * treatmentStd * treatmentStd) / 
                               (controlValues.length + treatmentValues.length - 2));
    const cohensD = (treatmentMean - controlMean) / pooledStd;
    
    // Determine winner
    const improvement = ((treatmentMean - controlMean) / controlMean) * 100;
    const winner = treatmentMean > controlMean ? 'treatment' : 'control';
    
    return {
      metric: metricName,
      control: {
        mean: controlMean,
        std: controlStd,
        samples: controlValues.length
      },
      treatment: {
        mean: treatmentMean,
        std: treatmentStd,
        samples: treatmentValues.length
      },
      pValue: tTest.pValue,
      tStatistic: tTest.tStatistic,
      isSignificant: tTest.pValue < this.config.significanceLevel,
      cohensD: cohensD,
      effect: this.interpretCohensD(cohensD),
      improvement: improvement,
      winner: winner,
      confidence: 1 - tTest.pValue
    };
  }

  extractMetricValues(results, metricName) {
    return results.map(result => {
      switch (metricName) {
        case 'success_rate':
          return result.metrics.success ? 1 : 0;
        case 'execution_time':
          return result.metrics.executionTime || 0;
        case 'error_rate':
          return result.metrics.success ? 0 : 1;
        case 'resource_utilization':
          return result.metrics.resourceUtilization || 0;
        case 'user_satisfaction':
          return result.metrics.userSatisfaction || 0;
        default:
          return result.metrics[metricName] || 0;
      }
    });
  }

  calculateMean(values) {
    return values.reduce((sum, val) => sum + val, 0) / values.length;
  }

  calculateStandardDeviation(values, mean) {
    const variance = values.reduce((sum, val) => sum + Math.pow(val - mean, 2), 0) / values.length;
    return Math.sqrt(variance);
  }

  performTTest(sample1, sample2) {
    const n1 = sample1.length;
    const n2 = sample2.length;
    const mean1 = this.calculateMean(sample1);
    const mean2 = this.calculateMean(sample2);
    const std1 = this.calculateStandardDeviation(sample1, mean1);
    const std2 = this.calculateStandardDeviation(sample2, mean2);
    
    // Welch's t-test (unequal variances)
    const se1 = (std1 * std1) / n1;
    const se2 = (std2 * std2) / n2;
    const se = Math.sqrt(se1 + se2);
    
    const tStatistic = (mean1 - mean2) / se;
    
    // Degrees of freedom (Welch-Satterthwaite equation)
    const df = Math.pow(se1 + se2, 2) / (Math.pow(se1, 2) / (n1 - 1) + Math.pow(se2, 2) / (n2 - 1));
    
    // Approximate p-value calculation (simplified)
    const pValue = this.approximatePValue(Math.abs(tStatistic), df);
    
    return {
      tStatistic,
      pValue,
      degreesOfFreedom: df
    };
  }

  approximatePValue(tStat, df) {
    // Simplified p-value approximation
    // For a more accurate implementation, use a proper t-distribution library
    if (tStat < 1.96) return 0.05;
    if (tStat < 2.58) return 0.01;
    if (tStat < 3.29) return 0.001;
    return 0.0001;
  }

  interpretCohensD(d) {
    const absD = Math.abs(d);
    if (absD < 0.2) return 'negligible';
    if (absD < 0.5) return 'small';
    if (absD < 0.8) return 'medium';
    return 'large';
  }

  generateRecommendation(metricAnalysis, primaryAnalysis) {
    if (!primaryAnalysis.isSignificant) {
      return {
        decision: 'continue_test',
        reason: 'No statistically significant difference detected',
        action: 'Continue test or collect more data'
      };
    }

    const winner = primaryAnalysis.winner;
    const improvement = Math.abs(primaryAnalysis.improvement);
    
    if (improvement < 5) {
      return {
        decision: 'minimal_impact',
        reason: `${winner} variant shows statistical significance but minimal practical impact (${improvement.toFixed(2)}%)`,
        action: 'Consider cost-benefit analysis before implementation'
      };
    }

    return {
      decision: 'implement_winner',
      reason: `${winner} variant shows significant improvement (${improvement.toFixed(2)}%) with ${primaryAnalysis.effect} effect size`,
      action: `Roll out ${winner} variant to all traffic`
    };
  }

  async endTest(testId) {
    const test = this.activeTests.get(testId) || await this.loadTest(testId);
    if (!test) {
      throw new Error(`Test ${testId} not found`);
    }

    console.log(`🏁 Ending test: ${test.name}`);
    
    // Final analysis
    const finalResults = await this.analyzeTest(testId);
    
    // Update test status
    test.status = 'completed';
    test.endTime = Date.now();
    test.results = finalResults;

    await this.redis.set(`ab_test:${testId}:config`, JSON.stringify(test));
    await this.redis.srem('ab_tests:active', testId);
    await this.redis.sadd('ab_tests:completed', testId);
    
    this.activeTests.delete(testId);
    this.testHistory.push(test);
    
    console.log(`✅ Test ${testId} completed`);
    console.log(`   📊 Result: ${finalResults.recommendation.decision}`);
    console.log(`   💡 Recommendation: ${finalResults.recommendation.action}`);
    
    return finalResults;
  }

  async loadTest(testId) {
    try {
      const testData = await this.redis.get(`ab_test:${testId}:config`);
      if (!testData) {
        return null;
      }

      const test = JSON.parse(testData);
      return test;
    } catch (error) {
      console.error(`❌ Error loading test ${testId}:`, error.message);
      return null;
    }
  }

  async getActiveTests() {
    const testIds = await this.redis.smembers('ab_tests:active');
    const tests = [];
    
    for (const testId of testIds) {
      const test = await this.loadTest(testId);
      if (test) {
        // Add current sample sizes
        const samples = await this.redis.hgetall(`ab_test:${testId}:samples`);
        test.variants.control.samples = parseInt(samples.control) || 0;
        test.variants.treatment.samples = parseInt(samples.treatment) || 0;
        
        tests.push(test);
      }
    }
    
    return tests;
  }

  async getTestResults(testId) {
    const test = await this.loadTest(testId);
    if (!test) {
      throw new Error(`Test ${testId} not found`);
    }
    
    if (test.status === 'active') {
      return await this.analyzeTest(testId);
    } else {
      const analysisData = await this.redis.get(`ab_test:${testId}:analysis`);
      return analysisData ? JSON.parse(analysisData) : null;
    }
  }

  async shutdown() {
    console.log('🔄 Shutting down A/B Testing Framework...');
    
    if (this.redis) {
      await this.redis.disconnect();
    }
    
    console.log('✅ A/B Testing Framework shutdown complete');
  }
}

module.exports = ABTestingFramework;

// Example usage if run directly
if (require.main === module) {
  const framework = new ABTestingFramework();
  
  // Example: Create a test comparing ML model vs. rule-based agent selection
  const exampleTest = {
    name: 'ML vs Rule-Based Agent Selection',
    description: 'Compare Random Forest model against rule-based specialization matching',
    hypothesis: 'ML-based agent selection will improve success rate by 15%',
    control: {
      name: 'Rule-Based Selection',
      strategy: 'specialization_matching',
      config: { threshold: 0.7 }
    },
    treatment: {
      name: 'ML Model Selection', 
      strategy: 'random_forest',
      config: { confidence_threshold: 0.8 }
    },
    targetMetrics: ['success_rate', 'execution_time', 'resource_utilization'],
    duration: 7 * 24 * 60 * 60 * 1000, // 7 days
    minSampleSize: 100
  };
  
  framework.createTest(exampleTest).then(() => {
    console.log('✅ Example test created successfully');
  }).catch(console.error);
  
  // Graceful shutdown
  process.on('SIGTERM', async () => {
    await framework.shutdown();
    process.exit(0);
  });
  
  process.on('SIGINT', async () => {
    await framework.shutdown();
    process.exit(0);
  });
}