// UI and component types

export interface ComponentProps {
  className?: string
  children?: React.ReactNode
  testId?: string
}

// Layout types
export interface LayoutProps extends ComponentProps {
  sidebar?: boolean
  header?: boolean
  footer?: boolean
  maxWidth?: string
}

export interface SidebarProps extends ComponentProps {
  open: boolean
  onToggle: () => void
  variant?: 'permanent' | 'temporary' | 'persistent'
  width?: number
}

export interface HeaderProps extends ComponentProps {
  title?: string
  actions?: React.ReactNode
  breadcrumbs?: BreadcrumbItem[]
}

export interface BreadcrumbItem {
  label: string
  href?: string
  active?: boolean
}

// Form types
export interface FormProps extends ComponentProps {
  onSubmit: (data: any) => void | Promise<void>
  loading?: boolean
  disabled?: boolean
  validationSchema?: any
  initialValues?: any
}

export interface InputProps extends ComponentProps {
  type?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url' | 'search'
  value?: string | number
  defaultValue?: string | number
  placeholder?: string
  disabled?: boolean
  readOnly?: boolean
  required?: boolean
  autoFocus?: boolean
  onChange?: (value: string) => void
  onBlur?: () => void
  onFocus?: () => void
  error?: string
  label?: string
  helpText?: string
  prefix?: React.ReactNode
  suffix?: React.ReactNode
  size?: 'sm' | 'md' | 'lg'
  variant?: 'outlined' | 'filled' | 'underlined'
}

export interface SelectProps extends ComponentProps {
  value?: string | string[]
  defaultValue?: string | string[]
  placeholder?: string
  disabled?: boolean
  required?: boolean
  multiple?: boolean
  searchable?: boolean
  clearable?: boolean
  options: SelectOption[]
  onChange?: (value: string | string[]) => void
  onSearch?: (query: string) => void
  error?: string
  label?: string
  helpText?: string
  size?: 'sm' | 'md' | 'lg'
  maxHeight?: number
}

export interface SelectOption {
  value: string
  label: string
  disabled?: boolean
  group?: string
  icon?: React.ReactNode
}

export interface ButtonProps extends ComponentProps {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'link'
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
  loading?: boolean
  icon?: React.ReactNode
  iconPosition?: 'left' | 'right'
  fullWidth?: boolean
  onClick?: () => void
  type?: 'button' | 'submit' | 'reset'
  href?: string
  target?: string
}

// Data display types
export interface TableProps<T = any> extends ComponentProps {
  data: T[]
  columns: TableColumn<T>[]
  loading?: boolean
  pagination?: PaginationConfig
  sorting?: SortConfig
  filtering?: FilterConfig
  selection?: SelectionConfig<T>
  onRowClick?: (row: T) => void
  emptyState?: React.ReactNode
  sticky?: boolean
  striped?: boolean
  bordered?: boolean
  dense?: boolean
}

export interface TableColumn<T = any> {
  key: string
  title: string
  dataIndex?: keyof T
  render?: (value: any, record: T, index: number) => React.ReactNode
  sortable?: boolean
  filterable?: boolean
  width?: number | string
  align?: 'left' | 'center' | 'right'
  fixed?: 'left' | 'right'
  ellipsis?: boolean
}

export interface PaginationConfig {
  current: number
  pageSize: number
  total: number
  showSizeChanger?: boolean
  showQuickJumper?: boolean
  showTotal?: (total: number, range: [number, number]) => string
  onChange?: (page: number, pageSize: number) => void
}

export interface SortConfig {
  field?: string
  order?: 'asc' | 'desc'
  onChange?: (field: string, order: 'asc' | 'desc' | null) => void
}

export interface FilterConfig {
  filters: Record<string, any>
  onChange?: (filters: Record<string, any>) => void
}

export interface SelectionConfig<T = any> {
  selectedRowKeys: string[]
  onChange?: (selectedRowKeys: string[], selectedRows: T[]) => void
  getCheckboxProps?: (record: T) => { disabled?: boolean }
  type?: 'checkbox' | 'radio'
}

// Card and panel types
export interface CardProps extends ComponentProps {
  title?: string
  subtitle?: string
  actions?: React.ReactNode
  footer?: React.ReactNode
  bordered?: boolean
  hoverable?: boolean
  loading?: boolean
  size?: 'sm' | 'md' | 'lg'
}

