package login

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"time"
)

func SetLanServerInstance(name string) {
	lanServerInstance = name
}

func IsFirstPlugInEvent() bool {
	if len(displayInfoList) == 0 {
		return true
	}
	return false
}

func IsLastPlugInEvent() bool {
	if len(displayInfoList) <= 1 {
		return true
	}
	return false
}

func IsChangeInstance(instance string) bool {
	if lanServerInstance == "" {
		return false
	}
	return !rtkUtils.ContentEqual([]byte(instance), []byte(lanServerInstance))
}

func SetPlugDisplayEventInfo(displayEventInfo rtkCommon.DisplayEventInfo) {
	if displayEventInfo.PlugEvent == 1 {
		displayInfoList = append(displayInfoList, displayEventInfo)
	} else {
		tmpDisplayList := make([]rtkCommon.DisplayEventInfo, 0)
		for _, display := range displayInfoList {
			if display.Source == displayEventInfo.Source && display.Port == displayEventInfo.Port {
				continue
			}
			tmpDisplayList = append(tmpDisplayList, display)
		}
		displayInfoList = displayInfoList[:0]
		displayInfoList = tmpDisplayList
	}
}

func IsSameDisplayInfo(srcPortInfo rtkMisc.SourcePortInfo) bool {
	for _, display := range displayInfoList {
		if display.Source == srcPortInfo.Source && display.Port == srcPortInfo.Port &&
			display.UdpMousePort == srcPortInfo.UdpMousePort && display.UdpKeyboardPort == srcPortInfo.UdpKeyboardPort {
			return true
		}
	}
	return false
}

func SetProductName(name string) {
	g_ProductName = name
}

func init() {
	lanServerInstance = ""
	g_ProductName = ""
	g_monitorName = ""
	lanServerRunning.Store(false)
	displayInfoList = make([]rtkCommon.DisplayEventInfo, 0)

	pSafeConnect = &safeConnect{
		connectMutex:     sync.RWMutex{},
		connectLanServer: nil,
		isAlive:          false,
	}
	heartBeatTicker = NewHeartBeatTicker(rtkCommon.PingInterval)
	cancelBrowse = nil

	disconnectAllClientFunc = nil
	cancelAllBusinessFunc = nil

	rtkPlatform.SetGoConnectLanServerCallback(mobileInitLanServer)

	rtkPlatform.SetGoBrowseLanServerCallback(func() {
		log.Printf("[%s][mobile] Browse LanServer monitor triggered!", rtkMisc.GetFuncInfo())
		lanServerInstance = ""
		stopLanServerBusiness()
		BrowseInstance()
	})

	rtkPlatform.SetGoSetMsgEventCallback(sendPlatformMsgEventToLanServer)
}

func computerInitLanServer(ctx context.Context) {
	stopBrowseInstance()
	retryCnt := 0
	bPrintErrLog := true
	for {
		errCode := initLanServer(ctx, bPrintErrLog)
		if errCode == rtkMisc.SUCCESS {
			break
		}

		if retryCnt <= g_retryServerMaxCnt {
			retryCnt++
		}
		if retryCnt == g_retryServerMaxCnt {
			bPrintErrLog = false
			log.Printf("initLanServer %d times failed, errCode:%d ! try to lookup Service over again ...", retryCnt, errCode)
			serverInstanceMap.Delete(lanServerInstance)
			NotifyDIASStatus(DIAS_Status_Connected_DiasService_Failed)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(g_retryServerInterval):
		}
	}
}

