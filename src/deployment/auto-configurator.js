/**
 * Auto-Detection and Configuration System for Ollamamax
 * Automatically detects environment and configures optimal settings
 */

const os = require('os');
const fs = require('fs').promises;
const path = require('path');
const { exec } = require('child_process');
const { promisify } = require('util');
const execAsync = promisify(exec);
const HardwareAbstractionLayer = require('../hardware/abstraction-layer');
const PlatformOptimizer = require('../hardware/platform-optimizer');

class AutoConfigurator {
  constructor() {
    this.hal = null;
    this.optimizer = null;
    this.detectedConfig = {};
    this.recommendedConfig = {};
    this.appliedConfig = {};
  }

  async initialize() {
    console.log('🔍 Starting auto-detection and configuration...');
    
    // Initialize hardware abstraction
    this.hal = new HardwareAbstractionLayer();
    await this.hal.initialize();
    
    // Initialize optimizer
    this.optimizer = new PlatformOptimizer(this.hal);
    
    // Run auto-detection
    await this.detect();
    
    // Generate recommendations
    await this.recommend();
    
    console.log('✅ Auto-configuration complete');
    return this;
  }

  async detect() {
    console.log('🔍 Detecting environment...');
    
    this.detectedConfig = {
      // Hardware detection
      hardware: {
        platform: this.hal.platform,
        capabilities: Object.fromEntries(this.hal.capabilities),
        metrics: this.hal.metrics
      },
      
      // Software detection
      software: await this.detectSoftware(),
      
      // Network detection
      network: await this.detectNetwork(),
      
      // Container detection
      container: await this.detectContainer(),
      
      // Cloud detection
      cloud: await this.detectCloud(),
      
      // Existing Ollama detection
      ollama: await this.detectOllama(),
      
      // Model detection
      models: await this.detectModels()
    };
    
    return this.detectedConfig;
  }

  async detectSoftware() {
    const software = {
      os: os.platform(),
      arch: os.arch(),
      nodeVersion: process.version,
      npm: null,
      python: null,
      cuda: null,
      docker: null,
      kubernetes: null
    };

    // Detect NPM version
    try {
      const { stdout: npmVersion } = await execAsync('npm --version');
      software.npm = npmVersion.trim();
    } catch {}

    // Detect Python version
    try {
      const { stdout: pythonVersion } = await execAsync('python3 --version');
      software.python = pythonVersion.trim().replace('Python ', '');
    } catch {}

    // Detect CUDA version
    try {
      const { stdout: cudaVersion } = await execAsync('nvcc --version | grep release');
      const match = cudaVersion.match(/release ([\d.]+)/);
      if (match) software.cuda = match[1];
    } catch {}

    // Detect Docker
    try {
      const { stdout: dockerVersion } = await execAsync('docker --version');
      software.docker = dockerVersion.trim();
    } catch {}

    // Detect Kubernetes
    try {
      const { stdout: k8sVersion } = await execAsync('kubectl version --client --short');
      software.kubernetes = k8sVersion.trim();
    } catch {}

    return software;
  }

  async detectNetwork() {
    const network = {
      interfaces: os.networkInterfaces(),
      publicIP: null,
      privateIP: null,
      bandwidth: null,
      latency: null,
      nat: false,
      firewall: null,
      proxy: null
    };

    // Get primary network interface
    const interfaces = Object.values(network.interfaces).flat();
    const primary = interfaces.find(i => !i.internal && i.family === 'IPv4');
    if (primary) {
      network.privateIP = primary.address;
    }

    // Detect public IP
    try {
      const { stdout } = await execAsync('curl -s https://api.ipify.org');
      network.publicIP = stdout.trim();
      network.nat = network.publicIP !== network.privateIP;
    } catch {}

    // Detect bandwidth (rough estimate)
    try {
      const start = Date.now();
      await execAsync('curl -s -o /dev/null https://speed.cloudflare.com/__down?bytes=1000000');
      const duration = Date.now() - start;
      network.bandwidth = Math.round((1000000 * 8) / (duration / 1000)); // bits per second
    } catch {}

    // Check for proxy
    network.proxy = process.env.HTTP_PROXY || process.env.HTTPS_PROXY || null;

    return network;
  }

