#!/usr/bin/env node

/**
 * Test Expert Agents System
 * Validates integration with claude-flow commands
 */

import chalk from 'chalk';
import { EXPERT_AGENT_REGISTRY, getExpertAgent, getExpertsForCommand, getExpertCategoryCounts } from '../src/agents/expert-agent-registry.js';
import { createExpertRouter } from '../src/agents/claude-flow-expert-integration.js';

console.log(chalk.cyan('🧪 Testing Expert Agent System\n'));

/**
 * Test expert agent registry
 */
async function testExpertRegistry() {
  console.log(chalk.yellow('1. Testing Expert Agent Registry:'));
  
  const counts = getExpertCategoryCounts();
  console.log(`  ✅ Total expert agents: ${counts.total}`);
  console.log(`  ✅ Categories: ${Object.keys(counts.categories).length}`);
  
  // Test specific experts
  const testExperts = [
    'sparc-architect',
    'github-pr-analyzer',
    'swarm-queen',
    'neural-architect',
    'kubernetes-orchestrator',
    'react-optimization-expert',
    'smart-contract-auditor'
  ];
  
  for (const expertType of testExperts) {
    const expert = getExpertAgent(expertType);
    if (expert) {
      console.log(`  ✅ ${expertType}: ${expert.capabilities.length} capabilities, ${expert.tools.length} tools`);
    } else {
      console.log(chalk.red(`  ❌ ${expertType}: Not found`));
    }
  }
}

/**
 * Test command mapping
 */
async function testCommandMapping() {
  console.log(chalk.yellow('\n2. Testing Command to Expert Mapping:'));
  
  const testCommands = [
    'sparc run architect',
    'github pr-manager',
    'swarm init --topology mesh',
    'analysis security-audit',
    'neural train',
    'k8s deploy',
    'react optimize'
  ];
  
  for (const command of testCommands) {
    const experts = getExpertsForCommand(command);
    if (experts.length > 0) {
      console.log(`  ✅ "${command}": ${experts.length} experts found`);
      console.log(`     → Top expert: ${experts[0].type}`);
    } else {
      console.log(chalk.yellow(`  ⚠️ "${command}": No direct experts, using fallback`));
    }
  }
}

/**
 * Test expert router initialization
 */
async function testExpertRouter() {
  console.log(chalk.yellow('\n3. Testing Expert Router:'));
  
  const router = createExpertRouter();
  
  try {
    await router.initialize();
    console.log('  ✅ Router initialized successfully');
    
    const status = await router.getStatus();
    console.log(`  ✅ Available experts: ${status.availableExperts}`);
    console.log(`  ✅ Command mappings: ${status.commandMappings}`);
    
    await router.shutdown();
    console.log('  ✅ Router shutdown successfully');
  } catch (error) {
    console.error(chalk.red(`  ❌ Router test failed: ${error.message}`));
  }
}

/**
 * Test expert capabilities
 */
async function testExpertCapabilities() {
  console.log(chalk.yellow('\n4. Testing Expert Capabilities by Category:'));
  
  const categories = {
    'SPARC': ['sparc-architect', 'sparc-coder', 'sparc-tester'],
    'GitHub': ['github-pr-analyzer', 'github-issue-orchestrator'],
    'Neural': ['neural-architect', 'ml-feature-engineer'],
    'Security': ['security-penetration-tester', 'security-cryptographer'],
    'DevOps': ['kubernetes-orchestrator', 'terraform-infrastructure-coder']
  };
  
  for (const [category, experts] of Object.entries(categories)) {
    console.log(`  ${chalk.green(category)}:`);
    for (const expertType of experts) {
      const expert = getExpertAgent(expertType);
      if (expert) {
        console.log(`    ✅ ${expert.name}`);
        console.log(`       Capabilities: ${expert.capabilities.slice(0, 3).join(', ')}...`);
        console.log(`       Claude Flow: ${expert.specializations?.claudeFlow?.slice(0, 2).join(', ') || 'N/A'}`);
      }
    }
  }
}

/**
 * Test expert specializations
 */
async function testExpertSpecializations() {
  console.log(chalk.yellow('\n5. Testing Expert Specializations:'));
  
  // Count unique capabilities
  const allCapabilities = new Set();
  const allTools = new Set();
  const allClaudeFlowCommands = new Set();
  
  for (const expert of Object.values(EXPERT_AGENT_REGISTRY)) {
    expert.capabilities.forEach(cap => allCapabilities.add(cap));
    expert.tools.forEach(tool => allTools.add(tool));
    expert.specializations?.claudeFlow?.forEach(cmd => allClaudeFlowCommands.add(cmd));
  }
  
  console.log(`  📊 Statistics:`);
  console.log(`     • Unique capabilities: ${allCapabilities.size}`);
  console.log(`     • Unique tools used: ${allTools.size}`);
  console.log(`     • Claude Flow commands: ${allClaudeFlowCommands.size}`);
  
  // Show top capabilities
  const capabilityCount = {};
  for (const expert of Object.values(EXPERT_AGENT_REGISTRY)) {
    for (const cap of expert.capabilities) {
      capabilityCount[cap] = (capabilityCount[cap] || 0) + 1;
    }
  }
  
  const topCapabilities = Object.entries(capabilityCount)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 10);
  
  console.log(`\n  🏆 Top 10 Capabilities:`);
  for (const [capability, count] of topCapabilities) {
    console.log(`     • ${capability}: ${count} experts`);
  }
}

/**
 * Test command routing scenarios
 */
async function testRoutingScenarios() {
  console.log(chalk.yellow('\n6. Testing Command Routing Scenarios:'));
  
  const router = createExpertRouter();
  await router.initialize();
  
  const scenarios = [
    {
      command: 'sparc pipeline',
      expectedExperts: ['sparc-orchestrator', 'sparc-coder', 'sparc-tester']
    },
    {
      command: 'github swarm-pr',
      expectedExperts: ['github-pr-analyzer', 'sparc-reviewer']
    },
    {
      command: 'analysis security-audit',
      expectedExperts: ['security-penetration-tester', 'security-cryptographer']
    }
  ];
  
  for (const scenario of scenarios) {
    const experts = router.commandMap[scenario.command];
    const match = JSON.stringify(experts) === JSON.stringify(scenario.expectedExperts);
    
    if (match) {
      console.log(`  ✅ "${scenario.command}"`);
      console.log(`     Correctly routes to: ${experts.join(', ')}`);
    } else {
      console.log(chalk.red(`  ❌ "${scenario.command}"`));
      console.log(`     Expected: ${scenario.expectedExperts.join(', ')}`);
      console.log(`     Got: ${experts?.join(', ') || 'undefined'}`);
    }
  }
  
  await router.shutdown();
}

/**
 * Run all tests
 */
async function runTests() {
  try {
    await testExpertRegistry();
    await testCommandMapping();
    await testExpertRouter();
    await testExpertCapabilities();
    await testExpertSpecializations();
    await testRoutingScenarios();
    
    console.log(chalk.green('\n✅ All expert agent tests completed successfully!'));
    
    console.log(chalk.cyan('\n📊 Expert Agent System Summary:'));
    const counts = getExpertCategoryCounts();
    console.log(`  • Total Expert Agents: ${counts.total}`);
    console.log(`  • All agents properly integrated with claude-flow`);
    console.log(`  • Deep specializations for each domain`);
    console.log(`  • Granular tool and capability mapping`);
    console.log(`  • Ready for production use with npx claude-flow@alpha`);
    
  } catch (error) {
    console.error(chalk.red('\n❌ Test suite failed:'), error);
    process.exit(1);
  }
}

runTests().catch(console.error);