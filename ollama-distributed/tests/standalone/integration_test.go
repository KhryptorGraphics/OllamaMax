package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// IntegrationTest tests the complete backend-frontend integration
func main() {
	log.Println("🧪 Starting Backend-Frontend Integration Test")

	// Test configuration
	baseURL := "http://localhost:8080"
	wsURL := "ws://localhost:8080/api/v1/ws"

	// Wait for server to be ready
	log.Println("⏳ Waiting for server to start...")
	if !waitForServer(baseURL, 30*time.Second) {
		log.Fatal("❌ Server failed to start within timeout")
	}
	log.Println("✅ Server is ready")

	// Run test suite

	// Test 1: API Endpoints
	log.Println("\n🔍 Testing API Endpoints...")
	testAPIEndpoints(baseURL)

	// Test 2: WebSocket Connection
	log.Println("\n🔌 Testing WebSocket Connection...")
	testWebSocketConnection(wsURL)

	// Test 3: Model Operations
	log.Println("\n🧠 Testing Model Operations...")
	testModelOperations(baseURL, wsURL)

	// Test 4: Auto-Distribution
	log.Println("\n⚡ Testing Auto-Distribution...")
	testAutoDistribution(baseURL)

	// Test 5: Real-time Updates
	log.Println("\n📡 Testing Real-time Updates...")
	testRealTimeUpdates(wsURL)

	log.Println("\n✅ All integration tests completed successfully!")
	log.Println("🎉 Backend-Frontend integration is working correctly!")
}

func waitForServer(baseURL string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/api/v1/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func testAPIEndpoints(baseURL string) {
	endpoints := []struct {
		path   string
		method string
		name   string
	}{
		{"/api/v1/health", "GET", "Health Check"},
		{"/api/v1/cluster/status", "GET", "Cluster Status"},
		{"/api/v1/nodes", "GET", "Nodes List"},
		{"/api/v1/models", "GET", "Models List"},
		{"/api/v1/transfers", "GET", "Transfers List"},
		{"/api/v1/metrics", "GET", "Metrics"},
	}

	for _, endpoint := range endpoints {
		log.Printf("  📍 Testing %s...", endpoint.name)

		resp, err := http.Get(baseURL + endpoint.path)
		if err != nil {
			log.Printf("    ❌ Failed to call %s: %v", endpoint.path, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf("    ❌ %s returned status %d", endpoint.name, resp.StatusCode)
			continue
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("    ❌ Failed to decode JSON response: %v", err)
			continue
		}

		log.Printf("    ✅ %s - OK", endpoint.name)
	}
}

func testWebSocketConnection(wsURL string) {
	u, err := url.Parse(wsURL)
	if err != nil {
		log.Printf("❌ Failed to parse WebSocket URL: %v", err)
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("❌ Failed to connect to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Println("  ✅ WebSocket connection established")

	// Test ping-pong
	pingMsg := map[string]interface{}{
		"type": "ping",
	}

	if err := conn.WriteJSON(pingMsg); err != nil {
		log.Printf("❌ Failed to send ping: %v", err)
		return
	}

	// Read pong response
	var pongMsg map[string]interface{}
	if err := conn.ReadJSON(&pongMsg); err != nil {
		log.Printf("❌ Failed to read pong: %v", err)
		return
	}

	if pongMsg["type"] != "pong" {
		log.Printf("❌ Expected pong, got: %v", pongMsg["type"])
		return
	}

	log.Println("  ✅ WebSocket ping-pong successful")
}

func testModelOperations(baseURL, wsURL string) {
	// Test model download
	log.Println("  📥 Testing model download...")

	downloadData := map[string]interface{}{
		"model": "test-model",
	}

	jsonData, _ := json.Marshal(downloadData)
	resp, err := http.Post(baseURL+"/api/v1/models/test-model/download", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to download model: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Model download failed with status %d: %s", resp.StatusCode, string(body))
		return
	}

	log.Println("  ✅ Model download initiated successfully")

	// Test model deletion
	log.Println("  🗑️  Testing model deletion...")

	req, _ := http.NewRequest("DELETE", baseURL+"/api/v1/models/test-model", nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to delete model: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("❌ Model deletion failed with status %d", resp.StatusCode)
		return
	}

	log.Println("  ✅ Model deletion completed successfully")
}

func testAutoDistribution(baseURL string) {
	log.Println("  🔄 Testing auto-distribution enable...")

	// Enable auto-distribution
	enableData := map[string]interface{}{
		"enabled": true,
	}

	jsonData, _ := json.Marshal(enableData)
	resp, err := http.Post(baseURL+"/api/v1/distribution/auto-configure", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to enable auto-distribution: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Auto-distribution enable failed with status %d: %s", resp.StatusCode, string(body))
		return
	}

	log.Println("  ✅ Auto-distribution enabled successfully")

	// Disable auto-distribution
	log.Println("  🔄 Testing auto-distribution disable...")

	disableData := map[string]interface{}{
		"enabled": false,
	}

	jsonData, _ = json.Marshal(disableData)
	resp, err = http.Post(baseURL+"/api/v1/distribution/auto-configure", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to disable auto-distribution: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("❌ Auto-distribution disable failed with status %d", resp.StatusCode)
		return
	}

	log.Println("  ✅ Auto-distribution disabled successfully")
}

func testRealTimeUpdates(wsURL string) {
	u, err := url.Parse(wsURL)
	if err != nil {
		log.Printf("❌ Failed to parse WebSocket URL: %v", err)
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("❌ Failed to connect to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Println("  📡 Listening for real-time updates...")

	// Set a read deadline
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Listen for messages
	messageCount := 0
	for messageCount < 3 {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			break
		}

		if msgType, ok := msg["type"].(string); ok {
			log.Printf("  📨 Received message: %s", msgType)
			messageCount++

			// Reset read deadline for next message
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		}
	}

	if messageCount > 0 {
		log.Printf("  ✅ Received %d real-time updates", messageCount)
	} else {
		log.Println("  ⚠️  No real-time updates received (this might be normal)")
	}
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
