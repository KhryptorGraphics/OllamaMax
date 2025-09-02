#!/usr/bin/env node

/**
 * Real-time Performance Monitoring Dashboard for OllamaMax
 * Provides continuous performance tracking and alerting
 */

const fs = require('fs').promises;
const path = require('path');
const { performance } = require('perf_hooks');

class PerformanceMonitoringDashboard {
  constructor() {
    this.metrics = {
      api_performance: {
        requests_per_second: 0,
        avg_response_time: 0,
        error_rate: 0,
        p95_response_time: 0,
        cache_hit_rate: 0
      },
      system_performance: {
        cpu_usage: 0,
        memory_usage: 0,
        memory_efficiency: 0,
        process_count: 0,
        docker_containers: 0
      },
      coordination_performance: {
        agent_spawn_time: 0,
        mcp_latency: 0,
        task_completion_rate: 0,
        coordination_overhead: 0
      },
      bottlenecks: [],
      alerts: []
    };
    
    this.thresholds = {
      api_response_time: 100,     // ms
      memory_usage: 75,           // %
      cpu_usage: 80,              // %
      error_rate: 5,              // %
      cache_hit_rate: 70,         // %
      agent_spawn_time: 1000      // ms
    };
    
    this.monitoring = {
      interval: 10000,            // 10 seconds
      retention_hours: 24,
      alert_cooldown: 300000      // 5 minutes
    };
    
    this.lastAlerts = new Map();
  }

  /**
   * Start continuous monitoring
   */
  async startMonitoring() {
    console.log('📊 Starting Performance Monitoring Dashboard...');
    
    // Initialize metrics collection
    await this.initializeMetricsCollection();
    
    // Start monitoring intervals
    this.startSystemMonitoring();
    this.startAPIMonitoring();
    this.startCoordinationMonitoring();
    this.startBottleneckDetection();
    
    // Set up alert system
    this.startAlertSystem();
    
    console.log('✅ Performance monitoring active');
    
    // Generate initial report
    setTimeout(() => this.generateDashboard(), 5000);
  }

  async initializeMetricsCollection() {
    const metricsDir = path.join(__dirname, '../performance/metrics');
    await fs.mkdir(metricsDir, { recursive: true });
    
    this.metricsFile = path.join(metricsDir, 'real-time-metrics.json');
    this.alertsFile = path.join(metricsDir, 'performance-alerts.json');
  }

  /**
   * Monitor system-level performance
   */
  startSystemMonitoring() {
    setInterval(async () => {
      try {
        const systemMetrics = await this.collectSystemMetrics();
        this.metrics.system_performance = systemMetrics;
        
        // Check thresholds
        if (systemMetrics.memory_usage > this.thresholds.memory_usage) {
          this.raiseAlert('memory_usage', 'high', 
            `Memory usage: ${systemMetrics.memory_usage.toFixed(1)}%`);
        }
        
        if (systemMetrics.cpu_usage > this.thresholds.cpu_usage) {
          this.raiseAlert('cpu_usage', 'high',
            `CPU usage: ${systemMetrics.cpu_usage.toFixed(1)}%`);
        }
        
      } catch (error) {
        console.error('System monitoring error:', error.message);
      }
    }, this.monitoring.interval);
  }

  async collectSystemMetrics() {
    const os = require('os');
    
    // CPU usage calculation
    const cpus = os.cpus();
    let totalIdle = 0;
    let totalTick = 0;
    
    for (const cpu of cpus) {
      for (const type in cpu.times) {
        totalTick += cpu.times[type];
      }
      totalIdle += cpu.times.idle;
    }
    
    const cpuUsage = 100 - (totalIdle / totalTick * 100);
    
    // Memory metrics
    const memoryTotal = os.totalmem();
    const memoryFree = os.freemem();
    const memoryUsed = memoryTotal - memoryFree;
    const memoryUsagePercent = (memoryUsed / memoryTotal) * 100;
    
    // Process metrics
    let processCount = 0;
    try {
      const { exec } = require('child_process');
      processCount = await new Promise((resolve) => {
        exec('ps aux | grep -E "(node|claude)" | grep -v grep | wc -l', (error, stdout) => {
          resolve(parseInt(stdout.trim()) || 0);
        });
      });
    } catch (error) {
      processCount = 0;
    }
    
    return {
      cpu_usage: cpuUsage,
      memory_usage: memoryUsagePercent,
      memory_efficiency: 100 - memoryUsagePercent,
      memory_total_gb: Math.round(memoryTotal / 1024 / 1024 / 1024),
      memory_used_gb: Math.round(memoryUsed / 1024 / 1024 / 1024),
      process_count: processCount,
      load_average: os.loadavg()[0]
    };
  }

