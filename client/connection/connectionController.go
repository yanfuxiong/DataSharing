package connection

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	if setupNode(rtkGlobal.ListenHost, rtkGlobal.ListenPort) != nil {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if setupNode(rtkGlobal.ListenHost, rtkGlobal.ListenPort) == nil {
					goto setNodeSuccessFlag
				}
			}
		}
	}

setNodeSuccessFlag:
	nodeMutex.RLock()
	defer nodeMutex.RUnlock()
	if node == nil {
		log.Fatalf("[%s] node is nil!", rtkMisc.GetFuncInfo())
	}
	pingServer = ping.NewPingService(node)
}

func cancelHostNode() {
	nodeMutex.Lock()
	defer nodeMutex.Unlock()
	if node != nil {
		log.Println("begin close p2p node info!")
		node.Peerstore().Close()
		node.Network().Close()
		node.ConnManager().Close()
		node.Close()
		node = nil
		log.Println("close p2p node info!")
	}
}

func Run(ctx context.Context) {
	defer cancelHostNode()
	defer CancelStreamPool()

	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows { // only windows need watch network info
		rtkMisc.GoSafe(func() { WatchNetworkInfo(ctx) })
	}

	buildListener()

	ticker := time.NewTicker(pingInterval)
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
	sourceAddrStr := fmt.Sprintf("/ip4/%s/tcp/%d", ip, port)
	sourceMultiAddr, err := ma.NewMultiaddr(sourceAddrStr)
	if err != nil {
		log.Printf("NewMultiaddr error:%+v, with addr:%s", err, sourceAddrStr)
		panic(err)
	}

	if port <= 0 {
		log.Println("p2p listen port is not set. Use a random port")
	}

	tempNode, err := libp2p.New(
		//libp2p.ListenAddrStrings(listen_addrs(rtkMdns.MdnsCfg.ListenPort)...), // Add mdns port with different initialization
		libp2p.ListenAddrs(sourceMultiAddr),
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

	if len(tempNode.Addrs()) == 0 {
		log.Printf("Failed to create node, Addrs is null!")
		tempNode.Close()
		return fmt.Errorf("addr is null!")
	}

	tempNode.Network().Listen(tempNode.Addrs()...)
	for _, p := range tempNode.Peerstore().Peers() {
		tempNode.Peerstore().ClearAddrs(p)
	}

	if rtkPlatform.IsHost() {
		rtkUtils.WriteNodeID(tempNode.ID().String(), rtkPlatform.GetHostIDPath())
	}

	rtkUtils.WriteNodeID(tempNode.ID().String(), rtkPlatform.GetIDPath())

	rtkGlobal.NodeInfo.IPAddr.LocalPort = rtkUtils.GetLocalPort(tempNode.Addrs())
	rtkGlobal.NodeInfo.ID = tempNode.ID().String()

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

	log.Println("=======================================================")
	log.Println("Self ID: ", rtkGlobal.NodeInfo.ID)
	log.Println("Self node Addr: ", tempNode.Addrs())
	log.Println("Self listen Addr: ", tempNode.Network().ListenAddresses())
	log.Println("Self device name: ", rtkGlobal.NodeInfo.DeviceName)
	log.Println("Self Platform: ", rtkGlobal.NodeInfo.Platform)
	log.Printf("Self Public IP[%s], Public Port[%s], LocalPort[%s]", rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort, rtkGlobal.NodeInfo.IPAddr.LocalPort)
	log.Println("=======================================================\n\n")

	ipAddr := rtkMisc.ConcatIP(rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort)
	serviceVer := "v" + rtkGlobal.ClientVersion + " (" + rtkBuildConfig.BuildDate + ")"
	rtkPlatform.GoUpdateSystemInfo(ipAddr, serviceVer)

	nodeMutex.Lock()
	node = tempNode
	nodeMutex.Unlock()

	return nil
}

func buildListener() {
	nodeMutex.RLock()
	defer nodeMutex.RUnlock()
	if node == nil {
		log.Fatalf("[%s] node is nil!", rtkMisc.GetFuncInfo())
	}

	node.SetStreamHandler(protocol.ID(rtkGlobal.ProtocolDirectID), func(stream network.Stream) {
		onlineEvent(stream, true, nil)
	})

	node.SetStreamHandler(protocol.ID(rtkGlobal.ProtocolImageTransmission), func(stream network.Stream) {
		updateFmtTypeStreamSrc(stream, rtkCommon.IMAGE_CB)
		noticeFmtTypeStreamReady(stream.Conn().RemotePeer().String(), rtkCommon.IMAGE_CB)
	})

	node.SetStreamHandler(protocol.ID(rtkGlobal.ProtocolFileTransmission), func(stream network.Stream) {
		updateFmtTypeStreamSrc(stream, rtkCommon.FILE_DROP)
		noticeFmtTypeStreamReady(stream.Conn().RemotePeer().String(), rtkCommon.FILE_DROP)
	})
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

	nodeMutex.RLock()
	defer nodeMutex.RUnlock()
	if node == nil {
		log.Printf("[%s] node is nil!", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_BIZ_P2P_NODE_NULL
	}

	ctx, cancel := context.WithTimeout(ctxMain, ctxTimeout_normal)
	defer cancel()

	if node.Network().Connectedness(peer.ID) != network.Connected {
		node.Network().ClosePeer(peer.ID)
		log.Printf("begin  to connect %+v ... \n\n", peer)
		if err := node.Connect(ctx, peer); err != nil {
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
		log.Printf("connect %+v end\n", peer)
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

	return onlineEvent(stream, false, &client)
}

func BuildFmtTypeTalker(id string, fmtType rtkCommon.TransFmtType) rtkMisc.CrossShareErr {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout_short)
	defer cancel()

	sInfo, ok := GetStreamInfo(id)
	if !ok {
		log.Printf("[%s] ID:[%s] IP[%s] get no stream info or stream is closed", rtkMisc.GetFuncInfo(), id, sInfo.ipAddr)
		return rtkMisc.ERR_BIZ_GET_STREAM_EMPTY
	}

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
	} else if fmtType == rtkCommon.IMAGE_CB {
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
		// DEBUG log
		// log.Println("Self is the first login Client , no any other client to build talker")
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
	// FIXME: this cause dead lock
	// mutex := getMutex(id)
	// mutex.Lock()
	// defer mutex.Unlock()
	if len(data) == 0 {
		log.Println("Write faile: empty data")
		return rtkMisc.ERR_BIZ_P2P_WRITE_EMPTY_DATA
	}

	sInfo, ok := GetStreamInfo(id)
	if !ok {
		log.Printf("[%s][%s] WriteSocket err, get no stream or stream is closed", rtkMisc.GetFuncInfo(), sInfo.ipAddr)
		return rtkMisc.ERR_BIZ_GET_STREAM_EMPTY
	}

	s := rtkUtils.NewConnFromStream(sInfo.s)
	if _, err := s.Write(data); err != nil {
		if CheckStreamReset(id, sInfo.timeStamp) {
			return rtkMisc.ERR_BIZ_GET_STREAM_RESET
		}

		log.Printf("[%s][%s] Write failed:[%+v], and execute offlineEvent ", rtkMisc.GetFuncInfo(), sInfo.ipAddr, err)
		offlineEvent(sInfo.s)

		// TODO: check if necessary
		if errors.Is(err, io.EOF) {
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

	bufio.NewWriter(s).Flush()
	log.Printf("[%s] DST ID:[%s] Write to socket successfully", rtkMisc.GetFuncInfo(), id)
	return rtkMisc.SUCCESS
}

func ReadSocket(id string, buffer []byte) (int, rtkMisc.CrossShareErr) {
	sInfo, ok := GetStreamInfo(id)
	if !ok {
		return 0, rtkMisc.ERR_BIZ_GET_STREAM_EMPTY
	}

	s := rtkUtils.NewConnFromStream(sInfo.s)
	n, err := s.Read(buffer)
	if err != nil {
		if CheckStreamReset(id, sInfo.timeStamp) {
			return 0, rtkMisc.ERR_BIZ_GET_STREAM_RESET
		}

		log.Printf("[%s][%s] Read failed [%+v],  execute offlineEvent ", rtkMisc.GetFuncInfo(), sInfo.ipAddr, err)
		offlineEvent(sInfo.s)

		if errors.Is(err, io.EOF) {
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

func onlineEvent(stream network.Stream, isFromListener bool, clientInfo *rtkMisc.ClientInfo) rtkMisc.CrossShareErr {
	id := stream.Conn().RemotePeer().String()
	mutex := getMutex(id)
	mutex.Lock()
	defer mutex.Unlock()

	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)
	var peerDeviceName, peerPlatForm, srcPortType, peerVer string
	if isFromListener {
		resultCode := handleNotice(stream, &peerPlatForm, &peerDeviceName, &srcPortType, &peerVer)
		if resultCode != rtkMisc.SUCCESS {
			stream.Reset()
			log.Printf("[%s] ID:[%s] IP:[%s] errCode:%d, so reset this stream, onlineEvent failed!", id, ipAddr, resultCode)
			return resultCode
		}
	} else {
		resultCode := noticeToPeer(stream)
		if resultCode != rtkMisc.SUCCESS {
			stream.Reset()
			log.Printf("[%s] ID:[%s] IP:[%s] errCode:%d, so reset this stream, onlineEvent failed!", id, ipAddr, resultCode)
			return resultCode
		}
		peerDeviceName = clientInfo.DeviceName
		peerPlatForm = clientInfo.Platform
		srcPortType = clientInfo.SourcePortType
		peerVer = clientInfo.Version
	}

	UpdateStream(id, stream)
	log.Println("****************************************************************************************")
	if isFromListener {
		log.Println("Connected from ID:", id, " IP:", ipAddr)
	} else {
		log.Println("Connected to ID:", id, " IP:", ipAddr)
	}
	log.Println("****************************************************************************************\n\n")

	updateUIOnlineStatus(true, id, ipAddr, peerPlatForm, peerDeviceName, srcPortType, peerVer)
	return rtkMisc.SUCCESS
}

func offlineEvent(stream network.Stream) {
	id := stream.Conn().RemotePeer().String()
	mutex := getMutex(id)
	mutex.Lock()
	defer mutex.Unlock()

	clientInfo, err := rtkUtils.GetClientInfo(id)
	if err == nil {
		updateUIOnlineStatus(false, id, clientInfo.IpAddr, clientInfo.Platform, clientInfo.DeviceName, clientInfo.SourcePortType, "")
	} else {
		log.Printf("[%s] %s, so not need updateUIOnlineStatus!", rtkMisc.GetFuncInfo(), err.Error())
	}

	// Disconnect and remove stream map
	CloseStream(id)
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
	offlineEvent(s)
}

func updateUIOnlineStatus(isOnline bool, id, ipAddr, platfrom, deviceName, srcPortType, ver string) {
	if isOnline {
		log.Printf("[%s] IP:[%s] Online: increase client count", rtkMisc.GetFuncInfo(), ipAddr)
		rtkPlatform.GoUpdateClientStatus(1, ipAddr, id, deviceName, srcPortType) // TODO: Deprecate , and replace with GoUpdateClientStatusEx
		rtkUtils.InsertClientInfoMap(id, ipAddr, platfrom, deviceName, srcPortType, ver)
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

func noticeToPeer(s network.Stream) rtkMisc.CrossShareErr {
	write := bufio.NewWriter(s)
	ipAddr := rtkUtils.GetRemoteAddrFromStream(s)

	registMsg := rtkCommon.RegistMdnsMessage{
		Host:           rtkPlatform.GetHostID(),
		Id:             rtkGlobal.NodeInfo.ID,
		Platform:       rtkGlobal.NodeInfo.Platform,
		DeviceName:     rtkGlobal.NodeInfo.DeviceName,
		SourcePortType: rtkGlobal.NodeInfo.SourcePortType,
		Version:        rtkGlobal.ClientVersion,
	}
	if err := json.NewEncoder(write).Encode(registMsg); err != nil {
		log.Printf("[%s] ID:[%s] Stream json.NewEncoder.Encode err:%+v", rtkMisc.GetFuncInfo(), s.Conn().RemotePeer().String(), err)
		if errors.Is(err, context.DeadlineExceeded) {
			return rtkMisc.ERR_NETWORK_P2P_WRITER_DEADLINE
		} else if errors.Is(err, context.Canceled) {
			return rtkMisc.ERR_NETWORK_P2P_WRITER_CANCELED
		} else if errors.Is(err, network.ErrReset) {
			return rtkMisc.ERR_NETWORK_P2P_WRITER_RESET
		}
		return rtkMisc.ERR_NETWORK_P2P_WRITER
	}
	if err := write.Flush(); err != nil {
		log.Printf("[%s] ID:[%s] Error flushing write buffer: %+v", rtkMisc.GetFuncInfo(), s.Conn().RemotePeer().String(), err)
		return rtkMisc.ERR_NETWORK_P2P_FLUSH
	}
	log.Printf("[%s] IP:[%s]noticeToPeer success!", rtkMisc.GetFuncInfo(), ipAddr)
	return rtkMisc.SUCCESS
}

func handleNotice(s network.Stream, platForm, name, srcPortType, ver *string) rtkMisc.CrossShareErr {
	read := bufio.NewReader(s)
	ipAddr := rtkUtils.GetRemoteAddrFromStream(s)

	var regMsg rtkCommon.RegistMdnsMessage
	err := json.NewDecoder(read).Decode(&regMsg)
	if err != nil {
		log.Printf("[%s] ID:[%s] Stream json.NewDecoder.Decode err:%+v", rtkMisc.GetFuncInfo(), s.Conn().RemotePeer().String(), err)
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
	*ver = regMsg.Version

	log.Printf("[%s] IP:[%s] handleNotice success!", rtkMisc.GetFuncInfo(), ipAddr)
	return rtkMisc.SUCCESS
}

func SetMDNSPeer(peer peer.AddrInfo) {
	mdnsPeerChan <- peer
}
