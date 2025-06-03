package p2p

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	rtkBuildConfig "rtk-cross-share/buildConfig"
	rtkClipboard "rtk-cross-share/clipboard"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
)

func P2PRead(s net.Conn, ipAddr string, ctx context.Context, cancelFunc context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			rtkGlobal.Handler.CtxMutex.Lock()
			if rtkGlobal.Handler.State.State != rtkCommon.UNINIT && rtkGlobal.Handler.State.State != rtkCommon.PROCESSING_TRAN_ING {
				rtkGlobal.Handler.CtxMutex.Unlock()
				time.Sleep(100 * time.Millisecond)
				continue
			}
			rtkGlobal.Handler.CtxMutex.Unlock()

			var msg rtkCommon.P2PMessage
			errSocket := ReadFromSocket(&msg, s)
			if errSocket != rtkCommon.OK {
				if errSocket == rtkCommon.ERR_NETWORK {
					cancelFunc()
					return
				} else {
					time.Sleep(3 * time.Second)
					continue
				}
			}

			//log.Println("read msg = ", msg)
			if !HandleDataTransferRead(s, ipAddr, msg) {
				HandleDataPasteToCBProcess(msg, ipAddr)
			}
		}
	}
}

func HandleDataTransferRead(s net.Conn, ipAddr string, gotMsg rtkCommon.P2PMessage) bool {
	ip, _ := rtkUtils.SplitIP(ipAddr)

	if (rtkGlobal.Handler.SourceIP != "") && (ip != rtkGlobal.Handler.SourceIP) && (rtkGlobal.Handler.SourceIP != rtkGlobal.NodeInfo.IPAddr.PublicIP) {
		log.Printf("HandleDataTransferRead ip %v source ip %v nodeInfo ip %v  return false", ip, rtkGlobal.Handler.SourceIP, rtkGlobal.NodeInfo.IPAddr.PublicIP)
		return false
	}

	rtkGlobal.Handler.CtxMutex.Lock()
	defer rtkGlobal.Handler.CtxMutex.Unlock()

	if rtkGlobal.Handler.State.State == rtkCommon.UNINIT {
		if gotMsg.InbandCmd == rtkCommon.FILE_REQ {
			rtkGlobal.Handler.State.State = rtkCommon.SRC_INIT
			err := HandleSrcDataTransfer(s, ipAddr, gotMsg.FmtType)
			if err != nil {
				var msg = rtkCommon.P2PMessage{
					SourceID:  rtkGlobal.NodeInfo.ID,
					InbandCmd: rtkCommon.FILE_TRAN_CANCEL_SRC,
					Hash:      "",
					PacketID:  0,
					FmtType:   "",
				}
				WriteToSocket(msg, s)
			}
			return true
		} else if gotMsg.InbandCmd == rtkCommon.FILE_DROP_INITIATE {
			targetCbData, _ := rtkUtils.GetNodeCBData(ipAddr)
			var cbData = rtkCommon.ClipBoardData{
				SourceID: gotMsg.SourceID,
				Hash:     gotMsg.Hash,
				FmtType:  gotMsg.FmtType,
				Content:  gotMsg.Buf,
				CopySize: gotMsg.CopySize,
			}
			if cbData.Hash != targetCbData.Hash {
				rtkGlobal.Handler.SourceID = gotMsg.SourceID
				rtkGlobal.Handler.SourceIP, _ = rtkUtils.SplitIP(ipAddr)
				rtkGlobal.Handler.CopyDataSize.SizeHigh = gotMsg.CopySize.SizeHigh
				rtkGlobal.Handler.CopyDataSize.SizeLow = gotMsg.CopySize.SizeLow
				rtkPlatform.GoSetupFileDrop(gotMsg.SourceID, string(gotMsg.Buf), gotMsg.SourcePlatform, gotMsg.CopySize.SizeHigh, gotMsg.CopySize.SizeLow)

				// Update all clipboard for every connected nodes
				rtkGlobal.CBData.Range(func(key, value interface{}) bool {
					rtkGlobal.CBData.Store(key, cbData)
					return true
				})
			}

			return true
		}
	} else if rtkGlobal.Handler.State.State == rtkCommon.PROCESSING_TRAN_ING {
		if gotMsg.InbandCmd == rtkCommon.FILE_REQ_ACK {
			rtkGlobal.Handler.State.State = rtkCommon.PROCESSING_TRAN_ING_ACK
			return true
		}
	}

	return false
}

