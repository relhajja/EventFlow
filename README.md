# EventFlow - Multi-Tenant Kubernetes FaaS Platform

<div align="center">

**Enterprise-grade Functions-as-a-Service with True Multi-Tenancy**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)](https://react.dev/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28+-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-4169E1?style=flat&logo=postgresql)](https://postgresql.org/)

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
- **Quick Setup**: Deploy to kind cluster in minutes

### 📊 Enterprise Ready
- **PostgreSQL Database**: Persistent function metadata storage
- **Prometheus Metrics**: Built-in observability
- **RBAC Integration**: Kubernetes-native authorization
- **Health Checks**: Liveness and readiness probes
- **Resource Limits**: Automatic CPU/memory requests and limits

## 🏗️ Architecture

### High-Level Overview

```
┌────────────────────────────────────────────────────────────┐
│                    EventFlow Platform                       │
├────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐       ┌──────────────┐                  │
│  │   Web UI     │──────▶│  API Server  │                  │
│  │  (React)     │       │    (Go)      │                  │
│  └──────────────┘       └───────┬──────┘                  │
│                                  │                          │
│                         ┌────────┼────────┐                │
│                         │        │        │                │
│                         ▼        ▼        ▼                │
│                  ┌─────────┐ ┌─────────┐ ┌─────────┐     │
│                  │   JWT   │ │Database │ │   K8s   │     │
│                  │  Auth   │ │  (Pg)   │ │   API   │     │
│                  └─────────┘ └─────────┘ └────┬────┘     │
│                                                │           │
│  ┌─────────────────────────────────────────────┼────────┐ │
│  │           Kubernetes Cluster                │        │ │
│  │                                              ▼        │ │
│  │                                      ┌──────────────┐│ │
│  │                                      │   Operator   ││ │
│  │                                      │(Kubebuilder) ││ │
│  │                                      └──────┬───────┘│ │
│  │                                             │        │ │
│  │     ┌───────────────────────────────────────┘        │ │
│  │     │                                                │ │
│  │     ▼                                                │ │
│  │  ┌──────────────────────────────────────┐          │ │
│  │  │      Tenant Namespaces               │          │ │
│  │  │                                      │          │ │
│  │  │  tenant-alice    tenant-bob          │          │ │
│  │  │  ├─ Quota        ├─ Quota            │          │ │
│  │  │  ├─ Functions    ├─ Functions        │          │ │
│  │  │  ├─ Deployments  ├─ Deployments      │          │ │
│  │  │  └─ Pods         └─ Pods             │          │ │
│  │  │                                      │          │ │
│  │  │  tenant-charlie  tenant-demo         │          │ │
│  │  │  ├─ Quota        ├─ Quota            │          │ │
│  │  │  ├─ Functions    ├─ Functions        │          │ │
│  │  │  ├─ Deployments  ├─ Deployments      │          │ │
│  │  │  └─ Pods         └─ Pods             │          │ │
│  │  └──────────────────────────────────────┘          │ │
│  └──────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────┘
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

```
1. User Login (Web UI)
   │
   ├─▶ Select User (alice/bob/charlie/demo)
   │
   └─▶ POST /auth/token
       ↓
       JWT Token (includes user_id, namespace)

2. Create Function
   │
   ├─▶ POST /v1/functions (with JWT)
   │
   ├─▶ API extracts user_id from token
   │
   ├─▶ Generate namespace: tenant-{user_id}
   │
   ├─▶ EnsureNamespace() creates if not exists
   │
   ├─▶ Apply ResourceQuota (CPU, memory, pods)
   │
   ├─▶ Save to PostgreSQL (user_id, namespace, metadata)
   │
   └─▶ Create Function CR in tenant namespace

3. Operator Reconciliation
   │
   ├─▶ Watch Function CRs across all namespaces
   │
   ├─▶ Create Deployment (fn-{name}) with owner ref
   │
   ├─▶ Set resource requests/limits automatically
   │
   ├─▶ Wait for pods to be ready
   │
   └─▶ Update Function.Status (Running/Failed)

4. Function Running
   │
   ├─▶ Pods scheduled in tenant namespace
   │
   ├─▶ User can invoke via API
   │
   ├─▶ User can stream logs
   │
   └─▶ HPA can scale based on load
```

## 🚀 Quick Start

### Prerequisites

```bash
# Install required tools
brew install docker kind kubectl go node

# Verify versions
docker --version   # 20.10+
kind --version     # 0.20+
kubectl version    # 1.28+
go version         # 1.22+
node --version     # 18+
```

### Deploy in 5 Minutes

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/eventflow.git
cd eventflow/eventflow

# 2. Create kind cluster
kind create cluster --name eventflow --config kind-config.yaml

# 3. Deploy PostgreSQL
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres.yaml
kubectl wait --for=condition=ready pod -l app=postgres -n eventflow --timeout=120s

# 4. Deploy the platform
make deploy

# 5. Access the dashboard
kubectl port-forward -n eventflow svc/eventflow-web 3000:80 &
open http://localhost:3000
```

### Test Multi-Tenancy

```bash
# Terminal 1: Port forward the API
kubectl port-forward -n eventflow svc/eventflow-api 8080:80

# Terminal 2: Create function as Alice
TOKEN_ALICE=$(curl -s -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","username":"alice","email":"alice@company.com"}' \
  | jq -r .token)

curl -X POST http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN_ALICE" \
  -H "Content-Type: application/json" \
  -d '{"name":"alice-app","image":"nginx:alpine","replicas":1}'

# Terminal 3: Create function as Bob
TOKEN_BOB=$(curl -s -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/json" \
  -d '{"user_id":"bob","username":"bob","email":"bob@company.com"}' \
  | jq -r .token)

curl -X POST http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN_BOB" \
  -H "Content-Type: application/json" \
  -d '{"name":"bob-app","image":"nginx:alpine","replicas":1}'

# Verify isolation - Alice can only see her functions
curl http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN_ALICE" | jq .

# Verify isolation - Bob can only see his functions
curl http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN_BOB" | jq .

# Check tenant namespaces
kubectl get namespaces | grep tenant
# tenant-alice    Active   2m
# tenant-bob      Active   1m

# Check resource quotas
kubectl get resourcequota -n tenant-alice
kubectl get resourcequota -n tenant-bob
```

## 📚 Documentation

Detailed documentation is available in the [`./docs`](./docs) directory:

- **[Architecture Guide](./docs/ARCHITECTURE.md)** - Deep dive into system design
- **[Deployment Guide](./docs/DEPLOYMENT.md)** - Production deployment strategies
- **[API Reference](./docs/API.md)** - Complete API documentation
- **[Operator Guide](./docs/OPERATOR.md)** - Kubernetes operator internals
- **[Security Guide](./docs/SECURITY.md)** - Security best practices
- **[Development Guide](./docs/DEVELOPMENT.md)** - Contributing and development setup

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
| **Backend** | Go 1.22 + chi router | High-performance API server |
| **Frontend** | React 18 + TypeScript | Modern web dashboard |
| **Database** | PostgreSQL 16 | Function metadata storage |
| **Operator** | Kubebuilder v4 | Kubernetes controller |
| **Container** | Docker | Containerization |
| **Orchestration** | Kubernetes 1.28+ | Container orchestration |
| **Auth** | JWT (HS256) | Stateless authentication |
| **Metrics** | Prometheus | Observability |
| **Build** | Vite | Fast frontend builds |
| **Styling** | Tailwind CSS | Utility-first CSS |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](./docs/CONTRIBUTING.md) for details.

### Quick Start for Contributors

```bash
# Fork and clone the repository
git clone https://github.com/yourusername/eventflow.git
cd eventflow

# Create a feature branch
git checkout -b feature/my-feature

# Make your changes and test
make test

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
