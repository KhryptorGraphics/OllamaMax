import { useState, useEffect, useCallback } from 'react';

interface BeforeInstallPromptEvent extends Event {
  readonly platforms: string[];
  readonly userChoice: Promise<{
    outcome: 'accepted' | 'dismissed';
    platform: string;
  }>;
  prompt(): Promise<void>;
}

interface PWAInstallState {
  isInstallable: boolean;
  isInstalled: boolean;
  isStandalone: boolean;
  canInstall: boolean;
  installPrompt: BeforeInstallPromptEvent | null;
  isOnline: boolean;
  updateAvailable: boolean;
}

interface PWAActions {
  installApp: () => Promise<boolean>;
  updateApp: () => Promise<void>;
  dismissInstallPrompt: () => void;
  checkForUpdates: () => Promise<boolean>;
  clearCache: () => Promise<void>;
  getCacheStatus: () => Promise<Record<string, number>>;
}

export interface PWAHookReturn extends PWAInstallState, PWAActions {}

// Detect if running in standalone mode (installed PWA)
const isStandaloneMode = (): boolean => {
  return (
    window.matchMedia('(display-mode: standalone)').matches ||
    window.matchMedia('(display-mode: fullscreen)').matches ||
    (window.navigator as any).standalone === true ||
    document.referrer.includes('android-app://')
  );
};

// Detect if PWA is installable
const isPWAInstallable = (): boolean => {
  return (
    'serviceWorker' in navigator &&
    window.matchMedia('(display-mode: browser)').matches &&
    !isStandaloneMode()
  );
};

// Check if app is already installed
const isPWAInstalled = (): boolean => {
  return isStandaloneMode() || localStorage.getItem('pwa-installed') === 'true';
};

