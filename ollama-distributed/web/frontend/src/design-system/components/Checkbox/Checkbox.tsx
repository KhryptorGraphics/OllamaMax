import React, { forwardRef, InputHTMLAttributes, useId, useState, useRef, useEffect } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { Check, Minus } from 'lucide-react'
import { cn } from '@/utils/cn'

// Checkbox variants using design tokens
const checkboxVariants = cva(
  [
    'shrink-0 border rounded transition-smooth',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
    'disabled:cursor-not-allowed disabled:opacity-50',
    'data-[state=checked]:bg-primary data-[state=checked]:border-primary',
    'data-[state=indeterminate]:bg-primary data-[state=indeterminate]:border-primary'
  ],
  {
    variants: {
      variant: {
        default: [
          'border-border bg-background',
          'hover:border-primary/50'
        ],
        error: [
          'border-destructive',
          'hover:border-destructive/80',
          'data-[state=checked]:bg-destructive data-[state=checked]:border-destructive'
        ],
        success: [
          'border-green-500',
          'hover:border-green-400',
          'data-[state=checked]:bg-green-500 data-[state=checked]:border-green-500'
        ],
        warning: [
          'border-yellow-500',
          'hover:border-yellow-400',
          'data-[state=checked]:bg-yellow-500 data-[state=checked]:border-yellow-500'
        ]
      },
      size: {
        sm: 'h-4 w-4',
        md: 'h-5 w-5',
        lg: 'h-6 w-6'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'md'
    }
  }
)

// Label variants
const labelVariants = cva(
  'font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
  {
    variants: {
      size: {
        sm: 'text-xs',
        md: 'text-sm',
        lg: 'text-base'
      }
    },
    defaultVariants: {
      size: 'md'
    }
  }
)

// Checkbox group layout variants
const groupVariants = cva(
  'flex gap-4',
  {
    variants: {
      layout: {
        horizontal: 'flex-row flex-wrap',
        vertical: 'flex-col',
        grid: 'grid grid-cols-2 sm:grid-cols-3'
      }
    },
    defaultVariants: {
      layout: 'vertical'
    }
  }
)

export interface CheckboxProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type' | 'size'>,
    VariantProps<typeof checkboxVariants> {
  /** Label for the checkbox */
  label?: string
  /** Whether the checkbox is in an indeterminate state */
  indeterminate?: boolean
  /** Custom icon for checked state */
  checkedIcon?: React.ReactNode
  /** Custom icon for indeterminate state */
  indeterminateIcon?: React.ReactNode
  /** Custom color for checked state */
  checkedColor?: string
  /** Additional CSS classes for the label */
  labelClassName?: string
  /** Description text below the label */
  description?: string
  /** Error message */
  error?: string
  /** Whether to hide the label visually but keep it for screen readers */
  hideLabel?: boolean
  /** Callback when indeterminate state changes */
  onIndeterminateChange?: (indeterminate: boolean) => void
}

/**
 * Checkbox component with comprehensive features and accessibility
 * 
 * Features:
 * - Basic checked/unchecked states
 * - Indeterminate state for partial selections
 * - Multiple variants and sizes
 * - Custom icons and colors
 * - Full keyboard navigation
 * - ARIA attributes for accessibility
 * - Description text support
 */
export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  (
    {
      className,
      labelClassName,
      variant,
      size,
      label,
      indeterminate = false,
      checkedIcon,
      indeterminateIcon,
      checkedColor,
      description,
      error,
      hideLabel = false,
      checked,
      defaultChecked,
      onChange,
      onIndeterminateChange,
      disabled,
      id,
      ...props
    },
    ref
  ) => {
    const [isChecked, setIsChecked] = useState(defaultChecked || false)
    const [isIndeterminate, setIsIndeterminate] = useState(indeterminate)
    const inputRef = useRef<HTMLInputElement>(null)
    const checkboxId = useId()
    const finalId = id || checkboxId
    const descriptionId = description ? `${finalId}-description` : undefined
    const errorId = error ? `${finalId}-error` : undefined

    // Merge refs
    useEffect(() => {
      const input = inputRef.current
      if (input) {
        input.indeterminate = isIndeterminate
      }
    }, [isIndeterminate])

    // Handle controlled/uncontrolled checked state
    const currentChecked = checked !== undefined ? checked : isChecked

    // Update indeterminate state when prop changes
    useEffect(() => {
      setIsIndeterminate(indeterminate)
    }, [indeterminate])

    // Handle change event
    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const newChecked = e.target.checked
      
      if (checked === undefined) {
        setIsChecked(newChecked)
      }
      
      // Clear indeterminate state when user interacts
      if (isIndeterminate) {
        setIsIndeterminate(false)
        onIndeterminateChange?.(false)
      }
      
      onChange?.(e)
    }

    // Determine state for styling
    const state = isIndeterminate ? 'indeterminate' : currentChecked ? 'checked' : 'unchecked'
    
    // Determine final variant
    const finalVariant = error ? 'error' : variant

    // Get icon based on state
    const getIcon = () => {
      if (isIndeterminate) {
        return indeterminateIcon || <Minus className="h-3 w-3" />
      }
      if (currentChecked) {
        return checkedIcon || <Check className="h-3 w-3" />
      }
      return null
    }

    // Build aria-describedby
    const ariaDescribedBy = [
      descriptionId,
      errorId
    ].filter(Boolean).join(' ') || undefined

    return (
      <div className="flex items-start space-x-2">
        <div className="relative flex items-center">
          <input
            type="checkbox"
            className="sr-only peer"
            ref={(node) => {
              // Handle both refs
              inputRef.current = node
              if (ref) {
                if (typeof ref === 'function') {
                  ref(node)
                } else {
                  ref.current = node
                }
              }
            }}
            id={finalId}
            checked={currentChecked}
            onChange={handleChange}
            disabled={disabled}
            aria-invalid={!!error}
            aria-describedby={ariaDescribedBy}
            {...props}
          />
          <div
            className={cn(
              checkboxVariants({ variant: finalVariant, size }),
              'flex items-center justify-center',
              className
            )}
            data-state={state}
            style={currentChecked && checkedColor ? { backgroundColor: checkedColor, borderColor: checkedColor } : undefined}
          >
            <span className="text-white pointer-events-none">
              {getIcon()}
            </span>
          </div>
        </div>
        
        {(label || description || error) && (
          <div className="flex flex-col space-y-1">
            {label && (
              <label
                htmlFor={finalId}
                className={cn(
                  labelVariants({ size }),
                  'cursor-pointer select-none',
                  disabled && 'cursor-not-allowed opacity-70',
                  hideLabel && 'sr-only',
                  labelClassName
                )}
              >
                {label}
              </label>
            )}
            {description && (
              <p 
                id={descriptionId}
                className="text-xs text-muted-foreground"
              >
                {description}
              </p>
            )}
            {error && (
              <p 
                id={errorId}
                className="text-xs text-destructive"
                role="alert"
              >
                {error}
              </p>
            )}
          </div>
        )}
      </div>
    )
  }
)

