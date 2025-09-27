package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	rtkBuildConfig "rtk-cross-share/client/buildConfig"
	rtkCommon "rtk-cross-share/client/common"
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
	cablePlugInFlagChan      = make(chan struct{}, 1)
	cablePlugOutFlagChan     = make(chan struct{})
	networkSwitchFlagChan    = make(chan struct{})
	clientVerInvalidFlagChan = make(chan struct{})
	getLanServerMacTimeStamp int64
)

func init() {
	getLanServerMacTimeStamp = 0

	rtkGlobal.ListenHost = rtkMisc.DefaultIp
	rtkGlobal.ListenPort = rtkGlobal.DefaultPort
	rtkGlobal.NodeInfo.Platform = rtkPlatform.GetPlatform()

	rtkConnection.SetStartProcessForPeerCallback(rtkPeer2Peer.StartProcessForPeer)
	rtkConnection.SetSendDisconnectMsgToPeerCallback(rtkPeer2Peer.SendDisconnectMsgToPeer)
	rtkLogin.SetDisconnectAllClientCallback(rtkConnection.CancelStreamPool)
	rtkLogin.SetCancelAllBusinessCallback(func() {
		clientVerInvalidFlagChan <- struct{}{}
	})

	rtkPlatform.SetGoNetworkSwitchCallback(func() {
		networkSwitchFlagChan <- struct{}{}
	})

	rtkPlatform.SetDetectPluginEventCallback(func(isPlugin bool, productName string) {
		if isPlugin {
			log.Printf("Detect cable plug-in")
			rtkLogin.SetProductName(productName)
			cablePlugInFlagChan <- struct{}{}
		} else {
			log.Printf("Detect cable plug-out")
			cablePlugOutFlagChan <- struct{}{}
		}
	})

	rtkPlatform.SetGoExtractDIASCallback(func() {
		log.Printf("Detect cable plug-out")
		cablePlugOutFlagChan <- struct{}{}
	})

	rtkPlatform.SetGoGetMacAddressCallback(func(mac string) {
		log.Printf("Detect cable plug-in, get MAC address: %s", mac)
		if getLanServerMacTimeStamp != 0 && (time.Now().UnixMilli()-getLanServerMacTimeStamp < 200) {
			log.Printf("GetMacAddress trigger interval time is too short, so discard it!")
			return
		}
		getLanServerMacTimeStamp = time.Now().UnixMilli()
		rtkLogin.SetLanServerInstance(mac)
		cablePlugInFlagChan <- struct{}{}
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
	log.Println("=======================================================")
	log.Println("Version: ", rtkGlobal.ClientVersion)
	log.Println("Build Date: ", rtkBuildConfig.BuildDate)
	log.Println("=======================================================\n\n")

	rtkLogin.NotifyDIASStatus(rtkLogin.DIAS_Status_Wait_DiasMonitor)
	lockFilePath := rtkPlatform.GetLockFilePath()
	file, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Println("Failed to open or create lock file:", err)
		return
	}
	defer file.Close()

	err = rtkPlatform.LockFile(file)
	if err != nil {
		log.Println("Another instance is already running.")
		return
	}
	defer rtkPlatform.UnlockFile(file)

	if rtkBuildConfig.CmdDebug == "1" {
		rtkMisc.GoSafe(func() { rtkDebug.DebugCmdLine() })
	}

	rtkMisc.GoSafe(func() { businessProcess(context.Background()) })

	select {}
}

func MainInit(serverId, serverIpInfo, listenHost string, listentPort int) {
	log.Println("=======================================================")
	log.Println("Version: ", rtkGlobal.ClientVersion)
	log.Println("Build Date: ", rtkBuildConfig.BuildDate)
	log.Println("=======================================================\n\n")

	rtkLogin.NotifyDIASStatus(rtkLogin.DIAS_Status_Wait_DiasMonitor)
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

	/*lockFilePath := rtkPlatform.GetLockFilePath()
	file, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Println("Failed to open or create lock file:", err)
		return
	}
	defer file.Close()

	err = rtkPlatform.LockFile(file)
	if err != nil {
		log.Println("Another instance is already running.\n\n")
		return
	}
	defer rtkPlatform.UnlockFile(file)*/

	rtkMisc.GoSafe(func() { businessProcess(context.Background()) })

	select {}
}

func businessProcess(ctx context.Context) {
	rtkLogin.BrowseInstance()

	var cancelBusinessFunc func(source rtkCommon.CancelBusinessSource)
	cancelBusinessFunc = nil
	var sonCtx context.Context
	for {
		select {
		case <-cablePlugInFlagChan:
			log.Println("===========================================================================")
			log.Println("******** DIAS is access, business start! ********")
			if cancelBusinessFunc != nil {
				log.Printf("******** Cancel the old business first! ********")
				cancelBusinessFunc(rtkCommon.SourceCablePlugIn)
				time.Sleep(100 * time.Millisecond) // wait for print cancel log
				cancelBusinessFunc = nil
				rtkLogin.BrowseInstance()
			}
			log.Println("===========================================================================\n\n")

			rtkLogin.NotifyDIASStatus(rtkLogin.DIAS_Status_Connectting_DiasService)
			sonCtx, cancelBusinessFunc = rtkUtils.WithCancelSource(ctx)
			rtkMisc.GoSafe(func() { businessStart(sonCtx) })
		case <-networkSwitchFlagChan:
			log.Println("===========================================================================")
			if cancelBusinessFunc != nil {
				log.Printf("******** Client Network is Switch, cancel old business! ******** ")
				cancelBusinessFunc(rtkCommon.SourceNetworkSwitch)
				time.Sleep(100 * time.Millisecond) // wait for print cancel log
				rtkLogin.BrowseInstance()
				log.Println("===========================================================================\n\n")
				log.Printf("[%s] business is restart!", rtkMisc.GetFuncInfo())

				sonCtx, cancelBusinessFunc = rtkUtils.WithCancelSource(ctx)
				rtkMisc.GoSafe(func() { businessStart(sonCtx) })
			} else {
				rtkLogin.BrowseInstance()
				log.Printf("******** Client Network is Switch, business is not start! ******** ")
				log.Println("===========================================================================\n\n")
			}
		case <-cablePlugOutFlagChan:
			log.Println("===========================================================================")
			if cancelBusinessFunc != nil {
				log.Printf("******** DIAS is extract, cancel all business! ******** ")
				cancelBusinessFunc(rtkCommon.SourceCablePlugOut)
				cancelBusinessFunc = nil
				time.Sleep(100 * time.Millisecond) // wait for print all cancel log
				rtkLogin.BrowseInstance()
			} else {
				log.Printf("******** DIAS is extract, business is not start! ******** ")
			}
			log.Println("===========================================================================\n\n")
		case <-clientVerInvalidFlagChan:
			log.Println("===========================================================================")
			log.Printf("******** Client Version is too old, and must be forcibly updated, cancel all business! ******** ")
			if cancelBusinessFunc != nil {
				cancelBusinessFunc(rtkCommon.SourceVerInvalid)
				cancelBusinessFunc = nil
				time.Sleep(100 * time.Millisecond) // wait for print all cancel log
			} else {
				log.Printf("******** business is not start! ******** ")
			}
			log.Println("===========================================================================\n\n")
		}
	}
}

func businessStart(ctx context.Context) {
	// Init connection
	rtkConnection.ConnectionInit(ctx)

	rtkMisc.GoSafe(func() { rtkConnection.Run(ctx) })

	rtkMisc.GoSafe(func() { rtkLogin.ConnectLanServerRun(ctx) })

}
