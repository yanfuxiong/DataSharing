package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	rtkBuildConfig "rtk-cross-share/buildConfig"
	rtkClipboard "rtk-cross-share/clipboard"
	rtkConnection "rtk-cross-share/connection"
	rtkDebug "rtk-cross-share/debug"
	rtkFileDrop "rtk-cross-share/filedrop"
	rtkGlobal "rtk-cross-share/global"
	rtkMdns "rtk-cross-share/mdns"
	rtkPeer2Peer "rtk-cross-share/peer2peer"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"time"

	"golang.design/x/clipboard"
	"gopkg.in/natefinch/lumberjack.v2"
)

func setupSettings() {
	rtkPlatform.SetupCallbackSettings()
	rtkClipboard.InitClipboard()
	rtkFileDrop.InitFileDrop()

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func setupLogFileSetting() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   rtkPlatform.GetLogFilePath(),
		MaxSize:    256,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	})
}

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

	setupSettings()
	// setupLogFileSetting()

	rtkGlobal.ListenHost = rtkGlobal.DefaultIp
	rtkGlobal.ListenPort = 0

	rtkUtils.GoSafe(func() { rtkDebug.DebugCmdLine() })

	rtkUtils.GoSafe(func() { BusinessProcess() })

	select {}
}

func MainInit(serverId, serverIpInfo, listenHost string, listentPort int) {
	log.Println("=======================================================")
	log.Println("Version: ", rtkBuildConfig.Version)
	log.Println("Build Date: ", rtkBuildConfig.BuildDate)
	log.Println("=======================================================\n\n")

	setupSettings()
	if len(serverId) > 0 && len(serverIpInfo) > 0 && listenHost != "" && listentPort > 0 {
		rtkGlobal.RelayServerID = serverId
		rtkGlobal.RelayServerIPInfo = serverIpInfo
		rtkGlobal.ListenPort = listentPort
		rtkGlobal.ListenHost = listenHost
		log.Printf("set relayServerID: [%s], relayServerIPInfo:[%s]", serverId, serverIpInfo)
		log.Printf("(MDNS) set host[%s] listen port: [%d]\n", listenHost, listentPort)
		rtkPlatform.SetNetWorkConnected(true)
	} else {
		log.Printf("MainInit  parameter  is  not set \n\n")
		return
	}

	//setupLogFileSetting()
	rtkUtils.GoSafe(func() { rtkDebug.DebugCmdLine() })

	rtkUtils.GoSafe(func() { BusinessProcess() })

	// TODO:  check below code if necessary  by TSTAS-40
	/*sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	<-sigs
	cancel()
	log.Println("Termination signal received, performing cleanup...")
	time.Sleep(10 * time.Millisecond)*/
	select {}
}

func BusinessProcess() {
	for {
		ctx, cancel := context.WithCancel(context.Background())
		businessStart(ctx)

		<-rtkConnection.GetNetworkSwitchFlag()
		log.Printf("[%s] Network is Switch, so cancel all business! *****************\n\n\n", rtkUtils.GetFuncInfo())
		cancel()
		time.Sleep(5 * time.Second)
		log.Printf("[%s] business is restart!", rtkUtils.GetFuncInfo())
	}
}

func businessStart(ctx context.Context) {
	rtkUtils.GoSafe(func() {
		for {
			select {
			case <-ctx.Done():
				rtkPeer2Peer.CaneclProcessForPeerMap()
				return
			case id := <-rtkConnection.StartProcessChan:
				rtkPeer2Peer.StartProcessForPeer(id)
			case id := <-rtkConnection.EndProcessChan:
				rtkPeer2Peer.EndProcessForPeer(id)
			case <-time.After(5 * time.Second):
				if rtkPeer2Peer.GetProcessForPeerCount() != rtkUtils.GetClientCount() {
					log.Printf("[%s] Attention please, get ClientCount:[%d] is not match ProcessCount:[%d] ", rtkUtils.GetFuncInfo(), rtkUtils.GetClientCount(), rtkPeer2Peer.GetProcessForPeerCount())
				}
			}
		}
	})

	// Init connection
	rtkConnection.ConnectionInit()
	rtkUtils.GoSafe(func() { rtkConnection.Run(ctx) })
	rtkUtils.GoSafe(func() { rtkMdns.MdnsServiceRun(ctx) })
}
