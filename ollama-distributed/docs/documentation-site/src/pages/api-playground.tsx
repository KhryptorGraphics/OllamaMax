import React, { useState, useEffect } from 'react';
import Layout from '@theme/Layout';
import clsx from 'clsx';
import styles from './api-playground.module.css';

interface ApiRequest {
  method: string;
  endpoint: string;
  headers: Record<string, string>;
  body?: string;
}

interface ApiResponse {
  status: number;
  headers: Record<string, string>;
  body: string;
  responseTime: number;
}

const API_ENDPOINTS = [
  { path: '/api/v1/health', method: 'GET', description: 'Get system health status' },
  { path: '/api/v1/cluster/status', method: 'GET', description: 'Get cluster status' },
  { path: '/api/v1/nodes', method: 'GET', description: 'List all nodes' },
  { path: '/api/v1/models', method: 'GET', description: 'List available models' },
  { path: '/api/v1/models/{name}/download', method: 'POST', description: 'Download a model' },
  { path: '/api/v1/generate', method: 'POST', description: 'Generate text' },
  { path: '/api/v1/chat', method: 'POST', description: 'Chat completion' },
  { path: '/api/v1/embeddings', method: 'POST', description: 'Generate embeddings' },
  { path: '/api/v1/metrics', method: 'GET', description: 'Get system metrics' },
];

const EXAMPLE_REQUESTS = {
  '/api/v1/generate': {
    method: 'POST',
    body: JSON.stringify({
      model: 'llama2',
      prompt: 'Explain artificial intelligence in simple terms',
      stream: false
    }, null, 2)
  },
  '/api/v1/chat': {
    method: 'POST',
    body: JSON.stringify({
      model: 'llama2',
      messages: [
        { role: 'user', content: 'Hello, how are you?' }
      ],
      stream: false
    }, null, 2)
  },
  '/api/v1/models/{name}/download': {
    method: 'POST',
    body: JSON.stringify({
      name: 'llama2'
    }, null, 2)
  }
};

