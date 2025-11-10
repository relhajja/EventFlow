# Make vs Task vs Shell Scripts

This document explains the different automation options available in the `infra-dev` directory.

## TL;DR

- **Use Make** if you want standard, universal tooling (recommended)
- **Use Task** if you prefer modern YAML-based configuration
- **Use Shell Scripts** if you want maximum control and transparency

## Comparison

### Makefile

**Pros:**
- ✅ Universal - installed on virtually every system
- ✅ Industry standard - everyone knows Make
- ✅ No installation needed
- ✅ Great for C/C++/Go projects (already familiar)
- ✅ Built-in dependency management
- ✅ Parallel execution with `-j`
- ✅ Tab completion support

**Cons:**
- ⚠️ Quirky syntax (tabs required, special variables)
- ⚠️ Shell escaping can be tricky
- ⚠️ Less readable for complex tasks

**Usage:**
```bash
make setup              # Complete setup
make rebuild-api        # Rebuild API only
make logs COMPONENT=api # View API logs
make status             # Show status
make help               # Show all targets
```

### Taskfile.yml (Task)

**Pros:**
- ✅ Modern, clean YAML syntax
- ✅ Better error messages
- ✅ Built-in variable interpolation
- ✅ Native cross-platform support
- ✅ Built-in file watching
- ✅ Better dependency management (deps vs cmds)
- ✅ Cleaner output

**Cons:**
- ⚠️ Requires installation (one extra step)
- ⚠️ Less universal (newer tool)
- ⚠️ Smaller ecosystem

**Installation:**
```bash
# macOS
brew install go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Or via Go
go install github.com/go-task/task/v3/cmd/task@latest
```

**Usage:**
```bash
task setup              # Complete setup
task rebuild:api        # Rebuild API only
task logs:api           # View API logs
task status             # Show status
task --list             # Show all tasks
```

### Shell Scripts

**Pros:**
- ✅ Maximum transparency - see exactly what runs
- ✅ Easy to customize inline
- ✅ No build tool needed
- ✅ Can be used from Make/Task
- ✅ Great for learning

**Cons:**
- ⚠️ No dependency management
- ⚠️ No parallel execution
- ⚠️ More verbose for complex workflows
- ⚠️ Manual orchestration

**Usage:**
```bash
./install-k3s.sh        # Install K3s
./build-images.sh       # Build images
./deploy.sh             # Deploy EventFlow
./dev.sh setup          # Complete setup
./dev.sh rebuild-api    # Rebuild API
./monitor.sh api        # Monitor logs
```

## Feature Comparison

| Feature | Make | Task | Shell Scripts |
|---------|------|------|---------------|
| Universal | ✅ Yes | ⚠️ Needs install | ✅ Yes |
| Syntax | ⚠️ Complex | ✅ Clean YAML | ✅ Bash |
| Dependencies | ✅ Yes | ✅ Yes | ❌ Manual |
| Parallel | ✅ Yes (`-j`) | ✅ Yes (`deps`) | ❌ No |
| Variables | ⚠️ Complex | ✅ Easy | ✅ Easy |
| Help/Docs | ✅ Built-in | ✅ Built-in | ⚠️ Manual |
| Tab Complete | ✅ Yes | ✅ Yes | ⚠️ Limited |
| Learning Curve | ⚠️ Medium | ✅ Low | ✅ Low |

## Examples

### Complete Setup

**Make:**
```bash
make setup
```

**Task:**
```bash
task setup
```

**Shell:**
```bash
./dev.sh setup
# Or manually:
./install-k3s.sh && ./build-images.sh && ./deploy.sh
```

### Rebuild API Only

**Make:**
```bash
make rebuild-api
```

**Task:**
```bash
task rebuild:api
```

**Shell:**
```bash
./dev.sh rebuild-api
```

### View Logs

**Make:**
```bash
make logs COMPONENT=api
# Or shorter:
make api  # Convenience alias
```

**Task:**
```bash
task logs:api
```

**Shell:**
```bash
./dev.sh logs api
```

### Show Status

**Make:**
```bash
make status
```

**Task:**
```bash
task status
```

**Shell:**
```bash
./dev.sh status
```

## Recommendations

### For EventFlow Contributors

**Use Make** - It's standard, works everywhere, and most Go developers are familiar with it.

```bash
# Quick start
make setup
make rebuild-api
make status
```

### For Modern Workflow Enthusiasts

**Use Task** - Cleaner syntax, better for complex workflows.

```bash
# Install once
brew install go-task  # or see installation above

# Then use
task setup
task rebuild:api
task status
```

### For Learning/Debugging

**Use Shell Scripts** - See exactly what's happening.

```bash
./dev.sh setup
./dev.sh logs api
```

## Best Practices

### All Three Coexist

The three approaches complement each other:

1. **Make/Task** - For common workflows
2. **Shell Scripts** - Can be called by Make/Task for complex logic
3. **Documentation** - Keep QUICKSTART.md updated

### Consistency

All three tools provide the same operations:
- `setup` - Complete installation
- `build` - Build all images
- `deploy` - Deploy to K3s
- `rebuild-api/operator/web` - Quick rebuilds
- `logs` - Follow logs
- `status` - Show cluster state
- `clean` - Remove EventFlow
- `reset` - Clean and redeploy

### Choose One, Master It

For daily development, pick one tool and stick with it:
- **Make** for universality
- **Task** for modern workflows
- **Scripts** for maximum control

All will get the job done!

## Migration Path

### From Shell Scripts to Make

The Makefile internally calls the same commands as the shell scripts. No migration needed - both work!

### From Make to Task

Task commands are similar but with `:` instead of `-`:

```bash
make rebuild-api   → task rebuild:api
make logs api      → task logs:api
make shell api     → task shell:api
```

### From Task to Make

Use `-` instead of `:`:

```bash
task rebuild:api   → make rebuild-api
task logs:api      → make logs COMPONENT=api
task shell:api     → make shell COMPONENT=api
```

## Summary

- ✅ **Makefile** - Best for most users (standard, universal)
- ✅ **Taskfile** - Best for modern workflows (clean, powerful)
- ✅ **Scripts** - Best for transparency (direct, simple)

All three are maintained and fully functional. Choose based on your preference!
