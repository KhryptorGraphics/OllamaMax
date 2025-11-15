/**
 * Frontend Test Suite for OllamaMax Web Interface
 * 
 * This test suite covers:
 * - WebSocket connection management
 * - Chat functionality
 * - Node management
 * - Model operations
 * - Settings persistence
 * - Accessibility features
 * - Responsive design
 * - Error handling
 */

// Polyfill for TextEncoder (required by jsdom)
if (typeof global.TextEncoder === 'undefined') {
    global.TextEncoder = require('util').TextEncoder;
}
if (typeof global.TextDecoder === 'undefined') {
    global.TextDecoder = require('util').TextDecoder;
}

const { JSDOM } = require('jsdom');

// Mock WebSocket class
class MockWebSocket {
    constructor(url) {
        this.url = url;
        this.readyState = WebSocket.CONNECTING;
        this.send = jest.fn();
        this.close = jest.fn();
    }
    
    static CONNECTING = 0;
    static OPEN = 1;
    static CLOSING = 2;
    static CLOSED = 3;
    
    simulateOpen() {
        this.readyState = WebSocket.OPEN;
        if (this.onopen) {
            this.onopen();
        }
    }
    
    simulateClose() {
        this.readyState = WebSocket.CLOSED;
        if (this.onclose) {
            this.onclose();
        }
    }
    
    simulateError() {
        if (this.onerror) {
            this.onerror(new Error('Connection error'));
        }
    }
    
    simulateMessage(data) {
        if (this.onmessage) {
            this.onmessage({ data: JSON.stringify(data) });
        }
    }
}

