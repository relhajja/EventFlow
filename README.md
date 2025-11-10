# EventFlow - Multi-Tenant Kubernetes FaaS Platform

<div align="center">

**Enterprise-grade Functions-as-a-Service with True Multi-Tenancy**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)](https://react.dev/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28+-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-4169E1?style=flat&logo=postgresql)](https://postgresql.org/)
[![NATS](https://img.shields.io/badge/NATS-2.10+-27AAE1?style=flat&logo=nats.io)](https://nats.io/)

[Why EventFlow?](#-why-eventflow) • [Key Features](#-key-features) • [Architecture](#-architecture) • [Quick Start](#-quick-start) • [Documentation](./docs)

</div>

---

## 🎯 Why EventFlow?

### The Problem

Modern cloud-native applications need:
- **Serverless compute** without vendor lock-in
- **Strong isolation** between different teams or customers
- **Resource governance** to prevent noisy neighbors
- **Self-service** function deployment without cluster admin access
- **Cost control** through resource quotas and limits

Existing FaaS platforms either:
- Lack true multi-tenancy (shared namespaces)
- Require complex infrastructure (AWS Lambda, Google Cloud Functions)
- Don't integrate natively with Kubernetes
- Have weak isolation boundaries

### The Solution: EventFlow

EventFlow is a **Kubernetes-native FaaS platform** that provides **namespace-per-tenant isolation**, making it ideal for:

🏢 **Enterprise SaaS Platforms**
- Give each customer their own isolated namespace
- Enforce resource quotas per customer
- Prevent cross-tenant access

👥 **Multi-Team Organizations**
- Separate development, staging, and production workloads
- Enforce team-based resource limits
- Audit and track resource usage per team

🎓 **Educational Institutions**
- Provide isolated environments for students
- Limit resource consumption per user
- Easy cleanup after courses end

☁️ **Cloud Service Providers**
- Offer FaaS as a managed service
- Strong tenant isolation guarantees
- Predictable billing per tenant

## ✨ Key Features

### 🔐 True Multi-Tenancy
- **Namespace Isolation**: Each user gets their own Kubernetes namespace (`tenant-{userId}`)
- **Resource Quotas**: Automatic CPU, memory, and pod limits per tenant
- **JWT Authentication**: User context embedded in every request
- **Scoped Operations**: Users can only see and manage their own functions

### ⚡ Kubernetes-Native
- **Custom Resource Definition (CRD)**: Functions are Kubernetes resources
- **Operator Pattern**: Automatic reconciliation of desired state
- **Standard Tools**: Works with kubectl, Helm, and other K8s tools
- **Owner References**: Automatic cleanup when functions are deleted

### 🎨 Developer-Friendly
- **Modern Web UI**: React + TypeScript dashboard
- **REST API**: Standard HTTP/JSON interface
- **Real-time Logs**: Stream function logs directly from pods
- **Source Code Builds**: Deploy from Python, Node.js, or Go source code
- **Event-Driven Architecture**: NATS-based async build system
- **Quick Setup**: Deploy to K3s cluster in minutes

### 📊 Enterprise Ready
- **PostgreSQL Database**: Persistent function metadata and build jobs
- **NATS JetStream**: Event-driven messaging and build queue
- **Docker Registry**: In-cluster image storage
- **Async Build Workers**: Background image building from source code
- **Prometheus Metrics**: Built-in observability
- **RBAC Integration**: Kubernetes-native authorization
- **Health Checks**: Liveness and readiness probes
- **Resource Limits**: Automatic CPU/memory requests and limits

## 🏗️ Architecture

### High-Level Overview

```
┌────────────────────────────────────────────────────────────────────┐
│                        EventFlow Platform                           │
├────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────┐       ┌──────────────┐      ┌───────────────┐  │
│  │   Web UI     │──────▶│  API Server  │─────▶│ NATS/JetStream│  │
│  │  (React)     │       │    (Go)      │      │  (Events)     │  │
│  └──────────────┘       └───────┬──────┘      └───────┬───────┘  │
│                                  │                     │           │
│                         ┌────────┼─────────────────────┘           │
│                         │        │        │                        │
│                         ▼        ▼        ▼                        │
│                  ┌─────────┐ ┌─────────┐ ┌─────────┐             │
│                  │   JWT   │ │Database │ │ Builder │             │
│                  │  Auth   │ │  (Pg)   │ │ Worker  │             │
│                  └─────────┘ └─────────┘ └────┬────┘             │
│                                                │                   │
│                                       ┌────────┼────────┐          │
│                                       │        │        │          │
│                                       ▼        ▼        ▼          │
│                                   ┌────────┐ ┌────────┐ ┌────────┐│
│                                   │ Docker │ │  K8s   │ │Registry││
│                                   │ Socket │ │  API   │ │        ││
│                                   └────────┘ └───┬────┘ └────────┘│
│                                                   │                 │
│  ┌────────────────────────────────────────────────┼──────────────┐│
│  │           Kubernetes Cluster                   │              ││
│  │                                                 ▼              ││
│  │                                         ┌──────────────┐      ││
│  │                                         │   Operator   │      ││
│  │                                         │(Kubebuilder) │      ││
│  │                                         └──────┬───────┘      ││
│  │                                                │              ││
│  │     ┌──────────────────────────────────────────┘              ││
│  │     │                                                         ││
│  │     ▼                                                         ││
│  │  ┌─────────────────────────────────────────────┐            ││
│  │  │      Tenant Namespaces                      │            ││
│  │  │                                             │            ││
│  │  │  tenant-alice    tenant-bob                 │            ││
│  │  │  ├─ Quota        ├─ Quota                   │            ││
│  │  │  ├─ Functions    ├─ Functions               │            ││
│  │  │  ├─ Deployments  ├─ Deployments             │            ││
│  │  │  └─ Pods         └─ Pods                    │            ││
│  │  │                                             │            ││
│  │  │  tenant-charlie  tenant-demo                │            ││
│  │  │  ├─ Quota        ├─ Quota                   │            ││
│  │  │  ├─ Functions    ├─ Functions               │            ││
│  │  │  ├─ Deployments  ├─ Deployments             │            ││
│  │  │  └─ Pods         └─ Pods                    │            ││
│  │  └─────────────────────────────────────────────┘            ││
│  └────────────────────────────────────────────────────────────────┘│
└────────────────────────────────────────────────────────────────────┘
```

### Multi-Tenant Isolation Model

**Namespace-Per-Tenant** provides the strongest isolation in Kubernetes:

| Aspect | Implementation | Benefit |
|--------|---------------|---------|
| **Compute** | Separate namespaces | Network policies, resource quotas |
| **Storage** | Namespaced PVCs | Isolated persistent volumes |
| **Network** | Network policies | Control ingress/egress traffic |
| **Identity** | RBAC per namespace | Fine-grained access control |
| **Resources** | Resource quotas | Prevent noisy neighbor problems |

### Request Flow

#### Option 1: Deploy Pre-built Image

```
1. User Login (Web UI)
   │
   └─▶ POST /auth/token → JWT Token

2. Create Function (Pre-built Image)
   │
   ├─▶ POST /v1/functions {"image": "nginx:alpine"}
   ├─▶ API saves to PostgreSQL
   └─▶ Creates Function CR in tenant namespace

3. Operator Reconciliation
   │
   ├─▶ Watches Function CRs
   ├─▶ Creates Deployment with image
   └─▶ Updates Function.Status → Running
```

#### Option 2: Build from Source Code (Event-Driven)

```
1. User Submits Source Code
   │
   ├─▶ POST /v1/functions {"runtime": "python", "source_code": "..."}
   │
   ├─▶ API creates build_job (status: pending)
   │
   ├─▶ API publishes NATS event → "build.created"
   │
   └─▶ Returns build_id immediately (202 Accepted)

2. Builder Worker (Event-Driven)
   │
   ├─▶ Subscribes to NATS "eventflow.events"
   │
   ├─▶ Receives build event instantly (<100ms)
   │
   ├─▶ Fetches job from database
   │
   ├─▶ Generates Dockerfile (Python/Node.js/Go)
   │
   ├─▶ Builds Docker image (status: building)
   │
   ├─▶ Pushes to in-cluster registry (status: pushing)
   │
   └─▶ Updates database (status: success)

3. Automatic Deployment
   │
   ├─▶ Worker creates Function CR with built image
   │
   ├─▶ Operator deploys to tenant namespace
   │
   └─▶ Function running from source code

Fallback: If NATS event missed, worker polls every 30s
```

## 🚀 Quick Start

### Prerequisites

```bash
# Install required tools
curl -sfL https://get.k3s.io | sh -
brew install docker kubectl go node task

# Verify versions
docker --version   # 20.10+
kubectl version    # 1.28+
go version         # 1.23+
node --version     # 18+
task --version     # 3.0+
```

### Deploy to K3s in 5 Minutes

```bash
# 1. Clone the repository
git clone https://github.com/relhajja/eventflow.git
cd eventflow

# 2. Install K3s (if not installed)
cd infra-dev && task k3s:install

# 3. Deploy everything
task deploy

# 4. Wait for all pods to be ready
task monitor

# 5. Access the dashboard
open http://localhost:30300
```

### Alternative: Deploy to Kind

```bash
cd eventflow
kind create cluster --name eventflow --config kind-config.yaml
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/nats.yaml
kubectl apply -f k8s/registry.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/operator.yaml
kubectl apply -f k8s/builder.yaml
```

### Test Multi-Tenancy

```bash
# Using Task (recommended)
task test:all

# Or manually:

# Get tokens
TOKEN_ALICE=$(task test:get-token USER=alice)
TOKEN_BOB=$(task test:get-token USER=bob)

# Create function from pre-built image (Alice)
curl -X POST http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN_ALICE" \
  -H "Content-Type: application/json" \
  -d '{"name":"alice-nginx","image":"nginx:alpine","replicas":1}'

# Create function from source code (Bob)
cat > handler.py <<EOF
from http.server import HTTPServer, BaseHTTPRequestHandler
import json

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        response = {"message": "Hello from Bob's Python function!"}
        self.wfile.write(json.dumps(response).encode())

HTTPServer(('0.0.0.0', 8080), Handler).serve_forever()
EOF

SOURCE_CODE=$(base64 -w0 handler.py)

curl -X POST http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN_BOB" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"bob-python\",
    \"runtime\": \"python\",
    \"source_code\": \"$SOURCE_CODE\",
    \"replicas\": 1
  }"

# Check build status
BUILD_ID=$(curl -s http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN_BOB" | jq -r '.[0].build_id')

curl http://localhost:30080/v1/builds/$BUILD_ID \
  -H "Authorization: Bearer $TOKEN_BOB" | jq .

# Stream build logs
curl -N http://localhost:30080/v1/builds/$BUILD_ID/logs/stream \
  -H "Authorization: Bearer $TOKEN_BOB"

# Verify isolation
curl http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN_ALICE" | jq .  # Only sees alice-nginx

curl http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN_BOB" | jq .    # Only sees bob-python

# Check tenant namespaces
kubectl get namespaces | grep tenant
kubectl get functions -A
kubectl get pods -A | grep fn-
```

## 📚 Documentation

Detailed documentation is available in the [`./docs`](./docs) directory:

- **[Architecture Guide](./docs/ARCHITECTURE.md)** - Deep dive into system design
- **[Deployment Guide](./docs/DEPLOYMENT.md)** - Production deployment strategies
- **[API Reference](./docs/API.md)** - Complete API documentation
- **[Event-Driven Builds](./docs/EVENT-DRIVEN-BUILDS.md)** - NATS-based async build system
- **[Builder Implementation](./docs/BUILDER-IMPLEMENTATION.md)** - Source code build architecture
- **[Async Build Quickstart](./docs/ASYNC-BUILD-QUICKSTART.md)** - Deploy build system in 5 minutes
- **[Task Reference](./infra-dev/TASK-REFERENCE.md)** - All Taskfile commands
- **[Troubleshooting](./infra-dev/TROUBLESHOOTING.md)** - Common issues and fixes

## 🎓 Use Cases

### 1. SaaS Platform with Customer Isolation

**Scenario**: You're building a data processing platform where each customer uploads files and runs processing jobs.

**EventFlow Solution**:
- Each customer gets their own namespace
- Resource quotas prevent one customer from consuming all resources
- Functions process customer data in isolation
- Billing based on actual resource usage per namespace

### 2. Internal Developer Platform

**Scenario**: Your company has 50+ engineering teams, each needs to deploy microservices and batch jobs.

**EventFlow Solution**:
- Each team gets a namespace (tenant-team-name)
- Teams self-service deploy functions without cluster admin access
- Central platform team sets resource quotas per team
- Audit logs track which team deployed what

### 3. Educational Platform

**Scenario**: University course with 200 students, each needs to deploy and test applications.

**EventFlow Solution**:
- Each student gets a namespace (tenant-student-id)
- Resource limits prevent accidental resource exhaustion
- Easy cleanup at end of semester (delete namespaces)
- Instructors can monitor student deployments

### 4. CI/CD Pipeline Functions

**Scenario**: Run isolated build, test, and deployment functions for each git branch.

**EventFlow Solution**:
- Each branch gets a temporary namespace
- Functions run tests, builds, deployments
- Automatic cleanup when branch is merged/deleted
- Parallel execution without interference

## 🔒 Security Model

### Authentication & Authorization

```
User Request
    │
    ├─▶ JWT Token (Bearer)
    │   ├─ user_id: "alice"
    │   ├─ namespace: "tenant-alice"
    │   └─ email: "alice@company.com"
    │
    ├─▶ API Middleware Validates Token
    │
    ├─▶ Extract User Context
    │
    ├─▶ All DB Queries: WHERE user_id = 'alice'
    │
    └─▶ All K8s Operations: namespace = 'tenant-alice'
```

### Isolation Guarantees

1. **API Level**: Database queries always filter by user_id
2. **Kubernetes Level**: Functions created in separate namespaces
3. **Network Level**: (Optional) NetworkPolicies between namespaces
4. **Resource Level**: ResourceQuotas prevent resource exhaustion
5. **RBAC Level**: API has limited permissions, users have no direct K8s access

## 📊 Resource Defaults

### Per-Tenant Quotas

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-quota
spec:
  hard:
    requests.cpu: "10"          # 10 CPU cores total
    requests.memory: 20Gi       # 20 GB memory total
    limits.cpu: "20"            # 20 CPU cores limit
    limits.memory: 40Gi         # 40 GB memory limit
    pods: "50"                  # Max 50 pods
    persistentvolumeclaims: "10" # Max 10 PVCs
```

### Per-Function Defaults

```yaml
resources:
  requests:
    cpu: 100m      # 0.1 CPU core
    memory: 128Mi  # 128 MB
  limits:
    cpu: 500m      # 0.5 CPU core
    memory: 512Mi  # 512 MB
```

## 🛠️ Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Backend** | Go 1.23 + chi router | High-performance API server |
| **Frontend** | React 18 + TypeScript | Modern web dashboard |
| **Database** | PostgreSQL 13 | Function metadata & build jobs |
| **Messaging** | NATS JetStream | Event-driven architecture |
| **Operator** | Kubebuilder v4 | Kubernetes controller |
| **Builder** | Docker + Go | Source code → container images |
| **Registry** | Docker Registry v2 | In-cluster image storage |
| **Container** | Docker | Containerization |
| **Orchestration** | K3s / Kubernetes 1.28+ | Container orchestration |
| **Auth** | JWT (HS256) | Stateless authentication |
| **Metrics** | Prometheus | Observability |
| **Build** | Vite | Fast frontend builds |
| **Styling** | Tailwind CSS | Utility-first CSS |
| **Task Runner** | Task (go-task) | Modern make alternative |

## 🤝 Contributing

We welcome contributions! 

### Quick Start for Contributors

```bash
# Fork and clone the repository
git clone https://github.com/relhajja/eventflow.git
cd eventflow

# Install dependencies
cd infra-dev && task install

# Create a feature branch
git checkout -b feature/my-feature

# Make your changes

# Test locally
task deploy
task test:all

# Commit and push
git commit -m "Add my feature"
git push origin feature/my-feature

# Open a Pull Request
```

## 📝 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with these amazing open-source projects:

- [Kubernetes](https://kubernetes.io/) - Container orchestration
- [Kubebuilder](https://book.kubebuilder.io/) - Kubernetes operator framework
- [chi](https://github.com/go-chi/chi) - Lightweight HTTP router
- [React](https://react.dev/) - UI library
- [PostgreSQL](https://postgresql.org/) - Relational database
- [Tailwind CSS](https://tailwindcss.com/) - CSS framework

## 📞 Support

- **Documentation**: [docs/](./docs)
- **Issues**: [GitHub Issues](https://github.com/relhajja/eventflow/issues)
- **Discussions**: [GitHub Discussions](https://github.com/relhajja/eventflow/discussions)

---

<div align="center">

**Built with ❤️ for the cloud-native community**

[⭐ Star us on GitHub](https://github.com/relhajja/eventflow) • [📖 Read the Docs](./docs) • [🐛 Report Bug](https://github.com/relhajja/eventflow/issues)

</div>
