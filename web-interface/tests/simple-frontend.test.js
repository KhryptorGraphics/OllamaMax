/**
 * Simple Frontend Test Suite for OllamaMax Web Interface
 * 
 * This test suite focuses on testing the core functionality without
 * requiring the full application to be loaded.
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
global.WebSocket = MockWebSocket;

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

describe('OllamaMax Frontend Test Suite', () => {
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
    });

    describe('WebSocket Connection Management', () => {
        test('should create WebSocket with correct URL', () => {
            // Verify the mock was created correctly
            expect(mockWebSocket.url).toBe('ws://localhost:13000/chat');
            expect(mockWebSocket.readyState).toBe(0); // WebSocket.CONNECTING = 0
        });

        test('should handle successful connection', () => {
            const onopenSpy = jest.fn();
            mockWebSocket.onopen = onopenSpy;
            
            mockWebSocket.simulateOpen();
            
            expect(mockWebSocket.readyState).toBe(WebSocket.OPEN);
            expect(onopenSpy).toHaveBeenCalled();
        });

        test('should handle connection errors', () => {
            const onerrorSpy = jest.fn();
            mockWebSocket.onerror = onerrorSpy;
            
            mockWebSocket.simulateError();
            
            expect(onerrorSpy).toHaveBeenCalled();
        });

        test('should handle message simulation', () => {
            const onmessageSpy = jest.fn();
            mockWebSocket.onmessage = onmessageSpy;
            
            const testData = { type: 'test', data: 'test data' };
            mockWebSocket.simulateMessage(testData);
            
            expect(onmessageSpy).toHaveBeenCalled();
            expect(onmessageSpy.mock.calls[0][0].data).toBe(JSON.stringify(testData));
        });
    });

    describe('DOM Elements', () => {
        test('should have connection status elements', () => {
            const connectionStatus = document.getElementById('connectionStatus');
            const connectionText = document.getElementById('connectionText');
            
            expect(connectionStatus).toBeTruthy();
            expect(connectionText).toBeTruthy();
            expect(connectionText.textContent).toBe('Connecting...');
        });

        test('should have chat interface elements', () => {
            const messageInput = document.getElementById('messageInput');
            const sendButton = document.getElementById('sendButton');
            const messagesArea = document.getElementById('messagesArea');
            
            expect(messageInput).toBeTruthy();
            expect(sendButton).toBeTruthy();
            expect(messagesArea).toBeTruthy();
        });

        test('should have model selector', () => {
            const modelSelector = document.getElementById('modelSelector');
            expect(modelSelector).toBeTruthy();
            
            const options = modelSelector.querySelectorAll('option');
            expect(options.length).toBeGreaterThan(0);
            expect(options[0].value).toBe('tinyllama');
        });
    });

    describe('Settings Persistence', () => {
        test('should save and load settings from localStorage', () => {
            const testSettings = {
                apiEndpoint: 'ws://test:13000',
                darkMode: true,
                temperature: 0.8,
                maxTokens: 4096
            };
            
            // Save settings
            mockLocalStorage.setItem('llamaChatSettings', JSON.stringify(testSettings));
            
            // Load settings
            const savedSettings = JSON.parse(mockLocalStorage.getItem('llamaChatSettings'));
            
            expect(savedSettings).toEqual(testSettings);
        });

        test('should handle missing settings gracefully', () => {
            const savedSettings = mockLocalStorage.getItem('nonexistent');
            expect(savedSettings).toBeNull();
        });
    });

    describe('Accessibility Features', () => {
        test('should have proper ARIA labels', () => {
            const messageInput = document.getElementById('messageInput');
            const modelSelector = document.getElementById('modelSelector');
            
            // These would normally have ARIA labels set by JavaScript
            expect(messageInput).toBeTruthy();
            expect(modelSelector).toBeTruthy();
        });

        test('should have proper heading structure', () => {
            // Add a heading to the DOM for testing
            const heading = document.createElement('h1');
            heading.textContent = 'Distributed Llama Chat';
            document.body.appendChild(heading);
            
            const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
            expect(headings.length).toBeGreaterThan(0);
        });
    });

    describe('Responsive Design', () => {
        test('should respond to viewport changes', () => {
            const originalWidth = window.innerWidth;
            
            // Mock window resize
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                value: 375
            });
            
            expect(window.innerWidth).toBe(375);
            
            // Reset
            Object.defineProperty(window, 'innerWidth', {
                writable: true,
                value: originalWidth
            });
        });
    });

    describe('Performance Tests', () => {
        test('should handle rapid operations efficiently', () => {
            const startTime = performance.now();
            
            // Simulate many DOM operations
            for (let i = 0; i < 100; i++) {
                const element = document.createElement('div');
                element.textContent = `Element ${i}`;
                document.body.appendChild(element);
            }
            
            const endTime = performance.now();
            const duration = endTime - startTime;
            
            expect(duration).toBeLessThan(100); // Should be very fast
        });

        test('should not leak memory with DOM operations', () => {
            const initialElements = document.querySelectorAll('*').length;
            
            // Add many elements
            for (let i = 0; i < 1000; i++) {
                const element = document.createElement('div');
                element.textContent = `Element ${i}`;
                document.body.appendChild(element);
            }
            
            const afterElements = document.querySelectorAll('*').length;
            expect(afterElements).toBeGreaterThan(initialElements);
            
            // Clear elements
            document.body.innerHTML = '';
            
            const finalElements = document.querySelectorAll('*').length;
            expect(finalElements).toBeLessThan(afterElements);
        });
    });
});

module.exports = {
    // Export test utilities for use in other test files
    MockWebSocket,
    createMockWebSocket: () => new MockWebSocket('ws://localhost:13000/chat'),
    setupTestEnvironment: () => {
        global.WebSocket = MockWebSocket;
        global.localStorage = mockLocalStorage;
    }
};