import { chromium } from 'playwright';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function quickTest() {
    console.log('🚀 Quick frontend test after fixing JavaScript errors...');
    
    const browser = await chromium.launch({ headless: true });
    const page = await browser.newPage();
    
    const screenshotDir = path.join(__dirname, 'fixed-test-screenshots');
    if (!fs.existsSync(screenshotDir)) {
        fs.mkdirSync(screenshotDir, { recursive: true });
    }
    
    // Monitor console messages and errors
    const issues = [];
    page.on('console', msg => {
        if (msg.type() === 'error') {
            issues.push(`Console Error: ${msg.text()}`);
        }
        console.log(`Browser: ${msg.type()}: ${msg.text()}`);
    });
    
    page.on('pageerror', error => {
        issues.push(`Page Error: ${error.message}`);
        console.error(`Page Error: ${error.message}`);
    });
    
    try {
        // Load the page
        await page.goto('file:///home/kp/ollamamax/web-interface/index.html');
        await page.waitForTimeout(3000);
        
        // Take initial screenshot
        await page.screenshot({
            path: path.join(screenshotDir, '01-initial-fixed.png'),
            fullPage: true
        });
        
        // Test navigation tabs
        const tabs = ['chat', 'nodes', 'models', 'settings'];
        for (const tab of tabs) {
            try {
                console.log(`Testing ${tab} tab...`);
                await page.click(`[data-tab="${tab}"]`);
                await page.waitForTimeout(1000);
                
                await page.screenshot({
                    path: path.join(screenshotDir, `02-tab-${tab}-fixed.png`),
                    fullPage: true
                });
                
                const isActive = await page.locator(`[data-tab="${tab}"]`).getAttribute('class');
                const tabContent = await page.locator(`#${tab}-tab`).isVisible();
                
                console.log(`  ${tab}: Active=${isActive?.includes('active')}, Content=${tabContent}`);
            } catch (error) {
                issues.push(`Tab ${tab} error: ${error.message}`);
            }
        }
        
        // Test chat functionality
        console.log('Testing chat interface...');
        await page.click('[data-tab="chat"]');
        await page.waitForTimeout(1000);
        
        try {
            // Check if input field exists and is functional
            const userInput = page.locator('#user-input');
            if (await userInput.isVisible()) {
                await userInput.fill('Test message for UI testing');
                await page.screenshot({
                    path: path.join(screenshotDir, '03-chat-input-test.png'),
                    fullPage: true
                });
                console.log('  ✅ Chat input field works');
            }
            
            // Test send button
            const sendButton = page.locator('#send-button');
            if (await sendButton.isVisible()) {
                console.log('  ✅ Send button exists');
            }
        } catch (error) {
            issues.push(`Chat test error: ${error.message}`);
        }
        
        // Test nodes interface
        console.log('Testing nodes interface...');
        await page.click('[data-tab="nodes"]');
        await page.waitForTimeout(1000);
        
        await page.screenshot({
            path: path.join(screenshotDir, '04-nodes-interface.png'),
            fullPage: true
        });
        
        // Test models interface
        console.log('Testing models interface...');
        await page.click('[data-tab="models"]');
        await page.waitForTimeout(1000);
        
        await page.screenshot({
            path: path.join(screenshotDir, '05-models-interface.png'),
            fullPage: true
        });
        
        // Test settings interface
        console.log('Testing settings interface...');
        await page.click('[data-tab="settings"]');
        await page.waitForTimeout(1000);
        
        await page.screenshot({
            path: path.join(screenshotDir, '06-settings-interface.png'),
            fullPage: true
        });
        
        // Create final report
        const report = {
            timestamp: new Date().toISOString(),
            totalIssues: issues.length,
            issues: issues,
            status: issues.length === 0 ? 'PASS' : 'ISSUES_FOUND'
        };
        
        fs.writeFileSync(path.join(__dirname, 'quick-test-report.json'), JSON.stringify(report, null, 2));
        
        console.log(`\n📊 Quick Test Results:`);
        console.log(`  Issues Found: ${issues.length}`);
        console.log(`  Screenshots: ${screenshotDir}`);
        console.log(`  Status: ${report.status}`);
        
        if (issues.length > 0) {
            console.log(`\n❌ Issues detected:`);
            issues.forEach((issue, i) => console.log(`  ${i + 1}. ${issue}`));
        } else {
            console.log(`\n✅ No JavaScript errors detected!`);
        }
        
    } catch (error) {
        console.error(`❌ Test failed: ${error.message}`);
    } finally {
        await browser.close();
    }
}

quickTest().catch(console.error);