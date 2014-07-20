# go-pubsocketd

Listen to a Redis PubSub chanhel and then rebroadcast over WebSockets.

## Building

	setenv GOPATH /path/to/go-pubsocketd
	go get code.google.com/p/go.net/websocket
	go get gopkg.in/redis.v1
	go build pubsocketd.go
