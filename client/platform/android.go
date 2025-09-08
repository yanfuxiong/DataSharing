//go:build android

package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync/atomic"
	"syscall"

	"github.com/libp2p/go-libp2p/core/crypto"
)

var (
	isConnecting             atomic.Bool
	imageData                bytes.Buffer
	copyTextChan             = make(chan string, 100)
	isNetWorkConnected       bool
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

// Notify to Android APK
type Callback interface {
	CallbackPasteXClipData(text, image, html string)
	CallbackMethod(result string)
	CallbackMethodImage(content string)
	LogMessageCallback(msg string)
	EventCallback(event int)
	CallbackMethodFileConfirm(id, platform, filename string, fileSize int64)
	CallbackFileDragNotify(id, platform, filename string, fileSize int64) // Deprecated: unused
	CallbackFileListDragNotify(ip, id, platform string, fileCnt int, totalSize, timestamp int64, firstFileName string, firstFileSize int64)
	CallbackFileListDragFolderNotify(ip, id, folderName string, timestamp int64)
	CallbackFilesTransferDone(filesInfo, platform, deviceName string, timestamp int64)
	CallbackMethodFoundPeer()
	CallbackUpdateProgressBar(ip, id, filename string, recvSize, total int64, timestamp int64) // Deprecated: unused
	CallbackUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt int, currentFileSize, totalSize, sentSize, timestamp int64)
	CallbackFileError(id, filename, err string, timestamp int64)
	CallbackNotifyErrEvent(id string, errCode int, arg1, arg2, arg3, arg4 string)
	CallbackUpdateDiasStatus(status int)
	CallbackGetAuthData() string
	CallbackUpdateMonitorName(monitorName string)
	CallbackRequestUpdateClientVersion(clienVersion string)
	CallbackNotifyBrowseResult(monitorName, instance, ipAddr, version string, timestamp int64)
	CallbackUpdateClientStatus(clientInfo string)
}

var CallbackInstance Callback = nil

func SetCallback(cb Callback) {
	CallbackInstance = cb
}

func initFilePath() {
	privKeyFile = ".priv.pem"
	hostID = ".HostID"
	nodeID = ".ID"
	lockFile = "singleton.lock"
	logFile = "p2p.log"
	crashLogFile = "crash.log"
	downloadPath = ""
}

func init() {
	ifConfirmDocumentsAccept = false
	rootPath = ""
	isNetWorkConnected = false
	isConnecting.Store(false)
}

func GetRootPath() string {
	return rootPath
}

func GetLogFilePath() string {
	return logFile
}

func GetCrashLogFilePath() string {
	return crashLogFile
}

func GetDownloadPath() string {
	return downloadPath
}

func GetLockFilePath() string {
	return lockFile
}

