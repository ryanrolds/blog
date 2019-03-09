.PHONY: install build all push_prod push_test tag docker_build

install:
	go get -u gopkg.in/russross/blackfriday.v2
	go get -u github.com/gorilla/mux
	go get -u github.com/sirupsen/logrus
	go get -u github.com/antchfx/htmlquery
	go get -u github.com/Depado/bfchroma
	go get -u github.com/alecthomas/chroma
	go get -u github.com/alecthomas/chroma/formatters/html
	go get -u github.com/gorilla/handlers

tag:
	export TAG_NAME=$(git rev-parse --short HEAD)
	docker tag pedantic:latest ryanrolds/pedantic_orderliness:$(TAG_NAME)
	docker push ryanrolds/pedantic_orderliness:$(TAG_NAME)

build:
	go build

push_prod: build docker_build
	docker tag pedantic:latest 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest
	docker push 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest
	aws ecs update-service --cluster pedantic --service pedantic-prod --force-new-deployment

docker_build: build
	docker build -t pedantic:test .

push_test: docker_build
	docker tag pedantic:test 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:test
	docker push 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:test
	aws ecs update-service --cluster pedantic --service pedantic-test --force-new-deployment

push_k8s: docker_build tag
	envsubst < k8s/deployment.manifest | kubectl create -f -

all: install build

