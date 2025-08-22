import { test, expect, TestHelpers } from './fixtures/test-base';

test.describe('Cluster Dashboard E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses
    await TestHelpers.mockApiResponse(page, 'nodes', {
      nodes: [
        {
          id: 'node-1',
          name: 'Node 1',
          status: 'healthy',
          cpu: 45,
          memory: 60,
          storage: 75,
          models: ['llama2', 'mistral'],
        },
        {
          id: 'node-2',
          name: 'Node 2',
          status: 'healthy',
          cpu: 30,
          memory: 40,
          storage: 50,
          models: ['gpt4'],
        },
      ],
    });

    await page.goto('/dashboard');
  });

  test('should display cluster overview', async ({ page }) => {
    // Check main dashboard elements
    await expect(page.locator('h1')).toContainText('Cluster Dashboard');
    await expect(page.locator('[data-testid="node-count"]')).toContainText('2');
    await expect(page.locator('[data-testid="healthy-nodes"]')).toContainText('2');
    
    // Check node cards
    const nodeCards = page.locator('[data-testid="node-card"]');
    await expect(nodeCards).toHaveCount(2);
  });

  test('should handle node selection', async ({ page }) => {
    // Click on first node
    await page.click('[data-testid="node-card"]:first-child');
    
    // Check if details panel opens
    await expect(page.locator('[data-testid="node-details"]')).toBeVisible();
    await expect(page.locator('[data-testid="node-details-title"]')).toContainText('Node 1');
    
    // Check resource metrics
    await expect(page.locator('[data-testid="cpu-usage"]')).toContainText('45%');
    await expect(page.locator('[data-testid="memory-usage"]')).toContainText('60%');
    await expect(page.locator('[data-testid="storage-usage"]')).toContainText('75%');
  });

  test('should update in real-time via WebSocket', async ({ page }) => {
    // Wait for WebSocket connection
    await TestHelpers.waitForWebSocket(page, 'ws://localhost:8080/ws');
    
    // Simulate WebSocket message
    await page.evaluate(() => {
      const ws = (window as any).__websocket;
      if (ws) {
        ws.dispatchEvent(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'node-update',
            data: {
              id: 'node-1',
              cpu: 80,
              memory: 90,
            },
          }),
        }));
      }
    });
    
    // Check if UI updates
    await expect(page.locator('[data-testid="node-1-cpu"]')).toContainText('80%');
    await expect(page.locator('[data-testid="node-1-memory"]')).toContainText('90%');
  });

  test('should be accessible', async ({ page }) => {
    await TestHelpers.checkAccessibility(page, 'Cluster Dashboard');
  });

  test('should be responsive', async ({ page }) => {
    const breakpoints = [
      { name: 'mobile', width: 375, height: 667 },
      { name: 'tablet', width: 768, height: 1024 },
      { name: 'desktop', width: 1920, height: 1080 },
    ];
    
    await TestHelpers.testResponsive(page, breakpoints);
    
    // Check layout changes
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.locator('[data-testid="mobile-menu"]')).toBeVisible();
    
    await page.setViewportSize({ width: 1920, height: 1080 });
    await expect(page.locator('[data-testid="desktop-sidebar"]')).toBeVisible();
  });

  test('should measure performance metrics', async ({ page }) => {
    const metrics = await TestHelpers.measurePerformance(page);
    
    // Assert performance budgets
    expect(metrics.domContentLoaded).toBeLessThan(1000);
    expect(metrics.loadComplete).toBeLessThan(3000);
    expect(metrics.firstContentfulPaint).toBeLessThan(1500);
  });
});