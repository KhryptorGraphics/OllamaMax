# API Overview

Welcome to the Ollama Distributed API documentation. This section provides comprehensive information about using our REST API and WebSocket interfaces.

## Base URL

```
https://api.ollama-distributed.example.com/v1
```

## Authentication

All API requests require authentication using Bearer tokens. See the [Authentication](./authentication.md) section for details.

## Rate Limiting

API requests are rate-limited to ensure fair usage and system stability:

- **Standard users**: 1000 requests per hour
- **Premium users**: 10000 requests per hour
- **Enterprise users**: Unlimited

## Response Format

All API responses use JSON format with consistent error handling:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "timestamp": "2025-08-27T19:42:00Z"
}
```

## Quick Start

1. Obtain an API key from the dashboard
2. Include the key in the Authorization header
3. Make your first API call

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.ollama-distributed.example.com/v1/models
```