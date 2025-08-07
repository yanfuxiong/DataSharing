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
	case rtkMisc.C2SMsg_AUTH_DATA_INDEX_MOBILE:
		msg.ClientIndex = rtkGlobal.NodeInfo.ClientIndex
		msg.ExtData = rtkMisc.AuthDataIndexMobileReq{mobileAuthData}
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
		rtkPlatform.GoMonitorNameNotify(monitorName)
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
		monitorName = initClientRsp.MonitorName
		heartBeatFlag <- struct{}{}
		log.Printf("Requst Init Client success, get Client Index:[%d]", initClientRsp.ClientIndex)

		rtkPlatform.GoMonitorNameNotify(monitorName)
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
	case rtkMisc.C2SMsg_AUTH_INDEX_MOBILE: //TODO: To remove it  and be replaced by C2SMsg_AUTH_DATA_INDEX_MOBILE
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
	case rtkMisc.C2SMsg_AUTH_DATA_INDEX_MOBILE:
		var authIndexMobileRsp rtkMisc.AuthDataIndexMobileResponse
		err = json.Unmarshal(rspMsg.ExtData, &authIndexMobileRsp)
		if err != nil {
			log.Printf("clientID:[%s] Index:[%d] Err: decode ExtDataText:%+v", rspMsg.ClientID, rspMsg.ClientIndex, err)
			return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
		}

		if authIndexMobileRsp.Code != rtkMisc.SUCCESS {
			return authIndexMobileRsp.Code
		}

		if authIndexMobileRsp.AuthStatus != true {
			log.Printf("[%s] clientID:[%s] Index[%d] Err: Unauthorized", rtkMisc.GetFuncInfo(), rspMsg.ClientID, rspMsg.ClientIndex)
			return rtkMisc.ERR_BIZ_S2C_UNAUTH
		}
		return SendReqClientListToLanServer()
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
