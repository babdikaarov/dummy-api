# Quick Start Guide

## Prerequisites
- Docker and Docker Compose installed
- Node.js 20+ (for local development)

## Development Mode

### Start the Project
```bash
npm run docker:dev:build
```

This will:
- Build and start PostgreSQL container
- Build and start API container
- Automatically generate database schema
- Automatically run migrations
- Automatically seed the database
- Start the app with hot reload on port 3000

### Access the Application
- **API**: http://localhost:3000
- **Swagger Docs**: http://localhost:3000/api/docs
- **API JSON**: http://localhost:3000/api-json
- **Database**: localhost:5432 (postgres/postgres)

### View Logs
```bash
docker-compose -f docker-compose.dev.yml logs -f api
```

### Stop the Project
```bash
npm run docker:down
```

---

## Production Mode

### Start the Project
```bash
npm run docker:prod:build
```

This will:
- Build and start PostgreSQL container
- Build and start optimized API container
- Automatically generate database schema
- Automatically run migrations
- Automatically seed the database
- Start the production app on port 3000

### Access the Application
- **API**: http://localhost:3000
- **Swagger Docs**: http://localhost:3000/api/docs

### View Logs
```bash
docker-compose -f docker-compose.prod.yml logs -f api
```

### Stop the Project
```bash
npm run docker:down
```

---

## Troubleshooting

### Container fails to start
```bash
# Check container logs
docker-compose -f docker-compose.dev.yml logs api

# If database connection fails, check if postgres is running
docker-compose -f docker-compose.dev.yml logs postgres
```

### Database issues
```bash
# Connect to database directly
psql -h localhost -U postgres -d gates_db

# Or execute commands inside the container
docker exec gates_api_dev npm run db:generate
docker exec gates_api_dev npm run db:migrate
docker exec gates_api_dev npm run db:seed
```

### Rebuild everything
```bash
# Remove all containers and volumes
npm run docker:down

# Rebuild from scratch
npm run docker:dev:build
```

### Manual database setup (if needed)
```bash
# Inside the running container
docker exec gates_api_dev bash -c "npm run db:generate && npm run db:migrate && npm run db:seed"
```

---

## Available Commands

| Command | Description |
|---------|-------------|
| `npm run docker:dev` | Start dev containers (must be already built) |
| `npm run docker:dev:build` | Build and start dev containers |
| `npm run docker:prod` | Start prod containers (must be already built) |
| `npm run docker:prod:build` | Build and start prod containers |
| `npm run docker:down` | Stop all containers |
| `npm run db:generate` | Generate database schema |
| `npm run db:migrate` | Run database migrations |
| `npm run db:seed` | Seed database with sample data |
| `npm run build` | Build the application |
| `npm run start:dev` | Start dev server locally (no Docker) |
| `npm run start:prod` | Start prod server locally (no Docker) |

---

## Environment Variables

### Development (.env.development)
```
NODE_ENV=development
PORT=3000
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gates_db
OLOLO_MOBILE_GATE_API_ORIGIN=*
```

### Production (.env.production)
```
NODE_ENV=production
PORT=3000
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gates_db
OLOLO_MOBILE_GATE_API_ORIGIN=http://host.docker.internal:8080
```

---

## Database

The project uses:
- **Database**: PostgreSQL 16
- **ORM**: Drizzle ORM
- **Migrations**: Drizzle Kit

Sample data includes:
- 20 locations (Russian-language place names)
- ~40 gates (with random distribution 1-3 per location)

---

## API Endpoints

- `GET /` - Health check
- `GET /locations` - Get all locations with gates
- `GET /locations/:id` - Get specific location with gates
- `POST /locations` - Create location
- `PUT /locations/:id` - Update location
- `DELETE /locations/:id` - Delete location
- `GET /api/docs` - Swagger documentation

---

## Notes

- Containers are named `gates_api_dev`/`gates_api_prod` and `gates_db_dev`/`gates_db_prod`
- Volume mounts in dev mode enable hot reload
- Production uses multi-stage Docker build for optimization
- All database setup is automated on container startup
