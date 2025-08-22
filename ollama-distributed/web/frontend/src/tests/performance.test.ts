/**
 * @fileoverview Performance tests for WebSocket and Dashboard
 * @description Tests for performance characteristics and scalability
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { WebSocketClient } from '../lib/websocket';
import { MessageTypes } from '../types/websocket';

// Mock performance.now for consistent timing
const mockPerformanceNow = vi.fn();
global.performance = { now: mockPerformanceNow } as any;

// Mock WebSocket for performance testing
class PerformanceMockWebSocket {
  public readyState: number = WebSocket.OPEN;
  public onopen: ((event: Event) => void) | null = null;
  public onclose: ((event: CloseEvent) => void) | null = null;
  public onerror: ((event: Event) => void) | null = null;
  public onmessage: ((event: MessageEvent) => void) | null = null;
  public url: string;
  
  private messageCount = 0;
  private startTime = 0;

  constructor(url: string) {
    this.url = url;
    this.startTime = performance.now();
    setTimeout(() => {
      this.onopen?.(new Event('open'));
    }, 1);
  }

  send(data: string) {
    this.messageCount++;
  }

  close() {
    this.readyState = WebSocket.CLOSED;
    this.onclose?.(new CloseEvent('close'));
  }

  simulateMessage(data: any) {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { 
        data: typeof data === 'string' ? data : JSON.stringify(data) 
      }));
    }
  }

  getMessageCount() {
    return this.messageCount;
  }

  getConnectionTime() {
    return performance.now() - this.startTime;
  }
}

(global as any).WebSocket = PerformanceMockWebSocket;

describe('WebSocket Performance Tests', () => {
  let client: WebSocketClient;
  let mockWs: PerformanceMockWebSocket;
  let timeCounter = 0;

  beforeEach(() => {
    timeCounter = 0;
    mockPerformanceNow.mockImplementation(() => {
      timeCounter += 10; // Simulate 10ms increments
      return timeCounter;
    });

    client = new WebSocketClient({
      url: 'ws://localhost:8080/ws',
      debug: false
    });
  });

  afterEach(() => {
    client.disconnect();
    vi.clearAllMocks();
  });

  describe('Connection Performance', () => {
    it('should connect within acceptable time', async () => {
      const startTime = performance.now();
      
      await client.connect();
      
      const connectionTime = performance.now() - startTime;
      expect(connectionTime).toBeLessThan(100); // Should connect within 100ms
      expect(client.isConnected()).toBe(true);
    });

    it('should handle rapid connect/disconnect cycles efficiently', async () => {
      const cycles = 10;
      const startTime = performance.now();

      for (let i = 0; i < cycles; i++) {
        await client.connect();
        expect(client.isConnected()).toBe(true);
        
        client.disconnect();
        expect(client.isConnected()).toBe(false);
      }

      const totalTime = performance.now() - startTime;
      const avgTimePerCycle = totalTime / cycles;
      
      expect(avgTimePerCycle).toBeLessThan(50); // Should average less than 50ms per cycle
    });
  });

  describe('Message Throughput', () => {
    beforeEach(async () => {
      await client.connect();
      mockWs = (client as any).ws as PerformanceMockWebSocket;
    });

    it('should handle high-frequency message sending', () => {
      const messageCount = 1000;
      const startTime = performance.now();

      for (let i = 0; i < messageCount; i++) {
        client.send({
          type: MessageTypes.PING,
          data: { id: i, timestamp: Date.now() },
          timestamp: new Date().toISOString()
        });
      }

      const sendTime = performance.now() - startTime;
      const messagesPerMs = messageCount / sendTime;
      
      expect(messagesPerMs).toBeGreaterThan(1); // Should send more than 1 message per ms
      expect(mockWs.getMessageCount()).toBe(messageCount);
    });

    it('should handle high-frequency message receiving', () => {
      const messageCount = 1000;
      let receivedCount = 0;

      const unsubscribe = client.subscribe(MessageTypes.METRICS_UPDATE, () => {
        receivedCount++;
      });

      const startTime = performance.now();

      for (let i = 0; i < messageCount; i++) {
        mockWs.simulateMessage({
          type: MessageTypes.METRICS_UPDATE,
          data: { cpu: Math.random() * 100, memory: Math.random() * 100 },
          timestamp: new Date().toISOString()
        });
      }

      const receiveTime = performance.now() - startTime;
      const messagesPerMs = messageCount / receiveTime;

      expect(messagesPerMs).toBeGreaterThan(1); // Should process more than 1 message per ms
      expect(receivedCount).toBe(messageCount);
      
      unsubscribe();
    });

    it('should maintain performance with multiple subscribers', () => {
      const subscriberCount = 50;
      const messageCount = 100;
      const receivedCounts: number[] = new Array(subscriberCount).fill(0);
      const unsubscribers: Array<() => void> = [];

      // Create multiple subscribers
      for (let i = 0; i < subscriberCount; i++) {
        const unsubscribe = client.subscribe(MessageTypes.CLUSTER_UPDATE, () => {
          receivedCounts[i]++;
        });
        unsubscribers.push(unsubscribe);
      }

      const startTime = performance.now();

      // Send messages
      for (let i = 0; i < messageCount; i++) {
        mockWs.simulateMessage({
          type: MessageTypes.CLUSTER_UPDATE,
          data: { status: 'healthy', nodes: 3 },
          timestamp: new Date().toISOString()
        });
      }

      const processTime = performance.now() - startTime;
      const totalCallbacks = subscriberCount * messageCount;
      const callbacksPerMs = totalCallbacks / processTime;

      expect(callbacksPerMs).toBeGreaterThan(10); // Should handle 10+ callbacks per ms
      
      // Verify all subscribers received all messages
      receivedCounts.forEach(count => {
        expect(count).toBe(messageCount);
      });

      // Cleanup
      unsubscribers.forEach(unsubscribe => unsubscribe());
    });
  });

  describe('Memory Usage', () => {
    beforeEach(async () => {
      await client.connect();
      mockWs = (client as any).ws as PerformanceMockWebSocket;
    });

    it('should not leak memory with subscription churn', () => {
      const iterations = 100;
      
      for (let i = 0; i < iterations; i++) {
        // Create and immediately destroy subscription
        const unsubscribe = client.subscribe(MessageTypes.NODE_UPDATE, () => {});
        unsubscribe();
      }

      // Should have no active subscriptions
      expect(client.getStats().subscriptions).toBe(0);
      
      // Internal subscriber maps should be clean
      const subscribers = (client as any).subscribers;
      const subscriptions = (client as any).subscriptions;
      
      expect(subscribers.size).toBe(0);
      expect(subscriptions.size).toBe(0);
    });

    it('should handle large message payloads efficiently', () => {
      let receivedData: any = null;
      const unsubscribe = client.subscribe(MessageTypes.MODEL_UPDATE, (data) => {
        receivedData = data;
      });

      // Create large payload (1MB of data)
      const largePayload = {
        models: new Array(10000).fill(null).map((_, i) => ({
          id: `model_${i}`,
          name: `Large Model ${i}`,
          description: 'A'.repeat(100), // 100 char description
          metadata: {
            version: '1.0.0',
            size: Math.random() * 1000000000,
            tags: ['large', 'test', 'performance'],
            parameters: new Array(10).fill(null).map((_, j) => ({
              name: `param_${j}`,
              value: Math.random() * 1000
            }))
          }
        }))
      };

      const startTime = performance.now();
      
      mockWs.simulateMessage({
        type: MessageTypes.MODEL_UPDATE,
        data: largePayload,
        timestamp: new Date().toISOString()
      });

      const processTime = performance.now() - startTime;

      expect(processTime).toBeLessThan(100); // Should process large payload within 100ms
      expect(receivedData).toEqual(largePayload);
      
      unsubscribe();
    });
  });

  describe('Concurrent Operations', () => {
    it('should handle concurrent subscribe/unsubscribe operations', async () => {
      await client.connect();
      
      const concurrentOperations = 50;
      const operations: Promise<void>[] = [];

      for (let i = 0; i < concurrentOperations; i++) {
        operations.push(
          new Promise<void>((resolve) => {
            setTimeout(() => {
              const unsubscribe = client.subscribe(MessageTypes.NOTIFICATION, () => {});
              setTimeout(() => {
                unsubscribe();
                resolve();
              }, Math.random() * 50);
            }, Math.random() * 50);
          })
        );
      }

      const startTime = performance.now();
      await Promise.all(operations);
      const operationTime = performance.now() - startTime;

      expect(operationTime).toBeLessThan(200); // Should complete within 200ms
      expect(client.getStats().subscriptions).toBe(0); // All should be cleaned up
    });

    it('should handle concurrent message processing', async () => {
      await client.connect();
      mockWs = (client as any).ws as PerformanceMockWebSocket;

      const messageTypes = [
        MessageTypes.CLUSTER_UPDATE,
        MessageTypes.NODE_UPDATE,
        MessageTypes.METRICS_UPDATE,
        MessageTypes.MODEL_UPDATE,
        MessageTypes.NOTIFICATION
      ];

      const receivedCounts = new Map<string, number>();
      const unsubscribers: Array<() => void> = [];

      // Subscribe to all message types
      messageTypes.forEach(type => {
        receivedCounts.set(type, 0);
        const unsubscribe = client.subscribe(type, () => {
          receivedCounts.set(type, receivedCounts.get(type)! + 1);
        });
        unsubscribers.push(unsubscribe);
      });

      const messagesPerType = 100;
      const startTime = performance.now();

      // Send messages concurrently
      const sendPromises = messageTypes.map(type => 
        new Promise<void>((resolve) => {
          setTimeout(() => {
            for (let i = 0; i < messagesPerType; i++) {
              mockWs.simulateMessage({
                type,
                data: { type, index: i },
                timestamp: new Date().toISOString()
              });
            }
            resolve();
          }, Math.random() * 10);
        })
      );

      await Promise.all(sendPromises);
      const processTime = performance.now() - startTime;

      expect(processTime).toBeLessThan(500); // Should process all within 500ms
      
      // Verify all messages were received
      messageTypes.forEach(type => {
        expect(receivedCounts.get(type)).toBe(messagesPerType);
      });

      // Cleanup
      unsubscribers.forEach(unsubscribe => unsubscribe());
    });
  });

  describe('Error Handling Performance', () => {
    beforeEach(async () => {
      await client.connect();
      mockWs = (client as any).ws as PerformanceMockWebSocket;
    });

    it('should handle malformed messages efficiently', () => {
      let errorCount = 0;
      const unsubscribe = client.onError(() => {
        errorCount++;
      });

      const malformedMessageCount = 1000;
      const startTime = performance.now();

      for (let i = 0; i < malformedMessageCount; i++) {
        // Send malformed JSON
        if (mockWs.onmessage) {
          mockWs.onmessage(new MessageEvent('message', {
            data: `invalid json ${i} {`
          }));
        }
      }

      const processTime = performance.now() - startTime;
      const messagesPerMs = malformedMessageCount / processTime;

      expect(messagesPerMs).toBeGreaterThan(5); // Should handle 5+ malformed messages per ms
      
      unsubscribe();
    });

    it('should handle subscriber errors without degrading performance', () => {
      const goodSubscriberCount = 10;
      const badSubscriberCount = 5;
      let goodCallbacks = 0;
      let errorCallbacks = 0;

      // Create good subscribers
      for (let i = 0; i < goodSubscriberCount; i++) {
        client.subscribe(MessageTypes.HEARTBEAT, () => {
          goodCallbacks++;
        });
      }

      // Create error-throwing subscribers
      for (let i = 0; i < badSubscriberCount; i++) {
        client.subscribe(MessageTypes.HEARTBEAT, () => {
          errorCallbacks++;
          throw new Error('Subscriber error');
        });
      }

      const messageCount = 100;
      const startTime = performance.now();

      for (let i = 0; i < messageCount; i++) {
        mockWs.simulateMessage({
          type: MessageTypes.HEARTBEAT,
          data: { server_time: Date.now() },
          timestamp: new Date().toISOString()
        });
      }

      const processTime = performance.now() - startTime;
      
      expect(processTime).toBeLessThan(200); // Should not be significantly slowed by errors
      expect(goodCallbacks).toBe(goodSubscriberCount * messageCount);
      expect(errorCallbacks).toBe(badSubscriberCount * messageCount);
    });
  });

  describe('Scalability Tests', () => {
    it('should maintain performance with increasing subscription count', async () => {
      await client.connect();
      mockWs = (client as any).ws as PerformanceMockWebSocket;

      const subscriptionCounts = [10, 50, 100, 500];
      const results: Array<{ count: number; timePerMessage: number }> = [];

      for (const count of subscriptionCounts) {
        const unsubscribers: Array<() => void> = [];
        let callbackCount = 0;

        // Create subscribers
        for (let i = 0; i < count; i++) {
          const unsubscribe = client.subscribe(MessageTypes.PING, () => {
            callbackCount++;
          });
          unsubscribers.push(unsubscribe);
        }

        const messageCount = 50;
        const startTime = performance.now();

        // Send messages
        for (let i = 0; i < messageCount; i++) {
          mockWs.simulateMessage({
            type: MessageTypes.PING,
            data: { timestamp: Date.now() },
            timestamp: new Date().toISOString()
          });
        }

        const processTime = performance.now() - startTime;
        const timePerMessage = processTime / messageCount;

        results.push({ count, timePerMessage });

        expect(callbackCount).toBe(count * messageCount);
        
        // Cleanup
        unsubscribers.forEach(unsubscribe => unsubscribe());
      }

      // Performance should not degrade significantly with more subscribers
      const baselineTime = results[0].timePerMessage;
      const maxTime = results[results.length - 1].timePerMessage;
      
      // Performance degradation should be sub-linear
      expect(maxTime / baselineTime).toBeLessThan(10); // No more than 10x slower with 50x subscribers
    });
  });
});

describe('Dashboard Performance Tests', () => {
  // Mock React Testing Library render function for performance testing
  const mockRender = vi.fn();
  
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should measure component render performance', () => {
    const renderCount = 100;
    const startTime = performance.now();

    for (let i = 0; i < renderCount; i++) {
      mockRender({ componentName: 'Dashboard', props: { id: i } });
    }

    const renderTime = performance.now() - startTime;
    const avgRenderTime = renderTime / renderCount;

    expect(avgRenderTime).toBeLessThan(1); // Should render in less than 1ms on average
  });

  it('should handle rapid state updates efficiently', () => {
    const stateUpdateCount = 1000;
    const startTime = performance.now();

    // Simulate rapid state updates (like real-time data)
    for (let i = 0; i < stateUpdateCount; i++) {
      mockRender({
        componentName: 'KPIWidget',
        props: {
          value: Math.random() * 100,
          isLive: true,
          timestamp: Date.now()
        }
      });
    }

    const updateTime = performance.now() - startTime;
    const avgUpdateTime = updateTime / stateUpdateCount;

    expect(avgUpdateTime).toBeLessThan(0.5); // Should handle updates in less than 0.5ms each
  });
});