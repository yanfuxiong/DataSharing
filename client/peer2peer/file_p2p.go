package peer2peer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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
	copyBufSize      = 4 << 20 // 4MB
	copyBufShortSize = 1 << 20 // 1MB
)

var (
	fileTransferInfoMap sync.Map //key: ID
)

type fileTransferInfo struct {
	cancelFn       func()
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

func OpenSrcFile(file **os.File, filePath string) rtkMisc.CrossShareErr {
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
	_, errSeek := (*file).Seek(0, io.SeekStart)
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
	*file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
	return err
}

func SetFileTransferInfo(id string, fn func()) {
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
			fileInfo.cancelFn()
		}
	} else {
		rtkFileDrop.CancelFileTransFromCacheMap(id, timestamp)
	}
}

func CancelDstFileTransfer(id string) {
	if value, ok := fileTransferInfoMap.Load(id); ok {
		fileInfo := value.(fileTransferInfo)
		fileInfo.isCancelByPeer = true
		fileTransferInfoMap.Store(id, fileInfo)
		log.Printf("[%s] (DST) ID:[%s] Cancel FileTransfer success!", rtkMisc.GetFuncInfo(), id)
		fileInfo.cancelFn()
	}
}

func IsInterruptByPeer(id string) bool {
	if value, ok := fileTransferInfoMap.Load(id); ok {
		fileInfo := value.(fileTransferInfo)
		return fileInfo.isCancelByPeer
	}
	return false
}

func dealFilesCacheDataProcess(id, ipAddr string) {
	resultCode := rtkMisc.SUCCESS

	for rtkFileDrop.GetFilesTransferDataCacheCount(id) > 0 {
		cacheData := rtkFileDrop.GetFilesTransferDataItem(id)
		if cacheData == nil {
			break
		}

		if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Src {
			resultCode = writeFileDataToSocket(id, ipAddr, cacheData)
			if resultCode != rtkMisc.SUCCESS {
				log.Printf("(SRC) ID[%s] IP[%s] Copy file data To Socket failed, timestamp:%d, ERR code:[%d]", id, ipAddr, cacheData.TimeStamp, resultCode)
				if resultCode != rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL && resultCode != rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_TIMEOUT {
					sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_SRC_INTERRUPT, resultCode, cacheData.TimeStamp)
					rtkPlatform.GoNotifyErrEvent(id, resultCode, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
				}
			}
		} else if cacheData.FileTransDirection == rtkFileDrop.FilesTransfer_As_Dst {
			resultCode = handleFileDataFromSocket(id, ipAddr, cacheData)
			if resultCode != rtkMisc.SUCCESS {
				log.Printf("(DST) ID[%s] IP[%s] Copy file data To Socket failed, timestamp:%d, ERR code:[%d]", id, ipAddr, cacheData.TimeStamp, resultCode)

				// any exceptions and user cancellation need to Notify to platfrom
				if resultCode != rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL && resultCode != rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT {
					sendFileTransInterruptMsgToPeer(id, COMM_FILE_TRANSFER_DST_INTERRUPT, resultCode, cacheData.TimeStamp)
					rtkPlatform.GoNotifyErrEvent(id, resultCode, ipAddr, strconv.Itoa(int(cacheData.TimeStamp)), "", "")
				}
			}
		} else {
			log.Printf("[%s] ID:[%s] Unknown direction type:[%s], skit it!", rtkMisc.GetFuncInfo(), id, cacheData.FileTransDirection)
		}

		rtkFileDrop.SetFilesCacheItemComplete(id, cacheData.TimeStamp)
	}
}

