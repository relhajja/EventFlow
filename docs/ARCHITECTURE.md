# EventFlow Architecture Guide

## Table of Contents

- [Overview](#overview)
- [System Components](#system-components)
- [Multi-Tenant Architecture](#multi-tenant-architecture)
- [Data Flow](#data-flow)
- [Security Architecture](#security-architecture)
- [Scalability](#scalability)

## Overview

EventFlow implements a **namespace-per-tenant** architecture where each user operates in an isolated Kubernetes namespace. This provides the strongest form of multi-tenancy in Kubernetes, with clear resource boundaries and security isolation.

### Design Principles

1. **Kubernetes-Native**: Everything is a Kubernetes resource
2. **Declarative**: Desired state expressed through CRDs
3. **Reconciliation**: Operator ensures actual state matches desired state
4. **Isolation**: Strong boundaries between tenants
5. **Self-Service**: Users manage their own functions without cluster access

## System Components

### 1. API Server (Go)

**Location**: `eventflow/api/`

**Responsibilities**:
- Handle HTTP requests from web UI and external clients
- JWT authentication and user context extraction
- PostgreSQL database operations
- Kubernetes API interactions via client-go
- Automatic namespace creation and quota management

**Key Files**:
- `main.go` - Entry point and server initialization
- `internal/auth/jwt.go` - JWT token generation and validation
- `internal/handlers/functions.go` - HTTP request handlers
- `internal/k8s/client.go` - Kubernetes client wrapper
- `internal/database/functions.go` - PostgreSQL repository

**API Flow**:
```
HTTP Request
    ↓
JWT Middleware (extract user_id)
    ↓
Handler (CreateFunction, ListFunctions, etc.)
    ↓
Database Repository (save metadata)
    ↓
Kubernetes Client (create Function CR)
    ↓
HTTP Response
```

### 2. Operator (Kubebuilder)

**Location**: `eventflow/operator/`

**Responsibilities**:
- Watch Function custom resources across all namespaces
- Reconcile Functions to Deployments
- Set resource requests and limits
- Update Function status based on pod health
- Handle Function lifecycle (create, update, delete)

**Key Files**:
- `api/v1alpha1/function_types.go` - Function CRD definition
- `internal/controller/function_controller.go` - Reconciliation logic
- `config/crd/` - CRD manifests
- `config/rbac/` - RBAC rules

**Reconciliation Loop**:
```
Watch Function CRs
    ↓
Compare desired vs actual state
    ↓
Create/Update/Delete Deployment
    ↓
Wait for Deployment to be ready
    ↓
Update Function.Status
    ↓
Requeue if needed
```

### 3. Web Dashboard (React + TypeScript)

**Location**: `eventflow/web/`

**Responsibilities**:
- User authentication (JWT)
- Function management UI
- Real-time status updates
- Log streaming viewer
- Error handling and notifications

**Key Files**:
- `src/pages/Login.tsx` - User selection and authentication
- `src/pages/Dashboard.tsx` - Function list view
- `src/pages/CreateFunction.tsx` - Function creation form
- `src/pages/FunctionDetails.tsx` - Function details and logs
- `src/context/AuthContext.tsx` - Authentication state management
- `src/services/api.ts` - API client

**Component Hierarchy**:
```
App (Router)
  ├─ AuthProvider
  │   ├─ Login
  │   └─ ProtectedRoute
  │       ├─ Layout
  │       │   ├─ Dashboard
  │       │   ├─ CreateFunction
  │       │   └─ FunctionDetails
```

### 4. PostgreSQL Database

**Location**: Deployed in Kubernetes

**Schema**:
```sql
CREATE TABLE functions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL,
    image VARCHAR(255) NOT NULL,
    replicas INT DEFAULT 1,
    command TEXT[],
    env JSONB,
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT functions_name_namespace_user_key 
        UNIQUE (name, namespace, user_id)
);

CREATE INDEX idx_functions_user_id 
    ON functions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_functions_namespace 
    ON functions(namespace) WHERE deleted_at IS NULL;
```

**Key Design Decisions**:
- **Composite Unique Key**: (name, namespace, user_id) allows same function name across tenants
- **Soft Deletes**: deleted_at timestamp for audit trail
- **JSONB for env**: Flexible environment variable storage
- **User ID**: Every query filters by user_id for isolation

## Multi-Tenant Architecture

### Namespace-Per-Tenant Model

Each user gets a dedicated Kubernetes namespace:

```
tenant-alice/
  ├─ ResourceQuota (tenant-quota)
  ├─ Function CRs
  ├─ Deployments (fn-*)
  └─ Pods

tenant-bob/
  ├─ ResourceQuota (tenant-quota)
  ├─ Function CRs
  ├─ Deployments (fn-*)
  └─ Pods
```

### Namespace Creation Flow

```go
// 1. User creates function
POST /v1/functions
Authorization: Bearer <JWT with user_id=alice>

// 2. API extracts user context
claims := auth.GetUserFromContext(ctx)
namespace := fmt.Sprintf("tenant-%s", claims.UserID) // tenant-alice

// 3. Ensure namespace exists
k8sClient.EnsureNamespace(ctx, namespace)

// 4. Apply resource quota
k8sClient.createResourceQuota(ctx, namespace)

// 5. Create Function CR in tenant namespace
dynamicClient.Create(ctx, functionCR, namespace)
```

### Resource Quotas

Each tenant namespace gets automatic quotas to prevent resource exhaustion:

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-quota
  namespace: tenant-alice
spec:
  hard:
    requests.cpu: "10"              # Total CPU requests
    requests.memory: "20Gi"         # Total memory requests
    limits.cpu: "20"                # Total CPU limits
    limits.memory: "40Gi"           # Total memory limits
    pods: "50"                      # Maximum pods
    persistentvolumeclaims: "10"    # Maximum PVCs
```

Benefits:
- **Predictable costs**: Know max resources per tenant
- **Fair sharing**: No single tenant can starve others
- **Billing**: Charge based on quota allocation
- **Safety**: Prevent accidental resource exhaustion

## Data Flow

### Creating a Function

```
┌──────────┐     1. POST /v1/functions      ┌──────────┐
│   Web    │────────────────────────────────>│   API    │
│    UI    │<────────────────────────────────│  Server  │
└──────────┘     2. HTTP 201 Created        └────┬─────┘
                                                  │
                                          3. Save metadata
                                                  │
                                                  v
                                            ┌──────────┐
                                            │PostgreSQL│
                                            └──────────┘
                                                  │
                                          4. Create Function CR
                                                  │
                                                  v
                            ┌─────────────────────────────────┐
                            │      Kubernetes Cluster          │
                            │                                  │
                            │  5. Operator watches             │
                            │     ┌──────────┐                │
                            │     │ Operator │                │
                            │     └────┬─────┘                │
                            │          │ 6. Reconcile         │
                            │          v                       │
                            │     ┌──────────┐                │
                            │     │Deployment│                │
                            │     └────┬─────┘                │
                            │          │ 7. Create            │
                            │          v                       │
                            │     ┌──────────┐                │
                            │     │   Pods   │                │
                            │     └──────────┘                │
                            │          │ 8. Running           │
                            │          v                       │
                            │     ┌──────────┐                │
                            │     │ Update   │                │
                            │     │ Status   │                │
                            │     └──────────┘                │
                            └─────────────────────────────────┘
```

### Listing Functions

```
┌──────────┐     1. GET /v1/functions       ┌──────────┐
│   Web    │────────────────────────────────>│   API    │
│    UI    │<────────────────────────────────│  Server  │
└──────────┘     2. JSON Response           └────┬─────┘
                                                  │
                                    3. SELECT * FROM functions
                                       WHERE user_id = ?
                                          AND deleted_at IS NULL
                                                  │
                                                  v
                                            ┌──────────┐
                                            │PostgreSQL│
                                            └──────────┘
```

### Invoking a Function

```
┌──────────┐  1. POST /v1/functions/name:invoke  ┌──────────┐
│   Web    │─────────────────────────────────────>│   API    │
│    UI    │<─────────────────────────────────────│  Server  │
└──────────┘  2. HTTP 200 Invocation Started    └────┬─────┘
                                                       │
                                           3. Create Kubernetes Job
                                                       │
                                                       v
                            ┌──────────────────────────────────┐
                            │      Kubernetes Cluster           │
                            │                                   │
                            │     ┌──────────┐                 │
                            │     │   Job    │                 │
                            │     └────┬─────┘                 │
                            │          │ 4. Run                │
                            │          v                        │
                            │     ┌──────────┐                 │
                            │     │   Pod    │                 │
                            │     └────┬─────┘                 │
                            │          │ 5. Execute            │
                            │          v                        │
                            │     ┌──────────┐                 │
                            │     │  Result  │                 │
                            │     └──────────┘                 │
                            └──────────────────────────────────┘
```

## Security Architecture

### Authentication Flow

```
1. User Login (Web UI)
   ├─> Select user (alice/bob/charlie)
   ├─> POST /auth/token
   │   Body: {user_id, username, email}
   └─> Response: {token, user, namespace}

2. JWT Token Structure
   {
     "user_id": "alice",
     "username": "Alice",
     "email": "alice@company.com",
     "namespace": "tenant-alice",
     "exp": 1699564800
   }

3. Authenticated Requests
   GET /v1/functions
   Header: Authorization: Bearer <token>
   
4. API Middleware
   ├─> Validate JWT signature
   ├─> Check expiration
   ├─> Extract claims
   └─> Store in request context

5. Handler Authorization
   ├─> Get user from context
   ├─> Filter queries by user_id
   └─> Scope K8s operations to user's namespace
```

### RBAC Model

**API Server ServiceAccount**:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: eventflow-api-clusterrole
rules:
  # Namespace management
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list", "watch", "create"]
  
  # Resource quota management
  - apiGroups: [""]
    resources: ["resourcequotas"]
    verbs: ["get", "list", "create", "update"]
  
  # Function CRD management
  - apiGroups: ["eventflow.eventflow.io"]
    resources: ["functions"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  
  # Read-only access to deployments/pods for status
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch"]
  
  - apiGroups: [""]
    resources: ["pods", "pods/log"]
    verbs: ["get", "list", "watch"]
```

**Operator ServiceAccount**:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: operator-manager-role
rules:
  # Full control over Functions
  - apiGroups: ["eventflow.eventflow.io"]
    resources: ["functions"]
    verbs: ["*"]
  
  - apiGroups: ["eventflow.eventflow.io"]
    resources: ["functions/status", "functions/finalizers"]
    verbs: ["get", "patch", "update"]
  
  # Deployment management
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["*"]
  
  # Pod read access for status
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
```

### Data Isolation

**Database Level**:
```sql
-- All queries filter by user_id
SELECT * FROM functions 
WHERE user_id = $1 
  AND deleted_at IS NULL;

-- Unique constraint includes user_id
CONSTRAINT functions_name_namespace_user_key 
    UNIQUE (name, namespace, user_id);
```

**Kubernetes Level**:
```go
// All Function CRs created in tenant namespace
namespace := fmt.Sprintf("tenant-%s", userID)
dynamicClient.Namespace(namespace).Create(ctx, functionCR)
```

**API Level**:
```go
// User context enforced at middleware level
claims := auth.GetUserFromContext(r.Context())
functions := db.List(ctx, claims.UserID)
```

## Scalability

### Horizontal Scaling

**API Server**:
- Stateless design allows horizontal scaling
- Current deployment: 2 replicas
- Can scale to 10+ replicas behind load balancer
- PostgreSQL connection pooling

**Operator**:
- Single active replica (leader election)
- Watches all namespaces efficiently
- Can handle 1000+ Function CRs
- Concurrent reconciliation with work queues

**Database**:
- PostgreSQL with read replicas
- Connection pooling (pgBouncer)
- Indexes on user_id and namespace
- Soft deletes for audit without delete overhead

### Vertical Scaling

**Per-Tenant Limits**:
- Adjust ResourceQuota based on tier (free/pro/enterprise)
- Different quotas for dev/staging/prod namespaces

**Per-Function Limits**:
- Default: 100m CPU / 128Mi memory
- Allow users to specify custom resources
- Operator validates against tenant quota

### Performance Optimizations

1. **Database**:
   - Indexed queries on user_id and namespace
   - JSONB for flexible env vars without schema changes
   - Soft deletes avoid expensive cascading deletes

2. **Kubernetes**:
   - Owner references for automatic garbage collection
   - Filtered watches (namespace-scoped when possible)
   - Informer caching reduces API server load

3. **API**:
   - JWT for stateless auth (no session storage)
   - Kubernetes client connection pooling
   - HTTP keep-alive for persistent connections

4. **Frontend**:
   - React Query for caching and deduplication
   - Lazy loading of function details
   - Optimistic UI updates

## Monitoring and Observability

### Metrics (Prometheus)

**API Metrics**:
- `eventflow_api_requests_total` - Total HTTP requests
- `eventflow_api_request_duration_seconds` - Request latency
- `eventflow_functions_total` - Total functions per user
- `eventflow_namespace_quota_cpu` - CPU quota per namespace
- `eventflow_namespace_quota_memory` - Memory quota per namespace

**Operator Metrics**:
- `eventflow_reconcile_total` - Total reconciliations
- `eventflow_reconcile_errors_total` - Reconciliation errors
- `eventflow_reconcile_duration_seconds` - Reconciliation time
- `eventflow_functions_phase` - Functions by phase (Pending/Running/Failed)

### Logging

**Structured Logging**:
```go
log.Info("Reconciling Function",
    "name", function.Name,
    "namespace", function.Namespace,
    "user_id", function.Spec.UserID,
    "phase", function.Status.Phase)
```

**Log Aggregation**:
- API and Operator logs to stdout
- Kubernetes captures to cluster logging (ELK/Loki)
- User function logs accessible via API

### Tracing

**Distributed Tracing** (Future):
- OpenTelemetry instrumentation
- Trace requests from Web UI → API → K8s API
- Correlate logs across components

## Future Architecture Enhancements

1. **Multi-Region Support**:
   - Replicate PostgreSQL across regions
   - Federated Kubernetes clusters
   - Global load balancer

2. **Event-Driven Functions**:
   - NATS/Kafka integration
   - Trigger functions from events
   - Async invocation with callbacks

3. **Function Builds**:
   - Git integration for source code
   - Automatic containerization (Buildpacks)
   - Internal container registry

4. **Advanced Isolation**:
   - NetworkPolicies between tenants
   - Pod Security Standards enforcement
   - Runtime security (Falco)

5. **Cost Management**:
   - Resource usage tracking per tenant
   - Cost allocation and chargeback
   - Budget alerts and limits
