//go:build windows

package main

/*
#cgo LDFLAGS: -static -lstdc++
#include <stdint.h>
#include <wchar.h>
#include <stdlib.h>


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

typedef void (*FileListNotifyCallback)(const char *ipPort, const char *clientID, uint32_t cFileCount, uint64_t totalSize, uint64_t timestamp, const wchar_t *firstFileName, uint64_t firstFileSize,const char *fileDetails);
static void FileListSendNotifyCallbackFunc(FileListNotifyCallback cb, const char *ipPort, const char *clientID, uint32_t cFileCount, uint64_t totalSize, uint64_t timestamp, const wchar_t *firstFileName, uint64_t firstFileSize,const char *fileDetails) {
    if (cb) cb(ipPort, clientID, cFileCount, totalSize, timestamp, firstFileName, firstFileSize,fileDetails);
}
static void FileListReceiveNotifyCallbackFunc(FileListNotifyCallback cb, const char *ipPort, const char *clientID, uint32_t cFileCount, uint64_t totalSize, uint64_t timestamp, const wchar_t *firstFileName, uint64_t firstFileSize,const char *fileDetails) {
    if (cb) cb(ipPort, clientID, cFileCount, totalSize, timestamp, firstFileName, firstFileSize, fileDetails);
}

typedef void (*UpdateProgressBarCallback)(const char *ipPort, const char *clientID, const wchar_t *currentFileName, uint32_t sentFilesCnt, uint32_t totalFilesCnt, uint64_t currentFileSize, uint64_t totalSize, uint64_t sentSize, uint64_t timestamp);
static void UpdateSendProgressBarCallbackFunc(UpdateProgressBarCallback cb, const char *ipPort, const char *clientID, const wchar_t *currentFileName, uint32_t sentFilesCnt, uint32_t totalFilesCnt, uint64_t currentFileSize, uint64_t totalSize, uint64_t sentSize, uint64_t timestamp) {
    if (cb) cb(ipPort, clientID, currentFileName, sentFilesCnt, totalFilesCnt, currentFileSize, totalSize, sentSize, timestamp);
}
static void UpdateReceiveProgressBarCallbackFunc(UpdateProgressBarCallback cb, const char *ipPort, const char *clientID, const wchar_t *currentFileName, uint32_t sentFilesCnt, uint32_t totalFilesCnt, uint64_t currentFileSize, uint64_t totalSize, uint64_t sentSize, uint64_t timestamp) {
    if (cb) cb(ipPort, clientID, currentFileName, sentFilesCnt, totalFilesCnt, currentFileSize, totalSize, sentSize, timestamp);
}

typedef void (*UpdateClientStatusExCallback)(const char *clientJson);
static void UpdateClientStatusExCallbackFunc(UpdateClientStatusExCallback cb, const char *clientJson) {
    if (cb) cb(clientJson);
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

typedef void (*SetupDstPasteXClipDataCallback)(const char *text,const char *image,const char *html,const char *rtf);
static void SetupDstPasteXClipDataCallbackFunc(SetupDstPasteXClipDataCallback cb, const char *text,const char *image,const char *html,const char *rtf) {
    if (cb) cb(text, image, html,rtf);
}

typedef void (*RequestUpdateClientVersionCallback)(const char *clienVersion);
static void RequestUpdateClientVersionCallbackFunc(RequestUpdateClientVersionCallback cb, const char *clienVersion) {
    if (cb) cb(clienVersion);
}

typedef void (*NotifyErrEventCallback)(const char *clientID, uint32_t errCode, const char *arg1, const char *arg2, const char *arg3, const char *arg4);
static void NotifyErrEventCallbackFunc(NotifyErrEventCallback cb, const char *clienID, uint32_t errCode, const char *arg1, const char *arg2, const char *arg3, const char *arg4) {
    if (cb) cb(clienID, errCode, arg1, arg2, arg3, arg4);
}

typedef void (*ReadyReCtrlCallback)(const char *ip, uint32_t mousePort, uint32_t kybrdPort);
static void ReadyReCtrlCallbackFunc(ReadyReCtrlCallback cb, const char *ip, uint32_t mousePort, uint32_t kybrdPort) {
    if (cb) cb(ip, mousePort, kybrdPort);
}
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"log"
	rtkCmd "rtk-cross-share/client/cmd"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strings"
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
	g_FileListSendNotifyCallback         C.FileListNotifyCallback             = nil
	g_FileListReceiveNotifyCallback      C.FileListNotifyCallback             = nil
	g_UpdateSendProgressBarCallback      C.UpdateProgressBarCallback          = nil
	g_UpdateReceiveProgressBarCallback   C.UpdateProgressBarCallback          = nil
	g_UpdateClientStatusExCallback       C.UpdateClientStatusExCallback       = nil
	g_UpdateClientStatusCallback         C.UpdateClientStatusCallback         = nil
	g_UpdateSystemInfoCallback           C.UpdateSystemInfoCallback           = nil
	g_NotiMessageCallback                C.NotiMessageCallback                = nil
	g_CleanClipboardCallback             C.CleanClipboardCallback             = nil
	g_AuthViaIndexCallback               C.AuthViaIndexCallback               = nil
	g_DIASStatusCallback                 C.DIASStatusCallback                 = nil
	g_RequestSourceAndPortCallback       C.RequestSourceAndPortCallback       = nil
	g_SetupDstPasteXClipDataCallback     C.SetupDstPasteXClipDataCallback     = nil
	g_RequestUpdateClientVersionCallback C.RequestUpdateClientVersionCallback = nil
	g_NotifyErrEventCallback             C.NotifyErrEventCallback             = nil
	g_ReadyReCtrlCallback                C.ReadyReCtrlCallback                = nil
)

func main() {}

func init() {
	rtkPlatform.SetAuthViaIndexCallback(GoTriggerCallbackSetAuthViaIndex)
	rtkPlatform.SetDIASStatusCallback(GoTriggerCallbackSetDIASStatus)
	rtkPlatform.SetRequestSourceAndPortCallback(GoTriggerCallbackRequestSourceAndPort)
	rtkPlatform.SetUpdateSystemInfoCallback(GoTriggerCallbackUpdateSystemInfo)
	rtkPlatform.SetUpdateClientStatusCallback(GoTriggerCallbackUpdateClientStatus)
	rtkPlatform.SetUpdateClientStatusExCallback(GoTriggerCallbackUpdateClientStatusEx)
	rtkPlatform.SetPasteXClipCallback(GoTriggerCallbackSetupDstPasteXClipData)
	rtkPlatform.SetCleanClipboardCallback(GoTriggerCallbackCleanClipboard)
	rtkPlatform.SetFileListSendNotifyCallback(GoTriggerCallbackFileListSendNotify)
	rtkPlatform.SetFileListReceiveNotifyCallback(GoTriggerCallbackFileListReceiveNotify)
	rtkPlatform.SetSendProgressBarCallback(GoTriggerCallbackSendProgressBar)
	rtkPlatform.SetReceiveProgressBarCallback(GoTriggerCallbackReceiveProgressBar)
	rtkPlatform.SetNotiMessageFileTransCallback(GoTriggerCallbackNotiMessage)
	rtkPlatform.SetReqClientUpdateVerCallback(GoTriggerCallbackReqClientUpdateVer)
	rtkPlatform.SetNotifyErrEventCallback(GoTriggerCallbackNotifyErrEvent)
	rtkPlatform.SetReadyReCtrlCallback(GoTriggerCallbackReadyReCtrl)

	rtkPlatform.SetConfirmDocumentsAccept(false)
}

/*======================================= Go Call Windows API =======================================*/

