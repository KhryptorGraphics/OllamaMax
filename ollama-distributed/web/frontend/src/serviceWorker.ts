/**
 * Service Worker for Progressive Web App (PWA)
 * Handles caching, offline functionality, and background sync
 */

const CACHE_NAME = 'ollama-distributed-v1'
const STATIC_CACHE = 'static-v1'
const DYNAMIC_CACHE = 'dynamic-v1'
const API_CACHE = 'api-v1'

// Assets to cache immediately
const STATIC_ASSETS = [
  '/',
  '/index.html',
  '/manifest.json',
  '/icons/icon-192x192.png',
  '/icons/icon-512x512.png',
  '/static/js/main.js',
  '/static/css/main.css'
]

// API endpoints to cache
const CACHEABLE_APIS = [
  '/api/cluster/status',
  '/api/models',
  '/api/monitoring/metrics'
]

// Maximum cache sizes
const MAX_STATIC_ITEMS = 100
const MAX_DYNAMIC_ITEMS = 50
const MAX_API_ITEMS = 30

self.addEventListener('install', (event: ExtendableEvent) => {
  console.log('Service Worker installing...')
  
  event.waitUntil(
    caches.open(STATIC_CACHE)
      .then(cache => {
        console.log('Caching static assets')
        return cache.addAll(STATIC_ASSETS)
      })
      .then(() => {
        // Force activation of new service worker
        return (self as any).skipWaiting()
      })
      .catch(error => {
        console.error('Failed to cache static assets:', error)
      })
  )
})

self.addEventListener('activate', (event: ExtendableEvent) => {
  console.log('Service Worker activating...')
  
  event.waitUntil(
    Promise.all([
      // Clean up old caches
      caches.keys().then(cacheNames => {
        return Promise.all(
          cacheNames.map(cacheName => {
            if (cacheName !== STATIC_CACHE && cacheName !== DYNAMIC_CACHE && cacheName !== API_CACHE) {
              console.log('Deleting old cache:', cacheName)
              return caches.delete(cacheName)
            }
          })
        )
      }),
      // Take control of all clients
      (self as any).clients.claim()
    ])
  )
})

self.addEventListener('fetch', (event: FetchEvent) => {
  const { request } = event
  const url = new URL(request.url)

  // Handle different types of requests
  if (request.method === 'GET') {
    if (isStaticAsset(request)) {
      event.respondWith(handleStaticAsset(request))
    } else if (isAPIRequest(request)) {
      event.respondWith(handleAPIRequest(request))
    } else if (isNavigationRequest(request)) {
      event.respondWith(handleNavigationRequest(request))
    } else {
      event.respondWith(handleDynamicRequest(request))
    }
  }
})

// Handle static assets (CSS, JS, images)
function handleStaticAsset(request: Request): Promise<Response> {
  return caches.open(STATIC_CACHE)
    .then(cache => {
      return cache.match(request)
        .then(response => {
          if (response) {
            return response
          }
          
          // Fetch and cache if not found
          return fetch(request)
            .then(fetchResponse => {
              if (fetchResponse.ok) {
                cache.put(request, fetchResponse.clone())
              }
              return fetchResponse
            })
        })
    })
    .catch(() => {
      // Return offline fallback for images
      if (request.destination === 'image') {
        return new Response(`
          <svg width="200" height="200" xmlns="http://www.w3.org/2000/svg">
            <rect width="200" height="200" fill="#f0f0f0"/>
            <text x="100" y="100" text-anchor="middle" dy=".3em" fill="#999">Offline</text>
          </svg>
        `, {
          headers: { 'Content-Type': 'image/svg+xml' }
        })
      }
      throw new Error('Resource not available offline')
    })
}

// Handle API requests with caching strategy
function handleAPIRequest(request: Request): Promise<Response> {
  const url = new URL(request.url)
  
  // Different strategies for different endpoints
  if (url.pathname.includes('/status') || url.pathname.includes('/metrics')) {
    // Stale-while-revalidate for frequently updated data
    return staleWhileRevalidate(request, API_CACHE)
  } else if (url.pathname.includes('/models')) {
    // Cache-first for relatively static data
    return cacheFirst(request, API_CACHE)
  } else {
    // Network-first for other API calls
    return networkFirst(request, API_CACHE)
  }
}

