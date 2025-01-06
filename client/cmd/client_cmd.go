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
	rtkConnection "rtk-cross-share/connection"
	rtkPlatform "rtk-cross-share/platform"
	rtkRelay "rtk-cross-share/relay"
	rtkUtils "rtk-cross-share/utils"
	"time"

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
		"/ip4/%s/tcp/%d",
		"/ip4/%s/udp/%d/quic",
		//"/ip6/::/tcp/%d",
		//"/ip6/::/udp/%d/quic",
	}

	for i, a := range addrs {
		addrs[i] = fmt.Sprintf(a, rtkGlobal.DefaultIp, port)
	}

	return addrs
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

	SetupSettings()
	// SetupLogFileSetting()

	rtkUtils.GoSafe(func() { rtkDebug.DebugCmdLine() })

	ctx := context.Background()

	// Init connection
	rtkConnection.ConnInit(rtkGlobal.DefaultIp)
	rtkUtils.GoSafe(func() { rtkConnection.Run(ctx) })

	// Main loop
	// TODO: refine after MDNS
	for id := range rtkUtils.GetDeviceInfoMap() {
		rtkUtils.GoSafe(func() { rtkPeer2Peer.ProcessEventsForPeer(id, ctx) })
	}

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
	// Init connection
	rtkConnection.ConnInit(listenHost)
	rtkUtils.GoSafe(func() { rtkConnection.Run(ctx) })

	// Main loop
	// TODO: refine after MDNS
	for id := range rtkUtils.GetDeviceInfoMap() {
		rtkUtils.GoSafe(func() { rtkPeer2Peer.ProcessEventsForPeer(id, ctx) })
	}

	rtkUtils.GoSafe(func() { rtkDebug.DebugCmdLine() })
	select {}
}
