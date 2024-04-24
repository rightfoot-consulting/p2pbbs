/*
Copyright © 2024 Stephen Owens steve.owens@rightfoot.consulting
*/
package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	p2pconfig "github.com/libp2p/go-libp2p/config"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"github.com/rightfoot-consulting/p2pbbs/bbscrypto"
)

var logger = log.Logger("chatnode")

type ChatNode struct {
	Config *Configuration
}

func NewChatNode(config *Configuration) (node *ChatNode, err error) {
	node = &ChatNode{
		Config: config,
	}
	return
}

func (node *ChatNode) Query() {
	config := node.Config
	listenAddresses, err := config.GetListenAddresses()
	if err != nil {
		panic(err)
	}
	if len(listenAddresses) < 1 {
		err = fmt.Errorf("no Listen IPs configured")
		panic(err)
	}
	var sk crypto.PrivKey = nil
	if node.Config.KeyFile != "" {
		sk, err = bbscrypto.LoadPrivateKey(config.KeyFile)
		if err != nil {
			panic(err)
		}
	}
	var options []p2pconfig.Option = []p2pconfig.Option{
		libp2p.ListenAddrs([]multiaddr.Multiaddr(listenAddresses)...),
		libp2p.Identity(sk),
	}
	host, err := libp2p.New(options...)
	if err != nil {
		panic(err)
	}
	ourAddresses := make([]string, len(host.Addrs()))
	for i, addr := range host.Addrs() {
		ourAddresses[i] = addr.String() + "/p2p/" + host.ID().String()
	}
	bootstrapPeers, err := config.GetBootstrapPeers(ourAddresses)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nThis node will listen on the following addresses:\n")
	for _, addr := range ourAddresses {
		fmt.Printf("\t%s\n", addr)
	}
	if sk != nil {
		id, err := peer.IDFromPrivateKey(sk)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\nNode id will be: %s\n", id.String())
	} else {
		fmt.Printf("\nNode id will be random\n")
	}
	fmt.Println("Bootstrap Peers: ")
	for _, addr := range bootstrapPeers {
		fmt.Printf("\t%s\n", addr.String())
	}
}

func (node *ChatNode) Run() {
	log.SetAllLoggers(log.LevelWarn)
	log.SetLogLevel("chatnode", "info")
	config := node.Config

	// libp2p.New constructs a new libp2p Host. Other options can be added
	// here.
	listenAddresses, err := config.GetListenAddresses()
	if err != nil {
		panic(err)
	}
	if len(listenAddresses) < 1 {
		err = fmt.Errorf("no Listen IPs configured")
		panic(err)
	}
	var sk crypto.PrivKey = nil
	if node.Config.KeyFile != "" {
		sk, err = bbscrypto.LoadPrivateKey(config.KeyFile)
		if err != nil {
			panic(err)
		}
	}

	//  Set options for
	var options []p2pconfig.Option = []p2pconfig.Option{
		libp2p.ListenAddrs([]multiaddr.Multiaddr(listenAddresses)...),
		libp2p.Identity(sk),
	}
	host, err := libp2p.New(options...)
	if err != nil {
		panic(err)
	}
	ourAddresses := make([]string, len(host.Addrs()))
	for i, addr := range host.Addrs() {
		ourAddresses[i] = addr.String() + "/p2p/" + host.ID().String()
	}
	logger.Info("Host created. We are:", ourAddresses)
	logger.Info(host.Addrs())

	// Set a function as stream handler. This function is called when a peer
	// initiates a connection and starts a stream with this peer.
	host.SetStreamHandler(protocol.ID(config.ProtocolID), handleStream)

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	ctx := context.Background()
	bsPeers, err := config.GetBootstrapPeers(ourAddresses)
	if err != nil {
		panic(err)
	}
	bootstrapPeers := make([]peer.AddrInfo, len(bsPeers))
	for i, addr := range bsPeers {
		var peerinfo *peer.AddrInfo
		peerinfo, err = peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			err = fmt.Errorf("unable to get address info from address: %v", addr)
			panic(err)
		}
		bootstrapPeers[i] = *peerinfo
	}
	logger.Infof("Final check of bootstrap peers: %v", bootstrapPeers)
	kademliaDHT, err := dht.New(ctx, host, dht.BootstrapPeers(bootstrapPeers...), dht.Mode(dht.ModeAutoServer))
	if err != nil {
		panic(err)
	}
	kademliaDHT.RoutingTable().PeerAdded = func(id peer.ID) {
		logger.Info("DHT Peer %s has been added.")
	}
	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	logger.Debug("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	refreshed := true
	for !refreshed {
		logger.Info("Refreshing the DHT")
		ch := kademliaDHT.RefreshRoutingTable()
		err = <-ch
		if err != nil {
			logger.Errorf("Error refreshing routing table: %v", err)
			// Wait a bit to let bootstrapping finish (really bootstrap should block until it's ready, but that isn't the case yet.)
			time.Sleep(10 * time.Second)
		} else {
			refreshed = true
		}
	}
	time.Sleep(1 * time.Second)

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	logger.Info("Announcing ourselves...")
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)
	logger.Debug("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	logger.Debug("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
	if err != nil {
		panic(err)
	}

	for peer := range peerChan {
		if peer.ID == host.ID() {
			continue
		}
		logger.Debug("Found peer:", peer)

		logger.Debug("Connecting to:", peer)
		stream, err := host.NewStream(ctx, peer.ID, protocol.ID(config.ProtocolID))

		if err != nil {
			logger.Warning("Connection failed:", err)
			continue
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
			go writeData(rw)
			go readData(rw)
		}

		logger.Info("Connected to:", peer)
	}

	select {}
}

func handleStream(stream network.Stream) {
	logger.Info("Got a new stream!")

	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}
