/**
 * Export Service
 * Handles data export in multiple formats (PDF, CSV, Excel) with advanced formatting
 */

import { 
  Report,
  ReportFormat,
  ExportOptions,
  ExportResult,
  BusinessMetrics,
  EnhancedPerformanceMetrics,
  ComplianceReport
} from '../../analytics/types'

// PDF generation library
import jsPDF from 'jspdf'
import 'jspdf-autotable'
import * as XLSX from 'xlsx'

interface ExportData {
  title: string
  data: any[]
  metadata?: Record<string, any>
  charts?: ChartData[]
  summary?: Record<string, any>
}

interface ChartData {
  title: string
  type: 'line' | 'bar' | 'pie' | 'table'
  data: any[]
  config?: any
}

export class ExportService {
  private static instance: ExportService

  static getInstance(): ExportService {
    if (!ExportService.instance) {
      ExportService.instance = new ExportService()
    }
    return ExportService.instance
  }

  async exportReport(
    data: ExportData,
    format: ReportFormat,
    options: ExportOptions = {
      format: 'pdf',
      includeCharts: true,
      includeData: true,
      compressed: false,
      branding: true
    }
  ): Promise<ExportResult> {
    const filename = this.generateFilename(data.title, format)
    
    let blob: Blob
    let size: number

    switch (format) {
      case 'pdf':
        blob = await this.generatePDF(data, options)
        break
      case 'csv':
        blob = this.generateCSV(data)
        break
      case 'excel':
        blob = await this.generateExcel(data, options)
        break
      case 'json':
        blob = this.generateJSON(data)
        break
      default:
        throw new Error(`Unsupported export format: ${format}`)
    }

    size = blob.size

    // Create download URL
    const downloadUrl = URL.createObjectURL(blob)
    
    // Auto-download
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    const result: ExportResult = {
      id: `export_${Date.now()}`,
      filename,
      size,
      downloadUrl,
      expiresAt: Date.now() + (24 * 60 * 60 * 1000), // 24 hours
      password: !!options.password
    }

    // Clean up URL after some time
    setTimeout(() => {
      URL.revokeObjectURL(downloadUrl)
    }, 10000)

    return result
  }

