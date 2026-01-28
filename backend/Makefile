include .env
export
CONN_STRING := "postgres://$(APP_DATABASE_USER):$(APP_DATABASE_PASSWORD)@$(APP_DATABASE_HOST):$(APP_DATABASE_PORT)/$(APP_DATABASE_DBNAME)?sslmode=$(APP_DATABASE_SSLMODE)"
.PHONY: help run build test clean migrate docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application
	go run cmd/api/main.go

build: ## Build the application
	go build -o bin/api cmd/api/main.go

test: ## Run tests
	go test -v -cover ./...

test-coverage: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	go mod download
	go mod tidy

lint: ## Run linter
	golangci-lint run

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

migrate-create: ## Create a new migration
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	migrate -path migrations -database "$(CONN_STRING)" up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	migrate -path migrations -database "$(CONN_STRING)" down

migrate-status: ## Check migration status
	@echo "Checking database tables..."
	migrate -path migrations -database "$(CONN_STRING)" status

dev: ## Run in development mode with hot reload (requires air)
	air

install-tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v2@latest
