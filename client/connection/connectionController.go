package connection

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"io"
	"log"
	"net"
	rtkBuildConfig "rtk-cross-share/buildConfig"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
)

func getMutex(id string) *sync.Mutex {
	value, ok := mutexMap.Load(id)
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

func getNetworkInfo() (string, bool) {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Failed to get network interfaces: %v", err)
		return rtkGlobal.DefaultIp, false
	}
	for _, iface := range interfaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Err: Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					return ipNet.IP.String(), true
				}
			}
		}
	}
	return rtkGlobal.DefaultIp, false
}

func ConnectionInit() {
	log.Printf("[%s] listen host[%s] port[%d] connection init", rtkUtils.GetFuncInfo(), rtkGlobal.ListenHost, rtkGlobal.ListenPort)
	node = setupNode(rtkGlobal.ListenHost, rtkGlobal.ListenPort)
	pingServer = ping.NewPingService(node)
}

func updateSystemInfo() {
	// Update system info
	ipAddr := rtkUtils.ConcatIP(rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort)
	serviceVer := "v" + rtkBuildConfig.Version + " (" + rtkBuildConfig.BuildDate + ")"
	//lastIp := rtkGlobal.DefaultIp
	rtkPlatform.GoUpdateSystemInfo(ipAddr, serviceVer)
	// TODO:  fix this by TSTAS-40
	/*rtkUtils.GoSafe(func() {
		for {
			ip, _ := getNetworkInfo()
			isChanged := ip != lastIp
			if isChanged {
				lastIp = ip
				rtkPlatform.GoUpdateSystemInfo(getLocalIpPort(lastIp), serviceVer)
			}

			time.Sleep(5 * time.Second)
		}
	})*/
}

func cancelHostNode() {
	if node != nil {
		node.Peerstore().Close()
		node.Network().Close()
		node.ConnManager().Close()
		node.Close()
		node = nil
	}
}

func Run(ctx context.Context) {
	defer cancelHostNode()
	defer CancelStreamPool()

	rtkPlatform.SetGoPipeConnectedCallback(func() {
		// Update system info
		updateSystemInfo()

		// Update all clients status
		clientMap := rtkUtils.GetClientMap()
		for _, info := range clientMap {
			rtkPlatform.GoUpdateClientStatus(1, info.IpAddr, info.ID, info.DeviceName)
		}
	})

	buildListener()

	rtkUtils.GoSafe(func() { BuildMDNSTalker(ctx) })

	for {
		select {
		case <-ctx.Done():
			log.Println("connectionController run is end by main context")
			return
		case <-time.After(30 * time.Second):
			CheckAllStreamAlive(ctx)
		case peerInfo := <-reConnectPeerChan:
			peerInfo.RetryCount++
			if peerInfo.RetryCount > peerInfo.MaxCount {
				if !IsStreamExisted(peerInfo.Peer.ID.String()) {
					OfflineStream(peerInfo.Peer.ID.String())
				}
				continue
			}

			if node.Network().Connectedness(peerInfo.Peer.ID) != network.Connected {
				log.Printf("Start connect to %+v,  %d times to retry, close peer first ", peerInfo.Peer.Addrs, peerInfo.RetryCount)
				ClosePeer(peerInfo.Peer.ID.String())
				if err := node.Connect(ctx, peerInfo.Peer); err != nil {
					log.Printf("Connect to peer %+v failed:%+v", peerInfo.Peer.Addrs, err)
					time.Sleep(peerInfo.DelayTime)
					reConnectPeerChan <- peerInfo
					continue
				}
			} else {
				log.Printf("Start open a new stream to %+v,  %d times to retry ...", peerInfo.Peer.Addrs, peerInfo.RetryCount)
			}

			if IsStreamExisted(peerInfo.Peer.ID.String()) {
				log.Printf("[%s] %+v a new stream is Opened , so skip it", rtkUtils.GetFuncInfo(), peerInfo.Peer.Addrs)
				continue
			}

			stream, err := node.NewStream(ctx, peerInfo.Peer.ID, protocol.ID(rtkGlobal.ProtocolDirectID))
			if err != nil {
				log.Println("Stream open failed: ", err)
				time.Sleep(peerInfo.DelayTime)
				reConnectPeerChan <- peerInfo
				continue
			}

			onlineEvent(stream, false)
		}
	}
}

