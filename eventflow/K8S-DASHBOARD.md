# Kubernetes Dashboard Access

## 🎯 Dashboard URL
http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/

## 🔑 Access Token Location
```bash
cat /home/ironik/GolandProjects/webapp/eventflow/k8s-dashboard-token.txt
```

## 📋 How to Access

1. Make sure kubectl proxy is running:
   ```bash
   kubectl proxy --port=8001
   ```

2. Open browser to the URL above

3. Select "Token" authentication method

4. Paste the token from `k8s-dashboard-token.txt`

5. Click "Sign In"

## 🔄 Restart Proxy if Needed
```bash
# Kill existing proxy
pkill -f "kubectl proxy"

# Start new proxy
kubectl proxy --port=8001 &
```

## 🎨 What You Can See
- All pods, deployments, services across namespaces
- EventFlow functions running in the cluster
- Real-time metrics and logs
- Resource usage graphs
- Events and troubleshooting info
