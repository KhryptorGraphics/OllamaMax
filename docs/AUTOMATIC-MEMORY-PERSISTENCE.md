# 🧠 Automatic Memory Persistence System

## Overview
A fully automated memory persistence system for Claude Code sessions that requires **ZERO manual intervention**. The system automatically saves, restores, and manages memory across all your Claude Code sessions.

## ✨ Features

### 🚀 Fully Automatic Operation
- **Auto-starts** when you enter the ollamamax directory
- **Auto-saves** memory every 5 minutes
- **Auto-saves** on file changes (with intelligent debouncing)
- **Auto-restores** memory when starting new sessions
- **Auto-exports** memory when exiting shell
- **Auto-cleanup** of old backups (keeps last 20)

### 🔄 Seamless Integration
- Integrated with shell profile (bash/zsh)
- Hooks into Claude Flow lifecycle events
- Watches for file system changes
- Monitors Claude Flow operations

### 📊 Smart Management
- Deduplication of memory entries
- MD5 hash checking to avoid redundant saves
- Intelligent debouncing (10-second delay)
- Automatic backup rotation
- Session tracking and summaries

## 🎯 Installation (One-Time Setup)

Run the installation script:
```bash
bash scripts/install-auto-persist.sh
```

Then restart your terminal or run:
```bash
source ~/.bashrc  # or ~/.zshrc
```

**That's it!** The system is now fully automatic.

## 🤖 How It Works

### 1. Automatic Daemon Management
When you `cd` into the ollamamax directory, the daemon automatically:
- Checks if it's already running
- Starts if not running
- Restores memory from last session
- Begins monitoring for changes

### 2. Continuous Monitoring
The daemon watches:
- `memory/memory-store.json` - Direct memory changes
- `.claude-flow/` directory - Task and operation tracking
- Source files - Development activity
- Every 5 minutes - Scheduled backups

### 3. Intelligent Saving
Saves are triggered by:
- **Interval**: Every 5 minutes automatically
- **File Changes**: When memory files are modified
- **Claude Flow Activity**: When tasks/operations complete
- **Shell Exit**: When leaving the directory or closing terminal

### 4. Automatic Restoration
Memory is restored:
- When daemon starts (entering directory)
- When Claude Flow hooks detect empty memory
- When starting new tasks (pre-task hooks)

## 📁 File Structure

```
ollamamax/
├── .claude-flow/
│   ├── auto-persist.pid       # Daemon process ID
│   ├── auto-persist.lock      # Lock file
│   ├── auto-persist.state     # Current state
│   └── auto-persist.log       # Daemon logs
├── memory/
│   ├── backups/
│   │   ├── auto-*.json        # Automatic backups
│   │   └── manual-*.json      # Manual backups
│   ├── sessions/              # Session data
│   └── memory-store.json      # Active memory
└── scripts/
    ├── auto-persist.js        # Main daemon
    ├── auto-persist-init.sh   # Shell integration
    └── install-auto-persist.sh # Installation script
```

## 🛠️ Manual Controls (Optional)

While the system is fully automatic, you can still manually control it:

### Quick Aliases (Added to your shell)
```bash
persist-status   # Check daemon status
persist-logs     # View recent logs
persist-stop     # Stop daemon
persist-start    # Start daemon
cf-save          # Manual memory save
cf-load          # Manual memory restore
```

### Daemon Commands
```bash
node scripts/auto-persist.js status   # Check status
node scripts/auto-persist.js logs     # View logs
node scripts/auto-persist.js stop     # Stop daemon
node scripts/auto-persist.js restart  # Restart daemon
```

## 📊 Monitoring

### Check Status
```bash
persist-status
```

Output shows:
- Daemon PID
- Current session ID
- Last save time
- Number of auto-backups
- Memory statistics

### View Logs
```bash
persist-logs
```

Shows recent daemon activity including:
- Save operations
- Restore operations
- Errors (if any)
- Cleanup operations

## 🔧 Configuration

Edit `.claude-flow/config/memory-persist.json`:

```json
{
  "memoryPersistence": {
    "enabled": true,
    "autoBackup": {
      "enabled": true,
      "intervalMinutes": 5,    // Auto-save interval
      "maxBackups": 20         // Maximum backups to keep
    },
    "sessionManagement": {
      "autoRestore": true,      // Restore on start
      "autoExport": true,       // Export on exit
      "mergeOnConflict": true   // Merge conflicting entries
    }
  }
}
```

## 🚨 Troubleshooting

### Daemon Not Starting
```bash
# Check for existing process
ps aux | grep auto-persist

# Kill stuck process
pkill -f auto-persist

# Restart
node scripts/auto-persist.js restart
```

### Memory Not Saving
```bash
# Check daemon status
persist-status

# Check logs for errors
persist-logs

# Manual save test
npx claude-flow@alpha memory store "test" "value"
```

### Shell Integration Issues
```bash
# Verify installation in shell profile
grep auto-persist ~/.bashrc  # or ~/.zshrc

# Reinstall if needed
bash scripts/install-auto-persist.sh
```

## 🎯 Benefits

1. **Zero Maintenance**: Set it and forget it
2. **Never Lose Work**: Automatic 5-minute backups
3. **Seamless Continuity**: Pick up exactly where you left off
4. **Resource Efficient**: Intelligent debouncing and deduplication
5. **Failure Recovery**: Automatic restoration after crashes
6. **Complete History**: Session tracking and summaries

## 📈 Performance Impact

- **CPU**: < 0.1% (mostly idle)
- **Memory**: ~20MB Node.js process
- **Disk**: ~1MB per backup (20 backups = ~20MB max)
- **Network**: None (all local operations)

## 🔒 Security

- All data stored locally
- No network transmission
- Backups in project directory only
- Process runs as current user
- No elevated permissions required

## 💡 Tips

1. **Let it run**: The daemon is designed to run continuously
2. **Trust the automation**: It will save your work
3. **Check occasionally**: Use `persist-status` to verify it's running
4. **Don't worry about duplicates**: Deduplication is automatic
5. **Old backups are cleaned**: Keeps only last 20 automatically

## 🎉 Summary

The automatic memory persistence system ensures you **never lose context** between Claude Code sessions. It runs silently in the background, saving your work every 5 minutes and whenever changes are detected. When you return to your project, your entire context is automatically restored, allowing you to pick up exactly where you left off.

**No commands to remember. No manual saves needed. It just works.**