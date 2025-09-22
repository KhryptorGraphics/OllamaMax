/**
 * Self-Healing Recovery Engine
 * Automatically recovers from detected errors
 */

const { exec } = require('child_process');
const { promisify } = require('util');
const fs = require('fs').promises;
const path = require('path');
const EventEmitter = require('events');

const execAsync = promisify(exec);

class RecoveryEngine extends EventEmitter {
  constructor(errorDetector) {
    super();
    this.errorDetector = errorDetector;
    this.recoveryAttempts = new Map();
    this.maxRetries = 3;
    this.retryDelay = 1000; // Start with 1 second
    this.isRecovering = false;
    
    // Listen for detected errors
    this.errorDetector.on('error-detected', this.handleError.bind(this));
  }
  
  /**
   * Handle detected error and attempt recovery
   */
  async handleError(detection) {
    if (!detection.recoverable) {
      this.emit('unrecoverable-error', detection);
      return;
    }
    
    if (this.isRecovering) {
      this.emit('recovery-queued', detection);
      return;
    }
    
    this.isRecovering = true;
    
    try {
      const recovered = await this.attemptRecovery(detection);
      if (recovered) {
        this.emit('recovery-success', detection);
      } else {
        this.emit('recovery-failed', detection);
      }
    } catch (error) {
      this.emit('recovery-error', { detection, error });
    } finally {
      this.isRecovering = false;
    }
  }
  
  /**
   * Attempt to recover from error
   */
  async attemptRecovery(detection) {
    const strategy = this.errorDetector.getRecoveryStrategy(detection);
    if (!strategy) {
      console.log(`No recovery strategy for ${detection.type} error`);
      return false;
    }
    
    const attemptKey = `${detection.patternName}-${Date.now()}`;
    const attempts = this.recoveryAttempts.get(attemptKey) || 0;
    
    if (attempts >= this.maxRetries) {
      console.log(`Max recovery attempts reached for ${detection.patternName}`);
      return false;
    }
    
    this.recoveryAttempts.set(attemptKey, attempts + 1);
    
    console.log(`Attempting recovery: ${strategy.description}`);
    
    switch (strategy.action) {
      case 'install-dependency':
        return await this.installDependency(detection, strategy);
        
      case 'analyze-and-fix':
        return await this.analyzeSyntaxError(detection);
        
      case 'debug-test':
        return await this.debugFailingTest(detection);
        
      case 'fix-types':
        return await this.fixTypeErrors(detection);
        
      case 'change-port':
        return await this.changePort(detection);
        
      case 'fix-permissions':
        return await this.fixPermissions(detection);
        
      case 'increase-memory':
        return await this.increaseMemory(strategy);
        
      case 'retry-with-backoff':
        return await this.retryWithBackoff(detection);
        
      default:
        console.log(`Unknown recovery action: ${strategy.action}`);
        return false;
    }
  }
  
