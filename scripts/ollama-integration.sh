#!/bin/bash
# OllamaMax Ollama Integration Script
# Handles integration with existing Ollama installations and model management

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m'

# Configuration
OLLAMA_PATH=""
OLLAMAMAX_CONFIG=""
MODELS_DIR=""
BACKUP_DIR=""
INTEGRATION_MODE="auto"
PRESERVE_MODELS=true
FORCE_INTEGRATION=false

# Ollama detection paths
OLLAMA_SEARCH_PATHS=(
    "/usr/local/bin/ollama"
    "/usr/bin/ollama"
    "$HOME/.local/bin/ollama"
    "$HOME/bin/ollama"
    "$(which ollama 2>/dev/null || true)"
)

usage() {
    echo -e "${CYAN}🔗 OllamaMax Ollama Integration${NC}"
    echo -e "${WHITE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help                 Show this help message"
    echo "  -m, --mode MODE           Integration mode: auto, migrate, coexist"
    echo "  -o, --ollama PATH         Path to existing Ollama binary"
    echo "  -c, --config PATH         OllamaMax configuration file"
    echo "  -d, --models-dir PATH     Models directory to use/migrate"
    echo "  -b, --backup-dir PATH     Backup directory for existing data"
    echo "  --preserve-models         Keep existing models (default)"
    echo "  --no-preserve             Don't preserve existing models"
    echo "  -f, --force               Force integration even with conflicts"
    echo "  --scan                    Scan for existing Ollama installations"
    echo ""
    echo "Integration Modes:"
    echo "  auto     - Automatically detect and choose best integration"
    echo "  migrate  - Migrate existing Ollama data to OllamaMax"
    echo "  coexist  - Run alongside existing Ollama (different ports)"
    echo ""
    echo "Examples:"
    echo "  $0 --scan                           # Scan for Ollama installations"
    echo "  $0 --mode migrate                   # Migrate existing Ollama"
    echo "  $0 --mode coexist --config my.yaml # Run alongside Ollama"
    echo "  $0 --force                          # Force integration"
    echo ""
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -m|--mode)
                INTEGRATION_MODE="$2"
                if [[ ! "$INTEGRATION_MODE" =~ ^(auto|migrate|coexist)$ ]]; then
                    echo -e "${RED}❌ Invalid mode: $INTEGRATION_MODE${NC}"
                    exit 1
                fi
                shift 2
                ;;
            -o|--ollama)
                OLLAMA_PATH="$2"
                shift 2
                ;;
            -c|--config)
                OLLAMAMAX_CONFIG="$2"
                shift 2
                ;;
            -d|--models-dir)
                MODELS_DIR="$2"
                shift 2
                ;;
            -b|--backup-dir)
                BACKUP_DIR="$2"
                shift 2
                ;;
            --preserve-models)
                PRESERVE_MODELS=true
                shift
                ;;
            --no-preserve)
                PRESERVE_MODELS=false
                shift
                ;;
            -f|--force)
                FORCE_INTEGRATION=true
                shift
                ;;
            --scan)
                scan_ollama_installations
                exit 0
                ;;
            *)
                echo -e "${RED}❌ Unknown option: $1${NC}"
                usage
                exit 1
                ;;
        esac
    done
}

# Print header
print_header() {
    echo ""
    echo -e "${CYAN}🔗 OllamaMax Ollama Integration${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${WHITE}Integrating with existing Ollama installations${NC}"
    echo ""
}

