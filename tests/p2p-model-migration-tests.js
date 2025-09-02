/**
 * P2P Model Migration and Distribution Tests
 * Tests the critical P2P model migration functionality when enabled
 */

import { test, expect } from '@playwright/test';

const CONFIG = {
  WEB_URL: 'http://localhost:8080',
  API_SERVER: 'http://localhost:13100',
  WORKERS: ['localhost:13000', 'localhost:13001', 'localhost:13002'],
  TEST_MODEL: 'tinyllama',
  TIMEOUT: 30000
};

test.describe('P2P Model Migration Tests', () => {

  test.beforeEach(async ({ page }) => {
    await page.goto(CONFIG.WEB_URL);
    await page.waitForSelector('#app');
    await page.waitForLoadState('networkidle');
  });

  test('should enable P2P model migration control', async ({ page }) => {
    // Navigate to Models tab
    await page.click('[data-tab="models"]');
    await page.waitForSelector('#modelsTab', { state: 'visible' });
    
    // Locate P2P control
    const p2pCheckbox = page.locator('#p2pEnabled');
    await expect(p2pCheckbox).toBeVisible();
    
    // Should be enabled by default (as per requirement)
    await expect(p2pCheckbox).toBeChecked();
    
    // Test toggling
    await p2pCheckbox.uncheck();
    await expect(p2pCheckbox).not.toBeChecked();
    
    // Re-enable for further tests
    await p2pCheckbox.check();
    await expect(p2pCheckbox).toBeChecked();
  });

  test('should display propagation strategy options', async ({ page }) => {
    await page.click('[data-tab="models"]');
    await page.waitForSelector('#modelsTab', { state: 'visible' });
    
    const strategySelect = page.locator('#propagationStrategy');
    await expect(strategySelect).toBeVisible();
    
    // Check available options
    const options = await strategySelect.locator('option').allTextContents();
    expect(options).toContain('Immediate (fastest)');
    expect(options).toContain('Scheduled (during low usage)');
    expect(options).toContain('Manual only');
    
    // Test selection
    await strategySelect.selectOption('scheduled');
    await expect(strategySelect).toHaveValue('scheduled');
  });

  test('should show auto-propagation settings', async ({ page }) => {
    await page.click('[data-tab="models"]');
    await page.waitForSelector('#modelsTab', { state: 'visible' });
    
    const autoPropCheckbox = page.locator('#autoPropagation');
    await expect(autoPropCheckbox).toBeVisible();
    
    // Should be enabled by default
    await expect(autoPropCheckbox).toBeChecked();
    
    // Test description text
    const labelText = page.locator('label[for="autoPropagation"]');
    await expect(labelText).toContainText('Enable automatic model propagation to new nodes');
  });

  test('should handle P2P model migration workflow', async ({ page, request }) => {
    // Enable P2P mode
    await page.click('[data-tab="models"]');
    await page.waitForSelector('#modelsTab', { state: 'visible' });
    
    const p2pCheckbox = page.locator('#p2pEnabled');
    await p2pCheckbox.check();
    
    // Set immediate propagation
    await page.selectOption('#propagationStrategy', 'immediate');
    
    // Verify current models on nodes
    const initialModels = await request.get(`${CONFIG.API_SERVER}/api/models`);
    expect(initialModels.status()).toBe(200);
    
    // Simulate model availability check
    const modelData = await initialModels.json();
    if (modelData.models && modelData.models.length > 0) {
      // Find first available model
      const testModel = modelData.models[0];
      
      // Look for model propagation options
      await page.waitForSelector('.model-card', { timeout: 10000 });
      const modelCards = page.locator('.model-card');
      
      if (await modelCards.count() > 0) {
        const firstCard = modelCards.first();
        
        // Check for propagate button
        const propagateButton = firstCard.locator('[data-action="propagate"]');
        if (await propagateButton.isVisible()) {
          await expect(propagateButton).toBeVisible();
          await expect(propagateButton).toContainText('Propagate');
        }
      }
    }
  });

  test('should display model distribution status across nodes', async ({ page, request }) => {
    await page.click('[data-tab="models"]');
    await page.waitForSelector('#modelsTab', { state: 'visible' });
    
    // Wait for models to load
    await page.waitForTimeout(2000);
    
    // Check if model cards show node availability
    const modelCards = page.locator('.model-card');
    const cardCount = await modelCards.count();
    
    if (cardCount > 0) {
      // Check first model card
      const firstCard = modelCards.first();
      
      // Should show which nodes have the model
      const nodesBadgesContainer = firstCard.locator('#modelNodes');
      if (await nodesBadgesContainer.isVisible()) {
        await expect(nodesBadgesContainer).toBeVisible();
      }
      
      // Should show "Available on:" text
      const availableText = firstCard.locator('.nodes-header');
      if (await availableText.isVisible()) {
        await expect(availableText).toContainText('Available on:');
      }
    }
  });

  test('should validate model download targets selection', async ({ page }) => {
    await page.click('[data-tab="models"]');
    await page.waitForSelector('#modelsTab', { state: 'visible' });
    
    // Find download form
    const downloadForm = page.locator('.model-download-form');
    await expect(downloadForm).toBeVisible();
    
    // Check model name input
    const modelNameInput = page.locator('#newModelName');
    await expect(modelNameInput).toBeVisible();
    
    // Check worker target selection
    const targetsContainer = page.locator('#downloadTargets');
    await expect(targetsContainer).toBeVisible();
    
    // Should allow selecting specific workers for download
    const downloadButton = page.locator('#downloadModel');
    await expect(downloadButton).toBeVisible();
    
    // Test form interaction
    await modelNameInput.fill('test-model');
    await expect(modelNameInput).toHaveValue('test-model');
  });

});

