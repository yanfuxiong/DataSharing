package peer2peer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-yamux/v5"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkConnection "rtk-cross-share/client/connection"
	rtkFileDrop "rtk-cross-share/client/filedrop"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strconv"
	"sync"
	"time"
)

const (
	copyBufSize       = 256 << 10 // 256KB
	truncateThreshold = 32 << 20  // 32MB

	interruptFailureInterval = 60 //seconds, Interrupt file data transfer time out: 60s
)

var (
	fileTransferInfoMap sync.Map //key: ID
)

type fileTransferInfo struct {
	cancelFn func(rtkCommon.CancelBusinessSource)
}

type cancelableReader struct {
	realReader io.Reader
	ctx        context.Context
}

type cancelableWriter struct {
	realWriter io.Writer
	ctx        context.Context
}

func (cRead *cancelableReader) Read(p []byte) (int, error) {
	select {
	case <-cRead.ctx.Done():
		log.Printf("[%s] cancel by cancelableReader!", rtkMisc.GetFuncInfo())
		return 0, cRead.ctx.Err()
	default:
		return cRead.realReader.Read(p) //maybe block here
	}
}

func (cWrite *cancelableWriter) Write(p []byte) (int, error) {
	select {
	case <-cWrite.ctx.Done():
		log.Printf("[%s] cancel by cancelableWriter!", rtkMisc.GetFuncInfo())
		return 0, cWrite.ctx.Err()
	default:
		return cWrite.realWriter.Write(p) //maybe block here
	}
}

func OpenSrcFile(file **os.File, filePath string, offset int64) rtkMisc.CrossShareErr {
	if *file != nil {
		(*file).Close()
		*file = nil
	}

	if !rtkMisc.FileExists(filePath) {
		log.Println("Error file not exist: ", filePath)
		return rtkMisc.ERR_BIZ_FT_FILE_NOT_EXISTS
	}

	var err error
	*file, err = os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		*file = nil
		log.Println("Error opening file:", err, "filePath:", filePath)
		return rtkMisc.ERR_BIZ_FD_SRC_OPEN_FILE
	}
	_, errSeek := (*file).Seek(offset, io.SeekStart)
	if errSeek != nil {
		(*file).Close()
		*file = nil
		log.Println("Error seek file:", err)
		return rtkMisc.ERR_BIZ_FD_SRC_FILE_SEEK
	}

	return rtkMisc.SUCCESS
}

func CloseFile(file **os.File) {
	if *file == nil {
		return
	}
	(*file).Close()
	*file = nil
}

func OpenDstFile(file **os.File, filePath string) error {
	if *file != nil {
		(*file).Close()
		*file = nil
	}

	var err error
	*file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Error opening Dst file path: %s err: %v ", filePath, err)
		*file = nil
		return err
	}
	return nil
}

func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		log.Printf("Remove file:[%s] err:%+v", filePath, err)
		return err
	}
	log.Printf("Remove file:[%s] success!", filePath)
	return nil
}

func SetFileTransferInfo(id string, fn func(rtkCommon.CancelBusinessSource)) {
	fileTransferInfoMap.Store(id, fileTransferInfo{
		cancelFn: fn,
	})
}

func CancelSrcFileTransfer(id, ipAddr string, timestamp uint64, errCode rtkMisc.CrossShareErr) {
	if errCode == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI {
		log.Printf("(SRC) [%s] IP:[%s] timestamp:[%d] Copy file operation was canceled by dst GUI !", rtkMisc.GetFuncInfo(), ipAddr, timestamp)
	} else {
		log.Printf("(SRC) [%s] IP:[%s] timestamp:[%d] Copy file operation was canceled by dst errCode:%d!", rtkMisc.GetFuncInfo(), ipAddr, timestamp, errCode)
		if errCode == rtkMisc.ERR_BIZ_FT_DST_OPEN_STREAM {
			rtkPlatform.GoNotifyErrEvent(id, errCode, ipAddr, strconv.Itoa(int(timestamp)), "", "") // notice  errCode to platform
			return
		}
	}
	if rtkFileDrop.IsFileTransInProgress(id, timestamp) {
		if value, ok := fileTransferInfoMap.Load(id); ok {
			fileInfo := value.(fileTransferInfo)
			fileTransferInfoMap.Store(id, fileInfo)
			fileInfo.cancelFn(rtkCommon.FileTransDstCancel)
			log.Printf("(SRC) [%s] ID:[%s] Cancel FileTransfer success, timestamp:%d", rtkMisc.GetFuncInfo(), id, timestamp)
		}
	} else {
		if rtkFileDrop.CancelFileTransFromCacheMap(id, timestamp) {
			rtkPlatform.GoNotifyErrEvent(id, errCode, ipAddr, strconv.Itoa(int(timestamp)), "", "") // notice  errCode to platform
			rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP)
			rtkConnection.CloseFileDropItemStream(id, timestamp)
		}
	}
}

