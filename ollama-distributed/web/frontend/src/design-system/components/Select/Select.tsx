import React, { useState, useRef, useEffect, useMemo, useCallback } from 'react';
import { ChevronDown, Check, X, Search } from 'lucide-react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../../../utils/cn';

const selectVariants = cva(
  'relative inline-block w-full text-left',
  {
    variants: {
      size: {
        sm: 'text-sm',
        md: 'text-base',
        lg: 'text-lg',
      },
    },
    defaultVariants: {
      size: 'md',
    },
  }
);

const triggerVariants = cva(
  'flex items-center justify-between w-full px-3 py-2 bg-white border rounded-lg shadow-sm transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500',
  {
    variants: {
      size: {
        sm: 'h-8 text-sm',
        md: 'h-10 text-base',
        lg: 'h-12 text-lg',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed bg-gray-50',
        false: 'hover:border-gray-400 cursor-pointer',
      },
      error: {
        true: 'border-red-500 focus:ring-red-500',
        false: 'border-gray-300',
      },
    },
    defaultVariants: {
      size: 'md',
      disabled: false,
      error: false,
    },
  }
);

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
  group?: string;
  icon?: React.ReactNode;
  description?: string;
}

export interface SelectProps extends VariantProps<typeof selectVariants> {
  options: SelectOption[];
  value?: string | string[];
  defaultValue?: string | string[];
  onChange?: (value: string | string[]) => void;
  placeholder?: string;
  disabled?: boolean;
  error?: boolean;
  multiple?: boolean;
  searchable?: boolean;
  clearable?: boolean;
  loading?: boolean;
  maxHeight?: number;
  renderOption?: (option: SelectOption, isSelected: boolean) => React.ReactNode;
  className?: string;
  name?: string;
  id?: string;
  'aria-label'?: string;
  'aria-describedby'?: string;
}

