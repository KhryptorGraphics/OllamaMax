import React, { forwardRef, HTMLAttributes, useEffect, useRef, useState } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/utils/cn'

// Progress bar variants using class-variance-authority with design tokens
const progressVariants = cva(
  'relative overflow-hidden',
  {
    variants: {
      variant: {
        primary: 'bg-primary-100 dark:bg-primary-900/20',
        secondary: 'bg-secondary-100 dark:bg-secondary-900/20', 
        success: 'bg-success-100 dark:bg-success-900/20',
        warning: 'bg-warning-100 dark:bg-warning-900/20',
        error: 'bg-error-100 dark:bg-error-900/20',
        info: 'bg-info-100 dark:bg-info-900/20',
        neutral: 'bg-neutral-200 dark:bg-neutral-700'
      },
      size: {
        xs: 'h-1',
        sm: 'h-2',
        md: 'h-3',
        lg: 'h-4',
        xl: 'h-6'
      }
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md'
    }
  }
)

// Progress fill variants
const progressFillVariants = cva(
  'h-full transition-all duration-300 ease-out',
  {
    variants: {
      variant: {
        primary: 'bg-primary-500 dark:bg-primary-400',
        secondary: 'bg-secondary-500 dark:bg-secondary-400',
        success: 'bg-success-500 dark:bg-success-400',
        warning: 'bg-warning-500 dark:bg-warning-400',
        error: 'bg-error-500 dark:bg-error-400',
        info: 'bg-info-500 dark:bg-info-400',
        neutral: 'bg-neutral-500 dark:bg-neutral-400'
      },
      animated: {
        true: 'relative overflow-hidden',
        false: ''
      },
      striped: {
        true: 'bg-gradient-to-r from-transparent via-white/20 to-transparent bg-[length:20px_100%]',
        false: ''
      }
    },
    compoundVariants: [
      {
        animated: true,
        striped: true,
        className: 'animate-[progress-stripes_1s_linear_infinite]'
      }
    ],
    defaultVariants: {
      variant: 'primary',
      animated: false,
      striped: false
    }
  }
)

// Circular progress sizes
const circularSizes = {
  xs: { size: 24, strokeWidth: 3 },
  sm: { size: 32, strokeWidth: 3 },
  md: { size: 48, strokeWidth: 4 },
  lg: { size: 64, strokeWidth: 5 },
  xl: { size: 96, strokeWidth: 6 }
}

// Progress label variants
const progressLabelVariants = cva(
  'font-medium tabular-nums',
  {
    variants: {
      size: {
        xs: 'text-xs',
        sm: 'text-xs',
        md: 'text-sm',
        lg: 'text-base',
        xl: 'text-lg'
      },
      position: {
        inside: 'absolute inset-0 flex items-center justify-center text-white mix-blend-difference',
        outside: 'mt-1',
        'outside-start': 'mb-1',
        inline: 'ml-2'
      }
    },
    defaultVariants: {
      size: 'md',
      position: 'outside'
    }
  }
)

// Base progress props
export interface ProgressBaseProps extends VariantProps<typeof progressVariants> {
  /** Current progress value (0-100 for percentage, or 0-max for custom) */
  value?: number
  /** Maximum value for progress (default: 100) */
  max?: number
  /** Minimum value for progress (default: 0) */
  min?: number
  /** Whether the progress bar is animated */
  animated?: boolean
  /** Whether to show striped pattern */
  striped?: boolean
  /** Indeterminate loading state (ignores value) */
  indeterminate?: boolean
  /** Show value display */
  showValue?: boolean
  /** Value display format */
  valueFormat?: 'percentage' | 'fraction' | 'custom'
  /** Custom value formatter function */
  formatValue?: (value: number, max: number) => string
  /** Position of the value label */
  labelPosition?: 'inside' | 'outside' | 'outside-start' | 'inline' | 'none'
  /** Custom label text (overrides value display) */
  label?: string
  /** Additional CSS classes */
  className?: string
  /** Accessible label for screen readers */
  'aria-label'?: string
  /** Accessible description */
  'aria-describedby'?: string
}

