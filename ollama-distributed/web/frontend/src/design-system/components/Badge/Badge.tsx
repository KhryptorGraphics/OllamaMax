import React, { forwardRef, HTMLAttributes } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { X } from 'lucide-react'
import { cn } from '@/utils/cn'

// Badge variants
const badgeVariants = cva(
  [
    'inline-flex items-center rounded-full text-xs font-medium',
    'transition-colors duration-200',
    'focus:outline-none focus:ring-2 focus:ring-offset-2'
  ],
  {
    variants: {
      variant: {
        default: [
          'bg-primary text-primary-foreground',
          'hover:bg-primary/80'
        ],
        secondary: [
          'bg-secondary text-secondary-foreground',
          'hover:bg-secondary/80'
        ],
        success: [
          'bg-success text-success-foreground',
          'hover:bg-success/80'
        ],
        warning: [
          'bg-warning text-warning-foreground',
          'hover:bg-warning/80'
        ],
        destructive: [
          'bg-destructive text-destructive-foreground',
          'hover:bg-destructive/80'
        ],
        outline: [
          'border border-input bg-background text-foreground',
          'hover:bg-accent hover:text-accent-foreground'
        ],
        ghost: [
          'text-foreground',
          'hover:bg-accent hover:text-accent-foreground'
        ]
      },
      size: {
        sm: 'px-2 py-0.5 text-xs',
        md: 'px-2.5 py-1 text-xs',
        lg: 'px-3 py-1.5 text-sm'
      },
      interactive: {
        true: 'cursor-pointer',
        false: 'pointer-events-none'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'md',
      interactive: false
    }
  }
)

export interface BadgeProps
  extends HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {
  /** Whether the badge is removable */
  removable?: boolean
  /** Callback when badge is removed */
  onRemove?: () => void
  /** Icon to display before the text */
  icon?: React.ReactNode
  /** Dot indicator variant */
  dot?: boolean
  /** Pulse animation for notifications */
  pulse?: boolean
}

/**
 * Badge component for labels, status indicators, and notifications
 * 
 * Features:
 * - Multiple variants (default, secondary, success, warning, destructive, outline, ghost)
 * - Size variants (sm, md, lg)
 * - Removable badges with close button
 * - Icon support
 * - Dot indicator variant
 * - Pulse animation for notifications
 * - Full accessibility support
 */
export const Badge = forwardRef<HTMLDivElement, BadgeProps>(
  (
    {
      className,
      variant,
      size,
      interactive,
      removable = false,
      onRemove,
      icon,
      dot = false,
      pulse = false,
      children,
      onClick,
      ...props
    },
    ref
  ) => {
    const isInteractive = interactive || removable || onClick

    const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
      if (onClick) {
        onClick(event)
      }
    }

    const handleRemove = (event: React.MouseEvent<HTMLButtonElement>) => {
      event.stopPropagation()
      onRemove?.()
    }

    const handleKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
      if (isInteractive && (event.key === 'Enter' || event.key === ' ')) {
        event.preventDefault()
        if (onClick) {
          onClick(event as any)
        }
      }
    }

    if (dot) {
      return (
        <span
          ref={ref}
          className={cn(
            'inline-flex items-center gap-1.5',
            className
          )}
          {...props}
        >
          <span
            className={cn(
              'w-2 h-2 rounded-full',
              variant === 'default' && 'bg-primary',
              variant === 'secondary' && 'bg-secondary',
              variant === 'success' && 'bg-success',
              variant === 'warning' && 'bg-warning',
              variant === 'destructive' && 'bg-destructive',
              variant === 'outline' && 'bg-foreground',
              variant === 'ghost' && 'bg-foreground',
              pulse && 'animate-pulse'
            )}
            aria-hidden="true"
          />
          {children && (
            <span className="text-sm text-foreground">
              {children}
            </span>
          )}
        </span>
      )
    }

    return (
      <div
        ref={ref}
        className={cn(
          badgeVariants({ variant, size, interactive: isInteractive }),
          pulse && 'animate-pulse',
          className
        )}
        onClick={isInteractive ? handleClick : undefined}
        onKeyDown={isInteractive ? handleKeyDown : undefined}
        tabIndex={isInteractive ? 0 : undefined}
        role={isInteractive ? 'button' : undefined}
        {...props}
      >
        {/* Icon */}
        {icon && (
          <span 
            className={cn(
              'flex items-center',
              size === 'sm' ? 'w-3 h-3' :
              size === 'md' ? 'w-3.5 h-3.5' :
              size === 'lg' ? 'w-4 h-4' : 'w-3.5 h-3.5',
              children && 'mr-1'
            )}
            aria-hidden="true"
          >
            {icon}
          </span>
        )}

        {/* Content */}
        {children && (
          <span className="truncate">
            {children}
          </span>
        )}

        {/* Remove button */}
        {removable && (
          <button
            type="button"
            onClick={handleRemove}
            className={cn(
              'ml-1 flex-shrink-0 rounded-full p-0.5',
              'hover:bg-black/10 dark:hover:bg-white/10',
              'focus:outline-none focus:ring-1 focus:ring-offset-1 focus:ring-current',
              'transition-colors duration-150'
            )}
            aria-label="Remove badge"
          >
            <X 
              className={cn(
                size === 'sm' ? 'w-2.5 h-2.5' :
                size === 'md' ? 'w-3 h-3' :
                size === 'lg' ? 'w-3.5 h-3.5' : 'w-3 h-3'
              )}
            />
          </button>
        )}
      </div>
    )
  }
)

