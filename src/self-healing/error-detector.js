/**
 * Self-Healing Error Detection and Monitoring System
 * Automatically detects and categorizes errors for recovery
 */

const EventEmitter = require('events');
const winston = require('winston');

class ErrorDetector extends EventEmitter {
  constructor() {
    super();
    this.patterns = new Map();
    this.recoveryStrategies = new Map();
    this.errorHistory = [];
    this.maxHistorySize = 1000;
    
    // Initialize error patterns
    this.initializePatterns();
    
    // Setup logger
    this.logger = winston.createLogger({
      level: 'info',
      format: winston.format.json(),
      transports: [
        new winston.transports.File({ filename: '.swarm/error-detection.log' })
      ]
    });
  }
  
  initializePatterns() {
    // Missing dependency patterns
    this.patterns.set('missing-module', {
      regex: /Cannot find module ['"](.+?)['"]/,
      type: 'dependency',
      severity: 'medium',
      recoverable: true
    });
    
    // Syntax error patterns
    this.patterns.set('syntax-error', {
      regex: /SyntaxError: (.+)/,
      type: 'syntax',
      severity: 'high',
      recoverable: true
    });
    
    // Test failure patterns
    this.patterns.set('test-failure', {
      regex: /(\d+) test\(s\) failed/,
      type: 'test',
      severity: 'medium',
      recoverable: true
    });
    
    // TypeScript error patterns
    this.patterns.set('type-error', {
      regex: /TS\d+: (.+)/,
      type: 'typescript',
      severity: 'medium',
      recoverable: true
    });
    
    // Port already in use
    this.patterns.set('port-in-use', {
      regex: /EADDRINUSE.*:(\d+)/,
      type: 'runtime',
      severity: 'low',
      recoverable: true
    });
    
    // Permission denied
    this.patterns.set('permission-denied', {
      regex: /EACCES|Permission denied/,
      type: 'permission',
      severity: 'high',
      recoverable: true
    });
    
    // Memory errors
    this.patterns.set('out-of-memory', {
      regex: /JavaScript heap out of memory/,
      type: 'memory',
      severity: 'critical',
      recoverable: true
    });
    
    // Network errors
    this.patterns.set('network-error', {
      regex: /ECONNREFUSED|ETIMEDOUT|ENETUNREACH/,
      type: 'network',
      severity: 'medium',
      recoverable: true
    });
  }
  
  /**
   * Detect error type and context from error message
   */
  detectError(errorMessage, context = {}) {
    const detection = {
      timestamp: new Date().toISOString(),
      message: errorMessage,
      context,
      type: 'unknown',
      severity: 'low',
      recoverable: false,
      matches: []
    };
    
    // Check against all patterns
    for (const [name, pattern] of this.patterns) {
      const match = errorMessage.match(pattern.regex);
      if (match) {
        detection.type = pattern.type;
        detection.severity = pattern.severity;
        detection.recoverable = pattern.recoverable;
        detection.patternName = name;
        detection.matches = match.slice(1);
        break;
      }
    }
    
    // Store in history
    this.addToHistory(detection);
    
    // Log detection
    this.logger.info('Error detected', detection);
    
    // Emit detection event
    this.emit('error-detected', detection);
    
    return detection;
  }
  
  /**
   * Analyze error patterns over time
   */
  analyzePatterns() {
    const analysis = {
      totalErrors: this.errorHistory.length,
      byType: {},
      bySeverity: {},
      recoveryRate: 0,
      commonErrors: [],
      trends: []
    };
    
    // Count by type and severity
    this.errorHistory.forEach(error => {
      // By type
      if (!analysis.byType[error.type]) {
        analysis.byType[error.type] = 0;
      }
      analysis.byType[error.type]++;
      
      // By severity
      if (!analysis.bySeverity[error.severity]) {
        analysis.bySeverity[error.severity] = 0;
      }
      analysis.bySeverity[error.severity]++;
    });
    
    // Calculate recovery rate
    const recoverableErrors = this.errorHistory.filter(e => e.recoverable).length;
    analysis.recoveryRate = this.errorHistory.length > 0 
      ? (recoverableErrors / this.errorHistory.length) * 100 
      : 0;
    
    // Find common errors
    const errorCounts = {};
    this.errorHistory.forEach(error => {
      const key = error.patternName || error.type;
      errorCounts[key] = (errorCounts[key] || 0) + 1;
    });
    
    analysis.commonErrors = Object.entries(errorCounts)
      .sort((a, b) => b[1] - a[1])
      .slice(0, 5)
      .map(([name, count]) => ({ name, count }));
    
    // Detect trends (errors increasing over time)
    if (this.errorHistory.length > 10) {
      const recentErrors = this.errorHistory.slice(-10);
      const olderErrors = this.errorHistory.slice(-20, -10);
      
      const recentTypes = {};
      const olderTypes = {};
      
      recentErrors.forEach(e => {
        recentTypes[e.type] = (recentTypes[e.type] || 0) + 1;
      });
      
      olderErrors.forEach(e => {
        olderTypes[e.type] = (olderTypes[e.type] || 0) + 1;
      });
      
      for (const type in recentTypes) {
        const recent = recentTypes[type];
        const older = olderTypes[type] || 0;
        if (recent > older * 1.5) {
          analysis.trends.push({
            type,
            trend: 'increasing',
            change: ((recent - older) / (older || 1)) * 100
          });
        }
      }
    }
    
    return analysis;
  }
  
  /**
   * Check if error matches known patterns
   */
  isKnownError(errorMessage) {
    for (const pattern of this.patterns.values()) {
      if (errorMessage.match(pattern.regex)) {
        return true;
      }
    }
    return false;
  }
  
  /**
   * Get recovery strategy for detected error
   */
  getRecoveryStrategy(detection) {
    const strategies = {
      'missing-module': {
        action: 'install-dependency',
        command: `npm install ${detection.matches[0]}`,
        description: 'Install missing dependency'
      },
      'syntax-error': {
        action: 'analyze-and-fix',
        description: 'Analyze syntax error and suggest fix'
      },
      'test-failure': {
        action: 'debug-test',
        description: 'Debug and fix failing test'
      },
      'type-error': {
        action: 'fix-types',
        description: 'Fix TypeScript type errors'
      },
      'port-in-use': {
        action: 'change-port',
        description: 'Use alternative port or kill process'
      },
      'permission-denied': {
        action: 'fix-permissions',
        description: 'Fix file permissions'
      },
      'out-of-memory': {
        action: 'increase-memory',
        command: 'export NODE_OPTIONS="--max-old-space-size=4096"',
        description: 'Increase Node.js memory limit'
      },
      'network-error': {
        action: 'retry-with-backoff',
        description: 'Retry operation with exponential backoff'
      }
    };
    
    return strategies[detection.patternName] || null;
  }
  
  /**
   * Add error to history with size limit
   */
  addToHistory(detection) {
    this.errorHistory.push(detection);
    
    // Maintain max history size
    if (this.errorHistory.length > this.maxHistorySize) {
      this.errorHistory.shift();
    }
  }
  
  /**
   * Clear error history
   */
  clearHistory() {
    this.errorHistory = [];
    this.logger.info('Error history cleared');
  }
  
  /**
   * Export error patterns for learning
   */
  exportPatterns() {
    return {
      patterns: Array.from(this.patterns.entries()),
      history: this.errorHistory,
      analysis: this.analyzePatterns()
    };
  }
  
  /**
   * Import learned patterns
   */
  importPatterns(data) {
    if (data.patterns) {
      data.patterns.forEach(([name, pattern]) => {
        this.patterns.set(name, pattern);
      });
    }
    this.logger.info(`Imported ${data.patterns?.length || 0} patterns`);
  }
}

module.exports = ErrorDetector;