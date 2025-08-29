# 🛠️ OllamaMax Installation & Configuration Scripts

This directory contains powerful scripts to help you install, configure, and integrate OllamaMax with ease.

## 📜 Available Scripts

### 🚀 install.sh
**Universal installation script for all platforms**

```bash
# Quick installation
curl -fsSL https://raw.githubusercontent.com/KhryptorGraphics/OllamaMax/main/scripts/install.sh | bash

# With options
./install.sh --enable-gpu --quick
./install.sh --version v1.2.0 --config ~/.ollamamax
./install.sh --force --skip-deps
```

**Features:**
- ✅ Cross-platform support (Linux, macOS, Windows WSL)
- ✅ Automatic dependency detection and installation
- ✅ GPU support detection and setup
- ✅ Shell integration and PATH management
- ✅ Comprehensive validation and error handling

### ⚙️ config-generator.sh
**Smart configuration generator with profiles**

```bash
# Interactive wizard
./config-generator.sh --interactive

# Profile-based generation
./config-generator.sh --profile production --security
./config-generator.sh --profile cluster --gpu
./config-generator.sh --profile development
```

**Available Profiles:**
- 🔧 **development** - Debug logging, minimal security
- 🏭 **production** - Optimized performance, security enabled
- 🌐 **cluster** - Multi-node distributed setup
- 💨 **edge** - Lightweight, minimal resources
- ⚡ **gpu** - GPU-accelerated workloads
- 🏢 **enterprise** - Full security, compliance, auditing

### 🔗 ollama-integration.sh
**Seamless integration with existing Ollama installations**

```bash
# Scan for existing Ollama
./ollama-integration.sh --scan

# Migration modes
./ollama-integration.sh --mode migrate    # Replace Ollama
./ollama-integration.sh --mode coexist    # Run alongside
./ollama-integration.sh --mode auto       # Automatic detection
```

**Integration Features:**
- 🔍 **Automatic Discovery** - Finds existing Ollama installations
- 📦 **Model Migration** - Preserves your downloaded models
- ⚙️ **Configuration Import** - Maintains your settings
- 🤝 **Coexistence Mode** - Run both systems simultaneously
- 💾 **Safe Backups** - Automatic data protection

## 🚀 Quick Start Examples

### Instant Setup
```bash
# Zero-configuration install and start
curl -fsSL https://install.ollamamax.com | bash
ollama-distributed quickstart
```

### Custom Development Setup
```bash
# Install with development profile
./install.sh --quick
./config-generator.sh --profile development --output dev-config.yaml
ollama-distributed start --config dev-config.yaml
```

### Production Deployment
```bash
# Install with security features
./install.sh --enable-gpu --version latest
./config-generator.sh --profile production --security --output prod-config.yaml
ollama-distributed validate --config prod-config.yaml --fix
ollama-distributed start --config prod-config.yaml
```

### Migrate from Ollama
```bash
# Seamlessly migrate from existing Ollama
./ollama-integration.sh --mode migrate --preserve-models
ollama-distributed start
```

## 🔧 Script Options Reference

### install.sh Options
| Option | Description | Example |
|--------|-------------|---------|
| `--version VER` | Install specific version | `--version v1.2.0` |
| `--dir DIR` | Installation directory | `--dir /opt/ollamamax` |
| `--config DIR` | Configuration directory | `--config ~/.ollamamax` |
| `--enable-gpu` | Enable GPU support | `--enable-gpu` |
| `--force` | Force reinstallation | `--force` |
| `--skip-deps` | Skip dependency installation | `--skip-deps` |
| `--quick` | Quick install with defaults | `--quick` |

### config-generator.sh Options
| Option | Description | Example |
|--------|-------------|---------|
| `--profile PROFILE` | Configuration profile | `--profile production` |
| `--output FILE` | Output configuration file | `--output config.yaml` |
| `--gpu` | Enable GPU support | `--gpu` |
| `--cluster` | Enable cluster mode | `--cluster` |
| `--security` | Enable security features | `--security` |
| `--interactive` | Interactive wizard | `--interactive` |

