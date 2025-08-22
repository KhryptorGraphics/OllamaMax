/**
 * Time Series Chart Component
 * Displays historical data visualization using Recharts
 */

import React, { useMemo } from 'react'
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine
} from 'recharts'
import { format, parseISO } from 'date-fns'
import { TimeSeriesPoint, ChartType, MetricValue } from '../../types/monitoring'

interface TimeSeriesChartProps {
  data: TimeSeriesPoint[]
  metric?: MetricValue
  title?: string
  chartType?: ChartType
  height?: number
  showGrid?: boolean
  showLegend?: boolean
  showThreshold?: boolean
  colors?: string[]
  timeFormat?: string
  valueFormat?: (value: number) => string
  className?: string
}

interface CustomTooltipProps {
  active?: boolean
  payload?: any[]
  label?: string
  valueFormat?: (value: number) => string
}

const CustomTooltip: React.FC<CustomTooltipProps> = ({
  active,
  payload,
  label,
  valueFormat
}) => {
  if (!active || !payload || payload.length === 0) {
    return null
  }
  
  const formatTime = (timestamp: string) => {
    try {
      return format(parseISO(timestamp), 'MMM dd, HH:mm:ss')
    } catch {
      return timestamp
    }
  }
  
  return (
    <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
      <p className="text-sm font-medium text-gray-900 mb-2">
        {formatTime(label || '')}
      </p>
      {payload.map((entry, index) => (
        <div key={index} className="flex items-center space-x-2">
          <div
            className="w-3 h-3 rounded-full"
            style={{ backgroundColor: entry.color }}
          />
          <span className="text-sm text-gray-700">
            {entry.name}: {valueFormat ? valueFormat(entry.value) : entry.value}
          </span>
        </div>
      ))}
    </div>
  )
}

const formatXAxisTick = (timestamp: string, timeFormat: string) => {
  try {
    return format(parseISO(timestamp), timeFormat)
  } catch {
    return timestamp
  }
}

const formatYAxisTick = (value: number, unit?: string) => {
  if (unit === 'bytes') {
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let unitIndex = 0
    let size = value
    
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024
      unitIndex++
    }
    
    return `${size.toFixed(1)}${units[unitIndex]}`
  }
  
  if (unit === '%') {
    return `${value.toFixed(0)}%`
  }
  
  if (unit === 'ms') {
    return `${value.toFixed(0)}ms`
  }
  
  return value.toFixed(2)
}

export const TimeSeriesChart: React.FC<TimeSeriesChartProps> = ({
  data,
  metric,
  title,
  chartType = 'line',
  height = 300,
  showGrid = true,
  showLegend = true,
  showThreshold = true,
  colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6'],
  timeFormat = 'HH:mm',
  valueFormat,
  className = ''
}) => {
  // Prepare chart data
  const chartData = useMemo(() => {
    return data.map(point => ({
      ...point,
      timestamp: point.timestamp,
      formattedTime: formatXAxisTick(point.timestamp, timeFormat)
    }))
  }, [data, timeFormat])
  
  // Default value formatter
  const defaultValueFormat = useMemo(() => {
    return (value: number) => {
      if (valueFormat) return valueFormat(value)
      return formatYAxisTick(value, metric?.unit)
    }
  }, [valueFormat, metric?.unit])
  
  // Chart component based on type
  const renderChart = () => {
    const commonProps = {
      data: chartData,
      margin: { top: 10, right: 30, left: 20, bottom: 5 }
    }
    
    const xAxisProps = {
      dataKey: 'timestamp',
      tickFormatter: (value: string) => formatXAxisTick(value, timeFormat),
      tick: { fontSize: 12 }
    }
    
    const yAxisProps = {
      tickFormatter: (value: number) => formatYAxisTick(value, metric?.unit),
      tick: { fontSize: 12 }
    }
    
    switch (chartType) {
      case 'area':
        return (
          <AreaChart {...commonProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />}
            <XAxis {...xAxisProps} />
            <YAxis {...yAxisProps} />
            <Tooltip content={<CustomTooltip valueFormat={defaultValueFormat} />} />
            {showLegend && <Legend />}
            {showThreshold && metric?.threshold && (
              <ReferenceLine
                y={metric.threshold}
                stroke="#EF4444"
                strokeDasharray="5 5"
                label="Threshold"
              />
            )}
            <Area
              type="monotone"
              dataKey="value"
              stroke={colors[0]}
              fill={colors[0]}
              fillOpacity={0.3}
              strokeWidth={2}
              name="Value"
            />
          </AreaChart>
        )
      
      case 'bar':
        return (
          <BarChart {...commonProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />}
            <XAxis {...xAxisProps} />
            <YAxis {...yAxisProps} />
            <Tooltip content={<CustomTooltip valueFormat={defaultValueFormat} />} />
            {showLegend && <Legend />}
            {showThreshold && metric?.threshold && (
              <ReferenceLine
                y={metric.threshold}
                stroke="#EF4444"
                strokeDasharray="5 5"
                label="Threshold"
              />
            )}
            <Bar
              dataKey="value"
              fill={colors[0]}
              name="Value"
            />
          </BarChart>
        )
      
      default: // line
        return (
          <LineChart {...commonProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />}
            <XAxis {...xAxisProps} />
            <YAxis {...yAxisProps} />
            <Tooltip content={<CustomTooltip valueFormat={defaultValueFormat} />} />
            {showLegend && <Legend />}
            {showThreshold && metric?.threshold && (
              <ReferenceLine
                y={metric.threshold}
                stroke="#EF4444"
                strokeDasharray="5 5"
                label="Threshold"
              />
            )}
            <Line
              type="monotone"
              dataKey="value"
              stroke={colors[0]}
              strokeWidth={2}
              dot={false}
              name="Value"
            />
          </LineChart>
        )
    }
  }
  
  if (!data || data.length === 0) {
    return (
      <div className={`${className}`}>
        {title && (
          <h3 className="text-lg font-medium text-gray-900 mb-4">{title}</h3>
        )}
        <div 
          className="flex items-center justify-center bg-gray-50 border-2 border-dashed border-gray-300 rounded-lg"
          style={{ height: `${height}px` }}
        >
          <div className="text-center">
            <div className="text-gray-400 text-lg mb-2">📊</div>
            <div className="text-gray-500">No data available</div>
          </div>
        </div>
      </div>
    )
  }
  
  return (
    <div className={`${className}`}>
      {title && (
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">{title}</h3>
          {metric && (
            <div className="flex items-center space-x-4 text-sm text-gray-600">
              <span>Current: {defaultValueFormat(metric.current)}</span>
              <span>Avg: {defaultValueFormat(metric.average)}</span>
              <span>Peak: {defaultValueFormat(metric.peak)}</span>
            </div>
          )}
        </div>
      )}
      
      <div className="bg-white p-4 border border-gray-200 rounded-lg">
        <ResponsiveContainer width="100%" height={height}>
          {renderChart()}
        </ResponsiveContainer>
      </div>
      
      {metric?.trend && (
        <div className="mt-2 text-sm text-gray-600">
          Trend: <span className={`font-medium ${
            metric.trend === 'up' ? 'text-green-600' :
            metric.trend === 'down' ? 'text-red-600' :
            'text-gray-600'
          }`}>
            {metric.trend === 'up' ? '↗ Increasing' :
             metric.trend === 'down' ? '↘ Decreasing' :
             '→ Stable'}
          </span>
        </div>
      )}
    </div>
  )
}

