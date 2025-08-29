#!/usr/bin/env node

// Test Runner for OllamaMax Web Interface
// Comprehensive test execution with reporting and validation

const { execSync, spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

class TestRunner {
  constructor() {
    this.results = {
      unit: null,
      integration: null,
      services: null,
      components: null,
      coverage: null,
      lint: null
    };
    this.startTime = Date.now();
  }

  log(message, type = 'info') {
    const timestamp = new Date().toISOString();
    const colors = {
      info: '\x1b[36m',    // Cyan
      success: '\x1b[32m', // Green
      warning: '\x1b[33m', // Yellow
      error: '\x1b[31m',   // Red
      reset: '\x1b[0m'     // Reset
    };
    
    console.log(`${colors[type]}[${timestamp}] ${message}${colors.reset}`);
  }

  async runCommand(command, description) {
    this.log(`Running: ${description}`, 'info');
    
    try {
      const output = execSync(command, { 
        encoding: 'utf8',
        cwd: __dirname,
        stdio: 'pipe'
      });
      
      this.log(`✅ ${description} completed successfully`, 'success');
      return { success: true, output };
    } catch (error) {
      this.log(`❌ ${description} failed: ${error.message}`, 'error');
      return { success: false, error: error.message, output: error.stdout };
    }
  }

  async runTests(type, pattern = '') {
    const commands = {
      unit: 'npm run test:unit -- --verbose',
      integration: 'npm run test:integration -- --verbose',
      services: 'npm run test:services -- --verbose',
      components: 'npm run test:components -- --verbose',
      coverage: 'npm run test:coverage -- --verbose',
      all: 'npm test -- --verbose'
    };

    const command = commands[type] || commands.all;
    const result = await this.runCommand(command, `${type} tests`);
    this.results[type] = result;
    return result;
  }

  async runLinting() {
    const result = await this.runCommand('npm run lint', 'ESLint checks');
    this.results.lint = result;
    return result;
  }

  async checkDependencies() {
    this.log('Checking dependencies...', 'info');
    
    try {
      // Check if package.json exists
      if (!fs.existsSync(path.join(__dirname, 'package.json'))) {
        throw new Error('package.json not found');
      }

      // Check if node_modules exists
      if (!fs.existsSync(path.join(__dirname, 'node_modules'))) {
        this.log('Installing dependencies...', 'warning');
        await this.runCommand('npm install', 'dependency installation');
      }

      this.log('✅ Dependencies check completed', 'success');
      return true;
    } catch (error) {
      this.log(`❌ Dependencies check failed: ${error.message}`, 'error');
      return false;
    }
  }

  async validateTestFiles() {
    this.log('Validating test files...', 'info');
    
    const testFiles = [
      'src/tests/services/api.test.js',
      'src/tests/services/websocket.test.js',
      'src/tests/services/auth.test.js',
      'src/tests/components/EnhancedApp.test.js',
      'src/tests/integration/full-stack.test.js',
      'src/tests/setup.js'
    ];

    const missingFiles = [];
    
    for (const file of testFiles) {
      const filePath = path.join(__dirname, file);
      if (!fs.existsSync(filePath)) {
        missingFiles.push(file);
      }
    }

    if (missingFiles.length > 0) {
      this.log(`❌ Missing test files: ${missingFiles.join(', ')}`, 'error');
      return false;
    }

    this.log('✅ All test files found', 'success');
    return true;
  }

  async generateReport() {
    const endTime = Date.now();
    const duration = (endTime - this.startTime) / 1000;

    const report = {
      timestamp: new Date().toISOString(),
      duration: `${duration}s`,
      results: this.results,
      summary: {
        total: Object.keys(this.results).length,
        passed: Object.values(this.results).filter(r => r?.success).length,
        failed: Object.values(this.results).filter(r => r && !r.success).length,
        skipped: Object.values(this.results).filter(r => r === null).length
      }
    };

    // Write detailed report
    const reportPath = path.join(__dirname, 'test-report.json');
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));

    // Generate summary
    this.log('\n' + '='.repeat(60), 'info');
    this.log('TEST EXECUTION SUMMARY', 'info');
    this.log('='.repeat(60), 'info');
    this.log(`Total Duration: ${duration}s`, 'info');
    this.log(`Tests Passed: ${report.summary.passed}`, 'success');
    this.log(`Tests Failed: ${report.summary.failed}`, report.summary.failed > 0 ? 'error' : 'info');
    this.log(`Tests Skipped: ${report.summary.skipped}`, 'warning');
    this.log('='.repeat(60), 'info');

    // Detailed results
    for (const [type, result] of Object.entries(this.results)) {
      if (result === null) {
        this.log(`${type.toUpperCase()}: SKIPPED`, 'warning');
      } else if (result.success) {
        this.log(`${type.toUpperCase()}: PASSED ✅`, 'success');
      } else {
        this.log(`${type.toUpperCase()}: FAILED ❌`, 'error');
        if (result.error) {
          this.log(`  Error: ${result.error}`, 'error');
        }
      }
    }

    this.log(`\nDetailed report saved to: ${reportPath}`, 'info');
    return report;
  }

  async runServiceValidation() {
    this.log('Running service validation tests...', 'info');
    
    const validationTests = [
      {
        name: 'API Service Connection',
        test: async () => {
          // Test if API service can be imported and initialized
          try {
            const apiModule = require('./src/services/api.js');
            return apiModule.default ? true : false;
          } catch (error) {
            throw new Error(`API service import failed: ${error.message}`);
          }
        }
      },
      {
        name: 'WebSocket Service Connection',
        test: async () => {
          try {
            const wsModule = require('./src/services/websocket.js');
            return wsModule.default ? true : false;
          } catch (error) {
            throw new Error(`WebSocket service import failed: ${error.message}`);
          }
        }
      },
      {
        name: 'Auth Service Connection',
        test: async () => {
          try {
            const authModule = require('./src/services/auth.js');
            return authModule.default ? true : false;
          } catch (error) {
            throw new Error(`Auth service import failed: ${error.message}`);
          }
        }
      }
    ];

    let allPassed = true;
    
    for (const validation of validationTests) {
      try {
        await validation.test();
        this.log(`✅ ${validation.name}`, 'success');
      } catch (error) {
        this.log(`❌ ${validation.name}: ${error.message}`, 'error');
        allPassed = false;
      }
    }

    return allPassed;
  }

  async runAll() {
    this.log('🚀 Starting comprehensive test suite...', 'info');
    
    // Pre-flight checks
    const depsOk = await this.checkDependencies();
    if (!depsOk) {
      this.log('❌ Dependency check failed, aborting tests', 'error');
      return false;
    }

    const filesOk = await this.validateTestFiles();
    if (!filesOk) {
      this.log('❌ Test file validation failed, aborting tests', 'error');
      return false;
    }

    // Service validation
    const servicesOk = await this.runServiceValidation();
    if (!servicesOk) {
      this.log('⚠️  Service validation failed, continuing with tests...', 'warning');
    }

    // Run linting first
    await this.runLinting();

    // Run test suites
    await this.runTests('services');
    await this.runTests('components');
    await this.runTests('integration');
    await this.runTests('coverage');

    // Generate final report
    const report = await this.generateReport();
    
    // Return success status
    return report.summary.failed === 0;
  }

  async runSpecific(testType) {
    this.log(`🎯 Running specific test suite: ${testType}`, 'info');
    
    const depsOk = await this.checkDependencies();
    if (!depsOk) return false;

    const result = await this.runTests(testType);
    await this.generateReport();
    
    return result.success;
  }
}

// CLI Interface
async function main() {
  const args = process.argv.slice(2);
  const testRunner = new TestRunner();

  if (args.length === 0) {
    // Run all tests
    const success = await testRunner.runAll();
    process.exit(success ? 0 : 1);
  } else {
    const testType = args[0];
    const validTypes = ['unit', 'integration', 'services', 'components', 'coverage', 'lint'];
    
    if (!validTypes.includes(testType)) {
      console.error(`Invalid test type: ${testType}`);
      console.error(`Valid types: ${validTypes.join(', ')}`);
      process.exit(1);
    }

    const success = await testRunner.runSpecific(testType);
    process.exit(success ? 0 : 1);
  }
}

// Handle uncaught errors
process.on('uncaughtException', (error) => {
  console.error('Uncaught Exception:', error);
  process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

// Run if called directly
if (require.main === module) {
  main().catch(error => {
    console.error('Test runner failed:', error);
    process.exit(1);
  });
}

module.exports = TestRunner;
