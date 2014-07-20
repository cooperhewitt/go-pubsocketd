/*
https://godoc.org/code.google.com/p/go.net/websocket
http://blog.golang.org/spotlight-on-external-go-libraries
https://gist.github.com/jweir/4528042
https://github.com/golang-samples/websocket/blob/master/simple/main.go
http://blog.jupo.org/2013/02/23/a-tale-of-two-queues/
http://stackoverflow.com/questions/19708330/serving-a-websocket-in-go
*/

package main

import (
	"code.google.com/p/go.net/websocket"
	"gopkg.in/redis.v1"
	"net/http"
	"fmt"
	"log"
	"flag"
	_ "os"
	_ "reflect"
)

var (
    redis_host string
    redis_port int
    redis_channel string
    redis_endpoint string
    websocket_host string
    websocket_port int
    websocket_endpoint string
    pubsub *redis.PubSub
)

func init() {

     flag.StringVar(&websocket_host, "ws-host", "127.0.0.1", "Websocket host")
     flag.IntVar(&websocket_port, "ws-port", 8080, "Websocket port")
     flag.StringVar(&redis_host, "rs-host", "127.0.0.1", "Redis host")
     flag.IntVar(&redis_port, "rs-port", 6379, "Redis port")
     flag.StringVar(&redis_channel, "rs-channel", "pubsocketd", "Redis channel")

     /*

     client := redis.NewTCPClient(&redis.Options{
	    Addr: redis_endpoint,
	})

	defer client.Close()

	pubsub := client.PubSub()
	defer pubsub.Close()

	err := pubsub.Subscribe(redis_channel)
	*/
}

func pubSubHandler(ws *websocket.Conn) {

	log.Printf("connecting!")

	client := redis.NewTCPClient(&redis.Options{
	    Addr: redis_endpoint,
	})

	defer client.Close()

	pubsub := client.PubSub()
	defer pubsub.Close()

	err := pubsub.Subscribe(redis_channel)

	if err != nil {
	   websocket.JSON.Send(ws, "FUBAR")
	   ws.Close()
	}

	for {
		i, _ := pubsub.Receive()
		msg, _ := i.(*redis.Message)

		if msg != nil {
		   websocket.JSON.Send(ws, msg.Payload)
		}	
	}
}

func main() {

     // not that init() is invoked before we get here

     flag.Parse()

     websocket_endpoint = fmt.Sprintf("%s:%d", websocket_host, websocket_port)
     redis_endpoint = fmt.Sprintf("%s:%d", redis_host, redis_port)

	http.HandleFunc("/", func (w http.ResponseWriter, req *http.Request){
        	s := websocket.Server{Handler: websocket.Handler(pubSubHandler)}
        	s.ServeHTTP(w, req)
    	});

	log.Printf("Listening for websocket requests on " + websocket_endpoint)
	log.Printf("Listening for pubsub messages from " + redis_endpoint + "#" + redis_channel)

	http_err := http.ListenAndServe(websocket_endpoint, nil)

	if http_err != nil {
		panic("ListenAndServe: " + http_err.Error())
	}
}
