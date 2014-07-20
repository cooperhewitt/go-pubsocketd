# go-pubsocketd

Listen to a Redis PubSub chanhel and then rebroadcast over WebSockets.

## Building

	$> setenv GOPATH /path/to/go-pubsocketd
	$> go get code.google.com/p/go.net/websocket
	$> go get gopkg.in/redis.v1
	$> go build pubsocketd.go

## Example

If we assume the following:

* `server` means pubsocketd itself
* `client` means a websocket client, for example the same HTML/JS page in the `client` folder
* `pubsub` means something that publishes pubsub messages

### server

The first thing to do is start the `pubsocketd` server to accept websocket connections and relay pubsub messages.

	$> ./pubsocketd 
	2014/07/20 13:43:50 [init] listening for websocket requests on 127.0.0.1:8080
	2014/07/20 13:43:50 [init] listening for pubsub messages from 127.0.0.1:6379#pubsocketd

### client

Next start a websocket client and connect to the `pubsocketd` server.

	var socket = new WebSocket('ws://127.0.0.1:8080');

	socket.onopen = function(e){
		console.log(e);
	};

Something like this should be printed to your browser's console log.

	open { target: WebSocket, eventPhase: 0, bubbles: false, cancelable: false, defaultPrevented: false, timeStamp: 1405878232662574, originalTarget: WebSocket, explicitOriginalTarget: WebSocket, isTrusted: true, NONE: 0, CAPTURING_PHASE: 1 } index.html:11

### server

Something like this should be printed to your `pubsocketd` server logs.

	2014/07/20 13:43:52 [127.0.0.1:50005][connect] hello world

### pubsub

Now publish a pubsub message to the `pubsocketd` channel. This example does so using Python but there are many other language implementations.

	$> python
	import redis
	r = redis.Redis()
	r.publish('pubsocketd', {'foo': 1, 'bar': 2})

### server

Something like this should be printed to your `pubsocketd` server logs.

	2014/07/20 13:43:57 [127.0.0.1:50005][send] {'foo': 1, 'bar': 2}

### client

Let's imagine that your client is set up to simply write websocket messages to the browser's console log.

	socket.onmessage = function(rsp){

		var data = rsp['data'];
		data = JSON.parse(data);

		console.log(rsp);
		console.log(data);
	};

The code above would yield something like this:

	message { target: WebSocket, data: ""{'foo': 1, 'bar': 2}"", origin: "ws://127.0.0.1:8080", lastEventId: "", isTrusted: true, eventPhase: 0, bubbles: false, cancelable: false, defaultPrevented: false, timeStamp: 1405878237791142, originalTarget: WebSocket } index.html:16
	Object { foo: 1, bar: 2}

The rest is up to you!
