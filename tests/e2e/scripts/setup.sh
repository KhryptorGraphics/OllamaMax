#!/bin/bash

# OllamaMax E2E Test Setup Script
# Automated setup for the comprehensive testing infrastructure

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo -e "${BLUE}🛠️  OllamaMax E2E Test Setup${NC}"
echo "================================"

# Check Node.js version
print_status "Checking Node.js version..."
NODE_VERSION=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 16 ]; then
    print_error "Node.js 16+ required. Found version: $(node --version)"
    exit 1
fi
print_success "Node.js version check passed: $(node --version)"

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    print_error "package.json not found. Please run this script from the tests/e2e directory"
    exit 1
fi

# Install dependencies
print_status "Installing npm dependencies..."
if npm ci; then
    print_success "Dependencies installed successfully"
else
    print_warning "npm ci failed, trying npm install..."
    if npm install; then
        print_success "Dependencies installed with npm install"
    else
        print_error "Failed to install dependencies"
        exit 1
    fi
fi

# Install Playwright browsers
print_status "Installing Playwright browsers..."
if npx playwright install --with-deps; then
    print_success "Playwright browsers installed"
else
    print_error "Failed to install Playwright browsers"
    exit 1
fi

# Create necessary directories
print_status "Creating directory structure..."
directories=(
    "reports"
    "reports/screenshots"
    "reports/performance"
    "reports/load-tests"
    "reports/security"
    "reports/playwright-report"
    "reports/allure-results"
    "reports/metrics"
    "test-results"
    "test-archives"
)

for dir in "${directories[@]}"; do
    mkdir -p "$dir"
done
print_success "Directory structure created"

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    print_status "Creating .env configuration file..."
    cat > .env << EOF
# OllamaMax E2E Test Configuration
BASE_URL=http://localhost:8080
NODE_ENV=test
HEADLESS=true
CI=false
DEBUG=false

# Performance test settings
RUN_PERFORMANCE=false
RUN_LOAD_TESTS=false

# Browser settings
BROWSERS=chromium,firefox,webkit
PARALLEL=true

# Test timeouts (milliseconds)
TEST_TIMEOUT=120000
ACTION_TIMEOUT=10000
NAVIGATION_TIMEOUT=30000

# Reporting
GENERATE_ALLURE=false
SAVE_SCREENSHOTS=true
SAVE_VIDEOS=on-failure
SAVE_TRACES=retain-on-failure
EOF
    print_success ".env file created"
else
    print_status ".env file already exists, skipping creation"
fi

# Create .gitignore additions for test artifacts
print_status "Updating .gitignore for test artifacts..."
cat >> .gitignore << EOF

# E2E Test artifacts
test-results/
reports/playwright-report/
reports/screenshots/
reports/performance/
reports/load-tests/
reports/security/
reports/allure-results/
reports/metrics/
test-archives/
*.log

# Environment files
.env.local
.env.test.local
EOF

# Validate TypeScript configuration
print_status "Validating TypeScript configuration..."
if npx tsc --noEmit; then
    print_success "TypeScript configuration is valid"
else
    print_warning "TypeScript validation found issues (this is normal for new setup)"
fi

# Run linting
print_status "Running initial code quality checks..."
if npm run lint; then
    print_success "Code quality checks passed"
else
    print_warning "Linting issues found - run 'npm run lint:fix' to auto-fix"
fi

# Test Playwright installation
print_status "Testing Playwright installation..."
if npx playwright --version > /dev/null 2>&1; then
    print_success "Playwright installation verified: $(npx playwright --version)"
else
    print_error "Playwright installation test failed"
    exit 1
fi

# Create a quick smoke test
print_status "Running smoke test..."
cat > smoke-test.js << 'EOF'
const { chromium } = require('playwright');

(async () => {
  try {
    const browser = await chromium.launch();
    const page = await browser.newPage();
    await page.goto('https://playwright.dev');
    const title = await page.title();
    await browser.close();
    
    if (title.includes('Playwright')) {
      console.log('✅ Smoke test passed');
      process.exit(0);
    } else {
      console.log('❌ Smoke test failed - unexpected title');
      process.exit(1);
    }
  } catch (error) {
    console.log('❌ Smoke test failed:', error.message);
    process.exit(1);
  }
})();
EOF

if node smoke-test.js; then
    print_success "Smoke test passed"
    rm smoke-test.js
else
    print_error "Smoke test failed"
    rm smoke-test.js
    exit 1
fi

# Generate initial documentation
print_status "Generating setup documentation..."
cat > SETUP.md << EOF
# OllamaMax E2E Testing - Setup Complete

## ✅ Installation Status
- Node.js: $(node --version)
- npm: $(npm --version)
- Playwright: $(npx playwright --version)
- Setup completed on: $(date)

## 🚀 Quick Start

### Run All Tests
\`\`\`bash
npm test
\`\`\`

### Run Specific Test Suites
\`\`\`bash
# Core functionality
npm run test:core

# Distributed inference
npm run test:inference

# Security tests
npm run test:security

# Performance tests (optional)
npm run test:performance

# Load tests (optional)
npm run test:load
\`\`\`

### View Reports
\`\`\`bash
# Open HTML report
npm run report

# Generate Allure report
npm run report:allure
\`\`\`

## 🔧 Configuration

Edit \`.env\` file to customize:
- BASE_URL: Target application URL
- Browsers to test
- Test timeouts
- Reporting options

## 📊 Available Scripts

- \`npm run test\` - Run all tests
- \`npm run test:headed\` - Run with visible browser
- \`npm run test:debug\` - Run in debug mode
- \`npm run test:ui\` - Interactive test runner
- \`npm run lint\` - Code quality checks
- \`npm run type-check\` - TypeScript validation

## 📁 Directory Structure

- \`tests/\` - Test specifications
- \`helpers/\` - Test utilities and helpers
- \`reports/\` - Generated test reports
- \`test-results/\` - Test execution artifacts

## 🆘 Troubleshooting

If tests fail to run:
1. Check if OllamaMax is running at the configured BASE_URL
2. Verify browser installations: \`npx playwright install\`
3. Check environment configuration in \`.env\`
4. Run setup again: \`./scripts/setup.sh\`

For detailed documentation, see README.md
EOF

print_success "Setup documentation generated (SETUP.md)"

# Final status
echo ""
echo -e "${GREEN}🎉 Setup Complete!${NC}"
echo "================================"
echo "✅ Dependencies installed"
echo "✅ Browsers configured"
echo "✅ Directory structure created"
echo "✅ Configuration files generated"
echo "✅ Code quality tools configured"
echo "✅ Documentation created"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Start your OllamaMax application"
echo "2. Update BASE_URL in .env if needed"
echo "3. Run tests: npm test"
echo "4. View results in reports/ directory"
echo ""
echo -e "${YELLOW}Quick Test:${NC}"
echo "npm run test:core"
echo ""
print_success "Ready for comprehensive E2E testing! 🚀"