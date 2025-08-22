import { useState, useEffect, useCallback } from 'react';

interface PushNotificationState {
  permission: NotificationPermission;
  isSupported: boolean;
  isSubscribed: boolean;
  subscription: PushSubscription | null;
  isLoading: boolean;
  error: string | null;
}

interface PushNotificationActions {
  requestPermission: () => Promise<boolean>;
  subscribe: () => Promise<boolean>;
  unsubscribe: () => Promise<boolean>;
  sendNotification: (title: string, options?: NotificationOptions) => Promise<void>;
  clearError: () => void;
}

export interface PushNotificationHookReturn extends PushNotificationState, PushNotificationActions {}

// VAPID keys for push notifications (in production, these should be from environment variables)
const VAPID_PUBLIC_KEY = 'BEl62iUYgUivxIkv69yViEuiBIa40HI80xeSNzy73GU';

// Convert VAPID key to Uint8Array
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - base64String.length % 4) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');

  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

export function usePushNotifications(): PushNotificationHookReturn {
  const [state, setState] = useState<PushNotificationState>({
    permission: Notification.permission,
    isSupported: 'Notification' in window && 'serviceWorker' in navigator && 'PushManager' in window,
    isSubscribed: false,
    subscription: null,
    isLoading: false,
    error: null,
  });

  // Check existing subscription
  const checkSubscription = useCallback(async () => {
    if (!state.isSupported) return;

    try {
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.getSubscription();
      
      setState(prev => ({
        ...prev,
        isSubscribed: !!subscription,
        subscription: subscription,
      }));

      return subscription;
    } catch (error) {
      console.error('[Push] Failed to check subscription:', error);
      setState(prev => ({
        ...prev,
        error: 'Failed to check subscription status',
      }));
    }
  }, [state.isSupported]);

  // Request notification permission
  const requestPermission = useCallback(async (): Promise<boolean> => {
    if (!state.isSupported) {
      setState(prev => ({
        ...prev,
        error: 'Push notifications are not supported in this browser',
      }));
      return false;
    }

    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      const permission = await Notification.requestPermission();
      
      setState(prev => ({
        ...prev,
        permission,
        isLoading: false,
      }));

      if (permission === 'granted') {
        console.log('[Push] Notification permission granted');
        
        // Track permission granted
        if ((window as any).gtag) {
          (window as any).gtag('event', 'notification_permission_granted', {
            event_category: 'Push Notifications',
          });
        }
        
        return true;
      } else {
        console.log('[Push] Notification permission denied');
        
        // Track permission denied
        if ((window as any).gtag) {
          (window as any).gtag('event', 'notification_permission_denied', {
            event_category: 'Push Notifications',
          });
        }
        
        setState(prev => ({
          ...prev,
          error: 'Notification permission was denied',
        }));
        
        return false;
      }
    } catch (error) {
      console.error('[Push] Failed to request permission:', error);
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: 'Failed to request notification permission',
      }));
      return false;
    }
  }, [state.isSupported]);

  // Subscribe to push notifications
  const subscribe = useCallback(async (): Promise<boolean> => {
    if (state.permission !== 'granted') {
      const granted = await requestPermission();
      if (!granted) return false;
    }

    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      const registration = await navigator.serviceWorker.ready;
      
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY),
      });

      // Send subscription to backend
      await fetch('/api/push/subscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          subscription: subscription.toJSON(),
          userAgent: navigator.userAgent,
          timestamp: new Date().toISOString(),
        }),
      });

      setState(prev => ({
        ...prev,
        isSubscribed: true,
        subscription,
        isLoading: false,
      }));

      console.log('[Push] Successfully subscribed to push notifications');
      
      // Track subscription
      if ((window as any).gtag) {
        (window as any).gtag('event', 'push_subscription', {
          event_category: 'Push Notifications',
        });
      }

      // Store subscription info locally for offline access
      localStorage.setItem('push-subscription', JSON.stringify(subscription.toJSON()));

      return true;
    } catch (error) {
      console.error('[Push] Failed to subscribe:', error);
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: 'Failed to subscribe to push notifications',
      }));
      return false;
    }
  }, [state.permission, requestPermission]);

  // Unsubscribe from push notifications
  const unsubscribe = useCallback(async (): Promise<boolean> => {
    if (!state.subscription) {
      return true;
    }

    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      // Unsubscribe from push manager
      await state.subscription.unsubscribe();

      // Notify backend
      await fetch('/api/push/unsubscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          endpoint: state.subscription.endpoint,
        }),
      });

      setState(prev => ({
        ...prev,
        isSubscribed: false,
        subscription: null,
        isLoading: false,
      }));

      console.log('[Push] Successfully unsubscribed from push notifications');
      
      // Track unsubscription
      if ((window as any).gtag) {
        (window as any).gtag('event', 'push_unsubscription', {
          event_category: 'Push Notifications',
        });
      }

      // Remove stored subscription
      localStorage.removeItem('push-subscription');

      return true;
    } catch (error) {
      console.error('[Push] Failed to unsubscribe:', error);
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: 'Failed to unsubscribe from push notifications',
      }));
      return false;
    }
  }, [state.subscription]);

  // Send a local notification
  const sendNotification = useCallback(async (
    title: string,
    options: NotificationOptions = {}
  ): Promise<void> => {
    if (state.permission !== 'granted') {
      throw new Error('Notification permission not granted');
    }

    try {
      const notification = new Notification(title, {
        icon: '/icons/icon-192x192.png',
        badge: '/icons/badge-72x72.png',
        vibrate: [200, 100, 200],
        tag: 'ollamamax-notification',
        renotify: true,
        requireInteraction: false,
        ...options,
      });

      // Handle notification click
      notification.onclick = (event) => {
        event.preventDefault();
        notification.close();
        
        // Focus or open app window
        if ('clients' in navigator.serviceWorker) {
          (async () => {
            const clients = await navigator.serviceWorker.ready;
            const windowClients = await (clients as any).clients.matchAll({
              type: 'window',
              includeUncontrolled: true,
            });

            if (windowClients.length > 0) {
              windowClients[0].focus();
            } else {
              window.open('/', '_blank');
            }
          })();
        } else {
          window.focus();
        }

        // Track notification click
        if ((window as any).gtag) {
          (window as any).gtag('event', 'notification_click', {
            event_category: 'Push Notifications',
            event_label: title,
          });
        }
      };

      // Auto-close after 5 seconds
      setTimeout(() => {
        notification.close();
      }, 5000);

      console.log('[Push] Notification sent:', title);
    } catch (error) {
      console.error('[Push] Failed to send notification:', error);
      throw error;
    }
  }, [state.permission]);

  // Clear error
  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  // Initialize and check subscription on mount
  useEffect(() => {
    if (state.isSupported) {
      checkSubscription();
    }
  }, [state.isSupported, checkSubscription]);

  // Listen for permission changes
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (!document.hidden && state.isSupported) {
        // Check if permission changed when app becomes visible
        const currentPermission = Notification.permission;
        if (currentPermission !== state.permission) {
          setState(prev => ({ ...prev, permission: currentPermission }));
          
          // Re-check subscription if permission was granted
          if (currentPermission === 'granted') {
            checkSubscription();
          }
        }
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [state.isSupported, state.permission, checkSubscription]);

  // Handle service worker messages
  useEffect(() => {
    if (!('serviceWorker' in navigator)) return;

    const handleMessage = (event: MessageEvent) => {
      if (event.data?.type === 'NOTIFICATION_CLICKED') {
        console.log('[Push] Notification clicked via service worker');
        
        // Handle notification click actions
        if (event.data.action === 'open_dashboard') {
          window.location.href = '/v2';
        }
      }
    };

    navigator.serviceWorker.addEventListener('message', handleMessage);

    return () => {
      navigator.serviceWorker.removeEventListener('message', handleMessage);
    };
  }, []);

  return {
    // State
    ...state,
    
    // Actions
    requestPermission,
    subscribe,
    unsubscribe,
    sendNotification,
    clearError,
  };
}