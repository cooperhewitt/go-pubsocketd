# https://godoc.org/code.google.com/p/go.net/websocket
# https://gist.github.com/jweir/4528042

package main

import (
	"log"       
)

func createSubscription(sh * subscriptionHandler, pubChan * string){
	log.Printf("creating channel %s",*pubChan)
 
	pubsub, err := redis.NewTCPClient(":6379","",-1).PubSubClient()
	haltOnErr(err)
 
	ch, err := pubsub.Subscribe(*pubChan)
	haltOnErr(err)
 
}

func main(){

}
