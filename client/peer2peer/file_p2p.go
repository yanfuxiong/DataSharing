package peer2peer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
)

var (
	fileTransferInfoMap sync.Map //key: ID
)

type fileTransferInfo struct {
	caneclFn       func()
	isCaneclByPeer bool // true: cancel by peer and not need send interrupt message to peer
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

func addSuffixBeforeExt(path, suffix string) string {
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)
	return fmt.Sprintf("%s%s%s", name, suffix, ext)
}

func getTargetDstPathName(dstFullPath, dstFileName string) (string, string) {
	index := uint(0)
	var dstPath string

	for {
		if index == 0 {
			dstPath = dstFullPath
		} else {
			dstPath = addSuffixBeforeExt(dstFullPath, fmt.Sprintf(" (%d)", index))
		}
		if !rtkMisc.FileExists(dstPath) {
			if index == 0 {
				return dstPath, dstFileName
			} else {
				return dstPath, addSuffixBeforeExt(dstFileName, fmt.Sprintf(" (%d)", index))
			}

		}
		index++
	}
}

func OpenSrcFile(file **os.File, filePath string) rtkMisc.CrossShareErr {
	if *file != nil {
		(*file).Close()
		*file = nil
	}

	if !rtkMisc.FileExists(filePath) {
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
		caneclFn:       fn,
		isCaneclByPeer: false,
	})
}

func CancelSrcFileTransfer(id string) {
	if value, ok := fileTransferInfoMap.Load(id); ok {
		fileInfo := value.(fileTransferInfo)
		fileInfo.isCaneclByPeer = true
		fileTransferInfoMap.Store(id, fileInfo)
		log.Printf("[%s] (SRC) ID:[%s] Cancel FileTransfer success!", rtkMisc.GetFuncInfo(), id)
		fileInfo.caneclFn()
	}
}

func CancelDstFileTransfer(id string) {
	if value, ok := fileTransferInfoMap.Load(id); ok {
		fileInfo := value.(fileTransferInfo)
		fileInfo.isCaneclByPeer = true
		fileTransferInfoMap.Store(id, fileInfo)
		log.Printf("[%s] (DST) ID:[%s] Cancel FileTransfer success!", rtkMisc.GetFuncInfo(), id)
		fileInfo.caneclFn()
	}
}

func IsInterruptByPeer(id string) bool {
	if value, ok := fileTransferInfoMap.Load(id); ok {
		fileInfo := value.(fileTransferInfo)
		return fileInfo.isCaneclByPeer
	}
	return false
}

func writeFileDataToSocket(id, ipAddr string) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	var sFileDrop network.Stream

	fileDropReqData, ok := rtkFileDrop.GetFileDropData(id)
	if !ok {
		log.Printf("[%s] Err: Not found fileDrop data", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_BIZ_FD_DATA_EMPTY
	}

	rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.FILE_DROP) // wait for file drop stream Ready
	sFileDrop, ok = rtkConnection.GetFmtTypeStream(id, rtkCommon.FILE_DROP)
	if !ok {
		log.Printf("[%s] Err: Not found file stream by ID:[%s]", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_FD_GET_STREAM_EMPTY
	}
	defer rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)

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
					rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
				}
				return
			// TODO: Update sender progress bar
			case <-timeOutTk.C:
				barCurrentBytes := progressBar.GetCurrentBytes()
				if barLastBytes == barCurrentBytes { // the file copy is timeout, 10s
					log.Printf("[%s] (SRC) IP[%s] Copy file data time out!", rtkMisc.GetFuncInfo(), ipAddr)

					//TODO:Need to handle timeout
					//io.Copy maybe block Exceeding 10 seconds, and the reason may be congestion control, packet loss retransmission, or disk I/O lag
					//rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
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

	if fileDropReqData.FileType == rtkCommon.P2PFile_Type_Multiple {
		log.Printf("(SRC) Start Copy file data to IP:[%s], file count:[%d] folder count:[%d] totalSize:[%d] TotalDescribe:[%s] ...", ipAddr, fileCount, folderCount, fileDropReqData.TotalSize, fileDropReqData.TotalDescribe)
	}

	for _, file := range fileDropReqData.SrcFileList {
		fileSize := uint64(file.FileSize_.SizeHigh)<<32 | uint64(file.FileSize_.SizeLow)
		errCode := writeFileToSocket(id, ipAddr, &cancelableWrite, &cancelableRead, &progressBar, file.FileName, file.FilePath, fileSize)
		if errCode != rtkMisc.SUCCESS {
			return errCode
		}
	}

	if fileDropReqData.FileType == rtkCommon.P2PFile_Type_Multiple {
		log.Printf("(SRC) End to Copy all file data to IP:[%s] success,TotalDescribe:[%s], total use [%d] ms", ipAddr, fileDropReqData.TotalDescribe, time.Now().UnixMilli()-startTime)
	}
	ShowNotiMessageSendFileTransferDone(fileDropReqData, id)
	return rtkMisc.SUCCESS
}

