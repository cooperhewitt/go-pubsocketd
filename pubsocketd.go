/*
https://godoc.org/code.google.com/p/go.net/websocket
http://blog.golang.org/spotlight-on-external-go-libraries
https://gist.github.com/jweir/4528042
https://github.com/golang-samples/websocket/blob/master/simple/main.go
https://github.com/golang-samples/websocket/blob/master/websocket-chat/src/chat/server.go
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
	"encoding/json"
)

var (
	redis_host string
	redis_port int
	redis_channel string
	redis_endpoint string
	websocket_host string
	websocket_port int
	websocket_endpoint string
	redis_client *redis.Client
)

func pubSubHandler(ws *websocket.Conn) {

     	remote_addr := ws.Request().RemoteAddr
	log.Printf("[%s][connect] hello world", remote_addr)

	pubsub_client := redis_client.PubSub()
	defer pubsub_client.Close()

	sub_err := pubsub_client.Subscribe(redis_channel)

	if sub_err != nil {
		log.Printf("Failed to subscribe to pubsub channel %s, because %s", redis_channel, sub_err.Error())
		ws.Close()
		return
	}

	for ws != nil {

		i, _ := pubsub_client.Receive()
		msg, _ := i.(*redis.Message)

		if msg != nil {

			log.Printf("[%s][send] %s", remote_addr, msg.Payload)

			var json_blob interface{}
			bytes_blob := []byte(msg.Payload)

			json_err := json.Unmarshal(bytes_blob, &json_blob)

			if json_err != nil{
				log.Printf("[%s][error] failed to parse JSON %s, because %s", msg.Payload, json_err.Error())
				continue
			}

		   	send_err := websocket.JSON.Send(ws, json_blob)

			if send_err != nil{
				log.Printf("[%s][error] failed to send JSON, because %s", remote_addr, send_err.Error())
				ws.Close()
				break
			}
		}
	}
}

func main() {

	flag.StringVar(&websocket_host, "ws-host", "127.0.0.1", "Websocket host")
	flag.IntVar(&websocket_port, "ws-port", 8080, "Websocket port")
	flag.StringVar(&redis_host, "rs-host", "127.0.0.1", "Redis host")
	flag.IntVar(&redis_port, "rs-port", 6379, "Redis port")
	flag.StringVar(&redis_channel, "rs-channel", "pubsocketd", "Redis channel")

	flag.Parse()

	websocket_endpoint = fmt.Sprintf("%s:%d", websocket_host, websocket_port)
	redis_endpoint = fmt.Sprintf("%s:%d", redis_host, redis_port)

	redis_client = redis.NewTCPClient(&redis.Options{
		Addr: redis_endpoint,
	})

	defer redis_client.Close()

	http.HandleFunc("/", func (w http.ResponseWriter, req *http.Request){
        	s := websocket.Server{Handler: websocket.Handler(pubSubHandler)}
        	s.ServeHTTP(w, req)
    	});

	log.Printf("[init] listening for websocket requests on " + websocket_endpoint)
	log.Printf("[init] listening for pubsub messages from " + redis_endpoint + " sent to the " + redis_channel + " channel")

	http_err := http.ListenAndServe(websocket_endpoint, nil)

	if http_err != nil {
		err, _ := fmt.Printf("Failed to start websocket server, because %s", http_err.Error())
		panic(err)
	}
}
