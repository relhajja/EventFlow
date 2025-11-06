.PHONY: help build run stop clean test deploy kind-setup

help: ## Show this help
	@echo "EventFlow - Functions as a Service"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build Docker images
	@echo "�� Building Docker images..."
	sudo docker compose build

run: ## Start services with docker-compose
	@echo "🚀 Starting EventFlow..."
	sudo docker compose up -d
	@echo ""
	@echo "✅ EventFlow is running!"
	@echo "   Dashboard: http://localhost:3000"
	@echo "   API:       http://localhost:8080"
	@echo "   Metrics:   http://localhost:9090/metrics"
	@echo ""
	@echo "Get a dev token:"
	@echo "   curl -X POST http://localhost:8080/auth/token"

stop: ## Stop services
	@echo "🛑 Stopping EventFlow..."
	sudo docker compose down

clean: ## Remove containers, images, and volumes
	@echo "🧹 Cleaning up..."
	sudo docker compose down -v --rmi all
	@echo "✅ Cleaned!"

logs: ## View logs
	sudo docker compose logs -f

test-api: ## Run backend tests
	@echo "🧪 Running API tests..."
	cd api && go test ./... -v

test-function: ## Create and test a sample function
	@echo "🧪 Testing function deployment..."
	@./scripts/test-function.sh

kind-setup: ## Create kind cluster and deploy EventFlow
	@echo "☸️  Setting up kind cluster..."
	kind create cluster --name eventflow || true
	@echo "🔨 Building images..."
	cd api && sudo docker build -t eventflow-api:latest .
	cd web && sudo docker build -t eventflow-web:latest .
	@echo "📦 Loading images into kind..."
	kind load docker-image eventflow-api:latest --name eventflow
	kind load docker-image eventflow-web:latest --name eventflow
	@echo "🚀 Deploying to Kubernetes..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/secrets.yaml
	kubectl apply -f k8s/rbac.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/hpa.yaml
	@echo ""
	@echo "✅ EventFlow deployed to kind cluster!"
	@echo "Run: kubectl port-forward -n eventflow svc/eventflow-api 8080:80"

kind-clean: ## Delete kind cluster
	kind delete cluster --name eventflow

deploy: ## Deploy to current Kubernetes context
	@echo "🚀 Deploying to Kubernetes..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/secrets.yaml
	kubectl apply -f k8s/rbac.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/hpa.yaml
	@echo "✅ Deployed!"

status: ## Check deployment status
	kubectl get pods,svc -n eventflow

dev-api: ## Run API locally (requires kubectl access)
	@echo "🔧 Running API in development mode..."
	cd api && go run main.go

dev-web: ## Run frontend dev server
	@echo "🔧 Running web dev server..."
	cd web && npm run dev
