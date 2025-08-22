// Enhanced Analytics and Reporting Types

// Core Analytics Types
export interface AnalyticsEvent {
  id: string
  type: AnalyticsEventType
  category: string
  action: string
  label?: string
  value?: number
  userId?: string
  sessionId: string
  timestamp: number
  metadata?: Record<string, any>
  context: EventContext
}

export type AnalyticsEventType = 
  | 'page_view'
  | 'click'
  | 'form_submit'
  | 'api_call'
  | 'error'
  | 'performance'
  | 'user_interaction'
  | 'custom'

export interface EventContext {
  url: string
  referrer: string
  userAgent: string
  viewport: { width: number; height: number }
  device: DeviceInfo
  location?: GeolocationCoordinates
}

export interface DeviceInfo {
  type: 'desktop' | 'tablet' | 'mobile'
  os: string
  browser: string
  version: string
}

// Enhanced Performance Metrics
export interface EnhancedPerformanceMetrics {
  webVitals: WebVitalsMetrics
  runtime: RuntimeMetrics
  network: NetworkPerformance
  resources: ResourceMetrics
  user: UserExperienceMetrics
}

export interface WebVitalsMetrics {
  fcp: number // First Contentful Paint
  lcp: number // Largest Contentful Paint
  fid: number // First Input Delay
  cls: number // Cumulative Layout Shift
  ttfb: number // Time to First Byte
  inp?: number // Interaction to Next Paint
}

export interface RuntimeMetrics {
  memory: MemoryMetrics
  cpu: CPUMetrics
  battery?: BatteryMetrics
}

export interface MemoryMetrics {
  used: number
  total: number
  limit: number
  peak: number
}

export interface CPUMetrics {
  usage: number
  cores: number
  speed: number
}

export interface BatteryMetrics {
  level: number
  charging: boolean
  dischargingTime?: number
  chargingTime?: number
}

export interface NetworkPerformance {
  effectiveType: string
  rtt: number
  downlink: number
  saveData: boolean
  requests: NetworkRequestMetrics[]
}

export interface NetworkRequestMetrics {
  url: string
  method: string
  status: number
  duration: number
  size: number
  cached: boolean
  type: 'fetch' | 'xhr' | 'beacon'
}

export interface ResourceMetrics {
  scripts: ResourceTiming[]
  stylesheets: ResourceTiming[]
  images: ResourceTiming[]
  fonts: ResourceTiming[]
  other: ResourceTiming[]
}

export interface ResourceTiming {
  name: string
  size: number
  duration: number
  cached: boolean
  compressed: boolean
}

export interface UserExperienceMetrics {
  totalSessionTime: number
  activeTime: number
  idleTime: number
  interactions: number
  scrollDepth: number
  bounceRate: number
}

// Business Intelligence Types
export interface BusinessMetrics {
  revenue: RevenueMetrics
  users: UserMetrics
  engagement: EngagementMetrics
  conversion: ConversionMetrics
  retention: RetentionMetrics
}

export interface RevenueMetrics {
  total: number
  recurring: number
  oneTime: number
  arpu: number // Average Revenue Per User
  ltv: number // Lifetime Value
  churnRate: number
  growthRate: number
  byProduct: ProductRevenue[]
  byRegion: RegionRevenue[]
}

export interface ProductRevenue {
  productId: string
  name: string
  revenue: number
  units: number
  averagePrice: number
}

export interface RegionRevenue {
  region: string
  revenue: number
  users: number
  averageOrderValue: number
}

export interface UserMetrics {
  total: number
  active: number
  new: number
  returning: number
  segments: UserSegment[]
  demographics: Demographics
  acquisition: AcquisitionMetrics
}

export interface UserSegment {
  id: string
  name: string
  count: number
  percentage: number
  criteria: Record<string, any>
}

export interface Demographics {
  ageGroups: AgeGroup[]
  genderDistribution: GenderDistribution
  locationDistribution: LocationDistribution[]
  deviceDistribution: DeviceDistribution[]
}

export interface AgeGroup {
  range: string
  count: number
  percentage: number
}

export interface GenderDistribution {
  male: number
  female: number
  other: number
  unknown: number
}

export interface LocationDistribution {
  country: string
  region?: string
  city?: string
  count: number
  percentage: number
}

export interface DeviceDistribution {
  type: string
  count: number
  percentage: number
}

export interface AcquisitionMetrics {
  channels: AcquisitionChannel[]
  cost: CostMetrics
  quality: QualityMetrics
}

export interface AcquisitionChannel {
  channel: string
  users: number
  cost: number
  conversionRate: number
  ltv: number
  roi: number
}

export interface CostMetrics {
  total: number
  cac: number // Customer Acquisition Cost
  cpm: number // Cost Per Mille
  cpc: number // Cost Per Click
  cpl: number // Cost Per Lead
}

