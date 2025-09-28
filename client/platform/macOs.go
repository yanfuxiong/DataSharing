//go:build darwin && !ios

package platform

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
)

var (
	isNetWorkConnected       bool // Deprecated: unused
	ifConfirmDocumentsAccept bool
	privKeyFile              string
	hostID                   string
	nodeID                   string
	lockFile                 string
	logFile                  string
	crashLogFile             string
	downloadPath             string
	rootPath                 string
)

func initFilePath() {
	privKeyFile = ".priv.pem"
	hostID = ".HostID"
	nodeID = ".ID"
	logFile = "p2p.log"
	lockFile = "singleton.lock"
	crashLogFile = "crash.log"
	downloadPath = ""
}

func GetRootPath() string {
	return rootPath
}

func GetDownloadPath() string {
	return downloadPath
}

func GetLogFilePath() string {
	return logFile
}

func GetCrashLogFilePath() string {
	return crashLogFile
}

func GetLockFilePath() string {
	return lockFile
}

type (
	CallbackNetworkSwitchFunc              func()
	CallbackCopyXClipFunc                  func(cbText, cbImage, cbHtml []byte)
	CallbackPasteXClipFunc                 func(text, image, html string)
	CallbackFileDropResponseFunc           func(string, rtkCommon.FileDropCmd, string)
	CallbackDragFileListRequestFunc        func([]rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackFileListDragNotify             func(string, string, string, uint32, uint64, uint64, string, uint64)
	CallbackFileListDragFolderNotify       func(string, string, string, uint64)
	CallbackFileListDropRequestFunc        func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackUpdateClientStatusFunc         func(clientInfo string)
	CallbackUpdateMultipleProgressBar      func(string, string, string, uint32, uint32, uint64, uint64, uint64, uint64)
	CallbackCancelFileTransFunc            func(string, string, uint64)
	CallbackNotifyErrEventFunc             func(id string, errCode uint32, arg1, arg2, arg3, arg4 string)
	CallbackGetMacAddressFunc              func(string)
	CallbackAuthStatusCodeFunc             func(uint8)
	CallbackExtractDIASFunc                func()
	CallbackDIASSourceAndPortFunc          func(uint8, uint8)
	CallbackMethodStartBrowseMdns          func(string, string)
	CallbackMethodStopBrowseMdns           func()
	CallbackMethodBrowseMdnsResultFunc     func(string, string, int, string, string, string, string)
	CallbackDetectPluginEventFunc          func(isPlugin bool, productName string)
	CallbackReqSourceAndPortFunc           func()
	CallbackDIASStatusFunc                 func(uint32)
	CallbackAuthViaIndexFunc               func(uint32)
	CallbackMonitorNameFunc                func(string)
	CallbackGetFilesTransCodeFunc          func(id string) rtkCommon.SendFilesRequestErrCode
	CallbackRequestUpdateClientVersionFunc func(string)
	CallbackNotifyBrowseResultFunc         func(monitorName, instance, ipAddr, version string, timestamp int64)
	CallbackConnectLanServerFunc           func(instance string)
	CallbackBrowseLanServerFunc            func()
	CallbackSetMsgEventFunc                func(event uint32, arg1, arg2, arg3, arg4 string)
)

var (
	callbackNetworkSwitch              CallbackNetworkSwitchFunc              = nil
	callbackCopyXClipDataCB            CallbackCopyXClipFunc                  = nil
	callbackPasteXClipDataCB           CallbackPasteXClipFunc                 = nil
	callbackInstanceFileDropResponseCB CallbackFileDropResponseFunc           = nil
	callbackDragFileListRequestCB      CallbackDragFileListRequestFunc        = nil
	callbackFileListDragNotify         CallbackFileListDragNotify             = nil
	callbackFileListDragFolderNotify   CallbackFileListDragFolderNotify       = nil
	callbackFileListDropRequest        CallbackFileListDropRequestFunc        = nil
	callbackUpdateClientStatus         CallbackUpdateClientStatusFunc         = nil
	callbackUpdateMultipleProgressBar  CallbackUpdateMultipleProgressBar      = nil
	callbackCancelFileTrans            CallbackCancelFileTransFunc            = nil
	callbackNotifyErrEvent             CallbackNotifyErrEventFunc             = nil
	callbackGetMacAddress              CallbackGetMacAddressFunc              = nil
	callbackAuthStatusCodeCB           CallbackAuthStatusCodeFunc             = nil
	callbackExtractDIAS                CallbackExtractDIASFunc                = nil
	callbackAuthViaIndex               CallbackAuthViaIndexFunc               = nil
	callbackDIASSourceAndPortCB        CallbackDIASSourceAndPortFunc          = nil
	callbackMethodStartBrowseMdns      CallbackMethodStartBrowseMdns          = nil
	callbackMethodStopBrowseMdns       CallbackMethodStopBrowseMdns           = nil
	callbackMethodBrowseMdnsResult     CallbackMethodBrowseMdnsResultFunc     = nil
	callbackDetectPluginEvent          CallbackDetectPluginEventFunc          = nil
	callbackReqSourceAndPort           CallbackReqSourceAndPortFunc           = nil
	callbackDIASStatus                 CallbackDIASStatusFunc                 = nil
	callbackMonitorName                CallbackMonitorNameFunc                = nil
	callbackGetFilesTransCode          CallbackGetFilesTransCodeFunc          = nil
	callbackRequestUpdateClientVersion CallbackRequestUpdateClientVersionFunc = nil
	callbackSetMsgEvent                CallbackSetMsgEventFunc                = nil
)

/*======================================= Used by main.go, set Callback =======================================*/

func SetCallbackPasteXClip(cb CallbackPasteXClipFunc) {
	callbackPasteXClipDataCB = cb
}

func SetCallbackFileListNotify(cb CallbackFileListDragNotify) {
	callbackFileListDragNotify = cb
}

func SetCallbackFileListFolderNotify(cb CallbackFileListDragFolderNotify) {
	callbackFileListDragFolderNotify = cb
}

func SetCallbackUpdateClientStatus(cb CallbackUpdateClientStatusFunc) {
	callbackUpdateClientStatus = cb
}

func SetCallbackUpdateMultipleProgressBar(cb CallbackUpdateMultipleProgressBar) {
	callbackUpdateMultipleProgressBar = cb
}

func SetCallbackMethodStartBrowseMdns(cb CallbackMethodStartBrowseMdns) {
	callbackMethodStartBrowseMdns = cb
}

func SetCallbackMethodStopBrowseMdns(cb CallbackMethodStopBrowseMdns) {
	callbackMethodStopBrowseMdns = cb
}

func SetCallbackDIASStatus(cb CallbackDIASStatusFunc) {
	callbackDIASStatus = cb
}

func SetCallbackAuthViaIndex(cb CallbackAuthViaIndexFunc) {
	callbackAuthViaIndex = cb
}

func SetCallbackRequestSourceAndPort(cb CallbackReqSourceAndPortFunc) {
	callbackReqSourceAndPort = cb
}

func SetCallbackMonitorName(cb CallbackMonitorNameFunc) {
	callbackMonitorName = cb
}

func SetCallbackRequestUpdateClientVersion(cb CallbackRequestUpdateClientVersionFunc) {
	callbackRequestUpdateClientVersion = cb
}

func SetCallbackNotifyErrEvent(cb CallbackNotifyErrEventFunc) {
	callbackNotifyErrEvent = cb
}

/*======================================= Used  by GO set Callback =======================================*/

func SetGoNetworkSwitchCallback(cb CallbackNetworkSwitchFunc) {
	callbackNetworkSwitch = cb
}

// Notify to Clipboard/FileDrop
func SetCopyXClipCallback(cb CallbackCopyXClipFunc) {
	callbackCopyXClipDataCB = cb
}

func SetGoFileDropResponseCallback(cb CallbackFileDropResponseFunc) {
	callbackInstanceFileDropResponseCB = cb
}

func SetGoFileListDropRequestCallback(cb CallbackFileListDropRequestFunc) {
	callbackFileListDropRequest = cb
}

func SetGoDragFileListRequestCallback(cb CallbackDragFileListRequestFunc) {
	callbackDragFileListRequestCB = cb
}

func SetGoCancelFileTransCallback(cb CallbackCancelFileTransFunc) {
	callbackCancelFileTrans = cb
}

func SetGoExtractDIASCallback(cb CallbackExtractDIASFunc) {
	callbackExtractDIAS = cb
}

func SetGoGetMacAddressCallback(cb CallbackGetMacAddressFunc) {
	callbackGetMacAddress = cb
}

func SetDetectPluginEventCallback(cb CallbackDetectPluginEventFunc) {
	callbackDetectPluginEvent = cb
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

func SetGoConnectLanServerCallback(cb CallbackConnectLanServerFunc) {
}

func SetGoBrowseLanServerCallback(cb CallbackBrowseLanServerFunc) {
}

func SetGoSetMsgEventCallback(cb CallbackSetMsgEventFunc) {
	callbackSetMsgEvent = cb
}

/*======================================= Used  by ios API =======================================*/

func SetupRootPath(path string) {
	if path == "" {
		return
	}
	rootPath = path
	initFilePath()

	getPath := func(dirPath, filePath string) string {
		return dirPath + "/" + filePath
	}

	settingsDir := ".Settings" // TODO: Be hidden folder in the future
	logDir := "Log"            // TODO: Be hidden folder in the future
	downloadDir := "Download"

	settingsPath := getPath(rootPath, settingsDir)
	logPath := getPath(rootPath, logDir)
	downloadPath = getPath(rootPath, downloadDir)

	rtkMisc.CreateDir(settingsPath)
	rtkMisc.CreateDir(logPath)
	rtkMisc.CreateDir(downloadPath)

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
}

func GoSetMsgEventFunc(event uint32, arg1, arg2, arg3, arg4 string) {
	if callbackSetMsgEvent == nil {
		log.Println("callbackSetMsgEvent is null!")
		return
	}
	rtkMisc.GoSafe(func() { callbackSetMsgEvent(event, arg1, arg2, arg3, arg4) })
}

func SetDeviceName(name string) {
	rtkGlobal.NodeInfo.DeviceName = name
}

func GoTriggerNetworkSwitch() {
	callbackNetworkSwitch()
}

func GoBrowseMdnsResultCallback(instance, ip string, port int, productName, mName, timestamp, version string) {
	if callbackMethodBrowseMdnsResult == nil {
		log.Println("CallbackMethodBrowseMdnsResult is null!")
		return
	}

	log.Printf("[%s] instance:[%s], ip:[%s], port:[%d], productName:[%s], mName:[%s], timestamp:[%s], verion:[%s]",
		rtkMisc.GetFuncInfo(), instance, ip, port, productName, mName, timestamp, version)
	callbackMethodBrowseMdnsResult(instance, ip, port, productName, mName, timestamp, version)
}

func GoSetMacAddress(mac string) {
	callbackGetMacAddress(mac)
}

func GoExtractDIASCallback() {
	callbackExtractDIAS()
}

func SetConfirmDocumentsAccept(ifConfirm bool) {
	ifConfirmDocumentsAccept = ifConfirm
}

func GoSetAuthStatusCode(status uint8) {
	if callbackAuthStatusCodeCB == nil {
		log.Println("callbackAuthStatusCodeCB is null!")
		return
	}
	callbackAuthStatusCodeCB(status)
}

func GoSetDIASSourceAndPort(src, port uint8) {
	if callbackDIASSourceAndPortCB == nil {
		log.Println("callbackDIASSourceAndPortCB is null!")
		return
	}
	callbackDIASSourceAndPortCB(src, port)
}

func GoCopyXClipData(text, image, html []byte) {
	if callbackCopyXClipDataCB == nil {
		log.Println("callbackCopyXClipDataCB is null!")
		return
	}

	callbackCopyXClipDataCB(text, image, html)
}

func GoFileDropResponse(id string, fileCmd rtkCommon.FileDropCmd, fileName string) {
	callbackInstanceFileDropResponseCB(id, fileCmd, fileName)
}

func GoMultiFilesDropRequest(id string, fileList *[]rtkCommon.FileInfo, folderList *[]string, totalSize, timestamp uint64, totalDesc string) rtkCommon.SendFilesRequestErrCode {
	if callbackFileListDropRequest == nil {
		log.Println("CallbackFileListDropRequest is null!")
		return rtkCommon.SendFilesRequestCallbackNotSet
	}

	if callbackGetFilesTransCode == nil {
		log.Println("callbackGetFilesTransCode is null!")
		return rtkCommon.SendFilesRequestCallbackNotSet
	}

	filesTransCode := callbackGetFilesTransCode(id)
	if filesTransCode != rtkCommon.SendFilesRequestSuccess {
		return filesTransCode
	}

	nMsgLength := int(rtkGlobal.P2PMsgMagicLength) //p2p null msg length

	for _, file := range *fileList {
		nMsgLength = nMsgLength + len(file.FileName) + rtkGlobal.FileInfoMagicLength
	}

	for _, folder := range *folderList {
		nMsgLength = nMsgLength + len(folder) + rtkGlobal.StringArrayMagicLength
	}

	if nMsgLength >= rtkGlobal.P2PMsgMaxLength {
		log.Printf("[%s] ID[%s] get file count:[%d] folder count:[%d], the p2p message is too long and over range!", rtkMisc.GetFuncInfo(), id, len(*fileList), len(*folderList))
		return rtkCommon.SendFilesRequestOverRange
	}

	callbackFileListDropRequest(id, *fileList, *folderList, totalSize, timestamp, totalDesc)
	return filesTransCode
}

func GoCancelFileTrans(ip, id string, timestamp uint64) {
	if callbackCancelFileTrans == nil {
		log.Println("callbackCancelFileTrans is null!")
		return
	}
	callbackCancelFileTrans(id, ip, timestamp)
}

/*======================================= Used  by GO business =======================================*/

func GoSetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {

}

func GoSetupFileListDrop(ip, id, platform, totalDesc string, fileCount, folderCount uint32, timestamp uint64) {
	log.Printf("(DST) GoSetupFileListDrop  ID:]%s] IP:[%s] totalDesc:%s  fileCount:%d  folderCount:%d", id, ip, totalDesc, fileCount, folderCount)
}

func GoMultiFilesDropNotify(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64) {
	if callbackFileListDragNotify == nil {
		log.Println(" callbackFileListDragNotify is null !")
		return
	}
	log.Printf("(DST) GoMultiFilesDropNotify  source id:%s ip:[%s] fileCnt:%d totalSize:%d firstFileName:%s firstFileSize:%d", id, ip, fileCnt, totalSize, firstFileName, firstFileSize)
	callbackFileListDragNotify(ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
}

func GoDragFileListNotify(ip, id, platform string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64) {
	if callbackFileListDragNotify == nil {
		log.Println(" callbackFileListDragNotify is null !")
		return
	}
	log.Printf("(DST) GoDragFileListNotify  source id:%s ip:[%s] fileCnt:%d totalSize:%d firstFileName:%s firstFileSize:%d", id, ip, fileCnt, totalSize, firstFileName, firstFileSize)
	callbackFileListDragNotify(ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
}

func GoDragFileListFolderNotify(ip, id, folderName string, timestamp uint64) {
	if callbackFileListDragFolderNotify == nil {
		log.Println(" callbackFileListDragFolderNotify is null !")
		return
	}
	log.Printf("(DST) GoDragFileListFolderNotify source id:%s ip:[%s] folder:[%s] timestamp:%d", id, ip, folderName, timestamp)
	callbackFileListDragFolderNotify(ip, id, folderName, timestamp)
}

func GoUpdateClientStatusEx(id string, status uint8) {
	if callbackUpdateClientStatus == nil {
		log.Printf("callbackUpdateClientStatus is null!\n\n")
		return
	}

	var clientInfo rtkCommon.ClientStatusInfo
	if status == 1 { // online
		info, err := rtkUtils.GetClientInfo(id)
		if err != nil {
			log.Printf("[%s] err:%+v", rtkMisc.GetFuncInfo(), err)
			return
		}
		clientInfo.ClientInfo = info
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

	callbackUpdateClientStatus(string(encodedData))
}

func FoundPeer() {

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

func GoUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	if callbackUpdateMultipleProgressBar == nil {
		log.Println("CallbackUpdateMultipleProgressBar is null !")
		return
	}
	//log.Printf("GoUpdateMultipleProgressBar ip:[%s] currentFileName:[%s] recvSize:[%d] total:[%d]", ip, currentFileName, sentSize, totalSize)
	callbackUpdateMultipleProgressBar(ip, id, currentFileName, sentFileCnt, totalFileCnt, currentFileSize, totalSize, sentSize, timestamp)
}

func GoUpdateSystemInfo(ip, serviceVer string) {

}

func GoUpdateClientStatus(status uint32, ip, id, name, deviceType string) {

}

func GoNotiMessageFileTransfer(fileName, clientName, platform string, timestamp uint64, isSender bool) {

}

func GoEventHandle(eventType rtkCommon.EventType, id, fileName string, timestamp uint64) {

}

func GoNotifyErrEvent(id string, errCode rtkMisc.CrossShareErr, arg1, arg2, arg3, arg4 string) {
	if callbackNotifyErrEvent == nil {
		log.Printf("callbackNotifyErrEvent is null!\n")
		return
	}

	callbackNotifyErrEvent(id, uint32(errCode), arg1, arg2, arg3, arg4)
}

func GoRequestUpdateClientVersion(ver string) {
	if callbackRequestUpdateClientVersion == nil {
		log.Println("callbackRequestUpdateClientVersion is null!")
		return
	}

	callbackRequestUpdateClientVersion(ver)
}

func GoCleanClipboard() {

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
	return rtkMisc.PlatformMac
}

func LockFile(file *os.File) error {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		log.Println("Failed to lock file:", err) //err:  resource temporarily unavailable
	}
	return err
}

func UnlockFile(file *os.File) error {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN|syscall.LOCK_NB)
	if err != nil {
		log.Println("Failed to lock file:", err)
	}
	return err
}

// Deprecated: unused
func SetNetWorkConnected(bConnected bool) {
	isNetWorkConnected = bConnected
}

// Deprecated: unused
func GetNetWorkConnected() bool {
	return isNetWorkConnected
}

func GetConfirmDocumentsAccept() bool {
	return ifConfirmDocumentsAccept
}

func GoNotifyBrowseResult(monitorName, instance, ipAddr, version string, timestamp int64) {

}

func GoAuthViaIndex(clientIndex uint32) {
	if callbackAuthViaIndex == nil {
		log.Printf("[%s] callbackAuthViaIndex is nil, GoAuthViaIndex failed!", rtkMisc.GetFuncInfo())
		return
	}
	callbackAuthViaIndex(clientIndex)
}

func GoReqSourceAndPort() {
	if callbackReqSourceAndPort == nil {
		log.Printf("[%s] callbackReqSourceAndPort is nil,!", rtkMisc.GetFuncInfo())
	}
	callbackReqSourceAndPort()
}

func GoMonitorNameNotify(name string) {
	if callbackMonitorName == nil {
		log.Printf("[%s] callbackMonitorName is nil, MonitorNameNotify failed!", rtkMisc.GetFuncInfo())
		return
	}
	callbackMonitorName(name)
}

func GoDIASStatusNotify(diasStatus uint32) {
	log.Printf("[%s] diasStatus:%d", rtkMisc.GetFuncInfo(), diasStatus)
	if callbackDIASStatus == nil {
		log.Printf("[%s] callbackDIASStatus is nil, DIASStatusNotify failed!", rtkMisc.GetFuncInfo())
		return
	}
	callbackDIASStatus(diasStatus)

}

func GetAuthData() (rtkMisc.CrossShareErr, rtkMisc.AuthDataInfo) {
	return rtkMisc.ERR_BIZ_GET_CALLBACK_INSTANCE_NULL, rtkMisc.AuthDataInfo{}
}

func GoGetSrcAndPortFromIni() rtkMisc.SourcePort {
	return rtkUtils.GetDeviceSrcPort()
}

// Specific Platform: iOS. Browse and lookup MDNS from iOS
func GoStartBrowseMdns(instance, serviceType string) {
	callbackMethodStartBrowseMdns(instance, serviceType)
}

func GoStopBrowseMdns() {
	callbackMethodStopBrowseMdns()
}
