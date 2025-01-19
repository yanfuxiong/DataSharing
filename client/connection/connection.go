package connection

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-reuseport"
)

func performSimultaneousOpen(conn network.Conn, localPort string, peerAddr string) net.Conn {
	peerID := conn.RemotePeer()
	log.Println("Performing simultaneous open to: ", peerID, " at ", peerAddr, " with ", localPort)

	var errCount = 0
	var connP2P net.Conn
	var err error
	const timeout = 6
	for errCount <= timeout {
		connP2P, err = reuseport.DialTimeout("tcp", fmt.Sprintf(":%v", localPort), peerAddr, 10*time.Second)
		if err != nil {
			log.Println("Request Fail!: %w", err)
			errCount++
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if errCount > timeout {
		log.Println("Cannot connect each node")
		if rtkPlatform.CallbackInstance != nil {
			rtkPlatform.CallbackInstance.EventCallback(rtkCommon.P2P_EVENT_CLIENT_CONNECT_FAIL)
		}
		return nil
	}

	log.Println("************************************************")
	log.Println("Connected to ID:", peerID, " IP:", peerAddr)
	log.Println("************************************************")

	if rtkPlatform.CallbackInstance != nil {
		rtkPlatform.CallbackInstance.EventCallback(rtkCommon.P2P_EVENT_CLIENT_CONNEDTED)
	}

	// convert net.Conn to *net.TCPConn
	tcpConn, ok := connP2P.(*net.TCPConn)
	if ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		//log.Println("KeepAlive is enabled, period: 30s")
	} else {
		log.Println("not tcp connection, skip setting KeepAlive")
	}
	return connP2P
}

func performDCUtRHandshake(s network.Stream) net.Conn {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	myAddr := rtkUtils.ConcatIP(rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort)
	var connP2P net.Conn = nil

	for {
		// Send Connect message
		connectMsg := rtkCommon.ConnectMessage{Tag: "CONNECT", ObservedAddrs: myAddr}
		if err := json.NewEncoder(rw).Encode(connectMsg); err != nil {
			log.Println("failed to send connect message: %w", err)
			return nil
		}
		if err := rw.Flush(); err != nil {
			log.Println("Error flushing write buffer: %w", err)
			return nil
		}

		// Start timer to measure RTT
		start := time.Now()
		log.Println(start)

		// Read Connect message
		var responseMsg rtkCommon.ConnectMessage
		if err := json.NewDecoder(rw).Decode(&responseMsg); err != nil {
			log.Println("failed to read connect message: %w", err)
			return nil
		}

		log.Println("Received Connect message with observed addresses:", connectMsg.ObservedAddrs)

		// Calculate RTT
		rtt := time.Since(start)
		log.Println(rtt.Milliseconds())

		// Send Sync message
		syncMsg := rtkCommon.SyncMessage{Tag: "SYNC"}
		if err := json.NewEncoder(rw).Encode(syncMsg); err != nil {
			log.Println("failed to send sync message: %w", err)
			return nil
		}
		if err := rw.Flush(); err != nil {
			log.Println("Error flushing write buffer: %w", err)
			return nil
		}

		// Wait for half the RTT
		time.Sleep(rtt / 2)
		rtkGlobal.RTT[responseMsg.ObservedAddrs] = rtt

		// Perform simultaneous open for hole punching
		connP2P = performSimultaneousOpen(s.Conn(), rtkGlobal.NodeInfo.IPAddr.LocalPort, responseMsg.ObservedAddrs)
		if connP2P != nil {
			break
		}

		time.Sleep(time.Second)
	}
	return connP2P
}

// FIXME: fix this
func ExecuteP2PConnect(ctx context.Context, stream network.Stream, node host.Host) {
	// connP2P := performDCUtRHandshake(stream)
	// if connP2P != nil {
	// 	ipAddr := connP2P.RemoteAddr().String()

	// 	connCtx, cancel := context.WithCancel(context.Background())
	// 	rtkUtils.GoSafe(func() {rtkPeer2Peer.ProcessEventsForPeer(connP2P, ipAddr, connCtx, cancel)})
	// 	rtkUtils.GoSafe(func() {
	// 		<-connCtx.Done()
	// 		log.Println("************************************************")
	// 		log.Println("Lost connection with ID:", stream.Conn().RemotePeer(), " IP:", ipAddr)
	// 		log.Println("************************************************")
	// 	})
	// }
}

// FIXME: fix this
func HandleStream(s network.Stream) {
	// rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	// myAddr := rtkUtils.ConcatIP(rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort)
	// var connP2P net.Conn

	// rtkUtils.GoSafe(func() {
	// 	defer s.Close()
	// 	for {
	// 		log.Println("Wait for CONNECT ...")
	// 		var connectMsg rtkCommon.ConnectMessage
	// 		err := json.NewDecoder(rw).Decode(&connectMsg)
	// 		if err != nil {
	// 			log.Println("Error decoding connect message: %w", err)
	// 			return
	// 		}

	// 		log.Println("Received Connect message with observed addresses: ", connectMsg.ObservedAddrs)

	// 		// Respond with a Connect message
	// 		responseMsg := rtkCommon.ConnectMessage{Tag: "CONNECT", ObservedAddrs: myAddr}
	// 		if err := json.NewEncoder(rw).Encode(responseMsg); err != nil {
	// 			log.Println("Error encoding connect message: %w", err)
	// 			return
	// 		}
	// 		if err := rw.Flush(); err != nil {
	// 			log.Println("Error flushing write buffer: %w", err)
	// 			return
	// 		}

	// 		// Read Sync message
	// 		var syncMsg rtkCommon.SyncMessage
	// 		err = json.NewDecoder(rw).Decode(&syncMsg)
	// 		if err != nil {
	// 			log.Println("Error decoding sync message: %w", err)
	// 			return
	// 		}

	// 		log.Println("Received Sync message:", syncMsg)

	// 		// Perform simultaneous open for hole punching
	// 		connP2P = performSimultaneousOpen(s.Conn(), rtkGlobal.NodeInfo.IPAddr.LocalPort, connectMsg.ObservedAddrs)
	// 		if connP2P != nil {
	// 			break
	// 		}
	// 	}

	// 	ipAddr := connP2P.RemoteAddr().String()

	// 	connCtx, cancel := context.WithCancel(context.Background())
	// 	rtkUtils.GoSafe(func() {rtkPeer2Peer.ProcessEventsForPeer(connP2P, ipAddr, connCtx, cancel)})
	// 	rtkUtils.GoSafe(func() {
	// 		<-connCtx.Done()
	// 		log.Println("************************************************")
	// 		log.Println("Lost connection with ID:", s.Conn().RemotePeer(), " IP:", ipAddr)
	// 		log.Println("************************************************")
	// 	})
	// })
}
