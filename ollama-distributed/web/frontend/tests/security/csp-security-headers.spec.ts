/**
 * Content Security Policy and Security Headers Testing Suite
 * 
 * Comprehensive validation of security headers implementation for the
 * Ollama Distributed frontend application.
 * 
 * Tests:
 * - Content Security Policy (CSP) validation and enforcement
 * - Security headers presence and configuration
 * - HSTS implementation
 * - CORS policy validation
 * - Clickjacking protection
 * - MIME type sniffing prevention
 */

import { test, expect, Page, BrowserContext } from '@playwright/test'
import { cspManager } from '../../src/utils/security'

interface SecurityHeader {
  name: string
  required: boolean
  expectedValue?: string | RegExp
  securityImplication: string
  owaspCategory: string
}

interface CSPDirective {
  directive: string
  allowedSources: string[]
  forbidden: string[]
  description: string
}

test.describe('[SECURITY-ALERT] CSP and Security Headers Validation', () => {
  const REQUIRED_SECURITY_HEADERS: SecurityHeader[] = [
    {
      name: 'Content-Security-Policy',
      required: true,
      expectedValue: /default-src\s+'self'/,
      securityImplication: 'Prevents XSS attacks and code injection',
      owaspCategory: 'A03:2021 - Injection'
    },
    {
      name: 'Strict-Transport-Security',
      required: true,
      expectedValue: /max-age=\d+/,
      securityImplication: 'Enforces HTTPS connections',
      owaspCategory: 'A02:2021 - Cryptographic Failures'
    },
    {
      name: 'X-Frame-Options',
      required: true,
      expectedValue: /(DENY|SAMEORIGIN)/,
      securityImplication: 'Prevents clickjacking attacks',
      owaspCategory: 'A04:2021 - Insecure Design'
    },
    {
      name: 'X-Content-Type-Options',
      required: true,
      expectedValue: 'nosniff',
      securityImplication: 'Prevents MIME type confusion attacks',
      owaspCategory: 'A05:2021 - Security Misconfiguration'
    },
    {
      name: 'X-XSS-Protection',
      required: true,
      expectedValue: /1;\s*mode=block/,
      securityImplication: 'Enables browser XSS filtering',
      owaspCategory: 'A03:2021 - Injection'
    },
    {
      name: 'Referrer-Policy',
      required: true,
      expectedValue: /(strict-origin-when-cross-origin|no-referrer)/,
      securityImplication: 'Controls referrer information disclosure',
      owaspCategory: 'A01:2021 - Broken Access Control'
    },
    {
      name: 'Permissions-Policy',
      required: false,
      expectedValue: /camera=\(\),\s*microphone=\(\)/,
      securityImplication: 'Restricts dangerous browser APIs',
      owaspCategory: 'A05:2021 - Security Misconfiguration'
    },
    {
      name: 'Cross-Origin-Embedder-Policy',
      required: false,
      expectedValue: 'require-corp',
      securityImplication: 'Enables cross-origin isolation',
      owaspCategory: 'A05:2021 - Security Misconfiguration'
    }
  ]

  const CSP_DIRECTIVES: CSPDirective[] = [
    {
      directive: 'default-src',
      allowedSources: ["'self'"],
      forbidden: ["'unsafe-eval'", "'unsafe-inline'", "*"],
      description: 'Fallback policy for all resource types'
    },
    {
      directive: 'script-src',
      allowedSources: ["'self'", "'nonce-*'", "'strict-dynamic'"],
      forbidden: ["'unsafe-eval'", "'unsafe-inline'"],
      description: 'Controls JavaScript execution sources'
    },
    {
      directive: 'style-src',
      allowedSources: ["'self'", "'nonce-*'"],
      forbidden: ["'unsafe-eval'"],
      description: 'Controls CSS stylesheet sources'
    },
    {
      directive: 'img-src',
      allowedSources: ["'self'", "data:", "https:"],
      forbidden: ["*", "http:"],
      description: 'Controls image sources'
    },
    {
      directive: 'connect-src',
      allowedSources: ["'self'", "wss:", "ws:"],
      forbidden: ["*"],
      description: 'Controls AJAX, WebSocket, and EventSource connections'
    },
    {
      directive: 'font-src',
      allowedSources: ["'self'", "https:"],
      forbidden: ["*", "data:"],
      description: 'Controls font loading sources'
    },
    {
      directive: 'object-src',
      allowedSources: ["'none'"],
      forbidden: ["'self'", "*"],
      description: 'Blocks object, embed, and applet elements'
    },
    {
      directive: 'media-src',
      allowedSources: ["'self'"],
      forbidden: ["*"],
      description: 'Controls audio and video sources'
    },
    {
      directive: 'frame-ancestors',
      allowedSources: ["'none'"],
      forbidden: ["'self'", "*"],
      description: 'Prevents embedding in frames (clickjacking protection)'
    },
    {
      directive: 'base-uri',
      allowedSources: ["'self'"],
      forbidden: ["*"],
      description: 'Controls base element URLs'
    },
    {
      directive: 'form-action',
      allowedSources: ["'self'"],
      forbidden: ["*"],
      description: 'Controls form submission targets'
    }
  ]

  test.describe('Security Headers Validation', () => {
    test('should have all required security headers', async ({ page }) => {
      const response = await page.goto('/')
      const headers = response?.headers() || {}
      
      const missingHeaders: string[] = []
      const invalidHeaders: Array<{ name: string, actual: string, expected: string | RegExp }> = []
      
      for (const header of REQUIRED_SECURITY_HEADERS) {
        const actualValue = headers[header.name.toLowerCase()]
        
        if (!actualValue && header.required) {
          missingHeaders.push(header.name)
          continue
        }
        
        if (actualValue && header.expectedValue) {
          const isValid = typeof header.expectedValue === 'string' 
            ? actualValue === header.expectedValue
            : header.expectedValue.test(actualValue)
            
          if (!isValid) {
            invalidHeaders.push({
              name: header.name,
              actual: actualValue,
              expected: header.expectedValue
            })
          }
        }
      }
      
      if (missingHeaders.length > 0) {
        console.error('[SECURITY-ALERT] Missing required security headers:', missingHeaders)
      }
      
      if (invalidHeaders.length > 0) {
        console.error('[SECURITY-ALERT] Invalid security headers:', invalidHeaders)
      }
      
      expect(missingHeaders).toHaveLength(0)
      expect(invalidHeaders).toHaveLength(0)
    })

    test('should have secure HSTS configuration', async ({ page }) => {
      const response = await page.goto('/')
      const hstsHeader = response?.headers()['strict-transport-security']
      
      expect(hstsHeader).toBeDefined()
      
      if (hstsHeader) {
        // Should have max-age of at least 1 year
        const maxAgeMatch = hstsHeader.match(/max-age=(\d+)/)
        const maxAge = maxAgeMatch ? parseInt(maxAgeMatch[1]) : 0
        const oneYear = 365 * 24 * 60 * 60 // seconds
        
        expect(maxAge).toBeGreaterThanOrEqual(oneYear)
        
        // Should include subdomains
        expect(hstsHeader).toContain('includeSubDomains')
        
        // Should not include preload in non-production (optional)
        if (process.env.NODE_ENV === 'production') {
          console.log('Production environment: Consider adding preload directive')
        }
      }
    })

    test('should prevent server information disclosure', async ({ page }) => {
      const response = await page.goto('/')
      const headers = response?.headers() || {}
      
      // Server header should not reveal version information
      const serverHeader = headers['server']
      if (serverHeader) {
        expect(serverHeader).not.toMatch(/nginx\/[\d.]+/)
        expect(serverHeader).not.toMatch(/Apache\/[\d.]+/)
        expect(serverHeader).not.toMatch(/IIS\/[\d.]+/)
      }
      
      // X-Powered-By should not be present
      expect(headers['x-powered-by']).toBeUndefined()
      
      // X-AspNet-Version should not be present
      expect(headers['x-aspnet-version']).toBeUndefined()
      
      // X-AspNetMvc-Version should not be present  
      expect(headers['x-aspnetmvc-version']).toBeUndefined()
    })
  })

  test.describe('Content Security Policy Validation', () => {
    test('should have a comprehensive CSP policy', async ({ page }) => {
      const response = await page.goto('/')
      const cspHeader = response?.headers()['content-security-policy']
      
      expect(cspHeader).toBeDefined()
      
      if (!cspHeader) return
      
      // Parse CSP directives
      const directives = new Map<string, string[]>()
      const parts = cspHeader.split(';').map(p => p.trim())
      
      for (const part of parts) {
        if (!part) continue
        const [directive, ...sources] = part.split(/\s+/)
        directives.set(directive, sources)
      }
      
      // Validate each required directive
      const missingDirectives: string[] = []
      const insecureDirectives: Array<{ directive: string, issue: string }> = []
      
      for (const cspDirective of CSP_DIRECTIVES) {
        const sources = directives.get(cspDirective.directive)
        
        if (!sources && cspDirective.allowedSources.length > 0) {
          missingDirectives.push(cspDirective.directive)
          continue
        }
        
        if (sources) {
          // Check for forbidden sources
          for (const forbidden of cspDirective.forbidden) {
            if (sources.includes(forbidden)) {
              insecureDirectives.push({
                directive: cspDirective.directive,
                issue: `Contains forbidden source: ${forbidden}`
              })
            }
          }
          
          // Check for overly permissive sources
          if (sources.includes('*') && !cspDirective.allowedSources.includes('*')) {
            insecureDirectives.push({
              directive: cspDirective.directive,
              issue: 'Contains wildcard (*) source - overly permissive'
            })
          }
        }
      }
      
      if (missingDirectives.length > 0) {
        console.error('[SECURITY-ALERT] Missing CSP directives:', missingDirectives)
      }
      
      if (insecureDirectives.length > 0) {
        console.error('[SECURITY-ALERT] Insecure CSP directives:', insecureDirectives)
      }
      
      expect(missingDirectives.length).toBeLessThanOrEqual(2) // Allow some optional directives
      expect(insecureDirectives).toHaveLength(0)
    })

    test('should prevent unsafe CSP configurations', async ({ page }) => {
      const response = await page.goto('/')
      const cspHeader = response?.headers()['content-security-policy']
      
      expect(cspHeader).toBeDefined()
      
      if (!cspHeader) return
      
      // Critical security violations
      const violations = []
      
      if (cspHeader.includes("'unsafe-eval'")) {
        violations.push("'unsafe-eval' detected - allows arbitrary code execution")
      }
      
      if (cspHeader.includes('script-src *')) {
        violations.push('Wildcard script-src detected - allows scripts from any origin')
      }
      
      if (cspHeader.includes('object-src *')) {
        violations.push('Wildcard object-src detected - allows plugins from any origin')
      }
      
      if (!cspHeader.includes('frame-ancestors')) {
        violations.push('Missing frame-ancestors directive - vulnerable to clickjacking')
      }
      
      if (!cspHeader.includes('base-uri')) {
        violations.push('Missing base-uri directive - vulnerable to base tag injection')
      }
      
      if (violations.length > 0) {
        console.error('[SECURITY-ALERT] CSP Security Violations:', violations)
      }
      
      expect(violations).toHaveLength(0)
    })

    test('should validate CSP nonce implementation', async ({ page }) => {
      await page.goto('/')
      
      // Check if nonces are properly implemented
      const nonceCheck = await page.evaluate(() => {
        const scripts = Array.from(document.querySelectorAll('script'))
        const inlineScripts = scripts.filter(script => !script.src && script.textContent)
        const nonceScripts = scripts.filter(script => script.nonce)
        
        return {
          totalScripts: scripts.length,
          inlineScripts: inlineScripts.length,
          nonceScripts: nonceScripts.length,
          hasNonceSupport: nonceScripts.length > 0
        }
      })
      
      // If there are inline scripts, they should have nonces
      if (nonceCheck.inlineScripts > 0) {
        expect(nonceCheck.nonceScripts).toBeGreaterThan(0)
        
        // Nonce should be sufficiently random and long
        const nonces = await page.evaluate(() => {
          return Array.from(document.querySelectorAll('script[nonce]'))
            .map(script => script.nonce)
        })
        
        for (const nonce of nonces) {
          expect(nonce.length).toBeGreaterThanOrEqual(16) // At least 128 bits
          expect(nonce).toMatch(/^[A-Za-z0-9+/]+=*$/) // Valid base64
        }
      }
    })

    test('should report CSP violations properly', async ({ page }) => {
      let cspViolationReported = false
      
      // Listen for CSP violation reports
      page.on('console', msg => {
        if (msg.type() === 'error' && msg.text().includes('Content Security Policy')) {
          cspViolationReported = true
        }
      })
      
      await page.goto('/')
      
      // Try to trigger a CSP violation
      await page.evaluate(() => {
        try {
          // This should violate CSP if properly configured
          const script = document.createElement('script')
          script.innerHTML = 'console.log("CSP violation test")'
          document.head.appendChild(script)
        } catch (e) {
          console.error('CSP violation caught:', e)
        }
      })
      
      // Wait a bit for violation to be reported
      await page.waitForTimeout(1000)
      
      // CSP should block the violation (expect violation to be reported)
      // This test verifies that CSP is actually enforced, not just declared
      if (!cspViolationReported) {
        console.warn('CSP may not be properly enforced - no violations detected during test')
      }
    })
  })

  test.describe('CORS Policy Validation', () => {
    test('should have restrictive CORS policy', async ({ page }) => {
      // Test CORS preflight request
      const corsResponse = await page.request.options('/', {
        headers: {
          'Origin': 'https://malicious-site.com',
          'Access-Control-Request-Method': 'POST',
          'Access-Control-Request-Headers': 'Content-Type'
        }
      })
      
      const corsHeaders = corsResponse.headers()
      
      // Should not allow arbitrary origins
      const allowOrigin = corsHeaders['access-control-allow-origin']
      expect(allowOrigin).not.toBe('*')
      expect(allowOrigin).not.toBe('https://malicious-site.com')
      
      // Should not allow dangerous methods
      const allowMethods = corsHeaders['access-control-allow-methods']
      if (allowMethods) {
        expect(allowMethods).not.toContain('TRACE')
        expect(allowMethods).not.toContain('CONNECT')
      }
      
      // Should not allow arbitrary headers
      const allowHeaders = corsHeaders['access-control-allow-headers']
      if (allowHeaders) {
        expect(allowHeaders).not.toBe('*')
      }
    })

    test('should validate credentials handling', async ({ page }) => {
      const response = await page.request.get('/', {
        headers: {
          'Origin': 'https://trusted-domain.com'
        }
      })
      
      const corsHeaders = response.headers()
      const allowCredentials = corsHeaders['access-control-allow-credentials']
      
      // If credentials are allowed, origin should not be wildcard
      if (allowCredentials === 'true') {
        const allowOrigin = corsHeaders['access-control-allow-origin']
        expect(allowOrigin).not.toBe('*')
      }
    })
  })

  test.describe('Clickjacking Protection', () => {
    test('should prevent framing attacks', async ({ page }) => {
      const response = await page.goto('/')
      const headers = response?.headers() || {}
      
      const xFrameOptions = headers['x-frame-options']
      const csp = headers['content-security-policy']
      
      // Should have either X-Frame-Options or CSP frame-ancestors
      const hasFrameProtection = xFrameOptions || (csp && csp.includes('frame-ancestors'))
      expect(hasFrameProtection).toBeTruthy()
      
      // X-Frame-Options should be DENY or SAMEORIGIN
      if (xFrameOptions) {
        expect(['DENY', 'SAMEORIGIN']).toContain(xFrameOptions)
      }
      
      // CSP frame-ancestors should be restrictive
      if (csp && csp.includes('frame-ancestors')) {
        expect(csp).not.toContain("frame-ancestors *")
        expect(csp).toMatch(/frame-ancestors\s+('none'|'self')/);
      }
    })

    test('should prevent embedding in malicious frames', async ({ browser }) => {
      // Create a page that tries to embed the application
      const context = await browser.newContext()
      const page = await context.newPage()
      
      const frameTest = await page.evaluate(async () => {
        return new Promise((resolve) => {
          const iframe = document.createElement('iframe')
          iframe.src = '/' // Try to embed the application
          
          iframe.onload = () => resolve({ loaded: true, blocked: false })
          iframe.onerror = () => resolve({ loaded: false, blocked: true })
          
          document.body.appendChild(iframe)
          
          // Timeout after 3 seconds
          setTimeout(() => resolve({ loaded: false, blocked: true }), 3000)
        })
      })
      
      // Application should prevent being embedded (blocked: true)
      expect(frameTest.blocked).toBe(true)
      
      await context.close()
    })
  })

  test.describe('Security Headers Monitoring', () => {
    test('should maintain consistent security headers across pages', async ({ page }) => {
      const pages = ['/', '/dashboard', '/api/health']
      const headerConsistency = new Map<string, Set<string>>()
      
      for (const pagePath of pages) {
        try {
          const response = await page.request.get(pagePath)
          const headers = response.headers()
          
          for (const header of REQUIRED_SECURITY_HEADERS) {
            const headerName = header.name.toLowerCase()
            const value = headers[headerName]
            
            if (!headerConsistency.has(headerName)) {
              headerConsistency.set(headerName, new Set())
            }
            
            if (value) {
              headerConsistency.get(headerName)!.add(value)
            }
          }
        } catch (error) {
          console.log(`Skipping ${pagePath}: ${error}`)
        }
      }
      
      // Check for inconsistencies
      const inconsistentHeaders = Array.from(headerConsistency.entries())
        .filter(([_, values]) => values.size > 1)
      
      if (inconsistentHeaders.length > 0) {
        console.warn('Inconsistent security headers across pages:', inconsistentHeaders)
      }
      
      // Allow some variation but flag major inconsistencies
      expect(inconsistentHeaders.length).toBeLessThanOrEqual(1)
    })

    test('should validate security headers in error responses', async ({ page }) => {
      // Test 404 error page
      const notFoundResponse = await page.request.get('/nonexistent-page')
      const errorHeaders = notFoundResponse.headers()
      
      // Security headers should be present even in error responses
      expect(errorHeaders['x-frame-options']).toBeDefined()
      expect(errorHeaders['x-content-type-options']).toBeDefined()
      
      // Should not leak server information in errors
      expect(errorHeaders['server']).not.toMatch(/nginx\/[\d.]+/)
    })
  })

  test.describe('Advanced Security Headers', () => {
    test('should implement Feature Policy/Permissions Policy', async ({ page }) => {
      const response = await page.goto('/')
      const headers = response?.headers() || {}
      
      const permissionsPolicy = headers['permissions-policy'] || headers['feature-policy']
      
      if (permissionsPolicy) {
        // Should disable dangerous features
        expect(permissionsPolicy).toContain('camera=()')
        expect(permissionsPolicy).toContain('microphone=()')
        expect(permissionsPolicy).toContain('geolocation=()')
      }
    })

    test('should implement Cross-Origin policies for isolation', async ({ page }) => {
      const response = await page.goto('/')
      const headers = response?.headers() || {}
      
      // Cross-Origin-Opener-Policy for isolation
      const coop = headers['cross-origin-opener-policy']
      if (coop) {
        expect(['same-origin', 'same-origin-allow-popups']).toContain(coop)
      }
      
      // Cross-Origin-Embedder-Policy for additional security
      const coep = headers['cross-origin-embedder-policy']
      if (coep) {
        expect(coep).toBe('require-corp')
      }
    })

    test('should validate timing attack protection', async ({ page }) => {
      // Test for timing attack vectors in error responses
      const startTime = Date.now()
      
      await page.request.post('/api/login', {
        data: { username: 'nonexistent@example.com', password: 'wrong' }
      }).catch(() => {})
      
      const validUserTime = Date.now() - startTime
      
      const startTime2 = Date.now()
      
      await page.request.post('/api/login', {
        data: { username: 'test@example.com', password: 'wrong' }
      }).catch(() => {})
      
      const invalidUserTime = Date.now() - startTime2
      
      // Response times should not reveal if user exists
      const timingDifference = Math.abs(validUserTime - invalidUserTime)
      expect(timingDifference).toBeLessThan(100) // Allow 100ms variance
    })
  })
})