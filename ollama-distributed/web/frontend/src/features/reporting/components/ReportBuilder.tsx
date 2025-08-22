/**
 * Report Builder Component
 * Provides an intuitive interface for creating and customizing reports
 */

import React, { useState, useCallback, useEffect } from 'react'
import { Card } from '../../../design-system/components/Card/Card'
import { Button } from '../../../design-system/components/Button/Button'
import { Input } from '../../../design-system/components/Input/Input'
import { Select } from '../../../design-system/components/Select/Select'
import { Chart } from '../../analytics/components/charts/Chart'
import {
  Report,
  ReportType,
  ReportFormat,
  ReportFilter,
  FilterOperator,
  ReportSchedule,
  DateRange,
  DateRangePreset,
  Widget,
  WidgetType,
  ChartType
} from '../../analytics/types'
import {
  Calendar,
  Filter,
  Download,
  Eye,
  Settings,
  Plus,
  Trash2,
  Copy,
  Save,
  Play,
  BarChart,
  PieChart,
  LineChart,
  Table,
  FileText,
  Clock,
  Users,
  Mail
} from 'lucide-react'

interface ReportBuilderProps {
  initialReport?: Partial<Report>
  onSave?: (report: Report) => void
  onPreview?: (report: Report) => void
  onCancel?: () => void
}

const REPORT_TYPES: { value: ReportType; label: string; icon: React.ReactNode }[] = [
  { value: 'analytics', label: 'Analytics Report', icon: <BarChart className="w-4 h-4" /> },
  { value: 'performance', label: 'Performance Report', icon: <LineChart className="w-4 h-4" /> },
  { value: 'business', label: 'Business Intelligence', icon: <PieChart className="w-4 h-4" /> },
  { value: 'compliance', label: 'Compliance Report', icon: <FileText className="w-4 h-4" /> },
  { value: 'security', label: 'Security Report', icon: <Shield className="w-4 h-4" /> },
  { value: 'user_behavior', label: 'User Behavior', icon: <Users className="w-4 h-4" /> },
  { value: 'custom', label: 'Custom Report', icon: <Settings className="w-4 h-4" /> }
]

const REPORT_FORMATS: { value: ReportFormat; label: string }[] = [
  { value: 'pdf', label: 'PDF Document' },
  { value: 'csv', label: 'CSV Spreadsheet' },
  { value: 'excel', label: 'Excel Workbook' },
  { value: 'html', label: 'HTML Page' },
  { value: 'json', label: 'JSON Data' }
]

const DATE_PRESETS: { value: DateRangePreset; label: string }[] = [
  { value: 'today', label: 'Today' },
  { value: 'yesterday', label: 'Yesterday' },
  { value: 'last_7_days', label: 'Last 7 Days' },
  { value: 'last_30_days', label: 'Last 30 Days' },
  { value: 'last_90_days', label: 'Last 90 Days' },
  { value: 'this_month', label: 'This Month' },
  { value: 'last_month', label: 'Last Month' },
  { value: 'this_quarter', label: 'This Quarter' },
  { value: 'last_quarter', label: 'Last Quarter' },
  { value: 'this_year', label: 'This Year' },
  { value: 'last_year', label: 'Last Year' },
  { value: 'custom', label: 'Custom Range' }
]

const WIDGET_TYPES: { value: WidgetType; label: string; icon: React.ReactNode }[] = [
  { value: 'metric', label: 'Metric Card', icon: <div className="w-4 h-4 bg-blue-500 rounded" /> },
  { value: 'chart', label: 'Chart', icon: <BarChart className="w-4 h-4" /> },
  { value: 'table', label: 'Data Table', icon: <Table className="w-4 h-4" /> },
  { value: 'text', label: 'Text Block', icon: <FileText className="w-4 h-4" /> }
]

const FILTER_OPERATORS: { value: FilterOperator; label: string }[] = [
  { value: 'equals', label: 'Equals' },
  { value: 'not_equals', label: 'Not Equals' },
  { value: 'contains', label: 'Contains' },
  { value: 'not_contains', label: 'Not Contains' },
  { value: 'starts_with', label: 'Starts With' },
  { value: 'ends_with', label: 'Ends With' },
  { value: 'greater_than', label: 'Greater Than' },
  { value: 'less_than', label: 'Less Than' },
  { value: 'between', label: 'Between' },
  { value: 'in', label: 'In List' },
  { value: 'not_in', label: 'Not In List' }
]

