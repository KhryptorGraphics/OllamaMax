export interface Identity {
  id: string;
  type: 'user' | 'service' | 'device' | 'application';
  name: string;
  attributes: Record<string, any>;
  certificates: Certificate[];
  policies: string[];
  roles: string[];
  groups: string[];
  status: 'active' | 'inactive' | 'suspended' | 'revoked';
  lastSeen: Date;
  createdAt: Date;
  updatedAt: Date;
  metadata: Record<string, any>;
}

export interface Certificate {
  id: string;
  type: 'client' | 'server' | 'ca' | 'intermediate';
  subject: string;
  issuer: string;
  serialNumber: string;
  fingerprint: string;
  notBefore: Date;
  notAfter: Date;
  keyUsage: string[];
  extendedKeyUsage: string[];
  subjectAltNames: string[];
  status: 'valid' | 'expired' | 'revoked' | 'pending';
  pemData: string;
  privateKey?: string;
  chain: string[];
  ocspUrl?: string;
  crlUrl?: string;
}

export interface NetworkPolicy {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  priority: number;
  source: PolicySelector;
  destination: PolicySelector;
  action: 'allow' | 'deny' | 'log' | 'monitor';
  conditions: PolicyCondition[];
  protocols: Protocol[];
  ports: PortRange[];
  timeRestrictions?: TimeRestriction[];
  rateLimit?: RateLimit;
  encryption: EncryptionRequirement;
  logging: LoggingConfig;
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;
}

export interface PolicySelector {
  identities: string[];
  labels: Record<string, string>;
  namespaces: string[];
  ipRanges: string[];
  domains: string[];
}

export interface PolicyCondition {
  type: 'identity' | 'location' | 'time' | 'device' | 'application' | 'risk';
  operator: 'equals' | 'not-equals' | 'in' | 'not-in' | 'contains' | 'matches';
  value: string | string[];
  metadata?: Record<string, any>;
}

export interface Protocol {
  name: 'TCP' | 'UDP' | 'ICMP' | 'HTTP' | 'HTTPS' | 'gRPC' | 'WebSocket';
  version?: string;
  options?: Record<string, any>;
}

export interface PortRange {
  start: number;
  end: number;
}

export interface TimeRestriction {
  days: number[]; // 0-6 (Sunday to Saturday)
  startTime: string; // HH:MM format
  endTime: string; // HH:MM format
  timezone: string;
}

export interface RateLimit {
  requests: number;
  period: string; // e.g., "1m", "1h", "1d"
  burst?: number;
}

export interface EncryptionRequirement {
  required: boolean;
  minTlsVersion: '1.0' | '1.1' | '1.2' | '1.3';
  cipherSuites: string[];
  certificateValidation: 'strict' | 'relaxed' | 'disabled';
  mutualTls: boolean;
}

export interface LoggingConfig {
  enabled: boolean;
  level: 'info' | 'warning' | 'error' | 'debug';
  destination: 'syslog' | 'file' | 'elasticsearch' | 'datadog';
  format: 'json' | 'text';
  includePayload: boolean;
}

export interface SecureChannel {
  id: string;
  name: string;
  type: 'mTLS' | 'IPSec' | 'WireGuard' | 'QUIC';
  source: Endpoint;
  destination: Endpoint;
  status: 'active' | 'inactive' | 'establishing' | 'failed';
  encryption: ChannelEncryption;
  authentication: ChannelAuthentication;
  metrics: ChannelMetrics;
  configuration: ChannelConfiguration;
  lastHandshake: Date;
  createdAt: Date;
}

export interface Endpoint {
  identity: string;
  address: string;
  port: number;
  protocol: string;
  certificate?: string;
}

export interface ChannelEncryption {
  algorithm: string;
  keySize: number;
  mode: string;
  keyExchange: string;
  perfectForwardSecrecy: boolean;
}

export interface ChannelAuthentication {
  method: 'certificate' | 'psk' | 'token';
  certificate?: string;
  presharedKey?: string;
  token?: string;
  verified: boolean;
}

