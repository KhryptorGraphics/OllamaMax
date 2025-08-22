import React, { forwardRef, LabelHTMLAttributes } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/utils/cn'

// Label variants using design tokens
const labelVariants = cva(
  [
    'text-sm font-medium leading-none transition-smooth',
    'peer-disabled:cursor-not-allowed peer-disabled:opacity-70'
  ],
  {
    variants: {
      variant: {
        default: 'text-foreground',
        muted: 'text-muted-foreground',
        error: 'text-destructive',
        success: 'text-green-600',
        warning: 'text-yellow-600'
      },
      size: {
        sm: 'text-xs',
        md: 'text-sm',
        lg: 'text-base'
      },
      required: {
        true: "after:content-['*'] after:text-destructive after:ml-1"
      },
      hidden: {
        true: 'sr-only'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'md'
    }
  }
)

export interface LabelProps
  extends LabelHTMLAttributes<HTMLLabelElement>,
    VariantProps<typeof labelVariants> {
  /** Whether the associated field is required */
  required?: boolean
  /** Whether to hide the label visually but keep it accessible */
  hidden?: boolean
}

/**
 * Label component with accessibility features and design token integration
 * 
 * Features:
 * - Multiple variants (default, muted, error, success, warning)
 * - Size variants (sm, md, lg)
 * - Required field indicator
 * - Screen reader support
 * - Proper form association
 * - Design token integration
 */
export const Label = forwardRef<HTMLLabelElement, LabelProps>(
  (
    {
      className,
      variant,
      size,
      required = false,
      hidden = false,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <label
        ref={ref}
        className={cn(
          labelVariants({ variant, size, required, hidden }),
          className
        )}
        {...props}
      >
        {children}
      </label>
    )
  }
)

Label.displayName = 'Label'

// Field set legend component for grouping related form fields
export interface FieldsetLegendProps
  extends React.HTMLAttributes<HTMLLegendElement>,
    Pick<LabelProps, 'variant' | 'size' | 'required'> {}

export const FieldsetLegend = forwardRef<HTMLLegendElement, FieldsetLegendProps>(
  (
    {
      className,
      variant = 'default',
      size = 'md',
      required = false,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <legend
        ref={ref}
        className={cn(
          labelVariants({ variant, size, required }),
          'mb-3',
          className
        )}
        {...props}
      >
        {children}
      </legend>
    )
  }
)

FieldsetLegend.displayName = 'FieldsetLegend'

// Form description component for additional context
export interface FormDescriptionProps
  extends React.HTMLAttributes<HTMLParagraphElement>,
    Pick<LabelProps, 'variant' | 'size'> {}

export const FormDescription = forwardRef<HTMLParagraphElement, FormDescriptionProps>(
  (
    {
      className,
      variant = 'muted',
      size = 'sm',
      children,
      ...props
    },
    ref
  ) => {
    return (
      <p
        ref={ref}
        className={cn(
          labelVariants({ variant, size }),
          'font-normal leading-relaxed',
          className
        )}
        {...props}
      >
        {children}
      </p>
    )
  }
)

FormDescription.displayName = 'FormDescription'

// Error message component
export interface FormErrorProps
  extends React.HTMLAttributes<HTMLParagraphElement> {
  /** Whether to show error icon */
  showIcon?: boolean
  /** Custom error icon */
  icon?: React.ReactNode
}

export const FormError = forwardRef<HTMLParagraphElement, FormErrorProps>(
  (
    {
      className,
      showIcon = true,
      icon,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <p
        ref={ref}
        className={cn(
          labelVariants({ variant: 'error', size: 'sm' }),
          'flex items-center gap-1.5 font-normal leading-relaxed',
          className
        )}
        role="alert"
        aria-live="polite"
        {...props}
      >
        {showIcon && (
          <span className="shrink-0" aria-hidden="true">
            {icon || (
              <svg
                className="w-3 h-3"
                fill="currentColor"
                viewBox="0 0 20 20"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
                  clipRule="evenodd"
                />
              </svg>
            )}
          </span>
        )}
        {children}
      </p>
    )
  }
)

FormError.displayName = 'FormError'

// Success message component
export interface FormSuccessProps
  extends React.HTMLAttributes<HTMLParagraphElement> {
  /** Whether to show success icon */
  showIcon?: boolean
  /** Custom success icon */
  icon?: React.ReactNode
}

export const FormSuccess = forwardRef<HTMLParagraphElement, FormSuccessProps>(
  (
    {
      className,
      showIcon = true,
      icon,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <p
        ref={ref}
        className={cn(
          labelVariants({ variant: 'success', size: 'sm' }),
          'flex items-center gap-1.5 font-normal leading-relaxed',
          className
        )}
        role="status"
        aria-live="polite"
        {...props}
      >
        {showIcon && (
          <span className="shrink-0" aria-hidden="true">
            {icon || (
              <svg
                className="w-3 h-3"
                fill="currentColor"
                viewBox="0 0 20 20"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  fillRule="evenodd"
                  d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                  clipRule="evenodd"
                />
              </svg>
            )}
          </span>
        )}
        {children}
      </p>
    )
  }
)

FormSuccess.displayName = 'FormSuccess'

// Warning message component
export interface FormWarningProps
  extends React.HTMLAttributes<HTMLParagraphElement> {
  /** Whether to show warning icon */
  showIcon?: boolean
  /** Custom warning icon */
  icon?: React.ReactNode
}

export const FormWarning = forwardRef<HTMLParagraphElement, FormWarningProps>(
  (
    {
      className,
      showIcon = true,
      icon,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <p
        ref={ref}
        className={cn(
          labelVariants({ variant: 'warning', size: 'sm' }),
          'flex items-center gap-1.5 font-normal leading-relaxed',
          className
        )}
        role="alert"
        aria-live="polite"
        {...props}
      >
        {showIcon && (
          <span className="shrink-0" aria-hidden="true">
            {icon || (
              <svg
                className="w-3 h-3"
                fill="currentColor"
                viewBox="0 0 20 20"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  fillRule="evenodd"
                  d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                  clipRule="evenodd"
                />
              </svg>
            )}
          </span>
        )}
        {children}
      </p>
    )
  }
)

FormWarning.displayName = 'FormWarning'

export default Label
export { labelVariants }