export interface GraphQLSchema {
  id: string;
  name: string;
  version: string;
  description: string;
  sdl: string; // Schema Definition Language
  introspection: IntrospectionResult;
  endpoints: GraphQLEndpoint[];
  resolvers: ResolverInfo[];
  directives: DirectiveInfo[];
  types: TypeInfo[];
  queries: FieldInfo[];
  mutations: FieldInfo[];
  subscriptions: FieldInfo[];
  createdAt: Date;
  updatedAt: Date;
  status: 'draft' | 'published' | 'deprecated';
  metadata: Record<string, any>;
}

export interface GraphQLEndpoint {
  id: string;
  name: string;
  url: string;
  environment: 'development' | 'staging' | 'production';
  authentication: AuthenticationConfig;
  headers: Record<string, string>;
  rateLimit: RateLimitConfig;
  caching: CachingConfig;
  monitoring: MonitoringConfig;
  status: 'healthy' | 'degraded' | 'offline';
  lastHealthCheck: Date;
  metrics: EndpointMetrics;
}

export interface AuthenticationConfig {
  type: 'none' | 'api-key' | 'bearer' | 'oauth2' | 'custom';
  apiKey?: string;
  bearerToken?: string;
  oauth2Config?: OAuth2Config;
  customHeaders?: Record<string, string>;
}

export interface OAuth2Config {
  clientId: string;
  clientSecret: string;
  tokenUrl: string;
  scopes: string[];
  grantType: 'client_credentials' | 'authorization_code';
}

export interface RateLimitConfig {
  enabled: boolean;
  requestsPerMinute: number;
  requestsPerHour: number;
  requestsPerDay: number;
  burst: number;
}

export interface CachingConfig {
  enabled: boolean;
  ttl: number;
  maxAge: number;
  staleWhileRevalidate: boolean;
  cacheKeyStrategy: 'query-hash' | 'custom' | 'headers';
  customKey?: string;
  invalidationRules: InvalidationRule[];
}

export interface InvalidationRule {
  mutations: string[];
  affectedQueries: string[];
  strategy: 'immediate' | 'lazy' | 'time-based';
  delay?: number;
}

export interface MonitoringConfig {
  enabled: boolean;
  logQueries: boolean;
  logErrors: boolean;
  logSlowQueries: boolean;
  slowQueryThreshold: number;
  alerting: AlertingConfig;
  metrics: MetricsConfig;
}

export interface AlertingConfig {
  errorRateThreshold: number;
  responseTimeThreshold: number;
  availabilityThreshold: number;
  channels: string[];
}

export interface MetricsConfig {
  collectRequestMetrics: boolean;
  collectFieldMetrics: boolean;
  collectErrorMetrics: boolean;
  retentionDays: number;
}

export interface EndpointMetrics {
  requestsPerSecond: number;
  averageResponseTime: number;
  errorRate: number;
  availability: number;
  cacheHitRate: number;
  concurrentConnections: number;
  totalRequests: number;
  totalErrors: number;
  lastUpdate: Date;
}

export interface IntrospectionResult {
  schemaHash: string;
  types: IntrospectionType[];
  queryType: IntrospectionType;
  mutationType?: IntrospectionType;
  subscriptionType?: IntrospectionType;
  directives: IntrospectionDirective[];
  lastIntrospection: Date;
}

export interface IntrospectionType {
  kind: 'SCALAR' | 'OBJECT' | 'INTERFACE' | 'UNION' | 'ENUM' | 'INPUT_OBJECT' | 'LIST' | 'NON_NULL';
  name: string;
  description?: string;
  fields?: IntrospectionField[];
  inputFields?: IntrospectionInputValue[];
  interfaces?: IntrospectionType[];
  enumValues?: IntrospectionEnumValue[];
  possibleTypes?: IntrospectionType[];
  ofType?: IntrospectionType;
}

export interface IntrospectionField {
  name: string;
  description?: string;
  args: IntrospectionInputValue[];
  type: IntrospectionType;
  isDeprecated: boolean;
  deprecationReason?: string;
}

export interface IntrospectionInputValue {
  name: string;
  description?: string;
  type: IntrospectionType;
  defaultValue?: string;
}

export interface IntrospectionEnumValue {
  name: string;
  description?: string;
  isDeprecated: boolean;
  deprecationReason?: string;
}

export interface IntrospectionDirective {
  name: string;
  description?: string;
  locations: string[];
  args: IntrospectionInputValue[];
}

export interface ResolverInfo {
  fieldName: string;
  typeName: string;
  complexity: number;
  avgExecutionTime: number;
  cachePolicy?: CachePolicy;
  rateLimiting?: RateLimitPolicy;
  authorization?: AuthorizationPolicy;
}

export interface CachePolicy {
  maxAge: number;
  scope: 'public' | 'private';
  hints: CacheHint[];
}

export interface CacheHint {
  path: string;
  maxAge: number;
  scope: 'public' | 'private';
}

export interface RateLimitPolicy {
  max: number;
  window: string;
  message?: string;
  skipSuccessfulRequests?: boolean;
}

