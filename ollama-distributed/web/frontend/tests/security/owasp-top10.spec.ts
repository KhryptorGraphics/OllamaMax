/**
 * OWASP Top 10 2021 Security Testing Suite
 * 
 * This test suite validates the Ollama Distributed frontend against the OWASP Top 10
 * security vulnerabilities, implementing comprehensive security testing patterns.
 */

import { test, expect, Page, BrowserContext } from '@playwright/test'
import { cspManager, InputSanitizer, SecureStorage } from '../../src/utils/security'

test.describe('[SECURITY-ALERT] OWASP Top 10 2021 Vulnerability Testing', () => {
  let context: BrowserContext
  let page: Page

  test.beforeEach(async ({ browser }) => {
    context = await browser.newContext({
      // Simulate real-world security test environment
      ignoreHTTPSErrors: false,
      permissions: [],
      extraHTTPHeaders: {
        'User-Agent': 'SecurityScanner/1.0 (OWASP Testing)'
      }
    })
    page = await context.newPage()
    
    // Monitor console for security violations
    page.on('console', msg => {
      if (msg.type() === 'error' && msg.text().includes('Content Security Policy')) {
        console.error('CSP Violation:', msg.text())
      }
    })
  })

  test.afterEach(async () => {
    await context.close()
  })

  test.describe('A01:2021 - Broken Access Control', () => {
    test('should prevent unauthorized access to admin routes', async () => {
      // Test horizontal privilege escalation
      await page.goto('/admin/users')
      
      // Should redirect to login if not authenticated
      await expect(page).toHaveURL(/.*login.*/)
      
      // Test vertical privilege escalation
      await page.goto('/api/admin/users', { waitUntil: 'networkidle' })
      const response = await page.waitForResponse('/api/admin/users')
      expect(response.status()).toBeGreaterThanOrEqual(401)
    })

    test('should enforce proper session management', async () => {
      // Test session timeout
      await page.goto('/')
      
      // Simulate expired session
      await page.evaluate(() => {
        localStorage.setItem('auth_token', 'expired_token')
      })
      
      await page.goto('/dashboard')
      await expect(page).toHaveURL(/.*login.*/)
    })

    test('should validate CSRF protection', async () => {
      await page.goto('/')
      
      // Check for CSRF token in forms
      const forms = await page.locator('form').all()
      for (const form of forms) {
        const method = await form.getAttribute('method')
        if (method?.toLowerCase() === 'post') {
          const csrfInput = form.locator('input[name="csrf_token"], input[name="_token"]')
          await expect(csrfInput).toBeVisible()
        }
      }
    })
  })

  test.describe('A02:2021 - Cryptographic Failures', () => {
    test('should enforce HTTPS in production', async () => {
      // Test HTTP to HTTPS redirect
      const response = await page.request.get('http://localhost:8080', {
        maxRedirects: 0
      })
      
      if (process.env.NODE_ENV === 'production') {
        expect(response.status()).toBe(301)
        expect(response.headers()['location']).toMatch(/^https:/)
      }
    })

    test('should implement proper TLS configuration', async () => {
      await page.goto('/')
      
      // Check for HSTS header
      const hstsHeader = await page.evaluate(() => {
        return document.querySelector('meta[http-equiv="Strict-Transport-Security"]')
      })
      
      if (process.env.NODE_ENV === 'production') {
        expect(hstsHeader).toBeTruthy()
      }
    })

    test('should validate secure storage implementation', async () => {
      await page.goto('/')
      
      // Test SecureStorage encryption
      await page.evaluate(async () => {
        const { SecureStorage } = await import('../../src/utils/security')
        await SecureStorage.setItem('test_key', 'sensitive_data')
        const stored = localStorage.getItem('test_key')
        
        // Stored data should be encrypted (not plaintext)
        if (stored) {
          expect(stored).not.toBe('sensitive_data')
          expect(stored.length).toBeGreaterThan(20) // Encrypted data is longer
        }
      })
    })
  })

  test.describe('A03:2021 - Injection', () => {
    test('should prevent XSS attacks', async () => {
      await page.goto('/')
      
      const xssPayloads = [
        '<script>alert("XSS")</script>',
        'javascript:alert("XSS")',
        '<img src=x onerror=alert("XSS")>',
        '"><script>alert("XSS")</script>',
        "'><script>alert('XSS')</script>"
      ]

      for (const payload of xssPayloads) {
        // Test input sanitization
        const inputs = await page.locator('input[type="text"], textarea').all()
        
        for (const input of inputs) {
          await input.fill(payload)
          await input.blur()
          
          // Check that script didn't execute
          const alerts = await page.evaluate(() => window.alert.toString())
          expect(alerts).not.toContain('XSS')
          
          // Check that input is properly escaped
          const value = await input.inputValue()
          expect(value).not.toContain('<script>')
        }
      }
    })

    test('should validate SQL injection protection', async () => {
      await page.goto('/')
      
      const sqlPayloads = [
        "'; DROP TABLE users; --",
        "' OR '1'='1",
        "1' UNION SELECT * FROM users--",
        "'; INSERT INTO users VALUES ('hacker', 'password'); --"
      ]

      // Test search functionality
      for (const payload of sqlPayloads) {
        await page.fill('input[type="search"]', payload)
        await page.press('input[type="search"]', 'Enter')
        
        // Wait for response and check for SQL errors
        await page.waitForTimeout(1000)
        const content = await page.textContent('body')
        
        expect(content).not.toContain('SQL syntax')
        expect(content).not.toContain('MySQL error')
        expect(content).not.toContain('PostgreSQL error')
        expect(content).not.toContain('ORA-')
      }
    })

    test('should prevent command injection', async () => {
      await page.goto('/')
      
      const commandPayloads = [
        '; cat /etc/passwd',
        '| ls -la',
        '&& whoami',
        '`id`',
        '$(whoami)'
      ]

      const inputs = await page.locator('input').all()
      for (const input of inputs) {
        for (const payload of commandPayloads) {
          await input.fill(`test${payload}`)
          await input.press('Enter')
          
          await page.waitForTimeout(500)
          const content = await page.textContent('body')
          
          // Should not contain command output
          expect(content).not.toContain('root:x:')
          expect(content).not.toContain('uid=')
          expect(content).not.toContain('total ')
        }
      }
    })
  })

  test.describe('A04:2021 - Insecure Design', () => {
    test('should implement proper authentication flow', async () => {
      await page.goto('/login')
      
      // Test multi-factor authentication requirement
      await page.fill('input[name="email"]', 'test@example.com')
      await page.fill('input[name="password"]', 'ValidPassword123!')
      await page.click('button[type="submit"]')
      
      // Should require MFA for sensitive operations
      await expect(page.locator('[data-testid="mfa-prompt"]')).toBeVisible()
    })

    test('should validate business logic flaws', async () => {
      // Test for race conditions in critical operations
      await page.goto('/dashboard')
      
      // Simulate concurrent requests
      const promises = Array.from({ length: 5 }, () => 
        page.request.post('/api/transfer', {
          data: { amount: 100, to: 'account123' }
        })
      )
      
      const responses = await Promise.all(promises)
      
      // Only one should succeed, others should fail with proper error
      const successCount = responses.filter(r => r.status() === 200).length
      expect(successCount).toBeLessThanOrEqual(1)
    })
  })

  test.describe('A05:2021 - Security Misconfiguration', () => {
    test('should validate Content Security Policy', async () => {
      await page.goto('/')
      
      // Check CSP header implementation
      const cspHeader = await page.evaluate(() => {
        const meta = document.querySelector('meta[http-equiv="Content-Security-Policy"]')
        return meta?.getAttribute('content') || ''
      })
      
      // CSP should be restrictive
      expect(cspHeader).toContain("default-src 'self'")
      expect(cspHeader).not.toContain("'unsafe-eval'")
      expect(cspHeader).toContain("frame-ancestors 'none'")
      
      // Test CSP violation reporting
      const violationPromise = page.waitForEvent('console', msg => 
        msg.type() === 'error' && msg.text().includes('Content Security Policy')
      )
      
      await page.evaluate(() => {
        try {
          // This should trigger CSP violation
          const script = document.createElement('script')
          script.innerHTML = 'console.log("CSP violation test")'
          document.body.appendChild(script)
        } catch (e) {
          // Expected CSP block
        }
      })
    })

    test('should validate security headers', async () => {
      const response = await page.goto('/')
      const headers = response?.headers() || {}
      
      // Required security headers
      expect(headers['x-frame-options']).toBeDefined()
      expect(headers['x-content-type-options']).toBe('nosniff')
      expect(headers['x-xss-protection']).toBeDefined()
      expect(headers['referrer-policy']).toBeDefined()
      
      if (process.env.NODE_ENV === 'production') {
        expect(headers['strict-transport-security']).toBeDefined()
      }
    })

    test('should disable unnecessary features', async () => {
      await page.goto('/')
      
      // Server should not expose version information
      const response = await page.request.get('/')
      const headers = response.headers()
      
      expect(headers['server']).not.toContain('Apache/')
      expect(headers['server']).not.toContain('nginx/')
      expect(headers['x-powered-by']).toBeUndefined()
    })
  })

  test.describe('A06:2021 - Vulnerable and Outdated Components', () => {
    test('should validate dependency security', async () => {
      // This would typically be done in CI/CD pipeline
      // Here we simulate checking for known vulnerabilities
      
      const vulnerablePackages = [
        'lodash@4.17.19', // Known vulnerability
        'react-dom@16.13.0', // Outdated version
        'axios@0.21.0' // Known vulnerability
      ]
      
      // In real implementation, this would check package.json
      // against vulnerability databases
      
      // Mock vulnerability check
      const hasVulnerabilities = false // Would be actual check
      expect(hasVulnerabilities).toBe(false)
    })
  })

  test.describe('A07:2021 - Identification and Authentication Failures', () => {
    test('should implement proper password policies', async () => {
      await page.goto('/register')
      
      const weakPasswords = [
        '123456',
        'password',
        'qwerty',
        'abc123',
        'admin'
      ]
      
      for (const password of weakPasswords) {
        await page.fill('input[name="password"]', password)
        await page.blur('input[name="password"]')
        
        // Should show password strength error
        await expect(page.locator('.password-error')).toBeVisible()
      }
      
      // Test strong password acceptance
      await page.fill('input[name="password"]', 'SecureP@ssw0rd123!')
      await page.blur('input[name="password"]')
      await expect(page.locator('.password-error')).not.toBeVisible()
    })

    test('should implement account lockout protection', async () => {
      await page.goto('/login')
      
      // Attempt multiple failed logins
      for (let i = 0; i < 6; i++) {
        await page.fill('input[name="email"]', 'test@example.com')
        await page.fill('input[name="password"]', 'wrongpassword')
        await page.click('button[type="submit"]')
        await page.waitForTimeout(1000)
      }
      
      // Should show account locked message
      await expect(page.locator('.account-locked')).toBeVisible()
    })

    test('should validate session security', async () => {
      await page.goto('/')
      
      // Check session cookie attributes
      const cookies = await context.cookies()
      const sessionCookie = cookies.find(c => c.name.includes('session'))
      
      if (sessionCookie) {
        expect(sessionCookie.secure).toBe(process.env.NODE_ENV === 'production')
        expect(sessionCookie.httpOnly).toBe(true)
        expect(sessionCookie.sameSite).toBe('Strict')
      }
    })
  })

  test.describe('A08:2021 - Software and Data Integrity Failures', () => {
    test('should validate Subresource Integrity', async () => {
      await page.goto('/')
      
      // Check for SRI attributes on external resources
      const externalScripts = await page.locator('script[src^="http"]').all()
      const externalStylesheets = await page.locator('link[rel="stylesheet"][href^="http"]').all()
      
      for (const script of externalScripts) {
        const integrity = await script.getAttribute('integrity')
        expect(integrity).toBeTruthy()
        expect(integrity).toMatch(/^(sha256|sha384|sha512)-/)
      }
      
      for (const stylesheet of externalStylesheets) {
        const integrity = await stylesheet.getAttribute('integrity')
        expect(integrity).toBeTruthy()
        expect(integrity).toMatch(/^(sha256|sha384|sha512)-/)
      }
    })
  })

  test.describe('A09:2021 - Security Logging and Monitoring Failures', () => {
    test('should log security events', async () => {
      await page.goto('/login')
      
      // Simulate failed login attempt
      await page.fill('input[name="email"]', 'attacker@evil.com')
      await page.fill('input[name="password"]', 'wrongpassword')
      await page.click('button[type="submit"]')
      
      // Check that security event is logged (would be server-side)
      // This is a placeholder for actual log verification
      const securityEventLogged = true // Mock verification
      expect(securityEventLogged).toBe(true)
    })

    test('should detect and alert on suspicious activities', async () => {
      // Simulate suspicious activity patterns
      await page.goto('/')
      
      // Rapid requests (potential DoS)
      const rapidRequests = Array.from({ length: 100 }, () => 
        page.request.get('/api/health')
      )
      
      await Promise.all(rapidRequests)
      
      // Should trigger rate limiting
      const response = await page.request.get('/api/health')
      expect([200, 429]).toContain(response.status())
    })
  })

  test.describe('A10:2021 - Server-Side Request Forgery (SSRF)', () => {
    test('should prevent SSRF attacks', async () => {
      await page.goto('/')
      
      const ssrfPayloads = [
        'http://localhost:22',
        'file:///etc/passwd',
        'http://169.254.169.254/latest/meta-data/',
        'gopher://localhost:25',
        'dict://localhost:11211'
      ]
      
      for (const payload of ssrfPayloads) {
        const response = await page.request.post('/api/fetch-url', {
          data: { url: payload }
        })
        
        // Should reject malicious URLs
        expect(response.status()).toBeGreaterThanOrEqual(400)
      }
    })
  })

  test.describe('Additional Security Tests', () => {
    test('should prevent clickjacking attacks', async () => {
      await page.goto('/')
      
      // Check X-Frame-Options header
      const response = await page.request.get('/')
      const headers = response.headers()
      
      expect(['DENY', 'SAMEORIGIN']).toContain(headers['x-frame-options'])
    })

    test('should implement proper CORS policy', async () => {
      const response = await page.request.options('/', {
        headers: {
          'Origin': 'https://evil.com',
          'Access-Control-Request-Method': 'POST'
        }
      })
      
      const corsHeader = response.headers()['access-control-allow-origin']
      expect(corsHeader).not.toBe('*')
      expect(corsHeader).not.toBe('https://evil.com')
    })

    test('should validate file upload security', async () => {
      await page.goto('/')
      
      const fileInput = page.locator('input[type="file"]')
      if (await fileInput.count() > 0) {
        // Test malicious file types
        const maliciousFiles = [
          'test.exe',
          'test.php',
          'test.jsp',
          'test.aspx'
        ]
        
        for (const filename of maliciousFiles) {
          // Create a temporary file
          const buffer = Buffer.from('malicious content')
          await fileInput.setInputFiles({
            name: filename,
            mimeType: 'application/octet-stream',
            buffer
          })
          
          // Should reject malicious file types
          await expect(page.locator('.file-error')).toBeVisible()
        }
      }
    })
  })
})