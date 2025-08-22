/**
 * Data Aggregation and Filtering Utilities
 * Advanced data processing for analytics and reporting
 */

import { 
  AnalyticsEvent, 
  BusinessMetrics, 
  DateRange, 
  ReportFilter, 
  FilterOperator,
  AggregationType 
} from '../types'

// Time period utilities
export enum TimePeriod {
  HOUR = 'hour',
  DAY = 'day',
  WEEK = 'week',
  MONTH = 'month',
  QUARTER = 'quarter',
  YEAR = 'year'
}

export interface AggregationConfig {
  groupBy: string[]
  metrics: AggregationMetric[]
  filters?: ReportFilter[]
  dateRange?: DateRange
  timePeriod?: TimePeriod
  limit?: number
  orderBy?: { field: string; direction: 'asc' | 'desc' }
}

export interface AggregationMetric {
  field: string
  aggregation: AggregationType
  alias?: string
}

export interface AggregationResult {
  groups: Record<string, any>[]
  totals: Record<string, number>
  summary: {
    totalRecords: number
    filteredRecords: number
    groupCount: number
    processingTime: number
  }
}

export class DataAggregator {
  private static instance: DataAggregator

  static getInstance(): DataAggregator {
    if (!DataAggregator.instance) {
      DataAggregator.instance = new DataAggregator()
    }
    return DataAggregator.instance
  }

  /**
   * Main aggregation method
   */
  aggregate<T = any>(data: T[], config: AggregationConfig): AggregationResult {
    const startTime = performance.now()
    
    // Apply filters first
    let filteredData = this.applyFilters(data, config.filters || [])
    
    // Apply date range filter if specified
    if (config.dateRange) {
      filteredData = this.applyDateRangeFilter(filteredData, config.dateRange)
    }

    // Group data
    const grouped = this.groupData(filteredData, config.groupBy)
    
    // Calculate aggregations
    const results = this.calculateAggregations(grouped, config.metrics)
    
    // Apply sorting
    if (config.orderBy) {
      results.sort((a, b) => {
        const aValue = a[config.orderBy!.field]
        const bValue = b[config.orderBy!.field]
        const multiplier = config.orderBy!.direction === 'asc' ? 1 : -1
        
        if (typeof aValue === 'number' && typeof bValue === 'number') {
          return (aValue - bValue) * multiplier
        }
        return String(aValue).localeCompare(String(bValue)) * multiplier
      })
    }

    // Apply limit
    const limitedResults = config.limit ? results.slice(0, config.limit) : results
    
    // Calculate totals
    const totals = this.calculateTotals(results, config.metrics)
    
    const endTime = performance.now()

    return {
      groups: limitedResults,
      totals,
      summary: {
        totalRecords: data.length,
        filteredRecords: filteredData.length,
        groupCount: results.length,
        processingTime: endTime - startTime
      }
    }
  }

  /**
   * Time-series aggregation for trend analysis
   */
  aggregateTimeSeries<T = any>(
    data: T[],
    dateField: string,
    valueField: string,
    period: TimePeriod,
    dateRange?: DateRange
  ): { timestamp: string; value: number }[] {
    let filteredData = data

    // Apply date range if specified
    if (dateRange) {
      filteredData = data.filter(item => {
        const itemDate = this.extractDate(item, dateField)
        return itemDate >= dateRange.start && itemDate <= dateRange.end
      })
    }

    // Group by time period
    const grouped = new Map<string, T[]>()
    
    filteredData.forEach(item => {
      const date = this.extractDate(item, dateField)
      const periodKey = this.formatDateToPeriod(date, period)
      
      if (!grouped.has(periodKey)) {
        grouped.set(periodKey, [])
      }
      grouped.get(periodKey)!.push(item)
    })

    // Calculate values for each period
    const results = Array.from(grouped.entries()).map(([timestamp, items]) => ({
      timestamp,
      value: this.sumField(items, valueField)
    }))

    // Sort by timestamp
    results.sort((a, b) => a.timestamp.localeCompare(b.timestamp))

    return results
  }

