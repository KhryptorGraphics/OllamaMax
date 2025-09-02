import { Page, APIRequestContext } from '@playwright/test';
import fs from 'fs/promises';
import path from 'path';

/**
 * Load Testing Helper for OllamaMax Platform
 * 
 * Provides utilities for:
 * - Concurrent request testing
 * - API load testing
 * - WebSocket stress testing
 * - Resource consumption monitoring
 * - Throughput measurement
 */

export interface LoadTestResult {
  testName: string;
  duration: number;
  totalRequests: number;
  successfulRequests: number;
  failedRequests: number;
  averageResponseTime: number;
  minResponseTime: number;
  maxResponseTime: number;
  requestsPerSecond: number;
  errors: Array<{ error: string; count: number }>;
  responseTimePercentiles: {
    p50: number;
    p90: number;
    p95: number;
    p99: number;
  };
}

export interface ConcurrentTestOptions {
  concurrentUsers: number;
  requestsPerUser: number;
  rampUpTime?: number;
  endpoint?: string;
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  payload?: any;
  headers?: Record<string, string>;
}

export class LoadTestHelper {
  constructor(private page: Page, private request?: APIRequestContext) {}

  /**
   * Run concurrent API load test
   */
  async runConcurrentAPITest(options: ConcurrentTestOptions): Promise<LoadTestResult> {
    const {
      concurrentUsers,
      requestsPerUser,
      rampUpTime = 0,
      endpoint = '/api/v1/health',
      method = 'GET',
      payload,
      headers = { 'Content-Type': 'application/json' }
    } = options;
    
    const startTime = Date.now();
    const responseTimes: number[] = [];
    const errors: { [key: string]: number } = {};
    let successCount = 0;
    let failCount = 0;
    
    console.log(`Starting load test: ${concurrentUsers} users, ${requestsPerUser} requests each`);
    
    // Create user sessions
    const userPromises: Promise<void>[] = [];
    
    for (let user = 0; user < concurrentUsers; user++) {
      const userPromise = this.runUserSession(user, requestsPerUser, endpoint, method, payload, headers, rampUpTime)
        .then((results) => {
          successCount += results.successes;
          failCount += results.failures;
          responseTimes.push(...results.responseTimes);
          
          // Aggregate errors
          for (const [error, count] of Object.entries(results.errors)) {
            errors[error] = (errors[error] || 0) + count;
          }
        });
      
      userPromises.push(userPromise);
      
      // Ramp up delay
      if (rampUpTime > 0 && user < concurrentUsers - 1) {
        await this.sleep(rampUpTime / concurrentUsers);
      }
    }
    
    // Wait for all users to complete
    await Promise.all(userPromises);
    
    const endTime = Date.now();
    const duration = endTime - startTime;
    
    // Calculate statistics
    const totalRequests = concurrentUsers * requestsPerUser;
    const averageResponseTime = responseTimes.length > 0 
      ? responseTimes.reduce((sum, time) => sum + time, 0) / responseTimes.length 
      : 0;
    
    const sortedTimes = responseTimes.sort((a, b) => a - b);
    const percentiles = {
      p50: this.getPercentile(sortedTimes, 50),
      p90: this.getPercentile(sortedTimes, 90),
      p95: this.getPercentile(sortedTimes, 95),
      p99: this.getPercentile(sortedTimes, 99)
    };
    
    const result: LoadTestResult = {
      testName: `Concurrent API Test (${concurrentUsers} users)`,
      duration,
      totalRequests,
      successfulRequests: successCount,
      failedRequests: failCount,
      averageResponseTime: Math.round(averageResponseTime),
      minResponseTime: sortedTimes.length > 0 ? sortedTimes[0] : 0,
      maxResponseTime: sortedTimes.length > 0 ? sortedTimes[sortedTimes.length - 1] : 0,
      requestsPerSecond: Math.round((successCount * 1000) / duration),
      errors: Object.entries(errors).map(([error, count]) => ({ error, count })),
      responseTimePercentiles: percentiles
    };
    
    console.log('Load test completed:', {
      duration: `${duration}ms`,
      totalRequests,
      successRate: `${Math.round((successCount / totalRequests) * 100)}%`,
      avgResponseTime: `${result.averageResponseTime}ms`,
      rps: result.requestsPerSecond
    });
    
    return result;
  }

