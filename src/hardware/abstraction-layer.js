/**
 * Hardware Abstraction Layer (HAL) for Ollamamax
 * Provides unified interface across diverse hardware platforms
 */

const os = require('os');
const fs = require('fs');
const { exec } = require('child_process');
const { promisify } = require('util');
const execAsync = promisify(exec);

class HardwareAbstractionLayer {
  constructor() {
    this.platform = null;
    this.capabilities = new Map();
    this.optimizations = new Map();
    this.metrics = {
      cpu: {},
      memory: {},
      gpu: {},
      network: {}
    };
  }

  async initialize() {
    console.log('🔧 Initializing Hardware Abstraction Layer...');
    
    await this.detectPlatform();
    await this.detectCapabilities();
    await this.loadOptimizations();
    await this.calibratePerformance();
    
    console.log(`✅ HAL initialized for ${this.platform.name}`);
    return this;
  }

  async detectPlatform() {
    const platform = {
      os: os.platform(),
      arch: os.arch(),
      cpus: os.cpus(),
      memory: os.totalmem(),
      type: 'unknown',
      name: 'Unknown Platform'
    };

    // Detect platform type
    if (await this.isRaspberryPi()) {
      platform.type = 'embedded';
      platform.name = 'Raspberry Pi';
      platform.constraints = {
        memory: 'limited',
        compute: 'edge',
        power: 'low'
      };
    } else if (await this.isNvidiaJetson()) {
      platform.type = 'edge-ai';
      platform.name = 'NVIDIA Jetson';
      platform.constraints = {
        memory: 'moderate',
        compute: 'gpu-accelerated',
        power: 'moderate'
      };
    } else if (await this.isMobileDevice()) {
      platform.type = 'mobile';
      platform.name = 'Mobile Device';
      platform.constraints = {
        memory: 'limited',
        compute: 'mobile',
        power: 'battery'
      };
    } else if (await this.isGPUServer()) {
      platform.type = 'datacenter';
      platform.name = 'GPU Server';
      platform.constraints = {
        memory: 'abundant',
        compute: 'high-performance',
        power: 'unlimited'
      };
    } else if (platform.memory > 32 * 1024 * 1024 * 1024) { // > 32GB RAM
      platform.type = 'workstation';
      platform.name = 'Workstation';
      platform.constraints = {
        memory: 'high',
        compute: 'standard',
        power: 'standard'
      };
    } else {
      platform.type = 'consumer';
      platform.name = 'Consumer PC';
      platform.constraints = {
        memory: 'moderate',
        compute: 'standard',
        power: 'standard'
      };
    }

    this.platform = platform;
    return platform;
  }

  async detectCapabilities() {
    // CPU capabilities
    this.capabilities.set('cpu', {
      cores: os.cpus().length,
      speed: os.cpus()[0].speed,
      architecture: os.arch(),
      features: await this.detectCPUFeatures()
    });

    // Memory capabilities
    this.capabilities.set('memory', {
      total: os.totalmem(),
      available: os.freemem(),
      swappiness: await this.getSwappiness(),
      hugepages: await this.checkHugepages()
    });

    // GPU capabilities
    this.capabilities.set('gpu', await this.detectGPU());

    // Network capabilities
    this.capabilities.set('network', {
      interfaces: os.networkInterfaces(),
      bandwidth: await this.estimateBandwidth(),
      latency: await this.measureLatency()
    });

    // Storage capabilities
    this.capabilities.set('storage', {
      type: await this.detectStorageType(),
      speed: await this.benchmarkStorage(),
      available: await this.getAvailableStorage()
    });

    return this.capabilities;
  }

