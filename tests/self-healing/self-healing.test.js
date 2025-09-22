/**
 * Self-Healing System Test Suite
 * Tests error detection, recovery, and learning capabilities
 */

const { SelfHealingSystem } = require('../../src/self-healing');
const fs = require('fs').promises;
const path = require('path');

describe('Self-Healing System', () => {
  let system;
  
  beforeEach(() => {
    system = new SelfHealingSystem({
      enabled: true,
      autoRecover: true,
      learningEnabled: true,
      maxRetries: 3
    });
  });
  
  afterEach(async () => {
    // Clean up test files
    try {
      await fs.rmdir('.swarm', { recursive: true });
    } catch (err) {
      // Directory might not exist
    }
  });
  
  describe('Error Detection', () => {
    test('should detect missing module errors', async () => {
      const errorMessage = "Error: Cannot find module 'express'";
      const detection = await system.processError(errorMessage);
      
      expect(detection).toBeDefined();
      expect(detection.type).toBe('dependency');
      expect(detection.severity).toBe('medium');
      expect(detection.recoverable).toBe(true);
      expect(detection.matches).toContain('express');
    });
    
    test('should detect syntax errors', async () => {
      const errorMessage = "SyntaxError: Unexpected token }";
      const detection = await system.processError(errorMessage);
      
      expect(detection).toBeDefined();
      expect(detection.type).toBe('syntax');
      expect(detection.severity).toBe('high');
      expect(detection.recoverable).toBe(true);
    });
    
    test('should detect test failures', async () => {
      const errorMessage = "Test failed: 5 test(s) failed";
      const detection = await system.processError(errorMessage);
      
      expect(detection).toBeDefined();
      expect(detection.type).toBe('test');
      expect(detection.severity).toBe('medium');
      expect(detection.recoverable).toBe(true);
    });
    
    test('should detect TypeScript errors', async () => {
      const errorMessage = "TS2345: Argument of type 'string' is not assignable";
      const detection = await system.processError(errorMessage);
      
      expect(detection).toBeDefined();
      expect(detection.type).toBe('typescript');
      expect(detection.severity).toBe('medium');
      expect(detection.recoverable).toBe(true);
    });
    
    test('should detect port in use errors', async () => {
      const errorMessage = "Error: EADDRINUSE: address already in use :::3000";
      const detection = await system.processError(errorMessage);
      
      expect(detection).toBeDefined();
      expect(detection.type).toBe('runtime');
      expect(detection.severity).toBe('low');
      expect(detection.recoverable).toBe(true);
      expect(detection.matches).toContain('3000');
    });
    
    test('should detect memory errors', async () => {
      const errorMessage = "FATAL ERROR: JavaScript heap out of memory";
      const detection = await system.processError(errorMessage);
      
      expect(detection).toBeDefined();
      expect(detection.type).toBe('memory');
      expect(detection.severity).toBe('critical');
      expect(detection.recoverable).toBe(true);
    });
    
    test('should handle unknown errors', async () => {
      const errorMessage = "Some random error message";
      const detection = await system.processError(errorMessage);
      
      expect(detection).toBeDefined();
      expect(detection.type).toBe('unknown');
      expect(detection.severity).toBe('low');
      expect(detection.recoverable).toBe(false);
    });
  });
  
  describe('Recovery Strategies', () => {
    test('should get recovery strategy for missing module', () => {
      const errorMessage = "Cannot find module 'lodash'";
      const suggestions = system.getRecoverySuggestions(errorMessage);
      
      expect(suggestions.defaultStrategy).toBeDefined();
      expect(suggestions.defaultStrategy.action).toBe('install-dependency');
      expect(suggestions.defaultStrategy.command).toContain('npm install lodash');
    });
    
    test('should get recovery strategy for syntax error', () => {
      const errorMessage = "SyntaxError: Unexpected token }";
      const suggestions = system.getRecoverySuggestions(errorMessage);
      
      expect(suggestions.defaultStrategy).toBeDefined();
      expect(suggestions.defaultStrategy.action).toBe('analyze-and-fix');
    });
    
    test('should get recovery strategy for port in use', () => {
      const errorMessage = "EADDRINUSE: address already in use :::8080";
      const suggestions = system.getRecoverySuggestions(errorMessage);
      
      expect(suggestions.defaultStrategy).toBeDefined();
      expect(suggestions.defaultStrategy.action).toBe('change-port');
    });
    
    test('should generate recommendations', () => {
      const errorMessage = "JavaScript heap out of memory";
      const suggestions = system.getRecoverySuggestions(errorMessage);
      
      expect(suggestions.recommendations).toBeDefined();
      expect(suggestions.recommendations.length).toBeGreaterThan(0);
      
      const highPriority = suggestions.recommendations.find(r => r.priority === 'high');
      expect(highPriority).toBeDefined();
    });
  });
  
  describe('Learning System', () => {
    test('should learn from errors', async () => {
      // Process same error multiple times
      const errorMessage = "Cannot find module 'axios'";
      
      await system.processError(errorMessage);
      await system.processError(errorMessage);
      await system.processError(errorMessage);
      
      const insights = system.learningSystem.getInsights();
      
      expect(insights.totalErrors).toBeGreaterThan(0);
      expect(insights.topErrors.length).toBeGreaterThan(0);
    });
    
    test('should predict error likelihood', async () => {
      // Train the system with errors
      await system.processError("Cannot find module 'express'", { file: 'server.js' });
      await system.processError("Cannot find module 'express'", { file: 'server.js' });
      
      // Predict errors for similar context
      const predictions = system.predictErrors({ file: 'server.js' });
      
      expect(predictions).toBeDefined();
      expect(Array.isArray(predictions)).toBe(true);
    });
    
    test('should save and load knowledge base', async () => {
      // Add some learning data
      await system.processError("Cannot find module 'react'");
      await system.processError("SyntaxError: Unexpected token");
      
      // Save knowledge base
      await system.learningSystem.saveKnowledgeBase();
      
      // Check if file exists
      const kbPath = path.join('.swarm', 'knowledge-base.json');
      const stats = await fs.stat(kbPath);
      expect(stats.isFile()).toBe(true);
      
      // Load knowledge base in new instance
      const newSystem = new SelfHealingSystem();
      await newSystem.learningSystem.loadKnowledgeBase();
      
      const insights = newSystem.learningSystem.getInsights();
      expect(insights.totalErrors).toBeGreaterThan(0);
    });
  });
  
  describe('Statistics', () => {
    test('should track error statistics', async () => {
      await system.processError("Error 1");
      await system.processError("Error 2");
      await system.processError("Error 3");
      
      const stats = system.getStats();
      
      expect(stats.errorsDetected).toBe(3);
      expect(stats.recoveryRate).toBeDefined();
      expect(stats.runtime).toBeDefined();
    });
    
    test('should analyze error patterns', async () => {
      // Create pattern by repeating errors
      await system.processError("Cannot find module 'express'");
      await system.processError("Cannot find module 'express'");
      await system.processError("SyntaxError: test");
      
      const stats = system.getStats();
      
      expect(stats.errorPatterns).toBeDefined();
      expect(stats.errorPatterns.totalErrors).toBe(3);
      expect(stats.errorPatterns.byType).toBeDefined();
    });
    
    test('should provide learning insights', async () => {
      await system.processError("Test error");
      
      const stats = system.getStats();
      
      expect(stats.learningInsights).toBeDefined();
      expect(stats.learningInsights.totalErrors).toBeGreaterThanOrEqual(0);
      expect(stats.learningInsights.recommendations).toBeDefined();
    });
  });
  
  describe('System Control', () => {
    test('should enable/disable self-healing', () => {
      expect(system.options.enabled).toBe(true);
      
      system.setEnabled(false);
      expect(system.options.enabled).toBe(false);
      
      system.setEnabled(true);
      expect(system.options.enabled).toBe(true);
    });
    
    test('should enable/disable auto-recovery', () => {
      expect(system.options.autoRecover).toBe(true);
      
      system.setAutoRecover(false);
      expect(system.options.autoRecover).toBe(false);
      
      system.setAutoRecover(true);
      expect(system.options.autoRecover).toBe(true);
    });
    
    test('should enable/disable learning', () => {
      expect(system.options.learningEnabled).toBe(true);
      
      system.setLearningEnabled(false);
      expect(system.options.learningEnabled).toBe(false);
      
      system.setLearningEnabled(true);
      expect(system.options.learningEnabled).toBe(true);
    });
    
    test('should reset system', async () => {
      // Add some data
      await system.processError("Test error");
      expect(system.getStats().errorsDetected).toBe(1);
      
      // Reset
      await system.reset();
      
      // Check reset
      expect(system.getStats().errorsDetected).toBe(0);
    });
  });
  
  describe('Export/Import', () => {
    test('should export system data', async () => {
      // Add some data
      await system.processError("Cannot find module 'test'");
      await system.processError("SyntaxError: test");
      
      const exportData = await system.exportData();
      
      expect(exportData).toBeDefined();
      expect(exportData.stats).toBeDefined();
      expect(exportData.patterns).toBeDefined();
      expect(exportData.learningExport).toBeDefined();
    });
    
    test('should import system data', async () => {
      // Create export data
      const exportData = {
        patterns: {
          patterns: [['test-pattern', {
            regex: /test/,
            type: 'test',
            severity: 'low',
            recoverable: true
          }]]
        }
      };
      
      await system.importData(exportData);
      
      // Test imported pattern
      const detection = await system.processError("test error");
      expect(detection.type).toBe('test');
    });
  });
  
  describe('Event System', () => {
    test('should emit error-detected event', (done) => {
      system.on('error-detected', (detection) => {
        expect(detection).toBeDefined();
        expect(detection.type).toBe('dependency');
        done();
      });
      
      system.processError("Cannot find module 'test'");
    });
    
    test('should emit recovery events', (done) => {
      let eventCount = 0;
      
      system.on('recovery-success', () => {
        eventCount++;
      });
      
      system.on('recovery-failed', () => {
        eventCount++;
      });
      
      system.on('unrecoverable-error', () => {
        eventCount++;
        expect(eventCount).toBeGreaterThan(0);
        done();
      });
      
      system.processError("Unknown error type");
    });
  });
});

// Run tests if Jest is available
if (typeof jest !== 'undefined') {
  console.log('🧪 Running Self-Healing System Tests...');
} else {
  console.log('ℹ️  Test suite created. Run with Jest to execute tests.');
}