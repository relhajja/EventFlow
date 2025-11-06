# 🚀 RiadCloud Development Pipeline

Complete development workflow with hot-reloading, testing, and automation.

## 📋 Quick Start

### 1. Install Development Tools
```bash
make install-tools
```
This installs:
- **Air** - Hot-reload tool for Go
- **golangci-lint** - Fast Go linters runner

### 2. Start Development
```bash
make dev
```
This will:
- Start Docker services (PostgreSQL, Redis, Adminer)
- Start the app with hot-reload enabled
- Automatically rebuild on file changes

## 🛠️ Available Commands

### Development
```bash
make dev          # Start with hot-reload (recommended)
make run          # Run without hot-reload
make watch        # Alternative watcher (uses inotifywait)
```

### Building
```bash
make build        # Build binary to bin/webapp
make clean        # Remove build artifacts
```

### Testing
```bash
make test              # Run all tests
make test-coverage     # Run tests with coverage report
```

### Code Quality
```bash
make fmt          # Format code
make lint         # Run linter
```

### Docker Management
```bash
make docker-up       # Start all services
make docker-down     # Stop all services
make docker-logs     # View logs
make docker-restart  # Restart services
```

### Database
```bash
make db-shell     # Open PostgreSQL shell
make db-migrate   # Run migrations
```

### CI/CD
```bash
make all         # Run fmt, lint, test, build
make ci          # Run lint, test, build (CI simulation)
```

## 📊 Development Workflow

### Typical Development Session

1. **Start services and development server:**
   ```bash
   make dev
   ```

2. **Make changes to code** - The server will auto-reload

3. **Run tests:**
   ```bash
   make test
   ```

4. **Check code quality:**
   ```bash
   make fmt lint
   ```

5. **Build for deployment:**
   ```bash
   make build
   ```

### Hot-Reload Details

When you run `make dev`, Air watches for changes in:
- `*.go` files
- `*.html` files
- `*.tpl`, `*.tmpl` template files

On changes, it will:
1. Rebuild the application
2. Kill the old process
3. Start the new binary
4. Show build errors in the terminal

## 🧪 Testing Pipeline

### Run Tests
```bash
make test
```

### With Coverage
```bash
make test-coverage
```
This generates:
- `coverage.out` - Coverage data
- `coverage.html` - Visual coverage report (open in browser)

### View Coverage Report
```bash
make test-coverage
xdg-open coverage.html  # Linux
```

## 🐳 Docker Services

### Start All Services
```bash
make docker-up
```

Running services:
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Adminer: `http://localhost:8081`
- Webapp: `http://localhost:8080`

### View Logs
```bash
make docker-logs
```

### Database Shell
```bash
make db-shell
```

Then you can run SQL:
```sql
SELECT * FROM services;
\dt  -- List tables
\q   -- Quit
```

## 📁 Project Structure

```
webapp/
├── main.go              # Main application
├── main_test.go         # Tests
├── index.html           # Frontend
├── Makefile            # Build automation
├── .air.toml           # Hot-reload config
├── docker-compose.yml  # Docker services
├── init.sql            # Database schema
├── Dockerfile          # App container
├── scripts/
│   └── watch.sh        # Alternative watcher
├── bin/                # Built binaries (gitignored)
└── tmp/                # Temporary build files (gitignored)
```

## 🔄 CI/CD Integration

### GitHub Actions Example

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make ci
```

### Local CI Simulation
```bash
make ci
```

## 🎯 Common Tasks

### First Time Setup
```bash
# 1. Install tools
make install-tools

# 2. Start services
make docker-up

# 3. Verify everything works
make test
make build
./bin/webapp
```

### Quick Test Before Commit
```bash
make all
```

### Rebuild Everything
```bash
make clean
make build
```

### Reset Database
```bash
make docker-down
sudo docker volume rm webapp_postgres_data
make docker-up
```

## 🐛 Troubleshooting

### Port Already in Use
```bash
# Find process using port 8080
sudo lsof -i :8080

# Kill it
kill -9 <PID>
```

### Docker Permission Denied
```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

### Air Not Found
```bash
# Make sure GOPATH/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Or add to ~/.bashrc
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

### Build Fails
```bash
# Check Go version
go version

# Clean and rebuild
make clean
make build
```

## 📈 Performance Tips

1. **Use hot-reload during development** - Faster than manual restarts
2. **Run tests in watch mode** - Use `make dev` in one terminal, tests in another
3. **Use Docker for dependencies** - Consistent environment
4. **Run linter before commit** - Catch issues early

## 🔐 Production Checklist

Before deploying:
- [ ] Run `make all` successfully
- [ ] All tests pass
- [ ] Coverage > 80%
- [ ] Linter shows no errors
- [ ] Update database passwords
- [ ] Set proper environment variables
- [ ] Review Dockerfile for security
- [ ] Enable HTTPS/TLS

## 📚 Additional Resources

- [Air Documentation](https://github.com/air-verse/air)
- [golangci-lint](https://golangci-lint.run/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)

---

**Questions?** Check `make help` for all available commands.