export interface ChannelMetrics {
  bytesTransmitted: number;
  bytesReceived: number;
  packetsTransmitted: number;
  packetsReceived: number;
  errors: number;
  latency: number;
  bandwidth: number;
  lastUpdate: Date;
}

export interface ChannelConfiguration {
  keepAlive: boolean;
  keepAliveInterval: number;
  timeout: number;
  retries: number;
  compression: boolean;
  mtu: number;
  bufferSize: number;
}

export interface TrustScore {
  identity: string;
  score: number; // 0-100
  factors: TrustFactor[];
  lastCalculated: Date;
  trend: 'increasing' | 'decreasing' | 'stable';
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
}

export interface TrustFactor {
  type: 'location' | 'device' | 'behavior' | 'time' | 'application' | 'network';
  name: string;
  value: number;
  weight: number;
  impact: 'positive' | 'negative' | 'neutral';
  description: string;
}

export interface SecurityEvent {
  id: string;
  type: 'authentication' | 'authorization' | 'policy-violation' | 'anomaly' | 'threat';
  severity: 'info' | 'warning' | 'error' | 'critical';
  identity: string;
  source: string;
  destination?: string;
  action: string;
  outcome: 'allowed' | 'denied' | 'logged';
  message: string;
  details: Record<string, any>;
  timestamp: Date;
  riskScore: number;
  resolved: boolean;
  resolvedAt?: Date;
  resolvedBy?: string;
}

export interface CertificateAuthority {
  id: string;
  name: string;
  type: 'root' | 'intermediate' | 'external';
  certificate: Certificate;
  privateKey?: string;
  status: 'active' | 'inactive' | 'revoked';
  validCertificates: number;
  expiredCertificates: number;
  revokedCertificates: number;
  nextSerial: string;
  crlDistributionPoint?: string;
  ocspResponder?: string;
  issuingPolicy: IssuingPolicy;
  createdAt: Date;
  updatedAt: Date;
}

export interface IssuingPolicy {
  maxValidityDays: number;
  keyUsage: string[];
  extendedKeyUsage: string[];
  basicConstraints: {
    ca: boolean;
    pathLength?: number;
  };
  subjectAltNameRequired: boolean;
  keyMinimumSize: number;
  autoRenewalDays: number;
}

export interface IdentityProvider {
  id: string;
  name: string;
  type: 'OIDC' | 'SAML' | 'LDAP' | 'OAuth2' | 'Certificate';
  enabled: boolean;
  configuration: ProviderConfiguration;
  mappings: AttributeMapping[];
  status: 'healthy' | 'degraded' | 'offline';
  lastSync: Date;
  syncedIdentities: number;
  errors: string[];
}

export interface ProviderConfiguration {
  endpoint: string;
  clientId?: string;
  clientSecret?: string;
  scopes?: string[];
  claims?: string[];
  certificate?: string;
  bindDn?: string;
  bindPassword?: string;
  baseDn?: string;
  filter?: string;
  attributes?: string[];
}

export interface AttributeMapping {
  source: string;
  target: string;
  required: boolean;
  defaultValue?: string;
  transformation?: string;
}

export interface ZeroTrustConfiguration {
  name: string;
  description: string;
  policies: NetworkPolicy[];
  identities: Identity[];
  certificateAuthorities: CertificateAuthority[];
  identityProviders: IdentityProvider[];
  channels: SecureChannel[];
  monitoring: {
    enabled: boolean;
    retention: number;
    alerting: AlertConfiguration[];
    anomalyDetection: boolean;
    behaviorAnalytics: boolean;
  };
  compliance: {
    frameworks: string[];
    auditLogging: boolean;
    reportGeneration: boolean;
    dataRetention: number;
  };
}

export interface AlertConfiguration {
  id: string;
  name: string;
  condition: string;
  threshold: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
  enabled: boolean;
  channels: string[];
  suppressionRules: SuppressionRule[];
}

export interface SuppressionRule {
  condition: string;
  duration: number;
  maxOccurrences: number;
}