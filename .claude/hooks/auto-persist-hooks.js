#!/usr/bin/env node

/**
 * Claude Flow Automatic Persistence Hooks
 * Integrates with Claude Flow lifecycle events for automatic memory management
 */

import { exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';
import { fileURLToPath } from 'url';

const execAsync = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const PROJECT_ROOT = path.resolve(__dirname, '../..');

class AutoPersistHooks {
  constructor() {
    this.config = {
      autoSave: true,
      saveOnTaskComplete: true,
      saveOnEdit: true,
      saveOnError: true,
      debounceMs: 5000
    };
    
    this.lastSaveTime = 0;
    this.pendingSave = null;
  }

  /**
   * Pre-task hook - Restore context before task
   */
  async preTask(taskDescription, taskId) {
    console.log('🔄 Auto-persist: Pre-task restoration...');
    
    try {
      // Check if memory needs restoration
      const { stdout } = await execAsync(
        'npx claude-flow@alpha memory list',
        { cwd: PROJECT_ROOT }
      );
      
      if (!stdout.includes('entries')) {
        // Memory is empty, try to restore
        console.log('   Restoring memory from latest backup...');
        await execAsync(
          'node scripts/memory-persist.js import',
          { cwd: PROJECT_ROOT }
        );
      }
      
      // Store task context
      await execAsync(
        `npx claude-flow@alpha memory store "current_task" "${taskDescription}"`,
        { cwd: PROJECT_ROOT }
      );
      
      console.log('✅ Context restored for task');
    } catch (error) {
      console.error('⚠️ Failed to restore context:', error.message);
    }
  }

  /**
   * Post-task hook - Save progress after task
   */
  async postTask(taskId, success = true) {
    if (!this.config.saveOnTaskComplete) return;
    
    console.log('💾 Auto-persist: Saving task progress...');
    
    try {
      const reason = success ? 'task-complete' : 'task-failed';
      await this.saveMemory(reason);
      
      // Update task history
      const timestamp = new Date().toISOString();
      await execAsync(
        `npx claude-flow@alpha memory store "task_history_${taskId}" "${timestamp}:${success}"`,
        { cwd: PROJECT_ROOT }
      );
      
      console.log('✅ Task progress saved');
    } catch (error) {
      console.error('⚠️ Failed to save progress:', error.message);
    }
  }

  /**
   * Pre-edit hook - Backup before file modification
   */
  async preEdit(filepath) {
    // Quick backup before edit
    try {
      await execAsync(
        `cp "${filepath}" "${filepath}.backup.$(date +%s)" 2>/dev/null || true`,
        { cwd: PROJECT_ROOT }
      );
    } catch {}
  }

  /**
   * Post-edit hook - Save after file modification
   */
  async postEdit(filepath) {
    if (!this.config.saveOnEdit) return;
    
    // Debounced save after edit
    this.debouncedSave('file-edit');
  }

  /**
   * Session start hook - Initialize persistence
   */
  async sessionStart() {
    console.log('🚀 Auto-persist: Initializing session...');
    
    try {
      // Start auto-persist daemon if not running
      const { stdout } = await execAsync(
        'node scripts/auto-persist.js status',
        { cwd: PROJECT_ROOT }
      );
      
      if (stdout.includes('not running')) {
        console.log('   Starting auto-persist daemon...');
        await execAsync(
          'nohup node scripts/auto-persist.js start > /dev/null 2>&1 &',
          { cwd: PROJECT_ROOT }
        );
        console.log('✅ Auto-persist daemon started');
      } else {
        console.log('✅ Auto-persist daemon already running');
      }
      
      // Restore last session
      await this.restoreLastSession();
      
    } catch (error) {
      console.error('⚠️ Session initialization error:', error.message);
    }
  }

  /**
   * Session end hook - Export and cleanup
   */
  async sessionEnd() {
    console.log('🔚 Auto-persist: Ending session...');
    
    try {
      // Final memory export
      await this.saveMemory('session-end');
      
      // Generate session summary
      await execAsync(
        'npx claude-flow@alpha hooks session-end --export-metrics --generate-summary',
        { cwd: PROJECT_ROOT }
      );
      
      console.log('✅ Session ended, memory persisted');
    } catch (error) {
      console.error('⚠️ Session end error:', error.message);
    }
  }

  /**
   * Error hook - Save on error
   */
  async onError(error) {
    if (!this.config.saveOnError) return;
    
    console.log('⚠️ Auto-persist: Saving after error...');
    
    try {
      // Save error context
      await execAsync(
        `npx claude-flow@alpha memory store "last_error" "${error.message || error}"`,
        { cwd: PROJECT_ROOT }
      );
      
      // Emergency save
      await this.saveMemory('error-recovery');
      
      console.log('✅ Error state saved');
    } catch (saveError) {
      console.error('❌ Failed to save error state:', saveError.message);
    }
  }

  /**
   * Save memory with reason
   */
  async saveMemory(reason) {
    try {
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const filename = `auto-${reason}-${timestamp}.json`;
      const filepath = path.join(PROJECT_ROOT, 'memory', 'backups', filename);
      
      await execAsync(
        `npx claude-flow@alpha memory export "${filepath}"`,
        { cwd: PROJECT_ROOT }
      );
      
      this.lastSaveTime = Date.now();
      return filepath;
    } catch (error) {
      console.error('Save error:', error.message);
      throw error;
    }
  }

  /**
   * Debounced save to prevent too frequent saves
   */
  debouncedSave(reason) {
    // Clear pending save
    if (this.pendingSave) {
      clearTimeout(this.pendingSave);
    }
    
    // Schedule new save
    this.pendingSave = setTimeout(async () => {
      const timeSinceLastSave = Date.now() - this.lastSaveTime;
      if (timeSinceLastSave > this.config.debounceMs) {
        await this.saveMemory(reason);
      }
      this.pendingSave = null;
    }, this.config.debounceMs);
  }

  /**
   * Restore last session
   */
  async restoreLastSession() {
    try {
      // Get latest backup
      const { stdout } = await execAsync(
        'ls -t memory/backups/*.json | head -1',
        { cwd: PROJECT_ROOT }
      );
      
      if (stdout.trim()) {
        console.log('   Restoring from:', path.basename(stdout.trim()));
        await execAsync(
          `npx claude-flow@alpha memory import "${stdout.trim()}"`,
          { cwd: PROJECT_ROOT }
        );
        console.log('✅ Memory restored');
      }
    } catch (error) {
      console.log('ℹ️ No previous session to restore');
    }
  }
}

// Export for use in Claude Flow hooks
export default AutoPersistHooks;

// CLI interface for testing
if (import.meta.url === `file://${process.argv[1]}`) {
  const hooks = new AutoPersistHooks();
  const command = process.argv[2];
  
  switch (command) {
    case 'pre-task':
      await hooks.preTask(process.argv[3] || 'Test task', 'test-id');
      break;
    case 'post-task':
      await hooks.postTask('test-id', true);
      break;
    case 'session-start':
      await hooks.sessionStart();
      break;
    case 'session-end':
      await hooks.sessionEnd();
      break;
    case 'test':
      console.log('Testing all hooks...');
      await hooks.sessionStart();
      await hooks.preTask('Test task', 'test-1');
      await hooks.postTask('test-1', true);
      await hooks.sessionEnd();
      console.log('✅ All hooks tested');
      break;
    default:
      console.log('Usage: node auto-persist-hooks.js {pre-task|post-task|session-start|session-end|test}');
  }
}