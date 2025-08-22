/**
 * Charts Components Export
 * Complete data visualization library with responsive charts, interactive features,
 * and comprehensive theming support
 */

// Core chart components
export { ChartContainer } from './ChartContainer'
export type { ChartContainerProps, ChartContainerRef } from './ChartContainer'

export { LineChart } from './LineChart'
export type { LineChartProps } from './LineChart'

export { AreaChart } from './AreaChart'
export type { AreaChartProps } from './AreaChart'

export { BarChart } from './BarChart'
export type { BarChartProps } from './BarChart'

export { PieChart } from './PieChart'
export type { PieChartProps } from './PieChart'

export { Heatmap } from './Heatmap'
export type { HeatmapProps, HeatmapDataPoint } from './Heatmap'

export { GaugeChart } from './GaugeChart'
export type { GaugeChartProps } from './GaugeChart'

// Chart utilities and configuration
export {
  dataFormatters,
  colorUtils,
  statsUtils,
  timeSeriesUtils,
  exportUtils,
  responsiveUtils
} from '@/utils/chartUtils'

export type {
  ChartDataPoint,
  TimeSeriesDataPoint,
  CategoryDataPoint,
  MultiSeriesDataPoint,
  ChartTheme
} from '@/utils/chartUtils'

export {
  animationConfig,
  responsiveConfig,
  accessibilityConfig,
  performanceConfig,
  themeConfig,
  chartDefaults,
  tooltipConfig,
  exportConfig,
  configUtils
} from './chartConfig'

export type {
  ResponsiveBreakpoint,
  ChartType,
  ColorPalette,
  ThemeMode
} from './chartConfig'

// Chart library information
export const chartLibraryInfo = {
  name: 'Ollama Distributed Charts',
  version: '1.0.0',
  description: 'Comprehensive data visualization library for distributed Ollama system monitoring',
  
  // Supported chart types
  chartTypes: [
    'LineChart',
    'AreaChart', 
    'BarChart',
    'PieChart',
    'Heatmap',
    'GaugeChart'
  ],
  
  // Key features
  features: [
    'Responsive design with mobile-first approach',
    'Dark/light mode support with automatic theme switching',
    'Interactive features (zoom, pan, hover, click)',
    'Real-time data streaming capabilities',
    'Export functionality (PNG, SVG, PDF, CSV)',
    'Accessibility compliance (WCAG 2.1 AA)',
    'Performance optimization for large datasets',
    'TypeScript support with comprehensive type definitions',
    'Customizable themes and color palettes',
    'Animation system with motion preference respect',
    'Comprehensive tooltip and legend systems',
    'Statistical calculations and data formatting utilities'
  ],
  
  // Browser compatibility
  browserSupport: {
    chrome: '90+',
    firefox: '88+',
    safari: '14+',
    edge: '90+'
  },
  
  // Dependencies
  dependencies: {
    recharts: '^2.15.4',
    'date-fns': '^4.1.0',
    jspdf: '^2.5.2',
    react: '^19.1.1'
  }
} as const

// Convenience exports for common chart patterns
export const ChartPatterns = {
  /**
   * Create a time series monitoring chart
   */
  createTimeSeriesChart: (data: any[], metrics: string[]) => ({
    component: 'LineChart',
    props: {
      data,
      metrics: metrics.map(key => ({
        key,
        name: key.charAt(0).toUpperCase() + key.slice(1),
        format: 'default' as const
      })),
      xAxis: {
        dataKey: 'timestamp',
        format: 'timestamp' as const,
        timeRange: 'hour' as const
      },
      interactive: {
        zoom: true,
        brush: true,
        crosshair: true
      }
    }
  }),
  
  /**
   * Create a resource utilization area chart
   */
  createResourceChart: (data: any[], resources: string[]) => ({
    component: 'AreaChart',
    props: {
      data,
      areas: resources.map(key => ({
        key,
        name: key.charAt(0).toUpperCase() + key.slice(1),
        format: 'percentage' as const,
        stackId: 'resources'
      })),
      type: 'stacked' as const,
      yAxis: {
        domain: [0, 100],
        format: 'percentage' as const
      }
    }
  }),
  
  /**
   * Create a distribution pie chart
   */
  createDistributionChart: (data: any[]) => ({
    component: 'PieChart',
    props: {
      data,
      variant: 'donut' as const,
      labels: {
        show: true,
        position: 'outside' as const,
        type: 'keyPercent' as const
      },
      legend: {
        show: true,
        position: 'bottom' as const
      }
    }
  }),
  
  /**
   * Create a performance gauge
   */
  createPerformanceGauge: (value: number, thresholds: any[]) => ({
    component: 'GaugeChart',
    props: {
      value,
      min: 0,
      max: 100,
      thresholds,
      style: 'half' as const,
      valueFormat: {
        format: 'percentage' as const
      },
      appearance: {
        showValue: true,
        showThresholds: true
      }
    }
  })
} as const

// Export helper functions for chart creation
export const ChartHelpers = {
  /**
   * Generate mock time series data for testing
   */
  generateMockTimeSeriesData: (
    metrics: string[],
    points: number = 24,
    timeRange: 'hour' | 'day' | 'week' = 'hour'
  ) => {
    const data: any[] = []
    const now = new Date()
    
    for (let i = 0; i < points; i++) {
      const timestamp = new Date(now)
      
      switch (timeRange) {
        case 'hour':
          timestamp.setMinutes(timestamp.getMinutes() - (points - i) * 5)
          break
        case 'day':
          timestamp.setHours(timestamp.getHours() - (points - i))
          break
        case 'week':
          timestamp.setDate(timestamp.getDate() - (points - i))
          break
      }
      
      const point: any = { timestamp: timestamp.toISOString() }
      
      metrics.forEach(metric => {
        point[metric] = Math.random() * 100
      })
      
      data.push(point)
    }
    
    return data
  },
  
  /**
   * Generate mock categorical data for testing
   */
  generateMockCategoryData: (categories: string[]) => {
    return categories.map(category => ({
      label: category,
      value: Math.random() * 100,
      category
    }))
  },
  
  /**
   * Generate mock heatmap data for testing
   */
  generateMockHeatmapData: (xLabels: string[], yLabels: string[]) => {
    const data: any[] = []
    
    xLabels.forEach((x, xi) => {
      yLabels.forEach((y, yi) => {
        data.push({
          x: xi,
          y: yi,
          value: Math.random() * 100,
          label: `${x} - ${y}`
        })
      })
    })
    
    return data
  }
} as const

export default {
  chartLibraryInfo,
  ChartPatterns,
  ChartHelpers
}