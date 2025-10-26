# Quick Reference Guide

## Start Services

```bash
# Development (hot reload enabled)
./deploy.sh dev

# Production (optimized)
./deploy.sh prod

# With rebuild
./deploy.sh dev rebuild
./deploy.sh prod rebuild
```

## Check Status

```bash
# All containers
docker ps -a

# Services only
docker-compose ps

# View logs
docker-compose logs -f

# Specific service
docker-compose logs -f dummy-api
docker-compose logs -f ololo-api
```

## Stop Services

```bash
# Stop all
docker-compose down

# Stop specific service
docker-compose stop dummy-api

# Remove volumes (warning: data loss!)
docker-compose down -v
```

## Database Access

```bash
# Dummy API database (Dev)
docker exec -it dummy-postgres-dev psql -U postgres -d gates_db

# Ololo database (Dev)
docker exec -it ololo-postgres-dev psql -U postgres -d ololo_gate

# List databases
\l

# List tables
\dt

# Exit psql
\q
```

## Rebuild Images

```bash
# Rebuild all
docker-compose build

# Rebuild specific service
docker-compose build dummy-api
docker-compose build ololo-api

# No cache
docker-compose build --no-cache
```

## Service Details

| Service | Port | Type | Status |
|---------|------|------|--------|
| Dummy API | 3000 | Node.js/NestJS | Running |
| Dummy DB | 5432 | PostgreSQL | Running |
| Ololo API | 8080 | Go | Running |
| Ololo DB | 5433 | PostgreSQL | Running |

## Useful Commands

```bash
# View resource usage
docker stats

# View network
docker network inspect backend-network

# View volume
docker volume inspect backend_dummy_postgres_data

# Remove dangling images
docker image prune

# Remove unused everything
docker system prune
```

## Troubleshooting

### Services won't start
```bash
# Check logs
docker-compose logs

# Check ports
lsof -i :3000
lsof -i :8080
```

### Database won't connect
```bash
# Check database container
docker ps | grep postgres

# Check database logs
docker logs [container_id]
```

### Hot reload not working
- Dummy API: Check /app has write permissions
- Ololo API: Check .air.toml configuration

### Port in use
- Change port in docker-compose file
- Or stop the service using that port

## Environment Variables

### Production (⚠️ Must customize before deployment)

**ololo-backend/.env.production:**
- DB_PASSWORD
- JWT_SECRET
- INIT_ADMIN_PASSWORD
- CORS_ALLOWED_ORIGINS
- THIRD_PARTY_API_URL

**dummy-backend-api/.env.production:**
- DB_PASSWORD
- All database credentials

## File Locations

```
Root:                ./deploy.sh
Production compose:  ./docker-compose.yml
Development compose: ./docker-compose.dev.yml
Docs:               ./DEPLOYMENT.md
                   ./README_SETUP.md
                   ./SETUP_SUMMARY.txt
```

## API Endpoints

- **Dummy API:** http://localhost:3000
- **Ololo API:** http://localhost:8080

## Documentation

| File | Purpose |
|------|---------|
| README_SETUP.md | Quick start guide |
| DEPLOYMENT.md | Full deployment guide |
| SETUP_SUMMARY.txt | Setup overview |
| QUICK_REFERENCE.md | This file |

## Emergency Cleanup

```bash
# Stop and remove everything
docker-compose down -v

# Remove all containers
docker container prune -f

# Remove all images
docker image prune -f

# Remove all volumes
docker volume prune -f

# Full reset
docker system prune -f --volumes
```

⚠️ Warning: These commands delete data. Use with caution!
