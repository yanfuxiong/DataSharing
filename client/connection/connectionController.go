package connection

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"io"
	"log"
	"net"
	rtkBuildConfig "rtk-cross-share/client/buildConfig"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkLogin "rtk-cross-share/client/login"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"

	"github.com/libp2p/go-libp2p"

	"sync"
	"time"

	"github.com/libp2p/go-libp2p/p2p/protocol/ping"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
)

func getMutex(id string) *sync.Mutex {
	value, ok := mutexMap.LoadOrStore(id, &sync.Mutex{})
	if !ok {
		mutex := &sync.Mutex{}
		mutexMap.Store(id, mutex)
		return mutex
	}
	return value.(*sync.Mutex)
}

func GetConnectNode() host.Host {
	return node
}

func ConnectionInit(ctx context.Context) {
	log.Printf("[%s] listen host[%s] port[%d] connection start init", rtkMisc.GetFuncInfo(), rtkGlobal.ListenHost, rtkGlobal.ListenPort)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		if setupNode(rtkGlobal.ListenHost, rtkGlobal.ListenPort) == nil {
			break
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}

	nodeMutex.RLock()
	defer nodeMutex.RUnlock()
	if node == nil {
		log.Fatalf("[%s] node is nil!", rtkMisc.GetFuncInfo())
	}
	pingServer = ping.NewPingService(node)
	log.Printf("[%s] connection init success!\n\n", rtkMisc.GetFuncInfo())
}

func cancelHostNode() {
	wait()

	nodeMutex.Lock()
	if node != nil {
		node.Peerstore().Close()
		node.Network().Close()
		node.ConnManager().Close()
		node.Close()
		node = nil
		log.Println("close p2p node info success!")
	}
	if fileTransNode != nil {
		fileTransNode.Peerstore().Close()
		fileTransNode.Network().Close()
		fileTransNode.Close()
		fileTransNode = nil
		log.Println("close p2p file Node info success!")
	}
	nodeMutex.Unlock()
}

func Run(ctx context.Context) {
	defer cancelHostNode()
	defer CancelAllStream(false)

	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows { // only windows need watch network info
		rtkMisc.GoSafe(func() { WatchNetworkInfo(ctx) })
	}

	buildListener(ctx)

	ticker := time.NewTicker(rtkCommon.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("connectionController Run is end by main context!")
			return
		case <-ticker.C:
			CheckAllStreamAlive(ctx)
		case clientList := <-rtkLogin.GetClientListFlag:
			rtkMisc.GoSafe(func() { buildClientListTalker(ctx, clientList) })
		}
	}
}

