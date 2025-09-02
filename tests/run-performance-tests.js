#!/usr/bin/env node

/**
 * OllamaMax Performance Test Runner
 * Orchestrates performance testing and generates comprehensive reports
 */

const { spawn } = require('child_process');
const fs = require('fs').promises;
const path = require('path');

class PerformanceTestRunner {
  constructor() {
    this.testSuites = [
      'performance-comprehensive.test.js',
      'performance-stress.test.js', 
      'performance-monitoring.test.js'
    ];
    
    this.results = {
      suites: [],
      metrics: {},
      timestamp: new Date().toISOString(),
      environment: {}
    };
  }
  
  async runAllTests() {
    console.log('🚀 Starting OllamaMax Performance Testing Suite...\n');
    
    // Collect environment info
    await this.collectEnvironmentInfo();
    
    // Start local server if needed
    const serverProcess = await this.startTestServer();
    
    try {
      // Run each test suite
      for (const suite of this.testSuites) {
        console.log(`\n📊 Running ${suite}...`);
        const result = await this.runTestSuite(suite);
        this.results.suites.push(result);
      }
      
      // Generate comprehensive report
      await this.generateReport();
      
    } finally {
      // Cleanup
      if (serverProcess) {
        serverProcess.kill();
      }
    }
  }
  
  async collectEnvironmentInfo() {
    const os = require('os');
    
    this.results.environment = {
      platform: os.platform(),
      arch: os.arch(),
      cpus: os.cpus().length,
      memory: Math.round(os.totalmem() / 1024 / 1024 / 1024), // GB
      nodeVersion: process.version,
      timestamp: Date.now()
    };
    
    console.log('Environment Info:');
    console.log(`  Platform: ${this.results.environment.platform} ${this.results.environment.arch}`);
    console.log(`  CPUs: ${this.results.environment.cpus}`);
    console.log(`  Memory: ${this.results.environment.memory}GB`);
    console.log(`  Node.js: ${this.results.environment.nodeVersion}`);
  }
  
  async startTestServer() {
    return new Promise((resolve) => {
      console.log('🌐 Starting test server...');
      
      const server = spawn('python3', ['-m', 'http.server', '8080'], {
        cwd: path.join(__dirname, '../web-interface'),
        stdio: 'pipe'
      });
      
      server.stdout.on('data', (data) => {
        if (data.toString().includes('Serving HTTP')) {
          console.log('✅ Test server started on port 8080');
          resolve(server);
        }
      });
      
      server.stderr.on('data', (data) => {
        console.error(`Server error: ${data}`);
      });
      
      // Fallback timeout
      setTimeout(() => resolve(server), 3000);
    });
  }
  
  async runTestSuite(suiteName) {
    return new Promise((resolve) => {
      const testPath = path.join(__dirname, suiteName);
      
      console.log(`  Running: ${suiteName}`);
      
      const startTime = Date.now();
      const playwright = spawn('npx', ['playwright', 'test', testPath, '--reporter=json'], {
        stdio: 'pipe'
      });
      
      let stdout = '';
      let stderr = '';
      
      playwright.stdout.on('data', (data) => {
        stdout += data.toString();
      });
      
      playwright.stderr.on('data', (data) => {
        stderr += data.toString();
      });
      
      playwright.on('close', (code) => {
        const endTime = Date.now();
        const duration = endTime - startTime;
        
        let testResults = {};
        try {
          // Parse Playwright JSON output
          testResults = JSON.parse(stdout);
        } catch (error) {
          console.warn(`  Warning: Could not parse test output for ${suiteName}`);
        }
        
        const result = {
          suite: suiteName,
          duration,
          exitCode: code,
          passed: code === 0,
          testResults,
          stdout: stdout.substring(0, 1000), // Truncate for storage
          stderr: stderr.substring(0, 1000)
        };
        
        console.log(`  ✅ Completed in ${duration}ms (exit code: ${code})`);
        resolve(result);
      });
    });
  }
  
