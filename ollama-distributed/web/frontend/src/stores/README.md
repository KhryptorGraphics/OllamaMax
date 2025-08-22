# Enhanced Zustand Store for Sprint C

Comprehensive state management solution for OllamaMax distributed system with real-time synchronization, performance optimizations, and robust error handling.

## 🎯 Overview

The enhanced Zustand store provides four specialized slices for Sprint C pages:

- **ModelsSlice**: Model management, deployment, and synchronization
- **NodesSlice**: Node operations, health monitoring, and metrics
- **MonitoringSlice**: Real-time metrics, alerts, and performance tracking
- **DashboardSlice**: System summary, activity feeds, and quick actions

## 🚀 Key Features

### ⚡ Performance Optimizations

- **Request Deduplication**: Prevents duplicate API calls
- **Intelligent Caching**: TTL-based caching with automatic invalidation
- **Debounced Updates**: Optimizes frequent state changes
- **Memoized Selectors**: Prevents unnecessary re-renders
- **Selective Updates**: Updates only changed data

### 🔄 Real-time Synchronization

- **WebSocket Integration**: Auto-connects on authentication
- **Channel Subscriptions**: Model, node, and metrics updates
- **Auto-refresh Logic**: Configurable polling intervals
- **Conflict Resolution**: Handles concurrent updates gracefully
- **State Persistence**: Maintains state across sessions

### 🛡️ Error Handling & Resilience

- **Retry Logic**: Exponential backoff with jitter
- **Circuit Breaker**: Prevents cascading failures
- **Optimistic Updates**: Immediate UI feedback with rollback
- **Graceful Degradation**: Fallback mechanisms
- **Comprehensive Logging**: Structured error tracking

## 📦 Store Structure

```
stores/
├── index.ts              # Main store with all slices
├── utils/
│   ├── debounce.ts      # Debouncing utilities
│   └── retry.ts         # Retry logic and circuit breaker
├── usage-examples.ts    # Implementation examples
└── README.md           # This documentation
```

## 🔧 Usage Examples

### Dashboard Page

```typescript
import { useDashboardSelector, useSystemHealthSelector } from '@/stores'

function DashboardPage() {
  const dashboard = useDashboardSelector()
  const systemHealth = useSystemHealthSelector()
  
  useEffect(() => {
    const { fetchDashboardData, subscribeToDashboardUpdates } = useStore.getState()
    
    // Fetch initial data (cached if available)
    fetchDashboardData()
    
    // Subscribe to real-time updates
    subscribeToDashboardUpdates()
    
    return () => {
      // Cleanup on unmount
      const { unsubscribeFromDashboardUpdates } = useStore.getState()
      unsubscribeFromDashboardUpdates()
    }
  }, [])

  return (
    <div>
      <h1>System Health: {systemHealth}</h1>
      <p>Total Nodes: {dashboard.summary.totalNodes}</p>
      <p>Healthy Nodes: {dashboard.summary.healthyNodes}</p>
    </div>
  )
}
```

### Models Page

```typescript
import { useModelsSelector, useModelsSummarySelector } from '@/stores'

function ModelsPage() {
  const models = useModelsSelector()
  const summary = useModelsSummarySelector()
  
  const handleDeployModel = async (modelName: string) => {
    const { deployModel } = useStore.getState()
    
    try {
      // Optimistic update - UI shows immediate feedback
      await deployModel(modelName, ['node-1', 'node-2'])
    } catch (error) {
      // UI automatically reverts on error
      console.error('Deployment failed:', error)
    }
  }

  return (
    <div>
      <h1>Models ({summary.total})</h1>
      <p>Syncing: {summary.syncing}</p>
      <p>Failed: {summary.failed}</p>
      
      {models.models.map(model => (
        <div key={model.name}>
          <h3>{model.name}</h3>
          <button onClick={() => handleDeployModel(model.name)}>
            Deploy
          </button>
        </div>
      ))}
    </div>
  )
}
```

### Monitoring Page

```typescript
import { useMonitoringSelector, useActiveAlertsSelector } from '@/stores'

function MonitoringPage() {
  const monitoring = useMonitoringSelector()
  const activeAlerts = useActiveAlertsSelector()
  
  useEffect(() => {
    const store = useStore.getState()
    
    // Start real-time monitoring
    store.fetchMetrics()
    store.subscribeToMetricsUpdates()
    store.setAutoRefresh(true)
    
    return () => {
      // Cleanup
      store.unsubscribeFromMetricsUpdates()
      store.setAutoRefresh(false)
    }
  }, [])

  const handleAcknowledgeAlert = async (alertId: string) => {
    const { acknowledgeAlert } = useStore.getState()
    await acknowledgeAlert(alertId)
  }

  return (
    <div>
      <h1>System Monitoring</h1>
      
      {monitoring.metrics && (
        <div>
          <p>CPU: {monitoring.metrics.system.cpu.current}%</p>
          <p>Memory: {monitoring.metrics.system.memory.current}%</p>
        </div>
      )}
      
      <h2>Active Alerts ({activeAlerts.length})</h2>
      {activeAlerts.map(alert => (
        <div key={alert.id}>
          <h4>{alert.message}</h4>
          <button onClick={() => handleAcknowledgeAlert(alert.id)}>
            Acknowledge
          </button>
        </div>
      ))}
    </div>
  )
}
```

## 🔌 API Integration

The store integrates with the enhanced API client providing:

### Dashboard APIs
- `GET /dashboard/summary` - System summary
- `GET /dashboard/activity` - Recent activity

