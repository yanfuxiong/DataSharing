//go:build ignore

package platform

/*
#cgo LDFLAGS: -L../../clipboard/libs -lclipboard -Wl,-Bstatic
#cgo LDFLAGS: -lws2_32 -lgdi32 -ldxva2 -lole32 -lshell32 -luuid -lstdc++ -pthread
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
//extern void SetupFileListDrop(char* ip, char* id, char* cTotalDesc, unsigned int cFileCount, unsigned int cFolderCount, unsigned long long cTimestamp);
extern void MultiFilesDropNotify(char *ipPort,char *clientID,unsigned int cFileCount, unsigned long long totalSize, unsigned long long timestamp, wchar_t *firstFileName, unsigned long long firstFileSize);
extern void DragFileNotify(char* ip, char* id, unsigned long long fileSize, unsigned long long timestamp, wchar_t* fileName);
extern void DragFileListNotify(char* ip, char* id, unsigned int cFileCount, unsigned long long totalSize, unsigned long long timestamp,   wchar_t* firstFileName, unsigned long long firstFileSize);
extern void SetupDstPasteImage(wchar_t* desc, IMAGE_HEADER imgHeader, unsigned long dataSize);
extern void DataTransfer(unsigned char* data, unsigned int size);
extern void UpdateProgressBar(char* ip, char* id, unsigned long long fileSize, unsigned long long sentSize, unsigned long long timestamp, wchar_t* fileName);
extern void UpdateMultipleProgressBar(char* ip, char* id, wchar_t* curentFileName, unsigned int sentFileCnt, unsigned int totalFileCnt, unsigned long long curentFileSize, unsigned long long totalSize, unsigned long long sentSize, unsigned long long timestamp);
extern void UpdateImageProgressBar(char *ip, char *id, unsigned long long fileSize, unsigned long long sentSize, unsigned long long timestamp);
extern void DeinitProgressBar();
extern void UpdateSystemInfo(char* ip, wchar_t* serviceVer);
extern void UpdateClientStatus(unsigned int status, char* ip, char* id, wchar_t* name, char* deviceType);
extern void NotiMessage(unsigned long long timestamp, unsigned int notiCode, wchar_t *notiParam[], int paramCount);
extern void EventHandle(EVENT_TYPE event);
extern void CleanClipboard();
extern const wchar_t* GetDeviceName();
extern void AuthViaIndex(unsigned int index);
extern void RequestSourceAndPort();
extern const wchar_t* GetDownloadPath();
extern void CoTaskMemFree(void* pv);
extern void DIASStatus(unsigned int status);

void clipboardCopyFileCallback(wchar_t* content, unsigned long, unsigned long);
void clipboardPasteFileCallback(char* content);
void fileDropCmdCallback(char*, unsigned long, wchar_t*);
void clipboardCopyImgCallback(IMAGE_HEADER, unsigned char*, unsigned long);

// Pipe
typedef void (*FileDropRequestCallback)(char*, char*, unsigned long long, unsigned long long, wchar_t*);
typedef void (*FileDropResponseCallback)(int, char*, char*, unsigned long long, unsigned long long, wchar_t*);
typedef void (*MultiFilesDropRequestCallback)(char *ipPort, char *clientID, unsigned long long timeStamp, wchar_t *filePathArry[], unsigned int arryLength);
typedef void (*DragFileListRequestCallback)(wchar_t*[],unsigned int, unsigned long long);
typedef void (*DragFileCallback)(unsigned long long, wchar_t*);
typedef void (*CancelFileTransferCallback)(char*, char*, unsigned long long);
typedef void (*PipeConnectedCallback)(void);
typedef void (*GetMacAddressCallback)(char*, int);
typedef void (*ExtractDIASCallback)();
typedef void (*AuthStatusCodeCallback)(unsigned char);
typedef void (*DIASSourceAndPortCallback)(unsigned char, unsigned char);
extern void StartPipeMonitor();
extern void StopPipeMonitor();
extern void SetFileDropRequestCallback(FileDropRequestCallback callback);
extern void SetMultiFilesDropRequestCallback(MultiFilesDropRequestCallback callback);
extern void SetFileDropResponseCallback(FileDropResponseCallback callback);
extern void SetDragFileCallback(DragFileCallback callback);
extern void SetDragFileListRequestCallback(DragFileListRequestCallback callback);
extern void SetCancelFileTransferCallback(CancelFileTransferCallback callback);
extern void SetPipeConnectedCallback(PipeConnectedCallback callback);
extern void SetGetMacAddressCallback(GetMacAddressCallback callback);
extern void SetExtractDIASCallback(ExtractDIASCallback callback);
extern void SetAuthStatusCodeCallback(AuthStatusCodeCallback callback);
extern void SetDIASSourceAndPortCallback(DIASSourceAndPortCallback callback);
void fileDropRequestCallback(char*, char*, unsigned long long, unsigned long long, wchar_t*);
void fileDropResponseCallback(int, char*, char*, unsigned long long, unsigned long long, wchar_t*);
void multiFilesDropRequestCallback(char *ipPort, char *clientID, unsigned long long timeStamp, wchar_t *filePathArry[], unsigned int arryLength);
void dragFileListRequestCallback(wchar_t*[], unsigned int, unsigned long long);
void dragFileCallback(unsigned long long, wchar_t*);
void cancelFileTransferCallback(char*, char*, unsigned long long);
void pipeConnectedCallback(void);
void getMacAddressCallback(char*, int);
void extractDIASCallback();
void authStatusCodeCallback(unsigned char);
void diasSourceAndPortCallback(unsigned char, unsigned char);
*/
import "C"
import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"time"
	"unsafe"

	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.design/x/clipboard"
	"golang.org/x/sys/windows"
)