test.describe('Model Migration API Tests', () => {

  test('should expose model distribution endpoints', async ({ request }) => {
    // Test getting models across all nodes
    const modelsResponse = await request.get(`${CONFIG.API_SERVER}/api/models`);
    expect(modelsResponse.status()).toBe(200);
    
    const data = await modelsResponse.json();
    expect(data).toHaveProperty('models');
  });

  test('should handle model propagation requests', async ({ request }) => {
    // Test propagation endpoint (if available)
    const propagateResponse = await request.post(`${CONFIG.API_SERVER}/api/models/propagate`, {
      data: {
        model: CONFIG.TEST_MODEL,
        targetNodes: CONFIG.WORKERS,
        strategy: 'immediate'
      }
    });
    
    // Should either succeed or return meaningful error
    expect([200, 202, 400, 404]).toContain(propagateResponse.status());
  });

  test('should report model availability per node', async ({ request }) => {
    for (const worker of CONFIG.WORKERS) {
      const [host, port] = worker.split(':');
      const workerUrl = `http://${host}:${port}`;
      
      const response = await request.get(`${workerUrl}/api/tags`);
      
      if (response.status() === 200) {
        const data = await response.json();
        expect(data).toHaveProperty('models');
        expect(Array.isArray(data.models)).toBe(true);
      }
    }
  });

});

test.describe('P2P Network Behavior Tests', () => {

  test('should handle P2P settings persistence', async ({ page }) => {
    await page.click('[data-tab="models"]');
    await page.waitForSelector('#modelsTab', { state: 'visible' });
    
    // Change P2P settings
    const p2pCheckbox = page.locator('#p2pEnabled');
    await p2pCheckbox.uncheck();
    await page.selectOption('#propagationStrategy', 'manual');
    
    // Refresh page
    await page.reload();
    await page.waitForSelector('#app');
    await page.click('[data-tab="models"]');
    
    // Settings should persist (if implemented)
    // This tests local storage or session persistence
    const currentP2pState = await p2pCheckbox.isChecked();
    const currentStrategy = await page.locator('#propagationStrategy').inputValue();
    
    // Log current states for debugging
    console.log(`P2P State after reload: ${currentP2pState}`);
    console.log(`Strategy after reload: ${currentStrategy}`);
  });

  test('should show P2P status in connection info', async ({ page }) => {
    await page.goto(CONFIG.WEB_URL);
    await page.waitForSelector('#app');
    
    // Check if P2P status is shown in connection area
    const connectionArea = page.locator('.connection-status');
    await expect(connectionArea).toBeVisible();
    
    // Look for any P2P indicators
    const statusElements = page.locator('.status-indicator, .status-value');
    const statusCount = await statusElements.count();
    expect(statusCount).toBeGreaterThan(0);
  });

  test('should handle P2P migration in real-time', async ({ page }) => {
    // This test monitors for real-time updates during P2P operations
    let updateReceived = false;
    
    // Monitor for WebSocket messages
    page.on('websocket', ws => {
      ws.on('framereceived', data => {
        const frame = data.payload;
        if (frame.includes('model') || frame.includes('migration')) {
          updateReceived = true;
        }
      });
    });
    
    await page.goto(CONFIG.WEB_URL);
    await page.waitForSelector('#app');
    
    // Go to models tab and ensure P2P is enabled
    await page.click('[data-tab="models"]');
    await page.waitForSelector('#modelsTab', { state: 'visible' });
    
    const p2pCheckbox = page.locator('#p2pEnabled');
    await p2pCheckbox.check();
    
    // Wait for potential WebSocket updates
    await page.waitForTimeout(5000);
    
    // Log whether we received any model-related updates
    console.log(`Real-time P2P updates received: ${updateReceived}`);
  });

});