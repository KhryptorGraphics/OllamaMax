#!/bin/bash
# Configuration Generator - Training Tool
# Generates different configuration profiles

PROFILE="${1:-development}"
OUTPUT="${2:-${PROFILE}-config.yaml}"

case $PROFILE in
    "development")
        cat > "$OUTPUT" << DEVEOF
# Development Configuration Profile
api:
  listen: ":8090"
  cors:
    enabled: true
    allowed_origins: ["*"]
p2p:
  listen_addr: "/ip4/127.0.0.1/tcp/4010"
  max_connections: 50
logging:
  level: "debug"
  file: "development.log"
auth:
  enabled: false
DEVEOF
        ;;
    "testing")
        cat > "$OUTPUT" << TESTEOF
# Testing Configuration Profile  
api:
  listen: ":0"  # Random port
p2p:
  listen_addr: "/ip4/127.0.0.1/tcp/0"  # Random port
logging:
  level: "warn"
  file: "/tmp/testing.log"
auth:
  enabled: false
TESTEOF
        ;;
    "production")
        cat > "$OUTPUT" << PRODEOF
# Production Configuration Profile
api:
  listen: ":8080"
  rate_limit:
    enabled: true
    requests_per: 100
  cors:
    enabled: true
    allowed_origins: ["https://your-domain.com"]
p2p:
  listen_addr: "/ip4/0.0.0.0/tcp/4001"
  max_connections: 200
logging:
  level: "info"
  file: "/var/log/ollama-distributed.log"
auth:
  enabled: true
  method: "jwt"
  secret_key: "CHANGE-THIS-SECRET-KEY"
PRODEOF
        ;;
    *)
        echo "Usage: $0 [development|testing|production] [output-file]"
        exit 1
        ;;
esac

echo "Generated $PROFILE configuration: $OUTPUT"
