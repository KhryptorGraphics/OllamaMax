/**
 * OllamaMax Performance Analysis and Reporting Tool
 * Generates detailed performance analysis reports and recommendations
 */

const fs = require('fs');
const path = require('path');

class PerformanceAnalyzer {
  constructor() {
    this.metrics = {
      pageLoad: [],
      webSocketLatency: [],
      memoryUsage: [],
      networkRequests: [],
      animationFrameRate: [],
      failoverTimes: []
    };
    
    this.thresholds = {
      pageLoadTime: 3000,
      webSocketLatency: 100,
      memoryGrowthLimit: 50,
      animationFrameRate: 50,
      failoverTime: 2000
    };
    
    this.recommendations = [];
  }
  
  addMetric(category, value, metadata = {}) {
    if (this.metrics[category]) {
      this.metrics[category].push({
        value,
        timestamp: Date.now(),
        metadata
      });
    }
  }
  
  analyzePageLoadPerformance() {
    const loadTimes = this.metrics.pageLoad.map(m => m.value);
    
    if (loadTimes.length === 0) return null;
    
    const analysis = {
      average: loadTimes.reduce((a, b) => a + b, 0) / loadTimes.length,
      median: this.calculateMedian(loadTimes),
      p95: this.calculatePercentile(loadTimes, 95),
      min: Math.min(...loadTimes),
      max: Math.max(...loadTimes),
      passRate: loadTimes.filter(t => t < this.thresholds.pageLoadTime).length / loadTimes.length
    };
    
    // Generate recommendations
    if (analysis.average > this.thresholds.pageLoadTime) {
      this.recommendations.push({
        category: 'Page Load',
        severity: 'high',
        issue: `Average load time (${analysis.average.toFixed(0)}ms) exceeds threshold (${this.thresholds.pageLoadTime}ms)`,
        recommendations: [
          'Implement resource bundling and minification',
          'Add lazy loading for non-critical components',
          'Optimize image assets and use modern formats (WebP)',
          'Implement service worker caching strategy',
          'Consider code splitting for larger modules'
        ]
      });
    }
    
    if (analysis.p95 > this.thresholds.pageLoadTime * 1.5) {
      this.recommendations.push({
        category: 'Page Load',
        severity: 'medium',
        issue: `95th percentile load time indicates inconsistent performance`,
        recommendations: [
          'Investigate network variability',
          'Add performance monitoring for slow devices',
          'Implement progressive loading strategies'
        ]
      });
    }
    
    return analysis;
  }
  
  analyzeWebSocketPerformance() {
    const latencies = this.metrics.webSocketLatency.map(m => m.value);
    
    if (latencies.length === 0) return null;
    
    const analysis = {
      average: latencies.reduce((a, b) => a + b, 0) / latencies.length,
      median: this.calculateMedian(latencies),
      p95: this.calculatePercentile(latencies, 95),
      passRate: latencies.filter(l => l < this.thresholds.webSocketLatency).length / latencies.length
    };
    
    if (analysis.average > this.thresholds.webSocketLatency) {
      this.recommendations.push({
        category: 'WebSocket',
        severity: 'high',
        issue: `WebSocket latency (${analysis.average.toFixed(0)}ms) exceeds real-time threshold`,
        recommendations: [
          'Optimize WebSocket message serialization',
          'Implement message batching for non-critical updates',
          'Consider WebSocket compression (permessage-deflate)',
          'Review network infrastructure and routing',
          'Implement connection pooling for multiple workers'
        ]
      });
    }
    
    return analysis;
  }
  
  analyzeMemoryUsage() {
    const memoryData = this.metrics.memoryUsage.map(m => m.value);
    
    if (memoryData.length < 2) return null;
    
    const growth = memoryData[memoryData.length - 1] - memoryData[0];
    const peakUsage = Math.max(...memoryData);
    
    const analysis = {
      growth: growth / 1024 / 1024, // Convert to MB
      peak: peakUsage / 1024 / 1024,
      trend: this.calculateTrend(memoryData),
      leakIndicators: this.detectMemoryLeaks(memoryData)
    };
    
    if (analysis.growth > this.thresholds.memoryGrowthLimit) {
      this.recommendations.push({
        category: 'Memory',
        severity: 'high',
        issue: `Memory growth (${analysis.growth.toFixed(2)}MB) indicates potential leaks`,
        recommendations: [
          'Implement proper cleanup of event listeners',
          'Review DOM node management and removal',
          'Add WeakMap/WeakSet for object references',
          'Implement virtual scrolling for large lists',
          'Review timer and interval cleanup'
        ]
      });
    }
    
    if (analysis.leakIndicators.suspiciousGrowth) {
      this.recommendations.push({
        category: 'Memory',
        severity: 'medium',
        issue: 'Continuous memory growth pattern detected',
        recommendations: [
          'Add memory profiling to identify leak sources',
          'Implement automatic garbage collection triggers',
          'Review WebSocket message handling for retention'
        ]
      });
    }
    
    return analysis;
  }
  
