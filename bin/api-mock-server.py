#!/usr/bin/env python3
"""
Ollama Distributed API Mock Server (Minimal)
Provides basic API endpoints for training module validation
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json
from datetime import datetime
import threading

class APIHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {
                "status": "healthy",
                "timestamp": datetime.utcnow().isoformat() + "Z",
                "version": "1.0.0",
                "node_id": "12D3KooW...",
                "services": {
                    "p2p": True,
                    "p2p_peers": 0,
                    "consensus": True,
                    "consensus_leader": False,
                    "scheduler": True,
                    "available_nodes": 1
                }
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/api/distributed/status':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {
                "distributed_mode": True,
                "fallback_mode": True,
                "cluster_size": 1,
                "active_nodes": ["node1"],
                "scheduler_stats": {},
                "runner_stats": {},
                "integration_stats": {}
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/api/distributed/nodes':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {
                "nodes": [
                    {
                        "id": "node1",
                        "status": "active",
                        "address": "127.0.0.1:8080",
                        "models": [],
                        "resources": {
                            "cpu": 0.15,
                            "memory": 0.25,
                            "disk": 0.20
                        }
                    }
                ]
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/api/distributed/metrics':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {
                "timestamp": datetime.utcnow().isoformat() + "Z",
                "node_id": "12D3KooW...",
                "connected_peers": 0,
                "is_leader": False,
                "requests_processed": 0,
                "models_loaded": 0,
                "nodes_total": 1,
                "nodes_online": 1,
                "uptime": 300,
                "websocket_connections": 0
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/api/tags':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {
                "models": [
                    {
                        "name": "llama2:7b",
                        "status": "available",
                        "size": "3.8GB"
                    },
                    {
                        "name": "phi3:mini",
                        "status": "ready",
                        "size": "2.3GB"
                    }
                ]
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/v1/models':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {
                "data": [
                    {
                        "id": "llama2:7b",
                        "object": "model",
                        "created": 1677652288,
                        "owned_by": "ollama"
                    },
                    {
                        "id": "phi3:mini",
                        "object": "model",
                        "created": 1677652288,
                        "owned_by": "ollama"
                    }
                ]
            }
            self.wfile.write(json.dumps(response).encode())
            
        else:
            self.send_response(404)
            self.end_headers()
            
    def do_POST(self):
        content_length = int(self.headers.get('Content-Length', 0))
        post_data = self.rfile.read(content_length)
        
        if self.path == '/api/generate':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            data = {}
            if post_data:
                try:
                    data = json.loads(post_data)
                except:
                    pass
            response = {
                "model": data.get("model", "test"),
                "response": "This is a placeholder response. Distributed inference not yet implemented.",
                "done": True
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/api/chat':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            data = {}
            if post_data:
                try:
                    data = json.loads(post_data)
                except:
                    pass
            response = {
                "model": data.get("model", "test"),
                "message": {
                    "role": "assistant",
                    "content": "This is a placeholder response. Distributed chat inference not yet implemented."
                },
                "done": True
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/api/show':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            data = {}
            if post_data:
                try:
                    data = json.loads(post_data)
                except:
                    pass
            model_name = data.get("name", "test")
            response = {
                "name": model_name,
                "size": "2.3GB",
                "modified_at": datetime.utcnow().isoformat() + "Z",
                "details": {
                    "format": "gguf",
                    "families": ["phi3"] if "phi3" in model_name else ["llama2"]
                }
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/api/embed':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {
                "embedding": [0.1, 0.2, 0.3, 0.4, 0.5] * 100  # Mock 500-dimensional vector
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/v1/chat/completions':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            data = {}
            if post_data:
                try:
                    data = json.loads(post_data)
                except:
                    pass
            response = {
                "id": "chatcmpl-123",
                "object": "chat.completion",
                "created": 1677652288,
                "model": data.get("model", "llama2:7b"),
                "choices": [{
                    "index": 0,
                    "message": {
                        "role": "assistant",
                        "content": "This is a placeholder OpenAI-compatible response."
                    },
                    "finish_reason": "stop"
                }],
                "usage": {
                    "prompt_tokens": 10,
                    "completion_tokens": 20,
                    "total_tokens": 30
                }
            }
            self.wfile.write(json.dumps(response).encode())
            
        elif self.path == '/v1/embeddings':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {
                "object": "list",
                "data": [{
                    "object": "embedding",
                    "embedding": [0.1, 0.2, 0.3, 0.4, 0.5] * 300,  # Mock 1500-dimensional vector
                    "index": 0
                }],
                "model": "text-embedding-ada-002",
                "usage": {
                    "prompt_tokens": 8,
                    "total_tokens": 8
                }
            }
            self.wfile.write(json.dumps(response).encode())
            
        else:
            self.send_response(404)
            self.end_headers()
            
    def log_message(self, format, *args):
        # Suppress log messages for cleaner output
        pass

def start_server(port=8080):
    server = HTTPServer(('0.0.0.0', port), APIHandler)
    print(f"🚀 Starting Ollama Distributed API Mock Server on port {port}")
    print(f"   API endpoint: http://localhost:{port}")
    print(f"   Health check: http://localhost:{port}/health")
    server.serve_forever()

if __name__ == '__main__':
    # Start server in background thread
    server_thread = threading.Thread(target=start_server, daemon=True)
    server_thread.start()
    
    # Keep main thread alive
    try:
        while True:
            pass
    except KeyboardInterrupt:
        print("\nShutting down server...")