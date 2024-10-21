package node

import (
	"log"
	"sync"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// NetworkNotifier implements the network.Notifiee interface
type NetworkNotifier struct {
	network *Network
	peers   map[peer.ID]network.Conn
	mu      sync.RWMutex
}

func NewNetworkNotifier(net *Network) *NetworkNotifier {
	return &NetworkNotifier{
		network: net,
		peers:   make(map[peer.ID]network.Conn),
	}
}

func (n *NetworkNotifier) Connected(net network.Network, conn network.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()

	remotePeer := conn.RemotePeer()
	remoteMultiaddr := conn.RemoteMultiaddr()

	n.peers[remotePeer] = conn

	log.Printf("➕ New peer connected:\n")
	log.Printf("   Peer ID: %s\n", remotePeer.String())
	log.Printf("   Address: %s\n", remoteMultiaddr.String())
	log.Printf("   Total peers: %d\n", len(n.peers))
}

func (n *NetworkNotifier) Disconnected(net network.Network, conn network.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()

	remotePeer := conn.RemotePeer()
	delete(n.peers, remotePeer)

	log.Printf("➖ Peer disconnected:\n")
	log.Printf("   Peer ID: %s\n", remotePeer.String())
	log.Printf("   Remaining peers: %d\n", len(n.peers))
}

// Get current connected peers
func (n *NetworkNotifier) GetConnectedPeers() []peer.ID {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peers := make([]peer.ID, 0, len(n.peers))
	for p := range n.peers {
		peers = append(peers, p)
	}
	return peers
}

// Required interface methods (can be left empty if not needed)
func (n *NetworkNotifier) Listen(net network.Network, multiaddr ma.Multiaddr)      {}
func (n *NetworkNotifier) ListenClose(net network.Network, multiaddr ma.Multiaddr) {}
func (n *NetworkNotifier) OpenedStream(net network.Network, stream network.Stream) {}
func (n *NetworkNotifier) ClosedStream(net network.Network, stream network.Stream) {}
