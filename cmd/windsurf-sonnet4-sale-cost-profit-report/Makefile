.PHONY: build run clean test docker-build docker-run

# Binary name
BINARY_NAME=sale-cost-profit-report
BINARY_PATH=./bin/$(BINARY_NAME)

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) .

# Run the application with default parameters (current month)
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

# Run with custom date range (usage: make run-range START=2024-01-01 END=2024-01-31)
run-range: build
	@echo "Running $(BINARY_NAME) with date range $(START) to $(END)..."
	@$(BINARY_PATH) $(START) $(END)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f *.csv

# Test the application
test:
	@echo "Running tests..."
	@go test -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME) .

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run --rm $(BINARY_NAME)

# Install dependencies and build
install: deps build

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Run with default date range (current month)"
	@echo "  run-range   - Run with custom date range (START=date END=date)"
	@echo "  clean       - Clean build artifacts and CSV files"
	@echo "  test        - Run tests"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  docker-build- Build Docker image"
	@echo "  docker-run  - Run in Docker container"
	@echo "  install     - Install deps and build"
	@echo "  help        - Show this help"
