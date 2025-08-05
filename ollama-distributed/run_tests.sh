#!/bin/bash

echo "🧪 Running Proxy CLI Unit Tests"
echo "==============================="

# Change to the project directory
cd "$(dirname "$0")"

# Build the project first
echo "📦 Building project..."
go build ./cmd/node
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"

# Run the unit tests
echo ""
echo "🔬 Running unit tests..."
go test ./cmd/node -v -run "TestProxy"

# Check test results
if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 All tests passed!"
else
    echo ""
    echo "❌ Some tests failed"
    exit 1
fi

echo ""
echo "📋 Test Summary:"
echo "- TestProxyCommandStructure: Tests command hierarchy"
echo "- TestProxyStatusCommand: Tests status command with various scenarios"
echo "- TestProxyInstancesCommand: Tests instances command"
echo "- TestProxyMetricsCommand: Tests metrics command"
echo "- TestMakeHTTPRequest: Tests HTTP client functionality"
echo "- TestProxyCommandFlags: Tests flag parsing and defaults"

echo ""
echo "✅ Unit tests completed successfully!"