  async loadOptimizations() {
    const platformOptimizations = {
      'embedded': {
        modelQuantization: 'int8',
        batchSize: 1,
        cacheStrategy: 'aggressive',
        memoryPool: 'limited',
        threading: 'conservative'
      },
      'edge-ai': {
        modelQuantization: 'int8-fp16',
        batchSize: 4,
        cacheStrategy: 'balanced',
        memoryPool: 'dynamic',
        threading: 'gpu-optimized'
      },
      'mobile': {
        modelQuantization: 'int8',
        batchSize: 1,
        cacheStrategy: 'minimal',
        memoryPool: 'adaptive',
        threading: 'power-aware'
      },
      'datacenter': {
        modelQuantization: 'fp16-fp32',
        batchSize: 32,
        cacheStrategy: 'distributed',
        memoryPool: 'unlimited',
        threading: 'aggressive'
      },
      'workstation': {
        modelQuantization: 'fp16',
        batchSize: 8,
        cacheStrategy: 'standard',
        memoryPool: 'large',
        threading: 'balanced'
      },
      'consumer': {
        modelQuantization: 'int8-fp16',
        batchSize: 2,
        cacheStrategy: 'standard',
        memoryPool: 'moderate',
        threading: 'conservative'
      }
    };

    const opts = platformOptimizations[this.platform.type] || platformOptimizations['consumer'];
    
    for (const [key, value] of Object.entries(opts)) {
      this.optimizations.set(key, value);
    }

    // Apply GPU-specific optimizations
    const gpu = this.capabilities.get('gpu');
    if (gpu && gpu.available) {
      if (gpu.vendor === 'nvidia') {
        this.optimizations.set('cudaEnabled', true);
        this.optimizations.set('tensorCores', gpu.tensorCores || false);
      } else if (gpu.vendor === 'amd') {
        this.optimizations.set('rocmEnabled', true);
      } else if (gpu.vendor === 'apple') {
        this.optimizations.set('metalEnabled', true);
      }
    }

    return this.optimizations;
  }

  async calibratePerformance() {
    console.log('📊 Calibrating performance metrics...');
    
    // CPU benchmark
    this.metrics.cpu = await this.benchmarkCPU();
    
    // Memory benchmark
    this.metrics.memory = await this.benchmarkMemory();
    
    // GPU benchmark (if available)
    if (this.capabilities.get('gpu')?.available) {
      this.metrics.gpu = await this.benchmarkGPU();
    }
    
    // Network benchmark
    this.metrics.network = await this.benchmarkNetwork();
    
    // Calculate composite score
    this.metrics.composite = this.calculateCompositeScore();
    
    return this.metrics;
  }

  // Platform detection helpers
  async isRaspberryPi() {
    try {
      const { stdout } = await execAsync('cat /proc/device-tree/model 2>/dev/null');
      return stdout.toLowerCase().includes('raspberry pi');
    } catch {
      return false;
    }
  }

  async isNvidiaJetson() {
    try {
      const { stdout } = await execAsync('cat /proc/device-tree/model 2>/dev/null');
      return stdout.toLowerCase().includes('jetson');
    } catch {
      return false;
    }
  }

  async isMobileDevice() {
    // Check for Android or iOS indicators
    try {
      const { stdout } = await execAsync('uname -a');
      return stdout.toLowerCase().includes('android') || 
             stdout.toLowerCase().includes('darwin');
    } catch {
      return false;
    }
  }

  async isGPUServer() {
    const gpu = await this.detectGPU();
    return gpu.available && gpu.memory > 16 * 1024 * 1024 * 1024; // > 16GB VRAM
  }

  // Capability detection helpers
  async detectCPUFeatures() {
    const features = [];
    try {
      if (os.arch() === 'x64' || os.arch() === 'x86') {
        const { stdout } = await execAsync('cat /proc/cpuinfo | grep flags | head -1');
        if (stdout.includes('avx2')) features.push('avx2');
        if (stdout.includes('avx512')) features.push('avx512');
        if (stdout.includes('sse4')) features.push('sse4');
        if (stdout.includes('fma')) features.push('fma');
      } else if (os.arch() === 'arm64') {
        features.push('neon');
        const { stdout } = await execAsync('cat /proc/cpuinfo | grep Features | head -1');
        if (stdout.includes('sve')) features.push('sve');
      }
    } catch {}
    return features;
  }

