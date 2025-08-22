import React from 'react'
import { Card } from '@/design-system/components/Card/Card'
import { MetricsGrid } from '@/components/monitoring/MetricsGrid'
import { AlertsPanel } from '@/components/monitoring/AlertsPanel'
import { LogViewer } from '@/components/monitoring/LogViewer'
import { TimeSeriesChart } from '@/components/monitoring/TimeSeriesChart'
import { useMonitoring } from '@/hooks/useMonitoring'

interface MonitoringPageProps {
  className?: string
}

const MonitoringPage: React.FC<MonitoringPageProps> = ({ className }) => {
  const { metrics, alerts, logs, isLoading } = useMonitoring()

  if (isLoading) {
    return (
      <div className="monitoring-page-loading">
        <div className="spinner" />
        <p>Loading monitoring data...</p>
      </div>
    )
  }

  return (
    <div className={`monitoring-page ${className || ''}`}>
      <div className="monitoring-header">
        <h1>System Monitoring</h1>
        <p>Real-time monitoring and observability dashboard</p>
      </div>

      <div className="monitoring-grid">
        {/* Metrics Overview */}
        <Card className="metrics-section">
          <h2>System Metrics</h2>
          <MetricsGrid metrics={metrics} />
        </Card>

        {/* Performance Charts */}
        <Card className="charts-section">
          <h2>Performance Trends</h2>
          <TimeSeriesChart data={metrics.timeSeries} />
        </Card>

        {/* Active Alerts */}
        <Card className="alerts-section">
          <h2>Active Alerts</h2>
          <AlertsPanel alerts={alerts} />
        </Card>

        {/* System Logs */}
        <Card className="logs-section">
          <h2>System Logs</h2>
          <LogViewer logs={logs} />
        </Card>
      </div>
    </div>
  )
}

export default MonitoringPage