# Component Specifications - Distributed Ollama Frontend

## Overview

This document provides detailed specifications for all UI components in the distributed Ollama frontend, organized by functional domains. Each component includes props interface, styling guidelines, accessibility requirements, and implementation notes.

## Design System Components

### Theme Provider

```typescript
interface ThemeProviderProps {
  theme?: 'light' | 'dark' | 'auto';
  children: React.ReactNode;
  customColors?: Partial<ColorTokens>;
}

const ThemeProvider: React.FC<ThemeProviderProps> = ({ 
  theme = 'auto', 
  children, 
  customColors 
}) => {
  // Implementation with styled-components ThemeProvider
  // System theme detection for 'auto' mode
  // CSS custom properties for theme switching
};
```

**Features**:
- Automatic system theme detection
- Smooth theme transitions (300ms)
- Custom color override support
- CSS custom properties integration

**Accessibility**:
- Respect user's system preference
- Maintain contrast ratios across themes
- No flash of unstyled content (FOUC)

---

### Button Component

```typescript
interface ButtonProps {
  variant?: 'primary' | 'secondary' | 'outlined' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  icon?: React.ReactNode;
  iconPosition?: 'left' | 'right';
  fullWidth?: boolean;
  children: React.ReactNode;
  onClick?: () => void;
  type?: 'button' | 'submit' | 'reset';
}

const Button: React.FC<ButtonProps> = ({ 
  variant = 'primary',
  size = 'md',
  disabled = false,
  loading = false,
  icon,
  iconPosition = 'left',
  fullWidth = false,
  children,
  onClick,
  type = 'button'
}) => {
  // Implementation with styled-components
  // Loading state with spinner
  // Icon positioning logic
};
```

