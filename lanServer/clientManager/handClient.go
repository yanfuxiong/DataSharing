package clientManager

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	rtkCommon "rtk-cross-share/lanServer/common"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

const (
	reconnListInternal = 5 * time.Second
)

// =============================
// TimingData get event
// =============================
type NotifyGetTimingDataCallback func() []rtkCommon.TimingData

var notifyGetTimingDataCallback NotifyGetTimingDataCallback

func SetNotifyGetTimingDataCallback(cb NotifyGetTimingDataCallback) {
	notifyGetTimingDataCallback = cb
}

func handleReadFromClientMsg(buffer []byte, IPAddr string, MsgRsp *rtkMisc.C2SMessage) rtkMisc.CrossShareErr {
	if len(buffer) == 0 {
		log.Printf("[%s] buffer is null!", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_BIZ_S2C_READ_EMPTY_DATA
	}
	buffer = bytes.Trim(buffer, "\x00")

	type TempMsg struct {
		ExtData json.RawMessage
		rtkMisc.C2SMessage
	}
	var msg TempMsg
	err := json.Unmarshal(buffer, &msg)
	if err != nil {
		log.Println("Failed to unmarshal C2SMessage data: ", err.Error())
		log.Printf("Err JSON len[%d] data:[%s] ", len(buffer), string(buffer))
		return rtkMisc.ERR_BIZ_JSON_UNMARSHAL
	}

	if msg.MsgType != rtkMisc.C2SMsg_CLIENT_HEARTBEAT {
		log.Printf("Received a msg from clientID:[%s] ClientIndex:[%d] IPAddr:[%s] MsgType:[%s]", msg.ClientID, msg.ClientIndex, IPAddr, msg.MsgType)
	}
	MsgRsp.ClientID = msg.ClientID
	MsgRsp.MsgType = msg.MsgType
	MsgRsp.ClientIndex = msg.ClientIndex
	MsgRsp.TimeStamp = msg.TimeStamp

	switch msg.MsgType {
	case rtkMisc.C2SMsg_CLIENT_HEARTBEAT:
	//rtkdbManager.UpdateHeartBeat(msg.ClientIndex)
	case rtkMisc.C2SMsg_RESET_CLIENT:
		resetRsp := rtkMisc.ResetClientResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS)}
		errCode := rtkdbManager.UpdateClientOnline(int(msg.ClientIndex))
		if errCode != rtkMisc.SUCCESS {
			resetRsp = rtkMisc.ResetClientResponse{Response: rtkMisc.GetResponse(errCode)}
		}
		MsgRsp.ExtData = resetRsp
	case rtkMisc.C2SMsg_INIT_CLIENT:
		var extData rtkMisc.InitClientMessageReq
		initClientRsp := rtkMisc.InitClientMessageResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS), ClientIndex: 0}
		err = json.Unmarshal(msg.ExtData, &extData)
		if err != nil {
			log.Printf("clientID:[%s] decode ExtDataText Err: %s", msg.ClientID, err.Error())
			initClientRsp.Response = rtkMisc.GetResponse(rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL)
		} else {
			pkIndex, errCode := rtkdbManager.UpsertClientInfo(extData.ClientID, extData.HOST, extData.IPAddr, extData.DeviceName, extData.Platform)
			if errCode != rtkMisc.SUCCESS {
				initClientRsp.Response = rtkMisc.GetResponse(errCode)
			} else {
				initClientRsp.ClientIndex = uint32(pkIndex)
				MsgRsp.ClientIndex = uint32(pkIndex)
			}
		}
		MsgRsp.ExtData = initClientRsp
	case rtkMisc.C2SMsg_AUTH_INDEX_MOBILE: //TODO: To remove it  and be replaced by C2SMsg_AUTH_DATA_INDEX_MOBILE
		var extData rtkMisc.AuthIndexMobileReq
		authIndexMobileRsp := rtkMisc.AuthIndexMobileResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS), AuthStatus: false}
		err = json.Unmarshal(msg.ExtData, &extData)
		if err != nil {
			log.Printf("clientID:[%s] decode ExtDataText Err: %s", msg.ClientID, err.Error())
			authIndexMobileRsp.Response = rtkMisc.GetResponse(rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL)
		} else {
			// TODO: Miracast case: Both of Miracast and CrossShare connected to server, then setup authStatus=true and response it
			authStatus := true
			errCode := rtkdbManager.UpdateAuthAndSrcPort(int(msg.ClientIndex), authStatus, extData.SourceAndPort.Source, extData.SourceAndPort.Port)
			if errCode != rtkMisc.SUCCESS {
				authIndexMobileRsp.Response = rtkMisc.GetResponse(errCode)
			} else {
				authIndexMobileRsp.AuthStatus = authStatus
			}
		}
		MsgRsp.ExtData = authIndexMobileRsp
	case rtkMisc.C2SMsg_AUTH_DATA_INDEX_MOBILE:
		authDataIndexMobileRsp := rtkMisc.AuthDataIndexMobileResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS), AuthStatus: false}
		var extData rtkMisc.AuthDataIndexMobileReq
		err = json.Unmarshal(msg.ExtData, &extData)
		if err != nil {
			log.Printf("[%s] clientID:[%s] decode ExtDataText Err: %s", rtkMisc.GetFuncInfo(), msg.ClientID, err.Error())
			authDataIndexMobileRsp.Response = rtkMisc.GetResponse(rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL)
		} else {
			log.Printf("[%s] Width[%d] Height[%d] Type[%d] Framerate[%d]  DisplayName:[%s]", rtkMisc.GetFuncInfo(), extData.AuthData.Width, extData.AuthData.Height, extData.AuthData.Type, extData.AuthData.Framerate, extData.AuthData.DisplayName)
			// TODO: Need to verify  AuthData,  Temporarily write the dead data for now
			authStatus := true
			errCode := rtkdbManager.UpdateAuthAndSrcPort(int(msg.ClientIndex), authStatus, 13, 0)
			if errCode != rtkMisc.SUCCESS {
				authDataIndexMobileRsp.Response = rtkMisc.GetResponse(errCode)
			} else {
				authDataIndexMobileRsp.AuthStatus = authStatus
			}
		}
		MsgRsp.ExtData = authDataIndexMobileRsp
	case rtkMisc.C2SMsg_REQ_CLIENT_LIST:
		getClientListRsp := rtkMisc.GetClientListResponse{Response: rtkMisc.GetResponse(rtkMisc.SUCCESS), ClientList: make([]rtkMisc.ClientInfo, 0)}
		clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
		errCode := rtkdbManager.QueryOnlineClientList(&clientInfoList)
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
			})
		}

		MsgRsp.ExtData = getClientListRsp
	default:
		log.Printf("[%s]Unknown MsgType:[%s]", rtkMisc.GetFuncInfo(), msg.MsgType)
		return rtkMisc.ERR_BIZ_S2C_UNKNOWN_MSG_TYPE
	}

	return rtkMisc.SUCCESS
}

