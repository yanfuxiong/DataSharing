package connection

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

const (
	pingInterval      = 3500 * time.Millisecond
	pingTimeout       = 3000 * time.Millisecond
	pingErrMaxCnt     = 3
	ctxTimeout_normal = 5 * time.Second
	ctxTimeout_short  = 2 * time.Second
)

type (
	callbackStartProcessForPeerFunc     func(id, ipAddr string) func()
	callbackSendDisconnectMsgToPeerFunc func(id string)
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

	callbackStartProcessForPeer     callbackStartProcessForPeerFunc
	callbackSendDisconnectMsgToPeer callbackSendDisconnectMsgToPeerFunc
)

func SetStartProcessForPeerCallback(cb callbackStartProcessForPeerFunc) {
	callbackStartProcessForPeer = cb
}

func SetSendDisconnectMsgToPeerCallback(cb callbackSendDisconnectMsgToPeerFunc) {
	callbackSendDisconnectMsgToPeer = cb
}
