package connection

import (
	"context"
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
)

type streamInfo struct {
	s network.Stream
	//isAlive    bool		//remove  param "isAlive", Reconnection triggered by LanServer
	timeStamp  int64
	pingErrCnt int

	sFileDrop network.Stream
	sImage    network.Stream
}

var (
	streamPoolMap   = make(map[string](streamInfo))
	streamPoolMutex sync.RWMutex
)

func CheckAllStreamAlive(ctx context.Context) {
	pingFailFunc := func(key string, sInfo streamInfo) {
		if CheckStreamReset(key, sInfo.timeStamp) { // if stream is update, not need go through this flow
			return
		}

		var pingErrCnt = sInfo.pingErrCnt + 1
		updateStreamPingErrCntInternal(key, pingErrCnt)
		if pingErrCnt < pingErrMaxCnt {
			return
		}

		// FIXME: It cannot be offline immediately due to the param "isAlive"
		// Invoke OffLineStream make the client remove from map
		offlineEvent(sInfo.s)
	}

	tempStreamMap := make(map[string](streamInfo))
	streamPoolMutex.RLock()
	for key, sInfo := range streamPoolMap {
		tempStreamMap[key] = sInfo
	}
	streamPoolMutex.RUnlock()

	for key, sInfo := range tempStreamMap {
		rtkMisc.GoSafeWithParam(func(args ...any) {
			// Default timeout is 10 sec in Ping.go
			// Use this context timeout instead of the timeout in Ping.go
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			defer cancel()
			select {
			case pingResult := <-pingServer.Ping(pingCtx, sInfo.s.Conn().RemotePeer()):
				ipAddr := rtkUtils.GetRemoteAddrFromStream(sInfo.s)
				if pingResult.Error != nil {
					log.Printf("[%s] IP[%s] Ping err:%+v", rtkMisc.GetFuncInfo(), ipAddr, pingResult.Error)
					pingFailFunc(key, sInfo)
				} else {
					if sInfo.pingErrCnt > 0 {
						log.Printf("[%s] ID:[%s] IP:[%s]  RTT [%d]ms", rtkMisc.GetFuncInfo(), sInfo.s.Conn().RemotePeer().String(), ipAddr, pingResult.RTT.Milliseconds())
						updateStreamPingErrCntInternal(key, 0)
					}
				}
			case <-pingCtx.Done():
				pingFailFunc(key, sInfo)
			}
		}, key, sInfo)
	}
}

func UpdateStream(id string, stream network.Stream) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)
	if oldSinfo, ok := streamPoolMap[id]; ok {
		log.Printf("[%s] UpdateStream id:%s Stream existed, ip[%s] old streamID:[%s] ", rtkMisc.GetFuncInfo(), id, ipAddr, oldSinfo.s.ID())
	}

	streamPoolMap[id] = streamInfo{
		s:          stream,
		timeStamp:  time.Now().UnixMilli(),
		pingErrCnt: 0,
		sFileDrop:  nil,
		sImage:     nil,
	}

	log.Printf("UpdateStream ID:[%s] IP:[%s] streamID:[%s]", id, ipAddr, stream.ID())
}

func GetStreamInfo(id string) (streamInfo, bool) {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		return sInfo, ok
	}

	return streamInfo{}, false
}

func GetStream(id string) (network.Stream, bool) {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		return sInfo.s, ok
	}

	return nil, false
}

func CheckStreamReset(id string, oldStamp int64) bool {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	if vInfo, ok := streamPoolMap[id]; ok {
		if oldStamp != vInfo.timeStamp {
			return true
		}
	}
	return false
}

