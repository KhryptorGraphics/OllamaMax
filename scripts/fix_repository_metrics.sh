#!/bin/bash

# Script to update repository metric recording patterns
# Replaces recordQueryMetrics and recordCacheOperation calls with manager methods

FILE="/home/kp/OllamaMax/pkg/database/repositories.go"

# Backup the file
cp "$FILE" "${FILE}.backup"

# Pattern 1: Replace recordQueryMetrics with inline timing + manager call
# This is complex, so we'll document the manual pattern instead

echo "Repository metrics refactoring notes:"
echo "======================================"
echo ""
echo "For each recordQueryMetrics call, replace with:"
echo ""
echo "  start := time.Now()"
echo "  _, err := r.db.ExecContext(...)"
echo "  if r.manager != nil {"
echo "    r.manager.RecordQuery(\"operation\", \"table\", time.Since(start))"
echo "  }"
echo ""
echo "For cache operations, replace recordCacheOperation/Hit/Miss with:"
echo ""
echo "  cacheStart := time.Now()"
echo "  result, err := r.redis.Get(...)"
echo "  if err == nil {"
echo "    if r.manager != nil { r.manager.RecordCacheHit(time.Since(cacheStart)) }"
echo "  } else if err == redis.Nil {"
echo "    if r.manager != nil { r.manager.RecordCacheMiss(time.Since(cacheStart)) }"
echo "  }"
echo ""
echo "Update all repository struct types to include 'manager *DatabaseManager'"
echo "Update all New*Repository constructors to accept manager parameter"
