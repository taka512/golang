.PHONY: build run clean test docker-build docker-run deps help

# Binary name
BINARY_NAME=profit-trend-display
BINARY_PATH=./bin/$(BINARY_NAME)

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) .

# Run the application with default parameters (last 30 days)
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

# Run with custom days (usage: make run-days DAYS=7)
run-days: build
	@echo "Running $(BINARY_NAME) with $(DAYS) days..."
	@$(BINARY_PATH) -days $(DAYS)

# Run with summary only
run-summary: build
	@echo "Running $(BINARY_NAME) with summary only..."
	@$(BINARY_PATH) -summary

# Run with custom chart size (usage: make run-large WIDTH=80 HEIGHT=20)
run-large: build
	@echo "Running $(BINARY_NAME) with large chart..."
	@$(BINARY_PATH) -width $(WIDTH) -height $(HEIGHT)

# Run with custom database (usage: make run-db DSN="user:pass@tcp(host:port)/db")
run-db: build
	@echo "Running $(BINARY_NAME) with custom database..."
	@$(BINARY_PATH) -dsn $(DSN)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/

# Test the application
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@golangci-lint run

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME) .

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run --rm --network host $(BINARY_NAME)

# Install dependencies and build
install: deps build

# Quick examples
example-week: build
	@echo "Example: Weekly trend (7 days)"
	@$(BINARY_PATH) -days 7

example-summary: build
	@echo "Example: Summary only"
	@$(BINARY_PATH) -summary -days 30

example-large: build
	@echo "Example: Large chart"
	@$(BINARY_PATH) -width 100 -height 25 -days 14

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run with default settings (30 days)"
	@echo "  run-days      - Run with custom days (DAYS=N)"
	@echo "  run-summary   - Run with summary only"
	@echo "  run-large     - Run with large chart (WIDTH=N HEIGHT=N)"
	@echo "  run-db        - Run with custom database (DSN=connection)"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code (requires golangci-lint)"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run in Docker container"
	@echo "  install       - Install deps and build"
	@echo ""
	@echo "Examples:"
	@echo "  example-week  - Show 7-day trend"
	@echo "  example-summary - Show summary only"
	@echo "  example-large - Show large chart (14 days)"
	@echo ""
	@echo "Custom usage:"
	@echo "  make run-days DAYS=14"
	@echo "  make run-large WIDTH=80 HEIGHT=20"
	@echo "  make run-db DSN=\"user:pass@tcp(host:port)/database\""
	@echo ""
	@echo "  help          - Show this help"