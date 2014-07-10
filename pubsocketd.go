/*
https://godoc.org/code.google.com/p/go.net/websocket
http://blog.golang.org/spotlight-on-external-go-libraries
https://gist.github.com/jweir/4528042
https://github.com/golang-samples/websocket/blob/master/simple/main.go
*/

package main

import (
	/* "log" */
	websocket "code.google.com/p/go.net/websocket"
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

	client := redis.NewTCPClient(&redis.Options{
	    Addr: "localhost:6379",
	})
	defer client.Close()

	pubsub := client.PubSub()
	defer pubsub.Close()

	err := pubsub.Subscribe("mychannel")
	_ = err

	msg, err := pubsub.Receive()
	fmt.Println(msg, err)

	io.Copy(ws, ws)
}

func main() {

	fmt.Println("foo")

	/* http://stackoverflow.com/questions/19708330/serving-a-websocket-in-go */

	http.HandleFunc("/",
    func (w http.ResponseWriter, req *http.Request) {
        s := websocket.Server{Handler: websocket.Handler(echoHandler)}
        s.ServeHTTP(w, req)
    });

	err := http.ListenAndServe("127.0.0.1:8080", nil)

	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
