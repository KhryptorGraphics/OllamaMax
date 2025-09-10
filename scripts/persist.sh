#!/bin/bash

# Simple wrapper for memory persistence operations
# Usage: ./persist.sh [command]

case "$1" in
    save|export)
        # Export current memory with timestamp
        REASON="${2:-manual}"
        echo "💾 Saving memory state..."
        npx claude-flow@alpha memory export "memory/backups/memory-$(date +%Y%m%d-%H%M%S)-$REASON.json"
        ;;
    
    load|import)
        # Import from latest or specified backup
        if [ -n "$2" ]; then
            echo "📥 Loading memory from: $2"
            npx claude-flow@alpha memory import "$2"
        else
            echo "📥 Loading latest memory backup..."
            node scripts/memory-persist.js import
        fi
        ;;
    
    start)
        # Start session with memory restoration
        echo "🚀 Starting persistent session..."
        ./scripts/session-hooks.sh start
        ;;
    
    end)
        # End session with memory export
        echo "🔚 Ending session with memory export..."
        ./scripts/session-hooks.sh end
        ;;
    
    auto)
        # Start auto-backup
        echo "⏱️ Starting auto-backup (Ctrl+C to stop)..."
        node scripts/memory-persist.js auto-backup
        ;;
    
    status)
        # Show memory status
        echo "📊 Memory Status:"
        npx claude-flow@alpha memory list
        echo ""
        echo "📂 Recent backups:"
        ls -lt memory/backups/*.json 2>/dev/null | head -5
        ;;
    
    clean)
        # Cleanup old backups
        echo "🗑️ Cleaning up old backups..."
        node scripts/memory-persist.js cleanup
        ;;
    
    *)
        echo "Memory Persistence Helper"
        echo ""
        echo "Usage: $0 {save|load|start|end|auto|status|clean}"
        echo ""
        echo "Commands:"
        echo "  save [reason]  Export current memory state"
        echo "  load [file]    Import memory from backup"
        echo "  start          Start session with memory restoration"
        echo "  end            End session with memory export"
        echo "  auto           Start automatic backups"
        echo "  status         Show memory and backup status"
        echo "  clean          Remove old backup files"
        echo ""
        echo "Quick usage:"
        echo "  ./persist.sh save         # Quick save"
        echo "  ./persist.sh load         # Load latest"
        echo "  ./persist.sh start        # Begin session"
        echo "  ./persist.sh end          # End session"
        ;;
esac