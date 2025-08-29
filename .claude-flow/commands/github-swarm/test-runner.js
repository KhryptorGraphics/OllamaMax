#!/usr/bin/env node

/**
 * GitHub Swarm Test Runner
 * Validates the GitHub swarm functionality
 */

const GitHubSwarmCLI = require('./index');
const GitHubSwarmManager = require('./github-swarm-manager');
const GitHubAPIIntegration = require('./github-api-integration');

class GitHubSwarmTestRunner {
  constructor() {
    this.tests = [];
    this.results = {
      passed: 0,
      failed: 0,
      total: 0
    };
  }

  async runAllTests() {
    console.log('🧪 Running GitHub Swarm Tests...\n');

    // Unit tests
    await this.testSwarmManager();
    await this.testCLIArgumentParsing();
    await this.testAgentSelection();
    await this.testFocusStrategies();

    // Integration tests (if GitHub token available)
    if (process.env.GITHUB_TOKEN) {
      await this.testGitHubAPIIntegration();
    } else {
      console.log('⚠️  Skipping GitHub API tests (no GITHUB_TOKEN)');
    }

    // Mock execution tests
    await this.testSwarmExecution();

    this.printResults();
    return this.results.failed === 0;
  }

  async testSwarmManager() {
    console.log('📋 Testing Swarm Manager...');

    await this.test('Create swarm with default config', async () => {
      const manager = new GitHubSwarmManager();
      const swarmId = await manager.createSwarm({
        repository: 'test/repo',
        maxAgents: 5,
        focus: 'maintenance'
      });

      if (!swarmId || typeof swarmId !== 'string') {
        throw new Error('Invalid swarm ID returned');
      }

      const status = await manager.getSwarmStatus(swarmId);
      if (status.repository !== 'test/repo') {
        throw new Error('Swarm config not stored correctly');
      }
    });

    await this.test('Create swarm with custom config', async () => {
      const manager = new GitHubSwarmManager();
      const swarmId = await manager.createSwarm({
        repository: 'owner/repo',
        maxAgents: 8,
        focus: 'development',
        features: { autoPr: true, codeReview: true }
      });

      const status = await manager.getSwarmStatus(swarmId);
      if (status.focus !== 'development') {
        throw new Error('Custom focus not applied');
      }
    });

    await this.test('List active swarms', async () => {
      const manager = new GitHubSwarmManager();
      await manager.createSwarm({ repository: 'test1/repo', focus: 'maintenance' });
      await manager.createSwarm({ repository: 'test2/repo', focus: 'development' });

      const swarms = await manager.listActiveSwarms();
      if (swarms.length < 2) {
        throw new Error('Not all swarms listed');
      }
    });
  }

  async testCLIArgumentParsing() {
    console.log('⚙️  Testing CLI Argument Parsing...');

    await this.test('Parse basic arguments', async () => {
      const cli = new GitHubSwarmCLI();
      const options = await cli.parseArguments(['-r', 'owner/repo', '-a', '6']);

      if (options.repository !== 'owner/repo' || options.agents !== 6) {
        throw new Error('Arguments not parsed correctly');
      }
    });

    await this.test('Parse feature flags', async () => {
      const cli = new GitHubSwarmCLI();
      const options = await cli.parseArguments([
        '-r', 'owner/repo', 
        '--auto-pr', 
        '--issue-labels', 
        '--code-review'
      ]);

      if (!options.autoPr || !options.issueLabels || !options.codeReview) {
        throw new Error('Feature flags not parsed correctly');
      }
    });

    await this.test('Validate repository format', async () => {
      const cli = new GitHubSwarmCLI();
      
      try {
        await cli.validateRepository('invalid-repo');
        throw new Error('Should have thrown validation error');
      } catch (error) {
        if (!error.message.includes('Invalid repository format')) {
          throw new Error('Wrong validation error message');
        }
      }
    });
  }

