import React from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/utils/cn'

interface SelectOption {
  value: string
  label: string
}

interface SimpleSelectProps {
  value: string
  onValueChange: (value: string) => void
  options: SelectOption[]
  placeholder?: string
  className?: string
}

export const SimpleSelect: React.FC<SimpleSelectProps> = ({
  value,
  onValueChange,
  options,
  placeholder = 'Select...',
  className
}) => {
  return (
    <div className={cn('relative', className)}>
      <select
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
        className={cn(
          'w-full px-3 py-2 pr-8 text-sm bg-background border border-border rounded-md',
          'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          'appearance-none cursor-pointer'
        )}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <div className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
        <ChevronDown className="w-4 h-4 text-muted-foreground" />
      </div>
    </div>
  )
}

export default SimpleSelect