/**
 * Test Web Interface with Playwright
 * Actually loads the page and checks what's displayed
 */

const { chromium } = require('playwright');

async function testWebInterface() {
  console.log('🌐 Testing OllamaMax Web Interface...\n');

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  // Track network requests
  const requests = [];
  const responses = [];

  page.on('request', request => {
    requests.push({
      url: request.url(),
      method: request.method(),
      resourceType: request.resourceType()
    });
  });

  page.on('response', response => {
    responses.push({
      url: response.url(),
      status: response.status(),
      contentType: response.headers()['content-type']
    });
  });

  // Track console messages
  const consoleLogs = [];
  page.on('console', msg => {
    consoleLogs.push(`[${msg.type()}] ${msg.text()}`);
  });

  try {
    // Navigate to the web interface
    console.log('📍 Navigating to http://localhost:13000/');
    await page.goto('http://localhost:13000/', { waitUntil: 'networkidle', timeout: 10000 });

    // Wait a bit for any JavaScript to execute
    await page.waitForTimeout(3000);
    
    // Get page title
    const title = await page.title();
    console.log(`📄 Page Title: ${title}\n`);
    
    // Check what's actually visible on the page
    console.log('👁️  Checking visible elements...\n');
    
    // Check for main elements
    const hasHeader = await page.locator('header').count() > 0;
    const hasChatTab = await page.locator('[data-tab="chat"]').count() > 0;
    const hasNodesTab = await page.locator('[data-tab="nodes"]').count() > 0;
    const hasModelsTab = await page.locator('[data-tab="models"]').count() > 0;
    const hasSettingsTab = await page.locator('[data-tab="settings"]').count() > 0;
    const hasMessagesArea = await page.locator('#messagesArea').count() > 0;
    const hasInputArea = await page.locator('#messageInput').count() > 0;
    
    console.log('Element Check:');
    console.log(`  Header: ${hasHeader ? '✅' : '❌'}`);
    console.log(`  Chat Tab: ${hasChatTab ? '✅' : '❌'}`);
    console.log(`  Nodes Tab: ${hasNodesTab ? '✅' : '❌'}`);
    console.log(`  Models Tab: ${hasModelsTab ? '✅' : '❌'}`);
    console.log(`  Settings Tab: ${hasSettingsTab ? '✅' : '❌'}`);
    console.log(`  Messages Area: ${hasMessagesArea ? '✅' : '❌'}`);
    console.log(`  Input Area: ${hasInputArea ? '✅' : '❌'}\n`);
    
    // Get all text content
    const bodyText = await page.locator('body').textContent();
    console.log('📝 Visible Text Content:');
    console.log('─'.repeat(60));
    console.log(bodyText.substring(0, 500));
    console.log('─'.repeat(60));
    console.log('');
    
    // Check console logs
    console.log('📋 Console Logs:');
    if (consoleLogs.length > 0) {
      consoleLogs.slice(0, 20).forEach(log => console.log(`  ${log}`));
      if (consoleLogs.length > 20) {
        console.log(`  ... and ${consoleLogs.length - 20} more`);
      }
    } else {
      console.log('  (no console output)');
    }
    console.log('');
    
    // Check if CSS is loaded
    const styles = await page.evaluate(() => {
      const styleSheets = Array.from(document.styleSheets);
      return styleSheets.map(sheet => ({
        href: sheet.href,
        rules: sheet.cssRules ? sheet.cssRules.length : 0
      }));
    });
    
    console.log('🎨 Loaded Stylesheets:');
    styles.forEach(style => {
      console.log(`  - ${style.href || 'inline'}: ${style.rules} rules`);
    });
    console.log('');
    
    // Check if JavaScript is loaded
    const scripts = await page.evaluate(() => {
      return Array.from(document.scripts).map(s => s.src || 'inline');
    });
    
    console.log('📜 Loaded Scripts:');
    scripts.forEach(script => {
      console.log(`  - ${script}`);
    });
    console.log('');
    
    // Take a screenshot
    await page.screenshot({ path: 'web-interface-screenshot.png', fullPage: true });
    console.log('📸 Screenshot saved to: web-interface-screenshot.png\n');

    // Show network requests
    console.log('🌐 Network Requests:');
    requests.forEach(req => {
      console.log(`  ${req.method} ${req.url} (${req.resourceType})`);
    });
    console.log('');

    console.log('📥 Network Responses:');
    responses.forEach(res => {
      const status = res.status >= 200 && res.status < 300 ? '✅' : '❌';
      console.log(`  ${status} ${res.status} ${res.url}`);
      console.log(`     Content-Type: ${res.contentType || 'unknown'}`);
    });
    console.log('');
    
    // Final assessment
    console.log('═'.repeat(60));
    console.log('ASSESSMENT:');
    console.log('═'.repeat(60));
    
    if (hasHeader && hasChatTab && hasMessagesArea && hasInputArea) {
      console.log('✅ Web interface appears to be loading correctly');
      console.log('✅ All main UI elements are present');
    } else {
      console.log('❌ Web interface is NOT loading correctly');
      console.log('❌ Missing critical UI elements');
    }
    
    if (errors.length > 0) {
      console.log('⚠️  JavaScript errors detected - functionality may be broken');
    }
    
    console.log('═'.repeat(60));
    
  } catch (error) {
    console.error('❌ Error testing web interface:', error.message);
  } finally {
    await browser.close();
  }
}

// Run the test
testWebInterface().catch(console.error);

