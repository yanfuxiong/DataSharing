package login

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

func SetLanServerInstance(name string) {
	lanServerInstance = name
}

func SetProductName(name string) {
	g_ProductName = name
}

func init() {
	lanServerInstance = ""
	lanServerAddr = ""
	g_ProductName = ""
	g_monitorName = ""
	isReconnectRunning.Store(false)
	lanServerRunning.Store(false)
	reconnectCancelFunc = nil

	pSafeConnect = &safeConnect{
		connectMutex:     sync.RWMutex{},
		connectLanServer: nil,
		isAlive:          false,
	}
	heartBeatTicker = nil
	cancelBrowse = nil

	disconnectAllClientFunc = nil
	cancelAllBusinessFunc = nil

	rtkPlatform.SetGoConnectLanServerCallback(mobileInitLanServer)

	rtkPlatform.SetGoBrowseLanServerCallback(func() {
		log.Printf("[%s][mobile] Browse LanServer monitor triggered!", rtkMisc.GetFuncInfo())
		lanServerInstance = ""
		lanServerAddr = ""
		serverInstanceMap.Clear()
		stopLanServerBusiness()
		BrowseInstance()
	})
}

func computerInitLanServer(ctx context.Context) {
	stopBrowseInstance()
	retryCnt := 0
	bPrintErrLog := true
	for {
		errCode := initLanServer(bPrintErrLog)
		if errCode == rtkMisc.SUCCESS {
			break
		}

		if retryCnt < g_retryServerMaxCnt {
			retryCnt++
		}
		if retryCnt == g_retryServerMaxCnt {
			bPrintErrLog = false
			log.Printf("initLanServer %d times failed, errCode:%d ! try to lookup Service over again ...", retryCnt, errCode)
			lanServerAddr = ""
			serverInstanceMap.Delete(lanServerInstance)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(g_retryServerInterval):
		}
	}
}

func mobileInitLanServer(instance string) {
	log.Printf("[%s][mobile] connect LanServer, Instance:%s", rtkMisc.GetFuncInfo(), instance)

	mapValue, ok := serverInstanceMap.Load(instance)
	if !ok {
		log.Printf("[%s][mobile] unknown instance:%s", rtkMisc.GetFuncInfo(), instance)
		NotifyDIASStatus(DIAS_Status_Connected_DiasService_Failed)
		return
	}

	if currentDiasStatus > DIAS_Status_Connectting_DiasService {
		log.Printf("[%s][mobile] currentDiasStatus:%d, not allowed connect LanServer over again! ", rtkMisc.GetFuncInfo(), currentDiasStatus)
		NotifyDIASStatus(DIAS_Status_Connected_DiasService_Failed)
		return
	}

	if lanServerInstance != "" {
		stopLanServerBusiness()
	}

	lanServerInstance = instance
	lanServerAddr = mapValue.(browseParam).ip
	g_monitorName = mapValue.(browseParam).monitorName

	retryCnt := 0
	bPrintErrLog := true
	for {
		retryCnt++
		errCode := initLanServer(bPrintErrLog)
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
			log.Printf("[%s] source:%d", rtkMisc.GetFuncInfo(), source)
		}
	}()

	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac {
		computerInitLanServer(ctx)
	} else {
		serverInstanceMap.Range(func(k, v any) bool {
			param := v.(browseParam)
			rtkPlatform.GoNotifyBrowseResult(param.monitorName, param.instance, param.ip, param.ver, param.timeStamp)
			return true
		})
		lanServerRunning.Store(true)
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
				buf := bufio.NewReader(conn)
				errCode := rtkMisc.SUCCESS
				readStrLine, err := buf.ReadString('\n')
				// _, err = pSafeConnect.Read(&buffer)  //this cause dead lock
				if err != nil {
					log.Printf("[%s] LanServer IPAddr:[%s] ReadString error:%+v ", rtkMisc.GetFuncInfo(), pSafeConnect.ConnectIPAddr(), err)
					errCode = rtkMisc.ERR_NETWORK_C2S_READ
					if errors.Is(err, io.EOF) {
						errCode = rtkMisc.ERR_NETWORK_C2S_READ_EOF
					} else if netErr, ok := err.(net.Error); ok {
						if netErr.Timeout() {
							errCode = rtkMisc.ERR_NETWORK_C2S_READ_TIME_OUT
						}
					}
				}
				readResult <- struct {
					buffer  string
					errCode rtkMisc.CrossShareErr
				}{buffer: readStrLine, errCode: errCode}
			}
		}
	})

	heartBeatTicker = time.NewTicker(time.Duration(999 * time.Hour))
	defer heartBeatTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartBeatTicker.C:
			heartBeatFunc()
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

