/**
 * GaugeChart Component
 * Single metric visualization with thresholds, progress indicators,
 * and customizable styling
 */

import React, { useMemo } from 'react'
import { dataFormatters, colorUtils, type ChartTheme } from '@/utils/chartUtils'

export interface GaugeChartProps {
  /** Current value */
  value: number
  
  /** Minimum value */
  min?: number
  
  /** Maximum value */
  max?: number
  
  /** Chart theme */
  theme?: ChartTheme
  
  /** Chart dimensions */
  dimensions?: { width: number; height: number }
  
  /** Enable animations */
  animations?: boolean
  
  /** Gauge style */
  style?: 'full' | 'half' | 'quarter'
  
  /** Value formatting */
  valueFormat?: {
    format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
    unit?: string
    precision?: number
  }
  
  /** Thresholds for color coding */
  thresholds?: Array<{
    value: number
    color?: string
    label?: string
  }>
  
  /** Gauge appearance */
  appearance?: {
    size?: number
    thickness?: number
    backgroundColor?: string
    showValue?: boolean
    showLabel?: boolean
    showThresholds?: boolean
    rounded?: boolean
  }
  
  /** Labels */
  labels?: {
    title?: string
    subtitle?: string
    unit?: string
  }
  
  /** Interactive features */
  interactive?: {
    clickable?: boolean
    highlightOnHover?: boolean
  }
  
  /** Target value indicator */
  target?: {
    value: number
    color?: string
    width?: number
    label?: string
  }
  
  /** Gradient configuration */
  gradient?: {
    enabled?: boolean
    colors?: string[]
    direction?: 'clockwise' | 'counterclockwise'
  }
  
  /** Loading state */
  loading?: boolean
  
  /** Click handler */
  onClick?: (value: number) => void
  
  /** Hover handler */
  onHover?: (value: number) => void
}

