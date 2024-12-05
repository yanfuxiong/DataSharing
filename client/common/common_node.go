package common

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

type IPAddrInfo struct {
	LocalPort  string
	PublicPort string
	PublicIP   string
}

type NodeInfo struct {
	IPAddr IPAddrInfo
	ID     string
}

type ClientInfo struct {
	ID       string
	IpAddr   string
	Platform string
}

type WaitToConnectInfo struct {
	PeerInfo    peer.AddrInfo
	ChExitTimer chan struct{}
}
