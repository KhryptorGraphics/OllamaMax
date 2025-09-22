#!/usr/bin/env node

/**
 * OllamaMax Docker Swarm Integration Test Suite
 * 
 * This script tests the deployed Docker Swarm stack to ensure all services
 * are running correctly and can communicate with each other.
 */

import http from 'http';
import https from 'https';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

// Test configuration
const config = {
    baseUrl: process.env.BASE_URL || 'http://localhost:8080',
    apiPort: process.env.API_PORT || 8080,
    prometheusUrl: 'http://localhost:9090',
    grafanaUrl: 'http://localhost:3001',
    kibanaUrl: 'http://localhost:5601',
    portainerUrl: 'http://localhost:9000',
    testTimeout: 60000, // 60 seconds
    retryAttempts: 5,
    retryDelay: 5000 // 5 seconds
};

// Color codes for output
const colors = {
    reset: '\x1b[0m',
    red: '\x1b[31m',
    green: '\x1b[32m',
    yellow: '\x1b[33m',
    blue: '\x1b[34m'
};

// Test results storage
const testResults = {
    passed: [],
    failed: [],
    warnings: []
};

// Helper functions
function log(message, color = colors.reset) {
    console.log(`${color}${message}${colors.reset}`);
}

function logTest(name, status) {
    const symbol = status === 'passed' ? '✓' : status === 'failed' ? '✗' : '⚠';
    const color = status === 'passed' ? colors.green : status === 'failed' ? colors.red : colors.yellow;
    log(`  ${symbol} ${name}`, color);
}

async function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function executeCommand(command) {
    try {
        const { stdout, stderr } = await execAsync(command);
        return { success: true, stdout, stderr };
    } catch (error) {
        return { success: false, error: error.message };
    }
}

async function checkHttpEndpoint(url, options = {}) {
    return new Promise((resolve) => {
        const protocol = url.startsWith('https') ? https : http;
        
        // Allow self-signed certificates for testing
        if (protocol === https) {
            options.rejectUnauthorized = false;
        }
        
        const req = protocol.get(url, options, (res) => {
            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => {
                resolve({
                    success: res.statusCode >= 200 && res.statusCode < 400,
                    statusCode: res.statusCode,
                    data
                });
            });
        });
        
        req.on('error', (error) => {
            resolve({ success: false, error: error.message });
        });
        
        req.setTimeout(5000, () => {
            req.destroy();
            resolve({ success: false, error: 'Request timeout' });
        });
    });
}

async function retryOperation(operation, name, maxAttempts = config.retryAttempts) {
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
        const result = await operation();
        if (result.success) {
            return result;
        }
        
        if (attempt < maxAttempts) {
            log(`    Retry ${attempt}/${maxAttempts} for ${name}...`, colors.yellow);
            await sleep(config.retryDelay);
        }
    }
    return { success: false, error: `Failed after ${maxAttempts} attempts` };
}

// Test functions
async function testDockerSwarm() {
    log('\n1. Testing Docker Swarm Status', colors.blue);
    
    const swarmInfo = await executeCommand('docker info --format "{{.Swarm.LocalNodeState}}"');
    if (swarmInfo.success && swarmInfo.stdout.trim() === 'active') {
        logTest('Docker Swarm is active', 'passed');
        testResults.passed.push('Docker Swarm active');
    } else {
        logTest('Docker Swarm is not active', 'failed');
        testResults.failed.push('Docker Swarm not active');
        return false;
    }
    
    const nodeInfo = await executeCommand('docker node ls --format "table {{.ID}}\t{{.Status}}\t{{.Availability}}"');
    if (nodeInfo.success) {
        logTest('Docker Swarm nodes accessible', 'passed');
        testResults.passed.push('Swarm nodes accessible');
    } else {
        logTest('Cannot access Docker Swarm nodes', 'failed');
        testResults.failed.push('Cannot access Swarm nodes');
    }
    
    return true;
}

