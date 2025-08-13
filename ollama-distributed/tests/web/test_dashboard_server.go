//go:build ignore

package web_tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Mock data structures
type ClusterStatus struct {
	NodeID    string `json:"node_id"`
	Leader    string `json:"leader"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type Node struct {
	ID      string  `json:"id"`
	Address string  `json:"address"`
	Status  string  `json:"status"`
	Usage   *Usage  `json:"usage,omitempty"`
	Models  []Model `json:"models,omitempty"`
}

type Usage struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Disk   float64 `json:"disk"`
}

type Model struct {
	Name           string `json:"name"`
	Size           int64  `json:"size"`
	Status         string `json:"status"`
	InferenceReady bool   `json:"inference_ready"`
}

type SecurityStatus struct {
	OverallScore int    `json:"overallScore"`
	Compliance   string `json:"compliance"`
}

type PerformanceMetrics struct {
	OverallScore  int     `json:"overallScore"`
	CacheHitRatio int     `json:"cacheHitRatio"`
	CPUUsage      float64 `json:"cpuUsage"`
	MemoryUsage   float64 `json:"memoryUsage"`
}

type Optimizations struct {
	TotalOptimizations int `json:"totalOptimizations"`
}

func TestDashboardServer() {
	fmt.Println("Starting Web Dashboard Test Server...")

	// Get the project root directory
	projectRoot := "/home/kp/ollamamax/ollama-distributed"
	webDir := filepath.Join(projectRoot, "web")

	// Setup HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
			return
		}
		http.NotFound(w, r)
	})

	// API endpoints
	http.HandleFunc("/api/v1/cluster/status", func(w http.ResponseWriter, r *http.Request) {
		status := ClusterStatus{
			NodeID:    "node-12345",
			Leader:    "node-12345",
			Status:    "healthy",
			Timestamp: time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	http.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, r *http.Request) {
		nodes := map[string]interface{}{
			"nodes": map[string]Node{
				"node-1": {
					ID:      "node-1",
					Address: "192.168.1.100:8080",
					Status:  "online",
					Usage: &Usage{
						CPU:    45.2,
						Memory: 68.5,
						Disk:   23.1,
					},
					Models: []Model{
						{Name: "llama2:7b", Size: 3800000000, Status: "available", InferenceReady: true},
						{Name: "codellama:13b", Size: 7300000000, Status: "loading", InferenceReady: false},
					},
				},
				"node-2": {
					ID:      "node-2",
					Address: "192.168.1.101:8080",
					Status:  "online",
					Usage: &Usage{
						CPU:    32.8,
						Memory: 54.2,
						Disk:   18.7,
					},
					Models: []Model{
						{Name: "mistral:7b", Size: 4100000000, Status: "available", InferenceReady: true},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodes)
	})

	http.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
		models := map[string]interface{}{
			"models": map[string]Model{
				"llama2:7b": {
					Name:           "llama2:7b",
					Size:           3800000000,
					Status:         "available",
					InferenceReady: true,
				},
				"codellama:13b": {
					Name:           "codellama:13b",
					Size:           7300000000,
					Status:         "loading",
					InferenceReady: false,
				},
				"mistral:7b": {
					Name:           "mistral:7b",
					Size:           4100000000,
					Status:         "available",
					InferenceReady: true,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models)
	})

	http.HandleFunc("/api/v1/transfers", func(w http.ResponseWriter, r *http.Request) {
		transfers := map[string]interface{}{
			"transfers": map[string]interface{}{
				"transfer-1": map[string]interface{}{
					"id":       "transfer-1",
					"model":    "llama2:7b",
					"from":     "node-1",
					"to":       "node-2",
					"progress": 75.5,
					"status":   "in_progress",
					"speed":    "15.2 MB/s",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transfers)
	})

	http.HandleFunc("/api/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := map[string]interface{}{
			"totalRequests": 1250,
			"avgLatency":    85.3,
			"cpu_usage":     0.42,
			"memory_usage":  0.68,
			"network_usage": 0.23,
			"timestamp":     time.Now().Unix(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	})

	// Security endpoints
	http.HandleFunc("/api/v1/security/status", func(w http.ResponseWriter, r *http.Request) {
		status := SecurityStatus{
			OverallScore: 85,
			Compliance:   "CIS",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	http.HandleFunc("/api/v1/security/threats", func(w http.ResponseWriter, r *http.Request) {
		threats := []interface{}{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(threats)
	})

	http.HandleFunc("/api/v1/security/alerts", func(w http.ResponseWriter, r *http.Request) {
		alerts := []interface{}{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alerts)
	})

	// Performance endpoints
	http.HandleFunc("/api/v1/performance/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := PerformanceMetrics{
			OverallScore:  92,
			CacheHitRatio: 85,
			CPUUsage:      32.5,
			MemoryUsage:   68.2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	})

	http.HandleFunc("/api/v1/performance/optimizations", func(w http.ResponseWriter, r *http.Request) {
		optimizations := Optimizations{
			TotalOptimizations: 15,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(optimizations)
	})

	http.HandleFunc("/api/v1/performance/bottlenecks", func(w http.ResponseWriter, r *http.Request) {
		bottlenecks := []interface{}{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bottlenecks)
	})

	// WebSocket endpoint (mock)
	http.HandleFunc("/api/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("WebSocket upgrade required"))
	})

	// Check if web directory exists
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		log.Fatalf("Web directory not found: %s", webDir)
	}

	// Check if index.html exists
	indexPath := filepath.Join(webDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		log.Fatalf("index.html not found: %s", indexPath)
	}

	port := ":12925"
	fmt.Printf("✅ Web Dashboard Test Server starting on http://localhost%s\n", port)
	fmt.Printf("📁 Serving files from: %s\n", webDir)
	fmt.Printf("🌐 Dashboard URL: http://localhost%s\n", port)
	fmt.Println("\n📋 Available API endpoints:")
	fmt.Println("  - GET /api/v1/cluster/status")
	fmt.Println("  - GET /api/v1/nodes")
	fmt.Println("  - GET /api/v1/models")
	fmt.Println("  - GET /api/v1/transfers")
	fmt.Println("  - GET /api/v1/metrics")
	fmt.Println("  - GET /api/v1/security/status")
	fmt.Println("  - GET /api/v1/security/threats")
	fmt.Println("  - GET /api/v1/security/alerts")
	fmt.Println("  - GET /api/v1/performance/metrics")
	fmt.Println("  - GET /api/v1/performance/optimizations")
	fmt.Println("  - GET /api/v1/performance/bottlenecks")
	fmt.Println("\n🚀 Server ready! Open http://localhost:12925 in your browser")
	fmt.Println("Press Ctrl+C to stop the server")

	log.Fatal(http.ListenAndServe(port, nil))
}
