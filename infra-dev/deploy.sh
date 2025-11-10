#!/bin/bash

set -e

echo "========================================"
echo "Deploying EventFlow to K3s"
echo "========================================"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
K8S_DIR="$PROJECT_ROOT/eventflow/k8s"

export KUBECONFIG=$HOME/.kube/config

echo "Using kubeconfig: $KUBECONFIG"
echo ""

# Create eventflow namespace
echo "Creating eventflow namespace..."
kubectl create namespace eventflow --dry-run=client -o yaml | kubectl apply -f -

# Apply secrets
echo "Applying secrets..."
kubectl apply -f "$K8S_DIR/secrets.yaml"

# Deploy PostgreSQL
echo "Deploying PostgreSQL..."
kubectl apply -f "$K8S_DIR/postgres.yaml"

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n eventflow --timeout=120s

# Apply RBAC configurations
echo "Applying RBAC configurations..."
kubectl apply -f "$K8S_DIR/rbac.yaml"
kubectl apply -f "$K8S_DIR/api-rbac.yaml"
kubectl apply -f "$K8S_DIR/api-cluster-rbac.yaml"
kubectl apply -f "$K8S_DIR/api-tenant-rbac.yaml"
kubectl apply -f "$K8S_DIR/operator-rbac.yaml"

# Deploy Function CRD
echo "Deploying Function CRD..."
kubectl apply -f "$K8S_DIR/crd-function.yaml"

# Deploy Operator
echo "Deploying EventFlow Operator..."
kubectl apply -f "$K8S_DIR/operator.yaml"

# Wait for operator to be ready
echo "Waiting for operator to be ready..."
kubectl wait --for=condition=ready pod -l app=operator-controller-manager -n eventflow --timeout=120s || true

# Deploy API and Web
echo "Deploying EventFlow API and Web..."
kubectl apply -f "$K8S_DIR/deployment.yaml"

# Wait for deployments to be ready
echo "Waiting for API to be ready..."
kubectl wait --for=condition=ready pod -l app=eventflow-api -n eventflow --timeout=120s

echo "Waiting for Web UI to be ready..."
kubectl wait --for=condition=ready pod -l app=eventflow-web -n eventflow --timeout=120s

echo ""
echo "✅ EventFlow deployed successfully!"
echo ""
echo "Access the application:"
echo "  Web UI: http://localhost:30080"
echo "  API: http://localhost:30080"
echo ""
echo "Check deployment status:"
echo "  kubectl get pods -n eventflow"
echo ""
echo "View logs:"
echo "  kubectl logs -n eventflow -l app=eventflow-api"
echo "  kubectl logs -n eventflow -l app=eventflow-web"
echo "  kubectl logs -n eventflow -l app=operator-controller-manager"
