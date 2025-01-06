package connection

import (
	"log"
	rtkUtils "rtk-cross-share/utils"
	"sync"

	"github.com/libp2p/go-libp2p/core/network"
)

var (
	streamPoolMap	= make(map[string](network.Stream))
	streamPoolMutex	sync.RWMutex
)

func UpdateStream(id string, stream network.Stream) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	streamPoolMap[id] = stream
	log.Println("UpdateStream id:", id)
}

func GetStream(id string) (network.Stream, bool) {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()

	stream, ok := streamPoolMap[id]
	return stream, ok
}

func AddStream(id string, pStream network.Stream) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	if stream, ok := streamPoolMap[id]; ok {
		stream.Close()
		log.Printf("[%s %d] Stream already existed, close first. id:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), id)
	}
	streamPoolMap[id] = pStream
}

func CloseStream(id string) {
	streamPoolMutex.Lock()
	defer streamPoolMutex.Unlock()

	if stream, ok := streamPoolMap[id]; ok {
		stream.Close()
		delete(streamPoolMap, id)
		log.Println("ClsoeStream id:", id)
	} else {
		log.Printf("[%s %d] Err: Unknown stream of id:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), id)
	}
}

func IsStreamExisted(id string) bool {
	streamPoolMutex.RLock()
	defer streamPoolMutex.RUnlock()

	if stream, ok := streamPoolMap[id]; ok {
		return stream.Conn().RemotePeer().String() != ""
	} else {
		return false
	}
}
