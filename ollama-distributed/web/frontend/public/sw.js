// Service Worker for OllamaMax PWA
// Implements offline-first architecture with intelligent caching strategies

const CACHE_NAME = 'ollamamax-v1.2.0';
const RUNTIME_CACHE = 'ollamamax-runtime-v1';
const OFFLINE_CACHE = 'ollamamax-offline-v1';
const API_CACHE = 'ollamamax-api-v1';

// Cache strategies for different resource types
const CACHE_STRATEGIES = {
  CACHE_FIRST: 'cache-first',
  NETWORK_FIRST: 'network-first',
  STALE_WHILE_REVALIDATE: 'stale-while-revalidate',
  NETWORK_ONLY: 'network-only',
  CACHE_ONLY: 'cache-only'
};

// Static assets to cache on install (App Shell)
const STATIC_CACHE_URLS = [
  '/',
  '/v2',
  '/manifest.json',
  '/icons/icon-192x192.png',
  '/icons/icon-512x512.png',
  '/offline.html' // Fallback page for offline
];

// Dynamic cache configurations
const CACHE_CONFIG = {
  // Static assets - cache first
  static: {
    urlPatterns: [/\.(?:js|css|html)$/],
    strategy: CACHE_STRATEGIES.CACHE_FIRST,
    cacheName: CACHE_NAME,
    maxAge: 30 * 24 * 60 * 60, // 30 days
    maxEntries: 100
  },
  
  // Images - stale while revalidate
  images: {
    urlPatterns: [/\.(?:png|jpg|jpeg|svg|gif|webp)$/],
    strategy: CACHE_STRATEGIES.STALE_WHILE_REVALIDATE,
    cacheName: 'ollamamax-images-v1',
    maxAge: 7 * 24 * 60 * 60, // 7 days
    maxEntries: 50
  },
  
  // API calls - network first with fallback
  api: {
    urlPatterns: [/^https?:\/\/.*\/api\//],
    strategy: CACHE_STRATEGIES.NETWORK_FIRST,
    cacheName: API_CACHE,
    maxAge: 5 * 60, // 5 minutes
    maxEntries: 50
  },
  
  // WebSocket connections - network only
  websocket: {
    urlPatterns: [/^wss?:\/\//],
    strategy: CACHE_STRATEGIES.NETWORK_ONLY
  }
};

// Background sync tags
const SYNC_TAGS = {
  MODEL_SYNC: 'model-sync',
  METRICS_SYNC: 'metrics-sync',
  NOTIFICATIONS_SYNC: 'notifications-sync'
};

// Install event - cache static resources
self.addEventListener('install', (event) => {
  console.log('[SW] Installing service worker...');
  
  event.waitUntil(
    Promise.all([
      // Cache static resources
      caches.open(CACHE_NAME).then((cache) => {
        console.log('[SW] Caching static resources');
        return cache.addAll(STATIC_CACHE_URLS);
      }),
      
      // Create offline fallback page
      createOfflinePage(),
      
      // Skip waiting to activate immediately
      self.skipWaiting()
    ])
  );
});

// Activate event - clean old caches and claim clients
self.addEventListener('activate', (event) => {
  console.log('[SW] Activating service worker...');
  
  event.waitUntil(
    Promise.all([
      // Clean old caches
      caches.keys().then((cacheNames) => {
        return Promise.all(
          cacheNames
            .filter((cacheName) => {
              return cacheName.startsWith('ollamamax-') && 
                     cacheName !== CACHE_NAME && 
                     cacheName !== RUNTIME_CACHE &&
                     cacheName !== OFFLINE_CACHE &&
                     cacheName !== API_CACHE;
            })
            .map((cacheName) => {
              console.log('[SW] Deleting old cache:', cacheName);
              return caches.delete(cacheName);
            })
        );
      }),
      
      // Claim all clients
      self.clients.claim(),
      
      // Initialize IndexedDB for offline data
      initializeOfflineStorage()
    ])
  );
});

// Fetch event - implement caching strategies
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);
  
  // Skip non-GET requests for caching
  if (request.method !== 'GET') {
    // Handle POST/PUT/DELETE for offline sync
    if (isApiRequest(request)) {
      event.respondWith(handleApiRequest(request));
    }
    return;
  }
  
  // Apply caching strategy based on request type
  event.respondWith(
    handleRequest(request).catch((error) => {
      console.error('[SW] Request failed:', error);
      return handleOfflineFallback(request);
    })
  );
});

