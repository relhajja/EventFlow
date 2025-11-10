.PHONY: help setup install build deploy clean reset rebuild rebuild-api rebuild-operator rebuild-web logs status shell monitor uninstall

# Default target
.DEFAULT_GOAL := help

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Project paths
PROJECT_ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
EVENTFLOW_ROOT := $(PROJECT_ROOT)/..
API_DIR := $(EVENTFLOW_ROOT)/eventflow/api
OPERATOR_DIR := $(EVENTFLOW_ROOT)/eventflow/operator
WEB_DIR := $(EVENTFLOW_ROOT)/eventflow/web
K8S_DIR := $(EVENTFLOW_ROOT)/eventflow/k8s

# Kubeconfig
export KUBECONFIG := $(HOME)/.kube/config

help: ## Show this help message
	@echo "$(BLUE)EventFlow K3s Development$(NC)"
	@echo ""
	@echo "$(GREEN)Usage:$(NC)"
	@echo "  make $(YELLOW)<target>$(NC)"
	@echo ""
	@echo "$(GREEN)Targets:$(NC)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(YELLOW)%-18s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

setup: install build deploy ## Complete setup (install K3s, build images, deploy)

install: ## Install K3s cluster
	@echo "$(GREEN)Installing K3s cluster...$(NC)"
	@if command -v k3s >/dev/null 2>&1; then \
		echo "$(YELLOW)K3s is already installed$(NC)"; \
		k3s --version; \
	else \
		curl -sfL https://get.k3s.io | sh -s - \
			--write-kubeconfig-mode 644 \
			--disable traefik \
			--disable servicelb \
			--node-name eventflow-node; \
		timeout 60 bash -c 'until sudo k3s kubectl get nodes 2>/dev/null; do sleep 2; done'; \
		mkdir -p $(HOME)/.kube; \
		sudo cp /etc/rancher/k3s/k3s.yaml $(HOME)/.kube/config; \
		sudo chown $(shell id -u):$(shell id -g) $(HOME)/.kube/config; \
		echo "$(GREEN)✅ K3s installed successfully!$(NC)"; \
	fi

build: build-api build-operator build-web ## Build all Docker images and import to K3s

build-api: ## Build API image
	@echo "$(GREEN)Building EventFlow API...$(NC)"
	@cd $(API_DIR) && docker build -t eventflow-api:latest .
	@docker save eventflow-api:latest | sudo k3s ctr images import -
	@echo "$(GREEN)✅ API image built and imported$(NC)"

build-operator: ## Build Operator image
	@echo "$(GREEN)Building EventFlow Operator...$(NC)"
	@cd $(OPERATOR_DIR) && docker build -t eventflow-operator:latest .
	@docker save eventflow-operator:latest | sudo k3s ctr images import -
	@echo "$(GREEN)✅ Operator image built and imported$(NC)"

build-web: ## Build Web UI image
	@echo "$(GREEN)Building EventFlow Web UI...$(NC)"
	@cd $(WEB_DIR) && docker build -t eventflow-web:latest .
	@docker save eventflow-web:latest | sudo k3s ctr images import -
	@echo "$(GREEN)✅ Web UI image built and imported$(NC)"

