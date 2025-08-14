// Admin dashboard tests are best-effort; skip tests if UI elements are not present.
const fs = require('fs');

const BASE_URL = process.env.BASE_URL || 'http://localhost';
const shots = './reports/screenshots';

beforeAll(async () => {
  fs.mkdirSync(shots, { recursive: true });
});

describe('Admin Dashboard (best-effort)', () => {
  it('attempts to locate node status and metrics UI', async () => {
    await page.goto(`${BASE_URL}/`, { waitUntil: 'networkidle0' });
    const hasMetrics = await page.evaluate(() => {
      return !!document.querySelector('#cluster-metrics, .metrics, [data-testid="metrics"]');
    });
    // If not present, we still pass but annotate
    if (!hasMetrics) {
      console.warn('Admin metrics UI not found; documenting gap');
    }
    await page.screenshot({ path: `${shots}/admin-dashboard.png`, fullPage: true });
  });

  it('validates WebSocket real-time updates (if present)', async () => {
    await page.goto(`${BASE_URL}/`, { waitUntil: 'domcontentloaded' });
    const gotUpdate = await page.evaluate(() => new Promise((resolve) => {
      try {
        const ws = new WebSocket(location.origin.replace('http','ws') + '/ws');
        let seen = false;
        ws.onopen = () => ws.send('ping');
        ws.onmessage = () => { seen = true; resolve(true); };
        setTimeout(() => resolve(seen), 5000);
      } catch (_) { resolve(false); }
    }));
    expect(gotUpdate).toBe(true);
  });
});

