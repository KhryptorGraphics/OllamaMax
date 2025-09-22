#!/usr/bin/env node

/**
 * Codebuff Help Command - Comprehensive help system
 */

import chalk from 'chalk';
import fs from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export class HelpCommand {
  constructor() {
    this.topics = {
      'overview': this.showOverview.bind(this),
      'getting-started': this.showGettingStarted.bind(this),
      'commands': this.showCommands.bind(this),
      'agents': this.showAgents.bind(this),
      'configuration': this.showConfiguration.bind(this),
      'api': this.showAPI.bind(this),
      'tools': this.showTools.bind(this),
      'examples': this.showExamples.bind(this),
      'troubleshooting': this.showTroubleshooting.bind(this),
      'tips': this.showTips.bind(this)
    };
  }

  /**
   * Execute help command
   */
  async execute(args = []) {
    const topic = args[0] || 'overview';
    
    if (this.topics[topic]) {
      await this.topics[topic]();
    } else {
      console.log(chalk.red(`Unknown help topic: ${topic}`));
      console.log(chalk.yellow('Available topics:'));
      Object.keys(this.topics).forEach(t => {
        console.log(`  ${chalk.green(t)}`);
      });
    }
  }

  /**
   * Show main overview
   */
  showOverview() {
    console.log(`
${chalk.cyan.bold('🚀 Codebuff + OllamaMax Help System')}
${chalk.cyan('=')}`.repeat(50));
    
    console.log(`
${chalk.yellow.bold('What is this?')}
Codebuff is an AI-powered coding assistant integrated with OllamaMax,
providing distributed AI model capabilities with enterprise-grade features.
`);
    
    console.log(`${chalk.yellow.bold('Quick Start:')}
  ${chalk.green('codebuff /help getting-started')}  # Getting started guide
  ${chalk.green('codebuff /help commands')}         # Available commands
  ${chalk.green('codebuff /help examples')}         # Usage examples
`);
    
    console.log(`${chalk.yellow.bold('Available Help Topics:')}
  ${chalk.green('overview')}         # This overview (default)
  ${chalk.green('getting-started')}  # Getting started guide
  ${chalk.green('commands')}         # All available commands
  ${chalk.green('agents')}           # AI agent system
  ${chalk.green('configuration')}    # Configuration options
  ${chalk.green('api')}              # API reference
  ${chalk.green('tools')}            # Available tools
  ${chalk.green('examples')}         # Usage examples
  ${chalk.green('troubleshooting')}  # Common issues
  ${chalk.green('tips')}             # Pro tips & tricks
`);
    
    console.log(`${chalk.gray('Usage:')} ${chalk.white('codebuff /help [topic]')}
`);
  }

  /**
   * Show getting started guide
   */
  showGettingStarted() {
    console.log(`
${chalk.cyan.bold('🚀 Getting Started with Codebuff + OllamaMax')}
`);
    
    console.log(`${chalk.yellow.bold('Step 1: Quick Setup')}
  ${chalk.green('# Start OllamaMax distributed system')}
  npm run build:distributed
  npm run docker:up

  ${chalk.green('# Verify installation')}
  curl http://localhost:8080/health
`);
    
    console.log(`${chalk.yellow.bold('Step 2: Run Analysis')}
  ${chalk.green('# Analyze your codebase')}
  claude-flow analyze
  
  ${chalk.green('# Or specific analysis types')}
  claude-flow analyze security
  claude-flow analyze performance
  claude-flow analyze hive-mind
`);
    
    console.log(`${chalk.yellow.bold('Step 3: Use AI Agents')}
  ${chalk.green('# Spawn AI agents for tasks')}
  npm run agents:spawn
  
  ${chalk.green('# Check agent status')}
  npm run agents:status
`);
    
    console.log(`${chalk.yellow.bold('Step 4: Access Web Interface')}
  ${chalk.green('# Open web dashboard')}
  open http://localhost:8081
  
  ${chalk.green('# API endpoint')}
  curl http://localhost:8080/api/v1/status
`);
    
    console.log(`${chalk.yellow.bold('Next Steps:')}
  • Read the configuration guide: ${chalk.cyan('codebuff /help configuration')}
  • Explore available commands: ${chalk.cyan('codebuff /help commands')}
  • See examples: ${chalk.cyan('codebuff /help examples')}
`);
  }

  /**
   * Show available commands
   */
  showCommands() {
    console.log(`
${chalk.cyan.bold('📋 Available Commands')}
`);
    
    console.log(`${chalk.yellow.bold('Analysis Commands:')}
  ${chalk.green('claude-flow analyze [mode] [target]')}  # AI-powered code analysis
    Modes: code, performance, security, architecture, hive-mind, all
    
  ${chalk.green('claude-flow analyse')}                  # British spelling alias
`);
    
    console.log(`${chalk.yellow.bold('Agent Commands:')}
  ${chalk.green('npm run agents')}                      # Launch smart agents
  ${chalk.green('npm run agents:status')}               # Check agent status
  ${chalk.green('npm run agents:spawn')}                # Spawn new agents
  ${chalk.green('npm run agents:orchestrate')}          # Orchestrate agents
  ${chalk.green('npm run agents:list')}                 # List active agents
`);
    
    console.log(`${chalk.yellow.bold('Development Commands:')}
  ${chalk.green('npm run build')}                       # Build Go binaries
  ${chalk.green('npm run build:distributed')}           # Build distributed system
  ${chalk.green('npm run dev')}                         # Development mode
  ${chalk.green('npm run validate')}                    # Run tests & linting
`);
    
    console.log(`${chalk.yellow.bold('Testing Commands:')}
  ${chalk.green('npm run test')}                        # Run all tests
  ${chalk.green('npm run test:comprehensive')}          # Comprehensive testing
  ${chalk.green('npm run test:performance')}            # Performance tests
  ${chalk.green('npm run test:api')}                    # API tests
  ${chalk.green('npm run test:ui')}                     # UI tests
`);
    
    console.log(`${chalk.yellow.bold('Docker Commands:')}
  ${chalk.green('npm run docker:build')}                # Build Docker images
  ${chalk.green('npm run docker:up')}                   # Start containers
  ${chalk.green('npm run docker:down')}                 # Stop containers
  ${chalk.green('npm run deploy:backend')}              # Deploy backend
`);
  }

  /**
   * Show agent information
   */
  showAgents() {
    console.log(`
${chalk.cyan.bold('🤖 AI Agent System')}
`);
    
    console.log(`${chalk.yellow.bold('Available Agents:')}
  ${chalk.green('codebuff/file-explorer@0.0.2')}       # Explore codebase
  ${chalk.green('codebuff/file-picker@0.0.2')}         # Find relevant files
  ${chalk.green('codebuff/researcher@0.0.2')}          # Research topics
  ${chalk.green('codebuff/thinker@0.0.2')}             # Deep thinking
  ${chalk.green('codebuff/reviewer@0.0.5')}            # Code review
  ${chalk.green('codebuff/context-pruner@0.0.6')}      # Manage context
  ${chalk.green('codebuff/agent-builder@0.0.2')}       # Build custom agents
`);
    
    console.log(`${chalk.yellow.bold('Agent Usage:')}
  ${chalk.green('# Spawn parallel agents')}
  spawn_agents({
    "agents": [
      {
        "agent_type": "codebuff/file-explorer",
        "prompt": "Find files related to authentication"
      }
    ]
  })
`);
    
    console.log(`${chalk.yellow.bold('Agent Workflows:')}
  • File exploration and analysis
  • Code research and documentation
  • Performance optimization
  • Security analysis
  • Architecture review
`);
  }

  /**
   * Show configuration options
   */
  showConfiguration() {
    console.log(`
${chalk.cyan.bold('⚙️ Configuration')}
`);
    
    console.log(`${chalk.yellow.bold('Configuration Files:')}
  ${chalk.green('codebuff.json')}          # Main Codebuff configuration
  ${chalk.green('config.yaml')}            # OllamaMax configuration
  ${chalk.green('docker-compose.yml')}     # Docker setup
  ${chalk.green('package.json')}           # NPM scripts & dependencies
`);
    
    console.log(`${chalk.yellow.bold('Key Configuration Options:')}
  ${chalk.green('maxAgentSteps')}          # Maximum agent execution steps (default: 25)
  ${chalk.green('baseAgent')}              # Default agent ("buffy")
  ${chalk.green('spawnableAgents')}        # Available agent types
  ${chalk.green('fileChangeHooks')}        # Commands to run on file changes
  ${chalk.green('startupProcesses')}       # Background processes to start
`);
    
    console.log(`${chalk.yellow.bold('Environment Variables:')}
  ${chalk.green('CODEBUFF_API_KEY')}       # Codebuff API key
  ${chalk.green('OLLAMA_HOST')}            # Ollama server host
  ${chalk.green('OLLAMA_PORT')}            # Ollama server port
  ${chalk.green('DEBUG')}                  # Enable debug logging
`);
    
    console.log(`${chalk.yellow.bold('Example codebuff.json:')}
${chalk.gray('{
  "maxAgentSteps": 25,
  "baseAgent": "buffy",
  "fileChangeHooks": [
    {
      "name": "lint",
      "command": "npm run lint",
      "filePattern": "**/*.{js,ts,go}"
    }
  ]
}')}
`);
  }

  /**
   * Show API reference
   */
  showAPI() {
    console.log(`
${chalk.cyan.bold('🌐 API Reference')}
`);
    
    console.log(`${chalk.yellow.bold('Main Endpoints:')}
  ${chalk.green('GET /health')}                        # Health check
  ${chalk.green('GET /api/v1/status')}                 # System status
  ${chalk.green('GET /api/v1/nodes')}                  # Cluster nodes
  ${chalk.green('POST /api/generate')}                 # Generate text
  ${chalk.green('POST /api/chat')}                     # Chat completion
  ${chalk.green('GET /api/v1/models')}                 # Available models
`);
    
    console.log(`${chalk.yellow.bold('Web Interface:')}
  ${chalk.green('http://localhost:8081')}              # Web dashboard
  ${chalk.green('http://localhost:8080')}              # API server
  ${chalk.green('http://localhost:9090')}              # Metrics (Prometheus)
`);
    
    console.log(`${chalk.yellow.bold('Example API Calls:')}
  ${chalk.green('# Health check')}
  curl http://localhost:8080/health
  
  ${chalk.green('# Generate text')}
  curl -X POST http://localhost:8080/api/generate \\
    -H 'Content-Type: application/json' \\
    -d '{"model":"llama2","prompt":"Hello AI!"}'
  
  ${chalk.green('# Chat completion')}
  curl -X POST http://localhost:8080/api/chat \\
    -H 'Content-Type: application/json' \\
    -d '{"model":"llama2","messages":[{"role":"user","content":"Hello!"}]}'
`);
  }

  /**
   * Show available tools
   */
  showTools() {
    console.log(`
${chalk.cyan.bold('🛠️ Available Tools')}
`);
    
    console.log(`${chalk.yellow.bold('Core Tools:')}
  ${chalk.green('read_files')}             # Read file contents
  ${chalk.green('write_file')}             # Write/edit files
  ${chalk.green('str_replace')}            # String replacement in files
  ${chalk.green('code_search')}            # Search code patterns
  ${chalk.green('run_terminal_command')}   # Execute shell commands
`);
    
    console.log(`${chalk.yellow.bold('Agent Tools:')}
  ${chalk.green('spawn_agents')}           # Spawn parallel agents
  ${chalk.green('spawn_agent_inline')}     # Spawn inline agent
  ${chalk.green('add_subgoal')}            # Create subgoals
  ${chalk.green('update_subgoal')}         # Update subgoal progress
  ${chalk.green('end_turn')}               # End conversation turn
`);
    
    console.log(`${chalk.yellow.bold('Analysis Tools:')}
  ${chalk.green('think_deeply')}           # Deep analysis
  ${chalk.green('create_plan')}            # Generate plans
  ${chalk.green('browser_logs')}           # Browser debugging
`);
    
    console.log(`${chalk.yellow.bold('Tool Usage Examples:')}
  ${chalk.green('# Read multiple files')}
  read_files(["src/main.go", "package.json"])
  
  ${chalk.green('# Search for patterns')}
  code_search("function.*authenticate", "-i -t ts")
  
  ${chalk.green('# Spawn research agent')}
  spawn_agents([{
    "agent_type": "codebuff/researcher",
    "prompt": "Research best practices for Go microservices"
  }])
`);
  }

  /**
   * Show usage examples
   */
  showExamples() {
    console.log(`
${chalk.cyan.bold('💡 Usage Examples')}
`);
    
    console.log(`${chalk.yellow.bold('Quick Start Example:')}
  ${chalk.green('# 1. Set up the environment')}
  npm install
  npm run build:distributed
  npm run docker:up
  
  ${chalk.green('# 2. Analyze your code')}
  claude-flow analyze all
  
  ${chalk.green('# 3. Start agents')}
  npm run agents:spawn
`);
    
    console.log(`${chalk.yellow.bold('Development Workflow:')}
  ${chalk.green('# 1. Start development server')}
  npm run dev
  
  ${chalk.green('# 2. Run tests continuously')}
  npm run test:watch
  
  ${chalk.green('# 3. Analyze changes')}
  claude-flow analyze security ./src
  
  ${chalk.green('# 4. Deploy when ready')}
  npm run validate && npm run deploy:backend
`);
    
    console.log(`${chalk.yellow.bold('AI Agent Examples:')}
  ${chalk.green('# Research and documentation')}
  spawn_agents([{
    "agent_type": "codebuff/researcher",
    "prompt": "Research microservices patterns for this codebase"
  }])
  
  ${chalk.green('# Code review')}
  spawn_agents([{
    "agent_type": "codebuff/reviewer",
    "prompt": "Review recent changes for security issues"
  }])
  
  ${chalk.green('# File exploration')}
  spawn_agents([{
    "agent_type": "codebuff/file-explorer",
    "prompt": "Find all authentication-related files"
  }])
`);
    
    console.log(`${chalk.yellow.bold('Production Deployment:')}
  ${chalk.green('# 1. Run comprehensive tests')}
  npm run test:comprehensive
  
  ${chalk.green('# 2. Security analysis')}
  claude-flow analyze security --deep
  
  ${chalk.green('# 3. Performance validation')}
  npm run test:performance
  
  ${chalk.green('# 4. Deploy with monitoring')}
  npm run deploy:backend
  docker-compose -f monitoring/docker-compose.yml up -d
`);
  }

  /**
   * Show troubleshooting guide
   */
  showTroubleshooting() {
    console.log(`
${chalk.cyan.bold('🔧 Troubleshooting Guide')}
`);
    
    console.log(`${chalk.red.bold('Common Issues:')}
`);
    
    console.log(`${chalk.yellow.bold('❌ Command not found')}
  Problem: ${chalk.red('claude-flow: command not found')}
  Solution:
  ${chalk.green('# Install dependencies')}
  npm install
  
  ${chalk.green('# Make CLI executable')}
  chmod +x src/cli.js
  
  ${chalk.green('# Or run directly')}
  node src/cli.js analyze
`);
    
    console.log(`${chalk.yellow.bold('❌ Port already in use')}
  Problem: ${chalk.red('EADDRINUSE: address already in use :::8080')}
  Solution:
  ${chalk.green('# Find process using port')}
  lsof -i :8080
  
  ${chalk.green('# Kill process')}
  kill -9 <PID>
  
  ${chalk.green('# Or use different port')}
  export OLLAMA_PORT=8090
`);
    
    console.log(`${chalk.yellow.bold('❌ Docker issues')}
  Problem: ${chalk.red('Docker containers not starting')}
  Solution:
  ${chalk.green('# Check Docker status')}
  docker-compose ps
  
  ${chalk.green('# View logs')}
  docker-compose logs -f
  
  ${chalk.green('# Rebuild containers')}
  docker-compose down && docker-compose up --build
`);
    
    console.log(`${chalk.yellow.bold('❌ Agent spawn failures')}
  Problem: ${chalk.red('Agents not responding or failing')}
  Solution:
  ${chalk.green('# Check agent status')}
  npm run agents:status
  
  ${chalk.green('# Restart agent system')}
  npm run agents:orchestrate
  
  ${chalk.green('# Check configuration')}
  cat codebuff.json
`);
    
    console.log(`${chalk.yellow.bold('🔍 Diagnostic Commands:')}
  ${chalk.green('# System health')}
  curl http://localhost:8080/health
  
  ${chalk.green('# Check processes')}
  ps aux | grep ollama
  
  ${chalk.green('# View logs')}
  tail -f logs/*.log
  
  ${chalk.green('# Test configuration')}
  npm run validate
`);
  }

  /**
   * Show pro tips
   */
  showTips() {
    console.log(`
${chalk.cyan.bold('💡 Pro Tips & Tricks')}
`);
    
    console.log(`${chalk.yellow.bold('⚡ Performance Tips:')}
  • Use ${chalk.green('hive-mind')} analysis for complex codebases
  • Run ${chalk.green('npm run test:performance')} regularly
  • Monitor resource usage with ${chalk.green('docker stats')}
  • Use ${chalk.green('--sequential')} for memory-constrained environments
`);
    
    console.log(`${chalk.yellow.bold('🤖 Agent Tips:')}
  • Spawn multiple agents in parallel for efficiency
  • Use specific prompts for better agent performance
  • Check agent status before spawning new ones
  • Save agent configurations in ${chalk.green('codebuff.json')}
`);
    
    console.log(`${chalk.yellow.bold('🔧 Development Tips:')}
  • Use ${chalk.green('npm run test:watch')} during development
  • Set up file change hooks for automatic testing
  • Use ${chalk.green('DEBUG=1')} for verbose logging
  • Keep ${chalk.green('codebuff.json')} in version control
`);
    
    console.log(`${chalk.yellow.bold('🚀 Deployment Tips:')}
  • Always run ${chalk.green('npm run validate')} before deployment
  • Use ${chalk.green('docker-compose')} for consistent environments
  • Monitor logs with ${chalk.green('docker-compose logs -f')}
  • Set up health checks for production
`);
    
    console.log(`${chalk.yellow.bold('📚 Learning Resources:')}
  • Read ${chalk.green('CLAUDE.md')} for detailed guidance
  • Check ${chalk.green('README.md')} for project overview
  • Explore ${chalk.green('docs/')} directory for documentation
  • Join GitHub discussions for community support
`);
  }
}

// Export for use in other modules
export default HelpCommand;
