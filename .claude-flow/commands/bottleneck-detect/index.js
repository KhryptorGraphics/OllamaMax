#!/usr/bin/env node

/**
 * Bottleneck Detection Command
 * Analyzes performance bottlenecks in swarm operations and suggests optimizations
 */

const fs = require('fs').promises;
const path = require('path');
const { performance } = require('perf_hooks');

// Import bottleneck analysis manager
const BottleneckAnalyzer = require('./bottleneck-analyzer');

class BottleneckDetectCLI {
  constructor() {
    this.analyzer = new BottleneckAnalyzer();
    this.activeAnalyses = new Map();
  }

  async parseArguments(args) {
    const options = {
      swarmId: null,
      timeRange: '1h',
      threshold: 20,
      export: null,
      fix: false,
      help: false
    };

    for (let i = 0; i < args.length; i++) {
      const arg = args[i];
      
      switch (arg) {
        case '--swarm-id':
        case '-s':
          options.swarmId = args[++i];
          break;
        case '--time-range':
        case '-t':
          options.timeRange = args[++i];
          break;
        case '--threshold':
          options.threshold = parseInt(args[++i]) || 20;
          break;
        case '--export':
        case '-e':
          options.export = args[++i];
          break;
        case '--fix':
          options.fix = true;
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
🔍 Bottleneck Detect - Performance Analysis & Optimization

Usage:
  bottleneck-detect [options]

Options:
  --swarm-id, -s <id>      Analyze specific swarm (default: current)
  --time-range, -t <range> Analysis period: 1h, 24h, 7d, all (default: 1h)
  --threshold <percent>    Bottleneck threshold percentage (default: 20)
  --export, -e <file>      Export analysis to file
  --fix                    Apply automatic optimizations
  --help, -h              Show this help message

Time Ranges:
  1h          - Last 1 hour
  24h         - Last 24 hours
  7d          - Last 7 days
  all         - All available data

Examples:
  # Basic bottleneck detection
  bottleneck-detect

  # Analyze specific swarm
  bottleneck-detect --swarm-id swarm-123

  # Last 24 hours with export
  bottleneck-detect -t 24h -e bottlenecks.json

  # Auto-fix detected issues
  bottleneck-detect --fix --threshold 15

Metrics Analyzed:
  • Communication Bottlenecks - Message delays, response times
  • Processing Bottlenecks    - Task completion, agent utilization
  • Memory Bottlenecks        - Cache performance, I/O patterns
  • Network Bottlenecks       - API latency, service timeouts

Automatic Fixes:
  • Topology optimization     - Switch to efficient patterns
  • Caching enhancement      - Enable smart caching
  • Concurrency tuning       - Adjust agent counts
  • Priority adjustment      - Optimize task queues
    `);
  }

  async validateOptions(options) {
    // Validate time range
    const validRanges = ['1h', '24h', '7d', 'all'];
    if (!validRanges.includes(options.timeRange)) {
      throw new Error(`Invalid time range: ${options.timeRange}. Valid options: ${validRanges.join(', ')}`);
    }

    // Validate threshold
    if (options.threshold < 1 || options.threshold > 100) {
      throw new Error('Threshold must be between 1 and 100 percent');
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

  async initializeAnalysis(options) {
    console.log('🔍 Initializing Bottleneck Analysis...');
    console.log(`📊 Swarm: ${options.swarmId || 'current'}`);
    console.log(`⏱️  Time Range: ${options.timeRange}`);
    console.log(`🎯 Threshold: ${options.threshold}%`);
    
    if (options.fix) {
      console.log('🔧 Auto-fix enabled');
    }

    const analysisId = await this.analyzer.createAnalysis({
      swarmId: options.swarmId,
      timeRange: options.timeRange,
      threshold: options.threshold,
      autoFix: options.fix
    });

    this.activeAnalyses.set(analysisId, {
      swarmId: options.swarmId,
      startTime: Date.now(),
      options
    });

    return analysisId;
  }

  async executeAnalysis(analysisId, options) {
    console.log('\n🔄 Executing Bottleneck Analysis...');
    
    try {
      const result = await this.analyzer.executeAnalysis(analysisId, {
        swarmId: options.swarmId,
        timeRange: options.timeRange,
        threshold: options.threshold,
        autoFix: options.fix
      });

      this.displayAnalysisReport(result);

      // Export if requested
      if (options.export) {
        await this.exportAnalysis(result, options.export);
        console.log(`\n📄 Analysis exported to: ${options.export}`);
      }

      // Apply fixes if requested
      if (options.fix && result.quickFixes.length > 0) {
        console.log('\n🔧 Applying automatic optimizations...');
        const fixResults = await this.analyzer.applyOptimizations(analysisId, result.quickFixes);
        this.displayFixResults(fixResults);
      }

      return result;
    } catch (error) {
      console.error('❌ Analysis execution failed:', error.message);
      throw error;
    }
  }

  displayAnalysisReport(result) {
    console.log('\n🔍 Bottleneck Analysis Report');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━');

    // Summary
    console.log('\n📊 Summary');
    console.log(`├── Time Range: ${result.summary.timeRange}`);
    console.log(`├── Agents Analyzed: ${result.summary.agentsAnalyzed}`);
    console.log(`├── Tasks Processed: ${result.summary.tasksProcessed}`);
    console.log(`└── Critical Issues: ${result.summary.criticalIssues}`);

    // Critical Bottlenecks
    if (result.criticalBottlenecks.length > 0) {
      console.log('\n🚨 Critical Bottlenecks');
      result.criticalBottlenecks.forEach((bottleneck, index) => {
        console.log(`${index + 1}. ${bottleneck.name} (${bottleneck.impact}% impact)`);
        console.log(`   └── ${bottleneck.description}`);
      });
    }

    // Warning Bottlenecks
    if (result.warningBottlenecks.length > 0) {
      console.log('\n⚠️ Warning Bottlenecks');
      result.warningBottlenecks.forEach((bottleneck, index) => {
        console.log(`${index + 1}. ${bottleneck.name} (${bottleneck.impact}% impact)`);
        console.log(`   └── ${bottleneck.description}`);
      });
    }

    // Recommendations
    if (result.recommendations.length > 0) {
      console.log('\n💡 Recommendations');
      result.recommendations.forEach((rec, index) => {
        console.log(`${index + 1}. ${rec.action} (est. ${rec.improvement}% improvement)`);
      });
    }

    // Quick Fixes
    if (result.quickFixes.length > 0) {
      console.log('\n✅ Quick Fixes Available');
      console.log('Run with --fix to apply:');
      result.quickFixes.forEach(fix => {
        console.log(`- ${fix.description}`);
      });
    }

    // Performance Metrics
    if (result.metrics) {
      console.log('\n📈 Performance Metrics');
      console.log(`├── Communication Efficiency: ${result.metrics.communicationEfficiency}%`);
      console.log(`├── Processing Efficiency: ${result.metrics.processingEfficiency}%`);
      console.log(`├── Memory Efficiency: ${result.metrics.memoryEfficiency}%`);
      console.log(`└── Overall Score: ${result.metrics.overallScore}%`);
    }
  }

  displayFixResults(fixResults) {
    console.log('\n🔧 Optimization Results');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━');

    fixResults.forEach((fix, index) => {
      const status = fix.success ? '✅' : '❌';
      console.log(`${status} ${fix.description}`);
      if (fix.improvement) {
        console.log(`   └── Improvement: ${fix.improvement}%`);
      }
      if (fix.error) {
        console.log(`   └── Error: ${fix.error}`);
      }
    });

    const successfulFixes = fixResults.filter(f => f.success).length;
    const totalImprovement = fixResults
      .filter(f => f.success && f.improvement)
      .reduce((sum, f) => sum + f.improvement, 0);

    console.log(`\n📊 Applied ${successfulFixes}/${fixResults.length} optimizations`);
    if (totalImprovement > 0) {
      console.log(`🚀 Estimated total improvement: ${Math.round(totalImprovement)}%`);
    }
  }

  async exportAnalysis(result, exportPath) {
    const exportData = {
      timestamp: new Date().toISOString(),
      analysis: result,
      metadata: {
        version: '1.0.0',
        generator: 'claude-flow-bottleneck-detect'
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
      
      const analysisId = await this.initializeAnalysis(options);
      const result = await this.executeAnalysis(analysisId, options);

      console.log('\n🎉 Bottleneck analysis completed successfully!');
      
      if (result.summary.criticalIssues === 0) {
        console.log('✨ No critical bottlenecks detected - system is performing well!');
      } else {
        console.log(`🔧 ${result.summary.criticalIssues} critical issues found - consider applying fixes`);
      }

    } catch (error) {
      console.error('❌ Error:', error.message);
      process.exit(1);
    }
  }
}

// CLI execution
if (require.main === module) {
  const cli = new BottleneckDetectCLI();
  const args = process.argv.slice(2);
  cli.run(args);
}

module.exports = BottleneckDetectCLI;
