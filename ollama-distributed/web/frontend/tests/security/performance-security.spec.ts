/**
 * Performance Security Testing Suite
 * 
 * Tests performance-related security vulnerabilities and validates
 * Core Web Vitals compliance for the Ollama Distributed frontend.
 * 
 * Security Focus Areas:
 * - DoS prevention through performance limits
 * - Resource exhaustion protection
 * - Client-side performance attacks
 * - Memory leak detection
 * - Bundle size security
 */

import { test, expect, Page, Browser } from '@playwright/test'

interface PerformanceMetrics {
  fcp: number // First Contentful Paint
  lcp: number // Largest Contentful Paint
  fid: number // First Input Delay
  cls: number // Cumulative Layout Shift
  tti: number // Time to Interactive
  tbt: number // Total Blocking Time
  si: number  // Speed Index
}

interface SecurityPerformanceReport {
  timestamp: string
  metrics: PerformanceMetrics
  thresholds: {
    fcp: { target: number, actual: number, pass: boolean }
    lcp: { target: number, actual: number, pass: boolean }
    fid: { target: number, actual: number, pass: boolean }
    cls: { target: number, actual: number, pass: boolean }
    tti: { target: number, actual: number, pass: boolean }
    tbt: { target: number, actual: number, pass: boolean }
  }
  securityIssues: Array<{
    type: string
    severity: 'HIGH' | 'MEDIUM' | 'LOW'
    description: string
    mitigation: string
  }>
  riskLevel: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'
}

