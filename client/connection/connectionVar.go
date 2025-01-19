package connection

import (
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"sync"
	"time"
)

const (
	retryConnection = 3

	// libp2p default backoff time is 5 seccond
	retryDelay = 5 * time.Second
)

var (
	node       host.Host
	pingServer *ping.PingService

	// mutexMap by ID
	mutexMap sync.Map

	StartProcessChan = make(chan string)
	EndProcessChan   = make(chan string)

	reConnectPeerChan = make(chan ReConnectPeerInfo, 5)

	mdnsPeerChan = make(chan peer.AddrInfo)
)

type ReConnectPeerInfo struct {
	Peer       peer.AddrInfo
	RetryCount uint8
	MaxCount   uint8
	DelayTime  time.Duration
}
