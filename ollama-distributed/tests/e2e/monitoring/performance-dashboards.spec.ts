import { test, expect } from '@playwright/test';
import { BrowserTestFramework } from '../../utils/browser-automation';

test.describe('Performance Dashboards and Monitoring', () => {
  let framework: BrowserTestFramework;

  test.beforeEach(async ({ page }) => {
    framework = new BrowserTestFramework(page);
    
    // Login and navigate to monitoring dashboard
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'admin@example.com');
    await page.fill('[data-testid="password-input"]', 'admin123');
    await page.click('[data-testid="login-button"]');
    await page.goto('/monitoring/performance');
    await expect(page.locator('[data-testid="performance-dashboard"]')).toBeVisible();
  });

  test('Real-time performance metrics visualization', async ({ page }) => {
    const result = await framework.testPerformanceMetrics();
    expect(result.success).toBeTruthy();
    expect(result.metrics).toBeDefined();
    
    // Test specific performance widgets
    const performanceWidgets = [
      'cpu-utilization-chart',
      'memory-usage-chart',
      'network-traffic-chart',
      'disk-io-chart',
      'request-latency-chart',
      'throughput-chart'
    ];
    
    for (const widget of performanceWidgets) {
      const chartWidget = page.locator(`[data-testid="${widget}"]`);
      await expect(chartWidget).toBeVisible();
      
      // Verify chart has data
      const canvas = chartWidget.locator('canvas');
      await expect(canvas).toBeVisible();
      
      // Check if chart legend is present
      const legend = chartWidget.locator('[data-testid="chart-legend"]');
      await expect(legend).toBeVisible();
      
      // Verify time range selector
      const timeRange = chartWidget.locator('[data-testid="time-range-selector"]');
      await expect(timeRange).toBeVisible();
    }
  });

  test('Interactive time range selection', async ({ page }) => {
    const timeRanges = [
      { label: '1 Hour', value: '1h' },
      { label: '6 Hours', value: '6h' },
      { label: '24 Hours', value: '24h' },
      { label: '7 Days', value: '7d' },
      { label: '30 Days', value: '30d' }
    ];
    
    for (const timeRange of timeRanges) {
      // Select time range
      await page.selectOption('[data-testid="global-time-range"]', timeRange.value);
      
      // Wait for charts to update
      await page.waitForTimeout(2000);
      
      // Verify all charts updated with new time range
      const charts = page.locator('[data-testid$="-chart"]');
      const chartCount = await charts.count();
      
      for (let i = 0; i < chartCount; i++) {
        const chart = charts.nth(i);
        const timeRangeDisplay = chart.locator('[data-testid="current-time-range"]');
        await expect(timeRangeDisplay).toContainText(timeRange.label);
      }
    }
  });

  test('Performance alerts and thresholds', async ({ page }) => {
    await page.goto('/monitoring/alerts');
    
    // Create a new performance alert
    await page.click('[data-testid="create-alert-button"]');
    
    const alertDialog = page.locator('[data-testid="create-alert-dialog"]');
    await expect(alertDialog).toBeVisible();
    
    // Configure alert parameters
    await page.fill('[data-testid="alert-name"]', 'High CPU Usage Alert');
    await page.selectOption('[data-testid="metric-type"]', 'cpu_utilization');
    await page.selectOption('[data-testid="condition"]', 'greater_than');
    await page.fill('[data-testid="threshold-value"]', '80');
    await page.fill('[data-testid="duration"]', '5'); // 5 minutes
    
    // Set notification channels
    await page.check('[data-testid="email-notification"]');
    await page.check('[data-testid="slack-notification"]');
    
    // Create alert
    await page.click('[data-testid="create-alert"]');
    await expect(page.locator('[data-testid="alert-created-success"]')).toBeVisible();
    
    // Verify alert appears in list
    const alertsList = page.locator('[data-testid="alerts-list"]');
    const newAlert = alertsList.locator('[data-alert-name="High CPU Usage Alert"]');
    await expect(newAlert).toBeVisible();
    
    // Test alert editing
    await newAlert.locator('[data-testid="edit-alert"]').click();
    
    const editDialog = page.locator('[data-testid="edit-alert-dialog"]');
    await expect(editDialog).toBeVisible();
    
    // Modify threshold
    await page.fill('[data-testid="threshold-value"]', '85');
    await page.click('[data-testid="save-alert"]');
    
    await expect(page.locator('[data-testid="alert-updated-success"]')).toBeVisible();
  });

  test('Custom dashboard creation', async ({ page }) => {
    await page.goto('/monitoring/dashboards');
    
    // Create new custom dashboard
    await page.click('[data-testid="create-dashboard-button"]');
    
    const dashboardBuilder = page.locator('[data-testid="dashboard-builder"]');
    await expect(dashboardBuilder).toBeVisible();
    
    // Set dashboard properties
    await page.fill('[data-testid="dashboard-name"]', 'Custom Performance Dashboard');
    await page.fill('[data-testid="dashboard-description"]', 'Custom metrics for specific monitoring needs');
    
    // Add widgets to dashboard
    const widgetLibrary = page.locator('[data-testid="widget-library"]');
    await expect(widgetLibrary).toBeVisible();
    
    // Drag and drop CPU widget
    const cpuWidget = widgetLibrary.locator('[data-widget="cpu-utilization"]');
    const dashboardCanvas = page.locator('[data-testid="dashboard-canvas"]');
    
    await cpuWidget.dragTo(dashboardCanvas);
    
    // Verify widget appears on canvas
    const addedWidget = dashboardCanvas.locator('[data-widget-type="cpu-utilization"]');
    await expect(addedWidget).toBeVisible();
    
    // Configure widget
    await addedWidget.locator('[data-testid="configure-widget"]').click();
    
    const widgetConfig = page.locator('[data-testid="widget-configuration"]');
    await expect(widgetConfig).toBeVisible();
    
    await page.fill('[data-testid="widget-title"]', 'Cluster CPU Usage');
    await page.selectOption('[data-testid="aggregation-method"]', 'average');
    await page.click('[data-testid="save-widget-config"]');
    
    // Add more widgets
    const memoryWidget = widgetLibrary.locator('[data-widget="memory-usage"]');
    await memoryWidget.dragTo(dashboardCanvas);
    
    const networkWidget = widgetLibrary.locator('[data-widget="network-traffic"]');
    await networkWidget.dragTo(dashboardCanvas);
    
    // Save dashboard
    await page.click('[data-testid="save-dashboard"]');
    await expect(page.locator('[data-testid="dashboard-saved-success"]')).toBeVisible();
    
    // Verify dashboard in list
    await page.goto('/monitoring/dashboards');
    const dashboardList = page.locator('[data-testid="dashboards-list"]');
    const customDashboard = dashboardList.locator('[data-dashboard-name="Custom Performance Dashboard"]');
    await expect(customDashboard).toBeVisible();
  });

  test('Performance data export', async ({ page }) => {
    // Navigate to data export section
    await page.goto('/monitoring/export');
    
    // Configure export parameters
    await page.selectOption('[data-testid="export-format"]', 'csv');
    await page.selectOption('[data-testid="time-range"]', '24h');
    
    // Select metrics to export
    const metrics = [
      'cpu-utilization',
      'memory-usage',
      'network-traffic',
      'request-latency',
      'error-rate'
    ];
    
    for (const metric of metrics) {
      await page.check(`[data-testid="export-${metric}"]`);
    }
    
    // Set export schedule
    await page.check('[data-testid="schedule-export"]');
    await page.selectOption('[data-testid="export-frequency"]', 'daily');
    await page.fill('[data-testid="export-time"]', '02:00');
    
    // Configure export destination
    await page.selectOption('[data-testid="export-destination"]', 'email');
    await page.fill('[data-testid="export-email"]', 'reports@example.com');
    
    // Start export
    await page.click('[data-testid="start-export"]');
    await expect(page.locator('[data-testid="export-started-success"]')).toBeVisible();
    
    // Test immediate export download
    await page.click('[data-testid="download-current-data"]');
    
    // Wait for download to start
    const downloadPromise = page.waitForEvent('download');
    const download = await downloadPromise;
    
    expect(download.suggestedFilename()).toMatch(/performance-metrics-.*\.csv/);
  });

  test('Multi-node performance comparison', async ({ page }) => {
    await page.goto('/monitoring/comparison');
    
    // Select nodes for comparison
    const nodeSelector = page.locator('[data-testid="node-selector"]');
    await expect(nodeSelector).toBeVisible();
    
    // Select multiple nodes
    await page.check('[data-testid="select-node-1"]');
    await page.check('[data-testid="select-node-2"]');
    await page.check('[data-testid="select-node-3"]');
    
    // Select metrics to compare
    await page.check('[data-testid="compare-cpu"]');
    await page.check('[data-testid="compare-memory"]');
    await page.check('[data-testid="compare-latency"]');
    
    // Generate comparison
    await page.click('[data-testid="generate-comparison"]');
    
    // Verify comparison charts appear
    const comparisonChart = page.locator('[data-testid="multi-node-comparison-chart"]');
    await expect(comparisonChart).toBeVisible();
    
    // Check for legend showing all selected nodes
    const chartLegend = comparisonChart.locator('[data-testid="chart-legend"]');
    await expect(chartLegend.locator('[data-node="node-1"]')).toBeVisible();
    await expect(chartLegend.locator('[data-node="node-2"]')).toBeVisible();
    await expect(chartLegend.locator('[data-node="node-3"]')).toBeVisible();
    
    // Test node selection toggle
    await chartLegend.locator('[data-node="node-2"]').click();
    
    // Node 2 should be hidden from chart
    await expect(chartLegend.locator('[data-node="node-2"]')).toHaveClass(/legend-item-disabled/);
  });

  test('Performance trend analysis', async ({ page }) => {
    await page.goto('/monitoring/trends');
    
    // Select analysis period
    await page.selectOption('[data-testid="analysis-period"]', '30d');
    
    // Configure trend analysis
    await page.selectOption('[data-testid="trend-metric"]', 'response_time');
    await page.selectOption('[data-testid="trend-granularity"]', 'hourly');
    
    // Run trend analysis
    await page.click('[data-testid="run-trend-analysis"]');
    
    // Wait for analysis results
    const trendResults = page.locator('[data-testid="trend-analysis-results"]');
    await expect(trendResults).toBeVisible({ timeout: 30000 });
    
    // Verify trend chart
    const trendChart = trendResults.locator('[data-testid="trend-chart"]');
    await expect(trendChart).toBeVisible();
    
    // Check trend statistics
    const trendStats = trendResults.locator('[data-testid="trend-statistics"]');
    await expect(trendStats.locator('[data-testid="average-value"]')).toBeVisible();
    await expect(trendStats.locator('[data-testid="trend-direction"]')).toBeVisible();
    await expect(trendStats.locator('[data-testid="volatility-score"]')).toBeVisible();
    
    // Test forecasting
    await page.click('[data-testid="enable-forecasting"]');
    await page.selectOption('[data-testid="forecast-period"]', '7d');
    await page.click('[data-testid="generate-forecast"]');
    
    // Verify forecast appears on chart
    const forecastLine = trendChart.locator('[data-testid="forecast-line"]');
    await expect(forecastLine).toBeVisible({ timeout: 15000 });
    
    // Check confidence intervals
    const confidenceInterval = trendChart.locator('[data-testid="confidence-interval"]');
    await expect(confidenceInterval).toBeVisible();
  });

  test('Resource utilization heatmaps', async ({ page }) => {
    await page.goto('/monitoring/heatmaps');
    
    // Configure heatmap view
    await page.selectOption('[data-testid="heatmap-metric"]', 'cpu_utilization');
    await page.selectOption('[data-testid="time-granularity"]', '1h');
    await page.selectOption('[data-testid="time-period"]', '7d');
    
    // Generate heatmap
    await page.click('[data-testid="generate-heatmap"]');
    
    // Verify heatmap visualization
    const heatmap = page.locator('[data-testid="resource-heatmap"]');
    await expect(heatmap).toBeVisible({ timeout: 15000 });
    
    // Check heatmap axes
    const xAxis = heatmap.locator('[data-testid="heatmap-x-axis"]');
    const yAxis = heatmap.locator('[data-testid="heatmap-y-axis"]');
    await expect(xAxis).toBeVisible();
    await expect(yAxis).toBeVisible();
    
    // Test heatmap interactivity
    const heatmapCells = heatmap.locator('[data-testid^="heatmap-cell-"]');
    const firstCell = heatmapCells.first();
    
    await firstCell.hover();
    
    // Verify tooltip appears
    const tooltip = page.locator('[data-testid="heatmap-tooltip"]');
    await expect(tooltip).toBeVisible();
    await expect(tooltip).toContainText('CPU Utilization');
    
    // Test color scale legend
    const colorScale = page.locator('[data-testid="heatmap-color-scale"]');
    await expect(colorScale).toBeVisible();
    await expect(colorScale.locator('[data-testid="scale-min"]')).toBeVisible();
    await expect(colorScale.locator('[data-testid="scale-max"]')).toBeVisible();
  });

  test('Performance anomaly detection', async ({ page }) => {
    await page.goto('/monitoring/anomalies');
    
    // Configure anomaly detection
    await page.selectOption('[data-testid="anomaly-metric"]', 'response_time');
    await page.selectOption('[data-testid="detection-sensitivity"]', 'medium');
    await page.selectOption('[data-testid="baseline-period"]', '7d');
    
    // Enable machine learning detection
    await page.check('[data-testid="enable-ml-detection"]');
    await page.selectOption('[data-testid="ml-algorithm"]', 'isolation_forest');
    
    // Run anomaly detection
    await page.click('[data-testid="run-anomaly-detection"]');
    
    // Wait for detection results
    const anomalyResults = page.locator('[data-testid="anomaly-detection-results"]');
    await expect(anomalyResults).toBeVisible({ timeout: 60000 });
    
    // Verify anomaly timeline
    const anomalyTimeline = anomalyResults.locator('[data-testid="anomaly-timeline"]');
    await expect(anomalyTimeline).toBeVisible();
    
    // Check for anomaly markers
    const anomalyMarkers = anomalyTimeline.locator('[data-testid^="anomaly-marker-"]');
    const markerCount = await anomalyMarkers.count();
    
    if (markerCount > 0) {
      // Test anomaly details
      await anomalyMarkers.first().click();
      
      const anomalyDetails = page.locator('[data-testid="anomaly-details-panel"]');
      await expect(anomalyDetails).toBeVisible();
      
      // Verify anomaly information
      await expect(anomalyDetails.locator('[data-testid="anomaly-timestamp"]')).toBeVisible();
      await expect(anomalyDetails.locator('[data-testid="anomaly-severity"]')).toBeVisible();
      await expect(anomalyDetails.locator('[data-testid="anomaly-deviation"]')).toBeVisible();
      await expect(anomalyDetails.locator('[data-testid="potential-causes"]')).toBeVisible();
    }
    
    // Test anomaly alert configuration
    await page.click('[data-testid="configure-anomaly-alerts"]');
    
    const alertConfig = page.locator('[data-testid="anomaly-alert-config"]');
    await expect(alertConfig).toBeVisible();
    
    await page.check('[data-testid="enable-anomaly-alerts"]');
    await page.selectOption('[data-testid="alert-severity-threshold"]', 'medium');
    await page.check('[data-testid="email-anomaly-alerts"]');
    
    await page.click('[data-testid="save-anomaly-alert-config"]');
    await expect(page.locator('[data-testid="anomaly-alerts-configured"]')).toBeVisible();
  });
});