import { useState, useEffect, useCallback } from 'react';

interface SyncStatus {
  isOnline: boolean;
  isSyncing: boolean;
  hasUnsyncedData: boolean;
  lastSyncTime: Date | null;
  pendingOperations: number;
  syncError: string | null;
}

interface SyncOperation {
  id: string;
  type: 'create' | 'update' | 'delete';
  resource: string;
  data: any;
  timestamp: number;
  retryCount: number;
  maxRetries: number;
}

interface OfflineSyncHookReturn extends SyncStatus {
  queueOperation: (operation: Omit<SyncOperation, 'id' | 'timestamp' | 'retryCount'>) => Promise<string>;
  syncNow: () => Promise<boolean>;
  clearPendingOperations: () => Promise<void>;
  getPendingOperations: () => Promise<SyncOperation[]>;
  cancelOperation: (operationId: string) => Promise<boolean>;
}

const DB_NAME = 'OllamaMaxOfflineDB';
const DB_VERSION = 1;
const SYNC_STORE = 'syncOperations';
const DATA_STORE = 'offlineData';

// IndexedDB helper class
class OfflineDB {
  private db: IDBDatabase | null = null;

  async init(): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        this.db = request.result;
        resolve();
      };

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;

        // Create sync operations store
        if (!db.objectStoreNames.contains(SYNC_STORE)) {
          const syncStore = db.createObjectStore(SYNC_STORE, { keyPath: 'id' });
          syncStore.createIndex('timestamp', 'timestamp', { unique: false });
          syncStore.createIndex('resource', 'resource', { unique: false });
        }

        // Create offline data store
        if (!db.objectStoreNames.contains(DATA_STORE)) {
          const dataStore = db.createObjectStore(DATA_STORE, { keyPath: 'key' });
          dataStore.createIndex('resource', 'resource', { unique: false });
          dataStore.createIndex('timestamp', 'timestamp', { unique: false });
        }
      };
    });
  }

  async addOperation(operation: SyncOperation): Promise<void> {
    if (!this.db) throw new Error('Database not initialized');

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([SYNC_STORE], 'readwrite');
      const store = transaction.objectStore(SYNC_STORE);
      const request = store.add(operation);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  async getOperations(): Promise<SyncOperation[]> {
    if (!this.db) throw new Error('Database not initialized');

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([SYNC_STORE], 'readonly');
      const store = transaction.objectStore(SYNC_STORE);
      const index = store.index('timestamp');
      const request = index.getAll();

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve(request.result);
    });
  }

  async removeOperation(id: string): Promise<void> {
    if (!this.db) throw new Error('Database not initialized');

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([SYNC_STORE], 'readwrite');
      const store = transaction.objectStore(SYNC_STORE);
      const request = store.delete(id);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  async updateOperation(operation: SyncOperation): Promise<void> {
    if (!this.db) throw new Error('Database not initialized');

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([SYNC_STORE], 'readwrite');
      const store = transaction.objectStore(SYNC_STORE);
      const request = store.put(operation);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  async clear(): Promise<void> {
    if (!this.db) throw new Error('Database not initialized');

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([SYNC_STORE], 'readwrite');
      const store = transaction.objectStore(SYNC_STORE);
      const request = store.clear();

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  async storeData(key: string, resource: string, data: any): Promise<void> {
    if (!this.db) throw new Error('Database not initialized');

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([DATA_STORE], 'readwrite');
      const store = transaction.objectStore(DATA_STORE);
      const request = store.put({
        key,
        resource,
        data,
        timestamp: Date.now()
      });

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  async getData(key: string): Promise<any> {
    if (!this.db) throw new Error('Database not initialized');

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([DATA_STORE], 'readonly');
      const store = transaction.objectStore(DATA_STORE);
      const request = store.get(key);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve(request.result?.data);
    });
  }
}

export function useOfflineSync(): OfflineSyncHookReturn {
  const [status, setStatus] = useState<SyncStatus>({
    isOnline: navigator.onLine,
    isSyncing: false,
    hasUnsyncedData: false,
    lastSyncTime: null,
    pendingOperations: 0,
    syncError: null,
  });

  const [db] = useState(() => new OfflineDB());

  // Initialize database
  useEffect(() => {
    const initDB = async () => {
      try {
        await db.init();
        await updatePendingCount();
        
        // Load last sync time from localStorage
        const lastSync = localStorage.getItem('lastSyncTime');
        if (lastSync) {
          setStatus(prev => ({
            ...prev,
            lastSyncTime: new Date(parseInt(lastSync))
          }));
        }
      } catch (error) {
        console.error('[OfflineSync] Failed to initialize database:', error);
        setStatus(prev => ({
          ...prev,
          syncError: 'Failed to initialize offline database'
        }));
      }
    };

    initDB();
  }, [db]);

  // Listen for online/offline events
  useEffect(() => {
    const handleOnline = () => {
      setStatus(prev => ({ ...prev, isOnline: true, syncError: null }));
      // Auto-sync when coming online
      if (status.hasUnsyncedData) {
        syncNow();
      }
    };

    const handleOffline = () => {
      setStatus(prev => ({ ...prev, isOnline: false }));
    };

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, [status.hasUnsyncedData]);

  // Update pending operations count
  const updatePendingCount = async () => {
    try {
      const operations = await db.getOperations();
      setStatus(prev => ({
        ...prev,
        pendingOperations: operations.length,
        hasUnsyncedData: operations.length > 0
      }));
    } catch (error) {
      console.error('[OfflineSync] Failed to update pending count:', error);
    }
  };

  // Queue an operation for sync
  const queueOperation = useCallback(async (
    operation: Omit<SyncOperation, 'id' | 'timestamp' | 'retryCount'>
  ): Promise<string> => {
    const syncOperation: SyncOperation = {
      ...operation,
      id: generateOperationId(),
      timestamp: Date.now(),
      retryCount: 0,
      maxRetries: operation.maxRetries || 3
    };

    try {
      await db.addOperation(syncOperation);
      await updatePendingCount();

      console.log('[OfflineSync] Operation queued:', syncOperation);

      // If online, attempt immediate sync
      if (status.isOnline) {
        syncNow();
      }

      return syncOperation.id;
    } catch (error) {
      console.error('[OfflineSync] Failed to queue operation:', error);
      throw error;
    }
  }, [status.isOnline, db]);

  // Execute API request
  const executeOperation = async (operation: SyncOperation): Promise<boolean> => {
    try {
      const endpoint = `/api/${operation.resource}`;
      const config: RequestInit = {
        method: operation.type === 'create' ? 'POST' : 
                operation.type === 'update' ? 'PUT' : 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        }
      };

      if (operation.type !== 'delete') {
        config.body = JSON.stringify(operation.data);
      } else if (operation.data?.id) {
        config.method = 'DELETE';
        // For delete operations, append ID to URL
        const deleteEndpoint = `${endpoint}/${operation.data.id}`;
        const response = await fetch(deleteEndpoint, config);
        return response.ok;
      }

      const response = await fetch(endpoint, config);
      
      if (response.ok) {
        console.log('[OfflineSync] Operation executed successfully:', operation.id);
        return true;
      } else {
        console.error('[OfflineSync] Operation failed:', response.status, response.statusText);
        return false;
      }
    } catch (error) {
      console.error('[OfflineSync] Network error executing operation:', error);
      return false;
    }
  };

  // Sync all pending operations
  const syncNow = useCallback(async (): Promise<boolean> => {
    if (!status.isOnline) {
      return false;
    }

    setStatus(prev => ({ ...prev, isSyncing: true, syncError: null }));

    try {
      const operations = await db.getOperations();
      
      if (operations.length === 0) {
        setStatus(prev => ({
          ...prev,
          isSyncing: false,
          lastSyncTime: new Date()
        }));
        localStorage.setItem('lastSyncTime', Date.now().toString());
        return true;
      }

      console.log('[OfflineSync] Starting sync of', operations.length, 'operations');

      let successCount = 0;
      let failureCount = 0;

      // Process operations in order
      for (const operation of operations) {
        const success = await executeOperation(operation);

        if (success) {
          await db.removeOperation(operation.id);
          successCount++;
        } else {
          // Increment retry count
          operation.retryCount++;
          
          if (operation.retryCount >= operation.maxRetries) {
            console.error('[OfflineSync] Operation exceeded max retries:', operation.id);
            await db.removeOperation(operation.id);
            failureCount++;
          } else {
            await db.updateOperation(operation);
          }
        }
      }

      await updatePendingCount();

      const allSuccessful = failureCount === 0;
      const now = new Date();

      setStatus(prev => ({
        ...prev,
        isSyncing: false,
        lastSyncTime: now,
        syncError: allSuccessful ? null : `${failureCount} operations failed to sync`
      }));

      localStorage.setItem('lastSyncTime', now.getTime().toString());

      console.log('[OfflineSync] Sync completed:', {
        successful: successCount,
        failed: failureCount,
        total: operations.length
      });

      return allSuccessful;
    } catch (error) {
      console.error('[OfflineSync] Sync failed:', error);
      setStatus(prev => ({
        ...prev,
        isSyncing: false,
        syncError: 'Sync operation failed'
      }));
      return false;
    }
  }, [status.isOnline, db]);

  // Clear all pending operations
  const clearPendingOperations = useCallback(async (): Promise<void> => {
    try {
      await db.clear();
      await updatePendingCount();
      console.log('[OfflineSync] All pending operations cleared');
    } catch (error) {
      console.error('[OfflineSync] Failed to clear pending operations:', error);
      throw error;
    }
  }, [db]);

  // Get all pending operations
  const getPendingOperations = useCallback(async (): Promise<SyncOperation[]> => {
    try {
      return await db.getOperations();
    } catch (error) {
      console.error('[OfflineSync] Failed to get pending operations:', error);
      return [];
    }
  }, [db]);

  // Cancel a specific operation
  const cancelOperation = useCallback(async (operationId: string): Promise<boolean> => {
    try {
      await db.removeOperation(operationId);
      await updatePendingCount();
      console.log('[OfflineSync] Operation cancelled:', operationId);
      return true;
    } catch (error) {
      console.error('[OfflineSync] Failed to cancel operation:', error);
      return false;
    }
  }, [db]);

  // Auto-sync interval when online
  useEffect(() => {
    if (!status.isOnline) return;

    const interval = setInterval(() => {
      if (status.hasUnsyncedData && !status.isSyncing) {
        syncNow();
      }
    }, 30000); // Sync every 30 seconds

    return () => clearInterval(interval);
  }, [status.isOnline, status.hasUnsyncedData, status.isSyncing, syncNow]);

  return {
    // Status
    ...status,
    
    // Actions
    queueOperation,
    syncNow,
    clearPendingOperations,
    getPendingOperations,
    cancelOperation,
  };
}

// Helper function to generate unique operation IDs
function generateOperationId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}