# CLI Reference

Complete reference for all Ollama Distributed CLI commands based on actual implementation.

## Main Command

```bash
ollama-distributed [command] [flags]
```

**Global Flags:**
- `-h, --help` - Help for any command
- `-v, --version` - Version information

## Core Commands

### quickstart

```bash
ollama-distributed quickstart [flags]
```

🚀 Instant setup with sensible defaults - Get OllamaMax running in 60 seconds with zero configuration.

**What it does:**
- Creates default configuration optimized for single-node deployment
- Downloads essential models (phi3, llama2-7b) if requested
- Starts the distributed node
- Opens web dashboard (optional)
- Provides usage examples

**Flags:**
- `--port int` - API server port (default 8080)
- `--no-models` - Skip downloading default models
- `--no-web` - Skip opening web dashboard

**Example:**
```bash
# Standard quickstart
ollama-distributed quickstart

# Custom port, skip models and web
ollama-distributed quickstart --port 8888 --no-models --no-web
```

### setup

```bash
ollama-distributed setup
```

⚙️ Interactive setup wizard for OllamaMax configuration.

**What it does:**
- Guides through configuring basic node settings
- Sets up network and clustering options
- Configures security settings
- Generates a configuration file
- Provides next steps for getting started

**Interactive Prompts:**
- Node name (default: ollama-node)
- API port (default: 8080)  
- Web port (default: 8081)
- GPU support (y/N)

**Example:**
```bash
ollama-distributed setup
# Follow interactive prompts
```

### start

```bash
ollama-distributed start [flags]
```

🏃 Start the OllamaMax node with specified configuration.

**Flags:**
- `-c, --config string` - Configuration file path

**Example:**
```bash
# Start with default config
ollama-distributed start

# Start with custom config
ollama-distributed start --config /path/to/config.yaml
```

### status

```bash
ollama-distributed status [flags]
```

🏥 Show comprehensive cluster health status with real-time health information.

**Flags:**
- `-o, --output string` - Output format: table, json, yaml (default "table")
- `-v, --verbose` - Show detailed metrics
- `-w, --watch` - Watch mode (refresh every 5s)

**Output includes:**
- Overall cluster status
- Node information (ID, status, role, uptime)
- Resource usage (when verbose)
- Model information (when verbose)  
- Network services (when verbose)
- Quick summary

**Example:**
```bash
# Basic status
ollama-distributed status

# Detailed status
ollama-distributed status --verbose

# Watch mode
ollama-distributed status --watch
```

### validate

```bash
ollama-distributed validate [flags]
```

🔍 Validate configuration and environment with comprehensive validation.

**Flags:**
- `--fix` - Attempt to fix common issues automatically
- `--quick` - Run only essential validation checks

**Validation checks:**
- Configuration file syntax
- API port availability
- System resources
- Directory permissions
- Network connectivity

**Example:**
```bash
# Basic validation
ollama-distributed validate

# Quick validation
ollama-distributed validate --quick

# Validate and fix issues
ollama-distributed validate --fix
```

## Utility Commands

### examples

```bash
ollama-distributed examples
```

💡 Show usage examples and common patterns.

**Shows examples for:**
- Quick Start
- Interactive Setup
- Start Node
- Check Status
- Download Model
- List Models
- Validate Config

### tutorial

```bash
ollama-distributed tutorial
```

📚 Interactive getting started tutorial with step-by-step guidance.

**Tutorial steps:**
1. Quick Start - `ollama-distributed quickstart`
2. Download Model - `ollama-distributed proxy pull phi3:mini`
3. Check Status - `ollama-distributed status`
4. Open Web UI - http://localhost:8081
5. Try API - `curl http://localhost:8080/health`

### troubleshoot

```bash
ollama-distributed troubleshoot
```

🔧 Diagnostic tools and common issue fixes.

**Diagnostic checks:**
- Service running status
- Port availability
- Disk space
- Memory availability
- Configuration validity

## Model Management (Proxy Commands)

### proxy pull

```bash
ollama-distributed proxy pull MODEL
```

Download a model to the distributed cluster.

**Arguments:**
- `MODEL` - Model name (e.g., llama2, phi3:mini, codellama)

**Example:**
```bash
ollama-distributed proxy pull llama2:7b
ollama-distributed proxy pull phi3:mini
```

### proxy list

```bash
ollama-distributed proxy list
```

List all available models with their status and size.

**Example output:**
```
🤖 Available Models
━━━━━━━━━━━━━━━━━━
phi3:mini       2.3GB    ✅ Ready
llama2:7b       3.8GB    ⏳ Downloading
codellama       3.8GB    💤 Available
```

## Shell Completion

Enable command completion for your shell:

### Bash
```bash
# Add to ~/.bashrc
echo 'source <(ollama-distributed completion bash)' >> ~/.bashrc
source ~/.bashrc
```

### Zsh  
```bash
# Add to ~/.zshrc
echo 'source <(ollama-distributed completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

### Fish
```bash
# Add to ~/.config/fish/config.fish
echo 'ollama-distributed completion fish | source' >> ~/.config/fish/config.fish
```

### PowerShell
```powershell
# Add to PowerShell profile
ollama-distributed completion powershell >> $PROFILE
```

## Configuration Directory

The CLI uses these directories:

- **Config Directory**: `~/.ollamamax/`
- **Data Directory**: `~/.ollamamax/data/`
- **Models Directory**: `~/.ollamamax/data/models/`
- **Logs Directory**: `~/.ollamamax/data/logs/`

## Environment Variables

The CLI respects these environment variables:

- `OLLAMA_HOST` - Default API host (default: localhost)
- `OLLAMA_PORT` - Default API port (default: 8080)
- `OLLAMA_CONFIG_DIR` - Config directory path
- `OLLAMA_DATA_DIR` - Data directory path
- `OLLAMA_DEBUG` - Enable debug logging
- `OLLAMA_ADMIN_TOKEN` - Admin authentication token

## Exit Codes

The CLI uses these exit codes:

- `0` - Success
- `1` - General error
- `2` - Configuration error
- `3` - Network error
- `4` - Permission error
- `5` - Resource error

## Real Usage Examples

### Development Setup
```bash
# Quick development setup
ollama-distributed quickstart
ollama-distributed proxy pull phi3:mini
ollama-distributed status --verbose
```

### Production Setup
```bash
# Production setup with custom config
ollama-distributed setup
# ... follow interactive prompts ...
ollama-distributed validate --fix
ollama-distributed start --config ~/.ollamamax/config.yaml
```

### Monitoring
```bash
# Continuous monitoring
ollama-distributed status --watch

# Get detailed system info
ollama-distributed status --verbose --output json
```

### Troubleshooting
```bash
# Diagnose issues
ollama-distributed validate --quick
ollama-distributed troubleshoot

# Fix common issues
ollama-distributed validate --fix
```

### Model Management
```bash
# Download and check models
ollama-distributed proxy pull llama2:7b
ollama-distributed proxy list
ollama-distributed status --verbose
```

## Command Output

All commands provide:
- **Clear status indicators** (✅ ❌ ⏳ 💤)
- **Professional formatting** with Unicode symbols
- **Consistent color coding** (when terminal supports it)
- **Structured information** with clear sections
- **Actionable next steps** where appropriate

## Getting Help

For any command, use the `--help` flag:

```bash
ollama-distributed --help
ollama-distributed quickstart --help
ollama-distributed status --help
```

The help system provides:
- Command description and usage
- Available flags and options
- Practical examples
- Related commands