  async generateReport() {
    console.log('\n📋 Generating performance report...');
    
    const report = {
      ...this.results,
      summary: this.generateSummary(),
      recommendations: this.generateRecommendations(),
      benchmarks: this.extractBenchmarks()
    };
    
    // Save JSON report
    const reportsDir = path.join(__dirname, '../test-results/performance');
    await fs.mkdir(reportsDir, { recursive: true });
    
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const jsonPath = path.join(reportsDir, `performance-report-${timestamp}.json`);
    
    await fs.writeFile(jsonPath, JSON.stringify(report, null, 2));
    console.log(`📄 JSON report saved: ${jsonPath}`);
    
    // Generate HTML report
    const htmlReport = this.generateHTMLReport(report);
    const htmlPath = path.join(reportsDir, `performance-report-${timestamp}.html`);
    await fs.writeFile(htmlPath, htmlReport);
    console.log(`📄 HTML report saved: ${htmlPath}`);
    
    // Generate markdown summary
    const markdownSummary = this.generateMarkdownSummary(report);
    const mdPath = path.join(reportsDir, `performance-summary-${timestamp}.md`);
    await fs.writeFile(mdPath, markdownSummary);
    console.log(`📄 Markdown summary saved: ${mdPath}`);
    
    // Display summary
    this.displaySummary(report);
  }
  
  generateSummary() {
    const totalTests = this.results.suites.reduce((sum, suite) => {
      return sum + (suite.testResults?.stats?.total || 0);
    }, 0);
    
    const passedTests = this.results.suites.reduce((sum, suite) => {
      return sum + (suite.testResults?.stats?.passed || 0);
    }, 0);
    
    const totalDuration = this.results.suites.reduce((sum, suite) => sum + suite.duration, 0);
    
    return {
      totalTests,
      passedTests,
      passRate: totalTests > 0 ? passedTests / totalTests : 0,
      totalDuration,
      suitesRun: this.results.suites.length,
      allPassed: this.results.suites.every(s => s.passed)
    };
  }
  
  generateRecommendations() {
    const recommendations = [];
    
    // Analyze test results for patterns
    const failedSuites = this.results.suites.filter(s => !s.passed);
    
    if (failedSuites.length > 0) {
      recommendations.push({
        category: 'Test Stability',
        severity: 'high',
        issue: `${failedSuites.length} test suite(s) failed`,
        action: 'Review failed tests and implement fixes',
        suites: failedSuites.map(s => s.suite)
      });
    }
    
    // Check for slow test execution
    const slowSuites = this.results.suites.filter(s => s.duration > 30000);
    if (slowSuites.length > 0) {
      recommendations.push({
        category: 'Test Performance',
        severity: 'medium',
        issue: 'Some test suites are running slowly',
        action: 'Optimize test execution or reduce scope',
        suites: slowSuites.map(s => ({ name: s.suite, duration: s.duration }))
      });
    }
    
    return recommendations;
  }
  
  extractBenchmarks() {
    // Extract performance metrics from test output
    const benchmarks = {
      pageLoadTime: { target: '<3000ms', measured: [] },
      webSocketLatency: { target: '<100ms', measured: [] },
      memoryUsage: { target: '<50MB growth', measured: [] },
      animationFPS: { target: '>50 FPS', measured: [] }
    };
    
    this.results.suites.forEach(suite => {
      if (suite.stdout) {
        // Extract metrics from console output
        const pageLoadMatches = suite.stdout.match(/Page load time: (\d+)ms/g);
        if (pageLoadMatches) {
          pageLoadMatches.forEach(match => {
            const time = parseInt(match.match(/(\d+)ms/)[1]);
            benchmarks.pageLoadTime.measured.push(time);
          });
        }
        
        const latencyMatches = suite.stdout.match(/latency: (\d+)ms/g);
        if (latencyMatches) {
          latencyMatches.forEach(match => {
            const latency = parseInt(match.match(/(\d+)ms/)[1]);
            benchmarks.webSocketLatency.measured.push(latency);
          });
        }
      }
    });
    
    return benchmarks;
  }
  
