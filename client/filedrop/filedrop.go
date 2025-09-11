package filedrop

import (
	"context"
	"log"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync"
)

type FileDropData struct {
	// Req data
	SrcFileList   []rtkCommon.FileInfo
	ActionType    rtkCommon.FileActionType
	TimeStamp     uint64
	FolderList    []string // must start with folder name and end with '/'.  eg: folderName/aaa/bbb/
	TotalDescribe string   // eg: 820MB /1.2GB
	TotalSize     uint64

	// Resp data
	DstFilePath string
	Cmd         rtkCommon.FileDropCmd
}

type cancelInfo struct {
	cancelFn      func()
	isCancelByGui bool
}

var (
	fileDropDataMap    = make(map[string]FileDropData) // key: ID
	fileDropDataMutex  sync.RWMutex
	fileDropReqIdChan  = make(chan string, 10)
	fileDropRespIdChan = make(chan string, 10)

	cancelFileTransMap sync.Map //key: ID

	dragFileInfoList  []rtkCommon.FileInfo
	dragFolderList    []string
	dragFileTimeStamp uint64
	dragTotalSize     uint64
	dragTotalDesc     string
)

func updateFileListDrop(id string, fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc string) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	fileDropDataMap[id] = FileDropData{
		SrcFileList:   append([]rtkCommon.FileInfo(nil), fileInfoList...),
		ActionType:    rtkCommon.P2PFileActionType_Drop,
		TimeStamp:     timeStamp,
		FolderList:    append([]string(nil), folderList...),
		TotalDescribe: totalDesc,
		TotalSize:     total,

		DstFilePath: "",
		Cmd:         rtkCommon.FILE_DROP_REQUEST,
	}

}

func UpdateFileListDropReqDataFromLocal(id string, fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc string) {
	updateFileListDrop(id, fileInfoList, folderList, total, timeStamp, totalDesc)

	nCount := rtkUtils.GetClientCount()
	for i := 0; i < nCount; i++ {
		fileDropReqIdChan <- id
	}
}

func UpdateFileListDropReqDataFromDst(id string, fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc string) {
	updateFileListDrop(id, fileInfoList, folderList, total, timeStamp, totalDesc)
}

func UpdateFileDropRespDataFromLocal(id string, cmd rtkCommon.FileDropCmd, filePath string) {
	updateFileDropRespData(id, cmd, filePath)

	nCount := rtkUtils.GetClientCount()
	for i := 0; i < nCount; i++ {
		fileDropRespIdChan <- id
	}
}

func UpdateFileDropRespDataFromDst(id string, cmd rtkCommon.FileDropCmd, filePath string) {
	updateFileDropRespData(id, cmd, filePath)

	nCount := rtkUtils.GetClientCount()
	for i := 0; i < nCount; i++ {
		fileDropRespIdChan <- id
	}
}