// Background sync event
self.addEventListener('sync', (event) => {
  console.log('[SW] Background sync triggered:', event.tag);
  
  switch (event.tag) {
    case SYNC_TAGS.MODEL_SYNC:
      event.waitUntil(syncModels());
      break;
    case SYNC_TAGS.METRICS_SYNC:
      event.waitUntil(syncMetrics());
      break;
    case SYNC_TAGS.NOTIFICATIONS_SYNC:
      event.waitUntil(syncNotifications());
      break;
  }
});

// Push notification event
self.addEventListener('push', (event) => {
  console.log('[SW] Push notification received');
  
  const options = {
    body: 'You have new updates from OllamaMax',
    icon: '/icons/icon-192x192.png',
    badge: '/icons/badge-72x72.png',
    vibrate: [200, 100, 200],
    data: {
      dateOfArrival: Date.now(),
      primaryKey: 1
    },
    actions: [
      {
        action: 'explore',
        title: 'Open Dashboard',
        icon: '/icons/action-dashboard.png'
      },
      {
        action: 'close',
        title: 'Close',
        icon: '/icons/action-close.png'
      }
    ]
  };
  
  if (event.data) {
    const payload = event.data.json();
    options.body = payload.body || options.body;
    options.data = { ...options.data, ...payload.data };
  }
  
  event.waitUntil(
    self.registration.showNotification('OllamaMax', options)
  );
});

// Notification click event
self.addEventListener('notificationclick', (event) => {
  console.log('[SW] Notification clicked:', event.action);
  
  event.notification.close();
  
  if (event.action === 'explore') {
    event.waitUntil(
      clients.openWindow('/v2')
    );
  }
});

// Message event - communicate with main thread
self.addEventListener('message', (event) => {
  console.log('[SW] Message received:', event.data);
  
  switch (event.data.type) {
    case 'SKIP_WAITING':
      self.skipWaiting();
      break;
      
    case 'CACHE_API_RESPONSE':
      cacheApiResponse(event.data.url, event.data.response);
      break;
      
    case 'GET_CACHE_STATUS':
      getCacheStatus().then((status) => {
        event.ports[0].postMessage(status);
      });
      break;
      
    case 'CLEAR_CACHE':
      clearCache(event.data.cacheName).then((success) => {
        event.ports[0].postMessage({ success });
      });
      break;
  }
});

// Request handling functions
async function handleRequest(request) {
  const url = new URL(request.url);
  
  // Determine caching strategy
  let config = null;
  for (const [key, value] of Object.entries(CACHE_CONFIG)) {
    if (value.urlPatterns.some(pattern => pattern.test(url.pathname))) {
      config = value;
      break;
    }
  }
  
  if (!config) {
    config = {
      strategy: CACHE_STRATEGIES.NETWORK_FIRST,
      cacheName: RUNTIME_CACHE
    };
  }
  
  switch (config.strategy) {
    case CACHE_STRATEGIES.CACHE_FIRST:
      return cacheFirst(request, config);
    case CACHE_STRATEGIES.NETWORK_FIRST:
      return networkFirst(request, config);
    case CACHE_STRATEGIES.STALE_WHILE_REVALIDATE:
      return staleWhileRevalidate(request, config);
    case CACHE_STRATEGIES.NETWORK_ONLY:
      return fetch(request);
    case CACHE_STRATEGIES.CACHE_ONLY:
      return caches.match(request);
    default:
      return networkFirst(request, config);
  }
}

