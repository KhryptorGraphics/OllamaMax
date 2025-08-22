#!/bin/sh

# Health check script for nginx container
# Checks if nginx is running and serving content properly

set -e

# Check if nginx process is running
if ! pgrep nginx > /dev/null; then
    echo "ERROR: nginx process not found"
    exit 1
fi

# Check if the health endpoint responds
if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "ERROR: Health endpoint not responding"
    exit 1
fi

# Check if main application is accessible
if ! curl -f -s http://localhost:8080/ | grep -q "<!DOCTYPE html>"; then
    echo "ERROR: Main application not serving HTML"
    exit 1
fi

# Check nginx configuration is valid
if ! nginx -t > /dev/null 2>&1; then
    echo "ERROR: nginx configuration is invalid"
    exit 1
fi

echo "Health check passed"
exit 0