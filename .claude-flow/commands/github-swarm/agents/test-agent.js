#!/usr/bin/env node

/**
 * Test Agent
 * Specialized agent for analyzing test coverage and quality
 */

const GitHubAPIIntegration = require('../github-api-integration');

class TestAgent {
  constructor(githubToken) {
    this.github = new GitHubAPIIntegration(githubToken);
    this.name = 'Test Agent';
    this.capabilities = ['test-analysis', 'coverage-check', 'test-generation', 'quality-assurance'];
  }

  async execute(config) {
    const { repository, features } = config;
    const [owner, repo] = repository.split('/');

    console.log(`🧪 ${this.name} analyzing repository: ${repository}`);

    const results = {
      agent: this.name,
      repository,
      tasksCompleted: 0,
      summary: '',
      actions: [],
      recommendations: []
    };

    try {
      // Analyze test structure
      const testStructure = await this.analyzeTestStructure(owner, repo);
      results.tasksCompleted++;

      // Check test coverage setup
      const coverageAnalysis = await this.analyzeCoverageSetup(owner, repo);
      if (coverageAnalysis.recommendations.length > 0) {
        results.recommendations.push(...coverageAnalysis.recommendations);
        results.tasksCompleted++;
      }

      // Analyze CI/CD test integration
      const ciAnalysis = await this.analyzeCITestIntegration(owner, repo);
      if (ciAnalysis.recommendations.length > 0) {
        results.recommendations.push(...ciAnalysis.recommendations);
        results.tasksCompleted++;
      }

      // Identify missing test patterns
      const missingTests = await this.identifyMissingTestPatterns(owner, repo);
      if (missingTests.length > 0) {
        results.actions.push({
          type: 'missing-test-patterns',
          patterns: missingTests
        });
        results.tasksCompleted++;
      }

      // Analyze test quality
      const qualityAnalysis = await this.analyzeTestQuality(testStructure);
      if (qualityAnalysis.issues.length > 0) {
        results.actions.push({
          type: 'test-quality-issues',
          issues: qualityAnalysis.issues
        });
        results.tasksCompleted++;
      }

      // Generate summary
      results.summary = this.generateSummary(testStructure, results.actions);

      console.log(`✅ ${this.name} completed analysis`);
      return results;

    } catch (error) {
      console.error(`❌ ${this.name} failed:`, error.message);
      throw error;
    }
  }

  async analyzeTestStructure(owner, repo) {
    const structure = {
      hasTests: false,
      testDirectories: [],
      testFiles: [],
      testFrameworks: [],
      estimatedCoverage: 0,
      testTypes: {
        unit: 0,
        integration: 0,
        e2e: 0
      }
    };

    // Common test directories and patterns
    const testPaths = [
      'test', 'tests', '__tests__', 'spec', 'specs',
      'src/test', 'src/tests', 'src/__tests__',
      'lib/test', 'lib/tests', 'lib/__tests__'
    ];

    // Check for test directories
    for (const testPath of testPaths) {
      try {
        const { data } = await this.github.octokit.rest.repos.getContent({
          owner,
          repo,
          path: testPath
        });

        if (Array.isArray(data)) {
          structure.hasTests = true;
          structure.testDirectories.push(testPath);
          
          // Count test files
          const testFiles = data.filter(file => 
            file.name.match(/\.(test|spec)\.(js|ts|py|rb|go|java|php)$/i)
          );
          structure.testFiles.push(...testFiles.map(f => `${testPath}/${f.name}`));
        }
      } catch (error) {
        // Directory doesn't exist, continue
      }
    }

    // Detect test frameworks from package.json
    try {
      const { data } = await this.github.octokit.rest.repos.getContent({
        owner,
        repo,
        path: 'package.json'
      });

      const packageJson = JSON.parse(Buffer.from(data.content, 'base64').toString());
      const dependencies = { ...packageJson.dependencies, ...packageJson.devDependencies };

      const frameworks = {
        'jest': 'Jest',
        'mocha': 'Mocha',
        'jasmine': 'Jasmine',
        'vitest': 'Vitest',
        'cypress': 'Cypress',
        'playwright': 'Playwright',
        'puppeteer': 'Puppeteer',
        'selenium': 'Selenium',
        'karma': 'Karma',
        'ava': 'AVA'
      };

      for (const [dep, name] of Object.entries(frameworks)) {
        if (dependencies[dep]) {
          structure.testFrameworks.push(name);
        }
      }
    } catch (error) {
      // package.json not found or not a Node.js project
    }

    // Categorize test types based on file names and paths
    structure.testFiles.forEach(file => {
      const fileName = file.toLowerCase();
      if (fileName.includes('unit') || fileName.includes('spec')) {
        structure.testTypes.unit++;
      } else if (fileName.includes('integration') || fileName.includes('int')) {
        structure.testTypes.integration++;
      } else if (fileName.includes('e2e') || fileName.includes('end-to-end')) {
        structure.testTypes.e2e++;
      } else {
        structure.testTypes.unit++; // Default to unit tests
      }
    });

    // Estimate coverage based on test file count and repository size
    if (structure.testFiles.length > 0) {
      const totalFiles = await this.estimateCodeFileCount(owner, repo);
      structure.estimatedCoverage = Math.min((structure.testFiles.length / totalFiles) * 100, 95);
    }

    return structure;
  }

