/**
 * Accessibility Enhancement Utilities
 * Tools for improving accessibility compliance and user experience
 */
import React from 'react'

// Accessibility types
export interface AccessibilityOptions {
  announcePageChanges?: boolean
  focusManagement?: boolean
  keyboardNavigation?: boolean
  highContrast?: boolean
  reducedMotion?: boolean
}

export interface FocusableElement {
  element: HTMLElement
  tabIndex: number
  role?: string
}

export interface AnnouncementOptions {
  priority?: 'polite' | 'assertive'
  delay?: number
  clear?: boolean
}

// Screen reader announcement system
export class ScreenReaderAnnouncer {
  private static instance: ScreenReaderAnnouncer
  private politeRegion: HTMLElement | null = null
  private assertiveRegion: HTMLElement | null = null

  static getInstance(): ScreenReaderAnnouncer {
    if (!ScreenReaderAnnouncer.instance) {
      ScreenReaderAnnouncer.instance = new ScreenReaderAnnouncer()
    }
    return ScreenReaderAnnouncer.instance
  }

  initialize(): void {
    if (typeof document === 'undefined') return

    // Create polite announcement region
    this.politeRegion = document.createElement('div')
    this.politeRegion.setAttribute('aria-live', 'polite')
    this.politeRegion.setAttribute('aria-atomic', 'true')
    this.politeRegion.className = 'sr-only'
    this.politeRegion.style.cssText = `
      position: absolute !important;
      width: 1px !important;
      height: 1px !important;
      padding: 0 !important;
      margin: -1px !important;
      overflow: hidden !important;
      clip: rect(0, 0, 0, 0) !important;
      white-space: nowrap !important;
      border: 0 !important;
    `

    // Create assertive announcement region
    this.assertiveRegion = document.createElement('div')
    this.assertiveRegion.setAttribute('aria-live', 'assertive')
    this.assertiveRegion.setAttribute('aria-atomic', 'true')
    this.assertiveRegion.className = 'sr-only'
    this.assertiveRegion.style.cssText = this.politeRegion.style.cssText

    // Add to DOM
    document.body.appendChild(this.politeRegion)
    document.body.appendChild(this.assertiveRegion)
  }

  announce(message: string, options: AnnouncementOptions = {}): void {
    const { priority = 'polite', delay = 0, clear = false } = options
    const region = priority === 'assertive' ? this.assertiveRegion : this.politeRegion

    if (!region) {
      this.initialize()
      return this.announce(message, options)
    }

    const doAnnounce = () => {
      if (clear) {
        region.textContent = ''
      }
      
      // Use a slight delay to ensure screen readers pick up the change
      setTimeout(() => {
        region.textContent = message
      }, 50)
    }

    if (delay > 0) {
      setTimeout(doAnnounce, delay)
    } else {
      doAnnounce()
    }
  }

  announcePageChange(title: string): void {
    this.announce(`Navigated to ${title}`, { priority: 'polite', delay: 100 })
  }

  announceLoadingState(isLoading: boolean, context?: string): void {
    const message = isLoading 
      ? `Loading${context ? ` ${context}` : ''}...`
      : `${context || 'Content'} loaded`
    
    this.announce(message, { priority: 'polite' })
  }

  announceError(error: string): void {
    this.announce(`Error: ${error}`, { priority: 'assertive' })
  }

  announceSuccess(message: string): void {
    this.announce(`Success: ${message}`, { priority: 'polite' })
  }
}

// Focus management utilities
export class FocusManager {
  private static instance: FocusManager
  private focusStack: HTMLElement[] = []
  private lastFocused: HTMLElement | null = null

  static getInstance(): FocusManager {
    if (!FocusManager.instance) {
      FocusManager.instance = new FocusManager()
    }
    return FocusManager.instance
  }

  // Get all focusable elements in a container
  getFocusableElements(container: HTMLElement = document.body): FocusableElement[] {
    const focusableSelectors = [
      'a[href]',
      'button:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      'textarea:not([disabled])',
      '[tabindex]:not([tabindex="-1"])',
      '[contenteditable="true"]'
    ].join(', ')

    const elements = Array.from(container.querySelectorAll(focusableSelectors)) as HTMLElement[]
    
    return elements
      .filter(el => this.isVisible(el) && !this.isInert(el))
      .map(el => ({
        element: el,
        tabIndex: parseInt(el.getAttribute('tabindex') || '0'),
        role: el.getAttribute('role') || undefined
      }))
      .sort((a, b) => {
        // Sort by tab index, then by DOM order
        if (a.tabIndex !== b.tabIndex) {
          if (a.tabIndex === 0) return 1
          if (b.tabIndex === 0) return -1
          return a.tabIndex - b.tabIndex
        }
        return 0
      })
  }

