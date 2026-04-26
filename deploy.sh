#!/bin/bash
#
# deploy.sh
# Deploys all StudyCafe services to Kubernetes in the correct dependency order.
#
# Usage: ./deploy.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[DEPLOY]${NC} $1"; }
info() { echo -e "${CYAN}[INFO]${NC} $1"; }
wait_for_ready() {
  local label=$1
  local ns=$2
  local timeout=$3
  info "Waiting for $label to be ready (timeout: ${timeout}s)..."
  kubectl wait --for=condition=Available deployment -l "app=$label" \
    -n "$ns" --timeout="${timeout}s" 2>/dev/null || {
    echo -e "${YELLOW}[WARN]${NC} $label not ready within ${timeout}s — continuing anyway"
  }
}

# ──────────────────────────────────────────────────────
# Step 1: Create Namespace
# ──────────────────────────────────────────────────────
log "━━━ Step 1/6: Creating Namespace ━━━"
kubectl apply -f "$SCRIPT_DIR/namespace.yaml"

# ──────────────────────────────────────────────────────
# Step 2: Create Secrets
# ──────────────────────────────────────────────────────
log "━━━ Step 2/6: Creating Secrets ━━━"
kubectl apply -f "$SCRIPT_DIR/secrets.yaml"

# ──────────────────────────────────────────────────────
# Step 3: Deploy Infrastructure (MySQL, Redis, Kafka)
# ──────────────────────────────────────────────────────
log "━━━ Step 3/6: Deploying Infrastructure ━━━"
kubectl apply -f "$SCRIPT_DIR/infra/"

info "Waiting for infrastructure to start..."
wait_for_ready "mysql-test"      "studycafe" 120
wait_for_ready "redis-test"      "studycafe" 60
wait_for_ready "kafka-controller" "studycafe" 60
wait_for_ready "kafka-broker"    "studycafe" 90

# ──────────────────────────────────────────────────────
# Step 4: Deploy Backend Services
# ──────────────────────────────────────────────────────
log "━━━ Step 4/6: Deploying Backend Services ━━━"
kubectl apply -f "$SCRIPT_DIR/apps/study-service-deployment.yaml"
kubectl apply -f "$SCRIPT_DIR/apps/study-service-service.yaml"
kubectl apply -f "$SCRIPT_DIR/apps/auth-service-deployment.yaml"
kubectl apply -f "$SCRIPT_DIR/apps/auth-service-service.yaml"
kubectl apply -f "$SCRIPT_DIR/apps/notification-service-deployment.yaml"
kubectl apply -f "$SCRIPT_DIR/apps/notification-service-service.yaml"

info "Waiting for backend services..."
wait_for_ready "study-service"        "studycafe" 120
wait_for_ready "auth-service"         "studycafe" 120
wait_for_ready "notification-service" "studycafe" 90

# ──────────────────────────────────────────────────────
# Step 5: Deploy API Gateway
# ──────────────────────────────────────────────────────
log "━━━ Step 5/6: Deploying API Gateway ━━━"
kubectl apply -f "$SCRIPT_DIR/apps/api-gateway-deployment.yaml"
kubectl apply -f "$SCRIPT_DIR/apps/api-gateway-service.yaml"

wait_for_ready "api-gateway" "studycafe" 90

# ──────────────────────────────────────────────────────
# Step 6: Deploy React Frontend
# ──────────────────────────────────────────────────────
log "━━━ Step 6/6: Deploying React Frontend ━━━"
kubectl apply -f "$SCRIPT_DIR/apps/react-frontend-deployment.yaml"
kubectl apply -f "$SCRIPT_DIR/apps/react-frontend-service.yaml"

wait_for_ready "react-frontend" "studycafe" 60

# ──────────────────────────────────────────────────────
# Summary
# ──────────────────────────────────────────────────────
echo ""
log "━━━ Deployment Complete! ━━━"
echo ""
info "Checking pod status:"
kubectl get pods -n studycafe -o wide
echo ""
info "Checking services:"
kubectl get svc -n studycafe
echo ""
echo "─────────────────────────────────────────────────"
echo -e "${GREEN}Access your app:${NC}"
echo ""
echo "  If you used kind-config.yaml with extraPortMappings:"
echo "    React Frontend:  http://localhost:3000"
echo "    API Gateway:     https://localhost:8083"
echo ""
echo "  If using the default Kind cluster (no port mappings):"
echo "    kubectl port-forward svc/react-frontend 3000:80 -n studycafe &"
echo "    kubectl port-forward svc/api-gateway 8083:8083 -n studycafe &"
echo "    → React: http://localhost:3000"
echo "    → API:   https://localhost:8083"
echo "─────────────────────────────────────────────────"
