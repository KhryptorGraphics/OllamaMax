/**
 * BarChart Component
 * Categorical data comparison with horizontal and vertical orientations,
 * grouped and stacked modes, and interactive features
 */

import React, { useMemo } from 'react'
import {
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  ReferenceLine,
  Cell
} from 'recharts'
import { 
  dataFormatters, 
  colorUtils,
  responsiveUtils,
  type ChartTheme, 
  type CategoryDataPoint,
  type MultiSeriesDataPoint 
} from '@/utils/chartUtils'

export interface BarChartProps {
  /** Chart data */
  data: (CategoryDataPoint | MultiSeriesDataPoint)[]
  
  /** Bars to display */
  bars: Array<{
    key: string
    name: string
    color?: string
    stackId?: string
    hide?: boolean
    radius?: number
    format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
    unit?: string
  }>
  
  /** Chart theme */
  theme?: ChartTheme
  
  /** Chart dimensions */
  dimensions?: { width: number; height: number }
  
  /** Enable animations */
  animations?: boolean
  
  /** Chart orientation */
  orientation?: 'vertical' | 'horizontal'
  
  /** Chart type */
  type?: 'grouped' | 'stacked'
  
  /** X-axis configuration */
  xAxis?: {
    dataKey?: string
    tickCount?: number
    angle?: number
    interval?: number | 'preserveStart' | 'preserveEnd' | 'preserveStartEnd'
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
    clickable?: boolean
    highlightOnHover?: boolean
  }
  
  /** Reference lines */
  referenceLines?: Array<{
    value: number
    label?: string
    color?: string
    strokeDashArray?: string
    axis?: 'x' | 'y'
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
  
  /** Color mapping for individual bars */
  colorMapping?: Record<string, string>
  
  /** Loading state */
  loading?: boolean
  
  /** Empty state message */
  emptyMessage?: string
  
  /** Click handler for bars */
  onBarClick?: (data: any, index: number) => void
  
  /** Hover handler for bars */
  onBarHover?: (data: any, index: number) => void
}

const BarChart: React.FC<BarChartProps> = ({
  data,
  bars,
  theme,
  dimensions,
  animations = true,
  orientation = 'vertical',
  type = 'grouped',
  xAxis = {
    dataKey: 'label'
  },
  yAxis = {},
  interactive = {
    clickable: false,
    highlightOnHover: true
  },
  referenceLines = [],
  tooltip = { show: true },
  legend = { show: true, position: 'bottom' },
  grid = { show: true, strokeDashArray: '3 3' },
  margin,
  colorMapping = {},
  loading = false,
  emptyMessage = 'No data available',
  onBarClick,
  onBarHover
}) => {
  // Generate responsive margins if not provided
  const chartMargin = useMemo(() => {
    if (margin) return margin
    const baseMargin = responsiveUtils.getResponsiveMargins()
    
    // Adjust margins for horizontal orientation
    if (orientation === 'horizontal') {
      return {
        ...baseMargin,
        left: Math.max(baseMargin.left, 100), // More space for labels
        bottom: baseMargin.bottom - 20
      }
    }
    
    return baseMargin
  }, [margin, orientation])

  // Generate colors for bars
  const barsWithColors = useMemo(() => {
    if (!theme) return bars
    
    const colors = colorUtils.generateColorPalette(bars.length, theme)
    return bars.map((bar, index) => ({
      ...bar,
      color: bar.color || colors[index]
    }))
  }, [bars, theme])

  // Filter visible bars
  const visibleBars = useMemo(() => 
    barsWithColors.filter(bar => !bar.hide),
    [barsWithColors]
  )

  // Custom tooltip component
  const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload || !payload.length) return null

