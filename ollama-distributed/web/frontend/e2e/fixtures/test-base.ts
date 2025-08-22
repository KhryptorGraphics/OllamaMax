import { test as base, expect } from '@playwright/test';
import { injectAxe, checkA11y, getViolations } from '@axe-core/playwright';

/**
 * Custom test fixture with enhanced capabilities
 */
export const test = base.extend({
  // Auto-inject axe for accessibility testing
  page: async ({ page }, use) => {
    await injectAxe(page);
    await use(page);
  },
});

export { expect };

/**
 * Helper functions for common test operations
 */
export class TestHelpers {
  /**
   * Check accessibility violations
   */
  static async checkAccessibility(page: any, context?: string) {
    const violations = await getViolations(page);
    if (violations.length > 0) {
      console.error(`Accessibility violations found${context ? ` in ${context}` : ''}:`, violations);
    }
    expect(violations).toHaveLength(0);
  }

  /**
   * Wait for WebSocket connection
   */
  static async waitForWebSocket(page: any, url: string) {
    return page.waitForFunction(
      (wsUrl) => {
        return new Promise((resolve) => {
          const ws = new WebSocket(wsUrl);
          ws.onopen = () => {
            ws.close();
            resolve(true);
          };
          ws.onerror = () => resolve(false);
        });
      },
      url,
      { timeout: 10000 }
    );
  }

  /**
   * Mock API response
   */
  static async mockApiResponse(page: any, endpoint: string, response: any) {
    await page.route(`**/api/${endpoint}`, (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(response),
      });
    });
  }

  /**
   * Measure performance metrics
   */
  static async measurePerformance(page: any) {
    const metrics = await page.evaluate(() => {
      const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return {
        domContentLoaded: perfData.domContentLoadedEventEnd - perfData.domContentLoadedEventStart,
        loadComplete: perfData.loadEventEnd - perfData.loadEventStart,
        firstPaint: performance.getEntriesByName('first-paint')[0]?.startTime || 0,
        firstContentfulPaint: performance.getEntriesByName('first-contentful-paint')[0]?.startTime || 0,
      };
    });
    return metrics;
  }

  /**
   * Test responsive design
   */
  static async testResponsive(page: any, breakpoints: { name: string; width: number; height: number }[]) {
    for (const breakpoint of breakpoints) {
      await page.setViewportSize({ width: breakpoint.width, height: breakpoint.height });
      await page.waitForTimeout(500); // Allow layout to stabilize
    }
  }
}