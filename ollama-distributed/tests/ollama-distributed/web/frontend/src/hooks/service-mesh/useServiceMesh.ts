import { useState, useEffect, useCallback } from 'react';
import { useWebSocket } from '../useWebSocket';
import {
  ServiceMeshConfiguration,
  ServiceMeshService,
  ServiceMeshWorkload,
  ServiceTopology,
  CanaryDeployment,
  SecurityPolicy,
  TrafficConfig,
  ObservabilityConfig
} from '../../types/service-mesh';

interface ServiceMeshState {
  configuration: ServiceMeshConfiguration | null;
  services: ServiceMeshService[];
  workloads: ServiceMeshWorkload[];
  topology: ServiceTopology | null;
  canaryDeployments: CanaryDeployment[];
  securityPolicies: SecurityPolicy[];
  trafficConfig: TrafficConfig | null;
  observabilityConfig: ObservabilityConfig | null;
  loading: boolean;
  error: string | null;
  connected: boolean;
}

interface ServiceMeshActions {
  // Configuration Management
  updateConfiguration: (config: Partial<ServiceMeshConfiguration>) => Promise<void>;
  restartControlPlane: () => Promise<void>;
  updateDataPlane: (config: Partial<any>) => Promise<void>;
  
  // Service Management
  refreshServices: () => Promise<void>;
  getServiceDetails: (serviceId: string) => Promise<ServiceMeshService>;
  updateServiceLabels: (serviceId: string, labels: Record<string, string>) => Promise<void>;
  
  // Workload Management
  refreshWorkloads: () => Promise<void>;
  getWorkloadDetails: (workloadId: string) => Promise<ServiceMeshWorkload>;
  injectSidecar: (workloadId: string) => Promise<void>;
  removeSidecar: (workloadId: string) => Promise<void>;
  restartWorkload: (workloadId: string) => Promise<void>;
  
  // Topology Management
  refreshTopology: () => Promise<void>;
  exportTopology: (format: 'json' | 'yaml' | 'dot') => Promise<string>;
  
