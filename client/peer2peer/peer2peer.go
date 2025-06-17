package peer2peer

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	rtkClipboard "rtk-cross-share/client/clipboard"
	rtkCommon "rtk-cross-share/client/common"
	rtkConnection "rtk-cross-share/client/connection"
	rtkFileDrop "rtk-cross-share/client/filedrop"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync/atomic"
	"time"
)

func HandleClipboardEvent(ctxMain context.Context, readSocketMode *atomic.Value, resultChan chan<- EventResult, id string) {
	resultChanText := make(chan rtkCommon.ClipBoardData)
	resultChanImg := make(chan rtkCommon.ClipBoardData)
	resultChanPasteImg := make(chan struct{})
	rtkMisc.GoSafe(func() { rtkClipboard.WatchClipboardText(ctxMain, id, resultChanText) })
	rtkMisc.GoSafe(func() { rtkClipboard.WatchClipboardImg(ctxMain, id, resultChanImg) })
	rtkMisc.GoSafe(func() { rtkClipboard.WatchClipboardPasteImg(ctxMain, id, resultChanPasteImg) })
	for {
		select {
		case <-ctxMain.Done():
			close(resultChan)
			return
		case <-resultChanPasteImg:
			//setIoSocketMode(readSocketMode, rtkCommon.IMAGE_CB)
			resultChan <- EventResult{
				Cmd: DispatchCmd{
					FmtType: rtkCommon.IMAGE_CB,
					State:   STATE_IO,
					Command: COMM_DST,
				},
				Data: "",
			}
		case cbData := <-resultChanText:
			//setMsgSocketMode(readSocketMode)
			resultChan <- EventResult{
				Cmd: DispatchCmd{
					FmtType: rtkCommon.TEXT_CB,
					State:   STATE_INIT,
					Command: COMM_SRC,
				},
				Data: cbData,
			}
		case cbData := <-resultChanImg:
			//setMsgSocketMode(readSocketMode)
			resultChan <- EventResult{
				Cmd: DispatchCmd{
					FmtType: rtkCommon.IMAGE_CB,
					State:   STATE_INIT,
					Command: COMM_SRC,
				},
				Data: cbData,
			}
		}
	}
}

