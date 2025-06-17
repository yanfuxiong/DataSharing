package connection

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

const (
	retryConnection = 3

	// libp2p default backoff time is 5 seccond
	retryDelay    = 5 * time.Second
	pingInterval  = 8 * time.Second // TODO: refine this. Expect to detect disconnection in 10 seconds
	pingTimeout   = 2 * time.Second
	pingErrMaxCnt = 3
)

var (
	node      host.Host
	nodeMutex sync.RWMutex

	pingServer *ping.PingService

	// mutexMap by ID
	mutexMap sync.Map

	MdnsStartTime = int64(0) // mdns services start time stamp

	mdnsPeerChan            = make(chan peer.AddrInfo)
	mdnsNoticeNetworkStatus = make(chan bool)

	noticeFmtTypeSteamReadyChanMap sync.Map

	callbackStartProcessForPeer func(id string)
	callbackStopProcessForPeer  func(id string)
)

type (
	callbackStartProcessForPeerFunc func(id string)
	callbackStopProcessForPeerFunc  func(id string)
)

func SetStartProcessForPeerCallback(cb callbackStartProcessForPeerFunc) {
	callbackStartProcessForPeer = cb
}

func SetStopProcessForPeerCallback(cb callbackStopProcessForPeerFunc) {
	callbackStopProcessForPeer = cb
}
