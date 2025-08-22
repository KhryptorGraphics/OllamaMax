import React, { forwardRef, HTMLAttributes } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/utils/cn'

/**
 * Spinner variants using class-variance-authority with design tokens
 */
const spinnerVariants = cva(
  // Base styles
  'inline-flex items-center justify-center',
  {
    variants: {
      size: {
        xs: 'w-3 h-3',
        sm: 'w-4 h-4',
        md: 'w-6 h-6',
        lg: 'w-8 h-8',
        xl: 'w-12 h-12'
      },
      variant: {
        primary: 'text-primary-500',
        secondary: 'text-secondary-500',
        success: 'text-success-500',
        warning: 'text-warning-500',
        danger: 'text-danger-500',
        info: 'text-info-500',
        neutral: 'text-neutral-500',
        current: 'text-current',
        white: 'text-white',
        black: 'text-black'
      },
      speed: {
        slow: '',
        normal: '',
        fast: ''
      },
      type: {
        spin: '',
        pulse: '',
        dots: '',
        bars: '',
        ring: '',
        ripple: ''
      }
    },
    defaultVariants: {
      size: 'md',
      variant: 'primary',
      speed: 'normal',
      type: 'spin'
    }
  }
)

/**
 * Spinner container variants for different contexts
 */
const spinnerContainerVariants = cva(
  'inline-flex flex-col items-center justify-center gap-2',
  {
    variants: {
      overlay: {
        true: 'fixed inset-0 z-50 bg-background/80 backdrop-blur-sm',
        false: ''
      },
      centered: {
        true: 'absolute inset-0',
        false: ''
      },
      inline: {
        true: 'inline-flex flex-row items-center',
        false: ''
      }
    },
    defaultVariants: {
      overlay: false,
      centered: false,
      inline: false
    }
  }
)

export interface SpinnerProps 
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'>,
    VariantProps<typeof spinnerVariants> {
  /** Loading text to display with spinner */
  label?: string
  /** Position of label relative to spinner */
  labelPosition?: 'top' | 'bottom' | 'left' | 'right'
  /** Whether to show as overlay */
  overlay?: boolean
  /** Whether to center in parent container */
  centered?: boolean
  /** Whether to display inline with content */
  inline?: boolean
  /** Custom spinner content for advanced animations */
  customSpinner?: React.ReactNode
  /** Screen reader text for accessibility */
  screenReaderText?: string
  /** Whether the spinner is visible */
  visible?: boolean
}

/**
 * Spinner component with multiple animation types and comprehensive features
 * 
 * Features:
 * - Multiple size variants (xs, sm, md, lg, xl)
 * - Different animation types (spin, pulse, dots, bars, ring, ripple)
 * - Color variants using design tokens
 * - Speed control (slow, normal, fast)
 * - Loading text with flexible positioning
 * - Overlay and centered display modes
 * - Full accessibility support with ARIA attributes
 * - Smooth enter/exit transitions
 * - Performance optimized animations
 * 
 * @example
 * ```tsx
 * // Basic spinner
 * <Spinner />
 * 
 * // With loading text
 * <Spinner label="Loading..." />
 * 
 * // Full screen overlay
 * <Spinner overlay label="Please wait..." />
 * 
 * // Custom size and color
 * <Spinner size="lg" variant="success" />
 * 
 * // Different animation types
 * <Spinner type="dots" />
 * <Spinner type="pulse" />
 * ```
 */
