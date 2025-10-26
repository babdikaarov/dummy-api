#!/bin/bash

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}   Production Container Setup Script${NC}"
echo -e "${BLUE}================================================${NC}\n"

# Check if .env.production exists
if [ ! -f ".env.production" ]; then
    echo -e "${RED}❌ Error: .env.production file not found!${NC}"
    echo -e "${YELLOW}Please create .env.production with the following variables:${NC}"
    echo "  DB_HOST=postgres"
    echo "  DB_PORT=5432"
    echo "  DB_USER=postgres"
    echo "  DB_PASSWORD=postgres"
    echo "  DB_NAME=gates_db"
    exit 1
fi

echo -e "${GREEN}✅ Found .env.production${NC}\n"

# Step 1: Build and start containers
echo -e "${BLUE}[Step 1/3]${NC} Building and starting Docker containers..."
docker-compose -f docker-compose.prod.yml up --build -d

# Wait for database to be ready
echo -e "\n${YELLOW}⏳ Waiting for database to be ready...${NC}"
sleep 15

# Step 2: Run migrations
echo -e "\n${BLUE}[Step 2/3]${NC} Running database migrations..."
docker exec gates_api_prod npm run db:migrate
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Migrations completed successfully${NC}"
else
    echo -e "${RED}❌ Migration failed!${NC}"
    exit 1
fi

# Step 3: Seed database (optional)
echo -e "\n${BLUE}[Step 3/3]${NC} Seeding database with sample data..."
echo -e "${YELLOW}ℹ️  Attempting to seed from localhost...${NC}"
npm run db:seed 2>&1 | grep -q "Seed completed" && {
    echo -e "${GREEN}✅ Database seeded successfully${NC}"
} || {
    echo -e "${YELLOW}⚠️  Seeding not critical - you can manually run:${NC}"
    echo -e "   ${BLUE}npm run db:seed${NC}"
}

# Verification
echo -e "\n${BLUE}================================================${NC}"
echo -e "${GREEN}✅ Production setup completed successfully!${NC}"
echo -e "${BLUE}================================================${NC}\n"

echo -e "${YELLOW}📊 System Status:${NC}"
echo -e "  🐳 API Container: $(docker inspect -f '{{.State.Status}}' gates_api_prod)"
echo -e "  🗄️  Database Container: $(docker inspect -f '{{.State.Status}}' gates_db_prod)"
echo -e "  🌐 API URL: ${GREEN}http://localhost:3000${NC}"
echo -e "  📚 Swagger Docs: ${GREEN}http://localhost:3000/api/docs${NC}"
echo ""

# Test API
echo -e "${YELLOW}🧪 Testing API endpoint...${NC}"
RESPONSE=$(curl -s http://localhost:3000 || echo "")
if [ -z "$RESPONSE" ]; then
    echo -e "${RED}⚠️  Could not connect to API. It may still be starting...${NC}"
else
    echo -e "${GREEN}✅ API is responding${NC}"
    echo -e "${BLUE}Response: ${NC}$RESPONSE\n"
fi

echo -e "${YELLOW}📝 Next Steps:${NC}"
echo "  1. Visit API: http://localhost:3000/locations"
echo "  2. View Swagger Docs: http://localhost:3000/api/docs"
echo "  3. Stop containers: npm run docker:down"
echo "  4. Start containers (next time): npm run docker:prod"
echo ""
