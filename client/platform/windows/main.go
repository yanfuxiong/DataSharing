//go:build windows

package main

/*
#cgo LDFLAGS: -static -lstdc++
#include <stdint.h>
#include <wchar.h>
#include <stdlib.h>

typedef struct {
    int width;
    int height;
    unsigned short planes;
    unsigned short bitCount;
    unsigned long compression;
} IMAGE_HEADER;

typedef enum NOTI_MSG_CODE
{
    NOTI_MSG_CODE_CONN_STATUS_SUCCESS       = 1,
    NOTI_MSG_CODE_CONN_STATUS_FAIL          = 2,
    NOTI_MSG_CODE_FILE_TRANS_DONE_SENDER    = 3,
    NOTI_MSG_CODE_FILE_TRANS_DONE_RECEIVER  = 4,
    NOTI_MSG_CODE_FILE_TRANS_REJECT         = 5,
} NOTI_MSG_CODE;

typedef void (*StartClipboardMonitorCallback)();
static void StartClipboardMonitorCallbackFunc(StartClipboardMonitorCallback cb) {
    if (cb) cb();
}

typedef void (*StopClipboardMonitorCallback)();
static void StopClipboardMonitorCallbackFunc(StopClipboardMonitorCallback cb) {
    if (cb) cb();
}

typedef void (*SetupFileDropCallback)(const char *ipPort, const char *id, uint64_t fileSize, uint64_t timestamp, const wchar_t *fileName);
static void SetupFileDropCallbackFunc(SetupFileDropCallback cb, const char *ipPort, const char *id, uint64_t fileSize, uint64_t timestamp, const wchar_t *fileName) {
    if (cb) cb(ipPort, id, fileSize, timestamp, fileName);
}

typedef void (*DragFileNotifyCallback)(const char *ipPort, const char *clientID, uint64_t fileSize, uint64_t timestamp, const wchar_t *fileName);
static void DragFileNotifyCallbackFunc(DragFileNotifyCallback cb, const char *ipPort, const char *clientID, uint64_t fileSize, uint64_t timestamp, const wchar_t *fileName) {
    if (cb) cb(ipPort, clientID, fileSize, timestamp, fileName);
}

typedef void (*DragFileListNotifyCallback)(const char *ipPort, const char *clientID, uint32_t cFileCount, uint64_t totalSize, uint64_t timestamp, const wchar_t *firstFileName, uint64_t firstFileSize);
static void DragFileListNotifyCallbackFunc(DragFileListNotifyCallback cb, const char *ipPort, const char *clientID, uint32_t cFileCount, uint64_t totalSize, uint64_t timestamp, const wchar_t *firstFileName, uint64_t firstFileSize) {
    if (cb) cb(ipPort, clientID, cFileCount, totalSize, timestamp, firstFileName, firstFileSize);
}

typedef void (*MultiFilesDropNotifyCallback)(const char *ipPort, const char *clientID, uint32_t cFileCount, uint64_t totalSize, uint64_t timestamp, const wchar_t *firstFileName, uint64_t firstFileSize);
static void MultiFilesDropNotifyCallbackFunc(MultiFilesDropNotifyCallback cb, const char *ipPort, const char *clientID, uint32_t cFileCount, uint64_t totalSize, uint64_t timestamp, const wchar_t *firstFileName, uint64_t firstFileSize) {
    if (cb) cb(ipPort, clientID, cFileCount, totalSize, timestamp, firstFileName, firstFileSize);
}

typedef void (*UpdateMultipleProgressBarCallback)(const char *ipPort, const char *clientID, const wchar_t *currentFileName, uint32_t sentFilesCnt, uint32_t totalFilesCnt, uint64_t currentFileSize, uint64_t totalSize, uint64_t sentSize, uint64_t timestamp);
static void UpdateMultipleProgressBarCallbackFunc(UpdateMultipleProgressBarCallback cb, const char *ipPort, const char *clientID, const wchar_t *currentFileName, uint32_t sentFilesCnt, uint32_t totalFilesCnt, uint64_t currentFileSize, uint64_t totalSize, uint64_t sentSize, uint64_t timestamp) {
    if (cb) cb(ipPort, clientID, currentFileName, sentFilesCnt, totalFilesCnt, currentFileSize, totalSize, sentSize, timestamp);
}

typedef void (*DataTransferCallback)(const unsigned char *data, uint32_t size);
static void DataTransferCallbackFunc(DataTransferCallback cb, const unsigned char *data, uint32_t size) {
    if (cb) cb(data, size);
}

typedef void (*UpdateProgressBarCallback)(const char *ipPort, const char *id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp, const wchar_t *fileName);
static void UpdateProgressBarCallbackFunc(UpdateProgressBarCallback cb, const char *ipPort, const char *id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp, const wchar_t *fileName) {
    if (cb) cb(ipPort, id, fileSize, sentSize, timestamp, fileName);
}

typedef void (*DeinitProgressBarCallback)();
static void DeinitProgressBarCallbackFunc(DeinitProgressBarCallback cb) {
    if (cb) cb();
}

typedef void (*UpdateImageProgressBarCallback)(const char *ipPort, const char *id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp);
static void UpdateImageProgressBarCallbackFunc(UpdateImageProgressBarCallback cb, const char *ipPort, const char *id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp) {
    if (cb) cb(ipPort, id, fileSize, sentSize, timestamp);
}

typedef void (*UpdateClientStatusCallback)(uint32_t status, const char *ipPort, const char *id, const wchar_t *name, const char *deviceType);
static void UpdateClientStatusCallbackFunc(UpdateClientStatusCallback cb, uint32_t status, const char *ipPort, const char *id, const wchar_t *name, const char *deviceType) {
    if (cb) cb(status, ipPort, id, name, deviceType);
}

typedef void (*UpdateSystemInfoCallback)(const char *ipPort, const wchar_t *serviceVer);
static void UpdateSystemInfoCallbackFunc(UpdateSystemInfoCallback cb, const char *ipPort, const wchar_t *serviceVer) {
    if (cb) cb(ipPort, serviceVer);
}

typedef void (*NotiMessageCallback)(uint64_t timestamp, uint32_t notiCode, const wchar_t *notiParam[], int paramCount);
static void NotiMessageCallbackFunc(NotiMessageCallback cb, uint64_t timestamp, uint32_t notiCode, const wchar_t *notiParam[], int paramCount) {
    if (cb) cb(timestamp, notiCode, notiParam, paramCount);
}

typedef void (*CleanClipboardCallback)();
static void CleanClipboardCallbackFunc(CleanClipboardCallback cb) {
    if (cb) cb();
}

typedef void (*AuthViaIndexCallback)(uint32_t index);
static void AuthViaIndexCallbackFunc(AuthViaIndexCallback cb, uint32_t index) {
    if (cb) cb(index);
}

typedef void (*DIASStatusCallback)(uint32_t statusCode);
static void DIASStatusCallbackFunc(DIASStatusCallback cb, uint32_t code) {
    if (cb) cb(code);
}

typedef void (*RequestSourceAndPortCallback)();
static void RequestSourceAndPortCallbackFunc(RequestSourceAndPortCallback cb) {
    if (cb) cb();
}

typedef void (*SetupDstPasteImageCallback)(const wchar_t* desc, IMAGE_HEADER imgHeader, uint32_t dataSize);
static void SetupDstPasteImageCallbackFunc(SetupDstPasteImageCallback cb, const wchar_t* desc, IMAGE_HEADER imgHeader, uint32_t dataSize) {
    if (cb) cb(desc, imgHeader, dataSize);
}

typedef void (*RequestUpdateClientVersionCallback)(const char *clienVersion);
static void RequestUpdateClientVersionCallbackFunc(RequestUpdateClientVersionCallback cb, const char *clienVersion) {
    if (cb) cb(clienVersion);
}

typedef void (*NotifyErrEventCallback)(const char *clientID, uint32_t errCode, const char *arg1, const char *arg2, const char *arg3, const char *arg4);
static void NotifyErrEventCallbackFunc(NotifyErrEventCallback cb, const char *clienID, uint32_t errCode, const char *arg1, const char *arg2, const char *arg3, const char *arg4) {
    if (cb) cb(clienID, errCode, arg1, arg2, arg3, arg4);
}
*/
import "C"
import (
	"fmt"
	"log"
	"path/filepath"
	rtkCmd "rtk-cross-share/client/cmd"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"unicode/utf16"
	"unsafe"
)