  private async generatePDF(data: ExportData, options: ExportOptions): Promise<Blob> {
    const pdf = new jsPDF()
    const pageWidth = pdf.internal.pageSize.width
    const pageHeight = pdf.internal.pageSize.height
    let yPosition = 20

    // Header with branding
    if (options.branding) {
      pdf.setFontSize(20)
      pdf.setTextColor(44, 62, 80)
      pdf.text('Ollama Distributed', 20, yPosition)
      yPosition += 10
      
      pdf.setFontSize(16)
      pdf.setTextColor(100, 100, 100)
      pdf.text(data.title, 20, yPosition)
      yPosition += 15
    } else {
      pdf.setFontSize(18)
      pdf.text(data.title, 20, yPosition)
      yPosition += 15
    }

    // Metadata
    if (data.metadata) {
      pdf.setFontSize(10)
      pdf.setTextColor(100, 100, 100)
      pdf.text(`Generated: ${new Date().toLocaleString()}`, 20, yPosition)
      yPosition += 8
      
      Object.entries(data.metadata).forEach(([key, value]) => {
        pdf.text(`${key}: ${value}`, 20, yPosition)
        yPosition += 6
      })
      yPosition += 10
    }

    // Summary section
    if (data.summary && Object.keys(data.summary).length > 0) {
      pdf.setFontSize(14)
      pdf.setTextColor(44, 62, 80)
      pdf.text('Executive Summary', 20, yPosition)
      yPosition += 10

      const summaryData = Object.entries(data.summary).map(([key, value]) => [
        this.formatLabel(key),
        this.formatValue(value)
      ])

      ;(pdf as any).autoTable({
        head: [['Metric', 'Value']],
        body: summaryData,
        startY: yPosition,
        theme: 'grid',
        headStyles: { fillColor: [52, 152, 219] },
        margin: { left: 20 }
      })

      yPosition = (pdf as any).lastAutoTable.finalY + 20
    }

    // Data tables
    if (options.includeData && data.data.length > 0) {
      if (yPosition > pageHeight - 100) {
        pdf.addPage()
        yPosition = 20
      }

      pdf.setFontSize(14)
      pdf.setTextColor(44, 62, 80)
      pdf.text('Detailed Data', 20, yPosition)
      yPosition += 10

      // Convert data to table format
      const headers = Object.keys(data.data[0])
      const rows = data.data.map(row => 
        headers.map(header => this.formatValue(row[header]))
      )

      ;(pdf as any).autoTable({
        head: [headers.map(h => this.formatLabel(h))],
        body: rows,
        startY: yPosition,
        theme: 'striped',
        headStyles: { fillColor: [52, 152, 219] },
        margin: { left: 20 },
        columnStyles: this.generateColumnStyles(headers),
        didDrawPage: (data: any) => {
          // Add page numbers
          pdf.setFontSize(8)
          pdf.setTextColor(100, 100, 100)
          pdf.text(
            `Page ${data.pageNumber}`,
            pageWidth - 30,
            pageHeight - 10
          )
        }
      })
    }

    // Charts section (placeholder - would need chart rendering library)
    if (options.includeCharts && data.charts && data.charts.length > 0) {
      pdf.addPage()
      yPosition = 20
      
      pdf.setFontSize(14)
      pdf.setTextColor(44, 62, 80)
      pdf.text('Charts and Visualizations', 20, yPosition)
      yPosition += 10
      
      data.charts.forEach((chart, index) => {
        pdf.setFontSize(12)
        pdf.text(chart.title, 20, yPosition)
        yPosition += 8
        
        // Placeholder for chart - in real implementation, you'd render the chart
        pdf.setDrawColor(200, 200, 200)
        pdf.rect(20, yPosition, 170, 80)
        pdf.setFontSize(10)
        pdf.setTextColor(150, 150, 150)
        pdf.text(`${chart.type} chart would be rendered here`, 105, yPosition + 40, { align: 'center' })
        yPosition += 90
        
        if (yPosition > pageHeight - 100 && index < data.charts.length - 1) {
          pdf.addPage()
          yPosition = 20
        }
      })
    }

    // Watermark
    if (options.watermark) {
      const totalPages = pdf.getNumberOfPages()
      for (let i = 1; i <= totalPages; i++) {
        pdf.setPage(i)
        pdf.saveGraphicsState()
        pdf.setGState(pdf.GState({ opacity: 0.1 }))
        pdf.setFontSize(50)
        pdf.setTextColor(200, 200, 200)
        pdf.text(options.watermark, pageWidth / 2, pageHeight / 2, {
          align: 'center',
          angle: 45
        })
        pdf.restoreGraphicsState()
      }
    }

    return new Blob([pdf.output('blob')], { type: 'application/pdf' })
  }

  private generateCSV(data: ExportData): Blob {
    if (!data.data.length) {
      return new Blob(['No data available'], { type: 'text/csv' })
    }

    const headers = Object.keys(data.data[0])
    const csvContent = [
      // Header row
      headers.map(h => this.escapeCSV(this.formatLabel(h))).join(','),
      // Data rows
      ...data.data.map(row =>
        headers.map(header => this.escapeCSV(this.formatValue(row[header]))).join(',')
      )
    ].join('\n')

    return new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  }

  private async generateExcel(data: ExportData, options: ExportOptions): Promise<Blob> {
    const workbook = XLSX.utils.book_new()

    // Summary sheet
    if (data.summary) {
      const summaryData = [
        ['Metric', 'Value'],
        ...Object.entries(data.summary).map(([key, value]) => [
          this.formatLabel(key),
          this.formatValue(value)
        ])
      ]
      
      const summarySheet = XLSX.utils.aoa_to_sheet(summaryData)
      
      // Style the header
      const headerStyle = {
        font: { bold: true },
        fill: { fgColor: { rgb: 'E3F2FD' } },
        border: {
          top: { style: 'thin' },
          bottom: { style: 'thin' },
          left: { style: 'thin' },
          right: { style: 'thin' }
        }
      }
      
      if (summarySheet['A1']) summarySheet['A1'].s = headerStyle
      if (summarySheet['B1']) summarySheet['B1'].s = headerStyle
      
      XLSX.utils.book_append_sheet(workbook, summarySheet, 'Summary')
    }

    // Data sheet
    if (options.includeData && data.data.length > 0) {
      const headers = Object.keys(data.data[0])
      const sheetData = [
        headers.map(h => this.formatLabel(h)),
        ...data.data.map(row =>
          headers.map(header => this.formatValue(row[header]))
        )
      ]

      const dataSheet = XLSX.utils.aoa_to_sheet(sheetData)
      
      // Auto-size columns
      const columnWidths = headers.map(header => ({
        wch: Math.max(
          header.length,
          ...data.data.map(row => String(this.formatValue(row[header])).length)
        )
      }))
      
      dataSheet['!cols'] = columnWidths
      
      XLSX.utils.book_append_sheet(workbook, dataSheet, 'Data')
    }

    // Charts data (as separate sheets)
    if (options.includeCharts && data.charts) {
      data.charts.forEach((chart, index) => {
        if (chart.data.length > 0) {
          const chartHeaders = Object.keys(chart.data[0])
          const chartSheetData = [
            chartHeaders,
            ...chart.data.map(row =>
              chartHeaders.map(header => row[header])
            )
          ]

          const chartSheet = XLSX.utils.aoa_to_sheet(chartSheetData)
          XLSX.utils.book_append_sheet(workbook, chartSheet, `Chart_${index + 1}`)
        }
      })
    }

    // Metadata sheet
    if (data.metadata) {
      const metadataData = [
        ['Property', 'Value'],
        ['Report Title', data.title],
        ['Generated At', new Date().toISOString()],
        ...Object.entries(data.metadata).map(([key, value]) => [key, String(value)])
      ]

      const metadataSheet = XLSX.utils.aoa_to_sheet(metadataData)
      XLSX.utils.book_append_sheet(workbook, metadataSheet, 'Metadata')
    }

    const excelBuffer = XLSX.write(workbook, { bookType: 'xlsx', type: 'array' })
    return new Blob([excelBuffer], { 
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' 
    })
  }

