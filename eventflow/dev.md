

kind create cluster --config /home/ironik/GolandProjects/webapp/eventflow/kind-config.yaml

kind delete cluster --name eventflow && kind create cluster --config /home/ironik/GolandProjects/webapp/eventflow/kind-config.yaml

kubectl apply -f k8s/secrets.yaml && kubectl apply -f k8s/rbac.yaml && kubectl apply -f k8s/postgres.yaml && kubectl apply -f k8s/nats.yaml

kind load docker-image postgres:16-alpine nats:2.10-alpine --name eventflow

## Test dispatcher:
docker build -t eventflow-dispatcher:latest . && kind load docker-image eventflow-dispatcher:latest --name eventflow && kubectl rollout restart deployment/eventflow-dispatcher -n eventflow

# Test api

curl -X POST http://localhost:8081/auth/token \
  -H "Content-Type: application/json" \
  -d '{"username": "admin"}'

TOKEN=$(curl -s -X POST http://localhost:8081/auth/token -H "Content-Type: application/json" -d '{"username": "admin"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4) && echo "Token: $TOKEN"


curl -X POST http://localhost:8081/v1/functions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImRldi11c2VyIiwibmFtZXNwYWNlIjoiZXZlbnRmbG93IiwiZXhwIjoxNzYyMzg2Mzg3fQ.NqQDWyiePgFXlwlEAZhCMCcT9GnrDJCJaD0kxeJMXYs" \
  -d '{
    "name": "test-func",
    "description": "Test function for dispatcher",
    "image": "alpine:latest",
    "command": ["echo", "Hello from EventFlow!"],
    "env": {"TEST_VAR": "test-value"}
  }'


## Debug dispatcher

ironik@ironik:~/GolandProjects/webapp/eventflow/dispatcher$ kubectl wait --for=condition=ready pod -l app=eventflow-dispatcher -n eventflow --timeout=60s && kubectl logs -n eventflow -l app=eventflow-dispatcher --tail=10
pod/eventflow-dispatcher-59f5fb4b98-8fshg condition met
2025/11/05 00:21:47 🚀 EventFlow Dispatcher starting...
2025/11/05 00:21:47 ✅ Connected to NATS
2025/11/05 00:21:47 📡 Listening for events on 'eventflow.events' stream...
2025/11/05 00:22:02 📨 Received event: http.invoke -> function: test-func
2025/11/05 00:22:02 🎯 Invoking function 'test-func' with payload: map[]
2025/11/05 00:22:02 ✅ Created invocation job for function 'test-func' with image 'alpine:latest'
2025/11/05 00:22:02 ✅ Event processed successfully


# Publish 

kubectl run nats-pub --rm -i --tty --image=natsio/nats-box:latest --restart=Never -n eventflow -- nats pub eventflow.events '{"id":"manual-test-001","type":"http.invoke","function":"riad-world","image":"alpine:latest","command":["sh","-c","echo Hello from manual NATS event! && env | grep EVENT_"],"payload":{"message":"Testing direct NATS publishing"},"timestamp":"2025-11-05T00:00:00Z"}' -s nats://nats:4222

## This works;
kubectl delete pod nats-pub -n eventflow 2>/dev/null; kubectl run nats-pub --rm -i --image=natsio/nats-box:latest --image-pull-policy=Never --restart=Never -n eventflow -- nats pub eventflow.events '{"id":"manual-test-001","type":"http.invoke","function":"hello-world","image":"alpine:latest","command":["sh","-c","echo Hello from manual NATS event! && env | grep EVENT_"],"payload":{"message":"Testing direct NATS publishing"},"timestamp":"2025-11-05T00:00:00Z"}' -s nats://nats:4222