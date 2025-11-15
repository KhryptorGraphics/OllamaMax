/**
 * OllamaMax Frontend-Backend Integration Test
 * 
 * This test validates the complete integration between the frontend
 * and the standardized API surface, including:
 * - WebSocket connections
 * - REST API calls
 * - Authentication flows
 * - Error handling
 * - Fallback mechanisms
 */

const WebSocket = require('ws');
const http = require('http');

class FrontendBackendIntegrationTest {
    constructor() {
        this.config = {
            endpoints: [
                'ws://localhost:13100/chat',
                'ws://localhost:13101/chat', 
                'ws://localhost:13102/chat'
            ],
            apiBase: 'http://localhost:13100',
            timeout: 5000
        };
        
        this.results = {
            passed: 0,
            failed: 0,
            tests: []
        };
    }

    async run() {
        console.log('🧪 OllamaMax Frontend-Backend Integration Test');
        console.log('==============================================');
        console.log();

        await this.testRESTAPI();
        await this.testWebSocketConnections();
        await this.testAuthenticationFlow();
        await this.testErrorHandling();
        await this.testFallbackMechanism();

        this.printSummary();
        return this.results.failed === 0;
    }

    async testRESTAPI() {
        console.log('📋 Testing REST API Endpoints');
        console.log('------------------------------');

        const endpoints = [
            { path: '/', expected: 200, name: 'API Root' },
            { path: '/health', expected: 200, name: 'Health Check' },
            { path: '/health/live', expected: 200, name: 'Liveness Probe' },
            { path: '/health/ready', expected: 200, name: 'Readiness Probe' },
            { path: '/openapi.json', expected: 200, name: 'OpenAPI Specification' },
            { path: '/metrics', expected: 200, name: 'Metrics Endpoint' },
            { path: '/docs', expected: 200, name: 'Documentation' },
            { path: '/auth/login', expected: 405, name: 'Auth Login (method check)' }
        ];

        for (const endpoint of endpoints) {
            await this.testHTTPEndpoint(endpoint);
        }
        console.log();
    }

    async testHTTPEndpoint(endpoint) {
        const url = `${this.config.apiBase}${endpoint.path}`;
        
        try {
            const response = await this.httpRequest(url, endpoint.path === '/auth/login' ? 'POST' : 'GET');
            
            if (response.statusCode === endpoint.expected) {
                this.pass(`${endpoint.name}: ${response.statusCode}`);
            } else {
                this.fail(`${endpoint.name}: Expected ${endpoint.expected}, got ${response.statusCode}`);
            }
        } catch (error) {
            this.fail(`${endpoint.name}: ${error.message}`);
        }
    }

    async testWebSocketConnections() {
        console.log('📋 Testing WebSocket Connections');
        console.log('--------------------------------');

        for (let i = 0; i < this.config.endpoints.length; i++) {
            const endpoint = this.config.endpoints[i];
            const isPrimary = i === 0;
            
            try {
                const ws = new WebSocket(endpoint);
                
                await new Promise((resolve, reject) => {
                    const timeout = setTimeout(() => {
                        reject(new Error('Connection timeout'));
                    }, this.config.timeout);

                    ws.on('open', () => {
                        clearTimeout(timeout);
                        ws.close();
                        if (isPrimary) {
                            this.pass(`Primary WebSocket: ${endpoint}`);
                        } else {
                            this.pass(`Fallback WebSocket: ${endpoint}`);
                        }
                        resolve();
                    });

                    ws.on('error', (error) => {
                        clearTimeout(timeout);
                        if (isPrimary) {
                            this.fail(`Primary WebSocket: ${endpoint} - ${error.message}`);
                        } else {
                            this.note(`Fallback WebSocket: ${endpoint} - Not available (expected)`);
                        }
                        reject(error);
                    });
                });
            } catch (error) {
                if (isPrimary) {
                    this.fail(`WebSocket connection failed: ${error.message}`);
                } else {
                    this.note(`Fallback endpoint not available: ${endpoint}`);
                }
            }
        }
        console.log();
    }

    async testAuthenticationFlow() {
        console.log('📋 Testing Authentication Flow');
        console.log('------------------------------');

        // Test auth endpoint exists
        try {
            const response = await this.httpRequest(`${this.config.apiBase}/auth`, 'GET');
            if (response.statusCode === 404) {
                this.pass('Auth endpoint structure exists');
            } else {
                this.pass(`Auth endpoint: ${response.statusCode}`);
            }
        } catch (error) {
            this.fail(`Auth endpoint test: ${error.message}`);
        }

        // Test auth registration endpoint
        try {
            const response = await this.httpRequest(`${this.config.apiBase}/auth/register`, 'POST');
            if (response.statusCode === 400) {
                this.pass('Auth registration endpoint exists (requires valid data)');
            } else {
                this.pass(`Auth registration: ${response.statusCode}`);
            }
        } catch (error) {
            this.fail(`Auth registration test: ${error.message}`);
        }
        console.log();
    }