func setupNode(ip string, port int) error {
	priv := rtkPlatform.GenKey()

	cancelHostNode()

	tcpAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, port))
	if err != nil {
		log.Printf("NewMultiaddr tcp addr error:%+v", err)
		return err
	}

	// UDP listen ip must be 0.0.0.0, Otherwise it will cause the android talker end to crash
	quicAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/udp/%d/quic-v1", rtkMisc.DefaultIp, rtkGlobal.DefaultPort))
	if err != nil {
		log.Printf("NewMultiaddr quic addr error:%+v", err)
		return err
	}

	if port <= 0 {
		log.Println("p2p listen port is not set. Use a random port")
	}

	tempNode, err := libp2p.New(
		//libp2p.ListenAddrStrings(listen_addrs(rtkMdns.MdnsCfg.ListenPort)...), // Add mdns port with different initialization
		libp2p.ListenAddrs(tcpAddr),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.NATPortMap(),
		libp2p.Identity(priv),
		libp2p.ForceReachabilityPrivate(),
		libp2p.ResourceManager(&network.NullResourceManager{}),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay(),
		libp2p.Ping(true),
	)
	if err != nil {
		log.Printf("Failed to create node: %v", err)
		return err
	}

	tempFileNode, err := libp2p.New(
		libp2p.ListenAddrs(quicAddr),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.ResourceManager(&network.NullResourceManager{}),
		libp2p.DisableRelay(),
	)
	if err != nil {
		log.Printf("Failed to create file node: %v", err)
		return err
	}

	if len(tempNode.Addrs()) == 0 {
		log.Printf("Failed to create node, Addrs is null!")
		tempNode.Close()
		return fmt.Errorf("addr is null!")
	}

	for _, p := range tempNode.Peerstore().Peers() {
		tempNode.Peerstore().ClearAddrs(p)
	}

	if rtkPlatform.IsHost() {
		//rtkUtils.WriteNodeID(tempNode.ID().String(), rtkPlatform.GetHostIDPath())
	}

	rtkUtils.WriteNodeID(tempNode.ID().String(), rtkPlatform.GetIDPath())

	rtkGlobal.NodeInfo.IPAddr.LocalPort = rtkUtils.GetLocalPort(tempNode.Addrs())
	rtkGlobal.NodeInfo.ID = tempNode.ID().String()

	rtkGlobal.NodeInfo.IPAddr.UpdPort = rtkUtils.GetQuicPort(tempFileNode.Addrs())
	rtkGlobal.NodeInfo.FileTransNodeID = tempFileNode.ID().String()

	// filter IP by skip [0.0.0.0, 127.0.0.1]
	filterAddr := func(addrs []ma.Multiaddr) (string, string) {
		for _, addr := range addrs {
			ip, port := rtkUtils.ExtractTCPIPandPort(addr)
			if ip != rtkMisc.DefaultIp && ip != rtkMisc.LoopBackIp {
				return ip, port
			}
		}
		return rtkUtils.ExtractTCPIPandPort(addrs[0])
	}
	publicIp, publicPort := filterAddr(tempNode.Addrs())
	rtkGlobal.NodeInfo.IPAddr.PublicIP = publicIp
	rtkGlobal.NodeInfo.IPAddr.PublicPort = publicPort
	serviceVer := "v" + rtkGlobal.ClientVersion + " (" + rtkBuildConfig.BuildDate + ")"

	log.Println("=======================================================")
	log.Println("Self ID: ", rtkGlobal.NodeInfo.ID)
	log.Println("Self node Addr: ", tempNode.Addrs())
	log.Println("Self listen Addr: ", tempNode.Network().ListenAddresses())
	log.Println("Self device name: ", rtkGlobal.NodeInfo.DeviceName)
	log.Println("Self Platform: ", rtkGlobal.NodeInfo.Platform)
	log.Printf("Self Public IP[%s], Public Port[%s], LocalPort[%s]", rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort, rtkGlobal.NodeInfo.IPAddr.LocalPort)
	log.Println("Self version info: ", serviceVer)
	log.Println("Self file node ID: ", rtkGlobal.NodeInfo.FileTransNodeID)
	log.Println("Self file node Addr: ", tempFileNode.Addrs())
	log.Println("Self file node listen Addr: ", tempFileNode.Network().ListenAddresses())
	log.Println("Self file node port: ", rtkGlobal.NodeInfo.IPAddr.UpdPort)
	log.Println("=======================================================")

	ipAddr := rtkMisc.ConcatIP(rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort)
	rtkPlatform.GoUpdateSystemInfo(ipAddr, serviceVer)

	nodeMutex.Lock()
	node = tempNode
	fileTransNode = tempFileNode
	nodeMutex.Unlock()

	return nil
}

func buildListener(ctx context.Context) {
	nodeMutex.RLock()
	defer nodeMutex.RUnlock()
	if node == nil {
		log.Fatalf("[%s] node is nil!", rtkMisc.GetFuncInfo())
	}

	node.SetStreamHandler(protocol.ID(rtkGlobal.ProtocolDirectID), func(stream network.Stream) {
		onlineEvent(ctx, stream, true, nil)
	})

	node.SetStreamHandler(protocol.ID(rtkGlobal.ProtocolImageTransmission), func(stream network.Stream) {
		updateFmtTypeStreamSrc(stream, rtkCommon.XCLIP_CB)
		noticeFmtTypeStreamReady(stream.Conn().RemotePeer().String(), rtkCommon.XCLIP_CB)
	})

	node.SetStreamHandler(protocol.ID(rtkGlobal.ProtocolFileTransmission), func(stream network.Stream) {
		updateFmtTypeStreamSrc(stream, rtkCommon.FILE_DROP)
		noticeFmtTypeStreamReady(stream.Conn().RemotePeer().String(), rtkCommon.FILE_DROP)
	})

}

func BuildFileDropItemStreamListener(timestamp uint64) {
	nodeMutex.Lock()
	defer nodeMutex.Unlock()
	if fileTransNode == nil {
		log.Printf("[%s] node is nil! set protocol handler failed!, timestamp:[%d]", rtkMisc.GetFuncInfo(), timestamp)
		return
	}
	fileTransNode.SetStreamHandler(protocol.ID(getFileDropStreamProtocol(timestamp)), handlerFileDropItemStream)
	log.Printf("[%s] set protocol handler success, timestamp:[%d]", rtkMisc.GetFuncInfo(), timestamp)
}

