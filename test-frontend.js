import { chromium } from 'playwright';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

class FrontendTester {
    constructor() {
        this.browser = null;
        this.page = null;
        this.testResults = [];
        this.screenshotDir = path.join(__dirname, 'test-screenshots');
        
        if (!fs.existsSync(this.screenshotDir)) {
            fs.mkdirSync(this.screenshotDir, { recursive: true });
        }
    }

    async initialize() {
        console.log('🚀 Starting comprehensive frontend testing...');
        this.browser = await chromium.launch({ 
            headless: true,
            slowMo: 500
        });
        this.page = await this.browser.newPage();
        
        this.page.on('console', msg => {
            console.log(`Browser Console: ${msg.text()}`);
        });
        
        this.page.on('pageerror', error => {
            console.error(`Page Error: ${error.message}`);
            this.testResults.push({
                component: 'Page Error',
                status: 'ERROR',
                error: error.message
            });
        });
    }

    async startWebInterface() {
        console.log('📡 Starting web interface...');
        try {
            await this.page.goto('file:///home/kp/ollamamax/web-interface/index.html');
            await this.page.waitForTimeout(2000);
            
            await this.page.screenshot({
                path: path.join(this.screenshotDir, '01-initial-load.png'),
                fullPage: true
            });
            
            console.log('✅ Web interface loaded successfully');
            return true;
        } catch (error) {
            console.error(`❌ Failed to load web interface: ${error.message}`);
            return false;
        }
    }

    async testNavigationTabs() {
        console.log('📋 Testing navigation tabs...');
        const tabs = ['chat', 'nodes', 'models', 'settings'];
        
        for (const tab of tabs) {
            try {
                console.log(`  Testing ${tab} tab...`);
                
                const tabButton = await this.page.locator(`[data-tab="${tab}"]`);
                await tabButton.click();
                await this.page.waitForTimeout(1000);
                
                await this.page.screenshot({
                    path: path.join(this.screenshotDir, `02-tab-${tab}.png`),
                    fullPage: true
                });
                
                const isActive = await tabButton.getAttribute('class');
                const tabContent = await this.page.locator(`#${tab}-tab`).isVisible();
                
                this.testResults.push({
                    component: `${tab} Tab`,
                    status: (isActive.includes('active') && tabContent) ? 'PASS' : 'FAIL',
                    details: `Active: ${isActive.includes('active')}, Content Visible: ${tabContent}`
                });
                
                console.log(`    ${tab} tab: ${isActive.includes('active') ? '✅' : '❌'}`);
            } catch (error) {
                console.error(`    ${tab} tab error: ${error.message}`);
                this.testResults.push({
                    component: `${tab} Tab`,
                    status: 'ERROR',
                    error: error.message
                });
            }
        }
    }

    async testChatInterface() {
        console.log('💬 Testing chat interface...');
        
        await this.page.locator('[data-tab="chat"]').click();
        await this.page.waitForTimeout(1000);
        
        const chatElements = [
            { selector: '#chat-messages', name: 'Chat Messages Container' },
            { selector: '#user-input', name: 'User Input Field' },
            { selector: '#send-button', name: 'Send Button' },
            { selector: '#clear-chat', name: 'Clear Chat Button' },
            { selector: '#model-select', name: 'Model Select Dropdown' }
        ];
        
        for (const element of chatElements) {
            try {
                const locator = this.page.locator(element.selector);
                const isVisible = await locator.isVisible();
                const isEnabled = await locator.isEnabled();
                
                await this.page.screenshot({
                    path: path.join(this.screenshotDir, `03-chat-${element.name.toLowerCase().replace(/\s+/g, '-')}.png`),
                    fullPage: true
                });
                
                this.testResults.push({
                    component: element.name,
                    status: (isVisible && isEnabled) ? 'PASS' : 'FAIL',
                    details: `Visible: ${isVisible}, Enabled: ${isEnabled}`
                });
                
                console.log(`  ${element.name}: ${isVisible && isEnabled ? '✅' : '❌'}`);
            } catch (error) {
                console.error(`  ${element.name} error: ${error.message}`);
                this.testResults.push({
                    component: element.name,
                    status: 'ERROR',
                    error: error.message
                });
            }
        }
        
        try {
            await this.page.fill('#user-input', 'Test message');
            await this.page.click('#send-button');
            await this.page.waitForTimeout(2000);
            
            await this.page.screenshot({
                path: path.join(this.screenshotDir, '04-chat-send-test.png'),
                fullPage: true
            });
            
            console.log('  Message send test completed');
        } catch (error) {
            console.error(`  Message send test error: ${error.message}`);
        }
    }