// Linear progress props
export interface LinearProgressProps extends ProgressBaseProps, 
  Omit<HTMLAttributes<HTMLDivElement>, keyof ProgressBaseProps> {
  /** Whether to show the progress as a thin line */
  thin?: boolean
  /** Border radius style */
  rounded?: 'none' | 'sm' | 'md' | 'lg' | 'full'
}

/**
 * Linear Progress component with comprehensive features
 * 
 * Features:
 * - Multiple color variants with design token integration
 * - Size variants (xs, sm, md, lg, xl)
 * - Animated and static modes
 * - Striped pattern support
 * - Indeterminate loading state
 * - Value display with multiple formats
 * - Full accessibility support (WCAG 2.1 AA)
 * - Dark mode support
 * - Smooth animations
 */
export const LinearProgress = forwardRef<HTMLDivElement, LinearProgressProps>(
  ({
    value = 0,
    max = 100,
    min = 0,
    variant = 'primary',
    size = 'md',
    animated = false,
    striped = false,
    indeterminate = false,
    showValue = false,
    valueFormat = 'percentage',
    formatValue,
    labelPosition = 'outside',
    label,
    thin = false,
    rounded = 'md',
    className,
    'aria-label': ariaLabel,
    'aria-describedby': ariaDescribedBy,
    ...props
  }, ref) => {
    // Calculate normalized progress (0-100)
    const normalizedValue = Math.min(Math.max(((value - min) / (max - min)) * 100, 0), 100)
    
    // Format value display
    const getFormattedValue = () => {
      if (label) return label
      if (!showValue || labelPosition === 'none') return null
      
      if (formatValue) {
        return formatValue(value, max)
      }
      
      switch (valueFormat) {
        case 'percentage':
          return `${Math.round(normalizedValue)}%`
        case 'fraction':
          return `${value}/${max}`
        case 'custom':
        default:
          return `${value}`
      }
    }
    
    const formattedValue = getFormattedValue()
    
    // Rounded classes
    const roundedClasses = {
      none: '',
      sm: 'rounded-sm',
      md: 'rounded-md',
      lg: 'rounded-lg',
      full: 'rounded-full'
    }
    
    // Render label
    const renderLabel = () => {
      if (!formattedValue || labelPosition === 'none') return null
      
      return (
        <span className={cn(progressLabelVariants({ size, position: labelPosition }))}>
          {formattedValue}
        </span>
      )
    }
    
    return (
      <div className={cn('w-full', className)} {...props}>
        {labelPosition === 'outside-start' && renderLabel()}
        
        <div className="flex items-center">
          <div
            ref={ref}
            role="progressbar"
            aria-valuenow={indeterminate ? undefined : value}
            aria-valuemin={min}
            aria-valuemax={max}
            aria-label={ariaLabel || (indeterminate ? 'Loading...' : `Progress: ${formattedValue}`)}
            aria-describedby={ariaDescribedBy}
            className={cn(
              progressVariants({ variant, size: thin ? 'xs' : size }),
              roundedClasses[rounded],
              'w-full'
            )}
          >
            <div
              className={cn(
                progressFillVariants({ 
                  variant, 
                  animated: animated || indeterminate,
                  striped 
                }),
                roundedClasses[rounded],
                indeterminate && 'animate-[progress-indeterminate_1.5s_ease-in-out_infinite]'
              )}
              style={!indeterminate ? { width: `${normalizedValue}%` } : { width: '30%' }}
            >
              {indeterminate && (
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent animate-shimmer" />
              )}
            </div>
            
            {labelPosition === 'inside' && renderLabel()}
          </div>
          
          {labelPosition === 'inline' && renderLabel()}
        </div>
        
        {labelPosition === 'outside' && renderLabel()}
      </div>
    )
  }
)

LinearProgress.displayName = 'LinearProgress'

// Circular progress props
export interface CircularProgressProps extends ProgressBaseProps,
  Omit<HTMLAttributes<HTMLDivElement>, keyof ProgressBaseProps> {
  /** Thickness of the progress stroke */
  thickness?: number
  /** Whether to show track behind progress */
  showTrack?: boolean
  /** Start angle in degrees (0 = top, 90 = right) */
  startAngle?: number
}

