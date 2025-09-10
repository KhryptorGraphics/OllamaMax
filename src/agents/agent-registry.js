#!/usr/bin/env node

/**
 * Agent Registry - Comprehensive agent definitions for Claude Code
 * 54 specialized agents with granular capabilities
 */

export const AGENT_REGISTRY = {
  // Core Development Agents
  'coder': {
    name: 'Coder',
    type: 'general-purpose',
    description: 'Implementation specialist for writing clean, efficient code',
    capabilities: ['code-generation', 'implementation', 'debugging', 'optimization'],
    tools: ['Write', 'Edit', 'MultiEdit', 'Read', 'Grep'],
    specializations: {
      languages: ['javascript', 'typescript', 'python', 'go', 'rust'],
      frameworks: ['react', 'vue', 'express', 'django', 'fastapi'],
      patterns: ['solid', 'dry', 'kiss', 'yagni', 'mvc', 'mvvm']
    },
    hooks: {
      pre: 'npx claude-flow@alpha hooks pre-task --agent coder',
      post: 'npx claude-flow@alpha hooks post-task --agent coder'
    }
  },
  
  'reviewer': {
    name: 'Reviewer',
    type: 'general-purpose',
    description: 'Code review and quality assurance specialist',
    capabilities: ['code-review', 'quality-assessment', 'security-audit', 'best-practices'],
    tools: ['Read', 'Grep', 'Glob', 'TodoWrite'],
    specializations: {
      focus: ['security', 'performance', 'maintainability', 'scalability'],
      standards: ['eslint', 'prettier', 'sonarqube', 'owasp'],
      metrics: ['complexity', 'coverage', 'debt', 'vulnerabilities']
    }
  },
  
  'tester': {
    name: 'Tester',
    type: 'general-purpose',
    description: 'Comprehensive testing and quality assurance specialist',
    capabilities: ['unit-testing', 'integration-testing', 'e2e-testing', 'tdd'],
    tools: ['Write', 'Edit', 'Bash', 'Read'],
    specializations: {
      frameworks: ['jest', 'mocha', 'playwright', 'cypress', 'selenium'],
      methodologies: ['tdd', 'bdd', 'atdd', 'property-based'],
      coverage: ['statement', 'branch', 'function', 'line']
    }
  },
  
  'planner': {
    name: 'Planner',
    type: 'general-purpose',
    description: 'Strategic planning and task orchestration agent',
    capabilities: ['requirement-analysis', 'task-breakdown', 'estimation', 'roadmapping'],
    tools: ['TodoWrite', 'Read', 'Write', 'Task'],
    specializations: {
      methodologies: ['agile', 'scrum', 'kanban', 'waterfall'],
      planning: ['sprint', 'epic', 'story', 'task'],
      metrics: ['velocity', 'burndown', 'throughput', 'cycle-time']
    }
  },
  
  'researcher': {
    name: 'Researcher',
    type: 'general-purpose',
    description: 'Deep research and information gathering specialist',
    capabilities: ['documentation-research', 'api-exploration', 'pattern-analysis', 'benchmarking'],
    tools: ['WebSearch', 'WebFetch', 'Read', 'Grep'],
    specializations: {
      domains: ['technical', 'business', 'competitive', 'market'],
      sources: ['documentation', 'papers', 'blogs', 'forums'],
      analysis: ['comparative', 'swot', 'feasibility', 'risk']
    }
  },

  // Swarm Coordination Agents
  'hierarchical-coordinator': {
    name: 'Hierarchical Coordinator',
    type: 'general-purpose',
    description: 'Queen-led hierarchical swarm coordination with specialized worker delegation',
    capabilities: ['hierarchy-management', 'task-delegation', 'worker-coordination', 'resource-allocation'],
    tools: ['Task', 'TodoWrite', 'mcp__flow-nexus__swarm_init', 'mcp__flow-nexus__agent_spawn'],
    specializations: {
      topology: 'hierarchical',
      roles: ['queen', 'supervisor', 'worker', 'scout'],
      delegation: ['top-down', 'priority-based', 'load-balanced']
    },
    config: {
      maxDepth: 5,
      maxWorkers: 20,
      delegationStrategy: 'intelligent'
    }
  },
  
  'mesh-coordinator': {
    name: 'Mesh Coordinator',
    type: 'general-purpose',
    description: 'Peer-to-peer mesh network swarm with distributed decision making',
    capabilities: ['peer-coordination', 'consensus-building', 'distributed-execution', 'fault-tolerance'],
    tools: ['Task', 'mcp__flow-nexus__swarm_init', 'mcp__flow-nexus__task_orchestrate'],
    specializations: {
      topology: 'mesh',
      consensus: ['voting', 'quorum', 'byzantine'],
      routing: ['shortest-path', 'load-balanced', 'redundant']
    }
  },
  
  'adaptive-coordinator': {
    name: 'Adaptive Coordinator',
    type: 'general-purpose',
    description: 'Dynamic topology switching coordinator with self-organizing patterns',
    capabilities: ['topology-switching', 'pattern-recognition', 'self-organization', 'optimization'],
    tools: ['Task', 'mcp__flow-nexus__swarm_scale', 'mcp__flow-nexus__swarm_monitor'],
    specializations: {
      patterns: ['emergent', 'swarm', 'flock', 'hive'],
      adaptation: ['load-based', 'performance-based', 'failure-based'],
      optimization: ['genetic', 'particle-swarm', 'ant-colony']
    }
  },

  // Consensus & Distributed Agents
  'byzantine-coordinator': {
    name: 'Byzantine Coordinator',
    type: 'general-purpose',
    description: 'Byzantine fault-tolerant consensus with malicious actor detection',
    capabilities: ['byzantine-consensus', 'fault-detection', 'malicious-detection', 'recovery'],
    tools: ['Task', 'mcp__flow-nexus__daa_agent_create', 'Bash'],
    specializations: {
      algorithms: ['pbft', 'tendermint', 'hotstuff'],
      tolerance: 'f < n/3',
      detection: ['behavior-analysis', 'voting-patterns', 'message-validation']
    }
  },
  
  'raft-manager': {
    name: 'Raft Manager',
    type: 'general-purpose',
    description: 'Raft consensus algorithm with leader election and log replication',
    capabilities: ['leader-election', 'log-replication', 'membership-changes', 'snapshots'],
    tools: ['Task', 'Write', 'Bash'],
    specializations: {
      states: ['leader', 'candidate', 'follower'],
      operations: ['append-entries', 'request-vote', 'install-snapshot'],
      guarantees: ['linearizability', 'durability', 'availability']
    }
  },
  
  'gossip-coordinator': {
    name: 'Gossip Coordinator',
    type: 'general-purpose',
    description: 'Gossip-based consensus for scalable eventually consistent systems',
    capabilities: ['gossip-protocol', 'epidemic-spread', 'convergence-detection', 'anti-entropy'],
    tools: ['Task', 'Bash', 'Write'],
    specializations: {
      protocols: ['push', 'pull', 'push-pull'],
      selection: ['random', 'round-robin', 'weighted'],
      convergence: ['probabilistic', 'deterministic', 'hybrid']
    }
  },

  // Performance & Optimization Agents
  'perf-analyzer': {
    name: 'Performance Analyzer',
    type: 'general-purpose',
    description: 'Performance bottleneck analyzer for workflow optimization',
    capabilities: ['profiling', 'bottleneck-detection', 'metric-analysis', 'optimization-recommendations'],
    tools: ['Bash', 'Read', 'mcp__flow-nexus__benchmark_run'],
    specializations: {
      metrics: ['cpu', 'memory', 'io', 'network'],
      analysis: ['flame-graphs', 'call-trees', 'heap-dumps'],
      optimization: ['caching', 'parallelization', 'algorithm-selection']
    }
  },
  
  'performance-benchmarker': {
    name: 'Performance Benchmarker',
    type: 'general-purpose',
    description: 'Comprehensive performance benchmarking for distributed protocols',
    capabilities: ['benchmark-design', 'load-testing', 'stress-testing', 'comparison'],
    tools: ['Bash', 'Write', 'mcp__flow-nexus__neural_performance_benchmark'],
    specializations: {
      types: ['throughput', 'latency', 'scalability', 'resilience'],
      tools: ['ab', 'jmeter', 'gatling', 'locust'],
      reporting: ['graphs', 'percentiles', 'trends', 'comparisons']
    }
  },

  // GitHub & Repository Agents
  'github-modes': {
    name: 'GitHub Modes',
    type: 'general-purpose',
    description: 'GitHub integration modes for workflow orchestration',
    capabilities: ['pr-management', 'issue-tracking', 'workflow-automation', 'release-coordination'],
    tools: ['Bash', 'Write', 'mcp__flow-nexus__github_repo_analyze'],
    specializations: {
      workflows: ['ci/cd', 'testing', 'deployment', 'release'],
      integrations: ['actions', 'webhooks', 'apps', 'oauth'],
      automation: ['auto-merge', 'auto-label', 'auto-assign']
    }
  },
  
  'pr-manager': {
    name: 'PR Manager',
    type: 'general-purpose',
    description: 'Comprehensive pull request management with swarm coordination',
    capabilities: ['pr-creation', 'review-coordination', 'merge-management', 'conflict-resolution'],
    tools: ['Bash', 'Git', 'Write', 'Task'],
    specializations: {
      review: ['code-quality', 'security', 'performance', 'style'],
      merge: ['squash', 'rebase', 'merge-commit'],
      automation: ['auto-review', 'auto-test', 'auto-merge']
    }
  },
  
  'code-review-swarm': {
    name: 'Code Review Swarm',
    type: 'general-purpose',
    description: 'AI agents for comprehensive intelligent code reviews',
    capabilities: ['multi-agent-review', 'pattern-detection', 'vulnerability-scanning', 'improvement-suggestions'],
    tools: ['Read', 'Grep', 'Task', 'TodoWrite'],
    specializations: {
      agents: ['security', 'performance', 'style', 'architecture'],
      depth: ['surface', 'semantic', 'behavioral', 'architectural'],
      feedback: ['inline', 'summary', 'actionable', 'educational']
    }
  },
  
  'issue-tracker': {
    name: 'Issue Tracker',
    type: 'general-purpose',
    description: 'Intelligent issue management and project coordination',
    capabilities: ['issue-triage', 'priority-assignment', 'duplicate-detection', 'progress-tracking'],
    tools: ['Bash', 'TodoWrite', 'Task', 'Write'],
    specializations: {
      classification: ['bug', 'feature', 'enhancement', 'task'],
      priority: ['critical', 'high', 'medium', 'low'],
      lifecycle: ['open', 'in-progress', 'review', 'closed']
    }
  },
  
  'release-manager': {
    name: 'Release Manager',
    type: 'general-purpose',
    description: 'Automated release coordination with ruv-swarm orchestration',
    capabilities: ['version-management', 'changelog-generation', 'deployment-coordination', 'rollback'],
    tools: ['Bash', 'Git', 'Write', 'Task'],
    specializations: {
      versioning: ['semantic', 'calendar', 'custom'],
      deployment: ['blue-green', 'canary', 'rolling', 'immediate'],
      automation: ['ci/cd', 'testing', 'notification', 'documentation']
    }
  },

  // SPARC Methodology Agents
  'sparc-coord': {
    name: 'SPARC Coordinator',
    type: 'general-purpose',
    description: 'SPARC methodology orchestrator for systematic development',
    capabilities: ['phase-coordination', 'methodology-enforcement', 'quality-gates', 'iteration'],
    tools: ['Task', 'TodoWrite', 'Write', 'Read'],
    specializations: {
      phases: ['specification', 'pseudocode', 'architecture', 'refinement', 'completion'],
      gates: ['review', 'approval', 'testing', 'validation'],
      iteration: ['waterfall', 'agile', 'spiral', 'incremental']
    }
  },
  
  'sparc-coder': {
    name: 'SPARC Coder',
    type: 'general-purpose',
    description: 'SPARC-specific implementation specialist',
    capabilities: ['sparc-implementation', 'phase-transition', 'code-generation', 'refinement'],
    tools: ['Write', 'Edit', 'MultiEdit', 'Read'],
    specializations: {
      patterns: ['specification-driven', 'test-first', 'iterative-refinement'],
      quality: ['correctness', 'completeness', 'consistency', 'clarity'],
      validation: ['unit-tests', 'integration-tests', 'acceptance-tests']
    }
  },
  
  'specification': {
    name: 'Specification Specialist',
    type: 'general-purpose',
    description: 'SPARC Specification phase specialist for requirements',
    capabilities: ['requirement-gathering', 'specification-writing', 'validation', 'traceability'],
    tools: ['Write', 'Read', 'TodoWrite', 'WebSearch'],
    specializations: {
      formats: ['user-stories', 'use-cases', 'formal-specs', 'bdd'],
      validation: ['completeness', 'consistency', 'feasibility', 'testability'],
      documentation: ['functional', 'non-functional', 'constraints', 'assumptions']
    }
  },
  
  'pseudocode': {
    name: 'Pseudocode Designer',
    type: 'general-purpose',
    description: 'SPARC Pseudocode phase specialist for algorithm design',
    capabilities: ['algorithm-design', 'pseudocode-writing', 'complexity-analysis', 'optimization'],
    tools: ['Write', 'Read', 'Edit'],
    specializations: {
      notation: ['structured', 'mathematical', 'flowchart', 'natural'],
      analysis: ['time-complexity', 'space-complexity', 'correctness'],
      optimization: ['algorithmic', 'data-structure', 'parallel']
    }
  },
  
  'architecture': {
    name: 'Architecture Specialist',
    type: 'general-purpose',
    description: 'SPARC Architecture phase specialist for system design',
    capabilities: ['system-design', 'component-design', 'interface-definition', 'pattern-application'],
    tools: ['Write', 'Read', 'Task', 'TodoWrite'],
    specializations: {
      patterns: ['mvc', 'microservices', 'event-driven', 'layered'],
      principles: ['solid', 'dry', 'kiss', 'yagni'],
      documentation: ['diagrams', 'adrs', 'interfaces', 'contracts']
    }
  },
  
  'refinement': {
    name: 'Refinement Specialist',
    type: 'general-purpose',
    description: 'SPARC Refinement phase specialist for iterative improvement',
    capabilities: ['code-refinement', 'optimization', 'refactoring', 'testing'],
    tools: ['Edit', 'MultiEdit', 'Read', 'Bash'],
    specializations: {
      techniques: ['refactoring', 'optimization', 'simplification', 'modularization'],
      metrics: ['complexity', 'coverage', 'performance', 'maintainability'],
      validation: ['regression-testing', 'performance-testing', 'acceptance-testing']
    }
  },

  // Specialized Development Agents
  'backend-dev': {
    name: 'Backend Developer',
    type: 'general-purpose',
    description: 'Specialized agent for backend API development',
    capabilities: ['api-design', 'database-design', 'authentication', 'optimization'],
    tools: ['Write', 'Edit', 'Bash', 'Read'],
    specializations: {
      frameworks: ['express', 'fastapi', 'django', 'gin', 'actix'],
      databases: ['postgresql', 'mongodb', 'redis', 'elasticsearch'],
      patterns: ['rest', 'graphql', 'grpc', 'websockets']
    }
  },
  
  'mobile-dev': {
    name: 'Mobile Developer',
    type: 'general-purpose',
    description: 'React Native mobile application development expert',
    capabilities: ['mobile-ui', 'cross-platform', 'native-integration', 'performance'],
    tools: ['Write', 'Edit', 'Bash', 'Read'],
    specializations: {
      platforms: ['ios', 'android', 'web'],
      frameworks: ['react-native', 'expo', 'flutter'],
      features: ['navigation', 'state', 'storage', 'networking']
    }
  },
  
  'ml-developer': {
    name: 'ML Developer',
    type: 'general-purpose',
    description: 'Machine learning model development and deployment',
    capabilities: ['model-training', 'feature-engineering', 'deployment', 'monitoring'],
    tools: ['Write', 'Bash', 'Read', 'mcp__flow-nexus__neural_train'],
    specializations: {
      frameworks: ['tensorflow', 'pytorch', 'scikit-learn', 'xgboost'],
      tasks: ['classification', 'regression', 'clustering', 'nlp', 'cv'],
      deployment: ['api', 'edge', 'batch', 'streaming']
    }
  },
  
  'cicd-engineer': {
    name: 'CI/CD Engineer',
    type: 'general-purpose',
    description: 'GitHub Actions CI/CD pipeline creation and optimization',
    capabilities: ['pipeline-design', 'automation', 'deployment', 'monitoring'],
    tools: ['Write', 'Bash', 'Git', 'Read'],
    specializations: {
      platforms: ['github-actions', 'jenkins', 'gitlab-ci', 'circleci'],
      stages: ['build', 'test', 'deploy', 'monitor'],
      strategies: ['blue-green', 'canary', 'rolling', 'feature-flags']
    }
  },
  
  'api-docs': {
    name: 'API Documentation Specialist',
    type: 'general-purpose',
    description: 'OpenAPI/Swagger documentation expert',
    capabilities: ['api-documentation', 'schema-generation', 'example-creation', 'validation'],
    tools: ['Write', 'Read', 'Edit'],
    specializations: {
      formats: ['openapi', 'swagger', 'raml', 'api-blueprint'],
      tools: ['swagger-ui', 'redoc', 'postman', 'insomnia'],
      documentation: ['endpoints', 'schemas', 'examples', 'authentication']
    }
  },
  
  'system-architect': {
    name: 'System Architect',
    type: 'general-purpose',
    description: 'System architecture design and technical decisions',
    capabilities: ['system-design', 'pattern-selection', 'scalability-planning', 'decision-making'],
    tools: ['Write', 'Read', 'TodoWrite', 'Task'],
    specializations: {
      patterns: ['microservices', 'serverless', 'event-driven', 'monolithic'],
      concerns: ['scalability', 'reliability', 'security', 'performance'],
      documentation: ['c4-model', 'uml', 'adrs', 'diagrams']
    }
  },
  
  'code-analyzer': {
    name: 'Code Analyzer',
    type: 'general-purpose',
    description: 'Advanced code quality analysis and improvements',
    capabilities: ['static-analysis', 'complexity-analysis', 'dependency-analysis', 'metrics'],
    tools: ['Read', 'Grep', 'Glob', 'Write'],
    specializations: {
      metrics: ['cyclomatic', 'cognitive', 'halstead', 'maintainability'],
      analysis: ['ast', 'data-flow', 'control-flow', 'dependency'],
      tools: ['eslint', 'sonarqube', 'codeclimate', 'deepscan']
    }
  },
  
  'base-template-generator': {
    name: 'Base Template Generator',
    type: 'general-purpose',
    description: 'Foundational templates and boilerplate code generator',
    capabilities: ['template-generation', 'boilerplate-creation', 'scaffolding', 'configuration'],
    tools: ['Write', 'Read', 'Edit', 'Bash'],
    specializations: {
      templates: ['component', 'service', 'api', 'test', 'config'],
      frameworks: ['react', 'vue', 'angular', 'express', 'django'],
      patterns: ['mvc', 'mvvm', 'repository', 'factory', 'singleton']
    }
  },

  // Testing & Validation Agents
  'tdd-london-swarm': {
    name: 'TDD London Swarm',
    type: 'general-purpose',
    description: 'Test-driven development with mock-driven design',
    capabilities: ['tdd', 'mocking', 'isolation-testing', 'outside-in'],
    tools: ['Write', 'Edit', 'Bash', 'Read'],
    specializations: {
      approach: ['outside-in', 'mockist', 'interaction-based'],
      tools: ['jest', 'sinon', 'mockito', 'wiremock'],
      patterns: ['aaa', 'given-when-then', 'test-doubles']
    }
  },
  
  'production-validator': {
    name: 'Production Validator',
    type: 'general-purpose',
    description: 'Production readiness validation specialist',
    capabilities: ['readiness-check', 'security-validation', 'performance-validation', 'compliance'],
    tools: ['Bash', 'Read', 'Grep', 'TodoWrite'],
    specializations: {
      checks: ['security', 'performance', 'scalability', 'reliability'],
      compliance: ['gdpr', 'hipaa', 'pci-dss', 'sox'],
      validation: ['smoke-tests', 'health-checks', 'monitoring', 'logging']
    }
  },

  // Migration & Infrastructure Agents
  'migration-planner': {
    name: 'Migration Planner',
    type: 'general-purpose',
    description: 'Migration planning for agent-based systems',
    capabilities: ['migration-planning', 'risk-assessment', 'rollback-planning', 'data-migration'],
    tools: ['Write', 'Read', 'TodoWrite', 'Task'],
    specializations: {
      types: ['database', 'platform', 'architecture', 'technology'],
      strategies: ['big-bang', 'phased', 'parallel-run', 'pilot'],
      validation: ['data-integrity', 'functionality', 'performance', 'rollback']
    }
  },
  
  'swarm-init': {
    name: 'Swarm Initializer',
    type: 'general-purpose',
    description: 'Swarm initialization and topology optimization',
    capabilities: ['swarm-setup', 'topology-design', 'agent-allocation', 'optimization'],
    tools: ['mcp__flow-nexus__swarm_init', 'Task', 'TodoWrite'],
    specializations: {
      topologies: ['mesh', 'hierarchical', 'star', 'ring'],
      optimization: ['load-balancing', 'latency', 'throughput', 'resilience'],
      scaling: ['horizontal', 'vertical', 'elastic', 'predictive']
    }
  },

  // Additional Specialized Agents
  'consensus-builder': {
    name: 'Consensus Builder',
    type: 'general-purpose',
    description: 'Multi-agent consensus coordination',
    capabilities: ['consensus-protocols', 'voting-mechanisms', 'conflict-resolution', 'agreement'],
    tools: ['Task', 'TodoWrite', 'mcp__flow-nexus__daa_agent_create'],
    specializations: {
      protocols: ['paxos', 'raft', 'pbft', 'tendermint'],
      voting: ['majority', 'quorum', 'weighted', 'ranked'],
      resolution: ['arbitration', 'mediation', 'voting', 'priority']
    }
  },
  
  'crdt-synchronizer': {
    name: 'CRDT Synchronizer',
    type: 'general-purpose',
    description: 'Conflict-free replicated data type synchronization',
    capabilities: ['crdt-implementation', 'merge-strategies', 'conflict-resolution', 'synchronization'],
    tools: ['Write', 'Edit', 'Task'],
    specializations: {
      types: ['g-counter', 'pn-counter', 'g-set', 'or-set', 'lww-register'],
      merge: ['commutative', 'associative', 'idempotent'],
      sync: ['state-based', 'operation-based', 'delta-based']
    }
  },
  
  'quorum-manager': {
    name: 'Quorum Manager',
    type: 'general-purpose',
    description: 'Dynamic quorum adjustment and membership management',
    capabilities: ['quorum-calculation', 'membership-management', 'split-brain-prevention', 'recovery'],
    tools: ['Task', 'Bash', 'mcp__flow-nexus__daa_agent_create'],
    specializations: {
      types: ['majority', 'weighted', 'hierarchical', 'dynamic'],
      membership: ['static', 'dynamic', 'elastic', 'federated'],
      recovery: ['auto-healing', 'manual', 'consensus-based']
    }
  },
  
  'security-manager': {
    name: 'Security Manager',
    type: 'general-purpose',
    description: 'Comprehensive security mechanisms for distributed protocols',
    capabilities: ['threat-modeling', 'encryption', 'authentication', 'authorization'],
    tools: ['Read', 'Grep', 'Write', 'Bash'],
    specializations: {
      threats: ['mitm', 'ddos', 'injection', 'replay'],
      crypto: ['tls', 'aes', 'rsa', 'ecdsa'],
      auth: ['jwt', 'oauth', 'saml', 'mtls']
    }
  },
  
  'task-orchestrator': {
    name: 'Task Orchestrator',
    type: 'general-purpose',
    description: 'Central coordination for task decomposition and synthesis',
    capabilities: ['task-breakdown', 'dependency-management', 'scheduling', 'result-aggregation'],
    tools: ['Task', 'TodoWrite', 'mcp__flow-nexus__task_orchestrate'],
    specializations: {
      decomposition: ['functional', 'data-parallel', 'pipeline', 'recursive'],
      scheduling: ['fifo', 'priority', 'deadline', 'fair-share'],
      aggregation: ['map-reduce', 'scatter-gather', 'fork-join']
    }
  },
  
  'memory-coordinator': {
    name: 'Memory Coordinator',
    type: 'general-purpose',
    description: 'Persistent memory and cross-agent memory sharing',
    capabilities: ['memory-management', 'state-persistence', 'cache-coordination', 'gc'],
    tools: ['mcp__flow-nexus__memory_usage', 'Write', 'Read'],
    specializations: {
      storage: ['in-memory', 'disk', 'distributed', 'hybrid'],
      sharing: ['broadcast', 'multicast', 'unicast', 'pubsub'],
      consistency: ['strong', 'eventual', 'causal', 'weak']
    }
  },
  
  'smart-agent': {
    name: 'Smart Agent',
    type: 'general-purpose',
    description: 'Intelligent agent coordination and dynamic spawning',
    capabilities: ['agent-selection', 'dynamic-spawning', 'load-prediction', 'auto-scaling'],
    tools: ['Task', 'mcp__flow-nexus__agent_spawn', 'mcp__flow-nexus__swarm_scale'],
    specializations: {
      selection: ['capability-based', 'load-based', 'cost-based', 'performance-based'],
      spawning: ['on-demand', 'predictive', 'scheduled', 'threshold-based'],
      scaling: ['reactive', 'proactive', 'predictive', 'scheduled']
    }
  },

  // GitHub Integration Agents
  'workflow-automation': {
    name: 'Workflow Automation',
    type: 'general-purpose',
    description: 'GitHub Actions workflow automation with adaptive coordination',
    capabilities: ['workflow-generation', 'pipeline-optimization', 'dependency-management', 'parallelization'],
    tools: ['Write', 'Bash', 'Git', 'Read'],
    specializations: {
      triggers: ['push', 'pull_request', 'schedule', 'webhook'],
      jobs: ['build', 'test', 'deploy', 'release'],
      optimization: ['caching', 'parallelization', 'conditionals', 'matrices']
    }
  },
  
  'project-board-sync': {
    name: 'Project Board Sync',
    type: 'general-purpose',
    description: 'Synchronize AI swarms with GitHub Projects',
    capabilities: ['board-sync', 'task-mapping', 'progress-tracking', 'automation'],
    tools: ['Bash', 'TodoWrite', 'Task'],
    specializations: {
      boards: ['kanban', 'scrum', 'custom'],
      sync: ['bidirectional', 'unidirectional', 'selective'],
      automation: ['rules', 'triggers', 'webhooks', 'actions']
    }
  },
  
  'release-swarm': {
    name: 'Release Swarm',
    type: 'general-purpose',
    description: 'Orchestrate complex releases with AI swarms',
    capabilities: ['release-orchestration', 'changelog-generation', 'multi-platform-deployment', 'validation'],
    tools: ['Bash', 'Git', 'Write', 'Task'],
    specializations: {
      platforms: ['github', 'npm', 'pypi', 'docker'],
      strategies: ['automated', 'manual', 'staged', 'continuous'],
      validation: ['tests', 'security', 'performance', 'compatibility']
    }
  },
  
  'repo-architect': {
    name: 'Repository Architect',
    type: 'general-purpose',
    description: 'Repository structure optimization and multi-repo management',
    capabilities: ['structure-design', 'monorepo-management', 'dependency-optimization', 'standardization'],
    tools: ['Write', 'Read', 'Git', 'Bash'],
    specializations: {
      structures: ['monorepo', 'polyrepo', 'hybrid'],
      tools: ['lerna', 'nx', 'rush', 'bazel'],
      optimization: ['build-times', 'dependencies', 'caching', 'sharing']
    }
  },
  
  'multi-repo-swarm': {
    name: 'Multi-Repo Swarm',
    type: 'general-purpose',
    description: 'Cross-repository swarm orchestration for organization-wide automation',
    capabilities: ['cross-repo-coordination', 'dependency-tracking', 'synchronized-updates', 'impact-analysis'],
    tools: ['Git', 'Bash', 'Task', 'TodoWrite'],
    specializations: {
      coordination: ['sequential', 'parallel', 'dependent', 'independent'],
      tracking: ['versions', 'dependencies', 'changes', 'impacts'],
      updates: ['cascade', 'selective', 'atomic', 'staged']
    }
  },
  
  'sync-coordinator': {
    name: 'Sync Coordinator',
    type: 'general-purpose',
    description: 'Multi-repository synchronization coordinator',
    capabilities: ['version-alignment', 'dependency-sync', 'cross-package-integration', 'conflict-resolution'],
    tools: ['Git', 'Bash', 'Write', 'Task'],
    specializations: {
      sync: ['versions', 'dependencies', 'configurations', 'schemas'],
      strategies: ['lock-step', 'independent', 'staged', 'continuous'],
      resolution: ['automatic', 'manual', 'policy-based', 'ai-assisted']
    }
  },
  
  'swarm-issue': {
    name: 'Swarm Issue Manager',
    type: 'general-purpose',
    description: 'GitHub issue-based swarm coordination',
    capabilities: ['issue-decomposition', 'task-generation', 'progress-tracking', 'auto-resolution'],
    tools: ['Bash', 'TodoWrite', 'Task', 'mcp__flow-nexus__task_orchestrate'],
    specializations: {
      decomposition: ['subtasks', 'dependencies', 'milestones', 'epics'],
      tracking: ['status', 'progress', 'blockers', 'dependencies'],
      resolution: ['auto-close', 'validation', 'verification', 'deployment']
    }
  },
  
  'swarm-pr': {
    name: 'Swarm PR Manager',
    type: 'general-purpose',
    description: 'Pull request swarm management with automated lifecycle',
    capabilities: ['pr-analysis', 'review-orchestration', 'conflict-resolution', 'auto-merge'],
    tools: ['Git', 'Bash', 'Task', 'Read'],
    specializations: {
      analysis: ['code-quality', 'security', 'performance', 'compatibility'],
      review: ['automated', 'multi-agent', 'progressive', 'gated'],
      merge: ['auto-merge', 'queue', 'train', 'stack']
    }
  },

  // Memory and Collective Intelligence
  'collective-intelligence-coordinator': {
    name: 'Collective Intelligence Coordinator',
    type: 'general-purpose',
    description: 'Orchestrates collective decision-making and knowledge synthesis',
    capabilities: ['knowledge-aggregation', 'decision-synthesis', 'pattern-emergence', 'wisdom-extraction'],
    tools: ['Task', 'mcp__flow-nexus__memory_usage', 'TodoWrite'],
    specializations: {
      aggregation: ['voting', 'weighted-consensus', 'bayesian', 'neural'],
      synthesis: ['merge', 'filter', 'transform', 'enhance'],
      patterns: ['emergent', 'learned', 'discovered', 'evolved']
    }
  },
  
  'swarm-memory-manager': {
    name: 'Swarm Memory Manager',
    type: 'general-purpose',
    description: 'Distributed memory management across agent swarms',
    capabilities: ['distributed-storage', 'memory-sharing', 'knowledge-persistence', 'context-management'],
    tools: ['mcp__flow-nexus__memory_usage', 'Write', 'Read', 'Task'],
    specializations: {
      storage: ['hierarchical', 'distributed', 'replicated', 'sharded'],
      sharing: ['broadcast', 'selective', 'on-demand', 'predictive'],
      persistence: ['session', 'permanent', 'versioned', 'temporal']
    }
  }
};

