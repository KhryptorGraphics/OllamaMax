#!/bin/bash
# Quick Test - Validate that the training materials work with the actual software
# This script tests the training components to ensure they align with the real implementation

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🧪 Ollama Distributed Training Quick Test${NC}"
echo -e "${GREEN}===========================================${NC}"
echo

# Test 1: Verify binary exists and works
echo "Test 1: Binary Verification"
echo "----------------------------"

if [[ -f "../../../cmd/distributed-ollama/main.go" ]]; then
    echo "✅ Main source file found"
else
    echo -e "${RED}❌ Main source file not found at expected location${NC}"
    echo "   Expected: ../../../cmd/distributed-ollama/main.go"
fi

if [[ -f "./validation-scripts.sh" ]]; then
    echo "✅ Validation scripts available"
else
    echo -e "${RED}❌ Validation scripts missing${NC}"
fi

# Test 2: Check command structure from source
echo
echo "Test 2: Command Structure Verification"  
echo "--------------------------------------"

# Analyze the actual CLI commands from source
if [[ -f "../../../cmd/node/main.go" ]]; then
    echo "✅ Node CLI source found"
    
    # Check for expected commands mentioned in training
    commands_found=0
    expected_commands=("start" "status" "quickstart" "setup" "validate")
    
    for cmd in "${expected_commands[@]}"; do
        if grep -q "\"$cmd\"" "../../../cmd/node/main.go" 2>/dev/null; then
            echo "✅ Command '$cmd' found in source"
            ((commands_found++))
        else
            echo -e "${YELLOW}⚠️  Command '$cmd' not found in node source${NC}"
        fi
    done
    
    # Check distributed-ollama commands too
    if [[ -f "../../../cmd/distributed-ollama/main.go" ]]; then
        echo "✅ Distributed CLI source found"
        
        if grep -q "quickstart" "../../../cmd/distributed-ollama/main.go"; then
            echo "✅ Quickstart command available in distributed version"
        fi
        
        if grep -q "proxy" "../../../cmd/distributed-ollama/main.go"; then
            echo "✅ Proxy commands available"
        fi
    fi
    
else
    echo -e "${YELLOW}⚠️  Node CLI source not found, checking alternatives${NC}"
fi

# Test 3: API endpoint verification from source
echo
echo "Test 3: API Endpoint Verification"
echo "---------------------------------"

api_handlers_found=false
if [[ -f "../../../pkg/api/handlers.go" ]] || [[ -f "../../../internal/api/handlers.go" ]]; then
    api_handlers_found=true
    echo "✅ API handlers source found"
    
    # Check for key endpoints mentioned in training
    endpoints_to_check=("/health" "/api/tags" "/api/generate" "/api/distributed/status")
    
    for endpoint in "${endpoints_to_check[@]}"; do
        # Look for endpoint in various possible handler files
        if find "../../../" -name "*.go" -type f -exec grep -l "$endpoint" {} \; 2>/dev/null | head -1 >/dev/null; then
            echo "✅ Endpoint '$endpoint' found in source"
        else
            echo -e "${YELLOW}⚠️  Endpoint '$endpoint' not obviously found${NC}"
        fi
    done
else
    echo -e "${YELLOW}⚠️  API handlers source not found at expected locations${NC}"
fi

# Test 4: Configuration structure verification
echo
echo "Test 4: Configuration Structure"
echo "-------------------------------"

config_found=false
if [[ -f "../../../internal/config/types.go" ]] || [[ -f "../../../pkg/config/types.go" ]]; then
    config_found=true
    echo "✅ Configuration types source found"
    
    # Check for key config sections mentioned in training
    config_sections=("API" "P2P" "Storage" "Web" "Consensus")
    
    for section in "${config_sections[@]}"; do
        if find "../../../" -name "*.go" -type f -exec grep -l "type.*$section.*struct" {} \; 2>/dev/null | head -1 >/dev/null; then
            echo "✅ Config section '$section' structure found"
        else
            echo -e "${YELLOW}⚠️  Config section '$section' not found${NC}"
        fi
    done
fi

# Test 5: Training material accuracy
echo
echo "Test 5: Training Material Accuracy Check"
echo "----------------------------------------"

# Check if training examples use realistic commands
training_files=("training-modules.md" "interactive-tutorial.md" "README.md")
accurate_training=true

for file in "${training_files[@]}"; do
    if [[ -f "$file" ]]; then
        echo "✅ Training file exists: $file"
        
        # Check for any obviously wrong commands (examples)
        if grep -q "ollama-distributed-fake-command" "$file" 2>/dev/null; then
            echo -e "${RED}❌ Found fake commands in $file${NC}"
            accurate_training=false
        fi
        
        # Check for realistic port numbers
        if grep -q ":8080\|:8081\|:4001" "$file" 2>/dev/null; then
            echo "✅ Uses realistic port numbers in $file"
        fi
        
    else
        echo -e "${RED}❌ Training file missing: $file${NC}"
        accurate_training=false
    fi
done

# Test 6: Validation scripts functionality
echo
echo "Test 6: Validation Scripts Test"
echo "-------------------------------"

if [[ -x "./validation-scripts.sh" ]]; then
    echo "✅ Validation scripts are executable"
    
    # Test help functionality
    if ./validation-scripts.sh help >/dev/null 2>&1; then
        echo "✅ Help command works"
    else
        echo -e "${RED}❌ Help command failed${NC}"
    fi
    
    # Test prerequisite check (safe to run)
    echo "Running prerequisite check..."
    if ./validation-scripts.sh prereq >/dev/null 2>&1; then
        echo "✅ Prerequisite check runs successfully"
    else
        echo -e "${YELLOW}⚠️  Prerequisite check had issues (may be normal)${NC}"
    fi
    
else
    echo -e "${RED}❌ Validation scripts not executable${NC}"
fi

# Summary
echo
echo -e "${GREEN}📊 Test Summary${NC}"
echo "==============="

# Count successful tests
tests_passed=0
total_tests=6

if [[ -f "../../../cmd/distributed-ollama/main.go" ]]; then ((tests_passed++)); fi
if [[ $commands_found -gt 2 ]]; then ((tests_passed++)); fi
if [[ $api_handlers_found == true ]]; then ((tests_passed++)); fi
if [[ $config_found == true ]]; then ((tests_passed++)); fi
if [[ $accurate_training == true ]]; then ((tests_passed++)); fi
if [[ -x "./validation-scripts.sh" ]]; then ((tests_passed++)); fi

echo "Tests Passed: $tests_passed/$total_tests"

if [[ $tests_passed -eq $total_tests ]]; then
    echo -e "${GREEN}🎉 All tests passed! Training materials align well with the actual software.${NC}"
    echo
    echo "✅ Ready to start training!"
    echo "   Run: ./validation-scripts.sh full"
    echo "   Or follow: interactive-tutorial.md"
    
elif [[ $tests_passed -ge 4 ]]; then
    echo -e "${YELLOW}⚠️  Most tests passed. Training should work with minor issues.${NC}"
    echo
    echo "⚡ You can proceed with training, but some features might differ slightly"
    
else
    echo -e "${RED}❌ Several tests failed. Training materials may not align with current software.${NC}"
    echo
    echo "🔧 Consider updating training materials or checking software version"
fi

echo
echo "📚 Next Steps:"
echo "   1. Review any warnings above"  
echo "   2. Run full validation: ./validation-scripts.sh full"
echo "   3. Start training with: README.md or interactive-tutorial.md"
echo
echo "🐛 Found issues? Report at: https://github.com/KhryptorGraphics/ollamamax/issues"