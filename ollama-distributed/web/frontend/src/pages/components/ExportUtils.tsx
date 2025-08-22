/**
 * ExportUtils Component - Provides data export capabilities for dashboard
 */

import React, { useState } from 'react'
import { Button } from '@/design-system/components/Button/Button'
import { Badge } from '@/design-system/components/Badge/Badge'
import { 
  Download,
  FileText,
  FileSpreadsheet,
  Image,
  Loader2
} from 'lucide-react'
import jsPDF from 'jspdf'
import 'jspdf-autotable'
import * as XLSX from 'xlsx'

interface ExportData {
  metrics: any
  activities: any[]
  alerts: any[]
  performanceData: any[]
  timestamp: string
}

interface ExportUtilsProps {
  data: ExportData
  filename?: string
  className?: string
}

const ExportUtils: React.FC<ExportUtilsProps> = ({
  data,
  filename = 'ollama-dashboard',
  className = ''
}) => {
  const [isExporting, setIsExporting] = useState<string | null>(null)

  const exportToPDF = async () => {
    setIsExporting('pdf')
    try {
      const pdf = new jsPDF()
      const pageWidth = pdf.internal.pageSize.getWidth()
      
      // Header
      pdf.setFontSize(20)
      pdf.setTextColor(40, 40, 40)
      pdf.text('OllamaMax Dashboard Report', 20, 30)
      
      // Timestamp
      pdf.setFontSize(10)
      pdf.setTextColor(100, 100, 100)
      pdf.text(`Generated: ${new Date(data.timestamp).toLocaleString()}`, 20, 40)
      
      let yPosition = 60
      
      // Metrics Summary
      pdf.setFontSize(16)
      pdf.setTextColor(40, 40, 40)
      pdf.text('System Metrics', 20, yPosition)
      yPosition += 10
      
      const metricsData = [
        ['Metric', 'Value', 'Status'],
        ['Active Nodes', `${data.metrics.nodes.healthy}/${data.metrics.nodes.total}`, 'Healthy'],
        ['Models Synced', `${data.metrics.models.synced}/${data.metrics.models.total}`, 'Good'],
        ['Active Tasks', `${data.metrics.tasks.running}`, 'Normal'],
        ['CPU Usage', `${data.metrics.performance.cpu}%`, data.metrics.performance.cpu > 80 ? 'High' : 'Normal'],
        ['Memory Usage', `${data.metrics.performance.memory}%`, data.metrics.performance.memory > 80 ? 'High' : 'Normal'],
      ]
      
      // @ts-ignore - jsPDF autoTable plugin
      pdf.autoTable({
        head: [metricsData[0]],
        body: metricsData.slice(1),
        startY: yPosition,
        styles: { fontSize: 8 },
        headStyles: { fillColor: [59, 130, 246] }
      })
      
      // @ts-ignore
      yPosition = pdf.lastAutoTable.finalY + 20
      
      // Recent Activities
      if (yPosition > 200) {
        pdf.addPage()
        yPosition = 20
      }
      
      pdf.setFontSize(16)
      pdf.text('Recent Activities', 20, yPosition)
      yPosition += 10
      
      const activitiesData = [
        ['Time', 'Event', 'Description', 'Severity'],
        ...data.activities.slice(0, 10).map(activity => [
          new Date(activity.timestamp).toLocaleTimeString(),
          activity.title,
          activity.description.substring(0, 50) + '...',
          activity.severity
        ])
      ]
      
      // @ts-ignore
      pdf.autoTable({
        head: [activitiesData[0]],
        body: activitiesData.slice(1),
        startY: yPosition,
        styles: { fontSize: 7 },
        headStyles: { fillColor: [59, 130, 246] },
        columnStyles: {
          2: { cellWidth: 60 }
        }
      })
      
      // Alerts Summary
      // @ts-ignore
      yPosition = pdf.lastAutoTable.finalY + 20
      
      if (yPosition > 220) {
        pdf.addPage()
        yPosition = 20
      }
      
      pdf.setFontSize(16)
      pdf.text('System Alerts', 20, yPosition)
      yPosition += 10
      
      const alertsData = [
        ['Type', 'Title', 'Source', 'Status'],
        ...data.alerts.slice(0, 10).map(alert => [
          alert.type,
          alert.title,
          alert.source,
          alert.acknowledged ? 'Acknowledged' : 'Active'
        ])
      ]
      
      // @ts-ignore
      pdf.autoTable({
        head: [alertsData[0]],
        body: alertsData.slice(1),
        startY: yPosition,
        styles: { fontSize: 8 },
        headStyles: { fillColor: [59, 130, 246] }
      })
      
      // Save the PDF
      pdf.save(`${filename}-${new Date().toISOString().split('T')[0]}.pdf`)
      
    } catch (error) {
      console.error('Failed to export PDF:', error)
    } finally {
      setIsExporting(null)
    }
  }

  const exportToExcel = async () => {
    setIsExporting('excel')
    try {
      const workbook = XLSX.utils.book_new()
      
      // Metrics worksheet
      const metricsWS = XLSX.utils.json_to_sheet([
        {
          'Metric': 'Active Nodes',
          'Value': `${data.metrics.nodes.healthy}/${data.metrics.nodes.total}`,
          'Details': `Healthy: ${data.metrics.nodes.healthy}, Degraded: ${data.metrics.nodes.degraded}, Offline: ${data.metrics.nodes.offline}`
        },
        {
          'Metric': 'Models',
          'Value': `${data.metrics.models.synced}/${data.metrics.models.total}`,
          'Details': `Synced: ${data.metrics.models.synced}, Syncing: ${data.metrics.models.syncing}, Failed: ${data.metrics.models.failed}`
        },
        {
          'Metric': 'Tasks',
          'Value': `${data.metrics.tasks.running} active`,
          'Details': `Running: ${data.metrics.tasks.running}, Pending: ${data.metrics.tasks.pending}, Completed: ${data.metrics.tasks.completed}, Failed: ${data.metrics.tasks.failed}`
        },
        {
          'Metric': 'CPU Usage',
          'Value': `${data.metrics.performance.cpu}%`,
          'Details': 'Current CPU utilization across cluster'
        },
        {
          'Metric': 'Memory Usage',
          'Value': `${data.metrics.performance.memory}%`,
          'Details': 'Current memory utilization across cluster'
        },
        {
          'Metric': 'Network Usage',
          'Value': `${data.metrics.performance.network}%`,
          'Details': 'Current network utilization across cluster'
        }
      ])
      XLSX.utils.book_append_sheet(workbook, metricsWS, 'Metrics')
      
      // Activities worksheet
      const activitiesData = data.activities.map(activity => ({
        'Timestamp': new Date(activity.timestamp).toLocaleString(),
        'Type': activity.type,
        'Title': activity.title,
        'Description': activity.description,
        'Severity': activity.severity
      }))
      const activitiesWS = XLSX.utils.json_to_sheet(activitiesData)
      XLSX.utils.book_append_sheet(workbook, activitiesWS, 'Activities')
      
      // Alerts worksheet
      const alertsData = data.alerts.map(alert => ({
        'Timestamp': new Date(alert.timestamp).toLocaleString(),
        'Type': alert.type,
        'Title': alert.title,
        'Message': alert.message,
        'Source': alert.source,
        'Status': alert.acknowledged ? 'Acknowledged' : 'Active'
      }))
      const alertsWS = XLSX.utils.json_to_sheet(alertsData)
      XLSX.utils.book_append_sheet(workbook, alertsWS, 'Alerts')
      
      // Performance Data worksheet
      const performanceWS = XLSX.utils.json_to_sheet(data.performanceData)
      XLSX.utils.book_append_sheet(workbook, performanceWS, 'Performance')
      
      // Save the Excel file
      XLSX.writeFile(workbook, `${filename}-${new Date().toISOString().split('T')[0]}.xlsx`)
      
    } catch (error) {
      console.error('Failed to export Excel:', error)
    } finally {
      setIsExporting(null)
    }
  }

  const exportToJSON = () => {
    setIsExporting('json')
    try {
      const exportData = {
        ...data,
        exportedAt: new Date().toISOString(),
        version: '1.0.0'
      }
      
      const dataStr = JSON.stringify(exportData, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `${filename}-${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
      
    } catch (error) {
      console.error('Failed to export JSON:', error)
    } finally {
      setIsExporting(null)
    }
  }

  const exportToCSV = () => {
    setIsExporting('csv')
    try {
      // Create CSV for metrics
      const metricsCSV = [
        ['Metric', 'Value', 'Details'],
        ['Active Nodes', `${data.metrics.nodes.healthy}/${data.metrics.nodes.total}`, `Healthy: ${data.metrics.nodes.healthy}, Degraded: ${data.metrics.nodes.degraded}`],
        ['Models', `${data.metrics.models.synced}/${data.metrics.models.total}`, `Synced: ${data.metrics.models.synced}, Syncing: ${data.metrics.models.syncing}`],
        ['Active Tasks', data.metrics.tasks.running, `Pending: ${data.metrics.tasks.pending}, Completed: ${data.metrics.tasks.completed}`],
        ['CPU Usage', `${data.metrics.performance.cpu}%`, 'Current cluster CPU utilization'],
        ['Memory Usage', `${data.metrics.performance.memory}%`, 'Current cluster memory utilization']
      ].map(row => row.join(',')).join('\\n')
      
      const dataBlob = new Blob([metricsCSV], { type: 'text/csv' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `${filename}-metrics-${new Date().toISOString().split('T')[0]}.csv`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
      
    } catch (error) {
      console.error('Failed to export CSV:', error)
    } finally {
      setIsExporting(null)
    }
  }

  const exportFormats = [
    {
      type: 'pdf',
      label: 'PDF Report',
      description: 'Complete dashboard report',
      icon: FileText,
      action: exportToPDF
    },
    {
      type: 'excel',
      label: 'Excel',
      description: 'Spreadsheet with all data',
      icon: FileSpreadsheet,
      action: exportToExcel
    },
    {
      type: 'json',
      label: 'JSON',
      description: 'Raw data format',
      icon: Download,
      action: exportToJSON
    },
    {
      type: 'csv',
      label: 'CSV',
      description: 'Metrics summary',
      icon: FileSpreadsheet,
      action: exportToCSV
    }
  ]

  return (
    <div className={`flex flex-wrap gap-2 ${className}`}>
      {exportFormats.map((format) => (
        <Button
          key={format.type}
          variant="outline"
          size="sm"
          disabled={isExporting === format.type}
          onClick={format.action}
          className="flex items-center gap-2"
        >
          {isExporting === format.type ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <format.icon className="h-4 w-4" />
          )}
          <span>{format.label}</span>
        </Button>
      ))}
    </div>
  )
}

export default ExportUtils