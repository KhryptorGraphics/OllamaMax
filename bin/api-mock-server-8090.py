#!/usr/bin/env python3
"""
Ollama Distributed API Mock Server on Port 8090
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json
from datetime import datetime

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
            data = json.loads(post_data) if post_data else {}
            response = {
                "model": data.get("model", "test"),
                "response": "This is a placeholder response. Distributed inference not yet implemented.",
                "done": True
            }
            self.wfile.write(json.dumps(response).encode())
        else:
            self.send_response(404)
            self.end_headers()
            
    def log_message(self, format, *args):
        pass

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', 8090), APIHandler)
    print("🚀 API Mock Server on port 8090")
    server.serve_forever()