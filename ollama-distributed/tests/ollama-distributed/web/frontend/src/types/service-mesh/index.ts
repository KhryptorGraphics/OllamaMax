export interface ServiceMeshConfiguration {
  id: string;
  name: string;
  type: 'istio' | 'linkerd' | 'consul-connect' | 'envoy' | 'custom';
  version: string;
  namespace: string;
  status: 'active' | 'inactive' | 'degraded' | 'failed';
  controlPlane: ControlPlaneConfig;
  dataPlane: DataPlaneConfig;
  security: SecurityConfig;
  observability: ObservabilityConfig;
  traffic: TrafficConfig;
  createdAt: Date;
  updatedAt: Date;
}

export interface ControlPlaneConfig {
  replicas: number;
  resources: ResourceRequirements;
  highAvailability: boolean;
  autoInjection: boolean;
  defaultPolicies: string[];
  telemetryCollection: boolean;
  metrics: ControlPlaneMetrics;
}

export interface DataPlaneConfig {
  proxyType: 'envoy' | 'linkerd-proxy' | 'consul-connect' | 'custom';
  proxyVersion: string;
  autoInjection: AutoInjectionConfig;
  resources: ResourceRequirements;
  configuration: ProxyConfiguration;
  updateStrategy: UpdateStrategy;
}

export interface AutoInjectionConfig {
  enabled: boolean;
  namespaces: string[];
  excludedNamespaces: string[];
  labels: Record<string, string>;
  annotations: Record<string, string>;
}

export interface ProxyConfiguration {
  logLevel: 'trace' | 'debug' | 'info' | 'warn' | 'error';
  tracingConfig: TracingConfig;
  metricsConfig: MetricsConfig;
  accessLogConfig: AccessLogConfig;
  concurrency: number;
  drainTime: number;
  parentShutdownTime: number;
}

export interface UpdateStrategy {
  type: 'rolling' | 'recreate' | 'canary';
  maxUnavailable: string;
  maxSurge: string;
  rollbackConfig?: RollbackConfig;
}

export interface RollbackConfig {
  enabled: boolean;
  healthCheckThreshold: number;
  timeoutSeconds: number;
  autoRollback: boolean;
}

export interface SecurityConfig {
  mtls: MutualTLSConfig;
  authorization: AuthorizationConfig;
  authentication: AuthenticationConfig;
  certificateManagement: CertificateManagementConfig;
  securityPolicies: SecurityPolicy[];
}

export interface MutualTLSConfig {
  mode: 'strict' | 'permissive' | 'disabled';
  autoMtls: boolean;
  certificateProvider: 'istiod' | 'cert-manager' | 'vault' | 'external';
  minTlsVersion: '1.0' | '1.1' | '1.2' | '1.3';
  cipherSuites: string[];
}

export interface AuthorizationConfig {
  enabled: boolean;
  defaultAction: 'allow' | 'deny';
  policies: AuthorizationPolicy[];
  auditLogging: boolean;
}

export interface AuthenticationConfig {
  providers: IdentityProvider[];
  jwtValidation: JWTValidationConfig;
  sessionManagement: SessionManagementConfig;
}

export interface IdentityProvider {
  name: string;
  type: 'oidc' | 'oauth2' | 'ldap' | 'custom';
  configuration: Record<string, any>;
  enabled: boolean;
}

export interface JWTValidationConfig {
  enabled: boolean;
  audiences: string[];
  issuers: string[];
  jwksUri: string;
  clockSkew: number;
}

export interface SessionManagementConfig {
  enabled: boolean;
  timeout: number;
  cookieConfig: CookieConfig;
}

export interface CookieConfig {
  name: string;
  secure: boolean;
  httpOnly: boolean;
  sameSite: 'strict' | 'lax' | 'none';
  domain?: string;
  path: string;
}

export interface CertificateManagementConfig {
  rootCa: CertificateConfig;
  intermediateCa: CertificateConfig;
  workloadCerts: WorkloadCertConfig;
  rotation: CertRotationConfig;
}

export interface CertificateConfig {
  provider: string;
  keySize: number;
  algorithm: string;
  validityDuration: string;
  renewBefore: string;
}