func RemoveFileDropItemStreamListener(timestamp uint64) {
	nodeMutex.Lock()
	defer nodeMutex.Unlock()
	if fileTransNode == nil {
		log.Printf("[%s] node is nil! remove protocol handler failed!, timestamp:[%d]", rtkMisc.GetFuncInfo(), timestamp)
		return
	}
	fileTransNode.RemoveStreamHandler(protocol.ID(getFileDropStreamProtocol(timestamp)))
	log.Printf("[%s] remove protocol handler success, timestamp:[%d]", rtkMisc.GetFuncInfo(), timestamp)
}

func handlerFileDropItemStream(stream network.Stream) {
	Reader := bufio.NewReader(stream)
	var reqMsg FileDropItemStreamInfo
	err := json.NewDecoder(Reader).Decode(&reqMsg)
	if err != nil {
		log.Printf("[%s] ID:[%s] Stream ID:[%s] json.NewDecoder.Decode err:%+v", rtkMisc.GetFuncInfo(), stream.Conn().RemotePeer().String(), stream.ID(), err)
		return
	}
	reqMsg.StreamId = stream.ID()
	addFileDropItemStreamAsSrc(reqMsg.ID, reqMsg.Timestamp, stream)
	noticeFmtTypeStreamReady(reqMsg.ID, rtkCommon.FILE_DROP)
}

func NewFileDropItemStream(ctx context.Context, id string, timestamp uint64) rtkMisc.CrossShareErr {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout_short)
	defer cancel()

	clientInfo, err := rtkUtils.GetClientInfo(id)
	if err != nil {
		log.Printf("[%s] ID:[%s] Not found Client Info data", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_GET_CLIENT_INFO_EMPTY
	}
	ip, _ := rtkUtils.SplitIPAddr(clientInfo.IpAddr)
	quicAddr := fmt.Sprintf("/ip4/%s/udp/%s/quic-v1", ip, clientInfo.UpdPort)
	addr := ma.StringCast(quicAddr)
	idB58, err := peer.Decode(clientInfo.FileTransNodeID)
	if err != nil {
		log.Printf("[%s] ID decode failed: %s", rtkMisc.GetFuncInfo(), clientInfo.ID)
		return rtkMisc.ERR_BIZ_P2P_PEER_DECODE
	}
	fileTransPeer := peer.AddrInfo{
		ID:    idB58,
		Addrs: []ma.Multiaddr{addr},
	}

	startTime := time.Now().UnixMilli()
	nodeMutex.RLock()
	defer nodeMutex.RUnlock()
	if fileTransNode.Network().Connectedness(fileTransPeer.ID) != network.Connected {
		fileTransNode.Network().ClosePeer(fileTransPeer.ID)
		log.Printf("begin  to connect %+v ...", fileTransPeer)
		if err = fileTransNode.Connect(ctx, fileTransPeer); err != nil {
			log.Printf("[%s] Connect peer%+v failed:%+v", rtkMisc.GetFuncInfo(), fileTransPeer, err)
			if errors.Is(err, context.DeadlineExceeded) {
				return rtkMisc.ERR_NETWORK_P2P_CONNECT_DEADLINE
			} else if errors.Is(err, context.Canceled) {
				return rtkMisc.ERR_NETWORK_P2P_CONNECT_CANCEL
			} else if netErr, ok := err.(net.Error); ok {
				log.Printf("[Socket][%s] Err: Read fail network error(%v)", rtkMisc.GetFuncInfo(), netErr.Error())
				if netErr.Timeout() {
					return rtkMisc.ERR_NETWORK_P2P_TIMEOUT
				}
			}
			return rtkMisc.ERR_NETWORK_P2P_CONNECT
		}
		log.Printf("connect %s success! use [%d] ms", fileTransPeer.ID.String(), time.Now().UnixMilli()-startTime)
	}

	protocolId := getFileDropStreamProtocol(timestamp)
	stream, err := fileTransNode.NewStream(ctx, fileTransPeer.ID, protocol.ID(protocolId))
	if err != nil {
		log.Printf("[%s] ID:[%s] IP:[%+v] open protocolId:%s stream failed:%+v", rtkMisc.GetFuncInfo(), id, fileTransPeer.Addrs, protocolId, err)
		if errors.Is(err, context.DeadlineExceeded) {
			return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM_DEADLINE
		} else if errors.Is(err, context.Canceled) {
			return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM_CANCEL
		}
		return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM
	}
	Writer := bufio.NewWriter(stream)
	registMsg := FileDropItemStreamInfo{
		Timestamp: timestamp,
		ID:        rtkGlobal.NodeInfo.ID,
		StreamId:  stream.ID(),
	}
	if err = json.NewEncoder(Writer).Encode(registMsg); err != nil {
		log.Println("failed to send register message: %w", err)
		return rtkMisc.ERR_NETWORK_P2P_WRITER
	}
	if err = Writer.Flush(); err != nil {
		log.Printf("[%s] ID:[%s] Error flushing write buffer: %+v", rtkMisc.GetFuncInfo(), id, err)
		return rtkMisc.ERR_NETWORK_P2P_FLUSH
	}

	addFileDropItemStreamAsDst(id, timestamp, stream)
	log.Printf("ID:[%s] new a file drop stream success! use [%d] ms", id, time.Now().UnixMilli()-startTime)
	return rtkMisc.SUCCESS
}

