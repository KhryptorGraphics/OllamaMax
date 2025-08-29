#!/bin/bash
# Ollama Distributed User Track Assessment Validation System
# This script validates practical assessment tasks automatically

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Score tracking
TOTAL_SCORE=0
MAX_SCORE=60

echo "═══════════════════════════════════════════════════════════════"
echo "   Ollama Distributed User Track - Practical Assessment"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Function to check command existence
check_command() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to add score
add_score() {
    local points=$1
    local description=$2
    TOTAL_SCORE=$((TOTAL_SCORE + points))
    echo -e "${GREEN}✓${NC} $description (+$points points)"
}

# Function to deduct score
deduct_score() {
    local description=$1
    echo -e "${RED}✗${NC} $description (0 points)"
}

echo "Starting Assessment Validation..."
echo "================================="
echo ""

# Task 1: Installation and Setup (15 points)
echo "Task 1: Installation and Setup Validation"
echo "-----------------------------------------"

# Check if ollama-distributed exists
if check_command ollama-distributed; then
    add_score 5 "Ollama Distributed installed correctly"
    
    # Check version
    if ollama-distributed --version &> /dev/null; then
        add_score 3 "Version check successful"
    else
        deduct_score "Version check failed"
    fi
else
    # Check in local directory
    if [ -f "./bin/ollama-distributed" ]; then
        add_score 3 "Ollama Distributed binary found locally"
        
        # Check version
        if ./bin/ollama-distributed --version &> /dev/null; then
            add_score 3 "Version check successful"
        else
            deduct_score "Version check failed"
        fi
    else
        deduct_score "Ollama Distributed not installed"
    fi
fi

# Check configuration file
if [ -f "$HOME/.ollama-distributed/config.yaml" ]; then
    add_score 3 "Configuration file exists"
    
    # Validate configuration
    if ollama-distributed validate &> /dev/null || ./bin/ollama-distributed validate &> /dev/null; then
        add_score 2 "Configuration is valid"
    else
        deduct_score "Configuration validation failed"
    fi
else
    deduct_score "Configuration file not found"
fi

# Check if node can start (we'll just check the command exists)
if ollama-distributed help start &> /dev/null || ./bin/ollama-distributed help start &> /dev/null; then
    add_score 2 "Start command available"
else
    deduct_score "Start command not available"
fi

echo ""

# Task 2: Monitoring Tool (15 points)
echo "Task 2: Monitoring Tool Validation"
echo "----------------------------------"

# Check for monitoring script
MONITOR_SCRIPT=""
if [ -f "monitoring-tool.sh" ]; then
    MONITOR_SCRIPT="monitoring-tool.sh"
elif [ -f "$HOME/monitoring-tool.sh" ]; then
    MONITOR_SCRIPT="$HOME/monitoring-tool.sh"
elif [ -f "./scripts/monitoring-tool.sh" ]; then
    MONITOR_SCRIPT="./scripts/monitoring-tool.sh"
fi

if [ -n "$MONITOR_SCRIPT" ]; then
    add_score 5 "Monitoring script found: $MONITOR_SCRIPT"
    
    # Check script content
    if grep -q "health" "$MONITOR_SCRIPT"; then
        add_score 3 "Health check functionality present"
    else
        deduct_score "Health check functionality missing"
    fi
    
    if grep -q "status\|Status\|STATUS" "$MONITOR_SCRIPT"; then
        add_score 3 "Status display functionality present"
    else
        deduct_score "Status display functionality missing"
    fi
    
    if grep -q "sleep\|while\|loop" "$MONITOR_SCRIPT"; then
        add_score 2 "Auto-refresh capability present"
    else
        deduct_score "Auto-refresh capability missing"
    fi
    
    if grep -q "error\|Error\|ERROR\|catch\|trap" "$MONITOR_SCRIPT"; then
        add_score 2 "Error handling present"
    else
        deduct_score "Error handling missing"
    fi
else
    deduct_score "Monitoring script not found"
    echo -e "${YELLOW}Hint: Create monitoring-tool.sh in current directory${NC}"
fi

echo ""

# Task 3: Dashboard Navigation (10 points)
echo "Task 3: Dashboard Navigation Check"
echo "----------------------------------"

# Check if user knows the dashboard URL
echo -e "${YELLOW}Manual Check Required:${NC}"
echo "1. Can you access http://localhost:8081? (2 points)"
echo "2. Can you navigate all sections? (2 points)"
echo "3. Can you find configuration? (2 points)"
echo "4. Can you locate metrics? (2 points)"
echo "5. Can you find help resources? (2 points)"
echo ""
echo "Please self-assess and add 0-10 points to your score."
echo ""

# Task 4: Diagnostic Report (10 points)
echo "Task 4: Diagnostic Report Validation"
echo "------------------------------------"

# Check for diagnostic script
DIAG_SCRIPT=""
if [ -f "diagnostic-report.sh" ]; then
    DIAG_SCRIPT="diagnostic-report.sh"
elif [ -f "$HOME/diagnostic-report.sh" ]; then
    DIAG_SCRIPT="$HOME/diagnostic-report.sh"
