#!/usr/bin/env node

/**
 * Agent Metrics HTTP Server
 * Provides HTTP endpoints for Prometheus scraping and API access
 */

const express = require('express');
const cors = require('cors');
const { performance } = require('perf_hooks');
const AgentMetricsCollector = require('./metrics-collector');

class MetricsServer {
  constructor(port = 8080) {
    this.app = express();
    this.port = port;
    this.collector = AgentMetricsCollector.instance;
    
    this.setupMiddleware();
    this.setupRoutes();
  }

  setupMiddleware() {
    this.app.use(cors());
    this.app.use(express.json({ limit: '10mb' }));
    this.app.use(express.urlencoded({ extended: true }));
    
    // Request logging
    this.app.use((req, res, next) => {
      const start = performance.now();
      res.on('finish', () => {
        const duration = performance.now() - start;
        console.log(`${req.method} ${req.url} - ${res.statusCode} - ${duration.toFixed(2)}ms`);
      });
      next();
    });
  }

  setupRoutes() {
    // Health check
    this.app.get('/health', (req, res) => {
      res.json({ status: 'healthy', timestamp: Date.now() });
    });

    // Prometheus metrics endpoint
    this.app.get('/metrics', async (req, res) => {
      try {
        const metrics = await this.generatePrometheusMetrics();
        res.set('Content-Type', 'text/plain; charset=utf-8');
        res.send(metrics);
      } catch (error) {
        console.error('Error generating Prometheus metrics:', error);
        res.status(500).json({ error: 'Failed to generate metrics' });
      }
    });

    // API endpoints for dashboard
    this.app.get('/api/agents/:agentId/metrics', async (req, res) => {
      try {
        const { agentId } = req.params;
        const { timeRange = '1h' } = req.query;
        
        const metrics = await this.collector.getAgentMetrics(agentId, timeRange);
        res.json(metrics);
      } catch (error) {
        console.error('Error fetching agent metrics:', error);
        res.status(500).json({ error: 'Failed to fetch agent metrics' });
      }
    });

    this.app.get('/api/timeseries/:metric', async (req, res) => {
      try {
        const { metric } = req.params;
        const { resolution = '1m', timeRange = '1h' } = req.query;
        
        const data = await this.collector.getTimeSeries(metric, resolution, timeRange);
        res.json(data);
      } catch (error) {
        console.error('Error fetching time series:', error);
        res.status(500).json({ error: 'Failed to fetch time series data' });
      }
    });

    this.app.get('/api/system/health', async (req, res) => {
      try {
        const health = await this.collector.getSystemHealth();
        res.json(health);
      } catch (error) {
        console.error('Error fetching system health:', error);
        res.status(500).json({ error: 'Failed to fetch system health' });
      }
    });

    // Submit metrics endpoint
    this.app.post('/api/metrics/agent', async (req, res) => {
      try {
        const { agentId, taskId, metrics } = req.body;
        
        if (!agentId || !taskId || !metrics) {
          return res.status(400).json({ error: 'Missing required fields' });
        }

        const result = await this.collector.collectAgentMetric(agentId, taskId, metrics);
        res.json({ success: true, metric: result });
      } catch (error) {
        console.error('Error collecting agent metric:', error);
        res.status(500).json({ error: 'Failed to collect metric' });
      }
    });

    this.app.post('/api/metrics/swarm', async (req, res) => {
      try {
        const metrics = req.body;
        
        if (!metrics || typeof metrics !== 'object') {
          return res.status(400).json({ error: 'Invalid metrics data' });
        }

        const result = await this.collector.collectSwarmMetric(metrics);
        res.json({ success: true, metric: result });
      } catch (error) {
        console.error('Error collecting swarm metric:', error);
        res.status(500).json({ error: 'Failed to collect swarm metric' });
      }
    });

    // Batch metrics submission
    this.app.post('/api/metrics/batch', async (req, res) => {
      try {
        const { metrics } = req.body;
        
        if (!Array.isArray(metrics)) {
          return res.status(400).json({ error: 'Metrics must be an array' });
        }

        const results = [];
        for (const metric of metrics) {
          if (metric.type === 'agent') {
            const result = await this.collector.collectAgentMetric(
              metric.agentId, 
              metric.taskId, 
              metric.data
            );
            results.push(result);
          } else if (metric.type === 'swarm') {
            const result = await this.collector.collectSwarmMetric(metric.data);
            results.push(result);
          }
        }

        res.json({ success: true, processed: results.length, results });
      } catch (error) {
        console.error('Error processing batch metrics:', error);
        res.status(500).json({ error: 'Failed to process batch metrics' });
      }
    });

    // Dashboard data aggregation endpoints
    this.app.get('/api/dashboard/overview', async (req, res) => {
      try {
        const overview = await this.generateDashboardOverview();
        res.json(overview);
      } catch (error) {
        console.error('Error generating dashboard overview:', error);
        res.status(500).json({ error: 'Failed to generate overview' });
      }
    });
  }

