#!/bin/bash
#
# build-and-load.sh
# Builds all Docker images locally and loads them into the Kind cluster.
#
# Usage: ./build-and-load.sh [--cluster-name <name>]
#
set -euo pipefail

CLUSTER_NAME="${1:-studycafe}"
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log()  { echo -e "${GREEN}[BUILD]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
err()  { echo -e "${RED}[ERROR]${NC} $1"; }

# ─────────────────────────────────────────────────
# 1) React Frontend
# ─────────────────────────────────────────────────
log "━━━ Building React Frontend ━━━"
cd "$PROJECT_ROOT/StudyCafe_React"
log "Installing npm dependencies..."
npm install
log "Cleaning old build..."
npm run clean || true
log "Building production bundle..."
npm run build
log "Building Docker image: kuuku123/react-apache-app:latest"
docker build -t kuuku123/react-apache-app:latest .

# ─────────────────────────────────────────────────
# 2) StudyCafe Server (Maven)
# ─────────────────────────────────────────────────
log "━━━ Building StudyCafe Server (Maven) ━━━"
cd "$PROJECT_ROOT/StudyCafe_Server_For_React"
log "Running mvn package..."
./mvnw clean package -DskipTests -q 2>/dev/null || mvn clean package -DskipTests -q
log "Building Docker image: kuuku123/study-service:latest"
docker build -t kuuku123/study-service:latest .

# ─────────────────────────────────────────────────
# 3) Auth Service (Gradle)
# ─────────────────────────────────────────────────
log "━━━ Building Auth Service (Gradle) ━━━"
cd "$PROJECT_ROOT/Auth_Service"
log "Running gradle bootJar..."
./gradlew clean bootJar -q
log "Building Docker image: kuuku123/auth-service:latest"
docker build -t kuuku123/auth-service:latest .

# ─────────────────────────────────────────────────
# 4) API Gateway (Gradle)
# ─────────────────────────────────────────────────
log "━━━ Building API Gateway (Gradle) ━━━"
cd "$PROJECT_ROOT/Api_Gateway"
log "Running gradle bootJar..."
./gradlew clean bootJar -q
log "Building Docker image: kuuku123/api-gateway:latest"
docker build -t kuuku123/api-gateway:latest .

# ─────────────────────────────────────────────────
# 5) WebFlux Notification Service (Gradle)
# ─────────────────────────────────────────────────
log "━━━ Building WebFlux Notification Service (Gradle) ━━━"
cd "$PROJECT_ROOT/StudyCafe_WebFlux"
log "Running gradle bootJar..."
./gradlew clean bootJar -q
log "Building Docker image: kuuku123/studycafe-webflux-notification:latest"
docker build -t kuuku123/studycafe-webflux-notification:latest .

# ─────────────────────────────────────────────────
# 6) Load all images into Kind
# ─────────────────────────────────────────────────
log "━━━ Loading images into Kind cluster '${CLUSTER_NAME}' ━━━"

IMAGES=(
  "kuuku123/react-apache-app:latest"
  "kuuku123/study-service:latest"
  "kuuku123/auth-service:latest"
  "kuuku123/api-gateway:latest"
  "kuuku123/studycafe-webflux-notification:latest"
)

for img in "${IMAGES[@]}"; do
  log "Loading $img ..."
  kind load docker-image "$img" --name "$CLUSTER_NAME"
done

log "━━━ All images built and loaded! ━━━"
echo ""
log "Next step: run ./deploy.sh to apply Kubernetes manifests."
