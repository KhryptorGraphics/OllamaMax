/**
 * Performance Optimization Utilities
 * Tools for monitoring, measuring, and optimizing application performance
 */
import React from 'react'

// Performance monitoring types
export interface PerformanceMetrics {
  fcp: number // First Contentful Paint
  lcp: number // Largest Contentful Paint
  fid: number // First Input Delay
  cls: number // Cumulative Layout Shift
  ttfb: number // Time to First Byte
  memory: number // Memory usage in MB
  timing: PerformanceTiming
}

export interface ComponentMetrics {
  componentName: string
  renderTime: number
  mountTime: number
  updateCount: number
  lastUpdate: number
}

export interface NetworkMetrics {
  requestCount: number
  totalSize: number
  cacheHits: number
  cacheMisses: number
  avgLatency: number
}

// Performance measurement utilities
export class PerformanceMonitor {
  private static instance: PerformanceMonitor
  private metrics: Map<string, number> = new Map()
  private observers: Map<string, PerformanceObserver> = new Map()
  private componentMetrics: Map<string, ComponentMetrics> = new Map()

  static getInstance(): PerformanceMonitor {
    if (!PerformanceMonitor.instance) {
      PerformanceMonitor.instance = new PerformanceMonitor()
    }
    return PerformanceMonitor.instance
  }

  // Initialize Web Vitals monitoring
  initializeWebVitals(): void {
    if (typeof window === 'undefined') return

    // First Contentful Paint
    this.observePerformanceEntry('paint', (entries) => {
      entries.forEach((entry) => {
        if (entry.name === 'first-contentful-paint') {
          this.metrics.set('fcp', entry.startTime)
        }
      })
    })

    // Largest Contentful Paint
    this.observePerformanceEntry('largest-contentful-paint', (entries) => {
      entries.forEach((entry) => {
        this.metrics.set('lcp', entry.startTime)
      })
    })

    // First Input Delay
    this.observePerformanceEntry('first-input', (entries) => {
      entries.forEach((entry) => {
        const fid = (entry as any).processingStart - entry.startTime
        this.metrics.set('fid', fid)
      })
    })

    // Cumulative Layout Shift
    this.observePerformanceEntry('layout-shift', (entries) => {
      let cls = 0
      entries.forEach((entry) => {
        if (!(entry as any).hadRecentInput) {
          cls += (entry as any).value
        }
      })
      this.metrics.set('cls', cls)
    })
  }

  // Observe specific performance entry types
  private observePerformanceEntry(
    type: string,
    callback: (entries: PerformanceEntry[]) => void
  ): void {
    if (typeof window === 'undefined' || !window.PerformanceObserver) return

    try {
      const observer = new PerformanceObserver((list) => {
        callback(list.getEntries())
      })
      
      observer.observe({ type, buffered: true })
      this.observers.set(type, observer)
    } catch (error) {
      console.warn(`Failed to observe ${type}:`, error)
    }
  }

  // Measure component performance
  measureComponent(componentName: string): {
    start: () => void
    end: () => void
    update: () => void
  } {
    let mountTime = 0
    
    const existing = this.componentMetrics.get(componentName) || {
      componentName,
      renderTime: 0,
      mountTime: 0,
      updateCount: 0,
      lastUpdate: 0
    }

    return {
      start: () => {
        this.metrics.set(`${componentName}_start`, performance.now())
      },
      
      end: () => {
        const endTime = performance.now()
        const renderTime = endTime - (this.metrics.get(`${componentName}_start`) || endTime)
        
        if (mountTime === 0) {
          mountTime = renderTime
        }

        this.componentMetrics.set(componentName, {
          ...existing,
          renderTime,
          mountTime: mountTime || existing.mountTime,
          lastUpdate: endTime
        })
        
        this.metrics.delete(`${componentName}_start`)
      },
      
      update: () => {
        const current = this.componentMetrics.get(componentName)
        if (current) {
          this.componentMetrics.set(componentName, {
            ...current,
            updateCount: current.updateCount + 1,
            lastUpdate: performance.now()
          })
        }
      }
    }
  }

  // Memory usage monitoring
  getMemoryUsage(): number {
    if (typeof window === 'undefined' || !(performance as any).memory) {
      return 0
    }

    const memory = (performance as any).memory
    return Math.round(memory.usedJSHeapSize / 1024 / 1024) // MB
  }

  // Network performance monitoring
  getNetworkMetrics(): NetworkMetrics {
    if (typeof window === 'undefined') {
      return {
        requestCount: 0,
        totalSize: 0,
        cacheHits: 0,
        cacheMisses: 0,
        avgLatency: 0
      }
    }

    const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[]
    
    let totalSize = 0
    let totalLatency = 0
    let cacheHits = 0
    let cacheMisses = 0

    resources.forEach((resource) => {
      totalSize += resource.transferSize || 0
      totalLatency += resource.responseEnd - resource.requestStart
      
      if (resource.transferSize === 0 && resource.decodedBodySize > 0) {
        cacheHits++
      } else {
        cacheMisses++
      }
    })

    return {
      requestCount: resources.length,
      totalSize,
      cacheHits,
      cacheMisses,
      avgLatency: resources.length > 0 ? totalLatency / resources.length : 0
    }
  }

