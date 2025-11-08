# EventFlow Deployment Guide

## Table of Contents

- [Local Development (kind)](#local-development-kind)
- [Production Kubernetes](#production-kubernetes)
- [Configuration](#configuration)
- [Monitoring](#monitoring)
- [Backup and Recovery](#backup-and-recovery)
- [Troubleshooting](#troubleshooting)

## Local Development (kind)

### Prerequisites

```bash
# macOS
brew install docker kind kubectl go node

# Linux
# Install Docker: https://docs.docker.com/engine/install/
# Install kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installation
# Install kubectl: https://kubernetes.io/docs/tasks/tools/
```

### Step-by-Step Deployment

#### 1. Create kind Cluster

```bash
cd eventflow

# Create cluster with custom configuration
kind create cluster --name eventflow --config kind-config.yaml

# Verify cluster
kubectl cluster-info --context kind-eventflow
kubectl get nodes
```

#### 2. Deploy PostgreSQL

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Deploy PostgreSQL
kubectl apply -f k8s/postgres.yaml

# Wait for PostgreSQL to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n eventflow --timeout=120s

# Initialize database
kubectl exec -n eventflow deployment/postgres -- \
  psql -U eventflow -d eventflow -c "$(cat scripts/init-db.sql)"
```

#### 3. Deploy the Operator

```bash
# Apply CRD
kubectl apply -f k8s/crd-function.yaml

# Build operator image
cd operator
make docker-build IMG=eventflow-operator:latest

# Load into kind
kind load docker-image eventflow-operator:latest --name eventflow

# Deploy operator
kubectl apply -f ../k8s/operator.yaml

# Verify operator is running
kubectl get pods -n eventflow -l app=operator
kubectl logs -n eventflow -l app=operator -f
```

#### 4. Deploy API Server

```bash
# Build API image
cd ../api
docker build -t eventflow-api:latest .

# Load into kind
kind load docker-image eventflow-api:latest --name eventflow

# Apply RBAC
kubectl apply -f ../k8s/api-cluster-rbac.yaml

# Deploy API
kubectl apply -f ../k8s/deployment.yaml

# Verify API is running
kubectl get pods -n eventflow -l app=eventflow-api
kubectl logs -n eventflow -l app=eventflow-api -f
```

#### 5. Deploy Web Dashboard

```bash
# Build web image
cd ../web
docker build -t eventflow-web:latest .

# Load into kind
kind load docker-image eventflow-web:latest --name eventflow

# Web is already in deployment.yaml
kubectl get pods -n eventflow -l app=eventflow-web
```

#### 6. Access the Platform

```bash
# Option 1: Port forward
kubectl port-forward -n eventflow svc/eventflow-web 3000:80 &
kubectl port-forward -n eventflow svc/eventflow-api 8080:80 &

open http://localhost:3000

# Option 2: NodePort (kind-specific)
# API is exposed on port 30080
# Web is exposed on port 30081
open http://localhost:30081
```

### Quick Deploy Script

Create a `deploy-local.sh` script:

```bash
#!/bin/bash
set -e

echo "🚀 Deploying EventFlow to kind cluster..."

# Create cluster
echo "📦 Creating kind cluster..."
kind create cluster --name eventflow --config kind-config.yaml

# Deploy PostgreSQL
echo "🐘 Deploying PostgreSQL..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres.yaml
kubectl wait --for=condition=ready pod -l app=postgres -n eventflow --timeout=120s

# Initialize database
echo "💾 Initializing database..."
kubectl exec -n eventflow deployment/postgres -- \
  psql -U eventflow -d eventflow -c "$(cat scripts/init-db.sql)"

# Deploy operator
echo "⚙️  Deploying operator..."
kubectl apply -f k8s/crd-function.yaml
cd operator && make docker-build IMG=eventflow-operator:latest
kind load docker-image eventflow-operator:latest --name eventflow
kubectl apply -f ../k8s/operator.yaml
cd ..

# Deploy API
echo "🔌 Deploying API..."
cd api && docker build -t eventflow-api:latest .
kind load docker-image eventflow-api:latest --name eventflow
cd ..
kubectl apply -f k8s/api-cluster-rbac.yaml

# Deploy web
echo "🎨 Deploying web dashboard..."
cd web && docker build -t eventflow-web:latest .
kind load docker-image eventflow-web:latest --name eventflow
cd ..

# Deploy services
kubectl apply -f k8s/deployment.yaml

# Wait for pods
echo "⏳ Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=eventflow-api -n eventflow --timeout=120s
kubectl wait --for=condition=ready pod -l app=eventflow-web -n eventflow --timeout=120s

echo "✅ Deployment complete!"
echo ""
echo "🌐 Access the dashboard:"
echo "   http://localhost:30081 (NodePort)"
echo ""
echo "   Or use port-forward:"
echo "   kubectl port-forward -n eventflow svc/eventflow-web 3000:80"
echo "   http://localhost:3000"
echo ""
echo "📡 API available at:"
echo "   http://localhost:30080 (NodePort)"
```

---

## Production Kubernetes

### Prerequisites

- Kubernetes 1.28+ cluster
- `kubectl` configured with admin access
- Persistent storage (StorageClass)
- Load balancer (for Ingress)
- TLS certificates (Let's Encrypt or cert-manager)

### Production Checklist

- [ ] Use managed PostgreSQL (AWS RDS, GCP Cloud SQL, Azure Database)
- [ ] Configure TLS/SSL for API and Web
- [ ] Set up Ingress with proper domain names
- [ ] Configure RBAC with least privilege
- [ ] Enable Pod Security Standards
- [ ] Set up NetworkPolicies
- [ ] Configure resource requests and limits
- [ ] Set up monitoring (Prometheus + Grafana)
- [ ] Configure log aggregation (ELK, Loki)
- [ ] Set up backup and disaster recovery
- [ ] Configure high availability (multiple replicas)
- [ ] Use Secrets management (Sealed Secrets, External Secrets)

### 1. Namespace and RBAC

```bash
# Create dedicated namespace
kubectl create namespace eventflow-prod

# Apply RBAC
kubectl apply -f k8s/api-cluster-rbac.yaml
kubectl apply -f k8s/operator.yaml
```

### 2. External PostgreSQL

Update `k8s/secrets.yaml` with production database credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: eventflow-secrets
  namespace: eventflow-prod
type: Opaque
stringData:
  JWT_SECRET: <generate-strong-secret>
  DB_HOST: <rds-endpoint>.rds.amazonaws.com
  DB_PORT: "5432"
  DB_USER: eventflow
  DB_PASSWORD: <secure-password>
  DB_NAME: eventflow
  DB_SSLMODE: require
```

Apply secrets:
```bash
kubectl apply -f k8s/secrets.yaml -n eventflow-prod
```

### 3. Deploy Operator

```bash
# Build and push operator image
cd operator
make docker-build IMG=your-registry.com/eventflow-operator:v1.0.0
docker push your-registry.com/eventflow-operator:v1.0.0

# Update operator deployment with correct image
kubectl set image deployment/operator-controller-manager \
  manager=your-registry.com/eventflow-operator:v1.0.0 \
  -n eventflow-prod
```

### 4. Deploy API with Resource Limits

Create `k8s/production/api-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: eventflow-api
  namespace: eventflow-prod
spec:
  replicas: 3
  selector:
    matchLabels:
      app: eventflow-api
  template:
    metadata:
      labels:
        app: eventflow-api
    spec:
      serviceAccountName: eventflow-api
      containers:
      - name: api
        image: your-registry.com/eventflow-api:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: eventflow-secrets
              key: JWT_SECRET
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: eventflow-secrets
              key: DB_HOST
        - name: DB_PORT
          valueFrom:
            secretKeyRef:
              name: eventflow-secrets
              key: DB_PORT
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: eventflow-secrets
              key: DB_USER
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: eventflow-secrets
              key: DB_PASSWORD
        - name: DB_NAME
          valueFrom:
            secretKeyRef:
              name: eventflow-secrets
              key: DB_NAME
        - name: DB_SSLMODE
          valueFrom:
            secretKeyRef:
              name: eventflow-secrets
              key: DB_SSLMODE
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 5. Set Up Ingress with TLS

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: eventflow-ingress
  namespace: eventflow-prod
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - eventflow.yourdomain.com
    - api.eventflow.yourdomain.com
    secretName: eventflow-tls
  rules:
  - host: eventflow.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: eventflow-web
            port:
              number: 80
  - host: api.eventflow.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: eventflow-api
            port:
              number: 80
```

### 6. Configure NetworkPolicies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-network-policy
  namespace: eventflow-prod
spec:
  podSelector:
    matchLabels:
      app: eventflow-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}
      podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443  # Kubernetes API
```

### 7. Set Up Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: eventflow-api-hpa
  namespace: eventflow-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: eventflow-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## Configuration

### Environment Variables

#### API Server

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `JWT_SECRET` | Secret key for JWT signing | - | Yes |
| `DB_HOST` | PostgreSQL host | `localhost` | Yes |
| `DB_PORT` | PostgreSQL port | `5432` | Yes |
| `DB_USER` | Database user | `eventflow` | Yes |
| `DB_PASSWORD` | Database password | - | Yes |
| `DB_NAME` | Database name | `eventflow` | Yes |
| `DB_SSLMODE` | SSL mode (`disable`/`require`) | `disable` | No |
| `PORT` | API server port | `8080` | No |
| `LOG_LEVEL` | Log level (`debug`/`info`/`warn`/`error`) | `info` | No |

#### Operator

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `METRICS_BIND_ADDRESS` | Metrics server address | `:8080` | No |
| `HEALTH_PROBE_BIND_ADDRESS` | Health probe address | `:8081` | No |
| `LEADER_ELECT` | Enable leader election | `true` | No |

### Resource Quotas

Adjust per-tenant quotas in `api/internal/k8s/client.go`:

```go
func (c *Client) createResourceQuota(ctx context.Context, namespace string) error {
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-quota",
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceRequestsCPU:              resource.MustParse("10"),    // Adjust for production
				corev1.ResourceRequestsMemory:           resource.MustParse("20Gi"),  // Adjust for production
				corev1.ResourceLimitsCPU:                resource.MustParse("20"),    // Adjust for production
				corev1.ResourceLimitsMemory:             resource.MustParse("40Gi"),  // Adjust for production
				corev1.ResourcePods:                     resource.MustParse("50"),    // Adjust for production
				corev1.ResourcePersistentVolumeClaims:   resource.MustParse("10"),    // Adjust for production
			},
		},
	}
	// ...
}
```

### Function Resource Defaults

Adjust default function resources in `operator/internal/controller/function_controller.go`:

```go
container.Resources = corev1.ResourceRequirements{
	Requests: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("100m"),  // Adjust
		corev1.ResourceMemory: resource.MustParse("128Mi"), // Adjust
	},
	Limits: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("500m"),  // Adjust
		corev1.ResourceMemory: resource.MustParse("512Mi"), // Adjust
	},
}
```

---

## Monitoring

### Prometheus Setup

```yaml
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: eventflow-api
  namespace: eventflow-prod
