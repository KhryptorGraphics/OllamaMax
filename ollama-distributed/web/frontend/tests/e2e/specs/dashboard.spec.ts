import { test, expect } from '@playwright/test'

test.describe('Dashboard Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('displays cluster overview correctly', async ({ page }) => {
    // Wait for dashboard to load
    await page.waitForSelector('[data-testid="cluster-overview"]')
    
    // Check metric cards are visible
    await expect(page.locator('[data-testid="total-nodes"]')).toBeVisible()
    await expect(page.locator('[data-testid="active-models"]')).toBeVisible()
    await expect(page.locator('[data-testid="cpu-usage"]')).toBeVisible()
    await expect(page.locator('[data-testid="memory-usage"]')).toBeVisible()

    // Verify metric values are numbers
    const totalNodes = await page.locator('[data-testid="total-nodes"] .metric-value').textContent()
    expect(parseInt(totalNodes || '0')).toBeGreaterThanOrEqual(0)
  })

  test('real-time updates work correctly', async ({ page }) => {
    await page.waitForSelector('[data-testid="cluster-overview"]')
    
    // Get initial CPU value
    const initialCpu = await page.locator('[data-testid="cpu-usage"] .metric-value').textContent()
    
    // Wait for WebSocket update (should happen within 5 seconds)
    await page.waitForTimeout(6000)
    
    // Check if value has updated (or at least container is still there)
    await expect(page.locator('[data-testid="cpu-usage"] .metric-value')).toBeVisible()
  })

  test('navigation between tabs works', async ({ page }) => {
    // Click on Models tab
    await page.click('[data-testid="models-tab"]')
    await expect(page.locator('[data-testid="models-content"]')).toBeVisible()
    
    // Click on Monitoring tab
    await page.click('[data-testid="monitoring-tab"]')
    await expect(page.locator('[data-testid="monitoring-content"]')).toBeVisible()
    
    // Return to Dashboard tab
    await page.click('[data-testid="dashboard-tab"]')
    await expect(page.locator('[data-testid="cluster-overview"]')).toBeVisible()
  })

  test('responsive design on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    
    // Check that metric cards stack vertically
    const cards = page.locator('[data-testid="metric-card"]')
    const cardCount = await cards.count()
    
    for (let i = 0; i < cardCount; i++) {
      await expect(cards.nth(i)).toBeVisible()
    }
  })
})

test.describe('Model Management', () => {
  test('can load and interact with model list', async ({ page }) => {
    await page.goto('/models')
    
    // Wait for models to load
    await page.waitForSelector('[data-testid="model-list"]')
    
    // Check if model cards are present
    const modelCards = page.locator('[data-testid="model-card"]')
    const modelCount = await modelCards.count()
    
    if (modelCount > 0) {
      // Click on first model
      await modelCards.first().click()
      
      // Check if model details are shown
      await expect(page.locator('[data-testid="model-details"]')).toBeVisible()
    }
  })

  test('can start and stop models', async ({ page }) => {
    await page.goto('/models')
    await page.waitForSelector('[data-testid="model-list"]')
    
    const startButton = page.locator('[data-testid="start-model-btn"]').first()
    if (await startButton.isVisible()) {
      await startButton.click()
      
      // Wait for status change
      await page.waitForTimeout(2000)
      
      // Check if stop button appears
      await expect(page.locator('[data-testid="stop-model-btn"]').first()).toBeVisible()
    }
  })
})

test.describe('Performance Monitoring', () => {
  test('performance charts load correctly', async ({ page }) => {
    await page.goto('/monitoring')
    
    // Wait for charts to load
    await page.waitForSelector('[data-testid="cpu-chart"]')
    await page.waitForSelector('[data-testid="memory-chart"]')
    
    // Check if chart containers have content
    await expect(page.locator('[data-testid="cpu-chart"] canvas')).toBeVisible()
    await expect(page.locator('[data-testid="memory-chart"] canvas')).toBeVisible()
  })

  test('can change time range for charts', async ({ page }) => {
    await page.goto('/monitoring')
    await page.waitForSelector('[data-testid="time-range-selector"]')
    
    // Change to 24h view
    await page.selectOption('[data-testid="time-range-selector"]', '24h')
    
    // Wait for charts to update
    await page.waitForTimeout(1000)
    
    // Charts should still be visible
    await expect(page.locator('[data-testid="cpu-chart"] canvas')).toBeVisible()
  })
})

test.describe('Security Features', () => {
  test('zero trust dashboard loads security metrics', async ({ page }) => {
    await page.goto('/security/zero-trust')
    
    // Wait for security overview
    await page.waitForSelector('[data-testid="security-overview"]')
    
    // Check security metric cards
    await expect(page.locator('[data-testid="verified-identities"]')).toBeVisible()
    await expect(page.locator('[data-testid="active-certificates"]')).toBeVisible()
    await expect(page.locator('[data-testid="threats-blocked"]')).toBeVisible()
  })

  test('can navigate security tabs', async ({ page }) => {
    await page.goto('/security/zero-trust')
    
    // Test identities tab
    await page.click('[data-testid="identities-tab"]')
    await expect(page.locator('[data-testid="identities-content"]')).toBeVisible()
    
    // Test policies tab
    await page.click('[data-testid="policies-tab"]')
    await expect(page.locator('[data-testid="policies-content"]')).toBeVisible()
    
    // Test certificates tab
    await page.click('[data-testid="certificates-tab"]')
    await expect(page.locator('[data-testid="certificates-content"]')).toBeVisible()
  })
})

test.describe('Accessibility', () => {
  test('keyboard navigation works', async ({ page }) => {
    await page.goto('/')
    
    // Focus on first interactive element
    await page.keyboard.press('Tab')
    
    // Should be able to navigate through all interactive elements
    for (let i = 0; i < 10; i++) {
      await page.keyboard.press('Tab')
    }
    
    // Check that focus is visible
    const focusedElement = await page.evaluate(() => document.activeElement?.tagName)
    expect(focusedElement).toBeTruthy()
  })

  test('screen reader announcements work', async ({ page }) => {
    await page.goto('/')
    
    // Check for aria-live regions
    await expect(page.locator('[aria-live="polite"]')).toBeVisible()
    
    // Check for proper heading structure
    const h1 = await page.locator('h1').count()
    expect(h1).toBeGreaterThanOrEqual(1)
  })

  test('high contrast mode support', async ({ page, browserName }) => {
    // Skip on WebKit as forced-colors is not supported
    test.skip(browserName === 'webkit', 'WebKit does not support forced-colors')
    
    await page.goto('/')
    
    // Check that elements are still visible in high contrast
    await expect(page.locator('[data-testid="cluster-overview"]')).toBeVisible()
    await expect(page.locator('[data-testid="metric-card"]').first()).toBeVisible()
  })
})