    async testNodesInterface() {
        console.log('🔗 Testing nodes interface...');
        
        await this.page.locator('[data-tab="nodes"]').click();
        await this.page.waitForTimeout(1000);
        
        const nodeElements = [
            { selector: '#nodes-list', name: 'Nodes List Container' },
            { selector: '#add-node-btn', name: 'Add Node Button' },
            { selector: '#refresh-nodes', name: 'Refresh Nodes Button' },
            { selector: '#node-url-input', name: 'Node URL Input' }
        ];
        
        for (const element of nodeElements) {
            try {
                const locator = this.page.locator(element.selector);
                const isVisible = await locator.isVisible();
                const isEnabled = await locator.isEnabled();
                
                await this.page.screenshot({
                    path: path.join(this.screenshotDir, `05-nodes-${element.name.toLowerCase().replace(/\s+/g, '-')}.png`),
                    fullPage: true
                });
                
                this.testResults.push({
                    component: element.name,
                    status: (isVisible && isEnabled) ? 'PASS' : 'FAIL',
                    details: `Visible: ${isVisible}, Enabled: ${isEnabled}`
                });
                
                console.log(`  ${element.name}: ${isVisible && isEnabled ? '✅' : '❌'}`);
            } catch (error) {
                console.error(`  ${element.name} error: ${error.message}`);
                this.testResults.push({
                    component: element.name,
                    status: 'ERROR',
                    error: error.message
                });
            }
        }
        
        try {
            await this.page.click('#refresh-nodes');
            await this.page.waitForTimeout(2000);
            
            await this.page.screenshot({
                path: path.join(this.screenshotDir, '06-nodes-refresh-test.png'),
                fullPage: true
            });
            
            console.log('  Nodes refresh test completed');
        } catch (error) {
            console.error(`  Nodes refresh test error: ${error.message}`);
        }
    }

    async testModelsInterface() {
        console.log('🧠 Testing models interface...');
        
        await this.page.locator('[data-tab="models"]').click();
        await this.page.waitForTimeout(1000);
        
        const modelElements = [
            { selector: '#models-list', name: 'Models List Container' },
            { selector: '#refresh-models', name: 'Refresh Models Button' },
            { selector: '#download-model-input', name: 'Download Model Input' },
            { selector: '#download-model-btn', name: 'Download Model Button' }
        ];
        
        for (const element of modelElements) {
            try {
                const locator = this.page.locator(element.selector);
                const isVisible = await locator.isVisible();
                const isEnabled = await locator.isEnabled();
                
                await this.page.screenshot({
                    path: path.join(this.screenshotDir, `07-models-${element.name.toLowerCase().replace(/\s+/g, '-')}.png`),
                    fullPage: true
                });
                
                this.testResults.push({
                    component: element.name,
                    status: (isVisible && isEnabled) ? 'PASS' : 'FAIL',
                    details: `Visible: ${isVisible}, Enabled: ${isEnabled}`
                });
                
                console.log(`  ${element.name}: ${isVisible && isEnabled ? '✅' : '❌'}`);
            } catch (error) {
                console.error(`  ${element.name} error: ${error.message}`);
                this.testResults.push({
                    component: element.name,
                    status: 'ERROR',
                    error: error.message
                });
            }
        }
        
        try {
            await this.page.click('#refresh-models');
            await this.page.waitForTimeout(2000);
            
            await this.page.screenshot({
                path: path.join(this.screenshotDir, '08-models-refresh-test.png'),
                fullPage: true
            });
            
            console.log('  Models refresh test completed');
        } catch (error) {
            console.error(`  Models refresh test error: ${error.message}`);
        }
    }

