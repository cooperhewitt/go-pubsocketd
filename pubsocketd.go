package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/redis.v1"
	"log"
	"net/http"
	"net/url"
	_ "strings"
)

var (
	redisHost                 string
	redisPort                 int
	redisChannel              string
	redisEndpoint             string
	websocketHost             string
	websocketPort             int
	websocketEndpoint         string
	websocketRoute            string
	websocketAllowableOrigins string
	// see below (20140727/straup)
	// websocketAllowableURLs   []url.URL
	redisClient *redis.Client
)

func pubsocketdHandler(w http.ResponseWriter, req *http.Request) {

	// Say what? See comments below where we assign http.HandleFunc
	// (20140727/straup)

	// This is meant to be a list of URLs but since I don't really
	// grok arrays in Go yet... it's not (20140727/straup)

	origin := websocketAllowableOrigins
	url, err := url.Parse(origin)

	if err != nil {
		err, _ := fmt.Printf("Failed to start websocket server, because Origin URL '%v' won't parse, %v", origin, err)
		panic(err)
	}

	config := websocket.Config{Origin: url}

	s := websocket.Server{
		Config:    config,
		Handler:   websocket.Handler(pubSubHandler),
		Handshake: pubsocketdHandshake,
	}

	s.ServeHTTP(w, req)
}

func pubsocketdHandshake(config *websocket.Config, req *http.Request) (err error) {

	remoteAddr := req.RemoteAddr
	headers := req.Header

	origin := headers.Get("Origin")
	realIP := headers.Get("X-Real-IP")

	if origin == "" {
		log.Printf("[%s][%s][handshake] missing origin", realIP, remoteAddr)
		return fmt.Errorf("missing origin")
	}

	parsed, err := url.Parse(origin)

	if err != nil {
		log.Printf("[%s][%s][handshake] failed to parse origin, %v", realIP, remoteAddr, origin)
		return fmt.Errorf("invalid origin")
	}

	// See above inre: config.Origin being/becoming a list of url.URLs
	// (20140727/straup)

	if parsed.String() != config.Origin.String() {
		log.Printf("[%s][%s][handshake] invalid origin, %v", realIP, remoteAddr, parsed)
		return fmt.Errorf("invalid origin")
	}

	log.Printf("[%s][%s][handshake] OK", realIP, remoteAddr)
	return
}

func pubSubHandler(ws *websocket.Conn) {

	remoteAddr := ws.Request().RemoteAddr
	headers := ws.Request().Header

	realIP := headers.Get("X-Real-IP")

	log.Printf("[%s][%s][request] OK", realIP, remoteAddr)

	pubsubClient := redisClient.PubSub()
	defer pubsubClient.Close()

	if err := pubsubClient.Subscribe(redisChannel); err != nil {
		log.Printf("[%s][%s][error] failed to subscribe to pubsub channel %v, because %s", realIP, remoteAddr, redisChannel, err)
		ws.Close()
		return
	}

	log.Printf("[%s][%s][connect] OK", realIP, remoteAddr)

	for ws != nil {

		i, _ := pubsubClient.Receive()

		if msg, _ := i.(*redis.Message); msg != nil {

			// log.Printf("[%s][%s][send] %s", realIP, remoteAddr, msg.Payload)

			var json_blob interface{}
			bytes_blob := []byte(msg.Payload)

			if err := json.Unmarshal(bytes_blob, &json_blob); err != nil {
				log.Printf("[%s][%s][error] failed to parse JSON %v, because %v", realIP, remoteAddr, msg.Payload, err)
				continue
			}

			if err := websocket.JSON.Send(ws, json_blob); err != nil {
				log.Printf("[%s][%s][error] failed to send JSON, because %v", realIP, remoteAddr, err)
				ws.Close()
				break
			}

			log.Printf("[%s][%s][send] OK", realIP, remoteAddr)
		}
	}
}

func main() {

	flag.StringVar(&websocketHost, "ws-host", "127.0.0.1", "Websocket host")
	flag.IntVar(&websocketPort, "ws-port", 8080, "Websocket port")
	flag.StringVar(&websocketRoute, "ws-route", "/", "Websocket route")
	flag.StringVar(&websocketAllowableOrigins, "ws-origin", "", "Websocket allowable origins")

	flag.StringVar(&redisHost, "rs-host", "127.0.0.1", "Redis host")
	flag.IntVar(&redisPort, "rs-port", 6379, "Redis port")
	flag.StringVar(&redisChannel, "rs-channel", "pubsocketd", "Redis channel")

	flag.Parse()

	if websocketAllowableOrigins == "" {
		err, _ := fmt.Printf("Missing allowable Origin (-ws-origin=http://example.com)")
		panic(err)
	}

	// Because I still don't really understand how arrays work in Go...
	// (20140727/straup)

	/*
		allowed := strings.Split(websocketAllowableOrigins, ",")

		for _, test := range allowed {

			test := strings.TrimSpace(test)

			url, err := url.Parse(test)

			if err != nil {
				err, _ := fmt.Printf("Invalid Origin parameter: %v, %v", test, err)
				panic(err)
			}

			log.Printf("%v, %T", url, url)

			omgwtf := []url.URL{ url }
			append(websocketAllowableURLs, omgwtf)
		}
	*/

	websocketEndpoint = fmt.Sprintf("%s:%d", websocketHost, websocketPort)
	redisEndpoint = fmt.Sprintf("%s:%d", redisHost, redisPort)

	redisClient = redis.NewTCPClient(&redis.Options{
		Addr: redisEndpoint,
	})

	defer redisClient.Close()

	// Normally this is the sort of thing you'd expect to do:
	// http.Handle(websocketRoute, websocket.Handler(pubSubHandler))

	// However since we're going to be aggressively paranoid about checking
	// the Origin headers we're going to set up our own websocket Server
	// thingy complete with custom Config and Handshake directive and
	// pass the whole thing off to HandleFunc (20140727/straup)

	// See also:
	// http://www.christian-schneider.net/CrossSiteWebSocketHijacking.html
	// https://code.google.com/p/go/source/browse/websocket/server.go?repo=net

	http.HandleFunc(websocketRoute, pubsocketdHandler)

	log.Printf("[init] listening for websocket requests on " + websocketEndpoint + websocketRoute + ", from " + websocketAllowableOrigins)
	log.Printf("[init] listening for pubsub messages from " + redisEndpoint + " sent to the " + redisChannel + " channel")

	if err := http.ListenAndServe(websocketEndpoint, nil); err != nil {
		err, _ := fmt.Printf("Failed to start websocket server, because %v", err)
		panic(err)
	}
}
