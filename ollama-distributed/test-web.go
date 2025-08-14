package main

import (
	"log"
	"net/http"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/web"
)

func main() {
	log.Println("Testing web server...")

	// Create web server config
	config := web.DefaultConfig()
	config.ListenAddress = ":8081"
	config.APIBaseURL = "http://localhost:8080"

	log.Printf("Creating web server with config: %+v", config)

	// Create web server (without API server for testing)
	webServer := web.NewWebServer(config, nil)

	log.Println("Starting web server...")

	// Start web server
	go func() {
		if err := webServer.Start(); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	log.Println("Web server started, waiting...")

	// Wait a bit to see if it starts
	time.Sleep(3 * time.Second)

	// Test the connection
	log.Println("Testing connection...")
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		log.Printf("Connection failed: %v", err)
	} else {
		log.Printf("Connection successful: %s", resp.Status)
		resp.Body.Close()
	}

	// Keep running for a bit longer
	time.Sleep(10 * time.Second)

	log.Println("Test complete")
}
