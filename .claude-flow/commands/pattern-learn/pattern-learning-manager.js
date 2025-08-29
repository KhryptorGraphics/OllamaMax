#!/usr/bin/env node

/**
 * Pattern Learning Manager
 * Core engine for extracting and analyzing patterns from successful operations
 */

const { performance } = require('perf_hooks');
const fs = require('fs').promises;
const path = require('path');

class PatternLearningManager {
  constructor() {
    this.activeSessions = new Map();
    this.operationHistory = new Map();
    this.patternTemplates = {
      'coordination': {
        name: 'Coordination Pattern',
        category: 'swarm',
        indicators: ['agent-sync', 'message-flow', 'task-distribution']
      },
      'task-execution': {
        name: 'Task Execution Pattern',
        category: 'tasks',
        indicators: ['execution-order', 'resource-usage', 'completion-time']
      },
      'communication': {
        name: 'Communication Pattern',
        category: 'communication',
        indicators: ['message-routing', 'response-time', 'bandwidth-usage']
      },
      'resource-allocation': {
        name: 'Resource Allocation Pattern',
        category: 'resources',
        indicators: ['memory-usage', 'cpu-utilization', 'load-balancing']
      },
      'optimization': {
        name: 'Optimization Pattern',
        category: 'performance',
        indicators: ['bottleneck-resolution', 'efficiency-gains', 'throughput']
      }
    };
  }

