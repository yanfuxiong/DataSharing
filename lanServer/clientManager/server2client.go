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

var VERSION_CONTROL = true

// =============================
// TimingData get event
// =============================
type (
	SendPlatformMsgEventCallback  func(event int, arg1, arg2, arg3, arg4 string)
	SendDragFileListStartCallback func(source, port, horzSize, vertSize, posX, posY int) rtkMisc.CrossShareErr
)

var (
	sendPlatformMsgEventCallback  SendPlatformMsgEventCallback  = nil
	sendDragFileListStartCallback SendDragFileListStartCallback = nil
)

func SetSendPlatformMsgEventCallback(cb SendPlatformMsgEventCallback) {
	sendPlatformMsgEventCallback = cb
}

func SetSendDragFileListStartCallback(cb SendDragFileListStartCallback) {
	sendDragFileListStartCallback = cb
}

func SendDragFileEvent(srcId, targetId string, srcClientIndex uint32) rtkMisc.CrossShareErr {
	msg := rtkMisc.C2SMessage{
		ClientID:    srcId,
		ClientIndex: srcClientIndex,
		MsgType:     rtkMisc.C2SMsg_DRAG_FILE_END,
		TimeStamp:   time.Now().UnixMilli(),
		ExtData:     targetId,
	}

	return writeMsg(&msg, 0)
}

func SendClientPlugEventUpdate(id string, clientIndex uint32, plugEvent bool) rtkMisc.CrossShareErr {
	msg := rtkMisc.C2SMessage{
		ClientID:    id,
		ClientIndex: clientIndex,
		MsgType:     rtkMisc.CS2Msg_UPDATE_PLUG_EVENT,
		TimeStamp:   time.Now().UnixMilli(),
		ExtData: rtkMisc.UpdatePlugEventReq{
			PlugEvent:   plugEvent,
			ProductName: "",
		},
	}

	return writeMsg(&msg, 0)
}

func PeriodicNotifyHandler(ctx context.Context) {
	ticker := time.NewTicker(periodicNotifyInternal)
	defer ticker.Stop()

	connDirectoin := rtkMisc.RECONN_GREATER
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			errCode := buildPeriodicNotify(connDirectoin)
			if errCode != rtkMisc.SUCCESS {
				log.Printf("[%s] Build PeriodicNotify failed: errCode[%d]", rtkMisc.GetFuncInfo(), errCode)
			}

			if connDirectoin == rtkMisc.RECONN_GREATER {
				connDirectoin = rtkMisc.RECONN_LESS
			} else {
				connDirectoin = rtkMisc.RECONN_GREATER
			}
		}
	}
}

func buildPeriodicNotify(reconnDirection rtkMisc.ReconnDirection) rtkMisc.CrossShareErr {
	periodicNotifyReq := rtkMisc.PeriodicNotifyReq{ClientList: make([]rtkMisc.ClientInfo, 0), ConnDirect: reconnDirection, ClientVersion: ""}
	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	errCode := rtkdbManager.QueryReconnList(&clientInfoList)
	if errCode != rtkMisc.SUCCESS {
		return errCode
	}

	nMaxVerValue := int(0)
	for _, client := range clientInfoList {
		periodicNotifyReq.ClientList = append(periodicNotifyReq.ClientList, rtkMisc.ClientInfo{
			ID:             client.ClientId,
			IpAddr:         client.IpAddr,
			Platform:       client.Platform,
			DeviceName:     client.DeviceName,
			SourcePortType: rtkCommon.GetClientSourcePortType(client.Source, client.Port),
		})

		clientVerVal := rtkMisc.GetVersionValue(client.Version)
		if nMaxVerValue < clientVerVal { // get max version
			periodicNotifyReq.ClientVersion = rtkMisc.GetShortVersion(client.Version)
			nMaxVerValue = clientVerVal
		}
	}

	periodicNotifyReq.Scenario = rtkGlobal.Scenario
	retErrCode := rtkMisc.SUCCESS
	for _, client := range periodicNotifyReq.ClientList {
		err := sendPeriodicNotify(client.ID, periodicNotifyReq)
		if err != rtkMisc.SUCCESS {
			retErrCode = err
		}
	}
	//log.Printf("[%s] client count: %d connDirect: %d", rtkMisc.GetFuncInfo(), len(periodicNotifyReq.ClientList), periodicNotifyReq.ConnDirect)
	return retErrCode
}

