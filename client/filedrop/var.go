package filedrop

import (
	rtkCommon "rtk-cross-share/client/common"
	"sync"
)

type FilesTransferDirectionType string

const (
	FilesTransfer_As_Src  FilesTransferDirectionType = "FilesTransfer_As_Src"
	FilesTransfer_As_Dst  FilesTransferDirectionType = "FilesTransfer_As_Dst"
	FilesTransfer_Unknown FilesTransferDirectionType = "FilesTransfer_Unknown"
)

type CallbackSendCancelFileTransMsgFunc func(id string, fileTransDataId uint64)

var (
	fileDropDataMap    = make(map[string]FileDropData)           // key: ID
	filesDataCacheMap  = make(map[string]filesDataTransferCache) //Key: ID Value: filesDataTransferCache
	fileDropDataMutex  sync.RWMutex
	fileDropReqIdChan  = make(chan string, 10)
	fileDropRespIdChan = make(chan string, 10)

	callbackSendCancelFileTransferMsgToPeer CallbackSendCancelFileTransMsgFunc

	dragFileInfoList  []rtkCommon.FileInfo
	dragFolderList    []string
	dragFileTimeStamp uint64
	dragTotalSize     uint64
	dragTotalDesc     string
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

type FilesTransferDataItem struct {
	FileDropData
	FileTransDirection FilesTransferDirectionType
}

type filesDataTransferCache struct {
	filesTransferDataQueue []FilesTransferDataItem
	isTransferInProgress   bool
	cancelFn               func()
	isCancelByGui          bool
}

func SetSendFileTransferCancelMsgToPeerCallback(cb CallbackSendCancelFileTransMsgFunc) {
	callbackSendCancelFileTransferMsgToPeer = cb
}

/*func SetRemove() {

}*/
