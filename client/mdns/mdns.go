package mdns

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"log"
	rtkCommon "rtk-cross-share/common"
	rtkConnection "rtk-cross-share/connection"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"time"
)

var reConnectPeer chan peer.AddrInfo

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

// interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

// Initialize the MDNS service
func initMDNS(peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)

	ser := mdns.NewMdnsService(peerhost, rendezvous, n)
	if err := ser.Start(rtkUtils.GetNetInterfaces()); err != nil {
		panic(err)
	}
	return n.PeerChan
}

func setCachePeerInfo(cachePeer peer.AddrInfo) {
	if rtkUtils.IsInMdnsClientList(cachePeer.ID.String()) {
		return
	}

	go func(peerId string) {
		timeOut := time.NewTimer(30 * time.Second)
		defer timeOut.Stop()

		waitCoonInfo := rtkCommon.WaitToConnectInfo{
			PeerInfo:    cachePeer,
			ChExitTimer: make(chan struct{}),
		}
		rtkGlobal.WaitConnPeerMap[peerId] = waitCoonInfo
		log.Printf("setCachePeerInfo: id:%s len:%d", peerId, len(rtkGlobal.WaitConnPeerMap))

		select {
		case <-timeOut.C:
			reConnectPeer <- waitCoonInfo.PeerInfo
			delete(rtkGlobal.WaitConnPeerMap, peerId)
			return
		case <-waitCoonInfo.ChExitTimer:
			return
		}

	}(cachePeer.ID.String())
}

func BuildMdnsListener(node host.Host) {
	node.SetStreamHandler(rtkGlobal.ProtocolDirectID, func(s network.Stream) {
		rtkConnection.MdnsHandleStream(s)
	})
}

func BuildMdnsTalker(ctx context.Context, node host.Host) {
	peerChan := initMDNS(node, rtkPlatform.GetHostID())

	rtkUtils.GoSafe(func() {
		for {
			var peer peer.AddrInfo
			select {
			case peer = <-peerChan:
				if peer.ID.String() > node.ID().String() {
					log.Println("Found peer:", peer, " id is greater than us, wait for it to connect to us")
					setCachePeerInfo(peer)
					continue
				}
				log.Println("Found peer:", peer, ", connecting...")
			case peer = <-reConnectPeer:
				log.Println("peer:", peer, " is greater than us and it not connect us, it time out, so connecting...")
			}

			if rtkUtils.IsInMdnsClientList(peer.ID.String()) {
				continue
			}

			if err := node.Connect(ctx, peer); err != nil {
				fmt.Println("Connection failed:", err)
				continue
			}

			stream, err := node.NewStream(ctx, peer.ID, protocol.ID(rtkGlobal.ProtocolDirectID))
			if err != nil {
				fmt.Println("Stream open failed", err)
				continue
			}

			rtkConnection.ExecuteDirectConnect(ctx, stream)
		}
	})
}