func setupNode(ip string, port int) host.Host {
	priv := rtkPlatform.GenKey()

	sourceAddrStr := fmt.Sprintf("/ip4/%s/tcp/%d", ip, port)
	sourceMultiAddr, err := ma.NewMultiaddr(sourceAddrStr)
	if err != nil {
		log.Printf("NewMultiaddr error:%+v, with addr:%s", err, sourceAddrStr)
		panic(err)
	}
	
	if port <= 0 {
		log.Println("(MDNS) listen port is not set. Use a random port")
	}

	node, err := libp2p.New(
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
		panic(err)
	}

	node.Network().Listen(node.Addrs()...)

	log.Println("=======================================================")
	log.Println("Self ID: ", node.ID().String())
	log.Println("Self node Addr: ", node.Addrs())
	log.Println("Self listen Addr: ", node.Network().ListenAddresses())
	log.Println("=======================================================\n\n")

	for _, p := range node.Peerstore().Peers() {
		node.Peerstore().ClearAddrs(p)
	}

	if rtkPlatform.IsHost() {
		rtkUtils.WriteNodeID(node.ID().String(), rtkPlatform.GetHostIDPath())
	}

	rtkUtils.WriteNodeID(node.ID().String(), rtkPlatform.GetIDPath())

	rtkGlobal.NodeInfo.IPAddr.LocalPort = rtkUtils.GetLocalPort(node.Addrs())
	rtkGlobal.NodeInfo.ID = node.ID().String()
	rtkGlobal.NodeInfo.DeviceName = rtkUtils.ConcatIP(rtkUtils.ExtractTCPIPandPort(node.Addrs()[0]))
	rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort = rtkUtils.ExtractTCPIPandPort(node.Addrs()[0])

	log.Printf("Public IP[%s], Public port[%s], LocalPort[%s]", rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort, rtkGlobal.NodeInfo.IPAddr.LocalPort)
	return node
}

// TODO : refine this flow: by TSTAS-61
func getDelayTime(id peer.ID) time.Duration {
	if node.ID() > id {
		return retryDelay
	}
	return retryDelay + 100*time.Millisecond
}

func buildListener() {
	log.Println("BuildListener")
	node.SetStreamHandler(rtkGlobal.ProtocolDirectID, func(stream network.Stream) {
		onlineEvent(stream, true)
	})
}

func buildTalker(ctx context.Context, peer peer.AddrInfo) error {
	if node.Network().Connectedness(peer.ID) == network.Connected {
		log.Printf("[%s] ID:[%s] Connected is already existed, skip connect.", rtkUtils.GetFuncInfo(), peer.ID.String())
		return nil
	}

	log.Printf("begin  to connect %+v \n\n", peer)
	if err := node.Connect(ctx, peer); err != nil {
		log.Printf("Connect peer%+v failed:%+v", peer, err)
		return err
	}
	log.Printf("connect %+v end\n", peer)

	if IsStreamExisted(peer.ID.String()) {
		log.Printf("[%s] ID:[%s] a Stream is already existed, skip NewStream.", rtkUtils.GetFuncInfo(), peer.ID.String())
		return nil
	}

	stream, err := node.NewStream(ctx, peer.ID, protocol.ID(rtkGlobal.ProtocolDirectID))
	if err != nil {
		log.Println("Stream open failed", err)
		return err
	}

	onlineEvent(stream, false)
	return nil

}

func WriteSocket(id string, data []byte) rtkCommon.SocketErr {
	// FIXME: this cause dead lock
	// mutex := getMutex(id)
	// mutex.Lock()
	// defer mutex.Unlock()
	if len(data) == 0 {
		log.Println("Write faile: empty data")
		return rtkCommon.ERR_OTHER
	}

	ipAddr := GetStreamIpAddr(id)
	stream, ok := GetStream(id)
	if !ok {
		log.Printf("[%s][%s] WriteSocket err, get no stream or stream is closed", rtkUtils.GetFuncInfo(), ipAddr)
		return rtkCommon.ERR_OTHER
	}

	s := rtkUtils.NewConnFromStream(stream)
	if _, err := s.Write(data); err != nil {
		log.Printf("[%s][%s] Write fail and execute offlineEvent ", rtkUtils.GetFuncInfo(), ipAddr)
		offlineEvent(stream)

		// TODO: check if necessary
		if errors.Is(err, io.EOF) {
			log.Println("Write fail: Stream closed by peer")
			return rtkCommon.ERR_EOF
		} else if netErr, ok := err.(net.Error); ok {
			log.Println("Write fail network error", netErr.Error())
			return rtkCommon.ERR_NETWORK
		} else {
			log.Println("Write fail", err.Error())
			return rtkCommon.ERR_OTHER
		}
	}
	log.Printf("[%s] Write to socket successfully", rtkUtils.GetFuncInfo())
	return rtkCommon.OK
}

func ReadSocket(id string, buffer []byte) (int, rtkCommon.SocketErr) {
	ipAddr := GetStreamIpAddr(id)
	stream, ok := GetStream(id)
	if !ok {
		return 0, rtkCommon.ERR_CONNECTION
	}
	s := rtkUtils.NewConnFromStream(stream)
	n, err := s.Read(buffer)
	if err != nil {
		log.Printf("[%s][%s] read failed and execute offlineEvent ", rtkUtils.GetFuncInfo(), ipAddr)
		offlineEvent(stream)

		// TODO: check if necessary
		if errors.Is(err, io.EOF) {
			log.Println("Read fail: Stream closed by peer")
			return n, rtkCommon.ERR_EOF
		} else if netErr, ok := err.(net.Error); ok {
			log.Printf("[Socket][%s] Err: Read fail network error(%v)", ipAddr, netErr.Error())
			if netErr.Timeout() {
				return n, rtkCommon.ERR_TIMEOUT
			}
			return n, rtkCommon.ERR_NETWORK
		} else {
			log.Printf("[Socket][%s] Err: Read fail(%v)", ipAddr, err.Error())
			return n, rtkCommon.ERR_OTHER
		}
	}

	return n, rtkCommon.OK
}

