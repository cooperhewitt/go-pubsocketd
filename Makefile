foo: fmt build

build:
	export GOPATH=$(shell pwd)
	go build ./pubsocketd.go

fmt:
	gofmt ./pubsocketd.go > ./pubsocketd.go.fmt
	mv ./pubsocketd.go.fmt ./pubsocketd.go
