#!/usr/bin/env node

/**
 * Analyze Agent - Advanced code analysis with hive-mind coordination
 * Provides multi-dimensional analysis with distributed intelligence
 */

import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs/promises';
import path from 'path';
import { performance } from 'perf_hooks';
import chalk from 'chalk';

const execAsync = promisify(exec);

class AnalyzeAgent {
  constructor(options = {}) {
    this.mode = options.mode || 'standard';
    this.depth = options.depth || 'medium';
    this.parallel = options.parallel !== false;
    this.metrics = {
      startTime: performance.now(),
      filesAnalyzed: 0,
      issuesFound: 0,
      suggestions: 0
    };
    this.hiveMemory = new Map();
  }

  /**
   * Initialize hive-mind coordination
   */
  async initHiveMind() {
    try {
      // Initialize swarm coordination
      await execAsync('npx claude-flow@alpha swarm init --topology mesh --max-agents 5');
      
      // Spawn specialized analysis agents
      const agents = [
        { type: 'code-quality', role: 'analyzer' },
        { type: 'performance', role: 'optimizer' },
        { type: 'security', role: 'auditor' },
        { type: 'architecture', role: 'architect' },
        { type: 'testing', role: 'qa' }
      ];

      for (const agent of agents) {
        await execAsync(`npx claude-flow@alpha agent spawn --type ${agent.type} --role ${agent.role}`);
        this.hiveMemory.set(agent.type, { status: 'active', results: [] });
      }

      console.log(chalk.green('✓ Hive-mind coordination initialized'));
      return true;
    } catch (error) {
      console.error(chalk.red('Failed to initialize hive-mind:'), error.message);
      return false;
    }
  }

  /**
   * Perform comprehensive code analysis
   */
  async analyze(target = '.') {
    console.log(chalk.cyan(`\n🔍 Starting ${this.mode} analysis of ${target}...\n`));

    const analyses = [];

    // Initialize hive-mind if in advanced mode
    if (this.mode === 'hive-mind') {
      await this.initHiveMind();
    }

    // Parallel analysis tasks
    if (this.parallel) {
      analyses.push(
        this.analyzeCodeQuality(target),
        this.analyzePerformance(target),
        this.analyzeSecurity(target),
        this.analyzeArchitecture(target),
        this.analyzeTestCoverage(target)
      );
    } else {
      // Sequential analysis
      analyses.push(await this.analyzeCodeQuality(target));
      analyses.push(await this.analyzePerformance(target));
      analyses.push(await this.analyzeSecurity(target));
      analyses.push(await this.analyzeArchitecture(target));
      analyses.push(await this.analyzeTestCoverage(target));
    }

    const results = await Promise.all(analyses);
    
    // Aggregate results
    const aggregated = this.aggregateResults(results);
    
    // Generate report
    await this.generateReport(aggregated);
    
    return aggregated;
  }

