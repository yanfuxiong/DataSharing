//go:build windows
// +build windows

package platform

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.design/x/clipboard"
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
	CallbackXClipCopyFunc              func(cbText, cbImage, cbHtml []byte)
	CallbackCopyImageFunc              func(rtkCommon.ImgHeader, []byte)
	CallbackSetupDstImageFunc          func(id string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32)
	CallbackPasteImageFunc             func()
	CallbackDataTransferFunc           func(data []byte)
	CallbackCleanClipboardFunc         func()
	CallbackFileListDropRequestFunc    func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackDragFileListRequestFunc    func([]rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackDragFileListNotifyFunc     func(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64)
	CallbackMultiFilesDropNotifyFunc   func(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64)
	CallbackMultipleProgressBarFunc    func(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64)
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
	CallbackDetectPluginEventFunc      func(isPlugin bool, productName string)
	CallbackGetFilesTransCodeFunc      func(id string) rtkCommon.SendFilesRequestErrCode
	CallbackReqClientUpdateVerFunc     func(clientVer string)
	CallbackNotifyErrEventFunc         func(id string, errCode uint32, arg1, arg2, arg3, arg4 string)
	CallbackConnectLanServerFunc       func(instance string)
	CallbackBrowseLanServerFunc        func()
	CallbackSetMsgEventFunc            func(event uint32, arg1, arg2, arg3, arg4 string)
)

var (
	// Go business Callback
	callbackNetworkSwitchCB            CallbackNetworkSwitchFunc          = nil
	callbackXClipCopyCB                CallbackXClipCopyFunc              = nil
	callbackInstanceCopyImageCB        CallbackCopyImageFunc              = nil
	callbackInstancePasteImageCB       CallbackPasteImageFunc             = nil
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
	callbackAuthViaIndex       CallbackAuthViaIndexCallbackFunc = nil
	callbackDIASStatus         CallbackDIASStatusFunc           = nil
	callbackReqSourceAndPort   CallbackReqSourceAndPortFunc     = nil
	callbackUpdateSystemInfo   CallbackUpdateSystemInfoFunc     = nil
	callbackUpdateClientStatus CallbackUpdateClientStatusFunc   = nil
	callbackSetupDstImageCB    CallbackSetupDstImageFunc        = nil
	callbackDataTransferCB     CallbackDataTransferFunc         = nil
	callbackCleanClipboard     CallbackCleanClipboardFunc       = nil
	callbackGetFilesTransCode  CallbackGetFilesTransCodeFunc    = nil
)

/*======================================= Used  by GO set Callback =======================================*/

func SetGoNetworkSwitchCallback(cb CallbackNetworkSwitchFunc) {
	callbackNetworkSwitchCB = cb
}

func SetCopyXClipCallback(cb CallbackXClipCopyFunc) {
	callbackXClipCopyCB = cb
}

func SetCopyImageCallback(cb CallbackCopyImageFunc) {
	callbackInstanceCopyImageCB = cb
}

func SetPasteImageCallback(cb CallbackPasteImageFunc) {
	callbackInstancePasteImageCB = cb
}

