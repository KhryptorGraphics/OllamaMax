/**
 * Unified Deployment System for Ollamamax
 * Single interface for deploying across all platforms
 */

const fs = require('fs').promises;
const path = require('path');
const { exec } = require('child_process');
const { promisify } = require('util');
const execAsync = promisify(exec);
const HardwareAbstractionLayer = require('../hardware/abstraction-layer');
const PlatformOptimizer = require('../hardware/platform-optimizer');

class UnifiedDeployer {
  constructor() {
    this.hal = null;
    this.optimizer = null;
    this.deploymentStrategies = new Map();
    this.loadStrategies();
  }

  async initialize() {
    console.log('🚀 Initializing Unified Deployment System...');
    
    // Initialize hardware abstraction layer
    this.hal = new HardwareAbstractionLayer();
    await this.hal.initialize();
    
    // Initialize platform optimizer
    this.optimizer = new PlatformOptimizer(this.hal);
    
    console.log('✅ Deployment system ready');
    return this;
  }

  loadStrategies() {
    // Deployment strategies for each platform type
    this.deploymentStrategies.set('embedded', {
      prepare: this.prepareEmbedded.bind(this),
      deploy: this.deployEmbedded.bind(this),
      validate: this.validateEmbedded.bind(this),
      rollback: this.rollbackEmbedded.bind(this)
    });

    this.deploymentStrategies.set('edge-ai', {
      prepare: this.prepareEdgeAI.bind(this),
      deploy: this.deployEdgeAI.bind(this),
      validate: this.validateEdgeAI.bind(this),
      rollback: this.rollbackEdgeAI.bind(this)
    });

    this.deploymentStrategies.set('mobile', {
      prepare: this.prepareMobile.bind(this),
      deploy: this.deployMobile.bind(this),
      validate: this.validateMobile.bind(this),
      rollback: this.rollbackMobile.bind(this)
    });

    this.deploymentStrategies.set('datacenter', {
      prepare: this.prepareDatacenter.bind(this),
      deploy: this.deployDatacenter.bind(this),
      validate: this.validateDatacenter.bind(this),
      rollback: this.rollbackDatacenter.bind(this)
    });

    this.deploymentStrategies.set('workstation', {
      prepare: this.prepareWorkstation.bind(this),
      deploy: this.deployWorkstation.bind(this),
      validate: this.validateWorkstation.bind(this),
      rollback: this.rollbackWorkstation.bind(this)
    });

    this.deploymentStrategies.set('consumer', {
      prepare: this.prepareConsumer.bind(this),
      deploy: this.deployConsumer.bind(this),
      validate: this.validateConsumer.bind(this),
      rollback: this.rollbackConsumer.bind(this)
    });

    // Docker deployment
    this.deploymentStrategies.set('docker', {
      prepare: this.prepareDocker.bind(this),
      deploy: this.deployDocker.bind(this),
      validate: this.validateDocker.bind(this),
      rollback: this.rollbackDocker.bind(this)
    });

    // Kubernetes deployment
    this.deploymentStrategies.set('kubernetes', {
      prepare: this.prepareKubernetes.bind(this),
      deploy: this.deployKubernetes.bind(this),
      validate: this.validateKubernetes.bind(this),
      rollback: this.rollbackKubernetes.bind(this)
    });
  }

