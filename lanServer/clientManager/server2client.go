package clientManager

import (
	"context"
	"encoding/json"
	"log"
	rtkCommon "rtk-cross-share/lanServer/common"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

var VERSION_CONTROL = false

// =============================
// TimingData get event
// =============================
type SendPlatformMsgEventCallback func(event int, arg1, arg2, arg3, arg4 string)

var sendPlatformMsgEventCallback SendPlatformMsgEventCallback

func SetSendPlatformMsgEventCallback(cb SendPlatformMsgEventCallback) {
	sendPlatformMsgEventCallback = cb
}

// TODO: call by via InterfaceMgr
func SendDragFileEvent(srcId, targetId string, srcClientIndex uint32) rtkMisc.CrossShareErr {
	msg := rtkMisc.C2SMessage{
		ClientID:    srcId,
		ClientIndex: srcClientIndex,
		MsgType:     rtkMisc.C2SMsg_REQ_CLIENT_DRAG_FILE,
		TimeStamp:   time.Now().UnixMilli(),
		ExtData:     targetId,
	}

	return writeMsg(&msg, 0)
}

func ReconnClientListHandler(ctx context.Context) {
	ticker := time.NewTicker(reconnListInternal)
	defer ticker.Stop()

	connDirectoin := rtkMisc.RECONN_GREATER
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			errCode := buildReconnClientList(connDirectoin)
			if errCode != rtkMisc.SUCCESS {
				log.Printf("[%s] Build ReconnClientListReq failed: errCode[%d]", rtkMisc.GetFuncInfo(), errCode)
			}

			if connDirectoin == rtkMisc.RECONN_GREATER {
				connDirectoin = rtkMisc.RECONN_LESS
			} else {
				connDirectoin = rtkMisc.RECONN_GREATER
			}
		}
	}
}

func buildReconnClientList(reconnDirection rtkMisc.ReconnDirection) rtkMisc.CrossShareErr {
	reconnListReq := rtkMisc.ReconnClientListReq{ClientList: make([]rtkMisc.ClientInfo, 0), ConnDirect: reconnDirection, ClientVersion: ""}
	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	errCode := rtkdbManager.QueryReconnList(&clientInfoList)
	if errCode != rtkMisc.SUCCESS {
		return errCode
	}

	nMaxVerValue := int(0)
	for _, client := range clientInfoList {
		reconnListReq.ClientList = append(reconnListReq.ClientList, rtkMisc.ClientInfo{
			ID:             client.ClientId,
			IpAddr:         client.IpAddr,
			Platform:       client.Platform,
			DeviceName:     client.DeviceName,
			SourcePortType: rtkCommon.GetClientSourcePortType(client.Source, client.Port),
		})

		clientVerVal := rtkMisc.GetVersionValue(client.Version)
		if nMaxVerValue < clientVerVal { // get max version
			reconnListReq.ClientVersion = rtkMisc.GetShortVersion(client.Version)
			nMaxVerValue = clientVerVal
		}
	}

	retErrCode := rtkMisc.SUCCESS
	for _, client := range reconnListReq.ClientList {
		err := sendReconnClientList(client.ID, reconnListReq)
		if err != rtkMisc.SUCCESS {
			retErrCode = err
		}
	}
	return retErrCode
}

func sendReconnClientList(id string, extData rtkMisc.ReconnClientListReq) rtkMisc.CrossShareErr {
	msg := rtkMisc.C2SMessage{
		ClientID:    id,
		ClientIndex: 0,
		MsgType:     rtkMisc.CS2Msg_RECONN_CLIENT_LIST,
		TimeStamp:   time.Now().UnixMilli(),
		ExtData:     extData,
	}

	return writeMsg(&msg, 0)
}

