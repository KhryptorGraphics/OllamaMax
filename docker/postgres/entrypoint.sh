#!/bin/bash
set -e

# Set the port in postgresql.conf
if [ "$POSTGRES_PORT" ]; then
    echo "port = $POSTGRES_PORT" >> "$PGDATA/postgresql.conf"
fi

# Call the original entrypoint
exec docker-entrypoint.sh "$@"