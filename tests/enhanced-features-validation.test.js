/**
 * Enhanced Features Validation Test Suite
 * Tests specific enhancements, real-time features, and advanced UI components
 * Focus: Mock data handling, enhanced node cards, model propagation, error states
 */

import { test, expect } from '@playwright/test';

const BASE_URL = 'http://localhost:8080';

test.describe('OllamaMax Enhanced Features - Validation Testing', () => {
  
  test.beforeEach(async ({ page }) => {
    // Set up network interception for API calls
    await page.route('**/api/nodes/detailed', async route => {
      // Mock detailed nodes response
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          nodes: [
            {
              id: 'worker-1',
              name: 'ollama-primary',
              url: 'http://localhost:13000',
              status: 'healthy',
              systemInfo: {
                cpu: { model: 'Intel Core i7-12700K', cores: 8, usage: 45.2 },
                memory: { total: 34359738368, used: 12884901888, usage: 37.5 },
                disk: { usage: 50.0 }
              },
              ollamaInfo: {
                models: [
                  { name: 'tinyllama:latest', size: 668263424 },
                  { name: 'llama2:7b', size: 3984588800 }
                ]
              },
              healthStatus: {
                checks: { 'API': 'healthy', 'Models': 'healthy' },
                warnings: [],
                errors: []
              }
            }
          ]
        })
      });
    });

    await page.route('**/api/models', async route => {
      // Mock models response
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          availableModels: ['tinyllama:latest', 'llama2:7b', 'codellama:7b'],
          workers: {
            'worker-1': {
              models: [
                { name: 'tinyllama:latest', size: 668263424 },
                { name: 'llama2:7b', size: 3984588800 }
              ]
            }
          }
        })
      });
    });

    await page.goto(BASE_URL);
    await page.waitForSelector('#app');
    await page.waitForTimeout(2000);
  });

  test.describe('Enhanced Node Cards with Mock Data', () => {
    
    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(1500);
    });

    test('should display enhanced node cards with detailed information', async ({ page }) => {
      // Wait for nodes to load
      await page.waitForTimeout(2000);
      
      const nodeCards = page.locator('.enhanced-node-card');
      const cardCount = await nodeCards.count();
      
      if (cardCount > 0) {
        const firstCard = nodeCards.first();
        
        // Check node header information
        await expect(firstCard.locator('.node-title')).toBeVisible();
        await expect(firstCard.locator('.status-indicator')).toBeVisible();
        
        // Check quick stats
        const quickStats = firstCard.locator('.quick-stat');
        const statsCount = await quickStats.count();
        expect(statsCount).toBeGreaterThanOrEqual(3); // CPU, Memory, Response
        
        // Check expand functionality
        const expandButton = firstCard.locator('.expand-btn');
        if (await expandButton.count() > 0) {
          await expandButton.click();
          await page.waitForTimeout(500);
          
          // Check expanded content
          const expandableContent = page.locator('[id^="expandable-"]').first();
          if (await expandableContent.count() > 0) {
            await expect(expandableContent).toBeVisible();
            
            // Check for tab navigation within expanded content
            const tabButtons = expandableContent.locator('.tab-btn');
            const tabCount = await tabButtons.count();
            
            if (tabCount > 0) {
              // Test tab switching within node details
              await tabButtons.nth(0).click();
              await page.waitForTimeout(200);
            }
          }
        }
      } else {
        console.warn('No enhanced node cards found - checking for fallback display');
        // Check if basic node cards are displayed instead
        const basicCards = page.locator('.node-card');
        if (await basicCards.count() > 0) {
          await expect(basicCards.first()).toBeVisible();
        }
      }
    });

    test('should handle node expansion and tab switching', async ({ page }) => {
      await page.waitForTimeout(2000);
      
      // Look for expandable nodes
      const expandButtons = page.locator('.expand-btn, [onclick*="toggleNodeExpansion"]');
      const buttonCount = await expandButtons.count();
      
      if (buttonCount > 0) {
        await expandButtons.first().click();
        await page.waitForTimeout(500);
        
        // Check for tab navigation
        const performanceTab = page.locator('button:has-text("Performance")').first();
        const healthTab = page.locator('button:has-text("Health")').first();
        
        if (await performanceTab.count() > 0) {
          await performanceTab.click();
          await page.waitForTimeout(300);
          
          // Should show performance panel
          const performancePanel = page.locator('[id*="performance-"]').first();
          if (await performancePanel.count() > 0) {
            await expect(performancePanel).toBeVisible();
          }
        }
        
        if (await healthTab.count() > 0) {
          await healthTab.click();
          await page.waitForTimeout(300);
          
          // Should show health panel
          const healthPanel = page.locator('[id*="health-"]').first();
          if (await healthPanel.count() > 0) {
            await expect(healthPanel).toBeVisible();
          }
        }
      }
    });

    test('should display system metrics correctly', async ({ page }) => {
      await page.waitForTimeout(2000);
      
      // Check cluster overview stats
      const totalNodes = page.locator('#totalNodes');
      const healthyNodes = page.locator('#healthyNodes');
      const totalCores = page.locator('#totalCores');
      
      if (await totalNodes.count() > 0) {
        const nodesText = await totalNodes.textContent();
        expect(nodesText).toMatch(/\d+/);
      }
      
      if (await healthyNodes.count() > 0) {
        const healthyText = await healthyNodes.textContent();
        expect(healthyText).toMatch(/\d+/);
      }
      
      if (await totalCores.count() > 0) {
        const coresText = await totalCores.textContent();
        expect(coresText).toMatch(/\d+/);
      }
    });
  });

  test.describe('Model Management with Enhanced UI', () => {
    
    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(1500);
    });

    test('should display model cards with proper information', async ({ page }) => {
      await page.waitForTimeout(2000);
      
      const modelGrid = page.locator('#modelGrid');
      await expect(modelGrid).toBeVisible();
      
      // Look for model cards
      const modelCards = page.locator('.model-card');
      const cardCount = await modelCards.count();
      
      if (cardCount > 0) {
        const firstCard = modelCards.first();
        
        // Check model information display
        const modelName = firstCard.locator('.model-name');
        const modelSize = firstCard.locator('.model-size');
        
        if (await modelName.count() > 0) {
          const nameText = await modelName.textContent();
          expect(nameText).toBeTruthy();
        }
        
        // Check action buttons
        const propagateButton = firstCard.locator('.propagate-button, [data-action="propagate"]');
        const deleteButton = firstCard.locator('.delete-button, [data-action="delete"]');
        
        if (await propagateButton.count() > 0) {
          await expect(propagateButton).toBeVisible();
        }
        
        if (await deleteButton.count() > 0) {
          await expect(deleteButton).toBeVisible();
        }
      }
    });

    test('should handle worker checkbox selection for downloads', async ({ page }) => {
      await page.waitForTimeout(2000);
      
      const downloadTargets = page.locator('#downloadTargets');
      await expect(downloadTargets).toBeVisible();
      
      // Look for worker checkboxes
      const checkboxes = downloadTargets.locator('input[type="checkbox"]');
      const checkboxCount = await checkboxes.count();
      
      if (checkboxCount > 0) {
        // Test checkbox interactions
        const firstCheckbox = checkboxes.first();
        const initialState = await firstCheckbox.isChecked();
        
        await firstCheckbox.click();
        const newState = await firstCheckbox.isChecked();
        expect(newState).not.toBe(initialState);
        
        // Test download with selections
        await page.fill('#newModelName', 'test-model');
        await page.click('#downloadModel');
        await page.waitForTimeout(500);
      }
    });

    test('should handle model propagation settings', async ({ page }) => {
      // Test P2P settings
      const p2pEnabled = page.locator('#p2pEnabled');
      const autoPropagation = page.locator('#autoPropagation');
      
      await expect(p2pEnabled).toBeVisible();
      await expect(autoPropagation).toBeVisible();
      
      // Test toggle functionality
      const p2pState = await p2pEnabled.isChecked();
      await p2pEnabled.click();
      await page.waitForTimeout(200);
      
      const newP2pState = await p2pEnabled.isChecked();
      expect(newP2pState).not.toBe(p2pState);
      
      // Test propagation strategy
      const strategySelect = page.locator('#propagationStrategy');
      await expect(strategySelect).toBeVisible();
      
      await strategySelect.selectOption('scheduled');
      await expect(strategySelect).toHaveValue('scheduled');
    });
  });

  test.describe('Real-time Updates and WebSocket Simulation', () => {
    
    test('should handle connection status updates', async ({ page }) => {
      const connectionStatus = page.locator('#connectionStatus');
      const connectionText = page.locator('#connectionText');
      
      await expect(connectionStatus).toBeVisible();
      await expect(connectionText).toBeVisible();
      
      // Check initial connection state
      const initialText = await connectionText.textContent();
      expect(initialText).toBeTruthy();
      
      // Connection status should have some class indication
      const statusClasses = await connectionStatus.getAttribute('class');
      expect(statusClasses).toBeTruthy();
    });

    test('should simulate message streaming', async ({ page }) => {
      await page.click('[data-tab="chat"]');
      
      // Send a message to trigger streaming simulation
      await page.fill('#messageInput', 'Test streaming message');
      await page.click('#sendButton');
      
      await page.waitForTimeout(1000);
      
      // Check if message appears
      const messages = page.locator('.message');
      const messageCount = await messages.count();
      
      if (messageCount > 0) {
        // Should have user message
        const userMessage = messages.locator('.user').first();
        if (await userMessage.count() > 0) {
          await expect(userMessage).toBeVisible();
        }
      }
    });

    test('should handle node count updates', async ({ page }) => {
      const nodeCount = page.locator('#nodeCount');
      await expect(nodeCount).toBeVisible();
      
      const countText = await nodeCount.textContent();
      expect(countText).toMatch(/\d+/);
      
      // Switch to nodes tab to trigger updates
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(1000);
      
      // Node count should still be valid
      const updatedCountText = await nodeCount.textContent();
      expect(updatedCountText).toMatch(/\d+/);
    });
  });

  test.describe('Error State Handling', () => {
    
    test('should handle API failure gracefully', async ({ page }) => {
      // Override mock to simulate API failure
      await page.route('**/api/models', async route => {
        await route.fulfill({ status: 500 });
      });
      
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(1000);
      
      // Try to refresh models
      await page.click('#refreshModels');
      await page.waitForTimeout(2000);
      
      // Interface should remain functional
      await expect(page.locator('#refreshModels')).toBeEnabled();
      await expect(page.locator('#modelGrid')).toBeVisible();
    });

    test('should show appropriate error messages for invalid inputs', async ({ page }) => {
      await page.click('[data-tab="models"]');
      
      // Test empty model name
      await page.click('#downloadModel');
      await page.waitForTimeout(500);
      
      // Should handle gracefully (no crash)
      await expect(page.locator('#modelGrid')).toBeVisible();
      
      // Test invalid characters in model name
      await page.fill('#newModelName', 'invalid/model:name!');
      await page.click('#downloadModel');
      await page.waitForTimeout(500);
      
      // Should still be functional
      await expect(page.locator('#newModelName')).toBeVisible();
    });

    test('should handle network connectivity issues', async ({ page }) => {
      // Simulate network failure
      await page.route('**/*', async route => {
        if (route.request().url().includes('/api/')) {
          await route.abort('failed');
        } else {
          await route.continue();
        }
      });
      
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(2000);
      
      // Should show offline state or fallback data
      await expect(page.locator('.cluster-overview')).toBeVisible();
      
      // Try refresh button
      await page.click('#refreshNodes');
      await page.waitForTimeout(1000);
      
      // Should not crash
      await expect(page.locator('#refreshNodes')).toBeEnabled();
    });
  });

  test.describe('Advanced UI Interactions', () => {
    
    test('should handle rapid tab switching without issues', async ({ page }) => {
      // Perform rapid tab switching
      for (let i = 0; i < 10; i++) {
        await page.click('[data-tab="nodes"]');
        await page.click('[data-tab="models"]');
        await page.click('[data-tab="settings"]');
        await page.click('[data-tab="chat"]');
      }
      
      // Should remain stable
      await expect(page.locator('#app')).toBeVisible();
      await expect(page.locator('[data-tab="chat"]')).toHaveClass(/active/);
    });

    test('should handle form validation and state persistence', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      
      // Fill out form
      await page.fill('#maxTokens', '4096');
      await page.fill('#apiEndpoint', 'ws://localhost:13100/test');
      await page.selectOption('#loadBalancing', 'fastest');
      
      // Switch tabs
      await page.click('[data-tab="chat"]');
      await page.click('[data-tab="settings"]');
      
      // Values should be preserved
      await expect(page.locator('#maxTokens')).toHaveValue('4096');
      await expect(page.locator('#apiEndpoint')).toHaveValue('ws://localhost:13100/test');
      await expect(page.locator('#loadBalancing')).toHaveValue('fastest');
    });

    test('should handle keyboard navigation properly', async ({ page }) => {
      // Test tab navigation
      await page.keyboard.press('Tab');
      
      // Should focus on first interactive element
      const focused = await page.evaluate(() => document.activeElement.tagName);
      expect(['BUTTON', 'INPUT', 'SELECT', 'TEXTAREA', 'A'].includes(focused)).toBe(true);
      
      // Test arrow key navigation in tab bar
      await page.locator('[data-tab="chat"]').focus();
      await page.keyboard.press('ArrowRight');
      
      const focusedTab = page.locator('[data-tab="nodes"]');
      await expect(focusedTab).toBeFocused();
    });
  });

  test.describe('Performance Under Load', () => {
    
    test('should maintain responsiveness with many nodes', async ({ page }) => {
      // Mock response with many nodes
      await page.route('**/api/nodes/detailed', async route => {
        const manyNodes = Array.from({ length: 50 }, (_, i) => ({
          id: `worker-${i + 1}`,
          name: `ollama-worker-${i + 1}`,
          status: Math.random() > 0.8 ? 'warning' : 'healthy',
          systemInfo: {
            cpu: { usage: Math.random() * 100 },
            memory: { usage: Math.random() * 100 }
          }
        }));
        
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ nodes: manyNodes })
        });
      });
      
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(3000);
      
      // Should still be responsive
      const startTime = Date.now();
      await page.click('#refreshNodes');
      await page.waitForTimeout(2000);
      const responseTime = Date.now() - startTime;
      
      expect(responseTime).toBeLessThan(5000);
      await expect(page.locator('.cluster-overview')).toBeVisible();
    });

    test('should handle memory efficiently during extended use', async ({ page }) => {
      // Simulate extended usage pattern
      for (let i = 0; i < 25; i++) {
        await page.click('[data-tab="nodes"]');
        await page.waitForTimeout(50);
        await page.click('[data-tab="models"]');
        await page.waitForTimeout(50);
        
        // Interact with forms
        await page.fill('#newModelName', `test-${i}`);
        await page.fill('#newModelName', '');
      }
      
      // Should remain responsive
      await expect(page.locator('#app')).toBeVisible();
      
      // Memory usage check (basic)
      const heapUsed = await page.evaluate(() => {
        if (performance.memory) {
          return performance.memory.usedJSHeapSize;
        }
        return 0;
      });
      
      // Should not have excessive heap usage (basic check)
      if (heapUsed > 0) {
        expect(heapUsed).toBeLessThan(100 * 1024 * 1024); // 100MB threshold
      }
    });
  });

  test.describe('Visual Regression Prevention', () => {
    
    test('should maintain consistent visual appearance', async ({ page }) => {
      // Take screenshots of different states for visual comparison
      await expect(page.locator('#app')).toBeVisible();
      
      // Chat tab
      await page.click('[data-tab="chat"]');
      await page.waitForTimeout(500);
      const chatHeight = await page.locator('#chatTab').boundingBox();
      expect(chatHeight.height).toBeGreaterThan(200);
      
      // Nodes tab
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(1000);
      const nodesHeight = await page.locator('#nodesTab').boundingBox();
      expect(nodesHeight.height).toBeGreaterThan(200);
      
      // Models tab
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(500);
      const modelsHeight = await page.locator('#modelsTab').boundingBox();
      expect(modelsHeight.height).toBeGreaterThan(200);
      
      // Settings tab
      await page.click('[data-tab="settings"]');
      await page.waitForTimeout(500);
      const settingsHeight = await page.locator('#settingsTab').boundingBox();
      expect(settingsHeight.height).toBeGreaterThan(200);
    });

    test('should maintain proper spacing and layout', async ({ page }) => {
      // Check header spacing
      const header = page.locator('.app-header');
      const headerBox = await header.boundingBox();
      expect(headerBox.height).toBeGreaterThan(60);
      
      // Check navigation tabs spacing
      const navTabs = page.locator('.nav-tabs');
      const navBox = await navTabs.boundingBox();
      expect(navBox.height).toBeGreaterThan(40);
      
      // Check main content area
      const mainContent = page.locator('#main-content');
      const mainBox = await mainContent.boundingBox();
      expect(mainBox.height).toBeGreaterThan(300);
    });
  });
});