func SetImageDataTransferCallback(cb CallbackDataTransferFunc) {
	callbackDataTransferCB = cb
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

func SetGoConnectLanServerCallback(cb CallbackConnectLanServerFunc) {
}

func SetGoBrowseLanServerCallback(cb CallbackBrowseLanServerFunc) {

}

func SetGoSetMsgEventCallback(cb CallbackSetMsgEventFunc) {
	callbackSetMsgEvent = cb
}

/*======================================= Used by main.go, set Callback =======================================*/

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

func SetUpdateClientStatusCallback(cb CallbackUpdateClientStatusFunc) {
	callbackUpdateClientStatus = cb
}

func SetSetupDstImageCallback(cb CallbackSetupDstImageFunc) {
	callbackSetupDstImageCB = cb
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

func GoCopyXClipData(text, image, html string) {

}

func GoCopyImage(imgHeader rtkCommon.ImgHeader, data []byte) {
	callbackInstanceCopyImageCB(imgHeader, data)
}

// Deprecated: unused
func GoPasteImage() {
	callbackInstancePasteImageCB()
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

	callbackFileListDropRequestCB(id, *fileList, *folderList, totalSize, timestamp, totalDesc)
	return filesTransCode
}

func GoDragFileListRequest(fileList *[]rtkCommon.FileInfo, folderList *[]string, totalSize, timestamp uint64, totalDesc string) {
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

func GetLockFilePath() string {
	return lockFile
}

// Monitor
func WatchClipboardText(ctx context.Context, resultChan chan<- string) {
	changeText := clipboard.Watch(ctx, clipboard.FmtText)
	isLastClipboardCopyText := false

	for {
		select {
		case <-ctx.Done():
			close(resultChan)
			return
		case <-chNotifyPasteText:
			isLastClipboardCopyText = true
		case contentText := <-changeText:
			/*if string(contentText) == "" || len(contentText) == 0 {
				continue
			}*/
			curClipboardCopyText := string(contentText) //we can copy a null text from local

			if !isLastClipboardCopyText { // we can copy same text from local
				log.Println("DEBUG: watchClipboardText - got new message: ", curClipboardCopyText)
				resultChan <- curClipboardCopyText
			} else {
				isLastClipboardCopyText = false
			}

		}
	}
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

func GetAuthData() (rtkMisc.CrossShareErr, rtkMisc.AuthDataInfo) {
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

}

func GoSetupDstPasteImage(id string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {
	// TODO: consider setup JPG image to windows C++
	bmpSize := uint32(imgHeader.Height) * uint32(imgHeader.Width) * uint32(imgHeader.BitCount) / 8
	log.Printf("GoSetupDstPasteImage from ID %s, len:[%d] dataSize:[%d] bmpSize:[%d]\n\n", id, len(content), dataSize, bmpSize)

	callbackSetupDstImageCB(id, content, imgHeader, bmpSize)
	callbackInstancePasteImageCB()
}

func GoDataTransfer(data []byte) {
	// TODO: avoid to convert to BMP here, move to C++ partition
	startConvertTime := time.Now().UnixNano()
	bmpData, err := rtkUtils.JpgToBmp(data)
	log.Printf("(DST) Convert jpg to bmp, size:[%d] use [%d] ms...", len(bmpData), (time.Now().UnixNano()-startConvertTime)/1e6)
	if err != nil {
		log.Printf("(DST) Err: Convert JPG to BMP failed")
		return
	}

	callbackDataTransferCB(bmpData)
}

func ReceiveImageCopyDataDone(fileSize int64, imgHeader rtkCommon.ImgHeader) {

}

func GoUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	callbackMultipleProgressBarCB(ip, id, deviceName, currentFileName, sentFileCnt, totalFileCnt, currentFileSize, totalSize, sentSize, timestamp)
}

func GoUpdateSystemInfo(ipAddr, serviceVer string) {
	log.Printf("[%s] Ip:[%s]  version[%s]", rtkMisc.GetFuncInfo(), ipAddr, serviceVer)
	callbackUpdateSystemInfo(ipAddr, serviceVer)
}

func GoUpdateClientStatus(status uint32, ip, id, name, deviceType string) {
	callbackUpdateClientStatus(status, ip, id, name, deviceType)
}

func GoNotiMessageFileTransfer(fileName, clientName, platform string, timestamp uint64, isSender bool) {
	callbackNotiMessageFileTransCB(fileName, clientName, platform, timestamp, isSender)
}

func GoEventHandle(eventType rtkCommon.EventType, ipAddr, fileName string, timestamp uint64) {

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

func GoSetupDstPasteText(content []byte) {
	log.Println("GoSetupDstPasteText :", string(content))

	nClientCount := rtkUtils.GetClientCount()
	for i := 0; i < nClientCount; i++ {
		chNotifyPasteText <- struct{}{}
	}
	time.Sleep(10 * time.Millisecond)
	clipboard.Write(clipboard.FmtText, content)
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

func LockFile(file *os.File) error {
	handle := windows.Handle(file.Fd())
	if handle == windows.InvalidHandle {
		return fmt.Errorf("invalid file handle")
	}

	var overlapped windows.Overlapped

	err := windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
	if err != nil {
		return fmt.Errorf("failed to lock file: %w", err)
	}

	return nil
}

func UnlockFile(file *os.File) error {
	handle := windows.Handle(file.Fd())
	if handle == windows.InvalidHandle {
		return fmt.Errorf("invalid file handle")
	}

	var overlapped windows.Overlapped

	err := windows.UnlockFileEx(handle, 0, 1, 0, &overlapped)
	if err != nil {
		return fmt.Errorf("failed to unlock file: %w", err)
	}

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
