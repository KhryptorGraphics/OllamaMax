import React, { useState, useRef, useEffect, useCallback, createContext, useContext } from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../../../utils/cn';

const tabsVariants = cva(
  'w-full',
  {
    variants: {
      orientation: {
        horizontal: 'flex flex-col',
        vertical: 'flex flex-row',
      },
    },
    defaultVariants: {
      orientation: 'horizontal',
    },
  }
);

const tabListVariants = cva(
  'flex',
  {
    variants: {
      orientation: {
        horizontal: 'flex-row border-b border-gray-200 space-x-1',
        vertical: 'flex-col border-r border-gray-200 space-y-1',
      },
      variant: {
        line: '',
        pill: 'p-1 bg-gray-100 rounded-lg',
        card: 'p-1',
      },
    },
    defaultVariants: {
      orientation: 'horizontal',
      variant: 'line',
    },
  }
);

const tabTriggerVariants = cva(
  'inline-flex items-center justify-center whitespace-nowrap transition-all focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        line: 'px-4 py-2 -mb-px border-b-2 border-transparent hover:text-gray-700 data-[state=active]:border-primary-500 data-[state=active]:text-primary-600',
        pill: 'px-4 py-2 rounded-md hover:bg-gray-200 data-[state=active]:bg-white data-[state=active]:shadow-sm',
        card: 'px-4 py-2 rounded-t-lg border border-transparent hover:bg-gray-50 data-[state=active]:bg-white data-[state=active]:border-gray-200 data-[state=active]:border-b-white',
      },
      size: {
        sm: 'text-sm h-8',
        md: 'text-base h-10',
        lg: 'text-lg h-12',
      },
    },
    defaultVariants: {
      variant: 'line',
      size: 'md',
    },
  }
);

const tabContentVariants = cva(
  'focus:outline-none',
  {
    variants: {
      orientation: {
        horizontal: 'w-full',
        vertical: 'flex-1 pl-4',
      },
    },
    defaultVariants: {
      orientation: 'horizontal',
    },
  }
);

interface TabsContextValue {
  value: string;
  onValueChange: (value: string) => void;
  orientation: 'horizontal' | 'vertical';
  variant: 'line' | 'pill' | 'card';
  size: 'sm' | 'md' | 'lg';
}

const TabsContext = createContext<TabsContextValue | undefined>(undefined);

const useTabsContext = () => {
  const context = useContext(TabsContext);
  if (!context) {
    throw new Error('Tabs components must be used within a Tabs provider');
  }
  return context;
};

export interface TabsProps extends VariantProps<typeof tabsVariants> {
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
  orientation?: 'horizontal' | 'vertical';
  variant?: 'line' | 'pill' | 'card';
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  children: React.ReactNode;
}

export const Tabs = React.forwardRef<HTMLDivElement, TabsProps>(
  (
    {
      value: controlledValue,
      defaultValue,
      onValueChange,
      orientation = 'horizontal',
      variant = 'line',
      size = 'md',
      className,
      children,
    },
    ref
  ) => {
    const [internalValue, setInternalValue] = useState(defaultValue || '');
    const value = controlledValue !== undefined ? controlledValue : internalValue;
    const isControlled = controlledValue !== undefined;

    const handleValueChange = useCallback((newValue: string) => {
      if (!isControlled) {
        setInternalValue(newValue);
      }
      onValueChange?.(newValue);
    }, [isControlled, onValueChange]);

    return (
      <TabsContext.Provider
        value={{
          value,
          onValueChange: handleValueChange,
          orientation: orientation || 'horizontal',
          variant: variant || 'line',
          size: size || 'md',
        }}
      >
        <div
          ref={ref}
          className={cn(tabsVariants({ orientation }), className)}
          data-orientation={orientation}
        >
          {children}
        </div>
      </TabsContext.Provider>
    );
  }
);

Tabs.displayName = 'Tabs';

export interface TabsListProps {
  className?: string;
  children: React.ReactNode;
  'aria-label'?: string;
}

