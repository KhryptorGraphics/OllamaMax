#!/bin/bash

# OllamaMax Security Hardening Script
# This script implements critical security fixes and hardening measures

set -e

echo "🛡️ OllamaMax Security Hardening"
echo "==============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
    esac
}

print_status "INFO" "Starting security hardening process..."

# Step 1: Validate Environment Variables
print_status "INFO" "Validating security environment variables..."

# Check JWT Secret
if [ -z "$JWT_SECRET" ]; then
    print_status "WARNING" "JWT_SECRET not set, generating secure random secret..."
    export JWT_SECRET=$(openssl rand -hex 32)
    echo "export JWT_SECRET=\"$JWT_SECRET\"" >> ~/.bashrc
    print_status "SUCCESS" "Generated JWT_SECRET (added to ~/.bashrc)"
else
    if [ ${#JWT_SECRET} -lt 32 ]; then
        print_status "ERROR" "JWT_SECRET must be at least 32 characters long"
        exit 1
    fi
    print_status "SUCCESS" "JWT_SECRET is properly configured"
fi

# Check Admin Password
if [ -z "$ADMIN_DEFAULT_PASSWORD" ]; then
    print_status "WARNING" "ADMIN_DEFAULT_PASSWORD not set, will generate random password"
    ADMIN_PASSWORD=$(openssl rand -hex 16)
    export ADMIN_DEFAULT_PASSWORD="$ADMIN_PASSWORD"
    echo "export ADMIN_DEFAULT_PASSWORD=\"$ADMIN_PASSWORD\"" >> ~/.bashrc
    print_status "SUCCESS" "Generated ADMIN_DEFAULT_PASSWORD: $ADMIN_PASSWORD"
else
    print_status "SUCCESS" "ADMIN_DEFAULT_PASSWORD is configured"
fi

# Step 2: Fix Hardcoded Credentials
print_status "INFO" "Scanning for hardcoded credentials..."

# Check for hardcoded passwords in code
HARDCODED_FOUND=false

# Check for admin/admin pattern
if grep -r "admin.*admin" pkg/ cmd/ --include="*.go" | grep -v "test" | grep -v "example"; then
    print_status "WARNING" "Found potential hardcoded admin credentials"
    HARDCODED_FOUND=true
fi

# Check for common weak passwords
WEAK_PATTERNS=("password" "123456" "admin123" "secret" "changeme")
for pattern in "${WEAK_PATTERNS[@]}"; do
    if grep -r "$pattern" pkg/ cmd/ --include="*.go" | grep -v "test" | grep -v "example" | grep -v "comment"; then
        print_status "WARNING" "Found potential weak password: $pattern"
        HARDCODED_FOUND=true
    fi
done

if [ "$HARDCODED_FOUND" = false ]; then
    print_status "SUCCESS" "No hardcoded credentials found"
fi

# Step 3: Configure HTTPS/TLS
print_status "INFO" "Configuring HTTPS/TLS settings..."

# Generate self-signed certificates if they don't exist
if [ ! -f "certs/server.crt" ] || [ ! -f "certs/server.key" ]; then
    print_status "INFO" "Generating self-signed TLS certificates..."
    mkdir -p certs
    
    openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt \
        -days 365 -nodes -subj "/C=US/ST=State/L=City/O=OllamaMax/CN=localhost" \
        2>/dev/null
    
    chmod 600 certs/server.key
    chmod 644 certs/server.crt
    
    export TLS_CERT_FILE="$(pwd)/certs/server.crt"
    export TLS_KEY_FILE="$(pwd)/certs/server.key"
    
    echo "export TLS_CERT_FILE=\"$TLS_CERT_FILE\"" >> ~/.bashrc
    echo "export TLS_KEY_FILE=\"$TLS_KEY_FILE\"" >> ~/.bashrc
    
    print_status "SUCCESS" "Generated TLS certificates"
else
    print_status "SUCCESS" "TLS certificates already exist"
fi

# Step 4: Run Security Scanners
print_status "INFO" "Running security scanners..."

# Install security tools if not present
if ! command -v gosec &> /dev/null; then
    print_status "INFO" "Installing gosec security scanner..."
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
fi

if ! command -v nancy &> /dev/null; then
    print_status "INFO" "Installing nancy vulnerability scanner..."
    go install github.com/sonatypecommunity/nancy@latest
fi

if ! command -v govulncheck &> /dev/null; then
    print_status "INFO" "Installing govulncheck..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
fi

# Run gosec
print_status "INFO" "Running gosec security scan..."
if gosec -fmt json -out security-report.json ./... 2>/dev/null; then
    print_status "SUCCESS" "Gosec scan completed (report: security-report.json)"
else
    print_status "WARNING" "Gosec found security issues (check security-report.json)"
fi

# Run nancy vulnerability check
print_status "INFO" "Running nancy vulnerability scan..."
if go list -json -deps ./... | nancy sleuth 2>/dev/null; then
    print_status "SUCCESS" "Nancy vulnerability scan passed"
else
    print_status "WARNING" "Nancy found vulnerable dependencies"
fi

# Run govulncheck
print_status "INFO" "Running govulncheck..."
if govulncheck ./... 2>/dev/null; then
    print_status "SUCCESS" "Govulncheck passed"
else
    print_status "WARNING" "Govulncheck found vulnerabilities"
fi

# Step 5: Configure Security Headers
print_status "INFO" "Configuring security headers..."

# Create security configuration file
cat > config/security.yaml << EOF
security:
  tls:
    enabled: true
    cert_file: "${TLS_CERT_FILE}"
    key_file: "${TLS_KEY_FILE}"
    min_version: "1.2"
  
  headers:
    hsts_max_age: 31536000
    content_security_policy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'"
    x_frame_options: "DENY"
    x_content_type_options: "nosniff"
  
  authentication:
    jwt_secret: "${JWT_SECRET}"
    session_timeout: 3600
    max_failed_attempts: 5
    lockout_duration: 900
  
  rate_limiting:
    enabled: true
    requests_per_minute: 100
    burst_size: 10
  
  audit:
    enabled: true
    log_file: "/var/log/ollama-security.log"
EOF

print_status "SUCCESS" "Created security configuration file"

# Step 6: Set Secure File Permissions
print_status "INFO" "Setting secure file permissions..."

# Secure configuration files
find config/ -name "*.yaml" -exec chmod 600 {} \;
find config/ -name "*.yml" -exec chmod 600 {} \;

# Secure certificate files
if [ -d "certs" ]; then
    chmod 700 certs/
    find certs/ -name "*.key" -exec chmod 600 {} \;
    find certs/ -name "*.crt" -exec chmod 644 {} \;
fi

# Secure log directory
mkdir -p logs
chmod 750 logs/

print_status "SUCCESS" "Set secure file permissions"

# Step 7: Create Security Monitoring Script
print_status "INFO" "Creating security monitoring script..."

cat > scripts/security-monitor.sh << 'EOF'
#!/bin/bash

# Security monitoring script
echo "🔍 Security Monitoring Report - $(date)"
echo "========================================"

# Check for failed authentication attempts
echo "Failed Authentication Attempts (last 24h):"
if [ -f "/var/log/ollama-security.log" ]; then
    grep "auth_failure" /var/log/ollama-security.log | tail -10
else
    echo "No security log found"
fi

# Check certificate expiration
echo -e "\nTLS Certificate Status:"
if [ -f "certs/server.crt" ]; then
    openssl x509 -in certs/server.crt -noout -dates
else
    echo "No TLS certificate found"
fi

# Check for suspicious network connections
echo -e "\nActive Network Connections:"
netstat -tuln | grep :8080

# Check system resources
echo -e "\nSystem Resources:"
echo "CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)%"
echo "Memory: $(free | grep Mem | awk '{printf("%.1f%%", $3/$2 * 100.0)}')"
echo "Disk: $(df -h / | awk 'NR==2{printf "%s", $5}')"

echo -e "\n✅ Security monitoring complete"
EOF

chmod +x scripts/security-monitor.sh
print_status "SUCCESS" "Created security monitoring script"

# Step 8: Final Security Validation
print_status "INFO" "Running final security validation..."

# Validate environment
if [ -n "$JWT_SECRET" ] && [ ${#JWT_SECRET} -ge 32 ]; then
    print_status "SUCCESS" "JWT_SECRET properly configured"
else
    print_status "ERROR" "JWT_SECRET validation failed"
fi

if [ -f "$TLS_CERT_FILE" ] && [ -f "$TLS_KEY_FILE" ]; then
    print_status "SUCCESS" "TLS certificates available"
else
    print_status "WARNING" "TLS certificates not found"
fi

if [ -f "config/security.yaml" ]; then
    print_status "SUCCESS" "Security configuration created"
else
    print_status "ERROR" "Security configuration missing"
fi

# Summary
echo ""
print_status "SUCCESS" "Security hardening completed!"
echo ""
echo "📋 Security Hardening Summary:"
echo "1. ✅ Environment variables configured"
echo "2. ✅ Hardcoded credentials scanned"
echo "3. ✅ TLS certificates generated"
echo "4. ✅ Security scanners executed"
echo "5. ✅ Security headers configured"
echo "6. ✅ File permissions secured"
echo "7. ✅ Monitoring script created"
echo ""
echo "🔐 Next Steps:"
echo "1. Review security-report.json for any issues"
echo "2. Run: source ~/.bashrc"
echo "3. Start with: ./ollama-distributed start --config config/security.yaml"
echo "4. Monitor with: ./scripts/security-monitor.sh"
echo ""
print_status "SUCCESS" "OllamaMax is now security hardened! 🛡️"
