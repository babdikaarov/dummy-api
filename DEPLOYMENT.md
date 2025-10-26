# Backend Deployment Guide

This guide covers running both the Dummy Backend API (Node.js/NestJS) and Ololo Backend (Go) projects together using Docker Compose.

## Project Structure

```
backend/
├── dummy-backend-api/          # Node.js/NestJS project
│   ├── Dockerfile.dev
│   ├── Dockerfile.prod
│   ├── docker-compose.dev.yml
│   ├── docker-compose.prod.yml
│   ├── .env.development
│   ├── .env.production
│   └── ... (other project files)
│
├── ololo-backend/              # Go project
│   ├── Dockerfile              # Dev stage
│   ├── Dockerfile.prod         # Production stage
│   ├── docker-compose.dev.yml
│   ├── docker-compose.prod.yml
│   ├── .env.development
│   ├── .env.production
│   └── ... (other project files)
│
├── docker-compose.yml          # Root production compose file
├── docker-compose.dev.yml      # Root development compose file
├── deploy.sh                   # Deployment automation script
└── DEPLOYMENT.md              # This file
```

## Quick Start

### Prerequisites

- Docker Desktop or Docker Engine (version 20.10+)
- Docker Compose (version 1.29+)
- Git
- Bash (for running the deployment script)

### Running Both Projects

#### Development Mode (with hot reload)

```bash
# Navigate to the backend directory
cd backend

# Run both projects with hot reload
./deploy.sh dev

# Or with image rebuild
./deploy.sh dev rebuild
```

#### Production Mode

```bash
# Navigate to the backend directory
cd backend

# Run both projects in production
./deploy.sh prod

# Or with image rebuild
./deploy.sh prod rebuild
```

### Manual Docker Compose Commands

If you prefer to run Docker Compose directly:

**Development:**
```bash
docker-compose -f docker-compose.dev.yml up -d
docker-compose -f docker-compose.dev.yml logs -f
```

**Production:**
```bash
docker-compose -f docker-compose.yml up -d
docker-compose -f docker-compose.yml logs -f
```

**Stop services:**
```bash
docker-compose -f docker-compose.yml down
# or for dev
docker-compose -f docker-compose.dev.yml down
```

## Service Endpoints

### Development Environment

| Service | Port | URL |
|---------|------|-----|
| Dummy API | 3000 | http://localhost:3000 |
| Dummy Database | 5432 | postgres://localhost:5432/gates_db |
| Ololo API | 8080 | http://localhost:8080 |
| Ololo Database | 5433 | postgres://localhost:5433/ololo_gate |

### Production Environment

| Service | Port | URL |
|---------|------|-----|
| Dummy API | 3000 | http://localhost:3000 |
| Ololo API | 8080 | http://localhost:8080 |

## Environment Configuration

### Dummy Backend API

**Development** (`dummy-backend-api/.env.development`):
- Default settings for local development
- Database credentials: `postgres/postgres`

**Production** (`dummy-backend-api/.env.production`):
- Optimized for production
- Update database credentials before deployment

### Ololo Backend

**Development** (`ololo-backend/.env.development`):
- Database: PostgreSQL (local container)
- Hot reload: Enabled (using Air)
- Debug features: Enabled
- CORS: Allow all origins

**Production** (`ololo-backend/.env.production`):
- **⚠️ IMPORTANT:** Must update before production deployment:
  - `DB_PASSWORD` - Change to secure password
  - `JWT_SECRET` - Change to secure secret key
  - `INIT_ADMIN_PASSWORD` - Change admin password
  - `CORS_ALLOWED_ORIGINS` - Set to your domain
  - `THIRD_PARTY_API_URL` - Set to production API endpoint

## Deployment on Remote Server

### 1. Clone Repository
```bash
git clone <your-repo-url> backend
cd backend
```

### 2. Update Production Environment Variables

```bash
# Edit ololo production config
vi ololo-backend/.env.production

# Edit dummy production config
vi dummy-backend-api/.env.production
```

**Critical variables to update:**

```env
# ololo-backend/.env.production
DB_PASSWORD=your-secure-password-here
JWT_SECRET=your-secure-jwt-secret-here
INIT_ADMIN_PASSWORD=your-secure-admin-password
CORS_ALLOWED_ORIGINS=https://yourdomain.com
THIRD_PARTY_API_URL=https://api.yourdomain.com:3000

# dummy-backend-api/.env.production
DB_PASSWORD=your-secure-password-here
DB_HOST=postgres  # or your actual postgres host
```