  async detectContainer() {
    const container = {
      isContainer: false,
      runtime: null,
      orchestrator: null
    };

    // Check if running in Docker
    try {
      await fs.access('/.dockerenv');
      container.isContainer = true;
      container.runtime = 'docker';
    } catch {}

    // Check for Kubernetes
    try {
      await fs.access('/var/run/secrets/kubernetes.io');
      container.orchestrator = 'kubernetes';
    } catch {}

    // Check for other container runtimes
    if (!container.isContainer) {
      try {
        const { stdout } = await execAsync('cat /proc/1/cgroup');
        if (stdout.includes('docker') || stdout.includes('containerd')) {
          container.isContainer = true;
          container.runtime = 'docker';
        } else if (stdout.includes('lxc')) {
          container.isContainer = true;
          container.runtime = 'lxc';
        }
      } catch {}
    }

    return container;
  }

  async detectCloud() {
    const cloud = {
      provider: null,
      instance: null,
      region: null,
      zone: null
    };

    // AWS detection
    try {
      const { stdout } = await execAsync('curl -s -m 1 http://169.254.169.254/latest/meta-data/instance-id');
      if (stdout) {
        cloud.provider = 'aws';
        cloud.instance = stdout.trim();
        
        try {
          const { stdout: region } = await execAsync('curl -s -m 1 http://169.254.169.254/latest/meta-data/placement/region');
          cloud.region = region.trim();
        } catch {}
      }
    } catch {}

    // GCP detection
    if (!cloud.provider) {
      try {
        const { stdout } = await execAsync('curl -s -m 1 -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/id');
        if (stdout) {
          cloud.provider = 'gcp';
          cloud.instance = stdout.trim();
        }
      } catch {}
    }

    // Azure detection
    if (!cloud.provider) {
      try {
        const { stdout } = await execAsync('curl -s -m 1 -H "Metadata: true" http://169.254.169.254/metadata/instance?api-version=2021-02-01');
        if (stdout) {
          const metadata = JSON.parse(stdout);
          cloud.provider = 'azure';
          cloud.instance = metadata.compute.vmId;
          cloud.region = metadata.compute.location;
        }
      } catch {}
    }

    return cloud;
  }

  async detectOllama() {
    const ollama = {
      installed: false,
      version: null,
      models: [],
      apiEndpoint: null,
      running: false
    };

    // Check if Ollama is installed
    try {
      const { stdout } = await execAsync('ollama --version');
      ollama.installed = true;
      ollama.version = stdout.trim();
    } catch {}

    // Check if Ollama is running
    try {
      const { stdout } = await execAsync('curl -s http://localhost:11434/api/tags');
      if (stdout) {
        ollama.running = true;
        ollama.apiEndpoint = 'http://localhost:11434';
        const data = JSON.parse(stdout);
        ollama.models = data.models || [];
      }
    } catch {}

    return ollama;
  }

  async detectModels() {
    const models = {
      available: [],
      downloaded: [],
      recommended: []
    };

    // Check for existing model files
    const modelPaths = [
      '/opt/models',
      '/usr/local/models',
      path.join(os.homedir(), '.ollama', 'models'),
      path.join(os.homedir(), '.cache', 'huggingface'),
      './models'
    ];

    for (const modelPath of modelPaths) {
      try {
        const files = await fs.readdir(modelPath);
        for (const file of files) {
          if (file.endsWith('.gguf') || file.endsWith('.bin')) {
            models.downloaded.push({
              name: file,
              path: path.join(modelPath, file),
              size: (await fs.stat(path.join(modelPath, file))).size
            });
          }
        }
      } catch {}
    }

    // Recommend models based on hardware
    const memory = this.hal.capabilities.get('memory').total;
    const gpu = this.hal.capabilities.get('gpu');

    if (memory < 8 * 1024 * 1024 * 1024) {
      models.recommended = ['llama-3.2-1b', 'phi-3-mini'];
    } else if (memory < 16 * 1024 * 1024 * 1024) {
      models.recommended = ['llama-3.2-3b', 'mistral-7b-instruct'];
    } else if (gpu && gpu.memory > 8 * 1024 * 1024 * 1024) {
      models.recommended = ['llama-3.1-8b', 'mixtral-8x7b'];
    } else {
      models.recommended = ['llama-3.1-8b', 'qwen2.5-14b'];
    }

    return models;
  }

  async recommend() {
    console.log('💡 Generating configuration recommendations...');
    
    const platform = this.hal.platform.type;
    const capabilities = this.hal.capabilities;
    const metrics = this.hal.metrics;
    
    this.recommendedConfig = {
      // Deployment method
      deployment: this.recommendDeployment(),
      
      // Model configuration
      models: this.recommendModels(),
      
      // Network configuration
      network: this.recommendNetwork(),
      
      // Performance settings
      performance: this.recommendPerformance(),
      
      // Security settings
      security: this.recommendSecurity(),
      
      // Integration settings
      integrations: this.recommendIntegrations()
    };
    
    return this.recommendedConfig;
  }