func getFileDropStreamProtocol(timestamp uint64) string {
	return fmt.Sprintf("%s%d", rtkGlobal.ProtocolFileTransQueue, timestamp)
}

func buildTalker(ctxMain context.Context, client rtkMisc.ClientInfo) rtkMisc.CrossShareErr {
	ip, port := rtkUtils.SplitIPAddr(client.IpAddr)
	ipAddr := fmt.Sprintf("/ip4/%s/tcp/%s", ip, port)
	addr := ma.StringCast(ipAddr)
	idB58, err := peer.Decode(client.ID)
	if err != nil {
		log.Printf("[%s] ID decode failed: %s", rtkMisc.GetFuncInfo(), client.ID)
		return rtkMisc.ERR_BIZ_P2P_PEER_DECODE
	}
	peer := peer.AddrInfo{
		ID:    idB58,
		Addrs: []ma.Multiaddr{addr},
	}

	nodeMutex.Lock()
	defer nodeMutex.Unlock()
	if node == nil {
		log.Printf("[%s] node is nil!", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_BIZ_P2P_NODE_NULL
	}

	ctx, cancel := context.WithTimeout(ctxMain, ctxTimeout_normal)
	defer cancel()

	if node.Network().Connectedness(peer.ID) != network.Connected {
		startTime := time.Now().UnixMilli()
		node.Network().ClosePeer(peer.ID)
		log.Printf("begin to connect %+v ...", peer)
		if err = node.Connect(ctx, peer); err != nil {
			log.Printf("[%s] Connect peer%+v failed:%+v", rtkMisc.GetFuncInfo(), peer, err)
			if errors.Is(err, context.DeadlineExceeded) {
				return rtkMisc.ERR_NETWORK_P2P_CONNECT_DEADLINE
			} else if errors.Is(err, context.Canceled) {
				return rtkMisc.ERR_NETWORK_P2P_CONNECT_CANCEL
			} else if netErr, ok := err.(net.Error); ok {
				log.Printf("[Socket][%s] Err: Read fail network error(%v)", rtkMisc.GetFuncInfo(), netErr.Error())
				if netErr.Timeout() {
					return rtkMisc.ERR_NETWORK_P2P_TIMEOUT
				}
			}
			return rtkMisc.ERR_NETWORK_P2P_CONNECT
		}
		log.Printf("connect %s success! use [%d] ms", peer.ID.String(), time.Now().UnixMilli()-startTime)
	}

	if IsStreamExisted(peer.ID.String()) {
		// DEBUG
		// log.Printf("[%s] ID:[%s] a Stream is already existed, skip NewStream.", rtkMisc.GetFuncInfo(), peer.ID.String())
		return rtkMisc.SUCCESS
	}

	stream, err := node.NewStream(ctx, peer.ID, protocol.ID(rtkGlobal.ProtocolDirectID))
	if err != nil {
		log.Printf("[%s] ID:[%s] open a stream failed:%+v", rtkMisc.GetFuncInfo(), peer.ID.String(), err)
		if errors.Is(err, context.DeadlineExceeded) {
			return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM_DEADLINE
		} else if errors.Is(err, context.Canceled) {
			return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM_CANCEL
		}
		return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM
	}

	return onlineEvent(ctxMain, stream, false, &client)
}

func BuildFmtTypeTalker(ctx context.Context, id string, fmtType rtkCommon.TransFmtType) rtkMisc.CrossShareErr {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout_short)
	defer cancel()

	sInfo, ok := GetStreamInfo(id)
	if !ok {
		log.Printf("[%s] ID:[%s] IP[%s] get no stream info or stream is closed", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr)
		return rtkMisc.ERR_BIZ_GET_STREAM_EMPTY
	}
	clearOldFmtStream(id, fmtType)

	nodeMutex.RLock()
	defer nodeMutex.RUnlock()
	var fmtTypeStream network.Stream
	var err error
	if fmtType == rtkCommon.FILE_DROP {
		fmtTypeStream, err = node.NewStream(ctx, sInfo.s.Conn().RemotePeer(), protocol.ID(rtkGlobal.ProtocolFileTransmission))
		if err != nil {
			log.Printf("[%s] ID:[%s] IP:[%s] open %s stream failed:%+v", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr, fmtType, err)
			if errors.Is(err, context.DeadlineExceeded) {
				return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM_DEADLINE
			} else if errors.Is(err, context.Canceled) {
				return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM_CANCEL
			}
			return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM
		}
	} else if fmtType == rtkCommon.XCLIP_CB {
		fmtTypeStream, err = node.NewStream(ctx, sInfo.s.Conn().RemotePeer(), protocol.ID(rtkGlobal.ProtocolImageTransmission))
		if err != nil {
			log.Printf("[%s] ID:[%s] IP:[%s] open %s stream failed:%+v", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr, fmtType, err)
			if errors.Is(err, context.DeadlineExceeded) {
				return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM_DEADLINE
			} else if errors.Is(err, context.Canceled) {
				return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM_CANCEL
			}
			return rtkMisc.ERR_NETWORK_P2P_OPEN_STREAM
		}
	} else {
		log.Printf("[%s] ID:[%s] IP:[%s]BuildFmtTypeTalker failed! Unknown fmtType:[%s]", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr, fmtType)
		return rtkMisc.ERR_BIZ_UNKNOWN_FMTTYPE
	}

	updateFmtTypeStreamDst(fmtTypeStream, fmtType)
	return rtkMisc.SUCCESS
}

