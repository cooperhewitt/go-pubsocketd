prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	rmdeps deps fmt bin

deps:   self
	@GOPATH=$(shell pwd) go get -u "gopkg.in/redis.v1"
	@GOPATH=$(shell pwd) go get -u "golang.org/x/net/websocket"

bin: 	self
	@GOPATH=$(shell pwd) go build -o bin/pubsocketd cmd/pubsocketd.go

fmt:
	go fmt cmd/*