export const Spinner = forwardRef<HTMLDivElement, SpinnerProps>(
  (
    {
      className,
      size = 'md',
      variant = 'primary',
      speed = 'normal',
      type = 'spin',
      label,
      labelPosition = 'bottom',
      overlay = false,
      centered = false,
      inline = false,
      customSpinner,
      screenReaderText = 'Loading',
      visible = true,
      ...props
    },
    ref
  ) => {
    // Don't render if not visible
    if (!visible) return null

    const speedDuration = {
      slow: '1.5s',
      normal: '1s',
      fast: '0.5s'
    }

    const renderSpinner = () => {
      if (customSpinner) return customSpinner

      switch (type) {
        case 'spin':
          return (
            <svg
              className={cn(spinnerVariants({ size, variant }), 'animate-spin')}
              style={{ animationDuration: speedDuration[speed] }}
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
          )

        case 'pulse':
          return (
            <div
              className={cn(
                spinnerVariants({ size, variant }),
                'rounded-full bg-current animate-pulse'
              )}
              style={{ animationDuration: speedDuration[speed] }}
              aria-hidden="true"
            />
          )

        case 'dots':
          return (
            <div className="flex space-x-1" aria-hidden="true">
              {[0, 1, 2].map((i) => (
                <div
                  key={i}
                  className={cn(
                    'rounded-full bg-current',
                    size === 'xs' ? 'w-1 h-1' :
                    size === 'sm' ? 'w-1.5 h-1.5' :
                    size === 'md' ? 'w-2 h-2' :
                    size === 'lg' ? 'w-2.5 h-2.5' :
                    'w-3 h-3',
                    variant === 'primary' ? 'bg-primary-500' :
                    variant === 'secondary' ? 'bg-secondary-500' :
                    variant === 'success' ? 'bg-success-500' :
                    variant === 'warning' ? 'bg-warning-500' :
                    variant === 'danger' ? 'bg-danger-500' :
                    variant === 'info' ? 'bg-info-500' :
                    variant === 'neutral' ? 'bg-neutral-500' :
                    variant === 'current' ? 'bg-current' :
                    variant === 'white' ? 'bg-white' :
                    'bg-black'
                  )}
                  style={{
                    animation: `bounce ${speedDuration[speed]} infinite`,
                    animationDelay: `${i * 0.1}s`
                  }}
                />
              ))}
            </div>
          )

        case 'bars':
          return (
            <div className="flex space-x-1" aria-hidden="true">
              {[0, 1, 2, 3].map((i) => (
                <div
                  key={i}
                  className={cn(
                    'bg-current',
                    size === 'xs' ? 'w-0.5 h-3' :
                    size === 'sm' ? 'w-0.5 h-4' :
                    size === 'md' ? 'w-1 h-6' :
                    size === 'lg' ? 'w-1.5 h-8' :
                    'w-2 h-12',
                    variant === 'primary' ? 'bg-primary-500' :
                    variant === 'secondary' ? 'bg-secondary-500' :
                    variant === 'success' ? 'bg-success-500' :
                    variant === 'warning' ? 'bg-warning-500' :
                    variant === 'danger' ? 'bg-danger-500' :
                    variant === 'info' ? 'bg-info-500' :
                    variant === 'neutral' ? 'bg-neutral-500' :
                    variant === 'current' ? 'bg-current' :
                    variant === 'white' ? 'bg-white' :
                    'bg-black'
                  )}
                  style={{
                    animation: `scale-y ${speedDuration[speed]} infinite ease-in-out`,
                    animationDelay: `${i * 0.1}s`,
                    transformOrigin: 'center'
                  }}
                />
              ))}
            </div>
          )

        case 'ring':
          return (
            <div
              className={cn(
                spinnerVariants({ size }),
                'relative'
              )}
              aria-hidden="true"
            >
              <div
                className={cn(
                  'absolute inset-0 rounded-full border-2',
                  variant === 'primary' ? 'border-primary-500' :
                  variant === 'secondary' ? 'border-secondary-500' :
                  variant === 'success' ? 'border-success-500' :
                  variant === 'warning' ? 'border-warning-500' :
                  variant === 'danger' ? 'border-danger-500' :
                  variant === 'info' ? 'border-info-500' :
                  variant === 'neutral' ? 'border-neutral-500' :
                  variant === 'current' ? 'border-current' :
                  variant === 'white' ? 'border-white' :
                  'border-black',
                  'opacity-25'
                )}
              />
              <div
                className={cn(
                  'absolute inset-0 rounded-full border-2 border-transparent',
                  variant === 'primary' ? 'border-t-primary-500' :
                  variant === 'secondary' ? 'border-t-secondary-500' :
                  variant === 'success' ? 'border-t-success-500' :
                  variant === 'warning' ? 'border-t-warning-500' :
                  variant === 'danger' ? 'border-t-danger-500' :
                  variant === 'info' ? 'border-t-info-500' :
                  variant === 'neutral' ? 'border-t-neutral-500' :
                  variant === 'current' ? 'border-t-current' :
                  variant === 'white' ? 'border-t-white' :
                  'border-t-black',
                  'animate-spin'
                )}
                style={{ animationDuration: speedDuration[speed] }}
              />
            </div>
          )

        case 'ripple':
          return (
            <div
              className={cn(
                spinnerVariants({ size }),
                'relative'
              )}
              aria-hidden="true"
            >
              {[0, 1].map((i) => (
                <div
                  key={i}
                  className={cn(
                    'absolute inset-0 rounded-full border-2',
                    variant === 'primary' ? 'border-primary-500' :
                    variant === 'secondary' ? 'border-secondary-500' :
                    variant === 'success' ? 'border-success-500' :
                    variant === 'warning' ? 'border-warning-500' :
                    variant === 'danger' ? 'border-danger-500' :
                    variant === 'info' ? 'border-info-500' :
                    variant === 'neutral' ? 'border-neutral-500' :
                    variant === 'current' ? 'border-current' :
                    variant === 'white' ? 'border-white' :
                    'border-black'
                  )}
                  style={{
                    animation: `ripple ${speedDuration[speed]} cubic-bezier(0, 0.2, 0.8, 1) infinite`,
                    animationDelay: `${i * 0.5}s`
                  }}
                />
              ))}
            </div>
          )

        default:
          return null
      }
    }

    const spinnerContent = (
      <>
        {renderSpinner()}
        {label && (
          <span 
            className={cn(
              'text-sm font-medium',
              variant === 'primary' ? 'text-primary-600' :
              variant === 'secondary' ? 'text-secondary-600' :
              variant === 'success' ? 'text-success-600' :
              variant === 'warning' ? 'text-warning-600' :
              variant === 'danger' ? 'text-danger-600' :
              variant === 'info' ? 'text-info-600' :
              variant === 'neutral' ? 'text-neutral-600' :
              variant === 'current' ? 'text-current' :
              variant === 'white' ? 'text-white' :
              'text-foreground'
            )}
          >
            {label}
          </span>
        )}
      </>
    )

    return (
      <div
        ref={ref}
        className={cn(
          spinnerContainerVariants({ overlay, centered, inline }),
          inline && labelPosition === 'left' && 'flex-row-reverse',
          inline && labelPosition === 'right' && 'flex-row',
          !inline && labelPosition === 'top' && 'flex-col-reverse',
          !inline && labelPosition === 'bottom' && 'flex-col',
          className
        )}
        role="status"
        aria-live="polite"
        aria-busy="true"
        {...props}
      >
        {spinnerContent}
        <span className="sr-only">{screenReaderText}</span>

        {/* Add required styles for custom animations */}
        <style jsx>{`
          @keyframes ripple {
            0% {
              transform: scale(0);
              opacity: 1;
            }
            100% {
              transform: scale(1);
              opacity: 0;
            }
          }
          
          @keyframes scale-y {
            0%, 80%, 100% {
              transform: scaleY(1);
            }
            40% {
              transform: scaleY(1.5);
            }
          }
        `}</style>
      </div>
    )
  }
)

