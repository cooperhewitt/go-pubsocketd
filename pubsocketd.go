# https://godoc.org/code.google.com/p/go.net/websocket
# https://gist.github.com/jweir/4528042
# https://github.com/golang-samples/websocket/blob/master/simple/main.go

package main

import (
	"log"
	"code.google.com/p/go.net/websocket"
	"io"
	"net/http"
)

func createSubscription(sh * subscriptionHandler, pubChan * string){
	log.Printf("creating channel %s",*pubChan)
 
	pubsub, err := redis.NewTCPClient(":6379","",-1).PubSubClient()
	haltOnErr(err)
 
	ch, err := pubsub.Subscribe(*pubChan)
	haltOnErr(err)
 
}

func echoHandler(ws *websocket.Conn) {

	# can I just plug createSubscription here or... ?
	# io.Copy(ws, ws)
}

func main() {

	http.Handle("/", websocket.Handler(echoHandler))
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
