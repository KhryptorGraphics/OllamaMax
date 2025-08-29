# Multi-Model Deployment

Deploy and manage multiple AI models simultaneously.

## Multi-Model Setup

```bash
# Deploy multiple models
ollama-distributed proxy pull llama2
ollama-distributed proxy pull phi3
ollama-distributed proxy pull codellama

# Check model status
ollama-distributed proxy list
```