func UpdateFmtTypeStream(stream network.Stream, fmtType rtkCommon.TransFmtType) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()
	id := stream.Conn().RemotePeer().String()
	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)
	if sInfo, ok := streamPoolMap[id]; ok {
		if fmtType == rtkCommon.IMAGE_CB {
			if sInfo.sImage != nil {
				log.Printf("[%s] ID:[%s] IP:[%s]  found old image stream is alive, close it first", rtkMisc.GetFuncInfo(), id, ipAddr)
				sInfo.sImage.Close()
			}
			sInfo.sImage = stream
		} else if fmtType == rtkCommon.FILE_DROP {
			if sInfo.sFileDrop != nil {
				log.Printf("[%s] ID:[%s] IP:[%s]  found old file Drop stream is alive, close it first", rtkMisc.GetFuncInfo(), id, ipAddr)
				sInfo.sFileDrop.Close()
			}
			sInfo.sFileDrop = stream
		} else {
			log.Printf("[%s] ID:[%s] IP:[%s] Unknown fmtType:[%s], update fmtType stream error!", rtkMisc.GetFuncInfo(), id, ipAddr, fmtType)
			stream.Close()
			return
		}
		streamPoolMap[id] = sInfo
		log.Printf("[%s] ID:[%s] IP:[%s] update %s stream success! ID:[%s]", rtkMisc.GetFuncInfo(), id, ipAddr, fmtType, stream.ID())
	} else {
		log.Printf("[%s] cannot found stream Info from streamPoolMap by ID:[%s]", rtkMisc.GetFuncInfo(), id)
	}
}

func GetFmtTypeStream(id string, fmtType rtkCommon.TransFmtType) (network.Stream, bool) {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		ipAddr := rtkUtils.GetRemoteAddrFromStream(sInfo.s)
		if fmtType == rtkCommon.IMAGE_CB {
			if sInfo.sImage != nil {
				return sInfo.sImage, true
			}
		} else if fmtType == rtkCommon.FILE_DROP {
			if sInfo.sFileDrop != nil {
				return sInfo.sFileDrop, true
			}
		} else {
			log.Printf("[%s] ID:[%s] IP:[%s] Unknown fmtType:[%s], get fmtType stream error!", rtkMisc.GetFuncInfo(), id, ipAddr, fmtType)
		}

	}
	return nil, false
}

func noticeFmtTypeStreamReady(id string, fmtType rtkCommon.TransFmtType) {
	key := id + string(fmtType)
	noticeChan, _ := noticeFmtTypeSteamReadyChanMap.LoadOrStore(key, make(chan struct{}, 1))
	noticeChan.(chan struct{}) <- struct{}{}
}

func HandleFmtTypeStreamReady(id string, fmtType rtkCommon.TransFmtType) {
	key := id + string(fmtType)
	noticeChan, _ := noticeFmtTypeSteamReadyChanMap.LoadOrStore(key, make(chan struct{}, 1))
	<-noticeChan.(chan struct{})
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
		log.Printf("[%s] Stream already existed, close first. id:%s", rtkMisc.GetFuncInfo(), id)
	}
	streamPoolMap[id] = streamInfo{
		s: pStream,
	}
}

func CloseStream(id string) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		if sInfo.sImage != nil {
			sInfo.sImage.Close()
			sInfo.sImage = nil
		}
		if sInfo.sFileDrop != nil {
			sInfo.sFileDrop.Close()
			sInfo.sFileDrop = nil
		}
		delete(streamPoolMap, id)
		log.Printf("ID:[%s] CloseStream streamID:[%s]", id, sInfo.s.ID())
		sInfo.s.Close()
	} else {
		log.Printf("[%s] Err: Unknown stream of id:%s", rtkMisc.GetFuncInfo(), id)
	}
}

