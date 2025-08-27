/**
 * Neural Learning System for Smart Agents Swarm
 * Implements continuous learning, pattern recognition, and performance optimization
 */

const fs = require('fs').promises;
const path = require('path');

class NeuralLearningSystem {
  constructor(memoryPath) {
    this.memoryPath = memoryPath || path.join(__dirname, '../memory');
    this.learningData = new Map();
    this.patterns = new Map();
    this.performanceMetrics = new Map();
    this.adaptationRules = new Map();
    
    // Learning configuration
    this.config = {
      learningRate: 0.1,
      memoryRetention: 1000, // Max items per pattern
      adaptationThreshold: 0.8,
      patternConfidenceThreshold: 0.7,
      performanceWindowSize: 50
    };

    this.initializeLearningSystem();
  }

  async initializeLearningSystem() {
    await this.ensureDirectoryExists(this.memoryPath);
    await this.loadMemoryData();
    this.setupLearningRules();
  }

  async ensureDirectoryExists(dir) {
    try {
      await fs.mkdir(dir, { recursive: true });
    } catch (error) {
      if (error.code !== 'EEXIST') throw error;
    }
  }

  /**
   * Load existing learning data from persistent storage
   */
  async loadMemoryData() {
    try {
      const memoryFiles = [
        'neural-memory.json',
        'performance-patterns.json',
        'adaptation-rules.json'
      ];

      for (const file of memoryFiles) {
        const filePath = path.join(this.memoryPath, file);
        try {
          const data = await fs.readFile(filePath, 'utf8');
          const parsedData = JSON.parse(data);
          
          switch (file) {
            case 'neural-memory.json':
              this.learningData = new Map(Object.entries(parsedData));
              break;
            case 'performance-patterns.json':
              this.performanceMetrics = new Map(Object.entries(parsedData));
              break;
            case 'adaptation-rules.json':
              this.adaptationRules = new Map(Object.entries(parsedData));
              break;
          }
        } catch (fileError) {
          console.log(`Creating new ${file}...`);
        }
      }

      console.log(`🧠 Neural learning system loaded with ${this.learningData.size} patterns`);
    } catch (error) {
      console.log('🧠 Initializing new neural learning system...');
    }
  }

  /**
   * Save learning data to persistent storage
   */
  async saveMemoryData() {
    try {
      const memoryData = {
        'neural-memory.json': Object.fromEntries(this.learningData),
        'performance-patterns.json': Object.fromEntries(this.performanceMetrics),
        'adaptation-rules.json': Object.fromEntries(this.adaptationRules)
      };

      for (const [filename, data] of Object.entries(memoryData)) {
        const filePath = path.join(this.memoryPath, filename);
        await fs.writeFile(filePath, JSON.stringify(data, null, 2));
      }
    } catch (error) {
      console.error('❌ Failed to save neural memory:', error.message);
    }
  }

  /**
   * Setup initial learning rules and patterns
   */
  setupLearningRules() {
    // Success pattern recognition
    this.addLearningRule('success_patterns', {
      trigger: (data) => data.success === true,
      action: (data) => this.reinforceSuccessPattern(data),
      weight: 1.0
    });

    // Failure analysis and recovery
    this.addLearningRule('failure_analysis', {
      trigger: (data) => data.success === false,
      action: (data) => this.analyzeFailurePattern(data),
      weight: 0.8
    });

    // Performance optimization
    this.addLearningRule('performance_optimization', {
      trigger: (data) => data.executionTime !== undefined,
      action: (data) => this.optimizePerformancePattern(data),
      weight: 0.9
    });

    // Agent specialization optimization
    this.addLearningRule('specialization_optimization', {
      trigger: (data) => data.specialization && data.taskType,
      action: (data) => this.optimizeSpecializationMapping(data),
      weight: 0.85
    });
  }

  /**
   * Add a learning rule to the system
   */
  addLearningRule(name, rule) {
    this.adaptationRules.set(name, {
      ...rule,
      activationCount: 0,
      lastActivated: null,
      effectiveness: 0.5
    });
  }