  async testAgentSelection() {
    console.log('🤖 Testing Agent Selection...');

    await this.test('Maintenance focus selects correct agents', async () => {
      const manager = new GitHubSwarmManager();
      const swarmId = await manager.createSwarm({
        repository: 'test/repo',
        focus: 'maintenance'
      });

      const status = await manager.getSwarmStatus(swarmId);
      const expectedAgents = ['issue-triager', 'documentation-agent', 'security-agent'];

      // Check if the selected agents match the expected ones for maintenance focus
      const selectedAgents = status.selectedAgents || [];
      const hasCorrectAgents = expectedAgents.every(agent => selectedAgents.includes(agent));

      if (!hasCorrectAgents) {
        console.log(`Expected: ${expectedAgents.join(', ')}`);
        console.log(`Got: ${selectedAgents.join(', ')}`);
        throw new Error('Incorrect agents selected for maintenance focus');
      }
    });

    await this.test('Development focus selects correct agents', async () => {
      const manager = new GitHubSwarmManager();
      const swarmId = await manager.createSwarm({
        repository: 'test/repo',
        focus: 'development'
      });

      const status = await manager.getSwarmStatus(swarmId);
      const expectedAgents = ['pr-reviewer', 'test-agent', 'documentation-agent'];

      // Check if the selected agents match the expected ones for development focus
      const selectedAgents = status.selectedAgents || [];
      const hasCorrectAgents = expectedAgents.every(agent => selectedAgents.includes(agent));

      if (!hasCorrectAgents) {
        console.log(`Expected: ${expectedAgents.join(', ')}`);
        console.log(`Got: ${selectedAgents.join(', ')}`);
        throw new Error('Incorrect agents selected for development focus');
      }
    });
  }

  async testFocusStrategies() {
    console.log('🎯 Testing Focus Strategies...');

    await this.test('All focus strategies are valid', async () => {
      const manager = new GitHubSwarmManager();
      const strategies = ['maintenance', 'development', 'review', 'triage'];

      for (const strategy of strategies) {
        const swarmId = await manager.createSwarm({
          repository: 'test/repo',
          focus: strategy
        });

        const status = await manager.getSwarmStatus(swarmId);
        if (status.focus !== strategy) {
          throw new Error(`Strategy ${strategy} not applied correctly`);
        }
      }
    });

    await this.test('Invalid focus strategy throws error', async () => {
      const manager = new GitHubSwarmManager();
      
      try {
        await manager.createSwarm({
          repository: 'test/repo',
          focus: 'invalid-strategy'
        });
        throw new Error('Should have thrown error for invalid strategy');
      } catch (error) {
        if (!error.message.includes('Unknown focus strategy')) {
          throw new Error('Wrong error message for invalid strategy');
        }
      }
    });
  }

  async testGitHubAPIIntegration() {
    console.log('🐙 Testing GitHub API Integration...');

    await this.test('GitHub API connection', async () => {
      const github = new GitHubAPIIntegration(process.env.GITHUB_TOKEN);
      const rateLimit = await github.getRateLimitStatus();

      if (!rateLimit.remaining || rateLimit.remaining < 0) {
        throw new Error('Invalid rate limit response');
      }
    });

    await this.test('Repository validation', async () => {
      const github = new GitHubAPIIntegration(process.env.GITHUB_TOKEN);
      const result = await github.validateRepository('octocat', 'Hello-World');

      if (!result.valid) {
        throw new Error('Failed to validate known public repository');
      }
    });
  }

  async testSwarmExecution() {
    console.log('🚀 Testing Swarm Execution...');

    await this.test('Execute swarm with mock data', async () => {
      const manager = new GitHubSwarmManager();

      // Enable mock mode for testing
      manager.mockMode = true;

      const swarmId = await manager.createSwarm({
        repository: 'test/repo',
        maxAgents: 3,
        focus: 'maintenance'
      });

      const result = await manager.executeSwarm(swarmId, {
        repository: 'test/repo',
        focus: 'maintenance',
        features: { issueLabels: true }
      });

      if (!result.success) {
        throw new Error('Swarm execution failed');
      }

      if (!result.summary || result.summary.length === 0) {
        throw new Error('No summary generated');
      }

      if (!result.nextSteps || result.nextSteps.length === 0) {
        throw new Error('No next steps generated');
      }

      // Verify that agents were deployed (even if some failed)
      if (result.agentsDeployed === 0) {
        throw new Error('No agents were deployed');
      }
    });
  }

  async test(name, testFunction) {
    this.results.total++;
    
    try {
      await testFunction();
      console.log(`  ✅ ${name}`);
      this.results.passed++;
    } catch (error) {
      console.log(`  ❌ ${name}: ${error.message}`);
      this.results.failed++;
    }
  }

  printResults() {
    console.log('\n📊 Test Results:');
    console.log(`  Total: ${this.results.total}`);
    console.log(`  Passed: ${this.results.passed}`);
    console.log(`  Failed: ${this.results.failed}`);
    console.log(`  Success Rate: ${Math.round((this.results.passed / this.results.total) * 100)}%`);

    if (this.results.failed === 0) {
      console.log('\n🎉 All tests passed!');
    } else {
      console.log('\n⚠️  Some tests failed. Check the output above.');
    }
  }
}

// Run tests if called directly
if (require.main === module) {
  const runner = new GitHubSwarmTestRunner();
  runner.runAllTests().then(success => {
    process.exit(success ? 0 : 1);
  });
}

module.exports = GitHubSwarmTestRunner;
