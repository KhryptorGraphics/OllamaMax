import React, { forwardRef, HTMLAttributes } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/utils/cn'

// Card variants using design tokens
const cardVariants = cva(
  [
    'rounded-card border bg-card text-card-foreground',
    'transition-smooth'
  ],
  {
    variants: {
      variant: {
        default: 'border-border shadow-elevation-1',
        elevated: 'border-border shadow-elevation-2',
        outlined: 'border-2 border-border shadow-none',
        filled: 'border-transparent bg-muted shadow-none',
        floating: 'border-border shadow-elevation-3',
        interactive: [
          'border-border shadow-elevation-1 cursor-pointer',
          'hover:shadow-elevation-2 hover:border-primary/50',
          'focus-visible:outline-none focus-visible:ring-focus focus-visible:ring-offset-2',
          'active:shadow-elevation-1 active:scale-[0.99]'
        ]
      },
      size: {
        sm: 'rounded-sm',
        md: 'rounded-card',
        lg: 'rounded-lg',
        xl: 'rounded-xl'
      },
      padding: {
        none: 'p-0',
        xs: 'p-2',
        sm: 'p-3',
        md: 'p-4',
        lg: 'p-6',
        xl: 'p-8',
        '2xl': 'p-10'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'md',
      padding: 'md'
    }
  }
)

export interface CardProps
  extends HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof cardVariants> {
  /** Whether the card is interactive (clickable) */
  interactive?: boolean
  /** Click handler for interactive cards */
  onCardClick?: () => void
  /** Loading state */
  loading?: boolean
  /** Disabled state */
  disabled?: boolean
  /** Custom ARIA label for accessibility */
  ariaLabel?: string
}

/**
 * Card component - A flexible container for grouping related content
 * 
 * Features:
 * - Multiple variants (default, elevated, outlined, filled, floating, interactive)
 * - Size variants (sm, md, lg, xl) with design token integration
 * - Padding variants (none, xs, sm, md, lg, xl, 2xl)
 * - Loading and disabled states
 * - Interactive states with hover, focus, and active feedback
 * - Full accessibility support with ARIA labels
 * - Composable with Card.Header, Card.Content, Card.Footer
 * - Design token integration for consistent theming
 */
export const Card = forwardRef<HTMLDivElement, CardProps>(
  (
    {
      className,
      variant,
      size,
      padding,
      interactive,
      onCardClick,
      loading = false,
      disabled = false,
      ariaLabel,
      children,
      ...props
    },
    ref
  ) => {
    const finalVariant = interactive ? 'interactive' : variant
    const Component = interactive || onCardClick ? 'button' : 'div'

    // Handle loading state
    if (loading) {
      return (
        <div
          ref={ref}
          className={cn(
            cardVariants({ variant: 'default', size, padding }),
            'animate-pulse',
            className
          )}
          aria-label="Loading content"
          {...props}
        >
          <div className="space-y-3">
            <div className="h-4 bg-muted rounded-sm" />
            <div className="h-4 bg-muted rounded-sm w-3/4" />
            <div className="h-20 bg-muted rounded-sm" />
          </div>
        </div>
      )
    }

    return (
      <Component
        className={cn(
          cardVariants({ variant: finalVariant, size, padding }),
          disabled && 'opacity-50 cursor-not-allowed',
          className
        )}
        ref={ref as any}
        onClick={disabled ? undefined : onCardClick}
        disabled={disabled}
        aria-label={ariaLabel}
        {...(interactive && !disabled && {
          role: 'button',
          tabIndex: 0,
          onKeyDown: (e: React.KeyboardEvent) => {
            if ((e.key === 'Enter' || e.key === ' ') && onCardClick) {
              e.preventDefault()
              onCardClick()
            }
          }
        })}
        {...props}
      >
        {children}
      </Component>
    )
  }
)

Card.displayName = 'Card'

// Card Header with design tokens
export interface CardHeaderProps extends HTMLAttributes<HTMLDivElement> {
  /** Whether to include bottom border */
  divided?: boolean
  /** Spacing variant */
  spacing?: 'none' | 'sm' | 'md' | 'lg'
  /** Alignment of header content */
  align?: 'start' | 'center' | 'end'
}

export const CardHeader = forwardRef<HTMLDivElement, CardHeaderProps>(
  ({ 
    className, 
    divided = false, 
    spacing = 'md',
    align = 'start',
    ...props 
  }, ref) => {
    const spacingClasses = {
      none: 'space-y-0',
      sm: 'space-y-1',
      md: 'space-y-1.5',
      lg: 'space-y-2'
    }

    const alignClasses = {
      start: 'items-start text-left',
      center: 'items-center text-center',
      end: 'items-end text-right'
    }

    return (
      <div
        ref={ref}
        className={cn(
          'flex flex-col',
          spacingClasses[spacing],
          alignClasses[align],
          divided && 'border-b border-border pb-spacing-4 mb-spacing-4',
          className
        )}
        {...props}
      />
    )
  }
)

CardHeader.displayName = 'CardHeader'

// Card Content with design tokens
export interface CardContentProps extends HTMLAttributes<HTMLDivElement> {
  /** Spacing variant for content */
  spacing?: 'none' | 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  /** Padding override for content area */
  padding?: 'none' | 'xs' | 'sm' | 'md' | 'lg' | 'xl'
}

export const CardContent = forwardRef<HTMLDivElement, CardContentProps>(
  ({ 
    className, 
    spacing = 'none',
    padding,
    ...props 
  }, ref) => {
    const spacingClasses = {
      none: '',
      xs: 'space-y-1',
      sm: 'space-y-2',
      md: 'space-y-4',
      lg: 'space-y-6',
      xl: 'space-y-8'
    }

    const paddingClasses = {
      none: 'p-0',
      xs: 'p-2',
      sm: 'p-3', 
      md: 'p-4',
      lg: 'p-6',
      xl: 'p-8'
    }

    return (
      <div
        ref={ref}
        className={cn(
          spacingClasses[spacing],
          padding && paddingClasses[padding],
          className
        )}
        {...props}
      />
    )
  }
)

CardContent.displayName = 'CardContent'

// Card Footer with design tokens
export interface CardFooterProps extends HTMLAttributes<HTMLDivElement> {
  /** Whether to include top border */
  divided?: boolean
  /** Justify content alignment */
  justify?: 'start' | 'center' | 'end' | 'between' | 'around' | 'evenly'
  /** Flex direction */
  direction?: 'row' | 'column'
  /** Gap between items */
  gap?: 'none' | 'xs' | 'sm' | 'md' | 'lg'
}

export const CardFooter = forwardRef<HTMLDivElement, CardFooterProps>(
  ({ 
    className, 
    divided = false, 
    justify = 'end',
    direction = 'row',
    gap = 'md',
    ...props 
  }, ref) => {
    const justifyClasses = {
      start: 'justify-start',
      center: 'justify-center',
      end: 'justify-end',
      between: 'justify-between',
      around: 'justify-around',
      evenly: 'justify-evenly'
    }

    const directionClasses = {
      row: 'flex-row',
      column: 'flex-col'
    }

    const gapClasses = {
      none: 'gap-0',
      xs: 'gap-1',
      sm: 'gap-2',
      md: 'gap-3',
      lg: 'gap-4'
    }

    return (
      <div
        ref={ref}
        className={cn(
          'flex items-center',
          directionClasses[direction],
          justifyClasses[justify],
          gapClasses[gap],
          divided && 'border-t border-border pt-spacing-4 mt-spacing-4',
          className
        )}
        {...props}
      />
    )
  }
)

CardFooter.displayName = 'CardFooter'

// Card Title with design tokens
export interface CardTitleProps extends HTMLAttributes<HTMLHeadingElement> {
  /** Heading level */
  level?: 1 | 2 | 3 | 4 | 5 | 6
  /** Size variant using design tokens */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl'
  /** Weight variant */
  weight?: 'normal' | 'medium' | 'semibold' | 'bold'
  /** Whether title should be truncated */
  truncate?: boolean
}

export const CardTitle = forwardRef<HTMLHeadingElement, CardTitleProps>(
  ({ 
    className, 
    level = 3, 
    size = 'md', 
    weight = 'semibold',
    truncate = false,
    ...props 
  }, ref) => {
    const Component = `h${level}` as keyof JSX.IntrinsicElements

    const sizeClasses = {
      xs: 'text-xs',
      sm: 'text-sm',
      md: 'text-base',
      lg: 'text-lg',
      xl: 'text-xl',
      '2xl': 'text-2xl'
    }

    const weightClasses = {
      normal: 'font-normal',
      medium: 'font-medium',
      semibold: 'font-semibold',
      bold: 'font-bold'
    }

    return (
      <Component
        ref={ref as any}
        className={cn(
          'leading-tight tracking-tight text-card-foreground',
          sizeClasses[size],
          weightClasses[weight],
          truncate && 'truncate',
          className
        )}
        {...props}
      />
    )
  }
)

CardTitle.displayName = 'CardTitle'

// Card Description with design tokens
export interface CardDescriptionProps extends HTMLAttributes<HTMLParagraphElement> {
  /** Size variant */
  size?: 'xs' | 'sm' | 'md' | 'lg'
  /** Line clamp for text truncation */
  lineClamp?: 1 | 2 | 3 | 4 | 5 | 6
  /** Variant for different text styles */
  variant?: 'default' | 'muted' | 'accent'
}

export const CardDescription = forwardRef<HTMLParagraphElement, CardDescriptionProps>(
  ({ 
    className, 
    size = 'sm', 
    lineClamp,
    variant = 'muted',
    ...props 
  }, ref) => {
    const sizeClasses = {
      xs: 'text-xs',
      sm: 'text-sm',
      md: 'text-base',
      lg: 'text-lg'
    }

    const variantClasses = {
      default: 'text-card-foreground',
      muted: 'text-muted-foreground',
      accent: 'text-accent-foreground'
    }

    const lineClampClasses = {
      1: 'line-clamp-1',
      2: 'line-clamp-2', 
      3: 'line-clamp-3',
      4: 'line-clamp-4',
      5: 'line-clamp-5',
      6: 'line-clamp-6'
    }

    return (
      <p
        ref={ref}
        className={cn(
          'leading-relaxed',
          sizeClasses[size],
          variantClasses[variant],
          lineClamp && lineClampClasses[lineClamp],
          className
        )}
        {...props}
      />
    )
  }
)

CardDescription.displayName = 'CardDescription'

// Card Image with design tokens
export interface CardImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  /** Image position */
  position?: 'top' | 'bottom' | 'left' | 'right' | 'full'
  /** Aspect ratio */
  aspectRatio?: 'square' | 'video' | 'portrait' | 'landscape' | 'auto'
  /** Whether image should cover the container */
  cover?: boolean
  /** Loading state */
  loading?: boolean
  /** Error fallback */
  fallback?: React.ReactNode
}