export function usePWA(): PWAHookReturn {
  const [state, setState] = useState<PWAInstallState>({
    isInstallable: isPWAInstallable(),
    isInstalled: isPWAInstalled(),
    isStandalone: isStandaloneMode(),
    canInstall: false,
    installPrompt: null,
    isOnline: navigator.onLine,
    updateAvailable: false,
  });

  // Handle install prompt
  const handleBeforeInstallPrompt = useCallback((e: Event) => {
    e.preventDefault();
    const installEvent = e as BeforeInstallPromptEvent;
    
    setState(prev => ({
      ...prev,
      installPrompt: installEvent,
      canInstall: true,
      isInstallable: true,
    }));

    console.log('[PWA] Install prompt captured');
  }, []);

  // Handle app installed
  const handleAppInstalled = useCallback(() => {
    console.log('[PWA] App was installed');
    localStorage.setItem('pwa-installed', 'true');
    
    setState(prev => ({
      ...prev,
      isInstalled: true,
      canInstall: false,
      installPrompt: null,
    }));

    // Track installation analytics
    if ((window as any).gtag) {
      (window as any).gtag('event', 'pwa_install', {
        event_category: 'PWA',
        event_label: 'App Installed'
      });
    }
  }, []);

  // Handle online/offline status
  const handleOnlineStatus = useCallback(() => {
    setState(prev => ({ ...prev, isOnline: navigator.onLine }));
  }, []);

  // Handle service worker updates
  const handleServiceWorkerUpdate = useCallback(() => {
    setState(prev => ({ ...prev, updateAvailable: true }));
  }, []);

  // Install the PWA
  const installApp = useCallback(async (): Promise<boolean> => {
    if (!state.installPrompt) {
      console.warn('[PWA] No install prompt available');
      return false;
    }

    try {
      await state.installPrompt.prompt();
      const choiceResult = await state.installPrompt.userChoice;
      
      console.log('[PWA] User choice:', choiceResult.outcome);
      
      if (choiceResult.outcome === 'accepted') {
        setState(prev => ({
          ...prev,
          installPrompt: null,
          canInstall: false,
        }));
        return true;
      }
      
      return false;
    } catch (error) {
      console.error('[PWA] Installation failed:', error);
      return false;
    }
  }, [state.installPrompt]);

  // Update the app
  const updateApp = useCallback(async (): Promise<void> => {
    if (!('serviceWorker' in navigator)) {
      return;
    }

    try {
      const registration = await navigator.serviceWorker.getRegistration();
      if (registration?.waiting) {
        // Tell the waiting SW to skip waiting and become the active SW
        registration.waiting.postMessage({ type: 'SKIP_WAITING' });
        
        // Reload the page to apply the update
        window.location.reload();
      }
    } catch (error) {
      console.error('[PWA] Update failed:', error);
    }
  }, []);

  // Dismiss install prompt
  const dismissInstallPrompt = useCallback(() => {
    setState(prev => ({
      ...prev,
      installPrompt: null,
      canInstall: false,
    }));
    
    // Remember dismissal for a day
    const dismissalTime = Date.now() + (24 * 60 * 60 * 1000);
    localStorage.setItem('pwa-install-dismissed', dismissalTime.toString());
  }, []);

  // Check for app updates
  const checkForUpdates = useCallback(async (): Promise<boolean> => {
    if (!('serviceWorker' in navigator)) {
      return false;
    }

    try {
      const registration = await navigator.serviceWorker.getRegistration();
      if (registration) {
        await registration.update();
        return !!registration.waiting;
      }
      return false;
    } catch (error) {
      console.error('[PWA] Update check failed:', error);
      return false;
    }
  }, []);

  // Clear app cache
  const clearCache = useCallback(async (): Promise<void> => {
    if (!('serviceWorker' in navigator)) {
      return;
    }

    try {
      const registration = await navigator.serviceWorker.getRegistration();
      if (registration?.active) {
        const messageChannel = new MessageChannel();
        
        await new Promise<void>((resolve) => {
          messageChannel.port1.onmessage = (event) => {
            if (event.data.success) {
              console.log('[PWA] Cache cleared successfully');
            } else {
              console.error('[PWA] Cache clear failed');
            }
            resolve();
          };
          
          registration.active!.postMessage(
            { type: 'CLEAR_CACHE' },
            [messageChannel.port2]
          );
        });
      }
      
      // Also clear local storage cache
      const cacheKeys = Object.keys(localStorage).filter(key => 
        key.startsWith('cache_') || key.startsWith('offline_')
      );
      
      cacheKeys.forEach(key => localStorage.removeItem(key));
      
    } catch (error) {
      console.error('[PWA] Cache clear failed:', error);
    }
  }, []);

  // Get cache status
  const getCacheStatus = useCallback(async (): Promise<Record<string, number>> => {
    if (!('serviceWorker' in navigator)) {
      return {};
    }

    try {
      const registration = await navigator.serviceWorker.getRegistration();
      if (registration?.active) {
        const messageChannel = new MessageChannel();
        
        return new Promise((resolve) => {
          messageChannel.port1.onmessage = (event) => {
            resolve(event.data || {});
          };
          
          registration.active!.postMessage(
            { type: 'GET_CACHE_STATUS' },
            [messageChannel.port2]
          );
        });
      }
      
      return {};
    } catch (error) {
      console.error('[PWA] Cache status check failed:', error);
      return {};
    }
  }, []);

  // Setup event listeners
  useEffect(() => {
    // Check if install was previously dismissed recently
    const dismissedTime = localStorage.getItem('pwa-install-dismissed');
    const isDismissed = dismissedTime && Date.now() < parseInt(dismissedTime);
    
    if (isDismissed) {
      setState(prev => ({ ...prev, canInstall: false }));
    }

    // Add event listeners
    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
    window.addEventListener('appinstalled', handleAppInstalled);
    window.addEventListener('online', handleOnlineStatus);
    window.addEventListener('offline', handleOnlineStatus);

    // Listen for service worker updates
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.addEventListener('message', (event) => {
        if (event.data?.type === 'UPDATE_AVAILABLE') {
          handleServiceWorkerUpdate();
        }
      });
    }

    // Cleanup
    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
      window.removeEventListener('appinstalled', handleAppInstalled);
      window.removeEventListener('online', handleOnlineStatus);
      window.removeEventListener('offline', handleOnlineStatus);
    };
  }, [handleBeforeInstallPrompt, handleAppInstalled, handleOnlineStatus, handleServiceWorkerUpdate]);

  // Detect display mode changes
  useEffect(() => {
    const mediaQuery = window.matchMedia('(display-mode: standalone)');
    
    const handleDisplayModeChange = (e: MediaQueryListEvent) => {
      setState(prev => ({
        ...prev,
        isStandalone: e.matches,
        isInstalled: e.matches || prev.isInstalled,
      }));
    };

    mediaQuery.addListener(handleDisplayModeChange);
    
    return () => {
      mediaQuery.removeListener(handleDisplayModeChange);
    };
  }, []);

  // Periodic update check
  useEffect(() => {
    const checkInterval = setInterval(async () => {
      if (navigator.onLine) {
        const hasUpdate = await checkForUpdates();
        if (hasUpdate && !state.updateAvailable) {
          handleServiceWorkerUpdate();
        }
      }
    }, 60000); // Check every minute

    return () => clearInterval(checkInterval);
  }, [checkForUpdates, state.updateAvailable, handleServiceWorkerUpdate]);

  return {
    // State
    ...state,
    
    // Actions
    installApp,
    updateApp,
    dismissInstallPrompt,
    checkForUpdates,
    clearCache,
    getCacheStatus,
  };
}