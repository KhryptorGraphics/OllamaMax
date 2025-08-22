/**
 * Reusable Chart Component System
 * Provides a unified interface for different chart types with TypeScript support
 */

import React, { useMemo, useCallback } from 'react'
import {
  LineChart,
  AreaChart,
  BarChart,
  PieChart,
  ScatterChart,
  ComposedChart,
  Line,
  Area,
  Bar,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine,
  Brush,
  Scatter
} from 'recharts'
import { ChartType, WidgetConfig } from '../../types'

// Chart configuration interfaces
export interface ChartProps {
  type: ChartType
  data: any[]
  config?: ChartConfig
  width?: number | string
  height?: number | string
  responsive?: boolean
  onDataPointClick?: (data: any, index: number) => void
  onLegendClick?: (data: any) => void
  loading?: boolean
  error?: string
  className?: string
}

export interface ChartConfig {
  // Data configuration
  xKey?: string
  yKey?: string | string[]
  groupBy?: string
  
  // Styling
  colors?: string[]
  colorScheme?: 'default' | 'blue' | 'green' | 'red' | 'purple' | 'orange' | 'custom'
  theme?: 'light' | 'dark'
  
  // Display options
  showGrid?: boolean
  showLegend?: boolean
  showTooltip?: boolean
  showAxis?: boolean
  showBrush?: boolean
  
  // Chart-specific options
  stacked?: boolean
  smooth?: boolean
  curved?: boolean
  fill?: boolean
  gradient?: boolean
  
  // Axes configuration
  xAxisLabel?: string
  yAxisLabel?: string
  xAxisFormatter?: (value: any) => string
  yAxisFormatter?: (value: any) => string
  
  // Tooltip configuration
  tooltipFormatter?: (value: any, name: string, props: any) => [string, string]
  labelFormatter?: (value: any) => string
  
  // Reference lines
  referenceLines?: ReferenceLine[]
  
  // Animation
  animationDuration?: number
  animationEasing?: string
  
  // Custom styling
  margin?: { top: number; right: number; bottom: number; left: number }
  padding?: string
}

interface ReferenceLine {
  value: number
  label?: string
  color?: string
  strokeDasharray?: string
}

// Color schemes
const COLOR_SCHEMES = {
  default: ['#8884d8', '#82ca9d', '#ffc658', '#ff7300', '#00ff00', '#ff00ff'],
  blue: ['#0066cc', '#3385d6', '#66a3e0', '#99c2ea', '#cce0f4'],
  green: ['#00b894', '#00d2a0', '#55efc4', '#81ecec', '#a7f3d0'],
  red: ['#e74c3c', '#f39c12', '#f1c40f', '#e67e22', '#d63031'],
  purple: ['#6c5ce7', '#a29bfe', '#fd79a8', '#fdcb6e', '#e84393'],
  orange: ['#fd7f6f', '#ffb347', '#ffcc5c', '#ff6b6b', '#4ecdc4']
}

// Default configuration
const DEFAULT_CONFIG: ChartConfig = {
  xKey: 'x',
  yKey: 'y',
  colorScheme: 'default',
  theme: 'light',
  showGrid: true,
  showLegend: true,
  showTooltip: true,
  showAxis: true,
  animationDuration: 300,
  margin: { top: 20, right: 30, bottom: 20, left: 20 }
}

