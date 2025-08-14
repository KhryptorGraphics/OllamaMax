package fault_tolerance

import (
	"fmt"
	"math"
	"time"
)

// Mathematical utility functions

// calculateMean calculates the mean of a slice of float64 values
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// calculateStdDev calculates the standard deviation of a slice of float64 values
func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	mean := calculateMean(values)
	variance := 0.0

	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}

	variance /= float64(len(values))
	return math.Sqrt(variance)
}

// calculateTrend calculates the trend (slope) of a time series
func calculateTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	n := float64(len(values))
	sumX := n * (n - 1) / 2
	sumY := 0.0
	sumXY := 0.0
	sumX2 := n * (n - 1) * (2*n - 1) / 6

	for i, y := range values {
		sumY += y
		sumXY += float64(i) * y
	}

	// Calculate slope using linear regression
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	return slope
}

// calculateVolatility calculates the volatility (standard deviation of returns)
func calculateVolatility(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	// Calculate returns (percentage changes)
	returns := make([]float64, len(values)-1)
	for i := 1; i < len(values); i++ {
		if values[i-1] != 0 {
			returns[i-1] = (values[i] - values[i-1]) / values[i-1]
		}
	}

	return calculateStdDev(returns)
}

// Conversion utility functions

// stringToFaultType converts string to FaultType
func stringToFaultType(s string) FaultType {
	switch s {
	case "resource_exhaustion":
		return FaultTypeResourceExhaustion
	case "performance_anomaly":
		return FaultTypePerformanceAnomaly
	case "network_partition":
		return FaultTypeNetworkPartition
	case "service_unavailable":
		return FaultTypeServiceUnavailable
	default:
		return FaultTypeServiceUnavailable
	}
}

// getTrendDirection returns the direction of a trend
func getTrendDirection(trend float64) string {
	if trend > 0.001 {
		return "increasing"
	} else if trend < -0.001 {
		return "decreasing"
	} else {
		return "stable"
	}
}

// ML utility functions

// featuresToVector converts feature map to vector for neural network
func (mlp *MLPredictor) featuresToVector(features map[string]float64) []float64 {
	vector := make([]float64, mlp.neuralNetwork.InputSize)

	// Map features to vector positions
	featureNames := []string{
		"cpu_usage", "memory_usage", "response_time", "error_rate", "connectivity",
		"cpu_usage_mean", "memory_usage_mean", "response_time_mean", "error_rate_mean", "connectivity_mean",
	}

	for i, featureName := range featureNames {
		if i >= len(vector) {
			break
		}
		if value, exists := features[featureName]; exists {
			vector[i] = value
		}
	}

	return vector
}

// forwardPass performs forward pass through neural network
func (mlp *MLPredictor) forwardPass(input []float64) []float64 {
	nn := mlp.neuralNetwork

	// Hidden layer
	hidden := make([]float64, nn.HiddenSize)
	for j := 0; j < nn.HiddenSize; j++ {
		sum := nn.Bias1[j]
		for i := 0; i < nn.InputSize; i++ {
			if i < len(input) && i < len(nn.Weights1) && j < len(nn.Weights1[i]) {
				sum += input[i] * nn.Weights1[i][j]
			}
		}
		hidden[j] = sigmoid(sum)
	}

	// Output layer
	output := make([]float64, nn.OutputSize)
	for k := 0; k < nn.OutputSize; k++ {
		sum := nn.Bias2[k]
		for j := 0; j < nn.HiddenSize; j++ {
			if j < len(nn.Weights2) && k < len(nn.Weights2[j]) {
				sum += hidden[j] * nn.Weights2[j][k]
			}
		}
		output[k] = sigmoid(sum)
	}

	return output
}

// sigmoid activation function
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// outputToPredictions converts neural network output to predictions
func (mlp *MLPredictor) outputToPredictions(output []float64, modelName string, features map[string]float64) []*FaultPrediction {
	var predictions []*FaultPrediction

	// Define fault types corresponding to output neurons
	faultTypes := []FaultType{
		FaultTypeResourceExhaustion,
		FaultTypePerformanceAnomaly,
		FaultTypeNetworkPartition,
		FaultTypeServiceUnavailable,
	}

	// Find the highest confidence prediction
	maxConfidence := 0.0
	maxIndex := -1
	for i, confidence := range output {
		if confidence > maxConfidence && confidence > 0.5 { // Threshold
			maxConfidence = confidence
			maxIndex = i
		}
	}

	// Create prediction if confidence is high enough
	if maxIndex >= 0 && maxIndex < len(faultTypes) {
		timeToFailure := time.Duration((1.0-maxConfidence)*10) * time.Minute

		predictions = append(predictions, &FaultPrediction{
			ID:            fmt.Sprintf("nn_pred_%d", time.Now().UnixNano()),
			PredictedType: faultTypes[maxIndex],
			Target:        "system",
			Confidence:    maxConfidence,
			TimeToFailure: timeToFailure,
			PredictedAt:   time.Now(),
			ExpectedAt:    time.Now().Add(timeToFailure),
			ModelUsed:     modelName,
			Features:      features,
			Metadata: map[string]interface{}{
				"output_vector":   output,
				"max_index":       maxIndex,
				"all_confidences": output,
			},
			Status: PredictionStatusPending,
		})
	}

	return predictions
}

// predictWithTree performs prediction using a decision tree
func (mlp *MLPredictor) predictWithTree(tree *DecisionTree, features map[string]float64) string {
	if tree.IsLeaf {
		return tree.Prediction
	}

	featureValue, exists := features[tree.Feature]
	if !exists {
		return ""
	}

	if featureValue <= tree.Threshold {
		if tree.Left != nil {
			return mlp.predictWithTree(tree.Left, features)
		}
	} else {
		if tree.Right != nil {
			return mlp.predictWithTree(tree.Right, features)
		}
	}

	return ""
}