func mobileInitLanServer(instance string) {
	initLanServerMutex.Lock()
	defer initLanServerMutex.Unlock()

	if !lanServerRunning.Load() {
		log.Printf("[%s] ConnectLanServerRun is not running!", rtkMisc.GetFuncInfo())
		return
	}

	log.Printf("[%s][mobile] connect LanServer, Instance:%s", rtkMisc.GetFuncInfo(), instance)

	mapValue, ok := serverInstanceMap.Load(instance)
	if !ok {
		log.Printf("[%s][mobile] unknown instance:%s", rtkMisc.GetFuncInfo(), instance)
		NotifyDIASStatus(DIAS_Status_Connected_DiasService_Failed)
		return
	}

	if currentDiasStatus > DIAS_Status_Connectting_DiasService {
		log.Printf("[%s][mobile] currentDiasStatus:%d, not allowed connect LanServer over again! ", rtkMisc.GetFuncInfo(), currentDiasStatus)
		if currentDiasStatus != DIAS_Status_Wait_Other_Clients && currentDiasStatus != DIAS_Status_Get_Clients_Success {
			NotifyDIASStatus(DIAS_Status_Connected_DiasService_Failed)
		}
		return
	}

	if lanServerInstance != "" {
		if lanServerInstance == instance &&
			pSafeConnect.IsAlive() &&
			(currentDiasStatus == DIAS_Status_Wait_screenCasting || currentDiasStatus == DIAS_Status_Wait_Other_Clients || currentDiasStatus == DIAS_Status_Get_Clients_Success) {
			log.Printf("[%s][mobile] Instance:%s is already connected, skip it!", rtkMisc.GetFuncInfo(), instance)
			return
		}
		stopLanServerBusiness()
	}

	lanServerInstance = instance
	g_monitorName = mapValue.(browseParam).monitorName

	retryCnt := 0
	bPrintErrLog := true
	for {
		retryCnt++
		errCode := initLanServer(context.Background(), bPrintErrLog)
		if errCode == rtkMisc.SUCCESS {
			break
		}

		if retryCnt == g_retryServerMaxCnt {
			bPrintErrLog = false
			log.Printf("initLanServer %d times failed, errCode:%d ! Browse service instances go on ...", retryCnt, errCode)
			NotifyDIASStatus(DIAS_Status_Connected_DiasService_Failed)
			return
		}
		<-time.After(g_retryServerInterval)
	}

}

func ConnectLanServerRun(ctx context.Context) {
	defer func() {
		cancelLanServerBusiness()
		if source, ok := rtkUtils.GetCancelSource(ctx); ok {
			if source == rtkCommon.SourceNetworkSwitch || source == rtkCommon.SourceVerInvalid || source == rtkCommon.SourceCablePlugIn {
				NotifyDIASStatus(DIAS_Status_Connectting_DiasService)
			} else if source == rtkCommon.SourceCablePlugOut {
				NotifyDIASStatus(DIAS_Status_Wait_DiasMonitor)
			}
			log.Printf("ConnectLanServerRun was canceled from source:%d", source)
		}
	}()

	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac {
		g_lookupByUnicast = true
		computerInitLanServer(ctx)
	} else {
		lanServerRunning.Store(true)
		serverInstanceMap.Range(func(k, v any) bool {
			param := v.(browseParam)
			rtkPlatform.GoNotifyBrowseResult(param.monitorName, param.instance, param.ip, param.ver, param.timeStamp)
			return true
		})
	}

	readResult := make(chan struct {
		buffer  string
		errCode rtkMisc.CrossShareErr
	}, 5) //add channel buffer to pervent stucking the channel receiver

	rtkMisc.GoSafe(func() {
		var printNetworkErr = true
		for {
			select {
			case <-ctx.Done():
				close(readResult)
				return
			default:
				conn := pSafeConnect.GetConnect() // TODO: refine this flow
				if conn == nil {
					if printNetworkErr {
						log.Printf("[%s] LanServer IPAddr:[%s] GetConnect is nil , try to reconnect", rtkMisc.GetFuncInfo(), pSafeConnect.ConnectIPAddr())
						printNetworkErr = false
					}
					time.Sleep(100 * time.Millisecond)
					continue
				}

				printNetworkErr = true
				errCode := rtkMisc.SUCCESS
				readStrLine, err := bufio.NewReader(conn).ReadString('\n')
				// _, err = pSafeConnect.Read(&buffer)  //this cause dead lock
				if err != nil {
					log.Printf("[%s] LanServer IPAddr:[%s] ReadString error:%+v ", rtkMisc.GetFuncInfo(), pSafeConnect.ConnectIPAddr(), err)
					errCode = rtkMisc.ERR_NETWORK_C2S_READ
					if errors.Is(err, io.EOF) {
						errCode = rtkMisc.ERR_NETWORK_C2S_READ_EOF
					} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						errCode = rtkMisc.ERR_NETWORK_C2S_READ_TIME_OUT
					}

				}
				readResult <- struct {
					buffer  string
					errCode rtkMisc.CrossShareErr
				}{buffer: readStrLine, errCode: errCode}
			}
		}
	})

	defer heartBeatTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartBeatTicker.C():
			rtkMisc.GoSafe(func() { heartBeatFunc(ctx) })
		case readData, ok := <-readResult:
			if !ok {
				continue
			}
			if readData.errCode != rtkMisc.SUCCESS {
				pSafeConnect.Close()
				log.Printf("[%s] read lanServer socket Data, errcode:%d", rtkMisc.GetFuncInfo(), readData.errCode)
				continue
			}
			if readData.buffer == "" || len(readData.buffer) == 0 {
				log.Printf("[%s] read lanServer socket Data is null, continue!", rtkMisc.GetFuncInfo())
				continue
			}
			bufferData := []byte(readData.buffer)
			handleReadMessageFromServer(bufferData)
		}
	}
}

