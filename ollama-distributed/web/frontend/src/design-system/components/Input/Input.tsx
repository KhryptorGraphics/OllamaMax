import React, { forwardRef, InputHTMLAttributes, useState, useId } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { Eye, EyeOff, AlertCircle, CheckCircle, Info, X, Search } from 'lucide-react'
import { cn } from '@/utils/cn'

// Input variants using design tokens
const inputVariants = cva(
  [
    'flex w-full border bg-background text-foreground transition-smooth',
    'placeholder:text-muted-foreground',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
    'disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-muted',
    'file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground',
    '[&::-webkit-search-cancel-button]:appearance-none'
  ],
  {
    variants: {
      variant: {
        default: [
          'border-border shadow-elevation-1',
          'hover:border-border/80 hover:shadow-elevation-2',
          'focus-visible:border-ring focus-visible:shadow-elevation-2'
        ],
        error: [
          'border-destructive text-foreground shadow-elevation-1',
          'focus-visible:ring-destructive focus-visible:border-destructive',
          'hover:border-destructive/80'
        ],
        success: [
          'border-green-500 text-foreground shadow-elevation-1',
          'focus-visible:ring-green-500 focus-visible:border-green-500',
          'hover:border-green-400'
        ],
        warning: [
          'border-yellow-500 text-foreground shadow-elevation-1',
          'focus-visible:ring-yellow-500 focus-visible:border-yellow-500',
          'hover:border-yellow-400'
        ]
      },
      size: {
        sm: 'h-8 px-2 text-xs rounded-sm',
        md: 'h-9 px-3 text-sm rounded-md',
        lg: 'h-10 px-4 text-base rounded-lg'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'md'
    }
  }
)

// Label styles
const labelVariants = cva(
  'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
  {
    variants: {
      variant: {
        default: 'text-foreground',
        error: 'text-destructive',
        success: 'text-success',
        warning: 'text-warning'
      },
      required: {
        true: "after:content-['*'] after:ml-1 after:text-destructive"
      }
    },
    defaultVariants: {
      variant: 'default'
    }
  }
)

export interface InputProps
  extends InputHTMLAttributes<HTMLInputElement>,
    VariantProps<typeof inputVariants> {
  /** Input label */
  label?: string
  /** Help text displayed below the input */
  helperText?: string
  /** Error message */
  error?: string
  /** Success message */
  success?: string
  /** Warning message */
  warning?: string
  /** Icon to display on the left side */
  leftIcon?: React.ReactNode
  /** Icon to display on the right side */
  rightIcon?: React.ReactNode
  /** Whether the field is required */
  required?: boolean
  /** Additional CSS classes for the container */
  containerClassName?: string
  /** Additional CSS classes for the label */
  labelClassName?: string
  /** Whether to show a clear button when input has value */
  clearable?: boolean
  /** Callback when clear button is clicked */
  onClear?: () => void
  /** Whether to hide the label visually but keep it for screen readers */
  hideLabel?: boolean
}

/**
 * Input component with comprehensive features and accessibility support
 * 
 * Features:
 * - Multiple variants (default, error, success, warning)
 * - Size variants (sm, md, lg)
 * - Icon support (left and right)
 * - Label and helper text
 * - Error, success, and warning states
 * - Password visibility toggle
 * - Full accessibility support
 * - Form validation integration
 */
export const Input = forwardRef<HTMLInputElement, InputProps>(
  (
    {
      className,
      containerClassName,
      labelClassName,
      variant,
      size,
      type,
      label,
      helperText,
      error,
      success,
      warning,
      leftIcon,
      rightIcon,
      required,
      clearable = false,
      onClear,
      hideLabel = false,
      id,
      value,
      onChange,
      ...props
    },
    ref
  ) => {
    const [showPassword, setShowPassword] = useState(false)
    const [internalValue, setInternalValue] = useState(value || '')
    
    const inputId = useId()
    const finalId = id || inputId
    const helperTextId = `${finalId}-helper`
    const errorId = `${finalId}-error`
    const successId = `${finalId}-success`
    const warningId = `${finalId}-warning`

    // Determine variant based on state
    const finalVariant = error ? 'error' : success ? 'success' : warning ? 'warning' : variant

    // Handle controlled/uncontrolled value
    const currentValue = value !== undefined ? value : internalValue
    const hasValue = Boolean(currentValue)

    // Password toggle for password inputs
    const isPassword = type === 'password'
    const isSearch = type === 'search'
    const inputType = isPassword && showPassword ? 'text' : type

    // Status icon based on state
    const getStatusIcon = () => {
      if (error) return <AlertCircle className="w-4 h-4 text-destructive" />
      if (success) return <CheckCircle className="w-4 h-4 text-green-600" />
      if (warning) return <Info className="w-4 h-4 text-yellow-600" />
      return null
    }

    const statusIcon = getStatusIcon()
    
    // Determine left icon
    const displayLeftIcon = isSearch ? <Search className="w-4 h-4" /> : leftIcon
    
    // Determine if we need right side spacing
    const hasRightContent = statusIcon || rightIcon || isPassword || (clearable && hasValue)

    // Handle value changes
    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = e.target.value
      if (value === undefined) {
        setInternalValue(newValue)
      }
      onChange?.(e)
    }

    // Handle clear action
    const handleClear = () => {
      const syntheticEvent = {
        target: { value: '' },
        currentTarget: { value: '' }
      } as React.ChangeEvent<HTMLInputElement>
      
      if (value === undefined) {
        setInternalValue('')
      }
      
      onClear?.()
      onChange?.(syntheticEvent)
    }

    // Build aria-describedby
    const ariaDescribedBy = [
      error && errorId,
      success && successId,
      warning && warningId,
      helperText && helperTextId
    ].filter(Boolean).join(' ') || undefined

    return (
      <div className={cn('space-y-2', containerClassName)}>
        {/* Label */}
        {label && (
          <label
            htmlFor={finalId}
            className={cn(
              labelVariants({ variant: finalVariant, required }),
              hideLabel && 'sr-only',
              labelClassName
            )}
          >
            {label}
          </label>
        )}

        {/* Input container */}
        <div className="relative">
          {/* Left icon */}
          {displayLeftIcon && (
            <div className="absolute left-3 top-1/2 -translate-y-1/2 flex items-center pointer-events-none">
              <span className="w-4 h-4 text-muted-foreground" aria-hidden="true">
                {displayLeftIcon}
              </span>
            </div>
          )}

          {/* Input */}
          <input
            type={inputType}
            value={currentValue}
            onChange={handleChange}
            className={cn(
              inputVariants({ variant: finalVariant, size }),
              displayLeftIcon && 'pl-10',
              hasRightContent && 'pr-10',
              className
            )}
            ref={ref}
            id={finalId}
            aria-invalid={!!error}
            aria-describedby={ariaDescribedBy}
            {...props}
          />

          {/* Right side icons */}
          {hasRightContent && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center space-x-1">
              {/* Clear button */}
              {clearable && hasValue && !props.disabled && (
                <button
                  type="button"
                  onClick={handleClear}
                  className="text-muted-foreground hover:text-foreground transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 rounded"
                  aria-label="Clear input"
                >
                  <X className="w-4 h-4" />
                </button>
              )}

              {/* Status icon */}
              {statusIcon && (
                <span aria-hidden="true">
                  {statusIcon}
                </span>
              )}

              {/* Custom right icon */}
              {rightIcon && !statusIcon && (
                <span className="w-4 h-4 text-muted-foreground" aria-hidden="true">
                  {rightIcon}
                </span>
              )}

              {/* Password toggle */}
              {isPassword && (
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="text-muted-foreground hover:text-foreground transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 rounded"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                >
                  {showPassword ? (
                    <EyeOff className="w-4 h-4" aria-hidden="true" />
                  ) : (
                    <Eye className="w-4 h-4" aria-hidden="true" />
                  )}
                </button>
              )}
            </div>
          )}
        </div>

        {/* Helper text, error, success, or warning message */}
        {(helperText || error || success || warning) && (
          <div className="space-y-1">
            {error && (
              <p
                id={errorId}
                className="text-sm text-destructive flex items-center"
                role="alert"
              >
                <AlertCircle className="w-3 h-3 mr-1 flex-shrink-0" />
                {error}
              </p>
            )}
            
            {success && !error && (
              <p className="text-sm text-success flex items-center">
                <CheckCircle className="w-3 h-3 mr-1 flex-shrink-0" />
                {success}
              </p>
            )}
            
            {warning && !error && !success && (
              <p className="text-sm text-warning flex items-center">
                <Info className="w-3 h-3 mr-1 flex-shrink-0" />
                {warning}
              </p>
            )}
            
            {helperText && !error && !success && !warning && (
              <p
                id={helperTextId}
                className="text-sm text-muted-foreground"
              >
                {helperText}
              </p>
            )}
          </div>
        )}
      </div>
    )
  }
)

