# Module 2: Node Configuration

**Duration**: 10 minutes  
**Objective**: Configure your first node with proper settings and understand the configuration system

Welcome to Module 2! Now that you have Ollama Distributed installed, it's time to configure your first node and understand how the configuration system works.

## 🎯 What You'll Learn

By the end of this module, you will:
- ✅ Use the interactive setup wizard
- ✅ Understand configuration file structure
- ✅ Customize settings for your environment
- ✅ Validate your configuration
- ✅ Learn about different configuration profiles

## 🛠️ Interactive Configuration

### Step 1: Launch the Setup Wizard

Let's use the built-in setup wizard to configure your node:

```bash
# Navigate to your project directory
cd /home/kp/ollamamax

# Launch the interactive setup wizard
./bin/ollama-distributed setup
```

**What happens next:**
The wizard will ask you several questions. Here are the recommended responses for learning:

**✅ Checkpoint 1**: Setup wizard launches and prompts for input.

### Step 2: Follow the Interactive Prompts

You'll see prompts like this:

```
⚙️ OllamaMax Interactive Setup
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Recommended Responses:**
- **Node name**: `training-node` (or your preferred name)
- **API port**: `8080` (default)
- **Web port**: `8081` (default)
- **Enable GPU support**: `N` (unless you have a GPU)

**Expected Output:**
```
📝 Configuration Summary:
   Node: training-node
   API Port: 8080
   Web Port: 8081
   GPU: false

✅ Setup complete! Configuration saved.

Next steps:
  1. Start: ollama-distributed start
  2. Status: ollama-distributed status
```

**✅ Checkpoint 2**: Interactive setup completes with configuration summary.

## 🧪 Hands-On Exercise 1: Quick Configuration

Let's also try the quickstart configuration:

```bash
# Generate a quickstart configuration
./bin/ollama-distributed quickstart --no-models --no-web
```

This creates a basic configuration optimized for single-node deployment.

**Expected Output:**
```
🚀 OllamaMax QuickStart
━━━━━━━━━━━━━━━━━━━━━
Getting you up and running in 60 seconds...

🔍 Validating environment...
✅ Environment ready
⚙️  Creating default configuration...
✅ Configuration created
📁 Setting up directories...
✅ Directories ready
🚀 Starting OllamaMax node...
✅ Node started

🎉 QuickStart Complete!
```

**✅ Checkpoint 3**: Quickstart configuration creates and starts a node successfully.

## 📁 Understanding Configuration Files

### Step 3: Explore the Configuration Directory

Let's examine what was created:

```bash
# Check the configuration directory
ls -la ~/.ollamamax/