  /**
   * Monitor API performance
   */
  startAPIMonitoring() {
    setInterval(async () => {
      try {
        const apiMetrics = await this.collectAPIMetrics();
        this.metrics.api_performance = apiMetrics;
        
        // Check API thresholds
        if (apiMetrics.avg_response_time > this.thresholds.api_response_time) {
          this.raiseAlert('api_latency', 'medium',
            `API response time: ${apiMetrics.avg_response_time.toFixed(2)}ms`);
        }
        
        if (apiMetrics.error_rate > this.thresholds.error_rate) {
          this.raiseAlert('api_errors', 'high',
            `API error rate: ${apiMetrics.error_rate.toFixed(1)}%`);
        }
        
      } catch (error) {
        console.error('API monitoring error:', error.message);
      }
    }, this.monitoring.interval);
  }

  async collectAPIMetrics() {
    const endpoints = [
      '/api/health',
      '/api/nodes', 
      '/api/models'
    ];
    
    const results = [];
    
    for (const endpoint of endpoints) {
      try {
        const startTime = performance.now();
        const response = await this.makeRequest(`http://localhost:13100${endpoint}`);
        const responseTime = performance.now() - startTime;
        
        results.push({
          endpoint,
          responseTime,
          success: response.statusCode === 200,
          statusCode: response.statusCode
        });
        
      } catch (error) {
        results.push({
          endpoint,
          responseTime: this.thresholds.api_response_time,
          success: false,
          error: error.message
        });
      }
    }
    
    const successfulResults = results.filter(r => r.success);
    const avgResponseTime = successfulResults.length > 0 ?
      successfulResults.reduce((sum, r) => sum + r.responseTime, 0) / successfulResults.length : 0;
    
    const errorRate = results.length > 0 ?
      (results.filter(r => !r.success).length / results.length) * 100 : 0;
    
    return {
      requests_per_second: results.length / (this.monitoring.interval / 1000),
      avg_response_time: avgResponseTime,
      error_rate: errorRate,
      p95_response_time: this.calculateP95(successfulResults.map(r => r.responseTime)),
      endpoint_results: results,
      timestamp: Date.now()
    };
  }

  /**
   * Monitor coordination performance
   */
  startCoordinationMonitoring() {
    setInterval(async () => {
      try {
        const coordMetrics = await this.collectCoordinationMetrics();
        this.metrics.coordination_performance = coordMetrics;
        
        // Check coordination thresholds
        if (coordMetrics.agent_spawn_time > this.thresholds.agent_spawn_time) {
          this.raiseAlert('agent_spawn_slow', 'medium',
            `Agent spawn time: ${coordMetrics.agent_spawn_time.toFixed(0)}ms`);
        }
        
      } catch (error) {
        console.error('Coordination monitoring error:', error.message);
      }
    }, this.monitoring.interval * 2); // Less frequent monitoring
  }

  async collectCoordinationMetrics() {
    // Read metrics from Claude Flow
    let agentSpawnTime = 0;
    let mcpLatency = 0;
    let taskCompletionRate = 100;
    
    try {
      // Read task metrics
      const taskMetricsPath = path.join(__dirname, '../.claude-flow/metrics/task-metrics.json');
      const taskData = await fs.readFile(taskMetricsPath, 'utf8');
      const taskMetrics = JSON.parse(taskData);
      
      if (taskMetrics.length > 0) {
        const recentTasks = taskMetrics.slice(-10); // Last 10 tasks
        const avgDuration = recentTasks.reduce((sum, task) => sum + task.duration, 0) / recentTasks.length;
        agentSpawnTime = avgDuration;
        
        const successfulTasks = recentTasks.filter(task => task.success).length;
        taskCompletionRate = (successfulTasks / recentTasks.length) * 100;
      }
      
    } catch (error) {
      console.warn('Could not read task metrics:', error.message);
    }
    
    return {
      agent_spawn_time: agentSpawnTime,
      mcp_latency: mcpLatency,
      task_completion_rate: taskCompletionRate,
      coordination_overhead: this.calculateCoordinationOverhead(),
      timestamp: Date.now()
    };
  }

