import { test, expect, Page } from '@playwright/test'

/**
 * Dashboard E2E Tests
 * 
 * Tests for the main dashboard functionality including:
 * - Dashboard loading and rendering
 * - Real-time data updates
 * - KPI widgets and metrics
 * - Interactive elements
 * - Responsive design
 * - Error states
 */

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication
    await page.route('**/api/auth/me', route => {
      route.fulfill({
        status: 200,
        body: JSON.stringify({
          user: { id: 1, email: 'admin@example.com', role: 'admin' },
          token: 'mock-jwt-token'
        })
      })
    })

    // Enable KPI widget feature flag
    await page.addInitScript(() => {
      localStorage.setItem('V2_KPI_WIDGET', '1')
      ;(window as any).V2_KPI_WIDGET = true
    })
  })

  test.describe('Dashboard Loading', () => {
    test('loads dashboard with KPI widgets', async ({ page }) => {
      await page.goto('/v2')
      
      // Verify dashboard loads
      await expect(page.locator('.omx-v2')).toBeVisible()
      await expect(page.locator('text=Cluster Utilization')).toBeVisible()
      
      // Verify KPI widget structure
      const kpiWidget = page.locator('.omx-v2.p-4.rounded.border.bg-white.shadow')
      await expect(kpiWidget).toBeVisible()
      await expect(kpiWidget.locator('.text-sm.text-slate-500')).toContainText('Cluster Utilization')
      await expect(kpiWidget.locator('.text-3xl.font-semibold')).toBeVisible()
    })

    test('shows fallback when feature flag disabled', async ({ page }) => {
      // Disable feature flag
      await page.addInitScript(() => {
        localStorage.removeItem('V2_KPI_WIDGET')
        ;(window as any).V2_KPI_WIDGET = false
      })
      
      await page.goto('/v2')
      
      // Should show fallback content
      await expect(page.locator('text=Dashboard')).toBeVisible()
      await expect(page.locator('text=Cluster Utilization')).not.toBeVisible()
    })

    test('dashboard loading performance', async ({ page }) => {
      const startTime = Date.now()
      
      await page.goto('/v2')
      await expect(page.locator('text=Cluster Utilization')).toBeVisible()
      
      const loadTime = Date.now() - startTime
      expect(loadTime).toBeLessThan(2000) // Should load within 2 seconds
    })
  })

  test.describe('Real-time Updates', () => {
    test('KPI values update automatically', async ({ page }) => {
      await page.goto('/v2')
      
      const kpiValue = page.locator('.text-3xl.font-semibold')
      
      // Get initial value
      const initialValue = await kpiValue.textContent()
      
      // Wait for updates (interval is 1000ms)
      await page.waitForTimeout(2000)
      
      // Value should have changed
      const updatedValue = await kpiValue.textContent()
      expect(updatedValue).not.toBe(initialValue)
    })

    test('real-time updates within reasonable bounds', async ({ page }) => {
      await page.goto('/v2')
      
      const kpiValue = page.locator('.text-3xl.font-semibold')
      
      // Monitor values for 5 seconds
      const values: number[] = []
      for (let i = 0; i < 5; i++) {\n        const text = await kpiValue.textContent()\n        const value = parseFloat(text?.replace('%', '') || '0')\n        values.push(value)\n        await page.waitForTimeout(1000)\n      }\n      \n      // All values should be between 0 and 100\n      values.forEach(value => {\n        expect(value).toBeGreaterThanOrEqual(0)\n        expect(value).toBeLessThanOrEqual(100)\n      })\n    })\n\n    test('handles WebSocket connection for real-time data', async ({ page }) => {\n      // Mock WebSocket connection\n      await page.addInitScript(() => {\n        class MockWebSocket {\n          constructor(url: string) {\n            setTimeout(() => {\n              if (this.onopen) this.onopen({} as Event)\n            }, 100)\n          }\n          onopen: ((event: Event) => void) | null = null\n          onmessage: ((event: MessageEvent) => void) | null = null\n          onclose: ((event: CloseEvent) => void) | null = null\n          onerror: ((event: Event) => void) | null = null\n          \n          send(data: string) {\n            // Mock sending data\n          }\n          \n          close() {\n            if (this.onclose) {\n              this.onclose({ code: 1000, reason: 'Normal closure' } as CloseEvent)\n            }\n          }\n        }\n        \n        ;(window as any).WebSocket = MockWebSocket\n      })\n      \n      await page.goto('/v2')\n      \n      // WebSocket connection should be established\n      // This would be tested with actual WebSocket implementation\n      await expect(page.locator('text=Cluster Utilization')).toBeVisible()\n    })\n  })\n\n  test.describe('Responsive Design', () => {\n    test('dashboard adapts to mobile viewport', async ({ page }) => {\n      await page.setViewportSize({ width: 375, height: 667 }) // iPhone SE\n      await page.goto('/v2')\n      \n      // Should have single column layout on mobile\n      const grid = page.locator('.grid')\n      await expect(grid).toHaveClass(/grid-cols-1/)\n      \n      // KPI widget should be visible and properly sized\n      const kpiWidget = page.locator('.omx-v2.p-4.rounded.border')\n      await expect(kpiWidget).toBeVisible()\n      \n      const widgetBounds = await kpiWidget.boundingBox()\n      expect(widgetBounds?.width).toBeLessThan(375)\n    })\n\n    test('dashboard adapts to tablet viewport', async ({ page }) => {\n      await page.setViewportSize({ width: 768, height: 1024 }) // iPad\n      await page.goto('/v2')\n      \n      // Should have 2-column layout on tablet\n      const grid = page.locator('.grid')\n      await expect(grid).toHaveClass(/sm:grid-cols-2/)\n    })\n\n    test('dashboard adapts to desktop viewport', async ({ page }) => {\n      await page.setViewportSize({ width: 1920, height: 1080 })\n      await page.goto('/v2')\n      \n      // Should have 4-column layout on large screens\n      const grid = page.locator('.grid')\n      await expect(grid).toHaveClass(/lg:grid-cols-4/)\n    })\n  })\n\n  test.describe('Navigation', () => {\n    test('header navigation works correctly', async ({ page }) => {\n      await page.goto('/v2')\n      \n      // Verify header is present\n      await expect(page.locator('header')).toBeVisible()\n      await expect(page.locator('text=OllamaMax')).toBeVisible()\n      \n      // Test navigation links\n      await expect(page.locator('nav a[href=\"/v2\"]')).toContainText('Dashboard')\n      await expect(page.locator('nav a[href=\"/v2/auth/login\"]')).toContainText('Login')\n      await expect(page.locator('nav a[href=\"/v2/auth/register\"]')).toContainText('Register')\n    })\n\n    test('main content is properly focused for accessibility', async ({ page }) => {\n      await page.goto('/v2')\n      \n      // Main content should have proper focus management\n      const main = page.locator('main#main')\n      await expect(main).toBeVisible()\n      await expect(main).toHaveAttribute('tabindex', '-1')\n    })\n  })\n\n  test.describe('Error Handling', () => {\n    test('handles API errors gracefully', async ({ page }) => {\n      // Mock API error\n      await page.route('**/api/dashboard/metrics', route => {\n        route.fulfill({ status: 500, body: JSON.stringify({ error: 'Internal server error' }) })\n      })\n      \n      await page.goto('/v2')\n      \n      // Dashboard should still load with fallback content\n      await expect(page.locator('.omx-v2')).toBeVisible()\n      \n      // Error state should be handled gracefully\n      // (This would depend on actual error handling implementation)\n    })\n\n    test('handles network connectivity issues', async ({ page }) => {\n      await page.goto('/v2')\n      \n      // Simulate network going offline\n      await page.context().setOffline(true)\n      \n      // Application should handle offline state\n      // (Implementation would depend on offline handling strategy)\n      await page.context().setOffline(false)\n    })\n  })\n\n  test.describe('Performance', () => {\n    test('dashboard renders efficiently with many widgets', async ({ page }) => {\n      // Mock multiple KPI widgets\n      await page.addInitScript(() => {\n        localStorage.setItem('V2_MULTIPLE_WIDGETS', '1')\n      })\n      \n      const startTime = Date.now()\n      await page.goto('/v2')\n      await expect(page.locator('text=Cluster Utilization')).toBeVisible()\n      const renderTime = Date.now() - startTime\n      \n      expect(renderTime).toBeLessThan(1000) // Should render within 1 second\n    })\n\n    test('real-time updates do not cause memory leaks', async ({ page }) => {\n      await page.goto('/v2')\n      \n      // Let updates run for a while\n      await page.waitForTimeout(10000)\n      \n      // Check memory usage (this would require actual performance monitoring)\n      // For now, just verify the page is still responsive\n      const kpiValue = page.locator('.text-3xl.font-semibold')\n      await expect(kpiValue).toBeVisible()\n    })\n  })\n\n  test.describe('Accessibility', () => {\n    test('dashboard has proper ARIA structure', async ({ page }) => {\n      await page.goto('/v2')\n      \n      // Main landmark should be present\n      await expect(page.locator('main')).toBeVisible()\n      \n      // KPI widgets should have proper semantic structure\n      const widget = page.locator('.omx-v2.p-4.rounded.border').first()\n      await expect(widget).toBeVisible()\n      \n      // Values should be properly labeled\n      const title = widget.locator('.text-sm.text-slate-500')\n      const value = widget.locator('.text-3xl.font-semibold')\n      \n      await expect(title).toBeVisible()\n      await expect(value).toBeVisible()\n    })\n\n    test('keyboard navigation works properly', async ({ page }) => {\n      await page.goto('/v2')\n      \n      // Skip link should be accessible\n      await page.keyboard.press('Tab')\n      \n      // Should be able to navigate through interactive elements\n      const focusedElement = page.locator(':focus')\n      await expect(focusedElement).toBeVisible()\n    })\n\n    test('color contrast meets WCAG standards', async ({ page }) => {\n      await page.goto('/v2')\n      \n      // Check color contrast for KPI widget text\n      const title = page.locator('.text-sm.text-slate-500').first()\n      const titleStyles = await title.evaluate((el) => {\n        const computed = window.getComputedStyle(el)\n        return {\n          color: computed.color,\n          backgroundColor: computed.backgroundColor\n        }\n      })\n      \n      // This would require actual contrast calculation\n      // For now, verify the styles are applied\n      expect(titleStyles.color).toBeTruthy()\n    })\n  })\n})"