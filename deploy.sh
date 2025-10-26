#!/bin/bash

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Configuration
DEV_MODE=${1:-"prod"}
REBUILD=${2:-"no-rebuild"}

# Functions
print_header() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

show_usage() {
    echo -e "${BLUE}Usage: ./deploy.sh [MODE] [BUILD]${NC}"
    echo ""
    echo "Modes:"
    echo "  dev   - Run in development mode with hot reload"
    echo "  prod  - Run in production mode (default)"
    echo ""
    echo "Build options:"
    echo "  rebuild     - Rebuild all Docker images"
    echo "  no-rebuild  - Use cached images (default)"
    echo ""
    echo "Examples:"
    echo "  ./deploy.sh              # Production with cached images"
    echo "  ./deploy.sh prod rebuild # Production with rebuild"
    echo "  ./deploy.sh dev          # Development with hot reload"
    echo "  ./deploy.sh dev rebuild  # Development with rebuild"
}

check_requirements() {
    print_header "Checking Requirements"

    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi
    print_success "Docker found"

    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi
    print_success "Docker Compose found"

    if ! command -v git &> /dev/null; then
        print_error "Git is not installed"
        exit 1
    fi
    print_success "Git found"
}

validate_env_files() {
    print_header "Validating Environment Files"

    # Check root .env.deploy file (shared for all services)
    if [ ! -f "$SCRIPT_DIR/.env.deploy" ]; then
        print_error "Missing .env.deploy at root"
        exit 1
    fi
    print_success ".env.deploy found (shared configuration)"

    # Check dummy-backend environment
    if [ ! -f "$SCRIPT_DIR/smart-dummy/.env.development" ]; then
        print_error "Missing smart-dummy/.env.development"
        exit 1
    fi
    print_success "smart-dummy/.env.development found"

    if [ ! -f "$SCRIPT_DIR/smart-dummy/.env.production" ]; then
        print_error "Missing smart-dummy/.env.production"
        exit 1
    fi
    print_success "smart-dummy/.env.production found"

    # Check smart-backend environment
    if [ ! -f "$SCRIPT_DIR/smart-backend/.env.development" ]; then
        print_warning "Missing smart-backend/.env.development - creating from example"
        if [ -f "$SCRIPT_DIR/smart-backend/.env.example" ]; then
            cp "$SCRIPT_DIR/smart-backend/.env.example" "$SCRIPT_DIR/smart-backend/.env.development"
            print_success "Created .env.development from .env.example"
        fi
    else
        print_success "smart-backend/.env.development found"
    fi

    if [ ! -f "$SCRIPT_DIR/smart-backend/.env.production" ]; then
        print_warning "Missing smart-backend/.env.production - creating from example"
        if [ -f "$SCRIPT_DIR/smart-backend/.env.example" ]; then
            cp "$SCRIPT_DIR/smart-backend/.env.example" "$SCRIPT_DIR/smart-backend/.env.production"
            # Update production settings
            sed -i '' 's/ENV=development/ENV=production/g' "$SCRIPT_DIR/smart-backend/.env.production"
            print_success "Created .env.production from .env.example"
        fi
    else
        print_success "smart-backend/.env.production found"
    fi
}

create_network() {
    print_header "Setting Up Docker Network"

    if ! docker network inspect backend-network &> /dev/null; then
        if ! docker network ls | grep -q "backend-network"; then
            docker network create backend-network
            print_success "Created backend-network"
        fi
    else
        print_success "Network backend-network already exists"
    fi
}

start_services() {
    print_header "Starting Services"

    local compose_file="docker-compose.shared.yml"
    local mode_name="PRODUCTION"

    if [ "$DEV_MODE" == "dev" ]; then
        compose_file="docker-compose.shared.dev.yml"
        mode_name="DEVELOPMENT"
    fi

    echo -e "${BLUE}Mode: $mode_name${NC}"
    echo -e "${BLUE}Compose file: $compose_file${NC}\n"

    if [ "$REBUILD" == "rebuild" ]; then
        echo -e "${YELLOW}Building images...${NC}"
        docker-compose -f "$SCRIPT_DIR/$compose_file" build
        if [ $? -ne 0 ]; then
            print_error "Build failed"
            exit 1
        fi
    fi

    docker-compose -f "$SCRIPT_DIR/$compose_file" up -d

    if [ $? -eq 0 ]; then
        print_success "Services started successfully"
    else
        print_error "Failed to start services"
        exit 1
    fi
}

show_service_info() {
    print_header "Service Information"

    if [ "$DEV_MODE" == "dev" ]; then
        echo -e "${BLUE}Development Environment${NC}"
        echo ""
        echo -e "${GREEN}Shared PostgreSQL Database${NC}"
        echo "  Container: deploy-postgres-dev"
        echo "  Host: deploy-postgres"
        echo "  Port: 5434 (localhost)"
        echo "  Connection: postgres://postgres:postgres@deploy-postgres:5432/deploy_db"
        echo ""
        echo -e "${GREEN}Smart Dummy Backend (Node.js)${NC}"
        echo "  URL: http://localhost:3000"
        echo "  Container: dummy-api-dev"
        echo "  Hot reload: Yes"
        echo ""
        echo -e "${GREEN}Smart Backend (Go)${NC}"
        echo "  URL: http://localhost:8080"
        echo "  Container: smart-api-dev"
        echo "  Hot reload: Yes (Air)"
        echo ""
    else
        echo -e "${BLUE}Production Environment${NC}"
        echo ""
        echo -e "${GREEN}Shared PostgreSQL Database${NC}"
        echo "  Container: deploy-postgres-prod"
        echo "  Host: deploy-postgres"
        echo "  Port: 5435 (localhost)"
        echo "  Connection: postgres://postgres:postgres@deploy-postgres:5432/deploy_db"
        echo ""
        echo -e "${GREEN}Smart Dummy Backend (Node.js)${NC}"
        echo "  URL: http://localhost:3000"
        echo "  Container: dummy-api-prod"
        echo ""
        echo -e "${GREEN}Smart Backend (Go)${NC}"
        echo "  URL: http://localhost:8080"
        echo "  Container: smart-api-prod"
        echo ""
    fi
}

show_status() {
    print_header "Container Status"
    docker ps --filter "name=dummy-api\|smart-api\|deploy-postgres" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
}

# Main script
if [ "$1" == "--help" ] || [ "$1" == "-h" ] || [ "$1" == "help" ]; then
    show_usage
    exit 0
fi

print_header "Backend Deployment Script"

check_requirements
validate_env_files
create_network
start_services
show_service_info
show_status

echo ""
print_success "Deployment complete!"
echo ""
echo -e "${BLUE}Useful commands:${NC}"

if [ "$DEV_MODE" == "dev" ]; then
    echo "  View logs:     docker-compose -f docker-compose.shared.dev.yml logs -f"
    echo "  Stop services: docker-compose -f docker-compose.shared.dev.yml down"
else
    echo "  View logs:     docker-compose -f docker-compose.shared.yml logs -f"
    echo "  Stop services: docker-compose -f docker-compose.shared.yml down"
fi

echo "  Rebuild:       ./deploy.sh $DEV_MODE rebuild"
echo ""
