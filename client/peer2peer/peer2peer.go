package peer2peer

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	rtkClipboard "rtk-cross-share/client/clipboard"
	rtkCommon "rtk-cross-share/client/common"
	rtkConnection "rtk-cross-share/client/connection"
	rtkFileDrop "rtk-cross-share/client/filedrop"
	rtkGlobal "rtk-cross-share/client/global"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"time"
	"unicode/utf8"
)

func HandleClipboardEvent(ctxMain context.Context, resultChan chan<- EventResult, id string) {
	resultXClipChan := make(chan rtkCommon.ClipBoardData)
	rtkMisc.GoSafe(func() { rtkClipboard.WatchXClipData(ctxMain, id, resultXClipChan) })

	for {
		select {
		case <-ctxMain.Done():
			close(resultChan)
			return
		case cbData := <-resultXClipChan:
			resultChan <- EventResult{
				Cmd: DispatchCmd{
					FmtType: rtkCommon.XCLIP_CB,
					State:   STATE_INIT,
					Command: COMM_SRC,
				},
				Data: cbData,
			}
		}
	}
}

func HandleFileDropEvent(ctxMain context.Context, resultChan chan<- EventResult, id string) {
	resultReqId := make(chan string)
	resultRespId := make(chan string)
	rtkMisc.GoSafe(func() { rtkFileDrop.WatchFileDropReqEvent(ctxMain, id, resultReqId) })
	rtkMisc.GoSafe(func() { rtkFileDrop.WatchFileDropRespEvent(ctxMain, id, resultRespId) })

	for {
		select {
		case <-ctxMain.Done():
			close(resultChan)
			return
		case fileDropId := <-resultReqId:
			if fileDropId == id {
				if data, ok := rtkFileDrop.GetFileDropData(id); ok {
					resultChan <- EventResult{
						Cmd: DispatchCmd{
							FmtType: rtkCommon.FILE_DROP,
							State:   STATE_INIT,
							Command: COMM_SRC,
						},
						Data: rtkCommon.ExtDataFile{
							SrcFileList:   rtkUtils.ClearSrcFileListFullPath(&data.SrcFileList),
							ActionType:    data.ActionType,
							FileType:      rtkCommon.P2PFile_Type_Multiple,
							TimeStamp:     data.TimeStamp,
							FolderList:    data.FolderList,
							TotalDescribe: data.TotalDescribe,
							TotalSize:     data.TotalSize,
						},
					}
				}
			}
		case fileDropId := <-resultRespId:
			if fileDropId == id {
				if data, ok := rtkFileDrop.GetFileDropData(id); ok {
					if data.Cmd == rtkCommon.FILE_DROP_ACCEPT {
						// Accept file: Prepeare to receive data
						resultChan <- EventResult{
							Cmd: DispatchCmd{
								FmtType: rtkCommon.FILE_DROP,
								State:   STATE_IO,
								Command: COMM_DST,
							},
							Data: "",
						}
					} else if data.Cmd == rtkCommon.FILE_DROP_REJECT {
						// Reject file: DO NOT setup SetReadSocketAsMsg
						resultChan <- EventResult{
							Cmd: DispatchCmd{
								FmtType: rtkCommon.FILE_DROP,
								State:   STATE_IO,
								Command: COMM_DST,
							},
							Data: "",
						}
						rtkFileDrop.ResetFileDropData(id)
					} else {
						log.Printf("[%s %d] Invalid fileDrop response data:%s with ID: %s", rtkMisc.GetFuncName(), rtkMisc.GetLine(), data.Cmd, id)
					}
				} else {
					log.Printf("[%s %d] Empty fileDrop response data with ID:%s", rtkMisc.GetFuncName(), rtkMisc.GetLine(), id)
				}
			}
		}
	}
}