type (
	CallbackNetworkSwitchFunc          func()
	CallbackCopyImageFunc              func(rtkCommon.ImgHeader, []byte)
	CallbackXClipCopyFunc              func(cbText, cbImage, cbHtml []byte)
	CallbackPasteImageFunc             func()
	CallbackFileDropResponseFunc       func(string, rtkCommon.FileDropCmd, string)
	CallbackFileListDropRequestFunc    func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackDragFileListRequestFunc    func([]rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackGetMacAddressFunc          func(string)
	CallbackCancelFileTransFunc        func(string, string, uint64)
	CallbackDetectPluginEventFunc      func(isPlugin bool, productName string)
	CallbackDIASSourceAndPortFunc      func(uint8, uint8)
	CallbackAuthStatusCodeFunc         func(uint8)
	CallbackExtractDIASFunc            func()
	CallbackMethodBrowseMdnsResultFunc func(string, string, int, string, string, string, string)
	CallbackGetFilesTransCodeFunc      func(id string) rtkCommon.SendFilesRequestErrCode
	CallbackConnectLanServerFunc       func(instance string)
	CallbackBrowseLanServerFunc        func()
	CallbackSetMsgEventFunc            func(event uint32, arg1, arg2, arg3, arg4 string)
)

var (
	callbackNetworkSwitchCB            CallbackNetworkSwitchFunc          = nil
	callbackInstanceCopyImageCB        CallbackCopyImageFunc              = nil
	callbackInstancePasteImageCB       CallbackPasteImageFunc             = nil
	callbackXClipCopyCB                CallbackXClipCopyFunc              = nil
	callbackInstanceFileDropResponseCB CallbackFileDropResponseFunc       = nil
	callbackFileListDropRequestCB      CallbackFileListDropRequestFunc    = nil
	callbackDragFileListRequestCB      CallbackDragFileListRequestFunc    = nil
	callbackGetMacAddressCB            CallbackGetMacAddressFunc          = nil
	callbackCancelFileTransDragCB      CallbackCancelFileTransFunc        = nil
	callbackDetectPluginEventCB        CallbackDetectPluginEventFunc      = nil
	callbackDIASSourceAndPortCB        CallbackDIASSourceAndPortFunc      = nil
	callbackAuthStatusCodeCB           CallbackAuthStatusCodeFunc         = nil
	callbackExtractDIASCB              CallbackExtractDIASFunc            = nil
	callbackMethodBrowseMdnsResult     CallbackMethodBrowseMdnsResultFunc = nil
	callbackGetFilesTransCode          CallbackGetFilesTransCodeFunc      = nil
	callbackConnectLanServer           CallbackConnectLanServerFunc       = nil
	callbackBrowseLanServer            CallbackBrowseLanServerFunc        = nil
	callbackSetMsgEvent                CallbackSetMsgEventFunc            = nil
)

func SetGoNetworkSwitchCallback(cb CallbackNetworkSwitchFunc) {
	callbackNetworkSwitchCB = cb
}

// Notify to Clipboard/FileDrop
func SetCopyXClipCallback(cb CallbackXClipCopyFunc) {
	callbackXClipCopyCB = cb
}

func SetCopyImageCallback(cb CallbackCopyImageFunc) {
	callbackInstanceCopyImageCB = cb
}

func SetPasteImageCallback(cb CallbackPasteImageFunc) {
	callbackInstancePasteImageCB = cb
}

func SetGoFileDropResponseCallback(cb CallbackFileDropResponseFunc) {
	callbackInstanceFileDropResponseCB = cb
}

func SetGoFileListDropRequestCallback(cb CallbackFileListDropRequestFunc) {
	callbackFileListDropRequestCB = cb
}

func SetGoDragFileListRequestCallback(cb CallbackDragFileListRequestFunc) {
	callbackDragFileListRequestCB = cb
}

func SetGoExtractDIASCallback(cb CallbackExtractDIASFunc) {
	callbackExtractDIASCB = cb
}

func SetGoGetMacAddressCallback(cb CallbackGetMacAddressFunc) {
	callbackGetMacAddressCB = cb
}

func SetDetectPluginEventCallback(cb CallbackDetectPluginEventFunc) {
	callbackDetectPluginEventCB = cb
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

func SetGoCancelFileTransCallback(cb CallbackCancelFileTransFunc) {
	callbackCancelFileTransDragCB = cb
}

func SetGetFilesTransCodeCallback(cb CallbackGetFilesTransCodeFunc) {
	callbackGetFilesTransCode = cb
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

/***************** Used  by android *****************/
func SetupRootPath(path string) {
	if path == "" {
		return
	}
	rootPath = path
	initFilePath()

	getPath := func(dirPath, filePath string) string {
		return filepath.Join(dirPath, filePath)
	}

	settingsDir := ".Settings" // TODO: Be hidden folder in the future
	logDir := "Log"
	//downloadDir := "Download"

	settingsPath := getPath(rootPath, settingsDir)
	logPath := getPath(rootPath, logDir)
	//downloadPath = getPath(rootPath, downloadDir)
	downloadPath = rootPath

	rtkMisc.CreateDir(settingsPath)
	rtkMisc.CreateDir(logPath)
	//rtkMisc.CreateDir(downloadPath)

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
	callbackSetMsgEvent(event, arg1, arg2, arg3, arg4)
}

func GoTriggerNetworkSwitch() {
	callbackNetworkSwitchCB()
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

func GoCopyXClipData(text, image, html []byte) {
	if callbackXClipCopyCB == nil {
		log.Println("callbackXClipCopyCB is null!")
		return
	}

	callbackXClipCopyCB(text, image, html)
}

func GoCopyImage(imgHeader rtkCommon.ImgHeader, data []byte) {
	callbackInstanceCopyImageCB(imgHeader, data)
}

func GoPasteImage() {
	callbackInstancePasteImageCB()
}

func GoFileDropResponse(id string, fileCmd rtkCommon.FileDropCmd, fileName string) {
	callbackInstanceFileDropResponseCB(id, fileCmd, fileName)
}

func GoMultiFilesDropRequest(id string, fileList *[]rtkCommon.FileInfo, folderList *[]string, totalSize, timestamp uint64, totalDesc string) rtkCommon.SendFilesRequestErrCode {
	if callbackFileListDropRequestCB == nil {
		log.Println("CallbackFileListDropRequestCB is null!")
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

	nMsgLength := int(rtkGlobal.P2PMsgMagicLength)

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

	callbackFileListDropRequestCB(id, *fileList, *folderList, totalSize, timestamp, totalDesc)
	return filesTransCode
}

func GoGetMacAddress(mac string) {
	callbackGetMacAddressCB(mac)
}

func GoTriggerDetectPluginEvent(isPlugin bool, productName string) {
	if callbackDetectPluginEventCB == nil {
		log.Println("callbackDetectPluginEventCB is null!")
		return
	}

	callbackDetectPluginEventCB(isPlugin, productName)
}

func GoCancelFileTrans(ip, id string, timestamp int64) {
	if callbackCancelFileTransDragCB == nil {
		log.Println("CallbackCancelFileTransDragCB is null!")
		return
	}
	callbackCancelFileTransDragCB(id, ip, uint64(timestamp))
}

func SetConfirmDocumentsAccept(ifConfirm bool) {
	ifConfirmDocumentsAccept = ifConfirm
}

func IsConnecting() bool {
	return isConnecting.Load()
}

func GoConnectLanServer(instance string) {
	isConnecting.Store(true)
	defer isConnecting.Store(false)

	if callbackConnectLanServer == nil {
		log.Println("callbackConnectLanServer is null!")
		return
	}

	callbackConnectLanServer(instance)
}

func GoBrowseLanServer() {
	if callbackBrowseLanServer == nil {
		log.Println("callbackBrowseLanServer is null!")
		return
	}

	callbackBrowseLanServer()
}

/***************** Used  by GO business *****************/

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
	fileSize := int64(fileSizeHigh)<<32 | int64(fileSizeLow)
	log.Printf("(DST) GoSetupDstPasteFile  sourceID:%s fileName:[%s] fileSize:[%d]", desc, fileName, fileSize)
	CallbackInstance.CallbackMethodFileConfirm("", platform, fileName, fileSize)
}

func GoSetupFileListDrop(ip, id, platform, totalDesc string, fileCount, folderCount uint32, timestamp uint64) {
	log.Printf("(DST) GoSetupFileListDrop  ID:]%s] IP:[%s] totalDesc:%s  fileCount:%d  folderCount:%d", id, ip, totalDesc, fileCount, folderCount)
}

func GoMultiFilesDropNotify(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64) {
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	log.Printf("(DST) GoMultiFilesDropNotify  source:%s ip:[%s]fileCnt:%d  totalSize:%d", id, ip, fileCnt, totalSize)
	CallbackInstance.CallbackFileListDragNotify(ip, id, platform, int(fileCnt), int64(totalSize), int64(timestamp), firstFileName, int64(firstFileSize))
}

func GoDragFileListNotify(ip, id, platform string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64) {
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	log.Printf("(DST) GoDragFileListNotify  source:%s ip:[%s]fileCnt:%d  totalSize:%d", id, ip, fileCnt, totalSize)
	CallbackInstance.CallbackFileListDragNotify(ip, id, platform, int(fileCnt), int64(totalSize), int64(timestamp), firstFileName, int64(firstFileSize))
}

func GoDragFileListFolderNotify(ip, id, folderName string, timestamp uint64) {
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	log.Printf("(DST) GoDragFileListFolderNotify source:%s ip:[%s] folder:[%s] timestamp:%d", id, ip, folderName, timestamp)
	CallbackInstance.CallbackFileListDragFolderNotify(ip, id, folderName, int64(timestamp))
}

func ReceiveImageCopyDataDone(fileSize int64, imgHeader rtkCommon.ImgHeader) {
	log.Printf("[%s]: size:%d, (width, height):(%d,%d)", rtkMisc.GetFuncInfo(), fileSize, imgHeader.Width, imgHeader.Height)
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	rtkMisc.GoSafe(func() {
		imageBase64 := rtkUtils.Base64Encode(imageData.Bytes())
		// log.Printf("len[%d][%d][%d][%+v]", len(ImageData), len(imageBase64), rtkGlobal.Handler.CopyImgHeader.Width, imageBase64)
		CallbackInstance.CallbackMethodImage(imageBase64)
		imageData.Reset()
	})
}

func FoundPeer() {
	log.Println("CallbackMethodFoundPeer")
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	CallbackInstance.CallbackMethodFoundPeer()
}

func GoUpdateClientStatusEx(id string, status uint8) {
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}

	var clientInfo rtkCommon.ClientStatusInfo
	if status == 1 {
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

	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), string(encodedData))
	CallbackInstance.CallbackUpdateClientStatus(string(encodedData))
}

func GoSetupDstPasteXClipData(cbText, cbImage, cbHtml []byte) {
	if CallbackInstance == nil {
		log.Println("GoSetupDstPasteText - failed - callbackInstance is nil")
		return
	}
	log.Printf("[%s] text len:%d , image len:%d, html:%d\n\n", rtkMisc.GetFuncInfo(), len(cbText), len(cbImage), len(cbHtml))

	imageStr := ""
	if len(cbImage) > 0 {
		imageBase64 := rtkUtils.Base64Encode(cbImage)
		imageStr = imageBase64
		log.Printf("call back image data:[%s]", imageStr[:20])
	}

	CallbackInstance.CallbackPasteXClipData(string(cbText), imageStr, string(cbHtml))
}

func GoSetupDstPasteText(content []byte) {
	log.Printf("GoSetupDstPasteText:%s \n\n", string(content))
	if CallbackInstance == nil {
		log.Println("GoSetupDstPasteText - failed - callbackInstance is nil")
		return
	}
	CallbackInstance.CallbackMethod(string(content))
}

func GoSetupDstPasteImage(desc string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {
	log.Printf("GoSetupDstPasteImage from ID %s, len:[%d] dataSize:[%d]\n\n", desc, len(content), dataSize)
	imageData.Reset()
	imageData.Grow(int(dataSize))
	callbackInstancePasteImageCB()
}

func GoDataTransfer(data []byte) {
	imageData.Write(data)
}

func GoUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	if CallbackInstance == nil {
		log.Println("CallbackUpdateMultipleProgressBar CallbackInstance is null !")
		return
	}
	//log.Printf("GoUpdateMultipleProgressBar ip:[%s] [%s] currentFileName:[%s] recvSize:[%d] total:[%d] timestamp:[%d]", ip, deviceName, currentFileName, sentSize, totalSize, timestamp)
	CallbackInstance.CallbackUpdateMultipleProgressBar(ip, id, deviceName, currentFileName, int(sentFileCnt), int(totalFileCnt), int64(currentFileSize), int64(totalSize), int64(sentSize), int64(timestamp))
}

func GoUpdateSystemInfo(ip, serviceVer string) {

}

func GoUpdateClientStatus(status uint32, ip, id, name, deviceType string) {

}

func GoNotiMessageFileTransfer(fileInfo, clientName, platform string, timestamp uint64, isSender bool) {
	if !isSender {
		return
	}
	log.Printf("[%s]: fileInfo:[%s], clientName:%s, timestamp:%d ", rtkMisc.GetFuncInfo(), fileInfo, clientName, timestamp)
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	CallbackInstance.CallbackFilesTransferDone(fileInfo, platform, clientName, int64(timestamp))
}

func GoEventHandle(eventType rtkCommon.EventType, id, fileName string, timestamp uint64) {
	if CallbackInstance == nil {
		log.Println("GoEventHandle CallbackInstance is null !")
		return
	}
	if eventType == rtkCommon.EVENT_TYPE_OPEN_FILE_ERR {
		strErr := "file datatransfer sender error"
		CallbackInstance.CallbackFileError(id, fileName, strErr, int64(timestamp))
	} else if eventType == rtkCommon.EVENT_TYPE_RECV_TIMEOUT {
		strErr := "file datatransfer receiving end error"
		CallbackInstance.CallbackFileError(id, fileName, strErr, int64(timestamp))
	}
	log.Printf("[%s %d]: id:%s, name:%s, error:%d", rtkMisc.GetFuncName(), rtkMisc.GetLine(), id, fileName, eventType)
}

func GoNotifyErrEvent(id string, errCode rtkMisc.CrossShareErr, arg1, arg2, arg3, arg4 string) {
	if CallbackInstance == nil {
		log.Println("GoEventHandle CallbackInstance is null !")
		return
	}

	log.Printf("[%s] id:[%s] errCode:%d arg1:%s, arg2:%s, arg3:%s, arg4:%s", rtkMisc.GetFuncInfo(), id, errCode, arg1, arg2, arg3, arg4)
	CallbackInstance.CallbackNotifyErrEvent(id, int(errCode), arg1, arg2, arg3, arg4)
}

func GoRequestUpdateClientVersion(ver string) {
	if CallbackInstance == nil {
		log.Println("GoRequestUpdateClientVersion CallbackInstance is null !")
		return
	}

	log.Printf("[%s] client version:%s \n\n", rtkMisc.GetFuncInfo(), ver)
	CallbackInstance.CallbackRequestUpdateClientVersion(ver)
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
	return rtkMisc.PlatformAndroid
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

func GoAuthViaIndex(clientIndex uint32) {

}

func GoReqSourceAndPort() {

}

func GoGetSrcAndPortFromIni() rtkMisc.SourcePort {
	return rtkUtils.GetDeviceSrcPort()
}

func GoNotifyBrowseResult(monitorName, instance, ipAddr, version string, timestamp int64) {
	if CallbackInstance == nil {
		log.Println("[%s] failed, callbackInstance is nil", rtkMisc.GetFuncInfo())
		return
	}
	log.Printf("[%s] name:%s, instance:%s, ip:%s, version:%s, timestamp:%d", rtkMisc.GetFuncInfo(), monitorName, instance, ipAddr, version, timestamp)
	CallbackInstance.CallbackNotifyBrowseResult(monitorName, instance, ipAddr, version, timestamp)
}

func GoMonitorNameNotify(name string) {
	if CallbackInstance == nil {
		log.Println("[%s] failed, callbackInstance is nil", rtkMisc.GetFuncInfo())
		return
	}
	log.Printf("[%s] monitor name: [%s]", rtkMisc.GetFuncInfo(), name)
	CallbackInstance.CallbackUpdateMonitorName(name)
}

func GoDIASStatusNotify(diasStatus uint32) {
	if CallbackInstance == nil {
		log.Println("GoDIASStatusNotify - failed - callbackInstance is nil")
		return
	}
	log.Printf("[%s] diasStatus:%d", rtkMisc.GetFuncInfo(), diasStatus)
	CallbackInstance.CallbackUpdateDiasStatus(int(diasStatus))
}

func GetAuthData() (rtkMisc.CrossShareErr, rtkMisc.AuthDataInfo) {
	if CallbackInstance == nil {
		log.Println("GetAuthData - failed - callbackInstance is nil")
		return rtkMisc.ERR_BIZ_GET_CALLBACK_INSTANCE_NULL, rtkMisc.AuthDataInfo{}
	}

	authDataJsonInfo := CallbackInstance.CallbackGetAuthData()
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

// Specific Platform: iOS. Browse and lookup MDNS from iOS
func GoStartBrowseMdns(instance, serviceType string) {
}

func GoStopBrowseMdns() {
}
