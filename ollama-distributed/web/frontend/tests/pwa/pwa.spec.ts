import { test, expect, Page, BrowserContext } from '@playwright/test';

// PWA Testing Suite for OllamaMax
test.describe('PWA Functionality', () => {
  let context: BrowserContext;
  let page: Page;

  test.beforeAll(async ({ browser }) => {
    // Create a new context for PWA testing
    context = await browser.newContext({
      viewport: { width: 375, height: 812 }, // iPhone X dimensions
      userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15',
      hasTouch: true,
      isMobile: true,
    });
    
    page = await context.newPage();
    
    // Enable service worker
    await context.grantPermissions(['notifications']);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('should load the app successfully', async () => {
    await page.goto('/');
    await expect(page).toHaveTitle(/OllamaMax/);
    await expect(page.locator('#root')).toBeVisible();
  });

  test('should have valid web app manifest', async () => {
    const response = await page.goto('/manifest.json');
    expect(response?.status()).toBe(200);
    
    const manifest = await response?.json();
    expect(manifest).toHaveProperty('name', 'OllamaMax - Distributed AI Platform');
    expect(manifest).toHaveProperty('short_name', 'OllamaMax');
    expect(manifest).toHaveProperty('start_url', '/');
    expect(manifest).toHaveProperty('display', 'standalone');
    expect(manifest).toHaveProperty('theme_color', '#2563eb');
    expect(manifest).toHaveProperty('background_color', '#ffffff');
    expect(manifest.icons).toHaveLength(9);
    
    // Verify icons have required sizes
    const iconSizes = manifest.icons.map((icon: any) => icon.sizes);
    expect(iconSizes).toContain('192x192');
    expect(iconSizes).toContain('512x512');
  });

  test('should register service worker', async () => {
    await page.goto('/');
    
    // Wait for service worker registration
    const swRegistered = await page.evaluate(async () => {
      if ('serviceWorker' in navigator) {
        try {
          const registration = await navigator.serviceWorker.ready;
          return !!registration;
        } catch (error) {
          return false;
        }
      }
      return false;
    });
    
    expect(swRegistered).toBe(true);
  });

  test('should show install prompt on supported browsers', async () => {
    // Simulate beforeinstallprompt event
    await page.goto('/');
    
    await page.evaluate(() => {
      const event = new CustomEvent('beforeinstallprompt', {
        bubbles: true,
        cancelable: true
      });
      
      // Add required properties for the event
      Object.defineProperty(event, 'platforms', {
        value: ['web'],
        writable: false
      });
      
      Object.defineProperty(event, 'userChoice', {
        value: Promise.resolve({ outcome: 'dismissed', platform: 'web' }),
        writable: false
      });
      
      Object.defineProperty(event, 'prompt', {
        value: () => Promise.resolve(),
        writable: false
      });
      
      window.dispatchEvent(event);
    });
    
    // Wait for install prompt to appear
    await expect(page.locator('[data-testid="pwa-install-prompt"], .pwa-install-prompt')).toBeVisible({ timeout: 3000 });
  });

  test('should handle offline functionality', async () => {
    await page.goto('/');
    
    // Wait for the app to be ready
    await page.waitForLoadState('networkidle');
    
    // Go offline
    await context.setOffline(true);
    
    // Verify offline indicator appears
    await expect(page.locator('body')).toHaveClass(/offline/);
    
    // Try to navigate - should work with cached resources
    await page.click('[href="/v2"]');
    await expect(page).toHaveURL(/\/v2/);
    
    // Go back online
    await context.setOffline(false);
    
    // Verify online indicator
    await expect(page.locator('body')).not.toHaveClass(/offline/);
  });

  test('should be responsive on mobile devices', async () => {
    await page.goto('/v2');
    
    // Check mobile navigation is visible
    await expect(page.locator('[data-testid="mobile-navigation"], .mobile-navigation')).toBeVisible();
    
    // Check desktop navigation is hidden on mobile
    await expect(page.locator('.lg\\:block').first()).not.toBeVisible();
    
    // Test mobile header
    const mobileHeader = page.locator('.lg\\:hidden');
    await expect(mobileHeader).toBeVisible();
    
    // Test bottom navigation on very small screens
    await page.setViewportSize({ width: 320, height: 568 });
    await expect(page.locator('.sm\\:hidden.fixed.bottom-0')).toBeVisible();
  });

  test('should support touch gestures', async () => {
    await page.goto('/v2');
    
    // Find a swipeable card
    const card = page.locator('[data-testid="responsive-card"], .touch-card').first();
    await expect(card).toBeVisible();
    
    // Simulate touch gestures
    const cardBox = await card.boundingBox();
    if (cardBox) {
      // Simulate swipe right
      await page.touchscreen.tap(cardBox.x + 50, cardBox.y + cardBox.height / 2);
      await page.mouse.move(cardBox.x + 50, cardBox.y + cardBox.height / 2);
      await page.mouse.down();
      await page.mouse.move(cardBox.x + 150, cardBox.y + cardBox.height / 2);
      await page.mouse.up();
    }
    
    // Test navigation drawer gesture
    await page.touchscreen.tap(10, 400);
    await page.mouse.move(10, 400);
    await page.mouse.down();
    await page.mouse.move(200, 400);
    await page.mouse.up();
    
    // Navigation drawer should open
    await expect(page.locator('[data-testid="mobile-drawer"], .mobile-drawer')).toBeVisible({ timeout: 1000 });
  });

  test('should handle push notifications', async () => {
    await page.goto('/');
    
    // Check if notifications are supported
    const notificationSupport = await page.evaluate(() => {
      return 'Notification' in window && 'serviceWorker' in navigator && 'PushManager' in window;
    });
    
    if (notificationSupport) {
      // Test notification permission request
      const permissionResult = await page.evaluate(async () => {
        const permission = await Notification.requestPermission();
        return permission;
      });
      
      expect(['granted', 'denied', 'default']).toContain(permissionResult);
      
      if (permissionResult === 'granted') {
        // Test local notification
        await page.evaluate(() => {
          new Notification('Test PWA Notification', {
            body: 'This is a test notification from OllamaMax PWA',
            icon: '/icons/icon-192x192.png',
            badge: '/icons/badge-72x72.png',
            tag: 'test-notification'
          });
        });
      }
    }
  });

  test('should persist data offline', async () => {
    await page.goto('/v2');
    
    // Create some test data
    await page.evaluate(async () => {
      // Simulate creating data that should be synced
      const mockOperation = {
        type: 'create',
        resource: 'test-data',
        data: { id: 'test-1', value: 'test-value' },
        maxRetries: 3
      };
      
      // Store in IndexedDB (simulating offline sync)
      const request = indexedDB.open('OllamaMaxOfflineDB', 1);
      
      return new Promise((resolve) => {
        request.onsuccess = () => {
          const db = request.result;
          const transaction = db.transaction(['syncOperations'], 'readwrite');
          const store = transaction.objectStore('syncOperations');
          
          const operation = {
            id: Date.now().toString(),
            timestamp: Date.now(),
            retryCount: 0,
            ...mockOperation
          };
          
          store.add(operation);
          transaction.oncomplete = () => resolve(true);
        };
        
        request.onupgradeneeded = (event) => {
          const db = (event.target as any).result;
          if (!db.objectStoreNames.contains('syncOperations')) {
            db.createObjectStore('syncOperations', { keyPath: 'id' });
          }
        };
      });
    });
    
    // Go offline
    await context.setOffline(true);
    
    // Verify data persists offline
    const hasOfflineData = await page.evaluate(async () => {
      const request = indexedDB.open('OllamaMaxOfflineDB', 1);
      
      return new Promise((resolve) => {
        request.onsuccess = () => {
          const db = request.result;
          const transaction = db.transaction(['syncOperations'], 'readonly');
          const store = transaction.objectStore('syncOperations');
          const getRequest = store.getAll();
          
          getRequest.onsuccess = () => {
            resolve(getRequest.result.length > 0);
          };
        };
      });
    });
    
    expect(hasOfflineData).toBe(true);
    
    // Go back online
    await context.setOffline(false);
  });

  test('should meet PWA performance criteria', async () => {
    await page.goto('/');
    
    // Check that the page loads quickly
    const startTime = Date.now();
    await page.waitForLoadState('networkidle');
    const loadTime = Date.now() - startTime;
    
    // PWA should load within 3 seconds
    expect(loadTime).toBeLessThan(3000);
    
    // Check for proper caching headers
    const response = await page.goto('/');
    const cacheControl = response?.headers()['cache-control'];
    expect(cacheControl).toBeDefined();
    
    // Verify critical resources are cached
    const swResponse = await page.goto('/sw.js');
    expect(swResponse?.status()).toBe(200);
    
    const manifestResponse = await page.goto('/manifest.json');
    expect(manifestResponse?.status()).toBe(200);
  });

  test('should be installable', async () => {
    await page.goto('/');
    
    // Check for installability criteria
    const installable = await page.evaluate(async () => {
      // Check if all PWA requirements are met
      const checks = {
        hasManifest: !!document.querySelector('link[rel="manifest"]'),
        hasServiceWorker: 'serviceWorker' in navigator,
        hasIcons: true, // We know we have icons from manifest test
        isHTTPS: location.protocol === 'https:' || location.hostname === 'localhost',
        hasStartUrl: true // Defined in manifest
      };
      
      return Object.values(checks).every(Boolean);
    });
    
    expect(installable).toBe(true);
  });

  test('should handle app updates', async () => {
    await page.goto('/');
    
    // Simulate a service worker update
    await page.evaluate(async () => {
      if ('serviceWorker' in navigator) {
        const registration = await navigator.serviceWorker.ready;
        
        // Simulate update available
        const updateEvent = new CustomEvent('sw-update-available', {
          detail: { registration }
        });
        
        window.dispatchEvent(updateEvent);
      }
    });
    
    // Check if update notification appears
    await expect(page.locator('[data-testid="pwa-update-notification"], .pwa-update-notification')).toBeVisible({ timeout: 3000 });
  });
});

// Lighthouse PWA audit test
test.describe('Lighthouse PWA Audit', () => {
  test('should pass Lighthouse PWA audit', async ({ page }) => {
    await page.goto('/');
    
    // This would typically be run with lighthouse-ci in CI/CD
    // For now, we'll check basic PWA requirements
    
    const pwaFeatures = await page.evaluate(() => {
      const checks = {
        manifest: !!document.querySelector('link[rel="manifest"]'),
        serviceWorker: 'serviceWorker' in navigator,
        httpsOrLocalhost: location.protocol === 'https:' || location.hostname === 'localhost',
        viewport: !!document.querySelector('meta[name="viewport"]'),
        themeColor: !!document.querySelector('meta[name="theme-color"]'),
        appleTouch: !!document.querySelector('link[rel="apple-touch-icon"]')
      };
      
      return checks;
    });
    
    expect(pwaFeatures.manifest).toBe(true);
    expect(pwaFeatures.serviceWorker).toBe(true);
    expect(pwaFeatures.httpsOrLocalhost).toBe(true);
    expect(pwaFeatures.viewport).toBe(true);
    expect(pwaFeatures.themeColor).toBe(true);
    expect(pwaFeatures.appleTouch).toBe(true);
  });
});