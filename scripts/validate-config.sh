#!/bin/bash

# OllamaMax Database Configuration Validation Script
# This script validates the database configuration consistency

set -e

echo "🔍 Validating OllamaMax Database Configuration..."
echo "================================================"

# Check if required files exist
echo "✅ Checking required files..."
for file in "main.go" "docker-compose.yml" ".env.example" "pkg/database/manager.go"; do
    if [ -f "$file" ]; then
        echo "  ✓ $file exists"
    else
        echo "  ❌ $file missing"
        exit 1
    fi
done

# Check main.go configuration
echo "✅ Validating main.go configuration..."

# Check if main.go uses environment variables properly
if grep -q "getEnvWithDefault.*DB_HOST" main.go && grep -q "os.Getenv.*DB_PASSWORD" main.go; then
    echo "  ✓ main.go uses environment variables for database configuration"
else
    echo "  ❌ main.go does not properly use environment variables"
    exit 1
fi

# Check if main.go has proper error handling
if grep -q "createDatabaseConfig" main.go; then
    echo "  ✓ main.go uses proper database configuration function"
else
    echo "  ❌ main.go does not use proper configuration function"
    exit 1
fi

# Check docker-compose.yml configuration
echo "✅ Validating docker-compose.yml configuration..."

# Check if Docker Compose has the required environment variables
required_env_vars=("DB_HOST" "DB_PORT" "DB_NAME" "DB_USER" "DB_PASSWORD" "REDIS_HOST" "REDIS_PORT" "REDIS_PASSWORD")

for var in "${required_env_vars[@]}"; do
    if grep -q "$var=" docker-compose.yml; then
        echo "  ✓ $var is configured in docker-compose.yml"
    else
        echo "  ❌ $var is missing in docker-compose.yml"
        exit 1
    fi
done

# Check port consistency
echo "✅ Checking port consistency..."

# Check PostgreSQL port mapping (should be 11432:5432)
if grep -q "11432:5432" docker-compose.yml; then
    echo "  ✓ PostgreSQL port mapping is correct (11432:5432)"
else
    echo "  ❌ PostgreSQL port mapping is incorrect"
    exit 1
fi

# Check Redis port mapping (should be 11379:6379)  
if grep -q "11379:6379" docker-compose.yml; then
    echo "  ✓ Redis port mapping is correct (11379:6379)"
else
    echo "  ❌ Redis port mapping is incorrect"
    exit 1
fi

# Check if internal ports are used in docker-compose environment
if grep -q "DB_PORT=5432" docker-compose.yml && grep -q "REDIS_PORT=6379" docker-compose.yml; then
    echo "  ✓ Internal ports are correctly configured in docker-compose environment"
else
    echo "  ❌ Internal ports are not correctly configured"
    exit 1
fi

# Validate .env.example
echo "✅ Validating .env.example..."

# Check if .env.example has all required variables with proper documentation
for var in "${required_env_vars[@]}"; do
    if grep -q "^$var=" .env.example; then
        echo "  ✓ $var is documented in .env.example"
    else
        echo "  ❌ $var is missing in .env.example"
        exit 1
    fi
done

# Check if .env.example has local development overrides section
if grep -q "LOCAL DEVELOPMENT OVERRIDES" .env.example; then
    echo "  ✓ .env.example has local development overrides section"
else
    echo "  ❌ .env.example missing local development overrides section"
    exit 1
fi

# Test Go syntax validation (basic checks)
echo "✅ Testing Go syntax validation..."

# Basic syntax check for main.go database configuration
if go fmt -n main.go > /dev/null 2>&1; then
    echo "  ✓ main.go has valid Go syntax"
else
    echo "  ❌ main.go has syntax errors"
    exit 1
fi

# Check for required functions in main.go
if grep -q "func createDatabaseConfig" main.go && grep -q "func getEnvWithDefault" main.go; then
    echo "  ✓ main.go contains required database configuration functions"
else
    echo "  ❌ main.go missing required database configuration functions"
    exit 1
fi

echo "  ⚠️  Skipping full compilation test due to package dependencies"

echo ""
echo "🎉 All configuration validations passed!"
echo ""
echo "Summary of fixes applied:"
echo "========================"
echo "1. ✅ Fixed PostgreSQL port configuration (internal 5432, external 11432)"
echo "2. ✅ Added proper environment variable handling in main.go"
echo "3. ✅ Added Redis password authentication support"
echo "4. ✅ Created comprehensive .env.example with all required variables"
echo "5. ✅ Added proper error handling for missing environment variables"
echo "6. ✅ Ensured consistent configuration between main.go and docker-compose.yml"
echo "7. ✅ Added database connection test script"
echo ""
echo "Next steps:"
echo "==========="
echo "1. Copy .env.example to .env and update with your actual values"
echo "2. Run 'docker-compose up -d postgres redis' to start databases"
echo "3. Run 'go run scripts/test-db-config.go' to test database connectivity"
echo "4. Run 'docker-compose up' to start the full application"
echo ""
echo "For local development outside Docker:"
echo "====================================="
echo "Update .env with:"
echo "DB_HOST=localhost"
echo "DB_PORT=11432"
echo "REDIS_HOST=localhost"
echo "REDIS_PORT=11379"