func WCharToGoString(wchar *C.wchar_t) string {
	if wchar == nil {
		return ""
	}
	var nLen int
	for *(*C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(wchar)) + uintptr(nLen)*unsafe.Sizeof(*wchar))) != 0 {
		nLen++
	}

	utf16Slice := make([]uint16, nLen)
	for i := 0; i < nLen; i++ {
		utf16Slice[i] = uint16(*(*C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(wchar)) + uintptr(i)*unsafe.Sizeof(*wchar))))
	}

	return string(utf16.Decode(utf16Slice))
}

func GoStringToWChar(str string) *C.wchar_t {
	utf16 := utf16.Encode([]rune(str))
	cStr := (*C.wchar_t)(C.malloc(C.size_t(len(utf16)+1) * 2)) // +1 for null terminator
	for i, v := range utf16 {
		*(*C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cStr)) + uintptr(i)*2)) = C.wchar_t(v)
	}

	//Add null terminator
	*(*C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cStr)) + uintptr(len(utf16))*2)) = 0
	return cStr
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
		cStr := GoStringToWChar(s)
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

var (
	g_StartClipboardMonitorCallback      C.StartClipboardMonitorCallback      = nil
	g_StopClipboardMonitorCallback       C.StopClipboardMonitorCallback       = nil
	g_DragFileNotifyCallback             C.DragFileNotifyCallback             = nil
	g_DragFileListNotifyCallback         C.DragFileListNotifyCallback         = nil
	g_MultiFilesDropNotifyCallback       C.MultiFilesDropNotifyCallback       = nil
	g_UpdateMultipleProgressBarCallback  C.UpdateMultipleProgressBarCallback  = nil
	g_DataTransferCallback               C.DataTransferCallback               = nil
	g_UpdateProgressBarCallback          C.UpdateProgressBarCallback          = nil
	g_UpdateImageProgressBarCallback     C.UpdateImageProgressBarCallback     = nil
	g_UpdateClientStatusCallback         C.UpdateClientStatusCallback         = nil
	g_UpdateSystemInfoCallback           C.UpdateSystemInfoCallback           = nil
	g_NotiMessageCallback                C.NotiMessageCallback                = nil
	g_CleanClipboardCallback             C.CleanClipboardCallback             = nil
	g_AuthViaIndexCallback               C.AuthViaIndexCallback               = nil
	g_DIASStatusCallback                 C.DIASStatusCallback                 = nil
	g_RequestSourceAndPortCallback       C.RequestSourceAndPortCallback       = nil
	g_SetupDstPasteImageCallback         C.SetupDstPasteImageCallback         = nil
	g_RequestUpdateClientVersionCallback C.RequestUpdateClientVersionCallback = nil
	g_NotifyErrEventCallback             C.NotifyErrEventCallback             = nil
)

