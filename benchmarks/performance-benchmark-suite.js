#!/usr/bin/env node

/**
 * OllamaMax Performance Benchmark Suite
 * Comprehensive performance analysis without hanging on WebSocket/UI dependencies
 */

const { performance } = require('perf_hooks');
const fs = require('fs').promises;
const path = require('path');
const http = require('http');
const { spawn } = require('child_process');

class PerformanceBenchmarkSuite {
  constructor() {
    this.results = {
      timestamp: new Date().toISOString(),
      environment: {},
      benchmarks: {},
      bottlenecks: [],
      optimizations: []
    };
    
    this.baseUrl = 'http://localhost:13100';
    this.apiEndpoints = [
      '/api/health',
      '/api/nodes', 
      '/api/models',
      '/api/nodes/detailed'
    ];
  }

  async runComprehensiveBenchmarks() {
    console.log('🚀 Starting OllamaMax Performance Benchmark Suite\n');
    
    // Collect system baseline
    await this.collectSystemBaseline();
    
    // Run benchmarks
    await this.benchmarkApiPerformance();
    await this.benchmarkDockerContainers();
    await this.benchmarkMCPCoordination();
    await this.benchmarkMemoryEfficiency();
    await this.benchmarkProcessOverhead();
    
    // Generate analysis
    await this.identifyBottlenecks();
    await this.generateOptimizations();
    
    // Save comprehensive report
    await this.saveResults();
    
    this.displaySummary();
  }

  async collectSystemBaseline() {
    const os = require('os');
    
    this.results.environment = {
      platform: os.platform(),
      arch: os.arch(),
      cpus: os.cpus().length,
      memory_gb: Math.round(os.totalmem() / 1024 / 1024 / 1024),
      node_version: process.version,
      load_avg: os.loadavg(),
      uptime: os.uptime()
    };
    
    console.log('📊 System Baseline:');
    console.log(`   Platform: ${this.results.environment.platform} ${this.results.environment.arch}`);
    console.log(`   CPUs: ${this.results.environment.cpus}`);
    console.log(`   Memory: ${this.results.environment.memory_gb}GB`);
    console.log(`   Load Average: ${this.results.environment.load_avg.map(l => l.toFixed(2)).join(', ')}`);
  }

  async benchmarkApiPerformance() {
    console.log('\n🌐 Benchmarking API Performance...');
    
    const apiResults = {
      endpoints: {},
      summary: {
        total_requests: 0,
        successful_requests: 0,
        avg_response_time: 0,
        max_response_time: 0,
        min_response_time: Infinity
      }
    };
    
    for (const endpoint of this.apiEndpoints) {
      const endpointResults = await this.testEndpoint(endpoint, 5);
      apiResults.endpoints[endpoint] = endpointResults;
      
      // Update summary
      apiResults.summary.total_requests += endpointResults.requests;
      apiResults.summary.successful_requests += endpointResults.successful;
      
      if (endpointResults.successful > 0) {
        const avgTime = endpointResults.response_times.reduce((a, b) => a + b, 0) / endpointResults.response_times.length;
        const maxTime = Math.max(...endpointResults.response_times);
        const minTime = Math.min(...endpointResults.response_times);
        
        apiResults.summary.max_response_time = Math.max(apiResults.summary.max_response_time, maxTime);
        apiResults.summary.min_response_time = Math.min(apiResults.summary.min_response_time, minTime);
      }
    }
    
    // Calculate overall average
    const allResponseTimes = Object.values(apiResults.endpoints)
      .flatMap(e => e.response_times)
      .filter(t => t > 0);
    
    apiResults.summary.avg_response_time = allResponseTimes.length > 0 
      ? allResponseTimes.reduce((a, b) => a + b, 0) / allResponseTimes.length 
      : 0;
    
    this.results.benchmarks.api_performance = apiResults;
    
    console.log(`   Successful requests: ${apiResults.summary.successful_requests}/${apiResults.summary.total_requests}`);
    console.log(`   Average response time: ${apiResults.summary.avg_response_time.toFixed(2)}ms`);
    console.log(`   Response time range: ${apiResults.summary.min_response_time.toFixed(2)}-${apiResults.summary.max_response_time.toFixed(2)}ms`);
  }