func GoTriggerCallbackReadyReCtrl(ip string, mousePort, kybrdPort uint32) {
	if g_ReadyReCtrlCallback == nil {
		log.Printf("%s g_AuthViaIndexCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cMousePort := C.uint32_t(mousePort)
	cKybrdPort := C.uint32_t(kybrdPort)

	log.Printf("[%s] readyReCtrl Addr:[%s - %d %d]", rtkMisc.GetFuncInfo(), ip, mousePort, kybrdPort)
	C.ReadyReCtrlCallbackFunc(g_ReadyReCtrlCallback, cIp, cMousePort, cKybrdPort)
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

func GoTriggerCallbackUpdateClientStatusEx(clientInfo string) {
	if g_UpdateClientStatusExCallback == nil {
		log.Printf("%s g_UpdateClientStatusExCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), clientInfo)
	cClientInfo := C.CString(clientInfo)
	defer C.free(unsafe.Pointer(cClientInfo))

	C.UpdateClientStatusExCallbackFunc(g_UpdateClientStatusExCallback, cClientInfo)
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

func GoTriggerCallbackSetupDstPasteXClipData(text, image, html, rtf string) {
	if g_SetupDstPasteXClipDataCallback == nil {
		log.Printf("%s g_SetupDstPasteXClipDataCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	cText := C.CString(text)
	cImage := C.CString(image)
	cHtml := C.CString(html)
	cRtf := C.CString(rtf)

	defer func() {
		C.free(unsafe.Pointer(cText))
		C.free(unsafe.Pointer(cImage))
		C.free(unsafe.Pointer(cHtml))
		C.free(unsafe.Pointer(cRtf))
	}()

	log.Printf("[%s] text len:%d, image len:%d, html len:%d, rtf len:%d \n\n", rtkMisc.GetFuncInfo(), len(text), len(image), len(html), len(rtf))
	C.SetupDstPasteXClipDataCallbackFunc(g_SetupDstPasteXClipDataCallback, cText, cImage, cHtml, cRtf)
}

func GoTriggerCallbackCleanClipboard() {
	if g_CleanClipboardCallback == nil {
		log.Printf("%s g_CleanClipboardCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	log.Printf("[%s] trigger!", rtkMisc.GetFuncInfo())
	C.CleanClipboardCallbackFunc(g_CleanClipboardCallback)
}

func GoTriggerCallbackSendProgressBar(ip, id, currentFileName string, sendFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sendSize, timestamp uint64) {
	if g_UpdateSendProgressBarCallback == nil {
		log.Printf("%s g_UpdateSendProgressBarCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}

	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))

	cSentFileCnt := C.uint(sendFileCnt)
	cTotalFileCnt := C.uint(totalFileCnt)

	cCurrentFileSize := C.ulonglong(currentFileSize)
	cSentSize := C.ulonglong(sendSize)
	cTotalSize := C.ulonglong(totalSize)
	cTimestamp := C.ulonglong(timestamp)
	cCurrentFileName := GoStringToWChar(currentFileName)
	defer C.free(unsafe.Pointer(cCurrentFileName))

	C.UpdateSendProgressBarCallbackFunc(g_UpdateSendProgressBarCallback, cIp, cId, cCurrentFileName, cSentFileCnt, cTotalFileCnt, cCurrentFileSize, cTotalSize, cSentSize, cTimestamp)
}

func GoTriggerCallbackReceiveProgressBar(ip, id, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	if g_UpdateReceiveProgressBarCallback == nil {
		log.Printf("%s g_UpdateReceiveProgressBarCallback is not set!", rtkMisc.GetFuncInfo())
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

	C.UpdateReceiveProgressBarCallbackFunc(g_UpdateReceiveProgressBarCallback, cIp, cId, cCurrentFileName, cSentFileCnt, cTotalFileCnt, cCurrentFileSize, cTotalSize, cSentSize, cTimestamp)
}

func GoTriggerCallbackFileListSendNotify(ip, id string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64, fileDetails string) {
	if g_FileListSendNotifyCallback == nil {
		log.Printf("%s g_FileListSendNotifyCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))
	cFileDetails := C.CString(fileDetails)
	defer C.free(unsafe.Pointer(cFileDetails))

	cFileCnt := C.uint(fileCnt)
	cToalSize := C.ulonglong(totalSize)
	cFirstFileSize := C.ulonglong(firstFileSize)
	cTimestamp := C.ulonglong(timestamp)
	cFirstFileName := GoStringToWChar(firstFileName)
	defer C.free(unsafe.Pointer(cFirstFileName))

	log.Printf("[%s] ID:[%s] IP:[%s] cnt:[%d] total:[%d] timestamp:[%d] firstFile:[%s] firstSize:[%d] %s", rtkMisc.GetFuncInfo(), id, ip, fileCnt, totalSize, timestamp, firstFileName, firstFileSize, fileDetails)
	C.FileListSendNotifyCallbackFunc(g_FileListSendNotifyCallback, cIp, cId, cFileCnt, cToalSize, cTimestamp, cFirstFileName, cFirstFileSize, cFileDetails)
}

func GoTriggerCallbackFileListReceiveNotify(ip, id string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64, fileDetails string) {
	if g_FileListReceiveNotifyCallback == nil {
		log.Printf("%s g_FileListReceiveNotifyCallback is not set!", rtkMisc.GetFuncInfo())
		return
	}
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))
	cId := C.CString(id)
	defer C.free(unsafe.Pointer(cId))
	cFileDetails := C.CString(fileDetails)
	defer C.free(unsafe.Pointer(cFileDetails))

	cFileCnt := C.uint(fileCnt)
	cToalSize := C.ulonglong(totalSize)
	cFirstFileSize := C.ulonglong(firstFileSize)
	cTimestamp := C.ulonglong(timestamp)
	cFirstFileName := GoStringToWChar(firstFileName)
	defer C.free(unsafe.Pointer(cFirstFileName))

	log.Printf("[%s] ID:[%s] IP:[%s] cnt:[%d] total:[%d] timestamp:[%d] firstFile:[%s] firstSize:[%d]", rtkMisc.GetFuncInfo(), id, ip, fileCnt, totalSize, timestamp, firstFileName, firstFileSize)
	C.FileListReceiveNotifyCallbackFunc(g_FileListReceiveNotifyCallback, cIp, cId, cFileCnt, cToalSize, cTimestamp, cFirstFileName, cFirstFileSize, cFileDetails)
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

/*======================================= Windows Call Go API =======================================*/

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

//export SendXClipData
func SendXClipData(cText, cImage, cHtml, cRtf *C.char) {
	text := C.GoString(cText)
	image := C.GoString(cImage)
	html := C.GoString(cHtml)
	rtf := C.GoString(cRtf)

	log.Printf("[%s] text:%d, image:%d, html:%d, rtf:%d \n\n", rtkMisc.GetFuncInfo(), len(text), len(image), len(html), len(rtf))
	rtkPlatform.GoCopyXClipData(text, image, html, rtf)
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
	//rtkPlatform.GoSetMacAddress(macAddress)
}

//export  SetDisplayEventInfo
func SetDisplayEventInfo(cDisplayInfoJson *C.char) {
	displayInfoJson := C.GoString(cDisplayInfoJson)

	var displayEventInfo rtkCommon.DisplayEventInfo
	err := json.Unmarshal([]byte(displayInfoJson), &displayEventInfo)
	if err != nil {
		log.Printf("[%s] Unmarshal[%s] err:%+v", rtkMisc.GetFuncInfo(), displayInfoJson, err)
		return
	}

	log.Printf("[%s] PlugEvent:[%d] MacAddr:[%s] source:[%s] port:[%d] UdpMousePort:[%s] UdpKeyboardPort:[%s]", rtkMisc.GetFuncInfo(), displayEventInfo.PlugEvent,
		displayEventInfo.MacAddr, displayEventInfo.Source, displayEventInfo.Port, displayEventInfo.UdpMousePort, displayEventInfo.UdpKeyboardPort)

	if displayEventInfo.PlugEvent != 0 && displayEventInfo.PlugEvent != 1 {
		displayEventInfo.PlugEvent = 1
	}
	rtkPlatform.GoSetDisplayEvent(&displayEventInfo)
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

//export SetConfirmDocumentsAccept
func SetConfirmDocumentsAccept(ifConfirm bool) {
	log.Printf("[%s], ifConfirm:[%+v]", rtkMisc.GetFuncInfo(), ifConfirm)
	//rtkPlatform.SetConfirmDocumentsAccept(ifConfirm)
}

//export SetDragFileListRequest
func SetDragFileListRequest(filePathArry **C.wchar_t, arryLength C.uint32_t, timeStamp C.uint64_t) C.uint {
	timestamp := uint64(timeStamp)
	fileCount := uint32(arryLength)
	fileList := make([]string, 0)

	for i := uint32(0); i < fileCount; i++ {
		wcharPtr := *(**C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(filePathArry)) + uintptr(i)*unsafe.Sizeof(*filePathArry)))
		file := WCharToGoString(wcharPtr)
		file = strings.ReplaceAll(file, "/", "\\")
		fileList = append(fileList, file)
	}

	return C.uint(rtkPlatform.GoDragFileListRequest(&fileList, timestamp))
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
	id := C.GoString(clientID)
	ip := C.GoString(ipPort)
	timestamp := uint64(timeStamp)
	fileCount := uint32(arryLength)
	fileList := make([]string, 0)

	for i := uint32(0); i < fileCount; i++ {
		wcharPtr := *(**C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(filePathArry)) + uintptr(i)*unsafe.Sizeof(*filePathArry)))
		file := WCharToGoString(wcharPtr)
		file = strings.ReplaceAll(file, "/", "\\")
		fileList = append(fileList, file)
	}

	return C.uint(rtkPlatform.GoMultiFilesDropRequest(id, ip, &fileList, timestamp))
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

//export GetClientList
func GetClientList() *C.char {
	clientList := rtkUtils.GetClientListEx()
	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), clientList)
	cClientListJson := C.CString(clientList)
	defer C.free(unsafe.Pointer(cClientListJson))

	return cClientListJson
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

/*======================================= Windows set Go Callback =======================================*/

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

//export SetFileListSendNotifyCallback
func SetFileListSendNotifyCallback(callback C.FileListNotifyCallback) {
	log.Println("SetFileListSendNotifyCallback")
	g_FileListSendNotifyCallback = callback
}

//export SetFileListReceiveNotifyCallback
func SetFileListReceiveNotifyCallback(callback C.FileListNotifyCallback) {
	log.Println("SetFileListReceiveNotifyCallback")
	g_FileListReceiveNotifyCallback = callback
}

//export SetUpdateSendProgressBarCallback
func SetUpdateSendProgressBarCallback(callback C.UpdateProgressBarCallback) {
	log.Println("SetUpdateSendProgressBarCallback")
	g_UpdateSendProgressBarCallback = callback
}

//export SetUpdateReceiveProgressBarCallback
func SetUpdateReceiveProgressBarCallback(callback C.UpdateProgressBarCallback) {
	log.Println("SetUpdateReceiveProgressBarCallback")
	g_UpdateReceiveProgressBarCallback = callback
}

//export SetUpdateClientStatusExCallback
func SetUpdateClientStatusExCallback(callback C.UpdateClientStatusExCallback) {
	log.Println("SetUpdateClientStatusExCallback")
	g_UpdateClientStatusExCallback = callback
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

//export SetSetupDstPasteXClipDataCallback
func SetSetupDstPasteXClipDataCallback(cb C.SetupDstPasteXClipDataCallback) {
	log.Println("SetSetupDstPasteXClipDataCallback")
	g_SetupDstPasteXClipDataCallback = cb
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

//export SetReadyReCtrlCallback
func SetReadyReCtrlCallback(cb C.ReadyReCtrlCallback) {
	log.Println("SetReadyReCtrlCallback")
	g_ReadyReCtrlCallback = cb
}
