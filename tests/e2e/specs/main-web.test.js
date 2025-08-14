const fs = require('fs');

const BASE_URL = process.env.BASE_URL || 'http://localhost';
const shots = './reports/screenshots';

beforeAll(async () => {
  fs.mkdirSync(shots, { recursive: true });
});

describe('Main Web Interface', () => {
  it('loads index and captures performance metrics', async () => {
    await page.goto(`${BASE_URL}/`, { waitUntil: 'networkidle0' });
    const timing = await page.evaluate(() => JSON.stringify(performance.getEntriesByType('navigation')[0]));
    fs.writeFileSync('./reports/index-perf.json', timing);
    await page.screenshot({ path: `${shots}/index.png`, fullPage: true });
  });

  it('loads /test and passes API connectivity', async () => {
    await page.goto(`${BASE_URL}/test`, { waitUntil: 'networkidle0' });
    await page.click('text=Test API');
    await page.waitForSelector('#api-response', { timeout: 15000 });
    const apiText = await page.$eval('#api-response', el => el.textContent);
    expect(apiText).toMatch(/webHealth/);
    await page.screenshot({ path: `${shots}/test_api.png`, fullPage: true });
  });

  it('validates React and CDN loading', async () => {
    await page.goto(`${BASE_URL}/`, { waitUntil: 'domcontentloaded' });
    // Check CDN scripts present
    const hasReact = await page.evaluate(() => !!window.React && !!window.ReactDOM);
    expect(hasReact).toBe(true);
  });

  it('tests responsive layouts', async () => {
    for (const size of [[375, 812], [768, 1024], [1280, 800]]) {
      await page.setViewport({ width: size[0], height: size[1] });
      await page.goto(`${BASE_URL}/`, { waitUntil: 'networkidle0' });
      await page.screenshot({ path: `${shots}/index_${size[0]}x${size[1]}.png`, fullPage: true });
    }
  });

  it('verifies WebSocket connectivity', async () => {
    await page.goto(`${BASE_URL}/`, { waitUntil: 'domcontentloaded' });
    const wsMessage = await page.evaluate(() => new Promise((resolve, reject) => {
      const ws = new WebSocket(location.origin.replace('http','ws') + '/ws');
      let timer = setTimeout(() => reject(new Error('timeout')), 10000);
      ws.onopen = () => ws.send('ping');
      ws.onmessage = (ev) => { clearTimeout(timer); resolve(ev.data); };
      ws.onerror = (e) => reject(e);
    }));
    expect(wsMessage).toBeTruthy();
  });
});