/**
 * Circular Progress component with SVG-based implementation
 * 
 * Features:
 * - Circular/radial progress indicator
 * - Multiple color variants
 * - Size variants (xs, sm, md, lg, xl)
 * - Animated rotation for indeterminate state
 * - Value display in center
 * - Customizable stroke thickness
 * - Full accessibility support
 */
export const CircularProgress = forwardRef<HTMLDivElement, CircularProgressProps>(
  ({
    value = 0,
    max = 100,
    min = 0,
    variant = 'primary',
    size = 'md',
    animated = false,
    indeterminate = false,
    showValue = false,
    valueFormat = 'percentage',
    formatValue,
    label,
    thickness,
    showTrack = true,
    startAngle = -90,
    className,
    'aria-label': ariaLabel,
    'aria-describedby': ariaDescribedBy,
    ...props
  }, ref) => {
    const { size: svgSize, strokeWidth: defaultStrokeWidth } = circularSizes[size || 'md']
    const strokeWidth = thickness || defaultStrokeWidth
    const radius = (svgSize - strokeWidth) / 2
    const circumference = 2 * Math.PI * radius
    
    // Calculate normalized progress (0-100)
    const normalizedValue = Math.min(Math.max(((value - min) / (max - min)) * 100, 0), 100)
    const strokeDashoffset = circumference - (normalizedValue / 100) * circumference
    
    // Format value display
    const getFormattedValue = () => {
      if (label) return label
      if (!showValue) return null
      
      if (formatValue) {
        return formatValue(value, max)
      }
      
      switch (valueFormat) {
        case 'percentage':
          return `${Math.round(normalizedValue)}%`
        case 'fraction':
          return `${value}/${max}`
        case 'custom':
        default:
          return `${value}`
      }
    }
    
    const formattedValue = getFormattedValue()
    
    // Variant colors
    const variantColors = {
      primary: 'text-primary-500 dark:text-primary-400',
      secondary: 'text-secondary-500 dark:text-secondary-400',
      success: 'text-success-500 dark:text-success-400',
      warning: 'text-warning-500 dark:text-warning-400',
      error: 'text-error-500 dark:text-error-400',
      info: 'text-info-500 dark:text-info-400',
      neutral: 'text-neutral-500 dark:text-neutral-400'
    }
    
    const trackColors = {
      primary: 'text-primary-100 dark:text-primary-900/20',
      secondary: 'text-secondary-100 dark:text-secondary-900/20',
      success: 'text-success-100 dark:text-success-900/20',
      warning: 'text-warning-100 dark:text-warning-900/20',
      error: 'text-error-100 dark:text-error-900/20',
      info: 'text-info-100 dark:text-info-900/20',
      neutral: 'text-neutral-200 dark:text-neutral-700'
    }
    
    return (
      <div
        ref={ref}
        role="progressbar"
        aria-valuenow={indeterminate ? undefined : value}
        aria-valuemin={min}
        aria-valuemax={max}
        aria-label={ariaLabel || (indeterminate ? 'Loading...' : `Progress: ${formattedValue}`)}
        aria-describedby={ariaDescribedBy}
        className={cn('relative inline-flex', className)}
        style={{ width: svgSize, height: svgSize }}
        {...props}
      >
        <svg
          width={svgSize}
          height={svgSize}
          viewBox={`0 0 ${svgSize} ${svgSize}`}
          className={cn(
            indeterminate && 'animate-spin',
            variantColors[variant || 'primary']
          )}
        >
          {/* Track circle */}
          {showTrack && (
            <circle
              cx={svgSize / 2}
              cy={svgSize / 2}
              r={radius}
              fill="none"
              stroke="currentColor"
              strokeWidth={strokeWidth}
              className={trackColors[variant || 'primary']}
            />
          )}
          
          {/* Progress circle */}
          <circle
            cx={svgSize / 2}
            cy={svgSize / 2}
            r={radius}
            fill="none"
            stroke="currentColor"
            strokeWidth={strokeWidth}
            strokeDasharray={circumference}
            strokeDashoffset={indeterminate ? circumference * 0.75 : strokeDashoffset}
            strokeLinecap="round"
            transform={`rotate(${startAngle} ${svgSize / 2} ${svgSize / 2})`}
            className={cn(
              'transition-all duration-300 ease-out',
              animated && !indeterminate && 'animate-pulse'
            )}
          />
        </svg>
        
        {/* Center value display */}
        {showValue && formattedValue && (
          <div className="absolute inset-0 flex items-center justify-center">
            <span className={cn(progressLabelVariants({ size }))}>
              {formattedValue}
            </span>
          </div>
        )}
      </div>
    )
  }
)