func HandleClient(ctx context.Context, conn net.Conn, timestamp int64) {
	clientIndex := uint32(0)
	clientID := ""

	defer func() {
		if closeConn(clientID, timestamp) {
			rtkdbManager.UpdateClientOffline(int(clientIndex))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("IPAddr:[%s] connect cancel by context! ", conn.RemoteAddr().String())
			return
		default:
			err := conn.SetDeadline(time.Now().Add(time.Duration(rtkMisc.ClientHeartbeatInterval+5) * time.Second))
			if err != nil {
				log.Printf("IPAddr:[%s] connect SetDeadline err:%+v !", conn.RemoteAddr().String(), err)
				return
			}

			// TODO: refine this flow
			buf := bufio.NewReader(conn)
			readStrLine, err := buf.ReadString('\n')
			if err != nil {
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					log.Printf("IPAddr:[%s] ClientIndex:[%d] ReadString timeout: %+v", conn.RemoteAddr().String(), clientIndex, err)
				} else {
					log.Printf("IPAddr:[%s] ClientIndex:[%d] ReadString error:%+v", conn.RemoteAddr().String(), clientIndex, err)
				}
				return
			}
			buffer := make([]byte, 1024)
			buffer = []byte(readStrLine)

			var C2SRsp rtkMisc.C2SMessage
			errCode := handleReadFromClientMsg(buffer, conn.RemoteAddr().String(), &C2SRsp)
			if errCode != rtkMisc.SUCCESS {
				continue
			}

			if clientIndex == 0 {
				clientIndex = C2SRsp.ClientIndex
				clientID = C2SRsp.ClientID
				updateConn(clientID, timestamp, conn)
			}

			if writeMsg(&C2SRsp, timestamp) != rtkMisc.SUCCESS {
				return
			}
		}

	}
}

func writeMsg(msg *rtkMisc.C2SMessage, timestamp int64) rtkMisc.CrossShareErr {
	encodedData, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal C2SMessage data:", err)
		return rtkMisc.ERR_BIZ_JSON_MARSHAL
	}

	if msg.MsgType != rtkMisc.C2SMsg_CLIENT_HEARTBEAT && msg.MsgType != rtkMisc.CS2Msg_RECONN_CLIENT_LIST {
		log.Printf("Write a msg to clientID:[%s] ClientIndex:[%d] MsgType:[%s]", msg.ClientID, msg.ClientIndex, msg.MsgType)
	}
	return write(encodedData, msg.ClientID, timestamp)
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
	reconnListReq := rtkMisc.ReconnClientListReq{ClientList: make([]rtkMisc.ClientInfo, 0), ConnDirect: reconnDirection}
	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	errCode := rtkdbManager.QueryReconnList(&clientInfoList)
	if errCode != rtkMisc.SUCCESS {
		return errCode
	}

	for _, client := range clientInfoList {
		reconnListReq.ClientList = append(reconnListReq.ClientList, rtkMisc.ClientInfo{
			ID:             client.ClientId,
			IpAddr:         client.IpAddr,
			Platform:       client.Platform,
			DeviceName:     client.DeviceName,
			SourcePortType: rtkCommon.GetClientSourcePortType(client.Source, client.Port),
		})
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