// Multi-series chart component
interface MultiSeriesChartProps extends Omit<TimeSeriesChartProps, 'data' | 'metric'> {
  series: Array<{
    name: string
    data: TimeSeriesPoint[]
    color?: string
    metric?: MetricValue
  }>
}

export const MultiSeriesChart: React.FC<MultiSeriesChartProps> = ({
  series,
  title,
  chartType = 'line',
  height = 300,
  showGrid = true,
  showLegend = true,
  colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6'],
  timeFormat = 'HH:mm',
  valueFormat,
  className = ''
}) => {
  // Combine all series data
  const chartData = useMemo(() => {
    const allTimestamps = new Set<string>()
    series.forEach(s => s.data.forEach(point => allTimestamps.add(point.timestamp)))
    
    const sortedTimestamps = Array.from(allTimestamps).sort()
    
    return sortedTimestamps.map(timestamp => {
      const dataPoint: any = { timestamp }
      
      series.forEach(s => {
        const point = s.data.find(p => p.timestamp === timestamp)
        dataPoint[s.name] = point?.value || null
      })
      
      return dataPoint
    })
  }, [series])
  
  const renderMultiChart = () => {
    const commonProps = {
      data: chartData,
      margin: { top: 10, right: 30, left: 20, bottom: 5 }
    }
    
    const xAxisProps = {
      dataKey: 'timestamp',
      tickFormatter: (value: string) => formatXAxisTick(value, timeFormat),
      tick: { fontSize: 12 }
    }
    
    const yAxisProps = {
      tickFormatter: (value: number) => formatYAxisTick(value),
      tick: { fontSize: 12 }
    }
    
    switch (chartType) {
      case 'area':
        return (
          <AreaChart {...commonProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />}
            <XAxis {...xAxisProps} />
            <YAxis {...yAxisProps} />
            <Tooltip content={<CustomTooltip valueFormat={valueFormat} />} />
            {showLegend && <Legend />}
            {series.map((s, index) => (
              <Area
                key={s.name}
                type="monotone"
                dataKey={s.name}
                stroke={s.color || colors[index % colors.length]}
                fill={s.color || colors[index % colors.length]}
                fillOpacity={0.3}
                strokeWidth={2}
              />
            ))}
          </AreaChart>
        )
      
      default: // line
        return (
          <LineChart {...commonProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />}
            <XAxis {...xAxisProps} />
            <YAxis {...yAxisProps} />
            <Tooltip content={<CustomTooltip valueFormat={valueFormat} />} />
            {showLegend && <Legend />}
            {series.map((s, index) => (
              <Line
                key={s.name}
                type="monotone"
                dataKey={s.name}
                stroke={s.color || colors[index % colors.length]}
                strokeWidth={2}
                dot={false}
                connectNulls={false}
              />
            ))}
          </LineChart>
        )
    }
  }
  
  return (
    <div className={`${className}`}>
      {title && (
        <h3 className="text-lg font-medium text-gray-900 mb-4">{title}</h3>
      )}
      
      <div className="bg-white p-4 border border-gray-200 rounded-lg">
        <ResponsiveContainer width="100%" height={height}>
          {renderMultiChart()}
        </ResponsiveContainer>
      </div>
    </div>
  )
}