func writeFileDataToSocket(id, ipAddr string, fileDropReqData *rtkFileDrop.FilesTransferDataItem) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()

	rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP) // wait for file drop stream Ready
	sFileDrop, ok := rtkConnection.GetFileDropItemStream(id, fileDropReqData.TimeStamp)
	if !ok {
		log.Printf("[%s] Err: Not found file stream by ID:[%s]", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_FD_GET_STREAM_EMPTY
	}
	defer rtkConnection.RemoveFileDropItemStreamListener(fileDropReqData.TimeStamp)
	defer rtkConnection.CloseFileDropItemStream(id, fileDropReqData.TimeStamp)

	fileCount := len(fileDropReqData.SrcFileList)
	folderCount := len(fileDropReqData.FolderList)
	if (fileCount == 0 && folderCount == 0) || fileDropReqData.TimeStamp == 0 {
		log.Printf("[%s] get file data is invalid! fileCount:[%d] folderCount:[%d] TimeStamp:[%d] ", rtkMisc.GetFuncInfo(), fileCount, folderCount, fileDropReqData.TimeStamp)
		return rtkMisc.ERR_BIZ_FD_DATA_INVALID
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
					rtkConnection.CloseFileDropItemStream(id, fileDropReqData.TimeStamp)
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

	copyBuffer := make([]byte, copyBufSize)
	log.Printf("(SRC) Start Copy file data to IP:[%s], id:[%d] file count:[%d] folder count:[%d] totalSize:[%d] TotalDescribe:[%s]...", ipAddr, fileDropReqData.TimeStamp, fileCount, folderCount, fileDropReqData.TotalSize, fileDropReqData.TotalDescribe)

	for _, file := range fileDropReqData.SrcFileList {
		fileSize := uint64(file.FileSize_.SizeHigh)<<32 | uint64(file.FileSize_.SizeLow)
		if fileSize < uint64(copyBufSize) {
			copyBuffer = copyBuffer[:copyBufShortSize]
		} else {
			copyBuffer = copyBuffer[:copyBufSize]
		}
		errCode := writeFileToSocket(id, ipAddr, &cancelableWrite, &cancelableRead, &progressBar, file.FileName, file.FilePath, fileSize, &copyBuffer)
		if errCode != rtkMisc.SUCCESS {
			return errCode
		}
	}

	log.Printf("(SRC) End Copy all file data to IP:[%s] success, id:[%d] file count:[%d] TotalDescribe:[%s], total use [%d] ms", ipAddr, fileDropReqData.TimeStamp, fileCount, fileDropReqData.TotalDescribe, time.Now().UnixMilli()-startTime)

	ShowNotiMessageSendFileTransferDone(fileDropReqData, id)
	return rtkMisc.SUCCESS
}

func writeFileToSocket(id, ip string, write *cancelableWriter, read *cancelableReader, totalBar **ProgressBar, fileName, filePath string, fileSize uint64, buf *[]byte) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	var srcFile *os.File
	errCode := OpenSrcFile(&srcFile, filePath)
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
				log.Printf("(SRC) IP[%s] Copy operation was canceled!", ip)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL
			} else if errors.Is(err, yamux.ErrStreamClosed) {
				if IsInterruptByPeer(id) {
					log.Printf("(SRC) sending file error:%+v", err)
					return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_CANCEL
				}
				log.Printf("(SRC) IP[%s] Copy operation was timeout!", ip)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE_TIMEOUT
			} else {
				log.Printf("[%s] Error sending file:%+v", rtkMisc.GetFuncInfo(), err)
				return rtkMisc.ERR_BIZ_FD_SRC_COPY_FILE
			}
		}
		bufio.NewWriter(write).Flush()
	}
	log.Printf("(SRC) End to copy file:[%s], size:[%d] use [%d] ms", fileName, nCopy, time.Now().UnixMilli()-startTime)
	return rtkMisc.SUCCESS
}

