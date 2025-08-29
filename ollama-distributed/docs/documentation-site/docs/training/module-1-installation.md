# Module 1: Installation and Setup

**Duration**: 10 minutes  
**Objective**: Get Ollama Distributed installed and running on your system

Welcome to your first hands-on module! In this session, you'll install Ollama Distributed and verify everything is working correctly.

## 🎯 What You'll Learn

By the end of this module, you will:
- ✅ Build Ollama Distributed from source
- ✅ Verify your installation is working
- ✅ Run your first CLI command
- ✅ Understand the basic directory structure

## 📋 Prerequisites Check

Before we start, let's verify your system is ready:

### System Requirements
```bash
# Check Go version (need 1.21+)
go version

# Check available disk space (need 2GB+)
df -h ~

# Check memory (need 4GB+ recommended)
free -h
```

**✅ Checkpoint 1**: Confirm you have Go 1.21+ and adequate disk space.

## 🛠️ Installation Process

### Step 1: Get the Source Code

If you don't already have the project:

```bash
# Navigate to your workspace
cd ~/workspace

# Check if project exists
ls -la ollamamax/
```

The project should already be available at `/home/kp/ollamamax/`.

### Step 2: Build the CLI Tool

Let's build the main CLI tool:

```bash
# Navigate to the project
cd /home/kp/ollamamax

# Build the CLI tool
go build -o bin/ollama-distributed ./cmd/ollama-distributed

# Verify the build succeeded
ls -la bin/ollama-distributed
```

**Expected Output:**
```
-rwxr-xr-x 1 user user 3600000 Aug 28 01:30 bin/ollama-distributed
```

**✅ Checkpoint 2**: Confirm the binary was created successfully.

### Step 3: Test the Installation

Let's verify the CLI tool works:

```bash
# Test the CLI help
./bin/ollama-distributed --help

# Check version
./bin/ollama-distributed --version
```

**Expected Output:**
```
🚀 OllamaMax - Enterprise Distributed AI Platform

A distributed, enterprise-grade version of Ollama that transforms the single-node
architecture into a horizontally scalable, fault-tolerant platform.

Features:
  🌐 Distributed AI model serving across multiple nodes
  🔒 Enterprise-grade security with JWT authentication
  📊 Real-time performance monitoring and optimization
  🎨 Beautiful web interface for easy management
  ⚡ Automatic load balancing and failover
  🔄 Seamless model distribution and synchronization
```

**✅ Checkpoint 3**: CLI help displays properly and shows all available commands.

### Step 4: Explore Available Commands

Let's see what commands are available:

```bash
# List all available commands
./bin/ollama-distributed --help
```

You should see these main commands:
- `quickstart` - 🚀 Instant setup with sensible defaults
- `setup` - ⚙️ Interactive setup wizard
- `start` - 🏃 Start the OllamaMax node
- `status` - 🏥 Show comprehensive cluster health status
- `validate` - 🔍 Validate configuration and environment
- `examples` - 💡 Show usage examples and common patterns
- `tutorial` - 📚 Interactive getting started tutorial
- `troubleshoot` - 🔧 Diagnostic tools and common issue fixes
- `proxy` - 🔗 Model management and proxy operations

**✅ Checkpoint 4**: All expected commands are listed and display help text.

## 🧪 Hands-On Exercise 1: First Run

Let's test the validate command to ensure everything is working:

```bash
# Run validation to check your system
./bin/ollama-distributed validate --quick
```

**Expected Output:**
```
🔍 OllamaMax Configuration Validation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Configuration file syntax: passed
✅ API port availability: passed
✅ System resources: passed
✅ Directory permissions: passed
✅ Network connectivity: passed

📊 Validation Summary
━━━━━━━━━━━━━━━━━━━
✅ All validations passed - ready to start!
```

**📝 Exercise Questions:**
1. What validation checks are performed?
2. What would happen if a port was already in use?
3. How would you get more detailed validation information?

**💡 Answers:**
1. Five checks: config syntax, port availability, system resources, directory permissions, network connectivity
2. The validation would fail and show which port is blocked
3. Run `./bin/ollama-distributed validate --help` to see more options

**✅ Checkpoint 5**: Validation passes successfully.

## 🧪 Hands-On Exercise 2: Directory Structure

Let's explore what gets created when you run Ollama Distributed:

```bash
# Check the current directory structure
ls -la

# Look at the bin directory
ls -la bin/

# Check what directories would be created (without actually running)
./bin/ollama-distributed quickstart --help
```

**Learning Point**: Ollama Distributed creates a `~/.ollamamax/` directory for configuration and data.

## 🔧 Troubleshooting Common Issues

### Issue 1: Build Fails
**Problem**: `go build` command fails
**Solution**:
```bash
# Ensure you're in the right directory
pwd
# Should show: /home/kp/ollamamax

# Check Go version
go version
# Should be 1.21 or higher

# Try building with verbose output
go build -v -o bin/ollama-distributed ./cmd/ollama-distributed
```

### Issue 2: Permission Denied
**Problem**: Cannot execute the binary
**Solution**:
```bash
# Make the binary executable
chmod +x bin/ollama-distributed

# Try running again
./bin/ollama-distributed --help
```

### Issue 3: Command Not Found
**Problem**: Binary seems to build but doesn't run
**Solution**:
```bash
# Use the full path
/home/kp/ollamamax/bin/ollama-distributed --help

# Or add to PATH temporarily
export PATH=$PATH:/home/kp/ollamamax/bin
ollama-distributed --help
```

## 📊 Module 1 Assessment

### Knowledge Check ✋
Answer these questions to test your understanding:

1. **Q**: What command would you use to validate your installation?
   **A**: `./bin/ollama-distributed validate --quick`

2. **Q**: Where does Ollama Distributed store its configuration files?
   **A**: `~/.ollamamax/` directory

3. **Q**: What are the minimum system requirements?
   **A**: Go 1.21+, 4GB RAM, 2GB disk space

4. **Q**: How do you see all available commands?
   **A**: `./bin/ollama-distributed --help`

### Practical Check ✋
Verify you can complete these tasks:

- [ ] Build the binary successfully
- [ ] Run `--help` and see command list  
- [ ] Execute `validate --quick` with all checks passing
- [ ] Understand the directory structure

## 🎉 Module 1 Complete!

**Congratulations!** You have successfully:

✅ **Installed** Ollama Distributed from source  
✅ **Verified** your installation is working correctly  
✅ **Tested** basic CLI functionality  
✅ **Understood** the project structure  

## 📚 What's Next?

You're now ready for **Module 2: Node Configuration** where you'll:
- Configure your first node
- Use the interactive setup wizard
- Customize settings for your environment
- Learn about different configuration profiles

**Time to continue:** [Module 2: Node Configuration →](./module-2-configuration.md)

## 💡 Pro Tips

1. **Save the Path**: Consider adding the bin directory to your PATH for easier access
2. **Keep Learning**: The `--help` flag works with every command for detailed information
3. **Stay Updated**: The project is actively developed, so features may evolve
4. **Practice**: Try different commands and explore the CLI interface

---

**Module 1 Status**: ✅ Complete  
**Next Module**: [Node Configuration →](./module-2-configuration.md)  
**Total Progress**: 1/5 modules (20%)