  async detectGPU() {
    const gpu = {
      available: false,
      vendor: null,
      model: null,
      memory: 0,
      compute: null
    };

    try {
      // Try NVIDIA
      const { stdout: nvidiaInfo } = await execAsync('nvidia-smi --query-gpu=name,memory.total --format=csv,noheader 2>/dev/null');
      if (nvidiaInfo) {
        const [model, memory] = nvidiaInfo.trim().split(',');
        gpu.available = true;
        gpu.vendor = 'nvidia';
        gpu.model = model.trim();
        gpu.memory = parseInt(memory) * 1024 * 1024; // Convert MB to bytes
        
        // Check for tensor cores
        if (model.includes('RTX') || model.includes('A100') || model.includes('V100')) {
          gpu.tensorCores = true;
        }
        
        return gpu;
      }
    } catch {}

    try {
      // Try AMD
      const { stdout: amdInfo } = await execAsync('rocm-smi --showmeminfo vram --csv 2>/dev/null');
      if (amdInfo) {
        gpu.available = true;
        gpu.vendor = 'amd';
        return gpu;
      }
    } catch {}

    try {
      // Try Apple Silicon
      const { stdout: appleInfo } = await execAsync('system_profiler SPDisplaysDataType 2>/dev/null');
      if (appleInfo && appleInfo.includes('Apple')) {
        gpu.available = true;
        gpu.vendor = 'apple';
        gpu.model = 'Apple Silicon GPU';
        return gpu;
      }
    } catch {}

    return gpu;
  }

  async getSwappiness() {
    try {
      const { stdout } = await execAsync('cat /proc/sys/vm/swappiness 2>/dev/null');
      return parseInt(stdout.trim());
    } catch {
      return 60; // Default Linux swappiness
    }
  }

  async checkHugepages() {
    try {
      const { stdout } = await execAsync('cat /proc/meminfo | grep HugePages_Total 2>/dev/null');
      const match = stdout.match(/HugePages_Total:\s+(\d+)/);
      return match ? parseInt(match[1]) > 0 : false;
    } catch {
      return false;
    }
  }

  async detectStorageType() {
    try {
      const { stdout } = await execAsync('lsblk -d -o name,rota | grep "\\s0$" 2>/dev/null');
      return stdout ? 'ssd' : 'hdd';
    } catch {
      return 'unknown';
    }
  }

  async benchmarkStorage() {
    try {
      const testFile = '/tmp/ollamamax_storage_test';
      const size = 100; // MB
      
      const start = Date.now();
      await execAsync(`dd if=/dev/zero of=${testFile} bs=1M count=${size} 2>/dev/null`);
      const writeTime = Date.now() - start;
      
      const readStart = Date.now();
      await execAsync(`dd if=${testFile} of=/dev/null bs=1M 2>/dev/null`);
      const readTime = Date.now() - readStart;
      
      await execAsync(`rm -f ${testFile}`);
      
      return {
        write: (size * 1024 * 1024) / (writeTime / 1000), // bytes/sec
        read: (size * 1024 * 1024) / (readTime / 1000) // bytes/sec
      };
    } catch {
      return { write: 0, read: 0 };
    }
  }

  async getAvailableStorage() {
    try {
      const { stdout } = await execAsync('df -B1 / | tail -1');
      const parts = stdout.trim().split(/\s+/);
      return parseInt(parts[3]); // Available bytes
    } catch {
      return 0;
    }
  }

  async estimateBandwidth() {
    // Simple bandwidth estimation
    return 1000 * 1000 * 100; // Default 100 Mbps
  }

  async measureLatency() {
    try {
      const { stdout } = await execAsync('ping -c 1 8.8.8.8 | grep "time=" 2>/dev/null');
      const match = stdout.match(/time=([\d.]+)/);
      return match ? parseFloat(match[1]) : 100;
    } catch {
      return 100; // Default 100ms
    }
  }

  // Benchmark methods
  async benchmarkCPU() {
    console.log('🔬 Benchmarking CPU...');
    const start = Date.now();
    let operations = 0;
    
    // Simple CPU benchmark
    while (Date.now() - start < 1000) {
      Math.sqrt(Math.random());
      operations++;
    }
    
    return {
      operationsPerSecond: operations,
      cores: os.cpus().length,
      speed: os.cpus()[0].speed
    };
  }

