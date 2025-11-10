#!/bin/bash

# Monitor EventFlow API for "too many open files" error
# Usage: ./monitor-api.sh

echo "Monitoring EventFlow API pods for errors..."
echo "Press Ctrl+C to stop"
echo ""

kubectl logs -f -n eventflow deployment/eventflow-api --all-containers=true 2>&1 | grep --line-buffered -E "(too many open files|error|Error|ERROR|failed|Failed|FAILED|panic|Panic|PANIC)"
