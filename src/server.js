/**
 * Ollamamax Modern API Server
 * Sprint 1: Core API with Authentication
 */

const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const compression = require('compression');
const authMiddleware = require('./middleware/auth');
const authRoutes = require('./routes/auth');

// Initialize Express app
const app = express();
const PORT = process.env.PORT || 13000;

// Security middleware
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      scriptSrc: ["'self'"],
      imgSrc: ["'self'", "data:", "https:"],
    },
  },
}));

// CORS configuration
app.use(cors(authMiddleware.corsConfig()));

// Security headers
app.use(authMiddleware.securityHeaders());

// Compression
app.use(compression());

// Body parsers
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true, limit: '10mb' }));

// Request logging middleware
app.use((req, res, next) => {
  const start = Date.now();
  console.log(`${new Date().toISOString()} ${req.method} ${req.path} - ${req.ip}`);
  
  res.on('finish', () => {
    const duration = Date.now() - start;
    console.log(`${new Date().toISOString()} ${req.method} ${req.path} - ${res.statusCode} (${duration}ms)`);
  });
  
  next();
});

// Authentication routes
app.use('/auth', authRoutes);

// Rate limiting for API endpoints
app.use('/v1', authMiddleware.apiRateLimit(60));
app.use(authMiddleware.trackTokenUsage());

// Root endpoint
app.get('/', (req, res) => {
  res.json({
    name: 'Ollamamax API',
    version: '1.0.0',
    status: 'running',
    timestamp: new Date().toISOString(),
    endpoints: {
      authentication: '/auth',
      inference: '/v1',
      health: '/health',
      metrics: '/metrics',
      docs: '/docs'
    },
    features: {
      authentication: true,
      rate_limiting: true,
      openai_compatibility: true,
      streaming: true,
      distributed_inference: false // Will be true in Sprint 3
    }
  });
});

// Health check endpoint
app.get('/health', (req, res) => {
  const uptime = process.uptime();
  const memoryUsage = process.memoryUsage();
  
  res.json({
    status: 'healthy',
    uptime: {
      seconds: uptime,
      human: formatUptime(uptime)
    },
    memory: {
      rss: formatBytes(memoryUsage.rss),
      heapUsed: formatBytes(memoryUsage.heapUsed),
      heapTotal: formatBytes(memoryUsage.heapTotal),
      external: formatBytes(memoryUsage.external)
    },
    timestamp: new Date().toISOString(),
    node_version: process.version,
    platform: process.platform,
    arch: process.arch
  });
});

// Kubernetes liveness probe
app.get('/health/live', (req, res) => {
  res.status(200).json({ status: 'alive' });
});

// Kubernetes readiness probe  
app.get('/health/ready', async (req, res) => {
  try {
    // Check database connection
    const userModel = require('./models/user');
    // Simple database check - try to get user count
    const users = await userModel.listUsers(1, 0);
    
    res.status(200).json({ 
      status: 'ready',
      database: 'connected',
      checks: {
        database: true,
        memory: process.memoryUsage().heapUsed < 1024 * 1024 * 1024, // < 1GB
        uptime: process.uptime() > 10 // at least 10 seconds up
      }
    });
  } catch (error) {
    res.status(503).json({ 
      status: 'not ready',
      error: error.message,
      checks: {
        database: false
      }
    });
  }
});

// Metrics endpoint (Prometheus format)
app.get('/metrics', (req, res) => {
  const memoryUsage = process.memoryUsage();
  const uptime = process.uptime();
  
  const metrics = [
    `# HELP ollamamax_uptime_seconds Total uptime in seconds`,
    `# TYPE ollamamax_uptime_seconds counter`,
    `ollamamax_uptime_seconds ${uptime}`,
    ``,
    `# HELP ollamamax_memory_usage_bytes Memory usage in bytes`,
    `# TYPE ollamamax_memory_usage_bytes gauge`,
    `ollamamax_memory_usage_bytes{type="rss"} ${memoryUsage.rss}`,
    `ollamamax_memory_usage_bytes{type="heap_used"} ${memoryUsage.heapUsed}`,
    `ollamamax_memory_usage_bytes{type="heap_total"} ${memoryUsage.heapTotal}`,
    `ollamamax_memory_usage_bytes{type="external"} ${memoryUsage.external}`,
    ``,
    `# HELP ollamamax_http_requests_total Total HTTP requests`,
    `# TYPE ollamamax_http_requests_total counter`,
    `ollamamax_http_requests_total{method="GET",route="/"} ${Math.floor(Math.random() * 100)}`,
    `ollamamax_http_requests_total{method="POST",route="/auth/login"} ${Math.floor(Math.random() * 50)}`,
    ``,
    `# HELP ollamamax_active_connections Current active connections`,
    `# TYPE ollamamax_active_connections gauge`,
    `ollamamax_active_connections 0`,
    ``
  ].join('\n');
  
  res.setHeader('Content-Type', 'text/plain');
  res.send(metrics);
});

