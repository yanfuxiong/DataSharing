package login

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	rtkFileDrop "rtk-cross-share/client/filedrop"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

func SetLanServerName(name string) {
	lanServerName = name
}

func init() {
	lanServerName = ""
	lanServerAddr = ""
	isReconnectRunning.Store(false)

	pSafeConnect = &safeConnect{
		connectMutex:     sync.RWMutex{},
		connectLanServer: nil,
		isAlive:          false,
	}
	heartBeatTicker = nil
	cancelBrowse = nil
}

func ConnectLanServerRun(ctx context.Context) {
	defer cancelLanServerBusiness()
	stopBrowseInstance()

	nCount := 0
	time.Sleep(50 * time.Millisecond) // Delay 50ms between "stop browse server" and "start lookup server"
	if initLanServer() != rtkMisc.SUCCESS {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			nCount++
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if initLanServer() == rtkMisc.SUCCESS {
					ticker.Stop()
					goto RunFlag
				}
				if nCount == 3 {
					log.Printf("connect To LanServerAddr:[%s] %d times failed!  try to lookup Service over again!", lanServerAddr, nCount)
					lanServerAddr = ""
					serverInstanceMap.Delete(lanServerName)
				}
			}
		}
	}

RunFlag:
	if rtkGlobal.NodeInfo.Platform == rtkGlobal.PlatformWindows {
		NotifyDIASStatus(DIAS_Status_Checking_Authorization)
	} else {
		NotifyDIASStatus(DIAS_Status_Wait_screenCasting)
	}

	readResult := make(chan struct {
		buffer  string
		errCode rtkMisc.CrossShareErr
	})

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
		case readData := <-readResult:
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
	if rtkGlobal.NodeInfo.Platform == rtkGlobal.PlatformiOS {
		err = browseLanServeriOS(ctx, rtkMisc.LanServiceType, resultChan)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if rtkGlobal.NodeInfo.Platform == rtkGlobal.PlatformiOS {
		return lookupLanServeriOS(ctx, lanServerName, rtkMisc.LanServiceType)
	} else {
		return lookupLanServer(ctx, lanServerName, rtkMisc.LanServiceType, rtkMisc.LanServerDomain)
	}
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

	nCount := 0
	reTryTicker := time.NewTicker(time.Duration(1) * time.Second)
	defer reTryTicker.Stop()
	for {
		nCount++
		select {
		case <-ctx.Done():
			return
		case <-reTryTicker.C:
			if connectToLanServer() == rtkMisc.SUCCESS {
				goto ReConnectSuccessFlag
			}
			if nCount == 3 {
				log.Printf("connect To LanServerAddr:[%s] %d times failed!  try to lookup Service over again!", lanServerAddr, nCount)
				lanServerAddr = ""
				serverInstanceMap.Delete(lanServerName)
			}
		}
	}

ReConnectSuccessFlag:

	if sendReqMsgToLanServer(rtkMisc.C2SMsg_RESET_CLIENT) == rtkMisc.SUCCESS {
		heartBeatFlag <- struct{}{}
		return
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
	DisconnectLanServerFlag <- struct{}{}
	heartBeatTicker.Reset(time.Duration(999 * time.Hour))
}

func cancelLanServerBusiness() {
	log.Printf("connect lanServer business is all cancel!")
	NotifyDIASStatus(DIAS_Status_Wait_DiasMonitor)
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

// TODO: Callback by windows process
func SendReqClientListToLanServer() {
	// TODO: retry a  few times or keep trying??
	nCount := 0
	for {
		nCount++
		if nCount > 3 {
			log.Printf("[%s] send requst client list 3 times failed!", rtkMisc.GetFuncInfo())
			break
		}

		if sendReqMsgToLanServer(rtkMisc.C2SMsg_REQ_CLIENT_LIST) == rtkMisc.SUCCESS {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func SendReqAuthIndexMobileToLanServer() rtkMisc.CrossShareErr {
	return sendReqMsgToLanServer(rtkMisc.C2SMsg_AUTH_INDEX_MOBILE)
}

func sendReqMsgToLanServer(MsgType rtkMisc.C2SMsgType) rtkMisc.CrossShareErr {
	var msg rtkMisc.C2SMessage
	msg.MsgType = MsgType
	resultCode := buildMessageReq(&msg)
	if resultCode != rtkMisc.SUCCESS {
		return resultCode
	}

	encodedData, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to Marshal C2SMessage data:", err)
		return rtkMisc.ERR_BIZ_JSON_MARSHAL
	}
	encodedData = bytes.Trim(encodedData, "\x00")
	errCode := pSafeConnect.Write(encodedData)
	if errCode != rtkMisc.SUCCESS {
		log.Printf("[%s] LanServer IPAddr:[%s]  sending msg[%s] errCode:%d ", rtkMisc.GetFuncInfo(), pSafeConnect.ConnectIPAddr(), MsgType, errCode)
		pSafeConnect.Close()
		return errCode
	}

	if MsgType != rtkMisc.C2SMsg_CLIENT_HEARTBEAT {
		log.Printf("[%s] MsgType:[%s], Write a message success!", rtkMisc.GetFuncInfo(), MsgType)
	}
	return rtkMisc.SUCCESS
}

func buildMessageReq(msg *rtkMisc.C2SMessage) rtkMisc.CrossShareErr {
	msg.TimeStamp = time.Now().UnixMilli()
	msg.ClientID = rtkGlobal.NodeInfo.ID

	switch msg.MsgType {
	case rtkMisc.C2SMsg_CLIENT_HEARTBEAT:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
	case rtkMisc.C2SMsg_INIT_CLIENT:
		reqData := rtkMisc.InitClientMessageReq{
			HOST:       rtkGlobal.HOST_ID,
			ClientID:   rtkGlobal.NodeInfo.ID,
			Platform:   rtkGlobal.NodeInfo.Platform,
			DeviceName: rtkGlobal.NodeInfo.DeviceName,
			IPAddr:     rtkMisc.ConcatIP(rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort),
		}
		msg.ExtData = reqData

	case rtkMisc.C2SMsg_RESET_CLIENT:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
	case rtkMisc.C2SMsg_AUTH_INDEX_MOBILE:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
		reqData := rtkMisc.AuthIndexMobileReq{
			SourceAndPort: rtkPlatform.GoGetSrcAndPortFromIni(),
		}
		msg.ExtData = reqData
	case rtkMisc.C2SMsg_REQ_CLIENT_LIST:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
	default:
		log.Printf("Unknown MsgType[%s]", msg.MsgType)
		return rtkMisc.ERR_BIZ_C2S_UNKNOWN_MSG_TYPE
	}

	return rtkMisc.SUCCESS
}

func handleReadMessageFromServer(buffer []byte) rtkMisc.CrossShareErr {
	buffer = bytes.Trim(buffer, "\x00")

	if len(buffer) == 0 {
		return rtkMisc.ERR_BIZ_C2S_READ_EMPTY_DATA
	}

	type TempMsg struct {
		ExtData json.RawMessage
		rtkMisc.C2SMessage
	}
	var rspMsg TempMsg
	err := json.Unmarshal(buffer, &rspMsg)
	if err != nil {
		log.Println("Failed to unmarshal C2SMessage data: ", err.Error())
		log.Printf("Err JSON len[%d] data:[%s] ", len(buffer), string(buffer))
		return rtkMisc.ERR_BIZ_JSON_UNMARSHAL
	}

	if rspMsg.MsgType != rtkMisc.C2SMsg_CLIENT_HEARTBEAT && rspMsg.MsgType != rtkMisc.CS2Msg_RECONN_CLIENT_LIST {
		log.Printf("Received a Response msg from Server, clientID:[%s] ClientIndex:[%d] MsgType:[%s] RTT:[%d]ms", rspMsg.ClientID, rspMsg.ClientIndex, rspMsg.MsgType, time.Now().UnixMilli()-rspMsg.TimeStamp)
	}

	switch rspMsg.MsgType {
	case rtkMisc.C2SMsg_CLIENT_HEARTBEAT:
	//log.Printf("HearBeat, RTT:[%d]ms", time.Now().UnixMilli()-rspMsg.TimeStamp)
	case rtkMisc.C2SMsg_RESET_CLIENT:
		var resetClientRsp rtkMisc.ResetClientResponse
		err = json.Unmarshal(rspMsg.ExtData, &resetClientRsp)
		if err != nil {
			log.Printf("clientID:[%s]decode ExtDataText  Err: %+v", rspMsg.ClientID, err)
			return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
		} else {
			if resetClientRsp.Code != rtkMisc.SUCCESS {
				log.Printf("Requst Reset Client failed,  code:[%d] errMsg:[%s]", resetClientRsp.Code, resetClientRsp.Msg)
				return resetClientRsp.Code
			}
		}
		log.Printf("Requst Reset Client Response success, request client list!")
		errCode := sendReqMsgToLanServer(rtkMisc.C2SMsg_REQ_CLIENT_LIST)
		if errCode != rtkMisc.SUCCESS {
			log.Printf("[%s] Err: Send REQ_CLIENT_LIST failed: [%d]", rtkMisc.GetFuncInfo(), errCode)
		}
	case rtkMisc.C2SMsg_INIT_CLIENT:
		var initClientRsp rtkMisc.InitClientMessageResponse
		err = json.Unmarshal(rspMsg.ExtData, &initClientRsp)
		if err != nil {
			log.Printf("clientID:[%s]decode ExtDataText  Err: %+v", rspMsg.ClientID, err)
			return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
		} else {
			if initClientRsp.Code != rtkMisc.SUCCESS {
				log.Printf("Requst Init Client failed,  code:[%d] errMsg:[%s]", initClientRsp.Code, initClientRsp.Msg)
				return initClientRsp.Code
			}
		}
		rtkGlobal.NodeInfo.ClientIndex = initClientRsp.ClientIndex
		heartBeatFlag <- struct{}{}
		log.Printf("Requst Init Client success, get Client Index:[%d]", initClientRsp.ClientIndex)

		if rtkGlobal.NodeInfo.Platform == rtkGlobal.PlatformAndroid || rtkGlobal.NodeInfo.Platform == rtkGlobal.PlatformiOS {
			SendReqAuthIndexMobileToLanServer()
		} else {
			rtkPlatform.GoAuthViaIndex(rtkGlobal.NodeInfo.ClientIndex)
		}
	case rtkMisc.C2SMsg_AUTH_INDEX_MOBILE:
		var authIndexMobileRsp rtkMisc.AuthIndexMobileResponse
		err = json.Unmarshal(rspMsg.ExtData, &authIndexMobileRsp)
		if err != nil {
			log.Printf("clientID:[%s] Index:[%d] Err: decode ExtDataText:%+v", rspMsg.ClientID, rspMsg.ClientIndex, err)
			return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
		}

		if authIndexMobileRsp.Code != rtkMisc.SUCCESS {
			return authIndexMobileRsp.Code
		}

		if authIndexMobileRsp.AuthStatus != true {
			log.Printf("clientID:[%s] Index[%d] Err: Unauthorized", rspMsg.ClientID, rspMsg.ClientIndex)
			return rtkMisc.ERR_BIZ_S2C_UNAUTH
		}
		SendReqClientListToLanServer()
	case rtkMisc.C2SMsg_REQ_CLIENT_LIST:
		var getClientListRsp rtkMisc.GetClientListResponse
		err = json.Unmarshal(rspMsg.ExtData, &getClientListRsp)
		if err != nil {
			log.Printf("clientID:[%s] Index:[%d] Err: decode ExtDataText:%+v", rspMsg.ClientID, rspMsg.ClientIndex, err)
			return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
		}

		if getClientListRsp.Code != rtkMisc.SUCCESS {
			return getClientListRsp.Code
		}
		clientList := make([]rtkMisc.ClientInfo, 0)
		for _, client := range getClientListRsp.ClientList {
			if client.ID != rtkGlobal.NodeInfo.ID {
				clientList = append(clientList, rtkMisc.ClientInfo{
					ID:             client.ID,
					IpAddr:         client.IpAddr,
					Platform:       client.Platform,
					DeviceName:     client.DeviceName,
					SourcePortType: client.SourcePortType,
				})
			} else {
				rtkGlobal.NodeInfo.SourcePortType = client.SourcePortType
			}
		}
		nClientCount := len(clientList)
		if nClientCount == 0 {
			NotifyDIASStatus(DIAS_Status_Wait_Other_Clients)
		} else {
			NotifyDIASStatus(DIAS_Status_Get_Clients_Success)
		}

		GetClientListFlag <- clientList
		log.Printf("Request Client List success, get online ClienList len [%d], self SourcePortType:[%s]", nClientCount, rtkGlobal.NodeInfo.SourcePortType)
	case rtkMisc.C2SMsg_REQ_CLIENT_DRAG_FILE:
		var targetID string
		err = json.Unmarshal(rspMsg.ExtData, &targetID)
		if err != nil {
			log.Printf("clientID:[%s] Index:[%d] Err: decode ExtDataText:%+v", rspMsg.ClientID, rspMsg.ClientIndex, err)
			return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
		}
		log.Printf("Call Client Drag file by LanServer success, get target Client id:[%s]", targetID)
		rtkFileDrop.UpdateDragFileReqDataFromLocal(targetID)
	case rtkMisc.CS2Msg_RECONN_CLIENT_LIST:
		var reconnListReq rtkMisc.ReconnClientListReq
		err = json.Unmarshal(rspMsg.ExtData, &reconnListReq)
		if err != nil {
			log.Printf("clientID:[%s] Index:[%d] Err: decode ExtDataText:%+v", rspMsg.ClientID, rspMsg.ClientIndex, err)
			return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
		}

		clientList := reconnListHandler(reconnListReq.ClientList, reconnListReq.ConnDirect)
		GetClientListFlag <- clientList
	default:
		log.Printf("[%s]Unknown MsgType:[%s]", rtkMisc.GetFuncInfo(), rspMsg.MsgType)
		return rtkMisc.ERR_BIZ_C2S_UNKNOWN_MSG_TYPE
	}

	return rtkMisc.SUCCESS
}

func NotifyDIASStatus(status CrossShareDiasStatus) {
	CurrentDiasStatus = status
	rtkPlatform.GoDIASStatusNotify(uint32(status))
}

func reconnListHandler(reconnList []rtkMisc.ClientInfo, connDirection rtkMisc.ReconnDirection) []rtkMisc.ClientInfo {
	clientList := make([]rtkMisc.ClientInfo, 0)
	for _, client := range reconnList {
		if client.ID == rtkGlobal.NodeInfo.ID {
			continue
		}

		if connDirection == rtkMisc.RECONN_GREATER {
			if rtkGlobal.NodeInfo.ID < client.ID {
				continue
			}
		} else {
			if rtkGlobal.NodeInfo.ID > client.ID {
				continue
			}
		}

		clientList = append(clientList, rtkMisc.ClientInfo{
			ID:             client.ID,
			IpAddr:         client.IpAddr,
			Platform:       client.Platform,
			DeviceName:     client.DeviceName,
			SourcePortType: client.SourcePortType,
		})
	}
	return clientList
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
	case entry := <-lanServerEntry:
		log.Printf("Lookup get Service success, use [%d] ms", time.Now().UnixMilli()-startTime)
		if len(entry.AddrIPv4) > 0 {
			lanServerIp := fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
			return lanServerIp, rtkMisc.SUCCESS
		} else {
			log.Printf("ServiceInstanceName [%s] get AddrIPv4 is null", entry.ServiceInstanceName())
			return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP_INVALID
		}
	}
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