func sendPeriodicNotify(id string, extData rtkMisc.PeriodicNotifyReq) rtkMisc.CrossShareErr {
	msg := rtkMisc.C2SMessage{
		ClientID:    id,
		ClientIndex: 0,
		MsgType:     rtkMisc.CS2Msg_PERIODIC_NOTIFY,
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

	if extData.Platform == rtkMisc.PlatformMnt {
		errCode = rtkdbManager.UpdateAuthAndSrcPort(pkIndex, true, rtkGlobal.Src_MNT, rtkGlobal.Port_MNT)
		if errCode != rtkMisc.SUCCESS {
			initClientRsp.Response = rtkMisc.GetResponse(errCode)
			return 0, initClientRsp
		}
	}

	initClientRsp.ClientIndex = uint32(pkIndex)
	initClientRsp.Scenario = rtkGlobal.Scenario
	initClientRsp.IsSupportFileDrag = rtkCommon.IsSupportFileDrag()
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
	maxRetryCnt := 20
	captrueMaxRetryCnt := 10
	// Only check capture index in USB-C type
	// It's workaround that checking USB-C port. Remove it after video capture done
	if (extData.AuthData.Type == rtkMisc.DisplayModeUsbC) && (len(rtkCommon.GetUsbCPort()) > 1) {
		if rtkMisc.GetVersionSerialValue(clientInfo.Version) >= int(rtkGlobal.ClientCaptureIndexVerSerial) {
			log.Printf("[%s] Get capture result. Index[%d]", rtkMisc.GetFuncInfo(), clientIndex)
			authStatus, source, port = checkCaptureResult(captrueMaxRetryCnt, int(clientIndex))
			maxRetryCnt -= captrueMaxRetryCnt
		}
	}

	if authStatus == false {
		// Compare timing with TimingData & AuthData
		log.Printf("[%s] Width[%d] Height[%d] Type[%d] Framerate[%d]  DisplayName:[%s]", rtkMisc.GetFuncInfo(), extData.AuthData.Width, extData.AuthData.Height, extData.AuthData.Type, extData.AuthData.Framerate, extData.AuthData.DisplayName)
		authStatus, source, port = checkMobileTiming(maxRetryCnt, int(clientIndex), extData.AuthData)
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
	authDataIndexMobileRsp.Source = source
	authDataIndexMobileRsp.Port = port
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

func dealC2SMsgReqUpdateClientSrcPortInfo(id string, clientIndex uint32, ext *json.RawMessage) interface{} {
	var extData rtkMisc.UpdateClientSrcPortInfoReq
	updateSrcPortInfoRsp := rtkMisc.UpdateClientSrcPortInfoResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS)}
	err := json.Unmarshal(*ext, &extData)
	if err != nil {
		log.Printf("clientID:[%s] decode ExtDataText Err: %s", id, err.Error())
		updateSrcPortInfoRsp.Response = rtkMisc.GetResponse(rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL)
		return updateSrcPortInfoRsp
	}

	errCode := rtkdbManager.UpdateSrcPortInfo(int(clientIndex), extData.ClientIndex, extData.SourcePortInfoList)
	if errCode != rtkMisc.SUCCESS {
		updateSrcPortInfoRsp.Response = rtkMisc.GetResponse(errCode)
	}

	for _, srcPortInfo := range extData.SourcePortInfoList {
		updateSrcPortInfoRsp.SourcePortList = append(updateSrcPortInfoRsp.SourcePortList, srcPortInfo.SourcePort)
	}

	return updateSrcPortInfoRsp
}

func dealC2SMsgReqDragFileListStart(id string, ext *json.RawMessage) interface{} {
	dragFileListStartRsp := rtkMisc.DragFileStartResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS)}
	var dragFileListStartReq rtkMisc.DragFileStartInfo
	err := json.Unmarshal(*ext, &dragFileListStartReq)
	if err != nil {
		log.Printf("clientID:[%s] decode ExtDataText Err: %s", id, err.Error())
		dragFileListStartRsp.Response = rtkMisc.GetResponse(rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL)
		return dragFileListStartRsp
	}

	resizeWidth, resizeHeight, resizeX, resizeY := resizeMobileRect(
		rtkMisc.SourcePort{Source: dragFileListStartReq.Source, Port: dragFileListStartReq.Port},
		dragFileListStartReq.HorzSize,
		dragFileListStartReq.VertSize,
		dragFileListStartReq.PosX,
		dragFileListStartReq.PosY,
	)
	errCode := sendDragFileListStartCallback(
		dragFileListStartReq.Source,
		dragFileListStartReq.Port,
		resizeWidth,
		resizeHeight,
		resizeX,
		resizeY,
	)
	if errCode != rtkMisc.SUCCESS {
		dragFileListStartRsp.Response = rtkMisc.GetResponse(errCode)
	} else {
		dragFileListStartRsp.SourcePort = dragFileListStartReq.SourcePort
	}

	return dragFileListStartRsp
}

func resizeMobileRect(srcPort rtkMisc.SourcePort, width, height, x, y int) (int, int, int, int) {
	timingData := notifyGetTimingDataBySrcPortCallback(srcPort.Source, srcPort.Port)
	if timingData.Width == 0 || timingData.Height == 0 || width == 0 || height == 0 {
		return width, height, x, y
	}

	if timingData.Width == width && timingData.Height == height {
		return width, height, x, y
	}

	retX := x * (timingData.Width / width)
	retY := y * (timingData.Height / height)
	return timingData.Width, timingData.Height, retX, retY
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
