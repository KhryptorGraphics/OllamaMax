import React, { forwardRef, InputHTMLAttributes, useId, useState } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { Circle } from 'lucide-react'
import { cn } from '@/utils/cn'

// Radio button variants using design tokens
const radioVariants = cva(
  [
    'shrink-0 rounded-full border transition-smooth',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
    'disabled:cursor-not-allowed disabled:opacity-50',
    'data-[state=checked]:border-primary'
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
          'data-[state=checked]:border-destructive'
        ],
        success: [
          'border-green-500',
          'hover:border-green-400',
          'data-[state=checked]:border-green-500'
        ],
        warning: [
          'border-yellow-500',
          'hover:border-yellow-400',
          'data-[state=checked]:border-yellow-500'
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

// Radio inner circle variants
const radioInnerVariants = cva(
  'rounded-full transition-smooth',
  {
    variants: {
      variant: {
        default: 'bg-primary',
        error: 'bg-destructive',
        success: 'bg-green-500',
        warning: 'bg-yellow-500'
      },
      size: {
        sm: 'h-1.5 w-1.5',
        md: 'h-2 w-2',
        lg: 'h-2.5 w-2.5'
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

// Radio group layout variants
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

export interface RadioProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type' | 'size'>,
    VariantProps<typeof radioVariants> {
  /** Label for the radio button */
  label?: string
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
}

/**
 * Radio button component with comprehensive features and accessibility
 * 
 * Features:
 * - Multiple variants and sizes
 * - Custom colors
 * - Full keyboard navigation
 * - ARIA attributes for accessibility
 * - Description text support
 */
export const Radio = forwardRef<HTMLInputElement, RadioProps>(
  (
    {
      className,
      labelClassName,
      variant,
      size,
      label,
      checkedColor,
      description,
      error,
      hideLabel = false,
      checked,
      disabled,
      id,
      ...props
    },
    ref
  ) => {
    const radioId = useId()
    const finalId = id || radioId
    const descriptionId = description ? `${finalId}-description` : undefined
    const errorId = error ? `${finalId}-error` : undefined

    // Determine state for styling
    const state = checked ? 'checked' : 'unchecked'
    
    // Determine final variant
    const finalVariant = error ? 'error' : variant

    // Build aria-describedby
    const ariaDescribedBy = [
      descriptionId,
      errorId
    ].filter(Boolean).join(' ') || undefined

    return (
      <div className="flex items-start space-x-2">
        <div className="relative flex items-center">
          <input
            type="radio"
            className="sr-only peer"
            ref={ref}
            id={finalId}
            checked={checked}
            disabled={disabled}
            aria-invalid={!!error}
            aria-describedby={ariaDescribedBy}
            {...props}
          />
          <div
            className={cn(
              radioVariants({ variant: finalVariant, size }),
              'flex items-center justify-center',
              className
            )}
            data-state={state}
            style={checked && checkedColor ? { borderColor: checkedColor } : undefined}
          >
            {checked && (
              <div 
                className={cn(radioInnerVariants({ variant: finalVariant, size }))}
                style={checkedColor ? { backgroundColor: checkedColor } : undefined}
              />
            )}
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

Radio.displayName = 'Radio'

export interface RadioGroupProps extends VariantProps<typeof groupVariants> {
  /** Group label */
  label?: string
  /** Array of radio options */
  options: Array<{
    value: string
    label: string
    description?: string
    disabled?: boolean
  }>
  /** Selected value */
  value?: string
  /** Default selected value */
  defaultValue?: string
  /** Callback when selection changes */
  onChange?: (value: string) => void
  /** Name attribute for the radio group */
  name?: string
  /** Whether the group is disabled */
  disabled?: boolean
  /** Error message */
  error?: string
  /** Required field */
  required?: boolean
  /** Radio button size */
  size?: 'sm' | 'md' | 'lg'
  /** Radio button variant */
  variant?: 'default' | 'error' | 'success' | 'warning'
  /** Additional CSS classes */
  className?: string
}

/**
 * RadioGroup component for managing radio button groups
 * 
 * Features:
 * - Group layouts (horizontal, vertical, grid)
 * - Controlled/uncontrolled modes
 * - Mutual exclusion
 * - Keyboard navigation (arrow keys)
 * - Accessibility support with fieldset and legend
 */
export const RadioGroup = forwardRef<HTMLFieldSetElement, RadioGroupProps>(
  (
    {
      label,
      options,
      value,
      defaultValue,
      onChange,
      name,
      disabled,
      error,
      required,
      size,
      variant,
      layout,
      className
    },
    ref
  ) => {
    const [selectedValue, setSelectedValue] = useState<string | undefined>(defaultValue)
    const groupId = useId()
    const groupName = name || groupId
    const errorId = error ? `${groupId}-error` : undefined
    
    // Handle controlled/uncontrolled state
    const currentValue = value !== undefined ? value : selectedValue

    // Handle radio change
    const handleRadioChange = (optionValue: string) => {
      if (value === undefined) {
        setSelectedValue(optionValue)
      }
      onChange?.(optionValue)
    }

    // Handle keyboard navigation
    const handleKeyDown = (e: React.KeyboardEvent, currentIndex: number) => {
      const enabledOptions = options.filter(opt => !opt.disabled)
      const currentEnabledIndex = enabledOptions.findIndex(opt => opt.value === options[currentIndex].value)
      
      let nextIndex = currentEnabledIndex
      
      switch (e.key) {
        case 'ArrowDown':
        case 'ArrowRight':
          e.preventDefault()
          nextIndex = (currentEnabledIndex + 1) % enabledOptions.length
          break
        case 'ArrowUp':
        case 'ArrowLeft':
          e.preventDefault()
          nextIndex = currentEnabledIndex === 0 ? enabledOptions.length - 1 : currentEnabledIndex - 1
          break
        case 'Home':
          e.preventDefault()
          nextIndex = 0
          break
        case 'End':
          e.preventDefault()
          nextIndex = enabledOptions.length - 1
          break
        default:
          return
      }
      
      const nextOption = enabledOptions[nextIndex]
      if (nextOption) {
        handleRadioChange(nextOption.value)
        // Focus the next radio button
        const nextElement = document.querySelector(`input[name="${groupName}"][value="${nextOption.value}"]`) as HTMLInputElement
        nextElement?.focus()
      }
    }

    return (
      <fieldset
        ref={ref}
        className={cn('space-y-3', className)}
        role="radiogroup"
        aria-invalid={!!error}
        aria-describedby={error ? errorId : undefined}
        aria-required={required}
      >
        {label && (
          <legend className="text-sm font-medium text-foreground">
            {label}
            {required && <span className="ml-1 text-destructive">*</span>}
          </legend>
        )}
        
        {/* Options */}
        <div className={cn(groupVariants({ layout }))}>
          {options.map((option, index) => (
            <Radio
              key={option.value}
              name={groupName}
              value={option.value}
              label={option.label}
              description={option.description}
              checked={currentValue === option.value}
              onChange={() => handleRadioChange(option.value)}
              onKeyDown={(e) => handleKeyDown(e, index)}
              disabled={disabled || option.disabled}
              size={size}
              variant={error ? 'error' : variant}
              tabIndex={currentValue === option.value ? 0 : -1}
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

RadioGroup.displayName = 'RadioGroup'

export default Radio