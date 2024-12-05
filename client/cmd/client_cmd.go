package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	rtkBuildConfig "rtk-cross-share/buildConfig"
	rtkDebug "rtk-cross-share/debug"
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

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	flag.Parse()
}

func SetupLogFileSetting() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "p2p.log",
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

	return node
}

func Run() {
	rtkMdns.MdnsCfg = rtkMdns.ParseFlags()
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

func MainInit(cb rtkPlatform.Callback, serverId, serverIpInfo, listenHost string, listentPort int) {
	log.Println("========================")
	log.Println("Version: ", rtkBuildConfig.Version)
	log.Println("Build Date: ", rtkBuildConfig.BuildDate)
	log.Printf("========================\n\n")

	rtkPlatform.CallbackInstance = cb

	rtkMdns.MdnsCfg = &rtkMdns.Config{}
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

	SetupSettings()
	//SetupLogFileSetting()

	ctx := context.Background()
	node := SetupNode()

	fmt.Println("Self ID: ", node.ID())
	fmt.Printf("========================\n\n")

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
