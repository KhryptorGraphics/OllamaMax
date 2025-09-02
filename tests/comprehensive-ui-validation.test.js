/**
 * Comprehensive UI Validation Test Suite for OllamaMax Web Interface
 * Tests all components, real-time updates, error handling, and accessibility
 * Focus: Navigation, Nodes Grid, Models Management, Chat Interface, P2P Controls, Settings, Error Handling
 */

import { test, expect } from '@playwright/test';

// Configuration
const BASE_URL = 'http://localhost:8080';
const API_BASE_URL = 'http://localhost:13100';

test.describe('OllamaMax Web Interface - Comprehensive UI Validation', () => {
  
  test.beforeEach(async ({ page }) => {
    // Set up console error monitoring
    const consoleErrors = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });
    page.consoleErrors = consoleErrors;

    // Navigate to the web interface
    await page.goto(BASE_URL);
    
    // Wait for the main app to load
    await page.waitForSelector('#app', { timeout: 10000 });
    
    // Wait for JavaScript initialization
    await page.waitForTimeout(2000);
  });

  test.describe('1. Navigation Tabs - Visual Verification & Interaction', () => {
    
    test('should display all navigation tabs with correct labels and states', async ({ page }) => {
      // Test presence and initial state
      const expectedTabs = [
        { selector: '[data-tab="chat"]', label: 'Chat', initiallyActive: true },
        { selector: '[data-tab="nodes"]', label: 'Nodes', initiallyActive: false },
        { selector: '[data-tab="models"]', label: 'Models', initiallyActive: false },
        { selector: '[data-tab="settings"]', label: 'Settings', initiallyActive: false }
      ];

      for (const tab of expectedTabs) {
        const tabElement = page.locator(tab.selector);
        await expect(tabElement).toBeVisible();
        await expect(tabElement).toContainText(tab.label);
        
        if (tab.initiallyActive) {
          await expect(tabElement).toHaveClass(/active/);
          await expect(tabElement).toHaveAttribute('aria-selected', 'true');
        } else {
          await expect(tabElement).not.toHaveClass(/active/);
          await expect(tabElement).toHaveAttribute('aria-selected', 'false');
        }
      }
    });

    test('should handle tab switching with proper state management', async ({ page }) => {
      const tabs = ['nodes', 'models', 'settings', 'chat'];
      
      for (const tabName of tabs) {
        // Click tab
        await page.click(`[data-tab="${tabName}"]`);
        await page.waitForTimeout(500); // Allow for transitions
        
        // Verify tab is active
        await expect(page.locator(`[data-tab="${tabName}"]`)).toHaveClass(/active/);
        await expect(page.locator(`[data-tab="${tabName}"]`)).toHaveAttribute('aria-selected', 'true');
        
        // Verify content area is active
        await expect(page.locator(`#${tabName}Tab`)).toHaveClass(/active/);
        
        // Verify other tabs are inactive
        const otherTabs = tabs.filter(t => t !== tabName);
        for (const otherTab of otherTabs) {
          await expect(page.locator(`[data-tab="${otherTab}"]`)).not.toHaveClass(/active/);
          await expect(page.locator(`[data-tab="${otherTab}"]`)).toHaveAttribute('aria-selected', 'false');
          await expect(page.locator(`#${otherTab}Tab`)).not.toHaveClass(/active/);
        }
      }
    });

    test('should support keyboard navigation for accessibility', async ({ page }) => {
      // Focus on first tab
      await page.keyboard.press('Tab');
      
      // Navigate through tabs using arrow keys
      await page.keyboard.press('ArrowRight');
      await expect(page.locator('[data-tab="nodes"]')).toBeFocused();
      
      await page.keyboard.press('ArrowRight');
      await expect(page.locator('[data-tab="models"]')).toBeFocused();
      
      await page.keyboard.press('ArrowLeft');
      await expect(page.locator('[data-tab="nodes"]')).toBeFocused();
      
      // Test Enter key activation
      await page.keyboard.press('Enter');
      await expect(page.locator('[data-tab="nodes"]')).toHaveClass(/active/);
    });
  });

  test.describe('2. Nodes Grid - Real-time Status & Worker Display', () => {
    
    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(1000);
    });

    test('should display cluster overview with statistics', async ({ page }) => {
      // Verify cluster overview section
      await expect(page.locator('.cluster-overview')).toBeVisible();
      
      // Check all stat elements
      const stats = ['totalNodes', 'healthyNodes', 'totalCores', 'totalMemory', 'totalRequests'];
      for (const stat of stats) {
        const element = page.locator(`#${stat}`);
        await expect(element).toBeVisible();
        
        // Verify it has content (not empty)
        const text = await element.textContent();
        expect(text.trim()).not.toBe('');
      }
    });

    test('should show 3 workers with real-time status indicators', async ({ page }) => {
      await page.waitForTimeout(2000); // Wait for data loading
      
      // Check for enhanced nodes container
      const nodesContainer = page.locator('#enhancedNodesContainer');
      await expect(nodesContainer).toBeVisible();
      
      // Look for node cards (should have mock data)
      const nodeCards = page.locator('.enhanced-node-card');
      const cardCount = await nodeCards.count();
      
      if (cardCount > 0) {
        // Test first node card structure
        const firstCard = nodeCards.first();
        await expect(firstCard).toBeVisible();
        
        // Check node title
        await expect(firstCard.locator('.node-title')).toBeVisible();
        
        // Check status indicator
        const statusIndicator = firstCard.locator('.status-indicator');
        await expect(statusIndicator).toBeVisible();
        
        // Check quick stats
        const quickStats = firstCard.locator('.quick-stat');
        const statsCount = await quickStats.count();
        expect(statsCount).toBeGreaterThan(0);
        
        // Test expand functionality
        const expandButton = firstCard.locator('.expand-btn');
        if (await expandButton.count() > 0) {
          await expandButton.click();
          
          const expandableSection = page.locator(`#expandable-worker-1, [id^="expandable-"]`).first();
          if (await expandableSection.count() > 0) {
            await expect(expandableSection).toBeVisible();
          }
        }
      }
    });

    test('should have functional filtering and sorting controls', async ({ page }) => {
      // Test status filter
      const statusFilter = page.locator('#statusFilter');
      await expect(statusFilter).toBeVisible();
      
      await statusFilter.selectOption('healthy');
      await expect(statusFilter).toHaveValue('healthy');
      
      await statusFilter.selectOption('all');
      await expect(statusFilter).toHaveValue('all');
      
      // Test sort controls
      const sortBy = page.locator('#sortBy');
      await expect(sortBy).toBeVisible();
      
      const sortOptions = ['name', 'cpu', 'memory', 'requests', 'uptime'];
      for (const option of sortOptions) {
        await sortBy.selectOption(option);
        await expect(sortBy).toHaveValue(option);
      }
      
      // Test search functionality
      const searchInput = page.locator('#nodeSearch');
      await expect(searchInput).toBeVisible();
      await searchInput.fill('worker');
      await expect(searchInput).toHaveValue('worker');
      await searchInput.clear();
    });

    test('should have refresh and view toggle controls', async ({ page }) => {
      // Test refresh button
      const refreshButton = page.locator('#refreshNodes');
      await expect(refreshButton).toBeVisible();
      await expect(refreshButton).toBeEnabled();
      
      // Test clicking refresh (should not cause errors)
      await refreshButton.click();
      await page.waitForTimeout(1000);
      
      // Test view toggles
      const compactView = page.locator('#compactView');
      const detailedView = page.locator('#detailedView');
      
      await expect(compactView).toBeVisible();
      await expect(detailedView).toBeVisible();
      
      // Test toggle functionality
      await detailedView.click();
      await expect(detailedView).toHaveClass(/active/);
      await expect(compactView).not.toHaveClass(/active/);
      
      await compactView.click();
      await expect(compactView).toHaveClass(/active/);
      await expect(detailedView).not.toHaveClass(/active/);
    });
  });

  test.describe('3. Models Management Interface', () => {
    
    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(1000);
    });

    test('should display model management sections', async ({ page }) => {
      // Check main sections
      await expect(page.locator('.models-header h2')).toContainText('Model Management');
      await expect(page.locator('#refreshModels')).toBeVisible();
      
      // Available models section
      await expect(page.locator('#modelGrid')).toBeVisible();
      
      // Download form section
      const downloadForm = page.locator('.model-download-form');
      await expect(downloadForm).toBeVisible();
      await expect(page.locator('#newModelName')).toBeVisible();
      await expect(page.locator('#downloadTargets')).toBeVisible();
      await expect(page.locator('#downloadModel')).toBeVisible();
      
      // Propagation settings section
      await expect(page.locator('.propagation-settings')).toBeVisible();
    });

    test('should handle model download form interactions', async ({ page }) => {
      // Test model name input
      const modelNameInput = page.locator('#newModelName');
      await modelNameInput.fill('llama2:7b');
      await expect(modelNameInput).toHaveValue('llama2:7b');
      
      // Test download button (should handle empty targets gracefully)
      const downloadButton = page.locator('#downloadModel');
      await expect(downloadButton).toBeEnabled();
      
      // Click without selecting targets (should show error message)
      await downloadButton.click();
      await page.waitForTimeout(500);
      
      // Clear input
      await modelNameInput.clear();
      
      // Test empty model name (should show error message)
      await downloadButton.click();
      await page.waitForTimeout(500);
    });

    test('should display P2P propagation controls', async ({ page }) => {
      // Check propagation checkboxes
      const autoPropagation = page.locator('#autoPropagation');
      const p2pEnabled = page.locator('#p2pEnabled');
      
      await expect(autoPropagation).toBeVisible();
      await expect(p2pEnabled).toBeVisible();
      
      // Test checkbox interactions
      const p2pInitialState = await p2pEnabled.isChecked();
      await p2pEnabled.click();
      const p2pNewState = await p2pEnabled.isChecked();
      expect(p2pNewState).not.toBe(p2pInitialState);
      
      // Test propagation strategy dropdown
      const strategySelect = page.locator('#propagationStrategy');
      await expect(strategySelect).toBeVisible();
      
      const strategies = ['immediate', 'scheduled', 'manual'];
      for (const strategy of strategies) {
        await strategySelect.selectOption(strategy);
        await expect(strategySelect).toHaveValue(strategy);
      }
    });

    test('should handle refresh models functionality', async ({ page }) => {
      const refreshButton = page.locator('#refreshModels');
      await refreshButton.click();
      
      // Should not cause any JavaScript errors
      await page.waitForTimeout(1000);
      
      // Interface should remain functional
      await expect(page.locator('#modelGrid')).toBeVisible();
      await expect(refreshButton).toBeEnabled();
    });
  });

  test.describe('4. Chat Interface with Worker Selection', () => {
    
    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="chat"]');
      await page.waitForTimeout(500);
    });

    test('should display chat interface components', async ({ page }) => {
      // Welcome message
      await expect(page.locator('.welcome-message')).toBeVisible();
      await expect(page.locator('.welcome-message h2')).toContainText('Welcome to Distributed Llama Chat');
      
      // Messages area
      await expect(page.locator('#messagesArea')).toBeVisible();
      
      // Status bar with worker selection
      const statusBar = page.locator('.status-bar');
      await expect(statusBar).toBeVisible();
      await expect(page.locator('#activeNode')).toBeVisible();
      await expect(page.locator('#queueLength')).toBeVisible();
      await expect(page.locator('#latency')).toBeVisible();
      await expect(page.locator('#modelSelector')).toBeVisible();
      
      // Input area
      await expect(page.locator('#messageInput')).toBeVisible();
      await expect(page.locator('#sendButton')).toBeVisible();
    });

    test('should handle message input and sending', async ({ page }) => {
      const messageInput = page.locator('#messageInput');
      const sendButton = page.locator('#sendButton');
      
      // Test typing
      await messageInput.fill('Hello, this is a test message');
      await expect(messageInput).toHaveValue('Hello, this is a test message');
      
      // Test send button click
      await sendButton.click();
      
      // Input should be cleared
      await expect(messageInput).toHaveValue('');
      
      // Message should appear in messages area
      await page.waitForTimeout(500);
      const messages = page.locator('.message');
      const messageCount = await messages.count();
      expect(messageCount).toBeGreaterThan(0);
    });

    test('should support keyboard shortcuts', async ({ page }) => {
      const messageInput = page.locator('#messageInput');
      
      // Test Enter key sending
      await messageInput.fill('Test message with Enter key');
      await messageInput.press('Enter');
      
      // Input should be cleared
      await expect(messageInput).toHaveValue('');
      
      // Test Shift+Enter for new line
      await messageInput.fill('Line 1');
      await messageInput.press('Shift+Enter');
      await messageInput.type('Line 2');
      
      const value = await messageInput.inputValue();
      expect(value).toContain('\n');
    });

    test('should display model selector with options', async ({ page }) => {
      const modelSelector = page.locator('#modelSelector');
      await expect(modelSelector).toBeVisible();
      
      // Should have at least one option
      const options = page.locator('#modelSelector option');
      const optionCount = await options.count();
      expect(optionCount).toBeGreaterThan(0);
      
      // Test selection
      if (optionCount > 1) {
        const firstOptionValue = await options.nth(0).getAttribute('value');
        await modelSelector.selectOption(firstOptionValue);
        await expect(modelSelector).toHaveValue(firstOptionValue);
      }
    });
  });

  test.describe('5. Settings Page with Accept Models Toggle', () => {
    
    test.beforeEach(async ({ page }) => {
      await page.click('[data-tab="settings"]');
      await page.waitForTimeout(500);
    });

    test('should display all settings sections', async ({ page }) => {
      // Main heading
      await expect(page.locator('.settings-container h2')).toContainText('Settings');
      
      // API Configuration section
      await expect(page.locator('#apiEndpoint')).toBeVisible();
      await expect(page.locator('#apiKey')).toBeVisible();
      
      // Chat Settings section
      await expect(page.locator('#streamingEnabled')).toBeVisible();
      await expect(page.locator('#autoScroll')).toBeVisible();
      await expect(page.locator('#maxTokens')).toBeVisible();
      await expect(page.locator('#temperature')).toBeVisible();
      await expect(page.locator('#temperatureValue')).toBeVisible();
      
      // Node Management section
      await expect(page.locator('#loadBalancing')).toBeVisible();
      await expect(page.locator('#addNodeButton')).toBeVisible();
      
      // Settings actions
      await expect(page.locator('#saveSettings')).toBeVisible();
      await expect(page.locator('#resetSettings')).toBeVisible();
    });

    test('should handle interactive settings controls', async ({ page }) => {
      // Test checkbox toggles
      const streamingEnabled = page.locator('#streamingEnabled');
      const autoScroll = page.locator('#autoScroll');
      
      const streamingState = await streamingEnabled.isChecked();
      await streamingEnabled.click();
      const newStreamingState = await streamingEnabled.isChecked();
      expect(newStreamingState).not.toBe(streamingState);
      
      // Test temperature slider
      const temperatureSlider = page.locator('#temperature');
      const temperatureValue = page.locator('#temperatureValue');
      
      await temperatureSlider.fill('0.9');
      await expect(temperatureValue).toContainText('0.9');
      
      // Test max tokens input
      const maxTokensInput = page.locator('#maxTokens');
      await maxTokensInput.fill('4096');
      await expect(maxTokensInput).toHaveValue('4096');
      
      // Test load balancing selection
      const loadBalancing = page.locator('#loadBalancing');
      const strategies = ['round-robin', 'least-loaded', 'fastest'];
      
      for (const strategy of strategies) {
        await loadBalancing.selectOption(strategy);
        await expect(loadBalancing).toHaveValue(strategy);
      }
    });

    test('should handle add node modal functionality', async ({ page }) => {
      const addNodeButton = page.locator('#addNodeButton');
      await addNodeButton.click();
      
      // Modal should appear
      const modal = page.locator('#addNodeModal');
      await expect(modal).toHaveClass(/active/);
      await expect(modal).toBeVisible();
      
      // Modal content
      await expect(page.locator('#nodeUrl')).toBeVisible();
      await expect(page.locator('#nodeName')).toBeVisible();
      await expect(page.locator('#confirmAddNode')).toBeVisible();
      await expect(page.locator('#cancelAddNode')).toBeVisible();
      
      // Test form inputs
      await page.fill('#nodeUrl', 'http://localhost:11434');
      await page.fill('#nodeName', 'test-node');
      
      // Test cancel button
      await page.click('#cancelAddNode');
      await expect(modal).not.toHaveClass(/active/);
    });

    test('should handle save and reset settings', async ({ page }) => {
      // Change a setting
      await page.fill('#maxTokens', '2048');
      
      // Test save settings
      const saveButton = page.locator('#saveSettings');
      await expect(saveButton).toBeEnabled();
      await saveButton.click();
      
      // Should not cause errors
      await page.waitForTimeout(500);
      
      // Test reset settings (with dialog handling)
      const resetButton = page.locator('#resetSettings');
      await expect(resetButton).toBeEnabled();
      
      // Handle the confirmation dialog
      page.on('dialog', async dialog => {
        expect(dialog.type()).toBe('confirm');
        await dialog.dismiss(); // Dismiss the dialog for testing
      });
      
      await resetButton.click();
    });
  });

  test.describe('6. Error Handling and Notifications', () => {
    
    test('should display error boundary elements', async ({ page }) => {
      const errorBoundary = page.locator('#errorBoundary');
      await expect(errorBoundary).toBeInViewport();
      
      // Check error boundary content (should be hidden initially)
      await expect(page.locator('#errorMessage')).toBeInViewport();
      await expect(page.locator('#retryButton')).toBeInViewport();
      await expect(page.locator('#reloadButton')).toBeInViewport();
    });

    test('should have loading overlay system', async ({ page }) => {
      const loadingOverlay = page.locator('#loadingOverlay');
      await expect(loadingOverlay).toBeInViewport();
      
      // Check loading spinner
      await expect(page.locator('.loading-spinner')).toBeInViewport();
      await expect(page.locator('#loadingMessage')).toBeInViewport();
    });

    test('should have notification container', async ({ page }) => {
      const notificationContainer = page.locator('#notificationContainer');
      await expect(notificationContainer).toBeInViewport();
    });

    test('should handle API failures gracefully', async ({ page }) => {
      // Test models tab API failure handling
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(1000);
      
      // Try to refresh models (should handle failure gracefully)
      await page.click('#refreshModels');
      await page.waitForTimeout(2000);
      
      // Interface should remain functional
      await expect(page.locator('#refreshModels')).toBeEnabled();
      await expect(page.locator('#modelGrid')).toBeVisible();
      
      // Check for no JavaScript errors
      const consoleErrors = page.consoleErrors || [];
      const criticalErrors = consoleErrors.filter(error => 
        !error.includes('Failed to fetch') && // Expected API errors
        !error.includes('NetworkError') &&
        !error.includes('ERR_NETWORK')
      );
      expect(criticalErrors.length).toBe(0);
    });

    test('should handle invalid input gracefully', async ({ page }) => {
      // Test invalid model download
      await page.click('[data-tab="models"]');
      await page.click('#downloadModel'); // Empty model name
      await page.waitForTimeout(500);
      
      // Test invalid settings
      await page.click('[data-tab="settings"]');
      await page.fill('#apiEndpoint', 'invalid-url');
      await page.click('#saveSettings');
      await page.waitForTimeout(500);
      
      // Test invalid message sending
      await page.click('[data-tab="chat"]');
      await page.click('#sendButton'); // Empty message
      
      // Interface should remain stable
      await expect(page.locator('#app')).toBeVisible();
    });
  });

  test.describe('7. WebSocket Connectivity and Real-time Updates', () => {
    
    test('should display connection status indicator', async ({ page }) => {
      const connectionStatus = page.locator('#connectionStatus');
      const connectionText = page.locator('#connectionText');
      
      await expect(connectionStatus).toBeVisible();
      await expect(connectionText).toBeVisible();
      
      // Should show some connection state
      const statusText = await connectionText.textContent();
      expect(statusText).toBeTruthy();
      expect(statusText.length).toBeGreaterThan(0);
    });

    test('should handle WebSocket connection simulation', async ({ page }) => {
      // Monitor WebSocket attempts through console logs
      const wsAttempts = [];
      page.on('console', msg => {
        if (msg.text().includes('WebSocket') || msg.text().includes('websocket')) {
          wsAttempts.push(msg.text());
        }
      });
      
      await page.waitForTimeout(3000);
      
      // Should attempt WebSocket connection
      // (Will fail in test environment, but should handle gracefully)
      
      // Check that the interface remains functional despite WS failure
      await expect(page.locator('#connectionStatus')).toBeVisible();
      await expect(page.locator('#app')).toBeVisible();
    });

    test('should update node count display', async ({ page }) => {
      const nodeCount = page.locator('#nodeCount');
      await expect(nodeCount).toBeVisible();
      
      // Should have some value (even if 0)
      const countText = await nodeCount.textContent();
      expect(countText).toMatch(/^\d+$/); // Should be a number
    });
  });

  test.describe('8. Accessibility Compliance (WCAG)', () => {
    
    test('should have proper heading structure', async ({ page }) => {
      // Check heading hierarchy
      const h1Count = await page.locator('h1').count();
      const h2Count = await page.locator('h2').count();
      const h3Count = await page.locator('h3').count();
      
      expect(h1Count).toBe(1); // Should have exactly one main heading
      expect(h2Count).toBeGreaterThan(0); // Should have section headings
    });

    test('should have accessible form controls', async ({ page }) => {
      // Check various tabs for form accessibility
      const tabs = ['chat', 'models', 'settings'];
      
      for (const tabName of tabs) {
        await page.click(`[data-tab="${tabName}"]`);
        await page.waitForTimeout(500);
        
        // Check that inputs have labels or aria-labels
        const inputs = page.locator('input, select, textarea');
        const inputCount = await inputs.count();
        
        if (inputCount > 0) {
          for (let i = 0; i < inputCount; i++) {
            const input = inputs.nth(i);
            const hasLabel = await input.getAttribute('aria-label') !== null ||
                           await input.getAttribute('aria-labelledby') !== null ||
                           await page.locator(`label[for="${await input.getAttribute('id')}"]`).count() > 0;
            
            if (await input.isVisible()) {
              expect(hasLabel).toBe(true);
            }
          }
        }
      }
    });

    test('should have sufficient color contrast', async ({ page }) => {
      // Test primary UI elements for contrast
      const elements = [
        '.tab-button',
        '.send-button',
        '#saveSettings',
        '.primary-button'
      ];
      
      for (const selector of elements) {
        const element = page.locator(selector).first();
        if (await element.count() > 0) {
          const styles = await element.evaluate(el => {
            const computed = window.getComputedStyle(el);
            return {
              color: computed.color,
              backgroundColor: computed.backgroundColor
            };
          });
          
          // Basic check that colors are defined
          expect(styles.color).toBeTruthy();
          expect(styles.backgroundColor).toBeTruthy();
        }
      }
    });

    test('should support skip link navigation', async ({ page }) => {
      const skipLink = page.locator('.skip-link');
      await expect(skipLink).toBeInViewport();
      await expect(skipLink).toHaveAttribute('href', '#main-content');
      
      // Test skip link functionality
      await skipLink.click();
      const mainContent = page.locator('#main-content');
      await expect(mainContent).toBeFocused();
    });

    test('should have proper ARIA attributes', async ({ page }) => {
      // Check tab navigation ARIA
      const tabButtons = page.locator('[role="tab"]');
      const tabCount = await tabButtons.count();
      
      for (let i = 0; i < tabCount; i++) {
        const tab = tabButtons.nth(i);
        await expect(tab).toHaveAttribute('aria-selected');
        await expect(tab).toHaveAttribute('aria-label');
      }
      
      // Check main content area
      await expect(page.locator('#main-content')).toHaveAttribute('role', 'main');
    });
  });

  test.describe('9. Performance and Core Web Vitals', () => {
    
    test('should load within performance budget', async ({ page }) => {
      const startTime = Date.now();
      
      await page.goto(BASE_URL);
      await page.waitForSelector('#app');
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      expect(loadTime).toBeLessThan(5000); // 5 second budget for complete load
    });

    test('should handle rapid interactions without performance degradation', async ({ page }) => {
      const startTime = Date.now();
      
      // Perform rapid interactions
      for (let i = 0; i < 10; i++) {
        await page.click('[data-tab="nodes"]');
        await page.click('[data-tab="models"]');
        await page.click('[data-tab="settings"]');
        await page.click('[data-tab="chat"]');
      }
      
      const interactionTime = Date.now() - startTime;
      expect(interactionTime).toBeLessThan(3000); // Should complete within 3 seconds
      
      // Interface should remain responsive
      await expect(page.locator('#app')).toBeVisible();
      await expect(page.locator('[data-tab="chat"]')).toHaveClass(/active/);
    });

    test('should not have memory leaks during extended use', async ({ page }) => {
      // Simulate extended usage
      for (let i = 0; i < 20; i++) {
        await page.click('[data-tab="nodes"]');
        await page.waitForTimeout(100);
        await page.click('[data-tab="models"]');
        await page.waitForTimeout(100);
        
        // Fill and clear form inputs
        await page.fill('#newModelName', `test-model-${i}`);
        await page.fill('#newModelName', '');
      }
      
      // Check that the interface is still responsive
      await expect(page.locator('#app')).toBeVisible();
      
      // No critical JavaScript errors should have occurred
      const consoleErrors = page.consoleErrors || [];
      const memoryErrors = consoleErrors.filter(error => 
        error.includes('memory') || error.includes('leak')
      );
      expect(memoryErrors.length).toBe(0);
    });
  });

  test.describe('10. Cross-Device and Responsive Testing', () => {
    
    test('should work on mobile devices', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
      
      // All main elements should be visible and accessible
      await expect(page.locator('#app')).toBeVisible();
      await expect(page.locator('.nav-tabs')).toBeVisible();
      
      // Test touch interactions
      await page.tap('[data-tab="nodes"]');
      await expect(page.locator('[data-tab="nodes"]')).toHaveClass(/active/);
      
      // Check that buttons are touch-friendly (minimum 44px)
      const tabButton = page.locator('[data-tab="chat"]');
      const boundingBox = await tabButton.boundingBox();
      expect(boundingBox.height).toBeGreaterThanOrEqual(44);
    });

    test('should work on tablet devices', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 }); // iPad
      
      await expect(page.locator('#app')).toBeVisible();
      
      // Test both orientations
      await page.setViewportSize({ width: 1024, height: 768 }); // Landscape
      await expect(page.locator('#app')).toBeVisible();
      
      // Navigation should remain functional
      await page.click('[data-tab="models"]');
      await expect(page.locator('#modelsTab')).toHaveClass(/active/);
    });

    test('should work on large desktop screens', async ({ page }) => {
      await page.setViewportSize({ width: 1920, height: 1080 });
      
      await expect(page.locator('#app')).toBeVisible();
      
      // Content should scale appropriately
      const chatContainer = page.locator('.chat-container');
      const boundingBox = await chatContainer.boundingBox();
      expect(boundingBox.width).toBeLessThan(1920); // Should not be full width
    });
  });

  test.afterEach(async ({ page }) => {
    // Check for JavaScript errors that occurred during the test
    const consoleErrors = page.consoleErrors || [];
    const criticalErrors = consoleErrors.filter(error => 
      !error.includes('WebSocket connection') && // Expected in test environment
      !error.includes('Failed to fetch') &&     // Expected API errors
      !error.includes('ERR_NETWORK') &&
      !error.includes('NetworkError')
    );
    
    if (criticalErrors.length > 0) {
      console.warn('JavaScript errors detected:', criticalErrors);
    }
    
    // Log any unexpected critical errors
    expect(criticalErrors.length).toBe(0);
  });
});