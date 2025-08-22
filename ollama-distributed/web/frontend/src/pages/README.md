# OllamaMax Dashboard

A comprehensive, real-time dashboard for monitoring and managing the OllamaMax distributed AI platform. Built with React, TypeScript, and Recharts for data visualization.

## 🚀 Features

### Real-time Monitoring
- **Live Metrics**: CPU, memory, disk, and network utilization
- **System Health**: Overall cluster health with detailed component status
- **Node Management**: Active nodes, their status, and distribution
- **Model Synchronization**: Track model sync status across nodes
- **Task Management**: Monitor running, pending, and completed tasks

### Interactive Components
- **Activity Feed**: Real-time events and system updates
- **Alert Management**: System alerts with acknowledgment capabilities
- **Quick Actions**: Common operations and emergency controls
- **Data Visualization**: Interactive charts for performance trends

### Export & Analytics
- **Multi-format Export**: PDF, Excel, JSON, and CSV formats
- **Historical Data**: Performance trends and historical analysis
- **Custom Reports**: Configurable time ranges and metrics

### Responsive Design
- **Mobile-first**: Optimized for all screen sizes
- **Dark Mode**: Full dark/light theme support
- **Accessibility**: WCAG 2.1 AA compliant

## 📁 Component Structure

```
src/pages/
├── Dashboard.tsx              # Main dashboard component
├── components/               # Dashboard-specific components
│   ├── MetricCard.tsx        # Key metrics display cards
│   ├── SystemHealthCard.tsx  # Overall system health
│   ├── ActivityFeedCard.tsx  # Real-time activity feed
│   ├── QuickActionsCard.tsx  # Common operations panel
│   ├── AlertsCard.tsx        # System alerts management
│   ├── ExportUtils.tsx       # Data export utilities
│   └── index.ts             # Component exports
└── dashboard/
    └── DashboardPage.tsx     # Legacy wrapper (redirects to Dashboard)
```

## 🧩 Component Architecture

### MetricCard
Displays key system metrics with trends and status indicators.

**Props:**
- `title`: Metric display name
- `value`: Current metric value
- `total?`: Total capacity (for ratio display)
- `icon`: Icon component
- `trend?`: Trend indicator (up/down arrow)
- `status`: Health status (healthy, warning, error, info)
- `subtitle?`: Additional context text
- `change?`: Change indicator with percentage and period

**Features:**
- Loading states with skeleton animation
- Trend indicators with colored arrows
- Status badges with semantic colors
- Responsive formatting for large numbers

### SystemHealthCard
Comprehensive system health overview with health checks.

**Props:**
- `clusterStatus?`: Cluster status data
- `metrics`: System performance metrics

**Features:**
- Overall health score calculation
- Individual component health checks
- 24-hour status timeline
- Resource utilization summary

### ActivityFeedCard
Real-time activity feed with filtering and categorization.

**Props:**
- `activities`: Array of activity items
- `onRefresh`: Refresh callback function
- `maxItems?`: Maximum items to display (default: 10)

**Features:**
- Activity type filtering (nodes, models, tasks, alerts)
- Real-time timestamp formatting
- Severity-based color coding
- Activity statistics summary

### QuickActionsCard
Panel for common operations and emergency controls.

**Features:**
- Primary actions (add node, sync models, download models)
- System actions (health check, view logs)
- Emergency controls (emergency stop)
- Action execution states with loading indicators

### AlertsCard
System alerts management with acknowledgment capabilities.

**Props:**
- `alerts`: Array of system alerts
- `onAcknowledge`: Alert acknowledgment callback
- `maxItems?`: Maximum alerts to display (default: 5)

**Features:**
- Alert type filtering (error, warning, info)
- Bulk acknowledgment
- Alert severity indicators
- Time-based sorting

### ExportUtils
Data export utilities supporting multiple formats.

**Props:**
- `data`: Dashboard data to export
- `filename?`: Base filename for exports
- `className?`: Additional CSS classes

**Supported Formats:**
- **PDF**: Complete dashboard report with charts and tables
- **Excel**: Multi-worksheet data with metrics, activities, and alerts
- **JSON**: Raw data format for programmatic access
- **CSV**: Metrics summary in comma-separated format

## 🎨 Design System Integration

The dashboard uses the OllamaMax design system with:

