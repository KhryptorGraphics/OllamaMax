/**
 * API Endpoints Tests
 * Tests for OpenAI-compatible API endpoints
 */

const request = require('supertest');
const { app } = require('../src/server');

describe('API Endpoints', () => {
  let authToken;
  const testUser = {
    email: 'api-test@ollamamax.com',
    password: 'TestPass123!'
  };

  beforeAll(async () => {
    // Wait for database initialization
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Register and login user for API tests
    await request(app)
      .post('/auth/register')
      .send(testUser);

    const loginResponse = await request(app)
      .post('/auth/login')
      .send(testUser);

    authToken = loginResponse.body.access_token;
  });

  describe('Root Endpoint', () => {
    it('should return API information', async () => {
      const response = await request(app)
        .get('/')
        .expect(200);

      expect(response.body).toMatchObject({
        name: 'Ollamamax API',
        version: '1.0.0',
        status: 'running',
        endpoints: {
          authentication: '/auth',
          inference: '/v1',
          health: '/health',
          docs: '/docs'
        },
        features: {
          authentication: true,
          rate_limiting: true,
          openai_compatibility: true,
          streaming: true
        }
      });
    });
  });

  describe('Health Endpoints', () => {
    it('should return health status', async () => {
      const response = await request(app)
        .get('/health')
        .expect(200);

      expect(response.body).toMatchObject({
        status: 'healthy',
        uptime: expect.any(Object),
        memory: expect.any(Object)
      });

      expect(response.body.uptime.seconds).toBeGreaterThan(0);
      expect(response.body.memory.rss).toBeTruthy();
    });

    it('should return liveness probe', async () => {
      const response = await request(app)
        .get('/health/live')
        .expect(200);

      expect(response.body.status).toBe('alive');
    });

    it('should return readiness probe', async () => {
      const response = await request(app)
        .get('/health/ready')
        .expect(200);

      expect(response.body.status).toBe('ready');
      expect(response.body.checks).toBeDefined();
      expect(response.body.checks.database).toBe(true);
    });
  });

  describe('Metrics Endpoint', () => {
    it('should return Prometheus metrics', async () => {
      const response = await request(app)
        .get('/metrics')
        .expect(200);

      expect(response.headers['content-type']).toBe('text/plain; charset=utf-8');
      expect(response.text).toContain('ollamamax_uptime_seconds');
      expect(response.text).toContain('ollamamax_memory_usage_bytes');
    });
  });

  describe('Models Endpoint', () => {
    it('should list available models without authentication', async () => {
      const response = await request(app)
        .get('/v1/models')
        .expect(200);

      expect(response.body).toMatchObject({
        object: 'list',
        data: expect.any(Array)
      });

      expect(response.body.data.length).toBeGreaterThan(0);
      expect(response.body.data[0]).toMatchObject({
        id: expect.any(String),
        object: 'model',
        owned_by: 'ollamamax'
      });
    });

    it('should include OpenAI compatibility models', async () => {
      const response = await request(app)
        .get('/v1/models')
        .expect(200);

      const modelIds = response.body.data.map(model => model.id);
      expect(modelIds).toContain('gpt-3.5-turbo');
      expect(modelIds).toContain('llama-3.2-3b');
    });
  });

  describe('Text Completions', () => {
    it('should create text completion with authentication', async () => {
      const response = await request(app)
        .post('/v1/completions')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          model: 'llama-3.2-3b',
          prompt: 'Hello, world!',
          max_tokens: 50,
          temperature: 0.7
        })
        .expect(200);

      expect(response.body).toMatchObject({
        id: expect.stringMatching(/^cmpl-/),
        object: 'text_completion',
        model: 'llama-3.2-3b',
        choices: expect.any(Array),
        usage: {
          prompt_tokens: expect.any(Number),
          completion_tokens: expect.any(Number),
          total_tokens: expect.any(Number)
        }
      });

      expect(response.body.choices[0]).toMatchObject({
        text: expect.any(String),
        index: 0,
        finish_reason: 'stop'
      });
    });

    it('should reject completion without authentication', async () => {
      const response = await request(app)
        .post('/v1/completions')
        .send({
          model: 'llama-3.2-3b',
          prompt: 'Hello, world!'
        })
        .expect(401);

      expect(response.body.error.code).toBe('missing_auth');
    });

    it('should validate required parameters', async () => {
      const response = await request(app)
        .post('/v1/completions')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          model: 'llama-3.2-3b'
          // Missing prompt
        })
        .expect(400);

      expect(response.body.error.param).toBe('prompt');
    });

    it('should handle streaming completions', async () => {
      const response = await request(app)
        .post('/v1/completions')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          model: 'llama-3.2-3b',
          prompt: 'Hello, world!',
          max_tokens: 10,
          stream: true
        })
        .expect(200);

      expect(response.headers['content-type']).toBe('text/event-stream; charset=utf-8');
      expect(response.text).toContain('data:');
      expect(response.text).toContain('[DONE]');
    });
  });

  describe('Chat Completions', () => {
    it('should create chat completion', async () => {
      const response = await request(app)
        .post('/v1/chat/completions')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          model: 'llama-3.2-3b',
          messages: [
            { role: 'system', content: 'You are a helpful assistant.' },
            { role: 'user', content: 'Hello!' }
          ],
          max_tokens: 50
        })
        .expect(200);

      expect(response.body).toMatchObject({
        id: expect.stringMatching(/^chatcmpl-/),
        object: 'chat.completion',
        model: 'llama-3.2-3b',
        choices: expect.any(Array),
        usage: expect.any(Object)
      });

      expect(response.body.choices[0]).toMatchObject({
        index: 0,
        message: {
          role: 'assistant',
          content: expect.any(String)
        },
        finish_reason: 'stop'
      });
    });

    it('should validate messages parameter', async () => {
      const response = await request(app)
        .post('/v1/chat/completions')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          model: 'llama-3.2-3b'
          // Missing messages
        })
        .expect(400);

      expect(response.body.error.param).toBe('messages');
    });

    it('should handle streaming chat completions', async () => {
      const response = await request(app)
        .post('/v1/chat/completions')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          model: 'llama-3.2-3b',
          messages: [
            { role: 'user', content: 'Hello!' }
          ],
          max_tokens: 10,
          stream: true
        })
        .expect(200);

      expect(response.headers['content-type']).toBe('text/event-stream; charset=utf-8');
      expect(response.text).toContain('data:');
      expect(response.text).toContain('[DONE]');
    });
  });

  describe('Embeddings', () => {
    it('should create embeddings', async () => {
      const response = await request(app)
        .post('/v1/embeddings')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          model: 'text-embedding-ada-002',
          input: 'The quick brown fox jumps over the lazy dog'
        })
        .expect(200);

      expect(response.body).toMatchObject({
        object: 'list',
        data: expect.any(Array),
        model: 'text-embedding-ada-002',
        usage: expect.any(Object)
      });

      expect(response.body.data[0]).toMatchObject({
        object: 'embedding',
        embedding: expect.any(Array),
        index: 0
      });

      expect(response.body.data[0].embedding).toHaveLength(1536); // Default dimensions
    });

    it('should handle array of inputs', async () => {
      const response = await request(app)
        .post('/v1/embeddings')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          model: 'text-embedding-ada-002',
          input: ['First text', 'Second text', 'Third text']
        })
        .expect(200);

      expect(response.body.data).toHaveLength(3);
      expect(response.body.data[0].index).toBe(0);
      expect(response.body.data[1].index).toBe(1);
      expect(response.body.data[2].index).toBe(2);
    });
  });

  describe('API Documentation', () => {
    it('should serve Swagger UI documentation', async () => {
      const response = await request(app)
        .get('/docs')
        .expect(200);

      expect(response.headers['content-type']).toContain('text/html');
      expect(response.text).toContain('swagger-ui');
      expect(response.text).toContain('Ollamamax API Documentation');
    });

    it('should serve OpenAPI specification', async () => {
      const response = await request(app)
        .get('/openapi.json')
        .expect(200);

      expect(response.body).toMatchObject({
        openapi: '3.0.0',
        info: {
          title: 'Ollamamax API',
          version: '1.0.0'
        },
        paths: expect.any(Object),
        components: expect.any(Object)
      });

      expect(response.body.paths['/v1/models']).toBeDefined();
      expect(response.body.paths['/v1/completions']).toBeDefined();
      expect(response.body.paths['/v1/chat/completions']).toBeDefined();
    });
  });

  describe('Error Handling', () => {
    it('should return 404 for non-existent endpoints', async () => {
      const response = await request(app)
        .get('/non-existent-endpoint')
        .expect(404);

      expect(response.body.error).toMatchObject({
        message: 'Not found',
        type: 'not_found_error',
        code: 'endpoint_not_found'
      });
    });

    it('should handle malformed JSON', async () => {
      const response = await request(app)
        .post('/v1/completions')
        .set('Authorization', `Bearer ${authToken}`)
        .set('Content-Type', 'application/json')
        .send('{ invalid json }')
        .expect(400);

      // Express handles this automatically
    });
  });

  describe('Rate Limiting', () => {
    it('should apply rate limits to API endpoints', async () => {
      // This test checks if rate limiting headers are present
      const response = await request(app)
        .get('/v1/models')
        .expect(200);

      // Rate limiting headers should be present
      expect(response.headers['x-ratelimit-limit']).toBeDefined();
      expect(response.headers['x-ratelimit-remaining']).toBeDefined();
    });
  });
});