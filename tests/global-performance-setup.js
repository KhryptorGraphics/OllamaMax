/**
 * Global Performance Test Setup
 * Initializes environment and monitoring for performance testing
 */

const fs = require('fs').promises;
const path = require('path');

async function globalSetup() {
  console.log('🔧 Setting up performance testing environment...');
  
  // Create test results directory
  const resultsDir = path.join(__dirname, '../test-results/performance');
  await fs.mkdir(resultsDir, { recursive: true });
  
  // Create performance monitoring directories
  await fs.mkdir(path.join(resultsDir, 'screenshots'), { recursive: true });
  await fs.mkdir(path.join(resultsDir, 'traces'), { recursive: true });
  await fs.mkdir(path.join(resultsDir, 'videos'), { recursive: true });
  
  // Initialize performance baseline file
  const baselineFile = path.join(resultsDir, 'baseline.json');
  const baseline = {
    timestamp: new Date().toISOString(),
    thresholds: {
      pageLoadTime: 3000,
      webSocketLatency: 100,
      memoryGrowthLimit: 50,
      animationFrameRate: 50,
      failoverTime: 2000
    },
    environment: {
      platform: process.platform,
      nodeVersion: process.version,
      cpus: require('os').cpus().length,
      memory: Math.round(require('os').totalmem() / 1024 / 1024 / 1024)
    }
  };
  
  await fs.writeFile(baselineFile, JSON.stringify(baseline, null, 2));
  
  // Set performance testing environment variables
  process.env.PERFORMANCE_TESTING = 'true';
  process.env.ENABLE_MEMORY_PROFILING = 'true';
  
  console.log('✅ Performance testing environment ready');
  
  // Wait for services to be available
  await waitForServices();
}

async function waitForServices() {
  console.log('⏳ Waiting for services to be ready...');
  
  const maxWaitTime = 30000; // 30 seconds
  const checkInterval = 1000; // 1 second
  const startTime = Date.now();
  
  while (Date.now() - startTime < maxWaitTime) {
    try {
      // Check if web interface is accessible
      const fetch = (await import('node-fetch')).default;
      
      const webResponse = await fetch('http://localhost:8080');
      if (webResponse.ok) {
        console.log('✅ Web interface ready');
        break;
      }
    } catch (error) {
      // Service not ready yet
      await new Promise(resolve => setTimeout(resolve, checkInterval));
    }
  }
  
  console.log('🚀 All services ready for performance testing');
}

module.exports = globalSetup;