//go:build windows
// +build windows

package platform

/*
#cgo LDFLAGS: -L../../clipboard/libs -lclipboard -Wl,-Bstatic
#cgo LDFLAGS: -lgdi32 -lole32 -luuid -lstdc++ -pthread
#include <stdlib.h>
#include <wchar.h>
#include "../../clipboard/MSPaste/MSCommonExt.h"

typedef void (*ClipboardCopyFileCallback)(wchar_t*, unsigned long, unsigned long);
typedef void (*ClipboardPasteFileCallback)(char*);
typedef void (*ClipboardCopyImgCallback)(IMAGE_HEADER, unsigned char*, unsigned long);

extern void SetClipboardCopyFileCallback(ClipboardCopyFileCallback callback);
extern void SetClipboardPasteFileCallback(ClipboardPasteFileCallback callback);
extern void SetClipboardCopyImgCallback(ClipboardCopyImgCallback callback);

extern void StartClipboardMonitor();
extern void StopClipboardMonitor();

extern void SetupDstPasteFile(wchar_t* desc, wchar_t* fileName, unsigned long fileSizeHigh, unsigned long fileSizeLow);
extern void SetupFileDrop(char* ip, char* id, unsigned long long fileSize, unsigned long long timestamp, wchar_t* fileName);
extern void SetupDstPasteImage(wchar_t* desc, IMAGE_HEADER imgHeader, unsigned long dataSize);
extern void DataTransfer(unsigned char* data, unsigned int size);
extern void UpdateProgressBar(char* ip, char* id, unsigned long long fileSize, unsigned long long sentSize, unsigned long long timestamp, wchar_t* fileName);
extern void DeinitProgressBar();
extern void UpdateClientStatus(unsigned int status, char* ip, char* id, wchar_t* name);
extern void EventHandle(EVENT_TYPE event);
extern void CleanClipboard();

void clipboardCopyFileCallback(wchar_t* content, unsigned long, unsigned long);
void clipboardPasteFileCallback(char* content);
void fileDropCmdCallback(char*, unsigned long, wchar_t*);
void clipboardCopyImgCallback(IMAGE_HEADER, unsigned char*, unsigned long);

// Pipe
typedef void (*FileDropRequestCallback)(char*, char*, unsigned long long, unsigned long long, wchar_t*);
typedef void (*FileDropResponseCallback)(int, char*, char*, unsigned long long, unsigned long long, wchar_t*);
typedef void (*PipeConnectedCallback)(void);
extern void StartPipeMonitor();
extern void StopPipeMonitor();
extern void SetFileDropRequestCallback(FileDropRequestCallback callback);
extern void SetFileDropResponseCallback(FileDropResponseCallback callback);
extern void SetPipeConnectedCallback(PipeConnectedCallback callback);
void fileDropRequestCallback(char*, char*, unsigned long long, unsigned long long, wchar_t*);
void fileDropResponseCallback(int, char*, char*, unsigned long long, unsigned long long, wchar_t*);
void pipeConnectedCallback(void);
*/
import "C"
import (
	"context"
	"log"
	"fmt"
	"os"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	rtkUtils "rtk-cross-share/utils"
	"unsafe"

	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.design/x/clipboard"
	"golang.org/x/sys/windows"
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
	CallbackMethodFileConfirm(ip, platform, fileName string, fileSize int64)
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

// download path
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

// Monitor
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
			log.Println("DEBUG: watchClipboardText - got new message: ", string(contentText))
			/*
				if CallbackInstanceResetCB != nil {
					CallbackInstanceResetCB(rtkCommon.CLIPBOARD_RESET_TYPE_TEXT)
				}
			*/
			resultChan <- string(contentText)
		}
	}
}

func wcharToString(wstr *C.wchar_t) string {
	var goStr string
	for ptr := wstr; *ptr != 0; ptr = (*C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + unsafe.Sizeof(*ptr))) {
		goStr += string(rune(*ptr))
	}
	return goStr
}

func stringToWChar(str string) *C.wchar_t {
	utf16Str := make([]uint16, len(str)+1)
	for i, r := range str {
		utf16Str[i] = uint16(r)
	}
	utf16Str[len(str)] = 0

	size := len(utf16Str) * int(unsafe.Sizeof(utf16Str[0]))
	cStr := C.malloc(C.size_t(size))
	if cStr == nil {
		panic("C.malloc failed")
	}

	C.memcpy(cStr, unsafe.Pointer(&utf16Str[0]), C.size_t(size))
	return (*C.wchar_t)(cStr)
}

