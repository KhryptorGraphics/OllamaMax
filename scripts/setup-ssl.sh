#!/bin/bash

# SSL/TLS Certificate Setup Script
# Generates self-signed certificates for development and provides instructions for production

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Check if running from project root
if [ ! -f "package.json" ]; then
    log_error "Please run this script from the project root directory"
    exit 1
fi

log_header "SSL/TLS Certificate Setup"

# Create certs directory
CERT_DIR="./certs"
mkdir -p "$CERT_DIR"
log_success "Created certificates directory: $CERT_DIR"

# Determine environment
ENV=${NODE_ENV:-development}
log_info "Environment: $ENV"

if [ "$ENV" = "production" ]; then
    log_header "Production SSL Setup"
    
    echo ""
    log_warning "For production, you should use certificates from a trusted Certificate Authority (CA)"
    echo ""
    echo "Recommended options:"
    echo ""
    echo "1. Let's Encrypt (Free, Automated)"
    echo "   - Install certbot: https://certbot.eff.org/"
    echo "   - Run: sudo certbot certonly --standalone -d yourdomain.com"
    echo "   - Certificates will be in: /etc/letsencrypt/live/yourdomain.com/"
    echo ""
    echo "2. Commercial CA (DigiCert, GlobalSign, etc.)"
    echo "   - Purchase certificate from CA"
    echo "   - Follow CA's verification process"
    echo "   - Download and install certificates"
    echo ""
    echo "3. Cloud Provider (AWS ACM, Google Cloud, Azure)"
    echo "   - Use cloud provider's certificate management"
    echo "   - Integrate with load balancer"
    echo ""
    
    read -p "Do you want to generate self-signed certificates for testing? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Skipping certificate generation"
        exit 0
    fi
fi

log_header "Generating Self-Signed Certificates"

# Get domain name
read -p "Enter domain name (default: localhost): " DOMAIN
DOMAIN=${DOMAIN:-localhost}

log_info "Generating certificates for: $DOMAIN"

# Generate private key
log_info "Generating private key..."
openssl genrsa -out "$CERT_DIR/server.key" 2048
log_success "Private key generated: $CERT_DIR/server.key"

# Generate certificate signing request
log_info "Generating certificate signing request..."
openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" \
    -subj "/C=US/ST=State/L=City/O=Organization/OU=IT/CN=$DOMAIN"
log_success "CSR generated: $CERT_DIR/server.csr"

# Generate self-signed certificate (valid for 365 days)
log_info "Generating self-signed certificate..."
openssl x509 -req -days 365 -in "$CERT_DIR/server.csr" \
    -signkey "$CERT_DIR/server.key" -out "$CERT_DIR/server.crt"
log_success "Certificate generated: $CERT_DIR/server.crt"

# Set proper permissions
chmod 600 "$CERT_DIR/server.key"
chmod 644 "$CERT_DIR/server.crt"
log_success "Set proper file permissions"

# Generate DH parameters for stronger security (optional, takes time)
read -p "Generate Diffie-Hellman parameters? (recommended but slow) (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Generating DH parameters (this may take several minutes)..."
    openssl dhparam -out "$CERT_DIR/dhparam.pem" 2048
    log_success "DH parameters generated: $CERT_DIR/dhparam.pem"
fi

# Display certificate information
log_header "Certificate Information"
openssl x509 -in "$CERT_DIR/server.crt" -text -noout | grep -A 2 "Subject:"
openssl x509 -in "$CERT_DIR/server.crt" -text -noout | grep -A 2 "Validity"

log_header "Next Steps"

echo ""
echo "1. Update your .env file:"
echo "   SSL_ENABLED=true"
echo "   SSL_CERT_PATH=$CERT_DIR/server.crt"
echo "   SSL_KEY_PATH=$CERT_DIR/server.key"
if [ -f "$CERT_DIR/dhparam.pem" ]; then
    echo "   SSL_DH_PARAM_PATH=$CERT_DIR/dhparam.pem"
fi
echo ""
echo "2. Restart the server:"
echo "   npm start"
echo ""
echo "3. Access via HTTPS:"
echo "   https://$DOMAIN:13000"
echo ""

if [ "$ENV" != "production" ]; then
    log_warning "Self-signed certificates will show security warnings in browsers"
    echo ""
    echo "To trust the certificate:"
    echo "  - Chrome/Edge: Click 'Advanced' → 'Proceed to $DOMAIN'"
    echo "  - Firefox: Click 'Advanced' → 'Accept the Risk and Continue'"
    echo "  - System-wide: Add $CERT_DIR/server.crt to your system's trusted certificates"
    echo ""
fi

log_success "SSL setup complete!"

