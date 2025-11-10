# EventFlow K3s Quick Start

Get EventFlow running on K3s in minutes!

## One-Command Setup

```bash
cd infra-dev
./dev.sh setup
```

This will:
1. ✅ Install K3s
2. ✅ Build all Docker images
3. ✅ Deploy EventFlow to K3s
4. ✅ Wait for everything to be ready

## Step-by-Step Setup

### 1. Install K3s

```bash
cd infra-dev
./install-k3s.sh
```

**What it does:**
- Installs K3s as a systemd service
- Configures kubeconfig at `~/.kube/config`
- Disables Traefik (we use NodePort)
- Sets up local storage

**Time:** ~1 minute

### 2. Build Images

```bash
./build-images.sh
```

**What it does:**
- Builds `eventflow-api:latest`
- Builds `eventflow-operator:latest`
- Builds `eventflow-web:latest`
- Imports all images into K3s containerd

**Time:** ~3-5 minutes (first build), ~30 seconds (cached)

### 3. Deploy EventFlow

```bash
./deploy.sh
```

**What it does:**
- Creates `eventflow` namespace
- Deploys PostgreSQL database
- Deploys EventFlow Operator
- Deploys EventFlow API
- Deploys EventFlow Web UI
- Sets up all RBAC permissions

**Time:** ~2 minutes

### 4. Access EventFlow

```bash
# Web UI
open http://localhost:30080

# Check status
./dev.sh status
```

## Daily Development Workflow

### Quick Rebuild After Code Changes

**Rebuild API only:**
```bash
./dev.sh rebuild-api
```

**Rebuild Operator only:**
```bash
./dev.sh rebuild-operator
```

**Rebuild Web UI only:**
```bash
./dev.sh rebuild-web
```

**Rebuild everything:**
```bash
./dev.sh rebuild
```

### View Logs

```bash
# API logs
./dev.sh logs api

# Operator logs
./dev.sh logs operator

# Web UI logs
./dev.sh logs web

# PostgreSQL logs
./dev.sh logs postgres
```

### Monitor for Errors

```bash
# Monitor API for errors
./monitor.sh api

# Monitor all components
./monitor.sh all
```

### Check Status

```bash
./dev.sh status
```

Shows:
- All pods in eventflow namespace
- Tenant namespaces
- All deployed functions

### Database Shell

```bash
./dev.sh shell postgres
```

Connects to PostgreSQL with psql.

### API Container Shell

```bash
./dev.sh shell api
```

Opens shell in API container.

## Common Tasks

### Create a Test Function

```bash
# Get JWT token for alice
export TOKEN="your-jwt-token"

# Create function
curl -X POST http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hello",
    "image": "hashicorp/http-echo:latest",
    "replicas": 1,
    "command": ["-text=Hello from EventFlow!"]
  }'

# List functions
curl http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN"
```

### Check Function Status

```bash
# Get all Functions CRs
kubectl get functions.eventflow.eventflow.io -A

# Describe specific function
kubectl describe function hello -n tenant-alice

# Check function pods
kubectl get pods -n tenant-alice
```

### Reset Everything

```bash
# Clean and redeploy
./dev.sh reset

# Or complete wipe
./cleanup.sh
./uninstall-k3s.sh
./dev.sh setup
```

## Troubleshooting

### API Not Starting

```bash
# Check logs
kubectl logs -n eventflow -l app=eventflow-api

# Check database connection
kubectl get pods -n eventflow -l app=postgres
```

### Images Not Found

```bash
# Verify images are imported
sudo k3s ctr images ls | grep eventflow

# Rebuild if missing
./build-images.sh
```

### Port 30080 Not Accessible

```bash
# Check service
kubectl get svc eventflow-api -n eventflow

# Check pods are ready
kubectl get pods -n eventflow

# Use port-forward as alternative
kubectl port-forward -n eventflow svc/eventflow-api 8080:80
```

See [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) for more help.

## Development Tips

### Fast Iteration on API Changes

```bash
# 1. Make changes to API code
vim ../eventflow/api/internal/handlers/functions.go

# 2. Quick rebuild and deploy
./dev.sh rebuild-api

# 3. Watch logs
./dev.sh logs api
```

### Testing Multi-Tenancy

```bash
# Get tokens for different users
export TOKEN_ALICE="..."
export TOKEN_BOB="..."

# Create function as alice
curl -X POST http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN_ALICE" \
  -d '{"name":"alice-func","image":"nginx:alpine",...}'

# Create function as bob
curl -X POST http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN_BOB" \
  -d '{"name":"bob-func","image":"nginx:alpine",...}'

# Verify isolation
kubectl get functions -n tenant-alice
kubectl get functions -n tenant-bob
```

### Debugging Operator

```bash
# Watch operator logs
kubectl logs -f -n eventflow -l app=operator-controller-manager

# Check reconciliation events
kubectl get events -n tenant-alice --watch

# Verify RBAC
kubectl auth can-i create deployments --as=system:serviceaccount:eventflow:operator-controller-manager -n tenant-alice
```

## Next Steps

- Read [K3S-VS-KIND.md](./K3S-VS-KIND.md) to understand the differences
- Check [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) if you encounter issues
- Explore the main [README.md](../README.md) for EventFlow architecture

## Cleanup

### Remove EventFlow (Keep K3s)

```bash
./cleanup.sh
```

### Complete Uninstall

```bash
./uninstall-k3s.sh
```

---

**Happy Coding! 🚀**
