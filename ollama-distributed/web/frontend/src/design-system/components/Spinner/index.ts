/**
 * Spinner Component Exports
 * 
 * Comprehensive loading indicators with multiple animation types,
 * sizes, and display modes. Fully accessible and integrated with
 * the design system tokens.
 */

export {
  Spinner,
  SpinnerOverlay,
  LoadingButton,
  Skeleton,
  spinnerVariants,
  spinnerContainerVariants,
  type SpinnerProps,
  type SpinnerOverlayProps,
  type LoadingButtonProps,
  type SkeletonProps
} from './Spinner'

// Re-export variant props type for external use
export type { VariantProps } from 'class-variance-authority'