  private generateJSON(data: ExportData): Blob {
    const jsonData = {
      title: data.title,
      generatedAt: new Date().toISOString(),
      metadata: data.metadata,
      summary: data.summary,
      data: data.data,
      charts: data.charts?.map(chart => ({
        title: chart.title,
        type: chart.type,
        data: chart.data
      }))
    }

    return new Blob([JSON.stringify(jsonData, null, 2)], {
      type: 'application/json'
    })
  }

  // Helper methods
  private generateFilename(title: string, format: ReportFormat): string {
    const sanitizedTitle = title.replace(/[^a-z0-9]/gi, '_').toLowerCase()
    const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-')
    return `${sanitizedTitle}_${timestamp}.${format}`
  }

  private formatLabel(key: string): string {
    return key
      .replace(/([A-Z])/g, ' $1')
      .replace(/^./, str => str.toUpperCase())
      .replace(/_/g, ' ')
      .trim()
  }

  private formatValue(value: any): string {
    if (value === null || value === undefined) return ''
    if (typeof value === 'number') {
      if (Number.isInteger(value)) return value.toString()
      return value.toFixed(2)
    }
    if (typeof value === 'boolean') return value ? 'Yes' : 'No'
    if (value instanceof Date) return value.toLocaleString()
    return String(value)
  }

  private escapeCSV(value: string): string {
    if (value.includes(',') || value.includes('"') || value.includes('\n')) {
      return `"${value.replace(/"/g, '""')}"`
    }
    return value
  }

  private generateColumnStyles(headers: string[]): Record<string, any> {
    const styles: Record<string, any> = {}
    
    headers.forEach((header, index) => {
      if (header.toLowerCase().includes('date') || header.toLowerCase().includes('time')) {
        styles[index] = { cellWidth: 'auto', fontSize: 9 }
      } else if (header.toLowerCase().includes('id')) {
        styles[index] = { cellWidth: 20, fontSize: 8 }
      } else if (header.toLowerCase().includes('amount') || header.toLowerCase().includes('price')) {
        styles[index] = { halign: 'right' }
      }
    })

    return styles
  }

  // Specialized export methods
  async exportBusinessMetrics(
    metrics: BusinessMetrics,
    format: ReportFormat,
    options?: Partial<ExportOptions>
  ): Promise<ExportResult> {
    const data: ExportData = {
      title: 'Business Intelligence Report',
      metadata: {
        'Report Type': 'Business Metrics',
        'Time Period': 'Last 30 Days',
        'Currency': 'USD'
      },
      summary: {
        'Total Revenue': `$${metrics.revenue.total.toLocaleString()}`,
        'Active Users': metrics.users.active.toLocaleString(),
        'Conversion Rate': `${metrics.conversion.overall.rate.toFixed(2)}%`,
        'Customer LTV': `$${metrics.revenue.ltv.toFixed(2)}`,
        'Churn Rate': `${(metrics.revenue.churnRate * 100).toFixed(1)}%`
      },
      data: [
        ...metrics.revenue.byProduct.map(p => ({
          type: 'Product Revenue',
          name: p.name,
          revenue: p.revenue,
          units: p.units,
          averagePrice: p.averagePrice
        })),
        ...metrics.users.segments.map(s => ({
          type: 'User Segment',
          name: s.name,
          count: s.count,
          percentage: s.percentage
        }))
      ],
      charts: [
        {
          title: 'Revenue by Product',
          type: 'bar',
          data: metrics.revenue.byProduct
        },
        {
          title: 'User Segments',
          type: 'pie',
          data: metrics.users.segments
        }
      ]
    }

    return this.exportReport(data, format, {
      format,
      includeCharts: true,
      includeData: true,
      compressed: false,
      branding: true,
      ...options
    })
  }