export default function ApiPlayground(): JSX.Element {
  const [baseUrl, setBaseUrl] = useState('http://localhost:8080');
  const [apiKey, setApiKey] = useState('');
  const [selectedEndpoint, setSelectedEndpoint] = useState(API_ENDPOINTS[0]);
  const [requestBody, setRequestBody] = useState('');
  const [customHeaders, setCustomHeaders] = useState('');
  const [response, setResponse] = useState<ApiResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [wsConnected, setWsConnected] = useState(false);
  const [wsMessages, setWsMessages] = useState<string[]>([]);

  useEffect(() => {
    // Load example request body when endpoint changes
    const exampleRequest = EXAMPLE_REQUESTS[selectedEndpoint.path];
    if (exampleRequest && exampleRequest.body) {
      setRequestBody(exampleRequest.body);
    } else {
      setRequestBody('');
    }
  }, [selectedEndpoint]);

  const makeRequest = async () => {
    setLoading(true);
    setResponse(null);

    try {
      const startTime = Date.now();
      const url = `${baseUrl}${selectedEndpoint.path.replace('{name}', 'llama2')}`;
      
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };

      if (apiKey) {
        headers['Authorization'] = `Bearer ${apiKey}`;
      }

      // Add custom headers
      if (customHeaders) {
        try {
          const parsed = JSON.parse(customHeaders);
          Object.assign(headers, parsed);
        } catch (e) {
          console.warn('Invalid custom headers JSON');
        }
      }

      const requestOptions: RequestInit = {
        method: selectedEndpoint.method,
        headers,
      };

      if (selectedEndpoint.method !== 'GET' && requestBody) {
        requestOptions.body = requestBody;
      }

      const response = await fetch(url, requestOptions);
      const responseBody = await response.text();
      const endTime = Date.now();

      setResponse({
        status: response.status,
        headers: Object.fromEntries(response.headers.entries()),
        body: responseBody,
        responseTime: endTime - startTime,
      });
    } catch (error) {
      setResponse({
        status: 0,
        headers: {},
        body: `Error: ${error.message}`,
        responseTime: 0,
      });
    } finally {
      setLoading(false);
    }
  };

  const connectWebSocket = () => {
    const ws = new WebSocket(`ws://${baseUrl.replace('http://', '').replace('https://', '')}/api/v1/ws`);
    
    ws.onopen = () => {
      setWsConnected(true);
      setWsMessages(prev => [...prev, 'Connected to WebSocket']);
      
      // Subscribe to metrics
      ws.send(JSON.stringify({ type: 'subscribe', channel: 'metrics' }));
    };

    ws.onmessage = (event) => {
      setWsMessages(prev => [...prev, event.data]);
    };

    ws.onclose = () => {
      setWsConnected(false);
      setWsMessages(prev => [...prev, 'WebSocket connection closed']);
    };

    ws.onerror = (error) => {
      setWsMessages(prev => [...prev, `WebSocket error: ${error}`]);
    };
  };

  const formatJson = (jsonString: string) => {
    try {
      return JSON.stringify(JSON.parse(jsonString), null, 2);
    } catch {
      return jsonString;
    }
  };

  return (
    <Layout
      title="API Playground"
      description="Interactive API testing playground for Ollama Distributed"
    >
      <div className={clsx('container', styles.playground)}>
        <div className={styles.header}>
          <h1>API Playground</h1>
          <p>Interactive testing environment for Ollama Distributed API endpoints</p>
        </div>

        <div className={styles.content}>
          {/* Configuration Panel */}
          <div className={styles.configPanel}>
            <h2>Configuration</h2>
            <div className={styles.configForm}>
              <div className={styles.formGroup}>
                <label>Base URL:</label>
                <input
                  type="text"
                  value={baseUrl}
                  onChange={(e) => setBaseUrl(e.target.value)}
                  placeholder="http://localhost:8080"
                />
              </div>
              <div className={styles.formGroup}>
                <label>API Key (optional):</label>
                <input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="Enter your API key"
                />
              </div>
            </div>
          </div>

          {/* Request Panel */}
          <div className={styles.requestPanel}>
            <h2>Request</h2>
            <div className={styles.endpointSelector}>
              <label>Endpoint:</label>
              <select
                value={selectedEndpoint.path}
                onChange={(e) => {
                  const endpoint = API_ENDPOINTS.find(ep => ep.path === e.target.value);
                  setSelectedEndpoint(endpoint);
                }}
              >
                {API_ENDPOINTS.map((endpoint) => (
                  <option key={endpoint.path} value={endpoint.path}>
                    {endpoint.method} {endpoint.path} - {endpoint.description}
                  </option>
                ))}
              </select>
            </div>

            <div className={styles.methodBadge}>
              <span className={clsx(styles.badge, styles[selectedEndpoint.method.toLowerCase()])}>
                {selectedEndpoint.method}
              </span>
              <code>{selectedEndpoint.path}</code>
            </div>

            {selectedEndpoint.method !== 'GET' && (
              <div className={styles.formGroup}>
                <label>Request Body:</label>
                <textarea
                  value={requestBody}
                  onChange={(e) => setRequestBody(e.target.value)}
                  placeholder="Enter JSON request body"
                  rows={8}
                />
              </div>
            )}

            <div className={styles.formGroup}>
              <label>Custom Headers (JSON):</label>
              <textarea
                value={customHeaders}
                onChange={(e) => setCustomHeaders(e.target.value)}
                placeholder='{"Custom-Header": "value"}'
                rows={3}
              />
            </div>

            <button
              className={styles.sendButton}
              onClick={makeRequest}
              disabled={loading}
            >
              {loading ? 'Sending...' : 'Send Request'}
            </button>
          </div>

          {/* Response Panel */}
          <div className={styles.responsePanel}>
            <h2>Response</h2>
            {response ? (
              <div className={styles.response}>
                <div className={styles.responseHeader}>
                  <span className={clsx(styles.statusBadge, {
                    [styles.success]: response.status >= 200 && response.status < 300,
                    [styles.error]: response.status >= 400,
                  })}>
                    {response.status}
                  </span>
                  <span className={styles.responseTime}>
                    {response.responseTime}ms
                  </span>
                </div>
                
                <div className={styles.responseBody}>
                  <h3>Response Body:</h3>
                  <pre>{formatJson(response.body)}</pre>
                </div>

                <div className={styles.responseHeaders}>
                  <h3>Response Headers:</h3>
                  <pre>{JSON.stringify(response.headers, null, 2)}</pre>
                </div>
              </div>
            ) : (
              <div className={styles.placeholder}>
                No response yet. Send a request to see the response.
              </div>
            )}
          </div>

          {/* WebSocket Panel */}
          <div className={styles.websocketPanel}>
            <h2>WebSocket Real-time Updates</h2>
            <div className={styles.wsControls}>
              <button
                className={styles.wsButton}
                onClick={connectWebSocket}
                disabled={wsConnected}
              >
                {wsConnected ? 'Connected' : 'Connect to WebSocket'}
              </button>
              <span className={clsx(styles.wsStatus, {
                [styles.connected]: wsConnected,
                [styles.disconnected]: !wsConnected,
              })}>
                {wsConnected ? 'Connected' : 'Disconnected'}
              </span>
            </div>
            
            <div className={styles.wsMessages}>
              {wsMessages.map((message, index) => (
                <div key={index} className={styles.wsMessage}>
                  <pre>{message}</pre>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}