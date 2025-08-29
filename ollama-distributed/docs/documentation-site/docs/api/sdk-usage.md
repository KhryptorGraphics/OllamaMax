# SDK Usage

Learn how to use the official Ollama Distributed SDKs in various programming languages.

## Go SDK

### Installation
```bash
go get github.com/ollama/ollama-distributed-go
```

### Usage
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ollama/ollama-distributed-go"
)

func main() {
    client := ollama.NewClient("https://api.ollama-distributed.example.com", "YOUR_API_KEY")

    response, err := client.Generate(context.Background(), &ollama.GenerateRequest{
        Model:  "llama2",
        Prompt: "Explain distributed systems",
        Stream: false,
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response.Response)
}
```

## Python SDK

### Installation
```bash
pip install ollama-distributed
```

### Usage
```python
from ollama_distributed import Client

client = Client(
    base_url="https://api.ollama-distributed.example.com",
    api_key="YOUR_API_KEY"
)

response = client.generate(
    model="llama2",
    prompt="What is artificial intelligence?",
    stream=False
)

print(response.text)
```

### Async Usage
```python
import asyncio
from ollama_distributed import AsyncClient

async def main():
    client = AsyncClient(
        base_url="https://api.ollama-distributed.example.com",
        api_key="YOUR_API_KEY"
    )

    response = await client.generate(
        model="llama2",
        prompt="Explain async programming",
        stream=False
    )

    print(response.text)

asyncio.run(main())
```

## JavaScript/TypeScript SDK

### Installation
```bash
npm install ollama-distributed-js
```

### Usage
```typescript
import { OllamaClient } from 'ollama-distributed-js';

const client = new OllamaClient({
  baseUrl: 'https://api.ollama-distributed.example.com',
  apiKey: 'YOUR_API_KEY'
});

async function generateText() {
  const response = await client.generate({
    model: 'llama2',
    prompt: 'Explain TypeScript benefits',
    stream: false
  });

  console.log(response.text);
}

generateText();
```

### Streaming
```typescript
const stream = await client.generateStream({
  model: 'llama2',
  prompt: 'Tell me about web development'
});

for await (const chunk of stream) {
  process.stdout.write(chunk.text);
}
```

## Error Handling

### Go
```go
response, err := client.Generate(ctx, req)
if err != nil {
    if apiErr, ok := err.(*ollama.APIError); ok {
        fmt.Printf("API Error: %s (Code: %d)\n", apiErr.Message, apiErr.Code)
    } else {
        fmt.Printf("Network Error: %v\n", err)
    }
}
```

### Python
```python
try:
    response = client.generate(model="llama2", prompt="Hello")
except ollama.APIError as e:
    print(f"API Error: {e.message} (Code: {e.code})")
except ollama.NetworkError as e:
    print(f"Network Error: {e}")
```

### JavaScript
```typescript
try {
  const response = await client.generate({
    model: 'llama2',
    prompt: 'Hello'
  });
} catch (error) {
  if (error instanceof APIError) {
    console.log(`API Error: ${error.message} (Code: ${error.code})`);
  } else {
    console.log(`Network Error: ${error.message}`);
  }
}
```