export interface WorkloadCertConfig extends CertificateConfig {
  san: SubjectAlternativeName[];
  autoProvision: boolean;
  trustDomain: string;
}

export interface SubjectAlternativeName {
  type: 'dns' | 'ip' | 'uri' | 'email';
  value: string;
}

export interface CertRotationConfig {
  enabled: boolean;
  schedule: string;
  gracePeriod: string;
  notifyBeforeExpiry: string;
}

export interface SecurityPolicy {
  id: string;
  name: string;
  type: 'authorization' | 'authentication' | 'network' | 'traffic';
  enabled: boolean;
  priority: number;
  rules: PolicyRule[];
  enforcement: 'strict' | 'permissive' | 'dry-run';
  createdAt: Date;
  updatedAt: Date;
}

export interface PolicyRule {
  id: string;
  source: PolicySelector;
  destination: PolicySelector;
  operation: OperationSelector;
  condition: PolicyCondition;
  action: 'allow' | 'deny' | 'log' | 'rate-limit';
  metadata: Record<string, any>;
}

export interface PolicySelector {
  principals: string[];
  namespaces: string[];
  labels: Record<string, string>;
  ipBlocks: string[];
}

export interface OperationSelector {
  methods: string[];
  paths: string[];
  headers: Record<string, string>;
  ports: number[];
}

export interface PolicyCondition {
  type: 'always' | 'time' | 'rate' | 'attribute' | 'expression';
  configuration: Record<string, any>;
}

export interface AuthorizationPolicy {
  id: string;
  name: string;
  namespace: string;
  selector: WorkloadSelector;
  rules: AuthorizationRule[];
  action: 'allow' | 'deny' | 'audit' | 'custom';
  enabled: boolean;
}

export interface WorkloadSelector {
  matchLabels: Record<string, string>;
  matchExpressions: LabelExpression[];
}

export interface LabelExpression {
  key: string;
  operator: 'in' | 'not-in' | 'exists' | 'does-not-exist';
  values: string[];
}

export interface AuthorizationRule {
  from: AuthorizationRuleFrom[];
  to: AuthorizationRuleTo[];
  when: AuthorizationRuleCondition[];
}

export interface AuthorizationRuleFrom {
  source: AuthorizationSource;
}

export interface AuthorizationSource {
  principals: string[];
  requestPrincipals: string[];
  namespaces: string[];
  ipBlocks: string[];
  remoteIpBlocks: string[];
}

export interface AuthorizationRuleTo {
  operation: AuthorizationOperation;
}

export interface AuthorizationOperation {
  methods: string[];
  hosts: string[];
  ports: string[];
  paths: string[];
}

export interface AuthorizationRuleCondition {
  key: string;
  values: string[];
  notValues: string[];
}

export interface ObservabilityConfig {
  metrics: MetricsConfig;
  tracing: TracingConfig;
  logging: LoggingConfig;
  visualization: VisualizationConfig;
  alerting: AlertingConfig;
}

export interface MetricsConfig {
  enabled: boolean;
  providers: MetricsProvider[];
  retention: string;
  scrapeInterval: string;
  resolution: string;
  customMetrics: CustomMetric[];
}

export interface MetricsProvider {
  name: string;
  type: 'prometheus' | 'datadog' | 'new-relic' | 'custom';
  endpoint: string;
  authentication: AuthConfig;
  configuration: Record<string, any>;
}

export interface AuthConfig {
  type: 'none' | 'basic' | 'bearer' | 'api-key' | 'oauth2';
  credentials: Record<string, string>;
}

export interface CustomMetric {
  name: string;
  type: 'counter' | 'gauge' | 'histogram' | 'summary';
  description: string;
  labels: string[];
  unit?: string;
}

export interface TracingConfig {
  enabled: boolean;
  samplingRate: number;
  providers: TracingProvider[];
  exporters: TracingExporter[];
  baggage: BaggageConfig;
}

export interface TracingProvider {
  name: string;
  type: 'jaeger' | 'zipkin' | 'datadog' | 'lightstep' | 'custom';
  endpoint: string;
  authentication: AuthConfig;
  configuration: Record<string, any>;
}

