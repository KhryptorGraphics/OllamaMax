export interface FederationCluster {
  id: string;
  name: string;
  region: string;
  status: 'online' | 'offline' | 'degraded' | 'maintenance';
  version: string;
  endpoint: string;
  lastSeen: Date;
  nodes: number;
  activeModels: number;
  health: ClusterHealth;
  metrics: ClusterMetrics;
  capabilities: ClusterCapabilities;
  authentication: AuthenticationConfig;
  networking: NetworkingConfig;
}

export interface ClusterHealth {
  overall: 'healthy' | 'warning' | 'critical';
  cpu: number;
  memory: number;
  disk: number;
  network: number;
  models: number;
  uptime: number;
  lastCheck: Date;
  issues: HealthIssue[];
}

export interface HealthIssue {
  id: string;
  type: 'warning' | 'error' | 'critical';
  message: string;
  component: string;
  timestamp: Date;
  resolved: boolean;
}

export interface ClusterMetrics {
  requestsPerSecond: number;
  responseTime: number;
  errorRate: number;
  throughput: number;
  concurrentUsers: number;
  modelInferences: number;
  networkLatency: number;
  storageUsage: number;
}

export interface ClusterCapabilities {
  maxNodes: number;
  maxModels: number;
  supportedArchitectures: string[];
  features: string[];
  protocols: string[];
  encryption: string[];
  authentication: string[];
}

export interface AuthenticationConfig {
  method: 'mTLS' | 'OAuth2' | 'OIDC' | 'APIKey';
  certificateAuthority: string;
  clientCertificate?: string;
  tokenEndpoint?: string;
  audience?: string;
  scopes?: string[];
}

export interface NetworkingConfig {
  discoveryProtocol: 'DNS' | 'Consul' | 'etcd' | 'Kubernetes';
  loadBalancing: 'round-robin' | 'least-connections' | 'weighted' | 'geographic';
  encryption: 'TLS' | 'mTLS' | 'Wireguard';
  compression: boolean;
  keepAlive: boolean;
  timeout: number;
  retries: number;
}

export interface FederationPolicy {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  priority: number;
  conditions: PolicyCondition[];
  actions: PolicyAction[];
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;
}

export interface PolicyCondition {
  type: 'region' | 'cluster' | 'load' | 'latency' | 'availability';
  operator: 'equals' | 'not-equals' | 'greater-than' | 'less-than' | 'contains';
  value: string | number | boolean;
}

export interface PolicyAction {
  type: 'route' | 'replicate' | 'failover' | 'scale' | 'notify';
  parameters: Record<string, any>;
}

export interface ServiceDiscovery {
  protocol: string;
  endpoint: string;
  services: DiscoveredService[];
  lastUpdate: Date;
  healthy: boolean;
}

export interface DiscoveredService {
  id: string;
  name: string;
  cluster: string;
  endpoint: string;
  port: number;
  protocol: string;
  health: 'healthy' | 'unhealthy' | 'unknown';
  metadata: Record<string, string>;
  tags: string[];
  lastSeen: Date;
}

export interface CrossRegionReplication {
  id: string;
  sourceCluster: string;
  targetClusters: string[];
  models: string[];
  strategy: 'immediate' | 'scheduled' | 'on-demand';
  status: 'active' | 'paused' | 'failed';
  lastSync: Date;
  syncProgress: number;
  conflictResolution: 'last-write-wins' | 'manual' | 'automatic';
}

export interface FederationEvent {
  id: string;
  type: 'cluster-join' | 'cluster-leave' | 'failover' | 'sync' | 'policy-change';
  cluster: string;
  message: string;
  severity: 'info' | 'warning' | 'error';
  timestamp: Date;
  metadata: Record<string, any>;
}

export interface FederationConfiguration {
  name: string;
  description: string;
  clusters: FederationCluster[];
  policies: FederationPolicy[];
  replication: CrossRegionReplication[];
  discovery: ServiceDiscovery;
  monitoring: {
    enabled: boolean;
    interval: number;
    retention: number;
    alerts: AlertConfiguration[];
  };
  security: {
    encryption: boolean;
    authentication: boolean;
    authorization: boolean;
    auditLogging: boolean;
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
}