export const TabsList = React.forwardRef<HTMLDivElement, TabsListProps>(
  ({ className, children, 'aria-label': ariaLabel }, ref) => {
    const { orientation, variant } = useTabsContext();
    const listRef = useRef<HTMLDivElement>(null);
    const [indicatorStyle, setIndicatorStyle] = useState<React.CSSProperties>({});

    useEffect(() => {
      const updateIndicator = () => {
        if (variant !== 'line' || !listRef.current) return;

        const activeTab = listRef.current.querySelector('[data-state="active"]');
        if (activeTab) {
          const rect = activeTab.getBoundingClientRect();
          const listRect = listRef.current.getBoundingClientRect();

          if (orientation === 'horizontal') {
            setIndicatorStyle({
              left: rect.left - listRect.left,
              width: rect.width,
              bottom: 0,
              height: 2,
            });
          } else {
            setIndicatorStyle({
              top: rect.top - listRect.top,
              height: rect.height,
              right: 0,
              width: 2,
            });
          }
        }
      };

      updateIndicator();
      window.addEventListener('resize', updateIndicator);
      return () => window.removeEventListener('resize', updateIndicator);
    }, [orientation, variant]);

    const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
      const triggers = Array.from(
        listRef.current?.querySelectorAll('[role="tab"]:not([disabled])') || []
      );
      const currentIndex = triggers.findIndex(trigger => 
        trigger === document.activeElement
      );

      let nextIndex = currentIndex;
      const isHorizontal = orientation === 'horizontal';
      const nextKey = isHorizontal ? 'ArrowRight' : 'ArrowDown';
      const prevKey = isHorizontal ? 'ArrowLeft' : 'ArrowUp';

      switch (e.key) {
        case nextKey:
          e.preventDefault();
          nextIndex = currentIndex < triggers.length - 1 ? currentIndex + 1 : 0;
          break;
        case prevKey:
          e.preventDefault();
          nextIndex = currentIndex > 0 ? currentIndex - 1 : triggers.length - 1;
          break;
        case 'Home':
          e.preventDefault();
          nextIndex = 0;
          break;
        case 'End':
          e.preventDefault();
          nextIndex = triggers.length - 1;
          break;
        default:
          return;
      }

      const nextTrigger = triggers[nextIndex] as HTMLElement;
      nextTrigger?.focus();
      nextTrigger?.click();
    }, [orientation]);

    return (
      <div
        ref={ref || listRef}
        role="tablist"
        aria-label={ariaLabel}
        aria-orientation={orientation}
        className={cn(tabListVariants({ orientation, variant }), 'relative', className)}
        onKeyDown={handleKeyDown}
      >
        {children}
        {variant === 'line' && (
          <div
            className="absolute bg-primary-500 transition-all duration-200"
            style={indicatorStyle}
          />
        )}
      </div>
    );
  }
);

TabsList.displayName = 'TabsList';

export interface TabsTriggerProps {
  value: string;
  disabled?: boolean;
  className?: string;
  children: React.ReactNode;
  icon?: React.ReactNode;
}

export const TabsTrigger = React.forwardRef<HTMLButtonElement, TabsTriggerProps>(
  ({ value: triggerValue, disabled, className, children, icon }, ref) => {
    const { value, onValueChange, variant, size } = useTabsContext();
    const isActive = value === triggerValue;

    return (
      <button
        ref={ref}
        type="button"
        role="tab"
        aria-selected={isActive}
        aria-controls={`tabpanel-${triggerValue}`}
        data-state={isActive ? 'active' : 'inactive'}
        disabled={disabled}
        className={cn(tabTriggerVariants({ variant, size }), className)}
        onClick={() => onValueChange(triggerValue)}
        tabIndex={isActive ? 0 : -1}
      >
        {icon && <span className="mr-2">{icon}</span>}
        {children}
      </button>
    );
  }
);

TabsTrigger.displayName = 'TabsTrigger';

export interface TabsContentProps {
  value: string;
  className?: string;
  children: React.ReactNode;
  lazy?: boolean;
  forceMount?: boolean;
}

export const TabsContent = React.forwardRef<HTMLDivElement, TabsContentProps>(
  ({ value: contentValue, className, children, lazy = false, forceMount = false }, ref) => {
    const { value, orientation } = useTabsContext();
    const isActive = value === contentValue;
    const [hasBeenActive, setHasBeenActive] = useState(isActive);

    useEffect(() => {
      if (isActive && !hasBeenActive) {
        setHasBeenActive(true);
      }
    }, [isActive, hasBeenActive]);

    // Don't render if not active and forceMount is false
    if (!forceMount && !isActive) {
      return null;
    }

    // For lazy loading, don't render until it has been active at least once
    if (lazy && !hasBeenActive && !isActive) {
      return null;
    }

    return (
      <div
        ref={ref}
        role="tabpanel"
        aria-labelledby={`tab-${contentValue}`}
        id={`tabpanel-${contentValue}`}
        data-state={isActive ? 'active' : 'inactive'}
        hidden={!isActive}
        tabIndex={0}
        className={cn(
          tabContentVariants({ orientation }),
          'mt-4',
          className
        )}
      >
        {children}
      </div>
    );
  }
);

TabsContent.displayName = 'TabsContent';

// Compound component exports
export const TabsCompound = Object.assign(Tabs, {
  List: TabsList,
  Trigger: TabsTrigger,
  Content: TabsContent,
});