async function testStackDeployment() {
    log('\n2. Testing Stack Deployment', colors.blue);
    
    const stackList = await executeCommand('docker stack ls --format "{{.Name}}"');
    if (stackList.success && (stackList.stdout.includes('ollamamax') || stackList.stdout.includes('ollamamax-test'))) {
        logTest('OllamaMax stack deployed', 'passed');
        testResults.passed.push('Stack deployed');
    } else {
        logTest('OllamaMax stack not found', 'failed');
        testResults.failed.push('Stack not deployed');
        return false;
    }
    
    // Check for either ollamamax or ollamamax-test namespace
    const servicesCmd = await executeCommand('docker service ls --format "{{.Name}}\t{{.Replicas}}"');
    if (servicesCmd.success) {
        const serviceLines = servicesCmd.stdout.trim().split('\n').filter(line => 
            line.includes('ollamamax') && line.length > 0
        );
        
        if (serviceLines.length === 0) {
            logTest('No OllamaMax services found', 'warning');
            testResults.warnings.push('No services found');
            return false;
        }
        
        let allHealthy = true;
        
        for (const line of serviceLines) {
            const parts = line.split('\t');
            if (parts.length >= 2) {
                const [name, replicas] = parts;
                const replicaParts = replicas.split('/');
                
                if (replicaParts.length === 2) {
                    const [current, desired] = replicaParts;
                    
                    if (current === desired && current !== '0') {
                        logTest(`Service ${name}: ${replicas}`, 'passed');
                        testResults.passed.push(`Service ${name} healthy`);
                    } else {
                        logTest(`Service ${name}: ${replicas}`, 'warning');
                        testResults.warnings.push(`Service ${name} not fully replicated`);
                        allHealthy = false;
                    }
                }
            }
        }
        
        return allHealthy;
    } else {
        logTest('Cannot list services', 'failed');
        testResults.failed.push('Cannot list services');
        return false;
    }
}

async function testAPIEndpoint() {
    log('\n3. Testing API Endpoint', colors.blue);
    
    const healthCheck = await retryOperation(
        () => checkHttpEndpoint(`${config.baseUrl}/health`),
        'API health check'
    );
    
    if (healthCheck.success) {
        logTest('API health endpoint accessible', 'passed');
        testResults.passed.push('API health check passed');
    } else {
        logTest(`API health endpoint not accessible: ${healthCheck.error}`, 'failed');
        testResults.failed.push('API health check failed');
        return false;
    }
    
    // Test API authentication endpoint
    const authCheck = await checkHttpEndpoint(`${config.baseUrl}/api/v1/auth/status`);
    if (authCheck.success || authCheck.statusCode === 401) {
        logTest('API authentication endpoint responding', 'passed');
        testResults.passed.push('API auth endpoint responding');
    } else {
        logTest('API authentication endpoint not responding', 'warning');
        testResults.warnings.push('API auth endpoint issue');
    }
    
    return true;
}

async function testMonitoringStack() {
    log('\n4. Testing Monitoring Stack', colors.blue);
    
    // Test Prometheus
    const prometheusCheck = await retryOperation(
        () => checkHttpEndpoint(`${config.prometheusUrl}/api/v1/query?query=up`),
        'Prometheus'
    );
    
    if (prometheusCheck.success) {
        logTest('Prometheus accessible', 'passed');
        testResults.passed.push('Prometheus running');
        
        // Check if Prometheus is scraping targets
        try {
            const data = JSON.parse(prometheusCheck.data);
            if (data.data && data.data.result && data.data.result.length > 0) {
                logTest(`Prometheus scraping ${data.data.result.length} targets`, 'passed');
                testResults.passed.push('Prometheus scraping targets');
            }
        } catch (e) {
            logTest('Prometheus data format issue', 'warning');
            testResults.warnings.push('Prometheus data format issue');
        }
    } else {
        logTest('Prometheus not accessible', 'warning');
        testResults.warnings.push('Prometheus not accessible');
    }
    
    // Test Grafana
    const grafanaCheck = await retryOperation(
        () => checkHttpEndpoint(`${config.grafanaUrl}/api/health`),
        'Grafana'
    );
    
    if (grafanaCheck.success) {
        logTest('Grafana accessible', 'passed');
        testResults.passed.push('Grafana running');
    } else {
        logTest('Grafana not accessible', 'warning');
        testResults.warnings.push('Grafana not accessible');
    }
    
    // Test Kibana
    const kibanaCheck = await retryOperation(
        () => checkHttpEndpoint(`${config.kibanaUrl}/api/status`),
        'Kibana'
    );
    
    if (kibanaCheck.success) {
        logTest('Kibana accessible', 'passed');
        testResults.passed.push('Kibana running');
    } else {
        logTest('Kibana not accessible', 'warning');
        testResults.warnings.push('Kibana not accessible');
    }
    
    // Test Portainer
    const portainerCheck = await checkHttpEndpoint(config.portainerUrl);
    if (portainerCheck.success) {
        logTest('Portainer accessible', 'passed');
        testResults.passed.push('Portainer running');
    } else {
        logTest('Portainer not accessible', 'warning');
        testResults.warnings.push('Portainer not accessible');
    }
    
    return true;
}

