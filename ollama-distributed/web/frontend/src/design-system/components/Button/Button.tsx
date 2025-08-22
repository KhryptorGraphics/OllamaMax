import React, { forwardRef, ButtonHTMLAttributes, ElementType } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { Loader2 } from 'lucide-react'
import { cn } from '@/utils/cn'

// Button variants using class-variance-authority with design tokens
const buttonVariants = cva(
  // Base styles using design tokens
  [
    'inline-flex items-center justify-center',
    'font-medium transition-smooth',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
    'disabled:pointer-events-none disabled:opacity-50',
    'active:scale-95 select-none',
    'border border-transparent',
    'relative overflow-hidden'
  ],
  {
    variants: {
      variant: {
        primary: [
          'bg-primary-500 text-white shadow-elevation-2',
          'hover:bg-primary-600 hover:shadow-elevation-3',
          'active:bg-primary-700 active:shadow-elevation-1',
          'focus-visible:ring-primary-400',
          'disabled:bg-muted disabled:text-muted-foreground disabled:shadow-none'
        ],
        secondary: [
          'bg-secondary-100 text-secondary-900 shadow-elevation-1',
          'hover:bg-secondary-200 hover:shadow-elevation-2',
          'active:bg-secondary-300 active:shadow-inner',
          'focus-visible:ring-secondary-400',
          'dark:bg-secondary-800 dark:text-secondary-100',
          'dark:hover:bg-secondary-700 dark:active:bg-secondary-600',
          'disabled:bg-muted disabled:text-muted-foreground disabled:shadow-none'
        ],
        outline: [
          'border-border bg-background text-foreground shadow-elevation-1',
          'hover:bg-accent hover:text-accent-foreground hover:shadow-elevation-2',
          'active:bg-accent/80 active:shadow-inner',
          'focus-visible:ring-primary-400',
          'disabled:bg-muted disabled:text-muted-foreground disabled:border-muted'
        ],
        ghost: [
          'text-foreground shadow-none',
          'hover:bg-accent hover:text-accent-foreground',
          'active:bg-accent/80',
          'focus-visible:ring-primary-400',
          'disabled:text-muted-foreground'
        ],
        link: [
          'text-primary-600 underline-offset-4 shadow-none',
          'hover:underline hover:text-primary-700',
          'active:text-primary-800',
          'focus-visible:ring-primary-400',
          'dark:text-primary-400 dark:hover:text-primary-300',
          'disabled:text-muted-foreground disabled:no-underline'
        ],
        destructive: [
          'bg-destructive text-destructive-foreground shadow-elevation-2',
          'hover:bg-destructive/90 hover:shadow-elevation-3',
          'active:bg-destructive/80 active:shadow-elevation-1',
          'focus-visible:ring-destructive',
          'disabled:bg-muted disabled:text-muted-foreground disabled:shadow-none'
        ]
      },
      size: {
        xs: 'h-7 px-2 text-xs rounded-sm gap-1',
        sm: 'h-8 px-3 text-sm rounded-md gap-1.5',
        md: 'h-9 px-4 text-sm rounded-md gap-2',
        lg: 'h-10 px-6 text-base rounded-lg gap-2',
        xl: 'h-12 px-8 text-base rounded-lg gap-2.5',
        icon: 'h-9 w-9 rounded-md'
      },
      fullWidth: {
        true: 'w-full'
      }
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md'
    }
  }
)

// Base button props with polymorphic support
type ButtonBaseProps = VariantProps<typeof buttonVariants> & {
  /** Whether the button is in a loading state */
  loading?: boolean
  /** Loading text to display when loading is true */
  loadingText?: string
  /** Icon to display before the button text */
  leftIcon?: React.ReactNode
  /** Icon to display after the button text */
  rightIcon?: React.ReactNode
  /** Additional CSS classes */
  className?: string
  /** Make button full width */
  fullWidth?: boolean
  /** Children content */
  children?: React.ReactNode
}

// Polymorphic button props
type PolymorphicButtonProps<T extends ElementType = 'button'> = ButtonBaseProps & {
  /** The element type to render as */
  as?: T
} & Omit<React.ComponentPropsWithoutRef<T>, keyof ButtonBaseProps>

export type ButtonProps<T extends ElementType = 'button'> = PolymorphicButtonProps<T>

/**
 * Button component with comprehensive variants, states, and accessibility features
 * 
 * Features:
 * - Multiple variants (primary, secondary, outline, ghost, link, destructive)
 * - Size variants (xs, sm, md, lg, xl, icon)
 * - Loading states with spinner
 * - Icon support (left and right)
 * - Polymorphic support (can render as different elements)
 * - Full accessibility support (WCAG 2.1 AA)
 * - Dark mode support
 * - Enhanced focus management
 * - Design token integration
 */
