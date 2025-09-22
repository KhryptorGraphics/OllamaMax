#!/usr/bin/env node

/**
 * CLI Integration for Self-Healing System
 * Provides command-line interface for self-healing features
 */

const { getSelfHealingSystem } = require('./index');
const { exec } = require('child_process');
const { promisify } = require('util');

const execAsync = promisify(exec);

class SelfHealingCLI {
  constructor() {
    this.system = getSelfHealingSystem({
      enabled: true,
      autoRecover: true,
      learningEnabled: true
    });
    
    this.setupProcessHandlers();
  }
  
  /**
   * Setup process error handlers
   */
  setupProcessHandlers() {
    // Catch uncaught exceptions
    process.on('uncaughtException', (error) => {
      console.error('🚨 Uncaught Exception:', error.message);
      this.handleProcessError(error);
    });
    
    // Catch unhandled promise rejections
    process.on('unhandledRejection', (reason, promise) => {
      console.error('🚨 Unhandled Rejection:', reason);
      this.handleProcessError(new Error(String(reason)));
    });
    
    // Catch process warnings
    process.on('warning', (warning) => {
      console.warn('⚠️  Process Warning:', warning.message);
      this.system.processError(warning.message, { type: 'warning' });
    });
  }
  
  /**
   * Handle process-level errors
   */
  async handleProcessError(error) {
    const detection = await this.system.processError(error.message, {
      stack: error.stack,
      type: 'process-error'
    });
    
    if (detection && detection.recoverable) {
      console.log('🔄 Attempting automatic recovery...');
    }
  }
  
  /**
   * Wrap command execution with self-healing
   */
  async executeWithHealing(command, options = {}) {
    console.log(`🔧 Executing: ${command}`);
    
    try {
      const { stdout, stderr } = await execAsync(command, options);
      
      // Check for errors in stderr
      if (stderr && stderr.trim()) {
        await this.system.processError(stderr, {
          command,
          type: 'command-stderr'
        });
      }
      
      return { success: true, stdout, stderr };
    } catch (error) {
      console.error(`❌ Command failed: ${error.message}`);
      
      // Process error through self-healing system
      const detection = await this.system.processError(error.message, {
        command,
        exitCode: error.code,
        type: 'command-error'
      });
      
      // If recovery was successful, retry the command
      if (detection && detection.recoverable) {
        console.log('🔄 Retrying command after recovery...');
        
        // Wait for recovery to complete
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        try {
          const { stdout, stderr } = await execAsync(command, options);
          console.log('✅ Command succeeded after recovery');
          return { success: true, stdout, stderr, recovered: true };
        } catch (retryError) {
          console.error('❌ Command failed after recovery attempt');
          return { success: false, error: retryError, recovered: false };
        }
      }
      
      return { success: false, error, recovered: false };
    }
  }
  
  /**
   * Monitor command execution
   */
  async monitorCommand(command, options = {}) {
    const startTime = Date.now();
    const result = await this.executeWithHealing(command, options);
    const duration = Date.now() - startTime;
    
    // Log execution metrics
    console.log(`⏱️  Execution time: ${duration}ms`);
    
    if (result.recovered) {
      console.log('🎉 Self-healing successful!');
    }
    
    return result;
  }
  
  /**
   * Process CLI commands
   */
  async processCommand(args) {
    const command = args[0];
    
    switch (command) {
      case 'status':
        return this.showStatus();
        
      case 'enable':
        return this.enableFeature(args[1]);
        
      case 'disable':
        return this.disableFeature(args[1]);
        
      case 'stats':
        return this.showStats();
        
      case 'learn':
        return this.showLearning();
        
      case 'predict':
        return this.predictErrors(args.slice(1).join(' '));
        
      case 'suggest':
        return this.suggestRecovery(args.slice(1).join(' '));
        
      case 'export':
        return this.exportData(args[1]);
        
      case 'import':
        return this.importData(args[1]);
        
      case 'reset':
        return this.resetSystem();
        
      case 'exec':
        return this.monitorCommand(args.slice(1).join(' '));
        
      default:
        return this.showHelp();
    }
  }
  
  /**
   * Show system status
   */
  showStatus() {
    console.log('\n🏥 Self-Healing System Status\n');
    console.log(`Enabled: ${this.system.options.enabled ? '✅' : '❌'}`);
    console.log(`Auto-Recovery: ${this.system.options.autoRecover ? '✅' : '❌'}`);
    console.log(`Learning: ${this.system.options.learningEnabled ? '✅' : '❌'}`);
    console.log(`Max Retries: ${this.system.options.maxRetries}`);
    console.log('');
  }
  
  /**
   * Show statistics
   */
  showStats() {
    const stats = this.system.getStats();
    
    console.log('\n📊 Self-Healing Statistics\n');
    console.log(`Runtime: ${stats.runtime}`);
    console.log(`Errors Detected: ${stats.errorsDetected}`);
    console.log(`Recoveries Attempted: ${stats.recoveriesAttempted}`);
    console.log(`Recoveries Successful: ${stats.recoveriesSuccessful}`);
    console.log(`Recovery Rate: ${stats.recoveryRate}`);
    
    if (stats.errorPatterns.commonErrors.length > 0) {
      console.log('\n🔝 Common Errors:');
      stats.errorPatterns.commonErrors.forEach(error => {
        console.log(`  - ${error.name}: ${error.count} occurrences`);
      });
    }
    
    if (stats.learningInsights.topErrors.length > 0) {
      console.log('\n📈 Top Error Types:');
      stats.learningInsights.topErrors.forEach(error => {
        console.log(`  - ${error.type}: ${error.occurrences} times (${(error.recoveryRate * 100).toFixed(1)}% recovery)`);
      });
    }
    
    console.log('');
  }
  
