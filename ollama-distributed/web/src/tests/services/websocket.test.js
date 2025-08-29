// WebSocket Service Tests
// Comprehensive tests for WebSocket service functionality

import wsService from '../../services/websocket.js';

// Mock WebSocket
class MockWebSocket {
  constructor(url) {
    this.url = url;
    this.readyState = WebSocket.CONNECTING;
    this.onopen = null;
    this.onclose = null;
    this.onmessage = null;
    this.onerror = null;
    
    // Simulate connection after a short delay
    setTimeout(() => {
      this.readyState = WebSocket.OPEN;
      if (this.onopen) this.onopen();
    }, 10);
  }

  send(data) {
    if (this.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket is not open');
    }
    this.lastSentData = data;
  }

  close(code = 1000, reason = '') {
    this.readyState = WebSocket.CLOSED;
    if (this.onclose) {
      this.onclose({ code, reason, wasClean: code === 1000 });
    }
  }

  // Test helper methods
  simulateMessage(data) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) });
    }
  }

  simulateError(error) {
    if (this.onerror) {
      this.onerror(error);
    }
  }
}

// Mock WebSocket constants
MockWebSocket.CONNECTING = 0;
MockWebSocket.OPEN = 1;
MockWebSocket.CLOSING = 2;
MockWebSocket.CLOSED = 3;

global.WebSocket = MockWebSocket;