export interface QualityMetrics {
  averageSessionDuration: number
  bounceRate: number
  pageDepth: number
  timeToConversion: number
}

export interface EngagementMetrics {
  pageViews: number
  uniquePageViews: number
  sessionDuration: number
  bounceRate: number
  exitRate: number
  interactions: InteractionMetrics
  content: ContentMetrics[]
}

export interface InteractionMetrics {
  clicks: number
  scrolls: number
  hovers: number
  keystrokes: number
  touches: number
  gestures: number
}

export interface ContentMetrics {
  path: string
  title: string
  views: number
  uniqueViews: number
  timeOnPage: number
  exitRate: number
  shareCount: number
}

export interface ConversionMetrics {
  overall: ConversionData
  byFunnel: FunnelMetrics[]
  bySource: SourceConversion[]
  goals: GoalMetrics[]
}

export interface ConversionData {
  rate: number
  total: number
  value: number
  averageOrderValue: number
}

export interface FunnelMetrics {
  name: string
  steps: FunnelStep[]
  overallConversion: number
  dropOffPoints: DropOffPoint[]
}

export interface FunnelStep {
  name: string
  users: number
  conversionRate: number
  dropOffRate: number
}

export interface DropOffPoint {
  step: string
  dropOffRate: number
  commonReasons: string[]
}

export interface SourceConversion {
  source: string
  users: number
  conversions: number
  rate: number
  value: number
}

export interface GoalMetrics {
  id: string
  name: string
  completions: number
  value: number
  conversionRate: number
}

export interface RetentionMetrics {
  cohorts: CohortMetrics[]
  overall: RetentionData
  bySegment: SegmentRetention[]
}

export interface CohortMetrics {
  cohort: string
  size: number
  retention: RetentionPeriod[]
}

export interface RetentionPeriod {
  period: number
  retained: number
  percentage: number
}

export interface RetentionData {
  day1: number
  day7: number
  day30: number
  day90: number
}

export interface SegmentRetention {
  segment: string
  retention: RetentionData
}

// Compliance and Reporting Types
export interface ComplianceReport {
  id: string
  type: ComplianceType
  generatedAt: number
  period: DateRange
  data: ComplianceData
  status: ReportStatus
  metadata: ComplianceMetadata
}

export type ComplianceType = 
  | 'gdpr'
  | 'ccpa'
  | 'hipaa'
  | 'sox'
  | 'pci_dss'
  | 'iso27001'
  | 'custom'

export interface ComplianceData {
  dataProcessing: DataProcessingRecord[]
  userRights: UserRightsRecord[]
  breaches: BreachRecord[]
  consents: ConsentRecord[]
  audits: AuditRecord[]
}

export interface DataProcessingRecord {
  id: string
  dataType: string
  purpose: string
  legalBasis: string
  retention: number
  processing: ProcessingActivity[]
  thirdParties: ThirdPartyProcessor[]
}

export interface ProcessingActivity {
  activity: string
  automated: boolean
  profiling: boolean
  timestamp: number
}

export interface ThirdPartyProcessor {
  name: string
  purpose: string
  dataShared: string[]
  safeguards: string[]
}

export interface UserRightsRecord {
  userId: string
  requestType: UserRightType
  timestamp: number
  status: RequestStatus
  resolution: string
}

export type UserRightType = 
  | 'access'
  | 'rectification'
  | 'erasure'
  | 'portability'
  | 'restriction'
  | 'objection'

export type RequestStatus = 
  | 'pending'
  | 'processing'
  | 'completed'
  | 'rejected'
  | 'partially_completed'

export interface BreachRecord {
  id: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  type: string
  description: string
  detectedAt: number
  reportedAt?: number
  affectedRecords: number
  mitigationSteps: string[]
  status: 'detected' | 'investigating' | 'contained' | 'resolved'
}

export interface ConsentRecord {
  userId: string
  consentType: string
  granted: boolean
  timestamp: number
  method: 'explicit' | 'implicit' | 'legitimate_interest'
  withdrawn?: number
}

export interface AuditRecord {
  id: string
  type: 'access' | 'modification' | 'deletion' | 'export'
  userId?: string
  adminId?: string
  resource: string
  action: string
  timestamp: number
  ipAddress: string
  userAgent: string
  result: 'success' | 'failure' | 'unauthorized'
}

export interface ComplianceMetadata {
  regulations: string[]
  jurisdiction: string
  officer: DataProtectionOfficer
  nextReview: number
  certifications: Certification[]
}

export interface DataProtectionOfficer {
  name: string
  email: string
  phone: string
  certified: boolean
}

export interface Certification {
  name: string
  issuer: string
  validFrom: number
  validUntil: number
  status: 'valid' | 'expired' | 'suspended'
}

