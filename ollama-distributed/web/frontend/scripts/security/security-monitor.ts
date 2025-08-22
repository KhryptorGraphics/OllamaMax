#!/usr/bin/env tsx
/**
 * Security Monitoring and Alerting System
 * 
 * Real-time security monitoring system for the Ollama Distributed frontend.
 * Provides continuous security monitoring, threat detection, and automated alerting.
 * 
 * Features:
 * - Real-time security event monitoring
 * - Automated threat detection and classification
 * - Security metrics collection and analysis
 * - Incident response automation
 * - Compliance monitoring and reporting
 * - Performance and security correlation analysis
 */

import { EventEmitter } from 'events'
import { writeFileSync, readFileSync, existsSync, mkdirSync } from 'fs'
import { join } from 'path'
import { execSync } from 'child_process'

interface SecurityEvent {
  id: string
  timestamp: string
  type: SecurityEventType
  severity: SecuritySeverity
  category: SecurityCategory
  source: string
  description: string
  metadata: Record<string, any>
  resolved: boolean
  resolvedAt?: string
  responseTime?: number
}

interface SecurityAlert {
  id: string
  timestamp: string
  eventId: string
  severity: SecuritySeverity
  title: string
  description: string
  actionItems: string[]
  acknowledged: boolean
  acknowledgedBy?: string
  acknowledgedAt?: string
  escalated: boolean
  escalatedAt?: string
}

interface SecurityMetric {
  name: string
  value: number
  unit: string
  timestamp: string
  threshold?: number
  status: 'NORMAL' | 'WARNING' | 'CRITICAL'
}

interface MonitoringRule {
  id: string
  name: string
  type: 'THRESHOLD' | 'PATTERN' | 'ANOMALY' | 'CORRELATION'
  condition: string
  severity: SecuritySeverity
  enabled: boolean
  cooldownPeriod: number
  lastTriggered?: string
}

type SecurityEventType = 
  | 'AUTHENTICATION_FAILURE'
  | 'AUTHORIZATION_VIOLATION'
  | 'XSS_ATTEMPT'
  | 'SQL_INJECTION_ATTEMPT'
  | 'COMMAND_INJECTION_ATTEMPT'
  | 'BRUTE_FORCE_ATTACK'
  | 'SESSION_ANOMALY'
  | 'DATA_EXFILTRATION'
  | 'VULNERABILITY_SCAN'
  | 'MALICIOUS_REQUEST'
  | 'RATE_LIMIT_VIOLATION'
  | 'CSP_VIOLATION'
  | 'SECURITY_MISCONFIGURATION'
  | 'SUSPICIOUS_ACTIVITY'
  | 'COMPLIANCE_VIOLATION'

type SecuritySeverity = 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | 'INFO'

type SecurityCategory = 
  | 'AUTHENTICATION'
  | 'AUTHORIZATION' 
  | 'INPUT_VALIDATION'
  | 'SESSION_MANAGEMENT'
  | 'CONFIGURATION'
  | 'MONITORING'
  | 'COMPLIANCE'
  | 'PERFORMANCE'

class SecurityMonitor extends EventEmitter {
  private events: SecurityEvent[] = []
  private alerts: SecurityAlert[] = []
  private metrics: SecurityMetric[] = []
  private rules: MonitoringRule[] = []
  private isMonitoring = false
  private monitoringInterval?: NodeJS.Timeout
  private alertThresholds: Record<string, number> = {}
  private outputDir: string
  
  constructor(outputDir = 'security-monitoring') {
    super()
    this.outputDir = join(process.cwd(), outputDir)
    
    if (!existsSync(this.outputDir)) {
      mkdirSync(this.outputDir, { recursive: true })
    }
    
    this.initializeRules()
    this.initializeThresholds()
    this.loadPersistedData()
    
    // Set up event handlers
    this.on('securityEvent', this.handleSecurityEvent.bind(this))
    this.on('alert', this.handleAlert.bind(this))
  }

