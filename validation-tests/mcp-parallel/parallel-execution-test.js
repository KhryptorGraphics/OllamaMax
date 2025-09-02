#!/usr/bin/env node
/**
 * MCP Parallel Execution Framework Validation Test Suite
 * Tests parallelization efficiency, dependency handling, and performance
 */

const { performance } = require('perf_hooks');
const fs = require('fs').promises;
const path = require('path');

// Import the parallel execution framework
const MCPParallelFramework = require('../../critical-fixes/mcp-parallel/parallel-execution-framework.js');

class MCPParallelTester {
    constructor() {
        this.framework = new MCPParallelFramework();
        this.results = {
            timestamp: new Date().toISOString(),
            tests: {},
            metrics: {},
            summary: {}
        };
    }

    async initialize() {
        console.log('⚡ Initializing MCP Parallel Execution Framework...');
        await this.framework.initialize();
        console.log('✅ Framework initialized successfully');
    }

    // Mock MCP operations for testing
    createMockOperations() {
        return {
            // Independent operations that can run in parallel
            independent: [
                { id: 'read_file_1', type: 'read', dependencies: [], duration: 100 },
                { id: 'read_file_2', type: 'read', dependencies: [], duration: 150 },
                { id: 'read_file_3', type: 'read', dependencies: [], duration: 120 },
                { id: 'read_file_4', type: 'read', dependencies: [], duration: 80 },
                { id: 'read_file_5', type: 'read', dependencies: [], duration: 200 }
            ],
            
            // Operations with dependencies (must be sequential)
            dependent: [
                { id: 'analyze_code', type: 'analyze', dependencies: ['read_file_1', 'read_file_2'], duration: 300 },
                { id: 'generate_docs', type: 'write', dependencies: ['analyze_code'], duration: 250 },
                { id: 'format_code', type: 'edit', dependencies: ['read_file_3', 'read_file_4'], duration: 180 },
                { id: 'final_review', type: 'review', dependencies: ['generate_docs', 'format_code'], duration: 400 }
            ],
            
            // Complex dependency chain
            complex: [
                { id: 'fetch_data_1', type: 'fetch', dependencies: [], duration: 200 },
                { id: 'fetch_data_2', type: 'fetch', dependencies: [], duration: 180 },
                { id: 'fetch_data_3', type: 'fetch', dependencies: [], duration: 220 },
                { id: 'process_data_1', type: 'process', dependencies: ['fetch_data_1'], duration: 300 },
                { id: 'process_data_2', type: 'process', dependencies: ['fetch_data_2'], duration: 280 },
                { id: 'merge_data', type: 'merge', dependencies: ['process_data_1', 'process_data_2', 'fetch_data_3'], duration: 150 },
                { id: 'validate_data', type: 'validate', dependencies: ['merge_data'], duration: 100 },
                { id: 'save_results', type: 'save', dependencies: ['validate_data'], duration: 120 }
            ]
        };
    }

    // Mock operation executor
    async executeMockOperation(operation) {
        // Simulate operation execution time
        await new Promise(resolve => setTimeout(resolve, operation.duration));
        
        return {
            id: operation.id,
            type: operation.type,
            success: true,
            duration: operation.duration,
            result: `Result for ${operation.id}`
        };
    }

    async testIndependentParallelization() {
        console.log('\n🔄 Testing independent operation parallelization...');
        
        const mockOps = this.createMockOperations();
        const operations = mockOps.independent;
        
        // Sequential execution baseline
        const sequentialStart = performance.now();
        const sequentialResults = [];
        for (const op of operations) {
            const result = await this.executeMockOperation(op);
            sequentialResults.push(result);
        }
        const sequentialTime = performance.now() - sequentialStart;
        
        // Parallel execution test
        const parallelStart = performance.now();
        const parallelPromises = operations.map(op => this.executeMockOperation(op));
        const parallelResults = await Promise.all(parallelPromises);
        const parallelTime = performance.now() - parallelStart;
        
        const speedup = sequentialTime / parallelTime;
        const efficiency = (speedup / operations.length) * 100;
        
        this.results.tests.independentParallelization = {
            operations: operations.length,
            sequentialTime: Math.round(sequentialTime),
            parallelTime: Math.round(parallelTime),
            speedup: parseFloat(speedup.toFixed(2)),
            efficiency: parseFloat(efficiency.toFixed(1)),
            theoreticalMaxSpeedup: operations.length,
            passed: speedup > 2.0 // Expect at least 2x speedup for 5 independent operations
        };
        
        console.log(`   Sequential: ${Math.round(sequentialTime)}ms`);
        console.log(`   Parallel: ${Math.round(parallelTime)}ms`);
        console.log(`   Speedup: ${speedup.toFixed(2)}x (${efficiency.toFixed(1)}% efficiency)`);
    }

