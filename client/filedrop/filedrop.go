package filedrop

import (
	"context"
	"log"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
)

func updateFileListDrop(id string, fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc string) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	fileDropDataMap[id] = FileDropData{
		SrcFileList:            fileInfoList,
		ActionType:             rtkCommon.P2PFileActionType_Drop,
		TimeStamp:              timeStamp,
		FolderList:             folderList,
		TotalDescribe:          totalDesc,
		TotalSize:              total,
		DstFilePath:            "",
		Cmd:                    rtkCommon.FILE_DROP_REQUEST,
		InterruptFileName:      "",
		InterruptFileOffSet:    0,
		InterruptFileTimeStamp: 0,
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

	dragFileInfoList = nil
	dragFolderList = nil
	dragFileTimeStamp = 0
	dragTotalSize = 0
	dragTotalDesc = ""
}

func init() {
	dragFileInfoList = nil
	dragFolderList = nil
	dragFileTimeStamp = 0
	dragTotalSize = 0
	dragTotalDesc = ""

	rtkPlatform.SetGoFileDropResponseCallback(UpdateFileDropRespDataFromLocal) //pop-up confirmation
	rtkPlatform.SetGoFileListDropRequestCallback(UpdateFileListDropReqDataFromLocal)

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
