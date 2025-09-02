/**
 * Global Performance Test Teardown
 * Cleanup and final report generation
 */

const fs = require('fs').promises;
const path = require('path');

async function globalTeardown() {
  console.log('🧹 Performance testing teardown...');
  
  try {
    // Generate final performance summary
    const resultsDir = path.join(__dirname, '../test-results/performance');
    const summaryPath = path.join(resultsDir, 'final-summary.json');
    
    const summary = {
      completedAt: new Date().toISOString(),
      testEnvironment: {
        platform: process.platform,
        nodeVersion: process.version
      },
      status: 'completed'
    };
    
    await fs.writeFile(summaryPath, JSON.stringify(summary, null, 2));
    
    console.log('✅ Performance testing cleanup completed');
    
  } catch (error) {
    console.error('❌ Teardown error:', error);
  }
}

module.exports = globalTeardown;