  analyzeNetworkPerformance() {
    const requests = this.metrics.networkRequests;
    
    if (requests.length === 0) return null;
    
    const responseTimes = requests.map(r => r.responseTime).filter(t => t > 0);
    const requestCounts = this.groupRequestsByEndpoint(requests);
    
    const analysis = {
      totalRequests: requests.length,
      averageResponseTime: responseTimes.reduce((a, b) => a + b, 0) / responseTimes.length || 0,
      slowestRequests: responseTimes.filter(t => t > 1000).length,
      requestPatterns: requestCounts,
      cacheHitRate: this.calculateCacheHitRate(requests)
    };
    
    if (analysis.averageResponseTime > 1000) {
      this.recommendations.push({
        category: 'Network',
        severity: 'high',
        issue: `Average API response time (${analysis.averageResponseTime.toFixed(0)}ms) is slow`,
        recommendations: [
          'Implement API response caching',
          'Add request deduplication logic',
          'Optimize database queries',
          'Consider CDN for static assets',
          'Implement request prioritization'
        ]
      });
    }
    
    // Check for excessive duplicate requests
    const duplicateRate = this.calculateDuplicateRequestRate(requests);
    if (duplicateRate > 0.3) {
      this.recommendations.push({
        category: 'Network',
        severity: 'medium',
        issue: `High duplicate request rate (${(duplicateRate * 100).toFixed(1)}%)`,
        recommendations: [
          'Implement request deduplication',
          'Add intelligent caching layer',
          'Review component lifecycle management'
        ]
      });
    }
    
    return analysis;
  }
  
