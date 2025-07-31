package login

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

func SetLanServerName(name string) {
	if lanServerName != "" && lanServerName != name {
		lanServerAddr = ""
	}
	lanServerName = name
}

func SetProductName(name string) {
	g_ProductName = name
}

func init() {
	lanServerName = ""
	lanServerAddr = ""
	g_ProductName = ""
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
	authStatus = false
}

func ConnectLanServerRun(ctx context.Context) {
	defer cancelLanServerBusiness()
	stopBrowseInstance()

	time.Sleep(50 * time.Millisecond) // Delay 50ms between "stop browse server" and "start lookup server"

	retryCnt := 0
	for {
		retryCnt++

		if initLanServer() == rtkMisc.SUCCESS {
			goto RunFlag
		}
		if retryCnt == g_retryServerMaxCnt {
			log.Printf("initLanServer %d times failed!  try to lookup Service over again!", retryCnt)
			lanServerAddr = ""
			serverInstanceMap.Delete(lanServerName)
			retryCnt = 0
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(g_retryServerInterval):
		}
	}

RunFlag:
	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformWindows {
		NotifyDIASStatus(DIAS_Status_Checking_Authorization)
	} else {
		NotifyDIASStatus(DIAS_Status_Wait_screenCasting)
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
		case <-heartBeatFlag:
			lanServerHeartbeatStart()
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
		if sendReqMsgToLanServer(rtkMisc.C2SMsg_CLIENT_HEARTBEAT) == rtkMisc.SUCCESS {
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
	}
	cancelBrowse = nil

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
		case param := <-resultChan:
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
		if lanServerName == "" {
			log.Printf("[%s] lanServerName is not set!", rtkMisc.GetFuncInfo())
			return "", rtkMisc.ERR_BIZ_C2S_GET_NO_SERVER_NAME
		}

		mapValue, ok := serverInstanceMap.Load(lanServerName)
		if ok {
			lanServerIp := mapValue.(string)
			if len(lanServerIp) > 0 {
				log.Printf("get LanServer addr from browse serverInstanceMap!")
				return lanServerIp, rtkMisc.SUCCESS
			}
		}
		return lookupLanServer(ctx, lanServerName, rtkMisc.LanServiceType, rtkMisc.LanServerDomain)
	} else {
		lanServerIp := ""
		// find the first service from map
		serverInstanceMap.Range(func(key, value any) bool {
			lanServerName = key.(string)
			lanServerIp = value.(string)
			return false
		})
		if lanServerName != "" && lanServerIp != "" {
			log.Printf("[%s][Mobile] get service name=(%s), ip=(%s) from map", rtkMisc.GetFuncInfo(), lanServerName, lanServerIp)
			return lanServerIp, rtkMisc.SUCCESS
		}

		resultChan := make(chan browseParam)
		var errCode rtkMisc.CrossShareErr
		if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformiOS {
			errCode = browseLanServeriOS(ctx, rtkMisc.LanServiceType, resultChan)
		} else {
			errCode = browseLanServerAndroid(ctx, rtkMisc.LanServiceType, rtkMisc.LanServerDomain, resultChan)
		}
		if errCode != rtkMisc.SUCCESS {
			return "", errCode
		}

		select {
		case <-ctx.Done():
			return "", rtkMisc.ERR_NETWORK_C2S_BROWSER_TIMEOUT
		case param, ok := <-resultChan:
			if !ok {
				break
			}
			if len(param.instance) > 0 && len(param.ip) > 0 {
				serverInstanceMap.Store(param.instance, param.ip)
				return param.ip, rtkMisc.SUCCESS
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
				log.Println("get  lanServerAddr error!\n\n")
				return errCode
			}
			lanServerAddr = serverAddr
			log.Printf("get LanServer addr:%s  by serverName:[%s], try to Dial it!", lanServerAddr, lanServerName)
		}

		pConnectLanServer, err := net.DialTimeout("tcp", lanServerAddr, time.Duration(5*time.Second))
		if err != nil {
			log.Printf("connecting to lanServerAddr[%s] Error:%+v ", lanServerAddr, err.Error())
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					return rtkMisc.ERR_NETWORK_C2S_DIAL_TIMEOUT
				}
			}
			return rtkMisc.ERR_NETWORK_C2S_DIAL
		}
		pSafeConnect.Reset(pConnectLanServer)
		log.Printf("Connect LanServerAddr:[%s] success!", lanServerAddr)
	}
	return rtkMisc.SUCCESS
}

func initLanServer() rtkMisc.CrossShareErr {
	resultCode := connectToLanServer()
	if resultCode != rtkMisc.SUCCESS {
		return resultCode
	}

	return sendReqMsgToLanServer(rtkMisc.C2SMsg_INIT_CLIENT)
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

	retryCnt := 0
	for {
		retryCnt++

		if connectToLanServer() == rtkMisc.SUCCESS {
			goto ReConnectSuccessFlag
		}
		if retryCnt == g_retryServerMaxCnt {
			log.Printf("connect To LanServerAddr:[%s] %d times failed!  try to lookup Service over again!", lanServerAddr, retryCnt)
			lanServerAddr = ""
			serverInstanceMap.Delete(lanServerName)
			retryCnt = 0
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(g_retryServerInterval):
		}
	}

ReConnectSuccessFlag:
	if sendReqMsgToLanServer(rtkMisc.C2SMsg_RESET_CLIENT) == rtkMisc.SUCCESS {
		heartBeatFlag <- struct{}{}
		log.Printf("ReConnectLanServer success!")
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
	NotifyDIASStatus(DIAS_Status_Wait_DiasMonitor)
	rtkPlatform.GoMonitorNameNotify("")
	authStatus = false
	pSafeConnect.Close()
	heartBeatTicker.Reset(time.Duration(999 * time.Hour))
	if disconnectAllClientFunc == nil {
		log.Printf("disconnectAllClientFunc is nil, not cancel all client stream and business!")
		return
	}
	disconnectAllClientFunc()
}

func cancelLanServerBusiness() {
	log.Printf("connect lanServer business is all cancel!")
	NotifyDIASStatus(DIAS_Status_Wait_DiasMonitor)
	rtkPlatform.GoMonitorNameNotify("")
	authStatus = false
	pSafeConnect.Close()
	lanServerAddr = ""
	stopBrowseInstance()
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

func SetAuthStatus(status bool) {
	authStatus = status
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
				SetLanServerName(entry.Instance)
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
