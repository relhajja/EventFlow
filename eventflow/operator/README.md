# EventFlow Operator

Kubernetes operator for managing EventFlow Functions using the Operator pattern.

## Description

The EventFlow Operator watches for `Function` custom resources and automatically creates and manages Kubernetes Deployments to run user-defined functions. It follows the Kubernetes Operator pattern using Kubebuilder.

**Architecture**: Frontend → Control API → Function CRD → Operator → Deployments

## Getting Started

### Prerequisites
- Go 1.22+
- Docker
- kubectl
- kind cluster (for local development)

### Local Development with kind

**Build and deploy the operator:**

```bash
# Build the operator image
make docker-build IMG=eventflow-operator:latest

# Load into kind cluster
kind load docker-image eventflow-operator:latest --name eventflow

# Deploy to cluster (applies CRDs, RBAC, and operator deployment)
make deploy
```

**Quick redeploy after code changes:**

```bash
make docker-build IMG=eventflow-operator:latest && \
  kind load docker-image eventflow-operator:latest --name eventflow && \
  kubectl rollout restart deployment/operator-controller-manager -n eventflow
```

### Create a Function

```bash
cat <<EOF | kubectl apply -f -
apiVersion: eventflow.io/v1alpha1
kind: Function
metadata:
  name: hello-function
  namespace: eventflow
spec:
  image: alpine:latest
  command: ["sh", "-c", "echo 'Hello from EventFlow!' && sleep 3600"]
  env:
    GREETING: "Hello"
  replicas: 1
EOF
```

### View Functions

```bash
# List all functions
kubectl get functions -n eventflow

# Get function details
kubectl describe function hello-function -n eventflow

# Check operator logs
kubectl logs -n eventflow -l control-plane=controller-manager --tail=50 -f
```

### To Uninstall

```bash
# Delete function instances
kubectl delete functions --all -n eventflow

# Undeploy operator
make undeploy

# Uninstall CRDs
make uninstall
```

## Development

**Generate manifests (after modifying types):**
```bash
make manifests
```

**Generate code (DeepCopy methods):**
```bash
make generate
```

**Run tests:**
```bash
make test
```

**Build locally without Docker:**
```bash
make build
```

## Project Structure

- `api/v1alpha1/` - Function CRD types and schema
- `internal/controller/` - Reconciliation logic
- `config/` - Kustomize manifests (CRD, RBAC, deployment)
- `config/samples/` - Example Function CRs

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

