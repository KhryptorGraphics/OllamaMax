/**
 * Agent Specializations Configuration
 * Defines the capabilities, priorities, and characteristics of each agent type
 */

const AgentSpecializations = {
  'general-purpose': {
    name: 'General Purpose Agent',
    description: 'Versatile agent for complex, multi-step tasks and adaptive problem-solving',
    priority: 7,
    complexityThreshold: 0.1,
    maxConcurrent: 5,
    tools: ['*'], // Has access to all tools
    strengths: [
      'Multi-domain analysis and synthesis',
      'Complex problem decomposition',
      'Cross-functional coordination',
      'Adaptive approach selection',
      'System integration perspective'
    ],
    specializedPrompts: {
      analysis: 'Provide comprehensive multi-domain analysis with system-wide perspective',
      implementation: 'Implement solution with consideration for all architectural layers',
      coordination: 'Coordinate between specialized domains and ensure integration'
    },
    learningPatterns: [
      'cross-domain-integration',
      'adaptive-problem-solving',
      'system-wide-optimization',
      'multi-tool-coordination'
    ]
  },

  'system-architect': {
    name: 'System Architect',
    description: 'High-level system design, scalability, and distributed architecture patterns',
    priority: 10,
    complexityThreshold: 0.8,
    maxConcurrent: 2,
    tools: ['Read', 'Write', 'Grep', 'Glob', 'Bash'],
    strengths: [
      'Distributed system design patterns',
      'Scalability architecture planning',
      'Technology stack evaluation',
      'System integration strategy',
      'Performance and reliability design'
    ],
    specializedPrompts: {
      analysis: 'Analyze system architecture requirements and design scalable, maintainable solutions',
      design: 'Create comprehensive system architecture with clear component boundaries and integration patterns',
      review: 'Review existing architecture for scalability, maintainability, and performance optimization opportunities'
    },
    learningPatterns: [
      'microservices-patterns',
      'event-driven-architecture',
      'distributed-data-management',
      'scalability-strategies',
      'system-integration-patterns'
    ]
  },

  'backend-architect': {
    name: 'Backend Architect',
    description: 'Server-side systems, APIs, databases, and backend service architecture',
    priority: 9,
    complexityThreshold: 0.6,
    maxConcurrent: 3,
    tools: ['Read', 'Write', 'Edit', 'MultiEdit', 'Bash', 'Grep'],
    strengths: [
      'Database design and optimization',
      'RESTful and GraphQL API design',
      'Microservices architecture',
      'Authentication and authorization',
      'Caching and performance optimization'
    ],
    specializedPrompts: {
      api: 'Design and implement robust, scalable APIs with proper error handling and documentation',
      database: 'Optimize database schema and queries for performance and maintainability',
      services: 'Architect backend services with clear separation of concerns and integration patterns'
    },
    learningPatterns: [
      'api-design-patterns',
      'database-optimization',
      'service-architecture',
      'authentication-strategies',
      'caching-patterns'
    ]
  },

  'frontend-architect': {
    name: 'Frontend Architect',
    description: 'User interfaces, user experience, and modern frontend architecture',
    priority: 8,
    complexityThreshold: 0.5,
    maxConcurrent: 3,
    tools: ['Read', 'Write', 'Edit', 'MultiEdit', 'Bash'],
    strengths: [
      'Modern UI framework expertise (React, Vue, Angular)',
      'Responsive and accessible design',
      'State management patterns',
      'Performance optimization',
      'Progressive web applications'
    ],
    specializedPrompts: {
      ui: 'Create modern, accessible user interfaces with optimal user experience',
      performance: 'Optimize frontend performance including bundle size, loading times, and runtime efficiency',
      architecture: 'Design scalable frontend architecture with clear component hierarchies and data flow'
    },
    learningPatterns: [
      'component-architecture',
      'state-management-patterns',
      'performance-optimization',
      'accessibility-patterns',
      'responsive-design-strategies'
    ]
  },

  'security-engineer': {
    name: 'Security Engineer',
    description: 'Security vulnerabilities, compliance, and defensive security measures',
    priority: 9,
    complexityThreshold: 0.7,
    maxConcurrent: 2,
    tools: ['Read', 'Grep', 'Glob', 'Bash', 'Write'],
    strengths: [
      'Security vulnerability assessment',
      'OWASP compliance and standards',
      'Authentication and authorization systems',
      'Data protection and encryption',
      'Security monitoring and incident response'
    ],
    specializedPrompts: {
      assessment: 'Conduct comprehensive security assessment identifying vulnerabilities and mitigation strategies',
      implementation: 'Implement security measures following industry best practices and compliance requirements',
      monitoring: 'Design security monitoring and alerting systems for threat detection and response'
    },
    learningPatterns: [
      'vulnerability-patterns',
      'authentication-security',
      'data-protection-strategies',
      'security-monitoring',
      'compliance-frameworks'
    ]
  },

  'performance-engineer': {
    name: 'Performance Engineer',
    description: 'System performance optimization, bottleneck analysis, and scalability',
    priority: 8,
    complexityThreshold: 0.6,
    maxConcurrent: 3,
    tools: ['Read', 'Grep', 'Glob', 'Bash', 'Write'],
    strengths: [
      'Performance profiling and analysis',
      'Bottleneck identification and resolution',
      'Load testing and capacity planning',
      'Caching strategies and optimization',
      'Resource utilization optimization'
    ],
    specializedPrompts: {
      analysis: 'Analyze system performance bottlenecks and identify optimization opportunities',
      optimization: 'Implement performance optimizations with measurable impact metrics',
      monitoring: 'Design performance monitoring systems with actionable alerting thresholds'
    },
    learningPatterns: [
      'performance-bottlenecks',
      'optimization-strategies',
      'caching-patterns',
      'resource-optimization',
      'monitoring-strategies'
    ]
  },

  'quality-engineer': {
    name: 'Quality Engineer',
    description: 'Testing strategies, quality assurance, and reliability engineering',
    priority: 8,
    complexityThreshold: 0.5,
    maxConcurrent: 4,
    tools: ['Read', 'Write', 'Bash', 'Grep'],
    strengths: [
      'Comprehensive testing strategy design',
      'Test automation frameworks',
      'Quality metrics and monitoring',
      'Edge case identification',
      'Continuous integration optimization'
    ],
    specializedPrompts: {
      strategy: 'Design comprehensive testing strategy covering unit, integration, and end-to-end testing',
      automation: 'Create robust test automation framework with efficient CI/CD integration',
      quality: 'Establish quality metrics and monitoring for continuous quality improvement'
    },
    learningPatterns: [
      'testing-strategies',
      'automation-patterns',
      'quality-metrics',
      'edge-case-patterns',
      'ci-cd-integration'
    ]
  },

  'devops-architect': {
    name: 'DevOps Architect',
    description: 'Infrastructure automation, deployment pipelines, and operational excellence',
    priority: 8,
    complexityThreshold: 0.6,
    maxConcurrent: 3,
    tools: ['Read', 'Write', 'Edit', 'Bash'],
    strengths: [
      'CI/CD pipeline design and optimization',
      'Infrastructure as Code implementation',
      'Container orchestration (Docker, Kubernetes)',
      'Monitoring and observability systems',
      'Automated deployment strategies'
    ],
    specializedPrompts: {
      pipeline: 'Design efficient CI/CD pipelines with proper testing, security, and deployment stages',
      infrastructure: 'Implement Infrastructure as Code with proper versioning and environment management',
      monitoring: 'Create comprehensive monitoring and alerting systems for operational visibility'
    },
    learningPatterns: [
      'ci-cd-optimization',
      'infrastructure-automation',
      'container-orchestration',
      'monitoring-patterns',
      'deployment-strategies'
    ]
  },

  'python-expert': {
    name: 'Python Expert',
    description: 'Advanced Python development, frameworks, and ecosystem expertise',
    priority: 7,
    complexityThreshold: 0.4,
    maxConcurrent: 4,
    tools: ['Read', 'Write', 'Edit', 'MultiEdit', 'Bash', 'Grep'],
    strengths: [
      'Advanced Python language features',
      'Framework-specific optimization (Django, Flask, FastAPI)',
      'Package management and distribution',
      'Python performance optimization',
      'Testing frameworks and best practices'
    ],
    specializedPrompts: {
      development: 'Implement Pythonic solutions following PEP standards and best practices',
      optimization: 'Optimize Python code for performance, memory usage, and maintainability',
      frameworks: 'Leverage Python frameworks effectively for scalable application development'
    },
    learningPatterns: [
      'pythonic-patterns',
      'framework-optimization',
      'performance-tuning',
      'testing-strategies',
      'package-management'
    ]
  },

  'refactoring-expert': {
    name: 'Refactoring Expert',
    description: 'Code quality improvement, technical debt reduction, and clean code principles',
    priority: 7,
    complexityThreshold: 0.5,
    maxConcurrent: 3,
    tools: ['Read', 'Edit', 'MultiEdit', 'Grep', 'Write', 'Bash'],
    strengths: [
      'Code smell identification and resolution',
      'Design pattern implementation',
      'SOLID principles application',
      'Legacy code modernization',
      'Technical debt assessment and reduction'
    ],
    specializedPrompts: {
      analysis: 'Identify code quality issues and create systematic improvement plan',
      refactoring: 'Refactor code following SOLID principles and design patterns',
      modernization: 'Modernize legacy code while maintaining functionality and improving maintainability'
    },
    learningPatterns: [
      'refactoring-strategies',
      'code-smell-patterns',
      'design-pattern-application',
      'modernization-approaches',
      'quality-metrics'
    ]
  },

  'requirements-analyst': {
    name: 'Requirements Analyst',
    description: 'Requirements engineering, stakeholder analysis, and specification development',
    priority: 8,
    complexityThreshold: 0.4,
    maxConcurrent: 3,
    tools: ['Read', 'Write', 'Edit', 'TodoWrite', 'Grep', 'Bash'],
    strengths: [
      'Stakeholder requirement elicitation',
      'Functional and non-functional specifications',
      'Use case and user story development',
      'Requirement traceability management',
      'Gap analysis and validation'
    ],
    specializedPrompts: {
      elicitation: 'Systematically gather and analyze stakeholder requirements',
      specification: 'Create clear, testable specifications with acceptance criteria',
      validation: 'Validate requirements against business objectives and technical constraints'
    },
    learningPatterns: [
      'requirement-elicitation',
      'specification-patterns',
      'validation-strategies',
      'stakeholder-analysis',
      'traceability-management'
    ]
  },

  'technical-writer': {
    name: 'Technical Writer',
    description: 'Technical documentation, API documentation, and knowledge management',
    priority: 6,
    complexityThreshold: 0.3,
    maxConcurrent: 4,
    tools: ['Read', 'Write', 'Edit', 'Bash'],
    strengths: [
      'Technical documentation creation',
      'API documentation and examples',
      'User guide and tutorial development',
      'Documentation architecture design',
      'Knowledge management systems'
    ],
    specializedPrompts: {
      documentation: 'Create comprehensive technical documentation with clear examples and usage instructions',
      api: 'Develop detailed API documentation with interactive examples and use cases',
      guides: 'Write user-friendly guides and tutorials for complex technical concepts'
    },
    learningPatterns: [
      'documentation-patterns',
      'api-documentation-standards',
      'tutorial-structures',
      'knowledge-organization',
      'user-experience-writing'
    ]
  }
};

