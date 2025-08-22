/**
 * @fileoverview Focus Trap Component
 * Provides focus management for modal dialogs and other contained interfaces
 */

import React, { useEffect, useRef, ReactNode } from 'react'
import { useAccessibilityContext } from './AccessibilityProvider'

interface FocusTrapProps {
  /** Whether focus trapping is active */
  active?: boolean
  /** Children to render within the focus trap */
  children: ReactNode
  /** Element to focus when trap activates */
  initialFocus?: HTMLElement | string
  /** Element to focus when trap deactivates */
  finalFocus?: HTMLElement
  /** Additional CSS classes */
  className?: string
  /** Callback when focus trap activates */
  onActivate?: () => void
  /** Callback when focus trap deactivates */
  onDeactivate?: () => void
}

/**
 * Focus Trap Component
 * 
 * Automatically manages focus within a container, ensuring keyboard users
 * cannot tab outside of the trapped area. Essential for modal dialogs,
 * dropdown menus, and other overlay components.
 */
export const FocusTrap: React.FC<FocusTrapProps> = ({
  active = true,
  children,
  initialFocus,
  finalFocus,
  className,
  onActivate,
  onDeactivate
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const previousActiveElement = useRef<HTMLElement | null>(null)
  const { settings, trapFocus, saveFocus, restoreFocus } = useAccessibilityContext()

  useEffect(() => {
    if (!active || !settings.focusTrapping) return

    const container = containerRef.current
    if (!container) return

    // Save current focus
    previousActiveElement.current = document.activeElement as HTMLElement
    saveFocus()

    // Set initial focus
    let elementToFocus: HTMLElement | null = null

    if (typeof initialFocus === 'string') {
      elementToFocus = container.querySelector(initialFocus)
    } else if (initialFocus instanceof HTMLElement) {
      elementToFocus = initialFocus
    } else {
      // Focus first focusable element
      const focusableElements = getFocusableElements(container)
      elementToFocus = focusableElements[0] || null
    }

    if (elementToFocus) {
      elementToFocus.focus()
    }

    // Set up focus trap
    const cleanup = trapFocus(container)

    onActivate?.()

    return () => {
      cleanup()
      onDeactivate?.()

      // Restore focus
      if (finalFocus) {
        finalFocus.focus()
      } else if (previousActiveElement.current) {
        restoreFocus()
      }
    }
  }, [active, initialFocus, finalFocus, settings.focusTrapping, trapFocus, saveFocus, restoreFocus, onActivate, onDeactivate])

  return (
    <div ref={containerRef} className={className}>
      {children}
    </div>
  )
}

/**
 * Get all focusable elements within a container
 */
function getFocusableElements(container: HTMLElement): HTMLElement[] {
  const focusableSelectors = [
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    'a[href]',
    '[tabindex]:not([tabindex="-1"])',
    '[contenteditable="true"]'
  ].join(', ')

  return Array.from(container.querySelectorAll(focusableSelectors))
    .filter((element): element is HTMLElement => {
      const htmlElement = element as HTMLElement
      return isVisible(htmlElement) && !isInert(htmlElement)
    })
}

/**
 * Check if element is visible
 */
function isVisible(element: HTMLElement): boolean {
  const style = window.getComputedStyle(element)
  return (
    style.display !== 'none' &&
    style.visibility !== 'hidden' &&
    style.opacity !== '0' &&
    element.offsetWidth > 0 &&
    element.offsetHeight > 0
  )
}

/**
 * Check if element is inert (disabled or aria-hidden)
 */
function isInert(element: HTMLElement): boolean {
  if (
    element.hasAttribute('disabled') ||
    element.getAttribute('aria-hidden') === 'true'
  ) {
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

export default FocusTrap