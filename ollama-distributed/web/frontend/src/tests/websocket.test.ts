/**
 * @fileoverview WebSocket client unit tests
 * @description Comprehensive test suite for WebSocket client reliability and functionality
 */

import { describe, it, expect, beforeEach, afterEach, vi, Mock } from 'vitest';
import { WebSocketClient } from '../lib/websocket';
import { ConnectionState, MessageTypes } from '../types/websocket';

// Mock WebSocket
class MockWebSocket {
  public readyState: number = WebSocket.CONNECTING;
  public onopen: ((event: Event) => void) | null = null;
  public onclose: ((event: CloseEvent) => void) | null = null;
  public onerror: ((event: Event) => void) | null = null;
  public onmessage: ((event: MessageEvent) => void) | null = null;
  public url: string;
  
  private messageQueue: string[] = [];

  constructor(url: string) {
    this.url = url;
    // Simulate async connection
    setTimeout(() => {
      this.readyState = WebSocket.OPEN;
      this.onopen?.(new Event('open'));
    }, 10);
  }

  send(data: string) {
    if (this.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket is not open');
    }
    this.messageQueue.push(data);
  }

  close(code?: number, reason?: string) {
    this.readyState = WebSocket.CLOSED;
    this.onclose?.(new CloseEvent('close', { code, reason }));
  }

  // Test helper methods
  simulateMessage(data: any) {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { 
        data: typeof data === 'string' ? data : JSON.stringify(data) 
      }));
    }
  }

  simulateError() {
    this.onerror?.(new Event('error'));
  }

  getLastMessage() {
    return this.messageQueue[this.messageQueue.length - 1];
  }

  getAllMessages() {
    return [...this.messageQueue];
  }

  clearMessages() {
    this.messageQueue = [];
  }
}

// Mock global WebSocket
(global as any).WebSocket = MockWebSocket;

