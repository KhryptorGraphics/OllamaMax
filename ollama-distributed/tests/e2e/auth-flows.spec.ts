import { test, expect, Page, BrowserContext } from '@playwright/test'

/**
 * Authentication Flow E2E Tests
 * 
 * Covers complete authentication workflows including:
 * - Login/logout flows
 * - Registration with validation
 * - Password reset functionality
 * - Email verification
 * - Session management
 * - Security features
 */

test.describe('Authentication Flows', () => {
  let context: BrowserContext
  let page: Page

  test.beforeEach(async ({ browser }) => {
    context = await browser.newContext()
    page = await context.newPage()
  })

  test.afterEach(async () => {
    await context.close()
  })

  test.describe('Login Flow', () => {
    test('successful login with valid credentials', async () => {
      await page.goto('/v2/auth/login')
      
      // Verify login form is present
      await expect(page.locator('form[aria-label="Login"]')).toBeVisible()
      await expect(page.locator('h1')).toContainText('Sign in')
      
      // Fill login form
      await page.fill('input[type="email"]', 'admin@example.com')
      await page.fill('input[type="password"]', 'SecurePass123!')
      
      // Submit form
      await page.click('button[type="submit"]')
      
      // Verify successful login
      await expect(page).toHaveURL('/v2')
      await expect(page.locator('header')).toContainText('OllamaMax')
    })

    test('login fails with invalid credentials', async () => {
      await page.goto('/v2/auth/login')
      
      await page.fill('input[type="email"]', 'invalid@example.com')
      await page.fill('input[type="password"]', 'wrongpassword')
      await page.click('button[type="submit"]')
      
      // Verify error message appears
      await expect(page.locator('[role="alert"]')).toBeVisible()
      await expect(page.locator('[role="alert"]')).toContainText(/login failed|invalid/i)
      
      // Verify still on login page
      await expect(page).toHaveURL(/login/)
    })

    test('login form validation', async () => {
      await page.goto('/v2/auth/login')
      
      // Test empty form submission
      await page.click('button[type="submit"]')
      
      // Verify HTML5 validation
      const emailInput = page.locator('input[type="email"]')
      const isEmailInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid)
      expect(isEmailInvalid).toBe(true)
      
      // Test invalid email format
      await page.fill('input[type="email"]', 'invalid-email')
      await page.click('button[type="submit"]')
      
      const emailValidityAfterInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid)
      expect(emailValidityAfterInvalid).toBe(true)
    })

    test('login form accessibility', async () => {
      await page.goto('/v2/auth/login')
      
      // Verify ARIA attributes
      await expect(page.locator('form')).toHaveAttribute('aria-label', 'Login')
      await expect(page.locator('input[type="email"]')).toHaveAttribute('aria-required', 'true')
      await expect(page.locator('input[type="password"]')).toHaveAttribute('aria-required', 'true')
      
      // Test keyboard navigation
      await page.keyboard.press('Tab')
      await expect(page.locator('input[type="email"]')).toBeFocused()
      
      await page.keyboard.press('Tab')
      await expect(page.locator('input[type="password"]')).toBeFocused()
      
      await page.keyboard.press('Tab')
      await expect(page.locator('button[type="submit"]')).toBeFocused()
    })

    test('loading state during login', async () => {
      await page.goto('/v2/auth/login')
      
      await page.fill('input[type="email"]', 'test@example.com')
      await page.fill('input[type="password"]', 'password123')
      
      // Mock slow network to test loading state
      await page.route('**/api/auth/login', route => {
        setTimeout(() => route.continue(), 2000)
      })
      
      const submitButton = page.locator('button[type="submit"]')
      await submitButton.click()
      
      // Verify loading state
      await expect(submitButton).toBeDisabled()
      await expect(submitButton).toContainText(/signing in/i)
    })
  })

  test.describe('Registration Flow', () => {
    test('successful registration with valid data', async () => {
      await page.goto('/v2/auth/register')
      
      // Fill registration form
      await page.fill('input[name="username"]', 'newuser123')
      await page.fill('input[type="email"]', 'newuser@example.com')
      await page.fill('input[type="password"]', 'SecurePass123!')
      await page.fill('input[name="confirm"]', 'SecurePass123!')
      
      // Accept terms and privacy
      await page.check('input[type="checkbox"]:has-text("Terms of Service")')
      await page.check('input[type="checkbox"]:has-text("Privacy Policy")')
      
      // Submit form
      await page.click('button[type="submit"]')
      
      // Verify success message
      await expect(page.locator('[role="status"]')).toBeVisible()
      await expect(page.locator('[role="status"]')).toContainText(/registration successful/i)
    })

    test('registration form validation', async () => {
      await page.goto('/v2/auth/register')
      
      // Test password strength indicator
      const passwordInput = page.locator('input[type="password"]').first()
      await passwordInput.fill('weak')\n      
      await expect(page.locator('text=Very weak')).toBeVisible()
      
      await passwordInput.fill('StrongPass123!')\n      
      await expect(page.locator('text=Strong')).toBeVisible()
      
      // Test password confirmation mismatch
      await page.fill('input[name="confirm"]', 'DifferentPass123!')
      await expect(page.locator('text=Passwords do not match')).toBeVisible()
    })

    test('email availability check', async () => {
      await page.goto('/v2/auth/register')
      
      // Mock email availability API
      await page.route('**/api/auth/check-email', route => {\n        const url = new URL(route.request().url());\n        const email = url.searchParams.get('email');\n        if (email === 'taken@example.com') {\n          route.fulfill({ status: 409, body: JSON.stringify({ error: 'Email taken' }) });\n        } else {\n          route.fulfill({ status: 200, body: JSON.stringify({ available: true }) });\n        }\n      })
      
      // Test taken email
      await page.fill('input[type="email"]', 'taken@example.com')
      await page.waitForTimeout(500) // Wait for debounced check
      await expect(page.locator('text=Email already in use')).toBeVisible()
    })

    test('terms and privacy policy requirements', async () => {
      await page.goto('/v2/auth/register')
      
      // Fill all fields except checkboxes
      await page.fill('input[name="username"]', 'testuser')
      await page.fill('input[type="email"]', 'test@example.com')
      await page.fill('input[type="password"]', 'SecurePass123!')
      await page.fill('input[name="confirm"]', 'SecurePass123!')
      
      // Verify submit button is disabled without checkboxes
      await expect(page.locator('button[type="submit"]')).toBeDisabled()
      
      // Check one checkbox - should still be disabled
      await page.check('input[type="checkbox"]:has-text("Terms of Service")')
      await expect(page.locator('button[type="submit"]')).toBeDisabled()
      
      // Check both checkboxes - should be enabled
      await page.check('input[type="checkbox"]:has-text("Privacy Policy")')
      await expect(page.locator('button[type="submit"]')).toBeEnabled()
    })
  })

  test.describe('Password Reset Flow', () => {
    test('forgot password request', async () => {
      await page.goto('/v2/auth/forgot-password')
      
      await page.fill('input[type="email"]', 'user@example.com')
      await page.click('button[type="submit"]')
      
      // Verify success message (generic for security)
      await expect(page.locator('[role="status"]')).toContainText(/if an account exists/i)
    })

    test('password reset with valid token', async () => {
      const resetToken = 'valid-reset-token-123'\n      await page.goto(`/v2/auth/reset-password?token=${resetToken}`)
      
      await page.fill('input[type="password"]', 'NewSecurePass123!')
      await page.fill('input[name="confirm"]', 'NewSecurePass123!')
      await page.click('button[type="submit"]')
      
      await expect(page.locator('[role="status"]')).toContainText(/password has been reset/i)
    })

    test('password reset rate limiting', async () => {
      await page.goto('/v2/auth/forgot-password')
      
      const emailInput = page.locator('input[type="email"]')
      const submitButton = page.locator('button[type="submit"]')
      
      // Submit 3 requests
      for (let i = 0; i < 3; i++) {
        await emailInput.fill('test@example.com')
        await submitButton.click()
        await page.waitForSelector('[role="status"]')
      }
      
      // 4th request should be blocked
      await emailInput.fill('test@example.com')
      await expect(submitButton).toBeDisabled()
      await expect(page.locator('text=3 reset emails per hour')).toBeVisible()
    })
  })

  test.describe('Email Verification', () => {
    test('successful email verification', async () => {
      const verificationToken = 'valid-verification-token'
      await page.goto(`/v2/auth/verify-email?token=${verificationToken}`)
      
      await expect(page.locator('[role="status"]')).toContainText(/email verified successfully/i)
    })

    test('failed email verification', async () => {
      await page.goto('/v2/auth/verify-email?token=invalid-token')
      
      await expect(page.locator('text=Verification failed')).toBeVisible()
    })
  })

  test.describe('Session Management', () => {
    test('logout functionality', async () => {
      // Login first
      await page.goto('/v2/auth/login')
      await page.fill('input[type="email"]', 'admin@example.com')
      await page.fill('input[type="password"]', 'SecurePass123!')
      await page.click('button[type="submit"]')
      
      await expect(page).toHaveURL('/v2')
      
      // Logout
      await page.click('button:has-text("Logout")')
      
      // Verify redirected to login
      await expect(page).toHaveURL(/login/)
    })

    test('session persistence across page refreshes', async () => {
      // Login
      await page.goto('/v2/auth/login')
      await page.fill('input[type="email"]', 'admin@example.com')
      await page.fill('input[type="password"]', 'SecurePass123!')
      await page.click('button[type="submit"]')
      
      // Refresh page
      await page.reload()
      
      // Should still be logged in
      await expect(page).toHaveURL('/v2')
      await expect(page.locator('header')).toContainText('OllamaMax')
    })

    test('automatic logout on token expiration', async () => {
      // Mock expired token response
      await page.route('**/api/**', route => {
        if (route.request().headers()['authorization']) {
          route.fulfill({ status: 401, body: JSON.stringify({ error: 'Token expired' }) })
        } else {
          route.continue()
        }
      })
      
      await page.goto('/v2/auth/login')
      await page.fill('input[type="email"]', 'admin@example.com')
      await page.fill('input[type="password"]', 'SecurePass123!')
      await page.click('button[type="submit"]')
      
      // Try to access protected resource
      await page.goto('/v2/admin')
      
      // Should be redirected to login
      await expect(page).toHaveURL(/login/)
    })
  })

  test.describe('Security Features', () => {
    test('XSS protection in forms', async () => {
      await page.goto('/v2/auth/login')
      
      const maliciousScript = '<script>alert("XSS")</script>'
      await page.fill('input[type="email"]', maliciousScript)
      
      // Verify script is not executed
      const emailValue = await page.inputValue('input[type="email"]')
      expect(emailValue).toBe(maliciousScript) // Should be treated as text
      
      // No alert should have appeared
      page.on('dialog', () => {
        throw new Error('XSS alert detected!')
      })
    })

    test('CSRF protection', async () => {
      // Mock CSRF token validation
      await page.route('**/api/auth/login', route => {
        const csrfToken = route.request().headers()['x-csrf-token']
        if (!csrfToken) {
          route.fulfill({ status: 403, body: JSON.stringify({ error: 'CSRF token missing' }) })
        } else {
          route.continue()
        }
      })
      
      await page.goto('/v2/auth/login')
      await page.fill('input[type="email"]', 'test@example.com')
      await page.fill('input[type="password"]', 'password123')
      await page.click('button[type="submit"]')
      
      // Should show CSRF error
      await expect(page.locator('[role="alert"]')).toContainText(/csrf/i)
    })
  })
})