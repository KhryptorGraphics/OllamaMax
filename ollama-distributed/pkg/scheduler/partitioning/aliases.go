package partitioning

import "strings"

// StrategyAliases maps various strategy names to their canonical form
var StrategyAliases = map[string]string{
	// Layer strategies
	"layer":               "layer",
	"layers":              "layer",
	"layerwise":           "layer",
	"layer-wise":          "layer",
	"layer_wise":          "layer",
	"layer_partitioning":  "layer",
	
	// Tensor strategies
	"tensor":              "tensor",
	"tensors":             "tensor",
	"tensor_parallelism":  "tensor",
	"tensor-parallelism":  "tensor",
	"tensor_parallel":     "tensor",
	"tp":                  "tensor",
	
	// Pipeline strategies
	"pipeline":            "pipeline",
	"pipeline_parallelism": "pipeline",
	"pipeline-parallelism": "pipeline",
	"pipeline_parallel":   "pipeline",
	"pp":                  "pipeline",
	
	// Hybrid strategies
	"hybrid":              "hybrid",
	"hybrid_parallelism":  "hybrid",
	"hybrid-parallelism":  "hybrid",
	"hybrid_parallel":     "hybrid",
	"mixed":               "hybrid",
	
	// Adaptive strategies
	"adaptive":            "adaptive",
	"adaptive_partitioning": "adaptive",
	"adaptive-partitioning": "adaptive",
	"auto":                "adaptive",
	"automatic":           "adaptive",
	"learning":            "adaptive",
}

// NormalizeStrategyName normalizes a strategy name to its canonical form
func NormalizeStrategyName(name string) string {
	// Convert to lowercase and trim spaces
	normalized := strings.ToLower(strings.TrimSpace(name))
	
	// Check if we have an alias for this name
	if canonical, exists := StrategyAliases[normalized]; exists {
		return canonical
	}
	
	// Return the normalized name if no alias found
	return normalized
}

// IsValidStrategy checks if a strategy name is valid (has a known alias)
func IsValidStrategy(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	_, exists := StrategyAliases[normalized]
	return exists
}

// GetCanonicalStrategies returns the list of canonical strategy names
func GetCanonicalStrategies() []string {
	// Use a map to deduplicate
	canonicalMap := make(map[string]bool)
	for _, canonical := range StrategyAliases {
		canonicalMap[canonical] = true
	}
	
	// Convert to slice
	var strategies []string
	for strategy := range canonicalMap {
		strategies = append(strategies, strategy)
	}
	
	return strategies
}