    return (
      <div className="rounded-lg border bg-popover p-3 shadow-lg">
        <div className="mb-2 text-sm font-medium">
          {tooltip.labelFormatter 
            ? tooltip.labelFormatter(label)
            : label
          }
        </div>
        <div className="space-y-1">
          {payload.map((item: any, index: number) => {
            const bar = barsWithColors.find(b => b.key === item.dataKey)
            if (!bar) return null

            const formattedValue = tooltip.formatter
              ? tooltip.formatter(item.value, item.name || bar.name, item)[0]
              : dataFormatters.formatNumber(item.value, {
                  format: bar.format,
                  unit: bar.unit,
                  precision: 2
                })

            return (
              <div key={index} className="flex items-center gap-2 text-sm">
                <div
                  className="h-3 w-3 rounded-sm"
                  style={{ backgroundColor: item.color }}
                />
                <span className="font-medium">{item.name || bar.name}:</span>
                <span>{formattedValue}</span>
              </div>
            )
          })}
        </div>
      </div>
    )
  }

  // Custom Y-axis tick formatter
  const formatYAxisTick = (value: any) => {
    if (orientation === 'horizontal') {
      return value // Labels for horizontal charts
    }
    
    return dataFormatters.formatNumber(value, {
      format: yAxis.format,
      unit: yAxis.unit,
      precision: 1
    })
  }

  // Custom X-axis tick formatter
  const formatXAxisTick = (value: any) => {
    if (orientation === 'vertical') {
      return value // Labels for vertical charts
    }
    
    return dataFormatters.formatNumber(value, {
      format: yAxis.format, // Use Y-axis format for horizontal orientation
      unit: yAxis.unit,
      precision: 1
    })
  }

  // Handle bar clicks
  const handleBarClick = (data: any, index: number) => {
    if (interactive.clickable && onBarClick) {
      onBarClick(data, index)
    }
  }

  // Get bar color based on data point and color mapping
  const getBarColor = (entry: any, barIndex: number) => {
    // Use color mapping if available
    if (colorMapping[entry[xAxis.dataKey]]) {
      return colorMapping[entry[xAxis.dataKey]]
    }
    
    // Use default bar color
    return visibleBars[barIndex]?.color || theme?.colors.primary[0]
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

  const ChartComponent = RechartsBarChart

  return (
    <div className="h-full w-full">
      <ResponsiveContainer width="100%" height="100%">
        <ChartComponent
          data={data}
          margin={chartMargin}
          onClick={handleBarClick}
          layout={orientation === 'horizontal' ? 'horizontal' : 'vertical'}
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
            type={orientation === 'horizontal' ? 'number' : 'category'}
            dataKey={orientation === 'vertical' ? xAxis.dataKey : undefined}
            domain={orientation === 'horizontal' ? yAxis.domain : undefined}
            tickFormatter={formatXAxisTick}
            tick={{ 
              fontSize: 12, 
              fill: theme?.colors.axis,
              angle: xAxis.angle || 0
            }}
            axisLine={{ stroke: theme?.colors.grid }}
            tickLine={{ stroke: theme?.colors.grid }}
            tickCount={xAxis.tickCount}
            interval={xAxis.interval}
          />

          {/* Y Axis */}
          <YAxis
            type={orientation === 'horizontal' ? 'category' : 'number'}
            dataKey={orientation === 'horizontal' ? xAxis.dataKey : undefined}
            domain={orientation === 'vertical' ? yAxis.domain : undefined}
            tickFormatter={formatYAxisTick}
            tick={{ 
              fontSize: 12, 
              fill: theme?.colors.axis 
            }}
            axisLine={{ stroke: theme?.colors.grid }}
            tickLine={{ stroke: theme?.colors.grid }}
            tickCount={yAxis.tickCount}
            width={orientation === 'horizontal' ? 80 : undefined}
          />

          {/* Tooltip */}
          {tooltip.show && (
            <Tooltip
              content={<CustomTooltip />}
              cursor={{
                fill: theme?.colors.grid,
                fillOpacity: 0.1
              }}
            />
          )}

          {/* Legend */}
          {legend.show && visibleBars.length > 1 && (
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
              {...(refLine.axis === 'x' || orientation === 'horizontal' 
                ? { x: refLine.value } 
                : { y: refLine.value }
              )}
              stroke={refLine.color || theme?.colors.semantic.warning}
              strokeDasharray={refLine.strokeDashArray || '5 5'}
              label={{
                value: refLine.label,
                position: 'topRight',
                style: { fill: theme?.colors.text, fontSize: '12px' }
              }}
            />
          ))}

          {/* Bars */}
          {visibleBars.map((bar, barIndex) => (
            <Bar
              key={bar.key}
              dataKey={bar.key}
              name={bar.name}
              stackId={type === 'stacked' ? (bar.stackId || 'default') : undefined}
              fill={bar.color}
              radius={bar.radius || 0}
              animationDuration={animations ? 1000 : 0}
              onMouseEnter={interactive.highlightOnHover ? onBarHover : undefined}
            >
              {/* Individual cell colors for single series */}
              {visibleBars.length === 1 && data.map((entry, index) => (
                <Cell 
                  key={`cell-${index}`} 
                  fill={getBarColor(entry, barIndex)}
                />
              ))}
            </Bar>
          ))}
        </ChartComponent>
      </ResponsiveContainer>
    </div>
  )
}

export { BarChart }
export type { BarChartProps }