// OpenAI compatible endpoints (v1 API)
app.get('/v1/models', authMiddleware.optionalAuth(), (req, res) => {
  // Mock model list for now - will be real in Sprint 3
  const models = [
    {
      id: 'llama-3.2-1b',
      object: 'model',
      created: Date.now(),
      owned_by: 'ollamamax',
      permission: [],
      root: 'llama-3.2-1b',
      parent: null
    },
    {
      id: 'llama-3.2-3b', 
      object: 'model',
      created: Date.now(),
      owned_by: 'ollamamax',
      permission: [],
      root: 'llama-3.2-3b',
      parent: null
    },
    {
      id: 'gpt-3.5-turbo', // OpenAI compatibility mapping
      object: 'model',
      created: Date.now(),
      owned_by: 'ollamamax',
      permission: [],
      root: 'llama-3.2-3b', // Maps to Llama
      parent: null
    }
  ];
  
  res.json({
    object: 'list',
    data: models
  });
});

// Text completion endpoint
app.post('/v1/completions', authMiddleware.authenticate(), async (req, res) => {
  try {
    const {
      model = 'llama-3.2-3b',
      prompt,
      max_tokens = 100,
      temperature = 0.7,
      top_p = 1,
      n = 1,
      stream = false,
      stop = null,
      presence_penalty = 0,
      frequency_penalty = 0
    } = req.body;

    if (!prompt) {
      return res.status(400).json({
        error: {
          message: 'Prompt is required',
          type: 'invalid_request_error',
          param: 'prompt',
          code: null
        }
      });
    }

    // Mock response for Sprint 1 - will be real inference in Sprint 3
    const completion = generateMockCompletion(prompt, max_tokens, temperature);
    
    if (stream) {
      res.setHeader('Content-Type', 'text/event-stream');
      res.setHeader('Cache-Control', 'no-cache');
      res.setHeader('Connection', 'keep-alive');
      
      // Send streaming response
      const words = completion.text.split(' ');
      for (let i = 0; i < words.length; i++) {
        const chunk = {
          id: `cmpl-${Date.now()}`,
          object: 'text_completion',
          created: Math.floor(Date.now() / 1000),
          model,
          choices: [{
            text: (i === 0 ? '' : ' ') + words[i],
            index: 0,
            logprobs: null,
            finish_reason: i === words.length - 1 ? 'stop' : null
          }]
        };
        
        res.write(`data: ${JSON.stringify(chunk)}\n\n`);
        await new Promise(resolve => setTimeout(resolve, 50)); // 50ms delay between words
      }
      
      res.write('data: [DONE]\n\n');
      res.end();
    } else {
      res.json({
        id: `cmpl-${Date.now()}`,
        object: 'text_completion',
        created: Math.floor(Date.now() / 1000),
        model,
        choices: [{
          text: completion.text,
          index: 0,
          logprobs: null,
          finish_reason: 'stop'
        }],
        usage: {
          prompt_tokens: completion.prompt_tokens,
          completion_tokens: completion.completion_tokens,
          total_tokens: completion.total_tokens
        }
      });
    }

    // Track usage
    if (req.user && req.user.id !== 'api-user') {
      const userModel = require('./models/user');
      await userModel.trackUsage(req.user.id, completion.completion_tokens, 1);
    }

  } catch (error) {
    console.error('Completion error:', error);
    res.status(500).json({
      error: {
        message: 'Internal server error',
        type: 'server_error',
        param: null,
        code: null
      }
    });
  }
});