  async benchmarkMemory() {
    console.log('🔬 Benchmarking memory...');
    const size = 100 * 1024 * 1024; // 100MB
    const buffer = Buffer.alloc(size);
    
    const writeStart = Date.now();
    for (let i = 0; i < size; i += 4096) {
      buffer.writeInt32LE(i, i);
    }
    const writeTime = Date.now() - writeStart;
    
    const readStart = Date.now();
    let sum = 0;
    for (let i = 0; i < size; i += 4096) {
      sum += buffer.readInt32LE(i);
    }
    const readTime = Date.now() - readStart;
    
    return {
      writeBandwidth: size / (writeTime / 1000),
      readBandwidth: size / (readTime / 1000),
      available: os.freemem(),
      total: os.totalmem()
    };
  }

  async benchmarkGPU() {
    console.log('🔬 Benchmarking GPU...');
    // Placeholder for GPU benchmarking
    // Would use CUDA/ROCm/Metal specific benchmarks
    return {
      compute: 'high',
      memory: this.capabilities.get('gpu').memory,
      bandwidth: 500 * 1024 * 1024 * 1024 // 500 GB/s estimate
    };
  }

  async benchmarkNetwork() {
    console.log('🔬 Benchmarking network...');
    return {
      bandwidth: await this.estimateBandwidth(),
      latency: await this.measureLatency()
    };
  }

  calculateCompositeScore() {
    const weights = {
      cpu: 0.3,
      memory: 0.2,
      gpu: 0.3,
      network: 0.1,
      storage: 0.1
    };
    
    let score = 0;
    
    // Normalize and weight each metric
    score += (this.metrics.cpu.operationsPerSecond / 1000000) * weights.cpu;
    score += (this.metrics.memory.available / os.totalmem()) * weights.memory;
    
    if (this.metrics.gpu.compute) {
      score += (this.metrics.gpu.compute === 'high' ? 1 : 0.5) * weights.gpu;
    }
    
    score += Math.min(1, this.metrics.network.bandwidth / (1000 * 1000 * 1000)) * weights.network;
    
    return Math.min(10, score * 10); // Scale to 0-10
  }

  // Public API methods
  getOptimalBatchSize() {
    return this.optimizations.get('batchSize') || 1;
  }

  getQuantizationLevel() {
    return this.optimizations.get('modelQuantization') || 'int8';
  }

  getCacheStrategy() {
    return this.optimizations.get('cacheStrategy') || 'standard';
  }

  getThreadingStrategy() {
    return this.optimizations.get('threading') || 'conservative';
  }

  shouldUseGPU() {
    return this.capabilities.get('gpu')?.available || false;
  }

  getMemoryLimit() {
    const total = os.totalmem();
    const poolType = this.optimizations.get('memoryPool');
    
    const limits = {
      'limited': total * 0.3,
      'adaptive': total * 0.5,
      'moderate': total * 0.6,
      'large': total * 0.8,
      'dynamic': total * 0.7,
      'unlimited': total * 0.9
    };
    
    return limits[poolType] || total * 0.5;
  }

  getPlatformInfo() {
    return {
      platform: this.platform,
      capabilities: Object.fromEntries(this.capabilities),
      optimizations: Object.fromEntries(this.optimizations),
      metrics: this.metrics
    };
  }

  // Adaptive optimization based on runtime metrics
  async adaptOptimizations(runtimeMetrics) {
    console.log('🔄 Adapting optimizations based on runtime metrics...');
    
    // Adjust batch size based on memory pressure
    if (runtimeMetrics.memoryPressure > 0.8) {
      const currentBatch = this.optimizations.get('batchSize');
      this.optimizations.set('batchSize', Math.max(1, Math.floor(currentBatch / 2)));
    }
    
    // Adjust threading based on CPU usage
    if (runtimeMetrics.cpuUsage > 0.9) {
      this.optimizations.set('threading', 'conservative');
    } else if (runtimeMetrics.cpuUsage < 0.5) {
      this.optimizations.set('threading', 'aggressive');
    }
    
    // Adjust cache strategy based on hit rate
    if (runtimeMetrics.cacheHitRate < 0.3) {
      this.optimizations.set('cacheStrategy', 'minimal');
    } else if (runtimeMetrics.cacheHitRate > 0.7) {
      this.optimizations.set('cacheStrategy', 'aggressive');
    }
    
    return this.optimizations;
  }
}

module.exports = HardwareAbstractionLayer;