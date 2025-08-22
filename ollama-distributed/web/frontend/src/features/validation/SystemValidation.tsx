import React, { useState, useEffect } from 'react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { useWebSocket } from '@/hooks/useWebSocket'
import {
  CheckCircle, XCircle, AlertTriangle, Clock, Play,
  RefreshCw, FileText, Shield, Zap, Database, Globe
} from 'lucide-react'

interface ValidationTest {
  id: string
  category: 'functional' | 'performance' | 'security' | 'integration'
  name: string
  description: string
  status: 'pending' | 'running' | 'passed' | 'failed' | 'skipped'
  duration?: number
  error?: string
  assertions: { name: string; passed: boolean }[]
}

interface BenchmarkResult {
  metric: string
  value: number
  unit: string
  baseline: number
  target: number
  percentChange: number
}

interface SecurityScan {
  vulnerability: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  affected: string[]
  remediation: string
  cve?: string
}

export const SystemValidation: React.FC = () => {
  const [tests, setTests] = useState<ValidationTest[]>([])
  const [benchmarks, setBenchmarks] = useState<BenchmarkResult[]>([])
  const [securityScans, setSecurityScans] = useState<SecurityScan[]>([])
  const [validationStatus, setValidationStatus] = useState('idle')
  const [progress, setProgress] = useState(0)

  const { sendMessage, lastMessage } = useWebSocket('/ws/validation')

  useEffect(() => {
    if (lastMessage) {
      const data = JSON.parse(lastMessage.data)
      if (data.type === 'tests') setTests(data.tests)
      if (data.type === 'benchmarks') setBenchmarks(data.benchmarks)
      if (data.type === 'security') setSecurityScans(data.scans)
      if (data.type === 'progress') setProgress(data.progress)
      if (data.type === 'status') setValidationStatus(data.status)
    }
  }, [lastMessage])

  const runValidation = (category?: string) => {
    setValidationStatus('running')
    sendMessage({ action: 'run_validation', category })
  }

  const exportReport = () => {
    const report = {
      timestamp: new Date().toISOString(),
      tests: tests.filter(t => t.status !== 'pending'),
      benchmarks,
      securityScans,
      summary: {
        passed: tests.filter(t => t.status === 'passed').length,
        failed: tests.filter(t => t.status === 'failed').length,
        total: tests.length
      }
    }

    const blob = new Blob([JSON.stringify(report, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `validation-report-${Date.now()}.json`
    a.click()
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'passed': return <CheckCircle className="w-5 h-5 text-green-500" />
      case 'failed': return <XCircle className="w-5 h-5 text-red-500" />
      case 'running': return <RefreshCw className="w-5 h-5 text-blue-500 animate-spin" />
      case 'skipped': return <AlertTriangle className="w-5 h-5 text-gray-400" />
      default: return <Clock className="w-5 h-5 text-gray-400" />
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'bg-red-100 text-red-700'
      case 'high': return 'bg-orange-100 text-orange-700'
      case 'medium': return 'bg-yellow-100 text-yellow-700'
      case 'low': return 'bg-blue-100 text-blue-700'
      default: return 'bg-gray-100 text-gray-700'
    }
  }

  const testsByCategory = tests.reduce((acc, test) => {
    if (!acc[test.category]) acc[test.category] = []
    acc[test.category].push(test)
    return acc
  }, {} as Record<string, ValidationTest[]>)

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">System Validation</h1>
        <div className="flex gap-4">
          <Button 
            onClick={() => runValidation()} 
            disabled={validationStatus === 'running'}
          >
            <Play className="w-4 h-4 mr-2" />
            Run Full Validation
          </Button>
          <Button variant="outline" onClick={exportReport}>
            <FileText className="w-4 h-4 mr-2" />
            Export Report
          </Button>
        </div>
      </div>

      {/* Validation Progress */}
      {validationStatus === 'running' && (
        <Card>
          <div className="p-4">
            <div className="flex justify-between mb-2">
              <span className="text-sm font-medium">Validation Progress</span>
              <span className="text-sm">{progress}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div 
                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                style={{ width: `${progress}%` }}
              />
            </div>
          </div>
        </Card>
      )}

      {/* Test Results by Category */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {Object.entries(testsByCategory).map(([category, categoryTests]) => (
          <Card key={category}>
            <div className="p-4">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold capitalize">{category} Tests</h3>
                <Button 
                  size="sm" 
                  variant="outline"
                  onClick={() => runValidation(category)}
                  disabled={validationStatus === 'running'}
                >
                  <RefreshCw className="w-3 h-3 mr-1" />
                  Rerun
                </Button>
              </div>
              <div className="space-y-2">
                {categoryTests.map(test => (
                  <div key={test.id} className="border rounded p-3">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        {getStatusIcon(test.status)}
                        <span className="font-medium">{test.name}</span>
                      </div>
                      {test.duration && (
                        <span className="text-xs text-gray-600">{test.duration}ms</span>
                      )}
                    </div>
                    <p className="text-sm text-gray-600 mb-2">{test.description}</p>
                    {test.error && (
                      <div className="bg-red-50 border border-red-200 rounded p-2 mb-2">
                        <p className="text-xs text-red-700">{test.error}</p>
                      </div>
                    )}
                    {test.assertions.length > 0 && (
                      <div className="space-y-1">
                        {test.assertions.map((assertion, idx) => (
                          <div key={idx} className="flex items-center gap-2 text-xs">
                            {assertion.passed ? (
                              <CheckCircle className="w-3 h-3 text-green-500" />
                            ) : (
                              <XCircle className="w-3 h-3 text-red-500" />
                            )}
                            <span>{assertion.name}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </Card>
        ))}
      </div>

      {/* Performance Benchmarks */}
      <Card>
        <div className="p-4">
          <h2 className="text-xl font-semibold mb-4">Performance Benchmarks</h2>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-2">Metric</th>
                  <th className="text-right py-2">Current</th>
                  <th className="text-right py-2">Baseline</th>
                  <th className="text-right py-2">Target</th>
                  <th className="text-right py-2">Change</th>
                  <th className="text-center py-2">Status</th>
                </tr>
              </thead>
              <tbody>
                {benchmarks.map(benchmark => (
                  <tr key={benchmark.metric} className="border-b">
                    <td className="py-2">{benchmark.metric}</td>
                    <td className="py-2 text-right">
                      {benchmark.value} {benchmark.unit}
                    </td>
                    <td className="py-2 text-right">
                      {benchmark.baseline} {benchmark.unit}
                    </td>
                    <td className="py-2 text-right">
                      {benchmark.target} {benchmark.unit}
                    </td>
                    <td className="py-2 text-right">
                      <span className={benchmark.percentChange > 0 ? 'text-red-600' : 'text-green-600'}>
                        {benchmark.percentChange > 0 ? '+' : ''}{benchmark.percentChange}%
                      </span>
                    </td>
                    <td className="py-2 text-center">
                      {benchmark.value <= benchmark.target ? (
                        <CheckCircle className="w-4 h-4 text-green-500 inline" />
                      ) : (
                        <XCircle className="w-4 h-4 text-red-500 inline" />
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </Card>

      {/* Security Vulnerabilities */}
      <Card>
        <div className="p-4">
          <h2 className="text-xl font-semibold mb-4">Security Scan Results</h2>
          {securityScans.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <Shield className="w-12 h-12 mx-auto mb-3" />
              <p>No vulnerabilities detected</p>
            </div>
          ) : (
            <div className="space-y-3">
              {securityScans.map((scan, idx) => (
                <div key={idx} className="border rounded p-4">
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <h4 className="font-medium">{scan.vulnerability}</h4>
                      {scan.cve && (
                        <p className="text-xs text-gray-600">CVE: {scan.cve}</p>
                      )}
                    </div>
                    <span className={`px-2 py-1 text-xs rounded-full ${getSeverityColor(scan.severity)}`}>
                      {scan.severity}
                    </span>
                  </div>
                  <div className="space-y-2 text-sm">
                    <div>
                      <p className="font-medium text-gray-700">Affected Components:</p>
                      <p className="text-gray-600">{scan.affected.join(', ')}</p>
                    </div>
                    <div>
                      <p className="font-medium text-gray-700">Remediation:</p>
                      <p className="text-gray-600">{scan.remediation}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>

      {/* Summary Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <div className="p-4 text-center">
            <CheckCircle className="w-8 h-8 mx-auto mb-2 text-green-500" />
            <p className="text-2xl font-bold">{tests.filter(t => t.status === 'passed').length}</p>
            <p className="text-sm text-gray-600">Tests Passed</p>
          </div>
        </Card>

        <Card>
          <div className="p-4 text-center">
            <XCircle className="w-8 h-8 mx-auto mb-2 text-red-500" />
            <p className="text-2xl font-bold">{tests.filter(t => t.status === 'failed').length}</p>
            <p className="text-sm text-gray-600">Tests Failed</p>
          </div>
        </Card>

        <Card>
          <div className="p-4 text-center">
            <Zap className="w-8 h-8 mx-auto mb-2 text-blue-500" />
            <p className="text-2xl font-bold">
              {benchmarks.filter(b => b.value <= b.target).length}/{benchmarks.length}
            </p>
            <p className="text-sm text-gray-600">Benchmarks Met</p>
          </div>
        </Card>

        <Card>
          <div className="p-4 text-center">
            <Shield className="w-8 h-8 mx-auto mb-2 text-purple-500" />
            <p className="text-2xl font-bold">
              {securityScans.filter(s => s.severity === 'critical' || s.severity === 'high').length}
            </p>
            <p className="text-sm text-gray-600">High-Risk Issues</p>
          </div>
        </Card>
      </div>
    </div>
  )
}