#!/usr/bin/env node

/**
 * Pattern Learning Command
 * Learn patterns from successful operations to improve future performance
 */

const fs = require('fs').promises;
const path = require('path');
const { performance } = require('perf_hooks');

// Import pattern learning manager
const PatternLearningManager = require('./pattern-learning-manager');

class PatternLearnCLI {
  constructor() {
    this.learningManager = new PatternLearningManager();
    this.activeLearning = new Map();
  }

  async parseArguments(args) {
    const options = {
      source: 'all',
      threshold: 0.8,
      save: null,
      export: null,
      analyze: false,
      help: false
    };

    for (let i = 0; i < args.length; i++) {
      const arg = args[i];
      
      switch (arg) {
        case '--source':
          options.source = args[++i];
          break;
        case '--threshold':
          options.threshold = parseFloat(args[++i]) || 0.8;
          break;
        case '--save':
          options.save = args[++i];
          break;
        case '--export':
          options.export = args[++i];
          break;
        case '--analyze':
          options.analyze = true;
          break;
        case '--help':
        case '-h':
          options.help = true;
          break;
      }
    }

    return options;
  }

  showHelp() {
    console.log(`
🧠 Pattern Learn - Extract Patterns from Successful Operations

Usage:
  pattern-learn [options]

Options:
  --source <type>          Pattern source (all, swarm, agents, tasks, communication)
  --threshold <score>      Success threshold (0.0-1.0, default: 0.8)
  --save <name>           Save learned patterns with name
  --export <file>         Export patterns to file
  --analyze               Show detailed pattern analysis
  --help, -h              Show this help message

Pattern Sources:
  all                     - Learn from all successful operations
  swarm                   - Focus on swarm coordination patterns
  agents                  - Learn agent behavior patterns
  tasks                   - Extract task execution patterns
  communication           - Communication efficiency patterns

Success Thresholds:
  0.9-1.0                 - Exceptional performance only
  0.8-0.9                 - High performance operations
  0.7-0.8                 - Good performance operations
  0.6-0.7                 - Acceptable performance operations

Examples:
  # Learn from all successful operations
  pattern-learn

  # High success threshold with analysis
  pattern-learn --threshold 0.9 --analyze

  # Learn communication patterns and save
  pattern-learn --source communication --save comm-patterns

  # Export all patterns for review
  pattern-learn --export patterns.json

Pattern Types Learned:
  • Coordination Patterns  - Successful agent coordination strategies
  • Task Patterns          - Efficient task execution sequences
  • Communication Patterns - Optimal message routing and timing
  • Resource Patterns      - Effective resource allocation strategies
  • Optimization Patterns  - Performance improvement techniques

Output:
  • Pattern Confidence     - Statistical confidence in pattern validity
  • Success Correlation    - How strongly pattern correlates with success
  • Applicability Scope    - Where and when to apply patterns
  • Implementation Guide   - How to implement learned patterns
    `);
  }

  async validateOptions(options) {
    // Validate source
    const validSources = ['all', 'swarm', 'agents', 'tasks', 'communication'];
    if (!validSources.includes(options.source)) {
      throw new Error(`Invalid source: ${options.source}. Valid options: ${validSources.join(', ')}`);
    }

    // Validate threshold
    if (options.threshold < 0.0 || options.threshold > 1.0) {
      throw new Error('Threshold must be between 0.0 and 1.0');
    }

    // Validate export path if provided
    if (options.export) {
      const exportDir = path.dirname(options.export);
      try {
        await fs.access(exportDir);
      } catch (error) {
        throw new Error(`Export directory does not exist: ${exportDir}`);
      }
    }

    return true;
  }

  async initializeLearning(options) {
    console.log('🧠 Initializing Pattern Learning...');
    console.log(`📊 Source: ${options.source}`);
    console.log(`🎯 Success Threshold: ${(options.threshold * 100).toFixed(1)}%`);
    
    if (options.save) {
      console.log(`💾 Save As: ${options.save}`);
    }

    const learningId = await this.learningManager.createLearningSession({
      source: options.source,
      threshold: options.threshold,
      saveAs: options.save
    });

    this.activeLearning.set(learningId, {
      source: options.source,
      startTime: Date.now(),
      options
    });

    return learningId;
  }

  async executeLearning(learningId, options) {
    console.log('\n🔄 Executing Pattern Learning...');
    
    try {
      const result = await this.learningManager.executeLearning(learningId, {
        source: options.source,
        threshold: options.threshold,
        analyze: options.analyze
      });

      this.displayLearningResults(result, options);

      // Save patterns if requested
      if (options.save) {
        await this.savePatterns(result, options.save);
        console.log(`\n💾 Patterns saved as: ${options.save}`);
      }

      // Export if requested
      if (options.export) {
        await this.exportPatterns(result, options.export);
        console.log(`\n📄 Patterns exported to: ${options.export}`);
      }

      return result;
    } catch (error) {
      console.error('❌ Pattern learning failed:', error.message);
      throw error;
    }
  }

