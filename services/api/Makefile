.PHONY: help build run test clean docker-build docker-run dev

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the API server
	@echo "Building API server..."
	go build -o bin/api ./cmd/api

run: ## Run the API server
	@echo "Running API server..."
	go run ./cmd/api

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t house-helper/api:latest .

docker-run: ## Run in Docker container
	@echo "Running in Docker container..."
	docker run --rm -p 8080:8080 house-helper/api:latest

dev: ## Run in development mode with hot reload
	@echo "Running in development mode..."
	go run ./cmd/api

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

migrate-up: ## Run database migrations up
	@echo "Running database migrations..."
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Run database migrations down
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir migrations $(NAME)

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs/

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install golang.org/x/tools/cmd/goimports@latest