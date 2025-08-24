//go:build darwin && !ios

package platform

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"syscall"

	"github.com/libp2p/go-libp2p/core/crypto"
)

var (
	imageData                bytes.Buffer
	copyTextChan             = make(chan string, 100)
	isNetWorkConnected       bool // Deprecated: unused
	strDeviceName            string
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
	CallbackMethodText                     func(string)
	CallbackMethodImage                    func(string)
	EventCallback                          func(int)
	CallbackCopyImageFunc                  func(rtkCommon.FileSize, rtkCommon.ImgHeader, []byte)
	CallbackPasteImageFunc                 func()
	CallbackMethodFileConfirm              func(string, string, string, int64)
	CallbackFileDropRequestFunc            func(string, rtkCommon.FileInfo, uint64)
	CallbackFileDropResponseFunc           func(string, rtkCommon.FileDropCmd, string)
	CallbackFileDragInfoFunc               func(rtkCommon.FileInfo, uint64)
	CallbackDragFileListRequestFunc        func([]rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackFileDragNotify                 func(string, string, string, string, uint64, uint64)
	CallbackFileListDragNotify             func(string, string, string, uint32, uint64, uint64, string, uint64)
	CallbackFileListDragFolderNotify       func(string, string, string, uint64)
	CallbackFileListDropRequestFunc        func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackMethodFoundPeer                func()
	CallbackUpdateProgressBar              func(string, string, uint64, uint64, uint64)
	CallbackUpdateMultipleProgressBar      func(string, string, string, string, uint32, uint32, uint64, uint64, uint64, uint64)
	CallbackCancelFileTransFunc            func(string, string, uint64)
	CallbackFileError                      func(string, string, string)
	CallbackNotifyErrEventFunc             func(id string, errCode uint32, arg1, arg2, arg3, arg4 string)
	CallbackGetMacAddressFunc              func(string)
	CallbackAuthStatusCodeFunc             func(uint8)
	CallbackExtractDIASFunc                func()
	CallbackDIASSourceAndPortFunc          func(uint8, uint8)
	CallbackMethodStartBrowseMdns          func(string, string)
	CallbackMethodStopBrowseMdns           func()
	CallbackMethodBrowseMdnsResultFunc     func(string, string, int, string, string, string, string)
	CallbackDetectPluginEventFunc          func(isPlugin bool, productName string)
	CallbackDIASStatusFunc                 func(uint32)
	CallbackMonitorNameFunc                func(string)
	CallbackGetFilesTransCodeFunc          func(id string) rtkCommon.SendFilesRequestErrCode
	CallbackRequestUpdateClientVersionFunc func(string)
)

var (
	callbackNetworkSwitch              CallbackNetworkSwitchFunc              = nil
	callbackMethodText                 CallbackMethodText                     = nil
	callbackMethodImage                CallbackMethodImage                    = nil
	eventCallback                      EventCallback                          = nil
	callbackInstanceCopyImage          CallbackCopyImageFunc                  = nil
	callbackInstancePasteImage         CallbackPasteImageFunc                 = nil
	callbackMethodFileConfirm          CallbackMethodFileConfirm              = nil
	callbackInstanceFileDropRequestCB  CallbackFileDropRequestFunc            = nil
	callbackInstanceFileDropResponseCB CallbackFileDropResponseFunc           = nil
	callbackInstanceFileDragCB         CallbackFileDragInfoFunc               = nil
	callbackDragFileListRequestCB      CallbackDragFileListRequestFunc        = nil
	callbackFileDragNotify             CallbackFileDragNotify                 = nil
	callbackFileListDragNotify         CallbackFileListDragNotify             = nil
	callbackFileListDragFolderNotify   CallbackFileListDragFolderNotify       = nil
	callbackFileListDropRequest        CallbackFileListDropRequestFunc        = nil
	callbackMethodFoundPeer            CallbackMethodFoundPeer                = nil
	callbackUpdateProgressBar          CallbackUpdateProgressBar              = nil
	callbackUpdateMultipleProgressBar  CallbackUpdateMultipleProgressBar      = nil
	callbackCancelFileTrans            CallbackCancelFileTransFunc            = nil
	callbackFileError                  CallbackFileError                      = nil
	callbackNotifyErrEvent             CallbackNotifyErrEventFunc             = nil
	callbackGetMacAddress              CallbackGetMacAddressFunc              = nil
	callbackAuthStatusCodeCB           CallbackAuthStatusCodeFunc             = nil
	callbackExtractDIAS                CallbackExtractDIASFunc                = nil
	callbackDIASSourceAndPortCB        CallbackDIASSourceAndPortFunc          = nil
	callbackMethodStartBrowseMdns      CallbackMethodStartBrowseMdns          = nil
	callbackMethodStopBrowseMdns       CallbackMethodStopBrowseMdns           = nil
	callbackMethodBrowseMdnsResult     CallbackMethodBrowseMdnsResultFunc     = nil
	callbackDetectPluginEvent          CallbackDetectPluginEventFunc          = nil
	callbackDIASStatus                 CallbackDIASStatusFunc                 = nil
	callbackMonitorName                CallbackMonitorNameFunc                = nil
	callbackGetFilesTransCode          CallbackGetFilesTransCodeFunc          = nil
	callbackRequestUpdateClientVersion CallbackRequestUpdateClientVersionFunc = nil
)

/*======================================= Used by main.go, set Callback =======================================*/

func SetCallbackMethodText(cb CallbackMethodText) {
	callbackMethodText = cb
}

func SetCallbackMethodImage(cb CallbackMethodImage) {
	callbackMethodImage = cb
}

func SetEventCallback(cb EventCallback) {
	eventCallback = cb
}

func SetCallbackMethodFileConfirm(cb CallbackMethodFileConfirm) {
	callbackMethodFileConfirm = cb
}

func SetCallbackFileNotify(cb CallbackFileDragNotify) {
	callbackFileDragNotify = cb
}

func SetCallbackFileListNotify(cb CallbackFileListDragNotify) {
	callbackFileListDragNotify = cb
}

func SetCallbackFileListFolderNotify(cb CallbackFileListDragFolderNotify) {
	callbackFileListDragFolderNotify = cb
}

func SetCallbackMethodFoundPeer(cb CallbackMethodFoundPeer) {
	callbackMethodFoundPeer = cb
}

func SetCallbackUpdateProgressBar(cb CallbackUpdateProgressBar) {
	callbackUpdateProgressBar = cb
}

func SetCallbackUpdateMultipleProgressBar(cb CallbackUpdateMultipleProgressBar) {
	callbackUpdateMultipleProgressBar = cb
}

func SetCallbackFileError(cb CallbackFileError) {
	callbackFileError = cb
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
func SetCopyImageCallback(cb CallbackCopyImageFunc) {
	callbackInstanceCopyImage = cb
}

func SetPasteImageCallback(cb CallbackPasteImageFunc) {
	callbackInstancePasteImage = cb
}

func SetGoFileDropRequestCallback(cb CallbackFileDropRequestFunc) {
	callbackInstanceFileDropRequestCB = cb
}

func SetGoFileDropResponseCallback(cb CallbackFileDropResponseFunc) {
	callbackInstanceFileDropResponseCB = cb
}

func SetGoDragFileCallback(cb CallbackFileDragInfoFunc) {
	callbackInstanceFileDragCB = cb
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

func GoReqSourceAndPort() {
}

func SetGoAuthStatusCodeCallback(cb CallbackAuthStatusCodeFunc) {
	callbackAuthStatusCodeCB = cb
}

func GoAuthViaIndex(clientIndex uint32) {

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

func SetDeviceName(name string) {
	strDeviceName = name
	log.Printf("SetDeviceName , device name:[%s]", strDeviceName)
}

func GoTriggerNetworkSwitch() {
	callbackNetworkSwitch()
}

func GoBrowseMdnsResultCallback(instance, ip string, port int) {
	if callbackMethodBrowseMdnsResult == nil {
		log.Println("CallbackMethodBrowseMdnsResult is null!")
		return
	}

	log.Printf("[%s] instance:[%s], ip:[%s], port:[%d]", rtkMisc.GetFuncInfo(), instance, ip, port)
	callbackMethodBrowseMdnsResult(instance, ip, port)
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

func GoCopyImage(fileSize rtkCommon.FileSize, imgHeader rtkCommon.ImgHeader, data []byte) {
	callbackInstanceCopyImage(fileSize, imgHeader, data)
}

func GoPasteImage() {
	callbackInstancePasteImage()
}

func GoFileDropRequest(id string, fileInfo rtkCommon.FileInfo, timestamp uint64) {
	callbackInstanceFileDropRequestCB(id, fileInfo, timestamp)
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

func GoSetSrcAndPort(source, port int) {
	log.Printf("[%s] , source:[%d], port:[%d]", rtkMisc.GetFuncInfo(), source, port)
	rtkUtils.DeviceSrcAndPort.Source = source
	rtkUtils.DeviceSrcAndPort.Port = port
}

func SendMessage(strText string) {
	log.Printf("SendMessage:[%s] ", strText)
	if strText == "" || len(strText) == 0 {
		return
	}

	nCount := rtkUtils.GetClientCount()
	for i := 0; i < nCount; i++ {
		copyTextChan <- strText
	}
}

func GoDragFileRequest(fileInfo rtkCommon.FileInfo, timestamp uint64) {
	callbackInstanceFileDragCB(fileInfo, timestamp)
}

/*======================================= Used  by GO business =======================================*/

func WatchClipboardText(ctx context.Context, resultChan chan<- string) {
	for {
		select {
		case <-ctx.Done():
			close(resultChan)
			return
		case curCopyText := <-copyTextChan:
			if len(curCopyText) > 0 {
				log.Println("DEBUG: watchClipboardText - got new message: ", curCopyText)
				resultChan <- curCopyText
			}
		}
	}
}

func GoSetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {
	if callbackMethodFileConfirm == nil {
		log.Println("CallbackMethodFileConfirm is null!")
		return
	}
	fileSize := int64(fileSizeHigh)<<32 | int64(fileSizeLow)
	log.Printf("(DST) GoSetupDstPasteFile  sourceID:%s fileName:[%s] fileSize:[%d]", desc, fileName, fileSize)
	callbackMethodFileConfirm("", platform, fileName, fileSize)
}

func GoSetupFileDrop(ip, id, fileName, platform string, fileSize uint64, timestamp uint64) {
	if callbackMethodFileConfirm == nil {
		log.Println("CallbackMethodFileConfirm is null!")
		return
	}
	log.Printf("(DST) GoSetupFileDrop  source:%s ip:[%s]fileName:%s  fileSize:%d", id, ip, fileName, fileSize)
	callbackMethodFileConfirm(id, platform, fileName, int64(fileSize))
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

func GoDragFileNotify(ip, id, fileName, platform string, fileSize uint64, timestamp uint64) {
	if callbackFileDragNotify == nil {
		log.Println("CallbackFileDragNotify is null!")
		return
	}
	log.Printf("(DST) GoDragFileNotify  source:%s ip:[%s]fileName:%s fileSize:%d timestamp:%d", id, ip, fileName, fileSize, timestamp)
	callbackFileDragNotify(ip, id, platform, fileName, fileSize, timestamp)
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
	log.Printf("(DST) GoDragFileListFolderNotify source id:%s ip:[%s] folderName:[%s] timestamp:%d", id, ip, folderName, timestamp)
	callbackFileListDragFolderNotify(ip, id, folderName, timestamp)
}

func ReceiveImageCopyDataDone(fileSize int64, imgHeader rtkCommon.ImgHeader) {
	log.Printf("[%s]: size:%d, (width, height):(%d,%d)", rtkMisc.GetFuncInfo(), fileSize, imgHeader.Width, imgHeader.Height)
	if callbackMethodImage == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	rtkMisc.GoSafe(func() {
		imageBase64 := rtkUtils.Base64Encode(imageData.Bytes())
		// log.Printf("len[%d][%d][%d][%+v]", len(ImageData), len(imageBase64), rtkGlobal.Handler.CopyImgHeader.Width, imageBase64)
		callbackMethodImage(imageBase64)
		imageData.Reset()
	})
}

func FoundPeer() {
	log.Println("CallbackMethodFoundPeer")
	if callbackMethodFoundPeer == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	callbackMethodFoundPeer()
}

func GoSetupDstPasteText(content []byte) {
	log.Printf("GoSetupDstPasteText:%s \n\n", string(content))
	if callbackMethodText == nil {
		log.Println("GoSetupDstPasteText - failed - callbackMethodText is nil")
		return
	}
	callbackMethodText(string(content))
}

func GoSetupDstPasteImage(desc string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {
	log.Printf("GoSetupDstPasteImage from ID %s, len:[%d] dataSize:[%d]\n\n", desc, len(content), dataSize)
	imageData.Reset()
	imageData.Grow(int(dataSize))
	callbackInstancePasteImage()
}

func GoDataTransfer(data []byte) {
	imageData.Write(data)
}

func GoUpdateProgressBar(ip, id string, fileSize, sentSize uint64, timestamp uint64, filePath string) {
	if callbackUpdateProgressBar == nil {
		log.Println("CallbackUpdateProgressBar is null !")
		return
	}
	fileName := filepath.Base(filePath)
	//log.Printf("GoUpdateProgressBar ip:[%s] name:[%s] recvSize:[%d] total:[%d]", ip, fileName, sentSize, fileSize)
	callbackUpdateProgressBar(id, fileName, uint64(sentSize), uint64(fileSize), uint64(timestamp))
}

func GoUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	if callbackUpdateMultipleProgressBar == nil {
		log.Println("CallbackUpdateMultipleProgressBar is null !")
		return
	}
	//log.Printf("GoUpdateMultipleProgressBar ip:[%s] [%s] currentFileName:[%s] recvSize:[%d] total:[%d]", ip, deviceName, currentFileName, sentSize, totalSize)
	callbackUpdateMultipleProgressBar(ip, id, deviceName, currentFileName, sentFileCnt, totalFileCnt, currentFileSize, totalSize, sentSize, timestamp)
}

func GoExtractDIASCallback() {

}

func GoUpdateSystemInfo(ip, serviceVer string) {

}

func GoUpdateClientStatus(status uint32, ip, id, name, deviceType string) {

}

func GoNotiMessageFileTransfer(fileName, clientName, platform string, timestamp uint64, isSender bool) {

}

func GoEventHandle(eventType rtkCommon.EventType, id, fileName string, timestamp uint64) {
	if callbackFileError == nil {
		log.Println("GoEventHandle CallbackInstance is null !")
		return
	}
	if eventType == rtkCommon.EVENT_TYPE_OPEN_FILE_ERR {
		strErr := "file datatransfer sender error"
		callbackFileError(id, fileName, strErr)
	} else if eventType == rtkCommon.EVENT_TYPE_RECV_TIMEOUT {
		strErr := "file datatransfer receiving end error"
		callbackFileError(id, fileName, strErr)
	}
	log.Printf("[%s %d]: id:%s, name:%s, error:%d", rtkMisc.GetFuncName(), rtkMisc.GetLine(), id, fileName, eventType)
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

func GoGetDeviceName() string {
	return strDeviceName
}

func GetConfirmDocumentsAccept() bool {
	return ifConfirmDocumentsAccept
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