deploy: ## Deploy EventFlow to K3s
	@echo "$(GREEN)Deploying EventFlow to K3s...$(NC)"
	@kubectl create namespace eventflow --dry-run=client -o yaml | kubectl apply -f -
	@kubectl apply -f $(K8S_DIR)/secrets.yaml
	@kubectl apply -f $(K8S_DIR)/postgres.yaml
	@echo "$(YELLOW)Waiting for PostgreSQL...$(NC)"
	@kubectl wait --for=condition=ready pod -l app=postgres -n eventflow --timeout=120s
	@kubectl apply -f $(K8S_DIR)/rbac.yaml
	@kubectl apply -f $(K8S_DIR)/api-rbac.yaml
	@kubectl apply -f $(K8S_DIR)/api-cluster-rbac.yaml
	@kubectl apply -f $(K8S_DIR)/api-tenant-rbac.yaml
	@kubectl apply -f $(K8S_DIR)/operator-rbac.yaml
	@kubectl apply -f $(K8S_DIR)/crd-function.yaml
	@kubectl apply -f $(K8S_DIR)/operator.yaml
	@echo "$(YELLOW)Waiting for operator...$(NC)"
	@kubectl wait --for=condition=ready pod -l app=operator-controller-manager -n eventflow --timeout=120s || true
	@kubectl apply -f $(K8S_DIR)/deployment.yaml
	@echo "$(YELLOW)Waiting for API...$(NC)"
	@kubectl wait --for=condition=ready pod -l app=eventflow-api -n eventflow --timeout=120s
	@echo "$(YELLOW)Waiting for Web UI...$(NC)"
	@kubectl wait --for=condition=ready pod -l app=eventflow-web -n eventflow --timeout=120s
	@echo ""
	@echo "$(GREEN)✅ EventFlow deployed successfully!$(NC)"
	@echo ""
	@echo "$(BLUE)Access the application:$(NC)"
	@echo "  Web UI: $(YELLOW)http://localhost:30080$(NC)"
	@echo "  API: $(YELLOW)http://localhost:30080$(NC)"
	@echo ""
	@echo "$(BLUE)Check status:$(NC)"
	@echo "  make status"

rebuild: build ## Rebuild all images and restart deployments
	@kubectl rollout restart deployment -n eventflow
	@echo "$(GREEN)✅ All deployments restarted$(NC)"

rebuild-api: build-api ## Rebuild and redeploy API only
	@kubectl rollout restart deployment/eventflow-api -n eventflow
	@echo "$(YELLOW)Waiting for rollout...$(NC)"
	@kubectl rollout status deployment/eventflow-api -n eventflow
	@echo "$(GREEN)✅ API redeployed$(NC)"

rebuild-operator: build-operator ## Rebuild and redeploy Operator only
	@kubectl rollout restart deployment/operator-controller-manager -n eventflow
	@echo "$(YELLOW)Waiting for rollout...$(NC)"
	@kubectl rollout status deployment/operator-controller-manager -n eventflow
	@echo "$(GREEN)✅ Operator redeployed$(NC)"

rebuild-web: build-web ## Rebuild and redeploy Web UI only
	@kubectl rollout restart deployment/eventflow-web -n eventflow
	@echo "$(YELLOW)Waiting for rollout...$(NC)"
	@kubectl rollout status deployment/eventflow-web -n eventflow
	@echo "$(GREEN)✅ Web UI redeployed$(NC)"

logs: ## Follow logs (usage: make logs COMPONENT=api)
	@COMPONENT=$${COMPONENT:-api}; \
	case $$COMPONENT in \
		api) kubectl logs -f -n eventflow -l app=eventflow-api --tail=50 ;; \
		operator) kubectl logs -f -n eventflow -l app=operator-controller-manager --tail=50 ;; \
		web) kubectl logs -f -n eventflow -l app=eventflow-web --tail=50 ;; \
		postgres) kubectl logs -f -n eventflow -l app=postgres --tail=50 ;; \
		*) echo "$(RED)Unknown component: $$COMPONENT$(NC)"; \
		   echo "Available: api, operator, web, postgres"; \
		   exit 1 ;; \
	esac

status: ## Show cluster and EventFlow status
	@echo "$(BLUE)EventFlow Status:$(NC)"
	@echo ""
	@kubectl get pods -n eventflow
	@echo ""
	@echo "$(BLUE)Tenant namespaces:$(NC)"
	@kubectl get namespaces -l type=tenant 2>/dev/null || echo "No tenant namespaces yet"
	@echo ""
	@echo "$(BLUE)Functions:$(NC)"
	@kubectl get functions.eventflow.eventflow.io --all-namespaces 2>/dev/null || echo "No functions deployed yet"

