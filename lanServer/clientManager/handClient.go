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
		MsgRsp.ClientIndex, MsgRsp.ExtData = dealC2SMsgInitClient(&msg.ExtData)
	case rtkMisc.C2SMsg_AUTH_DATA_INDEX_MOBILE:
		MsgRsp.ExtData = dealC2SMsgMobileAuthDataIndex(msg.ClientID, msg.ClientIndex, &msg.ExtData)
	case rtkMisc.C2SMsg_REQ_CLIENT_LIST:
		MsgRsp.ExtData = dealC2SMsgReqClientList()
	default:
		log.Printf("[%s]Unknown MsgType:[%s]", rtkMisc.GetFuncInfo(), msg.MsgType)
		return rtkMisc.ERR_BIZ_S2C_UNKNOWN_MSG_TYPE
	}

	return rtkMisc.SUCCESS
}

func checkMobileTiming(clientIndex int, authData rtkMisc.AuthDataInfo) (bool, int, int) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	maxRetryCnt := 20
	retryCnt := 0
	for {
		<-ticker.C
		src, port := getMobileTimingSrcPort(clientIndex, authData)
		if (src > 0) || (port > 0) {
			return true, src, port
		}

		retryCnt++
		log.Printf("[%s] Not found timing(%dx%d@%d), Retry:(%d/%d)",
			rtkMisc.GetFuncInfo(), authData.Width, authData.Height, authData.Framerate, retryCnt, maxRetryCnt)
		if retryCnt >= maxRetryCnt {
			return false, 0, 0
		}
	}
}

func getMobileTimingSrcPort(clientIndex int, authData rtkMisc.AuthDataInfo) (int, int) {
	timingDataList := notifyGetTimingDataCallback()
	source := 0
	port := 0

	clientInfo, err := rtkdbManager.QueryClientInfoByIndex(clientIndex)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s] Err(%d): Get Client info failed:(%d)", rtkMisc.GetFuncInfo(), err, clientIndex)
		return 0, 0
	}

	for idx, timingData := range timingDataList {
		log.Printf("[%s] Timing from DIAS: [%d](%dx%d@%d), Mode:%d",
			rtkMisc.GetFuncInfo(), idx, timingData.Width, timingData.Height, timingData.Framerate, timingData.DisplayMode)
		// Check DisplayMode
		if timingData.DisplayMode != authData.Type {
			continue
		}

		if timingData.DisplayMode == rtkMisc.DisplayModeMiracast {
			// Check DisplayName in MiraCast
			if timingData.DisplayName != authData.DisplayName {
				// DEBUG
				log.Printf("[%s] Different MiraCast displayName: DIAS(%s), Mobile(%s)",
					rtkMisc.GetFuncInfo(), timingData.DisplayName, authData.DisplayName)
				continue
			}

			// TODO: It doesn't match deviceName between phone and DIAS Miracast. Need to fix
			// // Check DeviceName in MiraCast
			// if timingData.DeviceName != clientInfo.DeviceName {
			// 	// DEBUG
			// 	log.Printf("[%s] Different MiraCast deviceName: DIAS(%s), Mobile(%s)",
			// 		rtkMisc.GetFuncInfo(), timingData.DeviceName, clientInfo.DeviceName)
			// 	continue
			// }
		} else if timingData.DisplayMode == rtkMisc.DisplayModeUsbC {
			// FIXME: Always allow timing in iOS platform. We cannot get correct timing in iOS platform now.
			if clientInfo.Platform == rtkMisc.PlatformiOS {
				if timingData.Source == 13 && timingData.Port == 0 && timingData.Width > 0 && timingData.Height > 0 && timingData.Framerate > 0 {
					log.Printf("[%s] iOS special case: Always allow if USB-C timing existed. (%dx%d@%d)(Source,Port)=(%d,%d)",
						rtkMisc.GetFuncInfo(), timingData.Width, timingData.Height, timingData.Framerate, timingData.Source, timingData.Port)
					return timingData.Source, timingData.Port
				}
			}
		}

		// Check Timing(width, height)
		// DO NOT check framerate in MiraCast and USBC, due to Application layer cannot get correct framerate from SurfaceFlinger
		if timingData.Width != authData.Width ||
			timingData.Height != authData.Height {
			log.Printf("[%s] Different resolution: DIAS resolution(%dx%d), Mobile resolution(%dx%d)",
				rtkMisc.GetFuncInfo(), timingData.Width, timingData.Height, authData.Width, authData.Height)
			continue
		}

		source = timingData.Source
		port = timingData.Port
		log.Printf("[%s] Found timing. (%dx%d@%d)(Source,Port)=(%d,%d)",
			rtkMisc.GetFuncInfo(), timingData.Width, timingData.Height, timingData.Framerate, source, port)
		return source, port
	}

	return 0, 0
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
