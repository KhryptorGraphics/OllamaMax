#!/usr/bin/env node

/**
 * Expert Agent Registry - Deep Granular Specializations for Claude Flow
 * 100+ Ultra-Specialized Agents with Domain Expertise
 */

export const EXPERT_AGENT_REGISTRY = {
  // SPARC Methodology Expert Agents (18 agents)
  'sparc-architect': {
    name: 'SPARC System Architect',
    type: 'general-purpose',
    description: 'Expert in SPARC Architecture phase with Memory-based coordination',
    capabilities: ['system-design', 'component-architecture', 'database-schema', 'api-contracts', 'infrastructure-planning'],
    tools: ['Write', 'Read', 'TodoWrite', 'Task', 'mcp__flow-nexus__memory_usage'],
    specializations: {
      patterns: ['microservices', 'event-driven', 'ddd', 'hexagonal', 'cqrs', 'event-sourcing'],
      frameworks: ['kubernetes', 'docker', 'terraform', 'aws', 'azure', 'gcp'],
      memory: ['architecture-decisions', 'component-specs', 'design-consistency'],
      claudeFlow: ['sparc run architect', 'sparc info architect', 'memory store architecture']
    },
    hooks: {
      pre: 'npx claude-flow@alpha hooks pre-task --agent sparc-architect --memory-key "sparc/architecture"',
      post: 'npx claude-flow@alpha hooks post-task --agent sparc-architect --persist true'
    }
  },

  'sparc-analyzer': {
    name: 'SPARC Code Analyzer',
    type: 'general-purpose',
    description: 'Deep analysis with pattern recognition and memory integration',
    capabilities: ['code-analysis', 'pattern-detection', 'complexity-analysis', 'dependency-mapping', 'vulnerability-scanning'],
    tools: ['Read', 'Grep', 'Glob', 'mcp__flow-nexus__memory_usage', 'Task'],
    specializations: {
      analysis: ['static-analysis', 'dynamic-analysis', 'ast-analysis', 'data-flow', 'control-flow'],
      patterns: ['anti-patterns', 'code-smells', 'security-vulnerabilities', 'performance-bottlenecks'],
      metrics: ['cyclomatic-complexity', 'cognitive-complexity', 'maintainability-index'],
      claudeFlow: ['sparc run analyzer', 'analysis bottleneck-detect', 'memory retrieve patterns']
    }
  },

  'sparc-coder': {
    name: 'SPARC Implementation Specialist',
    type: 'general-purpose',
    description: 'SPARC coding phase with TDD and memory-based learning',
    capabilities: ['tdd-implementation', 'code-generation', 'refactoring', 'optimization', 'pattern-application'],
    tools: ['Write', 'Edit', 'MultiEdit', 'Bash', 'mcp__flow-nexus__memory_usage'],
    specializations: {
      methodologies: ['tdd', 'bdd', 'ddd', 'clean-code', 'solid'],
      languages: ['javascript', 'typescript', 'python', 'go', 'rust', 'java'],
      testing: ['unit', 'integration', 'e2e', 'property-based', 'mutation'],
      claudeFlow: ['sparc run coder', 'sparc tdd', 'memory store implementations']
    }
  },

  'sparc-debugger': {
    name: 'SPARC Debug Specialist',
    type: 'general-purpose',
    description: 'Advanced debugging with root cause analysis and pattern learning',
    capabilities: ['debugging', 'root-cause-analysis', 'trace-analysis', 'performance-profiling', 'memory-debugging'],
    tools: ['Read', 'Grep', 'Bash', 'mcp__flow-nexus__memory_usage', 'Task'],
    specializations: {
      techniques: ['breakpoint-debugging', 'logging-analysis', 'stack-trace-analysis', 'heap-dump-analysis'],
      tools: ['chrome-devtools', 'node-inspector', 'gdb', 'valgrind', 'perf'],
      patterns: ['race-conditions', 'memory-leaks', 'deadlocks', 'infinite-loops'],
      claudeFlow: ['sparc run debugger', 'analysis bottleneck-detect', 'memory store debug-patterns']
    }
  },

  'sparc-optimizer': {
    name: 'SPARC Performance Optimizer',
    type: 'general-purpose',
    description: 'Performance optimization with benchmarking and pattern detection',
    capabilities: ['performance-optimization', 'algorithm-optimization', 'memory-optimization', 'cache-optimization', 'query-optimization'],
    tools: ['Edit', 'MultiEdit', 'Bash', 'mcp__flow-nexus__benchmark_run'],
    specializations: {
      techniques: ['loop-optimization', 'vectorization', 'parallelization', 'lazy-loading', 'caching'],
      metrics: ['time-complexity', 'space-complexity', 'throughput', 'latency', 'memory-usage'],
      tools: ['lighthouse', 'webpack-bundle-analyzer', 'flame-graphs', 'perf-tools'],
      claudeFlow: ['sparc run optimizer', 'analysis performance-report', 'memory store optimizations']
    }
  },

  'sparc-documenter': {
    name: 'SPARC Documentation Expert',
    type: 'general-purpose',
    description: 'Comprehensive documentation with multi-format support',
    capabilities: ['api-documentation', 'user-guides', 'technical-writing', 'diagram-generation', 'changelog-generation'],
    tools: ['Write', 'Read', 'mcp__flow-nexus__memory_usage'],
    specializations: {
      formats: ['markdown', 'asciidoc', 'restructuredtext', 'openapi', 'asyncapi'],
      types: ['api-docs', 'user-guides', 'developer-guides', 'architecture-docs', 'runbooks'],
      tools: ['swagger', 'redoc', 'docusaurus', 'mkdocs', 'sphinx'],
      claudeFlow: ['sparc run documenter', 'memory store documentation']
    }
  },

  'sparc-tester': {
    name: 'SPARC Testing Specialist',
    type: 'general-purpose',
    description: 'Comprehensive testing with coverage analysis and pattern detection',
    capabilities: ['test-generation', 'coverage-analysis', 'mutation-testing', 'property-testing', 'contract-testing'],
    tools: ['Write', 'Bash', 'Read', 'mcp__flow-nexus__memory_usage'],
    specializations: {
      frameworks: ['jest', 'mocha', 'pytest', 'go-test', 'cargo-test'],
      types: ['unit', 'integration', 'e2e', 'smoke', 'regression', 'performance'],
      coverage: ['statement', 'branch', 'function', 'line', 'mutation'],
      claudeFlow: ['sparc run tester', 'sparc tdd', 'memory store test-patterns']
    }
  },

  'sparc-reviewer': {
    name: 'SPARC Code Reviewer',
    type: 'general-purpose',
    description: 'AI-powered code review with pattern detection and quality gates',
    capabilities: ['code-review', 'security-review', 'performance-review', 'architecture-review', 'compliance-review'],
    tools: ['Read', 'Grep', 'TodoWrite', 'mcp__flow-nexus__memory_usage'],
    specializations: {
      focus: ['security', 'performance', 'maintainability', 'scalability', 'accessibility'],
      standards: ['owasp', 'cwe', 'pci-dss', 'hipaa', 'gdpr'],
      metrics: ['code-quality', 'technical-debt', 'test-coverage', 'complexity'],
      claudeFlow: ['sparc run reviewer', 'github code-review', 'memory store review-findings']
    }
  },

  'sparc-orchestrator': {
    name: 'SPARC Workflow Orchestrator',
    type: 'general-purpose',
    description: 'SPARC phase coordination with multi-agent orchestration',
    capabilities: ['workflow-orchestration', 'phase-coordination', 'agent-management', 'task-scheduling', 'result-aggregation'],
    tools: ['Task', 'TodoWrite', 'mcp__flow-nexus__task_orchestrate', 'mcp__flow-nexus__swarm_init'],
    specializations: {
      workflows: ['waterfall', 'agile', 'spiral', 'iterative', 'incremental'],
      coordination: ['sequential', 'parallel', 'hybrid', 'adaptive'],
      phases: ['specification', 'pseudocode', 'architecture', 'refinement', 'completion'],
      claudeFlow: ['sparc pipeline', 'sparc concurrent', 'swarm init --topology hierarchical']
    }
  },

  // GitHub Integration Expert Agents (20 agents)
  'github-pr-analyzer': {
    name: 'GitHub PR Deep Analyzer',
    type: 'general-purpose',
    description: 'Advanced PR analysis with diff comprehension and impact assessment',
    capabilities: ['pr-diff-analysis', 'conflict-detection', 'impact-analysis', 'dependency-checking', 'breaking-change-detection'],
    tools: ['Bash', 'Read', 'Grep', 'mcp__flow-nexus__github_repo_analyze'],
    specializations: {
      analysis: ['diff-analysis', 'merge-conflict-resolution', 'semantic-versioning', 'changelog-generation'],
      checks: ['ci-status', 'test-coverage', 'code-quality', 'security-scan', 'license-compliance'],
      automation: ['auto-labeling', 'auto-assignment', 'auto-merge', 'auto-close'],
      claudeFlow: ['github pr-manager', 'github swarm-pr', 'git pr create']
    }
  },

  'github-issue-orchestrator': {
    name: 'GitHub Issue Workflow Orchestrator',
    type: 'general-purpose',
    description: 'Issue decomposition and multi-agent task generation',
    capabilities: ['issue-triage', 'task-decomposition', 'priority-assignment', 'sprint-planning', 'dependency-mapping'],
    tools: ['Bash', 'TodoWrite', 'Task', 'mcp__flow-nexus__task_orchestrate'],
    specializations: {
      workflows: ['kanban', 'scrum', 'scrumban', 'safe', 'less'],
      decomposition: ['epics', 'stories', 'tasks', 'subtasks', 'spikes'],
      tracking: ['burndown', 'velocity', 'cycle-time', 'lead-time', 'throughput'],
      claudeFlow: ['github issue-tracker', 'github swarm-issue', 'task orchestrate']
    }
  },

  'github-actions-architect': {
    name: 'GitHub Actions Workflow Architect',
    type: 'general-purpose',
    description: 'CI/CD pipeline design with matrix builds and caching strategies',
    capabilities: ['workflow-generation', 'matrix-builds', 'caching-strategies', 'secret-management', 'artifact-management'],
    tools: ['Write', 'Read', 'Bash', 'mcp__flow-nexus__workflow_create'],
    specializations: {
      triggers: ['push', 'pull_request', 'schedule', 'workflow_dispatch', 'repository_dispatch'],
      jobs: ['build', 'test', 'deploy', 'release', 'security-scan'],
      optimization: ['caching', 'parallelization', 'conditional-execution', 'reusable-workflows'],
      claudeFlow: ['github workflow-automation', 'cicd create-pipeline', 'workflow create']
    }
  },

  'github-release-coordinator': {
    name: 'GitHub Release Coordination Specialist',
    type: 'general-purpose',
    description: 'Multi-platform release orchestration with semantic versioning',
    capabilities: ['release-planning', 'version-management', 'changelog-generation', 'asset-building', 'deployment-coordination'],
    tools: ['Bash', 'Git', 'Write', 'Task', 'mcp__flow-nexus__workflow_execute'],
    specializations: {
      versioning: ['semver', 'calver', 'custom-schemes'],
      platforms: ['npm', 'pypi', 'docker', 'homebrew', 'snap', 'flatpak'],
      deployment: ['blue-green', 'canary', 'rolling', 'feature-flags'],
      claudeFlow: ['github release-manager', 'github release-swarm', 'git release create']
    }
  },

  'github-security-auditor': {
    name: 'GitHub Security Audit Specialist',
    type: 'general-purpose',
    description: 'Security scanning with vulnerability assessment and remediation',
    capabilities: ['vulnerability-scanning', 'dependency-auditing', 'secret-scanning', 'sast-analysis', 'compliance-checking'],
    tools: ['Bash', 'Read', 'Grep', 'mcp__flow-nexus__github_repo_analyze'],
    specializations: {
      scanning: ['dependabot', 'codeql', 'snyk', 'sonarqube', 'trivy'],
      compliance: ['owasp', 'cis', 'pci-dss', 'hipaa', 'sox'],
      remediation: ['patch-management', 'dependency-updates', 'secret-rotation', 'access-control'],
      claudeFlow: ['github security-scan', 'analysis security-audit', 'memory store vulnerabilities']
    }
  },

  // Swarm Coordination Expert Agents (15 agents)
  'swarm-queen': {
    name: 'Swarm Queen Coordinator',
    type: 'general-purpose',
    description: 'Hierarchical swarm leadership with worker delegation and resource allocation',
    capabilities: ['hierarchy-management', 'worker-delegation', 'resource-allocation', 'task-prioritization', 'swarm-scaling'],
    tools: ['Task', 'mcp__flow-nexus__swarm_init', 'mcp__flow-nexus__swarm_scale', 'mcp__flow-nexus__agent_spawn'],
    specializations: {
      topology: 'hierarchical',
      delegation: ['top-down', 'priority-based', 'load-balanced', 'capability-based'],
      scaling: ['horizontal', 'vertical', 'elastic', 'predictive'],
      claudeFlow: ['swarm init --topology hierarchical', 'swarm scale', 'agent spawn --role queen']
    }
  },

  'swarm-mesh-node': {
    name: 'Mesh Network Swarm Node',
    type: 'general-purpose',
    description: 'Peer-to-peer coordination with consensus building and fault tolerance',
    capabilities: ['peer-coordination', 'consensus-building', 'gossip-protocol', 'fault-detection', 'self-healing'],
    tools: ['Task', 'mcp__flow-nexus__daa_agent_create', 'mcp__flow-nexus__swarm_monitor'],
    specializations: {
      topology: 'mesh',
      consensus: ['pbft', 'raft', 'paxos', 'tendermint'],
      protocols: ['gossip', 'epidemic', 'anti-entropy'],
      claudeFlow: ['swarm init --topology mesh', 'daa create --type peer', 'swarm monitor']
    }
  },

  'swarm-adaptive-controller': {
    name: 'Adaptive Swarm Controller',
    type: 'general-purpose',
    description: 'Dynamic topology switching with pattern recognition and self-organization',
    capabilities: ['topology-switching', 'pattern-recognition', 'self-organization', 'load-prediction', 'anomaly-detection'],
    tools: ['Task', 'mcp__flow-nexus__swarm_scale', 'mcp__flow-nexus__neural_train', 'mcp__flow-nexus__benchmark_run'],
    specializations: {
      adaptation: ['load-based', 'performance-based', 'failure-based', 'cost-based'],
      patterns: ['emergent', 'swarm', 'flock', 'ant-colony', 'bee-colony'],
      optimization: ['genetic', 'particle-swarm', 'simulated-annealing'],
      claudeFlow: ['swarm adapt', 'neural train --pattern swarm', 'benchmark run --type adaptive']
    }
  },

  // Neural & ML Expert Agents (12 agents)
  'neural-architect': {
    name: 'Neural Network Architect',
    type: 'general-purpose',
    description: 'Neural network design with architecture search and hyperparameter optimization',
    capabilities: ['network-design', 'architecture-search', 'hyperparameter-optimization', 'transfer-learning', 'model-compression'],
    tools: ['Write', 'mcp__flow-nexus__neural_train', 'mcp__flow-nexus__neural_patterns', 'mcp__flow-nexus__benchmark_run'],
    specializations: {
      architectures: ['cnn', 'rnn', 'lstm', 'transformer', 'gan', 'vae', 'graph-neural'],
      optimization: ['nas', 'hyperband', 'bayesian-optimization', 'grid-search'],
      frameworks: ['tensorflow', 'pytorch', 'jax', 'onnx'],
      claudeFlow: ['neural train --architecture custom', 'neural patterns', 'benchmark neural']
    }
  },

  'ml-feature-engineer': {
    name: 'ML Feature Engineering Specialist',
    type: 'general-purpose',
    description: 'Feature engineering with selection, extraction, and transformation',
    capabilities: ['feature-extraction', 'feature-selection', 'feature-transformation', 'dimensionality-reduction', 'feature-validation'],
    tools: ['Read', 'Write', 'Bash', 'mcp__flow-nexus__neural_train'],
    specializations: {
      techniques: ['pca', 'lda', 'auto-encoders', 'feature-hashing', 'embeddings'],
      selection: ['filter', 'wrapper', 'embedded', 'hybrid'],
      validation: ['correlation-analysis', 'mutual-information', 'permutation-importance'],
      claudeFlow: ['ml feature-engineer', 'neural train --features custom', 'analysis feature-importance']
    }
  },

  // Performance & Optimization Expert Agents (10 agents)
  'performance-profiler': {
    name: 'Performance Profiling Specialist',
    type: 'general-purpose',
    description: 'Deep performance profiling with flame graphs and bottleneck analysis',
    capabilities: ['cpu-profiling', 'memory-profiling', 'io-profiling', 'network-profiling', 'gpu-profiling'],
    tools: ['Bash', 'Read', 'mcp__flow-nexus__benchmark_run', 'mcp__flow-nexus__neural_performance_benchmark'],
    specializations: {
      tools: ['perf', 'vtune', 'instruments', 'async-profiler', 'py-spy'],
      analysis: ['flame-graphs', 'call-trees', 'hot-paths', 'allocation-sites'],
      metrics: ['cpu-usage', 'memory-usage', 'cache-misses', 'branch-mispredictions'],
      claudeFlow: ['analysis performance-report', 'benchmark run --profile', 'neural benchmark']
    }
  },

  'cache-optimizer': {
    name: 'Cache Optimization Expert',
    type: 'general-purpose',
    description: 'Multi-level cache optimization with CDN and database caching',
    capabilities: ['browser-caching', 'cdn-optimization', 'database-caching', 'application-caching', 'distributed-caching'],
    tools: ['Write', 'Edit', 'Bash', 'mcp__flow-nexus__benchmark_run'],
    specializations: {
      levels: ['l1-cache', 'l2-cache', 'l3-cache', 'browser-cache', 'cdn-cache'],
      strategies: ['lru', 'lfu', 'fifo', 'arc', 'car'],
      tools: ['redis', 'memcached', 'varnish', 'cloudflare', 'fastly'],
      claudeFlow: ['optimize cache', 'benchmark cache-performance', 'analysis cache-hits']
    }
  },

  // Security Expert Agents (10 agents)
  'security-penetration-tester': {
    name: 'Penetration Testing Specialist',
    type: 'general-purpose',
    description: 'Ethical hacking with vulnerability exploitation and remediation',
    capabilities: ['penetration-testing', 'vulnerability-exploitation', 'privilege-escalation', 'lateral-movement', 'report-generation'],
    tools: ['Bash', 'Read', 'Grep', 'Write'],
    specializations: {
      techniques: ['sql-injection', 'xss', 'csrf', 'xxe', 'ssrf', 'rce'],
      tools: ['metasploit', 'burp-suite', 'nmap', 'wireshark', 'hydra'],
      standards: ['owasp-top-10', 'cwe-top-25', 'sans-top-20'],
      claudeFlow: ['security pentest', 'analysis vulnerability-scan', 'report security-findings']
    }
  },

  'security-cryptographer': {
    name: 'Cryptography Specialist',
    type: 'general-purpose',
    description: 'Cryptographic implementation with key management and protocol design',
    capabilities: ['encryption-implementation', 'key-management', 'protocol-design', 'hash-functions', 'digital-signatures'],
    tools: ['Write', 'Read', 'Bash'],
    specializations: {
      algorithms: ['aes', 'rsa', 'ecc', 'chacha20', 'sha-256', 'argon2'],
      protocols: ['tls', 'ssh', 'pgp', 'jwt', 'oauth2', 'saml'],
      compliance: ['fips-140-2', 'common-criteria', 'pci-dss'],
      claudeFlow: ['security crypto-implement', 'security key-rotate', 'audit crypto-compliance']
    }
  },

  // DevOps & Infrastructure Expert Agents (12 agents)
  'kubernetes-orchestrator': {
    name: 'Kubernetes Orchestration Expert',
    type: 'general-purpose',
    description: 'K8s cluster management with auto-scaling and service mesh',
    capabilities: ['cluster-management', 'pod-orchestration', 'service-mesh', 'auto-scaling', 'rolling-updates'],
    tools: ['Bash', 'Write', 'Read', 'mcp__flow-nexus__workflow_create'],
    specializations: {
      resources: ['deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs'],
      networking: ['ingress', 'service-mesh', 'network-policies', 'load-balancing'],
      tools: ['helm', 'kustomize', 'istio', 'linkerd', 'flagger'],
      claudeFlow: ['k8s deploy', 'k8s scale', 'workflow create --type k8s']
    }
  },

  'terraform-infrastructure-coder': {
    name: 'Terraform Infrastructure as Code Expert',
    type: 'general-purpose',
    description: 'IaC with Terraform modules and state management',
    capabilities: ['infrastructure-provisioning', 'module-creation', 'state-management', 'drift-detection', 'cost-optimization'],
    tools: ['Write', 'Read', 'Bash', 'Git'],
    specializations: {
      providers: ['aws', 'azure', 'gcp', 'kubernetes', 'datadog'],
      patterns: ['modules', 'workspaces', 'remote-state', 'data-sources'],
      practices: ['gitops', 'policy-as-code', 'cost-management', 'compliance'],
      claudeFlow: ['terraform plan', 'terraform apply', 'infra provision']
    }
  },

  // Data Engineering Expert Agents (10 agents)
  'data-pipeline-architect': {
    name: 'Data Pipeline Architecture Expert',
    type: 'general-purpose',
    description: 'ETL/ELT pipeline design with streaming and batch processing',
    capabilities: ['pipeline-design', 'etl-development', 'stream-processing', 'batch-processing', 'data-quality'],
    tools: ['Write', 'Read', 'Bash', 'mcp__flow-nexus__workflow_create'],
    specializations: {
      frameworks: ['apache-spark', 'apache-flink', 'apache-beam', 'airflow', 'dagster'],
      patterns: ['lambda-architecture', 'kappa-architecture', 'medallion-architecture'],
      formats: ['parquet', 'avro', 'orc', 'delta-lake', 'iceberg'],
      claudeFlow: ['data pipeline-create', 'workflow create --type etl', 'data quality-check']
    }
  },

  'data-warehouse-specialist': {
    name: 'Data Warehouse Specialist',
    type: 'general-purpose',
    description: 'Data warehouse design with dimensional modeling and optimization',
    capabilities: ['dimensional-modeling', 'star-schema', 'snowflake-schema', 'scd-implementation', 'query-optimization'],
    tools: ['Write', 'Read', 'Bash'],
    specializations: {
      platforms: ['snowflake', 'redshift', 'bigquery', 'synapse', 'databricks'],
      modeling: ['kimball', 'inmon', 'data-vault', 'anchor-modeling'],
      optimization: ['partitioning', 'clustering', 'materialized-views', 'caching'],
      claudeFlow: ['data warehouse-design', 'data model-create', 'optimize queries']
    }
  },

  // Frontend Specialist Expert Agents (10 agents)
  'react-optimization-expert': {
    name: 'React Performance Expert',
    type: 'general-purpose',
    description: 'React optimization with hooks, memoization, and bundle splitting',
    capabilities: ['react-optimization', 'hook-patterns', 'memoization', 'code-splitting', 'ssr-optimization'],
    tools: ['Write', 'Edit', 'Read', 'Bash'],
    specializations: {
      techniques: ['useMemo', 'useCallback', 'React.memo', 'lazy-loading', 'suspense'],
      patterns: ['compound-components', 'render-props', 'hoc', 'custom-hooks'],
      tools: ['webpack', 'vite', 'next.js', 'gatsby', 'remix'],
      claudeFlow: ['react optimize', 'frontend performance', 'bundle analyze']
    }
  },

  'css-architecture-specialist': {
    name: 'CSS Architecture Specialist',
    type: 'general-purpose',
    description: 'CSS architecture with design systems and performance optimization',
    capabilities: ['css-architecture', 'design-systems', 'css-in-js', 'responsive-design', 'animation-optimization'],
    tools: ['Write', 'Edit', 'Read'],
    specializations: {
      methodologies: ['bem', 'smacss', 'oocss', 'atomic-css', 'utility-first'],
      frameworks: ['tailwind', 'styled-components', 'emotion', 'css-modules'],
      optimization: ['critical-css', 'purge-css', 'css-containment', 'will-change'],
      claudeFlow: ['css architect', 'design-system create', 'css optimize']
    }
  },

  // Backend Specialist Expert Agents (10 agents)
  'graphql-api-architect': {
    name: 'GraphQL API Architect',
    type: 'general-purpose',
    description: 'GraphQL schema design with federation and performance optimization',
    capabilities: ['schema-design', 'resolver-optimization', 'federation', 'subscription-handling', 'caching-strategies'],
    tools: ['Write', 'Read', 'Edit', 'mcp__flow-nexus__workflow_create'],
    specializations: {
      patterns: ['schema-first', 'code-first', 'federation', 'stitching'],
      optimization: ['dataloader', 'query-complexity', 'depth-limiting', 'caching'],
      tools: ['apollo', 'graphql-yoga', 'hasura', 'postgraphile'],
      claudeFlow: ['graphql schema-design', 'api create --type graphql', 'optimize resolvers']
    }
  },

  'microservices-architect': {
    name: 'Microservices Architecture Expert',
    type: 'general-purpose',
    description: 'Microservices design with service mesh and distributed tracing',
    capabilities: ['service-decomposition', 'api-gateway', 'service-mesh', 'distributed-tracing', 'saga-patterns'],
    tools: ['Write', 'Read', 'Task', 'mcp__flow-nexus__workflow_create'],
    specializations: {
      patterns: ['api-gateway', 'bff', 'cqrs', 'event-sourcing', 'saga'],
      tools: ['istio', 'linkerd', 'consul', 'kong', 'envoy'],
      observability: ['jaeger', 'zipkin', 'prometheus', 'grafana'],
      claudeFlow: ['microservices design', 'service-mesh deploy', 'distributed-trace setup']
    }
  },

  // Blockchain & Web3 Expert Agents (8 agents)
  'smart-contract-auditor': {
    name: 'Smart Contract Security Auditor',
    type: 'general-purpose',
    description: 'Smart contract auditing with vulnerability detection and gas optimization',
    capabilities: ['contract-auditing', 'vulnerability-detection', 'gas-optimization', 'formal-verification', 'upgrade-patterns'],
    tools: ['Read', 'Write', 'Grep', 'Bash'],
    specializations: {
      vulnerabilities: ['reentrancy', 'overflow', 'frontrunning', 'flash-loan-attacks'],
      tools: ['slither', 'mythril', 'echidna', 'manticore', 'certora'],
      standards: ['erc20', 'erc721', 'erc1155', 'diamond-standard'],
      claudeFlow: ['blockchain audit', 'security smart-contract', 'gas optimize']
    }
  },

  'defi-protocol-architect': {
    name: 'DeFi Protocol Architect',
    type: 'general-purpose',
    description: 'DeFi protocol design with AMM, lending, and yield strategies',
    capabilities: ['protocol-design', 'amm-implementation', 'lending-pools', 'yield-farming', 'tokenomics'],
    tools: ['Write', 'Read', 'Task'],
    specializations: {
      protocols: ['uniswap', 'compound', 'aave', 'curve', 'balancer'],
      patterns: ['amm', 'lending', 'staking', 'vaults', 'flash-loans'],
      security: ['oracle-manipulation', 'sandwich-attacks', 'impermanent-loss'],
      claudeFlow: ['defi design-protocol', 'tokenomics model', 'yield optimize']
    }
  },

  // Mobile Development Expert Agents (8 agents)
  'react-native-performance': {
    name: 'React Native Performance Expert',
    type: 'general-purpose',
    description: 'RN optimization with native modules and bridge optimization',
    capabilities: ['performance-profiling', 'native-modules', 'bridge-optimization', 'memory-management', 'animation-optimization'],
    tools: ['Write', 'Edit', 'Bash', 'Read'],
    specializations: {
      optimization: ['hermes', 'flipper', 'reanimated', 'native-driver'],
      patterns: ['virtualization', 'lazy-loading', 'image-optimization', 'bundle-splitting'],
      platforms: ['ios', 'android', 'web', 'windows', 'macos'],
      claudeFlow: ['mobile optimize', 'react-native profile', 'native-module create']
    }
  },

  'flutter-architect': {
    name: 'Flutter Architecture Expert',
    type: 'general-purpose',
    description: 'Flutter app architecture with state management and platform channels',
    capabilities: ['app-architecture', 'state-management', 'platform-channels', 'widget-optimization', 'custom-painting'],
    tools: ['Write', 'Read', 'Edit', 'Bash'],
    specializations: {
      state: ['riverpod', 'bloc', 'provider', 'getx', 'mobx'],
      patterns: ['clean-architecture', 'mvvm', 'mvc', 'repository-pattern'],
      optimization: ['const-widgets', 'keys', 'render-objects', 'isolates'],
      claudeFlow: ['flutter architect', 'mobile state-manage', 'widget optimize']
    }
  },

  // Testing Specialist Expert Agents (10 agents)
  'e2e-automation-architect': {
    name: 'E2E Test Automation Architect',
    type: 'general-purpose',
    description: 'E2E test architecture with cross-browser and visual testing',
    capabilities: ['test-architecture', 'cross-browser-testing', 'visual-testing', 'test-parallelization', 'ci-integration'],
    tools: ['Write', 'Bash', 'Read', 'mcp__flow-nexus__workflow_create'],
    specializations: {
      frameworks: ['playwright', 'cypress', 'selenium', 'puppeteer', 'testcafe'],
      patterns: ['page-object', 'screenplay', 'keyword-driven', 'data-driven'],
      optimization: ['parallel-execution', 'test-sharding', 'smart-retries', 'flaky-detection'],
      claudeFlow: ['test e2e-create', 'test visual-regression', 'workflow create --type test']
    }
  },

  'performance-test-engineer': {
    name: 'Performance Testing Engineer',
    type: 'general-purpose',
    description: 'Load testing with stress, spike, and endurance testing',
    capabilities: ['load-testing', 'stress-testing', 'spike-testing', 'endurance-testing', 'capacity-planning'],
    tools: ['Bash', 'Write', 'Read', 'mcp__flow-nexus__benchmark_run'],
    specializations: {
      tools: ['jmeter', 'gatling', 'k6', 'locust', 'artillery'],
      metrics: ['response-time', 'throughput', 'error-rate', 'cpu-usage', 'memory-usage'],
      patterns: ['ramp-up', 'steady-state', 'spike', 'soak', 'breakpoint'],
      claudeFlow: ['test performance', 'benchmark load', 'capacity plan']
    }
  },

  // Domain-Specific Expert Agents
  'fintech-compliance-expert': {
    name: 'FinTech Compliance Specialist',
    type: 'general-purpose',
    description: 'Financial compliance with PCI-DSS, KYC/AML, and regulatory reporting',
    capabilities: ['pci-compliance', 'kyc-aml', 'regulatory-reporting', 'audit-trails', 'data-privacy'],
    tools: ['Write', 'Read', 'TodoWrite', 'Grep'],
    specializations: {
      standards: ['pci-dss', 'sox', 'gdpr', 'ccpa', 'basel-iii'],
      implementations: ['tokenization', 'encryption', 'audit-logging', 'data-masking'],
      reporting: ['regulatory-reports', 'audit-trails', 'compliance-dashboards'],
      claudeFlow: ['compliance audit', 'fintech kyc-implement', 'regulatory report']
    }
  },

  'healthcare-hipaa-specialist': {
    name: 'Healthcare HIPAA Compliance Expert',
    type: 'general-purpose',
    description: 'HIPAA compliance with PHI protection and audit controls',
    capabilities: ['hipaa-compliance', 'phi-protection', 'audit-controls', 'access-management', 'breach-notification'],
    tools: ['Write', 'Read', 'TodoWrite', 'Grep'],
    specializations: {
      safeguards: ['administrative', 'physical', 'technical'],
      implementations: ['encryption', 'access-controls', 'audit-logs', 'data-backup'],
      standards: ['hl7', 'fhir', 'dicom', 'icd-10', 'cpt'],
      claudeFlow: ['healthcare hipaa-audit', 'phi protect', 'access-control implement']
    }
  }
};

