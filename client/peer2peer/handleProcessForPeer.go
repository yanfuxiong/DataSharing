package peer2peer

import (
	"context"
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkConnection "rtk-cross-share/client/connection"
	rtkFileDrop "rtk-cross-share/client/filedrop"
	rtkGlobal "rtk-cross-share/client/global"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

func init() {
	rtkFileDrop.SetSendFileTransferCancelMsgToPeerCallback(SendFileTransCancelByGuiMsgToPeer)
}

func StartProcessForPeer(ctx context.Context, id, ipAddr string) func(source rtkCommon.CancelBusinessSource) {
	sonCtx, cancel := rtkUtils.WithCancelSource(ctx)
	rtkMisc.GoSafe(func() { ProcessEventsForPeer(sonCtx, id, ipAddr) })
	log.Printf("[%s][%s][%s] ProcessEventsForPeer is Start !", rtkMisc.GetFuncInfo(), id, ipAddr)
	return cancel
}

func SendDisconnectMsgToPeer(id string) {
	sendCmdMsgToPeer(id, COMM_DISCONNECT, rtkCommon.TEXT_CB, rtkMisc.SUCCESS)
}

func SendFileTransCancelByGuiMsgToPeer(id string, fileTransDataId uint64) {
	log.Printf("[%s] ID:[%s] send cancel filesCachedata msg, id:%d", rtkMisc.GetFuncInfo(), id, fileTransDataId)
	sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_DST_INTERRUPT, rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI, fileTransDataId)
	rtkConnection.CloseFileDropItemStream(id, fileTransDataId)
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

func requestFileTransRecoverMsgToSrc(id, fileName string, timestamp uint64, offset, interrupTimeStamp int64, err rtkMisc.CrossShareErr) rtkMisc.CrossShareErr {
	extData := rtkCommon.ExtDataFilesTransferInterruptInfo{
		InterruptSrcFileName:   fileName,
		InterruptFileOffSet:    offset,
		TimeStamp:              timestamp,
		InterruptFileTimeStamp: interrupTimeStamp,
		InterruptErrCode:       err,
	}
	var msg Peer2PeerMessage
	msg.SourceID = rtkGlobal.NodeInfo.ID
	msg.SourcePlatform = rtkGlobal.NodeInfo.Platform
	msg.FmtType = rtkCommon.FILE_DROP
	msg.TimeStamp = uint64(time.Now().UnixMilli())
	msg.Command = COMM_FILE_TRANSFER_RECOVER_REQ
	msg.ExtData = extData
	return writeToSocket(&msg, id)
}

func responseFileTransRecoverMsgToDst(id string, errCode rtkMisc.CrossShareErr) rtkMisc.CrossShareErr {
	var msg Peer2PeerMessage
	msg.SourceID = rtkGlobal.NodeInfo.ID
	msg.SourcePlatform = rtkGlobal.NodeInfo.Platform
	msg.FmtType = rtkCommon.FILE_DROP
	msg.TimeStamp = uint64(time.Now().UnixMilli())
	msg.Command = COMM_FILE_TRANSFER_RECOVER_RSP
	msg.ExtData = errCode
	return writeToSocket(&msg, id)
}
