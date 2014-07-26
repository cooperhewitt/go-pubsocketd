fmt:
	gofmt ./pubsocketd.go > ./pubsocketd.go.fmt
	mv ./pubsocketd.go.fmt ./pubsocketd.go
