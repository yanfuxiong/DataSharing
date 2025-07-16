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
	"syscall"

	"github.com/libp2p/go-libp2p/core/crypto"
)

const (
	deviceSrcPort = "/storage/emulated/0/Android/data/com.realtek.crossshare/files/ID.SrcAndPort"
)

var (
	imageData                bytes.Buffer
	copyTextChan             = make(chan string, 100)
	isNetWorkConnected       bool // Deprecated: unused
	strDeviceName            string
	currentDiasStatus        uint32
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
	CallbackMethod(result string)
	CallbackMethodImage(content string)
	LogMessageCallback(msg string)
	EventCallback(event int)
	CallbackMethodFileConfirm(id, platform, filename string, fileSize int64)
	CallbackFileDragNotify(id, platform, filename string, fileSize int64)
	CallbackFileListDragNotify(ip, id, platform string, fileCnt int, totalSize, timestamp int64, firstFileName string, firstFileSize int64)
	CallbackFileListDragFolderNotify(ip, id, folderName string, timestamp int64)
	CallbackFilesTransferDone(filesInfo, platform, deviceName string, timestamp int64)
	CallbackMethodFoundPeer()
	CallbackUpdateProgressBar(ip, id, filename string, recvSize, total int64, timestamp int64)
	CallbackUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt int, currentFileSize, totalSize, sentSize, timestamp int64)
	CallbackFileError(id, filename, err string)
	CallbackUpdateDiasStatus(status int)
	CallbackGetAuthData() string
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
	rtkUtils.InitDeviceSrcAndPort(deviceSrcPort)
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
	CallbackCopyImageFunc              func(rtkCommon.FileSize, rtkCommon.ImgHeader, []byte)
	CallbackPasteImageFunc             func()
	CallbackFileDropRequestFunc        func(string, rtkCommon.FileInfo, uint64)
	CallbackFileDropResponseFunc       func(string, rtkCommon.FileDropCmd, string)
	CallbackFileDragInfoFunc           func(rtkCommon.FileInfo, uint64)
	CallbackFileListDropRequestFunc    func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackDragFileListRequestFunc    func([]rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackGetMacAddressFunc          func(string)
	CallbackCancelFileTransFunc        func(string, string, uint64)
	CallbackDetectPluginEventFunc      func(isPlugin bool, productName string)
	CallbackDIASSourceAndPortFunc      func(uint8, uint8)
	CallbackAuthStatusCodeFunc         func(uint8)
	CallbackExtractDIASFunc            func()
	CallbackMethodBrowseMdnsResultFunc func(string, string, int)
)

var (
	callbackNetworkSwitchCB            CallbackNetworkSwitchFunc          = nil
	callbackInstanceCopyImageCB        CallbackCopyImageFunc              = nil
	callbackInstancePasteImageCB       CallbackPasteImageFunc             = nil
	callbackInstanceFileDropRequestCB  CallbackFileDropRequestFunc        = nil
	callbackInstanceFileDropResponseCB CallbackFileDropResponseFunc       = nil
	callbackInstanceFileDragCB         CallbackFileDragInfoFunc           = nil
	callbackFileListDropRequestCB      CallbackFileListDropRequestFunc    = nil
	callbackDragFileListRequestCB      CallbackDragFileListRequestFunc    = nil
	callbackGetMacAddressCB            CallbackGetMacAddressFunc          = nil
	callbackCancelFileTransDragCB      CallbackCancelFileTransFunc        = nil
	callbackDetectPluginEventCB        CallbackDetectPluginEventFunc      = nil
	callbackDIASSourceAndPortCB        CallbackDIASSourceAndPortFunc      = nil
	callbackAuthStatusCodeCB           CallbackAuthStatusCodeFunc         = nil
	callbackExtractDIASCB              CallbackExtractDIASFunc            = nil
	callbackMethodBrowseMdnsResult     CallbackMethodBrowseMdnsResultFunc = nil
)

func SetGoNetworkSwitchCallback(cb CallbackNetworkSwitchFunc) {
	callbackNetworkSwitchCB = cb
}

// Notify to Clipboard/FileDrop
func SetCopyImageCallback(cb CallbackCopyImageFunc) {
	callbackInstanceCopyImageCB = cb
}