async function cacheFirst(request, config) {
  const cachedResponse = await caches.match(request);
  if (cachedResponse && !isExpired(cachedResponse, config.maxAge)) {
    return cachedResponse;
  }
  
  const networkResponse = await fetch(request);
  const cache = await caches.open(config.cacheName);
  
  // Clone response for caching
  const responseClone = networkResponse.clone();
  cache.put(request, responseClone);
  
  // Clean old entries if needed
  if (config.maxEntries) {
    await cleanCache(config.cacheName, config.maxEntries);
  }
  
  return networkResponse;
}

async function networkFirst(request, config) {
  try {
    const networkResponse = await fetch(request);
    
    // Cache successful responses
    if (networkResponse.ok) {
      const cache = await caches.open(config.cacheName);
      const responseClone = networkResponse.clone();
      cache.put(request, responseClone);
      
      if (config.maxEntries) {
        await cleanCache(config.cacheName, config.maxEntries);
      }
    }
    
    return networkResponse;
  } catch (error) {
    // Fallback to cache
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    throw error;
  }
}

async function staleWhileRevalidate(request, config) {
  const cache = await caches.open(config.cacheName);
  const cachedResponse = await cache.match(request);
  
  // Background update
  const networkResponsePromise = fetch(request).then((response) => {
    if (response.ok) {
      cache.put(request, response.clone());
    }
    return response;
  });
  
  return cachedResponse || networkResponsePromise;
}

// Offline handling functions
async function handleOfflineFallback(request) {
  const url = new URL(request.url);
  
  // Return cached version if available
  const cachedResponse = await caches.match(request);
  if (cachedResponse) {
    return cachedResponse;
  }
  
  // For navigation requests, return offline page
  if (request.mode === 'navigate') {
    return caches.match('/offline.html');
  }
  
  // For API requests, return offline response
  if (isApiRequest(request)) {
    return new Response(
      JSON.stringify({
        error: 'Offline',
        message: 'This request is not available offline',
        offline: true
      }),
      {
        status: 503,
        headers: { 'Content-Type': 'application/json' }
      }
    );
  }
  
  // Return generic offline response
  return new Response('Offline content not available', {
    status: 503,
    statusText: 'Service Unavailable'
  });
}

async function createOfflinePage() {
  const cache = await caches.open(OFFLINE_CACHE);
  const offlineHTML = `
    <!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>OllamaMax - Offline</title>
      <style>
        body {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
          margin: 0;
          padding: 2rem;
          text-align: center;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
          min-height: 100vh;
          display: flex;
          flex-direction: column;
          justify-content: center;
        }
        .offline-content {
          max-width: 400px;
          margin: 0 auto;
        }
        .logo {
          width: 80px;
          height: 80px;
          margin: 0 auto 2rem;
          background: white;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 2rem;
          color: #667eea;
        }
        h1 { margin-bottom: 0.5rem; }
        p { margin-bottom: 2rem; opacity: 0.9; }
        button {
          background: rgba(255,255,255,0.2);
          border: 2px solid white;
          color: white;
          padding: 0.75rem 1.5rem;
          border-radius: 6px;
          cursor: pointer;
          font-size: 1rem;
          transition: all 0.2s;
        }
        button:hover {
          background: rgba(255,255,255,0.3);
        }
      </style>
    </head>
    <body>
      <div class="offline-content">
        <div class="logo">🤖</div>
        <h1>You're Offline</h1>
        <p>OllamaMax is currently unavailable. Check your internet connection and try again.</p>
        <button onclick="window.location.reload()">Try Again</button>
      </div>
    </body>
    </html>
  `;
  
  await cache.put('/offline.html', new Response(offlineHTML, {
    headers: { 'Content-Type': 'text/html' }
  }));
}

// Background sync functions
async function syncModels() {
  console.log('[SW] Syncing models...');
  
  try {
    // Get pending sync data from IndexedDB
    const pendingSync = await getFromIndexedDB('pendingSync', 'models');
    
    if (pendingSync && pendingSync.length > 0) {
      for (const item of pendingSync) {
        try {
          const response = await fetch('/api/models', {
            method: item.method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item.data)
          });
          
          if (response.ok) {
            await removeFromIndexedDB('pendingSync', item.id);
          }
        } catch (error) {
          console.error('[SW] Failed to sync model:', error);
        }
      }
    }
    
    // Notify clients about sync completion
    const clients = await self.clients.matchAll();
    clients.forEach(client => {
      client.postMessage({
        type: 'SYNC_COMPLETE',
        tag: SYNC_TAGS.MODEL_SYNC
      });
    });
    
  } catch (error) {
    console.error('[SW] Model sync failed:', error);
  }
}

