#!/usr/bin/env node

/**
 * Claude Flow CLI - Main entry point
 */

import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import fs from 'fs';
import chalk from 'chalk';
import { createLogger, format, transports } from 'winston';

// Initialize structured logging
const logger = createLogger({
  level: process.env.LOG_LEVEL || 'info',
  format: format.combine(
    format.timestamp(),
    format.errors({ stack: true }),
    format.json()
  ),
  transports: [
    new transports.Console({
      format: format.combine(
        format.colorize(),
        format.simple()
      )
    })
  ]
});
import { AnalyzeCommand } from './commands/analyze.js';
import { HelpCommand } from './commands/help.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

class ClaudeFlowCLI {
  constructor() {
    this.commands = {
      analyze: new AnalyzeCommand(),
      analyse: new AnalyzeCommand(), // British spelling alias
      help: new HelpCommand()
    };
  }

  /**
   * Main CLI entry point
   */
  async run(args = process.argv.slice(2)) {
    if (args.length === 0) {
      this.showHelp();
      return;
    }

    const command = args[0];
    const commandArgs = args.slice(1);

    // Handle special flags
    if (command === '--help' || command === '-h') {
      this.showHelp();
      return;
    }

    if (command === '--version' || command === '-v') {
      this.showVersion();
      return;
    }

    // Execute command
    if (this.commands[command]) {
      try {
        await this.commands[command].execute(commandArgs);
        logger.info('Command executed successfully', { command, args: commandArgs });
      } catch (error) {
        logger.error('Command execution failed', {
          command,
          args: commandArgs,
          error: error.message,
          stack: error.stack
        });
        console.error(chalk.red(`Error executing ${command}:`), error.message);
        process.exit(1);
      }
    } else {
      logger.warn('Unknown command attempted', { command, args: commandArgs });
      console.error(chalk.red(`Unknown command: ${command}`));
      console.log(chalk.yellow(`Run 'claude-flow --help' for available commands`));
      process.exit(1);
    }
  }

  /**
   * Show help message
   */
  showHelp() {
    console.log(`
${chalk.cyan('Claude Flow - AI-Powered Development Assistant')}

${chalk.yellow('Usage:')} claude-flow <command> [options]

${chalk.yellow('Commands:')}
  ${chalk.green('analyze')} [mode] [target]   Analyze code with AI-powered insights
    Modes: code, performance, security, architecture, hive-mind, all
    
  ${chalk.green('analyse')}                   British spelling alias for analyze
  
  ${chalk.green('help')} [topic]              Show comprehensive help and guidance
    Topics: overview, getting-started, commands, agents, configuration,
           api, tools, examples, troubleshooting, tips

${chalk.yellow('Analysis Examples:')}
  claude-flow analyze                    # Complete analysis of current directory
  claude-flow analyze hive-mind          # Hive-mind coordinated analysis
  claude-flow analyze security ./src     # Security analysis of src folder
  claude-flow analyze all --deep         # Deep analysis of everything
  claude-flow analyse --hive-mind        # British spelling with hive-mind

${chalk.yellow('Options:')}
  --help, -h         Show this help message
  --version, -v      Show version information
  --hive-mind        Enable distributed hive-mind analysis
  --deep             Perform deep, detailed analysis
  --sequential       Run analyses sequentially (default: parallel)

${chalk.gray('For more information: https://github.com/ruvnet/claude-flow')}
    `);
  }

  /**
   * Show version information
   */
  showVersion() {
    try {
      const packagePath = join(__dirname, '..', 'package.json');
      const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf-8'));
      console.log(chalk.cyan(`Claude Flow v${packageJson.version}`));
    } catch (error) {
      console.log(chalk.cyan('Claude Flow v2.0.0'));
    }
  }
}

// Run CLI
const cli = new ClaudeFlowCLI();
// Graceful error handling and cleanup
process.on('uncaughtException', (error) => {
  logger.error('Uncaught exception', { error: error.message, stack: error.stack });
  console.error(chalk.red('Fatal error:'), error.message);
  process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  logger.error('Unhandled promise rejection', { reason, promise });
  console.error(chalk.red('Unhandled rejection:'), reason);
  process.exit(1);
});

cli.run().catch(error => {
  logger.error('CLI execution failed', { error: error.message, stack: error.stack });
  console.error(chalk.red('Fatal error:'), error.message);
  process.exit(1);
});