  async exportPerformanceMetrics(
    metrics: EnhancedPerformanceMetrics,
    format: ReportFormat,
    options?: Partial<ExportOptions>
  ): Promise<ExportResult> {
    const data: ExportData = {
      title: 'Performance Analysis Report',
      metadata: {
        'Report Type': 'Performance Metrics',
        'Measurement Time': new Date().toLocaleString(),
        'Browser': navigator.userAgent.split(' ').pop()
      },
      summary: {
        'First Contentful Paint': `${metrics.webVitals.fcp}ms`,
        'Largest Contentful Paint': `${metrics.webVitals.lcp}ms`,
        'First Input Delay': `${metrics.webVitals.fid}ms`,
        'Cumulative Layout Shift': metrics.webVitals.cls.toFixed(3),
        'Memory Usage': `${metrics.runtime.memory.used} MB`
      },
      data: [
        { metric: 'FCP', value: metrics.webVitals.fcp, unit: 'ms', category: 'Web Vitals' },
        { metric: 'LCP', value: metrics.webVitals.lcp, unit: 'ms', category: 'Web Vitals' },
        { metric: 'FID', value: metrics.webVitals.fid, unit: 'ms', category: 'Web Vitals' },
        { metric: 'CLS', value: metrics.webVitals.cls, unit: 'score', category: 'Web Vitals' },
        { metric: 'TTFB', value: metrics.webVitals.ttfb, unit: 'ms', category: 'Web Vitals' },
        { metric: 'Memory Used', value: metrics.runtime.memory.used, unit: 'MB', category: 'Runtime' },
        { metric: 'CPU Cores', value: metrics.runtime.cpu.cores, unit: 'cores', category: 'Runtime' }
      ]
    }

    return this.exportReport(data, format, {
      format,
      includeCharts: true,
      includeData: true,
      compressed: false,
      branding: true,
      ...options
    })
  }

  async exportComplianceReport(
    report: ComplianceReport,
    format: ReportFormat,
    options?: Partial<ExportOptions>
  ): Promise<ExportResult> {
    const data: ExportData = {
      title: `${report.type.toUpperCase()} Compliance Report`,
      metadata: {
        'Report ID': report.id,
        'Regulation': report.type.toUpperCase(),
        'Report Period': `${new Date(report.period.start).toLocaleDateString()} - ${new Date(report.period.end).toLocaleDateString()}`,
        'Officer': report.metadata.officer.name,
        'Jurisdiction': report.metadata.jurisdiction
      },
      summary: {
        'Data Processing Records': report.data.dataProcessing.length,
        'User Rights Requests': report.data.userRights.length,
        'Security Breaches': report.data.breaches.length,
        'Consent Records': report.data.consents.length,
        'Audit Records': report.data.audits.length
      },
      data: [
        ...report.data.dataProcessing.map(dp => ({
          category: 'Data Processing',
          dataType: dp.dataType,
          purpose: dp.purpose,
          legalBasis: dp.legalBasis,
          retention: `${dp.retention} days`
        })),
        ...report.data.userRights.map(ur => ({
          category: 'User Rights',
          userId: ur.userId,
          requestType: ur.requestType,
          status: ur.status,
          timestamp: new Date(ur.timestamp).toLocaleDateString()
        })),
        ...report.data.breaches.map(br => ({
          category: 'Security Breach',
          type: br.type,
          severity: br.severity,
          affectedRecords: br.affectedRecords,
          status: br.status,
          detectedAt: new Date(br.detectedAt).toLocaleDateString()
        }))
      ]
    }

    return this.exportReport(data, format, {
      format,
      includeCharts: false,
      includeData: true,
      compressed: false,
      branding: true,
      watermark: 'CONFIDENTIAL',
      ...options
    })
  }
}

export const exportService = ExportService.getInstance()