func dealC2SMsgInitClient(ext *json.RawMessage) (uint32, interface{}) {
	var extData rtkMisc.InitClientMessageReq
	initClientRsp := rtkMisc.InitClientMessageResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS), ClientIndex: 0, ClientVersion: ""}
	err := json.Unmarshal(*ext, &extData)
	if err != nil {
		log.Printf("clientID:[%s] decode ExtDataText Err: %s", extData.ClientID, err.Error())
		initClientRsp.Response = rtkMisc.GetResponse(rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL)
		return 0, initClientRsp
	}

	if VERSION_CONTROL {
		if !rtkMisc.CheckFullVersionVaild(extData.ClientVersion) {
			initClientRsp.Response = rtkMisc.GetResponse(rtkMisc.ERR_BIZ_VERSION_INVALID)
			log.Printf("clientID:[%s] get invalid version:[%s]", extData.ClientID, extData.ClientVersion)
			return 0, initClientRsp
		}

		reqVerValue := rtkMisc.GetVersionValue(extData.ClientVersion)
		curMaxVersion, errCode := rtkdbManager.QueryMaxVersion()
		if errCode != rtkMisc.SUCCESS {
			initClientRsp.Response = rtkMisc.GetResponse(errCode)
			return 0, initClientRsp
		}

		if curMaxVersion == "" {
			log.Printf("[%s] Always allow the first online client", rtkMisc.GetFuncInfo()) // No online clients in DB, allow current client to connect
		} else {
			curVerValue := rtkMisc.GetVersionValue(curMaxVersion)
			if curVerValue > reqVerValue { // Online clients version > current client version
				log.Printf("[%s] online clients list get max version:%s, and req client version:%s, notify to update!", rtkMisc.GetFuncInfo(), rtkMisc.GetShortVersion(curMaxVersion), extData.ClientVersion)
				initClientRsp.ClientVersion = rtkMisc.GetShortVersion(curMaxVersion)
				return 0, initClientRsp
			} else if curVerValue < reqVerValue { // Online clients version < current client version
				log.Printf("[%s] clientID:[%s] Version:[%s] is newer than current:[%s], notify other client list to update!", rtkMisc.GetFuncInfo(), extData.ClientID, extData.ClientVersion, curMaxVersion)
				buildNotifyClientVersion(extData.ClientID, rtkMisc.GetShortVersion(extData.ClientVersion))
			}
		}
	}

	pkIndex, errCode := rtkdbManager.UpsertClientInfo(extData.ClientID, extData.HOST, extData.IPAddr, extData.DeviceName, extData.Platform, extData.ClientVersion)
	if errCode != rtkMisc.SUCCESS {
		initClientRsp.Response = rtkMisc.GetResponse(errCode)
		return 0, initClientRsp
	}

	errCode = rtkdbManager.UpsertLinkInfo(pkIndex, extData.AppStoreLink)
	if errCode != rtkMisc.SUCCESS {
		initClientRsp.Response = rtkMisc.GetResponse(errCode)
		return 0, initClientRsp
	}

	initClientRsp.ClientIndex = uint32(pkIndex)
	return initClientRsp.ClientIndex, initClientRsp
}

func dealC2SMsgMobileAuthDataIndex(id string, clientIndex uint32, ext *json.RawMessage) interface{} {
	authDataIndexMobileRsp := rtkMisc.AuthDataIndexMobileResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS), AuthStatus: false}
	var extData rtkMisc.AuthDataIndexMobileReq
	err := json.Unmarshal(*ext, &extData)
	if err != nil {
		log.Printf("[%s] clientID:[%s] decode ExtDataText Err: %s", rtkMisc.GetFuncInfo(), id, err.Error())
		authDataIndexMobileRsp.Response = rtkMisc.GetResponse(rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL)
		return authDataIndexMobileRsp
	}

	clientInfo, errQueryClient := rtkdbManager.QueryClientInfoByIndex(int(clientIndex))
	if errQueryClient != rtkMisc.SUCCESS {
		log.Printf("[%s] Err(%d): Get Client info failed:(%d)", rtkMisc.GetFuncInfo(), errQueryClient, clientIndex)
		authDataIndexMobileRsp.Response = rtkMisc.GetResponse(errQueryClient)
		return authDataIndexMobileRsp
	}

	var authStatus = false
	var source = 0
	var port = 0
	// Only check capture index in USB-C type
	if extData.AuthData.Type == rtkMisc.DisplayModeUsbC {
		if rtkMisc.GetVersionSerialValue(clientInfo.Version) >= int(rtkGlobal.ClientCaptureIndexVerSerial) {
			log.Printf("[%s] Get capture result. Index[%d]", rtkMisc.GetFuncInfo(), clientIndex)
			authStatus, source, port = checkCaptureResult(int(clientIndex))
		}
	}

	if authStatus == false {
		// Compare timing with TimingData & AuthData
		log.Printf("[%s] Width[%d] Height[%d] Type[%d] Framerate[%d]  DisplayName:[%s]", rtkMisc.GetFuncInfo(), extData.AuthData.Width, extData.AuthData.Height, extData.AuthData.Type, extData.AuthData.Framerate, extData.AuthData.DisplayName)
		authStatus, source, port = checkMobileTiming(int(clientIndex), extData.AuthData)
	}

	errCode := rtkdbManager.UpdateAuthAndSrcPort(int(clientIndex), authStatus, source, port)
	if errCode != rtkMisc.SUCCESS {
		authDataIndexMobileRsp.Response = rtkMisc.GetResponse(errCode)
		return authDataIndexMobileRsp
	}

	if !authStatus {
		log.Printf("[%s] clientID:[%s] WARNING: Authorize failed", id, rtkMisc.GetFuncInfo())
	}
	authDataIndexMobileRsp.AuthStatus = authStatus
	return authDataIndexMobileRsp
}

