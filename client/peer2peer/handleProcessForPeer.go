package peer2peer

import (
	"context"
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

func StartProcessForPeer(id, ipAddr string) func() {
	ctx, cancel := context.WithCancel(context.Background())
	rtkMisc.GoSafe(func() { ProcessEventsForPeer(id, ipAddr, ctx) })
	log.Printf("[%s][%s][%s] ProcessEventsForPeer is Start !", rtkMisc.GetFuncInfo(), id, ipAddr)
	return cancel
}

func SendDisconnectMsgToPeer(id string) {
	sendCmdMsgToPeer(id, COMM_DISCONNECT, rtkCommon.TEXT_CB, rtkMisc.SUCCESS)
}

func sendCmdMsgToPeer(id string, cmd CommandType, fmtType rtkCommon.TransFmtType, errCode rtkMisc.CrossShareErr) {
	var msg Peer2PeerMessage
	msg.SourceID = rtkGlobal.NodeInfo.ID
	msg.SourcePlatform = rtkGlobal.NodeInfo.Platform
	msg.FmtType = fmtType
	msg.TimeStamp = uint64(time.Now().UnixMilli())
	msg.Command = cmd
	msg.ExtData = errCode
	writeToSocket(&msg, id)
}
