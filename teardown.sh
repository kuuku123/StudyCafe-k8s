#!/bin/bash
#
# teardown.sh
# Removes all StudyCafe resources from the Kind cluster.
#
# Usage: ./teardown.sh
#
set -euo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${RED}[TEARDOWN]${NC} Deleting all StudyCafe resources..."

kubectl delete namespace studycafe --ignore-not-found

echo -e "${GREEN}[TEARDOWN]${NC} Done. All resources in 'studycafe' namespace removed."
echo ""
echo "To also delete the Kind cluster:"
echo "  kind delete cluster --name studycafe"
