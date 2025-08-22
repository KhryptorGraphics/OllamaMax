import React, { forwardRef, useState, useMemo } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { User } from 'lucide-react'
import { cn } from '@/utils/cn'

// Avatar variants using design tokens
const avatarVariants = cva(
  [
    'relative flex shrink-0 overflow-hidden transition-smooth',
    'bg-muted border border-border/20'
  ],
  {
    variants: {
      size: {
        xs: 'w-6 h-6 text-xs',
        sm: 'w-8 h-8 text-sm',
        md: 'w-10 h-10 text-base',
        lg: 'w-12 h-12 text-lg',
        xl: 'w-16 h-16 text-xl',
        '2xl': 'w-20 h-20 text-2xl'
      },
      shape: {
        circle: 'rounded-full',
        rounded: 'rounded-lg',
        square: 'rounded-none'
      },
      variant: {
        default: 'shadow-elevation-1 hover:shadow-elevation-2',
        outline: 'border-2 border-border shadow-none',
        ghost: 'border-0 shadow-none bg-transparent'
      }
    },
    defaultVariants: {
      size: 'md',
      shape: 'circle',
      variant: 'default'
    }
  }
)

// Avatar image variants
const avatarImageVariants = cva(
  'aspect-square h-full w-full object-cover'
)

// Avatar fallback variants
const avatarFallbackVariants = cva(
  [
    'flex h-full w-full items-center justify-center',
    'bg-gradient-to-br from-primary-400 to-primary-600',
    'text-white font-medium select-none'
  ]
)

// Generate initials from name
const generateInitials = (name: string): string => {
  if (!name) return ''
  
  const words = name.trim().split(/\s+/)
  if (words.length === 1) {
    return words[0].charAt(0).toUpperCase()
  }
  
  return (words[0].charAt(0) + words[words.length - 1].charAt(0)).toUpperCase()
}

// Generate deterministic color from string
const generateColorFromString = (str: string): string => {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash)
  }
  
  // Convert to HSL for better color distribution
  const hue = Math.abs(hash) % 360
  return `hsl(${hue}, 65%, 50%)`
}

export interface AvatarProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof avatarVariants> {
  /** Image source URL */
  src?: string
  /** Alt text for the image */
  alt?: string
  /** Name for generating fallback initials and alt text */
  name?: string
  /** Custom fallback content */
  fallback?: React.ReactNode
  /** Loading state */
  loading?: boolean
  /** Whether to use colored background for initials */
  colorful?: boolean
  /** Custom fallback background color */
  fallbackBg?: string
  /** Image loading strategy */
  imageLoading?: 'eager' | 'lazy'
  /** Callback when image fails to load */
  onImageError?: () => void
  /** Callback when image loads successfully */
  onImageLoad?: () => void
}

/**
 * Avatar component with comprehensive fallback system and accessibility
 * 
 * Features:
 * - Multiple sizes (xs, sm, md, lg, xl, 2xl)
 * - Shape variants (circle, rounded, square)
 * - Style variants (default, outline, ghost)
 * - Automatic fallback system (image → initials → icon)
 * - Colorful initials based on name
 * - Loading states
 * - Full accessibility support
 * - Design token integration
 */
export const Avatar = forwardRef<HTMLDivElement, AvatarProps>(
  (
    {
      className,
      size,
      shape,
      variant,
      src,
      alt,
      name,
      fallback,
      loading = false,
      colorful = true,
      fallbackBg,
      imageLoading = 'lazy',
      onImageError,
      onImageLoad,
      ...props
    },
    ref
  ) => {
    const [imageError, setImageError] = useState(false)
    const [imageLoaded, setImageLoaded] = useState(false)

    // Generate fallback content
    const initials = useMemo(() => name ? generateInitials(name) : '', [name])
    const fallbackColor = useMemo(() => {
      if (fallbackBg) return fallbackBg
      if (colorful && name) return generateColorFromString(name)
      return undefined
    }, [fallbackBg, colorful, name])

    // Generate alt text
    const imageAlt = alt || (name ? `${name}'s avatar` : 'Avatar')

    // Handle image error
    const handleImageError = () => {
      setImageError(true)
      onImageError?.()
    }

    // Handle image load
    const handleImageLoad = () => {
      setImageLoaded(true)
      setImageError(false)
      onImageLoad?.()
    }

    // Determine what to show
    const shouldShowImage = src && !imageError && !loading
    const shouldShowFallback = !shouldShowImage

    return (
      <div
        ref={ref}
        className={cn(avatarVariants({ size, shape, variant }), className)}
        {...props}
      >
        {/* Loading state */}
        {loading && (
          <div className={cn(avatarFallbackVariants())}>
            <div className="animate-pulse bg-muted-foreground/20 rounded-full w-full h-full" />
          </div>
        )}

        {/* Image */}
        {shouldShowImage && (
          <img
            src={src}
            alt={imageAlt}
            loading={imageLoading}
            onError={handleImageError}
            onLoad={handleImageLoad}
            className={cn(
              avatarImageVariants(),
              shape === 'circle' && 'rounded-full',
              shape === 'rounded' && 'rounded-lg'
            )}
          />
        )}

        {/* Fallback */}
        {shouldShowFallback && !loading && (
          <div
            className={cn(
              avatarFallbackVariants(),
              shape === 'circle' && 'rounded-full',
              shape === 'rounded' && 'rounded-lg'
            )}
            style={fallbackColor ? { backgroundColor: fallbackColor } : undefined}
            aria-label={imageAlt}
          >
            {fallback ? (
              fallback
            ) : initials ? (
              <span
                className="font-semibold"
                aria-hidden="true"
              >
                {initials}
              </span>
            ) : (
              <User 
                className={cn(
                  'text-muted-foreground',
                  size === 'xs' && 'w-3 h-3',
                  size === 'sm' && 'w-4 h-4',
                  size === 'md' && 'w-5 h-5',
                  size === 'lg' && 'w-6 h-6',
                  size === 'xl' && 'w-8 h-8',
                  size === '2xl' && 'w-10 h-10'
                )}
                aria-hidden="true"
              />
            )}
          </div>
        )}
      </div>
    )
  }
)

