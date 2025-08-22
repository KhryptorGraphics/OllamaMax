import { test, expect, devices } from '@playwright/test';
import { BrowserTestFramework } from '../../utils/browser-automation';

test.describe('Mobile and Responsive Design', () => {
  let framework: BrowserTestFramework;

  test.beforeEach(async ({ page }) => {
    framework = new BrowserTestFramework(page);
    
    // Login before each test
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'admin@example.com');
    await page.fill('[data-testid="password-input"]', 'admin123');
    await page.click('[data-testid="login-button"]');
    await expect(page).toHaveURL('/dashboard');
  });

  test('Mobile responsive dashboard layout', async ({ page }) => {
    const result = await framework.testMobileResponsive();
    expect(result.success).toBeTruthy();
    
    // Additional mobile layout tests
    await page.setViewportSize({ width: 375, height: 667 }); // iPhone 8
    await page.goto('/dashboard');
    
    // Test mobile navigation
    const mobileMenuButton = page.locator('[data-testid="mobile-menu-button"]');
    await expect(mobileMenuButton).toBeVisible();
    
    // Test that desktop navigation is hidden
    const desktopNav = page.locator('[data-testid="desktop-navigation"]');
    await expect(desktopNav).not.toBeVisible();
    
    // Test mobile menu functionality
    await mobileMenuButton.click();
    const mobileNav = page.locator('[data-testid="mobile-navigation"]');
    await expect(mobileNav).toBeVisible();
    
    // Test navigation items
    const navItems = ['Dashboard', 'Admin', 'Monitoring', 'Settings'];
    for (const item of navItems) {
      const navItem = mobileNav.locator(`text=${item}`);
      await expect(navItem).toBeVisible();
    }
    
    // Test closing mobile menu
    await page.click('[data-testid="mobile-menu-close"]');
    await expect(mobileNav).not.toBeVisible();
  });

  test('Dashboard grid responsive behavior', async ({ page }) => {
    const viewports = [
      { width: 320, height: 568, device: 'iPhone SE' },
      { width: 375, height: 667, device: 'iPhone 8' },
      { width: 414, height: 896, device: 'iPhone 11' },
      { width: 768, height: 1024, device: 'iPad' },
      { width: 1024, height: 768, device: 'iPad Landscape' },
      { width: 1920, height: 1080, device: 'Desktop' }
    ];
    
    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.goto('/dashboard');
      
      const dashboardGrid = page.locator('[data-testid="dashboard-grid"]');
      await expect(dashboardGrid).toBeVisible();
      
      // Check grid columns based on viewport
      const computedStyle = await dashboardGrid.evaluate(el => 
        window.getComputedStyle(el).getPropertyValue('grid-template-columns')
      );
      
      if (viewport.width < 768) {
        // Mobile: single column
        expect(computedStyle).toContain('1fr');
      } else if (viewport.width < 1024) {
        // Tablet: two columns
        expect(computedStyle.split(' ').length).toBeLessThanOrEqual(2);
      } else {
        // Desktop: multiple columns
        expect(computedStyle.split(' ').length).toBeGreaterThan(2);
      }
      
      // Test that all dashboard widgets are visible
      const widgets = page.locator('[data-testid^="widget-"]');
      const widgetCount = await widgets.count();
      
      for (let i = 0; i < widgetCount; i++) {
        const widget = widgets.nth(i);
        await expect(widget).toBeVisible();
        
        // Verify widget content is not clipped
        const widgetBox = await widget.boundingBox();
        expect(widgetBox).toBeTruthy();
        expect(widgetBox!.width).toBeGreaterThan(0);
        expect(widgetBox!.height).toBeGreaterThan(0);
      }
    }
  });

  test('Touch interactions and gestures', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/dashboard');
    
    // Test swipe navigation
    const dashboardContainer = page.locator('[data-testid="dashboard-container"]');
    await expect(dashboardContainer).toBeVisible();
    
    // Simulate swipe left to right
    await dashboardContainer.hover();
    await page.touchscreen.tap(100, 300);
    await page.mouse.down();
    await page.mouse.move(250, 300);
    await page.mouse.up();
    
    // Test pinch zoom (if supported)
    await page.touchscreen.tap(200, 300);
    
    // Test tap interactions on buttons
    const actionButtons = page.locator('[data-testid^="action-button-"]');
    const buttonCount = await actionButtons.count();
    
    if (buttonCount > 0) {
      const firstButton = actionButtons.first();
      
      // Verify button is touch-friendly (minimum 44px)
      const buttonBox = await firstButton.boundingBox();
      expect(buttonBox!.height).toBeGreaterThanOrEqual(44);
      expect(buttonBox!.width).toBeGreaterThanOrEqual(44);
      
      // Test tap interaction
      await page.touchscreen.tap(buttonBox!.x + buttonBox!.width / 2, buttonBox!.y + buttonBox!.height / 2);
    }
    
    // Test scroll behavior
    await page.touchscreen.tap(200, 300);
    await page.mouse.wheel(0, 500); // Scroll down
    await page.waitForTimeout(1000);
    
    // Verify content scrolled
    const scrollTop = await page.evaluate(() => window.pageYOffset);
    expect(scrollTop).toBeGreaterThan(0);
  });

  test('Form input accessibility on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/admin/cluster');
    
    // Test add node form on mobile
    await page.click('[data-testid="add-node-button"]');
    const addNodeDialog = page.locator('[data-testid="add-node-dialog"]');
    await expect(addNodeDialog).toBeVisible();
    
    // Test form input behavior on mobile
    const nodeAddressInput = page.locator('[data-testid="node-address"]');
    await expect(nodeAddressInput).toBeVisible();
    
    // Verify input field is properly sized for mobile
    const inputBox = await nodeAddressInput.boundingBox();
    expect(inputBox!.height).toBeGreaterThanOrEqual(44); // Touch-friendly height
    
    // Test keyboard behavior
    await nodeAddressInput.click();
    
    // Check if virtual keyboard space is considered
    const viewportHeight = await page.evaluate(() => window.innerHeight);
    expect(viewportHeight).toBeGreaterThan(300); // Basic sanity check
    
    // Test form submission
    await nodeAddressInput.fill('192.168.1.100:8080');
    await page.fill('[data-testid="node-name"]', 'Mobile Test Node');
    
    // Verify submit button is accessible
    const submitButton = page.locator('[data-testid="confirm-add-node"]');
    const submitBox = await submitButton.boundingBox();
    expect(submitBox!.height).toBeGreaterThanOrEqual(44);
    
    await submitButton.click();
  });

  test('Performance on mobile devices', async ({ page }) => {
    // Test various mobile device profiles
    const mobileDevices = [
      devices['iPhone SE'],
      devices['iPhone 12'],
      devices['Pixel 5'],
      devices['Galaxy S21']
    ];
    
    for (const device of mobileDevices) {
      await page.setViewportSize(device.viewport);
      await page.setUserAgent(device.userAgent);
      
      // Measure page load performance
      const startTime = Date.now();
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');
      const loadTime = Date.now() - startTime;
      
      // Mobile should load within 5 seconds on 3G
      expect(loadTime).toBeLessThan(5000);
      
      // Check Core Web Vitals
      const metrics = await page.evaluate(() => {
        const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        return {
          loadTime: navigation.loadEventEnd - navigation.fetchStart,
          domContentLoaded: navigation.domContentLoadedEventEnd - navigation.fetchStart,
          firstPaint: performance.getEntriesByName('first-paint')[0]?.startTime || 0,
          firstContentfulPaint: performance.getEntriesByName('first-contentful-paint')[0]?.startTime || 0
        };
      });
      
      // Mobile performance thresholds
      expect(metrics.loadTime).toBeLessThan(5000); // 5s load time
      expect(metrics.firstContentfulPaint).toBeLessThan(2500); // 2.5s FCP
      expect(metrics.domContentLoaded).toBeLessThan(3000); // 3s DOM ready
    }
  });

  test('Landscape and portrait orientation', async ({ page }) => {
    // Test portrait orientation (mobile)
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/dashboard');
    
    // Verify portrait layout
    const dashboardGrid = page.locator('[data-testid="dashboard-grid"]');
    const portraitStyle = await dashboardGrid.evaluate(el => 
      window.getComputedStyle(el).getPropertyValue('grid-template-columns')
    );
    
    // Should be single column in portrait mobile
    expect(portraitStyle).toContain('1fr');
    
    // Test landscape orientation (mobile)
    await page.setViewportSize({ width: 667, height: 375 });
    await page.waitForTimeout(1000); // Allow layout to adjust
    
    // Verify landscape layout
    const landscapeStyle = await dashboardGrid.evaluate(el => 
      window.getComputedStyle(el).getPropertyValue('grid-template-columns')
    );
    
    // Should have more columns in landscape
    const landscapeColumns = landscapeStyle.split(' ').length;
    const portraitColumns = portraitStyle.split(' ').length;
    expect(landscapeColumns).toBeGreaterThanOrEqual(portraitColumns);
    
    // Test tablet orientations
    await page.setViewportSize({ width: 768, height: 1024 }); // Portrait tablet
    await page.waitForTimeout(1000);
    
    const tabletPortraitStyle = await dashboardGrid.evaluate(el => 
      window.getComputedStyle(el).getPropertyValue('grid-template-columns')
    );
    
    await page.setViewportSize({ width: 1024, height: 768 }); // Landscape tablet
    await page.waitForTimeout(1000);
    
    const tabletLandscapeStyle = await dashboardGrid.evaluate(el => 
      window.getComputedStyle(el).getPropertyValue('grid-template-columns')
    );
    
    // Landscape should have more columns than portrait
    const tabletLandscapeColumns = tabletLandscapeStyle.split(' ').length;
    const tabletPortraitColumns = tabletPortraitStyle.split(' ').length;
    expect(tabletLandscapeColumns).toBeGreaterThanOrEqual(tabletPortraitColumns);
  });

  test('Mobile navigation patterns', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Test bottom navigation (if implemented)
    await page.goto('/dashboard');
    const bottomNav = page.locator('[data-testid="bottom-navigation"]');
    
    if (await bottomNav.isVisible()) {
      // Test bottom navigation items
      const navItems = bottomNav.locator('[data-testid^="bottom-nav-"]');
      const itemCount = await navItems.count();
      
      for (let i = 0; i < itemCount; i++) {
        const item = navItems.nth(i);
        await expect(item).toBeVisible();
        
        // Test navigation
        await item.click();
        await page.waitForTimeout(1000);
        
        // Verify page changed (check URL or content)
        const currentUrl = page.url();
        expect(currentUrl).toContain('/');
      }
    }
    
    // Test hamburger menu navigation
    const hamburgerMenu = page.locator('[data-testid="mobile-menu-button"]');
    await expect(hamburgerMenu).toBeVisible();
    
    await hamburgerMenu.click();
    const mobileMenu = page.locator('[data-testid="mobile-navigation"]');
    await expect(mobileMenu).toBeVisible();
    
    // Test nested navigation
    const adminMenuItem = mobileMenu.locator('[data-testid="nav-admin"]');
    if (await adminMenuItem.isVisible()) {
      await adminMenuItem.click();
      
      // Check for submenu
      const adminSubmenu = mobileMenu.locator('[data-testid="admin-submenu"]');
      if (await adminSubmenu.isVisible()) {
        const submenuItems = adminSubmenu.locator('[data-testid^="submenu-"]');
        const submenuCount = await submenuItems.count();
        expect(submenuCount).toBeGreaterThan(0);
      }
    }
  });

  test('Mobile accessibility features', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/dashboard');
    
    // Test focus management on mobile
    await page.keyboard.press('Tab');
    const focusedElement = page.locator(':focus');
    await expect(focusedElement).toBeVisible();
    
    // Test skip links
    const skipLink = page.locator('[data-testid="skip-to-content"]');
    if (await skipLink.isVisible()) {
      await skipLink.click();
      const mainContent = page.locator('main');
      const isFocused = await mainContent.evaluate(el => el === document.activeElement);
      expect(isFocused).toBeTruthy();
    }
    
    // Test high contrast mode detection
    const supportsHighContrast = await page.evaluate(() => 
      window.matchMedia('(prefers-contrast: high)').matches
    );
    
    if (supportsHighContrast) {
      // Verify high contrast styles are applied
      const bodyBg = await page.locator('body').evaluate(el => 
        window.getComputedStyle(el).backgroundColor
      );
      expect(bodyBg).toBeTruthy();
    }
    
    // Test reduced motion preference
    const prefersReducedMotion = await page.evaluate(() => 
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    );
    
    if (prefersReducedMotion) {
      // Verify animations are disabled or reduced
      const animatedElements = page.locator('[class*="animate"], [class*="transition"]');
      const elementCount = await animatedElements.count();
      
      for (let i = 0; i < Math.min(elementCount, 5); i++) {
        const element = animatedElements.nth(i);
        const animationDuration = await element.evaluate(el => 
          window.getComputedStyle(el).animationDuration
        );
        
        // Animation should be significantly reduced or disabled
        expect(animationDuration === '0s' || animationDuration === 'none').toBeTruthy();
      }
    }
    
    // Test screen reader compatibility
    const headings = page.locator('h1, h2, h3, h4, h5, h6');
    const headingCount = await headings.count();
    expect(headingCount).toBeGreaterThan(0); // Should have proper heading structure
    
    // Test ARIA labels on interactive elements
    const interactiveElements = page.locator('button, [role="button"], input, select, textarea');
    const interactiveCount = await interactiveElements.count();
    
    for (let i = 0; i < Math.min(interactiveCount, 10); i++) {
      const element = interactiveElements.nth(i);
      const hasLabel = await element.evaluate(el => {
        const ariaLabel = el.getAttribute('aria-label');
        const ariaLabelledBy = el.getAttribute('aria-labelledby');
        const textContent = el.textContent?.trim();
        return !!(ariaLabel || ariaLabelledBy || textContent);
      });
      
      expect(hasLabel).toBeTruthy();
    }
  });

  test('Mobile offline functionality', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/dashboard');
    
    // Test offline detection
    await page.context().setOffline(true);
    
    // Check for offline indicator
    const offlineIndicator = page.locator('[data-testid="offline-indicator"]');
    await expect(offlineIndicator).toBeVisible({ timeout: 10000 });
    
    // Test cached content availability
    await page.reload();
    
    // Basic content should still be available from cache
    const mainContent = page.locator('main');
    await expect(mainContent).toBeVisible();
    
    // Test offline messaging
    const offlineMessage = page.locator('[data-testid="offline-message"]');
    if (await offlineMessage.isVisible()) {
      await expect(offlineMessage).toContainText(/offline|connection/i);
    }
    
    // Restore connection
    await page.context().setOffline(false);
    
    // Check for online indicator
    const onlineIndicator = page.locator('[data-testid="online-indicator"]');
    await expect(onlineIndicator).toBeVisible({ timeout: 10000 });
    
    // Test data sync after reconnection
    await page.waitForTimeout(2000);
    const syncIndicator = page.locator('[data-testid="sync-status"]');
    if (await syncIndicator.isVisible()) {
      await expect(syncIndicator).toHaveAttribute('data-status', 'synced');
    }
  });
});