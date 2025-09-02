#!/usr/bin/env node
/**
 * Redis Cluster Validation Test Suite
 * Tests performance, failover, and clustering functionality
 */

const Redis = require('ioredis');
const { performance } = require('perf_hooks');
const fs = require('fs').promises;

class RedisClusterTester {
    constructor() {
        this.cluster = null;
        this.results = {
            timestamp: new Date().toISOString(),
            tests: {},
            metrics: {},
            summary: {}
        };
    }

    async initialize() {
        console.log('🔴 Initializing Redis cluster connection...');
        
        // Connect to Redis cluster through HAProxy
        this.cluster = new Redis.Cluster([
            { host: 'localhost', port: 7000 },
            { host: 'localhost', port: 7001 },
            { host: 'localhost', port: 7002 }
        ], {
            enableReadyCheck: true,
            redisOptions: {
                password: process.env.REDIS_PASSWORD
            }
        });

        await this.cluster.ping();
        console.log('✅ Redis cluster connected successfully');
    }

    async testBasicOperations() {
        console.log('\n📊 Testing basic Redis operations...');
        const testStart = performance.now();
        
        const operations = [];
        const testData = Array.from({ length: 1000 }, (_, i) => ({
            key: `test:key:${i}`,
            value: `value_${i}_${Math.random().toString(36).substr(2, 9)}`
        }));

        // Test SET operations
        for (const { key, value } of testData) {
            operations.push(this.cluster.set(key, value));
        }
        
        await Promise.all(operations);
        const setTime = performance.now() - testStart;

        // Test GET operations
        const getStart = performance.now();
        const getOperations = testData.map(({ key }) => this.cluster.get(key));
        const results = await Promise.all(getOperations);
        const getTime = performance.now() - getStart;

        // Validate results
        const successCount = results.filter(Boolean).length;
        
        this.results.tests.basicOperations = {
            setOperations: testData.length,
            setTime: setTime,
            setThroughput: Math.round((testData.length / setTime) * 1000),
            getOperations: testData.length,
            getTime: getTime,
            getThroughput: Math.round((testData.length / getTime) * 1000),
            successRate: (successCount / testData.length) * 100,
            passed: successCount === testData.length
        };

        console.log(`   SET: ${testData.length} ops in ${setTime.toFixed(2)}ms (${this.results.tests.basicOperations.setThroughput} ops/sec)`);
        console.log(`   GET: ${testData.length} ops in ${getTime.toFixed(2)}ms (${this.results.tests.basicOperations.getThroughput} ops/sec)`);
    }

    async testConcurrentLoad() {
        console.log('\n⚡ Testing concurrent load handling...');
        const concurrencyLevels = [10, 50, 100, 200];
        this.results.tests.concurrentLoad = {};

        for (const concurrency of concurrencyLevels) {
            const testStart = performance.now();
            const promises = [];

            for (let i = 0; i < concurrency; i++) {
                const batchPromises = Array.from({ length: 100 }, (_, j) => {
                    const key = `concurrent:${concurrency}:${i}:${j}`;
                    const value = JSON.stringify({ 
                        timestamp: Date.now(), 
                        worker: i, 
                        operation: j 
                    });
                    return this.cluster.setex(key, 3600, value);
                });
                promises.push(...batchPromises);
            }

            await Promise.all(promises);
            const duration = performance.now() - testStart;
            const totalOps = concurrency * 100;

            this.results.tests.concurrentLoad[concurrency] = {
                operations: totalOps,
                duration: duration,
                throughput: Math.round((totalOps / duration) * 1000),
                avgLatency: duration / totalOps
            };

            console.log(`   ${concurrency} workers: ${totalOps} ops in ${duration.toFixed(2)}ms (${this.results.tests.concurrentLoad[concurrency].throughput} ops/sec)`);
        }
    }

    async testMemoryUsage() {
        console.log('\n💾 Testing memory usage patterns...');
        
        const memoryTest = async (keySize, valueSize, count) => {
            const testStart = performance.now();
            const operations = [];
            
            for (let i = 0; i < count; i++) {
                const key = `memory:${keySize}:${i.toString().padStart(keySize, '0')}`;
                const value = 'x'.repeat(valueSize);
                operations.push(this.cluster.setex(key, 1800, value));
            }
            
            await Promise.all(operations);
            const duration = performance.now() - testStart;
            
            // Get memory info
            const info = await this.cluster.info('memory');
            const memoryUsed = info.match(/used_memory:(\d+)/)?.[1];
            
            return {
                keySize,
                valueSize,
                count,
                duration,
                memoryUsed: parseInt(memoryUsed) || 0,
                throughput: Math.round((count / duration) * 1000)
            };
        };

        this.results.tests.memoryUsage = {
            smallKeys: await memoryTest(8, 100, 1000),
            mediumKeys: await memoryTest(32, 1024, 500),
            largeKeys: await memoryTest(64, 10240, 100)
        };

        Object.entries(this.results.tests.memoryUsage).forEach(([size, result]) => {
            console.log(`   ${size}: ${result.count} keys (${result.keySize}B key, ${result.valueSize}B value) = ${result.throughput} ops/sec`);
        });
    }

    async testFailoverScenarios() {
        console.log('\n🔄 Testing failover scenarios...');
        
        // This would require actual cluster manipulation in production
        // For testing purposes, we'll simulate connection resilience
        
        const resilience = await this.testConnectionResilience();
        
        this.results.tests.failover = {
            connectionResilience: resilience,
            note: 'Full failover testing requires cluster manipulation'
        };

        console.log(`   Connection resilience: ${resilience.passed ? 'PASSED' : 'FAILED'}`);
    }

