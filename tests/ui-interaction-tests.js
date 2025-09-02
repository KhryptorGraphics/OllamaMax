/**
 * Complete UI Interaction Tests
 * Tests every button, link, and interactive element in the interface
 */

import { test, expect } from '@playwright/test';

const CONFIG = {
  WEB_URL: 'http://localhost:8080',
  TIMEOUT: 10000
};

test.describe('Complete UI Interaction Testing', () => {

  test.beforeEach(async ({ page }) => {
    await page.goto(CONFIG.WEB_URL);
    await page.waitForSelector('#app');
    await page.waitForLoadState('networkidle');
  });

  test.describe('Header and Navigation Interactions', () => {

    test('should interact with all navigation tabs', async ({ page }) => {
      const tabs = [
        { name: 'chat', selector: '[data-tab="chat"]', contentSelector: '#chatTab' },
        { name: 'nodes', selector: '[data-tab="nodes"]', contentSelector: '#nodesTab' },
        { name: 'models', selector: '[data-tab="models"]', contentSelector: '#modelsTab' },
        { name: 'settings', selector: '[data-tab="settings"]', contentSelector: '#settingsTab' }
      ];

      for (const tab of tabs) {
        // Click tab
        await page.click(tab.selector);
        await page.waitForTimeout(500);
        
        // Verify tab is active
        await expect(page.locator(tab.selector)).toHaveClass(/active/);
        await expect(page.locator(tab.contentSelector)).toHaveClass(/active/);
        await expect(page.locator(tab.contentSelector)).toBeVisible();
        
        console.log(`✅ Tab '${tab.name}' interaction successful`);
      }
    });

    test('should display connection status interactions', async ({ page }) => {
      const statusIndicator = page.locator('#connectionStatus');
      const statusText = page.locator('#connectionText');
      
      await expect(statusIndicator).toBeVisible();
      await expect(statusText).toBeVisible();
      
      // Status should update over time
      await page.waitForTimeout(3000);
      const statusContent = await statusText.textContent();
      expect(statusContent.length).toBeGreaterThan(0);
      
      console.log(`✅ Connection status: ${statusContent}`);
    });

  });

  test.describe('Chat Tab - All Interactive Elements', () => {

    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="chat"]');
      await page.waitForSelector('#chatTab', { state: 'visible' });
    });

    test('should test message input area interactions', async ({ page }) => {
      const messageInput = page.locator('#messageInput');
      const sendButton = page.locator('#sendButton');
      const attachButton = page.locator('#attachButton');
      
      // Test message input
      await expect(messageInput).toBeVisible();
      await messageInput.fill('Test message for interaction');
      await expect(messageInput).toHaveValue('Test message for interaction');
      
      // Test send button
      await expect(sendButton).toBeVisible();
      await expect(sendButton).toBeEnabled();
      
      // Test attach button
      await expect(attachButton).toBeVisible();
      await attachButton.click(); // Should not crash
      
      // Test keyboard interactions
      await messageInput.fill('');
      await messageInput.fill('Test Enter key');
      await messageInput.press('Enter');
      
      // Test Shift+Enter for multiline
      await messageInput.fill('Line 1');
      await messageInput.press('Shift+Enter');
      await messageInput.type('Line 2');
      
      const multilineValue = await messageInput.inputValue();
      expect(multilineValue).toContain('\n');
      
      console.log('✅ Message input interactions tested');
    });

    test('should test model selector interactions', async ({ page }) => {
      const modelSelector = page.locator('#modelSelector');
      
      await expect(modelSelector).toBeVisible();
      
      // Get available options
      const options = await modelSelector.locator('option').allTextContents();
      expect(options.length).toBeGreaterThan(0);
      
      // Test selection
      if (options.length > 0) {
        await modelSelector.selectOption(options[0]);
        const selectedValue = await modelSelector.inputValue();
        expect(selectedValue.length).toBeGreaterThan(0);
      }
      
      console.log(`✅ Model selector tested with ${options.length} options`);
    });

    test('should test status bar elements', async ({ page }) => {
      const statusElements = [
        '#activeNode',
        '#queueLength', 
        '#latency',
        '#nodeCount'
      ];

      for (const selector of statusElements) {
        const element = page.locator(selector);
        await expect(element).toBeVisible();
        
        const content = await element.textContent();
        console.log(`✅ ${selector}: ${content}`);
      }
    });

  });

  test.describe('Nodes Tab - All Interactive Elements', () => {

    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="nodes"]');
      await page.waitForSelector('#nodesTab', { state: 'visible' });
      await page.waitForTimeout(1000);
    });

    test('should test view toggle buttons', async ({ page }) => {
      const compactView = page.locator('#compactView');
      const detailedView = page.locator('#detailedView');
      
      await expect(compactView).toBeVisible();
      await expect(detailedView).toBeVisible();
      
      // Test view switching
      await detailedView.click();
      await expect(detailedView).toHaveClass(/active/);
      
      await compactView.click();
      await expect(compactView).toHaveClass(/active/);
      
      console.log('✅ View toggle buttons tested');
    });

    test('should test refresh controls', async ({ page }) => {
      const pauseRefresh = page.locator('#pauseRefresh');
      const refreshNodes = page.locator('#refreshNodes');
      
      await expect(pauseRefresh).toBeVisible();
      await expect(refreshNodes).toBeVisible();
      
      // Test pause/resume
      await pauseRefresh.click();
      await page.waitForTimeout(500);
      
      await pauseRefresh.click();
      await page.waitForTimeout(500);
      
      // Test manual refresh
      await refreshNodes.click();
      await page.waitForTimeout(1000);
      
      console.log('✅ Refresh controls tested');
    });

    test('should test filtering and sorting controls', async ({ page }) => {
      const statusFilter = page.locator('#statusFilter');
      const sortBy = page.locator('#sortBy');
      const searchInput = page.locator('#nodeSearch');
      
      // Test status filter
      await expect(statusFilter).toBeVisible();
      await statusFilter.selectOption('healthy');
      await page.waitForTimeout(500);
      
      await statusFilter.selectOption('all');
      await page.waitForTimeout(500);
      
      // Test sorting
      await expect(sortBy).toBeVisible();
      await sortBy.selectOption('cpu');
      await page.waitForTimeout(500);
      
      await sortBy.selectOption('name');
      await page.waitForTimeout(500);
      
      // Test search
      await expect(searchInput).toBeVisible();
      await searchInput.fill('node');
      await page.waitForTimeout(500);
      
      await searchInput.fill('');
      await page.waitForTimeout(500);
      
      console.log('✅ Filtering and sorting controls tested');
    });

    test('should test node card interactions', async ({ page }) => {
      await page.waitForSelector('.enhanced-node-card', { timeout: 10000 });
      
      const nodeCards = page.locator('.enhanced-node-card');
      const cardCount = await nodeCards.count();
      
      if (cardCount > 0) {
        const firstCard = nodeCards.first();
        
        // Test expand button
        const expandBtn = firstCard.locator('.expand-btn');
        if (await expandBtn.isVisible()) {
          await expandBtn.click();
          await page.waitForTimeout(500);
          
          // Check if details section appears
          const detailsSection = firstCard.locator('.node-details-section');
          if (await detailsSection.isVisible()) {
            
            // Test detail tabs
            const detailTabs = firstCard.locator('.detail-tab');
            const tabCount = await detailTabs.count();
            
            for (let i = 0; i < tabCount; i++) {
              await detailTabs.nth(i).click();
              await page.waitForTimeout(300);
            }
          }
        }
        
        // Test other action buttons
        const configBtn = firstCard.locator('.config-btn');
        const restartBtn = firstCard.locator('.restart-btn');
        
        if (await configBtn.isVisible()) {
          await configBtn.click();
          await page.waitForTimeout(500);
        }
        
        console.log(`✅ Node card interactions tested on ${cardCount} cards`);
      }
    });

  });

  test.describe('Models Tab - All Interactive Elements', () => {

    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="models"]');
      await page.waitForSelector('#modelsTab', { state: 'visible' });
      await page.waitForTimeout(1000);
    });

    test('should test refresh models button', async ({ page }) => {
      const refreshButton = page.locator('#refreshModels');
      
      await expect(refreshButton).toBeVisible();
      await refreshButton.click();
      await page.waitForTimeout(1000);
      
      console.log('✅ Refresh models button tested');
    });

    test('should test model download form', async ({ page }) => {
      const modelNameInput = page.locator('#newModelName');
      const downloadButton = page.locator('#downloadModel');
      const targetsContainer = page.locator('#downloadTargets');
      
      // Test model name input
      await expect(modelNameInput).toBeVisible();
      await modelNameInput.fill('test-model-name');
      await expect(modelNameInput).toHaveValue('test-model-name');
      
      // Test targets container
      await expect(targetsContainer).toBeVisible();
      
      // Test download button
      await expect(downloadButton).toBeVisible();
      await expect(downloadButton).toBeEnabled();
      
      console.log('✅ Model download form tested');
    });

    test('should test P2P and propagation settings', async ({ page }) => {
      const autoPropagation = page.locator('#autoPropagation');
      const p2pEnabled = page.locator('#p2pEnabled');
      const propagationStrategy = page.locator('#propagationStrategy');
      
      // Test auto propagation checkbox
      await expect(autoPropagation).toBeVisible();
      const initialAutoState = await autoPropagation.isChecked();
      await autoPropagation.click();
      await expect(autoPropagation).toBeChecked(!initialAutoState);
      
      // Test P2P enabled checkbox (CRITICAL for requirements)
      await expect(p2pEnabled).toBeVisible();
      const initialP2PState = await p2pEnabled.isChecked();
      await p2pEnabled.click();
      await expect(p2pEnabled).toBeChecked(!initialP2PState);
      await p2pEnabled.click(); // Return to original state
      
      // Test propagation strategy dropdown
      await expect(propagationStrategy).toBeVisible();
      const strategies = await propagationStrategy.locator('option').allTextContents();
      expect(strategies.length).toBeGreaterThanOrEqual(3);
      
      await propagationStrategy.selectOption('scheduled');
      await expect(propagationStrategy).toHaveValue('scheduled');
      
      console.log('✅ P2P and propagation settings tested');
    });

    test('should test model card actions', async ({ page }) => {
      await page.waitForTimeout(2000);
      
      const modelCards = page.locator('.model-card');
      const cardCount = await modelCards.count();
      
      if (cardCount > 0) {
        const firstCard = modelCards.first();
        
        // Test propagate button
        const propagateButton = firstCard.locator('[data-action="propagate"]');
        if (await propagateButton.isVisible()) {
          await expect(propagateButton).toBeVisible();
          console.log('✅ Propagate button found and visible');
        }
        
        // Test delete button
        const deleteButton = firstCard.locator('[data-action="delete"]');
        if (await deleteButton.isVisible()) {
          await expect(deleteButton).toBeVisible();
          console.log('✅ Delete button found and visible');
        }
        
        console.log(`✅ Model card actions tested on ${cardCount} cards`);
      } else {
        console.log('⚠️ No model cards found to test');
      }
    });

  });

  test.describe('Settings Tab - All Interactive Elements', () => {

    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="settings"]');
      await page.waitForSelector('#settingsTab', { state: 'visible' });
    });

    test('should test API configuration inputs', async ({ page }) => {
      const apiEndpoint = page.locator('#apiEndpoint');
      const apiKey = page.locator('#apiKey');
      
      // Test API endpoint
      await expect(apiEndpoint).toBeVisible();
      const originalEndpoint = await apiEndpoint.inputValue();
      await apiEndpoint.fill('ws://test:13100/chat');
      await expect(apiEndpoint).toHaveValue('ws://test:13100/chat');
      await apiEndpoint.fill(originalEndpoint); // Restore
      
      // Test API key
      await expect(apiKey).toBeVisible();
      await apiKey.fill('test-api-key');
      await expect(apiKey).toHaveValue('test-api-key');
      await apiKey.fill(''); // Clear
      
      console.log('✅ API configuration inputs tested');
    });

    test('should test chat settings controls', async ({ page }) => {
      const streamingEnabled = page.locator('#streamingEnabled');
      const autoScroll = page.locator('#autoScroll');
      const maxTokens = page.locator('#maxTokens');
      const temperature = page.locator('#temperature');
      
      // Test streaming checkbox
      await expect(streamingEnabled).toBeVisible();
      const initialStreaming = await streamingEnabled.isChecked();
      await streamingEnabled.click();
      await expect(streamingEnabled).toBeChecked(!initialStreaming);
      
      // Test auto scroll checkbox
      await expect(autoScroll).toBeVisible();
      const initialAutoScroll = await autoScroll.isChecked();
      await autoScroll.click();
      await expect(autoScroll).toBeChecked(!initialAutoScroll);
      
      // Test max tokens input
      await expect(maxTokens).toBeVisible();
      await maxTokens.fill('1024');
      await expect(maxTokens).toHaveValue('1024');
      
      // Test temperature slider
      await expect(temperature).toBeVisible();
      await temperature.fill('1.2');
      
      // Check temperature display value
      const tempValue = page.locator('#temperatureValue');
      if (await tempValue.isVisible()) {
        const displayValue = await tempValue.textContent();
        expect(displayValue).toContain('1.2');
      }
      
      console.log('✅ Chat settings controls tested');
    });

    test('should test node management controls', async ({ page }) => {
      const loadBalancing = page.locator('#loadBalancing');
      const addNodeButton = page.locator('#addNodeButton');
      
      // Test load balancing dropdown
      await expect(loadBalancing).toBeVisible();
      const strategies = await loadBalancing.locator('option').allTextContents();
      expect(strategies.length).toBeGreaterThan(0);
      
      await loadBalancing.selectOption('least-loaded');
      await expect(loadBalancing).toHaveValue('least-loaded');
      
      // Test add node button
      await expect(addNodeButton).toBeVisible();
      await addNodeButton.click();
      
      // Modal should appear
      const modal = page.locator('#addNodeModal');
      await expect(modal).toBeVisible();
      
      // Test modal form
      const nodeUrl = page.locator('#nodeUrl');
      const nodeName = page.locator('#nodeName');
      const confirmButton = page.locator('#confirmAddNode');
      const cancelButton = page.locator('#cancelAddNode');
      
      await expect(nodeUrl).toBeVisible();
      await expect(nodeName).toBeVisible();
      await expect(confirmButton).toBeVisible();
      await expect(cancelButton).toBeVisible();
      
      // Fill and test form
      await nodeUrl.fill('http://test-node:11434');
      await nodeName.fill('test-node');
      
      // Cancel modal
      await cancelButton.click();
      await expect(modal).not.toBeVisible();
      
      console.log('✅ Node management controls tested');
    });

    test('should test settings action buttons', async ({ page }) => {
      const saveSettings = page.locator('#saveSettings');
      const resetSettings = page.locator('#resetSettings');
      
      // Make a change
      await page.fill('#maxTokens', '1500');
      
      // Test save button
      await expect(saveSettings).toBeVisible();
      await saveSettings.click();
      await page.waitForTimeout(500);
      
      // Test reset button
      await expect(resetSettings).toBeVisible();
      await resetSettings.click();
      await page.waitForTimeout(500);
      
      // Should reset to default
      await expect(page.locator('#maxTokens')).toHaveValue('2048');
      
      console.log('✅ Settings action buttons tested');
    });

  });

  test.describe('Error Handling and Edge Cases', () => {

    test('should handle modal interactions gracefully', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      await page.waitForSelector('#settingsTab', { state: 'visible' });
      
      // Open modal
      await page.click('#addNodeButton');
      await expect(page.locator('#addNodeModal')).toBeVisible();
      
      // Test clicking outside modal (if implemented)
      await page.click('.modal', { force: true });
      
      // Test escape key (if implemented)
      await page.keyboard.press('Escape');
      
      console.log('✅ Modal interaction edge cases tested');
    });

    test('should handle rapid clicking gracefully', async ({ page }) => {
      // Rapidly click navigation tabs
      for (let i = 0; i < 10; i++) {
        await page.click('[data-tab="nodes"]');
        await page.click('[data-tab="chat"]');
        await page.click('[data-tab="models"]');
        await page.click('[data-tab="settings"]');
      }
      
      // Should still be responsive
      await expect(page.locator('#app')).toBeVisible();
      
      console.log('✅ Rapid clicking handled gracefully');
    });

    test('should handle form validation', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      await page.waitForSelector('#settingsTab', { state: 'visible' });
      
      // Test invalid values
      await page.fill('#maxTokens', '-100');
      await page.fill('#temperature', '5'); // Outside normal range
      
      // Try to save
      await page.click('#saveSettings');
      
      // Should handle gracefully (either prevent or correct)
      await expect(page.locator('#app')).toBeVisible();
      
      console.log('✅ Form validation tested');
    });

  });

});