### 3. Start Services
```bash
./deploy.sh prod rebuild
```

### 4. Verify Services
```bash
docker ps
docker-compose logs -f
```

### 5. Update DNS/Firewall

- Configure your reverse proxy (Nginx/Traefik) to route traffic to ports 3000 and 8080
- Update firewall rules to allow incoming traffic
- Configure SSL/TLS certificates

## Docker Network

Both development and production setups use a shared Docker network (`backend-network` or `ololo-network`):

- Services can communicate using container names as hostnames
- Example: `http://dummy-api:3000` from ololo container

## Database Management

### Connect to Database

**Development:**

```bash
# Dummy database
docker exec -it dummy-postgres-dev psql -U postgres -d gates_db

# Ololo database
docker exec -it ololo-postgres-dev psql -U postgres -d ololo_gate
```

**Production:**

```bash
# Dummy database
docker exec -it dummy-postgres-prod psql -U postgres -d gates_db

# Ololo database
docker exec -it ololo-postgres-prod psql -U postgres -d ololo_gate
```

### Backup Databases

```bash
# Backup dummy database
docker exec dummy-postgres-prod pg_dump -U postgres -d gates_db > dummy_backup.sql

# Backup ololo database
docker exec ololo-postgres-prod pg_dump -U postgres -d ololo_gate > ololo_backup.sql
```

## Troubleshooting

### Services won't start

1. **Check Docker is running:**
   ```bash
   docker ps
   ```

2. **Check logs:**
   ```bash
   docker-compose logs [service_name]
   ```

3. **Network issues:**
   ```bash
   docker network inspect backend-network
   ```

### Port conflicts

If ports 3000, 5432, 5433, or 8080 are already in use:

1. Find the process using the port:
   ```bash
   lsof -i :3000  # or 5432, 5433, 8080
   ```

2. Stop it or use different ports in the docker-compose files

### Database connection errors

1. Ensure databases are healthy:
   ```bash
   docker ps
   # Check STATUS column, should show "Up"
   ```

2. Wait for database to be ready:
   ```bash
   docker-compose logs db
   # Look for "database system is ready to accept connections"
   ```

### Hot reload not working (development)

1. **Dummy API:** Uses nodemon/nest watch - file changes should auto-reload
2. **Ololo API:** Uses Air - ensure `.air.toml` is present and configured

### Memory/Resource issues

Increase Docker resource limits:
- Docker Desktop: Settings → Resources
- Production: Use cgroups memory limits

## Monitoring

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f dummy-api
docker-compose logs -f ololo-api

# Last 100 lines
docker-compose logs --tail 100
```

### Health Checks

Services include health checks that Docker monitors:

```bash
# Check health
docker inspect dummy-api-dev --format='{{.State.Health}}'
```

### Resource Usage

```bash
docker stats
```

## Security Considerations

1. **Never commit `.env.production`** with real secrets to git
2. **Use `.gitignore`** to exclude sensitive files
3. **Rotate secrets regularly** in production
4. **Use strong passwords** for database and JWT secrets
5. **Enable TLS/SSL** for all external connections
6. **Keep Docker images updated** with security patches
7. **Use read-only filesystems** where possible

## Scaling and Advanced Configuration

### Load Balancing

To run multiple instances behind a load balancer:

```yaml
services:
  dummy-api-1:
    # ... config
    ports: []  # Don't expose ports
  dummy-api-2:
    # ... config
    ports: []  # Use load balancer instead

  nginx:
    image: nginx:latest
    ports:
      - "3000:3000"
    # ... configure upstream servers
```

### Persistent Volumes

Databases already use persistent volumes:
- `dummy_postgres_data` - Dummy database storage
- `ololo_postgres_data` - Ololo database storage

To backup volumes:
```bash
docker run --rm -v dummy_postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/dummy_data.tar.gz -C / data
```

## Support & Issues

For issues:

1. Check logs: `docker-compose logs -f`
2. Verify `.env` files are correct
3. Ensure Docker/Compose are updated
4. Check disk space: `docker system df`
5. Clean up: `docker system prune` (carefully!)

## Next Steps

1. Deploy to staging environment first
2. Test all endpoints
3. Configure monitoring (Prometheus, Grafana)
4. Set up automated backups
5. Document your deployment specifics
