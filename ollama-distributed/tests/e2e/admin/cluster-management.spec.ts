import { test, expect } from '@playwright/test';
import { BrowserTestFramework } from '../../utils/browser-automation';

test.describe('Cluster Management Administration', () => {
  let framework: BrowserTestFramework;

  test.beforeEach(async ({ page }) => {
    framework = new BrowserTestFramework(page);
    
    // Login as admin
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'admin@example.com');
    await page.fill('[data-testid="password-input"]', 'admin123');
    await page.click('[data-testid="login-button"]');
    await expect(page).toHaveURL('/dashboard');
    
    // Navigate to admin panel
    await page.goto('/admin/cluster');
    await expect(page.locator('[data-testid="admin-cluster-panel"]')).toBeVisible();
  });

  test('Node management workflow', async ({ page }) => {
    const result = await framework.testClusterManagement();
    expect(result.success).toBeTruthy();
    
    // Additional node management tests
    const nodeList = page.locator('[data-testid="cluster-nodes"]');
    await expect(nodeList).toBeVisible();
    
    // Test node status monitoring
    const nodeItems = nodeList.locator('[data-testid^="node-"]');
    const nodeCount = await nodeItems.count();
    
    for (let i = 0; i < nodeCount; i++) {
      const node = nodeItems.nth(i);
      
      // Verify node has required information
      await expect(node.locator('[data-testid="node-id"]')).toBeVisible();
      await expect(node.locator('[data-testid="node-status"]')).toBeVisible();
      await expect(node.locator('[data-testid="node-cpu"]')).toBeVisible();
      await expect(node.locator('[data-testid="node-memory"]')).toBeVisible();
      
      // Test node actions menu
      await node.locator('[data-testid="node-actions-menu"]').click();
      const actionsDropdown = node.locator('[data-testid="node-actions-dropdown"]');
      await expect(actionsDropdown).toBeVisible();
      
      // Verify available actions
      await expect(actionsDropdown.locator('[data-action="restart"]')).toBeVisible();
      await expect(actionsDropdown.locator('[data-action="maintenance"]')).toBeVisible();
      await expect(actionsDropdown.locator('[data-action="remove"]')).toBeVisible();
      
      // Close dropdown
      await page.keyboard.press('Escape');
    }
  });

  test('Add new node to cluster', async ({ page }) => {
    // Open add node dialog
    await page.click('[data-testid="add-node-button"]');
    const addNodeDialog = page.locator('[data-testid="add-node-dialog"]');
    await expect(addNodeDialog).toBeVisible();
    
    // Test form validation
    await page.click('[data-testid="confirm-add-node"]');
    await expect(page.locator('[data-testid="node-address-error"]')).toBeVisible();
    
    // Fill in valid node information
    await page.fill('[data-testid="node-address"]', '192.168.1.101:8080');
    await page.fill('[data-testid="node-name"]', 'Test Node 1');
    await page.selectOption('[data-testid="node-type"]', 'worker');
    
    // Add authentication if required
    await page.check('[data-testid="requires-auth"]');
    await page.fill('[data-testid="auth-token"]', 'test-auth-token-123');
    
    // Submit form
    await page.click('[data-testid="confirm-add-node"]');
    
    // Verify node addition
    await expect(page.locator('[data-testid="node-addition-success"]')).toBeVisible();
    await expect(page.locator('[data-testid="node-192.168.1.101"]')).toBeVisible({ timeout: 15000 });
    
    // Verify node appears in cluster list with pending status
    const newNode = page.locator('[data-testid="node-192.168.1.101"]');
    await expect(newNode.locator('[data-testid="node-status"]')).toHaveText('connecting');
    
    // Wait for node to become ready
    await expect(newNode.locator('[data-testid="node-status"]')).toHaveText('ready', { timeout: 30000 });
  });

  test('Node maintenance and restart operations', async ({ page }) => {
    // Select a healthy node
    const healthyNode = page.locator('[data-testid^="node-"][data-status="ready"]').first();
    await expect(healthyNode).toBeVisible();
    
    const nodeId = await healthyNode.getAttribute('data-testid');
    
    // Put node in maintenance mode
    await healthyNode.locator('[data-testid="node-actions-menu"]').click();
    await page.click('[data-action="maintenance"]');
    
    // Confirm maintenance mode
    const confirmDialog = page.locator('[data-testid="confirm-maintenance-dialog"]');
    await expect(confirmDialog).toBeVisible();
    await page.click('[data-testid="confirm-maintenance"]');
    
    // Verify node status changes to maintenance
    await expect(healthyNode.locator('[data-testid="node-status"]')).toHaveText('maintenance', { timeout: 10000 });
    
    // Verify node stops receiving new tasks
    const activeTasksCount = healthyNode.locator('[data-testid="active-tasks"]');
    const initialTasks = await activeTasksCount.textContent();
    
    await page.waitForTimeout(5000);
    const currentTasks = await activeTasksCount.textContent();
    
    // Tasks should not increase (may decrease as they complete)
    expect(parseInt(currentTasks || '0')).toBeLessThanOrEqual(parseInt(initialTasks || '0'));
    
    // Restart node
    await healthyNode.locator('[data-testid="node-actions-menu"]').click();
    await page.click('[data-action="restart"]');
    
    const restartDialog = page.locator('[data-testid="confirm-restart-dialog"]');
    await expect(restartDialog).toBeVisible();
    await page.click('[data-testid="confirm-restart"]');
    
    // Verify node status progression: restarting -> ready
    await expect(healthyNode.locator('[data-testid="node-status"]')).toHaveText('restarting', { timeout: 5000 });
    await expect(healthyNode.locator('[data-testid="node-status"]')).toHaveText('ready', { timeout: 60000 });
  });

  test('Node removal and cleanup', async ({ page }) => {
    // Add a test node first
    await page.click('[data-testid="add-node-button"]');
    await page.fill('[data-testid="node-address"]', '192.168.1.102:8080');
    await page.fill('[data-testid="node-name"]', 'Test Node for Removal');
    await page.click('[data-testid="confirm-add-node"]');
    
    // Wait for node to be added
    const testNode = page.locator('[data-testid="node-192.168.1.102"]');
    await expect(testNode).toBeVisible({ timeout: 15000 });
    
    // Remove the node
    await testNode.locator('[data-testid="node-actions-menu"]').click();
    await page.click('[data-action="remove"]');
    
    // Confirm removal with safety checks
    const removeDialog = page.locator('[data-testid="confirm-remove-dialog"]');
    await expect(removeDialog).toBeVisible();
    
    // Verify warning message
    await expect(removeDialog.locator('[data-testid="removal-warning"]')).toContainText('permanently remove');
    
    // Type confirmation
    await page.fill('[data-testid="remove-confirmation-input"]', 'REMOVE');
    await page.click('[data-testid="confirm-remove"]');
    
    // Verify node removal
    await expect(page.locator('[data-testid="node-removal-success"]')).toBeVisible();
    await expect(testNode).not.toBeVisible({ timeout: 10000 });
  });

  test('Cluster health monitoring', async ({ page }) => {
    // Check cluster overview panel
    const clusterOverview = page.locator('[data-testid="cluster-overview"]');
    await expect(clusterOverview).toBeVisible();
    
    // Verify health metrics
    const healthMetrics = [
      'total-nodes',
      'healthy-nodes',
      'total-memory',
      'available-memory',
      'total-cpu',
      'cpu-utilization',
      'active-models',
      'request-rate'
    ];
    
    for (const metric of healthMetrics) {
      const metricElement = clusterOverview.locator(`[data-testid="${metric}"]`);
      await expect(metricElement).toBeVisible();
      
      // Verify metric has numerical value
      const value = await metricElement.textContent();
      expect(value).toMatch(/\d+/);
    }
    
    // Test cluster health status
    const healthStatus = page.locator('[data-testid="cluster-health-status"]');
    await expect(healthStatus).toBeVisible();
    
    const status = await healthStatus.getAttribute('data-status');
    expect(['healthy', 'warning', 'critical']).toContain(status);
  });

  test('Model distribution management', async ({ page }) => {
    await page.goto('/admin/models');
    
    // Test model deployment across cluster
    const modelList = page.locator('[data-testid="available-models"]');
    await expect(modelList).toBeVisible();
    
    // Select a model to deploy
    const firstModel = modelList.locator('[data-testid^="model-"]').first();
    await expect(firstModel).toBeVisible();
    
    await firstModel.locator('[data-testid="deploy-model-button"]').click();
    
    // Configure deployment
    const deployDialog = page.locator('[data-testid="deploy-model-dialog"]');
    await expect(deployDialog).toBeVisible();
    
    // Select target nodes
    await page.check('[data-testid="select-all-nodes"]');
    await page.selectOption('[data-testid="deployment-strategy"]', 'rolling');
    await page.fill('[data-testid="max-concurrent-deployments"]', '2');
    
    // Start deployment
    await page.click('[data-testid="start-deployment"]');
    
    // Monitor deployment progress
    const deploymentProgress = page.locator('[data-testid="deployment-progress"]');
    await expect(deploymentProgress).toBeVisible();
    
    // Wait for deployment completion
    await expect(page.locator('[data-testid="deployment-complete"]')).toBeVisible({ timeout: 120000 });
    
    // Verify model is deployed on all selected nodes
    await page.goto('/admin/cluster');
    const nodes = page.locator('[data-testid^="node-"]');
    const nodeCount = await nodes.count();
    
    for (let i = 0; i < nodeCount; i++) {
      const node = nodes.nth(i);
      const modelCount = node.locator('[data-testid="deployed-models-count"]');
      const count = await modelCount.textContent();
      expect(parseInt(count || '0')).toBeGreaterThan(0);
    }
  });

  test('Load balancing configuration', async ({ page }) => {
    await page.goto('/admin/load-balancing');
    
    // Test load balancing strategy configuration
    const strategySelect = page.locator('[data-testid="load-balancing-strategy"]');
    await expect(strategySelect).toBeVisible();
    
    // Test different strategies
    const strategies = ['round-robin', 'least-connections', 'weighted-round-robin', 'ip-hash'];
    
    for (const strategy of strategies) {
      await strategySelect.selectOption(strategy);
      
      // Apply configuration
      await page.click('[data-testid="apply-lb-config"]');
      await expect(page.locator('[data-testid="config-applied-success"]')).toBeVisible();
      
      // Verify strategy is active
      await expect(page.locator('[data-testid="active-strategy"]')).toContainText(strategy);
    }
    
    // Test weight configuration for weighted round-robin
    await strategySelect.selectOption('weighted-round-robin');
    
    const nodeWeights = page.locator('[data-testid="node-weights"]');
    await expect(nodeWeights).toBeVisible();
    
    // Configure weights for each node
    const weightInputs = nodeWeights.locator('[data-testid^="weight-node-"]');
    const weightCount = await weightInputs.count();
    
    for (let i = 0; i < weightCount; i++) {
      const weight = Math.floor(Math.random() * 10) + 1; // Random weight 1-10
      await weightInputs.nth(i).fill(weight.toString());
    }
    
    await page.click('[data-testid="apply-weights"]');
    await expect(page.locator('[data-testid="weights-applied-success"]')).toBeVisible();
  });

  test('Cluster scaling operations', async ({ page }) => {
    await page.goto('/admin/scaling');
    
    // Test auto-scaling configuration
    const autoScalingToggle = page.locator('[data-testid="auto-scaling-toggle"]');
    await expect(autoScalingToggle).toBeVisible();
    
    await autoScalingToggle.check();
    
    // Configure scaling parameters
    await page.fill('[data-testid="min-nodes"]', '2');
    await page.fill('[data-testid="max-nodes"]', '10');
    await page.fill('[data-testid="target-cpu-utilization"]', '70');
    await page.fill('[data-testid="scale-up-threshold"]', '80');
    await page.fill('[data-testid="scale-down-threshold"]', '30');
    
    // Set scaling cooldown periods
    await page.fill('[data-testid="scale-up-cooldown"]', '300'); // 5 minutes
    await page.fill('[data-testid="scale-down-cooldown"]', '600'); // 10 minutes
    
    // Apply auto-scaling configuration
    await page.click('[data-testid="apply-scaling-config"]');
    await expect(page.locator('[data-testid="scaling-config-applied"]')).toBeVisible();
    
    // Test manual scaling
    const currentNodeCount = await page.locator('[data-testid="current-node-count"]').textContent();
    const newNodeCount = parseInt(currentNodeCount || '0') + 1;
    
    await page.fill('[data-testid="target-node-count"]', newNodeCount.toString());
    await page.click('[data-testid="manual-scale"]');
    
    // Monitor scaling operation
    const scalingProgress = page.locator('[data-testid="scaling-progress"]');
    await expect(scalingProgress).toBeVisible();
    
    await expect(page.locator('[data-testid="scaling-complete"]')).toBeVisible({ timeout: 180000 });
    
    // Verify new node count
    const updatedNodeCount = await page.locator('[data-testid="current-node-count"]').textContent();
    expect(parseInt(updatedNodeCount || '0')).toBe(newNodeCount);
  });

  test('Cluster backup and recovery', async ({ page }) => {
    await page.goto('/admin/backup');
    
    // Test cluster state backup
    await page.click('[data-testid="create-backup-button"]');
    
    const backupDialog = page.locator('[data-testid="create-backup-dialog"]');
    await expect(backupDialog).toBeVisible();
    
    // Configure backup options
    await page.fill('[data-testid="backup-name"]', `cluster-backup-${Date.now()}`);
    await page.check('[data-testid="include-models"]');
    await page.check('[data-testid="include-configs"]');
    await page.check('[data-testid="include-logs"]');
    
    // Start backup
    await page.click('[data-testid="start-backup"]');
    
    // Monitor backup progress
    const backupProgress = page.locator('[data-testid="backup-progress"]');
    await expect(backupProgress).toBeVisible();
    
    await expect(page.locator('[data-testid="backup-complete"]')).toBeVisible({ timeout: 300000 });
    
    // Verify backup appears in list
    const backupList = page.locator('[data-testid="backup-list"]');
    const latestBackup = backupList.locator('[data-testid^="backup-"]').first();
    await expect(latestBackup).toBeVisible();
    
    // Test backup verification
    await latestBackup.locator('[data-testid="verify-backup"]').click();
    await expect(page.locator('[data-testid="backup-verification-success"]')).toBeVisible({ timeout: 60000 });
    
    // Test backup restore (dry run)
    await latestBackup.locator('[data-testid="restore-backup"]').click();
    
    const restoreDialog = page.locator('[data-testid="restore-backup-dialog"]');
    await expect(restoreDialog).toBeVisible();
    
    await page.check('[data-testid="dry-run-restore"]');
    await page.click('[data-testid="start-restore"]');
    
    await expect(page.locator('[data-testid="dry-run-complete"]')).toBeVisible({ timeout: 120000 });
  });
});