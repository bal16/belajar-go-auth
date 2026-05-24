# Simple Makefile for a Go project
-include .env
export

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Podman Compose"; \
		podman compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Podman Compose"; \
		podman compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
	
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
# 	@go test ./internal/database -v
# test-integration:
	RUN_INTEGRATION_TESTS=1 DB_HOST=localhost DB_PORT=${BLUEPRINT_DB_PORT} DB_USER=${BLUEPRINT_DB_USERNAME} DB_PASS=${BLUEPRINT_DB_PASSWORD} DB_NAME=${BLUEPRINT_DB_DATABASE} go test ./internal/connection -v -run TestGetDatabase_Integration

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch docker-run docker-down itest test-integration

dbmigrate:
	@echo "Running database migrations..."
	@go run ./cmd/database/migration.go
