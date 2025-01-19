package filedrop

import (
	"context"
	"log"
	"path/filepath"
	rtkCommon "rtk-cross-share/common"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"sync"
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
	fileDropDataMap    = make(map[string]FileDropData) // key: ID
	fileDropDataMutex  sync.RWMutex
	fileDropReqIdChan  = make(chan string, 10)
	fileDropRespIdChan = make(chan string, 10)
)

func UpdateFileDropReqDataFromLocal(id string, fileInfo rtkCommon.FileInfo, timestamp int64) {
	updateFileDropReqData(id, fileInfo, timestamp)

	for i := 0; i < rtkUtils.GetClientCount(); i++ {
		fileDropReqIdChan <- id
	}
}

func UpdateFileDropReqDataFromDst(id string, fileInfo rtkCommon.FileInfo, timestamp int64) {
	updateFileDropReqData(id, fileInfo, timestamp)
}

func updateFileDropReqData(id string, fileInfo rtkCommon.FileInfo, timestamp int64) {
	fileDropDataMutex.Lock()
	fileDropDataMap[id] = FileDropData{
		SrcFileInfo: fileInfo,
		TimeStamp:   uint64(timestamp),
		DstFilePath: "",
		Cmd:         rtkCommon.FILE_DROP_REQUEST,
	}
	fileDropDataMutex.Unlock()
}

func UpdateFileDropRespDataFromLocal(id string, cmd rtkCommon.FileDropCmd, filePath string) {
	updateFileDropRespData(id, cmd, filePath)

	for i := 0; i < rtkUtils.GetClientCount(); i++ {
		fileDropRespIdChan <- id
	}
}

func UpdateFileDropRespDataFromDst(id string, cmd rtkCommon.FileDropCmd, filePath string) {
	updateFileDropRespData(id, cmd, filePath)
}

func updateFileDropRespData(id string, cmd rtkCommon.FileDropCmd, filePath string) {
	fileDropDataMutex.Lock()
	if fileDropData, ok := fileDropDataMap[id]; ok {
		if fileDropData.Cmd == rtkCommon.FILE_DROP_REQUEST {
			fileDropData.DstFilePath = filePath
			fileDropData.Cmd = cmd
			fileDropDataMap[id] = fileDropData
		} else {
			log.Printf("[%s %d] Err: Update file drop failed. Invalid state", rtkUtils.GetFuncName(), rtkUtils.GetLine())
		}
	}
	fileDropDataMutex.Unlock()
}

func GetFileDropData(id string) (FileDropData, bool) {
	fileDropDataMutex.RLock()
	fileDropData, ok := fileDropDataMap[id]
	fileDropDataMutex.RUnlock()
	return fileDropData, ok
}

func ResetFileDropData(id string) {
	fileDropDataMutex.Lock()
	delete(fileDropDataMap, id)
	fileDropDataMutex.Unlock()
}

func InitFileDrop() {
	rtkPlatform.SetGoFileDropRequestCallback(UpdateFileDropReqDataFromLocal)
	rtkPlatform.SetGoFileDropResponseCallback(UpdateFileDropRespDataFromLocal)
}

func WatchFileDropReqEvent(ctx context.Context, id string, resultChan chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return
		case triggerId := <-fileDropReqIdChan:
			if triggerId == id {
				resultChan <- id
			}
		}
	}
}

func WatchFileDropRespEvent(ctx context.Context, id string, resultChan chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return
		case triggerId := <-fileDropRespIdChan:
			if triggerId == id {
				resultChan <- id
			}
		}
	}
}

func SetupDstFileDrop(id, filePath, platform string, fileSizeHigh uint32, fileSizeLow uint32, timestamp int64) {
	ipAddr, _ := rtkUtils.GetClientIp(id)

	fileInfo := rtkCommon.FileInfo{
		FileSize_: rtkCommon.FileSize{
			SizeHigh: fileSizeHigh,
			SizeLow:  fileSizeLow,
		},
		FilePath: filePath,
	}
	UpdateFileDropReqDataFromDst(id, fileInfo, timestamp)
	fileSize := uint64(fileSizeHigh)<<32 | uint64(fileSizeLow)
	rtkPlatform.GoSetupFileDrop(ipAddr, id, filepath.Base(filePath), platform, fileSize, timestamp)
}
