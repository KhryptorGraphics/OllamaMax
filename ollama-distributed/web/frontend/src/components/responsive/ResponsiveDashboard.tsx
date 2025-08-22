import React, { useState, useEffect } from 'react';
import { ResponsiveCard, CardAction } from './ResponsiveCard';
import { 
  Activity, 
  Cpu, 
  HardDrive, 
  Users, 
  TrendingUp, 
  AlertTriangle,
  CheckCircle,
  Clock,
  Zap,
  Database,
  Network,
  Settings,
  Eye,
  Download,
  Share
} from 'lucide-react';

interface MetricCardData {
  id: string;
  title: string;
  value: string | number;
  change?: {
    value: number;
    direction: 'up' | 'down';
    timeframe: string;
  };
  status?: 'healthy' | 'warning' | 'critical';
  icon: React.ComponentType<any>;
  description?: string;
}

interface ResponsiveDashboardProps {
  metrics?: MetricCardData[];
  onCardClick?: (id: string) => void;
  onRefresh?: () => void;
  className?: string;
}

const defaultMetrics: MetricCardData[] = [
  {
    id: 'cluster-status',
    title: 'Cluster Status',
    value: 'Online',
    status: 'healthy',
    icon: CheckCircle,
    description: '3 nodes active',
    change: { value: 0, direction: 'up', timeframe: '24h' }
  },
  {
    id: 'active-models',
    title: 'Active Models',
    value: 12,
    icon: Cpu,
    description: '8 inference tasks running',
    change: { value: 2, direction: 'up', timeframe: '1h' }
  },
  {
    id: 'cpu-usage',
    title: 'CPU Usage',
    value: '68%',
    status: 'warning',
    icon: Activity,
    description: 'Across 3 nodes',
    change: { value: 5, direction: 'up', timeframe: '5m' }
  },
  {
    id: 'memory-usage',
    title: 'Memory Usage',
    value: '4.2GB',
    icon: HardDrive,
    description: 'of 16GB total',
    change: { value: 200, direction: 'up', timeframe: '5m' }
  },
  {
    id: 'active-users',
    title: 'Active Users',
    value: 24,
    icon: Users,
    description: 'Current sessions',
    change: { value: 3, direction: 'up', timeframe: '30m' }
  },
  {
    id: 'requests-per-second',
    title: 'Requests/sec',
    value: '1.2K',
    icon: TrendingUp,
    description: 'Average response time: 120ms',
    change: { value: 150, direction: 'up', timeframe: '5m' }
  },
  {
    id: 'storage-used',
    title: 'Storage Used',
    value: '2.8TB',
    status: 'warning',
    icon: Database,
    description: 'of 5TB total (56%)',
    change: { value: 50, direction: 'up', timeframe: '24h' }
  },
  {
    id: 'network-throughput',
    title: 'Network I/O',
    value: '45 Mbps',
    icon: Network,
    description: 'Ingress/Egress combined',
    change: { value: 8, direction: 'down', timeframe: '5m' }
  }
];

