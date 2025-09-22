/**
 * Self-Healing System Main Module
 * Coordinates error detection, recovery, and learning
 */

const ErrorDetector = require('./error-detector');
const RecoveryEngine = require('./recovery-engine');
const LearningSystem = require('./learning-system');
const EventEmitter = require('events');

class SelfHealingSystem extends EventEmitter {
  constructor(options = {}) {
    super();
    
    this.options = {
      enabled: true,
      autoRecover: true,
      learningEnabled: true,
      maxRetries: 3,
      logLevel: 'info',
      ...options
    };
    
    // Initialize components
    this.errorDetector = new ErrorDetector();
    this.recoveryEngine = new RecoveryEngine(this.errorDetector);
    this.learningSystem = new LearningSystem();
    
    // Statistics
    this.stats = {
      errorsDetected: 0,
      recoveriesAttempted: 0,
      recoveriesSuccessful: 0,
      recoveriesFailed: 0,
      startTime: new Date().toISOString()
    };
    
    this.setupEventHandlers();
  }
  
  /**
   * Setup event handlers between components
   */
  setupEventHandlers() {
    // Error detection events
    this.errorDetector.on('error-detected', (detection) => {
      this.stats.errorsDetected++;
      console.log(`🔍 Error detected: ${detection.type} - ${detection.severity}`);
      
      // Learn from error
      if (this.options.learningEnabled) {
        this.learningSystem.learnFromError(detection);
      }
      
      // Check for learned recovery strategy
      if (this.options.autoRecover) {
        const bestStrategy = this.learningSystem.getBestStrategy(detection);
        if (bestStrategy && bestStrategy.confidence > 0.7) {
          console.log(`📚 Using learned strategy: ${bestStrategy.strategy} (${Math.round(bestStrategy.successRate * 100)}% success rate)`);
        }
      }
      
      this.emit('error-detected', detection);
    });
    
    // Recovery events
    this.recoveryEngine.on('recovery-success', (detection) => {
      this.stats.recoveriesSuccessful++;
      console.log(`✅ Recovery successful for ${detection.type} error`);
      
      // Learn from successful recovery
      if (this.options.learningEnabled) {
        const strategy = this.errorDetector.getRecoveryStrategy(detection);
        this.learningSystem.learnFromRecovery(detection, strategy, true);
      }
      
      this.emit('recovery-success', detection);
    });
    
    this.recoveryEngine.on('recovery-failed', (detection) => {
      this.stats.recoveriesFailed++;
      console.log(`❌ Recovery failed for ${detection.type} error`);
      
      // Learn from failed recovery
      if (this.options.learningEnabled) {
        const strategy = this.errorDetector.getRecoveryStrategy(detection);
        this.learningSystem.learnFromRecovery(detection, strategy, false);
      }
      
      this.emit('recovery-failed', detection);
    });
    
    this.recoveryEngine.on('recovery-queued', (detection) => {
      console.log(`⏳ Recovery queued for ${detection.type} error`);
    });
    
    this.recoveryEngine.on('unrecoverable-error', (detection) => {
      console.log(`🚨 Unrecoverable error: ${detection.type}`);
      this.emit('unrecoverable-error', detection);
    });
    
    // Learning events
    this.learningSystem.on('knowledge-saved', (metadata) => {
      console.log(`💾 Knowledge base saved: ${metadata.entries} entries`);
    });
    
    this.learningSystem.on('learned-error', ({ key, pattern }) => {
      console.log(`📚 Learned from error: ${key} (${pattern.occurrences} occurrences)`);
    });
    
    this.learningSystem.on('learned-recovery', ({ recoveryKey, recovery, success }) => {
      console.log(`📚 Learned from recovery: ${recoveryKey} - ${success ? 'Success' : 'Failed'}`);
    });
  }
  
  /**
   * Process error message for detection and recovery
   */
  async processError(errorMessage, context = {}) {
    if (!this.options.enabled) {
      return null;
    }
    
    // Detect error type
    const detection = this.errorDetector.detectError(errorMessage, context);
    
    // Attempt recovery is handled by RecoveryEngine through events
    if (this.options.autoRecover && detection.recoverable) {
      this.stats.recoveriesAttempted++;
    }
    
    return detection;
  }
  
