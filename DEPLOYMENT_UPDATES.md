# Deployment Script & Docker Compose Updates

## Overview
The backend deployment system has been refactored to use **shared Docker Compose files** that consolidate both dev and prod environments with a single PostgreSQL database per environment.

## Changes Made

### 1. Docker Compose Files

#### `docker-compose.shared.dev.yml` (Development)
- **Single PostgreSQL Instance**: `deploy-postgres-dev`
- **Port Mapping**: Database on `5434:5432`
- **Shared Configuration**: Uses `.env.deploy` at root level
- **Services**:
  - `dummy-api-dev`: Smart Dummy Backend (Node.js) - Port 3000
  - `smart-api-dev`: Smart Backend (Go) - Port 8080
  - `deploy-postgres-dev`: PostgreSQL - Port 5434
- **Features**:
  - Hot reload enabled (volumes mounted)
  - Development commands (`npm run start:dev`, `air -c .air.toml`)
  - Network: `backend-network`

#### `docker-compose.shared.yml` (Production)
- **Single PostgreSQL Instance**: `deploy-postgres-prod`
- **Port Mapping**: Database on `5435:5432`
- **Shared Configuration**: Uses `.env.deploy` at root level
- **Services**:
  - `dummy-api-prod`: Smart Dummy Backend (Node.js) - Port 3000
  - `smart-api-prod`: Smart Backend (Go) - Port 8080
  - `deploy-postgres-prod`: PostgreSQL - Port 5435
- **Features**:
  - Restart policies: `unless-stopped`
  - Health checks for all services (30s interval for APIs, 10s for DB)
  - Network: `backend-network`

### 2. Environment Configuration

#### `.env.deploy` (Root Level - Shared)
- **Single source of truth** for both dev and prod environments
- Contains:
  - `DB_HOST=deploy-postgres`
  - `DB_PORT=5432`
  - `DB_USER=postgres`
  - `DB_PASSWORD=postgres`
  - `DB_NAME=deploy_db`
  - API and service configurations
  - JWT settings
  - CORS and third-party service URLs

### 3. Deploy Script Updates (`./deploy.sh`)

#### Validation
- Added check for root-level `.env.deploy` file
- Validates project-specific environment files in `smart-dummy/` and `smart-backend/`

#### Service Information Display
**Development Mode**:
```
Shared PostgreSQL Database
  Container: deploy-postgres-dev
  Host: deploy-postgres
  Port: 5434 (localhost)
  Connection: postgres://postgres:postgres@deploy-postgres:5432/deploy_db

Smart Dummy Backend (Node.js)
  URL: http://localhost:3000
  Container: dummy-api-dev
  Hot reload: Yes

Smart Backend (Go)
  URL: http://localhost:8080
  Container: smart-api-dev
  Hot reload: Yes (Air)
```

**Production Mode**:
```
Shared PostgreSQL Database
  Container: deploy-postgres-prod
  Host: deploy-postgres
  Port: 5435 (localhost)
  Connection: postgres://postgres:postgres@deploy-postgres:5432/deploy_db

Smart Dummy Backend (Node.js)
  URL: http://localhost:3000
  Container: dummy-api-prod

Smart Backend (Go)
  URL: http://localhost:8080
  Container: smart-api-prod
```

#### Container Status
- Updated filter to show: `dummy-api`, `smart-api`, and `deploy-postgres` containers
- Displays names, status, and port mappings

#### Helper Commands
- Dynamically shows correct docker-compose file paths based on mode
- Examples:
  - Dev: `docker-compose -f docker-compose.shared.dev.yml logs -f`
  - Prod: `docker-compose -f docker-compose.shared.yml logs -f`

## Usage

### Development Deployment
```bash
./deploy.sh dev              # Start with cached images
./deploy.sh dev rebuild      # Start with image rebuild
```

### Production Deployment
```bash
./deploy.sh              # Start with cached images (default)
./deploy.sh prod rebuild # Start with image rebuild
```

### Help
```bash
./deploy.sh --help
```

## Database Setup

### Development
- **Container**: `deploy-postgres-dev`
- **Accessible from host**: `localhost:5434`
- **Accessible from containers**: `deploy-postgres:5432`
- **Database**: `deploy_db`
- **Volume**: `deploy_postgres_dev_data`

### Production
- **Container**: `deploy-postgres-prod`
- **Accessible from host**: `localhost:5435`
- **Accessible from containers**: `deploy-postgres:5432`
- **Database**: `deploy_db`
- **Volume**: `deploy_postgres_prod_data`

## Networking

Both environments use the `backend-network` bridge network for inter-container communication.

## Key Benefits

✅ **Simplified Management**: Single postgres instance per environment instead of separate databases
✅ **Unified Configuration**: `.env.deploy` file serves both services
✅ **Consistent Structure**: Dev and prod follow identical patterns
✅ **Better Scalability**: Easy to add more services using the same network
✅ **Clear Information**: Deploy script provides comprehensive service information
✅ **Port Isolation**: Dev (5434) and Prod (5435) PostgreSQL ports don't conflict

## Migration Notes

- Old separate docker-compose files (`docker-compose.dev.yml`, `docker-compose.prod.yml`) in individual project folders are no longer used
- New shared compose files at root level: `docker-compose.shared.dev.yml`, `docker-compose.shared.yml`
- All configuration now centralized in root `.env.deploy`
