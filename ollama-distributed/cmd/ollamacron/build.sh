#!/bin/bash

# Ollamacron Build Script
# Builds the Ollamacron executable with proper version information

set -e

# Build information
VERSION=${VERSION:-"dev"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION=$(go version | awk '{print $3}')

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.version=$VERSION"
LDFLAGS="$LDFLAGS -X main.commit=$COMMIT"
LDFLAGS="$LDFLAGS -X main.date=$DATE"
LDFLAGS="$LDFLAGS -X main.goVersion=$GO_VERSION"

# Build directory
BUILD_DIR=${BUILD_DIR:-"./build"}
mkdir -p "$BUILD_DIR"

# Platform detection
OS=${GOOS:-$(go env GOOS)}
ARCH=${GOARCH:-$(go env GOARCH)}

# Binary name
BINARY="ollamacron"
if [ "$OS" = "windows" ]; then
    BINARY="ollamacron.exe"
fi

OUTPUT="$BUILD_DIR/$BINARY"

echo "Building Ollamacron..."
echo "  Version: $VERSION"
echo "  Commit: $COMMIT"
echo "  Date: $DATE"
echo "  Go Version: $GO_VERSION"
echo "  OS/Arch: $OS/$ARCH"
echo "  Output: $OUTPUT"

# Build the binary
CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build \
    -ldflags "$LDFLAGS" \
    -o "$OUTPUT" \
    ./cmd/ollamacron

# Make executable
chmod +x "$OUTPUT"

echo "Build completed successfully!"
echo "Binary: $OUTPUT"

# Show version information
echo ""
echo "Version information:"
"$OUTPUT" version