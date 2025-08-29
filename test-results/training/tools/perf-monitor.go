package main

import (
	"fmt"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type PerformanceMetrics struct {
	ResponseTime    time.Duration `json:"response_time_ms"`
	StatusCode      int          `json:"status_code"`
	Success         bool         `json:"success"`
	Timestamp       time.Time    `json:"timestamp"`
}

func main() {
	baseURL := "http://localhost:8080"
	endpoint := "/health"
	
	fmt.Printf("Performance Monitor for %s%s\n", baseURL, endpoint)
	fmt.Println("Collecting metrics (Ctrl+C to stop)...")
	
	client := &http.Client{Timeout: 5 * time.Second}
	
	for {
		metrics := testEndpoint(client, baseURL+endpoint)
		
		fmt.Printf("[%s] Status: %d, Time: %dms, Success: %v\n",
			metrics.Timestamp.Format("15:04:05"),
			metrics.StatusCode,
			metrics.ResponseTime.Milliseconds(),
			metrics.Success)
		
		time.Sleep(2 * time.Second)
	}
}

func testEndpoint(client *http.Client, url string) PerformanceMetrics {
	start := time.Now()
	
	resp, err := client.Get(url)
	elapsed := time.Since(start)
	
	metrics := PerformanceMetrics{
		ResponseTime: elapsed,
		Timestamp:    time.Now(),
	}
	
	if err != nil {
		metrics.StatusCode = 0
		metrics.Success = false
		return metrics
	}
	defer resp.Body.Close()
	
	metrics.StatusCode = resp.StatusCode
	metrics.Success = resp.StatusCode < 400
	
	return metrics
}
