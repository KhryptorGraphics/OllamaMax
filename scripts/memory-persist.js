#!/usr/bin/env node

/**
 * Memory Persistence Manager for Claude Flow
 * Handles automated memory export/import across sessions
 */

import { promises as fs } from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const PROJECT_ROOT = path.resolve(__dirname, '..');

class MemoryPersistenceManager {
  constructor() {
    this.config = {
      backupDir: path.join(PROJECT_ROOT, 'memory', 'backups'),
      sessionDir: path.join(PROJECT_ROOT, 'memory', 'sessions'),
      memoryStore: path.join(PROJECT_ROOT, 'memory', 'memory-store.json'),
      sqliteDb: path.join(PROJECT_ROOT, '.claude-flow', 'memory', 'unified-memory.db'),
      maxBackups: 10,
      autoBackupInterval: 300000, // 5 minutes
      compressionEnabled: true
    };
    
    this.sessionId = `session-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    this.backupTimer = null;
  }

  async initialize() {
    // Create necessary directories
    await fs.mkdir(this.config.backupDir, { recursive: true });
    await fs.mkdir(this.config.sessionDir, { recursive: true });
    
    console.log('✅ Memory Persistence Manager initialized');
    console.log(`📋 Session ID: ${this.sessionId}`);
  }

  /**
   * Export memory to a timestamped backup file
   */
  async exportMemory(reason = 'manual') {
    try {
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const filename = `memory-${reason}-${timestamp}.json`;
      const filepath = path.join(this.config.backupDir, filename);
      
      // Export using claude-flow
      const { stdout, stderr } = await execAsync(
        `npx claude-flow@alpha memory export "${filepath}"`,
        { cwd: PROJECT_ROOT }
      );
      
      if (stderr && !stderr.includes('✅')) {
        throw new Error(stderr);
      }
      
      // Create session metadata
      const metadata = {
        sessionId: this.sessionId,
        timestamp: new Date().toISOString(),
        reason,
        filename,
        exportOutput: stdout
      };
      
      await fs.writeFile(
        path.join(this.config.sessionDir, `${this.sessionId}-export.json`),
        JSON.stringify(metadata, null, 2)
      );
      
      console.log(`✅ Memory exported: ${filename}`);
      console.log(`📦 Reason: ${reason}`);
      
      // Cleanup old backups
      await this.cleanupOldBackups();
      
      return filepath;
    } catch (error) {
      console.error('❌ Failed to export memory:', error.message);
      throw error;
    }
  }

  /**
   * Import memory from a backup file
   */
  async importMemory(filepath = null) {
    try {
      // If no filepath provided, use the latest backup
      if (!filepath) {
        filepath = await this.getLatestBackup();
        if (!filepath) {
          console.log('ℹ️ No backup found to import');
          return false;
        }
      }
      
      // Import using claude-flow
      const { stdout, stderr } = await execAsync(
        `npx claude-flow@alpha memory import "${filepath}"`,
        { cwd: PROJECT_ROOT }
      );
      
      if (stderr && !stderr.includes('✅')) {
        throw new Error(stderr);
      }
      
      console.log(`✅ Memory imported from: ${path.basename(filepath)}`);
      console.log(stdout);
      
      return true;
    } catch (error) {
      console.error('❌ Failed to import memory:', error.message);
      throw error;
    }
  }

  /**
   * Start a new session with memory restoration
   */
  async startSession() {
    console.log('🚀 Starting new session with memory persistence...');
    
    // Import latest memory state
    await this.importMemory();
    
    // Setup auto-backup
    this.startAutoBackup();
    
    // Register session hooks
    await this.registerSessionHooks();
    
    console.log('✅ Session started with memory persistence enabled');
  }

  /**
   * End session with memory export
   */
  async endSession() {
    console.log('🔚 Ending session...');
    
    // Stop auto-backup
    this.stopAutoBackup();
    
    // Final export
    await this.exportMemory('session-end');
    
    // Generate session summary
    await this.generateSessionSummary();
    
    console.log('✅ Session ended, memory persisted');
  }

  /**
   * Start automatic periodic backups
   */
  startAutoBackup() {
    this.backupTimer = setInterval(async () => {
      await this.exportMemory('auto-backup');
    }, this.config.autoBackupInterval);
    
    console.log(`⏱️ Auto-backup enabled (every ${this.config.autoBackupInterval / 60000} minutes)`);
  }

  /**
   * Stop automatic backups
   */
  stopAutoBackup() {
    if (this.backupTimer) {
      clearInterval(this.backupTimer);
      this.backupTimer = null;
      console.log('⏸️ Auto-backup disabled');
    }
  }

  /**
   * Get the latest backup file
   */
  async getLatestBackup() {
    try {
      const files = await fs.readdir(this.config.backupDir);
      const backupFiles = files
        .filter(f => f.startsWith('memory-') && f.endsWith('.json'))
        .sort()
        .reverse();
      
      if (backupFiles.length === 0) {
        return null;
      }
      
      return path.join(this.config.backupDir, backupFiles[0]);
    } catch (error) {
      console.error('❌ Failed to find latest backup:', error.message);
      return null;
    }
  }

  /**
   * Clean up old backup files
   */
  async cleanupOldBackups() {
    try {
      const files = await fs.readdir(this.config.backupDir);
      const backupFiles = files
        .filter(f => f.startsWith('memory-') && f.endsWith('.json'))
        .sort()
        .reverse();
      
      if (backupFiles.length <= this.config.maxBackups) {
        return;
      }
      
      // Remove oldest backups
      const filesToDelete = backupFiles.slice(this.config.maxBackups);
      for (const file of filesToDelete) {
        await fs.unlink(path.join(this.config.backupDir, file));
        console.log(`🗑️ Deleted old backup: ${file}`);
      }
    } catch (error) {
      console.error('❌ Failed to cleanup backups:', error.message);
    }
  }

  /**
   * Register session hooks with claude-flow
   */
  async registerSessionHooks() {
    try {
      // Register pre-task hook
      await execAsync(
        `npx claude-flow@alpha hooks pre-task --description "Session ${this.sessionId}" --task-id "session-${this.sessionId}"`,
        { cwd: PROJECT_ROOT }
      );
      
      console.log('🔗 Session hooks registered');
    } catch (error) {
      console.error('⚠️ Failed to register hooks:', error.message);
    }
  }

  /**
   * Generate session summary
   */
  async generateSessionSummary() {
    try {
      const summary = {
        sessionId: this.sessionId,
        startTime: new Date(parseInt(this.sessionId.split('-')[1])).toISOString(),
        endTime: new Date().toISOString(),
        backupsCreated: await this.countSessionBackups(),
        finalMemoryState: await this.getMemoryStats()
      };
      
      await fs.writeFile(
        path.join(this.config.sessionDir, `${this.sessionId}-summary.json`),
        JSON.stringify(summary, null, 2)
      );
      
      console.log('📊 Session summary generated');
    } catch (error) {
      console.error('⚠️ Failed to generate summary:', error.message);
    }
  }

  /**
   * Count backups created in this session
   */
  async countSessionBackups() {
    try {
      const files = await fs.readdir(this.config.backupDir);
      const sessionStart = parseInt(this.sessionId.split('-')[1]);
      
      let count = 0;
      for (const f of files) {
        if (!f.startsWith('memory-') || !f.endsWith('.json')) continue;
        const fileStats = await fs.stat(path.join(this.config.backupDir, f));
        if (fileStats.mtime.getTime() >= sessionStart) {
          count++;
        }
      }
      return count;
    } catch {
      return 0;
    }
  }

  /**
   * Get memory statistics
   */
  async getMemoryStats() {
    try {
      const { stdout } = await execAsync(
        'npx claude-flow@alpha memory list',
        { cwd: PROJECT_ROOT }
      );
      
      // Parse the output to extract stats
      const lines = stdout.split('\n');
      const stats = {};
      
      for (const line of lines) {
        const match = line.match(/(\w+)\s+\((\d+)\s+entries\)/);
        if (match) {
          stats[match[1]] = parseInt(match[2]);
        }
      }
      
      return stats;
    } catch {
      return {};
    }
  }

  /**
   * Merge memory from multiple sessions
   */
  async mergeMemory(sessionIds) {
    console.log('🔄 Merging memory from multiple sessions...');
    
    const mergedData = {};
    
    for (const sessionId of sessionIds) {
      try {
        const summaryPath = path.join(this.config.sessionDir, `${sessionId}-summary.json`);
        const exportPath = path.join(this.config.sessionDir, `${sessionId}-export.json`);
        
        if (await fs.access(exportPath).then(() => true).catch(() => false)) {
          const exportData = JSON.parse(await fs.readFile(exportPath, 'utf8'));
          const backupPath = path.join(this.config.backupDir, exportData.filename);
          
          if (await fs.access(backupPath).then(() => true).catch(() => false)) {
            const memoryData = JSON.parse(await fs.readFile(backupPath, 'utf8'));
            
            // Merge the data
            for (const [namespace, entries] of Object.entries(memoryData)) {
              if (!mergedData[namespace]) {
                mergedData[namespace] = [];
              }
              mergedData[namespace].push(...entries);
            }
          }
        }
      } catch (error) {
        console.error(`⚠️ Failed to merge session ${sessionId}:`, error.message);
      }
    }
    
    // Deduplicate entries
    for (const namespace of Object.keys(mergedData)) {
      const seen = new Map();
      mergedData[namespace] = mergedData[namespace].filter(entry => {
        const key = `${entry.key}-${entry.namespace}`;
        if (seen.has(key)) {
          // Keep the newest entry
          const existing = seen.get(key);
          if (entry.timestamp > existing.timestamp) {
            seen.set(key, entry);
            return true;
          }
          return false;
        }
        seen.set(key, entry);
        return true;
      });
    }
    
    // Save merged data
    const mergedPath = path.join(this.config.backupDir, `memory-merged-${Date.now()}.json`);
    await fs.writeFile(mergedPath, JSON.stringify(mergedData, null, 2));
    
    console.log(`✅ Memory merged to: ${path.basename(mergedPath)}`);
    return mergedPath;
  }
}

// CLI Interface
async function main() {
  const manager = new MemoryPersistenceManager();
  await manager.initialize();
  
  const command = process.argv[2];
  const args = process.argv.slice(3);
  
  switch (command) {
    case 'start':
      await manager.startSession();
      break;
      
    case 'end':
      await manager.endSession();
      break;
      
    case 'export':
      await manager.exportMemory(args[0] || 'manual');
      break;
      
    case 'import':
      await manager.importMemory(args[0]);
      break;
      
    case 'auto-backup':
      manager.startAutoBackup();
      console.log('Auto-backup started. Press Ctrl+C to stop.');
      // Keep process running
      process.on('SIGINT', async () => {
        manager.stopAutoBackup();
        await manager.exportMemory('interrupt');
        process.exit(0);
      });
      break;
      
    case 'merge':
      await manager.mergeMemory(args);
      break;
      
    case 'cleanup':
      await manager.cleanupOldBackups();
      break;
      
    case 'stats':
      const stats = await manager.getMemoryStats();
      console.log('📊 Memory Statistics:');
      console.log(JSON.stringify(stats, null, 2));
      break;
      
    default:
      console.log(`
Memory Persistence Manager

Usage: node memory-persist.js <command> [args]

Commands:
  start         Start a new session with memory restoration
  end           End session with memory export
  export [reason]   Export current memory state
  import [file]     Import memory from backup
  auto-backup   Start automatic periodic backups
  merge [ids]   Merge memory from multiple sessions
  cleanup       Remove old backup files
  stats         Show memory statistics

Examples:
  node memory-persist.js start
  node memory-persist.js export "before-major-change"
  node memory-persist.js import memory-backup.json
  node memory-persist.js auto-backup
`);
  }
}

// Run if executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch(console.error);
}

export { MemoryPersistenceManager };