export const Chart: React.FC<ChartProps> = ({
  type,
  data,
  config = {},
  width = '100%',
  height = 300,
  responsive = true,
  onDataPointClick,
  onLegendClick,
  loading = false,
  error,
  className = ''
}) => {
  const mergedConfig = useMemo(() => ({
    ...DEFAULT_CONFIG,
    ...config
  }), [config])

  const colors = useMemo(() => {
    if (mergedConfig.colors) return mergedConfig.colors
    if (mergedConfig.colorScheme && COLOR_SCHEMES[mergedConfig.colorScheme]) {
      return COLOR_SCHEMES[mergedConfig.colorScheme]
    }
    return COLOR_SCHEMES.default
  }, [mergedConfig.colors, mergedConfig.colorScheme])

  const handleClick = useCallback((data: any, index: number) => {
    onDataPointClick?.(data, index)
  }, [onDataPointClick])

  const renderTooltip = useCallback((props: any) => {
    if (!mergedConfig.showTooltip) return null

    const { active, payload, label } = props
    if (!active || !payload || !payload.length) return null

    return (
      <div className="bg-white dark:bg-gray-800 p-3 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg">
        {label && (
          <p className="font-medium text-gray-900 dark:text-gray-100 mb-2">
            {mergedConfig.labelFormatter ? mergedConfig.labelFormatter(label) : label}
          </p>
        )}
        {payload.map((entry: any, index: number) => (
          <p key={index} style={{ color: entry.color }} className="text-sm">
            {`${entry.name}: ${
              mergedConfig.tooltipFormatter
                ? mergedConfig.tooltipFormatter(entry.value, entry.name, entry)[0]
                : entry.value
            }`}
          </p>
        ))}
      </div>
    )
  }, [mergedConfig])

  const renderXAxis = () => {
    if (!mergedConfig.showAxis) return null

    return (
      <XAxis
        dataKey={mergedConfig.xKey}
        axisLine={false}
        tickLine={false}
        tick={{ fontSize: 12 }}
        tickFormatter={mergedConfig.xAxisFormatter}
      />
    )
  }

  const renderYAxis = () => {
    if (!mergedConfig.showAxis) return null

    return (
      <YAxis
        axisLine={false}
        tickLine={false}
        tick={{ fontSize: 12 }}
        tickFormatter={mergedConfig.yAxisFormatter}
      />
    )
  }

  const renderGrid = () => {
    if (!mergedConfig.showGrid) return null
    return <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
  }

  const renderLegend = () => {
    if (!mergedConfig.showLegend) return null
    return <Legend onClick={onLegendClick} />
  }

  const renderReferenceLines = () => {
    if (!mergedConfig.referenceLines) return null

    return mergedConfig.referenceLines.map((line, index) => (
      <ReferenceLine
        key={index}
        y={line.value}
        stroke={line.color || '#666'}
        strokeDasharray={line.strokeDasharray || '3 3'}
        label={line.label}
      />
    ))
  }

  const renderChart = () => {
    const commonProps = {
      data,
      margin: mergedConfig.margin,
      onClick: handleClick
    }

    switch (type) {
      case 'line':
        return (
          <LineChart {...commonProps}>
            {renderGrid()}
            {renderXAxis()}
            {renderYAxis()}
            <Tooltip content={renderTooltip} />
            {renderLegend()}
            {renderReferenceLines()}
            {Array.isArray(mergedConfig.yKey) ? (
              mergedConfig.yKey.map((key, index) => (
                <Line
                  key={key}
                  type={mergedConfig.smooth ? 'monotone' : 'linear'}
                  dataKey={key}
                  stroke={colors[index % colors.length]}
                  strokeWidth={2}
                  dot={{ r: 3 }}
                  activeDot={{ r: 5 }}
                  animationDuration={mergedConfig.animationDuration}
                />
              ))
            ) : (
              <Line
                type={mergedConfig.smooth ? 'monotone' : 'linear'}
                dataKey={mergedConfig.yKey}
                stroke={colors[0]}
                strokeWidth={2}
                dot={{ r: 3 }}
                activeDot={{ r: 5 }}
                animationDuration={mergedConfig.animationDuration}
              />
            )}
            {mergedConfig.showBrush && <Brush />}
          </LineChart>
        )

      case 'area':
        return (
          <AreaChart {...commonProps}>
            {renderGrid()}
            {renderXAxis()}
            {renderYAxis()}
            <Tooltip content={renderTooltip} />
            {renderLegend()}
            {renderReferenceLines()}
            {Array.isArray(mergedConfig.yKey) ? (
              mergedConfig.yKey.map((key, index) => (
                <Area
                  key={key}
                  type={mergedConfig.smooth ? 'monotone' : 'linear'}
                  dataKey={key}
                  stackId={mergedConfig.stacked ? '1' : undefined}
                  stroke={colors[index % colors.length]}
                  fill={colors[index % colors.length]}
                  fillOpacity={0.6}
                  animationDuration={mergedConfig.animationDuration}
                />
              ))
            ) : (
              <Area
                type={mergedConfig.smooth ? 'monotone' : 'linear'}
                dataKey={mergedConfig.yKey}
                stroke={colors[0]}
                fill={colors[0]}
                fillOpacity={0.6}
                animationDuration={mergedConfig.animationDuration}
              />
            )}
          </AreaChart>
        )

      case 'bar':
      case 'column':
        return (
          <BarChart {...commonProps}>
            {renderGrid()}
            {renderXAxis()}
            {renderYAxis()}
            <Tooltip content={renderTooltip} />
            {renderLegend()}
            {renderReferenceLines()}
            {Array.isArray(mergedConfig.yKey) ? (
              mergedConfig.yKey.map((key, index) => (
                <Bar
                  key={key}
                  dataKey={key}
                  stackId={mergedConfig.stacked ? '1' : undefined}
                  fill={colors[index % colors.length]}
                  animationDuration={mergedConfig.animationDuration}
                />
              ))
            ) : (
              <Bar
                dataKey={mergedConfig.yKey}
                fill={colors[0]}
                animationDuration={mergedConfig.animationDuration}
              />
            )}
          </BarChart>
        )

      case 'pie':
      case 'donut':
        return (
          <PieChart {...commonProps}>
            <Tooltip content={renderTooltip} />
            {renderLegend()}
            <Pie
              data={data}
              dataKey={mergedConfig.yKey}
              nameKey={mergedConfig.xKey}
              cx="50%"
              cy="50%"
              outerRadius={type === 'donut' ? 80 : 100}
              innerRadius={type === 'donut' ? 40 : 0}
              fill="#8884d8"
              animationDuration={mergedConfig.animationDuration}
              label={mergedConfig.showLegend ? false : true}
            >
              {data.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
              ))}
            </Pie>
          </PieChart>
        )

      case 'scatter':
        return (
          <ScatterChart {...commonProps}>
            {renderGrid()}
            {renderXAxis()}
            {renderYAxis()}
            <Tooltip content={renderTooltip} />
            {renderLegend()}
            <Scatter
              dataKey={mergedConfig.yKey}
              fill={colors[0]}
              animationDuration={mergedConfig.animationDuration}
            />
          </ScatterChart>
        )

      default:
        return <div>Unsupported chart type: {type}</div>
    }
  }

  if (loading) {
    return (
      <div className={`flex items-center justify-center ${className}`} style={{ height }}>
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className={`flex items-center justify-center ${className}`} style={{ height }}>
        <div className="text-red-500 text-center">
          <p className="font-medium">Error loading chart</p>
          <p className="text-sm">{error}</p>
        </div>
      </div>
    )
  }

  if (!data || data.length === 0) {
    return (
      <div className={`flex items-center justify-center ${className}`} style={{ height }}>
        <div className="text-gray-500 text-center">
          <p>No data available</p>
        </div>
      </div>
    )
  }

  const ContainerComponent = responsive ? ResponsiveContainer : 'div'
  const containerProps = responsive 
    ? { width: '100%', height } 
    : { style: { width, height } }

  return (
    <div className={`chart-container ${className}`}>
      <ContainerComponent {...containerProps}>
        {renderChart()}
      </ContainerComponent>
    </div>
  )
}

// Specialized chart components
export const LineChartComponent: React.FC<Omit<ChartProps, 'type'>> = (props) => (
  <Chart {...props} type="line" />
)

export const AreaChartComponent: React.FC<Omit<ChartProps, 'type'>> = (props) => (
  <Chart {...props} type="area" />
)

export const BarChartComponent: React.FC<Omit<ChartProps, 'type'>> = (props) => (
  <Chart {...props} type="bar" />
)

export const PieChartComponent: React.FC<Omit<ChartProps, 'type'>> = (props) => (
  <Chart {...props} type="pie" />
)

export const ScatterChartComponent: React.FC<Omit<ChartProps, 'type'>> = (props) => (
  <Chart {...props} type="scatter" />
)

export default Chart