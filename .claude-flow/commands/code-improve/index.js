#!/usr/bin/env node

/**
 * Code Improvement Command (/sc:improve)
 * Apply systematic improvements to code quality, performance, and maintainability
 */

const fs = require('fs').promises;
const path = require('path');
const { performance } = require('perf_hooks');

// Import improvement manager
const CodeImprovementManager = require('./code-improvement-manager');

class CodeImproveCLI {
  constructor() {
    this.improvementManager = new CodeImprovementManager();
    this.activeImprovements = new Map();
  }

  async parseArguments(args) {
    const options = {
      target: '.',
      type: 'quality',
      safe: false,
      interactive: false,
      preview: false,
      validate: false,
      help: false
    };

    // Check for help first
    if (args.includes('--help') || args.includes('-h')) {
      options.help = true;
      return options;
    }

    // Find target (first non-flag argument)
    let targetSet = false;
    for (let i = 0; i < args.length; i++) {
      const arg = args[i];

      if (!arg.startsWith('--') && !targetSet) {
        options.target = arg;
        targetSet = true;
        continue;
      }

      switch (arg) {
        case '--type':
          options.type = args[++i] || 'quality';
          break;
        case '--safe':
          options.safe = true;
          break;
        case '--interactive':
          options.interactive = true;
          break;
        case '--preview':
          options.preview = true;
          break;
        case '--validate':
          options.validate = true;
          break;
      }
    }

    return options;
  }

  showHelp() {
    console.log(`
🔧 Code Improvement (/sc:improve) - Systematic Code Enhancement

Usage:
  code-improve [target] [options]

Arguments:
  target                   Target file, directory, or pattern (default: current directory)

Options:
  --type <type>           Improvement type (quality, performance, maintainability, security)
  --safe                  Apply only safe improvements with rollback capability
  --interactive           Interactive mode for complex improvement decisions
  --preview               Preview changes before application
  --validate              Comprehensive validation after improvements
  --help, -h              Show this help message

Improvement Types:
  quality                 - Code structure, readability, technical debt reduction
  performance             - Optimization, bottleneck resolution, efficiency improvements
  maintainability         - Documentation, complexity reduction, modularity enhancement
  security                - Vulnerability fixes, security pattern application

Personas Activated:
  • Architect             - Structure analysis and design improvements
  • Performance Expert    - Speed optimization and bottleneck resolution
  • Quality Engineer      - Code quality and maintainability enhancement
  • Security Specialist   - Vulnerability analysis and security hardening

Examples:
  # Quality enhancement with safe refactoring
  code-improve src/ --type quality --safe

  # Performance optimization with interactive guidance
  code-improve api-endpoints --type performance --interactive

  # Maintainability improvements with preview
  code-improve legacy-modules --type maintainability --preview

  # Security hardening with validation
  code-improve auth-service --type security --validate

Key Features:
  • Multi-persona coordination for domain-specific expertise
  • Framework-specific optimization via Context7 integration
  • Systematic analysis via Sequential MCP for complex improvements
  • Safe refactoring with comprehensive validation and rollback
  • Progress tracking for complex multi-file operations
    `);
  }

  async validateOptions(options) {
    // Validate improvement type
    const validTypes = ['quality', 'performance', 'maintainability', 'security'];
    if (!validTypes.includes(options.type)) {
      throw new Error(`Invalid improvement type: ${options.type}. Valid options: ${validTypes.join(', ')}`);
    }

    // Validate target exists
    try {
      await fs.access(options.target);
    } catch (error) {
      throw new Error(`Target not found: ${options.target}`);
    }

    return true;
  }

  async initializeImprovement(options) {
    console.log('🔧 Initializing Code Improvement...');
    console.log(`📁 Target: ${options.target}`);
    console.log(`🎯 Type: ${options.type}`);
    console.log(`⚙️  Mode: ${this.getMode(options)}`);

    const improvementId = await this.improvementManager.createImprovement({
      target: options.target,
      type: options.type,
      safe: options.safe,
      interactive: options.interactive,
      preview: options.preview,
      validate: options.validate
    });

    this.activeImprovements.set(improvementId, {
      target: options.target,
      startTime: Date.now(),
      options
    });

    return improvementId;
  }

  getMode(options) {
    const modes = [];
    if (options.safe) modes.push('Safe');
    if (options.interactive) modes.push('Interactive');
    if (options.preview) modes.push('Preview');
    if (options.validate) modes.push('Validate');
    return modes.length > 0 ? modes.join(', ') : 'Standard';
  }

