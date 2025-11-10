#!/bin/bash

set -e

echo "========================================"
echo "Building EventFlow Images"
echo "========================================"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Project root: $PROJECT_ROOT"
echo ""

# Build API
echo "Building EventFlow API..."
cd "$PROJECT_ROOT/eventflow/api"
docker build -t eventflow-api:latest .
echo "✅ API image built"
echo ""

# Build Operator
echo "Building EventFlow Operator..."
cd "$PROJECT_ROOT/eventflow/operator"
docker build -t eventflow-operator:latest .
echo "✅ Operator image built"
echo ""

# Build Web UI
echo "Building EventFlow Web UI..."
cd "$PROJECT_ROOT/eventflow/web"
docker build -t eventflow-web:latest .
echo "✅ Web UI image built"
echo ""

# Import images to K3s
echo "Importing images to K3s..."

# K3s uses containerd, so we need to import via docker save/ctr import
echo "Importing eventflow-api..."
docker save eventflow-api:latest | sudo k3s ctr images import -

echo "Importing eventflow-operator..."
docker save eventflow-operator:latest | sudo k3s ctr images import -

echo "Importing eventflow-web..."
docker save eventflow-web:latest | sudo k3s ctr images import -

echo ""
echo "✅ All images built and imported to K3s!"
echo ""
echo "Verify images:"
sudo k3s ctr images ls | grep eventflow
