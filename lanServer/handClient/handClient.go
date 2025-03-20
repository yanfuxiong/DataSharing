package clientManager

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

// TODO:  ERR CODE  HANDLING
func handleReadFromClientMsg(buffer []byte, IPAddr string, MsgRsp *rtkMisc.C2SMessage) error {
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
		return err
	}

	log.Printf("Received a msg from clientID:[%s] ClientIndex:[%d] IPAddr:[%s] MsgType:[%s]", msg.ClientID, msg.ClientIndex, IPAddr, msg.MsgType)
	MsgRsp.ClientID = msg.ClientID
	MsgRsp.MsgType = msg.MsgType
	MsgRsp.ClientIndex = msg.ClientIndex
	MsgRsp.TimeStamp = msg.TimeStamp

	switch msg.MsgType {
	case rtkMisc.C2SMsg_CLIENT_HEARTBEAT:
	//rtkdbManager.UpdateHeartBeat(msg.ClientIndex)
	case rtkMisc.C2SMsg_RESET_CLIENT:
		resetRsp := rtkMisc.ResetClientResponse{Code: rtkMisc.SUCCESS, Msg: ""}
		err = rtkdbManager.OnlineClient(msg.ClientIndex)
		if err != nil {
			resetRsp.Code = -1
			resetRsp.Msg = err.Error()
		}
		MsgRsp.ExtData = resetRsp
	case rtkMisc.C2SMsg_INIT_CLIENT:
		var extData rtkMisc.InitClientMessageReq
		initClientRspss := rtkMisc.InitClientMessageResponse{Code: rtkMisc.SUCCESS}
		err = json.Unmarshal(msg.ExtData, &extData)
		if err != nil {
			log.Printf("clientID:[%s] decode ExtDataText Err: %s", msg.ClientID, err.Error())
			initClientRsp.Code = -1
			initClientRsp.Msg = err.Error()
		} else {
			pkIndex, err := rtkdbManager.UpSertClientInfo(&extData)
			if err != nil {
				initClientRsp.Code = -1
				initClientRsp.Msg = err.Error()
			} else {
				initClientRsp.ClientIndex = pkIndex
				MsgRsp.ClientIndex = pkIndex
			}
		}
		MsgRsp.ExtData = initClientRsp
	case rtkMisc.C2SMsg_REQ_CLIENT_LIST:
		var getClientListRsp rtkMisc.GetClientListResponse
		getClientListRsp.Code = rtkMisc.SUCCESS
		getClientListRsp.ClientList = make([]rtkMisc.ClientInfo, 0)
		err = rtkdbManager.QueryOnlineClientList(&getClientListRsp.ClientList)
		if err != nil {
			getClientListRsp.Code = -1
			getClientListRsp.Msg = err.Error()
		}

		MsgRsp.ExtData = getClientListRsp
	default:
		return errors.New("Unknown MsgType")
	}

	return nil
}

func HandleClient(conn net.Conn) {
	var clientIndex int
	clientIndex = 0
	defer func() {
		rtkdbManager.OfflineClient(clientIndex)
		conn.Close()
	}()

	for {
		err := conn.SetDeadline(time.Now().Add(time.Duration(rtkMisc.ClientHeartbeatInterval+5) * time.Second))
		if err != nil {
			log.Printf("IPAddr:[%s] connect SetDeadline err:%+v  retry after 1s!", conn.RemoteAddr().String(), err)
			time.Sleep(1 * time.Second)
			continue
		}

		buf := bufio.NewReader(conn)
		readStrLine, err := buf.ReadString('\n')
		if err != nil {
			log.Printf("IPAddr:[%s] ClientIndex:[%d] ReadString error:%s", conn.RemoteAddr().String(), clientIndex, err.Error())
			return
		}
		buffer := make([]byte, 1024)
		buffer = []byte(readStrLine)

		var C2SRsp rtkMisc.C2SMessage
		handleReadFromClientMsg(buffer, conn.RemoteAddr().String(), &C2SRsp)
		if clientIndex == 0 {
			clientIndex = C2SRsp.ClientIndex
		}

		encodedData, err := json.Marshal(C2SRsp)
		if err != nil {
			log.Println("Failed to marshal C2SMessage data:", err)
			continue
		}
		encodedData = bytes.Trim(encodedData, "\x00")
		_, err = conn.Write(append(encodedData, '\n'))
		if err != nil {
			log.Printf("IPAddr:[%s] Error sending response:%+v ", conn.RemoteAddr().String(), err.Error())
			return
		}

		err = bufio.NewWriter(conn).Flush()
		if err != nil {
			log.Printf("IPAddr:[%s]  Flush Error:%+v ", conn.RemoteAddr().String(), err.Error())
			return
		}

	}
}
