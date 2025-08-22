import { test, expect } from '@playwright/test'

test.describe('Authentication Parity Tests', () => {
  test('new app loads at /v2', async ({ page }) => {
    await page.goto('http://localhost:5173/v2/')

    // Check that the page loads
    await expect(page.locator('header h1').first()).toContainText('OllamaMax')
    
    // Check navigation links are present
    await expect(page.locator('nav a[href="/v2"]').first()).toContainText('Dashboard')
    await expect(page.locator('nav a[href="/v2/auth/login"]').first()).toContainText('Login')
    await expect(page.locator('nav a[href="/v2/auth/register"]').first()).toContainText('Register')
  })

  test('login form renders correctly', async ({ page }) => {
    await page.goto('http://localhost:5173/v2/auth/login')

    // Check form elements
    await expect(page.locator('form h1')).toContainText('Sign in')
    await expect(page.locator('input[type="email"]')).toBeVisible()
    await expect(page.locator('input[type="password"]')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toContainText('Sign in')
  })

  test('shared Button component renders with correct styles', async ({ page }) => {
    await page.goto('http://localhost:5173/v2/auth/login')
    
    const button = page.locator('button[type="submit"]')
    await expect(button).toBeVisible()
    
    // Check that CSS variables are applied (design tokens working)
    const buttonStyles = await button.evaluate((el) => {
      const computed = window.getComputedStyle(el)
      return {
        background: computed.backgroundColor,
        color: computed.color,
        borderRadius: computed.borderRadius
      }
    })
    
    // Verify that styles are applied (not default browser styles)
    expect(buttonStyles.background).not.toBe('rgba(0, 0, 0, 0)')
    expect(buttonStyles.borderRadius).not.toBe('0px')
  })

  test('CSS variables are available', async ({ page }) => {
    await page.goto('http://localhost:5173/v2/')
    
    const cssVars = await page.evaluate(() => {
      const root = document.documentElement
      const computed = window.getComputedStyle(root)
      return {
        brandColor: computed.getPropertyValue('--omx-color-brand-500').trim(),
        spacing: computed.getPropertyValue('--omx-spacing-4').trim(),
        radius: computed.getPropertyValue('--omx-radius-md').trim()
      }
    })
    
    expect(cssVars.brandColor).toBe('#0ea5e9')
    expect(cssVars.spacing).toBe('1rem')
    expect(cssVars.radius).toBe('0.375rem')
  })

  test('form validation works', async ({ page }) => {
    await page.goto('http://localhost:5173/v2/auth/login')
    
    // Try to submit empty form
    await page.click('button[type="submit"]')
    
    // Check HTML5 validation
    const emailInput = page.locator('input[type="email"]')
    const isEmailInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid)
    expect(isEmailInvalid).toBe(true)
  })
})