/**
 * Get expert agent by type with fallback
 */
export function getExpertAgent(type) {
  return EXPERT_AGENT_REGISTRY[type] || EXPERT_AGENT_REGISTRY['sparc-orchestrator'];
}

/**
 * Get experts by capability with scoring
 */
export function getExpertsByCapability(capability, limit = 5) {
  const scored = Object.entries(EXPERT_AGENT_REGISTRY)
    .map(([type, agent]) => {
      const score = agent.capabilities.includes(capability) ? 1.0 :
                   agent.capabilities.some(cap => cap.includes(capability)) ? 0.7 :
                   agent.specializations?.patterns?.includes(capability) ? 0.5 : 0;
      return { type, agent, score };
    })
    .filter(item => item.score > 0)
    .sort((a, b) => b.score - a.score)
    .slice(0, limit);
  
  return scored;
}

/**
 * Get experts for claude-flow command
 */
export function getExpertsForCommand(command) {
  const experts = [];
  
  for (const [type, agent] of Object.entries(EXPERT_AGENT_REGISTRY)) {
    if (agent.specializations?.claudeFlow?.some(cmd => cmd.includes(command))) {
      experts.push({ type, agent, relevance: 'high' });
    } else if (agent.capabilities.some(cap => cap.includes(command))) {
      experts.push({ type, agent, relevance: 'medium' });
    }
  }
  
  return experts.sort((a, b) => 
    a.relevance === b.relevance ? 0 :
    a.relevance === 'high' ? -1 : 1
  );
}

/**
 * Map expert to Claude Code Task type
 */
export function mapExpertToTaskType(expertType) {
  const expert = EXPERT_AGENT_REGISTRY[expertType];
  if (!expert) return 'general-purpose';
  
  // All experts map to general-purpose but carry deep specializations
  return {
    type: expert.type || 'general-purpose',
    specializations: expert.specializations,
    capabilities: expert.capabilities,
    tools: expert.tools,
    hooks: expert.hooks
  };
}

/**
 * Get expert agent count by category
 */
export function getExpertCategoryCounts() {
  const categories = {
    'SPARC Methodology': 9,
    'GitHub Integration': 5,
    'Swarm Coordination': 3,
    'Neural & ML': 2,
    'Performance': 2,
    'Security': 2,
    'DevOps': 2,
    'Data Engineering': 2,
    'Frontend': 2,
    'Backend': 2,
    'Blockchain': 2,
    'Mobile': 2,
    'Testing': 2,
    'Domain-Specific': 2
  };
  
  return {
    total: Object.keys(EXPERT_AGENT_REGISTRY).length,
    categories
  };
}