func buildClientListTalker(ctx context.Context, clientList []rtkMisc.ClientInfo) {
	if len(clientList) == 0 {
		return
	}

	for _, client := range clientList {
		if client.ID == "" || client.IpAddr == "" {
			log.Printf("ClientListFromLanServer get ID:[%s] IPAddr:[%s] , continue!", client.ID, client.IpAddr)
			continue
		}

		ip, port := rtkUtils.SplitIPAddr(client.IpAddr)
		if ip == "" || port == "" || ip == rtkMisc.DefaultIp || ip == rtkMisc.LoopBackIp {
			log.Printf("ClientListFromLanServer get ID:[%s] IPAddr:[%s] , continue!", client.ID, client.IpAddr)
			continue
		}

		rtkMisc.GoSafeWithParam(func(args ...any) {
			errCode := buildTalker(ctx, client)
			if errCode != rtkMisc.SUCCESS {
				log.Printf("[%s] ID:[%s] IPAddr:[%s] buildTalker failed, errCode:%d ", rtkMisc.GetFuncInfo(), client.ID, client.IpAddr, errCode)
			}
		}, client)
	}
}

func WriteSocket(id string, data []byte) rtkMisc.CrossShareErr {
	if len(data) == 0 {
		log.Println("Write faile: empty data")
		return rtkMisc.ERR_BIZ_P2P_WRITE_EMPTY_DATA
	}

	sInfo, ok := GetStreamInfo(id)
	if !ok {
		log.Printf("[%s][%s] WriteSocket err, get no stream or stream is closed", rtkMisc.GetFuncInfo(), sInfo.ipAddr)
		return rtkMisc.ERR_BIZ_GET_STREAM_EMPTY
	}

	if _, err := sInfo.s.Write(data); err != nil {
		if CheckStreamReset(id, sInfo.timeStamp) {
			return rtkMisc.ERR_BIZ_GET_STREAM_RESET
		}

		log.Printf("[%s][%s] Write failed:[%+v], and execute offlineEvent ", rtkMisc.GetFuncInfo(), sInfo.ipAddr, err)
		isEOF := false
		defer offlineEvent(sInfo.s, isEOF)

		if errors.Is(err, io.EOF) {
			isEOF = true
			return rtkMisc.ERR_NETWORK_P2P_EOF
		} else if netErr, ok := err.(net.Error); ok {
			log.Printf("[Socket][%s] Err: Read fail network error(%v)", sInfo.ipAddr, netErr.Error())
			if netErr.Timeout() {
				return rtkMisc.ERR_NETWORK_P2P_TIMEOUT
			}
		} else if errors.Is(err, network.ErrReset) {
			return rtkMisc.ERR_NETWORK_P2P_RESET
		} else if errors.Is(err, network.ErrNoConn) {
			return rtkMisc.ERR_NETWORK_P2P_CONNECT
		}
	}

	bufio.NewWriter(sInfo.s).Flush()
	log.Printf("[%s] DST ID:[%s] Write to socket successfully", rtkMisc.GetFuncInfo(), id)
	return rtkMisc.SUCCESS
}

