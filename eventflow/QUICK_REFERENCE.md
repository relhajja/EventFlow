# EventFlow - Quick Reference

## 🚀 Quick Commands

```bash
# Start everything (Docker Compose)
make run

# Stop everything
make stop

# View logs
make logs

# Test function deployment
make test-function

# Deploy to kind cluster
make kind-setup

# Clean everything
make clean
```

## 📡 API Endpoints

**Base URL**: `http://localhost:8080`

### Authentication
```bash
# Get dev token
curl -X POST http://localhost:8080/auth/token

# Response: {"token":"eyJhbGc..."}
```

### Functions
```bash
# List all functions
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/functions

# Create function
curl -X POST http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-func",
    "image": "nginx:alpine",
    "replicas": 1
  }'

# Get function details
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/functions/my-func

# Invoke function
curl -X POST http://localhost:8080/v1/functions/my-func:invoke \
  -H "Authorization: Bearer $TOKEN"

# Get logs
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/functions/my-func/logs

# Delete function
curl -X DELETE http://localhost:8080/v1/functions/my-func \
  -H "Authorization: Bearer $TOKEN"
```

## 🌐 URLs

- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **Metrics**: http://localhost:9090/metrics
- **Health**: http://localhost:8080/healthz
- **Ready**: http://localhost:8080/readyz

## 🐳 Docker Commands

```bash
# Build images
docker-compose build

# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Clean everything
docker-compose down -v --rmi all
```

## ☸️ Kubernetes Commands

```bash
# Apply manifests
kubectl apply -f k8s/

# Check status
kubectl get pods -n eventflow
kubectl get svc -n eventflow

# View logs
kubectl logs -n eventflow -l app=eventflow-api -f

# Port forward
kubectl port-forward -n eventflow svc/eventflow-api 8080:80

# Delete everything
kubectl delete namespace eventflow
```

## 📂 Project Structure

```
eventflow/
├── api/                    # Go backend
│   ├── main.go
│   ├── internal/
│   └── Dockerfile
├── web/                    # React frontend
│   ├── src/
│   ├── package.json
│   └── Dockerfile
├── k8s/                    # Kubernetes manifests
│   ├── namespace.yaml
│   ├── rbac.yaml
│   ├── deployment.yaml
│   └── hpa.yaml
├── scripts/
│   └── test-function.sh
├── Makefile
├── docker-compose.yaml
└── README.md
```

## 🔑 Environment Variables

### API
- `PORT=8080` - HTTP server port
- `NAMESPACE=default` - Kubernetes namespace
- `JWT_SECRET=secret` - JWT signing secret
- `LOG_LEVEL=info` - Log level

### Web
- `VITE_API_URL` - API base URL (optional)

## 🧪 Example Functions

### Simple Web Server
```json
{
  "name": "nginx",
  "image": "nginx:alpine",
  "replicas": 1
}
```

### Redis Cache
```json
{
  "name": "redis",
  "image": "redis:7-alpine",
  "replicas": 1,
  "env": {
    "REDIS_PASSWORD": "mysecret"
  }
}
```

### Custom App
```json
{
  "name": "my-app",
  "image": "myregistry/myapp:v1.0",
  "replicas": 3,
  "command": ["./app", "start"],
  "env": {
    "PORT": "8080",
    "DB_HOST": "postgres"
  }
}
```

## 🐛 Troubleshooting

### API won't start
```bash
# Check if running
curl http://localhost:8080/healthz

# View logs
docker-compose logs api

# Check kubectl access
kubectl cluster-info
```

### Frontend shows errors
```bash
# Check API is accessible
curl http://localhost:8080/v1/functions

# Check browser console
# Verify CORS settings
```

### Function won't deploy
```bash
# Check Kubernetes access
kubectl get nodes

# Check RBAC permissions
kubectl get sa,role,rolebinding -n eventflow

# View API logs
kubectl logs -n eventflow -l app=eventflow-api
```

## 📊 Monitoring

### Prometheus Metrics
```bash
# View all metrics
curl http://localhost:9090/metrics

# Function invocations
curl http://localhost:9090/metrics | grep eventflow_function_invocations

# Active functions
curl http://localhost:9090/metrics | grep eventflow_active_functions
```

### Health Checks
```bash
# Health check
curl http://localhost:8080/healthz

# Readiness check
curl http://localhost:8080/readyz
```

## 🔒 Security Checklist

- [ ] Change JWT secret in production
- [ ] Use OIDC instead of dev tokens
- [ ] Configure network policies
- [ ] Set resource limits
- [ ] Enable RBAC auditing
- [ ] Use private container registry
- [ ] Enable TLS/HTTPS
- [ ] Configure ingress authentication

## 📚 Documentation

- **Full Docs**: [README.md](README.md)
- **Setup Guide**: [SETUP.md](SETUP.md)
- **Project Summary**: [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)
- **This File**: QUICK_REFERENCE.md

---

**Need Help?** Check the full [README.md](README.md) for detailed documentation.
