package peer2peer

import (
	"context"
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkConnection "rtk-cross-share/client/connection"
	rtkFileDrop "rtk-cross-share/client/filedrop"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strconv"
	"time"
)

func init() {
	rtkFileDrop.SetSendFileTransferCancelMsgToPeerCallback(SendFileTransCancelByGuiMsgToPeer)
}

func StartProcessForPeer(ctx context.Context, id, ipAddr string) func(source rtkCommon.CancelBusinessSource) {
	sonCtx, cancel := rtkUtils.WithCancelSource(ctx)
	rtkMisc.GoSafe(func() { ProcessEventsForPeer(sonCtx, id, ipAddr) })
	log.Printf("[%s] ID:[%s] IP:[%s] ProcessEventsForPeer is Start !", rtkMisc.GetFuncInfo(), id, ipAddr)
	return cancel
}

func SendDisconnectMsgToPeer(id string) {
	sendCmdMsgToPeer(id, COMM_DISCONNECT, rtkCommon.TEXT_CB, rtkMisc.SUCCESS)
}

func SendFileTransCancelByGuiMsgToPeer(id, ipAddr string, fileTransDataId uint64, asSrc bool) {
	if asSrc {
		log.Printf("(SRC) [%s] IP:[%s] send cancel filesCachedata msg to dst, id:%d", rtkMisc.GetFuncInfo(), ipAddr, fileTransDataId)
		sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_SRC_INTERRUPT, rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_GUI, fileTransDataId)
		rtkPlatform.GoNotifyErrEvent(id, rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_GUI, ipAddr, strconv.Itoa(int(fileTransDataId)), "", "")
	} else {
		log.Printf("(DST) [%s] IP:[%s] send cancel filesCachedata msg to src, id:%d", rtkMisc.GetFuncInfo(), ipAddr, fileTransDataId)
		sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_DST_INTERRUPT, rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI, fileTransDataId)
		rtkPlatform.GoNotifyErrEvent(id, rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI, ipAddr, strconv.Itoa(int(fileTransDataId)), "", "")
	}
	rtkConnection.CloseFileDropItemStream(id, fileTransDataId)
}

func SendFileTransOpenStreamErrToSrc(id, ipAddr string, fileTransDataId uint64) {
	log.Printf("(DST) [%s] IP:[%s] timestamp:[%d] open file drop Item stream error!", rtkMisc.GetFuncInfo(), ipAddr, fileTransDataId)
	sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_DST_INTERRUPT, rtkMisc.ERR_BIZ_FT_DST_OPEN_STREAM, fileTransDataId)
	rtkPlatform.GoNotifyErrEvent(id, rtkMisc.ERR_BIZ_FT_DST_OPEN_STREAM, ipAddr, strconv.Itoa(int(fileTransDataId)), "", "")
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

func requestFileTransRecoverMsgToSrc(id, srcFileName string, timestamp uint64, offset int64, errCode rtkMisc.CrossShareErr) rtkMisc.CrossShareErr {
	extData := rtkCommon.ExtDataFilesTransferInterruptInfo{
		InterruptSrcFileName: srcFileName,
		InterruptFileOffSet:  offset,
		TimeStamp:            timestamp,
		InterruptErrCode:     errCode,
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
