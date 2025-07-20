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


help:
    @grep -E '^[a-zA-Z\/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

docker/build: ## コンテナビルド
	docker build ./ -t $(BINARY_NAME)
docker/run: docker/build ## コンテナ実行
	docker run --network=golang_mynetwork -t $(BINARY_NAME)

all: deps build-with-dir