const (
	privKeyFile  = ".priv.pem"
	hostID       = ".HostID"
	nodeID       = ".ID"
	logFile      = "p2p.log"
	crashLogFile = "crash.log"
)

var (
	downLoadFilePath         = ""
	chNotifyPasteText        = make(chan struct{}, 100)
	ifConfirmDocumentsAccept bool
)

func init() {
	ifConfirmDocumentsAccept = false
	downLoadFilePath = GoGetDownloadPath()
}

func DownloadPath() string {
	return downLoadFilePath
}

func GetLogFilePath() string {
	return logFile
}

func GetCrashLogFilePath() string {
	return crashLogFile
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

type CallbackNetworkSwitchFunc func()

var callbackNetworkSwitchCB CallbackNetworkSwitchFunc

func SetGoNetworkSwitchCallback(cb CallbackNetworkSwitchFunc) {
	callbackNetworkSwitchCB = cb
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

type CallbackFileDropRequestFunc func(string, rtkCommon.FileInfo, uint64)

var CallbackInstanceFileDropRequestCB CallbackFileDropRequestFunc = nil

func SetGoFileDropRequestCallback(cb CallbackFileDropRequestFunc) {
	CallbackInstanceFileDropRequestCB = cb
}

type CallbackFileListDropRequestFunc func(string, []rtkCommon.FileInfo, []string, uint64, uint64, string)

var CallbackFileListDropRequestCB CallbackFileListDropRequestFunc = nil

func SetGoFileListDropRequestCallback(cb CallbackFileListDropRequestFunc) {
	CallbackFileListDropRequestCB = cb
}

type CallbackDragFileListRequestFunc func([]rtkCommon.FileInfo, []string, uint64, uint64, string)

var CallbackDragFileListRequestCB CallbackDragFileListRequestFunc = nil

func SetGoDragFileListRequestCallback(cb CallbackDragFileListRequestFunc) {
	CallbackDragFileListRequestCB = cb
}

// download path
type CallbackFileDropResponseFunc func(string, rtkCommon.FileDropCmd, string)

var CallbackInstanceFileDropResponseCB CallbackFileDropResponseFunc = nil

func SetGoFileDropResponseCallback(cb CallbackFileDropResponseFunc) {
	CallbackInstanceFileDropResponseCB = cb
}

type CallbackFileDragInfoFunc func(rtkCommon.FileInfo, uint64)

var CallbackInstanceFileDragCB CallbackFileDragInfoFunc = nil

func SetGoDragFileCallback(cb CallbackFileDragInfoFunc) {
	CallbackInstanceFileDragCB = cb
}

type CallbackCancelFileTransFunc func(string, string, uint64)

var CallbackCancelFileTransDragCB CallbackCancelFileTransFunc = nil

func SetGoCancelFileTransCallback(cb CallbackCancelFileTransFunc) {
	CallbackCancelFileTransDragCB = cb
}

// TODO: replace with GetClientList
type CallbackPipeConnectedFunc func()

var CallbackPipeConnectedCB CallbackPipeConnectedFunc = nil

func SetGoPipeConnectedCallback(cb CallbackPipeConnectedFunc) {
	CallbackPipeConnectedCB = cb
}

type CallbackExtractDIASFunc func()

var CallbackExtractDIASCB CallbackExtractDIASFunc = nil

func SetGoExtractDIASCallback(cb CallbackExtractDIASFunc) {
	CallbackExtractDIASCB = cb
}

type CallbackGetMacAddressFunc func(string)

var CallbackGetMacAddressCB CallbackGetMacAddressFunc = nil

func SetGoGetMacAddressCallback(cb CallbackGetMacAddressFunc) {
	CallbackGetMacAddressCB = cb
}

type CallbackAuthStatusCodeFunc func(uint8)

var CallbackAuthStatusCodeCB CallbackAuthStatusCodeFunc = nil

func SetGoAuthStatusCodeCallback(cb CallbackAuthStatusCodeFunc) {
	CallbackAuthStatusCodeCB = cb
}

type CallbackDIASSourceAndPortFunc func(uint8, uint8)

var CallbackDIASSourceAndPortCB CallbackDIASSourceAndPortFunc = nil

func SetGoDIASSourceAndPortCallback(cb CallbackDIASSourceAndPortFunc) {
	CallbackDIASSourceAndPortCB = cb
}

type CallbackMethodBrowseMdnsResultFunc func(string, string, int)

var CallbackMethodBrowseMdnsResult CallbackMethodBrowseMdnsResultFunc = nil

func SetGoBrowseMdnsResultCallback(cb CallbackMethodBrowseMdnsResultFunc) {
	CallbackMethodBrowseMdnsResult = cb
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

type WCharArray struct {
	Array **C.wchar_t
	items []*C.wchar_t
}

func NewWCharArray(strs []string) *WCharArray {
	ptrSize := unsafe.Sizeof(uintptr(0))
	cArray := C.malloc(C.size_t(len(strs)+1) * C.size_t(ptrSize))
	if cArray == nil {
		panic("C.malloc failed")
	}

	cStrs := make([]*C.wchar_t, len(strs))
	for i, s := range strs {
		cStr := stringToWChar(s)
		cStrs[i] = cStr
		elemPtr := (**C.wchar_t)(unsafe.Pointer(uintptr(cArray) + uintptr(i)*ptrSize))
		*elemPtr = cStr
	}

	endPtr := (**C.wchar_t)(unsafe.Pointer(uintptr(cArray) + uintptr(len(strs))*ptrSize))
	*endPtr = nil

	return &WCharArray{
		Array: (**C.wchar_t)(cArray),
		items: cStrs,
	}
}

func (w *WCharArray) Free() {
	for _, ptr := range w.items {
		C.free(unsafe.Pointer(ptr))
	}
	C.free(unsafe.Pointer(w.Array))
}

//export extractDIASCallback
func extractDIASCallback() {
	if CallbackExtractDIASCB == nil {
		log.Printf("[%s] CallbackExtractDIASCB is not set, CallbackExtractDIASCB failed! ", rtkMisc.GetFuncInfo())
		return
	}
	log.Printf("[%s] extractDIASCallback ", rtkMisc.GetFuncInfo())
	CallbackExtractDIASCB()
}

//export getMacAddressCallback
func getMacAddressCallback(cMac *C.char, nLen int32) {
	if CallbackGetMacAddressCB == nil {
		log.Printf("[%s] CallbackGetMacAddressCB is not set, CallbackGetMacAddressCB failed! ", rtkMisc.GetFuncInfo())
		return
	}

	if nLen != 6 {
		log.Printf("[%s] getMacAddressCallback failed, invalid MAC length:%d", rtkMisc.GetFuncInfo(), nLen)
		return
	}
	macBytes := C.GoBytes(unsafe.Pointer(cMac), 6)

	macAddress := fmt.Sprintf("%02X%02X%02X%02X%02X%02X",
		macBytes[0], macBytes[1], macBytes[2],
		macBytes[3], macBytes[4], macBytes[5])

	log.Printf("[%s] getMacAddressCallback [%s]", rtkMisc.GetFuncInfo(), macAddress)
	CallbackGetMacAddressCB(macAddress)
}

//export authStatusCodeCallback
func authStatusCodeCallback(cStatus C.uchar) {
	if CallbackAuthStatusCodeCB == nil {
		log.Printf("[%s] CallbackAuthStatusCodeCB is not set, CallbackAuthStatusCodeCB failed!", rtkMisc.GetFuncInfo())
		return
	}
	authStatus := uint8(cStatus)
	log.Printf("[%s] authStatusCodeCallback [%d]", rtkMisc.GetFuncInfo(), authStatus)
	CallbackAuthStatusCodeCB(authStatus)
}

//export diasSourceAndPortCallback
func diasSourceAndPortCallback(cSource C.uchar, cPort C.uchar) {
	if CallbackDIASSourceAndPortCB == nil {
		log.Printf("[%s] CallbackDIASSourceAndPortCB is not set, CallbackDIASSourceAndPortCB failed!", rtkMisc.GetFuncInfo())
		return
	}
	source := uint8(cSource)
	port := uint8(cPort)
	log.Printf("[%s] diasSourceAndPortCallback (src,port): (%d,%d)", rtkMisc.GetFuncInfo(), source, port)
	CallbackDIASSourceAndPortCB(source, port)
}

//export GoAuthViaIndex
func GoAuthViaIndex(clientIndex uint32) {
	cIndex := C.uint(clientIndex)
	C.AuthViaIndex(cIndex)
}

func GoReqSourceAndPort() {
	C.RequestSourceAndPort()
}

//export GoDIASStatusNotify
func GoDIASStatusNotify(diasStatus uint32) {
	cStatus := C.uint(diasStatus)
	log.Printf("GoDIASStatusNotify:[%d]", cStatus)
	C.DIASStatus(cStatus)
}

// For GO DEBUG
func GoGetMacAddressCallback(mac string) {
	cContent := C.CString(mac)
	defer C.free(unsafe.Pointer(cContent))
	getMacAddressCallback(cContent, 0)
}

// For GO DEBUG
func GoExtractDIASCallback() {
	extractDIASCallback()
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

// TODO: remove in TSTAS-189
//
//export dragFileCallback
func dragFileCallback(cTimestamp C.ulonglong, cFilePath *C.wchar_t) {
	if CallbackInstanceFileDragCB == nil {
		return
	}

	filePath := wcharToString(cFilePath)
	fileSize, err := rtkMisc.FileSize(filePath)
	if err != nil {
		return
	}

	uFileSize := uint64(fileSize)
	fileSizeHigh := uint32(uFileSize >> 32)
	fileSizeLow := uint32(uFileSize & 0xFFFFFFFF)
	var fileInfo = rtkCommon.FileInfo{
		FileSize_: rtkCommon.FileSize{
			SizeHigh: fileSizeHigh,
			SizeLow:  fileSizeLow,
		},
		FilePath: filePath,
		FileName: filepath.Base(filePath),
	}

	timestamp := uint64(cTimestamp)
	log.Printf("[%s] path:[%s] size:[%d]", rtkMisc.GetFuncInfo(), fileInfo.FilePath, fileSize)
	CallbackInstanceFileDragCB(fileInfo, timestamp)
}

//export cancelFileTransferCallback
func cancelFileTransferCallback(cIp *C.char, cId *C.char, cTimestamp C.ulonglong) {
	if CallbackCancelFileTransDragCB == nil {
		return
	}

	id := C.GoString(cId)
	ip := C.GoString(cIp)
	timestamp := uint64(cTimestamp)
	log.Printf("[%s] cancelFileTransferCallback ID:[%s] IPAddr:[%s] timestamp:[%d]", rtkMisc.GetFuncInfo(), id, ip, timestamp)
	CallbackCancelFileTransDragCB(id, ip, timestamp)
}

// TODO: remove in TSTAS-189
//
//export fileDropRequestCallback
func fileDropRequestCallback(cIp *C.char, cId *C.char, cFileSize C.ulonglong, cTimestamp C.ulonglong, cFilePath *C.wchar_t) {
	if CallbackInstanceFileDropRequestCB == nil {
		return
	}
	filePath := wcharToString(cFilePath)
	fileSize, err := rtkMisc.FileSize(filePath)
	if err != nil {
		log.Printf("[%s] get file:[%s] size error, skit it!", rtkMisc.GetFuncInfo(), filePath)
		return
	}

	id := C.GoString(cId)
	ip := C.GoString(cIp)
	//fileSize := uint64(cFileSize)
	fileSizeHigh := uint32(fileSize >> 32)
	fileSizeLow := uint32(fileSize & 0xFFFFFFFF)
	var fileInfo = rtkCommon.FileInfo{
		FileSize_: rtkCommon.FileSize{
			SizeHigh: uint32(fileSizeHigh),
			SizeLow:  uint32(fileSizeLow),
		},
		FilePath: filePath,
		FileName: filepath.Base(filePath),
	}
	timestamp := uint64(cTimestamp)
	log.Printf("[%s %d] id[%s] ip[%s] path:[%s] fileSize:[%d] timestamp:[%d]", rtkMisc.GetFuncName(), rtkMisc.GetLine(), id, ip, fileInfo.FilePath, fileSize, timestamp)
	CallbackInstanceFileDropRequestCB(id, fileInfo, timestamp)
}

//export multiFilesDropRequestCallback
func multiFilesDropRequestCallback(cIp *C.char, cId *C.char, cTimestamp C.ulonglong, cFileList **C.wchar_t, cFileCount C.uint) {
	if CallbackFileListDropRequestCB == nil {
		return
	}
	id := C.GoString(cId)
	ip := C.GoString(cIp)
	fileList := make([]rtkCommon.FileInfo, 0)
	folderList := make([]string, 0)
	totalSize := uint64(0)
	fileCount := uint32(cFileCount)

	for i := uint32(0); i < fileCount; i++ {
		wcharPtr := *(**C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cFileList)) + uintptr(i)*unsafe.Sizeof(*cFileList)))
		file := wcharToString(wcharPtr)
		log.Printf("[%s] get file or path:[%s]", rtkMisc.GetFuncInfo(), file)
		if rtkMisc.FolderExists(file) {
			rtkUtils.WalkPath(file, &folderList, &fileList, &totalSize)
		} else if rtkMisc.FileExists(file) {
			fileSize, err := rtkMisc.FileSize(file)
			if err != nil {
				log.Printf("[%s] get file:[%s] size error, skit it!", rtkMisc.GetFuncInfo(), file)
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
			log.Printf("[%s] get file:[%s] type error, skit it!", rtkMisc.GetFuncInfo(), file)
		}
	}
	timestamp := uint64(cTimestamp)
	totalDesc := rtkMisc.FileSizeDesc(totalSize)

	log.Printf("[%s] ID[%s] IP:[%s] get file count:[%d] folder count:[%d], totalSize:[%d] totalDesc:[%s] timestamp:[%d]", rtkMisc.GetFuncInfo(), id, ip, len(fileList), len(folderList), totalSize, totalDesc, timestamp)
	CallbackFileListDropRequestCB(id, fileList, folderList, totalSize, timestamp, totalDesc)
}

//export dragFileListRequestCallback
func dragFileListRequestCallback(cFileList **C.wchar_t, cFileCount C.uint, cTimestamp C.ulonglong) {
	if CallbackDragFileListRequestCB == nil {
		return
	}
	fileList := make([]rtkCommon.FileInfo, 0)
	folderList := make([]string, 0)
	totalSize := uint64(0)
	fileCount := uint32(cFileCount)

	for i := uint32(0); i < fileCount; i++ {
		wcharPtr := *(**C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cFileList)) + uintptr(i)*unsafe.Sizeof(*cFileList)))
		file := wcharToString(wcharPtr)
		log.Printf("[%s] get file or path:[%s]", rtkMisc.GetFuncInfo(), file)
		if rtkMisc.FolderExists(file) {
			rtkUtils.WalkPath(file, &folderList, &fileList, &totalSize)
		} else if rtkMisc.FileExists(file) {
			fileSize, err := rtkMisc.FileSize(file)
			if err != nil {
				log.Printf("[%s] get file:[%s] size error, skit it!", rtkMisc.GetFuncInfo(), file)
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
			log.Printf("[%s] get file:[%s] type error, skit it!", rtkMisc.GetFuncInfo(), file)
		}
	}
	timestamp := uint64(cTimestamp)
	totalDesc := rtkMisc.FileSizeDesc(totalSize)

	log.Printf("[%s] get file count:[%d] folder count:[%d], totalSize:[%d] totalDesc:[%s] timestamp:[%d]", rtkMisc.GetFuncInfo(), len(fileList), len(folderList), totalSize, totalDesc, timestamp)
	CallbackDragFileListRequestCB(fileList, folderList, totalSize, timestamp, totalDesc)
}

