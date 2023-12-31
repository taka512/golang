.PHONY: help docker/*
CONTAINER    :=go.local
ECR_DOMAIN   := 762539400516.dkr.ecr.ap-northeast-1.amazonaws.com

help:
	@grep -E '^[a-zA-Z\/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

docker/prune:
	docker system prune
docker/up: ## containers up
	docker compose up -d
docker/logs: ## display containers log
	docker compose logs -f
docker/stop: ## stop containers
	docker compose stop
docker/buid: ## build container
	docker build ./ -t $(NAME)
docker/ssh: ## connect in container
	docker exec -it -e COLUMNS=$(shell tput cols) -e LINES=$(shell tput lines) $(CONTAINER) ash
mysql/ssh:
	mysql -h 127.0.0.1 -P 3307 -u root -p sample_mysql
docker/login:
	aws ecr get-login-password | docker login --username AWS --password-stdin $(ECR_DOMAIN)
docker/push: docker/login
	docker tag $(NAME) $(ECR_DOMAIN)/$(NAME)
	docker push $(ECR_DOMAIN)/$(NAME)
