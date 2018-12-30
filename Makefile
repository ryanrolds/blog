.PHONY: install build all

install:
	go get -u gopkg.in/russross/blackfriday.v2
	go get -u github.com/gorilla/mux
	go get -u github.com/sirupsen/logrus
	go get -u github.com/antchfx/htmlquery

build:
	go build

all: install build

