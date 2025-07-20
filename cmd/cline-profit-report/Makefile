.PHONY: build run clean

BINARY_NAME=profit-report

build:
	go build -o bin/$(BINARY_NAME) .

run: build
	./bin/$(BINARY_NAME)

clean:
	rm -rf bin/

deps:
	go mod tidy
	go mod download

# Create bin directory if it doesn't exist
bin:
	mkdir -p bin

# Build with bin directory creation
build-with-dir: bin
	go build -o bin/$(BINARY_NAME) .

all: deps build-with-dir