func main() {}

func init() {
	rtkPlatform.SetAuthViaIndexCallback(GoTriggerCallbackSetAuthViaIndex)
	rtkPlatform.SetDIASStatusCallback(GoTriggerCallbackSetDIASStatus)
	rtkPlatform.SetRequestSourceAndPortCallback(GoTriggerCallbackRequestSourceAndPort)
	rtkPlatform.SetUpdateSystemInfoCallback(GoTriggerCallbackUpdateSystemInfo)
	rtkPlatform.SetUpdateClientStatusCallback(GoTriggerCallbackUpdateClientStatus)
	rtkPlatform.SetImageDataTransferCallback(GoTriggerCallbackImageDataTransfer)
	rtkPlatform.SetSetupDstImageCallback(GoTriggerCallbackSetupDstPasteImage)
	rtkPlatform.SetCleanClipboardCallback(GoTriggerCallbackCleanClipboard)
	rtkPlatform.SetDragFileListNotifyCallback(GoTriggerCallbackDragFileListNotify)
	rtkPlatform.SetMultiFilesDropNotifyCallback(GoTriggerCallbackMultiFilesDropNotify)
	rtkPlatform.SetMultipleProgressBarCallback(GoTriggerCallbackMultipleProgressBar)
	rtkPlatform.SetNotiMessageFileTransCallback(GoTriggerCallbackNotiMessage)
	rtkPlatform.SetReqClientUpdateVerCallback(GoTriggerCallbackReqClientUpdateVer)
	rtkPlatform.SetNotifyErrEventCallback(GoTriggerCallbackNotifyErrEvent)

	rtkPlatform.SetConfirmDocumentsAccept(false)
}

