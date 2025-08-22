/**
 * Chart Utilities
 * Comprehensive utilities for data visualization including formatting, color generation,
 * statistical calculations, and export functionality
 */

import { format, parseISO, subDays, subHours, subMinutes } from 'date-fns'
import jsPDF from 'jspdf'
import { colors, semanticColors } from '@/design-system/tokens/colors'

// Types for chart data
export interface ChartDataPoint {
  timestamp?: string | Date
  value: number | null
  label?: string
  category?: string
  [key: string]: any
}

export interface TimeSeriesDataPoint extends ChartDataPoint {
  timestamp: string | Date
}

export interface CategoryDataPoint extends ChartDataPoint {
  label: string
  category?: string
}

export interface MultiSeriesDataPoint {
  timestamp?: string | Date
  label?: string
  [metricName: string]: number | string | Date | null | undefined
}

// Chart theme configuration
export interface ChartTheme {
  mode: 'light' | 'dark'
  colors: {
    primary: string[]
    semantic: {
      success: string
      warning: string
      error: string
      info: string
    }
    neutral: string[]
    background: string
    grid: string
    text: string
    axis: string
  }
}

// Data formatting utilities
export const dataFormatters = {
  /**
   * Format numbers with appropriate units and precision
   */
  formatNumber: (value: number, options: {
    precision?: number
    unit?: string
    format?: 'default' | 'percentage' | 'bytes' | 'duration' | 'currency'
  } = {}): string => {
    const { precision = 2, unit = '', format = 'default' } = options

    if (value === null || value === undefined || isNaN(value)) {
      return 'N/A'
    }

    switch (format) {
      case 'percentage':
        return `${(value * 100).toFixed(precision)}%`
      
      case 'bytes':
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
        if (value === 0) return '0 B'
        const i = Math.floor(Math.log(value) / Math.log(1024))
        return `${(value / Math.pow(1024, i)).toFixed(precision)} ${sizes[i]}`
      
      case 'duration':
        if (value < 1000) return `${value.toFixed(precision)}ms`
        if (value < 60000) return `${(value / 1000).toFixed(precision)}s`
        if (value < 3600000) return `${(value / 60000).toFixed(precision)}m`
        return `${(value / 3600000).toFixed(precision)}h`
      
      case 'currency':
        return new Intl.NumberFormat('en-US', {
          style: 'currency',
          currency: 'USD',
          minimumFractionDigits: precision,
          maximumFractionDigits: precision
        }).format(value)
      
      default:
        const formatted = value.toLocaleString('en-US', {
          minimumFractionDigits: precision,
          maximumFractionDigits: precision
        })
        return unit ? `${formatted} ${unit}` : formatted
    }
  },

  /**
   * Format timestamps for different time ranges
   */
  formatTimestamp: (timestamp: string | Date, range: 'minute' | 'hour' | 'day' | 'week' | 'month' = 'hour'): string => {
    const date = typeof timestamp === 'string' ? parseISO(timestamp) : timestamp

    switch (range) {
      case 'minute':
        return format(date, 'HH:mm:ss')
      case 'hour':
        return format(date, 'HH:mm')
      case 'day':
        return format(date, 'MMM dd HH:mm')
      case 'week':
        return format(date, 'MMM dd')
      case 'month':
        return format(date, 'MMM yyyy')
      default:
        return format(date, 'MMM dd HH:mm')
    }
  },

  /**
   * Generate time series labels based on range
   */
  generateTimeLabels: (range: 'hour' | 'day' | 'week' | 'month', count: number = 24): string[] => {
    const labels: string[] = []
    const now = new Date()

    for (let i = count - 1; i >= 0; i--) {
      let date: Date
      
      switch (range) {
        case 'hour':
          date = subMinutes(now, i * 5) // 5-minute intervals
          labels.push(format(date, 'HH:mm'))
          break
        case 'day':
          date = subHours(now, i)
          labels.push(format(date, 'HH:mm'))
          break
        case 'week':
          date = subDays(now, i)
          labels.push(format(date, 'MMM dd'))
          break
        case 'month':
          date = subDays(now, i)
          labels.push(format(date, 'MMM dd'))
          break
      }
    }

    return labels
  }
}

