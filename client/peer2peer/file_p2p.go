package peer2peer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p/core/network"
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

	"github.com/libp2p/go-yamux/v5"
)

const (
	copyBufNormalSize = 4 << 20  // 4MB
	copyBufShortSize  = 1 << 20  // 1MB
	truncateThreshold = 32 << 20 // 32MB

	interruptFailureInterval = 10 //seconds, Interrupt file data transfer time out: 10s
)

var (
	fileTransferInfoMap sync.Map //key: ID
)

type fileTransferInfo struct {
	cancelFn       func(rtkCommon.CancelBusinessSource)
	isCancelByPeer bool // true: cancel by peer and not need send interrupt message to peer
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
	}
	log.Printf("Remove file:[%s] success!", filePath)
	return err
}

func SetFileTransferInfo(id string, fn func(rtkCommon.CancelBusinessSource)) {
	fileTransferInfoMap.Store(id, fileTransferInfo{
		cancelFn:       fn,
		isCancelByPeer: false,
	})
}

func CancelSrcFileTransfer(id string, timestamp uint64) {
	if rtkFileDrop.IsFileTransInProgress(id, timestamp) {
		if value, ok := fileTransferInfoMap.Load(id); ok {
			fileInfo := value.(fileTransferInfo)
			fileInfo.isCancelByPeer = true
			fileTransferInfoMap.Store(id, fileInfo)
			log.Printf("[%s] (SRC) ID:[%s] Cancel FileTransfer success, timestamp:%d", rtkMisc.GetFuncInfo(), id, timestamp)
			fileInfo.cancelFn(rtkCommon.FileTransDstGuiCancel)
		}
	} else {
		rtkFileDrop.CancelFileTransFromCacheMap(id, timestamp)
		rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP)
		rtkConnection.CloseFileDropItemStream(id, timestamp)
	}
}

func CancelDstFileTransfer(id string) {
	if value, ok := fileTransferInfoMap.Load(id); ok {
		fileInfo := value.(fileTransferInfo)
		fileInfo.isCancelByPeer = true
		fileTransferInfoMap.Store(id, fileInfo)
		log.Printf("[%s] (DST) ID:[%s] Cancel FileTransfer success!", rtkMisc.GetFuncInfo(), id)
		fileInfo.cancelFn(rtkCommon.FileTransSrcCancel)
	}
}

