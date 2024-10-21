package node

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

const service string = "opennet"
const MessageTopic string = "messages"
const ConnectTopic string = "connection"
const ConnectionSuccess string = "connectionEstablished"

type NetworkOpts struct {
	Context   context.Context
	Networkch chan *Network
}

type Message struct {
	From    string `json:"from"`
	Content string `json:"content"`
	Topic   string `json:"topic"`
}

type Network struct {
	Node    host.Host
	ctx     context.Context
	dht     *dht.IpfsDHT
	pubsub  *pubsub.PubSub
	topics  map[string]*pubsub.Topic
	subs    map[string]*pubsub.Subscription
	msgChan chan *Message
}

func CreateNetwork(opts NetworkOpts) *Network {
	network := &Network{
		ctx:     opts.Context,
		topics:  make(map[string]*pubsub.Topic),
		subs:    make(map[string]*pubsub.Subscription),
		msgChan: make(chan *Message, 100),
	}

	defer network.start(opts.Networkch)
	return network
}

func (n *Network) initPubSub() error {

	if n.Node == nil {
		return fmt.Errorf("node must be initialized before pubsub")
	}

	ps, err := pubsub.NewGossipSub(n.ctx, n.Node)

	if err != nil {
		return fmt.Errorf("failed to create pubsub: %w", err)
	}

	n.pubsub = ps
	n.topics = make(map[string]*pubsub.Topic)
	n.subs = make(map[string]*pubsub.Subscription)
	n.msgChan = make(chan *Message, 100)

	topics := []string{MessageTopic, ConnectTopic, ConnectionSuccess}

	for _, topic := range topics {
		if err := n.JoinTopic(topic); err != nil {
			return fmt.Errorf("failed to join topic: %w", err)
		}
	}

	return nil
}
func (n *Network) JoinTopic(topicName string) error {
	topic, err := n.pubsub.Join(topicName)
	if err != nil {
		return err
	}

	sub, err := topic.Subscribe()

	if err != nil {
		return err
	}

	n.topics[topicName] = topic
	n.subs[topicName] = sub

	go n.handleSubscription(sub)

	return nil
}

func (n *Network) handleSubscription(sub *pubsub.Subscription) {
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			msg, err := sub.Next(n.ctx)
			if err != nil {
				log.Printf("Error getting next message: %s", err)
				continue
			}

			if msg.GetTopic() == ConnectTopic {
				if err := n.connectToPeer(string(msg.Data)); err != nil {
					log.Printf("error occured during connecting to peer: %s", err)
				}

				log.Printf("connected to a peer %s", string(msg.Data))

				// log.Printf("Connection notification: %s", string(msg.Data))
				continue
			}

			// Skip messages from ourselves
			if msg.ReceivedFrom == n.Node.ID() {
				continue
			}

			if msg.GetTopic() == ConnectionSuccess {
				log.Println(string(msg.Data))
				continue
			}

			message := &Message{
				From:    msg.ReceivedFrom.String(),
				Content: string(msg.Data),
				Topic:   sub.Topic(),
			}

			select {
			case n.msgChan <- message:
				log.Printf("Received message from %s: %s", message.From, message.Content)
			default:
				log.Printf("Message channel full, dropping message from %s", message.From)
			}
		}
	}
}