func onlineEvent(stream network.Stream, isFromListener bool) {
	id := stream.Conn().RemotePeer().String()
	UpdateStream(id, stream)

	mutex := getMutex(id)
	mutex.Lock() // TODO: refine this flow: cause dead lock both ends   or cause two connect

	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)
	var peerDeviceName, peerPlatForm string
	if isFromListener {
		handleRegister(stream, &peerPlatForm, &peerDeviceName)
	} else {
		registerToPeer(stream, &peerPlatForm, &peerDeviceName)
	}

	log.Println("************************************************")
	if isFromListener {
		log.Println("Connected from ID:", id, " IP:", ipAddr)
	} else {
		log.Println("Connected to ID:", id, " IP:", ipAddr)
	}
	log.Println("************************************************")

	if peerDeviceName == "" {
		peerDeviceName = ipAddr
	}

	updateUIOnlineStatus(true, id, ipAddr, peerPlatForm, peerDeviceName)
	StartProcessChan <- id
	mutex.Unlock()
}

func offlineEvent(stream network.Stream) {
	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)
	id := stream.Conn().RemotePeer().String()
	mutex := getMutex(id)
	mutex.Lock()
	defer mutex.Unlock()

	log.Println("************************************************")
	log.Println("Lost connection with ID:", id, " IP:", ipAddr)
	log.Println("************************************************")

	clientInfo, _ := rtkUtils.GetClientInfo(id)
	updateUIOnlineStatus(false, id, ipAddr, "", clientInfo.DeviceName)
	// Disconnect and update stream map
	CloseStream(id)
	EndProcessChan <- id
}

func updateUIOnlineStatus(isOnline bool, id, ipAddr, platfrom, deviceName string) {
	if isOnline {
		log.Printf("[%s ] IP:[%s] Online: increase client count", rtkUtils.GetFuncInfo(), ipAddr)
		rtkPlatform.GoUpdateClientStatus(1, ipAddr, id, deviceName)
		rtkUtils.InsertClientInfoMap(id, ipAddr, platfrom, deviceName)
		rtkPlatform.FoundPeer()
	} else {
		log.Printf("[%s] IP:[%s] Offline: decrease client count", rtkUtils.GetFuncInfo(), ipAddr)
		rtkPlatform.GoUpdateClientStatus(0, ipAddr, id, deviceName)
		rtkUtils.LostClientInfoMap(id)
		rtkPlatform.FoundPeer()
	}
}

func registerToPeer(s network.Stream, platForm, name *string) error {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	ipAddr := rtkUtils.GetRemoteAddrFromStream(s)

	registMsg := rtkCommon.RegistMdnsMessage{
		Host:       rtkPlatform.GetHostID(),
		Id:         rtkGlobal.NodeInfo.ID,
		Platform:   rtkPlatform.GetPlatform(),
		DeviceName: rtkGlobal.NodeInfo.DeviceName,
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

	*platForm = registMsg.Platform
	*name = regResonseMsg.DeviceName
	log.Printf("[%s] IP:[%s]registerToPeer success!", rtkUtils.GetFuncInfo(), ipAddr)
	return nil
}

func handleRegister(s network.Stream, platForm, name *string) error {
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

	*platForm = regMsg.Platform
	*name = regMsg.DeviceName

	registMsg := rtkCommon.RegistMdnsMessage{
		Host:       rtkPlatform.GetHostID(),
		Id:         rtkGlobal.NodeInfo.ID,
		Platform:   rtkPlatform.GetPlatform(),
		DeviceName: rtkGlobal.NodeInfo.DeviceName,
	}

	if err := json.NewEncoder(rw).Encode(&registMsg); err != nil {
		fmt.Println("failed to read register response message: ", err)
		return err
	}
	if err := rw.Flush(); err != nil {
		fmt.Println("Error flushing write buffer: ", err)
		return err
	}
	log.Printf("[%s] IP:[%s] handleRegister success!", rtkUtils.GetFuncInfo(), ipAddr)
	return nil
}

func SetMDNSPeer(peer peer.AddrInfo) {
	mdnsPeerChan <- peer
}

func BuildMDNSTalker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("BuildMDNSTalker is end by main context")
			return
		case mdnsPeer := <-mdnsPeerChan:
			if IsStreamExisted(mdnsPeer.ID.String()) {
				log.Printf("[MDNS] [%s]  Stream already existed, skip connect. ", mdnsPeer.ID.String())
				continue
			}

			err := buildTalker(ctx, mdnsPeer)
			if err != nil {
				reConnectInfo := ReConnectPeerInfo{
					Peer:       mdnsPeer,
					RetryCount: 0,
					MaxCount:   retryConnection,
					DelayTime:  getDelayTime(mdnsPeer.ID),
				}

				reConnectPeerChan <- reConnectInfo
			}
		}

	}
}