func OpenSrcFile() error {
	var err error
	rtkGlobal.Handler.SrcFile, err = os.OpenFile(rtkGlobal.Handler.CopyFilePath.Load().(string), os.O_RDONLY, 0644)
	if err != nil {
		rtkGlobal.Handler.SrcFile = nil
		log.Println("Error opening file:", err, "filePath:", rtkGlobal.Handler.CopyFilePath)
		return err
	}
	_, errSeek := rtkGlobal.Handler.SrcFile.Seek(0, io.SeekStart)
	if errSeek != nil {
		rtkGlobal.Handler.SrcFile = nil
		log.Println("Error seek file:", err)
		return errSeek
	}
	log.Println("OpenSrcFile")
	return nil
}

func CloseSrcFile() {
	if rtkGlobal.Handler.SrcFile == nil {
		return
	}
	rtkGlobal.Handler.SrcFile.Close()
	rtkGlobal.Handler.SrcFile = nil
	log.Println("CloseSrcFile")
}

func HandleSrcDataTransfer(s net.Conn, ipAddr string, fmtType rtkCommon.ClipboardFmtType) error {
	defer func() {
		rtkGlobal.Handler.State.State = rtkCommon.UNINIT
		rtkGlobal.Handler.DstFilePath = ""
		rtkGlobal.Handler.DstFile.Close()
		rtkGlobal.Handler.DstFile = nil
		rtkGlobal.Handler.AppointIpAddr = ""
	}()

	var msg = rtkCommon.P2PMessage{
		SourceID:  rtkGlobal.NodeInfo.ID,
		InbandCmd: rtkCommon.FILE_REQ_ACK,
		Hash:      "",
		PacketID:  0,
		FmtType:   "",
	}
	WriteToSocket(msg, s)

	// consider transfer data after receive ACK
	time.Sleep(rtkGlobal.RTT[ipAddr] + 2*time.Second)
	startTime := time.Now().UnixNano()
	if fmtType == rtkCommon.FILE {
		errOpenFile := OpenSrcFile()
		if errOpenFile != nil {
			return errOpenFile
		}

		defer CloseSrcFile()
		log.Println("(SRC) Start to copy file ...")
		// TODO: write data timeout
		// s.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err := io.Copy(s, rtkGlobal.Handler.SrcFile)
		// s.SetWriteDeadline(time.Time{})
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("Error sending file timeout:", netErr)
			} else {
				log.Println("Error sending file:", err)
			}
			return err
		}
		log.Printf("(SRC) End to copy file, use [%d] ms ...", (time.Now().UnixNano()-startTime)/1e6)
	} else if fmtType == rtkCommon.IMAGE {
		log.Printf("(SRC) Start to copy img, len[%d]...", len(rtkGlobal.Handler.CopyImgData))
		_, err := io.Copy(s, bytes.NewReader(rtkGlobal.Handler.CopyImgData))
		if err != nil {
			log.Println("Error sending img:", err)
			return err
		}
		log.Printf("(SRC) End to copy img, use [%d] ms ...", (time.Now().UnixNano()-startTime)/1e6)
	}

	return nil
}

