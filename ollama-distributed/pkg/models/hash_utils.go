package models

import (
	"crypto/sha256"
	"fmt"
)

// calculateModelHash calculates a hash for a model name
func calculateModelHash(modelName string) string {
	hash := sha256.Sum256([]byte(modelName))
	return fmt.Sprintf("%x", hash)
}