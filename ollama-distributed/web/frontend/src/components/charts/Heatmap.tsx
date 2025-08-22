/**
 * Heatmap Component
 * Resource utilization mapping with color gradients, tooltips, and cell interactions
 */

import React, { useMemo, useState } from 'react'
import { dataFormatters, colorUtils, type ChartTheme } from '@/utils/chartUtils'

export interface HeatmapDataPoint {
  x: number | string
  y: number | string
  value: number | null
  label?: string
  metadata?: Record<string, any>
}

export interface HeatmapProps {
  /** Chart data */
  data: HeatmapDataPoint[]
  
  /** Chart theme */
  theme?: ChartTheme
  
  /** Chart dimensions */
  dimensions?: { width: number; height: number }
  
  /** X-axis labels */
  xLabels?: string[]
  
  /** Y-axis labels */
  yLabels?: string[]
  
  /** Value formatting */
  valueFormat?: {
    format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
    unit?: string
    precision?: number
  }
  
  /** Color configuration */
  colorScale?: {
    type?: 'linear' | 'quantile' | 'threshold'
    colors?: string[]
    domain?: [number, number]
    thresholds?: number[]
  }
  
  /** Cell configuration */
  cell?: {
    size?: number | 'auto'
    gap?: number
    radius?: number
    border?: boolean
    borderColor?: string
    borderWidth?: number
  }
  
  /** Interactive features */
  interactive?: {
    clickable?: boolean
    highlightOnHover?: boolean
    showTooltip?: boolean
  }
  
  /** Tooltip configuration */
  tooltip?: {
    formatter?: (value: number, x: any, y: any, data: HeatmapDataPoint) => string
    position?: 'follow' | 'fixed'
  }
  
  /** Legend configuration */
  legend?: {
    show?: boolean
    position?: 'top' | 'bottom' | 'left' | 'right'
    width?: number
    height?: number
    steps?: number
  }
  
  /** Axis configuration */
  axes?: {
    showXAxis?: boolean
    showYAxis?: boolean
    labelRotation?: number
    fontSize?: number
  }
  
  /** Missing data configuration */
  missingData?: {
    color?: string
    pattern?: 'none' | 'diagonal' | 'dots'
    label?: string
  }
  
  /** Loading state */
  loading?: boolean
  
  /** Empty state message */
  emptyMessage?: string
  
  /** Click handler for cells */
  onCellClick?: (data: HeatmapDataPoint, x: any, y: any) => void
  
  /** Hover handler for cells */
  onCellHover?: (data: HeatmapDataPoint, x: any, y: any) => void
}