func processInbandRead(buffer []byte, len int, msg *Peer2PeerMessage) rtkMisc.CrossShareErr {
	buffer = buffer[:len]
	buffer = bytes.Trim(buffer, "\x00")
	buffer = bytes.Trim(buffer, "\x13")

	type TempMsg struct {
		ExtData json.RawMessage
		Peer2PeerMessage
	}

	var temp TempMsg
	err := json.Unmarshal(buffer, &temp)
	if err != nil {
		log.Println("Failed to unmarshal P2PMessage data", err.Error())
		log.Printf("Err JSON len[%d] data:[%s] ", len, string(buffer))
		//rtkUtils.WriteErrJson(s.RemoteAddr().String(), buffer)
		return rtkMisc.ERR_BIZ_JSON_UNMARSHAL
	}

	*msg = temp.Peer2PeerMessage
	if msg.Command == COMM_DISCONNECT {
		return rtkMisc.SUCCESS
	}

	switch msg.FmtType {
	case rtkCommon.TEXT_CB:
		var extDataText rtkCommon.ExtDataText
		err = json.Unmarshal(temp.ExtData, &extDataText)
		if err != nil {
			log.Println("Err: decode ExtDataText:", err)
			return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
		}

		msg.ExtData = rtkCommon.ExtDataXClip{
			Text:     []byte(extDataText.Text),
			Image:    nil,
			Html:     nil,
			Rtf:      nil,
			TextLen:  int64(utf8.RuneCountInString(extDataText.Text)),
			ImageLen: 0,
			HtmlLen:  0,
			RtfLen:   0,
		}
		msg.FmtType = rtkCommon.XCLIP_CB
	case rtkCommon.FILE_DROP:
		// Response accept or reject
		if msg.State == STATE_TRANS && msg.Command == COMM_DST {
			var extDataCmd rtkCommon.FileDropCmd
			err = json.Unmarshal(temp.ExtData, &extDataCmd)
			if err != nil {
				log.Printf("[%s] Err: decode ExtDataFile:%+v", rtkMisc.GetFuncInfo(), err)
				return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
			}
			msg.ExtData = extDataCmd
		} else if msg.Command == COMM_FILE_TRANSFER_SRC_INTERRUPT || msg.Command == COMM_FILE_TRANSFER_DST_INTERRUPT {
			if rtkUtils.GetPeerClientIsSupportQueueTrans(msg.SourceID) {
				var extData rtkCommon.ExtDataFilesTransferInterrupt
				err = json.Unmarshal(temp.ExtData, &extData)
				if err != nil {
					log.Printf("[%s] Err: decode ExtDataFile:%+v", rtkMisc.GetFuncInfo(), err)
					return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
				}
				msg.ExtData = extData
			} else {
				var resultCode rtkMisc.CrossShareErr
				err = json.Unmarshal(temp.ExtData, &resultCode)
				if err != nil {
					log.Printf("[%s] Err: decode ExtDataFile:%+v", rtkMisc.GetFuncInfo(), err)
					return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
				}
				msg.ExtData = resultCode
			}
		} else if msg.Command == COMM_FILE_TRANSFER_RECOVER_REQ {
			var extData rtkCommon.ExtDataFilesTransferInterruptInfo
			err = json.Unmarshal(temp.ExtData, &extData)
			if err != nil {
				log.Printf("[%s] Err: decode ExtDataFile:%+v", rtkMisc.GetFuncInfo(), err)
				return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
			}
			msg.ExtData = extData
		} else if msg.Command == COMM_FILE_TRANSFER_RECOVER_RSP {
			var resultCode rtkMisc.CrossShareErr
			err = json.Unmarshal(temp.ExtData, &resultCode)
			if err != nil {
				log.Printf("[%s] Err: decode ExtDataFile:%+v", rtkMisc.GetFuncInfo(), err)
				return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
			}
			msg.ExtData = resultCode
		} else {
			var extDataFile rtkCommon.ExtDataFile
			err = json.Unmarshal(temp.ExtData, &extDataFile)
			if err != nil {
				log.Printf("[%s] Err: decode ExtDataFile:%+v", rtkMisc.GetFuncInfo(), err)
				return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
			}
			msg.ExtData = extDataFile
		}
	case rtkCommon.IMAGE_CB:
		if msg.Command == COMM_CB_TRANSFER_SRC_INTERRUPT || msg.Command == COMM_CB_TRANSFER_DST_INTERRUPT {
			var resultCode rtkMisc.CrossShareErr
			err = json.Unmarshal(temp.ExtData, &resultCode)
			if err != nil {
				log.Printf("[%s] Err: decode CrossShareErr:%+v", rtkMisc.GetFuncInfo(), err)
				return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
			}
			msg.ExtData = resultCode
		} else {
			var extDataImg rtkCommon.ExtDataImg
			err = json.Unmarshal(temp.ExtData, &extDataImg)
			if err != nil {
				log.Printf("[%s] Err: decode ExtDataImg:%+v", rtkMisc.GetFuncInfo(), err)
				return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
			}
			msg.ExtData = rtkCommon.ExtDataXClip{
				Text:     nil,
				Image:    nil,
				Html:     nil,
				Rtf:      nil,
				TextLen:  0,
				ImageLen: int64(extDataImg.Size.SizeLow),
				HtmlLen:  0,
				RtfLen:   0,
			}
		}
		msg.FmtType = rtkCommon.XCLIP_CB
	case rtkCommon.XCLIP_CB:
		if msg.Command == COMM_CB_TRANSFER_SRC_INTERRUPT || msg.Command == COMM_CB_TRANSFER_DST_INTERRUPT {
			var resultCode rtkMisc.CrossShareErr
			err = json.Unmarshal(temp.ExtData, &resultCode)
			if err != nil {
				log.Printf("[%s] Err: decode CrossShareErr:%+v", rtkMisc.GetFuncInfo(), err)
				return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
			}
			msg.ExtData = resultCode
		} else {
			var extDataXClip rtkCommon.ExtDataXClip
			err = json.Unmarshal(temp.ExtData, &extDataXClip)
			if err != nil {
				log.Println("Err: decode ExtDataXClip:", err)
				return rtkMisc.ERR_BIZ_JSON_EXTDATA_UNMARSHAL
			}
			msg.ExtData = extDataXClip
		}
	}
	return rtkMisc.SUCCESS
}