export interface AuthorizationPolicy {
  requiresAuth: boolean;
  roles?: string[];
  permissions?: string[];
  customRule?: string;
}

export interface DirectiveInfo {
  name: string;
  description: string;
  locations: string[];
  arguments: ArgumentInfo[];
  isRepeatable: boolean;
}

export interface ArgumentInfo {
  name: string;
  type: string;
  defaultValue?: string;
  description?: string;
}

export interface TypeInfo {
  name: string;
  kind: string;
  description?: string;
  fields: FieldInfo[];
  interfaces: string[];
  enumValues?: EnumValueInfo[];
  inputFields?: InputFieldInfo[];
}

export interface FieldInfo {
  name: string;
  type: string;
  description?: string;
  arguments: ArgumentInfo[];
  isDeprecated: boolean;
  deprecationReason?: string;
}

export interface EnumValueInfo {
  name: string;
  description?: string;
  isDeprecated: boolean;
  deprecationReason?: string;
}

export interface InputFieldInfo {
  name: string;
  type: string;
  description?: string;
  defaultValue?: string;
}

export interface GraphQLQuery {
  id: string;
  name: string;
  query: string;
  variables?: Record<string, any>;
  operationType: 'query' | 'mutation' | 'subscription';
  endpoint: string;
  createdAt: Date;
  lastExecuted?: Date;
  executionCount: number;
  avgExecutionTime: number;
  tags: string[];
  folder?: string;
  isFavorite: boolean;
  description?: string;
}

export interface QueryExecution {
  id: string;
  queryId: string;
  query: string;
  variables?: Record<string, any>;
  endpoint: string;
  startTime: Date;
  endTime?: Date;
  duration?: number;
  status: 'running' | 'success' | 'error' | 'cancelled';
  result?: any;
  errors?: GraphQLError[];
  metrics: ExecutionMetrics;
}

export interface GraphQLError {
  message: string;
  locations?: ErrorLocation[];
  path?: (string | number)[];
  extensions?: Record<string, any>;
}

export interface ErrorLocation {
  line: number;
  column: number;
}

export interface ExecutionMetrics {
  requestSize: number;
  responseSize: number;
  networkTime: number;
  parsingTime: number;
  validationTime: number;
  executionTime: number;
  totalTime: number;
  cacheHit: boolean;
  rateLimited: boolean;
}

export interface Subscription {
  id: string;
  name: string;
  query: string;
  variables?: Record<string, any>;
  endpoint: string;
  status: 'connected' | 'connecting' | 'disconnected' | 'error';
  createdAt: Date;
  lastMessage?: Date;
  messageCount: number;
  autoReconnect: boolean;
  reconnectAttempts: number;
  maxReconnectAttempts: number;
  reconnectInterval: number;
}

export interface SubscriptionMessage {
  id: string;
  subscriptionId: string;
  type: 'start' | 'data' | 'error' | 'complete';
  payload?: any;
  timestamp: Date;
}

export interface SchemaComparison {
  id: string;
  name: string;
  baseSchema: string;
  targetSchema: string;
  comparison: ComparisonResult;
  createdAt: Date;
}

export interface ComparisonResult {
  breaking: BreakingChange[];
  dangerous: DangerousChange[];
  safe: SafeChange[];
  summary: ComparisonSummary;
}

export interface BreakingChange {
  type: string;
  description: string;
  path: string;
  severity: 'high' | 'medium' | 'low';
}

export interface DangerousChange {
  type: string;
  description: string;
  path: string;
  reason: string;
}

export interface SafeChange {
  type: string;
  description: string;
  path: string;
}

export interface ComparisonSummary {
  totalChanges: number;
  breakingChanges: number;
  dangerousChanges: number;
  safeChanges: number;
  compatibility: 'compatible' | 'potentially-breaking' | 'breaking';
}

export interface GraphQLConfiguration {
  defaultEndpoint: string;
  defaultHeaders: Record<string, string>;
  timeout: number;
  retries: number;
  caching: {
    enabled: boolean;
    defaultTtl: number;
    maxCacheSize: number;
  };
  subscriptions: {
    transport: 'websocket' | 'sse';
    reconnectInterval: number;
    maxReconnectAttempts: number;
    heartbeatInterval: number;
  };
  introspection: {
    enabled: boolean;
    interval: number;
    autoRefresh: boolean;
  };
  validation: {
    enabled: boolean;
    strictMode: boolean;
    customRules: string[];
  };
  security: {
    enableQueryWhitelist: boolean;
    maxQueryDepth: number;
    maxQueryComplexity: number;
    disableIntrospection: boolean;
  };
}

export interface QueryComplexityAnalysis {
  query: string;
  complexity: number;
  depth: number;
  fieldCount: number;
  aliasCount: number;
  warnings: string[];
  suggestions: string[];
}

export interface PlaygroundSettings {
  theme: 'light' | 'dark';
  fontSize: number;
  tabSize: number;
  wordWrap: boolean;
  autoComplete: boolean;
  linting: boolean;
  prettify: boolean;
  shareableUrls: boolean;
  requestCredentials: 'omit' | 'include' | 'same-origin';
}