package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/redis.v1"
	"log"
	"net/http"
)

var (
	redisHost         string
	redisPort         int
	redisChannel      string
	redisEndpoint     string
	websocketHost     string
	websocketPort     int
	websocketEndpoint string
	websocketRoute    string
	redisClient       *redis.Client
)

func pubSubHandler(ws *websocket.Conn) {

	remoteAddr := ws.Request().RemoteAddr
	log.Printf("[%s][connect] hello world", remoteAddr)

	pubsubClient := redisClient.PubSub()
	defer pubsubClient.Close()

	if err := pubsubClient.Subscribe(redisChannel); err != nil {
		log.Printf("Failed to subscribe to pubsub channel %v, because %s", redisChannel, err)
		ws.Close()
		return
	}

	for ws != nil {

		i, _ := pubsubClient.Receive()

		if msg, _ := i.(*redis.Message); msg != nil {

			log.Printf("[%s][send] %s", remoteAddr, msg.Payload)

			var json_blob interface{}
			bytes_blob := []byte(msg.Payload)

			if err := json.Unmarshal(bytes_blob, &json_blob); err != nil {
				log.Printf("[%s][error] failed to parse JSON %s, because %v", msg.Payload, err)
				continue
			}

			if err := websocket.JSON.Send(ws, json_blob); err != nil {
				log.Printf("[%v][error] failed to send JSON, because %v", remoteAddr, err)
				ws.Close()
				break
			}
		}
	}
}

func websocketHandler(w http.ResponseWriter, req *http.Request) {
	s := websocket.Server{Handler: websocket.Handler(pubSubHandler)}
	s.ServeHTTP(w, req)
}

func main() {

	flag.StringVar(&websocketHost, "ws-host", "127.0.0.1", "Websocket host")
	flag.IntVar(&websocketPort, "ws-port", 8080, "Websocket port")
	flag.StringVar(&websocketRoute, "ws-route", "/", "Websocket route")

	flag.StringVar(&redisHost, "rs-host", "127.0.0.1", "Redis host")
	flag.IntVar(&redisPort, "rs-port", 6379, "Redis port")
	flag.StringVar(&redisChannel, "rs-channel", "pubsocketd", "Redis channel")

	flag.Parse()

	websocketEndpoint = fmt.Sprintf("%s:%d", websocketHost, websocketPort)
	redisEndpoint = fmt.Sprintf("%s:%d", redisHost, redisPort)

	redisClient = redis.NewTCPClient(&redis.Options{
		Addr: redisEndpoint,
	})

	defer redisClient.Close()

	//http.HandleFunc(websocketRoute, websocketHandler)

	http.Handle(websocketRoute, websocket.Handler(pubSubHandler))

	log.Printf("[init] listening for websocket requests on " + websocketEndpoint + websocketRoute)
	log.Printf("[init] listening for pubsub messages from " + redisEndpoint + " sent to the " + redisChannel + " channel")

	if err := http.ListenAndServe(websocketEndpoint, nil); err != nil {
		err, _ := fmt.Printf("Failed to start websocket server, because %v", err)
		panic(err)
	}
}
