package main

import (
	// "code.google.com/p/go.net/websocket"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"gopkg.in/redis.v1"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	logger                    *log.Logger
	logFile                   string
	redisHost                 string
	redisPort                 int
	redisChannel              string
	redisEndpoint             string
	websocketHost             string
	websocketPort             int
	websocketEndpoint         string
	websocketRoute            string
	websocketAllowInsecure    bool
	websocketAllowableOrigins string
	websocketAllowableURLs    []url.URL
	websocketHeartbeat        bool
	redisClient               *redis.Client
	tlsCert                   string
	tlsKey                    string
)

// What I'd really like to do is pass in a list of allowable Origin
// URLs but the signature for websocket.Config only allows a single
// url.URL thingy and trying to subclass it resulted in a cascading
// series of "unexpected yak" style errors from the compilers. It
// seems like something that must be possible but my Go-fu is still
// weak... (20140729/straup)

// From Crowley (20140729):
// As for your multiple origins problem:  I recommend you construct
// a map[string]*websocket.Server mapping your full set of origins.
// Do that before you start serving traffic and then you can access
// the map without locks.  Each of those *websocket.Server values
// implements the http.Handler interface so you can in your handler
// look up the right one for the origin and call its ServeHTTP method.

// type pubsocketdConfig struct {
//     websocket.Config
//     Origin []url.URL
// }

func pubsocketdHandler(w http.ResponseWriter, req *http.Request) {

	// See above (20140729/straup)

	if websocketAllowInsecure {

		s := websocket.Server{
			Handler: websocket.Handler(pubSubHandler),
		}

		s.ServeHTTP(w, req)

	} else {

		originURL := websocketAllowableURLs[0]

		pubsocketdConfig := websocket.Config{Origin: &originURL}

		s := websocket.Server{
			Config:    pubsocketdConfig,
			Handshake: pubsocketdHandshake,
			Handler:   websocket.Handler(pubSubHandler),
		}

		s.ServeHTTP(w, req)
	}

}

func pubsocketdHandshake(config *websocket.Config, req *http.Request) (err error) {

	remoteAddr := req.RemoteAddr
	headers := req.Header

	origin := headers.Get("Origin")
	realIP := headers.Get("X-Real-IP")

	if origin == "" {
		logger.Printf("[%s][%s][handshake] missing origin", realIP, remoteAddr)
		return fmt.Errorf("missing origin")
	}

	parsed, err := url.Parse(origin)

	if err != nil {
		logger.Printf("[%s][%s][handshake] failed to parse origin, %v", realIP, remoteAddr, origin)
		return fmt.Errorf("failed to parse origin")
	}

	// See above inre: config.Origin being/becoming a list of url.URLs
	// (20140727/straup)

	if parsed.String() != config.Origin.String() {
		logger.Printf("[%s][%s][handshake] invalid origin, expected %v but got %v", realIP, remoteAddr, config.Origin, parsed)
		return fmt.Errorf("invalid origin")
	}

	logger.Printf("[%s][%s][handshake] OK", realIP, remoteAddr)
	return
}

func pubsocketHeartbeat(ws *websocket.Conn) {
	ticker := time.NewTicker(time.Second * 30)
	json_blob := "{\"heartbeat\": 1}"
	for t := range ticker.C {
		if err := websocket.JSON.Send(ws, json_blob); err != nil {
			logger.Printf("[error] failed to send heartbeat (%v) JSON, because %v", t, err)
			// ws.Close()
			break
		}
	}
}

