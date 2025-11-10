# Task Quick Reference

Quick reference for common Taskfile commands.

## 🚀 Getting Started

```bash
task setup              # Complete setup (install → build → deploy)
task verify             # Verify everything is working
```

## 🔨 Development Workflow

### Fast Iteration
```bash
task rebuild:api        # Rebuild & redeploy API (30-60s)
task rebuild:operator   # Rebuild & redeploy Operator
task rebuild:web        # Rebuild & redeploy Web UI
```

### View Logs
```bash
task logs:api           # Follow API logs
task logs:operator      # Follow Operator logs
task logs:web           # Follow Web UI logs
task logs:postgres      # Follow PostgreSQL logs
```

### Monitor for Errors
```bash
task monitor:api        # Watch API for errors only
task monitor:operator   # Watch Operator for errors only
task monitor:all        # Watch all components
```

### Check Status
```bash
task status             # Show all pods, namespaces, functions
task pods               # Show pods with detailed info
task events             # Show recent Kubernetes events
task images             # List imported images
```

## 🔍 Debugging

```bash
task describe:api       # Describe API pod
task describe:operator  # Describe Operator pod
task shell:api          # Open shell in API pod
task shell:postgres     # Open PostgreSQL shell
```

## 🧪 Testing

### Get JWT Token
```bash
task test:get-token USER=alice
task test:get-token USER=bob
```

### Create Test Function
```bash
export TOKEN=$(task test:get-token USER=alice)
task test:create-function
```

### List Functions
```bash
export TOKEN=$(task test:get-token USER=alice)
task test:list-functions
```

## 🔄 Restart

```bash
task restart            # Restart all deployments
task restart:api        # Restart API only
task restart:operator   # Restart Operator only
task restart:web        # Restart Web UI only
```

## 🧹 Cleanup

```bash
task clean              # Remove EventFlow (keep K3s)
task reset              # Clean and redeploy
task uninstall          # Remove K3s completely
```

## 🎯 Quick Aliases

```bash
task up                 # = task setup
task down               # = task clean
```

## 💡 Pro Tips

### Chain Commands
```bash
task clean && task deploy          # Clean and fresh deploy
task build:api && task restart:api # Build and restart API
```

### Use in Scripts
```bash
#!/bin/bash
task setup
task verify
if [ $? -eq 0 ]; then
    echo "✅ EventFlow is ready!"
fi
```

### Environment Variables
```bash
# Get a token and use it
export TOKEN=$(task test:get-token USER=alice)

# Create function
task test:create-function

# List functions
task test:list-functions
```

### Watch Mode (Auto-rebuild)
```bash
task dev                # Watch for changes and rebuild API
```

### Port Forwarding
```bash
task port-forward       # Forward localhost:8080 -> API
# Then access: http://localhost:8080
```

### Use k9s Dashboard
```bash
task dashboard          # Open k9s in eventflow namespace
```

## 📋 Complete Task List

Run this to see all available tasks:
```bash
task --list
```

## 🎨 Tips for Daily Use

### Morning Routine
```bash
cd infra-dev
task status             # Check if everything is running
task logs:api           # Check for any overnight issues
```

### After Code Changes
```bash
# Edit code in API
vim ../eventflow/api/internal/handlers/functions.go

# Quick rebuild
task rebuild:api        # Builds, imports, and restarts in ~30s

# Watch logs
task logs:api
```

### Before Committing
```bash
task verify             # Ensure everything works
task test:get-token USER=alice
export TOKEN=<token>
task test:create-function
task test:list-functions
```

### End of Day
```bash
task status             # Check final state
task down               # Optional: clean up resources
```

## 🐛 Troubleshooting

### API Won't Start
```bash
task describe:api       # Check pod events
task logs:api           # Check error messages
task restart:api        # Try restart
```

### Operator Not Working
```bash
task logs:operator      # Check logs
task describe:operator  # Check pod status
kubectl get functions.eventflow.eventflow.io -A  # Check CRs
```

### Database Issues
```bash
task logs:postgres      # Check database logs
task shell:postgres     # Connect to database
# In psql: \dt, SELECT * FROM functions;
```

### Image Not Found
```bash
task images             # Verify images are imported
task build:api          # Rebuild specific image
```

### "Too Many Open Files"
```bash
# Already fixed in latest version with:
# - QPS = 10
# - Burst = 15
# - Timeout = 10s
# - defer r.Body.Close()

# If still happening:
task rebuild:api        # Rebuild with latest fixes
```

## 🔗 Related Files

- `Taskfile.yml` - All task definitions
- `QUICKSTART.md` - Detailed getting started guide
- `TROUBLESHOOTING.md` - Common issues and solutions
- `TOOLS.md` - Comparison with Make and shell scripts

---

**Quick Start in 3 Commands:**
```bash
task setup              # Install and deploy everything
task verify             # Verify it's working
task logs:api           # Watch it run
```
