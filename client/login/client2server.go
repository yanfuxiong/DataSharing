package login

import (
	"bytes"
	"encoding/json"
	"log"
	rtkFileDrop "rtk-cross-share/client/filedrop"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

func SendReqClientListToLanServer() rtkMisc.CrossShareErr {
	return sendReqMsgToLanServer(rtkMisc.C2SMsg_REQ_CLIENT_LIST)
}

func SendReqAuthIndexMobileToLanServer() rtkMisc.CrossShareErr {
	return sendReqMsgToLanServer(rtkMisc.C2SMsg_AUTH_INDEX_MOBILE)
}

func sendReqAuthDataAndIndexMobileToLanServer() rtkMisc.CrossShareErr {
	return sendReqMsgToLanServer(rtkMisc.C2SMsg_AUTH_DATA_INDEX_MOBILE)
}

func sendReqInitClientToLanServer() rtkMisc.CrossShareErr {
	return sendReqMsgToLanServer(rtkMisc.C2SMsg_INIT_CLIENT)
}

func sendReqHeartbeatToLanServer() rtkMisc.CrossShareErr {
	return sendReqMsgToLanServer(rtkMisc.C2SMsg_CLIENT_HEARTBEAT)
}

func sendPlatformMsgEventToLanServer(event uint32, arg1, arg2, arg3, arg4 string) {
	extData := rtkMisc.PlatformMsgEventReq{
		Event: event,
		Arg1:  arg1,
		Arg2:  arg2,
		Arg3:  arg3,
		Arg4:  arg4,
	}
	sendReqMsgToLanServer(rtkMisc.CS2Msg_MESSAGE_EVENT, extData)
}

func sendReqMsgToLanServer(MsgType rtkMisc.C2SMsgType, extData ...interface{}) rtkMisc.CrossShareErr {
	var msg rtkMisc.C2SMessage
	msg.MsgType = MsgType
	resultCode := buildMessageReq(&msg, extData...)
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

func buildMessageReq(msg *rtkMisc.C2SMessage, extData ...interface{}) rtkMisc.CrossShareErr {
	msg.TimeStamp = time.Now().UnixMilli()
	msg.ClientID = rtkGlobal.NodeInfo.ID

	switch msg.MsgType {
	case rtkMisc.C2SMsg_CLIENT_HEARTBEAT:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
	case rtkMisc.C2SMsg_INIT_CLIENT:
		reqData := rtkMisc.InitClientMessageReq{
			HOST:          rtkGlobal.HOST_ID,
			ClientID:      rtkGlobal.NodeInfo.ID,
			Platform:      rtkGlobal.NodeInfo.Platform,
			DeviceName:    rtkGlobal.NodeInfo.DeviceName,
			IPAddr:        rtkMisc.ConcatIP(rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort),
			ClientVersion: rtkGlobal.ClientVersion,
			AppStoreLink:  rtkMisc.AppLink,
		}
		msg.ExtData = reqData

	case rtkMisc.C2SMsg_RESET_CLIENT:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
	case rtkMisc.C2SMsg_AUTH_DATA_INDEX_MOBILE:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
		msg.ExtData = rtkMisc.AuthDataIndexMobileReq{mobileAuthData}
	case rtkMisc.C2SMsg_REQ_CLIENT_LIST:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
	case rtkMisc.CS2Msg_MESSAGE_EVENT:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
		if len(extData) < 1 {
			log.Printf("ext data is null!")
			return rtkMisc.ERR_BIZ_C2S_EXT_DATA_EMPTY
		}
		msg.ExtData = extData[0]
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
		return dealS2CMsgInitClient(rspMsg.ClientID, rspMsg.ExtData)
	case rtkMisc.C2SMsg_AUTH_DATA_INDEX_MOBILE:
		return dealS2CMsgMobileAuthDataResp(rspMsg.ClientID, rspMsg.ClientIndex, rspMsg.ExtData)
	case rtkMisc.C2SMsg_REQ_CLIENT_LIST:
		return dealS2CMsgRespClientList(rspMsg.ClientID, rspMsg.ExtData)
	case rtkMisc.C2SMsg_REQ_CLIENT_DRAG_FILE:
		return dealS2CMsgReqClientDragFiles(rspMsg.ClientID, rspMsg.ExtData)
	case rtkMisc.CS2Msg_RECONN_CLIENT_LIST:
		return dealS2CMsgReconnClientList(rspMsg.ClientID, rspMsg.ExtData)
	case rtkMisc.CS2Msg_NOTIFY_CLIENT_VERSION:
		return dealS2CMsgNotifyClientVersion(rspMsg.ClientID, rspMsg.ExtData)
	case rtkMisc.CS2Msg_MESSAGE_EVENT:
		return dealS2CMsgMessageEvent(rspMsg.ClientID, rspMsg.ExtData)
	default:
		log.Printf("[%s]Unknown MsgType:[%s]", rtkMisc.GetFuncInfo(), rspMsg.MsgType)
		return rtkMisc.ERR_BIZ_C2S_UNKNOWN_MSG_TYPE
	}

	return rtkMisc.SUCCESS
}

func dealS2CMsgInitClient(id string, extData json.RawMessage) rtkMisc.CrossShareErr {
	var initClientRsp rtkMisc.InitClientMessageResponse
	err := json.Unmarshal(extData, &initClientRsp)
	if err != nil {
		log.Printf("clientID:[%s]decode ExtDataText  Err: %+v", id, err)
		return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
	}

	if initClientRsp.Code != rtkMisc.SUCCESS {
		log.Printf("Request Init Client failed,  code:[%d] errMsg:[%s]", initClientRsp.Code, initClientRsp.Msg)
		return initClientRsp.Code
	}

	if initClientRsp.ClientVersion != "" {
		log.Printf("Request Init Client, response version:[%s]", initClientRsp.ClientVersion)
		notifyVerValue := rtkMisc.GetVersionValue(initClientRsp.ClientVersion)
		if notifyVerValue < 0 {
			log.Printf("[%s] ClientVersion:[%s]  GetVersionValue failed!", rtkMisc.GetFuncInfo(), initClientRsp.ClientVersion)
			return rtkMisc.ERR_BIZ_VERSION_INVALID
		}

		if checkClientVersionInvalid(initClientRsp.ClientVersion) {
			return rtkMisc.SUCCESS
		}
	}

	rtkGlobal.NodeInfo.ClientIndex = initClientRsp.ClientIndex
	log.Printf("Requst Init Client success, get Client Index:[%d]", initClientRsp.ClientIndex)
	lanServerHeartbeatStart()

	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformAndroid || rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformiOS {
		errCode, authData := rtkPlatform.GetAuthData()
		if errCode != rtkMisc.SUCCESS {
			log.Printf("[%s] GetAuthData errCode:[%d]", rtkMisc.GetFuncInfo(), errCode)
			return errCode
		}
		mobileAuthData = authData
		return sendReqAuthDataAndIndexMobileToLanServer()
	} else {
		rtkPlatform.GoAuthViaIndex(rtkGlobal.NodeInfo.ClientIndex)
	}
	return rtkMisc.SUCCESS
}

func dealS2CMsgMobileAuthDataResp(id string, index uint32, extData json.RawMessage) rtkMisc.CrossShareErr {
	var authIndexMobileRsp rtkMisc.AuthDataIndexMobileResponse
	err := json.Unmarshal(extData, &authIndexMobileRsp)
	if err != nil {
		log.Printf("clientID:[%s] Index:[%d] Err: decode ExtDataText:%+v", id, index, err)
		return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
	}

	if authIndexMobileRsp.Code != rtkMisc.SUCCESS {
		return authIndexMobileRsp.Code
	}

	if authIndexMobileRsp.AuthStatus != true {
		log.Printf("[%s] clientID:[%s] Index[%d] Err: Unauthorized", rtkMisc.GetFuncInfo(), id, index)
		NotifyDIASStatus(DIAS_Status_Authorization_Failed)
		return rtkMisc.ERR_BIZ_S2C_UNAUTH
	}
	return SendReqClientListToLanServer()
}

func dealS2CMsgRespClientList(id string, extData json.RawMessage) rtkMisc.CrossShareErr {
	var getClientListRsp rtkMisc.GetClientListResponse
	err := json.Unmarshal(extData, &getClientListRsp)
	if err != nil {
		log.Printf("clientID:[%s] Err: decode ExtDataText:%+v", id, err)
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
				Version:        client.Version,
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
	return rtkMisc.SUCCESS
}

func dealS2CMsgReqClientDragFiles(id string, extData json.RawMessage) rtkMisc.CrossShareErr {
	var targetID string
	err := json.Unmarshal(extData, &targetID)
	if err != nil {
		log.Printf("clientID:[%s]  Err: decode ExtDataText:%+v", id, err)
		return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
	}
	log.Printf("Call Client Drag file by LanServer success, get target Client id:[%s]", targetID)
	rtkFileDrop.UpdateDragFileReqDataFromLocal(targetID)
	return rtkMisc.SUCCESS
}

func dealS2CMsgReconnClientList(id string, extData json.RawMessage) rtkMisc.CrossShareErr {
	var reconnListReq rtkMisc.ReconnClientListReq
	err := json.Unmarshal(extData, &reconnListReq)
	if err != nil {
		log.Printf("clientID:[%s] Err: decode ExtDataText:%+v", id, err)
		return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
	}

	notifyVerValue := rtkMisc.GetVersionValue(reconnListReq.ClientVersion)
	if notifyVerValue < 0 {
		log.Printf("[%s] ClientVersion:[%s]  GetVersionValue failed!", rtkMisc.GetFuncInfo(), reconnListReq.ClientVersion)
		return rtkMisc.ERR_BIZ_VERSION_INVALID
	}

	if checkClientVersionInvalid(reconnListReq.ClientVersion) {
		return rtkMisc.SUCCESS
	}

	clientList := reconnListHandler(reconnListReq.ClientList, reconnListReq.ConnDirect)
	GetClientListFlag <- clientList
	return rtkMisc.SUCCESS
}

func dealS2CMsgNotifyClientVersion(id string, extData json.RawMessage) rtkMisc.CrossShareErr {
	var notifyVersion rtkMisc.NotifyClientVersionReq
	err := json.Unmarshal(extData, &notifyVersion)
	if err != nil {
		log.Printf("[%s] clientID:[%s]  Err: decode ExtDataText:%+v", rtkMisc.GetFuncInfo(), id, err)
		return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
	}
	notifyVerValue := rtkMisc.GetVersionValue(notifyVersion.ClientVersion)
	if notifyVerValue < 0 {
		log.Printf("[%s] ClientVersion:[%s]  GetVersionValue failed!", rtkMisc.GetFuncInfo(), notifyVersion.ClientVersion)
		return rtkMisc.ERR_BIZ_VERSION_INVALID
	}
	checkClientVersionInvalid(notifyVersion.ClientVersion)
	return rtkMisc.SUCCESS
}

func dealS2CMsgMessageEvent(id string, extData json.RawMessage) rtkMisc.CrossShareErr {
	var messageEventRsp rtkMisc.PlatformMsgEventResponse
	err := json.Unmarshal(extData, &messageEventRsp)
	if err != nil {
		log.Printf("[%s] clientID:[%s]  Err: decode ExtDataText:%+v", rtkMisc.GetFuncInfo(), id, err)
		return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
	}

	if messageEventRsp.Code != rtkMisc.SUCCESS {
		log.Printf("[%s] Message Event Response failed, errCode:[%d]  errMsg:[%s]!", rtkMisc.GetFuncInfo(), messageEventRsp.Code, messageEventRsp.Msg)
		return messageEventRsp.Code
	}
	return rtkMisc.SUCCESS
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

func checkClientVersionInvalid(reqVer string) bool {
	curVerValue := rtkMisc.GetVersionValue(rtkGlobal.ClientVersion)
	if curVerValue < 0 {
		return false
	}

	reqVerValue := rtkMisc.GetVersionValue(reqVer)
	if curVerValue < reqVerValue {
		log.Printf("[%s] Current Client Version:[%s], request Client Version:[%s], need update!", rtkMisc.GetFuncInfo(), rtkGlobal.ClientVersion, reqVer)
		rtkPlatform.GoRequestUpdateClientVersion(rtkMisc.GetShortVersion(reqVer))
		if cancelAllBusinessFunc != nil {
			cancelAllBusinessFunc()
		} else {
			log.Printf("cancelAllBusinessFunc is nil, not cancel all business!")
		}
		return true
	}

	return false
}