  // Check if element is visible
  private isVisible(element: HTMLElement): boolean {
    const style = window.getComputedStyle(element)
    return style.display !== 'none' && 
           style.visibility !== 'hidden' && 
           style.opacity !== '0' &&
           element.offsetWidth > 0 && 
           element.offsetHeight > 0
  }

  // Check if element is inert (disabled or aria-hidden)
  private isInert(element: HTMLElement): boolean {
    if (element.hasAttribute('disabled') || element.getAttribute('aria-hidden') === 'true') {
      return true
    }
    
    // Check if any parent has aria-hidden
    let parent = element.parentElement
    while (parent) {
      if (parent.getAttribute('aria-hidden') === 'true') {
        return true
      }
      parent = parent.parentElement
    }
    
    return false
  }

  // Focus first focusable element in container
  focusFirst(container: HTMLElement = document.body): boolean {
    const focusable = this.getFocusableElements(container)
    if (focusable.length > 0) {
      focusable[0].element.focus()
      return true
    }
    return false
  }

  // Focus last focusable element in container
  focusLast(container: HTMLElement = document.body): boolean {
    const focusable = this.getFocusableElements(container)
    if (focusable.length > 0) {
      focusable[focusable.length - 1].element.focus()
      return true
    }
    return false
  }

  // Trap focus within container (for modals, dialogs)
  trapFocus(container: HTMLElement): () => void {
    const focusable = this.getFocusableElements(container)
    if (focusable.length === 0) return () => {}

    const firstElement = focusable[0].element
    const lastElement = focusable[focusable.length - 1].element

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Tab') return

      if (event.shiftKey) {
        // Shift + Tab
        if (document.activeElement === firstElement) {
          event.preventDefault()
          lastElement.focus()
        }
      } else {
        // Tab
        if (document.activeElement === lastElement) {
          event.preventDefault()
          firstElement.focus()
        }
      }
    }

    container.addEventListener('keydown', handleKeyDown)
    
    // Focus first element initially
    firstElement.focus()

    // Return cleanup function
    return () => {
      container.removeEventListener('keydown', handleKeyDown)
    }
  }

  // Save current focus to restore later
  saveFocus(): void {
    this.lastFocused = document.activeElement as HTMLElement
    if (this.lastFocused) {
      this.focusStack.push(this.lastFocused)
    }
  }

  // Restore previously saved focus
  restoreFocus(): boolean {
    const element = this.focusStack.pop() || this.lastFocused
    if (element && this.isVisible(element) && !this.isInert(element)) {
      element.focus()
      return true
    }
    return false
  }

  // Clear focus stack
  clearFocusStack(): void {
    this.focusStack = []
    this.lastFocused = null
  }
}

// Keyboard navigation utilities
export class KeyboardNavigationManager {
  private static instance: KeyboardNavigationManager
  // Commented out until implementation is complete
  // private _keyHandlers: Map<string, (event: KeyboardEvent) => void> = new Map()

  static getInstance(): KeyboardNavigationManager {
    if (!KeyboardNavigationManager.instance) {
      KeyboardNavigationManager.instance = new KeyboardNavigationManager()
    }
    return KeyboardNavigationManager.instance
  }

