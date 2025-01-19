package connection

import (
	"context"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"log"
	rtkUtils "rtk-cross-share/utils"
	"sync"
)

type streamInfo struct {
	s       network.Stream
	isAlive bool
}

var (
	streamPoolMap   = make(map[string](streamInfo))
	streamPoolMutex sync.RWMutex
)

func CheckAllStreamAlive(ctx context.Context) {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()

	if !IsNetworkConnected() {
		log.Printf("[%s] the Network is unavailable!", rtkUtils.GetFuncInfo())
		return
	}

	for _, sInfo := range streamPoolMap {
		pingResult := <-pingServer.Ping(ctx, sInfo.s.Conn().RemotePeer())
		ipAddr := rtkUtils.GetRemoteAddrFromStream(sInfo.s)
		if pingResult.Error != nil || !sInfo.isAlive {
			if pingResult.Error != nil {
				log.Printf("[%s] Ip[%s] Ping err:%+v, Retry connect", rtkUtils.GetFuncInfo(), ipAddr, pingResult.Error)
			} else {
				log.Printf("[%s] Ip[%s] stream is close , Retry open a new stream", rtkUtils.GetFuncInfo(), ipAddr)
			}

			reConnectInfo := ReConnectPeerInfo{
				Peer: peer.AddrInfo{
					ID:    sInfo.s.Conn().RemotePeer(),
					Addrs: []ma.Multiaddr{sInfo.s.Conn().RemoteMultiaddr()},
				},
				RetryCount: 0,
				MaxCount:   retryConnection,
				DelayTime:  getDelayTime(sInfo.s.Conn().RemotePeer()),
			}
			reConnectPeerChan <- reConnectInfo
		} else {
			log.Printf("[%s] id[%s] Ip[%s]  RTT [%d]ms", rtkUtils.GetFuncInfo(), sInfo.s.Conn().RemotePeer().String(), ipAddr, pingResult.RTT.Milliseconds())
		}
	}
}

func UpdateStream(id string, stream network.Stream) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)
	if oldSinfo, ok := streamPoolMap[id]; ok {
		if rtkUtils.GetRemoteAddrFromStream(oldSinfo.s) != ipAddr {
			oldSinfo.s.Reset()
		}
		log.Printf("[%s] UpdateStream id:%s Stream existed, ip[%s] alive:[%+v]", rtkUtils.GetFuncInfo(), id, ipAddr, oldSinfo.isAlive)
	}

	streamPoolMap[id] = streamInfo{
		s:       stream,
		isAlive: true,
	}

	log.Printf("UpdateStream id:[%s]  ip:[%s]", id, ipAddr)
}

func GetStream(id string) (network.Stream, bool) {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		if sInfo.isAlive { // only return a live stream
			return sInfo.s, ok
		}
	}

	return nil, false
}

func GetStreamIpAddr(id string) string {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		return rtkUtils.GetRemoteAddrFromStream(sInfo.s)
	}
	return "UnknownIp"
}

func AddStream(id string, pStream network.Stream) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		sInfo.s.Close() //attention: It will cause all stream closed
		log.Printf("[%s %d] Stream already existed, close first. id:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), id)
	}
	streamPoolMap[id] = streamInfo{
		s:       pStream,
		isAlive: true,
	}
}

func CloseStream(id string) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		sInfo.s.Reset() //Need to immediately notify the remote side to close
		//sInfo.s.Close()
		sInfo.isAlive = false
		streamPoolMap[id] = sInfo
		log.Println("CloseStream id:", id)
	} else {
		log.Printf("[%s %d] Err: Unknown stream of id:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), id)
	}
}

func ClosePeer(id string) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		sInfo.s.Reset()
		node.Network().ClosePeer(sInfo.s.Conn().RemotePeer())
		log.Println("ClosePeer id:", id)
	} else {
		log.Printf("[%s] Err: Unknown stream of id:%s", rtkUtils.GetFuncInfo(), id)
	}
}

func OfflineStream(id string) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		sInfo.s.Reset()
		node.Network().ClosePeer(sInfo.s.Conn().RemotePeer())
		delete(streamPoolMap, id)
		log.Printf("OfflineStream id:[%s] ip[%s]", id, rtkUtils.GetRemoteAddrFromStream(sInfo.s))
	} else {
		log.Printf("[%s] Err: Unknown stream of id:%s", rtkUtils.GetFuncInfo(), id)
	}
}

func IsStreamExisted(id string) bool {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		if sInfo.isAlive {
			return sInfo.s.Conn().RemotePeer().String() != ""
		}
	}
	return false
}

func CancelStreamPool() {
	log.Printf("CancelStreamPool")
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()
	for id, sInfo := range streamPoolMap {
		sInfo.s.Reset()
		node.Network().ClosePeer(sInfo.s.Conn().RemotePeer())
		delete(streamPoolMap, id)
	}
}

func PrintfStreamPool() {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	for k, v := range streamPoolMap {
		log.Printf("key:[%+v] stream:[%+v]", k, v)
	}
}
