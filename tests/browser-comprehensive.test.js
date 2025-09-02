/**
 * Comprehensive Browser Testing for OllamaMax Web Interface
 * Tests all components, interactions, and UI elements
 */

import { test, expect } from '@playwright/test';

// Configuration
const BASE_URL = 'http://localhost:13100';
const WEB_INTERFACE_URL = `${BASE_URL}/web-interface/index.html`;

test.describe('OllamaMax Web Interface - Comprehensive Testing', () => {
  
  test.beforeEach(async ({ page }) => {
    // Navigate to the web interface
    await page.goto(WEB_INTERFACE_URL);
    
    // Wait for the main app to load
    await page.waitForSelector('#app');
    
    // Wait a bit for JavaScript initialization
    await page.waitForTimeout(2000);
  });

  test.describe('Initial Load and Layout', () => {
    
    test('should display the main header and navigation', async ({ page }) => {
      // Check header title
      await expect(page.locator('h1')).toContainText('Distributed Llama Chat');
      
      // Check all navigation tabs are present
      const tabs = ['Chat', 'Nodes', 'Models', 'Settings'];
      for (const tab of tabs) {
        await expect(page.locator(`button:has-text("${tab}")`)).toBeVisible();
      }
      
      // Check connection status indicator
      await expect(page.locator('#connectionStatus')).toBeVisible();
      await expect(page.locator('#connectionText')).toBeVisible();
    });

    test('should show correct initial tab state', async ({ page }) => {
      // Chat tab should be active by default
      await expect(page.locator('[data-tab="chat"]')).toHaveClass(/active/);
      await expect(page.locator('#chatTab')).toHaveClass(/active/);
      
      // Other tabs should not be active
      await expect(page.locator('[data-tab="nodes"]')).not.toHaveClass(/active/);
      await expect(page.locator('[data-tab="models"]')).not.toHaveClass(/active/);
      await expect(page.locator('[data-tab="settings"]')).not.toHaveClass(/active/);
    });

    test('should display welcome message and node count', async ({ page }) => {
      await expect(page.locator('.welcome-message h2')).toContainText('Welcome to Distributed Llama Chat');
      await expect(page.locator('#nodeCount')).toBeVisible();
    });
  });

  test.describe('Tab Navigation', () => {
    
    test('should switch to Nodes tab correctly', async ({ page }) => {
      await page.click('[data-tab="nodes"]');
      
      // Check tab activation
      await expect(page.locator('[data-tab="nodes"]')).toHaveClass(/active/);
      await expect(page.locator('#nodesTab')).toHaveClass(/active/);
      
      // Check nodes dashboard elements
      await expect(page.locator('.nodes-header h2')).toContainText('Distributed Nodes Dashboard');
      await expect(page.locator('.cluster-overview')).toBeVisible();
      await expect(page.locator('#enhancedNodesContainer')).toBeVisible();
    });

    test('should switch to Models tab correctly', async ({ page }) => {
      await page.click('[data-tab="models"]');
      
      // Check tab activation
      await expect(page.locator('[data-tab="models"]')).toHaveClass(/active/);
      await expect(page.locator('#modelsTab')).toHaveClass(/active/);
      
      // Check models management elements
      await expect(page.locator('.models-header h2')).toContainText('Model Management');
      await expect(page.locator('#modelGrid')).toBeVisible();
      await expect(page.locator('.model-download-form')).toBeVisible();
    });

    test('should switch to Settings tab correctly', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      
      // Check tab activation
      await expect(page.locator('[data-tab="settings"]')).toHaveClass(/active/);
      await expect(page.locator('#settingsTab')).toHaveClass(/active/);
      
      // Check settings elements
      await expect(page.locator('.settings-container h2')).toContainText('Settings');
      await expect(page.locator('#apiEndpoint')).toBeVisible();
      await expect(page.locator('#saveSettings')).toBeVisible();
    });
  });

  test.describe('Chat Interface', () => {
    
    test('should have functional message input', async ({ page }) => {
      const messageInput = page.locator('#messageInput');
      const sendButton = page.locator('#sendButton');
      
      await expect(messageInput).toBeVisible();
      await expect(sendButton).toBeVisible();
      
      // Test typing in message input
      await messageInput.fill('Test message');
      await expect(messageInput).toHaveValue('Test message');
    });

    test('should display status bar with metrics', async ({ page }) => {
      await expect(page.locator('#activeNode')).toBeVisible();
      await expect(page.locator('#queueLength')).toBeVisible();
      await expect(page.locator('#latency')).toBeVisible();
      await expect(page.locator('#modelSelector')).toBeVisible();
    });

    test('should handle Enter key for message sending', async ({ page }) => {
      const messageInput = page.locator('#messageInput');
      
      await messageInput.fill('Test message for Enter key');
      
      // Simulate Enter key press
      await messageInput.press('Enter');
      
      // Message input should be cleared after sending
      await expect(messageInput).toHaveValue('');
    });
  });

  test.describe('Nodes Management', () => {
    
    test('should load and display cluster overview', async ({ page }) => {
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(1000);
      
      // Check cluster stats elements
      await expect(page.locator('#totalNodes')).toBeVisible();
      await expect(page.locator('#healthyNodes')).toBeVisible();
      await expect(page.locator('#totalCores')).toBeVisible();
      await expect(page.locator('#totalMemory')).toBeVisible();
    });

    test('should have functional node filters', async ({ page }) => {
      await page.click('[data-tab="nodes"]');
      
      // Check filter elements
      await expect(page.locator('#statusFilter')).toBeVisible();
      await expect(page.locator('#sortBy')).toBeVisible();
      await expect(page.locator('#nodeSearch')).toBeVisible();
      
      // Test status filter
      await page.selectOption('#statusFilter', 'healthy');
      await expect(page.locator('#statusFilter')).toHaveValue('healthy');
    });

    test('should have refresh and view toggle buttons', async ({ page }) => {
      await page.click('[data-tab="nodes"]');
      
      await expect(page.locator('#refreshNodes')).toBeVisible();
      await expect(page.locator('#compactView')).toBeVisible();
      await expect(page.locator('#detailedView')).toBeVisible();
      
      // Test view toggles
      await page.click('#detailedView');
      await expect(page.locator('#detailedView')).toHaveClass(/active/);
    });
  });

  test.describe('Models Management', () => {
    
    test('should display model management interface', async ({ page }) => {
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(1000);
      
      // Check main elements
      await expect(page.locator('#refreshModels')).toBeVisible();
      await expect(page.locator('#modelGrid')).toBeVisible();
      await expect(page.locator('#newModelName')).toBeVisible();
      await expect(page.locator('#downloadModel')).toBeVisible();
    });

    test('should have functional model download form', async ({ page }) => {
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(1000);
      
      // Fill model name
      await page.fill('#newModelName', 'llama2:7b');
      await expect(page.locator('#newModelName')).toHaveValue('llama2:7b');
      
      // Check download targets container
      await expect(page.locator('#downloadTargets')).toBeVisible();
    });

    test('should display propagation settings', async ({ page }) => {
      await page.click('[data-tab="models"]');
      
      await expect(page.locator('#autoPropagation')).toBeVisible();
      await expect(page.locator('#p2pEnabled')).toBeVisible();
      await expect(page.locator('#propagationStrategy')).toBeVisible();
      
      // Test checkbox functionality
      const p2pCheckbox = page.locator('#p2pEnabled');
      const isChecked = await p2pCheckbox.isChecked();
      await p2pCheckbox.click();
      await expect(p2pCheckbox).toBeChecked(!isChecked);
    });
  });

  test.describe('Settings Interface', () => {
    
    test('should display all settings sections', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      
      // API Configuration
      await expect(page.locator('#apiEndpoint')).toBeVisible();
      await expect(page.locator('#apiKey')).toBeVisible();
      
      // Chat Settings  
      await expect(page.locator('#streamingEnabled')).toBeVisible();
      await expect(page.locator('#autoScroll')).toBeVisible();
      await expect(page.locator('#maxTokens')).toBeVisible();
      await expect(page.locator('#temperature')).toBeVisible();
      
      // Node Management
      await expect(page.locator('#loadBalancing')).toBeVisible();
      await expect(page.locator('#addNodeButton')).toBeVisible();
    });

    test('should have functional settings controls', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      
      // Test temperature slider
      const tempSlider = page.locator('#temperature');
      await tempSlider.fill('0.8');
      await expect(page.locator('#temperatureValue')).toContainText('0.8');
      
      // Test max tokens input
      await page.fill('#maxTokens', '4096');
      await expect(page.locator('#maxTokens')).toHaveValue('4096');
    });

    test('should have save and reset buttons', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      
      await expect(page.locator('#saveSettings')).toBeVisible();
      await expect(page.locator('#resetSettings')).toBeVisible();
      
      // Test buttons are clickable
      await expect(page.locator('#saveSettings')).toBeEnabled();
      await expect(page.locator('#resetSettings')).toBeEnabled();
    });
  });

  test.describe('Modal and Interactive Elements', () => {
    
    test('should open add node modal', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      await page.click('#addNodeButton');
      
      // Modal should be visible
      await expect(page.locator('#addNodeModal')).toHaveClass(/active/);
      
      // Modal content should be present
      await expect(page.locator('#nodeUrl')).toBeVisible();
      await expect(page.locator('#nodeName')).toBeVisible();
      await expect(page.locator('#confirmAddNode')).toBeVisible();
      await expect(page.locator('#cancelAddNode')).toBeVisible();
    });

    test('should close add node modal', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      await page.click('#addNodeButton');
      
      // Close modal
      await page.click('#cancelAddNode');
      
      // Modal should not have active class
      await expect(page.locator('#addNodeModal')).not.toHaveClass(/active/);
    });
  });

  test.describe('Responsive Design and Accessibility', () => {
    
    test('should be responsive on mobile viewport', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      // Main elements should still be visible
      await expect(page.locator('#app')).toBeVisible();
      await expect(page.locator('.nav-tabs')).toBeVisible();
      await expect(page.locator('.tab-content')).toBeVisible();
    });

    test('should have accessible form labels', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      
      // Check for proper labels
      const labels = await page.locator('label').count();
      expect(labels).toBeGreaterThan(0);
    });

    test('should support keyboard navigation', async ({ page }) => {
      // Test tab key navigation
      await page.keyboard.press('Tab');
      await page.keyboard.press('Tab');
      
      // Focus should be on a focusable element
      const focusedElement = await page.evaluate(() => document.activeElement.tagName);
      expect(['BUTTON', 'INPUT', 'SELECT', 'TEXTAREA']).toContain(focusedElement);
    });
  });

  test.describe('Error Handling and Edge Cases', () => {
    
    test('should handle empty model name gracefully', async ({ page }) => {
      await page.click('[data-tab="models"]');
      
      // Try to download with empty name
      await page.click('#downloadModel');
      
      // Should not crash - page should still be functional
      await expect(page.locator('#modelGrid')).toBeVisible();
    });

    test('should handle invalid API endpoint', async ({ page }) => {
      await page.click('[data-tab="settings"]');
      
      // Set invalid endpoint
      await page.fill('#apiEndpoint', 'invalid-url');
      await page.click('#saveSettings');
      
      // Should not crash the interface
      await expect(page.locator('#apiEndpoint')).toBeVisible();
    });

    test('should maintain state during tab switching', async ({ page }) => {
      // Fill a form in settings
      await page.click('[data-tab="settings"]');
      await page.fill('#maxTokens', '4096');
      
      // Switch to another tab and back
      await page.click('[data-tab="chat"]');
      await page.click('[data-tab="settings"]');
      
      // Value should be preserved
      await expect(page.locator('#maxTokens')).toHaveValue('4096');
    });
  });

  test.describe('Performance and Loading', () => {
    
    test('should load within reasonable time', async ({ page }) => {
      const startTime = Date.now();
      await page.goto(WEB_INTERFACE_URL);
      await page.waitForSelector('#app');
      const loadTime = Date.now() - startTime;
      
      // Should load within 5 seconds
      expect(loadTime).toBeLessThan(5000);
    });

    test('should handle rapid tab switching', async ({ page }) => {
      // Rapidly switch between tabs
      for (let i = 0; i < 5; i++) {
        await page.click('[data-tab="nodes"]');
        await page.waitForTimeout(100);
        await page.click('[data-tab="models"]');
        await page.waitForTimeout(100);
        await page.click('[data-tab="settings"]');
        await page.waitForTimeout(100);
        await page.click('[data-tab="chat"]');
        await page.waitForTimeout(100);
      }
      
      // Interface should remain functional
      await expect(page.locator('#app')).toBeVisible();
    });
  });
});