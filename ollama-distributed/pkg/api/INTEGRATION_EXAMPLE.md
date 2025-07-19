# API Integration Example

This document shows how to integrate the distributed API layer with the existing Ollama server.

## Integration with Existing routes.go

To integrate with the existing `/home/kp/ollamamax/ollama/server/routes.go`, you would modify the `GenerateRoutes` function as follows:

```go
// In /home/kp/ollamamax/ollama/server/routes.go
// Add this import at the top
import "github.com/ollama/ollama-distributed/pkg/api"

// Modify the GenerateRoutes function
func (s *Server) GenerateRoutes(rc *ollama.Registry) (http.Handler, error) {
    // Check if distributed mode is enabled
    if os.Getenv("OLLAMA_DISTRIBUTED") != "false" {
        // Initialize distributed components
        scheduler := getScheduler() // Your scheduler instance
        modelDist := getModelDistribution() // Your model distribution instance
        localAddr := "http://localhost:11434" // Local Ollama address
        
        // Create route integration
        integration, err := api.NewRouteIntegration(scheduler, modelDist, localAddr)
        if err != nil {
            log.Printf("Failed to create route integration: %v", err)
            // Fall back to original routes
            return s.generateOriginalRoutes(rc)
        }
        
        // Initialize with this server
        if err := integration.Initialize(s, scheduler, modelDist, localAddr); err != nil {
            log.Printf("Failed to initialize route integration: %v", err)
            // Fall back to original routes
            return s.generateOriginalRoutes(rc)
        }
        
        // Use wrapped routes
        return integration.WrapGenerateRoutes(s, scheduler, modelDist, localAddr)(rc)
    }
    
    // Use original routes
    return s.generateOriginalRoutes(rc)
}

// Rename the original GenerateRoutes to this
func (s *Server) generateOriginalRoutes(rc *ollama.Registry) (http.Handler, error) {
    // ... original GenerateRoutes implementation
}
```

## Seamless Integration Example

```go
package main

import (
    "context"
    "log"
    "net"
    "os"
    
    "github.com/ollama/ollama/server"
    "github.com/ollama/ollama-distributed/pkg/api"
    "github.com/ollama/ollama-distributed/pkg/scheduler"
    "github.com/ollama/ollama-distributed/pkg/models"
)

func main() {
    // Create original server
    originalServer := &server.Server{}
    
    // Create distributed components
    scheduler := scheduler.NewEngine()
    modelDist := models.NewDistribution()
    localAddr := "http://localhost:11434"
    
    // Create distributed wrapper
    wrapper, err := api.NewDistributedServerWrapper(
        originalServer, 
        scheduler, 
        modelDist, 
        localAddr,
    )
    if err != nil {
        log.Fatalf("Failed to create distributed wrapper: %v", err)
    }
    
    // Start distributed components
    ctx := context.Background()
    if err := wrapper.Start(ctx); err != nil {
        log.Fatalf("Failed to start distributed components: %v", err)
    }
    
    // Generate routes with distributed capabilities
    handler, err := wrapper.GenerateRoutesWithDistributed(nil)
    if err != nil {
        log.Fatalf("Failed to generate routes: %v", err)
    }
    
    // Start server
    ln, err := net.Listen("tcp", ":11434")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    
    log.Println("Server starting with distributed capabilities")
    if err := server.Serve(ln); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
```

## Environment Variables

Configure the distributed behavior with environment variables:

```bash
# Enable distributed mode
export OLLAMA_DISTRIBUTED=true

# Enable standalone mode
export OLLAMA_STANDALONE=true

# Set local Ollama address
export OLLAMA_LOCAL_ADDR=http://localhost:11434

# Enable fallback mode
export OLLAMA_FALLBACK=true

# Set admin token for admin endpoints
export OLLAMA_ADMIN_TOKEN=your-secret-token
```

## API Endpoints

### Original Ollama Endpoints (with distributed capabilities)
- `POST /api/generate` - Generate text with distributed routing
- `POST /api/chat` - Chat completions with distributed routing
- `POST /api/embed` - Embeddings with distributed routing
- `GET /api/tags` - List models (local + distributed)
- `POST /api/pull` - Pull models with distributed management
- `GET /api/ps` - Process status (local + distributed)
- `GET /api/version` - Version with distributed info

### Distributed-specific Endpoints
- `GET /api/distributed/status` - Distributed system status
- `GET /api/distributed/nodes` - List cluster nodes
- `GET /api/distributed/models` - List distributed models
- `POST /api/distributed/rebalance` - Rebalance models
- `POST /api/distributed/migrate` - Migrate models

### Admin Endpoints (require authorization)
- `POST /admin/mode` - Set distributed/local mode
- `POST /admin/fallback` - Enable/disable fallback
- `GET /admin/stats` - Detailed statistics

## Fallback Behavior

The system provides multiple fallback mechanisms:

1. **Local Fallback**: Requests fall back to local Ollama instance
2. **Cached Responses**: Use cached responses for temporary failures
3. **Standalone Mode**: Operate independently when cluster is unavailable

## Headers

The system adds several headers to responses:

- `X-Ollama-Distributed`: Indicates distributed mode
- `X-Ollama-Mode`: Current mode (distributed/local/standalone)
- `X-Ollama-Node`: Node that processed the request
- `X-Ollama-Fallback`: Indicates fallback was used
- `X-Ollama-Fallback-Reason`: Reason for fallback

## Testing

Test the integration:

```bash
# Test distributed generation
curl -X POST http://localhost:11434/api/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2", "prompt": "Hello world"}'

# Check distributed status
curl http://localhost:11434/api/distributed/status

# Test fallback (with local instance stopped)
curl -X POST http://localhost:11434/api/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2", "prompt": "Hello world"}'
```

## Integration Benefits

1. **Backward Compatibility**: Existing clients work without changes
2. **Transparent Distribution**: Automatic routing to best available node
3. **Fault Tolerance**: Multiple fallback mechanisms
4. **Performance**: Load balancing across cluster nodes
5. **Monitoring**: Comprehensive metrics and health checks
6. **Administration**: Runtime configuration changes