# Testing Database Setup

## Quick Test

After running `./deploy.sh dev`, the database should be automatically set up with migrations and initial data.

### Verify Setup

1. **Check containers are running:**
```bash
docker-compose ps
```

Expected output:
```
NAME                 STATUS          PORTS
gates_api_dev        Up (healthy)    0.0.0.0:3000->3000/tcp
gates_db_dev         Up (healthy)    0.0.0.0:5432->5432/tcp
ololo-gate-app-dev   Up (healthy)    0.0.0.0:8080->8080/tcp
ololo-gate-db-dev    Up (healthy)    0.0.0.0:5433->5432/tcp
```

2. **Check application logs:**
```bash
docker-compose logs dummy-api | grep -E "migrations|Seed|listening"
```

Expected to see:
```
Waiting for database to be ready...
Running migrations...
Seeding database with initial data...
Starting application with hot reload...
[Nest] X - XX/XX/XXXX, X:XX:XX PM LOG [NestFactory] Starting Nest application
[Nest] X - XX/XX/XXXX, X:XX:XX PM LOG [InstanceLoader] TypeOrmModule dependencies initialized
```

3. **Connect to database and verify tables:**
```bash
docker exec -it gates_db_dev psql -U postgres -d gates_db

# Inside psql:
\dt                    # List tables
SELECT * FROM users;   # Check if users table has data
\q                     # Exit
```

4. **Test API endpoint:**
```bash
curl http://localhost:3000

# Or with specific endpoint
curl http://localhost:3000/health
```

## Database Contents

After successful seed, you should have:

### Tables Created
- users
- roles
- permissions
- migrations
- (plus any others defined in your migrations)

### Initial Data
- Default admin user (credentials in .env.development)
- Default roles (admin, user, etc.)
- Default permissions

## Troubleshooting

### Migrations not running
```bash
# Check logs
docker-compose logs dummy-api

# Run manually
docker exec gates_api_dev npx drizzle-kit migrate
```

### Database empty
```bash
# Reconnect to database
docker exec -it gates_db_dev psql -U postgres -d gates_db

# Check migrations table
SELECT * FROM drizzle_migrations;

# Manually seed if needed
docker exec gates_api_dev npm run db:seed
```

### Connection refused
```bash
# Wait for database to be ready
sleep 5

# Check database health
docker logs gates_db_dev

# Restart if needed
docker-compose restart dummy-postgres-dev
```

### Application not starting
```bash
# Check full logs
docker-compose logs dummy-api -f

# Check for errors
docker-compose logs dummy-api | grep -i error
```

## Production Test

For production testing:

```bash
./deploy.sh prod rebuild

# Verify
docker-compose ps
docker-compose logs dummy-api

# Test endpoint
curl http://localhost:3000

# Connect to database
docker exec dummy-postgres-prod psql -U postgres -d gates_db
```

## Database Persistence

- Development: Data persists in `postgres_dev_data` volume
- Production: Data persists in `dummy_postgres_data` volume

To start fresh:
```bash
# Development
docker-compose down -v
./deploy.sh dev

# Production
docker-compose down -v
./deploy.sh prod rebuild
```

## Automated Testing Script

```bash
#!/bin/bash

echo "Testing database setup..."

# Check containers
echo "✓ Checking containers..."
docker-compose ps | grep -q "gates_api_dev" && echo "  API container: OK" || echo "  API container: FAILED"

# Check database
echo "✓ Testing database connection..."
docker exec gates_db_dev pg_isready -U postgres && echo "  Database: OK" || echo "  Database: FAILED"

# Check migrations
echo "✓ Checking migrations..."
MIGRATIONS=$(docker exec gates_db_dev psql -U postgres -d gates_db -t -c "SELECT COUNT(*) FROM drizzle_migrations;")
echo "  Migrations applied: $MIGRATIONS"

# Check tables
echo "✓ Checking tables..."
TABLES=$(docker exec gates_db_dev psql -U postgres -d gates_db -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';")
echo "  Tables created: $TABLES"

# Check API
echo "✓ Testing API..."
curl -s http://localhost:3000 > /dev/null && echo "  API: OK" || echo "  API: FAILED"

echo ""
echo "Database setup test complete!"
```

Save as `test-db.sh`, make executable with `chmod +x test-db.sh`, and run with `./test-db.sh`.
