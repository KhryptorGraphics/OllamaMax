#!/bin/bash

# Session Hooks for Claude Flow Memory Persistence
# This script manages session lifecycle with memory persistence

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SESSION_ID="session-cf-$(date +%s)-$$"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

# Start session with memory restoration
session_start() {
    print_status "Starting session: $SESSION_ID"
    
    # Create session marker
    echo "$SESSION_ID" > "$PROJECT_ROOT/.claude-flow/current-session"
    
    # Restore memory from latest backup
    if [ -d "$PROJECT_ROOT/memory/backups" ]; then
        LATEST_BACKUP=$(ls -t "$PROJECT_ROOT/memory/backups"/memory-*.json 2>/dev/null | head -1)
        if [ -n "$LATEST_BACKUP" ]; then
            print_status "Restoring memory from: $(basename "$LATEST_BACKUP")"
            npx claude-flow@alpha memory import "$LATEST_BACKUP"
        else
            print_warning "No previous memory backup found"
        fi
    fi
    
    # Register session hooks
    npx claude-flow@alpha hooks pre-task \
        --description "Session $SESSION_ID initialized" \
        --task-id "$SESSION_ID" \
        --auto-spawn-agents
    
    # Start auto-backup in background
    nohup node "$SCRIPT_DIR/memory-persist.js" auto-backup > "$PROJECT_ROOT/.claude-flow/auto-backup.log" 2>&1 &
    echo $! > "$PROJECT_ROOT/.claude-flow/auto-backup.pid"
    
    print_status "Session started with memory persistence enabled"
}

# End session with memory export
session_end() {
    print_status "Ending session: $SESSION_ID"
    
    # Stop auto-backup
    if [ -f "$PROJECT_ROOT/.claude-flow/auto-backup.pid" ]; then
        PID=$(cat "$PROJECT_ROOT/.claude-flow/auto-backup.pid")
        kill $PID 2>/dev/null
        rm "$PROJECT_ROOT/.claude-flow/auto-backup.pid"
        print_status "Auto-backup stopped"
    fi
    
    # Export final memory state
    EXPORT_FILE="$PROJECT_ROOT/memory/backups/memory-session-end-$(date +%Y%m%d-%H%M%S).json"
    npx claude-flow@alpha memory export "$EXPORT_FILE"
    print_status "Memory exported to: $(basename "$EXPORT_FILE")"
    
    # Generate session summary
    npx claude-flow@alpha hooks session-end \
        --export-metrics \
        --generate-summary \
        --swarm-id "$SESSION_ID"
    
    # Clean up session marker
    rm -f "$PROJECT_ROOT/.claude-flow/current-session"
    
    print_status "Session ended successfully"
}

# Task pre-hook
task_pre() {
    TASK_DESC="$1"
    print_status "Pre-task hook: $TASK_DESC"
    
    # Save task context
    echo "{\"task\": \"$TASK_DESC\", \"timestamp\": \"$(date -Iseconds)\"}" > \
        "$PROJECT_ROOT/.claude-flow/current-task.json"
    
    npx claude-flow@alpha hooks pre-task \
        --description "$TASK_DESC" \
        --task-id "task-$(date +%s)"
}

# Task post-hook
task_post() {
    TASK_ID="$1"
    print_status "Post-task hook: $TASK_ID"
    
    npx claude-flow@alpha hooks post-task \
        --task-id "$TASK_ID" \
        --analyze-performance \
        --generate-insights
    
    # Quick memory backup after task
    npx claude-flow@alpha memory export \
        "$PROJECT_ROOT/memory/backups/memory-task-$(date +%Y%m%d-%H%M%S).json"
}

# Edit pre-hook
edit_pre() {
    FILE="$1"
    print_status "Pre-edit hook: $FILE"
    
    # Backup file before edit
    if [ -f "$FILE" ]; then
        cp "$FILE" "$FILE.backup.$(date +%s)"
    fi
    
    npx claude-flow@alpha hooks pre-edit \
        --file "$FILE" \
        --operation edit
}

# Edit post-hook
edit_post() {
    FILE="$1"
    print_status "Post-edit hook: $FILE"
    
    npx claude-flow@alpha hooks post-edit \
        --file "$FILE" \
        --memory-key "edits/$(date +%s)/$FILE"
}

# Memory consolidation
consolidate_memory() {
    print_status "Consolidating memory..."
    
    # Get all memory namespaces
    NAMESPACES=$(npx claude-flow@alpha memory list | grep -o '^\s*\w\+' | xargs)
    
    for NS in $NAMESPACES; do
        print_status "Processing namespace: $NS"
        # Export and reimport to consolidate
        TEMP_FILE="/tmp/memory-$NS-$(date +%s).json"
        npx claude-flow@alpha memory export "$TEMP_FILE" --namespace "$NS"
        npx claude-flow@alpha memory clear --namespace "$NS"
        npx claude-flow@alpha memory import "$TEMP_FILE"
        rm "$TEMP_FILE"
    done
    
    print_status "Memory consolidation complete"
}

# Main command handler
case "$1" in
    start)
        session_start
        ;;
    end)
        session_end
        ;;
    task-pre)
        task_pre "$2"
        ;;
    task-post)
        task_post "$2"
        ;;
    edit-pre)
        edit_pre "$2"
        ;;
    edit-post)
        edit_post "$2"
        ;;
    consolidate)
        consolidate_memory
        ;;
    status)
        if [ -f "$PROJECT_ROOT/.claude-flow/current-session" ]; then
            CURRENT_SESSION=$(cat "$PROJECT_ROOT/.claude-flow/current-session")
            print_status "Active session: $CURRENT_SESSION"
            
            if [ -f "$PROJECT_ROOT/.claude-flow/auto-backup.pid" ]; then
                print_status "Auto-backup is running"
            else
                print_warning "Auto-backup is not running"
            fi
            
            # Show memory stats
            npx claude-flow@alpha memory list
        else
            print_warning "No active session"
        fi
        ;;
    *)
        echo "Usage: $0 {start|end|task-pre|task-post|edit-pre|edit-post|consolidate|status}"
        echo ""
        echo "Commands:"
        echo "  start         Start a new session with memory restoration"
        echo "  end           End current session with memory export"
        echo "  task-pre      Pre-task hook (requires task description)"
        echo "  task-post     Post-task hook (requires task ID)"
        echo "  edit-pre      Pre-edit hook (requires file path)"
        echo "  edit-post     Post-edit hook (requires file path)"
        echo "  consolidate   Consolidate and deduplicate memory"
        echo "  status        Show current session status"
        exit 1
        ;;
esac