  generateHTMLReport(report) {
    return `
<!DOCTYPE html>
<html>
<head>
    <title>OllamaMax Performance Test Report</title>
    <meta charset="UTF-8">
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0; padding: 20px; background: #f8f9fa;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white; padding: 30px; border-radius: 12px; margin-bottom: 20px;
        }
        .metric-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .metric-card { 
            background: white; padding: 20px; border-radius: 8px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .metric-value { font-size: 2em; font-weight: bold; color: #667eea; }
        .status-good { color: #10b981; }
        .status-warning { color: #f59e0b; }
        .status-critical { color: #ef4444; }
        .recommendations { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 10px 0; }
        .test-suite { background: white; margin: 10px 0; padding: 15px; border-radius: 6px; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #e5e7eb; }
        th { background: #f9fafb; font-weight: 600; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 OllamaMax Performance Test Report</h1>
            <p>Generated: ${report.timestamp}</p>
            <p>Environment: ${report.environment.platform} (${report.environment.cpus} CPUs, ${report.environment.memory}GB RAM)</p>
        </div>
        
        <div class="metric-grid">
            <div class="metric-card">
                <h3>Test Summary</h3>
                <div class="metric-value ${report.summary.allPassed ? 'status-good' : 'status-critical'}">
                    ${report.summary.passedTests}/${report.summary.totalTests}
                </div>
                <p>Tests Passed (${(report.summary.passRate * 100).toFixed(1)}%)</p>
            </div>
            
            <div class="metric-card">
                <h3>Overall Score</h3>
                <div class="metric-value ${report.summary.overallScore > 80 ? 'status-good' : 
                                         report.summary.overallScore > 60 ? 'status-warning' : 'status-critical'}">
                    ${report.summary.overallScore}/100
                </div>
                <p>Performance Score</p>
            </div>
            
            <div class="metric-card">
                <h3>Test Duration</h3>
                <div class="metric-value">${(report.summary.totalDuration / 1000).toFixed(1)}s</div>
                <p>Total Execution Time</p>
            </div>
        </div>
        
        <h2>Performance Benchmarks</h2>
        <table>
            <thead>
                <tr><th>Metric</th><th>Target</th><th>Measured</th><th>Status</th></tr>
            </thead>
            <tbody>
                ${Object.entries(report.benchmarks).map(([key, benchmark]) => {
                  const hasMeasurement = benchmark.measured && benchmark.measured.length > 0;
                  const avgMeasured = hasMeasurement ? 
                    (benchmark.measured.reduce((a, b) => a + b, 0) / benchmark.measured.length).toFixed(1) : 'N/A';
                  const status = hasMeasured ? '✅' : '❌';
                  
                  return `
                    <tr>
                        <td><strong>${key}</strong></td>
                        <td>${benchmark.target}</td>
                        <td>${avgMeasured}${hasMeasurement ? ' (avg)' : ''}</td>
                        <td>${status}</td>
                    </tr>
                  `;
                }).join('')}
            </tbody>
        </table>
        
        <h2>Test Suite Results</h2>
        ${report.suites.map(suite => `
            <div class="test-suite">
                <h3>${suite.suite} ${suite.passed ? '✅' : '❌'}</h3>
                <p><strong>Duration:</strong> ${(suite.duration / 1000).toFixed(2)}s</p>
                <p><strong>Exit Code:</strong> ${suite.exitCode}</p>
                ${suite.stderr ? `<p><strong>Errors:</strong> ${suite.stderr}</p>` : ''}
            </div>
        `).join('')}
        
        ${report.recommendations.length > 0 ? `
            <h2>🔧 Recommendations</h2>
            ${report.recommendations.map(rec => `
                <div class="recommendations">
                    <h3>${rec.category} - ${rec.severity.toUpperCase()}</h3>
                    <p><strong>Issue:</strong> ${rec.issue}</p>
                    <p><strong>Action:</strong> ${rec.action}</p>
                </div>
            `).join('')}
        ` : '<h2>✅ No Issues Found</h2>'}
        
        <h2>📈 Performance Optimization Recommendations</h2>
        <div class="recommendations">
            <h3>Frontend Optimizations</h3>
            <ul>
                <li><strong>Bundle Optimization:</strong> Implement code splitting and tree shaking</li>
                <li><strong>Asset Optimization:</strong> Use WebP images and lazy loading</li>
                <li><strong>Caching Strategy:</strong> Implement service worker for offline capability</li>
                <li><strong>Virtual Scrolling:</strong> For large node/message lists</li>
            </ul>
        </div>
        
        <div class="recommendations">
            <h3>Backend Optimizations</h3>
            <ul>
                <li><strong>Redis Caching:</strong> Cache frequent API responses</li>
                <li><strong>WebSocket Optimization:</strong> Implement message batching</li>
                <li><strong>Load Balancing:</strong> Enhance node selection algorithms</li>
                <li><strong>Monitoring:</strong> Add real-time performance metrics</li>
            </ul>
        </div>
        
        <div class="recommendations">
            <h3>Infrastructure Optimizations</h3>
            <ul>
                <li><strong>CDN Integration:</strong> Serve static assets from CDN</li>
                <li><strong>HTTP/2:</strong> Enable HTTP/2 for multiplexing</li>
                <li><strong>Compression:</strong> Enable gzip/brotli compression</li>
                <li><strong>Health Checks:</strong> Implement advanced node health monitoring</li>
            </ul>
        </div>
    </div>
</body>
</html>`;
  }
  