func ReadSocket(id string, buffer []byte) (int, rtkMisc.CrossShareErr) {
	sInfo, ok := GetStreamInfo(id)
	if !ok {
		return 0, rtkMisc.ERR_BIZ_GET_STREAM_EMPTY
	}

	sInfo.s.SetReadDeadline(time.Time{}) //Cancel timeout limit
	n, err := sInfo.s.Read(buffer)
	if err != nil {
		if CheckStreamReset(id, sInfo.timeStamp) {
			return 0, rtkMisc.ERR_BIZ_GET_STREAM_RESET
		}

		log.Printf("[%s][%s] Read failed [%+v],  execute offlineEvent ", rtkMisc.GetFuncInfo(), sInfo.ipAddr, err)

		isEOF := false
		defer offlineEvent(sInfo.s, isEOF)

		if errors.Is(err, io.EOF) {
			isEOF = true
			return n, rtkMisc.ERR_NETWORK_P2P_EOF
		} else if netErr, ok := err.(net.Error); ok {
			log.Printf("[Socket][%s] Err: Read fail network error(%v)", sInfo.ipAddr, netErr.Error())
			if netErr.Timeout() {
				return n, rtkMisc.ERR_NETWORK_P2P_TIMEOUT
			}
		} else if errors.Is(err, network.ErrReset) {
			return n, rtkMisc.ERR_NETWORK_P2P_RESET
		} else if errors.Is(err, network.ErrNoConn) {
			return n, rtkMisc.ERR_NETWORK_P2P_CONNECT
		}

		return n, rtkMisc.ERR_NETWORK_P2P_OTHER
	}

	return n, rtkMisc.SUCCESS
}

func onlineEvent(ctx context.Context, stream network.Stream, isFromListener bool, clientInfo *rtkMisc.ClientInfo) rtkMisc.CrossShareErr {
	id := stream.Conn().RemotePeer().String()
	mutex := getMutex(id)
	mutex.Lock()
	defer mutex.Unlock()

	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)
	var peerDeviceName, peerPlatForm, srcPortType, peerVer, peerFileTransID, peerUdpPort string
	if isFromListener {
		resultCode := handleNotice(stream, &peerPlatForm, &peerDeviceName, &srcPortType, &peerVer, &peerFileTransID, &peerUdpPort)
		if resultCode != rtkMisc.SUCCESS {
			stream.Reset()
			log.Printf("[%s] ID:[%s] IP:[%s] errCode:%d, so reset this stream, onlineEvent failed!", rtkMisc.GetFuncInfo(), id, ipAddr, resultCode)
			return resultCode
		}
	} else {
		resultCode := noticeToPeer(stream, &peerVer, &peerFileTransID, &peerUdpPort)
		if resultCode != rtkMisc.SUCCESS {
			stream.Reset()
			log.Printf("[%s] ID:[%s] IP:[%s] errCode:%d, so reset this stream, onlineEvent failed!", rtkMisc.GetFuncInfo(), id, ipAddr, resultCode)
			return resultCode
		}
		peerDeviceName = clientInfo.DeviceName
		peerPlatForm = clientInfo.Platform
		srcPortType = clientInfo.SourcePortType
	}

	updateStream(ctx, id, stream)
	log.Println("****************************************************************************************")
	if isFromListener {
		log.Println("Connected from ID:", id, " IP:", ipAddr)
	} else {
		log.Println("Connected to ID:", id, " IP:", ipAddr)
	}
	log.Println("****************************************************************************************")

	updateUIOnlineStatus(true, id, ipAddr, peerPlatForm, peerDeviceName, srcPortType, peerVer, peerFileTransID, peerUdpPort)
	return rtkMisc.SUCCESS
}

func offlineEvent(stream network.Stream, isFromPeer bool) {
	id := stream.Conn().RemotePeer().String()
	mutex := getMutex(id)
	mutex.Lock()
	defer mutex.Unlock()

	// Disconnect and remove stream map
	closeStream(id, isFromPeer)

	clientInfo, err := rtkUtils.GetClientInfo(id)
	if err == nil {
		updateUIOnlineStatus(false, id, clientInfo.IpAddr, clientInfo.Platform, clientInfo.DeviceName, clientInfo.SourcePortType, "", "", "")
	} else {
		log.Printf("[%s] %s, so not need updateUIOnlineStatus!", rtkMisc.GetFuncInfo(), err.Error())
		return
	}

	log.Println("************************************************************************************************")
	log.Println("Lost connection with ID:", id, " IP:", clientInfo.IpAddr)
	log.Println("************************************************************************************************\n\n")
}