**Styling Guidelines**:
- Primary: Love gradient background (#FF6B6B → #FFD93D)
- Secondary: White background, primary border
- Height: sm(32px), md(40px), lg(48px)
- Border radius: 8px
- Transition: all 300ms ease-in-out

**Accessibility**:
- ARIA labels for loading state
- Keyboard focus indicators
- Disabled state properly announced
- Minimum 44px touch target on mobile

---

### Input Component

```typescript
interface InputProps {
  type?: 'text' | 'email' | 'password' | 'number' | 'search';
  placeholder?: string;
  value?: string;
  defaultValue?: string;
  onChange?: (value: string) => void;
  onBlur?: () => void;
  onFocus?: () => void;
  disabled?: boolean;
  error?: string;
  label?: string;
  helperText?: string;
  required?: boolean;
  autoComplete?: string;
  icon?: React.ReactNode;
  iconPosition?: 'left' | 'right';
}

const Input: React.FC<InputProps> = ({ 
  type = 'text',
  placeholder,
  value,
  defaultValue,
  onChange,
  onBlur,
  onFocus,
  disabled = false,
  error,
  label,
  helperText,
  required = false,
  autoComplete,
  icon,
  iconPosition = 'left'
}) => {
  // Implementation with controlled/uncontrolled patterns
  // Error state styling
  // Icon positioning
};
```

**Features**:
- Controlled and uncontrolled modes
- Built-in validation styling
- Icon support with positioning
- Password visibility toggle

**Accessibility**:
- Label association with htmlFor
- Error announcement with aria-describedby
- Required field indication
- Appropriate autocomplete attributes

---

## Admin Dashboard Components

### NodeCard Component

```typescript
interface NodeInfo {
  id: string;
  name: string;
  status: 'online' | 'offline' | 'degraded' | 'maintenance';
  cpu: number; // percentage
  memory: number; // percentage
  network: number; // MB/s
  models: number; // count
  lastSeen: Date;
  version: string;
  region: string;
}

interface NodeCardProps {
  node: NodeInfo;
  onSelect?: (nodeId: string) => void;
  onAction?: (nodeId: string, action: string) => void;
  compact?: boolean;
}

const NodeCard: React.FC<NodeCardProps> = ({ 
  node, 
  onSelect, 
  onAction, 
  compact = false 
}) => {
  // Implementation with status indicators
  // Resource utilization bars
  // Action menu integration
};
```

**Visual Design**:
- Card container with subtle shadow
- Status indicator dot (green/red/yellow/gray)
- Resource bars with color coding
- Hover states with elevation

**Interactions**:
- Click to select/focus node
- Right-click for context menu
- Keyboard navigation support
- Touch-friendly on mobile

---

### ClusterTopology Component

```typescript
interface ClusterNode {
  id: string;
  name: string;
  position: { x: number; y: number };
  status: NodeStatus;
  connections: string[];
  metrics: NodeMetrics;
}

interface ClusterTopologyProps {
  nodes: ClusterNode[];
  selectedNode?: string;
  onNodeSelect?: (nodeId: string) => void;
  onNodeAction?: (nodeId: string, action: string) => void;
  layout?: 'force' | 'grid' | 'circle';
  showMetrics?: boolean;
}

const ClusterTopology: React.FC<ClusterTopologyProps> = ({ 
  nodes, 
  selectedNode, 
  onNodeSelect, 
  onNodeAction, 
  layout = 'force',
  showMetrics = true 
}) => {
  // D3.js force simulation
  // SVG rendering with React
  // Interactive zoom and pan
  // Real-time position updates
};
```

**Features**:
- Force-directed graph layout
- Interactive zoom and pan (mouse/touch)
- Real-time node status updates
- Connection strength visualization

**Performance**:
- Canvas fallback for 100+ nodes
- Efficient re-rendering with React.memo
- Debounced layout calculations
- Memory management for large graphs

---

### ModelCatalog Component

```typescript
interface ModelInfo {
  id: string;
  name: string;
  version: string;
  size: number; // bytes
  status: 'available' | 'downloading' | 'syncing' | 'error';
  description: string;
  tags: string[];
  downloadProgress?: number;
  lastUpdated: Date;
  nodeCount: number; // nodes with this model
}

interface ModelCatalogProps {
  models: ModelInfo[];
  view?: 'grid' | 'list';
  onModelSelect?: (modelId: string) => void;
  onModelAction?: (modelId: string, action: string) => void;
  filters?: ModelFilters;
  onFilterChange?: (filters: ModelFilters) => void;
  searchQuery?: string;
  onSearchChange?: (query: string) => void;
}

const ModelCatalog: React.FC<ModelCatalogProps> = ({ 
  models, 
  view = 'grid', 
  onModelSelect, 
  onModelAction, 
  filters, 
  onFilterChange, 
  searchQuery, 
  onSearchChange 
}) => {
  // Virtual scrolling for large lists
  // Search and filter integration
  // Grid/list view switching
  // Bulk operations support
};
```

**Features**:
- Grid and list view modes
- Search with highlighting
- Filter by status, size, tags
- Bulk selection and actions

**Performance**:
- Virtual scrolling for 1000+ models
- Debounced search input
- Memoized filter calculations
- Efficient re-rendering patterns

---

## Real-time Monitoring Components

### PerformanceChart Component

```typescript
interface DataPoint {
  timestamp: number;
  value: number;
  label?: string;
}

interface PerformanceChartProps {
  data: DataPoint[];
  type: 'line' | 'area' | 'bar';
  title: string;
  unit?: string;
  height?: number;
  timeRange?: '1h' | '24h' | '7d' | '30d';
  onTimeRangeChange?: (range: string) => void;
  realTime?: boolean;
  threshold?: number;
  showGrid?: boolean;
  interactive?: boolean;
}

const PerformanceChart: React.FC<PerformanceChartProps> = ({ 
  data, 
  type, 
  title, 
  unit, 
  height = 300, 
  timeRange = '1h', 
  onTimeRangeChange, 
  realTime = false, 
  threshold, 
  showGrid = true, 
  interactive = true 
}) => {
  // D3.js or Chart.js integration
  // Real-time data streaming
  // Interactive tooltips
  // Responsive design
};
```

**Features**:
- Multiple chart types (line, area, bar)
- Real-time data updates
- Interactive tooltips and zoom
- Threshold line indicators

**Performance**:
- 60fps smooth animations
- Efficient data point updates
- Canvas rendering for large datasets
- Memory optimization for long sessions

---

### AlertCenter Component

```typescript
interface Alert {
  id: string;
  title: string;
  description: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  category: 'performance' | 'security' | 'system' | 'user';
  timestamp: Date;
  status: 'new' | 'acknowledged' | 'resolved';
  source: string;
  nodeId?: string;
  actions?: AlertAction[];
}

interface AlertCenterProps {
  alerts: Alert[];
  onAlertAction?: (alertId: string, action: string) => void;
  onBulkAction?: (alertIds: string[], action: string) => void;
  filters?: AlertFilters;
  onFilterChange?: (filters: AlertFilters) => void;
  realTime?: boolean;
}

const AlertCenter: React.FC<AlertCenterProps> = ({ 
  alerts, 
  onAlertAction, 
  onBulkAction, 
  filters, 
  onFilterChange, 
  realTime = true 
}) => {
  // Real-time alert streaming
  // Bulk operations support
  // Advanced filtering
  // Sound/visual notifications
};
```

**Features**:
- Real-time alert delivery
- Severity-based color coding
- Bulk acknowledgment actions
- Audio notifications for critical alerts

**Accessibility**:
- Screen reader announcements for new alerts
- Keyboard navigation for alert actions
- High contrast severity indicators
- Alternative notification methods

---

## Security Management Components

### UserTable Component

```typescript
interface User {
  id: string;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  status: 'active' | 'inactive' | 'suspended';
  roles: string[];
  lastLogin: Date | null;
  createdAt: Date;
  mfaEnabled: boolean;
}

interface UserTableProps {
  users: User[];
  selectedUsers?: string[];
  onUserSelect?: (userIds: string[]) => void;
  onUserAction?: (userId: string, action: string) => void;
  onBulkAction?: (userIds: string[], action: string) => void;
  sortBy?: string;
  sortDirection?: 'asc' | 'desc';
  onSort?: (field: string, direction: 'asc' | 'desc') => void;
  pageSize?: number;
  currentPage?: number;
  onPageChange?: (page: number) => void;
}

const UserTable: React.FC<UserTableProps> = ({ 
  users, 
  selectedUsers = [], 
  onUserSelect, 
  onUserAction, 
  onBulkAction, 
  sortBy, 
  sortDirection, 
  onSort, 
  pageSize = 50, 
  currentPage = 1, 
  onPageChange 
}) => {
  // Sortable table implementation
  // Row selection with checkboxes
  // Pagination controls
  // Responsive design
};
```

**Features**:
- Sortable columns with indicators
- Multi-select with checkboxes
- Pagination with page size options
- Responsive table scrolling

**Performance**:
- Virtual scrolling for large user lists
- Efficient sorting algorithms
- Debounced filter updates
- Memoized row rendering

---

### RoleMatrix Component

```typescript
interface Permission {
  resource: string;
  actions: string[];
  conditions?: string[];
}

interface Role {
  id: string;
  name: string;
  description: string;
  permissions: Permission[];
  isSystemRole: boolean;
  userCount: number;
}

interface RoleMatrixProps {
  roles: Role[];
  permissions: Permission[];
  onRoleChange?: (roleId: string, permissions: Permission[]) => void;
  onRoleCreate?: (role: Omit<Role, 'id'>) => void;
  onRoleDelete?: (roleId: string) => void;
  readOnly?: boolean;
}

const RoleMatrix: React.FC<RoleMatrixProps> = ({ 
  roles, 
  permissions, 
  onRoleChange, 
  onRoleCreate, 
  onRoleDelete, 
  readOnly = false 
}) => {
  // Grid layout with checkboxes
  // Drag and drop support
  // Inherited permission indication
  // Conflict detection
};
```

**Features**:
- Grid-based permission matrix
- Visual permission inheritance
- Drag and drop role assignment
- Permission conflict detection

**Accessibility**:
- Screen reader table navigation
- Keyboard checkbox controls
- Clear permission descriptions
- Role hierarchy announcements

---

## Enterprise Features Components

### TenantDashboard Component

```typescript
interface Tenant {
  id: string;
  name: string;
  domain: string;
  status: 'active' | 'trial' | 'suspended';
  userCount: number;
  resourceUsage: ResourceUsage;
  billingInfo: BillingInfo;
  settings: TenantSettings;
  createdAt: Date;
}

interface TenantDashboardProps {
  tenant: Tenant;
  onTenantUpdate?: (tenant: Partial<Tenant>) => void;
  onResourceAction?: (action: string) => void;
  onBillingAction?: (action: string) => void;
}

const TenantDashboard: React.FC<TenantDashboardProps> = ({ 
  tenant, 
  onTenantUpdate, 
  onResourceAction, 
  onBillingAction 
}) => {
  // Tenant overview cards
  // Resource usage charts
  // Billing information
  // Quick actions
};
```

**Features**:
- Tenant status overview
- Resource usage visualization
- Billing integration display
- Quick action buttons

**Customization**:
- Tenant-specific branding
- Custom color schemes
- Logo integration
- Feature flag display

---

### FederationTopology Component

```typescript
interface Cluster {
  id: string;
  name: string;
  region: string;
  status: 'online' | 'offline' | 'degraded';
  nodeCount: number;
  connections: ClusterConnection[];
  latency: { [clusterId: string]: number };
  position: { lat: number; lng: number };
}

interface FederationTopologyProps {
  clusters: Cluster[];
  selectedCluster?: string;
  onClusterSelect?: (clusterId: string) => void;
  onClusterAction?: (clusterId: string, action: string) => void;
  view?: 'map' | 'graph';
  showLatency?: boolean;
  showTraffic?: boolean;
}

const FederationTopology: React.FC<FederationTopologyProps> = ({ 
  clusters, 
  selectedCluster, 
  onClusterSelect, 
  onClusterAction, 
  view = 'map', 
  showLatency = true, 
  showTraffic = false 
}) => {
  // Geographic map visualization
  // Network graph alternative
  // Latency indicators
  // Traffic flow animation
};
```

**Features**:
- Geographic cluster distribution
- Inter-cluster latency display
- Traffic flow visualization
- Connection status indicators

**Visualization**:
- Interactive world map
- Force-directed graph option
- Animated traffic flows
- Responsive design for mobile

---

## Shared Components

### Modal Component

```typescript
interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  closeOnOverlayClick?: boolean;
  closeOnEscape?: boolean;
  showCloseButton?: boolean;
  children: React.ReactNode;
  footer?: React.ReactNode;
}

const Modal: React.FC<ModalProps> = ({ 
  isOpen, 
  onClose, 
  title, 
  size = 'md', 
  closeOnOverlayClick = true, 
  closeOnEscape = true, 
  showCloseButton = true, 
  children, 
  footer 
}) => {
  // Portal rendering
  // Focus trap implementation
  // Smooth enter/exit animations
  // Backdrop blur effect
};
```

**Features**:
- Portal rendering outside DOM tree
- Focus trap for keyboard navigation
- Smooth animations (300ms)
- Backdrop blur effect

**Accessibility**:
- Focus management (trap and restore)
- Escape key closing
- ARIA modal attributes
- Screen reader announcements

---

### Table Component

```typescript
interface Column<T> {
  key: keyof T;
  title: string;
  width?: number;
  sortable?: boolean;
  render?: (value: any, row: T, index: number) => React.ReactNode;
  align?: 'left' | 'center' | 'right';
}

interface TableProps<T> {
  data: T[];
  columns: Column<T>[];
  keyField: keyof T;
  loading?: boolean;
  emptyMessage?: string;
  sortBy?: keyof T;
  sortDirection?: 'asc' | 'desc';
  onSort?: (field: keyof T, direction: 'asc' | 'desc') => void;
  onRowClick?: (row: T, index: number) => void;
  selectedRows?: (T[keyof T])[];
  onRowSelect?: (selectedRows: (T[keyof T])[]) => void;
  pagination?: PaginationProps;
}

const Table = <T extends Record<string, any>>({ 
  data, 
  columns, 
  keyField, 
  loading = false, 
  emptyMessage = 'No data available', 
  sortBy, 
  sortDirection, 
  onSort, 
  onRowClick, 
  selectedRows = [], 
  onRowSelect, 
  pagination 
}: TableProps<T>) => {
  // Generic table implementation
  // Sorting with visual indicators
  // Row selection with checkboxes
  // Loading and empty states
};
```

**Features**:
- Generic type-safe implementation
- Sortable columns with indicators
- Row selection (single/multi)
- Loading skeleton states

**Performance**:
- Virtual scrolling for large datasets
- Memoized row rendering
- Efficient sorting algorithms
- Debounced filter updates

---

## Implementation Guidelines

### Component Development Standards

**File Structure**:
```
src/components/
├── ui/           # Basic UI components
├── charts/       # Data visualization components
├── forms/        # Form-related components
├── layout/       # Layout and navigation
├── admin/        # Admin dashboard components
├── security/     # Security management components
├── enterprise/   # Enterprise feature components
└── shared/       # Shared utility components
```

**Component Template**:
```typescript
import React from 'react';
import styled from 'styled-components';
import { ComponentProps } from './ComponentName.types';

const StyledComponent = styled.div`
  /* Styled-components implementation */
`;

export const ComponentName: React.FC<ComponentProps> = ({ 
  prop1, 
  prop2, 
  ...props 
}) => {
  // Component logic

  return (
    <StyledComponent {...props}>
      {/* Component JSX */}
    </StyledComponent>
  );
};

ComponentName.displayName = 'ComponentName';

export default ComponentName;
```

**Testing Requirements**:
- Unit tests with React Testing Library
- Accessibility tests with jest-axe
- Visual regression tests with Playwright
- Performance tests for complex components

**Documentation Standards**:
- Storybook stories for all components
- Props documentation with examples
- Accessibility guidelines
- Performance considerations

This comprehensive component specification provides the foundation for building a robust, accessible, and performant distributed Ollama frontend interface.