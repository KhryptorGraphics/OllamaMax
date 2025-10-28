.PHONY: build-app clean test fmt vet all

# Build the application binary
build-app:
	go build -o bin/ollamamax ./cmd/ollama-distributed

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Run tests
test:
	go test -v -coverprofile=coverage.out -covermode=atomic ./internal/... ./pkg/... ./cmd/...

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Vet code
vet:
	go vet ./...

# Run all checks
all: fmt vet test build-app
