# SPARC Orchestrator Documentation

## Overview

The SPARC Orchestrator is a sophisticated multi-agent task coordination system that implements various orchestration patterns for efficient task decomposition, parallel execution, and result synthesis.

## Features

### Core Capabilities
- **Task Decomposition**: Breaks complex tasks into agent-specific subtasks
- **Agent Coordination**: Manages multiple specialized agents working together
- **Resource Allocation**: Efficiently distributes work across available agents
- **Progress Tracking**: Real-time monitoring of task and agent status
- **Result Synthesis**: Aggregates and synthesizes results from multiple agents
- **Memory Sharing**: Enables agents to share context and intermediate results

### Orchestration Patterns

#### 1. Hierarchical Coordination
- Master agent delegates to specialized workers
- Top-down communication flow
- Centralized control and review

```javascript
const results = await orchestrator.coordinateTask(task, {
  strategy: 'domain',
  pattern: 'hierarchical'
});
```

#### 2. Parallel Pipeline
- Multiple parallel execution streams
- Synchronization at checkpoints
- Efficient for independent tasks

```javascript
const results = await orchestrator.coordinateTask(task, {
  strategy: 'parallel',
  parallel: true
});
```

#### 3. Sequential Pipeline
- Step-by-step execution with dependencies
- Context passing between stages
- SPARC methodology implementation

```javascript
const results = await orchestrator.coordinateTask(task, {
  strategy: 'sequential',
  parallel: false
});
```

#### 4. Event-Driven
- Reactive agent coordination
- Event-based triggering
- Dynamic workflow adaptation

```javascript
const pattern = OrchestrationPatterns.eventDriven;
const results = await pattern.execute(task, orchestrator);
```

#### 5. Adaptive Strategy
- Dynamic strategy selection based on complexity
- Automatic agent allocation
- Self-optimizing execution

```javascript
const results = await orchestrator.coordinateTask(task, {
  strategy: 'adaptive'
});
```

#### 6. Consensus Building
- Multiple validators working together
- Majority voting mechanism
- Quality assurance through agreement

```javascript
const pattern = OrchestrationPatterns.consensus;
const results = await pattern.execute(task, orchestrator);
```

#### 7. MapReduce
- Distributed processing pattern
- Parallel map phase
- Aggregated reduce phase

```javascript
const pattern = OrchestrationPatterns.mapReduce;
const results = await pattern.execute(task, orchestrator);
```

## Usage

### Basic Setup

```javascript
import SPARCOrchestrator from './src/sparc-orchestrator.js';

const orchestrator = new SPARCOrchestrator();
await orchestrator.initialize();
```

### Simple Task Coordination

```javascript
// Coordinate a simple task
const results = await orchestrator.coordinateTask('Build user authentication system');

console.log(results.summary);
console.log(results.metrics);
```

### Advanced Configuration

```javascript
// Complex task with specific options
const results = await orchestrator.coordinateTask('Refactor entire codebase', {
  strategy: 'adaptive',      // Use adaptive strategy
  parallel: true,            // Enable parallel execution
  timeout: 600000,          // 10 minute timeout
  retryOnFailure: true      // Retry failed subtasks
});
```

### Pattern-Based Execution

```javascript
import { selectPattern } from './src/orchestration-patterns.js';

// Auto-select pattern based on task
const pattern = selectPattern(task);
const results = await pattern.execute(task, orchestrator);

// Or specify pattern explicitly
const pattern = selectPattern(task, { pattern: 'consensus' });
```

### Memory Sharing

```javascript
// Share data between agents
await orchestrator.shareMemory('context-key', contextData);

// Retrieve shared data
const sharedData = await orchestrator.getMemory('context-key');
```

### Progress Monitoring

```javascript
// Monitor ongoing tasks
const status = await orchestrator.monitorProgress();

console.log(`Active agents: ${status.agents.length}`);
console.log(`Tasks in progress: ${status.tasks.filter(t => t.status === 'in_progress').length}`);
console.log(`Memory entries: ${status.memory.entries}`);
```

## Task Decomposition Strategies

### Domain Strategy
Decomposes tasks based on domain expertise:
- Researcher: Requirements analysis
- Architect: System design
- Coder: Implementation
- Tester: Quality assurance
- Reviewer: Code review and optimization

### Parallel Strategy
Decomposes for maximum parallelization:
- Analyzer: Component analysis
- Designer: Interface design
- Documenter: Documentation

### Sequential Strategy
SPARC methodology implementation:
1. Specification: Define requirements
2. Pseudocode: Algorithm design
3. Architecture: System structure
4. Refinement: Iterative improvement
5. Completion: Final integration

### Adaptive Strategy
Dynamically selects strategy based on:
- Task complexity assessment
- Domain detection
- Resource availability
- Historical performance

## Agent Types

### Core Agents
- `researcher`: Research and analysis
- `architect`: System design
- `coder`: Implementation
- `tester`: Testing and validation
- `reviewer`: Code review

### Specialized Agents
- `analyzer`: Deep analysis
- `designer`: UI/UX design
- `documenter`: Documentation
- `optimizer`: Performance optimization
- `security`: Security analysis

### SPARC Agents
- `specification`: Requirements specification
- `pseudocode`: Algorithm design
- `architecture`: System architecture
- `refinement`: Code refinement
- `completion`: Integration and completion

## Event System

The orchestrator emits events for monitoring and integration:

