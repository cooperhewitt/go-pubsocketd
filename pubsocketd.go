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
	"reflect"
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
	    Addr: "127.0.0.1:6379",
	})
	defer client.Close()

	pubsub := client.PubSub()
	defer pubsub.Close()

	err := pubsub.Subscribe("mychannel")
	haltOnErr(err)

	/* http://golangtutorials.blogspot.com/2011/06/interfaces-in-go.html */

	for{
		fmt.Println("for")
		msg, er := pubsub.Receive()
		fmt.Println(msg, er)

		fmt.Println(reflect.TypeOf(msg))
	}

	fmt.Println("WHAT")

	/* http://stackoverflow.com/questions/19708330/serving-a-websocket-in-go */

	http.HandleFunc("/", func (w http.ResponseWriter, req *http.Request){
        	s := websocket.Server{Handler: websocket.Handler(echoHandler)}
        	s.ServeHTTP(w, req)
    	});

	http_err := http.ListenAndServe("127.0.0.1:8080", nil)

	if http_err != nil {
		panic("ListenAndServe: " + http_err.Error())
	}
}
