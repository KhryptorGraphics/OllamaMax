/**
 * Self-Healing Learning System
 * Learns from errors and recoveries to improve future handling
 */

const fs = require('fs').promises;
const path = require('path');
const EventEmitter = require('events');

class LearningSystem extends EventEmitter {
  constructor() {
    super();
    this.knowledgeBase = new Map();
    this.errorPatterns = new Map();
    this.recoverySuccess = new Map();
    this.confidenceThreshold = 0.7;
    this.learningRate = 0.1;
    
    this.loadKnowledgeBase();
  }
  
  /**
   * Load existing knowledge base
   */
  async loadKnowledgeBase() {
    try {
      const kbPath = path.join('.swarm', 'knowledge-base.json');
      const data = await fs.readFile(kbPath, 'utf8');
      const kb = JSON.parse(data);
      
      // Restore knowledge base
      kb.patterns?.forEach(([key, value]) => {
        this.errorPatterns.set(key, value);
      });
      
      kb.recoveries?.forEach(([key, value]) => {
        this.recoverySuccess.set(key, value);
      });
      
      kb.knowledge?.forEach(([key, value]) => {
        this.knowledgeBase.set(key, value);
      });
      
      console.log(`Loaded ${this.knowledgeBase.size} knowledge entries`);
    } catch (error) {
      // Knowledge base doesn't exist yet
      console.log('Starting with empty knowledge base');
    }
  }
  
  /**
   * Save knowledge base to disk
   */
  async saveKnowledgeBase() {
    try {
      const kbPath = path.join('.swarm', 'knowledge-base.json');
      
      const kb = {
        patterns: Array.from(this.errorPatterns.entries()),
        recoveries: Array.from(this.recoverySuccess.entries()),
        knowledge: Array.from(this.knowledgeBase.entries()),
        metadata: {
          updated: new Date().toISOString(),
          entries: this.knowledgeBase.size,
          patterns: this.errorPatterns.size,
          recoveries: this.recoverySuccess.size
        }
      };
      
      await fs.mkdir('.swarm', { recursive: true });
      await fs.writeFile(kbPath, JSON.stringify(kb, null, 2));
      
      this.emit('knowledge-saved', kb.metadata);
    } catch (error) {
      console.error(`Failed to save knowledge base: ${error.message}`);
    }
  }
  
  /**
   * Learn from error occurrence
   */
  learnFromError(detection) {
    const key = this.generateErrorKey(detection);
    const pattern = this.errorPatterns.get(key) || {
      occurrences: 0,
      firstSeen: new Date().toISOString(),
      lastSeen: null,
      contexts: [],
      recoveryAttempts: 0,
      successfulRecoveries: 0,
      confidence: 0
    };
    
    // Update pattern
    pattern.occurrences++;
    pattern.lastSeen = new Date().toISOString();
    pattern.contexts.push(detection.context);
    
    // Keep only last 10 contexts
    if (pattern.contexts.length > 10) {
      pattern.contexts = pattern.contexts.slice(-10);
    }
    
    // Update confidence based on occurrences
    pattern.confidence = Math.min(1, pattern.occurrences * 0.1);
    
    this.errorPatterns.set(key, pattern);
    
    // Emit learning event
    this.emit('learned-error', { key, pattern });
    
    // Auto-save periodically
    if (pattern.occurrences % 10 === 0) {
      this.saveKnowledgeBase();
    }
    
    return pattern;
  }
  
