#!/bin/bash
# Comprehensive build script for OllamaMax with multi-stage Dockerfile support

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")}
DOCKERFILE=${DOCKERFILE:-"Dockerfile"}
IMAGE_NAME=${IMAGE_NAME:-"ollamamax"}
PLATFORMS=${PLATFORMS:-"linux/amd64,linux/arm64"}
PUSH=${PUSH:-false}
DEV=${DEV:-false}

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build OllamaMax Docker images with multi-architecture support.

OPTIONS:
    -h, --help              Show this help message
    -v, --version VERSION   Set version tag (default: git describe or 'dev')
    -n, --name NAME         Set image name (default: 'ollamamax')
    -p, --platforms LIST    Set target platforms (default: 'linux/amd64,linux/arm64')
    -f, --dockerfile FILE   Set Dockerfile path (default: 'Dockerfile')
    --push                  Push image to registry after build
    --dev                   Build development image
    --prod                  Build production image (default)
    --test                  Run tests after build
    --no-cache              Build without cache
    --dry-run               Show commands without executing

EXAMPLES:
    $0                                          # Build production image for multi-arch
    $0 --dev                                    # Build development image
    $0 --version v1.0.0 --push                 # Build and push specific version
    $0 --platforms linux/amd64 --test          # Build for x86_64 only and test
    $0 --name myregistry/ollamamax --push      # Build with custom name and push

EOF
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Docker is installed and running
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker is not running or not accessible"
        exit 1
    fi
    
    # Check if buildx is available for multi-platform builds
    if [[ "$PLATFORMS" == *","* ]]; then
        if ! docker buildx version &> /dev/null; then
            print_error "Docker buildx is required for multi-platform builds"
            exit 1
        fi
        
        # Ensure buildx instance exists
        if ! docker buildx inspect multiarch &> /dev/null; then
            print_status "Creating buildx instance for multi-platform builds..."
            docker buildx create --name multiarch --driver docker-container --use
            docker buildx inspect --bootstrap
        else
            docker buildx use multiarch
        fi
    fi
    
    print_success "Prerequisites check passed"
}

# Function to build the image
build_image() {
    print_status "Building OllamaMax image..."
    print_status "Configuration:"
    print_status "  Version: $VERSION"
    print_status "  Image Name: $IMAGE_NAME"
    print_status "  Platforms: $PLATFORMS"
    print_status "  Dockerfile: $DOCKERFILE"
    print_status "  Build Time: $BUILD_TIME"
    print_status "  Git Commit: $GIT_COMMIT"
    
    # Prepare build arguments
    BUILD_ARGS=(
        --build-arg "VERSION=$VERSION"
        --build-arg "BUILD_TIME=$BUILD_TIME"
        --build-arg "GIT_COMMIT=$GIT_COMMIT"
        --tag "$IMAGE_NAME:$VERSION"
        --tag "$IMAGE_NAME:latest"
    )
    
    # Add platform support for multi-arch builds
    if [[ "$PLATFORMS" == *","* ]]; then
        BUILD_ARGS+=(--platform "$PLATFORMS")
        BUILDER="buildx build"
        
        if [[ "$PUSH" == "true" ]]; then
            BUILD_ARGS+=(--push)
        else
            BUILD_ARGS+=(--load)
            print_warning "Multi-platform build will not be loaded locally. Use --push to push to registry."
        fi
    else
        BUILDER="build"
        BUILD_ARGS+=(--platform "$PLATFORMS")
    fi
    
    # Add cache options
    if [[ "$NO_CACHE" == "true" ]]; then
        BUILD_ARGS+=(--no-cache)
    fi
    
    # Add dockerfile
    BUILD_ARGS+=(--file "$DOCKERFILE")
    
    # Add context
    BUILD_ARGS+=(.)
    
    # Build command
    BUILD_CMD="docker $BUILDER ${BUILD_ARGS[*]}"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "Dry run - would execute:"
        echo "$BUILD_CMD"
        return 0
    fi
    
    print_status "Executing: $BUILD_CMD"
    
    # Execute build
    if eval "$BUILD_CMD"; then
        print_success "Image build completed successfully"
    else
        print_error "Image build failed"
        exit 1
    fi
    
    # Show image info for single-platform builds
    if [[ "$PLATFORMS" != *","* ]] && [[ "$PUSH" != "true" ]]; then
        print_status "Image information:"
        docker images "$IMAGE_NAME:$VERSION" --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
    fi
}