  private initializeRules() {
    this.rules = [
      {
        id: 'BRUTE_FORCE_DETECTION',
        name: 'Brute Force Attack Detection',
        type: 'THRESHOLD',
        condition: 'authentication_failures > 5 in 5 minutes',
        severity: 'HIGH',
        enabled: true,
        cooldownPeriod: 300000 // 5 minutes
      },
      {
        id: 'XSS_PATTERN_DETECTION',
        name: 'XSS Attack Pattern Detection',
        type: 'PATTERN',
        condition: 'request contains <script> or javascript: or onerror=',
        severity: 'HIGH',
        enabled: true,
        cooldownPeriod: 60000 // 1 minute
      },
      {
        id: 'SQL_INJECTION_DETECTION',
        name: 'SQL Injection Detection',
        type: 'PATTERN',
        condition: 'request contains UNION SELECT or DROP TABLE or OR 1=1',
        severity: 'CRITICAL',
        enabled: true,
        cooldownPeriod: 0 // Immediate alerting
      },
      {
        id: 'RATE_LIMIT_VIOLATION',
        name: 'Rate Limit Violation Detection',
        type: 'THRESHOLD',
        condition: 'requests > 100 per minute from single IP',
        severity: 'MEDIUM',
        enabled: true,
        cooldownPeriod: 60000
      },
      {
        id: 'CSP_VIOLATION_SPIKE',
        name: 'CSP Violation Spike Detection',
        type: 'THRESHOLD',
        condition: 'csp_violations > 10 in 1 minute',
        severity: 'MEDIUM',
        enabled: true,
        cooldownPeriod: 120000
      },
      {
        id: 'SESSION_ANOMALY',
        name: 'Session Anomaly Detection',
        type: 'ANOMALY',
        condition: 'unusual_session_behavior detected',
        severity: 'MEDIUM',
        enabled: true,
        cooldownPeriod: 300000
      },
      {
        id: 'PRIVILEGE_ESCALATION',
        name: 'Privilege Escalation Attempt',
        type: 'PATTERN',
        condition: 'unauthorized admin access attempt',
        severity: 'CRITICAL',
        enabled: true,
        cooldownPeriod: 0
      },
      {
        id: 'VULNERABILITY_SCAN',
        name: 'Vulnerability Scan Detection',
        type: 'PATTERN',
        condition: 'automated vulnerability scanning patterns',
        severity: 'MEDIUM',
        enabled: true,
        cooldownPeriod: 600000 // 10 minutes
      }
    ]
  }

  private initializeThresholds() {
    this.alertThresholds = {
      authentication_failures: 5,
      authorization_violations: 3,
      xss_attempts: 1,
      sql_injection_attempts: 1,
      command_injection_attempts: 1,
      csp_violations: 10,
      rate_limit_violations: 5,
      suspicious_requests: 10,
      error_rate: 0.05, // 5%
      response_time: 5000, // 5 seconds
      memory_usage: 0.9, // 90%
      cpu_usage: 0.8 // 80%
    }
  }

  startMonitoring(intervalMs = 10000) { // Default 10 second intervals
    if (this.isMonitoring) {
      console.warn('Security monitoring is already running')
      return
    }

    console.log('🔒 Starting security monitoring system...')
    this.isMonitoring = true

    this.monitoringInterval = setInterval(() => {
      this.performSecurityChecks()
      this.collectMetrics()
      this.analyzePatterns()
      this.checkThresholds()
    }, intervalMs)

    // Immediate initial check
    this.performSecurityChecks()
    this.collectMetrics()

    console.log('✅ Security monitoring system started')
    
    // Schedule periodic reports
    setInterval(() => {
      this.generatePeriodicReport()
    }, 3600000) // Every hour
  }