  /**
   * Process learning data from agent execution
   */
  async processLearningData(agentResult) {
    const learningEntry = {
      timestamp: Date.now(),
      agentId: agentResult.agentId,
      specialization: agentResult.specialization,
      taskData: agentResult.taskData,
      success: agentResult.success,
      executionTime: agentResult.executionTime,
      output: agentResult.output,
      errors: agentResult.errors || [],
      patterns: agentResult.patterns || []
    };

    // Apply learning rules
    for (const [ruleName, rule] of this.adaptationRules.entries()) {
      if (rule.trigger(learningEntry)) {
        try {
          await rule.action(learningEntry);
          rule.activationCount++;
          rule.lastActivated = Date.now();
        } catch (error) {
          console.error(`❌ Learning rule ${ruleName} failed:`, error.message);
        }
      }
    }

    // Store in learning data
    const patternKey = this.generatePatternKey(learningEntry);
    const existingPattern = this.learningData.get(patternKey) || {
      samples: [],
      confidence: 0,
      successRate: 0,
      avgExecutionTime: 0,
      lastUpdated: Date.now()
    };

    existingPattern.samples.push(learningEntry);
    
    // Maintain memory size limits
    if (existingPattern.samples.length > this.config.memoryRetention) {
      existingPattern.samples = existingPattern.samples.slice(-this.config.memoryRetention);
    }

    // Update pattern statistics
    this.updatePatternStatistics(existingPattern);
    this.learningData.set(patternKey, existingPattern);

    // Periodic memory consolidation
    if (Math.random() < 0.1) { // 10% chance
      await this.consolidateMemory();
    }
  }

  /**
   * Generate a pattern key for similar learning contexts
   */
  generatePatternKey(entry) {
    const taskComplexity = entry.taskData?.complexity || 0;
    const complexityBand = Math.floor(taskComplexity * 10) / 10;
    
    return `${entry.specialization}-${complexityBand}-${entry.taskData?.taskType || 'general'}`;
  }

  /**
   * Update statistical measures for a pattern
   */
  updatePatternStatistics(pattern) {
    const samples = pattern.samples;
    const successfulSamples = samples.filter(s => s.success);
    
    pattern.successRate = successfulSamples.length / samples.length;
    pattern.avgExecutionTime = samples.reduce((sum, s) => sum + (s.executionTime || 0), 0) / samples.length;
    pattern.confidence = Math.min(0.95, Math.sqrt(samples.length / 100) * pattern.successRate);
    pattern.lastUpdated = Date.now();
  }

  /**
   * Reinforce successful patterns
   */
  async reinforceSuccessPattern(data) {
    const patternKey = `success-${data.specialization}-${data.taskData?.taskType || 'general'}`;
    const successPattern = this.patterns.get(patternKey) || {
      reinforcements: 0,
      strategies: new Map(),
      performance: []
    };

    successPattern.reinforcements++;
    successPattern.performance.push({
      timestamp: Date.now(),
      executionTime: data.executionTime,
      confidence: data.confidence || 0.8
    });

    // Extract successful strategies
    if (data.output && data.patterns) {
      data.patterns.forEach(pattern => {
        const existing = successPattern.strategies.get(pattern.pattern) || {
          frequency: 0,
          avgConfidence: 0,
          effectiveness: 0
        };
        
        existing.frequency++;
        existing.avgConfidence = (existing.avgConfidence + pattern.confidence) / 2;
        existing.effectiveness = Math.min(1.0, existing.effectiveness + this.config.learningRate);
        
        successPattern.strategies.set(pattern.pattern, existing);
      });
    }

    this.patterns.set(patternKey, successPattern);
  }

  /**
   * Analyze failure patterns for learning
   */
  async analyzeFailurePattern(data) {
    const failureKey = `failure-${data.specialization}-${data.errors[0]?.type || 'unknown'}`;
    const failurePattern = this.patterns.get(failureKey) || {
      occurrences: 0,
      contexts: [],
      recoveryStrategies: new Map()
    };

    failurePattern.occurrences++;
    failurePattern.contexts.push({
      timestamp: Date.now(),
      taskComplexity: data.taskData?.complexity || 0,
      errorDetails: data.errors,
      context: data.taskData
    });

    // Learn from failure context
    if (data.errors && data.errors.length > 0) {
      data.errors.forEach(error => {
        const recoveryKey = error.type || error.message?.substring(0, 50);
        const recovery = failurePattern.recoveryStrategies.get(recoveryKey) || {
          attempts: 0,
          successfulRecoveries: 0,
          strategies: []
        };
        
        recovery.attempts++;
        failurePattern.recoveryStrategies.set(recoveryKey, recovery);
      });
    }

    this.patterns.set(failureKey, failurePattern);
  }

