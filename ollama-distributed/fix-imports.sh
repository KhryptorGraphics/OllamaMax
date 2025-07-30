#!/bin/bash

# Fix all import paths from spark to ollamamax
echo "Fixing import paths throughout the codebase..."

# Replace in all Go files
find . -name "*.go" -type f -exec sed -i 's|github.com/khryptorgraphics/spark|github.com/khryptorgraphics/ollamamax/ollama-distributed|g' {} +

# Replace in go.mod files  
find . -name "go.mod" -type f -exec sed -i 's|github.com/khryptorgraphics/spark|github.com/khryptorgraphics/ollamamax/ollama-distributed|g' {} +

# Replace in any go.sum files
find . -name "go.sum" -type f -exec sed -i 's|github.com/khryptorgraphics/spark|github.com/khryptorgraphics/ollamamax/ollama-distributed|g' {} +

echo "Import paths fixed. Running go mod tidy..."
go mod tidy

echo "Done!"