package filedrop

import (
	rtkCommon "rtk-cross-share/client/common"
	rtkMisc "rtk-cross-share/misc"
	"sync"
)

type FilesTransferDirectionType string

const (
	FilesTransfer_As_Src  FilesTransferDirectionType = "FilesTransfer_As_Src"
	FilesTransfer_As_Dst  FilesTransferDirectionType = "FilesTransfer_As_Dst"
	FilesTransfer_Unknown FilesTransferDirectionType = "FilesTransfer_Unknown"
)

type CallbackSendCancelFileTransMsgFunc func(id, ipAddr string, fileTransDataId uint64, asSrc bool)

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
	dragSrcRootPath   string
)

type FileDropData struct {
	// Req data
	SrcFileList   []rtkCommon.FileInfo
	ActionType    rtkCommon.FileActionType
	TimeStamp     uint64
	FolderList    []string // must start with folder name and end with '/'.  eg: folderName/aaa/bbb/
	SrcRootPath   string   // Src root Folder
	TotalDescribe string   // eg: 820MB /1.2GB
	TotalSize     uint64

	// Resp data
	DstFilePath string //DownloadPath
	Cmd         rtkCommon.FileDropCmd
}

type FileInfoEx struct {
	FileSize uint64
	FilePath string //full path
	FileName string //this must start with folder name, eg: folderName/aaa/bbb/ccc.txt
}

type FileDataTransDetails struct {
	ID            string
	IPAddr        string
	TimeStamp     uint64
	FileList      []FileInfoEx
	FolderList    []string // must start with folder name and end with '/'.  eg: folderName/aaa/bbb/
	RootPath      string   // dst:downloadPath  ; src:rootPath
	TotalSize     uint64
	TotalDescribe string
	FirstFileName string
	FirstFileSize uint64
}

type FilesTransferDataItem struct {
	FileDropData
	FileTransDirection FilesTransferDirectionType

	// Transfer Interrupt Info
	InterruptSrcFileName        string                `json:"-"` // Src fileName
	InterruptDstFileName        string                `json:"-"` // Dst fileName
	InterruptDstFullPath        string                `json:"-"` // Dst fullPath
	InterruptFileOffSet         int64                 `json:"-"`
	InterruptLastErrCode        rtkMisc.CrossShareErr `json:"-"`
	RecoverFileTransTimerCancel func()                `json:"-"`
}

type filesDataTransferCache struct {
	filesTransferDataQueue []FilesTransferDataItem
}

func SetSendFileTransferCancelMsgToPeerCallback(cb CallbackSendCancelFileTransMsgFunc) {
	callbackSendCancelFileTransferMsgToPeer = cb
}