func IsInterruptByPeer(id string) bool {
	if value, ok := fileTransferInfoMap.Load(id); ok {
		fileInfo := value.(fileTransferInfo)
		return fileInfo.isCancelByPeer
	}
	return false
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

			/*if source == rtkCommon.UpperLevelBusinessCancel ||
				source == rtkCommon.OldP2PBusinessCancel {
				rtkConnection.CancelFileTransNode()
			}*/

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

func dealFilesCacheDataProcess(ctx context.Context, id, ipAddr string) {
	for rtkFileDrop.GetFilesTransferDataCacheCount(id) > 0 {
		cacheData := rtkFileDrop.GetFilesTransferDataItem(id)
		if cacheData == nil {
			break
		}

		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Src {
			resultCode, reTry := writeFileDataItemToSocket(ctx, id, ipAddr, cacheData)
			if resultCode != rtkMisc.SUCCESS {
				if !reTry {
					log.Printf("(SRC) ID[%s] IP[%s] Copy file data To Socket failed, timestamp:%d, ERR code:[%d],  and not resend!", id, ipAddr, cacheData.TimeStamp, resultCode)
					// need notify to dst
					sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_SRC_INTERRUPT, resultCode, cacheData.TimeStamp)
					rtkPlatform.GoNotifyErrEvent(id, resultCode, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
				} else {
					log.Printf("(SRC) ID[%s] IP[%s] Copy file data To Socket is interrupt, timestamp:[%d], wait to resend...", id, ipAddr, cacheData.TimeStamp)
					rtkMisc.GoSafe(func() { checkRecoverProcessEventsForPeerAsSrc(id, ipAddr, resultCode) })
					return
				}
			}
			rtkConnection.RemoveFileDropItemStreamListener(cacheData.TimeStamp)
		} else if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Dst {
			resultCode, reTry := readFileDataItemFromSocket(ctx, id, ipAddr, cacheData)
			if resultCode != rtkMisc.SUCCESS {
				if !reTry { // any exceptions and user cancellation need to Notify to platfrom
					log.Printf("(DST) ID[%s] IP[%s] Copy file data To Socket failed, timestamp:%d, ERR code:[%d]   and not retry", id, ipAddr, cacheData.TimeStamp, resultCode)

					if resultCode != rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS && resultCode != rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI { // need notify to src
						sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_DST_INTERRUPT, resultCode, cacheData.TimeStamp)
					}

					rtkPlatform.GoNotifyErrEvent(id, resultCode, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
				} else {
					log.Printf("(DST) ID[%s] IP[%s] Copy file data To Socket is interrupt, timestamp:[%d], wait to retry...", id, ipAddr, cacheData.TimeStamp)
					rtkMisc.GoSafe(func() { checkRecoverProcessEventsForPeerAsDst(id, ipAddr, resultCode) })
					return
				}
			}
		} else {
			log.Printf("[%s] ID:[%s] Unknown direction type:[%s], skit it!", rtkMisc.GetFuncInfo(), id, cacheData.FileTransDirection)
		}

		rtkFileDrop.SetFilesCacheItemComplete(id, cacheData.TimeStamp)
	}
}

func writeFileDataItemToSocket(ctx context.Context, id, ipAddr string, fileDropReqData *rtkFileDrop.FilesTransferDataItem) (errCode rtkMisc.CrossShareErr, needReTry bool) {
	startTime := time.Now().UnixMilli()
	errCode = rtkMisc.SUCCESS
	needReTry = true

	rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP) // wait for file drop stream Ready
	var sFileDrop network.Stream
	var ok bool
	if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
		sFileDrop, ok = rtkConnection.GetFileDropItemStream(id, fileDropReqData.TimeStamp)
		defer rtkConnection.CloseFileDropItemStream(id, fileDropReqData.TimeStamp)
	} else {
		sFileDrop, ok = rtkConnection.GetFmtTypeStream(id, rtkCommon.FILE_DROP)
		defer rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
	}
	if !ok {
		log.Printf("[%s] Err: Not found file stream by ID:[%s]", rtkMisc.GetFuncInfo(), id)
		errCode = rtkMisc.ERR_BIZ_FT_GET_STREAM_EMPTY
		return
	}

	fileCount := len(fileDropReqData.SrcFileList)
	folderCount := len(fileDropReqData.FolderList)
	if (fileCount == 0 && folderCount == 0) || fileDropReqData.TimeStamp == 0 {
		log.Printf("[%s] get file data is invalid! fileCount:[%d] folderCount:[%d] TimeStamp:[%d] ", rtkMisc.GetFuncInfo(), fileCount, folderCount, fileDropReqData.TimeStamp)
		needReTry = false
		errCode = rtkMisc.ERR_BIZ_FT_DATA_INVALID
		return
	}

	ctx, cancel := rtkUtils.WithCancelSource(ctx)
	defer cancel(rtkCommon.FileTransDone)
	SetFileTransferInfo(id, cancel) //Used when the receiving end is exception or cancellation
	progressBar := New64(int64(fileDropReqData.TotalSize))

	rtkMisc.GoSafe(func() {
		timeOutTk := time.NewTicker(10 * time.Second)
		defer timeOutTk.Stop()
		barLastBytes := int64(0)
		barMax := progressBar.GetMax()
		for {
			select {
			case <-ctx.Done():
				//cancel by (DST) maybe block at stream Write, so need interrupt by close stream
				if IsInterruptByPeer(id) {
					time.Sleep(1 * time.Millisecond)
					if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
						rtkConnection.CloseFileDropItemStream(id, fileDropReqData.TimeStamp)
					} else {
						rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
					}
				}
				return
			// TODO: Update sender progress bar
			case <-timeOutTk.C:
				barCurrentBytes := progressBar.GetCurrentBytes()
				if barLastBytes == barCurrentBytes { // the file copy is timeout, 10s
					log.Printf("[%s] (SRC) IP[%s] Copy file data time out!", rtkMisc.GetFuncInfo(), ipAddr)

					//TODO:Need to handle timeout
					//io.Copy maybe block Exceeding 10 seconds, and the reason may be congestion control, packet loss retransmission, or disk I/O lag
					//rtkConnection.CloseFileDropItemStream(id, fileDropReqData.TimeStamp)
					//return
				}
				//log.Printf("[%s] barCurrentBytes: %d", rtkMisc.GetFuncInfo(), barCurrentBytes)
				barLastBytes = barCurrentBytes

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

	copyBuffer := make([]byte, copyBufNormalSize)
	isResend := false
	if fileDropReqData.InterruptSrcFileName != "" && fileDropReqData.InterruptFileTimeStamp > 0 {
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
				log.Printf("[%d] Retry Copy file data to IP:[%s], id:[%d], get invalid interrupt offset:[%d]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropReqData.TimeStamp, fileDropReqData.InterruptFileOffSet)
				needReTry = false
				errCode = rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
				return
			}
			log.Printf("(SRC) Retry Copy file data to IP:[%s], id:[%d], Starting from this file:[%s], offset:[%d] ...", ipAddr, fileDropReqData.TimeStamp, fileInfo.FileName, fileDropReqData.InterruptFileOffSet)
		} else {
			offSet = int64(0)
		}

		if fileSize < uint64(copyBufNormalSize) {
			copyBuffer = copyBuffer[:copyBufShortSize]
		} else {
			copyBuffer = copyBuffer[:copyBufNormalSize]
		}

		srcTransResult := writeFileToSocket(id, ipAddr, &cancelableWrite, &cancelableRead, &progressBar, fileInfo.FileName, fileInfo.FilePath, fileSize, fileDropReqData.TimeStamp, offSet, &copyBuffer)
		if srcTransResult != rtkMisc.SUCCESS {
			if srcTransResult == rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS ||
				srcTransResult == rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_GUI ||
				srcTransResult == rtkMisc.ERR_BIZ_FT_FILE_NOT_EXISTS {
				needReTry = false
			}

			errCode = srcTransResult
			return
		}
	}

	if isResend && !getInterruptFile {
		log.Printf("[%d] Retry Copy file data to IP:[%s], id:[%d], get invalid interrupt src file name:[%s]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropReqData.TimeStamp, fileDropReqData.InterruptSrcFileName)
		needReTry = false
		errCode = rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
		return
	}

	log.Printf("(SRC) End Copy all file data to IP:[%s] success, id:[%d] file count:[%d] TotalDescribe:[%s], total use [%d] ms", ipAddr, fileDropReqData.TimeStamp, fileCount, fileDropReqData.TotalDescribe, time.Now().UnixMilli()-startTime)
	ShowNotiMessageSendFileTransferDone(fileDropReqData, id)
	return
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
		nCopy, err = io.CopyBuffer(io.MultiWriter(write, *totalBar), read, *buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("Error sending file timeout:", netErr)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_TIMEOUT
			} else if errors.Is(err, context.Canceled) {
				return getFileDataSendCancelErrCode(read.ctx, ipAddr, timeStamp)
			} else if errors.Is(err, yamux.ErrStreamClosed) {
				/*if IsInterruptByPeer(id) {
					log.Printf("(SRC) IP[%s] Copy operation was canceled!", ip)
					return  errCode = rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL
				}
				log.Printf("(SRC) IP[%s] Copy operation was timeout!", ip)*/
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_TIMEOUT
			} else {
				log.Printf("(SRC) IP:[%s] timeStamp:[%d] Copy file Error:%+v", ipAddr, timeStamp, err)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE
			}
		}
		bufio.NewWriter(write).Flush()
	}
	log.Printf("(SRC) End to copy file:[%s], size:[%d] use [%d] ms", fileName, nCopy, time.Now().UnixMilli()-startTime)
	return rtkMisc.SUCCESS
}