Input.displayName = 'Input'

// Textarea component with similar styling
export interface TextareaProps
  extends Omit<React.TextareaHTMLAttributes<HTMLTextAreaElement>, 'size'>,
    Pick<InputProps, 'label' | 'helperText' | 'error' | 'success' | 'warning' | 'required' | 'containerClassName' | 'labelClassName'> {
  /** Textarea size */
  size?: 'sm' | 'md' | 'lg'
  /** Auto-resize the textarea */
  autoResize?: boolean
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  (
    {
      className,
      containerClassName,
      labelClassName,
      label,
      helperText,
      error,
      success,
      warning,
      required,
      id,
      size = 'md',
      autoResize = false,
      ...props
    },
    ref
  ) => {
    const inputId = useId()
    const finalId = id || inputId
    const helperTextId = `${finalId}-helper`
    const errorId = `${finalId}-error`

    const variant = error ? 'error' : success ? 'success' : warning ? 'warning' : 'default'

    const sizeClasses = {
      sm: 'min-h-[80px] px-2 py-1.5 text-xs',
      md: 'min-h-[100px] px-3 py-2 text-sm',
      lg: 'min-h-[120px] px-4 py-3 text-base'
    }

    return (
      <div className={cn('space-y-2', containerClassName)}>
        {/* Label */}
        {label && (
          <label
            htmlFor={finalId}
            className={cn(
              labelVariants({ variant, required }),
              labelClassName
            )}
          >
            {label}
          </label>
        )}

        {/* Textarea */}
        <textarea
          className={cn(
            'flex w-full rounded-md border border-input bg-background',
            'placeholder:text-muted-foreground',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2',
            'disabled:cursor-not-allowed disabled:opacity-50',
            'transition-colors duration-200',
            variant === 'default' && 'focus-visible:ring-primary-500 hover:border-primary-300',
            variant === 'error' && 'border-destructive focus-visible:ring-destructive',
            variant === 'success' && 'border-success focus-visible:ring-success',
            variant === 'warning' && 'border-warning focus-visible:ring-warning',
            sizeClasses[size],
            autoResize && 'resize-none',
            className
          )}
          ref={ref}
          id={finalId}
          aria-invalid={!!error}
          aria-describedby={cn(
            helperText && helperTextId,
            error && errorId
          )}
          {...props}
        />

        {/* Helper text, error, success, or warning message */}
        {(helperText || error || success || warning) && (
          <div className="space-y-1">
            {error && (
              <p
                id={errorId}
                className="text-sm text-destructive flex items-center"
                role="alert"
              >
                <AlertCircle className="w-3 h-3 mr-1 flex-shrink-0" />
                {error}
              </p>
            )}
            
            {success && !error && (
              <p className="text-sm text-success flex items-center">
                <CheckCircle className="w-3 h-3 mr-1 flex-shrink-0" />
                {success}
              </p>
            )}
            
            {warning && !error && !success && (
              <p className="text-sm text-warning flex items-center">
                <Info className="w-3 h-3 mr-1 flex-shrink-0" />
                {warning}
              </p>
            )}
            
            {helperText && !error && !success && !warning && (
              <p
                id={helperTextId}
                className="text-sm text-muted-foreground"
              >
                {helperText}
              </p>
            )}
          </div>
        )}
      </div>
    )
  }
)

Textarea.displayName = 'Textarea'

export default Input