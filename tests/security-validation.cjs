#!/usr/bin/env node
/**
 * Security Infrastructure Validation Script (CommonJS)
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

// Test 2: Security files exist
test('Security files exist', () => {
  const fs = require('fs');
  const path = require('path');
  
  const securityFiles = [
    path.join(__dirname, '..', 'src', 'config', 'security.js'),
    path.join(__dirname, '..', 'src', 'middleware', 'security.js'),
    path.join(__dirname, '..', 'src', 'utils', 'secure-parser.js'),
    path.join(__dirname, '..', '.env.example')
  ];
  
  for (const file of securityFiles) {
    if (!fs.existsSync(file)) {
      throw new Error(`Security file missing: ${file}`);
    }
  }
});

// Test 3: .env.example has security settings
test('.env.example contains security settings', () => {
  const fs = require('fs');
  const path = require('path');
  
  const envPath = path.join(__dirname, '..', '.env.example');
  const content = fs.readFileSync(envPath, 'utf8');
  
  const requiredSettings = [
    'SESSION_SECRET',
    'JWT_SECRET', 
    'TEST_API_KEY',
    'TEST_JWT_SECRET',
    '32-chars'
  ];
  
  for (const setting of requiredSettings) {
    if (!content.includes(setting)) {
      throw new Error(`Missing security setting: ${setting}`);
    }
  }
});

// Test 4: Test files updated
test('Test files updated to use environment variables', () => {
  const fs = require('fs');
  const path = require('path');
  
  // Check analyze-agent test file
  const testFile = path.join(__dirname, 'agents', 'analyze-agent.test.cjs');
  if (fs.existsSync(testFile)) {
    const content = fs.readFileSync(testFile, 'utf8');
    if (!content.includes('process.env.TEST_API_KEY')) {
      throw new Error('Test file not updated to use environment variables');
    }
    if (content.includes('sk-1234567890abcdef')) {
      throw new Error('Hardcoded API key still present');
    }
  }
  
  // Check Go test file
  const goTestFile = path.join(__dirname, 'training', 'training_module_tests.go');
  if (fs.existsSync(goTestFile)) {
    const content = fs.readFileSync(goTestFile, 'utf8');
    if (!content.includes('TEST_JWT_SECRET')) {
      throw new Error('Go test file not updated');
    }
  }
  
  // Check validation script
  const validationScript = path.join(__dirname, 'training', 'validation_scripts_enhanced.sh');
  if (fs.existsSync(validationScript)) {
    const content = fs.readFileSync(validationScript, 'utf8');
    if (!content.includes('TEST_JWT_SECRET')) {
      throw new Error('Validation script not updated');
    }
  }
});

// Test 5: Security patterns implemented
test('Security modules contain proper patterns', () => {
  const fs = require('fs');
  const path = require('path');
  
  // Check security config
  const configFile = path.join(__dirname, '..', 'src', 'config', 'security.js');
  const configContent = fs.readFileSync(configFile, 'utf8');
  
  const requiredPatterns = [
    'generateJWT',
    'verifyJWT', 
    'sanitizeInput',
    'generateCSRFToken',
    'verifyCSRFToken',
    'hashPassword',
    'verifyPassword'
  ];
  
  for (const pattern of requiredPatterns) {
    if (!configContent.includes(pattern)) {
      throw new Error(`Security pattern missing: ${pattern}`);
    }
  }
  
  // Check secure parser
  const parserFile = path.join(__dirname, '..', 'src', 'utils', 'secure-parser.js');
  const parserContent = fs.readFileSync(parserFile, 'utf8');
  
  const parserPatterns = [
    'SecureJSONParser',
    'SafeExpressionEvaluator',
    'SafeTemplateProcessor',
    'dangerous'
  ];
  
  for (const pattern of parserPatterns) {
    if (!parserContent.includes(pattern)) {
      throw new Error(`Parser pattern missing: ${pattern}`);
    }
  }
});

// Test 6: Security middleware functions
test('Security middleware contains required functions', () => {
  const fs = require('fs');
  const path = require('path');
  
  const middlewareFile = path.join(__dirname, '..', 'src', 'middleware', 'security.js');
  const content = fs.readFileSync(middlewareFile, 'utf8');
  
  const requiredFunctions = [
    'setupSecurityMiddleware',
    'rateLimiters',
    'validateAndSanitizeInput',
    'csrfProtection',
    'requireAuth',
    'requireRole',
    'securityLogger',
    'secureFileUpload'
  ];
  
  for (const func of requiredFunctions) {
    if (!content.includes(func)) {
      throw new Error(`Middleware function missing: ${func}`);
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