// Color palette generation
export const colorUtils = {
  /**
   * Generate chart theme based on design system
   */
  generateChartTheme: (mode: 'light' | 'dark' = 'light'): ChartTheme => {
    const themeColors = semanticColors[mode]
    
    return {
      mode,
      colors: {
        primary: [
          colors.primary[500],
          colors.primary[400],
          colors.primary[600],
          colors.primary[300],
          colors.primary[700],
          colors.secondary[500],
          colors.secondary[400],
          colors.secondary[600]
        ],
        semantic: {
          success: colors.success[500],
          warning: colors.warning[500],
          error: colors.error[500],
          info: colors.info[500]
        },
        neutral: [
          colors.neutral[500],
          colors.neutral[400],
          colors.neutral[600],
          colors.neutral[300],
          colors.neutral[700]
        ],
        background: themeColors.background.primary,
        grid: themeColors.border.primary,
        text: themeColors.text.primary,
        axis: themeColors.text.secondary
      }
    }
  },

  /**
   * Generate color palette for multiple series
   */
  generateColorPalette: (count: number, theme: ChartTheme): string[] => {
    const { primary, neutral } = theme.colors
    const allColors = [...primary, ...neutral]
    
    if (count <= allColors.length) {
      return allColors.slice(0, count)
    }

    // Generate additional colors by adjusting lightness
    const colors: string[] = [...allColors]
    const baseColors = primary
    
    while (colors.length < count) {
      baseColors.forEach((color, index) => {
        if (colors.length >= count) return
        // Add lighter/darker variants
        colors.push(adjustColorOpacity(color, 0.7))
      })
    }

    return colors.slice(0, count)
  },

  /**
   * Get semantic color based on value and thresholds
   */
  getSemanticColor: (value: number, thresholds: {
    error?: number
    warning?: number
    success?: number
  }, theme: ChartTheme): string => {
    const { semantic } = theme.colors
    
    if (thresholds.error !== undefined && value >= thresholds.error) {
      return semantic.error
    }
    if (thresholds.warning !== undefined && value >= thresholds.warning) {
      return semantic.warning
    }
    if (thresholds.success !== undefined && value >= thresholds.success) {
      return semantic.success
    }
    
    return theme.colors.primary[0]
  }
}

// Statistical calculations
export const statsUtils = {
  /**
   * Calculate basic statistics for a dataset
   */
  calculateStats: (data: number[]): {
    min: number
    max: number
    mean: number
    median: number
    sum: number
    count: number
    variance: number
    stdDev: number
  } => {
    const validData = data.filter(value => value !== null && value !== undefined && !isNaN(value))
    
    if (validData.length === 0) {
      return {
        min: 0, max: 0, mean: 0, median: 0, sum: 0, count: 0, variance: 0, stdDev: 0
      }
    }

    const sorted = [...validData].sort((a, b) => a - b)
    const sum = validData.reduce((acc, val) => acc + val, 0)
    const mean = sum / validData.length
    
    const variance = validData.reduce((acc, val) => acc + Math.pow(val - mean, 2), 0) / validData.length
    const stdDev = Math.sqrt(variance)
    
    const median = validData.length % 2 === 0
      ? (sorted[validData.length / 2 - 1] + sorted[validData.length / 2]) / 2
      : sorted[Math.floor(validData.length / 2)]

    return {
      min: Math.min(...validData),
      max: Math.max(...validData),
      mean,
      median,
      sum,
      count: validData.length,
      variance,
      stdDev
    }
  },

  /**
   * Calculate moving average
   */
  calculateMovingAverage: (data: number[], window: number): number[] => {
    const result: number[] = []
    
    for (let i = 0; i < data.length; i++) {
      const start = Math.max(0, i - window + 1)
      const subset = data.slice(start, i + 1)
      const validSubset = subset.filter(val => val !== null && val !== undefined && !isNaN(val))
      
      if (validSubset.length > 0) {
        const average = validSubset.reduce((sum, val) => sum + val, 0) / validSubset.length
        result.push(average)
      } else {
        result.push(0)
      }
    }
    
    return result
  },

  /**
   * Calculate percentiles
   */
  calculatePercentile: (data: number[], percentile: number): number => {
    const validData = data.filter(value => value !== null && value !== undefined && !isNaN(value))
    
    if (validData.length === 0) return 0
    
    const sorted = [...validData].sort((a, b) => a - b)
    const index = (percentile / 100) * (sorted.length - 1)
    
    if (Number.isInteger(index)) {
      return sorted[index]
    }
    
    const lower = Math.floor(index)
    const upper = Math.ceil(index)
    const weight = index - lower
    
    return sorted[lower] * (1 - weight) + sorted[upper] * weight
  }
}

