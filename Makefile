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

docker_build: build
	docker build -t pedantic:test .

tag: docker_build
	docker tag pedantic:latest ryanrolds/pedantic_orderliness:$(TAG_NAME)
	docker push ryanrolds/pedantic_orderliness:$(TAG_NAME)

build:
	go build

push_prod:
	docker tag pedantic:latest 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest
	docker push 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest
	aws ecs update-service --cluster pedantic --service pedantic-prod --force-new-deployment

push_test:
	docker tag pedantic:test 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:test
	docker push 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:test
	aws ecs update-service --cluster pedantic --service pedantic-test --force-new-deployment

push_k8s:
	TAG_NAME=$(TAG_NAME) ENV=production envsubst < k8s/deployment.manifest | kubectl replace -f -
	TAG_NAME=$(TAG_NAME) ENV=test envsubst < k8s/deployment.manifest | kubectl replace -f -

all: install build

