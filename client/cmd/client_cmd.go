package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	rtkBuildConfig "rtk-cross-share/client/buildConfig"
	rtkConnection "rtk-cross-share/client/connection"
	rtkDebug "rtk-cross-share/client/debug"
	rtkGlobal "rtk-cross-share/client/global"
	rtkLogin "rtk-cross-share/client/login"
	rtkPeer2Peer "rtk-cross-share/client/peer2peer"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strconv"
	"time"
)

var (
	getLanServerMacFlagChan  = make(chan struct{}, 1)
	extractDIASFlagChan      = make(chan struct{})
	networkSwitchFlagChan    = make(chan struct{})
	getLanServerMacTimeStamp int64
)

func init() {
	getLanServerMacTimeStamp = 0
	rtkPlatform.SetupCallbackSettings()

	rtkGlobal.ListenHost = rtkMisc.DefaultIp
	rtkGlobal.ListenPort = rtkGlobal.DefaultPort
	rtkGlobal.NodeInfo.Platform = rtkPlatform.GetPlatform()

	rtkPlatform.SetGoNetworkSwitchCallback(func() {
		networkSwitchFlagChan <- struct{}{}
	})

	rtkPlatform.SetGoPipeConnectedCallback(func() {
		// Update system info
		ipAddr := rtkMisc.ConcatIP(rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort)
		serviceVer := "v" + rtkBuildConfig.Version + " (" + rtkBuildConfig.BuildDate + ")"
		rtkPlatform.GoUpdateSystemInfo(ipAddr, serviceVer)

		// Update all clients status
		clientMap := rtkUtils.GetClientMap()
		for _, info := range clientMap {
			rtkPlatform.GoUpdateClientStatus(1, info.IpAddr, info.ID, info.DeviceName, info.SourcePortType)
		}

		// Update DIAS status
		rtkLogin.NotifyDIASStatus(rtkLogin.CurrentDiasStatus)
	})

	rtkPlatform.SetGoExtractDIASCallback(func() {
		log.Printf("Detect cable plug-out")
		extractDIASFlagChan <- struct{}{}
	})

	rtkPlatform.SetGoGetMacAddressCallback(func(mac string) {
		log.Printf("Get MAC address: %s", mac)
		if getLanServerMacTimeStamp != 0 && (time.Now().UnixMilli()-getLanServerMacTimeStamp < 200) {
			log.Printf("GetMacAddress trigger interval time is too short, so discard it!")
			return
		}
		getLanServerMacTimeStamp = time.Now().UnixMilli()
		rtkLogin.SetLanServerName(mac)
		getLanServerMacFlagChan <- struct{}{}
	})

	rtkPlatform.SetGoAuthStatusCodeCallback(func(status uint8) {
		log.Printf("Get auth status: %d", status)
		if status == 1 {
			rtkPlatform.GoReqSourceAndPort()
		} else {
			rtkLogin.NotifyDIASStatus(rtkLogin.DIAS_Status_Authorization_Failed)
			log.Printf("Warning: UNAUTHORIZED Client!")
		}
	})

	rtkPlatform.SetGoDIASSourceAndPortCallback(func(source uint8, port uint8) {
		// Reserved param about source and port
		log.Printf("Get (source,port)=(%d, %d)", source, port)
		rtkLogin.SendReqClientListToLanServer()
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
		addrs[i] = fmt.Sprintf(a, rtkMisc.DefaultIp, port)
	}

	return addrs
}

func Run() {
	rtkLogin.NotifyDIASStatus(rtkLogin.DIAS_Status_Wait_DiasMonitor)
	rtkMisc.InitLog(rtkPlatform.GetLogFilePath(), rtkPlatform.GetCrashLogFilePath(), 0)
	if rtkBuildConfig.Debug == "1" {
		rtkMisc.SetupLogConsoleFile()
	} else {
		rtkMisc.SetupLogFile()
	}

	log.Println("========================")
	log.Println("Version: ", rtkBuildConfig.Version)
	log.Println("Build Date: ", rtkBuildConfig.BuildDate)
	log.Printf("========================\n\n")

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

	rtkMisc.GoSafe(func() { rtkDebug.DebugCmdLine() })

	rtkMisc.GoSafe(func() { businessProcess() })

	select {}
}

func MainInit(serverId, serverIpInfo, listenHost string, listentPort int) {
	rtkLogin.NotifyDIASStatus(rtkLogin.DIAS_Status_Wait_DiasMonitor)
	rtkMisc.InitLog(rtkPlatform.GetLogFilePath(), rtkPlatform.GetCrashLogFilePath(), 0)
	if rtkBuildConfig.Debug == "1" {
		rtkMisc.SetupLogConsoleFile()
	} else {
		rtkMisc.SetupLogFile()
	}

	log.Println("=======================================================")
	log.Println("Version: ", rtkBuildConfig.Version)
	log.Println("Build Date: ", rtkBuildConfig.BuildDate)
	log.Println("=======================================================\n\n")

	if len(serverId) > 0 && len(serverIpInfo) > 0 &&
		listenHost != "" &&
		listenHost != rtkMisc.DefaultIp &&
		listenHost != rtkMisc.LoopBackIp &&
		listentPort > rtkGlobal.DefaultPort {
		rtkGlobal.RelayServerID = serverId
		rtkGlobal.RelayServerIPInfo = serverIpInfo
		rtkGlobal.ListenPort = listentPort
		rtkGlobal.ListenHost = listenHost
		rtkGlobal.NodeInfo.IPAddr.PublicIP = listenHost
		rtkGlobal.NodeInfo.IPAddr.PublicPort = strconv.Itoa(listentPort)
		log.Printf("set relayServerID: [%s], relayServerIPInfo:[%s]", serverId, serverIpInfo)
		log.Printf("p2p set host[%s] listen port: [%d]\n", listenHost, listentPort)
		rtkPlatform.SetNetWorkConnected(true)
	} else {
		log.Printf("listenHost:[%s] listentPort:[%d]", listenHost, listentPort)
		log.Fatalf("MainInit  parameter is invalid \n\n")
	}

	rtkMisc.GoSafe(func() { businessProcess() })

	select {}
}

func businessProcess() {
	rtkLogin.BrowseInstance()

	var cancelFunc func()
	cancelFunc = nil
	for {
		select {
		case <-getLanServerMacFlagChan:
			log.Println("===========================================================================")
			log.Println("******** Get lan Server mac Address, business start! ********")
			if cancelFunc != nil {
				log.Printf("******** Cancel the old business first! ********")
				cancelFunc()
				time.Sleep(100 * time.Millisecond) // wait for print cancel log
				cancelFunc = nil
			}
			log.Println("===========================================================================\n\n")
			ctx, cancel := context.WithCancel(context.Background())
			cancelFunc = cancel
			rtkLogin.NotifyDIASStatus(rtkLogin.DIAS_Status_Connectting_DiasService)
			businessStart(ctx)
		case <-networkSwitchFlagChan:
			log.Println("===========================================================================")
			if cancelFunc != nil {
				log.Printf("******** Client Network is Switch, cancel old business! ******** ")
				cancelFunc()
				time.Sleep(5 * time.Second)
				log.Println("===========================================================================\n\n")
				log.Printf("[%s] business is restart!", rtkMisc.GetFuncInfo())
				ctx, cancel := context.WithCancel(context.Background())
				cancelFunc = cancel
				businessStart(ctx)
			} else {
				log.Printf("******** Client Network is Switch, business is not start! ******** ")
				log.Println("===========================================================================\n\n")
			}
		case <-extractDIASFlagChan:
			log.Println("===========================================================================")
			if cancelFunc != nil {
				log.Printf("******** DIAS is extract, cancel all business! ******** ")
				cancelFunc()
				cancelFunc = nil
				rtkLogin.BrowseInstance()
				time.Sleep(100 * time.Millisecond) // wait for print all cancel log
			} else {
				log.Printf("******** DIAS is extract, business is not start! ******** ")
			}
			log.Println("===========================================================================\n\n")
		}
	}
}

func businessStart(ctx context.Context) {
	rtkMisc.GoSafe(func() {
		ticker := time.NewTicker(time.Duration(5) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				// TODO: refine this cancel flow, we need send disconnect message to all peer befor cancel
				rtkPeer2Peer.SendDisconnectToAllPeer(true)
				return
			case id := <-rtkConnection.StartProcessChan:
				rtkPeer2Peer.StartProcessForPeer(id)
			case id := <-rtkConnection.EndProcessChan:
				rtkPeer2Peer.EndProcessForPeer(id)
			case <-rtkConnection.CancelAllProcess:
				rtkPeer2Peer.CaneclProcessForPeerMap()
			case <-rtkLogin.DisconnectLanServerFlag:
				//TODO: check this flow
				rtkPeer2Peer.SendDisconnectToAllPeer(false)
				rtkConnection.OfflineAllStreamEvent()
			case <-ticker.C:
				nProcessCount := rtkPeer2Peer.GetProcessForPeerCount()
				nClientCount := rtkUtils.GetClientCount()
				if nProcessCount != nClientCount {
					log.Printf("[%s] Attention please, get ClientCount:[%d] is not match ProcessCount:[%d] \n\n", rtkMisc.GetFuncInfo(), nClientCount, nProcessCount)
				}
			}
		}
	})

	// Init connection
	rtkConnection.ConnectionInit()
	rtkMisc.GoSafe(func() { rtkConnection.Run(ctx) })

	rtkMisc.GoSafe(func() { rtkLogin.ConnectLanServerRun(ctx) })

}