```javascript
orchestrator.on('initialized', () => {
  console.log('Orchestrator ready');
});

orchestrator.on('agent-spawned', (agent) => {
  console.log(`Agent ${agent.id} spawned`);
});

orchestrator.on('subtask-completed', ({ agent, subtask, result }) => {
  console.log(`${agent.type} completed: ${subtask.work}`);
});

orchestrator.on('task-completed', ({ taskId, results }) => {
  console.log(`Task ${taskId} completed`);
});

orchestrator.on('progress-update', (status) => {
  console.log('Progress:', status);
});
```

## Metrics and Performance

### Metrics Collection
The orchestrator automatically collects:
- Task duration
- Success rates
- Resource utilization
- Parallelization efficiency

### Performance Optimization
- Automatic complexity assessment
- Dynamic strategy selection
- Efficient agent allocation
- Result caching and reuse

## Integration with Claude Flow

### Using with Claude Flow CLI

```bash
# Initialize orchestrator
npx claude-flow@alpha sparc run orchestrator "coordinate feature development"

# With specific strategy
npx claude-flow@alpha sparc run orchestrator "build authentication" --strategy adaptive

# With pattern preference
npx claude-flow@alpha sparc run orchestrator "review code" --pattern consensus
```

### Memory Integration

```bash
# Store orchestration results
npx claude-flow@alpha memory store --key "orchestration-results" --value '{"taskId":"...","results":"..."}'

# Retrieve for next session
npx claude-flow@alpha memory retrieve --key "orchestration-results"
```

## Best Practices

### Task Description
- Be specific and detailed in task descriptions
- Include domain keywords for better decomposition
- Specify constraints and requirements

### Strategy Selection
- Use `adaptive` for unknown complexity
- Use `parallel` for independent tasks
- Use `sequential` for dependent workflows
- Use `consensus` for critical decisions

### Memory Management
- Share context between related agents
- Clean up memory after task completion
- Use meaningful keys for shared data

### Error Handling
- Enable `retryOnFailure` for critical tasks
- Monitor progress for long-running tasks
- Implement proper cleanup in error cases

## Examples

### Example 1: Feature Development

```javascript
const task = 'Develop user profile management feature with avatar upload';

const results = await orchestrator.coordinateTask(task, {
  strategy: 'domain',
  parallel: false  // Sequential to maintain dependencies
});

// Results include work from all domain experts
console.log(results.summary.insights);
console.log(results.summary.recommendations);
```

### Example 2: Code Review Process

```javascript
const task = 'Review pull request #123 for security and performance';

const pattern = OrchestrationPatterns.consensus;
const results = await pattern.execute(task, orchestrator);

if (results.consensus) {
  console.log('PR approved by consensus');
} else {
  console.log('Further review needed');
}
```

### Example 3: System Analysis

```javascript
const task = 'Analyze system architecture for bottlenecks and optimization opportunities';

const results = await orchestrator.coordinateTask(task, {
  strategy: 'adaptive',
  timeout: 900000  // 15 minutes for thorough analysis
});

// Access detailed metrics
console.log('Performance metrics:', results.metrics);
console.log('Optimization recommendations:', results.summary.recommendations);
```

## Troubleshooting

### Common Issues

1. **Agent spawn failures**
   - Ensure claude-flow is properly installed
   - Check swarm initialization status
   - Verify available system resources

2. **Task timeout**
   - Increase timeout for complex tasks
   - Consider breaking into smaller subtasks
   - Use parallel execution when possible

3. **Memory issues**
   - Implement proper cleanup
   - Monitor memory usage
   - Use TTL for temporary data

4. **Poor parallelization**
   - Review task dependencies
   - Use appropriate decomposition strategy
   - Ensure agents are properly configured

## CLI Usage

### Command Line Interface

```bash
# Run orchestrator directly
node src/sparc-orchestrator.js "task description" [strategy]

# Examples
node src/sparc-orchestrator.js "Build REST API" adaptive
node src/sparc-orchestrator.js "Analyze codebase" parallel
node src/sparc-orchestrator.js "Implement SPARC workflow" sequential
```

### Integration with npm scripts

Add to package.json:

```json
{
  "scripts": {
    "orchestrate": "node src/sparc-orchestrator.js",
    "orchestrate:adaptive": "node src/sparc-orchestrator.js \"$npm_config_task\" adaptive",
    "orchestrate:parallel": "node src/sparc-orchestrator.js \"$npm_config_task\" parallel"
  }
}
```

Usage:
```bash
npm run orchestrate:adaptive --task="Build feature"
```

## API Reference

### SPARCOrchestrator Class

#### Constructor
```javascript
new SPARCOrchestrator()
```

#### Methods

##### initialize()
Initialize the orchestrator environment
- Returns: `Promise<boolean>`

##### decomposeTask(task, strategy)
Decompose task into subtasks
- `task`: Task description string
- `strategy`: 'domain' | 'parallel' | 'sequential' | 'adaptive'
- Returns: Array of subtask objects

##### spawnAgent(type, capabilities)
Spawn a new agent
- `type`: Agent type string
- `capabilities`: Array of capability strings
- Returns: `Promise<Agent>`

##### coordinateTask(task, options)
Coordinate task execution
- `task`: Task description string
- `options`: Configuration object
- Returns: `Promise<Results>`

##### shareMemory(key, value)
Share data in memory
- `key`: Memory key string
- `value`: Data to share
- Returns: `Promise<void>`

##### getMemory(key)
Retrieve from memory
- `key`: Memory key string
- Returns: `Promise<any>`

##### monitorProgress()
Get current progress status
- Returns: `Promise<Status>`

##### cleanup()
Clean up resources
- Returns: `Promise<void>`

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on contributing to the SPARC Orchestrator.

## License

MIT License - See [LICENSE](../LICENSE) for details.