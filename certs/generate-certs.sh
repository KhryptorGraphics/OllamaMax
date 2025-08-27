#!/bin/bash

# Certificate Generation Script for OllamaMax Cluster
# Creates self-signed certificates for secure communication

set -e

CERT_DIR="/home/kp/ollamamax/certs"
DOMAIN="ollamamax.local"
COUNTRY="US"
STATE="CA"
CITY="San Francisco"
ORG="OllamaMax"
OU="Distributed Systems"

echo "Creating certificate directory..."
mkdir -p "$CERT_DIR"
cd "$CERT_DIR"

echo "Generating private key..."
openssl genrsa -out server.key 4096

echo "Generating certificate signing request..."
openssl req -new -key server.key -out server.csr -subj "/C=${COUNTRY}/ST=${STATE}/L=${CITY}/O=${ORG}/OU=${OU}/CN=${DOMAIN}"

echo "Generating self-signed certificate..."
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt \
  -extensions v3_req -extfile <(
cat <<EOF
[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = ollamamax.local
DNS.3 = node-1
DNS.4 = node-2
DNS.5 = node-3
DNS.6 = ollamamax-node-1
DNS.7 = ollamamax-node-2
DNS.8 = ollamamax-node-3
IP.1 = 127.0.0.1
IP.2 = 172.20.0.2
IP.3 = 172.20.0.3
IP.4 = 172.20.0.4
EOF
)

echo "Setting appropriate permissions..."
chmod 600 server.key
chmod 644 server.crt

echo "Cleaning up temporary files..."
rm -f server.csr

echo "Certificate generation completed successfully!"
echo "Certificate: $CERT_DIR/server.crt"
echo "Private Key: $CERT_DIR/server.key"

# Verify certificate
echo "Certificate details:"
openssl x509 -in server.crt -text -noout | grep -A 1 "Subject:" | head -2