Spinner.displayName = 'Spinner'

/**
 * SpinnerOverlay component for full-screen loading states
 */
export interface SpinnerOverlayProps extends Omit<SpinnerProps, 'overlay'> {
  /** Whether the overlay is visible */
  isOpen?: boolean
  /** Callback when overlay is clicked */
  onOverlayClick?: () => void
  /** Whether clicking overlay should close it */
  closeOnClick?: boolean
}

export const SpinnerOverlay = forwardRef<HTMLDivElement, SpinnerOverlayProps>(
  ({ isOpen = true, onOverlayClick, closeOnClick = false, ...props }, ref) => {
    if (!isOpen) return null

    const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
      if (e.target === e.currentTarget && closeOnClick) {
        onOverlayClick?.()
      }
    }

    return (
      <Spinner
        ref={ref}
        overlay
        onClick={handleOverlayClick}
        {...props}
      />
    )
  }
)

SpinnerOverlay.displayName = 'SpinnerOverlay'

/**
 * LoadingButton component that integrates spinner with button
 */
export interface LoadingButtonProps {
  /** Whether the button is loading */
  isLoading?: boolean
  /** Loading text to display */
  loadingText?: string
  /** Spinner props to customize loading indicator */
  spinnerProps?: Omit<SpinnerProps, 'label'>
  /** Button content */
  children: React.ReactNode
  /** Additional button props */
  className?: string
  onClick?: () => void
  disabled?: boolean
}