  async estimateCodeFileCount(owner, repo) {
    try {
      // Get repository languages to estimate code files
      const { data } = await this.github.octokit.rest.repos.listLanguages({
        owner,
        repo
      });

      // Rough estimation based on repository size and languages
      const languages = Object.keys(data);
      const hasMainLanguage = languages.length > 0;
      
      // Very rough estimation - in a real implementation, you'd traverse the repo
      return hasMainLanguage ? Math.max(languages.length * 10, 20) : 20;
    } catch (error) {
      return 20; // Default estimate
    }
  }

  async analyzeCoverageSetup(owner, repo) {
    const recommendations = [];

    // Check for coverage configuration files
    const coverageFiles = [
      '.coveragerc',
      'coverage.xml',
      'lcov.info',
      '.nyc_output',
      'jest.config.js',
      'jest.config.json'
    ];

    let hasCoverageSetup = false;

    for (const file of coverageFiles) {
      try {
        await this.github.octokit.rest.repos.getContent({
          owner,
          repo,
          path: file
        });
        hasCoverageSetup = true;
        break;
      } catch (error) {
        // File doesn't exist
      }
    }

    if (!hasCoverageSetup) {
      recommendations.push({
        type: 'coverage-setup',
        priority: 'medium',
        message: 'No test coverage configuration found - consider adding coverage reporting'
      });
    }

    // Check for coverage badges in README
    try {
      const { data } = await this.github.octokit.rest.repos.getContent({
        owner,
        repo,
        path: 'README.md'
      });

      const content = Buffer.from(data.content, 'base64').toString();
      const hasCoverageBadge = content.includes('coverage') && content.includes('badge');

      if (!hasCoverageBadge) {
        recommendations.push({
          type: 'coverage-badge',
          priority: 'low',
          message: 'Consider adding a coverage badge to README'
        });
      }
    } catch (error) {
      // README not found
    }

    return { recommendations };
  }