func readFileDataItemFromSocket(ctx context.Context, id, ipAddr string, fileDropData *rtkFileDrop.FilesTransferDataItem) (errCode rtkMisc.CrossShareErr, needReTry bool) {
	startTime := time.Now().UnixMilli()
	needReTry = true
	errCode = rtkMisc.SUCCESS

	var sFileDrop network.Stream
	var ok bool
	if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
		sFileDrop, ok = rtkConnection.GetFileDropItemStream(id, fileDropData.TimeStamp)
		defer rtkConnection.CloseFileDropItemStream(id, fileDropData.TimeStamp)
	} else {
		sFileDrop, ok = rtkConnection.GetFmtTypeStream(id, rtkCommon.FILE_DROP)
		defer rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
	}
	if !ok {
		log.Printf("[%s] Err: Not found FileDrop stream by ID: %s", rtkMisc.GetFuncInfo(), id)
		errCode = rtkMisc.ERR_BIZ_FT_GET_STREAM_EMPTY
		return
	}

	if fileDropData.Cmd != rtkCommon.FILE_DROP_ACCEPT {
		log.Printf("[%s] Invalid fildDrop cmd state:%s", rtkMisc.GetFuncInfo(), fileDropData.Cmd)
		needReTry = false
		errCode = rtkMisc.ERR_BIZ_FT_UNKNOWN_CMD
		return
	}

	nSrcFileCount := uint32(len(fileDropData.SrcFileList))
	folderCount := uint32(len(fileDropData.FolderList))
	if (nSrcFileCount == 0 && folderCount == 0) || fileDropData.TimeStamp == 0 {
		log.Printf("[%s] get file data is invalid! fileCount:[%d] folderCount:[%d] TimeStamp:[%d] ", rtkMisc.GetFuncInfo(), nSrcFileCount, folderCount, fileDropData.TimeStamp)
		needReTry = false
		errCode = rtkMisc.ERR_BIZ_FT_DATA_INVALID
		return
	}

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

	progressBar := New64(int64(fileDropData.TotalSize))
	ctx, cancel := rtkUtils.WithCancelSource(ctx)
	defer cancel(rtkCommon.FileTransDone)

	currentFileSize := uint64(0)
	sentCount := uint32(0)
	dstFullFilePath := fileDropData.DstFilePath
	rtkMisc.GoSafe(func() {
		barTicker := time.NewTicker(100 * time.Millisecond)
		defer barTicker.Stop()

		barLastBytes := int64(0)
		timeoutBarCnt := int(0)
		barMax := progressBar.GetMax()
		for {
			select {
			case <-ctx.Done():
				//cancel by (SRC) maybe block at stream Read, so need interrupt by close stream
				/*if IsInterruptByPeer(id) {
					time.Sleep(1 * time.Millisecond)
					if rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
						rtkConnection.CloseFileDropItemStream(id, fileDropData.TimeStamp)
					} else {
						rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
					}
				}*/
				return
			case <-barTicker.C:
				barCurrentBytes := progressBar.GetCurrentBytes()
				if barLastBytes != barCurrentBytes {
					rtkPlatform.GoUpdateMultipleProgressBar(ipAddr, id, dstFullFilePath, sentCount, nSrcFileCount, currentFileSize, fileDropData.TotalSize, uint64(barCurrentBytes), fileDropData.TimeStamp)
					barLastBytes = barCurrentBytes
					timeoutBarCnt = 0
				} else {
					timeoutBarCnt++
					if timeoutBarCnt >= 300 { // dst copy file data timeout: 30s
						log.Printf("[%s] (DST) IP[%s] id:[%d] Copy file data time out! wait to retry...", rtkMisc.GetFuncInfo(), ipAddr, fileDropData.TimeStamp)
						//rtkConnection.ResetFileDropItemStream(id, fileDropData.TimeStamp) //set all file data transfer stream Reset, wait to retry...
						return
					}
				}

				if barCurrentBytes >= barMax {
					return
				}
			}
		}
	})

	//SetFileTransferInfo(id, cancel)                   //Used when there is an exception on the sending end
	rtkFileDrop.SetCancelFileTransferFunc(id, cancel) //Used when the recipient cancels
	cancelableRead := cancelableReader{
		realReader: nil,
		ctx:        ctx,
	}
	cancelableWrite := cancelableWriter{
		realWriter: nil,
		ctx:        ctx,
	}

	isRetry := false
	if fileDropData.InterruptSrcFileName != "" && fileDropData.InterruptDstFileName != "" && fileDropData.InterruptFileTimeStamp > 0 {
		isRetry = true
		log.Printf("(DST) Retry Copy file data from IP:[%s], id:[%d], count:[%d], totalSize:[%d], TotalDescribe:[%s]...", ipAddr, fileDropData.TimeStamp, nSrcFileCount, fileDropData.TotalSize, fileDropData.TotalDescribe)
	} else {
		log.Printf("(DST) Start Copy file data from IP:[%s], id:[%d], count:[%d], totalSize:[%d], TotalDescribe:[%s]...", ipAddr, fileDropData.TimeStamp, nSrcFileCount, fileDropData.TotalSize, fileDropData.TotalDescribe)
	}

	offset := int64(0)
	getInterruptFile := false
	isRecoverFile := false
	copyBuffer := make([]byte, copyBufNormalSize)
	for i, fileInfo := range fileDropData.SrcFileList {
		currentFileSize = uint64(fileInfo.FileSize_.SizeHigh)<<32 | uint64(fileInfo.FileSize_.SizeLow)
		if isRetry && fileInfo.FileName != fileDropData.InterruptSrcFileName && !getInterruptFile {
			progressBar.Add64(int64(currentFileSize))
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

			log.Printf("(DST) Retry Copy file data from IP:[%s], id:[%d], received:[%d], percentage:[%.2f%%], Starting from this file:[%s], offset:[%d]...", ipAddr, fileDropData.TimeStamp, receivedBytes, rounded, fileInfo.FileName, fileDropData.InterruptFileOffSet)
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

		if currentFileSize < uint64(copyBufNormalSize) {
			copyBuffer = copyBuffer[:copyBufShortSize]
		} else {
			copyBuffer = copyBuffer[:copyBufNormalSize]
		}

		dstTransResult := readFileFromSocket(id, ipAddr, &cancelableWrite, &cancelableRead, &progressBar, currentFileSize, fileDropData.TimeStamp, dstFileName, dstFullFilePath, &copyBuffer, &offset, isRecoverFile)
		if dstTransResult != rtkMisc.SUCCESS {
			if dstTransResult == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS || dstTransResult == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI {
				needReTry = false // not need retry
			} else {
				interruptTimeStampSec := time.Now().Unix()
				rtkFileDrop.SetFilesTransferDataInterrupt(id, fileInfo.FileName, dstFileName, fileDropData.TimeStamp, offset, interruptTimeStampSec, dstTransResult)
				//rtkConnection.CloseAllFileDropStream(id) // set all file data transfer stream Reset, wait to retry...
			}
			errCode = dstTransResult
			return
		}
		sentCount++
		if uint32(i) != (nSrcFileCount - 1) {
			rtkPlatform.GoUpdateMultipleProgressBar(ipAddr, id, dstFullFilePath, sentCount, nSrcFileCount, currentFileSize, fileDropData.TotalSize, uint64(progressBar.GetCurrentBytes()), fileDropData.TimeStamp)
		}
	}

	if isRetry && !getInterruptFile {
		log.Printf("[%d] Retry Copy file data from IP:[%s], id:[%d], get invalid interrupt file name:[%s]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropData.TimeStamp, fileDropData.InterruptSrcFileName)
		needReTry = false
		errCode = rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
		return
	}

	rtkPlatform.GoUpdateMultipleProgressBar(ipAddr, id, dstFullFilePath, sentCount, nSrcFileCount, currentFileSize, fileDropData.TotalSize, fileDropData.TotalSize, fileDropData.TimeStamp)
	log.Printf("(DST) End Copy file data from IP:[%s] success, id:[%d] count:[%d] totalSize:[%d] totalDescribe:[%s] total use:[%d]ms", ipAddr, fileDropData.TimeStamp, nSrcFileCount, fileDropData.TotalSize, fileDropData.TotalDescribe, time.Now().UnixMilli()-startTime)
	ShowNotiMessageRecvFileTransferDone(fileDropData, id)
	return
}

func readFileFromSocket(id, ipAddr string, write *cancelableWriter, read *cancelableReader, totalBar **ProgressBar, fileSize, timeStamp uint64, dstfileName, dstFullPath string, buf *[]byte, offset *int64, isRecover bool) rtkMisc.CrossShareErr {
	var dstFile *os.File
	startTime := time.Now().UnixMilli()

	err := OpenDstFile(&dstFile, dstFullPath)
	if err != nil {
		return rtkMisc.ERR_BIZ_FD_DST_OPEN_FILE
	}
	defer CloseFile(&dstFile)

	log.Printf("(DST) IP[%s] Start to copy file:[%s], size:[%d] ...", ipAddr, dstfileName, fileSize)
	nDstWrite := int64(0)
	if fileSize > 0 {
		if fileSize > uint64(truncateThreshold) && !isRecover {
			dstFile.Truncate(int64(fileSize))
		}
		write.realWriter = dstFile
		nDstWrite, err = io.CopyBuffer(io.MultiWriter(write, *totalBar), read, *buf)
		if err != nil {
			*offset = nDstWrite
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("(DST) IP:[%s] Error Read file timeout:%s", ipAddr, netErr)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT
			} else if errors.Is(err, context.Canceled) {
				return getFileDataReceiveCancelErrCode(read.ctx, ipAddr, timeStamp)
			} else {
				log.Printf("(DST) IP:[%s] timeStamp:[%d] Copy file Error:%+v", ipAddr, timeStamp, err)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE
			}
		}
	}

	if uint64(nDstWrite) >= fileSize {
		log.Printf("(DST) IP[%s] End to Copy file:[%s] success, total:[%d] use [%d] ms", ipAddr, dstfileName, nDstWrite, time.Now().UnixMilli()-startTime)
		return rtkMisc.SUCCESS
	} else {
		log.Printf("(DST) IP[%s] End to Copy file:[%s] failed, total:[%d], it less then filesize:[%d]...", ipAddr, dstfileName, nDstWrite, fileSize)
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
	timer := time.NewTimer(20 * time.Second)
	defer timer.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recoverFileTransferTimerMap.Store(id, cancel)

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
	}

	log.Printf("[%s] ID:[%s] IP:[%s] cache file data resend time out!", rtkMisc.GetFuncInfo(), id, ipAddr)
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

		if nCount >= 20 {
			log.Printf("[%s] ID:[%s] IP:[%s] cache file data retry time out!", rtkMisc.GetFuncInfo(), id, ipAddr)
			clearCacheFileDataList(id, ipAddr, code)
			return
		}

		nCount++
	}

	buildFileDropItemStreamListener(id)

	cacheData := rtkFileDrop.GetFilesTransferDataItem(id)
	if cacheData == nil {
		return
	}

	requestFileTransRecoverMsgToSrc(id, cacheData.InterruptDstFileName, cacheData.TimeStamp, cacheData.InterruptFileOffSet, cacheData.InterruptFileTimeStamp, cacheData.InterruptLastErrCode)
}

