/*
Copyright © 2024 Stephen Owens steve.owens@rightfoot.consulting
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
)

// pingPongCmd represents the launch command
var pingPongCmd = &cobra.Command{
	Use:   "ping-pong",
	Short: "Starts the BBS node for remote peer ping test",
	Long: `If a --remote-peer has been passed on the command line, connect to it
and send it 5 ping messages, otherwise wait for a signal to stop`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("command line args: %v\n", args)
		fmt.Println("ping-pong called")
		remotePeer, err := cmd.Flags().GetString("remote-peer")
		if err != nil {
			panic(err)
		}

		server := &PingPongServer{
			peer: strings.Replace(remotePeer, "//", "/", 1),
		}
		server.run()
	},
}

type PingPongServer struct {
	peer        string
	node        host.Host
	peerInfo    *peerstore.AddrInfo
	addrs       []multiaddr.Multiaddr
	pingService *ping.PingService
}

func (pps *PingPongServer) run() {
	var err error
	pps.node, err = libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Ping(false),
	)
	if err != nil {
		panic(err)
	}

	// configure our own ping protocol
	pps.pingService = &ping.PingService{Host: pps.node}
	pps.node.SetStreamHandler(ping.ID, pps.pingService.PingHandler)

	// print the node's PeerInfo in multiaddr format
	pps.peerInfo = &peerstore.AddrInfo{
		ID:    pps.node.ID(),
		Addrs: pps.node.Addrs(),
	}
	pps.addrs, err = peerstore.AddrInfoToP2pAddrs(pps.peerInfo)
	if err != nil {
		panic(err)
	}
	fmt.Printf("libp2p node address: %v\n", pps.addrs[0])

	// if a remote peer has been passed on the command line, connect to it
	// and send it 5 ping messages, otherwise wait for a signal to stop
	if len(pps.peer) > 1 {
		fmt.Printf("Attempting to ping: %s\n\n", pps.peer)
		pps.ping()
	} else {
		pps.pong()
	}
}

func (pps *PingPongServer) ping() {
	addr, err := multiaddr.NewMultiaddr(pps.peer)
	if err != nil {
		panic(err)
	}
	peer, err := peerstore.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}
	if err := pps.node.Connect(context.Background(), *peer); err != nil {
		panic(err)
	}
	fmt.Printf("sending 5 ping messages to: %s\n", addr.String())
	ch := pps.pingService.Ping(context.Background(), peer.ID)
	for i := 0; i < 5; i++ {
		res := <-ch
		fmt.Println("pinged", addr, "in", res.RTT)
	}
}

func (pps *PingPongServer) pong() {
	fmt.Println("Listening for ping messages use CRTRL-C to quit.")

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Received signal, shutting down...")

	// shut the node down
	if err := pps.node.Close(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(pingPongCmd)
	pingPongCmd.Flags().StringP("remote-peer", "p", "", "Send pings to specified remote peer")
}