func dealC2SMsgReqClientList(clientIndex uint32) interface{} {
	getClientListRsp := rtkMisc.GetClientListResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS), ClientList: make([]rtkMisc.ClientInfo, 0)}
	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	errCode := rtkdbManager.ClientQueryOnlineClientList(int(clientIndex), &clientInfoList)
	if errCode != rtkMisc.SUCCESS {
		getClientListRsp.Response = rtkMisc.GetResponse(errCode)
	}

	for _, client := range clientInfoList {
		getClientListRsp.ClientList = append(getClientListRsp.ClientList, rtkMisc.ClientInfo{
			ID:             client.ClientId,
			IpAddr:         client.IpAddr,
			Platform:       client.Platform,
			DeviceName:     client.DeviceName,
			SourcePortType: rtkCommon.GetClientSourcePortType(client.Source, client.Port),
			Version:        client.Version,
		})
	}

	return getClientListRsp
}

func dealC2SMsgReqPlatformMsgEvent(id string, ext *json.RawMessage) interface{} {
	var extData rtkMisc.PlatformMsgEventReq
	msgEventRsp := rtkMisc.PlatformMsgEventResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS)}
	err := json.Unmarshal(*ext, &extData)
	if err != nil {
		log.Printf("clientID:[%s] decode ExtDataText Err: %s", id, err.Error())
		msgEventRsp.Response = rtkMisc.GetResponse(rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL)
		return msgEventRsp
	}
	log.Printf("[%s] id:[%s] event:[%d], arg1:%s, arg2:%s, arg3:%s, arg4:%s\n", rtkMisc.GetFuncInfo(), id, extData.Event, extData.Arg1, extData.Arg2, extData.Arg3, extData.Arg4)
	sendPlatformMsgEventCallback(int(extData.Event), extData.Arg1, extData.Arg2, extData.Arg3, extData.Arg4)

	return msgEventRsp
}

func buildNotifyClientVersion(id, version string) rtkMisc.CrossShareErr {
	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	errCode := rtkdbManager.QueryOnlineClientList(&clientInfoList)
	if errCode != rtkMisc.SUCCESS {
		return errCode
	}

	reqClientVer := rtkMisc.NotifyClientVersionReq{ClientVersion: version}
	retErrCode := rtkMisc.SUCCESS
	for _, client := range clientInfoList {
		if id == client.ClientId {
			continue
		}
		err := sendNotifyClientVersion(client.ClientId, reqClientVer)
		if err != rtkMisc.SUCCESS {
			retErrCode = err
		}
	}
	return retErrCode
}

func sendNotifyClientVersion(id string, extData rtkMisc.NotifyClientVersionReq) rtkMisc.CrossShareErr {
	msg := rtkMisc.C2SMessage{
		ClientID:    id,
		ClientIndex: 0,
		MsgType:     rtkMisc.CS2Msg_NOTIFY_CLIENT_VERSION,
		TimeStamp:   time.Now().UnixMilli(),
		ExtData:     extData,
	}

	return writeMsg(&msg, 0)
}