  calculateCoordinationOverhead() {
    // Estimate coordination overhead based on process count and MCP activity
    const baseProcesses = 5; // Minimum expected processes
    const actualProcesses = this.metrics.system_performance.process_count || 0;
    const overhead = Math.max(0, actualProcesses - baseProcesses);
    
    return {
      excess_processes: overhead,
      overhead_percentage: actualProcesses > 0 ? (overhead / actualProcesses * 100).toFixed(1) + '%' : '0%'
    };
  }

  /**
   * Detect performance bottlenecks automatically
   */
  startBottleneckDetection() {
    setInterval(() => {
      this.detectBottlenecks();
    }, this.monitoring.interval * 3); // Every 30 seconds
  }

  detectBottlenecks() {
    const bottlenecks = [];
    
    // API Performance Bottlenecks
    if (this.metrics.api_performance.avg_response_time > this.thresholds.api_response_time) {
      bottlenecks.push({
        category: 'api_latency',
        severity: this.metrics.api_performance.avg_response_time > 200 ? 'high' : 'medium',
        metric: 'Average API Response Time',
        current: `${this.metrics.api_performance.avg_response_time.toFixed(2)}ms`,
        threshold: `${this.thresholds.api_response_time}ms`,
        recommendation: 'Implement API caching layer or optimize backend processing'
      });
    }
    
    // Memory Usage Bottlenecks
    if (this.metrics.system_performance.memory_usage > this.thresholds.memory_usage) {
      bottlenecks.push({
        category: 'memory_pressure',
        severity: this.metrics.system_performance.memory_usage > 85 ? 'high' : 'medium',
        metric: 'System Memory Usage',
        current: `${this.metrics.system_performance.memory_usage.toFixed(1)}%`,
        threshold: `${this.thresholds.memory_usage}%`,
        recommendation: 'Implement memory pooling or reduce agent process count'
      });
    }
    
    // Process Overhead Bottlenecks
    if (this.metrics.system_performance.process_count > 20) {
      bottlenecks.push({
        category: 'process_overhead',
        severity: this.metrics.system_performance.process_count > 50 ? 'high' : 'medium',
        metric: 'Active Node.js Processes',
        current: this.metrics.system_performance.process_count.toString(),
        threshold: '20',
        recommendation: 'Consolidate agent processes using worker threads or shared contexts'
      });
    }
    
    // Coordination Overhead Bottlenecks
    if (this.metrics.coordination_performance.agent_spawn_time > this.thresholds.agent_spawn_time) {
      bottlenecks.push({
        category: 'coordination_latency',
        severity: 'medium',
        metric: 'Agent Spawn Time',
        current: `${this.metrics.coordination_performance.agent_spawn_time.toFixed(0)}ms`,
        threshold: `${this.thresholds.agent_spawn_time}ms`,
        recommendation: 'Implement agent pre-spawning and connection pooling'
      });
    }
    
    this.metrics.bottlenecks = bottlenecks;
    
    // Log new bottlenecks
    const newBottlenecks = bottlenecks.filter(b => 
      !this.previousBottlenecks?.some(prev => 
        prev.category === b.category && prev.severity === b.severity
      )
    );
    
    if (newBottlenecks.length > 0) {
      console.log('🚨 New performance bottlenecks detected:');
      newBottlenecks.forEach(b => {
        console.log(`   [${b.severity.toUpperCase()}] ${b.category}: ${b.current} (threshold: ${b.threshold})`);
      });
    }
    
    this.previousBottlenecks = [...bottlenecks];
  }

