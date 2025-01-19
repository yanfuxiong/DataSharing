package peer2peer

import (
	"context"
	"log"
	rtkConnection "rtk-cross-share/connection"
	rtkUtils "rtk-cross-share/utils"
	"sync"
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
		rtkUtils.GoSafe(func() { ProcessEventsForPeer(id, ctx) })
		processForPeerMap[id] = canecl
		log.Printf("[%s][%s][%s] ProcessEventsForPeer is Start !", rtkUtils.GetFuncInfo(), id, ipAddr)
	} else {
		log.Printf("[%s][%s][%s] ProcessEventsForPeer is existed, not need restart!", rtkUtils.GetFuncInfo(), id, ipAddr)
	}
}

func EndProcessForPeer(id string) {
	processMutex.Lock()
	defer processMutex.Unlock()
	processForPeerMap[id]()
	delete(processForPeerMap, id)
	log.Printf("[%s][%s][%s] ProcessEventsForPeer is Cancel !", rtkUtils.GetFuncInfo(), id, rtkConnection.GetStreamIpAddr(id))
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