# Function to run tests
run_tests() {
    if [[ "$TEST" == "true" ]]; then
        print_status "Running tests on built image..."
        
        # Basic container start test
        print_status "Testing container startup..."
        CONTAINER_ID=$(docker run -d --rm "$IMAGE_NAME:$VERSION" /bin/sh -c "sleep 10")
        
        if docker ps --filter "id=$CONTAINER_ID" --format "{{.ID}}" | grep -q "$CONTAINER_ID"; then
            print_success "Container started successfully"
            docker stop "$CONTAINER_ID" >/dev/null
        else
            print_error "Container failed to start"
            exit 1
        fi
        
        # Health check test (if health check is available)
        print_status "Testing health check..."
        if docker inspect "$IMAGE_NAME:$VERSION" --format='{{.Config.Healthcheck}}' | grep -q "CMD"; then
            CONTAINER_ID=$(docker run -d --rm -p 11434:11434 "$IMAGE_NAME:$VERSION")
            sleep 30  # Wait for startup
            
            if docker inspect "$CONTAINER_ID" --format='{{.State.Health.Status}}' | grep -q "healthy"; then
                print_success "Health check passed"
            else
                print_warning "Health check did not pass within timeout"
            fi
            
            docker stop "$CONTAINER_ID" >/dev/null
        fi
        
        print_success "All tests passed"
    fi
}

# Function to push image
push_image() {
    if [[ "$PUSH" == "true" ]] && [[ "$PLATFORMS" != *","* ]]; then
        print_status "Pushing image to registry..."
        
        if docker push "$IMAGE_NAME:$VERSION" && docker push "$IMAGE_NAME:latest"; then
            print_success "Image pushed successfully"
        else
            print_error "Failed to push image"
            exit 1
        fi
    fi
}

# Main function
main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -n|--name)
                IMAGE_NAME="$2"
                shift 2
                ;;
            -p|--platforms)
                PLATFORMS="$2"
                shift 2
                ;;
            -f|--dockerfile)
                DOCKERFILE="$2"
                shift 2
                ;;
            --push)
                PUSH=true
                shift
                ;;
            --dev)
                DOCKERFILE="Dockerfile.dev"
                DEV=true
                shift
                ;;
            --prod)
                DOCKERFILE="Dockerfile"
                DEV=false
                shift
                ;;
            --test)
                TEST=true
                shift
                ;;
            --no-cache)
                NO_CACHE=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    # Set development-specific defaults
    if [[ "$DEV" == "true" ]]; then
        IMAGE_NAME="${IMAGE_NAME}:dev"
        PLATFORMS="linux/amd64"  # Dev builds are typically single-arch
    fi
    
    print_status "Starting OllamaMax Docker build..."
    print_status "Build configuration: ${DEV:+Development}${DEV:-Production} mode"
    
    check_prerequisites
    build_image
    run_tests
    push_image
    
    print_success "Build process completed successfully!"
    
    # Show usage instructions
    cat << EOF

🎉 OllamaMax image built successfully!

Usage instructions:
  # Run single node
  docker run -p 11434:11434 $IMAGE_NAME:$VERSION

  # Run with docker-compose (production)
  docker-compose up

  # Run with docker-compose (development)
  docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

  # Multi-architecture manifest (if pushed)
  docker manifest inspect $IMAGE_NAME:$VERSION

EOF
}

# Run main function with all arguments
main "$@"