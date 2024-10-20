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

		go server.StartServer(net)

		if len(*dest) != 0 {
			log.Printf("Attempting to connect to peer: %s", *dest)
			net.ConnectToPeer(*dest)
		}

		// Run discovery and connection attempts periodically
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Running periodic discovery and connection attempts...")
				net.Discover()
				logPeerInfo(net)
			case <-ctx.Done():
				return
			}
		}

	case <-time.After(60 * time.Second):
		log.Println("Network initialization timed out.")
		cancel()
	}
}

func logPeerInfo(net *node.Network) {
	connectedPeers := net.Node.Network().Peers()
	log.Printf("Connected peers: %d", len(connectedPeers))
	for _, peer := range connectedPeers {
		log.Printf("  - %s", peer.String())
	}
}