Badge.displayName = 'Badge'

// Badge Group for multiple badges
export interface BadgeGroupProps {
  children: React.ReactNode
  /** Maximum number of badges to show before truncating */
  max?: number
  /** Spacing between badges */
  spacing?: 'sm' | 'md' | 'lg'
  /** Orientation */
  orientation?: 'horizontal' | 'vertical'
  /** Custom className */
  className?: string
}

export const BadgeGroup: React.FC<BadgeGroupProps> = ({
  children,
  max,
  spacing = 'sm',
  orientation = 'horizontal',
  className
}) => {
  const spacingClasses = {
    sm: orientation === 'horizontal' ? 'gap-1' : 'gap-1',
    md: orientation === 'horizontal' ? 'gap-2' : 'gap-2',
    lg: orientation === 'horizontal' ? 'gap-3' : 'gap-3'
  }

  const badges = React.Children.toArray(children)
  const visibleBadges = max ? badges.slice(0, max) : badges
  const hiddenCount = max && badges.length > max ? badges.length - max : 0

  return (
    <div
      className={cn(
        'flex flex-wrap items-center',
        orientation === 'vertical' && 'flex-col items-start',
        spacingClasses[spacing],
        className
      )}
    >
      {visibleBadges}
      {hiddenCount > 0 && (
        <Badge variant="outline" size="sm">
          +{hiddenCount}
        </Badge>
      )}
    </div>
  )
}

BadgeGroup.displayName = 'BadgeGroup'

// Status Badge for common status indicators
export interface StatusBadgeProps extends Omit<BadgeProps, 'variant'> {
  status: 'online' | 'offline' | 'busy' | 'away' | 'active' | 'inactive' | 'pending' | 'approved' | 'rejected'
}

export const StatusBadge = forwardRef<HTMLDivElement, StatusBadgeProps>(
  ({ status, ...props }, ref) => {
    const statusConfig = {
      online: { variant: 'success' as const, children: 'Online' },
      offline: { variant: 'secondary' as const, children: 'Offline' },
      busy: { variant: 'destructive' as const, children: 'Busy' },
      away: { variant: 'warning' as const, children: 'Away' },
      active: { variant: 'success' as const, children: 'Active' },
      inactive: { variant: 'secondary' as const, children: 'Inactive' },
      pending: { variant: 'warning' as const, children: 'Pending' },
      approved: { variant: 'success' as const, children: 'Approved' },
      rejected: { variant: 'destructive' as const, children: 'Rejected' }
    }

    const config = statusConfig[status]

    return (
      <Badge
        ref={ref}
        variant={config.variant}
        {...props}
      >
        {props.children || config.children}
      </Badge>
    )
  }
)

StatusBadge.displayName = 'StatusBadge'

// Notification Badge for counts
export interface NotificationBadgeProps extends Omit<BadgeProps, 'children'> {
  count: number
  /** Maximum count to display before showing "99+" */
  max?: number
  /** Show dot instead of count when count is 0 */
  showZero?: boolean
}

export const NotificationBadge = forwardRef<HTMLDivElement, NotificationBadgeProps>(
  ({ count, max = 99, showZero = false, ...props }, ref) => {
    if (count === 0 && !showZero) {
      return null
    }

    const displayCount = count > max ? `${max}+` : count.toString()

    return (
      <Badge
        ref={ref}
        variant="destructive"
        size="sm"
        {...props}
      >
        {count === 0 ? '' : displayCount}
      </Badge>
    )
  }
)

NotificationBadge.displayName = 'NotificationBadge'

export default Badge