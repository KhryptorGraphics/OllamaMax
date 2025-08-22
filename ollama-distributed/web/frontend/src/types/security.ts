// Security and compliance types

export interface SecurityState {
  status: SecurityStatus
  threats: ThreatInfo[]
  alerts: SecurityAlert[]
  compliance: ComplianceStatus
  lastScan: string | null
  loading: boolean
  error: string | null
}

export type SecurityStatus = 'secure' | 'warning' | 'critical' | 'unknown'

export interface ThreatInfo {
  id: string
  type: ThreatType
  severity: 'low' | 'medium' | 'high' | 'critical'
  source: string
  target?: string
  description: string
  mitigated: boolean
  timestamp: string
}

export type ThreatType = 
  | 'unauthorized_access'
  | 'brute_force'
  | 'suspicious_activity'
  | 'malware'
  | 'data_breach'
  | 'policy_violation'

export interface SecurityAlert {
  id: string
  type: ThreatType
  severity: 'low' | 'medium' | 'high' | 'critical'
  message: string
  details: Record<string, any>
  acknowledged: boolean
  timestamp: string
}

export interface ComplianceStatus {
  overall: 'compliant' | 'partial' | 'non_compliant'
  frameworks: ComplianceFramework[]
  lastAssessment: string
}

export interface ComplianceFramework {
  name: string
  version: string
  status: 'compliant' | 'partial' | 'non_compliant'
  score: number
  controls: ComplianceControl[]
}

export interface ComplianceControl {
  id: string
  name: string
  status: 'implemented' | 'partial' | 'not_implemented'
  evidence?: string[]
  gaps?: string[]
}