package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	rtkBuildConfig "rtk-cross-share/buildConfig"
	rtkClipboard "rtk-cross-share/clipboard"
	rtkDebug "rtk-cross-share/debug"
	rtkFileDrop "rtk-cross-share/filedrop"
	rtkGlobal "rtk-cross-share/global"
	rtkMdns "rtk-cross-share/mdns"
	rtkPlatform "rtk-cross-share/platform"
	rtkRelay "rtk-cross-share/relay"
	rtkUtils "rtk-cross-share/utils"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	"golang.design/x/clipboard"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupSettings() {
	rtkPlatform.SetupCallbackSettings()
	rtkClipboard.InitClipboard()
	rtkFileDrop.InitFileDrop()

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	rtkMdns.MdnsCfg = rtkMdns.ParseFlags()
}

func SetupLogFileSetting() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   rtkPlatform.GetLogFilePath(),
		MaxSize:    256,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	})
}

var sourceAddrStr string

func listen_addrs(port int) []string {
	addrs := []string{
		"/ip4/0.0.0.0/tcp/%d",
		"/ip4/0.0.0.0/udp/%d/quic",
		//"/ip6/::/tcp/%d",
		//"/ip6/::/udp/%d/quic",
	}

	for i, a := range addrs {
		addrs[i] = fmt.Sprintf(a, port)
	}

	return addrs
}

func SetupNode() host.Host {
	priv := rtkPlatform.GenKey()
	rtkUtils.InitDeviceTable(rtkPlatform.GetDeviceTablePath())

	if rtkMdns.MdnsCfg.ListenPort <= 0 {
		log.Println("(MDNS) listen port is not set. Use a random port")
		/*if content := rtkUtils.ReadMdnsPort(rtkPlatform.GetMdnsPortConfigPath()); content != "" {
			rtkMdns.MdnsCfg.ListenPort, _ = strconv.Atoi(content)
		} else {
			rand.Seed(time.Now().UnixNano())
			rtkMdns.MdnsCfg.ListenPort = rand.Intn(65535)
			rtkUtils.WriteMdnsPort(string(rtkMdns.MdnsCfg.ListenPort), rtkPlatform.GetMdnsPortConfigPath())
		}*/
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(sourceAddrStr)
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
	log.Println("Self listen Addr: ", node.Network().ListenAddresses())
	log.Println("Self LocalPort:", rtkUtils.GetLocalPort(node.Addrs()))
	log.Println("========================\n\n")

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

func Run() {
	lockFilePath := "singleton.lock"
	file, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Failed to open or create lock file:", err)
		return
	}
	defer file.Close()

	err = rtkPlatform.LockFile(file)
	if err != nil {
		fmt.Println("Another instance is already running.")
		return
	}
	defer rtkPlatform.UnlockFile(file)

	// TODO: Replace with GetClientList
	rtkPlatform.SetGoPipeConnectedCallback(func() {
		for _, info := range rtkGlobal.MdnsClientList {
			deviceName := rtkUtils.QueryDeviceName(info.ID)
			if deviceName == "" {
				deviceName = info.IpAddr
			}
			rtkPlatform.GoUpdateClientStatus(1, info.IpAddr, info.ID, deviceName)
		}
	})
	SetupSettings()
	if rtkMdns.MdnsCfg.LogSwitch == 0 {
		SetupLogFileSetting()
	}

	sourceAddrStr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", rtkMdns.MdnsCfg.ListenPort)
	ctx := context.Background()
	node := SetupNode()

	rtkRelay.BuildListener(ctx, node)

	rtkMdns.BuildMdnsListener(node)
	rtkMdns.BuildMdnsTalker(ctx, node)

	<-time.After(3 * time.Second)        // wait mdns discovery all peers
	rtkUtils.RemoveMdnsClientFromGuest() //delete mdns found peer in global list

	if len(rtkGlobal.GuestList) == 0 {
		log.Println("Wait for node")
		rtkUtils.GoSafe(func() { rtkDebug.DebugCmdLine() })
		select {}
	}

	for _, target := range rtkGlobal.GuestList {
		rtkRelay.BuildTalker(ctx, node, target)
	}

	rtkUtils.GoSafe(func() { rtkDebug.DebugCmdLine() })
	select {}
}

func MainInit(serverId, serverIpInfo, listenHost string, listentPort int) {
	log.Println("========================")
	log.Println("Version: ", rtkBuildConfig.Version)
	log.Println("Build Date: ", rtkBuildConfig.BuildDate)
	log.Printf("========================\n\n")

	SetupSettings()
	if len(serverId) > 0 && len(serverIpInfo) > 0 && listenHost != "" && listentPort > 0 {
		rtkGlobal.RelayServerID = serverId
		rtkGlobal.RelayServerIPInfo = serverIpInfo
		rtkMdns.MdnsCfg.ListenPort = listentPort
		rtkMdns.MdnsCfg.ListenHost = listenHost
		log.Printf("set relayServerID: %s", serverId)
		log.Printf("set relayServerIPInfo: %s", serverIpInfo)
		log.Printf("(MDNS) set host[%s] listen port: %d\n", listenHost, listentPort)
		sourceAddrStr = fmt.Sprintf("/ip4/%s/tcp/%d", rtkMdns.MdnsCfg.ListenHost, rtkMdns.MdnsCfg.ListenPort)
	}

	//SetupLogFileSetting()

	ctx := context.Background()
	node := SetupNode()

	log.Println("Self ID: ", node.ID())
	log.Printf("========================\n\n")

	rtkRelay.BuildListener(ctx, node)

	rtkMdns.BuildMdnsListener(node)
	rtkMdns.BuildMdnsTalker(ctx, node)

	<-time.After(3 * time.Second) // wait mdns discovery all peers
	if len(rtkGlobal.GuestList) == 0 {
		log.Println("Wait for node")
		rtkUtils.GoSafe(func() { rtkDebug.DebugCmdLine() })
		select {}
	}

	rtkUtils.RemoveMdnsClientFromGuest() //delete mdns found peer
	for _, targetId := range rtkGlobal.GuestList {
		rtkRelay.BuildTalker(ctx, node, targetId)
	}

	rtkUtils.GoSafe(func() { rtkDebug.DebugCmdLine() })
	select {}
}