  recommendDeployment() {
    const container = this.detectedConfig.container;
    const cloud = this.detectedConfig.cloud;
    const kubernetes = this.detectedConfig.software.kubernetes;
    const docker = this.detectedConfig.software.docker;
    
    if (container.orchestrator === 'kubernetes' || kubernetes) {
      return {
        method: 'kubernetes',
        reasoning: 'Kubernetes detected, using cloud-native deployment'
      };
    } else if (container.isContainer || docker) {
      return {
        method: 'docker',
        reasoning: 'Docker environment detected'
      };
    } else if (cloud.provider) {
      return {
        method: 'cloud-native',
        provider: cloud.provider,
        reasoning: `Running on ${cloud.provider}, using cloud-optimized deployment`
      };
    } else if (this.hal.platform.type === 'embedded') {
      return {
        method: 'systemd',
        reasoning: 'Embedded platform, using lightweight systemd service'
      };
    } else {
      return {
        method: 'native',
        reasoning: 'Standard installation for this platform'
      };
    }
  }

  recommendModels() {
    const memory = this.hal.capabilities.get('memory').total;
    const gpu = this.hal.capabilities.get('gpu');
    const quantization = this.hal.getQuantizationLevel();
    
    const recommendations = {
      primary: null,
      fallback: null,
      quantization,
      batchSize: this.hal.getOptimalBatchSize(),
      contextLength: 2048,
      reasoning: []
    };
    
    // Memory-based recommendations
    const memoryGB = Math.floor(memory / (1024 * 1024 * 1024));
    
    if (memoryGB < 4) {
      recommendations.primary = 'llama-3.2-1b-q4';
      recommendations.contextLength = 512;
      recommendations.reasoning.push('Limited memory, using smallest model');
    } else if (memoryGB < 8) {
      recommendations.primary = 'llama-3.2-3b-q4';
      recommendations.contextLength = 1024;
      recommendations.reasoning.push('Moderate memory, using 3B parameter model');
    } else if (memoryGB < 16) {
      recommendations.primary = 'mistral-7b-q4';
      recommendations.contextLength = 2048;
      recommendations.reasoning.push('Good memory, using 7B parameter model');
    } else if (gpu && gpu.memory > 16 * 1024 * 1024 * 1024) {
      recommendations.primary = 'mixtral-8x7b-q4';
      recommendations.contextLength = 4096;
      recommendations.reasoning.push('High-end GPU detected, using MoE model');
    } else {
      recommendations.primary = 'llama-3.1-8b-q4';
      recommendations.contextLength = 4096;
      recommendations.reasoning.push('Ample memory, using 8B parameter model');
    }
    
    // Set fallback model (one size smaller)
    const modelSizes = ['1b', '3b', '7b', '8b', '14b', '70b'];
    const currentSize = recommendations.primary.match(/(\d+b)/)?.[1];
    const currentIndex = modelSizes.indexOf(currentSize);
    if (currentIndex > 0) {
      recommendations.fallback = recommendations.primary.replace(currentSize, modelSizes[currentIndex - 1]);
    }
    
    return recommendations;
  }

  recommendNetwork() {
    const network = this.detectedConfig.network;
    const platform = this.hal.platform.type;
    
    return {
      port: network.nat ? 13000 : 13000, // Use non-standard port
      bind: platform === 'datacenter' ? '0.0.0.0' : '127.0.0.1',
      protocol: platform === 'datacenter' ? 'grpc' : 'http',
      compression: network.bandwidth < 10000000 ? 'gzip' : 'none',
      ssl: {
        enabled: platform === 'datacenter' || this.detectedConfig.cloud.provider,
        selfSigned: !this.detectedConfig.cloud.provider
      },
      cors: {
        enabled: true,
        origins: ['http://localhost:*', 'https://localhost:*']
      }
    };
  }

  recommendPerformance() {
    const optimizations = this.optimizer.optimizations;
    
    return {
      threading: this.hal.getThreadingStrategy(),
      caching: this.hal.getCacheStrategy(),
      memoryLimit: this.hal.getMemoryLimit(),
      swapEnabled: this.hal.platform.type !== 'datacenter',
      gpuEnabled: this.hal.shouldUseGPU(),
      quantization: this.hal.getQuantizationLevel(),
      batchSize: this.hal.getOptimalBatchSize(),
      parallelRequests: Math.min(4, this.hal.capabilities.get('cpu').cores),
      requestTimeout: 300000, // 5 minutes
      keepAlive: true
    };
  }

