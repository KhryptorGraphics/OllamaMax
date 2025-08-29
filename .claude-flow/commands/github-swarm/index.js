#!/usr/bin/env node

/**
 * GitHub Swarm Command
 * Specialized swarm for GitHub repository management
 */

const { spawn } = require('child_process');
const fs = require('fs').promises;
const path = require('path');
const { performance } = require('perf_hooks');

// Import GitHub-specific swarm implementation
const GitHubSwarmManager = require('./github-swarm-manager');

class GitHubSwarmCLI {
  constructor() {
    this.swarmManager = new GitHubSwarmManager();
    this.activeSwarms = new Map();
  }

  async parseArguments(args) {
    const options = {
      repository: null,
      agents: 5,
      focus: 'maintenance',
      autoPr: false,
      issueLabels: false,
      codeReview: false,
      help: false
    };

    for (let i = 0; i < args.length; i++) {
      const arg = args[i];
      
      switch (arg) {
        case '--repository':
        case '-r':
          options.repository = args[++i];
          break;
        case '--agents':
        case '-a':
          options.agents = parseInt(args[++i]) || 5;
          break;
        case '--focus':
        case '-f':
          options.focus = args[++i];
          break;
        case '--auto-pr':
          options.autoPr = true;
          break;
        case '--issue-labels':
          options.issueLabels = true;
          break;
        case '--code-review':
          options.codeReview = true;
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
🐙 GitHub Swarm - Specialized Repository Management

Usage:
  github-swarm [options]

Options:
  --repository, -r <owner/repo>  Target GitHub repository
  --agents, -a <number>          Number of specialized agents (default: 5)
  --focus, -f <type>             Focus area: maintenance, development, review, triage
  --auto-pr                      Enable automatic pull request enhancements
  --issue-labels                 Auto-categorize and label issues
  --code-review                  Enable AI-powered code reviews
  --help, -h                     Show this help message

Examples:
  github-swarm -r owner/repo
  github-swarm -r owner/repo -f maintenance --issue-labels
  github-swarm -r owner/repo -f development --auto-pr --code-review
  github-swarm -r owner/repo -a 8 -f triage --issue-labels --auto-pr

Agent Types:
  • Issue Triager    - Analyzes and categorizes issues
  • PR Reviewer      - Reviews code changes and suggests improvements
  • Documentation    - Updates README files and creates API docs
  • Test Agent       - Identifies missing tests and validates coverage
  • Security Agent   - Scans for vulnerabilities and security issues

Focus Areas:
  • maintenance      - Repository health, dependencies, documentation
  • development      - Code quality, testing, CI/CD improvements
  • review          - Pull request analysis and enhancement
  • triage          - Issue management and prioritization
    `);
  }

  async validateRepository(repository) {
    if (!repository) {
      throw new Error('Repository is required. Use --repository owner/repo');
    }

    const repoPattern = /^[a-zA-Z0-9_.-]+\/[a-zA-Z0-9_.-]+$/;
    if (!repoPattern.test(repository)) {
      throw new Error('Invalid repository format. Use owner/repo format');
    }

    return true;
  }

  async initializeSwarm(options) {
    console.log('🚀 Initializing GitHub Swarm...');
    console.log(`📁 Repository: ${options.repository}`);
    console.log(`🤖 Agents: ${options.agents}`);
    console.log(`🎯 Focus: ${options.focus}`);
    console.log(`⚙️  Features: ${this.getEnabledFeatures(options)}`);

    const swarmId = await this.swarmManager.createSwarm({
      repository: options.repository,
      maxAgents: options.agents,
      focus: options.focus,
      features: {
        autoPr: options.autoPr,
        issueLabels: options.issueLabels,
        codeReview: options.codeReview
      }
    });

    this.activeSwarms.set(swarmId, {
      repository: options.repository,
      startTime: Date.now(),
      options
    });

    return swarmId;
  }

  getEnabledFeatures(options) {
    const features = [];
    if (options.autoPr) features.push('Auto-PR');
    if (options.issueLabels) features.push('Issue Labels');
    if (options.codeReview) features.push('Code Review');
    return features.length > 0 ? features.join(', ') : 'Basic';
  }

  async executeSwarm(swarmId, options) {
    console.log('\n🔄 Executing GitHub Swarm...');
    
    try {
      const result = await this.swarmManager.executeSwarm(swarmId, {
        repository: options.repository,
        focus: options.focus,
        features: {
          autoPr: options.autoPr,
          issueLabels: options.issueLabels,
          codeReview: options.codeReview
        }
      });

      console.log('\n📊 Swarm Execution Results:');
      console.log(`✅ Success: ${result.success}`);
      console.log(`🤖 Agents Deployed: ${result.agentsDeployed}`);
      console.log(`📋 Tasks Completed: ${result.tasksCompleted}`);
      console.log(`⏱️  Execution Time: ${result.executionTime}ms`);

      if (result.summary) {
        console.log('\n📝 Summary:');
        result.summary.forEach(item => {
          console.log(`  • ${item}`);
        });
      }

      return result;
    } catch (error) {
      console.error('❌ Swarm execution failed:', error.message);
      throw error;
    }
  }

  async run(args) {
    try {
      const options = await this.parseArguments(args);

      if (options.help) {
        this.showHelp();
        return;
      }

      await this.validateRepository(options.repository);
      
      const swarmId = await this.initializeSwarm(options);
      const result = await this.executeSwarm(swarmId, options);

      console.log('\n🎉 GitHub Swarm completed successfully!');
      
      if (result.nextSteps) {
        console.log('\n🔮 Suggested Next Steps:');
        result.nextSteps.forEach(step => {
          console.log(`  • ${step}`);
        });
      }

    } catch (error) {
      console.error('❌ Error:', error.message);
      process.exit(1);
    }
  }
}

// CLI execution
if (require.main === module) {
  const cli = new GitHubSwarmCLI();
  const args = process.argv.slice(2);
  cli.run(args);
}

module.exports = GitHubSwarmCLI;
