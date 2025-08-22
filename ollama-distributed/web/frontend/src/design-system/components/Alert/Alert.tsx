import React, { forwardRef, HTMLAttributes, useState } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { AlertCircle, CheckCircle, Info, AlertTriangle, X } from 'lucide-react'
import { cn } from '@/utils/cn'

// Alert variants
const alertVariants = cva(
  [
    'relative w-full rounded-lg border p-4',
    'transition-all duration-200'
  ],
  {
    variants: {
      variant: {
        default: [
          'bg-background text-foreground border-border'
        ],
        info: [
          'bg-info/10 text-info-foreground border-info/20',
          'dark:bg-info/10 dark:text-info dark:border-info/30'
        ],
        success: [
          'bg-success/10 text-success-foreground border-success/20',
          'dark:bg-success/10 dark:text-success dark:border-success/30'
        ],
        warning: [
          'bg-warning/10 text-warning-foreground border-warning/20',
          'dark:bg-warning/10 dark:text-warning dark:border-warning/30'
        ],
        destructive: [
          'bg-destructive/10 text-destructive-foreground border-destructive/20',
          'dark:bg-destructive/10 dark:text-destructive dark:border-destructive/30'
        ]
      },
      size: {
        sm: 'p-3 text-sm',
        md: 'p-4 text-sm',
        lg: 'p-6 text-base'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'md'
    }
  }
)

export interface AlertProps
  extends HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof alertVariants> {
  /** Alert title */
  title?: string
  /** Whether the alert is dismissible */
  dismissible?: boolean
  /** Callback when alert is dismissed */
  onDismiss?: () => void
  /** Custom icon */
  icon?: React.ReactNode
  /** Whether to show default icon */
  showIcon?: boolean
  /** Actions to display in the alert */
  actions?: React.ReactNode
}

/**
 * Alert component for displaying important messages and notifications
 * 
 * Features:
 * - Multiple variants (default, info, success, warning, destructive)
 * - Size variants (sm, md, lg)
 * - Dismissible alerts
 * - Custom icons or default variant icons
 * - Action buttons support
 * - Full accessibility support
 * - Auto-dismiss capability
 */
export const Alert = forwardRef<HTMLDivElement, AlertProps>(
  (
    {
      className,
      variant,
      size,
      title,
      dismissible = false,
      onDismiss,
      icon,
      showIcon = true,
      actions,
      children,
      ...props
    },
    ref
  ) => {
    const [isVisible, setIsVisible] = useState(true)

    const handleDismiss = () => {
      setIsVisible(false)
      onDismiss?.()
    }

    const getDefaultIcon = () => {
      switch (variant) {
        case 'info':
          return <Info className="w-4 h-4" />
        case 'success':
          return <CheckCircle className="w-4 h-4" />
        case 'warning':
          return <AlertTriangle className="w-4 h-4" />
        case 'destructive':
          return <AlertCircle className="w-4 h-4" />
        default:
          return <Info className="w-4 h-4" />
      }
    }

    const displayIcon = icon || (showIcon ? getDefaultIcon() : null)

    if (!isVisible) {
      return null
    }

    return (
      <div
        ref={ref}
        role="alert"
        className={cn(alertVariants({ variant, size }), className)}
        {...props}
      >
        <div className="flex">
          {/* Icon */}
          {displayIcon && (
            <div className="flex-shrink-0 mr-3">
              <span 
                className={cn(
                  'flex items-center justify-center',
                  variant === 'info' && 'text-info',
                  variant === 'success' && 'text-success',
                  variant === 'warning' && 'text-warning',
                  variant === 'destructive' && 'text-destructive'
                )}
                aria-hidden="true"
              >
                {displayIcon}
              </span>
            </div>
          )}

          {/* Content */}
          <div className="flex-1 min-w-0">
            {title && (
              <h3 className={cn(
                'font-medium mb-1',
                size === 'sm' ? 'text-sm' : 'text-base'
              )}>
                {title}
              </h3>
            )}
            
            {children && (
              <div className={cn(
                'text-sm leading-relaxed',
                variant === 'info' && 'text-info-foreground/90',
                variant === 'success' && 'text-success-foreground/90',
                variant === 'warning' && 'text-warning-foreground/90',
                variant === 'destructive' && 'text-destructive-foreground/90'
              )}>
                {children}
              </div>
            )}

            {/* Actions */}
            {actions && (
              <div className="mt-3 flex items-center space-x-2">
                {actions}
              </div>
            )}
          </div>

          {/* Dismiss button */}
          {dismissible && (
            <div className="flex-shrink-0 ml-3">
              <button
                type="button"
                onClick={handleDismiss}
                className={cn(
                  'inline-flex rounded-md p-1.5 transition-colors',
                  'hover:bg-black/5 dark:hover:bg-white/5',
                  'focus:outline-none focus:ring-2 focus:ring-offset-2',
                  variant === 'info' && 'text-info focus:ring-info',
                  variant === 'success' && 'text-success focus:ring-success',
                  variant === 'warning' && 'text-warning focus:ring-warning',
                  variant === 'destructive' && 'text-destructive focus:ring-destructive',
                  variant === 'default' && 'text-foreground focus:ring-primary'
                )}
                aria-label="Dismiss alert"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          )}
        </div>
      </div>
    )
  }
)

Alert.displayName = 'Alert'

// Alert Title component
export interface AlertTitleProps extends HTMLAttributes<HTMLHeadingElement> {
  level?: 1 | 2 | 3 | 4 | 5 | 6
}

export const AlertTitle = forwardRef<HTMLHeadingElement, AlertTitleProps>(
  ({ className, level = 3, ...props }, ref) => {
    const Component = `h${level}` as keyof JSX.IntrinsicElements

    return (
      <Component
        ref={ref as any}
        className={cn('font-medium leading-none tracking-tight mb-1', className)}
        {...props}
      />
    )
  }
)

AlertTitle.displayName = 'AlertTitle'

// Alert Description component
export interface AlertDescriptionProps extends HTMLAttributes<HTMLDivElement> {}

export const AlertDescription = forwardRef<HTMLDivElement, AlertDescriptionProps>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('text-sm leading-relaxed [&_p]:leading-relaxed', className)}
      {...props}
    />
  )
)