func SetPasteImageCallback(cb CallbackPasteImageFunc) {
	callbackInstancePasteImageCB = cb
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

func SetDeviceName(name string) {
	strDeviceName = name
	log.Printf("SetDeviceName , device name:[%s]", strDeviceName)
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

func GoCopyImage(fileSize rtkCommon.FileSize, imgHeader rtkCommon.ImgHeader, data []byte) {
	callbackInstanceCopyImageCB(fileSize, imgHeader, data)
}

func GoPasteImage() {
	callbackInstancePasteImageCB()
}

func GoFileDropRequest(id string, fileInfo rtkCommon.FileInfo, timestamp uint64) {
	callbackInstanceFileDropRequestCB(id, fileInfo, timestamp)
}

func GoFileDropResponse(id string, fileCmd rtkCommon.FileDropCmd, fileName string) {
	callbackInstanceFileDropResponseCB(id, fileCmd, fileName)
}

func GoMultiFilesDropRequest(id string, fileList *[]rtkCommon.FileInfo, folderList *[]string, totalSize, timestamp uint64, totalDesc string) {
	if callbackFileListDropRequestCB == nil {
		log.Println("CallbackFileListDropRequestCB is null!")
		return
	}
	callbackFileListDropRequestCB(id, *fileList, *folderList, totalSize, timestamp, totalDesc)
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

func GoSetupFileDrop(ip, id, fileName, platform string, fileSize uint64, timestamp uint64) {
	log.Printf("(DST) GoSetupFileDrop  source:%s ip:[%s]fileName:%s  fileSize:%d", id, ip, fileName, fileSize)
	CallbackInstance.CallbackMethodFileConfirm(id, platform, fileName, int64(fileSize))
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

func GoDragFileNotify(ip, id, fileName, platform string, fileSize uint64, timestamp uint64) {
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	log.Printf("(DST) GoDragFileNotify  source:%s ip:[%s]fileName:%s  fileSize:%d", id, ip, fileName, fileSize)
	CallbackInstance.CallbackFileDragNotify(id, platform, fileName, int64(fileSize))
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
	log.Printf("(DST) GoDragFileListFolderNotify  source:%s ip:[%s]  folderName:[%s]  timestamp:%d", id, ip, folderName, timestamp)
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

func GoSetupDstPasteImage(desc string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {
	log.Printf("GoSetupDstPasteImage from ID %s, len:[%d] dataSize:[%d]\n\n", desc, len(content), dataSize)
	imageData.Reset()
	imageData.Grow(int(dataSize))
	callbackInstancePasteImageCB()
}

func GoDataTransfer(data []byte) {
	imageData.Write(data)
}

func GoUpdateProgressBar(ip, id string, fileSize, sentSize, timestamp uint64, filePath string) {
	fileName := filepath.Base(filePath)
	//log.Printf("GoUpdateProgressBar ip:[%s] name:[%s] recvSize:[%d] total:[%d]", ip, fileName, sentSize, fileSize)
	CallbackInstance.CallbackUpdateProgressBar(ip, id, fileName, int64(sentSize), int64(fileSize), int64(timestamp))
}

func GoUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	if CallbackInstance == nil {
		log.Println("CallbackUpdateMultipleProgressBar CallbackInstance is null !")
		return
	}
	//log.Printf("GoUpdateMultipleProgressBar ip:[%s] [%s] currentFileName:[%s] recvSize:[%d] total:[%d]", ip, deviceName, currentFileName, sentSize, totalSize)
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

func GoEventHandle(eventType rtkCommon.EventType, id, fileName string) {
	if CallbackInstance == nil {
		log.Println("GoEventHandle CallbackInstance is null !")
		return
	}
	if eventType == rtkCommon.EVENT_TYPE_OPEN_FILE_ERR {
		strErr := "file datatransfer sender error"
		CallbackInstance.CallbackFileError(id, fileName, strErr)
	} else if eventType == rtkCommon.EVENT_TYPE_RECV_TIMEOUT {
		strErr := "file datatransfer receiving end error"
		CallbackInstance.CallbackFileError(id, fileName, strErr)
	}
	log.Printf("[%s %d]: id:%s, name:%s, error:%d", rtkMisc.GetFuncName(), rtkMisc.GetLine(), id, fileName, eventType)
}

func GoCleanClipboard() {
}

func GoSetupDstPasteText(content []byte) {
	log.Printf("GoSetupDstPasteText:%s \n\n", string(content))
	if CallbackInstance == nil {
		log.Println("GoSetupDstPasteText - failed - callbackInstance is nil")
		return
	}
	CallbackInstance.CallbackMethod(string(content))
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
	return rtkGlobal.PlatformAndroid
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

func GoAuthViaIndex(clientIndex uint32) {

}

func GoReqSourceAndPort() {

}

func GoGetSrcAndPortFromIni() rtkMisc.SourcePort {
	return rtkUtils.GetDeviceSrcPort()
}

func GoDIASStatusNotify(diasStatus uint32) {
	currentDiasStatus = diasStatus
	log.Printf("[%s] diasStatus:%d", rtkMisc.GetFuncInfo(), currentDiasStatus)
	if CallbackInstance == nil {
		log.Println("GoDIASStatusNotify - failed - callbackInstance is nil")
		return
	}
	CallbackInstance.CallbackUpdateDiasStatus(int(diasStatus))
}

func GoGetDIASStatus() uint32 {
	log.Printf("[%s] status=%d", rtkMisc.GetFuncInfo(), currentDiasStatus)
	return currentDiasStatus
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