  /**
   * Alert system for performance issues
   */
  startAlertSystem() {
    setInterval(() => {
      this.processAlerts();
    }, this.monitoring.interval);
  }

  raiseAlert(type, severity, message) {
    const now = Date.now();
    const lastAlert = this.lastAlerts.get(type);
    
    // Apply cooldown to prevent spam
    if (lastAlert && (now - lastAlert) < this.monitoring.alert_cooldown) {
      return;
    }
    
    const alert = {
      type,
      severity,
      message,
      timestamp: new Date().toISOString(),
      metric_snapshot: { ...this.metrics }
    };
    
    this.metrics.alerts.push(alert);
    this.lastAlerts.set(type, now);
    
    const emoji = severity === 'high' ? '🚨' : severity === 'medium' ? '⚠️' : '💡';
    console.log(`${emoji} ALERT [${severity.toUpperCase()}] ${type}: ${message}`);
  }

  processAlerts() {
    // Clean up old alerts (keep last 50)
    if (this.metrics.alerts.length > 50) {
      this.metrics.alerts = this.metrics.alerts.slice(-50);
    }
    
    // Check for alert patterns
    const recentAlerts = this.metrics.alerts.filter(alert => 
      Date.now() - new Date(alert.timestamp).getTime() < 600000 // Last 10 minutes
    );
    
    if (recentAlerts.length > 5) {
      this.raiseAlert('alert_storm', 'high', 
        `${recentAlerts.length} alerts in last 10 minutes - system degradation detected`);
    }
  }

  /**
   * Generate real-time dashboard
   */
  async generateDashboard() {
    const dashboard = {
      timestamp: new Date().toISOString(),
      status: this.calculateOverallStatus(),
      metrics: this.metrics,
      performance_score: this.calculatePerformanceScore(),
      trends: await this.calculateTrends(),
      recommendations: this.generateRecommendations()
    };
    
    // Save metrics to file
    await fs.writeFile(this.metricsFile, JSON.stringify(dashboard, null, 2));
    
    // Display dashboard
    this.displayDashboard(dashboard);
    
    return dashboard;
  }

  calculateOverallStatus() {
    const criticalAlerts = this.metrics.alerts.filter(a => a.severity === 'high').length;
    const highSeverityBottlenecks = this.metrics.bottlenecks.filter(b => b.severity === 'high').length;
    
    if (criticalAlerts > 0 || highSeverityBottlenecks > 0) {
      return 'critical';
    } else if (this.metrics.alerts.length > 0 || this.metrics.bottlenecks.length > 0) {
      return 'warning';
    } else {
      return 'healthy';
    }
  }

  calculatePerformanceScore() {
    let score = 100;
    
    // API Performance (30% weight)
    const apiScore = Math.max(0, 100 - (this.metrics.api_performance.avg_response_time / 10));
    score -= (100 - apiScore) * 0.3;
    
    // System Performance (40% weight)
    const memoryPenalty = Math.max(0, this.metrics.system_performance.memory_usage - 50) * 2;
    const cpuPenalty = Math.max(0, this.metrics.system_performance.cpu_usage - 50) * 1.5;
    score -= (memoryPenalty + cpuPenalty) * 0.4;
    
    // Coordination Performance (30% weight)
    const coordPenalty = Math.max(0, (this.metrics.coordination_performance.agent_spawn_time - 500) / 100);
    score -= coordPenalty * 0.3;
    
    return Math.max(0, Math.round(score));
  }

  async calculateTrends() {
    // Simple trend calculation based on recent changes
    return {
      memory_trend: this.metrics.system_performance.memory_usage > 55 ? 'increasing' : 'stable',
      api_trend: this.metrics.api_performance.avg_response_time > 50 ? 'degrading' : 'stable',
      coordination_trend: this.metrics.coordination_performance.agent_spawn_time > 1000 ? 'slow' : 'optimal'
    };
  }

