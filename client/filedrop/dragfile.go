package filedrop

import (
	"log"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
)

func init() {

	rtkPlatform.SetGoDragFileCallback(UpdateFileDragInfo)
	rtkPlatform.SetGoDragFileListRequestCallback(UpdateDragFileListFromLocal)
}

func UpdateFileDragInfo(fileinfo rtkCommon.FileInfo, timestamp uint64) {
	dragFileInfoList = []rtkCommon.FileInfo{fileinfo}
	dragFolderList = nil
	dragFileTimeStamp = timestamp
	dragTotalSize = uint64(fileinfo.FileSize_.SizeHigh)<<32 | uint64(fileinfo.FileSize_.SizeLow)
	dragTotalDesc = rtkMisc.FileSizeDesc(dragTotalSize)
}

func UpdateDragFileReqDataFromLocal(id string) rtkMisc.CrossShareErr {
	fileCnt := len(dragFileInfoList)
	folderCnt := len(dragFolderList)
	if fileCnt == 0 && folderCnt == 0 {
		log.Printf("[%s] get Drag File info is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_BIZ_DF_DATA_EMPTY
	}

	if dragFileTimeStamp == 0 {
		log.Printf("[%s] get Drag File timestamp is 0", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_BIZ_DF_INVALID_TIMESTAMP
	}

	if fileCnt > 0 && !rtkMisc.FileExists(dragFileInfoList[0].FilePath) {
		log.Printf("[%s] get File:[%s] info error", rtkMisc.GetFuncInfo(), dragFileInfoList[0].FilePath)
		return rtkMisc.ERR_BIZ_DF_FILE_NOT_EXISTS
	}

	updateDragFileReqData(id)
	nCount := rtkUtils.GetClientCount()
	for i := 0; i < nCount; i++ {
		fileDropReqIdChan <- id
	}

	return rtkMisc.SUCCESS
}

func UpdateDragFileReqDataFromDst(id string) {
	updateDragFileReqData(id)
}

func updateDragFileReqData(id string) {
	targetData := FileDropData{
		SrcFileList:   append([]rtkCommon.FileInfo(nil), dragFileInfoList...),
		TimeStamp:     dragFileTimeStamp,
		ActionType:    rtkCommon.P2PFileActionType_Drag,
		TotalDescribe: dragTotalDesc,
		TotalSize:     dragTotalSize,

		DstFilePath: "",
		Cmd:         rtkCommon.FILE_DROP_REQUEST,
	}

	if len(dragFileInfoList) == 1 && len(dragFolderList) == 0 {
		targetData.FileType = rtkCommon.P2PFile_Type_Single
	} else {
		targetData.FileType = rtkCommon.P2PFile_Type_Multiple
		if len(dragFolderList) > 0 {
			targetData.FolderList = append([]string(nil), dragFolderList...)
		}
	}

	fileDropDataMutex.Lock()
	fileDropDataMap[id] = targetData
	fileDropDataMutex.Unlock()
}

func updateDragFileRespData(id string) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()
	if fileDragData, ok := fileDropDataMap[id]; ok {
		if fileDragData.ActionType != rtkCommon.P2PFileActionType_Drag {
			log.Printf("[%s] Err: Update file drag failed. Invalid ActionType: %s", rtkMisc.GetFuncInfo(), fileDragData.ActionType)
			return
		}
		if fileDragData.Cmd == rtkCommon.FILE_DROP_REQUEST {
			fileDragData.Cmd = rtkCommon.FILE_DROP_ACCEPT
			fileDragData.DstFilePath = rtkPlatform.DownloadPath()
			fileDropDataMap[id] = fileDragData
		} else {
			log.Printf("[%s] Err: Update file drag failed. Invalid cmd: %s", rtkMisc.GetFuncInfo(), fileDragData.Cmd)
		}
	}
}

func UpdateDragFileRespDataFromDst(id string) {
	updateDragFileRespData(id)

	nCount := rtkUtils.GetClientCount()
	for i := 0; i < nCount; i++ {
		fileDropRespIdChan <- id
	}
}

func UpdateDragFileRespDataFromLocal(id string, cmd rtkCommon.FileDropCmd, filePath string) {
	updateDragFileRespData(id)
}

func UpdateDragFileList(fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc string) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	dragFileInfoList = append([]rtkCommon.FileInfo(nil), fileInfoList...)
	dragFileTimeStamp = timeStamp
	dragFolderList = append([]string(nil), folderList...)
	dragTotalDesc = totalDesc
	dragTotalSize = total
}

func UpdateDragFileListFromLocal(fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc string) {
	UpdateDragFileList(fileInfoList, folderList, total, timeStamp, totalDesc)
}

func UpdateDragFileListFromDst(fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc string) {
	UpdateDragFileList(fileInfoList, folderList, total, timeStamp, totalDesc)
}

// ********************  Setup Dst file info ****************

func SetupDstDragFile(id, ip, platform string, fileInfo rtkCommon.FileInfo, fileSize, timestamp uint64) {
	UpdateFileDragInfo(fileInfo, timestamp)
	UpdateDragFileReqDataFromDst(id)
	UpdateDragFileRespDataFromDst(id)

	fileName := rtkMisc.AdaptationPath(filepath.Base(fileInfo.FilePath))
	rtkPlatform.GoDragFileNotify(ip, id, fileName, platform, fileSize, timestamp)
}

func SetupDstDragFileList(id, ip, platform string, fileInfoList []rtkCommon.FileInfo, folderList []string, totalSize, timeStamp uint64, totalDesc string) {
	UpdateDragFileListFromDst(fileInfoList, folderList, totalSize, timeStamp, totalDesc)
	UpdateDragFileReqDataFromDst(id)
	UpdateDragFileRespDataFromDst(id)

	nFileCount := uint32(len(fileInfoList))
	firstFileSize := uint64(0)
	firstFileName := string("")
	if nFileCount > 0 {
		firstFileSize = uint64(fileInfoList[0].FileSize_.SizeHigh)<<32 | uint64(fileInfoList[0].FileSize_.SizeLow)
		firstFileName = rtkMisc.AdaptationPath(fileInfoList[0].FileName)
	} else {
		firstFileName = folderList[0]
	}

	rtkPlatform.GoDragFileListNotify(ip, id, platform, nFileCount, totalSize, timeStamp, firstFileName, firstFileSize)
}
