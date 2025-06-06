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

typedef const wchar_t* (*GetDeviceNameCallback)();
static const wchar_t* GetDeviceNameCallbackFunc(GetDeviceNameCallback cb) {
    if (cb) {
		return cb();
	} else {
	 	return NULL;
	}
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

typedef const wchar_t* (*GetDownloadPathCallback)();
static const wchar_t* GetDownloadPathCallbackFunc(GetDownloadPathCallback cb) {
        if (cb) {
		return cb();
	} else {
	 	return NULL;
	}
}

typedef void (*SetupDstPasteImageCallback)(const wchar_t* desc, IMAGE_HEADER imgHeader, uint32_t dataSize);
static void SetupDstPasteImageCallbackFunc(SetupDstPasteImageCallback cb, const wchar_t* desc, IMAGE_HEADER imgHeader, uint32_t dataSize) {
    if (cb) cb(desc, imgHeader, dataSize);
}
*/
import "C"
import (
	"fmt"
	"log"
	rtkCmd "rtk-cross-share/client/cmd"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkMisc "rtk-cross-share/misc"
	"unsafe"
)

func WCharToGoString(wstr *C.wchar_t) string {
	var goStr string
	for ptr := wstr; *ptr != 0; ptr = (*C.wchar_t)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + unsafe.Sizeof(*ptr))) {
		goStr += string(rune(*ptr))
	}
	return goStr
}

