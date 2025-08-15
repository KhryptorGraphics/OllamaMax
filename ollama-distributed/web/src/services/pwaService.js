/**
 * PWA Service
 * 
 * Manages Progressive Web App features including service worker registration,
 * push notifications, and offline functionality.
 */

class PWAService {
  constructor() {
    this.swRegistration = null;
    this.isOnline = navigator.onLine;
    this.installPrompt = null;
    this.notificationPermission = 'default';
    
    this.init();
  }

  // Initialize PWA service
  async init() {
    // Register service worker
    await this.registerServiceWorker();
    
    // Setup online/offline listeners
    this.setupNetworkListeners();
    
    // Setup install prompt listener
    this.setupInstallPrompt();
    
    // Check notification permission
    this.checkNotificationPermission();
    
    // Setup beforeinstallprompt listener
    this.setupBeforeInstallPrompt();
  }

  // Register service worker
  async registerServiceWorker() {
    if ('serviceWorker' in navigator) {
      try {
        console.log('[PWA] Registering service worker...');
        
        this.swRegistration = await navigator.serviceWorker.register('/sw.js', {
          scope: '/'
        });
        
        console.log('[PWA] Service worker registered successfully');
        
        // Handle service worker updates
        this.swRegistration.addEventListener('updatefound', () => {
          const newWorker = this.swRegistration.installing;
          
          newWorker.addEventListener('statechange', () => {
            if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
              // New service worker is available
              this.showUpdateAvailable();
            }
          });
        });
        
        // Listen for messages from service worker
        navigator.serviceWorker.addEventListener('message', (event) => {
          this.handleServiceWorkerMessage(event);
        });
        
        return this.swRegistration;
      } catch (error) {
        console.error('[PWA] Service worker registration failed:', error);
        return null;
      }
    } else {
      console.warn('[PWA] Service workers not supported');
      return null;
    }
  }

  // Setup network status listeners
  setupNetworkListeners() {
    window.addEventListener('online', () => {
      console.log('[PWA] Back online');
      this.isOnline = true;
      this.onNetworkStatusChange(true);
    });
    
    window.addEventListener('offline', () => {
      console.log('[PWA] Gone offline');
      this.isOnline = false;
      this.onNetworkStatusChange(false);
    });
  }

  // Setup install prompt
  setupInstallPrompt() {
    // Check if app is already installed
    if (window.matchMedia('(display-mode: standalone)').matches) {
      console.log('[PWA] App is running in standalone mode');
      return;
    }
    
    // Check if app can be installed
    if ('getInstalledRelatedApps' in navigator) {
      navigator.getInstalledRelatedApps().then((relatedApps) => {
        if (relatedApps.length > 0) {
          console.log('[PWA] Related app is already installed');
        }
      });
    }
  }

  // Setup beforeinstallprompt event
  setupBeforeInstallPrompt() {
    window.addEventListener('beforeinstallprompt', (event) => {
      console.log('[PWA] Install prompt available');
      
      // Prevent the mini-infobar from appearing
      event.preventDefault();
      
      // Store the event for later use
      this.installPrompt = event;
      
      // Notify app that install is available
      this.onInstallAvailable();
    });
    
    window.addEventListener('appinstalled', () => {
      console.log('[PWA] App was installed');
      this.installPrompt = null;
      this.onAppInstalled();
    });
  }

  // Show install prompt
  async showInstallPrompt() {
    if (!this.installPrompt) {
      console.warn('[PWA] Install prompt not available');
      return false;
    }
    
    try {
      // Show the install prompt
      this.installPrompt.prompt();
      
      // Wait for user response
      const result = await this.installPrompt.userChoice;
      
      console.log('[PWA] Install prompt result:', result.outcome);
      
      // Clear the prompt
      this.installPrompt = null;
      
      return result.outcome === 'accepted';
    } catch (error) {
      console.error('[PWA] Error showing install prompt:', error);
      return false;
    }
  }

  // Check notification permission
  checkNotificationPermission() {
    if ('Notification' in window) {
      this.notificationPermission = Notification.permission;
      console.log('[PWA] Notification permission:', this.notificationPermission);
    } else {
      console.warn('[PWA] Notifications not supported');
    }
  }

  // Request notification permission
  async requestNotificationPermission() {
    if (!('Notification' in window)) {
      console.warn('[PWA] Notifications not supported');
      return false;
    }
    
    if (this.notificationPermission === 'granted') {
      return true;
    }
    
    try {
      const permission = await Notification.requestPermission();
      this.notificationPermission = permission;
      
      console.log('[PWA] Notification permission result:', permission);
      
      if (permission === 'granted') {
        await this.subscribeToPushNotifications();
        return true;
      }
      
      return false;
    } catch (error) {
      console.error('[PWA] Error requesting notification permission:', error);
      return false;
    }
  }

  // Subscribe to push notifications
  async subscribeToPushNotifications() {
    if (!this.swRegistration) {
      console.warn('[PWA] Service worker not registered');
      return null;
    }
    
    try {
      // Check if already subscribed
      let subscription = await this.swRegistration.pushManager.getSubscription();
      
      if (!subscription) {
        // Create new subscription
        const vapidPublicKey = await this.getVAPIDPublicKey();
        
        subscription = await this.swRegistration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: this.urlBase64ToUint8Array(vapidPublicKey)
        });
        
        console.log('[PWA] Push subscription created');
      } else {
        console.log('[PWA] Already subscribed to push notifications');
      }
      
      // Send subscription to server
      await this.sendSubscriptionToServer(subscription);
      
      return subscription;
    } catch (error) {
      console.error('[PWA] Error subscribing to push notifications:', error);
      return null;
    }
  }

  // Get VAPID public key from server
  async getVAPIDPublicKey() {
    try {
      const response = await fetch('/api/v1/push/vapid-public-key');
      const data = await response.json();
      return data.publicKey;
    } catch (error) {
      console.error('[PWA] Error getting VAPID key:', error);
      // Fallback key for development
      return 'BEl62iUYgUivxIkv69yViEuiBIa40HI80NM9f8HnKJuOmLWjMpS_7VnYkYdYWjZfkpZn4_qKSXgdvVhSVsNAT9w';
    }
  }

  // Send subscription to server
  async sendSubscriptionToServer(subscription) {
    try {
      const response = await fetch('/api/v1/push/subscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('ollama_auth_token')}`
        },
        body: JSON.stringify({
          subscription: subscription.toJSON()
        })
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      console.log('[PWA] Subscription sent to server');
    } catch (error) {
      console.error('[PWA] Error sending subscription to server:', error);
    }
  }

  // Convert VAPID key
  urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding)
      .replace(/-/g, '+')
      .replace(/_/g, '/');
    
    const rawData = window.atob(base64);
    const outputArray = new Uint8Array(rawData.length);
    
    for (let i = 0; i < rawData.length; ++i) {
      outputArray[i] = rawData.charCodeAt(i);
    }
    
    return outputArray;
  }

  // Show local notification
  showNotification(title, options = {}) {
    if (this.notificationPermission !== 'granted') {
      console.warn('[PWA] Notification permission not granted');
      return;
    }
    
    const defaultOptions = {
      icon: '/icons/icon-192x192.png',
      badge: '/icons/badge-72x72.png',
      tag: 'ollama-notification',
      requireInteraction: false,
      silent: false
    };
    
    const notificationOptions = { ...defaultOptions, ...options };
    
    if (this.swRegistration) {
      this.swRegistration.showNotification(title, notificationOptions);
    } else {
      new Notification(title, notificationOptions);
    }
  }

  // Handle service worker messages
  handleServiceWorkerMessage(event) {
    const { data } = event;
    
    switch (data.type) {
      case 'CACHE_UPDATED':
        console.log('[PWA] Cache updated');
        break;
      
      case 'OFFLINE_READY':
        console.log('[PWA] App ready for offline use');
        this.onOfflineReady();
        break;
      
      default:
        console.log('[PWA] Unknown message from service worker:', data);
    }
  }

  // Show update available notification
  showUpdateAvailable() {
    console.log('[PWA] App update available');
    this.onUpdateAvailable();
  }

  // Update service worker
  async updateServiceWorker() {
    if (this.swRegistration && this.swRegistration.waiting) {
      this.swRegistration.waiting.postMessage({ type: 'SKIP_WAITING' });
      
      // Reload page after update
      window.location.reload();
    }
  }

  // Get app info
  getAppInfo() {
    return {
      isOnline: this.isOnline,
      isInstalled: window.matchMedia('(display-mode: standalone)').matches,
      canInstall: !!this.installPrompt,
      notificationPermission: this.notificationPermission,
      serviceWorkerSupported: 'serviceWorker' in navigator,
      pushSupported: 'PushManager' in window
    };
  }

  // Event handlers (to be overridden by app)
  onNetworkStatusChange(isOnline) {
    // Override in app
    console.log('[PWA] Network status changed:', isOnline);
  }

  onInstallAvailable() {
    // Override in app
    console.log('[PWA] Install available');
  }

  onAppInstalled() {
    // Override in app
    console.log('[PWA] App installed');
  }

  onUpdateAvailable() {
    // Override in app
    console.log('[PWA] Update available');
  }

  onOfflineReady() {
    // Override in app
    console.log('[PWA] Offline ready');
  }

  // Cleanup
  destroy() {
    // Remove event listeners
    window.removeEventListener('online', this.onNetworkStatusChange);
    window.removeEventListener('offline', this.onNetworkStatusChange);
    window.removeEventListener('beforeinstallprompt', this.setupBeforeInstallPrompt);
    window.removeEventListener('appinstalled', this.onAppInstalled);
  }
}

// Create singleton instance
const pwaService = new PWAService();

export default pwaService;
