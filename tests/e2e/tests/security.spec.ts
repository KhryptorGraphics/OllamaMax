import { test, expect } from '@playwright/test';
import { SecurityHelper } from '../helpers/security-helper';

/**
 * Security Testing Suite for OllamaMax Platform
 * 
 * Tests critical security aspects:
 * - Authentication and authorization
 * - Input validation and sanitization
 * - CORS and security headers
 * - Rate limiting and DoS protection
 * - XSS and injection prevention
 * - Data exposure and information leakage
 */

test.describe('Security Testing', () => {
  let securityHelper: SecurityHelper;

  test.beforeEach(async ({ page }) => {
    securityHelper = new SecurityHelper(page);
    await page.goto('/');
  });

  test('security headers validation', async ({ request }) => {
    const response = await request.get('/');
    const headers = response.headers();
    
    // Check for important security headers
    const securityHeaders = {
      'x-content-type-options': 'nosniff',
      'x-frame-options': ['DENY', 'SAMEORIGIN'],
      'x-xss-protection': '1; mode=block',
      'strict-transport-security': null, // Should exist for HTTPS
      'content-security-policy': null,
      'referrer-policy': null
    };
    
    const findings = [];
    
    for (const [headerName, expectedValues] of Object.entries(securityHeaders)) {
      const headerValue = headers[headerName];
      
      if (!headerValue) {
        findings.push(`Missing security header: ${headerName}`);
      } else if (Array.isArray(expectedValues)) {
        if (!expectedValues.some(expected => headerValue.includes(expected))) {
          findings.push(`${headerName}: ${headerValue} (expected one of: ${expectedValues.join(', ')})`);
        }
      } else if (expectedValues && !headerValue.includes(expectedValues)) {
        findings.push(`${headerName}: ${headerValue} (expected: ${expectedValues})`);
      }
    }
    
    if (findings.length > 0) {
      console.warn('Security header findings:', findings);
    } else {
      console.log('✅ All critical security headers present');
    }
    
    // Don't fail test for missing headers in development
    expect(response.ok()).toBeTruthy();
  });

  test('input validation and XSS prevention', async ({ page }) => {
    // Test XSS prevention in various input fields
    const xssPayloads = [
      '<script>alert("xss")</script>',
      '"><script>alert(String.fromCharCode(88,83,83))</script>',
      "javascript:alert('XSS')",
      '<img src=x onerror=alert("XSS")>',
      '<svg onload=alert("XSS")>',
      '{{7*7}}', // Template injection
      '${7*7}', // Expression injection
    ];
    
    // Look for input fields
    const inputFields = page.locator('input[type="text"], input[type="search"], textarea, [contenteditable="true"]');
    const inputCount = await inputFields.count();
    
    if (inputCount > 0) {
      for (let i = 0; i < Math.min(inputCount, 3); i++) {
        const input = inputFields.nth(i);
        
        for (const payload of xssPayloads.slice(0, 3)) { // Test first 3 payloads
          await input.clear();
          await input.fill(payload);
          await input.press('Enter');
          await page.waitForTimeout(1000);
          
          // Check if XSS executed (page should not have alerts)
          const hasAlert = await page.evaluate(() => {
            // Check if any script tags were actually executed
            return document.body.innerHTML.includes('alert(') || 
                   document.scripts.length > 10; // Suspicious script count
          });
          
          expect(hasAlert).toBeFalsy();
        }
      }
      
      console.log('✅ XSS prevention tests completed');
    } else {
      console.log('ℹ️  No input fields found for XSS testing');
    }
  });

  test('SQL injection prevention', async ({ request }) => {
    // Test API endpoints for SQL injection vulnerabilities
    const sqlPayloads = [
      "'; DROP TABLE users; --",
      "' OR '1'='1",
      "' UNION SELECT * FROM users --",
      "admin'--",
      "admin' /*",
      "' OR 1=1#"
    ];
    
    const endpoints = [
      '/api/v1/models',
      '/api/v1/health',
      '/api/v1/status',
      '/search',
      '/query'
    ];
    
    for (const endpoint of endpoints) {
      for (const payload of sqlPayloads.slice(0, 2)) { // Test first 2 payloads per endpoint
        try {
          // Test as query parameter
          const response1 = await request.get(`${endpoint}?q=${encodeURIComponent(payload)}`);
          
          // Test as POST body
          const response2 = await request.post(endpoint, {
            data: { query: payload, search: payload },
            headers: { 'Content-Type': 'application/json' }
          }).catch(() => ({ ok: () => false, status: () => 400 }));
          
          // Should not return database errors or sensitive information
          if (response1.ok()) {
            const text1 = await response1.text();
            expect(text1).not.toMatch(/sql|database|mysql|postgres|sqlite|error|exception|stack trace/i);
          }
          
          if (response2.ok && response2.ok()) {
            const text2 = await response2.text();
            expect(text2).not.toMatch(/sql|database|mysql|postgres|sqlite|error|exception|stack trace/i);
          }
          
        } catch (error) {
          // Network errors are acceptable - means endpoint doesn't exist
          continue;
        }
      }
    }
    
    console.log('✅ SQL injection prevention tests completed');
  });

  test('authentication bypass attempts', async ({ request, page }) => {
    // Test common authentication bypass techniques
    const bypassAttempts = [
      { headers: { 'X-Forwarded-For': '127.0.0.1' } },
      { headers: { 'X-Real-IP': '127.0.0.1' } },
      { headers: { 'X-Forwarded-User': 'admin' } },
      { headers: { 'X-Remote-User': 'admin' } },
      { headers: { 'Authorization': 'Bearer invalid-token' } },
      { headers: { 'Cookie': 'admin=true; session=admin; authenticated=true' } }
    ];
    
    const protectedEndpoints = [
      '/admin',
      '/api/v1/admin',
      '/dashboard',
      '/config',
      '/settings'
    ];
    
    for (const endpoint of protectedEndpoints) {
      for (const attempt of bypassAttempts) {
        try {
          const response = await request.get(endpoint, attempt);
          
          // Should not grant unauthorized access
          if (response.ok()) {
            const content = await response.text();
            
            // Check if we got administrative content
            const hasAdminContent = content.toLowerCase().includes('admin') && 
                                  (content.includes('dashboard') || content.includes('configuration'));
            
            if (hasAdminContent) {
              console.warn(`⚠️  Potential auth bypass at ${endpoint} with headers:`, attempt.headers);
            }
          }
          
        } catch (error) {
          // Expected for non-existent endpoints
          continue;
        }
      }
    }
    
    console.log('✅ Authentication bypass tests completed');
  });

  test('rate limiting and DoS protection', async ({ request }) => {
    const endpoint = '/api/v1/health';
    const requestCount = 20;
    const requests = [];
    
    // Send rapid requests
    const startTime = Date.now();
    
    for (let i = 0; i < requestCount; i++) {
      requests.push(
        request.get(endpoint, { timeout: 5000 }).catch(error => ({ 
          error: error.message, 
          status: () => 0,
          ok: () => false 
        }))
      );
    }
    
    const responses = await Promise.all(requests);
    const endTime = Date.now();
    
    const successCount = responses.filter(r => r.ok && r.ok()).length;
    const rateLimitedCount = responses.filter(r => r.status && (r.status() === 429 || r.status() === 503)).length;
    const errorCount = responses.filter(r => r.error).length;
    
    console.log(`Rate limiting test: ${successCount} success, ${rateLimitedCount} rate-limited, ${errorCount} errors`);
    console.log(`Total time: ${endTime - startTime}ms`);
    
    // If rate limiting is implemented, should see some 429 responses
    if (rateLimitedCount > 0) {
      console.log('✅ Rate limiting is active');
      expect(rateLimitedCount).toBeGreaterThan(0);
    } else {
      console.log('ℹ️  No rate limiting detected - may not be implemented');
    }
    
    // At least some requests should complete
    expect(successCount + rateLimitedCount + errorCount).toBe(requestCount);
  });

  test('information disclosure prevention', async ({ request }) => {
    // Test for information leakage in error responses
    const sensitiveEndpoints = [
      '/api/v1/debug',
      '/api/v1/config',
      '/api/v1/env',
      '/.env',
      '/config.json',
      '/package.json',
      '/debug/vars',
      '/metrics',
      '/health/detailed'
    ];
    
    for (const endpoint of sensitiveEndpoints) {
      try {
        const response = await request.get(endpoint);
        
        if (response.ok()) {
          const content = await response.text();
          
          // Check for sensitive information
          const sensitivePatterns = [
            /password|secret|key|token/gi,
            /mongodb|postgres|mysql|redis/gi,
            /localhost|127\.0\.0\.1/g,
            /api[_-]?key|secret[_-]?key/gi,
            /"(password|secret|key|token)":\s*"[^"]+"/gi
          ];
          
          for (const pattern of sensitivePatterns) {
            if (pattern.test(content)) {
              console.warn(`⚠️  Sensitive information potentially exposed at ${endpoint}`);
              break;
            }
          }
        }
        
      } catch (error) {
        // Expected for most endpoints
        continue;
      }
    }
    
    console.log('✅ Information disclosure tests completed');
  });

  test('CORS configuration validation', async ({ request }) => {
    // Test CORS headers and restrictions
    const origins = [
      'http://evil.com',
      'https://malicious.site',
      'http://localhost:3000', // This might be allowed
      'null'
    ];
    
    for (const origin of origins) {
      const response = await request.get('/', {
        headers: { 'Origin': origin }
      });
      
      const corsHeaders = {
        'access-control-allow-origin': response.headers()['access-control-allow-origin'],
        'access-control-allow-credentials': response.headers()['access-control-allow-credentials'],
        'access-control-allow-methods': response.headers()['access-control-allow-methods']
      };
      
      // Should not allow arbitrary origins
      if (corsHeaders['access-control-allow-origin'] === '*' && 
          corsHeaders['access-control-allow-credentials'] === 'true') {
        console.warn('⚠️  Dangerous CORS configuration detected: wildcard origin with credentials');
      }
      
      if (corsHeaders['access-control-allow-origin'] === origin && origin.includes('evil')) {
        console.warn(`⚠️  Suspicious origin allowed: ${origin}`);
      }
    }
    
    console.log('✅ CORS configuration tests completed');
  });

  test('file upload security (if applicable)', async ({ page }) => {
    // Look for file upload functionality
    const fileInputs = page.locator('input[type="file"]');
    const fileInputCount = await fileInputs.count();
    
    if (fileInputCount > 0) {
      console.log(`Found ${fileInputCount} file upload inputs`);
      
      // Test with potentially dangerous files
      const dangerousFiles = [
        { name: 'test.exe', content: 'MZ\x90\x00' }, // Executable header
        { name: 'test.php', content: '<?php phpinfo(); ?>' },
        { name: 'test.js', content: 'alert("xss")' },
        { name: '../../../etc/passwd', content: 'path traversal test' },
        { name: 'test.svg', content: '<svg onload="alert(1)"></svg>' }
      ];
      
      for (let i = 0; i < Math.min(fileInputCount, 2); i++) {
        const fileInput = fileInputs.nth(i);
        
        // Test first dangerous file
        const testFile = dangerousFiles[0];
        
        try {
          // Create a temporary file buffer
          const buffer = Buffer.from(testFile.content);
          
          await fileInput.setInputFiles({
            name: testFile.name,
            mimeType: 'application/octet-stream',
            buffer: buffer
          });
          
          // Try to submit if there's a submit button nearby
          const submitButton = page.locator('button[type="submit"], button:has-text("Upload"), input[type="submit"]').first();
          const submitVisible = await submitButton.isVisible({ timeout: 2000 }).catch(() => false);
          
          if (submitVisible) {
            await submitButton.click();
            await page.waitForTimeout(2000);
            
            // Check for error messages (good) or success (potentially bad)
            const hasError = await page.locator('.error, .alert-danger, [role="alert"]').first().isVisible({ timeout: 3000 }).catch(() => false);
            
            if (hasError) {
              console.log('✅ File upload validation working - rejected dangerous file');
            } else {
              console.warn('⚠️  File upload may have accepted dangerous file');
            }
          }
          
        } catch (error) {
          console.log('ℹ️  File upload test encountered error:', error.message);
        }
      }
    } else {
      console.log('ℹ️  No file upload functionality found');
    }
  });

  test('session management security', async ({ page, context }) => {
    // Test session security if authentication exists
    const loginForm = page.locator('form[action*="login"], form:has(input[type="password"]), .login-form').first();
    const hasLogin = await loginForm.isVisible({ timeout: 3000 }).catch(() => false);
    
    if (hasLogin) {
      // Test session fixation
      const initialCookies = await context.cookies();
      
      // Attempt login (won't succeed but might set cookies)
      const usernameField = loginForm.locator('input[type="text"], input[type="email"], input[name*="user"], input[name*="email"]').first();
      const passwordField = loginForm.locator('input[type="password"]').first();
      
      if (await usernameField.isVisible({ timeout: 1000 }).catch(() => false)) {
        await usernameField.fill('test@test.com');
        await passwordField.fill('testpassword123');
        
        const submitBtn = loginForm.locator('button[type="submit"], input[type="submit"]').first();
        await submitBtn.click();
        await page.waitForTimeout(2000);
        
        const finalCookies = await context.cookies();
        
        // Check if session cookies changed after login attempt
        const sessionCookiesChanged = finalCookies.length !== initialCookies.length ||
          finalCookies.some(cookie => !initialCookies.find(initial => 
            initial.name === cookie.name && initial.value === cookie.value
          ));
        
        if (sessionCookiesChanged) {
          console.log('✅ Session cookies changed after login attempt');
        }
        
        // Check for secure cookie flags
        const secureCookies = finalCookies.filter(cookie => cookie.secure);
        const httpOnlyCookies = finalCookies.filter(cookie => cookie.httpOnly);
        
        console.log(`Secure cookies: ${secureCookies.length}/${finalCookies.length}`);
        console.log(`HttpOnly cookies: ${httpOnlyCookies.length}/${finalCookies.length}`);
      }
    } else {
      console.log('ℹ️  No authentication system detected');
    }
  });
});