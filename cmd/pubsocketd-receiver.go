package main

import (
	"flag"
	"golang.org/x/net/websocket"
	"log"
)

func main() {

	var url = flag.String("url", "ws://127.0.0.1:8080", "The websocket URL to connect to")
	var origin = flag.String("origin", "", "The origin header to send")

	flag.Parse()

	log.Printf("dialing %s...\n", *url)

	ws, err := websocket.Dial(*url, "", *origin)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("connected to %s and ready to receive new messages\n", *url)

	for {
		var msg = make([]byte, 512)
		_, err = ws.Read(msg)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s\n", msg)
	}
}