func (n *Network) PublishMessage(topicName string, content string) error {
	topic, exists := n.topics[topicName]

	if !exists {
		return fmt.Errorf("topic does not exists %s", topicName)
	}

	return topic.Publish(n.ctx, []byte(content))
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
		libp2p.NoTransports,
		libp2p.Transport(tcp.NewTCPTransport),
	)

	if err != nil {
		log.Fatal(err)
	}

	// node.Network().Notify()

	notifier := NewNetworkNotifier(n)
	node.Network().Notify(notifier)

	for _, addr := range node.Addrs() {
		log.Printf("Listening on %s/p2p/%s \n", addr, node.ID().ShortString())
	}

	log.Printf("Node ID %s\n", node.ID().Loggable())

	kademliaDHT, err := dht.New(n.ctx, node, dht.Mode(dht.ModeClient))

	if err != nil {
		log.Fatalf("failed to create dht %s", err)
	}

	// if err := kademliaDHT.Bootstrap(n.ctx); err != nil {
	// 	log.Fatalf("failed to bootstrap dht %s", err)
	// }

	displayNodeAddress(node)

	n.Node = node
	n.dht = kademliaDHT

	if err := n.initPubSub(); err != nil {
		log.Fatalf("failed to initiliazed pubsub %s \n", err)
	}

	// go n.connectToBootstrap()
	// go n.startDiscovery()

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

// func (n *Network) connectToBootstrap() {
// 	log.Printf("connecting bootstrap \n")
// 	bootstrapPeers := dht.DefaultBootstrapPeers

// 	var wg sync.WaitGroup

// 	for _, peerAddr := range bootstrapPeers {
// 		peerInfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()

// 			if peerInfo.ID != n.Node.ID() {

// 				if err := n.Node.Connect(n.ctx, *peerInfo); err != nil {
// 					fmt.Printf("failed to connect %s \n", err)
// 				}

// 				log.Printf("connected to bootstrap peer: %s\n", peerInfo.ID)
// 			} else {
// 				log.Fatalf("skipping conencting to itself")
// 			}

// 		}()
// 	}

// 	wg.Wait()
// }

// func (n *Network) startDiscovery() {
// 	log.Printf("starting discovery \n")

// 	routingDiscovery := routing.NewRoutingDiscovery(n.dht)
// 	routingDiscovery.Advertise(n.ctx, service)

// 	ticker := time.NewTicker(time.Second * 5)

// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-n.ctx.Done():
// 			return
// 		case <-ticker.C:
// 			peers, err := routingDiscovery.FindPeers(n.ctx, service)
// 			if err != nil {
// 				log.Printf("error finding peers %s \n", err)
// 				continue
// 			}

// 			log.Printf("total peers %d: ", len(peers))
// 			log.Printf("total peers in node %d: ", len(n.Node.Network().Peers()))

// 			for peer := range peers {
// 				if peer.ID == n.Node.ID() {
// 					continue
// 				}

// 				if len(peer.Addrs) == 0 {
// 					log.Printf("peer %s has no addresses", peer.ID)
// 					continue
// 				}

// 				err := n.Node.Connect(n.ctx, peer)

// 				if err != nil {
// 					log.Printf("error connecting to peer %s", err)
// 				} else {
// 					log.Printf("connected to peer %s", peer.ID.ShortString())
// 				}

// 			}

// 		}
// 	}

// }

// func (n *Network) Discover() {
// 	routingDiscovery := routing.NewRoutingDiscovery(n.dht)
// 	routingDiscovery.Advertise(n.ctx, service)

// 	peers, err := routingDiscovery.FindPeers(n.ctx, service)

// 	if err != nil {
// 		log.Printf("error finding peers %s", err)
// 	}

// 	for peer := range peers {
// 		if peer.ID == n.Node.ID() {
// 			continue
// 		}

// 		if err := n.Node.Connect(n.ctx, peer); err != nil {
// 			log.Printf("failed to connect to a peer %s", err)
// 		} else {
// 			log.Printf("connected with a peer")
// 		}
// 	}
// }

func (n *Network) connectToPeer(peerAddr string) error {
	fmt.Printf("connect to peer \n")
	multiAddr, err := multiaddr.NewMultiaddr(peerAddr)

	if err != nil {
		return err
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)

	if err != nil {
		return err
	}

	err = n.Node.Connect(n.ctx, *peerInfo)

	if err != nil {
		return err
	}

	log.Printf("Successfully connected to peer: %s \n", peerInfo.ID)

	return nil
}