/**
 * Get agent configuration by type
 */
export function getAgent(type) {
  const agent = AGENT_REGISTRY[type];
  if (!agent) {
    // Fallback to general-purpose type
    return {
      name: type,
      type: 'general-purpose',
      description: `Custom agent: ${type}`,
      capabilities: ['general'],
      tools: ['Read', 'Write', 'Edit', 'Bash'],
      specializations: {}
    };
  }
  return agent;
}

/**
 * Get all available agent types
 */
export function getAvailableAgents() {
  return Object.keys(AGENT_REGISTRY);
}

/**
 * Get agents by capability
 */
export function getAgentsByCapability(capability) {
  return Object.entries(AGENT_REGISTRY)
    .filter(([_, agent]) => agent.capabilities.includes(capability))
    .map(([type, agent]) => ({ type, ...agent }));
}

/**
 * Get agents by tool requirement
 */
export function getAgentsByTool(tool) {
  return Object.entries(AGENT_REGISTRY)
    .filter(([_, agent]) => agent.tools.includes(tool))
    .map(([type, agent]) => ({ type, ...agent }));
}

/**
 * Map agent to Claude Code Task tool type
 */
export function mapToClaudeCodeType(agentType) {
  const agent = AGENT_REGISTRY[agentType];
  if (!agent) {
    return 'general-purpose';
  }
  
  // All agents map to general-purpose for Claude Code Task tool
  // but carry their specific configurations
  return agent.type || 'general-purpose';
}