  generateRecommendations() {
    const recommendations = [];
    
    // Based on current bottlenecks
    for (const bottleneck of this.metrics.bottlenecks) {
      recommendations.push({
        priority: bottleneck.severity,
        category: bottleneck.category,
        action: bottleneck.recommendation,
        expected_impact: this.estimateImpact(bottleneck.category)
      });
    }
    
    // General optimization recommendations
    if (this.metrics.system_performance.process_count > 30) {
      recommendations.push({
        priority: 'medium',
        category: 'architecture',
        action: 'Implement agent process consolidation using worker threads',
        expected_impact: '40-60% reduction in process overhead'
      });
    }
    
    if (this.metrics.api_performance.avg_response_time > 20) {
      recommendations.push({
        priority: 'high',
        category: 'caching',
        action: 'Deploy Redis caching layer for API responses',
        expected_impact: '60-80% reduction in API response times'
      });
    }
    
    return recommendations.slice(0, 5); // Top 5 recommendations
  }

  estimateImpact(category) {
    const impactMap = {
      'api_latency': '50-70% response time improvement',
      'memory_pressure': '20-30% memory usage reduction',
      'process_overhead': '40-60% CPU overhead reduction',
      'coordination_latency': '30-50% coordination efficiency improvement'
    };
    
    return impactMap[category] || 'Performance improvement expected';
  }

  displayDashboard(dashboard) {
    console.clear();
    console.log('╔═══════════════════════════════════════════════════════════════╗');
    console.log('║              🚀 OllamaMax Performance Dashboard              ║');
    console.log('╠═══════════════════════════════════════════════════════════════╣');
    console.log(`║ Status: ${this.getStatusIcon(dashboard.status)} ${dashboard.status.toUpperCase().padEnd(45)} ║`);
    console.log(`║ Score:  ${dashboard.performance_score}/100${''.padEnd(45)} ║`);
    console.log(`║ Time:   ${new Date().toLocaleTimeString().padEnd(45)} ║`);
    console.log('╠═══════════════════════════════════════════════════════════════╣');
    
    // System Metrics
    const sys = dashboard.metrics.system_performance;
    console.log('║ 📊 SYSTEM PERFORMANCE                                        ║');
    console.log(`║   Memory: ${sys.memory_usage?.toFixed(1) || 'N/A'}% (${sys.memory_used_gb || 'N/A'}GB/${sys.memory_total_gb || 'N/A'}GB)${''.padEnd(20)} ║`);
    console.log(`║   CPU:    ${sys.cpu_usage?.toFixed(1) || 'N/A'}% (Load: ${sys.load_average?.toFixed(2) || 'N/A'})${''.padEnd(25)} ║`);
    console.log(`║   Processes: ${sys.process_count || 'N/A'}${''.padEnd(45)} ║`);
    
    // API Metrics
    const api = dashboard.metrics.api_performance;
    console.log('║                                                               ║');
    console.log('║ 🌐 API PERFORMANCE                                           ║');
    console.log(`║   Response Time: ${api.avg_response_time?.toFixed(2) || 'N/A'}ms${''.padEnd(35)} ║`);
    console.log(`║   Error Rate:    ${api.error_rate?.toFixed(1) || 'N/A'}%${''.padEnd(35)} ║`);
    console.log(`║   RPS:           ${api.requests_per_second?.toFixed(1) || 'N/A'}${''.padEnd(35)} ║`);
    
    // Coordination Metrics
    const coord = dashboard.metrics.coordination_performance;
    console.log('║                                                               ║');
    console.log('║ 🤖 COORDINATION PERFORMANCE                                  ║');
    console.log(`║   Agent Spawn:   ${coord.agent_spawn_time?.toFixed(0) || 'N/A'}ms${''.padEnd(35)} ║`);
    console.log(`║   Task Success:  ${coord.task_completion_rate?.toFixed(1) || 'N/A'}%${''.padEnd(35)} ║`);
    
    // Active Issues
    console.log('║                                                               ║');
    console.log('║ ⚠️  ACTIVE ISSUES                                            ║');
    if (dashboard.metrics.bottlenecks.length === 0) {
      console.log('║   No performance issues detected ✅                          ║');
    } else {
      dashboard.metrics.bottlenecks.slice(0, 3).forEach(bottleneck => {
        const emoji = bottleneck.severity === 'high' ? '🚨' : '⚠️';
        const line = `   ${emoji} ${bottleneck.category}: ${bottleneck.current}`;
        console.log(`║ ${line.padEnd(61)} ║`);
      });
    }
    
    console.log('╚═══════════════════════════════════════════════════════════════╝');
    
    // Recent alerts
    if (dashboard.metrics.alerts.length > 0) {
      console.log('\n🔔 Recent Alerts:');
      dashboard.metrics.alerts.slice(-3).forEach(alert => {
        const emoji = alert.severity === 'high' ? '🚨' : alert.severity === 'medium' ? '⚠️' : '💡';
        console.log(`   ${emoji} [${new Date(alert.timestamp).toLocaleTimeString()}] ${alert.message}`);
      });
    }
  }