// Time series aggregation
export const timeSeriesUtils = {
  /**
   * Aggregate time series data by time interval
   */
  aggregateByInterval: (
    data: TimeSeriesDataPoint[],
    interval: 'minute' | 'hour' | 'day',
    aggregation: 'sum' | 'avg' | 'min' | 'max' | 'count' = 'avg'
  ): TimeSeriesDataPoint[] => {
    const groupedData = new Map<string, number[]>()
    
    data.forEach(point => {
      const date = typeof point.timestamp === 'string' ? parseISO(point.timestamp) : point.timestamp
      let key: string
      
      switch (interval) {
        case 'minute':
          key = format(date, 'yyyy-MM-dd HH:mm')
          break
        case 'hour':
          key = format(date, 'yyyy-MM-dd HH')
          break
        case 'day':
          key = format(date, 'yyyy-MM-dd')
          break
      }
      
      if (!groupedData.has(key)) {
        groupedData.set(key, [])
      }
      
      if (point.value !== null && point.value !== undefined) {
        groupedData.get(key)!.push(point.value)
      }
    })
    
    const result: TimeSeriesDataPoint[] = []
    
    groupedData.forEach((values, key) => {
      let aggregatedValue: number
      
      switch (aggregation) {
        case 'sum':
          aggregatedValue = values.reduce((sum, val) => sum + val, 0)
          break
        case 'min':
          aggregatedValue = Math.min(...values)
          break
        case 'max':
          aggregatedValue = Math.max(...values)
          break
        case 'count':
          aggregatedValue = values.length
          break
        case 'avg':
        default:
          aggregatedValue = values.reduce((sum, val) => sum + val, 0) / values.length
          break
      }
      
      result.push({
        timestamp: key,
        value: aggregatedValue
      })
    })
    
    return result.sort((a, b) => 
      new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    )
  },

  /**
   * Fill missing time points in time series data
   */
  fillMissingTimePoints: (
    data: TimeSeriesDataPoint[],
    interval: 'minute' | 'hour' | 'day',
    fillValue: number | null = null
  ): TimeSeriesDataPoint[] => {
    if (data.length === 0) return []
    
    const sortedData = [...data].sort((a, b) => 
      new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    )
    
    const start = new Date(sortedData[0].timestamp)
    const end = new Date(sortedData[sortedData.length - 1].timestamp)
    const result: TimeSeriesDataPoint[] = []
    
    const current = new Date(start)
    const dataMap = new Map(
      sortedData.map(point => [
        format(new Date(point.timestamp), 'yyyy-MM-dd HH:mm:ss'),
        point
      ])
    )
    
    while (current <= end) {
      const key = format(current, 'yyyy-MM-dd HH:mm:ss')
      const existingPoint = dataMap.get(key)
      
      if (existingPoint) {
        result.push(existingPoint)
      } else {
        result.push({
          timestamp: new Date(current),
          value: fillValue
        })
      }
      
      switch (interval) {
        case 'minute':
          current.setMinutes(current.getMinutes() + 1)
          break
        case 'hour':
          current.setHours(current.getHours() + 1)
          break
        case 'day':
          current.setDate(current.getDate() + 1)
          break
      }
    }
    
    return result
  }
}

