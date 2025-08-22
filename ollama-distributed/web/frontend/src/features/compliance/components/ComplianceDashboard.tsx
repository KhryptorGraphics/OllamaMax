import React, { useState, useEffect } from 'react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { useWebSocket } from '@/hooks/useWebSocket'
import {
  Shield, CheckCircle, AlertTriangle, XCircle, FileText,
  Lock, Key, Eye, Users, Database, Globe, Clock, Download
} from 'lucide-react'

interface ComplianceStatus {
  framework: string
  status: 'compliant' | 'partial' | 'non-compliant' | 'pending'
  score: number
  lastAudit: string
  nextAudit: string
  findings: number
  criticalIssues: number
}

interface DataPrivacy {
  dataTypes: string[]
  retention: string
  encryption: 'at-rest' | 'in-transit' | 'both'
  anonymization: boolean
  rightsRequests: number
}

interface AuditLog {
  id: string
  timestamp: string
  user: string
  action: string
  resource: string
  outcome: 'success' | 'failure'
  ip: string
}

export const ComplianceDashboard: React.FC = () => {
  const [frameworks, setFrameworks] = useState<ComplianceStatus[]>([
    {
      framework: 'GDPR',
      status: 'compliant',
      score: 95,
      lastAudit: '2024-01-15',
      nextAudit: '2024-04-15',
      findings: 2,
      criticalIssues: 0
    },
    {
      framework: 'SOC 2 Type II',
      status: 'compliant',
      score: 92,
      lastAudit: '2024-01-10',
      nextAudit: '2024-07-10',
      findings: 5,
      criticalIssues: 0
    },
    {
      framework: 'HIPAA',
      status: 'partial',
      score: 78,
      lastAudit: '2024-01-20',
      nextAudit: '2024-03-20',
      findings: 12,
      criticalIssues: 2
    },
    {
      framework: 'ISO 27001',
      status: 'compliant',
      score: 88,
      lastAudit: '2023-12-01',
      nextAudit: '2024-06-01',
      findings: 8,
      criticalIssues: 1
    }
  ])
  const [dataPrivacy, setDataPrivacy] = useState<DataPrivacy>({
    dataTypes: ['PII', 'PHI', 'Financial', 'Behavioral'],
    retention: '7 years',
    encryption: 'both',
    anonymization: true,
    rightsRequests: 42
  })
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])
  const [selectedTab, setSelectedTab] = useState('overview')

  const { sendMessage, lastMessage } = useWebSocket('/ws/compliance')

  useEffect(() => {
    if (lastMessage) {
      const data = JSON.parse(lastMessage.data)
      if (data.type === 'frameworks') setFrameworks(data.frameworks)
      if (data.type === 'privacy') setDataPrivacy(data.privacy)
      if (data.type === 'audit_logs') setAuditLogs(data.logs)
    }
  }, [lastMessage])

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'compliant': return <CheckCircle className="w-5 h-5 text-green-500" />
      case 'partial': return <AlertTriangle className="w-5 h-5 text-yellow-500" />
      case 'non-compliant': return <XCircle className="w-5 h-5 text-red-500" />
      default: return <Clock className="w-5 h-5 text-gray-500" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'compliant': return 'bg-green-100 text-green-700'
      case 'partial': return 'bg-yellow-100 text-yellow-700'
      case 'non-compliant': return 'bg-red-100 text-red-700'
      default: return 'bg-gray-100 text-gray-700'
    }
  }

  const generateComplianceReport = (framework: string) => {
    sendMessage({ action: 'generate_report', framework })
  }

  const exportAuditLogs = () => {
    const csv = [
      ['Timestamp', 'User', 'Action', 'Resource', 'Outcome', 'IP'],
      ...auditLogs.map(log => [
        log.timestamp, log.user, log.action, log.resource, log.outcome, log.ip
      ])
    ].map(row => row.join(',')).join('\n')

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'audit-logs.csv'
    a.click()
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Compliance Management</h1>
        <Button onClick={exportAuditLogs}>
          <Download className="w-4 h-4 mr-2" />
          Export Audit Logs
        </Button>
      </div>

      {/* Compliance Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {frameworks.map(framework => (
          <Card key={framework.framework}>
            <div className="p-4">
              <div className="flex items-center justify-between mb-3">
                {getStatusIcon(framework.status)}
                <span className={`px-2 py-1 text-xs rounded-full ${getStatusColor(framework.status)}`}>
                  {framework.status}
                </span>
              </div>
              <h3 className="font-semibold text-lg mb-2">{framework.framework}</h3>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Score</span>
                  <span className="font-medium">{framework.score}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className={`h-2 rounded-full ${
                      framework.score >= 90 ? 'bg-green-600' :
                      framework.score >= 70 ? 'bg-yellow-600' :
                      'bg-red-600'
                    }`}
                    style={{ width: `${framework.score}%` }}
                  />
                </div>
                <div className="flex justify-between text-xs text-gray-600">
                  <span>{framework.findings} findings</span>
                  {framework.criticalIssues > 0 && (
                    <span className="text-red-600">{framework.criticalIssues} critical</span>
                  )}
                </div>
                <div className="pt-2">
                  <Button 
                    size="sm" 
                    variant="outline" 
                    className="w-full"
                    onClick={() => generateComplianceReport(framework.framework)}
                  >
                    <FileText className="w-3 h-3 mr-1" />
                    Generate Report
                  </Button>
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>

      {/* Data Privacy & Protection */}
      <Card>
        <div className="p-4">
          <h2 className="text-xl font-semibold mb-4">Data Privacy & Protection</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <h3 className="font-medium mb-2 flex items-center gap-2">
                <Database className="w-4 h-4" />
                Data Classification
              </h3>
              <div className="space-y-1">
                {dataPrivacy.dataTypes.map(type => (
                  <div key={type} className="flex items-center gap-2">
                    <CheckCircle className="w-3 h-3 text-green-500" />
                    <span className="text-sm">{type}</span>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <h3 className="font-medium mb-2 flex items-center gap-2">
                <Lock className="w-4 h-4" />
                Security Measures
              </h3>
              <div className="space-y-1 text-sm">
                <div>Encryption: {dataPrivacy.encryption}</div>
                <div>Retention: {dataPrivacy.retention}</div>
                <div>Anonymization: {dataPrivacy.anonymization ? 'Enabled' : 'Disabled'}</div>
              </div>
            </div>

            <div>
              <h3 className="font-medium mb-2 flex items-center gap-2">
                <Users className="w-4 h-4" />
                User Rights
              </h3>
              <div className="space-y-1 text-sm">
                <div>Rights Requests: {dataPrivacy.rightsRequests}</div>
                <div>Avg Response: 24 hours</div>
                <div>Compliance Rate: 100%</div>
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* Audit Trail */}
      <Card>
        <div className="p-4">
          <h2 className="text-xl font-semibold mb-4">Recent Audit Activity</h2>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-2">Timestamp</th>
                  <th className="text-left py-2">User</th>
                  <th className="text-left py-2">Action</th>
                  <th className="text-left py-2">Resource</th>
                  <th className="text-left py-2">Status</th>
                </tr>
              </thead>
              <tbody>
                {auditLogs.slice(0, 10).map(log => (
                  <tr key={log.id} className="border-b">
                    <td className="py-2 text-sm">{new Date(log.timestamp).toLocaleString()}</td>
                    <td className="py-2 text-sm">{log.user}</td>
                    <td className="py-2 text-sm">{log.action}</td>
                    <td className="py-2 text-sm">{log.resource}</td>
                    <td className="py-2">
                      <span className={`text-xs px-2 py-1 rounded-full ${
                        log.outcome === 'success' 
                          ? 'bg-green-100 text-green-700' 
                          : 'bg-red-100 text-red-700'
                      }`}>
                        {log.outcome}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </Card>

      {/* Compliance Actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <div className="p-4 text-center">
            <Shield className="w-12 h-12 mx-auto mb-3 text-blue-500" />
            <h3 className="font-semibold mb-2">Security Assessment</h3>
            <p className="text-sm text-gray-600 mb-3">
              Run comprehensive security and compliance assessment
            </p>
            <Button variant="outline" className="w-full">Start Assessment</Button>
          </div>
        </Card>

        <Card>
          <div className="p-4 text-center">
            <Eye className="w-12 h-12 mx-auto mb-3 text-green-500" />
            <h3 className="font-semibold mb-2">Privacy Impact Assessment</h3>
            <p className="text-sm text-gray-600 mb-3">
              Evaluate data processing activities for privacy risks
            </p>
            <Button variant="outline" className="w-full">Begin PIA</Button>
          </div>
        </Card>

        <Card>
          <div className="p-4 text-center">
            <Key className="w-12 h-12 mx-auto mb-3 text-purple-500" />
            <h3 className="font-semibold mb-2">Access Review</h3>
            <p className="text-sm text-gray-600 mb-3">
              Review and audit user access permissions
            </p>
            <Button variant="outline" className="w-full">Review Access</Button>
          </div>
        </Card>
      </div>
    </div>
  )
}