# EventFlow - Quick Setup Guide

## Option 1: Docker Compose (Simplest)

```bash
cd eventflow

# Start everything
make run

# Get a dev token
curl -X POST http://localhost:8080/auth/token

# Open dashboard and login with the token
open http://localhost:3000

# Test creating a function
make test-function

# View logs
make logs

# Stop everything
make stop
```

## Option 2: Local Development

### Backend

```bash
cd api

# Download dependencies
go mod download

# Set environment variables
export NAMESPACE=default
export JWT_SECRET=dev-secret-change-in-production

# Run (requires kubectl access to a cluster)
go run main.go
```

### Frontend

```bash
cd web

# Install dependencies
npm install

# Run dev server (proxies API to localhost:8080)
npm run dev

# Open browser
open http://localhost:3000
```

## Option 3: Kubernetes (kind)

```bash
# Create kind cluster and deploy everything
make kind-setup

# Port forward to access
kubectl port-forward -n eventflow svc/eventflow-api 8080:80

# Get token
curl -X POST http://localhost:8080/auth/token

# Test the API
make test-function

# Check status
make status

# Clean up
make kind-clean
```

## First Steps

1. **Login**: Use the "Get Dev Token & Login" button
2. **Create Function**: Click "+ Create Function"
   - Name: `hello-world`
   - Image: `nginx:alpine`
   - Replicas: `1`
3. **View Function**: Click on the function card
4. **Invoke**: Click the "Invoke" button
5. **View Logs**: Logs appear at the bottom
6. **Delete**: Click the trash icon

## Example Functions to Try

### Nginx Web Server
```json
{
  "name": "web-server",
  "image": "nginx:alpine",
  "replicas": 2
}
```

### Redis Cache
```json
{
  "name": "redis-cache",
  "image": "redis:7-alpine",
  "replicas": 1
}
```

### Custom Container
```json
{
  "name": "my-app",
  "image": "your-registry/your-app:latest",
  "replicas": 3,
  "env": {
    "PORT": "8080",
    "ENV": "production"
  },
  "command": ["./app", "start"]
}
```

## Troubleshooting

### "Cannot connect to Kubernetes cluster"
- Make sure you have a Kubernetes cluster running
- Check `kubectl cluster-info`
- For Docker Compose: mount your `~/.kube/config`

### "Failed to get token"
- Check API is running: `curl http://localhost:8080/healthz`
- Check logs: `docker-compose logs api`

### Frontend shows connection error
- Verify API is running on port 8080
- Check browser console for errors
- Try accessing API directly: `curl http://localhost:8080/healthz`

## Next Steps

- Read the full [README.md](README.md)
- Check the [API Documentation](README.md#api-documentation)
- Explore the Kubernetes manifests in `k8s/`
- Customize the JWT secret in production
- Add ingress for external access
- Configure HPA for autoscaling
