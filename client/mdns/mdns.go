package mdns

import (
	"context"
	"log"
	"net"
	rtkConnection "rtk-cross-share/connection"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

// interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

func isNetworkConnected() bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Failed to get network interfaces: %v", err)
		return false
	}

	for _, iface := range interfaces {
		if (iface.Flags & net.FlagUp) == 0 || (iface.Flags & net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Err: Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil || ipNet.IP.To16() != nil {
					return true
				}
			}
		}
	}

	return false
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

func BuildMdnsListener(node host.Host) {
	node.SetStreamHandler(rtkGlobal.ProtocolDirectID, func(s network.Stream) {
		if rtkConnection.IsStreamExisted(s.Conn().RemotePeer().String()) {
			log.Printf("[MDNS] H: Stream already existed, skip connect. ID: %s", s.Conn().RemotePeer().String())
		} else {
			rtkConnection.MdnsHandleStream(s)
			rtkConnection.AddStream(s.Conn().RemotePeer().String(), s)
		}
	})
}

func BuildMdnsTalker(ctx context.Context, node host.Host) {
	peerChan := initMDNS(node, rtkPlatform.GetHostID())

	/*
	// Auto reconnection
	// register with service so that we get notified about peer discovery
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)

	// An hour might be a long long period in practical applications. But this is fine for us
	ser := mdns.NewMdnsService(node, rtkPlatform.GetHostID(), n)
	if err := ser.Start(rtkUtils.GetNetInterfaces()); err != nil {
		panic(err)
	}
	peerChan := n.PeerChan

	rtkUtils.GoSafe(func() {
		for {
			if !isNetworkConnected() {
				log.Println("[MDNS] Network disconnected. Storping MDNS services")
				close(peerChan)
				ser.Close()

				for !isNetworkConnected() {
					time.Sleep(5*time.Second)
					log.Println("[MDNS] Network unavailable")
				}

				log.Println("[MDNS] Network reconnected. Restarting mDNS")
				// mdnsService, peerChan := initMDNS(node, rtkPlatform.GetHostID())
				n := &discoveryNotifee{}
				n.PeerChan = make(chan peer.AddrInfo)

				// An hour might be a long long period in practical applications. But this is fine for us
				newSer := mdns.NewMdnsService(node, rtkPlatform.GetHostID(), n)
				if err := newSer.Start(rtkUtils.GetNetInterfaces()); err != nil {
					panic(err)
				}
				ser = newSer
				peerChan = n.PeerChan
			}
			time.Sleep(5*time.Second)
		}
	})
	*/

	rtkUtils.GoSafe(func() {
		for {
			peer := <-peerChan
			if rtkConnection.IsStreamExisted(peer.ID.String()) {
				log.Printf("[MDNS] E: Stream already existed, skip connect. ID: %s", peer.ID.String())
				continue
			}

			if peer.ID > node.ID() {
				// if other end peer id greater than us, don't connect to it, just wait for it to connect us
				log.Println("[MDNS] Found peer:", peer, " id is greater than us, wait for it to connect to us")
				time.Sleep(15 * time.Second)
				if rtkConnection.IsStreamExisted(peer.ID.String()) {
					continue
				} else {
					log.Printf("[MDNS] Wait for node timeout, execute direct connect: %s", peer.ID.String())
				}
			}
			log.Println("Found peer:", peer, ", connecting...")

			if err := node.Connect(ctx, peer); err != nil {
				log.Println("Connection failed:", err)
				time.Sleep(3*time.Second)
				continue
			}
			// open a stream, this stream will be handled by handleStream other end
			stream, err := node.NewStream(ctx, peer.ID, protocol.ID(rtkGlobal.ProtocolDirectID))
			if err != nil {
				log.Println("Stream open failed", err)
				time.Sleep(3*time.Second)
				continue
			}

			rtkConnection.ExecuteDirectConnect(ctx, stream)
			rtkConnection.AddStream(stream.Conn().RemotePeer().String(), stream)
			log.Println("BuildMdnsTalker Connected to:", peer.ID.String(), " success!")
		}
	})

}
