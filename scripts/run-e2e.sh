#!/bin/bash
##
# E2E Test Runner with Environment Defaults
# Sets default values for environment variables if not already set
# Properly forwards all CLI arguments to Playwright
##

# Set defaults only if not already set
: "${BASE_URL:=http://localhost:8080}"
: "${API_BASE_URL:=http://localhost:11434}"
: "${BACKEND_UP:=1}"

# Export for child processes
export BASE_URL
export API_BASE_URL
export BACKEND_UP

# Forward all arguments to Playwright using "$@" for proper quoting
# This ensures args like --project=chromium are passed correctly
exec npx playwright test tests/e2e "$@"
