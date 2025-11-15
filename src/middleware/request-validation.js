// Request Validation Middleware
// Implements comprehensive request size limits and security protections
// Addresses security and performance requirements from KNOWN_ISSUES.md

const rateLimit = require('express-rate-limit');
const slowDown = require('express-slow-down');

// Configuration constants
const REQUEST_LIMITS = {
  // HTTP body size limits
  JSON_LIMIT: parseInt(process.env.MAX_JSON_SIZE) || 10 * 1024 * 1024,        // 10MB
  URLENCODED_LIMIT: parseInt(process.env.MAX_URLENCODED_SIZE) || 10 * 1024 * 1024, // 10MB
  RAW_LIMIT: parseInt(process.env.MAX_RAW_SIZE) || 5 * 1024 * 1024,           // 5MB
  
  // WebSocket message limits
  WEBSOCKET_MESSAGE_LIMIT: parseInt(process.env.MAX_WS_MESSAGE_SIZE) || 1024 * 1024, // 1MB
  
  // Application-specific limits
  INFERENCE_PROMPT_LIMIT: parseInt(process.env.MAX_PROMPT_SIZE) || 100 * 1024,      // 100KB
  MODEL_NAME_LIMIT: parseInt(process.env.MAX_MODEL_NAME_SIZE) || 256,               // 256 bytes
  SESSION_ID_LIMIT: parseInt(process.env.MAX_SESSION_ID_SIZE) || 128,               // 128 bytes
  
  // Request header limits
  HEADER_SIZE_LIMIT: parseInt(process.env.MAX_HEADER_SIZE) || 8 * 1024,            // 8KB
  HEADER_COUNT_LIMIT: parseInt(process.env.MAX_HEADER_COUNT) || 100,               // 100 headers
  
  // Query parameter limits
  QUERY_PARAM_LIMIT: parseInt(process.env.MAX_QUERY_PARAMS) || 50,                 // 50 params
  QUERY_VALUE_LIMIT: parseInt(process.env.MAX_QUERY_VALUE_SIZE) || 1024,           // 1KB per value
};

// Rate limiting configuration
const createRateLimit = (windowMs, max, message, skipSuccessfulRequests = false) => {
  return rateLimit({
    windowMs,
    max,
    message: {
      error: 'Too Many Requests',
      message,
      code: 'RATE_LIMIT_EXCEEDED',
      retryAfter: Math.ceil(windowMs / 1000)
    },
    standardHeaders: true,
    legacyHeaders: false,
    skipSuccessfulRequests,
    handler: (req, res) => {
      console.warn(`Rate limit exceeded for ${req.ip} on ${req.path}: ${req.method} ${req.originalUrl}`);
      res.status(429).json({
        error: 'Too Many Requests',
        message,
        code: 'RATE_LIMIT_EXCEEDED',
        retryAfter: Math.ceil(windowMs / 1000)
      });
    }
  });
};

// Speed limiting (progressive delay)
const createSpeedLimit = (windowMs, delayAfter, delayMs, maxDelayMs) => {
  return slowDown({
    windowMs,
    delayAfter,
    delayMs,
    maxDelayMs,
    skipFailedRequests: false,
    skipSuccessfulRequests: false,
    onLimitReached: (req, res, options) => {
      console.warn(`Speed limit reached for ${req.ip} on ${req.path}`);
    }
  });
};

// Request size validation middleware
const validateRequestSize = (req, res, next) => {
  const contentLength = parseInt(req.get('content-length') || '0');
  const contentType = req.get('content-type') || '';
  
  // Determine appropriate limit based on content type
  let limit = REQUEST_LIMITS.RAW_LIMIT;
  if (contentType.includes('application/json')) {
    limit = REQUEST_LIMITS.JSON_LIMIT;
  } else if (contentType.includes('application/x-www-form-urlencoded')) {
    limit = REQUEST_LIMITS.URLENCODED_LIMIT;
  }
  
  if (contentLength > limit) {
    console.warn(`Request too large: ${contentLength} bytes (limit: ${limit}) from ${req.ip} to ${req.path}`);
    return res.status(413).json({
      error: 'Request Entity Too Large',
      message: `Request body size (${contentLength} bytes) exceeds maximum allowed size (${limit} bytes)`,
      code: 'REQUEST_TOO_LARGE',
      maxSize: limit,
      contentType: contentType
    });
  }
  
  next();
};

