package filedrop

import (
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
)

func init() {
	rtkPlatform.SetGoDragFileListRequestCallback(UpdateDragFileListFromLocal)
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

	nCacheCount := GetFilesTransferDataSendCacheCount(id)
	if nCacheCount >= rtkGlobal.SendFilesRequestMaxQueueSize {
		log.Printf("[%s] ID[%s] this user file drop cache count:[%d] is too large and over range !", rtkMisc.GetFuncInfo(), id, nCacheCount)
		return rtkMisc.ERR_BIZ_DF_CACHE_OVER_RANGE
	}

	updateDragFileReqData(id)

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
		return rtkMisc.ERR_BIZ_GET_CLIENT_INFO_EMPTY
	}

	firstFileSize := uint64(0)
	var firstFileName string
	if fileCnt > 0 {
		firstFileSize = uint64(dragFileInfoList[0].FileSize_.SizeHigh)<<32 | uint64(dragFileInfoList[0].FileSize_.SizeLow)
		firstFileName = dragFileInfoList[0].FileName
	} else {
		firstFileName = dragFolderList[0]
	}

	rtkPlatform.GoFileListSendNotify(ipAddr, id, uint32(fileCnt), dragTotalSize, dragFileTimeStamp, firstFileName, firstFileSize, getFileDropDataDetails(id, ipAddr))
	return rtkMisc.SUCCESS
}

func UpdateDragFileReqDataFromDst(id string) {
	updateDragFileReqData(id)
}

func updateDragFileReqData(id string) {
	targetData := FileDropData{
		SrcFileList:          dragFileInfoList,
		ActionType:           rtkCommon.P2PFileActionType_Drag,
		TimeStamp:            dragFileTimeStamp,
		FolderList:           dragFolderList,
		SrcRootPath:   	      dragSrcRootPath,
		TotalDescribe:        dragTotalDesc,
		TotalSize:            dragTotalSize,
		DstFilePath:          "",
		Cmd:                  rtkCommon.FILE_DROP_REQUEST,
		InterruptSrcFileName: "",
		InterruptDstFileName: "",
		InterruptFileOffSet:  0,
		InterruptLastErrCode: rtkMisc.SUCCESS,
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
			fileDragData.DstFilePath = rtkPlatform.GetDownloadPath()
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

func UpdateDragFileList(fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc, srcRootPath string) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	dragFileInfoList = fileInfoList
	dragFileTimeStamp = timeStamp
	dragFolderList = folderList
	dragTotalDesc = totalDesc
	dragSrcRootPath = srcRootPath
	dragTotalSize = total
}

func UpdateDragFileListFromLocal(fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc, srcRootPath string) {
	UpdateDragFileList(fileInfoList, folderList, total, timeStamp, totalDesc, srcRootPath)
}

func UpdateDragFileListFromDst(fileInfoList []rtkCommon.FileInfo, folderList []string, total, timeStamp uint64, totalDesc string) {
	UpdateDragFileList(fileInfoList, folderList, total, timeStamp, totalDesc, "")
}

// ********************  Setup Dst file info ****************

func SetupDstDragFileList(id, ip string, fileInfoList []rtkCommon.FileInfo, folderList []string, totalSize, timeStamp uint64, totalDesc string) {
	if len(folderList) > 0 {
		fileInfoList, folderList = rtkUtils.GetTargetFileList(rtkPlatform.GetDownloadPath(), fileInfoList, folderList)
	}
	UpdateDragFileListFromDst(fileInfoList, folderList, totalSize, timeStamp, totalDesc)
	UpdateDragFileReqDataFromDst(id)

	nFileCount := uint32(len(fileInfoList))
	firstFileSize := uint64(0)
	firstFileName := string("")
	if nFileCount > 0 {
		firstFileSize = uint64(fileInfoList[0].FileSize_.SizeHigh)<<32 | uint64(fileInfoList[0].FileSize_.SizeLow)
		firstFileName = rtkMisc.AdaptationPath(fileInfoList[0].FileName)
	} else {
		firstFileName = folderList[0]
	}

	UpdateDragFileRespDataFromDst(id)
	rtkPlatform.GoFileListReceiveNotify(ip, id, nFileCount, totalSize, timeStamp, firstFileName, firstFileSize, getFileDropDataDetails(id, ip))
}
