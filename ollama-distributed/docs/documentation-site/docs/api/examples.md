# API Examples

Practical examples for integrating with the Ollama Distributed API.

## Basic Examples

### Generate Text (cURL)
```bash
curl -X POST https://api.ollama-distributed.example.com/v1/generate \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "prompt": "Explain quantum computing",
    "stream": false
  }'
```

### Python Example
```python
import requests

api_key = "YOUR_API_KEY"
url = "https://api.ollama-distributed.example.com/v1/generate"

headers = {
    "Authorization": f"Bearer {api_key}",
    "Content-Type": "application/json"
}

data = {
    "model": "llama2",
    "prompt": "Write a haiku about programming",
    "stream": False
}

response = requests.post(url, json=data, headers=headers)
result = response.json()

if result["success"]:
    print(result["data"]["response"])
```

### JavaScript/Node.js Example
```javascript
const fetch = require('node-fetch');

async function generateText(prompt) {
  const response = await fetch('https://api.ollama-distributed.example.com/v1/generate', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      model: 'llama2',
      prompt: prompt,
      stream: false
    })
  });

  const data = await response.json();
  return data.success ? data.data.response : null;
}

generateText("Explain machine learning").then(console.log);
```

## Streaming Examples

### Streaming with cURL
```bash
curl -X POST https://api.ollama-distributed.example.com/v1/generate \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "prompt": "Tell me a story",
    "stream": true
  }' \
  --no-buffer
```

### Streaming with Python
```python
import requests
import json

def stream_generate(prompt):
    url = "https://api.ollama-distributed.example.com/v1/generate"
    headers = {
        "Authorization": "Bearer YOUR_API_KEY",
        "Content-Type": "application/json"
    }
    data = {
        "model": "llama2",
        "prompt": prompt,
        "stream": True
    }

    response = requests.post(url, json=data, headers=headers, stream=True)
    
    for line in response.iter_lines():
        if line:
            chunk = json.loads(line)
            if chunk.get("success") and chunk["data"].get("response"):
                print(chunk["data"]["response"], end="", flush=True)

stream_generate("Write a poem about the ocean")
```