/**
 * OpenRouter JavaScript Client for Node.js
 * Provides integration with OpenRouter API for Sonoma Sky Alpha model
 */

const https = require('https');
const http = require('http');
const { URL } = require('url');

class OpenRouterClient {
    constructor(config = {}) {
        this.apiKey = config.apiKey || process.env.OPENROUTER_API_KEY;
        this.baseURL = config.baseURL || 'https://openrouter.ai/api/v1';
        this.timeout = config.timeout || 300000; // 5 minutes
        this.maxRetries = config.maxRetries || 3;
        this.retryDelay = config.retryDelay || 2000;
        this.userAgent = config.userAgent || 'OllamaMax/1.0.0';
        this.debugMode = config.debugMode || false;
        
        if (!this.apiKey) {
            throw new Error('OpenRouter API key is required');
        }
    }
    
    /**
     * Send chat completion request to OpenRouter
     */
    async chatCompletion(request) {
        // Default to Sonoma Sky Alpha if no model specified
        if (!request.model) {
            request.model = 'alpindale/sonoma-sky-alpha';
        }
        
        // Validate request
        this.validateRequest(request);
        
        let lastError;
        
        // Retry logic
        for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
            try {
                const response = await this.makeRequest('/chat/completions', 'POST', request);
                
                if (this.debugMode) {
                    console.log('OpenRouter response:', {
                        id: response.id,
                        model: response.model,
                        usage: response.usage
                    });
                }
                
                return response;
            } catch (error) {
                lastError = error;
                
                if (this.debugMode) {
                    console.warn(`OpenRouter request attempt ${attempt + 1} failed:`, error.message);
                }
                
                if (attempt < this.maxRetries) {
                    await this.delay(this.retryDelay * (attempt + 1));
                }
            }
        }
        
        throw new Error(`OpenRouter request failed after ${this.maxRetries + 1} attempts: ${lastError.message}`);
    }
    
    /**
     * Get available models from OpenRouter
     */
    async getModels() {
        try {
            const response = await this.makeRequest('/models', 'GET');
            return response.data || [];
        } catch (error) {
            throw new Error(`Failed to get models: ${error.message}`);
        }
    }
    
    /**
     * Check health of OpenRouter API
     */
    async health() {
        try {
            const testRequest = {
                model: 'alpindale/sonoma-sky-alpha',
                messages: [{ role: 'user', content: 'Hello' }],
                max_tokens: 5
            };
            
            await this.chatCompletion(testRequest);
            return true;
        } catch (error) {
            throw new Error(`Health check failed: ${error.message}`);
        }
    }
    
    /**
     * Make HTTP request to OpenRouter API
     */
    async makeRequest(endpoint, method, body = null) {
        return new Promise((resolve, reject) => {
            const url = new URL(endpoint, this.baseURL);
            const isHttps = url.protocol === 'https:';
            const httpModule = isHttps ? https : http;
            
            const requestData = body ? JSON.stringify(body) : null;
            
            const options = {
                hostname: url.hostname,
                port: url.port || (isHttps ? 443 : 80),
                path: url.pathname + url.search,
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.apiKey}`,
                    'User-Agent': this.userAgent,
                    'HTTP-Referer': 'https://github.com/khryptorgraphics/ollamamax',
                    'X-Title': 'OllamaMax Distributed AI Platform'
                },
                timeout: this.timeout
            };
            
            if (requestData) {
                options.headers['Content-Length'] = Buffer.byteLength(requestData);
            }
            
            if (this.debugMode) {
                console.log('OpenRouter request:', {
                    url: url.toString(),
                    method: method,
                    headers: options.headers,
                    body: body
                });
            }
            
            const req = httpModule.request(options, (res) => {
                let data = '';
                
                res.on('data', (chunk) => {
                    data += chunk;
                });
                
                res.on('end', () => {
                    try {
                        const parsedData = JSON.parse(data);
                        
                        if (this.debugMode) {
                            console.log('OpenRouter response:', {
                                statusCode: res.statusCode,
                                headers: res.headers,
                                body: parsedData
                            });
                        }
                        
                        if (res.statusCode >= 400) {
                            const error = parsedData.error || { message: `HTTP ${res.statusCode}` };
                            reject(new Error(error.message || 'Unknown error'));
                        } else {
                            resolve(parsedData);
                        }
                    } catch (parseError) {
                        reject(new Error(`Failed to parse response: ${parseError.message}`));
                    }
                });
            });
            
            req.on('error', (error) => {
                reject(new Error(`Request failed: ${error.message}`));
            });
            
            req.on('timeout', () => {
                req.destroy();
                reject(new Error('Request timeout'));
            });
            
            if (requestData) {
                req.write(requestData);
            }
            
            req.end();
        });
    }
    
    /**
     * Validate chat completion request
     */
    validateRequest(request) {
        if (!request.model) {
            throw new Error('Model is required');
        }
        
        if (!request.messages || !Array.isArray(request.messages) || request.messages.length === 0) {
            throw new Error('At least one message is required');
        }
        
        for (let i = 0; i < request.messages.length; i++) {
            const message = request.messages[i];
            
            if (!message.role) {
                throw new Error(`Message ${i}: role is required`);
            }
            
            if (!message.content) {
                throw new Error(`Message ${i}: content is required`);
            }
            
            if (!['system', 'user', 'assistant'].includes(message.role)) {
                throw new Error(`Message ${i}: invalid role ${message.role}`);
            }
        }
        
        if (request.temperature !== undefined && (request.temperature < 0 || request.temperature > 2)) {
            throw new Error('Temperature must be between 0 and 2');
        }
        
        if (request.top_p !== undefined && (request.top_p < 0 || request.top_p > 1)) {
            throw new Error('top_p must be between 0 and 1');
        }
        
        if (request.max_tokens !== undefined && (request.max_tokens < 1 || request.max_tokens > 131072)) {
            throw new Error('max_tokens must be between 1 and 131072');
        }
    }
    
    /**
     * Delay helper for retries
     */
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
    
    /**
     * Get client configuration
     */
    getConfig() {
        return {
            baseURL: this.baseURL,
            timeout: this.timeout,
            maxRetries: this.maxRetries,
            retryDelay: this.retryDelay,
            userAgent: this.userAgent,
            debugMode: this.debugMode
        };
    }
    
    /**
     * Update client configuration
     */
    updateConfig(config) {
        if (config.timeout !== undefined) this.timeout = config.timeout;
        if (config.maxRetries !== undefined) this.maxRetries = config.maxRetries;
        if (config.retryDelay !== undefined) this.retryDelay = config.retryDelay;
        if (config.debugMode !== undefined) this.debugMode = config.debugMode;
    }
}

module.exports = { OpenRouterClient };
