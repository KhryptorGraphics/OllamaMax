/**
 * AreaChart Component
 * Stacked area charts for resource usage visualization with gradient fills
 * and interactive features
 */

import React, { useMemo } from 'react'
import {
  AreaChart as RechartsAreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  ReferenceLine,
  Brush
} from 'recharts'
import { 
  dataFormatters, 
  colorUtils,
  responsiveUtils,
  type ChartTheme, 
  type MultiSeriesDataPoint 
} from '@/utils/chartUtils'

export interface AreaChartProps {
  /** Chart data */
  data: MultiSeriesDataPoint[]
  
  /** Areas to display */
  areas: Array<{
    key: string
    name: string
    color?: string
    stackId?: string
    hide?: boolean
    fillOpacity?: number
    strokeWidth?: number
    gradient?: boolean
    format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
    unit?: string
  }>
  
  /** Chart theme */
  theme?: ChartTheme
  
  /** Chart dimensions */
  dimensions?: { width: number; height: number }
  
  /** Enable animations */
  animations?: boolean
  
  /** Chart type */
  type?: 'stacked' | 'overlapping' | 'percent'
  
  /** X-axis configuration */
  xAxis?: {
    dataKey?: string
    format?: 'timestamp' | 'label'
    timeRange?: 'minute' | 'hour' | 'day' | 'week' | 'month'
    tickCount?: number
  }
  
  /** Y-axis configuration */
  yAxis?: {
    domain?: [number | 'auto', number | 'auto']
    tickCount?: number
    format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
    unit?: string
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
  }>
  
  /** Tooltip configuration */
  tooltip?: {
    formatter?: (value: any, name: string, props: any) => [string, string]
    labelFormatter?: (label: string) => string
    show?: boolean
    shared?: boolean
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
}

const AreaChart: React.FC<AreaChartProps> = ({
  data,
  areas,
  theme,
  dimensions,
  animations = true,
  type = 'stacked',
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
  tooltip = { show: true, shared: true },
  legend = { show: true, position: 'bottom' },
  grid = { show: true, strokeDashArray: '3 3' },
  margin,
  loading = false,
  emptyMessage = 'No data available',
  onDataPointClick
}) => {
  // Generate responsive margins if not provided
  const chartMargin = useMemo(() => {
    if (margin) return margin
    return responsiveUtils.getResponsiveMargins()
  }, [margin])

  // Generate colors for areas
  const areasWithColors = useMemo(() => {
    if (!theme) return areas
    
    const colors = colorUtils.generateColorPalette(areas.length, theme)
    return areas.map((area, index) => ({
      ...area,
      color: area.color || colors[index]
    }))
  }, [areas, theme])

  // Filter visible areas
  const visibleAreas = useMemo(() => 
    areasWithColors.filter(area => !area.hide),
    [areasWithColors]
  )

  // Generate gradient definitions
  const gradientDefs = useMemo(() => {
    return visibleAreas
      .filter(area => area.gradient !== false)
      .map((area, index) => (
        <defs key={`gradient-${area.key}`}>
          <linearGradient id={`gradient-${area.key}`} x1="0" y1="0" x2="0" y2="1">
            <stop 
              offset="5%" 
              stopColor={area.color} 
              stopOpacity={area.fillOpacity || 0.8} 
            />
            <stop 
              offset="95%" 
              stopColor={area.color} 
              stopOpacity={0.1} 
            />
          </linearGradient>
        </defs>
      ))
  }, [visibleAreas])

  // Process data for percent type
  const processedData = useMemo(() => {
    if (type !== 'percent') return data

    return data.map(item => {
      const total = visibleAreas.reduce((sum, area) => {
        const value = item[area.key]
        return sum + (typeof value === 'number' ? value : 0)
      }, 0)

      const processedItem = { ...item }
      visibleAreas.forEach(area => {
        const value = item[area.key]
        if (typeof value === 'number' && total > 0) {
          processedItem[area.key] = (value / total) * 100
        }
      })

      return processedItem
    })
  }, [data, type, visibleAreas])

  // Custom tooltip component
  const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload || !payload.length) return null

    const sortedPayload = [...payload].sort((a, b) => b.value - a.value)

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
          {sortedPayload.map((item: any, index: number) => {
            const area = areasWithColors.find(a => a.key === item.dataKey)
            if (!area) return null

            const formattedValue = tooltip.formatter
              ? tooltip.formatter(item.value, item.name || area.name, item)[0]
              : dataFormatters.formatNumber(item.value, {
                  format: type === 'percent' ? 'percentage' : area.format,
                  unit: type === 'percent' ? '' : area.unit,
                  precision: 2
                })

            return (
              <div key={index} className="flex items-center gap-2 text-sm">
                <div
                  className="h-3 w-3 rounded-sm"
                  style={{ backgroundColor: item.color }}
                />
                <span className="font-medium">{item.name || area.name}:</span>
                <span>{formattedValue}</span>
              </div>
            )
          })}
          
          {/* Show total for stacked charts */}
          {type === 'stacked' && sortedPayload.length > 1 && (
            <div className="mt-2 pt-2 border-t">
              <div className="flex items-center gap-2 text-sm font-medium">
                <span>Total:</span>
                <span>
                  {dataFormatters.formatNumber(
                    sortedPayload.reduce((sum, item) => sum + item.value, 0),
                    {
                      format: visibleAreas[0]?.format,
                      unit: visibleAreas[0]?.unit,
                      precision: 2
                    }
                  )}
                </span>
              </div>
            </div>
          )}
        </div>
      </div>
    )
  }

  // Custom Y-axis tick formatter
  const formatYAxisTick = (value: any) => {
    return dataFormatters.formatNumber(value, {
      format: type === 'percent' ? 'percentage' : yAxis.format,
      unit: type === 'percent' ? '' : yAxis.unit,
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
        <RechartsAreaChart
          data={processedData}
          margin={chartMargin}
          onClick={handleDataPointClick}
          stackOffset={type === 'percent' ? 'expand' : undefined}
        >
          {/* Gradient definitions */}
          {gradientDefs}

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

          {/* Y Axis */}
          <YAxis
            domain={type === 'percent' ? [0, 100] : yAxis.domain}
            tickCount={yAxis.tickCount}
            tickFormatter={formatYAxisTick}
            tick={{ 
              fontSize: 12, 
              fill: theme?.colors.axis 
            }}
            axisLine={{ stroke: theme?.colors.grid }}
            tickLine={{ stroke: theme?.colors.grid }}
          />

          {/* Tooltip */}
          {tooltip.show && (
            <Tooltip
              content={<CustomTooltip />}
              shared={tooltip.shared}
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
              stroke={refLine.color || theme?.colors.semantic.warning}
              strokeDasharray={refLine.strokeDashArray || '5 5'}
              label={{
                value: refLine.label,
                position: 'topRight',
                style: { fill: theme?.colors.text, fontSize: '12px' }
              }}
            />
          ))}

          {/* Areas */}
          {visibleAreas.map((area) => (
            <Area
              key={area.key}
              type="monotone"
              dataKey={area.key}
              name={area.name}
              stackId={type === 'overlapping' ? undefined : (area.stackId || 'default')}
              stroke={area.color}
              strokeWidth={area.strokeWidth || 1}
              fill={
                area.gradient !== false 
                  ? `url(#gradient-${area.key})` 
                  : area.color
              }
              fillOpacity={area.gradient !== false ? 1 : (area.fillOpacity || 0.6)}
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
        </RechartsAreaChart>
      </ResponsiveContainer>
    </div>
  )
}

export { AreaChart }
export type { AreaChartProps }