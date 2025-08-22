import React, { forwardRef, HTMLAttributes } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/utils/cn'

// Container variants
const containerVariants = cva(
  'w-full mx-auto',
  {
    variants: {
      size: {
        sm: 'max-w-screen-sm',      // 640px
        md: 'max-w-screen-md',      // 768px
        lg: 'max-w-screen-lg',      // 1024px
        xl: 'max-w-screen-xl',      // 1280px
        '2xl': 'max-w-screen-2xl',  // 1536px
        full: 'max-w-full',
        prose: 'max-w-prose'        // ~65ch for optimal reading
      },
      padding: {
        none: 'px-0',
        sm: 'px-4',
        md: 'px-6',
        lg: 'px-8',
        xl: 'px-12'
      }
    },
    defaultVariants: {
      size: 'xl',
      padding: 'md'
    }
  }
)

// Grid variants
const gridVariants = cva(
  'grid',
  {
    variants: {
      cols: {
        1: 'grid-cols-1',
        2: 'grid-cols-1 md:grid-cols-2',
        3: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
        4: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4',
        5: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5',
        6: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6',
        12: 'grid-cols-12'
      },
      gap: {
        none: 'gap-0',
        sm: 'gap-2',
        md: 'gap-4',
        lg: 'gap-6',
        xl: 'gap-8',
        '2xl': 'gap-12'
      }
    },
    defaultVariants: {
      cols: 1,
      gap: 'md'
    }
  }
)

// Flex variants
const flexVariants = cva(
  'flex',
  {
    variants: {
      direction: {
        row: 'flex-row',
        'row-reverse': 'flex-row-reverse',
        col: 'flex-col',
        'col-reverse': 'flex-col-reverse'
      },
      wrap: {
        nowrap: 'flex-nowrap',
        wrap: 'flex-wrap',
        'wrap-reverse': 'flex-wrap-reverse'
      },
      justify: {
        start: 'justify-start',
        end: 'justify-end',
        center: 'justify-center',
        between: 'justify-between',
        around: 'justify-around',
        evenly: 'justify-evenly'
      },
      align: {
        start: 'items-start',
        end: 'items-end',
        center: 'items-center',
        baseline: 'items-baseline',
        stretch: 'items-stretch'
      },
      gap: {
        none: 'gap-0',
        sm: 'gap-2',
        md: 'gap-4',
        lg: 'gap-6',
        xl: 'gap-8',
        '2xl': 'gap-12'
      }
    },
    defaultVariants: {
      direction: 'row',
      wrap: 'nowrap',
      justify: 'start',
      align: 'stretch',
      gap: 'md'
    }
  }
)

// Stack variants (vertical spacing)
const stackVariants = cva(
  'flex flex-col',
  {
    variants: {
      spacing: {
        none: 'space-y-0',
        xs: 'space-y-1',
        sm: 'space-y-2',
        md: 'space-y-4',
        lg: 'space-y-6',
        xl: 'space-y-8',
        '2xl': 'space-y-12'
      },
      align: {
        start: 'items-start',
        end: 'items-end',
        center: 'items-center',
        stretch: 'items-stretch'
      }
    },
    defaultVariants: {
      spacing: 'md',
      align: 'stretch'
    }
  }
)

// Container Component
export interface ContainerProps
  extends HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof containerVariants> {}

export const Container = forwardRef<HTMLDivElement, ContainerProps>(
  ({ className, size, padding, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(containerVariants({ size, padding }), className)}
      {...props}
    />
  )
)

Container.displayName = 'Container'

// Grid Component
export interface GridProps
  extends HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof gridVariants> {
  /** Responsive column configuration */
  responsive?: {
    sm?: number
    md?: number
    lg?: number
    xl?: number
    '2xl'?: number
  }
}

export const Grid = forwardRef<HTMLDivElement, GridProps>(
  ({ className, cols, gap, responsive, ...props }, ref) => {
    let responsiveClasses = ''
    
    if (responsive) {
      Object.entries(responsive).forEach(([breakpoint, columns]) => {
        const prefix = breakpoint === 'sm' ? '' : `${breakpoint}:`
        responsiveClasses += ` ${prefix}grid-cols-${columns}`
      })
    }

    return (
      <div
        ref={ref}
        className={cn(
          gridVariants({ cols: responsive ? undefined : cols, gap }),
          responsive && responsiveClasses,
          className
        )}
        {...props}
      />
    )
  }
)

Grid.displayName = 'Grid'

// Flex Component
export interface FlexProps
  extends HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof flexVariants> {}

export const Flex = forwardRef<HTMLDivElement, FlexProps>(
  ({ className, direction, wrap, justify, align, gap, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        flexVariants({ direction, wrap, justify, align, gap }),
        className
      )}
      {...props}
    />
  )
)

Flex.displayName = 'Flex'

// Stack Component (Vertical Layout)
export interface StackProps
  extends HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof stackVariants> {}

export const Stack = forwardRef<HTMLDivElement, StackProps>(
  ({ className, spacing, align, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(stackVariants({ spacing, align }), className)}
      {...props}
    />
  )
)

Stack.displayName = 'Stack'

// Box Component (Generic Container)
export interface BoxProps extends HTMLAttributes<HTMLDivElement> {
  /** Element to render as */
  as?: keyof JSX.IntrinsicElements
}

export const Box = forwardRef<HTMLDivElement, BoxProps>(
  ({ as: Component = 'div', ...props }, ref) => (
    <Component ref={ref} {...props} />
  )
)

Box.displayName = 'Box'

// Spacer Component
export interface SpacerProps {
  /** Size of the spacer */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl'
  /** Direction of the spacer */
  direction?: 'horizontal' | 'vertical'
  /** Custom height/width */
  height?: string
  width?: string
}