export const Select = React.forwardRef<HTMLDivElement, SelectProps>(
  (
    {
      options,
      value: controlledValue,
      defaultValue,
      onChange,
      placeholder = 'Select an option',
      disabled = false,
      error = false,
      multiple = false,
      searchable = false,
      clearable = false,
      loading = false,
      maxHeight = 300,
      renderOption,
      size = 'md',
      className,
      name,
      id,
      'aria-label': ariaLabel,
      'aria-describedby': ariaDescribedBy,
    },
    ref
  ) => {
    const [isOpen, setIsOpen] = useState(false);
    const [searchTerm, setSearchTerm] = useState('');
    const [highlightedIndex, setHighlightedIndex] = useState(-1);
    const [internalValue, setInternalValue] = useState<string | string[]>(
      defaultValue || (multiple ? [] : '')
    );

    const containerRef = useRef<HTMLDivElement>(null);
    const triggerRef = useRef<HTMLButtonElement>(null);
    const searchInputRef = useRef<HTMLInputElement>(null);
    const optionsRef = useRef<HTMLDivElement>(null);

    const value = controlledValue !== undefined ? controlledValue : internalValue;
    const isControlled = controlledValue !== undefined;

    // Filter options based on search term
    const filteredOptions = useMemo(() => {
      if (!searchTerm) return options;
      return options.filter(option =>
        option.label.toLowerCase().includes(searchTerm.toLowerCase()) ||
        option.description?.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }, [options, searchTerm]);

    // Group options
    const groupedOptions = useMemo(() => {
      const groups: Record<string, SelectOption[]> = {};
      const ungrouped: SelectOption[] = [];

      filteredOptions.forEach(option => {
        if (option.group) {
          if (!groups[option.group]) {
            groups[option.group] = [];
          }
          groups[option.group].push(option);
        } else {
          ungrouped.push(option);
        }
      });

      return { groups, ungrouped };
    }, [filteredOptions]);

    // Get selected options
    const selectedOptions = useMemo(() => {
      if (multiple) {
        const values = Array.isArray(value) ? value : [];
        return options.filter(option => values.includes(option.value));
      } else {
        return options.filter(option => option.value === value);
      }
    }, [options, value, multiple]);

    // Handle value change
    const handleValueChange = useCallback((newValue: string | string[]) => {
      if (!isControlled) {
        setInternalValue(newValue);
      }
      onChange?.(newValue);
    }, [isControlled, onChange]);

    // Handle option selection
    const handleOptionSelect = useCallback((option: SelectOption) => {
      if (option.disabled) return;

      if (multiple) {
        const currentValues = Array.isArray(value) ? value : [];
        const newValues = currentValues.includes(option.value)
          ? currentValues.filter(v => v !== option.value)
          : [...currentValues, option.value];
        handleValueChange(newValues);
      } else {
        handleValueChange(option.value);
        setIsOpen(false);
      }

      setSearchTerm('');
    }, [value, multiple, handleValueChange]);

    // Handle clear
    const handleClear = useCallback((e: React.MouseEvent) => {
      e.stopPropagation();
      handleValueChange(multiple ? [] : '');
    }, [multiple, handleValueChange]);

    // Handle keyboard navigation
    const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
      if (disabled) return;

      switch (e.key) {
        case 'Enter':
        case ' ':
          if (!isOpen) {
            setIsOpen(true);
          } else if (highlightedIndex >= 0) {
            handleOptionSelect(filteredOptions[highlightedIndex]);
          }
          e.preventDefault();
          break;
        case 'Escape':
          setIsOpen(false);
          setSearchTerm('');
          triggerRef.current?.focus();
          break;
        case 'ArrowDown':
          e.preventDefault();
          if (!isOpen) {
            setIsOpen(true);
          } else {
            setHighlightedIndex(prev =>
              prev < filteredOptions.length - 1 ? prev + 1 : 0
            );
          }
          break;
        case 'ArrowUp':
          e.preventDefault();
          if (!isOpen) {
            setIsOpen(true);
          } else {
            setHighlightedIndex(prev =>
              prev > 0 ? prev - 1 : filteredOptions.length - 1
            );
          }
          break;
        case 'Home':
          e.preventDefault();
          setHighlightedIndex(0);
          break;
        case 'End':
          e.preventDefault();
          setHighlightedIndex(filteredOptions.length - 1);
          break;
      }
    }, [disabled, isOpen, highlightedIndex, filteredOptions, handleOptionSelect]);

    // Handle click outside
    useEffect(() => {
      const handleClickOutside = (event: MouseEvent) => {
        if (
          containerRef.current &&
          !containerRef.current.contains(event.target as Node)
        ) {
          setIsOpen(false);
          setSearchTerm('');
        }
      };

      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    // Focus search input when opened
    useEffect(() => {
      if (isOpen && searchable) {
        searchInputRef.current?.focus();
      }
    }, [isOpen, searchable]);

    // Scroll highlighted option into view
    useEffect(() => {
      if (highlightedIndex >= 0 && optionsRef.current) {
        const highlightedOption = optionsRef.current.children[highlightedIndex] as HTMLElement;
        if (highlightedOption) {
          highlightedOption.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
        }
      }
    }, [highlightedIndex]);

    const renderSelectedValue = () => {
      if (loading) {
        return <span className="text-gray-400">Loading...</span>;
      }

      if (selectedOptions.length === 0) {
        return <span className="text-gray-400">{placeholder}</span>;
      }

      if (multiple) {
        return (
          <div className="flex flex-wrap gap-1">
            {selectedOptions.map(option => (
              <span
                key={option.value}
                className="inline-flex items-center px-2 py-1 text-xs bg-primary-100 text-primary-700 rounded-md"
              >
                {option.label}
                {!disabled && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleOptionSelect(option);
                    }}
                    className="ml-1 hover:text-primary-900"
                    aria-label={`Remove ${option.label}`}
                  >
                    <X className="w-3 h-3" />
                  </button>
                )}
              </span>
            ))}
          </div>
        );
      }

      return (
        <span className="truncate">
          {selectedOptions[0]?.icon && (
            <span className="inline-block mr-2">{selectedOptions[0].icon}</span>
          )}
          {selectedOptions[0]?.label}
        </span>
      );
    };

    const renderOptionItem = (option: SelectOption, index: number) => {
      const isSelected = multiple
        ? (Array.isArray(value) ? value : []).includes(option.value)
        : value === option.value;
      const isHighlighted = index === highlightedIndex;

      if (renderOption) {
        return renderOption(option, isSelected);
      }

      return (
        <div
          key={option.value}
          className={cn(
            'px-3 py-2 cursor-pointer transition-colors',
            {
              'bg-primary-50 text-primary-700': isHighlighted,
              'bg-primary-100 text-primary-800': isSelected && !isHighlighted,
              'hover:bg-gray-50': !isSelected && !isHighlighted && !option.disabled,
              'opacity-50 cursor-not-allowed': option.disabled,
            }
          )}
          onClick={() => handleOptionSelect(option)}
          onMouseEnter={() => setHighlightedIndex(index)}
          role="option"
          aria-selected={isSelected}
          aria-disabled={option.disabled}
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center flex-1">
              {multiple && (
                <div className="mr-2">
                  <input
                    type="checkbox"
                    checked={isSelected}
                    onChange={() => {}}
                    disabled={option.disabled}
                    className="w-4 h-4 text-primary-600 border-gray-300 rounded focus:ring-primary-500"
                    aria-hidden="true"
                  />
                </div>
              )}
              {option.icon && (
                <span className="mr-2 flex-shrink-0">{option.icon}</span>
              )}
              <div className="flex-1">
                <div className="font-medium">{option.label}</div>
                {option.description && (
                  <div className="text-xs text-gray-500 mt-0.5">
                    {option.description}
                  </div>
                )}
              </div>
            </div>
            {!multiple && isSelected && (
              <Check className="w-4 h-4 text-primary-600 flex-shrink-0 ml-2" />
            )}
          </div>
        </div>
      );
    };

    return (
      <div
        ref={ref || containerRef}
        className={cn(selectVariants({ size }), className)}
      >
        <button
          ref={triggerRef}
          type="button"
          className={cn(triggerVariants({ size, disabled, error }))}
          onClick={() => !disabled && setIsOpen(!isOpen)}
          onKeyDown={handleKeyDown}
          disabled={disabled}
          aria-haspopup="listbox"
          aria-expanded={isOpen}
          aria-label={ariaLabel}
          aria-describedby={ariaDescribedBy}
          id={id}
        >
          <div className="flex-1 text-left overflow-hidden">
            {renderSelectedValue()}
          </div>
          <div className="flex items-center ml-2 space-x-1">
            {clearable && selectedOptions.length > 0 && !disabled && (
              <button
                onClick={handleClear}
                className="p-0.5 hover:bg-gray-100 rounded"
                aria-label="Clear selection"
              >
                <X className="w-4 h-4 text-gray-400" />
              </button>
            )}
            <ChevronDown
              className={cn(
                'w-4 h-4 text-gray-400 transition-transform',
                isOpen && 'transform rotate-180'
              )}
            />
          </div>
        </button>

        {isOpen && (
          <div className="absolute z-50 w-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg">
            {searchable && (
              <div className="p-2 border-b border-gray-200">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                  <input
                    ref={searchInputRef}
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                    placeholder="Search..."
                    aria-label="Search options"
                  />
                </div>
              </div>
            )}

            <div
              ref={optionsRef}
              className="overflow-auto"
              style={{ maxHeight: `${maxHeight}px` }}
              role="listbox"
              aria-multiselectable={multiple}
            >
              {loading ? (
                <div className="px-3 py-8 text-center text-gray-500">
                  Loading options...
                </div>
              ) : filteredOptions.length === 0 ? (
                <div className="px-3 py-8 text-center text-gray-500">
                  No options found
                </div>
              ) : (
                <>
                  {groupedOptions.ungrouped.length > 0 && (
                    <>
                      {groupedOptions.ungrouped.map((option, index) =>
                        renderOptionItem(option, index)
                      )}
                    </>
                  )}
                  {Object.entries(groupedOptions.groups).map(([groupName, groupOptions]) => (
                    <div key={groupName}>
                      <div className="px-3 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider bg-gray-50">
                        {groupName}
                      </div>
                      {groupOptions.map((option, index) =>
                        renderOptionItem(
                          option,
                          groupedOptions.ungrouped.length +
                            Object.values(groupedOptions.groups)
                              .slice(0, Object.keys(groupedOptions.groups).indexOf(groupName))
                              .reduce((acc, g) => acc + g.length, 0) +
                            index
                        )
                      )}
                    </div>
                  ))}
                </>
              )}
            </div>
          </div>
        )}

        {/* Hidden input for form submission */}
        {name && (
          <input
            type="hidden"
            name={name}
            value={multiple ? JSON.stringify(value) : String(value)}
          />
        )}
      </div>
    );
  }
);

Select.displayName = 'Select';