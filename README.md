# 🚀 RiadCloud - Docker Test

Simple Docker setup to test the complete stack.

## Quick Start

```bash
make up     # Start all services
make test   # Test endpoints  
make down   # Stop services
```

## What's Included

- React Frontend (port 3000)
- Go Backend API (port 8080)
- PostgreSQL Database (port 5432) 
- Redis Cache (port 6379)
- Adminer UI (port 8081)

## Commands

- `make up` - Build and start all services
- `make test` - Test all endpoints
- `make down` - Stop services
- `make clean` - Remove everything

## Access

- **Frontend**: http://localhost:3000 (React UI)
- **Backend API**: http://localhost:8080
- **Adminer**: http://localhost:8081
  - Server: postgres
  - User: riadcloud
  - Password: riadcloud_dev_pass

## Architecture

Frontend (React) → Backend (Go) → Database (PostgreSQL) + Cache (Redis)

- Frontend makes API calls to `/api/*` which are proxied to the Go backend
- Backend serves API responses and handles database/cache operations
- All services run in isolated Docker containers

## That's It!

Three commands to test everything in Docker:
```bash
make up
make test
make down
```