// Export utilities
export const exportUtils = {
  /**
   * Export chart as PNG
   */
  exportAsPNG: async (chartElement: HTMLElement, filename: string = 'chart.png'): Promise<void> => {
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    
    if (!ctx) throw new Error('Could not get canvas context')
    
    const rect = chartElement.getBoundingClientRect()
    canvas.width = rect.width * 2 // High DPI
    canvas.height = rect.height * 2
    canvas.style.width = `${rect.width}px`
    canvas.style.height = `${rect.height}px`
    
    ctx.scale(2, 2)
    
    // Convert SVG to canvas (simplified approach)
    const svgElement = chartElement.querySelector('svg')
    if (svgElement) {
      const svgData = new XMLSerializer().serializeToString(svgElement)
      const img = new Image()
      
      return new Promise((resolve, reject) => {
        img.onload = () => {
          ctx.fillStyle = '#ffffff'
          ctx.fillRect(0, 0, canvas.width, canvas.height)
          ctx.drawImage(img, 0, 0)
          
          canvas.toBlob(blob => {
            if (blob) {
              const url = URL.createObjectURL(blob)
              const a = document.createElement('a')
              a.href = url
              a.download = filename
              a.click()
              URL.revokeObjectURL(url)
              resolve()
            } else {
              reject(new Error('Failed to create blob'))
            }
          }, 'image/png')
        }
        
        img.onerror = reject
        img.src = `data:image/svg+xml;base64,${btoa(svgData)}`
      })
    }
  },

  /**
   * Export chart as SVG
   */
  exportAsSVG: (chartElement: HTMLElement, filename: string = 'chart.svg'): void => {
    const svgElement = chartElement.querySelector('svg')
    if (!svgElement) throw new Error('No SVG element found')
    
    const svgData = new XMLSerializer().serializeToString(svgElement)
    const blob = new Blob([svgData], { type: 'image/svg+xml' })
    const url = URL.createObjectURL(blob)
    
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  },

  /**
   * Export chart data as CSV
   */
  exportAsCSV: (data: any[], filename: string = 'chart-data.csv'): void => {
    if (data.length === 0) return
    
    const headers = Object.keys(data[0])
    const csvContent = [
      headers.join(','),
      ...data.map(row => 
        headers.map(header => {
          const value = row[header]
          return typeof value === 'string' && value.includes(',') 
            ? `"${value}"` 
            : value
        }).join(',')
      )
    ].join('\n')
    
    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  },

  /**
   * Export chart as PDF
   */
  exportAsPDF: async (chartElement: HTMLElement, filename: string = 'chart.pdf'): Promise<void> => {
    const pdf = new jsPDF({
      orientation: 'landscape',
      unit: 'mm',
      format: 'a4'
    })
    
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    
    if (!ctx) throw new Error('Could not get canvas context')
    
    const rect = chartElement.getBoundingClientRect()
    canvas.width = rect.width * 2
    canvas.height = rect.height * 2
    
    ctx.scale(2, 2)
    
    const svgElement = chartElement.querySelector('svg')
    if (svgElement) {
      const svgData = new XMLSerializer().serializeToString(svgElement)
      const img = new Image()
      
      return new Promise((resolve, reject) => {
        img.onload = () => {
          ctx.fillStyle = '#ffffff'
          ctx.fillRect(0, 0, canvas.width, canvas.height)
          ctx.drawImage(img, 0, 0)
          
          const imgData = canvas.toDataURL('image/png')
          const pdfWidth = pdf.internal.pageSize.getWidth()
          const pdfHeight = pdf.internal.pageSize.getHeight()
          
          const imgWidth = pdfWidth - 20 // 10mm margin on each side
          const imgHeight = (rect.height * imgWidth) / rect.width
          
          const x = 10 // 10mm left margin
          const y = (pdfHeight - imgHeight) / 2 // Center vertically
          
          pdf.addImage(imgData, 'PNG', x, y, imgWidth, imgHeight)
          pdf.save(filename)
          resolve()
        }
        
        img.onerror = reject
        img.src = `data:image/svg+xml;base64,${btoa(svgData)}`
      })
    }
  }
}

// Helper functions
const adjustColorOpacity = (color: string, opacity: number): string => {
  // Simple opacity adjustment for hex colors
  if (color.startsWith('#')) {
    const hex = color.slice(1)
    const alpha = Math.round(opacity * 255).toString(16).padStart(2, '0')
    return `${color}${alpha}`
  }
  return color
}

// Responsive utilities
export const responsiveUtils = {
  /**
   * Get responsive chart dimensions based on container
   */
  getResponsiveDimensions: (container: HTMLElement): { width: number; height: number } => {
    const rect = container.getBoundingClientRect()
    const aspectRatio = 16 / 9 // Default aspect ratio
    
    let width = rect.width
    let height = rect.height
    
    // If height is not explicitly set, calculate from aspect ratio
    if (height === 0 || height < width / aspectRatio) {
      height = width / aspectRatio
    }
    
    // Ensure minimum dimensions
    width = Math.max(width, 300)
    height = Math.max(height, 200)
    
    return { width, height }
  },

  /**
   * Get responsive margins based on screen size
   */
  getResponsiveMargins: (): { top: number; right: number; bottom: number; left: number } => {
    const width = window.innerWidth
    
    if (width < 640) {
      // Mobile
      return { top: 20, right: 20, bottom: 40, left: 40 }
    } else if (width < 1024) {
      // Tablet
      return { top: 30, right: 30, bottom: 50, left: 60 }
    } else {
      // Desktop
      return { top: 40, right: 40, bottom: 60, left: 80 }
    }
  }
}

export {
  type ChartDataPoint,
  type TimeSeriesDataPoint,
  type CategoryDataPoint,
  type MultiSeriesDataPoint,
  type ChartTheme
}