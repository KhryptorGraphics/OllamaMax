#!/bin/bash

# Installation script for automatic memory persistence

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🔧 Installing Automatic Memory Persistence System..."
echo ""

# Make scripts executable
echo "📝 Making scripts executable..."
chmod +x "$SCRIPT_DIR/auto-persist.js"
chmod +x "$SCRIPT_DIR/auto-persist-init.sh"
chmod +x "$SCRIPT_DIR/memory-persist.js"
chmod +x "$SCRIPT_DIR/session-hooks.sh"
chmod +x "$SCRIPT_DIR/persist.sh"

# Detect shell
SHELL_NAME=$(basename "$SHELL")
SHELL_RC=""

case "$SHELL_NAME" in
    bash)
        SHELL_RC="$HOME/.bashrc"
        ;;
    zsh)
        SHELL_RC="$HOME/.zshrc"
        ;;
    *)
        echo "⚠️ Unsupported shell: $SHELL_NAME"
        echo "   Please manually add the following to your shell profile:"
        echo "   source $SCRIPT_DIR/auto-persist-init.sh"
        exit 1
        ;;
esac

# Check if already installed
if grep -q "auto-persist-init.sh" "$SHELL_RC" 2>/dev/null; then
    echo "✅ Auto-persist already installed in $SHELL_RC"
else
    echo "📝 Adding auto-persist to $SHELL_RC..."
    echo "" >> "$SHELL_RC"
    echo "# Ollamamax Automatic Memory Persistence" >> "$SHELL_RC"
    echo "if [ -f \"$SCRIPT_DIR/auto-persist-init.sh\" ]; then" >> "$SHELL_RC"
    echo "    source \"$SCRIPT_DIR/auto-persist-init.sh\"" >> "$SHELL_RC"
    echo "fi" >> "$SHELL_RC"
    echo "✅ Added to $SHELL_RC"
fi

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p "$PROJECT_ROOT/.claude-flow"
mkdir -p "$PROJECT_ROOT/memory/backups"
mkdir -p "$PROJECT_ROOT/memory/sessions"
mkdir -p "$PROJECT_ROOT/memory/exports"

# Create initial configuration if not exists
CONFIG_FILE="$PROJECT_ROOT/.claude-flow/config/memory-persist.json"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "📝 Creating configuration..."
    mkdir -p "$(dirname "$CONFIG_FILE")"
    cat > "$CONFIG_FILE" << 'EOF'
{
  "memoryPersistence": {
    "enabled": true,
    "autoBackup": {
      "enabled": true,
      "intervalMinutes": 5,
      "maxBackups": 20
    },
    "sessionManagement": {
      "autoRestore": true,
      "autoExport": true,
      "mergeOnConflict": true
    },
    "directories": {
      "backups": "./memory/backups",
      "sessions": "./memory/sessions",
      "exports": "./memory/exports"
    }
  }
}
EOF
fi

# Start the daemon
echo "🚀 Starting auto-persist daemon..."
cd "$PROJECT_ROOT"
node scripts/auto-persist.js stop 2>/dev/null || true
sleep 1
nohup node scripts/auto-persist.js start > /dev/null 2>&1 &
sleep 2

# Check if daemon started
if node scripts/auto-persist.js status | grep -q "not running"; then
    echo "⚠️ Failed to start daemon. Please check logs:"
    echo "   node scripts/auto-persist.js logs"
else
    echo "✅ Auto-persist daemon is running"
fi

# Create systemd service (optional, for Linux systems)
if command -v systemctl &> /dev/null; then
    echo ""
    echo "📋 Optional: Create systemd service for boot startup?"
    echo "   This will ensure memory persistence starts automatically on system boot."
    read -p "   Install systemd service? (y/N): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        SERVICE_FILE="/tmp/ollamamax-persist.service"
        cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Ollamamax Memory Persistence Daemon
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$PROJECT_ROOT
ExecStart=/usr/bin/node $SCRIPT_DIR/auto-persist.js start
ExecStop=/usr/bin/node $SCRIPT_DIR/auto-persist.js stop
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
        
        echo "   Installing systemd service..."
        sudo mv "$SERVICE_FILE" /etc/systemd/system/ollamamax-persist.service
        sudo systemctl daemon-reload
        sudo systemctl enable ollamamax-persist.service
        sudo systemctl start ollamamax-persist.service
        
        echo "✅ Systemd service installed and started"
        echo "   Use 'systemctl status ollamamax-persist' to check status"
    fi
fi

echo ""
echo "✨ Installation Complete!"
echo ""
echo "📋 The system will now:"
echo "   • Auto-start when you enter the ollamamax directory"
echo "   • Save memory every 5 minutes"
echo "   • Save on file changes"
echo "   • Restore memory on session start"
echo "   • Export memory on shell exit"
echo ""
echo "🎯 Quick Commands:"
echo "   persist-status  - Check daemon status"
echo "   persist-logs    - View recent logs"
echo "   persist-stop    - Stop daemon"
echo "   persist-start   - Start daemon"
echo ""
echo "⚡ To apply changes immediately:"
echo "   source $SHELL_RC"
echo ""
echo "Or simply restart your terminal."