/**
 * Agent Selection Algorithm
 * Determines optimal agent configuration based on task characteristics
 */
class AgentSelector {
  constructor() {
    this.specializations = AgentSpecializations;
    this.selectionHistory = new Map();
  }

  /**
   * Select optimal agents for a given task
   */
  selectAgents(taskAnalysis, constraints = {}) {
    const {
      complexity,
      requiredSpecializations,
      estimatedAgentCount,
      taskType,
      priority
    } = taskAnalysis;

    const selectedAgents = [];
    const availableSlots = constraints.maxAgents || estimatedAgentCount;

    // Always include general-purpose for coordination
    selectedAgents.push({
      specialization: 'general-purpose',
      priority: 10,
      role: 'coordinator'
    });

    // Add specifically required specializations
    requiredSpecializations.forEach(spec => {
      if (spec !== 'general-purpose' && this.specializations[spec]) {
        selectedAgents.push({
          specialization: spec,
          priority: this.specializations[spec].priority,
          role: 'specialist'
        });
      }
    });

    // Fill remaining slots with complementary agents
    const remainingSlots = availableSlots - selectedAgents.length;
    if (remainingSlots > 0) {
      const complementaryAgents = this.selectComplementaryAgents(
        taskAnalysis,
        selectedAgents.map(a => a.specialization),
        remainingSlots
      );
      selectedAgents.push(...complementaryAgents);
    }

    // Sort by priority and return
    return selectedAgents.sort((a, b) => b.priority - a.priority);
  }

