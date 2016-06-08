package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/redis.v1"
	"log"
	"os"
)

type Message struct {
	Data string
}

func main() {

	var host = flag.String("host", "127.0.0.1", "Redis host")
	var port = flag.Int("port", 6379, "Redis port")
	var channel = flag.String("channel", "pubsocketd", "Redis channel")

	flag.Parse()

	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("connecting to %s...\n", endpoint)

	client := redis.NewTCPClient(&redis.Options{
		Addr: endpoint,
	})

	defer client.Close()

	buf := bufio.NewScanner(os.Stdin)

	log.Printf("connected to %s and ready to send new messages\n", endpoint)

	for buf.Scan() {

		txt := buf.Text()

		msg := Message{Data: txt}
		body, err := json.Marshal(msg)

		if err != nil {
			log.Fatal(err)
		}

		client.Publish(*channel, string(body))
	}

	os.Exit(0)
}