//export fileDropResponseCallback
func fileDropResponseCallback(cStatus int32, cIp *C.char, cId *C.char, cFileSize C.ulonglong, cTimestamp C.ulonglong, cFilePath *C.wchar_t) {
	if CallbackInstanceFileDropResponseCB == nil {
		return
	}

	id := C.GoString(cId)
	ip := C.GoString(cIp)
	if cStatus == 0 { // FILE_DROP_REJECT
		log.Println("FILE_DROP_REJECT")
		CallbackInstanceFileDropResponseCB(id, rtkCommon.FILE_DROP_REJECT, "")
	} else if cStatus == 1 { // FILE_DROP_ACCEPT
		//path := filepath.Dir(wcharToString(cFilePath))
		path := wcharToString(cFilePath)
		log.Printf("FILE_DROP_ACCEPT, ip:[%s] path:[%s]", ip, path)
		CallbackInstanceFileDropResponseCB(id, rtkCommon.FILE_DROP_ACCEPT, path)
	}
}

//export clipboardCopyImgCallback
func clipboardCopyImgCallback(cHeader C.IMAGE_HEADER, cData *C.uchar, cDataSize C.ulong) {
	if CallbackInstanceCopyImageCB == nil {
		log.Println("[clipboardCopyImgCallback] Err: null callback instance")
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
func GoSetupFileDrop(ip, id, fileName, platform string, fileSize uint64, timestamp uint64) {
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

// export SetupFileListDrop
func GoSetupFileListDrop(ip, id, platform, totalDesc string, fileCount, folderCount uint32, timestamp uint64) {
	log.Printf("[%s] fileCnt:[%d] folderCnt:[%d] totalDesc:[%s]", rtkMisc.GetFuncInfo(), fileCount, folderCount, totalDesc)
	/*cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))
	cTotalDesc := C.CString(totalDesc)
	defer C.free(unsafe.Pointer(cTotalDesc))

	cFileCount := C.uint(fileCount)
	cFolderCount := C.uint(folderCount)
	cTimestamp := C.ulonglong(timestamp)

	C.SetupFileListDrop(cIp, cId, cTotalDesc, cFileCount, cFolderCount, cTimestamp)*/
}

// export MultiFilesDropNotify
func GoMultiFilesDropNotify(ip, id, platform string, fileCnt uint32, totalSize, timestamp uint64, firstFileName string, firstFileSize uint64) {
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))

	cFileCnt := C.uint(fileCnt)
	cFirstFileSize := C.ulonglong(firstFileSize)
	cTotalSize := C.ulonglong(totalSize)
	cTimestamp := C.ulonglong(timestamp)
	cFirstFileName := stringToWChar(firstFileName)
	defer C.free(unsafe.Pointer(cFirstFileName))
	C.MultiFilesDropNotify(cIp, cId, cFileCnt, cTotalSize, cTimestamp, cFirstFileName, cFirstFileSize)
}

// export DragFileNotify
func GoDragFileNotify(ip, id, fileName, platform string, fileSize uint64, timestamp uint64) {
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))
	cFileSize := C.ulonglong(fileSize)
	cTimestamp := C.ulonglong(timestamp)
	cFileName := stringToWChar(fileName)
	defer C.free(unsafe.Pointer(cFileName))
	C.DragFileNotify(cIp, cId, cFileSize, cTimestamp, cFileName)
}

