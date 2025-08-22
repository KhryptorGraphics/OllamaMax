/**
 * @fileoverview Accessibility Provider Component
 * Provides global accessibility context and features
 */

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { useAccessibility } from '@/utils/accessibility'

interface AccessibilitySettings {
  /** Reduced motion preference */
  reducedMotion: boolean
  /** High contrast mode */
  highContrast: boolean
  /** Large text mode */
  largeText: boolean
  /** Dark mode preference */
  darkMode: boolean
  /** Screen reader optimizations */
  screenReaderOptimized: boolean
  /** Keyboard navigation enhanced */
  keyboardEnhanced: boolean
  /** Focus trap enabled */
  focusTrapping: boolean
  /** Announcement preferences */
  announcements: {
    navigation: boolean
    loading: boolean
    errors: boolean
    success: boolean
  }
}

interface AccessibilityContextType {
  settings: AccessibilitySettings
  updateSettings: (updates: Partial<AccessibilitySettings>) => void
  announce: (message: string, priority?: 'polite' | 'assertive') => void
  announcePageChange: (title: string) => void
  announceLoadingState: (isLoading: boolean, context?: string) => void
  announceError: (error: string) => void
  announceSuccess: (message: string) => void
  saveFocus: () => void
  restoreFocus: () => boolean
  trapFocus: (container: HTMLElement) => () => void
  focusFirst: (container?: HTMLElement) => boolean
  focusLast: (container?: HTMLElement) => boolean
}

const AccessibilityContext = createContext<AccessibilityContextType | undefined>(undefined)

export interface AccessibilityProviderProps {
  children: ReactNode
  /** Initial accessibility settings */
  initialSettings?: Partial<AccessibilitySettings>
  /** Storage key for persisting settings */
  storageKey?: string
}

const DEFAULT_SETTINGS: AccessibilitySettings = {
  reducedMotion: false,
  highContrast: false,
  largeText: false,
  darkMode: false,
  screenReaderOptimized: false,
  keyboardEnhanced: false,
  focusTrapping: true,
  announcements: {
    navigation: true,
    loading: true,
    errors: true,
    success: true
  }
}

/**
 * Accessibility Provider Component
 * 
 * Provides global accessibility context, settings management, and utilities
 * for the entire application. Automatically detects user preferences and
 * provides methods for accessible interactions.
 */