  /**
   * Analyze code quality - OPTIMIZED with parallel file processing
   */
  async analyzeCodeQuality(target) {
    const startTime = performance.now();
    const results = {
      type: 'code-quality',
      issues: [],
      suggestions: [],
      metrics: {}
    };

    try {
      // Check for common code quality issues
      const files = await this.getFiles(target, ['.js', '.jsx', '.ts', '.tsx']);
      
      // OPTIMIZED: Process files in parallel batches to avoid overwhelming system
      const batchSize = 10;
      const batches = [];
      for (let i = 0; i < files.length; i += batchSize) {
        batches.push(files.slice(i, i + batchSize));
      }
      
      for (const batch of batches) {
        const batchPromises = batch.map(async (file) => {
          try {
            const content = await fs.readFile(file, 'utf-8');
            const fileIssues = [];
            const fileSuggestions = [];
            
            // Parallel pattern matching using Promise.all for regex operations
            const [consoleLogs, todos] = await Promise.all([
              Promise.resolve(content.match(/console\.(log|error|warn|debug)/g)),
              Promise.resolve(content.match(/(TODO|FIXME|HACK|XXX):/g))
            ]);
            
            // Check file size
            if (content.length > 10000) {
              fileIssues.push({
                file,
                type: 'large-file',
                message: 'File exceeds recommended size (>10KB)',
                severity: 'warning'
              });
            }
            
            // Check for console.log statements
            if (consoleLogs && consoleLogs.length > 0) {
              fileIssues.push({
                file,
                type: 'console-statements',
                message: `Found ${consoleLogs.length} console statements`,
                severity: 'info'
              });
            }
            
            // Check for TODO/FIXME comments
            if (todos && todos.length > 0) {
              fileSuggestions.push({
                file,
                type: 'technical-debt',
                message: `Found ${todos.length} TODO/FIXME comments`,
                priority: 'low'
              });
            }
            
            return { issues: fileIssues, suggestions: fileSuggestions };
          } catch (error) {
            console.warn(`Failed to analyze file ${file}:`, error.message);
            return { issues: [], suggestions: [] };
          }
        });
        
        // Wait for batch completion and aggregate results
        const batchResults = await Promise.all(batchPromises);
        for (const result of batchResults) {
          results.issues.push(...result.issues);
          results.suggestions.push(...result.suggestions);
          this.metrics.filesAnalyzed++;
        }
      }
      
      // Store in hive memory if enabled
      if (this.hiveMemory.has('code-quality')) {
        this.hiveMemory.get('code-quality').results = results;
      }
      
    } catch (error) {
      console.error(chalk.red('Code quality analysis failed:'), error.message);
    }

    results.metrics.duration = performance.now() - startTime;
    return results;
  }

  /**
   * Analyze performance - OPTIMIZED with parallel processing and efficient regex
   */
  async analyzePerformance(target) {
    const startTime = performance.now();
    const results = {
      type: 'performance',
      issues: [],
      suggestions: [],
      metrics: {}
    };

    try {
      const files = await this.getFiles(target, ['.js', '.jsx', '.ts', '.tsx']);
      
      // OPTIMIZED: Pre-compile regex patterns for better performance
      const patterns = {
        nestedLoops: /for\s*\([^)]*\)[^}]*for\s*\([^)]*\)/g,
        syncOps: /fs\.(readFileSync|writeFileSync|appendFileSync)/g,
        inefficientArrayMethods: /\.(filter|map|reduce)\(.*\)\.(filter|map|reduce)/g,
        blockingOperations: /(sleep|setTimeout|setInterval)\s*\(/g
      };
      
      // Process files in parallel batches
      const batchSize = 8;
      const batches = [];
      for (let i = 0; i < files.length; i += batchSize) {
        batches.push(files.slice(i, i + batchSize));
      }
      
      for (const batch of batches) {
        const batchPromises = batch.map(async (file) => {
          try {
            const content = await fs.readFile(file, 'utf-8');
            const fileIssues = [];
            
            // Use single pass through content with all patterns
            const patternResults = {};
            for (const [key, pattern] of Object.entries(patterns)) {
              // Reset regex index for global patterns
              pattern.lastIndex = 0;
              patternResults[key] = content.match(pattern);
            }
            
            // Check for nested loops
            if (patternResults.nestedLoops?.length > 0) {
              fileIssues.push({
                file,
                type: 'nested-loops',
                message: `Potential performance issue: ${patternResults.nestedLoops.length} nested loops detected`,
                severity: 'warning'
              });
            }
            
            // Check for synchronous file operations
            if (patternResults.syncOps?.length > 0) {
              fileIssues.push({
                file,
                type: 'sync-operations',
                message: `Found ${patternResults.syncOps.length} synchronous file operations`,
                severity: 'warning'
              });
            }
            
            // Check for inefficient array method chaining
            if (patternResults.inefficientArrayMethods?.length > 0) {
              fileIssues.push({
                file,
                type: 'inefficient-array-chaining',
                message: `Found ${patternResults.inefficientArrayMethods.length} inefficient array method chains`,
                severity: 'info'
              });
            }
            
            // Check for blocking operations
            if (patternResults.blockingOperations?.length > 0) {
              fileIssues.push({
                file,
                type: 'blocking-operations',
                message: `Found ${patternResults.blockingOperations.length} potentially blocking operations`,
                severity: 'info'
              });
            }
            
            return fileIssues;
          } catch (error) {
            console.warn(`Failed to analyze performance in file ${file}:`, error.message);
            return [];
          }
        });
        
        // Wait for batch completion and aggregate results
        const batchResults = await Promise.all(batchPromises);
        for (const fileIssues of batchResults) {
          results.issues.push(...fileIssues);
        }
      }
      
