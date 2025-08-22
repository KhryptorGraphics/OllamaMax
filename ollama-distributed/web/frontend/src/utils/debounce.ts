/**
 * Debounce utility for optimizing frequent function calls
 * Used for performance optimization in state updates
 */

export function debounce<T extends (...args: any[]) => any>(
  func: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: NodeJS.Timeout | undefined

  return (...args: Parameters<T>) => {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => func(...args), delay)
  }
}

/**
 * Throttle utility for limiting function execution frequency
 * Useful for scroll events and real-time updates
 */
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean = false

  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}

/**
 * Debounce with immediate execution option
 * Executes immediately on first call, then debounces subsequent calls
 */
export function debounceImmediate<T extends (...args: any[]) => any>(
  func: T,
  delay: number,
  immediate: boolean = false
): (...args: Parameters<T>) => void {
  let timeoutId: NodeJS.Timeout | undefined

  return (...args: Parameters<T>) => {
    const callNow = immediate && !timeoutId

    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => {
      timeoutId = undefined
      if (!immediate) func(...args)
    }, delay)

    if (callNow) func(...args)
  }
}

/**
 * Batch debounce for collecting multiple calls and executing them together
 * Useful for batching state updates or API calls
 */
export function batchDebounce<T>(
  func: (items: T[]) => void,
  delay: number
): (item: T) => void {
  let timeoutId: NodeJS.Timeout | undefined
  let items: T[] = []

  return (item: T) => {
    items.push(item)
    
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => {
      func([...items])
      items = []
    }, delay)
  }
}