const GaugeChart: React.FC<GaugeChartProps> = ({
  value,
  min = 0,
  max = 100,
  theme,
  dimensions,
  animations = true,
  style = 'half',
  valueFormat = {
    format: 'default',
    precision: 1
  },
  thresholds = [],
  appearance = {
    size: 200,
    thickness: 20,
    showValue: true,
    showLabel: true,
    showThresholds: false,
    rounded: true
  },
  labels = {},
  interactive = {
    clickable: false,
    highlightOnHover: false
  },
  target,
  gradient = {
    enabled: false
  },
  loading = false,
  onClick,
  onHover
}) => {
  // Calculate gauge parameters
  const gaugeParams = useMemo(() => {
    const size = appearance.size || 200
    const thickness = appearance.thickness || 20
    const radius = (size - thickness) / 2
    const centerX = size / 2
    const centerY = size / 2
    
    let startAngle: number
    let endAngle: number
    let angleRange: number
    
    switch (style) {
      case 'full':
        startAngle = -90
        endAngle = 270
        angleRange = 360
        break
      case 'quarter':
        startAngle = 0
        endAngle = 90
        angleRange = 90
        break
      case 'half':
      default:
        startAngle = 180
        endAngle = 0
        angleRange = 180
        break
    }
    
    return {
      size,
      thickness,
      radius,
      centerX,
      centerY,
      startAngle,
      endAngle,
      angleRange
    }
  }, [appearance.size, appearance.thickness, style])

  // Normalize value to 0-1 range
  const normalizedValue = useMemo(() => {
    const clampedValue = Math.max(min, Math.min(max, value))
    return (clampedValue - min) / (max - min)
  }, [value, min, max])

  // Calculate angles
  const valueAngle = useMemo(() => {
    const { startAngle, angleRange } = gaugeParams
    return startAngle + (normalizedValue * angleRange)
  }, [normalizedValue, gaugeParams])

  // Get color based on thresholds
  const getValueColor = useMemo(() => {
    if (thresholds.length === 0) {
      return theme?.colors.primary[0] || '#3b82f6'
    }

    // Find the appropriate threshold
    const sortedThresholds = [...thresholds].sort((a, b) => a.value - b.value)
    const matchingThreshold = sortedThresholds.find(t => value <= t.value)
    
    if (matchingThreshold?.color) {
      return matchingThreshold.color
    }

    // Default semantic colors based on position
    const position = normalizedValue
    if (!theme) return '#3b82f6'
    
    if (position < 0.5) return theme.colors.semantic.success
    if (position < 0.8) return theme.colors.semantic.warning
    return theme.colors.semantic.error
  }, [value, thresholds, normalizedValue, theme])

  // Generate gradient if enabled
  const gradientId = `gauge-gradient-${Math.random().toString(36).substr(2, 9)}`
  const gradientDefinition = useMemo(() => {
    if (!gradient.enabled || !gradient.colors) return null

    const colors = gradient.colors.length > 0 
      ? gradient.colors 
      : theme 
        ? [theme.colors.semantic.success, theme.colors.semantic.warning, theme.colors.semantic.error]
        : ['#22c55e', '#f59e0b', '#ef4444']

    return (
      <defs>
        <linearGradient id={gradientId} gradientUnits="userSpaceOnUse">
          {colors.map((color, index) => (
            <stop
              key={index}
              offset={`${(index / (colors.length - 1)) * 100}%`}
              stopColor={color}
            />
          ))}
        </linearGradient>
      </defs>
    )
  }, [gradient, theme, gradientId])

  // Create arc path
  const createArcPath = (
    centerX: number,
    centerY: number,
    radius: number,
    startAngle: number,
    endAngle: number
  ): string => {
    const start = polarToCartesian(centerX, centerY, radius, startAngle)
    const end = polarToCartesian(centerX, centerY, radius, endAngle)
    const largeArcFlag = Math.abs(endAngle - startAngle) <= 180 ? '0' : '1'
    
    return [
      'M', start.x, start.y,
      'A', radius, radius, 0, largeArcFlag, 1, end.x, end.y
    ].join(' ')
  }

  // Convert polar coordinates to cartesian
  const polarToCartesian = (
    centerX: number,
    centerY: number,
    radius: number,
    angleInDegrees: number
  ) => {
    const angleInRadians = (angleInDegrees - 90) * Math.PI / 180.0
    return {
      x: centerX + (radius * Math.cos(angleInRadians)),
      y: centerY + (radius * Math.sin(angleInRadians))
    }
  }

  // Handle interactions
  const handleClick = () => {
    if (interactive.clickable && onClick) {
      onClick(value)
    }
  }

  const handleMouseEnter = () => {
    if (interactive.highlightOnHover && onHover) {
      onHover(value)
    }
  }

  // Background arc path
  const backgroundPath = createArcPath(
    gaugeParams.centerX,
    gaugeParams.centerY,
    gaugeParams.radius,
    gaugeParams.startAngle,
    gaugeParams.endAngle
  )

  // Value arc path
  const valuePath = createArcPath(
    gaugeParams.centerX,
    gaugeParams.centerY,
    gaugeParams.radius,
    gaugeParams.startAngle,
    valueAngle
  )

  // Target indicator position
  const targetPosition = useMemo(() => {
    if (!target) return null
    
    const normalizedTarget = (target.value - min) / (max - min)
    const targetAngle = gaugeParams.startAngle + (normalizedTarget * gaugeParams.angleRange)
    return polarToCartesian(
      gaugeParams.centerX,
      gaugeParams.centerY,
      gaugeParams.radius,
      targetAngle
    )
  }, [target, min, max, gaugeParams])

  // Threshold indicators
  const thresholdIndicators = useMemo(() => {
    if (!appearance.showThresholds || thresholds.length === 0) return []
    
    return thresholds.map((threshold, index) => {
      const normalizedThreshold = (threshold.value - min) / (max - min)
      const thresholdAngle = gaugeParams.startAngle + (normalizedThreshold * gaugeParams.angleRange)
      const position = polarToCartesian(
        gaugeParams.centerX,
        gaugeParams.centerY,
        gaugeParams.radius + 5,
        thresholdAngle
      )
      
      return {
        ...threshold,
        position,
        angle: thresholdAngle
      }
    })
  }, [thresholds, appearance.showThresholds, min, max, gaugeParams])

  // Loading state
  if (loading) {
    return (
      <div 
        className="flex items-center justify-center"
        style={{ 
          width: gaugeParams.size, 
          height: style === 'half' ? gaugeParams.size / 2 + 40 : gaugeParams.size 
        }}
      >
        <div className="text-sm text-muted-foreground">Loading...</div>
      </div>
    )
  }

  const containerHeight = style === 'half' 
    ? gaugeParams.size / 2 + 60 
    : style === 'quarter'
      ? gaugeParams.size / 2 + 40
      : gaugeParams.size + 40

  return (
    <div 
      className="relative flex flex-col items-center"
      style={{ width: gaugeParams.size, height: containerHeight }}
    >
      {/* Title */}
      {labels.title && (
        <div 
          className="text-lg font-semibold mb-2 text-center"
          style={{ color: theme?.colors.text }}
        >
          {labels.title}
        </div>
      )}

      {/* SVG Gauge */}
      <svg
        width={gaugeParams.size}
        height={style === 'half' ? gaugeParams.size / 2 + 20 : gaugeParams.size}
        className={interactive.clickable ? 'cursor-pointer' : ''}
        onClick={handleClick}
        onMouseEnter={handleMouseEnter}
      >
        {gradientDefinition}

        {/* Background arc */}
        <path
          d={backgroundPath}
          stroke={appearance.backgroundColor || theme?.colors.grid || '#e5e7eb'}
          strokeWidth={gaugeParams.thickness}
          fill="none"
          strokeLinecap={appearance.rounded ? 'round' : 'butt'}
        />

        {/* Value arc */}
        <path
          d={valuePath}
          stroke={gradient.enabled ? `url(#${gradientId})` : getValueColor}
          strokeWidth={gaugeParams.thickness}
          fill="none"
          strokeLinecap={appearance.rounded ? 'round' : 'butt'}
          style={{
            transition: animations ? 'stroke-dasharray 0.5s ease-in-out' : undefined
          }}
        />

        {/* Target indicator */}
        {target && targetPosition && (
          <g>
            <line
              x1={targetPosition.x - 8}
              y1={targetPosition.y - 8}
              x2={targetPosition.x + 8}
              y2={targetPosition.y + 8}
              stroke={target.color || theme?.colors.semantic.info || '#0ea5e9'}
              strokeWidth={target.width || 3}
              strokeLinecap="round"
            />
            <line
              x1={targetPosition.x - 8}
              y1={targetPosition.y + 8}
              x2={targetPosition.x + 8}
              y2={targetPosition.y - 8}
              stroke={target.color || theme?.colors.semantic.info || '#0ea5e9'}
              strokeWidth={target.width || 3}
              strokeLinecap="round"
            />
          </g>
        )}

        {/* Threshold indicators */}
        {thresholdIndicators.map((threshold, index) => (
          <circle
            key={index}
            cx={threshold.position.x}
            cy={threshold.position.y}
            r={3}
            fill={threshold.color || theme?.colors.text || '#374151'}
          />
        ))}

        {/* Center value display */}
        {appearance.showValue && (
          <text
            x={gaugeParams.centerX}
            y={style === 'half' ? gaugeParams.centerY + 10 : gaugeParams.centerY}
            textAnchor="middle"
            dominantBaseline="central"
            className="text-2xl font-bold"
            fill={theme?.colors.text || '#374151'}
          >
            {dataFormatters.formatNumber(value, valueFormat)}
          </text>
        )}

        {/* Unit label */}
        {labels.unit && appearance.showValue && (
          <text
            x={gaugeParams.centerX}
            y={style === 'half' ? gaugeParams.centerY + 30 : gaugeParams.centerY + 20}
            textAnchor="middle"
            dominantBaseline="central"
            className="text-sm"
            fill={theme?.colors.axis || '#6b7280'}
          >
            {labels.unit}
          </text>
        )}
      </svg>

      {/* Subtitle */}
      {labels.subtitle && (
        <div 
          className="text-sm text-center mt-2"
          style={{ color: theme?.colors.axis }}
        >
          {labels.subtitle}
        </div>
      )}

      {/* Target label */}
      {target?.label && (
        <div 
          className="text-xs text-center mt-1"
          style={{ color: theme?.colors.axis }}
        >
          Target: {dataFormatters.formatNumber(target.value, valueFormat)}
        </div>
      )}

      {/* Threshold labels */}
      {appearance.showThresholds && thresholds.length > 0 && (
        <div className="flex flex-wrap justify-center gap-2 mt-2">
          {thresholds.map((threshold, index) => (
            <div 
              key={index}
              className="text-xs flex items-center gap-1"
              style={{ color: theme?.colors.text }}
            >
              <div
                className="w-2 h-2 rounded-full"
                style={{ backgroundColor: threshold.color || theme?.colors.text }}
              />
              {threshold.label || dataFormatters.formatNumber(threshold.value, valueFormat)}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export { GaugeChart }
export type { GaugeChartProps }