# Neural Training & AI-Powered Optimization System

## Overview

This directory contains the Sprint 3 neural training and AI optimization system that bridges Sprint 2 ML Pipeline, Advanced Swarm Coordination, and Claude-Flow Infrastructure.

## Components

### Core Neural Training

- **agent-lstm-predictor.js** - Per-agent LSTM load and performance forecasting
- **neural-pattern-trainer.js** - Pattern recognition using neural networks
- **agent-performance-forecaster.js** - Ensemble performance prediction (LSTM + Random Forest + Neural Learning)
- **historical-data-aggregator.js** - Multi-source data collection and normalization
- **unified-neural-orchestrator.js** - Coordinated training across all systems
- **swarm-algorithm-optimizer.js** - ML-powered coordination algorithm selection

## Quick Start

### Training Neural Models

```bash
# Train all neural models
npm run neural:train

# Train specific models
npm run neural:train:agent-lstm
npm run neural:train:patterns
npm run neural:train:algorithms
```

### Making Predictions

```bash
# Predict agent performance
npm run neural:predict:agent

# Predict swarm load
npm run neural:predict:load

# Predict pattern success
npm run neural:predict:patterns
```

### Data Aggregation

```bash
# Aggregate historical data
npm run neural:aggregate

# Sync Claude-Flow memory to Redis
npm run neural:sync
```

## Architecture

### Data Flow

1. **Hook Events** → ML Training Hooks
2. **Historical Data** → Data Aggregator
3. **Training Data** → Neural Models
4. **Predictions** → Swarm Coordination
5. **Results** → Training Loop

### Integration Points

- **EnhancedSmartAgentsSwarm**: Uses ML predictions for agent selection
- **QueenCoordinator**: Leverages ML insights for strategic planning
- **CrossAgentLearningSystem**: Integrates with ML for bidirectional learning
- **SwarmPerformanceOptimizer**: Uses neural insights for optimization

## API Reference

### Agent LSTM Predictor

```javascript
const predictor = new AgentLSTMPredictor();

// Predict agent load
const loadPrediction = await predictor.predictAgentLoad(agentId);
// Returns: { load: number, confidence: number }

// Predict availability
const availability = await predictor.predictAgentAvailability(agentId);
// Returns: { available: boolean, probability: number }

// Predict performance
const performance = await predictor.predictAgentPerformance(agentId, taskType);
// Returns: { successRate: number, estimatedDuration: number }
```

### Agent Performance Forecaster

```javascript
const forecaster = new AgentPerformanceForecaster();
await forecaster.initialize();

// Comprehensive performance forecast
const forecast = await forecaster.predictAgentPerformance(agentId, taskRequest);

// Rank agents by predicted performance
const ranked = await forecaster.rankAgentsByPredictedPerformance(taskRequest, availableAgents);

// Get optimal agent for task
const optimal = await forecaster.getOptimalAgentForTask(taskRequest);
```

### Neural Pattern Trainer

```javascript
const trainer = new NeuralPatternTrainer();

// Train on patterns
const patterns = await trainer.collectPatterns();
const result = await trainer.trainOnPatterns(patterns);

// Predict pattern success
const prediction = await trainer.predictPatternSuccess(pattern);

// Recommend patterns
const recommendations = await trainer.recommendPatterns(taskContext);
```

### Historical Data Aggregator

```javascript
const aggregator = new HistoricalDataAggregator();

// Aggregate historical data
const data = await aggregator.aggregateHistoricalData(startTime, endTime);

// Get training dataset
const dataset = await aggregator.getTrainingDataset('agent_selection', 1000);

// Sync Claude-Flow memory to Redis
await aggregator.syncClaudeFlowMemoryToRedis();

// Get data quality report
const report = await aggregator.getDataQualityReport();
```

### Unified Neural Orchestrator

```javascript
const orchestrator = new UnifiedNeuralOrchestrator();

// Orchestrate full training
const result = await orchestrator.orchestrateFullTraining();

// Train specific model
const modelResult = await orchestrator.trainSpecificModel('agent-lstm');

// Get training status
const status = await orchestrator.getTrainingStatus();

// Validate all models
const validation = await orchestrator.validateAllModels();
```

## Performance Metrics

### Training Performance
- **Training Time**: <5 minutes per model
- **Training Frequency**: Every 2-6 hours depending on model
- **Data Sources**: Redis metrics, Claude-Flow memory, swarm metrics

### Prediction Performance
- **Prediction Latency**: <100ms for agent selection, <500ms for scaling
- **Model Accuracy**: >80% for agent selection, >75% for LSTM
- **Pattern Recognition**: >90% confidence
- **Cache Hit Rate**: >70% for repeated predictions

## Deployment

### Prerequisites
- Sprint 1 (Redis cluster) running
- Sprint 2 (ML pipeline) running
- Node.js 18+ with dependencies installed

### Deploy Neural Training Services

```bash
# Deploy to Kubernetes
npm run deploy:neural

# Validate deployment
npm run deploy:validate:neural
```

### Monitoring

- **Grafana Dashboard**: `monitoring/grafana/dashboards/neural-training.json`
- **Prometheus Metrics**: Scraped from ports 8087-8090
- **Logs**: Available via `kubectl logs`

## Testing

```bash
# Run neural training tests
npm run test:neural

# Run integration tests
npm run test:neural:integration

# Check training status
npm run neural:status
```

## Troubleshooting

### Common Issues

**Issue**: Insufficient training data
**Solution**: Ensure agents have been running for at least 1 hour to collect sufficient metrics

**Issue**: Model accuracy below threshold
**Solution**: Check data quality report and increase training data collection period

**Issue**: Prediction latency too high
**Solution**: Enable prediction caching and reduce model complexity

### Debug Commands

```bash
# Check data aggregation
node src/agents/historical-data-aggregator.js validate

# Test LSTM predictor
node src/agents/agent-lstm-predictor.js test

# Check training status
node src/agents/unified-neural-orchestrator.js status
```

## Best Practices

1. **Training Data Quality**: Ensure data is collected from multiple sources and normalized
2. **Model Validation**: Always validate models before deploying to production
3. **Prediction Confidence**: Only use predictions with confidence >0.7 for critical decisions
4. **Continuous Improvement**: Retrain models regularly based on latest performance data
5. **A/B Testing**: Test new models against existing ones before full rollout

## Contributing

When adding new neural models:
1. Implement in `src/agents/` directory
2. Register with `unified-neural-orchestrator.js`
3. Add validation logic
4. Create tests in `tests/ml/`
5. Update documentation

## License

Part of OllamaMax project.