  /**
   * Cohort analysis for retention metrics
   */
  analyzeCohorts<T = any>(
    data: T[],
    cohortDateField: string,
    returnDateField: string,
    userIdField: string,
    periods: number[] = [1, 7, 30, 90]
  ): Array<{
    cohort: string
    size: number
    retention: Array<{ period: number; retained: number; percentage: number }>
  }> {
    // Group users by cohort (first activity month)
    const cohorts = new Map<string, Set<string>>()
    const userFirstSeen = new Map<string, number>()

    data.forEach(item => {
      const userId = String(item[userIdField as keyof T])
      const date = this.extractDate(item, cohortDateField)
      
      if (!userFirstSeen.has(userId) || date < userFirstSeen.get(userId)!) {
        userFirstSeen.set(userId, date)
      }
    })

    // Create cohorts based on first seen month
    userFirstSeen.forEach((firstDate, userId) => {
      const cohortKey = this.formatDateToPeriod(firstDate, TimePeriod.MONTH)
      
      if (!cohorts.has(cohortKey)) {
        cohorts.set(cohortKey, new Set())
      }
      cohorts.get(cohortKey)!.add(userId)
    })

    // Calculate retention for each cohort
    const results = Array.from(cohorts.entries()).map(([cohort, users]) => {
      const cohortSize = users.size
      const cohortDate = new Date(cohort + '-01').getTime()
      
      const retention = periods.map(periodDays => {
        const periodEnd = cohortDate + (periodDays * 24 * 60 * 60 * 1000)
        
        // Count users who returned within the period
        const returnedUsers = new Set<string>()
        
        data.forEach(item => {
          const userId = String(item[userIdField as keyof T])
          const returnDate = this.extractDate(item, returnDateField)
          
          if (users.has(userId) && 
              returnDate > cohortDate && 
              returnDate <= periodEnd) {
            returnedUsers.add(userId)
          }
        })

        return {
          period: periodDays,
          retained: returnedUsers.size,
          percentage: (returnedUsers.size / cohortSize) * 100
        }
      })

      return {
        cohort,
        size: cohortSize,
        retention
      }
    })

    return results.sort((a, b) => a.cohort.localeCompare(b.cohort))
  }

  /**
   * Funnel analysis for conversion tracking
   */
  analyzeFunnel<T = any>(
    data: T[],
    steps: Array<{ name: string; condition: (item: T) => boolean }>,
    userIdField: string
  ): Array<{
    step: string
    users: number
    conversionRate: number
    dropOffRate: number
  }> {
    const userProgression = new Map<string, number>()
    
    // Track each user's progression through the funnel
    data.forEach(item => {
      const userId = String(item[userIdField as keyof T])
      const currentStep = userProgression.get(userId) || 0
      
      // Check if user qualifies for next steps
      for (let i = currentStep; i < steps.length; i++) {
        if (steps[i].condition(item)) {
          userProgression.set(userId, i + 1)
        } else {
          break
        }
      }
    })

    // Calculate metrics for each step
    const totalUsers = userProgression.size
    const results = steps.map((step, index) => {
      const usersAtStep = Array.from(userProgression.values()).filter(
        stepReached => stepReached > index
      ).length
      
      const conversionRate = totalUsers > 0 ? (usersAtStep / totalUsers) * 100 : 0
      const dropOffRate = index > 0 
        ? 100 - conversionRate 
        : 0

      return {
        step: step.name,
        users: usersAtStep,
        conversionRate,
        dropOffRate
      }
    })

    return results
  }

  /**
   * Statistical analysis utilities
   */
  calculateStatistics(values: number[]): {
    count: number
    sum: number
    mean: number
    median: number
    mode: number[]
    min: number
    max: number
    variance: number
    standardDeviation: number
    percentiles: { p25: number; p50: number; p75: number; p95: number; p99: number }
  } {
    if (values.length === 0) {
      return {
        count: 0, sum: 0, mean: 0, median: 0, mode: [],
        min: 0, max: 0, variance: 0, standardDeviation: 0,
        percentiles: { p25: 0, p50: 0, p75: 0, p95: 0, p99: 0 }
      }
    }

    const sorted = [...values].sort((a, b) => a - b)
    const count = values.length
    const sum = values.reduce((acc, val) => acc + val, 0)
    const mean = sum / count

    // Median
    const median = count % 2 === 0
      ? (sorted[count / 2 - 1] + sorted[count / 2]) / 2
      : sorted[Math.floor(count / 2)]

    // Mode
    const frequency = new Map<number, number>()
    values.forEach(val => {
      frequency.set(val, (frequency.get(val) || 0) + 1)
    })
    const maxFreq = Math.max(...frequency.values())
    const mode = Array.from(frequency.entries())
      .filter(([_, freq]) => freq === maxFreq)
      .map(([val, _]) => val)

    // Min/Max
    const min = Math.min(...values)
    const max = Math.max(...values)

    // Variance and Standard Deviation
    const variance = values.reduce((acc, val) => acc + Math.pow(val - mean, 2), 0) / count
    const standardDeviation = Math.sqrt(variance)

    // Percentiles
    const percentiles = {
      p25: this.calculatePercentile(sorted, 25),
      p50: this.calculatePercentile(sorted, 50),
      p75: this.calculatePercentile(sorted, 75),
      p95: this.calculatePercentile(sorted, 95),
      p99: this.calculatePercentile(sorted, 99)
    }

    return {
      count, sum, mean, median, mode, min, max,
      variance, standardDeviation, percentiles
    }
  }