  /**
   * Select complementary agents based on task characteristics
   */
  selectComplementaryAgents(taskAnalysis, existingSpecializations, slots) {
    const { complexity, taskType } = taskAnalysis;
    const complementary = [];

    // High complexity tasks benefit from additional specialized agents
    if (complexity > 0.7 && slots >= 2) {
      if (!existingSpecializations.includes('system-architect')) {
        complementary.push({
          specialization: 'system-architect',
          priority: 9,
          role: 'architect'
        });
      }
      if (!existingSpecializations.includes('security-engineer')) {
        complementary.push({
          specialization: 'security-engineer',
          priority: 8,
          role: 'security'
        });
      }
    }

    // Performance-critical tasks
    if (taskType?.includes('performance') || taskType?.includes('optimization')) {
      if (!existingSpecializations.includes('performance-engineer')) {
        complementary.push({
          specialization: 'performance-engineer',
          priority: 8,
          role: 'performance'
        });
      }
    }

    // Quality assurance for complex implementations
    if (complexity > 0.5 && !existingSpecializations.includes('quality-engineer')) {
      complementary.push({
        specialization: 'quality-engineer',
        priority: 7,
        role: 'quality'
      });
    }

    return complementary.slice(0, slots);
  }

  /**
   * Get agent configuration for specialization
   */
  getAgentConfig(specialization) {
    return this.specializations[specialization] || this.specializations['general-purpose'];
  }

