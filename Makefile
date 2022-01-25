.PHONY: install build all push_prod push_test tag docker_build

TAG_NAME := $(shell git rev-parse --short HEAD)

install: build

build:
	go build

docker_build: build

push_k8s: docker_build
	docker build -t $(TAG_NAME) .
	docker tag $(TAG_NAME) docker.pedanticorderliness.com/pedantic:$(TAG_NAME)
	docker push docker.pedanticorderliness.com/pedantic:$(TAG_NAME)

login_aws:
	aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 756280430156.dkr.ecr.us-west-2.amazonaws.com

push_aws: docker_build
	docker build -t pedantic:latest .
	docker tag pedantic:latest 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest
	docker push 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest

deploy_prod:
	aws ecs update-service --cluster pedantic --service pedantic-prod --force-new-deployment

deploy_k8s:
	TAG_NAME=$(TAG_NAME) ENV=test envsubst < k8s/deployment.yaml | kubectl replace -f -

all: install build