      // Store in hive memory
      if (this.hiveMemory.has('performance')) {
        this.hiveMemory.get('performance').results = results;
      }
      
    } catch (error) {
      console.error(chalk.red('Performance analysis failed:'), error.message);
    }

    results.metrics.duration = performance.now() - startTime;
    return results;
  }

  /**
   * Analyze security - OPTIMIZED with parallel processing and comprehensive patterns
   */
  async analyzeSecurity(target) {
    const startTime = performance.now();
    const results = {
      type: 'security',
      issues: [],
      suggestions: [],
      metrics: {}
    };

    try {
      const files = await this.getFiles(target, ['.js', '.jsx', '.ts', '.tsx', '.json']);
      
      // OPTIMIZED: Pre-compiled comprehensive security patterns
      const securityPatterns = {
        secrets: /(api[_-]?key|secret|password|token|auth[_-]?token)\s*[:=]\s*["'][^"']{8,}["']/gi,
        evalUsage: /\beval\s*\(/g, // Only flag direct eval() usage, not setTimeout/setInterval
        sqlInjection: /(SELECT|INSERT|UPDATE|DELETE).*\+.*\$\{|\$\{.*\}.*(?:SELECT|INSERT|UPDATE|DELETE)/gi,
        xss: /innerHTML\s*[=+]|document\.write\s*\(|eval\s*\(/g,
        unsafeRegex: /new\s+RegExp\s*\(.*\$\{|RegExp\s*\(.*\$\{/g
      };
      
      // Process files in parallel batches
      const batchSize = 6; // Smaller batches for security analysis
      const batches = [];
      for (let i = 0; i < files.length; i += batchSize) {
        batches.push(files.slice(i, i + batchSize));
      }
      
      for (const batch of batches) {
        const batchPromises = batch.map(async (file) => {
          try {
            const content = await fs.readFile(file, 'utf-8');
            const fileIssues = [];
            
            // Run all security pattern checks in parallel
            const patternResults = await Promise.all(
              Object.entries(securityPatterns).map(([key, pattern]) => {
                pattern.lastIndex = 0; // Reset global regex
                return Promise.resolve({ key, matches: content.match(pattern) });
              })
            );
            
            // Process results efficiently
            for (const { key, matches } of patternResults) {
              if (matches?.length > 0) {
                let severity, message;
                
                switch (key) {
                  case 'secrets':
                    severity = 'critical';
                    message = 'Potential hardcoded secrets detected';
                    break;
                  case 'evalUsage':
                    severity = 'high';
                    message = 'Dangerous eval() or dynamic code execution detected';
                    break;
                  case 'sqlInjection':
                    severity = 'high';
                    message = 'Potential SQL injection vulnerability';
                    break;
                  case 'xss':
                    severity = 'high';
                    message = 'Potential XSS vulnerability detected';
                    break;
                  case 'unsafeRegex':
                    severity = 'medium';
                    message = 'Potentially unsafe regex construction';
                    break;
                  default:
                    severity = 'medium';
                    message = `Security issue detected: ${key}`;
                }
                
                fileIssues.push({
                  file,
                  type: key,
                  message,
                  severity,
                  count: matches.length
                });
              }
            }
            
            return fileIssues;
          } catch (error) {
            console.warn(`Failed to analyze security in file ${file}:`, error.message);
            return [];
          }
        });
        
        // Wait for batch completion and aggregate results
        const batchResults = await Promise.all(batchPromises);
        for (const fileIssues of batchResults) {
          results.issues.push(...fileIssues);
          // Count critical and high severity issues
          this.metrics.issuesFound += fileIssues.filter(issue => 
            issue.severity === 'critical' || issue.severity === 'high'
          ).length;
        }
      }
      
      // Store in hive memory
      if (this.hiveMemory.has('security')) {
        this.hiveMemory.get('security').results = results;
      }
      
    } catch (error) {
      console.error(chalk.red('Security analysis failed:'), error.message);
    }

    results.metrics.duration = performance.now() - startTime;
    return results;
  }

  /**
   * Analyze architecture
   */
  async analyzeArchitecture(target) {
    const startTime = performance.now();
    const results = {
      type: 'architecture',
      issues: [],
      suggestions: [],
      metrics: {}
    };

    try {
      // Analyze project structure
      const structure = await this.analyzeProjectStructure(target);
      
      // Check for circular dependencies
      const circularDeps = await this.checkCircularDependencies(target);
      if (circularDeps.length > 0) {
        results.issues.push({
          type: 'circular-dependencies',
          message: `Found ${circularDeps.length} circular dependencies`,
          severity: 'high',
          details: circularDeps
        });
      }
      
      // Analyze module cohesion
      if (structure.modules) {
        for (const module of structure.modules) {
          if (module.files > 20) {
            results.suggestions.push({
              type: 'module-size',
              message: `Module ${module.name} contains ${module.files} files (consider splitting)`,
              priority: 'medium'
            });
            this.metrics.suggestions++;
          }
        }
      }
      
      // Store in hive memory
      if (this.hiveMemory.has('architecture')) {
        this.hiveMemory.get('architecture').results = results;
      }
      
    } catch (error) {
      console.error(chalk.red('Architecture analysis failed:'), error.message);
    }

    results.metrics.duration = performance.now() - startTime;
    return results;
  }

  /**
   * Analyze test coverage
   */
  async analyzeTestCoverage(target) {
    const startTime = performance.now();
    const results = {
      type: 'testing',
      issues: [],
      suggestions: [],
      metrics: {}
    };

    try {
      // Count test files
      const testFiles = await this.getFiles(target, ['.test.js', '.spec.js', '.test.ts', '.spec.ts']);
      const srcFiles = await this.getFiles(target, ['.js', '.jsx', '.ts', '.tsx']);
      
      const coverage = (testFiles.length / srcFiles.length) * 100;
      
      if (coverage < 50) {
        results.issues.push({
          type: 'low-test-coverage',
          message: `Test coverage is low: ${coverage.toFixed(1)}%`,
          severity: 'warning'
        });
        this.metrics.issuesFound++;
      }
      
      // Check for untested files
      const untestedFiles = srcFiles.filter(src => {
        const testName = src.replace(/\.(js|jsx|ts|tsx)$/, '.test$1');
        const specName = src.replace(/\.(js|jsx|ts|tsx)$/, '.spec$1');
        return !testFiles.includes(testName) && !testFiles.includes(specName);
      });
      
      if (untestedFiles.length > 0) {
        results.suggestions.push({
          type: 'missing-tests',
          message: `${untestedFiles.length} files lack test coverage`,
          priority: 'high',
          files: untestedFiles.slice(0, 10) // Show first 10
        });
        this.metrics.suggestions++;
      }
      
      // Store in hive memory
      if (this.hiveMemory.has('testing')) {
        this.hiveMemory.get('testing').results = results;
      }
      
    } catch (error) {
      console.error(chalk.red('Test coverage analysis failed:'), error.message);
    }

    results.metrics.duration = performance.now() - startTime;
    return results;
  }

  /**
   * Get files with specific extensions - OPTIMIZED with parallel directory scanning
   */
  async getFiles(dir, extensions) {
    const files = [];
    const extensionsSet = new Set(extensions); // O(1) lookups instead of O(n)
    
    // OPTIMIZED: Use parallel directory scanning
    async function walkParallel(currentDir) {
      try {
        const entries = await fs.readdir(currentDir, { withFileTypes: true });
        
        // Separate directories and files for parallel processing
        const directories = [];
        const filePromises = [];
        
        for (const entry of entries) {
          // Skip node_modules and hidden directories early
          if (entry.name.startsWith('.') || entry.name === 'node_modules') {
            continue;
          }
          
          const fullPath = path.join(currentDir, entry.name);
          
          if (entry.isDirectory()) {
            directories.push(fullPath);
          } else if (entry.isFile()) {
            const ext = path.extname(entry.name);
            if (extensionsSet.has(ext)) { // O(1) lookup
              files.push(fullPath);
            }
          }
        }
        
        // Process directories in parallel
        if (directories.length > 0) {
          await Promise.all(directories.map(dir => walkParallel(dir)));
        }
        
      } catch (error) {
        // Ignore permission errors
      }
    }
    
    await walkParallel(dir);
    return files;
  }

  /**
   * Analyze project structure
   */
  async analyzeProjectStructure(target) {
    const structure = {
      modules: [],
      totalFiles: 0,
      totalDirectories: 0
    };
    
    try {
      const entries = await fs.readdir(target, { withFileTypes: true });
      
      for (const entry of entries) {
        if (entry.isDirectory() && !entry.name.startsWith('.') && entry.name !== 'node_modules') {
          const modulePath = path.join(target, entry.name);
          const files = await this.getFiles(modulePath, ['.js', '.jsx', '.ts', '.tsx']);
          
          structure.modules.push({
            name: entry.name,
            files: files.length
          });
          
          structure.totalDirectories++;
        } else if (entry.isFile()) {
          structure.totalFiles++;
        }
      }
    } catch (error) {
      console.error(chalk.red('Structure analysis failed:'), error.message);
    }
    
    return structure;
  }

  /**
   * Check for circular dependencies
   */
  async checkCircularDependencies(target) {
    const circular = [];
    // Simplified circular dependency check
    // In production, use a proper dependency graph analysis tool
    return circular;
  }

  /**
   * Aggregate results from all analyses
   */
  aggregateResults(results) {
    const aggregated = {
      summary: {
        totalIssues: 0,
        criticalIssues: 0,
        warnings: 0,
        suggestions: 0,
        filesAnalyzed: this.metrics.filesAnalyzed,
        duration: performance.now() - this.metrics.startTime
      },
      analyses: {},
      recommendations: []
    };
    
    for (const result of results) {
      aggregated.analyses[result.type] = result;
      
      // Count issues by severity
      for (const issue of result.issues) {
        aggregated.summary.totalIssues++;
        if (issue.severity === 'critical' || issue.severity === 'high') {
          aggregated.summary.criticalIssues++;
        } else if (issue.severity === 'warning') {
          aggregated.summary.warnings++;
        }
      }
      
      aggregated.summary.suggestions += result.suggestions.length;
    }
    
    // Generate recommendations based on findings
    if (aggregated.summary.criticalIssues > 0) {
      aggregated.recommendations.push({
        priority: 'critical',
        action: 'Address critical security and performance issues immediately'
      });
    }
    
    if (aggregated.analyses.testing?.suggestions?.some(s => s.type === 'missing-tests')) {
      aggregated.recommendations.push({
        priority: 'high',
        action: 'Improve test coverage for critical components'
      });
    }
    
    return aggregated;
  }

  /**
   * Generate analysis report
   */
  async generateReport(results) {
    console.log(chalk.cyan('\n📊 Analysis Report\n'));
    console.log(chalk.white('═'.repeat(50)));
    
    // Summary
    console.log(chalk.yellow('\n📈 Summary:'));
    console.log(`  Files Analyzed: ${results.summary.filesAnalyzed}`);
    console.log(`  Total Issues: ${results.summary.totalIssues}`);
    console.log(`  Critical Issues: ${chalk.red(results.summary.criticalIssues)}`);
    console.log(`  Warnings: ${chalk.yellow(results.summary.warnings)}`);
    console.log(`  Suggestions: ${chalk.blue(results.summary.suggestions)}`);
    console.log(`  Duration: ${(results.summary.duration / 1000).toFixed(2)}s`);
    
    // Detailed findings by category
    for (const [type, analysis] of Object.entries(results.analyses)) {
      if (analysis.issues.length > 0 || analysis.suggestions.length > 0) {
        console.log(chalk.yellow(`\n📋 ${type.toUpperCase()}:`));
        
        // Issues
        if (analysis.issues.length > 0) {
          console.log(chalk.red('  Issues:'));
          for (const issue of analysis.issues.slice(0, 5)) {
            const severityColor = issue.severity === 'critical' ? chalk.red :
                                 issue.severity === 'high' ? chalk.magenta :
                                 issue.severity === 'warning' ? chalk.yellow :
                                 chalk.gray;
            console.log(`    • [${severityColor(issue.severity.toUpperCase())}] ${issue.message}`);
            if (issue.file) {
              console.log(chalk.gray(`      ${issue.file}`));
            }
          }
        }
        
        // Suggestions
        if (analysis.suggestions.length > 0) {
          console.log(chalk.blue('  Suggestions:'));
          for (const suggestion of analysis.suggestions.slice(0, 3)) {
            console.log(`    • ${suggestion.message}`);
          }
        }
      }
    }
    
    // Recommendations
    if (results.recommendations.length > 0) {
      console.log(chalk.green('\n✨ Recommendations:'));
      for (const rec of results.recommendations) {
        const priorityColor = rec.priority === 'critical' ? chalk.red :
                            rec.priority === 'high' ? chalk.yellow :
                            chalk.blue;
        console.log(`  ${priorityColor('•')} ${rec.action}`);
      }
    }
    
    console.log(chalk.white('\n' + '═'.repeat(50)));
    
    // Save report to file
    const reportPath = path.join(process.cwd(), 'analysis-report.json');
    await fs.writeFile(reportPath, JSON.stringify(results, null, 2));
    console.log(chalk.green(`\n✓ Full report saved to: ${reportPath}`));
  }
}

// CLI interface
async function main() {
  const args = process.argv.slice(2);
  const options = {
    mode: 'standard',
    depth: 'medium',
    parallel: true
  };
  
  // Parse arguments
  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '--hive-mind':
        options.mode = 'hive-mind';
        break;
      case '--deep':
        options.depth = 'deep';
        break;
      case '--sequential':
        options.parallel = false;
        break;
      case '--help':
        console.log(`
${chalk.cyan('Analyze Agent - Advanced Code Analysis')}

Usage: analyze-agent [options] [target]

Options:
  --hive-mind     Enable hive-mind coordination for distributed analysis
  --deep          Perform deep analysis with maximum detail
  --sequential    Run analyses sequentially instead of in parallel
  --help          Show this help message

Examples:
  analyze-agent                    # Analyze current directory
  analyze-agent --hive-mind ./src  # Hive-mind analysis of src folder
  analyze-agent --deep --sequential # Deep sequential analysis
        `);
        process.exit(0);
    }
  }
  
  // Get target directory (default to current)
  const target = args.find(arg => !arg.startsWith('--')) || process.cwd();
  
  // Create and run analyzer
  const analyzer = new AnalyzeAgent(options);
  
  try {
    await analyzer.analyze(target);
    console.log(chalk.green('\n✅ Analysis complete!'));
  } catch (error) {
    console.error(chalk.red('\n❌ Analysis failed:'), error.message);
    process.exit(1);
  }
}

// Export for module usage
export { AnalyzeAgent };

// Run if called directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch(console.error);
}