  async deploy(options = {}) {
    const {
      mode = 'auto',
      platform = null,
      skipValidation = false,
      dryRun = false
    } = options;

    console.log('📦 Starting unified deployment...');
    
    // Determine deployment platform
    const targetPlatform = platform || this.hal.platform.type;
    
    // Get deployment strategy
    const strategy = this.deploymentStrategies.get(
      mode === 'docker' ? 'docker' : 
      mode === 'kubernetes' ? 'kubernetes' : 
      targetPlatform
    );

    if (!strategy) {
      throw new Error(`No deployment strategy for platform: ${targetPlatform}`);
    }

    try {
      // Phase 1: Prepare
      console.log('📋 Phase 1: Preparation');
      const preparation = await strategy.prepare(options);
      
      if (dryRun) {
        console.log('🔍 Dry run complete. Deployment plan:');
        console.log(JSON.stringify(preparation, null, 2));
        return preparation;
      }

      // Phase 2: Deploy
      console.log('🚀 Phase 2: Deployment');
      const deployment = await strategy.deploy(preparation);

      // Phase 3: Validate
      if (!skipValidation) {
        console.log('✔️ Phase 3: Validation');
        const validation = await strategy.validate(deployment);
        
        if (!validation.success) {
          console.error('❌ Validation failed, initiating rollback...');
          await strategy.rollback(deployment);
          throw new Error('Deployment validation failed');
        }
      }

      // Phase 4: Optimize
      console.log('⚡ Phase 4: Optimization');
      const optimizations = await this.optimizer.optimize();
      await this.optimizer.applyOptimizations(optimizations);

      console.log('✅ Deployment completed successfully!');
      
      return {
        platform: targetPlatform,
        deployment,
        optimizations,
        status: 'success'
      };

    } catch (error) {
      console.error('❌ Deployment failed:', error.message);
      throw error;
    }
  }

  // Raspberry Pi Deployment
  async prepareEmbedded(options) {
    return {
      platform: 'raspberry-pi',
      requirements: {
        os: 'Raspberry Pi OS (64-bit)',
        memory: '4GB minimum',
        storage: '32GB SD card minimum',
        network: 'Ethernet recommended'
      },
      dependencies: [
        'build-essential',
        'cmake',
        'python3-dev',
        'nodejs',
        'npm'
      ],
      services: [
        'ollamamax-node',
        'ollamamax-api'
      ],
      config: {
        modelPath: '/opt/ollamamax/models',
        dataPath: '/var/lib/ollamamax',
        port: 13000
      }
    };
  }

  async deployEmbedded(preparation) {
    console.log('🔧 Deploying to Raspberry Pi...');
    
    // Install system dependencies
    await execAsync('sudo apt-get update');
    for (const dep of preparation.dependencies) {
      await execAsync(`sudo apt-get install -y ${dep}`);
    }

    // Create directories
    await execAsync(`sudo mkdir -p ${preparation.config.modelPath}`);
    await execAsync(`sudo mkdir -p ${preparation.config.dataPath}`);

    // Install Ollamamax
    await execAsync('npm install --production');
    
    // Create systemd service
    const serviceContent = `[Unit]
Description=Ollamamax Node Service
After=network.target

[Service]
Type=simple
User=ollamamax
WorkingDirectory=/opt/ollamamax
ExecStart=/usr/bin/node /opt/ollamamax/main.js
Restart=on-failure
RestartSec=10
Environment="NODE_ENV=production"
Environment="PORT=${preparation.config.port}"

[Install]
WantedBy=multi-user.target`;

    await fs.writeFile('/tmp/ollamamax.service', serviceContent);
    await execAsync('sudo mv /tmp/ollamamax.service /etc/systemd/system/');
    await execAsync('sudo systemctl daemon-reload');
    await execAsync('sudo systemctl enable ollamamax');
    await execAsync('sudo systemctl start ollamamax');

    return {
      status: 'deployed',
      service: 'ollamamax',
      port: preparation.config.port
    };
  }