  generateMarkdownSummary(report) {
    return `# OllamaMax Performance Test Summary

**Generated:** ${report.timestamp}
**Environment:** ${report.environment.platform} (${report.environment.cpus} CPUs, ${report.environment.memory}GB RAM)

## 📊 Results Overview

- **Overall Score:** ${report.summary.overallScore}/100
- **Tests Passed:** ${report.summary.passedTests}/${report.summary.totalTests} (${(report.summary.passRate * 100).toFixed(1)}%)
- **Total Duration:** ${(report.summary.totalDuration / 1000).toFixed(1)}s
- **Critical Issues:** ${report.summary.criticalIssues || 0}

## 🎯 Performance Benchmarks

| Metric | Target | Status |
|--------|--------|--------|
${Object.entries(report.benchmarks).map(([key, benchmark]) => {
  const status = benchmark.measured && benchmark.measured.length > 0 ? '✅ Measured' : '❌ No Data';
  return `| ${key} | ${benchmark.target} | ${status} |`;
}).join('\n')}

## 🚀 Optimization Priorities

### High Priority
1. **Page Load Optimization** - Target <3s load time
2. **WebSocket Latency** - Achieve <100ms real-time updates
3. **Memory Management** - Prevent memory leaks during extended use

### Medium Priority
1. **Animation Performance** - Maintain >50 FPS during interactions
2. **Network Optimization** - Implement request deduplication
3. **Error Handling** - Improve graceful degradation

### Low Priority
1. **Code Splitting** - Reduce initial bundle size
2. **Service Worker** - Add offline capability
3. **Performance Monitoring** - Real-time metrics dashboard

## 📈 Performance Trends

${report.suites.map(suite => `
### ${suite.suite}
- **Status:** ${suite.passed ? '✅ PASSED' : '❌ FAILED'}
- **Duration:** ${(suite.duration / 1000).toFixed(2)}s
`).join('')}

---
*Report generated by OllamaMax Performance Testing Suite*
`;
  }
  
  displaySummary(report) {
    console.log('\n' + '='.repeat(60));
    console.log('🎯 PERFORMANCE TEST SUMMARY');
    console.log('='.repeat(60));
    
    console.log(`\n📊 Overall Results:`);
    console.log(`   Score: ${report.summary.overallScore}/100`);
    console.log(`   Tests: ${report.summary.passedTests}/${report.summary.totalTests} passed`);
    console.log(`   Duration: ${(report.summary.totalDuration / 1000).toFixed(1)}s`);
    
    console.log(`\n🎯 Performance Status:`);
    Object.entries(report.benchmarks).forEach(([key, benchmark]) => {
      const status = benchmark.measured && benchmark.measured.length > 0 ? 
        '✅ Measured' : '❌ No Data';
      console.log(`   ${key}: ${benchmark.target} ${status}`);
    });
    
    if (report.recommendations.length > 0) {
      console.log(`\n⚠️  Issues Found:`);
      report.recommendations.forEach(rec => {
        const emoji = rec.severity === 'high' ? '🚨' : '⚠️';
        console.log(`   ${emoji} ${rec.category}: ${rec.issue}`);
      });
    } else {
      console.log(`\n✅ No performance issues detected!`);
    }
    
    console.log('\n' + '='.repeat(60));
  }
}

// CLI execution
if (require.main === module) {
  const runner = new PerformanceTestRunner();
  
  runner.runAllTests().catch(error => {
    console.error('Performance test runner failed:', error);
    process.exit(1);
  });
}

module.exports = { PerformanceTestRunner };