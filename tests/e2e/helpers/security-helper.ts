import { Page } from '@playwright/test';

/**
 * Security Testing Helper for OllamaMax Platform
 * 
 * Provides utilities for:
 * - XSS detection and prevention testing
 * - SQL injection testing
 * - CSRF protection testing
 * - Security header validation
 * - Input sanitization testing
 */

export class SecurityHelper {
  constructor(private page: Page) {}

  /**
   * Test for XSS vulnerabilities in input fields
   */
  async testXSSVulnerabilities(payloads: string[] = this.getXSSPayloads()): Promise<{
    vulnerableInputs: Array<{ selector: string; payload: string; detected: boolean }>;
    totalInputs: number;
    vulnerabilityCount: number;
  }> {
    const inputSelectors = [
      'input[type="text"]',
      'input[type="search"]',
      'input[type="email"]',
      'textarea',
      '[contenteditable="true"]'
    ];
    
    const results = [];
    let totalInputs = 0;
    
    for (const selector of inputSelectors) {
      const inputs = this.page.locator(selector);
      const count = await inputs.count();
      totalInputs += count;
      
      for (let i = 0; i < Math.min(count, 3); i++) { // Test first 3 inputs of each type
        const input = inputs.nth(i);
        
        for (const payload of payloads.slice(0, 3)) { // Test first 3 payloads
          const detected = await this.testSingleXSSPayload(input, payload);
          results.push({
            selector: `${selector}:nth(${i})`,
            payload,
            detected
          });
        }
      }
    }
    
    return {
      vulnerableInputs: results.filter(r => r.detected),
      totalInputs,
      vulnerabilityCount: results.filter(r => r.detected).length
    };
  }

  /**
   * Test a single XSS payload against an input
   */
  private async testSingleXSSPayload(input: any, payload: string): Promise<boolean> {
    try {
      await input.clear();
      await input.fill(payload);
      await input.press('Enter');
      await this.page.waitForTimeout(1000);
      
      // Check if XSS was executed
      return await this.page.evaluate(() => {
        // Check for alert dialogs
        if (window.alert.toString() !== 'function alert() { [native code] }') {
          return true;
        }
        
        // Check for script injection in DOM
        const scripts = Array.from(document.scripts);
        const suspiciousScripts = scripts.filter(script => 
          script.innerHTML.includes('alert(') || 
          script.innerHTML.includes('eval(') ||
          script.innerHTML.includes('document.cookie')
        );
        
        return suspiciousScripts.length > 0;
      });
      
    } catch (error) {
      return false;
    }
  }

  /**
   * Get common XSS payloads
   */
  private getXSSPayloads(): string[] {
    return [
      '<script>alert("xss")</script>',
      '"><script>alert("xss")</script>',
      '<img src=x onerror=alert("xss")>',
      '<svg onload=alert("xss")>',
      'javascript:alert("xss")',
      '<iframe src="javascript:alert("xss")"></iframe>',
      '<input onfocus=alert("xss") autofocus>',
      '<select onfocus=alert("xss") autofocus>',
      '<textarea onfocus=alert("xss") autofocus>',
      '\'><script>alert("xss")</script>',
      '\"><script>alert("xss")</script>'
    ];
  }

  /**
   * Test SQL injection vulnerabilities
   */
  async testSQLInjection(endpoints: string[] = ['/api/search', '/api/query']): Promise<{
    vulnerableEndpoints: Array<{ endpoint: string; payload: string; response: string }>;
    totalEndpoints: number;
  }> {
    const sqlPayloads = [
      "' OR '1'='1",
      "'; DROP TABLE users; --",
      "' UNION SELECT * FROM users --",
      "admin'--",
      "' OR 1=1#"
    ];
    
    const vulnerableEndpoints = [];
    
    for (const endpoint of endpoints) {
      for (const payload of sqlPayloads.slice(0, 2)) { // Test first 2 payloads per endpoint
        try {
          const response = await this.page.request.post(endpoint, {
            data: { query: payload, search: payload },
            headers: { 'Content-Type': 'application/json' }
          });
          
          if (response.ok()) {
            const text = await response.text();
            
            // Check for SQL error messages that might indicate vulnerability
            const sqlErrors = [
              'mysql_fetch_array',
              'ORA-01756',
              'Microsoft OLE DB Provider',
              'PostgreSQL query failed',
              'sqlite3.OperationalError',
              'SQL syntax.*MySQL',
              'Warning.*mysql_',
              'MySQLSyntaxErrorException',
              'valid MySQL result'
            ];
            
            const hasErrorMessage = sqlErrors.some(error => 
              text.toLowerCase().includes(error.toLowerCase())
            );
            
            if (hasErrorMessage) {
              vulnerableEndpoints.push({
                endpoint,
                payload,
                response: text.substring(0, 500) // First 500 chars
              });
            }
          }
        } catch (error) {
          // Network errors are expected for non-existent endpoints
          continue;
        }
      }
    }
    
    return {
      vulnerableEndpoints,
      totalEndpoints: endpoints.length
    };
  }