  async createLearningSession(config) {
    const sessionId = `learning-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    const sessionConfig = {
      id: sessionId,
      source: config.source || 'all',
      threshold: config.threshold || 0.8,
      saveAs: config.saveAs,
      createdAt: Date.now(),
      status: 'initialized'
    };

    this.activeSessions.set(sessionId, sessionConfig);

    console.log(`✅ Learning session ${sessionId} created successfully`);
    console.log(`🎯 Focus: ${sessionConfig.source} operations`);
    console.log(`📊 Success threshold: ${(sessionConfig.threshold * 100).toFixed(1)}%`);

    return sessionId;
  }

  async executeLearning(sessionId, executionConfig) {
    const session = this.activeSessions.get(sessionId);
    if (!session) {
      throw new Error(`Learning session ${sessionId} not found`);
    }

    const startTime = performance.now();
    session.status = 'learning';

    console.log(`🧠 Analyzing operations for pattern extraction...`);

    // Collect operation data
    const operations = await this.collectOperationData(session.source, session.threshold);
    
    // Extract patterns
    const patterns = await this.extractPatterns(operations, session.source);
    
    // Validate and score patterns
    const validatedPatterns = await this.validatePatterns(patterns, operations);
    
    // Generate recommendations
    const recommendations = await this.generateRecommendations(validatedPatterns);

    const endTime = performance.now();
    const learningTime = Math.round(endTime - startTime);

    const result = {
      sessionId,
      source: session.source,
      threshold: session.threshold,
      summary: {
        operationsAnalyzed: operations.length,
        successfulOperations: operations.filter(op => op.success >= session.threshold).length,
        patternsDiscovered: validatedPatterns.length,
        highConfidencePatterns: validatedPatterns.filter(p => p.confidence > 0.9).length,
        learningTime
      },
      patterns: validatedPatterns,
      categories: this.categorizePatterns(validatedPatterns),
      recommendations,
      quality: this.assessPatternQuality(validatedPatterns),
      rawData: operations
    };

    session.status = 'completed';
    session.result = result;

    return result;
  }

  async collectOperationData(source, threshold) {
    console.log(`📊 Collecting operation data for ${source}...`);
    
    // Simulate operation data collection - in real implementation, this would query actual logs
    const operations = [];
    const operationCount = Math.floor(Math.random() * 100) + 50;

    for (let i = 0; i < operationCount; i++) {
      const operation = {
        id: `op-${i + 1}`,
        type: this.getRandomOperationType(source),
        timestamp: Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000, // Last 7 days
        success: Math.random(),
        duration: Math.random() * 5000 + 500, // 500ms to 5.5s
        agentsInvolved: Math.floor(Math.random() * 8) + 2,
        tasksCompleted: Math.floor(Math.random() * 20) + 1,
        resourceUsage: {
          memory: Math.random() * 100,
          cpu: Math.random() * 100,
          network: Math.random() * 100
        },
        metrics: {
          messageLatency: Math.random() * 1000 + 100,
          taskCompletionRate: Math.random(),
          coordinationEfficiency: Math.random(),
          errorRate: Math.random() * 0.1
        },
        context: {
          complexity: Math.random(),
          priority: ['low', 'medium', 'high', 'critical'][Math.floor(Math.random() * 4)],
          topology: ['mesh', 'hierarchical', 'star', 'ring'][Math.floor(Math.random() * 4)]
        }
      };

      operations.push(operation);
    }

    // Filter by threshold
    const successfulOps = operations.filter(op => op.success >= threshold);
    console.log(`✅ Found ${successfulOps.length}/${operations.length} operations above threshold`);

    return operations;
  }

  getRandomOperationType(source) {
    const types = {
      'all': ['coordination', 'task-execution', 'communication', 'optimization', 'analysis'],
      'swarm': ['coordination', 'topology-switch', 'load-balancing'],
      'agents': ['task-assignment', 'agent-communication', 'resource-allocation'],
      'tasks': ['task-execution', 'task-prioritization', 'task-optimization'],
      'communication': ['message-routing', 'protocol-optimization', 'bandwidth-management']
    };

    const sourceTypes = types[source] || types['all'];
    return sourceTypes[Math.floor(Math.random() * sourceTypes.length)];
  }

  async extractPatterns(operations, source) {
    console.log(`🔍 Extracting patterns from ${operations.length} operations...`);
    
    const patterns = [];
    const successfulOps = operations.filter(op => op.success > 0.7);

    // Extract coordination patterns
    if (source === 'all' || source === 'swarm') {
      const coordPatterns = this.extractCoordinationPatterns(successfulOps);
      patterns.push(...coordPatterns);
    }

    // Extract task execution patterns
    if (source === 'all' || source === 'tasks') {
      const taskPatterns = this.extractTaskPatterns(successfulOps);
      patterns.push(...taskPatterns);
    }

    // Extract communication patterns
    if (source === 'all' || source === 'communication') {
      const commPatterns = this.extractCommunicationPatterns(successfulOps);
      patterns.push(...commPatterns);
    }

    // Extract resource patterns
    if (source === 'all' || source === 'agents') {
      const resourcePatterns = this.extractResourcePatterns(successfulOps);
      patterns.push(...resourcePatterns);
    }

    console.log(`🧠 Extracted ${patterns.length} potential patterns`);
    return patterns;
  }

  extractCoordinationPatterns(operations) {
    const patterns = [];

    // High-efficiency coordination pattern
    const highEfficiencyOps = operations.filter(op => 
      op.metrics.coordinationEfficiency > 0.8 && op.agentsInvolved > 4
    );

    if (highEfficiencyOps.length > 1) {
      patterns.push({
        id: 'coord-high-efficiency',
        name: 'High-Efficiency Coordination',
        category: 'coordination',
        description: 'Coordination strategy that maintains high efficiency with multiple agents',
        conditions: {
          agentCount: { min: 4, max: 12 },
          topology: this.getMostCommon(highEfficiencyOps, 'context.topology'),
          complexity: { max: 0.7 }
        },
        outcomes: {
          coordinationEfficiency: this.getAverage(highEfficiencyOps, 'metrics.coordinationEfficiency'),
          taskCompletionRate: this.getAverage(highEfficiencyOps, 'metrics.taskCompletionRate'),
          duration: this.getAverage(highEfficiencyOps, 'duration')
        },
        sampleSize: highEfficiencyOps.length,
        rawConfidence: highEfficiencyOps.length / operations.length
      });
    }

    // Fast coordination pattern
    const fastCoordOps = operations.filter(op => 
      op.duration < 2000 && op.metrics.coordinationEfficiency > 0.6
    );

    if (fastCoordOps.length > 1) {
      patterns.push({
        id: 'coord-fast-execution',
        name: 'Fast Coordination Execution',
        category: 'coordination',
        description: 'Coordination approach optimized for speed while maintaining efficiency',
        conditions: {
          maxDuration: 2000,
          minEfficiency: 0.6,
          topology: this.getMostCommon(fastCoordOps, 'context.topology')
        },
        outcomes: {
          averageDuration: this.getAverage(fastCoordOps, 'duration'),
          coordinationEfficiency: this.getAverage(fastCoordOps, 'metrics.coordinationEfficiency')
        },
        sampleSize: fastCoordOps.length,
        rawConfidence: fastCoordOps.length / operations.length
      });
    }

    return patterns;
  }

  extractTaskPatterns(operations) {
    const patterns = [];

    // High task completion rate pattern
    const highCompletionOps = operations.filter(op => 
      op.metrics.taskCompletionRate > 0.9 && op.tasksCompleted > 5
    );

    if (highCompletionOps.length > 1) {
      patterns.push({
        id: 'task-high-completion',
        name: 'High Task Completion Rate',
        category: 'task-execution',
        description: 'Task execution strategy that achieves high completion rates',
        conditions: {
          minTasks: 5,
          priority: this.getMostCommon(highCompletionOps, 'context.priority'),
          agentRange: {
            min: Math.min(...highCompletionOps.map(op => op.agentsInvolved)),
            max: Math.max(...highCompletionOps.map(op => op.agentsInvolved))
          }
        },
        outcomes: {
          completionRate: this.getAverage(highCompletionOps, 'metrics.taskCompletionRate'),
          averageDuration: this.getAverage(highCompletionOps, 'duration'),
          errorRate: this.getAverage(highCompletionOps, 'metrics.errorRate')
        },
        sampleSize: highCompletionOps.length,
        rawConfidence: highCompletionOps.length / operations.length
      });
    }

    return patterns;
  }

  extractCommunicationPatterns(operations) {
    const patterns = [];

    // Low latency communication pattern
    const lowLatencyOps = operations.filter(op => 
      op.metrics.messageLatency < 300 && op.agentsInvolved > 3
    );

    if (lowLatencyOps.length > 2) {
      patterns.push({
        id: 'comm-low-latency',
        name: 'Low-Latency Communication',
        category: 'communication',
        description: 'Communication pattern that minimizes message latency',
        conditions: {
          maxLatency: 300,
          minAgents: 3,
          topology: this.getMostCommon(lowLatencyOps, 'context.topology')
        },
        outcomes: {
          averageLatency: this.getAverage(lowLatencyOps, 'metrics.messageLatency'),
          networkUsage: this.getAverage(lowLatencyOps, 'resourceUsage.network'),
          coordinationEfficiency: this.getAverage(lowLatencyOps, 'metrics.coordinationEfficiency')
        },
        sampleSize: lowLatencyOps.length,
        rawConfidence: lowLatencyOps.length / operations.length
      });
    }

    return patterns;
  }

  extractResourcePatterns(operations) {
    const patterns = [];

    // Efficient resource usage pattern
    const efficientResourceOps = operations.filter(op => 
      op.resourceUsage.memory < 60 && 
      op.resourceUsage.cpu < 70 && 
      op.success > 0.8
    );

    if (efficientResourceOps.length > 2) {
      patterns.push({
        id: 'resource-efficient',
        name: 'Efficient Resource Utilization',
        category: 'resource-allocation',
        description: 'Resource allocation strategy that maintains high success with low resource usage',
        conditions: {
          maxMemory: 60,
          maxCpu: 70,
          minSuccess: 0.8
        },
        outcomes: {
          memoryUsage: this.getAverage(efficientResourceOps, 'resourceUsage.memory'),
          cpuUsage: this.getAverage(efficientResourceOps, 'resourceUsage.cpu'),
          successRate: this.getAverage(efficientResourceOps, 'success'),
          duration: this.getAverage(efficientResourceOps, 'duration')
        },
        sampleSize: efficientResourceOps.length,
        rawConfidence: efficientResourceOps.length / operations.length
      });
    }

    return patterns;
  }

  async validatePatterns(patterns, operations) {
    console.log(`🔬 Validating ${patterns.length} patterns...`);
    
    const validatedPatterns = [];

    for (const pattern of patterns) {
      // Calculate confidence based on sample size and consistency
      const confidence = this.calculatePatternConfidence(pattern, operations);
      
      // Calculate success rate
      const successRate = this.calculateSuccessRate(pattern, operations);
      
      // Determine applicability
      const applicability = this.determineApplicability(pattern, operations);

      if (confidence > 0.3) { // Only include patterns with reasonable confidence
        validatedPatterns.push({
          ...pattern,
          confidence,
          successRate,
          applicability,
          details: {
            sampleSize: pattern.sampleSize,
            validationScore: confidence * successRate,
            implementationComplexity: this.assessImplementationComplexity(pattern)
          }
        });
      }
    }

    console.log(`✅ Validated ${validatedPatterns.length} patterns`);
    return validatedPatterns.sort((a, b) => b.confidence - a.confidence);
  }

  calculatePatternConfidence(pattern, operations) {
    const sampleRatio = pattern.sampleSize / operations.length;
    const baseConfidence = Math.min(sampleRatio * 5, 1.0); // Scale sample ratio
    
    // Adjust based on consistency
    const consistencyBonus = pattern.rawConfidence > 0.3 ? 0.2 : 0;
    
    return Math.min(baseConfidence + consistencyBonus, 1.0);
  }

  calculateSuccessRate(pattern, operations) {
    // Find operations that match pattern conditions
    const matchingOps = operations.filter(op => this.matchesPattern(op, pattern));
    const successfulMatches = matchingOps.filter(op => op.success > 0.7);
    
    return matchingOps.length > 0 ? successfulMatches.length / matchingOps.length : 0;
  }

  matchesPattern(operation, pattern) {
    // Simplified pattern matching - in real implementation, this would be more sophisticated
    if (pattern.conditions.agentCount) {
      const { min, max } = pattern.conditions.agentCount;
      if (operation.agentsInvolved < min || operation.agentsInvolved > max) {
        return false;
      }
    }
    
    if (pattern.conditions.topology && operation.context.topology !== pattern.conditions.topology) {
      return false;
    }
    
    return true;
  }

  determineApplicability(pattern, operations) {
    const categories = {
      'coordination': 'Multi-agent coordination scenarios',
      'task-execution': 'Complex task processing workflows',
      'communication': 'High-frequency agent communication',
      'resource-allocation': 'Resource-constrained environments'
    };
    
    return categories[pattern.category] || 'General swarm operations';
  }

  assessImplementationComplexity(pattern) {
    const complexityFactors = {
      'coordination': 'medium',
      'task-execution': 'low',
      'communication': 'high',
      'resource-allocation': 'medium'
    };
    
    return complexityFactors[pattern.category] || 'medium';
  }

  async generateRecommendations(patterns) {
    const recommendations = [];

    const highConfidencePatterns = patterns.filter(p => p.confidence > 0.8);
    
    if (highConfidencePatterns.length > 0) {
      recommendations.push({
        action: `Implement ${highConfidencePatterns.length} high-confidence patterns`,
        expectedImpact: 'Significant performance improvement',
        priority: 'high',
        patterns: highConfidencePatterns.map(p => p.id)
      });
    }

    const communicationPatterns = patterns.filter(p => p.category === 'communication');
    if (communicationPatterns.length > 0) {
      recommendations.push({
        action: 'Optimize communication protocols based on learned patterns',
        expectedImpact: 'Reduced latency and improved coordination',
        priority: 'medium',
        patterns: communicationPatterns.map(p => p.id)
      });
    }

    const resourcePatterns = patterns.filter(p => p.category === 'resource-allocation');
    if (resourcePatterns.length > 0) {
      recommendations.push({
        action: 'Apply resource optimization patterns',
        expectedImpact: 'Better resource utilization and cost reduction',
        priority: 'medium',
        patterns: resourcePatterns.map(p => p.id)
      });
    }

    return recommendations;
  }

  categorizePatterns(patterns) {
    const categories = {};
    
    patterns.forEach(pattern => {
      categories[pattern.category] = (categories[pattern.category] || 0) + 1;
    });
    
    return categories;
  }

  assessPatternQuality(patterns) {
    if (patterns.length === 0) {
      return {
        averageConfidence: 0,
        diversity: 0,
        validationScore: 0,
        applicabilityRange: 'None'
      };
    }

    const averageConfidence = patterns.reduce((sum, p) => sum + p.confidence, 0) / patterns.length;
    const uniqueCategories = new Set(patterns.map(p => p.category)).size;
    const diversity = uniqueCategories / Object.keys(this.patternTemplates).length;
    const validationScore = patterns.reduce((sum, p) => sum + (p.confidence * p.successRate), 0) / patterns.length;
    
    return {
      averageConfidence,
      diversity,
      validationScore,
      applicabilityRange: uniqueCategories > 2 ? 'Broad' : uniqueCategories > 1 ? 'Moderate' : 'Narrow'
    };
  }

  // Utility functions
  getMostCommon(operations, property) {
    const values = operations.map(op => this.getNestedProperty(op, property)).filter(v => v);
    const counts = {};
    values.forEach(v => counts[v] = (counts[v] || 0) + 1);
    return Object.keys(counts).reduce((a, b) => counts[a] > counts[b] ? a : b, null);
  }

  getAverage(operations, property) {
    const values = operations.map(op => this.getNestedProperty(op, property)).filter(v => typeof v === 'number');
    return values.length > 0 ? values.reduce((sum, v) => sum + v, 0) / values.length : 0;
  }

  getNestedProperty(obj, path) {
    return path.split('.').reduce((current, key) => current && current[key], obj);
  }
}

module.exports = PatternLearningManager;
