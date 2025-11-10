#!/bin/bash

set -e

echo "========================================"
echo "Installing K3s Cluster"
echo "========================================"

# Check if K3s is already installed
if command -v k3s &> /dev/null; then
    echo "K3s is already installed"
    k3s --version
    exit 0
fi

# Install K3s with custom options
echo "Installing K3s..."
curl -sfL https://get.k3s.io | sh -s - \
    --write-kubeconfig-mode 644 \
    --disable traefik \
    --disable servicelb \
    --node-name eventflow-node

# Wait for K3s to be ready
echo "Waiting for K3s to be ready..."
timeout 60 bash -c 'until sudo k3s kubectl get nodes 2>/dev/null; do sleep 2; done'

# Setup kubeconfig for current user
echo "Setting up kubeconfig..."
mkdir -p $HOME/.kube
sudo cp /etc/rancher/k3s/k3s.yaml $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
export KUBECONFIG=$HOME/.kube/config

echo ""
echo "✅ K3s installation completed!"
echo ""
echo "Cluster information:"
kubectl version --short 2>/dev/null || kubectl version
echo ""
kubectl get nodes
echo ""
echo "To use kubectl, run:"
echo "  export KUBECONFIG=\$HOME/.kube/config"
echo ""
echo "Or use k3s kubectl directly:"
echo "  sudo k3s kubectl get nodes"