// Header validation middleware
const validateHeaders = (req, res, next) => {
  const headers = req.headers;
  const headerCount = Object.keys(headers).length;
  
  // Check header count
  if (headerCount > REQUEST_LIMITS.HEADER_COUNT_LIMIT) {
    console.warn(`Too many headers: ${headerCount} (limit: ${REQUEST_LIMITS.HEADER_COUNT_LIMIT}) from ${req.ip}`);
    return res.status(400).json({
      error: 'Bad Request',
      message: `Too many headers (${headerCount}). Maximum allowed: ${REQUEST_LIMITS.HEADER_COUNT_LIMIT}`,
      code: 'TOO_MANY_HEADERS'
    });
  }
  
  // Check individual header sizes
  for (const [name, value] of Object.entries(headers)) {
    const headerSize = Buffer.byteLength(`${name}: ${value}`, 'utf8');
    if (headerSize > REQUEST_LIMITS.HEADER_SIZE_LIMIT) {
      console.warn(`Header too large: ${name} (${headerSize} bytes) from ${req.ip}`);
      return res.status(400).json({
        error: 'Bad Request',
        message: `Header '${name}' is too large (${headerSize} bytes). Maximum allowed: ${REQUEST_LIMITS.HEADER_SIZE_LIMIT} bytes`,
        code: 'HEADER_TOO_LARGE',
        headerName: name
      });
    }
  }
  
  next();
};

// Query parameter validation middleware
const validateQueryParams = (req, res, next) => {
  const queryParams = Object.keys(req.query);
  
  // Check parameter count
  if (queryParams.length > REQUEST_LIMITS.QUERY_PARAM_LIMIT) {
    console.warn(`Too many query parameters: ${queryParams.length} from ${req.ip} to ${req.path}`);
    return res.status(400).json({
      error: 'Bad Request',
      message: `Too many query parameters (${queryParams.length}). Maximum allowed: ${REQUEST_LIMITS.QUERY_PARAM_LIMIT}`,
      code: 'TOO_MANY_QUERY_PARAMS'
    });
  }
  
  // Check individual parameter values
  for (const [key, value] of Object.entries(req.query)) {
    const valueStr = Array.isArray(value) ? value.join(',') : String(value);
    const valueSize = Buffer.byteLength(valueStr, 'utf8');
    
    if (valueSize > REQUEST_LIMITS.QUERY_VALUE_LIMIT) {
      console.warn(`Query parameter value too large: ${key} (${valueSize} bytes) from ${req.ip}`);
      return res.status(400).json({
        error: 'Bad Request',
        message: `Query parameter '${key}' value is too large (${valueSize} bytes). Maximum allowed: ${REQUEST_LIMITS.QUERY_VALUE_LIMIT} bytes`,
        code: 'QUERY_VALUE_TOO_LARGE',
        parameterName: key
      });
    }
  }
  
  next();
};