func CancelDstFileTransfer(id, ipAddr string, timestamp uint64, errCode rtkMisc.CrossShareErr) {
	if errCode == rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_GUI {
		log.Printf("(DST) [%s] IP:[%s] timestamp:[%d] Copy file operation was canceled by src GUI !", rtkMisc.GetFuncInfo(), ipAddr, timestamp)
	} else {
		log.Printf("(DST) [%s] IP:[%s] timestamp:[%d] Copy file operation was canceled by src errCode:%d!", rtkMisc.GetFuncInfo(), ipAddr, timestamp, errCode)
	}
	if rtkFileDrop.IsFileTransInProgress(id, timestamp) {
		if value, ok := fileTransferInfoMap.Load(id); ok {
			fileInfo := value.(fileTransferInfo)
			fileTransferInfoMap.Store(id, fileInfo)
			fileInfo.cancelFn(rtkCommon.FileTransSrcCancel)
			log.Printf("(DST) [%s] ID:[%s] Cancel FileTransfer success, timestamp:%d", rtkMisc.GetFuncInfo(), id, timestamp)
		}
	} else {
		if rtkFileDrop.CancelFileTransFromCacheMap(id, timestamp) {
			rtkPlatform.GoNotifyErrEvent(id, errCode, ipAddr, strconv.Itoa(int(timestamp)), "", "") // notice  errCode to platform
			rtkConnection.CloseFileDropItemStream(id, timestamp)
		}
	}
}

func getFileDataSendCancelErrCode(ctx context.Context, ipAddr string, timeStamp uint64) rtkMisc.CrossShareErr {
	return getFileTransferCancelErrCode(ctx, ipAddr, timeStamp, true)
}

func getFileDataReceiveCancelErrCode(ctx context.Context, ipAddr string, timeStamp uint64) rtkMisc.CrossShareErr {
	return getFileTransferCancelErrCode(ctx, ipAddr, timeStamp, false)
}