//export clipboardCopyFileCallback
func clipboardCopyFileCallback(cContent *C.wchar_t, cFileSizeHigh C.ulong, cFileSizeLow C.ulong) {
	content := wcharToString(cContent)
	fileSizeHigh := uint32(cFileSizeHigh)
	fileSizeLow := uint32(cFileSizeLow)
	log.Println("Clipboard file content:", content, "fileSize high:", fileSizeHigh, "low:", fileSizeLow)
}

// For DEBUG
func GoClipboardPasteFileCallback(content string) {
	cContent := C.CString(content)
	defer C.free(unsafe.Pointer(cContent))
	clipboardPasteFileCallback(cContent)
}

//export clipboardPasteFileCallback
func clipboardPasteFileCallback(cContent *C.char) {
	if CallbackInstancePasteImageCB == nil {
		return
	}
	content := C.GoString(cContent)
	CallbackInstancePasteImageCB()
	log.Println("Paste Clipboard file content:", content)
}

//export fileDropRequestCallback
func fileDropRequestCallback(cIp *C.char, cId *C.char, cFileSize C.ulonglong, cTimestamp C.ulonglong, cFilePath *C.wchar_t) {
	if CallbackInstanceFileDropRequestCB == nil {
		return
	}
	ip := C.GoString(cIp)
	fileSize := uint64(cFileSize)
	fileSizeHigh := uint32(fileSize >> 32)
	fileSizeLow := uint32(fileSize & 0xFFFFFFFF)
	var fileInfo = rtkCommon.FileInfo{
		FileSize_: rtkCommon.FileSize{
			SizeHigh: uint32(fileSizeHigh),
			SizeLow:  uint32(fileSizeLow),
		},
		FilePath: wcharToString(cFilePath),
	}
	timestamp := int64(cTimestamp)
	log.Printf("[%s %d] ip[%s] path:[%s]", rtkUtils.GetFuncName(), rtkUtils.GetLine(), ip, fileInfo.FilePath)
	CallbackInstanceFileDropRequestCB(ip, fileInfo, timestamp)
}

//export fileDropResponseCallback
func fileDropResponseCallback(cStatus int32, cIp *C.char, cId *C.char, cFileSize C.ulonglong, cTimestamp C.ulonglong, cFilePath *C.wchar_t) {
	if CallbackInstanceFileDropResponseCB == nil {
		return
	}

	ip := C.GoString(cIp)
	if cStatus == 0 { // FILE_DROP_REJECT
		log.Println("FILE_DROP_REJECT")
		CallbackInstanceFileDropResponseCB(ip, rtkCommon.FILE_DROP_REJECT, "")
	} else if cStatus == 1 { // FILE_DROP_ACCEPT
		path := wcharToString(cFilePath)
		log.Printf("FILE_DROP_ACCEPT, ip:[%s] path:[%s]", ip, path)
		CallbackInstanceFileDropResponseCB(ip, rtkCommon.FILE_DROP_ACCEPT, path)
	}
}

//export clipboardCopyImgCallback
func clipboardCopyImgCallback(cHeader C.IMAGE_HEADER, cData *C.uchar, cDataSize C.ulong) {
	if CallbackInstanceCopyImageCB == nil {
		return
	}
	imgHeader := rtkCommon.ImgHeader{
		Width:       int32(cHeader.width),
		Height:      int32(cHeader.height),
		Planes:      uint16(cHeader.planes),
		BitCount:    uint16(cHeader.bitCount),
		Compression: uint32(cHeader.compression),
	}
	data := C.GoBytes(unsafe.Pointer(cData), C.int(cDataSize))
	dataSize := uint32(cDataSize)
	// FIXME
	filesize := rtkCommon.FileSize{
		SizeHigh: 0,
		SizeLow:  dataSize,
	}
	CallbackInstanceCopyImageCB(filesize, imgHeader, data)
	log.Printf("Clipboard image content, width[%d] height[%d] data size[%d] \n", imgHeader.Width, imgHeader.Height, dataSize)
}

//export pipeConnectedCallback
func pipeConnectedCallback() {
	if CallbackPipeConnectedCB == nil {
		return
	}
	CallbackPipeConnectedCB()
}

// export SetupDstPasteFile
func GoSetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {
	cDesc := stringToWChar(desc)
	defer C.free(unsafe.Pointer(cDesc))
	cFileName := stringToWChar(fileName)
	defer C.free(unsafe.Pointer(cFileName))
	cFileSizeHigh := C.ulong(fileSizeHigh)
	cFileSizeLow := C.ulong(fileSizeLow)
	C.SetupDstPasteFile(cDesc, cFileName, cFileSizeHigh, cFileSizeLow)
}

// export SetupFileDrop
func GoSetupFileDrop(ip, id, fileName, platform string, fileSize uint64, timestamp int64) {
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))
	cFileSize := C.ulonglong(fileSize)
	cTimestamp := C.ulonglong(timestamp)
	cFileName := stringToWChar(fileName)
	defer C.free(unsafe.Pointer(cFileName))
	C.SetupFileDrop(cIp, cId, cFileSize, cTimestamp, cFileName)
}