async function syncMetrics() {
  console.log('[SW] Syncing metrics...');
  // Similar implementation for metrics sync
}

async function syncNotifications() {
  console.log('[SW] Syncing notifications...');
  // Similar implementation for notifications sync
}

// API request handling
async function handleApiRequest(request) {
  try {
    return await fetch(request);
  } catch (error) {
    // Store request for background sync
    const requestData = {
      id: generateId(),
      url: request.url,
      method: request.method,
      headers: Object.fromEntries(request.headers.entries()),
      data: request.method !== 'GET' ? await request.json() : null,
      timestamp: Date.now()
    };
    
    await storeInIndexedDB('pendingSync', requestData);
    
    // Register background sync
    if ('serviceWorker' in navigator && 'sync' in window.ServiceWorkerRegistration.prototype) {
      await self.registration.sync.register(SYNC_TAGS.MODEL_SYNC);
    }
    
    return new Response(JSON.stringify({
      offline: true,
      message: 'Request queued for sync when online',
      id: requestData.id
    }), {
      status: 202,
      headers: { 'Content-Type': 'application/json' }
    });
  }
}

// Utility functions
function isApiRequest(request) {
  const url = new URL(request.url);
  return url.pathname.startsWith('/api/');
}

function isExpired(response, maxAge) {
  if (!maxAge) return false;
  
  const dateHeader = response.headers.get('date');
  if (!dateHeader) return false;
  
  const date = new Date(dateHeader);
  return (Date.now() - date.getTime()) > (maxAge * 1000);
}

async function cleanCache(cacheName, maxEntries) {
  const cache = await caches.open(cacheName);
  const keys = await cache.keys();
  
  if (keys.length > maxEntries) {
    const keysToDelete = keys.slice(0, keys.length - maxEntries);
    await Promise.all(keysToDelete.map(key => cache.delete(key)));
  }
}

async function getCacheStatus() {
  const cacheNames = await caches.keys();
  const status = {};
  
  for (const cacheName of cacheNames) {
    const cache = await caches.open(cacheName);
    const keys = await cache.keys();
    status[cacheName] = keys.length;
  }
  
  return status;
}

async function clearCache(cacheName) {
  return await caches.delete(cacheName);
}

function generateId() {
  return Date.now().toString(36) + Math.random().toString(36).substr(2);
}

// IndexedDB helper functions
async function initializeOfflineStorage() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('OllamaMaxDB', 1);
    
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
    
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      
      // Create object stores
      if (!db.objectStoreNames.contains('pendingSync')) {
        db.createObjectStore('pendingSync', { keyPath: 'id' });
      }
      
      if (!db.objectStoreNames.contains('offlineData')) {
        db.createObjectStore('offlineData', { keyPath: 'key' });
      }
    };
  });
}

async function storeInIndexedDB(storeName, data) {
  const db = await initializeOfflineStorage();
  const transaction = db.transaction([storeName], 'readwrite');
  const store = transaction.objectStore(storeName);
  
  return new Promise((resolve, reject) => {
    const request = store.add(data);
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

async function getFromIndexedDB(storeName, key) {
  const db = await initializeOfflineStorage();
  const transaction = db.transaction([storeName], 'readonly');
  const store = transaction.objectStore(storeName);
  
  return new Promise((resolve, reject) => {
    const request = key ? store.get(key) : store.getAll();
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

async function removeFromIndexedDB(storeName, key) {
  const db = await initializeOfflineStorage();
  const transaction = db.transaction([storeName], 'readwrite');
  const store = transaction.objectStore(storeName);
  
  return new Promise((resolve, reject) => {
    const request = store.delete(key);
    request.onsuccess = () => resolve();
    request.onerror = () => reject(request.error);
  });
}

console.log('[SW] Service worker loaded successfully');