func HandleDataPasteToCBProcess(msg rtkCommon.P2PMessage, ipAddr string) {
	var cbData = rtkCommon.ClipBoardData{
		SourceID: msg.SourceID,
		Hash:     msg.Hash,
		FmtType:  msg.FmtType,
		Content:  msg.Buf,
		CopySize: msg.CopySize,
	}

	targetCbData, _ := rtkUtils.GetNodeCBData(ipAddr)
	if cbData.Hash != targetCbData.Hash {
		log.Printf("HandleDataPasteToCBProcess Clipboard changed at %s: FmtType:%s cmd:%s addr:%s len:[%d]", cbData.Hash, cbData.FmtType, msg.InbandCmd, ipAddr, len(cbData.Content))

		if cbData.FmtType == rtkCommon.TEXT {
			go rtkPlatform.GoSetupDstPasteText([]byte(cbData.Content))
		} else if cbData.FmtType == rtkCommon.FILE && msg.InbandCmd == rtkCommon.TEXT_TRAN {
			rtkGlobal.Handler.SourceID = cbData.SourceID
			rtkGlobal.Handler.SourceIP, _ = rtkUtils.SplitIP(ipAddr)
			rtkGlobal.Handler.CopyDataSize.SizeHigh = msg.CopySize.SizeHigh
			rtkGlobal.Handler.CopyDataSize.SizeLow = msg.CopySize.SizeLow
			rtkClipboard.ResetClipboard(rtkCommon.CLIPBOARD_RESET_TYPE_IMAGE | rtkCommon.CLIPBOARD_RESET_TYPE_FILE) // reset local clipboard data
			go rtkPlatform.GoSetupDstPasteFile(msg.SourceID, string(cbData.Content), msg.SourcePlatform, msg.CopySize.SizeHigh, msg.CopySize.SizeLow)
		} else if cbData.FmtType == rtkCommon.IMAGE && msg.InbandCmd == rtkCommon.TEXT_TRAN {
			rtkGlobal.Handler.SourceID = cbData.SourceID
			rtkGlobal.Handler.SourceIP, _ = rtkUtils.SplitIP(ipAddr)
			rtkGlobal.Handler.CopyDataSize.SizeHigh = msg.CopySize.SizeHigh
			rtkGlobal.Handler.CopyDataSize.SizeLow = msg.CopySize.SizeLow
			rtkClipboard.ResetClipboard(rtkCommon.CLIPBOARD_RESET_TYPE_IMAGE | rtkCommon.CLIPBOARD_RESET_TYPE_FILE) // reset local clipboard data
			go rtkPlatform.GoSetupDstPasteImage(msg.SourceID, cbData.Content, msg.CopyImgHeader, msg.CopySize.SizeLow)
		}

		// Update all clipboard for every connected nodes
		rtkGlobal.CBData.Range(func(key, value interface{}) bool {
			rtkGlobal.CBData.Store(key, cbData)
			return true
		})
	}
}

func WatchP2PFileTransferState(ctx context.Context, resultChan chan<- bool) {
	cur_state := rtkGlobal.Handler.State.State
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			if cur_state != rtkGlobal.Handler.State.State {
				log.Println("WatchP2PFileTransferState change: ", rtkGlobal.Handler.State.State)
				cur_state = rtkGlobal.Handler.State.State
				if cur_state != rtkCommon.UNINIT {
					resultChan <- true
				}
			} else {
				resultChan <- false
			}
		}
	}
}

