//go:build darwin

package platform

import (
	"context"
	"log"
	"os"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	rtkUtils "rtk-cross-share/utils"

	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.design/x/clipboard"
)

const (
	logFile = "p2p.log"
)

func GetLogFilePath() string {
	return logFile
}

type Callback interface {
	CallbackMethod(result string)
	CallbackMethodImage(content []byte)
	LogMessageCallback(msg string)
	EventCallback(event int)
	CallbackMethodFileConfirm(id, platform, fileName string, fileSize int64)
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

/*
type CallbackFunc func(rtkCommon.ClipboardResetType)

var CallbackInstanceResetCB CallbackFunc = nil

func SetResetCBCallback(cb CallbackFunc) {
	CallbackInstanceResetCB = cb
}
*/

func WatchClipboardText(ctx context.Context, resultChan chan<- string) {
	changeText := clipboard.Watch(ctx, clipboard.FmtText)
	for {
		select {
		case <-ctx.Done():
			return
		case contentText := <-changeText:
			if string(contentText) == "" || len(contentText) == 0 {
				continue
			}
			log.Println("watchClipboardText - got new message: ", string(contentText))
			/*
				if CallbackInstanceResetCB != nil {
					CallbackInstanceResetCB(rtkCommon.CLIPBOARD_RESET_TYPE_TEXT)
				}
			*/
			resultChan <- string(contentText)
		}
	}
}

func SetupCallbackSettings() {

}

func GoClipboardPasteFileCallback(content string) {

}

func GoSetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {

}

func GoSetupFileDrop(ip, id, fileName, platform string, fileSize uint64, timestamp int64) {

}

func GoSetupDstPasteImage(desc string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {

}

func GoDataTransfer(data []byte) {

}

func GoUpdateProgressBar(ip, id string, fileSize, sentSize uint64, timestamp uint64, fileName string) {

}

func GoDeinitProgressBar() {

}

func GoUpdateSystemInfo(ip, serviceVer string) {

}

func GoUpdateClientStatus(status uint32, ip string, id string, name string) {

}

func GoEventHandle(eventType rtkCommon.EventType, id, fileName string) {

}

func GoCleanClipboard() {

}

func GoSetupDstPasteText(content []byte) {
	clipboard.Write(clipboard.FmtText, content)
}

func ReceiveFileConfirm(fileSize int64) {

}

func ReceiveImageCopyDataDone(fileSize int64, imgHeader rtkCommon.ImgHeader) {
	log.Println("ReceiveImageCopyDataDone size:", fileSize, " imgHeader: ", imgHeader)
}

func ReceiveFileDropCopyDataDone(fileSize int64, dstFilePath string) {
	log.Println("ReceiveFileDropCopyDataDone size:", fileSize, " dstFilePath: ", dstFilePath)
}

func FoundPeer() {

}

func GenKey() crypto.PrivKey {
	privKeyFile := ".priv.pem"
	return rtkUtils.GenKey(privKeyFile)
}

func GetIDPath() string {
	return ".ID"
}

func GetHostIDPath() string {
	return ".HostID"
}

func IsHost() bool {
	return rtkUtils.FileExists(".HostID")
}

func GetHostID() string {
	file, err := os.Open(".HostID")
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

func GetPlatform() string {
	return rtkGlobal.PlatformMac
}

func GetMdnsPortConfigPath() string {
	return ".MdnsPort"
}

func SetNetWorkConnected(bConnected bool) {
}

func GetNetWorkConnected() bool {
	return false
}

// Deprecated: replace with GetDeviceInfoPath
func GetDeviceTablePath() string {
	return ".DeviceTable"
}

func GetDeviceInfoPath() string {
	return ".DeviceInfo"
}

func LockFile(file *os.File) error {
	return nil
}

func UnlockFile(file *os.File) error {
	return nil
}
