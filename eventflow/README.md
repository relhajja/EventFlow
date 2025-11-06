# EventFlow - Kubernetes-Native Functions as a Service

<div align="center">
  
**A scalable, Kubernetes-native "Functions-as-a-Service" platform**

Built with Go backend + React TypeScript frontend

[Features](#features) • [Quick Start](#quick-start) • [Architecture](#architecture) • [API](#api-documentation) • [Development](#development)

</div>

---

## 🚀 Features

- **Kubernetes-Native**: Deploy functions as Kubernetes Deployments + Services
- **Multi-Tenancy**: Namespace isolation with JWT authentication
- **Real-time Monitoring**: Live function status, logs, and metrics
- **Scalable**: HPA support for auto-scaling based on CPU/memory
- **Modern Stack**: Go (chi router) + React (TypeScript, Tailwind, React Query)
- **Production Ready**: Docker images, Kubernetes manifests, Helm charts

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Configuration](#configuration)

## ⚡ Quick Start

### Prerequisites

- Docker & Docker Compose
- Kubernetes cluster (kind/minikube for local development)
- kubectl configured
- Go 1.22+ (for local development)
- Node.js 18+ (for frontend development)

### Local Development with Docker Compose

```bash
# Clone the repository
cd eventflow

# Start all services
docker-compose up -d

# Access the dashboard
open http://localhost:3000

# Get a dev token
curl -X POST http://localhost:8080/auth/token

# Use the token in the dashboard login
```

### Local Development with kind

```bash
# Create a kind cluster
kind create cluster --name eventflow

# Build Docker images
cd api && docker build -t eventflow-api:latest .
cd ../web && docker build -t eventflow-web:latest .

# Load images into kind
kind load docker-image eventflow-api:latest --name eventflow
kind load docker-image eventflow-web:latest --name eventflow

# Deploy to Kubernetes
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/rbac.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/hpa.yaml

# Port forward to access
kubectl port-forward -n eventflow svc/eventflow-api 8080:80
```

## 🏗️ Architecture

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │
       ▼
┌─────────────────┐      ┌──────────────────┐
│  React Frontend │─────▶│   Go Backend API │
│  (TypeScript)   │      │   (chi router)   │
└─────────────────┘      └────────┬─────────┘
                                  │
                         ┌────────┴─────────┐
                         │  Kubernetes API  │
                         └────────┬─────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        ▼                         ▼                         ▼
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│ Deployment 1 │         │ Deployment 2 │         │ Deployment N │
│  (Function)  │         │  (Function)  │         │  (Function)  │
└──────────────┘         └──────────────┘         └──────────────┘
```

### Components

**Backend (Go)**
- Framework: `chi` (HTTP router)
- Kubernetes: `client-go` (in-cluster config)
- Auth: JWT with OIDC support
- Metrics: Prometheus `/metrics` endpoint
- Health: `/healthz`, `/readyz` endpoints

**Frontend (React + TypeScript)**
- Build: Vite
- State: React Query (data fetching/caching)
- UI: Tailwind CSS
- Forms: React Hook Form
- Charts: Recharts
- Icons: Lucide React

## 📡 API Documentation

### Authentication

All API endpoints require JWT authentication:

```bash
# Get dev token
curl -X POST http://localhost:8080/auth/token

# Use in requests
curl -H "Authorization: Bearer <token>" http://localhost:8080/v1/functions
```

### Endpoints

#### **Create Function**
```http
POST /v1/functions
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "my-function",
  "image": "nginx:latest",
  "replicas": 2,
  "command": ["nginx", "-g", "daemon off;"],
  "env": {
    "KEY": "value"
  }
}
```

#### **List Functions**
```http
GET /v1/functions
Authorization: Bearer <token>
```

Response:
```json
[
  {
    "name": "my-function",
    "image": "nginx:latest",
    "replicas": 2,
    "available_replicas": 2,
    "ready_replicas": 2,
    "updated_replicas": 2,
    "status": "Running",
    "created_at": "2025-01-01T00:00:00Z"
  }
]
```

#### **Get Function Details**
```http
GET /v1/functions/{name}
Authorization: Bearer <token>
```

#### **Invoke Function**
```http
POST /v1/functions/{name}:invoke
Authorization: Bearer <token>
Content-Type: application/json

{
  "payload": {}
}
```

#### **Delete Function**
```http
DELETE /v1/functions/{name}
Authorization: Bearer <token>
```

#### **Get Function Logs**
```http
GET /v1/functions/{name}/logs?follow=true
Authorization: Bearer <token>
```

#### **Health & Metrics**
```http
GET /healthz      # Health check
GET /readyz       # Readiness check
GET /metrics      # Prometheus metrics
```

## 💻 Development

### Backend Development

```bash
cd api

# Install dependencies
go mod download

