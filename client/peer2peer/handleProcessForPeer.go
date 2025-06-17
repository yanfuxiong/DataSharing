package peer2peer

import (
	"context"
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkConnection "rtk-cross-share/client/connection"
	rtkGlobal "rtk-cross-share/client/global"
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"time"
)

var (
	processForPeerMap = make(map[string]func())
	processMutex      sync.Mutex
)

func StartProcessForPeer(id string) {
	processMutex.Lock()
	defer processMutex.Unlock()

	ipAddr := rtkConnection.GetStreamIpAddr(id)
	if _, ok := processForPeerMap[id]; !ok {
		ctx, canecl := context.WithCancel(context.Background())
		rtkMisc.GoSafe(func() { ProcessEventsForPeer(id, ipAddr, ctx) })
		processForPeerMap[id] = canecl
		log.Printf("[%s][%s][%s] ProcessEventsForPeer is Start !", rtkMisc.GetFuncInfo(), id, ipAddr)
	} else {
		log.Printf("[%s][%s][%s] ProcessEventsForPeer is existed, not need restart!", rtkMisc.GetFuncInfo(), id, ipAddr)
	}
}

func EndProcessForPeer(id string) {
	processMutex.Lock()
	defer processMutex.Unlock()
	if _, ok := processForPeerMap[id]; ok {
		processForPeerMap[id]()
		delete(processForPeerMap, id)
		log.Printf("[%s][%s][%s] ProcessEventsForPeer is Cancel !", rtkMisc.GetFuncInfo(), id, rtkConnection.GetStreamIpAddr(id))
	} else {
		log.Printf("[%s][%s][%s] ProcessEventsForPeer is already Cancel, not need cancel again !", rtkMisc.GetFuncInfo(), id, rtkConnection.GetStreamIpAddr(id))
	}
}

func CaneclProcessForPeerMap() {
	processMutex.Lock()
	defer processMutex.Unlock()
	for id, _ := range processForPeerMap {
		processForPeerMap[id]()
		delete(processForPeerMap, id)
	}
	log.Printf("Canecl all ProcessForPeer")
}

func GetProcessForPeerCount() int {
	processMutex.Lock()
	defer processMutex.Unlock()
	return len(processForPeerMap)
}

func SendDisconnectToAllPeer(isFromExtractDIAS bool) {
	nSendMsgCount := uint32(0)
	nCancelProcessForPeerCount := uint32(0)
	processMutex.Lock()
	defer processMutex.Unlock()
	for id, _ := range processForPeerMap {
		SendCmdMsgToPeer(id, COMM_DISCONNECT, rtkCommon.TEXT_CB, rtkMisc.SUCCESS)
		nSendMsgCount++
		if isFromExtractDIAS {
			processForPeerMap[id]()
			delete(processForPeerMap, id)
			nCancelProcessForPeerCount++
		}
	}
	log.Printf("Send %d disconnect message and Canecl %d ProcessForPeer", nSendMsgCount, nCancelProcessForPeerCount)
}

func SendCmdMsgToPeer(id string, cmd CommandType, fmtType rtkCommon.TransFmtType, errCode rtkMisc.CrossShareErr) {
	var msg Peer2PeerMessage
	msg.SourceID = rtkGlobal.NodeInfo.ID
	msg.SourcePlatform = rtkGlobal.NodeInfo.Platform
	msg.FmtType = fmtType
	msg.TimeStamp = uint64(time.Now().UnixMilli())
	msg.Command = cmd
	msg.ExtData = errCode
	writeToSocket(&msg, id)
}