  /**
   * Predict potential errors based on context
   */
  predictErrors(context) {
    if (!this.options.learningEnabled) {
      return [];
    }
    
    return this.learningSystem.predictErrorLikelihood(context);
  }
  
  /**
   * Get system statistics
   */
  getStats() {
    const runtime = Date.now() - new Date(this.stats.startTime).getTime();
    const hours = runtime / (1000 * 60 * 60);
    
    return {
      ...this.stats,
      runtime: `${hours.toFixed(2)} hours`,
      recoveryRate: this.stats.recoveriesAttempted > 0 
        ? (this.stats.recoveriesSuccessful / this.stats.recoveriesAttempted * 100).toFixed(2) + '%'
        : '0%',
      errorPatterns: this.errorDetector.analyzePatterns(),
      learningInsights: this.learningSystem.getInsights()
    };
  }
  
  /**
   * Get recovery suggestions for an error
   */
  getRecoverySuggestions(errorMessage) {
    const detection = this.errorDetector.detectError(errorMessage);
    const strategy = this.errorDetector.getRecoveryStrategy(detection);
    const learned = this.learningSystem.getBestStrategy(detection);
    
    return {
      detection,
      defaultStrategy: strategy,
      learnedStrategy: learned,
      recommendations: this.generateRecommendations(detection, strategy, learned)
    };
  }
  
  /**
   * Generate recovery recommendations
   */
  generateRecommendations(detection, defaultStrategy, learnedStrategy) {
    const recommendations = [];
    
    if (learnedStrategy && learnedStrategy.confidence > 0.8) {
      recommendations.push({
        priority: 'high',
        action: learnedStrategy.strategy,
        reason: `High success rate (${Math.round(learnedStrategy.successRate * 100)}%) based on ${learnedStrategy.attempts} attempts`
      });
    }
    
    if (defaultStrategy) {
      recommendations.push({
        priority: learnedStrategy ? 'medium' : 'high',
        action: defaultStrategy.action,
        reason: 'Default recovery strategy for this error type'
      });
    }
    
    if (detection.severity === 'critical') {
      recommendations.push({
        priority: 'high',
        action: 'escalate',
        reason: 'Critical severity requires immediate attention'
      });
    }
    
    return recommendations;
  }
  
  /**
   * Export system data
   */
  async exportData() {
    const patterns = this.errorDetector.exportPatterns();
    const learning = await this.learningSystem.exportLearning();
    
    return {
      stats: this.getStats(),
      patterns,
      learningExport: learning
    };
  }
  
  /**
   * Import system data
   */
  async importData(data) {
    if (data.patterns) {
      this.errorDetector.importPatterns(data.patterns);
    }
    
    if (data.learningExport) {
      await this.learningSystem.importLearning(data.learningExport);
    }
    
    console.log('System data imported successfully');
  }
  
  /**
   * Enable/disable self-healing
   */
  setEnabled(enabled) {
    this.options.enabled = enabled;
    console.log(`Self-healing ${enabled ? 'enabled' : 'disabled'}`);
  }
  
  /**
   * Enable/disable auto-recovery
   */
  setAutoRecover(enabled) {
    this.options.autoRecover = enabled;
    console.log(`Auto-recovery ${enabled ? 'enabled' : 'disabled'}`);
  }
  
  /**
   * Enable/disable learning
   */
  setLearningEnabled(enabled) {
    this.options.learningEnabled = enabled;
    console.log(`Learning ${enabled ? 'enabled' : 'disabled'}`);
  }
  
  /**
   * Clear all history and learning data
   */
  async reset() {
    this.errorDetector.clearHistory();
    this.recoveryEngine.clearAttempts();
    this.stats = {
      errorsDetected: 0,
      recoveriesAttempted: 0,
      recoveriesSuccessful: 0,
      recoveriesFailed: 0,
      startTime: new Date().toISOString()
    };
    
    console.log('Self-healing system reset');
  }
}

// Singleton instance
let instance = null;

/**
 * Get or create self-healing system instance
 */
function getSelfHealingSystem(options) {
  if (!instance) {
    instance = new SelfHealingSystem(options);
  }
  return instance;
}

module.exports = {
  SelfHealingSystem,
  getSelfHealingSystem,
  ErrorDetector,
  RecoveryEngine,
  LearningSystem
};