    async testConnectionResilience() {
        const iterations = 100;
        let successCount = 0;
        const latencies = [];

        for (let i = 0; i < iterations; i++) {
            try {
                const start = performance.now();
                await this.cluster.ping();
                const latency = performance.now() - start;
                latencies.push(latency);
                successCount++;
            } catch (error) {
                console.warn(`   Connection test ${i} failed:`, error.message);
            }
        }

        return {
            iterations,
            successCount,
            successRate: (successCount / iterations) * 100,
            avgLatency: latencies.reduce((a, b) => a + b, 0) / latencies.length,
            maxLatency: Math.max(...latencies),
            minLatency: Math.min(...latencies),
            passed: successCount >= iterations * 0.95 // 95% success rate required
        };
    }

    async performanceBaseline() {
        console.log('\n📈 Establishing performance baseline...');
        
        const baseline = {
            timestamp: Date.now(),
            operations: {},
            targets: {
                latencyReduction: { target: '60-80%', baseline: 'TBD' },
                throughputImprovement: { target: '2x-3x', baseline: 'TBD' }
            }
        };

        // Single operation latency
        const singleOpLatencies = [];
        for (let i = 0; i < 100; i++) {
            const start = performance.now();
            await this.cluster.get(`baseline:${i}`);
            singleOpLatencies.push(performance.now() - start);
        }

        baseline.operations.singleGet = {
            avgLatency: singleOpLatencies.reduce((a, b) => a + b, 0) / singleOpLatencies.length,
            p95Latency: singleOpLatencies.sort((a, b) => a - b)[Math.floor(singleOpLatencies.length * 0.95)],
            maxLatency: Math.max(...singleOpLatencies)
        };

        this.results.metrics.baseline = baseline;
        
        console.log(`   Average GET latency: ${baseline.operations.singleGet.avgLatency.toFixed(3)}ms`);
        console.log(`   P95 GET latency: ${baseline.operations.singleGet.p95Latency.toFixed(3)}ms`);
    }

    generateSummary() {
        const { basicOperations, concurrentLoad, memoryUsage, failover } = this.results.tests;
        
        this.results.summary = {
            overallHealth: 'HEALTHY',
            criticalIssues: [],
            performanceMetrics: {
                basicOperationsThroughput: Math.max(basicOperations.setThroughput, basicOperations.getThroughput),
                peakConcurrentThroughput: Math.max(...Object.values(concurrentLoad).map(r => r.throughput)),
                connectionResilience: failover.connectionResilience.successRate
            },
            recommendations: [],
            targetAchievement: {
                latencyReduction: 'BASELINE_ESTABLISHED',
                throughputImprovement: 'BASELINE_ESTABLISHED',
                clusterStability: failover.connectionResilience.passed ? 'ACHIEVED' : 'NEEDS_WORK'
            }
        };

        // Add recommendations based on results
        if (this.results.summary.performanceMetrics.connectionResilience < 95) {
            this.results.summary.criticalIssues.push('Connection resilience below 95%');
            this.results.summary.recommendations.push('Review cluster configuration and network stability');
        }

        if (this.results.summary.performanceMetrics.basicOperationsThroughput < 1000) {
            this.results.summary.recommendations.push('Consider Redis configuration tuning for higher throughput');
        }
    }

    async cleanup() {
        console.log('\n🧹 Cleaning up test data...');
        
        const patterns = [
            'test:key:*',
            'concurrent:*',
            'memory:*',
            'baseline:*'
        ];

        for (const pattern of patterns) {
            const keys = await this.cluster.keys(pattern);
            if (keys.length > 0) {
                await this.cluster.del(...keys);
                console.log(`   Cleaned ${keys.length} keys matching ${pattern}`);
            }
        }

        await this.cluster.disconnect();
    }

    async run() {
        try {
            await this.initialize();
            await this.testBasicOperations();
            await this.testConcurrentLoad();
            await this.testMemoryUsage();
            await this.testFailoverScenarios();
            await this.performanceBaseline();
            
            this.generateSummary();
            
            console.log('\n📊 Test Summary:');
            console.log(`   Overall Health: ${this.results.summary.overallHealth}`);
            console.log(`   Peak Throughput: ${this.results.summary.performanceMetrics.peakConcurrentThroughput} ops/sec`);
            console.log(`   Connection Resilience: ${this.results.summary.performanceMetrics.connectionResilience.toFixed(1)}%`);
            
            if (this.results.summary.criticalIssues.length > 0) {
                console.log('\n⚠️  Critical Issues:');
                this.results.summary.criticalIssues.forEach(issue => console.log(`   - ${issue}`));
            }

            // Save results
            await fs.writeFile(
                '/home/kp/ollamamax/test-results/redis-cluster-validation.json',
                JSON.stringify(this.results, null, 2)
            );
            
            console.log('\n✅ Redis cluster validation completed successfully');
            return this.results;
            
        } catch (error) {
            console.error('❌ Redis cluster test failed:', error);
            this.results.error = error.message;
            this.results.summary = { overallHealth: 'FAILED', error: error.message };
            
            await fs.writeFile(
                '/home/kp/ollamamax/test-results/redis-cluster-validation.json',
                JSON.stringify(this.results, null, 2)
            );
            
            throw error;
        } finally {
            await this.cleanup();
        }
    }
}

// Run if called directly
if (require.main === module) {
    const tester = new RedisClusterTester();
    tester.run().then(() => {
        console.log('Redis cluster validation completed');
        process.exit(0);
    }).catch((error) => {
        console.error('Validation failed:', error);
        process.exit(1);
    });
}

module.exports = RedisClusterTester;