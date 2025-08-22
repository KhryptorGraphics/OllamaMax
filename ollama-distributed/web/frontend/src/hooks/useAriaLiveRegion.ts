/**
 * @fileoverview useAriaLiveRegion Hook
 * Provides utilities for managing ARIA live regions for dynamic content announcements
 */

import { useCallback, useEffect, useRef } from 'react'

export interface AriaLiveRegionOptions {
  /** The politeness level of the live region */
  politeness?: 'off' | 'polite' | 'assertive'
  /** Whether to clear previous announcements */
  clearOnAnnounce?: boolean
  /** Delay before announcing (in milliseconds) */
  delay?: number
  /** Whether the entire region should be announced when changed */
  atomic?: boolean
}

/**
 * Hook for managing ARIA live regions
 * 
 * Provides utilities for creating and managing live regions that announce
 * dynamic content changes to screen readers.
 */
export function useAriaLiveRegion(options: AriaLiveRegionOptions = {}) {
  const {
    politeness = 'polite',
    clearOnAnnounce = false,
    delay = 0,
    atomic = true
  } = options

  const liveRegionRef = useRef<HTMLDivElement | null>(null)
  const timeoutRef = useRef<NodeJS.Timeout>()

  // Create live region element
  useEffect(() => {
    if (!liveRegionRef.current) {
      const liveRegion = document.createElement('div')
      liveRegion.setAttribute('aria-live', politeness)
      liveRegion.setAttribute('aria-atomic', atomic.toString())
      liveRegion.className = 'sr-only'
      liveRegion.style.cssText = `
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
      
      document.body.appendChild(liveRegion)
      liveRegionRef.current = liveRegion
    }

    return () => {
      if (liveRegionRef.current) {
        document.body.removeChild(liveRegionRef.current)
        liveRegionRef.current = null
      }
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [politeness, atomic])

  // Function to announce messages
  const announce = useCallback((message: string) => {
    if (!liveRegionRef.current) return

    const liveRegion = liveRegionRef.current

    // Clear existing timeout
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    const doAnnounce = () => {
      if (clearOnAnnounce) {
        liveRegion.textContent = ''
        // Small delay to ensure screen readers notice the change
        setTimeout(() => {
          liveRegion.textContent = message
        }, 50)
      } else {
        liveRegion.textContent = message
      }
    }

    if (delay > 0) {
      timeoutRef.current = setTimeout(doAnnounce, delay)
    } else {
      doAnnounce()
    }
  }, [clearOnAnnounce, delay])

  // Function to clear the live region
  const clear = useCallback(() => {
    if (liveRegionRef.current) {
      liveRegionRef.current.textContent = ''
    }
  }, [])

  // Function to update politeness level
  const setPoliteness = useCallback((newPoliteness: 'off' | 'polite' | 'assertive') => {
    if (liveRegionRef.current) {
      liveRegionRef.current.setAttribute('aria-live', newPoliteness)
    }
  }, [])

  return {
    announce,
    clear,
    setPoliteness,
    isSupported: typeof document !== 'undefined'
  }
}

/**
 * Hook for announcing status messages (polite)
 */
export function useStatusAnnouncer() {
  return useAriaLiveRegion({
    politeness: 'polite',
    clearOnAnnounce: true,
    delay: 100
  })
}

/**
 * Hook for announcing alerts and errors (assertive)
 */
export function useAlertAnnouncer() {
  return useAriaLiveRegion({
    politeness: 'assertive',
    clearOnAnnounce: true,
    delay: 0
  })
}

/**
 * Hook for announcing loading states
 */
export function useLoadingAnnouncer() {
  const { announce, clear } = useAriaLiveRegion({
    politeness: 'polite',
    clearOnAnnounce: false,
    delay: 500 // Delay to avoid announcing very quick loads
  })

  const announceLoading = useCallback((message = 'Loading...') => {
    announce(message)
  }, [announce])

  const announceLoaded = useCallback((message = 'Content loaded') => {
    announce(message)
  }, [announce])

  return {
    announceLoading,
    announceLoaded,
    clear
  }
}

export default useAriaLiveRegion