# Run locally (requires kubectl access)
export NAMESPACE=default
export JWT_SECRET=dev-secret
go run main.go

# Run tests
go test ./...

# Build
go build -o eventflow-api .
```

### Frontend Development

```bash
cd web

# Install dependencies
npm install

# Run dev server (with API proxy)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Project Structure

```
eventflow/
├── api/                      # Go backend
│   ├── main.go              # Entry point
│   ├── internal/
│   │   ├── auth/            # JWT authentication
│   │   ├── config/          # Configuration
│   │   ├── handlers/        # HTTP handlers
│   │   ├── k8s/             # Kubernetes client
│   │   ├── metrics/         # Prometheus metrics
│   │   ├── models/          # Data models
│   │   └── server/          # HTTP server
│   ├── Dockerfile
│   └── go.mod
├── web/                     # React frontend
│   ├── src/
│   │   ├── components/      # React components
│   │   ├── pages/           # Page components
│   │   ├── services/        # API client
│   │   ├── context/         # React context
│   │   ├── types/           # TypeScript types
│   │   └── main.tsx         # Entry point
│   ├── Dockerfile
│   ├── nginx.conf
│   ├── package.json
│   └── vite.config.ts
├── k8s/                     # Kubernetes manifests
│   ├── namespace.yaml
│   ├── secrets.yaml
│   ├── rbac.yaml
│   ├── deployment.yaml
│   └── hpa.yaml
├── helm/                    # Helm chart (stretch goal)
├── docker-compose.yaml
└── README.md
```

## ☸️ Kubernetes Deployment

### Deploy to Kubernetes

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Create secrets (modify first!)
kubectl apply -f k8s/secrets.yaml

# Create RBAC
kubectl apply -f k8s/rbac.yaml

# Deploy API
kubectl apply -f k8s/deployment.yaml

# Enable HPA
kubectl apply -f k8s/hpa.yaml

# Check status
kubectl get pods -n eventflow
kubectl get svc -n eventflow

# View logs
kubectl logs -n eventflow -l app=eventflow-api -f
```

### Access the Dashboard

```bash
# Port forward
kubectl port-forward -n eventflow svc/eventflow-api 8080:80

# Or use ingress (configure as needed)
```

### RBAC Permissions

The API requires the following Kubernetes permissions:

- **Deployments**: create, get, list, watch, update, patch, delete
- **Services**: create, get, list, watch, update, patch, delete
- **Pods**: get, list, watch
- **Pod Logs**: get
- **Jobs**: create, get, list, watch, delete

## ⚙️ Configuration

### Environment Variables

**Backend (`api`)**

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `NAMESPACE` | `default` | Kubernetes namespace for functions |
| `JWT_SECRET` | - | JWT signing secret (required) |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `METRICS_PORT` | `9090` | Prometheus metrics port |

**Frontend (`web`)**

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_URL` | - | API base URL (optional, uses proxy) |

### Kubernetes Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: eventflow-secrets
  namespace: eventflow
type: Opaque
stringData:
  jwt-secret: "your-secure-secret-here-min-32-chars"
```

## 🧪 Testing

### Test Function Deployment

```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/token | jq -r '.token')

# Create a function
curl -X POST http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-nginx",
    "image": "nginx:alpine",
    "replicas": 1
  }'

# List functions
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/functions

# Invoke function
curl -X POST "http://localhost:8080/v1/functions/test-nginx:invoke" \
  -H "Authorization: Bearer $TOKEN"

# Get logs
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/v1/functions/test-nginx/logs"

# Delete function
curl -X DELETE "http://localhost:8080/v1/functions/test-nginx" \
  -H "Authorization: Bearer $TOKEN"
```

## 📊 Metrics

Prometheus metrics available at `/metrics`:

- `eventflow_function_invocations_total` - Total function invocations
- `eventflow_function_duration_seconds` - Function invocation duration
- `eventflow_http_requests_total` - HTTP requests by method/path/status
- `eventflow_active_functions` - Number of active functions

## 🔒 Security

- JWT authentication with configurable secrets
- Kubernetes RBAC for fine-grained permissions
- Namespace isolation for multi-tenancy
- Secret management via Kubernetes secrets
- Network policies (can be added)

## 🚧 Roadmap / Stretch Goals

- [x] WebSocket/SSE log streaming
- [x] Prometheus metrics integration
- [ ] Helm chart for easy deployment
- [ ] OpenTelemetry tracing
- [ ] Function templates/marketplace
- [ ] Autoscaling based on custom metrics
- [ ] Multi-cluster support
- [ ] Function versioning and rollbacks

## 📝 License

MIT License - feel free to use this project as you wish!

## 🤝 Contributing

Contributions welcome! Please open an issue or PR.

---

**Built with ❤️ using Go, React, and Kubernetes**