describe('WebSocket Service', () => {
  let mockWebSocket;

  beforeEach(() => {
    // Reset service state
    wsService.disconnect();
    
    // Mock window.location
    Object.defineProperty(window, 'location', {
      value: {
        hostname: 'localhost',
        protocol: 'http:'
      },
      writable: true
    });
  });

  afterEach(() => {
    wsService.disconnect();
  });

  describe('Connection Management', () => {
    test('should connect to WebSocket server', async () => {
      const connectPromise = new Promise(resolve => {
        wsService.on('connected', resolve);
      });

      wsService.connect('test-token');
      
      await connectPromise;
      
      expect(wsService.getStatus().isConnected).toBe(true);
      expect(wsService.getStatus().readyState).toBe(WebSocket.OPEN);
    });

    test('should send authentication on connection', async () => {
      const token = 'test-auth-token';
      
      wsService.connect(token);
      
      // Wait for connection
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });

      // Check if auth message was sent
      const connection = wsService.connection;
      expect(connection.lastSentData).toBe(JSON.stringify({
        type: 'auth',
        token: token
      }));
    });

    test('should handle connection errors', async () => {
      const errorPromise = new Promise(resolve => {
        wsService.on('error', resolve);
      });

      wsService.connect();
      
      // Simulate connection error
      setTimeout(() => {
        wsService.connection.simulateError(new Error('Connection failed'));
      }, 20);

      const error = await errorPromise;
      expect(error).toBeInstanceOf(Error);
    });

    test('should disconnect cleanly', () => {
      wsService.connect();
      
      const status = wsService.getStatus();
      expect(status.isConnected).toBe(false); // Initially connecting
      
      wsService.disconnect();
      
      const finalStatus = wsService.getStatus();
      expect(finalStatus.isConnected).toBe(false);
      expect(finalStatus.readyState).toBe(WebSocket.CLOSED);
    });
  });

  describe('Message Handling', () => {
    beforeEach(async () => {
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });
    });

    test('should handle incoming messages', async () => {
      const messagePromise = new Promise(resolve => {
        wsService.on('test_message', resolve);
      });

      const testData = { type: 'test_message', payload: 'test data' };
      wsService.connection.simulateMessage(testData);

      const receivedData = await messagePromise;
      expect(receivedData).toEqual(testData);
    });

    test('should handle heartbeat messages', async () => {
      const pongData = { type: 'pong', timestamp: Date.now() };
      
      // Send pong message
      wsService.connection.simulateMessage(pongData);
      
      // Check that lastHeartbeat was updated
      const status = wsService.getStatus();
      expect(status.lastHeartbeat).toBeGreaterThan(0);
    });

    test('should handle authentication responses', async () => {
      const authSuccessPromise = new Promise(resolve => {
        wsService.on('authenticated', resolve);
      });

      wsService.connection.simulateMessage({
        type: 'auth_response',
        success: true
      });

      await authSuccessPromise;
    });

    test('should handle authentication failures', async () => {
      const authFailPromise = new Promise(resolve => {
        wsService.on('auth_failed', resolve);
      });

      const errorMessage = 'Invalid token';
      wsService.connection.simulateMessage({
        type: 'auth_response',
        success: false,
        error: errorMessage
      });

      const error = await authFailPromise;
      expect(error).toBe(errorMessage);
    });

    test('should handle malformed messages gracefully', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      // Simulate malformed JSON
      if (wsService.connection.onmessage) {
        wsService.connection.onmessage({ data: 'invalid json' });
      }

      expect(consoleSpy).toHaveBeenCalledWith(
        'WebSocket message parsing error:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Message Sending', () => {
    beforeEach(async () => {
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });
    });

    test('should send messages when connected', () => {
      const testMessage = { type: 'test', data: 'hello' };
      
      wsService.send(testMessage);
      
      expect(wsService.connection.lastSentData).toBe(JSON.stringify(testMessage));
    });

    test('should queue messages when disconnected', () => {
      wsService.disconnect();
      
      const testMessage = { type: 'test', data: 'queued' };
      wsService.send(testMessage);
      
      const status = wsService.getStatus();
      expect(status.queuedMessages).toBe(1);
    });

    test('should flush queued messages on reconnection', async () => {
      // Disconnect and queue a message
      wsService.disconnect();
      const queuedMessage = { type: 'queued', data: 'test' };
      wsService.send(queuedMessage);
      
      // Reconnect
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });
      
      // Check that queued message was sent
      expect(wsService.connection.lastSentData).toBe(JSON.stringify(queuedMessage));
      expect(wsService.getStatus().queuedMessages).toBe(0);
    });

    test('should limit queue size', () => {
      wsService.disconnect();
      
      // Send more than 100 messages
      for (let i = 0; i < 150; i++) {
        wsService.send({ type: 'test', id: i });
      }
      
      const status = wsService.getStatus();
      expect(status.queuedMessages).toBe(100); // Should be limited to 100
    });
  });

  describe('Subscriptions', () => {
    beforeEach(async () => {
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });
    });

    test('should subscribe to data streams', () => {
      const callback = jest.fn();
      
      wsService.subscribe('metrics', callback);
      
      // Check that subscription message was sent
      expect(wsService.connection.lastSentData).toBe(JSON.stringify({
        type: 'subscribe',
        stream: 'metrics',
        timestamp: expect.any(Number)
      }));
    });

    test('should receive subscribed data', async () => {
      const callback = jest.fn();
      
      wsService.subscribe('metrics', callback);
      
      // Simulate receiving metrics data
      const metricsData = { type: 'metrics', cpu: 45, memory: 67 };
      wsService.connection.simulateMessage(metricsData);
      
      expect(callback).toHaveBeenCalledWith(metricsData);
    });

    test('should unsubscribe from data streams', () => {
      const callback = jest.fn();
      
      wsService.subscribe('metrics', callback);
      wsService.unsubscribe('metrics', callback);
      
      // Check that unsubscription message was sent
      expect(wsService.connection.lastSentData).toBe(JSON.stringify({
        type: 'unsubscribe',
        stream: 'metrics',
        timestamp: expect.any(Number)
      }));
    });

    test('should remove all listeners when unsubscribing without callback', () => {
      const callback1 = jest.fn();
      const callback2 = jest.fn();
      
      wsService.subscribe('metrics', callback1);
      wsService.subscribe('metrics', callback2);
      wsService.unsubscribe('metrics'); // No callback = remove all
      
      // Simulate receiving data
      const metricsData = { type: 'metrics', cpu: 45 };
      wsService.connection.simulateMessage(metricsData);
      
      expect(callback1).not.toHaveBeenCalled();
      expect(callback2).not.toHaveBeenCalled();
    });
  });

  describe('Heartbeat Mechanism', () => {
    beforeEach(async () => {
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });
    });

    test('should send ping messages periodically', async () => {
      // Wait for heartbeat interval
      await new Promise(resolve => setTimeout(resolve, 16000));
      
      // Check that ping was sent
      const lastMessage = JSON.parse(wsService.connection.lastSentData);
      expect(lastMessage.type).toBe('ping');
      expect(lastMessage.timestamp).toBeGreaterThan(0);
    });

    test('should close connection on heartbeat timeout', async () => {
      const disconnectPromise = new Promise(resolve => {
        wsService.on('disconnected', resolve);
      });

      // Set last heartbeat to old timestamp
      wsService.lastHeartbeat = Date.now() - 35000; // 35 seconds ago
      
      // Trigger heartbeat check
      await new Promise(resolve => setTimeout(resolve, 16000));
      
      // Connection should be closed due to timeout
      await disconnectPromise;
    });
  });

  describe('Reconnection Logic', () => {
    test('should attempt reconnection on unexpected disconnect', async () => {
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });

      const reconnectPromise = new Promise(resolve => {
        // Listen for second connection (reconnection)
        let connectionCount = 0;
        wsService.on('connected', () => {
          connectionCount++;
          if (connectionCount === 2) resolve();
        });
      });

      // Simulate unexpected disconnect
      wsService.connection.close(1006, 'Connection lost');
      
      await reconnectPromise;
    });

    test('should not reconnect on clean disconnect', async () => {
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });

      let reconnectionAttempted = false;
      wsService.on('connected', () => {
        reconnectionAttempted = true;
      });

      // Clean disconnect
      wsService.connection.close(1000, 'Normal closure');
      
      // Wait to see if reconnection is attempted
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      expect(reconnectionAttempted).toBe(false);
    });

    test('should limit reconnection attempts', async () => {
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });

      // Simulate multiple failed reconnections
      for (let i = 0; i < 6; i++) {
        wsService.connection.close(1006, 'Connection lost');
        await new Promise(resolve => setTimeout(resolve, 100));
      }

      const status = wsService.getStatus();
      expect(status.reconnectAttempts).toBeLessThanOrEqual(5);
    });
  });

  describe('Event Management', () => {
    test('should add and remove event listeners', () => {
      const callback = jest.fn();
      
      wsService.on('test_event', callback);
      wsService.emit('test_event', 'test data');
      
      expect(callback).toHaveBeenCalledWith('test data');
      
      wsService.off('test_event', callback);
      wsService.emit('test_event', 'test data 2');
      
      expect(callback).toHaveBeenCalledTimes(1); // Should not be called again
    });

    test('should handle errors in event listeners gracefully', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      const faultyCallback = () => { throw new Error('Callback error'); };
      
      wsService.on('test_event', faultyCallback);
      wsService.emit('test_event', 'test data');
      
      expect(consoleSpy).toHaveBeenCalledWith(
        "Error in WebSocket event listener for 'test_event':",
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Status Reporting', () => {
    test('should report correct connection status', () => {
      const initialStatus = wsService.getStatus();
      expect(initialStatus.isConnected).toBe(false);
      expect(initialStatus.readyState).toBe(WebSocket.CLOSED);
      expect(initialStatus.reconnectAttempts).toBe(0);
      expect(initialStatus.queuedMessages).toBe(0);
    });

    test('should update status on connection', async () => {
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });

      const status = wsService.getStatus();
      expect(status.isConnected).toBe(true);
      expect(status.readyState).toBe(WebSocket.OPEN);
    });
  });

  describe('Utility Methods', () => {
    beforeEach(async () => {
      wsService.connect();
      await new Promise(resolve => {
        wsService.on('connected', resolve);
      });
    });

    test('should request metrics', () => {
      wsService.requestMetrics(['system', 'cluster']);
      
      const lastMessage = JSON.parse(wsService.connection.lastSentData);
      expect(lastMessage.type).toBe('request_metrics');
      expect(lastMessage.metrics).toEqual(['system', 'cluster']);
    });

    test('should request node updates', () => {
      wsService.requestNodeUpdates();
      
      const lastMessage = JSON.parse(wsService.connection.lastSentData);
      expect(lastMessage.type).toBe('request_node_updates');
    });

    test('should send user activity', () => {
      const activity = { action: 'page_view', page: '/dashboard' };
      
      wsService.sendActivity(activity);
      
      const lastMessage = JSON.parse(wsService.connection.lastSentData);
      expect(lastMessage.type).toBe('user_activity');
      expect(lastMessage.activity).toEqual(activity);
    });
  });
});
