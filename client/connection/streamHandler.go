package connection

import (
	"context"
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
)

type TransFileStateType string

const (
	TRANS_FILE_IN_PROGRESS_SRC TransFileStateType = "TransFile_In_Progress_SRC"
	TRANS_FILE_IN_PROGRESS_DST TransFileStateType = "TransFile_In_Progress_DST"
	TRANS_FILE_NOT_PREFORMED   TransFileStateType = "TransFile_Not_Performed"
)

type streamInfo struct {
	s              network.Stream
	ipAddr         string
	timeStamp      int64
	pingErrCnt     int
	sFileDrop      network.Stream
	sImage         network.Stream
	transFileState TransFileStateType

	cancelFn func()
}

var (
	streamPoolMap   = make(map[string](streamInfo))
	streamPoolMutex sync.RWMutex
)

func init() {
	rtkPlatform.SetGetFilesTransCodeCallback(GetFileTransErrCode)
}

func CheckAllStreamAlive(ctx context.Context) {
	pingFailFunc := func(key string, sInfo streamInfo) {
		if CheckStreamReset(key, sInfo.timeStamp) { // if stream is update, not need go through this flow
			return
		}

		pingErrCnt := updateStreamPingErrCntIncrease(key)
		if (pingErrCnt >= pingErrMaxCnt) && (sInfo.transFileState == TRANS_FILE_NOT_PREFORMED) {
			offlineEvent(sInfo.s)
		}
	}

	tempStreamMap := make(map[string](streamInfo))
	streamPoolMutex.RLock()
	for key, sInfo := range streamPoolMap {
		tempStreamMap[key] = sInfo
	}
	streamPoolMutex.RUnlock()

	for key, sInfo := range tempStreamMap {
		if sInfo.transFileState == TRANS_FILE_IN_PROGRESS_SRC || sInfo.transFileState == TRANS_FILE_IN_PROGRESS_DST {
			continue
		}

		rtkMisc.GoSafeWithParam(func(args ...any) {
			// Default timeout is 10 sec in Ping.go
			// Use this context timeout instead of the timeout in Ping.go
			pingCtx, cancelFun := context.WithTimeout(ctx, pingTimeout)
			defer cancelFun()
			select {
			case pingResult := <-pingServer.Ping(pingCtx, sInfo.s.Conn().RemotePeer()):
				if pingResult.Error != nil {
					log.Printf("[%s] IP[%s] Ping err:%+v", rtkMisc.GetFuncInfo(), sInfo.ipAddr, pingResult.Error)
					pingFailFunc(key, sInfo)
				} else {
					if sInfo.pingErrCnt > 0 {
						log.Printf("[%s] ID:[%s] IP:[%s]  RTT [%d]ms", rtkMisc.GetFuncInfo(), sInfo.s.Conn().RemotePeer().String(), sInfo.ipAddr, pingResult.RTT.Milliseconds())
						updateStreamPingErrCntReset(key)
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
		log.Printf("[%s] UpdateStream ID:%s  IP:[%s],Stream existed, the old streamID:[%s] ", rtkMisc.GetFuncInfo(), id, ipAddr, oldSinfo.s.ID())
		if oldSinfo.cancelFn != nil {
			oldSinfo.cancelFn()
			log.Printf("[%s] UpdateStream ID:[%s] IP:[%s], ProcessForPeer existed, Cancel the old StartProcessForPeer first!", rtkMisc.GetFuncInfo(), id, ipAddr)
		}
	}

	streamPoolMap[id] = streamInfo{
		s:              stream,
		ipAddr:         ipAddr,
		timeStamp:      time.Now().UnixMilli(),
		pingErrCnt:     0,
		sFileDrop:      nil,
		sImage:         nil,
		transFileState: TRANS_FILE_NOT_PREFORMED,
		cancelFn:       callbackStartProcessForPeer(id, ipAddr), // StartProcessForPeer
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

func updateFmtTypeStreamInternal(stream network.Stream, fmtType rtkCommon.TransFmtType, isDst bool) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()
	id := stream.Conn().RemotePeer().String()
	if sInfo, ok := streamPoolMap[id]; ok {
		if fmtType == rtkCommon.IMAGE_CB {
			if sInfo.sImage != nil {
				log.Printf("[%s] ID:[%s] IP:[%s]  found old image stream is alive, stop it !", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr)
				if isDst {
					sInfo.sImage.Reset() // cover the last image
				}
			}
			sInfo.sImage = stream
		} else if fmtType == rtkCommon.FILE_DROP {
			if sInfo.sFileDrop != nil {
				log.Printf("[%s] ID:[%s] IP:[%s]  found old file Drop stream is alive, not close it !", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr)
				//sInfo.sFileDrop.Close()
			}
			sInfo.sFileDrop = stream
			if isDst {
				sInfo.transFileState = TRANS_FILE_IN_PROGRESS_DST
			} else {
				sInfo.transFileState = TRANS_FILE_IN_PROGRESS_SRC
			}
		} else {
			log.Printf("[%s] ID:[%s] IP:[%s] Unknown fmtType:[%s], update fmtType stream error!", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr, fmtType)
			stream.Close()
			return
		}
		streamPoolMap[id] = sInfo
		log.Printf("[%s] ID:[%s] IP:[%s] update %s stream success! ID:[%s]", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr, fmtType, stream.ID())
	} else {
		log.Printf("[%s] cannot found stream Info from streamPoolMap by ID:[%s]", rtkMisc.GetFuncInfo(), id)
	}
}
func updateFmtTypeStreamSrc(stream network.Stream, fmtType rtkCommon.TransFmtType) {
	updateFmtTypeStreamInternal(stream, fmtType, false)
}

func updateFmtTypeStreamDst(stream network.Stream, fmtType rtkCommon.TransFmtType) {
	updateFmtTypeStreamInternal(stream, fmtType, true)
}

func GetFmtTypeStream(id string, fmtType rtkCommon.TransFmtType) (network.Stream, bool) {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		if fmtType == rtkCommon.IMAGE_CB {
			if sInfo.sImage != nil {
				return sInfo.sImage, true
			}
		} else if fmtType == rtkCommon.FILE_DROP {
			if sInfo.sFileDrop != nil {
				return sInfo.sFileDrop, true
			}
		} else {
			log.Printf("[%s] ID:[%s] IP:[%s] Unknown fmtType:[%s], get fmtType stream error!", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr, fmtType)
		}

	}
	return nil, false
}

func GetFileTransErrCode(id string) rtkCommon.SendFilesRequestErrCode {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		if sInfo.transFileState == TRANS_FILE_IN_PROGRESS_SRC {
			log.Printf("[%s] ID:[%s] Currently sending documents to this user", rtkMisc.GetFuncInfo(), id)
			return rtkCommon.SendFilesRequestInProgressBySrc
		} else if sInfo.transFileState == TRANS_FILE_IN_PROGRESS_DST {
			log.Printf("[%s] ID:[%s] Currently receiving documents from this user", rtkMisc.GetFuncInfo(), id)
			return rtkCommon.SendFilesRequestInProgressByDst
		}
	}

	return rtkCommon.SendFilesRequestSuccess
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
		if sInfo.cancelFn != nil { // StopProcessForPeer
			log.Printf("ID:[%s] IP:[%s] ProcessEventsForPeer is Cancel! ", id, sInfo.ipAddr)
			sInfo.cancelFn()
			sInfo.cancelFn = nil
		}
		delete(streamPoolMap, id)
		log.Printf("ID:[%s] IP:[%s] CloseStream,  StreamID:[%s]", id, sInfo.ipAddr, sInfo.s.ID())
		sInfo.s.Close()
	} else {
		log.Printf("[%s] Err: Unknown stream of id:%s", rtkMisc.GetFuncInfo(), id)
	}
}

func CloseFmtTypeStream(id string, fmtType rtkCommon.TransFmtType) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		if fmtType == rtkCommon.FILE_DROP {
			if sInfo.sFileDrop != nil {
				sInfo.sFileDrop.Close()
				sInfo.sFileDrop = nil
				sInfo.transFileState = TRANS_FILE_NOT_PREFORMED
				streamPoolMap[id] = sInfo
				log.Printf("[%s] ID:[%s] IP:[%s] fmtType:[%s] CloseFmtTypeStream success!", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr, fmtType)
			}
		} else if fmtType == rtkCommon.IMAGE_CB {
			if sInfo.sImage != nil {
				sInfo.sImage.Close()
				sInfo.sImage = nil
				streamPoolMap[id] = sInfo
				log.Printf("[%s] ID:[%s] IP:[%s] fmtType:[%s] CloseFmtTypeStream success!", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr, fmtType)
			}
		} else {
			log.Printf("[%s] ID:[%s] IP:[%s] Unknown fmtType:[%s], close fmtType stream error!", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr, fmtType)
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
		callbackSendDisconnectMsgToPeer(id)
		if sInfo.sImage != nil {
			sInfo.sImage.Close()
			sInfo.sImage = nil
		}
		if sInfo.sFileDrop != nil {
			sInfo.sFileDrop.Close()
			sInfo.sFileDrop = nil
			sInfo.transFileState = TRANS_FILE_NOT_PREFORMED
		}

		if sInfo.cancelFn != nil { // StopProcessForPeer
			log.Printf("ID:[%s] IP:[%s] ProcessEventsForPeer is Cancel! ", id, sInfo.ipAddr)
			sInfo.cancelFn()
		}
		sInfo.s.Close()
		node.Network().ClosePeer(sInfo.s.Conn().RemotePeer())
		delete(streamPoolMap, id)
		log.Println("ClosePeer id:", id)
	} else {
		log.Printf("[%s] Err: Unknown stream of id:%s", rtkMisc.GetFuncInfo(), id)
	}
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
	tempStreamMap := make(map[string](streamInfo))
	streamPoolMutex.RLock()
	for key, sInfo := range streamPoolMap {
		tempStreamMap[key] = sInfo
	}
	streamPoolMutex.RUnlock()

	nCount := uint8(0)
	for id, sInfo := range tempStreamMap {
		updateUIOnlineStatus(false, id, sInfo.ipAddr, "", "", "", "")
		callbackSendDisconnectMsgToPeer(id)

		if sInfo.sFileDrop != nil {
			sInfo.sFileDrop.Close()
		}
		if sInfo.sImage != nil {
			sInfo.sImage.Close()
		}
		if sInfo.cancelFn != nil { // StopProcessForPeer
			log.Printf("[%s] ID:[%s] IP:[%s] ProcessEventsForPeer is Cancel! ", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr)
			sInfo.cancelFn()
			sInfo.cancelFn = nil
		}
		sInfo.s.Close()

		streamPoolMutex.Lock()
		delete(streamPoolMap, id)
		streamPoolMutex.Unlock()
		nCount++
	}

	log.Printf("[%s] Cancel stream count:%d", rtkMisc.GetFuncInfo(), nCount)
}

func PrintfStreamPool() {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()
	for k, v := range streamPoolMap {
		log.Printf("key:[%+v] stream:[%+v]", k, v)
	}
}

func updateStreamPingErrCntIncrease(id string) int {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		sInfo.pingErrCnt = sInfo.pingErrCnt + 1
		streamPoolMap[id] = sInfo
		log.Printf("Update PingErrCnt id:[%s] Ping err cnt:[%d]", id, sInfo.pingErrCnt)
		return sInfo.pingErrCnt
	}
	return 0
}

func updateStreamPingErrCntReset(id string) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()
	if sInfo, ok := streamPoolMap[id]; ok {
		sInfo.pingErrCnt = 0
		streamPoolMap[id] = sInfo
		log.Printf("Update PingErrCnt id:[%s] Ping err cnt reset!", id)
	}
}
