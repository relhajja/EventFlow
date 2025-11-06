#!/bin/bash

# Test function deployment script
set -e

API_URL="${API_URL:-http://localhost:8080}"

echo "🧪 Testing EventFlow Function Deployment"
echo "========================================"
echo ""

# Get token
echo "📝 Getting dev token..."
TOKEN=$(curl -s -X POST "$API_URL/auth/token" | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "❌ Failed to get token"
  exit 1
fi

echo "✅ Got token: ${TOKEN:0:20}..."
echo ""

# Create function
echo "🚀 Creating test function..."
curl -X POST "$API_URL/v1/functions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-nginx",
    "image": "nginx:alpine",
    "replicas": 1,
    "env": {
      "TEST_VAR": "hello-eventflow"
    }
  }' | jq .

echo ""
echo "⏳ Waiting for function to be ready..."
sleep 5

# List functions
echo ""
echo "📋 Listing functions..."
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_URL/v1/functions" | jq .

# Get function details
echo ""
echo "🔍 Getting function details..."
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_URL/v1/functions/test-nginx" | jq .

# Invoke function
echo ""
echo "▶️  Invoking function..."
curl -X POST "$API_URL/v1/functions/test-nginx:invoke" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"payload": {}}' | jq .

# Get logs
echo ""
echo "📜 Getting function logs..."
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_URL/v1/functions/test-nginx/logs" | head -n 20

# Delete function
echo ""
echo "🗑️  Deleting function..."
curl -X DELETE "$API_URL/v1/functions/test-nginx" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo ""
echo "✅ Test completed successfully!"
