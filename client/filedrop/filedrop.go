package filedrop

import (
	"context"
	"log"
	"path/filepath"
	rtkCommon "rtk-cross-share/common"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"sync"
	"time"
)

type FileDropData struct {
	// Req data
	SrcFileInfo rtkCommon.FileInfo
	TimeStamp   uint64
	// Resp data
	DstFilePath string
	Cmd         rtkCommon.FileDropCmd
}

var (
	fileDropDataMap             = make(map[string]FileDropData) // key: IP
	fileDropDataMutex           sync.RWMutex
	isFileDropReqDataFromLocal  = make(map[string]bool) // key: IP
	isFileDropRespDataFromLocal = make(map[string]bool) // key: IP
)

func UpdateFileDropReqDataFromLocal(ip string, fileInfo rtkCommon.FileInfo, timestamp int64) {
	updateFileDropReqData(ip, fileInfo, timestamp)
	isFileDropReqDataFromLocal[ip] = true
}

func UpdateFileDropReqDataFromDst(ip string, fileInfo rtkCommon.FileInfo, timestamp int64) {
	updateFileDropReqData(ip, fileInfo, timestamp)
	isFileDropReqDataFromLocal[ip] = false
}

func updateFileDropReqData(ip string, fileInfo rtkCommon.FileInfo, timestamp int64) {
	fileDropDataMutex.Lock()
	fileDropDataMap[ip] = FileDropData{
		SrcFileInfo: fileInfo,
		TimeStamp:   uint64(timestamp),
		DstFilePath: "",
		Cmd:         rtkCommon.FILE_DROP_REQUEST,
	}
	fileDropDataMutex.Unlock()
}

func UpdateFileDropRespDataFromLocal(ip string, cmd rtkCommon.FileDropCmd, filePath string) {
	updateFileDropRespData(ip, cmd, filePath)
	isFileDropRespDataFromLocal[ip] = true
}

func UpdateFileDropRespDataFromDst(ip string, cmd rtkCommon.FileDropCmd, filePath string) {
	updateFileDropRespData(ip, cmd, filePath)
	isFileDropRespDataFromLocal[ip] = false
}

func updateFileDropRespData(ip string, cmd rtkCommon.FileDropCmd, filePath string) {
	fileDropDataMutex.Lock()
	if fileDropData, ok := fileDropDataMap[ip]; ok {
		if fileDropData.Cmd == rtkCommon.FILE_DROP_REQUEST {
			fileDropData.DstFilePath = filePath
			fileDropData.Cmd = cmd
			fileDropDataMap[ip] = fileDropData
		} else {
			log.Printf("[%s %d] Err: Update file drop failed. Invalid state", rtkUtils.GetFuncName(), rtkUtils.GetLine())
		}
	}
	fileDropDataMutex.Unlock()
}

func GetFileDropData(ip string) (FileDropData, bool) {
	fileDropDataMutex.RLock()
	fileDropData, ok := fileDropDataMap[ip]
	fileDropDataMutex.RUnlock()
	return fileDropData, ok
}

func ResetFileDropData(ip string) {
	fileDropDataMutex.Lock()
	delete(fileDropDataMap, ip)
	fileDropDataMutex.Unlock()
}

func WatchFileDropReqEvent(ctx context.Context, ipAddr string, resultChan chan<- string) {
	rtkPlatform.SetGoFileDropRequestCallback(UpdateFileDropReqDataFromLocal)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			if isLocal, ok := isFileDropReqDataFromLocal[ipAddr]; ok {
				if isLocal {
					resultChan <- ipAddr
					isFileDropReqDataFromLocal[ipAddr] = false
				}
			}
		}
	}
}

func WatchFileDropRespEvent(ctx context.Context, ipAddr string, resultChan chan<- string) {
	rtkPlatform.SetGoFileDropResponseCallback(UpdateFileDropRespDataFromLocal)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			if isLocal, ok := isFileDropRespDataFromLocal[ipAddr]; ok {
				if isLocal {
					resultChan <- ipAddr
					isFileDropRespDataFromLocal[ipAddr] = false
				}
			}
		}
	}
}

func SetupDstFileDrop(ip, id, filePath, platform string, fileSizeHigh uint32, fileSizeLow uint32, timestamp int64) {
	fileInfo := rtkCommon.FileInfo{
		FileSize_: rtkCommon.FileSize{
			SizeHigh: fileSizeHigh,
			SizeLow:  fileSizeLow,
		},
		FilePath: filePath,
	}
	UpdateFileDropReqDataFromDst(ip, fileInfo, timestamp)
	fileSize := uint64(fileSizeHigh)<<32 | uint64(fileSizeLow)
	rtkPlatform.GoSetupFileDrop(ip, id, filepath.Base(filePath), platform, fileSize, timestamp)
}