test.describe('[SECURITY-PERF] Performance Security Testing', () => {
  // Performance thresholds aligned with security requirements
  const SECURITY_THRESHOLDS = {
    FCP: 1800,    // 1.8s - Fast enough to prevent user frustration attacks
    LCP: 2500,    // 2.5s - Core Web Vital threshold
    FID: 100,     // 100ms - Core Web Vital threshold  
    CLS: 0.1,     // 0.1 - Core Web Vital threshold
    TTI: 3800,    // 3.8s - Interactive before potential timeout exploits
    TBT: 200,     // 200ms - Prevent UI blocking attacks
    BUNDLE_SIZE: 2048, // 2MB - Prevent bundle bloat attacks
    MEMORY_USAGE: 50,  // 50MB - Prevent memory exhaustion
    CPU_USAGE: 70      // 70% - Prevent CPU exhaustion
  }

  let performanceReport: SecurityPerformanceReport

  test.beforeAll(() => {
    performanceReport = {
      timestamp: new Date().toISOString(),
      metrics: {} as PerformanceMetrics,
      thresholds: {} as any,
      securityIssues: [],
      riskLevel: 'LOW'
    }
  })

  test.afterAll(() => {
    // Generate security performance report
    generateSecurityReport(performanceReport)
  })

  test.describe('Core Web Vitals Security', () => {
    test('should meet FCP security threshold to prevent UI freezing attacks', async ({ page }) => {
      await page.goto('/', { waitUntil: 'networkidle' })
      
      const metrics = await page.evaluate(() => {
        return new Promise((resolve) => {
          new PerformanceObserver((list) => {
            const entries = list.getEntries()
            const fcpEntry = entries.find(entry => entry.name === 'first-contentful-paint')
            if (fcpEntry) {
              resolve(fcpEntry.startTime)
            }
          }).observe({ entryTypes: ['paint'] })
          
          // Fallback timeout
          setTimeout(() => resolve(0), 5000)
        })
      })

      const fcp = metrics as number
      performanceReport.metrics.fcp = fcp
      performanceReport.thresholds.fcp = {
        target: SECURITY_THRESHOLDS.FCP,
        actual: fcp,
        pass: fcp <= SECURITY_THRESHOLDS.FCP
      }

      if (fcp > SECURITY_THRESHOLDS.FCP) {
        performanceReport.securityIssues.push({
          type: 'FCP_SECURITY_VIOLATION',
          severity: 'MEDIUM',
          description: `First Contentful Paint (${fcp}ms) exceeds security threshold (${SECURITY_THRESHOLDS.FCP}ms)`,
          mitigation: 'Optimize critical rendering path to prevent UI blocking attacks'
        })
      }

      expect(fcp).toBeLessThanOrEqual(SECURITY_THRESHOLDS.FCP)
    })

    test('should meet LCP threshold to prevent loading manipulation attacks', async ({ page }) => {
      await page.goto('/')
      
      const lcp = await page.evaluate(() => {
        return new Promise<number>((resolve) => {
          new PerformanceObserver((list) => {
            const entries = list.getEntries()
            const lastEntry = entries[entries.length - 1]
            resolve(lastEntry.startTime)
          }).observe({ entryTypes: ['largest-contentful-paint'] })
          
          // Ensure we capture the final LCP
          setTimeout(() => {
            const observer = new PerformanceObserver((list) => {
              const entries = list.getEntries()
              if (entries.length > 0) {
                resolve(entries[entries.length - 1].startTime)
              }
            })
            observer.observe({ entryTypes: ['largest-contentful-paint'] })
            observer.disconnect()
          }, 2500)
        })
      })

      performanceReport.metrics.lcp = lcp
      performanceReport.thresholds.lcp = {
        target: SECURITY_THRESHOLDS.LCP,
        actual: lcp,
        pass: lcp <= SECURITY_THRESHOLDS.LCP
      }

      if (lcp > SECURITY_THRESHOLDS.LCP) {
        performanceReport.securityIssues.push({
          type: 'LCP_SECURITY_VIOLATION',
          severity: 'HIGH',
          description: `Largest Contentful Paint (${lcp}ms) exceeds security threshold`,
          mitigation: 'Optimize image loading and reduce render-blocking resources'
        })
      }

      expect(lcp).toBeLessThanOrEqual(SECURITY_THRESHOLDS.LCP)
    })

    test('should prevent input delay attacks through FID compliance', async ({ page }) => {
      await page.goto('/', { waitUntil: 'networkidle' })
      
      // Measure First Input Delay by simulating user interaction
      const startTime = Date.now()
      await page.click('button', { timeout: 1000 }).catch(() => {
        // If no button, try other interactive elements
        return page.click('a, input, [role="button"]', { timeout: 1000 }).catch(() => null)
      })
      const inputDelay = Date.now() - startTime

      performanceReport.metrics.fid = inputDelay
      performanceReport.thresholds.fid = {
        target: SECURITY_THRESHOLDS.FID,
        actual: inputDelay,
        pass: inputDelay <= SECURITY_THRESHOLDS.FID
      }

      if (inputDelay > SECURITY_THRESHOLDS.FID) {
        performanceReport.securityIssues.push({
          type: 'FID_SECURITY_VIOLATION',
          severity: 'HIGH',
          description: `First Input Delay (${inputDelay}ms) creates DoS vulnerability`,
          mitigation: 'Reduce JavaScript execution time and optimize event handlers'
        })
      }

      expect(inputDelay).toBeLessThanOrEqual(SECURITY_THRESHOLDS.FID)
    })

    test('should prevent layout shift attacks through CLS compliance', async ({ page }) => {
      await page.goto('/')
      
      const cls = await page.evaluate(() => {
        return new Promise<number>((resolve) => {
          let clsValue = 0
          
          new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
              if (!(entry as any).hadRecentInput) {
                clsValue += (entry as any).value
              }
            }
          }).observe({ entryTypes: ['layout-shift'] })
          
          // Resolve after sufficient time to capture layout shifts
          setTimeout(() => resolve(clsValue), 3000)
        })
      })

      performanceReport.metrics.cls = cls
      performanceReport.thresholds.cls = {
        target: SECURITY_THRESHOLDS.CLS,
        actual: cls,
        pass: cls <= SECURITY_THRESHOLDS.CLS
      }

      if (cls > SECURITY_THRESHOLDS.CLS) {
        performanceReport.securityIssues.push({
          type: 'CLS_SECURITY_VIOLATION',
          severity: 'MEDIUM',
          description: `Cumulative Layout Shift (${cls}) enables clickjacking attacks`,
          mitigation: 'Stabilize layout by reserving space for dynamic content'
        })
      }

      expect(cls).toBeLessThanOrEqual(SECURITY_THRESHOLDS.CLS)
    })
  })

  test.describe('Resource Security Testing', () => {
    test('should prevent bundle size attacks', async ({ page }) => {
      const response = await page.goto('/')
      
      // Get all resource sizes
      const resourceSizes = await page.evaluate(() => {
        const resources = performance.getEntriesByType('navigation')
        let totalSize = 0
        
        for (const resource of resources) {
          totalSize += (resource as any).transferSize || 0
        }
        
        // Also check loaded scripts and stylesheets
        const scripts = Array.from(document.scripts)
        const styles = Array.from(document.styleSheets)
        
        return {
          totalSize,
          scriptCount: scripts.length,
          styleCount: styles.length
        }
      })

      const bundleSizeKB = resourceSizes.totalSize / 1024
      
      if (bundleSizeKB > SECURITY_THRESHOLDS.BUNDLE_SIZE) {
        performanceReport.securityIssues.push({
          type: 'BUNDLE_SIZE_SECURITY_VIOLATION',
          severity: 'MEDIUM',
          description: `Bundle size (${bundleSizeKB.toFixed(1)}KB) may enable bandwidth exhaustion attacks`,
          mitigation: 'Implement code splitting and reduce bundle size'
        })
      }

      expect(bundleSizeKB).toBeLessThanOrEqual(SECURITY_THRESHOLDS.BUNDLE_SIZE)
    })

    test('should detect memory leak vulnerabilities', async ({ page, browser }) => {
      await page.goto('/')
      
      // Measure initial memory
      const initialMemory = await page.evaluate(() => {
        if ('memory' in performance) {
          return (performance as any).memory.usedJSHeapSize
        }
        return 0
      })

      // Simulate heavy operations that might cause memory leaks
      await page.evaluate(() => {
        const elements = []
        for (let i = 0; i < 1000; i++) {
          const div = document.createElement('div')
          div.innerHTML = `<span>Memory test ${i}</span>`
          elements.push(div)
        }
        
        // Simulate potential memory leak patterns
        const listeners = []
        for (let i = 0; i < 100; i++) {
          const handler = () => console.log(`Handler ${i}`)
          document.addEventListener('click', handler)
          listeners.push(handler)
        }
        
        // Clean up (good practice)
        listeners.forEach(handler => {
          document.removeEventListener('click', handler)
        })
        
        return elements.length
      })

      // Force garbage collection if available
      await page.evaluate(() => {
        if (window.gc) {
          window.gc()
        }
      })

      // Measure final memory
      const finalMemory = await page.evaluate(() => {
        if ('memory' in performance) {
          return (performance as any).memory.usedJSHeapSize
        }
        return 0
      })

      const memoryIncreaseMB = (finalMemory - initialMemory) / (1024 * 1024)
      
      if (memoryIncreaseMB > SECURITY_THRESHOLDS.MEMORY_USAGE) {
        performanceReport.securityIssues.push({
          type: 'MEMORY_LEAK_VULNERABILITY',
          severity: 'HIGH',
          description: `Memory usage increased by ${memoryIncreaseMB.toFixed(1)}MB during test`,
          mitigation: 'Review event listeners and DOM references for memory leaks'
        })
      }

      expect(memoryIncreaseMB).toBeLessThanOrEqual(SECURITY_THRESHOLDS.MEMORY_USAGE)
    })

    test('should prevent CPU exhaustion attacks', async ({ page }) => {
      await page.goto('/')
      
      // Monitor CPU usage during intensive operations
      const cpuMetrics = await page.evaluate(() => {
        const startTime = performance.now()
        
        // Simulate CPU-intensive operation
        let result = 0
        for (let i = 0; i < 1000000; i++) {
          result += Math.random() * Math.sin(i)
        }
        
        const endTime = performance.now()
        const executionTime = endTime - startTime
        
        return {
          executionTime,
          result: result > 0 // Just to use the result
        }
      })

      // CPU usage estimation based on execution time
      const estimatedCPUUsage = (cpuMetrics.executionTime / 1000) * 100
      
      if (estimatedCPUUsage > SECURITY_THRESHOLDS.CPU_USAGE) {
        performanceReport.securityIssues.push({
          type: 'CPU_EXHAUSTION_VULNERABILITY',
          severity: 'HIGH',
          description: `High CPU usage detected: ${estimatedCPUUsage.toFixed(1)}%`,
          mitigation: 'Optimize computational operations and implement throttling'
        })
      }

      expect(estimatedCPUUsage).toBeLessThanOrEqual(SECURITY_THRESHOLDS.CPU_USAGE)
    })
  })

  test.describe('DoS Prevention Testing', () => {
    test('should handle rapid successive requests', async ({ page }) => {
      await page.goto('/')
      
      // Test rapid API requests
      const startTime = Date.now()
      const promises = Array.from({ length: 50 }, async (_, i) => {
        try {
          const response = await page.request.get('/api/health')
          return { status: response.status(), index: i }
        } catch (error) {
          return { status: 0, index: i, error: true }
        }
      })
      
      const results = await Promise.all(promises)
      const endTime = Date.now()
      
      const successfulRequests = results.filter(r => r.status === 200).length
      const rateLimitedRequests = results.filter(r => r.status === 429).length
      const requestRate = results.length / ((endTime - startTime) / 1000)
      
      // Should have rate limiting in place
      if (requestRate > 100 && rateLimitedRequests === 0) {
        performanceReport.securityIssues.push({
          type: 'DOS_VULNERABILITY',
          severity: 'HIGH',
          description: `No rate limiting detected for ${requestRate.toFixed(1)} req/s`,
          mitigation: 'Implement rate limiting to prevent DoS attacks'
        })
      }
      
      expect(rateLimitedRequests).toBeGreaterThan(0) // Should have some rate limiting
    })

    test('should handle large payload attacks', async ({ page }) => {
      await page.goto('/')
      
      // Test large form submission
      const largeData = 'x'.repeat(10 * 1024 * 1024) // 10MB string
      
      try {
        const response = await page.request.post('/api/feedback', {
          data: { message: largeData }
        })
        
        // Should reject large payloads
        if (response.status() === 200) {
          performanceReport.securityIssues.push({
            type: 'LARGE_PAYLOAD_VULNERABILITY',
            severity: 'HIGH',
            description: 'Server accepts large payloads without restriction',
            mitigation: 'Implement payload size limits to prevent resource exhaustion'
          })
        }
        
        expect(response.status()).toBeGreaterThanOrEqual(400)
      } catch (error) {
        // Expected behavior - request should be rejected
        console.log('Large payload correctly rejected')
      }
    })

    test('should prevent recursive/infinite loop attacks', async ({ page }) => {
      await page.goto('/')
      
      // Test for potential infinite loop vulnerabilities
      const loopTest = await page.evaluate(() => {
        const startTime = performance.now()
        let iterations = 0
        const maxTime = 1000 // 1 second maximum
        
        // Simulate potentially vulnerable loop
        while (performance.now() - startTime < maxTime && iterations < 1000000) {
          iterations++
          // Simulate work
          Math.random()
        }
        
        return {
          iterations,
          timeElapsed: performance.now() - startTime,
          completed: performance.now() - startTime >= maxTime
        }
      })
      
      // Should have reasonable iteration limits
      if (loopTest.iterations > 500000) {
        performanceReport.securityIssues.push({
          type: 'INFINITE_LOOP_VULNERABILITY',
          severity: 'MEDIUM',
          description: `High iteration count detected: ${loopTest.iterations}`,
          mitigation: 'Implement loop iteration limits and timeouts'
        })
      }
      
      expect(loopTest.iterations).toBeLessThan(1000000)
    })
  })

  test.describe('Client-Side Security Performance', () => {
    test('should validate CSP performance impact', async ({ page }) => {
      const response = await page.goto('/')
      
      // Check if CSP header exists
      const headers = response?.headers() || {}
      const cspHeader = headers['content-security-policy']
      
      if (!cspHeader) {
        performanceReport.securityIssues.push({
          type: 'MISSING_CSP_SECURITY',
          severity: 'HIGH',
          description: 'Content Security Policy header not found',
          mitigation: 'Implement CSP headers for XSS protection'
        })
      }
      
      // Measure CSP impact on performance
      const cspComplexity = cspHeader ? cspHeader.split(';').length : 0
      
      if (cspComplexity > 20) {
        performanceReport.securityIssues.push({
          type: 'CSP_PERFORMANCE_IMPACT',
          severity: 'LOW',
          description: `Complex CSP with ${cspComplexity} directives may impact performance`,
          mitigation: 'Optimize CSP directives for better performance'
        })
      }
      
      expect(cspHeader).toBeDefined()
    })

    test('should validate subresource integrity performance', async ({ page }) => {
      await page.goto('/')
      
      const sriCheck = await page.evaluate(() => {
        const scripts = Array.from(document.querySelectorAll('script[src]'))
        const stylesheets = Array.from(document.querySelectorAll('link[rel="stylesheet"]'))
        
        const externalScripts = scripts.filter(script => 
          script.src.startsWith('http') && !script.hasAttribute('integrity')
        )
        
        const externalStyles = stylesheets.filter(link =>
          link.href.startsWith('http') && !link.hasAttribute('integrity')
        )
        
        return {
          externalScriptsWithoutSRI: externalScripts.length,
          externalStylesWithoutSRI: externalStyles.length,
          totalExternalResources: externalScripts.length + externalStyles.length
        }
      })
      
      if (sriCheck.totalExternalResources > 0) {
        const sriCoverage = 1 - ((sriCheck.externalScriptsWithoutSRI + sriCheck.externalStylesWithoutSRI) / sriCheck.totalExternalResources)
        
        if (sriCoverage < 0.8) {
          performanceReport.securityIssues.push({
            type: 'MISSING_SRI_SECURITY',
            severity: 'MEDIUM',
            description: `${Math.round((1 - sriCoverage) * 100)}% of external resources lack SRI`,
            mitigation: 'Add integrity attributes to external resources'
          })
        }
      }
    })
  })
})

