//go:build android

package platform

import (
	"bytes"
	"context"
	"github.com/libp2p/go-libp2p/core/crypto"
	"log"
	"os"
	"path/filepath"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	rtkUtils "rtk-cross-share/utils"
	"strings"
)

const (
	hostID       = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/ID.HostID"
	nodeID       = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/ID.ID"
	receiveFile  = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/"
	logFile      = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/p2p.log"
	crashLogFile = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/crash.log"

	// Deprecated: replace with deviceInfo
	deviceTable = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/ID.DeviceTable"
	deviceInfo  = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/ID.DeviceInfo"
)

var (
	imageData          bytes.Buffer
	copyTextChan       = make(chan string, 100)
	isNetWorkConnected bool
)

func GetLogFilePath() string {
	return logFile
}

func GetCrashLogFilePath() string {
	return crashLogFile
}

// Notify to Android APK
type Callback interface {
	CallbackMethod(result string)
	CallbackMethodImage(content string)
	LogMessageCallback(msg string)
	EventCallback(event int)
	CallbackMethodFileConfirm(id, platform, filename string, fileSize int64)
	CallbackMethodFileDone(name string, fileSize int64)
	CallbackMethodFoundPeer()
	CallbackUpdateProgressBar(id, filename string, recvSize, total int64)
	CallbackFileError(id, filename, err string)
}

var CallbackInstance Callback = nil

func SetCallback(cb Callback) {
	CallbackInstance = cb
}

// Notify to Clipboard/FileDrop
type CallbackCopyImageFunc func(rtkCommon.FileSize, rtkCommon.ImgHeader, []byte)

var CallbackInstanceCopyImageCB CallbackCopyImageFunc = nil

func SetCopyImageCallback(cb CallbackCopyImageFunc) {
	CallbackInstanceCopyImageCB = cb
}

type CallbackPasteImageFunc func()

var CallbackInstancePasteImageCB CallbackPasteImageFunc = nil

func SetPasteImageCallback(cb CallbackPasteImageFunc) {
	CallbackInstancePasteImageCB = cb
}

type CallbackFileDropRequestFunc func(string, rtkCommon.FileInfo, int64)

var CallbackInstanceFileDropRequestCB CallbackFileDropRequestFunc = nil

func SetGoFileDropRequestCallback(cb CallbackFileDropRequestFunc) {
	CallbackInstanceFileDropRequestCB = cb
}

type CallbackFileDropResponseFunc func(string, rtkCommon.FileDropCmd, string)

var CallbackInstanceFileDropResponseCB CallbackFileDropResponseFunc = nil

func SetGoFileDropResponseCallback(cb CallbackFileDropResponseFunc) {
	CallbackInstanceFileDropResponseCB = cb
}

// TODO: replace with GetClientList
type CallbackPipeConnectedFunc func()

var CallbackPipeConnectedCB CallbackPipeConnectedFunc = nil

func SetGoPipeConnectedCallback(cb CallbackPipeConnectedFunc) {
	CallbackPipeConnectedCB = cb
}

// Android only
func GoCopyImage(fileSize rtkCommon.FileSize, imgHeader rtkCommon.ImgHeader, data []byte) {
	CallbackInstanceCopyImageCB(fileSize, imgHeader, data)
}

func GoPasteImage() {
	CallbackInstancePasteImageCB()
}

func GoFileDropRequest(id string, fileInfo rtkCommon.FileInfo, timestamp int64) {
	CallbackInstanceFileDropRequestCB(id, fileInfo, timestamp)
}

func GoFileDropResponse(id string, fileCmd rtkCommon.FileDropCmd, fileName string) {
	CallbackInstanceFileDropResponseCB(id, fileCmd, fileName)
}

func WatchClipboardText(ctx context.Context, resultChan chan<- string) {
	var lastCopyText string
	for {
		select {
		case <-ctx.Done():
			close(resultChan)
			return

		case curCopyText := <-copyTextChan:
			if len(curCopyText) > 0 && !strings.EqualFold(curCopyText, lastCopyText) {
				log.Println("DEBUG: watchClipboardText - got new message: ", curCopyText)
				lastCopyText = curCopyText
				resultChan <- curCopyText
			}
		}
	}
}