// Handle navigation requests (page loads)
function handleNavigationRequest(request: Request): Promise<Response> {
  return networkFirst(request, DYNAMIC_CACHE)
    .catch(() => {
      // Return cached index.html for offline navigation
      return caches.match('/index.html')
        .then(response => {
          if (response) {
            return response
          }
          throw new Error('No offline page available')
        })
    })
}

// Handle other dynamic requests
function handleDynamicRequest(request: Request): Promise<Response> {
  return networkFirst(request, DYNAMIC_CACHE)
}

// Caching strategies
function cacheFirst(request: Request, cacheName: string): Promise<Response> {
  return caches.open(cacheName)
    .then(cache => {
      return cache.match(request)
        .then(response => {
          if (response) {
            return response
          }
          
          return fetch(request)
            .then(fetchResponse => {
              if (fetchResponse.ok) {
                cache.put(request, fetchResponse.clone())
                limitCacheSize(cacheName, getMaxCacheSize(cacheName))
              }
              return fetchResponse
            })
        })
    })
}

function networkFirst(request: Request, cacheName: string): Promise<Response> {
  return fetch(request)
    .then(response => {
      if (response.ok) {
        caches.open(cacheName)
          .then(cache => {
            cache.put(request, response.clone())
            limitCacheSize(cacheName, getMaxCacheSize(cacheName))
          })
      }
      return response
    })
    .catch(() => {
      return caches.match(request)
        .then(response => {
          if (response) {
            return response
          }
          throw new Error('Request failed and no cache available')
        })
    })
}

function staleWhileRevalidate(request: Request, cacheName: string): Promise<Response> {
  return caches.open(cacheName)
    .then(cache => {
      return cache.match(request)
        .then(cachedResponse => {
          const fetchPromise = fetch(request)
            .then(response => {
              if (response.ok) {
                cache.put(request, response.clone())
                limitCacheSize(cacheName, getMaxCacheSize(cacheName))
              }
              return response
            })
            .catch(() => cachedResponse)

          return cachedResponse || fetchPromise
        })
    })
}

// Utility functions
function isStaticAsset(request: Request): boolean {
  const url = new URL(request.url)
  return url.pathname.startsWith('/static/') ||
         url.pathname.includes('.css') ||
         url.pathname.includes('.js') ||
         url.pathname.includes('.png') ||
         url.pathname.includes('.jpg') ||
         url.pathname.includes('.svg') ||
         url.pathname.includes('.ico')
}

function isAPIRequest(request: Request): boolean {
  const url = new URL(request.url)
  return url.pathname.startsWith('/api/')
}

function isNavigationRequest(request: Request): boolean {
  return request.mode === 'navigate' ||
         (request.method === 'GET' && request.headers.get('accept')?.includes('text/html'))
}

function getMaxCacheSize(cacheName: string): number {
  switch (cacheName) {
    case STATIC_CACHE: return MAX_STATIC_ITEMS
    case DYNAMIC_CACHE: return MAX_DYNAMIC_ITEMS
    case API_CACHE: return MAX_API_ITEMS
    default: return 50
  }
}

function limitCacheSize(cacheName: string, maxItems: number): Promise<void> {
  return caches.open(cacheName)
    .then(cache => {
      return cache.keys()
        .then(keys => {
          if (keys.length > maxItems) {
            // Delete oldest items (first in, first out)
            const itemsToDelete = keys.slice(0, keys.length - maxItems)
            return Promise.all(
              itemsToDelete.map(key => cache.delete(key))
            )
          }
        })
    })
}

// Background sync for offline actions
self.addEventListener('sync', (event: any) => {
  console.log('Background sync triggered:', event.tag)
  
  if (event.tag === 'background-sync') {
    event.waitUntil(doBackgroundSync())
  }
})

