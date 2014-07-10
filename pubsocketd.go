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
	"io"
	"net/http"
	"fmt"
)

func haltOnErr(err error){
	if err != nil { panic(err) }
}

func echoHandler(ws *websocket.Conn) {

     	fmt.Println("handler...")
	io.Copy(ws, ws)
}

func main() {

	fmt.Println("foo")

	client := redis.NewTCPClient(&redis.Options{
	    Addr: "localhost:6379",
	})
	defer client.Close()

	pubsub := client.PubSub()
	defer pubsub.Close()

	e := pubsub.Subscribe("mychannel")
	_ = e

	/* wtf... pass it a callback or something or what... */
	msg, er := pubsub.Receive()
	fmt.Println(msg, er)

	/* http://stackoverflow.com/questions/19708330/serving-a-websocket-in-go */

	http.HandleFunc("/", func (w http.ResponseWriter, req *http.Request){
        	s := websocket.Server{Handler: websocket.Handler(echoHandler)}
        	s.ServeHTTP(w, req)
    	});

	err := http.ListenAndServe("127.0.0.1:8080", nil)

	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
