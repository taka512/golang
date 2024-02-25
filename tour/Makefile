FILE :=

help:
    @grep -E '^[a-zA-Z\/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build:
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/$(FILE) $(FILE).go
run: build
	./bin/$(FILE)