# Look at the configuration file
ls -la ~/.ollamamax/*.yaml
```

**Expected Structure:**
```
~/.ollamamax/
├── quickstart-config.yaml
├── data/
│   ├── models/
│   └── logs/
```

**✅ Checkpoint 4**: Configuration directory and files are created.

### Step 4: Examine Configuration Content

Let's look at the configuration file structure:

```bash
# View the generated configuration
cat ~/.ollamamax/quickstart-config.yaml
```

**Expected Content:**
```yaml
# OllamaMax QuickStart Configuration
node:
  id: "quickstart-node"
  name: "quickstart-node"
  data_dir: "/home/user/.ollamamax/data"

api:
  host: "0.0.0.0"
  port: 8080

web:
  enabled: true
  port: 8081

models:
  store_path: "/home/user/.ollamamax/data/models"
  auto_cleanup: true

performance:
  max_concurrency: 4
  gpu_enabled: false
```

**📝 Learning Points:**
1. Configuration uses YAML format
2. Each section controls different aspects of the system
3. Paths are automatically set based on your home directory
4. GPU detection is automatic but can be overridden

## 🧪 Hands-On Exercise 2: Configuration Profiles

Let's explore different configuration profiles using the configuration generator:

```bash
# Check what configuration scripts are available
ls -la /home/kp/ollamamax/scripts/

# Generate a development configuration
/home/kp/ollamamax/scripts/config-generator.sh --profile development --output dev-config.yaml

# View the generated config
head -20 dev-config.yaml
```

**Expected Output:**
```yaml
# OllamaMax Development Configuration
# Generated on: 2025-08-28
# Profile: development
# Features: Debug logging, single node, minimal security

node:
  id: "dev-node-001"
  name: "Development Node"
  environment: "development"
  data_dir: "./data"

api:
  host: "127.0.0.1"
  port: 8080
  enable_tls: false

logging:
  level: "debug"
  format: "text"
```

**✅ Checkpoint 5**: Different configuration profiles can be generated successfully.

## 🔍 Configuration Validation

### Step 5: Validate Your Configuration

Let's validate the configuration we created:

```bash
# Validate the current configuration
./bin/ollama-distributed validate

# Run a more comprehensive validation
./bin/ollama-distributed validate --fix
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

**✅ Checkpoint 6**: Configuration validation passes all checks.

## 🧪 Hands-On Exercise 3: Custom Configuration

Let's create a custom configuration file:

```bash
# Create a custom config file
cat > custom-training-config.yaml << EOF
node:
  id: "custom-training-node"
  name: "My Training Node"
  environment: "learning"

api:
  host: "0.0.0.0"
  port: 8082

web:
  enabled: true
  port: 8083

models:
  store_path: "./training-models"
  max_cache_size: "5GB"

performance:
  max_concurrency: 2
  memory_limit: "4GB"

logging:
  level: "info"
  format: "text"
EOF

# Validate the custom configuration
echo "Validating custom configuration..."
# Note: The current validate command doesn't take a config file parameter
# This is an area for future enhancement
```

**📝 Learning Point**: You can create custom configuration files, though the current CLI validate command needs enhancement to accept custom config paths.

## 🔧 Troubleshooting Configuration Issues

### Common Configuration Problems

#### Issue 1: Port Already in Use
**Problem**: API port is already occupied
**Detection**: Validation fails on port availability
**Solution**:
```bash
# Check what's using the port
lsof -i :8080

# Use a different port in configuration
# Edit the config file to change the port number
```

#### Issue 2: Permission Denied
**Problem**: Cannot write to configuration directory
**Solution**:
```bash
# Check permissions
ls -la ~/.ollamamax

# Fix permissions if needed
chmod 755 ~/.ollamamax
chmod 644 ~/.ollamamax/*.yaml
```

#### Issue 3: Invalid YAML Format
**Problem**: Configuration file has syntax errors
**Detection**: Validation fails on syntax check
**Solution**:
```bash
# Use a YAML validator to check syntax
python3 -c "import yaml; yaml.safe_load(open('~/.ollamamax/quickstart-config.yaml'))"
```

## 📊 Module 2 Assessment

### Knowledge Check ✋

1. **Q**: What command starts the interactive setup wizard?
   **A**: `./bin/ollama-distributed setup`

2. **Q**: Where are configuration files stored by default?
   **A**: `~/.ollamamax/` directory

3. **Q**: What format is used for configuration files?
   **A**: YAML format

4. **Q**: How do you validate a configuration?
   **A**: `./bin/ollama-distributed validate`

5. **Q**: What are the default API and Web ports?
   **A**: API: 8080, Web: 8081

### Practical Check ✋

Verify you can complete these tasks:

- [ ] Run the interactive setup wizard
- [ ] Generate a quickstart configuration  
- [ ] Validate configuration successfully
- [ ] Understand the YAML configuration structure
- [ ] Create a custom configuration file

### Advanced Challenge 🚀

Try creating configurations for different scenarios:

```bash
# Production-style configuration
/home/kp/ollamamax/scripts/config-generator.sh --profile production --security

# GPU-optimized configuration (if you have GPU)
/home/kp/ollamamax/scripts/config-generator.sh --profile gpu

# Cluster configuration
/home/kp/ollamamax/scripts/config-generator.sh --profile cluster --nodes 3
```

## 🎉 Module 2 Complete!

**Congratulations!** You have successfully:

✅ **Configured** your first Ollama Distributed node  
✅ **Learned** the interactive setup process  
✅ **Understood** the configuration file structure  
✅ **Validated** your configuration settings  
✅ **Explored** different configuration profiles  

### Key Takeaways

1. **Interactive Setup**: The setup wizard makes configuration easy for beginners
2. **YAML Format**: Configuration uses human-readable YAML format
3. **Validation**: Always validate configuration before starting services
4. **Profiles**: Different profiles optimize for different use cases
5. **Flexibility**: Custom configurations can be created for specific needs

## 📚 What's Next?

You're now ready for **Module 3: Basic Cluster Operations** where you'll:
- Start your configured node
- Check cluster status and health
- Understand node networking
- Monitor cluster operations
- Learn about distributed system concepts

**Time to continue:** [Module 3: Basic Cluster Operations →](./module-3-cluster.md)

## 💡 Pro Tips

1. **Keep Backups**: Save working configurations for different environments
2. **Document Changes**: Comment your custom configurations
3. **Test Validation**: Always run validation after making changes
4. **Use Profiles**: Leverage existing profiles as starting points
5. **Environment Variables**: Configuration can be overridden with env vars

---

**Module 2 Status**: ✅ Complete  
**Next Module**: [Basic Cluster Operations →](./module-3-cluster.md)  
**Total Progress**: 2/5 modules (40%)