Avatar.displayName = 'Avatar'

// Avatar group component for displaying multiple avatars
export interface AvatarGroupProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Maximum number of avatars to show before showing count */
  max?: number
  /** Size of avatars in the group */
  size?: AvatarProps['size']
  /** Shape of avatars in the group */
  shape?: AvatarProps['shape']
  /** Variant of avatars in the group */
  variant?: AvatarProps['variant']
  /** Spacing between avatars */
  spacing?: 'tight' | 'normal' | 'loose'
  /** Show count of remaining avatars */
  showRemaining?: boolean
  /** Custom content for remaining count */
  remainingText?: (count: number) => string
}

export const AvatarGroup = forwardRef<HTMLDivElement, AvatarGroupProps>(
  (
    {
      className,
      children,
      max = 3,
      size = 'md',
      shape = 'circle',
      variant = 'default',
      spacing = 'normal',
      showRemaining = true,
      remainingText = (count) => `+${count}`,
      ...props
    },
    ref
  ) => {
    const childrenArray = React.Children.toArray(children)
    const visibleChildren = childrenArray.slice(0, max)
    const remainingCount = Math.max(0, childrenArray.length - max)

    const spacingClasses = {
      tight: '-space-x-1',
      normal: '-space-x-2',
      loose: '-space-x-3'
    }

    return (
      <div
        ref={ref}
        className={cn(
          'flex items-center',
          spacingClasses[spacing],
          className
        )}
        {...props}
      >
        {/* Visible avatars */}
        {visibleChildren.map((child, index) => (
          <div
            key={index}
            className="relative ring-2 ring-background"
            style={{ zIndex: visibleChildren.length - index }}
          >
            {React.isValidElement(child) 
              ? React.cloneElement(child as React.ReactElement<AvatarProps>, {
                  size,
                  shape,
                  variant
                })
              : child
            }
          </div>
        ))}

        {/* Remaining count */}
        {showRemaining && remainingCount > 0 && (
          <div
            className="relative ring-2 ring-background"
            style={{ zIndex: 0 }}
          >
            <Avatar
              size={size}
              shape={shape}
              variant={variant}
              fallback={
                <span className="text-xs font-semibold">
                  {remainingText(remainingCount)}
                </span>
              }
              colorful={false}
              fallbackBg="hsl(var(--muted))"
            />
          </div>
        )}
      </div>
    )
  }
)

AvatarGroup.displayName = 'AvatarGroup'

// Avatar with status indicator
export interface AvatarWithStatusProps extends AvatarProps {
  /** Status indicator */
  status?: 'online' | 'offline' | 'away' | 'busy'
  /** Custom status color */
  statusColor?: string
  /** Status position */
  statusPosition?: 'top-right' | 'bottom-right' | 'top-left' | 'bottom-left'
}

export const AvatarWithStatus = forwardRef<HTMLDivElement, AvatarWithStatusProps>(
  (
    {
      status,
      statusColor,
      statusPosition = 'bottom-right',
      className,
      ...props
    },
    ref
  ) => {
    const statusColors = {
      online: 'bg-green-500',
      offline: 'bg-gray-400',
      away: 'bg-yellow-500',
      busy: 'bg-red-500'
    }

    const positionClasses = {
      'top-right': 'top-0 right-0',
      'bottom-right': 'bottom-0 right-0',
      'top-left': 'top-0 left-0',
      'bottom-left': 'bottom-0 left-0'
    }

    return (
      <div className={cn('relative inline-block', className)}>
        <Avatar ref={ref} {...props} />
        {status && (
          <span
            className={cn(
              'absolute block rounded-full ring-2 ring-background',
              'w-3 h-3',
              positionClasses[statusPosition],
              statusColor ? '' : statusColors[status]
            )}
            style={statusColor ? { backgroundColor: statusColor } : undefined}
            aria-label={`Status: ${status}`}
          />
        )}
      </div>
    )
  }
)

AvatarWithStatus.displayName = 'AvatarWithStatus'

export default Avatar
export { avatarVariants }