  /**
   * Learn from recovery attempt
   */
  learnFromRecovery(detection, strategy, success) {
    const errorKey = this.generateErrorKey(detection);
    const recoveryKey = this.generateRecoveryKey(detection, strategy);
    
    // Update error pattern
    const pattern = this.errorPatterns.get(errorKey);
    if (pattern) {
      pattern.recoveryAttempts++;
      if (success) {
        pattern.successfulRecoveries++;
      }
      this.errorPatterns.set(errorKey, pattern);
    }
    
    // Update recovery success rate
    const recovery = this.recoverySuccess.get(recoveryKey) || {
      attempts: 0,
      successes: 0,
      failures: 0,
      successRate: 0,
      lastAttempt: null,
      avgRecoveryTime: 0,
      confidence: 0
    };
    
    recovery.attempts++;
    if (success) {
      recovery.successes++;
    } else {
      recovery.failures++;
    }
    
    recovery.successRate = recovery.successes / recovery.attempts;
    recovery.lastAttempt = new Date().toISOString();
    recovery.confidence = Math.min(1, recovery.attempts * 0.05);
    
    this.recoverySuccess.set(recoveryKey, recovery);
    
    // Store learned knowledge
    this.addKnowledge(errorKey, {
      type: 'recovery',
      strategy: strategy.action,
      success,
      confidence: recovery.confidence
    });
    
    // Emit learning event
    this.emit('learned-recovery', { recoveryKey, recovery, success });
    
    return recovery;
  }
  
  /**
   * Get best recovery strategy based on learning
   */
  getBestStrategy(detection) {
    const errorKey = this.generateErrorKey(detection);
    const pattern = this.errorPatterns.get(errorKey);
    
    if (!pattern || pattern.confidence < this.confidenceThreshold) {
      return null; // Not enough data
    }
    
    // Find all recovery strategies for this error
    const strategies = [];
    for (const [key, recovery] of this.recoverySuccess) {
      if (key.startsWith(errorKey)) {
        strategies.push({
          key,
          recovery,
          score: recovery.successRate * recovery.confidence
        });
      }
    }
    
    // Sort by score and return best
    strategies.sort((a, b) => b.score - a.score);
    
    if (strategies.length > 0 && strategies[0].score > this.confidenceThreshold) {
      return {
        strategy: strategies[0].key.split('::')[1],
        confidence: strategies[0].recovery.confidence,
        successRate: strategies[0].recovery.successRate,
        attempts: strategies[0].recovery.attempts
      };
    }
    
    return null;
  }
  
  /**
   * Predict error likelihood
   */
  predictErrorLikelihood(context) {
    const predictions = [];
    
    for (const [key, pattern] of this.errorPatterns) {
      // Simple context matching
      let contextMatch = 0;
      pattern.contexts.forEach(ctx => {
        if (JSON.stringify(ctx).includes(JSON.stringify(context))) {
          contextMatch++;
        }
      });
      
      const contextScore = contextMatch / pattern.contexts.length;
      const recencyScore = this.calculateRecencyScore(pattern.lastSeen);
      const frequencyScore = Math.min(1, pattern.occurrences / 100);
      
      const likelihood = (contextScore * 0.5 + recencyScore * 0.3 + frequencyScore * 0.2) * pattern.confidence;
      
      if (likelihood > 0.3) {
        predictions.push({
          error: key,
          likelihood,
          pattern,
          preventable: pattern.successfulRecoveries > 0
        });
      }
    }
    
    predictions.sort((a, b) => b.likelihood - a.likelihood);
    return predictions.slice(0, 5); // Top 5 predictions
  }
  
  /**
   * Add knowledge entry
   */
  addKnowledge(key, value) {
    const knowledge = this.knowledgeBase.get(key) || [];
    knowledge.push({
      ...value,
      timestamp: new Date().toISOString()
    });
    
    // Keep only last 20 entries per key
    if (knowledge.length > 20) {
      knowledge.shift();
    }
    
    this.knowledgeBase.set(key, knowledge);
  }
  