func HandleDataCopyProcess(s net.Conn, cbData rtkCommon.ClipBoardData, ipAddr string) {
	targetCbData, _ := rtkUtils.GetNodeCBData(ipAddr)
	if cbData.Hash != targetCbData.Hash {
		if cbData.FmtType == rtkCommon.FILE || cbData.FmtType == rtkCommon.IMAGE {
			if rtkGlobal.Handler.AppointIpAddr != "" && ipAddr != rtkGlobal.Handler.AppointIpAddr {
				log.Println("specify the transmission peer is: ", rtkGlobal.Handler.AppointIpAddr, " peer:", ipAddr, " continue")
				return
			}
		}

		rtkGlobal.CBData.Store(ipAddr, cbData)
		rtkGlobal.Handler.SourceID = cbData.SourceID
		rtkGlobal.Handler.SourceIP = rtkGlobal.NodeInfo.IPAddr.PublicIP

		var msg = rtkCommon.P2PMessage{
			SourceID:       rtkGlobal.NodeInfo.ID,
			SourcePlatform: rtkBuildConfig.Platform,
			InbandCmd:      rtkCommon.TEXT_TRAN,
			Hash:           cbData.Hash,
			PacketID:       0,
			FmtType:        cbData.FmtType,
		}

		if cbData.FmtType == rtkCommon.FILE {
			msg.CopySize.SizeHigh = rtkGlobal.Handler.CopyDataSize.SizeHigh
			msg.CopySize.SizeLow = rtkGlobal.Handler.CopyDataSize.SizeLow
			msg.Buf = []byte(filepath.Base(string(cbData.Content)))
		} else if cbData.FmtType == rtkCommon.IMAGE {
			msg.CopySize.SizeHigh = rtkGlobal.Handler.CopyDataSize.SizeHigh
			msg.CopySize.SizeLow = rtkGlobal.Handler.CopyDataSize.SizeLow
			msg.CopyImgHeader = rtkGlobal.Handler.CopyImgHeader
			msg.Buf = []byte{}
		} else {
			msg.Buf = cbData.Content
		}

		log.Printf("HandleDataCopyProcess Clipboard changed ,WriteToSocket, hash:%s  FmtType: %s, %s, content:[%s]", cbData.Hash, msg.FmtType, ipAddr, string(msg.Buf))
		WriteToSocket(msg, s)
	}
}

func P2PWrite(s net.Conn, ipAddr string, ctx context.Context) {

	select {
	case <-ctx.Done():
		return
	default:
		watchCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		resultChanText := make(chan rtkCommon.ClipBoardData)
		resultChanFile := make(chan rtkCommon.ClipBoardData)
		resultChanImg := make(chan rtkCommon.ClipBoardData)
		resultChanIsDataTranState := make(chan bool)
		go rtkClipboard.WatchClipboardText(watchCtx, resultChanText)
		//go rtkClipboard.WatchClipboardFiles(watchCtx, resultChanFile, ipAddr)
		go rtkClipboard.WatchClipboardImg(watchCtx, resultChanImg)
		go WatchP2PFileTransferState(watchCtx, resultChanIsDataTranState)
		for {
			select {
			case <-ctx.Done():
				return
			case isDataTranState := <-resultChanIsDataTranState:
				if isDataTranState == true {
					HandleDataTransferWrite(s, ipAddr)
				}
			case cbData := <-resultChanText:
				HandleDataCopyProcess(s, cbData, ipAddr)
			case cbData := <-resultChanFile:
				HandleDataCopyProcess(s, cbData, ipAddr)
			case cbData := <-resultChanImg:
				HandleDataCopyProcess(s, cbData, ipAddr)
			}
		}
	}
}

func OpenDstFile() {
	var err error
	rtkGlobal.Handler.DstFile, err = os.OpenFile(rtkGlobal.Handler.DstFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Error opening Dst file err: %v path: %v", err, rtkGlobal.Handler.DstFilePath)
		return
	}
	log.Println("OpenDstFile")
}

func ReadFromSocket(msg *rtkCommon.P2PMessage, s net.Conn) rtkCommon.SocketErr {
	buffer := make([]byte, 65535)
	n, err := s.Read(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			log.Println("Read fail network error", netErr.Error())
			return rtkCommon.ERR_NETWORK
		} else {
			log.Println("Read fail", err.Error())
			return rtkCommon.ERR_OTHER
		}
	}

	buffer = buffer[:n]
	buffer = bytes.Trim(buffer, "\x00")
	buffer = bytes.Trim(buffer, "\x13")

	err = json.Unmarshal(buffer, msg)
	if err != nil {
		log.Println("Failed to unmarshal P2PMessage data", err.Error())
		log.Printf("Err JSON len[%d] data:[%s] ", n, string(buffer))
		rtkUtils.WriteErrJson(s.RemoteAddr().String(), buffer)
		return rtkCommon.ERR_JSON
	}

	// log.Println("Read msg from:", s.RemoteAddr(), "InbandCmd =", msg.InbandCmd, "FmtType =", msg.FmtType, "Buf =", msg.Buf, "Buf size=", len(msg.Buf))
	return rtkCommon.OK
}

