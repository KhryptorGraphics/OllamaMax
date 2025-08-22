import { useState, useEffect, useCallback } from 'react';
import { useWebSocket } from '../useWebSocket';
import {
  FederationCluster,
  FederationPolicy,
  FederationEvent,
  CrossRegionReplication,
  ServiceDiscovery,
  FederationConfiguration
} from '../../types/federation';

interface FederationState {
  clusters: FederationCluster[];
  policies: FederationPolicy[];
  events: FederationEvent[];
  replication: CrossRegionReplication[];
  discovery: ServiceDiscovery | null;
  configuration: FederationConfiguration | null;
  loading: boolean;
  error: string | null;
  connected: boolean;
}

interface FederationActions {
  addCluster: (cluster: Omit<FederationCluster, 'id' | 'lastSeen'>) => Promise<void>;
  removeCluster: (clusterId: string) => Promise<void>;
  updateCluster: (clusterId: string, updates: Partial<FederationCluster>) => Promise<void>;
  createPolicy: (policy: Omit<FederationPolicy, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  updatePolicy: (policyId: string, updates: Partial<FederationPolicy>) => Promise<void>;
  deletePolicy: (policyId: string) => Promise<void>;
  enablePolicy: (policyId: string) => Promise<void>;
  disablePolicy: (policyId: string) => Promise<void>;
  initiateFailover: (sourceCluster: string, targetCluster: string) => Promise<void>;
  configureReplication: (config: Omit<CrossRegionReplication, 'id'>) => Promise<void>;
  refreshDiscovery: () => Promise<void>;
  exportConfiguration: () => Promise<string>;
  importConfiguration: (config: string) => Promise<void>;
}

export function useFederation(): FederationState & FederationActions {
  const [state, setState] = useState<FederationState>({
    clusters: [],
    policies: [],
    events: [],
    replication: [],
    discovery: null,
    configuration: null,
    loading: true,
    error: null,
    connected: false
  });

  const { isConnected, sendMessage, lastMessage } = useWebSocket('/api/ws/federation');

  // Handle WebSocket messages
  useEffect(() => {
    if (lastMessage) {
      try {
        const message = JSON.parse(lastMessage.data);
        handleWebSocketMessage(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    }
  }, [lastMessage]);

  // Update connection status
  useEffect(() => {
    setState(prev => ({ ...prev, connected: isConnected }));
  }, [isConnected]);

  const handleWebSocketMessage = useCallback((message: any) => {
    switch (message.type) {
      case 'cluster-update':
        setState(prev => ({
          ...prev,
          clusters: prev.clusters.map(cluster =>
            cluster.id === message.data.id ? { ...cluster, ...message.data } : cluster
          )
        }));
        break;

      case 'cluster-added':
        setState(prev => ({
          ...prev,
          clusters: [...prev.clusters, message.data]
        }));
        break;

      case 'cluster-removed':
        setState(prev => ({
          ...prev,
          clusters: prev.clusters.filter(cluster => cluster.id !== message.data.id)
        }));
        break;

      case 'policy-updated':
        setState(prev => ({
          ...prev,
          policies: prev.policies.map(policy =>
            policy.id === message.data.id ? { ...policy, ...message.data } : policy
          )
        }));
        break;

      case 'federation-event':
        setState(prev => ({
          ...prev,
          events: [message.data, ...prev.events].slice(0, 1000) // Keep last 1000 events
        }));
        break;

      case 'replication-status':
        setState(prev => ({
          ...prev,
          replication: prev.replication.map(rep =>
            rep.id === message.data.id ? { ...rep, ...message.data } : rep
          )
        }));
        break;

      case 'discovery-update':
        setState(prev => ({
          ...prev,
          discovery: message.data
        }));
        break;

      case 'error':
        setState(prev => ({
          ...prev,
          error: message.data.message
        }));
        break;
    }
  }, []);

  // Initialize federation data
  useEffect(() => {
    const initializeFederation = async () => {
      try {
        setState(prev => ({ ...prev, loading: true, error: null }));

        const [
          clustersResponse,
          policiesResponse,
          eventsResponse,
          replicationResponse,
          discoveryResponse,
          configResponse
        ] = await Promise.all([
          fetch('/api/federation/clusters'),
          fetch('/api/federation/policies'),
          fetch('/api/federation/events?limit=100'),
          fetch('/api/federation/replication'),
          fetch('/api/federation/discovery'),
          fetch('/api/federation/configuration')
        ]);

        const [clusters, policies, events, replication, discovery, configuration] = await Promise.all([
          clustersResponse.json(),
          policiesResponse.json(),
          eventsResponse.json(),
          replicationResponse.json(),
          discoveryResponse.json(),
          configResponse.json()
        ]);

        setState(prev => ({
          ...prev,
          clusters,
          policies,
          events,
          replication,
          discovery,
          configuration,
          loading: false
        }));
      } catch (error) {
        setState(prev => ({
          ...prev,
          error: error instanceof Error ? error.message : 'Failed to load federation data',
          loading: false
        }));
      }
    };

    initializeFederation();
  }, []);

  const addCluster = useCallback(async (cluster: Omit<FederationCluster, 'id' | 'lastSeen'>) => {
    try {
      const response = await fetch('/api/federation/clusters', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(cluster)
      });

      if (!response.ok) {
        throw new Error('Failed to add cluster');
      }

      const newCluster = await response.json();
      setState(prev => ({
        ...prev,
        clusters: [...prev.clusters, newCluster]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to add cluster'
      }));
      throw error;
    }
  }, []);

  const removeCluster = useCallback(async (clusterId: string) => {
    try {
      const response = await fetch(`/api/federation/clusters/${clusterId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to remove cluster');
      }

      setState(prev => ({
        ...prev,
        clusters: prev.clusters.filter(cluster => cluster.id !== clusterId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to remove cluster'
      }));
      throw error;
    }
  }, []);

  const updateCluster = useCallback(async (clusterId: string, updates: Partial<FederationCluster>) => {
    try {
      const response = await fetch(`/api/federation/clusters/${clusterId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        throw new Error('Failed to update cluster');
      }

      const updatedCluster = await response.json();
      setState(prev => ({
        ...prev,
        clusters: prev.clusters.map(cluster =>
          cluster.id === clusterId ? updatedCluster : cluster
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update cluster'
      }));
      throw error;
    }
  }, []);

  const createPolicy = useCallback(async (policy: Omit<FederationPolicy, 'id' | 'createdAt' | 'updatedAt'>) => {
    try {
      const response = await fetch('/api/federation/policies', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(policy)
      });

      if (!response.ok) {
        throw new Error('Failed to create policy');
      }

      const newPolicy = await response.json();
      setState(prev => ({
        ...prev,
        policies: [...prev.policies, newPolicy]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to create policy'
      }));
      throw error;
    }
  }, []);

  const updatePolicy = useCallback(async (policyId: string, updates: Partial<FederationPolicy>) => {
    try {
      const response = await fetch(`/api/federation/policies/${policyId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        throw new Error('Failed to update policy');
      }

      const updatedPolicy = await response.json();
      setState(prev => ({
        ...prev,
        policies: prev.policies.map(policy =>
          policy.id === policyId ? updatedPolicy : policy
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update policy'
      }));
      throw error;
    }
  }, []);

  const deletePolicy = useCallback(async (policyId: string) => {
    try {
      const response = await fetch(`/api/federation/policies/${policyId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to delete policy');
      }

      setState(prev => ({
        ...prev,
        policies: prev.policies.filter(policy => policy.id !== policyId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to delete policy'
      }));
      throw error;
    }
  }, []);

  const enablePolicy = useCallback(async (policyId: string) => {
    await updatePolicy(policyId, { enabled: true });
  }, [updatePolicy]);

  const disablePolicy = useCallback(async (policyId: string) => {
    await updatePolicy(policyId, { enabled: false });
  }, [updatePolicy]);

  const initiateFailover = useCallback(async (sourceCluster: string, targetCluster: string) => {
    try {
      const response = await fetch('/api/federation/failover', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sourceCluster, targetCluster })
      });

      if (!response.ok) {
        throw new Error('Failed to initiate failover');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to initiate failover'
      }));
      throw error;
    }
  }, []);