export interface TracingExporter {
  name: string;
  type: 'otlp' | 'jaeger' | 'zipkin' | 'logging';
  endpoint: string;
  compression: 'none' | 'gzip';
  headers: Record<string, string>;
}

export interface BaggageConfig {
  enabled: boolean;
  maxEntries: number;
  maxKeyLength: number;
  maxValueLength: number;
}

export interface LoggingConfig {
  enabled: boolean;
  level: 'trace' | 'debug' | 'info' | 'warn' | 'error';
  format: 'json' | 'text';
  providers: LoggingProvider[];
  structured: boolean;
  sampling: LogSamplingConfig;
}

export interface LoggingProvider {
  name: string;
  type: 'elasticsearch' | 'fluentd' | 'datadog' | 'splunk' | 'loki' | 'custom';
  endpoint: string;
  authentication: AuthConfig;
  configuration: Record<string, any>;
}

export interface LogSamplingConfig {
  enabled: boolean;
  rate: number;
  burst: number;
  thereafter: number;
}

export interface AccessLogConfig {
  enabled: boolean;
  format: string;
  encoding: 'text' | 'json';
  providers: string[];
  includeFilterState: boolean;
  includeRequestBody: boolean;
  includeResponseBody: boolean;
}

export interface VisualizationConfig {
  enabled: boolean;
  dashboards: Dashboard[];
  grafana: GrafanaConfig;
  kiali: KialiConfig;
}

export interface Dashboard {
  id: string;
  name: string;
  type: 'service-topology' | 'traffic-flow' | 'performance' | 'security' | 'custom';
  configuration: Record<string, any>;
  panels: DashboardPanel[];
}

export interface DashboardPanel {
  id: string;
  title: string;
  type: 'graph' | 'table' | 'stat' | 'heatmap' | 'logs';
  query: string;
  visualization: VisualizationSettings;
}

export interface VisualizationSettings {
  displayMode: string;
  colorScheme: string;
  thresholds: Threshold[];
  overrides: Override[];
}

export interface Threshold {
  value: number;
  color: string;
  condition: 'gt' | 'lt' | 'eq';
}

export interface Override {
  matcher: string;
  properties: Record<string, any>;
}

export interface GrafanaConfig {
  enabled: boolean;
  endpoint: string;
  authentication: AuthConfig;
  dashboardIds: string[];
}

export interface KialiConfig {
  enabled: boolean;
  endpoint: string;
  authentication: AuthConfig;
  configuration: Record<string, any>;
}

export interface AlertingConfig {
  enabled: boolean;
  rules: AlertRule[];
  channels: AlertChannel[];
  silences: AlertSilence[];
}

export interface AlertRule {
  id: string;
  name: string;
  description: string;
  expression: string;
  severity: 'info' | 'warning' | 'critical';
  duration: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  enabled: boolean;
}

export interface AlertChannel {
  id: string;
  name: string;
  type: 'email' | 'slack' | 'webhook' | 'pagerduty' | 'custom';
  configuration: Record<string, any>;
  enabled: boolean;
}

export interface AlertSilence {
  id: string;
  matchers: AlertMatcher[];
  startsAt: Date;
  endsAt: Date;
  comment: string;
  createdBy: string;
}

export interface AlertMatcher {
  name: string;
  value: string;
  isRegex: boolean;
}

export interface TrafficConfig {
  loadBalancing: LoadBalancingConfig;
  circuitBreaker: CircuitBreakerConfig;
  retryPolicy: RetryPolicyConfig;
  timeout: TimeoutConfig;
  routing: RoutingConfig;
  rateLimit: RateLimitConfig;
}

export interface LoadBalancingConfig {
  algorithm: 'round-robin' | 'least-connections' | 'random' | 'weighted' | 'consistent-hash';
  healthCheck: HealthCheckConfig;
  outlierDetection: OutlierDetectionConfig;
  localityPreference: LocalityPreference;
}

export interface HealthCheckConfig {
  enabled: boolean;
  interval: string;
  timeout: string;
  healthyThreshold: number;
  unhealthyThreshold: number;
  path?: string;
  method?: string;
  expectedStatuses: number[];
}