// export DragFileListNotify
func GoDragFileListNotify(ip, id, platform string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64) {
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))

	cFileCnt := C.uint(fileCnt)
	cToalSize := C.ulonglong(totalSize)
	cFirstFileSize := C.ulonglong(firstFileSize)
	cTimestamp := C.ulonglong(timestamp)
	cFirstFileName := stringToWChar(firstFileName)
	defer C.free(unsafe.Pointer(cFirstFileName))
	C.DragFileListNotify(cIp, cId, cFileCnt, cToalSize, cTimestamp, cFirstFileName, cFirstFileSize)
}

func GoDragFileListFolderNotify(ip, id, folderName string, timestamp uint64) {
}

// export SetupDstPasteImage
func GoSetupDstPasteImage(desc string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {
	// TODO: consider setup JPG image to windows clipboard
	// calculate Bitmap size with JPG image
	bmpSize := uint32(imgHeader.Height) * uint32(imgHeader.Width) * uint32(imgHeader.BitCount) / 8
	log.Printf("[Windows] SetupDstPasteImage with compression, height=%d, width=%d, bitCount=%d, size=%d",
		imgHeader.Height, imgHeader.Width, imgHeader.BitCount, bmpSize)

	cDesc := stringToWChar(desc)
	defer C.free(unsafe.Pointer(cDesc))
	cImgHeader := C.IMAGE_HEADER{
		width:       C.int(imgHeader.Width),
		height:      C.int(imgHeader.Height),
		planes:      C.ushort(imgHeader.Planes),
		bitCount:    C.ushort(imgHeader.BitCount),
		compression: C.ulong(imgHeader.Compression),
	}
	cDataSize := C.ulong(bmpSize)
	C.SetupDstPasteImage(cDesc, cImgHeader, cDataSize)
}

// export DataTransfer
func GoDataTransfer(data []byte) {
	// TODO: avoid to convert to BMP here, move to C++ partition
	startConvertTime := time.Now().UnixNano()
	bmpData, err := rtkUtils.JpgToBmp(data)
	log.Printf("(DST) Convert jpg to bmp, size:[%d] use [%d] ms...", len(bmpData), (time.Now().UnixNano()-startConvertTime)/1e6)
	if err != nil {
		log.Printf("(DST) Err: Convert JPG to BMP failed")
	}

	cData := (*C.uchar)(unsafe.Pointer(&bmpData[0]))
	cSize := C.uint(len(bmpData))
	C.DataTransfer(cData, cSize)
}

// export UpdateProgressBar
func GoUpdateProgressBar(ip, id string, fileSize, sentSize, timestamp uint64, fileName string) {
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

// export UpdateMultipleProgressBar
func GoUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))

	cSentFileCnt := C.uint(sentFileCnt)
	cTotalFileCnt := C.uint(totalFileCnt)

	cCurrentFileSize := C.ulonglong(currentFileSize)
	cSentSize := C.ulonglong(sentSize)
	cTotalSize := C.ulonglong(totalSize)
	cTimestamp := C.ulonglong(timestamp)
	cCurrentFileName := stringToWChar(currentFileName)
	defer C.free(unsafe.Pointer(cCurrentFileName))
	// on the Windows client, the deviceName can be queried via the clientID,
	// so it is not needed here and can be ignored.
	C.UpdateMultipleProgressBar(cIp, cId, cCurrentFileName, cSentFileCnt, cTotalFileCnt, cCurrentFileSize, cTotalSize, cSentSize, cTimestamp)
}