  // Get all collected metrics
  getAllMetrics(): PerformanceMetrics {
    const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
    
    return {
      fcp: this.metrics.get('fcp') || 0,
      lcp: this.metrics.get('lcp') || 0,
      fid: this.metrics.get('fid') || 0,
      cls: this.metrics.get('cls') || 0,
      ttfb: navigation ? navigation.responseStart - navigation.requestStart : 0,
      memory: this.getMemoryUsage(),
      timing: performance.timing
    }
  }

  // Get component performance data
  getComponentMetrics(): ComponentMetrics[] {
    return Array.from(this.componentMetrics.values())
  }

  // Performance budget checking
  checkPerformanceBudget(budgets: {
    fcp?: number
    lcp?: number
    fid?: number
    cls?: number
    memory?: number
  }): { metric: string; actual: number; budget: number; passed: boolean }[] {
    const metrics = this.getAllMetrics()
    const results: { metric: string; actual: number; budget: number; passed: boolean }[] = []

    Object.entries(budgets).forEach(([metric, budget]) => {
      const actual = metrics[metric as keyof PerformanceMetrics] as number
      results.push({
        metric,
        actual,
        budget,
        passed: actual <= budget
      })
    })

    return results
  }

  // Send metrics to analytics
  sendMetrics(endpoint: string): void {
    const metrics = this.getAllMetrics()
    const componentMetrics = this.getComponentMetrics()
    const networkMetrics = this.getNetworkMetrics()

    if (typeof window === 'undefined') return

    // Use beacon API for reliable delivery
    const data = JSON.stringify({
      url: window.location.href,
      userAgent: navigator.userAgent,
      timestamp: Date.now(),
      webVitals: {
        fcp: metrics.fcp,
        lcp: metrics.lcp,
        fid: metrics.fid,
        cls: metrics.cls,
        ttfb: metrics.ttfb
      },
      memory: metrics.memory,
      components: componentMetrics,
      network: networkMetrics
    })

    if (navigator.sendBeacon) {
      navigator.sendBeacon(endpoint, data)
    } else {
      // Fallback to fetch
      fetch(endpoint, {
        method: 'POST',
        body: data,
        headers: { 'Content-Type': 'application/json' },
        keepalive: true
      }).catch(console.error)
    }
  }

  // Cleanup observers
  cleanup(): void {
    this.observers.forEach((observer) => {
      observer.disconnect()
    })
    this.observers.clear()
    this.metrics.clear()
    this.componentMetrics.clear()
  }
}

// React hook for component performance monitoring
export function usePerformanceMonitor(componentName: string) {
  const monitor = PerformanceMonitor.getInstance()
  const measurement = monitor.measureComponent(componentName)

  React.useEffect(() => {
    measurement.start()
    return () => {
      measurement.end()
    }
  }, [])

  React.useEffect(() => {
    measurement.update()
  })

  return {
    getMetrics: () => monitor.getComponentMetrics().find(m => m.componentName === componentName),
    getAllMetrics: () => monitor.getAllMetrics()
  }
}

// Performance optimization utilities
export const performanceUtils = {
  // Debounce function calls
  debounce<T extends (...args: any[]) => any>(
    func: T,
    wait: number
  ): (...args: Parameters<T>) => void {
    let timeout: NodeJS.Timeout
    return (...args: Parameters<T>) => {
      clearTimeout(timeout)
      timeout = setTimeout(() => func(...args), wait)
    }
  },

  // Throttle function calls
  throttle<T extends (...args: any[]) => any>(
    func: T,
    limit: number
  ): (...args: Parameters<T>) => void {
    let inThrottle: boolean
    return (...args: Parameters<T>) => {
      if (!inThrottle) {
        func(...args)
        inThrottle = true
        setTimeout(() => inThrottle = false, limit)
      }
    }
  },

  // Lazy load images with intersection observer
  lazyLoadImage(img: HTMLImageElement, src: string): void {
    if ('IntersectionObserver' in window) {
      const observer = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            img.src = src
            img.classList.remove('lazy')
            observer.unobserve(img)
          }
        })
      })
      observer.observe(img)
    } else {
      // Fallback for older browsers
      img.src = src
    }
  },

  // Preload critical resources
  preloadResource(href: string, as: string): void {
    if (typeof document === 'undefined') return

    const link = document.createElement('link')
    link.rel = 'preload'
    link.href = href
    link.as = as
    document.head.appendChild(link)
  },

  // Measure bundle size impact
  measureBundleSize(): Promise<{ total: number; resources: Array<{ name: string; size: number }> }> {
    return new Promise((resolve) => {
      if (typeof window === 'undefined') {
        resolve({ total: 0, resources: [] })
        return
      }

      window.addEventListener('load', () => {
        const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[]
        
        const bundleResources = resources
          .filter(r => r.name.includes('.js') || r.name.includes('.css'))
          .map(r => ({
            name: r.name.split('/').pop() || r.name,
            size: r.transferSize || r.encodedBodySize || 0
          }))

        const total = bundleResources.reduce((sum, r) => sum + r.size, 0)

        resolve({
          total,
          resources: bundleResources.sort((a, b) => b.size - a.size)
        })
      })
    })
  }
}

// Initialize performance monitoring
if (typeof window !== 'undefined') {
  const monitor = PerformanceMonitor.getInstance()
  monitor.initializeWebVitals()

  // Send metrics on page unload
  window.addEventListener('beforeunload', () => {
    monitor.sendMetrics('/api/analytics/performance')
  })

  // Send metrics periodically for SPAs
  setInterval(() => {
    monitor.sendMetrics('/api/analytics/performance')
  }, 30000) // Every 30 seconds
}

export default PerformanceMonitor