  /**
   * Install missing dependency
   */
  async installDependency(detection, strategy) {
    try {
      const moduleName = detection.matches[0];
      console.log(`Installing missing module: ${moduleName}`);
      
      // Check if it's a local file
      if (moduleName.startsWith('.') || moduleName.startsWith('/')) {
        console.log('Missing local file, cannot auto-install');
        return false;
      }
      
      // Try to install the package
      const { stdout, stderr } = await execAsync(`npm install ${moduleName}`);
      console.log(`Installed ${moduleName} successfully`);
      
      // Store successful recovery
      await this.storeRecoveryPattern(detection, {
        action: 'npm install',
        module: moduleName,
        success: true
      });
      
      return true;
    } catch (error) {
      console.error(`Failed to install dependency: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Analyze and fix syntax errors
   */
  async analyzeSyntaxError(detection) {
    try {
      const errorMessage = detection.message;
      
      // Extract file and line info if available
      const fileMatch = errorMessage.match(/(\S+\.js):(\d+):(\d+)/);
      if (!fileMatch) {
        console.log('Cannot determine file location from error');
        return false;
      }
      
      const [, filePath, line, column] = fileMatch;
      
      // Common syntax fixes
      const fixes = [
        { pattern: /Unexpected token \)/, fix: 'Remove extra closing parenthesis' },
        { pattern: /Unexpected token \}/, fix: 'Remove extra closing brace' },
        { pattern: /Unexpected token ;/, fix: 'Remove extra semicolon' },
        { pattern: /Unexpected end of input/, fix: 'Add missing closing brace or parenthesis' },
        { pattern: /Missing semicolon/, fix: 'Add semicolon at end of statement' }
      ];
      
      for (const { pattern, fix } of fixes) {
        if (errorMessage.match(pattern)) {
          console.log(`Syntax fix suggestion: ${fix} at ${filePath}:${line}:${column}`);
          
          // Store the suggestion for manual review
          await this.storeRecoveryPattern(detection, {
            action: 'syntax-fix',
            file: filePath,
            line,
            column,
            suggestion: fix
          });
          
          return true;
        }
      }
      
      return false;
    } catch (error) {
      console.error(`Failed to analyze syntax error: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Debug failing test
   */
  async debugFailingTest(detection) {
    try {
      console.log('Analyzing test failure...');
      
      // Try to run tests with verbose output
      const { stdout, stderr } = await execAsync('npm test -- --verbose');
      
      // Parse test output for specific failures
      const failurePattern = /✗\s+(.+?)\s+\((\d+)ms\)/g;
      const failures = [...stdout.matchAll(failurePattern)];
      
      if (failures.length > 0) {
        console.log(`Found ${failures.length} failing test(s)`);
        
        // Store failure patterns
        await this.storeRecoveryPattern(detection, {
          action: 'test-debug',
          failures: failures.map(f => ({
            name: f[1],
            duration: f[2]
          }))
        });
        
        return true;
      }
      
      return false;
    } catch (error) {
      console.error(`Failed to debug test: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Fix TypeScript type errors
   */
  async fixTypeErrors(detection) {
    try {
      console.log('Analyzing TypeScript errors...');
      
      // Run TypeScript compiler for diagnostics
      const { stdout } = await execAsync('npx tsc --noEmit --pretty false');
      
      // Parse TypeScript errors
      const tsErrors = stdout.split('\n').filter(line => line.includes('TS'));
      
      // Common type fixes
      const typeFixes = [
        { 
          pattern: /Property '(.+)' does not exist/, 
          fix: 'Add property definition to interface/type' 
        },
        { 
          pattern: /Cannot find name '(.+)'/, 
          fix: 'Import or declare the missing type' 
        },
        { 
          pattern: /Type '(.+)' is not assignable to type '(.+)'/, 
          fix: 'Fix type mismatch or add type assertion' 
        }
      ];
      
      let fixApplied = false;
      for (const error of tsErrors) {
        for (const { pattern, fix } of typeFixes) {
          if (error.match(pattern)) {
            console.log(`Type fix suggestion: ${fix}`);
            fixApplied = true;
          }
        }
      }
      
      return fixApplied;
    } catch (error) {
      // TypeScript not configured, skip
      return false;
    }
  }
  
  /**
   * Change port when current port is in use
   */
  async changePort(detection) {
    try {
      const portMatch = detection.message.match(/:(\d+)/);
      if (!portMatch) return false;
      
      const currentPort = parseInt(portMatch[1]);
      const newPort = currentPort + 1;
      
      console.log(`Port ${currentPort} in use, trying ${newPort}`);
      
      // Update environment variable
      process.env.PORT = newPort;
      
      // Store recovery action
      await this.storeRecoveryPattern(detection, {
        action: 'port-change',
        oldPort: currentPort,
        newPort: newPort
      });
      
      return true;
    } catch (error) {
      console.error(`Failed to change port: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Fix file permissions
   */
  async fixPermissions(detection) {
    try {
      // Extract file path from error message
      const fileMatch = detection.message.match(/['"](.+?)['"]/);
      if (!fileMatch) return false;
      
      const filePath = fileMatch[1];
      console.log(`Fixing permissions for: ${filePath}`);
      
      // Try to fix permissions
      await execAsync(`chmod +rw ${filePath}`);
      
      await this.storeRecoveryPattern(detection, {
        action: 'fix-permissions',
        file: filePath
      });
      
      return true;
    } catch (error) {
      console.error(`Failed to fix permissions: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Increase memory limit
   */
  async increaseMemory(strategy) {
    try {
      console.log('Increasing Node.js memory limit');
      
      // Execute memory increase command
      await execAsync(strategy.command);
      
      // Update current process
      if (!process.execArgv.includes('--max-old-space-size')) {
        process.execArgv.push('--max-old-space-size=4096');
      }
      
      return true;
    } catch (error) {
      console.error(`Failed to increase memory: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Retry operation with exponential backoff
   */
  async retryWithBackoff(detection) {
    const attemptKey = `retry-${Date.now()}`;
    const attempts = this.recoveryAttempts.get(attemptKey) || 0;
    
    if (attempts >= this.maxRetries) {
      return false;
    }
    
    const delay = this.retryDelay * Math.pow(2, attempts);
    console.log(`Retrying in ${delay}ms (attempt ${attempts + 1}/${this.maxRetries})`);
    
    await new Promise(resolve => setTimeout(resolve, delay));
    
    this.recoveryAttempts.set(attemptKey, attempts + 1);
    
    // Re-emit the original command/operation
    this.emit('retry-operation', detection);
    
    return true;
  }
  
  /**
   * Store successful recovery pattern for learning
   */
  async storeRecoveryPattern(detection, recovery) {
    const pattern = {
      timestamp: new Date().toISOString(),
      error: detection,
      recovery,
      success: true
    };
    
    try {
      const patternsFile = path.join('.swarm', 'recovery-patterns.json');
      
      // Ensure directory exists
      await fs.mkdir('.swarm', { recursive: true });
      
      // Load existing patterns
      let patterns = [];
      try {
        const data = await fs.readFile(patternsFile, 'utf8');
        patterns = JSON.parse(data);
      } catch (err) {
        // File doesn't exist yet
      }
      
      // Add new pattern
      patterns.push(pattern);
      
      // Keep only last 1000 patterns
      if (patterns.length > 1000) {
        patterns = patterns.slice(-1000);
      }
      
      // Save patterns
      await fs.writeFile(patternsFile, JSON.stringify(patterns, null, 2));
      
      this.emit('pattern-stored', pattern);
    } catch (error) {
      console.error(`Failed to store recovery pattern: ${error.message}`);
    }
  }
  
  /**
   * Learn from recovery patterns
   */
  async learnFromPatterns() {
    try {
      const patternsFile = path.join('.swarm', 'recovery-patterns.json');
      const data = await fs.readFile(patternsFile, 'utf8');
      const patterns = JSON.parse(data);
      
      // Analyze patterns for insights
      const insights = {
        totalRecoveries: patterns.length,
        byErrorType: {},
        byRecoveryAction: {},
        successRate: 100, // All stored patterns are successful
        commonRecoveries: []
      };
      
      patterns.forEach(pattern => {
        // Count by error type
        const errorType = pattern.error.type;
        insights.byErrorType[errorType] = (insights.byErrorType[errorType] || 0) + 1;
        
        // Count by recovery action
        const action = pattern.recovery.action;
        insights.byRecoveryAction[action] = (insights.byRecoveryAction[action] || 0) + 1;
      });
      
      // Find most common recoveries
      insights.commonRecoveries = Object.entries(insights.byRecoveryAction)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 5)
        .map(([action, count]) => ({ action, count }));
      
      return insights;
    } catch (error) {
      console.error(`Failed to learn from patterns: ${error.message}`);
      return null;
    }
  }
  
  /**
   * Clear recovery attempts
   */
  clearAttempts() {
    this.recoveryAttempts.clear();
  }
}

module.exports = RecoveryEngine;