  /**
   * Get insights from learning
   */
  getInsights() {
    const insights = {
      totalErrors: this.errorPatterns.size,
      totalRecoveries: this.recoverySuccess.size,
      knowledgeEntries: this.knowledgeBase.size,
      topErrors: [],
      bestRecoveries: [],
      predictions: [],
      recommendations: []
    };
    
    // Top errors by frequency
    const errors = Array.from(this.errorPatterns.entries())
      .map(([key, pattern]) => ({ key, ...pattern }))
      .sort((a, b) => b.occurrences - a.occurrences);
    
    insights.topErrors = errors.slice(0, 5).map(e => ({
      type: e.key.split('::')[0],
      occurrences: e.occurrences,
      recoveryRate: e.successfulRecoveries / e.recoveryAttempts || 0
    }));
    
    // Best recovery strategies
    const recoveries = Array.from(this.recoverySuccess.entries())
      .map(([key, recovery]) => ({ key, ...recovery }))
      .sort((a, b) => b.successRate - a.successRate);
    
    insights.bestRecoveries = recoveries.slice(0, 5).map(r => ({
      strategy: r.key.split('::')[1],
      successRate: r.successRate,
      attempts: r.attempts
    }));
    
    // Generate recommendations
    if (errors.length > 0) {
      errors.slice(0, 3).forEach(error => {
        if (error.occurrences > 5 && error.successfulRecoveries / error.recoveryAttempts < 0.5) {
          insights.recommendations.push({
            type: 'improve-recovery',
            error: error.key.split('::')[0],
            reason: 'Low recovery success rate',
            suggestion: 'Review and improve recovery strategy'
          });
        }
        
        if (error.occurrences > 10 && error.confidence > 0.8) {
          insights.recommendations.push({
            type: 'prevent-error',
            error: error.key.split('::')[0],
            reason: 'Frequent occurrence',
            suggestion: 'Implement preventive measures'
          });
        }
      });
    }
    
    return insights;
  }
  
  /**
   * Generate error key for pattern matching
   */
  generateErrorKey(detection) {
    return `${detection.type}::${detection.patternName || 'unknown'}`;
  }
  
  /**
   * Generate recovery key
   */
  generateRecoveryKey(detection, strategy) {
    return `${this.generateErrorKey(detection)}::${strategy.action}`;
  }
  
  /**
   * Calculate recency score
   */
  calculateRecencyScore(lastSeen) {
    if (!lastSeen) return 0;
    
    const now = Date.now();
    const then = new Date(lastSeen).getTime();
    const hoursSince = (now - then) / (1000 * 60 * 60);
    
    // Decay over 24 hours
    return Math.max(0, 1 - (hoursSince / 24));
  }
  
  /**
   * Export learning data
   */
  async exportLearning() {
    const exportData = {
      timestamp: new Date().toISOString(),
      patterns: Array.from(this.errorPatterns.entries()),
      recoveries: Array.from(this.recoverySuccess.entries()),
      knowledge: Array.from(this.knowledgeBase.entries()),
      insights: this.getInsights()
    };
    
    const exportPath = path.join('.swarm', `learning-export-${Date.now()}.json`);
    await fs.writeFile(exportPath, JSON.stringify(exportData, null, 2));
    
    return exportPath;
  }
  
  /**
   * Import learning data
   */
  async importLearning(filePath) {
    try {
      const data = await fs.readFile(filePath, 'utf8');
      const importData = JSON.parse(data);
      
      // Merge patterns
      importData.patterns?.forEach(([key, value]) => {
        const existing = this.errorPatterns.get(key);
        if (existing) {
          // Merge occurrences and contexts
          existing.occurrences += value.occurrences;
          existing.contexts = [...existing.contexts, ...value.contexts].slice(-10);
          existing.successfulRecoveries += value.successfulRecoveries;
          existing.recoveryAttempts += value.recoveryAttempts;
          this.errorPatterns.set(key, existing);
        } else {
          this.errorPatterns.set(key, value);
        }
      });
      
      // Merge recoveries
      importData.recoveries?.forEach(([key, value]) => {
        const existing = this.recoverySuccess.get(key);
        if (existing) {
          existing.attempts += value.attempts;
          existing.successes += value.successes;
          existing.failures += value.failures;
          existing.successRate = existing.successes / existing.attempts;
          this.recoverySuccess.set(key, existing);
        } else {
          this.recoverySuccess.set(key, value);
        }
      });
      
      // Merge knowledge
      importData.knowledge?.forEach(([key, value]) => {
        const existing = this.knowledgeBase.get(key) || [];
        this.knowledgeBase.set(key, [...existing, ...value].slice(-20));
      });
      
      await this.saveKnowledgeBase();
      
      return {
        patternsImported: importData.patterns?.length || 0,
        recoveriesImported: importData.recoveries?.length || 0,
        knowledgeImported: importData.knowledge?.length || 0
      };
    } catch (error) {
      throw new Error(`Failed to import learning data: ${error.message}`);
    }
  }
}

module.exports = LearningSystem;