// Reporting Types
export interface Report {
  id: string
  name: string
  type: ReportType
  format: ReportFormat
  schedule: ReportSchedule
  recipients: string[]
  filters: ReportFilter[]
  data: any
  status: ReportStatus
  generatedAt?: number
  size?: number
  downloadUrl?: string
}

export type ReportType = 
  | 'analytics'
  | 'performance'
  | 'business'
  | 'compliance'
  | 'security'
  | 'user_behavior'
  | 'financial'
  | 'operational'
  | 'custom'

export type ReportFormat = 
  | 'pdf'
  | 'csv'
  | 'excel'
  | 'json'
  | 'html'
  | 'xml'

export interface ReportSchedule {
  frequency: 'once' | 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly'
  time?: string
  dayOfWeek?: number
  dayOfMonth?: number
  timezone: string
  enabled: boolean
  nextRun?: number
}

export interface ReportFilter {
  field: string
  operator: FilterOperator
  value: any
  label: string
}

export type FilterOperator = 
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'starts_with'
  | 'ends_with'
  | 'greater_than'
  | 'less_than'
  | 'between'
  | 'in'
  | 'not_in'

export type ReportStatus = 
  | 'draft'
  | 'scheduled'
  | 'generating'
  | 'completed'
  | 'failed'
  | 'cancelled'

// Date Range
export interface DateRange {
  start: number
  end: number
  timezone?: string
  preset?: DateRangePreset
}

export type DateRangePreset = 
  | 'today'
  | 'yesterday'
  | 'last_7_days'
  | 'last_30_days'
  | 'last_90_days'
  | 'this_month'
  | 'last_month'
  | 'this_quarter'
  | 'last_quarter'
  | 'this_year'
  | 'last_year'
  | 'custom'

// Dashboard and Visualization
export interface Dashboard {
  id: string
  name: string
  description: string
  widgets: Widget[]
  layout: DashboardLayout
  shared: boolean
  owner: string
  permissions: DashboardPermission[]
  createdAt: number
  updatedAt: number
}

export interface Widget {
  id: string
  type: WidgetType
  title: string
  config: WidgetConfig
  position: WidgetPosition
  dataSource: DataSource
  refreshInterval?: number
}

export type WidgetType = 
  | 'metric'
  | 'chart'
  | 'table'
  | 'map'
  | 'heatmap'
  | 'funnel'
  | 'gauge'
  | 'progress'
  | 'text'
  | 'iframe'

export interface WidgetConfig {
  chartType?: ChartType
  aggregation?: AggregationType
  groupBy?: string[]
  filters?: ReportFilter[]
  limit?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
  colorScheme?: string
  showLegend?: boolean
  showAxis?: boolean
  [key: string]: any
}

export type ChartType = 
  | 'line'
  | 'area'
  | 'bar'
  | 'column'
  | 'pie'
  | 'donut'
  | 'scatter'
  | 'bubble'
  | 'radar'
  | 'polar'
  | 'treemap'
  | 'sankey'

export type AggregationType = 
  | 'sum'
  | 'average'
  | 'count'
  | 'distinct'
  | 'min'
  | 'max'
  | 'median'
  | 'percentile'

export interface WidgetPosition {
  x: number
  y: number
  width: number
  height: number
}

export interface DataSource {
  type: 'api' | 'websocket' | 'static'
  url?: string
  query?: string
  params?: Record<string, any>
  headers?: Record<string, string>
  refreshRate?: number
}

export interface DashboardLayout {
  columns: number
  rowHeight: number
  margin: [number, number]
  padding: [number, number]
  responsive: boolean
}

export interface DashboardPermission {
  userId: string
  role: 'viewer' | 'editor' | 'admin'
  granted: number
  grantedBy: string
}

// Real-time Analytics
export interface RealTimeMetrics {
  activeUsers: number
  pageViews: number
  events: AnalyticsEvent[]
  performance: EnhancedPerformanceMetrics
  errors: ErrorEvent[]
  timestamp: number
}

export interface ErrorEvent {
  id: string
  type: 'javascript' | 'network' | 'security' | 'performance'
  message: string
  stack?: string
  url: string
  line?: number
  column?: number
  userId?: string
  sessionId: string
  timestamp: number
  severity: 'low' | 'medium' | 'high' | 'critical'
  resolved?: boolean
  tags: string[]
}

// Export utilities
export interface ExportOptions {
  format: ReportFormat
  includeCharts: boolean
  includeData: boolean
  compressed: boolean
  password?: string
  watermark?: string
  branding: boolean
}

export interface ExportResult {
  id: string
  filename: string
  size: number
  downloadUrl: string
  expiresAt: number
  password?: boolean
}
