# Agent System Usage Guide

## 🚀 Overview

The agent system provides 54 specialized agents that integrate seamlessly with Claude Code's Task tool. All agents are properly mapped to the `general-purpose` type while maintaining their unique capabilities and specializations.

## 📋 Available Agents (54 Total)

### Core Development (5)
- `coder` - Implementation specialist
- `reviewer` - Code review specialist  
- `tester` - Testing specialist
- `planner` - Strategic planning
- `researcher` - Information gathering

### Swarm Coordination (3)
- `hierarchical-coordinator` - Queen-led hierarchical swarm
- `mesh-coordinator` - Peer-to-peer mesh network
- `adaptive-coordinator` - Dynamic topology switching

### Consensus & Distributed (7)
- `byzantine-coordinator` - Byzantine fault tolerance
- `raft-manager` - Raft consensus algorithm
- `gossip-coordinator` - Gossip-based consensus
- `consensus-builder` - Multi-agent consensus
- `crdt-synchronizer` - Conflict-free replicated data
- `quorum-manager` - Dynamic quorum adjustment
- `security-manager` - Security mechanisms

### Performance & Optimization (5)
- `perf-analyzer` - Performance bottleneck analysis
- `performance-benchmarker` - Comprehensive benchmarking
- `task-orchestrator` - Task decomposition
- `memory-coordinator` - Memory management
- `smart-agent` - Intelligent agent coordination

### GitHub & Repository (14)
- `github-modes` - GitHub workflow orchestration
- `pr-manager` - Pull request management
- `code-review-swarm` - AI code review swarm
- `issue-tracker` - Issue management
- `release-manager` - Release coordination
- `workflow-automation` - GitHub Actions automation
- `project-board-sync` - Project board sync
- `release-swarm` - Release orchestration
- `repo-architect` - Repository optimization
- `multi-repo-swarm` - Cross-repository swarm
- `sync-coordinator` - Multi-repo synchronization
- `swarm-issue` - Issue-based coordination
- `swarm-pr` - PR swarm management

### SPARC Methodology (6)
- `sparc-coord` - SPARC orchestrator
- `sparc-coder` - SPARC implementation
- `specification` - Specification specialist
- `pseudocode` - Pseudocode designer
- `architecture` - Architecture specialist
- `refinement` - Refinement specialist

### Specialized Development (10)
- `backend-dev` - Backend API development
- `mobile-dev` - React Native development
- `ml-developer` - Machine learning development
- `cicd-engineer` - CI/CD pipeline creation
- `api-docs` - OpenAPI/Swagger documentation
- `system-architect` - System architecture design
- `code-analyzer` - Code quality analysis
- `base-template-generator` - Template generation
- `tdd-london-swarm` - TDD with mocks
- `production-validator` - Production readiness

### Migration & Infrastructure (2)
- `migration-planner` - Migration planning
- `swarm-init` - Swarm initialization

### Memory & Intelligence (2)
- `collective-intelligence-coordinator` - Collective decision-making
- `swarm-memory-manager` - Distributed memory management

## 🎯 Using Agents with Claude Code

### Method 1: Direct Task Tool Usage

```javascript
// In Claude Code, use the Task tool directly
Task("Backend API Development", "Build REST API with authentication using Express", "general-purpose")
Task("Security Audit", "Review code for security vulnerabilities", "general-purpose")
Task("Test Creation", "Create comprehensive test suite with 90% coverage", "general-purpose")
```

### Method 2: Using NPM Scripts

```bash
# List all available agents
npm run agents:list

# Spawn specific agent
npm run agents:spawn backend-dev "Build user authentication API"

# Orchestrate task with optimal agents
npm run agents:orchestrate "Build secure e-commerce platform"
```

### Method 3: Programmatic Usage

```javascript
import { createIntegration } from './src/agents/claude-flow-integration.js';

const integration = await createIntegration();
await integration.initialize();

// Spawn specific agent
const result = await integration.spawnAgent('backend-dev', 'Build REST API');

// Or orchestrate with multiple agents
const orchestration = await integration.orchestrateTask(
  'Build secure authentication system',
  { maxAgents: 5 }
);
```

## 📊 Agent Capability Matching

The system automatically analyzes tasks and selects optimal agents:

| Task Keywords | Selected Agents |
|---------------|-----------------|
| implement, build, create | coder, backend-dev |
| test, tdd, coverage | tester, tdd-london-swarm |
| security, auth, encrypt | security-manager, reviewer |
| performance, optimize | perf-analyzer, performance-benchmarker |
| design, architecture | system-architect, architecture |

## 🔧 Configuration

### Integration Options

```javascript
const options = {
  autoSwarm: true,        // Auto-initialize swarm coordination
  topology: 'mesh',       // Swarm topology (mesh, hierarchical, star, ring)
  maxAgents: 10,          // Maximum concurrent agents
  memorySharing: true,    // Enable cross-agent memory
  hooks: true             // Enable coordination hooks
};

const integration = createIntegration(options);
```

