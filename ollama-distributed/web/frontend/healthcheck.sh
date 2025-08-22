#!/bin/sh
# Health check script for production deployment

# Check if nginx is running
if ! pgrep nginx > /dev/null; then
    echo "Nginx is not running"
    exit 1
fi

# Check if the application is responding
if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "Health endpoint not responding"
    exit 1
fi

# Check if index.html exists
if [ ! -f /usr/share/nginx/html/index.html ]; then
    echo "Application files not found"
    exit 1
fi

echo "Health check passed"
exit 0