  // Security Policy Management
  createSecurityPolicy: (policy: Omit<SecurityPolicy, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  updateSecurityPolicy: (policyId: string, updates: Partial<SecurityPolicy>) => Promise<void>;
  deleteSecurityPolicy: (policyId: string) => Promise<void>;
  enableSecurityPolicy: (policyId: string) => Promise<void>;
  disableSecurityPolicy: (policyId: string) => Promise<void>;
  testSecurityPolicy: (policy: SecurityPolicy, scenario: any) => Promise<{ allowed: boolean; reason: string }>;
  
  // Traffic Management
  updateTrafficConfig: (config: Partial<TrafficConfig>) => Promise<void>;
  createTrafficRule: (rule: any) => Promise<void>;
  updateTrafficRule: (ruleId: string, updates: any) => Promise<void>;
  deleteTrafficRule: (ruleId: string) => Promise<void>;
  
  // Canary Deployments
  createCanaryDeployment: (deployment: Omit<CanaryDeployment, 'id' | 'startTime'>) => Promise<void>;
  promoteCanary: (deploymentId: string) => Promise<void>;
  rollbackCanary: (deploymentId: string) => Promise<void>;
  pauseCanary: (deploymentId: string) => Promise<void>;
  resumeCanary: (deploymentId: string) => Promise<void>;
  abortCanary: (deploymentId: string) => Promise<void>;
  
  // Observability
  updateObservabilityConfig: (config: Partial<ObservabilityConfig>) => Promise<void>;
  getMetrics: (query: string, timeRange: string) => Promise<any>;
  getTraces: (serviceId: string, timeRange: string) => Promise<any>;
  getLogs: (workloadId: string, timeRange: string) => Promise<any>;
  
  // Certificate Management
  rotateCertificates: () => Promise<void>;
  getCertificateStatus: () => Promise<any>;
  
  // Health and Diagnostics
  runHealthCheck: () => Promise<any>;
  runConfigValidation: () => Promise<{ valid: boolean; errors: string[]; warnings: string[] }>;
  getDiagnostics: () => Promise<any>;
  
  // Import/Export
  exportConfiguration: () => Promise<string>;
  importConfiguration: (config: string) => Promise<void>;
}

export function useServiceMesh(): ServiceMeshState & ServiceMeshActions {
  const [state, setState] = useState<ServiceMeshState>({
    configuration: null,
    services: [],
    workloads: [],
    topology: null,
    canaryDeployments: [],
    securityPolicies: [],
    trafficConfig: null,
    observabilityConfig: null,
    loading: true,
    error: null,
    connected: false
  });

  const { isConnected, sendMessage, lastMessage } = useWebSocket('/api/ws/service-mesh');

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
      case 'configuration-updated':
        setState(prev => ({
          ...prev,
          configuration: { ...prev.configuration, ...message.data }
        }));
        break;

      case 'service-updated':
        setState(prev => ({
          ...prev,
          services: prev.services.map(service =>
            service.id === message.data.id ? { ...service, ...message.data } : service
          )
        }));
        break;

      case 'workload-updated':
        setState(prev => ({
          ...prev,
          workloads: prev.workloads.map(workload =>
            workload.id === message.data.id ? { ...workload, ...message.data } : workload
          )
        }));
        break;

      case 'topology-updated':
        setState(prev => ({
          ...prev,
          topology: message.data
        }));
        break;

      case 'canary-status-changed':
        setState(prev => ({
          ...prev,
          canaryDeployments: prev.canaryDeployments.map(canary =>
            canary.id === message.data.id ? { ...canary, ...message.data } : canary
          )
        }));
        break;

      case 'security-policy-updated':
        setState(prev => ({
          ...prev,
          securityPolicies: prev.securityPolicies.map(policy =>
            policy.id === message.data.id ? { ...policy, ...message.data } : policy
          )
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

  // Initialize service mesh data
  useEffect(() => {
    const initializeServiceMesh = async () => {
      try {
        setState(prev => ({ ...prev, loading: true, error: null }));

        const [
          configResponse,
          servicesResponse,
          workloadsResponse,
          topologyResponse,
          canaryResponse,
          policiesResponse,
          trafficResponse,
          observabilityResponse
        ] = await Promise.all([
          fetch('/api/service-mesh/configuration'),
          fetch('/api/service-mesh/services'),
          fetch('/api/service-mesh/workloads'),
          fetch('/api/service-mesh/topology'),
          fetch('/api/service-mesh/canary-deployments'),
          fetch('/api/service-mesh/security-policies'),
          fetch('/api/service-mesh/traffic-config'),
          fetch('/api/service-mesh/observability-config')
        ]);

        const [
          configuration,
          services,
          workloads,
          topology,
          canaryDeployments,
          securityPolicies,
          trafficConfig,
          observabilityConfig
        ] = await Promise.all([
          configResponse.json(),
          servicesResponse.json(),
          workloadsResponse.json(),
          topologyResponse.json(),
          canaryResponse.json(),
          policiesResponse.json(),
          trafficResponse.json(),
          observabilityResponse.json()
        ]);

        setState(prev => ({
          ...prev,
          configuration,
          services,
          workloads,
          topology,
          canaryDeployments,
          securityPolicies,
          trafficConfig,
          observabilityConfig,
          loading: false
        }));
      } catch (error) {
        setState(prev => ({
          ...prev,
          error: error instanceof Error ? error.message : 'Failed to load service mesh data',
          loading: false
        }));
      }
    };

    initializeServiceMesh();
  }, []);

  // Configuration Management
  const updateConfiguration = useCallback(async (config: Partial<ServiceMeshConfiguration>) => {
    try {
      const response = await fetch('/api/service-mesh/configuration', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        throw new Error('Failed to update configuration');
      }

      const updatedConfig = await response.json();
      setState(prev => ({
        ...prev,
        configuration: updatedConfig
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update configuration'
      }));
      throw error;
    }
  }, []);

  const restartControlPlane = useCallback(async () => {
    try {
      const response = await fetch('/api/service-mesh/control-plane/restart', {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to restart control plane');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to restart control plane'
      }));
      throw error;
    }
  }, []);

  const updateDataPlane = useCallback(async (config: Partial<any>) => {
    try {
      const response = await fetch('/api/service-mesh/data-plane', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        throw new Error('Failed to update data plane');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update data plane'
      }));
      throw error;
    }
  }, []);

  // Service Management
  const refreshServices = useCallback(async () => {
    try {
      const response = await fetch('/api/service-mesh/services');
      if (!response.ok) {
        throw new Error('Failed to refresh services');
      }

      const services = await response.json();
      setState(prev => ({ ...prev, services }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to refresh services'
      }));
      throw error;
    }
  }, []);

