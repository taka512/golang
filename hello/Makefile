ECR_DOMAIN         := 762539400516.dkr.ecr.ap-northeast-1.amazonaws.com
DOCKER_IMAGE       := $(ECR_DOMAIN)/develop

docker/login:
	aws ecr get-login-password | docker login --username AWS --password-stdin $(ECR_DOMAIN)


docker/push: docker/login
	docker tag hello $(ECR_DOMAIN)/$(NAME)
	docker push $(ECR_DOMAIN)/$(NAME)