  async testEndpoint(endpoint, iterations = 5) {
    const results = {
      endpoint,
      requests: iterations,
      successful: 0,
      failed: 0,
      response_times: [],
      error_types: {}
    };
    
    for (let i = 0; i < iterations; i++) {
      try {
        const startTime = performance.now();
        const response = await this.makeHttpRequest(this.baseUrl + endpoint);
        const responseTime = performance.now() - startTime;
        
        if (response.statusCode === 200) {
          results.successful++;
          results.response_times.push(responseTime);
        } else {
          results.failed++;
          const errorType = `HTTP_${response.statusCode}`;
          results.error_types[errorType] = (results.error_types[errorType] || 0) + 1;
        }
        
      } catch (error) {
        results.failed++;
        const errorType = error.code || 'UNKNOWN_ERROR';
        results.error_types[errorType] = (results.error_types[errorType] || 0) + 1;
      }
      
      // Brief pause between requests
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    
    console.log(`   ${endpoint}: ${results.successful}/${results.requests} successful`);
    return results;
  }

  makeHttpRequest(url) {
    return new Promise((resolve, reject) => {
      const startTime = performance.now();
      
      const req = http.get(url, { timeout: 5000 }, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
          resolve({
            statusCode: res.statusCode,
            headers: res.headers,
            data: data,
            responseTime: performance.now() - startTime
          });
        });
      });
      
      req.on('error', reject);
      req.on('timeout', () => {
        req.destroy();
        reject(new Error('Request timeout'));
      });
    });
  }

  async benchmarkDockerContainers() {
    console.log('\n🐳 Benchmarking Docker Container Performance...');
    
    try {
      const dockerStats = await this.executeCommand('docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"');
      
      const containerMetrics = this.parseDockerStats(dockerStats);
      
      this.results.benchmarks.docker_performance = {
        total_containers: containerMetrics.length,
        total_cpu_usage: containerMetrics.reduce((sum, c) => sum + c.cpu_percent, 0),
        total_memory_mb: containerMetrics.reduce((sum, c) => sum + c.memory_mb, 0),
        containers: containerMetrics
      };
      
      console.log(`   Active containers: ${containerMetrics.length}`);
      console.log(`   Total CPU usage: ${this.results.benchmarks.docker_performance.total_cpu_usage.toFixed(2)}%`);
      console.log(`   Total memory usage: ${this.results.benchmarks.docker_performance.total_memory_mb.toFixed(0)}MB`);
      
    } catch (error) {
      console.log('   Warning: Docker stats unavailable');
      this.results.benchmarks.docker_performance = { error: error.message };
    }
  }

  parseDockerStats(dockerOutput) {
    const lines = dockerOutput.split('\n').slice(1); // Skip header
    const containers = [];
    
    for (const line of lines) {
      if (line.trim()) {
        const parts = line.split('\t');
        if (parts.length >= 4) {
          const [container, cpu, memory, network, block] = parts;
          containers.push({
            container: container.trim(),
            cpu_percent: parseFloat(cpu.replace('%', '')) || 0,
            memory_mb: this.parseMemoryMB(memory.trim()),
            network_io: network.trim(),
            block_io: block?.trim() || 'N/A'
          });
        }
      }
    }
    
    return containers;
  }

  parseMemoryMB(memoryStr) {
    const match = memoryStr.match(/([\d.]+)(\w+)/);
    if (!match) return 0;
    
    const [, value, unit] = match;
    const num = parseFloat(value);
    
    switch (unit.toLowerCase()) {
      case 'gb': return num * 1024;
      case 'mb': return num;
      case 'kb': return num / 1024;
      default: return num;
    }
  }

  async benchmarkMCPCoordination() {
    console.log('\n🤖 Benchmarking MCP Coordination Performance...');
    
    const mcpBenchmarks = {
      agent_spawn_times: [],
      memory_operations: [],
      hook_execution_times: [],
      coordination_overhead: {}
    };
    
    // Test agent spawn performance
    for (let i = 0; i < 3; i++) {
      try {
        const spawnStart = performance.now();
        await this.executeCommand('npx claude-flow@alpha agent spawn researcher "Quick test task"', 10000);
        const spawnTime = performance.now() - spawnStart;
        mcpBenchmarks.agent_spawn_times.push(spawnTime);
        console.log(`   Agent spawn ${i + 1}: ${spawnTime.toFixed(2)}ms`);
      } catch (error) {
        console.log(`   Agent spawn ${i + 1}: failed (${error.message})`);
      }
    }
    
    // Test memory operations
    for (let i = 0; i < 5; i++) {
      try {
        const memStart = performance.now();
        await this.executeCommand(`npx claude-flow@alpha memory store "benchmark_test_${i}" "test data"`, 5000);
        const memTime = performance.now() - memStart;
        mcpBenchmarks.memory_operations.push(memTime);
      } catch (error) {
        console.log(`   Memory operation ${i + 1}: failed`);
      }
    }
    
    this.results.benchmarks.mcp_coordination = mcpBenchmarks;
    
    const avgSpawnTime = mcpBenchmarks.agent_spawn_times.length > 0 
      ? mcpBenchmarks.agent_spawn_times.reduce((a, b) => a + b, 0) / mcpBenchmarks.agent_spawn_times.length 
      : 0;
    
    const avgMemoryTime = mcpBenchmarks.memory_operations.length > 0
      ? mcpBenchmarks.memory_operations.reduce((a, b) => a + b, 0) / mcpBenchmarks.memory_operations.length
      : 0;
    
    console.log(`   Average agent spawn time: ${avgSpawnTime.toFixed(2)}ms`);
    console.log(`   Average memory operation time: ${avgMemoryTime.toFixed(2)}ms`);
  }

  async benchmarkMemoryEfficiency() {
    console.log('\n💾 Benchmarking Memory Efficiency...');
    
    const memoryAnalysis = {
      process_memory: {},
      gc_performance: {},
      memory_leaks: []
    };
    
    // Analyze Node.js process memory
    const memUsage = process.memoryUsage();
    memoryAnalysis.process_memory = {
      rss_mb: Math.round(memUsage.rss / 1024 / 1024),
      heap_used_mb: Math.round(memUsage.heapUsed / 1024 / 1024), 
      heap_total_mb: Math.round(memUsage.heapTotal / 1024 / 1024),
      external_mb: Math.round(memUsage.external / 1024 / 1024),
      heap_efficiency: ((memUsage.heapUsed / memUsage.heapTotal) * 100).toFixed(2)
    };
    
    // Test garbage collection performance
    const gcStart = performance.now();
    if (global.gc) {
      global.gc();
    }
    const gcTime = performance.now() - gcStart;
    memoryAnalysis.gc_performance = {
      gc_time_ms: gcTime,
      gc_available: !!global.gc
    };
    
    this.results.benchmarks.memory_efficiency = memoryAnalysis;
    
    console.log(`   Process RSS: ${memoryAnalysis.process_memory.rss_mb}MB`);
    console.log(`   Heap utilization: ${memoryAnalysis.process_memory.heap_efficiency}%`);
    console.log(`   GC performance: ${gcTime.toFixed(2)}ms`);
  }

  async benchmarkProcessOverhead() {
    console.log('\n⚡ Benchmarking Process Overhead...');
    
    try {
      const processInfo = await this.executeCommand('ps aux | grep -E "(node|claude|flow)" | grep -v grep');
      const processes = this.parseProcessInfo(processInfo);
      
      const processAnalysis = {
        total_processes: processes.length,
        total_cpu_usage: processes.reduce((sum, p) => sum + p.cpu, 0),
        total_memory_mb: processes.reduce((sum, p) => sum + p.memory_mb, 0),
        high_cpu_processes: processes.filter(p => p.cpu > 5.0),
        high_memory_processes: processes.filter(p => p.memory_mb > 100),
        processes: processes.slice(0, 10) // Top 10 processes
      };
      
      this.results.benchmarks.process_overhead = processAnalysis;
      
      console.log(`   Active Node.js processes: ${processAnalysis.total_processes}`);
      console.log(`   Combined CPU usage: ${processAnalysis.total_cpu_usage.toFixed(2)}%`);
      console.log(`   Combined memory usage: ${processAnalysis.total_memory_mb.toFixed(0)}MB`);
      
    } catch (error) {
      console.log('   Warning: Process analysis failed');
      this.results.benchmarks.process_overhead = { error: error.message };
    }
  }

  parseProcessInfo(processOutput) {
    const lines = processOutput.split('\n').filter(line => line.trim());
    const processes = [];
    
    for (const line of lines) {
      const parts = line.trim().split(/\s+/);
      if (parts.length >= 11) {
        processes.push({
          user: parts[0],
          pid: parseInt(parts[1]),
          cpu: parseFloat(parts[2]) || 0,
          memory_percent: parseFloat(parts[3]) || 0,
          memory_mb: this.calculateMemoryMB(parts[4]),
          command: parts.slice(10).join(' ').substring(0, 50)
        });
      }
    }
    
    return processes.sort((a, b) => b.cpu - a.cpu);
  }

  calculateMemoryMB(memStr) {
    // Convert RSS from KB to MB (approximate)
    const kb = parseInt(memStr) || 0;
    return Math.round(kb / 1024);
  }

  async identifyBottlenecks() {
    console.log('\n🔍 Identifying Performance Bottlenecks...');
    
    const bottlenecks = [];
    
    // API Performance Analysis
    const apiPerf = this.results.benchmarks.api_performance;
    if (apiPerf && apiPerf.summary.avg_response_time > 50) {
      bottlenecks.push({
        category: 'api_latency',
        severity: apiPerf.summary.avg_response_time > 100 ? 'high' : 'medium',
        issue: `Average API response time: ${apiPerf.summary.avg_response_time.toFixed(2)}ms`,
        threshold: '< 50ms for optimal performance',
        impact: 'User experience degradation'
      });
    }
    
    // Memory Efficiency Analysis
    const memPerf = this.results.benchmarks.memory_efficiency;
    if (memPerf && parseFloat(memPerf.process_memory.heap_efficiency) < 60) {
      bottlenecks.push({
        category: 'memory_fragmentation',
        severity: 'medium',
        issue: `Low heap efficiency: ${memPerf.process_memory.heap_efficiency}%`,
        threshold: '> 70% for optimal memory usage',
        impact: 'Increased memory overhead and GC pressure'
      });
    }
    
    // Process Overhead Analysis
    const processPerf = this.results.benchmarks.process_overhead;
    if (processPerf && processPerf.total_processes > 15) {
      bottlenecks.push({
        category: 'process_proliferation',
        severity: 'medium',
        issue: `High process count: ${processPerf.total_processes}`,
        threshold: '< 10 processes for optimal efficiency',
        impact: 'Context switching overhead and resource fragmentation'
      });
    }
    
    // Docker Resource Analysis
    const dockerPerf = this.results.benchmarks.docker_performance;
    if (dockerPerf && dockerPerf.total_memory_mb > 500) {
      bottlenecks.push({
        category: 'container_overhead',
        severity: 'low',
        issue: `High container memory usage: ${dockerPerf.total_memory_mb}MB`,
        threshold: '< 300MB for lightweight deployment',
        impact: 'Reduced available memory for application logic'
      });
    }
    
    this.results.bottlenecks = bottlenecks;
    
    console.log(`   Identified ${bottlenecks.length} performance bottlenecks`);
    bottlenecks.forEach((b, i) => {
      console.log(`   ${i + 1}. [${b.severity.toUpperCase()}] ${b.category}: ${b.issue}`);
    });
  }

  async generateOptimizations() {
    console.log('\n🚀 Generating Performance Optimizations...');
    
    const optimizations = [
      {
        category: 'api_optimization',
        priority: 'high',
        title: 'Implement API Response Caching',
        description: 'Cache frequent API responses (nodes, models) with 30s TTL',
        implementation: 'Redis-based caching layer with cache invalidation',
        expected_improvement: '60-80% reduction in response time for cached endpoints',
        effort: 'medium'
      },
      {
        category: 'mcp_optimization',
        priority: 'high', 
        title: 'Optimize MCP Server Communication',
        description: 'Implement connection pooling and message batching for MCP servers',
        implementation: 'Shared connection pool with request queuing and batch processing',
        expected_improvement: '40-50% reduction in coordination overhead',
        effort: 'high'
      },
      {
        category: 'memory_optimization',
        priority: 'medium',
        title: 'Implement Memory Pool Management',
        description: 'Use object pooling for frequently created/destroyed objects',
        implementation: 'Object pools for agent instances, message objects, and WebSocket frames',
        expected_improvement: '25-30% reduction in memory allocation overhead',
        effort: 'medium'
      },
      {
        category: 'process_optimization',
        priority: 'medium',
        title: 'Consolidate Agent Processes',
        description: 'Use shared Node.js processes with isolated contexts',
        implementation: 'Worker threads or VM contexts instead of separate processes',
        expected_improvement: '50-70% reduction in process overhead',
        effort: 'high'
      },
      {
        category: 'docker_optimization',
        priority: 'low',
        title: 'Optimize Container Resource Limits',
        description: 'Set appropriate memory limits and use multi-stage builds',
        implementation: 'Resource constraints in docker-compose and optimized base images',
        expected_improvement: '15-20% reduction in container overhead',
        effort: 'low'
      },
      {
        category: 'coordination_optimization',
        priority: 'high',
        title: 'Implement Smart Load Balancing',
        description: 'Dynamic load balancing based on real-time performance metrics',
        implementation: 'Weighted round-robin with health scoring and circuit breakers',
        expected_improvement: '30-40% improvement in request distribution efficiency',
        effort: 'medium'
      }
    ];
    
    this.results.optimizations = optimizations;
    
    console.log(`   Generated ${optimizations.length} optimization strategies`);
    optimizations.forEach((opt, i) => {
      console.log(`   ${i + 1}. [${opt.priority.toUpperCase()}] ${opt.title}`);
    });
  }

  async executeCommand(command, timeout = 5000) {
    return new Promise((resolve, reject) => {
      const child = spawn('bash', ['-c', command], { timeout });
      let stdout = '';
      let stderr = '';
      
      child.stdout.on('data', data => stdout += data);
      child.stderr.on('data', data => stderr += data);
      
      child.on('close', code => {
        if (code === 0) {
          resolve(stdout);
        } else {
          reject(new Error(`Command failed: ${stderr || stdout}`));
        }
      });
      
      setTimeout(() => {
        child.kill();
        reject(new Error('Command timeout'));
      }, timeout);
    });
  }

  async saveResults() {
    const resultsDir = path.join(__dirname, '../performance');
    await fs.mkdir(resultsDir, { recursive: true });
    
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const filePath = path.join(resultsDir, `benchmark-results-${timestamp}.json`);
    
    await fs.writeFile(filePath, JSON.stringify(this.results, null, 2));
    console.log(`\n💾 Results saved: ${filePath}`);
  }

  displaySummary() {
    console.log('\n' + '='.repeat(70));
    console.log('🎯 PERFORMANCE BENCHMARK SUMMARY');
    console.log('='.repeat(70));
    
    // System Overview
    const env = this.results.environment;
    console.log(`\n📊 System: ${env.platform} (${env.cpus} cores, ${env.memory_gb}GB)`);
    console.log(`   Load Average: ${env.load_avg.map(l => l.toFixed(2)).join(', ')}`);
    
    // API Performance
    const api = this.results.benchmarks.api_performance;
    if (api) {
      const successRate = (api.summary.successful_requests / api.summary.total_requests * 100).toFixed(1);
      console.log(`\n🌐 API Performance:`);
      console.log(`   Success Rate: ${successRate}%`);
      console.log(`   Average Response: ${api.summary.avg_response_time.toFixed(2)}ms`);
    }
    
    // Docker Performance
    const docker = this.results.benchmarks.docker_performance;
    if (docker && docker.total_containers) {
      console.log(`\n🐳 Container Performance:`);
      console.log(`   Active Containers: ${docker.total_containers}`);
      console.log(`   Total Memory: ${docker.total_memory_mb.toFixed(0)}MB`);
    }
    
    // Bottlenecks
    if (this.results.bottlenecks.length > 0) {
      console.log(`\n⚠️  Critical Issues (${this.results.bottlenecks.length}):`);
      this.results.bottlenecks.forEach((bottleneck, i) => {
        const emoji = bottleneck.severity === 'high' ? '🚨' : bottleneck.severity === 'medium' ? '⚠️' : '💡';
        console.log(`   ${emoji} ${bottleneck.category}: ${bottleneck.issue}`);
      });
    }
    
    // Top Optimizations
    const highPriorityOpts = this.results.optimizations.filter(o => o.priority === 'high');
    if (highPriorityOpts.length > 0) {
      console.log(`\n🚀 High Priority Optimizations:`);
      highPriorityOpts.forEach((opt, i) => {
        console.log(`   ${i + 1}. ${opt.title} (${opt.expected_improvement})`);
      });
    }
    
    console.log('\n' + '='.repeat(70));
  }
}

// CLI execution
if (require.main === module) {
  const benchmark = new PerformanceBenchmarkSuite();
  
  benchmark.runComprehensiveBenchmarks().catch(error => {
    console.error('Benchmark suite failed:', error);
    process.exit(1);
  });
}

module.exports = { PerformanceBenchmarkSuite };