export const LoadingButton: React.FC<LoadingButtonProps> = ({
  isLoading = false,
  loadingText = 'Loading...',
  spinnerProps,
  children,
  className,
  onClick,
  disabled
}) => {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center gap-2 px-4 py-2',
        'bg-primary-500 text-white rounded-md',
        'hover:bg-primary-600 transition-colors',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        className
      )}
      onClick={onClick}
      disabled={disabled || isLoading}
      aria-busy={isLoading}
    >
      {isLoading ? (
        <>
          <Spinner
            size="sm"
            variant="white"
            inline
            {...spinnerProps}
          />
          <span>{loadingText}</span>
        </>
      ) : (
        children
      )}
    </button>
  )
}

LoadingButton.displayName = 'LoadingButton'

/**
 * Skeleton loader component for content placeholders
 */
export interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  /** Width of the skeleton */
  width?: string | number
  /** Height of the skeleton */
  height?: string | number
  /** Shape variant */
  variant?: 'text' | 'circular' | 'rectangular'
  /** Animation type */
  animation?: 'pulse' | 'wave' | 'none'
}

export const Skeleton: React.FC<SkeletonProps> = ({
  width,
  height,
  variant = 'text',
  animation = 'pulse',
  className,
  style,
  ...props
}) => {
  const baseClasses = 'bg-muted'
  const animationClasses = {
    pulse: 'animate-pulse',
    wave: 'animate-shimmer',
    none: ''
  }
  const variantClasses = {
    text: 'rounded',
    circular: 'rounded-full',
    rectangular: 'rounded-md'
  }

  return (
    <div
      className={cn(
        baseClasses,
        animationClasses[animation],
        variantClasses[variant],
        className
      )}
      style={{
        width: width || (variant === 'text' ? '100%' : undefined),
        height: height || (variant === 'text' ? '1em' : undefined),
        ...style
      }}
      aria-busy="true"
      aria-live="polite"
      role="status"
      {...props}
    >
      <span className="sr-only">Loading content</span>
      
      {/* Add shimmer animation styles */}
      {animation === 'wave' && (
        <style jsx>{`
          @keyframes shimmer {
            0% {
              background-position: -200% 0;
            }
            100% {
              background-position: 200% 0;
            }
          }
          
          .animate-shimmer {
            background: linear-gradient(
              90deg,
              var(--muted) 25%,
              var(--muted-foreground) 50%,
              var(--muted) 75%
            );
            background-size: 200% 100%;
            animation: shimmer 1.5s ease-in-out infinite;
          }
        `}</style>
      )}
    </div>
  )
}

Skeleton.displayName = 'Skeleton'

// Export all types for external use
export type { VariantProps }
export { spinnerVariants, spinnerContainerVariants }