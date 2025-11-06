# EventFlow - Event-Driven FaaS Platform

## Architecture

```
┌─────────────┐      ┌──────────┐      ┌────────────┐      ┌──────────────┐
│   Client    │─────>│   API    │─────>│    NATS    │─────>│  Dispatcher  │
│  (HTTP/Web) │      │  Server  │      │  (Queue)   │      │   (Worker)   │
└─────────────┘      └──────────┘      └────────────┘      └──────────────┘
                           │                                        │
                           │                                        ▼
                           ▼                                ┌──────────────┐
                    ┌─────────────┐                         │  Kubernetes  │
                    │  Dashboard  │                         │   Function   │
                    │    (Web)    │                         │     Pods     │
                    └─────────────┘                         └──────────────┘
```

## Components

1. **API Server** (Port 8080)
   - REST API for managing functions
   - Publishes events to NATS
   - JWT authentication

2. **NATS** (Port 4222)
   - Event message queue with JetStream
   - Persistent event storage (24h retention)
   - Reliable event delivery

3. **Dispatcher** 
   - Consumes events from NATS
   - Invokes functions via Kubernetes Jobs
   - Auto-scales based on queue depth

4. **Web Dashboard** (Port 3000)
   - Function management UI
   - Event monitoring
   - Real-time metrics

## Quick Start

```bash
# Start all services
make run

# Check status
make status

# View logs
make logs
```

## Services

- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **NATS Monitoring**: http://localhost:8222
- **Metrics**: http://localhost:9090/metrics

## Testing the Event Flow

### 1. Get Authentication Token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)
```

### 2. Create a Function (Demo Mode)

```bash
curl -X POST http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hello-world",
    "image": "myregistry/hello-function:latest",
    "replicas": 1,
    "env": {
      "ENV": "production"
    }
  }'
```

### 3. Trigger Function via Event

```bash
# This publishes an event to NATS
curl -X POST http://localhost:8080/v1/functions/hello-world:invoke \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user": "john",
    "action": "process_order",
    "orderId": "12345"
  }'
```

### 4. Watch Dispatcher Process Events

```bash
sudo docker logs -f eventflow-dispatcher
```

You should see:
```
📨 Received event: http.invoke -> function: hello-world
🎯 Invoking function 'hello-world' with payload: map[action:process_order orderId:12345 user:john]
✅ Event processed successfully
```

### 5. List Functions

```bash
curl http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Event Types

EventFlow supports different event types:

- `http.invoke` - Direct HTTP invocation
- `webhook.received` - External webhooks
- `schedule.triggered` - Cron/scheduled events
- `message.queued` - Message queue events
- Custom event types

## Current Status (Skeleton Mode)

✅ **Implemented:**
- NATS event queue with JetStream
- Event publisher in API
- Event dispatcher worker
- HTTP API for function management
- Demo mode (works without Kubernetes)

⏳ **Next Steps:**
- Custom Resource Definition (CRD) for functions
- Kubernetes controller for function lifecycle
- Auto-scaling based on event queue depth
- Multi-region support
- Build service for code bundles

## Development

### Run Locally (without Docker)

```bash
# Terminal 1: Start NATS
docker run -p 4222:4222 nats:2.10-alpine -js

# Terminal 2: Start API
cd api && go run main.go

# Terminal 3: Start Dispatcher  
cd dispatcher && go run main.go

# Terminal 4: Start Web
cd web && npm run dev
```

### Deploy to Kubernetes

```bash
# Create kind cluster and deploy
make kind-setup

# Port forward API
kubectl port-forward -n eventflow svc/eventflow-api 8080:80

# Port forward Web
kubectl port-forward -n eventflow svc/eventflow-web 3000:80
```

## Architecture Decisions

### Why NATS?
- Lightweight (< 20MB memory)
- JetStream provides persistence
- Built-in clustering and HA
- Perfect for event-driven systems

### Why Separate Dispatcher?
- Decouples API from function execution
- Allows independent scaling
- Better fault tolerance
- Can add multiple dispatchers for parallel processing

### Why Demo Mode?
- Test locally without Kubernetes
- Faster development iteration
- CI/CD friendly
- Easy onboarding

## Next Evolution

This is a **working skeleton**. To make it production-ready:

1. **Add Kubernetes CRD**: Define `Function` as a custom resource
2. **Build Controller**: Watch `Function` objects and manage deployments
3. **Add HPA**: Scale based on NATS queue depth
4. **Add Build Service**: Accept code bundles and build containers
5. **Add Storage**: S3/Minio for function artifacts
6. **Add Observability**: Logging, tracing, metrics
7. **Add Multi-tenancy**: Namespaces, resource quotas
8. **Add Registry**: Internal container registry




## License

MIT
