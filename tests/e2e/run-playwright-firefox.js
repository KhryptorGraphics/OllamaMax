const { firefox } = require('playwright');
const fs = require('fs');

(async () => {
  const BASE_URL = process.env.BASE_URL || 'http://localhost';
  const browser = await firefox.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await context.newPage();

  const screenshotsDir = './reports/screenshots-firefox';
  fs.mkdirSync(screenshotsDir, { recursive: true });

  try {
    await page.goto(`${BASE_URL}/test`, { waitUntil: 'networkidle' });
    await page.screenshot({ path: `${screenshotsDir}/test_page.png` });

    // Click Test API button
    await page.getByText('Test API').click();
    await page.waitForSelector('#api-response', { timeout: 15000 });

    // Check health via diagnostic endpoint
    await page.goto(`${BASE_URL}/diagnostic`, { waitUntil: 'networkidle' });
    await page.screenshot({ path: `${screenshotsDir}/diagnostic.png` });

    console.log('Firefox smoke: PASS');
  } catch (e) {
    console.error('Firefox smoke: FAIL', e);
    process.exitCode = 1;
  } finally {
    await browser.close();
  }
})();

