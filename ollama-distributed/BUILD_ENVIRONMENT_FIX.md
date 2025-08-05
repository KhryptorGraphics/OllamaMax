# Build Environment Fix Guide

## 🎯 Issue Diagnosis

The build environment has issues where Go commands hang indefinitely. This guide provides comprehensive fixes.

## 🔧 Common Causes and Solutions

### 1. **GOPROXY Issues**
**Symptoms:** `go mod download` hangs
**Solution:** Configure alternative proxy

```bash
# Set direct mode (bypass proxy)
export GOPROXY=direct
export GOSUMDB=off

# Or use alternative proxy
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn
```

### 2. **Network Connectivity**
**Symptoms:** All Go commands hang
**Solution:** Configure network settings

```bash
# Check DNS resolution
nslookup proxy.golang.org

# Configure Git for Go modules
git config --global url."https://".insteadOf git://

# Set timeout for Go operations
export GOTIMEOUT=30s
```

### 3. **Module Cache Issues**
**Symptoms:** Commands hang during module resolution
**Solution:** Clear and reconfigure cache

```bash
# Clear module cache
rm -rf $(go env GOMODCACHE)
rm -rf $(go env GOCACHE)

# Set new cache location
export GOMODCACHE=/tmp/gomodcache
export GOCACHE=/tmp/gocache
```

### 4. **Go Version Compatibility**
**Symptoms:** Incompatible Go version
**Solution:** Use compatible version

```bash
# Check Go version
go version

# Use Go 1.21+ for this project
# Download from: https://golang.org/dl/
```

## 🚀 Quick Fix Script

Create and run this script to fix common issues:

```bash
#!/bin/bash
# build-env-fix.sh

echo "🔧 Fixing Go Build Environment"

# Set environment variables
export GOPROXY=direct
export GOSUMDB=off
export GOTIMEOUT=30s
export GOMODCACHE=/tmp/gomodcache
export GOCACHE=/tmp/gocache

# Create cache directories
mkdir -p /tmp/gomodcache
mkdir -p /tmp/gocache

# Configure Git
git config --global url."https://".insteadOf git://

echo "✅ Environment configured"
echo "Now try: go build ./cmd/node"
```

## 🐳 Docker Alternative

If local Go environment issues persist, use Docker:

```dockerfile
# Dockerfile.build
FROM golang:1.21-alpine

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o ollama-distributed ./cmd/node

CMD ["./ollama-distributed"]
```

Build with Docker:
```bash
docker build -f Dockerfile.build -t ollama-distributed .
docker run ollama-distributed proxy --help
```

## 🔍 Diagnostic Commands

Test environment step by step:

```bash
# 1. Test basic Go
echo 'package main; import "fmt"; func main() { fmt.Println("Go works") }' > test.go
timeout 10s go run test.go

# 2. Test module system
timeout 10s go mod verify

# 3. Test specific package
timeout 30s go build -o /dev/null ./pkg/config

# 4. Test main command
timeout 60s go build -o /dev/null ./cmd/node
```

## 🛠️ Alternative Build Methods

### Method 1: Offline Build
```bash
# Download dependencies separately
go mod download -x

# Build without network
go build -mod=readonly ./cmd/node
```

### Method 2: Vendor Dependencies
```bash
# Create vendor directory
go mod vendor

# Build using vendor
go build -mod=vendor ./cmd/node
```

### Method 3: Cross-Platform Build
```bash
# Build for different platforms
GOOS=linux GOARCH=amd64 go build ./cmd/node
GOOS=windows GOARCH=amd64 go build ./cmd/node
GOOS=darwin GOARCH=amd64 go build ./cmd/node
```

## 📋 Troubleshooting Checklist

- [ ] Go version 1.21+ installed
- [ ] GOPROXY configured (try `direct`)
- [ ] GOSUMDB disabled if needed
- [ ] Network connectivity verified
- [ ] Module cache cleared
- [ ] Git configured for HTTPS
- [ ] Sufficient disk space
- [ ] No firewall blocking Go traffic

## 🎯 Success Verification

Once fixed, these should work:
```bash
go version                    # Shows Go version
go env GOPROXY               # Shows proxy config
go mod verify                # Verifies modules
go build ./cmd/node          # Builds main command
./node proxy --help          # Shows CLI help
```

## 🚨 Emergency Workaround

If all else fails, use pre-built binaries or alternative build systems:

1. **Use GitHub Actions** to build remotely
2. **Use Docker** for isolated build environment
3. **Use Go Playground** for testing small components
4. **Use alternative Go distributions** (TinyGo, etc.)

## 📞 Support Resources

- **Go Documentation**: https://golang.org/doc/
- **Module Documentation**: https://golang.org/ref/mod
- **Proxy Documentation**: https://proxy.golang.org/
- **Build Troubleshooting**: https://golang.org/doc/faq#build_fails