export const ReportBuilder: React.FC<ReportBuilderProps> = ({
  initialReport,
  onSave,
  onPreview,
  onCancel
}) => {
  const [report, setReport] = useState<Partial<Report>>({
    id: '',
    name: '',
    type: 'analytics',
    format: 'pdf',
    schedule: {
      frequency: 'once',
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      enabled: false
    },
    recipients: [],
    filters: [],
    status: 'draft',
    ...initialReport
  })

  const [activeTab, setActiveTab] = useState('basic')
  const [previewData, setPreviewData] = useState<any>(null)
  const [isGeneratingPreview, setIsGeneratingPreview] = useState(false)

  const updateReport = useCallback((updates: Partial<Report>) => {
    setReport(prev => ({ ...prev, ...updates }))
  }, [])

  const addFilter = useCallback(() => {
    const newFilter: ReportFilter = {
      field: '',
      operator: 'equals',
      value: '',
      label: 'New Filter'
    }
    setReport(prev => ({
      ...prev,
      filters: [...(prev.filters || []), newFilter]
    }))
  }, [])

  const updateFilter = useCallback((index: number, updates: Partial<ReportFilter>) => {
    setReport(prev => ({
      ...prev,
      filters: prev.filters?.map((filter, i) => 
        i === index ? { ...filter, ...updates } : filter
      ) || []
    }))
  }, [])

  const removeFilter = useCallback((index: number) => {
    setReport(prev => ({
      ...prev,
      filters: prev.filters?.filter((_, i) => i !== index) || []
    }))
  }, [])

  const addRecipient = useCallback((email: string) => {
    if (email && !report.recipients?.includes(email)) {
      setReport(prev => ({
        ...prev,
        recipients: [...(prev.recipients || []), email]
      }))
    }
  }, [report.recipients])

  const removeRecipient = useCallback((email: string) => {
    setReport(prev => ({
      ...prev,
      recipients: prev.recipients?.filter(r => r !== email) || []
    }))
  }, [])

  const generatePreview = useCallback(async () => {
    if (!report.name || !report.type) return

    setIsGeneratingPreview(true)
    try {
      // Simulate API call to generate preview data
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      // Mock preview data based on report type
      const mockData = generateMockData(report.type!)
      setPreviewData(mockData)
      
      if (onPreview) {
        onPreview(report as Report)
      }
    } catch (error) {
      console.error('Failed to generate preview:', error)
    } finally {
      setIsGeneratingPreview(false)
    }
  }, [report, onPreview])

  const handleSave = useCallback(() => {
    if (!report.name || !report.type) return

    const completeReport: Report = {
      ...report,
      id: report.id || `report_${Date.now()}`,
      generatedAt: Date.now(),
      status: 'draft'
    } as Report

    onSave?.(completeReport)
  }, [report, onSave])

  const generateMockData = (type: ReportType) => {
    switch (type) {
      case 'analytics':
        return {
          metrics: {
            totalUsers: 1234,
            pageViews: 5678,
            sessionDuration: 180,
            bounceRate: 0.25
          },
          chartData: Array.from({ length: 30 }, (_, i) => ({
            date: new Date(Date.now() - i * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
            users: Math.floor(Math.random() * 1000) + 500,
            sessions: Math.floor(Math.random() * 1500) + 800
          }))
        }
      
      case 'performance':
        return {
          webVitals: {
            fcp: 1200,
            lcp: 2100,
            fid: 80,
            cls: 0.05,
            ttfb: 400
          },
          chartData: Array.from({ length: 24 }, (_, i) => ({
            hour: `${i}:00`,
            responseTime: Math.floor(Math.random() * 200) + 100,
            throughput: Math.floor(Math.random() * 1000) + 500
          }))
        }
      
      default:
        return {
          metrics: { total: 100, active: 80, success: 90 },
          chartData: []
        }
    }
  }

  const renderBasicSettings = () => (
    <div className="space-y-6">
      <div>
        <label className="block text-sm font-medium mb-2">Report Name</label>
        <Input
          value={report.name || ''}
          onChange={(e) => updateReport({ name: e.target.value })}
          placeholder="Enter report name"
          className="w-full"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">Report Type</label>
        <Select
          value={report.type || 'analytics'}
          onChange={(value) => updateReport({ type: value as ReportType })}
          className="w-full"
        >
          {REPORT_TYPES.map(type => (
            <option key={type.value} value={type.value}>
              {type.label}
            </option>
          ))}
        </Select>
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">Output Format</label>
        <Select
          value={report.format || 'pdf'}
          onChange={(value) => updateReport({ format: value as ReportFormat })}
          className="w-full"
        >
          {REPORT_FORMATS.map(format => (
            <option key={format.value} value={format.value}>
              {format.label}
            </option>
          ))}
        </Select>
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">Date Range</label>
        <Select
          value="last_30_days"
          onChange={() => {}}
          className="w-full"
        >
          {DATE_PRESETS.map(preset => (
            <option key={preset.value} value={preset.value}>
              {preset.label}
            </option>
          ))}
        </Select>
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">Description</label>
        <textarea
          value={report.description || ''}
          onChange={(e) => updateReport({ description: e.target.value })}
          placeholder="Brief description of this report"
          className="w-full p-3 border border-gray-300 rounded-lg resize-none"
          rows={3}
        />
      </div>
    </div>
  )

  const renderFilters = () => (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-medium">Data Filters</h3>
        <Button onClick={addFilter} variant="outline" size="sm">
          <Plus className="w-4 h-4 mr-2" />
          Add Filter
        </Button>
      </div>

      {report.filters?.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          <Filter className="w-12 h-12 mx-auto mb-3 text-gray-400" />
          <p>No filters configured</p>
          <p className="text-sm">Add filters to narrow down your data</p>
        </div>
      )}

      <div className="space-y-4">
        {report.filters?.map((filter, index) => (
          <Card key={index}>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
              <div>
                <label className="block text-sm font-medium mb-1">Field</label>
                <Input
                  value={filter.field}
                  onChange={(e) => updateFilter(index, { field: e.target.value })}
                  placeholder="Field name"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">Operator</label>
                <Select
                  value={filter.operator}
                  onChange={(value) => updateFilter(index, { operator: value as FilterOperator })}
                >
                  {FILTER_OPERATORS.map(op => (
                    <option key={op.value} value={op.value}>
                      {op.label}
                    </option>
                  ))}
                </Select>
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">Value</label>
                <Input
                  value={filter.value?.toString() || ''}
                  onChange={(e) => updateFilter(index, { value: e.target.value })}
                  placeholder="Filter value"
                />
              </div>
              
              <div>
                <Button
                  onClick={() => removeFilter(index)}
                  variant="outline"
                  size="sm"
                  className="text-red-600 hover:bg-red-50"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  )

  const renderSchedule = () => (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <input
          type="checkbox"
          id="enable-schedule"
          checked={report.schedule?.enabled || false}
          onChange={(e) => updateReport({
            schedule: { ...report.schedule!, enabled: e.target.checked }
          })}
          className="rounded"
        />
        <label htmlFor="enable-schedule" className="font-medium">
          Enable Scheduled Generation
        </label>
      </div>

      {report.schedule?.enabled && (
        <>
          <div>
            <label className="block text-sm font-medium mb-2">Frequency</label>
            <Select
              value={report.schedule?.frequency || 'once'}
              onChange={(value) => updateReport({
                schedule: { ...report.schedule!, frequency: value as any }
              })}
              className="w-full"
            >
              <option value="once">One Time</option>
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="monthly">Monthly</option>
              <option value="quarterly">Quarterly</option>
              <option value="yearly">Yearly</option>
            </Select>
          </div>

          {report.schedule?.frequency !== 'once' && (
            <div>
              <label className="block text-sm font-medium mb-2">Time</label>
              <Input
                type="time"
                value={report.schedule?.time || '09:00'}
                onChange={(e) => updateReport({
                  schedule: { ...report.schedule!, time: e.target.value }
                })}
                className="w-full"
              />
            </div>
          )}

          <div>
            <label className="block text-sm font-medium mb-2">Timezone</label>
            <Select
              value={report.schedule?.timezone || 'UTC'}
              onChange={(value) => updateReport({
                schedule: { ...report.schedule!, timezone: value }
              })}
              className="w-full"
            >
              <option value="UTC">UTC</option>
              <option value="America/New_York">Eastern Time</option>
              <option value="America/Los_Angeles">Pacific Time</option>
              <option value="Europe/London">London</option>
              <option value="Asia/Tokyo">Tokyo</option>
            </Select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Recipients</label>
            <div className="space-y-2">
              <div className="flex gap-2">
                <Input
                  type="email"
                  placeholder="Enter email address"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      addRecipient((e.target as HTMLInputElement).value)
                      ;(e.target as HTMLInputElement).value = ''
                    }
                  }}
                  className="flex-1"
                />
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    const input = document.querySelector('input[type="email"]') as HTMLInputElement
                    if (input?.value) {
                      addRecipient(input.value)
                      input.value = ''
                    }
                  }}
                >
                  <Plus className="w-4 h-4" />
                </Button>
              </div>
              
              <div className="flex flex-wrap gap-2">
                {report.recipients?.map(email => (
                  <div
                    key={email}
                    className="flex items-center gap-2 bg-blue-50 text-blue-700 px-3 py-1 rounded-full text-sm"
                  >
                    <Mail className="w-3 h-3" />
                    {email}
                    <button
                      onClick={() => removeRecipient(email)}
                      className="text-blue-500 hover:text-blue-700"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  )

  const renderPreview = () => (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-medium">Report Preview</h3>
        <Button
          onClick={generatePreview}
          disabled={isGeneratingPreview || !report.name || !report.type}
          variant="outline"
        >
          {isGeneratingPreview ? (
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500 mr-2" />
          ) : (
            <Eye className="w-4 h-4 mr-2" />
          )}
          Generate Preview
        </Button>
      </div>

      {previewData ? (
        <div className="space-y-6">
          <Card>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {Object.entries(previewData.metrics || {}).map(([key, value]) => (
                <div key={key} className="text-center">
                  <p className="text-2xl font-bold">{value as number}</p>
                  <p className="text-sm text-gray-600 capitalize">
                    {key.replace(/([A-Z])/g, ' $1').trim()}
                  </p>
                </div>
              ))}
            </div>
          </Card>

          {previewData.chartData && previewData.chartData.length > 0 && (
            <Card>
              <Chart
                type="line"
                data={previewData.chartData}
                config={{
                  xKey: Object.keys(previewData.chartData[0])[0],
                  yKey: Object.keys(previewData.chartData[0]).slice(1),
                  showGrid: true,
                  showLegend: true
                }}
                height={300}
              />
            </Card>
          )}
        </div>
      ) : (
        <div className="text-center py-12 text-gray-500">
          <Eye className="w-12 h-12 mx-auto mb-3 text-gray-400" />
          <p>Click "Generate Preview" to see how your report will look</p>
        </div>
      )}
    </div>
  )

  const tabs = [
    { id: 'basic', label: 'Basic Settings', icon: <Settings className="w-4 h-4" /> },
    { id: 'filters', label: 'Filters', icon: <Filter className="w-4 h-4" /> },
    { id: 'schedule', label: 'Schedule', icon: <Clock className="w-4 h-4" /> },
    { id: 'preview', label: 'Preview', icon: <Eye className="w-4 h-4" /> }
  ]

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold">
            {initialReport ? 'Edit Report' : 'Create New Report'}
          </h1>
          <p className="text-gray-600 mt-1">
            Configure your report settings and data filters
          </p>
        </div>
        
        <div className="flex gap-3">
          {onCancel && (
            <Button onClick={onCancel} variant="outline">
              Cancel
            </Button>
          )}
          <Button onClick={handleSave} disabled={!report.name || !report.type}>
            <Save className="w-4 h-4 mr-2" />
            Save Report
          </Button>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          {tabs.map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`py-2 px-1 border-b-2 font-medium text-sm flex items-center gap-2 ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <Card className="min-h-[500px]">
        {activeTab === 'basic' && renderBasicSettings()}
        {activeTab === 'filters' && renderFilters()}
        {activeTab === 'schedule' && renderSchedule()}
        {activeTab === 'preview' && renderPreview()}
      </Card>
    </div>
  )
}

export default ReportBuilder