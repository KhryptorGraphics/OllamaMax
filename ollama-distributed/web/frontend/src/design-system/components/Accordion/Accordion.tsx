import React, { useState, useRef, useEffect, useCallback, createContext, useContext } from 'react';
import { ChevronDown, ChevronRight, Plus, Minus } from 'lucide-react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../../../utils/cn';

const accordionVariants = cva(
  'w-full',
  {
    variants: {
      variant: {
        default: 'border border-gray-200 rounded-lg',
        bordered: 'space-y-2',
        ghost: '',
        filled: 'bg-gray-50 rounded-lg p-2 space-y-2',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

const accordionItemVariants = cva(
  'w-full',
  {
    variants: {
      variant: {
        default: 'border-b last:border-b-0',
        bordered: 'border border-gray-200 rounded-lg',
        ghost: 'border-b border-gray-100',
        filled: 'bg-white rounded-lg shadow-sm',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

const accordionTriggerVariants = cva(
  'flex items-center justify-between w-full text-left transition-all hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-inset',
  {
    variants: {
      variant: {
        default: 'px-4 py-3',
        bordered: 'px-4 py-3 rounded-t-lg',
        ghost: 'px-2 py-3',
        filled: 'px-4 py-3 rounded-t-lg',
      },
      size: {
        sm: 'text-sm',
        md: 'text-base',
        lg: 'text-lg',
      },
      expanded: {
        true: 'font-medium',
        false: '',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'md',
      expanded: false,
    },
  }
);

const accordionContentVariants = cva(
  'overflow-hidden transition-all duration-200 ease-in-out',
  {
    variants: {
      variant: {
        default: 'px-4',
        bordered: 'px-4',
        ghost: 'px-2',
        filled: 'px-4',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

interface AccordionContextValue {
  expandedItems: Set<string>;
  toggleItem: (value: string) => void;
  variant: 'default' | 'bordered' | 'ghost' | 'filled';
  size: 'sm' | 'md' | 'lg';
  iconPosition: 'left' | 'right';
  iconType: 'chevron' | 'arrow' | 'plus';
  animated: boolean;
}

const AccordionContext = createContext<AccordionContextValue | undefined>(undefined);

const useAccordionContext = () => {
  const context = useContext(AccordionContext);
  if (!context) {
    throw new Error('Accordion components must be used within an Accordion provider');
  }
  return context;
};

export interface AccordionProps extends VariantProps<typeof accordionVariants> {
  type?: 'single' | 'multiple';
  value?: string | string[];
  defaultValue?: string | string[];
  onValueChange?: (value: string | string[]) => void;
  variant?: 'default' | 'bordered' | 'ghost' | 'filled';
  size?: 'sm' | 'md' | 'lg';
  iconPosition?: 'left' | 'right';
  iconType?: 'chevron' | 'arrow' | 'plus';
  animated?: boolean;
  className?: string;
  children: React.ReactNode;
}

export const Accordion = React.forwardRef<HTMLDivElement, AccordionProps>(
  (
    {
      type = 'single',
      value: controlledValue,
      defaultValue,
      onValueChange,
      variant = 'default',
      size = 'md',
      iconPosition = 'right',
      iconType = 'chevron',
      animated = true,
      className,
      children,
    },
    ref
  ) => {
    const [internalExpandedItems, setInternalExpandedItems] = useState<Set<string>>(() => {
      if (defaultValue) {
        return new Set(Array.isArray(defaultValue) ? defaultValue : [defaultValue]);
      }
      return new Set();
    });

    const expandedItems = controlledValue !== undefined
      ? new Set(Array.isArray(controlledValue) ? controlledValue : [controlledValue])
      : internalExpandedItems;

    const isControlled = controlledValue !== undefined;

    const toggleItem = useCallback((value: string) => {
      const newExpandedItems = new Set(expandedItems);

      if (type === 'single') {
        if (newExpandedItems.has(value)) {
          newExpandedItems.clear();
        } else {
          newExpandedItems.clear();
          newExpandedItems.add(value);
        }
      } else {
        if (newExpandedItems.has(value)) {
          newExpandedItems.delete(value);
        } else {
          newExpandedItems.add(value);
        }
      }

      if (!isControlled) {
        setInternalExpandedItems(newExpandedItems);
      }

      const newValue = type === 'single'
        ? (newExpandedItems.size > 0 ? Array.from(newExpandedItems)[0] : '')
        : Array.from(newExpandedItems);

      onValueChange?.(newValue);
    }, [expandedItems, type, isControlled, onValueChange]);

    return (
      <AccordionContext.Provider
        value={{
          expandedItems,
          toggleItem,
          variant: variant || 'default',
          size: size || 'md',
          iconPosition: iconPosition || 'right',
          iconType: iconType || 'chevron',
          animated: animated !== false,
        }}
      >
        <div
          ref={ref}
          className={cn(accordionVariants({ variant }), className)}
          data-variant={variant}
        >
          {children}
        </div>
      </AccordionContext.Provider>
    );
  }
);

Accordion.displayName = 'Accordion';

export interface AccordionItemProps {
  value: string;
  className?: string;
  children: React.ReactNode;
}

export const AccordionItem = React.forwardRef<HTMLDivElement, AccordionItemProps>(
  ({ value, className, children }, ref) => {
    const { variant } = useAccordionContext();

    return (
      <div
        ref={ref}
        className={cn(accordionItemVariants({ variant }), className)}
        data-accordion-item
        data-value={value}
      >
        {children}
      </div>
    );
  }
);

AccordionItem.displayName = 'AccordionItem';

export interface AccordionTriggerProps {
  value: string;
  className?: string;
  children: React.ReactNode;
  icon?: React.ReactNode;
  disabled?: boolean;
}

export const AccordionTrigger = React.forwardRef<HTMLButtonElement, AccordionTriggerProps>(
  ({ value, className, children, icon: customIcon, disabled }, ref) => {
    const { expandedItems, toggleItem, variant, size, iconPosition, iconType, animated } = useAccordionContext();
    const isExpanded = expandedItems.has(value);

    const renderIcon = () => {
      if (customIcon) {
        return (
          <span
            className={cn(
              'transition-transform duration-200',
              animated && isExpanded && 'rotate-180'
            )}
          >
            {customIcon}
          </span>
        );
      }

      const iconClass = cn(
        'w-4 h-4 transition-transform duration-200',
        animated && isExpanded && (iconType === 'chevron' ? 'rotate-180' : 'rotate-90')
      );

      switch (iconType) {
        case 'arrow':
          return <ChevronRight className={iconClass} />;
        case 'plus':
          return isExpanded ? (
            <Minus className="w-4 h-4" />
          ) : (
            <Plus className="w-4 h-4" />
          );
        case 'chevron':
        default:
          return <ChevronDown className={iconClass} />;
      }
    };

    return (
      <button
        ref={ref}
        type="button"
        className={cn(
          accordionTriggerVariants({ variant, size, expanded: isExpanded }),
          disabled && 'opacity-50 cursor-not-allowed',
          className
        )}
        onClick={() => !disabled && toggleItem(value)}
        disabled={disabled}
        aria-expanded={isExpanded}
        aria-controls={`accordion-content-${value}`}
      >
        {iconPosition === 'left' && (
          <span className="mr-2">{renderIcon()}</span>
        )}
        <span className="flex-1 text-left">{children}</span>
        {iconPosition === 'right' && (
          <span className="ml-2">{renderIcon()}</span>
        )}
      </button>
    );
  }
);

AccordionTrigger.displayName = 'AccordionTrigger';

export interface AccordionContentProps {
  value: string;
  className?: string;
  children: React.ReactNode;
}

export const AccordionContent = React.forwardRef<HTMLDivElement, AccordionContentProps>(
  ({ value, className, children }, ref) => {
    const { expandedItems, variant, animated } = useAccordionContext();
    const isExpanded = expandedItems.has(value);
    const contentRef = useRef<HTMLDivElement>(null);
    const [height, setHeight] = useState<number | undefined>(isExpanded ? undefined : 0);

    useEffect(() => {
      if (!animated) {
        setHeight(undefined);
        return;
      }

      if (contentRef.current) {
        if (isExpanded) {
          const contentHeight = contentRef.current.scrollHeight;
          setHeight(contentHeight);
          
          // After animation completes, remove fixed height to allow dynamic content
          const timeout = setTimeout(() => {
            setHeight(undefined);
          }, 200);
          
          return () => clearTimeout(timeout);
        } else {
          // First set to current height, then to 0 for smooth animation
          const currentHeight = contentRef.current.scrollHeight;
          setHeight(currentHeight);
          
          requestAnimationFrame(() => {
            setHeight(0);
          });
        }
      }
    }, [isExpanded, animated]);

    return (
      <div
        ref={ref}
        id={`accordion-content-${value}`}
        role="region"
        aria-labelledby={`accordion-trigger-${value}`}
        className={cn(
          accordionContentVariants({ variant }),
          !isExpanded && 'invisible',
          className
        )}
        style={{
          height: animated ? height : isExpanded ? 'auto' : 0,
        }}
      >
        <div ref={contentRef} className={cn('pb-4', !isExpanded && 'invisible')}>
          {children}
        </div>
      </div>
    );
  }
);

AccordionContent.displayName = 'AccordionContent';

// Compound component exports
export const AccordionCompound = Object.assign(Accordion, {
  Item: AccordionItem,
  Trigger: AccordionTrigger,
  Content: AccordionContent,
});