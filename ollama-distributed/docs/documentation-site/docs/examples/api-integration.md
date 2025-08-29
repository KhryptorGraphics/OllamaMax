# REST API Integration

Integrate with OllamaMax using the REST API.

## Basic API Usage

```bash
# Generate text
curl -X POST http://localhost:8081/api/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2", "prompt": "Hello world"}'
```

See the [API Reference](../api/overview.md) for complete documentation.