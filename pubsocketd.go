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
    foo = "bar"
)

func haltOnErr(err error){
	if err != nil { panic(err) }
}

func echoHandler(ws *websocket.Conn) {

	fmt.Println("connecting!")

	client := redis.NewTCPClient(&redis.Options{
	    Addr: "127.0.0.1:6379",
	})

	defer client.Close()

	pubsub := client.PubSub()
	defer pubsub.Close()

	err := pubsub.Subscribe("mychannel")
	haltOnErr(err)

	if err != nil{
		return
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
        	s := websocket.Server{Handler: websocket.Handler(echoHandler)}
        	s.ServeHTTP(w, req)
    	});

	http_err := http.ListenAndServe("127.0.0.1:8080", nil)

	if http_err != nil {
		panic("ListenAndServe: " + http_err.Error())
	}

	fmt.Println("Listening on 127.0.0.1:8080")
}
