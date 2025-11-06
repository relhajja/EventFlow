## Kubebuilder Operator

### Build and Deploy (Kubebuilder)
```bash
cd /home/ironik/GolandProjects/webapp/eventflow/operator
make docker-build IMG=eventflow-operator:latest
kind load docker-image eventflow-operator:latest --name eventflow
make deploy
```

### Quick Redeploy (after code changes)
```bash
cd /home/ironik/GolandProjects/webapp/eventflow/operator && \
  make docker-build IMG=eventflow-operator:latest && \
  kind load docker-image eventflow-operator:latest --name eventflow && \
  kubectl rollout restart deployment/operator-controller-manager -n eventflow
```

### Test Function CR
```bash
cat <<EOF | kubectl apply -f -
apiVersion: eventflow.io/v1alpha1
kind: Function
metadata:
  name: test-operator-func
  namespace: eventflow
spec:
  image: alpine:latest
  command: ["sh", "-c", "echo 'Hello from Operator!' && sleep 3600"]
  env:
    TEST_VAR: "operator-test"
  replicas: 1
EOF
```

### Check operator logs
```bash
kubectl logs -n eventflow -l control-plane=controller-manager --tail=50 -f
```