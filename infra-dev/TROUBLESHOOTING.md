# K3s Troubleshooting Guide

## Common Issues and Solutions

### 1. K3s Installation Fails

**Symptom:** Installation script fails or hangs

**Solutions:**
```bash
# Check if port 6443 is already in use
sudo netstat -tlnp | grep 6443

# Check if another Kubernetes is running
ps aux | grep kube

# Clean up any existing K3s
sudo /usr/local/bin/k3s-uninstall.sh

# Try installation again
./install-k3s.sh
```

### 2. Cannot Access Cluster

**Symptom:** `kubectl` commands fail with "connection refused"

**Solutions:**
```bash
# Check if K3s is running
sudo systemctl status k3s

# Restart K3s
sudo systemctl restart k3s

# Setup kubeconfig
export KUBECONFIG=$HOME/.kube/config
sudo cp /etc/rancher/k3s/k3s.yaml $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# Verify
kubectl get nodes
```

### 3. Images Not Found

**Symptom:** Pods show `ImagePullBackOff` or `ErrImagePull`

**Solutions:**
```bash
# Check if images are imported
sudo k3s ctr images ls | grep eventflow

# Rebuild and import images
./build-images.sh

# Check pod events
kubectl describe pod <pod-name> -n eventflow
```

### 4. Pods in CrashLoopBackOff

**Symptom:** Pods keep restarting

**Solutions:**
```bash
# Check logs
kubectl logs <pod-name> -n eventflow --previous

# Check resource limits
kubectl describe pod <pod-name> -n eventflow

# Check if database is ready
kubectl get pods -n eventflow -l app=postgres

# Restart deployment
kubectl rollout restart deployment/<deployment-name> -n eventflow
```

### 5. "Too Many Open Files" Error

**Symptom:** API or Operator shows file descriptor errors

**Solutions:**
```bash
# Check current limits
ulimit -n

# Increase limits temporarily
ulimit -n 65536

# Increase limits permanently (edit /etc/security/limits.conf)
sudo tee -a /etc/security/limits.conf <<EOF
*               soft    nofile          65536
*               hard    nofile          65536
EOF

# Rebuild with lower connection limits
# (Already configured in current version)
./dev.sh rebuild-api
```

### 6. Database Connection Errors

**Symptom:** API cannot connect to PostgreSQL

**Solutions:**
```bash
# Check if PostgreSQL is running
kubectl get pods -n eventflow -l app=postgres

# Check PostgreSQL logs
kubectl logs -n eventflow -l app=postgres

# Test connection manually
kubectl exec -it -n eventflow <api-pod-name> -- /bin/sh
# Then try: nc -zv postgres 5432

# Restart PostgreSQL
kubectl rollout restart deployment/postgres -n eventflow
```

### 7. NodePort Service Not Accessible

**Symptom:** Cannot access http://localhost:30080

**Solutions:**
```bash
# Check if service is created
kubectl get svc -n eventflow

# Check if pods are ready
kubectl get pods -n eventflow

# Check firewall rules
sudo iptables -L -n | grep 30080

# Test service directly
curl http://localhost:30080/healthz

# Port forward as alternative
kubectl port-forward -n eventflow svc/eventflow-api 8080:80
```

### 8. Operator Not Creating Resources

**Symptom:** Function CRs created but no Deployments

**Solutions:**
```bash
# Check operator logs
kubectl logs -n eventflow -l app=operator-controller-manager

# Check if CRD is installed
kubectl get crd functions.eventflow.eventflow.io

# Check operator RBAC
kubectl get clusterrole operator-manager-role
kubectl get clusterrolebinding operator-manager-rolebinding

# Restart operator
kubectl rollout restart deployment/operator-controller-manager -n eventflow
```

### 9. Namespace Stuck in Terminating

**Symptom:** Namespace won't delete

**Solutions:**
```bash
# Check for finalizers
kubectl get namespace <namespace> -o json | jq '.spec.finalizers'

# Remove finalizers
kubectl patch namespace <namespace> -p '{"spec":{"finalizers":null}}' --type=merge

# Or use force delete
kubectl delete namespace <namespace> --grace-period=0 --force
```

### 10. Disk Space Issues

**Symptom:** Cannot pull images or create pods due to disk space

**Solutions:**
```bash
# Check disk usage
df -h

# Clean up K3s images
sudo k3s crictl rmi --prune

# Clean up Docker images (if Docker is installed)
docker system prune -a -f

# Check K3s data usage
sudo du -sh /var/lib/rancher/k3s
```

## Diagnostic Commands

### Check Cluster Health
```bash
# Node status
kubectl get nodes -o wide

# All pods
kubectl get pods -A

# System components
kubectl get pods -n kube-system

# Events
kubectl get events -n eventflow --sort-by='.lastTimestamp'
```

### Check Resources
```bash
# CPU and memory usage
kubectl top nodes
kubectl top pods -n eventflow

# Storage
kubectl get pv
kubectl get pvc -A
```

### Check Networking
```bash
# Services
kubectl get svc -A

# Endpoints
kubectl get endpoints -n eventflow

# Test internal DNS
kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup postgres.eventflow.svc.cluster.local
```

### Logs
```bash
# API logs
kubectl logs -n eventflow -l app=eventflow-api --tail=100

# Operator logs
kubectl logs -n eventflow -l app=operator-controller-manager --tail=100

# All logs
kubectl logs -n eventflow --all-containers=true --tail=100

# Follow logs
kubectl logs -f -n eventflow <pod-name>
```

## Performance Tuning

### Reduce Memory Usage
```bash
# Edit deployment.yaml and reduce resources:
resources:
  requests:
    memory: "64Mi"
    cpu: "50m"
  limits:
    memory: "256Mi"
    cpu: "200m"
```

### Increase Connection Limits
```bash
# Already configured in current K3s client:
config.QPS = 10
config.Burst = 15
config.Timeout = 10

# Increase if needed in api/internal/k8s/client.go
```

## Complete Reset

If all else fails, completely reset the environment:

```bash
# 1. Cleanup EventFlow
./cleanup.sh

# 2. Uninstall K3s
./uninstall-k3s.sh

# 3. Clean up any remaining files
sudo rm -rf /var/lib/rancher/k3s
sudo rm -rf /etc/rancher/k3s
rm -rf $HOME/.kube

# 4. Reinstall everything
./install-k3s.sh
./build-images.sh
./deploy.sh
```

## Getting Help

If issues persist:

1. Check K3s logs: `sudo journalctl -u k3s -f`
2. Check dmesg for system errors: `sudo dmesg | tail -100`
3. Check K3s GitHub issues: https://github.com/k3s-io/k3s/issues
4. Check EventFlow logs with: `./dev.sh logs api`
