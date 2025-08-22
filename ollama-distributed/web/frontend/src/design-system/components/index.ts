/**
 * Design System Components Index
 * Centralized export of all design system components
 */

// Base components
export { default as Button, ButtonGroup, IconButton, ToggleButton } from './Button/Button'
export type { ButtonProps, ButtonGroupProps, IconButtonProps, ToggleButtonProps } from './Button/Button'

export { default as Input, Textarea } from './Input/Input'
export type { InputProps, TextareaProps } from './Input/Input'

export { default as Card } from './Card/Card'
export type { CardProps, CardHeaderProps, CardContentProps, CardFooterProps, CardTitleProps, CardDescriptionProps, CardImageProps, CardActionsProps } from './Card/Card'

export { default as Badge, BadgeGroup, StatusBadge, NotificationBadge } from './Badge/Badge'
export type { BadgeProps, BadgeGroupProps, StatusBadgeProps, NotificationBadgeProps } from './Badge/Badge'

export { default as Alert, ToastAlert, BannerAlert } from './Alert/Alert'
export type { AlertProps, AlertTitleProps, AlertDescriptionProps, AlertActionsProps, ToastAlertProps, BannerAlertProps } from './Alert/Alert'

// Loading components
export { Spinner, SpinnerOverlay, LoadingButton, Skeleton } from './Spinner'
export type { SpinnerProps, SpinnerOverlayProps, LoadingButtonProps, SkeletonProps } from './Spinner'

// Layout components
export { default as Layout } from './Layout/Layout'
export type { ContainerProps, GridProps, FlexProps, StackProps, BoxProps, SpacerProps, DividerProps, CenterProps, AspectRatioProps, MasonryProps } from './Layout/Layout'

// Data display components
export { 
  default as Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  TableFooter,
  TableCaption,
  TablePagination,
  TableFilter,
  EmptyState
} from './Table/Table'
export type { 
  TableProps,
  TableHeaderProps,
  TableBodyProps,
  TableRowProps,
  TableHeadProps,
  TableCellProps,
  TableFooterProps,
  TableCaptionProps,
  TablePaginationProps,
  TableFilterProps,
  EmptyStateProps
} from './Table/Table'

// Component utilities
export const componentUtils = {
  Button: () => import('./Button/Button'),
  Input: () => import('./Input/Input'),
  Card: () => import('./Card/Card'),
  Badge: () => import('./Badge/Badge'),
  Alert: () => import('./Alert/Alert'),
  Spinner: () => import('./Spinner'),
  Layout: () => import('./Layout/Layout'),
  Table: () => import('./Table/Table')
} as const

export default {
  componentUtils
}