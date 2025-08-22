/**
 * Compliance Reporting Dashboard
 * GDPR, CCPA, HIPAA, SOX, PCI-DSS compliance monitoring and reporting
 */

import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { Card } from '../../../design-system/components/Card/Card'
import { Button } from '../../../design-system/components/Button/Button'
import { Select } from '../../../design-system/components/Select/Select'
import { Input } from '../../../design-system/components/Input/Input'
import { useWebSocket } from '../../../hooks/useWebSocket'
import {
  ComplianceReport,
  ComplianceType,
  UserRightsRecord,
  BreachRecord,
  AuditRecord,
  ComplianceData,
  RequestStatus,
  UserRightType
} from '../../analytics/types'
import {
  Shield,
  AlertTriangle,
  CheckCircle,
  Clock,
  User,
  FileText,
  Download,
  Eye,
  Settings,
  Search,
  Filter,
  Calendar,
  Mail,
  Lock,
  Unlock,
  Database,
  Users,
  Activity,
  Trash2,
  Edit,
  ExternalLink,
  RefreshCw,
  Bell,
  Warning
} from 'lucide-react'

interface ComplianceMetrics {
  overall: {
    score: number
    status: 'compliant' | 'at_risk' | 'non_compliant'
    lastAudit: number
    nextAudit: number
  }
  regulations: ComplianceRegulationStatus[]
  userRights: {
    totalRequests: number
    pendingRequests: number
    averageResponseTime: number
    completionRate: number
  }
  dataProcessing: {
    totalRecords: number
    processedToday: number
    retentionViolations: number
    thirdPartyShares: number
  }
  breaches: {
    totalBreaches: number
    openBreaches: number
    averageResolutionTime: number
    reportedToAuthority: number
  }
  audits: {
    totalAudits: number
    passedAudits: number
    failedAudits: number
    averageScore: number
  }
}

interface ComplianceRegulationStatus {
  regulation: ComplianceType
  name: string
  status: 'compliant' | 'at_risk' | 'non_compliant'
  score: number
  lastAssessment: number
  violations: number
  requirements: ComplianceRequirement[]
}

interface ComplianceRequirement {
  id: string
  name: string
  status: 'met' | 'partial' | 'not_met'
  description: string
  evidence: string[]
  dueDate?: number
}

const COMPLIANCE_REGULATIONS: { value: ComplianceType; label: string; description: string }[] = [
  { value: 'gdpr', label: 'GDPR', description: 'General Data Protection Regulation (EU)' },
  { value: 'ccpa', label: 'CCPA', description: 'California Consumer Privacy Act (US)' },
  { value: 'hipaa', label: 'HIPAA', description: 'Health Insurance Portability and Accountability Act (US)' },
  { value: 'sox', label: 'SOX', description: 'Sarbanes-Oxley Act (US)' },
  { value: 'pci_dss', label: 'PCI DSS', description: 'Payment Card Industry Data Security Standard' },
  { value: 'iso27001', label: 'ISO 27001', description: 'Information Security Management Standard' }
]

const USER_RIGHT_TYPES: { value: UserRightType; label: string; description: string }[] = [
  { value: 'access', label: 'Data Access', description: 'Request access to personal data' },
  { value: 'rectification', label: 'Data Rectification', description: 'Request correction of personal data' },
  { value: 'erasure', label: 'Data Erasure', description: 'Request deletion of personal data (Right to be Forgotten)' },
  { value: 'portability', label: 'Data Portability', description: 'Request data in portable format' },
  { value: 'restriction', label: 'Processing Restriction', description: 'Request restriction of data processing' },
  { value: 'objection', label: 'Processing Objection', description: 'Object to data processing' }
]