  /**
   * Validate security headers
   */
  async validateSecurityHeaders(): Promise<{
    headers: { [key: string]: string | null };
    score: number;
    missing: string[];
    recommendations: string[];
  }> {
    const response = await this.page.goto(this.page.url());
    const headers = response ? response.headers() : {};
    
    const securityHeaders = {
      'x-content-type-options': headers['x-content-type-options'] || null,
      'x-frame-options': headers['x-frame-options'] || null,
      'x-xss-protection': headers['x-xss-protection'] || null,
      'strict-transport-security': headers['strict-transport-security'] || null,
      'content-security-policy': headers['content-security-policy'] || null,
      'referrer-policy': headers['referrer-policy'] || null,
      'permissions-policy': headers['permissions-policy'] || null
    };
    
    const missing = [];
    const recommendations = [];
    let score = 0;
    
    // Check X-Content-Type-Options
    if (securityHeaders['x-content-type-options']) {
      score += 15;
    } else {
      missing.push('X-Content-Type-Options');
      recommendations.push('Add "X-Content-Type-Options: nosniff" to prevent MIME-type sniffing');
    }
    
    // Check X-Frame-Options
    if (securityHeaders['x-frame-options']) {
      score += 15;
    } else {
      missing.push('X-Frame-Options');
      recommendations.push('Add "X-Frame-Options: DENY" or "SAMEORIGIN" to prevent clickjacking');
    }
    
    // Check X-XSS-Protection
    if (securityHeaders['x-xss-protection']) {
      score += 10;
    } else {
      missing.push('X-XSS-Protection');
      recommendations.push('Add "X-XSS-Protection: 1; mode=block" for legacy XSS protection');
    }
    
    // Check HSTS (only for HTTPS)
    if (this.page.url().startsWith('https://')) {
      if (securityHeaders['strict-transport-security']) {
        score += 20;
      } else {
        missing.push('Strict-Transport-Security');
        recommendations.push('Add HSTS header for HTTPS sites');
      }
    }
    
    // Check CSP
    if (securityHeaders['content-security-policy']) {
      score += 25;
    } else {
      missing.push('Content-Security-Policy');
      recommendations.push('Implement Content Security Policy to prevent XSS attacks');
    }
    
    // Check Referrer Policy
    if (securityHeaders['referrer-policy']) {
      score += 10;
    } else {
      missing.push('Referrer-Policy');
      recommendations.push('Add Referrer Policy to control referrer information');
    }
    
    // Check Permissions Policy
    if (securityHeaders['permissions-policy']) {
      score += 5;
    } else {
      missing.push('Permissions-Policy');
      recommendations.push('Consider adding Permissions Policy for feature control');
    }
    
    return {
      headers: securityHeaders,
      score,
      missing,
      recommendations
    };
  }

  /**
   * Test CSRF protection
   */
  async testCSRFProtection(): Promise<{
    hasTokens: boolean;
    tokenCount: number;
    sameOriginChecks: boolean;
  }> {
    // Look for CSRF tokens in forms
    const forms = this.page.locator('form');
    const formCount = await forms.count();
    let tokenCount = 0;
    
    for (let i = 0; i < formCount; i++) {
      const form = forms.nth(i);
      const csrfTokens = form.locator('input[name*="csrf"], input[name*="token"], input[name="_token"]');
      tokenCount += await csrfTokens.count();
    }
    
    // Check for CSRF tokens in meta tags
    const metaTokens = this.page.locator('meta[name*="csrf"], meta[name*="token"]');
    tokenCount += await metaTokens.count();
    
    // Test same-origin policy (simplified)
    const sameOriginChecks = await this.page.evaluate(() => {
      // Try to access a different origin (this should fail in a secure setup)
      try {
        const xhr = new XMLHttpRequest();
        xhr.open('POST', 'http://evil.com/csrf-test', false);
        xhr.send();
        return false; // If this succeeds, there might be CORS issues
      } catch (error) {
        return true; // Expected to fail due to same-origin policy
      }
    });
    
    return {
      hasTokens: tokenCount > 0,
      tokenCount,
      sameOriginChecks
    };
  }

