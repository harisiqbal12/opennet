package node

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	routing "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
)

const service string = "opennet"

type NetworkOpts struct {
	Context   context.Context
	Networkch chan *Network
}

type Network struct {
	Node host.Host
	ctx  context.Context
	dht  *dht.IpfsDHT
}

func CreateNetwork(opts NetworkOpts) *Network {
	network := &Network{
		ctx: opts.Context,
	}
	defer network.start(opts.Networkch)
	return network
}

func (n *Network) start(networkch chan *Network) {

	conn, err := connmgr.NewConnManager(0, 500, connmgr.WithGracePeriod(time.Minute))

	if err != nil {
		log.Fatalf("error initializing a conn manager %s", err)
	}

	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.ConnectionManager(conn),
		libp2p.NATPortMap(),
		libp2p.EnableRelay(),
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("im here \n")

	for _, addr := range node.Addrs() {
		log.Printf("Listening on %s/p2p/%s \n", addr, node.ID().ShortString())
	}

	log.Printf("Node ID %s\n", node.ID().Loggable())

	kademliaDHT, err := dht.New(n.ctx, node)

	if err != nil {
		log.Fatalf("failed to create dht %s", err)
	}

	if err := kademliaDHT.Bootstrap(n.ctx); err != nil {
		log.Fatalf("failed to bootstrap dht %s", err)
	}

	displayNodeAddress(node)

	go n.connectToBootstrap()
	go n.startDiscovery()

	n.Node = node
	n.dht = kademliaDHT

	networkch <- n
}

func displayNodeAddress(node host.Host) {
	for _, addr := range node.Addrs() {
		fullAddr := fmt.Sprintf("%s/p2p/%s", addr, node.ID().ShortString())
		log.Printf("Node address: %s", fullAddr)
		fmt.Printf("\nTo connect to this node, run the program with:\n")
		fmt.Printf("-address %s\n\n", fullAddr)
	}
}

func (n *Network) connectToBootstrap() {
	log.Printf("connecting bootstrap \n")
	bootstrapPeers := dht.DefaultBootstrapPeers

	var wg sync.WaitGroup

	for _, peerAddr := range bootstrapPeers {
		peerInfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()

			if peerInfo.ID != n.Node.ID() {

				if err := n.Node.Connect(n.ctx, *peerInfo); err != nil {
					fmt.Printf("failed to connect %s \n", err)
				}

				log.Printf("connected to bootstrap peer: %s\n", peerInfo.ID)
			} else {
				log.Fatalf("skipping conencting to itself")
			}

		}()
	}

	wg.Wait()
}

func (n *Network) startDiscovery() {
	log.Printf("starting discovery \n")

	routingDiscovery := routing.NewRoutingDiscovery(n.dht)
	routingDiscovery.Advertise(n.ctx, service)

	ticker := time.NewTicker(time.Second * 5)

	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			peers, err := routingDiscovery.FindPeers(n.ctx, service)
			if err != nil {
				log.Printf("error finding peers %s \n", err)
				continue
			}

			log.Printf("total peers %d: ", len(peers))
			log.Printf("total peers in node %d: ", len(n.Node.Network().Peers()))

			for peer := range peers {
				if peer.ID == n.Node.ID() {
					continue
				}

				if len(peer.Addrs) == 0 {
					log.Printf("peer %s has no addresses", peer.ID)
					continue
				}

				err := n.Node.Connect(n.ctx, peer)

				if err != nil {
					log.Printf("error connecting to peer %s", err)
				} else {
					log.Printf("connected to peer %s", peer.ID.ShortString())
				}

			}

		}
	}

	// routing.DiscoveryRouting.Advertise()
}

// func (n *Network) Discovery() error {
// 	n.Node.Connect()
// }

func (n *Network) Discover() {
	routingDiscovery := routing.NewRoutingDiscovery(n.dht)
	routingDiscovery.Advertise(n.ctx, service)

	peers, err := routingDiscovery.FindPeers(n.ctx, service)

	if err != nil {
		log.Printf("error finding peers %s", err)
	}

	for peer := range peers {
		if peer.ID == n.Node.ID() {
			continue
		}

		if err := n.Node.Connect(n.ctx, peer); err != nil {
			log.Printf("failed to connect to a peer %s", err)
		} else {
			log.Printf("connected with a peer")
		}
	}
}

func (n *Network) ConnectToPeer(peerAddr string) {
	fmt.Printf("connect to peer \n")
	multiAddr, err := multiaddr.NewMultiaddr(peerAddr)

	if err != nil {
		log.Fatalf("Invalid multiaddress: %s", err)
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		log.Fatalf("Failed to get peer info: %s", err)
	}

	err = n.Node.Connect(n.ctx, *peerInfo)
	if err != nil {
		log.Printf("Failed to connect to peer: %s", err)
	} else {
		log.Printf("Successfully connected to peer: %s", peerInfo.ID)
	}
}
