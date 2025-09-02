/**
 * Test Runner and Execution Plan
 * Coordinates comprehensive testing of OllamaMax platform
 */

import { execSync } from 'child_process';
import { writeFileSync } from 'fs';

class TestExecutor {
  constructor() {
    this.results = {
      total: 0,
      passed: 0,
      failed: 0,
      skipped: 0,
      tests: []
    };
    
    this.testSuites = [
      {
        name: 'API Health Tests',
        file: 'api-health-tests.js',
        priority: 1,
        description: 'Validates API server and worker health'
      },
      {
        name: 'UI Interaction Tests', 
        file: 'ui-interaction-tests.js',
        priority: 2,
        description: 'Tests all UI elements and interactions'
      },
      {
        name: 'P2P Model Migration Tests',
        file: 'p2p-model-migration-tests.js', 
        priority: 3,
        description: 'Tests P2P model migration functionality'
      },
      {
        name: 'Comprehensive Integration Tests',
        file: 'comprehensive-test-strategy.js',
        priority: 4,
        description: 'Complete end-to-end system validation'
      }
    ];
  }

  async runPreFlightChecks() {
    console.log('🔍 Running pre-flight checks...');
    
    const checks = [
      {
        name: 'Web Interface Server',
        command: 'curl -s http://localhost:8080 -o /dev/null',
        required: true
      },
      {
        name: 'API Server',
        command: 'curl -s http://localhost:13100/health -o /dev/null',
        required: true
      },
      {
        name: 'Worker Node 1',
        command: 'curl -s http://localhost:13000/api/version -o /dev/null',
        required: false
      },
      {
        name: 'Worker Node 2', 
        command: 'curl -s http://localhost:13001/api/version -o /dev/null',
        required: false
      },
      {
        name: 'Worker Node 3',
        command: 'curl -s http://localhost:13002/api/version -o /dev/null',
        required: false
      }
    ];

    let criticalFailures = 0;
    
    for (const check of checks) {
      try {
        execSync(check.command, { stdio: 'ignore' });
        console.log(`✅ ${check.name}: OK`);
      } catch (error) {
        if (check.required) {
          console.error(`❌ ${check.name}: FAILED (CRITICAL)`);
          criticalFailures++;
        } else {
          console.warn(`⚠️ ${check.name}: FAILED (NON-CRITICAL)`);
        }
      }
    }

    if (criticalFailures > 0) {
      console.error(`\n❌ ${criticalFailures} critical services are not running.`);
      console.error('Please ensure the web interface and API server are started before running tests.');
      return false;
    }

    console.log('\n✅ Pre-flight checks passed!\n');
    return true;
  }

  async runTestSuite(suite, browser = 'chromium') {
    console.log(`\n📋 Running ${suite.name}...`);
    console.log(`📄 Description: ${suite.description}`);
    
    const startTime = Date.now();
    
    try {
      const command = `npx playwright test tests/${suite.file} --project=${browser} --reporter=json`;
      console.log(`🚀 Executing: ${command}`);
      
      const output = execSync(command, { 
        encoding: 'utf8',
        stdio: ['inherit', 'pipe', 'pipe']
      });
      
      // Parse results if available
      try {
        const results = JSON.parse(output);
        const duration = Date.now() - startTime;
        
        const testResult = {
          suite: suite.name,
          status: 'passed',
          duration,
          tests: results.tests || [],
          stats: results.stats || {}
        };
        
        this.results.tests.push(testResult);
        this.results.passed++;
        
        console.log(`✅ ${suite.name} completed successfully (${duration}ms)`);
        return testResult;
        
      } catch (parseError) {
        // Fallback for non-JSON output
        const duration = Date.now() - startTime;
        const testResult = {
          suite: suite.name,
          status: 'passed',
          duration,
          output: output.substring(0, 500) + '...'
        };
        
        this.results.tests.push(testResult);
        this.results.passed++;
        
        console.log(`✅ ${suite.name} completed (${duration}ms)`);
        return testResult;
      }
      
    } catch (error) {
      const duration = Date.now() - startTime;
      const testResult = {
        suite: suite.name,
        status: 'failed',
        duration,
        error: error.message
      };
      
      this.results.tests.push(testResult);
      this.results.failed++;
      
      console.error(`❌ ${suite.name} failed (${duration}ms)`);
      console.error(`Error: ${error.message}`);
      
      return testResult;
    }
  }