  const configureReplication = useCallback(async (config: Omit<CrossRegionReplication, 'id'>) => {
    try {
      const response = await fetch('/api/federation/replication', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        throw new Error('Failed to configure replication');
      }

      const newReplication = await response.json();
      setState(prev => ({
        ...prev,
        replication: [...prev.replication, newReplication]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to configure replication'
      }));
      throw error;
    }
  }, []);

  const refreshDiscovery = useCallback(async () => {
    try {
      const response = await fetch('/api/federation/discovery/refresh', {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to refresh service discovery');
      }

      const discovery = await response.json();
      setState(prev => ({
        ...prev,
        discovery
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to refresh service discovery'
      }));
      throw error;
    }
  }, []);

  const exportConfiguration = useCallback(async (): Promise<string> => {
    try {
      const response = await fetch('/api/federation/configuration/export');
      
      if (!response.ok) {
        throw new Error('Failed to export configuration');
      }

      return await response.text();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to export configuration'
      }));
      throw error;
    }
  }, []);

  const importConfiguration = useCallback(async (config: string) => {
    try {
      const response = await fetch('/api/federation/configuration/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ configuration: config })
      });

      if (!response.ok) {
        throw new Error('Failed to import configuration');
      }

      // Reload federation data after import
      window.location.reload();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to import configuration'
      }));
      throw error;
    }
  }, []);

  return {
    ...state,
    addCluster,
    removeCluster,
    updateCluster,
    createPolicy,
    updatePolicy,
    deletePolicy,
    enablePolicy,
    disablePolicy,
    initiateFailover,
    configureReplication,
    refreshDiscovery,
    exportConfiguration,
    importConfiguration
  };
}