  async analyzeCITestIntegration(owner, repo) {
    const recommendations = [];

    // Check for CI configuration files
    const ciFiles = [
      '.github/workflows',
      '.gitlab-ci.yml',
      '.travis.yml',
      'circle.yml',
      '.circleci/config.yml',
      'azure-pipelines.yml',
      'Jenkinsfile'
    ];

    let hasCISetup = false;
    let hasTestsInCI = false;

    for (const file of ciFiles) {
      try {
        const { data } = await this.github.octokit.rest.repos.getContent({
          owner,
          repo,
          path: file
        });

        hasCISetup = true;

        // Check if CI includes test commands
        if (Array.isArray(data)) {
          // GitHub Actions workflows directory
          for (const workflow of data) {
            try {
              const { data: workflowData } = await this.github.octokit.rest.repos.getContent({
                owner,
                repo,
                path: `${file}/${workflow.name}`
              });

              const content = Buffer.from(workflowData.content, 'base64').toString();
              if (content.includes('test') || content.includes('npm test') || content.includes('yarn test')) {
                hasTestsInCI = true;
                break;
              }
            } catch (error) {
              // Skip if can't read workflow file
            }
          }
        } else {
          // Single CI file
          const content = Buffer.from(data.content, 'base64').toString();
          if (content.includes('test') || content.includes('npm test') || content.includes('yarn test')) {
            hasTestsInCI = true;
          }
        }

        if (hasTestsInCI) break;
      } catch (error) {
        // File doesn't exist
      }
    }

    if (!hasCISetup) {
      recommendations.push({
        type: 'ci-setup',
        priority: 'high',
        message: 'No CI/CD configuration found - consider setting up automated testing'
      });
    } else if (!hasTestsInCI) {
      recommendations.push({
        type: 'ci-tests',
        priority: 'medium',
        message: 'CI/CD found but no test execution detected - add test commands to CI'
      });
    }

    return { recommendations };
  }

  async identifyMissingTestPatterns(owner, repo) {
    const missing = [];

    // Check for common test patterns
    const patterns = [
      {
        name: 'Unit Tests',
        description: 'Tests for individual functions/methods',
        check: async () => {
          // Look for unit test files
          const unitTestPatterns = ['unit', 'spec', '.test.'];
          // This would be more sophisticated in a real implementation
          return false; // Simplified for demo
        }
      },
      {
        name: 'Integration Tests',
        description: 'Tests for component interactions',
        check: async () => {
          // Look for integration test files
          return false; // Simplified for demo
        }
      },
      {
        name: 'End-to-End Tests',
        description: 'Full application workflow tests',
        check: async () => {
          // Look for e2e test files
          return false; // Simplified for demo
        }
      }
    ];

    for (const pattern of patterns) {
      const exists = await pattern.check();
      if (!exists) {
        missing.push({
          type: pattern.name,
          description: pattern.description,
          priority: pattern.name === 'Unit Tests' ? 'high' : 'medium'
        });
      }
    }

    return missing;
  }

  async analyzeTestQuality(testStructure) {
    const issues = [];

    // Check test coverage
    if (testStructure.estimatedCoverage < 50) {
      issues.push({
        type: 'low-coverage',
        message: `Estimated test coverage is low (${Math.round(testStructure.estimatedCoverage)}%)`,
        severity: 'high'
      });
    } else if (testStructure.estimatedCoverage < 80) {
      issues.push({
        type: 'medium-coverage',
        message: `Test coverage could be improved (${Math.round(testStructure.estimatedCoverage)}%)`,
        severity: 'medium'
      });
    }

    // Check test distribution
    const totalTests = testStructure.testTypes.unit + testStructure.testTypes.integration + testStructure.testTypes.e2e;
    if (totalTests > 0) {
      const unitPercentage = (testStructure.testTypes.unit / totalTests) * 100;
      
      if (unitPercentage < 60) {
        issues.push({
          type: 'test-distribution',
          message: 'Consider adding more unit tests for better test pyramid',
          severity: 'medium'
        });
      }

      if (testStructure.testTypes.integration === 0 && totalTests > 5) {
        issues.push({
          type: 'missing-integration-tests',
          message: 'No integration tests found - consider adding some',
          severity: 'medium'
        });
      }
    }

    // Check for test frameworks
    if (testStructure.hasTests && testStructure.testFrameworks.length === 0) {
      issues.push({
        type: 'no-test-framework',
        message: 'Test files found but no recognized test framework detected',
        severity: 'medium'
      });
    }

    return { issues };
  }

  generateSummary(testStructure, actions) {
    const issueCount = actions.reduce((sum, action) => sum + (action.issues?.length || action.patterns?.length || 1), 0);
    
    return `Found ${testStructure.testFiles.length} test files, ` +
           `estimated ${Math.round(testStructure.estimatedCoverage)}% coverage. ` +
           `Frameworks: ${testStructure.testFrameworks.join(', ') || 'None detected'}. ` +
           `Identified ${issueCount} improvement areas.`;
  }
}

module.exports = TestAgent;