  /**
   * Run a single user session
   */
  private async runUserSession(
    userId: number,
    requestCount: number,
    endpoint: string,
    method: string,
    payload?: any,
    headers?: Record<string, string>,
    rampUpTime?: number
  ): Promise<{
    successes: number;
    failures: number;
    responseTimes: number[];
    errors: { [key: string]: number };
  }> {
    const responseTimes: number[] = [];
    const errors: { [key: string]: number } = {};
    let successes = 0;
    let failures = 0;
    
    for (let i = 0; i < requestCount; i++) {
      const requestStart = Date.now();
      
      try {
        let response;
        
        if (this.request) {
          // Use API request context
          response = await this.request.fetch(endpoint, {
            method,
            data: payload,
            headers,
            timeout: 30000
          });
        } else {
          // Use page context
          response = await this.page.evaluate(async ({ endpoint, method, payload, headers }) => {
            const fetchResponse = await fetch(endpoint, {
              method,
              headers,
              body: payload ? JSON.stringify(payload) : undefined
            });
            
            return {
              ok: fetchResponse.ok,
              status: fetchResponse.status,
              statusText: fetchResponse.statusText
            };
          }, { endpoint, method, payload, headers });
        }
        
        const responseTime = Date.now() - requestStart;
        responseTimes.push(responseTime);
        
        if (response.ok || (response as any).ok) {
          successes++;
        } else {
          failures++;
          const errorKey = `HTTP_${(response as any).status || response.status()}`;
          errors[errorKey] = (errors[errorKey] || 0) + 1;
        }
        
      } catch (error) {
        const responseTime = Date.now() - requestStart;
        responseTimes.push(responseTime);
        failures++;
        
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        errors[errorMessage] = (errors[errorMessage] || 0) + 1;
      }
      
      // Small delay between requests from same user
      if (i < requestCount - 1) {
        await this.sleep(100);
      }
    }
    
    return { successes, failures, responseTimes, errors };
  }

  /**
   * Test WebSocket load handling
   */
  async testWebSocketLoad(options: {
    concurrentConnections: number;
    messagesPerConnection: number;
    messageInterval?: number;
  }): Promise<LoadTestResult> {
    const { concurrentConnections, messagesPerConnection, messageInterval = 1000 } = options;
    const startTime = Date.now();
    
    console.log(`Starting WebSocket load test: ${concurrentConnections} connections, ${messagesPerConnection} messages each`);
    
    const connectionPromises = [];
    let totalSuccessful = 0;
    let totalFailed = 0;
    const responseTimes: number[] = [];
    const errors: { [key: string]: number } = {};
    
    for (let i = 0; i < concurrentConnections; i++) {
      const connectionPromise = this.runWebSocketSession(i, messagesPerConnection, messageInterval)
        .then((result) => {
          totalSuccessful += result.successful;
          totalFailed += result.failed;
          responseTimes.push(...result.responseTimes);
          
          for (const [error, count] of Object.entries(result.errors)) {
            errors[error] = (errors[error] || 0) + count;
          }
        });
      
      connectionPromises.push(connectionPromise);
      
      // Small delay between connection attempts
      await this.sleep(50);
    }
    
    await Promise.all(connectionPromises);
    
    const endTime = Date.now();
    const duration = endTime - startTime;
    const totalMessages = concurrentConnections * messagesPerConnection;
    
    const sortedTimes = responseTimes.sort((a, b) => a - b);
    const averageResponseTime = responseTimes.length > 0 
      ? responseTimes.reduce((sum, time) => sum + time, 0) / responseTimes.length 
      : 0;
    
    return {
      testName: `WebSocket Load Test (${concurrentConnections} connections)`,
      duration,
      totalRequests: totalMessages,
      successfulRequests: totalSuccessful,
      failedRequests: totalFailed,
      averageResponseTime: Math.round(averageResponseTime),
      minResponseTime: sortedTimes.length > 0 ? sortedTimes[0] : 0,
      maxResponseTime: sortedTimes.length > 0 ? sortedTimes[sortedTimes.length - 1] : 0,
      requestsPerSecond: Math.round((totalSuccessful * 1000) / duration),
      errors: Object.entries(errors).map(([error, count]) => ({ error, count })),
      responseTimePercentiles: {
        p50: this.getPercentile(sortedTimes, 50),
        p90: this.getPercentile(sortedTimes, 90),
        p95: this.getPercentile(sortedTimes, 95),
        p99: this.getPercentile(sortedTimes, 99)
      }
    };
  }

