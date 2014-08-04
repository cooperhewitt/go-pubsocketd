# go-pubsocketd

Listen to a Redis PubSub chanhel and then rebroadcast over WebSockets.

The long version is available in this blog post titled [The Medium is the Message](http://labs.cooperhewitt.org/2014/the-medium-is-the-message-and-pubsocketd/).

## Building

	$> export GOPATH=/path/to/go-pubsocketd
	$> go get code.google.com/p/go.net/websocket
	$> go get gopkg.in/redis.v1
	$> go build pubsocketd.go

Or, if you are working on a system with the `make` command installed:

	$> make deps
	$> make build

## Example

If we assume the following:

* `server` means pubsocketd itself
* `client` means a WebSocket client, for example the same HTML/JS page in the `client` folder
* `pubsub` means something that publishes pubsub messages

### server

The first thing to do is start the `pubsocketd` server to accept WebSocket connections and relay pubsub messages.

	$> ./pubsocketd -ws-origin=http://example.com
	2014/07/20 13:43:50 [init] listening for websocket requests on 127.0.0.1:8080
	2014/07/20 13:43:50 [init] listening for pubsub messages from 127.0.0.1:6379#pubsocketd

See the `-ws-origin` flag? That's important and is discussed in detail below.

### client

Next start a WebSocket client and connect to the `pubsocketd` server.

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

Let's imagine that your client is set up to simply write WebSocket messages to the browser's console log.

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

## pubsocketd command-line options

In addition to the default command-line options exported by any Go application the following can be specified:

### -rs-channel="pubsocketd"

The Redis channel you want to listen (and relay messages for).

### -rs-host="127.0.0.1"

The Redis host you are connecting to.

### -rs-port=6379

The Redis port you are connecting to.

### -ws-host="127.0.0.1"

The host that this server will listen for connections on.

### -ws-port=8080

The port that this server will listen for connections on.

### -ws-route="/"

The path that this server will listen for connections on.

### -ws-origin= _required_

A list of valid hosts that may connect to this server.

Currently multiple hosts are not supported but will be in time.

### -ws-insecure=false

A boolean flag indicating that the WebSocket server should be run in "insecure" mode which will allow connections from any host.

This is available only for debugging and should **not** be enabled for production use.

### -ps-log-file="/var/log/pubsocketd.log"

Write all logging to this file, as well as STDOUT.

## See also

* http://redis.io/topics/pubsub
* https://godoc.org/code.google.com/p/go.net/websocket
* http://blog.golang.org/spotlight-on-external-go-libraries
* https://gist.github.com/jweir/4528042
* https://github.com/golang-samples/websocket/blob/master/simple/main.go
* https://github.com/golang-samples/websocket/blob/master/websocket-chat/src/chat/server.go
* http://blog.jupo.org/2013/02/23/a-tale-of-two-queues/
* http://stackoverflow.com/questions/19708330/serving-a-websocket-in-go
* http://www.christian-schneider.net/CrossSiteWebSocketHijacking.html
* http://tools.ietf.org/id/draft-abarth-origin-03.html
* https://github.com/mroth/sseserver
 
## Shout-outs

Props to [Richard Crowley](https://github.com/rcrowley) for patient comments and suggestions along the way.