func heartBeatFunc(ctx context.Context) {
	checkPingServerTimeout()

	if !pSafeConnect.IsAlive() {
		rtkMisc.GoSafe(func() { ReConnectLanServer(ctx) })
		return
	}

	sendReqHeartbeatToLanServer()
}

func getLanServerAddr(ctx context.Context, bPrintErr bool) (string, rtkMisc.CrossShareErr) {
	if lanServerInstance == "" {
		if bPrintErr {
			log.Printf("[%s] lanServerInstance is not set!", rtkMisc.GetFuncInfo())
		}
		return "", rtkMisc.ERR_BIZ_C2S_GET_NO_SERVER_NAME
	}

	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac {
		mapValue, ok := serverInstanceMap.Load(lanServerInstance)
		if ok {
			lanServerIp := mapValue.(browseParam).ip
			g_monitorName = mapValue.(browseParam).monitorName
			if len(lanServerIp) > 0 {
				log.Printf("get LanServer addr %s from browse serverInstanceMap!", lanServerIp)
				return lanServerIp, rtkMisc.SUCCESS
			}
		}

		time.Sleep(50 * time.Millisecond) // Delay 50ms between "stop browse server" and "start lookup server"
		tOctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		return lookupLanServer(tOctx, lanServerInstance, rtkMisc.LanServiceType, rtkMisc.LanServerDomain, bPrintErr)
	} else {
		mapValue, ok := serverInstanceMap.Load(lanServerInstance)
		if ok {
			lanServerIp := mapValue.(browseParam).ip
			g_monitorName = mapValue.(browseParam).monitorName
			log.Printf("[%s][Mobile] get service Instance=(%s), ip=(%s) from map", rtkMisc.GetFuncInfo(), lanServerInstance, lanServerIp)
			return lanServerIp, rtkMisc.SUCCESS
		} else {
			log.Printf("[%s] lanServerInstance:[%s] is invalid, get no data from instance map!", rtkMisc.GetFuncInfo(), lanServerInstance)
		}

	}
	return "", rtkMisc.ERR_NETWORK_C2S_BROWSER_INVALID
}

func connectToLanServer(ctx context.Context, bPrintErr bool) rtkMisc.CrossShareErr {
	if !pSafeConnect.IsAlive() {
		serverAddr, errCode := getLanServerAddr(ctx, bPrintErr)
		if errCode != rtkMisc.SUCCESS {
			return errCode
		}

		if bPrintErr {
			log.Printf("get LanServer addr:%s  by serverInstance:[%s], try to Dial it!", serverAddr, lanServerInstance)
		}

		tOctx, dialCancelFn := context.WithTimeout(ctx, time.Duration(5*time.Second))
		defer dialCancelFn()
		d := net.Dialer{Timeout: time.Duration(5 * time.Second)}
		pConnectLanServer, err := d.DialContext(tOctx, "tcp", serverAddr)

		if err != nil {
			if bPrintErr {
				log.Printf("connecting to lanServerAddr[%s] Error:%+v ", serverAddr, err.Error())
			}

			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					return rtkMisc.ERR_NETWORK_C2S_DIAL_TIMEOUT
				}
			}
			return rtkMisc.ERR_NETWORK_C2S_DIAL
		}

		pSafeConnect.Reset(pConnectLanServer)
		log.Printf("Connect LanServerAddr:[%s] success! LocalAddr:[%s]", serverAddr, pConnectLanServer.LocalAddr().String())

		stopBrowseInstance() // mobile need stop Browse
	}
	return rtkMisc.SUCCESS
}