    async testSettingsInterface() {
        console.log('⚙️ Testing settings interface...');
        
        await this.page.locator('[data-tab="settings"]').click();
        await this.page.waitForTimeout(1000);
        
        const settingsElements = [
            { selector: '#api-endpoint', name: 'API Endpoint Input' },
            { selector: '#max-tokens', name: 'Max Tokens Input' },
            { selector: '#temperature', name: 'Temperature Slider' },
            { selector: '#save-settings', name: 'Save Settings Button' },
            { selector: '#reset-settings', name: 'Reset Settings Button' }
        ];
        
        for (const element of settingsElements) {
            try {
                const locator = this.page.locator(element.selector);
                const isVisible = await locator.isVisible();
                const isEnabled = await locator.isEnabled();
                
                await this.page.screenshot({
                    path: path.join(this.screenshotDir, `09-settings-${element.name.toLowerCase().replace(/\s+/g, '-')}.png`),
                    fullPage: true
                });
                
                this.testResults.push({
                    component: element.name,
                    status: (isVisible && isEnabled) ? 'PASS' : 'FAIL',
                    details: `Visible: ${isVisible}, Enabled: ${isEnabled}`
                });
                
                console.log(`  ${element.name}: ${isVisible && isEnabled ? '✅' : '❌'}`);
            } catch (error) {
                console.error(`  ${element.name} error: ${error.message}`);
                this.testResults.push({
                    component: element.name,
                    status: 'ERROR',
                    error: error.message
                });
            }
        }
    }

    async testAllButtons() {
        console.log('🔘 Testing all clickable buttons...');
        
        const allButtons = await this.page.locator('button, .btn, [role="button"]').all();
        
        for (let i = 0; i < allButtons.length; i++) {
            try {
                const button = allButtons[i];
                const buttonText = await button.textContent() || `Button-${i}`;
                const isVisible = await button.isVisible();
                const isEnabled = await button.isEnabled();
                
                if (isVisible) {
                    await this.page.screenshot({
                        path: path.join(this.screenshotDir, `10-button-${buttonText.toLowerCase().replace(/\s+/g, '-')}-${i}.png`),
                        fullPage: true
                    });
                    
                    this.testResults.push({
                        component: `Button: ${buttonText}`,
                        status: isEnabled ? 'PASS' : 'DISABLED',
                        details: `Text: "${buttonText}", Enabled: ${isEnabled}`
                    });
                    
                    console.log(`  ${buttonText}: ${isEnabled ? '✅' : '⚠️  (disabled)'}`);
                }
            } catch (error) {
                console.error(`  Button ${i} error: ${error.message}`);
            }
        }
    }

    async testAllInputFields() {
        console.log('📝 Testing all input fields...');
        
        const allInputs = await this.page.locator('input, textarea, select').all();
        
        for (let i = 0; i < allInputs.length; i++) {
            try {
                const input = allInputs[i];
                const inputId = await input.getAttribute('id') || `Input-${i}`;
                const inputType = await input.getAttribute('type') || 'text';
                const isVisible = await input.isVisible();
                const isEnabled = await input.isEnabled();
                
                if (isVisible) {
                    await this.page.screenshot({
                        path: path.join(this.screenshotDir, `11-input-${inputId.toLowerCase()}-${i}.png`),
                        fullPage: true
                    });
                    
                    this.testResults.push({
                        component: `Input: ${inputId}`,
                        status: isEnabled ? 'PASS' : 'DISABLED',
                        details: `ID: "${inputId}", Type: "${inputType}", Enabled: ${isEnabled}`
                    });
                    
                    console.log(`  ${inputId} (${inputType}): ${isEnabled ? '✅' : '⚠️  (disabled)'}`);
                }
            } catch (error) {
                console.error(`  Input ${i} error: ${error.message}`);
            }
        }
    }