  recommendSecurity() {
    const isProduction = this.detectedConfig.cloud.provider || 
                        this.detectedConfig.container.orchestrator === 'kubernetes';
    
    return {
      authentication: {
        enabled: isProduction,
        method: 'jwt',
        expiresIn: '24h'
      },
      rateLimit: {
        enabled: true,
        requests: isProduction ? 100 : 1000,
        window: '1m'
      },
      encryption: {
        atRest: isProduction,
        inTransit: isProduction,
        algorithm: 'aes-256-gcm'
      },
      audit: {
        enabled: isProduction,
        level: 'info',
        retention: '30d'
      },
      sandbox: {
        enabled: this.hal.platform.type === 'datacenter',
        runtime: 'gvisor'
      }
    };
  }

  recommendIntegrations() {
    const ollama = this.detectedConfig.ollama;
    const docker = this.detectedConfig.software.docker;
    
    const integrations = {
      ollama: {
        enabled: ollama.installed,
        endpoint: ollama.apiEndpoint,
        compatibility: true
      },
      openai: {
        enabled: true,
        compatibility: true,
        endpoint: '/v1'
      },
      langchain: {
        enabled: true,
        vectorStore: 'chroma'
      },
      monitoring: {
        prometheus: docker || this.detectedConfig.container.isContainer,
        grafana: docker || this.detectedConfig.container.isContainer
      }
    };
    
    return integrations;
  }

  async apply(config = null) {
    console.log('⚙️ Applying configuration...');
    
    const targetConfig = config || this.recommendedConfig;
    
    // Create configuration file
    const configPath = path.join(process.cwd(), 'ollamamax.config.json');
    
    const finalConfig = {
      version: '1.0.0',
      detected: this.detectedConfig,
      applied: targetConfig,
      timestamp: new Date().toISOString()
    };
    
    await fs.writeFile(configPath, JSON.stringify(finalConfig, null, 2));
    
    // Apply system-level optimizations
    await this.applySystemOptimizations(targetConfig);
    
    // Create startup script
    await this.createStartupScript(targetConfig);
    
    // Set environment variables
    await this.setEnvironmentVariables(targetConfig);
    
    this.appliedConfig = targetConfig;
    
    console.log('✅ Configuration applied successfully');
    console.log(`📄 Configuration saved to: ${configPath}`);
    
    return this.appliedConfig;
  }

  async applySystemOptimizations(config) {
    const platform = this.hal.platform.type;
    
    // Linux-specific optimizations
    if (os.platform() === 'linux') {
      try {
        // Increase file descriptors
        await execAsync('ulimit -n 65536').catch(() => {});
        
        // Set swappiness
        if (config.performance.swapEnabled) {
          await execAsync('sudo sysctl vm.swappiness=10').catch(() => {});
        }
        
        // Enable huge pages if recommended
        if (platform === 'datacenter') {
          await execAsync('sudo sysctl vm.nr_hugepages=512').catch(() => {});
        }
      } catch (error) {
        console.warn('Some system optimizations could not be applied:', error.message);
      }
    }
  }

  async createStartupScript(config) {
    const scriptContent = `#!/bin/bash
# Ollamamax Auto-Generated Startup Script
# Generated: ${new Date().toISOString()}

# Environment setup
export NODE_ENV=production
export OLLAMAMAX_PORT=${config.network.port}
export OLLAMAMAX_HOST=${config.network.bind}
export OLLAMAMAX_MODEL=${config.models.primary}
export OLLAMAMAX_BATCH_SIZE=${config.performance.batchSize}
export OLLAMAMAX_CACHE_STRATEGY=${config.performance.caching}
export OLLAMAMAX_GPU_ENABLED=${config.performance.gpuEnabled}

# Memory limits
export NODE_OPTIONS="--max-old-space-size=${Math.floor(config.performance.memoryLimit / (1024 * 1024))}"

# Start Ollamamax
echo "Starting Ollamamax with auto-detected configuration..."
node main.js
`;
    
    const scriptPath = path.join(process.cwd(), 'start-ollamamax.sh');
    await fs.writeFile(scriptPath, scriptContent);
    await execAsync(`chmod +x ${scriptPath}`);
    
    console.log(`📝 Startup script created: ${scriptPath}`);
  }

