//go:build ios

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
	lockFd                   *os.File
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

type (
	CallbackNetworkSwitchFunc              func()
	CallbackCopyXClipFunc                  func(cbText, cbImage, cbHtml, cbRtf []byte)
	CallbackPasteXClipFunc                 func(text, image, html, rtf string)
	CallbackFileDropResponseFunc           func(string, rtkCommon.FileDropCmd, string)
	CallbackDragFileListRequestFunc        func([]rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackFileListNotify                 func(string, string, uint32, uint64, uint64, string, uint64)
	CallbackFileListDragFolderNotify       func(string, string, string, uint64)
	CallbackFileListDropRequestFunc        func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackUpdateClientStatusFunc         func(clientInfo string)
	CallbackUpdateProgressBar              func(string, string, string, uint32, uint32, uint64, uint64, uint64, uint64)
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
	CallbackGetAuthDataFunc                func(uint32) string
	CallbackDIASStatusFunc                 func(uint32)
	CallbackMonitorNameFunc                func(string)
	CallbackGetFilesTransCodeFunc          func(id string) rtkCommon.SendFilesRequestErrCode
	CallbackGetFilesCacheSendCountFunc     func(id string) int
	CallbackRequestUpdateClientVersionFunc func(string)
	CallbackNotifyBrowseResultFunc         func(monitorName, instance, ipAddr, version string, timestamp int64)
	CallbackConnectLanServerFunc           func(instance string)
	CallbackBrowseLanServerFunc            func()
	CallbackSetMsgEventFunc                func(event uint32, arg1, arg2, arg3, arg4 string)
)

var (
	callbackNetworkSwitch              CallbackNetworkSwitchFunc              = nil
	callbackCopyXClipData              CallbackCopyXClipFunc                  = nil
	callbackPasteXClipData             CallbackPasteXClipFunc                 = nil
	callbackInstanceFileDropResponseCB CallbackFileDropResponseFunc           = nil
	callbackDragFileListRequestCB      CallbackDragFileListRequestFunc        = nil
	callbackFileListSendNotify         CallbackFileListNotify                 = nil
	callbackFileListReceiveNotify      CallbackFileListNotify                 = nil
	callbackFileListDragFolderNotify   CallbackFileListDragFolderNotify       = nil
	callbackFileListDropRequest        CallbackFileListDropRequestFunc        = nil
	callbackUpdateClientStatus         CallbackUpdateClientStatusFunc         = nil
	callbackUpdateSendProgressBar      CallbackUpdateProgressBar              = nil
	callbackUpdateReceiveProgressBar   CallbackUpdateProgressBar              = nil
	callbackCancelFileTrans            CallbackCancelFileTransFunc            = nil
	callbackNotifyErrEvent             CallbackNotifyErrEventFunc             = nil
	callbackGetMacAddress              CallbackGetMacAddressFunc              = nil
	callbackAuthStatusCodeCB           CallbackAuthStatusCodeFunc             = nil
	callbackExtractDIAS                CallbackExtractDIASFunc                = nil
	callbackDIASSourceAndPortCB        CallbackDIASSourceAndPortFunc          = nil
	callbackMethodStartBrowseMdns      CallbackMethodStartBrowseMdns          = nil
	callbackMethodStopBrowseMdns       CallbackMethodStopBrowseMdns           = nil
	callbackMethodBrowseMdnsResult     CallbackMethodBrowseMdnsResultFunc     = nil
	callbackDetectPluginEvent          CallbackDetectPluginEventFunc          = nil
	callbackGetAuthData                CallbackGetAuthDataFunc                = nil
	callbackDIASStatus                 CallbackDIASStatusFunc                 = nil
	callbackMonitorName                CallbackMonitorNameFunc                = nil
	callbackGetFilesTransCode          CallbackGetFilesTransCodeFunc          = nil
	callbackGetFilesCacheSendCount     CallbackGetFilesCacheSendCountFunc     = nil
	callbackRequestUpdateClientVersion CallbackRequestUpdateClientVersionFunc = nil
	callbackNotifyBrowseResult         CallbackNotifyBrowseResultFunc         = nil
	callbackConnectLanServer           CallbackConnectLanServerFunc           = nil
	callbackBrowseLanServer            CallbackBrowseLanServerFunc            = nil
	callbackSetMsgEvent                CallbackSetMsgEventFunc                = nil
)

/*======================================= Used by main.go, set Callback =======================================*/

func SetCallbackFileListSendNotify(cb CallbackFileListNotify) {
	callbackFileListSendNotify = cb
}

func SetCallbackFileListReceiveNotify(cb CallbackFileListNotify) {
	callbackFileListReceiveNotify = cb
}

func SetCallbackFileListFolderNotify(cb CallbackFileListDragFolderNotify) {
	callbackFileListDragFolderNotify = cb
}

func SetCallbackUpdateClientStatus(cb CallbackUpdateClientStatusFunc) {
	callbackUpdateClientStatus = cb
}

func SetCallbackUpdateSendProgressBar(cb CallbackUpdateProgressBar) {
	callbackUpdateSendProgressBar = cb
}

func SetCallbackUpdateReceiveProgressBar(cb CallbackUpdateProgressBar) {
	callbackUpdateReceiveProgressBar = cb
}

func SetCallbackMethodStartBrowseMdns(cb CallbackMethodStartBrowseMdns) {
	callbackMethodStartBrowseMdns = cb
}

func SetCallbackMethodStopBrowseMdns(cb CallbackMethodStopBrowseMdns) {
	callbackMethodStopBrowseMdns = cb
}

func SetCallbackGetAuthData(cb CallbackGetAuthDataFunc) {
	callbackGetAuthData = cb
}

func SetCallbackDIASStatus(cb CallbackDIASStatusFunc) {
	callbackDIASStatus = cb
}

func SetCallbackMonitorName(cb CallbackMonitorNameFunc) {
	callbackMonitorName = cb
}

func SetCallbackPasteXClipData(cb CallbackPasteXClipFunc) {
	callbackPasteXClipData = cb
}

func SetCallbackRequestUpdateClientVersion(cb CallbackRequestUpdateClientVersionFunc) {
	callbackRequestUpdateClientVersion = cb
}

func SetCallbackNotifyErrEvent(cb CallbackNotifyErrEventFunc) {
	callbackNotifyErrEvent = cb
}

func SetCallbackNotifyBrowseResult(cb CallbackNotifyBrowseResultFunc) {
	callbackNotifyBrowseResult = cb
}

/*======================================= Used  by GO set Callback =======================================*/

func SetGoNetworkSwitchCallback(cb CallbackNetworkSwitchFunc) {
	callbackNetworkSwitch = cb
}

// Notify to Clipboard/FileDrop
func SetCopyXClipCallback(cb CallbackCopyXClipFunc) {
	callbackCopyXClipData = cb
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

func SetGetFilesCacheSendCountCallback(cb CallbackGetFilesCacheSendCountFunc) {
	callbackGetFilesCacheSendCount = cb
}

func SetGoConnectLanServerCallback(cb CallbackConnectLanServerFunc) {
	callbackConnectLanServer = cb
}

func SetGoBrowseLanServerCallback(cb CallbackBrowseLanServerFunc) {
	callbackBrowseLanServer = cb
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

func GoGetMacAddressCallback(mac string) {
	log.Printf("[%s]  mac :[%s]", rtkMisc.GetFuncInfo(), mac)
	callbackGetMacAddress(mac)
}

func GoTriggerDetectPluginEvent(isPlugin bool) {
	callbackDetectPluginEvent(isPlugin, "")
}

func SetConfirmDocumentsAccept(ifConfirm bool) {
	ifConfirmDocumentsAccept = ifConfirm
}

func GoConnectLanServer(instance string) {
	if callbackConnectLanServer == nil {
		log.Println("callbackConnectLanServer is null!")
		return
	}

	rtkMisc.GoSafe(func() { callbackConnectLanServer(instance) })
}

func GoBrowseLanServer() {
	if callbackBrowseLanServer == nil {
		return
	}

	rtkMisc.GoSafe(func() { callbackBrowseLanServer() })
}

func GoCopyXClipData(text, image, html, rtf []byte) {
	if callbackCopyXClipData == nil {
		log.Println("callbackCopyXClipData is null!")
		return
	}

	callbackCopyXClipData(text, image, html, rtf)
}

func GoFileDropResponse(id string, fileCmd rtkCommon.FileDropCmd, fileName string) {
	callbackInstanceFileDropResponseCB(id, fileCmd, fileName)
}

func GoMultiFilesDropRequest(id string, fileList *[]rtkCommon.FileInfo, folderList *[]string, totalSize, timestamp uint64, totalDesc string) rtkCommon.SendFilesRequestErrCode {
	if callbackFileListDropRequest == nil {
		log.Println("CallbackFileListDropRequest is null!")
		return rtkCommon.SendFilesRequestCallbackNotSet
	}

	if callbackGetFilesCacheSendCount == nil {
		log.Println("callbackGetFilesCacheSendCount is null!")
		return rtkCommon.SendFilesRequestCallbackNotSet
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

	callbackFileListDropRequest(id, *fileList, *folderList, totalSize, timestamp, totalDesc)
	return rtkCommon.SendFilesRequestSuccess
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

func GoFileListSendNotify(ip, id string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64) {
	if callbackFileListSendNotify == nil {
		log.Println(" callbackFileListSendNotify is null !")
		return
	}
	callbackFileListSendNotify(ip, id, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
}

func GoFileListReceiveNotify(ip, id string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64) {
	if callbackFileListReceiveNotify == nil {
		log.Println(" callbackFileListReceiveNotify is null !")
		return
	}

	callbackFileListReceiveNotify(ip, id, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
}

func GoDragFileListFolderNotify(ip, id, folderName string, timestamp uint64) {
	if callbackFileListDragFolderNotify == nil {
		//log.Println(" callbackFileListDragFolderNotify is null !")
		return
	}
	log.Printf("(DST) GoDragFileListFolderNotify source id:%s ip:[%s] folderName:[%s] timestamp:%d", id, ip, folderName, timestamp)
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
	log.Printf("[%s] Str:%s", rtkMisc.GetFuncInfo(), string(encodedData))
	callbackUpdateClientStatus(string(encodedData))
}

func GoSetupDstPasteXClipData(cbText, cbImage, cbHtml, cbRtf []byte) {
	if callbackPasteXClipData == nil {
		log.Printf("callbackPasteXClipData is null!\n\n")
		return
	}

	imageStr := ""
	if len(cbImage) > 0 {
		imageBase64 := rtkUtils.Base64Encode(cbImage)
		imageStr = imageBase64
	}

	callbackPasteXClipData(string(cbText), imageStr, string(cbHtml), string(cbRtf))
}

func GoUpdateSendProgressBar(ip, id, currentFileName string, sendFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sendSize, timestamp uint64) {
	if callbackUpdateSendProgressBar == nil {
		log.Println("callbackUpdateSendProgressBar is null !")
		return
	}
	//log.Printf("GoUpdateSendProgressBar ip:[%s] currentFileName:[%s] sendSize:[%d] total:[%d]", ip, currentFileName, sendSize, totalSize)
	callbackUpdateSendProgressBar(ip, id, currentFileName, sendFileCnt, totalFileCnt, currentFileSize, totalSize, sendSize, timestamp)
}

func GoUpdateReceiveProgressBar(ip, id, currentFileName string, recvFileCnt, totalFileCnt uint32, currentFileSize, totalSize, recvSize, timestamp uint64) {
	if callbackUpdateReceiveProgressBar == nil {
		log.Println("callbackUpdateReceiveProgressBar is null !")
		return
	}
	//log.Printf("GoUpdateReceiveProgressBar ip:[%s] currentFileName:[%s] recvSize:[%d] total:[%d]", ip, currentFileName, recvSize, totalSize)
	callbackUpdateReceiveProgressBar(ip, id, currentFileName, recvFileCnt, totalFileCnt, currentFileSize, totalSize, recvSize, timestamp)
}

func GoUpdateSystemInfo(ip, serviceVer string) {

}

func GoNotiMessageFileTransfer(fileName, clientName, platform string, timestamp uint64, isSender bool) {

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
	return rtkMisc.PlatformiOS
}

func LockFile() error {
	var err error
	lockFd, err = os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Failed to open or create lock file:[%s] err:%+v", lockFile, err)
		return err
	}

	err = syscall.Flock(int(lockFd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		log.Printf("Failed to lock file[%s] err:%+v", lockFile, err) //err:  resource temporarily unavailable
	}
	return err
}

func UnlockFile() error {
	err := syscall.Flock(int(lockFd.Fd()), syscall.LOCK_UN|syscall.LOCK_NB)
	if err != nil {
		log.Printf("Failed to unlock file[%s] err:%+v", lockFile, err)
	}

	lockFd.Close()
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
	if callbackNotifyBrowseResult == nil {
		log.Println("[%s] failed, callbackNotifyBrowseResult is nil", rtkMisc.GetFuncInfo())
		return
	}

	callbackNotifyBrowseResult(monitorName, instance, ipAddr, version, timestamp)
}

func GoAuthViaIndex(clientIndex uint32) {

}

func GoReqSourceAndPort() {
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

func GetAuthData(clientIndex uint32) (rtkMisc.CrossShareErr, rtkMisc.AuthDataInfo) {
	if callbackGetAuthData == nil {
		log.Println("callbackGetAuthData is null !")
		return rtkMisc.ERR_BIZ_GET_CALLBACK_INSTANCE_NULL, rtkMisc.AuthDataInfo{}
	}
	authDataJsonInfo := callbackGetAuthData(clientIndex)
	log.Printf("[%s] get json data:[%s]", rtkMisc.GetFuncInfo(), authDataJsonInfo)

	var authData rtkMisc.AuthDataInfo
	err := json.Unmarshal([]byte(authDataJsonInfo), &authData)
	if err != nil {
		log.Printf("[%s] Unmarshal[%s] err:%+v", rtkMisc.GetFuncInfo(), authDataJsonInfo, err)
		return rtkMisc.ERR_BIZ_JSON_UNMARSHAL, rtkMisc.AuthDataInfo{}
	}

	log.Printf("[%s] width:[%d] height:[%d] Framerate:[%d] type:[%d] DisplayName:[%s]", rtkMisc.GetFuncInfo(), authData.Width, authData.Height, authData.Framerate, authData.Type, authData.DisplayName)
	return rtkMisc.SUCCESS, authData
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

func GoSetupAppLink(link string) {
	rtkMisc.AppLink = link
}
