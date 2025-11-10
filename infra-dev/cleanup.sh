#!/bin/bash

set -e

echo "========================================"
echo "Cleaning up EventFlow from K3s"
echo "========================================"

export KUBECONFIG=$HOME/.kube/config

echo "Deleting EventFlow deployments..."
kubectl delete deployment --all -n eventflow --ignore-not-found=true

echo "Deleting EventFlow services..."
kubectl delete service --all -n eventflow --ignore-not-found=true

echo "Deleting EventFlow Function CRs..."
kubectl delete functions.eventflow.eventflow.io --all --all-namespaces --ignore-not-found=true

echo "Deleting tenant namespaces..."
kubectl delete namespace -l type=tenant --ignore-not-found=true

echo "Deleting EventFlow namespace..."
kubectl delete namespace eventflow --ignore-not-found=true

echo "Deleting CRDs..."
kubectl delete crd functions.eventflow.eventflow.io --ignore-not-found=true

echo "Deleting cluster-level RBAC..."
kubectl delete clusterrole eventflow-api-cluster-role --ignore-not-found=true
kubectl delete clusterrolebinding eventflow-api-cluster-binding --ignore-not-found=true

echo ""
echo "✅ EventFlow cleanup completed!"
echo ""
echo "Note: This does NOT uninstall K3s. To uninstall K3s, run:"
echo "  ./uninstall-k3s.sh"