func CloseFmtTypeStream(id string, fmtType rtkCommon.TransFmtType) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		ipAddr := rtkUtils.GetRemoteAddrFromStream(sInfo.s)
		if fmtType == rtkCommon.FILE_DROP {
			if sInfo.sFileDrop != nil {
				sInfo.sFileDrop.Close()
				sInfo.sFileDrop = nil
				streamPoolMap[id] = sInfo
				log.Printf("[%s] ID:[%s] IP:[%s] fmtType:[%s] CloseFmtTypeStream success!", rtkMisc.GetFuncInfo(), id, ipAddr, fmtType)
			}
		} else if fmtType == rtkCommon.IMAGE_CB {
			if sInfo.sImage != nil {
				sInfo.sImage.Close()
				sInfo.sImage = nil
				streamPoolMap[id] = sInfo
				log.Printf("[%s] ID:[%s] IP:[%s] fmtType:[%s] CloseFmtTypeStream success!", rtkMisc.GetFuncInfo(), id, ipAddr, fmtType)
			}
		} else {
			log.Printf("[%s] ID:[%s] IP:[%s] Unknown fmtType:[%s], close fmtType stream error!", rtkMisc.GetFuncInfo(), id, ipAddr, fmtType)
			return
		}
	} else {
		log.Printf("[%s] Err: Unknown stream info of id:%s", rtkMisc.GetFuncInfo(), id)
	}
}

func ClosePeer(id string) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		sInfo.s.Close()
		if sInfo.sImage != nil {
			sInfo.sImage.Close()
			sInfo.sImage = nil
		}
		if sInfo.sFileDrop != nil {
			sInfo.sFileDrop.Close()
			sInfo.sFileDrop = nil
		}
		node.Network().ClosePeer(sInfo.s.Conn().RemotePeer())
		log.Println("ClosePeer id:", id)
	} else {
		log.Printf("[%s] Err: Unknown stream of id:%s", rtkMisc.GetFuncInfo(), id)
	}
}

func OfflineStream(id string) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	if sInfo, ok := streamPoolMap[id]; ok {
		if sInfo.sImage != nil {
			sInfo.sImage.Close()
		}
		if sInfo.sFileDrop != nil {
			sInfo.sFileDrop.Close()
		}
		delete(streamPoolMap, id)
		log.Printf("OfflineStream ID:[%s] IP[%s] streamID:[%s]", id, rtkUtils.GetRemoteAddrFromStream(sInfo.s), sInfo.s.ID())
		sInfo.s.Close()
	} else {
		log.Printf("[%s] Err: Unknown stream of id:%s", rtkMisc.GetFuncInfo(), id)
	}
}

func IsInStreamPool(id string) bool {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	if _, ok := streamPoolMap[id]; ok {
		return true
	}
	return false
}

func IsStreamExisted(id string) bool {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		return sInfo.s.Conn().RemotePeer().String() != ""
	}
	return false
}

func CancelStreamPool() {
	nCount := uint8(0)
	streamPoolMutex.Lock()
	for id, sInfo := range streamPoolMap {
		sInfo.s.Close()
		if sInfo.sFileDrop != nil {
			sInfo.sFileDrop.Close()
		}
		if sInfo.sImage != nil {
			sInfo.sImage.Close()
		}
		delete(streamPoolMap, id)
		nCount++

		ipAddr := rtkUtils.GetRemoteAddrFromStream(sInfo.s)
		updateUIOnlineStatus(false, id, ipAddr, "", "", "")
	}
	streamPoolMutex.Unlock()
	log.Printf("CancelStreamPool stream count:%d", nCount)
}

func OfflineAllStreamEvent() {
	tempStreamMap := make(map[string](streamInfo))
	streamPoolMutex.RLock()
	for key, sInfo := range streamPoolMap {
		tempStreamMap[key] = sInfo
	}
	streamPoolMutex.RUnlock()

	nCount := uint8(0)
	for _, sInfo := range tempStreamMap {
		nCount++
		offlineEvent(sInfo.s)
	}
	log.Printf("OfflineAllStreamEvent stream count:%d", nCount)
}

func PrintfStreamPool() {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	for k, v := range streamPoolMap {
		log.Printf("key:[%+v] stream:[%+v]", k, v)
	}
}

func updateStreamPingErrCntInternal(id string, pingErrCnt int) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		sInfo.pingErrCnt = pingErrCnt
		streamPoolMap[id] = sInfo
		if pingErrCnt > 0 {
			log.Printf("UpdateStream id:[%s] Ping err cnt:[%d]", id, pingErrCnt)
		}
	}
}
