# EventFlow Infrastructure - Development Environment

This directory contains the infrastructure setup for running EventFlow on K3s (lightweight Kubernetes).

## Prerequisites

- Linux machine (Ubuntu/Debian recommended)
- Minimum 2GB RAM
- Docker (for building images)
- [Task](https://taskfile.dev) - Modern task runner

## Quick Start

### Using Task (Recommended)

```bash
# Install Task first: https://taskfile.dev/installation/
# Complete setup in one command
task setup

# Or step by step:
task k3s:install    # Install K3s
task build          # Build all images
task deploy         # Deploy EventFlow

# Access the application
# Web UI: http://localhost:30300
# API: http://localhost:30080
```

### Alternative: Using Shell Scripts

```bash
# 1. Install K3s
./install-k3s.sh

# 2. Build and load images
./build-images.sh

# 3. Deploy EventFlow
./deploy.sh

# 4. Access the application
# Web UI: http://localhost:30080
# API: http://localhost:30080

# 5. Clean up
./cleanup.sh
```

## Available Tools

### Taskfile.yml (Primary)
Modern task runner with better syntax and built-in parallelization.

```bash
task --list        # Show all available tasks
task setup         # Complete setup
task rebuild:api   # Quick rebuild API
task logs:api      # Follow API logs
task status        # Show cluster status
task test:all      # Run all tests
```

See [TASK-REFERENCE.md](./TASK-REFERENCE.md) for complete command reference.

### Shell Scripts (Standalone)
Scripts for manual execution or CI/CD pipelines.

- `install-k3s.sh` - Installs K3s cluster
- `build-images.sh` - Builds Docker images
- `deploy.sh` - Deploys EventFlow
- `cleanup.sh` - Removes EventFlow
- `uninstall-k3s.sh` - Removes K3s
- `dev.sh` - All-in-one development tool
- `monitor.sh` - Monitor logs for errors

## K3s Configuration

K3s is configured with:
- Traefik disabled (using our own ingress)
- Local path provisioner enabled
- Metrics server enabled
- Default storage class

## Accessing K3s

K3s kubeconfig is located at `/etc/rancher/k3s/k3s.yaml`

To use kubectl:
```bash
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
kubectl get nodes
```

Or use k3s kubectl directly:
```bash
sudo k3s kubectl get nodes
```