  /**
   * Show learning insights
   */
  showLearning() {
    const insights = this.system.learningSystem.getInsights();
    
    console.log('\n🧠 Learning Insights\n');
    console.log(`Knowledge Entries: ${insights.knowledgeEntries}`);
    console.log(`Error Patterns: ${insights.totalErrors}`);
    console.log(`Recovery Strategies: ${insights.totalRecoveries}`);
    
    if (insights.bestRecoveries.length > 0) {
      console.log('\n🏆 Best Recovery Strategies:');
      insights.bestRecoveries.forEach(recovery => {
        console.log(`  - ${recovery.strategy}: ${(recovery.successRate * 100).toFixed(1)}% success (${recovery.attempts} attempts)`);
      });
    }
    
    if (insights.recommendations.length > 0) {
      console.log('\n💡 Recommendations:');
      insights.recommendations.forEach(rec => {
        console.log(`  - ${rec.suggestion} (${rec.reason})`);
      });
    }
    
    console.log('');
  }
  
  /**
   * Predict errors based on context
   */
  predictErrors(context) {
    const predictions = this.system.predictErrors({ context });
    
    if (predictions.length === 0) {
      console.log('\n✨ No errors predicted for this context\n');
      return;
    }
    
    console.log('\n🔮 Error Predictions\n');
    predictions.forEach(pred => {
      console.log(`${pred.error}:`);
      console.log(`  Likelihood: ${(pred.likelihood * 100).toFixed(1)}%`);
      console.log(`  Preventable: ${pred.preventable ? '✅' : '❌'}`);
    });
    console.log('');
  }
  
  /**
   * Suggest recovery for error
   */
  suggestRecovery(errorMessage) {
    const suggestions = this.system.getRecoverySuggestions(errorMessage);
    
    console.log('\n💊 Recovery Suggestions\n');
    console.log(`Error Type: ${suggestions.detection.type}`);
    console.log(`Severity: ${suggestions.detection.severity}`);
    console.log(`Recoverable: ${suggestions.detection.recoverable ? '✅' : '❌'}`);
    
    if (suggestions.recommendations.length > 0) {
      console.log('\nRecommended Actions:');
      suggestions.recommendations.forEach(rec => {
        console.log(`  [${rec.priority.toUpperCase()}] ${rec.action}`);
        console.log(`    Reason: ${rec.reason}`);
      });
    }
    
    console.log('');
  }
  
  /**
   * Enable feature
   */
  enableFeature(feature) {
    switch (feature) {
      case 'auto-recover':
        this.system.setAutoRecover(true);
        break;
      case 'learning':
        this.system.setLearningEnabled(true);
        break;
      default:
        this.system.setEnabled(true);
    }
    console.log(`✅ ${feature || 'Self-healing'} enabled`);
  }
  
  /**
   * Disable feature
   */
  disableFeature(feature) {
    switch (feature) {
      case 'auto-recover':
        this.system.setAutoRecover(false);
        break;
      case 'learning':
        this.system.setLearningEnabled(false);
        break;
      default:
        this.system.setEnabled(false);
    }
    console.log(`❌ ${feature || 'Self-healing'} disabled`);
  }
  
  /**
   * Export system data
   */
  async exportData(filename) {
    const file = filename || `self-healing-export-${Date.now()}.json`;
    const data = await this.system.exportData();
    
    const fs = require('fs').promises;
    await fs.writeFile(file, JSON.stringify(data, null, 2));
    
    console.log(`✅ Data exported to ${file}`);
  }
  
  /**
   * Import system data
   */
  async importData(filename) {
    if (!filename) {
      console.error('❌ Please provide a filename to import');
      return;
    }
    
    const fs = require('fs').promises;
    const data = JSON.parse(await fs.readFile(filename, 'utf8'));
    
    await this.system.importData(data);
    console.log(`✅ Data imported from ${filename}`);
  }
  
  /**
   * Reset system
   */
  async resetSystem() {
    await this.system.reset();
    console.log('✅ System reset complete');
  }
  
  /**
   * Show help
   */
  showHelp() {
    console.log(`
🏥 Self-Healing System CLI

Usage: self-heal [command] [options]

Commands:
  status              Show system status
  enable [feature]    Enable self-healing or specific feature
  disable [feature]   Disable self-healing or specific feature
  stats              Show statistics
  learn              Show learning insights
  predict <context>   Predict errors for context
  suggest <error>     Suggest recovery for error
  export [file]       Export system data
  import <file>       Import system data
  reset              Reset system
  exec <command>      Execute command with self-healing
  help               Show this help

Features:
  auto-recover       Automatic error recovery
  learning          Pattern learning system

Examples:
  self-heal status
  self-heal enable auto-recover
  self-heal stats
  self-heal exec "npm test"
  self-heal predict "deploying to production"
  self-heal suggest "Cannot find module 'express'"
`);
  }
}

// Export for module usage
module.exports = SelfHealingCLI;

// Run CLI if executed directly
if (require.main === module) {
  const cli = new SelfHealingCLI();
  const args = process.argv.slice(2);
  
  if (args.length === 0) {
    cli.showHelp();
  } else {
    cli.processCommand(args).catch(error => {
      console.error('❌ CLI Error:', error.message);
      process.exit(1);
    });
  }
}