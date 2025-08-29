#!/bin/bash
# Make deployment scripts executable

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Making deployment scripts executable..."

# Make all shell scripts executable
find "$SCRIPT_DIR" -name "*.sh" -exec chmod +x {} \;

echo "✅ All scripts are now executable"

# List executable scripts
echo ""
echo "Executable scripts:"
find "$SCRIPT_DIR" -name "*.sh" -executable -exec basename {} \;