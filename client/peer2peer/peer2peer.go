package peer2peer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	rtkClipboard "rtk-cross-share/clipboard"
	rtkCommon "rtk-cross-share/common"
	rtkFileDrop "rtk-cross-share/filedrop"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"sync/atomic"
	"time"
)

func HandleClipboardEvent(ctxMain context.Context, readSocketMode *atomic.Value, resultChan chan<- EventResult, s net.Conn, ipAddr string) {
	resultChanText := make(chan rtkCommon.ClipBoardData)
	resultChanImg := make(chan rtkCommon.ClipBoardData)
	resultChanPasteImg := make(chan bool)
	rtkUtils.GoSafe(func() {rtkClipboard.WatchClipboardText(ctxMain, ipAddr, resultChanText)})
	rtkUtils.GoSafe(func() {rtkClipboard.WatchClipboardImg(ctxMain, ipAddr, resultChanImg)})
	rtkUtils.GoSafe(func() {rtkClipboard.WatchClipboardPasteImg(ctxMain, ipAddr, s.RemoteAddr().String(), resultChanPasteImg)})
	for {
		select {
		case <-ctxMain.Done():
			return
		case isPasted := <-resultChanPasteImg:
			if isPasted {
				setIoSocketMode(readSocketMode, rtkCommon.IMAGE_CB)
				resultChan <- EventResult{
					Cmd: DispatchCmd{
						FmtType: rtkCommon.IMAGE_CB,
						State:   STATE_IO,
						Command: COMM_DST,
					},
					Data: "",
				}
			}
		case cbData := <-resultChanText:
			setMsgSocketMode(readSocketMode)
			resultChan <- EventResult{
				Cmd: DispatchCmd{
					FmtType: rtkCommon.TEXT_CB,
					State:   STATE_INIT,
					Command: COMM_SRC,
				},
				Data: cbData,
			}
		case cbData := <-resultChanImg:
			setMsgSocketMode(readSocketMode)
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

func HandleFileDropEvent(ctxMain context.Context, readSocketMode *atomic.Value, resultChan chan<- EventResult, ipAddr string) {
	resultReqIp := make(chan string)
	resultRespIp := make(chan string)
	rtkUtils.GoSafe(func() {rtkFileDrop.WatchFileDropReqEvent(ctxMain, ipAddr, resultReqIp)})
	rtkUtils.GoSafe(func() {rtkFileDrop.WatchFileDropRespEvent(ctxMain, ipAddr, resultRespIp)})
	for {
		select {
		case <-ctxMain.Done():
			return
		case ip := <-resultReqIp:
			if ip == ipAddr {
				if data, ok := rtkFileDrop.GetFileDropData(ip); ok {
					setMsgSocketMode(readSocketMode)
					resultChan <- EventResult{
						Cmd: DispatchCmd{
							FmtType: rtkCommon.FILE_DROP,
							State:   STATE_INIT,
							Command: COMM_SRC,
						},
						Data: data.SrcFileInfo,
					}
				}
			}
		case ip := <-resultRespIp:
			if ip == ipAddr {
				if data, ok := rtkFileDrop.GetFileDropData(ip); ok {
					if data.Cmd == rtkCommon.FILE_DROP_ACCEPT {
						// Accept file: Prepeare to receive data
						setIoSocketMode(readSocketMode, rtkCommon.FILE_DROP)
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
						rtkFileDrop.ResetFileDropData(ipAddr)
					} else {
						log.Printf("[%s %d] Invalid fileDrop response data:%s with IP: %s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), data.Cmd, ip)
					}
				} else {
					log.Printf("[%s %d] Empty fileDrop response data with IP:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), ip)
				}
			}
		}
	}
}

func handleReadFromSocketMsg(buffer []byte, len int, msg *Peer2PeerMessage, s net.Conn) rtkCommon.SocketErr {
	buffer = buffer[:len]
	buffer = bytes.Trim(buffer, "\x00")
	buffer = bytes.Trim(buffer, "\x13")

	type TempMsg struct {
		ExtData json.RawMessage
		Peer2PeerMessage
	}

	var temp TempMsg
	err := json.Unmarshal(buffer, &temp)
	// err := json.Unmarshal(buffer, msg)
	if err != nil {
		log.Println("Failed to unmarshal P2PMessage data", err.Error())
		log.Printf("Err JSON len[%d] data:[%s] ", len, string(buffer))
		//rtkUtils.WriteErrJson(s.RemoteAddr().String(), buffer)
		return rtkCommon.ERR_JSON
	}

	*msg = temp.Peer2PeerMessage
	switch msg.FmtType {
	case rtkCommon.TEXT_CB:
		var extDataText rtkCommon.ExtDataText
		err := json.Unmarshal(temp.ExtData, &extDataText)
		if err != nil {
			log.Println("Err: decode ExtDataText:", err)
			return rtkCommon.ERR_JSON
		}
		msg.ExtData = extDataText
	case rtkCommon.FILE_DROP:
		// Response accept or reject
		if msg.State == STATE_TRANS && msg.Command == COMM_DST {
			var extDataCmd rtkCommon.FileDropCmd
			err := json.Unmarshal(temp.ExtData, &extDataCmd)
			if err != nil {
				log.Println("Err: decode ExtDataFile:", err)
				return rtkCommon.ERR_JSON
			}
			msg.ExtData = extDataCmd
		} else {
			var extDataFile rtkCommon.ExtDataFile
			err := json.Unmarshal(temp.ExtData, &extDataFile)
			if err != nil {
				log.Println("Err: decode ExtDataFile:", err)
				return rtkCommon.ERR_JSON
			}
			msg.ExtData = extDataFile
		}
	case rtkCommon.IMAGE_CB:
		var extDataImg rtkCommon.ExtDataImg
		err := json.Unmarshal(temp.ExtData, &extDataImg)
		if err != nil {
			log.Println("Err: decode ExtDataImg:", err)
			return rtkCommon.ERR_JSON
		}
		msg.ExtData = extDataImg
	}
	return rtkCommon.OK
}

func initFileDropDataTransfer(ip string, name *string, dstFile **os.File, receivedBytes *uint64, fileSize *uint64, startTime *int64) rtkCommon.TransferErr {
	fileDropData, ok := rtkFileDrop.GetFileDropData(ip)
	if !ok {
		log.Printf("[%s %d] Not found fileDrop data", rtkUtils.GetFuncName(), rtkUtils.GetLine())
		return rtkCommon.TRANS_ERR_OTHER
	}
	if fileDropData.Cmd != rtkCommon.FILE_DROP_ACCEPT {
		log.Printf("[%s %d] Invalid fildDrop cmd state:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), fileDropData.Cmd)
		return rtkCommon.TRANS_ERR_UNKNOWN_CMD
	}

	errOpenFile := OpenDstFile(dstFile, fileDropData.DstFilePath)
	if errOpenFile != nil {
		return rtkCommon.TRANS_ERR_CREATE_FAILED
	}

	*name = filepath.Base(fileDropData.DstFilePath)
	*startTime = time.Now().UnixNano()
	*fileSize = uint64(fileDropData.SrcFileInfo.FileSize_.SizeHigh)<<32 | uint64(fileDropData.SrcFileInfo.FileSize_.SizeLow)
	log.Printf("(DST) Start to Copy size:[%d]...", *fileSize)
	*receivedBytes = uint64(0)
	return rtkCommon.TRANS_OK
}

func initImageDataTransfer(receivedBytes *uint64, fileSize *uint64, startTime *int64) rtkCommon.TransferErr {
	cbData := rtkClipboard.GetLastClipboardData()
	extData, ok := cbData.ExtData.(rtkCommon.ExtDataImg)
	if !ok || cbData.FmtType != rtkCommon.IMAGE_CB {
		log.Printf("[%s %d] Invalid image ext data type", rtkUtils.GetFuncName(), rtkUtils.GetLine())
		return rtkCommon.TRANS_ERR_OTHER
	}

	*startTime = time.Now().UnixNano()
	*fileSize = uint64(extData.Size.SizeHigh)<<32 | uint64(extData.Size.SizeLow)
	log.Printf("(DST) Start to Copy size:[%d]...", *fileSize)
	*receivedBytes = uint64(0)
	return rtkCommon.TRANS_OK
}

func handleReadFromSocketFileRaw(Id, ipAddr, fileName string, file **os.File, buffer []byte, len int, receivedRawBytes *uint64, fileSize uint64, startTime int64) bool {
	log.Println("Receive data size:", len)
	if len == 0 {
		CloseDstFile(file)
		return true
	}

	if IsTransferError((buffer[:len])) {
		HandleDataTransferError(COMM_CANCEL_SRC, ipAddr, fileName)
		CloseDstFile(file)
		return true
	}
	*receivedRawBytes += uint64(len)
	WriteDstFile(file, buffer[:len])

	if *receivedRawBytes >= fileSize {
		log.Printf("(DST) End to Copy file, total:[%d] use [%d] ms...", fileSize, (time.Now().UnixNano()-startTime)/1e6)
		if fileDropData, ok := rtkFileDrop.GetFileDropData(ipAddr); ok {
			// For Windows file
			rtkPlatform.GoUpdateProgressBar(ipAddr, Id, fileSize, fileSize, int64(fileDropData.TimeStamp), fileDropData.DstFilePath)
			//  For Android file
			rtkPlatform.ReceiveFileDropCopyDataDone(int64(fileSize), fileDropData.DstFilePath)
			rtkFileDrop.ResetFileDropData(ipAddr)
		}
		CloseDstFile(file)
		return true
	} else {
		if fileDropData, ok := rtkFileDrop.GetFileDropData(ipAddr); ok {
			rtkPlatform.GoUpdateProgressBar(ipAddr, Id, fileSize, *receivedRawBytes, int64(fileDropData.TimeStamp), fileDropData.DstFilePath)
		}
	}

	return false
}
func handleReadFromSocketImageRaw(Id, ipAddr string, buffer []byte, len int, receivedRawBytes *uint64, fileSize uint64, startTime int64) bool {
	log.Println("Receive data size:", len)
	if len == 0 {
		return true
	}

	if IsTransferError((buffer[:len])) {
		//HandleDataTransferError(COMM_CANCEL_SRC)
		return true
	}

	rtkPlatform.GoDataTransfer(buffer[:len])
	*receivedRawBytes += uint64(len)
	if *receivedRawBytes >= fileSize {
		log.Printf("(DST) End to Copy img, total:[%d] use [%d] ms...", fileSize, (time.Now().UnixNano()-startTime)/1e6)
		if extData, ok := rtkClipboard.GetLastClipboardData().ExtData.(rtkCommon.ExtDataImg); ok { // Only For Android
			rtkPlatform.ReceiveImageCopyDataDone(int64(fileSize), extData.Header)
		}
		return true
	}

	return false
}

func HandleReadFromSocket(ctxMain context.Context, readSocketMode *atomic.Value, resultChan chan<- EventResult, s net.Conn, ipAddr string, errChan chan<- rtkCommon.SocketErr) {

	var dstFileName string
	defer func() {
		if !isMsgSocketMode(readSocketMode) {
			fmtType := getIoSocketFmtType(readSocketMode)
			if fmtType == rtkCommon.FILE_DROP {
				HandleDataTransferError(COMM_CANCEL_DST, ipAddr, dstFileName)
			}
		}
	}()

	// Params for data tranfer
	startTime := int64(0)
	isLastMsgSocketMode := isMsgSocketMode(readSocketMode)
	fileSize := uint64(0)
	receivedRawBytes := uint64(0)
	var dstFile *os.File

	for {
		select {
		case <-ctxMain.Done():
			log.Printf("[Socket][%s] Err: Read operation is done by main context", ipAddr)
			return
		default:
			buffer := make([]byte, 32*1024) // 32KB
			n, err := s.Read(buffer)
			if err != nil {
				// TODO: handle cancel status
				if netErr, ok := err.(net.Error); ok {
					log.Printf("[Socket][%s] Err: Read fail network error(%v)", ipAddr, netErr.Error())
					if netErr.Timeout() {
						errChan <- rtkCommon.ERR_TIMEOUT
						return
					}
					errChan <- rtkCommon.ERR_NETWORK
					return
				} else {
					log.Printf("[Socket][%s] Err: Read fail(%v)", ipAddr, err.Error())
					errChan <- rtkCommon.ERR_OTHER
					return
				}
			}

			// Reveice msg data
			if isMsgSocketMode(readSocketMode) {
				isLastMsgSocketMode = true

				var msg Peer2PeerMessage
				errSocket := handleReadFromSocketMsg(buffer, n, &msg, s)
				if errSocket != rtkCommon.OK {
					if errSocket == rtkCommon.ERR_CANCEL { // how it happened??
						if ctxMain.Err() == context.Canceled {
							log.Printf("[%s %d] Err: Read canceled, retrying...", rtkUtils.GetFuncName(), rtkUtils.GetLine())
						}
						continue
					} else if errSocket == rtkCommon.ERR_JSON {
						log.Printf("[%s %d] Err: Read invalid JSON msg, retrying...", rtkUtils.GetFuncName(), rtkUtils.GetLine())
						continue
					}
				}
				log.Printf("[%s %d] EventResult fmt=%s, state=%s, cmd=%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), msg.FmtType, msg.State, msg.Command)

				resultChan <- EventResult{
					Cmd: DispatchCmd{
						FmtType: msg.FmtType,
						State:   msg.State,
						Command: msg.Command,
					},
					Data: msg.ExtData,
				}
			} else { // Receive raw data
				fmtType := getIoSocketFmtType(readSocketMode)
				// TODO: refine this flow: how to check if first time downloading
				if isLastMsgSocketMode {
					// Init data transfer params
					switch fmtType {
					case rtkCommon.FILE_DROP:
						isLastMsgSocketMode = false
						ret := initFileDropDataTransfer(ipAddr, &dstFileName, &dstFile, &receivedRawBytes, &fileSize, &startTime)
						if ret != rtkCommon.TRANS_OK {
							log.Printf("[%s %d] Err: FileDrop tranfer initialized failed. Code:%d", rtkUtils.GetFuncName(), rtkUtils.GetLine(), ret)
							setMsgSocketMode(readSocketMode)
							isLastMsgSocketMode = true
							// TODO: display error dialog
							// TODO: send msg to SRC for cancel
							continue
						}
					case rtkCommon.IMAGE_CB:
						isLastMsgSocketMode = false
						ret := initImageDataTransfer(&receivedRawBytes, &fileSize, &startTime)
						if ret != rtkCommon.TRANS_OK {
							log.Printf("[%s %d] Err: Image tranfer initialized failed. Code:%d", rtkUtils.GetFuncName(), rtkUtils.GetLine(), ret)
							setMsgSocketMode(readSocketMode)
							isLastMsgSocketMode = true
							// TODO: display error dialog
							// TODO: send msg to SRC for cancel
							continue
						}
					default:
						log.Printf("[%s %d] Unknown data transfer type:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), fmtType)
					}
				}

				if fmtType == rtkCommon.FILE_DROP {
					if handleReadFromSocketFileRaw(s.RemoteAddr().String(), ipAddr, dstFileName, &dstFile, buffer, n, &receivedRawBytes, fileSize, startTime) {
						setMsgSocketMode(readSocketMode)
						isLastMsgSocketMode = true
					}
				} else if fmtType == rtkCommon.IMAGE_CB {
					if handleReadFromSocketImageRaw(s.RemoteAddr().String(), ipAddr, buffer, n, &receivedRawBytes, fileSize, startTime) {
						setMsgSocketMode(readSocketMode)
						isLastMsgSocketMode = true
					}
				}

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
		log.Printf("[%s %d] Invalid state: cur(%s, %s), next(%s, %s)", rtkUtils.GetFuncName(), rtkUtils.GetLine(), *curState, *curCommand, event.Cmd.State, event.Cmd.Command)
		return false
	}
}

func buildMessage(msg *Peer2PeerMessage, ipAddr string, event EventResult) bool {
	msg.SourceID = rtkGlobal.NodeInfo.ID
	msg.SourcePlatform = rtkPlatform.GetPlatform()
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
				log.Printf("[%s %d] Err: Import Clipboard - Text to msg failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
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
				log.Printf("[%s %d] Err: Import Clipboard - Image to msg failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
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
			if extData, ok := rtkFileDrop.GetFileDropData(ipAddr); ok {
				msg.ExtData = extData.Cmd
				return true
			}
		} else if event.Cmd.Command == COMM_SRC {
			if extData, ok := rtkFileDrop.GetFileDropData(ipAddr); ok {
				msg.ExtData = rtkCommon.ExtDataFile{
					Size:     extData.SrcFileInfo.FileSize_,
					FilePath: extData.SrcFileInfo.FilePath,
				}
				return true
			} else {
				log.Printf("[%s %d] Err: Import FileDrop - File to msg failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
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

func writeToSocket(msg Peer2PeerMessage, s net.Conn) rtkCommon.SocketErr {
	log.Printf("Write msg to: %s, platform=%s, fmt=%s, state=%s, cmd=%s", s.RemoteAddr(), msg.SourcePlatform, msg.FmtType, msg.State, msg.Command)
	encodedData, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal P2PMessage data:", err)
		return rtkCommon.ERR_JSON
	}

	encodedData = bytes.Trim(encodedData, "\x00")
	encodedData = bytes.Trim(encodedData, "\x13")

	if _, err := s.Write(encodedData); err != nil {
		if netErr, ok := err.(net.Error); ok {
			log.Println("Write fail network error", netErr.Error())
			return rtkCommon.ERR_NETWORK
		} else {
			log.Println("Write fail", err.Error())
			return rtkCommon.ERR_OTHER
		}
	}
	return rtkCommon.OK
}

func OpenSrcFile(file **os.File, filePath string) error {
	if *file != nil {
		(*file).Close()
		*file = nil
	}

	var err error
	*file, err = os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		*file = nil
		log.Println("Error opening file:", err, "filePath:", filePath)
		return err
	}
	_, errSeek := (*file).Seek(0, io.SeekStart)
	if errSeek != nil {
		(*file).Close()
		*file = nil
		log.Println("Error seek file:", err)
		return errSeek
	}
	log.Println("OpenSrcFile")
	return nil
}

func CloseSrcFile(file **os.File) {
	if *file == nil {
		return
	}
	(*file).Close()
	*file = nil
	log.Println("CloseSrcFile")
}

func OpenDstFile(file **os.File, filePath string) error {
	if *file != nil {
		(*file).Close()
		*file = nil
	}

	var err error
	*file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Error opening Dst file err: %v path: %v", err, filePath)
		*file = nil
		return err
	}
	log.Println("OpenDstFile")
	return nil
}

func CloseDstFile(file **os.File) {
	if *file == nil {
		return
	}

	log.Println("CloseDstFile")
	(*file).Close()
	*file = nil
}

func HandleDataTransferError(inbandCmd CommandType, ipAddr, fileName string) {
	switch inbandCmd {
	case COMM_CANCEL_SRC:
		rtkPlatform.GoEventHandle(rtkCommon.EVENT_TYPE_OPEN_FILE_ERR, ipAddr, fileName)
	case COMM_CANCEL_DST:
		rtkPlatform.GoEventHandle(rtkCommon.EVENT_TYPE_RECV_TIMEOUT, ipAddr, fileName)
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

func WriteDstFile(file **os.File, content []byte) {
	if *file != nil {
		if _, err := (*file).Write(content); err != nil {
			log.Println("Error writing to dst file:", err)
			return
		}
	} else {
		log.Printf("[%s %d] Err: Dst file is not open!", rtkUtils.GetFuncName(), rtkUtils.GetLine())
	}
}

func processIoWrite(s net.Conn, ipAddr string, fmtType rtkCommon.TransFmtType) {
	if fmtType == rtkCommon.FILE_DROP {
		fileDropReqData, ok := rtkFileDrop.GetFileDropData(ipAddr)
		if !ok {
			log.Printf("[%s %d] Err: Not found fileDrop data", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			goto ErrFlag
		}
		if fileDropReqData.SrcFileInfo.FilePath == "" ||
			(fileDropReqData.SrcFileInfo.FileSize_.SizeHigh == 0 && fileDropReqData.SrcFileInfo.FileSize_.SizeLow == 0) {
			log.Printf("[%s %d] Err: Invalid fileDrop data", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			goto ErrFlag
		}
		var srcFile *os.File
		errOpenFile := OpenSrcFile(&srcFile, fileDropReqData.SrcFileInfo.FilePath)
		if errOpenFile != nil {
			goto ErrFlag
		}

		defer CloseSrcFile(&srcFile)
		defer rtkFileDrop.ResetFileDropData(ipAddr)
		log.Println("(SRC) Start to copy file ...")
		startTime := time.Now().UnixNano()
		// TODO: write data timeout
		// s.SetWriteDeadline(time.Now().Add(10 * time.Second))
		// TODO: Update progress bar
		n, err := io.Copy(s, srcFile)
		// s.SetWriteDeadline(time.Time{})
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("Error sending file timeout:", netErr)
			} else {
				log.Println("Error sending file:", err)
			}
			return
		}
		log.Printf("(SRC) End to copy file, size:[%d] use [%d] ms ...", n, (time.Now().UnixNano()-startTime)/1e6)
	} else if fmtType == rtkCommon.IMAGE_CB {
		if extData, ok := rtkClipboard.GetLastClipboardData().ExtData.(rtkCommon.ExtDataImg); ok {
			log.Printf("(SRC) Start to copy img, len[%d]...", len(extData.Data))
			startTime := time.Now().UnixNano()
			_, err := io.Copy(s, bytes.NewReader(extData.Data))
			if err != nil {
				log.Println("Error sending img:", err)
				return
			}
			log.Printf("(SRC) End to copy img, use [%d] ms ...", (time.Now().UnixNano()-startTime)/1e6)
		} else {
			log.Printf("[%s %d] Unknown ext data type", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			goto ErrFlag
		}
	}

	return

ErrFlag:
	var msg Peer2PeerMessage
	msg.SourceID = rtkGlobal.NodeInfo.ID
	msg.SourcePlatform = rtkPlatform.GetPlatform()
	msg.FmtType = fmtType
	msg.TimeStamp = uint64(time.Now().Unix())
	msg.Command = COMM_CANCEL_SRC
	writeToSocket(msg, s)
}

func updateStateCommand(curState *StateType, curCommand *CommandType, updateState StateType, updateCommand CommandType) {
	log.Printf("[%s %d] Current state from (%s, %s) to (%s, %s)", rtkUtils.GetFuncName(), rtkUtils.GetLine(), *curState, *curCommand, updateState, updateCommand)
	if updateState == STATE_INIT {
		// convert STATE_INIT to STATE_INFO for INIT
		*curState = STATE_INFO
	} else {
		*curState = updateState
	}
	*curCommand = updateCommand
}

func processTextCB(s net.Conn, ipAddr string, event EventResult) bool {
	nextState := event.Cmd.State
	nextCommand := event.Cmd.Command

	if nextState == STATE_INFO && nextCommand == COMM_DST {
		if extData, ok := event.Data.(rtkCommon.ExtDataText); ok {
			log.Printf("[%s %d] Ready to paste text", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			// [Dst]: Setup clipboard and DO NOT send msg
			rtkUtils.GoSafe(func() { rtkClipboard.SetupDstPasteText(s.RemoteAddr().String(), []byte(extData.Text)) })
		} else {
			log.Printf("[%s %d] Err: Setup past text failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			return false
		}
	} else {
		var msg Peer2PeerMessage
		if buildMessage(&msg, ipAddr, event) {
			writeToSocket(msg, s)
		} else {
			log.Printf("[%s %d] Build message failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			return false
		}
	}
	return true
}

func processImageCB(s net.Conn, ipAddr string, event EventResult) bool {
	nextState := event.Cmd.State
	nextCommand := event.Cmd.Command

	if nextState == STATE_IO {
		if nextCommand == COMM_SRC {
			rtkUtils.GoSafe(func() { processIoWrite(s, ipAddr, event.Cmd.FmtType) }) // [Src]: Start to trans file
		} else if nextCommand == COMM_DST {
			// TODO: refine this flow: put buildMessage together
			var msg Peer2PeerMessage
			if buildMessage(&msg, ipAddr, event) {
				writeToSocket(msg, s)
			} else {
				log.Printf("[%s %d] Build message failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
				return false
			}
		}
	} else if nextState == STATE_TRANS && nextCommand == COMM_DST {
		if extData, ok := event.Data.(rtkCommon.ExtDataImg); ok {
			log.Printf("[%s %d] Ready to paste image", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			// [Dst]: Setup clipboard and DO NOT send msg
			rtkClipboard.SetupDstPasteImage(s.RemoteAddr().String(), "", extData.Data, extData.Header, extData.Size.SizeLow)
		} else {
			log.Printf("[%s %d] Err: Setup past image failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			return false
		}
	} else {
		var msg Peer2PeerMessage
		if buildMessage(&msg, ipAddr, event) {
			writeToSocket(msg, s)
		} else {
			log.Printf("[%s %d] Build message failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			return false
		}
	}
	return true
}

func processFileDrop(s net.Conn, ipAddr string, event EventResult) bool {
	nextState := event.Cmd.State
	nextCommand := event.Cmd.Command

	if nextState == STATE_IO {
		if nextCommand == COMM_SRC {
			// Receive response from dst
			if extData, ok := event.Data.(rtkCommon.FileDropCmd); ok {
				if extData == rtkCommon.FILE_DROP_ACCEPT {
					rtkUtils.GoSafe(func() {processIoWrite(s, ipAddr, event.Cmd.FmtType)}) // [Src]: Start to trans file
				} else if extData == rtkCommon.FILE_DROP_REJECT {
					// TODO: send response to platform (accept or reject)
					rtkFileDrop.ResetFileDropData(ipAddr)
				} else {
					log.Printf("[%s %d] Unknown file drop response type: %s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), extData)
				}
			} else {
				log.Printf("[%s %d] Invalid file drop response: %s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), extData)
			}
		} else if nextCommand == COMM_DST {
			// TODO: refine this flow: put builderMessage together
			var msg Peer2PeerMessage
			if buildMessage(&msg, ipAddr, event) {
				writeToSocket(msg, s)
			} else {
				log.Printf("[%s %d] Build message failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
				return false
			}
		}
	} else if nextState == STATE_TRANS && nextCommand == COMM_DST {
		if extData, ok := event.Data.(rtkCommon.ExtDataFile); ok {
			log.Printf("[%s %d] Ready to accept file", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			// [Dst]: Setup clipboard and DO NOT send msg
			rtkFileDrop.SetupDstFileDrop(ipAddr, s.RemoteAddr().String(), extData.FilePath, rtkPlatform.GetPlatform(), extData.Size.SizeHigh, extData.Size.SizeLow, time.Now().UnixMilli())
		} else {
			log.Printf("[%s %d] Err: Setup file drop failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			return false
		}
	} else {
		var msg Peer2PeerMessage
		if buildMessage(&msg, ipAddr, event) {
			writeToSocket(msg, s)
		} else {
			log.Printf("[%s %d] Build message failed", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			return false
		}
	}
	return true
}

func processTask(curState *StateType, curCommand *CommandType, s net.Conn, ipAddr string, event EventResult) {
	ret := true
	switch event.Cmd.FmtType {
	case rtkCommon.TEXT_CB:
		log.Println("ProcessTextCB triggered")
		ret = processTextCB(s, ipAddr, event)
	case rtkCommon.IMAGE_CB:
		log.Println("ProcessImageCB triggered")
		ret = processImageCB(s, ipAddr, event)
	case rtkCommon.FILE_DROP:
		log.Println("ProcessFileDrop triggered")
		ret = processFileDrop(s, ipAddr, event)
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
		log.Printf("[%s %d] Invalid state: cur(%s, %s), next(%s, %s)", rtkUtils.GetFuncName(), rtkUtils.GetLine(), curState, curCommand, nextState, nextCommand)
	}
	return ret
}


type ReadSocketMode struct {
	AllowReadMsg bool
	DataType	 rtkCommon.TransFmtType
}

func setMsgSocketMode(data *atomic.Value) {
	data.Store(ReadSocketMode{true, rtkCommon.TEXT_CB})
}

func setIoSocketMode(data *atomic.Value, dataType rtkCommon.TransFmtType) {
	data.Store(ReadSocketMode{false, dataType})
}

func isMsgSocketMode(data *atomic.Value) bool {
	return data.Load().(ReadSocketMode).AllowReadMsg
}

func getIoSocketFmtType(data *atomic.Value) rtkCommon.TransFmtType {
	return data.Load().(ReadSocketMode).DataType
}

func ProcessEventsForPeer(s net.Conn, ipAddr string, ctx context.Context, cancelFunc context.CancelFunc) {
	curState := STATE_INIT
	curCommand := COMM_INIT

	var readSocketMode atomic.Value
	setMsgSocketMode(&readSocketMode)
	eventResultClipboard := make(chan EventResult)
	eventResultFileDrop := make(chan EventResult)
	eventResultSocket := make(chan EventResult)
	errChan := make(chan rtkCommon.SocketErr)
	rtkUtils.GoSafe(func() {HandleClipboardEvent(ctx, &readSocketMode, eventResultClipboard, s, ipAddr)})
	rtkUtils.GoSafe(func() {HandleFileDropEvent(ctx, &readSocketMode, eventResultFileDrop, ipAddr)})
	rtkUtils.GoSafe(func() {HandleReadFromSocket(ctx, &readSocketMode, eventResultSocket, s, ipAddr, errChan)})

	handleEvent := func(event EventResult) {
		buildState := curState
		buildCommand := curCommand

		if buildCmd(&curState, &curCommand, event, &buildState, &buildCommand) {
			// Move to next state and process
			event.Cmd.State = buildState
			event.Cmd.Command = buildCommand
			processTask(&curState, &curCommand, s, ipAddr, event)
		}
	}
	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s %d] Err: End of context", rtkUtils.GetFuncName(), rtkUtils.GetLine())
			if rtkClipboard.GetLastClipboardData().SourceID == s.RemoteAddr().String() {
				rtkClipboard.ResetLastClipboardData()
			}
			return
		default:
			select {
			case err := <-errChan:
				log.Printf("[%s %d] Err: Socket error: %d", rtkUtils.GetFuncName(), rtkUtils.GetLine(), err)
				cancelFunc()
			case event := <-eventResultClipboard:
				log.Printf("[ProcessEvent Clipboard][%s] EventResult fmt=%s, state=%s, cmd=%s", ipAddr, event.Cmd.FmtType, event.Cmd.State, event.Cmd.Command)
				handleEvent(event)
			case event := <-eventResultFileDrop:
				log.Printf("[ProcessEvent FileDrop][%s] EventResult fmt=%s, state=%s, cmd=%s", ipAddr, event.Cmd.FmtType, event.Cmd.State, event.Cmd.Command)
				handleEvent(event)
			case event := <-eventResultSocket:
				log.Printf("[ProcessEvent Socket][%s] EventResult fmt=%s, state=%s, cmd=%s", ipAddr, event.Cmd.FmtType, event.Cmd.State, event.Cmd.Command)
				handleEvent(event)
			}
		}
	}
}