export const ComplianceReportingDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<ComplianceMetrics | null>(null)
  const [selectedRegulation, setSelectedRegulation] = useState<ComplianceType>('gdpr')
  const [activeTab, setActiveTab] = useState('overview')
  const [userRightsRequests, setUserRightsRequests] = useState<UserRightsRecord[]>([])
  const [breaches, setBreaches] = useState<BreachRecord[]>([])
  const [audits, setAudits] = useState<AuditRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')

  const { sendMessage, lastMessage, isConnected } = useWebSocket()

  useEffect(() => {
    if (lastMessage) {
      const data = JSON.parse(lastMessage.data)
      
      if (data.type === 'compliance_metrics') {
        setMetrics(data.metrics)
        setLoading(false)
      } else if (data.type === 'user_rights') {
        setUserRightsRequests(data.requests)
      } else if (data.type === 'breaches') {
        setBreaches(data.breaches)
      } else if (data.type === 'audits') {
        setAudits(data.audits)
      }
    }
  }, [lastMessage])

  useEffect(() => {
    sendMessage({
      action: 'subscribe_compliance',
      regulation: selectedRegulation
    })

    return () => sendMessage({ action: 'unsubscribe_compliance' })
  }, [selectedRegulation, sendMessage])

  const filteredUserRights = useMemo(() => {
    return userRightsRequests.filter(request => {
      const matchesSearch = !searchTerm || 
        request.userId.toLowerCase().includes(searchTerm.toLowerCase()) ||
        request.requestType.toLowerCase().includes(searchTerm.toLowerCase())
      
      const matchesStatus = statusFilter === 'all' || request.status === statusFilter
      
      return matchesSearch && matchesStatus
    })
  }, [userRightsRequests, searchTerm, statusFilter])

  const generateComplianceReport = useCallback(async (regulation: ComplianceType, format: 'pdf' | 'csv' | 'json') => {
    try {
      sendMessage({
        action: 'generate_compliance_report',
        regulation,
        format,
        includeDetails: true
      })
    } catch (error) {
      console.error('Failed to generate compliance report:', error)
    }
  }, [sendMessage])

  const handleUserRightRequest = useCallback((requestId: string, action: 'approve' | 'reject', notes?: string) => {
    sendMessage({
      action: 'handle_user_right',
      requestId,
      decision: action,
      notes
    })
  }, [sendMessage])

  const reportBreach = useCallback((breach: Omit<BreachRecord, 'id' | 'detectedAt' | 'status'>) => {
    sendMessage({
      action: 'report_breach',
      breach: {
        ...breach,
        detectedAt: Date.now(),
        status: 'detected'
      }
    })
  }, [sendMessage])

  const getComplianceStatusColor = (status: string) => {
    switch (status) {
      case 'compliant':
        return { bg: 'bg-green-100', text: 'text-green-700', border: 'border-green-200' }
      case 'at_risk':
        return { bg: 'bg-yellow-100', text: 'text-yellow-700', border: 'border-yellow-200' }
      case 'non_compliant':
        return { bg: 'bg-red-100', text: 'text-red-700', border: 'border-red-200' }
      default:
        return { bg: 'bg-gray-100', text: 'text-gray-700', border: 'border-gray-200' }
    }
  }

  const getRequestStatusColor = (status: RequestStatus) => {
    switch (status) {
      case 'completed':
        return { bg: 'bg-green-100', text: 'text-green-700' }
      case 'processing':
        return { bg: 'bg-blue-100', text: 'text-blue-700' }
      case 'pending':
        return { bg: 'bg-yellow-100', text: 'text-yellow-700' }
      case 'rejected':
        return { bg: 'bg-red-100', text: 'text-red-700' }
      default:
        return { bg: 'bg-gray-100', text: 'text-gray-700' }
    }
  }

  const renderOverview = () => (
    <div className="space-y-6">
      {/* Overall Compliance Score */}
      <Card>
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">Overall Compliance Status</h3>
            <div className="flex items-center gap-2">
              <div className={`w-3 h-3 rounded-full ${
                metrics?.overall.status === 'compliant' ? 'bg-green-500' :
                metrics?.overall.status === 'at_risk' ? 'bg-yellow-500' : 'bg-red-500'
              }`} />
              <span className="text-sm font-medium capitalize">
                {metrics?.overall.status?.replace('_', ' ')}
              </span>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            <div className="text-center">
              <div className="relative w-24 h-24 mx-auto mb-3">
                <svg className="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
                  <circle
                    cx="50"
                    cy="50"
                    r="40"
                    stroke="#e5e7eb"
                    strokeWidth="8"
                    fill="none"
                  />
                  <circle
                    cx="50"
                    cy="50"
                    r="40"
                    stroke={
                      (metrics?.overall.score || 0) >= 80 ? '#10b981' :
                      (metrics?.overall.score || 0) >= 60 ? '#f59e0b' : '#ef4444'
                    }
                    strokeWidth="8"
                    fill="none"
                    strokeLinecap="round"
                    strokeDasharray={`${(metrics?.overall.score || 0) * 2.51} ${100 * 2.51}`}
                    className="transition-all duration-300"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-xl font-bold">{metrics?.overall.score || 0}</span>
                </div>
              </div>
              <p className="text-sm text-gray-600">Compliance Score</p>
            </div>

            <div className="text-center">
              <p className="text-2xl font-bold">{metrics?.regulations.length || 0}</p>
              <p className="text-sm text-gray-600">Regulations</p>
              <p className="text-xs text-green-600 mt-1">
                {metrics?.regulations.filter(r => r.status === 'compliant').length || 0} compliant
              </p>
            </div>

            <div className="text-center">
              <p className="text-2xl font-bold">{metrics?.userRights.totalRequests || 0}</p>
              <p className="text-sm text-gray-600">User Rights Requests</p>
              <p className="text-xs text-blue-600 mt-1">
                {metrics?.userRights.pendingRequests || 0} pending
              </p>
            </div>

            <div className="text-center">
              <p className="text-2xl font-bold">{metrics?.breaches.totalBreaches || 0}</p>
              <p className="text-sm text-gray-600">Security Breaches</p>
              <p className="text-xs text-red-600 mt-1">
                {metrics?.breaches.openBreaches || 0} open
              </p>
            </div>
          </div>
        </div>
      </Card>

      {/* Regulations Status */}
      <Card>
        <div className="p-4">
          <h3 className="text-lg font-semibold mb-4">Regulation Compliance Status</h3>
          <div className="space-y-4">
            {metrics?.regulations.map(regulation => {
              const colors = getComplianceStatusColor(regulation.status)
              return (
                <div key={regulation.regulation} className={`flex items-center justify-between p-4 rounded-lg border ${colors.border} ${colors.bg}`}>
                  <div className="flex items-center gap-3">
                    <Shield className={`w-5 h-5 ${colors.text}`} />
                    <div>
                      <h4 className="font-medium">{regulation.name}</h4>
                      <p className="text-sm text-gray-600">
                        Score: {regulation.score}% • {regulation.violations} violations
                      </p>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-2">
                    <span className={`px-3 py-1 text-xs font-medium rounded-full ${colors.bg} ${colors.text}`}>
                      {regulation.status.replace('_', ' ').toUpperCase()}
                    </span>
                    <Button
                      onClick={() => setSelectedRegulation(regulation.regulation)}
                      variant="outline"
                      size="sm"
                    >
                      <Eye className="w-4 h-4 mr-1" />
                      Details
                    </Button>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </Card>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <div className="p-4">
            <h3 className="text-lg font-semibold mb-4">Recent User Rights Requests</h3>
            <div className="space-y-3">
              {userRightsRequests.slice(0, 5).map(request => {
                const colors = getRequestStatusColor(request.status)
                return (
                  <div key={request.userId} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div className="flex items-center gap-3">
                      <User className="w-4 h-4 text-gray-600" />
                      <div>
                        <p className="font-medium text-sm">{request.requestType.replace('_', ' ')}</p>
                        <p className="text-xs text-gray-600">
                          User: {request.userId} • {new Date(request.timestamp).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                    <span className={`px-2 py-1 text-xs rounded-full ${colors.bg} ${colors.text}`}>
                      {request.status}
                    </span>
                  </div>
                )
              })}
            </div>
          </div>
        </Card>

        <Card>
          <div className="p-4">
            <h3 className="text-lg font-semibold mb-4">Security Incidents</h3>
            <div className="space-y-3">
              {breaches.slice(0, 5).map(breach => (
                <div key={breach.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <div className="flex items-center gap-3">
                    <AlertTriangle className={`w-4 h-4 ${
                      breach.severity === 'critical' ? 'text-red-600' :
                      breach.severity === 'high' ? 'text-orange-600' :
                      breach.severity === 'medium' ? 'text-yellow-600' : 'text-gray-600'
                    }`} />
                    <div>
                      <p className="font-medium text-sm">{breach.type}</p>
                      <p className="text-xs text-gray-600">
                        {breach.affectedRecords} records • {new Date(breach.detectedAt).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                  <span className={`px-2 py-1 text-xs rounded-full ${
                    breach.status === 'resolved' ? 'bg-green-100 text-green-700' :
                    breach.status === 'contained' ? 'bg-blue-100 text-blue-700' :
                    breach.status === 'investigating' ? 'bg-yellow-100 text-yellow-700' :
                    'bg-red-100 text-red-700'
                  }`}>
                    {breach.status}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </Card>
      </div>
    </div>
  )

  const renderUserRights = () => (
    <div className="space-y-6">
      {/* Filters */}
      <Card>
        <div className="p-4">
          <div className="flex flex-wrap gap-4 items-center">
            <div className="flex-1 min-w-64">
              <Input
                type="text"
                placeholder="Search by user ID or request type..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full"
                icon={<Search className="w-4 h-4" />}
              />
            </div>
            
            <Select
              value={statusFilter}
              onChange={setStatusFilter}
              className="w-40"
            >
              <option value="all">All Status</option>
              <option value="pending">Pending</option>
              <option value="processing">Processing</option>
              <option value="completed">Completed</option>
              <option value="rejected">Rejected</option>
            </Select>
          </div>
        </div>
      </Card>

      {/* Requests List */}
      <Card>
        <div className="p-4">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold">User Rights Requests</h3>
            <p className="text-sm text-gray-600">
              {filteredUserRights.length} of {userRightsRequests.length} requests
            </p>
          </div>
          
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-3 px-2">User ID</th>
                  <th className="text-left py-3 px-2">Request Type</th>
                  <th className="text-left py-3 px-2">Date</th>
                  <th className="text-left py-3 px-2">Status</th>
                  <th className="text-left py-3 px-2">Response Time</th>
                  <th className="text-left py-3 px-2">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredUserRights.map(request => {
                  const colors = getRequestStatusColor(request.status)
                  const responseTime = request.status === 'completed' && request.resolution 
                    ? Math.round((Date.now() - request.timestamp) / (1000 * 60 * 60 * 24))
                    : null
                  
                  return (
                    <tr key={`${request.userId}-${request.timestamp}`} className="border-b hover:bg-gray-50">
                      <td className="py-3 px-2 font-medium">{request.userId}</td>
                      <td className="py-3 px-2 capitalize">
                        {request.requestType.replace('_', ' ')}
                      </td>
                      <td className="py-3 px-2">
                        {new Date(request.timestamp).toLocaleDateString()}
                      </td>
                      <td className="py-3 px-2">
                        <span className={`px-2 py-1 text-xs rounded-full ${colors.bg} ${colors.text}`}>
                          {request.status}
                        </span>
                      </td>
                      <td className="py-3 px-2">
                        {responseTime ? `${responseTime} days` : '-'}
                      </td>
                      <td className="py-3 px-2">
                        <div className="flex gap-2">
                          {request.status === 'pending' && (
                            <>
                              <Button
                                onClick={() => handleUserRightRequest(`${request.userId}-${request.timestamp}`, 'approve')}
                                variant="outline"
                                size="sm"
                                className="text-green-600 hover:bg-green-50"
                              >
                                <CheckCircle className="w-3 h-3" />
                              </Button>
                              <Button
                                onClick={() => handleUserRightRequest(`${request.userId}-${request.timestamp}`, 'reject')}
                                variant="outline"
                                size="sm"
                                className="text-red-600 hover:bg-red-50"
                              >
                                <Trash2 className="w-3 h-3" />
                              </Button>
                            </>
                          )}
                          <Button variant="outline" size="sm">
                            <Eye className="w-3 h-3" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      </Card>
    </div>
  )

  const tabs = [
    { id: 'overview', label: 'Overview', icon: <Eye className="w-4 h-4" /> },
    { id: 'user_rights', label: 'User Rights', icon: <User className="w-4 h-4" /> },
    { id: 'breaches', label: 'Security Incidents', icon: <AlertTriangle className="w-4 h-4" /> },
    { id: 'audits', label: 'Audit Trail', icon: <FileText className="w-4 h-4" /> }
  ]

  if (loading && !metrics) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Compliance Dashboard</h1>
          <p className="text-gray-600 mt-1">GDPR, CCPA, HIPAA and other regulatory compliance</p>
        </div>

        <div className="flex items-center gap-3">
          <Select
            value={selectedRegulation}
            onChange={(value) => setSelectedRegulation(value as ComplianceType)}
            className="w-40"
          >
            {COMPLIANCE_REGULATIONS.map(regulation => (
              <option key={regulation.value} value={regulation.value}>
                {regulation.label}
              </option>
            ))}
          </Select>

          <Button
            onClick={() => generateComplianceReport(selectedRegulation, 'pdf')}
            variant="outline"
            size="sm"
          >
            <Download className="w-4 h-4 mr-2" />
            Export Report
          </Button>

          <Button variant="outline" size="sm">
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
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

      {/* Content */}
      {activeTab === 'overview' && renderOverview()}
      {activeTab === 'user_rights' && renderUserRights()}
      {activeTab === 'breaches' && (
        <Card>
          <div className="p-4">
            <h3 className="text-lg font-semibold">Security Incidents - Coming Soon</h3>
            <p className="text-gray-600 mt-2">Detailed security incident management interface</p>
          </div>
        </Card>
      )}
      {activeTab === 'audits' && (
        <Card>
          <div className="p-4">
            <h3 className="text-lg font-semibold">Audit Trail - Coming Soon</h3>
            <p className="text-gray-600 mt-2">Comprehensive audit log and compliance tracking</p>
          </div>
        </Card>
      )}
    </div>
  )
}

export default ComplianceReportingDashboard