func getFileTransferCancelErrCode(ctx context.Context, ipAddr string, timeStamp uint64, isSrc bool) (errCode rtkMisc.CrossShareErr) {
	if source, ok := rtkUtils.GetCancelSource(ctx); ok {
		if isSrc {
			log.Printf("(SRC) IP:[%s] timeStamp:[%d] file data transfer is cancel source:%d!", ipAddr, timeStamp, source)
		} else {
			log.Printf("(DST) IP:[%s] timeStamp:[%d] file data transfer is cancel source:%d!", ipAddr, timeStamp, source)
		}

		if source == rtkCommon.SourceCablePlugIn ||
			source == rtkCommon.SourceCablePlugOut ||
			source == rtkCommon.SourceVerInvalid ||
			source == rtkCommon.SourceNetworkSwitch ||
			source == rtkCommon.OldP2PBusinessCancel ||
			source == rtkCommon.PeerDisconnectCancel {
			if isSrc {
				errCode = rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS
			} else {
				errCode = rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS
			}

		} else if source == rtkCommon.FileTransDstGuiCancel {
			errCode = rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI
		} else if source == rtkCommon.FileTransSrcGuiCancel {
			errCode = rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_GUI
		}
		return
	} else {
		log.Printf("[%s] IP:[%s] timeStamp:[%d] file data transfer is cancel, Unknown source!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
	}

	if isSrc {
		errCode = rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL
	} else {
		errCode = rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL
	}
	return
}

func dealFilesCacheDataProcess(p2pCtx context.Context, id, ipAddr string, timeStamp uint64) {
	for rtkFileDrop.GetFilesTransferDataCacheCount(id) > 0 {
		cacheData := rtkFileDrop.GetFilesTransferDataItem(id, timeStamp)
		if cacheData == nil {
			break
		}

		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Src {
			resultCode, reTry := writeFileDataItemToSocket(p2pCtx, id, ipAddr, cacheData)
			if resultCode != rtkMisc.SUCCESS {
				if !reTry {
					log.Printf("(SRC) ID[%s] IP[%s] Copy file data To Socket failed, timestamp:%d, ERR code:[%d],  and not resend!", id, ipAddr, cacheData.TimeStamp, resultCode)

					// need notify to dst
					if resultCode != rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI {
						sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_SRC_INTERRUPT, resultCode, cacheData.TimeStamp)
					}

					rtkPlatform.GoNotifyErrEvent(id, resultCode, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
				} else {
					log.Printf("(SRC) ID[%s] IP[%s] Copy file data To Socket is interrupt, timestamp:[%d], wait to resend...", id, ipAddr, cacheData.TimeStamp)
					rtkConnection.CloseAllFileDropStream(id)
					rtkMisc.GoSafe(func() { checkRecoverProcessEventsForPeerAsSrc(id, ipAddr, resultCode) })
					return
				}
			}
			rtkConnection.CloseFileDropItemStream(id, cacheData.TimeStamp)
			rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
			rtkConnection.RemoveFileDropItemStreamListener(cacheData.TimeStamp)
		} else if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Dst {
			resultCode, reTry := readFileDataItemFromSocket(p2pCtx, id, ipAddr, cacheData)
			if resultCode != rtkMisc.SUCCESS {
				if !reTry { // any exceptions and user cancellation need to Notify to platfrom
					log.Printf("(DST) ID[%s] IP[%s] Copy file data To Socket failed, timestamp:%d, ERR code:[%d]  and not retry", id, ipAddr, cacheData.TimeStamp, resultCode)

					// need notify to src
					sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_DST_INTERRUPT, resultCode, cacheData.TimeStamp)
					rtkPlatform.GoNotifyErrEvent(id, resultCode, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
				} else {
					log.Printf("(DST) ID[%s] IP[%s] Copy file data To Socket is interrupt, timestamp:[%d], wait to retry...", id, ipAddr, cacheData.TimeStamp)
					rtkConnection.CloseAllFileDropStream(id)
					//rtkMisc.GoSafe(func() { checkRecoverProcessEventsForPeerAsDst(id, ipAddr, resultCode) })
					return
				}
			}
			rtkConnection.CloseFileDropItemStream(id, cacheData.TimeStamp)
			rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
		} else {
			log.Printf("[%s] ID:[%s] Invalid direction type:[%s], skit it!", rtkMisc.GetFuncInfo(), id, cacheData.FileTransDirection)
		}

		rtkFileDrop.SetFilesCacheItemComplete(id, cacheData.TimeStamp)
	}
}

func writeItemFileDataToSocket(p2pCtx context.Context, id, ipAddr string, fileDropReqData *rtkFileDrop.FilesTransferDataItem) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP) // wait for file drop stream Ready
	if p2pCtx.Err() != nil {                                        // deal file cache Data must return errCode and clear cache when p2p business is canceled
		return getFileDataSendCancelErrCode(p2pCtx, ipAddr, fileDropReqData.TimeStamp)
	}

	var sFileDrop network.Stream
	var ok bool
	if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
		sFileDrop, ok = rtkConnection.GetFileDropItemStream(id, fileDropReqData.TimeStamp)
	} else {
		sFileDrop, ok = rtkConnection.GetFmtTypeStream(id, rtkCommon.FILE_DROP)
	}
	if !ok {
		log.Printf("[%s] Err: Not found file stream by ID:[%s]", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_FD_GET_STREAM_EMPTY
	}

	nTotalFileCnt := uint32(len(fileDropReqData.SrcFileList))
	nTotalFolderCnt := len(fileDropReqData.FolderList)
	if (nTotalFileCnt == 0 && nTotalFolderCnt == 0) || fileDropReqData.TimeStamp == 0 {
		log.Printf("[%s] get file data is invalid! fileCount:[%d] folderCount:[%d] TimeStamp:[%d] ", rtkMisc.GetFuncInfo(), nTotalFileCnt, nTotalFolderCnt, fileDropReqData.TimeStamp)
		return rtkMisc.ERR_BIZ_FD_DATA_INVALID
	}

	ctx, cancel := rtkUtils.WithCancelSource(p2pCtx)
	defer cancel(rtkCommon.FileTransDone)
	SetFileTransferInfo(id, cancel)                   //Used when the receiving end is exception or cancellation
	rtkFileDrop.SetCancelFileTransferFunc(id, fileDropReqData.TimeStamp, cancel) //Used when the user cancels

	progressBar := New64(int64(fileDropReqData.TotalSize))
	var curFileName string
	curFileSize := uint64(0)
	fileDoneCnt := uint32(0)

	rtkMisc.GoSafe(func() {
		barTicker := time.NewTicker(100 * time.Millisecond)
		defer barTicker.Stop()
		barLastBytes := int64(0)
		timeoutBarCnt := int(0)
		barMax := progressBar.GetMax()
		for {
			select {
			case <-ctx.Done():
				//cancel io.Copy maybe block at stream Write, so need interrupt by set deadline
				sFileDrop.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))
				rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP) // old version client maybe trigger this case
				return
			case <-barTicker.C:
				barCurrentBytes := progressBar.GetCurrentBytes()
				if barLastBytes != barCurrentBytes {
					rtkPlatform.GoUpdateSendProgressBar(ipAddr, id, curFileName, fileDoneCnt, nTotalFileCnt, curFileSize, fileDropReqData.TotalSize, uint64(barCurrentBytes), fileDropReqData.TimeStamp)
					barLastBytes = barCurrentBytes
					timeoutBarCnt = 0
				} else {
					timeoutBarCnt++
					if timeoutBarCnt >= 100 { // TODO: Check if it is necessary to determine timeout
						log.Printf("(SRC) [%s] IP[%s] timestamp:[%d] Copy file no data in last 10 seconds!", rtkMisc.GetFuncInfo(), ipAddr, fileDropReqData.TimeStamp)
						timeoutBarCnt = 0
					}
				}

				if barCurrentBytes >= barMax {
					return
				}
			}
		}
	})

	cancelableRead := cancelableReader{
		realReader: nil,
		ctx:        ctx,
	}
	cancelableWrite := cancelableWriter{
		realWriter: sFileDrop,
		ctx:        ctx,
	}

	copyBuffer := make([]byte, copyBufSize)
	isResend := false
	if fileDropReqData.InterruptSrcFileName != "" && fileDropReqData.InterruptLastErrCode != rtkMisc.SUCCESS {
		isResend = true
		log.Printf("(SRC) Retry Copy file data to IP:[%s], id:[%d] file count:[%d] folder count:[%d] totalSize:[%d] TotalDescribe:[%s]...", ipAddr, fileDropReqData.TimeStamp, fileCount, folderCount, fileDropReqData.TotalSize, fileDropReqData.TotalDescribe)
	} else {
		log.Printf("(SRC) Start Copy file data to IP:[%s], id:[%d] file count:[%d] folder count:[%d] totalSize:[%d] TotalDescribe:[%s]...", ipAddr, fileDropReqData.TimeStamp, fileCount, folderCount, fileDropReqData.TotalSize, fileDropReqData.TotalDescribe)
	}

	getInterruptFile := false
	offSet := int64(0)
	for _, fileInfo := range fileDropReqData.SrcFileList {
		fileSize := uint64(fileInfo.FileSize_.SizeHigh)<<32 | uint64(fileInfo.FileSize_.SizeLow)

		if isResend && fileInfo.FileName != fileDropReqData.InterruptSrcFileName && !getInterruptFile {
			continue
		}

		if isResend && fileInfo.FileName == fileDropReqData.InterruptSrcFileName {
			getInterruptFile = true
			offSet = fileDropReqData.InterruptFileOffSet
			if fileDropReqData.InterruptFileOffSet < 0 || fileDropReqData.InterruptFileOffSet > int64(fileSize) {
				log.Printf("[%s] Retry Copy file data to IP:[%s], id:[%d], get invalid interrupt offset:[%d]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropReqData.TimeStamp, fileDropReqData.InterruptFileOffSet)
				needReTry = false
				errCode = rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
				return
			}
			fileSize = fileSize - uint64(fileDropReqData.InterruptFileOffSet)
			log.Printf("(SRC) Retry Copy file data to IP:[%s], id:[%d], Starting from this file:[%s], offset:[%d] ...", ipAddr, fileDropReqData.TimeStamp, fileInfo.FileName, fileDropReqData.InterruptFileOffSet)
		} else {
			offSet = int64(0)
		}

		srcTransResult := writeFileToSocket(id, ipAddr, &cancelableWrite, &cancelableRead, &progressBar, fileInfo.FileName, fileInfo.FilePath, fileSize, fileDropReqData.TimeStamp, offSet, &copyBuffer)
		if srcTransResult != rtkMisc.SUCCESS {
			if srcTransResult == rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS ||
				srcTransResult == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS ||
				srcTransResult == rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_GUI ||
				srcTransResult == rtkMisc.ERR_BIZ_FT_FILE_NOT_EXISTS {
				needReTry = false
			}

			return srcTransResult
		}
	}

	if isResend && !getInterruptFile {
		log.Printf("[%d] Retry Copy file data to IP:[%s], id:[%d], get invalid interrupt src file name:[%s]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropReqData.TimeStamp, fileDropReqData.InterruptSrcFileName)
		return rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
	}
	rtkPlatform.GoUpdateSendProgressBar(ipAddr, id, curFileName, fileDoneCnt, fileCount, curFileSize, fileDropReqData.TotalSize, fileDropReqData.TotalSize, fileDropReqData.TimeStamp)

	log.Printf("(SRC) End Copy all file data to IP:[%s] success, id:[%d] file count:[%d] TotalDescribe:[%s], total use [%d] ms", ipAddr, fileDropReqData.TimeStamp, fileCount, fileDropReqData.TotalDescribe, time.Now().UnixMilli()-startTime)
	ShowNotiMessageSendFileTransferDone(fileDropReqData, id)
	return rtkMisc.SUCCESS
}

func writeFileToSocket(id, ipAddr string, write *cancelableWriter, read *cancelableReader, totalBar **ProgressBar, fileName, filePath string, fileSize, timeStamp uint64, offset int64, buf *[]byte) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()

	var srcFile *os.File
	errCode := OpenSrcFile(&srcFile, filePath, offset)
	if errCode != rtkMisc.SUCCESS {
		log.Printf("[%s] OpenSrcFile err code:[%d]", rtkMisc.GetFuncInfo(), errCode)
		return errCode
	}
	defer CloseFile(&srcFile)
	read.realReader = srcFile

	log.Printf("(SRC) Start to copy file:[%s] size:[%d] ...", fileName, fileSize)
	nCopy := int64(0)
	var err error
	if fileSize > 0 {
		nCopy, err = io.CopyBuffer(io.MultiWriter(*totalBar, write), read, *buf)
		if err != nil {
			if rtkConnection.IsQuicEOF(err) { // cancel by dst
				log.Printf("(SRC) [%s] IP:[%s] timestamp:[%d] quic EOF!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS
			} else if rtkConnection.IsQuicClose(err) { //this case unused
				log.Printf("(SRC) [%s] IP:[%s] timestamp:[%d] quic local Closed!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
				if read.ctx.Err() != nil {
					return getFileDataSendCancelErrCode(read.ctx, ipAddr, timeStamp)
				}
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("(SRC) [%s] IP:[%s] Error sending file timeout:%v", rtkMisc.GetFuncInfo(), ipAddr, netErr)
				if read.ctx.Err() != nil {
					return getFileDataSendCancelErrCode(read.ctx, ipAddr, timeStamp)
				}
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_TIMEOUT
			} else if errors.Is(err, context.Canceled) {
				return getFileDataSendCancelErrCode(read.ctx, ipAddr, timeStamp)
			} else if errors.Is(err, yamux.ErrStreamClosed) || errors.Is(err, yamux.ErrStreamReset) { // old  version client trigger this case
				if read.ctx.Err() != nil {
					return getFileDataSendCancelErrCode(read.ctx, ipAddr, timeStamp)
				}
				log.Printf("(SRC) IP[%s] Copy operation was canceled by close stream!", ipAddr)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL
			} else {
				log.Printf("(SRC) [%s] IP:[%s] timeStamp:[%d] Copy file Error:%+v", rtkMisc.GetFuncInfo(), ipAddr, timeStamp, err)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE
			}
		}

		if !rtkUtils.GetPeerClientIsSupportQueueTrans(id) { //quic no need flush
			bufio.NewWriter(write).Flush()
		}
	}
	log.Printf("(SRC) End to copy file:[%s], size:[%d] use [%d] ms", fileName, nCopy, time.Now().UnixMilli()-startTime)
	return rtkMisc.SUCCESS
}

func readItemFileDataFromSocket(p2pCtx context.Context, id, ipAddr string, fileDropData *rtkFileDrop.FilesTransferDataItem) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	if p2pCtx.Err() != nil { // deal file cache Data must return errCode and clear cache when p2p business is canceled
		return getFileDataReceiveCancelErrCode(p2pCtx, ipAddr, fileDropData.TimeStamp)
	}

	var sFileDrop network.Stream
	var ok bool
	if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
		sFileDrop, ok = rtkConnection.GetFileDropItemStream(id, fileDropData.TimeStamp)
	} else {
		sFileDrop, ok = rtkConnection.GetFmtTypeStream(id, rtkCommon.FILE_DROP)
	}
	if !ok {
		log.Printf("[%s] Err: Not found FileDrop stream by ID: %s", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_FD_GET_STREAM_EMPTY
	}

	if fileDropData.Cmd != rtkCommon.FILE_DROP_ACCEPT {
		log.Printf("[%s] Invalid fildDrop cmd state:%s", rtkMisc.GetFuncInfo(), fileDropData.Cmd)
		return rtkMisc.ERR_BIZ_FD_UNKNOWN_CMD
	}

	nSrcFileCount := uint32(len(fileDropData.SrcFileList))
	folderCount := uint32(len(fileDropData.FolderList))
	if (nSrcFileCount == 0 && folderCount == 0) || fileDropData.TimeStamp == 0 {
		log.Printf("[%s] get file data is invalid! fileCount:[%d] folderCount:[%d] TimeStamp:[%d] ", rtkMisc.GetFuncInfo(), nSrcFileCount, folderCount, fileDropData.TimeStamp)
		return rtkMisc.ERR_BIZ_FD_DATA_INVALID
	}

	isRetry := false //interrupt and retry transmission flag
	if fileDropData.InterruptSrcFileName != "" && fileDropData.InterruptDstFileName != "" && fileDropData.InterruptLastErrCode != rtkMisc.SUCCESS {
		isRetry = true
		log.Printf("(DST) Retry Copy file data from IP:[%s], id:[%d], count:[%d], totalSize:[%d], TotalDescribe:[%s]...", ipAddr, fileDropData.TimeStamp, nSrcFileCount, fileDropData.TotalSize, fileDropData.TotalDescribe)
	} else {
		log.Printf("(DST) Start Copy file data from IP:[%s], id:[%d], count:[%d], totalSize:[%d], TotalDescribe:[%s]...", ipAddr, fileDropData.TimeStamp, nSrcFileCount, fileDropData.TotalSize, fileDropData.TotalDescribe)
	}

	if !isRetry {
		nFolderCount := 0
		for _, dir := range fileDropData.FolderList {
			path := filepath.Join(fileDropData.DstFilePath, rtkMisc.AdaptationPath(dir))
			err := rtkMisc.CreateDir(path, 0755)
			if err != nil {
				log.Printf("[%s] CreateDir:[%s] err:[%+v]", rtkMisc.GetFuncInfo(), path, err)
				continue
			}
			rtkPlatform.GoDragFileListFolderNotify(ipAddr, id, path, fileDropData.TimeStamp)
			nFolderCount++
		}
		if nFolderCount > 0 {
			log.Printf("(DST) Create  %d dir success!", nFolderCount)
		}
	}

	progressBar := New64(int64(fileDropData.TotalSize))
	ctx, cancel := rtkUtils.WithCancelSource(p2pCtx)
	defer cancel(rtkCommon.FileTransDone)
	SetFileTransferInfo(id, cancel)                   //Used when there is an exception on the sending end
	rtkFileDrop.SetCancelFileTransferFunc(id, fileDropData.TimeStamp, cancel) //Used when the recipient cancels

	currentFileSize := uint64(0)
	fileDoneCnt := uint32(0)
	dstFileName := fileDropData.DstFilePath
	rtkMisc.GoSafe(func() {
		barTicker := time.NewTicker(100 * time.Millisecond)
		defer barTicker.Stop()

		barLastBytes := int64(0)
		timeoutBarCnt := int(0)
		barMax := progressBar.GetMax()
		for {
			select {
			case <-ctx.Done():
				//cancel io.Copy maybe block at stream Read, so need interrupt by set deadline
				sFileDrop.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
				rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP) // old version client maybe trigger this case
				return
			case <-barTicker.C:
				barCurrentBytes := progressBar.GetCurrentBytes()
				if barLastBytes != barCurrentBytes {
					rtkPlatform.GoUpdateReceiveProgressBar(ipAddr, id, dstFileName, fileDoneCnt, nSrcFileCount, currentFileSize, fileDropData.TotalSize, uint64(barCurrentBytes), fileDropData.TimeStamp)
					barLastBytes = barCurrentBytes
					timeoutBarCnt = 0
				} else {
					timeoutBarCnt++
					if timeoutBarCnt >= 100 { // TODO: Check if it is necessary to determine timeout
						log.Printf("(DST) [%s] IP[%s] timestamp:[%d] Copy file no data in last 10 seconds!", rtkMisc.GetFuncInfo(), ipAddr, fileDropData.TimeStamp)
						timeoutBarCnt = 0
					}
				}

				if barCurrentBytes >= barMax {
					return
				}
			}
		}
	})

	cancelableRead := cancelableReader{
		realReader: nil,
		ctx:        ctx,
	}
	cancelableWrite := cancelableWriter{
		realWriter: nil,
		ctx:        ctx,
	}

	offset := int64(0)
	getInterruptFile := false
	isRecoverFile := false
	copyBuffer := make([]byte, copyBufSize)
	for i, fileInfo := range fileDropData.SrcFileList {
		currentFileSize = uint64(fileInfo.FileSize_.SizeHigh)<<32 | uint64(fileInfo.FileSize_.SizeLow)
		if isRetry && fileInfo.FileName != fileDropData.InterruptSrcFileName && !getInterruptFile {
			progressBar.Add64(int64(currentFileSize))
			doneFileCnt++
			continue
		}

		var dstFileName string
		if isRetry && fileInfo.FileName == fileDropData.InterruptSrcFileName {
			getInterruptFile = true
			if fileDropData.InterruptFileOffSet < 0 || fileDropData.InterruptFileOffSet > int64(currentFileSize) {
				log.Printf("[%d] Retry Copy file data from IP:[%s], id:[%d], get invalid interrupt offset:[%d]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropData.TimeStamp, fileDropData.InterruptFileOffSet)
				needReTry = false
				errCode = rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
				return
			}
			progressBar.Add64(fileDropData.InterruptFileOffSet)
			receivedBytes := progressBar.GetCurrentBytes()
			rounded := float64(receivedBytes) / float64(fileDropData.TotalSize) * 100
			log.Printf("(DST) Retry Copy file data from IP:[%s], id:[%d], already received:[%d], percentage:[%.2f%%], Starting from this file:[%s], offset:[%d]...", ipAddr, fileDropData.TimeStamp, receivedBytes, rounded, fileInfo.FileName, fileDropData.InterruptFileOffSet)

			dstFileName = fileDropData.InterruptDstFileName
			dstFullFilePath = filepath.Join(fileDropData.DstFilePath, dstFileName)
			isRecoverFile = true
		} else {
			isRecoverFile = false
			dstFileName = rtkMisc.AdaptationPath(fileInfo.FileName)
			dstFullFilePath, dstFileName = rtkUtils.GetTargetDstPathName(filepath.Join(fileDropData.DstFilePath, dstFileName), dstFileName)
		}

		if isRetry && isRecoverFile {
			currentFileSize = currentFileSize - uint64(fileDropData.InterruptFileOffSet)
		}

		cancelableRead.realReader = io.LimitReader(sFileDrop, int64(currentFileSize))
		
		dstTransResult := readFileFromSocket(id, ipAddr, &cancelableWrite, &cancelableRead, &progressBar, currentFileSize, fileDropData.TimeStamp, dstFileName, dstFullFilePath, &copyBuffer, &offset, isRecoverFile)
		if dstTransResult != rtkMisc.SUCCESS {
			if dstTransResult == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS ||
				dstTransResult == rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS ||
				dstTransResult == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI {
				needReTry = false // not need retry
			} else {
				rtkFileDrop.SetFilesTransferDataInterrupt(id, fileInfo.FileName, dstFileName, dstFullFilePath, fileDropData.TimeStamp, offset, dstTransResult)
			}

			return dstTransResult
		}

		doneFileCnt++
		if uint32(i) != (nSrcFileCount - 1) {
			rtkPlatform.GoUpdateReceiveProgressBar(ipAddr, id, dstFullFilePath, doneFileCnt, nSrcFileCount, currentFileSize, fileDropData.TotalSize, uint64(progressBar.GetCurrentBytes()), fileDropData.TimeStamp)
		}
	}

	if isRetry && !getInterruptFile {
		log.Printf("[%d] Retry Copy file data from IP:[%s], id:[%d], get invalid interrupt file name:[%s]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropData.TimeStamp, fileDropData.InterruptSrcFileName)
		return rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
	}

	rtkPlatform.GoUpdateReceiveProgressBar(ipAddr, id, dstFullFilePath, doneFileCnt, nSrcFileCount, currentFileSize, fileDropData.TotalSize, fileDropData.TotalSize, fileDropData.TimeStamp)
	log.Printf("(DST) End Copy file data from IP:[%s] success, id:[%d] count:[%d] totalSize:[%d] totalDescribe:[%s] total use:[%d]ms", ipAddr, fileDropData.TimeStamp, nSrcFileCount, fileDropData.TotalSize, fileDropData.TotalDescribe, time.Now().UnixMilli()-startTime)
	ShowNotiMessageRecvFileTransferDone(fileDropData, id)
	return rtkMisc.SUCCESS
}

func readFileFromSocket(id, ipAddr string, write *cancelableWriter, read *cancelableReader, totalBar **ProgressBar, fileSize, timeStamp uint64, dstFileName, dstFullPath string, buf *[]byte, offset *int64, isRecover bool) rtkMisc.CrossShareErr {
	var dstFile *os.File
	startTime := time.Now().UnixMilli()

	err := OpenDstFile(&dstFile, dstFullPath)
	if err != nil {
		return rtkMisc.ERR_BIZ_FD_DST_OPEN_FILE
	}
	defer CloseFile(&dstFile)

	if isRecover {
		log.Printf("(DST) IP[%s] Retry to copy file:[%s], still has [%d] left ...", ipAddr, dstFileName, fileSize)
	} else {
		log.Printf("(DST) IP[%s] Start to copy file:[%s], size:[%d] ...", ipAddr, dstFileName, fileSize)
	}

	nDstWrite := int64(0)
	if fileSize > 0 {
		if fileSize > uint64(truncateThreshold) && !isRecover {
			dstFile.Truncate(int64(fileSize))
		}
		write.realWriter = dstFile
		nDstWrite, err = io.CopyBuffer(io.MultiWriter(write, *totalBar), read, *buf)
		if err != nil {
			*offset = nDstWrite
			if rtkConnection.IsQuicEOF(err) { // cancel by src
				log.Printf("(DST) [%s] IP:[%s] timestamp:[%d] quic EOF!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS
			} else if rtkConnection.IsQuicClose(err) { //this case unused
				log.Printf("(DST) [%s] IP:[%s] timestamp:[%d] quic local Closed!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
				if read.ctx.Err() != nil {
					return getFileDataReceiveCancelErrCode(read.ctx, ipAddr, timeStamp)
				}
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("(DST) [%s] IP:[%s] Error Read file timeout:%v", rtkMisc.GetFuncInfo(), ipAddr, netErr)
				if read.ctx.Err() != nil {
					return getFileDataReceiveCancelErrCode(read.ctx, ipAddr, timeStamp)
				}
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT
			} else if errors.Is(err, context.Canceled) {
				return getFileDataReceiveCancelErrCode(read.ctx, ipAddr, timeStamp)
			} else if errors.Is(err, yamux.ErrStreamClosed) || errors.Is(err, yamux.ErrStreamReset) { // old  version client trigger this case
				if read.ctx.Err() != nil {
					return getFileDataReceiveCancelErrCode(read.ctx, ipAddr, timeStamp)
				}
				log.Printf("(DST) IP[%s] Copy operation was canceled by close stream!", ipAddr)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL
			} else {
				log.Printf("(DST) [%s] IP:[%s] timeStamp:[%d] Copy file Error:%+v", rtkMisc.GetFuncInfo(), ipAddr, timeStamp, err)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE
			}
		}
	}

	if uint64(nDstWrite) >= fileSize {
		log.Printf("(DST) IP[%s] End to Copy file:[%s] success, total:[%d] use [%d] ms", ipAddr, dstFileName, nDstWrite, time.Now().UnixMilli()-startTime)
		return rtkMisc.SUCCESS
	} else {
		log.Printf("(DST) IP[%s] End to Copy file:[%s] failed, total:[%d], it less then filesize:[%d]...", ipAddr, dstFileName, nDstWrite, fileSize)
		return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_LOSS
	}
}

func ShowNotiMessageSendFileTransferDone(fileDropData *rtkFileDrop.FilesTransferDataItem, id string) {
	clientInfo, err := rtkUtils.GetClientInfo(id)
	if err != nil {
		log.Printf("[%s] %s", rtkMisc.GetFuncInfo(), err.Error())
		return
	}

	fileCnt := len(fileDropData.SrcFileList)
	fileUnit := "files"
	if fileCnt <= 1 {
		fileUnit = "file"
	}
	filename := fmt.Sprintf("%d %s", fileCnt, fileUnit)
	rtkPlatform.GoNotiMessageFileTransfer(filename, clientInfo.DeviceName, clientInfo.Platform, fileDropData.TimeStamp, true)
}

func ShowNotiMessageRecvFileTransferDone(fileDropData *rtkFileDrop.FilesTransferDataItem, id string) {
	clientInfo, err := rtkUtils.GetClientInfo(id)
	if err != nil {
		log.Printf("[%s] %s", rtkMisc.GetFuncInfo(), err.Error())
		return
	}

	fileCnt := len(fileDropData.SrcFileList)
	fileUnit := "files"
	if fileCnt <= 1 {
		fileUnit = "file"
	}
	filename := fmt.Sprintf("%d %s", fileCnt, fileUnit)
	rtkPlatform.GoNotiMessageFileTransfer(filename, clientInfo.DeviceName, clientInfo.Platform, fileDropData.TimeStamp, false)
}

func checkRecoverProcessEventsForPeerAsSrc(id, ipAddr string, code rtkMisc.CrossShareErr) {
	timer := time.NewTimer((interruptFailureInterval + 5) * time.Second)
	defer timer.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recoverFileTransferTimerMap.Store(id, cancel)

	select {
	case <-ctx.Done():
		log.Printf("(SRC) [%s] ID:[%s] IP:[%s] stop cache file data resend time out timer success!", rtkMisc.GetFuncInfo(), id, ipAddr)
		return
	case <-timer.C:
	}

	log.Printf("(SRC) [%s] ID:[%s] IP:[%s] cache file data resend time out!", rtkMisc.GetFuncInfo(), id, ipAddr)
	clearCacheFileDataList(id, ipAddr, code)
}

func checkRecoverProcessEventsForPeerAsDst(id, ipAddr string, code rtkMisc.CrossShareErr) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	nCount := 0
	for {
		<-ticker.C

		if rtkConnection.IsStreamExisted(id) {
			break
		}

		if nCount >= interruptFailureInterval {
			log.Printf("(DST) [%s] ID:[%s] IP:[%s] cache file data retry time out!", rtkMisc.GetFuncInfo(), id, ipAddr)
			clearCacheFileDataList(id, ipAddr, code)
			return
		}

		nCount++
	}

	cacheData := rtkFileDrop.GetFilesTransferDataItem(id,0)
	if cacheData == nil {
		return
	}

	requestFileTransRecoverMsgToSrc(id, cacheData.InterruptSrcFileName, cacheData.TimeStamp, cacheData.InterruptFileOffSet, cacheData.InterruptLastErrCode)
}

func recoverFileTransferProcessAsDst(ctx context.Context, id, ipAddr string) {
	cacheData := rtkFileDrop.GetFilesTransferDataItem(id,0)
	if cacheData == nil {
		return
	}

	if cacheData.InterruptSrcFileName == "" || cacheData.InterruptDstFileName == "" || cacheData.InterruptLastErrCode == rtkMisc.SUCCESS {
		log.Printf("[%s] ID:[%s] IP:[%s] Invalid Interrupt info!", rtkMisc.GetFuncInfo(), id, ipAddr)
		return
	}

	if cacheData.FileTransDirection != rtkFileDrop.FilesTransfer_As_Dst {
		log.Printf("[%s] ID:[%s] IP:[%s] Invalid Interrupt FileTransDirection:[%s]!", rtkMisc.GetFuncInfo(), id, ipAddr, cacheData.FileTransDirection)
		return
	}

	buildFileDropItemStream(ctx, id) // build UDP connect

	dealFilesCacheDataProcess(ctx, id, ipAddr,0)
}

func recoverFileTransferProcessAsSrc(ctx context.Context, id, ipAddr string, errCode rtkMisc.CrossShareErr) {
	if value, ok := recoverFileTransferTimerMap.Load(id); ok {
		fn := value.(context.CancelFunc)
		if fn != nil {
			fn()
			recoverFileTransferTimerMap.Store(id, nil)
		} else {
			log.Printf("[%s] ID:[%s] IP:[%s] get cache file data resend time out timer invalid !", rtkMisc.GetFuncInfo(), id, ipAddr)
		}
	}

	code := buildFileDropItemStream(ctx, id)
	if code != rtkMisc.SUCCESS {
		errCode = code
	}

	cacheData := rtkFileDrop.GetFilesTransferDataItem(id,0)
	if cacheData == nil {
		errCode = rtkMisc.ERR_BIZ_FT_DATA_EMPTY
	} else {
		if cacheData.InterruptSrcFileName == "" || cacheData.InterruptLastErrCode == rtkMisc.SUCCESS {
			log.Printf("[%s] ID:[%s] IP:[%s] Invalid Interrupt info!", rtkMisc.GetFuncInfo(), id, ipAddr)
			errCode = rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
		}

		if cacheData.FileTransDirection != rtkFileDrop.FilesTransfer_As_Src {
			log.Printf("[%s] ID:[%s] IP:[%s] Invalid Interrupt FileTransDirection:[%s]!", rtkMisc.GetFuncInfo(), id, ipAddr, cacheData.FileTransDirection)
			errCode = rtkMisc.ERR_BIZ_FT_DIRECTION_TYPE_INVALID
		}
	}

	responseFileTransRecoverMsgToDst(id, errCode)
	if errCode != rtkMisc.SUCCESS {
		log.Printf("(SRC) [%s] ID:[%s] request recover file transfer failed, errCode:[%d]", rtkMisc.GetFuncInfo(), id, errCode)
		clearCacheFileDataList(id, ipAddr, errCode)
		return
	}

	dealFilesCacheDataProcess(ctx, id, ipAddr,0)
}

func setFileDataItemProtocolListener(id string) rtkMisc.CrossShareErr {
	/*cacheList := rtkFileDrop.GetFilesTransferDataList(id)
	if cacheList == nil {
		return rtkMisc.ERR_BIZ_FT_DATA_EMPTY
	}
	for _, cacheData := range cacheList {
		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Src {
			rtkConnection.BuildFileDropItemStreamListener(cacheData.TimeStamp)
		}
	}*/

	return rtkMisc.SUCCESS
}

func buildFileDropItemStream(ctx context.Context, id string) rtkMisc.CrossShareErr {
	/*cacheList := rtkFileDrop.GetFilesTransferDataList(id)
	if cacheList == nil {
		return rtkMisc.ERR_BIZ_FT_DATA_EMPTY
	}
	for _, cacheData := range cacheList {
		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Dst {
			errCode := rtkConnection.NewFileDropItemStream(ctx, id, cacheData.TimeStamp)
			if errCode != rtkMisc.SUCCESS {
				log.Printf("[%s] ID:[%s] new File Drop Item stream err, errCode:%+v ", rtkMisc.GetFuncInfo(), id, errCode)
				return errCode
			}
		}
	}*/
	return rtkMisc.SUCCESS
}

func clearCacheFileDataList(id, ipAddr string, code rtkMisc.CrossShareErr) {
	rtkConnection.ClearFmtTypeStreamReadyFlag(id) // clear all wait for file drop stream Ready chan flag
	rtkConnection.CloseAllFileDropStream(id)      // close all file data transfer stream

	i := 0
	for rtkFileDrop.GetFilesTransferDataCacheCount(id) > 0 {
		cacheData := rtkFileDrop.GetFilesTransferDataItem(id,0)
		if cacheData == nil {
			break
		}

		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Src {
			log.Printf("(SRC) ID[%s] IP[%s] (cache) Copy file data To Socket failed, timestamp:%d, ERR code:[%d]!", id, ipAddr, cacheData.TimeStamp, code)
			rtkConnection.RemoveFileDropItemStreamListener(cacheData.TimeStamp)
		} else if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Dst {
			log.Printf("(DST) ID[%s] IP[%s] (cache) Copy file data To Socket failed, timestamp:%d, ERR code:[%d]", id, ipAddr, cacheData.TimeStamp, code)
			if i == 0 && cacheData.InterruptDstFullPath != "" {
				DeleteFile(cacheData.InterruptDstFullPath)
			}
		} else {
			log.Printf("[%s] ID:[%s] Invalid direction type:[%s]!", rtkMisc.GetFuncInfo(), id, cacheData.FileTransDirection)
		}

		i++
		rtkPlatform.GoNotifyErrEvent(id, code, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
		rtkFileDrop.SetFilesCacheItemComplete(id, cacheData.TimeStamp)
	}
}
