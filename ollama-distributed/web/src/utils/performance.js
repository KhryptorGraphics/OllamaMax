/**
 * Performance Optimization Utilities
 * 
 * Provides code splitting, lazy loading, bundle optimization, and performance monitoring.
 */

import { lazy, Suspense } from 'react';

// Performance monitoring
class PerformanceMonitor {
  constructor() {
    this.metrics = new Map();
    this.observers = new Map();
    this.init();
  }

  init() {
    // Setup performance observers
    this.setupNavigationObserver();
    this.setupResourceObserver();
    this.setupLongTaskObserver();
    this.setupLayoutShiftObserver();
    this.setupLargestContentfulPaintObserver();
  }

  // Navigation timing
  setupNavigationObserver() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          this.recordMetric('navigation', {
            domContentLoaded: entry.domContentLoadedEventEnd - entry.domContentLoadedEventStart,
            loadComplete: entry.loadEventEnd - entry.loadEventStart,
            domInteractive: entry.domInteractive - entry.navigationStart,
            firstPaint: this.getFirstPaint(),
            firstContentfulPaint: this.getFirstContentfulPaint(),
          });
        }
      });
      
      try {
        observer.observe({ entryTypes: ['navigation'] });
        this.observers.set('navigation', observer);
      } catch (error) {
        console.warn('Navigation observer not supported:', error);
      }
    }
  }

  // Resource loading
  setupResourceObserver() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.initiatorType === 'script' || entry.initiatorType === 'link') {
            this.recordMetric('resource', {
              name: entry.name,
              type: entry.initiatorType,
              duration: entry.duration,
              size: entry.transferSize,
              cached: entry.transferSize === 0 && entry.decodedBodySize > 0,
            });
          }
        }
      });
      
      try {
        observer.observe({ entryTypes: ['resource'] });
        this.observers.set('resource', observer);
      } catch (error) {
        console.warn('Resource observer not supported:', error);
      }
    }
  }

  // Long tasks (blocking main thread)
  setupLongTaskObserver() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          this.recordMetric('longTask', {
            duration: entry.duration,
            startTime: entry.startTime,
            attribution: entry.attribution,
          });
        }
      });
      
      try {
        observer.observe({ entryTypes: ['longtask'] });
        this.observers.set('longtask', observer);
      } catch (error) {
        console.warn('Long task observer not supported:', error);
      }
    }
  }

  // Cumulative Layout Shift
  setupLayoutShiftObserver() {
    if ('PerformanceObserver' in window) {
      let cumulativeScore = 0;
      
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (!entry.hadRecentInput) {
            cumulativeScore += entry.value;
          }
        }
        
        this.recordMetric('layoutShift', {
          cumulativeScore,
          currentShift: list.getEntries()[list.getEntries().length - 1]?.value || 0,
        });
      });
      
      try {
        observer.observe({ entryTypes: ['layout-shift'] });
        this.observers.set('layout-shift', observer);
      } catch (error) {
        console.warn('Layout shift observer not supported:', error);
      }
    }
  }

  // Largest Contentful Paint
  setupLargestContentfulPaintObserver() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        const lastEntry = entries[entries.length - 1];
        
        this.recordMetric('largestContentfulPaint', {
          startTime: lastEntry.startTime,
          size: lastEntry.size,
          element: lastEntry.element?.tagName || 'unknown',
        });
      });
      
      try {
        observer.observe({ entryTypes: ['largest-contentful-paint'] });
        this.observers.set('largest-contentful-paint', observer);
      } catch (error) {
        console.warn('LCP observer not supported:', error);
      }
    }
  }

  // Get First Paint
  getFirstPaint() {
    const paintEntries = performance.getEntriesByType('paint');
    const firstPaint = paintEntries.find(entry => entry.name === 'first-paint');
    return firstPaint ? firstPaint.startTime : null;
  }

  // Get First Contentful Paint
  getFirstContentfulPaint() {
    const paintEntries = performance.getEntriesByType('paint');
    const firstContentfulPaint = paintEntries.find(entry => entry.name === 'first-contentful-paint');
    return firstContentfulPaint ? firstContentfulPaint.startTime : null;
  }

  // Record metric
  recordMetric(type, data) {
    if (!this.metrics.has(type)) {
      this.metrics.set(type, []);
    }
    
    this.metrics.get(type).push({
      ...data,
      timestamp: Date.now(),
    });

    // Emit event for external monitoring
    window.dispatchEvent(new CustomEvent('performance-metric', {
      detail: { type, data }
    }));
  }

  // Get metrics
  getMetrics(type) {
    return this.metrics.get(type) || [];
  }

  // Get performance summary
  getPerformanceSummary() {
    return {
      navigation: this.getMetrics('navigation')[0] || {},
      resources: this.getMetrics('resource'),
      longTasks: this.getMetrics('longTask'),
      layoutShift: this.getMetrics('layoutShift')[0] || {},
      largestContentfulPaint: this.getMetrics('largestContentfulPaint')[0] || {},
      memoryUsage: this.getMemoryUsage(),
      connectionInfo: this.getConnectionInfo(),
    };
  }

  // Get memory usage
  getMemoryUsage() {
    if ('memory' in performance) {
      return {
        usedJSHeapSize: performance.memory.usedJSHeapSize,
        totalJSHeapSize: performance.memory.totalJSHeapSize,
        jsHeapSizeLimit: performance.memory.jsHeapSizeLimit,
      };
    }
    return null;
  }

  // Get connection info
  getConnectionInfo() {
    if ('connection' in navigator) {
      return {
        effectiveType: navigator.connection.effectiveType,
        downlink: navigator.connection.downlink,
        rtt: navigator.connection.rtt,
        saveData: navigator.connection.saveData,
      };
    }
    return null;
  }

  // Cleanup
  destroy() {
    this.observers.forEach(observer => observer.disconnect());
    this.observers.clear();
    this.metrics.clear();
  }
}

