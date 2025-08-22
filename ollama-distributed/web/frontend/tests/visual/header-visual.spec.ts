import { test, expect } from '@playwright/test'

// Visual baseline for shared Header in /v2 and legacy /
// Uses toHaveScreenshot with default thresholds from config

test.describe('Header visual regression', () => {
  test('v2 header baseline', async ({ page }) => {
    await page.goto('/v2/')
    const header = page.locator('header[role="banner"]')
    await expect(header).toBeVisible()
    await expect(header).toHaveScreenshot('v2-header.png')
  })

  test('legacy header baseline (shared via flag)', async ({ page, context }) => {
    // Enable shared header in legacy
    await context.addInitScript(() => {
      window.localStorage.setItem('USE_SHARED_HEADER', '1')
    })

    // Legacy app is served at another port by a static script if needed
    const legacyBase = process.env.LEGACY_BASE || 'http://localhost:8090/'

    await page.goto(legacyBase)
    const header = page.locator('#ui-header header[role="banner"]')
    await expect(header).toBeVisible()
    await expect(header).toHaveScreenshot('legacy-header.png')
  })
})

