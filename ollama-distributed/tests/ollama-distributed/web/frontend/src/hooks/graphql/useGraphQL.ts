import { useState, useEffect, useCallback, useRef } from 'react';
import { useWebSocket } from '../useWebSocket';
import {
  GraphQLSchema,
  GraphQLEndpoint,
  GraphQLQuery,
  QueryExecution,
  Subscription,
  SubscriptionMessage,
  SchemaComparison,
  GraphQLConfiguration,
  QueryComplexityAnalysis,
  PlaygroundSettings
} from '../../types/graphql';

interface GraphQLState {
  schemas: GraphQLSchema[];
  endpoints: GraphQLEndpoint[];
  queries: GraphQLQuery[];
  executions: QueryExecution[];
  subscriptions: Subscription[];
  messages: SubscriptionMessage[];
  comparisons: SchemaComparison[];
  configuration: GraphQLConfiguration | null;
  playgroundSettings: PlaygroundSettings;
  loading: boolean;
  error: string | null;
  connected: boolean;
}

interface GraphQLActions {
  // Schema Management
  createSchema: (schema: Omit<GraphQLSchema, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  updateSchema: (schemaId: string, updates: Partial<GraphQLSchema>) => Promise<void>;
  deleteSchema: (schemaId: string) => Promise<void>;
  introspectSchema: (endpointId: string) => Promise<GraphQLSchema>;
  validateSchema: (sdl: string) => Promise<{ valid: boolean; errors: string[] }>;
  
  // Endpoint Management
  createEndpoint: (endpoint: Omit<GraphQLEndpoint, 'id' | 'metrics' | 'lastHealthCheck'>) => Promise<void>;
  updateEndpoint: (endpointId: string, updates: Partial<GraphQLEndpoint>) => Promise<void>;
  deleteEndpoint: (endpointId: string) => Promise<void>;
  testEndpoint: (endpointId: string) => Promise<{ success: boolean; responseTime: number; error?: string }>;
  
  // Query Management
  createQuery: (query: Omit<GraphQLQuery, 'id' | 'createdAt' | 'executionCount' | 'avgExecutionTime'>) => Promise<void>;
  updateQuery: (queryId: string, updates: Partial<GraphQLQuery>) => Promise<void>;
  deleteQuery: (queryId: string) => Promise<void>;
  duplicateQuery: (queryId: string) => Promise<void>;
  
  // Query Execution
  executeQuery: (query: GraphQLQuery, variables?: Record<string, any>) => Promise<QueryExecution>;
  cancelExecution: (executionId: string) => Promise<void>;
  
  // Subscriptions
  createSubscription: (subscription: Omit<Subscription, 'id' | 'createdAt' | 'messageCount' | 'reconnectAttempts'>) => Promise<void>;
  startSubscription: (subscriptionId: string) => Promise<void>;
  stopSubscription: (subscriptionId: string) => Promise<void>;
  deleteSubscription: (subscriptionId: string) => Promise<void>;
  
  // Schema Comparison
  compareSchemas: (baseSchemaId: string, targetSchemaId: string, name: string) => Promise<SchemaComparison>;
  deleteComparison: (comparisonId: string) => Promise<void>;
  
  // Analysis
  analyzeQueryComplexity: (query: string, schemaId: string) => Promise<QueryComplexityAnalysis>;
  
  // Configuration
  updateConfiguration: (config: Partial<GraphQLConfiguration>) => Promise<void>;
  updatePlaygroundSettings: (settings: Partial<PlaygroundSettings>) => Promise<void>;
  
  // Import/Export
  exportQueries: (queryIds: string[]) => Promise<string>;
  importQueries: (data: string) => Promise<void>;
  exportSchema: (schemaId: string) => Promise<string>;
  importSchema: (data: string) => Promise<void>;
}

export function useGraphQL(): GraphQLState & GraphQLActions {
  const [state, setState] = useState<GraphQLState>({
    schemas: [],
    endpoints: [],
    queries: [],
    executions: [],
    subscriptions: [],
    messages: [],
    comparisons: [],
    configuration: null,
    playgroundSettings: {
      theme: 'light',
      fontSize: 14,
      tabSize: 2,
      wordWrap: true,
      autoComplete: true,
      linting: true,
      prettify: true,
      shareableUrls: false,
      requestCredentials: 'same-origin'
    },
    loading: true,
    error: null,
    connected: false
  });

  const { isConnected, sendMessage, lastMessage } = useWebSocket('/api/ws/graphql');
  const subscriptionConnections = useRef<Map<string, WebSocket>>(new Map());

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
      case 'schema-updated':
        setState(prev => ({
          ...prev,
          schemas: prev.schemas.map(schema =>
            schema.id === message.data.id ? { ...schema, ...message.data } : schema
          )
        }));
        break;

      case 'endpoint-updated':
        setState(prev => ({
          ...prev,
          endpoints: prev.endpoints.map(endpoint =>
            endpoint.id === message.data.id ? { ...endpoint, ...message.data } : endpoint
          )
        }));
        break;

      case 'query-executed':
        setState(prev => ({
          ...prev,
          executions: [message.data, ...prev.executions].slice(0, 1000) // Keep last 1000 executions
        }));
        break;

      case 'subscription-message':
        setState(prev => ({
          ...prev,
          messages: [message.data, ...prev.messages].slice(0, 10000) // Keep last 10k messages
        }));
        break;

      case 'subscription-status-changed':
        setState(prev => ({
          ...prev,
          subscriptions: prev.subscriptions.map(sub =>
            sub.id === message.data.id ? { ...sub, ...message.data } : sub
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

  // Initialize GraphQL data
  useEffect(() => {
    const initializeGraphQL = async () => {
      try {
        setState(prev => ({ ...prev, loading: true, error: null }));

        const [
          schemasResponse,
          endpointsResponse,
          queriesResponse,
          executionsResponse,
          subscriptionsResponse,
          messagesResponse,
          comparisonsResponse,
          configResponse,
          settingsResponse
        ] = await Promise.all([
          fetch('/api/graphql/schemas'),
          fetch('/api/graphql/endpoints'),
          fetch('/api/graphql/queries'),
          fetch('/api/graphql/executions?limit=100'),
          fetch('/api/graphql/subscriptions'),
          fetch('/api/graphql/subscription-messages?limit=1000'),
          fetch('/api/graphql/comparisons'),
          fetch('/api/graphql/configuration'),
          fetch('/api/graphql/playground/settings')
        ]);

        const [
          schemas,
          endpoints,
          queries,
          executions,
          subscriptions,
          messages,
          comparisons,
          configuration,
          playgroundSettings
        ] = await Promise.all([
          schemasResponse.json(),
          endpointsResponse.json(),
          queriesResponse.json(),
          executionsResponse.json(),
          subscriptionsResponse.json(),
          messagesResponse.json(),
          comparisonsResponse.json(),
          configResponse.json(),
          settingsResponse.json()
        ]);

        setState(prev => ({
          ...prev,
          schemas,
          endpoints,
          queries,
          executions,
          subscriptions,
          messages,
          comparisons,
          configuration,
          playgroundSettings: { ...prev.playgroundSettings, ...playgroundSettings },
          loading: false
        }));
      } catch (error) {
        setState(prev => ({
          ...prev,
          error: error instanceof Error ? error.message : 'Failed to load GraphQL data',
          loading: false
        }));
      }
    };

    initializeGraphQL();
  }, []);

  // Schema Management
  const createSchema = useCallback(async (schema: Omit<GraphQLSchema, 'id' | 'createdAt' | 'updatedAt'>) => {
    try {
      const response = await fetch('/api/graphql/schemas', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(schema)
      });

      if (!response.ok) {
        throw new Error('Failed to create schema');
      }

      const newSchema = await response.json();
      setState(prev => ({
        ...prev,
        schemas: [...prev.schemas, newSchema]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to create schema'
      }));
      throw error;
    }
  }, []);

  const updateSchema = useCallback(async (schemaId: string, updates: Partial<GraphQLSchema>) => {
    try {
      const response = await fetch(`/api/graphql/schemas/${schemaId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        throw new Error('Failed to update schema');
      }

      const updatedSchema = await response.json();
      setState(prev => ({
        ...prev,
        schemas: prev.schemas.map(schema =>
          schema.id === schemaId ? updatedSchema : schema
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update schema'
      }));
      throw error;
    }
  }, []);

  const deleteSchema = useCallback(async (schemaId: string) => {
    try {
      const response = await fetch(`/api/graphql/schemas/${schemaId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to delete schema');
      }

      setState(prev => ({
        ...prev,
        schemas: prev.schemas.filter(schema => schema.id !== schemaId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to delete schema'
      }));
      throw error;
    }
  }, []);

  const introspectSchema = useCallback(async (endpointId: string): Promise<GraphQLSchema> => {
    try {
      const response = await fetch(`/api/graphql/endpoints/${endpointId}/introspect`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to introspect schema');
      }

      const schema = await response.json();
      setState(prev => ({
        ...prev,
        schemas: [...prev.schemas, schema]
      }));

      return schema;
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to introspect schema'
      }));
      throw error;
    }
  }, []);

  const validateSchema = useCallback(async (sdl: string): Promise<{ valid: boolean; errors: string[] }> => {
    try {
      const response = await fetch('/api/graphql/schemas/validate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sdl })
      });

      if (!response.ok) {
        throw new Error('Failed to validate schema');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to validate schema'
      }));
      throw error;
    }
  }, []);

