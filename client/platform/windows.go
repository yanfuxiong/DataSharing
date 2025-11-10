//go:build windows

package platform

import (
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/sys/windows"
	"log"
	"os"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

var (
	privKeyFile              = ".priv.pem"
	hostID                   = ".HostID"
	nodeID                   = ".ID"
	lockFile                 = "singleton.lock"
	lockFd                   *os.File
	logFile                  = "p2p.log"
	crashLogFile             = "crash.log"
	downloadPath             = ""
	chNotifyPasteText        = make(chan struct{}, 100)
	ifConfirmDocumentsAccept bool
)

func InitPlatform(rootPath, downLoadPath, deviceName string) {
	downloadPath = downLoadPath
	if deviceName == "" {
		rtkGlobal.NodeInfo.DeviceName = "UnknownDeviceName"
	} else {
		rtkGlobal.NodeInfo.DeviceName = deviceName
	}

	getPath := func(dirPath, filePath string) string {
		return filepath.Join(dirPath, filePath)
	}

	settingsDir := ".Settings"
	logDir := "Log"
	settingsPath := getPath(rootPath, settingsDir)
	logPath := getPath(rootPath, logDir)

	if !rtkMisc.FolderExists(settingsPath) {
		rtkMisc.CreateDir(settingsPath)
	}

	if !rtkMisc.FolderExists(logPath) {
		rtkMisc.CreateDir(logPath)
	}

	privKeyFile = getPath(settingsPath, privKeyFile)
	hostID = getPath(settingsPath, hostID)
	nodeID = getPath(settingsPath, nodeID)
	lockFile = getPath(settingsPath, lockFile)

	logFile = getPath(logPath, logFile)
	crashLogFile = getPath(logPath, crashLogFile)

	rtkMisc.InitLog(logFile, crashLogFile, 0)
	n, fErr := fmt.Fprintln(os.Stdout, "CheckStatus")
	if fErr == nil && n == 12 {
		rtkMisc.SetupLogConsoleFile()
	} else {
		rtkMisc.SetupLogFile()
	}

	log.Printf("[%s] init rootPath:[%s] downLoadPath:[%s], deviceName:[%s] success!", rtkMisc.GetFuncInfo(), rootPath, downLoadPath, deviceName)
}

type (
	CallbackNetworkSwitchFunc          func()
	CallbackCopyXClipFunc              func(cbText, cbImage, cbHtml []byte)
	CallbackPasteXClipFunc             func(text, image, html string)
	CallbackCleanClipboardFunc         func()
	CallbackFileListDropRequestFunc    func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackDragFileListRequestFunc    func([]rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackDragFileListNotifyFunc     func(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64)
	CallbackMultiFilesDropNotifyFunc   func(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64)
	CallbackMultipleProgressBarFunc    func(ip, id, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64)
	CallbackNotiMessageFileTransFunc   func(fileName, clientName, platform string, timestamp uint64, isSender bool)
	CallbackFileDropResponseFunc       func(string, rtkCommon.FileDropCmd, string)
	CallbackCancelFileTransFunc        func(string, string, uint64)
	CallbackExtractDIASFunc            func()
	CallbackGetMacAddressFunc          func(string)
	CallbackAuthStatusCodeFunc         func(uint8)
	CallbackDIASSourceAndPortFunc      func(uint8, uint8)
	CallbackMethodBrowseMdnsResultFunc func(string, string, int, string, string, string, string)
	CallbackAuthViaIndexCallbackFunc   func(uint32)
	CallbackDIASStatusFunc             func(uint32)
	CallbackReqSourceAndPortFunc       func()
	CallbackUpdateSystemInfoFunc       func(ipAddr string, verInfo string)
	CallbackUpdateClientStatusFunc     func(status uint32, ip, id, deviceName, deviceType string)
	CallbackUpdateClientStatusExFunc   func(clientInfo string)
	CallbackDetectPluginEventFunc      func(isPlugin bool, productName string)
	CallbackGetFilesTransCodeFunc      func(id string) rtkCommon.SendFilesRequestErrCode
	CallbackGetFilesCacheSendCountFunc func(id string) int
	CallbackReqClientUpdateVerFunc     func(clientVer string)
	CallbackNotifyErrEventFunc         func(id string, errCode uint32, arg1, arg2, arg3, arg4 string)
	CallbackConnectLanServerFunc       func(instance string)
	CallbackBrowseLanServerFunc        func()
	CallbackSetMsgEventFunc            func(event uint32, arg1, arg2, arg3, arg4 string)
)

var (
	// Go business Callback
	callbackNetworkSwitchCB            CallbackNetworkSwitchFunc          = nil
	callbackCopyXClipDataCB            CallbackCopyXClipFunc              = nil
	callbackPasteXClipDataCB           CallbackPasteXClipFunc             = nil
	callbackFileListDropRequestCB      CallbackFileListDropRequestFunc    = nil
	callbackDragFileListRequestCB      CallbackDragFileListRequestFunc    = nil
	callbackDragFileListNotifyCB       CallbackDragFileListNotifyFunc     = nil
	callbackMultiFilesDropNotifyCB     CallbackMultiFilesDropNotifyFunc   = nil
	callbackMultipleProgressBarCB      CallbackMultipleProgressBarFunc    = nil
	callbackNotiMessageFileTransCB     CallbackNotiMessageFileTransFunc   = nil
	callbackInstanceFileDropResponseCB CallbackFileDropResponseFunc       = nil
	callbackCancelFileTransDragCB      CallbackCancelFileTransFunc        = nil
	callbackExtractDIASCB              CallbackExtractDIASFunc            = nil
	callbackGetMacAddressCB            CallbackGetMacAddressFunc          = nil
	callbackAuthStatusCodeCB           CallbackAuthStatusCodeFunc         = nil
	callbackDIASSourceAndPortCB        CallbackDIASSourceAndPortFunc      = nil
	callbackMethodBrowseMdnsResult     CallbackMethodBrowseMdnsResultFunc = nil
	callbackReqClientUpdateVer         CallbackReqClientUpdateVerFunc     = nil
	callbackNotifyErrEvent             CallbackNotifyErrEventFunc         = nil
	callbackSetMsgEvent                CallbackSetMsgEventFunc            = nil

	// main.go Callback
	callbackAuthViaIndex           CallbackAuthViaIndexCallbackFunc   = nil
	callbackDIASStatus             CallbackDIASStatusFunc             = nil
	callbackReqSourceAndPort       CallbackReqSourceAndPortFunc       = nil
	callbackUpdateSystemInfo       CallbackUpdateSystemInfoFunc       = nil
	callbackUpdateClientStatus     CallbackUpdateClientStatusFunc     = nil
	callbackUpdateClientStatusEx   CallbackUpdateClientStatusExFunc   = nil
	callbackCleanClipboard         CallbackCleanClipboardFunc         = nil
	callbackGetFilesTransCode      CallbackGetFilesTransCodeFunc      = nil
	callbackGetFilesCacheSendCount CallbackGetFilesCacheSendCountFunc = nil
)

/*======================================= Used  by GO set Callback =======================================*/

func SetGoNetworkSwitchCallback(cb CallbackNetworkSwitchFunc) {
	callbackNetworkSwitchCB = cb
}

func SetCopyXClipCallback(cb CallbackCopyXClipFunc) {
	callbackCopyXClipDataCB = cb
}

func SetGoFileListDropRequestCallback(cb CallbackFileListDropRequestFunc) {
	callbackFileListDropRequestCB = cb
}

func SetGoDragFileListRequestCallback(cb CallbackDragFileListRequestFunc) {
	callbackDragFileListRequestCB = cb
}

func SetGoFileDropResponseCallback(cb CallbackFileDropResponseFunc) {
	callbackInstanceFileDropResponseCB = cb
}

func SetGoCancelFileTransCallback(cb CallbackCancelFileTransFunc) {
	callbackCancelFileTransDragCB = cb
}

func SetGoExtractDIASCallback(cb CallbackExtractDIASFunc) {
	callbackExtractDIASCB = cb
}

func SetGoGetMacAddressCallback(cb CallbackGetMacAddressFunc) {
	callbackGetMacAddressCB = cb
}

func SetDetectPluginEventCallback(cb CallbackDetectPluginEventFunc) {
	//callbackDetectPluginEventCB = cb
}

func SetGoAuthStatusCodeCallback(cb CallbackAuthStatusCodeFunc) {
	callbackAuthStatusCodeCB = cb
}

func SetGoDIASSourceAndPortCallback(cb CallbackDIASSourceAndPortFunc) {
	callbackDIASSourceAndPortCB = cb
}

func SetGoBrowseMdnsResultCallback(cb CallbackMethodBrowseMdnsResultFunc) {
	callbackMethodBrowseMdnsResult = cb
}

func SetGetFilesTransCodeCallback(cb CallbackGetFilesTransCodeFunc) {
	callbackGetFilesTransCode = cb
}

func SetGetFilesCacheSendCountCallback(cb CallbackGetFilesCacheSendCountFunc) {
	callbackGetFilesCacheSendCount = cb
}

func SetGoConnectLanServerCallback(cb CallbackConnectLanServerFunc) {
}

func SetGoBrowseLanServerCallback(cb CallbackBrowseLanServerFunc) {

}

func SetGoSetMsgEventCallback(cb CallbackSetMsgEventFunc) {
	callbackSetMsgEvent = cb
}

/*======================================= Used by main.go, set Callback =======================================*/

func SetPasteXClipCallback(cb CallbackPasteXClipFunc) {
	callbackPasteXClipDataCB = cb
}

func SetAuthViaIndexCallback(cb CallbackAuthViaIndexCallbackFunc) {
	callbackAuthViaIndex = cb
}

func SetDIASStatusCallback(cb CallbackDIASStatusFunc) {
	callbackDIASStatus = cb
}

func SetRequestSourceAndPortCallback(cb CallbackReqSourceAndPortFunc) {
	callbackReqSourceAndPort = cb
}

func SetUpdateSystemInfoCallback(cb CallbackUpdateSystemInfoFunc) {
	callbackUpdateSystemInfo = cb
}

func SetUpdateClientStatusExCallback(cb CallbackUpdateClientStatusExFunc) {
	callbackUpdateClientStatusEx = cb
}

func SetUpdateClientStatusCallback(cb CallbackUpdateClientStatusFunc) {
	callbackUpdateClientStatus = cb
}

func SetCleanClipboardCallback(cb CallbackCleanClipboardFunc) {
	callbackCleanClipboard = cb
}

func SetMultipleProgressBarCallback(cb CallbackMultipleProgressBarFunc) {
	callbackMultipleProgressBarCB = cb
}

func SetDragFileListNotifyCallback(cb CallbackDragFileListNotifyFunc) {
	callbackDragFileListNotifyCB = cb
}

func SetMultiFilesDropNotifyCallback(cb CallbackMultiFilesDropNotifyFunc) {
	callbackMultiFilesDropNotifyCB = cb
}

func SetNotiMessageFileTransCallback(cb CallbackNotiMessageFileTransFunc) {
	callbackNotiMessageFileTransCB = cb
}

func SetReqClientUpdateVerCallback(cb CallbackReqClientUpdateVerFunc) {
	callbackReqClientUpdateVer = cb
}

func SetNotifyErrEventCallback(cb CallbackNotifyErrEventFunc) {
	callbackNotifyErrEvent = cb
}

/*======================================= Used by main.go, Called by C++ =======================================*/
func GoSetMsgEventFunc(event uint32, arg1, arg2, arg3, arg4 string) {
	if callbackSetMsgEvent == nil {
		log.Println("callbackSetMsgEvent is null!")
		return
	}
	callbackSetMsgEvent(event, arg1, arg2, arg3, arg4)
}

func GoCopyXClipData(text, image, html []byte) {
	if callbackCopyXClipDataCB == nil {
		log.Println("callbackCopyXClipDataCB is null!")
		return
	}

	callbackCopyXClipDataCB(text, image, html)
}

func GoSetAuthStatusCode(status uint8) {
	callbackAuthStatusCodeCB(status)
}

func GoSetDIASSourceAndPort(src, port uint8) {
	callbackDIASSourceAndPortCB(src, port)
}

func GoExtractDIASCallback() {
	callbackExtractDIASCB()
}

func GoSetMacAddress(macAddr string) {
	callbackGetMacAddressCB(macAddr)
}

func GoMultiFilesDropRequest(id string, fileList *[]rtkCommon.FileInfo, folderList *[]string, totalSize, timestamp uint64, totalDesc string) rtkCommon.SendFilesRequestErrCode {
	if callbackFileListDropRequestCB == nil {
		log.Println("callbackFileListDropRequestCB is null!")
		return rtkCommon.SendFilesRequestCallbackNotSet
	}

	if callbackGetFilesCacheSendCount == nil {
		log.Println("callbackGetFilesCacheSendCount is null!")
		return rtkCommon.SendFilesRequestCallbackNotSet
	}

	if len(*fileList) == 0 && len(*folderList) == 0 {
		log.Println("file content is null!")
		return rtkCommon.SendFilesRequestParameterErr
	}

	if !rtkUtils.GetPeerClientIsSupportQueueTrans(id) {
		if callbackGetFilesTransCode == nil {
			log.Println("callbackGetFilesTransCode is null!")
			return rtkCommon.SendFilesRequestCallbackNotSet
		}

		filesTransCode := callbackGetFilesTransCode(id)
		if filesTransCode != rtkCommon.SendFilesRequestSuccess {
			return filesTransCode
		}
	}

	nCacheCount := callbackGetFilesCacheSendCount(id)
	if nCacheCount >= rtkGlobal.SendFilesRequestMaxQueueSize {
		log.Printf("[%s] ID[%s] this user file drop cache count:[%d] is too large and over range !", rtkMisc.GetFuncInfo(), id, nCacheCount)
		return rtkCommon.SendFilesRequestCacheOverRange
	}

	if totalSize > uint64(rtkGlobal.SendFilesRequestMaxSize) {
		log.Printf("[%s] ID[%s] this file drop total size:[%d] [%s] is too large and over range !", rtkMisc.GetFuncInfo(), id, totalSize, totalDesc)
		return rtkCommon.SendFilesRequestSizeOverRange
	}

	nMsgLength := int(rtkGlobal.P2PMsgMagicLength) //p2p null msg length

	for _, file := range *fileList {
		nMsgLength = nMsgLength + len(file.FileName) + rtkGlobal.FileInfoMagicLength
	}

	for _, folder := range *folderList {
		nMsgLength = nMsgLength + len(folder) + rtkGlobal.StringArrayMagicLength
	}

	if nMsgLength >= rtkGlobal.P2PMsgMaxLength {
		log.Printf("[%s] ID[%s] file count:[%d] folder count:[%d], the p2p message is too long and over range!", rtkMisc.GetFuncInfo(), id, len(*fileList), len(*folderList))
		return rtkCommon.SendFilesRequestLengthOverRange
	}

	callbackFileListDropRequestCB(id, *fileList, *folderList, totalSize, timestamp, totalDesc)
	return rtkCommon.SendFilesRequestSuccess
}

func GoDragFileListRequest(fileList *[]rtkCommon.FileInfo, folderList *[]string, totalSize, timestamp uint64, totalDesc string) {
	if len(*fileList) == 0 && len(*folderList) == 0 {
		log.Println("file content is null!")
		return
	}
	callbackDragFileListRequestCB(*fileList, *folderList, totalSize, timestamp, totalDesc)
}

func GoCancelFileTrans(ip, id string, timestamp int64) {
	callbackCancelFileTransDragCB(id, ip, uint64(timestamp))
}

func GoUpdateDownloadPath(path string) {
	downloadPath = path
}

/*======================================= Used by GO business =======================================*/

func GetDownloadPath() string {
	return downloadPath
}

func GetLogFilePath() string {
	return logFile
}

func GetCrashLogFilePath() string {
	return crashLogFile
}

func GoAuthViaIndex(clientIndex uint32) {
	callbackAuthViaIndex(clientIndex)
}

func GoReqSourceAndPort() {
	callbackReqSourceAndPort()
}

func GoMonitorNameNotify(name string) {
}

func GoDIASStatusNotify(diasStatus uint32) {
	callbackDIASStatus(diasStatus)
}

func GetAuthData(clientIndex uint32) (rtkMisc.CrossShareErr, rtkMisc.AuthDataInfo) {
	return rtkMisc.ERR_BIZ_GET_CALLBACK_INSTANCE_NULL, rtkMisc.AuthDataInfo{}
}

func GoSetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {

}

func GoSetupFileListDrop(ip, id, platform, totalDesc string, fileCount, folderCount uint32, timestamp uint64) {
	log.Printf("[%s] fileCnt:[%d] folderCnt:[%d] totalDesc:[%s]", rtkMisc.GetFuncInfo(), fileCount, folderCount, totalDesc)
}

func GoMultiFilesDropNotify(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64) {
	callbackMultiFilesDropNotifyCB(ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
}

func GoDragFileListNotify(ip, id, platform string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64) {
	callbackDragFileListNotifyCB(ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
}

func GoDragFileListFolderNotify(ip, id, folderName string, timestamp uint64) {
}

func GoSetupDstPasteXClipData(cbText, cbImage, cbHtml []byte) {
	if callbackPasteXClipDataCB == nil {
		log.Printf("callbackPasteXClipDataCB is null!\n\n")
		return
	}

	imageStr := ""
	if len(cbImage) > 0 {
		imageBase64 := rtkUtils.Base64Encode(cbImage)
		imageStr = imageBase64
	}

	callbackPasteXClipDataCB(string(cbText), imageStr, string(cbHtml))
}

func GoUpdateMultipleProgressBar(ip, id, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	callbackMultipleProgressBarCB(ip, id, currentFileName, sentFileCnt, totalFileCnt, currentFileSize, totalSize, sentSize, timestamp)
}

func GoUpdateSystemInfo(ipAddr, serviceVer string) {
	log.Printf("[%s] Ip:[%s]  version[%s]", rtkMisc.GetFuncInfo(), ipAddr, serviceVer)
	callbackUpdateSystemInfo(ipAddr, serviceVer)
}

func GoUpdateClientStatusEx(id string, status uint8) {
	if callbackUpdateClientStatusEx == nil {
		log.Printf("callbackUpdateClientStatusEx is null!\n\n")
		return
	}

	var clientInfo rtkCommon.ClientStatusInfo
	if status == 1 { // online
		info, err := rtkUtils.GetClientInfo(id)
		if err != nil {
			log.Printf("[%s] err:%+v", rtkMisc.GetFuncInfo(), err)
			return
		}
		clientInfo.ClientInfo = info.ClientInfo
	} else {
		clientInfo.ID = id
	}

	clientInfo.TimeStamp = time.Now().UnixMilli()
	clientInfo.Status = status
	encodedData, err := json.Marshal(clientInfo)
	if err != nil {
		log.Println("Failed to Marshal ClientStatusInfo data, err:", err)
		return
	}

	callbackUpdateClientStatusEx(string(encodedData))
}

func GoUpdateClientStatus(status uint32, ip, id, name, deviceType string) {
	callbackUpdateClientStatus(status, ip, id, name, deviceType)
}

func GoNotiMessageFileTransfer(fileName, clientName, platform string, timestamp uint64, isSender bool) {
	callbackNotiMessageFileTransCB(fileName, clientName, platform, timestamp, isSender)
}

func GoNotifyErrEvent(id string, errCode rtkMisc.CrossShareErr, arg1, arg2, arg3, arg4 string) {
	if callbackNotifyErrEvent == nil {
		log.Printf("callbackNotifyErrEvent is null!\n")
		return
	}

	callbackNotifyErrEvent(id, uint32(errCode), arg1, arg2, arg3, arg4)
}

func GoRequestUpdateClientVersion(ver string) {
	if callbackReqClientUpdateVer == nil {
		log.Printf("callbackReqClientUpdateVer is null!\n")
		return
	}

	callbackReqClientUpdateVer(ver)
}

func GoCleanClipboard() {
	callbackCleanClipboard()
}

func FoundPeer() {
}

func GenKey() crypto.PrivKey {
	return rtkUtils.GenKey(privKeyFile)
}

func IsHost() bool {
	return rtkMisc.FileExists(hostID)
}

func GetHostID() string {
	file, err := os.Open(hostID)
	if err != nil {
		log.Println(err)
		return rtkGlobal.HOST_ID
	}
	defer file.Close()

	data := make([]byte, 1024)
	_, err = file.Read(data)
	if err != nil {
		log.Println(err)
		return rtkGlobal.HOST_ID
	}

	return string(data)
}

func GetIDPath() string {
	return nodeID
}

func GetHostIDPath() string {
	return hostID
}

func GetPlatform() string {
	return rtkMisc.PlatformWindows
}

// Deprecated: unused
func SetNetWorkConnected(bConnected bool) {
}

// Deprecated: unused
func GetNetWorkConnected() bool {
	return false
}

func LockFile() (err error) {
	lockFd, err = os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Failed to open or create lock file:[%s] err:%+v", lockFile, err)
		return
	}

	handle := windows.Handle(lockFd.Fd())
	if handle == windows.InvalidHandle {
		err = fmt.Errorf("invalid file handle")
		log.Printf("[%s] Failed to get file handle", rtkMisc.GetFuncInfo())
		return
	}

	var overlapped windows.Overlapped
	err = windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
	if err != nil {
		log.Printf("Failed to lock file[%s] err:%+v", lockFile, err)
	}

	return
}

func UnlockFile() error {
	handle := windows.Handle(lockFd.Fd())
	if handle == windows.InvalidHandle {
		return fmt.Errorf("invalid file handle")
	}

	var overlapped windows.Overlapped

	err := windows.UnlockFileEx(handle, 0, 1, 0, &overlapped)
	if err != nil {
		return fmt.Errorf("failed to unlock file: %w", err)
	}

	lockFd.Close()
	return nil
}

func GoTriggerNetworkSwitch() {
	callbackNetworkSwitchCB()
}

func GoGetSrcAndPortFromIni() rtkMisc.SourcePort {
	return rtkUtils.GetDeviceSrcPort()
}

func SetConfirmDocumentsAccept(ifConfirm bool) {
	ifConfirmDocumentsAccept = ifConfirm
}

func GetConfirmDocumentsAccept() bool {
	return ifConfirmDocumentsAccept
}

func GoNotifyBrowseResult(monitorName, instance, ipAddr, version string, timestamp int64) {

}

// Specific Platform: iOS. Browse and lookup MDNS from iOS
func GoStartBrowseMdns(instance, serviceType string) {
}

func GoStopBrowseMdns() {
}

func GoSetupAppLink(link string) {
	rtkMisc.AppLink = link
}