  /**
   * Anomaly detection using statistical methods
   */
  detectAnomalies(
    values: number[],
    method: 'zscore' | 'iqr' | 'isolation' = 'zscore',
    threshold: number = 2
  ): { value: number; index: number; score: number }[] {
    const stats = this.calculateStatistics(values)
    const anomalies: { value: number; index: number; score: number }[] = []

    switch (method) {
      case 'zscore':
        values.forEach((value, index) => {
          const zScore = Math.abs((value - stats.mean) / stats.standardDeviation)
          if (zScore > threshold) {
            anomalies.push({ value, index, score: zScore })
          }
        })
        break

      case 'iqr':
        const iqr = stats.percentiles.p75 - stats.percentiles.p25
        const lowerBound = stats.percentiles.p25 - (1.5 * iqr)
        const upperBound = stats.percentiles.p75 + (1.5 * iqr)
        
        values.forEach((value, index) => {
          if (value < lowerBound || value > upperBound) {
            const score = Math.min(
              Math.abs(value - lowerBound) / iqr,
              Math.abs(value - upperBound) / iqr
            )
            anomalies.push({ value, index, score })
          }
        })
        break
    }

    return anomalies.sort((a, b) => b.score - a.score)
  }

  // Private helper methods
  private applyFilters<T>(data: T[], filters: ReportFilter[]): T[] {
    return data.filter(item => {
      return filters.every(filter => {
        const value = this.getNestedValue(item, filter.field)
        return this.evaluateFilter(value, filter.operator, filter.value)
      })
    })
  }

  private applyDateRangeFilter<T>(data: T[], dateRange: DateRange): T[] {
    return data.filter(item => {
      // Assume items have a timestamp field
      const timestamp = (item as any).timestamp || (item as any).createdAt || (item as any).date
      if (!timestamp) return true
      
      const date = typeof timestamp === 'number' ? timestamp : new Date(timestamp).getTime()
      return date >= dateRange.start && date <= dateRange.end
    })
  }

  private groupData<T>(data: T[], groupBy: string[]): Map<string, T[]> {
    const groups = new Map<string, T[]>()
    
    data.forEach(item => {
      const key = groupBy.map(field => 
        String(this.getNestedValue(item, field) || 'null')
      ).join('|')
      
      if (!groups.has(key)) {
        groups.set(key, [])
      }
      groups.get(key)!.push(item)
    })
    
    return groups
  }

  private calculateAggregations<T>(
    groups: Map<string, T[]>,
    metrics: AggregationMetric[]
  ): Record<string, any>[] {
    return Array.from(groups.entries()).map(([groupKey, items]) => {
      const result: Record<string, any> = {}
      
      // Add group fields
      const keyParts = groupKey.split('|')
      keyParts.forEach((part, index) => {
        result[`group_${index}`] = part === 'null' ? null : part
      })
      
      // Calculate aggregations
      metrics.forEach(metric => {
        const values = items
          .map(item => this.getNestedValue(item, metric.field))
          .filter(val => val !== null && val !== undefined && !isNaN(Number(val)))
          .map(val => Number(val))
        
        const fieldName = metric.alias || `${metric.field}_${metric.aggregation}`
        
        switch (metric.aggregation) {
          case 'sum':
            result[fieldName] = values.reduce((acc, val) => acc + val, 0)
            break
          case 'average':
            result[fieldName] = values.length > 0 ? values.reduce((acc, val) => acc + val, 0) / values.length : 0
            break
          case 'count':
            result[fieldName] = items.length
            break
          case 'distinct':
            result[fieldName] = new Set(values).size
            break
          case 'min':
            result[fieldName] = values.length > 0 ? Math.min(...values) : 0
            break
          case 'max':
            result[fieldName] = values.length > 0 ? Math.max(...values) : 0
            break
          case 'median':
            const sorted = values.sort((a, b) => a - b)
            result[fieldName] = sorted.length > 0 ? this.calculatePercentile(sorted, 50) : 0
            break
          case 'percentile':
            // Default to 95th percentile
            const sortedValues = values.sort((a, b) => a - b)
            result[fieldName] = sortedValues.length > 0 ? this.calculatePercentile(sortedValues, 95) : 0
            break
        }
      })
      
      return result
    })
  }

  private calculateTotals(
    results: Record<string, any>[],
    metrics: AggregationMetric[]
  ): Record<string, number> {
    const totals: Record<string, number> = {}
    
    metrics.forEach(metric => {
      const fieldName = metric.alias || `${metric.field}_${metric.aggregation}`
      const values = results.map(r => Number(r[fieldName]) || 0)
      
      switch (metric.aggregation) {
        case 'sum':
        case 'count':
        case 'distinct':
          totals[fieldName] = values.reduce((acc, val) => acc + val, 0)
          break
        case 'average':
        case 'median':
        case 'percentile':
          totals[fieldName] = values.length > 0 ? values.reduce((acc, val) => acc + val, 0) / values.length : 0
          break
        case 'min':
          totals[fieldName] = values.length > 0 ? Math.min(...values) : 0
          break
        case 'max':
          totals[fieldName] = values.length > 0 ? Math.max(...values) : 0
          break
      }
    })
    
    return totals
  }

