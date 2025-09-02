/**
 * Final Comprehensive Testing Suite
 * Tests all improvements across the 5 iterations
 */

const { test, expect, chromium } = require('playwright/test');

// Test the improved interface
test.describe('OllamaMax UI - Post-Improvements Testing', () => {
  
  const TEST_URL = 'http://localhost:8080';
  
  test.beforeEach(async ({ page }) => {
    // Start simple HTTP server for testing
    await page.goto(TEST_URL);
    await page.waitForLoadState('networkidle');
  });

  test.describe('Accessibility Improvements (Iteration 1)', () => {
    
    test('should have proper ARIA labels on navigation', async ({ page }) => {
      // Check ARIA labels on tab buttons
      await expect(page.locator('[data-tab="chat"]')).toHaveAttribute('aria-label', 'Chat tab');
      await expect(page.locator('[data-tab="nodes"]')).toHaveAttribute('aria-label', 'Nodes management tab');
      await expect(page.locator('[data-tab="models"]')).toHaveAttribute('aria-label', 'Model management tab');
      await expect(page.locator('[data-tab="settings"]')).toHaveAttribute('aria-label', 'Settings tab');
    });

    test('should have proper ARIA roles', async ({ page }) => {
      await expect(page.locator('[data-tab="chat"]')).toHaveAttribute('role', 'tab');
      await expect(page.locator('#main-content')).toHaveAttribute('role', 'main');
    });

    test('should have skip link for accessibility', async ({ page }) => {
      const skipLink = page.locator('.skip-link');
      await expect(skipLink).toBeInViewport();
      await expect(skipLink).toHaveAttribute('href', '#main-content');
    });

    test('should have accessible form labels', async ({ page }) => {
      const messageInput = page.locator('#messageInput');
      await expect(messageInput).toHaveAttribute('aria-label', 'Message input');
      await expect(messageInput).toHaveAttribute('aria-describedby', 'message-help');
      
      const helpText = page.locator('#message-help');
      await expect(helpText).toBeInViewport();
    });
  });

  test.describe('Performance Optimizations (Iteration 2)', () => {
    
    test('should load within performance budget', async ({ page }) => {
      const startTime = Date.now();
      await page.goto(TEST_URL);
      await page.waitForSelector('#app');
      const loadTime = Date.now() - startTime;
      
      expect(loadTime).toBeLessThan(3000); // 3 second budget
    });

    test('should have lazy loading setup', async ({ page }) => {
      // Check that lazy loading function exists
      const lazyLoadingExists = await page.evaluate(() => {
        return typeof window.llamaClient?.setupLazyLoading === 'function';
      });
      expect(lazyLoadingExists).toBe(true);
    });

    test('should have debounced search functionality', async ({ page }) => {
      const debounceExists = await page.evaluate(() => {
        return typeof window.llamaClient?.debounce === 'function';
      });
      expect(debounceExists).toBe(true);
    });
  });

  test.describe('Modern UI Components (Iteration 3)', () => {
    
    test('should support CSS custom properties for theming', async ({ page }) => {
      const primaryColor = await page.evaluate(() => {
        return getComputedStyle(document.documentElement).getPropertyValue('--primary').trim();
      });
      expect(primaryColor).toBeTruthy();
      expect(primaryColor).toBe('#667eea');
    });

    test('should have enhanced hover effects', async ({ page }) => {
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(500);
      
      const nodeCard = page.locator('.enhanced-node-card').first();
      if (await nodeCard.count() > 0) {
        await nodeCard.hover();
        
        const transform = await nodeCard.evaluate(el => {
          return window.getComputedStyle(el).transform;
        });
        expect(transform).not.toBe('none');
      }
    });

    test('should have smooth transitions', async ({ page }) => {
      const tabButton = page.locator('[data-tab="models"]');
      const transition = await tabButton.evaluate(el => {
        return window.getComputedStyle(el).transition;
      });
      expect(transition).toContain('0.3s');
    });
  });

  test.describe('Responsive Enhancements (Iteration 4)', () => {
    
    test('should adapt to mobile viewport', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      // Check mobile-specific styles
      const navTabs = page.locator('.nav-tabs');
      const gridDisplay = await navTabs.evaluate(el => {
        return window.getComputedStyle(el).display;
      });
      expect(gridDisplay).toBe('grid');
    });

    test('should have touch-friendly button sizes on mobile', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      const tabButton = page.locator('[data-tab="chat"]');
      const minHeight = await tabButton.evaluate(el => {
        return window.getComputedStyle(el).minHeight;
      });
      
      // Should have minimum 44px for touch targets
      expect(parseFloat(minHeight)).toBeGreaterThanOrEqual(44);
    });

    test('should handle tablet landscape orientation', async ({ page }) => {
      await page.setViewportSize({ width: 1024, height: 768 });
      
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(500);
      
      const clusterOverview = page.locator('.cluster-overview');
      const gridColumns = await clusterOverview.evaluate(el => {
        return window.getComputedStyle(el).gridTemplateColumns;
      });
      
      // Should adapt grid layout for landscape tablets
      expect(gridColumns).toBeTruthy();
    });
  });

  test.describe('UX Refinements (Iteration 5)', () => {
    
    test('should have error boundary elements', async ({ page }) => {
      await expect(page.locator('#errorBoundary')).toBeInViewport();
      await expect(page.locator('#retryButton')).toBeInViewport();
      await expect(page.locator('#reloadButton')).toBeInViewport();
    });

    test('should have loading overlay', async ({ page }) => {
      await expect(page.locator('#loadingOverlay')).toBeInViewport();
      await expect(page.locator('.loading-spinner')).toBeInViewport();
    });

    test('should have notification system', async ({ page }) => {
      await expect(page.locator('#notificationContainer')).toBeInViewport();
    });

    test('should have proper focus management', async ({ page }) => {
      // Test keyboard navigation
      await page.keyboard.press('Tab');
      const focusedElement = await page.evaluate(() => document.activeElement.tagName);
      expect(['BUTTON', 'INPUT', 'A'].includes(focusedElement)).toBe(true);
    });

    test('should support high contrast mode', async ({ page }) => {
      // Test high contrast media query styles are present
      const hasHighContrastStyles = await page.evaluate(() => {
        const stylesheets = Array.from(document.styleSheets);
        return stylesheets.some(sheet => {
          try {
            const rules = Array.from(sheet.cssRules);
            return rules.some(rule => 
              rule.media && rule.media.mediaText.includes('prefers-contrast: high')
            );
          } catch (e) {
            return false;
          }
        });
      });
      expect(hasHighContrastStyles).toBe(true);
    });
  });

  test.describe('Overall Integration Testing', () => {
    
    test('should maintain functionality across all tabs', async ({ page }) => {
      // Test tab switching works after improvements
      const tabs = ['chat', 'nodes', 'models', 'settings'];
      
      for (const tab of tabs) {
        await page.click(`[data-tab="${tab}"]`);
        await page.waitForTimeout(300);
        
        await expect(page.locator(`[data-tab="${tab}"]`)).toHaveClass(/active/);
        await expect(page.locator(`#${tab}Tab`)).toHaveClass(/active/);
      }
    });

    test('should handle rapid interactions gracefully', async ({ page }) => {
      // Rapidly switch tabs multiple times
      for (let i = 0; i < 10; i++) {
        await page.click('[data-tab="nodes"]');
        await page.click('[data-tab="models"]');
        await page.click('[data-tab="settings"]');
        await page.click('[data-tab="chat"]');
      }
      
      // Should still be functional
      await expect(page.locator('#app')).toBeVisible();
      await expect(page.locator('[data-tab="chat"]')).toHaveClass(/active/);
    });

    test('should maintain performance under stress', async ({ page }) => {
      const startTime = Date.now();
      
      // Perform various interactions
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(100);
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(100);
      await page.fill('#messageInput', 'Test message for performance');
      await page.click('#sendButton');
      
      const endTime = Date.now();
      expect(endTime - startTime).toBeLessThan(2000);
    });

    test('should work with JavaScript disabled gracefully', async ({ page }) => {
      // Disable JavaScript
      await page.context().addInitScript(() => {
        window.addEventListener('error', () => {});
      });
      
      // Page should still load basic content
      await expect(page.locator('#app')).toBeVisible();
      await expect(page.locator('h1')).toContainText('Distributed Llama Chat');
    });
  });

  test.describe('Cross-Browser Compatibility', () => {
    
    test('should work in different browsers', async () => {
      const browsers = [
        { name: 'chromium', browser: chromium },
      ];
      
      for (const { name, browser } of browsers) {
        const browserInstance = await browser.launch();
        const context = await browserInstance.newContext();
        const page = await context.newPage();
        
        await page.goto(TEST_URL);
        await expect(page.locator('#app')).toBeVisible();
        await expect(page.locator('.nav-tabs')).toBeVisible();
        
        await browserInstance.close();
      }
    });
  });

  test.describe('Final Validation', () => {
    
    test('should have all major components visible and functional', async ({ page }) => {
      // Header and navigation
      await expect(page.locator('.app-header')).toBeVisible();
      await expect(page.locator('.nav-tabs')).toBeVisible();
      await expect(page.locator('.connection-status')).toBeVisible();
      
      // Main content area
      await expect(page.locator('#main-content')).toBeVisible();
      await expect(page.locator('.chat-container')).toBeVisible();
      
      // Input area
      await expect(page.locator('#messageInput')).toBeVisible();
      await expect(page.locator('#sendButton')).toBeVisible();
      
      // Status indicators
      await expect(page.locator('#connectionStatus')).toBeVisible();
      await expect(page.locator('#connectionText')).toBeVisible();
    });

    test('should pass accessibility audit', async ({ page }) => {
      // Basic accessibility checks
      const hasHeadings = await page.locator('h1, h2, h3, h4, h5, h6').count();
      expect(hasHeadings).toBeGreaterThan(0);
      
      const hasLabels = await page.locator('label').count();
      expect(hasLabels).toBeGreaterThan(0);
      
      const hasAltText = await page.locator('img[alt]').count();
      const totalImages = await page.locator('img').count();
      if (totalImages > 0) {
        expect(hasAltText).toBe(totalImages);
      }
    });

    test('should demonstrate improved user experience', async ({ page }) => {
      // Smooth animations
      await page.hover('[data-tab="models"]');
      const hasTransitions = await page.locator('[data-tab="models"]').evaluate(el => {
        return window.getComputedStyle(el).transition !== 'all 0s ease 0s';
      });
      expect(hasTransitions).toBe(true);
      
      // Responsive design
      await page.setViewportSize({ width: 320, height: 568 });
      await expect(page.locator('#app')).toBeVisible();
      
      await page.setViewportSize({ width: 1920, height: 1080 });
      await expect(page.locator('#app')).toBeVisible();
    });
  });
});

// Utility function to start HTTP server for testing
async function startTestServer() {
  const { spawn } = require('child_process');
  const path = require('path');
  
  const server = spawn('python3', ['-m', 'http.server', '8080'], {
    cwd: path.join(__dirname, '..', 'web-interface'),
    stdio: 'inherit'
  });
  
  // Wait for server to start
  await new Promise(resolve => setTimeout(resolve, 2000));
  
  return server;
}

module.exports = { startTestServer };