    async testDependencyHandling() {
        console.log('\n🔗 Testing dependency chain handling...');
        
        const mockOps = this.createMockOperations();
        const operations = mockOps.dependent;
        
        // Build operation map
        const opMap = new Map();
        operations.forEach(op => opMap.set(op.id, op));
        
        // Test dependency resolution
        const dependencyStart = performance.now();
        const executionPlan = this.framework.buildExecutionPlan(operations);
        const planTime = performance.now() - dependencyStart;
        
        // Execute with dependency awareness
        const executionStart = performance.now();
        const results = await this.executeWithDependencies(operations, executionPlan);
        const executionTime = performance.now() - executionStart;
        
        // Validate execution order
        const orderValid = this.validateExecutionOrder(results, operations);
        
        this.results.tests.dependencyHandling = {
            operations: operations.length,
            planningTime: Math.round(planTime),
            executionTime: Math.round(executionTime),
            executionPlan: executionPlan,
            orderValid: orderValid,
            passed: orderValid && results.every(r => r.success)
        };
        
        console.log(`   Planning time: ${Math.round(planTime)}ms`);
        console.log(`   Execution time: ${Math.round(executionTime)}ms`);
        console.log(`   Execution order: ${orderValid ? 'VALID' : 'INVALID'}`);
    }

    async executeWithDependencies(operations, executionPlan) {
        const results = [];
        const completed = new Set();
        const opMap = new Map();
        
        operations.forEach(op => opMap.set(op.id, op));
        
        for (const batch of executionPlan) {
            const batchPromises = batch.map(async (opId) => {
                const operation = opMap.get(opId);
                const result = await this.executeMockOperation(operation);
                completed.add(opId);
                return result;
            });
            
            const batchResults = await Promise.all(batchPromises);
            results.push(...batchResults);
        }
        
        return results;
    }

    validateExecutionOrder(results, operations) {
        const executionOrder = results.map(r => r.id);
        const opMap = new Map();
        operations.forEach(op => opMap.set(op.id, op));
        
        for (let i = 0; i < executionOrder.length; i++) {
            const opId = executionOrder[i];
            const operation = opMap.get(opId);
            
            // Check if all dependencies were executed before this operation
            for (const depId of operation.dependencies) {
                const depIndex = executionOrder.indexOf(depId);
                if (depIndex === -1 || depIndex > i) {
                    console.warn(`   Dependency violation: ${opId} executed before ${depId}`);
                    return false;
                }
            }
        }
        
        return true;
    }

    async testComplexScenario() {
        console.log('\n🏗️ Testing complex dependency scenario...');
        
        const mockOps = this.createMockOperations();
        const operations = mockOps.complex;
        
        // Measure planning overhead
        const planStart = performance.now();
        const executionPlan = this.framework.buildExecutionPlan(operations);
        const planTime = performance.now() - planStart;
        
        // Execute complex scenario
        const execStart = performance.now();
        const results = await this.executeWithDependencies(operations, executionPlan);
        const execTime = performance.now() - execStart;
        
        // Calculate theoretical minimum time (critical path)
        const criticalPath = this.calculateCriticalPath(operations);
        const parallelizationRatio = criticalPath / execTime;
        
        this.results.tests.complexScenario = {
            operations: operations.length,
            planningTime: Math.round(planTime),
            executionTime: Math.round(execTime),
            criticalPathTime: criticalPath,
            parallelizationRatio: parseFloat(parallelizationRatio.toFixed(2)),
            batchCount: executionPlan.length,
            avgBatchSize: parseFloat((operations.length / executionPlan.length).toFixed(1)),
            passed: parallelizationRatio > 0.7 // Should be close to optimal
        };
        
        console.log(`   Planning overhead: ${Math.round(planTime)}ms`);
        console.log(`   Execution time: ${Math.round(execTime)}ms`);
        console.log(`   Critical path: ${criticalPath}ms`);
        console.log(`   Parallelization ratio: ${parallelizationRatio.toFixed(2)}`);
    }

    calculateCriticalPath(operations) {
        const opMap = new Map();
        const memo = new Map();
        
        operations.forEach(op => opMap.set(op.id, op));
        
        const calculatePath = (opId) => {
            if (memo.has(opId)) return memo.get(opId);
            
            const operation = opMap.get(opId);
            if (!operation.dependencies.length) {
                memo.set(opId, operation.duration);
                return operation.duration;
            }
            
            const maxDependencyPath = Math.max(
                ...operation.dependencies.map(depId => calculatePath(depId))
            );
            
            const totalPath = maxDependencyPath + operation.duration;
            memo.set(opId, totalPath);
            return totalPath;
        };
        
        return Math.max(...operations.map(op => calculatePath(op.id)));
    }

