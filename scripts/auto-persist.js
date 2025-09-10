#!/usr/bin/env node

/**
 * Automatic Memory Persistence Daemon
 * Runs in background and automatically manages memory persistence
 */

import { promises as fs } from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { exec, spawn } from 'child_process';
import { promisify } from 'util';
import { watch } from 'fs';
import { createHash } from 'crypto';

const execAsync = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const PROJECT_ROOT = path.resolve(__dirname, '..');

class AutoPersistDaemon {
  constructor() {
    this.config = {
      projectRoot: PROJECT_ROOT,
      pidFile: path.join(PROJECT_ROOT, '.claude-flow', 'auto-persist.pid'),
      lockFile: path.join(PROJECT_ROOT, '.claude-flow', 'auto-persist.lock'),
      stateFile: path.join(PROJECT_ROOT, '.claude-flow', 'auto-persist.state'),
      logFile: path.join(PROJECT_ROOT, '.claude-flow', 'auto-persist.log'),
      memoryDir: path.join(PROJECT_ROOT, 'memory'),
      backupDir: path.join(PROJECT_ROOT, 'memory', 'backups'),
      autoSaveInterval: 300000, // 5 minutes
      fileWatchDebounce: 10000, // 10 seconds
      maxBackups: 20,
      watchPatterns: [
        'memory/memory-store.json',
        '.claude-flow/memory/*.db',
        'src/**/*.{js,ts,jsx,tsx}',
        'tests/**/*.{js,ts}',
        '*.md'
      ]
    };
    
    this.sessionId = null;
    this.lastSaveTime = 0;
    this.lastMemoryHash = null;
    this.saveTimer = null;
    this.fileWatchers = [];
    this.isRunning = false;
    this.pendingSave = false;
  }

  async start() {
    try {
      // Check if already running
      if (await this.isAlreadyRunning()) {
        console.log('⚠️ Auto-persist daemon is already running');
        const pid = await fs.readFile(this.config.pidFile, 'utf8');
        console.log(`   PID: ${pid}`);
        return false;
      }

      // Create lock file
      await this.createLock();

      // Initialize
      await this.initialize();

      // Start daemon
      console.log('🚀 Starting auto-persist daemon...');
      this.isRunning = true;

      // Write PID file
      await fs.writeFile(this.config.pidFile, process.pid.toString());

      // Restore last session if exists
      await this.restoreLastSession();

      // Start auto-save timer
      this.startAutoSave();

      // Setup file watchers
      await this.setupFileWatchers();

      // Setup signal handlers
      this.setupSignalHandlers();

      // Log start
      await this.log('Auto-persist daemon started');
      console.log('✅ Auto-persist daemon running');
      console.log(`   PID: ${process.pid}`);
      console.log(`   Session: ${this.sessionId}`);
      console.log(`   Auto-save: Every ${this.config.autoSaveInterval / 60000} minutes`);
      console.log('   Press Ctrl+C to stop');

      // Keep process running
      process.stdin.resume();

    } catch (error) {
      console.error('❌ Failed to start daemon:', error.message);
      await this.cleanup();
      process.exit(1);
    }
  }

  async stop() {
    console.log('🔚 Stopping auto-persist daemon...');
    
    this.isRunning = false;
    
    // Stop timers
    if (this.saveTimer) {
      clearInterval(this.saveTimer);
    }
    
    // Close file watchers
    this.fileWatchers.forEach(watcher => watcher.close());
    
    // Final save
    await this.saveMemory('daemon-stop');
    
    // Cleanup
    await this.cleanup();
    
    console.log('✅ Auto-persist daemon stopped');
  }