### Agent Task Configuration

Each agent receives:
- Task description with focus area
- Specialized capabilities list
- Available tools
- Coordination instructions
- Hook integration for pre/post task

## 🐝 Swarm Coordination

### Automatic Swarm Mode

When multiple agents are needed, the system automatically:
1. Initializes swarm with optimal topology
2. Spawns agents based on task requirements
3. Coordinates via memory sharing
4. Aggregates results

### Manual Swarm Control

```bash
# Initialize swarm
npx claude-flow@alpha swarm init --topology mesh --max-agents 10

# Spawn specific agents
npx claude-flow@alpha agent spawn --type backend-dev
npx claude-flow@alpha agent spawn --type tester
npx claude-flow@alpha agent spawn --type security-manager

# Orchestrate task
npx claude-flow@alpha task orchestrate --task "Build API" --strategy parallel
```

## 📝 Example Workflows

### 1. Building a REST API

```javascript
// Orchestrate with multiple specialized agents
await integration.orchestrateTask(
  "Build REST API with authentication, testing, and documentation",
  { maxAgents: 5 }
);

// This will spawn:
// - backend-dev: API implementation
// - security-manager: Authentication security
// - tester: Test suite creation
// - api-docs: OpenAPI documentation
// - production-validator: Production readiness
```

### 2. Code Review Swarm

```javascript
// Spawn code review swarm
await integration.spawnAgent('code-review-swarm', 
  'Review pull request #123 for security, performance, and quality'
);

// The swarm coordinates multiple review agents:
// - Security review
// - Performance analysis
// - Code quality check
// - Architecture review
```

### 3. SPARC Development

```javascript
// Use SPARC methodology agents
const sparcAgents = [
  { agentType: 'specification', task: 'Define requirements' },
  { agentType: 'pseudocode', task: 'Design algorithms' },
  { agentType: 'architecture', task: 'Design system architecture' },
  { agentType: 'sparc-coder', task: 'Implement solution' },
  { agentType: 'refinement', task: 'Refine and optimize' }
];

await integration.spawnAgents(sparcAgents);
```

## 🔍 Monitoring & Status

```javascript
// Get integration status
const status = await integration.getStatus();
console.log(`Active agents: ${status.activeAgents.length}`);
console.log(`Swarm ID: ${status.swarmId}`);

// Get bridge statistics
console.log(`Bridge status:`, status.bridgeStatus);
```

## ⚡ Performance Tips

1. **Batch Agent Spawning**: Spawn multiple agents in one call for parallel execution
2. **Memory Sharing**: Enable memory sharing for better coordination
3. **Topology Selection**: Choose appropriate swarm topology for your use case
4. **Agent Selection**: Let the system auto-select agents based on task analysis
5. **Hook Integration**: Use hooks for coordination and progress tracking

## 🐛 Troubleshooting

### Agent Not Found Error
If you see "Agent type 'X' not found", ensure:
- The agent type exists in the registry
- You're using the correct agent name from the list
- The integration is properly initialized

### Swarm Initialization Failed
If swarm fails to initialize:
- Check claude-flow is installed: `npm list -g claude-flow`
- Try without auto-swarm: `{ autoSwarm: false }`
- Verify MCP server connectivity

### Task Execution Timeout
For long-running tasks:
- Increase timeout in options: `{ timeout: 600000 }`
- Break down complex tasks into smaller subtasks
- Use orchestration for parallel execution

## 🚀 Advanced Usage

### Custom Agent Selection

```javascript
// Find agents by capability
const securityAgents = integration.findAgentsByCapability('security-audit');

// Custom task analysis
const requirements = integration.analyzeTaskRequirements(task);
const agents = integration.selectOptimalAgents(requirements, { maxAgents: 3 });
```

### Cross-Session Memory

```javascript
// Store in memory
await integration.storeInMemory('project/api/design', apiDesign);

// Retrieve from memory
const design = await integration.retrieveFromMemory('project/api/design');
```

### Dynamic Agent Spawning

```javascript
// Spawn agents based on complexity
const complexity = analyzeComplexity(task);
const agentCount = complexity > 0.8 ? 5 : 3;

await integration.orchestrateTask(task, { maxAgents: agentCount });
```

## 📚 Resources

- Agent Registry: `/src/agents/agent-registry.js`
- Integration Module: `/src/agents/claude-flow-integration.js`
- Bridge System: `/src/agents/claude-flow-bridge.js`
- Test Suite: `/tests/test-agent-system.js`

## 🎯 Best Practices

1. **Use Orchestration** for complex tasks requiring multiple specializations
2. **Enable Hooks** for better coordination and progress tracking
3. **Leverage Memory** for cross-agent knowledge sharing
4. **Monitor Status** to track agent performance and resource usage
5. **Test First** with smaller tasks before scaling to complex orchestrations

---

The agent system is fully integrated and ready for production use with Claude Code's Task tool!