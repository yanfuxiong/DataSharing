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

	rtkPlatform.SetGoConnectLanServerCallback(func(monitorName, instance, ipAddr string) {
		log.Printf("mobile confirm LanServer monitor name:%s Instance:%s, IpAddr:%d", monitorName, instance, ipAddr)
		serverInstanceMap.Store(instance, ipAddr)
		lanServerInstance = instance
		g_monitorName = monitorName
	})

	rtkPlatform.SetGoBrowseLanServerCallback(func() {
		lanServerInstance = ""
		lanServerAddr = ""
		log.Printf("mobile Browse LanServer monitor triggered!")
		stopLanServerBusiness()
		time.Sleep(50 * time.Millisecond)
		BrowseInstance()
	})
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
		stopBrowseInstance()
	}

	time.Sleep(50 * time.Millisecond) // Delay 50ms between "stop browse server" and "start lookup server"

	retryCnt := 0
	for {
		retryCnt++

		errCode := initLanServer()
		if errCode == rtkMisc.SUCCESS {
			break
		}

		if retryCnt == g_retryServerMaxCnt && (rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac) {
			log.Printf("initLanServer %d times failed, errCode:%d !  try to lookup Service over again!", retryCnt, errCode)
			lanServerAddr = ""
			serverInstanceMap.Delete(lanServerInstance)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(g_retryServerInterval):
		}
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

type browseParam struct {
	instance string
	ip       string
}

func BrowseInstance() rtkMisc.CrossShareErr {
	if cancelBrowse != nil {
		cancelBrowse()
		cancelBrowse = nil
		time.Sleep(50 * time.Millisecond)
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
		select {
		case <-ctx.Done():
			return
		case param, ok := <-resultChan:
			if !ok {
				break
			}
			if len(param.instance) > 0 && len(param.ip) > 0 {
				serverInstanceMap.Store(param.instance, param.ip)
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

func getLanServerAddr() (string, rtkMisc.CrossShareErr) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformMac {
		if lanServerInstance == "" {
			log.Printf("[%s] lanServerInstance is not set!", rtkMisc.GetFuncInfo())
			return "", rtkMisc.ERR_BIZ_C2S_GET_NO_SERVER_NAME
		}

		mapValue, ok := serverInstanceMap.Load(lanServerInstance)
		if ok {
			lanServerIp := mapValue.(string)
			if len(lanServerIp) > 0 {
				log.Printf("get LanServer addr from browse serverInstanceMap!")
				return lanServerIp, rtkMisc.SUCCESS
			}
		}
		return lookupLanServer(ctx, lanServerInstance, rtkMisc.LanServiceType, rtkMisc.LanServerDomain)
	} else {
		if lanServerInstance != "" {
			lanServerIp, ok := serverInstanceMap.Load(lanServerInstance)
			if ok && lanServerIp != "" {
				log.Printf("[%s][Mobile] get service Instance=(%s), ip=(%s) from map", rtkMisc.GetFuncInfo(), lanServerInstance, lanServerIp)
				return lanServerIp.(string), rtkMisc.SUCCESS
			}

		}
	}
	return "", rtkMisc.ERR_NETWORK_C2S_BROWSER_INVALID
}

func connectToLanServer() rtkMisc.CrossShareErr {
	if !pSafeConnect.IsAlive() {
		if lanServerAddr == "" {
			serverAddr, errCode := getLanServerAddr()
			if errCode != rtkMisc.SUCCESS {
				//log.Printf("get lanServerAddr error, err code:%d \n\n", errCode)
				return errCode
			}
			lanServerAddr = serverAddr
		}

		log.Printf("get LanServer addr:%s  by serverInstance:[%s], try to Dial it!", lanServerAddr, lanServerInstance)
		pConnectLanServer, err := net.DialTimeout("tcp", lanServerAddr, time.Duration(5*time.Second))
		if err != nil {
			log.Printf("connecting to lanServerAddr[%s] Error:%+v ", lanServerAddr, err.Error())
			NotifyDIASStatus(DIAS_Status_Connected_LAN_Server_Failed)
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

func initLanServer() rtkMisc.CrossShareErr {
	resultCode := connectToLanServer()
	if resultCode != rtkMisc.SUCCESS {
		return resultCode
	}

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
		heartBeatTicker.Reset(time.Duration(999 * time.Hour))
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

	retryCnt := 0
	for {
		retryCnt++

		if initLanServer() == rtkMisc.SUCCESS {
			log.Printf("ReConnectLanServer success!")
			lanServerHeartbeatStart()
			break
		}
		if retryCnt == g_retryServerMaxCnt {
			NotifyDIASStatus(DIAS_Status_Connectting_DiasService)
			rtkPlatform.GoMonitorNameNotify("")
			log.Printf("connect To LanServerAddr:[%s] %d times failed!  try to lookup Service over again!", lanServerAddr, retryCnt)
			lanServerAddr = ""
			serverInstanceMap.Delete(lanServerInstance)
			lanServerInstance = ""
			if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformAndroid || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformiOS {
				BrowseInstance()
			}
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
	log.Println("lanServerHeartbeatStart is Running!")
}

func StopLanServerRun() {
	stopLanServerBusiness()
}

func stopLanServerBusiness() {
	log.Printf("connect lanServer business is all stop!")
	pSafeConnect.Close()
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
	serverInstanceMap.Clear() // check if need clear
	if isReconnectRunning.Load() {
		if reconnectCancelFunc != nil {
			reconnectCancelFunc()
			reconnectCancelFunc = nil
		}
	}
}

func NotifyDIASStatus(status CrossShareDiasStatus) {
	CurrentDiasStatus = status
	rtkPlatform.GoDIASStatusNotify(uint32(status))
}

func browseLanServer(ctx context.Context, serviceType, domain string, resultChan chan<- browseParam) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(rtkUtils.GetNetInterfaces(), nil)
	if err != nil {
		log.Printf("[%s] Failed to initialize resolver:%+v", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_NETWORK_C2S_RESOLVER
	}

	entries := make(chan *zeroconf.ServiceEntry)
	rtkMisc.GoSafe(func() {
		for entry := range entries {
			if len(entry.AddrIPv4) > 0 {
				lanServerIp := fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
				log.Printf("Browse get a Service:[%s] IP:[%s],use [%d] ms", entry.Instance, lanServerIp, time.Now().UnixMilli()-startTime)
				resultChan <- browseParam{entry.Instance, lanServerIp}
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

				log.Printf("Found target! Service:[%s] IP:[%s], use [%d] ms", entry.Instance, lanServerIp, time.Now().UnixMilli()-startTime)
				// hack code
				//SetLanServerInstance(entry.Instance)
				rtkPlatform.GoNotifyBrowseResult("cross_share_lan_serv", entry.Instance, lanServerIp, "2.2.13", time.Now().UnixMilli())
				rtkPlatform.GoNotifyBrowseResult("monitorName22_test", "entry.Instance", "127.0.0.1:8080", "2.2.10", time.Now().UnixMilli())

				resultChan <- browseParam{entry.Instance, lanServerIp}
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
	rtkPlatform.SetGoBrowseMdnsResultCallback(func(instance, ip string, port int) {
		lanServerIp := fmt.Sprintf("%s:%d", ip, port)
		log.Printf("Browse get a Service:[%s] IP:[%s],use [%d] ms", instance, lanServerIp, time.Now().UnixMilli()-startTime)

		rtkPlatform.GoNotifyBrowseResult("cross_share_lan_serv", instance, lanServerIp, "2.2.13", time.Now().UnixMilli())
		rtkPlatform.GoNotifyBrowseResult("cross_share_lan_test11", "macmacmacmac", "127.0.0.1:8080", "2.2.13", time.Now().UnixMilli())
		resultChan <- browseParam{instance, lanServerIp}
	})
	rtkPlatform.GoStartBrowseMdns("", serviceType)

	rtkMisc.GoSafe(func() {
		<-ctx.Done()
		log.Printf("Stop Browse service instances...")
		rtkPlatform.GoStopBrowseMdns()
		rtkPlatform.SetGoBrowseMdnsResultCallback(nil)
	})

	return rtkMisc.SUCCESS
}

func lookupLanServer(ctx context.Context, instance, serviceType, domain string) (string, rtkMisc.CrossShareErr) {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(rtkUtils.GetNetInterfaces(), nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return "", rtkMisc.ERR_NETWORK_C2S_RESOLVER
	}

	lanServerEntry := make(chan *zeroconf.ServiceEntry)
	log.Printf("Start Lookup service  by name:%s  type:%s", instance, serviceType)
	err = resolver.Lookup(ctx, instance, serviceType, domain, lanServerEntry)
	if err != nil {
		log.Println("Failed to browse:", err.Error())
		return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP
	}

	select {
	case <-ctx.Done():
		log.Printf("Lookup Timeout, get no entries")
		return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP_TIMEOUT
	case entry, ok := <-lanServerEntry:
		if !ok {
			break
		}

		log.Printf("Lookup get Service success, use [%d] ms", time.Now().UnixMilli()-startTime)
		if len(entry.AddrIPv4) > 0 {
			lanServerIp := fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
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
	rtkPlatform.SetGoBrowseMdnsResultCallback(func(instance, ip string, port int) {
		lanServerIp := fmt.Sprintf("%s:%d", ip, port)
		lanServerEntry <- browseParam{instance, lanServerIp}
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
