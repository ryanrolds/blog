.PHONY: install build all push_prod push_test tag docker_build

TAG_NAME := $(shell git rev-parse --short HEAD)

install:
	go get -u gopkg.in/russross/blackfriday.v2
	go get -u github.com/gorilla/mux
	go get -u github.com/sirupsen/logrus
	go get -u github.com/antchfx/htmlquery
	go get -u github.com/Depado/bfchroma
	go get -u github.com/alecthomas/chroma
	go get -u github.com/alecthomas/chroma/formatters/html
	go get -u github.com/gorilla/handlers

build:
	go build

docker_build: build

push_docker_hub: docker_build
	docker build -t pedantic:test .
	docker tag pedantic:test ryanrolds/pedantic_orderliness:$(TAG_NAME)
	docker push ryanrolds/pedantic_orderliness:$(TAG_NAME)

push_aws: docker_build
	docker build -t pedantic:latest .
	docker tag pedantic:latest 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest
	docker push 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest

deploy_prod:
	aws ecs update-service --cluster pedantic --service pedantic-prod --force-new-deployment

deploy_k8s:
	TAG_NAME=$(TAG_NAME) ENV=production envsubst < k8s/deployment.manifest | kubectl replace -f -
	TAG_NAME=$(TAG_NAME) ENV=test envsubst < k8s/deployment.manifest | kubectl replace -f -

all: install build

