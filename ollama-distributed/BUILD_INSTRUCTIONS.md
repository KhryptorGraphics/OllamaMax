# OllamaMax Build Instructions

## 🎯 Quick Start (Recommended)

### Option 1: Docker Build (Most Reliable)
```bash
# Build using Docker
docker-compose -f docker-compose.build.yml up ollama-build

# The binary will be in ./bin/ollama-distributed
./bin/ollama-distributed proxy --help
```

### Option 2: Local Build (After Environment Setup)
```bash
# Setup environment
chmod +x scripts/setup-build-env.sh
./scripts/setup-build-env.sh

# Source new environment
source ~/.bashrc

# Build
go build -o ollama-distributed ./cmd/node

# Test
./ollama-distributed proxy --help
```

## 🔧 Build Environment Issues?

If you encounter hanging Go commands, follow these steps:

### Step 1: Environment Diagnosis
```bash
# Check if Go is working
go version

# Verify environment
./scripts/verify-go-env.sh
```

### Step 2: Apply Fixes
```bash
# Run environment setup
./scripts/setup-build-env.sh

# Or reset and try again
./scripts/reset-go-env.sh
./scripts/setup-build-env.sh
```

### Step 3: Alternative Methods
```bash
# Method A: Vendor build
./scripts/build-with-vendor.sh

# Method B: Docker build
docker build -f Dockerfile.build -t ollama-distributed .

# Method C: Development container
docker-compose -f docker-compose.build.yml up ollama-dev
```

## 🐳 Docker Development

### Start Development Environment
```bash
# Start development container
docker-compose -f docker-compose.build.yml up -d ollama-dev

# Enter container
docker exec -it ollama-dev bash

# Inside container:
build.sh    # Build the project
test.sh     # Run tests
lint.sh     # Run linter
dev.sh      # Start development server
```

### Build and Test
```bash
# Build only
docker-compose -f docker-compose.build.yml up ollama-build

# Run tests
docker-compose -f docker-compose.build.yml up ollama-test

# Both build and test
docker-compose -f docker-compose.build.yml up
```

## 🛠️ Manual Build Steps

If automated scripts don't work:

### 1. Environment Setup
```bash
export GOPROXY=direct
export GOSUMDB=off
export GOMODCACHE=/tmp/gomodcache
export GOCACHE=/tmp/gocache
mkdir -p /tmp/gomodcache /tmp/gocache
```

### 2. Dependencies
```bash
# Download dependencies
go mod download

# Or create vendor
go mod vendor
```

### 3. Build
```bash
# Standard build
go build -o ollama-distributed ./cmd/node

# Or vendor build
go build -mod=vendor -o ollama-distributed ./cmd/node

# Or with specific flags
go build -ldflags="-s -w" -o ollama-distributed ./cmd/node
```

## 🧪 Testing

### Unit Tests
```bash
# All tests
go test ./...

# Proxy tests only
go test ./cmd/node -run TestProxy

# With coverage
go test -cover ./...
```

### Integration Tests
```bash
# Start the system
./ollama-distributed start &

# Test CLI commands
./ollama-distributed proxy status
./ollama-distributed proxy instances
./ollama-distributed proxy metrics
```

## 📦 Build Outputs

### Successful Build
```
✅ ollama-distributed binary created
✅ CLI commands work: ./ollama-distributed --help
✅ Proxy commands work: ./ollama-distributed proxy --help
```

### Build Artifacts
- `ollama-distributed` - Main binary
- `bin/ollama-distributed` - Docker build output
- `vendor/` - Vendored dependencies (if used)

## 🚨 Troubleshooting

### Common Issues

**Go commands hang:**
- Use Docker build method
- Check network connectivity
- Clear module cache
- Use vendor build

**Import errors:**
- Run `go mod tidy`
- Check go.mod file
- Verify Go version (1.21+)

**Permission errors:**
- Check file permissions
- Use sudo if needed
- Check disk space

**Network errors:**
- Configure GOPROXY=direct
- Check firewall settings
- Use offline build

### Getting Help

1. **Check logs**: Look for error messages in build output
2. **Environment**: Run `./scripts/verify-go-env.sh`
3. **Docker**: Use Docker build as fallback
4. **Documentation**: See BUILD_ENVIRONMENT_FIX.md

## ✅ Verification

After successful build:

```bash
# Test binary exists
ls -la ollama-distributed

# Test CLI works
./ollama-distributed --help

# Test proxy commands
./ollama-distributed proxy --help
./ollama-distributed proxy status --help

# Test version
./ollama-distributed version
```

## 🎉 Success!

Once built successfully, you can:

1. **Use CLI commands**: `./ollama-distributed proxy status`
2. **Start the system**: `./ollama-distributed start`
3. **Monitor cluster**: `./ollama-distributed proxy metrics --watch`
4. **Deploy**: Copy binary to target systems

The OllamaMax distributed system is now ready for use!