func updateFileDropRespData(id string, cmd rtkCommon.FileDropCmd, filePath string) {
	fileDropDataMutex.Lock()
	if fileDropData, ok := fileDropDataMap[id]; ok {
		if fileDropData.ActionType != rtkCommon.P2PFileActionType_Drop {
			log.Printf("[%s] Err: Update file drop failed. Invalid ActionType: %s", rtkMisc.GetFuncInfo(), fileDropData.ActionType)
			return
		}

		if fileDropData.Cmd == rtkCommon.FILE_DROP_REQUEST {
			if cmd == rtkCommon.FILE_DROP_ACCEPT {
				if rtkMisc.FolderExists(filePath) {
					fileDropData.DstFilePath = filePath
				} else {
					fileDropData.DstFilePath = filepath.Dir(filePath)
					fileDropData.SrcFileList[0].FileName = filepath.Base(filePath)
				}
			}

			fileDropData.Cmd = cmd
			fileDropDataMap[id] = fileDropData
		} else {
			log.Printf("[%s] Err: Update file drop failed. Invalid state:%s", rtkMisc.GetFuncInfo(), fileDropData.Cmd)
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

	cancelFileTransMap.Delete(id)
	dragFileInfoList = nil
	dragFolderList = nil
	dragFileTimeStamp = 0
	dragTotalSize = 0
	dragTotalDesc = ""
}

func init() {
	cancelFileTransMap.Clear()
	dragFileInfoList = nil
	dragFolderList = nil
	dragFileTimeStamp = 0
	dragTotalSize = 0
	dragTotalDesc = ""

	rtkPlatform.SetGoFileDropResponseCallback(UpdateFileDropRespDataFromLocal) //pop-up confirmation
	rtkPlatform.SetGoFileListDropRequestCallback(UpdateFileListDropReqDataFromLocal)

	rtkPlatform.SetGoCancelFileTransCallback(CancelFileTransfer)
}

func CancelFileTransfer(id, ip string, timestamp uint64) {
	if value, ok := cancelFileTransMap.Load(id); ok {
		currentFileData, exist := GetFileDropData(id)
		if !exist {
			log.Printf("[%s] ID:[%s],IPAddr:[%s], Not fount file data", rtkMisc.GetFuncInfo(), id, ip)
			return
		}
		if currentFileData.TimeStamp != timestamp {
			log.Printf("[%s] ID:[%s],IPAddr:[%s], timestamp:[%d] invalid! ", rtkMisc.GetFuncInfo(), id, ip, timestamp)
			return
		}
		cancelData := value.(cancelInfo)
		if !cancelData.isCancelByGui {
			cancelData.cancelFn()
			cancelData.cancelFn = nil
			cancelData.isCancelByGui = true
			cancelFileTransMap.Store(id, cancelData)
			log.Printf("ID:[%s],IPAddr:[%s] CancelFileTransfer success by platform GUI!", id, ip)
		}
	} else {
		log.Printf("[%s] ID:[%s],IPAddr:[%s] get no cancelFileTransMap data!", rtkMisc.GetFuncInfo(), id, ip)
	}
}

func SetCancelFileTransferFunc(id string, fn func()) {
	cancelFileTransMap.Store(id, cancelInfo{
		cancelFn:      fn,
		isCancelByGui: false,
	})
}

func IsCancelFileTransferByGui(id string) bool {
	if value, ok := cancelFileTransMap.Load(id); ok {
		return value.(cancelInfo).isCancelByGui
	}
	return false
}

func WatchFileDropReqEvent(ctx context.Context, id string, resultChan chan<- string) {
	for {
		select {
		case <-ctx.Done():
			close(resultChan)
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
			close(resultChan)
			return
		case triggerId := <-fileDropRespIdChan:
			if triggerId == id {
				resultChan <- id
			}
		}
	}
}

// ********************  Setup Dst file info ****************

func SetupDstFileListDrop(id, ip, platform, totalDesc string, fileList []rtkCommon.FileInfo, folderList []string, totalSize, timestamp uint64) {
	UpdateFileListDropReqDataFromDst(id, fileList, folderList, totalSize, timestamp, totalDesc)

	if rtkPlatform.GetConfirmDocumentsAccept() {
		rtkPlatform.GoSetupFileListDrop(ip, id, platform, totalDesc, uint32(len(fileList)), uint32(len(folderList)), timestamp) // need pop-up confirmation
	} else {
		nFileCount := uint32(len(fileList))
		firstFileSize := uint64(0)
		firstFileName := string("")
		if nFileCount > 0 {
			firstFileSize = uint64(fileList[0].FileSize_.SizeHigh)<<32 | uint64(fileList[0].FileSize_.SizeLow)
			firstFileName = filepath.Join(rtkPlatform.GetDownloadPath(), rtkMisc.AdaptationPath(fileList[0].FileName))
		} else {
			firstFileName = filepath.Join(rtkMisc.AdaptationPath(folderList[0]))
		}

		firstFileName, _ = rtkUtils.GetTargetDstPathName(firstFileName, "")
		rtkPlatform.GoMultiFilesDropNotify(ip, id, platform, nFileCount, totalSize, timestamp, firstFileName, firstFileSize) //No need to confirm
		UpdateFileDropRespDataFromDst(id, rtkCommon.FILE_DROP_ACCEPT, rtkPlatform.GetDownloadPath())
	}
}
