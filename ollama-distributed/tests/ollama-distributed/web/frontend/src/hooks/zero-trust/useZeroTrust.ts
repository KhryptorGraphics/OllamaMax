import { useState, useEffect, useCallback } from 'react';
import { useWebSocket } from '../useWebSocket';
import {
  Identity,
  NetworkPolicy,
  Certificate,
  SecureChannel,
  TrustScore,
  SecurityEvent,
  CertificateAuthority,
  IdentityProvider,
  ZeroTrustConfiguration
} from '../../types/zero-trust';

interface ZeroTrustState {
  identities: Identity[];
  policies: NetworkPolicy[];
  certificates: Certificate[];
  channels: SecureChannel[];
  trustScores: TrustScore[];
  events: SecurityEvent[];
  certificateAuthorities: CertificateAuthority[];
  identityProviders: IdentityProvider[];
  configuration: ZeroTrustConfiguration | null;
  loading: boolean;
  error: string | null;
  connected: boolean;
}

interface ZeroTrustActions {
  // Identity Management
  createIdentity: (identity: Omit<Identity, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  updateIdentity: (identityId: string, updates: Partial<Identity>) => Promise<void>;
  deleteIdentity: (identityId: string) => Promise<void>;
  revokeIdentity: (identityId: string) => Promise<void>;
  
  // Policy Management
  createPolicy: (policy: Omit<NetworkPolicy, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  updatePolicy: (policyId: string, updates: Partial<NetworkPolicy>) => Promise<void>;
  deletePolicy: (policyId: string) => Promise<void>;
  enablePolicy: (policyId: string) => Promise<void>;
  disablePolicy: (policyId: string) => Promise<void>;
  testPolicy: (policy: NetworkPolicy, scenario: any) => Promise<boolean>;
  
  // Certificate Management
  issueCertificate: (request: CertificateRequest) => Promise<Certificate>;
  renewCertificate: (certificateId: string) => Promise<Certificate>;
  revokeCertificate: (certificateId: string, reason: string) => Promise<void>;
  validateCertificate: (certificateId: string) => Promise<CertificateValidation>;
  
  // Channel Management
  establishChannel: (config: ChannelConfig) => Promise<SecureChannel>;
  closeChannel: (channelId: string) => Promise<void>;
  refreshChannel: (channelId: string) => Promise<void>;
  
  // Trust Management
  calculateTrustScore: (identityId: string) => Promise<TrustScore>;
  updateTrustFactors: (identityId: string, factors: any[]) => Promise<void>;
  
  // Event Management
  resolveEvent: (eventId: string, resolution: string) => Promise<void>;
  acknowledgeEvent: (eventId: string) => Promise<void>;
  
  // Configuration
  exportConfiguration: () => Promise<string>;
  importConfiguration: (config: string) => Promise<void>;
  validateConfiguration: (config: ZeroTrustConfiguration) => Promise<ValidationResult>;
}

interface CertificateRequest {
  identity: string;
  type: 'client' | 'server';
  subject: string;
  subjectAltNames: string[];
  keyUsage: string[];
  validityDays: number;
  caId: string;
}

interface CertificateValidation {
  valid: boolean;
  errors: string[];
  warnings: string[];
  chain: Certificate[];
  ocspStatus: 'good' | 'revoked' | 'unknown';
  expiresInDays: number;
}

interface ChannelConfig {
  name: string;
  type: 'mTLS' | 'IPSec' | 'WireGuard';
  source: string;
  destination: string;
  encryption: any;
  authentication: any;
}

interface ValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
}

export function useZeroTrust(): ZeroTrustState & ZeroTrustActions {
  const [state, setState] = useState<ZeroTrustState>({
    identities: [],
    policies: [],
    certificates: [],
    channels: [],
    trustScores: [],
    events: [],
    certificateAuthorities: [],
    identityProviders: [],
    configuration: null,
    loading: true,
    error: null,
    connected: false
  });

  const { isConnected, sendMessage, lastMessage } = useWebSocket('/api/ws/zero-trust');

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
      case 'identity-updated':
        setState(prev => ({
          ...prev,
          identities: prev.identities.map(identity =>
            identity.id === message.data.id ? { ...identity, ...message.data } : identity
          )
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

      case 'certificate-issued':
        setState(prev => ({
          ...prev,
          certificates: [...prev.certificates, message.data]
        }));
        break;

      case 'certificate-revoked':
        setState(prev => ({
          ...prev,
          certificates: prev.certificates.map(cert =>
            cert.id === message.data.id ? { ...cert, status: 'revoked' } : cert
          )
        }));
        break;

      case 'channel-established':
        setState(prev => ({
          ...prev,
          channels: [...prev.channels, message.data]
        }));
        break;

      case 'channel-closed':
        setState(prev => ({
          ...prev,
          channels: prev.channels.filter(channel => channel.id !== message.data.id)
        }));
        break;

      case 'trust-score-updated':
        setState(prev => ({
          ...prev,
          trustScores: prev.trustScores.map(score =>
            score.identity === message.data.identity ? message.data : score
          )
        }));
        break;

      case 'security-event':
        setState(prev => ({
          ...prev,
          events: [message.data, ...prev.events].slice(0, 10000) // Keep last 10k events
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

  // Initialize zero trust data
  useEffect(() => {
    const initializeZeroTrust = async () => {
      try {
        setState(prev => ({ ...prev, loading: true, error: null }));

        const [
          identitiesResponse,
          policiesResponse,
          certificatesResponse,
          channelsResponse,
          trustScoresResponse,
          eventsResponse,
          casResponse,
          providersResponse,
          configResponse
        ] = await Promise.all([
          fetch('/api/zero-trust/identities'),
          fetch('/api/zero-trust/policies'),
          fetch('/api/zero-trust/certificates'),
          fetch('/api/zero-trust/channels'),
          fetch('/api/zero-trust/trust-scores'),
          fetch('/api/zero-trust/events?limit=1000'),
          fetch('/api/zero-trust/certificate-authorities'),
          fetch('/api/zero-trust/identity-providers'),
          fetch('/api/zero-trust/configuration')
        ]);

        const [
          identities,
          policies,
          certificates,
          channels,
          trustScores,
          events,
          certificateAuthorities,
          identityProviders,
          configuration
        ] = await Promise.all([
          identitiesResponse.json(),
          policiesResponse.json(),
          certificatesResponse.json(),
          channelsResponse.json(),
          trustScoresResponse.json(),
          eventsResponse.json(),
          casResponse.json(),
          providersResponse.json(),
          configResponse.json()
        ]);

        setState(prev => ({
          ...prev,
          identities,
          policies,
          certificates,
          channels,
          trustScores,
          events,
          certificateAuthorities,
          identityProviders,
          configuration,
          loading: false
        }));
      } catch (error) {
        setState(prev => ({
          ...prev,
          error: error instanceof Error ? error.message : 'Failed to load zero trust data',
          loading: false
        }));
      }
    };

    initializeZeroTrust();
  }, []);

  // Identity Management
  const createIdentity = useCallback(async (identity: Omit<Identity, 'id' | 'createdAt' | 'updatedAt'>) => {
    try {
      const response = await fetch('/api/zero-trust/identities', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(identity)
      });

      if (!response.ok) {
        throw new Error('Failed to create identity');
      }

      const newIdentity = await response.json();
      setState(prev => ({
        ...prev,
        identities: [...prev.identities, newIdentity]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to create identity'
      }));
      throw error;
    }
  }, []);

  const updateIdentity = useCallback(async (identityId: string, updates: Partial<Identity>) => {
    try {
      const response = await fetch(`/api/zero-trust/identities/${identityId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        throw new Error('Failed to update identity');
      }

      const updatedIdentity = await response.json();
      setState(prev => ({
        ...prev,
        identities: prev.identities.map(identity =>
          identity.id === identityId ? updatedIdentity : identity
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update identity'
      }));
      throw error;
    }
  }, []);

  const deleteIdentity = useCallback(async (identityId: string) => {
    try {
      const response = await fetch(`/api/zero-trust/identities/${identityId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to delete identity');
      }

      setState(prev => ({
        ...prev,
        identities: prev.identities.filter(identity => identity.id !== identityId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to delete identity'
      }));
      throw error;
    }
  }, []);

  const revokeIdentity = useCallback(async (identityId: string) => {
    await updateIdentity(identityId, { status: 'revoked' });
  }, [updateIdentity]);

  // Policy Management
  const createPolicy = useCallback(async (policy: Omit<NetworkPolicy, 'id' | 'createdAt' | 'updatedAt'>) => {
    try {
      const response = await fetch('/api/zero-trust/policies', {
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

  const updatePolicy = useCallback(async (policyId: string, updates: Partial<NetworkPolicy>) => {
    try {
      const response = await fetch(`/api/zero-trust/policies/${policyId}`, {
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
      const response = await fetch(`/api/zero-trust/policies/${policyId}`, {
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

  const testPolicy = useCallback(async (policy: NetworkPolicy, scenario: any): Promise<boolean> => {
    try {
      const response = await fetch('/api/zero-trust/policies/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ policy, scenario })
      });

      if (!response.ok) {
        throw new Error('Failed to test policy');
      }

      const result = await response.json();
      return result.allowed;
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to test policy'
      }));
      throw error;
    }
  }, []);

  // Certificate Management
  const issueCertificate = useCallback(async (request: CertificateRequest): Promise<Certificate> => {
    try {
      const response = await fetch('/api/zero-trust/certificates', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request)
      });

      if (!response.ok) {
        throw new Error('Failed to issue certificate');
      }

      const certificate = await response.json();
      setState(prev => ({
        ...prev,
        certificates: [...prev.certificates, certificate]
      }));

      return certificate;
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to issue certificate'
      }));
      throw error;
    }
  }, []);

  const renewCertificate = useCallback(async (certificateId: string): Promise<Certificate> => {
    try {
      const response = await fetch(`/api/zero-trust/certificates/${certificateId}/renew`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to renew certificate');
      }

      const certificate = await response.json();
      setState(prev => ({
        ...prev,
        certificates: prev.certificates.map(cert =>
          cert.id === certificateId ? certificate : cert
        )
      }));

      return certificate;
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to renew certificate'
      }));
      throw error;
    }
  }, []);

  const revokeCertificate = useCallback(async (certificateId: string, reason: string) => {
    try {
      const response = await fetch(`/api/zero-trust/certificates/${certificateId}/revoke`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ reason })
      });

      if (!response.ok) {
        throw new Error('Failed to revoke certificate');
      }

      setState(prev => ({
        ...prev,
        certificates: prev.certificates.map(cert =>
          cert.id === certificateId ? { ...cert, status: 'revoked' } : cert
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to revoke certificate'
      }));
      throw error;
    }
  }, []);

  const validateCertificate = useCallback(async (certificateId: string): Promise<CertificateValidation> => {
    try {
      const response = await fetch(`/api/zero-trust/certificates/${certificateId}/validate`);

      if (!response.ok) {
        throw new Error('Failed to validate certificate');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to validate certificate'
      }));
      throw error;
    }
  }, []);

  // Channel Management
  const establishChannel = useCallback(async (config: ChannelConfig): Promise<SecureChannel> => {
    try {
      const response = await fetch('/api/zero-trust/channels', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        throw new Error('Failed to establish channel');
      }

      const channel = await response.json();
      setState(prev => ({
        ...prev,
        channels: [...prev.channels, channel]
      }));

      return channel;
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to establish channel'
      }));
      throw error;
    }
  }, []);

  const closeChannel = useCallback(async (channelId: string) => {
    try {
      const response = await fetch(`/api/zero-trust/channels/${channelId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to close channel');
      }

      setState(prev => ({
        ...prev,
        channels: prev.channels.filter(channel => channel.id !== channelId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to close channel'
      }));
      throw error;
    }
  }, []);

  const refreshChannel = useCallback(async (channelId: string) => {
    try {
      const response = await fetch(`/api/zero-trust/channels/${channelId}/refresh`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to refresh channel');
      }

      const channel = await response.json();
      setState(prev => ({
        ...prev,
        channels: prev.channels.map(ch =>
          ch.id === channelId ? channel : ch
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to refresh channel'
      }));
      throw error;
    }
  }, []);

  // Trust Management
  const calculateTrustScore = useCallback(async (identityId: string): Promise<TrustScore> => {
    try {
      const response = await fetch(`/api/zero-trust/trust-scores/${identityId}/calculate`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to calculate trust score');
      }

      const trustScore = await response.json();
      setState(prev => ({
        ...prev,
        trustScores: prev.trustScores.map(score =>
          score.identity === identityId ? trustScore : score
        )
      }));

      return trustScore;
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to calculate trust score'
      }));
      throw error;
    }
  }, []);

  const updateTrustFactors = useCallback(async (identityId: string, factors: any[]) => {
    try {
      const response = await fetch(`/api/zero-trust/trust-scores/${identityId}/factors`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ factors })
      });

      if (!response.ok) {
        throw new Error('Failed to update trust factors');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update trust factors'
      }));
      throw error;
    }
  }, []);

  // Event Management
  const resolveEvent = useCallback(async (eventId: string, resolution: string) => {
    try {
      const response = await fetch(`/api/zero-trust/events/${eventId}/resolve`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ resolution })
      });

      if (!response.ok) {
        throw new Error('Failed to resolve event');
      }

      setState(prev => ({
        ...prev,
        events: prev.events.map(event =>
          event.id === eventId ? { ...event, resolved: true, resolvedAt: new Date() } : event
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to resolve event'
      }));
      throw error;
    }
  }, []);

  const acknowledgeEvent = useCallback(async (eventId: string) => {
    try {
      const response = await fetch(`/api/zero-trust/events/${eventId}/acknowledge`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to acknowledge event');
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to acknowledge event'
      }));
      throw error;
    }
  }, []);

  // Configuration Management
  const exportConfiguration = useCallback(async (): Promise<string> => {
    try {
      const response = await fetch('/api/zero-trust/configuration/export');
      
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
      const response = await fetch('/api/zero-trust/configuration/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ configuration: config })
      });

      if (!response.ok) {
        throw new Error('Failed to import configuration');
      }

      // Reload zero trust data after import
      window.location.reload();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to import configuration'
      }));
      throw error;
    }
  }, []);

  const validateConfiguration = useCallback(async (config: ZeroTrustConfiguration): Promise<ValidationResult> => {
    try {
      const response = await fetch('/api/zero-trust/configuration/validate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
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

  return {
    ...state,
    createIdentity,
    updateIdentity,
    deleteIdentity,
    revokeIdentity,
    createPolicy,
    updatePolicy,
    deletePolicy,
    enablePolicy,
    disablePolicy,
    testPolicy,
    issueCertificate,
    renewCertificate,
    revokeCertificate,
    validateCertificate,
    establishChannel,
    closeChannel,
    refreshChannel,
    calculateTrustScore,
    updateTrustFactors,
    resolveEvent,
    acknowledgeEvent,
    exportConfiguration,
    importConfiguration,
    validateConfiguration
  };
}