### ollama-integration.sh Options
| Option | Description | Example |
|--------|-------------|---------|
| `--mode MODE` | Integration mode | `--mode migrate` |
| `--ollama PATH` | Path to Ollama binary | `--ollama /usr/bin/ollama` |
| `--models-dir PATH` | Models directory | `--models-dir ~/.ollama/models` |
| `--preserve-models` | Keep existing models | `--preserve-models` |
| `--force` | Force integration | `--force` |
| `--scan` | Scan for installations | `--scan` |

## 🎯 Common Use Cases

### 1. First Time Installation
```bash
# Complete beginner setup
curl -fsSL https://install.ollamamax.com | bash
ollama-distributed quickstart
```

### 2. Development Environment  
```bash
./install.sh --quick
./config-generator.sh --profile development
ollama-distributed setup  # Interactive configuration
ollama-distributed start
```

### 3. Production Deployment
```bash
./install.sh --enable-gpu --version stable
./config-generator.sh --profile production --security --gpu
ollama-distributed validate --fix
ollama-distributed start
```

### 4. Cluster Setup
```bash
# Node 1 (Leader)
./config-generator.sh --profile cluster --output cluster-leader.yaml
ollama-distributed start --config cluster-leader.yaml --cluster-init

# Node 2+ (Followers)  
./config-generator.sh --profile cluster --output cluster-node.yaml
ollama-distributed join --config cluster-node.yaml --peer leader-ip:8080
```

### 5. Migrate from Ollama
```bash
./ollama-integration.sh --scan                    # See what you have
./ollama-integration.sh --mode migrate --backup   # Migrate safely
ollama-distributed start                          # Start OllamaMax
```

### 6. Side-by-Side Testing
```bash
./ollama-integration.sh --mode coexist   # Different ports
# Ollama: localhost:11434
# OllamaMax: localhost:8080
```

## 🛟 Troubleshooting

### Permission Issues
```bash
# Fix file permissions
sudo chmod +x scripts/*.sh
sudo chown -R $USER:$USER ~/.ollamamax

# Use custom directories
./install.sh --dir ~/ollamamax --config ~/ollamamax-config
```

### Port Conflicts
```bash
# Check what's using ports
sudo lsof -i :8080
sudo lsof -i :8081

# Use different ports
./config-generator.sh --interactive  # Choose custom ports
```

### Integration Problems
```bash
# Reset integration
./ollama-integration.sh --force --mode auto

# Restore from backup
./ollama-integration.sh --restore ~/backups/ollama_20231201
```

### Clean Reinstall
```bash
# Remove everything
rm -rf ~/.ollamamax /usr/local/bin/ollama-distributed

# Fresh install
./install.sh --force --quick
ollama-distributed quickstart
```

## 🔍 Advanced Features

### Environment Detection
Scripts automatically detect:
- ✅ Operating system and architecture
- ✅ Available memory and disk space  
- ✅ Existing Ollama installations
- ✅ GPU capabilities (CUDA, ROCm)
- ✅ Network port availability
- ✅ Package managers (apt, yum, brew)

### Security Features
- 🔐 **Automatic TLS certificate generation**
- 🛡️ **JWT secret generation**
- 🔑 **Permission management**
- 📋 **Security validation checks**
- 🔍 **Vulnerability scanning**

### Backup & Recovery
- 💾 **Automatic data backups before changes**
- 📋 **Backup manifests with restore instructions** 
- 🔄 **Rollback capabilities**
- ✅ **Data integrity verification**

## 📚 Additional Resources

- 📖 **Full Documentation**: https://docs.ollamamax.com
- 🎓 **Installation Tutorial**: https://docs.ollamamax.com/install
- 🔧 **Configuration Guide**: https://docs.ollamamax.com/config  
- 🤝 **Integration Guide**: https://docs.ollamamax.com/migrate
- 💬 **Community Support**: https://github.com/KhryptorGraphics/OllamaMax/discussions

---

**Need Help?** 🆘

Run any script with `--help` for detailed options, or visit our [GitHub Discussions](https://github.com/KhryptorGraphics/OllamaMax/discussions) for community support!