  /**
   * Optimize performance patterns
   */
  async optimizePerformancePattern(data) {
    const perfKey = `performance-${data.specialization}`;
    const perfPattern = this.performanceMetrics.get(perfKey) || {
      executionTimes: [],
      optimizations: new Map(),
      trends: []
    };

    perfPattern.executionTimes.push({
      timestamp: Date.now(),
      time: data.executionTime,
      taskComplexity: data.taskData?.complexity || 0
    });

    // Maintain sliding window
    if (perfPattern.executionTimes.length > this.config.performanceWindowSize) {
      perfPattern.executionTimes = perfPattern.executionTimes.slice(-this.config.performanceWindowSize);
    }

    // Calculate performance trends
    if (perfPattern.executionTimes.length >= 10) {
      const recent = perfPattern.executionTimes.slice(-10);
      const older = perfPattern.executionTimes.slice(-20, -10);
      
      if (older.length > 0) {
        const recentAvg = recent.reduce((sum, p) => sum + p.time, 0) / recent.length;
        const olderAvg = older.reduce((sum, p) => sum + p.time, 0) / older.length;
        
        const trend = recentAvg < olderAvg ? 'improving' : 
                     recentAvg > olderAvg * 1.1 ? 'degrading' : 'stable';
        
        perfPattern.trends.push({
          timestamp: Date.now(),
          trend,
          improvement: (olderAvg - recentAvg) / olderAvg
        });
      }
    }

    this.performanceMetrics.set(perfKey, perfPattern);
  }

  /**
   * Optimize agent specialization mappings
   */
  async optimizeSpecializationMapping(data) {
    const mappingKey = `mapping-${data.taskData?.taskType || 'general'}`;
    const mapping = this.patterns.get(mappingKey) || {
      specializations: new Map(),
      effectiveness: new Map()
    };

    const specData = mapping.specializations.get(data.specialization) || {
      attempts: 0,
      successes: 0,
      avgExecutionTime: 0,
      complexityHandled: []
    };

    specData.attempts++;
    if (data.success) specData.successes++;
    specData.avgExecutionTime = (specData.avgExecutionTime + data.executionTime) / 2;
    specData.complexityHandled.push(data.taskData?.complexity || 0);

    const effectiveness = specData.successes / specData.attempts;
    mapping.specializations.set(data.specialization, specData);
    mapping.effectiveness.set(data.specialization, effectiveness);

    this.patterns.set(mappingKey, mapping);
  }

  /**
   * Consolidate memory by removing outdated or low-confidence patterns
   */
  async consolidateMemory() {
    const now = Date.now();
    const consolidationAge = 7 * 24 * 60 * 60 * 1000; // 7 days

    // Remove old, low-confidence patterns
    for (const [key, pattern] of this.learningData.entries()) {
      if (now - pattern.lastUpdated > consolidationAge && pattern.confidence < this.config.patternConfidenceThreshold) {
        this.learningData.delete(key);
      }
    }

    // Merge similar patterns
    await this.mergeSimilarPatterns();

    // Save consolidated memory
    await this.saveMemoryData();
    
    console.log(`🧠 Memory consolidated: ${this.learningData.size} patterns retained`);
  }

  /**
   * Merge similar patterns to reduce memory fragmentation
   */
  async mergeSimilarPatterns() {
    const patternsArray = Array.from(this.learningData.entries());
    const merged = new Set();

    for (let i = 0; i < patternsArray.length; i++) {
      if (merged.has(i)) continue;

      const [key1, pattern1] = patternsArray[i];
      
      for (let j = i + 1; j < patternsArray.length; j++) {
        if (merged.has(j)) continue;

        const [key2, pattern2] = patternsArray[j];
        
        if (this.calculatePatternSimilarity(pattern1, pattern2) > 0.85) {
          // Merge patterns
          const mergedPattern = this.mergePatterns(pattern1, pattern2);
          this.learningData.set(key1, mergedPattern);
          this.learningData.delete(key2);
          merged.add(j);
        }
      }
    }
  }