shell: ## Open shell in component (usage: make shell COMPONENT=api)
	@COMPONENT=$${COMPONENT:-api}; \
	case $$COMPONENT in \
		api) POD=$$(kubectl get pod -n eventflow -l app=eventflow-api -o jsonpath='{.items[0].metadata.name}'); \
		     kubectl exec -it -n eventflow $$POD -- /bin/sh ;; \
		postgres) POD=$$(kubectl get pod -n eventflow -l app=postgres -o jsonpath='{.items[0].metadata.name}'); \
		          kubectl exec -it -n eventflow $$POD -- psql -U eventflow -d eventflow ;; \
		*) echo "$(RED)Unknown component: $$COMPONENT$(NC)"; \
		   echo "Available: api, postgres"; \
		   exit 1 ;; \
	esac

monitor: ## Monitor logs for errors (usage: make monitor COMPONENT=api)
	@COMPONENT=$${COMPONENT:-all}; \
	case $$COMPONENT in \
		api) kubectl logs -f -n eventflow -l app=eventflow-api --all-containers=true 2>&1 | \
		     grep --line-buffered -E "(too many open files|error|Error|ERROR|failed|Failed|FAILED|panic|Panic|PANIC)" ;; \
		operator) kubectl logs -f -n eventflow -l app=operator-controller-manager --all-containers=true 2>&1 | \
		          grep --line-buffered -E "(error|Error|ERROR|failed|Failed|FAILED|panic|Panic|PANIC)" ;; \
		all) kubectl logs -f -n eventflow --all-containers=true 2>&1 | \
		     grep --line-buffered -E "(too many open files|error|Error|ERROR|failed|Failed|FAILED|panic|Panic|PANIC)" ;; \
		*) echo "$(RED)Unknown component: $$COMPONENT$(NC)"; \
		   echo "Available: api, operator, all"; \
		   exit 1 ;; \
	esac

clean: ## Remove EventFlow from K3s (keeps K3s installed)
	@echo "$(YELLOW)Cleaning up EventFlow...$(NC)"
	@kubectl delete deployment --all -n eventflow --ignore-not-found=true
	@kubectl delete service --all -n eventflow --ignore-not-found=true
	@kubectl delete functions.eventflow.eventflow.io --all --all-namespaces --ignore-not-found=true
	@kubectl delete namespace -l type=tenant --ignore-not-found=true
	@kubectl delete namespace eventflow --ignore-not-found=true
	@kubectl delete crd functions.eventflow.eventflow.io --ignore-not-found=true
	@kubectl delete clusterrole eventflow-api-cluster-role --ignore-not-found=true
	@kubectl delete clusterrolebinding eventflow-api-cluster-binding --ignore-not-found=true
	@echo "$(GREEN)✅ EventFlow cleaned up$(NC)"

reset: clean deploy ## Clean and redeploy EventFlow

uninstall: ## Completely uninstall K3s
	@echo "$(RED)This will completely remove K3s from your system.$(NC)"
	@read -p "Are you sure? [yes/no]: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		if [ -f /usr/local/bin/k3s-uninstall.sh ]; then \
			sudo /usr/local/bin/k3s-uninstall.sh; \
			rm -f $(HOME)/.kube/config; \
			echo "$(GREEN)✅ K3s uninstalled$(NC)"; \
		else \
			echo "$(RED)K3s uninstall script not found$(NC)"; \
			exit 1; \
		fi \
	else \
		echo "$(YELLOW)Uninstall cancelled$(NC)"; \
	fi

# Convenience targets for common component operations
.PHONY: api operator web postgres
api: logs ## Alias for 'make logs COMPONENT=api'
	@COMPONENT=api $(MAKE) logs

operator: ## Alias for 'make logs COMPONENT=operator'
	@COMPONENT=operator $(MAKE) logs

web: ## Alias for 'make logs COMPONENT=web'
	@COMPONENT=web $(MAKE) logs

postgres: ## Alias for 'make logs COMPONENT=postgres'
	@COMPONENT=postgres $(MAKE) logs