// Code splitting utilities
export const createLazyComponent = (importFn, fallback = null) => {
  const LazyComponent = lazy(importFn);
  
  return (props) => (
    <Suspense fallback={fallback || <div>Loading...</div>}>
      <LazyComponent {...props} />
    </Suspense>
  );
};

// Bundle optimization
export const preloadRoute = (routeImport) => {
  const componentImport = routeImport();
  return componentImport;
};

// Image optimization
export const createOptimizedImage = (src, options = {}) => {
  const {
    width,
    height,
    quality = 80,
    format = 'webp',
    fallback = 'jpg'
  } = options;

  // Check WebP support
  const supportsWebP = (() => {
    const canvas = document.createElement('canvas');
    canvas.width = 1;
    canvas.height = 1;
    return canvas.toDataURL('image/webp').indexOf('data:image/webp') === 0;
  })();

  const optimizedSrc = supportsWebP ? 
    `${src}?format=${format}&quality=${quality}${width ? `&w=${width}` : ''}${height ? `&h=${height}` : ''}` :
    `${src}?format=${fallback}&quality=${quality}${width ? `&w=${width}` : ''}${height ? `&h=${height}` : ''}`;

  return optimizedSrc;
};

// Debounce utility for performance
export const debounce = (func, wait, immediate = false) => {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      timeout = null;
      if (!immediate) func(...args);
    };
    const callNow = immediate && !timeout;
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
    if (callNow) func(...args);
  };
};

// Throttle utility for performance
export const throttle = (func, limit) => {
  let inThrottle;
  return function(...args) {
    if (!inThrottle) {
      func.apply(this, args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
};

// Intersection Observer utility
export const createIntersectionObserver = (callback, options = {}) => {
  const defaultOptions = {
    root: null,
    rootMargin: '0px',
    threshold: 0.1,
  };

  const observerOptions = { ...defaultOptions, ...options };

  if ('IntersectionObserver' in window) {
    return new IntersectionObserver(callback, observerOptions);
  }

  // Fallback for browsers without IntersectionObserver
  return {
    observe: () => {},
    unobserve: () => {},
    disconnect: () => {},
  };
};

// Virtual scrolling utility
export const createVirtualList = (items, itemHeight, containerHeight) => {
  const visibleCount = Math.ceil(containerHeight / itemHeight);
  const bufferSize = Math.floor(visibleCount / 2);
  
  return {
    getVisibleRange: (scrollTop) => {
      const startIndex = Math.floor(scrollTop / itemHeight);
      const endIndex = Math.min(startIndex + visibleCount + bufferSize, items.length - 1);
      const bufferedStartIndex = Math.max(0, startIndex - bufferSize);
      
      return {
        startIndex: bufferedStartIndex,
        endIndex,
        visibleItems: items.slice(bufferedStartIndex, endIndex + 1),
        offsetY: bufferedStartIndex * itemHeight,
        totalHeight: items.length * itemHeight,
      };
    },
  };
};

// Performance budget checker
export const checkPerformanceBudget = (budget) => {
  const summary = performanceMonitor.getPerformanceSummary();
  const violations = [];

  // Check FCP budget
  if (budget.firstContentfulPaint && summary.navigation.firstContentfulPaint > budget.firstContentfulPaint) {
    violations.push({
      metric: 'First Contentful Paint',
      actual: summary.navigation.firstContentfulPaint,
      budget: budget.firstContentfulPaint,
    });
  }

  // Check LCP budget
  if (budget.largestContentfulPaint && summary.largestContentfulPaint.startTime > budget.largestContentfulPaint) {
    violations.push({
      metric: 'Largest Contentful Paint',
      actual: summary.largestContentfulPaint.startTime,
      budget: budget.largestContentfulPaint,
    });
  }

  // Check CLS budget
  if (budget.cumulativeLayoutShift && summary.layoutShift.cumulativeScore > budget.cumulativeLayoutShift) {
    violations.push({
      metric: 'Cumulative Layout Shift',
      actual: summary.layoutShift.cumulativeScore,
      budget: budget.cumulativeLayoutShift,
    });
  }

  // Check bundle size budget
  const totalBundleSize = summary.resources
    .filter(resource => resource.type === 'script')
    .reduce((total, resource) => total + resource.size, 0);

  if (budget.bundleSize && totalBundleSize > budget.bundleSize) {
    violations.push({
      metric: 'Bundle Size',
      actual: totalBundleSize,
      budget: budget.bundleSize,
    });
  }

  return {
    passed: violations.length === 0,
    violations,
    summary,
  };
};

// Create singleton performance monitor
export const performanceMonitor = new PerformanceMonitor();

// Export utilities
export default {
  performanceMonitor,
  createLazyComponent,
  preloadRoute,
  createOptimizedImage,
  debounce,
  throttle,
  createIntersectionObserver,
  createVirtualList,
  checkPerformanceBudget,
};
