import { test, expect } from '@playwright/test'

// Visual baselines for SideNav and Breadcrumbs

test.describe('Navigation visuals', () => {
  test('v2 breadcrumbs (flag on)', async ({ page }) => {
    await page.addInitScript(() => localStorage.setItem('USE_SHARED_BREADCRUMBS', '1'))
    await page.goto('/v2/auth/login')
    const bc = page.locator('nav[aria-label="Breadcrumb"]')
    await expect(bc).toBeVisible()
    await expect(bc).toHaveScreenshot('v2-breadcrumbs.png')
  })

  test('v2 sidenav (flag on)', async ({ page }) => {
    await page.addInitScript(() => localStorage.setItem('USE_SHARED_SIDENAV', '1'))
    await page.goto('/v2/auth/login')
    const nav = page.locator('aside[role="navigation"]')
    await expect(nav).toBeVisible()
    await expect(nav).toHaveScreenshot('v2-sidenav.png')
  })

  test('legacy breadcrumbs (flag on)', async ({ page }) => {
    await page.addInitScript(() => localStorage.setItem('USE_SHARED_BREADCRUMBS', '1'))
    const legacyBase = process.env.LEGACY_BASE || 'http://localhost:8090/'
    await page.goto(legacyBase)
    const bc = page.locator('#ui-breadcrumbs nav[aria-label="Breadcrumb"]')
    await expect(bc).toBeVisible()
    await expect(bc).toHaveScreenshot('legacy-breadcrumbs.png')
  })

  test('legacy sidenav (flag on)', async ({ page }) => {
    await page.addInitScript(() => localStorage.setItem('USE_SHARED_SIDENAV', '1'))
    const legacyBase = process.env.LEGACY_BASE || 'http://localhost:8090/'
    await page.goto(legacyBase)
    const nav = page.locator('#ui-sidenav aside[role="navigation"]')
    await expect(nav).toBeVisible()
    await expect(nav).toHaveScreenshot('legacy-sidenav.png')
  })
})

