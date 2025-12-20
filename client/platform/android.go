//go:build android

package platform

import (
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"log"
	"os"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"syscall"
	"time"
)

type NOTI_MSG_CODE int

const (
	NOTI_MSG_CODE_CONN_STATUS_SUCCESS NOTI_MSG_CODE = iota + 1
	NOTI_MSG_CODE_CONN_STATUS_FAIL
	NOTI_MSG_CODE_FILE_TRANS_DONE_SENDER
	NOTI_MSG_CODE_FILE_TRANS_DONE_RECEIVER
	NOTI_MSG_CODE_FILE_TRANS_REJECT
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

// Notify to Android APK
type Callback interface {
	CallbackPasteXClipData(text, image, html, rtf string)
	LogMessageCallback(msg string)
	CallbackMethodFileConfirm(id, platform, filename string, fileSize int64)
	CallbackFileListSendNotify(ip, id string, fileCnt int, totalSize, timestamp int64, firstFileName string, firstFileSize int64, fileDetails string)
	CallbackFileListReceiveNotify(ip, id string, fileCnt int, totalSize, timestamp int64, firstFileName string, firstFileSize int64, fileDetails string)
	CallbackFileListDragFolderNotify(ip, id, folderName string, timestamp int64)
	CallbackSendFilesDone(filesInfo, platform, deviceName string, timestamp int64)
	CallbackNotiMessage(filesInfo, platform, deviceName string, notiCode, onlineCnt int, timestamp int64)
	CallbackUpdateClientStatus(clientInfo string)
	CallbackUpdateSendProgressBar(ip, id, currentFileName string, sendFileCnt, totalFileCnt int, currentFileSize, totalSize, sendSize, timestamp int64)
	CallbackUpdateReceiveProgressBar(ip, id, currentFileName string, recvFileCnt, totalFileCnt int, currentFileSize, totalSize, recvSize, timestamp int64)
	CallbackNotifyErrEvent(id string, errCode int, arg1, arg2, arg3, arg4 string)
	CallbackUpdateDiasStatus(status int)
	CallbackGetAuthData(clientIndex int) string
	CallbackUpdateMonitorName(monitorName string)
	CallbackRequestUpdateClientVersion(clienVersion string)
	CallbackNotifyBrowseResult(monitorName, instance, ipAddr, version string, timestamp int64)
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

type (
	CallbackNetworkSwitchFunc          func()
	CallbackCopyXClipFunc              func(cbText, cbImage, cbHtml, cbRtf []byte)
	CallbackFileDropResponseFunc       func(string, rtkCommon.FileDropCmd, string)
	CallbackFileListDropRequestFunc    func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string, string)
	CallbackDragFileListRequestFunc    func([]rtkCommon.FileInfo, []string, uint64, uint64, string, string)
	CallbackGetMacAddressFunc          func(string)
	CallbackCancelFileTransFunc        func(string, string, uint64)
	CallbackDetectPluginEventFunc      func(isPlugin bool, productName string)
	CallbackDIASSourceAndPortFunc      func(uint8, uint8)
	CallbackAuthStatusCodeFunc         func(uint8)
	CallbackExtractDIASFunc            func()
	CallbackMethodBrowseMdnsResultFunc func(string, string, int, string, string, string, string)
	CallbackGetFilesTransCodeFunc      func(id string) rtkCommon.SendFilesRequestErrCode
	CallbackGetFilesSendCacheCountFunc func(id string) int
	CallbackConnectLanServerFunc       func(instance string)
	CallbackBrowseLanServerFunc        func()
	CallbackSetMsgEventFunc            func(event uint32, arg1, arg2, arg3, arg4 string)
)

var (
	callbackNetworkSwitchCB            CallbackNetworkSwitchFunc          = nil
	callbackXClipCopyCB                CallbackCopyXClipFunc              = nil
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
	callbackGetFilesSendCacheCount     CallbackGetFilesSendCacheCountFunc = nil
	callbackConnectLanServer           CallbackConnectLanServerFunc       = nil
	callbackBrowseLanServer            CallbackBrowseLanServerFunc        = nil
	callbackSetMsgEvent                CallbackSetMsgEventFunc            = nil
)

func SetGoNetworkSwitchCallback(cb CallbackNetworkSwitchFunc) {
	callbackNetworkSwitchCB = cb
}

// Notify to Clipboard/FileDrop
func SetCopyXClipCallback(cb CallbackCopyXClipFunc) {
	callbackXClipCopyCB = cb
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

func SetGetFilesCacheSendCountCallback(cb CallbackGetFilesSendCacheCountFunc) {
	callbackGetFilesSendCacheCount = cb
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
	rtkMisc.GoSafe(func() { callbackSetMsgEvent(event, arg1, arg2, arg3, arg4) })
}

func SetDeviceName(name string) {
	rtkGlobal.NodeInfo.DeviceName = name
}

func GoTriggerNetworkSwitch() {
}

func GoSetHostListenAddr(listenHost string, listenPort int) {
	if listenHost == "" || listenHost == rtkMisc.DefaultIp || listenHost == rtkMisc.LoopBackIp || listenPort <= rtkGlobal.DefaultPort {
		return
	}
	if rtkGlobal.ListenHost != rtkMisc.DefaultIp &&
		rtkGlobal.ListenHost != "" &&
		rtkGlobal.ListenPort != rtkGlobal.DefaultPort &&
		(listenHost != rtkGlobal.ListenHost || listenPort != rtkGlobal.ListenPort) {
		log.Printf("[%s] The previous host Addr:[%s:%d], new host Addr:[%s:%d] ", rtkMisc.GetFuncInfo(), rtkGlobal.ListenHost, rtkGlobal.ListenPort, listenHost, listenPort)
		log.Println("**************** Attention please, the host listen addr is switch! ********************\n\n")
		rtkGlobal.ListenHost = listenHost
		rtkGlobal.ListenPort = listenPort
		callbackNetworkSwitchCB()
	}
}

func GoCopyXClipData(text, image, html, rtf string) {
	if callbackXClipCopyCB == nil {
		log.Println("callbackXClipCopyCB is null!")
		return
	}

	imgData := []byte(nil)
	if image != "" {
		startTime := time.Now().UnixMilli()
		data := rtkUtils.Base64Decode(image)
		if data == nil {
			return
		}

		format, width, height := rtkUtils.GetByteImageInfo(data)
		jpegData, err := rtkUtils.ImageToJpeg(format, data)
		if err != nil {
			return
		}
		if len(jpegData) == 0 {
			log.Printf("[CopyXClip] Error: jpeg data is empty")
			return
		}

		imgData = jpegData
		log.Printf("image get jpg size:[%d](%d,%d),decode use:[%d]ms", len(imgData), width, height, time.Now().UnixMilli()-startTime)
	}

	callbackXClipCopyCB([]byte(text), imgData, []byte(html), []byte(rtf))
}

func GoFileDropResponse(id string, fileCmd rtkCommon.FileDropCmd, fileName string) {
	callbackInstanceFileDropResponseCB(id, fileCmd, fileName)
}

func GoMultiFilesDropRequest(filesDataInfoJson string) rtkCommon.SendFilesRequestErrCode {
	if callbackFileListDropRequestCB == nil {
		log.Println("CallbackFileListDropRequestCB is null!")
		return rtkCommon.SendFilesRequestCallbackNotSet
	}

	if callbackGetFilesSendCacheCount == nil {
		log.Println("callbackGetFilesSendCacheCount is null!")
		return rtkCommon.SendFilesRequestCallbackNotSet
	}

	var fileDataInfo rtkCommon.FilesDataRequestInfo
	err := json.Unmarshal([]byte(filesDataInfoJson), &fileDataInfo)
	if err != nil {
		log.Printf("[%s] Unmarshal[%s] err:%+v", rtkMisc.GetFuncInfo(), filesDataInfoJson, err)
		return rtkCommon.SendFilesRequestParameterErr
	}
	log.Printf("[%s] ID:[%s] IPAddr:[%s] len:[%d] timestamp:[%d]", rtkMisc.GetFuncInfo(), fileDataInfo.Id, fileDataInfo.Ip, len(fileDataInfo.PathList), fileDataInfo.TimeStamp)

	fileList := make([]rtkCommon.FileInfo, 0)
	folderList := make([]string, 0)
	totalSize := uint64(0)

	for _, file := range fileDataInfo.PathList {
		if rtkMisc.FolderExists(file) { //unused,  Android does not support folder sharing
			rtkUtils.WalkPath(file, &folderList, &fileList, &totalSize)
		} else if rtkMisc.FileExists(file) {
			fileSize, sizeErr := rtkMisc.FileSize(file)
			if sizeErr != nil {
				log.Printf("[%s] get file:[%s] size error:%+v, skit it!", rtkMisc.GetFuncInfo(), file, sizeErr)
				continue
			}
			fileList = append(fileList, rtkCommon.FileInfo{
				FileSize_: rtkCommon.FileSize{
					SizeHigh: uint32(fileSize >> 32),
					SizeLow:  uint32(fileSize & 0xFFFFFFFF),
				},
				FilePath: file,
				FileName: filepath.Base(file),
			})
			totalSize += fileSize
		} else {
			log.Printf("[%s] get file or path:[%s] is not exist , so skit it!", rtkMisc.GetFuncInfo(), file)
		}
	}
	totalDesc := rtkMisc.FileSizeDesc(totalSize)

	if fileDataInfo.TimeStamp == 0 {
		fileDataInfo.TimeStamp = time.Now().UnixMilli()
	}
	log.Printf("[%s] ID[%s] IP:[%s] get file count:[%d] folder count:[%d], totalSize:[%d] totalDesc:[%s] timestamp:[%d]", rtkMisc.GetFuncInfo(), fileDataInfo.Id, fileDataInfo.Ip, len(fileList), len(folderList), totalSize, totalDesc, fileDataInfo.TimeStamp)

	if len(fileList) == 0 && len(folderList) == 0 {
		log.Println("file content is null!")
		return rtkCommon.SendFilesRequestParameterErr
	}

	if !rtkUtils.GetPeerClientIsSupportQueueTrans(fileDataInfo.Id) {
		if callbackGetFilesTransCode == nil {
			log.Println("callbackGetFilesTransCode is null!")
			return rtkCommon.SendFilesRequestCallbackNotSet
		}

		filesTransCode := callbackGetFilesTransCode(fileDataInfo.Id)
		if filesTransCode != rtkCommon.SendFilesRequestSuccess {
			return filesTransCode
		}
	}

	nCacheCount := callbackGetFilesSendCacheCount(fileDataInfo.Id)
	if nCacheCount >= rtkGlobal.SendFilesRequestMaxQueueSize {
		log.Printf("[%s] ID[%s] this user file drop cache count:[%d] is too large and over range !", rtkMisc.GetFuncInfo(), fileDataInfo.Id, nCacheCount)
		return rtkCommon.SendFilesRequestCacheOverRange
	}

	/*if totalSize > uint64(rtkGlobal.SendFilesRequestMaxSize) {
		log.Printf("[%s] ID[%s] this file drop total size:[%d] [%s] is too large and over range !", rtkMisc.GetFuncInfo(), fileDataInfo.Id, totalSize, totalDesc)
		return rtkCommon.SendFilesRequestSizeOverRange
	}*/

	nMsgLength := int(rtkGlobal.P2PMsgMagicLength)

	for _, file := range fileList {
		nMsgLength = nMsgLength + len(file.FileName) + rtkGlobal.FileInfoMagicLength
	}

	for _, folder := range folderList {
		nMsgLength = nMsgLength + len(folder) + rtkGlobal.StringArrayMagicLength
	}

	if nMsgLength >= rtkGlobal.P2PMsgMaxLength {
		log.Printf("[%s] ID[%s] file count:[%d] folder count:[%d], the p2p message is too long and over range!", rtkMisc.GetFuncInfo(), fileDataInfo.Id, len(fileList), len(folderList))
		return rtkCommon.SendFilesRequestLengthOverRange
	}

	callbackFileListDropRequestCB(fileDataInfo.Id, fileList, folderList, totalSize, uint64(fileDataInfo.TimeStamp), totalDesc, "")
	return rtkCommon.SendFilesRequestSuccess
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

func GoConnectLanServer(instance string) {
	if callbackConnectLanServer == nil {
		log.Println("callbackConnectLanServer is null!")
		return
	}

	rtkMisc.GoSafe(func() { callbackConnectLanServer(instance) })
}

func GoBrowseLanServer() {
	if callbackBrowseLanServer == nil {
		log.Println("callbackBrowseLanServer is null!")
		return
	}

	rtkMisc.GoSafe(func() { callbackBrowseLanServer() })
}

/***************** Used  by GO business *****************/

func GoSetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {
	fileSize := int64(fileSizeHigh)<<32 | int64(fileSizeLow)
	log.Printf("(DST) GoSetupDstPasteFile  sourceID:%s fileName:[%s] fileSize:[%d]", desc, fileName, fileSize)
	CallbackInstance.CallbackMethodFileConfirm("", platform, fileName, fileSize)
}

func GoSetupFileListDrop(ip, id, platform, totalDesc string, fileCount, folderCount uint32, timestamp uint64) {
	log.Printf("(DST) GoSetupFileListDrop  ID:]%s] IP:[%s] totalDesc:%s  fileCount:%d  folderCount:%d", id, ip, totalDesc, fileCount, folderCount)
}

func GoFileListSendNotify(ip, id string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64, fileDetails string) {
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	log.Printf("(SRC) GoFileListSendNotify  dst:%s ip:[%s] timestamp:%d fileCnt:%d  totalSize:%d", id, ip, timestamp, fileCnt, totalSize)
	CallbackInstance.CallbackFileListSendNotify(ip, id, int(fileCnt), int64(totalSize), int64(timestamp), firstFileName, int64(firstFileSize), fileDetails)
}

func GoFileListReceiveNotify(ip, id string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64, fileDetails string) {
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	log.Printf("(DST) GoFileListReceiveNotify  src:%s ip:[%s] timestamp:%d fileCnt:%d  totalSize:%d", id, ip, timestamp, fileCnt, totalSize)
	CallbackInstance.CallbackFileListReceiveNotify(ip, id, int(fileCnt), int64(totalSize), int64(timestamp), firstFileName, int64(firstFileSize), fileDetails)
}

func GoDragFileListFolderNotify(ip, id, folderName string, timestamp uint64) {
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	log.Printf("(DST) GoDragFileListFolderNotify source:%s ip:[%s] folder:[%s] timestamp:%d", id, ip, folderName, timestamp)
	CallbackInstance.CallbackFileListDragFolderNotify(ip, id, folderName, int64(timestamp))
}

func GoUpdateClientStatusEx(id string, status uint8) {
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
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

	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), string(encodedData))
	CallbackInstance.CallbackUpdateClientStatus(string(encodedData))
}

func GoSetupDstPasteXClipData(cbText, cbImage, cbHtml, cbRtf []byte) {
	if CallbackInstance == nil {
		log.Println("GoSetupDstPasteXClipData failed - callbackInstance is nil")
		return
	}
	log.Printf("[%s] text:%d , image:%d, html:%d, rtf:%d \n\n", rtkMisc.GetFuncInfo(), len(cbText), len(cbImage), len(cbHtml), len(cbRtf))

	imageStr := ""
	if len(cbImage) > 0 {
		imageBase64 := rtkUtils.Base64Encode(cbImage)
		imageStr = imageBase64
	}

	CallbackInstance.CallbackPasteXClipData(string(cbText), imageStr, string(cbHtml), string(cbRtf))
}

func GoUpdateSendProgressBar(ip, id, currentFileName string, sendFileCnt, totalFileCnt uint32, currentFileSize, totalFileSize, sendSize, timestamp uint64) {
	if CallbackInstance == nil {
		log.Println("CallbackUpdateSendProgressBar CallbackInstance is null !")
		return
	}
	//log.Printf("GoUpdateSendProgressBar ip:[%s] [%s] currentFileName:[%s] sendSize:[%d] total:[%d] timestamp:[%d]", ip, deviceName, currentFileName, sendSize, totalFileSize, timestamp)
	CallbackInstance.CallbackUpdateSendProgressBar(ip, id, currentFileName, int(sendFileCnt), int(totalFileCnt), int64(currentFileSize), int64(totalFileSize), int64(sendSize), int64(timestamp))
}

func GoUpdateReceiveProgressBar(ip, id, currentFileName string, recvFileCnt, totalFileCnt uint32, currentFileSize, totalFileSize, recvSize, timestamp uint64) {
	if CallbackInstance == nil {
		log.Println("CallbackUpdateReceiveProgressBar CallbackInstance is null !")
		return
	}
	//log.Printf("GoUpdateReceiveProgressBar ip:[%s] [%s] currentFileName:[%s] recvSize:[%d] total:[%d] timestamp:[%d]", ip, deviceName, currentFileName, sentSize, totalSize, timestamp)
	CallbackInstance.CallbackUpdateReceiveProgressBar(ip, id, currentFileName, int(recvFileCnt), int(totalFileCnt), int64(currentFileSize), int64(totalFileSize), int64(recvSize), int64(timestamp))
}

func GoUpdateSystemInfo(ip, serviceVer string) {

}

func GoNotiMessageFileTransfer(fileInfo, clientName, platform string, timestamp uint64, isSender bool) {
	log.Printf("[%s]: fileInfo:[%s], clientName:%s, timestamp:%d ", rtkMisc.GetFuncInfo(), fileInfo, clientName, timestamp)
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}

	var code NOTI_MSG_CODE
	if isSender {
		code = NOTI_MSG_CODE_FILE_TRANS_DONE_SENDER
		CallbackInstance.CallbackSendFilesDone(fileInfo, platform, clientName, int64(timestamp))
	} else {
		code = NOTI_MSG_CODE_FILE_TRANS_DONE_RECEIVER
	}

	CallbackInstance.CallbackNotiMessage(fileInfo, platform, clientName, int(code), 0, int64(timestamp))
}

