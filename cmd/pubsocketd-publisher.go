package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/redis.v1"
	"os"
)

func main() {

	var host = flag.String("host", "127.0.0.1", "Redis host")
	var port = flag.Int("port", 6379, "Redis port")
	var channel = flag.String("channel", "", "Redis channel")

	flag.Parse()

	endpoint := fmt.Sprintf("%s:%d", *host, *port)

	client := redis.NewTCPClient(&redis.Options{
		Addr: endpoint,
	})

	defer client.Close()

	buf := bufio.NewScanner(os.Stdin)

	for buf.Scan() {
		txt := buf.Text()
		client.Publish(*channel, txt)
	}

	os.Exit(0)
}
