/**
 * Neural Training Integration Tests
 * Comprehensive test suite for neural training and ML integration
 */

const AgentLSTMPredictor = require('../../src/agents/agent-lstm-predictor.js');
const AgentPerformanceForecaster = require('../../src/agents/agent-performance-forecaster.js');
const NeuralPatternTrainer = require('../../src/agents/neural-pattern-trainer.js');
const HistoricalDataAggregator = require('../../src/agents/historical-data-aggregator.js');
const UnifiedNeuralOrchestrator = require('../../src/agents/unified-neural-orchestrator.js');
const MLTrainingHooks = require('../../.claude-flow/hooks/ml-training-hooks.js');

class NeuralTrainingTests {
  constructor() {
    this.results = [];
    this.startTime = Date.now();
  }

  async runAllTests() {
    console.log('🧪 Starting Neural Training Integration Tests...\n');

    await this.testMLTrainingHooks();
    await this.testAgentLSTMPredictor();
    await this.testNeuralPatternTrainer();
    await this.testHistoricalDataAggregator();
    await this.testAgentPerformanceForecaster();
    await this.testUnifiedNeuralOrchestrator();
    await this.testIntegrationWorkflow();
    await this.testPerformanceBenchmarks();

    this.printSummary();
  }

  async testMLTrainingHooks() {
    console.log('📋 Testing ML Training Hooks...');

    try {
      const hooks = new MLTrainingHooks();
      await hooks.initialize();

      // Test onPostTask
      await hooks.onPostTask('test-task-1', 'test-agent-1', {
        success: true,
        duration: 1500,
        complexity: 'medium',
        type: 'code-generation',
        specialization: 'coder'
      });

      // Test execution stats
      const stats = hooks.getExecutionStats();
      this.assert(stats.avg < 50, 'Hook execution time should be <50ms', stats.avg);

      await hooks.close();

      this.recordResult('ML Training Hooks', 'PASS', 'All hook tests passed');
    } catch (error) {
      this.recordResult('ML Training Hooks', 'FAIL', error.message);
    }
  }

  async testAgentLSTMPredictor() {
    console.log('📋 Testing Agent LSTM Predictor...');

    try {
      const predictor = new AgentLSTMPredictor();

      // Run built-in test
      const testResult = await predictor.test();
      this.assert(testResult, 'LSTM predictor test should pass');

      await predictor.close();

      this.recordResult('Agent LSTM Predictor', 'PASS', 'Model training and prediction successful');
    } catch (error) {
      this.recordResult('Agent LSTM Predictor', 'FAIL', error.message);
    }
  }

  async testNeuralPatternTrainer() {
    console.log('📋 Testing Neural Pattern Trainer...');

    try {
      const trainer = new NeuralPatternTrainer();

      // Test pattern prediction
      const testPattern = {
        agentType: 'coder',
        taskType: 'code-generation',
        complexity: 'medium',
        teamSize: 3,
        parallelism: true
      };

      const prediction = await trainer.predictPatternSuccess(testPattern);
      this.assert(prediction.success !== undefined, 'Pattern prediction should return success probability');

      await trainer.close();

      this.recordResult('Neural Pattern Trainer', 'PASS', 'Pattern training and prediction successful');
    } catch (error) {
      this.recordResult('Neural Pattern Trainer', 'FAIL', error.message);
    }
  }

  async testHistoricalDataAggregator() {
    console.log('📋 Testing Historical Data Aggregator...');

    try {
      const aggregator = new HistoricalDataAggregator();

      // Test data quality report
      const report = await aggregator.getDataQualityReport();
      this.assert(report.quality !== undefined, 'Data quality report should be generated');

      await aggregator.close();

      this.recordResult('Historical Data Aggregator', 'PASS', 'Data aggregation successful');
    } catch (error) {
      this.recordResult('Historical Data Aggregator', 'FAIL', error.message);
    }
  }

