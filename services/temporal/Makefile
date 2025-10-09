.PHONY: help install test worker api docker-build docker-push clean

# Variables
WORKER_IMAGE := househelper/temporal-worker
API_IMAGE := househelper/temporal-api
VERSION ?= latest

help: ## Display this help message
	@echo "Temporal Service - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

install: ## Install dependencies
	go mod download
	go mod tidy

test: ## Run tests
	go test ./internal/workflows/... -v -cover

test-coverage: ## Run tests with coverage report
	go test ./internal/workflows/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

worker: ## Run worker locally
	go run cmd/worker/main.go

api: ## Run API server locally
	go run cmd/api/main.go

build-worker: ## Build worker binary
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/worker cmd/worker/main.go

build-api: ## Build API binary
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api cmd/api/main.go

docker-build: ## Build Docker images
	docker build -f Dockerfile.worker -t $(WORKER_IMAGE):$(VERSION) .
	docker build -f Dockerfile.api -t $(API_IMAGE):$(VERSION) .

docker-push: ## Push Docker images
	docker push $(WORKER_IMAGE):$(VERSION)
	docker push $(API_IMAGE):$(VERSION)

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	goimports -w .

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

dev: ## Run in development mode (requires Temporal server)
	@echo "Starting worker and API..."
	@trap 'kill 0' EXIT; \
	go run cmd/worker/main.go & \
	go run cmd/api/main.go & \
	wait
