package mdns

import (
	"context"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"log"
	rtkConnection "rtk-cross-share/connection"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"time"
)

var node host.Host

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

// interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

// Initialize the MDNS service
func initMDNS(peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	// register with service so that we get notified about peer discovery
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)

	// An hour might be a long long period in practical applications. But this is fine for us
	ser := mdns.NewMdnsService(peerhost, rendezvous, n)
	if err := ser.Start(rtkUtils.GetNetInterfaces()); err != nil {
		panic(err)
	}
	return n.PeerChan
}

func MdnsServiceRun(ctx context.Context) {
	node = rtkConnection.GetConnectNode()

	/*	peerChan := initMDNS(node, rtkPlatform.GetHostID())*/
	// register with service so that we get notified about peer discovery
	Notifee := &discoveryNotifee{}
	Notifee.PeerChan = make(chan peer.AddrInfo, 5)

	// An hour might be a long long period in practical applications. But this is fine for us
	mdnsSer := mdns.NewMdnsService(node, rtkPlatform.GetHostID(), Notifee)
	if err := mdnsSer.Start(rtkUtils.GetNetInterfaces()); err != nil {
		panic(err)
	}
	peerChan := Notifee.PeerChan
	rtkConnection.MdnsStartTime = time.Now().UnixMilli()
	rtkUtils.GoSafe(func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("MdnsServiceRun is end by main context")
				mdnsSer.Close()
				return
			case peer, isOpen := <-peerChan:
				if !isOpen {
					time.Sleep(1 * time.Second)
					continue
				}
				log.Printf("[MDNS]Found peer:%+v use[%d]ms", peer, time.Now().UnixMilli()-rtkConnection.MdnsStartTime)
				rtkConnection.SetMDNSPeer(peer)
			case networkStatus := <-rtkConnection.GetNetworkStatus(): // Auto reconnection
				if !networkStatus {
					log.Println("[MDNS] Network disconnected . Stop MDNS services")
					mdnsSer.Close()
					continue
				} else {
					log.Println("[MDNS] Network reconnected. Restarting MDNS Services")
					reStartServer := false
					reTryServerTimes := 0
					for !reStartServer {
						select {
						case <-ctx.Done():
							log.Printf("MdnsServiceRun is end by main context")
							return
						case <-time.After(5 * time.Second):
							for _, p := range node.Peerstore().Peers() {
								node.Peerstore().ClearAddrs(p)
							}

							if len(node.Addrs()) == 0 || len(node.Network().ListenAddresses()) == 0 {
								log.Printf("Addr is null, retry local keep addr. node Addr:[%+v] listsen:[%+v] local:[%+v]", node.Addrs(), node.Network().ListenAddresses(), rtkConnection.ListenMultAddr())
								node.Network().Listen(rtkConnection.ListenMultAddr()...)
							} else {
								node.Network().Listen(node.Addrs()...)
							}

							newSer := mdns.NewMdnsService(node, rtkPlatform.GetHostID(), Notifee)
							if err := newSer.Start(rtkUtils.GetNetInterfaces()); err != nil {
								newSer.Close()
								reTryServerTimes++
								log.Printf("[MDNS] MDNS services restart err:[%+v], %d times retry after 5s...", err, reTryServerTimes)
							} else {
								reStartServer = true
								rtkConnection.MdnsStartTime = time.Now().UnixMilli()
								log.Printf("[MDNS] MDNS services restart successed!")
								mdnsSer = newSer
								peerChan = Notifee.PeerChan
							}

						}
					}

				}

			}

		}
	})

}