async function doBackgroundSync(): Promise<void> {
  try {
    // Get offline actions from IndexedDB
    const offlineActions = await getOfflineActions()
    
    for (const action of offlineActions) {
      try {
        await fetch(action.url, {
          method: action.method,
          headers: action.headers,
          body: action.body
        })
        
        // Remove from offline queue on success
        await removeOfflineAction(action.id)
        
        // Notify clients of successful sync
        await notifyClients({
          type: 'sync-success',
          action: action.type
        })
        
      } catch (error) {
        console.error('Failed to sync action:', action, error)
        
        // Increment retry count
        await incrementRetryCount(action.id)
        
        // Remove if max retries reached
        if (action.retries >= 3) {
          await removeOfflineAction(action.id)
          await notifyClients({
            type: 'sync-failed',
            action: action.type,
            error: 'Max retries reached'
          })
        }
      }
    }
  } catch (error) {
    console.error('Background sync failed:', error)
  }
}

// Push notifications
self.addEventListener('push', (event: any) => {
  console.log('Push notification received:', event)
  
  const options = {
    body: 'You have new updates in Ollama Distributed',
    icon: '/icons/icon-192x192.png',
    badge: '/icons/badge.png',
    vibrate: [100, 50, 100],
    data: {
      dateOfArrival: Date.now(),
      primaryKey: 1
    },
    actions: [
      {
        action: 'explore',
        title: 'Open App',
        icon: '/icons/checkmark.png'
      },
      {
        action: 'close',
        title: 'Close',
        icon: '/icons/xmark.png'
      }
    ]
  }
  
  event.waitUntil(
    self.registration.showNotification('Ollama Distributed', options)
  )
})

// Notification click handling
self.addEventListener('notificationclick', (event: any) => {
  console.log('Notification click received:', event)
  
  event.notification.close()
  
  if (event.action === 'explore') {
    event.waitUntil(
      (self as any).clients.openWindow('/')
    )
  }
})

// Helper functions for IndexedDB operations
function getOfflineActions(): Promise<any[]> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('OfflineActions', 1)
    
    request.onerror = () => reject(request.error)
    request.onsuccess = () => {
      const db = request.result
      const transaction = db.transaction(['actions'], 'readonly')
      const store = transaction.objectStore('actions')
      const getAllRequest = store.getAll()
      
      getAllRequest.onsuccess = () => resolve(getAllRequest.result)
      getAllRequest.onerror = () => reject(getAllRequest.error)
    }
    
    request.onupgradeneeded = () => {
      const db = request.result
      const store = db.createObjectStore('actions', { keyPath: 'id' })
      store.createIndex('timestamp', 'timestamp', { unique: false })
    }
  })
}

function removeOfflineAction(id: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('OfflineActions', 1)
    
    request.onsuccess = () => {
      const db = request.result
      const transaction = db.transaction(['actions'], 'readwrite')
      const store = transaction.objectStore('actions')
      const deleteRequest = store.delete(id)
      
      deleteRequest.onsuccess = () => resolve()
      deleteRequest.onerror = () => reject(deleteRequest.error)
    }
  })
}

function incrementRetryCount(id: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('OfflineActions', 1)
    
    request.onsuccess = () => {
      const db = request.result
      const transaction = db.transaction(['actions'], 'readwrite')
      const store = transaction.objectStore('actions')
      const getRequest = store.get(id)
      
      getRequest.onsuccess = () => {
        const action = getRequest.result
        if (action) {
          action.retries = (action.retries || 0) + 1
          store.put(action)
        }
        resolve()
      }
      getRequest.onerror = () => reject(getRequest.error)
    }
  })
}

function notifyClients(message: any): Promise<void> {
  return (self as any).clients.matchAll()
    .then((clients: any[]) => {
      clients.forEach(client => {
        client.postMessage(message)
      })
    })
}

// Cache warming for critical resources
self.addEventListener('message', (event: any) => {
  if (event.data && event.data.type === 'WARM_CACHE') {
    event.waitUntil(warmCache(event.data.urls))
  }
  
  if (event.data && event.data.type === 'SKIP_WAITING') {
    (self as any).skipWaiting()
  }
})

function warmCache(urls: string[]): Promise<void> {
  return caches.open(DYNAMIC_CACHE)
    .then(cache => {
      return Promise.all(
        urls.map(url => {
          return fetch(url)
            .then(response => {
              if (response.ok) {
                return cache.put(url, response)
              }
            })
            .catch(error => {
              console.warn('Failed to warm cache for:', url, error)
            })
        })
      )
    })
    .then(() => {
      console.log('Cache warming completed')
    })
}

export {}