  async generatePrometheusMetrics() {
    const health = await this.collector.getSystemHealth();
    const globalStats = await this.collector.redisCluster.hgetall('global:stats:realtime');
    
    let metrics = [];

    // Collector metrics
    metrics.push('# HELP agent_collector_total_metrics Total metrics collected');
    metrics.push('# TYPE agent_collector_total_metrics counter');
    metrics.push(`agent_collector_total_metrics ${health.collector.totalMetrics}`);

    metrics.push('# HELP agent_collector_batches_processed Total batches processed');
    metrics.push('# TYPE agent_collector_batches_processed counter');
    metrics.push(`agent_collector_batches_processed ${health.collector.batchesProcessed}`);

    metrics.push('# HELP agent_collector_errors Total errors');
    metrics.push('# TYPE agent_collector_errors counter');
    metrics.push(`agent_collector_errors ${health.collector.errors}`);

    metrics.push('# HELP agent_collector_processing_time Average processing time in ms');
    metrics.push('# TYPE agent_collector_processing_time gauge');
    metrics.push(`agent_collector_processing_time ${health.collector.avgProcessingTime}`);

    metrics.push('# HELP agent_collector_buffer_size Current buffer size');
    metrics.push('# TYPE agent_collector_buffer_size gauge');
    metrics.push(`agent_collector_buffer_size ${health.bufferSize}`);

    // Redis cluster metrics
    metrics.push('# HELP redis_cluster_healthy Redis cluster health status');
    metrics.push('# TYPE redis_cluster_healthy gauge');
    metrics.push(`redis_cluster_healthy ${health.redis.healthy ? 1 : 0}`);

    metrics.push('# HELP redis_cluster_nodes Number of Redis nodes');
    metrics.push('# TYPE redis_cluster_nodes gauge');
    metrics.push(`redis_cluster_nodes ${health.redis.nodes}`);

    // Global agent metrics
    if (globalStats && Object.keys(globalStats).length > 0) {
      metrics.push('# HELP agent_tasks_total Total tasks processed');
      metrics.push('# TYPE agent_tasks_total counter');
      metrics.push(`agent_tasks_total ${globalStats.totalTasks || 0}`);

      metrics.push('# HELP agent_tasks_successful Total successful tasks');
      metrics.push('# TYPE agent_tasks_successful counter');
      metrics.push(`agent_tasks_successful ${globalStats.successfulTasks || 0}`);

      metrics.push('# HELP agent_success_rate Task success rate');
      metrics.push('# TYPE agent_success_rate gauge');
      metrics.push(`agent_success_rate ${globalStats.successRate || 0}`);

      metrics.push('# HELP agent_avg_execution_time Average execution time in ms');
      metrics.push('# TYPE agent_avg_execution_time gauge');
      metrics.push(`agent_avg_execution_time ${globalStats.avgDuration || 0}`);

      metrics.push('# HELP agent_avg_cpu_usage Average CPU usage percentage');
      metrics.push('# TYPE agent_avg_cpu_usage gauge');
      metrics.push(`agent_avg_cpu_usage ${globalStats.avgCpu || 0}`);

      metrics.push('# HELP agent_avg_memory_usage Average memory usage in MB');
      metrics.push('# TYPE agent_avg_memory_usage gauge');
      metrics.push(`agent_avg_memory_usage ${globalStats.avgMemory || 0}`);
    }

    // Get individual agent metrics for detailed monitoring
    try {
      const agentKeys = await this.collector.redisCluster.keys('agent:*:aggregated');
      for (const key of agentKeys.slice(0, 50)) { // Limit to 50 agents for performance
        const agentId = key.split(':')[1];
        const aggData = await this.collector.redisCluster.hgetall(key);
        
        if (aggData && Object.keys(aggData).length > 0) {
          const labels = `{agent_id="${agentId}"}`;
          
          metrics.push(`agent_tasks_total${labels} ${aggData.totalTasks || 0}`);
          metrics.push(`agent_tasks_successful${labels} ${aggData.successfulTasks || 0}`);
          metrics.push(`agent_success_rate${labels} ${aggData.successRate || 0}`);
          metrics.push(`agent_avg_execution_time${labels} ${aggData.avgDuration || 0}`);
          metrics.push(`agent_avg_cpu_usage${labels} ${aggData.avgCpu || 0}`);
          metrics.push(`agent_avg_memory_usage${labels} ${aggData.avgMemory || 0}`);
          metrics.push(`agent_error_count${labels} ${aggData.errorCount || 0}`);
        }
      }
    } catch (error) {
      console.error('Error fetching individual agent metrics:', error);
    }

    return metrics.join('\n');
  }