const Heatmap: React.FC<HeatmapProps> = ({
  data,
  theme,
  dimensions,
  xLabels = [],
  yLabels = [],
  valueFormat = {
    format: 'default',
    precision: 2
  },
  colorScale = {
    type: 'linear',
    colors: ['#1e40af', '#3b82f6', '#60a5fa', '#93c5fd', '#dbeafe']
  },
  cell = {
    size: 'auto',
    gap: 1,
    radius: 2,
    border: false,
    borderWidth: 1
  },
  interactive = {
    clickable: false,
    highlightOnHover: true,
    showTooltip: true
  },
  tooltip = {
    position: 'follow'
  },
  legend = {
    show: true,
    position: 'right',
    width: 20,
    height: 200,
    steps: 10
  },
  axes = {
    showXAxis: true,
    showYAxis: true,
    labelRotation: 0,
    fontSize: 12
  },
  missingData = {
    color: '#f3f4f6',
    pattern: 'diagonal',
    label: 'No data'
  },
  loading = false,
  emptyMessage = 'No data available',
  onCellClick,
  onCellHover
}) => {
  const [hoveredCell, setHoveredCell] = useState<{
    data: HeatmapDataPoint
    x: number
    y: number
    clientX: number
    clientY: number
  } | null>(null)

  // Generate grid structure
  const { grid, uniqueXValues, uniqueYValues } = useMemo(() => {
    const xValues = xLabels.length > 0 ? xLabels : Array.from(new Set(data.map(d => d.x))).sort()
    const yValues = yLabels.length > 0 ? yLabels : Array.from(new Set(data.map(d => d.y))).sort()
    
    // Create grid matrix
    const gridMatrix: (HeatmapDataPoint | null)[][] = []
    const dataMap = new Map<string, HeatmapDataPoint>()
    
    // Create lookup map
    data.forEach(point => {
      const key = `${point.x}-${point.y}`
      dataMap.set(key, point)
    })
    
    // Fill grid
    yValues.forEach(y => {
      const row: (HeatmapDataPoint | null)[] = []
      xValues.forEach(x => {
        const key = `${x}-${y}`
        row.push(dataMap.get(key) || null)
      })
      gridMatrix.push(row)
    })
    
    return {
      grid: gridMatrix,
      uniqueXValues: xValues,
      uniqueYValues: yValues
    }
  }, [data, xLabels, yLabels])

  // Calculate value domain
  const valueDomain = useMemo(() => {
    const values = data
      .map(d => d.value)
      .filter((v): v is number => v !== null && v !== undefined && !isNaN(v))
    
    if (values.length === 0) return [0, 1]
    
    if (colorScale.domain) {
      return colorScale.domain
    }
    
    return [Math.min(...values), Math.max(...values)]
  }, [data, colorScale.domain])

  // Generate color scale function
  const getColorForValue = useMemo(() => {
    const colors = colorScale.colors || (theme ? [
      theme.colors.primary[4],
      theme.colors.primary[3],
      theme.colors.primary[2],
      theme.colors.primary[1],
      theme.colors.primary[0]
    ] : ['#e5e7eb', '#9ca3af', '#6b7280', '#374151', '#1f2937'])

    return (value: number | null): string => {
      if (value === null || value === undefined || isNaN(value)) {
        return missingData.color || '#f3f4f6'
      }

      const [min, max] = valueDomain
      const normalizedValue = max > min ? (value - min) / (max - min) : 0
      const clampedValue = Math.max(0, Math.min(1, normalizedValue))

      if (colorScale.type === 'threshold' && colorScale.thresholds) {
        const thresholdIndex = colorScale.thresholds.findIndex(t => value <= t)
        return colors[thresholdIndex >= 0 ? thresholdIndex : colors.length - 1]
      }

      // Linear interpolation
      const colorIndex = clampedValue * (colors.length - 1)
      const lowerIndex = Math.floor(colorIndex)
      const upperIndex = Math.ceil(colorIndex)
      
      if (lowerIndex === upperIndex) {
        return colors[lowerIndex]
      }

      // Simple color interpolation (hex to hex)
      return colors[lowerIndex] // Simplified - could implement proper color interpolation
    }
  }, [colorScale, theme, valueDomain, missingData.color])

  // Calculate cell dimensions
  const cellDimensions = useMemo(() => {
    if (!dimensions) return { width: 40, height: 40 }
    
    const availableWidth = dimensions.width - (legend.show && legend.position === 'right' ? legend.width! + 20 : 0) - 80
    const availableHeight = dimensions.height - 80
    
    if (cell.size === 'auto') {
      const cellWidth = Math.max(20, availableWidth / uniqueXValues.length)
      const cellHeight = Math.max(20, availableHeight / uniqueYValues.length)
      return { width: cellWidth, height: cellHeight }
    }
    
    return { width: cell.size, height: cell.size }
  }, [dimensions, cell.size, uniqueXValues.length, uniqueYValues.length, legend])

  // Handle cell interactions
  const handleCellMouseEnter = (
    cellData: HeatmapDataPoint | null,
    x: any,
    y: any,
    event: React.MouseEvent
  ) => {
    if (!interactive.highlightOnHover && !interactive.showTooltip) return
    
    if (cellData) {
      setHoveredCell({
        data: cellData,
        x,
        y,
        clientX: event.clientX,
        clientY: event.clientY
      })
      
      if (onCellHover) {
        onCellHover(cellData, x, y)
      }
    }
  }

  const handleCellMouseLeave = () => {
    setHoveredCell(null)
  }

  const handleCellClick = (cellData: HeatmapDataPoint | null, x: any, y: any) => {
    if (interactive.clickable && cellData && onCellClick) {
      onCellClick(cellData, x, y)
    }
  }

  // Render missing data pattern
  const renderMissingDataPattern = (x: number, y: number, width: number, height: number) => {
    if (missingData.pattern === 'diagonal') {
      return (
        <pattern
          id={`missing-pattern-${x}-${y}`}
          patternUnits="userSpaceOnUse"
          width="4"
          height="4"
        >
          <rect width="4" height="4" fill={missingData.color} />
          <path d="M0,4 L4,0 M-1,1 L1,-1 M3,5 L5,3" stroke="#d1d5db" strokeWidth="0.5" />
        </pattern>
      )
    }
    
    if (missingData.pattern === 'dots') {
      return (
        <pattern
          id={`missing-pattern-${x}-${y}`}
          patternUnits="userSpaceOnUse"
          width="6"
          height="6"
        >
          <rect width="6" height="6" fill={missingData.color} />
          <circle cx="3" cy="3" r="1" fill="#d1d5db" />
        </pattern>
      )
    }
    
    return null
  }

  // Render color legend
  const renderLegend = () => {
    if (!legend.show) return null

    const steps = legend.steps || 10
    const [min, max] = valueDomain
    const stepSize = (max - min) / steps

    return (
      <div className="flex flex-col items-center">
        <div className="text-xs font-medium mb-2" style={{ color: theme?.colors.text }}>
          {dataFormatters.formatNumber(max, valueFormat)}
        </div>
        
        <div
          className="border"
          style={{
            width: legend.width,
            height: legend.height,
            borderColor: theme?.colors.grid
          }}
        >
          {Array.from({ length: steps }, (_, i) => {
            const value = max - (i * stepSize)
            const color = getColorForValue(value)
            
            return (
              <div
                key={i}
                style={{
                  backgroundColor: color,
                  height: `${100 / steps}%`,
                  width: '100%'
                }}
              />
            )
          })}
        </div>
        
        <div className="text-xs font-medium mt-2" style={{ color: theme?.colors.text }}>
          {dataFormatters.formatNumber(min, valueFormat)}
        </div>
      </div>
    )
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
    <div className="h-full w-full relative">
      <div className="flex h-full">
        {/* Main heatmap area */}
        <div className="flex-1">
          <svg
            width="100%"
            height="100%"
            viewBox={`0 0 ${uniqueXValues.length * (cellDimensions.width + cell.gap!)} ${uniqueYValues.length * (cellDimensions.height + cell.gap!)}`}
          >
            {/* Missing data patterns */}
            <defs>
              {grid.flatMap((row, rowIndex) =>
                row.map((cellData, colIndex) => {
                  if (cellData === null || cellData.value === null) {
                    return renderMissingDataPattern(
                      colIndex,
                      rowIndex,
                      cellDimensions.width,
                      cellDimensions.height
                    )
                  }
                  return null
                })
              )}
            </defs>

            {/* Render cells */}
            {grid.map((row, rowIndex) =>
              row.map((cellData, colIndex) => {
                const x = colIndex * (cellDimensions.width + cell.gap!)
                const y = rowIndex * (cellDimensions.height + cell.gap!)
                const isHovered = hoveredCell?.x === colIndex && hoveredCell?.y === rowIndex
                
                const cellColor = cellData === null || cellData.value === null
                  ? `url(#missing-pattern-${colIndex}-${rowIndex})`
                  : getColorForValue(cellData.value)

                return (
                  <rect
                    key={`${colIndex}-${rowIndex}`}
                    x={x}
                    y={y}
                    width={cellDimensions.width}
                    height={cellDimensions.height}
                    fill={cellColor}
                    rx={cell.radius}
                    ry={cell.radius}
                    stroke={cell.border ? (cell.borderColor || theme?.colors.grid) : 'none'}
                    strokeWidth={cell.borderWidth}
                    style={{
                      opacity: isHovered && interactive.highlightOnHover ? 0.8 : 1,
                      cursor: interactive.clickable ? 'pointer' : 'default'
                    }}
                    onMouseEnter={(e) => handleCellMouseEnter(cellData, colIndex, rowIndex, e)}
                    onMouseLeave={handleCellMouseLeave}
                    onClick={() => handleCellClick(cellData, colIndex, rowIndex)}
                  />
                )
              })
            )}
          </svg>

          {/* Axis labels */}
          {axes.showXAxis && (
            <div className="flex justify-start mt-2">
              {uniqueXValues.map((label, index) => (
                <div
                  key={index}
                  className="text-center"
                  style={{
                    width: cellDimensions.width + cell.gap!,
                    fontSize: axes.fontSize,
                    color: theme?.colors.text,
                    transform: axes.labelRotation ? `rotate(${axes.labelRotation}deg)` : undefined
                  }}
                >
                  {label}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Legend */}
        {legend.show && legend.position === 'right' && (
          <div className="ml-4">
            {renderLegend()}
          </div>
        )}
      </div>

      {/* Y-axis labels */}
      {axes.showYAxis && (
        <div className="absolute left-0 top-0 flex flex-col justify-start">
          {uniqueYValues.map((label, index) => (
            <div
              key={index}
              className="flex items-center justify-end pr-2"
              style={{
                height: cellDimensions.height + cell.gap!,
                fontSize: axes.fontSize,
                color: theme?.colors.text
              }}
            >
              {label}
            </div>
          ))}
        </div>
      )}

      {/* Tooltip */}
      {interactive.showTooltip && hoveredCell && (
        <div
          className="absolute z-10 rounded-lg border bg-popover p-3 shadow-lg pointer-events-none"
          style={{
            left: tooltip.position === 'follow' ? hoveredCell.clientX + 10 : '50%',
            top: tooltip.position === 'follow' ? hoveredCell.clientY - 10 : '50%',
            transform: tooltip.position === 'fixed' ? 'translate(-50%, -50%)' : undefined
          }}
        >
          <div className="space-y-1">
            <div className="text-sm font-medium">
              {uniqueXValues[hoveredCell.x]} × {uniqueYValues[hoveredCell.y]}
            </div>
            <div className="text-sm">
              {hoveredCell.data.value !== null
                ? tooltip.formatter
                  ? tooltip.formatter(
                      hoveredCell.data.value,
                      uniqueXValues[hoveredCell.x],
                      uniqueYValues[hoveredCell.y],
                      hoveredCell.data
                    )
                  : dataFormatters.formatNumber(hoveredCell.data.value, valueFormat)
                : missingData.label
              }
            </div>
            {hoveredCell.data.label && (
              <div className="text-xs text-muted-foreground">
                {hoveredCell.data.label}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export { Heatmap }
export type { HeatmapProps, HeatmapDataPoint }