  const getServiceDetails = useCallback(async (serviceId: string): Promise<ServiceMeshService> => {
    try {
      const response = await fetch(`/api/service-mesh/services/${serviceId}`);
      if (!response.ok) {
        throw new Error('Failed to get service details');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to get service details'
      }));
      throw error;
    }
  }, []);

  const updateServiceLabels = useCallback(async (serviceId: string, labels: Record<string, string>) => {
    try {
      const response = await fetch(`/api/service-mesh/services/${serviceId}/labels`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ labels })
      });

      if (!response.ok) {
        throw new Error('Failed to update service labels');
      }

      const updatedService = await response.json();
      setState(prev => ({
        ...prev,
        services: prev.services.map(service =>
          service.id === serviceId ? updatedService : service
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update service labels'
      }));
      throw error;
    }
  }, []);

  // Workload Management
  const refreshWorkloads = useCallback(async () => {
    try {
      const response = await fetch('/api/service-mesh/workloads');
      if (!response.ok) {
        throw new Error('Failed to refresh workloads');
      }

      const workloads = await response.json();
      setState(prev => ({ ...prev, workloads }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to refresh workloads'
      }));
      throw error;
    }
  }, []);

  const getWorkloadDetails = useCallback(async (workloadId: string): Promise<ServiceMeshWorkload> => {
    try {
      const response = await fetch(`/api/service-mesh/workloads/${workloadId}`);
      if (!response.ok) {
        throw new Error('Failed to get workload details');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to get workload details'
      }));
      throw error;
    }
  }, []);

  const injectSidecar = useCallback(async (workloadId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/workloads/${workloadId}/sidecar/inject`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to inject sidecar');
      }

      // Refresh workloads to get updated state
      await refreshWorkloads();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to inject sidecar'
      }));
      throw error;
    }
  }, [refreshWorkloads]);

  const removeSidecar = useCallback(async (workloadId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/workloads/${workloadId}/sidecar/remove`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to remove sidecar');
      }

      // Refresh workloads to get updated state
      await refreshWorkloads();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to remove sidecar'
      }));
      throw error;
    }
  }, [refreshWorkloads]);

  const restartWorkload = useCallback(async (workloadId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/workloads/${workloadId}/restart`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to restart workload');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to restart workload'
      }));
      throw error;
    }
  }, []);

  // Topology Management
  const refreshTopology = useCallback(async () => {
    try {
      const response = await fetch('/api/service-mesh/topology');
      if (!response.ok) {
        throw new Error('Failed to refresh topology');
      }

      const topology = await response.json();
      setState(prev => ({ ...prev, topology }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to refresh topology'
      }));
      throw error;
    }
  }, []);

  const exportTopology = useCallback(async (format: 'json' | 'yaml' | 'dot'): Promise<string> => {
    try {
      const response = await fetch(`/api/service-mesh/topology/export?format=${format}`);
      if (!response.ok) {
        throw new Error('Failed to export topology');
      }

      return await response.text();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to export topology'
      }));
      throw error;
    }
  }, []);

  // Security Policy Management
  const createSecurityPolicy = useCallback(async (policy: Omit<SecurityPolicy, 'id' | 'createdAt' | 'updatedAt'>) => {
    try {
      const response = await fetch('/api/service-mesh/security-policies', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(policy)
      });

      if (!response.ok) {
        throw new Error('Failed to create security policy');
      }

      const newPolicy = await response.json();
      setState(prev => ({
        ...prev,
        securityPolicies: [...prev.securityPolicies, newPolicy]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to create security policy'
      }));
      throw error;
    }
  }, []);

  const updateSecurityPolicy = useCallback(async (policyId: string, updates: Partial<SecurityPolicy>) => {
    try {
      const response = await fetch(`/api/service-mesh/security-policies/${policyId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        throw new Error('Failed to update security policy');
      }

      const updatedPolicy = await response.json();
      setState(prev => ({
        ...prev,
        securityPolicies: prev.securityPolicies.map(policy =>
          policy.id === policyId ? updatedPolicy : policy
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update security policy'
      }));
      throw error;
    }
  }, []);

  const deleteSecurityPolicy = useCallback(async (policyId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/security-policies/${policyId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to delete security policy');
      }

      setState(prev => ({
        ...prev,
        securityPolicies: prev.securityPolicies.filter(policy => policy.id !== policyId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to delete security policy'
      }));
      throw error;
    }
  }, []);

  const enableSecurityPolicy = useCallback(async (policyId: string) => {
    await updateSecurityPolicy(policyId, { enabled: true });
  }, [updateSecurityPolicy]);

  const disableSecurityPolicy = useCallback(async (policyId: string) => {
    await updateSecurityPolicy(policyId, { enabled: false });
  }, [updateSecurityPolicy]);

  const testSecurityPolicy = useCallback(async (policy: SecurityPolicy, scenario: any): Promise<{ allowed: boolean; reason: string }> => {
    try {
      const response = await fetch('/api/service-mesh/security-policies/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ policy, scenario })
      });

      if (!response.ok) {
        throw new Error('Failed to test security policy');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to test security policy'
      }));
      throw error;
    }
  }, []);

  // Traffic Management
  const updateTrafficConfig = useCallback(async (config: Partial<TrafficConfig>) => {
    try {
      const response = await fetch('/api/service-mesh/traffic-config', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        throw new Error('Failed to update traffic config');
      }

      const updatedConfig = await response.json();
      setState(prev => ({
        ...prev,
        trafficConfig: updatedConfig
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update traffic config'
      }));
      throw error;
    }
  }, []);

  const createTrafficRule = useCallback(async (rule: any) => {
    try {
      const response = await fetch('/api/service-mesh/traffic-rules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(rule)
      });

      if (!response.ok) {
        throw new Error('Failed to create traffic rule');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to create traffic rule'
      }));
      throw error;
    }
  }, []);

  const updateTrafficRule = useCallback(async (ruleId: string, updates: any) => {
    try {
      const response = await fetch(`/api/service-mesh/traffic-rules/${ruleId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        throw new Error('Failed to update traffic rule');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update traffic rule'
      }));
      throw error;
    }
  }, []);

  const deleteTrafficRule = useCallback(async (ruleId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/traffic-rules/${ruleId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to delete traffic rule');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to delete traffic rule'
      }));
      throw error;
    }
  }, []);

  // Canary Deployments
  const createCanaryDeployment = useCallback(async (deployment: Omit<CanaryDeployment, 'id' | 'startTime'>) => {
    try {
      const response = await fetch('/api/service-mesh/canary-deployments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(deployment)
      });

      if (!response.ok) {
        throw new Error('Failed to create canary deployment');
      }

      const newDeployment = await response.json();
      setState(prev => ({
        ...prev,
        canaryDeployments: [...prev.canaryDeployments, newDeployment]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to create canary deployment'
      }));
      throw error;
    }
  }, []);

  const promoteCanary = useCallback(async (deploymentId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/canary-deployments/${deploymentId}/promote`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to promote canary');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to promote canary'
      }));
      throw error;
    }
  }, []);

  const rollbackCanary = useCallback(async (deploymentId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/canary-deployments/${deploymentId}/rollback`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to rollback canary');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to rollback canary'
      }));
      throw error;
    }
  }, []);

  const pauseCanary = useCallback(async (deploymentId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/canary-deployments/${deploymentId}/pause`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to pause canary');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to pause canary'
      }));
      throw error;
    }
  }, []);

  const resumeCanary = useCallback(async (deploymentId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/canary-deployments/${deploymentId}/resume`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to resume canary');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to resume canary'
      }));
      throw error;
    }
  }, []);

  const abortCanary = useCallback(async (deploymentId: string) => {
    try {
      const response = await fetch(`/api/service-mesh/canary-deployments/${deploymentId}/abort`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to abort canary');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to abort canary'
      }));
      throw error;
    }
  }, []);

  // Observability
  const updateObservabilityConfig = useCallback(async (config: Partial<ObservabilityConfig>) => {
    try {
      const response = await fetch('/api/service-mesh/observability-config', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        throw new Error('Failed to update observability config');
      }

      const updatedConfig = await response.json();
      setState(prev => ({
        ...prev,
        observabilityConfig: updatedConfig
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update observability config'
      }));
      throw error;
    }
  }, []);

  const getMetrics = useCallback(async (query: string, timeRange: string): Promise<any> => {
    try {
      const response = await fetch('/api/service-mesh/metrics', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query, timeRange })
      });

      if (!response.ok) {
        throw new Error('Failed to get metrics');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to get metrics'
      }));
      throw error;
    }
  }, []);

  const getTraces = useCallback(async (serviceId: string, timeRange: string): Promise<any> => {
    try {
      const response = await fetch(`/api/service-mesh/traces/${serviceId}?timeRange=${timeRange}`);

      if (!response.ok) {
        throw new Error('Failed to get traces');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to get traces'
      }));
      throw error;
    }
  }, []);

  const getLogs = useCallback(async (workloadId: string, timeRange: string): Promise<any> => {
    try {
      const response = await fetch(`/api/service-mesh/logs/${workloadId}?timeRange=${timeRange}`);

      if (!response.ok) {
        throw new Error('Failed to get logs');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to get logs'
      }));
      throw error;
    }
  }, []);

  // Certificate Management
  const rotateCertificates = useCallback(async () => {
    try {
      const response = await fetch('/api/service-mesh/certificates/rotate', {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to rotate certificates');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to rotate certificates'
      }));
      throw error;
    }
  }, []);

  const getCertificateStatus = useCallback(async (): Promise<any> => {
    try {
      const response = await fetch('/api/service-mesh/certificates/status');

      if (!response.ok) {
        throw new Error('Failed to get certificate status');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to get certificate status'
      }));
      throw error;
    }
  }, []);

  // Health and Diagnostics
  const runHealthCheck = useCallback(async (): Promise<any> => {
    try {
      const response = await fetch('/api/service-mesh/health', {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to run health check');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to run health check'
      }));
      throw error;
    }
  }, []);

  const runConfigValidation = useCallback(async (): Promise<{ valid: boolean; errors: string[]; warnings: string[] }> => {
    try {
      const response = await fetch('/api/service-mesh/validate-config', {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to validate configuration');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to validate configuration'
      }));
      throw error;
    }
  }, []);

  const getDiagnostics = useCallback(async (): Promise<any> => {
    try {
      const response = await fetch('/api/service-mesh/diagnostics');

      if (!response.ok) {
        throw new Error('Failed to get diagnostics');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to get diagnostics'
      }));
      throw error;
    }
  }, []);

  // Import/Export
  const exportConfiguration = useCallback(async (): Promise<string> => {
    try {
      const response = await fetch('/api/service-mesh/configuration/export');

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
      const response = await fetch('/api/service-mesh/configuration/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ configuration: config })
      });

      if (!response.ok) {
        throw new Error('Failed to import configuration');
      }

      // Reload service mesh data after import
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
    updateConfiguration,
    restartControlPlane,
    updateDataPlane,
    refreshServices,
    getServiceDetails,
    updateServiceLabels,
    refreshWorkloads,
    getWorkloadDetails,
    injectSidecar,
    removeSidecar,
    restartWorkload,
    refreshTopology,
    exportTopology,
    createSecurityPolicy,
    updateSecurityPolicy,
    deleteSecurityPolicy,
    enableSecurityPolicy,
    disableSecurityPolicy,
    testSecurityPolicy,
    updateTrafficConfig,
    createTrafficRule,
    updateTrafficRule,
    deleteTrafficRule,
    createCanaryDeployment,
    promoteCanary,
    rollbackCanary,
    pauseCanary,
    resumeCanary,
    abortCanary,
    updateObservabilityConfig,
    getMetrics,
    getTraces,
    getLogs,
    rotateCertificates,
    getCertificateStatus,
    runHealthCheck,
    runConfigValidation,
    getDiagnostics,
    exportConfiguration,
    importConfiguration
  };
}