func HandleReadInbandFromSocket(ctxMain context.Context, resultChan chan<- EventResult, id, ipAddr string) {
	defer close(resultChan)
	for {
		select {
		case <-ctxMain.Done():
			log.Printf("[%s][Socket] ID:[%s] Err: Read operation is done by context", rtkMisc.GetFuncInfo(), id)
			return
		default:
			buffer := make([]byte, rtkGlobal.P2PMsgMaxLength) // 32KB
			nLen, errCode := rtkConnection.ReadSocket(id, buffer)
			if errCode == rtkMisc.ERR_BIZ_GET_STREAM_RESET {
				continue
			} else if errCode == rtkMisc.ERR_BIZ_GET_STREAM_EMPTY {
				log.Printf("[%s] ID:[%s] ReadSocket failed  errCode:[%d]", rtkMisc.GetFuncInfo(), id, errCode)
				time.Sleep(10 * time.Millisecond)
				continue
			} else {
				if errCode != rtkMisc.SUCCESS {
					log.Printf("[%s] ID:[%s] ReadSocket failed  errCode:[%d]", rtkMisc.GetFuncInfo(), id, errCode)
					continue
				}
			}

			var msg Peer2PeerMessage
			errCode = processInbandRead(buffer, nLen, &msg)
			if errCode != rtkMisc.SUCCESS {
				log.Printf("[%s] handle Read message error, errCode:%d, retrying...", rtkMisc.GetFuncInfo(), errCode)
				continue
			}
			log.Printf("[%s] EventResult fmt=%s, state=%s, cmd=%s", rtkMisc.GetFuncInfo(), msg.FmtType, msg.State, msg.Command)

			if msg.Command == COMM_DISCONNECT {
				//rtkConnection.OfflineEvent(id)
				continue
			} else if msg.Command == COMM_FILE_TRANSFER_SRC_INTERRUPT {
				if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
					if fileTransInfo, ok := msg.ExtData.(rtkCommon.ExtDataFilesTransferInterrupt); ok {
						CancelDstFileTransfer(id, ipAddr, fileTransInfo.TimeStamp, fileTransInfo.Code) // Dst: Copy need cancel
					}
				} else {
					if fileDropData, ok := rtkFileDrop.GetFileDropData(id); ok {
						if fileTransCode, extOk := msg.ExtData.(rtkMisc.CrossShareErr); extOk {
							CancelDstFileTransfer(id, ipAddr, fileDropData.TimeStamp, fileTransCode) // Dst: Copy need cancel
						}
					}
				}
				continue
			} else if msg.Command == COMM_FILE_TRANSFER_DST_INTERRUPT {
				if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
					if fileTransInfo, ok := msg.ExtData.(rtkCommon.ExtDataFilesTransferInterrupt); ok {
						CancelSrcFileTransfer(id, ipAddr, fileTransInfo.TimeStamp, fileTransInfo.Code) // Src: Copy need cancel
					}
				} else {
					if fileDropData, ok := rtkFileDrop.GetFileDropData(id); ok {
						if fileTransCode, extOk := msg.ExtData.(rtkMisc.CrossShareErr); extOk {
							CancelSrcFileTransfer(id, ipAddr, fileDropData.TimeStamp, fileTransCode) // Src: Copy need cancel
						}
					}
				}
				continue
			} else if msg.Command == COMM_FILE_TRANSFER_RECOVER_REQ { // Src
				if interruptInfo, ok := msg.ExtData.(rtkCommon.ExtDataFilesTransferInterruptInfo); ok {
					reqErrCode := rtkMisc.SUCCESS
					if !rtkFileDrop.SetFilesTransferDataInterrupt(id, interruptInfo.InterruptSrcFileName, "", "", interruptInfo.TimeStamp, interruptInfo.InterruptFileOffSet, interruptInfo.InterruptErrCode) {
						reqErrCode = rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
					}
					rtkMisc.GoSafe(func() { recoverFileTransferProcessAsSrc(ctxMain, id, ipAddr, reqErrCode) })
				}
				continue
			} else if msg.Command == COMM_FILE_TRANSFER_RECOVER_RSP { // Dst
				if recoverRspCode, ok := msg.ExtData.(rtkMisc.CrossShareErr); ok {
					if recoverRspCode == rtkMisc.SUCCESS {
						rtkMisc.GoSafe(func() { recoverFileTransferProcessAsDst(ctxMain, id, ipAddr) })
					} else {
						log.Printf("[%s](DST) request recover file transfer failed, errCode:[%d]", rtkMisc.GetFuncInfo(), recoverRspCode)
						clearCacheFileDataList(id, ipAddr, recoverRspCode)
					}
				}
				continue
			} else if msg.Command == COMM_CB_TRANSFER_SRC_INTERRUPT {
				log.Printf("[%s] (DST) Copy image operation was canceled by src !", rtkMisc.GetFuncInfo())
				continue
			} else if msg.Command == COMM_CB_TRANSFER_DST_INTERRUPT {
				log.Printf("[%s] (SRC) Copy image operation was canceled by dst !", rtkMisc.GetFuncInfo())
				continue
			}

			resultChan <- EventResult{
				Cmd: DispatchCmd{
					FmtType: msg.FmtType,
					State:   msg.State,
					Command: msg.Command,
				},
				Data: msg.ExtData,
			}
		}
	}
}