AlertDescription.displayName = 'AlertDescription'

// Alert Actions component
export interface AlertActionsProps extends HTMLAttributes<HTMLDivElement> {
  /** Spacing between actions */
  spacing?: 'sm' | 'md' | 'lg'
  /** Actions alignment */
  align?: 'start' | 'center' | 'end'
}

export const AlertActions = forwardRef<HTMLDivElement, AlertActionsProps>(
  ({ className, spacing = 'md', align = 'start', ...props }, ref) => {
    const spacingClasses = {
      sm: 'space-x-2',
      md: 'space-x-3',
      lg: 'space-x-4'
    }

    const alignClasses = {
      start: 'justify-start',
      center: 'justify-center',
      end: 'justify-end'
    }

    return (
      <div
        ref={ref}
        className={cn(
          'flex items-center mt-3',
          spacingClasses[spacing],
          alignClasses[align],
          className
        )}
        {...props}
      />
    )
  }
)

AlertActions.displayName = 'AlertActions'

// Toast Alert for notifications
export interface ToastAlertProps extends Omit<AlertProps, 'dismissible'> {
  /** Auto-dismiss timeout in milliseconds */
  autoHideDuration?: number
  /** Position of the toast */
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | 'top-center' | 'bottom-center'
}

export const ToastAlert = forwardRef<HTMLDivElement, ToastAlertProps>(
  (
    {
      className,
      autoHideDuration = 5000,
      position = 'top-right',
      onDismiss,
      ...props
    },
    ref
  ) => {
    const [isVisible, setIsVisible] = useState(true)

    React.useEffect(() => {
      if (autoHideDuration > 0) {
        const timer = setTimeout(() => {
          setIsVisible(false)
          onDismiss?.()
        }, autoHideDuration)

        return () => clearTimeout(timer)
      }
    }, [autoHideDuration, onDismiss])

    const positionClasses = {
      'top-right': 'fixed top-4 right-4 z-50',
      'top-left': 'fixed top-4 left-4 z-50',
      'bottom-right': 'fixed bottom-4 right-4 z-50',
      'bottom-left': 'fixed bottom-4 left-4 z-50',
      'top-center': 'fixed top-4 left-1/2 transform -translate-x-1/2 z-50',
      'bottom-center': 'fixed bottom-4 left-1/2 transform -translate-x-1/2 z-50'
    }

    if (!isVisible) {
      return null
    }

    return (
      <Alert
        ref={ref}
        dismissible
        onDismiss={() => {
          setIsVisible(false)
          onDismiss?.()
        }}
        className={cn(
          positionClasses[position],
          'max-w-sm shadow-lg animate-in slide-in-from-top-2',
          className
        )}
        {...props}
      />
    )
  }
)

ToastAlert.displayName = 'ToastAlert'

// Banner Alert for page-level notifications
export interface BannerAlertProps extends AlertProps {
  /** Whether the banner is sticky */
  sticky?: boolean
}

export const BannerAlert = forwardRef<HTMLDivElement, BannerAlertProps>(
  ({ className, sticky = false, ...props }, ref) => (
    <Alert
      ref={ref}
      className={cn(
        'rounded-none border-x-0 border-t-0',
        sticky && 'sticky top-0 z-40',
        className
      )}
      {...props}
    />
  )
)

BannerAlert.displayName = 'BannerAlert'

// Compound component exports
Alert.Title = AlertTitle
Alert.Description = AlertDescription
Alert.Actions = AlertActions

export default Alert