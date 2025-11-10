#!/bin/bash

# Monitor EventFlow components for errors
# Usage: ./monitor.sh [component]

set -e

export KUBECONFIG=$HOME/.kube/config

component="${1:-all}"

case "$component" in
    "api")
        echo "Monitoring EventFlow API for errors..."
        kubectl logs -f -n eventflow -l app=eventflow-api --all-containers=true 2>&1 | \
            grep --line-buffered -E "(too many open files|error|Error|ERROR|failed|Failed|FAILED|panic|Panic|PANIC)"
        ;;
    
    "operator")
        echo "Monitoring EventFlow Operator for errors..."
        kubectl logs -f -n eventflow -l app=operator-controller-manager --all-containers=true 2>&1 | \
            grep --line-buffered -E "(error|Error|ERROR|failed|Failed|FAILED|panic|Panic|PANIC)"
        ;;
    
    "all")
        echo "Monitoring all EventFlow components for errors..."
        kubectl logs -f -n eventflow --all-containers=true 2>&1 | \
            grep --line-buffered -E "(too many open files|error|Error|ERROR|failed|Failed|FAILED|panic|Panic|PANIC)"
        ;;
    
    *)
        echo "Usage: $0 [component]"
        echo ""
        echo "Components:"
        echo "  api      - Monitor API only"
        echo "  operator - Monitor Operator only"
        echo "  all      - Monitor all components (default)"
        exit 1
        ;;
esac
