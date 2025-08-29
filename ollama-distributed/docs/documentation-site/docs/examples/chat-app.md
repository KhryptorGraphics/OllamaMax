# Simple Chat Application

Build a simple chat application using OllamaMax.

## Overview

This example demonstrates how to create a basic chat application that connects to your OllamaMax cluster.

## Implementation

```javascript
// Simple chat app example
const API_URL = 'http://localhost:8081/api';

async function sendMessage(message) {
  const response = await fetch(`${API_URL}/generate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      model: 'llama2',
      prompt: message,
      stream: false
    })
  });
  return response.json();
}
```

For more examples, see the [Getting Started Guide](../getting-started.md).