  /**
   * Update selection history for learning
   */
  updateSelectionHistory(taskAnalysis, selectedAgents, results) {
    const historyKey = this.generateHistoryKey(taskAnalysis);
    const historyEntry = {
      timestamp: Date.now(),
      taskAnalysis,
      selectedAgents,
      results,
      success: results.every(r => r.success),
      totalExecutionTime: results.reduce((sum, r) => sum + r.executionTime, 0)
    };

    const existing = this.selectionHistory.get(historyKey) || [];
    existing.push(historyEntry);
    this.selectionHistory.set(historyKey, existing);
  }

  /**
   * Generate history key for similar tasks
   */
  generateHistoryKey(taskAnalysis) {
    const { complexity, taskType, requiredSpecializations } = taskAnalysis;
    const complexityBand = Math.floor(complexity * 10) / 10; // Round to 0.1
    const specKey = requiredSpecializations.sort().join('-');
    return `${complexityBand}-${taskType || 'general'}-${specKey}`;
  }

  /**
   * Get learning insights from selection history
   */
  getLearningInsights() {
    const insights = {
      mostSuccessfulCombinations: [],
      performancePatterns: new Map(),
      specializationEfficiency: new Map()
    };

    this.selectionHistory.forEach((history, key) => {
      const successRate = history.filter(h => h.success).length / history.length;
      const avgExecutionTime = history.reduce((sum, h) => sum + h.totalExecutionTime, 0) / history.length;

      insights.performancePatterns.set(key, {
        successRate,
        avgExecutionTime,
        frequency: history.length
      });

      // Analyze specialization efficiency
      history.forEach(entry => {
        entry.selectedAgents.forEach(agent => {
          const existing = insights.specializationEfficiency.get(agent.specialization) || {
            totalTasks: 0,
            successfulTasks: 0,
            avgExecutionTime: 0
          };

          existing.totalTasks++;
          if (entry.success) existing.successfulTasks++;
          existing.avgExecutionTime = (existing.avgExecutionTime + entry.totalExecutionTime) / 2;

          insights.specializationEfficiency.set(agent.specialization, existing);
        });
      });
    });

    return insights;
  }
}

module.exports = {
  AgentSpecializations,
  AgentSelector
};