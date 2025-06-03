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
	pingInternal  = 3 * time.Second
	pingTimeout   = 2 * time.Second
	pingErrMaxCnt = 3
)

var (
	node      host.Host
	nodeMutex sync.RWMutex

	pingServer *ping.PingService

	// mutexMap by ID
	mutexMap sync.Map

	StartProcessChan = make(chan string)
	EndProcessChan   = make(chan string)
	CancelAllProcess = make(chan struct{})
	MdnsStartTime    = int64(0) // mdns services start time stamp

	mdnsPeerChan            = make(chan peer.AddrInfo)
	mdnsNoticeNetworkStatus = make(chan bool)

	noticeFmtTypeSteamReadyChanMap sync.Map
)