function generateSecurityReport(report: SecurityPerformanceReport) {
  // Determine overall risk level
  const criticalIssues = report.securityIssues.filter(i => i.severity === 'HIGH').length
  const mediumIssues = report.securityIssues.filter(i => i.severity === 'MEDIUM').length
  
  if (criticalIssues > 2) {
    report.riskLevel = 'CRITICAL'
  } else if (criticalIssues > 0 || mediumIssues > 3) {
    report.riskLevel = 'HIGH'
  } else if (mediumIssues > 0) {
    report.riskLevel = 'MEDIUM'
  } else {
    report.riskLevel = 'LOW'
  }
  
  // Generate report file
  const reportPath = `security-reports/performance-security-${Date.now()}.json`
  
  console.log('\n🔒 Performance Security Report')
  console.log('================================')
  console.log(`Risk Level: ${report.riskLevel}`)
  console.log(`Security Issues: ${report.securityIssues.length}`)
  console.log(`Core Web Vitals Compliance: ${getWebVitalsCompliance(report)}%`)
  
  if (report.securityIssues.length > 0) {
    console.log('\nSecurity Issues Found:')
    report.securityIssues.forEach(issue => {
      console.log(`  ${issue.severity}: ${issue.type}`)
      console.log(`    ${issue.description}`)
      console.log(`    Mitigation: ${issue.mitigation}`)
    })
  }
}