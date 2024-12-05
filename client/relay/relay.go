package relay

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	ma "github.com/multiformats/go-multiaddr"

	rtkCommon "rtk-cross-share/common"
	rtkConnection "rtk-cross-share/connection"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
)

func regiterToServer(ctx context.Context, s network.Stream, node host.Host) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	regMsg := rtkCommon.RegMessage{
		HOST:  rtkPlatform.GetHostID(),
		GUEST: node.ID().String(),
	}
	if err := json.NewEncoder(rw).Encode(regMsg); err != nil {
		log.Println("failed to send register message: %w", err)
		return
	}
	if err := rw.Flush(); err != nil {
		log.Println("Error flushing write buffer: %w", err)
		return
	}

	var regResonseMsg rtkCommon.RegResponseMessage
	if err := json.NewDecoder(rw).Decode(&regResonseMsg); err != nil {
		log.Println("failed to read register response message: %w", err)
		return
	}

	//fmt.Println(regResonseMsg)
	rtkGlobal.GuestList = rtkUtils.RemoveMySelfID(regResonseMsg.GUEST_LIST, node.ID().String())
	rtkGlobal.NodeInfo.IPAddr.PublicIP = regResonseMsg.GUEST_PUBLIC_TCP_IP
	rtkGlobal.NodeInfo.IPAddr.PublicPort = regResonseMsg.GUEST_PUBLIC_TCP_PORT
}

func setupRelayServerConnection(ctx context.Context, node host.Host, relayServerAddr string) {
	msg, _ := ma.NewMultiaddr(relayServerAddr)
	relayinfo, err := peer.AddrInfoFromP2pAddr(msg)
	if err != nil {
		log.Println(err.Error())
		if rtkPlatform.CallbackInstance != nil {
			rtkPlatform.CallbackInstance.EventCallback(rtkCommon.P2P_EVENT_SERVER_CONNECT_FAIL)
		}
		return
	}

	if err := node.Connect(ctx, *relayinfo); err != nil {
		log.Printf("Failed to connect node and relay: %v", err)
		if rtkPlatform.CallbackInstance != nil {
			rtkPlatform.CallbackInstance.EventCallback(rtkCommon.P2P_EVENT_SERVER_CONNECT_FAIL)
		}
		return
	}

	// Open a stream to the relay server
	stream, err := node.NewStream(network.WithAllowLimitedConn(ctx, "register"), relayinfo.ID, rtkGlobal.HostProtocolID)
	if err != nil {
		log.Printf("Unexpected error here. Failed to open stream: %v", err)
		if rtkPlatform.CallbackInstance != nil {
			rtkPlatform.CallbackInstance.EventCallback(rtkCommon.P2P_EVENT_SERVER_CONNECT_FAIL)
		}
		return
	}

	defer stream.Close()
	regiterToServer(ctx, stream, node)
	reservation, err := client.Reserve(ctx, node, *relayinfo)
	if err != nil {
		log.Printf("unreachable failed to receive a relay reservation from relay. %v", err)
		if rtkPlatform.CallbackInstance != nil {
			rtkPlatform.CallbackInstance.EventCallback(rtkCommon.P2P_EVENT_SERVER_CONNECT_FAIL)
		}
		return
	}
	fmt.Println("Dump reservation")
	fmt.Println(reservation)

	if rtkPlatform.CallbackInstance != nil {
		rtkPlatform.CallbackInstance.EventCallback(rtkCommon.P2P_EVENT_SERVER_CONNEDTED)
	}
}

func BuildListener(ctx context.Context, node host.Host) {
	relayServerAddr := rtkGlobal.RelayServerIPInfo + rtkGlobal.RelayServerID
	setupRelayServerConnection(ctx, node, relayServerAddr)
	node.SetStreamHandler(rtkGlobal.ProtocolID, func(s network.Stream) {
		rtkUtils.GoSafe(func() {rtkConnection.HandleStream(s)})
	})
}

func BuildTalker(ctx context.Context, node host.Host, targetId string) {
	relayServerAddr := rtkGlobal.RelayServerIPInfo + rtkGlobal.RelayServerID
	targetAddr := relayServerAddr + "/p2p-circuit/p2p/" + targetId
	targetAddrWithRelay, err := ma.NewMultiaddr(targetAddr)
	if err != nil {
		log.Println(err)
		return
	}

	// FIXME:
	// Should check connect directly first
	// If we cannot communicate directly and then connect with relay
	useRelayToDestInfo, _ := peer.AddrInfoFromP2pAddr(targetAddrWithRelay)
	if err := node.Connect(ctx, *useRelayToDestInfo); err != nil {
		log.Printf("Unexpected error here. Failed to connect remote node by relay: %v", err)
		// Skip this connection due to invalid ID
		return
	}

	// Open a stream to the peer
	stream, err := node.NewStream(network.WithAllowLimitedConn(ctx, "chatStream"), useRelayToDestInfo.ID, rtkGlobal.ProtocolID)
	if err != nil {
		log.Printf("Unexpected error here. Failed to open stream: %v", err)
		return
	}

	rtkUtils.GoSafe(func() {rtkConnection.ExecuteP2PConnect(ctx, stream, node)})
}