  /**
   * Calculate similarity between two patterns
   */
  calculatePatternSimilarity(pattern1, pattern2) {
    // Simple similarity based on success rate and execution time
    const successRateDiff = Math.abs(pattern1.successRate - pattern2.successRate);
    const timeDiff = Math.abs(pattern1.avgExecutionTime - pattern2.avgExecutionTime) / 
                     Math.max(pattern1.avgExecutionTime, pattern2.avgExecutionTime, 1);
    
    return 1 - (successRateDiff * 0.6 + timeDiff * 0.4);
  }

  /**
   * Merge two similar patterns
   */
  mergePatterns(pattern1, pattern2) {
    const totalSamples = pattern1.samples.length + pattern2.samples.length;
    
    return {
      samples: [...pattern1.samples, ...pattern2.samples].slice(-this.config.memoryRetention),
      confidence: (pattern1.confidence * pattern1.samples.length + pattern2.confidence * pattern2.samples.length) / totalSamples,
      successRate: (pattern1.successRate * pattern1.samples.length + pattern2.successRate * pattern2.samples.length) / totalSamples,
      avgExecutionTime: (pattern1.avgExecutionTime * pattern1.samples.length + pattern2.avgExecutionTime * pattern2.samples.length) / totalSamples,
      lastUpdated: Math.max(pattern1.lastUpdated, pattern2.lastUpdated)
    };
  }

  /**
   * Get learning recommendations based on accumulated patterns
   */
  getLearningRecommendations() {
    const recommendations = [];

    // Analyze performance trends
    for (const [key, metrics] of this.performanceMetrics.entries()) {
      if (metrics.trends.length > 0) {
        const latestTrend = metrics.trends[metrics.trends.length - 1];
        if (latestTrend.trend === 'degrading') {
          recommendations.push({
            type: 'performance',
            specialization: key.replace('performance-', ''),
            issue: 'Performance degrading',
            suggestion: 'Review recent changes and optimize execution patterns',
            priority: 8
          });
        }
      }
    }

    // Analyze failure patterns
    for (const [key, pattern] of this.patterns.entries()) {
      if (key.startsWith('failure-') && pattern.occurrences > 5) {
        recommendations.push({
          type: 'reliability',
          pattern: key,
          issue: `Recurring failures: ${pattern.occurrences} occurrences`,
          suggestion: 'Implement better error handling and recovery strategies',
          priority: 9
        });
      }
    }

    // Analyze specialization effectiveness
    for (const [key, mapping] of this.patterns.entries()) {
      if (key.startsWith('mapping-')) {
        for (const [spec, effectiveness] of mapping.effectiveness.entries()) {
          if (effectiveness < 0.6 && mapping.specializations.get(spec)?.attempts > 10) {
            recommendations.push({
              type: 'specialization',
              taskType: key.replace('mapping-', ''),
              specialization: spec,
              issue: `Low effectiveness: ${(effectiveness * 100).toFixed(1)}%`,
              suggestion: 'Consider alternative specialization or additional training',
              priority: 7
            });
          }
        }
      }
    }

    return recommendations.sort((a, b) => b.priority - a.priority);
  }