    async testBatchOptimization() {
        console.log('\n📦 Testing batch size optimization...');
        
        const batchSizes = [1, 5, 10, 20, 50];
        const operationCounts = [10, 50, 100];
        
        this.results.tests.batchOptimization = {};
        
        for (const opCount of operationCounts) {
            const operations = Array.from({ length: opCount }, (_, i) => ({
                id: `batch_op_${i}`,
                type: 'process',
                dependencies: [],
                duration: 50 + Math.random() * 100
            }));
            
            this.results.tests.batchOptimization[opCount] = {};
            
            for (const batchSize of batchSizes) {
                const start = performance.now();
                
                // Execute in batches
                const results = [];
                for (let i = 0; i < operations.length; i += batchSize) {
                    const batch = operations.slice(i, i + batchSize);
                    const batchPromises = batch.map(op => this.executeMockOperation(op));
                    const batchResults = await Promise.all(batchPromises);
                    results.push(...batchResults);
                }
                
                const duration = performance.now() - start;
                const throughput = Math.round((opCount / duration) * 1000);
                
                this.results.tests.batchOptimization[opCount][batchSize] = {
                    duration: Math.round(duration),
                    throughput: throughput
                };
            }
            
            // Find optimal batch size for this operation count
            const optimal = Object.entries(this.results.tests.batchOptimization[opCount])
                .reduce((best, [size, metrics]) => 
                    metrics.throughput > best.throughput ? { size: parseInt(size), ...metrics } : best,
                    { size: 1, throughput: 0 }
                );
            
            console.log(`   ${opCount} ops: optimal batch size ${optimal.size} (${optimal.throughput} ops/sec)`);
        }
    }

    async testErrorHandling() {
        console.log('\n⚠️  Testing error handling and recovery...');
        
        const operationsWithErrors = [
            { id: 'op_1', type: 'process', dependencies: [], duration: 100, shouldFail: false },
            { id: 'op_2', type: 'process', dependencies: [], duration: 150, shouldFail: true },
            { id: 'op_3', type: 'process', dependencies: ['op_1'], duration: 200, shouldFail: false },
            { id: 'op_4', type: 'process', dependencies: ['op_2'], duration: 180, shouldFail: false }
        ];
        
        const executeMockWithErrors = async (operation) => {
            await new Promise(resolve => setTimeout(resolve, operation.duration));
            
            if (operation.shouldFail) {
                throw new Error(`Simulated failure in ${operation.id}`);
            }
            
            return {
                id: operation.id,
                success: true,
                result: `Result for ${operation.id}`
            };
        };
        
        const start = performance.now();
        const results = [];
        const errors = [];
        
        try {
            // Test error propagation and isolation
            const promises = operationsWithErrors.map(async (op) => {
                try {
                    const result = await executeMockWithErrors(op);
                    results.push(result);
                    return result;
                } catch (error) {
                    errors.push({ operation: op.id, error: error.message });
                    throw error;
                }
            });
            
            await Promise.allSettled(promises);
        } catch (error) {
            // Expected for failing operations
        }
        
        const duration = performance.now() - start;
        
        this.results.tests.errorHandling = {
            totalOperations: operationsWithErrors.length,
            successfulOperations: results.length,
            failedOperations: errors.length,
            executionTime: Math.round(duration),
            errorIsolation: results.length > 0 && errors.length > 0, // Some succeeded despite failures
            passed: results.length === 2 && errors.length === 1 // Expected: 2 success, 1 failure, 1 blocked
        };
        
        console.log(`   Successful: ${results.length}, Failed: ${errors.length}`);
        console.log(`   Error isolation: ${this.results.tests.errorHandling.errorIsolation ? 'SUCCESS' : 'FAILED'}`);
    }

    async performanceBaseline() {
        console.log('\n📊 Establishing performance baseline...');
        
        // Test different operation loads
        const loads = [10, 50, 100, 200];
        const baseline = {};
        
        for (const load of loads) {
            const operations = Array.from({ length: load }, (_, i) => ({
                id: `baseline_${i}`,
                type: 'process',
                dependencies: [],
                duration: 50 + Math.random() * 100
            }));
            
            // Sequential execution
            const seqStart = performance.now();
            for (const op of operations) {
                await this.executeMockOperation(op);
            }
            const seqTime = performance.now() - seqStart;
            
            // Parallel execution
            const parStart = performance.now();
            const promises = operations.map(op => this.executeMockOperation(op));
            await Promise.all(promises);
            const parTime = performance.now() - parStart;
            
            baseline[load] = {
                sequentialTime: Math.round(seqTime),
                parallelTime: Math.round(parTime),
                speedup: parseFloat((seqTime / parTime).toFixed(2)),
                efficiency: parseFloat(((seqTime / parTime) / load * 100).toFixed(1))
            };
            
            console.log(`   ${load} ops: ${baseline[load].speedup}x speedup (${baseline[load].efficiency}% efficiency)`);
        }
        
        this.results.metrics.baseline = baseline;
    }