func OfflineEvent(id string) {
	s, bExisted := GetStream(id)
	if !bExisted {
		log.Printf("[%s] the stream is already not existed!", id)
		return
	}
	log.Printf("ID:[%s] offline event come from peer!", id)
	offlineEvent(s, true)
}

func updateUIOnlineStatus(isOnline bool, id, ipAddr, platfrom, deviceName, srcPortType, ver, fileTransId, udpPort string) {
	if isOnline {
		log.Printf("[%s] IP:[%s] Online: increase client count\n\n", rtkMisc.GetFuncInfo(), ipAddr)
		rtkPlatform.GoUpdateClientStatus(1, ipAddr, id, deviceName, srcPortType) // TODO: Deprecate , and replace with GoUpdateClientStatusEx
		rtkUtils.InsertClientInfoMap(id, ipAddr, platfrom, deviceName, srcPortType, ver, fileTransId, udpPort)
		rtkPlatform.FoundPeer() // TODO: Deprecate , and replace with GoUpdateClientStatusEx
		rtkPlatform.GoUpdateClientStatusEx(id, 1)
	} else {
		log.Printf("[%s] IP:[%s] Offline: decrease client count", rtkMisc.GetFuncInfo(), ipAddr)
		rtkPlatform.GoUpdateClientStatus(0, ipAddr, id, deviceName, srcPortType) // TODO: Deprecate , and replace with GoUpdateClientStatusEx
		rtkUtils.LostClientInfoMap(id)
		rtkPlatform.FoundPeer() // TODO: Deprecate , and replace with GoUpdateClientStatusEx
		rtkPlatform.GoUpdateClientStatusEx(id, 0)
	}
}

func noticeToPeer(s network.Stream, ver, fileTransId, udpPort *string) rtkMisc.CrossShareErr {
	ipAddr := rtkUtils.GetRemoteAddrFromStream(s)
	id := s.Conn().RemotePeer().String()
	registMsg := rtkCommon.RegistMdnsMessage{
		Host:            rtkPlatform.GetHostID(),
		Id:              rtkGlobal.NodeInfo.ID,
		Platform:        rtkGlobal.NodeInfo.Platform,
		DeviceName:      rtkGlobal.NodeInfo.DeviceName,
		SourcePortType:  rtkGlobal.NodeInfo.SourcePortType,
		Version:         rtkGlobal.ClientVersion,
		FileTransNodeID: rtkGlobal.NodeInfo.FileTransNodeID,
		UdpPort:         rtkGlobal.NodeInfo.IPAddr.UpdPort,
	}

	write := bufio.NewWriter(s)
	err := json.NewEncoder(write).Encode(registMsg)
	if err != nil {
		log.Printf("[%s] ID:[%s] IP:[%s] Stream json.NewEncoder.Encode err:%+v", rtkMisc.GetFuncInfo(), id, ipAddr, err)
		if errors.Is(err, context.DeadlineExceeded) {
			return rtkMisc.ERR_NETWORK_P2P_WRITER_DEADLINE
		} else if errors.Is(err, context.Canceled) {
			return rtkMisc.ERR_NETWORK_P2P_WRITER_CANCELED
		} else if errors.Is(err, network.ErrReset) {
			return rtkMisc.ERR_NETWORK_P2P_WRITER_RESET
		}
		return rtkMisc.ERR_NETWORK_P2P_WRITER
	}
	if err = write.Flush(); err != nil {
		log.Printf("[%s] ID:[%s] Error flushing write buffer: %+v", rtkMisc.GetFuncInfo(), id, err)
		return rtkMisc.ERR_NETWORK_P2P_FLUSH
	}

	reqMsg := rtkCommon.RegistMdnsMessage{Version: ""}
	s.SetReadDeadline(time.Now().Add(1 * time.Second)) //Only valid for the current goroutine
	read := bufio.NewReader(s)
	err = json.NewDecoder(read).Decode(&reqMsg)
	if err != nil {
		s.SetReadDeadline(time.Time{})
		log.Printf("[%s] ID:[%s] IP:[%s] Stream json.NewDecoder.Decode err:%+v", rtkMisc.GetFuncInfo(), id, ipAddr, err)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			*ver = rtkGlobal.ClientDefaultVersion
			log.Printf("[%s] IP:[%s] handle decoder time out(1s)! use defalut value:%s", rtkMisc.GetFuncInfo(), ipAddr, rtkGlobal.ClientDefaultVersion)
		} else if errors.Is(err, context.DeadlineExceeded) {
			return rtkMisc.ERR_NETWORK_P2P_READER_DEADLINE
		} else if errors.Is(err, context.Canceled) {
			return rtkMisc.ERR_NETWORK_P2P_READER_CANCELED
		} else if errors.Is(err, network.ErrReset) {
			return rtkMisc.ERR_NETWORK_P2P_READER_RESET
		} else {
			return rtkMisc.ERR_NETWORK_P2P_READER
		}
	} else {
		log.Printf("[%s] IP:[%s] handle decoder success! verison:%s", rtkMisc.GetFuncInfo(), ipAddr, reqMsg.Version)
		*ver = reqMsg.Version
		*fileTransId = reqMsg.FileTransNodeID
		*udpPort = reqMsg.UdpPort
	}

	return rtkMisc.SUCCESS
}

