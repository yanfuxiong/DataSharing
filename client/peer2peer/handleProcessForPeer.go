package peer2peer

import (
	"context"
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkFileDrop "rtk-cross-share/client/filedrop"
	rtkGlobal "rtk-cross-share/client/global"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

func init() {
	rtkFileDrop.SetSendFileTransferCancelMsgToPeerCallback(SendFileTransCancelByGuiMsgToPeer)
}

func StartProcessForPeer(id, ipAddr string) func() {
	ctx, cancel := context.WithCancel(context.Background())
	rtkMisc.GoSafe(func() { ProcessEventsForPeer(id, ipAddr, ctx) })
	log.Printf("[%s][%s][%s] ProcessEventsForPeer is Start !", rtkMisc.GetFuncInfo(), id, ipAddr)
	return cancel
}

func SendDisconnectMsgToPeer(id string) {
	sendCmdMsgToPeer(id, COMM_DISCONNECT, rtkCommon.TEXT_CB, rtkMisc.SUCCESS)
}

func SendFileTransCancelByGuiMsgToPeer(id string, fileTransDataId uint64) {
	log.Printf("[%s] ID:[%s] send cancel filesCachedata msg, id:%d", rtkMisc.GetFuncInfo(), id, fileTransDataId)
	sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_DST_INTERRUPT, rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI, fileTransDataId)
}

func sendFileTransInterruptMsgToPeer(id string, cmd CommandType, errCode rtkMisc.CrossShareErr, timestamp uint64) {
	extData := rtkCommon.ExtDataFilesTransferInterrupt{
		Code:      errCode,
		TimeStamp: timestamp,
	}
	var msg Peer2PeerMessage
	msg.SourceID = rtkGlobal.NodeInfo.ID
	msg.SourcePlatform = rtkGlobal.NodeInfo.Platform
	msg.FmtType = rtkCommon.FILE_DROP
	msg.TimeStamp = uint64(time.Now().UnixMilli())
	msg.Command = cmd
	msg.ExtData = extData
	writeToSocket(&msg, id)
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