  // Endpoint Management
  const createEndpoint = useCallback(async (endpoint: Omit<GraphQLEndpoint, 'id' | 'metrics' | 'lastHealthCheck'>) => {
    try {
      const response = await fetch('/api/graphql/endpoints', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(endpoint)
      });

      if (!response.ok) {
        throw new Error('Failed to create endpoint');
      }

      const newEndpoint = await response.json();
      setState(prev => ({
        ...prev,
        endpoints: [...prev.endpoints, newEndpoint]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to create endpoint'
      }));
      throw error;
    }
  }, []);

  const updateEndpoint = useCallback(async (endpointId: string, updates: Partial<GraphQLEndpoint>) => {
    try {
      const response = await fetch(`/api/graphql/endpoints/${endpointId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        throw new Error('Failed to update endpoint');
      }

      const updatedEndpoint = await response.json();
      setState(prev => ({
        ...prev,
        endpoints: prev.endpoints.map(endpoint =>
          endpoint.id === endpointId ? updatedEndpoint : endpoint
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update endpoint'
      }));
      throw error;
    }
  }, []);

  const deleteEndpoint = useCallback(async (endpointId: string) => {
    try {
      const response = await fetch(`/api/graphql/endpoints/${endpointId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to delete endpoint');
      }

      setState(prev => ({
        ...prev,
        endpoints: prev.endpoints.filter(endpoint => endpoint.id !== endpointId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to delete endpoint'
      }));
      throw error;
    }
  }, []);

  const testEndpoint = useCallback(async (endpointId: string): Promise<{ success: boolean; responseTime: number; error?: string }> => {
    try {
      const response = await fetch(`/api/graphql/endpoints/${endpointId}/test`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to test endpoint');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to test endpoint'
      }));
      throw error;
    }
  }, []);

  // Query Management
  const createQuery = useCallback(async (query: Omit<GraphQLQuery, 'id' | 'createdAt' | 'executionCount' | 'avgExecutionTime'>) => {
    try {
      const response = await fetch('/api/graphql/queries', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(query)
      });

      if (!response.ok) {
        throw new Error('Failed to create query');
      }

      const newQuery = await response.json();
      setState(prev => ({
        ...prev,
        queries: [...prev.queries, newQuery]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to create query'
      }));
      throw error;
    }
  }, []);

  const updateQuery = useCallback(async (queryId: string, updates: Partial<GraphQLQuery>) => {
    try {
      const response = await fetch(`/api/graphql/queries/${queryId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        throw new Error('Failed to update query');
      }

      const updatedQuery = await response.json();
      setState(prev => ({
        ...prev,
        queries: prev.queries.map(query =>
          query.id === queryId ? updatedQuery : query
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update query'
      }));
      throw error;
    }
  }, []);

  const deleteQuery = useCallback(async (queryId: string) => {
    try {
      const response = await fetch(`/api/graphql/queries/${queryId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to delete query');
      }

      setState(prev => ({
        ...prev,
        queries: prev.queries.filter(query => query.id !== queryId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to delete query'
      }));
      throw error;
    }
  }, []);

  const duplicateQuery = useCallback(async (queryId: string) => {
    try {
      const response = await fetch(`/api/graphql/queries/${queryId}/duplicate`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to duplicate query');
      }

      const duplicatedQuery = await response.json();
      setState(prev => ({
        ...prev,
        queries: [...prev.queries, duplicatedQuery]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to duplicate query'
      }));
      throw error;
    }
  }, []);

  // Query Execution
  const executeQuery = useCallback(async (query: GraphQLQuery, variables?: Record<string, any>): Promise<QueryExecution> => {
    try {
      const response = await fetch('/api/graphql/execute', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          queryId: query.id,
          query: query.query,
          variables,
          endpoint: query.endpoint
        })
      });

      if (!response.ok) {
        throw new Error('Failed to execute query');
      }

      const execution = await response.json();
      setState(prev => ({
        ...prev,
        executions: [execution, ...prev.executions]
      }));

      return execution;
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to execute query'
      }));
      throw error;
    }
  }, []);

  const cancelExecution = useCallback(async (executionId: string) => {
    try {
      const response = await fetch(`/api/graphql/executions/${executionId}/cancel`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to cancel execution');
      }

      setState(prev => ({
        ...prev,
        executions: prev.executions.map(execution =>
          execution.id === executionId ? { ...execution, status: 'cancelled' } : execution
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to cancel execution'
      }));
      throw error;
    }
  }, []);

  // Subscriptions
  const createSubscription = useCallback(async (subscription: Omit<Subscription, 'id' | 'createdAt' | 'messageCount' | 'reconnectAttempts'>) => {
    try {
      const response = await fetch('/api/graphql/subscriptions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(subscription)
      });

      if (!response.ok) {
        throw new Error('Failed to create subscription');
      }

      const newSubscription = await response.json();
      setState(prev => ({
        ...prev,
        subscriptions: [...prev.subscriptions, newSubscription]
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to create subscription'
      }));
      throw error;
    }
  }, []);

  const startSubscription = useCallback(async (subscriptionId: string) => {
    try {
      const response = await fetch(`/api/graphql/subscriptions/${subscriptionId}/start`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to start subscription');
      }

      setState(prev => ({
        ...prev,
        subscriptions: prev.subscriptions.map(sub =>
          sub.id === subscriptionId ? { ...sub, status: 'connecting' } : sub
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to start subscription'
      }));
      throw error;
    }
  }, []);

  const stopSubscription = useCallback(async (subscriptionId: string) => {
    try {
      const response = await fetch(`/api/graphql/subscriptions/${subscriptionId}/stop`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error('Failed to stop subscription');
      }

      setState(prev => ({
        ...prev,
        subscriptions: prev.subscriptions.map(sub =>
          sub.id === subscriptionId ? { ...sub, status: 'disconnected' } : sub
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to stop subscription'
      }));
      throw error;
    }
  }, []);

  const deleteSubscription = useCallback(async (subscriptionId: string) => {
    try {
      const response = await fetch(`/api/graphql/subscriptions/${subscriptionId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to delete subscription');
      }

      setState(prev => ({
        ...prev,
        subscriptions: prev.subscriptions.filter(sub => sub.id !== subscriptionId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to delete subscription'
      }));
      throw error;
    }
  }, []);

  // Schema Comparison
  const compareSchemas = useCallback(async (baseSchemaId: string, targetSchemaId: string, name: string): Promise<SchemaComparison> => {
    try {
      const response = await fetch('/api/graphql/comparisons', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ baseSchemaId, targetSchemaId, name })
      });

      if (!response.ok) {
        throw new Error('Failed to compare schemas');
      }

      const comparison = await response.json();
      setState(prev => ({
        ...prev,
        comparisons: [...prev.comparisons, comparison]
      }));

      return comparison;
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to compare schemas'
      }));
      throw error;
    }
  }, []);

  const deleteComparison = useCallback(async (comparisonId: string) => {
    try {
      const response = await fetch(`/api/graphql/comparisons/${comparisonId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error('Failed to delete comparison');
      }

      setState(prev => ({
        ...prev,
        comparisons: prev.comparisons.filter(comp => comp.id !== comparisonId)
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to delete comparison'
      }));
      throw error;
    }
  }, []);

  // Analysis
  const analyzeQueryComplexity = useCallback(async (query: string, schemaId: string): Promise<QueryComplexityAnalysis> => {
    try {
      const response = await fetch('/api/graphql/analyze/complexity', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query, schemaId })
      });

      if (!response.ok) {
        throw new Error('Failed to analyze query complexity');
      }

      return await response.json();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to analyze query complexity'
      }));
      throw error;
    }
  }, []);

  // Configuration
  const updateConfiguration = useCallback(async (config: Partial<GraphQLConfiguration>) => {
    try {
      const response = await fetch('/api/graphql/configuration', {
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

  const updatePlaygroundSettings = useCallback(async (settings: Partial<PlaygroundSettings>) => {
    try {
      const response = await fetch('/api/graphql/playground/settings', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(settings)
      });

      if (!response.ok) {
        throw new Error('Failed to update playground settings');
      }

      const updatedSettings = await response.json();
      setState(prev => ({
        ...prev,
        playgroundSettings: { ...prev.playgroundSettings, ...updatedSettings }
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to update playground settings'
      }));
      throw error;
    }
  }, []);

  // Import/Export
  const exportQueries = useCallback(async (queryIds: string[]): Promise<string> => {
    try {
      const response = await fetch('/api/graphql/queries/export', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ queryIds })
      });

      if (!response.ok) {
        throw new Error('Failed to export queries');
      }

      return await response.text();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to export queries'
      }));
      throw error;
    }
  }, []);

  const importQueries = useCallback(async (data: string) => {
    try {
      const response = await fetch('/api/graphql/queries/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ data })
      });

      if (!response.ok) {
        throw new Error('Failed to import queries');
      }

      // Reload queries after import
      const queriesResponse = await fetch('/api/graphql/queries');
      const queries = await queriesResponse.json();
      
      setState(prev => ({
        ...prev,
        queries
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to import queries'
      }));
      throw error;
    }
  }, []);

  const exportSchema = useCallback(async (schemaId: string): Promise<string> => {
    try {
      const response = await fetch(`/api/graphql/schemas/${schemaId}/export`);

      if (!response.ok) {
        throw new Error('Failed to export schema');
      }

      return await response.text();
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to export schema'
      }));
      throw error;
    }
  }, []);

  const importSchema = useCallback(async (data: string) => {
    try {
      const response = await fetch('/api/graphql/schemas/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ data })
      });

      if (!response.ok) {
        throw new Error('Failed to import schema');
      }

      // Reload schemas after import
      const schemasResponse = await fetch('/api/graphql/schemas');
      const schemas = await schemasResponse.json();
      
      setState(prev => ({
        ...prev,
        schemas
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to import schema'
      }));
      throw error;
    }
  }, []);

  return {
    ...state,
    createSchema,
    updateSchema,
    deleteSchema,
    introspectSchema,
    validateSchema,
    createEndpoint,
    updateEndpoint,
    deleteEndpoint,
    testEndpoint,
    createQuery,
    updateQuery,
    deleteQuery,
    duplicateQuery,
    executeQuery,
    cancelExecution,
    createSubscription,
    startSubscription,
    stopSubscription,
    deleteSubscription,
    compareSchemas,
    deleteComparison,
    analyzeQueryComplexity,
    updateConfiguration,
    updatePlaygroundSettings,
    exportQueries,
    importQueries,
    exportSchema,
    importSchema
  };
}