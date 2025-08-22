/**
 * LineChart Component
 * Time series data visualization with multiple metrics support,
 * interactive features, and responsive design
 */

import React, { useMemo } from 'react'
import {
  LineChart as RechartsLineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  ReferenceLine,
  Brush,
  Area
} from 'recharts'
import { 
  dataFormatters, 
  colorUtils,
  responsiveUtils,
  type ChartTheme, 
  type TimeSeriesDataPoint,
  type MultiSeriesDataPoint 
} from '@/utils/chartUtils'

export interface LineChartProps {
  /** Chart data */
  data: MultiSeriesDataPoint[]
  
  /** Metrics to display as lines */
  metrics: Array<{
    key: string
    name: string
    color?: string
    strokeWidth?: number
    strokeDashArray?: string
    hide?: boolean
    yAxisId?: 'left' | 'right'
    format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
    unit?: string
  }>
  
  /** Chart theme */
  theme?: ChartTheme
  
  /** Chart dimensions */
  dimensions?: { width: number; height: number }
  
  /** Enable animations */
  animations?: boolean
  
  /** X-axis configuration */
  xAxis?: {
    dataKey?: string
    format?: 'timestamp' | 'label'
    timeRange?: 'minute' | 'hour' | 'day' | 'week' | 'month'
    tickCount?: number
  }
  
  /** Y-axis configuration */
  yAxis?: {
    left?: {
      domain?: [number | 'auto', number | 'auto']
      tickCount?: number
      format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
      unit?: string
    }
    right?: {
      domain?: [number | 'auto', number | 'auto']
      tickCount?: number
      format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
      unit?: string
    }
  }
  
  /** Interactive features */
  interactive?: {
    zoom?: boolean
    brush?: boolean
    crosshair?: boolean
    clickable?: boolean
  }
  
  /** Reference lines */
  referenceLines?: Array<{
    value: number
    label?: string
    color?: string
    strokeDashArray?: string
    yAxisId?: 'left' | 'right'
  }>
  
  /** Tooltip configuration */
  tooltip?: {
    formatter?: (value: any, name: string, props: any) => [string, string]
    labelFormatter?: (label: string) => string
    show?: boolean
  }
  
  /** Legend configuration */
  legend?: {
    show?: boolean
    position?: 'top' | 'bottom' | 'left' | 'right'
  }
  
  /** Grid configuration */
  grid?: {
    show?: boolean
    strokeDashArray?: string
  }
  
  /** Margin configuration */
  margin?: {
    top?: number
    right?: number
    bottom?: number
    left?: number
  }
  
  /** Loading state */
  loading?: boolean
  
  /** Empty state message */
  emptyMessage?: string
  
  /** Click handler for data points */
  onDataPointClick?: (data: any, index: number) => void
  
  /** Hover handler for data points */
  onDataPointHover?: (data: any, index: number) => void
}