  async setEnvironmentVariables(config) {
    const envPath = path.join(process.cwd(), '.env.auto');
    
    const envContent = `# Ollamamax Auto-Generated Environment Variables
# Generated: ${new Date().toISOString()}

# Network Configuration
PORT=${config.network.port}
HOST=${config.network.bind}
PROTOCOL=${config.network.protocol}

# Model Configuration
DEFAULT_MODEL=${config.models.primary}
FALLBACK_MODEL=${config.models.fallback}
QUANTIZATION=${config.models.quantization}
BATCH_SIZE=${config.models.batchSize}
CONTEXT_LENGTH=${config.models.contextLength}

# Performance Configuration
THREADING_STRATEGY=${config.performance.threading}
CACHE_STRATEGY=${config.performance.caching}
MEMORY_LIMIT=${config.performance.memoryLimit}
GPU_ENABLED=${config.performance.gpuEnabled}
PARALLEL_REQUESTS=${config.performance.parallelRequests}

# Security Configuration
AUTH_ENABLED=${config.security.authentication.enabled}
RATE_LIMIT_ENABLED=${config.security.rateLimit.enabled}
ENCRYPTION_ENABLED=${config.security.encryption.inTransit}

# Integration Configuration
OLLAMA_COMPAT=${config.integrations.ollama.enabled}
OPENAI_COMPAT=${config.integrations.openai.enabled}
`;
    
    await fs.writeFile(envPath, envContent);
    console.log(`📝 Environment variables saved: ${envPath}`);
  }

  // Interactive configuration
  async interactive() {
    console.log('\n🎮 Interactive Configuration Mode');
    console.log('================================\n');
    
    // Show detection results
    console.log('📊 Detected Environment:');
    console.log(`  Platform: ${this.hal.platform.name}`);
    console.log(`  Memory: ${Math.floor(this.hal.capabilities.get('memory').total / (1024 * 1024 * 1024))}GB`);
    console.log(`  GPU: ${this.hal.capabilities.get('gpu')?.model || 'None'}`);
    console.log(`  Container: ${this.detectedConfig.container.isContainer ? 'Yes' : 'No'}`);
    console.log(`  Cloud: ${this.detectedConfig.cloud.provider || 'None'}`);
    
    console.log('\n📋 Recommended Configuration:');
    console.log(`  Deployment: ${this.recommendedConfig.deployment.method}`);
    console.log(`  Model: ${this.recommendedConfig.models.primary}`);
    console.log(`  Port: ${this.recommendedConfig.network.port}`);
    console.log(`  GPU: ${this.recommendedConfig.performance.gpuEnabled ? 'Enabled' : 'Disabled'}`);
    
    console.log('\n✅ Configuration ready to apply');
    console.log('Run `await configurator.apply()` to apply these settings');
    
    return this.recommendedConfig;
  }

  // Export configuration for different formats
  async export(format = 'json') {
    const config = {
      detected: this.detectedConfig,
      recommended: this.recommendedConfig,
      applied: this.appliedConfig
    };
    
    switch (format) {
      case 'yaml':
        // Would use a YAML library here
        return JSON.stringify(config, null, 2);
        
      case 'env':
        return this.exportAsEnv(config);
        
      case 'docker':
        return this.exportAsDockerCompose(config);
        
      case 'kubernetes':
        return this.exportAsKubernetes(config);
        
      default:
        return JSON.stringify(config, null, 2);
    }
  }

  exportAsEnv(config) {
    const env = [];
    const rec = config.recommended;
    
    env.push(`PORT=${rec.network.port}`);
    env.push(`MODEL=${rec.models.primary}`);
    env.push(`GPU_ENABLED=${rec.performance.gpuEnabled}`);
    env.push(`CACHE_STRATEGY=${rec.performance.caching}`);
    
    return env.join('\n');
  }

  exportAsDockerCompose(config) {
    const rec = config.recommended;
    
    return `version: '3.8'

services:
  ollamamax:
    image: ollamamax/ollamamax:latest
    ports:
      - "${rec.network.port}:${rec.network.port}"
    environment:
      - MODEL=${rec.models.primary}
      - GPU_ENABLED=${rec.performance.gpuEnabled}
      - CACHE_STRATEGY=${rec.performance.caching}
    ${rec.performance.gpuEnabled ? `deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]` : ''}
`;
  }

  exportAsKubernetes(config) {
    const rec = config.recommended;
    
    return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollamamax
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ollamamax
  template:
    metadata:
      labels:
        app: ollamamax
    spec:
      containers:
      - name: ollamamax
        image: ollamamax/ollamamax:latest
        env:
        - name: MODEL
          value: "${rec.models.primary}"
        - name: GPU_ENABLED
          value: "${rec.performance.gpuEnabled}"
        ports:
        - containerPort: ${rec.network.port}
        ${rec.performance.gpuEnabled ? `resources:
          limits:
            nvidia.com/gpu: 1` : ''}
`;
  }
}

module.exports = AutoConfigurator;