# Backend Projects Setup - Complete

## What Was Created

This setup allows you to run both the Dummy Backend API and Ololo Backend from a single root directory using Docker Compose.

### Files Created/Modified

#### Root Directory (`backend/`)

1. **`docker-compose.yml`**
   - Production setup for both projects
   - Includes health checks and restart policies
   - Services communicate via `backend-network`

2. **`docker-compose.dev.yml`**
   - Development setup with hot reload enabled
   - Volume mounts for live code changes
   - Local database access on different ports

3. **`deploy.sh`**
   - Automated deployment script
   - Validates environment, creates networks, starts services
   - Supports dev/prod modes with optional rebuild
   - Usage: `./deploy.sh [dev|prod] [rebuild|no-rebuild]`

4. **`DEPLOYMENT.md`**
   - Complete deployment guide
   - Troubleshooting tips
   - Security recommendations
   - Database management commands

#### Ololo Backend (`ololo-backend/`)

1. **`Dockerfile.prod`**
   - Multi-stage production build
   - Optimized binary with stripped symbols
   - Uses distroless base image for minimal size

2. **`docker-compose.dev.yml`**
   - Development-only compose file for ololo
   - Hot reload with Air
   - Separate network from other services (optional)

3. **`docker-compose.prod.yml`**
   - Production-only compose file
   - Restart policies and health checks
   - Secure PostgreSQL configuration

4. **`.env.development`**
   - Development environment variables
   - Credentials safe for local development

5. **`.env.production`**
   - Production environment template
   - **⚠️ Must be customized before deployment**
   - All sensitive values marked with placeholders

## Quick Start

### For Local Development

```bash
cd backend
./deploy.sh dev
```

This will start:
- Dummy API on `http://localhost:3000` (hot reload enabled)
- Ololo API on `http://localhost:8080` (hot reload enabled)
- Both PostgreSQL databases

### For Production Deployment

```bash
cd backend

# Edit environment files
vi ololo-backend/.env.production
vi dummy-backend-api/.env.production

# Start services
./deploy.sh prod rebuild
```

## Service Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Network                           │
│              (backend-network / ololo-network)              │
│                                                               │
│  ┌──────────────────┐              ┌──────────────────┐    │
│  │  Dummy API       │              │   Ololo API      │    │
│  │  Port: 3000      │              │   Port: 8080     │    │
│  │  (Node.js)       │              │   (Go)           │    │
│  └────────┬─────────┘              └────────┬─────────┘    │
│           │                                 │               │
│  ┌────────▼─────────┐              ┌────────▼─────────┐    │
│  │ Dummy PostgreSQL │              │ Ololo PostgreSQL │    │
│  │ Port: 5432/5433  │              │ Port: 5432/5433  │    │
│  └──────────────────┘              └──────────────────┘    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Environment Variables

### Development (`localhost`)

| Project | API Port | DB Port | DB Name |
|---------|----------|---------|---------|
| Dummy | 3000 | 5432 | gates_db |
| Ololo | 8080 | 5433 | ololo_gate |

### Production (Docker containers)

Both projects use container-to-container networking:
- Services connect using container names as hostnames
- Database passwords must be changed in `.env.production`

## Commands

### Deployment Script

```bash
# Development with rebuild
./deploy.sh dev rebuild

# Development without rebuild
./deploy.sh dev

# Production with rebuild
./deploy.sh prod rebuild

# Production without rebuild
./deploy.sh prod

# Show help
./deploy.sh --help
```

### Docker Compose Direct

```bash
# Start services
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Stop services
docker-compose -f docker-compose.dev.yml down

# Rebuild images
docker-compose -f docker-compose.dev.yml build

# View container status
docker-compose -f docker-compose.dev.yml ps
```

## Security Checklist for Production

Before deploying to production:

- [ ] Change `DB_PASSWORD` in both projects
- [ ] Change `JWT_SECRET` in ololo-backend
- [ ] Change `INIT_ADMIN_PASSWORD` in ololo-backend
- [ ] Set `CORS_ALLOWED_ORIGINS` to your domain
- [ ] Set `THIRD_PARTY_API_URL` to production endpoint
- [ ] Add `.env.production` to `.gitignore`
- [ ] Configure TLS/SSL certificates
- [ ] Set up database backups
- [ ] Configure monitoring/logging

## Troubleshooting

### Services won't start

1. Verify Docker is running:
   ```bash
   docker ps
   ```

2. Check logs:
   ```bash
   docker-compose logs -f
   ```

3. Verify port availability:
   ```bash
   lsof -i :3000  # or other ports
   ```

### Database connection errors

1. Wait for databases to be ready:
   ```bash
   docker-compose logs db
   # Look for "ready to accept connections"
   ```

2. Check database credentials in `.env` files match docker-compose files

### Hot reload not working

For **Dummy API**: File changes should auto-reload (needs ~2-3 seconds)
For **Ololo API**: File changes trigger Air to rebuild (check `.air.toml`)

### Port conflicts

Edit `docker-compose.yml` or `docker-compose.dev.yml` to use different ports:

```yaml
ports:
  - "3001:3000"  # Map external 3001 to internal 3000
```

## Next Steps

1. **Read the full guide:** See `DEPLOYMENT.md` for comprehensive documentation
2. **Test locally:** Run `./deploy.sh dev` and verify both services work
3. **Customize configs:** Update `.env` files with your values
4. **Deploy to remote:** Follow "Deployment on Remote Server" in `DEPLOYMENT.md`
5. **Monitor services:** Set up monitoring tools (optional but recommended)

## File Structure Reference

```
backend/
├── docker-compose.yml              # Root prod config
├── docker-compose.dev.yml          # Root dev config
├── deploy.sh                       # Deployment script
├── DEPLOYMENT.md                   # Full deployment guide
├── README_SETUP.md                 # This file
│
├── dummy-backend-api/
│   ├── Dockerfile.dev
│   ├── Dockerfile.prod
│   ├── docker-compose.dev.yml
│   ├── docker-compose.prod.yml
│   ├── .env.development
│   ├── .env.production
│   └── ...
│
└── ololo-backend/
    ├── Dockerfile
    ├── Dockerfile.prod
    ├── docker-compose.dev.yml
    ├── docker-compose.prod.yml
    ├── .env
    ├── .env.development
    ├── .env.production
    ├── .env.example
    └── ...
```

## Support

For detailed information on:
- **Deployment:** See `DEPLOYMENT.md`
- **Troubleshooting:** Section in `DEPLOYMENT.md`
- **Database management:** Section in `DEPLOYMENT.md`
- **Security:** Section in `DEPLOYMENT.md`

---

**Created:** 2025-10-26
**Setup includes:** Both Dummy Backend API and Ololo Backend with dev/prod configurations
