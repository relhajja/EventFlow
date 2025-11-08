# EventFlow API Reference

## Base URL

```
http://localhost:30080  # NodePort (kind cluster)
http://localhost:8080   # Port forward
```

## Authentication

All endpoints (except `/auth/token`) require JWT authentication.

### Headers

```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

## Endpoints

### Authentication

#### Generate Token

```http
POST /auth/token
```

Generate a JWT token for authentication.

**Request Body:**
```json
{
  "user_id": "alice",
  "username": "Alice",
  "email": "alice@company.com"
}
```

**Response:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "alice",
  "username": "Alice",
  "email": "alice@company.com",
  "namespace": "tenant-alice"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:30080/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "alice",
    "username": "Alice",
    "email": "alice@company.com"
  }'
```

---

### Functions

#### Create Function

```http
POST /v1/functions
```

Create a new function in the authenticated user's namespace.

**Request Body:**
```json
{
  "name": "my-function",
  "image": "nginx:alpine",
  "replicas": 2,
  "command": ["nginx", "-g", "daemon off;"],
  "env": {
    "ENV": "production",
    "LOG_LEVEL": "info"
  }
}
```

**Field Descriptions:**
- `name` (required): Unique function name within your namespace
- `image` (required): Container image (Docker Hub or private registry)
- `replicas` (optional): Number of pod replicas (default: 1)
- `command` (optional): Container command override
- `env` (optional): Environment variables as key-value pairs

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "alice",
  "name": "my-function",
  "namespace": "tenant-alice",
  "image": "nginx:alpine",
  "replicas": 2,
  "command": ["nginx", "-g", "daemon off;"],
  "env": {
    "ENV": "production",
    "LOG_LEVEL": "info"
  },
  "status": "Pending",
  "created_at": "2025-11-08T10:30:00Z",
  "updated_at": "2025-11-08T10:30:00Z"
}
```

**cURL Example:**
```bash
TOKEN="your_jwt_token_here"

curl -X POST http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-server",
    "image": "nginx:alpine",
    "replicas": 2,
    "env": {
      "ENV": "production"
    }
  }'
```

**Error Responses:**

- `400 Bad Request` - Invalid request body
```json
{
  "error": "Bad Request",
  "message": "name is required"
}
```

- `401 Unauthorized` - Missing or invalid JWT token
```json
{
  "error": "Unauthorized",
  "message": "missing authorization header"
}
```

- `409 Conflict` - Function with same name already exists
```json
{
  "error": "Conflict",
  "message": "function 'my-function' already exists in namespace 'tenant-alice'"
}
```

---

#### List Functions

```http
GET /v1/functions
```

List all functions for the authenticated user.

**Response:** `200 OK`
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "alice",
    "name": "web-server",
    "namespace": "tenant-alice",
    "image": "nginx:alpine",
    "replicas": 2,
    "status": "Running",
    "available_replicas": 2,
    "created_at": "2025-11-08T10:30:00Z",
    "updated_at": "2025-11-08T10:31:00Z"
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "alice",
    "name": "api-backend",
    "namespace": "tenant-alice",
    "image": "node:18-alpine",
    "replicas": 3,
    "status": "Running",
    "available_replicas": 3,
    "created_at": "2025-11-08T09:15:00Z",
    "updated_at": "2025-11-08T09:16:00Z"
  }
]
```

**cURL Example:**
```bash
curl http://localhost:30080/v1/functions \
  -H "Authorization: Bearer $TOKEN"
```

---

#### Get Function

```http
GET /v1/functions/{name}
```

Get details of a specific function.

**Path Parameters:**
- `name`: Function name

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "alice",
  "name": "web-server",
  "namespace": "tenant-alice",
  "image": "nginx:alpine",
  "replicas": 2,
  "command": ["nginx", "-g", "daemon off;"],
  "env": {
    "ENV": "production",
    "LOG_LEVEL": "info"
  },
  "status": "Running",
  "available_replicas": 2,
  "created_at": "2025-11-08T10:30:00Z",
  "updated_at": "2025-11-08T10:31:00Z"
}
```

**cURL Example:**
```bash
curl http://localhost:30080/v1/functions/web-server \
  -H "Authorization: Bearer $TOKEN"