export interface PanelProps extends ComponentProps {
  title: string
  collapsible?: boolean
  collapsed?: boolean
  onCollapse?: (collapsed: boolean) => void
  actions?: React.ReactNode
  size?: 'sm' | 'md' | 'lg'
}

// Modal and dialog types
export interface ModalProps extends ComponentProps {
  open: boolean
  onClose: () => void
  title?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  centered?: boolean
  closable?: boolean
  maskClosable?: boolean
  keyboard?: boolean
  footer?: React.ReactNode
  loading?: boolean
}

export interface DialogProps extends ComponentProps {
  open: boolean
  onClose: () => void
  onConfirm?: () => void
  title: string
  content: string
  type?: 'info' | 'success' | 'warning' | 'error' | 'confirm'
  confirmText?: string
  cancelText?: string
  loading?: boolean
}

// Navigation types
export interface TabsProps extends ComponentProps {
  activeKey: string
  onChange: (key: string) => void
  type?: 'line' | 'card' | 'pills'
  size?: 'sm' | 'md' | 'lg'
  centered?: boolean
  items: TabItem[]
}

export interface TabItem {
  key: string
  label: string
  content: React.ReactNode
  disabled?: boolean
  closable?: boolean
  icon?: React.ReactNode
}

export interface MenuProps extends ComponentProps {
  items: MenuItem[]
  mode?: 'horizontal' | 'vertical' | 'inline'
  theme?: 'light' | 'dark'
  selectedKeys?: string[]
  openKeys?: string[]
  onSelect?: (key: string) => void
  onOpenChange?: (openKeys: string[]) => void
  collapsed?: boolean
}

export interface MenuItem {
  key: string
  label: string
  icon?: React.ReactNode
  href?: string
  disabled?: boolean
  danger?: boolean
  children?: MenuItem[]
  onClick?: () => void
}

// Feedback types
export interface AlertProps extends ComponentProps {
  type: 'info' | 'success' | 'warning' | 'error'
  message: string
  description?: string
  closable?: boolean
  onClose?: () => void
  showIcon?: boolean
  banner?: boolean
  actions?: React.ReactNode
}

export interface ProgressProps extends ComponentProps {
  percent: number
  type?: 'line' | 'circle' | 'dashboard'
  status?: 'normal' | 'success' | 'exception' | 'active'
  size?: 'sm' | 'md' | 'lg'
  showInfo?: boolean
  format?: (percent?: number) => React.ReactNode
  strokeColor?: string
  trailColor?: string
  strokeWidth?: number
}

export interface SpinnerProps extends ComponentProps {
  size?: 'sm' | 'md' | 'lg'
  tip?: string
  spinning?: boolean
  delay?: number
  indicator?: React.ReactNode
}

// Chart and visualization types
export interface ChartProps extends ComponentProps {
  data: any[]
  type: 'line' | 'bar' | 'pie' | 'area' | 'scatter' | 'donut'
  xField?: string
  yField?: string
  colorField?: string
  seriesField?: string
  config?: any
  height?: number
  loading?: boolean
  error?: string
}

// Theme types
export interface ThemeColors {
  primary: string
  secondary: string
  success: string
  warning: string
  error: string
  info: string
  background: string
  surface: string
  text: string
  textSecondary: string
  border: string
  divider: string
}

export interface ThemeSpacing {
  xs: number
  sm: number
  md: number
  lg: number
  xl: number
  xxl: number
}

export interface ThemeTypography {
  fontFamily: string
  fontSize: {
    xs: number
    sm: number
    md: number
    lg: number
    xl: number
    xxl: number
  }
  fontWeight: {
    light: number
    normal: number
    medium: number
    semibold: number
    bold: number
  }
  lineHeight: {
    tight: number
    normal: number
    relaxed: number
  }
}

export interface ThemeBreakpoints {
  xs: number
  sm: number
  md: number
  lg: number
  xl: number
  xxl: number
}

export interface Theme {
  colors: ThemeColors
  spacing: ThemeSpacing
  typography: ThemeTypography
  breakpoints: ThemeBreakpoints
  shadows: string[]
  borderRadius: {
    sm: number
    md: number
    lg: number
    full: number
  }
  transitions: {
    duration: {
      fast: string
      normal: string
      slow: string
    }
    easing: {
      ease: string
      easeIn: string
      easeOut: string
      easeInOut: string
    }
  }
}