  // Handle arrow key navigation for lists/grids
  handleArrowNavigation(
    container: HTMLElement,
    options: {
      orientation?: 'horizontal' | 'vertical' | 'both'
      wrap?: boolean
      columns?: number
    } = {}
  ): () => void {
    const { orientation = 'vertical', wrap = true, columns } = options
    
    const handleKeyDown = (event: KeyboardEvent) => {
      const focusable = FocusManager.getInstance().getFocusableElements(container)
      const currentIndex = focusable.findIndex(item => item.element === document.activeElement)
      
      if (currentIndex === -1) return

      let targetIndex = currentIndex
      
      switch (event.key) {
        case 'ArrowDown':
          if (orientation === 'vertical' || orientation === 'both') {
            event.preventDefault()
            if (columns) {
              targetIndex = Math.min(currentIndex + columns, focusable.length - 1)
            } else {
              targetIndex = wrap 
                ? (currentIndex + 1) % focusable.length
                : Math.min(currentIndex + 1, focusable.length - 1)
            }
          }
          break
          
        case 'ArrowUp':
          if (orientation === 'vertical' || orientation === 'both') {
            event.preventDefault()
            if (columns) {
              targetIndex = Math.max(currentIndex - columns, 0)
            } else {
              targetIndex = wrap
                ? currentIndex === 0 ? focusable.length - 1 : currentIndex - 1
                : Math.max(currentIndex - 1, 0)
            }
          }
          break
          
        case 'ArrowRight':
          if (orientation === 'horizontal' || orientation === 'both') {
            event.preventDefault()
            targetIndex = wrap
              ? (currentIndex + 1) % focusable.length
              : Math.min(currentIndex + 1, focusable.length - 1)
          }
          break
          
        case 'ArrowLeft':
          if (orientation === 'horizontal' || orientation === 'both') {
            event.preventDefault()
            targetIndex = wrap
              ? currentIndex === 0 ? focusable.length - 1 : currentIndex - 1
              : Math.max(currentIndex - 1, 0)
          }
          break
          
        case 'Home':
          event.preventDefault()
          targetIndex = 0
          break
          
        case 'End':
          event.preventDefault()
          targetIndex = focusable.length - 1
          break
      }
      
      if (targetIndex !== currentIndex && focusable[targetIndex]) {
        focusable[targetIndex].element.focus()
      }
    }

    container.addEventListener('keydown', handleKeyDown)
    
    return () => {
      container.removeEventListener('keydown', handleKeyDown)
    }
  }

  // Handle escape key to close modals/dropdowns
  handleEscapeKey(callback: () => void): () => void {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault()
        callback()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }

  // Handle Enter/Space activation
  handleActivationKeys(element: HTMLElement, callback: () => void): () => void {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault()
        callback()
      }
    }

    element.addEventListener('keydown', handleKeyDown)
    
    return () => {
      element.removeEventListener('keydown', handleKeyDown)
    }
  }
}

// User preferences detection
class AccessibilityPreferencesClass {
  static detectPreferences(): {
    reducedMotion: boolean
    highContrast: boolean
    darkMode: boolean
    largeFonts: boolean
  } {
    if (typeof window === 'undefined') {
      return {
        reducedMotion: false,
        highContrast: false,
        darkMode: false,
        largeFonts: false
      }
    }

    return {
      reducedMotion: window.matchMedia('(prefers-reduced-motion: reduce)').matches,
      highContrast: window.matchMedia('(prefers-contrast: high)').matches,
      darkMode: window.matchMedia('(prefers-color-scheme: dark)').matches,
      largeFonts: window.matchMedia('(min-resolution: 192dpi)').matches
    }
  }

  static applyPreferences(preferences = AccessibilityPreferencesClass.detectPreferences()): void {
    if (typeof document === 'undefined') return

    const root = document.documentElement

    // Apply reduced motion
    if (preferences.reducedMotion) {
      root.style.setProperty('--animation-duration', '0.01ms')
      root.style.setProperty('--transition-duration', '0.01ms')
    }

    // Apply high contrast
    if (preferences.highContrast) {
      root.classList.add('high-contrast')
    }

    // Apply dark mode
    if (preferences.darkMode) {
      root.classList.add('dark')
    }

    // Apply large fonts
    if (preferences.largeFonts) {
      root.style.setProperty('--base-font-size', '18px')
    }
  }
}