CircularProgress.displayName = 'CircularProgress'

// Steps progress props
export interface StepsProgressProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  /** Total number of steps */
  steps: number
  /** Current step (0-indexed) */
  currentStep: number
  /** Variant for styling */
  variant?: VariantProps<typeof progressVariants>['variant']
  /** Size of the steps indicator */
  size?: 'sm' | 'md' | 'lg'
  /** Labels for each step */
  stepLabels?: string[]
  /** Whether to show step numbers */
  showStepNumbers?: boolean
  /** Orientation of steps */
  orientation?: 'horizontal' | 'vertical'
  /** Click handler for steps */
  onStepClick?: (step: number) => void
}

/**
 * Steps Progress component for multi-step processes
 * 
 * Features:
 * - Visual step indicators
 * - Clickable steps for navigation
 * - Step labels and numbers
 * - Horizontal and vertical layouts
 * - Accessibility support
 */
export const StepsProgress: React.FC<StepsProgressProps> = ({
  steps,
  currentStep,
  variant = 'primary',
  size = 'md',
  stepLabels,
  showStepNumbers = true,
  orientation = 'horizontal',
  onStepClick,
  className,
  ...props
}) => {
  const sizeClasses = {
    sm: 'w-6 h-6 text-xs',
    md: 'w-8 h-8 text-sm',
    lg: 'w-10 h-10 text-base'
  }
  
  const variantColors = {
    primary: 'bg-primary-500 text-white border-primary-500',
    secondary: 'bg-secondary-500 text-white border-secondary-500',
    success: 'bg-success-500 text-white border-success-500',
    warning: 'bg-warning-500 text-white border-warning-500',
    error: 'bg-error-500 text-white border-error-500',
    info: 'bg-info-500 text-white border-info-500',
    neutral: 'bg-neutral-500 text-white border-neutral-500'
  }
  
  const inactiveColors = 'bg-neutral-200 dark:bg-neutral-700 text-neutral-500 dark:text-neutral-400 border-neutral-300 dark:border-neutral-600'
  
  return (
    <div
      className={cn(
        'flex',
        orientation === 'horizontal' ? 'flex-row items-center' : 'flex-col',
        className
      )}
      role="group"
      aria-label="Progress steps"
      {...props}
    >
      {Array.from({ length: steps }, (_, index) => {
        const isActive = index <= currentStep
        const isCurrent = index === currentStep
        const isClickable = onStepClick !== undefined
        
        return (
          <React.Fragment key={index}>
            <button
              type="button"
              onClick={() => onStepClick?.(index)}
              disabled={!isClickable}
              className={cn(
                'relative flex items-center justify-center rounded-full border-2 transition-all',
                sizeClasses[size],
                isActive ? variantColors[variant || 'primary'] : inactiveColors,
                isCurrent && 'ring-2 ring-offset-2 ring-offset-background',
                isClickable && 'cursor-pointer hover:scale-110',
                !isClickable && 'cursor-default'
              )}
              aria-label={stepLabels?.[index] || `Step ${index + 1}`}
              aria-current={isCurrent ? 'step' : undefined}
            >
              {showStepNumbers && (
                <span className="font-semibold">{index + 1}</span>
              )}
            </button>
            
            {index < steps - 1 && (
              <div
                className={cn(
                  'transition-all',
                  orientation === 'horizontal' ? 'flex-1 h-0.5 mx-2' : 'w-0.5 flex-1 my-2',
                  index < currentStep ? variantColors[variant || 'primary'] : inactiveColors
                )}
                aria-hidden="true"
              />
            )}
          </React.Fragment>
        )
      })}
    </div>
  )
}

StepsProgress.displayName = 'StepsProgress'

// Export compound component
export const Progress = Object.assign(LinearProgress, {
  Linear: LinearProgress,
  Circular: CircularProgress,
  Steps: StepsProgress
})

export default Progress