func GoNotifyErrEvent(id string, errCode rtkMisc.CrossShareErr, arg1, arg2, arg3, arg4 string) {
	if CallbackInstance == nil {
		log.Println("GoNotifyErrEvent CallbackInstance is null !")
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

func LockFile() error {
	var err error
	lockFd, err = os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Failed to open or create lock file:[%s] err:%+v", lockFile, err)
		return err
	}

	err = syscall.Flock(int(lockFd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		log.Printf("Failed to lock file[%s] err:%+v", lockFile, err)
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

func GetAuthData(clientIndex uint32) (rtkMisc.CrossShareErr, rtkMisc.AuthDataInfo) {
	if CallbackInstance == nil {
		log.Println("GetAuthData - failed - callbackInstance is nil")
		return rtkMisc.ERR_BIZ_GET_CALLBACK_INSTANCE_NULL, rtkMisc.AuthDataInfo{}
	}

	authDataJsonInfo := CallbackInstance.CallbackGetAuthData(int(clientIndex))

	var authData rtkMisc.AuthDataInfo
	err := json.Unmarshal([]byte(authDataJsonInfo), &authData)
	if err != nil {
		log.Printf("[%s] Unmarshal[%s] err:%+v", rtkMisc.GetFuncInfo(), authDataJsonInfo, err)
		return rtkMisc.ERR_BIZ_JSON_UNMARSHAL, rtkMisc.AuthDataInfo{}
	}

	log.Printf("[%s] width:[%d] height:[%d] Framerate:[%d] type:[%d] DisplayName:[%s] Center[%d,%d]", rtkMisc.GetFuncInfo(), authData.Width, authData.Height, authData.Framerate, authData.Type, authData.DisplayName, authData.CenterX, authData.CenterY)
	return rtkMisc.SUCCESS, authData
}

// Specific Platform: iOS. Browse and lookup MDNS from iOS
func GoStartBrowseMdns(instance, serviceType string) {
}

func GoStopBrowseMdns() {
}

func GoSetupAppLink(link string) {
	rtkMisc.AppLink = link
}