async function testDataStores() {
    log('\n5. Testing Data Stores', colors.blue);
    
    // Test Redis connectivity
    const redisCheck = await executeCommand('docker exec $(docker ps -q -f name=redis-master) redis-cli ping');
    if (redisCheck.success && redisCheck.stdout.trim() === 'PONG') {
        logTest('Redis responding', 'passed');
        testResults.passed.push('Redis healthy');
    } else {
        logTest('Redis not responding', 'failed');
        testResults.failed.push('Redis unhealthy');
    }
    
    // Test PostgreSQL connectivity
    const postgresCheck = await executeCommand('docker exec $(docker ps -q -f name=postgres) pg_isready -U ollamamax');
    if (postgresCheck.success && postgresCheck.stdout.includes('accepting connections')) {
        logTest('PostgreSQL accepting connections', 'passed');
        testResults.passed.push('PostgreSQL healthy');
    } else {
        logTest('PostgreSQL not accepting connections', 'failed');
        testResults.failed.push('PostgreSQL unhealthy');
    }
    
    // Test Elasticsearch
    const elasticCheck = await checkHttpEndpoint('http://localhost:9200/_cluster/health');
    if (elasticCheck.success) {
        try {
            const health = JSON.parse(elasticCheck.data);
            if (health.status === 'green' || health.status === 'yellow') {
                logTest(`Elasticsearch cluster status: ${health.status}`, 'passed');
                testResults.passed.push('Elasticsearch healthy');
            } else {
                logTest(`Elasticsearch cluster status: ${health.status}`, 'warning');
                testResults.warnings.push('Elasticsearch degraded');
            }
        } catch (e) {
            logTest('Elasticsearch response parsing error', 'warning');
            testResults.warnings.push('Elasticsearch response issue');
        }
    } else {
        logTest('Elasticsearch not accessible', 'warning');
        testResults.warnings.push('Elasticsearch not accessible');
    }
    
    return true;
}

async function testSwarmCommunication() {
    log('\n6. Testing Inter-Service Communication', colors.blue);
    
    // Test service discovery
    const dnsCheck = await executeCommand('docker exec $(docker ps -q -f name=ollamamax-api -f status=running | head -1) nslookup redis-master');
    if (dnsCheck.success && !dnsCheck.stdout.includes('NXDOMAIN')) {
        logTest('Service discovery working', 'passed');
        testResults.passed.push('Service discovery working');
    } else {
        logTest('Service discovery issues', 'warning');
        testResults.warnings.push('Service discovery issues');
    }
    
    // Test network connectivity between services
    const networkCheck = await executeCommand('docker exec $(docker ps -q -f name=ollamamax-api -f status=running | head -1) ping -c 1 redis-master');
    if (networkCheck.success && networkCheck.stdout.includes('1 packets transmitted, 1 received')) {
        logTest('Inter-service networking working', 'passed');
        testResults.passed.push('Inter-service networking working');
    } else {
        logTest('Inter-service networking issues', 'warning');
        testResults.warnings.push('Inter-service networking issues');
    }
    
    return true;
}

