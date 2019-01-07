.PHONY: install build all push_prod push_test 

install:
	go get -u gopkg.in/russross/blackfriday.v2
	go get -u github.com/sirupsen/logrus
	go get -u github.com/antchfx/htmlquery
	go get -u github.com/Depado/bfchroma
	go get -u github.com/alecthomas/chroma
	go get -u github.com/alecthomas/chroma/formatters/html

build:
	go build

push_prod: 
	docker build -t pedantic .
	docker tag pedantic:latest 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest
	docker push 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:latest
	aws ecs update-service --cluster pedantic --service pedantic-prod --force-new-deployment

push_test:
	docker build -t pedantic:test .
	docker tag pedantic:test 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:test
	docker push 756280430156.dkr.ecr.us-west-2.amazonaws.com/pedantic:test
	aws ecs update-service --cluster pedantic --service pedantic-test --force-new-deployment

all: install build

