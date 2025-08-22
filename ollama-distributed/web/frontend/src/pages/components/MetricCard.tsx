/**
 * MetricCard Component - Displays key metrics with trends and status indicators
 */

import React from 'react'
import { Card, CardHeader, CardContent, CardTitle } from '@/design-system/components/Card/Card'
import { Badge } from '@/design-system/components/Badge/Badge'
import { ArrowUp, ArrowDown, TrendingUp, TrendingDown } from 'lucide-react'

interface MetricCardProps {
  title: string
  value: string | number
  total?: number
  icon: React.ReactNode
  trend?: React.ReactNode
  status: 'healthy' | 'warning' | 'error' | 'info'
  subtitle?: string
  change?: {
    value: number
    type: 'increase' | 'decrease'
    period: string
  }
  loading?: boolean
}

const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  total,
  icon,
  trend,
  status,
  subtitle,
  change,
  loading = false
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'text-success-600 bg-success-50 border-success-200'
      case 'warning':
        return 'text-warning-600 bg-warning-50 border-warning-200'
      case 'error':
        return 'text-error-600 bg-error-50 border-error-200'
      default:
        return 'text-info-600 bg-info-50 border-info-200'
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'secondary'
      case 'warning':
        return 'warning'
      case 'error':
        return 'destructive'
      default:
        return 'secondary'
    }
  }

  const formatValue = (val: string | number) => {
    if (typeof val === 'number') {
      if (val >= 1000000) {
        return `${(val / 1000000).toFixed(1)}M`
      } else if (val >= 1000) {
        return `${(val / 1000).toFixed(1)}K`
      }
      return val.toString()
    }
    return val
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="animate-pulse">
            <div className="flex items-center justify-between mb-4">
              <div className="h-4 bg-muted rounded w-20"></div>
              <div className="h-6 w-6 bg-muted rounded"></div>
            </div>
            <div className="h-8 bg-muted rounded w-16 mb-2"></div>
            <div className="h-3 bg-muted rounded w-24"></div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="transition-all duration-200 hover:shadow-lg">
      <CardContent className="p-6">
        <div className="flex items-center justify-between mb-4">
          <CardTitle className="text-sm font-medium text-muted-foreground">
            {title}
          </CardTitle>
          <div className={`p-2 rounded-lg ${getStatusColor(status)}`}>
            {icon}
          </div>
        </div>
        
        <div className="space-y-2">
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-foreground">
              {formatValue(value)}
            </span>
            {total && (
              <span className="text-sm text-muted-foreground">
                / {formatValue(total)}
              </span>
            )}
            {trend && (
              <div className="ml-auto">
                {trend}
              </div>
            )}
          </div>
          
          {subtitle && (
            <p className="text-xs text-muted-foreground">
              {subtitle}
            </p>
          )}
          
          {change && (
            <div className="flex items-center gap-2">
              <div className={`flex items-center gap-1 text-xs ${
                change.type === 'increase' ? 'text-success-600' : 'text-error-600'
              }`}>
                {change.type === 'increase' ? (
                  <TrendingUp className="h-3 w-3" />
                ) : (
                  <TrendingDown className="h-3 w-3" />
                )}
                <span>{Math.abs(change.value)}%</span>
              </div>
              <span className="text-xs text-muted-foreground">
                vs {change.period}
              </span>
            </div>
          )}
          
          <div className="flex justify-end">
            <Badge variant={getStatusBadge(status)} className="text-xs">
              {status}
            </Badge>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export default MetricCard