func pubSubHandler(ws *websocket.Conn) {

	if websocketHeartbeat {
		go pubsocketHeartbeat(ws)
	}

	remoteAddr := ws.Request().RemoteAddr
	headers := ws.Request().Header

	realIP := headers.Get("X-Real-IP")

	logger.Printf("[%s][%s][request] OK", realIP, remoteAddr)

	pubsubClient := redisClient.PubSub()
	defer pubsubClient.Close()

	if err := pubsubClient.Subscribe(redisChannel); err != nil {
		logger.Printf("[%s][%s][error] failed to subscribe to pubsub channel %v, because %s", realIP, remoteAddr, redisChannel, err)
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
				logger.Printf("[%s][%s][error] failed to parse JSON %v, because %v", realIP, remoteAddr, msg.Payload, err)
				continue
			}

			if err := websocket.JSON.Send(ws, json_blob); err != nil {
				logger.Printf("[%s][%s][error] failed to send JSON, because %v", realIP, remoteAddr, err)
				ws.Close()
				break
			}

			logger.Printf("[%s][%s][send] OK", realIP, remoteAddr)
		}
	}
}

func main() {

	flag.StringVar(&websocketHost, "ws-host", "127.0.0.1", "WebSocket host")
	flag.IntVar(&websocketPort, "ws-port", 8080, "WebSocket port")
	flag.StringVar(&websocketRoute, "ws-route", "/", "WebSocket route")
	flag.StringVar(&websocketAllowableOrigins, "ws-origin", "", "WebSocket allowable origin(s)")
	flag.BoolVar(&websocketAllowInsecure, "ws-insecure", false, "Allow WebSocket server to run in insecure mode")
	flag.BoolVar(&websocketHeartbeat, "ws-heartbeat", false, "Keep WebSocket alive with heartbeat messages")
	flag.StringVar(&redisHost, "rs-host", "127.0.0.1", "Redis host")
	flag.IntVar(&redisPort, "rs-port", 6379, "Redis port")
	flag.StringVar(&redisChannel, "rs-channel", "pubsocketd", "Redis channel")
	flag.StringVar(&logFile, "ps-log-file", "", "Log requests to this file")
	flag.StringVar(&tlsCert, "tls-cert", "", "Path to TLS certificate")
	flag.StringVar(&tlsKey, "tls-key", "", "Path to TLS key")

	flag.Parse()

	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)

		if err != nil {
			panic(err)
		}

		multi := io.MultiWriter(file, os.Stdout)
		logger = log.New(multi, "[pubsocketd] ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		multi := io.MultiWriter(os.Stdout)
		logger = log.New(multi, "[pubsocketd] ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	if !websocketAllowInsecure {
		if websocketAllowableOrigins == "" {
			logger.Fatalf("Missing allowable Origin (-ws-origin=http://example.com)")
		}

		allowed := strings.Split(websocketAllowableOrigins, ",")
		count := len(allowed)

		if count > 1 {
			logger.Fatalf("Only one origin server is supported at the moment")
		}

		websocketAllowableURLs = make([]url.URL, 0, count)

		for _, test := range allowed {

			test := strings.TrimSpace(test)

			url, err := url.Parse(test)

			if err != nil {
				logger.Fatalf("Invalid Origin parameter: %v, %v", test, err)
			}

			websocketAllowableURLs = append(websocketAllowableURLs, *url)
		}
	}

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

	if websocketAllowInsecure {
		logger.Printf("[init] listening for websocket requests on " + websocketEndpoint + websocketRoute + ", in INSECURE MODE which is not advised for production use")

	} else {
		logger.Printf("[init] listening for websocket requests on " + websocketEndpoint + websocketRoute + ", from " + websocketAllowableOrigins)
	}

	logger.Printf("[init] listening for pubsub messages from " + redisEndpoint + " sent to the " + redisChannel + " channel")

	if tlsCert != "" && tlsKey != "" {
		if err := http.ListenAndServeTLS(websocketEndpoint, tlsCert, tlsKey, nil); err != nil {
			logger.Fatalf("Failed to start websocket server, because %v", err)
		}
	} else {
		if err := http.ListenAndServe(websocketEndpoint, nil); err != nil {
			logger.Fatalf("Failed to start websocket server, because %v", err)
		}
	}
}