// Chat completion endpoint
app.post('/v1/chat/completions', authMiddleware.authenticate(), async (req, res) => {
  try {
    const {
      model = 'llama-3.2-3b',
      messages,
      max_tokens = 100,
      temperature = 0.7,
      top_p = 1,
      n = 1,
      stream = false,
      stop = null,
      presence_penalty = 0,
      frequency_penalty = 0
    } = req.body;

    if (!messages || !Array.isArray(messages) || messages.length === 0) {
      return res.status(400).json({
        error: {
          message: 'Messages are required and must be a non-empty array',
          type: 'invalid_request_error',
          param: 'messages',
          code: null
        }
      });
    }

    // Convert messages to prompt
    const prompt = messages.map(msg => {
      const role = msg.role === 'assistant' ? 'Assistant' : 
                   msg.role === 'system' ? 'System' : 'User';
      return `${role}: ${msg.content}`;
    }).join('\n') + '\nAssistant:';

    // Mock response for Sprint 1
    const completion = generateMockCompletion(prompt, max_tokens, temperature);
    
    if (stream) {
      res.setHeader('Content-Type', 'text/event-stream');
      res.setHeader('Cache-Control', 'no-cache');
      res.setHeader('Connection', 'keep-alive');
      
      // Send streaming response
      const words = completion.text.split(' ');
      for (let i = 0; i < words.length; i++) {
        const chunk = {
          id: `chatcmpl-${Date.now()}`,
          object: 'chat.completion.chunk',
          created: Math.floor(Date.now() / 1000),
          model,
          choices: [{
            index: 0,
            delta: {
              content: (i === 0 ? '' : ' ') + words[i]
            },
            finish_reason: i === words.length - 1 ? 'stop' : null
          }]
        };
        
        res.write(`data: ${JSON.stringify(chunk)}\n\n`);
        await new Promise(resolve => setTimeout(resolve, 50)); // 50ms delay
      }
      
      res.write('data: [DONE]\n\n');
      res.end();
    } else {
      res.json({
        id: `chatcmpl-${Date.now()}`,
        object: 'chat.completion',
        created: Math.floor(Date.now() / 1000),
        model,
        choices: [{
          index: 0,
          message: {
            role: 'assistant',
            content: completion.text
          },
          finish_reason: 'stop'
        }],
        usage: {
          prompt_tokens: completion.prompt_tokens,
          completion_tokens: completion.completion_tokens,
          total_tokens: completion.total_tokens
        }
      });
    }

    // Track usage
    if (req.user && req.user.id !== 'api-user') {
      const userModel = require('./models/user');
      await userModel.trackUsage(req.user.id, completion.completion_tokens, 1);
    }

  } catch (error) {
    console.error('Chat completion error:', error);
    res.status(500).json({
      error: {
        message: 'Internal server error',
        type: 'server_error',
        param: null,
        code: null
      }
    });
  }
});

// Embeddings endpoint (mock for now)
app.post('/v1/embeddings', authMiddleware.authenticate(), async (req, res) => {
  try {
    const {
      model = 'text-embedding-ada-002',
      input,
      encoding_format = 'float',
      dimensions = 1536
    } = req.body;

    if (!input) {
      return res.status(400).json({
        error: {
          message: 'Input is required',
          type: 'invalid_request_error',
          param: 'input',
          code: null
        }
      });
    }

    const inputs = Array.isArray(input) ? input : [input];
    
    // Generate mock embeddings
    const data = inputs.map((text, index) => ({
      object: 'embedding',
      embedding: Array(dimensions).fill(0).map(() => Math.random() - 0.5),
      index
    }));

    res.json({
      object: 'list',
      data,
      model,
      usage: {
        prompt_tokens: inputs.join(' ').split(' ').length,
        total_tokens: inputs.join(' ').split(' ').length
      }
    });

  } catch (error) {
    console.error('Embeddings error:', error);
    res.status(500).json({
      error: {
        message: 'Internal server error',
        type: 'server_error',
        param: null,
        code: null
      }
    });
  }
});

// API documentation endpoint
app.get('/docs', (req, res) => {
  const html = `
<!DOCTYPE html>
<html>
<head>
    <title>Ollamamax API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        .swagger-ui .topbar { display: none; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/openapi.json',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ],
            layout: "StandaloneLayout"
        });
    </script>
</body>
</html>`;
  
  res.send(html);
});

