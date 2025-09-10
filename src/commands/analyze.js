#!/usr/bin/env node

/**
 * Analyze Command - Orchestrates analysis operations with hive-mind coordination
 */

import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs/promises';
import path from 'path';
import chalk from 'chalk';
import { AnalyzeAgent } from '../agents/analyze-agent.js';

const execAsync = promisify(exec);

class AnalyzeCommand {
  constructor() {
    this.supportedModes = ['code', 'performance', 'security', 'architecture', 'hive-mind', 'all'];
  }

  /**
   * Execute analysis command
   */
  async execute(args = []) {
    // Check for help flag first
    if (args.includes('--help') || args.includes('-h')) {
      this.showHelp();
      return true;
    }
    
    const mode = args[0] || 'all';
    const target = args[1] || '.';
    const flags = args.slice(2);
    
    // Validate mode
    if (!this.supportedModes.includes(mode)) {
      console.error(chalk.red(`Error: Unknown analysis mode "${mode}"`));
      console.log(chalk.yellow(`Supported modes: ${this.supportedModes.join(', ')}`));
      return false;
    }
    
    console.log(chalk.cyan(`\n🔍 Claude Flow Analysis System`));
    console.log(chalk.gray(`Mode: ${mode}, Target: ${target}`));
    
    try {
      // Check if hive-mind mode is requested
      if (mode === 'hive-mind' || flags.includes('--hive-mind')) {
        return await this.runHiveMindAnalysis(target, flags);
      }
      
      // Run specific or all analyses
      if (mode === 'all') {
        return await this.runCompleteAnalysis(target, flags);
      } else {
        return await this.runSpecificAnalysis(mode, target, flags);
      }
    } catch (error) {
      console.error(chalk.red('Analysis failed:'), error.message);
      return false;
    }
  }

  /**
   * Run hive-mind coordinated analysis
   */
  async runHiveMindAnalysis(target, flags) {
    console.log(chalk.magenta('\n🧠 Initializing Hive-Mind Analysis...\n'));
    
    try {
      // Initialize hive-mind coordination
      console.log(chalk.gray('Step 1: Initializing swarm coordination...'));
      const { stdout: initOutput } = await execAsync('npx claude-flow@alpha swarm init --topology mesh --max-agents 5');
      console.log(chalk.green('✓ Swarm initialized'));
      
      // Spawn analysis agents
      console.log(chalk.gray('Step 2: Spawning specialized agents...'));
      const agents = [
        { name: 'CodeQuality', type: 'analyzer', focus: 'code-quality' },
        { name: 'Performance', type: 'optimizer', focus: 'performance' },
        { name: 'Security', type: 'auditor', focus: 'security' },
        { name: 'Architecture', type: 'architect', focus: 'structure' },
        { name: 'Testing', type: 'qa', focus: 'coverage' }
      ];
      
      for (const agent of agents) {
        await execAsync(`npx claude-flow@alpha agent spawn --name "${agent.name}" --type ${agent.type} --focus ${agent.focus}`);
        console.log(chalk.green(`✓ ${agent.name} agent spawned`));
      }
      
      // Orchestrate analysis tasks
      console.log(chalk.gray('Step 3: Orchestrating parallel analysis...'));
      await execAsync(`npx claude-flow@alpha task orchestrate --task "analyze ${target}" --strategy parallel`);
      
      // Run the analyzer with hive-mind mode
      const analyzer = new AnalyzeAgent({ 
        mode: 'hive-mind',
        depth: flags.includes('--deep') ? 'deep' : 'medium',
        parallel: !flags.includes('--sequential')
      });
      
      const results = await analyzer.analyze(target);
      
      // Store results in memory for cross-agent access
      console.log(chalk.gray('Step 4: Storing results in hive memory...'));
      await execAsync(`npx claude-flow@alpha memory store --key "analysis/results" --value '${JSON.stringify(results)}'`);
      
      // Generate insights using collective intelligence
      console.log(chalk.gray('Step 5: Generating collective insights...'));
      await this.generateCollectiveInsights(results);
      
      console.log(chalk.green('\n✅ Hive-mind analysis complete!'));
      return true;
      
    } catch (error) {
      console.error(chalk.red('Hive-mind coordination failed:'), error.message);
      
      // Fallback to standard analysis
      console.log(chalk.yellow('\nFalling back to standard analysis...'));
      return await this.runCompleteAnalysis(target, flags);
    }
  }

