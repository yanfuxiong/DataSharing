package login

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/grandcat/zeroconf"
	"github.com/robfig/cron/v3"
	"log"
	"net"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"time"
)

var (
	lanServerAddr string
	lanServerName string
	pSafeConnect  = &safeConnect{
		connectMutex:     sync.RWMutex{},
		connectLanServer: nil,
		isAlive:          false,
	}

	pCron                  *cron.Cron
	clientInitFlag         = make(chan struct{}, 1)
	cancelLanServerRunFlag = make(chan struct{})

	// Call by connection
	GetClientListFlag       = make(chan struct{})
	ClientListFromLanServer []rtkCommon.ClientInfo
	DisconnectLanServer     = make(chan struct{})
)

func SetLanServerName(name string) {
	lanServerName = name
}

func ConnectLanServerRun(ctx context.Context) {
	pCron = nil
	lanServerAddr = ""

	if !initLanServer() {
		//TODO: notice to GUI  failed, some ERR CODE must be retry
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				cancelLanServerBusiness()
				return
			case <-ticker.C:
				if initLanServer() {
					ticker.Stop()
					break
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			cancelLanServerBusiness()
			return
		case <-clientInitFlag:
			lanServerConnectHeartbeat()
		case <-cancelLanServerRunFlag:
			cancelLanServerBusiness()
			return
		default:
			buffer := make([]byte, 1024)
			conn := pSafeConnect.GetConnect() // TODO: refine this flow
			if conn == nil {
				log.Printf("[%s] LanServer IPAddr:[%s] GetConnect is nil , try to reconnect", rtkMisc.GetFuncInfo(), pSafeConnect.ConnectIPAddr())
				reConnectLanServer()
				time.Sleep(100 * time.Millisecond)
				continue
			}
			buf := bufio.NewReader(conn)
			readStrLine, err := buf.ReadString('\n')
			// _, err = pSafeConnect.Read(&buffer)  //this cause dead lock
			if err != nil {
				log.Printf("[%s] LanServer IPAddr:[%s] ReadString error:%+v, try to reconnect", rtkMisc.GetFuncInfo(), pSafeConnect.ConnectIPAddr(), err.Error())
				reConnectLanServer()
				time.Sleep(100 * time.Millisecond)
				continue
			}
			buffer = []byte(readStrLine)
			handleReadMessageFromServer(buffer)
		}
	}
}