  private evaluateFilter(value: any, operator: FilterOperator, filterValue: any): boolean {
    const val = String(value || '').toLowerCase()
    const filter = String(filterValue || '').toLowerCase()
    
    switch (operator) {
      case 'equals':
        return val === filter
      case 'not_equals':
        return val !== filter
      case 'contains':
        return val.includes(filter)
      case 'not_contains':
        return !val.includes(filter)
      case 'starts_with':
        return val.startsWith(filter)
      case 'ends_with':
        return val.endsWith(filter)
      case 'greater_than':
        return Number(value) > Number(filterValue)
      case 'less_than':
        return Number(value) < Number(filterValue)
      case 'between':
        if (Array.isArray(filterValue) && filterValue.length === 2) {
          return Number(value) >= Number(filterValue[0]) && Number(value) <= Number(filterValue[1])
        }
        return false
      case 'in':
        return Array.isArray(filterValue) && filterValue.includes(value)
      case 'not_in':
        return !Array.isArray(filterValue) || !filterValue.includes(value)
      default:
        return true
    }
  }

  private getNestedValue(obj: any, path: string): any {
    return path.split('.').reduce((current, key) => {
      return current && typeof current === 'object' ? current[key] : undefined
    }, obj)
  }

  private extractDate(item: any, dateField: string): number {
    const value = this.getNestedValue(item, dateField)
    if (typeof value === 'number') return value
    if (typeof value === 'string') return new Date(value).getTime()
    if (value instanceof Date) return value.getTime()
    return Date.now()
  }

  private formatDateToPeriod(timestamp: number, period: TimePeriod): string {
    const date = new Date(timestamp)
    
    switch (period) {
      case TimePeriod.HOUR:
        return `${date.getFullYear()}-${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')} ${date.getHours().toString().padStart(2, '0')}:00`
      case TimePeriod.DAY:
        return `${date.getFullYear()}-${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')}`
      case TimePeriod.WEEK:
        const weekStart = new Date(date)
        weekStart.setDate(date.getDate() - date.getDay())
        return `${weekStart.getFullYear()}-W${Math.ceil((weekStart.getDate()) / 7).toString().padStart(2, '0')}`
      case TimePeriod.MONTH:
        return `${date.getFullYear()}-${(date.getMonth() + 1).toString().padStart(2, '0')}`
      case TimePeriod.QUARTER:
        const quarter = Math.ceil((date.getMonth() + 1) / 3)
        return `${date.getFullYear()}-Q${quarter}`
      case TimePeriod.YEAR:
        return date.getFullYear().toString()
      default:
        return date.toISOString().split('T')[0]
    }
  }

  private sumField<T>(items: T[], field: string): number {
    return items.reduce((sum, item) => {
      const value = this.getNestedValue(item, field)
      return sum + (Number(value) || 0)
    }, 0)
  }

  private calculatePercentile(sortedValues: number[], percentile: number): number {
    if (sortedValues.length === 0) return 0
    
    const index = (percentile / 100) * (sortedValues.length - 1)
    const lower = Math.floor(index)
    const upper = Math.ceil(index)
    
    if (lower === upper) {
      return sortedValues[lower]
    }
    
    return sortedValues[lower] + (sortedValues[upper] - sortedValues[lower]) * (index - lower)
  }
}

export const dataAggregator = DataAggregator.getInstance()

// Convenience functions
export const aggregateData = (data: any[], config: AggregationConfig) => 
  dataAggregator.aggregate(data, config)

export const aggregateTimeSeries = (
  data: any[],
  dateField: string,
  valueField: string,
  period: TimePeriod,
  dateRange?: DateRange
) => dataAggregator.aggregateTimeSeries(data, dateField, valueField, period, dateRange)

export const analyzeCohorts = (
  data: any[],
  cohortDateField: string,
  returnDateField: string,
  userIdField: string,
  periods?: number[]
) => dataAggregator.analyzeCohorts(data, cohortDateField, returnDateField, userIdField, periods)

export const analyzeFunnel = (
  data: any[],
  steps: Array<{ name: string; condition: (item: any) => boolean }>,
  userIdField: string
) => dataAggregator.analyzeFunnel(data, steps, userIdField)

export const calculateStatistics = (values: number[]) => 
  dataAggregator.calculateStatistics(values)

export const detectAnomalies = (
  values: number[],
  method: 'zscore' | 'iqr' | 'isolation' = 'zscore',
  threshold: number = 2
) => dataAggregator.detectAnomalies(values, method, threshold)