  displayLearningResults(result, options) {
    console.log('\n🧠 Pattern Learning Results');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━');

    // Summary
    console.log('\n📊 Learning Summary');
    console.log(`├── Operations Analyzed: ${result.summary.operationsAnalyzed}`);
    console.log(`├── Successful Operations: ${result.summary.successfulOperations}`);
    console.log(`├── Patterns Discovered: ${result.summary.patternsDiscovered}`);
    console.log(`├── High Confidence: ${result.summary.highConfidencePatterns}`);
    console.log(`└── Learning Time: ${result.summary.learningTime}ms`);

    // Discovered Patterns
    if (result.patterns.length > 0) {
      console.log('\n🔍 Discovered Patterns');
      result.patterns.forEach((pattern, index) => {
        const confidenceIcon = pattern.confidence > 0.9 ? '🟢' : pattern.confidence > 0.7 ? '🟡' : '🔴';
        console.log(`${index + 1}. ${confidenceIcon} ${pattern.name} (${(pattern.confidence * 100).toFixed(1)}% confidence)`);
        console.log(`   └── ${pattern.description}`);
        
        if (options.analyze && pattern.details) {
          console.log(`   ├── Success Rate: ${(pattern.successRate * 100).toFixed(1)}%`);
          console.log(`   ├── Sample Size: ${pattern.sampleSize} operations`);
          console.log(`   └── Applicability: ${pattern.applicability}`);
        }
      });
    }

    // Pattern Categories
    if (result.categories) {
      console.log('\n📋 Pattern Categories');
      Object.entries(result.categories).forEach(([category, count]) => {
        console.log(`├── ${category}: ${count} patterns`);
      });
    }

    // Implementation Recommendations
    if (result.recommendations.length > 0) {
      console.log('\n💡 Implementation Recommendations');
      result.recommendations.forEach((rec, index) => {
        console.log(`${index + 1}. ${rec.action}`);
        console.log(`   └── Expected Impact: ${rec.expectedImpact}`);
      });
    }

    // Quality Metrics
    if (result.quality) {
      console.log('\n📈 Pattern Quality Metrics');
      console.log(`├── Average Confidence: ${(result.quality.averageConfidence * 100).toFixed(1)}%`);
      console.log(`├── Pattern Diversity: ${(result.quality.diversity * 100).toFixed(1)}%`);
      console.log(`├── Validation Score: ${(result.quality.validationScore * 100).toFixed(1)}%`);
      console.log(`└── Applicability Range: ${result.quality.applicabilityRange}`);
    }
  }

  async savePatterns(result, saveName) {
    const patternsDir = path.join(process.cwd(), '.claude-flow', 'patterns');
    
    try {
      await fs.mkdir(patternsDir, { recursive: true });
    } catch (error) {
      // Directory might already exist
    }

    const saveData = {
      name: saveName,
      timestamp: new Date().toISOString(),
      patterns: result.patterns,
      summary: result.summary,
      quality: result.quality,
      metadata: {
        version: '1.0.0',
        source: result.source,
        threshold: result.threshold
      }
    };

    const savePath = path.join(patternsDir, `${saveName}.json`);
    await fs.writeFile(savePath, JSON.stringify(saveData, null, 2));
  }

  async exportPatterns(result, exportPath) {
    const exportData = {
      timestamp: new Date().toISOString(),
      learningResults: result,
      metadata: {
        version: '1.0.0',
        generator: 'claude-flow-pattern-learn'
      }
    };

    await fs.writeFile(exportPath, JSON.stringify(exportData, null, 2));
  }

  async run(args) {
    try {
      const options = await this.parseArguments(args);

      if (options.help) {
        this.showHelp();
        return;
      }

      await this.validateOptions(options);
      
      const learningId = await this.initializeLearning(options);
      const result = await this.executeLearning(learningId, options);

      console.log('\n🎉 Pattern learning completed successfully!');
      
      if (result.summary.patternsDiscovered === 0) {
        console.log('📝 No patterns discovered - try lowering the success threshold');
      } else {
        console.log(`🧠 Discovered ${result.summary.patternsDiscovered} patterns for future optimization`);
        
        if (result.summary.highConfidencePatterns > 0) {
          console.log(`✨ ${result.summary.highConfidencePatterns} high-confidence patterns ready for implementation`);
        }
      }

    } catch (error) {
      console.error('❌ Error:', error.message);
      process.exit(1);
    }
  }
}

// CLI execution
if (require.main === module) {
  const cli = new PatternLearnCLI();
  const args = process.argv.slice(2);
  cli.run(args);
}

module.exports = PatternLearnCLI;