// export UpdateSystemInfo
func GoUpdateSystemInfo(ip, serviceVer string) {
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cServiceVer := stringToWChar(serviceVer)
	defer C.free(unsafe.Pointer(cServiceVer))
	C.UpdateSystemInfo(cIp, cServiceVer)
}

// export UpdateClientStatus
func GoUpdateClientStatus(status uint32, ip, id, name, deviceType string) {
	cStatus := C.uint(status)
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))
	cName := stringToWChar(name)
	defer C.free(unsafe.Pointer(cName))
	cDeviceType := C.CString(deviceType)
	defer C.free(unsafe.Pointer(cDeviceType))
	C.UpdateClientStatus(cStatus, cIp, cId, cName, cDeviceType)
}

func GoNotiMessageFileTransfer(fileName, clientName, platform string, timestamp uint64, isSender bool) {
	cTimestamp := C.ulonglong(timestamp)
	cCode := C.uint(C.NOTI_MSG_CODE_FILE_TRANS_DONE_RECEIVER)
	if isSender {
		cCode = C.uint(C.NOTI_MSG_CODE_FILE_TRANS_DONE_SENDER)
	}
	cFileName := stringToWChar(fileName)
	defer C.free(unsafe.Pointer(cFileName))
	cClientName := stringToWChar(clientName)
	defer C.free(unsafe.Pointer(cClientName))
	paramArray := []string{fileName, clientName}
	cParamArray := NewWCharArray(paramArray)
	defer cParamArray.Free()
	cParamCnt := C.int(len(paramArray))
	C.NotiMessage(cTimestamp, cCode, cParamArray.Array, cParamCnt)
}

