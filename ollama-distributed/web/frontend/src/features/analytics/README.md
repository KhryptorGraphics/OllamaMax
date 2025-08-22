# Analytics & Reporting System

A comprehensive analytics and reporting system for the Ollama Distributed frontend, providing real-time insights, business intelligence, and regulatory compliance reporting.

## Features

### 📊 Core Analytics
- **Event Tracking**: Page views, clicks, form submissions, API calls, errors
- **Performance Monitoring**: Web Vitals, runtime metrics, resource usage
- **User Behavior Analysis**: Session tracking, interaction patterns, conversion funnels
- **Real-time Streaming**: WebSocket-based live data updates

### 📈 Business Intelligence
- **Revenue Analytics**: Product performance, regional analysis, growth metrics
- **User Segmentation**: Demographics, behavior-based segments, cohort analysis
- **Conversion Tracking**: Multi-step funnels, goal completion, attribution
- **Retention Analysis**: Cohort retention, churn prediction, lifetime value

### 📋 Compliance Reporting
- **GDPR Compliance**: Data processing records, user rights management, breach reporting
- **CCPA Support**: Consumer privacy requests, data deletion, opt-out tracking
- **HIPAA Ready**: Healthcare data compliance, audit trails, security monitoring
- **SOX/PCI-DSS**: Financial and payment card industry compliance

### 📑 Advanced Reporting
- **Export Formats**: PDF, CSV, Excel, JSON with professional formatting
- **Scheduled Reports**: Automated report generation and distribution
- **Interactive Dashboards**: Real-time widgets, customizable layouts
- **Data Visualization**: Charts, heatmaps, trend analysis

## Architecture

```
src/features/
├── analytics/
│   ├── components/
│   │   ├── AnalyticsDashboard.tsx      # Main analytics dashboard
│   │   ├── BusinessIntelligence.tsx    # BI dashboard and insights
│   │   └── charts/
│   │       └── Chart.tsx               # Reusable chart components
│   ├── services/
│   │   └── analyticsService.ts         # Core analytics service
│   ├── hooks/
│   │   └── useRealTimeAnalytics.ts     # Real-time data hook
│   ├── utils/
│   │   └── dataAggregation.ts          # Data processing utilities
│   └── types/
│       └── index.ts                    # TypeScript definitions
├── reporting/
│   ├── components/
│   │   ├── ReportBuilder.tsx           # Interactive report builder
│   │   └── PerformanceDashboard.tsx    # Performance metrics dashboard
│   └── services/
│       └── exportService.ts            # Export functionality
└── compliance/
    └── components/
        └── ComplianceReportingDashboard.tsx  # Compliance dashboard
```

## Quick Start

### 1. Import Components

```tsx
import { 
  AnalyticsDashboard, 
  BusinessIntelligence,
  ReportBuilder,
  PerformanceDashboard,
  ComplianceReportingDashboard 
} from '@/features/analytics'
```

### 2. Initialize Analytics Service

```tsx
import { analyticsService } from '@/features/analytics'

// Configure analytics
analyticsService.configure({
  enableRealTime: true,
  enablePerformanceTracking: true,
  enableErrorTracking: true,
  batchEvents: true
})

// Set user context
analyticsService.setUserId('user-123')
```

### 3. Track Events

```tsx
// Page views
analyticsService.trackPageView('/dashboard', 'Dashboard')

// User interactions
analyticsService.trackClick('export-button', 'reporting')

// API calls
analyticsService.trackApiCall('/api/users', 'GET', 200, 150, 1024)

// Custom events
analyticsService.trackCustomEvent('feature_used', {
  feature: 'advanced_filtering',
  context: 'user_dashboard'
})
```

### 4. Use Real-time Hook

```tsx
import { useRealTimeAnalytics } from '@/features/analytics'

function Dashboard() {
  const {
    metrics,
    events,
    errors,
    isConnected,
    stats
  } = useRealTimeAnalytics({
    enablePerformanceTracking: true,
    enableErrorTracking: true,
    updateInterval: 5000
  })

  return (
    <div>
      <p>Active Users: {metrics?.activeUsers}</p>
      <p>Events/sec: {stats.eventsPerSecond.toFixed(2)}</p>
      <p>Connection: {isConnected ? 'Connected' : 'Disconnected'}</p>
    </div>
  )
}
```

### 5. Data Aggregation

```tsx
import { aggregateData, TimePeriod } from '@/features/analytics'

// Aggregate events by type
const aggregated = aggregateData(events, {
  groupBy: ['type', 'category'],
  metrics: [
    { field: 'timestamp', aggregation: 'count', alias: 'event_count' },
    { field: 'value', aggregation: 'average', alias: 'avg_value' }
  ],
  orderBy: { field: 'event_count', direction: 'desc' },
  limit: 10
})

// Time series analysis
const timeSeries = aggregateTimeSeries(
  events,
  'timestamp',
  'value',
  TimePeriod.HOUR
)
```

### 6. Export Reports

