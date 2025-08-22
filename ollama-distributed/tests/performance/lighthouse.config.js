module.exports = {
  ci: {
    collect: {
      url: [
        'http://localhost:3000',
        'http://localhost:3000/login',
        'http://localhost:3000/dashboard',
        'http://localhost:3000/admin',
        'http://localhost:3000/monitoring'
      ],
      startServerCommand: 'npm run dev',
      startServerReadyPattern: 'ready',
      startServerReadyTimeout: 30000,
      numberOfRuns: 3,
      settings: {
        chromeFlags: '--no-sandbox --headless'
      }
    },
    assert: {
      assertions: {
        'categories:performance': ['error', { minScore: 0.9 }],
        'categories:accessibility': ['error', { minScore: 0.9 }],
        'categories:best-practices': ['error', { minScore: 0.9 }],
        'categories:seo': ['error', { minScore: 0.8 }],
        'categories:pwa': ['warn', { minScore: 0.8 }],
        
        // Performance metrics
        'first-contentful-paint': ['error', { maxNumericValue: 1500 }],
        'largest-contentful-paint': ['error', { maxNumericValue: 2500 }],
        'first-meaningful-paint': ['warn', { maxNumericValue: 2000 }],
        'speed-index': ['error', { maxNumericValue: 3000 }],
        'interactive': ['error', { maxNumericValue: 3500 }],
        'cumulative-layout-shift': ['error', { maxNumericValue: 0.1 }],
        'total-blocking-time': ['warn', { maxNumericValue: 300 }],
        
        // Resource optimization
        'unused-css-rules': ['warn', { maxNumericValue: 10000 }],
        'unused-javascript': ['warn', { maxNumericValue: 10000 }],
        'render-blocking-resources': ['warn', { maxNumericValue: 500 }],
        'unminified-css': ['error', { maxNumericValue: 0 }],
        'unminified-javascript': ['error', { maxNumericValue: 0 }],
        'modern-image-formats': ['warn', { minScore: 0.8 }],
        'efficient-animated-content': ['warn', { minScore: 0.8 }],
        
        // Network and caching
        'uses-long-cache-ttl': ['warn', { minScore: 0.75 }],
        'uses-optimized-images': ['warn', { minScore: 0.8 }],
        'uses-webp-images': ['warn', { minScore: 0.8 }],
        'uses-responsive-images': ['warn', { minScore: 0.8 }],
        'offscreen-images': ['warn', { minScore: 0.8 }],
        
        // Accessibility
        'color-contrast': ['error', { minScore: 1 }],
        'image-alt': ['error', { minScore: 1 }],
        'button-name': ['error', { minScore: 1 }],
        'link-name': ['error', { minScore: 1 }],
        'document-title': ['error', { minScore: 1 }],
        'html-has-lang': ['error', { minScore: 1 }],
        'meta-description': ['warn', { minScore: 1 }],
        
        // Security
        'is-on-https': ['error', { minScore: 1 }],
        'uses-http2': ['warn', { minScore: 1 }],
        'no-vulnerable-libraries': ['error', { minScore: 1 }],
        
        // PWA
        'service-worker': ['warn', { minScore: 1 }],
        'offline-start-url': ['warn', { minScore: 1 }],
        'is-on-https': ['error', { minScore: 1 }],
        'redirects-http': ['warn', { minScore: 1 }],
        'viewport': ['error', { minScore: 1 }],
        'without-javascript': ['warn', { minScore: 1 }],
        
        // Best practices
        'uses-passive-event-listeners': ['warn', { minScore: 1 }],
        'no-document-write': ['error', { minScore: 1 }],
        'external-anchors-use-rel-noopener': ['error', { minScore: 1 }],
        'geolocation-on-start': ['error', { minScore: 1 }],
        'notification-on-start': ['error', { minScore: 1 }]
      }
    },
    upload: {
      target: 'filesystem',
      outputDir: './tests/reports/lighthouse'
    },
    server: {
      port: 9001,
      storage: './tests/reports/lighthouse'
    }
  }
};