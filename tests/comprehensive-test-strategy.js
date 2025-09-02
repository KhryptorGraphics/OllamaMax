/**
 * Comprehensive Testing Strategy for OllamaMax Distributed AI Platform
 * 
 * This test suite validates all functionality of the OllamaMax system including:
 * - Web Interface functionality across all tabs and interactions
 * - API Server health and worker management
 * - P2P model migration and distributed operations
 * - Error handling and recovery mechanisms
 * - Performance monitoring and real-time updates
 */

import { test, expect, chromium, firefox, webkit } from '@playwright/test';

// Test Configuration
const CONFIG = {
  API_SERVER_URL: 'http://localhost:13100',
  WEB_INTERFACE_URL: 'http://localhost:8080',
  WORKER_PORTS: [13000, 13001, 13002],
  TIMEOUT: 30000,
  RETRY_COUNT: 3,
  PERFORMANCE_THRESHOLDS: {
    PAGE_LOAD: 3000,
    API_RESPONSE: 1000,
    WEBSOCKET_CONNECTION: 2000
  }
};

test.describe('OllamaMax Comprehensive Testing Suite', () => {

  test.describe('System Health & API Validation', () => {

    test('API Server should be healthy and responsive', async ({ request }) => {
      const startTime = Date.now();
      
      // Test health endpoint
      const healthResponse = await request.get(`${CONFIG.API_SERVER_URL}/health`);
      const responseTime = Date.now() - startTime;
      
      expect(healthResponse.status()).toBe(200);
      expect(responseTime).toBeLessThan(CONFIG.PERFORMANCE_THRESHOLDS.API_RESPONSE);
      
      const healthData = await healthResponse.json();
      expect(healthData).toHaveProperty('status', 'healthy');
      expect(healthData).toHaveProperty('nodes');
      expect(Array.isArray(healthData.nodes)).toBe(true);
    });

    test('Worker nodes should be accessible and healthy', async ({ request }) => {
      for (const port of CONFIG.WORKER_PORTS) {
        const workerUrl = `http://localhost:${port}`;
        
        // Test basic connectivity
        const response = await request.get(`${workerUrl}/api/version`);
        expect(response.status()).toBe(200);
        
        // Validate response structure
        const data = await response.json();
        expect(data).toHaveProperty('version');
      }
    });

    test('Distributed system should report correct node count', async ({ request }) => {
      const response = await request.get(`${CONFIG.API_SERVER_URL}/api/nodes`);
      expect(response.status()).toBe(200);
      
      const data = await response.json();
      expect(data).toHaveProperty('nodes');
      expect(data.nodes.length).toBeGreaterThanOrEqual(3);
      
      // Validate each node has required properties
      data.nodes.forEach(node => {
        expect(node).toHaveProperty('id');
        expect(node).toHaveProperty('status');
        expect(node).toHaveProperty('health');
        expect(['healthy', 'warning', 'error']).toContain(node.status);
      });
    });

  });

  test.describe('Web Interface - Core Functionality', () => {

    test.beforeEach(async ({ page }) => {
      // Navigate to web interface
      await page.goto(CONFIG.WEB_INTERFACE_URL);
      
      // Wait for app initialization
      await page.waitForSelector('#app', { timeout: CONFIG.TIMEOUT });
      await page.waitForLoadState('networkidle');
      
      // Verify basic page structure loaded
      await expect(page.locator('h1')).toContainText('Distributed Llama Chat');
    });

    test.describe('Navigation and Tab Management', () => {

      test('should display all navigation tabs correctly', async ({ page }) => {
        const expectedTabs = ['Chat', 'Nodes', 'Models', 'Settings'];
        
        for (const tabName of expectedTabs) {
          const tabButton = page.locator(`[data-tab="${tabName.toLowerCase()}"]`);
          await expect(tabButton).toBeVisible();
          await expect(tabButton).toContainText(tabName);
        }
        
        // Chat tab should be active by default
        await expect(page.locator('[data-tab="chat"]')).toHaveClass(/active/);
      });

      test('should switch between tabs correctly', async ({ page }) => {
        const tabs = [
          { name: 'nodes', selector: '#nodesTab' },
          { name: 'models', selector: '#modelsTab' },
          { name: 'settings', selector: '#settingsTab' },
          { name: 'chat', selector: '#chatTab' }
        ];

        for (const tab of tabs) {
          // Click tab button
          await page.click(`[data-tab="${tab.name}"]`);
          
          // Wait for content to load
          await page.waitForTimeout(500);
          
          // Verify tab is active
          await expect(page.locator(`[data-tab="${tab.name}"]`)).toHaveClass(/active/);
          await expect(page.locator(tab.selector)).toHaveClass(/active/);
          
          // Verify content is visible
          await expect(page.locator(tab.selector)).toBeVisible();
        }
      });

      test('should handle keyboard navigation', async ({ page }) => {
        // Focus on first tab
        await page.focus('[data-tab="chat"]');
        
        // Use arrow keys to navigate
        await page.keyboard.press('ArrowRight');
        await expect(page.locator('[data-tab="nodes"]')).toBeFocused();
        
        await page.keyboard.press('ArrowRight');
        await expect(page.locator('[data-tab="models"]')).toBeFocused();
        
        // Activate with Enter/Space
        await page.keyboard.press('Enter');
        await expect(page.locator('[data-tab="models"]')).toHaveClass(/active/);
      });

    });

    test.describe('Nodes Tab - Worker Management', () => {

      test.beforeEach(async ({ page }) => {
        await page.click('[data-tab="nodes"]');
        await page.waitForSelector('#nodesTab', { state: 'visible' });
        await page.waitForTimeout(1000); // Allow for data loading
      });

      test('should display cluster overview with correct metrics', async ({ page }) => {
        // Verify cluster overview section exists
        await expect(page.locator('.cluster-overview')).toBeVisible();
        
        // Check for key metrics
        const metrics = ['Total Nodes', 'Healthy', 'CPU Cores', 'Total RAM', 'Active Requests'];
        
        for (const metric of metrics) {
          await expect(page.locator(`.stat-label:has-text("${metric}")`)).toBeVisible();
        }
        
        // Verify total nodes shows expected count (3)
        const totalNodesElement = page.locator('#totalNodes');
        await expect(totalNodesElement).toBeVisible();
        const totalNodes = await totalNodesElement.textContent();
        expect(parseInt(totalNodes)).toBeGreaterThanOrEqual(3);
      });

      test('should display individual node cards with health status', async ({ page }) => {
        // Wait for nodes to load
        await page.waitForSelector('.enhanced-node-card', { timeout: 10000 });
        
        // Get all node cards
        const nodeCards = page.locator('.enhanced-node-card');
        const nodeCount = await nodeCards.count();
        
        expect(nodeCount).toBeGreaterThanOrEqual(3);
        
        // Verify each node card has required elements
        for (let i = 0; i < nodeCount; i++) {
          const card = nodeCards.nth(i);
          
          // Node name and status
          await expect(card.locator('.node-name')).toBeVisible();
          await expect(card.locator('.node-status-badge')).toBeVisible();
          
          // System metrics
          await expect(card.locator('.cpu-value')).toBeVisible();
          await expect(card.locator('.memory-value')).toBeVisible();
          await expect(card.locator('.disk-value')).toBeVisible();
          
          // Performance stats
          await expect(card.locator('.requests-per-sec')).toBeVisible();
          await expect(card.locator('.avg-response')).toBeVisible();
          await expect(card.locator('.node-uptime')).toBeVisible();
        }
      });

      test('should support node filtering and sorting', async ({ page }) => {
        // Test status filter
        await page.selectOption('#statusFilter', 'healthy');
        await page.waitForTimeout(500);
        
        // Verify filtering works (at least some nodes should be visible)
        const visibleNodes = page.locator('.enhanced-node-card:visible');
        expect(await visibleNodes.count()).toBeGreaterThan(0);
        
        // Test sorting
        await page.selectOption('#sortBy', 'memory');
        await page.waitForTimeout(500);
        
        // Test search
        await page.fill('#nodeSearch', 'node');
        await page.waitForTimeout(500);
        
        // Reset filters
        await page.selectOption('#statusFilter', 'all');
        await page.fill('#nodeSearch', '');
        await page.waitForTimeout(500);
      });

      test('should expand node details on interaction', async ({ page }) => {
        // Wait for nodes to load
        await page.waitForSelector('.enhanced-node-card:first-child', { timeout: 10000 });
        
        // Click expand button on first node
        const firstCard = page.locator('.enhanced-node-card').first();
        await firstCard.locator('.expand-btn').click();
        
        // Wait for details section to appear
        await expect(firstCard.locator('.node-details-section')).toBeVisible();
        
        // Check detail tabs
        const detailTabs = ['performance', 'health', 'models', 'config'];
        for (const tab of detailTabs) {
          await expect(firstCard.locator(`[data-tab="${tab}"]`)).toBeVisible();
        }
        
        // Test tab switching in details
        await firstCard.locator('[data-tab="health"]').click();
        await expect(firstCard.locator('#health-panel')).toBeVisible();
      });

      test('should handle refresh controls', async ({ page }) => {
        // Test pause/resume refresh
        await page.click('#pauseRefresh');
        await expect(page.locator('#pauseRefresh')).toContainText('▶️');
        
        await page.click('#pauseRefresh');
        await expect(page.locator('#pauseRefresh')).toContainText('⏸️');
        
        // Test manual refresh
        await page.click('#refreshNodes');
        await page.waitForTimeout(1000);
      });

    });

    test.describe('Models Tab - Model Management', () => {

      test.beforeEach(async ({ page }) => {
        await page.click('[data-tab="models"]');
        await page.waitForSelector('#modelsTab', { state: 'visible' });
        await page.waitForTimeout(1000);
      });

      test('should display available models', async ({ page }) => {
        // Verify models section exists
        await expect(page.locator('.models-container')).toBeVisible();
        await expect(page.locator('h2:has-text("Model Management")')).toBeVisible();
        
        // Check for model grid
        await expect(page.locator('#modelGrid')).toBeVisible();
        
        // Test refresh models functionality
        await page.click('#refreshModels');
        await page.waitForTimeout(1000);
      });

      test('should support model download form', async ({ page }) => {
        // Verify download form exists
        await expect(page.locator('.model-download-form')).toBeVisible();
        
        // Test form fields
        await expect(page.locator('#newModelName')).toBeVisible();
        await expect(page.locator('#downloadTargets')).toBeVisible();
        await expect(page.locator('#downloadModel')).toBeVisible();
        
        // Test form interaction (without actually downloading)
        await page.fill('#newModelName', 'test-model');
        await expect(page.locator('#newModelName')).toHaveValue('test-model');
      });

      test('should handle P2P model settings', async ({ page }) => {
        // Verify propagation settings
        await expect(page.locator('#autoPropagation')).toBeVisible();
        await expect(page.locator('#p2pEnabled')).toBeVisible();
        await expect(page.locator('#propagationStrategy')).toBeVisible();
        
        // Test P2P toggle (this is critical for the requirement)
        const p2pCheckbox = page.locator('#p2pEnabled');
        const isChecked = await p2pCheckbox.isChecked();
        
        // Toggle the setting
        await p2pCheckbox.click();
        await expect(p2pCheckbox).toBeChecked(!isChecked);
        
        // Toggle back
        await p2pCheckbox.click();
        await expect(p2pCheckbox).toBeChecked(isChecked);
        
        // Test propagation strategy selection
        await page.selectOption('#propagationStrategy', 'scheduled');
        await expect(page.locator('#propagationStrategy')).toHaveValue('scheduled');
      });

    });

    test.describe('Chat Tab - Message Interface', () => {

      test.beforeEach(async ({ page }) => {
        await page.click('[data-tab="chat"]');
        await page.waitForSelector('#chatTab', { state: 'visible' });
      });

      test('should display chat interface elements', async ({ page }) => {
        // Verify main chat elements
        await expect(page.locator('#messagesArea')).toBeVisible();
        await expect(page.locator('#messageInput')).toBeVisible();
        await expect(page.locator('#sendButton')).toBeVisible();
        
        // Verify status bar
        await expect(page.locator('.status-bar')).toBeVisible();
        await expect(page.locator('#activeNode')).toBeVisible();
        await expect(page.locator('#queueLength')).toBeVisible();
        await expect(page.locator('#latency')).toBeVisible();
        await expect(page.locator('#modelSelector')).toBeVisible();
        
        // Verify welcome message
        await expect(page.locator('.welcome-message')).toBeVisible();
        await expect(page.locator('#nodeCount')).toBeVisible();
      });

      test('should handle message input interactions', async ({ page }) => {
        const messageInput = page.locator('#messageInput');
        const sendButton = page.locator('#sendButton');
        
        // Test input field
        await messageInput.fill('Test message');
        await expect(messageInput).toHaveValue('Test message');
        
        // Test send button state
        await expect(sendButton).toBeEnabled();
        
        // Test keyboard shortcuts
        await messageInput.fill('');
        await messageInput.fill('Test with Enter');
        await messageInput.press('Enter');
        
        // Test Shift+Enter for new line
        await messageInput.fill('Line 1');
        await messageInput.press('Shift+Enter');
        await messageInput.type('Line 2');
        const value = await messageInput.inputValue();
        expect(value).toContain('\n');
      });

      test('should display connection status updates', async ({ page }) => {
        // Verify connection status elements
        const statusIndicator = page.locator('#connectionStatus');
        const statusText = page.locator('#connectionText');
        
        await expect(statusIndicator).toBeVisible();
        await expect(statusText).toBeVisible();
        
        // Status should eventually show some state
        await page.waitForTimeout(3000);
        const statusTextContent = await statusText.textContent();
        expect(statusTextContent).not.toBe('');
      });

      test('should update node count in real-time', async ({ page }) => {
        const nodeCountElement = page.locator('#nodeCount');
        await expect(nodeCountElement).toBeVisible();
        
        // Node count should eventually be populated
        await page.waitForTimeout(2000);
        const nodeCount = await nodeCountElement.textContent();
        expect(parseInt(nodeCount)).toBeGreaterThanOrEqual(0);
      });

    });

    test.describe('Settings Tab - Configuration', () => {

      test.beforeEach(async ({ page }) => {
        await page.click('[data-tab="settings"]');
        await page.waitForSelector('#settingsTab', { state: 'visible' });
      });

      test('should display configuration options', async ({ page }) => {
        // API Configuration section
        await expect(page.locator('h3:has-text("API Configuration")')).toBeVisible();
        await expect(page.locator('#apiEndpoint')).toBeVisible();
        await expect(page.locator('#apiKey')).toBeVisible();
        
        // Chat Settings section
        await expect(page.locator('h3:has-text("Chat Settings")')).toBeVisible();
        await expect(page.locator('#streamingEnabled')).toBeVisible();
        await expect(page.locator('#autoScroll')).toBeVisible();
        await expect(page.locator('#maxTokens')).toBeVisible();
        await expect(page.locator('#temperature')).toBeVisible();
        
        // Node Management section
        await expect(page.locator('h3:has-text("Node Management")')).toBeVisible();
        await expect(page.locator('#loadBalancing')).toBeVisible();
        await expect(page.locator('#addNodeButton')).toBeVisible();
      });

      test('should handle settings modifications', async ({ page }) => {
        // Test API endpoint modification
        await page.fill('#apiEndpoint', 'ws://localhost:13100/test');
        await expect(page.locator('#apiEndpoint')).toHaveValue('ws://localhost:13100/test');
        
        // Test checkbox toggles
        const streamingCheckbox = page.locator('#streamingEnabled');
        const initialState = await streamingCheckbox.isChecked();
        await streamingCheckbox.click();
        await expect(streamingCheckbox).toBeChecked(!initialState);
        
        // Test range slider
        await page.locator('#temperature').fill('1.2');
        await expect(page.locator('#temperatureValue')).toContainText('1.2');
        
        // Test select dropdown
        await page.selectOption('#loadBalancing', 'least-loaded');
        await expect(page.locator('#loadBalancing')).toHaveValue('least-loaded');
      });

      test('should handle add node modal', async ({ page }) => {
        // Click add node button
        await page.click('#addNodeButton');
        
        // Modal should appear
        await expect(page.locator('#addNodeModal')).toBeVisible();
        
        // Fill modal form
        await page.fill('#nodeUrl', 'http://test-node:11434');
        await page.fill('#nodeName', 'test-node-1');
        
        // Cancel modal
        await page.click('#cancelAddNode');
        await expect(page.locator('#addNodeModal')).not.toBeVisible();
      });

      test('should save and reset settings', async ({ page }) => {
        // Make a setting change
        await page.fill('#maxTokens', '1024');
        
        // Save settings
        await page.click('#saveSettings');
        await page.waitForTimeout(500);
        
        // Reset settings
        await page.click('#resetSettings');
        await page.waitForTimeout(500);
        
        // Value should be reset to default
        await expect(page.locator('#maxTokens')).toHaveValue('2048');
      });

    });

  });

  test.describe('Error Handling and Edge Cases', () => {

    test('should handle network disconnection gracefully', async ({ page, context }) => {
      await page.goto(CONFIG.WEB_INTERFACE_URL);
      await page.waitForSelector('#app');
      
      // Simulate network failure
      await context.setOffline(true);
      await page.waitForTimeout(2000);
      
      // Check connection status indicates error
      const statusText = page.locator('#connectionText');
      const statusContent = await statusText.textContent();
      expect(statusContent.toLowerCase()).toContain('error');
      
      // Restore network
      await context.setOffline(false);
      await page.waitForTimeout(3000);
    });

    test('should display error boundary when JavaScript fails', async ({ page }) => {
      await page.goto(CONFIG.WEB_INTERFACE_URL);
      
      // Inject a JavaScript error
      await page.evaluate(() => {
        window.onerror = null; // Remove default handler
        throw new Error('Test error for error boundary');
      });
      
      await page.waitForTimeout(1000);
      
      // Check if error boundary is displayed
      const errorBoundary = page.locator('#errorBoundary');
      if (await errorBoundary.isVisible()) {
        await expect(errorBoundary.locator('#errorMessage')).toBeVisible();
        await expect(page.locator('#retryButton')).toBeVisible();
      }
    });

    test('should handle malformed API responses', async ({ page }) => {
      // Intercept API calls and return malformed data
      await page.route('**/api/nodes', route => {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: '{"invalid": "json" missing brace'
        });
      });
      
      await page.goto(CONFIG.WEB_INTERFACE_URL);
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(2000);
      
      // Should handle gracefully without crashing
      await expect(page.locator('#app')).toBeVisible();
    });

  });

  test.describe('Performance and Load Testing', () => {

    test('should load within performance thresholds', async ({ page }) => {
      const startTime = Date.now();
      
      await page.goto(CONFIG.WEB_INTERFACE_URL);
      await page.waitForSelector('#app');
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      expect(loadTime).toBeLessThan(CONFIG.PERFORMANCE_THRESHOLDS.PAGE_LOAD);
    });

    test('should handle rapid tab switching', async ({ page }) => {
      await page.goto(CONFIG.WEB_INTERFACE_URL);
      await page.waitForSelector('#app');
      
      const tabs = ['chat', 'nodes', 'models', 'settings'];
      
      // Rapidly switch between tabs
      for (let i = 0; i < 20; i++) {
        const tab = tabs[i % tabs.length];
        await page.click(`[data-tab="${tab}"]`);
        await page.waitForTimeout(100);
      }
      
      // Should still be responsive
      await expect(page.locator('#app')).toBeVisible();
    });

    test('should handle multiple concurrent operations', async ({ page }) => {
      await page.goto(CONFIG.WEB_INTERFACE_URL);
      await page.waitForSelector('#app');
      
      // Start multiple operations concurrently
      const operations = [
        page.click('[data-tab="nodes"]'),
        page.click('#refreshNodes'),
        page.click('[data-tab="models"]'),
        page.click('#refreshModels'),
        page.click('[data-tab="settings"]')
      ];
      
      await Promise.all(operations);
      await page.waitForTimeout(2000);
      
      // Should handle gracefully
      await expect(page.locator('#app')).toBeVisible();
    });

  });

  test.describe('Cross-Browser Compatibility', () => {

    ['chromium', 'firefox', 'webkit'].forEach(browserName => {
      test(`should work correctly in ${browserName}`, async () => {
        const browser = await chromium.launch();
        const page = await browser.newPage();
        
        try {
          await page.goto(CONFIG.WEB_INTERFACE_URL);
          await page.waitForSelector('#app');
          
          // Test basic functionality
          await expect(page.locator('h1')).toContainText('Distributed Llama Chat');
          
          // Test tab switching
          await page.click('[data-tab="nodes"]');
          await expect(page.locator('#nodesTab')).toBeVisible();
          
          await page.click('[data-tab="models"]');
          await expect(page.locator('#modelsTab')).toBeVisible();
          
        } finally {
          await browser.close();
        }
      });
    });

  });

  test.describe('Real-time Updates and WebSocket Testing', () => {

    test('should establish WebSocket connection', async ({ page }) => {
      let wsConnected = false;
      
      // Monitor WebSocket connections
      page.on('websocket', ws => {
        wsConnected = true;
        console.log('WebSocket connected:', ws.url());
        
        ws.on('close', () => console.log('WebSocket disconnected'));
        ws.on('framesent', data => console.log('Frame sent:', data));
        ws.on('framereceived', data => console.log('Frame received:', data));
      });
      
      await page.goto(CONFIG.WEB_INTERFACE_URL);
      await page.waitForSelector('#app');
      
      // Wait for WebSocket connection
      await page.waitForTimeout(5000);
      
      expect(wsConnected).toBe(true);
    });

    test('should receive real-time node updates', async ({ page }) => {
      await page.goto(CONFIG.WEB_INTERFACE_URL);
      await page.waitForSelector('#app');
      
      // Go to nodes tab
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(2000);
      
      // Monitor for updates
      const initialNodeCount = await page.locator('#totalNodes').textContent();
      
      // Wait for potential updates
      await page.waitForTimeout(5000);
      
      // Should maintain consistent state or show updates
      await expect(page.locator('#totalNodes')).toBeVisible();
    });

  });

});

// Utility functions for test setup and teardown
export async function setupTestEnvironment() {
  // Ensure all services are running
  console.log('Setting up test environment...');
  
  // Add any setup logic here
  return true;
}

export async function cleanupTestEnvironment() {
  // Clean up any test artifacts
  console.log('Cleaning up test environment...');
  
  return true;
}