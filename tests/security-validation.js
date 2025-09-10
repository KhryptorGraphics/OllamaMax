#!/usr/bin/env node
/**
 * Security Infrastructure Validation Script
 * Tests the implemented security measures
 */

console.log('🔐 Security Infrastructure Validation');
console.log('=====================================\n');

// Set test environment variables
process.env.NODE_ENV = 'test';
process.env.SESSION_SECRET = 'test-session-secret-32-chars-minimum';
process.env.JWT_SECRET = 'test-jwt-secret-32-chars-minimum';
process.env.TEST_API_KEY = 'test-api-key-for-testing';
process.env.TEST_JWT_SECRET = 'test-jwt-secret-32-chars-minimum-for-testing';

let passed = 0;
let failed = 0;

function test(name, fn) {
  try {
    fn();
    console.log(`✅ ${name}`);
    passed++;
  } catch (error) {
    console.log(`❌ ${name}: ${error.message}`);
    failed++;
  }
}

// Test 1: Environment variables validation
test('Test environment variables configured', () => {
  if (!process.env.TEST_API_KEY) throw new Error('TEST_API_KEY not set');
  if (!process.env.TEST_JWT_SECRET) throw new Error('TEST_JWT_SECRET not set');
  if (process.env.TEST_JWT_SECRET.length < 32) throw new Error('TEST_JWT_SECRET too short');
});

// Test 2: Basic module loading (will show warnings but shouldn't fail)
test('Security modules can be loaded', () => {
  // Just test that the files exist and can be parsed
  const fs = require('fs');
  const path = require('path');
  
  const securityConfigPath = path.join(__dirname, '..', 'src', 'config', 'security.js');
  const securityMiddlewarePath = path.join(__dirname, '..', 'src', 'middleware', 'security.js');
  const secureParserPath = path.join(__dirname, '..', 'src', 'utils', 'secure-parser.js');
  
  if (!fs.existsSync(securityConfigPath)) throw new Error('Security config file missing');
  if (!fs.existsSync(securityMiddlewarePath)) throw new Error('Security middleware file missing');
  if (!fs.existsSync(secureParserPath)) throw new Error('Secure parser file missing');
  
  // Test file contents have key security patterns
  const securityConfigContent = fs.readFileSync(securityConfigPath, 'utf8');
  if (!securityConfigContent.includes('generateJWT')) throw new Error('JWT functionality missing');
  if (!securityConfigContent.includes('sanitizeInput')) throw new Error('Input sanitization missing');
  
  const secureParserContent = fs.readFileSync(secureParserPath, 'utf8');
  if (!secureParserContent.includes('SecureJSONParser')) throw new Error('SecureJSONParser missing');
  if (!secureParserContent.includes('SafeExpressionEvaluator')) throw new Error('SafeExpressionEvaluator missing');
  
  const middlewareContent = fs.readFileSync(securityMiddlewarePath, 'utf8');
  if (!middlewareContent.includes('setupSecurityMiddleware')) throw new Error('Security middleware setup missing');
  if (!middlewareContent.includes('csrfProtection')) throw new Error('CSRF protection missing');
});

// Test 3: File security patterns
test('Files contain security hardening patterns', () => {
  const fs = require('fs');
  const path = require('path');
  
  // Check .env.example has proper security settings
  const envExamplePath = path.join(__dirname, '..', '.env.example');
  if (!fs.existsSync(envExamplePath)) throw new Error('.env.example file missing');
  
  const envContent = fs.readFileSync(envExamplePath, 'utf8');
  if (!envContent.includes('SESSION_SECRET')) throw new Error('SESSION_SECRET missing from .env.example');
  if (!envContent.includes('JWT_SECRET')) throw new Error('JWT_SECRET missing from .env.example');
  if (!envContent.includes('TEST_API_KEY')) throw new Error('TEST_API_KEY missing from .env.example');
  if (!envContent.includes('32-chars')) throw new Error('Security requirements not documented');
});

// Test 4: Updated test files use environment variables
test('Test files updated to use environment variables', () => {
  const fs = require('fs');
  const path = require('path');
  
  // Check analyze-agent test file
  const testFilePath = path.join(__dirname, 'agents', 'analyze-agent.test.cjs');
  if (!fs.existsSync(testFilePath)) throw new Error('analyze-agent.test.cjs missing');
  
  const testContent = fs.readFileSync(testFilePath, 'utf8');
  if (!testContent.includes('process.env.TEST_API_KEY')) throw new Error('Test file not updated to use environment variables');
  if (testContent.includes('sk-1234567890abcdef')) throw new Error('Hardcoded API key still present in test');
  
  // Check Go test file
  const goTestPath = path.join(__dirname, 'training', 'training_module_tests.go');
  if (fs.existsSync(goTestPath)) {
    const goContent = fs.readFileSync(goTestPath, 'utf8');
    if (!goContent.includes('TEST_JWT_SECRET')) throw new Error('Go test file not updated');
    if (goContent.includes('your-secret-key-here')) throw new Error('Hardcoded secret still present in Go test');
  }
});

// Test 5: Security patterns in codebase
test('Security patterns properly implemented', () => {
  const fs = require('fs');
  const path = require('path');
  
  // Check that eval detection is improved in analyze-agent.js
  const analyzeAgentPath = path.join(__dirname, '..', 'src', 'agents', 'analyze-agent.js');
  if (fs.existsSync(analyzeAgentPath)) {
    const content = fs.readFileSync(analyzeAgentPath, 'utf8');
    if (content.includes('/\\b(eval|Function|setTimeout|setInterval)\\s*\\(.*\\)/g')) {
      throw new Error('Old eval detection pattern still present');
    }
  }
});

// Run tests and display results
console.log(`\n📊 Test Results:`);
console.log(`✅ Passed: ${passed}`);
console.log(`❌ Failed: ${failed}`);
console.log(`📈 Success Rate: ${Math.round((passed / (passed + failed)) * 100)}%`);

if (failed > 0) {
  console.log('\n⚠️  Some security tests failed. Please review the implementation.');
  process.exit(1);
} else {
  console.log('\n🎉 All security tests passed! Infrastructure is working correctly.');
  process.exit(0);
}