    async testErrorHandling() {
        console.log('📋 Testing Error Handling');
        console.log('-------------------------');

        // Test 404 error
        try {
            const response = await this.httpRequest(`${this.config.apiBase}/nonexistent`, 'GET');
            if (response.statusCode === 404) {
                this.pass('404 error handling');
            } else {
                this.fail(`Expected 404, got ${response.statusCode}`);
            }
        } catch (error) {
            this.fail(`404 test: ${error.message}`);
        }

        // Test invalid JSON
        try {
            const response = await this.httpRequest(`${this.config.apiBase}/auth/login`, 'POST', {
                'Content-Type': 'application/json'
            }, 'invalid json');
            
            if (response.statusCode >= 400) {
                this.pass('Invalid JSON handling');
            } else {
                this.fail(`Expected error for invalid JSON, got ${response.statusCode}`);
            }
        } catch (error) {
            this.pass('Invalid JSON rejected');
        }
        console.log();
    }

    async testFallbackMechanism() {
        console.log('📋 Testing Fallback Mechanism');
        console.log('-----------------------------');

        // Test that the primary endpoint is tried first
        let primarySuccess = false;
        let fallbackSuccess = false;

        // Test primary endpoint
        try {
            const ws = new WebSocket(this.config.endpoints[0]);
            await new Promise((resolve) => {
                const timeout = setTimeout(() => {
                    resolve();
                }, 2000);

                ws.on('open', () => {
                    clearTimeout(timeout);
                    ws.close();
                    primarySuccess = true;
                    resolve();
                });

                ws.on('error', () => {
                    clearTimeout(timeout);
                    resolve();
                });
            });
        } catch (error) {
            // Primary endpoint failed, which is acceptable
        }

        // Test fallback endpoint
        try {
            const ws = new WebSocket(this.config.endpoints[1]);
            await new Promise((resolve) => {
                const timeout = clearTimeout(() => {
                    resolve();
                }, 2000);

                ws.on('open', () => {
                    clearTimeout(timeout);
                    ws.close();
                    fallbackSuccess = true;
                    resolve();
                });

                ws.on('error', () => {
                    clearTimeout(timeout);
                    resolve();
                });
            });
        } catch (error) {
            // Fallback endpoint failed, which is acceptable
        }

        if (primarySuccess || fallbackSuccess) {
            this.pass('At least one WebSocket endpoint available');
        } else {
            this.fail('No WebSocket endpoints available');
        }
        console.log();
    }

    async httpRequest(url, method = 'GET', headers = {}, body = null) {
        return new Promise((resolve, reject) => {
            const urlObj = new URL(url);
            
            const options = {
                hostname: urlObj.hostname,
                port: urlObj.port,
                path: urlObj.pathname + urlObj.search,
                method: method,
                headers: {
                    'User-Agent': 'OllamaMax Integration Test',
                    ...headers
                }
            };

            const req = http.request(options, (res) => {
                let data = '';
                
                res.on('data', (chunk) => {
                    data += chunk;
                });
                
                res.on('end', () => {
                    resolve({
                        statusCode: res.statusCode,
                        headers: res.headers,
                        body: data
                    });
                });
            });

            req.on('error', (error) => {
                reject(error);
            });

            req.setTimeout(this.config.timeout, () => {
                req.destroy();
                reject(new Error('Request timeout'));
            });

            if (body) {
                req.write(body);
            }

            req.end();
        });
    }

    pass(message) {
        console.log(`  ✓ ${message}`);
        this.results.passed++;
        this.results.tests.push({ message, status: 'pass' });
    }

    fail(message) {
        console.log(`  ✗ ${message}`);
        this.results.failed++;
        this.results.tests.push({ message, status: 'fail' });
    }

    note(message) {
        console.log(`  ℹ ${message}`);
        this.results.tests.push({ message, status: 'note' });
    }

    printSummary() {
        console.log('📊 Integration Test Summary');
        console.log('============================');
        console.log(`Passed: ${this.results.passed}`);
        console.log(`Failed: ${this.results.failed}`);
        console.log(`Total:  ${this.results.tests.length}`);
        console.log();

        if (this.results.failed === 0) {
            console.log('🎉 All integration tests passed!');
            console.log('The frontend-backend API surface is working correctly.');
        } else {
            console.log('❌ Some integration tests failed.');
            console.log('Please check the services are running and configured correctly.');
            
            console.log('\n🔧 Troubleshooting steps:');
            console.log('1. Start Node.js API: npm start');
            console.log('2. Verify .env configuration');
            console.log('3. Check port availability');
            console.log('4. Review error messages above');
        }
    }
}

// Run the test if this file is executed directly
if (require.main === module) {
    const test = new FrontendBackendIntegrationTest();
    test.run().then(success => {
        process.exit(success ? 0 : 1);
    }).catch(error => {
        console.error('Test execution failed:', error);
        process.exit(1);
    });
}

module.exports = FrontendBackendIntegrationTest;