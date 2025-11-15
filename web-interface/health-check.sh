#!/bin/sh
# Health check script for OllamaMax web interface

# Check if nginx is running
if ! pgrep nginx > /dev/null; then
    echo "nginx process not found"
    exit 1
fi

# Check if the main page is accessible
if ! curl -f http://localhost/health > /dev/null 2>&1; then
    echo "Health endpoint not accessible"
    exit 1
fi

# Check if index.html exists and is readable
if [ ! -f /usr/share/nginx/html/index.html ]; then
    echo "index.html not found"
    exit 1
fi

# Check if main JS file exists
if [ ! -f /usr/share/nginx/html/main.*.js ] && [ ! -f /usr/share/nginx/html/main.js ]; then
    echo "main JS file not found"
    exit 1
fi

# Check available disk space (warn if less than 10% free)
df /usr/share/nginx/html | awk 'NR==2 {if ($5+0 > 90) exit 1}'

echo "Health check passed"
exit 0