# Detect existing Ollama installations
detect_ollama() {
    echo -e "${BLUE}🔍 Detecting Ollama installations...${NC}"
    
    local found_installations=()
    
    # Check common paths
    for path in "${OLLAMA_SEARCH_PATHS[@]}"; do
        if [[ -n "$path" ]] && [[ -x "$path" ]]; then
            local version=$("$path" --version 2>/dev/null | head -n1 || echo "unknown")
            found_installations+=("$path:$version")
            echo -e "   ${GREEN}✅ Found: $path${NC} ($version)"
        fi
    done
    
    # Check for running Ollama services
    if pgrep -f "ollama" > /dev/null; then
        echo -e "   ${YELLOW}⚠️  Ollama service is currently running${NC}"
        
        # Get port information
        local ollama_ports=$(netstat -tlpn 2>/dev/null | grep ollama | awk '{print $4}' | cut -d: -f2 | sort -n)
        if [[ -n "$ollama_ports" ]]; then
            echo -e "   ${CYAN}   Active ports: $(echo $ollama_ports | tr '\n' ' ')${NC}"
        fi
    fi
    
    # Check for Ollama data directories
    local ollama_data_dirs=(
        "$HOME/.ollama"
        "/usr/share/ollama"
        "/opt/ollama"
        "$(ollama env OLLAMA_MODELS 2>/dev/null || true)"
    )
    
    for dir in "${ollama_data_dirs[@]}"; do
        if [[ -n "$dir" ]] && [[ -d "$dir" ]]; then
            local size=$(du -sh "$dir" 2>/dev/null | cut -f1 || echo "unknown")
            echo -e "   ${GREEN}📁 Data directory: $dir${NC} ($size)"
            
            # Count models
            local model_count=0
            if [[ -d "$dir/models" ]]; then
                model_count=$(find "$dir/models" -name "*.bin" -o -name "*.gguf" 2>/dev/null | wc -l)
                if [[ $model_count -gt 0 ]]; then
                    echo -e "   ${GREEN}   📦 Models found: $model_count${NC}"
                fi
            fi
        fi
    done
    
    if [[ ${#found_installations[@]} -eq 0 ]]; then
        echo -e "   ${YELLOW}⚠️  No existing Ollama installations found${NC}"
        return 1
    fi
    
    # Set default Ollama path if not specified
    if [[ -z "$OLLAMA_PATH" ]] && [[ ${#found_installations[@]} -gt 0 ]]; then
        OLLAMA_PATH=$(echo "${found_installations[0]}" | cut -d: -f1)
        echo -e "   ${CYAN}Using: $OLLAMA_PATH${NC}"
    fi
    
    echo ""
    return 0
}

# Scan and display all Ollama installations
scan_ollama_installations() {
    print_header
    
    echo -e "${BLUE}🔍 Comprehensive Ollama Installation Scan${NC}"
    echo -e "${WHITE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    detect_ollama
    
    # Additional scanning
    echo -e "${BLUE}🔍 Checking system services...${NC}"
    
    # Check systemd services
    if command -v systemctl >/dev/null 2>&1; then
        local ollama_services=$(systemctl list-units --type=service | grep -i ollama || true)
        if [[ -n "$ollama_services" ]]; then
            echo -e "   ${GREEN}📋 Systemd services:${NC}"
            echo "$ollama_services" | sed 's/^/      /'
        else
            echo -e "   ${YELLOW}📋 No Ollama systemd services found${NC}"
        fi
    fi
    
    # Check Docker containers
    if command -v docker >/dev/null 2>&1; then
        local ollama_containers=$(docker ps -a --filter="name=ollama" --format="table {{.Names}}\t{{.Status}}" 2>/dev/null || true)
        if [[ -n "$ollama_containers" ]] && [[ "$ollama_containers" != "NAMES	STATUS" ]]; then
            echo -e "   ${GREEN}🐳 Docker containers:${NC}"
            echo "$ollama_containers" | sed 's/^/      /'
        else
            echo -e "   ${YELLOW}🐳 No Ollama Docker containers found${NC}"
        fi
    fi
    
    # Check configuration files
    echo ""
    echo -e "${BLUE}🔍 Checking configuration files...${NC}"
    
    local config_locations=(
        "$HOME/.ollama/config.json"
        "/etc/ollama/config.json"
        "$HOME/.config/ollama/config.json"
    )
    
    for config in "${config_locations[@]}"; do
        if [[ -f "$config" ]]; then
            echo -e "   ${GREEN}⚙️  Config: $config${NC}"
        fi
    done
    
    echo ""
    echo -e "${GREEN}🎯 Integration Recommendations:${NC}"
    
    if [[ ${#found_installations[@]} -eq 0 ]]; then
        echo -e "   • No existing Ollama found - proceed with clean OllamaMax installation"
        echo -e "   • Run: ${CYAN}ollama-distributed quickstart${NC}"
    else
        echo -e "   • Existing Ollama detected - choose integration mode:"
        echo -e "   • Migrate data: ${CYAN}$0 --mode migrate${NC}"
        echo -e "   • Run alongside: ${CYAN}$0 --mode coexist${NC}"
        echo -e "   • Automatic: ${CYAN}$0 --mode auto${NC}"
    fi
    
    echo ""
}

# Analyze existing Ollama configuration
analyze_ollama_config() {
    local ollama_path="$1"
    
    echo -e "${BLUE}📊 Analyzing Ollama configuration...${NC}"
    
    # Get Ollama environment variables
    local ollama_env_vars=""
    if [[ -x "$ollama_path" ]]; then
        ollama_env_vars=$("$ollama_path" env 2>/dev/null || true)
    fi
    
    # Parse important settings
    local models_dir=""
    local host=""
    local port=""
    
    if [[ -n "$ollama_env_vars" ]]; then
        models_dir=$(echo "$ollama_env_vars" | grep "OLLAMA_MODELS=" | cut -d= -f2 | tr -d '"' || true)
        host=$(echo "$ollama_env_vars" | grep "OLLAMA_HOST=" | cut -d= -f2 | tr -d '"' || echo "127.0.0.1")
        port=$(echo "$host" | cut -d: -f2 || echo "11434")
        host=$(echo "$host" | cut -d: -f1)
    fi
    
    # Set defaults if not found
    models_dir=${models_dir:-"$HOME/.ollama/models"}
    host=${host:-"127.0.0.1"}
    port=${port:-"11434"}
    
    echo -e "   Models directory: ${CYAN}$models_dir${NC}"
    echo -e "   Host: ${CYAN}$host${NC}"
    echo -e "   Port: ${CYAN}$port${NC}"
    
    # Check for conflicts with OllamaMax defaults
    local conflicts=()
    
    if [[ "$port" == "8080" ]] || [[ "$port" == "8081" ]]; then
        conflicts+=("Port conflict: Ollama using $port (OllamaMax default)")
    fi
    
    if [[ ${#conflicts[@]} -gt 0 ]]; then
        echo -e "   ${YELLOW}⚠️  Potential conflicts:${NC}"
        for conflict in "${conflicts[@]}"; do
            echo -e "      • $conflict"
        done
    else
        echo -e "   ${GREEN}✅ No conflicts detected${NC}"
    fi
    
    # Store configuration for later use
    export DETECTED_MODELS_DIR="$models_dir"
    export DETECTED_HOST="$host"
    export DETECTED_PORT="$port"
    
    echo ""
}

# Backup existing Ollama data
backup_ollama_data() {
    local backup_timestamp=$(date +"%Y%m%d_%H%M%S")
    BACKUP_DIR=${BACKUP_DIR:-"$HOME/.ollamamax/backups/ollama_$backup_timestamp"}
    
    echo -e "${BLUE}💾 Backing up existing Ollama data...${NC}"
    
    mkdir -p "$BACKUP_DIR"
    
    # Backup directories
    local dirs_to_backup=(
        "$HOME/.ollama"
        "${DETECTED_MODELS_DIR:-$HOME/.ollama/models}"
    )
    
    for dir in "${dirs_to_backup[@]}"; do
        if [[ -d "$dir" ]]; then
            local basename=$(basename "$dir")
            local dest="$BACKUP_DIR/$basename"
            
            echo -e "   ${CYAN}Backing up: $dir → $dest${NC}"
            cp -r "$dir" "$dest"
            
            local size=$(du -sh "$dest" | cut -f1)
            echo -e "   ${GREEN}✅ Backup complete: $size${NC}"
        fi
    done
    
    # Create backup manifest
    cat > "$BACKUP_DIR/manifest.txt" << EOF
Ollama Backup Manifest
Created: $(date)
Backup Directory: $BACKUP_DIR

Original Locations:
$(for dir in "${dirs_to_backup[@]}"; do echo "- $dir"; done)

Restore Instructions:
# Stop OllamaMax
ollama-distributed stop

# Restore data
$(for dir in "${dirs_to_backup[@]}"; do 
    if [[ -d "$dir" ]]; then
        basename=$(basename "$dir")
        echo "cp -r \"$BACKUP_DIR/$basename\" \"$dir\""
    fi
done)

# Restart
ollama-distributed start
EOF
    
    echo -e "   ${GREEN}✅ Backup completed: $BACKUP_DIR${NC}"
    echo -e "   ${CYAN}Manifest: $BACKUP_DIR/manifest.txt${NC}"
    echo ""
}

# Migrate Ollama models to OllamaMax
migrate_models() {
    local source_models_dir="${DETECTED_MODELS_DIR:-$HOME/.ollama/models}"
    local target_models_dir="${MODELS_DIR:-$HOME/.ollamamax/data/models}"
    
    echo -e "${BLUE}🚚 Migrating Ollama models...${NC}"
    
    if [[ ! -d "$source_models_dir" ]]; then
        echo -e "   ${YELLOW}⚠️  Source models directory not found: $source_models_dir${NC}"
        return 0
    fi
    
    # Create target directory
    mkdir -p "$target_models_dir"
    
    # Count and list models
    local model_files=$(find "$source_models_dir" -name "*.bin" -o -name "*.gguf" 2>/dev/null | wc -l)
    
    if [[ $model_files -eq 0 ]]; then
        echo -e "   ${YELLOW}⚠️  No models found in $source_models_dir${NC}"
        return 0
    fi
    
    echo -e "   ${CYAN}Found $model_files model files${NC}"
    echo -e "   ${CYAN}Migrating: $source_models_dir → $target_models_dir${NC}"
    
    # Copy or link models based on space availability
    local available_space=$(df -BG "$target_models_dir" | awk 'NR==2 {print $4}' | sed 's/G//')
    local required_space=$(du -BG "$source_models_dir" | cut -f1 | sed 's/G//')
    
    if [[ $available_space -lt $required_space ]]; then
        echo -e "   ${YELLOW}⚠️  Insufficient space for copy ($required_space GB needed, $available_space GB available)${NC}"
        echo -e "   ${CYAN}Creating symbolic links instead...${NC}"
        
        # Create symbolic links
        find "$source_models_dir" -type f \( -name "*.bin" -o -name "*.gguf" \) | while read -r model_file; do
            local rel_path="${model_file#$source_models_dir/}"
            local target_path="$target_models_dir/$rel_path"
            local target_dir=$(dirname "$target_path")
            
            mkdir -p "$target_dir"
            ln -sf "$model_file" "$target_path"
            echo -e "   ${GREEN}🔗 Linked: $(basename "$model_file")${NC}"
        done
    else
        echo -e "   ${CYAN}Copying models (sufficient space available)...${NC}"
        
        # Copy models
        cp -r "$source_models_dir"/* "$target_models_dir/"
        echo -e "   ${GREEN}✅ Models copied successfully${NC}"
    fi
    
    # Update model registry if needed
    local model_list="$target_models_dir/models.json"
    if [[ ! -f "$model_list" ]]; then
        echo -e "   ${CYAN}Generating model registry...${NC}"
        
        cat > "$model_list" << EOF
{
  "version": "1.0",
  "migrated_from": "ollama",
  "migration_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "models": [
$(find "$target_models_dir" -name "*.bin" -o -name "*.gguf" | while read -r model; do
    local size=$(stat -f%z "$model" 2>/dev/null || stat -c%s "$model" 2>/dev/null || echo "0")
    local name=$(basename "$model" | sed 's/\.(bin|gguf)$//')
    echo "    {\"name\": \"$name\", \"file\": \"$(basename "$model")\", \"size\": $size},"
done | sed '$s/,$//')
  ]
}
EOF
    fi
    
    echo -e "   ${GREEN}✅ Model migration complete${NC}"
    echo ""
}

# Configure OllamaMax for coexistence
configure_coexistence() {
    echo -e "${BLUE}⚙️  Configuring coexistence mode...${NC}"
    
    # Load or create OllamaMax configuration
    local config_file="${OLLAMAMAX_CONFIG:-config.yaml}"
    
    if [[ ! -f "$config_file" ]]; then
        echo -e "   ${CYAN}Creating OllamaMax configuration...${NC}"
        
        # Generate configuration with different ports
        local api_port=8080
        local web_port=8081
        local p2p_port=8180
        
        # Adjust ports if Ollama is using standard ports
        if [[ "$DETECTED_PORT" == "8080" ]]; then
            api_port=8090
            web_port=8091
            p2p_port=8190
            echo -e "   ${YELLOW}Adjusting ports to avoid conflict with Ollama${NC}"
        fi
        
        cat > "$config_file" << EOF
# OllamaMax Configuration - Coexistence Mode
# Generated for running alongside existing Ollama installation

node:
  id: "ollamamax-coexist-$(date +%s)"
  name: "ollamamax-node"
  data_dir: "$HOME/.ollamamax/data"
  log_level: "info"
  environment: "coexistence"

api:
  host: "0.0.0.0"
  port: $api_port
  enable_tls: false

web:
  enabled: true
  host: "0.0.0.0" 
  port: $web_port

p2p:
  enabled: false
  listen_port: $p2p_port

models:
  store_path: "$HOME/.ollamamax/data/models"
  max_cache_size: "4GB"
  auto_cleanup: true
  # Use separate model directory to avoid conflicts
  isolated_storage: true

performance:
  max_concurrency: 4
  memory_limit: "2GB"
  # Reduce resource usage when coexisting
  cooperative_mode: true

# Coexistence settings
coexistence:
  enabled: true
  ollama_port: $DETECTED_PORT
  ollama_models_dir: "$DETECTED_MODELS_DIR"
  share_models: $PRESERVE_MODELS

logging:
  level: "info"
  file: "$HOME/.ollamamax/data/logs/ollamamax.log"
  # Use different log files to avoid conflicts
  separate_logs: true
EOF
        
        echo -e "   ${GREEN}✅ Configuration created: $config_file${NC}"
        echo -e "   ${CYAN}API Port: $api_port (Ollama: $DETECTED_PORT)${NC}"
        echo -e "   ${CYAN}Web Port: $web_port${NC}"
    else
        echo -e "   ${YELLOW}⚠️  Using existing configuration: $config_file${NC}"
    fi
    
    echo ""
}

# Run integration based on selected mode
run_integration() {
    case "$INTEGRATION_MODE" in
        auto)
            echo -e "${BLUE}🤖 Running automatic integration...${NC}"
            
            # Check if Ollama is currently running
            if pgrep -f "ollama" > /dev/null; then
                echo -e "   ${YELLOW}Ollama is running - choosing coexistence mode${NC}"
                INTEGRATION_MODE="coexist"
            else
                echo -e "   ${CYAN}Ollama not running - choosing migration mode${NC}"
                INTEGRATION_MODE="migrate"
            fi
            
            # Recursively call with determined mode
            INTEGRATION_MODE="$INTEGRATION_MODE"
            run_integration
            ;;
            
        migrate)
            echo -e "${BLUE}🚚 Running migration integration...${NC}"
            echo ""
            
            # Check if Ollama is running and offer to stop it
            if pgrep -f "ollama" > /dev/null; then
                echo -e "${YELLOW}⚠️  Ollama is currently running${NC}"
                if [[ "$FORCE_INTEGRATION" == false ]]; then
                    echo -n "Stop Ollama service for migration? [Y/n]: "
                    read -r stop_choice
                    if [[ ! "$stop_choice" =~ ^[Nn]$ ]]; then
                        echo -e "${CYAN}Stopping Ollama service...${NC}"
                        pkill -f "ollama" || true
                        sleep 2
                    else
                        echo -e "${RED}❌ Cannot migrate while Ollama is running${NC}"
                        exit 1
                    fi
                fi
            fi
            
            # Backup existing data
            backup_ollama_data
            
            # Migrate models if requested
            if [[ "$PRESERVE_MODELS" == true ]]; then
                migrate_models
            fi
            
            # Create migration configuration
            configure_migration
            
            echo -e "${GREEN}✅ Migration integration complete${NC}"
            ;;
            
        coexist)
            echo -e "${BLUE}🤝 Running coexistence integration...${NC}"
            echo ""
            
            # Configure for coexistence
            configure_coexistence
            
            # Optionally link models
            if [[ "$PRESERVE_MODELS" == true ]]; then
                migrate_models
            fi
            
            echo -e "${GREEN}✅ Coexistence integration complete${NC}"
            ;;
    esac
    
    echo ""
}

# Configure OllamaMax for migration mode  
configure_migration() {
    echo -e "${BLUE}⚙️  Configuring migration mode...${NC}"
    
    local config_file="${OLLAMAMAX_CONFIG:-config.yaml}"
    
    cat > "$config_file" << EOF
# OllamaMax Configuration - Migration Mode
# Migrated from existing Ollama installation

node:
  id: "ollamamax-migrated-$(date +%s)"
  name: "ollamamax-migrated"
  data_dir: "$HOME/.ollamamax/data"
  log_level: "info"
  environment: "migrated"

api:
  host: "0.0.0.0"
  port: ${DETECTED_PORT:-8080}
  enable_tls: false

web:
  enabled: true
  host: "0.0.0.0"
  port: $((${DETECTED_PORT:-8080} + 1))

models:
  store_path: "$HOME/.ollamamax/data/models"
  max_cache_size: "8GB"
  auto_cleanup: true
  # Migrated from Ollama
  migration_source: "$DETECTED_MODELS_DIR"

performance:
  max_concurrency: 4
  memory_limit: "4GB"

# Migration metadata
migration:
  enabled: true
  source: "ollama"
  migration_date: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  backup_location: "$BACKUP_DIR"
  original_config:
    host: "$DETECTED_HOST"
    port: "$DETECTED_PORT"
    models_dir: "$DETECTED_MODELS_DIR"

logging:
  level: "info"
  file: "$HOME/.ollamamax/data/logs/ollamamax.log"
EOF
    
    echo -e "   ${GREEN}✅ Migration configuration created: $config_file${NC}"
    echo ""
}

# Display integration summary
show_integration_summary() {
    echo -e "${GREEN}🎉 Integration Summary${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    echo -e "${WHITE}📊 Integration Details:${NC}"
    echo -e "   Mode: ${CYAN}$INTEGRATION_MODE${NC}"
    echo -e "   Ollama Path: ${CYAN}${OLLAMA_PATH:-Not detected}${NC}"
    echo -e "   Models Preserved: ${CYAN}$PRESERVE_MODELS${NC}"
    
    if [[ -n "$BACKUP_DIR" ]]; then
        echo -e "   Backup Location: ${CYAN}$BACKUP_DIR${NC}"
    fi
    
    echo ""
    
    case "$INTEGRATION_MODE" in
        migrate)
            echo -e "${WHITE}🚀 Next Steps (Migration):${NC}"
            echo -e "   1. Validate: ${CYAN}ollama-distributed validate${NC}"
            echo -e "   2. Start: ${CYAN}ollama-distributed start${NC}"
            echo -e "   3. Test: ${CYAN}curl http://localhost:${DETECTED_PORT:-8080}/health${NC}"
            echo ""
            echo -e "${YELLOW}💡 Note: OllamaMax is now using Ollama's original port${NC}"
            ;;
            
        coexist)
            echo -e "${WHITE}🚀 Next Steps (Coexistence):${NC}"
            echo -e "   1. Validate: ${CYAN}ollama-distributed validate${NC}"
            echo -e "   2. Start: ${CYAN}ollama-distributed start${NC}"
            echo -e "   3. Test OllamaMax: ${CYAN}curl http://localhost:8080/health${NC}"
            echo -e "   4. Test Ollama: ${CYAN}curl http://localhost:${DETECTED_PORT:-11434}/health${NC}"
            echo ""
            echo -e "${YELLOW}💡 Both services will run on different ports${NC}"
            ;;
    esac
    
    echo -e "${WHITE}📚 Additional Commands:${NC}"
    echo -e "   Status: ${CYAN}ollama-distributed status${NC}"
    echo -e "   Examples: ${CYAN}ollama-distributed examples${NC}"
    echo -e "   Troubleshoot: ${CYAN}ollama-distributed troubleshoot${NC}"
    echo ""
    
    if [[ -n "$BACKUP_DIR" ]]; then
        echo -e "${WHITE}🔄 Rollback Instructions:${NC}"
        echo -e "   See: ${CYAN}$BACKUP_DIR/manifest.txt${NC}"
        echo ""
    fi
}

# Main function
main() {
    parse_args "$@"
    
    print_header
    
    # Detect existing Ollama installations
    if detect_ollama; then
        analyze_ollama_config "$OLLAMA_PATH"
        run_integration
        show_integration_summary
    else
        echo -e "${GREEN}✅ No existing Ollama installations found${NC}"
        echo -e "${CYAN}You can proceed with a clean OllamaMax installation:${NC}"
        echo -e "   ${CYAN}ollama-distributed quickstart${NC}"
        echo ""
    fi
}

# Run main function
main "$@"