func WriteToSocket(msg rtkCommon.P2PMessage, s net.Conn) rtkCommon.SocketErr {
	log.Println("Write msg to:", s.RemoteAddr(), "InbandCmd =", msg.InbandCmd)
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

func WriteDstFile(content []byte) {
	if rtkGlobal.Handler.DstFile != nil {
		if _, err := rtkGlobal.Handler.DstFile.Write(content); err != nil {
			log.Println("Error writing to dst file:", err)
			return
		}
		rtkPlatform.GoUpdateProgressBar(len(content))
	} else {
		rtkPlatform.GoDataTransfer(content)
	}
}

func HandleDataTransferError(inbandCmd rtkCommon.P2PFileTransferInbandEnum) {
	switch inbandCmd {
	case rtkCommon.FILE_TRAN_CANCEL_SRC:
		rtkPlatform.GoEventHandle(rtkCommon.EVENT_TYPE_OPEN_FILE_ERR)
	case rtkCommon.FILE_TRAN_CANCEL_DST:
		rtkPlatform.GoEventHandle(rtkCommon.EVENT_TYPE_RECV_TIMEOUT)
	default:
		log.Println("[DataTransferError]: Unhandled type")
	}
}

func IsTransferError(buffer []byte) bool {
	var msg rtkCommon.P2PMessage
	var js json.RawMessage
	if json.Unmarshal(buffer, &js) != nil {
		return false
	}

	data := bytes.Trim(buffer, "\x00")
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return false
	}

	if msg.InbandCmd == rtkCommon.FILE_TRAN_CANCEL_SRC {
		return true
	}
	return false
}

func CloseDstFile() {
	if rtkGlobal.Handler.DstFile == nil {
		return
	}

	log.Println("CloseDstFile")
	rtkGlobal.Handler.DstFile.Close()
	rtkGlobal.Handler.DstFile = nil
	rtkGlobal.Handler.DstFilePath = ""
	rtkGlobal.Handler.CopyFileName = ""
	rtkPlatform.GoDeinitProgressBar()
}

func ClearCopySourceData() {
	rtkGlobal.Handler.SourceIP = ""
	rtkGlobal.Handler.SourceID = ""
	rtkGlobal.Handler.CopyDataSize.SizeHigh = 0
	rtkGlobal.Handler.CopyDataSize.SizeLow = 0
}