export interface OutlierDetectionConfig {
  enabled: boolean;
  consecutiveErrors: number;
  interval: string;
  baseEjectionTime: string;
  maxEjectionPercent: number;
  minHealthPercent: number;
}

export interface LocalityPreference {
  enabled: boolean;
  failover: LocalityFailover[];
  distribute: LocalityDistribution[];
}

export interface LocalityFailover {
  from: string;
  to: string;
}

export interface LocalityDistribution {
  locality: string;
  weight: number;
}

export interface CircuitBreakerConfig {
  enabled: boolean;
  maxConnections: number;
  maxPendingRequests: number;
  maxRequests: number;
  maxRetries: number;
  trackRemaining: boolean;
}

export interface RetryPolicyConfig {
  enabled: boolean;
  attempts: number;
  perTryTimeout: string;
  retryOn: string[];
  retryRemoteLocalities: boolean;
  backoff: BackoffConfig;
}

export interface BackoffConfig {
  baseInterval: string;
  maxInterval: string;
  multiplier: number;
}

export interface TimeoutConfig {
  enabled: boolean;
  requestTimeout: string;
  idleTimeout: string;
  streamIdleTimeout: string;
  maxStreamDuration: string;
}

export interface RoutingConfig {
  rules: RoutingRule[];
  defaultRoute: DefaultRoute;
  faultInjection: FaultInjectionConfig;
  trafficSplitting: TrafficSplittingConfig;
}

export interface RoutingRule {
  id: string;
  name: string;
  priority: number;
  match: RouteMatch;
  route: RouteDestination[];
  redirect?: RouteRedirect;
  rewrite?: RouteRewrite;
  fault?: FaultInjection;
  mirror?: RouteMirror;
}

export interface RouteMatch {
  headers: HeaderMatch[];
  method: HTTPMethod;
  path: PathMatch;
  queryParams: QueryParamMatch[];
}

export interface HeaderMatch {
  name: string;
  value?: string;
  regex?: string;
  invert: boolean;
}

export interface HTTPMethod {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' | 'HEAD' | 'OPTIONS';
}

export interface PathMatch {
  type: 'exact' | 'prefix' | 'regex';
  value: string;
  caseSensitive: boolean;
}

export interface QueryParamMatch {
  name: string;
  value?: string;
  regex?: string;
}

export interface RouteDestination {
  destination: ServiceDestination;
  weight: number;
  headers?: HeaderManipulation;
}

export interface ServiceDestination {
  host: string;
  subset?: string;
  port?: number;
}

export interface HeaderManipulation {
  set: Record<string, string>;
  add: Record<string, string>;
  remove: string[];
}

export interface RouteRedirect {
  uri?: string;
  authority?: string;
  scheme?: string;
  redirectCode: number;
}

export interface RouteRewrite {
  uri?: string;
  authority?: string;
}

export interface FaultInjection {
  delay?: DelayFault;
  abort?: AbortFault;
}

export interface DelayFault {
  percentage: number;
  fixedDelay: string;
}

export interface AbortFault {
  percentage: number;
  httpStatus: number;
}

export interface RouteMirror {
  host: string;
  subset?: string;
  percentage: number;
}

export interface DefaultRoute {
  destination: ServiceDestination;
  timeout?: string;
  retryPolicy?: RetryPolicyConfig;
}

export interface FaultInjectionConfig {
  enabled: boolean;
  rules: FaultInjectionRule[];
}

export interface FaultInjectionRule {
  id: string;
  name: string;
  selector: WorkloadSelector;
  faults: FaultInjection[];
  enabled: boolean;
}

export interface TrafficSplittingConfig {
  enabled: boolean;
  experiments: TrafficExperiment[];
}

export interface TrafficExperiment {
  id: string;
  name: string;
  status: 'active' | 'paused' | 'completed';
  variants: TrafficVariant[];
  metrics: ExperimentMetric[];
  duration: string;
  startTime: Date;
  endTime?: Date;
}

export interface TrafficVariant {
  name: string;
  weight: number;
  destination: ServiceDestination;
  headers?: HeaderManipulation;
}