  async runAllTests(browsers = ['chromium']) {
    console.log('🎯 OllamaMax Comprehensive Testing Suite');
    console.log('==========================================\n');
    
    // Pre-flight checks
    const preFlightPassed = await this.runPreFlightChecks();
    if (!preFlightPassed) {
      process.exit(1);
    }

    // Sort test suites by priority
    const sortedSuites = this.testSuites.sort((a, b) => a.priority - b.priority);
    
    for (const browser of browsers) {
      console.log(`\n🌐 Testing with ${browser} browser...\n`);
      
      for (const suite of sortedSuites) {
        this.results.total++;
        await this.runTestSuite(suite, browser);
        
        // Brief pause between test suites
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
    }

    await this.generateReport();
    this.displaySummary();
  }

  async generateReport() {
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        total: this.results.total,
        passed: this.results.passed,
        failed: this.results.failed,
        skipped: this.results.skipped,
        successRate: ((this.results.passed / this.results.total) * 100).toFixed(1)
      },
      testResults: this.results.tests,
      recommendations: this.generateRecommendations()
    };

    const reportPath = '/home/kp/ollamamax/test-results/comprehensive-test-report.json';
    
    try {
      writeFileSync(reportPath, JSON.stringify(report, null, 2));
      console.log(`\n📊 Test report saved to: ${reportPath}`);
    } catch (error) {
      console.warn(`⚠️ Could not save report: ${error.message}`);
    }

    return report;
  }

  generateRecommendations() {
    const recommendations = [];
    
    if (this.results.failed > 0) {
      recommendations.push('❌ Some tests failed - review error details and fix critical issues');
    }
    
    if (this.results.passed === this.results.total) {
      recommendations.push('✅ All tests passed - system is ready for production');
    }
    
    // Check for specific test patterns
    const failedTests = this.results.tests.filter(t => t.status === 'failed');
    const apiFailures = failedTests.filter(t => t.suite.includes('API'));
    const uiFailures = failedTests.filter(t => t.suite.includes('UI'));
    const p2pFailures = failedTests.filter(t => t.suite.includes('P2P'));
    
    if (apiFailures.length > 0) {
      recommendations.push('🔧 API issues detected - check server health and worker connectivity');
    }
    
    if (uiFailures.length > 0) {
      recommendations.push('🎨 UI issues detected - verify web interface is properly served');
    }
    
    if (p2pFailures.length > 0) {
      recommendations.push('🔄 P2P migration issues detected - verify model management functionality');
    }

    return recommendations;
  }

  displaySummary() {
    console.log('\n📊 TEST EXECUTION SUMMARY');
    console.log('========================');
    console.log(`Total Test Suites: ${this.results.total}`);
    console.log(`✅ Passed: ${this.results.passed}`);
    console.log(`❌ Failed: ${this.results.failed}`);
    console.log(`⏭️ Skipped: ${this.results.skipped}`);
    console.log(`🎯 Success Rate: ${((this.results.passed / this.results.total) * 100).toFixed(1)}%`);
    
    const totalDuration = this.results.tests.reduce((sum, test) => sum + (test.duration || 0), 0);
    console.log(`⏱️ Total Duration: ${(totalDuration / 1000).toFixed(2)}s`);
    
    console.log('\n📋 DETAILED RESULTS:');
    this.results.tests.forEach(test => {
      const icon = test.status === 'passed' ? '✅' : '❌';
      const duration = test.duration ? `(${test.duration}ms)` : '';
      console.log(`${icon} ${test.suite} ${duration}`);
    });

    const recommendations = this.generateRecommendations();
    if (recommendations.length > 0) {
      console.log('\n💡 RECOMMENDATIONS:');
      recommendations.forEach(rec => console.log(`  ${rec}`));
    }
    
    console.log('\n🏁 Testing completed!');
    
    return this.results.failed === 0;
  }
}

// Test execution scenarios
const scenarios = {
  quick: {
    browsers: ['chromium'],
    suites: ['api-health-tests.js', 'ui-interaction-tests.js']
  },
  
  p2p: {
    browsers: ['chromium'],  
    suites: ['p2p-model-migration-tests.js']
  },
  
  full: {
    browsers: ['chromium', 'firefox'],
    suites: 'all'
  },
  
  crossBrowser: {
    browsers: ['chromium', 'firefox', 'webkit'],
    suites: ['ui-interaction-tests.js']
  }
};

// Command-line execution
if (import.meta.url === `file://${process.argv[1]}`) {
  const scenario = process.argv[2] || 'quick';
  
  console.log(`🚀 Running ${scenario} test scenario...`);
  
  const executor = new TestExecutor();
  
  if (scenario === 'full' || scenario === 'all') {
    executor.runAllTests(['chromium']).then(success => {
      process.exit(success ? 0 : 1);
    });
  } else if (scenarios[scenario]) {
    const config = scenarios[scenario];
    const filteredSuites = config.suites === 'all' ? 
      executor.testSuites : 
      executor.testSuites.filter(suite => config.suites.includes(suite.file));
    
    executor.testSuites = filteredSuites;
    executor.runAllTests(config.browsers).then(success => {
      process.exit(success ? 0 : 1);
    });
  } else {
    console.error(`Unknown scenario: ${scenario}`);
    console.log('Available scenarios: quick, p2p, full, crossBrowser');
    process.exit(1);
  }
}

export { TestExecutor, scenarios };