func HandleDataTransferWrite(s net.Conn, ipAddr string) bool {
	ip, _ := rtkUtils.SplitIP(ipAddr)
	if (ip != rtkGlobal.Handler.SourceIP) && (rtkGlobal.Handler.SourceIP != rtkGlobal.NodeInfo.IPAddr.PublicIP) {
		log.Printf("HandleDataTransferWrite ip %v source ip %v nodeInfo ip %v  return false", ip, rtkGlobal.Handler.SourceIP, rtkGlobal.NodeInfo.IPAddr.PublicIP)
		return false
	}

	if rtkGlobal.Handler.State.State == rtkCommon.FILE_DROP_INIT &&
		rtkGlobal.Handler.AppointIpAddr != "" && ipAddr != rtkGlobal.Handler.AppointIpAddr {
		log.Println("specify the transmission peer is: ", rtkGlobal.Handler.AppointIpAddr, " peer:", ipAddr, " continue")
		return false
	}

	rtkGlobal.Handler.CtxMutex.Lock()
	defer rtkGlobal.Handler.CtxMutex.Unlock()

	targetCbData, _ := rtkUtils.GetNodeCBData(ipAddr)
	var msg = rtkCommon.P2PMessage{
		SourceID:       rtkGlobal.NodeInfo.ID,
		SourcePlatform: rtkBuildConfig.Platform,
		InbandCmd:      "",
		Hash:           "",
		PacketID:       0,
		FmtType:        targetCbData.FmtType,
	}

	if rtkGlobal.Handler.State.State == rtkCommon.DEST_INIT {
		rtkGlobal.Handler.State.State = rtkCommon.PROCESSING_TRAN_ING
		msg.InbandCmd = rtkCommon.FILE_REQ
		msg.CopySize.SizeHigh = rtkGlobal.Handler.CopyDataSize.SizeHigh
		msg.CopySize.SizeLow = rtkGlobal.Handler.CopyDataSize.SizeLow
		if msg.FmtType == rtkCommon.FILE {
			OpenDstFile()
		}
		WriteToSocket(msg, s)
		return true
	} else if rtkGlobal.Handler.State.State == rtkCommon.PROCESSING_TRAN_ING_ACK {
		fileSize := int64(rtkGlobal.Handler.CopyDataSize.SizeHigh)<<32 | int64(rtkGlobal.Handler.CopyDataSize.SizeLow)
		log.Printf("(DST) Start to Copy  size:[%d]...", fileSize)
		receivedBytes := int64(0)
		buffer := make([]byte, 32*1024)

		startTime := time.Now().UnixNano()
		// TODO: read data timeout
		// s.SetReadDeadline(time.Now().Add(10 * time.Second))
		for receivedBytes < fileSize {
			n, err := s.Read(buffer)
			log.Println("Receive data size:", n)
			if err != nil && err != io.EOF {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					var msg = rtkCommon.P2PMessage{
						SourceID:  rtkGlobal.NodeInfo.ID,
						InbandCmd: rtkCommon.FILE_TRAN_CANCEL_DST,
						Hash:      "",
						PacketID:  0,
						FmtType:   "",
					}
					WriteToSocket(msg, s)
					HandleDataTransferError(rtkCommon.FILE_TRAN_CANCEL_DST)
				}
				log.Println("Error reading from connection:", err)
				break
			}
			if n == 0 {
				break
			}

			// Source cancel transfer file
			if IsTransferError(buffer[:n]) {
				HandleDataTransferError(rtkCommon.FILE_TRAN_CANCEL_SRC)
				break
			}
			WriteDstFile(buffer[:n])
			receivedBytes += int64(n)
		}
		// s.SetReadDeadline(time.Time{})
		log.Printf("(DST) End to Copy, use [%d] ms...", (time.Now().UnixNano()-startTime)/1e6)
		rtkPlatform.ReceiveCopyDataDone(msg.FmtType, fileSize)
		if msg.FmtType == rtkCommon.FILE {
			CloseDstFile()
		}
		ClearCopySourceData()
		rtkGlobal.Handler.State.State = rtkCommon.UNINIT
		return true
	} else if rtkGlobal.Handler.State.State == rtkCommon.FILE_DROP_INIT {
		msg.FmtType = rtkCommon.FILE
		msg.InbandCmd = rtkCommon.FILE_DROP_INITIATE
		msg.CopySize.SizeHigh = rtkGlobal.Handler.CopyDataSize.SizeHigh
		msg.CopySize.SizeLow = rtkGlobal.Handler.CopyDataSize.SizeLow
		msg.Buf = []byte(filepath.Base(rtkGlobal.Handler.CopyFilePath.Load().(string)))
		hash, err := rtkUtils.CreateMD5Hash(msg.Buf)
		if err != nil {
			log.Fatalf("Failed to create hash: %v", err)
		} else {
			msg.Hash = hash.B58String()
		}

		cbData := rtkCommon.ClipBoardData{
			SourceID: rtkGlobal.NodeInfo.ID,
			FmtType:  rtkCommon.FILE,
			Hash:     msg.Hash,
			Content:  msg.Buf,
			CopySize: rtkCommon.FileSize{
				SizeHigh: msg.CopySize.SizeHigh,
				SizeLow:  msg.CopySize.SizeLow,
			},
		}
		rtkGlobal.CBData.Store(ipAddr, cbData)

		log.Printf("(SRC) Init file-drop filename:[%s] size:[%d] ...", string(msg.Buf), msg.CopySize.SizeLow)
		WriteToSocket(msg, s)
		rtkGlobal.Handler.State.State = rtkCommon.UNINIT
		return true
	}

	return false
}
