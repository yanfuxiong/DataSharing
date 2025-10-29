package connection

import (
	"context"
	rtkCommon "rtk-cross-share/client/common"
	rtkUtils "rtk-cross-share/client/utils"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

const (
	ctxTimeout_normal = 5 * time.Second
	ctxTimeout_short  = 2 * time.Second
)

type FileDropItemStreamInfo struct {
	Timestamp uint64
	ID        string
	StreamId  string // stream ID
}

type (
	callbackStartProcessForPeerFunc     func(ctx context.Context, id, ipAddr string) func(source rtkCommon.CancelBusinessSource)
	callbackSendDisconnectMsgToPeerFunc func(id string)
)

var (
	node          host.Host
	fileTransNode host.Host
	nodeMutex     sync.RWMutex

	pingServer *ping.PingService
	cg         *rtkUtils.CondGroup

	// mutexMap by ID
	mutexMap sync.Map

	MdnsStartTime = int64(0) // mdns services start time stamp

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

func CondGroupAdd() {
	cg.Add(1)
}

func CondGroupDone() {
	cg.Done()
}

func wait() {
	cg.Wait()
}