### Models APIs
- `GET /models` - List models
- `POST /models/{name}/deploy` - Deploy model
- `POST /models/{name}/undeploy` - Undeploy model
- `POST /models/upload` - Upload model
- `DELETE /models/{name}` - Delete model

### Nodes APIs
- `GET /nodes` - List nodes
- `GET /nodes/{id}` - Node details
- `GET /nodes/{id}/metrics` - Node metrics
- `POST /nodes/{id}/drain` - Drain node
- `POST /nodes/{id}/enable` - Enable node
- `DELETE /nodes/{id}` - Remove node

### Monitoring APIs
- `GET /metrics` - Performance metrics
- `GET /alerts` - System alerts
- `POST /alerts/{id}/acknowledge` - Acknowledge alert
- `POST /alerts/{id}/resolve` - Resolve alert

## 🎛️ Configuration

### Cache TTL Settings

```typescript
const CACHE_TTL = {
  models: 30000,    // 30 seconds
  nodes: 15000,     // 15 seconds 
  metrics: 10000,   // 10 seconds
  dashboard: 20000, // 20 seconds
}
```

### WebSocket Subscriptions

```typescript
// Auto-subscribed channels on authentication
const channels = [
  'models:*',       // Model updates
  'nodes:*',        // Node status
  'metrics:*',      // Performance metrics
  'alerts:*',       // System alerts
  'dashboard:*',    // Dashboard updates
]
```

### Auto-refresh Settings

```typescript
const defaultSettings = {
  autoRefresh: false,
  refreshInterval: 10000, // 10 seconds
}
```

## 🔍 Memoized Selectors

Pre-built selectors for optimal performance:

```typescript
// Basic selectors
const models = useModelsSelector()
const nodes = useNodesSelector()
const monitoring = useMonitoringSelector()
const dashboard = useDashboardSelector()

// Derived selectors
const systemHealth = useSystemHealthSelector()
const modelsSummary = useModelsSummarySelector()
const activeAlerts = useActiveAlertsSelector()
```

## 🚨 Error Handling

### Automatic Retry

```typescript
// Retry with exponential backoff
const response = await retryWithBackoff(() => 
  apiClient.getModels(), 
  {
    maxRetries: 3,
    baseDelay: 1000,
    backoffFactor: 2,
  }
)
```

### Optimistic Updates

```typescript
// Immediate UI feedback with automatic rollback on error
const deployModel = async (modelName: string) => {
  return optimisticUpdate(
    // Optimistic state update
    (optimisticData) => updateModelState(optimisticData),
    // Optimistic data
    { status: 'syncing', progress: 0 },
    // Actual API call
    () => apiClient.deployModel(modelName),
    // Rollback on error
    (error) => revertModelState(error)
  )
}
```

### Circuit Breaker

```typescript
const circuitBreaker = new CircuitBreaker(apiCall, {
  failureThreshold: 5,
  timeout: 60000,
  resetTimeout: 30000,
})

await circuitBreaker.execute()
```

## 📊 Performance Monitoring

### Request Deduplication

```typescript
// Prevents duplicate requests for same data
const data = await dedupedRequest('cache-key', () => 
  apiClient.getModels()
)
```

### Batch Operations

```typescript
// Execute multiple operations efficiently
const results = await retryBulk([
  () => fetchModels(),
  () => fetchNodes(),
  () => fetchMetrics(),
])
```

## 🔐 Security Features

- **Authentication Integration**: Auto-connects WebSocket on login
- **Token Management**: Automatic token refresh
- **Data Sanitization**: Validates all API responses
- **Secure Cleanup**: Clears sensitive data on logout

## 🧪 Testing

The store includes comprehensive error boundaries and fallback mechanisms:

- **Network Failures**: Automatic retry with backoff
- **Authentication Errors**: Graceful logout and cleanup
- **Server Errors**: Circuit breaker protection
- **Data Corruption**: Validation and sanitization

## 📈 Monitoring & Analytics

- **Performance Metrics**: Request timing and success rates
- **Error Tracking**: Structured error logging
- **Cache Efficiency**: Hit/miss ratios
- **Real-time Updates**: WebSocket message counts

## 🔄 Migration Guide

From basic store to enhanced store:

1. **Update imports**:
   ```typescript
   // Before
   import { useStore } from '@/stores'
   
   // After
   import { useModelsSelector, useNodesSelector } from '@/stores'
   ```

2. **Use memoized selectors**:
   ```typescript
   // Before
   const models = useStore(state => state.models)
   
   // After
   const models = useModelsSelector()
   ```

3. **Enable real-time updates**:
   ```typescript
   useEffect(() => {
     const { subscribeToModelUpdates } = useStore.getState()
     subscribeToModelUpdates()
     
     return () => {
       const { unsubscribeFromModelUpdates } = useStore.getState()
       unsubscribeFromModelUpdates()
     }
   }, [])
   ```

## 📚 Additional Resources

- [usage-examples.ts](./usage-examples.ts) - Comprehensive usage examples
- [utils/debounce.ts](./utils/debounce.ts) - Debouncing utilities
- [utils/retry.ts](./utils/retry.ts) - Retry logic and resilience patterns
- [API Client Documentation](../services/api/README.md) - API integration guide

## 🤝 Contributing

When extending the store:

1. **Follow TypeScript patterns**: Use proper type definitions
2. **Implement error handling**: Add try-catch and rollback logic
3. **Add caching**: Use TTL-based caching for performance
4. **Enable real-time**: Add WebSocket subscription support
5. **Write tests**: Include unit tests for new functionality
6. **Update documentation**: Keep this README current

---

**Note**: This enhanced store is specifically designed for Sprint C requirements and provides the foundation for all subsequent development phases.