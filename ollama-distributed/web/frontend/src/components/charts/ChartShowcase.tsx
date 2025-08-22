/**
 * Chart Showcase Component
 * Comprehensive demonstration of all chart components with real-time data streaming,
 * interactive features, and various configuration options
 */

import React, { useState, useEffect, useRef } from 'react'
import {
  ChartContainer,
  LineChart,
  AreaChart,
  BarChart,
  PieChart,
  Heatmap,
  GaugeChart,
  ChartHelpers,
  configUtils,
  type ChartContainerRef
} from './index'
import { colorUtils } from '@/utils/chartUtils'
import { Button } from '@/design-system/components/Button/Button'
import { Select } from '@/design-system/components/Select/Select'

export interface ChartShowcaseProps {
  /** Theme mode */
  theme?: 'light' | 'dark'
  
  /** Enable real-time data updates */
  realTime?: boolean
  
  /** Update interval in milliseconds */
  updateInterval?: number
}

const ChartShowcase: React.FC<ChartShowcaseProps> = ({
  theme = 'light',
  realTime = false,
  updateInterval = 2000
}) => {
  // Chart theme
  const chartTheme = colorUtils.generateChartTheme(theme)
  
  // Chart container refs for export functionality
  const lineChartRef = useRef<ChartContainerRef>(null)
  const areaChartRef = useRef<ChartContainerRef>(null)
  const barChartRef = useRef<ChartContainerRef>(null)
  const pieChartRef = useRef<ChartContainerRef>(null)
  const heatmapRef = useRef<ChartContainerRef>(null)
  const gaugeRef = useRef<ChartContainerRef>(null)

  // State for demo data
  const [timeSeriesData, setTimeSeriesData] = useState(() =>
    ChartHelpers.generateMockTimeSeriesData(['cpu', 'memory', 'network'], 24, 'hour')
  )
  
  const [resourceData, setResourceData] = useState(() =>
    ChartHelpers.generateMockTimeSeriesData(['cpu', 'memory', 'disk', 'network'], 24, 'hour')
  )
  
  const [categoryData, setCategoryData] = useState(() =>
    ChartHelpers.generateMockCategoryData(['Model A', 'Model B', 'Model C', 'Model D', 'Model E'])
  )
  
  const [distributionData, setDistributionData] = useState(() =>
    ChartHelpers.generateMockCategoryData(['Active', 'Idle', 'Processing', 'Error'])
  )
  
  const [heatmapData, setHeatmapData] = useState(() =>
    ChartHelpers.generateMockHeatmapData(
      ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
      ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00']
    )
  )
  
  const [gaugeValue, setGaugeValue] = useState(75)
  const [loading, setLoading] = useState(false)
  
  // Real-time data updates
  useEffect(() => {
    if (!realTime) return

    const interval = setInterval(() => {
      // Update time series data
      setTimeSeriesData(prev => {
        const newData = [...prev.slice(1)]
        const lastPoint = prev[prev.length - 1]
        const newTimestamp = new Date(new Date(lastPoint.timestamp).getTime() + 5 * 60 * 1000)
        
        newData.push({
          timestamp: newTimestamp.toISOString(),
          cpu: Math.max(0, Math.min(100, lastPoint.cpu + (Math.random() - 0.5) * 10)),
          memory: Math.max(0, Math.min(100, lastPoint.memory + (Math.random() - 0.5) * 8)),
          network: Math.max(0, Math.min(100, lastPoint.network + (Math.random() - 0.5) * 15))
        })
        
        return newData
      })
      
      // Update gauge value
      setGaugeValue(prev => Math.max(0, Math.min(100, prev + (Math.random() - 0.5) * 20)))
    }, updateInterval)

    return () => clearInterval(interval)
  }, [realTime, updateInterval])

  // Refresh all data
  const refreshData = () => {
    setLoading(true)
    
    setTimeout(() => {
      setTimeSeriesData(ChartHelpers.generateMockTimeSeriesData(['cpu', 'memory', 'network'], 24, 'hour'))
      setResourceData(ChartHelpers.generateMockTimeSeriesData(['cpu', 'memory', 'disk', 'network'], 24, 'hour'))
      setCategoryData(ChartHelpers.generateMockCategoryData(['Model A', 'Model B', 'Model C', 'Model D', 'Model E']))
      setDistributionData(ChartHelpers.generateMockCategoryData(['Active', 'Idle', 'Processing', 'Error']))
      setHeatmapData(ChartHelpers.generateMockHeatmapData(
        ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
        ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00']
      ))
      setGaugeValue(Math.random() * 100)
      setLoading(false)
    }, 500)
  }

  return (
    <div className="space-y-8 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Chart Showcase</h1>
          <p className="text-muted-foreground mt-2">
            Comprehensive data visualization components for distributed Ollama monitoring
          </p>
        </div>
        
        <div className="flex items-center gap-4">
          <Button
            variant="outline"
            onClick={refreshData}
            disabled={loading}
          >
            {loading ? 'Refreshing...' : 'Refresh Data'}
          </Button>
          
          <Button
            variant="outline"
            onClick={() => {
              // Export all charts as PNG
              const refs = [lineChartRef, areaChartRef, barChartRef, pieChartRef, heatmapRef, gaugeRef]
              refs.forEach((ref, index) => {
                if (ref.current) {
                  setTimeout(() => ref.current?.exportAsPNG(), index * 200)
                }
              })
            }}
          >
            Export All
          </Button>
        </div>
      </div>

      {/* Line Chart - System Metrics */}
      <ChartContainer
        ref={lineChartRef}
        title="System Performance Metrics"
        description="Real-time CPU, Memory, and Network utilization over time"
        loading={loading}
        data={timeSeriesData}
        theme={theme}
        autoRefresh={realTime ? updateInterval : undefined}
        onRefresh={refreshData}
        exportFilename="system-metrics"
        height={400}
      >
        <LineChart
          data={timeSeriesData}
          metrics={[
            { 
              key: 'cpu', 
              name: 'CPU Usage', 
              color: chartTheme.colors.semantic.error,
              format: 'percentage'
            },
            { 
              key: 'memory', 
              name: 'Memory Usage', 
              color: chartTheme.colors.semantic.warning,
              format: 'percentage'
            },
            { 
              key: 'network', 
              name: 'Network I/O', 
              color: chartTheme.colors.semantic.info,
              format: 'bytes',
              unit: 'MB/s'
            }
          ]}
          interactive={{
            zoom: true,
            brush: true,
            crosshair: true,
            clickable: true
          }}
          referenceLines={[
            { value: 80, label: 'Warning Threshold', color: chartTheme.colors.semantic.warning },
            { value: 95, label: 'Critical Threshold', color: chartTheme.colors.semantic.error }
          ]}
        />
      </ChartContainer>

      {/* Area Chart - Resource Usage */}
      <ChartContainer
        ref={areaChartRef}
        title="Resource Utilization Distribution"
        description="Stacked view of system resource usage across different components"
        loading={loading}
        data={resourceData}
        theme={theme}
        exportFilename="resource-utilization"
        height={350}
      >
        <AreaChart
          data={resourceData}
          type="stacked"
          areas={[
            { key: 'cpu', name: 'CPU', color: chartTheme.colors.primary[0], format: 'percentage' },
            { key: 'memory', name: 'Memory', color: chartTheme.colors.primary[1], format: 'percentage' },
            { key: 'disk', name: 'Disk I/O', color: chartTheme.colors.primary[2], format: 'percentage' },
            { key: 'network', name: 'Network', color: chartTheme.colors.primary[3], format: 'percentage' }
          ]}
          yAxis={{
            domain: [0, 400], // Stacked percentages
            format: 'percentage'
          }}
          interactive={{
            crosshair: true,
            clickable: true
          }}
        />
      </ChartContainer>

      {/* Bar Chart - Model Performance */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChartContainer
          ref={barChartRef}
          title="Model Performance Comparison"
          description="Response time comparison across different AI models"
          loading={loading}
          data={categoryData}
          theme={theme}
          exportFilename="model-performance"
          height={300}
        >
          <BarChart
            data={categoryData}
            bars={[
              { 
                key: 'value', 
                name: 'Response Time',
                format: 'duration',
                unit: 'ms'
              }
            ]}
            colorMapping={{
              'Model A': chartTheme.colors.semantic.success,
              'Model B': chartTheme.colors.semantic.info,
              'Model C': chartTheme.colors.semantic.warning,
              'Model D': chartTheme.colors.semantic.error,
              'Model E': chartTheme.colors.primary[0]
            }}
            interactive={{
              clickable: true,
              highlightOnHover: true
            }}
            yAxis={{
              format: 'duration'
            }}
          />
        </ChartContainer>

        {/* Pie Chart - System Status */}
        <ChartContainer
          ref={pieChartRef}
          title="System Status Distribution"
          description="Current distribution of system states across the cluster"
          loading={loading}
          data={distributionData}
          theme={theme}
          exportFilename="system-status"
          height={300}
        >
          <PieChart
            data={distributionData}
            variant="donut"
            colors={[
              chartTheme.colors.semantic.success,
              chartTheme.colors.semantic.info,
              chartTheme.colors.semantic.warning,
              chartTheme.colors.semantic.error
            ]}
            labels={{
              show: true,
              position: 'outside',
              type: 'keyPercent'
            }}
            interactive={{
              clickable: true,
              highlightOnHover: true,
              expandOnHover: true
            }}
            donut={{
              innerRadius: '50%',
              centerContent: (
                <div className="text-center">
                  <div className="text-2xl font-bold">
                    {distributionData.reduce((sum, item) => sum + item.value, 0).toFixed(0)}
                  </div>
                  <div className="text-sm text-muted-foreground">Total Nodes</div>
                </div>
              )
            }}
          />
        </ChartContainer>
      </div>

      {/* Heatmap - Usage Patterns */}
      <ChartContainer
        ref={heatmapRef}
        title="Weekly Usage Patterns"
        description="System utilization heatmap showing usage patterns throughout the week"
        loading={loading}
        data={heatmapData}
        theme={theme}
        exportFilename="usage-patterns"
        height={300}
      >
        <Heatmap
          data={heatmapData}
          xLabels={['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']}
          yLabels={['00:00', '04:00', '08:00', '12:00', '16:00', '20:00']}
          valueFormat={{
            format: 'percentage',
            precision: 1
          }}
          colorScale={{
            type: 'linear',
            colors: [
              chartTheme.colors.primary[4],
              chartTheme.colors.primary[3],
              chartTheme.colors.primary[2],
              chartTheme.colors.primary[1],
              chartTheme.colors.primary[0]
            ]
          }}
          cell={{
            size: 'auto',
            gap: 2,
            radius: 4,
            border: true
          }}
          interactive={{
            clickable: true,
            highlightOnHover: true,
            showTooltip: true
          }}
          legend={{
            show: true,
            position: 'right'
          }}
        />
      </ChartContainer>

      {/* Gauge Charts - Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <ChartContainer
          ref={gaugeRef}
          title="CPU Utilization"
          loading={loading}
          theme={theme}
          exportFilename="cpu-gauge"
          height={250}
        >
          <GaugeChart
            value={gaugeValue}
            min={0}
            max={100}
            style="half"
            valueFormat={{
              format: 'percentage',
              precision: 1
            }}
            thresholds={[
              { value: 50, color: chartTheme.colors.semantic.success, label: 'Normal' },
              { value: 80, color: chartTheme.colors.semantic.warning, label: 'Warning' },
              { value: 95, color: chartTheme.colors.semantic.error, label: 'Critical' }
            ]}
            appearance={{
              showValue: true,
              showThresholds: true,
              rounded: true
            }}
            labels={{
              title: 'CPU',
              unit: '%'
            }}
            target={{
              value: 70,
              color: chartTheme.colors.semantic.info,
              label: 'Target'
            }}
            gradient={{
              enabled: true,
              colors: [
                chartTheme.colors.semantic.success,
                chartTheme.colors.semantic.warning,
                chartTheme.colors.semantic.error
              ]
            }}
          />
        </ChartContainer>

        <ChartContainer
          title="Memory Usage"
          loading={loading}
          theme={theme}
          height={250}
        >
          <GaugeChart
            value={85}
            min={0}
            max={100}
            style="half"
            valueFormat={{ format: 'percentage' }}
            labels={{ title: 'Memory', unit: '%' }}
            appearance={{ showValue: true, rounded: true }}
          />
        </ChartContainer>

        <ChartContainer
          title="Network Throughput"
          loading={loading}
          theme={theme}
          height={250}
        >
          <GaugeChart
            value={65}
            min={0}
            max={1000}
            style="half"
            valueFormat={{ format: 'bytes', unit: 'MB/s' }}
            labels={{ title: 'Network', unit: 'MB/s' }}
            appearance={{ showValue: true, rounded: true }}
          />
        </ChartContainer>

        <ChartContainer
          title="System Health"
          loading={loading}
          theme={theme}
          height={250}
        >
          <GaugeChart
            value={92}
            min={0}
            max={100}
            style="half"
            valueFormat={{ format: 'default', precision: 0 }}
            labels={{ title: 'Health Score' }}
            appearance={{ showValue: true, rounded: true }}
            thresholds={[
              { value: 70, color: chartTheme.colors.semantic.error },
              { value: 85, color: chartTheme.colors.semantic.warning },
              { value: 100, color: chartTheme.colors.semantic.success }
            ]}
          />
        </ChartContainer>
      </div>

      {/* Chart Configuration Info */}
      <div className="mt-12 p-6 border rounded-lg bg-card">
        <h3 className="text-lg font-semibold mb-4">Chart Configuration Features</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 text-sm">
          <div>
            <h4 className="font-medium text-green-600 mb-2">Interactive Features</h4>
            <ul className="space-y-1 text-muted-foreground">
              <li>• Zoom and pan support</li>
              <li>• Hover tooltips</li>
              <li>• Click interactions</li>
              <li>• Brush selection</li>
              <li>• Real-time updates</li>
            </ul>
          </div>
          
          <div>
            <h4 className="font-medium text-blue-600 mb-2">Export Options</h4>
            <ul className="space-y-1 text-muted-foreground">
              <li>• PNG with high DPI</li>
              <li>• SVG vector format</li>
              <li>• PDF documents</li>
              <li>• CSV data export</li>
              <li>• Batch export support</li>
            </ul>
          </div>
          
          <div>
            <h4 className="font-medium text-purple-600 mb-2">Accessibility</h4>
            <ul className="space-y-1 text-muted-foreground">
              <li>• WCAG 2.1 AA compliant</li>
              <li>• Keyboard navigation</li>
              <li>• Screen reader support</li>
              <li>• Motion preference respect</li>
              <li>• High contrast mode</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}

export { ChartShowcase }
export type { ChartShowcaseProps }