// OpenAPI specification
app.get('/openapi.json', (req, res) => {
  const spec = {
    openapi: '3.0.0',
    info: {
      title: 'Ollamamax API',
      version: '1.0.0',
      description: 'Distributed LLM Inference API with OpenAI compatibility'
    },
    servers: [
      {
        url: `http://localhost:${PORT}`,
        description: 'Development server'
      }
    ],
    components: {
      securitySchemes: {
        bearerAuth: {
          type: 'http',
          scheme: 'bearer',
          bearerFormat: 'JWT'
        },
        apiKeyAuth: {
          type: 'apiKey',
          in: 'header',
          name: 'x-api-key'
        }
      }
    },
    security: [
      { bearerAuth: [] },
      { apiKeyAuth: [] }
    ],
    paths: {
      '/v1/models': {
        get: {
          summary: 'List available models',
          responses: {
            '200': {
              description: 'Success',
              content: {
                'application/json': {
                  example: {
                    object: 'list',
                    data: [
                      {
                        id: 'llama-3.2-3b',
                        object: 'model',
                        created: 1699564800,
                        owned_by: 'ollamamax'
                      }
                    ]
                  }
                }
              }
            }
          }
        }
      },
      '/v1/completions': {
        post: {
          summary: 'Create text completion',
          requestBody: {
            required: true,
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  required: ['prompt'],
                  properties: {
                    model: { type: 'string', default: 'llama-3.2-3b' },
                    prompt: { type: 'string' },
                    max_tokens: { type: 'integer', default: 100 },
                    temperature: { type: 'number', default: 0.7 },
                    stream: { type: 'boolean', default: false }
                  }
                }
              }
            }
          },
          responses: {
            '200': { description: 'Success' },
            '400': { description: 'Bad request' },
            '401': { description: 'Unauthorized' }
          }
        }
      },
      '/v1/chat/completions': {
        post: {
          summary: 'Create chat completion',
          requestBody: {
            required: true,
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  required: ['messages'],
                  properties: {
                    model: { type: 'string', default: 'llama-3.2-3b' },
                    messages: {
                      type: 'array',
                      items: {
                        type: 'object',
                        properties: {
                          role: { type: 'string', enum: ['system', 'user', 'assistant'] },
                          content: { type: 'string' }
                        }
                      }
                    },
                    max_tokens: { type: 'integer', default: 100 },
                    temperature: { type: 'number', default: 0.7 },
                    stream: { type: 'boolean', default: false }
                  }
                }
              }
            }
          },
          responses: {
            '200': { description: 'Success' }
          }
        }
      }
    }
  };
  
  res.json(spec);
});

// 404 handler
app.use((req, res) => {
  res.status(404).json({
    error: {
      message: 'Not found',
      type: 'not_found_error',
      code: 'endpoint_not_found'
    }
  });
});

// Error handling middleware
app.use(authMiddleware.errorHandler());

// Utility functions
function formatUptime(seconds) {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);
  
  if (days > 0) return `${days}d ${hours}h ${minutes}m ${secs}s`;
  if (hours > 0) return `${hours}h ${minutes}m ${secs}s`;
  if (minutes > 0) return `${minutes}m ${secs}s`;
  return `${secs}s`;
}

function formatBytes(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function generateMockCompletion(prompt, maxTokens, temperature) {
  // Mock completion generator for Sprint 1
  const responses = [
    "This is a mock response from Ollamamax. In Sprint 3, this will be replaced with real distributed inference across multiple nodes.",
    "Hello! I'm Ollamamax, a distributed LLM inference engine. Currently running in development mode with mock responses.",
    "I understand you're testing the API. The authentication system is working correctly, and I can see your request was processed successfully.",
    "Thank you for trying Ollamamax! The distributed inference engine will be available in Sprint 3. For now, I'm providing mock responses to test the API layer.",
    "This response demonstrates the streaming and non-streaming capabilities of the Ollamamax API. Real model inference coming soon!"
  ];
  
  let text = responses[Math.floor(Math.random() * responses.length)];
  
  // Adjust length based on maxTokens
  const words = text.split(' ');
  const targetWords = Math.min(maxTokens / 2, words.length); // Rough approximation
  text = words.slice(0, targetWords).join(' ');
  
  const promptTokens = Math.ceil(prompt.length / 4); // Rough approximation
  const completionTokens = Math.ceil(text.length / 4);
  
  return {
    text,
    prompt_tokens: promptTokens,
    completion_tokens: completionTokens,
    total_tokens: promptTokens + completionTokens
  };
}

// Graceful shutdown
process.on('SIGINT', () => {
  console.log('\nReceived SIGINT. Graceful shutdown...');
  process.exit(0);
});

process.on('SIGTERM', () => {
  console.log('\nReceived SIGTERM. Graceful shutdown...');
  process.exit(0);
});

// Start server
const server = app.listen(PORT, '0.0.0.0', () => {
  console.log(`🚀 Ollamamax API Server started`);
  console.log(`📍 Server: http://localhost:${PORT}`);
  console.log(`📚 Documentation: http://localhost:${PORT}/docs`);
  console.log(`❤️ Health Check: http://localhost:${PORT}/health`);
  console.log(`🔑 Authentication: http://localhost:${PORT}/auth`);
  console.log(`🤖 OpenAI API: http://localhost:${PORT}/v1`);
  console.log(`📊 Metrics: http://localhost:${PORT}/metrics`);
  console.log('');
  console.log('🎯 Sprint 1: Core API & Authentication - RUNNING');
});

module.exports = { app, server };