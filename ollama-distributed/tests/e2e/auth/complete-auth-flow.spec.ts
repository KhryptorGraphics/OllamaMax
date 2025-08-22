import { test, expect } from '@playwright/test';
import { BrowserTestFramework } from '../../utils/browser-automation';

test.describe('Complete Authentication Flow', () => {
  let framework: BrowserTestFramework;

  test.beforeEach(async ({ page }) => {
    framework = new BrowserTestFramework(page);
  });

  test('User registration to dashboard workflow', async ({ page }) => {
    // Test complete user journey from registration to dashboard access
    
    // 1. Registration
    await page.goto('/register');
    await page.fill('[data-testid="username-input"]', 'testuser@example.com');
    await page.fill('[data-testid="password-input"]', 'SecurePassword123!');
    await page.fill('[data-testid="confirm-password-input"]', 'SecurePassword123!');
    await page.click('[data-testid="register-button"]');

    // Verify registration success
    await expect(page.locator('[data-testid="registration-success"]')).toBeVisible();
    
    // 2. Email verification simulation
    await page.goto('/verify-email?token=test-verification-token');
    await expect(page.locator('[data-testid="email-verified"]')).toBeVisible();
    
    // 3. Login with new credentials
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'testuser@example.com');
    await page.fill('[data-testid="password-input"]', 'SecurePassword123!');
    await page.click('[data-testid="login-button"]');

    // 4. Dashboard access
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('[data-testid="user-dashboard"]')).toBeVisible();
    
    // 5. Session persistence test
    await page.reload();
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('[data-testid="user-welcome"]')).toContainText('testuser@example.com');
  });

  test('Login form validation', async ({ page }) => {
    await page.goto('/login');
    
    // Test empty form submission
    await page.click('[data-testid="login-button"]');
    await expect(page.locator('[data-testid="email-error"]')).toBeVisible();
    await expect(page.locator('[data-testid="password-error"]')).toBeVisible();
    
    // Test invalid email format
    await page.fill('[data-testid="email-input"]', 'invalid-email');
    await page.click('[data-testid="login-button"]');
    await expect(page.locator('[data-testid="email-error"]')).toContainText('Invalid email format');
    
    // Test password requirements
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="password-input"]', '123');
    await page.click('[data-testid="login-button"]');
    await expect(page.locator('[data-testid="password-error"]')).toContainText('Password too short');
  });

  test('Password reset flow', async ({ page }) => {
    // 1. Initiate password reset
    await page.goto('/login');
    await page.click('[data-testid="forgot-password-link"]');
    
    await page.fill('[data-testid="reset-email-input"]', 'test@example.com');
    await page.click('[data-testid="send-reset-button"]');
    
    await expect(page.locator('[data-testid="reset-sent-message"]')).toBeVisible();
    
    // 2. Reset password with token
    await page.goto('/reset-password?token=test-reset-token');
    await page.fill('[data-testid="new-password-input"]', 'NewSecurePassword123!');
    await page.fill('[data-testid="confirm-new-password-input"]', 'NewSecurePassword123!');
    await page.click('[data-testid="reset-password-button"]');
    
    await expect(page.locator('[data-testid="password-reset-success"]')).toBeVisible();
    
    // 3. Login with new password
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="password-input"]', 'NewSecurePassword123!');
    await page.click('[data-testid="login-button"]');
    
    await expect(page).toHaveURL('/dashboard');
  });

  test('Session timeout and renewal', async ({ page }) => {
    // Login
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    
    await expect(page).toHaveURL('/dashboard');
    
    // Simulate session timeout
    await page.evaluate(() => {
      localStorage.removeItem('auth_token');
      sessionStorage.clear();
    });
    
    // Navigate to protected route
    await page.goto('/admin');
    
    // Should redirect to login
    await expect(page).toHaveURL('/login');
    await expect(page.locator('[data-testid="session-expired-message"]')).toBeVisible();
  });

  test('Two-factor authentication flow', async ({ page }) => {
    // Login with 2FA enabled account
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', '2fa-user@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    
    // Should show 2FA challenge
    await expect(page.locator('[data-testid="2fa-challenge"]')).toBeVisible();
    
    // Enter 2FA code
    await page.fill('[data-testid="2fa-code-input"]', '123456');
    await page.click('[data-testid="verify-2fa-button"]');
    
    // Should reach dashboard
    await expect(page).toHaveURL('/dashboard');
    
    // Test backup codes
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', '2fa-user@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    
    await page.click('[data-testid="use-backup-code-link"]');
    await page.fill('[data-testid="backup-code-input"]', 'backup-123456789');
    await page.click('[data-testid="verify-backup-code-button"]');
    
    await expect(page).toHaveURL('/dashboard');
  });

  test('OAuth social login integration', async ({ page }) => {
    await page.goto('/login');
    
    // Test Google OAuth flow
    const googleLoginPromise = page.waitForEvent('popup');
    await page.click('[data-testid="google-login-button"]');
    const googlePopup = await googleLoginPromise;
    
    // Simulate OAuth success
    await googlePopup.goto('/oauth/callback?provider=google&token=success');
    await googlePopup.close();
    
    // Should redirect to dashboard
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('[data-testid="user-avatar"]')).toBeVisible();
  });

  test('Account lockout protection', async ({ page }) => {
    await page.goto('/login');
    
    // Attempt multiple failed logins
    for (let i = 0; i < 5; i++) {
      await page.fill('[data-testid="email-input"]', 'test@example.com');
      await page.fill('[data-testid="password-input"]', 'wrongpassword');
      await page.click('[data-testid="login-button"]');
      
      if (i < 4) {
        await expect(page.locator('[data-testid="login-error"]')).toBeVisible();
      }
    }
    
    // Account should be locked
    await expect(page.locator('[data-testid="account-locked-message"]')).toBeVisible();
    
    // Even correct password should be rejected
    await page.fill('[data-testid="password-input"]', 'correctpassword');
    await page.click('[data-testid="login-button"]');
    await expect(page.locator('[data-testid="account-locked-message"]')).toBeVisible();
  });

  test('Concurrent session management', async ({ page, context }) => {
    // Login in first session
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    await page.click('[data-testid="login-button"]');
    
    await expect(page).toHaveURL('/dashboard');
    
    // Open second browser session
    const secondPage = await context.newPage();
    await secondPage.goto('/login');
    await secondPage.fill('[data-testid="email-input"]', 'test@example.com');
    await secondPage.fill('[data-testid="password-input"]', 'password123');
    await secondPage.click('[data-testid="login-button"]');
    
    // Should show concurrent session warning
    await expect(secondPage.locator('[data-testid="concurrent-session-warning"]')).toBeVisible();
    
    // First session should be invalidated
    await page.reload();
    await expect(page).toHaveURL('/login');
    await expect(page.locator('[data-testid="session-terminated-message"]')).toBeVisible();
  });
});