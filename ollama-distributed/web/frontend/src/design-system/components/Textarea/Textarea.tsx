import React, { forwardRef, TextareaHTMLAttributes, useId, useState, useRef, useEffect, useCallback } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { AlertCircle, CheckCircle, Info } from 'lucide-react'
import { cn } from '@/utils/cn'

// Textarea variants using design tokens
const textareaVariants = cva(
  [
    'flex w-full min-h-[80px] border bg-background text-foreground transition-smooth',
    'placeholder:text-muted-foreground',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
    'disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-muted',
    'resize-vertical'
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
        sm: 'px-2 py-1.5 text-xs rounded-sm',
        md: 'px-3 py-2 text-sm rounded-md',
        lg: 'px-4 py-3 text-base rounded-lg'
      },
      resize: {
        none: 'resize-none',
        vertical: 'resize-y',
        horizontal: 'resize-x',
        both: 'resize'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'md',
      resize: 'vertical'
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

export interface TextareaProps
  extends TextareaHTMLAttributes<HTMLTextAreaElement>,
    VariantProps<typeof textareaVariants> {
  /** Textarea label */
  label?: string
  /** Help text displayed below the textarea */
  helperText?: string
  /** Error message */
  error?: string
  /** Success message */
  success?: string
  /** Warning message */
  warning?: string
  /** Whether the field is required */
  required?: boolean
  /** Additional CSS classes for the container */
  containerClassName?: string
  /** Additional CSS classes for the label */
  labelClassName?: string
  /** Whether to hide the label visually but keep it for screen readers */
  hideLabel?: boolean
  /** Maximum character length with counter display */
  maxLength?: number
  /** Whether to show character counter */
  showCounter?: boolean
  /** Whether to auto-resize based on content */
  autoResize?: boolean
  /** Minimum number of rows for auto-resize */
  minRows?: number
  /** Maximum number of rows for auto-resize */
  maxRows?: number
  /** Whether to show validation icon */
  showValidationIcon?: boolean
}

/**
 * Textarea component with comprehensive features and accessibility
 * 
 * Features:
 * - Multi-line text input
 * - Auto-resize functionality
 * - Character counting with limits
 * - Validation states (error, success, warning)
 * - Size variants and resize control
 * - Disabled and readonly states
 * - Full accessibility support
 */
export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  (
    {
      className,
      containerClassName,
      labelClassName,
      variant,
      size,
      resize,
      label,
      helperText,
      error,
      success,
      warning,
      required,
      hideLabel = false,
      maxLength,
      showCounter = true,
      autoResize = false,
      minRows = 3,
      maxRows = 10,
      showValidationIcon = true,
      id,
      value,
      defaultValue,
      onChange,
      disabled,
      readOnly,
      rows = 4,
      ...props
    },
    ref
  ) => {
    const [internalValue, setInternalValue] = useState(defaultValue || '')
    const [textareaHeight, setTextareaHeight] = useState<string | undefined>()
    const textareaRef = useRef<HTMLTextAreaElement>(null)
    const hiddenTextareaRef = useRef<HTMLTextAreaElement>(null)
    const textareaId = useId()
    const finalId = id || textareaId
    const helperTextId = `${finalId}-helper`
    const errorId = `${finalId}-error`
    const successId = `${finalId}-success`
    const warningId = `${finalId}-warning`
    const counterId = `${finalId}-counter`

    // Determine variant based on state
    const finalVariant = error ? 'error' : success ? 'success' : warning ? 'warning' : variant

    // Handle controlled/uncontrolled value
    const currentValue = value !== undefined ? value : internalValue
    const characterCount = typeof currentValue === 'string' ? currentValue.length : 0
    const isOverLimit = maxLength ? characterCount > maxLength : false

    // Status icon based on state
    const getStatusIcon = () => {
      if (!showValidationIcon) return null
      if (error) return <AlertCircle className="w-4 h-4 text-destructive" />
      if (success) return <CheckCircle className="w-4 h-4 text-green-600" />
      if (warning) return <Info className="w-4 h-4 text-yellow-600" />
      return null
    }

    const statusIcon = getStatusIcon()

    // Calculate auto-resize height
    const calculateHeight = useCallback(() => {
      if (!autoResize || !hiddenTextareaRef.current || !textareaRef.current) return

      const hiddenTextarea = hiddenTextareaRef.current
      const textarea = textareaRef.current

      // Reset height to get accurate scrollHeight
      hiddenTextarea.value = currentValue as string || ''
      hiddenTextarea.style.height = 'auto'
      
      const scrollHeight = hiddenTextarea.scrollHeight
      const minHeight = parseInt(getComputedStyle(textarea).lineHeight) * minRows
      const maxHeight = parseInt(getComputedStyle(textarea).lineHeight) * maxRows
      
      let newHeight = Math.max(scrollHeight, minHeight)
      if (maxHeight) {
        newHeight = Math.min(newHeight, maxHeight)
      }
      
      setTextareaHeight(`${newHeight}px`)
    }, [autoResize, currentValue, minRows, maxRows])

    // Update height when value changes
    useEffect(() => {
      calculateHeight()
    }, [calculateHeight])

    // Initial height calculation
    useEffect(() => {
      if (autoResize && textareaRef.current) {
        calculateHeight()
        // Add resize observer for font size changes
        const resizeObserver = new ResizeObserver(() => calculateHeight())
        resizeObserver.observe(textareaRef.current)
        return () => resizeObserver.disconnect()
      }
    }, [autoResize, calculateHeight])

    // Handle value changes
    const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const newValue = e.target.value
      
      // Enforce maxLength if specified
      if (maxLength && newValue.length > maxLength && !props.placeholder?.includes('optional')) {
        return
      }
      
      if (value === undefined) {
        setInternalValue(newValue)
      }
      
      onChange?.(e)
    }

    // Build aria-describedby
    const ariaDescribedBy = [
      error && errorId,
      success && successId,
      warning && warningId,
      helperText && helperTextId,
      showCounter && maxLength && counterId
    ].filter(Boolean).join(' ') || undefined

    // Merge refs
    const mergeRefs = (node: HTMLTextAreaElement | null) => {
      textareaRef.current = node
      if (ref) {
        if (typeof ref === 'function') {
          ref(node)
        } else {
          ref.current = node
        }
      }
    }

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

        {/* Textarea container */}
        <div className="relative">
          {/* Hidden textarea for height calculation */}
          {autoResize && (
            <textarea
              ref={hiddenTextareaRef}
              className={cn(
                textareaVariants({ variant: finalVariant, size, resize: 'none' }),
                'absolute -z-10 opacity-0 pointer-events-none overflow-hidden',
                className
              )}
              rows={minRows}
              tabIndex={-1}
              aria-hidden="true"
            />
          )}

          {/* Main textarea */}
          <textarea
            ref={mergeRefs}
            value={currentValue}
            onChange={handleChange}
            className={cn(
              textareaVariants({ variant: finalVariant, size, resize: autoResize ? 'none' : resize }),
              isOverLimit && 'border-destructive focus-visible:ring-destructive',
              statusIcon && 'pr-10',
              className
            )}
            id={finalId}
            rows={autoResize ? minRows : rows}
            style={autoResize && textareaHeight ? { height: textareaHeight } : undefined}
            disabled={disabled}
            readOnly={readOnly}
            aria-invalid={!!error || isOverLimit}
            aria-describedby={ariaDescribedBy}
            {...props}
          />

          {/* Status icon */}
          {statusIcon && (
            <div className="absolute right-3 top-3">
              <span aria-hidden="true">
                {statusIcon}
              </span>
            </div>
          )}
        </div>

        {/* Helper text, error messages, and counter */}
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1 space-y-1">
            {helperText && !error && !success && !warning && (
              <p id={helperTextId} className="text-xs text-muted-foreground">
                {helperText}
              </p>
            )}
            {error && (
              <p id={errorId} className="text-xs text-destructive" role="alert">
                {error}
              </p>
            )}
            {success && (
              <p id={successId} className="text-xs text-green-600">
                {success}
              </p>
            )}
            {warning && (
              <p id={warningId} className="text-xs text-yellow-600">
                {warning}
              </p>
            )}
          </div>

          {/* Character counter */}
          {showCounter && maxLength && (
            <div 
              id={counterId}
              className={cn(
                'text-xs tabular-nums',
                isOverLimit ? 'text-destructive font-medium' : 'text-muted-foreground'
              )}
              aria-live="polite"
              aria-atomic="true"
            >
              {characterCount}/{maxLength}
            </div>
          )}
        </div>
      </div>
    )
  }
)

Textarea.displayName = 'Textarea'

export default Textarea