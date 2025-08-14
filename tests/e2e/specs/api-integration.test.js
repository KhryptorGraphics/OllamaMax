const fs = require('fs');

const BASE_URL = process.env.BASE_URL || 'http://localhost';
const shots = './reports/screenshots';

beforeAll(async () => {
  fs.mkdirSync(shots, { recursive: true });
});

describe('API Integration via Web', () => {
  it('calls /api/v1/health through the LB and validates JSON', async () => {
    const res = await page.goto(`${BASE_URL}/api/v1/health`, { waitUntil: 'networkidle0' });
    expect(res.status()).toBeLessThan(400);
    const body = await res.json();
    expect(body).toHaveProperty('status');
  });

  it('handles API error responses gracefully', async () => {
    const res = await page.goto(`${BASE_URL}/api/v1/does-not-exist`, { waitUntil: 'networkidle0' });
    expect(res.status()).toBeGreaterThanOrEqual(400);
  });

  it('validates CORS by fetching from page context', async () => {
    await page.goto(`${BASE_URL}/test`, { waitUntil: 'networkidle0' });
    const status = await page.evaluate(async () => {
      const r = await fetch('/api/v1/health', { mode: 'cors' });
      return r.status;
    });
    expect(status).toBe(200);
  });
});