export const CardImage = forwardRef<HTMLImageElement, CardImageProps>(
  ({ 
    className, 
    position = 'top', 
    aspectRatio = 'auto', 
    cover = true,
    loading = false,
    fallback,
    alt,
    onError,
    ...props 
  }, ref) => {
    const [hasError, setHasError] = React.useState(false)

    const aspectClasses = {
      square: 'aspect-square',
      video: 'aspect-video',
      portrait: 'aspect-[3/4]',
      landscape: 'aspect-[4/3]',
      auto: ''
    }

    const positionClasses = {
      top: 'rounded-t-card',
      bottom: 'rounded-b-card',
      left: 'rounded-l-card',
      right: 'rounded-r-card',
      full: 'rounded-card'
    }

    const handleError = (e: React.SyntheticEvent<HTMLImageElement>) => {
      setHasError(true)
      onError?.(e)
    }

    if (loading) {
      return (
        <div
          className={cn(
            'w-full bg-muted animate-pulse flex items-center justify-center',
            aspectClasses[aspectRatio],
            positionClasses[position],
            className
          )}
        >
          <div className="text-muted-foreground text-xs">Loading...</div>
        </div>
      )
    }

    if (hasError && fallback) {
      return (
        <div
          className={cn(
            'w-full bg-muted flex items-center justify-center',
            aspectClasses[aspectRatio],
            positionClasses[position],
            className
          )}
        >
          {fallback}
        </div>
      )
    }

    return (
      <img
        ref={ref}
        className={cn(
          'w-full transition-smooth',
          cover && 'object-cover',
          aspectClasses[aspectRatio],
          positionClasses[position],
          className
        )}
        alt={alt}
        onError={handleError}
        {...props}
      />
    )
  }
)