func SendMessage(strText string) {
	log.Printf("SendMessage:[%s] ", strText)
	if strText == "" || len(strText) == 0 {
		return
	}

	for i := 0; i < rtkUtils.GetClientCount(); i++ {
		copyTextChan <- strText
	}
}

func SetupCallbackSettings() {

}

func GoClipboardPasteFileCallback(content string) {

}

func GoSetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {
	fileSize := int64(fileSizeHigh)<<32 | int64(fileSizeLow)
	log.Printf("(DST) GoSetupDstPasteFile  sourceID:%s fileName:[%s] fileSize:[%d]", desc, fileName, fileSize)
	CallbackInstance.CallbackMethodFileConfirm("", platform, fileName, fileSize)
}

func GoSetupFileDrop(ip, id, fileName, platform string, fileSize uint64, timestamp int64) {
	log.Printf("(DST) GoSetupFileDrop  source:%s ip:[%s]fileName:%s  fileSize:%d", id, ip, fileName, fileSize)
	CallbackInstance.CallbackMethodFileConfirm(id, platform, fileName, int64(fileSize))
}

func ReceiveImageCopyDataDone(fileSize int64, imgHeader rtkCommon.ImgHeader) {
	log.Printf("[%s %d]: size:%d, (width, height):(%d,%d)", rtkUtils.GetFuncName(), rtkUtils.GetLine(), fileSize, imgHeader.Width, imgHeader.Height)
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	rtkUtils.GoSafe(func() {
		imageBase64 := rtkUtils.Base64Encode(imageData.Bytes())
		// log.Printf("len[%d][%d][%d][%+v]", len(ImageData), len(imageBase64), rtkGlobal.Handler.CopyImgHeader.Width, imageBase64)
		CallbackInstance.CallbackMethodImage(imageBase64)
		imageData.Reset()
	})
}

func ReceiveFileDropCopyDataDone(fileSize int64, dstFilePath string) {
	log.Printf("[%s %d]: size:%d, dstFilePath:%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), fileSize, dstFilePath)
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	CallbackInstance.CallbackMethodFileDone(dstFilePath, fileSize)
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
	CallbackInstancePasteImageCB()
}

func GoDataTransfer(data []byte) {
	imageData.Write(data)
}

func GoUpdateProgressBar(ip, id string, fileSize, sentSize uint64, timestamp int64, filePath string) {
	fileName := filepath.Base(filePath)
	log.Printf("GoUpdateProgressBar ip:[%s] name:[%s] recvSize:[%d] total:[%d]", ip, fileName, sentSize, fileSize)
	CallbackInstance.CallbackUpdateProgressBar(id, fileName, int64(sentSize), int64(fileSize))
}

func GoDeinitProgressBar() {

}

func GoUpdateSystemInfo(ip, serviceVer string) {

}

func GoUpdateClientStatus(status uint32, ip string, id string, name string) {

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
	log.Printf("[%s %d]: id:%s, name:%s, error:%d", rtkUtils.GetFuncName(), rtkUtils.GetLine(), id, fileName, eventType)
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
	privKeyFile := "/storage/emulated/0/Android/data/com.rtk.myapplication/files/priv.pem"
	return rtkUtils.GenKey(privKeyFile)
}

func IsHost() bool {
	return rtkUtils.FileExists(hostID)
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

func GetReceiveFilePath() string {
	return receiveFile
}

func GetMdnsPortConfigPath() string {
	return ""
}

// Deprecated: replace with GetDeviceInfoPath
func GetDeviceTablePath() string {
	return deviceTable
}

func GetDeviceInfoPath() string {
	return deviceInfo
}

func LockFile(file *os.File) error {
	return nil
}

func UnlockFile(file *os.File) error {
	return nil
}

func SetNetWorkConnected(bConnected bool) {
	isNetWorkConnected = bConnected
}

func GetNetWorkConnected() bool {
	return isNetWorkConnected
}
