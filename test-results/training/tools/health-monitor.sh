#!/bin/bash
# API Health Monitor - Training Tool
# Monitors the health of Ollama Distributed API

BASE_URL="${1:-http://localhost:8080}"
INTERVAL="${2:-5}"
LOG_FILE="health-monitor.log"

echo "Starting health monitor for $BASE_URL (checking every ${INTERVAL}s)"
echo "Logs will be written to: $LOG_FILE"

while true; do
    TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
    
    if RESPONSE=$(curl -s --connect-timeout 3 "$BASE_URL/health" 2>&1); then
        if echo "$RESPONSE" | jq -e '.status' >/dev/null 2>&1; then
            STATUS=$(echo "$RESPONSE" | jq -r '.status')
            echo "[$TIMESTAMP] SUCCESS - Status: $STATUS" | tee -a "$LOG_FILE"
        else
            echo "[$TIMESTAMP] WARNING - Got response but no status field" | tee -a "$LOG_FILE"
        fi
    else
        echo "[$TIMESTAMP] ERROR - Health check failed: $RESPONSE" | tee -a "$LOG_FILE"
    fi
    
    sleep "$INTERVAL"
done
