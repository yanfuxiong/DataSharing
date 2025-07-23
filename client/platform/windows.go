//go:build windows

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
	strDeviceName            string
	ifConfirmDocumentsAccept bool
)

func InitPlatform(rootPath, downLoadPath, deviceName string) {
	downloadPath = downLoadPath
	if deviceName == "" {
		strDeviceName = "UnknownDeviceName"
	} else {
		strDeviceName = deviceName
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

	log.Printf("[%s] init rootPath:[%s] downLoadPath:[%s], deviceName:[%s] success!", rtkMisc.GetFuncInfo(), rootPath, downLoadPath, strDeviceName)
}

type (
	CallbackNetworkSwitchFunc          func()
	CallbackCopyImageFunc              func(rtkCommon.FileSize, rtkCommon.ImgHeader, []byte)
	CallbackSetupDstImageFunc          func(id string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32)
	CallbackPasteImageFunc             func()
	CallbackDataTransferFunc           func(data []byte)
	CallbackCleanClipboardFunc         func()
	CallbackFileDropRequestFunc        func(string, rtkCommon.FileInfo, uint64)
	CallbackFileListDropRequestFunc    func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackDragFileListRequestFunc    func([]rtkCommon.FileInfo, []string, uint64, uint64, string)
	CallbackDragFileListNotifyFunc     func(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64)
	CallbackMultiFilesDropNotifyFunc   func(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64)
	CallbacMultipleProgressBarFunc     func(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64)
	CallbacNotiMessageFileTransFunc    func(fileName, clientName, platform string, timestamp uint64, isSender bool)
	CallbackFileDropResponseFunc       func(string, rtkCommon.FileDropCmd, string)
	CallbackFileDragInfoFunc           func(rtkCommon.FileInfo, uint64)
	CallbackCancelFileTransFunc        func(string, string, uint64)
	CallbackExtractDIASFunc            func()
	CallbackGetMacAddressFunc          func(string)
	CallbackAuthStatusCodeFunc         func(uint8)
	CallbackDIASSourceAndPortFunc      func(uint8, uint8)
	CallbackMethodBrowseMdnsResultFunc func(string, string, int)
	CallbackAuthViaIndexCallbackFunc   func(uint32)
	CallbackDIASStatusFunc             func(uint32)
	CallbackDeviceNameFunc             func() string
	CallbackDownloadPathFunc           func() string
	CallbackReqSourceAndPortFunc       func()
	CallbackUpdateSystemInfoFunc       func(ipAddr string, verInfo string)
	CallbackUpdateClientStatusFunc     func(status uint32, ip, id, deviceName, deviceType string)
	CallbackDetectPluginEventFunc      func(isPlugin bool, productName string)
)

var (
	// Go business Callback
	callbackNetworkSwitchCB            CallbackNetworkSwitchFunc          = nil
	callbackInstanceCopyImageCB        CallbackCopyImageFunc              = nil
	callbackInstancePasteImageCB       CallbackPasteImageFunc             = nil
	callbackInstanceFileDropRequestCB  CallbackFileDropRequestFunc        = nil
	callbackFileListDropRequestCB      CallbackFileListDropRequestFunc    = nil
	callbackDragFileListRequestCB      CallbackDragFileListRequestFunc    = nil
	callbackDragFileListNotifyCB       CallbackDragFileListNotifyFunc     = nil
	callbackMultiFilesDropNotifyCB     CallbackMultiFilesDropNotifyFunc   = nil
	callbacMultipleProgressBarCB       CallbacMultipleProgressBarFunc     = nil
	callbacNotiMessageFileTransCB      CallbacNotiMessageFileTransFunc    = nil
	callbackInstanceFileDropResponseCB CallbackFileDropResponseFunc       = nil
	callbackInstanceFileDragCB         CallbackFileDragInfoFunc           = nil
	callbackCancelFileTransDragCB      CallbackCancelFileTransFunc        = nil
	callbackExtractDIASCB              CallbackExtractDIASFunc            = nil
	callbackGetMacAddressCB            CallbackGetMacAddressFunc          = nil
	callbackAuthStatusCodeCB           CallbackAuthStatusCodeFunc         = nil
	callbackDIASSourceAndPortCB        CallbackDIASSourceAndPortFunc      = nil
	callbackMethodBrowseMdnsResult     CallbackMethodBrowseMdnsResultFunc = nil

	// main.go Callback
	callbackAuthViaIndex       CallbackAuthViaIndexCallbackFunc = nil
	callbackDIASStatus         CallbackDIASStatusFunc           = nil
	callbackReqSourceAndPort   CallbackReqSourceAndPortFunc     = nil
	callbackDeviceName         CallbackDeviceNameFunc           = nil
	callbackDownloadPath       CallbackDownloadPathFunc         = nil
	callbackUpdateSystemInfo   CallbackUpdateSystemInfoFunc     = nil
	callbackUpdateClientStatus CallbackUpdateClientStatusFunc   = nil
	callbackSetupDstImageCB    CallbackSetupDstImageFunc        = nil
	callbackDataTransferCB     CallbackDataTransferFunc         = nil
	callbackCleanClipboard     CallbackCleanClipboardFunc       = nil
)

func getDeviceName() string {
	if callbackDeviceName == nil {
		log.Fatalf("callbackDeviceName is nil!")
	}

	return callbackDeviceName()
}

func getDownloadPathInternal() string {
	if callbackDownloadPath == nil {
		log.Fatalf("callbackDownloadPath is nil!")
	}

	return callbackDownloadPath()
}

/*======================================= Used  by GO set Callback =======================================*/

func SetGoNetworkSwitchCallback(cb CallbackNetworkSwitchFunc) {
	callbackNetworkSwitchCB = cb
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

func SetGoFileDropRequestCallback(cb CallbackFileDropRequestFunc) {
	callbackInstanceFileDropRequestCB = cb
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

func SetGoDragFileCallback(cb CallbackFileDragInfoFunc) {
	callbackInstanceFileDragCB = cb
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

/*======================================= Used by main.go, set Callback =======================================*/

func SetAuthViaIndexCallback(cb CallbackAuthViaIndexCallbackFunc) {
	callbackAuthViaIndex = cb
}

func SetDIASStatusCallback(cb CallbackDIASStatusFunc) {
	callbackDIASStatus = cb
}

func SetDeviceNameCallback(cb CallbackDeviceNameFunc) {
	callbackDeviceName = cb
}

func SetDownloadPathCallback(cb CallbackDownloadPathFunc) {
	callbackDownloadPath = cb
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

func SetMultipleProgressBarCallback(cb CallbacMultipleProgressBarFunc) {
	callbacMultipleProgressBarCB = cb
}

func SetDragFileListNotifyCallback(cb CallbackDragFileListNotifyFunc) {
	callbackDragFileListNotifyCB = cb
}

func SetMultiFilesDropNotifyCallback(cb CallbackMultiFilesDropNotifyFunc) {
	callbackMultiFilesDropNotifyCB = cb
}

func SetNotiMessageFileTransCallback(cb CallbacNotiMessageFileTransFunc) {
	callbacNotiMessageFileTransCB = cb
}

/*======================================= Used by main.go, Called by C++ =======================================*/

func GoCopyImage(fileSize rtkCommon.FileSize, imgHeader rtkCommon.ImgHeader, data []byte) {
	callbackInstanceCopyImageCB(fileSize, imgHeader, data)
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

func GoMultiFilesDropRequest(id string, fileList *[]rtkCommon.FileInfo, folderList *[]string, totalSize, timestamp uint64, totalDesc string) {
	if callbackFileListDropRequestCB == nil {
		log.Println("CallbackFileListDropRequestCB is null!")
		return
	}
	callbackFileListDropRequestCB(id, *fileList, *folderList, totalSize, timestamp, totalDesc)
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

func GoSetupFileDrop(ip, id, fileName, platform string, fileSize uint64, timestamp uint64) {

}

func GoSetupFileListDrop(ip, id, platform, totalDesc string, fileCount, folderCount uint32, timestamp uint64) {
	log.Printf("[%s] fileCnt:[%d] folderCnt:[%d] totalDesc:[%s]", rtkMisc.GetFuncInfo(), fileCount, folderCount, totalDesc)

}

func GoMultiFilesDropNotify(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64) {
	callbackMultiFilesDropNotifyCB(ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
}

func GoDragFileNotify(ip, id, fileName, platform string, fileSize uint64, timestamp uint64) {

}

func GoDragFileListNotify(ip, id, platform string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64) {
	callbackDragFileListNotifyCB(ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
}

func GoDragFileListFolderNotify(ip, id, folderName string, timestamp uint64) {
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

func GoUpdateProgressBar(ip, id string, fileSize, sentSize, timestamp uint64, fileName string) {

}

func GoUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	callbacMultipleProgressBarCB(ip, id, deviceName, currentFileName, sentFileCnt, totalFileCnt, currentFileSize, totalSize, sentSize, timestamp)
}

func GoUpdateSystemInfo(ipAddr, serviceVer string) {
	log.Printf("[%s] Ip:[%s]  version[%s]", rtkMisc.GetFuncInfo(), ipAddr, serviceVer)
	callbackUpdateSystemInfo(ipAddr, serviceVer)
}

func GoUpdateClientStatus(status uint32, ip, id, name, deviceType string) {
	callbackUpdateClientStatus(status, ip, id, name, deviceType)
}

func GoNotiMessageFileTransfer(fileName, clientName, platform string, timestamp uint64, isSender bool) {
	callbacNotiMessageFileTransCB(fileName, clientName, platform, timestamp, isSender)
}

func GoEventHandle(eventType rtkCommon.EventType, ipAddr, fileName string) {

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

func GoGetDeviceName() string {
	return strDeviceName
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

// Specific Platform: iOS. Browse and lookup MDNS from iOS
func GoStartBrowseMdns(instance, serviceType string) {
}

func GoStopBrowseMdns() {
}