func handleFileDataFromSocket(id, ipAddr string, fileDropData *rtkFileDrop.FilesTransferDataItem) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()

	sFileDrop, ok := rtkConnection.GetFileDropItemStream(id, fileDropData.TimeStamp)
	if !ok {
		log.Printf("[%s] Err: Not found FileDrop stream by ID: %s", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_FD_GET_STREAM_EMPTY
	}
	defer rtkConnection.CloseFileDropItemStream(id, fileDropData.TimeStamp)

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	currentFileSize := uint64(0)
	sentCount := uint32(0)
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
				//cancel by (SRC) maybe block at stream Read, so need interrupt by close stream
				if IsInterruptByPeer(id) {
					time.Sleep(1 * time.Millisecond)
					rtkConnection.CloseFileDropItemStream(id, fileDropData.TimeStamp)
				}
				return
			case <-barTicker.C:
				barCurrentBytes := progressBar.GetCurrentBytes()
				if barLastBytes != barCurrentBytes {
					rtkPlatform.GoUpdateMultipleProgressBar(ipAddr, id, dstFileName, sentCount, nSrcFileCount, currentFileSize, fileDropData.TotalSize, uint64(barCurrentBytes), fileDropData.TimeStamp)
					barLastBytes = barCurrentBytes
					timeoutBarCnt = 0
				} else {
					timeoutBarCnt++
					if timeoutBarCnt >= 100 { // copy file data timeout: 10s
						log.Printf("[%s] (DST) IP[%s] Copy file data time out!", rtkMisc.GetFuncInfo(), ipAddr)
						//TODO:Need to handle timeout
						//rtkConnection.CloseFileDropItemStream(id, fileDropData.TimeStamp)
						//return
						timeoutBarCnt = 0
					}
				}

				if barCurrentBytes >= barMax {
					return
				}
			}
		}
	})

	SetFileTransferInfo(id, cancel)                   //Used when there is an exception on the sending end
	rtkFileDrop.SetCancelFileTransferFunc(id, cancel) //Used when the recipient cancels
	cancelableRead := cancelableReader{
		realReader: nil,
		ctx:        ctx,
	}
	cancelableWrite := cancelableWriter{
		realWriter: nil,
		ctx:        ctx,
	}

	log.Printf("(DST) Start Copy file data from IP:[%s], id:[%d], count:[%d], totalSize:[%d], TotalDescribe:[%s]...", ipAddr, fileDropData.TimeStamp, nSrcFileCount, fileDropData.TotalSize, fileDropData.TotalDescribe)

	copyBuffer := make([]byte, copyBufSize)
	isLastFile := false
	for i, fileInfo := range fileDropData.SrcFileList {
		currentFileSize = uint64(fileInfo.FileSize_.SizeHigh)<<32 | uint64(fileInfo.FileSize_.SizeLow)
		fileName := rtkMisc.AdaptationPath(fileInfo.FileName)
		dstFileName, fileName = rtkUtils.GetTargetDstPathName(filepath.Join(fileDropData.DstFilePath, fileName), fileName)

		if uint32(i) == (nSrcFileCount - 1) {
			isLastFile = true
		}
		cancelableRead.realReader = io.LimitReader(sFileDrop, int64(currentFileSize))

		if currentFileSize < uint64(copyBufSize) {
			copyBuffer = copyBuffer[:copyBufShortSize]
		} else {
			copyBuffer = copyBuffer[:copyBufSize]
		}

		dstTransResult := handleFileFromSocket(ipAddr, id, &cancelableWrite, &cancelableRead, &progressBar, currentFileSize, fileName, dstFileName, &copyBuffer)
		if dstTransResult != rtkMisc.SUCCESS {
			return dstTransResult
		}
		sentCount++
		if !isLastFile {
			rtkPlatform.GoUpdateMultipleProgressBar(ipAddr, id, dstFileName, sentCount, nSrcFileCount, currentFileSize, fileDropData.TotalSize, uint64(progressBar.GetCurrentBytes()), fileDropData.TimeStamp)
		}
	}

	rtkPlatform.GoUpdateMultipleProgressBar(ipAddr, id, dstFileName, sentCount, nSrcFileCount, currentFileSize, fileDropData.TotalSize, fileDropData.TotalSize, fileDropData.TimeStamp)
	log.Printf("(DST) End Copy file data from IP:[%s] success, id:[%d] count:[%d] totalSize:[%d] totalDescribe:[%s] total use:[%d]ms", ipAddr, fileDropData.TimeStamp, nSrcFileCount, fileDropData.TotalSize, fileDropData.TotalDescribe, time.Now().UnixMilli()-startTime)
	ShowNotiMessageRecvFileTransferDone(fileDropData, id)
	return rtkMisc.SUCCESS
}

func handleFileFromSocket(ipAddr, id string, write *cancelableWriter, read *cancelableReader, totalBar **ProgressBar, fileSize uint64, filename, dstFullPath string, buf *[]byte) rtkMisc.CrossShareErr {
	var dstFile *os.File
	startTime := time.Now().UnixMilli()
	err := OpenDstFile(&dstFile, dstFullPath)
	if err != nil {
		return rtkMisc.ERR_BIZ_FD_DST_OPEN_FILE
	}
	defer CloseFile(&dstFile)
	write.realWriter = dstFile

	log.Printf("(DST) IP[%s] Start to copy file:[%s], size:[%d] ...", ipAddr, filename, fileSize)
	nDstWrite := int64(0)
	if fileSize > 0 {
		nDstWrite, err = io.CopyBuffer(io.MultiWriter(write, *totalBar), read, *buf)
		if err != nil {
			CloseFile(&dstFile)
			DeleteFile(dstFullPath)
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("Error Read file timeout:", netErr)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT
			} else if errors.Is(err, context.Canceled) {
				if rtkFileDrop.IsCancelFileTransferByGui(id) {
					log.Printf("(DST) IP[%s] Copy operation was canceled by dst platform GUI!", ipAddr)
					return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI
				}
				log.Printf("(DST) IP[%s] Copy operation was canceled!", ipAddr)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL
			} else if errors.Is(err, yamux.ErrStreamClosed) {
				if IsInterruptByPeer(id) {
					log.Printf("(DST) IP[%s] Copy operation was canceled!", ipAddr)
					return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_CANCEL
				}
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT
			} else {
				log.Printf("[%s] IP:[%s] Copy file Error:%+v", rtkMisc.GetFuncInfo(), ipAddr, err)
				return rtkMisc.ERR_BIZ_FD_DST_COPY_FILE
			}
		}
	}

	if uint64(nDstWrite) >= fileSize {
		log.Printf("(DST) IP[%s] End to Copy file:[%s] success, total:[%d] use [%d] ms", ipAddr, filename, nDstWrite, time.Now().UnixMilli()-startTime)
		return rtkMisc.SUCCESS
	} else {
		log.Printf("(DST) IP[%s] End to Copy file:[%s] failed, total:[%d], it less then filesize:[%d]...", ipAddr, filename, nDstWrite, fileSize)
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
