import { test, expect } from '@playwright/test'
test.describe('Auth pages accessibility', () => {
  const routes = ['/v2/auth/login', '/v2/auth/register', '/v2/auth/forgot-password', '/v2/auth/reset-password?token=TEST', '/v2/auth/verify-email?token=TEST']

  for (const route of routes) {
    test(`no critical a11y violations on ${route}`, async ({ page }) => {
      await page.goto(route)
      let AxeBuilder: any
      try {
        AxeBuilder = (await import('@axe-core/playwright')).default
      } catch (e) {
        test.skip(true, 'Skipping axe-core test because @axe-core/playwright is not installed')
        return
      }
      const results = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa']).analyze()
      const critical = results.violations.filter((v: any) => ['critical', 'serious'].includes(v.impact || ''))
      expect(critical, JSON.stringify(results.violations, null, 2)).toHaveLength(0)
    })
  }
})

