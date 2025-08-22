import { test, expect } from '@playwright/test';
import { BrowserTestFramework } from '../../utils/browser-automation';

test.describe('Real-time Dashboard Updates', () => {
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

  test('WebSocket connection establishment and metrics updates', async ({ page }) => {
    const result = await framework.testRealTimeUpdates();
    expect(result.success).toBeTruthy();
    expect(result.duration).toBeLessThan(30000); // 30 seconds max
    
    if (result.errors.length > 0) {
      console.error('Real-time update errors:', result.errors);
    }
  });

  test('Live cluster status monitoring', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Wait for initial cluster status load
    const clusterStatus = page.locator('[data-testid="cluster-status"]');
    await expect(clusterStatus).toBeVisible();
    
    // Monitor for status changes
    const initialStatus = await clusterStatus.getAttribute('data-status');
    
    // Simulate node status change
    await page.evaluate(() => {
      window.dispatchEvent(new CustomEvent('cluster-update', {
        detail: {
          type: 'node_status_change',
          nodeId: 'node-1',
          status: 'degraded'
        }
      }));
    });
    
    // Verify status update within 2 seconds
    await expect(clusterStatus.locator('[data-node="node-1"]')).toHaveAttribute('data-status', 'degraded', { timeout: 2000 });
    
    // Test status color indicators
    const degradedNode = clusterStatus.locator('[data-node="node-1"][data-status="degraded"]');
    await expect(degradedNode).toHaveClass(/status-degraded/);
  });

  test('Real-time performance metrics visualization', async ({ page }) => {
    await page.goto('/dashboard/performance');
    
    // Wait for metrics charts to load
    const metricsChart = page.locator('[data-testid="performance-chart"]');
    await expect(metricsChart).toBeVisible();
    
    // Verify chart updates with live data
    const chartCanvas = metricsChart.locator('canvas');
    await expect(chartCanvas).toBeVisible();
    
    // Wait for data updates (charts should refresh every 5 seconds)
    await page.waitForTimeout(6000);
    
    // Verify chart has rendered data points
    const hasDataPoints = await page.evaluate(() => {
      const canvas = document.querySelector('[data-testid="performance-chart"] canvas') as HTMLCanvasElement;
      if (!canvas) return false;
      
      const ctx = canvas.getContext('2d');
      const imageData = ctx?.getImageData(0, 0, canvas.width, canvas.height);
      
      // Check if canvas has non-transparent pixels (indicating rendered content)
      if (!imageData) return false;
      
      for (let i = 3; i < imageData.data.length; i += 4) {
        if (imageData.data[i] > 0) return true; // Found non-transparent pixel
      }
      return false;
    });
    
    expect(hasDataPoints).toBeTruthy();
  });

  test('Live notification system', async ({ page }) => {
    await page.goto('/dashboard');
    
    const notificationContainer = page.locator('[data-testid="notifications"]');
    await expect(notificationContainer).toBeVisible();
    
    // Test different notification types
    const notificationTypes = [
      { type: 'success', message: 'Model deployment completed' },
      { type: 'warning', message: 'Node memory usage high' },
      { type: 'error', message: 'Connection to node lost' },
      { type: 'info', message: 'System maintenance scheduled' }
    ];
    
    for (const notification of notificationTypes) {
      // Trigger notification
      await page.evaluate((notif) => {
        window.dispatchEvent(new CustomEvent('system-notification', {
          detail: notif
        }));
      }, notification);
      
      // Verify notification appears
      const notificationElement = notificationContainer.locator(
        `[data-notification-type="${notification.type}"]`
      ).last();
      
      await expect(notificationElement).toBeVisible({ timeout: 2000 });
      await expect(notificationElement).toContainText(notification.message);
      
      // Test auto-dismiss for non-error notifications
      if (notification.type !== 'error') {
        await expect(notificationElement).not.toBeVisible({ timeout: 8000 });
      }
    }
  });

  test('WebSocket connection recovery', async ({ page }) => {
    const result = await framework.testWebSocketConnection();
    expect(result.success).toBeTruthy();
    
    // Additional connection recovery tests
    await page.goto('/dashboard');
    
    // Verify connection status indicator
    const connectionStatus = page.locator('[data-testid="connection-status"]');
    await expect(connectionStatus).toHaveAttribute('data-status', 'connected');
    
    // Simulate network interruption
    await page.context().setOffline(true);
    
    // Connection status should show disconnected
    await expect(connectionStatus).toHaveAttribute('data-status', 'disconnected', { timeout: 5000 });
    
    // Restore connection
    await page.context().setOffline(false);
    
    // Verify reconnection
    await expect(connectionStatus).toHaveAttribute('data-status', 'connected', { timeout: 10000 });
    
    // Verify data sync after reconnection
    const lastUpdateTime = page.locator('[data-testid="last-update-time"]');
    const timeBeforeDisconnect = await lastUpdateTime.textContent();
    
    // Wait for new data after reconnection
    await page.waitForTimeout(3000);
    const timeAfterReconnect = await lastUpdateTime.textContent();
    
    expect(timeAfterReconnect).not.toBe(timeBeforeDisconnect);
  });

  test('Multi-user real-time collaboration', async ({ page, context }) => {
    // First user session
    await page.goto('/dashboard/shared-workspace');
    
    // Open second user session
    const secondPage = await context.newPage();
    await secondPage.goto('/login');
    await secondPage.fill('[data-testid="email-input"]', 'user2@example.com');
    await secondPage.fill('[data-testid="password-input"]', 'user123');
    await secondPage.click('[data-testid="login-button"]');
    await secondPage.goto('/dashboard/shared-workspace');
    
    // Test collaborative editing
    const sharedDocument = page.locator('[data-testid="shared-document"]');
    const secondUserDocument = secondPage.locator('[data-testid="shared-document"]');
    
    // First user makes changes
    await sharedDocument.fill('User 1 changes');
    
    // Second user should see changes in real-time
    await expect(secondUserDocument).toHaveValue('User 1 changes', { timeout: 3000 });
    
    // Test presence indicators
    const presenceIndicators = page.locator('[data-testid="user-presence"]');
    await expect(presenceIndicators).toContainText('user2@example.com');
    
    // Test cursor position sharing
    await sharedDocument.click({ position: { x: 50, y: 10 } });
    
    const otherUserCursor = secondPage.locator('[data-testid="cursor-user1"]');
    await expect(otherUserCursor).toBeVisible({ timeout: 2000 });
  });

  test('Real-time data filtering and search', async ({ page }) => {
    await page.goto('/dashboard/logs');
    
    // Wait for initial log data
    const logContainer = page.locator('[data-testid="log-container"]');
    await expect(logContainer.locator('.log-entry').first()).toBeVisible();
    
    // Test real-time filtering
    const filterInput = page.locator('[data-testid="log-filter"]');
    await filterInput.fill('error');
    
    // Verify filtering
    const visibleLogs = logContainer.locator('.log-entry:visible');
    const logCount = await visibleLogs.count();
    
    for (let i = 0; i < logCount; i++) {
      const logText = await visibleLogs.nth(i).textContent();
      expect(logText?.toLowerCase()).toContain('error');
    }
    
    // Test live filtering with new incoming logs
    await page.evaluate(() => {
      window.dispatchEvent(new CustomEvent('new-log-entry', {
        detail: {
          level: 'info',
          message: 'System information message',
          timestamp: new Date().toISOString()
        }
      }));
    });
    
    // Info log should not appear due to error filter
    await page.waitForTimeout(1000);
    const newLogVisible = await logContainer.locator(':text("System information message")').isVisible();
    expect(newLogVisible).toBeFalsy();
    
    // Clear filter
    await filterInput.clear();
    
    // Now info log should appear
    await expect(logContainer.locator(':text("System information message")')).toBeVisible({ timeout: 2000 });
  });

  test('Dashboard widget auto-refresh', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Test different widget refresh intervals
    const widgets = [
      { selector: '[data-testid="cpu-usage-widget"]', interval: 5000 },
      { selector: '[data-testid="memory-usage-widget"]', interval: 5000 },
      { selector: '[data-testid="network-traffic-widget"]', interval: 10000 },
      { selector: '[data-testid="disk-usage-widget"]', interval: 30000 }
    ];
    
    for (const widget of widgets) {
      const widgetElement = page.locator(widget.selector);
      await expect(widgetElement).toBeVisible();
      
      // Get initial value
      const initialValue = await widgetElement.locator('[data-testid="metric-value"]').textContent();
      
      // Wait for refresh interval
      await page.waitForTimeout(widget.interval + 1000);
      
      // Value should have updated (or at least timestamp should change)
      const lastUpdated = widgetElement.locator('[data-testid="last-updated"]');
      const updateTime = await lastUpdated.textContent();
      
      // Verify recent update (within last 10 seconds)
      const now = new Date();
      const updateTimestamp = new Date(updateTime || '');
      const timeDiff = now.getTime() - updateTimestamp.getTime();
      
      expect(timeDiff).toBeLessThan(10000); // Within 10 seconds
    }
  });

  test('Real-time alert management', async ({ page }) => {
    await page.goto('/dashboard/alerts');
    
    // Test alert creation and real-time display
    const alertContainer = page.locator('[data-testid="active-alerts"]');
    await expect(alertContainer).toBeVisible();
    
    // Simulate new alert
    await page.evaluate(() => {
      window.dispatchEvent(new CustomEvent('new-alert', {
        detail: {
          id: 'alert-001',
          severity: 'critical',
          title: 'High CPU Usage',
          description: 'Node CPU usage exceeded 90%',
          timestamp: new Date().toISOString(),
          source: 'node-1'
        }
      }));
    });
    
    // Verify alert appears immediately
    const newAlert = alertContainer.locator('[data-alert-id="alert-001"]');
    await expect(newAlert).toBeVisible({ timeout: 1000 });
    await expect(newAlert).toHaveClass(/severity-critical/);
    
    // Test alert acknowledgment
    await newAlert.locator('[data-testid="acknowledge-button"]').click();
    await expect(newAlert).toHaveClass(/acknowledged/);
    
    // Test alert resolution
    await page.evaluate(() => {
      window.dispatchEvent(new CustomEvent('alert-resolved', {
        detail: { id: 'alert-001' }
      }));
    });
    
    await expect(newAlert).not.toBeVisible({ timeout: 2000 });
  });
});