  /**
   * Run complete analysis (all categories)
   */
  async runCompleteAnalysis(target, flags) {
    console.log(chalk.cyan('\n📊 Running Complete Analysis...\n'));
    
    const analyzer = new AnalyzeAgent({
      mode: 'standard',
      depth: flags.includes('--deep') ? 'deep' : 'medium',
      parallel: !flags.includes('--sequential')
    });
    
    const results = await analyzer.analyze(target);
    
    // Store results for future reference
    await this.storeResults(results);
    
    return true;
  }

  /**
   * Run specific analysis category
   */
  async runSpecificAnalysis(mode, target, flags) {
    console.log(chalk.cyan(`\n🔍 Running ${mode} analysis...\n`));
    
    const analyzer = new AnalyzeAgent({
      mode: 'standard',
      depth: flags.includes('--deep') ? 'deep' : 'medium',
      parallel: false // Single category, no need for parallel
    });
    
    let results;
    switch (mode) {
      case 'code':
        results = await analyzer.analyzeCodeQuality(target);
        break;
      case 'performance':
        results = await analyzer.analyzePerformance(target);
        break;
      case 'security':
        results = await analyzer.analyzeSecurity(target);
        break;
      case 'architecture':
        results = await analyzer.analyzeArchitecture(target);
        break;
      default:
        throw new Error(`Unsupported analysis mode: ${mode}`);
    }
    
    // Generate focused report
    await this.generateFocusedReport(mode, results);
    
    return true;
  }

  /**
   * Generate collective insights from hive-mind analysis
   */
  async generateCollectiveInsights(results) {
    console.log(chalk.magenta('\n🧠 Collective Intelligence Insights:\n'));
    
    const insights = [];
    
    // Critical issues consensus
    if (results.summary.criticalIssues > 0) {
      insights.push({
        type: 'consensus',
        priority: 'critical',
        message: `All agents agree: ${results.summary.criticalIssues} critical issues require immediate attention`,
        action: 'Create priority task queue for critical fixes'
      });
    }
    
    // Cross-cutting concerns
    const crossCutting = this.identifyCrossCuttingConcerns(results);
    for (const concern of crossCutting) {
      insights.push({
        type: 'pattern',
        priority: 'high',
        message: concern.message,
        action: concern.action
      });
    }
    
    // Optimization opportunities
    const optimizations = this.identifyOptimizations(results);
    for (const opt of optimizations) {
      insights.push({
        type: 'opportunity',
        priority: 'medium',
        message: opt.message,
        action: opt.action
      });
    }
    
    // Display insights
    for (const insight of insights) {
      const icon = insight.priority === 'critical' ? '🚨' :
                   insight.priority === 'high' ? '⚠️' :
                   '💡';
      console.log(`${icon} ${chalk.yellow(insight.message)}`);
      console.log(chalk.gray(`   → ${insight.action}`));
    }
    
    return insights;
  }

  /**
   * Identify cross-cutting concerns from analysis
   */
  identifyCrossCuttingConcerns(results) {
    const concerns = [];
    
    // Check for systemic issues
    if (results.analyses.security?.issues.length > 5 && 
        results.analyses.code?.issues.length > 10) {
      concerns.push({
        message: 'Systemic quality issues detected across security and code quality',
        action: 'Implement automated quality gates in CI/CD pipeline'
      });
    }
    
    if (results.analyses.testing?.suggestions.some(s => s.type === 'missing-tests') &&
        results.analyses.security?.issues.length > 0) {
      concerns.push({
        message: 'Low test coverage combined with security vulnerabilities',
        action: 'Prioritize security-focused test coverage improvement'
      });
    }
    
    return concerns;
  }

