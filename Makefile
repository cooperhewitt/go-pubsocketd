CWD=$(shell pwd)
GOPATH := $(CWD)/vendor:$(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:   self
	@GOPATH=$(GOPATH) go get -u "gopkg.in/redis.v1"
	@GOPATH=$(GOPATH) go get -u "golang.org/x/net/websocket"

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/pubsocketd cmd/pubsocketd.go
	@GOPATH=$(GOPATH) go build -o bin/pubsocketd-publisher cmd/pubsocketd-publisher.go
	@GOPATH=$(GOPATH) go build -o bin/pubsocketd-receiver cmd/pubsocketd-receiver.go

vendor: rmdeps deps
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +

fmt:
	go fmt cmd/*