  async executeImprovement(improvementId, options) {
    console.log('\n🔄 Executing Code Improvement...');
    
    try {
      const result = await this.improvementManager.executeImprovement(improvementId, {
        target: options.target,
        type: options.type,
        safe: options.safe,
        interactive: options.interactive,
        preview: options.preview,
        validate: options.validate
      });

      this.displayImprovementResults(result, options);

      // Show preview if requested
      if (options.preview && result.changes.length > 0) {
        console.log('\n👀 Preview Mode - Changes Not Applied');
        console.log('Run without --preview to apply improvements');
      }

      // Validate if requested
      if (options.validate && !options.preview) {
        console.log('\n🔍 Running Validation...');
        const validationResult = await this.improvementManager.validateImprovements(improvementId);
        this.displayValidationResults(validationResult);
      }

      return result;
    } catch (error) {
      console.error('❌ Code improvement failed:', error.message);
      throw error;
    }
  }

  displayImprovementResults(result, options) {
    console.log('\n🔧 Code Improvement Results');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━');

    // Summary
    console.log('\n📊 Improvement Summary');
    console.log(`├── Files Analyzed: ${result.summary.filesAnalyzed}`);
    console.log(`├── Issues Found: ${result.summary.issuesFound}`);
    console.log(`├── Improvements Applied: ${result.summary.improvementsApplied}`);
    console.log(`├── Files Modified: ${result.summary.filesModified}`);
    console.log(`└── Execution Time: ${result.summary.executionTime}ms`);

    // Persona Analysis
    if (result.personaAnalysis) {
      console.log('\n👥 Persona Analysis');
      Object.entries(result.personaAnalysis).forEach(([persona, analysis]) => {
        console.log(`├── ${persona}: ${analysis.issuesFound} issues, ${analysis.improvementsApplied} improvements`);
      });
    }

    // Applied Improvements
    if (result.improvements.length > 0) {
      console.log('\n✅ Applied Improvements');
      result.improvements.forEach((improvement, index) => {
        const priorityIcon = improvement.priority === 'high' ? '🔴' : improvement.priority === 'medium' ? '🟡' : '🟢';
        console.log(`${index + 1}. ${priorityIcon} ${improvement.description}`);
        console.log(`   └── File: ${improvement.file} (${improvement.category})`);
        
        if (improvement.impact) {
          console.log(`   └── Impact: ${improvement.impact}`);
        }
      });
    }

    // Quality Metrics
    if (result.metrics) {
      console.log('\n📈 Quality Metrics');
      console.log(`├── Code Quality Score: ${result.metrics.qualityScore}/100`);
      console.log(`├── Maintainability Index: ${result.metrics.maintainabilityIndex}/100`);
      console.log(`├── Technical Debt Reduction: ${result.metrics.technicalDebtReduction}%`);
      console.log(`└── Performance Improvement: ${result.metrics.performanceImprovement}%`);
    }

    // Recommendations
    if (result.recommendations.length > 0) {
      console.log('\n💡 Additional Recommendations');
      result.recommendations.forEach((rec, index) => {
        console.log(`${index + 1}. ${rec.action}`);
        console.log(`   └── ${rec.description}`);
      });
    }

    // Rollback Information
    if (result.rollbackInfo && options.safe) {
      console.log('\n🔄 Rollback Information');
      console.log(`└── Backup created: ${result.rollbackInfo.backupPath}`);
      console.log(`└── Rollback command: ${result.rollbackInfo.rollbackCommand}`);
    }
  }

  displayValidationResults(validationResult) {
    console.log('\n🔍 Validation Results');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━');

    if (validationResult.success) {
      console.log('✅ All improvements validated successfully');
      console.log(`├── Tests Passed: ${validationResult.testsPassed}/${validationResult.totalTests}`);
      console.log(`├── Quality Checks: ${validationResult.qualityChecks} passed`);
      console.log(`└── Performance Impact: ${validationResult.performanceImpact}`);
    } else {
      console.log('❌ Validation failed');
      validationResult.failures.forEach(failure => {
        console.log(`├── ${failure.type}: ${failure.message}`);
      });
      
      if (validationResult.rollbackAvailable) {
        console.log('\n🔄 Rollback available - run: npm run rollback');
      }
    }
  }

  async run(args) {
    try {
      const options = await this.parseArguments(args);

      if (options.help) {
        this.showHelp();
        return;
      }

      await this.validateOptions(options);
      
      const improvementId = await this.initializeImprovement(options);
      const result = await this.executeImprovement(improvementId, options);

      console.log('\n🎉 Code improvement completed successfully!');
      
      if (result.summary.improvementsApplied === 0) {
        console.log('✨ No improvements needed - code quality is already excellent!');
      } else {
        console.log(`🔧 Applied ${result.summary.improvementsApplied} improvements across ${result.summary.filesModified} files`);
        
        if (result.metrics && result.metrics.qualityScore > 90) {
          console.log('🌟 Excellent code quality achieved!');
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
  const cli = new CodeImproveCLI();
  const args = process.argv.slice(2);
  cli.run(args);
}

module.exports = CodeImproveCLI;
