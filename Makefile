SHELL:=/bin/bash
deps:
	export GOPATH=$(shell pwd)
	go get code.google.com/p/go.net/websocket
	go get gopkg.in/redis.v1

build:
	export GOPATH=$(shell pwd)
	go build ./pubsocketd.go

fmt:
	gofmt ./pubsocketd.go > ./pubsocketd.go.fmt
	mv ./pubsocketd.go.fmt ./pubsocketd.go