// TODO:  ERR CODE  HANDLING
func discoverLanServer() string {
	if lanServerName == "" {
		log.Println("lanServerName is null")
		return ""
	}

	resolver, err := zeroconf.NewResolver(rtkUtils.GetNetInterfaces(), nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return ""
	}

	var startTime int64
	lanServerEntry := make(chan *zeroconf.ServiceEntry)
	entries := make(chan *zeroconf.ServiceEntry)
	rtkMisc.GoSafe(func() {
		for entry := range entries {
			log.Printf("Found a Service:[%s] Type:[%s],use [%d] ms", entry.Instance, entry.Service, time.Now().UnixMilli()-startTime)
			if entry.Instance == lanServerName {
				lanServerEntry <- entry
				break
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	startTime = time.Now().UnixMilli()
	log.Printf("Start Browse service  by ServiceType:%s", rtkMisc.LanServiceType)
	err = resolver.Browse(ctx, rtkMisc.LanServiceType, rtkMisc.LanServerDomain, entries)
	if err != nil {
		log.Println("Failed to browse:", err.Error())
		return ""
	}

	select {
	case <-ctx.Done():
		log.Printf("Browse Timeout by %s, get no entries", rtkMisc.LanServiceType)
		return ""
	case entry := <-lanServerEntry:
		if len(entry.AddrIPv4) > 0 {
			return fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
		} else {
			log.Printf("ServiceInstanceName [%s] get AddrIPv4 is null", entry.ServiceInstanceName())
			return ""
		}
	}

}

func connectToLanServer() rtkMisc.CrossShareErr {
	if !pSafeConnect.IsAlive() {
		if lanServerAddr == "" {
			lanServerAddr = discoverLanServer()
			if lanServerAddr == "" {
				log.Println("discover LanServer is null")
				return -1
			}
			log.Printf("discovery LanServer addr:%s, try to Dial it!", lanServerAddr)
		}

		pConnectLanServer, err := net.Dial("tcp", lanServerAddr)
		if err != nil {
			log.Printf("connecting to Lan Server[%s] Error:%+v ", lanServerAddr, err.Error())
			return -1
		}
		pSafeConnect.Reset(pConnectLanServer)
		log.Printf("Connect LanServer success!")
	}
	return rtkMisc.SUCCESS
}

func initLanServer() bool {
	if connectToLanServer() != rtkMisc.SUCCESS {
		return false
	}

	if sendReqMsgToLanServer(rtkMisc.C2SMsg_INIT_CLIENT) != rtkMisc.SUCCESS {
		return false
	}
	return true
}

func reConnectLanServer() {
	cancelLanServerBusiness()
	DisconnectLanServer <- struct{}{}

	nCount := 0
	for {
		nCount++
		if connectToLanServer() == rtkMisc.SUCCESS {
			break
		}
		if nCount == 3 {
			lanServerAddr = ""
		}
		time.Sleep(3 * time.Second)
	}

	if sendReqMsgToLanServer(rtkMisc.C2SMsg_RESET_CLIENT) != rtkMisc.SUCCESS {
		time.Sleep(3 * time.Second)
		sendReqMsgToLanServer(rtkMisc.C2SMsg_RESET_CLIENT)
	}

}

func lanServerConnectHeartbeat() {
	pCron = cron.New(cron.WithSeconds())
	specStr := fmt.Sprintf("*/%d * * * * *", rtkMisc.ClientHeartbeatInterval)
	_, err := pCron.AddFunc(specStr, func() {
		// TODO: handle heart beat error,  need  rediscovery
		nCount := 0
		for {
			nCount++
			if sendReqMsgToLanServer(rtkMisc.C2SMsg_CLIENT_HEARTBEAT) == rtkMisc.SUCCESS {
				break
			}

			if nCount >= 3 {
				log.Printf("lanServer heartBeat %d times failed!, try rediscovery", nCount)
				break
			}
			time.Sleep(1 * time.Second)
		}
	})

	// TODO:  ERR CODE  HANDLING
	if err != nil {
		log.Println("lanServerConnectHeartbeat cron AddFunc:", err)
		return
	}
	pCron.Start()
	log.Println("lanServerConnectHeartbeat start!")
}

func CancelLanServerRun() {
	cancelLanServerRunFlag <- struct{}{}
}

func cancelLanServerBusiness() {
	log.Printf("connect lanServer business is all cancel!")
	if pCron != nil {
		pCron.Stop()
		pCron = nil
	}

	if pSafeConnect != nil {
		pSafeConnect.Close()
	}
	ClientListFromLanServer = nil
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

// TODO: hanndle ERR CODE
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
		return -1
	}
	encodedData = bytes.Trim(encodedData, "\x00")
	_, err = pSafeConnect.Write(encodedData)
	if err != nil {
		log.Printf("[%s] LanServer IPAddr:[%s]  sending msg[%s] Err:%+v ", rtkMisc.GetFuncInfo(), pSafeConnect.ConnectIPAddr(), MsgType, err)
		reConnectLanServer()
		return -1
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
	case rtkMisc.C2SMsg_REQ_CLIENT_LIST:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
	default:
		log.Printf("Unknown MsgType[%s]", msg.MsgType)
		return -1
	}

	return rtkMisc.SUCCESS
}

// TODO:  ERR CODE  HANDLING
func handleReadMessageFromServer(buffer []byte) rtkMisc.CrossShareErr {
	buffer = bytes.Trim(buffer, "\x00")

	type TempMsg struct {
		ExtData json.RawMessage
		rtkMisc.C2SMessage
	}
	var rspMsg TempMsg
	err := json.Unmarshal(buffer, &rspMsg)
	if err != nil {
		log.Println("Failed to unmarshal C2SMessage data: ", err.Error())
		log.Printf("Err JSON len[%d] data:[%s] ", len(buffer), string(buffer))
		return -1
	}

	log.Printf("Received a Response msg from Server, clientID:[%s] ClientIndex:[%d] MsgType:[%s] RTT:[%d]ms", rspMsg.ClientID, rspMsg.ClientIndex, rspMsg.MsgType, time.Now().UnixMilli()-rspMsg.TimeStamp)

	switch rspMsg.MsgType {
	case rtkMisc.C2SMsg_CLIENT_HEARTBEAT:
	//log.Printf("HearBeat, RTT:[%d]ms", time.Now().UnixMilli()-rspMsg.TimeStamp)
	case rtkMisc.C2SMsg_RESET_CLIENT:
		var resetClientRsp rtkMisc.ResetClientResponse
		err = json.Unmarshal(rspMsg.ExtData, &resetClientRsp)
		if err != nil {
			log.Printf("clientID:[%s]decode ExtDataText  Err: %+v", rspMsg.ClientID, err)
			return -1
		} else {
			if resetClientRsp.Code != rtkMisc.SUCCESS {
				log.Printf("Requst Reset Client failed,  code:[%d] errMsg:[%s]", resetClientRsp.Code, resetClientRsp.Msg)
				return resetClientRsp.Code
			}
		}
		log.Printf("Requst Rest Client success!")
	case rtkMisc.C2SMsg_INIT_CLIENT:
		var initClientRsp rtkMisc.InitClientMessageResponse
		err = json.Unmarshal(rspMsg.ExtData, &initClientRsp)
		if err != nil {
			log.Printf("clientID:[%s]decode ExtDataText  Err: %+v", rspMsg.ClientID, err)
			return -1
		} else {
			if initClientRsp.Code != rtkMisc.SUCCESS {
				log.Printf("Requst Init Client failed,  code:[%d] errMsg:[%s]", initClientRsp.Code, initClientRsp.Msg)
				return initClientRsp.Code
			}
		}
		rtkGlobal.NodeInfo.ClientIndex = initClientRsp.ClientIndex
		clientInitFlag <- struct{}{}
		log.Printf("Requst Init Client success, get Client Index:[%d]", initClientRsp.ClientIndex)
	case rtkMisc.C2SMsg_REQ_CLIENT_LIST:
		var getClientListRsp rtkMisc.GetClientListResponse
		err = json.Unmarshal(rspMsg.ExtData, &getClientListRsp)
		if err != nil {
			log.Printf("clientID:[%s] Index:[%d] Err: decode ExtDataText:", rspMsg.ClientID, rspMsg.ClientIndex, err)
			return -1
		}

		if getClientListRsp.Code != rtkMisc.SUCCESS {
			return -1
		}
		ClientListFromLanServer = make([]rtkCommon.ClientInfo, 0)
		for _, client := range getClientListRsp.ClientList {
			if client.ID != rtkGlobal.NodeInfo.ID {
				ClientListFromLanServer = append(ClientListFromLanServer, rtkCommon.ClientInfo{
					ID:         client.ID,
					IpAddr:     client.IpAddr,
					Platform:   client.Platform,
					DeviceName: client.DeviceName,
				})
			}
		}

		GetClientListFlag <- struct{}{}
		log.Printf("Requst Client List from LanServer success, get online ClienList len [%d]", len(ClientListFromLanServer))
	default:
		log.Printf("[%s]Unknown MsgType:[%s]", rtkMisc.GetFuncInfo(), rspMsg.MsgType)
		return -1

	}

	return rtkMisc.SUCCESS
}