```tsx
import { exportService } from '@/features/reporting'

// Export as PDF
const result = await exportService.exportReport(
  {
    title: 'Monthly Analytics Report',
    data: aggregatedData,
    summary: { totalEvents: 1000, avgResponseTime: 250 },
    charts: [
      {
        title: 'Events Over Time',
        type: 'line',
        data: timeSeries
      }
    ]
  },
  'pdf',
  {
    includeCharts: true,
    includeData: true,
    branding: true,
    watermark: 'CONFIDENTIAL'
  }
)

console.log(`Report exported: ${result.filename}`)
```

## Configuration

### Analytics Service Configuration

```tsx
analyticsService.configure({
  enableRealTime: true,           // WebSocket streaming
  enablePerformanceTracking: true, // Web Vitals monitoring
  enableErrorTracking: true,      // Error event capture
  enableUserTracking: true,       // User behavior analysis
  batchEvents: true,              // Batch API calls
  batchSize: 50,                  // Events per batch
  batchInterval: 30000,           // Batch frequency (ms)
  endpoint: '/api/analytics/events', // API endpoint
  websocketUrl: '/ws/analytics'   // WebSocket URL
})
```

### Chart Configuration

```tsx
<Chart
  type="line"
  data={chartData}
  config={{
    xKey: 'timestamp',
    yKey: ['users', 'sessions'],
    colorScheme: 'blue',
    smooth: true,
    showGrid: true,
    showLegend: true,
    yAxisFormatter: (value) => `${value.toLocaleString()}`,
    referenceLines: [
      { value: 1000, label: 'Target', color: '#ef4444' }
    ]
  }}
  height={400}
/>
```

## API Reference

### AnalyticsService

| Method | Description | Parameters |
|--------|-------------|------------|
| `track()` | Track custom event | `type, category, action, label?, value?, metadata?` |
| `trackPageView()` | Track page navigation | `path, title?, referrer?` |
| `trackClick()` | Track element clicks | `element, category?` |
| `trackFormSubmit()` | Track form submissions | `formName, success, fields?` |
| `trackApiCall()` | Track API requests | `endpoint, method, status, duration, size?` |
| `trackError()` | Track error events | `error, context?` |
| `getRealTimeMetrics()` | Get current metrics | None |
| `exportData()` | Export analytics data | `format, dateRange?` |

### DataAggregator

| Method | Description | Parameters |
|--------|-------------|------------|
| `aggregate()` | Aggregate data with grouping | `data, config` |
| `aggregateTimeSeries()` | Time-based aggregation | `data, dateField, valueField, period` |
| `analyzeCohorts()` | Cohort retention analysis | `data, cohortField, returnField, userField` |
| `analyzeFunnel()` | Conversion funnel analysis | `data, steps, userField` |
| `calculateStatistics()` | Statistical analysis | `values` |
| `detectAnomalies()` | Anomaly detection | `values, method?, threshold?` |

### ExportService

| Method | Description | Parameters |
|--------|-------------|------------|
| `exportReport()` | Export data in various formats | `data, format, options?` |
| `exportBusinessMetrics()` | Export BI data | `metrics, format, options?` |
| `exportPerformanceMetrics()` | Export performance data | `metrics, format, options?` |
| `exportComplianceReport()` | Export compliance data | `report, format, options?` |

## WebSocket Protocol

The real-time analytics system uses WebSocket for live data streaming:

### Client Messages
```json
{
  "type": "subscribe",
  "config": {
    "performance": true,
    "errors": true,
    "business": false,
    "interval": 5000
  }
}
```

### Server Messages
```json
{
  "type": "real_time_metrics",
  "payload": {
    "activeUsers": 1250,
    "pageViews": 15000,
    "events": [...],
    "performance": {...},
    "errors": [...],
    "timestamp": 1674567890000
  }
}
```

## Privacy & Compliance

### GDPR Compliance
- **Consent Management**: Track and honor user consent preferences
- **Data Minimization**: Only collect necessary analytics data
- **Right to Access**: Export user's analytics data
- **Right to Erasure**: Delete user's analytics data
- **Data Portability**: Provide data in machine-readable format

### Security Features
- **Data Encryption**: All data encrypted in transit and at rest
- **Access Control**: Role-based access to analytics data
- **Audit Logging**: Complete audit trail of data access
- **Anonymization**: Remove PII from analytics events
- **Retention Policies**: Automatic data cleanup after retention period

## Testing

The analytics system includes comprehensive tests:

```bash
# Run all analytics tests
npm run test src/tests/analytics.test.ts

# Run with coverage
npm run test:coverage

# Run E2E tests
npm run test:e2e
```

## Performance Considerations

### Optimization Strategies
- **Event Batching**: Reduce API calls by batching events
- **Local Storage**: Cache events offline for reliability
- **WebSocket Streaming**: Real-time updates without polling
- **Data Compression**: Compress exports and large datasets
- **Lazy Loading**: Load dashboards components on demand

### Performance Budgets
- **Bundle Size**: Analytics features <100KB gzipped
- **Memory Usage**: <50MB for dashboard with 10K events
- **API Response**: <200ms for aggregation queries
- **Export Speed**: <5s for 100K events CSV export
- **Real-time Latency**: <100ms for live event streaming

## Contributing

1. Follow TypeScript best practices
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure GDPR compliance for data handling
5. Optimize for performance and bundle size

## License

This analytics system is part of the Ollama Distributed project and follows the same license terms.