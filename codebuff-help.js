#!/usr/bin/env node

/**
 * Codebuff Help Command Handler
 * Responds to 'codebuff /help' commands
 */

import { HelpCommand } from './src/commands/help.js';
import chalk from 'chalk';

class CodebuffHelp {
  constructor() {
    this.helpCommand = new HelpCommand();
  }

  /**
   * Handle /help command
   */
  async handleHelp(args = []) {
    // Remove '/help' or 'help' from args if present
    const cleanArgs = args.filter(arg => !arg.startsWith('/help') && arg !== 'help');
    
    console.log(chalk.cyan.bold('\n🤖 Codebuff Help System Activated'));
    console.log(chalk.gray('─'.repeat(50)));
    
    await this.helpCommand.execute(cleanArgs);
    
    console.log(chalk.gray('\n' + '─'.repeat(50)));
    console.log(chalk.cyan('💡 Need more help? Try:'));
    console.log(`  ${chalk.green('codebuff /help [topic]')}     # Specific help topic`);
    console.log(`  ${chalk.green('claude-flow --help')}        # CLI help`);
    console.log(`  ${chalk.green('npm run help')}              # NPM script version`);
    console.log();
  }

  /**
   * Quick help summary
   */
  quickHelp() {
    console.log(`
${chalk.cyan.bold('🚀 Codebuff Quick Help')}
`);
    
    console.log(`${chalk.yellow.bold('Most Common Commands:')}
  ${chalk.green('codebuff /help getting-started')}  # Start here
  ${chalk.green('claude-flow analyze')}             # Analyze code
  ${chalk.green('npm run agents:spawn')}            # Start AI agents
  ${chalk.green('npm run docker:up')}               # Start services
`);
    
    console.log(`${chalk.yellow.bold('Quick Links:')}
  • Web UI: ${chalk.cyan('http://localhost:8081')}
  • API: ${chalk.cyan('http://localhost:8080')}
  • Documentation: ${chalk.cyan('README.md')}
`);
    
    console.log(`${chalk.gray('For complete help:')} ${chalk.white('codebuff /help')}\n`);
  }
}

// Handle command line usage
if (import.meta.url === `file://${process.argv[1]}`) {
  const codebuffHelp = new CodebuffHelp();
  const args = process.argv.slice(2);
  
  // Check if this is a quick help request
  if (args.includes('--quick') || args.includes('-q')) {
    codebuffHelp.quickHelp();
  } else {
    codebuffHelp.handleHelp(args).catch(error => {
      console.error(chalk.red('Error:'), error.message);
      process.exit(1);
    });
  }
}

export default CodebuffHelp;