func recoverFileTransferProcessAsDst(ctx context.Context, id, ipAddr string) {
	cacheData := rtkFileDrop.GetFilesTransferDataItem(id)
	if cacheData == nil {
		return
	}

	if cacheData.InterruptDstFileName == "" || cacheData.InterruptFileTimeStamp == 0 {
		log.Printf("[%s] ID:[%s] IP:[%s] Unknown Interrupt info!", rtkMisc.GetFuncInfo(), id, ipAddr)
		return
	}

	if cacheData.FileTransDirection != rtkFileDrop.FilesTransfer_As_Dst {
		log.Printf("[%s] ID:[%s] IP:[%s] Unknown Interrupt FileTransDirection:[%s]!", rtkMisc.GetFuncInfo(), id, ipAddr, cacheData.FileTransDirection)
		return
	}

	buildFileDropItemStream(ctx, id) // build UDP connect

	//TODO: check is file trans in process
	dealFilesCacheDataProcess(ctx, id, ipAddr)
}

func recoverFileTransferProcessAsSrc(ctx context.Context, id, ipAddr string, errCode rtkMisc.CrossShareErr) {
	if value, ok := recoverFileTransferTimerMap.Load(id); ok {
		fn := value.(context.CancelFunc)
		if fn != nil {
			fn()
			recoverFileTransferTimerMap.Store(id, nil)
			log.Printf("[%s] ID:[%s] IP:[%s] stop cache file data resend time out timer success!", rtkMisc.GetFuncInfo(), id, ipAddr)
		} else {
			log.Printf("[%s] ID:[%s] IP:[%s] get cache file data resend time out timer invalid !", rtkMisc.GetFuncInfo(), id, ipAddr)
		}
	}

	code := buildFileDropItemStreamListener(id)
	if code != rtkMisc.SUCCESS {
		errCode = code
	}

	code = buildFileDropItemStream(ctx, id)
	if code != rtkMisc.SUCCESS {
		errCode = code
	}

	cacheData := rtkFileDrop.GetFilesTransferDataItem(id)
	if cacheData == nil {
		errCode = rtkMisc.ERR_BIZ_FT_DATA_EMPTY
	}

	if cacheData.InterruptSrcFileName == "" || cacheData.InterruptFileTimeStamp == 0 {
		log.Printf("[%s] ID:[%s] IP:[%s] Unknown Interrupt info!", rtkMisc.GetFuncInfo(), id, ipAddr)
		errCode = rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
	}

	if cacheData.FileTransDirection != rtkFileDrop.FilesTransfer_As_Src {
		log.Printf("[%s] ID:[%s] IP:[%s] Unknown Interrupt FileTransDirection:[%s]!", rtkMisc.GetFuncInfo(), id, ipAddr, cacheData.FileTransDirection)
		errCode = rtkMisc.ERR_BIZ_FT_DIRECTION_TYPE_INVALID
	}

	responseFileTransRecoverMsgToDst(id, errCode)
	if errCode != rtkMisc.SUCCESS {
		log.Printf("[%s](SRC) request recover file transfer failed, errCode:[%d]", rtkMisc.GetFuncInfo(), errCode)
		clearCacheFileDataList(id, ipAddr, errCode)
		return
	}

	//TODO: check is file trans in process
	dealFilesCacheDataProcess(ctx, id, ipAddr)
}