func GoStringToWChar(str string) *C.wchar_t {
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

var (
	g_StartClipboardMonitorCallback     C.StartClipboardMonitorCallback     = nil
	g_StopClipboardMonitorCallback      C.StopClipboardMonitorCallback      = nil
	g_DragFileNotifyCallback            C.DragFileNotifyCallback            = nil
	g_DragFileListNotifyCallback        C.DragFileListNotifyCallback        = nil
	g_MultiFilesDropNotifyCallback      C.MultiFilesDropNotifyCallback      = nil
	g_UpdateMultipleProgressBarCallback C.UpdateMultipleProgressBarCallback = nil
	g_DataTransferCallback              C.DataTransferCallback              = nil
	g_UpdateProgressBarCallback         C.UpdateProgressBarCallback         = nil
	g_UpdateImageProgressBarCallback    C.UpdateImageProgressBarCallback    = nil
	g_UpdateClientStatusCallback        C.UpdateClientStatusCallback        = nil
	g_UpdateSystemInfoCallback          C.UpdateSystemInfoCallback          = nil
	g_NotiMessageCallback               C.NotiMessageCallback               = nil
	g_CleanClipboardCallback            C.CleanClipboardCallback            = nil
	g_GetDeviceNameCallback             C.GetDeviceNameCallback             = nil
	g_AuthViaIndexCallback              C.AuthViaIndexCallback              = nil
	g_DIASStatusCallback                C.DIASStatusCallback                = nil
	g_RequestSourceAndPortCallback      C.RequestSourceAndPortCallback      = nil
	g_GetDownloadPathCallback           C.GetDownloadPathCallback           = nil
	g_SetupDstPasteImageCallback        C.SetupDstPasteImageCallback        = nil
)

func main() {}

func init() {
	rtkPlatform.SetAuthViaIndexCallback(GoTriggerCallbackSetAuthViaIndex)
	rtkPlatform.SetDIASStatusCallback(GoTriggerCallbackSetDIASStatus)
	rtkPlatform.SetDeviceNameCallback(GoTriggerCallbackGetDeviceName)
	rtkPlatform.SetDownloadPathCallback(GoTriggerCallbackGetDownloadPath)
	rtkPlatform.SetRequestSourceAndPortCallback(GoTriggerCallbackRequestSourceAndPort)
	rtkPlatform.SetUpdateSystemInfoCallback(GoTriggerCallbackUpdateSystemInfo)
	rtkPlatform.SetUpdateClientStatusCallback(GoTriggerCallbackUpdateClientStatus)
	rtkPlatform.SetImageDataTransferCallback(GoTriggerCallbackImageDataTransfer)
	rtkPlatform.SetSetupDstImageCallback(GoTriggerCallbackSetupDstPasteImage)

	rtkPlatform.SetDragFileListNotifyCallback(GoTriggerCallbackDragFileListNotify)
	rtkPlatform.SetConfirmDocumentsAccept(false)

	/*C.MultiFilesDropNotifyCallbackFunc(g_MultiFilesDropNotifyCallback, ipPort, clientID, 2, 2048, 1625097600, fileName, 1024)
	C.UpdateMultipleProgressBarCallbackFunc(g_UpdateMultipleProgressBarCallback, ipPort, clientID, fileName, 1, 2, 1024, 2048, 512, 1625097600)

	C.NotiMessageCallbackFunc(g_NotiMessageCallback, 1625097600, 2001, (**C.wchar_t)(unsafe.Pointer(&notiParams[0])), 2)
	C.CleanClipboardCallbackFunc(g_CleanClipboardCallback)
	*/
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

func GoTriggerCallbackGetDeviceName() string {
	if g_GetDeviceNameCallback == nil {
		log.Printf("%s g_GetDeviceNameCallback is not set!", rtkMisc.GetFuncInfo())
		return ""
	}

	deviceName := WCharToGoString(C.GetDeviceNameCallbackFunc(g_GetDeviceNameCallback))
	log.Printf("[%s] get device name:[%s]", rtkMisc.GetFuncInfo(), deviceName)
	return deviceName
}

func GoTriggerCallbackGetDownloadPath() string {
	if g_GetDownloadPathCallback == nil {
		log.Printf("%s g_GetDownloadPathCallback is not set!", rtkMisc.GetFuncInfo())
		return ""
	}
	path := WCharToGoString(C.GetDownloadPathCallbackFunc(g_GetDownloadPathCallback))
	log.Printf("[%s] get download path:[%s]", rtkMisc.GetFuncInfo(), path)
	return path
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

//export InitGoServer
func InitGoServer() {
	rtkPlatform.InitPlatform()
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

//export SetSetupDstPasteImageCallback
func SetSetupDstPasteImageCallback(cb C.SetupDstPasteImageCallback) {
	fmt.Println("SetSetupDstPasteImageCallback")
	g_SetupDstPasteImageCallback = cb
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
	fmt.Printf("SetMacAddress(%q, %d)\n", C.GoString(cMacAddress), length)
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

//export SetDragFileListRequest
func SetDragFileListRequest(filePathArry **C.wchar_t, arryLength C.uint32_t, timeStamp C.uint64_t) {
	paths := (*[1 << 30]*C.wchar_t)(unsafe.Pointer(filePathArry))[:arryLength:arryLength]
	goPaths := make([]string, arryLength)
	for i := range goPaths {
		goPaths[i] = WCharToGoString(paths[i])
	}
	fmt.Printf("SetDragFileListRequest(%v, %d, %d)\n", goPaths, arryLength, timeStamp)
}

//export SetCancelFileTransfer
func SetCancelFileTransfer(ipPort *C.char, clientID *C.char, timeStamp C.uint64_t) {
	fmt.Printf("SetCancelFileTransfer(%q, %q, %d)\n",
		C.GoString(ipPort),
		C.GoString(clientID),
		timeStamp)

	cIpAddr := C.GoString(ipPort)
	cId := C.GoString(clientID)
	cTimestamp := int64(timeStamp)

	rtkPlatform.GoCancelFileTrans(cIpAddr, cId, cTimestamp)
}

//export SetMultiFilesDropRequest
func SetMultiFilesDropRequest(ipPort *C.char, clientID *C.char, timeStamp C.uint64_t, filePathArry **C.wchar_t, arryLength C.uint32_t) {
	paths := (*[1 << 30]*C.wchar_t)(unsafe.Pointer(filePathArry))[:arryLength:arryLength]
	goPaths := make([]string, arryLength)
	for i := range goPaths {
		goPaths[i] = WCharToGoString(paths[i])
	}
	fmt.Printf("SetMultiFilesDropRequest(%q, %q, %d, %v, %d)\n",
		C.GoString(ipPort),
		C.GoString(clientID),
		timeStamp,
		goPaths,
		arryLength)
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
	fmt.Println("SetDragFileListNotifyCallback")
	g_DragFileListNotifyCallback = callback
}

//export SetMultiFilesDropNotifyCallback
func SetMultiFilesDropNotifyCallback(callback C.MultiFilesDropNotifyCallback) {
	fmt.Println("SetMultiFilesDropNotifyCallback")
	g_MultiFilesDropNotifyCallback = callback
}

//export SetUpdateMultipleProgressBarCallback
func SetUpdateMultipleProgressBarCallback(callback C.UpdateMultipleProgressBarCallback) {
	fmt.Println("SetUpdateMultipleProgressBarCallback")
	g_UpdateMultipleProgressBarCallback = callback
}

//export SetDataTransferCallback
func SetDataTransferCallback(callback C.DataTransferCallback) {
	fmt.Println("SetDataTransferCallback")
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
	fmt.Println("SetUpdateClientStatusCallback")
	g_UpdateClientStatusCallback = callback
}

//export SetUpdateSystemInfoCallback
func SetUpdateSystemInfoCallback(callback C.UpdateSystemInfoCallback) {
	fmt.Println("SetUpdateSystemInfoCallback")
	g_UpdateSystemInfoCallback = callback
}

//export SetNotiMessageCallback
func SetNotiMessageCallback(callback C.NotiMessageCallback) {
	fmt.Println("SetNotiMessageCallback")
	g_NotiMessageCallback = callback
}

//export SetCleanClipboardCallback
func SetCleanClipboardCallback(callback C.CleanClipboardCallback) {
	fmt.Println("SetCleanClipboardCallback")
	g_CleanClipboardCallback = callback
}

//export SetGetDeviceNameCallback
func SetGetDeviceNameCallback(callback C.GetDeviceNameCallback) {
	fmt.Println("SetGetDeviceNameCallback")
	g_GetDeviceNameCallback = callback
}

//export SetAuthViaIndexCallback
func SetAuthViaIndexCallback(callback C.AuthViaIndexCallback) {
	fmt.Println("SetAuthViaIndexCallback")
	g_AuthViaIndexCallback = callback
}

//export SetDIASStatusCallback
func SetDIASStatusCallback(callback C.DIASStatusCallback) {
	fmt.Println("SetDIASStatusCallback")
	g_DIASStatusCallback = callback
}

//export SetRequestSourceAndPortCallback
func SetRequestSourceAndPortCallback(callback C.RequestSourceAndPortCallback) {
	fmt.Println("SetRequestSourceAndPortCallback")
	g_RequestSourceAndPortCallback = callback
}

//export SetGetDownloadPathCallback
func SetGetDownloadPathCallback(callback C.GetDownloadPathCallback) {
	fmt.Println("SetGetDownloadPathCallback")
	g_GetDownloadPathCallback = callback
}
