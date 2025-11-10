/**
 * MCP (Model Context Protocol) Client
 * Connects to and uses MCP servers for enhanced functionality
 */

const { spawn } = require('child_process');
const EventEmitter = require('events');

class MCPClient extends EventEmitter {
  constructor(serverConfig) {
    super();
    this.serverConfig = serverConfig;
    this.servers = new Map();
    this.initialized = false;
  }

  /**
   * Initialize all configured MCP servers
   */
  async initialize() {
    console.log('🔌 Initializing MCP servers...');
    
    for (const [name, config] of Object.entries(this.serverConfig)) {
      if (config.disabled) {
        console.log(`  ⏭️  Skipping disabled server: ${name}`);
        continue;
      }

      try {
        await this.connectServer(name, config);
        console.log(`  ✓ Connected to ${name}`);
      } catch (error) {
        console.error(`  ✗ Failed to connect to ${name}:`, error.message);
      }
    }

    this.initialized = true;
    console.log('✓ MCP client initialized');
  }

  /**
   * Connect to a single MCP server
   */
  async connectServer(name, config) {
    const process = spawn(config.command, config.args, {
      env: { ...process.env, ...config.env },
      stdio: ['pipe', 'pipe', 'pipe']
    });

    const server = {
      name,
      process,
      config,
      tools: [],
      resources: [],
      prompts: []
    };

    // Handle server output
    process.stdout.on('data', (data) => {
      this.handleServerMessage(name, data);
    });

    process.stderr.on('data', (data) => {
      console.error(`MCP Server ${name} error:`, data.toString());
    });

    process.on('exit', (code) => {
      console.log(`MCP Server ${name} exited with code ${code}`);
      this.servers.delete(name);
    });

    // Send initialization request
    this.sendRequest(process, {
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2024-11-05',
        capabilities: {
          roots: { listChanged: true },
          sampling: {}
        },
        clientInfo: {
          name: 'OllamaMax',
          version: '1.0.0'
        }
      }
    });

    this.servers.set(name, server);
    return server;
  }

  /**
   * Send JSON-RPC request to MCP server
   */
  sendRequest(process, request) {
    const message = JSON.stringify(request) + '\n';
    process.stdin.write(message);
  }

  /**
   * Handle messages from MCP server
   */
  handleServerMessage(serverName, data) {
    try {
      const messages = data.toString().split('\n').filter(Boolean);
      
      for (const msg of messages) {
        const response = JSON.parse(msg);
        
        if (response.method === 'notifications/initialized') {
          this.emit('server-ready', serverName);
        } else if (response.result) {
          this.emit('response', { serverName, response });
        }
      }
    } catch (error) {
      console.error(`Error parsing MCP message from ${serverName}:`, error);
    }
  }

  /**
   * Call a tool on an MCP server
   */
  async callTool(serverName, toolName, args = {}) {
    const server = this.servers.get(serverName);
    if (!server) {
      throw new Error(`MCP server ${serverName} not connected`);
    }

    const requestId = Date.now();
    
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error(`Tool call timeout: ${toolName}`));
      }, 30000);

      const handler = ({ serverName: respServer, response }) => {
        if (respServer === serverName && response.id === requestId) {
          clearTimeout(timeout);
          this.off('response', handler);
          
          if (response.error) {
            reject(new Error(response.error.message));
          } else {
            resolve(response.result);
          }
        }
      };

      this.on('response', handler);

      this.sendRequest(server.process, {
        jsonrpc: '2.0',
        id: requestId,
        method: 'tools/call',
        params: {
          name: toolName,
          arguments: args
        }
      });
    });
  }

  /**
   * Shutdown all MCP servers
   */
  async shutdown() {
    console.log('Shutting down MCP servers...');
    
    for (const [name, server] of this.servers) {
      server.process.kill();
      console.log(`  ✓ Stopped ${name}`);
    }
    
    this.servers.clear();
  }
}

module.exports = MCPClient;