export interface ExperimentMetric {
  name: string;
  type: 'success-rate' | 'latency' | 'error-rate' | 'custom';
  target: number;
  tolerance: number;
}

export interface RateLimitConfig {
  enabled: boolean;
  rules: RateLimitRule[];
  rateLimitService: RateLimitService;
}

export interface RateLimitRule {
  id: string;
  name: string;
  selector: WorkloadSelector;
  limits: RateLimit[];
  enabled: boolean;
}

export interface RateLimit {
  unit: 'second' | 'minute' | 'hour' | 'day';
  requests: number;
  dimensions: RateLimitDimension[];
}

export interface RateLimitDimension {
  type: 'header' | 'remote-address' | 'source-cluster' | 'destination-cluster' | 'custom';
  key?: string;
  defaultValue?: string;
}

export interface RateLimitService {
  enabled: boolean;
  service: string;
  timeout: string;
  failureMode: 'allow' | 'deny';
}

export interface ResourceRequirements {
  requests: ResourceSpec;
  limits: ResourceSpec;
}

export interface ResourceSpec {
  cpu: string;
  memory: string;
  storage?: string;
}

export interface ControlPlaneMetrics {
  istiodMemoryUsage: number;
  istiodCpuUsage: number;
  configUpdates: number;
  connectedProxies: number;
  certificateRotations: number;
  lastConfigPush: Date;
}

export interface ServiceMeshService {
  id: string;
  name: string;
  namespace: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  ports: ServicePort[];
  endpoints: ServiceEndpoint[];
  virtualServices: string[];
  destinationRules: string[];
  health: ServiceHealth;
  metrics: ServiceMetrics;
  status: 'healthy' | 'degraded' | 'unhealthy';
}

export interface ServicePort {
  name: string;
  port: number;
  targetPort: number;
  protocol: 'TCP' | 'UDP' | 'HTTP' | 'HTTPS' | 'GRPC';
}

export interface ServiceEndpoint {
  ip: string;
  port: number;
  ready: boolean;
  labels: Record<string, string>;
}

export interface ServiceHealth {
  healthy: number;
  unhealthy: number;
  unknown: number;
  lastCheck: Date;
}

export interface ServiceMetrics {
  requestRate: number;
  errorRate: number;
  p50Latency: number;
  p95Latency: number;
  p99Latency: number;
  inboundSuccessRate: number;
  outboundSuccessRate: number;
  tcpConnections: number;
}

export interface ServiceMeshWorkload {
  id: string;
  name: string;
  namespace: string;
  type: 'deployment' | 'daemonset' | 'statefulset' | 'job';
  labels: Record<string, string>;
  annotations: Record<string, string>;
  replicas: WorkloadReplicas;
  containers: Container[];
  sidecar: SidecarConfig;
  health: WorkloadHealth;
  metrics: WorkloadMetrics;
}

export interface WorkloadReplicas {
  desired: number;
  current: number;
  ready: number;
  available: number;
}

export interface Container {
  name: string;
  image: string;
  resources: ResourceRequirements;
  ports: ContainerPort[];
  env: EnvVar[];
}

export interface ContainerPort {
  name: string;
  containerPort: number;
  protocol: 'TCP' | 'UDP';
}

export interface EnvVar {
  name: string;
  value?: string;
  valueFrom?: EnvVarSource;
}

export interface EnvVarSource {
  configMapKeyRef?: ConfigMapKeySelector;
  secretKeyRef?: SecretKeySelector;
  fieldRef?: FieldSelector;
}

export interface ConfigMapKeySelector {
  name: string;
  key: string;
  optional: boolean;
}

export interface SecretKeySelector {
  name: string;
  key: string;
  optional: boolean;
}

export interface FieldSelector {
  fieldPath: string;
}

export interface SidecarConfig {
  injected: boolean;
  image: string;
  version: string;
  resources: ResourceRequirements;
  configuration: ProxyConfiguration;
  status: 'running' | 'pending' | 'failed' | 'unknown';
}

export interface WorkloadHealth {
  overallHealth: 'healthy' | 'degraded' | 'unhealthy';
  podHealth: PodHealth[];
  readinessProbe: ProbeHealth;
  livenessProbe: ProbeHealth;
}

