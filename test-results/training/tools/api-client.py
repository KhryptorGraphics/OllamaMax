#!/usr/bin/env python3
"""
API Test Client - Training Tool
Comprehensive testing client for Ollama Distributed API
"""

import requests
import json
import time
import sys
from datetime import datetime

class OllamaAPIClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.timeout = 10
        
    def test_endpoint(self, endpoint, method="GET", data=None):
        """Test a single API endpoint"""
        url = f"{self.base_url}{endpoint}"
        timestamp = datetime.now().isoformat()
        
        try:
            if method.upper() == "GET":
                response = self.session.get(url)
            elif method.upper() == "POST":
                response = self.session.post(url, json=data)
            else:
                response = self.session.request(method, url, json=data)
            
            result = {
                "timestamp": timestamp,
                "endpoint": endpoint,
                "method": method,
                "status_code": response.status_code,
                "success": response.status_code < 400,
                "response_time": response.elapsed.total_seconds(),
            }
            
            try:
                result["response_json"] = response.json()
            except:
                result["response_text"] = response.text[:200]
            
            return result
            
        except requests.exceptions.RequestException as e:
            return {
                "timestamp": timestamp,
                "endpoint": endpoint,
                "method": method,
                "success": False,
                "error": str(e)
            }
    
    def run_test_suite(self):
        """Run comprehensive test suite"""
        endpoints = [
            "/health",
            "/api/v1/health", 
            "/api/v1/nodes",
            "/api/v1/models",
            "/api/v1/stats",
        ]
        
        results = []
        for endpoint in endpoints:
            result = self.test_endpoint(endpoint)
            results.append(result)
            print(f"{result['endpoint']}: {result.get('status_code', 'ERROR')} ({'OK' if result['success'] else 'FAIL'})")
        
        return results

if __name__ == "__main__":
    base_url = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8080"
    client = OllamaAPIClient(base_url)
    
    print(f"Testing API at: {base_url}")
    print("=" * 40)
    
    results = client.run_test_suite()
    
    # Save results
    with open("api-test-results.json", "w") as f:
        json.dump(results, f, indent=2)
    
    print("\nResults saved to: api-test-results.json")
