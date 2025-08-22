/**
 * @fileoverview Accessibility Announcer Component
 * Provides a React component for making screen reader announcements
 */

import React, { useEffect } from 'react'
import { useAriaLiveRegion } from '@/hooks/useAriaLiveRegion'

interface AccessibilityAnnouncerProps {
  /** Message to announce */
  message?: string
  /** Priority level for announcements */
  priority?: 'polite' | 'assertive'
  /** Whether to clear the message after announcing */
  clearAfterDelay?: number
  /** Additional CSS classes */
  className?: string
}

/**
 * Accessibility Announcer Component
 * 
 * A React component that creates ARIA live regions for announcing
 * dynamic content changes to screen readers. Use this component
 * to announce status updates, errors, and other important information.
 */
export const AccessibilityAnnouncer: React.FC<AccessibilityAnnouncerProps> = ({
  message,
  priority = 'polite',
  clearAfterDelay,
  className
}) => {
  const { announce, clear } = useAriaLiveRegion({
    politeness: priority,
    clearOnAnnounce: true
  })

  useEffect(() => {
    if (message) {
      announce(message)

      if (clearAfterDelay && clearAfterDelay > 0) {
        const timer = setTimeout(() => {
          clear()
        }, clearAfterDelay)

        return () => clearTimeout(timer)
      }
    }
  }, [message, announce, clear, clearAfterDelay])

  // This component doesn't render anything visible
  return (
    <div
      className={`accessibility-announcer sr-only ${className || ''}`}
      aria-hidden="true"
    />
  )
}

/**
 * Status Announcer Component
 * For polite status updates
 */
export const StatusAnnouncer: React.FC<Omit<AccessibilityAnnouncerProps, 'priority'>> = (props) => (
  <AccessibilityAnnouncer {...props} priority="polite" />
)

/**
 * Alert Announcer Component
 * For urgent alerts and errors
 */
export const AlertAnnouncer: React.FC<Omit<AccessibilityAnnouncerProps, 'priority'>> = (props) => (
  <AccessibilityAnnouncer {...props} priority="assertive" />
)

export default AccessibilityAnnouncer