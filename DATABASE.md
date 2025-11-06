# RiadCloud Database Setup

## Docker Compose Services

This setup includes:
- **PostgreSQL 16** (port 5432) - Main database
- **Redis 7** (port 6379) - Caching and session storage
- **Adminer** (port 8081) - Database management UI

## Quick Start

### 1. Start all services:
```bash
docker-compose up -d
```

### 2. View logs:
```bash
docker-compose logs -f
```

### 3. Stop all services:
```bash
docker-compose down
```

### 4. Stop and remove all data:
```bash
docker-compose down -v
```

## Database Connection Details

### PostgreSQL
- **Host:** localhost
- **Port:** 5432
- **Database:** riadcloud
- **Username:** riadcloud
- **Password:** riadcloud_dev_pass

**Connection String:**
```
postgresql://riadcloud:riadcloud_dev_pass@localhost:5432/riadcloud
```

### Redis
- **Host:** localhost
- **Port:** 6379
- **No password required (development)**

**Connection String:**
```
redis://localhost:6379
```

## Adminer (Database UI)

Access the database management interface at:
```
http://localhost:8081
```

Login credentials:
- System: PostgreSQL
- Server: postgres
- Username: riadcloud
- Password: riadcloud_dev_pass
- Database: riadcloud

## Database Schema

The `init.sql` script creates:
- `services` table - Cloud service offerings
- `deployments` table - User deployment instances
- `users` table - Platform users
- Indexes for performance
- `active_deployments` view

## Testing Connection

### PostgreSQL:
```bash
docker exec -it riadcloud-postgres psql -U riadcloud -d riadcloud
```

Then run:
```sql
SELECT * FROM services;
```

### Redis:
```bash
docker exec -it riadcloud-redis redis-cli
```

Then run:
```
PING
SET test "Hello RiadCloud"
GET test
```

## Useful Commands

### Check service status:
```bash
docker-compose ps
```

### Restart specific service:
```bash
docker-compose restart postgres
docker-compose restart redis
```

### View PostgreSQL logs:
```bash
docker-compose logs postgres
```

### Backup database:
```bash
docker exec riadcloud-postgres pg_dump -U riadcloud riadcloud > backup.sql
```

### Restore database:
```bash
docker exec -i riadcloud-postgres psql -U riadcloud riadcloud < backup.sql
```

## Production Notes

⚠️ **Important:** Change the default passwords before deploying to production!

Update in `docker-compose.yml`:
```yaml
POSTGRES_PASSWORD: your_secure_password_here
```

Also consider:
- Using Docker secrets for sensitive data
- Setting up SSL/TLS for PostgreSQL
- Configuring Redis authentication
- Setting up regular backups
- Using persistent volume mounts
