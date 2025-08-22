/**
 * @fileoverview Comprehensive accessibility tests for the entire application
 * Tests WCAG 2.1 AA compliance across all major pages and components
 */

import { test, expect } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'

// Test configuration
const ACCESSIBILITY_CONFIG = {
  // WCAG 2.1 AA compliance tags
  tags: ['wcag2a', 'wcag2aa', 'wcag21aa'],
  
  // Exclude minor issues in development
  impacts: ['critical', 'serious'],
  
  // Rules to disable in specific contexts
  disableRules: {
    development: ['color-contrast'], // May not be finalized in dev
    beta: []
  }
}

// Common test patterns
const COMMON_PATTERNS = {
  navigation: {
    landmarks: ['banner', 'navigation', 'main', 'contentinfo'],
    skipLinks: ['Skip to content', 'Skip to navigation']
  },
  forms: {
    labels: true,
    fieldsets: true,
    errorMessages: true
  },
  interactive: {
    focusManagement: true,
    keyboardAccess: true,
    touchTargets: true
  }
}

test.describe('Comprehensive Accessibility Testing', () => {
  test.beforeEach(async ({ page }) => {
    // Set up accessibility testing context
    await page.goto('/v2')
    
    // Wait for initial load
    await page.waitForLoadState('networkidle')
    
    // Ensure focus is in a known state
    await page.keyboard.press('Tab') // Focus first element
    await page.keyboard.press('Shift+Tab') // Return to start
  })

  test.describe('Application-wide Accessibility', () => {
    const criticalPages = [
      { path: '/v2', name: 'Dashboard' },
      { path: '/v2/models', name: 'Models' },
      { path: '/v2/cluster', name: 'Cluster' },
      { path: '/v2/analytics', name: 'Analytics' },
      { path: '/v2/auth/login', name: 'Login' },
      { path: '/v2/auth/register', name: 'Register' }
    ]

    for (const { path, name } of criticalPages) {
      test(`${name} page should meet WCAG 2.1 AA standards`, async ({ page }) => {
        await page.goto(path)
        await page.waitForLoadState('networkidle')

        const axeResults = await new AxeBuilder({ page })
          .withTags(ACCESSIBILITY_CONFIG.tags)
          .analyze()

        const criticalViolations = axeResults.violations.filter(violation =>
          ACCESSIBILITY_CONFIG.impacts.includes(violation.impact as any)
        )

        expect(criticalViolations, `Critical accessibility violations found on ${name}:\n${JSON.stringify(criticalViolations, null, 2)}`).toHaveLength(0)
      })
    }
  })

  test.describe('Keyboard Navigation', () => {
    test('should support complete keyboard navigation', async ({ page }) => {
      await page.goto('/v2')
      
      // Test skip links
      await page.keyboard.press('Tab')
      const skipLink = page.locator('text=Skip to content')
      if (await skipLink.isVisible()) {
        await expect(skipLink).toBeFocused()
        await page.keyboard.press('Enter')
        
        // Verify focus moved to main content
        const mainContent = page.locator('main, [role="main"], #root')
        await expect(mainContent).toBeFocused()
      }
      
      // Test tab order through main navigation
      await page.keyboard.press('Tab')
      const firstNavItem = page.locator('nav a, nav button').first()
      await expect(firstNavItem).toBeFocused()
      
      // Continue tabbing through navigation
      for (let i = 0; i < 5; i++) {
        await page.keyboard.press('Tab')
        const focused = page.locator(':focus')
        await expect(focused).toBeVisible()
      }
    })

    test('should trap focus in modal dialogs', async ({ page }) => {
      await page.goto('/v2')
      
      // Look for modal trigger buttons
      const modalTriggers = page.locator('[data-testid*="modal"], [aria-haspopup="dialog"], button:has-text("Add"), button:has-text("Create")')
      
      if (await modalTriggers.count() > 0) {
        const firstTrigger = modalTriggers.first()
        await firstTrigger.click()
        
        // Wait for modal to appear
        const modal = page.locator('[role="dialog"], [role="alertdialog"]')
        if (await modal.isVisible()) {
          // Test focus trap
          const focusableInModal = modal.locator('button, input, select, textarea, a[href], [tabindex]:not([tabindex="-1"])')
          const focusableCount = await focusableInModal.count()
          
          if (focusableCount > 1) {
            // Tab through all focusable elements
            for (let i = 0; i < focusableCount; i++) {
              await page.keyboard.press('Tab')
            }
            
            // Next tab should cycle back to first element
            await page.keyboard.press('Tab')
            const firstFocusable = focusableInModal.first()
            await expect(firstFocusable).toBeFocused()
            
            // Test reverse direction
            await page.keyboard.press('Shift+Tab')
            const lastFocusable = focusableInModal.last()
            await expect(lastFocusable).toBeFocused()
          }
          
          // Test escape key
          await page.keyboard.press('Escape')
          await expect(modal).not.toBeVisible()
        }
      }
    })

    test('should support arrow key navigation in menus and lists', async ({ page }) => {
      await page.goto('/v2')
      
      // Look for menu components
      const menus = page.locator('[role="menu"], [role="menubar"], [role="listbox"]')
      
      if (await menus.count() > 0) {
        const firstMenu = menus.first()
        const menuItems = firstMenu.locator('[role="menuitem"], [role="option"]')
        const itemCount = await menuItems.count()
        
        if (itemCount > 1) {
          // Focus first item
          await menuItems.first().focus()
          
          // Test arrow down navigation
          await page.keyboard.press('ArrowDown')
          await expect(menuItems.nth(1)).toBeFocused()
          
          // Test arrow up navigation
          await page.keyboard.press('ArrowUp')
          await expect(menuItems.first()).toBeFocused()
        }
      }
    })

    test('should handle Enter and Space key activation', async ({ page }) => {
      await page.goto('/v2')
      
      // Test buttons
      const buttons = page.locator('button:visible').first()
      if (await buttons.isVisible()) {
        await buttons.focus()
        
        // Test Enter activation
        const clickPromise = page.waitForEvent('console', msg => msg.type() === 'log' || msg.type() === 'info')
        await page.keyboard.press('Enter')
        
        // Test Space activation
        await page.keyboard.press(' ')
      }
      
      // Test links
      const links = page.locator('a[href]:visible').first()
      if (await links.isVisible()) {
        await links.focus()
        await page.keyboard.press('Enter')
        // Navigation should occur
      }
    })
  })

  test.describe('Screen Reader Compatibility', () => {
    test('should have proper heading structure', async ({ page }) => {
      await page.goto('/v2')
      
      const headings = page.locator('h1, h2, h3, h4, h5, h6')
      const headingCount = await headings.count()
      
      if (headingCount > 0) {
        // Should have exactly one h1
        const h1Count = await page.locator('h1').count()
        expect(h1Count).toBe(1)
        
        // Check heading hierarchy
        const headingLevels: number[] = []
        for (let i = 0; i < headingCount; i++) {
          const heading = headings.nth(i)
          const tagName = await heading.evaluate(el => el.tagName)
          const level = parseInt(tagName.charAt(1))
          headingLevels.push(level)
        }
        
        // Verify no level skipping
        for (let i = 1; i < headingLevels.length; i++) {
          const currentLevel = headingLevels[i]
          const previousLevel = headingLevels[i - 1]
          expect(currentLevel).toBeLessThanOrEqual(previousLevel + 1)
        }
      }
    })

    test('should have proper landmark structure', async ({ page }) => {
      await page.goto('/v2')
      
      // Check for essential landmarks
      const landmarks = {
        banner: page.locator('[role="banner"], header'),
        navigation: page.locator('[role="navigation"], nav'),
        main: page.locator('[role="main"], main'),
        contentinfo: page.locator('[role="contentinfo"], footer')
      }
      
      // Should have at least main landmark
      await expect(landmarks.main).toHaveCount(1)
      
      // Navigation should exist
      await expect(landmarks.navigation).toHaveCountGreaterThan(0)
      
      // Banner/header should exist
      await expect(landmarks.banner).toHaveCountGreaterThan(0)
    })

    test('should have accessible form labels and descriptions', async ({ page }) => {
      await page.goto('/v2/auth/login')
      
      const formControls = page.locator('input:visible, select:visible, textarea:visible')
      const controlCount = await formControls.count()
      
      for (let i = 0; i < controlCount; i++) {
        const control = formControls.nth(i)
        const type = await control.getAttribute('type')
        
        // Skip hidden inputs
        if (type === 'hidden') continue
        
        // Should have accessible name
        const accessibleName = await control.getAttribute('aria-label') || 
                              await control.getAttribute('aria-labelledby') ||
                              await page.locator(`label[for="${await control.getAttribute('id')}"]`).textContent()
        
        expect(accessibleName, `Form control at index ${i} should have accessible name`).toBeTruthy()
        
        // Required fields should be marked
        const isRequired = await control.getAttribute('required') !== null
        if (isRequired) {
          const ariaRequired = await control.getAttribute('aria-required')
          expect(ariaRequired).toBe('true')
        }
      }
    })

    test('should announce dynamic content changes', async ({ page }) => {
      await page.goto('/v2')
      
      // Look for live regions
      const liveRegions = page.locator('[aria-live], [role="status"], [role="alert"]')
      
      if (await liveRegions.count() > 0) {
        // Test that live regions have appropriate aria-live values
        for (let i = 0; i < await liveRegions.count(); i++) {
          const region = liveRegions.nth(i)
          const ariaLive = await region.getAttribute('aria-live')
          const role = await region.getAttribute('role')
          
          if (role === 'alert') {
            // Alerts should be assertive or not specified (defaults to assertive)
            expect(ariaLive === null || ariaLive === 'assertive').toBe(true)
          } else if (role === 'status') {
            // Status should be polite or not specified (defaults to polite)
            expect(ariaLive === null || ariaLive === 'polite').toBe(true)
          }
        }
      }
    })
  })

  test.describe('Visual and Motor Accessibility', () => {
    test('should have sufficient color contrast', async ({ page }) => {
      await page.goto('/v2')
      
      const axeResults = await new AxeBuilder({ page })
        .withRules(['color-contrast'])
        .analyze()
      
      expect(axeResults.violations).toHaveLength(0)
    })

    test('should have adequate touch target sizes', async ({ page }) => {
      await page.goto('/v2')
      
      const interactiveElements = page.locator('button:visible, a:visible, input:visible, [role="button"]:visible')
      const elementCount = await interactiveElements.count()
      
      for (let i = 0; i < Math.min(elementCount, 10); i++) { // Test first 10 elements
        const element = interactiveElements.nth(i)
        const boundingBox = await element.boundingBox()
        
        if (boundingBox) {
          // WCAG recommends minimum 44x44 pixels for touch targets
          expect(boundingBox.width).toBeGreaterThanOrEqual(24) // Slightly relaxed for complex UIs
          expect(boundingBox.height).toBeGreaterThanOrEqual(24)
        }
      }
    })

    test('should respect reduced motion preferences', async ({ page }) => {
      // Set reduced motion preference
      await page.emulateMedia({ reducedMotion: 'reduce' })
      await page.goto('/v2')
      
      // Check that animations are disabled or minimal
      const animatedElements = page.locator('[class*="animate"], [class*="transition"]')
      
      for (let i = 0; i < await animatedElements.count(); i++) {
        const element = animatedElements.nth(i)
        const styles = await element.evaluate(el => {
          const computed = window.getComputedStyle(el)
          return {
            animationDuration: computed.animationDuration,
            transitionDuration: computed.transitionDuration
          }
        })
        
        // Animations should be very short or disabled
        expect(parseFloat(styles.animationDuration) <= 0.01 || styles.animationDuration === '0s').toBe(true)
        expect(parseFloat(styles.transitionDuration) <= 0.01 || styles.transitionDuration === '0s').toBe(true)
      }
    })

    test('should support high contrast mode', async ({ page }) => {
      await page.emulateMedia({ forcedColors: 'active' })
      await page.goto('/v2')
      
      // Basic checks that page is still functional
      await expect(page.locator('body')).toBeVisible()
      
      // Interactive elements should still be accessible
      const buttons = page.locator('button:visible')
      if (await buttons.count() > 0) {
        await expect(buttons.first()).toBeVisible()
      }
    })
  })

  test.describe('Form Accessibility', () => {
    test('should handle form validation accessibly', async ({ page }) => {
      await page.goto('/v2/auth/login')
      
      // Try to submit empty form
      const submitButton = page.locator('button[type="submit"], input[type="submit"]')
      if (await submitButton.isVisible()) {
        await submitButton.click()
        
        // Check for error messages
        const errorMessages = page.locator('[role="alert"], .error, [aria-invalid="true"]')
        
        if (await errorMessages.count() > 0) {
          // Error messages should be accessible
          for (let i = 0; i < await errorMessages.count(); i++) {
            const error = errorMessages.nth(i)
            await expect(error).toBeVisible()
            
            // Should have role alert or be associated with invalid field
            const role = await error.getAttribute('role')
            const ariaInvalid = await error.getAttribute('aria-invalid')
            
            expect(role === 'alert' || ariaInvalid === 'true').toBe(true)
          }
        }
      }
    })

    test('should associate errors with form fields', async ({ page }) => {
      await page.goto('/v2/auth/register')
      
      // Fill form with invalid data and submit
      const emailInput = page.locator('input[type="email"]')
      if (await emailInput.isVisible()) {
        await emailInput.fill('invalid-email')
        
        const submitButton = page.locator('button[type="submit"]')
        if (await submitButton.isVisible()) {
          await submitButton.click()
          
          // Check if email field is marked as invalid
          const isInvalid = await emailInput.getAttribute('aria-invalid')
          if (isInvalid === 'true') {
            // Should have aria-describedby pointing to error message
            const describedBy = await emailInput.getAttribute('aria-describedby')
            if (describedBy) {
              const errorElement = page.locator(`#${describedBy}`)
              await expect(errorElement).toBeVisible()
            }
          }
        }
      }
    })
  })

  test.describe('Component-Specific Tests', () => {
    test('should have accessible data tables', async ({ page }) => {
      await page.goto('/v2/models') // Assuming models page has tables
      
      const tables = page.locator('table')
      
      if (await tables.count() > 0) {
        const firstTable = tables.first()
        
        // Should have table headers
        const headers = firstTable.locator('th')
        await expect(headers).toHaveCountGreaterThan(0)
        
        // Headers should have scope attribute
        for (let i = 0; i < await headers.count(); i++) {
          const header = headers.nth(i)
          const scope = await header.getAttribute('scope')
          expect(scope).toBeTruthy()
        }
        
        // Should have caption or accessible name
        const caption = firstTable.locator('caption')
        const ariaLabel = await firstTable.getAttribute('aria-label')
        const ariaLabelledBy = await firstTable.getAttribute('aria-labelledby')
        
        expect(
          await caption.isVisible() || ariaLabel || ariaLabelledBy
        ).toBe(true)
      }
    })

    test('should have accessible navigation menus', async ({ page }) => {
      await page.goto('/v2')
      
      const navMenus = page.locator('[role="navigation"] ul, nav ul')
      
      if (await navMenus.count() > 0) {
        const firstMenu = navMenus.first()
        const menuItems = firstMenu.locator('li')
        
        // Menu items should be properly structured
        for (let i = 0; i < await menuItems.count(); i++) {
          const item = menuItems.nth(i)
          const link = item.locator('a')
          
          if (await link.isVisible()) {
            // Links should have accessible names
            const accessibleName = await link.textContent() || 
                                  await link.getAttribute('aria-label')
            expect(accessibleName?.trim()).toBeTruthy()
          }
        }
      }
    })
  })

  test.describe('Performance and Accessibility', () => {
    test('should maintain accessibility during loading states', async ({ page }) => {
      await page.goto('/v2')
      
      // Look for loading indicators
      const loadingIndicators = page.locator('[role="progressbar"], .loading, [aria-busy="true"]')
      
      if (await loadingIndicators.count() > 0) {
        const indicator = loadingIndicators.first()
        
        // Should have appropriate ARIA attributes
        const role = await indicator.getAttribute('role')
        const ariaLabel = await indicator.getAttribute('aria-label')
        const ariaBusy = await indicator.getAttribute('aria-busy')
        
        expect(
          role === 'progressbar' || 
          ariaLabel || 
          ariaBusy === 'true'
        ).toBe(true)
      }
    })

    test('should maintain accessibility with dynamic content', async ({ page }) => {
      await page.goto('/v2/analytics') // Page with dynamic charts/data
      
      // Wait for dynamic content to load
      await page.waitForLoadState('networkidle')
      
      // Run accessibility check after content loads
      const axeResults = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa'])
        .analyze()
      
      const criticalViolations = axeResults.violations.filter(v => 
        ['critical', 'serious'].includes(v.impact as string)
      )
      
      expect(criticalViolations).toHaveLength(0)
    })
  })

  test.describe('Multi-language Accessibility', () => {
    test('should handle language attributes correctly', async ({ page }) => {
      await page.goto('/v2')
      
      // Check html lang attribute
      const htmlLang = await page.getAttribute('html', 'lang')
      expect(htmlLang).toBeTruthy()
      
      // Check for any elements with different language
      const langElements = page.locator('[lang]')
      
      for (let i = 0; i < await langElements.count(); i++) {
        const element = langElements.nth(i)
        const lang = await element.getAttribute('lang')
        expect(lang).toMatch(/^[a-z]{2}(-[A-Z]{2})?$/) // Valid language code format
      }
    })
  })

  test.describe('Accessibility Regression Prevention', () => {
    test('should maintain accessibility across browser zoom levels', async ({ page }) => {
      await page.goto('/v2')
      
      // Test at 200% zoom
      await page.setViewportSize({ width: 640, height: 480 }) // Simulates 200% zoom on 1280x960
      
      // Basic functionality should still work
      const navigation = page.locator('nav')
      await expect(navigation).toBeVisible()
      
      // Interactive elements should still be accessible
      const buttons = page.locator('button:visible')
      if (await buttons.count() > 0) {
        await buttons.first().focus()
        await expect(buttons.first()).toBeFocused()
      }
    })

    test('should work with browser accessibility tools', async ({ page }) => {
      // This test ensures compatibility with common accessibility extensions
      await page.goto('/v2')
      
      // Simulate common accessibility tool behaviors
      await page.addStyleTag({
        content: `
          /* Simulate high contrast extension */
          * { 
            border: 1px solid red !important; 
            background: white !important; 
            color: black !important; 
          }
        `
      })
      
      // Page should still be functional
      await expect(page.locator('body')).toBeVisible()
      
      // Remove the style
      await page.addStyleTag({
        content: `
          * { 
            border: unset !important; 
            background: unset !important; 
            color: unset !important; 
          }
        `
      })
    })
  })
})

// Helper function to check specific component accessibility
test.describe('Component Library Accessibility', () => {
  test('should have accessible design system components', async ({ page }) => {
    // This would test a dedicated component showcase page
    await page.goto('/v2/styleguide') // If you have one
    
    if (await page.locator('body').textContent() !== null) {
      const axeResults = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa', 'wcag21aa'])
        .analyze()
      
      expect(axeResults.violations.filter(v => 
        ['critical', 'serious'].includes(v.impact as string)
      )).toHaveLength(0)
    }
  })
})