const LineChart: React.FC<LineChartProps> = ({
  data,
  metrics,
  theme,
  dimensions,
  animations = true,
  xAxis = {
    dataKey: 'timestamp',
    format: 'timestamp',
    timeRange: 'hour'
  },
  yAxis = {},
  interactive = {
    zoom: false,
    brush: false,
    crosshair: true,
    clickable: false
  },
  referenceLines = [],
  tooltip = { show: true },
  legend = { show: true, position: 'bottom' },
  grid = { show: true, strokeDashArray: '3 3' },
  margin,
  loading = false,
  emptyMessage = 'No data available',
  onDataPointClick,
  onDataPointHover
}) => {
  // Generate responsive margins if not provided
  const chartMargin = useMemo(() => {
    if (margin) return margin
    return responsiveUtils.getResponsiveMargins()
  }, [margin])

  // Generate colors for metrics
  const metricsWithColors = useMemo(() => {
    if (!theme) return metrics
    
    const colors = colorUtils.generateColorPalette(metrics.length, theme)
    return metrics.map((metric, index) => ({
      ...metric,
      color: metric.color || colors[index]
    }))
  }, [metrics, theme])

  // Filter visible metrics
  const visibleMetrics = useMemo(() => 
    metricsWithColors.filter(metric => !metric.hide),
    [metricsWithColors]
  )

  // Custom tooltip component
  const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload || !payload.length) return null

    return (
      <div className="rounded-lg border bg-popover p-3 shadow-lg">
        <div className="mb-2 text-sm font-medium">
          {tooltip.labelFormatter 
            ? tooltip.labelFormatter(label)
            : xAxis.format === 'timestamp' 
              ? dataFormatters.formatTimestamp(label, xAxis.timeRange)
              : label
          }
        </div>
        <div className="space-y-1">
          {payload.map((item: any, index: number) => {
            const metric = metricsWithColors.find(m => m.key === item.dataKey)
            if (!metric) return null

            const formattedValue = tooltip.formatter
              ? tooltip.formatter(item.value, item.name || metric.name, item)[0]
              : dataFormatters.formatNumber(item.value, {
                  format: metric.format,
                  unit: metric.unit,
                  precision: 2
                })

            return (
              <div key={index} className="flex items-center gap-2 text-sm">
                <div
                  className="h-3 w-3 rounded-sm"
                  style={{ backgroundColor: item.color }}
                />
                <span className="font-medium">{item.name || metric.name}:</span>
                <span>{formattedValue}</span>
              </div>
            )
          })}
        </div>
      </div>
    )
  }

  // Custom Y-axis tick formatter
  const formatYAxisTick = (value: any, axisId: 'left' | 'right') => {
    const config = yAxis[axisId]
    if (!config) return value

    return dataFormatters.formatNumber(value, {
      format: config.format,
      unit: config.unit,
      precision: 1
    })
  }

  // Custom X-axis tick formatter
  const formatXAxisTick = (value: any) => {
    if (xAxis.format === 'timestamp') {
      return dataFormatters.formatTimestamp(value, xAxis.timeRange)
    }
    return value
  }

  // Handle data point clicks
  const handleDataPointClick = (data: any, index: number) => {
    if (interactive.clickable && onDataPointClick) {
      onDataPointClick(data, index)
    }
  }

  // Loading or empty state
  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-sm text-muted-foreground">Loading...</div>
      </div>
    )
  }

  if (!data || data.length === 0) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-sm text-muted-foreground">{emptyMessage}</div>
      </div>
    )
  }

  return (
    <div className="h-full w-full">
      <ResponsiveContainer width="100%" height="100%">
        <RechartsLineChart
          data={data}
          margin={chartMargin}
          onClick={handleDataPointClick}
        >
          {/* Grid */}
          {grid.show && (
            <CartesianGrid
              strokeDasharray={grid.strokeDashArray}
              stroke={theme?.colors.grid}
              opacity={0.3}
            />
          )}

          {/* X Axis */}
          <XAxis
            dataKey={xAxis.dataKey}
            tickFormatter={formatXAxisTick}
            tick={{ 
              fontSize: 12, 
              fill: theme?.colors.axis 
            }}
            axisLine={{ stroke: theme?.colors.grid }}
            tickLine={{ stroke: theme?.colors.grid }}
            tickCount={xAxis.tickCount}
          />

          {/* Left Y Axis */}
          <YAxis
            yAxisId="left"
            orientation="left"
            domain={yAxis.left?.domain}
            tickCount={yAxis.left?.tickCount}
            tickFormatter={(value) => formatYAxisTick(value, 'left')}
            tick={{ 
              fontSize: 12, 
              fill: theme?.colors.axis 
            }}
            axisLine={{ stroke: theme?.colors.grid }}
            tickLine={{ stroke: theme?.colors.grid }}
          />

          {/* Right Y Axis (if needed) */}
          {visibleMetrics.some(m => m.yAxisId === 'right') && (
            <YAxis
              yAxisId="right"
              orientation="right"
              domain={yAxis.right?.domain}
              tickCount={yAxis.right?.tickCount}
              tickFormatter={(value) => formatYAxisTick(value, 'right')}
              tick={{ 
                fontSize: 12, 
                fill: theme?.colors.axis 
              }}
              axisLine={{ stroke: theme?.colors.grid }}
              tickLine={{ stroke: theme?.colors.grid }}
            />
          )}

          {/* Tooltip */}
          {tooltip.show && (
            <Tooltip
              content={<CustomTooltip />}
              cursor={
                interactive.crosshair
                  ? {
                      stroke: theme?.colors.grid,
                      strokeWidth: 1,
                      strokeDasharray: '5 5'
                    }
                  : false
              }
            />
          )}

          {/* Legend */}
          {legend.show && (
            <Legend
              verticalAlign={
                legend.position === 'top' || legend.position === 'bottom'
                  ? legend.position
                  : 'bottom'
              }
              align={
                legend.position === 'left' || legend.position === 'right'
                  ? legend.position
                  : 'center'
              }
              wrapperStyle={{
                color: theme?.colors.text,
                fontSize: '12px'
              }}
            />
          )}

          {/* Reference Lines */}
          {referenceLines.map((refLine, index) => (
            <ReferenceLine
              key={index}
              y={refLine.value}
              yAxisId={refLine.yAxisId || 'left'}
              stroke={refLine.color || theme?.colors.semantic.warning}
              strokeDasharray={refLine.strokeDashArray || '5 5'}
              label={{
                value: refLine.label,
                position: 'topRight',
                style: { fill: theme?.colors.text, fontSize: '12px' }
              }}
            />
          ))}

          {/* Lines */}
          {visibleMetrics.map((metric) => (
            <Line
              key={metric.key}
              type="monotone"
              dataKey={metric.key}
              name={metric.name}
              stroke={metric.color}
              strokeWidth={metric.strokeWidth || 2}
              strokeDasharray={metric.strokeDashArray}
              yAxisId={metric.yAxisId || 'left'}
              dot={interactive.clickable ? { r: 4 } : false}
              activeDot={{
                r: 6,
                stroke: metric.color,
                strokeWidth: 2,
                fill: theme?.colors.background
              }}
              animationDuration={animations ? 1000 : 0}
              connectNulls={false}
            />
          ))}

          {/* Brush for zoom */}
          {interactive.brush && (
            <Brush
              dataKey={xAxis.dataKey}
              height={30}
              stroke={theme?.colors.primary[0]}
              fill={theme?.colors.background}
              tickFormatter={formatXAxisTick}
            />
          )}
        </RechartsLineChart>
      </ResponsiveContainer>
    </div>
  )
}

export { LineChart }
export type { LineChartProps }