// export SetupDstPasteImage
func GoSetupDstPasteImage(desc string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {
	cDesc := stringToWChar(desc)
	defer C.free(unsafe.Pointer(cDesc))
	cImgHeader := C.IMAGE_HEADER{
		width:       C.int(imgHeader.Width),
		height:      C.int(imgHeader.Height),
		planes:      C.ushort(imgHeader.Planes),
		bitCount:    C.ushort(imgHeader.BitCount),
		compression: C.ulong(imgHeader.Compression),
	}
	cDataSize := C.ulong(dataSize)
	C.SetupDstPasteImage(cDesc, cImgHeader, cDataSize)
}

// export DataTransfer
func GoDataTransfer(data []byte) {
	cData := (*C.uchar)(unsafe.Pointer(&data[0]))
	cSize := C.uint(len(data))
	C.DataTransfer(cData, cSize)
}

// export UpdateProgressBar
func GoUpdateProgressBar(ip, id string, fileSize, sentSize uint64, timestamp int64, fileName string) {
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))
	cFileSize := C.ulonglong(fileSize)
	cSentSize := C.ulonglong(sentSize)
	cTimestamp := C.ulonglong(timestamp)
	cName := stringToWChar(fileName)
	defer C.free(unsafe.Pointer(cName))
	C.UpdateProgressBar(cIp, cId, cFileSize, cSentSize, cTimestamp, cName)
}

// export DeinitProgressBar
func GoDeinitProgressBar() {
	C.DeinitProgressBar()
}

// export UpdateClientStatus
func GoUpdateClientStatus(status uint32, ip string, id string, name string) {
	cStatus := C.uint(status)
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))
	cName := stringToWChar(name)
	defer C.free(unsafe.Pointer(cName))
	C.UpdateClientStatus(cStatus, cIp, cId, cName)
}

// export EventHandle
func GoEventHandle(eventType rtkCommon.EventType, ipAddr, fileName string) {
	C.EventHandle(C.EVENT_TYPE(eventType))
}

// export CleanClipboard
func GoCleanClipboard() {
	C.CleanClipboard()
}

func SetupCallbackSettings() {
	C.SetClipboardCopyFileCallback((C.ClipboardCopyFileCallback)(unsafe.Pointer(C.clipboardCopyFileCallback)))
	C.SetClipboardPasteFileCallback((C.ClipboardPasteFileCallback)(unsafe.Pointer(C.clipboardPasteFileCallback)))
	C.SetFileDropRequestCallback((C.FileDropRequestCallback)(unsafe.Pointer(C.fileDropRequestCallback)))
	C.SetFileDropResponseCallback((C.FileDropResponseCallback)(unsafe.Pointer(C.fileDropResponseCallback)))
	C.SetClipboardCopyImgCallback((C.ClipboardCopyImgCallback)(unsafe.Pointer(C.clipboardCopyImgCallback)))
	C.SetPipeConnectedCallback((C.PipeConnectedCallback)(unsafe.Pointer(C.pipeConnectedCallback)))
	C.StartClipboardMonitor()
	C.StartPipeMonitor()
}

func GoSetupDstPasteText(content []byte) {
	log.Println("GoSetupDstPasteText :", string(content))
	clipboard.Write(clipboard.FmtText, content)
}

func ReceiveFileConfirm(fileSize int64) {
	log.Println("ReceiveFileConfirm:", fileSize)
}

func ReceiveImageCopyDataDone(fileSize int64, imgHeader rtkCommon.ImgHeader) {
	log.Println("ReceiveImageCopyDataDone size:", fileSize, " imgHeader: ", imgHeader)
}

func ReceiveFileDropCopyDataDone(fileSize int64, dstFilePath string) {
	log.Println("ReceiveFileDropCopyDataDone size:", fileSize, " dstFilePath: ", dstFilePath)
}

func FoundPeer() {
	log.Println("CallbackMethodFoundPeer")
}

func GetClientList() string {
	clientList := rtkUtils.GetClientList()
	log.Printf("GetClientList :[%s]", clientList)
	return clientList
}

func GenKey() crypto.PrivKey {
	privKeyFile := ".priv.pem"
	return rtkUtils.GenKey(privKeyFile)
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

func GetIDPath() string {
	return ".ID"
}

func GetHostIDPath() string {
	return ".HostID"
}

func GetPlatform() string {
	return "windows"
}

func GetMdnsPortConfigPath() string {
	return ".MdnsPort"
}

func GetDeviceTablePath() string {
	return ".DeviceTable"
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