spec:
  selector:
    matchLabels:
      app: eventflow-api
  endpoints:
  - port: metrics
    path: /metrics
    interval: 30s
```

### Grafana Dashboard

Import the EventFlow dashboard (create `grafana-dashboard.json`):

**Key Metrics**:
- API request rate
- API latency (p50, p95, p99)
- Function count by status
- Namespace resource usage
- Operator reconciliation rate
- Database connection pool usage

### Alerts

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: eventflow-alerts
  namespace: eventflow-prod
spec:
  groups:
  - name: eventflow
    rules:
    - alert: HighAPIErrorRate
      expr: rate(eventflow_api_requests_total{status=~"5.."}[5m]) > 0.05
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High API error rate"
        description: "API error rate is {{ $value }} req/s"
    
    - alert: DatabaseDown
      expr: up{job="postgres"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "PostgreSQL is down"
    
    - alert: OperatorNotReconciling
      expr: rate(eventflow_reconcile_total[5m]) == 0
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "Operator not reconciling"
```

---

## Backup and Recovery

### Database Backup

```bash
# Daily backup cron job
kubectl create -f - <<EOF
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: eventflow-prod
spec:
  schedule: "0 2 * * *"  # 2 AM daily
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: eventflow-secrets
                  key: DB_PASSWORD
            command:
            - sh
            - -c
            - |
              pg_dump -h $DB_HOST -U $DB_USER $DB_NAME | \
              gzip > /backup/eventflow-$(date +%Y%m%d).sql.gz
              # Upload to S3/GCS
            volumeMounts:
            - name: backup
              mountPath: /backup
          volumes:
          - name: backup
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
EOF
```