- **Colors**: Semantic color tokens for status, themes, and branding
- **Typography**: Consistent text hierarchy and readability
- **Spacing**: 8px grid system for consistent layouts
- **Borders**: Rounded corners and elevation shadows
- **Animations**: Smooth transitions with 300ms ease-in-out

### Theme Support
- Light and dark mode variants
- High contrast mode compatibility
- CSS custom properties for theming
- Tailwind CSS utility classes

## 📊 Data Visualization

Built with Recharts for interactive charts:

### Performance Charts
- **Line Charts**: CPU, memory, network utilization over time
- **Area Charts**: Task throughput (completed vs failed)
- **Pie Charts**: Node distribution by status

### Chart Features
- Responsive container sizing
- Theme-aware colors
- Interactive tooltips
- Legend integration
- Real-time data updates

## 🔌 WebSocket Integration

Real-time updates through WebSocket connections:

```typescript
const { isConnected, connectionState } = useWebSocket()
const { data: clusterStatus } = useClusterStatus()
const { data: metrics } = useMetrics()
const { notifications } = useNotifications()
```

### Connection Management
- Automatic reconnection with exponential backoff
- Connection state indicators
- Error handling and fallback states
- Configurable refresh intervals

## 📱 Responsive Behavior

### Desktop (1024px+)
- Full 3-column layout with sidebar
- All components visible
- Detailed charts and tables

### Tablet (768px - 1023px)
- 2-column layout
- Collapsible sidebar
- Simplified charts

### Mobile (< 768px)
- Single column stack
- Horizontal scroll for tables
- Touch-optimized controls
- Simplified metrics cards

## ♿ Accessibility Features

### WCAG 2.1 AA Compliance
- **Keyboard Navigation**: Full keyboard accessibility
- **Screen Readers**: ARIA labels and live regions
- **Color Contrast**: Minimum 4.5:1 contrast ratios
- **Focus Management**: Visible focus indicators
- **Semantic HTML**: Proper heading hierarchy

### Interactive Elements
- Button states and feedback
- Form validation messages
- Error announcements
- Loading state indicators

## 🚀 Performance Optimizations

### Code Splitting
- Lazy component loading
- Dynamic imports for charts
- Route-based splitting

### Data Management
- Efficient re-renders with useMemo/useCallback
- Optimistic updates for real-time data
- Debounced search and filters

### Asset Optimization
- SVG icons for scalability
- Compressed images
- Optimized bundle sizes

## 🧪 Testing Strategy

### Unit Tests
- Component rendering tests
- Props and state management
- Event handling verification
- Accessibility compliance

### Integration Tests
- WebSocket connection handling
- Data flow between components
- Export functionality
- Theme switching

### E2E Tests
- Complete user workflows
- Cross-browser compatibility
- Performance benchmarks
- Accessibility audits

## 🔧 Development

### Local Development
```bash
npm run dev          # Start development server
npm run typecheck    # Run TypeScript checks
npm run lint         # Run ESLint
npm run test         # Run unit tests
npm run test:e2e     # Run E2E tests
```

### Building
```bash
npm run build        # Production build
npm run preview      # Preview production build
```

### Environment Configuration
- Development: Real-time WebSocket connections
- Production: Optimized bundles and CDN assets
- Testing: Mock data and offline mode

## 📋 Usage Examples

### Basic Implementation
```tsx
import Dashboard from '@/pages/Dashboard'

function App() {
  return <Dashboard />
}
```

### Custom Configuration
```tsx
import { DashboardProvider } from '@/pages/dashboard/context'

function App() {
  return (
    <DashboardProvider
      refreshInterval={5000}
      autoRefresh={true}
      theme="dark"
    >
      <Dashboard />
    </DashboardProvider>
  )
}
```

### Component Usage
```tsx
import { MetricCard, SystemHealthCard } from '@/pages/components'

function CustomDashboard() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      <MetricCard
        title="Active Nodes"
        value={5}
        total={8}
        icon={<Server />}
        status="healthy"
      />
      <SystemHealthCard
        metrics={systemMetrics}
      />
    </div>
  )
}
```

## 🤝 Contributing

1. Follow the existing component patterns
2. Add TypeScript types for all props
3. Include accessibility features
4. Write comprehensive tests
5. Update documentation

## 📄 License

Part of the OllamaMax distributed AI platform.