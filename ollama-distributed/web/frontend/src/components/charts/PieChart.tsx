/**
 * PieChart Component
 * Distribution and percentage data visualization with donut mode,
 * custom labels, and interactive features
 */

import React, { useMemo, useState } from 'react'
import {
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
  Legend
} from 'recharts'
import { 
  dataFormatters, 
  colorUtils,
  responsiveUtils,
  type ChartTheme, 
  type CategoryDataPoint 
} from '@/utils/chartUtils'

export interface PieChartProps {
  /** Chart data */
  data: CategoryDataPoint[]
  
  /** Chart theme */
  theme?: ChartTheme
  
  /** Chart dimensions */
  dimensions?: { width: number; height: number }
  
  /** Enable animations */
  animations?: boolean
  
  /** Chart variant */
  variant?: 'pie' | 'donut'
  
  /** Data key for values */
  dataKey?: string
  
  /** Data key for labels */
  labelKey?: string
  
  /** Value formatting */
  valueFormat?: {
    format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
    unit?: string
    precision?: number
  }
  
  /** Label configuration */
  labels?: {
    show?: boolean
    position?: 'inside' | 'outside' | 'center'
    type?: 'key' | 'value' | 'percent' | 'keyPercent' | 'valuePercent'
    fontSize?: number
  }
  
  /** Interactive features */
  interactive?: {
    clickable?: boolean
    highlightOnHover?: boolean
    expandOnHover?: boolean
  }
  
  /** Tooltip configuration */
  tooltip?: {
    formatter?: (value: any, name: string, props: any) => [string, string]
    show?: boolean
  }
  
  /** Legend configuration */
  legend?: {
    show?: boolean
    position?: 'top' | 'bottom' | 'left' | 'right'
    align?: 'left' | 'center' | 'right'
  }
  
  /** Color configuration */
  colors?: string[] | 'semantic' | 'auto'
  
  /** Donut configuration */
  donut?: {
    innerRadius?: number | string
    centerContent?: React.ReactNode
  }
  
  /** Start and end angles */
  startAngle?: number
  endAngle?: number
  
  /** Minimum angle for small slices */
  minAngle?: number
  
  /** Padding between slices */
  paddingAngle?: number
  
  /** Loading state */
  loading?: boolean
  
  /** Empty state message */
  emptyMessage?: string
  
  /** Click handler for slices */
  onSliceClick?: (data: CategoryDataPoint, index: number) => void
  
  /** Hover handler for slices */
  onSliceHover?: (data: CategoryDataPoint, index: number) => void
}