describe('WebSocketClient', () => {
  let client: WebSocketClient;
  let mockWs: MockWebSocket;

  beforeEach(() => {
    client = new WebSocketClient({
      url: 'ws://localhost:8080/ws',
      reconnectAttempts: 3,
      reconnectInterval: 100,
      heartbeatInterval: 1000,
      debug: false
    });

    // Access the private ws property for testing
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
    vi.spyOn(console, 'warn').mockImplementation(() => {});
  });

  afterEach(() => {
    client.disconnect();
    vi.restoreAllMocks();
    vi.clearAllTimers();
  });

  describe('Connection Management', () => {
    it('should connect successfully', async () => {
      const promise = client.connect();
      
      // Wait for connection
      await promise;
      
      expect(client.isConnected()).toBe(true);
      expect(client.getState()).toBe(ConnectionState.CONNECTED);
    });

    it('should handle connection errors', async () => {
      // Create error handler
      let errorReceived: Error | null = null;
      const unsubscribe = client.onError((error) => {
        errorReceived = error;
      });

      try {
        // Start connection and immediately simulate error
        const promise = client.connect();
        
        // Access private ws to simulate error
        setTimeout(() => {
          const ws = (client as any).ws as MockWebSocket;
          if (ws) {
            ws.simulateError();
          }
        }, 5);

        await expect(promise).rejects.toThrow();
        expect(errorReceived).toBeInstanceOf(Error);
      } finally {
        unsubscribe();
      }
    });

    it('should disconnect gracefully', async () => {
      await client.connect();
      expect(client.isConnected()).toBe(true);
      
      client.disconnect();
      expect(client.isConnected()).toBe(false);
      expect(client.getState()).toBe(ConnectionState.DISCONNECTED);
    });

    it('should not connect when already connected', async () => {
      await client.connect();
      expect(client.isConnected()).toBe(true);
      
      // Second connect should return immediately
      await client.connect();
      expect(client.isConnected()).toBe(true);
    });
  });

  describe('Message Handling', () => {
    beforeEach(async () => {
      await client.connect();
      mockWs = (client as any).ws as MockWebSocket;
    });

    it('should send messages when connected', () => {
      const message = {
        type: MessageTypes.PING,
        data: { test: 'data' },
        timestamp: new Date().toISOString()
      };

      client.send(message);
      
      const sentMessage = mockWs.getLastMessage();
      expect(JSON.parse(sentMessage)).toEqual(message);
    });

    it('should not send messages when disconnected', () => {
      client.disconnect();
      
      const message = {
        type: MessageTypes.PING,
        data: { test: 'data' },
        timestamp: new Date().toISOString()
      };

      client.send(message);
      
      // No message should be sent
      expect(mockWs.getAllMessages()).toHaveLength(0);
    });

    it('should receive and parse messages correctly', () => {
      let receivedData: any = null;
      const unsubscribe = client.subscribe(MessageTypes.CLUSTER_UPDATE, (data) => {
        receivedData = data;
      });

      const testData = { status: 'healthy', nodes: 3 };
      mockWs.simulateMessage({
        type: MessageTypes.CLUSTER_UPDATE,
        data: testData,
        timestamp: new Date().toISOString()
      });

      expect(receivedData).toEqual(testData);
      unsubscribe();
    });

    it('should handle malformed messages gracefully', () => {
      let receivedData: any = null;
      const unsubscribe = client.subscribe(MessageTypes.CLUSTER_UPDATE, (data) => {
        receivedData = data;
      });

      // Send malformed JSON
      mockWs.simulateMessage('invalid json {');

      // Should not crash and should not call subscriber
      expect(receivedData).toBeNull();
      unsubscribe();
    });
  });

  describe('Subscription Management', () => {
    beforeEach(async () => {
      await client.connect();
      mockWs = (client as any).ws as MockWebSocket;
    });

    it('should subscribe to topics', () => {
      let callCount = 0;
      const unsubscribe = client.subscribe(MessageTypes.NODE_UPDATE, () => {
        callCount++;
      });

      mockWs.simulateMessage({
        type: MessageTypes.NODE_UPDATE,
        data: { node_id: 'test' },
        timestamp: new Date().toISOString()
      });

      expect(callCount).toBe(1);
      unsubscribe();
    });

    it('should unsubscribe from topics', () => {
      let callCount = 0;
      const unsubscribe = client.subscribe(MessageTypes.NODE_UPDATE, () => {
        callCount++;
      });

      // First message should be received
      mockWs.simulateMessage({
        type: MessageTypes.NODE_UPDATE,
        data: { node_id: 'test1' },
        timestamp: new Date().toISOString()
      });

      expect(callCount).toBe(1);

      // Unsubscribe
      unsubscribe();

      // Second message should not be received
      mockWs.simulateMessage({
        type: MessageTypes.NODE_UPDATE,
        data: { node_id: 'test2' },
        timestamp: new Date().toISOString()
      });

      expect(callCount).toBe(1);
    });

    it('should handle multiple subscribers to same topic', () => {
      let callCount1 = 0;
      let callCount2 = 0;

      const unsubscribe1 = client.subscribe(MessageTypes.METRICS_UPDATE, () => {
        callCount1++;
      });

      const unsubscribe2 = client.subscribe(MessageTypes.METRICS_UPDATE, () => {
        callCount2++;
      });

      mockWs.simulateMessage({
        type: MessageTypes.METRICS_UPDATE,
        data: { cpu: 50 },
        timestamp: new Date().toISOString()
      });

      expect(callCount1).toBe(1);
      expect(callCount2).toBe(1);

      unsubscribe1();
      unsubscribe2();
    });

    it('should send subscription message to server', () => {
      mockWs.clearMessages();
      
      client.subscribe(MessageTypes.MODEL_UPDATE, () => {});

      const messages = mockWs.getAllMessages();
      expect(messages).toHaveLength(1);
      
      const subscribeMessage = JSON.parse(messages[0]);
      expect(subscribeMessage.type).toBe(MessageTypes.SUBSCRIBE);
      expect(subscribeMessage.data.topic).toBe(MessageTypes.MODEL_UPDATE);
    });
  });

  describe('Reconnection Logic', () => {
    it('should attempt reconnection on unexpected disconnect', async () => {
      await client.connect();
      expect(client.isConnected()).toBe(true);

      let stateChanges: ConnectionState[] = [];
      const unsubscribe = client.onStateChange((state: any) => {
        stateChanges.push(state);
      });

      // Simulate unexpected disconnect
      const ws = (client as any).ws as MockWebSocket;
      ws.close(1006, 'Connection lost');

      // Wait for reconnection state
      await new Promise(resolve => setTimeout(resolve, 50));

      expect(stateChanges).toContain(ConnectionState.RECONNECTING);
      unsubscribe();
    });

    it('should not reconnect on manual disconnect', async () => {
      await client.connect();
      expect(client.isConnected()).toBe(true);

      let stateChanges: ConnectionState[] = [];
      const unsubscribe = client.onStateChange((state: any) => {
        stateChanges.push(state);
      });

      client.disconnect();

      // Wait a bit to ensure no reconnection attempt
      await new Promise(resolve => setTimeout(resolve, 200));

      expect(stateChanges).not.toContain(ConnectionState.RECONNECTING);
      expect(client.getState()).toBe(ConnectionState.DISCONNECTED);
      unsubscribe();
    });

    it('should respect maximum reconnection attempts', async () => {
      // Create client with max 2 attempts
      const limitedClient = new WebSocketClient({
        url: 'ws://localhost:8080/ws',
        reconnectAttempts: 2,
        reconnectInterval: 50,
        debug: false
      });

      let stateChanges: ConnectionState[] = [];
      const unsubscribe = limitedClient.onStateChange((state: any) => {
        stateChanges.push(state);
      });

      try {
        await limitedClient.connect();
        
        // Simulate repeated disconnects
        for (let i = 0; i < 3; i++) {
          const ws = (limitedClient as any).ws as MockWebSocket;
          if (ws) {
            ws.close(1006, 'Connection lost');
            await new Promise(resolve => setTimeout(resolve, 100));
          }
        }

        // Should eventually give up and stay disconnected
        await new Promise(resolve => setTimeout(resolve, 200));
        expect(limitedClient.getState()).toBe(ConnectionState.DISCONNECTED);
      } finally {
        limitedClient.disconnect();
        unsubscribe();
      }
    });
  });

  describe('Heartbeat and Health Monitoring', () => {
    beforeEach(async () => {
      await client.connect();
      mockWs = (client as any).ws as MockWebSocket;
    });

    it('should handle heartbeat messages', () => {
      const serverTime = Date.now() / 1000;
      mockWs.simulateMessage({
        type: MessageTypes.HEARTBEAT,
        data: { server_time: serverTime },
        timestamp: new Date().toISOString()
      });

      const timeOffset = client.getServerTimeOffset();
      expect(Math.abs(timeOffset)).toBeLessThan(1); // Should be close to 0
    });

    it('should handle pong messages', () => {
      // Set a ping time
      (client as any).lastPingTime = Date.now() - 100;

      mockWs.simulateMessage({
        type: MessageTypes.PONG,
        data: { timestamp: Date.now() },
        timestamp: new Date().toISOString()
      });

      // Should not throw and should handle gracefully
      expect(client.getStats().lastPingTime).toBeGreaterThan(0);
    });
  });

  describe('Statistics and State', () => {
    it('should provide connection statistics', () => {
      const stats = client.getStats();
      
      expect(stats).toHaveProperty('state');
      expect(stats).toHaveProperty('reconnectAttempts');
      expect(stats).toHaveProperty('subscriptions');
      expect(stats).toHaveProperty('serverTimeOffset');
      expect(stats).toHaveProperty('lastPingTime');
    });

    it('should track subscription count', () => {
      const unsubscribe1 = client.subscribe(MessageTypes.CLUSTER_UPDATE, () => {});
      const unsubscribe2 = client.subscribe(MessageTypes.NODE_UPDATE, () => {});

      expect(client.getStats().subscriptions).toBe(2);

      unsubscribe1();
      expect(client.getStats().subscriptions).toBe(1);

      unsubscribe2();
      expect(client.getStats().subscriptions).toBe(0);
    });
  });

  describe('Error Handling', () => {
    it('should handle subscriber callback errors gracefully', async () => {
      await client.connect();
      mockWs = (client as any).ws as MockWebSocket;

      // Subscribe with a callback that throws
      const unsubscribe = client.subscribe(MessageTypes.NOTIFICATION, () => {
        throw new Error('Subscriber error');
      });

      // Should not crash when message is received
      expect(() => {
        mockWs.simulateMessage({
          type: MessageTypes.NOTIFICATION,
          data: { message: 'test' },
          timestamp: new Date().toISOString()
        });
      }).not.toThrow();

      unsubscribe();
    });

    it('should handle state listener errors gracefully', () => {
      const unsubscribe = client.onStateChange(() => {
        throw new Error('State listener error');
      });

      // Should not crash when state changes
      expect(() => {
        client.disconnect();
      }).not.toThrow();

      unsubscribe();
    });

    it('should handle error listener errors gracefully', () => {
      const unsubscribe = client.onError(() => {
        throw new Error('Error listener error');
      });

      // Should not crash when error is emitted
      expect(() => {
        (client as any).emitError(new Error('Test error'));
      }).not.toThrow();

      unsubscribe();
    });
  });

  describe('Edge Cases', () => {
    it('should handle rapid connect/disconnect cycles', async () => {
      for (let i = 0; i < 5; i++) {
        await client.connect();
        expect(client.isConnected()).toBe(true);
        
        client.disconnect();
        expect(client.isConnected()).toBe(false);
      }
    });

    it('should handle subscription during disconnected state', () => {
      expect(client.isConnected()).toBe(false);
      
      // Should not throw
      const unsubscribe = client.subscribe(MessageTypes.CLUSTER_UPDATE, () => {});
      expect(typeof unsubscribe).toBe('function');
      
      unsubscribe();
    });

    it('should resubscribe after reconnection', async () => {
      await client.connect();
      mockWs = (client as any).ws as MockWebSocket;
      
      // Subscribe to a topic
      let messageCount = 0;
      const unsubscribe = client.subscribe(MessageTypes.MODEL_UPDATE, () => {
        messageCount++;
      });

      // Clear initial subscription message
      mockWs.clearMessages();

      // Simulate disconnect and reconnect
      mockWs.close(1006, 'Connection lost');
      
      // Wait for reconnection
      await new Promise(resolve => setTimeout(resolve, 150));

      // Should have sent resubscription
      const messages = mockWs.getAllMessages();
      const resubscribeMessage = messages.find(msg => {
        const parsed = JSON.parse(msg);
        return parsed.type === MessageTypes.SUBSCRIBE && 
               parsed.data.topic === MessageTypes.MODEL_UPDATE;
      });

      expect(resubscribeMessage).toBeDefined();
      unsubscribe();
    });
  });
});

describe('WebSocket Client Factory Functions', () => {
  it('should create client instance', () => {
    const { createWebSocketClient } = require('../lib/websocket');
    
    const client = createWebSocketClient({
      url: 'ws://localhost:8080/ws',
      debug: true
    });

    expect(client).toBeInstanceOf(WebSocketClient);
  });

  it('should return singleton instance', () => {
    const { getWebSocketClient, createWebSocketClient } = require('../lib/websocket');
    
    const client1 = createWebSocketClient({
      url: 'ws://localhost:8080/ws'
    });

    const client2 = getWebSocketClient();
    
    expect(client1).toBe(client2);
  });
});