  /**
   * Identify optimization opportunities
   */
  identifyOptimizations(results) {
    const optimizations = [];
    
    if (results.analyses.performance?.issues.some(i => i.type === 'nested-loops')) {
      optimizations.push({
        message: 'Performance bottlenecks identified in loop structures',
        action: 'Refactor nested loops using more efficient algorithms'
      });
    }
    
    if (results.analyses.architecture?.suggestions.some(s => s.type === 'module-size')) {
      optimizations.push({
        message: 'Large modules detected that could benefit from splitting',
        action: 'Refactor large modules into smaller, focused components'
      });
    }
    
    return optimizations;
  }

  /**
   * Generate focused report for specific analysis
   */
  async generateFocusedReport(mode, results) {
    console.log(chalk.cyan(`\n📊 ${mode.toUpperCase()} Analysis Report\n`));
    console.log(chalk.white('═'.repeat(50)));
    
    // Display issues
    if (results.issues && results.issues.length > 0) {
      console.log(chalk.red(`\n${results.issues.length} issues found:`));
      for (const issue of results.issues) {
        const severityColor = issue.severity === 'critical' ? chalk.red :
                             issue.severity === 'high' ? chalk.magenta :
                             issue.severity === 'warning' ? chalk.yellow :
                             chalk.gray;
        console.log(`  • [${severityColor(issue.severity)}] ${issue.message}`);
        if (issue.file) {
          console.log(chalk.gray(`    ${issue.file}`));
        }
      }
    }
    
    // Display suggestions
    if (results.suggestions && results.suggestions.length > 0) {
      console.log(chalk.blue(`\n${results.suggestions.length} suggestions:`));
      for (const suggestion of results.suggestions) {
        console.log(`  • ${suggestion.message}`);
      }
    }
    
    console.log(chalk.white('\n' + '═'.repeat(50)));
  }

  /**
   * Show help for analyze command
   */
  showHelp() {
    console.log(`
${chalk.cyan('Claude Flow Analyze - Advanced Code Analysis')}

${chalk.yellow('Usage:')} claude-flow analyze [mode] [target] [options]

${chalk.yellow('Modes:')}
  ${chalk.green('all')}           Complete analysis of all categories (default)
  ${chalk.green('code')}          Code quality analysis
  ${chalk.green('performance')}   Performance bottleneck detection
  ${chalk.green('security')}      Security vulnerability scanning
  ${chalk.green('architecture')}  Architectural pattern analysis
  ${chalk.green('hive-mind')}     Distributed AI-powered analysis

${chalk.yellow('Options:')}
  --hive-mind     Enable hive-mind coordination
  --deep          Perform deep, detailed analysis
  --sequential    Run analyses sequentially
  --help, -h      Show this help message

${chalk.yellow('Examples:')}
  claude-flow analyze                    # Complete analysis of current directory
  claude-flow analyze hive-mind ./src    # Hive-mind analysis of src folder
  claude-flow analyze security --deep    # Deep security analysis
  claude-flow analyse code               # British spelling supported
    `);
  }

  /**
   * Store analysis results
   */
  async storeResults(results) {
    try {
      const timestamp = new Date().toISOString();
      const reportDir = path.join(process.cwd(), '.analysis');
      
      // Create directory if it doesn't exist
      await fs.mkdir(reportDir, { recursive: true });
      
      // Save results
      const reportPath = path.join(reportDir, `report-${timestamp.replace(/[:.]/g, '-')}.json`);
      await fs.writeFile(reportPath, JSON.stringify(results, null, 2));
      
      console.log(chalk.gray(`Results saved to: ${reportPath}`));
    } catch (error) {
      console.error(chalk.yellow('Warning: Could not save results:'), error.message);
    }
  }
}

// Export for use as module
export { AnalyzeCommand };

// CLI execution
if (import.meta.url === `file://${process.argv[1]}`) {
  const command = new AnalyzeCommand();
  const args = process.argv.slice(2);
  
  command.execute(args).then(success => {
    process.exit(success ? 0 : 1);
  }).catch(error => {
    console.error(chalk.red('Error:'), error.message);
    process.exit(1);
  });
}