func HandleFileDropEvent(ctxMain context.Context, readSocketMode *atomic.Value, resultChan chan<- EventResult, id string) {
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
					//setMsgSocketMode(readSocketMode)
					resultChan <- EventResult{
						Cmd: DispatchCmd{
							FmtType: rtkCommon.FILE_DROP,
							State:   STATE_INIT,
							Command: COMM_SRC,
						},
						Data: rtkCommon.ExtDataFile{
							SrcFileList:   data.SrcFileList,
							ActionType:    data.ActionType,
							FileType:      data.FileType,
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
						//setIoSocketMode(readSocketMode, rtkCommon.FILE_DROP)
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

func handleReadFromSocketMsg(buffer []byte, len int, msg *Peer2PeerMessage) rtkMisc.CrossShareErr {
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
		msg.ExtData = extDataText
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
				log.Printf("[%s] Err: decode ExtDataImg:%+v", rtkMisc.GetFuncInfo(), err)
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
			msg.ExtData = extDataImg
		}
	}
	return rtkMisc.SUCCESS
}

func HandleReadFromSocket(ctxMain context.Context, readSocketMode *atomic.Value, resultChan chan<- EventResult, id string) {
	defer close(resultChan)
	for {
		select {
		case <-ctxMain.Done():
			log.Printf("[%s][Socket] ID:[%s] Err: Read operation is done by context", rtkMisc.GetFuncInfo(), id)
			return
		default:
			buffer := make([]byte, 32*1024) // 32KB
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
			errCode = handleReadFromSocketMsg(buffer, nLen, &msg)
			if errCode != rtkMisc.SUCCESS {
				log.Printf("[%s] handle Read message error, errCode:%d, retrying...", rtkMisc.GetFuncInfo(), errCode)
				continue
			}
			log.Printf("[%s] EventResult fmt=%s, state=%s, cmd=%s", rtkMisc.GetFuncInfo(), msg.FmtType, msg.State, msg.Command)

			if msg.Command == COMM_DISCONNECT {
				rtkConnection.OfflineEvent(id)
				continue
			} else if msg.Command == COMM_FILE_TRANSFER_SRC_INTERRUPT {
				fileDropData, ok := rtkFileDrop.GetFileDropData(id)
				if ok {
					HandleDataTransferError(COMM_CANCEL_SRC, id, fileDropData.DstFilePath)
				}
				// TODO: check if necessary to notice GUI the SRC interrupt file transfer
				CancelDstFileTransfer(id) // Dst: Copy need cancel
				if fileTransCode, ok := msg.ExtData.(rtkMisc.CrossShareErr); ok {
					log.Printf("[%s] (DST) Copy operation was canceled by src errCode:%d!", rtkMisc.GetFuncInfo(), fileTransCode)
				}
				continue
			} else if msg.Command == COMM_FILE_TRANSFER_DST_INTERRUPT {
				CancelSrcFileTransfer(id) // Src: Copy need cancel
				if fileTransCode, ok := msg.ExtData.(rtkMisc.CrossShareErr); ok {
					if fileTransCode == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI {
						// TODO: check if necessary to notice GUI the DST interrupt file transfer
						log.Printf("[%s](SRC) ID:[%s] Copy file operation was canceled by dst GUI !", rtkMisc.GetFuncInfo(), id)
					} else {
						log.Printf("[%s](SRC) ID:[%s] Copy file operation was canceled by dst errCode:%d!", rtkMisc.GetFuncInfo(), id, fileTransCode)
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

	switch msg.FmtType {
	case rtkCommon.TEXT_CB:
		if cbData, ok := event.Data.(rtkCommon.ClipBoardData); ok {
			if extData, ok := cbData.ExtData.(rtkCommon.ExtDataText); ok {
				msg.ExtData = rtkCommon.ExtDataText{
					Text: extData.Text,
				}
				return true
			} else {
				log.Printf("[%s %d] Err: Import Clipboard - Text to msg failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
				return false
			}
		}

	case rtkCommon.IMAGE_CB:
		// Paste image and require data
		if event.Cmd.State == STATE_IO && event.Cmd.Command == COMM_DST {
			msg.State = STATE_TRANS
			msg.Command = COMM_DST
			return true
		} else if event.Cmd.Command == COMM_SRC {
			if extData, ok := rtkClipboard.GetLastClipboardData().ExtData.(rtkCommon.ExtDataImg); ok {
				msg.ExtData = rtkCommon.ExtDataImg{
					Size:   extData.Size,
					Header: extData.Header,
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
					SrcFileList:   extData.SrcFileList,
					ActionType:    extData.ActionType,
					FileType:      extData.FileType,
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
	}

	return false
}

func WaitForReply(s net.Conn, match CommandType) rtkCommon.SocketErr {
	timeout := 5 * time.Second
	s.SetReadDeadline(time.Now().Add(timeout))

	buffer := make([]byte, 65535)
	n, err := s.Read(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			log.Println("Read fail network error", netErr.Error())
			if netErr.Timeout() {
				return rtkCommon.ERR_TIMEOUT
			}
			return rtkCommon.ERR_NETWORK
		} else {
			log.Println("Read fail", err.Error())
			return rtkCommon.ERR_OTHER
		}
	}

	buffer = buffer[:n]
	buffer = bytes.Trim(buffer, "\x00")
	buffer = bytes.Trim(buffer, "\x13")

	var msg Peer2PeerMessage
	err = json.Unmarshal(buffer, msg)
	if err != nil {
		log.Println("Failed to unmarshal P2PMessage data", err.Error())
		log.Printf("Err JSON len[%d] data:[%s] ", n, string(buffer))
		rtkUtils.WriteErrJson(s.RemoteAddr().String(), buffer)
		return rtkCommon.ERR_JSON
	}

	log.Println("Read msg from:", s.RemoteAddr(), "Cmd =", msg.Command, "FmtType =", msg.FmtType)
	if msg.Command != match {
		return rtkCommon.ERR_OTHER
	}
	return rtkCommon.OK
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

func HandleDataTransferError(inbandCmd CommandType, id, fileName string) {
	switch inbandCmd {
	case COMM_CANCEL_SRC:
		rtkPlatform.GoEventHandle(rtkCommon.EVENT_TYPE_OPEN_FILE_ERR, id, fileName)
	case COMM_CANCEL_DST:
		rtkPlatform.GoEventHandle(rtkCommon.EVENT_TYPE_RECV_TIMEOUT, id, fileName)
	default:
		log.Println("[DataTransferError]: Unhandled type")
	}
}

func IsTransferError(buffer []byte) bool {
	var msg Peer2PeerMessage
	var js json.RawMessage
	if json.Unmarshal(buffer, &js) != nil {
		return false
	}

	data := bytes.Trim(buffer, "\x00")
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return false
	}

	if msg.Command == COMM_CANCEL_SRC {
		return true
	}
	return false
}

func processIoWrite(id, ipAddr string, fmtType rtkCommon.TransFmtType) {
	resultCode := rtkMisc.SUCCESS
	if fmtType == rtkCommon.FILE_DROP {
		resultCode = writeFileDataToSocket(id, ipAddr)
		if resultCode != rtkMisc.SUCCESS {
			log.Printf("(SRC) ID[%s] IP[%s] Copy file list To Socket failed, ERR code:[%d]", id, ipAddr, resultCode)
			if resultCode != rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL {
				SendCmdMsgToPeer(id, COMM_FILE_TRANSFER_SRC_INTERRUPT, fmtType, resultCode)
			}
		}
	} else if fmtType == rtkCommon.IMAGE_CB {
		resultCode = writeImageToSocket(id)
		if resultCode != rtkMisc.SUCCESS {
			log.Printf("(SRC) ID[%s] IP[%s] Copy image To Socket failed, ERR code:[%d]", id, ipAddr, resultCode)
			SendCmdMsgToPeer(id, COMM_CB_TRANSFER_SRC_INTERRUPT, fmtType, resultCode)
		}
	}
}

func processIoRead(id, ipAddr, deviceName string, fmtType rtkCommon.TransFmtType) {
	if fmtType == rtkCommon.FILE_DROP {
		dstFileName, resultCode := handleFileDataFromSocket(id, ipAddr, deviceName)
		if resultCode != rtkMisc.SUCCESS {
			log.Printf("(DST) ID[%s] IP[%s] Copy file [%s] From Socket failed, ERR code:[%d]", id, ipAddr, dstFileName, resultCode)
			// Both user cancellations and exceptions require notification to the other end
			HandleDataTransferError(COMM_CANCEL_DST, id, dstFileName)
			if resultCode != rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL {
				SendCmdMsgToPeer(id, COMM_FILE_TRANSFER_DST_INTERRUPT, fmtType, resultCode)
			}
		}
	} else if fmtType == rtkCommon.IMAGE_CB {
		resultCode := handleCopyImageFromSocket(id, ipAddr)
		if resultCode != rtkMisc.SUCCESS {
			log.Printf("(DST) ID[%s] IP[%s] Copy image From Socket failed, ERR code:[%d] ...", id, ipAddr, resultCode)
			SendCmdMsgToPeer(id, COMM_CB_TRANSFER_DST_INTERRUPT, fmtType, resultCode)
			HandleDataTransferError(COMM_CANCEL_DST, id, "")
		}
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

func processTextCB(id string, event EventResult) bool {
	nextState := event.Cmd.State
	nextCommand := event.Cmd.Command

	if nextState == STATE_INFO && nextCommand == COMM_DST {
		if extData, ok := event.Data.(rtkCommon.ExtDataText); ok {
			log.Printf("[%s %d] Ready to paste text", rtkMisc.GetFuncName(), rtkMisc.GetLine())
			// [Dst]: Setup clipboard and DO NOT send msg
			rtkMisc.GoSafe(func() { rtkClipboard.SetupDstPasteText(id, []byte(extData.Text)) })
		} else {
			log.Printf("[%s %d] Err: Setup past text failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
			return false
		}
	} else {
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

func processImageCB(id string, event EventResult) bool {
	nextState := event.Cmd.State
	nextCommand := event.Cmd.Command

	if nextState == STATE_IO && nextCommand == COMM_SRC {
		// [Src]: Start to trans image
		clientInfo, err := rtkUtils.GetClientInfo(id)
		if err != nil {
			log.Printf("[%s] %s", rtkMisc.GetFuncInfo(), err.Error())
			return false
		}
		rtkMisc.GoSafe(func() { processIoWrite(id, clientInfo.IpAddr, event.Cmd.FmtType) })
	} else if nextState == STATE_TRANS && nextCommand == COMM_DST {
		if extData, ok := event.Data.(rtkCommon.ExtDataImg); ok {
			log.Printf("[%s %d] Ready to paste image", rtkMisc.GetFuncName(), rtkMisc.GetLine())
			// [Dst]: Setup clipboard and DO NOT send msg
			rtkMisc.GoSafe(func() { rtkClipboard.SetupDstPasteImage(id, "", extData.Data, extData.Header, extData.Size.SizeLow) })
		} else {
			log.Printf("[%s %d] Err: Setup past image failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
			return false
		}
	} else {
		if nextState == STATE_IO && nextCommand == COMM_DST {
			ipAddr, ok := rtkUtils.GetClientIp(id)
			if !ok {
				//return rtkMisc.ERR_BIZ_GET_CLIENT_INFO_EMPTY
				return false
			}
			if errCode := rtkConnection.BuildFmtTypeTalker(id, rtkCommon.IMAGE_CB); errCode != rtkMisc.SUCCESS {
				log.Printf("[%s]BuildImageTalker errCode:%+v ", rtkMisc.GetFuncInfo(), errCode)
				return false
			}
			rtkMisc.GoSafe(func() { processIoRead(id, ipAddr, "", event.Cmd.FmtType) })
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

func processFileDrop(id string, event EventResult) bool {
	nextState := event.Cmd.State
	nextCommand := event.Cmd.Command

	if nextState == STATE_IO && nextCommand == COMM_SRC {
		// Receive response from dst
		if extData, ok := event.Data.(rtkCommon.FileDropCmd); ok {
			if extData == rtkCommon.FILE_DROP_ACCEPT {
				clientInfo, err := rtkUtils.GetClientInfo(id)
				if err != nil {
					log.Printf("[%s] %s", rtkMisc.GetFuncInfo(), err.Error())
					return false
				}
				rtkMisc.GoSafe(func() { processIoWrite(id, clientInfo.IpAddr, event.Cmd.FmtType) }) // [Src]: Start to trans file
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
			//log.Printf("[%s %d] Ready to accept file", rtkMisc.GetFuncName(), rtkMisc.GetLine())
			log.Printf("[%s %d] Ready to accept file,ActionType:[%s] FileType:[%s]", rtkMisc.GetFuncName(), rtkMisc.GetLine(), extData.ActionType, extData.FileType)
			// [Dst]: Setup file drop Data and DO NOT send msg

			if len(extData.SrcFileList) == 0 && len(extData.FolderList) == 0 {
				log.Printf("[%s] ID[%s] get file drop data is null", rtkMisc.GetFuncInfo(), id)
				return false
			}
			clientInfo, err := rtkUtils.GetClientInfo(id)
			if err != nil {
				log.Printf("[%s] %s", rtkMisc.GetFuncInfo(), err.Error())
				return false
			}

			if extData.ActionType == rtkCommon.P2PFileActionType_Drop {
				if extData.FileType == rtkCommon.P2PFile_Type_Single {
					rtkFileDrop.SetupDstFileDrop(id, clientInfo.IpAddr, clientInfo.Platform, extData.SrcFileList[0], extData.TimeStamp)
				} else if extData.FileType == rtkCommon.P2PFile_Type_Multiple {
					rtkFileDrop.SetupDstFileListDrop(id, clientInfo.IpAddr, clientInfo.Platform, extData.TotalDescribe, extData.SrcFileList, extData.FolderList, extData.TotalSize, extData.TimeStamp)
				} else {
					log.Printf("[%s] ID[%s] IP:[%s] get file unknown FileType:[%s] ", rtkMisc.GetFuncInfo(), id, clientInfo.IpAddr, extData.FileType)
					return false
				}
			} else if extData.ActionType == rtkCommon.P2PFileActionType_Drag {
				if extData.FileType == rtkCommon.P2PFile_Type_Single {
					rtkFileDrop.SetupDstDragFile(id, clientInfo.IpAddr, clientInfo.Platform, extData.SrcFileList[0], extData.TotalSize, extData.TimeStamp)
				} else if extData.FileType == rtkCommon.P2PFile_Type_Multiple {
					rtkFileDrop.SetupDstDragFileList(id, clientInfo.IpAddr, clientInfo.Platform, extData.SrcFileList, extData.FolderList, extData.TotalSize, extData.TimeStamp, extData.TotalDescribe)
				} else {
					log.Printf("[%s] ID[%s] IP:[%s] get file unknown FileType:[%s] ", rtkMisc.GetFuncInfo(), id, clientInfo.IpAddr, extData.FileType)
					return false
				}
			} else {
				log.Printf("[%s] ID[%s] IP:[%s] get file unknown ActionType:[%s] ", rtkMisc.GetFuncInfo(), id, clientInfo.IpAddr, extData.ActionType)
				return false
			}
		} else {
			log.Printf("[%s %d] Err: Setup file drop failed", rtkMisc.GetFuncName(), rtkMisc.GetLine())
			return false
		}
	} else {
		if nextState == STATE_IO && nextCommand == COMM_DST { // [Dst]: build fmtType Stream first
			clientInfo, err := rtkUtils.GetClientInfo(id)
			if err != nil {
				//return rtkMisc.ERR_BIZ_GET_CLIENT_INFO_EMPTY
				return false
			}
			if errCode := rtkConnection.BuildFmtTypeTalker(id, event.Cmd.FmtType); errCode != rtkMisc.SUCCESS {
				log.Printf("[%s]BuildFileTalker errCode:%+v ", rtkMisc.GetFuncInfo(), errCode)
				return false
			}
			rtkMisc.GoSafe(func() { processIoRead(id, clientInfo.IpAddr, clientInfo.DeviceName, event.Cmd.FmtType) }) // [Dst]: be ready to receive file drop raw data
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

func processTask(curState *StateType, curCommand *CommandType, id string, event EventResult) {
	ret := true
	switch event.Cmd.FmtType {
	case rtkCommon.TEXT_CB:
		log.Println("ProcessTextCB triggered")
		ret = processTextCB(id, event)
	case rtkCommon.IMAGE_CB:
		log.Println("ProcessImageCB triggered")
		ret = processImageCB(id, event)
	case rtkCommon.FILE_DROP:
		log.Println("ProcessFileDrop triggered")
		ret = processFileDrop(id, event)
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

func ProcessEventsForPeer(id, ipAddr string, ctx context.Context) {
	curState := STATE_INIT
	curCommand := COMM_INIT

	var readSocketMode atomic.Value
	//setMsgSocketMode(&readSocketMode)
	eventResultClipboard := make(chan EventResult)
	eventResultFileDrop := make(chan EventResult)
	eventResultSocket := make(chan EventResult)
	rtkMisc.GoSafe(func() { HandleClipboardEvent(ctx, &readSocketMode, eventResultClipboard, id) })
	rtkMisc.GoSafe(func() { HandleFileDropEvent(ctx, &readSocketMode, eventResultFileDrop, id) })
	rtkMisc.GoSafe(func() { HandleReadFromSocket(ctx, &readSocketMode, eventResultSocket, id) })

	handleEvent := func(event EventResult) {
		buildState := curState
		buildCommand := curCommand

		if buildCmd(&curState, &curCommand, event, &buildState, &buildCommand) {
			// Move to next state and process
			event.Cmd.State = buildState
			event.Cmd.Command = buildCommand
			processTask(&curState, &curCommand, id, event)
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
		case event := <-eventResultClipboard:
			if event.Cmd.State == "" || event.Cmd.Command == "" {
				continue
			}
			log.Printf("[ProcessEvent Clipboard][%s] EventResult fmt=%s, state=%s, cmd=%s", ipAddr, event.Cmd.FmtType, event.Cmd.State, event.Cmd.Command)
			handleEvent(event)
		case event := <-eventResultFileDrop:
			if event.Cmd.State == "" || event.Cmd.Command == "" {
				continue
			}
			log.Printf("[ProcessEvent FileDrop][%s] EventResult fmt=%s, state=%s, cmd=%s", ipAddr, event.Cmd.FmtType, event.Cmd.State, event.Cmd.Command)
			handleEvent(event)
		case event := <-eventResultSocket:
			if event.Cmd.State == "" || event.Cmd.Command == "" {
				continue
			}
			log.Printf("[ProcessEvent Socket][%s] EventResult fmt=%s, state=%s, cmd=%s", ipAddr, event.Cmd.FmtType, event.Cmd.State, event.Cmd.Command)
			handleEvent(event)
		}
	}
}
