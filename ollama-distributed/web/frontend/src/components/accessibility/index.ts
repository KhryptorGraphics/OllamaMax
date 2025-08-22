/**
 * @fileoverview Accessibility Components Barrel Export
 * Centralized exports for all accessibility-related components and utilities
 */

// Main provider and context
export { 
  AccessibilityProvider, 
  useAccessibilityContext, 
  useComponentAccessibility,
  withAccessibility 
} from './AccessibilityProvider'

// UI Components
export { AccessibilityPanel } from './AccessibilityPanel'
export { FocusTrap } from './FocusTrap'
export { SkipLinks } from './SkipLinks'
export { 
  AccessibilityAnnouncer,
  StatusAnnouncer,
  AlertAnnouncer 
} from './AccessibilityAnnouncer'

// Hooks
export { 
  useAriaLiveRegion,
  useStatusAnnouncer,
  useAlertAnnouncer,
  useLoadingAnnouncer 
} from '@/hooks/useAriaLiveRegion'

// Testing utilities
export {
  AccessibilityTester,
  accessibilityTester,
  testAccessibility,
  testAxeCompliance,
  testKeyboardNavigation,
  testScreenReaderCompatibility
} from '@/utils/accessibility-testing'

// Core accessibility utilities
export {
  useAccessibility,
  ScreenReaderAnnouncer,
  FocusManager,
  KeyboardNavigationManager,
  accessibilityUtils,
  AccessibilityPreferences
} from '@/utils/accessibility'

// Type exports
export type {
  AccessibilityTestOptions,
  KeyboardTestOptions,
  ScreenReaderTestOptions
} from '@/utils/accessibility-testing'

export type {
  AccessibilityOptions,
  FocusableElement,
  AnnouncementOptions
} from '@/utils/accessibility'