func initLanServer(ctx context.Context, bPrintErr bool) rtkMisc.CrossShareErr {
	resultCode := connectToLanServer(ctx, bPrintErr)
	if resultCode != rtkMisc.SUCCESS {
		return resultCode
	}

	rtkPlatform.GoMonitorNameNotify(g_monitorName)
	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac {
		NotifyDIASStatus(DIAS_Status_Checking_Authorization)
	} else {
		NotifyDIASStatus(DIAS_Status_Wait_screenCasting)
	}

	lanServerHeartbeatStart()
	return sendReqInitClientToLanServer()
}

func ReConnectLanServer(ctx context.Context) {
	stopLanServerBusiness()
	time.Sleep(100 * time.Millisecond) // Delay 100ms between "disconnect all client" and "start reconnect lan server"

	log.Println("Try to connect to LanServer over again!\n")

	retryCnt := 0
	bPrintErrLog := true
	for {
		errCode := initLanServer(ctx, bPrintErrLog)
		if errCode == rtkMisc.SUCCESS {
			log.Printf("ReConnectLanServer success!")
			break
		}

		if retryCnt <= g_retryServerMaxCnt {
			retryCnt++
		}
		if retryCnt == g_retryServerMaxCnt {
			NotifyDIASStatus(DIAS_Status_Connectting_DiasService)
			rtkPlatform.GoMonitorNameNotify("")

			if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac {
				serverInstanceMap.Delete(lanServerInstance)
				log.Printf("initLanServer %d times failed, errCode:%d ! try to Lookup Service over again...", retryCnt, errCode)
			} else {
				log.Printf("initLanServer %d times failed, errCode:%d ! try to Browse Service over again...", retryCnt, errCode)
				BrowseInstance()
			}

			bPrintErrLog = false
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(g_retryServerInterval):
		}
	}
}

func lanServerHeartbeatStart() {
	heartBeatTicker.Start()
	log.Println("lanServer heartbeat is Running...")
}

func lanServerHeartbeatStop() {
	heartBeatTicker.Stop()
	log.Println("lanServer heartbeat is Stop!")
	updatePingServerErrCntReset()
}

func StopLanServerRun() {
	stopLanServerBusiness()
}

func stopLanServerBusiness() {
	log.Printf("connect lanServer business is all stop!")
	lanServerHeartbeatStop()
	pSafeConnect.Close()
	rtkPlatform.GoMonitorNameNotify("")
	NotifyDIASStatus(DIAS_Status_Connectting_DiasService)
	displayInfoList = displayInfoList[:0]

	if disconnectAllClientFunc == nil {
		log.Printf("disconnectAllClientFunc is nil, not cancel all client stream and business!")
		return
	}
	disconnectAllClientFunc()
}

func cancelLanServerBusiness() {
	log.Printf("connect lanServer business is all cancel!")
	rtkPlatform.GoMonitorNameNotify("")
	pSafeConnect.Close()
	lanServerInstance = ""
	stopBrowseInstance()
	lanServerRunning.Store(false)
	displayInfoList = displayInfoList[:0]
}

func NotifyDIASStatus(status CrossShareDiasStatus) {
	if currentDiasStatus != status {
		rtkPlatform.GoDIASStatusNotify(uint32(status))

		if (status == DIAS_Status_Wait_DiasMonitor) ||
			(status == DIAS_Status_Connectting_DiasService) ||
			(status == DIAS_Status_Wait_Other_Clients) ||
			(status == DIAS_Status_Get_Clients_Success) {
			currentDiasStatus = status
		}
	}
}
