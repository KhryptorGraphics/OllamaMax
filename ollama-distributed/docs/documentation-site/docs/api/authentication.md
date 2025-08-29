# Authentication

Learn how to authenticate with the Ollama Distributed API using various methods.

## API Keys

The primary method for authenticating with the API is using API keys.

### Obtaining an API Key

1. Log in to your dashboard
2. Navigate to Settings > API Keys
3. Click "Generate New Key"
4. Copy and store your key securely

### Using API Keys

Include your API key in the Authorization header:

```bash
curl -H "Authorization: Bearer sk-your-api-key-here" \
  https://api.ollama-distributed.example.com/v1/models
```

## JWT Tokens

For web applications, you can use JWT tokens obtained through the authentication flow.

```javascript
// Login and get JWT token
const response = await fetch('/api/auth/login', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({username: 'user', password: 'pass'})
});

const {token} = await response.json();

// Use token in subsequent requests
const modelsResponse = await fetch('/api/v1/models', {
  headers: {'Authorization': `Bearer ${token}`}
});
```

## Security Best Practices

- Never expose API keys in client-side code
- Rotate keys regularly
- Use environment variables for key storage
- Implement proper key management in production