CardImage.displayName = 'CardImage'

// Card Actions with design tokens
export interface CardActionsProps extends HTMLAttributes<HTMLDivElement> {
  /** Spacing between actions */
  spacing?: 'none' | 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  /** Actions alignment */
  align?: 'start' | 'center' | 'end' | 'between' | 'around'
  /** Direction of actions */
  direction?: 'row' | 'column'
  /** Whether actions should be full width */
  fullWidth?: boolean
}

export const CardActions = forwardRef<HTMLDivElement, CardActionsProps>(
  ({ 
    className, 
    spacing = 'md', 
    align = 'end',
    direction = 'row',
    fullWidth = false,
    ...props 
  }, ref) => {
    const spacingClasses = {
      none: direction === 'row' ? 'space-x-0' : 'space-y-0',
      xs: direction === 'row' ? 'space-x-1' : 'space-y-1',
      sm: direction === 'row' ? 'space-x-2' : 'space-y-2',
      md: direction === 'row' ? 'space-x-3' : 'space-y-3',
      lg: direction === 'row' ? 'space-x-4' : 'space-y-4',
      xl: direction === 'row' ? 'space-x-6' : 'space-y-6'
    }

    const alignClasses = {
      start: 'justify-start',
      center: 'justify-center',
      end: 'justify-end',
      between: 'justify-between',
      around: 'justify-around'
    }

    const directionClasses = {
      row: 'flex-row items-center',
      column: 'flex-col items-stretch'
    }

    return (
      <div
        ref={ref}
        className={cn(
          'flex',
          directionClasses[direction],
          alignClasses[align],
          spacingClasses[spacing],
          fullWidth && 'w-full',
          className
        )}
        {...props}
      />
    )
  }
)

CardActions.displayName = 'CardActions'

// Compound component exports
Card.Header = CardHeader
Card.Content = CardContent
Card.Footer = CardFooter
Card.Title = CardTitle
Card.Description = CardDescription
Card.Image = CardImage
Card.Actions = CardActions

export default Card