export const Spacer: React.FC<SpacerProps> = ({
  size = 'md',
  direction = 'vertical',
  height,
  width
}) => {
  const sizeMap = {
    xs: direction === 'vertical' ? 'h-1' : 'w-1',
    sm: direction === 'vertical' ? 'h-2' : 'w-2',
    md: direction === 'vertical' ? 'h-4' : 'w-4',
    lg: direction === 'vertical' ? 'h-6' : 'w-6',
    xl: direction === 'vertical' ? 'h-8' : 'w-8',
    '2xl': direction === 'vertical' ? 'h-12' : 'w-12'
  }

  const style: React.CSSProperties = {}
  if (height) style.height = height
  if (width) style.width = width

  return (
    <div
      className={cn(
        'flex-shrink-0',
        !height && !width && sizeMap[size]
      )}
      style={style}
      aria-hidden="true"
    />
  )
}

Spacer.displayName = 'Spacer'

// Divider Component
export interface DividerProps extends HTMLAttributes<HTMLHRElement> {
  /** Orientation of the divider */
  orientation?: 'horizontal' | 'vertical'
  /** Thickness of the divider */
  thickness?: 'thin' | 'medium' | 'thick'
  /** Style of the divider */
  variant?: 'solid' | 'dashed' | 'dotted'
  /** Label for the divider */
  label?: string
}

export const Divider = forwardRef<HTMLHRElement, DividerProps>(
  ({ 
    className, 
    orientation = 'horizontal',
    thickness = 'thin',
    variant = 'solid',
    label,
    ...props 
  }, ref) => {
    const thicknessClasses = {
      thin: orientation === 'horizontal' ? 'border-t' : 'border-l',
      medium: orientation === 'horizontal' ? 'border-t-2' : 'border-l-2',
      thick: orientation === 'horizontal' ? 'border-t-4' : 'border-l-4'
    }

    const variantClasses = {
      solid: '',
      dashed: 'border-dashed',
      dotted: 'border-dotted'
    }

    const orientationClasses = {
      horizontal: 'w-full',
      vertical: 'h-full min-h-[1rem]'
    }

    if (label) {
      return (
        <div className={cn(
          'relative flex items-center',
          orientation === 'horizontal' ? 'w-full' : 'flex-col h-full',
          className
        )}>
          <div className={cn(
            'flex-1 border-border',
            thicknessClasses[thickness],
            variantClasses[variant]
          )} />
          <span className={cn(
            'text-sm text-muted-foreground bg-background',
            orientation === 'horizontal' ? 'px-3' : 'py-3'
          )}>
            {label}
          </span>
          <div className={cn(
            'flex-1 border-border',
            thicknessClasses[thickness],
            variantClasses[variant]
          )} />
        </div>
      )
    }

    return (
      <hr
        ref={ref}
        className={cn(
          'border-0 border-border',
          thicknessClasses[thickness],
          variantClasses[variant],
          orientationClasses[orientation],
          className
        )}
        {...props}
      />
    )
  }
)

Divider.displayName = 'Divider'

// Center Component
export interface CenterProps extends HTMLAttributes<HTMLDivElement> {
  /** Whether to center inline (horizontally) */
  inline?: boolean
}

export const Center = forwardRef<HTMLDivElement, CenterProps>(
  ({ className, inline = false, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        'flex items-center justify-center',
        inline ? 'inline-flex' : 'flex',
        className
      )}
      {...props}
    />
  )
)

Center.displayName = 'Center'

// Aspect Ratio Component
export interface AspectRatioProps extends HTMLAttributes<HTMLDivElement> {
  /** Aspect ratio */
  ratio?: 'square' | 'video' | 'photo' | 'golden' | number
}

export const AspectRatio = forwardRef<HTMLDivElement, AspectRatioProps>(
  ({ className, ratio = 'square', children, ...props }, ref) => {
    const ratioClasses = {
      square: 'aspect-square',      // 1:1
      video: 'aspect-video',        // 16:9
      photo: 'aspect-[4/3]',        // 4:3
      golden: 'aspect-[1.618/1]'    // Golden ratio
    }

    const ratioClass = typeof ratio === 'number' 
      ? `aspect-[${ratio}/1]` 
      : ratioClasses[ratio] || ratioClasses.square

    return (
      <div
        ref={ref}
        className={cn('relative w-full', ratioClass, className)}
        {...props}
      >
        <div className="absolute inset-0">
          {children}
        </div>
      </div>
    )
  }
)

AspectRatio.displayName = 'AspectRatio'

// Masonry Component (CSS Grid Masonry when supported)
export interface MasonryProps extends HTMLAttributes<HTMLDivElement> {
  /** Number of columns */
  columns?: number
  /** Gap between items */
  gap?: 'sm' | 'md' | 'lg' | 'xl'
}

export const Masonry = forwardRef<HTMLDivElement, MasonryProps>(
  ({ className, columns = 3, gap = 'md', ...props }, ref) => {
    const gapClasses = {
      sm: 'gap-2',
      md: 'gap-4',
      lg: 'gap-6',
      xl: 'gap-8'
    }

    return (
      <div
        ref={ref}
        className={cn(
          'columns-1',
          columns >= 2 && 'sm:columns-2',
          columns >= 3 && 'md:columns-3',
          columns >= 4 && 'lg:columns-4',
          columns >= 5 && 'xl:columns-5',
          gapClasses[gap],
          className
        )}
        style={{
          columnCount: columns,
          columnFill: 'balance'
        }}
        {...props}
      />
    )
  }
)

Masonry.displayName = 'Masonry'

export default {
  Container,
  Grid,
  Flex,
  Stack,
  Box,
  Spacer,
  Divider,
  Center,
  AspectRatio,
  Masonry
}