// export EventHandle
func GoEventHandle(eventType rtkCommon.EventType, ipAddr, fileName string) {
	C.EventHandle(C.EVENT_TYPE(eventType))
}

// export CleanClipboard
func GoCleanClipboard() {
	// TODO: it cause crash after disconnection. Related issue: TSTAS-44
	// C.CleanClipboard()
}

func SetupCallbackSettings() {
	C.SetClipboardCopyFileCallback((C.ClipboardCopyFileCallback)(unsafe.Pointer(C.clipboardCopyFileCallback)))
	C.SetClipboardPasteFileCallback((C.ClipboardPasteFileCallback)(unsafe.Pointer(C.clipboardPasteFileCallback)))
	C.SetFileDropRequestCallback((C.FileDropRequestCallback)(unsafe.Pointer(C.fileDropRequestCallback)))
	C.SetFileDropResponseCallback((C.FileDropResponseCallback)(unsafe.Pointer(C.fileDropResponseCallback)))
	C.SetDragFileCallback((C.DragFileCallback)(unsafe.Pointer(C.dragFileCallback)))
	C.SetDragFileListRequestCallback((C.DragFileListRequestCallback)(unsafe.Pointer(C.dragFileListRequestCallback)))
	C.SetMultiFilesDropRequestCallback((C.MultiFilesDropRequestCallback)(unsafe.Pointer(C.multiFilesDropRequestCallback)))
	C.SetCancelFileTransferCallback((C.CancelFileTransferCallback)(unsafe.Pointer(C.cancelFileTransferCallback)))
	C.SetClipboardCopyImgCallback((C.ClipboardCopyImgCallback)(unsafe.Pointer(C.clipboardCopyImgCallback)))
	C.SetPipeConnectedCallback((C.PipeConnectedCallback)(unsafe.Pointer(C.pipeConnectedCallback)))
	C.SetGetMacAddressCallback((C.GetMacAddressCallback)(unsafe.Pointer(C.getMacAddressCallback)))
	C.SetExtractDIASCallback((C.ExtractDIASCallback)(unsafe.Pointer(C.extractDIASCallback)))
	C.SetAuthStatusCodeCallback((C.AuthStatusCodeCallback)(unsafe.Pointer(C.authStatusCodeCallback)))
	C.SetDIASSourceAndPortCallback((C.DIASSourceAndPortCallback)(unsafe.Pointer(C.diasSourceAndPortCallback)))
	C.StartClipboardMonitor()
	C.StartPipeMonitor()
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

func ReceiveImageCopyDataDone(fileSize int64, imgHeader rtkCommon.ImgHeader) {
}

func ReceiveFileDropCopyDataDone(fileSize int64, dstFilePath string) {
}

func FoundPeer() {
}

func GetClientList() string {
	clientList := rtkUtils.GetClientList()
	log.Printf("GetClientList :[%s]", clientList)
	return clientList
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
	return rtkGlobal.PlatformWindows
}

func GetMdnsPortConfigPath() string {
	return ".MdnsPort"
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

// export GetDeviceName
func GoGetDeviceName() string {
	return wcharToString(C.GetDeviceName())
}

func getDownloadPathInternal() string {
	path := C.GetDownloadPath()
	if path == nil {
		log.Fatalf("[%s] C.GetDownloadPath :[%s] invalid!", rtkMisc.GetFuncInfo(), path)
	}
	defer C.CoTaskMemFree(unsafe.Pointer(path))
	downLoadPath := wcharToString(path)
	if !rtkMisc.FolderExists(downLoadPath) {
		log.Fatalf("[%s] getDownloadPath :[%s] invalid!", rtkMisc.GetFuncInfo(), downLoadPath)
	}

	log.Printf(" getDownloadPath:[%s] success!", downLoadPath)
	return downLoadPath
}

func GoGetSrcAndPortFromIni() rtkMisc.SourcePort {
	return rtkUtils.GetDeviceSrcPort()
}

// TODO: handle after TSTAS-189 complete
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