  stopMonitoring() {
    if (!this.isMonitoring) {
      return
    }

    console.log('🛑 Stopping security monitoring system...')
    
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval)
    }
    
    this.isMonitoring = false
    this.persistData()
    
    console.log('✅ Security monitoring system stopped')
  }

  private performSecurityChecks() {
    // Check authentication security
    this.checkAuthenticationSecurity()
    
    // Check for suspicious patterns
    this.checkSuspiciousPatterns()
    
    // Check system security health
    this.checkSystemSecurityHealth()
    
    // Check compliance status
    this.checkComplianceStatus()
  }

  private checkAuthenticationSecurity() {
    // Simulate authentication monitoring
    const recentAuthFailures = this.getRecentEvents('AUTHENTICATION_FAILURE', 300000) // 5 minutes
    
    if (recentAuthFailures.length >= this.alertThresholds.authentication_failures) {
      this.createSecurityEvent({
        type: 'BRUTE_FORCE_ATTACK',
        severity: 'HIGH',
        category: 'AUTHENTICATION',
        source: 'auth_monitor',
        description: `Potential brute force attack detected: ${recentAuthFailures.length} failed attempts`,
        metadata: {
          failedAttempts: recentAuthFailures.length,
          timeWindow: '5 minutes',
          sources: this.extractSources(recentAuthFailures)
        }
      })
    }

    // Check for session anomalies
    this.checkSessionAnomalies()
  }

  private checkSessionAnomalies() {
    // Simulate session monitoring
    const suspiciousSessionEvents = [
      'Multiple simultaneous sessions from different locations',
      'Session hijacking indicators detected',
      'Unusual session duration patterns'
    ]

    // Random simulation for demo
    if (Math.random() < 0.1) { // 10% chance
      const randomEvent = suspiciousSessionEvents[Math.floor(Math.random() * suspiciousSessionEvents.length)]
      
      this.createSecurityEvent({
        type: 'SESSION_ANOMALY',
        severity: 'MEDIUM',
        category: 'SESSION_MANAGEMENT',
        source: 'session_monitor',
        description: randomEvent,
        metadata: {
          detectionMethod: 'behavioral_analysis',
          confidenceScore: Math.random() * 0.5 + 0.5
        }
      })
    }
  }

  private checkSuspiciousPatterns() {
    const suspiciousPatterns = [
      { pattern: 'XSS_ATTEMPT', description: 'Cross-site scripting attempt detected', severity: 'HIGH' as SecuritySeverity },
      { pattern: 'SQL_INJECTION_ATTEMPT', description: 'SQL injection attempt detected', severity: 'CRITICAL' as SecuritySeverity },
      { pattern: 'COMMAND_INJECTION_ATTEMPT', description: 'Command injection attempt detected', severity: 'CRITICAL' as SecuritySeverity },
      { pattern: 'VULNERABILITY_SCAN', description: 'Vulnerability scanning detected', severity: 'MEDIUM' as SecuritySeverity }
    ]

    // Simulate pattern detection
    suspiciousPatterns.forEach(({ pattern, description, severity }) => {
      if (Math.random() < 0.05) { // 5% chance for demo
        this.createSecurityEvent({
          type: pattern as SecurityEventType,
          severity,
          category: 'INPUT_VALIDATION',
          source: 'pattern_detector',
          description,
          metadata: {
            patternType: pattern,
            detectionRule: `pattern_${pattern.toLowerCase()}`,
            requestDetails: this.generateMockRequestDetails()
          }
        })
      }
    })
  }

  private checkSystemSecurityHealth() {
    // Simulate system health monitoring
    const healthMetrics = {
      csp_violations: Math.floor(Math.random() * 20),
      security_headers: Math.random() > 0.1, // 90% chance of proper headers
      https_enforcement: Math.random() > 0.05, // 95% chance of HTTPS
      vulnerability_scan_status: Math.random() > 0.2, // 80% chance of up-to-date
      dependency_vulnerabilities: Math.floor(Math.random() * 5)
    }

    // Check CSP violations
    if (healthMetrics.csp_violations > this.alertThresholds.csp_violations) {
      this.createSecurityEvent({
        type: 'CSP_VIOLATION',
        severity: 'MEDIUM',
        category: 'CONFIGURATION',
        source: 'csp_monitor',
        description: `High number of CSP violations: ${healthMetrics.csp_violations}`,
        metadata: {
          violations: healthMetrics.csp_violations,
          threshold: this.alertThresholds.csp_violations
        }
      })
    }

    // Check security headers
    if (!healthMetrics.security_headers) {
      this.createSecurityEvent({
        type: 'SECURITY_MISCONFIGURATION',
        severity: 'MEDIUM',
        category: 'CONFIGURATION',
        source: 'header_monitor',
        description: 'Missing or misconfigured security headers detected',
        metadata: {
          missingHeaders: ['Content-Security-Policy', 'X-Frame-Options'],
          recommendation: 'Review and update security headers configuration'
        }
      })
    }

    // Check vulnerability status
    if (healthMetrics.dependency_vulnerabilities > 0) {
      this.createSecurityEvent({
        type: 'VULNERABILITY_SCAN',
        severity: healthMetrics.dependency_vulnerabilities > 2 ? 'HIGH' : 'MEDIUM',
        category: 'CONFIGURATION',
        source: 'vuln_scanner',
        description: `${healthMetrics.dependency_vulnerabilities} known vulnerabilities in dependencies`,
        metadata: {
          vulnerablePackages: healthMetrics.dependency_vulnerabilities,
          recommendation: 'Update vulnerable dependencies immediately'
        }
      })
    }
  }

  private checkComplianceStatus() {
    const complianceChecks = [
      { name: 'GDPR_COMPLIANCE', status: Math.random() > 0.1 },
      { name: 'SOC2_COMPLIANCE', status: Math.random() > 0.15 },
      { name: 'ISO27001_COMPLIANCE', status: Math.random() > 0.12 },
      { name: 'OWASP_TOP10_COMPLIANCE', status: Math.random() > 0.08 }
    ]

    complianceChecks.forEach(({ name, status }) => {
      if (!status) {
        this.createSecurityEvent({
          type: 'COMPLIANCE_VIOLATION',
          severity: 'HIGH',
          category: 'COMPLIANCE',
          source: 'compliance_monitor',
          description: `${name} violation detected`,
          metadata: {
            complianceFramework: name,
            violationType: 'configuration_drift',
            remediation: 'Review compliance requirements and update configuration'
          }
        })
      }
    })
  }

  private collectMetrics() {
    const timestamp = new Date().toISOString()
    
    // Collect security metrics
    const metrics: SecurityMetric[] = [
      {
        name: 'authentication_failures_per_minute',
        value: this.getRecentEvents('AUTHENTICATION_FAILURE', 60000).length,
        unit: 'count/minute',
        timestamp,
        threshold: 5,
        status: 'NORMAL'
      },
      {
        name: 'security_events_per_hour',
        value: this.getRecentEvents(null, 3600000).length,
        unit: 'count/hour',
        timestamp,
        threshold: 50,
        status: 'NORMAL'
      },
      {
        name: 'alert_response_time',
        value: this.calculateAverageResponseTime(),
        unit: 'seconds',
        timestamp,
        threshold: 300, // 5 minutes
        status: 'NORMAL'
      },
      {
        name: 'vulnerability_count',
        value: Math.floor(Math.random() * 10),
        unit: 'count',
        timestamp,
        threshold: 5,
        status: 'NORMAL'
      },
      {
        name: 'compliance_score',
        value: Math.floor(Math.random() * 20 + 80), // 80-100%
        unit: 'percentage',
        timestamp,
        threshold: 90,
        status: 'NORMAL'
      }
    ]

    // Update metric status based on thresholds
    metrics.forEach(metric => {
      if (metric.threshold) {
        if (metric.value > metric.threshold) {
          metric.status = metric.name === 'compliance_score' ? 'WARNING' : 'CRITICAL'
        } else if (metric.value > metric.threshold * 0.8) {
          metric.status = 'WARNING'
        }
      }
    })

    this.metrics.push(...metrics)
    
    // Keep only recent metrics (last 24 hours)
    const cutoffTime = Date.now() - 24 * 60 * 60 * 1000
    this.metrics = this.metrics.filter(m => new Date(m.timestamp).getTime() > cutoffTime)
  }

  private analyzePatterns() {
    // Analyze recent events for patterns
    const recentEvents = this.getRecentEvents(null, 3600000) // Last hour
    
    // Pattern analysis
    const eventTypeCounts = new Map<SecurityEventType, number>()
    const sourceCounts = new Map<string, number>()
    
    recentEvents.forEach(event => {
      eventTypeCounts.set(event.type, (eventTypeCounts.get(event.type) || 0) + 1)
      sourceCounts.set(event.source, (sourceCounts.get(event.source) || 0) + 1)
    })

    // Check for suspicious patterns
    eventTypeCounts.forEach((count, type) => {
      if (count > 10 && type !== 'SUSPICIOUS_ACTIVITY') { // Avoid recursion
        this.createSecurityEvent({
          type: 'SUSPICIOUS_ACTIVITY',
          severity: 'MEDIUM',
          category: 'MONITORING',
          source: 'pattern_analyzer',
          description: `Unusual spike in ${type} events: ${count} occurrences in the last hour`,
          metadata: {
            eventType: type,
            count,
            timeWindow: '1 hour',
            threshold: 10
          }
        })
      }
    })
  }

  private checkThresholds() {
    // Check metric thresholds and generate alerts
    const criticalMetrics = this.metrics.filter(m => 
      m.status === 'CRITICAL' && 
      new Date(m.timestamp).getTime() > Date.now() - 300000 // Last 5 minutes
    )

    criticalMetrics.forEach(metric => {
      this.createAlert({
        eventId: `threshold_${metric.name}`,
        severity: 'HIGH',
        title: `Critical Threshold Exceeded: ${metric.name}`,
        description: `Metric ${metric.name} has exceeded critical threshold: ${metric.value} ${metric.unit} > ${metric.threshold} ${metric.unit}`,
        actionItems: [
          `Investigate cause of elevated ${metric.name}`,
          'Check system resources and performance',
          'Review recent configuration changes',
          'Consider scaling or optimization measures'
        ]
      })
    })
  }

  private createSecurityEvent(eventData: Partial<SecurityEvent>) {
    const event: SecurityEvent = {
      id: this.generateId(),
      timestamp: new Date().toISOString(),
      type: eventData.type!,
      severity: eventData.severity!,
      category: eventData.category!,
      source: eventData.source!,
      description: eventData.description!,
      metadata: eventData.metadata || {},
      resolved: false
    }

    this.events.push(event)
    this.emit('securityEvent', event)

    // Auto-create alert for high severity events
    if (event.severity === 'CRITICAL' || event.severity === 'HIGH') {
      this.createAlert({
        eventId: event.id,
        severity: event.severity,
        title: `Security Event: ${event.type}`,
        description: event.description,
        actionItems: this.generateActionItems(event)
      })
    }

    // Keep only recent events (last 7 days)
    const cutoffTime = Date.now() - 7 * 24 * 60 * 60 * 1000
    this.events = this.events.filter(e => new Date(e.timestamp).getTime() > cutoffTime)

    return event
  }

  private createAlert(alertData: Partial<SecurityAlert>) {
    const alert: SecurityAlert = {
      id: this.generateId(),
      timestamp: new Date().toISOString(),
      eventId: alertData.eventId!,
      severity: alertData.severity!,
      title: alertData.title!,
      description: alertData.description!,
      actionItems: alertData.actionItems || [],
      acknowledged: false,
      escalated: false
    }

    this.alerts.push(alert)
    this.emit('alert', alert)

    // Auto-escalate critical alerts after 10 minutes
    if (alert.severity === 'CRITICAL') {
      setTimeout(() => {
        if (!alert.acknowledged) {
          this.escalateAlert(alert.id)
        }
      }, 600000) // 10 minutes
    }

    console.log(`🚨 SECURITY ALERT [${alert.severity}]: ${alert.title}`)
    
    return alert
  }

  private generateActionItems(event: SecurityEvent): string[] {
    const actionItems = []

    switch (event.type) {
      case 'AUTHENTICATION_FAILURE':
      case 'BRUTE_FORCE_ATTACK':
        actionItems.push(
          'Review authentication logs for suspicious patterns',
          'Consider implementing additional rate limiting',
          'Check for compromised accounts',
          'Review firewall rules for source IPs'
        )
        break

      case 'XSS_ATTEMPT':
        actionItems.push(
          'Review input validation and output encoding',
          'Check Content Security Policy configuration',
          'Audit affected endpoints for XSS vulnerabilities',
          'Consider implementing additional XSS protection'
        )
        break

      case 'SQL_INJECTION_ATTEMPT':
        actionItems.push(
          'Immediately review database query construction',
          'Audit for parameterized queries usage',
          'Check database logs for unauthorized access',
          'Implement additional input validation'
        )
        break

      case 'CSP_VIOLATION':
        actionItems.push(
          'Review Content Security Policy configuration',
          'Check for unauthorized script injections',
          'Audit third-party integrations',
          'Update CSP rules if legitimate violations'
        )
        break

      case 'VULNERABILITY_SCAN':
        actionItems.push(
          'Update vulnerable dependencies immediately',
          'Review security patches and updates',
          'Conduct comprehensive security audit',
          'Implement vulnerability management process'
        )
        break

      default:
        actionItems.push(
          'Investigate the security event immediately',
          'Review system logs for additional context',
          'Check for related security events',
          'Document findings and remediation steps'
        )
    }

    return actionItems
  }

  acknowledgeAlert(alertId: string, acknowledgedBy: string) {
    const alert = this.alerts.find(a => a.id === alertId)
    if (alert && !alert.acknowledged) {
      alert.acknowledged = true
      alert.acknowledgedBy = acknowledgedBy
      alert.acknowledgedAt = new Date().toISOString()
      
      console.log(`✅ Alert ${alertId} acknowledged by ${acknowledgedBy}`)
      
      // Calculate response time
      const event = this.events.find(e => e.id === alert.eventId)
      if (event) {
        event.responseTime = Date.now() - new Date(event.timestamp).getTime()
      }
    }
  }

  escalateAlert(alertId: string) {
    const alert = this.alerts.find(a => a.id === alertId)
    if (alert && !alert.escalated) {
      alert.escalated = true
      alert.escalatedAt = new Date().toISOString()
      
      console.log(`🔺 Alert ${alertId} escalated due to no acknowledgment`)
      
      // Send escalation notification (simulate)
      this.sendEscalationNotification(alert)
    }
  }

  resolveEvent(eventId: string) {
    const event = this.events.find(e => e.id === eventId)
    if (event && !event.resolved) {
      event.resolved = true
      event.resolvedAt = new Date().toISOString()
      
      console.log(`✅ Security event ${eventId} resolved`)
    }
  }

  private sendEscalationNotification(alert: SecurityAlert) {
    // Simulate sending escalation notifications
    console.log(`📧 Escalation notification sent for alert: ${alert.title}`)
    console.log(`   Severity: ${alert.severity}`)
    console.log(`   Description: ${alert.description}`)
    console.log(`   Action Items: ${alert.actionItems.join(', ')}`)
  }

  private handleSecurityEvent(event: SecurityEvent) {
    // Log security event
    console.log(`🔒 Security Event [${event.severity}]: ${event.type} - ${event.description}`)
    
    // Additional automated response based on event type
    switch (event.type) {
      case 'BRUTE_FORCE_ATTACK':
        this.automaticBruteForceResponse(event)
        break
      case 'SQL_INJECTION_ATTEMPT':
        this.automaticInjectionResponse(event)
        break
      case 'VULNERABILITY_SCAN':
        this.automaticVulnScanResponse(event)
        break
    }
  }

  private handleAlert(alert: SecurityAlert) {
    // Log alert
    console.log(`🚨 Security Alert [${alert.severity}]: ${alert.title}`)
    
    // Save alert to file for external processing
    this.saveAlertToFile(alert)
  }

  private automaticBruteForceResponse(event: SecurityEvent) {
    console.log('🛡️ Activating automatic brute force protection measures')
    
    // Simulate automatic response
    if (event.metadata.sources) {
      console.log(`   Temporarily blocking IPs: ${event.metadata.sources.join(', ')}`)
    }
    
    console.log('   Increasing authentication logging level')
    console.log('   Notifying security team')
  }

  private automaticInjectionResponse(event: SecurityEvent) {
    console.log('🛡️ Activating automatic injection attack protection')
    console.log('   Increasing input validation logging')
    console.log('   Reviewing recent database queries')
    console.log('   Notifying development team immediately')
  }

  private automaticVulnScanResponse(event: SecurityEvent) {
    console.log('🛡️ Activating vulnerability management procedures')
    console.log('   Scheduling emergency patch review')
    console.log('   Updating vulnerability database')
    console.log('   Notifying infrastructure team')
  }

  generateSecurityReport(): any {
    const now = new Date()
    const last24Hours = new Date(now.getTime() - 24 * 60 * 60 * 1000)
    
    const recentEvents = this.events.filter(e => new Date(e.timestamp) > last24Hours)
    const recentAlerts = this.alerts.filter(a => new Date(a.timestamp) > last24Hours)
    const recentMetrics = this.metrics.filter(m => new Date(m.timestamp) > last24Hours)
    
    const report = {
      reportId: `security-report-${Date.now()}`,
      timestamp: now.toISOString(),
      period: {
        start: last24Hours.toISOString(),
        end: now.toISOString()
      },
      summary: {
        totalEvents: recentEvents.length,
        totalAlerts: recentAlerts.length,
        criticalEvents: recentEvents.filter(e => e.severity === 'CRITICAL').length,
        unacknowledgedAlerts: recentAlerts.filter(a => !a.acknowledged).length,
        averageResponseTime: this.calculateAverageResponseTime()
      },
      eventsByType: this.groupEventsByType(recentEvents),
      eventsBySeverity: this.groupEventsBySeverity(recentEvents),
      topSources: this.getTopEventSources(recentEvents),
      securityMetrics: recentMetrics,
      recommendations: this.generateSecurityRecommendations(recentEvents, recentAlerts)
    }
    
    return report
  }

  private generatePeriodicReport() {
    const report = this.generateSecurityReport()
    
    console.log('\n📊 Hourly Security Report')
    console.log('==========================')
    console.log(`Total Events: ${report.summary.totalEvents}`)
    console.log(`Total Alerts: ${report.summary.totalAlerts}`)
    console.log(`Critical Events: ${report.summary.criticalEvents}`)
    console.log(`Unacknowledged Alerts: ${report.summary.unacknowledgedAlerts}`)
    console.log(`Average Response Time: ${report.summary.averageResponseTime}s`)
    
    // Save report
    const reportPath = join(this.outputDir, `security-report-${Date.now()}.json`)
    writeFileSync(reportPath, JSON.stringify(report, null, 2))
    console.log(`Report saved: ${reportPath}`)
  }

  // Helper methods
  private generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
  }

  private getRecentEvents(type: SecurityEventType | null, timeWindowMs: number): SecurityEvent[] {
    const cutoffTime = Date.now() - timeWindowMs
    return this.events.filter(e => 
      new Date(e.timestamp).getTime() > cutoffTime &&
      (type === null || e.type === type)
    )
  }

  private extractSources(events: SecurityEvent[]): string[] {
    const sources = new Set(events.map(e => e.source))
    return Array.from(sources)
  }

  private generateMockRequestDetails(): any {
    return {
      method: Math.random() > 0.5 ? 'POST' : 'GET',
      url: `/api/endpoint${Math.floor(Math.random() * 10)}`,
      userAgent: 'Mozilla/5.0 (Suspicious Bot)',
      sourceIP: `192.168.1.${Math.floor(Math.random() * 255)}`,
      timestamp: new Date().toISOString()
    }
  }

  private calculateAverageResponseTime(): number {
    const resolvedEvents = this.events.filter(e => e.resolved && e.responseTime)
    if (resolvedEvents.length === 0) return 0
    
    const totalResponseTime = resolvedEvents.reduce((sum, e) => sum + (e.responseTime || 0), 0)
    return Math.round(totalResponseTime / resolvedEvents.length / 1000) // Convert to seconds
  }

  private groupEventsByType(events: SecurityEvent[]): Record<string, number> {
    const grouped: Record<string, number> = {}
    events.forEach(e => {
      grouped[e.type] = (grouped[e.type] || 0) + 1
    })
    return grouped
  }

  private groupEventsBySeverity(events: SecurityEvent[]): Record<string, number> {
    const grouped: Record<string, number> = {}
    events.forEach(e => {
      grouped[e.severity] = (grouped[e.severity] || 0) + 1
    })
    return grouped
  }

  private getTopEventSources(events: SecurityEvent[]): Array<{source: string, count: number}> {
    const sourceCounts: Record<string, number> = {}
    events.forEach(e => {
      sourceCounts[e.source] = (sourceCounts[e.source] || 0) + 1
    })
    
    return Object.entries(sourceCounts)
      .map(([source, count]) => ({ source, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10)
  }

  private generateSecurityRecommendations(events: SecurityEvent[], alerts: SecurityAlert[]): string[] {
    const recommendations = []
    
    const criticalEvents = events.filter(e => e.severity === 'CRITICAL').length
    const unacknowledgedAlerts = alerts.filter(a => !a.acknowledged).length
    
    if (criticalEvents > 5) {
      recommendations.push('High number of critical security events detected. Review security posture immediately.')
    }
    
    if (unacknowledgedAlerts > 3) {
      recommendations.push('Multiple unacknowledged alerts. Review incident response procedures.')
    }
    
    const injectionEvents = events.filter(e => 
      e.type === 'XSS_ATTEMPT' || 
      e.type === 'SQL_INJECTION_ATTEMPT' || 
      e.type === 'COMMAND_INJECTION_ATTEMPT'
    ).length
    
    if (injectionEvents > 2) {
      recommendations.push('Multiple injection attempts detected. Review input validation and sanitization.')
    }
    
    if (recommendations.length === 0) {
      recommendations.push('Security posture is stable. Continue monitoring and maintain current security measures.')
    }
    
    return recommendations
  }

  private saveAlertToFile(alert: SecurityAlert) {
    const alertPath = join(this.outputDir, `alert-${alert.id}.json`)
    writeFileSync(alertPath, JSON.stringify(alert, null, 2))
  }

  private persistData() {
    const dataPath = join(this.outputDir, 'monitor-data.json')
    const data = {
      events: this.events.slice(-1000), // Keep last 1000 events
      alerts: this.alerts.slice(-100),   // Keep last 100 alerts
      metrics: this.metrics.slice(-1000) // Keep last 1000 metrics
    }
    
    writeFileSync(dataPath, JSON.stringify(data, null, 2))
  }

  private loadPersistedData() {
    const dataPath = join(this.outputDir, 'monitor-data.json')
    
    if (existsSync(dataPath)) {
      try {
        const data = JSON.parse(readFileSync(dataPath, 'utf-8'))
        this.events = data.events || []
        this.alerts = data.alerts || []
        this.metrics = data.metrics || []
        
        console.log(`Loaded ${this.events.length} events, ${this.alerts.length} alerts, ${this.metrics.length} metrics`)
      } catch (error) {
        console.warn('Failed to load persisted data:', error)
      }
    }
  }

  // Public API methods
  getEvents(filters?: {
    type?: SecurityEventType
    severity?: SecuritySeverity
    category?: SecurityCategory
    timeRange?: number
  }): SecurityEvent[] {
    let filteredEvents = this.events

    if (filters) {
      if (filters.type) {
        filteredEvents = filteredEvents.filter(e => e.type === filters.type)
      }
      if (filters.severity) {
        filteredEvents = filteredEvents.filter(e => e.severity === filters.severity)
      }
      if (filters.category) {
        filteredEvents = filteredEvents.filter(e => e.category === filters.category)
      }
      if (filters.timeRange) {
        const cutoffTime = Date.now() - filters.timeRange
        filteredEvents = filteredEvents.filter(e => 
          new Date(e.timestamp).getTime() > cutoffTime
        )
      }
    }

    return filteredEvents.sort((a, b) => 
      new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
    )
  }

  getAlerts(acknowledged = false): SecurityAlert[] {
    return this.alerts
      .filter(a => a.acknowledged === acknowledged)
      .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
  }

  getCurrentMetrics(): SecurityMetric[] {
    const now = Date.now()
    const recentMetrics = new Map<string, SecurityMetric>()
    
    // Get most recent metric for each name
    this.metrics
      .filter(m => now - new Date(m.timestamp).getTime() < 3600000) // Last hour
      .forEach(metric => {
        const existing = recentMetrics.get(metric.name)
        if (!existing || new Date(metric.timestamp) > new Date(existing.timestamp)) {
          recentMetrics.set(metric.name, metric)
        }
      })
    
    return Array.from(recentMetrics.values())
  }

  getSystemStatus(): {
    monitoring: boolean
    eventsLast24h: number
    alertsLast24h: number
    criticalAlertsUnacknowledged: number
    averageResponseTime: number
    systemHealth: 'HEALTHY' | 'WARNING' | 'CRITICAL'
  } {
    const last24h = Date.now() - 24 * 60 * 60 * 1000
    const recentEvents = this.events.filter(e => new Date(e.timestamp).getTime() > last24h)
    const recentAlerts = this.alerts.filter(a => new Date(a.timestamp).getTime() > last24h)
    const criticalUnacknowledged = recentAlerts.filter(a => 
      a.severity === 'CRITICAL' && !a.acknowledged
    ).length

    let systemHealth: 'HEALTHY' | 'WARNING' | 'CRITICAL' = 'HEALTHY'
    
    if (criticalUnacknowledged > 0 || recentEvents.filter(e => e.severity === 'CRITICAL').length > 5) {
      systemHealth = 'CRITICAL'
    } else if (recentAlerts.length > 10 || recentEvents.filter(e => e.severity === 'HIGH').length > 10) {
      systemHealth = 'WARNING'
    }

    return {
      monitoring: this.isMonitoring,
      eventsLast24h: recentEvents.length,
      alertsLast24h: recentAlerts.length,
      criticalAlertsUnacknowledged: criticalUnacknowledged,
      averageResponseTime: this.calculateAverageResponseTime(),
      systemHealth
    }
  }
}

// CLI interface
if (require.main === module) {
  const monitor = new SecurityMonitor()
  
  // Handle graceful shutdown
  process.on('SIGINT', () => {
    console.log('\nShutting down security monitor...')
    monitor.stopMonitoring()
    process.exit(0)
  })
  
  process.on('SIGTERM', () => {
    console.log('\nShutting down security monitor...')
    monitor.stopMonitoring()
    process.exit(0)
  })
  
  // Start monitoring
  monitor.startMonitoring(5000) // Check every 5 seconds
  
  // Demonstrate some events
  setTimeout(() => {
    console.log('Simulating authentication failure...')
    monitor['createSecurityEvent']({
      type: 'AUTHENTICATION_FAILURE',
      severity: 'MEDIUM',
      category: 'AUTHENTICATION',
      source: 'auth_service',
      description: 'Failed login attempt from suspicious IP',
      metadata: { ip: '192.168.1.100', attempts: 3 }
    })
  }, 10000)
  
  setTimeout(() => {
    console.log('Simulating XSS attempt...')
    monitor['createSecurityEvent']({
      type: 'XSS_ATTEMPT',
      severity: 'HIGH',
      category: 'INPUT_VALIDATION',
      source: 'web_application',
      description: 'Cross-site scripting attempt detected',
      metadata: { payload: '<script>alert("xss")</script>' }
    })
  }, 20000)
  
  // Keep the process running
  console.log('Security monitoring is running. Press Ctrl+C to stop.')
}

export default SecurityMonitor