  async generateDashboardOverview() {
    const health = await this.collector.getSystemHealth();
    const globalStats = await this.collector.redisCluster.hgetall('global:stats:realtime');
    
    // Get recent swarm metrics
    const swarmKeys = await this.collector.redisCluster.keys('swarm:metrics:*');
    const latestSwarmKey = swarmKeys.sort().reverse()[0];
    const swarmMetrics = latestSwarmKey ? 
      await this.collector.redisCluster.hgetall(latestSwarmKey) : {};

    // Get active agents count
    const agentKeys = await this.collector.redisCluster.keys('agent:*:aggregated');
    const activeAgents = [];
    
    for (const key of agentKeys.slice(0, 20)) { // Limit for performance
      const agentId = key.split(':')[1];
      const aggData = await this.collector.redisCluster.hgetall(key);
      
      if (aggData && aggData.lastUpdate && 
          (Date.now() - parseInt(aggData.lastUpdate)) < 300000) { // Active in last 5 minutes
        activeAgents.push({
          agentId,
          totalTasks: parseInt(aggData.totalTasks) || 0,
          successRate: parseFloat(aggData.successRate) || 0,
          avgDuration: parseFloat(aggData.avgDuration) || 0,
          lastUpdate: parseInt(aggData.lastUpdate) || 0
        });
      }
    }

    return {
      timestamp: Date.now(),
      system: {
        health: health.redis.healthy,
        uptime: health.uptime,
        bufferSize: health.bufferSize
      },
      agents: {
        active: activeAgents.length,
        total: agentKeys.length,
        details: activeAgents
      },
      performance: {
        totalTasks: parseInt(globalStats.totalTasks) || 0,
        successfulTasks: parseInt(globalStats.successfulTasks) || 0,
        successRate: parseFloat(globalStats.successRate) || 0,
        avgDuration: parseFloat(globalStats.avgDuration) || 0,
        avgCpu: parseFloat(globalStats.avgCpu) || 0,
        avgMemory: parseFloat(globalStats.avgMemory) || 0
      },
      swarm: {
        topology: swarmMetrics.topology || 'unknown',
        utilization: parseFloat(swarmMetrics.utilization) || 0,
        throughput: parseFloat(swarmMetrics.throughput) || 0,
        queueLength: parseInt(swarmMetrics.queueLength) || 0
      },
      collector: health.collector
    };
  }

  start() {
    return new Promise((resolve) => {
      this.server = this.app.listen(this.port, () => {
        console.log(`🚀 Metrics Server running on port ${this.port}`);
        console.log(`📊 Prometheus metrics: http://localhost:${this.port}/metrics`);
        console.log(`🔍 Health check: http://localhost:${this.port}/health`);
        console.log(`📈 Dashboard API: http://localhost:${this.port}/api/dashboard/overview`);
        resolve();
      });
    });
  }

  stop() {
    return new Promise((resolve) => {
      if (this.server) {
        this.server.close(() => {
          console.log('✅ Metrics Server stopped');
          resolve();
        });
      } else {
        resolve();
      }
    });
  }
}

module.exports = MetricsServer;

// Start server if run directly
if (require.main === module) {
  const server = new MetricsServer(process.env.METRICS_PORT || 8080);
  server.start().catch(console.error);
  
  // Graceful shutdown
  process.on('SIGTERM', async () => {
    await server.stop();
    process.exit(0);
  });
  
  process.on('SIGINT', async () => {
    await server.stop();
    process.exit(0);
  });
}