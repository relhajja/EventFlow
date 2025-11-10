#!/bin/bash

# Quick development workflow script
# Usage: ./dev.sh [command]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export KUBECONFIG=$HOME/.kube/config

case "$1" in
    "setup")
        echo "Setting up K3s development environment..."
        "$SCRIPT_DIR/install-k3s.sh"
        "$SCRIPT_DIR/build-images.sh"
        "$SCRIPT_DIR/deploy.sh"
        ;;
    
    "rebuild")
        echo "Rebuilding and redeploying..."
        "$SCRIPT_DIR/build-images.sh"
        kubectl rollout restart deployment -n eventflow
        ;;
    
    "rebuild-api")
        echo "Rebuilding API only..."
        cd "$SCRIPT_DIR/../eventflow/api"
        docker build -t eventflow-api:latest .
        docker save eventflow-api:latest | sudo k3s ctr images import -
        kubectl rollout restart deployment/eventflow-api -n eventflow
        echo "Waiting for rollout..."
        kubectl rollout status deployment/eventflow-api -n eventflow
        ;;
    
    "rebuild-operator")
        echo "Rebuilding Operator only..."
        cd "$SCRIPT_DIR/../eventflow/operator"
        docker build -t eventflow-operator:latest .
        docker save eventflow-operator:latest | sudo k3s ctr images import -
        kubectl rollout restart deployment/operator-controller-manager -n eventflow
        echo "Waiting for rollout..."
        kubectl rollout status deployment/operator-controller-manager -n eventflow
        ;;
    
    "rebuild-web")
        echo "Rebuilding Web UI only..."
        cd "$SCRIPT_DIR/../eventflow/web"
        docker build -t eventflow-web:latest .
        docker save eventflow-web:latest | sudo k3s ctr images import -
        kubectl rollout restart deployment/eventflow-web -n eventflow
        echo "Waiting for rollout..."
        kubectl rollout status deployment/eventflow-web -n eventflow
        ;;
    
    "logs")
        component="${2:-api}"
        case "$component" in
            "api")
                kubectl logs -f -n eventflow -l app=eventflow-api --tail=50
                ;;
            "operator")
                kubectl logs -f -n eventflow -l app=operator-controller-manager --tail=50
                ;;
            "web")
                kubectl logs -f -n eventflow -l app=eventflow-web --tail=50
                ;;
            "postgres")
                kubectl logs -f -n eventflow -l app=postgres --tail=50
                ;;
            *)
                echo "Unknown component: $component"
                echo "Available: api, operator, web, postgres"
                exit 1
                ;;
        esac
        ;;
    
    "status")
        echo "EventFlow Status:"
        echo ""
        kubectl get pods -n eventflow
        echo ""
        echo "Tenant namespaces:"
        kubectl get namespaces -l type=tenant
        echo ""
        echo "Functions:"
        kubectl get functions.eventflow.eventflow.io --all-namespaces
        ;;
    
    "clean")
        "$SCRIPT_DIR/cleanup.sh"
        ;;
    
    "reset")
        echo "Cleaning up and redeploying..."
        "$SCRIPT_DIR/cleanup.sh"
        sleep 5
        "$SCRIPT_DIR/deploy.sh"
        ;;
    
    "shell")
        component="${2:-api}"
        case "$component" in
            "api")
                pod=$(kubectl get pod -n eventflow -l app=eventflow-api -o jsonpath='{.items[0].metadata.name}')
                kubectl exec -it -n eventflow "$pod" -- /bin/sh
                ;;
            "postgres")
                pod=$(kubectl get pod -n eventflow -l app=postgres -o jsonpath='{.items[0].metadata.name}')
                kubectl exec -it -n eventflow "$pod" -- psql -U eventflow -d eventflow
                ;;
            *)
                echo "Unknown component: $component"
                echo "Available: api, postgres"
                exit 1
                ;;
        esac
        ;;
    
    *)
        echo "EventFlow Development Tool"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  setup           - Install K3s, build images, and deploy EventFlow"
        echo "  rebuild         - Rebuild all images and restart deployments"
        echo "  rebuild-api     - Rebuild and redeploy API only"
        echo "  rebuild-operator - Rebuild and redeploy Operator only"
        echo "  rebuild-web     - Rebuild and redeploy Web UI only"
        echo "  logs [component] - Follow logs (api, operator, web, postgres)"
        echo "  status          - Show cluster and EventFlow status"
        echo "  clean           - Remove EventFlow from cluster"
        echo "  reset           - Clean and redeploy everything"
        echo "  shell [component] - Open shell in component (api, postgres)"
        echo ""
        exit 1
        ;;
esac
