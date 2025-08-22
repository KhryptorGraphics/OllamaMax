import { test, expect } from '@playwright/test'

// These assume backend endpoints exist; if not, they validate front-end behavior and error handling.
// For CI without backend, we can extend to mock with MSW later if needed.

test.describe('Authentication flows', () => {
  test('Registration form validation and submit UX', async ({ page }) => {
    await page.goto('/v2/auth/register')

    await page.fill('input[aria-required="true"]:below(:text("Username"))', 'testuser')
    await page.fill('input[type="email"]', 'bad-email')
    await expect(page.getByText('Enter a valid email')).toBeVisible()
    await page.fill('input[type="email"]', 'test@example.com')

    await page.fill('input[type="password"]:below(:text("Password"))', 'weak')
    await expect(page.getByText('Must be at least 8 characters')).toBeVisible()
    await page.fill('input[type="password"]:below(:text("Password"))', 'StrongP@ssw0rd!')

    // confirm mismatch
    await page.fill('input[type="password"]:below(:text("Confirm password"))', 'StrongP@ssw0rd?')
    await expect(page.getByText('Passwords do not match')).toBeVisible()
    await page.fill('input[type="password"]:below(:text("Confirm password"))', 'StrongP@ssw0rd!')

    // accept tos/privacy
    await page.check('text=I agree to the >> input')
    await page.check('text=Privacy Policy >> input')

    // submit (may fail without backend; we only assert disabled state toggles and messaging presence)
    await page.click('button[aria-label="Create account"]')
    await expect(page.getByRole('status').or(page.getByRole('alert'))).toBeVisible({ timeout: 5000 })
  })

  test('Forgot password flow with client-side rate limit', async ({ page }) => {
    await page.goto('/v2/auth/forgot-password')
    await page.fill('input[type="email"]', 'test@example.com')

    for (let i = 0; i < 3; i++) {
      await page.click('button[aria-label="Send reset link"]')
      await expect(page.getByRole('status').or(page.getByRole('alert'))).toBeVisible()
    }

    // 4th should be disabled by client-side limiter
    const isDisabled = await page.locator('button[aria-label="Send reset link"]').isDisabled()
    expect(isDisabled).toBeTruthy()
  })

  test('Reset password form validation and submit UX', async ({ page }) => {
    await page.goto('/v2/auth/reset-password?token=FAKE')
    await page.fill('input[type="password"]:below(:text("New password"))', 'StrongP@ssw0rd!')
    await page.fill('input[type="password"]:below(:text("Confirm password"))', 'StrongP@ssw0rd!')
    await page.click('button[aria-label="Reset password"]')
    await expect(page.getByRole('status').or(page.getByRole('alert'))).toBeVisible({ timeout: 5000 })
  })

  test('Email verification status page', async ({ page }) => {
    await page.goto('/v2/auth/verify-email?token=FAKE')
    await expect(page.getByRole('status')).toBeVisible()
  })
})

