.PHONY: help docker/*
CONTAINER    :=go.local
ECR_DOMAIN   := 762539400516.dkr.ecr.ap-northeast-1.amazonaws.com

help:
	@grep -E '^[a-zA-Z\/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

docker/prune: ## 色々掃除
	docker system prune
docker/image/rm: ## イメージを全削除
	docker rmi -f $$(docker images -q)
docker/up: ## コンテナ起動
	docker compose up -d
docker/logs: ## コンテナログの表示
	docker compose logs -f
docker/stop: ## コンテナ停止
	docker compose stop
docker/ssh: ## コンテナにsshで接続
	docker exec -it -e COLUMNS=$(shell tput cols) -e LINES=$(shell tput lines) $(CONTAINER) bash
mysql/ssh:
	mysql -h 127.0.0.1 -P 3307 -u root -p sample_mysql
docker/login:
	aws ecr get-login-password | docker login --username AWS --password-stdin $(ECR_DOMAIN)
docker/push: docker/login
	docker tag $(NAME) $(ECR_DOMAIN)/golang-hello
	docker push $(ECR_DOMAIN)/golang-hello
#	docker tag $(NAME) $(ECR_DOMAIN)/$(NAME)
#	docker push $(ECR_DOMAIN)/$(NAME)