```

**Error Responses:**

- `404 Not Found` - Function doesn't exist or doesn't belong to user
```json
{
  "error": "Not Found",
  "message": "function 'web-server' not found"
}
```

---

#### Delete Function

```http
DELETE /v1/functions/{name}
```

Delete a function (soft delete).

**Path Parameters:**
- `name`: Function name

**Response:** `204 No Content`

**cURL Example:**
```bash
curl -X DELETE http://localhost:30080/v1/functions/web-server \
  -H "Authorization: Bearer $TOKEN"
```

**Error Responses:**

- `404 Not Found` - Function doesn't exist
```json
{
  "error": "Not Found",
  "message": "function 'web-server' not found"
}
```

---

#### Invoke Function

```http
POST /v1/functions/{name}:invoke
```

Invoke a function with optional payload.

**Path Parameters:**
- `name`: Function name

**Request Body (Optional):**
```json
{
  "input": {
    "key": "value",
    "data": "payload"
  }
}
```

**Response:** `200 OK`
```json
{
  "message": "Function 'web-server' invoked successfully",
  "job_name": "web-server-invoke-1699564800"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:30080/v1/functions/web-server:invoke \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "action": "process",
      "data": "test"
    }
  }'
```

---

#### Get Function Logs

```http
GET /v1/functions/{name}/logs
```

Stream logs from function pods.

**Path Parameters:**
- `name`: Function name

**Query Parameters:**
- `follow` (optional): Follow logs in real-time (`true`/`false`, default: `false`)
- `tail` (optional): Number of lines to tail (default: 100)

**Response:** `200 OK` (text/plain or event-stream)

```
2025-11-08T10:30:15Z [INFO] Server started on port 8080
2025-11-08T10:30:20Z [INFO] Received request: GET /health
2025-11-08T10:30:21Z [INFO] Health check passed
```

**cURL Examples:**
```bash
# Get last 100 lines
curl http://localhost:30080/v1/functions/web-server/logs \
  -H "Authorization: Bearer $TOKEN"

# Follow logs in real-time
curl http://localhost:30080/v1/functions/web-server/logs?follow=true \
  -H "Authorization: Bearer $TOKEN"

# Tail last 50 lines
curl http://localhost:30080/v1/functions/web-server/logs?tail=50 \
  -H "Authorization: Bearer $TOKEN"
```

---

### Health Checks

#### Health Check

```http
GET /healthz
```

Check API server health.

**Response:** `200 OK`
```json
{
  "status": "healthy"
}
```

---

#### Readiness Check

```http
GET /readyz
```

Check if API server is ready to accept traffic.

**Response:** `200 OK`
```json
{
  "status": "ready"
}
```

---

### Metrics

#### Prometheus Metrics

```http
GET /metrics
```

Get Prometheus metrics.

**Response:** `200 OK` (text/plain)
```
# HELP eventflow_api_requests_total Total number of API requests
# TYPE eventflow_api_requests_total counter
eventflow_api_requests_total{method="GET",endpoint="/v1/functions",status="200"} 145

# HELP eventflow_api_request_duration_seconds API request duration in seconds
# TYPE eventflow_api_request_duration_seconds histogram
eventflow_api_request_duration_seconds_bucket{method="POST",endpoint="/v1/functions",le="0.1"} 89
eventflow_api_request_duration_seconds_bucket{method="POST",endpoint="/v1/functions",le="0.5"} 95
eventflow_api_request_duration_seconds_bucket{method="POST",endpoint="/v1/functions",le="1"} 98

# HELP eventflow_functions_total Total number of functions
# TYPE eventflow_functions_total gauge
eventflow_functions_total{namespace="tenant-alice",status="Running"} 5
eventflow_functions_total{namespace="tenant-bob",status="Running"} 3
```

---

## Error Response Format

All error responses follow this format:

```json
{
  "error": "Error Type",
  "message": "Detailed error message"
}
```

### HTTP Status Codes

| Code | Meaning | Usage |
|------|---------|-------|
| 200 | OK | Successful GET request |
| 201 | Created | Successful POST (create) |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Invalid request body or parameters |
| 401 | Unauthorized | Missing or invalid JWT token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource already exists |
| 500 | Internal Server Error | Server error |

---

## Rate Limiting

Currently no rate limiting is enforced. For production deployments, consider:
- API Gateway with rate limiting
- Per-tenant rate limits
- Resource quota enforcement

---

## Pagination

Currently, all list endpoints return all results. For large datasets, implement pagination:

```http
GET /v1/functions?page=1&limit=20
```

---

## Webhook Support (Future)

Future support for webhooks to notify external systems:

```http
POST /v1/webhooks
```

Register a webhook URL to receive events:
- `function.created`
- `function.updated`
- `function.deleted`
- `function.invoked`

---

## SDK Examples

### JavaScript/TypeScript

```typescript
import axios from 'axios';