func GoTriggerCallbackSetAuthViaIndex(index uint32) {
	if g_AuthViaIndexCallback == nil {
		log.Printf("%s g_AuthViaIndexCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	cIndex := C.uint32_t(index)
	log.Printf("[%s] set AuthViaIndex:[%d]", rtkMisc.GetFuncInfo(), index)
	C.AuthViaIndexCallbackFunc(g_AuthViaIndexCallback, cIndex)
}

func GoTriggerCallbackSetDIASStatus(status uint32) {
	if g_DIASStatusCallback == nil {
		log.Printf("%s g_DIASStatusCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

	cStatus := C.uint32_t(status)
	log.Printf("[%s] set DIASStatus:[%d]", rtkMisc.GetFuncInfo(), status)
	C.DIASStatusCallbackFunc(g_DIASStatusCallback, cStatus)
}

func GoTriggerCallbackRequestSourceAndPort() {
	if g_RequestSourceAndPortCallback == nil {
		log.Printf("%s g_RequestSourceAndPortCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	C.RequestSourceAndPortCallbackFunc(g_RequestSourceAndPortCallback)
}

func GoTriggerCallbackUpdateSystemInfo(ipAddr, versionInfo string) {
	if g_UpdateSystemInfoCallback == nil {
		log.Printf("%s g_UpdateSystemInfoCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

	cIpAddr := C.CString(ipAddr)
	defer C.free(unsafe.Pointer(cIpAddr))

	cServiceVer := GoStringToWChar(versionInfo)
	defer C.free(unsafe.Pointer(cServiceVer))

	C.UpdateSystemInfoCallbackFunc(g_UpdateSystemInfoCallback, cIpAddr, cServiceVer)
}

func GoTriggerCallbackUpdateClientStatus(status uint32, ip, id, deviceName, deviceType string) {
	if g_UpdateClientStatusCallback == nil {
		log.Printf("%s g_UpdateClientStatusCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	cStatus := C.uint32_t(status)
	cIp := C.CString(ip)
	cId := C.CString(id)
	cDeviceName := GoStringToWChar(deviceName)
	cDeviceType := C.CString(deviceType)

	defer func() {
		C.free(unsafe.Pointer(cIp))
		C.free(unsafe.Pointer(cId))
		C.free(unsafe.Pointer(cDeviceName))
		C.free(unsafe.Pointer(cDeviceType))
	}()

	C.UpdateClientStatusCallbackFunc(g_UpdateClientStatusCallback, cStatus, cIp, cId, cDeviceName, cDeviceType)
}

func GoTriggerCallbackSetupDstPasteImage(Id string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {
	if g_SetupDstPasteImageCallback == nil {
		log.Printf("%s g_SetupDstPasteImageCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

	log.Printf("[Windows] SetupDstPasteImage with compression, height=%d, width=%d, bitCount=%d, size=%d",
		imgHeader.Height, imgHeader.Width, imgHeader.BitCount, dataSize)

	cId := GoStringToWChar(Id)
	defer C.free(unsafe.Pointer(cId))

	cImgHeader := C.IMAGE_HEADER{
		width:       C.int(imgHeader.Width),
		height:      C.int(imgHeader.Height),
		planes:      C.ushort(imgHeader.Planes),
		bitCount:    C.ushort(imgHeader.BitCount),
		compression: C.ulong(imgHeader.Compression),
	}
	cDataSize := C.uint32_t(dataSize)

	C.SetupDstPasteImageCallbackFunc(g_SetupDstPasteImageCallback, cId, cImgHeader, cDataSize)
}

func GoTriggerCallbackImageDataTransfer(data []byte) {
	if g_DataTransferCallback == nil {
		log.Printf("%s g_DataTransferCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

	cLen := C.uint32_t(len(data))
	log.Printf("[%s] len:%d", rtkMisc.GetFuncInfo(), cLen)
	C.DataTransferCallbackFunc(g_DataTransferCallback, (*C.uchar)(unsafe.Pointer(&data[0])), cLen)
}

func GoTriggerCallbackCleanClipboard() {
	if g_CleanClipboardCallback == nil {
		log.Printf("%s g_CleanClipboardCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	log.Printf("[%s] trigger!", rtkMisc.GetFuncInfo())
	C.CleanClipboardCallbackFunc(g_CleanClipboardCallback)
}

func GoTriggerCallbackMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	if g_UpdateMultipleProgressBarCallback == nil {
		log.Printf("%s g_UpdateMultipleProgressBarCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

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
	cCurrentFileName := GoStringToWChar(currentFileName)
	defer C.free(unsafe.Pointer(cCurrentFileName))

	//const char *ipPort, const char *clientID, const wchar_t *currentFileName, uint32_t sentFilesCnt, uint32_t totalFilesCnt, uint64_t currentFileSize, uint64_t totalSize, uint64_t sentSize, uint64_t timestamp
	C.UpdateMultipleProgressBarCallbackFunc(g_UpdateMultipleProgressBarCallback, cIp, cId, cCurrentFileName, cSentFileCnt, cTotalFileCnt, cCurrentFileSize, cTotalSize, cSentSize, cTimestamp)
}

func GoTriggerCallbackDragFileListNotify(ip, id, platform string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64) {
	if g_DragFileListNotifyCallback == nil {
		log.Printf("%s g_DragFileListNotifyCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))

	cFileCnt := C.uint(fileCnt)
	cToalSize := C.ulonglong(totalSize)
	cFirstFileSize := C.ulonglong(firstFileSize)
	cTimestamp := C.ulonglong(timestamp)
	cFirstFileName := GoStringToWChar(firstFileName)
	defer C.free(unsafe.Pointer(cFirstFileName))

	log.Printf("[%s] ID:[%s] IP:[%s] cnt:[%d] total:[%d] timestamp:[%d] firstFile:[%s] firstSize:[%d]", rtkMisc.GetFuncInfo(), ip, id, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
	C.DragFileListNotifyCallbackFunc(g_DragFileListNotifyCallback, cIp, cId, cFileCnt, cToalSize, cTimestamp, cFirstFileName, cFirstFileSize)
}

func GoTriggerCallbackMultiFilesDropNotify(ip, id, platform string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64) {
	if g_MultiFilesDropNotifyCallback == nil {
		log.Printf("%s g_MultiFilesDropNotifyCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))

	cFileCnt := C.uint(fileCnt)
	cToalSize := C.ulonglong(totalSize)
	cFirstFileSize := C.ulonglong(firstFileSize)
	cTimestamp := C.ulonglong(timestamp)
	cFirstFileName := GoStringToWChar(firstFileName)
	defer C.free(unsafe.Pointer(cFirstFileName))

	log.Printf("[%s] ID:[%s] IP:[%s] cnt:[%d] total:[%d] timestamp:[%d] firstFile:[%s] firstSize:[%d]", rtkMisc.GetFuncInfo(), ip, id, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
	C.MultiFilesDropNotifyCallbackFunc(g_MultiFilesDropNotifyCallback, cIp, cId, cFileCnt, cToalSize, cTimestamp, cFirstFileName, cFirstFileSize)
}

func GoTriggerCallbackNotiMessage(fileName, clientName, platform string, timestamp uint64, isSender bool) {
	if g_NotiMessageCallback == nil {
		log.Printf("%s g_NotiMessageCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

	cTimestamp := C.ulonglong(timestamp)
	cCode := C.uint(C.NOTI_MSG_CODE_FILE_TRANS_DONE_RECEIVER)
	if isSender {
		cCode = C.uint(C.NOTI_MSG_CODE_FILE_TRANS_DONE_SENDER)
	}
	cFileName := GoStringToWChar(fileName)
	defer C.free(unsafe.Pointer(cFileName))
	cClientName := GoStringToWChar(clientName)
	defer C.free(unsafe.Pointer(cClientName))
	paramArray := []string{fileName, clientName}
	cParamArray := NewWCharArray(paramArray)
	defer cParamArray.Free()
	cParamCnt := C.int(len(paramArray))

	log.Printf("[%s] timestamp:[%d] code:[%d] notiParam:[%+v] cParamCnt:[%d]", rtkMisc.GetFuncInfo(), cTimestamp, cCode, cParamArray.Array, cParamCnt)
	//(uint64_t timestamp, uint32_t notiCode, const wchar_t *notiParam[], int paramCount)
	C.NotiMessageCallbackFunc(g_NotiMessageCallback, cTimestamp, cCode, cParamArray.Array, cParamCnt)
}

func GoTriggerCallbackReqClientUpdateVer(ver string) {
	if g_RequestUpdateClientVersionCallback == nil {
		log.Printf("[%s] g_RequestUpdateClientVersionCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

	cVer := C.CString(ver)
	defer C.free(unsafe.Pointer(cVer))

	log.Printf("[%s] version:%s", rtkMisc.GetFuncInfo(), ver)
	C.RequestUpdateClientVersionCallbackFunc(g_RequestUpdateClientVersionCallback, cVer)
}

func GoTriggerCallbackNotifyErrEvent(id string, errCode uint32, arg1, arg2, arg3, arg4 string) {
	if g_NotifyErrEventCallback == nil {
		log.Printf("[%s] g_NotifyErrEventCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

	cId := C.CString(id)
	cErrCode := C.uint(errCode)
	cArg1 := C.CString(arg1)
	cArg2 := C.CString(arg2)
	cArg3 := C.CString(arg3)
	cArg4 := C.CString(arg4)
	defer func() {
		C.free(unsafe.Pointer(cId))
		C.free(unsafe.Pointer(cArg1))
		C.free(unsafe.Pointer(cArg2))
		C.free(unsafe.Pointer(cArg3))
		C.free(unsafe.Pointer(cArg4))
	}()

	log.Printf("[%s] id:[%s] errCode:%d arg1:%s, arg2:%s, arg3:%s, arg4:%s", rtkMisc.GetFuncInfo(), id, errCode, arg1, arg2, arg3, arg4)
	C.NotifyErrEventCallbackFunc(g_NotifyErrEventCallback, cId, cErrCode, cArg1, cArg2, cArg3, cArg4)
}

//export InitGoServer
func InitGoServer(cRootPath, cDownloadPath, cDeviceName *C.wchar_t) {
	rootPath := WCharToGoString(cRootPath)
	if rootPath == "" || !rtkMisc.FolderExists(rootPath) {
		log.Fatalf("[%s] get rootPath [%s] is invalid!", rtkMisc.GetFuncInfo(), rootPath)
	}

	downloadPath := WCharToGoString(cDownloadPath)
	if downloadPath == "" || !rtkMisc.FolderExists(downloadPath) {
		log.Fatalf("[%s] get downloadPath [%s] is invalid!", rtkMisc.GetFuncInfo(), downloadPath)
	}

	deviceName := WCharToGoString(cDeviceName)
	rtkPlatform.InitPlatform(rootPath, downloadPath, deviceName)
	rtkCmd.Run()
}

//export SetClipboardCopyImg
func SetClipboardCopyImg(cHeader C.IMAGE_HEADER, bitmapData *C.uchar, cDataSize C.ulong) {

	imgHeader := rtkCommon.ImgHeader{
		Width:       int32(cHeader.width),
		Height:      int32(cHeader.height),
		Planes:      uint16(cHeader.planes),
		BitCount:    uint16(cHeader.bitCount),
		Compression: uint32(cHeader.compression),
	}

	data := C.GoBytes(unsafe.Pointer(bitmapData), C.int(cDataSize))
	dataSize := uint32(cDataSize)
	// FIXME
	filesize := rtkCommon.FileSize{
		SizeHigh: 0,
		SizeLow:  dataSize,
	}

	rtkPlatform.GoCopyImage(filesize, imgHeader, data)
	log.Printf("Clipboard image content, width[%d] height[%d] data size[%d] \n", imgHeader.Width, imgHeader.Height, dataSize)
}

//export SetFileDropRequest
func SetFileDropRequest(ipPort *C.char, clientID *C.char, fileSize C.uint64_t, timestamp C.uint64_t, fileName *C.wchar_t) {
	fmt.Printf("SetFileDropRequest(%q, %q, %d, %d, %q)\n",
		C.GoString(ipPort),
		C.GoString(clientID),
		fileSize,
		timestamp,
		WCharToGoString(fileName))
}

//export SetFileDropResponse
func SetFileDropResponse(statusCode C.int, ipPort *C.char, clientID *C.char, fileSize C.uint64_t, timestamp C.uint64_t, fileName *C.wchar_t) {
	fmt.Printf("SetFileDropResponse(%d, %q, %q, %d, %d, %q)\n",
		statusCode,
		C.GoString(ipPort),
		C.GoString(clientID),
		fileSize,
		timestamp,
		WCharToGoString(fileName))
}

//export SetMacAddress
func SetMacAddress(cMacAddress *C.char, length C.int) {
	log.Printf("SetMacAddress(%q, %d)\n", C.GoString(cMacAddress), length)
	if length != 6 {
		log.Printf("[%s] getMacAddressCallback failed, invalid MAC length:%d", rtkMisc.GetFuncInfo(), length)
		return
	}
	macBytes := C.GoBytes(unsafe.Pointer(cMacAddress), 6)

	macAddress := fmt.Sprintf("%02X%02X%02X%02X%02X%02X",
		macBytes[0], macBytes[1], macBytes[2],
		macBytes[3], macBytes[4], macBytes[5])

	log.Printf("[%s] MacAddress [%s]", rtkMisc.GetFuncInfo(), macAddress)
	rtkPlatform.GoSetMacAddress(macAddress)
}

//export SetExtractDIAS
func SetExtractDIAS() {
	log.Printf("[%s] ExtractDIAS", rtkMisc.GetFuncInfo())
	rtkPlatform.GoExtractDIASCallback()
}

//export SetAuthStatusCode
func SetAuthStatusCode(authResult C.uchar) {
	log.Printf("SetAuthStatusCode(%d)\n", authResult)
	authStatus := uint8(authResult)
	rtkPlatform.GoSetAuthStatusCode(authStatus)
}

//export SetDIASSourceAndPort
func SetDIASSourceAndPort(cSource C.uchar, cPort C.uchar) {
	source := uint8(cSource)
	port := uint8(cPort)
	log.Printf("[%s] diasSourceAndPortCallback (src,port): (%d,%d)", rtkMisc.GetFuncInfo(), source, port)
	rtkPlatform.GoSetDIASSourceAndPort(source, port)
}

//export SetDragFile
func SetDragFile(timeStamp C.uint64_t, filePath *C.wchar_t) {
	fmt.Printf("SetDragFile(%d, %q)\n", timeStamp, WCharToGoString(filePath))
}

//export SetConfirmDocumentsAccept
func SetConfirmDocumentsAccept(ifConfirm bool) {
	log.Printf("[%s], ifConfirm:[%+v]", rtkMisc.GetFuncInfo(), ifConfirm)
	//rtkPlatform.SetConfirmDocumentsAccept(ifConfirm)
}

//export SetDragFileListRequest
func SetDragFileListRequest(filePathArry **C.wchar_t, arryLength C.uint32_t, timeStamp C.uint64_t) {
	fileList := make([]rtkCommon.FileInfo, 0)
	folderList := make([]string, 0)
	totalSize := uint64(0)
	fileCount := uint32(arryLength)
	nFileCnt := 0
	nFolderCnt := 0
	nPathSize := uint64(0)

	for i := uint32(0); i < fileCount; i++ {
		wcharPtr := *(**C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(filePathArry)) + uintptr(i)*unsafe.Sizeof(*filePathArry)))
		file := WCharToGoString(wcharPtr)
		if rtkMisc.FolderExists(file) {
			nFileCnt = len(fileList)
			nFolderCnt = len(folderList)
			nPathSize = totalSize
			rtkUtils.WalkPath(file, &folderList, &fileList, &totalSize)
			log.Printf("[%s] walk a path:[%s], get [%d] files and [%d] folders, total size:[%d]", rtkMisc.GetFuncInfo(), file, len(fileList)-nFileCnt, len(folderList)-nFolderCnt, totalSize-nPathSize)
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
			log.Printf("[%s] get a file:[%s], size:[%d] ", rtkMisc.GetFuncInfo(), file, fileSize)
		} else {
			log.Printf("[%s] get file or path:[%s] is invalid, skit it!", rtkMisc.GetFuncInfo(), file)
		}
	}
	timestamp := uint64(timeStamp)
	totalDesc := rtkMisc.FileSizeDesc(totalSize)

	log.Printf("[%s] get file count:[%d] folder count:[%d], totalSize:[%d] totalDesc:[%s] timestamp:[%d]", rtkMisc.GetFuncInfo(), len(fileList), len(folderList), totalSize, totalDesc, timestamp)
	rtkPlatform.GoDragFileListRequest(&fileList, &folderList, totalSize, timestamp, totalDesc)
}

//export SetCancelFileTransfer
func SetCancelFileTransfer(ipPort *C.char, clientID *C.char, timeStamp C.uint64_t) {
	log.Printf("SetCancelFileTransfer(%q, %q, %d)\n",
		C.GoString(ipPort),
		C.GoString(clientID),
		timeStamp)

	cIpAddr := C.GoString(ipPort)
	cId := C.GoString(clientID)
	cTimestamp := int64(timeStamp)

	rtkPlatform.GoCancelFileTrans(cIpAddr, cId, cTimestamp)
}

//export SetMultiFilesDropRequest
func SetMultiFilesDropRequest(ipPort *C.char, clientID *C.char, timeStamp C.uint64_t, filePathArry **C.wchar_t, arryLength C.uint32_t) C.uint {
	//(cIp *C.char, cId *C.char, cTimestamp C.ulonglong, cFileList **C.wchar_t, cFileCount C.uint)
	id := C.GoString(clientID)
	ip := C.GoString(ipPort)
	fileList := make([]rtkCommon.FileInfo, 0)
	folderList := make([]string, 0)
	totalSize := uint64(0)
	fileCount := uint32(arryLength)
	nFileCnt := 0
	nFolderCnt := 0
	nPathSize := uint64(0)

	for i := uint32(0); i < fileCount; i++ {
		wcharPtr := *(**C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(filePathArry)) + uintptr(i)*unsafe.Sizeof(*filePathArry)))
		file := WCharToGoString(wcharPtr)
		if rtkMisc.FolderExists(file) {
			nFileCnt = len(fileList)
			nFolderCnt = len(folderList)
			nPathSize = totalSize
			rtkUtils.WalkPath(file, &folderList, &fileList, &totalSize)
			log.Printf("[%s] walk a path:[%s], get [%d] files and [%d] folders , total size:[%d] ", rtkMisc.GetFuncInfo(), file, len(fileList)-nFileCnt, len(folderList)-nFolderCnt, totalSize-nPathSize)
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
			log.Printf("[%s] get a file:[%s], size:[%d] ", rtkMisc.GetFuncInfo(), file, fileSize)
		} else {
			log.Printf("[%s] get file or path:[%s] is invalid, skit it!", rtkMisc.GetFuncInfo(), file)
		}
	}
	timestamp := uint64(timeStamp)
	totalDesc := rtkMisc.FileSizeDesc(totalSize)

	log.Printf("[%s] ID[%s] IP:[%s] get file count:[%d] folder count:[%d], totalSize:[%d] totalDesc:[%s] timestamp:[%d]", rtkMisc.GetFuncInfo(), id, ip, len(fileList), len(folderList), totalSize, totalDesc, timestamp)
	return C.uint(rtkPlatform.GoMultiFilesDropRequest(id, &fileList, &folderList, totalSize, timestamp, totalDesc))
}

//export SetMsgEventFunc
func SetMsgEventFunc(cEvent C.uint32_t, cArg1 *C.char, cArg2 *C.char, cArg3 *C.char, cArg4 *C.char) {
	event := uint32(cEvent)
	arg1 := C.GoString(cArg1)
	arg2 := C.GoString(cArg2)
	arg3 := C.GoString(cArg3)
	arg4 := C.GoString(cArg4)

	log.Printf("[%s] event:[%d], arg1:%s, arg2:%s, arg3:%s, arg4:%s\n", rtkMisc.GetFuncInfo(), event, arg1, arg2, arg3, arg4)
	rtkPlatform.GoSetMsgEventFunc(event, arg1, arg2, arg3, arg4)
}

//export RequestUpdateDownloadPath
func RequestUpdateDownloadPath(cDownloadPath *C.wchar_t) {
	downloadPath := WCharToGoString(cDownloadPath)
	if downloadPath == "" || !rtkMisc.FolderExists(downloadPath) {
		log.Printf("[%s] get downloadPath [%s] is invalid!", rtkMisc.GetFuncInfo(), downloadPath)
		return
	}
	log.Printf("[%s] update downloadPath:[%s] success!", rtkMisc.GetFuncInfo(), downloadPath)
	rtkPlatform.GoUpdateDownloadPath(downloadPath)
}

//export SetStartClipboardMonitorCallback
func SetStartClipboardMonitorCallback(callback C.StartClipboardMonitorCallback) {
	fmt.Println("SetStartClipboardMonitorCallback")
	g_StartClipboardMonitorCallback = callback
	C.StartClipboardMonitorCallbackFunc(callback)
}

//export SetStopClipboardMonitorCallback
func SetStopClipboardMonitorCallback(callback C.StopClipboardMonitorCallback) {
	fmt.Println("SetStopClipboardMonitorCallback")
	g_StopClipboardMonitorCallback = callback
}

//export SetSetupFileDropCallback
func SetSetupFileDropCallback(callback C.SetupFileDropCallback) {
	fmt.Println("SetSetupFileDropCallback")
	//g_SetupFileDropCallback = callback
}

//export SetDragFileNotifyCallback
func SetDragFileNotifyCallback(callback C.DragFileNotifyCallback) {
	fmt.Println("SetDragFileNotifyCallback")
	g_DragFileNotifyCallback = callback
}

//export SetDragFileListNotifyCallback
func SetDragFileListNotifyCallback(callback C.DragFileListNotifyCallback) {
	log.Println("SetDragFileListNotifyCallback")
	g_DragFileListNotifyCallback = callback
}

//export SetMultiFilesDropNotifyCallback
func SetMultiFilesDropNotifyCallback(callback C.MultiFilesDropNotifyCallback) {
	log.Println("SetMultiFilesDropNotifyCallback")
	g_MultiFilesDropNotifyCallback = callback
}

//export SetUpdateMultipleProgressBarCallback
func SetUpdateMultipleProgressBarCallback(callback C.UpdateMultipleProgressBarCallback) {
	log.Println("SetUpdateMultipleProgressBarCallback")
	g_UpdateMultipleProgressBarCallback = callback
}

//export SetDataTransferCallback
func SetDataTransferCallback(callback C.DataTransferCallback) {
	log.Println("SetDataTransferCallback")
	g_DataTransferCallback = callback
}

//export SetUpdateProgressBarCallback
func SetUpdateProgressBarCallback(callback C.UpdateProgressBarCallback) {
	fmt.Println("SetUpdateProgressBarCallback")
	g_UpdateProgressBarCallback = callback
}

//export SetDeinitProgressBarCallback
func SetDeinitProgressBarCallback(callback C.DeinitProgressBarCallback) {
	fmt.Println("SetDeinitProgressBarCallback")
	//g_DeinitProgressBarCallback = callback
}

//export SetUpdateImageProgressBarCallback
func SetUpdateImageProgressBarCallback(callback C.UpdateImageProgressBarCallback) {
	fmt.Println("SetUpdateImageProgressBarCallback")
	g_UpdateImageProgressBarCallback = callback
}

//export SetUpdateClientStatusCallback
func SetUpdateClientStatusCallback(callback C.UpdateClientStatusCallback) {
	log.Println("SetUpdateClientStatusCallback")
	g_UpdateClientStatusCallback = callback
}

//export SetUpdateSystemInfoCallback
func SetUpdateSystemInfoCallback(callback C.UpdateSystemInfoCallback) {
	log.Println("SetUpdateSystemInfoCallback")
	g_UpdateSystemInfoCallback = callback
}

//export SetNotiMessageCallback
func SetNotiMessageCallback(callback C.NotiMessageCallback) {
	log.Println("SetNotiMessageCallback")
	g_NotiMessageCallback = callback
}

//export SetCleanClipboardCallback
func SetCleanClipboardCallback(callback C.CleanClipboardCallback) {
	fmt.Println("SetCleanClipboardCallback")
	g_CleanClipboardCallback = callback
}

//export SetAuthViaIndexCallback
func SetAuthViaIndexCallback(callback C.AuthViaIndexCallback) {
	log.Println("SetAuthViaIndexCallback")
	g_AuthViaIndexCallback = callback
}

//export SetDIASStatusCallback
func SetDIASStatusCallback(callback C.DIASStatusCallback) {
	log.Println("SetDIASStatusCallback")
	g_DIASStatusCallback = callback
}

//export SetRequestSourceAndPortCallback
func SetRequestSourceAndPortCallback(callback C.RequestSourceAndPortCallback) {
	log.Println("SetRequestSourceAndPortCallback")
	g_RequestSourceAndPortCallback = callback
}

//export SetSetupDstPasteImageCallback
func SetSetupDstPasteImageCallback(cb C.SetupDstPasteImageCallback) {
	log.Println("SetSetupDstPasteImageCallback")
	g_SetupDstPasteImageCallback = cb
}

//export SetRequestUpdateClientVersionCallback
func SetRequestUpdateClientVersionCallback(cb C.RequestUpdateClientVersionCallback) {
	log.Println("SetRequestUpdateClientVersionCallback")
	g_RequestUpdateClientVersionCallback = cb
}

//export SetNotifyErrEventCallback
func SetNotifyErrEventCallback(cb C.NotifyErrEventCallback) {
	log.Println("SetNotifyErrEventCallback")
	g_NotifyErrEventCallback = cb
}