func buildFileDropItemStreamListener(id string) rtkMisc.CrossShareErr {
	cacheList := rtkFileDrop.GetFilesTransferDataList(id)
	if cacheList == nil {
		return rtkMisc.ERR_BIZ_FT_DATA_EMPTY
	}
	for _, cacheData := range cacheList {
		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Src {
			rtkConnection.BuildFileDropItemStreamListener(cacheData.TimeStamp)
		}
	}

	return rtkMisc.SUCCESS
}

func buildFileDropItemStream(ctx context.Context, id string) rtkMisc.CrossShareErr {
	cacheList := rtkFileDrop.GetFilesTransferDataList(id)
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
	}
	return rtkMisc.SUCCESS
}

func clearCacheFileDataList(id, ipAddr string, code rtkMisc.CrossShareErr) rtkMisc.CrossShareErr {
	cacheList := rtkFileDrop.GetFilesTransferDataList(id)
	if cacheList == nil {
		return rtkMisc.ERR_BIZ_FT_DATA_EMPTY
	}

	for i, cacheData := range cacheList {
		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Src {
			log.Printf("(SRC) ID[%s] IP[%s] (cache) Copy file data To Socket failed, timestamp:%d, ERR code:[%d]!", id, ipAddr, cacheData.TimeStamp, code)
			if i > 0 {
				rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP) // clear wait for file drop stream Ready chan flag Except for the first one
			}
			rtkConnection.RemoveFileDropItemStreamListener(cacheData.TimeStamp)
		} else if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Dst {
			log.Printf("(DST) ID[%s] IP[%s] (cache) Copy file data To Socket failed, timestamp:%d, ERR code:[%d]", id, ipAddr, cacheData.TimeStamp, code)
			if i == 0 && cacheData.InterruptDstFileName != "" {
				DeleteFile(cacheData.InterruptDstFileName)
			}
		} else {
			log.Printf("[%s] ID:[%s] Unknown direction type:[%s]!", rtkMisc.GetFuncInfo(), id, cacheData.FileTransDirection)
		}

		rtkPlatform.GoNotifyErrEvent(id, code, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
		rtkFileDrop.SetFilesCacheItemComplete(id, cacheData.TimeStamp)
	}

	return rtkMisc.SUCCESS
}