// Accessibility validation utilities
export const accessibilityUtils = {
  // Validate color contrast
  checkColorContrast(foreground: string, background: string): {
    ratio: number
    aa: boolean
    aaa: boolean
  } {
    const getLuminance = (color: string): number => {
      // Simplified luminance calculation
      const rgb = color.match(/\d+/g)?.map(Number) || [0, 0, 0]
      const [r, g, b] = rgb.map(c => {
        c = c / 255
        return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
      })
      return 0.2126 * r + 0.7152 * g + 0.0722 * b
    }

    const l1 = getLuminance(foreground)
    const l2 = getLuminance(background)
    const ratio = (Math.max(l1, l2) + 0.05) / (Math.min(l1, l2) + 0.05)

    return {
      ratio,
      aa: ratio >= 4.5,
      aaa: ratio >= 7
    }
  },

  // Check for proper heading structure
  validateHeadingStructure(): { valid: boolean; issues: string[] } {
    if (typeof document === 'undefined') {
      return { valid: true, issues: [] }
    }

    const headings = Array.from(document.querySelectorAll('h1, h2, h3, h4, h5, h6'))
    const issues: string[] = []
    let lastLevel = 0

    // Check for h1
    const h1Count = document.querySelectorAll('h1').length
    if (h1Count === 0) {
      issues.push('No h1 element found')
    } else if (h1Count > 1) {
      issues.push('Multiple h1 elements found')
    }

    // Check heading hierarchy
    headings.forEach((heading, index) => {
      const level = parseInt(heading.tagName.charAt(1))
      
      if (index === 0 && level !== 1) {
        issues.push('First heading should be h1')
      }
      
      if (level > lastLevel + 1) {
        issues.push(`Heading level jumps from h${lastLevel} to h${level}`)
      }
      
      lastLevel = level
    })

    return {
      valid: issues.length === 0,
      issues
    }
  },

  // Check for alt text on images
  validateImageAltText(): { valid: boolean; issues: string[] } {
    if (typeof document === 'undefined') {
      return { valid: true, issues: [] }
    }

    const images = Array.from(document.querySelectorAll('img'))
    const issues: string[] = []

    images.forEach((img, index) => {
      if (!img.hasAttribute('alt')) {
        issues.push(`Image ${index + 1} missing alt attribute`)
      } else if (img.getAttribute('alt') === '') {
        // Empty alt is valid for decorative images
        if (!img.hasAttribute('aria-hidden') && img.getAttribute('role') !== 'presentation') {
          issues.push(`Image ${index + 1} has empty alt but is not marked as decorative`)
        }
      }
    })

    return {
      valid: issues.length === 0,
      issues
    }
  },

  // Check form accessibility
  validateFormAccessibility(): { valid: boolean; issues: string[] } {
    if (typeof document === 'undefined') {
      return { valid: true, issues: [] }
    }

    const formControls = Array.from(document.querySelectorAll('input, select, textarea'))
    const issues: string[] = []

    formControls.forEach((control, index) => {
      const id = control.id
      const type = control.getAttribute('type')
      
      // Check for labels
      if (type !== 'hidden' && type !== 'submit' && type !== 'button') {
        const hasLabel = !!document.querySelector(`label[for="${id}"]`) ||
                         !!control.getAttribute('aria-label') ||
                         !!control.getAttribute('aria-labelledby')
        
        if (!hasLabel) {
          issues.push(`Form control ${index + 1} missing label`)
        }
      }
      
      // Check required fields
      if (control.hasAttribute('required') && !control.getAttribute('aria-required')) {
        control.setAttribute('aria-required', 'true')
      }
    })

    return {
      valid: issues.length === 0,
      issues
    }
  }
}

// React hook for accessibility features
export function useAccessibility(options: AccessibilityOptions = {}) {
  const announcer = ScreenReaderAnnouncer.getInstance()
  const focusManager = FocusManager.getInstance()
  const keyboardManager = KeyboardNavigationManager.getInstance()

  React.useEffect(() => {
    announcer.initialize()
    
    if (options.announcePageChanges) {
      const title = document.title
      announcer.announcePageChange(title)
    }

    // Apply user preferences
    AccessibilityPreferencesClass.applyPreferences()
  }, [])

  return {
    announce: announcer.announce.bind(announcer),
    announcePageChange: announcer.announcePageChange.bind(announcer),
    announceLoadingState: announcer.announceLoadingState.bind(announcer),
    announceError: announcer.announceError.bind(announcer),
    announceSuccess: announcer.announceSuccess.bind(announcer),
    saveFocus: focusManager.saveFocus.bind(focusManager),
    restoreFocus: focusManager.restoreFocus.bind(focusManager),
    trapFocus: focusManager.trapFocus.bind(focusManager),
    focusFirst: focusManager.focusFirst.bind(focusManager),
    focusLast: focusManager.focusLast.bind(focusManager),
    handleArrowNavigation: keyboardManager.handleArrowNavigation.bind(keyboardManager),
    handleEscapeKey: keyboardManager.handleEscapeKey.bind(keyboardManager),
    handleActivationKeys: keyboardManager.handleActivationKeys.bind(keyboardManager)
  }
}

// Initialize accessibility features
if (typeof window !== 'undefined') {
  const announcer = ScreenReaderAnnouncer.getInstance()
  announcer.initialize()
  
  // Apply user preferences on load
  AccessibilityPreferencesClass.applyPreferences()
  
  // Listen for preference changes
  const mediaQueries = [
    window.matchMedia('(prefers-reduced-motion: reduce)'),
    window.matchMedia('(prefers-contrast: high)'),
    window.matchMedia('(prefers-color-scheme: dark)')
  ]

  mediaQueries.forEach(mq => {
    mq.addEventListener('change', () => {
      AccessibilityPreferencesClass.applyPreferences()
    })
  })
}

// Export the class with static methods
export { AccessibilityPreferencesClass as AccessibilityPreferences }