    async testWebSocketConnection() {
        console.log('🔌 Testing WebSocket connection...');
        
        try {
            const wsStatus = await this.page.evaluate(() => {
                return new Promise((resolve) => {
                    const ws = new WebSocket('ws://localhost:13100/chat');
                    
                    ws.onopen = () => {
                        ws.close();
                        resolve('CONNECTED');
                    };
                    
                    ws.onerror = () => {
                        resolve('FAILED');
                    };
                    
                    setTimeout(() => {
                        resolve('TIMEOUT');
                    }, 5000);
                });
            });
            
            this.testResults.push({
                component: 'WebSocket Connection',
                status: wsStatus === 'CONNECTED' ? 'PASS' : 'FAIL',
                details: `Connection status: ${wsStatus}`
            });
            
            console.log(`  WebSocket: ${wsStatus === 'CONNECTED' ? '✅' : '❌'} (${wsStatus})`);
        } catch (error) {
            console.error(`  WebSocket test error: ${error.message}`);
            this.testResults.push({
                component: 'WebSocket Connection',
                status: 'ERROR',
                error: error.message
            });
        }
    }

    async generateReport() {
        console.log('📊 Generating comprehensive test report...');
        
        const report = {
            timestamp: new Date().toISOString(),
            totalTests: this.testResults.length,
            passed: this.testResults.filter(r => r.status === 'PASS').length,
            failed: this.testResults.filter(r => r.status === 'FAIL').length,
            errors: this.testResults.filter(r => r.status === 'ERROR').length,
            disabled: this.testResults.filter(r => r.status === 'DISABLED').length,
            results: this.testResults
        };
        
        const reportPath = path.join(__dirname, 'frontend-test-report.json');
        fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
        
        const summaryPath = path.join(__dirname, 'frontend-test-summary.md');
        const summary = `# Frontend Test Summary

## Overview
- **Total Tests**: ${report.totalTests}
- **Passed**: ${report.passed} ✅
- **Failed**: ${report.failed} ❌
- **Errors**: ${report.errors} 🚨
- **Disabled**: ${report.disabled} ⚠️
- **Test Date**: ${new Date().toLocaleString()}

## Failed/Error Components
${this.testResults
    .filter(r => r.status === 'FAIL' || r.status === 'ERROR')
    .map(r => `- **${r.component}**: ${r.status} - ${r.details || r.error}`)
    .join('\n')}

## All Test Results
${this.testResults
    .map(r => `- ${r.component}: ${r.status} ${r.status === 'PASS' ? '✅' : r.status === 'FAIL' ? '❌' : r.status === 'ERROR' ? '🚨' : '⚠️'}`)
    .join('\n')}

## Screenshots
All UI component screenshots have been saved to: \`test-screenshots/\`
`;
        
        fs.writeFileSync(summaryPath, summary);
        
        console.log(`\n📋 Test Report Generated:`);
        console.log(`  JSON Report: ${reportPath}`);
        console.log(`  Summary: ${summaryPath}`);
        console.log(`  Screenshots: ${this.screenshotDir}`);
        
        return report;
    }

    async close() {
        if (this.browser) {
            await this.browser.close();
        }
    }

    async runFullTest() {
        try {
            await this.initialize();
            
            const loaded = await this.startWebInterface();
            if (!loaded) {
                console.error('❌ Cannot continue - web interface failed to load');
                return;
            }
            
            await this.testNavigationTabs();
            await this.testChatInterface();
            await this.testNodesInterface();
            await this.testModelsInterface();
            await this.testSettingsInterface();
            await this.testAllButtons();
            await this.testAllInputFields();
            await this.testWebSocketConnection();
            
            const report = await this.generateReport();
            
            console.log(`\n🎉 Testing Complete!`);
            console.log(`📊 Results: ${report.passed} passed, ${report.failed} failed, ${report.errors} errors`);
            
        } catch (error) {
            console.error(`❌ Test execution failed: ${error.message}`);
        } finally {
            await this.close();
        }
    }
}

if (import.meta.url === `file://${process.argv[1]}`) {
    const tester = new FrontendTester();
    tester.runFullTest().catch(console.error);
}

export default FrontendTester;