async function testScaling() {
    log('\n7. Testing Service Scaling', colors.blue);
    
    // Test scaling up
    const scaleUp = await executeCommand('docker service scale ollamamax_ollamamax-api=5');
    if (scaleUp.success) {
        logTest('Service scale up command executed', 'passed');
        
        // Wait for scaling to complete
        await sleep(10000);
        
        // Check scaled replicas
        const checkScale = await executeCommand('docker service ls --filter name=ollamamax_ollamamax-api --format "{{.Replicas}}"');
        if (checkScale.success && checkScale.stdout.trim().startsWith('5/')) {
            logTest('Service successfully scaled to 5 replicas', 'passed');
            testResults.passed.push('Service scaling working');
            
            // Scale back down
            await executeCommand('docker service scale ollamamax_ollamamax-api=3');
        } else {
            logTest('Service scaling incomplete', 'warning');
            testResults.warnings.push('Service scaling incomplete');
        }
    } else {
        logTest('Service scale command failed', 'warning');
        testResults.warnings.push('Service scaling failed');
    }
    
    return true;
}

async function testLoadBalancing() {
    log('\n8. Testing Load Balancing', colors.blue);
    
    const responses = new Set();
    const requests = 10;
    
    for (let i = 0; i < requests; i++) {
        const response = await checkHttpEndpoint(`${config.baseUrl}/health`);
        if (response.success && response.data) {
            // In a real scenario, the response would include server ID
            responses.add(response.data);
        }
    }
    
    if (responses.size > 1) {
        logTest(`Load balancing working (${responses.size} different responses)`, 'passed');
        testResults.passed.push('Load balancing working');
    } else if (responses.size === 1) {
        logTest('Load balancing not verified (single response)', 'warning');
        testResults.warnings.push('Load balancing not verified');
    } else {
        logTest('Load balancing test failed', 'failed');
        testResults.failed.push('Load balancing failed');
    }
    
    return true;
}

// Main test runner
async function runTests() {
    console.log('');
    log('========================================', colors.blue);
    log('  OllamaMax Docker Swarm Integration Tests', colors.blue);
    log('========================================', colors.blue);
    
    const startTime = Date.now();
    
    // Run tests
    const swarmOk = await testDockerSwarm();
    if (!swarmOk) {
        log('\n⚠️  Docker Swarm not properly initialized. Skipping remaining tests.', colors.yellow);
    } else {
        await testStackDeployment();
        await testAPIEndpoint();
        await testMonitoringStack();
        await testDataStores();
        await testSwarmCommunication();
        await testScaling();
        await testLoadBalancing();
    }
    
    const duration = ((Date.now() - startTime) / 1000).toFixed(2);
    
    // Print summary
    console.log('');
    log('========================================', colors.blue);
    log('  Test Results Summary', colors.blue);
    log('========================================', colors.blue);
    console.log('');
    
    log(`✓ Passed: ${testResults.passed.length}`, colors.green);
    log(`⚠ Warnings: ${testResults.warnings.length}`, colors.yellow);
    log(`✗ Failed: ${testResults.failed.length}`, colors.red);
    log(`⏱ Duration: ${duration}s`, colors.blue);
    
    if (testResults.failed.length > 0) {
        console.log('');
        log('Failed Tests:', colors.red);
        testResults.failed.forEach(test => log(`  - ${test}`, colors.red));
    }
    
    if (testResults.warnings.length > 0) {
        console.log('');
        log('Warnings:', colors.yellow);
        testResults.warnings.forEach(warning => log(`  - ${warning}`, colors.yellow));
    }
    
    console.log('');
    
    // Exit with appropriate code
    process.exit(testResults.failed.length > 0 ? 1 : 0);
}

// Handle errors
process.on('unhandledRejection', (error) => {
    log(`\n✗ Unhandled error: ${error.message}`, colors.red);
    process.exit(1);
});

// Run tests
runTests();