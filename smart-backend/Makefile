.PHONY: swagger docs swagger-serve run test build clean help

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@which swag >/dev/null 2>&1 && swag init -g cmd/main.go --output ./docs || go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go --output ./docs
	@echo "✅ Swagger docs generated at ./docs"


# Alias for swagger
docs: swagger

# Run the application
run:
	@echo "Starting Ololo Gate API..."
	@go run cmd/main.go

# Run with Docker Compose
docker-up:
	@echo "Starting Docker containers..."
	@docker compose up

# Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	@docker compose down

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -cover

# Build the application
build:
	@echo "Building Ololo Gate..."
	@go build -o bin/ololo-gate ./cmd/main.go
	@echo "✅ Build complete: bin/ololo-gate"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf docs/
	@echo "✅ Cleaned"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Help
help:
	@echo "Ololo Gate Backend - Available Commands:"
	@echo ""
	@echo "  make swagger         - Generate Swagger documentation"
	@echo "  make docs            - Alias for 'make swagger'"
	@echo "  make run             - Run the application locally"
	@echo "  make docker-up       - Start Docker containers"
	@echo "  make docker-down     - Stop Docker containers"
	@echo "  make test            - Run all tests"
	@echo "  make test-coverage   - Run tests with coverage"
	@echo "  make build           - Build the binary"
	@echo "  make clean           - Clean build artifacts and docs"
	@echo "  make deps            - Install/update dependencies"
	@echo "  make fmt             - Format code"
	@echo "  make lint            - Run linter"
	@echo "  make help            - Show this help message"
	@echo ""