export const AccessibilityProvider: React.FC<AccessibilityProviderProps> = ({
  children,
  initialSettings = {},
  storageKey = 'accessibility-settings'
}) => {
  const [settings, setSettings] = useState<AccessibilitySettings>(() => {
    // Load saved settings from localStorage
    if (typeof window !== 'undefined') {
      try {
        const saved = localStorage.getItem(storageKey)
        if (saved) {
          return { ...DEFAULT_SETTINGS, ...JSON.parse(saved), ...initialSettings }
        }
      } catch (error) {
        console.warn('Failed to load accessibility settings:', error)
      }
    }
    
    return { ...DEFAULT_SETTINGS, ...initialSettings }
  })

  // Initialize accessibility utilities
  const accessibilityUtils = useAccessibility({
    announcePageChanges: settings.announcements.navigation,
    focusManagement: true,
    keyboardNavigation: settings.keyboardEnhanced,
    reducedMotion: settings.reducedMotion
  })

  // Detect user preferences from media queries
  useEffect(() => {
    if (typeof window === 'undefined') return

    const mediaQueries = {
      reducedMotion: window.matchMedia('(prefers-reduced-motion: reduce)'),
      highContrast: window.matchMedia('(prefers-contrast: high)'),
      darkMode: window.matchMedia('(prefers-color-scheme: dark)')
    }

    const handleMediaChange = () => {
      const detectedSettings = {
        reducedMotion: mediaQueries.reducedMotion.matches,
        highContrast: mediaQueries.highContrast.matches,
        darkMode: mediaQueries.darkMode.matches
      }

      // Only update if user hasn't manually set these preferences
      setSettings(prevSettings => {
        const shouldUpdate = (
          !localStorage.getItem(`${storageKey}-user-reduced-motion`) ||
          !localStorage.getItem(`${storageKey}-user-high-contrast`) ||
          !localStorage.getItem(`${storageKey}-user-dark-mode`)
        )

        if (shouldUpdate) {
          return { ...prevSettings, ...detectedSettings }
        }
        return prevSettings
      })
    }

    // Initial detection
    handleMediaChange()

    // Listen for changes
    Object.values(mediaQueries).forEach(mq => {
      mq.addEventListener('change', handleMediaChange)
    })

    return () => {
      Object.values(mediaQueries).forEach(mq => {
        mq.removeEventListener('change', handleMediaChange)
      })
    }
  }, [storageKey])

  // Apply settings to document
  useEffect(() => {
    if (typeof document === 'undefined') return

    const root = document.documentElement

    // Apply reduced motion
    if (settings.reducedMotion) {
      root.style.setProperty('--animation-duration', '0.01ms')
      root.style.setProperty('--transition-duration', '0.01ms')
      root.classList.add('reduce-motion')
    } else {
      root.style.removeProperty('--animation-duration')
      root.style.removeProperty('--transition-duration')
      root.classList.remove('reduce-motion')
    }

    // Apply high contrast
    if (settings.highContrast) {
      root.classList.add('high-contrast')
    } else {
      root.classList.remove('high-contrast')
    }

    // Apply large text
    if (settings.largeText) {
      root.style.setProperty('--base-font-size', '18px')
      root.classList.add('large-text')
    } else {
      root.style.removeProperty('--base-font-size')
      root.classList.remove('large-text')
    }

    // Apply dark mode
    if (settings.darkMode) {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }

    // Apply screen reader optimizations
    if (settings.screenReaderOptimized) {
      root.classList.add('screen-reader-optimized')
    } else {
      root.classList.remove('screen-reader-optimized')
    }

    // Apply keyboard enhanced mode
    if (settings.keyboardEnhanced) {
      root.classList.add('keyboard-enhanced')
    } else {
      root.classList.remove('keyboard-enhanced')
    }
  }, [settings])

  // Save settings to localStorage
  useEffect(() => {
    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem(storageKey, JSON.stringify(settings))
      } catch (error) {
        console.warn('Failed to save accessibility settings:', error)
      }
    }
  }, [settings, storageKey])

  const updateSettings = (updates: Partial<AccessibilitySettings>) => {
    setSettings(prevSettings => {
      const newSettings = { ...prevSettings, ...updates }
      
      // Mark user-set preferences
      if (typeof window !== 'undefined') {
        Object.keys(updates).forEach(key => {
          localStorage.setItem(`${storageKey}-user-${key}`, 'true')
        })
      }
      
      return newSettings
    })
  }

  const announce = (message: string, priority: 'polite' | 'assertive' = 'polite') => {
    accessibilityUtils.announce(message, { priority })
  }

  const announcePageChange = (title: string) => {
    if (settings.announcements.navigation) {
      accessibilityUtils.announcePageChange(title)
    }
  }

  const announceLoadingState = (isLoading: boolean, context?: string) => {
    if (settings.announcements.loading) {
      accessibilityUtils.announceLoadingState(isLoading, context)
    }
  }

  const announceError = (error: string) => {
    if (settings.announcements.errors) {
      accessibilityUtils.announceError(error)
    }
  }

  const announceSuccess = (message: string) => {
    if (settings.announcements.success) {
      accessibilityUtils.announceSuccess(message)
    }
  }

  const contextValue: AccessibilityContextType = {
    settings,
    updateSettings,
    announce,
    announcePageChange,
    announceLoadingState,
    announceError,
    announceSuccess,
    saveFocus: accessibilityUtils.saveFocus,
    restoreFocus: accessibilityUtils.restoreFocus,
    trapFocus: accessibilityUtils.trapFocus,
    focusFirst: accessibilityUtils.focusFirst,
    focusLast: accessibilityUtils.focusLast
  }

  return (
    <AccessibilityContext.Provider value={contextValue}>
      {children}
    </AccessibilityContext.Provider>
  )
}

/**
 * Hook to access accessibility context
 */
export const useAccessibilityContext = (): AccessibilityContextType => {
  const context = useContext(AccessibilityContext)
  if (!context) {
    throw new Error('useAccessibilityContext must be used within an AccessibilityProvider')
  }
  return context
}

/**
 * Hook for component-specific accessibility features
 */
export const useComponentAccessibility = (componentName: string) => {
  const { announce, settings } = useAccessibilityContext()
  
  const announceComponentState = (state: string, details?: string) => {
    const message = details 
      ? `${componentName} ${state}: ${details}`
      : `${componentName} ${state}`
    announce(message, 'polite')
  }

  const announceComponentError = (error: string) => {
    announce(`${componentName} error: ${error}`, 'assertive')
  }

  return {
    settings,
    announceComponentState,
    announceComponentError,
    announce
  }
}

/**
 * Higher-order component for adding accessibility features
 */
export function withAccessibility<P extends object>(
  Component: React.ComponentType<P>,
  componentName: string
) {
  return React.forwardRef<any, P>((props, ref) => {
    const accessibility = useComponentAccessibility(componentName)
    
    return (
      <Component
        {...props}
        ref={ref}
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        {...(props as any).accessibility && { accessibility }}
      />
    )
  })
}

export default AccessibilityProvider