func buildCmd(curState *StateType, curCommand *CommandType, event EventResult, buildState *StateType, buildCommand *CommandType) bool {
	ret := true
	if event.Cmd.State == STATE_INIT {
		*buildState = STATE_INFO
		*buildCommand = COMM_SRC
	} else if event.Cmd.State == STATE_INFO {
		if event.Cmd.Command == COMM_SRC {
			*buildState = STATE_INFO
			*buildCommand = COMM_DST
		} else if event.Cmd.Command == COMM_DST {
			*buildState = STATE_TRANS
			*buildCommand = COMM_SRC
		} else {
			ret = false
		}
	} else if event.Cmd.State == STATE_TRANS {
		if event.Cmd.Command == COMM_SRC {
			*buildState = STATE_TRANS
			*buildCommand = COMM_DST
		} else if event.Cmd.Command == COMM_DST {
			// Start to transfer file
			*buildState = STATE_IO
			*buildCommand = COMM_SRC
		} else {
			ret = false
		}
	} else if event.Cmd.State == STATE_IO {
		// Only from require file
		if event.Cmd.Command == COMM_DST {
			*buildState = STATE_IO
			*buildCommand = COMM_DST
		} else {
			ret = false
		}
	} else {
		ret = false
	}

	if ret && isValidState(*curState, *curCommand, *buildState, *buildCommand) {
		updateStateCommand(curState, curCommand, *buildState, *buildCommand)
		return true
	} else {
		log.Printf("[%s %d] Invalid state: cur(%s, %s), next(%s, %s)", rtkMisc.GetFuncName(), rtkMisc.GetLine(), *curState, *curCommand, event.Cmd.State, event.Cmd.Command)
		return false
	}
}