// Application-specific validation middleware
const validateInferenceRequest = (req, res, next) => {
  if (req.body && req.body.type === 'inference') {
    // Validate prompt size
    if (req.body.content) {
      const promptSize = Buffer.byteLength(req.body.content, 'utf8');
      if (promptSize > REQUEST_LIMITS.INFERENCE_PROMPT_LIMIT) {
        console.warn(`Inference prompt too large: ${promptSize} bytes from ${req.ip}`);
        return res.status(400).json({
          error: 'Bad Request',
          message: `Inference prompt is too large (${promptSize} bytes). Maximum allowed: ${REQUEST_LIMITS.INFERENCE_PROMPT_LIMIT} bytes`,
          code: 'PROMPT_TOO_LARGE',
          maxSize: REQUEST_LIMITS.INFERENCE_PROMPT_LIMIT
        });
      }
    }
    
    // Validate model name
    if (req.body.model) {
      const modelNameSize = Buffer.byteLength(req.body.model, 'utf8');
      if (modelNameSize > REQUEST_LIMITS.MODEL_NAME_LIMIT) {
        console.warn(`Model name too large: ${modelNameSize} bytes from ${req.ip}`);
        return res.status(400).json({
          error: 'Bad Request',
          message: `Model name is too large (${modelNameSize} bytes). Maximum allowed: ${REQUEST_LIMITS.MODEL_NAME_LIMIT} bytes`,
          code: 'MODEL_NAME_TOO_LARGE'
        });
      }
    }
    
    // Validate session ID
    if (req.body.sessionId) {
      const sessionIdSize = Buffer.byteLength(req.body.sessionId, 'utf8');
      if (sessionIdSize > REQUEST_LIMITS.SESSION_ID_LIMIT) {
        console.warn(`Session ID too large: ${sessionIdSize} bytes from ${req.ip}`);
        return res.status(400).json({
          error: 'Bad Request',
          message: `Session ID is too large (${sessionIdSize} bytes). Maximum allowed: ${REQUEST_LIMITS.SESSION_ID_LIMIT} bytes`,
          code: 'SESSION_ID_TOO_LARGE'
        });
      }
    }
  }
  
  next();
};

// WebSocket message validation
const validateWebSocketMessage = (message, ws) => {
  // Check message size
  if (message.length > REQUEST_LIMITS.WEBSOCKET_MESSAGE_LIMIT) {
    const error = {
      type: 'error',
      error: 'Message Too Large',
      message: `WebSocket message size (${message.length} bytes) exceeds maximum allowed size (${REQUEST_LIMITS.WEBSOCKET_MESSAGE_LIMIT} bytes)`,
      code: 'MESSAGE_TOO_LARGE',
      maxSize: REQUEST_LIMITS.WEBSOCKET_MESSAGE_LIMIT
    };
    
    console.warn(`WebSocket message too large: ${message.length} bytes from ${ws._socket?.remoteAddress}`);
    ws.send(JSON.stringify(error));
    return false;
  }
  
  try {
    const data = JSON.parse(message);
    
    // Validate inference-specific content
    if (data.type === 'inference' && data.content) {
      const promptSize = Buffer.byteLength(data.content, 'utf8');
      if (promptSize > REQUEST_LIMITS.INFERENCE_PROMPT_LIMIT) {
        const error = {
          type: 'error',
          error: 'Prompt Too Large',
          message: `Inference prompt size (${promptSize} bytes) exceeds maximum allowed size (${REQUEST_LIMITS.INFERENCE_PROMPT_LIMIT} bytes)`,
          code: 'PROMPT_TOO_LARGE',
          maxSize: REQUEST_LIMITS.INFERENCE_PROMPT_LIMIT
        };
        
        console.warn(`WebSocket inference prompt too large: ${promptSize} bytes from ${ws._socket?.remoteAddress}`);
        ws.send(JSON.stringify(error));
        return false;
      }
    }
    
    return true;
  } catch (parseError) {
    const error = {
      type: 'error',
      error: 'Invalid JSON',
      message: 'Failed to parse WebSocket message as JSON',
      code: 'INVALID_JSON'
    };
    
    console.warn(`Invalid WebSocket JSON from ${ws._socket?.remoteAddress}: ${parseError.message}`);
    ws.send(JSON.stringify(error));
    return false;
  }
};

module.exports = {
  REQUEST_LIMITS,
  createRateLimit,
  createSpeedLimit,
  validateRequestSize,
  validateHeaders,
  validateQueryParams,
  validateInferenceRequest,
  validateWebSocketMessage
};