const PieChart: React.FC<PieChartProps> = ({
  data,
  theme,
  dimensions,
  animations = true,
  variant = 'pie',
  dataKey = 'value',
  labelKey = 'label',
  valueFormat = {
    format: 'default',
    precision: 1
  },
  labels = {
    show: true,
    position: 'outside',
    type: 'keyPercent'
  },
  interactive = {
    clickable: false,
    highlightOnHover: true,
    expandOnHover: false
  },
  tooltip = { show: true },
  legend = { 
    show: true, 
    position: 'bottom',
    align: 'center'
  },
  colors = 'auto',
  donut = {
    innerRadius: '40%'
  },
  startAngle = 90,
  endAngle = -270,
  minAngle = 0,
  paddingAngle = 0,
  loading = false,
  emptyMessage = 'No data available',
  onSliceClick,
  onSliceHover
}) => {
  const [activeIndex, setActiveIndex] = useState<number | null>(null)

  // Calculate total value for percentage calculations
  const totalValue = useMemo(() => {
    return data.reduce((sum, item) => sum + (item[dataKey] || 0), 0)
  }, [data, dataKey])

  // Generate colors for slices
  const sliceColors = useMemo(() => {
    if (Array.isArray(colors)) {
      return colors
    }
    
    if (!theme) return []
    
    if (colors === 'semantic') {
      // Use semantic colors for status-like data
      const semanticColors = [
        theme.colors.semantic.success,
        theme.colors.semantic.info,
        theme.colors.semantic.warning,
        theme.colors.semantic.error
      ]
      return colorUtils.generateColorPalette(data.length, {
        ...theme,
        colors: { ...theme.colors, primary: semanticColors }
      })
    }
    
    return colorUtils.generateColorPalette(data.length, theme)
  }, [colors, theme, data.length])

  // Process data with calculated percentages
  const processedData = useMemo(() => {
    return data.map((item, index) => ({
      ...item,
      percentage: totalValue > 0 ? (item[dataKey] / totalValue) * 100 : 0,
      color: sliceColors[index] || theme?.colors.primary[0] || '#3b82f6'
    }))
  }, [data, dataKey, totalValue, sliceColors, theme])

  // Custom label rendering function
  const renderCustomLabel = (entry: any) => {
    if (!labels.show) return null

    const { cx, cy, midAngle, innerRadius, outerRadius, index } = entry
    const item = processedData[index]
    
    if (!item) return null

    let labelText = ''
    
    switch (labels.type) {
      case 'key':
        labelText = item[labelKey]
        break
      case 'value':
        labelText = dataFormatters.formatNumber(item[dataKey], valueFormat)
        break
      case 'percent':
        labelText = `${item.percentage.toFixed(1)}%`
        break
      case 'keyPercent':
        labelText = `${item[labelKey]} (${item.percentage.toFixed(1)}%)`
        break
      case 'valuePercent':
        labelText = `${dataFormatters.formatNumber(item[dataKey], valueFormat)} (${item.percentage.toFixed(1)}%)`
        break
      default:
        labelText = item[labelKey]
    }

    if (labels.position === 'center' && variant === 'donut') {
      return (
        <text
          x={cx}
          y={cy}
          textAnchor="middle"
          dominantBaseline="central"
          fontSize={labels.fontSize || 14}
          fill={theme?.colors.text}
        >
          {labelText}
        </text>
      )
    }

    if (labels.position === 'inside') {
      const radius = innerRadius + (outerRadius - innerRadius) * 0.5
      const x = cx + radius * Math.cos(-midAngle * Math.PI / 180)
      const y = cy + radius * Math.sin(-midAngle * Math.PI / 180)

      return (
        <text
          x={x}
          y={y}
          textAnchor={x > cx ? 'start' : 'end'}
          dominantBaseline="central"
          fontSize={labels.fontSize || 12}
          fill="#ffffff"
        >
          {labelText}
        </text>
      )
    }

    // Outside labels
    const RADIAN = Math.PI / 180
    const radius = outerRadius + 30
    const x = cx + radius * Math.cos(-midAngle * RADIAN)
    const y = cy + radius * Math.sin(-midAngle * RADIAN)

    return (
      <text
        x={x}
        y={y}
        textAnchor={x > cx ? 'start' : 'end'}
        dominantBaseline="central"
        fontSize={labels.fontSize || 12}
        fill={theme?.colors.text}
      >
        {labelText}
      </text>
    )
  }

  // Custom tooltip component
  const CustomTooltip = ({ active, payload }: any) => {
    if (!active || !payload || !payload.length) return null

    const data = payload[0].payload
    const formattedValue = tooltip.formatter
      ? tooltip.formatter(data[dataKey], data[labelKey], data)[0]
      : dataFormatters.formatNumber(data[dataKey], valueFormat)

    return (
      <div className="rounded-lg border bg-popover p-3 shadow-lg">
        <div className="space-y-1">
          <div className="flex items-center gap-2 text-sm">
            <div
              className="h-3 w-3 rounded-sm"
              style={{ backgroundColor: data.color }}
            />
            <span className="font-medium">{data[labelKey]}:</span>
          </div>
          <div className="ml-5 text-sm">
            <div>Value: {formattedValue}</div>
            <div>Percentage: {data.percentage.toFixed(1)}%</div>
          </div>
        </div>
      </div>
    )
  }

  // Handle slice interactions
  const handleSliceEnter = (data: any, index: number) => {
    if (interactive.highlightOnHover) {
      setActiveIndex(index)
    }
    if (onSliceHover) {
      onSliceHover(data, index)
    }
  }

  const handleSliceLeave = () => {
    if (interactive.highlightOnHover) {
      setActiveIndex(null)
    }
  }

  const handleSliceClick = (data: any, index: number) => {
    if (interactive.clickable && onSliceClick) {
      onSliceClick(data, index)
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

  if (!data || data.length === 0 || totalValue === 0) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-sm text-muted-foreground">{emptyMessage}</div>
      </div>
    )
  }

  const pieProps = {
    data: processedData,
    dataKey,
    nameKey: labelKey,
    cx: '50%',
    cy: '50%',
    startAngle,
    endAngle,
    innerRadius: variant === 'donut' ? donut.innerRadius : 0,
    outerRadius: '80%',
    paddingAngle,
    minAngle,
    animationDuration: animations ? 1000 : 0,
    label: labels.position !== 'center' ? renderCustomLabel : false,
    labelLine: labels.position === 'outside'
  }

  return (
    <div className="h-full w-full relative">
      <ResponsiveContainer width="100%" height="100%">
        <RechartsPieChart>
          <Pie {...pieProps}>
            {processedData.map((entry, index) => (
              <Cell
                key={`cell-${index}`}
                fill={entry.color}
                stroke={theme?.colors.background}
                strokeWidth={2}
                style={{
                  filter: activeIndex === index && interactive.highlightOnHover
                    ? 'brightness(1.1)'
                    : undefined,
                  transform: activeIndex === index && interactive.expandOnHover
                    ? 'scale(1.05)'
                    : undefined,
                  transformOrigin: 'center',
                  transition: 'all 0.2s ease-in-out',
                  cursor: interactive.clickable ? 'pointer' : 'default'
                }}
                onMouseEnter={() => handleSliceEnter(entry, index)}
                onMouseLeave={handleSliceLeave}
                onClick={() => handleSliceClick(entry, index)}
              />
            ))}
          </Pie>

          {/* Center content for donut charts */}
          {variant === 'donut' && donut.centerContent && (
            <text
              x="50%"
              y="50%"
              textAnchor="middle"
              dominantBaseline="central"
              fill={theme?.colors.text}
            >
              {donut.centerContent}
            </text>
          )}

          {/* Center label for donut charts */}
          {variant === 'donut' && labels.position === 'center' && activeIndex !== null && (
            <text
              x="50%"
              y="50%"
              textAnchor="middle"
              dominantBaseline="central"
              fontSize={labels.fontSize || 16}
              fill={theme?.colors.text}
            >
              {processedData[activeIndex] && renderCustomLabel({
                cx: 0, cy: 0, midAngle: 0, innerRadius: 0, outerRadius: 0,
                index: activeIndex
              })}
            </text>
          )}

          {/* Tooltip */}
          {tooltip.show && (
            <Tooltip content={<CustomTooltip />} />
          )}

          {/* Legend */}
          {legend.show && (
            <Legend
              verticalAlign={legend.position === 'top' || legend.position === 'bottom' 
                ? legend.position 
                : 'middle'
              }
              align={legend.align}
              layout={legend.position === 'left' || legend.position === 'right' 
                ? 'vertical' 
                : 'horizontal'
              }
              wrapperStyle={{
                color: theme?.colors.text,
                fontSize: '12px'
              }}
            />
          )}
        </RechartsPieChart>
      </ResponsiveContainer>
    </div>
  )
}

export { PieChart }
export type { PieChartProps }