export const ResponsiveDashboard: React.FC<ResponsiveDashboardProps> = ({
  metrics = defaultMetrics,
  onCardClick,
  onRefresh,
  className = ''
}) => {
  const [refreshing, setRefreshing] = useState(false);
  const [selectedMetrics, setSelectedMetrics] = useState<Set<string>>(new Set());
  const [isSelectionMode, setIsSelectionMode] = useState(false);

  // Handle refresh with loading state
  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await onRefresh?.();
      // Simulate network delay for UX
      await new Promise(resolve => setTimeout(resolve, 1000));
    } finally {
      setRefreshing(false);
    }
  };

  // Handle card selection (for batch operations)
  const handleCardSelection = (id: string) => {
    const newSelected = new Set(selectedMetrics);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedMetrics(newSelected);
  };

  // Handle long press to enter selection mode
  const handleLongPress = (id: string) => {
    setIsSelectionMode(true);
    setSelectedMetrics(new Set([id]));
  };

  // Exit selection mode
  const exitSelectionMode = () => {
    setIsSelectionMode(false);
    setSelectedMetrics(new Set());
  };

  // Get status color
  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'healthy': return 'text-green-600';
      case 'warning': return 'text-yellow-600';
      case 'critical': return 'text-red-600';
      default: return 'text-gray-600';
    }
  };

  // Get status background
  const getStatusBg = (status?: string) => {
    switch (status) {
      case 'healthy': return 'bg-green-50 border-green-200';
      case 'warning': return 'bg-yellow-50 border-yellow-200';
      case 'critical': return 'bg-red-50 border-red-200';
      default: return 'bg-white border-gray-200';
    }
  };

  // Format change value
  const formatChange = (change?: MetricCardData['change']) => {
    if (!change) return null;
    
    const sign = change.direction === 'up' ? '+' : '-';
    const color = change.direction === 'up' ? 'text-green-600' : 'text-red-600';
    
    return (
      <span className={`text-xs ${color} flex items-center space-x-1`}>
        <span>{sign}{Math.abs(change.value)}</span>
        <span className="text-gray-400">({change.timeframe})</span>
      </span>
    );
  };

  return (
    <div className={`responsive-dashboard ${className}`}>
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6 space-y-4 sm:space-y-0">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-sm text-gray-600">Real-time system metrics and status</p>
        </div>
        
        <div className="flex items-center space-x-3">
          {isSelectionMode && (
            <div className="flex items-center space-x-2 bg-blue-50 px-3 py-2 rounded-lg">
              <span className="text-sm text-blue-700">
                {selectedMetrics.size} selected
              </span>
              <button
                onClick={exitSelectionMode}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Cancel
              </button>
            </div>
          )}
          
          <button
            onClick={handleRefresh}
            disabled={refreshing}
            className="inline-flex items-center px-4 py-2 bg-white border border-gray-300 rounded-lg shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 transition-colors"
          >
            <Clock className={`w-4 h-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            {refreshing ? 'Refreshing...' : 'Refresh'}
          </button>
        </div>
      </div>

      {/* Metrics Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 sm:gap-6">
        {metrics.map((metric) => {
          const Icon = metric.icon;
          const isSelected = selectedMetrics.has(metric.id);
          
          return (
            <ResponsiveCard
              key={metric.id}
              title={metric.title}
              subtitle={metric.description}
              onClick={() => {
                if (isSelectionMode) {
                  handleCardSelection(metric.id);
                } else {
                  onCardClick?.(metric.id);
                }
              }}
              onSwipeRight={() => handleLongPress(metric.id)}
              onSwipeLeft={() => onCardClick?.(metric.id)}
              className={`
                ${getStatusBg(metric.status)}
                ${isSelected ? 'ring-2 ring-blue-500 ring-offset-2' : ''}
                ${isSelectionMode ? 'cursor-pointer' : ''}
              `}
              isClickable={true}
              variant="elevated"
              size="medium"
              actions={
                <div className="py-1">
                  <CardAction
                    icon={Eye}
                    label="View Details"
                    onClick={() => onCardClick?.(metric.id)}
                  />
                  <CardAction
                    icon={Download}
                    label="Export Data"
                    onClick={() => console.log('Export', metric.id)}
                  />
                  <CardAction
                    icon={Share}
                    label="Share Metric"
                    onClick={() => console.log('Share', metric.id)}
                  />
                  <CardAction
                    icon={Settings}
                    label="Configure"
                    onClick={() => console.log('Configure', metric.id)}
                  />
                </div>
              }
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className={`p-2 rounded-lg ${metric.status === 'critical' ? 'bg-red-100' : metric.status === 'warning' ? 'bg-yellow-100' : 'bg-blue-100'}`}>
                    <Icon className={`w-5 h-5 ${getStatusColor(metric.status)}`} />
                  </div>
                  
                  <div>
                    <div className="text-2xl font-bold text-gray-900">
                      {metric.value}
                    </div>
                    {formatChange(metric.change)}
                  </div>
                </div>

                {/* Status Indicator */}
                {metric.status && (
                  <div className="flex items-center">
                    {metric.status === 'healthy' && <CheckCircle className="w-5 h-5 text-green-500" />}
                    {metric.status === 'warning' && <AlertTriangle className="w-5 h-5 text-yellow-500" />}
                    {metric.status === 'critical' && <AlertTriangle className="w-5 h-5 text-red-500" />}
                  </div>
                )}
              </div>

              {/* Additional Context for Mobile */}
              <div className="mt-3 pt-3 border-t border-gray-100 sm:hidden">
                <div className="flex items-center justify-between text-xs text-gray-500">
                  <span>Last updated: just now</span>
                  {metric.change && (
                    <span className="flex items-center space-x-1">
                      <Zap className="w-3 h-3" />
                      <span>Real-time</span>
                    </span>
                  )}
                </div>
              </div>
            </ResponsiveCard>
          );
        })}
      </div>

      {/* Selection Mode Actions */}
      {isSelectionMode && selectedMetrics.size > 0 && (
        <div className="fixed bottom-4 left-4 right-4 sm:bottom-6 sm:left-6 sm:right-6 bg-white border border-gray-200 rounded-lg shadow-lg p-4 z-40">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-gray-900">
              {selectedMetrics.size} metric{selectedMetrics.size > 1 ? 's' : ''} selected
            </span>
            
            <div className="flex items-center space-x-2">
              <button className="inline-flex items-center px-3 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors">
                <Download className="w-4 h-4 mr-2" />
                Export
              </button>
              
              <button className="inline-flex items-center px-3 py-2 bg-gray-100 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-200 transition-colors">
                <Share className="w-4 h-4 mr-2" />
                Share
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Pull-to-Refresh Indicator */}
      {refreshing && (
        <div className="fixed top-16 left-1/2 transform -translate-x-1/2 bg-white border border-gray-200 rounded-lg shadow-lg px-4 py-2 z-50">
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
            <span className="text-sm text-gray-600">Refreshing data...</span>
          </div>
        </div>
      )}
    </div>
  );
};