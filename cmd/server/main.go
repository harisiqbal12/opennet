package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/harisiqbal12/opennet/pkg/node"
	"github.com/harisiqbal12/opennet/pkg/server"
)

func main() {
	fmt.Println("Starting Opennet!!")

	dest := flag.String("address", "", "Destination multiaddr string")
	flag.Parse()

	log.Printf("Destination: %s", *dest)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	networkChan := make(chan *node.Network)

	opts := node.NetworkOpts{
		Context:   ctx,
		Networkch: networkChan,
	}

	go node.CreateNetwork(opts)

	select {
	case net := <-networkChan:
		fmt.Println("Network initialized!")
		server.StartServer(net)

	case <-time.After(60 * time.Second):
		log.Println("Network initialization timed out.")
		cancel()
	}
}
