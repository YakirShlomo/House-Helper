.PHONY: help dev-up dev-down dev-logs seed test test-mobile test-api test-integration test-load build build-mobile build-api docker-build deploy-dev deploy-prod infra-plan infra-apply demo-timer

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Local Development
dev-up: ## Start all local services
	@echo "Starting local development environment..."
	docker-compose -f deploy/docker-compose.yml up -d
	@echo "Services started. Wait for initialization..."
	@sleep 10
	@echo "✅ Development environment ready!"

dev-down: ## Stop all local services
	@echo "Stopping local development environment..."
	docker-compose -f deploy/docker-compose.yml down

dev-logs: ## View service logs
	docker-compose -f deploy/docker-compose.yml logs -f

seed: ## Load demo data
	@echo "Loading demo data..."
	docker-compose -f deploy/docker-compose.yml exec api go run cmd/seed/main.go
	@echo "✅ Demo data loaded!"

# Testing
test: test-mobile test-api ## Run all tests
	@echo "✅ All tests completed!"

test-mobile: ## Run Flutter tests
	@echo "Running Flutter tests..."
	cd apps/mobile_flutter && flutter test

test-api: ## Run Go tests
	@echo "Running Go API tests..."
	cd services/api && go test ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	cd services/api && go test ./tests/integration/...

test-load: ## Run k6 load tests
	@echo "Running load tests..."
	k6 run scripts/load-test.js

# Building
build: build-mobile build-api ## Build all components
	@echo "✅ All components built!"

build-mobile: ## Build Flutter apps
	@echo "Building Flutter app..."
	cd apps/mobile_flutter && flutter build apk --release
	cd apps/mobile_flutter && flutter build ios --release --no-codesign

build-api: ## Build Go API
	@echo "Building Go API..."
	cd services/api && make build

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker build -t house-helper/api:latest services/api/
	docker build -t house-helper/notifier:latest services/notifier/

# Deployment
deploy-dev: ## Deploy to development
	@echo "Deploying to development..."
	cd infra/helm && helm upgrade --install house-helper-api ./api -f values-dev.yaml

deploy-prod: ## Deploy to production
	@echo "Deploying to production..."
	cd infra/helm && helm upgrade --install house-helper-api ./api -f values-prod.yaml

infra-plan: ## Terraform plan
	@echo "Planning infrastructure changes..."
	cd infra/terraform/envs/dev && terraform plan

infra-apply: ## Terraform apply
	@echo "Applying infrastructure changes..."
	cd infra/terraform/envs/dev && terraform apply

# Demo
demo-timer: ## Demonstrate timer workflow
	@echo "🧺 Starting laundry timer demo..."
	curl -X POST http://localhost:8080/v1/timers/start \
		-H "Content-Type: application/json" \
		-d '{"type":"laundry","duration":"45m","task_id":"demo-123"}'
	@echo "\n✅ Timer started! Check notifications in 45 minutes."