  /**
   * Run a single WebSocket session
   */
  private async runWebSocketSession(
    connectionId: number,
    messageCount: number,
    messageInterval: number
  ): Promise<{
    successful: number;
    failed: number;
    responseTimes: number[];
    errors: { [key: string]: number };
  }> {
    return await this.page.evaluate(async ({ connectionId, messageCount, messageInterval }) => {
      return new Promise((resolve) => {
        const responseTimes: number[] = [];
        const errors: { [key: string]: number } = {};
        let successful = 0;
        let failed = 0;
        let messagesSent = 0;
        let messagesReceived = 0;
        const pendingMessages: { [key: string]: number } = {};
        
        try {
          const wsUrl = location.origin.replace('http', 'ws') + '/ws';
          const ws = new WebSocket(wsUrl);
          
          const timeout = setTimeout(() => {
            ws.close();
            errors['Connection timeout'] = 1;
            resolve({ successful, failed: failed + 1, responseTimes, errors });
          }, 30000);
          
          ws.onopen = () => {
            const sendMessage = () => {
              if (messagesSent >= messageCount) {
                // Wait for remaining responses
                setTimeout(() => {
                  clearTimeout(timeout);
                  ws.close();
                  resolve({ successful, failed, responseTimes, errors });
                }, 5000);
                return;
              }
              
              const messageId = `msg-${connectionId}-${messagesSent}`;
              const startTime = Date.now();
              pendingMessages[messageId] = startTime;
              
              ws.send(JSON.stringify({
                id: messageId,
                type: 'test',
                data: `Test message ${messagesSent} from connection ${connectionId}`
              }));
              
              messagesSent++;
              
              if (messagesSent < messageCount) {
                setTimeout(sendMessage, messageInterval);
              }
            };
            
            sendMessage();
          };
          
          ws.onmessage = (event) => {
            messagesReceived++;
            
            try {
              const data = JSON.parse(event.data);
              const messageId = data.id;
              
              if (messageId && pendingMessages[messageId]) {
                const responseTime = Date.now() - pendingMessages[messageId];
                responseTimes.push(responseTime);
                delete pendingMessages[messageId];
                successful++;
              } else {
                successful++; // Generic response
                responseTimes.push(50); // Estimated response time
              }
            } catch (e) {
              successful++; // Non-JSON response is still a response
              responseTimes.push(50);
            }
          };
          
          ws.onerror = (error) => {
            errors['WebSocket error'] = (errors['WebSocket error'] || 0) + 1;
            failed++;
          };
          
          ws.onclose = () => {
            // Handle any remaining pending messages as failures
            failed += Object.keys(pendingMessages).length;
          };
          
        } catch (error) {
          errors[error.message] = 1;
          resolve({ successful: 0, failed: 1, responseTimes: [], errors });
        }
      });
    }, { connectionId, messageCount, messageInterval });
  }

  /**
   * Monitor resource usage during load test
   */
  async monitorResourceUsage(duration: number = 30000): Promise<Array<{
    timestamp: number;
    memory: any;
    cpu?: number;
    networkActivity: boolean;
  }>> {
    const samples = [];
    const interval = 2000; // Sample every 2 seconds
    const endTime = Date.now() + duration;
    
    while (Date.now() < endTime) {
      const sample = await this.page.evaluate(() => {
        const memory = (performance as any).memory;
        const timing = performance.timing;
        const now = Date.now();
        
        // Check for recent network activity
        const resources = performance.getEntriesByType('resource');
        const recentResources = resources.filter((r: any) => 
          r.responseEnd > now - 5000 // Resources loaded in last 5 seconds
        );
        
        return {
          timestamp: now,
          memory: memory ? {
            totalJSHeapSize: memory.totalJSHeapSize,
            usedJSHeapSize: memory.usedJSHeapSize,
            jsHeapSizeLimit: memory.jsHeapSizeLimit
          } : null,
          networkActivity: recentResources.length > 0
        };
      });
      
      samples.push(sample);
      await this.sleep(interval);
    }
    
    return samples;
  }

  /**
   * Test different payload sizes for API performance
   */
  async testPayloadSizes(endpoint: string, payloadSizes: number[]): Promise<LoadTestResult[]> {
    const results: LoadTestResult[] = [];
    
    for (const size of payloadSizes) {
      const payload = {
        data: 'x'.repeat(size),
        size: size,
        timestamp: Date.now()
      };
      
      const result = await this.runConcurrentAPITest({
        concurrentUsers: 5,
        requestsPerUser: 10,
        endpoint,
        method: 'POST',
        payload
      });
      
      result.testName = `Payload Size Test (${size} bytes)`;
      results.push(result);
      
      console.log(`Payload size ${size} bytes: ${result.averageResponseTime}ms avg response time`);
    }
    
    return results;
  }

  /**
   * Save load test results
   */
  async saveResults(testName: string, results: LoadTestResult | LoadTestResult[]): Promise<void> {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const filename = `load-test-${testName}-${timestamp}.json`;
    const filepath = path.join('reports', 'load-tests', filename);
    
    await fs.mkdir(path.dirname(filepath), { recursive: true }).catch(() => {});
    
    const data = {
      testName,
      timestamp: new Date().toISOString(),
      userAgent: await this.page.evaluate(() => navigator.userAgent),
      results: Array.isArray(results) ? results : [results]
    };
    
    await fs.writeFile(filepath, JSON.stringify(data, null, 2));
    console.log(`Load test results saved to: ${filepath}`);
  }

  /**
   * Utility functions
   */
  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
  
  private getPercentile(sortedArray: number[], percentile: number): number {
    if (sortedArray.length === 0) return 0;
    const index = Math.ceil((percentile / 100) * sortedArray.length) - 1;
    return sortedArray[Math.max(0, index)];
  }
}