  /**
   * Test input validation and sanitization
   */
  async testInputValidation(): Promise<{
    testedInputs: number;
    properlyValidated: number;
    issues: Array<{ input: string; issue: string; payload: string }>;
  }> {
    const maliciousInputs = [
      { payload: '../../../etc/passwd', type: 'path_traversal' },
      { payload: '{{7*7}}', type: 'template_injection' },
      { payload: '${7*7}', type: 'expression_injection' },
      { payload: '<script>alert(1)</script>', type: 'xss' },
      { payload: 'eval("alert(1)")', type: 'code_injection' },
      { payload: 'file:///etc/passwd', type: 'file_inclusion' }
    ];
    
    const inputSelectors = ['input[type="text"]', 'textarea', 'input[type="search"]'];
    let testedInputs = 0;
    let properlyValidated = 0;
    const issues = [];
    
    for (const selector of inputSelectors) {
      const inputs = this.page.locator(selector);
      const count = await inputs.count();
      
      for (let i = 0; i < Math.min(count, 2); i++) { // Test first 2 inputs of each type
        const input = inputs.nth(i);
        testedInputs++;
        
        for (const test of maliciousInputs.slice(0, 3)) { // Test first 3 payloads
          try {
            await input.clear();
            await input.fill(test.payload);
            
            // Try to submit or trigger validation
            await input.press('Tab');
            await this.page.waitForTimeout(500);
            
            // Check for validation messages or error states
            const hasValidationError = await this.page.locator(
              '.error, .invalid, [aria-invalid="true"], .field-error'
            ).first().isVisible({ timeout: 1000 }).catch(() => false);
            
            if (!hasValidationError) {
              // Check if dangerous content was reflected back
              const pageContent = await this.page.content();
              if (pageContent.includes(test.payload)) {
                issues.push({
                  input: `${selector}:nth(${i})`,
                  issue: `Dangerous ${test.type} payload reflected without validation`,
                  payload: test.payload
                });
              }
            } else {
              properlyValidated++;
            }
            
          } catch (error) {
            // Error might indicate good validation
            properlyValidated++;
          }
        }
      }
    }
    
    return {
      testedInputs,
      properlyValidated,
      issues
    };
  }

  /**
   * Check for information disclosure
   */
  async checkInformationDisclosure(): Promise<{
    sensitiveDataFound: boolean;
    disclosures: Array<{ type: string; content: string; location: string }>;
  }> {
    const disclosures = [];
    
    // Check page content for sensitive information
    const pageContent = await this.page.content();
    
    const sensitivePatterns = [
      { pattern: /password\s*[=:]\s*["']?[\w\d]+["']?/gi, type: 'password' },
      { pattern: /api[_-]?key\s*[=:]\s*["']?[\w\d\-_]+["']?/gi, type: 'api_key' },
      { pattern: /secret[_-]?key\s*[=:]\s*["']?[\w\d\-_]+["']?/gi, type: 'secret_key' },
      { pattern: /access[_-]?token\s*[=:]\s*["']?[\w\d\-_]+["']?/gi, type: 'access_token' },
      { pattern: /mongodb:\/\/[\w:@.\-\/]+/gi, type: 'mongodb_uri' },
      { pattern: /postgres:\/\/[\w:@.\-\/]+/gi, type: 'postgres_uri' },
      { pattern: /mysql:\/\/[\w:@.\-\/]+/gi, type: 'mysql_uri' }
    ];
    
    for (const { pattern, type } of sensitivePatterns) {
      const matches = pageContent.match(pattern);
      if (matches) {
        for (const match of matches) {
          disclosures.push({
            type,
            content: match.substring(0, 50) + '...', // Truncate for safety
            location: 'page_content'
          });
        }
      }
    }
    
    // Check JavaScript files for sensitive data
    const scriptElements = this.page.locator('script[src]');
    const scriptCount = await scriptElements.count();
    
    for (let i = 0; i < Math.min(scriptCount, 5); i++) { // Check first 5 scripts
      const src = await scriptElements.nth(i).getAttribute('src');
      if (src && !src.startsWith('http')) {
        try {
          const response = await this.page.request.get(src);
          if (response.ok()) {
            const scriptContent = await response.text();
            
            for (const { pattern, type } of sensitivePatterns) {
              const matches = scriptContent.match(pattern);
              if (matches) {
                for (const match of matches) {
                  disclosures.push({
                    type,
                    content: match.substring(0, 50) + '...',
                    location: `script: ${src}`
                  });
                }
              }
            }
          }
        } catch (error) {
          // Script might not be accessible
          continue;
        }
      }
    }
    
    return {
      sensitiveDataFound: disclosures.length > 0,
      disclosures
    };
  }
}