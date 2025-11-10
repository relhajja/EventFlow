#!/bin/bash

set -e

echo "========================================"
echo "Uninstalling K3s Cluster"
echo "========================================"

# Check if K3s is installed
if ! command -v k3s &> /dev/null; then
    echo "K3s is not installed"
    exit 0
fi

echo "This will completely remove K3s from your system."
read -p "Are you sure? (yes/no): " -r
echo

if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Uninstall cancelled"
    exit 0
fi

# Run K3s uninstall script
if [ -f /usr/local/bin/k3s-uninstall.sh ]; then
    echo "Running K3s uninstall script..."
    sudo /usr/local/bin/k3s-uninstall.sh
else
    echo "K3s uninstall script not found"
    exit 1
fi

# Clean up kubeconfig
if [ -f "$HOME/.kube/config" ]; then
    echo "Removing kubeconfig..."
    rm -f "$HOME/.kube/config"
fi

echo ""
echo "✅ K3s uninstalled successfully!"
