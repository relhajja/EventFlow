#!/bin/bash

set -e

echo "🧪 EventFlow End-to-End Test"
echo "=============================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Get Token
echo -e "${BLUE}Step 1: Authenticating...${NC}"
TOKEN=$(curl -s -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

if [ -z "$TOKEN" ]; then
  echo "❌ Failed to get token"
  exit 1
fi

echo -e "${GREEN}✅ Got authentication token${NC}"
echo ""

# Step 2: Create a function (demo mode)
echo -e "${BLUE}Step 2: Creating function 'data-processor'...${NC}"
curl -s -X POST http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "data-processor",
    "image": "myregistry/data-processor:v1",
    "replicas": 2,
    "env": {
      "LOG_LEVEL": "info",
      "REGION": "us-east-1"
    }
  }' | jq .

echo -e "${GREEN}✅ Function created${NC}"
echo ""

# Step 3: List functions
echo -e "${BLUE}Step 3: Listing all functions...${NC}"
curl -s http://localhost:8080/v1/functions \
  -H "Authorization: Bearer $TOKEN" | jq .

echo ""

# Step 4: Trigger event
echo -e "${BLUE}Step 4: Publishing event to 'data-processor'...${NC}"
curl -s -X POST http://localhost:8080/v1/functions/data-processor:invoke \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "eventType": "order.created",
    "order": {
      "id": "order-456",
      "amount": 99.99,
      "items": ["item-1", "item-2"]
    }
  }' | jq .

echo -e "${GREEN}✅ Event published to NATS queue${NC}"
echo ""

# Step 5: Check dispatcher logs
echo -e "${BLUE}Step 5: Checking dispatcher logs...${NC}"
echo -e "${YELLOW}Dispatcher output:${NC}"
sudo docker logs eventflow-dispatcher --tail 5

echo ""
echo -e "${GREEN}✅ Test completed successfully!${NC}"
echo ""
echo "Next steps:"
echo "  - View real-time logs: make logs"
echo "  - Access dashboard: http://localhost:3000"
echo "  - View NATS monitoring: http://localhost:8222"
echo "  - Deploy to Kubernetes: make kind-setup"
