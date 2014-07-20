/*
https://godoc.org/code.google.com/p/go.net/websocket
http://blog.golang.org/spotlight-on-external-go-libraries
https://gist.github.com/jweir/4528042
https://github.com/golang-samples/websocket/blob/master/simple/main.go
http://blog.jupo.org/2013/02/23/a-tale-of-two-queues/
*/

package main

import (
	"code.google.com/p/go.net/websocket"
	"gopkg.in/redis.v1"
	"net/http"
	"fmt"
	_ "reflect"
)

var (
    redis_host = "127.0.0.1"
    redis_port = "6379"
    redis_channel = "pubsocketd"
    /* see below - I have no idea what I am doing */
    /* pubsub *redis.PubSub */
)

/*
func init() {

	addr := redis_host + ":" + redis_port

	client := redis.NewTCPClient(&redis.Options{
	    Addr: addr,
	})

	defer client.Close()

	pubsub := client.PubSub()
	defer pubsub.Close()

	err := pubsub.Subscribe(redis_channel)
}
*/

func pubSubHandler(ws *websocket.Conn) {

	fmt.Println("connecting!")
	// fmt.Println(ws.RemoteAddr())

	addr := redis_host + ":" + redis_port

	client := redis.NewTCPClient(&redis.Options{
	    Addr: addr,
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
		   // fmt.Println(msg.Payload)
		   websocket.JSON.Send(ws, msg.Payload)
		}	
	}
}

func main() {

	/* http://stackoverflow.com/questions/19708330/serving-a-websocket-in-go */

	http.HandleFunc("/", func (w http.ResponseWriter, req *http.Request){
        	s := websocket.Server{Handler: websocket.Handler(pubSubHandler)}
        	s.ServeHTTP(w, req)
    	});

	fmt.Println("Listening on " + redis_host + ":" + redis_port + " and relaying messages from '" + redis_channel + "'")

	http_err := http.ListenAndServe("127.0.0.1:8080", nil)

	if http_err != nil {
		panic("ListenAndServe: " + http_err.Error())
	}
}
