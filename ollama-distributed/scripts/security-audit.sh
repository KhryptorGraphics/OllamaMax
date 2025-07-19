#!/bin/bash
# Security Audit Script for Ollama Distributed
# Validates security configurations and identifies vulnerabilities

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$PROJECT_ROOT/config"
DEPLOY_DIR="$PROJECT_ROOT/deploy"
REPORT_FILE="$PROJECT_ROOT/security-audit-report.txt"

# Counters
PASSED=0
FAILED=0
WARNINGS=0

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    ((WARNINGS++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

check_file_exists() {
    local file="$1"
    local description="$2"
    
    if [[ -f "$file" ]]; then
        log_success "$description exists: $file"
        return 0
    else
        log_error "$description missing: $file"
        return 1
    fi
}

check_file_permissions() {
    local file="$1"
    local expected_perms="$2"
    local description="$3"
    
    if [[ -f "$file" ]]; then
        local actual_perms
        actual_perms=$(stat -c "%a" "$file" 2>/dev/null || stat -f "%Lp" "$file" 2>/dev/null)
        if [[ "$actual_perms" == "$expected_perms" ]]; then
            log_success "$description has correct permissions ($expected_perms): $file"
        else
            log_warning "$description has incorrect permissions ($actual_perms, expected $expected_perms): $file"
        fi
    else
        log_error "$description not found: $file"
    fi
}

check_config_security() {
    local config_file="$1"
    local description="$2"
    
    if [[ ! -f "$config_file" ]]; then
        log_error "$description not found: $config_file"
        return 1
    fi
    
    log_info "Checking $description security..."
    
    # Check for hardcoded secrets
    if grep -q "CHANGE_THIS" "$config_file"; then
        log_error "$description contains default/placeholder secrets"
    else
        log_success "$description has no obvious placeholder secrets"
    fi
    
    # Check for wildcard CORS origins
    if grep -q "allowed_origins.*\*" "$config_file"; then
        log_error "$description has wildcard CORS origins (*)"
    else
        log_success "$description has restricted CORS origins"
    fi
    
    # Check TLS configuration
    if grep -q "enabled.*true" "$config_file" && grep -q "min_version.*1\.3" "$config_file"; then
        log_success "$description has TLS 1.3 enabled"
    else
        log_error "$description has weak or disabled TLS configuration"
    fi
    
    # Check for environment variable usage
    if grep -q "\${" "$config_file"; then
        log_success "$description uses environment variables for sensitive data"
    else
        log_warning "$description may have hardcoded sensitive values"
    fi
}

check_docker_security() {
    log_info "Checking Docker configuration security..."
    
    local docker_compose="$DEPLOY_DIR/docker/docker-compose.yml"
    
    if [[ -f "$docker_compose" ]]; then
        # Check for hardcoded passwords
        if grep -q "password.*admin" "$docker_compose" || grep -q "password.*password" "$docker_compose"; then
            log_error "Docker Compose has default/weak passwords"
        else
            log_success "Docker Compose passwords appear to be secure"
        fi
        
        # Check for environment variable usage
        if grep -q "\${" "$docker_compose"; then
            log_success "Docker Compose uses environment variables"
        else
            log_warning "Docker Compose may have hardcoded values"
        fi
        
        # Check for privileged containers
        if grep -q "privileged.*true" "$docker_compose"; then
            log_error "Docker Compose has privileged containers"
        else
            log_success "Docker Compose has no privileged containers"
        fi
        
        # Check for host network mode
        if grep -q "network_mode.*host" "$docker_compose"; then
            log_warning "Docker Compose uses host networking"
        else
            log_success "Docker Compose uses isolated networking"
        fi
    else
        log_error "Docker Compose file not found"
    fi
}

check_kubernetes_security() {
    log_info "Checking Kubernetes configuration security..."
    
    local k8s_values="$DEPLOY_DIR/kubernetes/helm/ollamacron/values.yaml"
    local k8s_secrets="$DEPLOY_DIR/kubernetes/secrets-template.yaml"
    
    if [[ -f "$k8s_values" ]]; then
        # Check security context
        if grep -q "runAsNonRoot.*true" "$k8s_values"; then
            log_success "Kubernetes runs as non-root user"
        else
            log_error "Kubernetes may run as root user"
        fi
        
        # Check for privilege escalation
        if grep -q "allowPrivilegeEscalation.*false" "$k8s_values"; then
            log_success "Kubernetes prevents privilege escalation"
        else
            log_error "Kubernetes allows privilege escalation"
        fi
        
        # Check read-only root filesystem
        if grep -q "readOnlyRootFilesystem.*true" "$k8s_values"; then
            log_success "Kubernetes uses read-only root filesystem"
        else
            log_warning "Kubernetes root filesystem is writable"
        fi
        
        # Check network policy
        if grep -q "networkPolicy.*enabled.*true" "$k8s_values"; then
            log_success "Kubernetes network policy is enabled"
        else
            log_warning "Kubernetes network policy is disabled"
        fi
        
        # Check resource limits
        if grep -q "limits:" "$k8s_values" && grep -q "requests:" "$k8s_values"; then
            log_success "Kubernetes has resource limits and requests"
        else
            log_warning "Kubernetes missing resource constraints"
        fi
    else
        log_error "Kubernetes values file not found"
    fi
    
    if [[ -f "$k8s_secrets" ]]; then
        # Check for base64 encoded secrets
        if grep -q "Q0hBTkdF" "$k8s_secrets"; then
            log_error "Kubernetes secrets contain placeholder values"
        else
            log_success "Kubernetes secrets appear to be properly configured"
        fi
    else
        log_error "Kubernetes secrets template not found"
    fi
}

check_tls_configuration() {
    log_info "Checking TLS configuration..."
    
    local tls_config="$CONFIG_DIR/tls-config.yaml"
    
    if [[ -f "$tls_config" ]]; then
        # Check minimum TLS version
        if grep -q "min_version.*1\.3" "$tls_config"; then
            log_success "TLS minimum version is 1.3"
        else
            log_error "TLS minimum version is not 1.3"
        fi
        
        # Check cipher suites
        if grep -q "TLS_AES_256_GCM_SHA384" "$tls_config"; then
            log_success "TLS uses strong cipher suites"
        else
            log_warning "TLS cipher suites may be weak"
        fi
        
        # Check certificate validation
        if grep -q "verify_hostname.*true" "$tls_config"; then
            log_success "TLS hostname verification is enabled"
        else
            log_error "TLS hostname verification is disabled"
        fi
    else
        log_error "TLS configuration file not found"
    fi
}

check_backup_security() {
    log_info "Checking backup configuration security..."
    
    local backup_config="$CONFIG_DIR/backup-restore.yaml"
    
    if [[ -f "$backup_config" ]]; then
        # Check encryption
        if grep -q "encryption.*enabled.*true" "$backup_config"; then
            log_success "Backup encryption is enabled"
        else
            log_error "Backup encryption is disabled"
        fi
        
        # Check compression
        if grep -q "compression.*true" "$backup_config"; then
            log_success "Backup compression is enabled"
        else
            log_warning "Backup compression is disabled"
        fi
        
        # Check retention policy
        if grep -q "retention_policy" "$backup_config"; then
            log_success "Backup retention policy is configured"
        else
            log_warning "Backup retention policy is not configured"
        fi
    else
        log_error "Backup configuration file not found"
    fi
}

check_environment_files() {
    log_info "Checking environment files..."
    
    local env_example="$PROJECT_ROOT/.env.example"
    local docker_env="$DEPLOY_DIR/docker/.env.docker"
    
    # Check .env.example
    if [[ -f "$env_example" ]]; then
        if grep -q "CHANGE_THIS" "$env_example"; then
            log_success ".env.example contains placeholder values (correct)"
        else
            log_warning ".env.example may contain actual secrets"
        fi
    else
        log_error ".env.example file not found"
    fi
    
    # Check for actual .env files with secrets
    if [[ -f "$PROJECT_ROOT/.env" ]]; then
        log_warning "Found .env file in project root - ensure it's not committed to git"
        check_file_permissions "$PROJECT_ROOT/.env" "600" ".env file"
    fi
    
    # Check Docker .env file
    if [[ -f "$docker_env" ]]; then
        if grep -q "CHANGE_THIS" "$docker_env"; then
            log_success "Docker .env file contains placeholder values (correct)"
        else
            log_warning "Docker .env file may contain actual secrets"
        fi
    fi
}

check_port_conflicts() {
    log_info "Checking for port conflicts..."
    
    local docker_compose="$DEPLOY_DIR/docker/docker-compose.yml"
    
    if [[ -f "$docker_compose" ]]; then
        # Extract external ports from Docker Compose
        local ports
        ports=$(grep -E "^\s*-\s*\"[0-9]+:" "$docker_compose" | sed -E 's/.*"([0-9]+):.*/\1/' | sort -n)
        
        # Check for duplicates
        local duplicates
        duplicates=$(echo "$ports" | uniq -d)
        
        if [[ -n "$duplicates" ]]; then
            log_error "Port conflicts found: $duplicates"
        else
            log_success "No port conflicts found in Docker Compose"
        fi
        
        # Check for standard port usage
        if echo "$ports" | grep -q "^80$\|^443$\|^22$\|^21$"; then
            log_warning "Using standard system ports (80, 443, 22, 21)"
        else
            log_success "No conflicts with standard system ports"
        fi
    fi
}

check_file_structure() {
    log_info "Checking security-related file structure..."
    
    # Check for required security files
    check_file_exists "$CONFIG_DIR/security-headers.yaml" "Security headers configuration"
    check_file_exists "$CONFIG_DIR/tls-config.yaml" "TLS configuration"
    check_file_exists "$CONFIG_DIR/backup-restore.yaml" "Backup configuration"
    check_file_exists "$PROJECT_ROOT/.env.example" "Environment template"
    check_file_exists "$DEPLOY_DIR/kubernetes/secrets-template.yaml" "Kubernetes secrets template"
    
    # Check permissions on sensitive files
    find "$CONFIG_DIR" -name "*.yaml" -type f | while read -r file; do
        check_file_permissions "$file" "644" "Configuration file"
    done
    
    # Check for executable scripts
    if [[ -f "$SCRIPT_DIR/security-audit.sh" ]]; then
        check_file_permissions "$SCRIPT_DIR/security-audit.sh" "755" "Security audit script"
    fi
}

generate_report() {
    log_info "Generating security audit report..."
    
    {
        echo "# Ollama Distributed Security Audit Report"
        echo "Generated on: $(date)"
        echo "Project root: $PROJECT_ROOT"
        echo ""
        echo "## Summary"
        echo "- Passed: $PASSED"
        echo "- Failed: $FAILED"
        echo "- Warnings: $WARNINGS"
        echo "- Total checks: $((PASSED + FAILED + WARNINGS))"
        echo ""
        
        if [[ $FAILED -eq 0 ]]; then
            echo "✅ **Security Status: GOOD**"
        elif [[ $FAILED -le 3 ]]; then
            echo "⚠️ **Security Status: NEEDS ATTENTION**"
        else
            echo "❌ **Security Status: CRITICAL ISSUES**"
        fi
        
        echo ""
        echo "## Recommendations"
        
        if [[ $FAILED -gt 0 ]]; then
            echo "### Critical Issues (Fix Immediately)"
            echo "- Review failed security checks above"
            echo "- Replace all placeholder secrets with secure values"
            echo "- Enable TLS 1.3 with strong cipher suites"
            echo "- Implement proper CORS policies"
            echo "- Enable backup encryption"
        fi
        
        if [[ $WARNINGS -gt 0 ]]; then
            echo "### Warnings (Recommended Fixes)"
            echo "- Review warning items above"
            echo "- Implement network policies in Kubernetes"
            echo "- Set proper file permissions"
            echo "- Enable additional security features"
        fi
        
        echo ""
        echo "## Next Steps"
        echo "1. Fix all critical security issues"
        echo "2. Generate and configure secure secrets"
        echo "3. Implement TLS certificates"
        echo "4. Test security configurations"
        echo "5. Set up monitoring and alerting"
        echo "6. Regular security audits"
        
    } > "$REPORT_FILE"
    
    log_success "Security audit report saved to: $REPORT_FILE"
}

main() {
    echo "🔒 Ollama Distributed Security Audit"
    echo "=================================="
    echo ""
    
    # Run all security checks
    check_file_structure
    check_config_security "$CONFIG_DIR/config.yaml" "Main configuration"
    check_config_security "$CONFIG_DIR/node.yaml" "Node configuration"
    check_config_security "$DEPLOY_DIR/config/environments/production.yaml" "Production configuration"
    check_tls_configuration
    check_backup_security
    check_docker_security
    check_kubernetes_security
    check_environment_files
    check_port_conflicts
    
    echo ""
    echo "🔒 Security Audit Complete"
    echo "========================="
    echo -e "Passed: ${GREEN}$PASSED${NC}"
    echo -e "Failed: ${RED}$FAILED${NC}"
    echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
    echo "Total: $((PASSED + FAILED + WARNINGS))"
    echo ""
    
    # Generate report
    generate_report
    
    # Exit with appropriate code
    if [[ $FAILED -eq 0 ]]; then
        echo -e "${GREEN}✅ Security audit passed!${NC}"
        exit 0
    elif [[ $FAILED -le 3 ]]; then
        echo -e "${YELLOW}⚠️  Security audit has minor issues${NC}"
        exit 1
    else
        echo -e "${RED}❌ Security audit failed with critical issues${NC}"
        exit 2
    fi
}

# Run main function
main "$@"