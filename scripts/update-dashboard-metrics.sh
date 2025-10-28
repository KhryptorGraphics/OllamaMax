#!/bin/bash

# Script to update Grafana dashboard JSON files with correct metric names

DASHBOARD_DIR="/home/kp/OllamaMax/monitoring/grafana/dashboards"

echo "Updating Grafana dashboard metric names..."
echo "==========================================="

# Backup dashboards
mkdir -p "${DASHBOARD_DIR}.backup"
cp -r "${DASHBOARD_DIR}"/*.json "${DASHBOARD_DIR}.backup/"
echo "✓ Backed up dashboards to ${DASHBOARD_DIR}.backup/"

# Update database-performance.json
echo ""
echo "Updating database-performance.json..."

sed -i 's/"db_connections_open"/"ollamamax_database_db_connections_open"/g' \
    "${DASHBOARD_DIR}/database-performance.json"

sed -i 's/"db_connections_in_use"/"ollamamax_database_db_connections_active"/g' \
    "${DASHBOARD_DIR}/database-performance.json"

sed -i 's/"db_connections_idle"/"ollamamax_database_db_connections_idle"/g' \
    "${DASHBOARD_DIR}/database-performance.json"

sed -i 's/"db_queries_total"/"ollamamax_database_db_queries_total"/g' \
    "${DASHBOARD_DIR}/database-performance.json"

sed -i 's/"cache_hits_total"/"ollamamax_database_cache_hits_total"/g' \
    "${DASHBOARD_DIR}/database-performance.json"

sed -i 's/"cache_misses_total"/"ollamamax_database_cache_misses_total"/g' \
    "${DASHBOARD_DIR}/database-performance.json"

sed -i 's/"db_query_duration_seconds"/"ollamamax_database_db_query_duration_seconds"/g' \
    "${DASHBOARD_DIR}/database-performance.json"

echo "✓ Updated database-performance.json"

# Update API performance dashboard if it references DB metrics
echo ""
echo "Updating api-performance.json..."

sed -i 's/"db_connections_open"/"ollamamax_database_db_connections_open"/g' \
    "${DASHBOARD_DIR}/api-performance.json"

sed -i 's/"db_connections_in_use"/"ollamamax_database_db_connections_active"/g' \
    "${DASHBOARD_DIR}/api-performance.json"

sed -i 's/"db_queries_total"/"ollamamax_database_db_queries_total"/g' \
    "${DASHBOARD_DIR}/api-performance.json"

echo "✓ Updated api-performance.json"

# Update overview dashboard
echo ""
echo "Updating ollamamax-overview.json..."

sed -i 's/"db_connections_open"/"ollamamax_database_db_connections_open"/g' \
    "${DASHBOARD_DIR}/ollamamax-overview.json"

sed -i 's/"db_connections_in_use"/"ollamamax_database_db_connections_active"/g' \
    "${DASHBOARD_DIR}/ollamamax-overview.json"

sed -i 's/"db_queries_total"/"ollamamax_database_db_queries_total"/g' \
    "${DASHBOARD_DIR}/ollamamax-overview.json"

sed -i 's/"cache_hits_total"/"ollamamax_database_cache_hits_total"/g' \
    "${DASHBOARD_DIR}/ollamamax-overview.json"

sed -i 's/"cache_misses_total"/"ollamamax_database_cache_misses_total"/g' \
    "${DASHBOARD_DIR}/ollamamax-overview.json"

echo "✓ Updated ollamamax-overview.json"

# Check for any remaining old metric names
echo ""
echo "Checking for remaining old metric names..."
echo "==========================================="

OLD_METRICS=(
    "db_connections_open"
    "db_connections_in_use"
    "db_connections_idle"
    "db_queries_total"
    "cache_hits_total"
    "cache_misses_total"
    "db_query_duration_seconds"
)

FOUND_OLD=0
for metric in "${OLD_METRICS[@]}"; do
    if grep -q "\"${metric}\"" "${DASHBOARD_DIR}"/*.json; then
        echo "⚠ Found remaining '${metric}' in:"
        grep -l "\"${metric}\"" "${DASHBOARD_DIR}"/*.json
        FOUND_OLD=1
    fi
done

if [ $FOUND_OLD -eq 0 ]; then
    echo "✓ No old metric names found - all updated!"
fi

echo ""
echo "Update complete!"
echo "==========================================="
echo "Summary:"
echo "  - Backups saved to: ${DASHBOARD_DIR}.backup/"
echo "  - Updated dashboards in: ${DASHBOARD_DIR}/"
echo ""
echo "Next: Restart Grafana to load updated dashboards"