export const Button = <T extends ElementType = 'button'>(
  {
    as,
    className,
    variant = 'primary',
    size = 'md',
    fullWidth,
    loading = false,
    loadingText,
    leftIcon,
    rightIcon,
    disabled,
    children,
    ...props
  }: ButtonProps<T>
) => {
  const Component = as || 'button'
  const isDisabled = disabled || loading
  
  // Determine if component should have button role
  const shouldHaveButtonRole = Component !== 'button' && !props.role
  
  // Generate unique ID for loading state announcement
  const loadingId = React.useId()

  return (
    <Component
      className={cn(buttonVariants({ variant, size, fullWidth }), className)}
      disabled={Component === 'button' ? isDisabled : undefined}
      aria-disabled={isDisabled}
      aria-busy={loading}
      aria-describedby={loading ? loadingId : undefined}
      role={shouldHaveButtonRole ? 'button' : undefined}
      tabIndex={isDisabled ? -1 : undefined}
      {...props}
    >
      {/* Loading spinner with ARIA live region */}
      {loading && (
        <>
          <Loader2 
            className={cn(
              'animate-spin shrink-0',
              size === 'xs' ? 'w-3 h-3' :
              size === 'sm' ? 'w-3 h-3' :
              size === 'md' ? 'w-4 h-4' :
              size === 'lg' ? 'w-5 h-5' :
              size === 'xl' ? 'w-5 h-5' : 'w-4 h-4'
            )}
            aria-hidden="true"
          />
          {/* Screen reader announcement for loading state */}
          <span id={loadingId} className="sr-only">
            Loading...
          </span>
        </>
      )}

      {/* Left icon */}
      {!loading && leftIcon && (
        <span 
          className={cn(
            'flex items-center justify-center shrink-0',
            size === 'xs' ? 'w-3 h-3' :
            size === 'sm' ? 'w-3 h-3' :
            size === 'md' ? 'w-4 h-4' :
            size === 'lg' ? 'w-5 h-5' :
            size === 'xl' ? 'w-5 h-5' : 'w-4 h-4'
          )}
          aria-hidden="true"
        >
          {leftIcon}
        </span>
      )}

      {/* Button content */}
      {loading && loadingText ? (
        <span className="truncate">{loadingText}</span>
      ) : children ? (
        <span className="truncate">{children}</span>
      ) : null}

      {/* Right icon */}
      {!loading && rightIcon && (
        <span 
          className={cn(
            'flex items-center justify-center shrink-0',
            size === 'xs' ? 'w-3 h-3' :
            size === 'sm' ? 'w-3 h-3' :
            size === 'md' ? 'w-4 h-4' :
            size === 'lg' ? 'w-5 h-5' :
            size === 'xl' ? 'w-5 h-5' : 'w-4 h-4'
          )}
          aria-hidden="true"
        >
          {rightIcon}
        </span>
      )}
    </Component>
  )
}

Button.displayName = 'Button'

// Button group component for related actions
export interface ButtonGroupProps {
  children: React.ReactNode
  orientation?: 'horizontal' | 'vertical'
  className?: string
  spacing?: 'none' | 'sm' | 'md' | 'lg'
}

export const ButtonGroup: React.FC<ButtonGroupProps> = ({
  children,
  orientation = 'horizontal',
  className,
  spacing = 'sm'
}) => {
  const spacingClasses = {
    none: '',
    sm: orientation === 'horizontal' ? 'space-x-2' : 'space-y-2',
    md: orientation === 'horizontal' ? 'space-x-4' : 'space-y-4',
    lg: orientation === 'horizontal' ? 'space-x-6' : 'space-y-6'
  }

  return (
    <div
      className={cn(
        'flex',
        orientation === 'horizontal' ? 'flex-row items-center' : 'flex-col',
        spacingClasses[spacing],
        className
      )}
      role="group"
      aria-label="Button group"
    >
      {children}
    </div>
  )
}

// Icon button variant for actions with only icons
export interface IconButtonProps<T extends ElementType = 'button'> 
  extends Omit<ButtonProps<T>, 'leftIcon' | 'rightIcon' | 'children'> {
  /** Icon to display in the button */
  icon: React.ReactNode
  /** Accessible label for screen readers (required) */
  'aria-label': string
  /** Visual label that appears on hover (optional tooltip) */
  title?: string
}

export const IconButton = <T extends ElementType = 'button'>({
  icon, 
  className, 
  size = 'md', 
  ...props
}: IconButtonProps<T>) => {
  return (
    <Button
      size="icon"
      className={cn(
        // Override icon size based on size prop
        size === 'xs' && 'h-6 w-6',
        size === 'sm' && 'h-7 w-7', 
        size === 'md' && 'h-9 w-9',
        size === 'lg' && 'h-10 w-10',
        size === 'xl' && 'h-12 w-12',
        className
      )}
      {...props}
    >
      <span 
        className={cn(
          'flex items-center justify-center',
          size === 'xs' ? 'w-3 h-3' :
          size === 'sm' ? 'w-3 h-3' :
          size === 'md' ? 'w-4 h-4' :
          size === 'lg' ? 'w-5 h-5' :
          size === 'xl' ? 'w-6 h-6' : 'w-4 h-4'
        )}
        aria-hidden="true"
      >
        {icon}
      </span>
    </Button>
  )
}

IconButton.displayName = 'IconButton'

// Toggle button for binary states
export interface ToggleButtonProps extends Omit<ButtonProps, 'variant'> {
  pressed?: boolean
  onPressedChange?: (pressed: boolean) => void
}

export const ToggleButton = forwardRef<HTMLButtonElement, ToggleButtonProps>(
  ({ pressed = false, onPressedChange, onClick, className, ...props }, ref) => {
    const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
      onPressedChange?.(!pressed)
      onClick?.(event)
    }

    return (
      <Button
        ref={ref}
        variant={pressed ? 'primary' : 'outline'}
        onClick={handleClick}
        aria-pressed={pressed}
        className={cn(
          pressed && 'shadow-inner',
          className
        )}
        {...props}
      />
    )
  }
)

ToggleButton.displayName = 'ToggleButton'

export default Button