  /**
   * Generate learning insights report
   */
  generateLearningReport() {
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        totalPatterns: this.learningData.size,
        learningRulesActive: this.adaptationRules.size,
        performanceMetrics: this.performanceMetrics.size,
        memoryUtilization: this.calculateMemoryUtilization()
      },
      recommendations: this.getLearningRecommendations(),
      topPerformingPatterns: this.getTopPerformingPatterns(),
      learningTrends: this.analyzeLearningTrends()
    };

    return report;
  }

  /**
   * Calculate memory utilization statistics
   */
  calculateMemoryUtilization() {
    let totalSamples = 0;
    let highConfidencePatterns = 0;

    for (const pattern of this.learningData.values()) {
      totalSamples += pattern.samples.length;
      if (pattern.confidence > this.config.patternConfidenceThreshold) {
        highConfidencePatterns++;
      }
    }

    return {
      totalSamples,
      averageSamplesPerPattern: totalSamples / (this.learningData.size || 1),
      highConfidencePatterns,
      confidenceRatio: highConfidencePatterns / (this.learningData.size || 1)
    };
  }

  /**
   * Get top performing patterns
   */
  getTopPerformingPatterns() {
    return Array.from(this.learningData.entries())
      .filter(([, pattern]) => pattern.confidence > 0.8 && pattern.successRate > 0.9)
      .sort((a, b) => (b[1].confidence * b[1].successRate) - (a[1].confidence * a[1].successRate))
      .slice(0, 10)
      .map(([key, pattern]) => ({
        pattern: key,
        confidence: pattern.confidence,
        successRate: pattern.successRate,
        samples: pattern.samples.length
      }));
  }

  /**
   * Analyze learning trends over time
   */
  analyzeLearningTrends() {
    const now = Date.now();
    const timeWindows = [
      { name: 'last24h', ms: 24 * 60 * 60 * 1000 },
      { name: 'last7d', ms: 7 * 24 * 60 * 60 * 1000 },
      { name: 'last30d', ms: 30 * 24 * 60 * 60 * 1000 }
    ];

    const trends = {};

    timeWindows.forEach(window => {
      const cutoff = now - window.ms;
      let newPatterns = 0;
      let updatedPatterns = 0;
      let totalSamples = 0;

      for (const pattern of this.learningData.values()) {
        if (pattern.lastUpdated > cutoff) {
          updatedPatterns++;
          const recentSamples = pattern.samples.filter(s => s.timestamp > cutoff);
          totalSamples += recentSamples.length;
          
          if (pattern.samples.length === recentSamples.length) {
            newPatterns++;
          }
        }
      }

      trends[window.name] = {
        newPatterns,
        updatedPatterns,
        totalSamples,
        learningRate: totalSamples / (window.ms / (60 * 60 * 1000)) // samples per hour
      };
    });

    return trends;
  }

  /**
   * Train the system with focused learning on specific patterns
   */
  async trainSystem(focusAreas = []) {
    console.log('🧠 Starting focused training session...');

    // Reinforcement learning on high-performing patterns
    for (const [key, pattern] of this.learningData.entries()) {
      if (pattern.confidence > 0.8 && pattern.successRate > 0.9) {
        // Increase pattern weight
        pattern.confidence = Math.min(0.95, pattern.confidence + this.config.learningRate * 0.1);
      }
    }

    // Focus training on specific areas if provided
    if (focusAreas.length > 0) {
      for (const area of focusAreas) {
        await this.focusedTraining(area);
      }
    }

    // Consolidate and save
    await this.consolidateMemory();
    
    console.log('✅ Training session completed');
    return this.generateLearningReport();
  }

  /**
   * Focused training on specific domain
   */
  async focusedTraining(focusArea) {
    console.log(`🎯 Focused training on: ${focusArea}`);
    
    // Find relevant patterns
    const relevantPatterns = new Map();
    for (const [key, pattern] of this.learningData.entries()) {
      if (key.includes(focusArea)) {
        relevantPatterns.set(key, pattern);
      }
    }

    // Apply focused learning algorithms
    for (const [key, pattern] of relevantPatterns.entries()) {
      // Increase learning rate for this domain
      const focusedLearningRate = this.config.learningRate * 1.5;
      
      // Analyze recent performance
      const recentSamples = pattern.samples.slice(-20);
      const recentSuccessRate = recentSamples.filter(s => s.success).length / recentSamples.length;
      
      if (recentSuccessRate > pattern.successRate) {
        // Pattern is improving, reinforce
        pattern.confidence = Math.min(0.95, pattern.confidence + focusedLearningRate);
      } else if (recentSuccessRate < pattern.successRate * 0.8) {
        // Pattern is degrading, needs attention
        console.log(`⚠️  Pattern ${key} needs attention: success rate dropped`);
      }
    }
  }
}

module.exports = NeuralLearningSystem;