elif [ -f "./scripts/diagnostic-report.sh" ]; then
    DIAG_SCRIPT="./scripts/diagnostic-report.sh"
fi

if [ -n "$DIAG_SCRIPT" ]; then
    add_score 2 "Diagnostic script found: $DIAG_SCRIPT"
    
    # Check script content
    if grep -q "uname\|system\|System" "$DIAG_SCRIPT"; then
        add_score 2 "System information collection present"
    else
        deduct_score "System information collection missing"
    fi
    
    if grep -q "status\|Status" "$DIAG_SCRIPT"; then
        add_score 2 "Status capture present"
    else
        deduct_score "Status capture missing"
    fi
    
    if grep -q "config\|Config\|configuration" "$DIAG_SCRIPT"; then
        add_score 2 "Configuration capture present"
    else
        deduct_score "Configuration capture missing"
    fi
    
    if grep -q "log\|Log\|LOG" "$DIAG_SCRIPT"; then
        add_score 2 "Log capture present"
    else
        deduct_score "Log capture missing"
    fi
else
    deduct_score "Diagnostic script not found"
    echo -e "${YELLOW}Hint: Create diagnostic-report.sh in current directory${NC}"
fi

echo ""

# Task 5: Troubleshooting (10 points)
echo "Task 5: Troubleshooting Knowledge Check"
echo "---------------------------------------"

# Create a simple troubleshooting quiz
echo "Please answer the following troubleshooting questions:"
echo ""

echo "1. How would you check what's using port 8080?"
echo "   Expected: lsof -i :8080 or netstat -tuln | grep 8080"
read -p "   Your answer: " answer1

if [[ "$answer1" == *"lsof"* ]] || [[ "$answer1" == *"netstat"* ]]; then
    add_score 2 "Correct port conflict diagnosis"
else
    deduct_score "Incorrect port conflict diagnosis"
fi

echo ""
echo "2. What command validates the configuration?"
echo "   Expected: ollama-distributed validate"
read -p "   Your answer: " answer2

if [[ "$answer2" == *"validate"* ]]; then
    add_score 2 "Correct validation command"
else
    deduct_score "Incorrect validation command"
fi

echo ""
echo "3. Where are the log files stored?"
echo "   Expected: ~/.ollama-distributed/logs/"
read -p "   Your answer: " answer3

if [[ "$answer3" == *".ollama-distributed/logs"* ]]; then
    add_score 2 "Correct log location"
else
    deduct_score "Incorrect log location"
fi

echo ""
echo "4. What status code means 'Not Implemented'?"
echo "   Expected: 501"
read -p "   Your answer: " answer4

if [[ "$answer4" == "501" ]]; then
    add_score 2 "Correct status code knowledge"
else
    deduct_score "Incorrect status code knowledge"
fi

echo ""
echo "5. What's the minimum Go version required?"
echo "   Expected: 1.21 or 1.21+"
read -p "   Your answer: " answer5

if [[ "$answer5" == *"1.21"* ]]; then
    add_score 2 "Correct Go version requirement"
else
    deduct_score "Incorrect Go version requirement"
fi

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "                    ASSESSMENT COMPLETE"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Calculate percentage
PERCENTAGE=$((TOTAL_SCORE * 100 / MAX_SCORE))

# Determine pass/fail
if [ $PERCENTAGE -ge 70 ]; then
    echo -e "${GREEN}✅ PRACTICAL ASSESSMENT: PASSED${NC}"
    echo ""
    echo "Your Score: $TOTAL_SCORE / $MAX_SCORE ($PERCENTAGE%)"
    
    if [ $PERCENTAGE -ge 90 ]; then
        echo "Level: DISTINCTION (Gold Badge)"
    elif [ $PERCENTAGE -ge 80 ]; then
        echo "Level: MERIT (Silver Badge)"
    else
        echo "Level: PASS (Bronze Badge)"
    fi
else
    echo -e "${RED}❌ PRACTICAL ASSESSMENT: NOT PASSED${NC}"
    echo ""
    echo "Your Score: $TOTAL_SCORE / $MAX_SCORE ($PERCENTAGE%)"
    echo "Required: 42 / 60 (70%)"
    echo ""
    echo "Please review the materials and try again after 7 days."
fi

echo ""
echo "Note: Add manual dashboard navigation score (0-10) to final total."
echo ""

# Generate certificate data if passed
if [ $PERCENTAGE -ge 70 ]; then
    CERT_ID=$(uuidgen 2>/dev/null || echo "CERT-$(date +%s)")
    echo "Certificate Data:"
    echo "=================="
    echo "ID: $CERT_ID"
    echo "Date: $(date +%Y-%m-%d)"
    echo "Valid Until: $(date -d '+18 months' +%Y-%m-%d 2>/dev/null || date +%Y-%m-%d)"
    echo "Score: $TOTAL_SCORE / $MAX_SCORE"
    echo ""
    echo "Save this information for your certificate."
fi

echo ""
echo "Assessment validation complete. Good luck!"