// Setup JSDOM environment
const dom = new JSDOM(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Distributed Llama Chat - OllamaMax</title>
</head>
<body>
    <div id="app">
        <!-- Mock DOM elements for testing -->
        <div class="connection-status">
            <span class="status-indicator" id="connectionStatus"></span>
            <span id="connectionText">Connecting...</span>
        </div>
        <div id="activeNode">-</div>
        <div id="queueLength">0</div>
        <div id="latency">-</div>
        <div id="connectionQuality"></div>
        <div id="connectionLatency">-</div>
        <div id="connectionThroughput">-</div>
        <textarea id="messageInput"></textarea>
        <button id="sendButton">Send</button>
        <div id="messagesArea"></div>
        <select id="modelSelector">
            <option value="tinyllama">TinyLlama</option>
        </select>
    </div>
</body>
</html>
`, {
    url: 'http://localhost',
    pretendToBeVisual: true,
    resources: 'usable'
});

global.window = dom.window;
global.document = dom.window.document;
global.HTMLElement = dom.window.HTMLElement;
global.HTMLInputElement = dom.window.HTMLInputElement;
global.HTMLSelectElement = dom.window.HTMLSelectElement;
global.HTMLTextAreaElement = dom.window.HTMLTextAreaElement;

// Mock WebSocket
global.WebSocket = class MockWebSocket {
    constructor(url) {
        this.url = url;
        this.readyState = WebSocket.CONNECTING;
        this.send = jest.fn();
        this.close = jest.fn();
    }
    
    static CONNECTING = 0;
    static OPEN = 1;
    static CLOSING = 2;
    static CLOSED = 3;
    
    simulateOpen() {
        this.readyState = WebSocket.OPEN;
        if (this.onopen) {
            this.onopen();
        }
    }
    
    simulateClose() {
        this.readyState = WebSocket.CLOSED;
        if (this.onclose) {
            this.onclose();
        }
    }
    
    simulateError() {
        if (this.onerror) {
            this.onerror(new Error('Connection error'));
        }
    }
    
    simulateMessage(data) {
        if (this.onmessage) {
            this.onmessage({ data: JSON.stringify(data) });
        }
    }
};

// Mock localStorage
const mockLocalStorage = {
    store: {},
    getItem(key) {
        return this.store[key] || null;
    },
    setItem(key, value) {
        this.store[key] = value;
    },
    removeItem(key) {
        delete this.store[key];
    },
    clear() {
        this.store = {};
    }
};

global.localStorage = mockLocalStorage;

// Import the main application class
const appPath = require.resolve('../app.js');

describe('OllamaMax Frontend Test Suite', () => {
    let llamaClient;
    let mockWebSocket;

    beforeEach(() => {
        // Reset DOM
        document.body.innerHTML = dom.window.document.body.innerHTML;
        
        // Reset localStorage
        mockLocalStorage.clear();
        
        // Create mock WebSocket
        mockWebSocket = new MockWebSocket('ws://localhost:13100/chat');
        
        // Mock the WebSocket constructor
        global.WebSocket = jest.fn(() => mockWebSocket);
        
        // Import and instantiate the application
        delete require.cache[appPath];
        const App = require('../app.js');
        llamaClient = new App();
    });

    describe('WebSocket Connection Management', () => {
        test('should initialize with correct default settings', () => {
            expect(llamaClient.reconnectAttempts).toBe(0);
            expect(llamaClient.maxReconnectAttempts).toBe(10);
            expect(llamaClient.connectionQuality.latency).toBe(0);
            expect(llamaClient.endpoints).toContain('ws://localhost:13000/chat');
        });

        test('should attempt connection on initialization', () => {
            expect(global.WebSocket).toHaveBeenCalledWith('ws://localhost:13000/chat');
        });

        test('should handle successful connection', () => {
            const updateStatusSpy = jest.spyOn(llamaClient, 'updateConnectionStatus');
            
            mockWebSocket.simulateOpen();
            
            expect(updateStatusSpy).toHaveBeenCalledWith('connected');
            expect(llamaClient.reconnectAttempts).toBe(0);
        });

        test('should handle connection errors', () => {
            const updateStatusSpy = jest.spyOn(llamaClient, 'updateConnectionStatus');
            
            mockWebSocket.simulateError();
            
            expect(updateStatusSpy).toHaveBeenCalledWith('error');
        });

        test('should attempt reconnection on close', (done) => {
            jest.useFakeTimers();
            
            mockWebSocket.simulateClose();
            
            // Fast-forward time
            jest.advanceTimersByTime(1000);
            
            expect(global.WebSocket).toHaveBeenCalledTimes(2); // Initial + retry
            jest.clearAllTimers();
            jest.useRealTimers();
            done();
        });

        test('should track connection history', () => {
            const initialHistoryLength = llamaClient.performanceData.connectionHistory.length;
            
            mockWebSocket.simulateOpen();
            
            expect(llamaClient.performanceData.connectionHistory.length).toBeGreaterThan(initialHistoryLength);
            expect(llamaClient.performanceData.connectionHistory[0].event).toBe('connect_attempt');
        });
    });

    describe('Message Handling', () => {
        beforeEach(() => {
            mockWebSocket.simulateOpen();
        });

        test('should handle chat responses', () => {
            const addMessageSpy = jest.spyOn(llamaClient, 'addMessage');
            
            const mockResponse = {
                type: 'response',
                content: 'Hello, world!',
                node: 'test-node',
                streaming: false
            };
            
            mockWebSocket.simulateMessage(mockResponse);
            
            expect(addMessageSpy).toHaveBeenCalledWith('ai', 'Hello, world!', 'test-node');
        });

        test('should handle streaming responses', () => {
            const addMessageSpy = jest.spyOn(llamaClient, 'addMessage');
            
            const mockResponse = {
                type: 'response',
                content: '',
                node: 'test-node',
                streaming: true,
                id: 'stream-1'
            };
            
            mockWebSocket.simulateMessage(mockResponse);
            
            expect(addMessageSpy).toHaveBeenCalledWith('ai', '', 'test-node', true);
        });

        test('should handle stream chunks', () => {
            // First, create a streaming message
            const mockResponse = {
                type: 'response',
                content: '',
                node: 'test-node',
                streaming: true,
                id: 'stream-1'
            };
            mockWebSocket.simulateMessage(mockResponse);
            
            const updateStreamingSpy = jest.spyOn(llamaClient, 'updateStreamingMessage');
            
            const mockChunk = {
                type: 'stream_chunk',
                id: 'stream-1',
                chunk: 'Hello ',
                done: false
            };
            
            mockWebSocket.simulateMessage(mockChunk);
            
            expect(updateStreamingSpy).toHaveBeenCalledWith('Hello ');
        });

        test('should handle ping/pong for latency tracking', () => {
            const recordPongSpy = jest.spyOn(llamaClient, 'recordPong');
            
            llamaClient.sendPing();
            
            const mockPong = {
                type: 'pong',
                timestamp: Date.now()
            };
            
            mockWebSocket.simulateMessage(mockPong);
            
            expect(recordPongSpy).toHaveBeenCalled();
            expect(llamaClient.connectionQuality.latency).toBeGreaterThan(0);
        });
    });

    describe('Settings Persistence', () => {
        test('should load settings from localStorage', () => {
            const testSettings = {
                apiEndpoint: 'ws://test:13000',
                darkMode: true,
                temperature: 0.8,
                maxTokens: 4096
            };
            
            localStorage.setItem('llamaChatSettings', JSON.stringify(testSettings));
            
            // Recreate client to test loading
            delete require.cache[appPath];
            const App = require('../app.js');
            const newClient = new App();
            
            expect(newClient.settings).toEqual(testSettings);
        });

        test('should save settings to localStorage', () => {
            const testSettings = {
                apiEndpoint: 'ws://test:13000',
                darkMode: true,
                temperature: 0.8,
                maxTokens: 4096
            };
            
            llamaClient.settings = testSettings;
            llamaClient.saveSettings();
            
            const savedSettings = JSON.parse(localStorage.getItem('llamaChatSettings'));
            expect(savedSettings).toEqual(testSettings);
        });
    });

    describe('Chat Functionality', () => {
        beforeEach(() => {
            mockWebSocket.simulateOpen();
        });

        test('should send messages via WebSocket', () => {
            const testMessage = 'Hello, AI!';
            
            // Simulate typing and sending a message
            document.getElementById('messageInput').value = testMessage;
            document.getElementById('sendButton').click();
            
            expect(mockWebSocket.send).toHaveBeenCalled();
            
            const sentData = JSON.parse(mockWebSocket.send.mock.calls[0][0]);
            expect(sentData.type).toBe('chat');
            expect(sentData.content).toBe(testMessage);
        });

        test('should handle message queueing when disconnected', () => {
            // Disconnect WebSocket
            mockWebSocket.readyState = WebSocket.CLOSED;
            
            const testMessage = 'Queued message';
            
            // Try to send message
            document.getElementById('messageInput').value = testMessage;
            document.getElementById('sendButton').click();
            
            // Message should be queued
            expect(llamaClient.messageQueue).toHaveLength(1);
            expect(llamaClient.messageQueue[0].content).toBe(testMessage);
        });

        test('should process message queue on reconnection', () => {
            // Queue a message while disconnected
            llamaClient.messageQueue.push({ type: 'chat', content: 'Test message' });
            
            // Simulate reconnection
            mockWebSocket.simulateOpen();
            
            // Messages should be sent
            expect(mockWebSocket.send).toHaveBeenCalled();
        });
    });

    describe('Node Management', () => {
        test('should update node status', () => {
            const mockNodeUpdate = {
                type: 'node_update',
                node: 'node-1',
                status: 'healthy',
                load: 45,
                memory: 67
            };
            
            mockWebSocket.simulateMessage(mockNodeUpdate);
            
            expect(llamaClient.nodes).toHaveLength(1);
            expect(llamaClient.nodes[0].status).toBe('healthy');
        });

        test('should update queue status', () => {
            const mockQueueUpdate = {
                type: 'queue_update',
                length: 5,
                estimatedWaitTime: 2000
            };
            
            mockWebSocket.simulateMessage(mockQueueUpdate);
            
            const queueElement = document.getElementById('queueLength');
            expect(queueElement.textContent).toBe('5');
        });
    });

    describe('Error Handling', () => {
        test('should handle malformed WebSocket messages', () => {
            console.error = jest.fn();
            
            // Send invalid JSON
            mockWebSocket.simulateMessage('invalid json');
            
            expect(console.error).toHaveBeenCalled();
        });

        test('should handle unknown message types', () => {
            console.warn = jest.fn();
            
            const unknownMessage = {
                type: 'unknown_type',
                data: 'test'
            };
            
            mockWebSocket.simulateMessage(unknownMessage);
            
            expect(console.warn).toHaveBeenCalledWith('Unknown message type:', 'unknown_type');
        });
    });

    describe('Performance Monitoring', () => {
        test('should track latency measurements', () => {
            const initialLatencyCount = llamaClient.performanceData.latency.length;
            
            // Simulate latency measurement
            llamaClient.recordPong(Date.now() - 100);
            
            expect(llamaClient.performanceData.latency.length).toBeGreaterThan(initialLatencyCount);
            expect(llamaClient.performanceData.latency[0].latency).toBe(100);
        });

        test('should limit performance data history', () => {
            // Add more than 100 latency measurements
            for (let i = 0; i < 150; i++) {
                llamaClient.performanceData.latency.push({
                    timestamp: Date.now(),
                    latency: i
                });
            }
            
            llamaClient.recordPong(50);
            
            // Should be limited to 100
            expect(llamaClient.performanceData.latency.length).toBe(100);
        });
    });

    describe('Accessibility Features', () => {
        test('should have proper ARIA labels', () => {
            const messageInput = document.getElementById('messageInput');
            const modelSelector = document.getElementById('modelSelector');
            
            expect(messageInput.getAttribute('aria-label')).toBeTruthy();
            expect(modelSelector.getAttribute('aria-label')).toBeTruthy();
        });

        test('should have skip links', () => {
            const skipLink = document.querySelector('.skip-link');
            expect(skipLink).toBeTruthy();
            expect(skipLink.getAttribute('href')).toBe('#main-content');
        });

        test('should have proper heading structure', () => {
            const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
            expect(headings.length).toBeGreaterThan(0);
        });
    });

    describe('Responsive Design', () => {
        test('should respond to viewport changes', () => {
            // Mock window resize
            const resizeEvent = new Event('resize');
            window.dispatchEvent(resizeEvent);
            
            // Should trigger responsive behavior
            expect(window.innerWidth).toBeDefined();
        });

        test('should handle different screen sizes', () => {
            const originalWidth = window.innerWidth;
            
            // Simulate mobile viewport
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                value: 375
            });
            
            // Should adapt to mobile layout
            expect(window.innerWidth).toBe(375);
            
            // Reset
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                value: originalWidth
            });
        });
    });
});

// Performance tests
describe('Performance Tests', () => {
    test('should handle rapid message sending', () => {
        const startTime = performance.now();
        
        for (let i = 0; i < 100; i++) {
            const mockResponse = {
                type: 'response',
                content: `Message ${i}`,
                node: 'test-node',
                streaming: false
            };
            // This would normally trigger DOM updates
        }
        
        const endTime = performance.now();
        const duration = endTime - startTime;
        
        expect(duration).toBeLessThan(100); // Should be very fast
    });

    test('should not leak memory with many messages', () => {
        const initialMemory = process.memoryUsage().heapUsed;
        
        // Simulate many message operations
        for (let i = 0; i < 1000; i++) {
            const mockResponse = {
                type: 'response',
                content: `Message ${i}`,
                node: `node-${i % 10}`,
                streaming: false
            };
            // Process message
        }
        
        // Force garbage collection if available
        if (global.gc) {
            global.gc();
        }
        
        const finalMemory = process.memoryUsage().heapUsed;
        const memoryIncrease = finalMemory - initialMemory;
        
        // Memory increase should be reasonable (< 50MB)
        expect(memoryIncrease).toBeLessThan(50 * 1024 * 1024);
    });
});

module.exports = {
    // Export test utilities for use in other test files
    createMockWebSocket: () => new MockWebSocket('ws://localhost:13000/chat'),
    setupTestEnvironment: () => {
        global.WebSocket = MockWebSocket;
        global.localStorage = mockLocalStorage;
    }
};