func handleNotice(s network.Stream, platForm, name, srcPortType, ver, fileTransId, udpPort *string) rtkMisc.CrossShareErr {
	id := s.Conn().RemotePeer().String()
	ipAddr := rtkUtils.GetRemoteAddrFromStream(s)

	regMsg := rtkCommon.RegistMdnsMessage{Version: ""}
	read := bufio.NewReader(s)
	err := json.NewDecoder(read).Decode(&regMsg)
	if err != nil {
		log.Printf("[%s] ID:[%s] IP:[%s] json.NewDecoder.Decode err:%+v", rtkMisc.GetFuncInfo(), id, ipAddr, err)
		if errors.Is(err, context.DeadlineExceeded) {
			return rtkMisc.ERR_NETWORK_P2P_READER_DEADLINE
		} else if errors.Is(err, context.Canceled) {
			return rtkMisc.ERR_NETWORK_P2P_READER_CANCELED
		} else if errors.Is(err, network.ErrReset) {
			return rtkMisc.ERR_NETWORK_P2P_READER_RESET
		}
		return rtkMisc.ERR_NETWORK_P2P_READER
	}

	*platForm = regMsg.Platform
	*name = regMsg.DeviceName
	*srcPortType = regMsg.SourcePortType
	*fileTransId = regMsg.FileTransNodeID
	*udpPort = regMsg.UdpPort

	if regMsg.Version == "" { // old version
		*ver = rtkGlobal.ClientDefaultVersion
		log.Printf("[%s] IP:[%s] handle decoder success! version is null, use defalut value:%s", rtkMisc.GetFuncInfo(), ipAddr, rtkGlobal.ClientDefaultVersion)
	} else {
		*ver = regMsg.Version
		registMsg := rtkCommon.RegistMdnsMessage{
			Host:            rtkPlatform.GetHostID(),
			Id:              rtkGlobal.NodeInfo.ID,
			Platform:        rtkGlobal.NodeInfo.Platform,
			DeviceName:      rtkGlobal.NodeInfo.DeviceName,
			SourcePortType:  rtkGlobal.NodeInfo.SourcePortType,
			Version:         rtkGlobal.ClientVersion,
			FileTransNodeID: rtkGlobal.NodeInfo.FileTransNodeID,
			UdpPort:         rtkGlobal.NodeInfo.IPAddr.UpdPort,
		}

		write := bufio.NewWriter(s)
		if err = json.NewEncoder(write).Encode(registMsg); err != nil {
			log.Printf("[%s] ID:[%s] IP:[%s] json.NewDecoder.Decode err:%+v", rtkMisc.GetFuncInfo(), id, ipAddr, err)
			if errors.Is(err, context.DeadlineExceeded) {
				return rtkMisc.ERR_NETWORK_P2P_WRITER_DEADLINE
			} else if errors.Is(err, context.Canceled) {
				return rtkMisc.ERR_NETWORK_P2P_WRITER_CANCELED
			} else if errors.Is(err, network.ErrReset) {
				return rtkMisc.ERR_NETWORK_P2P_WRITER_RESET
			}
			return rtkMisc.ERR_NETWORK_P2P_WRITER
		}
		if err = write.Flush(); err != nil {
			log.Printf("[%s] ID:[%s] Error flushing write buffer: %+v", rtkMisc.GetFuncInfo(), id, err)
			return rtkMisc.ERR_NETWORK_P2P_FLUSH
		}
		log.Printf("[%s] IP:[%s] handle encoder success! version:%s", rtkMisc.GetFuncInfo(), ipAddr, regMsg.Version)
	}

	return rtkMisc.SUCCESS
}
