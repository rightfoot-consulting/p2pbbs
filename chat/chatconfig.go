/*
Copyright © 2024 Stephen Owens steve.owens@rightfoot.consulting
*/
package chat

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	maddr "github.com/multiformats/go-multiaddr"
)

type Configuration struct {
	Port             int      `json:"port"`
	RendezvousString string   `json:"redezvous_string"`
	BootstrapPeers   []string `json:"bootstrap_peers"`
	ListenIps        []string `json:"listen_ips"`
	ProtocolID       string   `json:"protocol_id"`
	KeyFile          string   `json:"key_file"`
}

func LoadChatConfig(filename string) (config *Configuration, err error) {
	var cfg Configuration
	jsonBytes, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonBytes, &cfg)
	if err == nil {
		config = &cfg
	}
	return
}

func (cfg *Configuration) GetBootstrapPeers(exclude []string) (peers []maddr.Multiaddr, err error) {
	peers = make([]maddr.Multiaddr, 0, len(cfg.BootstrapPeers))
	for _, addrString := range cfg.BootstrapPeers {
		if addrString == "" {
			continue
		}
		var addr maddr.Multiaddr
		addrString := strings.Replace(addrString, "//", "/", 1)
		addr, err = maddr.NewMultiaddr(strings.Replace("//", "/", addrString, 1))
		if err != nil {
			peers = nil
			return
		}
		// Don't add self to bootstrap peers
		excluded := false
		for _, e := range exclude {
			if e == addr.String() {
				excluded = true
				break
			}
		}
		if !excluded {
			logger.Infof("Adding bootstrap peer: %s", addr.String())
			peers = append(peers, addr)
		}
	}
	return
}

func (cfg *Configuration) GetListenAddresses() (addresses []maddr.Multiaddr, err error) {
	addresses = make([]maddr.Multiaddr, 0)
	ipList := cfg.ListenIps
	if len(ipList) < 1 {
		fmt.Printf("Configuration has no IPs for listening, defaulting to 0.0.0.0\n")
		ipList = append(ipList, "0.0.0.0")
	}
	fmt.Printf("iplist: %v\n", ipList)
	for _, ipString := range ipList {
		addrString := fmt.Sprintf("/ip4/%s/tcp/%d", ipString, cfg.Port)
		var addr maddr.Multiaddr
		addr, err = maddr.NewMultiaddr(addrString)
		if err != nil {
			addresses = nil
			return
		}
		addresses = append(addresses, addr)
	}
	return
}