class EventFlowClient {
  private baseURL: string;
  private token: string;

  constructor(baseURL: string, token: string) {
    this.baseURL = baseURL;
    this.token = token;
  }

  async createFunction(name: string, image: string, replicas: number = 1) {
    const response = await axios.post(
      `${this.baseURL}/v1/functions`,
      { name, image, replicas },
      {
        headers: {
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json'
        }
      }
    );
    return response.data;
  }

  async listFunctions() {
    const response = await axios.get(
      `${this.baseURL}/v1/functions`,
      {
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      }
    );
    return response.data;
  }

  async deleteFunction(name: string) {
    await axios.delete(
      `${this.baseURL}/v1/functions/${name}`,
      {
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      }
    );
  }
}

// Usage
const client = new EventFlowClient('http://localhost:30080', 'your_token');
await client.createFunction('my-app', 'nginx:alpine', 2);
const functions = await client.listFunctions();
await client.deleteFunction('my-app');
```

### Python

```python
import requests

class EventFlowClient:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    def create_function(self, name, image, replicas=1):
        response = requests.post(
            f'{self.base_url}/v1/functions',
            json={'name': name, 'image': image, 'replicas': replicas},
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()
    
    def list_functions(self):
        response = requests.get(
            f'{self.base_url}/v1/functions',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()
    
    def delete_function(self, name):
        response = requests.delete(
            f'{self.base_url}/v1/functions/{name}',
            headers=self.headers
        )
        response.raise_for_status()

# Usage
client = EventFlowClient('http://localhost:30080', 'your_token')
client.create_function('my-app', 'nginx:alpine', 2)
functions = client.list_functions()
client.delete_function('my-app')
```

### Go

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EventFlowClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

type Function struct {
	Name     string            `json:"name"`
	Image    string            `json:"image"`
	Replicas int               `json:"replicas"`
	Env      map[string]string `json:"env,omitempty"`
}

func (c *EventFlowClient) CreateFunction(fn Function) error {
	body, _ := json.Marshal(fn)
	req, _ := http.NewRequest("POST", c.BaseURL+"/v1/functions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func (c *EventFlowClient) ListFunctions() ([]Function, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/v1/functions", nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var functions []Function
	if err := json.NewDecoder(resp.Body).Decode(&functions); err != nil {
		return nil, err
	}
	return functions, nil
}

// Usage
func main() {
	client := &EventFlowClient{
		BaseURL: "http://localhost:30080",
		Token:   "your_token",
		Client:  &http.Client{},
	}
	
	fn := Function{
		Name:     "my-app",
		Image:    "nginx:alpine",
		Replicas: 2,
	}
	
	if err := client.CreateFunction(fn); err != nil {
		panic(err)
	}
	
	functions, err := client.ListFunctions()
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Found %d functions\n", len(functions))
}
```

---

## Testing with Postman

### Import Collection

Create a Postman collection with these environment variables:

```json
{
  "base_url": "http://localhost:30080",
  "token": "{{token}}"
}
```

### Pre-request Script (Get Token)

```javascript
pm.sendRequest({
    url: pm.environment.get("base_url") + "/auth/token",
    method: 'POST',
    header: {
        'Content-Type': 'application/json'
    },
    body: {
        mode: 'raw',
        raw: JSON.stringify({
            user_id: "alice",
            username: "Alice",
            email: "alice@company.com"
        })
    }
}, function (err, response) {
    const data = response.json();
    pm.environment.set("token", data.token);
});
```

---

## API Versioning

Current version: `v1`

Future versions will be available at:
- `/v2/functions`
- `/v3/functions`

Breaking changes will result in a new version.
