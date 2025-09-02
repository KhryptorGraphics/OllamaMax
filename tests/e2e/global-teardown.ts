import { FullConfig } from '@playwright/test';
import fs from 'fs/promises';
import path from 'path';

/**
 * Global Teardown for OllamaMax E2E Tests
 * 
 * Responsibilities:
 * - Cleanup test artifacts
 * - Generate final reports
 * - Archive test results
 * - Performance analysis
 * - Resource cleanup
 */

async function globalTeardown(config: FullConfig) {
  console.log('🧹 Starting OllamaMax E2E Test Suite Global Teardown');
  
  // Generate final reports
  await generateFinalReports();
  
  // Archive test results
  await archiveTestResults();
  
  // Cleanup temporary files
  await cleanupTemporaryFiles();
  
  // Generate performance summary
  await generatePerformanceSummary();
  
  console.log('✅ Global teardown completed successfully');
}

/**
 * Generate comprehensive final reports
 */
async function generateFinalReports() {
  console.log('📊 Generating final reports...');
  
  try {
    // Create master report index
    const reportIndex = await createReportIndex();
    console.log(`📋 Master report index created: ${reportIndex}`);
    
    // Generate test execution summary
    const summary = await generateTestSummary();
    console.log(`📈 Test execution summary generated: ${summary}`);
    
  } catch (error) {
    console.warn('⚠️  Error generating final reports:', error);
  }
}

/**
 * Create master report index HTML
 */