Checkbox.displayName = 'Checkbox'

export interface CheckboxGroupProps extends VariantProps<typeof groupVariants> {
  /** Group label */
  label?: string
  /** Array of checkbox options */
  options: Array<{
    value: string
    label: string
    description?: string
    disabled?: boolean
  }>
  /** Selected values */
  value?: string[]
  /** Default selected values */
  defaultValue?: string[]
  /** Callback when selection changes */
  onChange?: (values: string[]) => void
  /** Whether the group is disabled */
  disabled?: boolean
  /** Error message */
  error?: string
  /** Required field */
  required?: boolean
  /** Additional CSS classes */
  className?: string
}

/**
 * CheckboxGroup component for managing multiple checkboxes
 * 
 * Features:
 * - Group layouts (horizontal, vertical, grid)
 * - Controlled/uncontrolled modes
 * - Select all functionality
 * - Accessibility support with fieldset and legend
 */
export const CheckboxGroup = forwardRef<HTMLFieldSetElement, CheckboxGroupProps>(
  (
    {
      label,
      options,
      value,
      defaultValue = [],
      onChange,
      disabled,
      error,
      required,
      layout,
      className
    },
    ref
  ) => {
    const [selectedValues, setSelectedValues] = useState<string[]>(defaultValue)
    const groupId = useId()
    const errorId = error ? `${groupId}-error` : undefined
    
    // Handle controlled/uncontrolled state
    const currentValues = value !== undefined ? value : selectedValues
    
    // Check if all options are selected
    const allSelected = options.every(opt => 
      currentValues.includes(opt.value) || opt.disabled
    )
    
    // Check if some options are selected
    const someSelected = options.some(opt => 
      currentValues.includes(opt.value)
    ) && !allSelected

    // Handle individual checkbox change
    const handleCheckboxChange = (optionValue: string, checked: boolean) => {
      let newValues: string[]
      
      if (checked) {
        newValues = [...currentValues, optionValue]
      } else {
        newValues = currentValues.filter(v => v !== optionValue)
      }
      
      if (value === undefined) {
        setSelectedValues(newValues)
      }
      
      onChange?.(newValues)
    }

    // Handle select all
    const handleSelectAll = (checked: boolean) => {
      const newValues = checked 
        ? options.filter(opt => !opt.disabled).map(opt => opt.value)
        : []
      
      if (value === undefined) {
        setSelectedValues(newValues)
      }
      
      onChange?.(newValues)
    }

    return (
      <fieldset
        ref={ref}
        className={cn('space-y-3', className)}
        aria-invalid={!!error}
        aria-describedby={error ? errorId : undefined}
      >
        {label && (
          <legend className="text-sm font-medium text-foreground">
            {label}
            {required && <span className="ml-1 text-destructive">*</span>}
          </legend>
        )}
        
        {/* Select all checkbox */}
        {options.length > 1 && (
          <Checkbox
            label="Select All"
            checked={allSelected}
            indeterminate={someSelected}
            onChange={(e) => handleSelectAll(e.target.checked)}
            disabled={disabled}
            className="mb-2"
          />
        )}
        
        {/* Options */}
        <div className={cn(groupVariants({ layout }))}>
          {options.map((option) => (
            <Checkbox
              key={option.value}
              label={option.label}
              description={option.description}
              checked={currentValues.includes(option.value)}
              onChange={(e) => handleCheckboxChange(option.value, e.target.checked)}
              disabled={disabled || option.disabled}
            />
          ))}
        </div>
        
        {error && (
          <p 
            id={errorId}
            className="text-xs text-destructive mt-1"
            role="alert"
          >
            {error}
          </p>
        )}
      </fieldset>
    )
  }
)

CheckboxGroup.displayName = 'CheckboxGroup'

export default Checkbox