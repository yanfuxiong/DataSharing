package mdns

import (
	"context"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"log"
	rtkConnection "rtk-cross-share/connection"
	rtkPeer2Peer "rtk-cross-share/peer2peer"
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

	rtkUtils.GoSafe(func() {
		checkNetWorkConnectCount := 0
		bNetWorkFlag := true
		reStartServer := false
		reTryServerTimes := 0
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
				log.Printf("[MDNS]Found peer:%+v", peer)
				rtkConnection.SetMDNSPeer(peer)
			case <-time.After(1 * time.Second): // Auto reconnection
				if !rtkConnection.IsNetworkConnected() {
					if bNetWorkFlag {
						bNetWorkFlag = false
						log.Println("[MDNS] Network disconnected . Stop MDNS services")
						//close(peerChan)
						mdnsSer.Close()
						rtkPeer2Peer.CaneclProcessForPeerMap()
						rtkConnection.CancelStreamPool()
					}
					checkNetWorkConnectCount++
					if checkNetWorkConnectCount > 5 {
						checkNetWorkConnectCount = 0
						log.Println("[MDNS] Network is unavailable!")
					}
					continue
				} else {
					if !bNetWorkFlag || reStartServer {
						if !bNetWorkFlag {
							log.Println("[MDNS] Network reconnected. Restarting MDNS Services")
						}
						bNetWorkFlag = true

						if reStartServer {
							log.Printf("[MDNS] is %d times to retry restart MDNS services", reTryServerTimes)
						}

						for _, p := range node.Peerstore().Peers() {
							node.Peerstore().ClearAddrs(p)
						}

						if len(node.Addrs()) == 0 || len(node.Network().ListenAddresses()) == 0 {
							log.Printf("Addr is null, retry local keep addr. node Addr:[%+v] listsen:[%+v] local:[%+v]", node.Addrs(), node.Network().ListenAddresses(), rtkConnection.ListenMultAddr())
							node.Network().Listen(rtkConnection.ListenMultAddr()...)
						} else {
							node.Network().Listen(node.Addrs()...)
						}

						/*n := &discoveryNotifee{}
						n.PeerChan = make(chan peer.AddrInfo, 5)*/

						// An hour might be a long long period in practical applications. But this is fine for us
						newSer := mdns.NewMdnsService(node, rtkPlatform.GetHostID(), Notifee)
						if err := newSer.Start(rtkUtils.GetNetInterfaces()); err != nil {
							newSer.Close()
							//panic(err)
							log.Printf("[MDNS] MDNS services restart err:[%+v], retry after 3s...", err)
							time.Sleep(3 * time.Second)
							reTryServerTimes++
							reStartServer = true
						} else {
							reTryServerTimes = 0
							reStartServer = false
							log.Printf("[MDNS] MDNS services restart successed!")
						}
						mdnsSer = newSer
						peerChan = Notifee.PeerChan
					}
				}

			}

		}
	})

}
