package filedrop

import (
	"context"
	"encoding/json"
	"log"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
)

func updateFileListDrop(id string, fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc, srcRootPath string) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	fileDropDataMap[id] = FileDropData{
		SrcFileList:          fileInfoList,
		ActionType:           rtkCommon.P2PFileActionType_Drop,
		TimeStamp:            timeStamp,
		FolderList:           folderList,
		SrcRootPath:   	      srcRootPath,
		TotalDescribe:        totalDesc,
		TotalSize:            total,
		DstFilePath:          "",
		Cmd:                  rtkCommon.FILE_DROP_REQUEST,
		InterruptSrcFileName: "",
		InterruptDstFileName: "",
		InterruptFileOffSet:  0,
		InterruptLastErrCode: rtkMisc.SUCCESS,
	}
}

func UpdateFileListDropReqDataFromLocal(id string, fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc, srcRootPath string) {
	updateFileListDrop(id, fileInfoList, folderList, total, timeStamp, totalDesc, srcRootPath)

	clientListMap := rtkUtils.GetClientMap()
	ipAddr := string("")
	for _, clientInfo := range clientListMap {
		if clientInfo.ID == id {
			ipAddr = clientInfo.IpAddr
		}
		fileDropReqIdChan <- id
	}

	if ipAddr == "" {
		log.Printf("[%s] ID:[%s] Not found client map data\n\n", rtkMisc.GetFuncInfo(), id)
		return
	}

	firstFileSize := uint64(0)
	var firstFileName string
	if len(fileInfoList) > 0 {
		firstFileSize = uint64(fileInfoList[0].FileSize_.SizeHigh)<<32 | uint64(fileInfoList[0].FileSize_.SizeLow)
		firstFileName = fileInfoList[0].FileName
	} else {
		firstFileName = folderList[0]
	}

	rtkPlatform.GoFileListSendNotify(ipAddr, id, uint32(len(fileInfoList)), total, timeStamp, firstFileName, firstFileSize, getFileDropDataDetails(id, ipAddr))
}

func UpdateFileListDropReqDataFromDst(id string, fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc string) {
	updateFileListDrop(id, fileInfoList, folderList, total, timeStamp, totalDesc, "")
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

func getFileDropDataDetails(id, ipAddr string) string {
	fileDropDataMutex.RLock()
	fileDropData, ok := fileDropDataMap[id]
	fileDropDataMutex.RUnlock()

	if !ok {
		log.Printf("[%s], ID[%s] Not fount file data", rtkMisc.GetFuncInfo(), id)
		return ""
	}

	notifyInfo := FileDataTransDetails{
		ID:            id,
		IPAddr:        ipAddr,
		TimeStamp:     fileDropData.TimeStamp,
		FileList:      make([]FileInfoEx, 0),
		TotalSize:     fileDropData.TotalSize,
		TotalDescribe: fileDropData.TotalDescribe,
	}

	for i, fileInfo := range fileDropData.SrcFileList {
		var fileInfoEx FileInfoEx
		fileInfoEx.FileSize = uint64(fileInfo.FileSize_.SizeHigh)<<32 | uint64(fileInfo.FileSize_.SizeLow)
		if fileDropData.Cmd == rtkCommon.FILE_DROP_ACCEPT && fileDropData.DstFilePath != "" { // as Dst Notify,  update dst full path
			fileName := rtkMisc.AdaptationPath(fileInfo.FileName)
			fileInfoEx.FilePath, fileInfoEx.FileName = rtkUtils.GetTargetDstPathName(filepath.Join(fileDropData.DstFilePath, fileName), fileName)
		} else {
			fileInfoEx.FilePath = fileInfo.FilePath
			fileInfoEx.FileName = fileInfo.FileName
		}

		if i == 0 {
			notifyInfo.FirstFileSize = fileInfoEx.FileSize
			notifyInfo.FirstFileName = fileInfoEx.FileName
		}

		notifyInfo.FileList = append(notifyInfo.FileList, fileInfoEx)
	}

	if fileDropData.Cmd == rtkCommon.FILE_DROP_ACCEPT && fileDropData.DstFilePath != "" { // DST
		notifyInfo.FolderList = make([]string, 0)
		for _, folder := range fileDropData.FolderList {
			notifyInfo.FolderList = append(notifyInfo.FolderList, rtkMisc.AdaptationPath(folder))
		}
		notifyInfo.RootPath = fileDropData.DstFilePath
	} else { // SRC
		notifyInfo.FolderList = fileDropData.FolderList
		notifyInfo.RootPath = fileDropData.SrcRootPath
	}

	encodedData, err := json.Marshal(notifyInfo)
	if err != nil {
		log.Printf("[%s] Failed to Marshal FileDataTransDetails data, err: %+v", rtkMisc.GetFuncInfo(), err)
		return ""
	}

	return string(encodedData)
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
		UpdateFileDropRespDataFromDst(id, rtkCommon.FILE_DROP_ACCEPT, rtkPlatform.GetDownloadPath())
		rtkPlatform.GoFileListReceiveNotify(ip, id, nFileCount, totalSize, timestamp, firstFileName, firstFileSize, getFileDropDataDetails(id, ip)) //No need to confirm
	}
}