func buildMessage(msg *Peer2PeerMessage, id string, event EventResult) bool {
	msg.SourceID = rtkGlobal.NodeInfo.ID
	msg.SourcePlatform = rtkGlobal.NodeInfo.Platform
	msg.FmtType = event.Cmd.FmtType
	msg.State = event.Cmd.State
	msg.Command = event.Cmd.Command
	msg.TimeStamp = uint64(time.Now().UnixMilli())

	if !rtkUtils.GetPeerClientIsSupportXClip(id) && event.Cmd.FmtType == rtkCommon.XCLIP_CB {
		if event.Cmd.Command == COMM_SRC {
			if extData, ok := rtkClipboard.GetLastClipboardData().ExtData.(rtkCommon.ExtDataXClip); ok {
				if extData.ImageLen > 0 {
					msg.FmtType = rtkCommon.IMAGE_CB
				} else if extData.TextLen > 0 {
					msg.FmtType = rtkCommon.TEXT_CB
				}
			} else {
				log.Printf("[%s] Err: Import Clipboard data failed", rtkMisc.GetFuncInfo())
				return false
			}
		} else if event.Cmd.Command == COMM_DST { // only image have COMM_DST msg
			msg.FmtType = rtkCommon.IMAGE_CB
		}
	}

	switch msg.FmtType {
	case rtkCommon.TEXT_CB:
		if extData, ok := rtkClipboard.GetLastClipboardData().ExtData.(rtkCommon.ExtDataXClip); ok {
			msg.ExtData = rtkCommon.ExtDataText{
				Text: string(extData.Text),
			}
			return true
		} else {
			log.Printf("[%s] Err: Import ExtDataXClip - Text to msg failed", rtkMisc.GetFuncInfo())
			return false
		}
	case rtkCommon.IMAGE_CB:
		// Paste image and require data
		if event.Cmd.State == STATE_IO && event.Cmd.Command == COMM_DST {
			msg.State = STATE_TRANS
			msg.Command = COMM_DST
			return true
		} else if event.Cmd.Command == COMM_SRC {
			if extData, ok := rtkClipboard.GetLastClipboardData().ExtData.(rtkCommon.ExtDataXClip); ok {
				_, width, height := rtkUtils.GetByteImageInfo(extData.Image)
				msg.ExtData = rtkCommon.ExtDataImg{
					Size: rtkCommon.FileSize{
						SizeHigh: uint32(0),
						SizeLow:  uint32(extData.ImageLen),
					},
					Header: rtkCommon.ImgHeader{
						Width:       int32(width),
						Height:      int32(height),
						Planes:      1,
						BitCount:    32,
						Compression: 0,
					},
				}
				return true
			} else {
				log.Printf("[%s %d] Err: Import Clipboard - Image to msg failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
				return false
			}
		} else if event.Cmd.Command == COMM_DST {
			return true
		}

	case rtkCommon.FILE_DROP:
		// Accept file and require data
		if event.Cmd.State == STATE_IO && event.Cmd.Command == COMM_DST {
			msg.State = STATE_TRANS
			msg.Command = COMM_DST
			if extData, ok := rtkFileDrop.GetFileDropData(id); ok {
				msg.ExtData = extData.Cmd
				return true
			}
		} else if event.Cmd.Command == COMM_SRC {
			if extData, ok := rtkFileDrop.GetFileDropData(id); ok {
				msg.ExtData = rtkCommon.ExtDataFile{
					SrcFileList:   rtkUtils.ClearSrcFileListFullPath(&extData.SrcFileList),
					ActionType:    extData.ActionType,
					FileType:      rtkCommon.P2PFile_Type_Multiple,
					TimeStamp:     extData.TimeStamp,
					FolderList:    extData.FolderList,
					TotalDescribe: extData.TotalDescribe,
					TotalSize:     extData.TotalSize,
				}
				return true
			} else {
				log.Printf("[%s %d] Err: Import FileDrop - File to msg failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
				return false
			}
		} else if event.Cmd.Command == COMM_DST {
			return true
		}
	case rtkCommon.XCLIP_CB:
		// Accept XClip and require data
		if event.Cmd.State == STATE_IO && event.Cmd.Command == COMM_DST {
			msg.State = STATE_TRANS
			msg.Command = COMM_DST
			return true
		} else if event.Cmd.Command == COMM_SRC {
			if extData, ok := rtkClipboard.GetLastClipboardData().ExtData.(rtkCommon.ExtDataXClip); ok {
				msg.ExtData = rtkCommon.ExtDataXClip{
					Text:     nil,
					Image:    nil,
					Html:     nil,
					Rtf:      nil,
					TextLen:  extData.TextLen,
					ImageLen: extData.ImageLen,
					HtmlLen:  extData.HtmlLen,
					RtfLen:   extData.RtfLen,
				}
				return true
			} else {
				log.Printf("[%s %d] Err: Import Clipboard - data to msg failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
				return false
			}
		} else if event.Cmd.Command == COMM_DST {
			return true
		}
	}

	return false
}

func writeToSocket(msg *Peer2PeerMessage, id string) rtkMisc.CrossShareErr {
	log.Printf("Write msg to: %s, platform=%s, fmt=%s, state=%s, cmd=%s", id, msg.SourcePlatform, msg.FmtType, msg.State, msg.Command)
	encodedData, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal P2PMessage data:", err)
		return rtkMisc.ERR_BIZ_JSON_MARSHAL
	}

	encodedData = bytes.Trim(encodedData, "\x00")
	encodedData = bytes.Trim(encodedData, "\x13")

	errCode := rtkConnection.WriteSocket(id, encodedData)
	if errCode == rtkMisc.ERR_BIZ_GET_STREAM_RESET {
		return rtkConnection.WriteSocket(id, encodedData)
	}
	return errCode
}

func processIoWrite(ctx context.Context, id, ipAddr string, fmtType rtkCommon.TransFmtType, timeStamp uint64) {
	resultCode := rtkMisc.SUCCESS
	if fmtType == rtkCommon.FILE_DROP {
		dealFilesCacheDataProcess(ctx, id, ipAddr, timeStamp)
	} else if fmtType == rtkCommon.XCLIP_CB {
		resultCode = writeXClipDataToSocket(id, ipAddr)
		if resultCode != rtkMisc.SUCCESS {
			log.Printf("(SRC) ID[%s] IP[%s] Copy XClip data To Socket failed, ERR code:[%d]", id, ipAddr, resultCode)
			sendCmdMsgToPeer(id, COMM_CB_TRANSFER_SRC_INTERRUPT, fmtType, resultCode)
		}
	} else {
		log.Printf("[%s]Unknown FmtType:[%s]", rtkMisc.GetFuncInfo(), fmtType)
	}
}

func processIoRead(ctx context.Context, id, ipAddr string, fmtType rtkCommon.TransFmtType, timeStamp uint64) {
	resultCode := rtkMisc.SUCCESS
	if fmtType == rtkCommon.FILE_DROP {
		dealFilesCacheDataProcess(ctx, id, ipAddr, timeStamp)
	} else if fmtType == rtkCommon.XCLIP_CB {
		resultCode = handleXClipDataFromSocket(id, ipAddr)
		if resultCode != rtkMisc.SUCCESS {
			log.Printf("(DST) ID[%s] IP[%s] Copy XClip data From Socket failed, ERR code:[%d] ...", id, ipAddr, resultCode)
			sendCmdMsgToPeer(id, COMM_CB_TRANSFER_DST_INTERRUPT, fmtType, resultCode)
		}
	} else {
		log.Printf("[%s]Unknown FmtType:[%s]", rtkMisc.GetFuncInfo(), fmtType)
	}
}

func updateStateCommand(curState *StateType, curCommand *CommandType, updateState StateType, updateCommand CommandType) {
	log.Printf("[%s %d] Current state from (%s, %s) to (%s, %s)", rtkMisc.GetFuncName(), rtkMisc.GetLine(), *curState, *curCommand, updateState, updateCommand)
	if updateState == STATE_INIT {
		// convert STATE_INIT to STATE_INFO for INIT
		*curState = STATE_INFO
	} else {
		*curState = updateState
	}
	*curCommand = updateCommand
}

func processFileDrop(ctx context.Context, id, ipAddr string, event EventResult) bool {
	nextState := event.Cmd.State
	nextCommand := event.Cmd.Command

	if nextState == STATE_IO && nextCommand == COMM_SRC {
		// Receive response from dst
		if extData, ok := event.Data.(rtkCommon.FileDropCmd); ok {
			if extData == rtkCommon.FILE_DROP_ACCEPT {
				timeStamp := rtkFileDrop.SetFilesDataToCacheAsSrc(id)
				if rtkFileDrop.GetFilesTransferDataCacheCount(id) <= rtkGlobal.FilesConcurrentTransferMaxSize {
					rtkMisc.GoSafe(func() { processIoWrite(ctx, id, ipAddr, event.Cmd.FmtType, timeStamp) }) // [Src]: Start to trans file
				} else {
					log.Printf("[%s] ID:[%s] there are file data transfer is in progress, queue up and wait!", rtkMisc.GetFuncInfo(), id)
				}
			} else if extData == rtkCommon.FILE_DROP_REJECT {
				// TODO: send response to platform (accept or reject)
				rtkFileDrop.ResetFileDropData(id)
			} else {
				log.Printf("[%s %d] Unknown file drop response type: %s", rtkMisc.GetFuncName(), rtkMisc.GetLine(), extData)
			}
		} else {
			log.Printf("[%s %d] Invalid file drop response: %+v", rtkMisc.GetFuncName(), rtkMisc.GetLine(), extData)
		}
	} else if nextState == STATE_TRANS && nextCommand == COMM_DST {
		if extData, ok := event.Data.(rtkCommon.ExtDataFile); ok {
			log.Printf("[%s] Ready to accept file,ActionType:[%s]", rtkMisc.GetFuncInfo(), extData.ActionType)
			// [Dst]: Setup file drop Data and DO NOT send msg

			if len(extData.SrcFileList) == 0 && len(extData.FolderList) == 0 {
				log.Printf("[%s] ID[%s] get file drop data is null", rtkMisc.GetFuncInfo(), id)
				return false
			}

			if extData.ActionType == rtkCommon.P2PFileActionType_Drop {
				rtkFileDrop.SetupDstFileListDrop(id, ipAddr, "", extData.TotalDescribe, extData.SrcFileList, extData.FolderList, extData.TotalSize, extData.TimeStamp)
			} else if extData.ActionType == rtkCommon.P2PFileActionType_Drag {
				rtkFileDrop.SetupDstDragFileList(id, ipAddr, extData.SrcFileList, extData.FolderList, extData.TotalSize, extData.TimeStamp, extData.TotalDescribe)
			} else {
				log.Printf("[%s] ID[%s] IP:[%s] get file unknown ActionType:[%s] ", rtkMisc.GetFuncInfo(), id, ipAddr, extData.ActionType)
				return false
			}
		} else {
			log.Printf("[%s %d] Err: Setup file drop failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
			return false
		}
	} else {
		if nextState == STATE_TRANS && nextCommand == COMM_SRC {
			if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
				if fileDropInfo, ok := rtkFileDrop.GetFileDropData(id); ok {
					rtkConnection.BuildFileDropItemStreamListener(fileDropInfo.TimeStamp)
				} else {
					log.Printf("[%s] ID:[%s] Not found fileDrop data", rtkMisc.GetFuncInfo(), id)
					return false
				}
			}
		} else if nextState == STATE_IO && nextCommand == COMM_DST {
			buildItemFileDropStream := func() (uint64, rtkMisc.CrossShareErr) { // [Dst]: every FileDropData need build a new stream
				fileDropInfo, ok := rtkFileDrop.GetFileDropData(id)
				if !ok {
					log.Printf("[%s] ID:[%s] Not found fileDrop data", rtkMisc.GetFuncInfo(), id)
					return 0, rtkMisc.ERR_BIZ_FT_DATA_EMPTY
				}

				if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
 					if errCode := rtkConnection.NewFileDropItemStream(ctx, id, fileDropInfo.TimeStamp); errCode != rtkMisc.SUCCESS {
 						return fileDropInfo.TimeStamp, errCode
 					}
				} else {
					if errCode := rtkConnection.BuildFmtTypeTalker(ctx, id, event.Cmd.FmtType); errCode != rtkMisc.SUCCESS {
						log.Printf("[%s]BuildFmtTypeTalker errCode:%+v ", rtkMisc.GetFuncInfo(), errCode)
						return fileDropInfo.TimeStamp, errCode
					}
 				}
				return fileDropInfo.TimeStamp, rtkMisc.SUCCESS
			}

			fileTransDataId, errCode := buildItemFileDropStream()
			if errCode != rtkMisc.SUCCESS {
				SendFileTransOpenStreamErrToSrc(id, ipAddr, fileTransDataId)
				return false
 			}
			
			timeStamp := rtkFileDrop.SetFilesDataToCacheAsDst(id)
			if rtkFileDrop.GetFilesTransferDataCacheCount(id) <= rtkGlobal.FilesConcurrentTransferMaxSize {
				rtkMisc.GoSafe(func() { processIoRead(ctx, id, ipAddr, event.Cmd.FmtType, timeStamp) }) // [Dst]: be ready to receive file drop raw data
			} else {
				log.Printf("[%s] ID:[%s] there are file data transfer is in progress, queue up and wait!", rtkMisc.GetFuncInfo(), id)
			}
		}
		var msg Peer2PeerMessage
		if buildMessage(&msg, id, event) {
			writeToSocket(&msg, id)
		} else {
			log.Printf("[%s %d] Build message failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
			return false
		}
	}
	return true
}

func processXClip(ctx context.Context, id, ipAddr string, event EventResult) bool {
	nextState := event.Cmd.State
	nextCommand := event.Cmd.Command

	if nextState == STATE_IO && nextCommand == COMM_SRC {
		// [Src]: Start to trans XClip
		rtkMisc.GoSafe(func() { processIoWrite(ctx, id, ipAddr, event.Cmd.FmtType, 0) })
	} else {
		if nextState == STATE_INFO && nextCommand == COMM_DST && !rtkUtils.GetPeerClientIsSupportXClip(id) { // not support XClip
			if extData, ok := event.Data.(rtkCommon.ExtDataXClip); ok {
				if extData.TextLen > 0 {
					rtkClipboard.SetupDstPasteXClipData(id, extData.Text, nil, nil, nil)
					return true
				}
			}
		} else if nextState == STATE_TRANS && nextCommand == COMM_DST {
			if extData, ok := event.Data.(rtkCommon.ExtDataXClip); ok {
				rtkClipboard.SetupDstPasteXClipHead(id, extData.TextLen, extData.ImageLen, extData.HtmlLen, extData.RtfLen)
			} else {
				log.Printf("[%s] Err: Setup past XClip failed", rtkMisc.GetFuncInfo())
				return false
			}

			event.Cmd.State = STATE_IO // XClip data has no paste event , directly enter IO state

			if errCode := rtkConnection.BuildFmtTypeTalker(ctx, id, rtkCommon.XCLIP_CB); errCode != rtkMisc.SUCCESS {
				log.Printf("[%s]BuildFmtTypeTalker errCode:%+v ", rtkMisc.GetFuncInfo(), errCode)
				return false
			}
			// [Dst]: Start to trans XClip
			rtkMisc.GoSafe(func() { processIoRead(ctx, id, ipAddr, event.Cmd.FmtType, 0) })
		}

		var msg Peer2PeerMessage
		if buildMessage(&msg, id, event) {
			writeToSocket(&msg, id)
		} else {
			log.Printf("[%s %d] Build message failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
			return false
		}
	}
	return true
}

func processTask(ctx context.Context, curState *StateType, curCommand *CommandType, id, ipAddr string, event EventResult) {
	ret := true
	switch event.Cmd.FmtType {
	case rtkCommon.FILE_DROP:
		log.Println("ProcessFileDrop triggered")
		ret = processFileDrop(ctx, id, ipAddr, event)
	case rtkCommon.XCLIP_CB:
		log.Println("processXClip triggered")
		ret = processXClip(ctx, id, ipAddr, event)
	default:
		log.Printf("[%s]Unknown cmd FmtType:[%s]", rtkMisc.GetFuncInfo(), event.Cmd.FmtType)
	}

	if !ret {
		log.Printf("Invalid state: cur(%s, %s), next(%s, %s)", *curState, *curCommand, event.Cmd.State, event.Cmd.Command)
	}
}

func isValidState(curState StateType, curCommand CommandType, nextState StateType, nextCommand CommandType) bool {
	// Src: Always allow to send info msg
	enforceSend := ((nextState == STATE_INIT) && (nextCommand == COMM_SRC)) ||
		((nextState == STATE_INFO) && (nextCommand == COMM_SRC))
	if enforceSend {
		return true
	}

	// Dst: Always allow to receive info msg
	// Maybe we can use this case ignore msg when data transferring
	enforceRec := ((nextState == STATE_INFO) && (nextCommand == COMM_DST))
	if enforceRec {
		return true
	}

	// Dst: Always allow to Accept file Drop Response, ignore curState
	enforceAccept := (nextState == STATE_IO) && (nextCommand == COMM_DST)
	if enforceAccept {
		return true
	}

	ret := false
	if curState == STATE_INIT {
		ret = (nextState == STATE_INFO && nextCommand == COMM_DST)
	} else if curState == STATE_INFO {
		if curCommand == COMM_SRC {
			ret = (nextState == STATE_TRANS && nextCommand == COMM_SRC)
		} else if curCommand == COMM_DST {
			ret = (nextState == STATE_TRANS && nextCommand == COMM_DST)
		}
	} else if curState == STATE_TRANS {
		if curCommand == COMM_SRC {
			ret = (nextState == STATE_IO && nextCommand == COMM_SRC)
		} else if curCommand == COMM_DST {
			ret = (nextState == STATE_IO && nextCommand == COMM_DST)
		}
	} else if curState == STATE_IO {
		if curCommand == COMM_SRC {
			ret = (nextState == STATE_IO && nextCommand == COMM_SRC)
		} else if curCommand == COMM_DST {
			ret = (nextState == STATE_IO && nextCommand == COMM_DST)
		}
	}

	if !ret {
		log.Printf("[%s %d] Invalid state: cur(%s, %s), next(%s, %s)", rtkMisc.GetFuncName(), rtkMisc.GetLine(), curState, curCommand, nextState, nextCommand)
	}
	return ret
}

func ProcessEventsForPeer(ctx context.Context, id, ipAddr string) {

	curState := STATE_INIT
	curCommand := COMM_INIT

	eventResultClipboard := make(chan EventResult)
	eventResultFileDrop := make(chan EventResult)
	eventResultSocket := make(chan EventResult)
	rtkMisc.GoSafe(func() { HandleClipboardEvent(ctx, eventResultClipboard, id) })
	rtkMisc.GoSafe(func() { HandleFileDropEvent(ctx, eventResultFileDrop, id) })
	rtkMisc.GoSafe(func() { HandleReadInbandFromSocket(ctx, eventResultSocket, id, ipAddr) })

	handleEvent := func(event EventResult) {
		buildState := curState
		buildCommand := curCommand

		if buildCmd(&curState, &curCommand, event, &buildState, &buildCommand) {
			// Move to next state and process
			event.Cmd.State = buildState
			event.Cmd.Command = buildCommand
			processTask(ctx, &curState, &curCommand, id, ipAddr, event)
		}
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] ID:[%s] ProcessEventsForPeer is End by context", rtkMisc.GetFuncInfo(), id)
			if rtkClipboard.GetLastClipboardData().SourceID == id {
				rtkClipboard.ResetLastClipboardData()
			}
			return
		case event, ok := <-eventResultClipboard:
			if !ok {
				continue
			}
			log.Printf("[ProcessEvent Clipboard][%s] EventResult fmt=%s, state=%s, cmd=%s", ipAddr, event.Cmd.FmtType, event.Cmd.State, event.Cmd.Command)
			handleEvent(event)
		case event, ok := <-eventResultFileDrop:
			if !ok {
				continue
			}
			log.Printf("[ProcessEvent FileDrop][%s] EventResult fmt=%s, state=%s, cmd=%s", ipAddr, event.Cmd.FmtType, event.Cmd.State, event.Cmd.Command)
			handleEvent(event)
		case event, ok := <-eventResultSocket:
			if !ok {
				continue
			}
			log.Printf("[ProcessEvent Socket][%s] EventResult fmt=%s, state=%s, cmd=%s", ipAddr, event.Cmd.FmtType, event.Cmd.State, event.Cmd.Command)
			handleEvent(event)
		}
	}
}