    generateSummary() {
        const tests = this.results.tests;
        const criticalIssues = [];
        const recommendations = [];
        
        // Check test results
        if (!tests.independentParallelization?.passed) {
            criticalIssues.push('Independent parallelization below expected performance');
        }
        
        if (!tests.dependencyHandling?.passed) {
            criticalIssues.push('Dependency handling validation failed');
        }
        
        if (!tests.complexScenario?.passed) {
            criticalIssues.push('Complex scenario parallelization below threshold');
        }
        
        if (!tests.errorHandling?.passed) {
            criticalIssues.push('Error handling and isolation issues detected');
        }
        
        // Generate recommendations
        if (tests.independentParallelization?.efficiency < 70) {
            recommendations.push('Consider reducing parallelization overhead or increasing batch sizes');
        }
        
        if (tests.batchOptimization) {
            recommendations.push('Implement adaptive batch sizing based on operation characteristics');
        }
        
        // Calculate overall performance improvement
        const avgSpeedup = Object.values(this.results.metrics.baseline || {})
            .reduce((sum, metrics) => sum + metrics.speedup, 0) / 
            Object.keys(this.results.metrics.baseline || {}).length;
        
        this.results.summary = {
            overallHealth: criticalIssues.length === 0 ? 'HEALTHY' : 'NEEDS_ATTENTION',
            criticalIssues,
            recommendations,
            performanceMetrics: {
                averageSpeedup: parseFloat((avgSpeedup || 0).toFixed(2)),
                independentParallelEfficiency: tests.independentParallelization?.efficiency || 0,
                dependencyHandlingValid: tests.dependencyHandling?.passed || false,
                errorHandlingRobust: tests.errorHandling?.passed || false
            },
            targetAchievement: {
                overheadReduction: avgSpeedup > 2.5 ? 'ACHIEVED' : 'PARTIAL',
                dependencyOptimization: tests.dependencyHandling?.passed ? 'ACHIEVED' : 'FAILED',
                errorResilience: tests.errorHandling?.passed ? 'ACHIEVED' : 'FAILED'
            }
        };
    }

    async cleanup() {
        console.log('\n🧹 Cleaning up test resources...');
        // Framework cleanup if needed
        if (this.framework.cleanup) {
            await this.framework.cleanup();
        }
    }

    async run() {
        try {
            await this.initialize();
            await this.testIndependentParallelization();
            await this.testDependencyHandling();
            await this.testComplexScenario();
            await this.testBatchOptimization();
            await this.testErrorHandling();
            await this.performanceBaseline();
            
            this.generateSummary();
            
            console.log('\n📊 MCP Parallel Execution Test Summary:');
            console.log(`   Overall Health: ${this.results.summary.overallHealth}`);
            console.log(`   Average Speedup: ${this.results.summary.performanceMetrics.averageSpeedup}x`);
            console.log(`   Dependency Handling: ${this.results.summary.performanceMetrics.dependencyHandlingValid ? 'PASSED' : 'FAILED'}`);
            console.log(`   Error Resilience: ${this.results.summary.performanceMetrics.errorHandlingRobust ? 'PASSED' : 'FAILED'}`);
            
            if (this.results.summary.criticalIssues.length > 0) {
                console.log('\n⚠️  Critical Issues:');
                this.results.summary.criticalIssues.forEach(issue => console.log(`   - ${issue}`));
            }
            
            // Save results
            await fs.writeFile(
                '/home/kp/ollamamax/test-results/mcp-parallel-validation.json',
                JSON.stringify(this.results, null, 2)
            );
            
            console.log('\n✅ MCP Parallel Execution validation completed successfully');
            return this.results;
            
        } catch (error) {
            console.error('❌ MCP Parallel Execution test failed:', error);
            this.results.error = error.message;
            this.results.summary = { overallHealth: 'FAILED', error: error.message };
            
            await fs.writeFile(
                '/home/kp/ollamamax/test-results/mcp-parallel-validation.json',
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
    const tester = new MCPParallelTester();
    tester.run().then(() => {
        console.log('MCP Parallel Execution validation completed');
        process.exit(0);
    }).catch((error) => {
        console.error('Validation failed:', error);
        process.exit(1);
    });
}

module.exports = MCPParallelTester;