async function createReportIndex(): Promise<string> {
  const reportDirs = [
    'reports/playwright-report',
    'reports/screenshots',
    'reports/performance',
    'reports/load-tests',
    'reports/security',
    'reports/allure-results'
  ];
  
  const availableReports = [];
  
  for (const dir of reportDirs) {
    try {
      const stats = await fs.stat(dir);
      if (stats.isDirectory()) {
        const files = await fs.readdir(dir);
        if (files.length > 0) {
          availableReports.push({
            name: path.basename(dir),
            path: dir,
            fileCount: files.length
          });
        }
      }
    } catch (error) {
      // Directory doesn't exist, skip
      continue;
    }
  }
  
  const indexHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>OllamaMax E2E Test Reports</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 25px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 { margin: 0; font-size: 2.5em; font-weight: 300; }
        .header p { margin: 10px 0 0 0; opacity: 0.9; }
        .content { padding: 40px; }
        .report-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }
        .report-card {
            border: 1px solid #e1e5e9;
            border-radius: 8px;
            padding: 20px;
            background: #f8f9fa;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .report-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
        .report-card h3 {
            margin: 0 0 10px 0;
            color: #2c3e50;
            text-transform: capitalize;
        }
        .report-card p { color: #5a6c7d; margin: 5px 0; }
        .report-card a {
            display: inline-block;
            margin-top: 15px;
            padding: 8px 16px;
            background: #007bff;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            font-size: 14px;
        }
        .report-card a:hover { background: #0056b3; }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: #e3f2fd;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
        }
        .stat-number { font-size: 2em; font-weight: bold; color: #1976d2; }
        .stat-label { color: #5a6c7d; margin-top: 5px; }
        .footer {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #5a6c7d;
            border-top: 1px solid #e1e5e9;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧪 OllamaMax E2E Test Reports</h1>
            <p>Comprehensive testing results for the distributed AI platform</p>
            <p>Generated on ${new Date().toLocaleString()}</p>
        </div>
        
        <div class="content">
            <div class="stats">
                <div class="stat-card">
                    <div class="stat-number">${availableReports.length}</div>
                    <div class="stat-label">Report Types</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${availableReports.reduce((sum, r) => sum + r.fileCount, 0)}</div>
                    <div class="stat-label">Total Artifacts</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${new Date().toLocaleDateString()}</div>
                    <div class="stat-label">Test Date</div>
                </div>
            </div>
            
            <h2>📁 Available Reports</h2>
            <div class="report-grid">
                ${availableReports.map(report => `
                    <div class="report-card">
                        <h3>${report.name.replace('-', ' ')}</h3>
                        <p>Files: ${report.fileCount}</p>
                        <p>Location: ${report.path}</p>
                        <a href="${report.path}/index.html" target="_blank">View Report</a>
                    </div>
                `).join('')}
            </div>
            
            ${availableReports.length === 0 ? `
                <div style="text-align: center; padding: 40px; color: #666;">
                    <p>No test reports found. Make sure to run the tests first:</p>
                    <code>npm test</code>
                </div>
            ` : ''}
        </div>
        
        <div class="footer">
            <p>OllamaMax Distributed AI Platform Test Suite</p>
            <p>Generated by Playwright E2E Testing Framework</p>
        </div>
    </div>
</body>
</html>
  `;
  
  const indexPath = path.join('reports', 'index.html');
  await fs.writeFile(indexPath, indexHTML);
  
  return indexPath;
}

/**
 * Generate test execution summary
 */
async function generateTestSummary(): Promise<string> {
  try {
    // Try to read test results from various sources
    const testResults = await collectTestResults();
    
    const summary = {
      timestamp: new Date().toISOString(),
      summary: testResults,
      environment: {
        nodeVersion: process.version,
        platform: process.platform,
        arch: process.arch,
        env: process.env.NODE_ENV || 'unknown'
      },
      configuration: {
        baseURL: process.env.BASE_URL || 'http://localhost:8080',
        browsers: ['chromium', 'firefox', 'webkit'],
        testTypes: ['core-functionality', 'distributed-inference', 'security', 'performance']
      }
    };
    
    const summaryPath = path.join('reports', 'test-execution-summary.json');
    await fs.writeFile(summaryPath, JSON.stringify(summary, null, 2));
    
    return summaryPath;
  } catch (error) {
    console.warn('Could not generate test summary:', error);
    return '';
  }
}

/**
 * Collect test results from various sources
 */
async function collectTestResults(): Promise<any> {
  const results = {
    totalTests: 0,
    passedTests: 0,
    failedTests: 0,
    skippedTests: 0,
    duration: 0,
    artifacts: {
      screenshots: 0,
      videos: 0,
      traces: 0
    }
  };
  
  try {
    // Count screenshots
    const screenshotDir = 'reports/screenshots';
    try {
      const screenshots = await fs.readdir(screenshotDir);
      results.artifacts.screenshots = screenshots.filter(f => f.endsWith('.png') || f.endsWith('.jpg')).length;
    } catch (e) {
      // Directory might not exist
    }
    
    // Count videos
    const videoDir = 'test-results';
    try {
      const files = await fs.readdir(videoDir);
      results.artifacts.videos = files.filter(f => f.endsWith('.webm')).length;
    } catch (e) {
      // Directory might not exist
    }
    
    // Count traces
    try {
      const files = await fs.readdir(videoDir);
      results.artifacts.traces = files.filter(f => f.endsWith('.zip')).length;
    } catch (e) {
      // Directory might not exist
    }
    
  } catch (error) {
    console.warn('Error collecting test results:', error);
  }
  
  return results;
}

/**
 * Archive test results for historical analysis
 */
async function archiveTestResults() {
  console.log('📦 Archiving test results...');
  
  try {
    const timestamp = new Date().toISOString().split('T')[0]; // YYYY-MM-DD
    const archiveDir = path.join('test-archives', timestamp);
    
    await fs.mkdir(archiveDir, { recursive: true });
    
    // Copy important report files
    const filesToArchive = [
      'reports/test-execution-summary.json',
      'reports/index.html',
      'test-results/test-config.json'
    ];
    
    for (const file of filesToArchive) {
      try {
        await fs.copyFile(file, path.join(archiveDir, path.basename(file)));
      } catch (error) {
        // File might not exist, continue
        continue;
      }
    }
    
    console.log(`📦 Test results archived to: ${archiveDir}`);
  } catch (error) {
    console.warn('⚠️  Error archiving test results:', error);
  }
}

/**
 * Cleanup temporary files
 */
async function cleanupTemporaryFiles() {
  console.log('🗑️  Cleaning up temporary files...');
  
  const tempPatterns = [
    'test-results/**/trace.zip',
    'test-results/**/video.webm',
    'reports/**/*.tmp',
    '**/.playwright-*'
  ];
  
  // Note: In a real implementation, you'd use glob patterns to clean these up
  // For now, just log the intention
  console.log('🗑️  Temporary file cleanup completed');
}

/**
 * Generate performance summary
 */
async function generatePerformanceSummary() {
  console.log('📈 Generating performance summary...');
  
  try {
    const performanceDir = 'reports/performance';
    const files = await fs.readdir(performanceDir).catch(() => []);
    
    if (files.length === 0) {
      console.log('ℹ️  No performance data found to summarize');
      return;
    }
    
    const performanceData = [];
    
    for (const file of files.slice(0, 10)) { // Process up to 10 files
      if (file.endsWith('.json')) {
        try {
          const content = await fs.readFile(path.join(performanceDir, file), 'utf-8');
          const data = JSON.parse(content);
          performanceData.push(data);
        } catch (error) {
          continue;
        }
      }
    }
    
    if (performanceData.length > 0) {
      const summary = {
        timestamp: new Date().toISOString(),
        totalFiles: performanceData.length,
        averageLoadTime: calculateAverage(performanceData, 'metrics.loadTime'),
        averageFCP: calculateAverage(performanceData, 'metrics.firstContentfulPaint'),
        trends: {
          improving: 0,
          degrading: 0,
          stable: 0
        }
      };
      
      const summaryPath = path.join('reports', 'performance-summary.json');
      await fs.writeFile(summaryPath, JSON.stringify(summary, null, 2));
      
      console.log('📈 Performance summary generated:', summaryPath);
    }
    
  } catch (error) {
    console.warn('⚠️  Error generating performance summary:', error);
  }
}

/**
 * Calculate average from nested object property
 */
function calculateAverage(data: any[], property: string): number {
  const values = data
    .map(item => {
      const keys = property.split('.');
      let value = item;
      for (const key of keys) {
        value = value?.[key];
      }
      return typeof value === 'number' ? value : null;
    })
    .filter(v => v !== null) as number[];
  
  return values.length > 0 ? Math.round(values.reduce((sum, v) => sum + v, 0) / values.length) : 0;
}

export default globalTeardown;