// Correlation utility functions

// updateCausalityGraph updates the causality graph
func (ce *CorrelationEngine) updateCausalityGraph(timeSeriesData map[string]*TimeSeriesData) {
	// Simple causality detection based on correlation and lag
	for metric1, data1 := range timeSeriesData {
		for metric2, data2 := range timeSeriesData {
			if metric1 != metric2 {
				// Calculate cross-correlation with different lags
				maxCorrelation := 0.0
				bestLag := 0

				for lag := 0; lag < ce.config.CrossCorrelationLags; lag++ {
					correlation := ce.calculateLaggedCorrelation(data1.Values, data2.Values, lag)
					if math.Abs(correlation) > math.Abs(maxCorrelation) {
						maxCorrelation = correlation
						bestLag = lag
					}
				}

				// Create causality edge if correlation is significant
				if math.Abs(maxCorrelation) > 0.5 {
					ce.addCausalityEdge(metric1, metric2, maxCorrelation, bestLag)
				}
			}
		}
	}
}

// calculateLaggedCorrelation calculates correlation with lag
func (ce *CorrelationEngine) calculateLaggedCorrelation(x, y []float64, lag int) float64 {
	if lag >= len(x) || lag >= len(y) {
		return 0
	}

	// Adjust arrays for lag
	var x1, y1 []float64
	if lag > 0 {
		x1 = x[:len(x)-lag]
		y1 = y[lag:]
	} else {
		x1 = x
		y1 = y
	}

	return ce.calculateCorrelation(x1, y1)
}

// addCausalityEdge adds an edge to the causality graph
func (ce *CorrelationEngine) addCausalityEdge(from, to string, strength float64, lag int) {
	// Add nodes if they don't exist
	if ce.causalityGraph.Nodes[from] == nil {
		ce.causalityGraph.Nodes[from] = &CausalityNode{
			Name:       from,
			Type:       "metric",
			Importance: 0.5,
			LastSeen:   time.Now(),
		}
	}

	if ce.causalityGraph.Nodes[to] == nil {
		ce.causalityGraph.Nodes[to] = &CausalityNode{
			Name:       to,
			Type:       "metric",
			Importance: 0.5,
			LastSeen:   time.Now(),
		}
	}

	// Add edge
	edge := &CausalityEdge{
		From:       from,
		To:         to,
		Strength:   strength,
		Lag:        lag,
		Confidence: math.Abs(strength),
	}

	if ce.causalityGraph.Edges[from] == nil {
		ce.causalityGraph.Edges[from] = make([]*CausalityEdge, 0)
	}

	ce.causalityGraph.Edges[from] = append(ce.causalityGraph.Edges[from], edge)
}

// generateCorrelationPredictions generates predictions based on correlations
func (ce *CorrelationEngine) generateCorrelationPredictions(metrics map[string]interface{}, timeSeriesData map[string]*TimeSeriesData) []*FaultPrediction {
	var predictions []*FaultPrediction

	// Look for strong correlations that might indicate impending faults
	for metric1, correlation1 := range ce.correlationMatrix {
		for metric2, correlationValue := range correlation1 {
			if math.Abs(correlationValue) > 0.8 { // Strong correlation
				// Check if one metric is showing anomalous behavior
				if data1, exists1 := timeSeriesData[metric1]; exists1 {
					if data2, exists2 := timeSeriesData[metric2]; exists2 {
						if ce.isAnomalousCorrelation(data1, data2, correlationValue) {
							prediction := ce.createCorrelationPrediction(metric1, metric2, correlationValue)
							if prediction != nil {
								predictions = append(predictions, prediction)
							}
						}
					}
				}
			}
		}
	}

	return predictions
}

// isAnomalousCorrelation checks if correlation indicates anomaly
func (ce *CorrelationEngine) isAnomalousCorrelation(data1, data2 *TimeSeriesData, expectedCorrelation float64) bool {
	if len(data1.Values) < 10 || len(data2.Values) < 10 {
		return false
	}

	// Calculate recent correlation
	recentSize := 10
	recent1 := data1.Values[len(data1.Values)-recentSize:]
	recent2 := data2.Values[len(data2.Values)-recentSize:]

	recentCorrelation := ce.calculateCorrelation(recent1, recent2)

	// Check if recent correlation deviates significantly from expected
	return math.Abs(recentCorrelation-expectedCorrelation) > 0.3
}

// createCorrelationPrediction creates a prediction based on correlation anomaly
func (ce *CorrelationEngine) createCorrelationPrediction(metric1, metric2 string, correlation float64) *FaultPrediction {
	confidence := math.Abs(correlation) * 0.8 // Scale confidence
	timeToFailure := 5 * time.Minute

	return &FaultPrediction{
		ID:            fmt.Sprintf("corr_pred_%d", time.Now().UnixNano()),
		PredictedType: FaultTypePerformanceAnomaly,
		Target:        fmt.Sprintf("%s-%s", metric1, metric2),
		Confidence:    confidence,
		TimeToFailure: timeToFailure,
		PredictedAt:   time.Now(),
		ExpectedAt:    time.Now().Add(timeToFailure),
		ModelUsed:     "correlation_analysis",
		Features: map[string]float64{
			"correlation": correlation,
		},
		Metadata: map[string]interface{}{
			"metric1":     metric1,
			"metric2":     metric2,
			"correlation": correlation,
		},
		Status: PredictionStatusPending,
	}
}
