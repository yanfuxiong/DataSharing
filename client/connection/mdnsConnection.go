package connection

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	rtkPeer2Peer "rtk-cross-share/peer2peer"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"

	"github.com/libp2p/go-libp2p/core/network"
)

func MdnsHandleStream(stream network.Stream) {
	connP2P := rtkUtils.NewConnFromStream(stream)
	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)

	if stream.Conn().RemotePeer().String() > rtkGlobal.NodeInfo.ID {
		rtkUtils.ClearCachePeerInfo(stream.Conn().RemotePeer().String())
	}

	handleRegister(stream)

	ip, port := rtkUtils.ExtractTCPIPandPort(stream.Conn().LocalMultiaddr())
	rtkGlobal.NodeInfo.IPAddr.PublicIP = ip
	rtkGlobal.NodeInfo.IPAddr.PublicPort = port
	fmt.Printf("public ip [%s] port[%s]\n", ip, port)
	fmt.Println("************************************************")
	log.Println("H Connected to ID:", stream.Conn().RemotePeer().String(), " IP:", ipAddr)
	fmt.Println("************************************************")
	rtkPlatform.GoUpdateClientStatus(1, ipAddr, stream.Conn().RemotePeer().String(), ipAddr)

	connCtx, cancel := context.WithCancel(context.Background())
	rtkUtils.GoSafe(func() { rtkPeer2Peer.ProcessEventsForPeer(connP2P, ipAddr, connCtx, cancel) })
	rtkUtils.GoSafe(func() {
		<-connCtx.Done()
		rtkUtils.LostMdnsClientList(stream.Conn().RemotePeer().String())
		rtkPlatform.FoundPeer()
		fmt.Println("************************************************")
		log.Println("Lost connection with ID:", stream.Conn().RemotePeer().String(), " IP:", ipAddr)
		fmt.Println("************************************************")
		rtkPlatform.GoUpdateClientStatus(0, ipAddr, stream.Conn().RemotePeer().String(), ipAddr)
		stream.Close()
	})

}

func ExecuteDirectConnect(ctx context.Context, stream network.Stream) {
	connP2P := rtkUtils.NewConnFromStream(stream)
	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)

	if registerToPeer(stream) != nil {
		rtkUtils.InsertMdnsClientList(stream.Conn().RemotePeer().String(), ipAddr, "unknownPlatform")
		rtkPlatform.FoundPeer()
	}

	ip, port := rtkUtils.ExtractTCPIPandPort(stream.Conn().LocalMultiaddr())
	rtkGlobal.NodeInfo.IPAddr.PublicIP = ip
	rtkGlobal.NodeInfo.IPAddr.PublicPort = port
	fmt.Printf("public ip [%s] port[%s]\n", ip, port)
	fmt.Println("************************************************")
	log.Println("E Connected to ID:", stream.Conn().RemotePeer().String(), " IP:", ipAddr)
	fmt.Println("************************************************")
	rtkPlatform.GoUpdateClientStatus(1, ipAddr, stream.Conn().RemotePeer().String(), ipAddr)

	connCtx, cancel := context.WithCancel(context.Background())
	rtkUtils.GoSafe(func() { rtkPeer2Peer.ProcessEventsForPeer(connP2P, ipAddr, connCtx, cancel) })
	rtkUtils.GoSafe(func() {
		<-connCtx.Done()
		rtkUtils.LostMdnsClientList(stream.Conn().RemotePeer().String())
		rtkPlatform.FoundPeer()
		fmt.Println("************************************************")
		log.Println("Lost connection with ID:", stream.Conn().RemotePeer().String(), " IP:", ipAddr)
		fmt.Println("************************************************")
		rtkPlatform.GoUpdateClientStatus(0, ipAddr, stream.Conn().RemotePeer().String(), ipAddr)
		stream.Close()
	})
}

func registerToPeer(s network.Stream) error {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	ipAddr := rtkUtils.GetRemoteAddrFromStream(s)

	registMsg := rtkCommon.RegistMdnsMessage{
		Host:     rtkPlatform.GetHostID(),
		Id:       rtkGlobal.NodeInfo.ID,
		Platform: rtkPlatform.GetPlatform(),
	}
	if err := json.NewEncoder(rw).Encode(registMsg); err != nil {
		log.Println("failed to send register message: %w", err)
		return err
	}
	if err := rw.Flush(); err != nil {
		log.Println("Error flushing write buffer: %w", err)
		return err
	}
	var regResonseMsg rtkCommon.RegistMdnsMessage
	if err := json.NewDecoder(rw).Decode(&regResonseMsg); err != nil {
		log.Println("failed to read register response message: %w", err)
		return err
	}

	rtkUtils.InsertMdnsClientList(regResonseMsg.Id, ipAddr, regResonseMsg.Platform)
	rtkPlatform.FoundPeer()
	log.Println("registerToPeer success!")
	return nil
}

func handleRegister(s network.Stream) error {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	ipAddr := rtkUtils.GetRemoteAddrFromStream(s)

	var regMsg rtkCommon.RegistMdnsMessage
	err := json.NewDecoder(rw).Decode(&regMsg)
	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			fmt.Println("Stream context canceled or deadline exceeded:", err)
		}
		if err.Error() == "stream reset" {
			fmt.Println("Stream reset by peer:", err)
		}
		return err
	}
	rtkUtils.InsertMdnsClientList(regMsg.Id, ipAddr, regMsg.Platform)
	rtkPlatform.FoundPeer()

	registMsg := rtkCommon.RegistMdnsMessage{
		Host:     rtkPlatform.GetHostID(),
		Id:       rtkGlobal.NodeInfo.ID,
		Platform: rtkPlatform.GetPlatform(),
	}

	if err := json.NewEncoder(rw).Encode(&registMsg); err != nil {
		fmt.Println("failed to read register response message: ", err)
		return err
	}
	if err := rw.Flush(); err != nil {
		fmt.Println("Error flushing write buffer: ", err)
		return err
	}
	log.Println("handleRegister success!")
	return nil
}