  async testAgentPerformanceForecaster() {
    console.log('📋 Testing Agent Performance Forecaster...');

    try {
      const forecaster = new AgentPerformanceForecaster();
      await forecaster.initialize();

      // Test ensemble prediction
      const taskRequest = { type: 'code-generation', complexity: 'medium' };
      const prediction = await forecaster.predictAgentPerformance('test-agent', taskRequest);

      this.assert(prediction.prediction !== undefined, 'Performance forecast should be generated');
      this.assert(prediction.confidence !== undefined, 'Confidence score should be provided');

      await forecaster.close();

      this.recordResult('Agent Performance Forecaster', 'PASS', 'Ensemble forecasting successful');
    } catch (error) {
      this.recordResult('Agent Performance Forecaster', 'FAIL', error.message);
    }
  }

  async testUnifiedNeuralOrchestrator() {
    console.log('📋 Testing Unified Neural Orchestrator...');

    try {
      const orchestrator = new UnifiedNeuralOrchestrator();

      // Test training status
      const status = await orchestrator.getTrainingStatus();
      this.assert(status.status !== undefined, 'Training status should be available');

      await orchestrator.close();

      this.recordResult('Unified Neural Orchestrator', 'PASS', 'Orchestration successful');
    } catch (error) {
      this.recordResult('Unified Neural Orchestrator', 'FAIL', error.message);
    }
  }

  async testIntegrationWorkflow() {
    console.log('📋 Testing Integration Workflow...');

    try {
      // End-to-end workflow test
      const hooks = new MLTrainingHooks();
      await hooks.initialize();

      // Simulate agent task completion
      await hooks.onPostTask('workflow-task-1', 'workflow-agent-1', {
        success: true,
        duration: 2000,
        complexity: 'high',
        type: 'research',
        specialization: 'researcher'
      });

      await hooks.close();

      this.recordResult('Integration Workflow', 'PASS', 'End-to-end workflow successful');
    } catch (error) {
      this.recordResult('Integration Workflow', 'FAIL', error.message);
    }
  }

  async testPerformanceBenchmarks() {
    console.log('📋 Testing Performance Benchmarks...');

    try {
      const startTime = Date.now();

      const predictor = new AgentLSTMPredictor();
      await predictor.predictAgentLoad('benchmark-agent');
      const predictionTime = Date.now() - startTime;

      this.assert(predictionTime < 100, 'Prediction latency should be <100ms', predictionTime);

      await predictor.close();

      this.recordResult('Performance Benchmarks', 'PASS', `Prediction latency: ${predictionTime}ms`);
    } catch (error) {
      this.recordResult('Performance Benchmarks', 'FAIL', error.message);
    }
  }

  assert(condition, message, value = null) {
    if (!condition) {
      throw new Error(`Assertion failed: ${message}${value ? ` (value: ${value})` : ''}`);
    }
  }

  recordResult(testName, status, details) {
    this.results.push({ testName, status, details, timestamp: Date.now() });
    console.log(`  ${status === 'PASS' ? '✅' : '❌'} ${testName}: ${details}\n`);
  }

  printSummary() {
    const duration = Date.now() - this.startTime;
    const passed = this.results.filter(r => r.status === 'PASS').length;
    const failed = this.results.filter(r => r.status === 'FAIL').length;

    console.log('═'.repeat(60));
    console.log('📊 Neural Training Test Summary');
    console.log('═'.repeat(60));
    console.log(`Total Tests: ${this.results.length}`);
    console.log(`✅ Passed: ${passed}`);
    console.log(`❌ Failed: ${failed}`);
    console.log(`⏱️  Duration: ${(duration / 1000).toFixed(2)}s`);
    console.log('═'.repeat(60));

    if (failed > 0) {
      console.log('\n❌ Failed Tests:');
      this.results.filter(r => r.status === 'FAIL').forEach(r => {
        console.log(`  - ${r.testName}: ${r.details}`);
      });
    }

    process.exit(failed > 0 ? 1 : 0);
  }
}

// Run tests
if (require.main === module) {
  const tests = new NeuralTrainingTests();
  tests.runAllTests().catch(error => {
    console.error('❌ Test suite failed:', error);
    process.exit(1);
  });
}

module.exports = NeuralTrainingTests;
