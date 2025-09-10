#!/usr/bin/env node

/**
 * Wrapper script to enable 'claude-flow analyse' command
 * Maps 'analyse' to our custom analyze implementation
 */

import { spawn } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Get command arguments
const args = process.argv.slice(2);

// Check if this is an analyze/analyse request
if (args[0] === 'analyse' || args[0] === 'analyze') {
  // Run our custom analyze implementation
  const analyzeScript = path.join(__dirname, '..', 'src', 'cli.js');
  const child = spawn('node', [analyzeScript, 'analyze', ...args.slice(1)], {
    stdio: 'inherit'
  });
  
  child.on('exit', (code) => {
    process.exit(code || 0);
  });
} else {
  // Pass through to regular claude-flow
  const child = spawn('claude-flow', args, {
    stdio: 'inherit'
  });
  
  child.on('exit', (code) => {
    process.exit(code || 0);
  });
}