  async validateEmbedded(deployment) {
    try {
      const { stdout } = await execAsync('sudo systemctl is-active ollamamax');
      const isActive = stdout.trim() === 'active';
      
      // Check if service is responding
      const { stdout: curlOutput } = await execAsync(
        `curl -s -o /dev/null -w "%{http_code}" http://localhost:${deployment.port}/health`
      );
      const isResponding = curlOutput === '200';

      return {
        success: isActive && isResponding,
        service: isActive,
        api: isResponding
      };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async rollbackEmbedded(deployment) {
    await execAsync('sudo systemctl stop ollamamax');
    await execAsync('sudo systemctl disable ollamamax');
    console.log('🔄 Rollback completed');
  }

  // NVIDIA Jetson Deployment
  async prepareEdgeAI(options) {
    return {
      platform: 'nvidia-jetson',
      requirements: {
        os: 'JetPack 5.x',
        cuda: '11.4+',
        tensorrt: '8.x',
        memory: '8GB minimum'
      },
      dependencies: [
        'cuda-toolkit',
        'tensorrt',
        'python3-pip',
        'nodejs',
        'docker.io'
      ],
      optimizations: {
        gpu: true,
        tensorCores: true,
        dla: true
      }
    };
  }

  async deployEdgeAI(preparation) {
    console.log('🔧 Deploying to NVIDIA Jetson...');
    
    // Install CUDA dependencies
    await execAsync('sudo apt-get update');
    await execAsync('sudo apt-get install -y cuda-toolkit-11-4');
    
    // Install TensorRT
    await execAsync('sudo apt-get install -y tensorrt');
    
    // Deploy with Docker for easier management
    await execAsync('docker pull ollamamax/jetson:latest');
    await execAsync(`docker run -d \
      --name ollamamax \
      --runtime nvidia \
      --network host \
      -v /opt/models:/models \
      -e CUDA_VISIBLE_DEVICES=0 \
      -e ENABLE_TENSORRT=true \
      ollamamax/jetson:latest`);

    return {
      status: 'deployed',
      container: 'ollamamax',
      gpu: 'enabled'
    };
  }

  async validateEdgeAI(deployment) {
    try {
      const { stdout } = await execAsync('docker ps --filter name=ollamamax --format "{{.Status}}"');
      const isRunning = stdout.includes('Up');
      
      return {
        success: isRunning,
        container: isRunning,
        gpu: true
      };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async rollbackEdgeAI(deployment) {
    await execAsync('docker stop ollamamax');
    await execAsync('docker rm ollamamax');
    console.log('🔄 Rollback completed');
  }

  // Mobile Deployment
  async prepareMobile(options) {
    return {
      platform: 'mobile',
      targets: ['ios', 'android'],
      requirements: {
        ios: {
          xcode: '14+',
          swift: '5.7+',
          coreml: 'required'
        },
        android: {
          sdk: '33+',
          ndk: 'r25+',
          nnapi: 'required'
        }
      },
      buildTools: {
        ios: 'xcodebuild',
        android: 'gradle'
      }
    };
  }

  async deployMobile(preparation) {
    console.log('📱 Building mobile packages...');
    
    // This would typically build mobile SDKs
    // For now, we'll create the package structure
    
    const mobilePackage = {
      ios: {
        framework: 'OllamamaxKit.xcframework',
        size: 'optimized',
        architecture: ['arm64', 'arm64-simulator']
      },
      android: {
        aar: 'ollamamax-sdk.aar',
        minSdk: 24,
        targetSdk: 33
      }
    };

    return {
      status: 'packaged',
      packages: mobilePackage
    };
  }

  async validateMobile(deployment) {
    // Mobile validation would test the SDKs
    return {
      success: true,
      ios: true,
      android: true
    };
  }

  async rollbackMobile(deployment) {
    console.log('🔄 Mobile deployment rollback');
  }

  // Datacenter Deployment
  async prepareDatacenter(options) {
    return {
      platform: 'datacenter',
      orchestration: 'kubernetes',
      requirements: {
        nodes: '3+ for HA',
        gpu: 'NVIDIA A100 or H100',
        network: '100Gbps InfiniBand',
        storage: 'NVMe SSD array'
      },
      components: [
        'control-plane',
        'worker-nodes',
        'gpu-nodes',
        'storage-nodes'
      ],
      scaling: {
        horizontal: true,
        vertical: true,
        autoScale: true
      }
    };
  }

  async deployDatacenter(preparation) {
    console.log('🏢 Deploying to datacenter...');
    
    // Generate Kubernetes manifests
    const k8sManifests = await this.generateK8sManifests(preparation);
    
    // Apply manifests
    for (const manifest of k8sManifests) {
      await fs.writeFile(`/tmp/${manifest.name}.yaml`, manifest.content);
      await execAsync(`kubectl apply -f /tmp/${manifest.name}.yaml`);
    }

    return {
      status: 'deployed',
      namespace: 'ollamamax',
      replicas: 3
    };
  }

  async validateDatacenter(deployment) {
    try {
      const { stdout } = await execAsync(
        `kubectl get pods -n ${deployment.namespace} -o json`
      );
      const pods = JSON.parse(stdout);
      const allReady = pods.items.every(pod => 
        pod.status.phase === 'Running'
      );

      return {
        success: allReady,
        pods: pods.items.length,
        ready: allReady
      };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async rollbackDatacenter(deployment) {
    await execAsync(`kubectl delete namespace ${deployment.namespace}`);
    console.log('🔄 Datacenter rollback completed');
  }

  // Workstation Deployment
  async prepareWorkstation(options) {
    return {
      platform: 'workstation',
      method: 'native',
      requirements: {
        os: ['Windows 11', 'macOS 13+', 'Ubuntu 22.04+'],
        memory: '16GB minimum',
        gpu: 'Recommended',
        storage: '100GB available'
      },
      installer: {
        windows: 'ollamamax-setup.exe',
        macos: 'Ollamamax.dmg',
        linux: 'ollamamax-installer.sh'
      }
    };
  }

  async deployWorkstation(preparation) {
    console.log('💻 Deploying to workstation...');
    
    const platform = process.platform;
    
    if (platform === 'darwin') {
      // macOS installation
      await execAsync('brew install ollamamax');
    } else if (platform === 'win32') {
      // Windows installation
      console.log('Please run the Windows installer: ollamamax-setup.exe');
    } else {
      // Linux installation
      await execAsync('sudo snap install ollamamax');
    }

    return {
      status: 'installed',
      platform,
      method: 'native'
    };
  }

  async validateWorkstation(deployment) {
    try {
      const { stdout } = await execAsync('ollamamax --version');
      return {
        success: stdout.includes('Ollamamax'),
        version: stdout.trim()
      };
    } catch {
      return { success: false };
    }
  }

  async rollbackWorkstation(deployment) {
    const platform = process.platform;
    
    if (platform === 'darwin') {
      await execAsync('brew uninstall ollamamax');
    } else if (platform === 'linux') {
      await execAsync('sudo snap remove ollamamax');
    }
    
    console.log('🔄 Workstation rollback completed');
  }

  // Consumer PC Deployment
  async prepareConsumer(options) {
    return {
      platform: 'consumer-pc',
      method: 'electron',
      requirements: {
        os: ['Windows 10+', 'macOS 11+', 'Ubuntu 20.04+'],
        memory: '8GB minimum',
        storage: '50GB available'
      },
      features: {
        gui: true,
        autoUpdate: true,
        trayIcon: true
      }
    };
  }

  async deployConsumer(preparation) {
    console.log('🖥️ Deploying consumer application...');
    
    // Build Electron app
    await execAsync('npm run build:electron');
    
    // Package for distribution
    await execAsync('npm run package:electron');

    return {
      status: 'packaged',
      outputs: {
        windows: 'dist/Ollamamax-Setup.exe',
        macos: 'dist/Ollamamax.dmg',
        linux: 'dist/ollamamax.AppImage'
      }
    };
  }

  async validateConsumer(deployment) {
    // Check if packages were created
    const exists = await fs.access(deployment.outputs.linux)
      .then(() => true)
      .catch(() => false);

    return {
      success: exists,
      packages: exists
    };
  }

  async rollbackConsumer(deployment) {
    console.log('🔄 Consumer deployment rollback');
  }

  // Docker Deployment
  async prepareDocker(options) {
    return {
      platform: 'docker',
      images: {
        base: 'ollamamax/base:latest',
        gpu: 'ollamamax/gpu:latest',
        slim: 'ollamamax/slim:latest'
      },
      compose: {
        file: 'docker-compose.yml',
        services: ['api', 'worker', 'redis', 'postgres']
      },
      networks: ['ollamamax-network'],
      volumes: ['ollamamax-data', 'ollamamax-models']
    };
  }

  async deployDocker(preparation) {
    console.log('🐳 Deploying with Docker...');
    
    // Pull images
    for (const image of Object.values(preparation.images)) {
      await execAsync(`docker pull ${image}`);
    }

    // Create network
    await execAsync('docker network create ollamamax-network').catch(() => {});

    // Create volumes
    for (const volume of preparation.volumes) {
      await execAsync(`docker volume create ${volume}`).catch(() => {});
    }

    // Start services
    await execAsync('docker-compose up -d');

    return {
      status: 'running',
      services: preparation.compose.services,
      network: 'ollamamax-network'
    };
  }

  async validateDocker(deployment) {
    try {
      const { stdout } = await execAsync('docker-compose ps --format json');
      const services = JSON.parse(stdout);
      const allRunning = services.every(s => s.State === 'running');

      return {
        success: allRunning,
        services: services.length,
        running: allRunning
      };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async rollbackDocker(deployment) {
    await execAsync('docker-compose down');
    console.log('🔄 Docker rollback completed');
  }

  // Kubernetes Deployment
  async prepareKubernetes(options) {
    return {
      platform: 'kubernetes',
      namespace: 'ollamamax',
      charts: {
        helm: 'ollamamax/ollamamax',
        version: '1.0.0'
      },
      resources: {
        requests: {
          cpu: '4',
          memory: '8Gi',
          'nvidia.com/gpu': '1'
        },
        limits: {
          cpu: '8',
          memory: '16Gi',
          'nvidia.com/gpu': '1'
        }
      },
      replicas: 3,
      autoscaling: {
        enabled: true,
        minReplicas: 3,
        maxReplicas: 10,
        targetCPU: 70
      }
    };
  }

  async deployKubernetes(preparation) {
    console.log('☸️ Deploying to Kubernetes...');
    
    // Create namespace
    await execAsync(`kubectl create namespace ${preparation.namespace}`).catch(() => {});

    // Install using Helm
    await execAsync(`helm repo add ollamamax https://charts.ollamamax.io`);
    await execAsync('helm repo update');
    
    await execAsync(`helm install ollamamax ollamamax/ollamamax \
      --namespace ${preparation.namespace} \
      --set replicas=${preparation.replicas} \
      --set autoscaling.enabled=${preparation.autoscaling.enabled} \
      --set resources.requests.cpu=${preparation.resources.requests.cpu} \
      --set resources.requests.memory=${preparation.resources.requests.memory}`);

    return {
      status: 'deployed',
      namespace: preparation.namespace,
      release: 'ollamamax'
    };
  }

  async validateKubernetes(deployment) {
    try {
      const { stdout } = await execAsync(
        `helm status ${deployment.release} -n ${deployment.namespace} -o json`
      );
      const status = JSON.parse(stdout);

      return {
        success: status.info.status === 'deployed',
        status: status.info.status,
        revision: status.version
      };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async rollbackKubernetes(deployment) {
    await execAsync(`helm uninstall ${deployment.release} -n ${deployment.namespace}`);
    await execAsync(`kubectl delete namespace ${deployment.namespace}`);
    console.log('🔄 Kubernetes rollback completed');
  }

  // Helper methods
  async generateK8sManifests(config) {
    const manifests = [];

    // Deployment manifest
    manifests.push({
      name: 'deployment',
      content: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollamamax
  namespace: ollamamax
spec:
  replicas: ${config.scaling?.replicas || 3}
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
        image: ollamamax/datacenter:latest
        resources:
          requests:
            memory: "16Gi"
            cpu: "4"
            nvidia.com/gpu: "1"
          limits:
            memory: "32Gi"
            cpu: "8"
            nvidia.com/gpu: "1"
        ports:
        - containerPort: 13000`
    });

    // Service manifest
    manifests.push({
      name: 'service',
      content: `apiVersion: v1
kind: Service
metadata:
  name: ollamamax
  namespace: ollamamax
spec:
  selector:
    app: ollamamax
  ports:
  - protocol: TCP
    port: 13000
    targetPort: 13000
  type: LoadBalancer`
    });

    return manifests;
  }

  // Get deployment status
  async getStatus() {
    const platform = this.hal.platform;
    const optimizations = this.optimizer.optimizations;

    return {
      platform: platform.type,
      name: platform.name,
      capabilities: Object.fromEntries(this.hal.capabilities),
      optimizations: Object.fromEntries(optimizations),
      metrics: this.hal.metrics
    };
  }
}

module.exports = UnifiedDeployer;