  generateComprehensiveReport() {
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        overallScore: this.calculateOverallScore(),
        criticalIssues: this.recommendations.filter(r => r.severity === 'high').length,
        warnings: this.recommendations.filter(r => r.severity === 'medium').length
      },
      analysis: {
        pageLoad: this.analyzePageLoadPerformance(),
        webSocket: this.analyzeWebSocketPerformance(),
        memory: this.analyzeMemoryUsage(),
        network: this.analyzeNetworkPerformance()
      },
      recommendations: this.recommendations,
      benchmarks: this.generateBenchmarks()
    };
    
    return report;
  }
  
  generateBenchmarks() {
    return {
      pageLoadTime: {
        target: '< 3000ms',
        current: this.metrics.pageLoad.length > 0 ? 
          `${(this.metrics.pageLoad.reduce((a, b) => a + b.value, 0) / this.metrics.pageLoad.length).toFixed(0)}ms` : 'N/A'
      },
      webSocketLatency: {
        target: '< 100ms',
        current: this.metrics.webSocketLatency.length > 0 ?
          `${(this.metrics.webSocketLatency.reduce((a, b) => a + b.value, 0) / this.metrics.webSocketLatency.length).toFixed(0)}ms` : 'N/A'
      },
      memoryEfficiency: {
        target: '< 50MB growth',
        current: this.metrics.memoryUsage.length > 1 ?
          `${((this.metrics.memoryUsage[this.metrics.memoryUsage.length - 1].value - this.metrics.memoryUsage[0].value) / 1024 / 1024).toFixed(2)}MB` : 'N/A'
      },
      animationPerformance: {
        target: '> 50 FPS',
        current: this.metrics.animationFrameRate.length > 0 ?
          `${(this.metrics.animationFrameRate.reduce((a, b) => a + b.value, 0) / this.metrics.animationFrameRate.length).toFixed(1)} FPS` : 'N/A'
      }
    };
  }
  
  calculateOverallScore() {
    let score = 100;
    
    // Deduct points for each critical issue
    const criticalIssues = this.recommendations.filter(r => r.severity === 'high').length;
    const warnings = this.recommendations.filter(r => r.severity === 'medium').length;
    
    score -= criticalIssues * 20;
    score -= warnings * 10;
    
    return Math.max(score, 0);
  }
  
  // Utility methods
  calculateMedian(values) {
    const sorted = [...values].sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    return sorted.length % 2 !== 0 ? sorted[mid] : (sorted[mid - 1] + sorted[mid]) / 2;
  }
  
  calculatePercentile(values, percentile) {
    const sorted = [...values].sort((a, b) => a - b);
    const index = Math.ceil((percentile / 100) * sorted.length) - 1;
    return sorted[index];
  }
  
  calculateTrend(values) {
    if (values.length < 2) return 'insufficient_data';
    
    const firstHalf = values.slice(0, Math.floor(values.length / 2));
    const secondHalf = values.slice(Math.floor(values.length / 2));
    
    const firstAvg = firstHalf.reduce((a, b) => a + b, 0) / firstHalf.length;
    const secondAvg = secondHalf.reduce((a, b) => a + b, 0) / secondHalf.length;
    
    const change = (secondAvg - firstAvg) / firstAvg;
    
    if (change > 0.1) return 'increasing';
    if (change < -0.1) return 'decreasing';
    return 'stable';
  }
  
  detectMemoryLeaks(memoryData) {
    if (memoryData.length < 3) return { suspiciousGrowth: false };
    
    // Check for continuous growth pattern
    let growthCount = 0;
    for (let i = 1; i < memoryData.length; i++) {
      if (memoryData[i] > memoryData[i - 1]) {
        growthCount++;
      }
    }
    
    const growthRate = growthCount / (memoryData.length - 1);
    
    return {
      suspiciousGrowth: growthRate > 0.7, // More than 70% of samples show growth
      growthRate: growthRate
    };
  }
  
  groupRequestsByEndpoint(requests) {
    const groups = {};
    
    requests.forEach(req => {
      const endpoint = req.url.split('?')[0]; // Remove query params
      if (!groups[endpoint]) {
        groups[endpoint] = {
          count: 0,
          totalTime: 0,
          errors: 0
        };
      }
      
      groups[endpoint].count++;
      if (req.responseTime) {
        groups[endpoint].totalTime += req.responseTime;
      }
      if (req.error) {
        groups[endpoint].errors++;
      }
    });
    
    // Calculate averages
    Object.keys(groups).forEach(endpoint => {
      const group = groups[endpoint];
      group.averageTime = group.count > 0 ? group.totalTime / group.count : 0;
      group.errorRate = group.count > 0 ? group.errors / group.count : 0;
    });
    
    return groups;
  }
  
  calculateCacheHitRate(requests) {
    // Simplified cache hit detection based on response times
    const fastRequests = requests.filter(r => r.responseTime < 50).length;
    return requests.length > 0 ? fastRequests / requests.length : 0;
  }
  
  calculateDuplicateRequestRate(requests) {
    const uniqueRequests = new Set(requests.map(r => `${r.method}:${r.url}`));
    return requests.length > 0 ? 1 - (uniqueRequests.size / requests.length) : 0;
  }
  
  exportReport(format = 'json') {
    const report = this.generateComprehensiveReport();
    
    switch (format) {
      case 'json':
        return JSON.stringify(report, null, 2);
      
      case 'html':
        return this.generateHTMLReport(report);
      
      case 'markdown':
        return this.generateMarkdownReport(report);
      
      default:
        return report;
    }
  }
  
  generateHTMLReport(report) {
    return `
<!DOCTYPE html>
<html>
<head>
    <title>OllamaMax Performance Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 8px; }
        .metric { margin: 10px 0; padding: 10px; border-left: 4px solid #007acc; }
        .critical { border-left-color: #d73a49; }
        .warning { border-left-color: #f66a0a; }
        .good { border-left-color: #28a745; }
        .recommendations { background: #fff3cd; padding: 15px; border-radius: 5px; margin: 10px 0; }
        table { border-collapse: collapse; width: 100%; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>OllamaMax Performance Analysis Report</h1>
        <p>Generated: ${report.timestamp}</p>
        <p>Overall Performance Score: <strong>${report.summary.overallScore}/100</strong></p>
    </div>
    
    <h2>Performance Metrics</h2>
    
    ${report.analysis.pageLoad ? `
    <div class="metric ${report.analysis.pageLoad.average < 3000 ? 'good' : 'critical'}">
        <h3>Page Load Performance</h3>
        <p>Average: ${report.analysis.pageLoad.average.toFixed(0)}ms</p>
        <p>95th Percentile: ${report.analysis.pageLoad.p95.toFixed(0)}ms</p>
        <p>Pass Rate: ${(report.analysis.pageLoad.passRate * 100).toFixed(1)}%</p>
    </div>` : ''}
    
    ${report.analysis.webSocket ? `
    <div class="metric ${report.analysis.webSocket.average < 100 ? 'good' : 'warning'}">
        <h3>WebSocket Performance</h3>
        <p>Average Latency: ${report.analysis.webSocket.average.toFixed(0)}ms</p>
        <p>Pass Rate: ${(report.analysis.webSocket.passRate * 100).toFixed(1)}%</p>
    </div>` : ''}
    
    <h2>Performance Benchmarks</h2>
    <table>
        <tr><th>Metric</th><th>Target</th><th>Current</th><th>Status</th></tr>
        ${Object.entries(report.benchmarks).map(([key, benchmark]) => `
        <tr>
            <td>${key}</td>
            <td>${benchmark.target}</td>
            <td>${benchmark.current}</td>
            <td>${benchmark.current.includes('N/A') ? 'No Data' : '✓'}</td>
        </tr>
        `).join('')}
    </table>
    
    <h2>Recommendations</h2>
    ${report.recommendations.map(rec => `
    <div class="recommendations ${rec.severity}">
        <h3>${rec.category} - ${rec.severity.toUpperCase()}</h3>
        <p><strong>Issue:</strong> ${rec.issue}</p>
        <p><strong>Recommendations:</strong></p>
        <ul>
            ${rec.recommendations.map(r => `<li>${r}</li>`).join('')}
        </ul>
    </div>
    `).join('')}
    
</body>
</html>`;
  }
  
  generateMarkdownReport(report) {
    let markdown = `# OllamaMax Performance Analysis Report

Generated: ${report.timestamp}
Overall Performance Score: **${report.summary.overallScore}/100**

## Performance Summary

- Critical Issues: ${report.summary.criticalIssues}
- Warnings: ${report.summary.warnings}

## Benchmarks

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
`;
    
    Object.entries(report.benchmarks).forEach(([key, benchmark]) => {
      const status = benchmark.current.includes('N/A') ? '❌ No Data' : '✅ Measured';
      markdown += `| ${key} | ${benchmark.target} | ${benchmark.current} | ${status} |\n`;
    });
    
    if (report.recommendations.length > 0) {
      markdown += '\n## Recommendations\n\n';
      
      report.recommendations.forEach(rec => {
        const emoji = rec.severity === 'high' ? '🚨' : '⚠️';
        markdown += `### ${emoji} ${rec.category} - ${rec.severity.toUpperCase()}\n\n`;
        markdown += `**Issue:** ${rec.issue}\n\n`;
        markdown += '**Recommendations:**\n';
        rec.recommendations.forEach(r => {
          markdown += `- ${r}\n`;
        });
        markdown += '\n';
      });
    }
    
    return markdown;
  }
}