export interface PodHealth {
  name: string;
  status: 'running' | 'pending' | 'failed' | 'succeeded' | 'unknown';
  ready: boolean;
  restarts: number;
}

export interface ProbeHealth {
  configured: boolean;
  healthy: boolean;
  lastProbe: Date;
  failureThreshold: number;
  successThreshold: number;
}

export interface WorkloadMetrics {
  cpuUsage: number;
  memoryUsage: number;
  networkIn: number;
  networkOut: number;
  requestsPerSecond: number;
  errorsPerSecond: number;
  averageResponseTime: number;
}

export interface ServiceTopology {
  services: TopologyService[];
  connections: ServiceConnection[];
  clusters: TopologyCluster[];
  lastUpdated: Date;
}

export interface TopologyService {
  id: string;
  name: string;
  namespace: string;
  type: 'internal' | 'external' | 'gateway';
  health: 'healthy' | 'degraded' | 'unhealthy';
  position: Position;
  metadata: Record<string, any>;
}

export interface ServiceConnection {
  source: string;
  destination: string;
  protocol: string;
  encrypted: boolean;
  requestRate: number;
  errorRate: number;
  responseTime: number;
  health: 'healthy' | 'degraded' | 'unhealthy';
}

export interface TopologyCluster {
  id: string;
  name: string;
  services: string[];
  position: Position;
  collapsed: boolean;
}

export interface Position {
  x: number;
  y: number;
}

export interface CanaryDeployment {
  id: string;
  name: string;
  service: string;
  namespace: string;
  status: 'pending' | 'running' | 'succeeded' | 'failed' | 'paused';
  strategy: CanaryStrategy;
  baseline: DeploymentVersion;
  canary: DeploymentVersion;
  metrics: CanaryMetrics;
  analysis: CanaryAnalysis;
  steps: CanaryStep[];
  currentStep: number;
  startTime: Date;
  endTime?: Date;
  duration: string;
}

export interface CanaryStrategy {
  type: 'blue-green' | 'rolling' | 'traffic-splitting';
  maxUnavailable: string;
  maxSurge: string;
  steps: CanaryStepConfig[];
  analysis: AnalysisConfig;
}

export interface CanaryStepConfig {
  weight: number;
  duration: string;
  analysis?: AnalysisConfig;
}

export interface AnalysisConfig {
  enabled: boolean;
  metrics: AnalysisMetric[];
  successCondition: string;
  failureCondition: string;
  interval: string;
  count: number;
  successfulRunHistoryLimit: number;
  unsuccessfulRunHistoryLimit: number;
}

export interface AnalysisMetric {
  name: string;
  provider: string;
  query: string;
  successCondition: string;
  failureCondition: string;
  interval: string;
  count: number;
}

export interface DeploymentVersion {
  image: string;
  replicas: number;
  weight: number;
  health: 'healthy' | 'degraded' | 'unhealthy';
  readyReplicas: number;
}

export interface CanaryMetrics {
  successRate: number;
  errorRate: number;
  averageResponseTime: number;
  requestsPerSecond: number;
  cpuUsage: number;
  memoryUsage: number;
  customMetrics: Record<string, number>;
}

export interface CanaryAnalysis {
  status: 'pending' | 'running' | 'successful' | 'failed' | 'error';
  phase: 'initializing' | 'running' | 'successful' | 'failed';
  message: string;
  runs: AnalysisRun[];
}

export interface AnalysisRun {
  name: string;
  phase: 'pending' | 'running' | 'successful' | 'failed' | 'error';
  message: string;
  startedAt: Date;
  finishedAt?: Date;
  measurements: Measurement[];
}

export interface Measurement {
  name: string;
  phase: 'pending' | 'running' | 'successful' | 'failed' | 'error';
  value: string;
  message: string;
  startedAt: Date;
  finishedAt?: Date;
}

export interface CanaryStep {
  index: number;
  name: string;
  status: 'pending' | 'running' | 'successful' | 'failed' | 'skipped';
  weight: number;
  duration: string;
  startTime?: Date;
  endTime?: Date;
  message?: string;
}