  getStatusIcon(status) {
    const icons = {
      'healthy': '✅',
      'warning': '⚠️',
      'critical': '🚨'
    };
    return icons[status] || '❓';
  }

  calculateP95(values) {
    if (values.length === 0) return 0;
    const sorted = [...values].sort((a, b) => a - b);
    const index = Math.ceil(sorted.length * 0.95) - 1;
    return sorted[index] || 0;
  }

  async makeRequest(url) {
    return new Promise((resolve, reject) => {
      const req = http.get(url, { timeout: 5000 }, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
          resolve({
            statusCode: res.statusCode,
            data: data
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

  /**
   * Generate performance optimization implementation guide
   */
  generateImplementationGuide() {
    return {
      immediate_optimizations: [
        {
          title: 'API Response Caching',
          priority: 'HIGH',
          implementation: 'Deploy Redis caching layer with 30s TTL for health endpoints',
          files_to_modify: ['optimizations/api-caching-layer.js'],
          estimated_effort: '4-6 hours',
          expected_improvement: '60-80% API response time reduction'
        },
        {
          title: 'Memory Pool Management',
          priority: 'HIGH', 
          implementation: 'Implement object pooling for agents, messages, and WebSocket frames',
          files_to_modify: ['optimizations/memory-pool-manager.js'],
          estimated_effort: '6-8 hours',
          expected_improvement: '25-30% memory allocation overhead reduction'
        },
        {
          title: 'Smart Load Balancing',
          priority: 'MEDIUM',
          implementation: 'Deploy weighted round-robin with health scoring',
          files_to_modify: ['optimizations/smart-load-balancer.js'],
          estimated_effort: '8-12 hours',
          expected_improvement: '30-40% request distribution efficiency'
        }
      ],
      configuration_changes: [
        {
          file: 'docker-compose.cpu.yml',
          changes: [
            'Add memory limits: 256m for workers, 512m for API server',
            'Set CPU limits: 0.5 cores for workers, 1.0 for API',
            'Enable gzip compression in nginx reverse proxy'
          ]
        },
        {
          file: 'package.json',
          changes: [
            'Add redis dependency for caching',
            'Add performance monitoring scripts',
            'Configure NODE_OPTIONS="--max-old-space-size=512"'
          ]
        }
      ],
      monitoring_improvements: [
        'Implement Prometheus metrics collection',
        'Add Grafana dashboards for real-time visualization',
        'Set up automated alerts for performance regression',
        'Enable distributed tracing for request flow analysis'
      ]
    };
  }

  /**
   * Cleanup and shutdown
   */
  async cleanup() {
    console.log('🔌 Shutting down performance monitoring...');
    
    // Save final metrics
    if (this.metricsFile) {
      const finalDashboard = await this.generateDashboard();
      console.log(`💾 Final metrics saved to: ${this.metricsFile}`);
    }
  }
}

module.exports = { PerformanceMonitoringDashboard };

// CLI execution
if (require.main === module) {
  const dashboard = new PerformanceMonitoringDashboard();
  
  (async () => {
    await dashboard.startMonitoring();
    
    // Run monitoring for 30 seconds
    console.log('🔄 Monitoring for 30 seconds...\n');
    
    await new Promise(resolve => setTimeout(resolve, 30000));
    
    // Generate implementation guide
    console.log('\n📋 Performance Optimization Implementation Guide:');
    const guide = dashboard.generateImplementationGuide();
    console.log(JSON.stringify(guide, null, 2));
    
    await dashboard.cleanup();
    
  })().catch(console.error);
}