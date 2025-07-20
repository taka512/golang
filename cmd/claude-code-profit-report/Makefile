.PHONY: build run clean help test deps

# Variables
BINARY_NAME=claude-code-profit-report
GO=go
GOFLAGS=
LDFLAGS=

# Default target
all: build

## help: Show this help message
help:
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## deps: Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

## build: Build the binary
build: deps
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

## run: Run the application with sample parameters
run: build
	./$(BINARY_NAME) -s 2025-07-14 -e 2025-07-20

## run-company: Run for specific company (all warehouses)
run-company: build
	./$(BINARY_NAME) -c 1 -s 2024-01-01 -e 2024-01-31

## run-warehouse: Run for specific warehouse (all companies)
run-warehouse: build
	./$(BINARY_NAME) -w 1 -s 2024-01-01 -e 2024-01-31

## run-specific: Run for specific company and warehouse
run-specific: build
	./$(BINARY_NAME) -c 1 -w 1 -s 2024-01-01 -e 2024-01-31

## run-slack: Run with Slack output (requires SLACK_HOOK env var)
run-slack: build
	./$(BINARY_NAME) -c 1 -w 1 -s 2024-01-01 -e 2024-01-31 --slack

## test: Run tests
test:
	$(GO) test -v ./...

## clean: Clean build artifacts
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

## install: Install the binary to $GOPATH/bin
install: build
	$(GO) install

# Development targets
## dev: Run with live reload (requires air)
dev:
	@which air > /dev/null || (echo "Please install air: go install github.com/cosmtrek/air@latest" && exit 1)
	air

## fmt: Format code
fmt:
	$(GO) fmt ./...

## vet: Run go vet
vet:
	$(GO) vet ./...

## lint: Run golangci-lint (requires golangci-lint)
lint:
	@which golangci-lint > /dev/null || (echo "Please install golangci-lint" && exit 1)
	golangci-lint run
