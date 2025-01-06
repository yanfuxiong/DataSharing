package connection

import (
	"context"
	"errors"
	"fmt"
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

const (
	retryConnection = 2
	retryDelay      = 1 * time.Second
)

var (
	node host.Host
	// mutexMap by ID
	mutexMap sync.Map
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

func ConnInit(ip string) {
	rtkUtils.InitDeviceInfo(rtkPlatform.GetDeviceInfoPath())
	node = setupNode(ip, rtkUtils.DeviceStaticPort)
	// Remove self ID from map
	delete(rtkUtils.GetDeviceInfoMap(), node.ID().String())
}

func updateSystemInfo() {
	// Update system info
	getLocalIpPort := func(ip string) string {
		return ip + ":" + rtkUtils.DeviceStaticPort
	}
	serviceVer := "v" + rtkBuildConfig.Version + " (" + rtkBuildConfig.BuildDate + ")"
	lastIp := rtkGlobal.DefaultIp
	rtkPlatform.GoUpdateSystemInfo(getLocalIpPort(lastIp), serviceVer)

	rtkUtils.GoSafe(func() {
		for {
			ip, _ := getNetworkInfo()
			isChanged := ip != lastIp
			if isChanged {
				lastIp = ip
				rtkPlatform.GoUpdateSystemInfo(getLocalIpPort(lastIp), serviceVer)
			}

			time.Sleep(5 * time.Second)
		}
	})
}

func Run(ctx context.Context) {
	// TODO: Replace with GetClientList
	rtkPlatform.SetGoPipeConnectedCallback(func() {
		// Update system info
		updateSystemInfo()

		// Update all clients status
		for _, info := range rtkGlobal.MdnsClientList {
			deviceInfo, err := rtkUtils.GetDeviceInfo(info.ID)
			if err != nil {
				continue
			}

			deviceName := deviceInfo.Name
			if deviceName == "" {
				deviceName = info.IpAddr
			}
			rtkPlatform.GoUpdateClientStatus(1, info.IpAddr, info.ID, deviceName)
		}
	})

	buildListener()

	log.Println("Connection rule: (self ID > peer ID)")
	var wg sync.WaitGroup
	for id := range rtkUtils.GetDeviceInfoMap() {
		wg.Add(1)

		rtkUtils.GoSafe(func() {
			defer wg.Done()
			peerId, errDecode := peer.Decode(id)
			if errDecode != nil {
				log.Printf("[%s %d] Err: decode ID failed: %s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), errDecode.Error())
				return
			}
			// Make one listener and one talker
			if node.ID() > peerId {
				_, ok := GetStream(id)
				if !ok {
					err := buildTalker(ctx, id)
					if err != nil {
						// FIXME: Debug log
						// log.Printf("[%s %d] Err: buildTalker failed. id:%s, err:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), id, err.Error())
					}
				}
			}
		})

	}
	wg.Wait()
	log.Println("Change Connection rule: (self ID < peer ID)")
	for id := range rtkUtils.GetDeviceInfoMap() {
		wg.Add(1)
		rtkUtils.GoSafe(func() {
			defer wg.Done()
			peerId, errDecode := peer.Decode(id)
			if errDecode != nil {
				log.Printf("[%s %d] Err: decode ID failed: %s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), errDecode.Error())
				return
			}
			// Make one listener and one talker
			if node.ID() < peerId {
				_, ok := GetStream(id)
				if !ok {
					err := buildTalker(ctx, id)
					if err != nil {
						log.Printf("[%s %d] Err: buildTalker failed. id:%s, err:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), id, err.Error())
					}
				}
			}
		})
	}
	wg.Wait()

	select {}
}

func setupNode(ip, port string) host.Host {
	priv := rtkPlatform.GenKey()
	// rtkUtils.InitDeviceTable(rtkPlatform.GetDeviceTablePath())

	sourceAddrStr := fmt.Sprintf("/ip4/%s/tcp/%s", ip, port)
	sourceMultiAddr, _ := ma.NewMultiaddr(sourceAddrStr)

	node, err := libp2p.New(
		//libp2p.ListenAddrStrings(listen_addrs(rtkMdns.MdnsCfg.ListenPort)...), // Add mdns port with different initialization
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.NATPortMap(),
		libp2p.Identity(priv),
		libp2p.ForceReachabilityPrivate(),
		libp2p.ResourceManager(&network.NullResourceManager{}),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Printf("Failed to create node: %v", err)
		return nil
	}

	node.Network().Listen(node.Addrs()[0])

	log.Println("Self ID: ", node.ID().String())
	log.Println("Self node Addr: ", node.Addrs())
	log.Printf("========================\n\n")

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

	return node
}

func buildListener() {
	log.Println("BuildListener")
	node.SetStreamHandler(rtkGlobal.ProtocolDirectID, func(stream network.Stream) {
		onlineEvent(stream, true)
	})
}

func buildTalker(ctx context.Context, id string) error {
	mutex := getMutex(id)
	mutex.Lock()
	defer mutex.Unlock()

	deviceInfo, err := rtkUtils.GetDeviceInfo(id)
	if err != nil {
		return err
	}

	// Use static port temporarily
	ipAddr := fmt.Sprintf("/ip4/%s/tcp/%s", deviceInfo.IP, rtkUtils.DeviceStaticPort)
	addr := ma.StringCast(ipAddr)
	idB85, err := peer.Decode(id)
	if err != nil {
		log.Printf("[%s %d] ID decode failed: %s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), id)
		return err
	}
	peer := peer.AddrInfo{
		ID:    idB85,
		Addrs: []ma.Multiaddr{addr},
	}

	connCount := 0
	var finalErr error = nil
	for connCount <= retryConnection {
		if connCount > 0 {
			log.Printf("Connect to %s, retry count:%d", ipAddr, connCount)
		}
		connCount++

		if err := node.Connect(ctx, peer); err != nil {
			// FIXME: Debug log
			// log.Println("Connection failed:", err)
			finalErr = err
			time.Sleep(retryDelay)
			continue
		}

		stream, err := node.NewStream(ctx, peer.ID, protocol.ID(rtkGlobal.ProtocolDirectID))
		if err != nil {
			log.Println("Stream open failed", err)
			finalErr = err
			time.Sleep(retryDelay)
			continue
		}

		onlineEvent(stream, false)
		return nil
	}

	return finalErr
}

func WriteSocket(id string, data []byte) rtkCommon.SocketErr {
	ipAddr, err := rtkUtils.GetDeviceIp(id)
	if err != nil {
		log.Printf("[Socket] Err: not found node")
		return rtkCommon.ERR_OTHER
	}

	if len(data) == 0 {
		log.Println("Write faile: empty data")
		return rtkCommon.ERR_OTHER
	}

	ctx := context.Background()
	_, ok := GetStream(id)
	if !ok {
		if err := buildTalker(ctx, id); err != nil {
			return rtkCommon.ERR_CONNECTION
		}
	}

	executeWriteData := func(data []byte) rtkCommon.SocketErr {
		// FIXME: this cause dead lock
		// mutex := getMutex(id)
		// mutex.Lock()
		// defer mutex.Unlock()

		stream, ok := GetStream(id)
		if !ok {
			return rtkCommon.ERR_OTHER
		}

		s := rtkUtils.NewConnFromStream(stream)
		if _, err := s.Write(data); err != nil {
			// Write failed and delete from stream map
			// TODO: check if necessary
			log.Printf("[%s %d] Write failed and close stream: %s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), ipAddr)
			offlineEvent(stream)

			if errors.Is(err, io.EOF) {
				log.Println("Write faile: Stream closed by peer")
				return rtkCommon.ERR_EOF
			} else if netErr, ok := err.(net.Error); ok {
				log.Println("Write fail network error", netErr.Error())
				return rtkCommon.ERR_NETWORK
			} else {
				log.Println("Write fail", err.Error())
				return rtkCommon.ERR_OTHER
			}
		}
		log.Printf("[%s %d] Write to socket successfully", rtkUtils.GetFuncName(), rtkUtils.GetLine())
		return rtkCommon.OK
	}

	rtkErr := executeWriteData(data)
	if rtkErr != rtkCommon.OK {
		log.Printf("[%s %d] Lost connection. Retry connect to %s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), ipAddr)
		if err := buildTalker(ctx, id); err != nil {
			return rtkCommon.ERR_CONNECTION
		}

		// Send again after connected
		return executeWriteData(data)
	}

	return rtkErr
}

func ReadSocket(id string, buffer []byte) (int, rtkCommon.SocketErr) {
	ipAddr, err := rtkUtils.GetDeviceIp(id)
	if err != nil {
		log.Printf("[Socket] Err: not found node")
		return 0, rtkCommon.ERR_OTHER
	}

	mutex := getMutex(id)
	mutex.Lock()
	defer mutex.Unlock()

	stream, ok := GetStream(id)
	if !ok {
		return 0, rtkCommon.ERR_CONNECTION
	}
	s := rtkUtils.NewConnFromStream(stream)

	executeReadData := func(buffer []byte) (int, rtkCommon.SocketErr) {
		n, err := s.Read(buffer)
		if err != nil {
			// Write failed and delete from stream map
			// TODO: check if necessary
			log.Printf("[%s %d] read failed and close stream: %s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), ipAddr)
			offlineEvent(stream)

			if errors.Is(err, io.EOF) {
				log.Println("Write faile: Stream closed by peer")
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

	return executeReadData(buffer)
}

func updatePublicInfo(ip, port string) {
	rtkGlobal.NodeInfo.IPAddr.PublicIP = ip
	rtkGlobal.NodeInfo.IPAddr.PublicPort = port
}

func onlineEvent(stream network.Stream, isFromListener bool) {
	ip, port := rtkUtils.ExtractTCPIPandPort(stream.Conn().LocalMultiaddr())
	updatePublicInfo(ip, port)

	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)
	id := stream.Conn().RemotePeer().String()
	log.Println("************************************************")
	if isFromListener {
		log.Println("Connected from ID:", id, " IP:", ipAddr)
	} else {
		log.Println("Connected to ID:", id, " IP:", ipAddr)
	}
	log.Println("************************************************")

	// TODO: consider reconnection case
	// var peerDeviceName string
	// handleRegister(stream, &peerDeviceName)
	// registerToPeer(stream, &peerDeviceName)
	// TODO: refine after MDNS
	deviceName := ipAddr
	deviceInfo, err := rtkUtils.GetDeviceInfo(id)
	if err == nil && deviceInfo.Name != "" {
		deviceName = deviceInfo.Name
	}
	UpdateUIOnlineStatus(true, id, ipAddr, deviceName)
	// Connect and update stream map
	_, ok := GetStream(id)
	if ok {
		log.Printf("[%s %d] Stream existed. DO NOT update stream", rtkUtils.GetFuncName(), rtkUtils.GetLine())
	} else {
		UpdateStream(id, stream)
	}
}

func offlineEvent(stream network.Stream) {
	rtkUtils.LostMdnsClientList(stream.Conn().RemotePeer().String())
	rtkPlatform.FoundPeer()

	ipAddr := rtkUtils.GetRemoteAddrFromStream(stream)
	id := stream.Conn().RemotePeer().String()
	log.Println("************************************************")
	log.Println("Lost connection with ID:", id, " IP:", ipAddr)
	log.Println("************************************************")

	// TODO: refine after MDNS
	deviceName := ipAddr
	deviceInfo, err := rtkUtils.GetDeviceInfo(id)
	if err == nil && deviceInfo.Name != "" {
		deviceName = deviceInfo.Name
	}
	UpdateUIOnlineStatus(false, id, ipAddr, deviceName)
	// Disconnect and update stream map
	CloseStream(id)
}

// TODO: refine with GetClientStatus
func UpdateUIOnlineStatus(isOnline bool, id, ipAddr, deviceName string) {
	_, ok := GetStream(id)
	if isOnline {
		if !ok {
			log.Printf("[%s %d] Online: increase client count", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			rtkPlatform.GoUpdateClientStatus(1, ipAddr, id, deviceName)
			rtkUtils.InsertMdnsClientList(id, ipAddr, rtkPlatform.GetPlatform(), deviceName)
			rtkPlatform.FoundPeer()
		} else {
			log.Printf("[%s %d] Online: existed, skip", rtkUtils.GetFuncName(), rtkUtils.GetLine())
		}
	} else {
		if ok {
			log.Printf("[%s %d] Offline: decrease client count", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			rtkPlatform.GoUpdateClientStatus(0, ipAddr, id, deviceName)
			rtkUtils.LostMdnsClientList(id)
			rtkPlatform.FoundPeer()
		} else {
			log.Printf("[%s %d] Offline: not existed, skip", rtkUtils.GetFuncName(), rtkUtils.GetLine())
		}
	}
}