### Disaster Recovery

1. **Database Restore**:
```bash
kubectl exec -it postgres-0 -n eventflow-prod -- \
  psql -U eventflow -d eventflow < backup.sql
```

2. **Redeploy Platform**:
```bash
kubectl apply -f k8s/production/
```

3. **Verify Functions**:
```bash
kubectl get functions --all-namespaces
```

---

## Troubleshooting

### Common Issues

#### API Pod Not Starting

```bash
# Check logs
kubectl logs -n eventflow-prod -l app=eventflow-api

# Check events
kubectl describe pod -n eventflow-prod -l app=eventflow-api

# Common issues:
# - Database connection failed
# - Missing secrets
# - Resource limits too low
```

#### Operator Not Reconciling

```bash
# Check operator logs
kubectl logs -n eventflow-prod -l app=operator -f

# Check RBAC
kubectl auth can-i create deployments --as=system:serviceaccount:eventflow-prod:operator-controller-manager

# Check CRD
kubectl get crd functions.eventflow.eventflow.io
```

#### Functions Stuck in Pending

```bash
# Check function status
kubectl get functions -n tenant-alice

# Check deployment
kubectl get deployments -n tenant-alice

# Check pods
kubectl get pods -n tenant-alice

# Check events
kubectl describe function my-function -n tenant-alice

# Common issues:
# - Resource quota exceeded
# - Image pull failed
# - Node resources exhausted
```

