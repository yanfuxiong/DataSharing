//go:build android

package platform

import (
	"context"
	"log"
	"os"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	rtkUtils "rtk-cross-share/utils"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
)

const (
	hostID      = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/ID.HostID"
	nodeID      = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/ID.ID"
	receiveFile = "/storage/emulated/0/Android/data/com.rtk.myapplication/files/"
)

var ImageData []byte
var curInputText string

// Notify to Android APK
type Callback interface {
	CallbackMethod(result string)
	CallbackMethodImage(content string)
	LogMessageCallback(msg string)
	EventCallback(event int)
	CallbackMethodFileConfirm(ipAddr, platform, filename string, fileSize int64)
	CallbackMethodFileDone(name string, fileSize int64)
	CallbackMethodFoundPeer()
	CallbackUpdateProgressBar(ipAddr, filename string, recvSize, total int64)
	CallbackFileError(ipAddr, filename, err string)
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

// Android only
func GoCopyImage(fileSize rtkCommon.FileSize, imgHeader rtkCommon.ImgHeader, data []byte) {
	CallbackInstanceCopyImageCB(fileSize, imgHeader, data)
}

func GoPasteImage() { //Android  have no Paste
	CallbackInstancePasteImageCB()
}

func GoFileDropRequest(ip string, fileInfo rtkCommon.FileInfo, timestamp int64) {
	CallbackInstanceFileDropRequestCB(ip, fileInfo, timestamp)
}

func GoFileDropResponse(ip string, fileCmd rtkCommon.FileDropCmd, fileName string) {
	CallbackInstanceFileDropResponseCB(ip, fileCmd, fileName)
}

func WatchClipboardText(ctx context.Context, resultChan chan<- string) {
	var lastInputText string // this must be local variable

	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(100 * time.Millisecond):
			if len(curInputText) > 0 && !strings.EqualFold(curInputText, lastInputText) {
				log.Println("watchClipboardText - got new message: ", curInputText)
				lastInputText = curInputText
				resultChan <- curInputText
			}
		}
	}
}

func SendMessage(s string) {
	log.Printf("SendMessage:[%s] ", s)
	if s == "" || len(s) == 0 {
		return
	}
	curInputText = s
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

func GoSetupFileDrop(ipAddr string, desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {
	fileSize := int64(fileSizeHigh)<<32 | int64(fileSizeLow)
	log.Printf("(DST) GoSetupFileDrop  source:%s ip:[%s]fileName:%s  fileSize:%d", desc, ipAddr, fileName, fileSize)
	CallbackInstance.CallbackMethodFileConfirm(ipAddr, platform, fileName, fileSize)
}

func ReceiveImageCopyDataDone(ImageSize int64, imgHeader rtkCommon.ImgHeader) {
	log.Printf("[%s %d]: size:%d, (width, height):(%d,%d)", rtkUtils.GetFuncName(), rtkUtils.GetLine(), ImageSize, imgHeader.Width, imgHeader.Height)
	if CallbackInstance == nil {
		log.Println(" CallbackInstance is null !")
		return
	}
	rtkUtils.GoSafe(func() {
		imageBase64 := rtkUtils.Base64Encode(rtkUtils.BitmapToImage(ImageData, int(imgHeader.Width), int(imgHeader.Height)))
		// log.Printf("len[%d][%d][%d][%+v]", len(ImageData), len(imageBase64), rtkGlobal.Handler.CopyImgHeader.Width, imageBase64)
		CallbackInstance.CallbackMethodImage(imageBase64)
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
	ImageData = []byte{}
	CallbackInstancePasteImageCB()
}

func GoDataTransfer(data []byte) {
	ImageData = append(ImageData, data...)
}

func GoUpdateProgressBar(ipAddr, filename string, size int, recvSize, totalSize int64) {
	//log.Printf("GoUpdateProgressBar ip:[%s] name:[%s] recvSize:[%d] total:[%d]", ipAddr, filename, recvSize, totalSize)
	CallbackInstance.CallbackUpdateProgressBar(ipAddr, filename, recvSize, totalSize)
}

func GoDeinitProgressBar() {

}

func GoUpdateClientStatus(status uint32, ip string, id string, name string) {

}

func GoEventHandle(eventType rtkCommon.EventType, ipAddr, fileName string) {
	if CallbackInstance == nil {
		log.Println("GoEventHandle CallbackInstance is null !")
		return
	}
	if eventType == rtkCommon.EVENT_TYPE_OPEN_FILE_ERR {
		strErr := "file datatransfer sender error"
		CallbackInstance.CallbackFileError(ipAddr, fileName, strErr)
	} else if eventType == rtkCommon.EVENT_TYPE_RECV_TIMEOUT {
		strErr := "file datatransfer receiving end error"
		CallbackInstance.CallbackFileError(ipAddr, fileName, strErr)
	}
	log.Printf("[%s %d]: ipAddr:%s, name:%s, error:%d", rtkUtils.GetFuncName(), rtkUtils.GetLine(), ipAddr, fileName, eventType)
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
	return "android"
}

func GetReceiveFilePath() string {
	return receiveFile
}

func GetMdnsPortConfigPath() string {
	return ""
}