func writeFileToSocket(id, ip string, write *cancelableWriter, read *cancelableReader, totalBar **ProgressBar, fileName, filePath string, filesize uint64) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	var srcFile *os.File
	errCode := OpenSrcFile(&srcFile, filePath)
	if errCode != rtkMisc.SUCCESS {
		log.Printf("[%s] OpenSrcFile err code:[%d]", rtkMisc.GetFuncInfo(), errCode)
		return errCode
	}
	defer CloseFile(&srcFile)
	read.realReader = srcFile

	log.Printf("(SRC) Start to copy file:[%s] size:[%d] ...", fileName, filesize)
	nCopy := int64(0)
	var err error
	if filesize > 0 {
		nCopy, err = io.Copy(io.MultiWriter(write, *totalBar), read)
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

func handleFileDataFromSocket(id, ipAddr, deviceName string) (string, rtkMisc.CrossShareErr) {
	startTime := time.Now().UnixMilli()

	fileDropData, ok := rtkFileDrop.GetFileDropData(id)
	if !ok {
		log.Printf("[%s] Not found fileDrop data", rtkMisc.GetFuncInfo())
		return "", rtkMisc.ERR_BIZ_FD_DATA_EMPTY
	}

	dstFileName := fileDropData.DstFilePath
	sFileDrop, ok := rtkConnection.GetFmtTypeStream(id, rtkCommon.FILE_DROP)
	if !ok {
		log.Printf("[%s] Err: Not found FileDrop stream by ID: %s", rtkMisc.GetFuncInfo(), id)
		return dstFileName, rtkMisc.ERR_BIZ_FD_GET_STREAM_EMPTY
	}
	defer rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)

	if fileDropData.Cmd != rtkCommon.FILE_DROP_ACCEPT {
		log.Printf("[%s] Invalid fildDrop cmd state:%s", rtkMisc.GetFuncInfo(), fileDropData.Cmd)
		return dstFileName, rtkMisc.ERR_BIZ_FD_UNKNOWN_CMD
	}

	if fileDropData.FileType != rtkCommon.P2PFile_Type_Single && fileDropData.FileType != rtkCommon.P2PFile_Type_Multiple {
		log.Printf("[%s] Invalid FileType :[%s]", rtkMisc.GetFuncInfo(), fileDropData.FileType)
		return dstFileName, rtkMisc.ERR_BIZ_FD_UNKNOWN_TYPE
	}

	nSrcFileCount := uint32(len(fileDropData.SrcFileList))
	folderCount := uint32(len(fileDropData.FolderList))
	if (nSrcFileCount == 0 && folderCount == 0) || fileDropData.TimeStamp == 0 {
		log.Printf("[%s] get file data is invalid! fileCount:[%d] folderCount:[%d] TimeStamp:[%d] ", rtkMisc.GetFuncInfo(), nSrcFileCount, folderCount, fileDropData.TimeStamp)
		return dstFileName, rtkMisc.ERR_BIZ_FD_DATA_INVALID
	}

	var currentFileSize uint64
	sentCount := uint32(0)
	if fileDropData.FileType == rtkCommon.P2PFile_Type_Multiple {
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
		log.Printf("(DST) Start Copy file list, count:[%d], totalSize:[%d], TotalDescribe:[%s]", nSrcFileCount, fileDropData.TotalSize, fileDropData.TotalDescribe)
	}

	progressBar := New64(int64(fileDropData.TotalSize))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
					rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
				}
				return
			case <-barTicker.C:
				barCurrentBytes := progressBar.GetCurrentBytes()
				if barLastBytes != barCurrentBytes {
					rtkPlatform.GoUpdateMultipleProgressBar(ipAddr, id, deviceName, dstFileName, sentCount, nSrcFileCount, currentFileSize, fileDropData.TotalSize, uint64(barCurrentBytes), fileDropData.TimeStamp)
					barLastBytes = barCurrentBytes
					timeoutBarCnt = 0
				} else {
					timeoutBarCnt++
					if timeoutBarCnt >= 100 { // copy file data timeout: 10s
						log.Printf("[%s] (DST) IP[%s] Copy file data time out!", rtkMisc.GetFuncInfo(), ipAddr)
						//TODO:Need to handle timeout
						//rtkConnection.CloseFmtTypeStream(id, rtkCommon.FILE_DROP)
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
		realReader: sFileDrop,
		ctx:        ctx,
	}
	cancelableWrite := cancelableWriter{
		realWriter: nil,
		ctx:        ctx,
	}

	for i, fileInfo := range fileDropData.SrcFileList {
		currentFileSize = uint64(fileInfo.FileSize_.SizeHigh)<<32 | uint64(fileInfo.FileSize_.SizeLow)
		fileName := rtkMisc.AdaptationPath(fileInfo.FileName)
		dstFileName, fileName = getTargetDstPathName(filepath.Join(fileDropData.DstFilePath, fileName), fileName)

		dstTransResult := handleFileFromSocket(ipAddr, id, &cancelableWrite, &cancelableRead, &progressBar, currentFileSize, fileName, dstFileName)
		if dstTransResult != rtkMisc.SUCCESS {
			return dstFileName, dstTransResult
		}
		sentCount++
		if fileDropData.FileType == rtkCommon.P2PFile_Type_Multiple && uint32(i) != (nSrcFileCount-1) {
			rtkPlatform.GoUpdateMultipleProgressBar(ipAddr, id, deviceName, dstFileName, sentCount, nSrcFileCount, currentFileSize, fileDropData.TotalSize, uint64(progressBar.GetCurrentBytes()), fileDropData.TimeStamp)
		}
	}

	if fileDropData.FileType == rtkCommon.P2PFile_Type_Multiple {
		rtkPlatform.GoUpdateMultipleProgressBar(ipAddr, id, deviceName, dstFileName, sentCount, nSrcFileCount, currentFileSize, fileDropData.TotalSize, fileDropData.TotalSize, fileDropData.TimeStamp)
		log.Printf("(DST) End Copy file list success, count:[%d] totalSize:[%d] total use:[%d]ms", nSrcFileCount, fileDropData.TotalSize, time.Now().UnixMilli()-startTime)
	} else {
		rtkPlatform.GoUpdateProgressBar(ipAddr, id, fileDropData.TotalSize, fileDropData.TotalSize, fileDropData.TimeStamp, dstFileName)
	}
	ShowNotiMessageRecvFileTransferDone(fileDropData, id)
	return dstFileName, rtkMisc.SUCCESS
}

func handleFileFromSocket(ipAddr, id string, write *cancelableWriter, read *cancelableReader, totalBar **ProgressBar, fileSize uint64, filename, dstFullPath string) rtkMisc.CrossShareErr {
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
		nDstWrite, err = io.CopyN(io.MultiWriter(write, *totalBar), read, int64(fileSize))
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

func ShowNotiMessageSendFileTransferDone(fileDropData rtkFileDrop.FileDropData, id string) {
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

func ShowNotiMessageRecvFileTransferDone(fileDropData rtkFileDrop.FileDropData, id string) {
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