  async initialize() {
    // Create necessary directories
    await fs.mkdir(path.join(PROJECT_ROOT, '.claude-flow'), { recursive: true });
    await fs.mkdir(this.config.backupDir, { recursive: true });
    await fs.mkdir(path.join(PROJECT_ROOT, 'memory', 'sessions'), { recursive: true });
    
    // Generate session ID
    this.sessionId = `auto-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    // Save initial state
    await this.saveState();
  }

  async isAlreadyRunning() {
    try {
      const pid = await fs.readFile(this.config.pidFile, 'utf8');
      // Check if process is actually running
      process.kill(parseInt(pid), 0);
      return true;
    } catch {
      // Not running or PID file doesn't exist
      return false;
    }
  }

  async createLock() {
    await fs.writeFile(this.config.lockFile, JSON.stringify({
      pid: process.pid,
      startTime: new Date().toISOString(),
      sessionId: this.sessionId
    }));
  }

  async restoreLastSession() {
    try {
      // Check for last session state
      const stateData = await fs.readFile(this.config.stateFile, 'utf8');
      const state = JSON.parse(stateData);
      
      console.log('📥 Restoring last session...');
      
      // Import last memory backup
      if (state.lastBackup) {
        await execAsync(
          `npx claude-flow@alpha memory import "${state.lastBackup}"`,
          { cwd: PROJECT_ROOT }
        );
        console.log(`   Restored from: ${path.basename(state.lastBackup)}`);
      }
      
    } catch (error) {
      // No previous session or error restoring
      console.log('ℹ️ Starting fresh session');
    }
  }

  async saveState() {
    const state = {
      sessionId: this.sessionId,
      lastSaveTime: this.lastSaveTime,
      lastBackup: await this.getLatestBackup(),
      startTime: new Date().toISOString()
    };
    
    await fs.writeFile(this.config.stateFile, JSON.stringify(state, null, 2));
  }

  async saveMemory(reason = 'auto') {
    // Prevent duplicate saves
    if (this.pendingSave) return;
    
    this.pendingSave = true;
    
    try {
      // Check if memory has changed
      const memoryHash = await this.getMemoryHash();
      if (memoryHash === this.lastMemoryHash && reason === 'auto') {
        this.pendingSave = false;
        return;
      }
      
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const filename = `auto-${reason}-${timestamp}.json`;
      const filepath = path.join(this.config.backupDir, filename);
      
      // Export memory
      const { stdout, stderr } = await execAsync(
        `npx claude-flow@alpha memory export "${filepath}"`,
        { cwd: PROJECT_ROOT }
      );
      
      if (!stderr || stderr.includes('✅')) {
        this.lastSaveTime = Date.now();
        this.lastMemoryHash = memoryHash;
        
        await this.log(`Memory saved: ${filename} (${reason})`);
        console.log(`💾 Auto-saved: ${filename}`);
        
        // Update state
        await this.saveState();
        
        // Cleanup old backups
        await this.cleanupOldBackups();
      }
      
    } catch (error) {
      await this.log(`Save error: ${error.message}`);
      console.error('⚠️ Auto-save failed:', error.message);
    } finally {
      this.pendingSave = false;
    }
  }

  async getMemoryHash() {
    try {
      const memoryFile = path.join(PROJECT_ROOT, 'memory', 'memory-store.json');
      const content = await fs.readFile(memoryFile, 'utf8');
      return createHash('md5').update(content).digest('hex');
    } catch {
      return null;
    }
  }

  async getLatestBackup() {
    try {
      const files = await fs.readdir(this.config.backupDir);
      const backupFiles = files
        .filter(f => f.startsWith('auto-') && f.endsWith('.json'))
        .sort()
        .reverse();
      
      return backupFiles.length > 0 
        ? path.join(this.config.backupDir, backupFiles[0])
        : null;
    } catch {
      return null;
    }
  }

  startAutoSave() {
    this.saveTimer = setInterval(async () => {
      await this.saveMemory('interval');
    }, this.config.autoSaveInterval);
  }

  async setupFileWatchers() {
    console.log('👁️ Setting up file watchers...');
    
    // Watch memory store
    const memoryStore = path.join(PROJECT_ROOT, 'memory', 'memory-store.json');
    if (await this.fileExists(memoryStore)) {
      const watcher = watch(memoryStore, async (eventType) => {
        if (eventType === 'change') {
          await this.debouncedSave('memory-change');
        }
      });
      this.fileWatchers.push(watcher);
    }
    
    // Watch for Claude Flow operations
    const claudeFlowDir = path.join(PROJECT_ROOT, '.claude-flow');
    const claudeWatcher = watch(claudeFlowDir, { recursive: true }, async (eventType, filename) => {
      if (filename && (filename.includes('task') || filename.includes('memory'))) {
        await this.debouncedSave('claude-flow-activity');
      }
    });
    this.fileWatchers.push(claudeWatcher);
    
    console.log('   Watching memory files and Claude Flow activity');
  }

  async debouncedSave(reason) {
    // Simple debounce - wait before saving
    setTimeout(async () => {
      const timeSinceLastSave = Date.now() - this.lastSaveTime;
      if (timeSinceLastSave > this.config.fileWatchDebounce) {
        await this.saveMemory(reason);
      }
    }, this.config.fileWatchDebounce);
  }

  async cleanupOldBackups() {
    try {
      const files = await fs.readdir(this.config.backupDir);
      const autoBackups = files
        .filter(f => f.startsWith('auto-') && f.endsWith('.json'))
        .sort()
        .reverse();
      
      if (autoBackups.length <= this.config.maxBackups) return;
      
      // Remove oldest backups
      const toDelete = autoBackups.slice(this.config.maxBackups);
      for (const file of toDelete) {
        await fs.unlink(path.join(this.config.backupDir, file));
        await this.log(`Deleted old backup: ${file}`);
      }
    } catch (error) {
      await this.log(`Cleanup error: ${error.message}`);
    }
  }

  setupSignalHandlers() {
    // Graceful shutdown
    ['SIGINT', 'SIGTERM'].forEach(signal => {
      process.on(signal, async () => {
        console.log(`\n📛 Received ${signal}`);
        await this.stop();
        process.exit(0);
      });
    });
    
    // Save on exit
    process.on('exit', () => {
      console.log('👋 Goodbye!');
    });
    
    // Handle errors
    process.on('uncaughtException', async (error) => {
      console.error('💥 Uncaught exception:', error);
      await this.log(`Error: ${error.message}`);
      await this.stop();
      process.exit(1);
    });
  }

  async cleanup() {
    try {
      // Remove PID file
      await fs.unlink(this.config.pidFile).catch(() => {});
      
      // Remove lock file
      await fs.unlink(this.config.lockFile).catch(() => {});
      
    } catch (error) {
      console.error('⚠️ Cleanup error:', error.message);
    }
  }

  async log(message) {
    const timestamp = new Date().toISOString();
    const logEntry = `[${timestamp}] ${message}\n`;
    
    await fs.appendFile(this.config.logFile, logEntry).catch(() => {});
  }

  async fileExists(path) {
    try {
      await fs.access(path);
      return true;
    } catch {
      return false;
    }
  }

  async status() {
    try {
      if (!await this.isAlreadyRunning()) {
        console.log('❌ Auto-persist daemon is not running');
        return;
      }
      
      const pid = await fs.readFile(this.config.pidFile, 'utf8');
      const lock = JSON.parse(await fs.readFile(this.config.lockFile, 'utf8'));
      const state = JSON.parse(await fs.readFile(this.config.stateFile, 'utf8'));
      
      console.log('📊 Auto-Persist Daemon Status:');
      console.log(`   PID: ${pid}`);
      console.log(`   Session: ${lock.sessionId}`);
      console.log(`   Started: ${lock.startTime}`);
      console.log(`   Last Save: ${state.lastSaveTime ? new Date(state.lastSaveTime).toLocaleString() : 'Never'}`);
      console.log(`   Last Backup: ${state.lastBackup ? path.basename(state.lastBackup) : 'None'}`);
      
      // Count backups
      const files = await fs.readdir(this.config.backupDir);
      const autoBackups = files.filter(f => f.startsWith('auto-')).length;
      console.log(`   Auto Backups: ${autoBackups}`);
      
    } catch (error) {
      console.error('❌ Error checking status:', error.message);
    }
  }
}

// CLI Interface
async function main() {
  const command = process.argv[2];
  const daemon = new AutoPersistDaemon();
  
  switch (command) {
    case 'start':
      await daemon.start();
      break;
      
    case 'stop':
      // Send signal to running daemon
      try {
        const pid = await fs.readFile(daemon.config.pidFile, 'utf8');
        process.kill(parseInt(pid), 'SIGTERM');
        console.log('✅ Stop signal sent to daemon');
      } catch (error) {
        console.error('❌ No running daemon found');
      }
      break;
      
    case 'restart':
      // Stop then start
      try {
        const pid = await fs.readFile(daemon.config.pidFile, 'utf8');
        process.kill(parseInt(pid), 'SIGTERM');
        console.log('⏸️ Stopping daemon...');
        await new Promise(resolve => setTimeout(resolve, 2000));
      } catch {}
      await daemon.start();
      break;
      
    case 'status':
      await daemon.status();
      break;
      
    case 'logs':
      // Show recent logs
      try {
        const logs = await fs.readFile(daemon.config.logFile, 'utf8');
        const lines = logs.split('\n').slice(-20).join('\n');
        console.log('📜 Recent logs:');
        console.log(lines);
      } catch {
        console.log('No logs available');
      }
      break;
      
    default:
      console.log(`
Automatic Memory Persistence Daemon

Usage: node auto-persist.js <command>

Commands:
  start     Start the daemon
  stop      Stop the daemon
  restart   Restart the daemon
  status    Show daemon status
  logs      Show recent logs

The daemon will:
  • Auto-save memory every 5 minutes
  • Watch for file changes and save on activity
  • Restore memory on startup
  • Clean up old backups automatically
`);
  }
}

// Run if executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch(console.error);
}

export default AutoPersistDaemon;