func heartBeatFunc() {
	if isNeedReconnectProcess() {
		return
	}

	nCount := 0
	for {
		nCount++
		if sendReqHeartbeatToLanServer() == rtkMisc.SUCCESS {
			break
		}
		if isNeedReconnectProcess() {
			return
		}

		if nCount >= 3 {
			log.Printf("**************** Attention please, lanServer heartBeat %d times failed! stop try again!**************** \n\n", nCount)
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func BrowseInstance() rtkMisc.CrossShareErr {
	if cancelBrowse != nil {
		cancelBrowse()
		cancelBrowse = nil
		time.Sleep(50 * time.Millisecond) // Delay 50ms between "stop browse server" and "start lookup server"
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancelBrowse = cancel

	resultChan := make(chan browseParam)

	var err rtkMisc.CrossShareErr
	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformiOS {
		err = browseLanServeriOS(ctx, rtkMisc.LanServiceType, resultChan)
	} else if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformAndroid {
		err = browseLanServerAndroid(ctx, rtkMisc.LanServiceType, rtkMisc.LanServerDomain, resultChan)
	} else {
		err = browseLanServer(ctx, rtkMisc.LanServiceType, rtkMisc.LanServerDomain, resultChan)
	}

	rtkMisc.GoSafe(func() {
		for param := range resultChan {
			if len(param.instance) > 0 && len(param.ip) > 0 {
				serverInstanceMap.Store(param.instance, param)
			}
		}
	})
	return err
}

func stopBrowseInstance() {
	if cancelBrowse != nil {
		cancelBrowse()
		cancelBrowse = nil
	}
}

func getLanServerAddr(bPrintErr bool) (string, rtkMisc.CrossShareErr) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac {
		if lanServerInstance == "" {
			if bPrintErr {
				log.Printf("[%s] lanServerInstance is not set!", rtkMisc.GetFuncInfo())
			}
			return "", rtkMisc.ERR_BIZ_C2S_GET_NO_SERVER_NAME
		}

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
		return lookupLanServer(ctx, lanServerInstance, rtkMisc.LanServiceType, rtkMisc.LanServerDomain, bPrintErr)
	} else {
		if lanServerInstance != "" {
			mapValue, ok := serverInstanceMap.Load(lanServerInstance)
			if ok {
				lanServerIp := mapValue.(browseParam).ip
				g_monitorName = mapValue.(browseParam).monitorName
				log.Printf("[%s][Mobile] get service Instance=(%s), ip=(%s) from map", rtkMisc.GetFuncInfo(), lanServerInstance, lanServerIp)
				return lanServerIp, rtkMisc.SUCCESS
			} else {
				log.Printf("[%s] lanServerInstance:[%s] is invalid or get no data from map!", rtkMisc.GetFuncInfo(), lanServerInstance)
			}
		} else {
			log.Printf("[%s] lanServerInstance is null!", rtkMisc.GetFuncInfo())
		}
	}
	return "", rtkMisc.ERR_NETWORK_C2S_BROWSER_INVALID
}

func connectToLanServer(bPrintErr bool) rtkMisc.CrossShareErr {
	if !pSafeConnect.IsAlive() {
		if lanServerAddr == "" {
			serverAddr, errCode := getLanServerAddr(bPrintErr)
			if errCode != rtkMisc.SUCCESS {
				return errCode
			}
			lanServerAddr = serverAddr
		}

		if bPrintErr {
			log.Printf("get LanServer addr:%s  by serverInstance:[%s], try to Dial it!", lanServerAddr, lanServerInstance)
		}
		pConnectLanServer, err := net.DialTimeout("tcp", lanServerAddr, time.Duration(5*time.Second))
		if err != nil {
			if bPrintErr {
				log.Printf("connecting to lanServerAddr[%s] Error:%+v ", lanServerAddr, err.Error())
			}
			NotifyDIASStatus(DIAS_Status_Connected_DiasService_Failed)
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					return rtkMisc.ERR_NETWORK_C2S_DIAL_TIMEOUT
				}
			}
			return rtkMisc.ERR_NETWORK_C2S_DIAL
		}
		pSafeConnect.Reset(pConnectLanServer)
		log.Printf("Connect LanServerAddr:[%s] success!", lanServerAddr)

		stopBrowseInstance() // mobile need stop Browse
	}
	return rtkMisc.SUCCESS
}

func initLanServer(bPrintErr bool) rtkMisc.CrossShareErr {
	resultCode := connectToLanServer(bPrintErr)
	if resultCode != rtkMisc.SUCCESS {
		return resultCode
	}

	rtkPlatform.GoMonitorNameNotify(g_monitorName)
	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac {
		NotifyDIASStatus(DIAS_Status_Checking_Authorization)
	} else {
		NotifyDIASStatus(DIAS_Status_Wait_screenCasting)
	}

	return sendReqInitClientToLanServer()
}

func isNeedReconnectProcess() bool {
	if isReconnectRunning.Load() {
		return true
	}

	if !pSafeConnect.IsAlive() && !isReconnectRunning.Load() {
		stopLanServerBusiness()
		rtkMisc.GoSafe(func() { ReConnectLanServer() })
		return true
	}
	return false
}

func ReConnectLanServer() {
	if isReconnectRunning.Load() {
		return
	}
	isReconnectRunning.Store(true)
	defer isReconnectRunning.Store(false)

	log.Println("Try to connect to LanServer over again!\n")

	ctx, cancecl := context.WithCancel(context.Background())
	reconnectCancelFunc = cancecl
	defer func() {
		if reconnectCancelFunc != nil {
			reconnectCancelFunc()
			reconnectCancelFunc = nil
		}
	}()

	retryCnt := 0
	bPrintErrLog := true
	for {
		errCode := initLanServer(bPrintErrLog)
		if errCode == rtkMisc.SUCCESS {
			log.Printf("ReConnectLanServer success!")
			break
		}

		if retryCnt < g_retryServerMaxCnt {
			retryCnt++
		}
		if retryCnt == g_retryServerMaxCnt {
			NotifyDIASStatus(DIAS_Status_Connectting_DiasService)
			rtkPlatform.GoMonitorNameNotify("")

			lanServerAddr = ""
			if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac {
				serverInstanceMap.Delete(lanServerInstance)
				log.Printf("initLanServer %d times failed, errCode:%d ! try to Lookup Service over again...", retryCnt, errCode)
			} else {
				lanServerInstance = ""
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
	heartBeatTicker.Reset(time.Duration(rtkMisc.ClientHeartbeatInterval * time.Second))
	log.Println("lanServer heartbeat is Running...")
}

func lanServerHeartbeatStop() {
	heartBeatTicker.Reset(time.Duration(999 * time.Hour))
	//heartBeatTicker.Stop()  //Go Version must be 1.23 or greater
	log.Println("lanServer heartbeat is Stop!")
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
	lanServerAddr = ""
	stopBrowseInstance()
	serverInstanceMap.Clear()
	lanServerRunning.Store(false)
	if isReconnectRunning.Load() {
		if reconnectCancelFunc != nil {
			reconnectCancelFunc()
			reconnectCancelFunc = nil
		}
	}
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

func browseLanServer(ctx context.Context, serviceType, domain string, resultChan chan<- browseParam) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(rtkUtils.GetNetInterfaces(), nil)
	if err != nil {
		log.Printf("[%s] Failed to initialize resolver:%+v", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_NETWORK_C2S_RESOLVER
	}

	getTextRecordMap := func(textRecord []string) map[string]string {
		txtMap := make(map[string]string)
		for _, txt := range textRecord {
			parts := strings.SplitN(txt, "=", 2)
			if len(parts) == 2 {
				txtMap[parts[0]] = parts[1]
			}
		}
		return txtMap
	}
	entries := make(chan *zeroconf.ServiceEntry)
	rtkMisc.GoSafe(func() {
		for entry := range entries {
			if len(entry.AddrIPv4) > 0 {
				lanServerIp := fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
				txtMap := getTextRecordMap(entry.Text)
				textRecordmonitorName := txtMap[rtkMisc.TextRecordKeyMonitorName]
				textRecordTimeStamp := txtMap[rtkMisc.TextRecordKeyTimestamp]
				TextRecordKeyVersion := txtMap[rtkMisc.TextRecordKeyVersion]
				log.Printf("Browse get a Service, mName:[%s] instance:[%s] IP:[%s] ver:[%s] timestamp:[%s], use %d ms", textRecordmonitorName, entry.Instance, lanServerIp, TextRecordKeyVersion, textRecordTimeStamp, time.Now().UnixMilli()-startTime)

				resultChan <- browseParam{entry.Instance, lanServerIp, textRecordmonitorName, TextRecordKeyVersion, 0}
			}
		}
		log.Printf("Stop Browse service instances...")
		close(resultChan)
	})

	err = resolver.Browse(ctx, serviceType, domain, entries)
	if err != nil {
		log.Printf("[%s] Failed to browse:%+v", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_NETWORK_C2S_BROWSER
	}

	log.Printf("Start Browse service instances...")
	return rtkMisc.SUCCESS
}

func browseLanServerAndroid(ctx context.Context, serviceType, domain string, resultChan chan<- browseParam) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(rtkUtils.GetNetInterfaces(), nil)
	if err != nil {
		log.Printf("[%s] Failed to initialize resolver:%+v", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_NETWORK_C2S_RESOLVER
	}

	entries := make(chan *zeroconf.ServiceEntry)

	getTextRecordMap := func(textRecord []string) map[string]string {
		txtMap := make(map[string]string)
		for _, txt := range textRecord {
			parts := strings.SplitN(txt, "=", 2)
			if len(parts) == 2 {
				txtMap[parts[0]] = parts[1]
			}
		}
		return txtMap
	}
	rtkMisc.GoSafe(func() {
		for entry := range entries {
			if len(entry.AddrIPv4) > 0 {
				entryIp := entry.AddrIPv4[0].String()
				lanServerIp := fmt.Sprintf("%s:%d", entryIp, entry.Port)
				log.Printf("Browse get a Service:[%s] IP:[%s], use [%d] ms", entry.Instance, lanServerIp, time.Now().UnixMilli()-startTime)

				txtMap := getTextRecordMap(entry.Text)
				textRecordIp := txtMap[rtkMisc.TextRecordKeyIp]
				textRecordProductName := txtMap[rtkMisc.TextRecordKeyProductName]
				textRecordmName := txtMap[rtkMisc.TextRecordKeyMonitorName]
				textRecordTimeStamp := txtMap[rtkMisc.TextRecordKeyTimestamp]
				textRecordKeyVersion := txtMap[rtkMisc.TextRecordKeyVersion]
				if textRecordIp != entryIp {
					log.Printf("[%s] WARNING: Different IP. Entry:(%s); TextRecord:(%s)", rtkMisc.GetFuncInfo(), entryIp, textRecordIp)
					continue
				}

				if (textRecordProductName != "") && (g_ProductName != "") {
					if textRecordProductName != g_ProductName {
						log.Printf("[%s] WARNING: Different ProductName. Mobile:(%s); TextRecord:(%s)", rtkMisc.GetFuncInfo(), g_ProductName, textRecordProductName)
						continue
					}
				}
				stamp, err := strconv.Atoi(textRecordTimeStamp)
				if err != nil {
					log.Printf("[%s] WARNING: invalid timestamp:%s, err:%+v", rtkMisc.GetFuncInfo(), rtkMisc.TextRecordKeyTimestamp, err)
				}

				log.Printf("Found target Service, mName:[%s] instance:[%s] IP:[%s] ver:[%s] timestamp:[%s], use %d ms", textRecordmName, entry.Instance, lanServerIp, textRecordKeyVersion, textRecordTimeStamp, time.Now().UnixMilli()-startTime)

				if lanServerRunning.Load() {
					rtkPlatform.GoNotifyBrowseResult(textRecordmName, entry.Instance, lanServerIp, textRecordKeyVersion, int64(stamp))
				}
				rtkPlatform.GoNotifyBrowseResult("testMonitorName", "test-Instance", "10.24.136.104:8080", "2.12.14", 0)
				serverInstanceMap.Store("test-Instance", browseParam{
					instance:    "test-Instance",
					ip:          "10.24.136.104:8080",
					monitorName: "testMonitorName",
				})

				resultChan <- browseParam{entry.Instance, lanServerIp, textRecordmName, textRecordKeyVersion, int64(stamp)}
			}
		}
		log.Printf("Stop Browse service instances...")
		close(resultChan)
	})

	err = resolver.Browse(ctx, serviceType, domain, entries)
	if err != nil {
		log.Printf("[%s] Failed to browse:%+v", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_NETWORK_C2S_BROWSER
	}

	log.Printf("Start Browse service instances...")
	return rtkMisc.SUCCESS
}

func browseLanServeriOS(ctx context.Context, serviceType string, resultChan chan<- browseParam) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	log.Printf("Start Browse service instances...")
	rtkPlatform.SetGoBrowseMdnsResultCallback(func(instance, ip string, port int, productName, mName, timestamp, version string) {
		lanServerIp := fmt.Sprintf("%s:%d", ip, port)
		log.Printf("Browse get a Service:[%s] IP:[%s],use [%d] ms", instance, lanServerIp, time.Now().UnixMilli()-startTime)

		stamp, err := strconv.Atoi(timestamp)
		if err != nil {
			log.Printf("[%s] WARNING: invalid[%s]:%d. err:%s", rtkMisc.GetFuncInfo(), rtkMisc.TextRecordKeyTimestamp, stamp, err)
		}

		if lanServerRunning.Load() {
			rtkPlatform.GoNotifyBrowseResult(mName, instance, lanServerIp, version, int64(stamp))
		}
		resultChan <- browseParam{instance, lanServerIp, mName, version, int64(stamp)}
	})
	rtkPlatform.GoStartBrowseMdns("", serviceType)

	rtkMisc.GoSafe(func() {
		<-ctx.Done()
		log.Printf("Stop Browse service instances...")
		rtkPlatform.GoStopBrowseMdns()
		rtkPlatform.SetGoBrowseMdnsResultCallback(nil)
		close(resultChan)
	})

	return rtkMisc.SUCCESS
}

func lookupLanServer(ctx context.Context, instance, serviceType, domain string, bPrintErr bool) (string, rtkMisc.CrossShareErr) {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(rtkUtils.GetNetInterfaces(), nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return "", rtkMisc.ERR_NETWORK_C2S_RESOLVER
	}

	lanServerEntry := make(chan *zeroconf.ServiceEntry)
	if bPrintErr {
		log.Printf("Start Lookup service  by name:%s  type:%s", instance, serviceType)
	}
	err = resolver.Lookup(ctx, instance, serviceType, domain, lanServerEntry)
	if err != nil {
		log.Println("Failed to browse:", err.Error())
		return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP
	}

	getTextRecordMap := func(textRecord []string) map[string]string {
		txtMap := make(map[string]string)
		for _, txt := range textRecord {
			parts := strings.SplitN(txt, "=", 2)
			if len(parts) == 2 {
				txtMap[parts[0]] = parts[1]
			}
		}
		return txtMap
	}
	select {
	case <-ctx.Done():
		return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP_TIMEOUT
	case entry, ok := <-lanServerEntry:
		if !ok {
			break
		}
		txtMap := getTextRecordMap(entry.Text)
		textRecordmonitorName := txtMap[rtkMisc.TextRecordKeyMonitorName]
		textRecordTimeStamp := txtMap[rtkMisc.TextRecordKeyTimestamp]
		textRecordKeyVersion := txtMap[rtkMisc.TextRecordKeyVersion]

		if entry.Instance != instance {
			log.Printf("Expect instance[%s], ignore instance [%s]", instance, entry.Instance)
		} else if len(entry.AddrIPv4) > 0 {
			lanServerIp := fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
			log.Printf("Lookup get Service, mName:[%s] instance:[%s] IP:[%s] ver:[%s] timestamp:[%s], use %d ms", textRecordmonitorName, entry.Instance, lanServerIp, textRecordKeyVersion, textRecordTimeStamp, time.Now().UnixMilli()-startTime)
			stamp, err := strconv.Atoi(textRecordTimeStamp)
			if err != nil {
				log.Printf("[%s] WARNING: invalid[%s]:%d. err:%s", rtkMisc.GetFuncInfo(), rtkMisc.TextRecordKeyTimestamp, stamp, err)
			}
			param := browseParam{entry.Instance, lanServerIp, textRecordmonitorName, textRecordKeyVersion, int64(stamp)}
			serverInstanceMap.Store(param.instance, param)
			return lanServerIp, rtkMisc.SUCCESS
		} else {
			log.Printf("ServiceInstanceName [%s] get AddrIPv4 is null", entry.ServiceInstanceName())
		}
	}
	return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP_INVALID
}

func lookupLanServeriOS(ctx context.Context, instance, serviceType string) (string, rtkMisc.CrossShareErr) {
	startTime := time.Now().UnixMilli()
	log.Printf("Start Lookup service  by name:%s  type:%s", instance, serviceType)
	lanServerEntry := make(chan browseParam)
	rtkPlatform.SetGoBrowseMdnsResultCallback(func(instance, ip string, port int, productName, mName, timestamp, version string) {
		lanServerIp := fmt.Sprintf("%s:%d", ip, port)
		stamp, err := strconv.Atoi(timestamp)
		if err != nil {
			log.Printf("[%s] WARNING: invalid[%s]:%d. err:%s", rtkMisc.GetFuncInfo(), rtkMisc.TextRecordKeyTimestamp, stamp, err)
		}
		lanServerEntry <- browseParam{instance, lanServerIp, mName, version, int64(stamp)}
	})
	rtkPlatform.GoStartBrowseMdns(instance, serviceType)

	select {
	case <-ctx.Done():
		log.Printf("Lookup Timeout, get no entries")
		rtkPlatform.GoStopBrowseMdns()
		rtkPlatform.SetGoBrowseMdnsResultCallback(nil)
		return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP_TIMEOUT
	case val := <-lanServerEntry:
		log.Printf("Lookup get Service success, use [%d] ms", time.Now().UnixMilli()-startTime)
		return val.ip, rtkMisc.SUCCESS
	}
}