#### Database Connection Issues

```bash
# Test connection from API pod
kubectl exec -it deployment/eventflow-api -n eventflow-prod -- \
  psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# Check connection pool
# Look for "too many connections" errors in logs
```

### Debug Mode

Enable debug logging:

```yaml
# API
env:
- name: LOG_LEVEL
  value: "debug"

# Operator
args:
- --zap-log-level=debug
```

### Performance Tuning

1. **Increase API replicas**:
```bash
kubectl scale deployment eventflow-api -n eventflow-prod --replicas=5
```

2. **Increase database connections**:
```go
// api/internal/database/db.go
config.MaxConns = 50
config.MinConns = 10
```

3. **Tune operator workers**:
```go
// operator/cmd/main.go
MaxConcurrentReconciles: 10
```

---

## Upgrade Strategy

### Rolling Update

```bash
# Build new version
docker build -t your-registry.com/eventflow-api:v1.1.0 api/
docker push your-registry.com/eventflow-api:v1.1.0

# Update deployment
kubectl set image deployment/eventflow-api \
  api=your-registry.com/eventflow-api:v1.1.0 \
  -n eventflow-prod

# Monitor rollout
kubectl rollout status deployment/eventflow-api -n eventflow-prod

# Rollback if needed
kubectl rollout undo deployment/eventflow-api -n eventflow-prod
```

### Database Migration

```bash
# 1. Backup database
# 2. Test migration on staging
# 3. Apply migration
kubectl exec -n eventflow-prod deployment/postgres -- \
  psql -U eventflow -d eventflow -f /path/to/migration.sql
# 4. Deploy new API version
# 5. Verify
```
