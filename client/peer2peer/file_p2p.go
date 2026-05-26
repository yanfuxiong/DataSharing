package peer2peer

import (
	"bufio"
	"bytes"
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
		return rtkMisc.ERR_BIZ_FD_FILE_NOT_EXISTS
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

func CancelSrcFileTransfer(id, ipAddr string, timestamp uint64, errCode rtkMisc.CrossShareErr) {
	if errCode == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI {
		log.Printf("(SRC) [%s] IP:[%s] timestamp:[%d] Copy file operation was canceled by dst GUI !", rtkMisc.GetFuncInfo(), ipAddr, timestamp)
	} else {
		log.Printf("(SRC) [%s] IP:[%s] timestamp:[%d] Copy file operation was canceled by dst errCode:%d!", rtkMisc.GetFuncInfo(), ipAddr, timestamp, errCode)
		if errCode == rtkMisc.ERR_BIZ_FT_DST_COPY_DETAILS {
			rtkFileDrop.CancelFileTransFromCacheMap(id, timestamp)
			rtkPlatform.GoNotifyErrEvent(id, errCode, ipAddr, strconv.Itoa(int(timestamp)), "", "")
			rtkConnection.CloseFileDropItemStream(id, timestamp)
			rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP)
			return
		} else if errCode == rtkMisc.ERR_BIZ_FT_DST_OPEN_STREAM {
			rtkFileDrop.CancelFileTransFromCacheMap(id, timestamp)
			rtkPlatform.GoNotifyErrEvent(id, errCode, ipAddr, strconv.Itoa(int(timestamp)), "", "")
			return
		}
	}

	if rtkFileDrop.IsCancelFileTransInProgress(id, timestamp, errCode) {
		log.Printf("(SRC) [%s] ID:[%s] Cancel FileTransfer success, timestamp:%d", rtkMisc.GetFuncInfo(), id, timestamp)
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
	if rtkFileDrop.IsCancelFileTransInProgress(id, timestamp, errCode) {
		log.Printf("(DST) [%s] ID:[%s] Cancel FileTransfer success, timestamp:%d", rtkMisc.GetFuncInfo(), id, timestamp)
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

func getFileTransferCancelErrCode(ctx context.Context, ipAddr string, timeStamp uint64, isSrc bool) rtkMisc.CrossShareErr {
	var errCode rtkMisc.CrossShareErr
	if isSrc {
		errCode = rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL
	} else {
		errCode = rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL
	}

	if source, ok := rtkUtils.GetCancelSource(ctx); ok {
		if isSrc {
			log.Printf("(SRC) IP:[%s] timeStamp:[%d] file data transfer is cancel source:%d!", ipAddr, timeStamp, source)
		} else {
			log.Printf("(DST) IP:[%s] timeStamp:[%d] file data transfer is cancel source:%d!", ipAddr, timeStamp, source)
		}

		if source == rtkCommon.TcpNetworkCancel || // need retry
			source == rtkCommon.LanServerBusinessCancel ||
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
	} else {
		log.Printf("[%s] IP:[%s] timeStamp:[%d] file data transfer is cancel, Unknown source!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
	}

	return errCode
}

func writeItemFileDataDetailsToSocket(ctx context.Context, id, ipAddr string, timeStamp uint64) {
	startTime := time.Now().UnixMilli()
	rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP) // wait for file drop stream Ready
	sFileDrop, ok := rtkConnection.GetFileDropItemStream(id, timeStamp)
	if !ok {
		log.Printf("[%s] Err: Not found file stream by ID:[%s]", rtkMisc.GetFuncInfo(), id)
		return
	}
	fileDetailsData := rtkFileDrop.GetFileDropDataDetailsData(id)
	if fileDetailsData == nil {
		rtkConnection.CloseFileDropItemStream(id, timeStamp)
		return
	}

	writer := cancelableWriter{
		realWriter: sFileDrop,
		ctx:        ctx,
	}
	nCopy, err := io.Copy(&writer, bytes.NewReader(fileDetailsData))
	if err != nil {
		log.Printf("(SRC) [%s] IP:[%s] timeStamp:[%d] Copy file details Error:%+v", rtkMisc.GetFuncInfo(), ipAddr, timeStamp, err)
		rtkConnection.CloseFileDropItemStream(id, timeStamp)
		return
	}
	log.Printf("(SRC) Copy file details to IP:[%s] success, timestamp:[%d] details size:[%d] total use [%d] ms", ipAddr, timeStamp, nCopy, time.Now().UnixMilli()-startTime)
}

func readItemFileDataDetailsFromSocket(ctx context.Context, id, ipAddr string, timeStamp uint64, detailsLen int) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	if errCode := rtkConnection.NewFileDropItemStream(ctx, id, timeStamp); errCode != rtkMisc.SUCCESS {
		log.Printf("[%s] new File Drop Item stream errCode:%+v ", rtkMisc.GetFuncInfo(), errCode)
		return rtkMisc.ERR_BIZ_FT_DST_OPEN_STREAM
	}
	sFileDrop, ok := rtkConnection.GetFileDropItemStream(id, timeStamp)
	if !ok {
		log.Printf("[%s] Err: Not found file stream by ID:[%s]", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_FD_GET_STREAM_EMPTY
	}

	reader := cancelableReader{
		realReader: sFileDrop,
		ctx:        ctx,
	}

	var detailsBuffer bytes.Buffer
	detailsBuffer.Reset()
	detailsBuffer.Grow(detailsLen)
	nCopy, err := io.CopyN(&detailsBuffer, &reader, int64(detailsLen))
	if err != nil {
		log.Printf("(DST) [%s] IP:[%s] timeStamp:[%d] Copy file details Error:%+v", rtkMisc.GetFuncInfo(), ipAddr, timeStamp, err)
		return rtkMisc.ERR_BIZ_FT_COPY_DETAILS
	}
	log.Printf("(DST) Copy file details from IP:[%s] success, timestamp:[%d] details size:[%d] total use [%d] ms", ipAddr, timeStamp, nCopy, time.Now().UnixMilli()-startTime)
	return rtkFileDrop.SetupDstFileDropDataDetails(id, ipAddr, detailsBuffer.Bytes())
}

func dealFilesCacheDataProcess(p2pCtx context.Context, id, ipAddr string, timeStamp uint64) {
	for rtkFileDrop.GetFilesTransferDataCacheCount(id) > 0 {
		cacheData := rtkFileDrop.GetFilesTransferDataItem(id, timeStamp)
		if cacheData == nil {
			break
		}

		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Src {
			resultCode = writeItemFileDataToSocket(p2pCtx, id, ipAddr, cacheData)
			if resultCode != rtkMisc.SUCCESS {
				if resultCode == rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS {
					log.Printf("(SRC) ID[%s] IP[%s] Copy file data To Socket is interrupt, timestamp:[%d], wait to resend...", id, ipAddr, cacheData.TimeStamp)
					rtkConnection.CloseAllFileDropStream(id)
					rtkMisc.GoSafe(func() { watchRecoverFileTransferCacheTimeoutAsSrc(id, ipAddr, cacheData.TimeStamp, resultCode) })
					return
				} else {
					log.Printf("(SRC) ID[%s] IP[%s] Copy file data To Socket failed, timestamp:%d, ERR code:[%d],  and not resend!", id, ipAddr, cacheData.TimeStamp, resultCode)
					sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_SRC_INTERRUPT, resultCode, cacheData.TimeStamp)
					rtkPlatform.GoNotifyErrEvent(id, resultCode, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
				}
			}
			rtkConnection.CloseFileDropItemStream(id, cacheData.TimeStamp)
			rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP) //  keep for support old version
			rtkConnection.RemoveFileDropItemStreamListener(cacheData.TimeStamp)
		} else if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Dst {
			resultCode = readItemFileDataFromSocket(p2pCtx, id, ipAddr, cacheData)
			if resultCode != rtkMisc.SUCCESS {
				if resultCode == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS {
					log.Printf("(DST) ID[%s] IP[%s] Copy file data To Socket is interrupt, timestamp:[%d], wait to retry...", id, ipAddr, cacheData.TimeStamp)
					rtkConnection.CloseAllFileDropStream(id)
					rtkMisc.GoSafe(func() { watchRecoverFileTransferCacheTimeoutAsDst(id, ipAddr, cacheData.TimeStamp, resultCode) })
					return
				} else {
					log.Printf("(DST) ID[%s] IP[%s] Copy file data To Socket failed, timestamp:%d, ERR code:[%d]  and not retry", id, ipAddr, cacheData.TimeStamp, resultCode)
					sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_DST_INTERRUPT, resultCode, cacheData.TimeStamp)
					rtkPlatform.GoNotifyErrEvent(id, resultCode, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
				}
			}
			rtkConnection.CloseFileDropItemStream(id, cacheData.TimeStamp)
			rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP) //  keep for support old version
		} else {
			log.Printf("[%s] ID:[%s] Invalid direction type:[%s], skit it!", rtkMisc.GetFuncInfo(), id, cacheData.FileTransDirection)
		}

		rtkFileDrop.SetFilesCacheItemComplete(id, cacheData.TimeStamp)
	}
}

func writeItemFileDataToSocket(p2pCtx context.Context, id, ipAddr string, fileDropReqData *rtkFileDrop.FilesTransferDataItem) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	if !rtkUtils.GetPeerClientIsRmFCL(id) {
		rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP) // wait for file drop stream Ready
	}

	if p2pCtx.Err() != nil { // deal file cache Data must return errCode and clear cache when p2p business is canceled
		return getFileDataSendCancelErrCode(p2pCtx, ipAddr, fileDropReqData.TimeStamp)
	}

	isResend := false
	if fileDropReqData.InterruptSrcFileName != "" && fileDropReqData.InterruptLastErrCode != rtkMisc.SUCCESS { //InterruptFileOffSet maybe is 0
		isResend = true
		rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP) // wait for file drop stream Ready
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
	nTotalFolderCnt := uint32(len(fileDropReqData.FolderList))
	if (nTotalFileCnt == 0 && nTotalFolderCnt == 0) || fileDropReqData.TimeStamp == 0 {
		log.Printf("[%s] get file data is invalid! fileCount:[%d] folderCount:[%d] TimeStamp:[%d] ", rtkMisc.GetFuncInfo(), nTotalFileCnt, nTotalFolderCnt, fileDropReqData.TimeStamp)
		return rtkMisc.ERR_BIZ_FD_DATA_INVALID
	}

	ctx, cancel := rtkUtils.WithCancelSource(p2pCtx)
	defer cancel(rtkCommon.FileTransDone)
	rtkFileDrop.SetCancelFileTransferFunc(id, cancel)

	if isResend {
		log.Printf("(SRC) Retry Copy file data to IP:[%s], id:[%d] file count:[%d] folder count:[%d] totalSize:[%d] TotalDescribe:[%s]...", ipAddr, fileDropReqData.TimeStamp, nTotalFileCnt, nTotalFolderCnt, fileDropReqData.TotalSize, fileDropReqData.TotalDescribe)
	} else {
		log.Printf("(SRC) Start Copy file data to IP:[%s], id:[%d] file count:[%d] folder count:[%d] totalSize:[%d] TotalDescribe:[%s]...", ipAddr, fileDropReqData.TimeStamp, nTotalFileCnt, nTotalFolderCnt, fileDropReqData.TotalSize, fileDropReqData.TotalDescribe)
	}

	progressBar := New64(int64(fileDropReqData.TotalSize))
	var curFilePath string
	var curFileSize uint64
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
					rtkPlatform.GoUpdateSendProgressBar(ipAddr, id, curFilePath, fileDoneCnt, nTotalFileCnt, curFileSize, fileDropReqData.TotalSize, uint64(barCurrentBytes), fileDropReqData.TimeStamp)
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
	getInterruptFile := false
	offSet := int64(0)
	for i, fileInfo := range fileDropReqData.SrcFileList {
		curFileSize = uint64(fileInfo.FileSize_.SizeHigh)<<32 | uint64(fileInfo.FileSize_.SizeLow)
		curFilePath = fileInfo.FilePath
		if isResend && fileInfo.FileName != fileDropReqData.InterruptSrcFileName && !getInterruptFile {
			progressBar.Add64(int64(curFileSize))
			fileDoneCnt++
			continue
		}

		fileSize := curFileSize
		if isResend && fileInfo.FileName == fileDropReqData.InterruptSrcFileName {
			getInterruptFile = true
			offSet = fileDropReqData.InterruptFileOffSet
			if fileDropReqData.InterruptFileOffSet < 0 || fileDropReqData.InterruptFileOffSet > int64(curFileSize) {
				log.Printf("[%s] Retry Copy file data to IP:[%s], id:[%d], get invalid interrupt offset:[%d]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropReqData.TimeStamp, offSet)
				return rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
			}
			progressBar.Add64(fileDropReqData.InterruptFileOffSet)
			sendBytes := progressBar.GetCurrentBytes()
			rounded := float64(sendBytes) / float64(fileDropReqData.TotalSize) * 100
			fileSize = curFileSize - uint64(fileDropReqData.InterruptFileOffSet)
			log.Printf("(SRC) Retry Copy file data to IP:[%s], id:[%d], already send:[%d], percentage:[%.2f%%], Starting from this file:[%s], offset:[%d] ...", ipAddr, fileDropReqData.TimeStamp, sendBytes, rounded, fileInfo.FileName, offSet)
		} else {
			offSet = int64(0)
		}

		errCode := writeFileToSocket(id, ipAddr, &cancelableWrite, &cancelableRead, &progressBar, fileInfo.FileName, fileInfo.FilePath, fileSize, fileDropReqData.TimeStamp, offSet, &copyBuffer)
		if errCode != rtkMisc.SUCCESS {
			return errCode
		}
		fileDoneCnt++
		if uint32(i) != (nTotalFileCnt - 1) {
			rtkPlatform.GoUpdateSendProgressBar(ipAddr, id, curFilePath, fileDoneCnt, nTotalFileCnt, curFileSize, fileDropReqData.TotalSize, uint64(progressBar.GetCurrentBytes()), fileDropReqData.TimeStamp)
		}
	}

	if isResend && !getInterruptFile {
		log.Printf("[%d] Retry Copy file data to IP:[%s], id:[%d], get invalid interrupt src file name:[%s]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropReqData.TimeStamp, fileDropReqData.InterruptSrcFileName)
		return rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
	}

	rtkPlatform.GoUpdateSendProgressBar(ipAddr, id, curFilePath, fileDoneCnt, nTotalFileCnt, curFileSize, fileDropReqData.TotalSize, fileDropReqData.TotalSize, fileDropReqData.TimeStamp)
	log.Printf("(SRC) End Copy all file data to IP:[%s] success, id:[%d] file count:[%d] folder count:[%d] TotalDescribe:[%s], total use [%d] ms", ipAddr, fileDropReqData.TimeStamp, nTotalFileCnt, nTotalFolderCnt, fileDropReqData.TotalDescribe, time.Now().UnixMilli()-startTime)
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

	if offset != 0 {
		log.Printf("(SRC) IP[%s] Retry copy file:[%s], still has size:[%d] left ...", ipAddr, fileName, fileSize)
	} else {
		log.Printf("(SRC) IP[%s] Start copy file:[%s], size:[%d] ...", ipAddr, fileName, fileSize)
	}

	nCopy := int64(0)
	var err error
	if fileSize > 0 {
		nCopy, err = io.CopyBuffer(io.MultiWriter(*totalBar, write), read, *buf)
		if err != nil {
			log.Printf("(SRC) [%s] IP:[%s] timestamp:[%d] Copy file Error:%+v!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp, err)
			if read.ctx.Err() != nil {
				return getFileDataSendCancelErrCode(read.ctx, ipAddr, timeStamp)
			}
			if rtkConnection.IsQuicEOF(err) { // close by remote, need retry  TODO: check it
				log.Printf("(SRC) [%s] IP:[%s] timestamp:[%d] quic remote Close!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS
			} else if rtkConnection.IsQuicClose(err) { // close by local, need retry  TODO: check it
				log.Printf("(SRC) [%s] IP:[%s] timestamp:[%d] quic local Closed!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("(SRC) [%s] IP:[%s] Error sending file timeout:%v", rtkMisc.GetFuncInfo(), ipAddr, netErr)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL_BUSINESS
			} else if errors.Is(err, yamux.ErrStreamClosed) || errors.Is(err, yamux.ErrStreamReset) { // old  version client use tcp stream trigger this case
				log.Printf("(SRC) IP[%s] IP:[%s]  Copy operation was canceled by close stream!", rtkMisc.GetFuncInfo(), ipAddr)
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
	log.Printf("(SRC) IP[%s] End to copy file:[%s] success, total:[%d] use [%d] ms", ipAddr, fileName, nCopy, time.Now().UnixMilli()-startTime)
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

	if !rtkMisc.FolderExists(fileDropData.DstFilePath) {
		log.Printf("[%s] Invalid fildDrop DstFilePath:%s", rtkMisc.GetFuncInfo(), fileDropData.DstFilePath)
		return rtkMisc.ERR_BIZ_FD_FOLDER_NOT_EXISTS
	}

	nTotalFileCnt := uint32(len(fileDropData.SrcFileList))
	nTotalFolderCnt := uint32(len(fileDropData.FolderList))
	if (nTotalFileCnt == 0 && nTotalFolderCnt == 0) || fileDropData.TimeStamp == 0 {
		log.Printf("[%s] get file data is invalid! fileCount:[%d] folderCount:[%d] TimeStamp:[%d] ", rtkMisc.GetFuncInfo(), nTotalFileCnt, nTotalFolderCnt, fileDropData.TimeStamp)
		return rtkMisc.ERR_BIZ_FD_DATA_INVALID
	}

	ctx, cancel := rtkUtils.WithCancelSource(p2pCtx)
	defer cancel(rtkCommon.FileTransDone)
	rtkFileDrop.SetCancelFileTransferFunc(id, cancel)

	isRetry := false                                                                                                                                //interrupt and retry transmission flag
	if fileDropData.InterruptSrcFileName != "" && fileDropData.InterruptDstFileName != "" && fileDropData.InterruptLastErrCode != rtkMisc.SUCCESS { //InterruptFileOffSet maybe is 0
		isRetry = true
		log.Printf("(DST) Retry Copy file data from IP:[%s] id:[%d] file count:[%d] folder count:[%d] totalSize:[%d] TotalDescribe:[%s]...", ipAddr, fileDropData.TimeStamp, nTotalFileCnt, nTotalFolderCnt, fileDropData.TotalSize, fileDropData.TotalDescribe)
	} else {
		log.Printf("(DST) Start Copy file data from IP:[%s] id:[%d] file count:[%d] folder count:[%d] totalSize:[%d] TotalDescribe:[%s]...", ipAddr, fileDropData.TimeStamp, nTotalFileCnt, nTotalFolderCnt, fileDropData.TotalSize, fileDropData.TotalDescribe)
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
			log.Printf("(DST) Create  %d folder success!", nFolderCount)
		}
	}

	progressBar := New64(int64(fileDropData.TotalSize))
	var curFileName string
	var curFileSize uint64
	fileDoneCnt := uint32(0)
	dstFilePath := fileDropData.DstFilePath
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
					rtkPlatform.GoUpdateReceiveProgressBar(ipAddr, id, dstFilePath, fileDoneCnt, nTotalFileCnt, curFileSize, fileDropData.TotalSize, uint64(barCurrentBytes), fileDropData.TimeStamp)
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

	copyBuffer := make([]byte, copyBufSize)
	getInterruptFile := false
	isInterruptFile := false
	offset := int64(0)
	for i, fileInfo := range fileDropData.SrcFileList {
		curFileSize = uint64(fileInfo.FileSize_.SizeHigh)<<32 | uint64(fileInfo.FileSize_.SizeLow)
		if isRetry && fileInfo.FileName != fileDropData.InterruptSrcFileName && !getInterruptFile {
			progressBar.Add64(int64(curFileSize))
			fileDoneCnt++
			continue
		}

		fileSize := curFileSize
		if isRetry && fileInfo.FileName == fileDropData.InterruptSrcFileName {
			getInterruptFile = true
			if fileDropData.InterruptFileOffSet < 0 || fileDropData.InterruptFileOffSet > int64(curFileSize) {
				log.Printf("[%d] Retry Copy file data from IP:[%s], id:[%d], get invalid interrupt offset:[%d]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropData.TimeStamp, fileDropData.InterruptFileOffSet)
				return rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
			}
			progressBar.Add64(fileDropData.InterruptFileOffSet)
			receivedBytes := progressBar.GetCurrentBytes()
			rounded := float64(receivedBytes) / float64(fileDropData.TotalSize) * 100
			log.Printf("(DST) Retry Copy file data from IP:[%s], id:[%d], already received:[%d], percentage:[%.2f%%], Starting from this file:[%s], offset:[%d]...", ipAddr, fileDropData.TimeStamp, receivedBytes, rounded, fileInfo.FileName, fileDropData.InterruptFileOffSet)

			curFileName = fileDropData.InterruptDstFileName
			dstFilePath = filepath.Join(fileDropData.DstFilePath, curFileName)
			isInterruptFile = true
			fileSize = curFileSize - uint64(fileDropData.InterruptFileOffSet)
		} else {
			isInterruptFile = false
			curFileName = rtkMisc.AdaptationPath(fileInfo.FileName)
			dstFilePath, curFileName = rtkUtils.GetTargetDstPathName(filepath.Join(fileDropData.DstFilePath, curFileName), curFileName)
		}

		cancelableRead.realReader = io.LimitReader(sFileDrop, int64(fileSize))

		errCode := readFileFromSocket(id, ipAddr, &cancelableWrite, &cancelableRead, &progressBar, fileSize, fileDropData.TimeStamp, curFileName, dstFilePath, &copyBuffer, &offset, isInterruptFile)
		if errCode != rtkMisc.SUCCESS {
			if errCode == rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS {
				rtkFileDrop.SetFilesTransferDataInterrupt(id, fileInfo.FileName, curFileName, dstFilePath, fileDropData.TimeStamp, offset, errCode)
			} else {
				DeleteFile(dstFilePath)
			}
			return errCode
		}

		fileDoneCnt++
		if uint32(i) != (nTotalFileCnt - 1) {
			rtkPlatform.GoUpdateReceiveProgressBar(ipAddr, id, dstFilePath, fileDoneCnt, nTotalFileCnt, curFileSize, fileDropData.TotalSize, uint64(progressBar.GetCurrentBytes()), fileDropData.TimeStamp)
		}
	}

	if isRetry && !getInterruptFile {
		log.Printf("[%d] Retry Copy file data from IP:[%s], id:[%d], unknown interrupt file name:[%s]!", rtkMisc.GetFuncInfo(), ipAddr, fileDropData.TimeStamp, fileDropData.InterruptSrcFileName)
		return rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
	}

	rtkPlatform.GoUpdateReceiveProgressBar(ipAddr, id, dstFilePath, fileDoneCnt, nTotalFileCnt, curFileSize, fileDropData.TotalSize, fileDropData.TotalSize, fileDropData.TimeStamp)
	log.Printf("(DST) End Copy file data from IP:[%s] success, id:[%d] file count:[%d] folder count:[%d] totalSize:[%d] totalDescribe:[%s] total use:[%d]ms", ipAddr, fileDropData.TimeStamp, nTotalFileCnt, nTotalFolderCnt, fileDropData.TotalSize, fileDropData.TotalDescribe, time.Now().UnixMilli()-startTime)
	ShowNotiMessageRecvFileTransferDone(fileDropData, id)
	return rtkMisc.SUCCESS
}

func readFileFromSocket(id, ipAddr string, write *cancelableWriter, read *cancelableReader, totalBar **ProgressBar, fileSize, timeStamp uint64, dstFileName, dstFullPath string, buf *[]byte, offset *int64, isRetry bool) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	var dstFile *os.File
	err := OpenDstFile(&dstFile, dstFullPath)
	if err != nil {
		return rtkMisc.ERR_BIZ_FD_DST_OPEN_FILE
	}
	defer CloseFile(&dstFile)
	write.realWriter = dstFile

	if isRetry {
		log.Printf("(DST) IP[%s] Retry copy file:[%s], still has size:[%d] left ...", ipAddr, dstFileName, fileSize)
	} else {
		log.Printf("(DST) IP[%s] Start copy file:[%s], size:[%d] ...", ipAddr, dstFileName, fileSize)
	}

	nDstWrite := int64(0)
	if fileSize > 0 {
		if fileSize > uint64(truncateThreshold) && !isRetry {
			dstFile.Truncate(int64(fileSize))
		}
		nDstWrite, err = io.CopyBuffer(io.MultiWriter(write, *totalBar), read, *buf)
		if err != nil {
			*offset = nDstWrite
			log.Printf("(DST) [%s] IP:[%s] timestamp:[%d] Copy file Error:%+v!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp, err)
			if read.ctx.Err() != nil {
				return getFileDataReceiveCancelErrCode(read.ctx, ipAddr, timeStamp)
			}
			if rtkConnection.IsQuicEOF(err) { // close by remote, need retry  TODO: check it
				log.Printf("(DST) [%s] IP:[%s] timestamp:[%d] quic remote Closed!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS
			} else if rtkConnection.IsQuicClose(err) { // close by local, need retry  TODO: check it
				log.Printf("(DST) [%s] IP:[%s] timestamp:[%d] quic local Closed!", rtkMisc.GetFuncInfo(), ipAddr, timeStamp)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("(DST) [%s] IP:[%s] Error Read file timeout:%v", rtkMisc.GetFuncInfo(), ipAddr, netErr)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_BUSINESS
			} else if errors.Is(err, yamux.ErrStreamClosed) || errors.Is(err, yamux.ErrStreamReset) { // old  version client trigger this case
				log.Printf("(DST) IP[%s] Copy operation was canceled by close stream!", ipAddr)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL
			} else {
				log.Printf("(DST) [%s] IP:[%s] timeStamp:[%d] Copy file Error:%+v", rtkMisc.GetFuncInfo(), ipAddr, timeStamp, err)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE
			}
		}
		dstFile.Sync()
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

func watchRecoverFileTransferCacheTimeout(id, ipAddr string, timestamp uint64, code rtkMisc.CrossShareErr, isSrc bool) {
	timer := time.NewTimer((interruptFailureInterval + 5) * time.Second)
	defer timer.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rtkFileDrop.SetFilesTransferRecoverTimerCancel(id, timestamp, cancel)
	tag := "SRC"
	if !isSrc {
		tag = "DST"
	}

	select {
	case <-ctx.Done():
		log.Printf("(%s) [%s] IP:[%s] timestamp:[%d] stop cache file data recover timer success!", tag, rtkMisc.GetFuncInfo(), ipAddr, timestamp)
		return
	case <-timer.C:
	}

	log.Printf("(%s) [%s] IP:[%s] timestamp:[%d] cache file data recover time out!  clear all cache data!", tag, rtkMisc.GetFuncInfo(), ipAddr, timestamp)
	clearFilesTransferCacheList(id, ipAddr, code)
}

func watchRecoverFileTransferCacheTimeoutAsSrc(id, ipAddr string, timestamp uint64, code rtkMisc.CrossShareErr) {
	watchRecoverFileTransferCacheTimeout(id, ipAddr, timestamp, code, true)
}

func watchRecoverFileTransferCacheTimeoutAsDst(id, ipAddr string, timestamp uint64, code rtkMisc.CrossShareErr) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	rtkMisc.GoSafe(func() { watchRecoverFileTransferCacheTimeout(id, ipAddr, timestamp, code, false) })

	nCount := 0
	for {
		<-ticker.C

		if rtkConnection.IsStreamExisted(id) { //check TCP reconnect
			break
		}

		if nCount > (interruptFailureInterval / 2) {
			log.Printf("(DST) [%s] ID:[%s] IP:[%s] cache file data retry timer time out! ", rtkMisc.GetFuncInfo(), id, ipAddr)
			return
		}
		nCount++
	}

	cacheData := rtkFileDrop.GetFilesTransferDataItem(id)
	if cacheData == nil {
		return
	}

	sendFileTransRecoverRequestToSrc(id, cacheData.InterruptSrcFileName, cacheData.TimeStamp, cacheData.InterruptFileOffSet, cacheData.InterruptLastErrCode)
}

func recoverFileTransferProcessAsDst(ctx context.Context, id, ipAddr string) {
	cacheData := rtkFileDrop.GetFilesTransferDataItem(id)
	if cacheData == nil {
		return
	}

	if cacheData.InterruptSrcFileName == "" ||
		cacheData.InterruptDstFileName == "" ||
		cacheData.InterruptLastErrCode == rtkMisc.SUCCESS ||
		cacheData.RecoverFileTransTimerCancel == nil {
		log.Printf("[%s] ID:[%s] IP:[%s] Invalid Interrupt info!", rtkMisc.GetFuncInfo(), id, ipAddr)
		return
	}

	if cacheData.FileTransDirection != rtkFileDrop.FilesTransfer_As_Dst {
		log.Printf("[%s] ID:[%s] IP:[%s] Invalid Interrupt FileTransDirection:[%s]!", rtkMisc.GetFuncInfo(), id, ipAddr, cacheData.FileTransDirection)
		return
	}

	code := buildFileDropItemStream(ctx, id) // build quic connect
	if code != rtkMisc.SUCCESS {
		return
	}

	cacheData.RecoverFileTransTimerCancel()
	dealFilesCacheDataProcess(ctx, id, ipAddr)
}

func recoverFileTransferProcessAsSrc(ctx context.Context, id, ipAddr string) {
	errCode := buildFileDropItemStream(ctx, id)

	cacheData := rtkFileDrop.GetFilesTransferDataItem(id)
	if cacheData == nil {
		errCode = rtkMisc.ERR_BIZ_FD_DATA_EMPTY
	} else {
		if cacheData.InterruptSrcFileName == "" || cacheData.InterruptLastErrCode == rtkMisc.SUCCESS || cacheData.RecoverFileTransTimerCancel == nil {
			log.Printf("[%s] ID:[%s] IP:[%s] Invalid Interrupt info!", rtkMisc.GetFuncInfo(), id, ipAddr)
			errCode = rtkMisc.ERR_BIZ_FT_INTERRUPT_INFO_INVALID
		}

		if cacheData.FileTransDirection != rtkFileDrop.FilesTransfer_As_Src {
			log.Printf("[%s] ID:[%s] IP:[%s] TimeStamp:[%d] Invalid Interrupt FileTransDirection:[%s]!", rtkMisc.GetFuncInfo(), id, ipAddr, cacheData.TimeStamp, cacheData.FileTransDirection)
			errCode = rtkMisc.ERR_BIZ_FD_DIRECTION_TYPE_INVALID
		}
	}

	if sendFileTransRecoverResponseToDst(id, errCode) != rtkMisc.SUCCESS || errCode != rtkMisc.SUCCESS {
		return
	}
	cacheData.RecoverFileTransTimerCancel()
	dealFilesCacheDataProcess(ctx, id, ipAddr)
}

func buildFileDropItemStream(ctx context.Context, id string) rtkMisc.CrossShareErr {
	cacheList := rtkFileDrop.GetFilesTransferDataList(id)
	if cacheList == nil {
		return rtkMisc.ERR_BIZ_FD_DATA_EMPTY
	}
	for _, cacheData := range cacheList {
		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Dst {
			errCode := rtkConnection.NewFileDropItemStream(ctx, id, cacheData.TimeStamp)
			if errCode != rtkMisc.SUCCESS {
				log.Printf("[%s] ID:[%s] TimeStamp:[%d]  new File Drop Item stream err, errCode:%+v ", rtkMisc.GetFuncInfo(), id, cacheData.TimeStamp, errCode)
				return errCode
			}
		}
	}
	return rtkMisc.SUCCESS
}

func clearFilesTransferCacheList(id, ipAddr string, code rtkMisc.CrossShareErr) {
	rtkConnection.CloseAllFileDropStream(id) // close all file data transfer stream

	i := 0
	for rtkFileDrop.GetFilesTransferDataCacheCount(id) > 0 {
		cacheData := rtkFileDrop.GetFilesTransferDataItem(id)
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
