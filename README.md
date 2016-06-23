# go-pubsocketd

Listen to a Redis PubSub channel and then rebroadcast over WebSockets.

The long version is available in this blog post titled [The Medium is the Message](http://labs.cooperhewitt.org/2014/the-medium-is-the-message-and-pubsocketd/).

## Building

The easiest way to get started is to use the handy `build` target in local Makefile

```
make build
```

All of the package's dependencies are included with this repository (in the `vendor` directory) but you will still need to have a copy of [Go](http://www.golang.org) installed on your computer.

## Example

If we assume the following:

* `server` means pubsocketd itself
* `client` means a WebSocket client, for example the same HTML/JS page in the `client` folder
* `pubsub` means something that publishes pubsub messages

### server

The first thing to do is start the `pubsocketd` server to accept WebSocket connections and relay pubsub messages.

```
$> ./pubsocketd -ws-origin=http://example.com
2014/07/20 13:43:50 [init] listening for websocket requests on 127.0.0.1:8080
2014/07/20 13:43:50 [init] listening for pubsub messages from 127.0.0.1:6379#pubsocketd
```

See the `-ws-origin` flag? That's important and is discussed in detail below. If you include `-tls-cert` and `-tls-key` flags `pubsocketd` will speak TLS and you should use `wss:` instead of `ws:` on the client.

### client

Next start a WebSocket client and connect to the `pubsocketd` server.

```
var socket = new WebSocket('ws://127.0.0.1:8080');

socket.onopen = function(e){
	console.log(e);
};
```

Something like this should be printed to your browser's console log.

	open { target: WebSocket, eventPhase: 0, bubbles: false, cancelable: false, defaultPrevented: false, timeStamp: 1405878232662574, originalTarget: WebSocket, explicitOriginalTarget: WebSocket, isTrusted: true, NONE: 0, CAPTURING_PHASE: 1 } index.html:11

### server

Something like this should be printed to your `pubsocketd` server logs.

```
	2014/07/20 13:43:52 [127.0.0.1:50005][connect] hello world
```

### pubsub

Now publish a pubsub message to the `pubsocketd` channel. This example does so using Python but there are many other language implementations.

```
$> python
import redis
r = redis.Redis()
r.publish('pubsocketd', {'foo': 1, 'bar': 2})
```

### server

Something like this should be printed to your `pubsocketd` server logs.

```
2014/07/20 13:43:57 [127.0.0.1:50005][send] {'foo': 1, 'bar': 2}
```

### client

Let's imagine that your client is set up to simply write WebSocket messages to the browser's console log.

```
socket.onmessage = function(rsp){

	var data = rsp['data'];
	data = JSON.parse(data);

	console.log(rsp);
	console.log(data);
};
```

The code above would yield something like this:

```
	message { target: WebSocket, data: ""{'foo': 1, 'bar': 2}"", origin: "ws://127.0.0.1:8080", lastEventId: "", isTrusted: true, eventPhase: 0, bubbles: false, cancelable: false, defaultPrevented: false, timeStamp: 1405878237791142, originalTarget: WebSocket } index.html:16
	Object { foo: 1, bar: 2}
```

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

### -ws-origin=http://example.com

A list of valid hosts that may connect to this server. This flag is required unless you are running `pubsocketd` in "insecure mode" (details below) which you should not do for production use.

Currently multiple hosts are not supported but will be in time.

### -ws-insecure=false

A boolean flag indicating that the WebSocket server should be run in "insecure" mode which will allow connections from any host.

This is available only for debugging and should **not** be enabled for production use.

### -ws-heartbeat=false

A boolean flag that activates a keep-alive heartbeat message, sent once every 30 seconds.

### -ps-log-file="/var/log/pubsocketd.log"

Write all logging to this file, as well as STDOUT.

## Utilities

There are command-line tools that are included to help test and debug your instance of `pubsocketd`. First of all let's assume that you've started a copy of `pubsocketd` with the following arguments:

```
./bin/pubsocketd -ws-origin http://localhost
[pubsocketd] 2016/06/08 04:31:21 pubsocketd.go:250: [init] listening for websocket requests on 127.0.0.1:8080/, from http://localhost
[pubsocketd] 2016/06/08 04:31:21 pubsocketd.go:253: [init] listening for pubsub messages from 127.0.0.1:6379 sent to the pubsocketd channel
```

### pubsocketd-publisher

`pubsocketd-publisher` reads input from STDIN, encodes it as a JSON string and broadcasts the message over a Redis PubSub channel.

```
./bin/pubsocketd-publisher -h
Usage of ./bin/pubsocketd-publisher:
  -channel string
    	   Redis channel (default "pubsocketd")
  -host string
    	Redis host (default "127.0.0.1")
  -port int
    	Redis port (default 6379)
```

For example:

```
./bin/pubsocketd-publisher
2016/06/08 04:35:48 connecting to 127.0.0.1:6379...
2016/06/08 04:35:48 connected to 127.0.0.1:6379 and ready to send new messages
PEW PEW PEW
WOO WOO WOO
```

### pubsocketd-receiver

`pubsocketd-receiver` listens for messages from a WebSocket connections and prints each one to STDOUT.

```
./bin/pubsocketd-receiver -h
Usage of ./bin/pubsocketd-receiver:
  -origin string
    	  The origin header to send
  -url string
       The websocket URL to connect to (default "ws://127.0.0.1:8080")
```

For example:

```
./bin/pubsocketd-receiver -origin http://localhost
2016/06/08 04:35:59 dialing ws://127.0.0.1:8080...
2016/06/08 04:35:59 connected to ws://127.0.0.1:8080 and ready to receive new messages
2016/06/08 04:36:09 {"Data":"PEW PEW PEW"}
2016/06/08 04:36:25 {"Data":"WOO WOO WOO"}
```

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

## Gotchas

If you're running the server in a VM, you may need to explicitly run on host `0.0.0.0` instead of the default `127.0.0.1`:

```
-ws-host=0.0.0.0
```

Also, if you're using a self-signed SSL certificate, you'll probably need to visit `https://localhost:8080/` or whatever the address is, and create an exception for the SSL validation error.

There's some discussion of the port forwarding issue on [Stack Overflow](http://stackoverflow.com/questions/23840098/empty-reply-from-server-cant-connect-to-vagrant-vm-w-port-forwarding).

---

There's another gotcha that has to do with maintaining an open connection when the site is running through an AWS Elastic Load Balancer. It seems that after 60 seconds of inactivity the connection consistently stops working (even with some reconnection logic on the client side). There might be an ELB configuration to turn this timeout off, but I do not presently know what it is, and we probably want to keep a timeout in place to keep our servers happy. (We did try enabling `proxy_prococol` with nginx, which _seemed_ like the right solution. It wasn't.)

A workaround that does work is to send out a ping every 30 seconds (really any span of time less than 60 seconds). That keeps the connection alive. Use the option `-ws-heartbeat` to send out a `{"heartbeat": 1}` message every 30 seconds.

## Shout-outs

Props to [Richard Crowley](https://github.com/rcrowley) for patient comments and suggestions along the way.