// Test integration with performance analyzer
test.describe('Performance Analysis Integration', () => {
  
  test('should generate comprehensive performance analysis report', async ({ page }) => {
    const analyzer = new PerformanceAnalyzer();
    
    // Collect performance data
    const loadStart = Date.now();
    await page.goto(WEB_INTERFACE_URL, { waitUntil: 'networkidle' });
    const loadTime = Date.now() - loadStart;
    
    analyzer.addMetric('pageLoad', loadTime, { viewport: await page.viewportSize() });
    
    // Simulate WebSocket latency
    await page.waitForFunction(() => 
      window.llamaClient?.ws?.readyState === WebSocket.OPEN
    );
    
    analyzer.addMetric('webSocketLatency', 75, { connection: 'healthy' });
    
    // Memory usage simulation
    const memoryStart = await page.evaluate(() => 
      performance.memory ? performance.memory.usedJSHeapSize : 0
    );
    analyzer.addMetric('memoryUsage', memoryStart);
    
    // Perform some operations
    await page.click('[data-tab="nodes"]');
    await page.waitForTimeout(500);
    
    const memoryAfter = await page.evaluate(() => 
      performance.memory ? performance.memory.usedJSHeapSize : 0
    );
    analyzer.addMetric('memoryUsage', memoryAfter);
    
    // Generate and validate report
    const report = analyzer.generateComprehensiveReport();
    
    console.log('\n=== PERFORMANCE ANALYSIS REPORT ===');
    console.log(`Overall Score: ${report.summary.overallScore}/100`);
    console.log(`Critical Issues: ${report.summary.criticalIssues}`);
    console.log(`Warnings: ${report.summary.warnings}`);
    
    // Export reports in different formats
    const jsonReport = analyzer.exportReport('json');
    const markdownReport = analyzer.exportReport('markdown');
    
    // Save reports to files
    await page.evaluate((jsonData) => {
      // Store in localStorage for potential retrieval
      localStorage.setItem('performanceReport', jsonData);
    }, jsonReport);
    
    console.log('\nMarkdown Report Preview:');
    console.log(markdownReport.substring(0, 500) + '...');
    
    // Validate report structure
    expect(report.timestamp).toBeTruthy();
    expect(report.summary.overallScore).toBeGreaterThanOrEqual(0);
    expect(report.summary.overallScore).toBeLessThanOrEqual(100);
    expect(Array.isArray(report.recommendations)).toBe(true);
  });
});

module.exports = { PerformanceAnalyzer };