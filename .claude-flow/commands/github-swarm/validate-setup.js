#!/usr/bin/env node

/**
 * GitHub Swarm Setup Validator
 * Validates the installation and configuration
 */

const fs = require('fs').promises;
const path = require('path');

class SetupValidator {
  constructor() {
    this.errors = [];
    this.warnings = [];
    this.info = [];
  }

  async validateSetup() {
    console.log('🔍 Validating GitHub Swarm Setup...\n');

    await this.checkNodeVersion();
    await this.checkRequiredFiles();
    await this.checkDependencies();
    await this.checkGitHubToken();
    await this.checkPermissions();

    this.printResults();
    return this.errors.length === 0;
  }

  async checkNodeVersion() {
    const nodeVersion = process.version;
    const majorVersion = parseInt(nodeVersion.slice(1).split('.')[0]);

    if (majorVersion >= 14) {
      this.info.push(`✅ Node.js version: ${nodeVersion}`);
    } else {
      this.errors.push(`❌ Node.js ${nodeVersion} is too old. Requires 14+`);
    }
  }

  async checkRequiredFiles() {
    const requiredFiles = [
      'index.js',
      'github-swarm-manager.js',
      'github-api-integration.js',
      'package.json',
      'README.md',
      'test-runner.js'
    ];

    const agentFiles = [
      'agents/issue-triager.js',
      'agents/pr-reviewer.js',
      'agents/documentation-agent.js',
      'agents/test-agent.js',
      'agents/security-agent.js'
    ];

    for (const file of requiredFiles) {
      try {
        await fs.access(file);
        this.info.push(`✅ Found: ${file}`);
      } catch (error) {
        this.errors.push(`❌ Missing: ${file}`);
      }
    }

    for (const file of agentFiles) {
      try {
        await fs.access(file);
        this.info.push(`✅ Found: ${file}`);
      } catch (error) {
        this.warnings.push(`⚠️  Missing: ${file} (agent will use simulation)`);
      }
    }
  }

  async checkDependencies() {
    try {
      const packageJson = JSON.parse(await fs.readFile('package.json', 'utf8'));
      const dependencies = packageJson.dependencies || {};

      const requiredDeps = ['@octokit/rest', 'commander', 'chalk', 'inquirer', 'ora'];
      
      for (const dep of requiredDeps) {
        if (dependencies[dep]) {
          this.info.push(`✅ Dependency: ${dep}@${dependencies[dep]}`);
        } else {
          this.errors.push(`❌ Missing dependency: ${dep}`);
        }
      }

      // Check if node_modules exists
      try {
        await fs.access('node_modules');
        this.info.push('✅ Dependencies installed');
      } catch (error) {
        this.errors.push('❌ Dependencies not installed. Run: npm install');
      }

    } catch (error) {
      this.errors.push('❌ Cannot read package.json');
    }
  }

  async checkGitHubToken() {
    const token = process.env.GITHUB_TOKEN;
    
    if (!token) {
      this.warnings.push('⚠️  No GITHUB_TOKEN environment variable found');
      this.warnings.push('   Some features will be limited without GitHub API access');
      this.warnings.push('   Set token with: export GITHUB_TOKEN="your_token"');
      return;
    }

    if (token.length < 20) {
      this.warnings.push('⚠️  GITHUB_TOKEN seems too short - verify it\'s correct');
      return;
    }

    // Test token validity
    try {
      const { Octokit } = require('@octokit/rest');
      const octokit = new Octokit({ auth: token });
      
      const { data } = await octokit.rest.users.getAuthenticated();
      this.info.push(`✅ GitHub token valid for user: ${data.login}`);
      
      // Check rate limit
      const rateLimit = await octokit.rest.rateLimit.get();
      const remaining = rateLimit.data.rate.remaining;
      
      if (remaining > 1000) {
        this.info.push(`✅ Rate limit: ${remaining} requests remaining`);
      } else {
        this.warnings.push(`⚠️  Rate limit low: ${remaining} requests remaining`);
      }

    } catch (error) {
      this.errors.push(`❌ GitHub token invalid: ${error.message}`);
    }
  }

  async checkPermissions() {
    try {
      // Check if main script is executable
      const stats = await fs.stat('index.js');
      const isExecutable = !!(stats.mode & parseInt('111', 8));
      
      if (isExecutable) {
        this.info.push('✅ Main script is executable');
      } else {
        this.warnings.push('⚠️  Main script not executable. Run: chmod +x index.js');
      }

    } catch (error) {
      this.errors.push('❌ Cannot check file permissions');
    }
  }

  printResults() {
    console.log('\n📊 Validation Results:');
    console.log('======================');

    if (this.info.length > 0) {
      console.log('\n✅ Success:');
      this.info.forEach(msg => console.log(`  ${msg}`));
    }

    if (this.warnings.length > 0) {
      console.log('\n⚠️  Warnings:');
      this.warnings.forEach(msg => console.log(`  ${msg}`));
    }

    if (this.errors.length > 0) {
      console.log('\n❌ Errors:');
      this.errors.forEach(msg => console.log(`  ${msg}`));
    }

    console.log('\n📈 Summary:');
    console.log(`  ✅ Success: ${this.info.length}`);
    console.log(`  ⚠️  Warnings: ${this.warnings.length}`);
    console.log(`  ❌ Errors: ${this.errors.length}`);

    if (this.errors.length === 0) {
      console.log('\n🎉 Setup validation passed!');
      console.log('\nNext steps:');
      console.log('  1. Test with: ./index.js --help');
      console.log('  2. Run tests: npm test');
      console.log('  3. Try a swarm: ./index.js -r owner/repo');
    } else {
      console.log('\n❌ Setup validation failed!');
      console.log('Please fix the errors above before using GitHub Swarm.');
    }
  }
}

// Run validation if called directly
if (require.main === module) {
  const validator = new SetupValidator();
  validator.validateSetup().then(success => {
    process.exit(success ? 0 : 1);
  });
}

module.exports = SetupValidator;
