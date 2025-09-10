#!/bin/bash

# Auto-Persist Initialization Script
# Add this to your shell profile (.bashrc, .zshrc, etc.)

OLLAMAMAX_DIR="$HOME/ollamamax"
AUTO_PERSIST_SCRIPT="$OLLAMAMAX_DIR/scripts/auto-persist.js"
AUTO_PERSIST_PID="$OLLAMAMAX_DIR/.claude-flow/auto-persist.pid"

# Function to check if auto-persist is running
check_auto_persist() {
    if [ -f "$AUTO_PERSIST_PID" ]; then
        PID=$(cat "$AUTO_PERSIST_PID" 2>/dev/null)
        if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
            return 0  # Running
        fi
    fi
    return 1  # Not running
}

# Function to start auto-persist if in ollamamax directory
start_auto_persist_if_needed() {
    # Check if we're in the ollamamax directory or a subdirectory
    if [[ "$PWD" == *"/ollamamax"* ]]; then
        if ! check_auto_persist; then
            echo "🚀 Starting automatic memory persistence..."
            (cd "$OLLAMAMAX_DIR" && nohup node "$AUTO_PERSIST_SCRIPT" start > /dev/null 2>&1 &)
            sleep 1
            
            if check_auto_persist; then
                echo "✅ Auto-persist daemon started"
            else
                echo "⚠️ Failed to start auto-persist daemon"
            fi
        fi
    fi
}

# Auto-start on shell initialization if in project
start_auto_persist_if_needed

# Hook into cd command to auto-start when entering project
cd() {
    builtin cd "$@"
    start_auto_persist_if_needed
}

# Cleanup function for shell exit
cleanup_auto_persist() {
    if check_auto_persist && [[ "$PWD" == *"/ollamamax"* ]]; then
        echo "💾 Saving final memory state..."
        (cd "$OLLAMAMAX_DIR" && node "$AUTO_PERSIST_SCRIPT" stop > /dev/null 2>&1)
    fi
}

# Register cleanup on exit (for bash)
if [ -n "$BASH_VERSION" ]; then
    trap cleanup_auto_persist EXIT
fi

# For zsh
if [ -n "$ZSH_VERSION" ]; then
    zshexit() {
        cleanup_auto_persist
    }
fi

# Aliases for manual control
alias persist-start="cd '$OLLAMAMAX_DIR' && node scripts/auto-persist.js start"
alias persist-stop="cd '$OLLAMAMAX_DIR' && node scripts/auto-persist.js stop"
alias persist-status="cd '$OLLAMAMAX_DIR' && node scripts/auto-persist.js status"
alias persist-logs="cd '$OLLAMAMAX_DIR' && node scripts/auto-persist.js logs"

# Claude Flow aliases with automatic memory operations
alias cf="npx claude-flow@alpha"
alias cf-save="npx claude-flow@alpha memory export 'memory/backups/manual-$(date +%Y%m%d-%H%M%S).json'"
alias cf-load="cd '$OLLAMAMAX_DIR' && node scripts/memory-persist.js import"

# Function to show memory persistence status in prompt (optional)
memory_persist_status() {
    if check_auto_persist; then
        echo "💾"
    fi
}

# Example: Add to PS1 for bash
# PS1='$(memory_persist_status) \u@\h:\w\$ '

echo "🧠 Memory persistence environment loaded"
echo "   Use 'persist-status' to check daemon status"
echo "   Auto-persist will start when you enter the ollamamax directory"