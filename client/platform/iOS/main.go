//go:build ios
// +build ios

package main

/*
#include <stdlib.h>
#include <stdint.h>

typedef enum NOTI_MSG_CODE
{
    NOTI_MSG_CODE_CONN_STATUS_SUCCESS       = 1,
    NOTI_MSG_CODE_CONN_STATUS_FAIL          = 2,
    NOTI_MSG_CODE_FILE_TRANS_DONE_SENDER    = 3,
    NOTI_MSG_CODE_FILE_TRANS_DONE_RECEIVER  = 4,
    NOTI_MSG_CODE_FILE_TRANS_REJECT         = 5,
} NOTI_MSG_CODE;

typedef void (*CallbackUpdateClientStatus)(char* clientJsonStr);
typedef void (*CallbackMethodFileListNotify)(char* ip, char* id, unsigned int fileCnt, unsigned long long totalSize,unsigned long long timestamp, char* firstFileName, unsigned long long firstFileSize, char*  fileDetails);
typedef void (*CallbackUpdateProgressBar)(char* ip,char* id, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp);
typedef void (*CallbackNotiMessage)(unsigned long long timestamp, unsigned int notiCode, char* notiParam[], int paramCount);
typedef void (*CallbackMethodStartBrowseMdns)(char* instance, char* serviceType);
typedef void (*CallbackMethodStopBrowseMdns)();
typedef char* (*CallbackAuthData)(unsigned int clientIndex);
typedef void (*CallbackSetDIASStatus)(unsigned int status);
typedef void (*CallbackSetMonitorName)(char* monitorName);
typedef void (*CallbackPasteXClipData)(char *text, char *image, char *html, char* rtf);
typedef void (*CallbackRequestUpdateClientVersion)(char* clientVer);
typedef void (*CallbackNotifyErrEvent)(char* id, unsigned int errCode, char* arg1, char* arg2, char* arg3, char* arg4);
typedef void (*CallbackNotifyBrowseResult)(char* monitorName, char* instance, char* ip, char* version, unsigned long long timestamp);

static CallbackUpdateClientStatus gCallbackUpdateClientStatus = 0;
static CallbackMethodFileListNotify gCallbackFileListSendNotify = 0;
static CallbackMethodFileListNotify gCallbackFileListReceiveNotify = 0;
static CallbackUpdateProgressBar gCallbackUpdateSendProgressBar = 0;
static CallbackUpdateProgressBar gCallbackUpdateReceiveProgressBar = 0;
static CallbackNotiMessage gCallbackNotiMessage = 0;
static CallbackMethodStartBrowseMdns gCallbackMethodStartBrowseMdns = 0;
static CallbackMethodStopBrowseMdns gCallbackMethodStopBrowseMdns = 0;
static CallbackAuthData gCallbackAuthData = 0;
static CallbackSetDIASStatus gCallbackSetDIASStatus = 0;
static CallbackSetMonitorName gCallbackSetMonitorName = 0;
static CallbackPasteXClipData gCallbackPasteXClipData = 0;
static CallbackRequestUpdateClientVersion gCallbackRequestUpdateClientVersion = 0;
static CallbackNotifyErrEvent gCallbackNotifyErrEvent = 0;
static CallbackNotifyBrowseResult gCallbackNotifyBrowseResult = 0;


static void setCallbackUpdateClientStatus(CallbackUpdateClientStatus cb) {gCallbackUpdateClientStatus = cb;}
static void invokeCallbackUpdateClientStatus(char* clientJsonStr) {
	if (gCallbackUpdateClientStatus) {gCallbackUpdateClientStatus(clientJsonStr);}
}
static void setCallbackFileListSendNotify(CallbackMethodFileListNotify cb) {gCallbackFileListSendNotify = cb;}
static void invokeCallbackFileListSendNotify(char* ip, char* id, unsigned int fileCnt, unsigned long long totalSize,unsigned long long timestamp, char* firstFileName, unsigned long long firstFileSize, char*  fileDetails) {
	if (gCallbackFileListSendNotify) {gCallbackFileListSendNotify(ip, id, fileCnt, totalSize, timestamp, firstFileName, firstFileSize, fileDetails);}
}
static void setCallbackFileListReceiveNotify(CallbackMethodFileListNotify cb) {gCallbackFileListReceiveNotify = cb;}
static void invokeCallbackFileListReceiveNotify(char* ip, char* id, unsigned int fileCnt, unsigned long long totalSize,unsigned long long timestamp, char* firstFileName, unsigned long long firstFileSize, char*  fileDetails) {
	if (gCallbackFileListReceiveNotify) {gCallbackFileListReceiveNotify(ip, id, fileCnt, totalSize, timestamp, firstFileName, firstFileSize, fileDetails);}
}
static void setCallbackUpdateSendProgressBar(CallbackUpdateProgressBar cb) {gCallbackUpdateSendProgressBar = cb;}
static void invokeCallbackUpdateSendProgressBar(char* ip,char* id, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp) {
	if (gCallbackUpdateSendProgressBar) {gCallbackUpdateSendProgressBar(ip,id,currentfileName,recvFileCnt,totalFileCnt,currentFileSize,totalSize, recvSize, timestamp);}
}
static void setCallbackUpdateReceiveProgressBar(CallbackUpdateProgressBar cb) {gCallbackUpdateReceiveProgressBar = cb;}
static void invokeCallbackUpdateReceiveProgressBar(char* ip,char* id, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp) {
	if (gCallbackUpdateReceiveProgressBar) {gCallbackUpdateReceiveProgressBar(ip,id,currentfileName,recvFileCnt,totalFileCnt,currentFileSize,totalSize, recvSize, timestamp);}
}
static void setCallbackNotiMessage(CallbackNotiMessage cb) {gCallbackNotiMessage = cb;}
static void invokeCallbackNotiMessage(unsigned long long timestamp, unsigned int notiCode, char* notiParam[], int paramCount) {
	if (gCallbackNotiMessage) {gCallbackNotiMessage(timestamp, notiCode, notiParam, paramCount);}
}
static void setCallbackMethodStartBrowseMdns(CallbackMethodStartBrowseMdns cb) {gCallbackMethodStartBrowseMdns = cb;}
static void invokeCallbackMethodStartBrowseMdns(char* instance, char* serviceType) {
	if (gCallbackMethodStartBrowseMdns) {gCallbackMethodStartBrowseMdns(instance, serviceType);}
}
static void setCallbackMethodStopBrowseMdns(CallbackMethodStopBrowseMdns cb) {gCallbackMethodStopBrowseMdns = cb;}
static void invokeCallbackMethodStopBrowseMdns() {
	if (gCallbackMethodStopBrowseMdns) {gCallbackMethodStopBrowseMdns();}
}
static void setCallbackGetAuthData(CallbackAuthData cb) {gCallbackAuthData = cb;}
static char* invokeCallbackGetAuthData(unsigned int clientIndex) {
	if (gCallbackAuthData) { return gCallbackAuthData(clientIndex);}
    return NULL;
}
static void setCallbackSetDIASStatus(CallbackSetDIASStatus cb) {gCallbackSetDIASStatus = cb;}
static void invokeCallbackSetDIASStatus(unsigned int status) {
	if (gCallbackSetDIASStatus) { gCallbackSetDIASStatus(status);}
}
static void setCallbackSetMonitorName(CallbackSetMonitorName cb) {gCallbackSetMonitorName = cb;}
static void invokeCallbackSetMonitorName(char* monitorName) {
	if (gCallbackSetMonitorName) { gCallbackSetMonitorName(monitorName);}
}
static void setCallbackPasteXClipData(CallbackPasteXClipData cb) {gCallbackPasteXClipData = cb;}
static void invokeCallbackPasteXClipData(char *text, char *image, char *html, char* rtf) {
	if (gCallbackPasteXClipData) { gCallbackPasteXClipData(text, image, html, rtf);}
}
static void setCallbackRequestUpdateClientVersion(CallbackRequestUpdateClientVersion cb) {gCallbackRequestUpdateClientVersion = cb;}
static void invokeCallbackRequestUpdateClientVersion(char* clientVer) {
	if (gCallbackRequestUpdateClientVersion) { gCallbackRequestUpdateClientVersion(clientVer);}
}
static void setCallbackNotifyErrEvent(CallbackNotifyErrEvent cb) {gCallbackNotifyErrEvent = cb;}
static void invokeCallbackNotifyErrEvent(char* id, unsigned int errCode, char* arg1, char* arg2, char* arg3, char* arg4) {
	if (gCallbackNotifyErrEvent) { gCallbackNotifyErrEvent(id,errCode,arg1,arg2,arg3,arg4);}
}
static void setCallbackNotifyBrowseResult(CallbackNotifyBrowseResult cb) {gCallbackNotifyBrowseResult = cb;}
static void invokeCallbackNotifyBrowseResult(char* monitorName, char* instance, char* ip, char* version, unsigned long long timestamp) {
	if (gCallbackNotifyBrowseResult) { gCallbackNotifyBrowseResult(monitorName, instance, ip, version, timestamp);}
}
*/
import "C"

import (
	"log"
	rtkBuildConfig "rtk-cross-share/client/buildConfig"
	rtkCmd "rtk-cross-share/client/cmd"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strings"
	"unsafe"
)

type CharArray struct {
	Array **C.char
	items []*C.char
}

func NewCharArray(strs []string) *CharArray {
	ptrSize := unsafe.Sizeof(uintptr(0))
	cArray := C.malloc(C.size_t(len(strs)+1) * C.size_t(ptrSize))
	if cArray == nil {
		panic("C.malloc failed")
	}

	cStrs := make([]*C.char, len(strs))
	for i, s := range strs {
		cStr := C.CString(s)
		cStrs[i] = cStr
		elemPtr := (**C.char)(unsafe.Pointer(uintptr(cArray) + uintptr(i)*ptrSize))
		*elemPtr = cStr
	}

	endPtr := (**C.char)(unsafe.Pointer(uintptr(cArray) + uintptr(len(strs))*ptrSize))
	*endPtr = nil

	return &CharArray{
		Array: (**C.char)(cArray),
		items: cStrs,
	}
}

func (a *CharArray) Free() {
	for _, ptr := range a.items {
		C.free(unsafe.Pointer(ptr))
	}
	C.free(unsafe.Pointer(a.Array))
}

func main() {
}

func init() {
	rtkPlatform.SetCallbackUpdateClientStatus(GoTriggerCallbackUpdateClientStatus)
	rtkPlatform.SetCallbackFileListSendNotify(GoTriggerCallbackFileListSendNotify)
	rtkPlatform.SetCallbackFileListReceiveNotify(GoTriggerCallbackFileListReceiveNotify)
	rtkPlatform.SetCallbackUpdateSendProgressBar(GoTriggerCallbackUpdateSendProgressBar)
	rtkPlatform.SetCallbackUpdateReceiveProgressBar(GoTriggerCallbackUpdateReceiveProgressBar)
	rtkPlatform.SetCallbackNotiMessageFileTrans(GoTriggerCallbackNotiMessage)
	rtkPlatform.SetCallbackMethodStartBrowseMdns(GoTriggerCallbackMethodStartBrowseMdns)
	rtkPlatform.SetCallbackMethodStopBrowseMdns(GoTriggerCallbackMethodStopBrowseMdns)
	rtkPlatform.SetCallbackGetAuthData(GoTriggerCallbackGetAuthData)
	rtkPlatform.SetCallbackDIASStatus(GoTriggerCallbackSetDIASStatus)
	rtkPlatform.SetCallbackMonitorName(GoTriggerCallbackSetMonitorName)
	rtkPlatform.SetCallbackPasteXClipData(GoTriggerCallbackPasteXClipData)
	rtkPlatform.SetCallbackRequestUpdateClientVersion(GoTriggerCallbackReqClientUpdateVer)
	rtkPlatform.SetCallbackNotifyErrEvent(GoTriggerCallbackNotifyErrEvent)
	rtkPlatform.SetCallbackNotifyBrowseResult(GoTriggerCallbackNotifyBrowseResult)

	rtkPlatform.SetConfirmDocumentsAccept(false)
}

func GoTriggerCallbackUpdateClientStatus(clientInfo string) {
	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), clientInfo)
	cClientInfo := C.CString(clientInfo)
	defer C.free(unsafe.Pointer(cClientInfo))

	C.invokeCallbackUpdateClientStatus(cClientInfo)
}

func GoTriggerCallbackFileListSendNotify(ip, id string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64, fileDetails string) {
	cip := C.CString(ip)
	cid := C.CString(id)
	cfirstFileName := C.CString(firstFileName)
	cFileDetails := C.CString(fileDetails)

	defer func() {
		C.free(unsafe.Pointer(cip))
		C.free(unsafe.Pointer(cid))
		C.free(unsafe.Pointer(cfirstFileName))
		C.free(unsafe.Pointer(cFileDetails))
	}()

	cFileCnt := C.uint(fileCnt)
	ctotalSize := C.ulonglong(totalSize)
	ctimeStamp := C.ulonglong(timestamp)
	cfirstFileSize := C.ulonglong(firstFileSize)

	log.Printf("[%s] (SRC) dst id:%s ip:[%s] fileCnt:%d totalSize:%d firstFileName:%s firstFileSize:%d", rtkMisc.GetFuncInfo(), id, ip, fileCnt, totalSize, firstFileName, firstFileSize)
	C.invokeCallbackFileListSendNotify(cip, cid, cFileCnt, ctotalSize, ctimeStamp, cfirstFileName, cfirstFileSize, cFileDetails)
}

func GoTriggerCallbackFileListReceiveNotify(ip, id string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64, fileDetails string) {
	cip := C.CString(ip)
	cid := C.CString(id)
	cfirstFileName := C.CString(firstFileName)
	cFileDetails := C.CString(fileDetails)

	defer func() {
		C.free(unsafe.Pointer(cip))
		C.free(unsafe.Pointer(cid))
		C.free(unsafe.Pointer(cfirstFileName))
		C.free(unsafe.Pointer(cFileDetails))
	}()

	cFileCnt := C.uint(fileCnt)
	ctotalSize := C.ulonglong(totalSize)
	ctimeStamp := C.ulonglong(timestamp)
	cfirstFileSize := C.ulonglong(firstFileSize)

	log.Printf("[%s] (DST) src id:%s ip:[%s] fileCnt:%d totalSize:%d firstFileName:%s firstFileSize:%d", rtkMisc.GetFuncInfo(), id, ip, fileCnt, totalSize, firstFileName, firstFileSize)
	C.invokeCallbackFileListReceiveNotify(cip, cid, cFileCnt, ctotalSize, ctimeStamp, cfirstFileName, cfirstFileSize, cFileDetails)
}

func GoTriggerCallbackUpdateSendProgressBar(ip, id, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	cip := C.CString(ip)
	cid := C.CString(id)
	ccurrentFileName := C.CString(currentFileName)

	defer func() {
		C.free(unsafe.Pointer(cip))
		C.free(unsafe.Pointer(cid))
		C.free(unsafe.Pointer(ccurrentFileName))
	}()

	crecvFileCnt := C.uint(sentFileCnt)
	ctotalFileCnt := C.uint(totalFileCnt)

	ccurrentFileSize := C.ulonglong(currentFileSize)
	ctotalSize := C.ulonglong(totalSize)
	crecvSize := C.ulonglong(sentSize)
	ctimeStamp := C.ulonglong(timestamp)

	C.invokeCallbackUpdateSendProgressBar(cip, cid, ccurrentFileName, crecvFileCnt, ctotalFileCnt, ccurrentFileSize, ctotalSize, crecvSize, ctimeStamp)
}

func GoTriggerCallbackUpdateReceiveProgressBar(ip, id, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	cip := C.CString(ip)
	cid := C.CString(id)
	ccurrentFileName := C.CString(currentFileName)

	defer func() {
		C.free(unsafe.Pointer(cip))
		C.free(unsafe.Pointer(cid))
		C.free(unsafe.Pointer(ccurrentFileName))
	}()

	crecvFileCnt := C.uint(sentFileCnt)
	ctotalFileCnt := C.uint(totalFileCnt)

	ccurrentFileSize := C.ulonglong(currentFileSize)
	ctotalSize := C.ulonglong(totalSize)
	crecvSize := C.ulonglong(sentSize)
	ctimeStamp := C.ulonglong(timestamp)

	C.invokeCallbackUpdateReceiveProgressBar(cip, cid, ccurrentFileName, crecvFileCnt, ctotalFileCnt, ccurrentFileSize, ctotalSize, crecvSize, ctimeStamp)
}

func GoTriggerCallbackNotiMessage(fileName, clientName, platform string, timestamp uint64, isSender bool) {
	cTimestamp := C.ulonglong(timestamp)
	cCode := C.uint(C.NOTI_MSG_CODE_FILE_TRANS_DONE_RECEIVER)
	if isSender {
		cCode = C.uint(C.NOTI_MSG_CODE_FILE_TRANS_DONE_SENDER)
	}
	cFileName := C.CString(fileName)
	cClientName := C.CString(clientName)
	paramArray := []string{fileName, clientName}
	cParamArray := NewCharArray(paramArray)
	defer func() {
		defer C.free(unsafe.Pointer(cFileName))
		defer C.free(unsafe.Pointer(cClientName))
		defer cParamArray.Free()
	}()
	cParamCnt := C.int(len(paramArray))

	log.Printf("[%s] timestamp:[%d] code:[%d] notiParam:[%+v] cParamCnt:[%d]", rtkMisc.GetFuncInfo(), cTimestamp, cCode, cParamArray.Array, cParamCnt)
	C.invokeCallbackNotiMessage(cTimestamp, cCode, cParamArray.Array, cParamCnt)
}

func GoTriggerCallbackMethodStartBrowseMdns(instance, serviceType string) {
	cinstance := C.CString(instance)
	cserviceType := C.CString(serviceType)
	defer C.free(unsafe.Pointer(cinstance))
	defer C.free(unsafe.Pointer(cserviceType))
	C.invokeCallbackMethodStartBrowseMdns(cinstance, cserviceType)
}

func GoTriggerCallbackMethodStopBrowseMdns() {
	C.invokeCallbackMethodStopBrowseMdns()
}

func GoTriggerCallbackGetAuthData(clientIndex uint32) string {
	cClientIndex := C.uint(clientIndex)
	cAuthData := C.invokeCallbackGetAuthData(cClientIndex)
	defer C.free(unsafe.Pointer(cAuthData))

	authData := C.GoString(cAuthData)
	log.Printf("[%s] %s", rtkMisc.GetFuncInfo(), authData)
	return authData
}

func GoTriggerCallbackSetDIASStatus(status uint32) {
	cStatus := C.uint(status)
	C.invokeCallbackSetDIASStatus(cStatus)
	log.Printf("[%s] status:%d", rtkMisc.GetFuncInfo(), status)
}

func GoTriggerCallbackSetMonitorName(name string) {
	cMonitorName := C.CString(name)
	defer C.free(unsafe.Pointer(cMonitorName))
	C.invokeCallbackSetMonitorName(cMonitorName)
	log.Printf("[%s] MonitorName:[%s]", rtkMisc.GetFuncInfo(), name)
}

func GoTriggerCallbackPasteXClipData(text, image, html, rtf string) {
	cText := C.CString(text)
	cImage := C.CString(image)
	cHtml := C.CString(html)
	cRtf := C.CString(rtf)
	defer C.free(unsafe.Pointer(cText))
	defer C.free(unsafe.Pointer(cImage))
	defer C.free(unsafe.Pointer(cHtml))
	defer C.free(unsafe.Pointer(cRtf))

	log.Printf("[%s] text:%d, image:%d, html:%d, rtf:%d \n\n", rtkMisc.GetFuncInfo(), len(text), len(image), len(html), len(rtf))
	C.invokeCallbackPasteXClipData(cText, cImage, cHtml, cRtf)
}

func GoTriggerCallbackReqClientUpdateVer(ver string) {
	cVer := C.CString(ver)
	defer C.free(unsafe.Pointer(cVer))

	log.Printf("[%s] version:%s", rtkMisc.GetFuncInfo(), ver)
	C.invokeCallbackRequestUpdateClientVersion(cVer)
}

func GoTriggerCallbackNotifyErrEvent(id string, errCode uint32, arg1, arg2, arg3, arg4 string) {
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
	C.invokeCallbackNotifyErrEvent(cId, cErrCode, cArg1, cArg2, cArg3, cArg4)
}

func GoTriggerCallbackNotifyBrowseResult(monitorName, instance, ipAddr, version string, timestamp int64) {
	cMonitorName := C.CString(monitorName)
	cInstance := C.CString(instance)
	cIpAddr := C.CString(ipAddr)
	cVersion := C.CString(version)
	cTimeStamp := C.ulonglong(timestamp)

	defer func() {
		C.free(unsafe.Pointer(cMonitorName))
		C.free(unsafe.Pointer(cInstance))
		C.free(unsafe.Pointer(cIpAddr))
		C.free(unsafe.Pointer(cVersion))
	}()

	log.Printf("[%s] name:%s, instance:%s, ip:%s, version:%s, timestamp:%d", rtkMisc.GetFuncInfo(), monitorName, instance, ipAddr, version, timestamp)
	C.invokeCallbackNotifyBrowseResult(cMonitorName, cInstance, cIpAddr, cVersion, cTimeStamp)
}

//export SetCallbackUpdateClientStatus
func SetCallbackUpdateClientStatus(cb C.CallbackUpdateClientStatus) {
	C.setCallbackUpdateClientStatus(cb)
}

//export SetCallbackFileListSendNotify
func SetCallbackFileListSendNotify(cb C.CallbackMethodFileListNotify) {
	C.setCallbackFileListSendNotify(cb)
}

//export SetCallbackFileListReceiveNotify
func SetCallbackFileListReceiveNotify(cb C.CallbackMethodFileListNotify) {
	C.setCallbackFileListReceiveNotify(cb)
}

//export SetCallbackUpdateSendProgressBar
func SetCallbackUpdateSendProgressBar(cb C.CallbackUpdateProgressBar) {
	C.setCallbackUpdateSendProgressBar(cb)
}

//export SetCallbackUpdateReceiveProgressBar
func SetCallbackUpdateReceiveProgressBar(cb C.CallbackUpdateProgressBar) {
	C.setCallbackUpdateReceiveProgressBar(cb)
}

//export SetCallbackNotiMessage
func SetCallbackNotiMessage(cb C.CallbackNotiMessage) {
	C.setCallbackNotiMessage(cb)
}

//export SetCallbackMethodStartBrowseMdns
func SetCallbackMethodStartBrowseMdns(cb C.CallbackMethodStartBrowseMdns) {
	C.setCallbackMethodStartBrowseMdns(cb)
}

//export SetCallbackMethodStopBrowseMdns
func SetCallbackMethodStopBrowseMdns(cb C.CallbackMethodStopBrowseMdns) {
	C.setCallbackMethodStopBrowseMdns(cb)
}

//export SetCallbackGetAuthData
func SetCallbackGetAuthData(cb C.CallbackAuthData) {
	log.Printf("[%s] SetCallbackGetAuthData", rtkMisc.GetFuncInfo())
	C.setCallbackGetAuthData(cb)
}

//export SetCallbackDIASStatus
func SetCallbackDIASStatus(cb C.CallbackSetDIASStatus) {
	log.Printf("[%s] SetCallbackDIASStatus", rtkMisc.GetFuncInfo())
	C.setCallbackSetDIASStatus(cb)
}

//export SetCallbackMonitorName
func SetCallbackMonitorName(cb C.CallbackSetMonitorName) {
	log.Printf("[%s] SetCallbackMonitorName", rtkMisc.GetFuncInfo())
	C.setCallbackSetMonitorName(cb)
}

//export SetCallbackPasteXClipData
func SetCallbackPasteXClipData(cb C.CallbackPasteXClipData) {
	log.Printf("[%s] SetCallbackPasteXClipData", rtkMisc.GetFuncInfo())
	C.setCallbackPasteXClipData(cb)
}

//export SetCallbackRequestUpdateClientVersion
func SetCallbackRequestUpdateClientVersion(cb C.CallbackRequestUpdateClientVersion) {
	log.Printf("[%s] SetCallbackRequestUpdateClientVersion", rtkMisc.GetFuncInfo())
	C.setCallbackRequestUpdateClientVersion(cb)
}

//export SetCallbackNotifyErrEvent
func SetCallbackNotifyErrEvent(cb C.CallbackNotifyErrEvent) {
	log.Printf("[%s] SetCallbackNotifyErrEvent", rtkMisc.GetFuncInfo())
	C.setCallbackNotifyErrEvent(cb)
}

//export SetCallbackNotifyBrowseResult
func SetCallbackNotifyBrowseResult(cb C.CallbackNotifyBrowseResult) {
	log.Printf("[%s] SetCallbackNotifyBrowseResult", rtkMisc.GetFuncInfo())
	C.setCallbackNotifyBrowseResult(cb)
}

//export MainInit
func MainInit(deviceName, rootPath, serverId, serverIpInfo, listenHost string, listenPort int) {
	rtkPlatform.SetDeviceName(deviceName)

	if rootPath == "" || !rtkMisc.FolderExists(rootPath) {
		log.Fatalf("[%s] RootPath :[%s] is invalid!", rtkMisc.GetFuncInfo(), rootPath)
	}
	rtkPlatform.SetupRootPath(rootPath)
	log.Printf("[%s] deviceName:[%s] rootPath:[%s] host:[%s] port:[%d]", rtkMisc.GetFuncInfo(), deviceName, rootPath, listenHost, listenPort)
	rtkCmd.MainInit(serverId, serverIpInfo, listenHost, listenPort)
}

//export SetMsgEventFunc
func SetMsgEventFunc(event int, arg1, arg2, arg3, arg4 string) {
	log.Printf("[%s] event:[%d], arg1:%s, arg2:%s, arg3:%s, arg4:%s\n", rtkMisc.GetFuncInfo(), event, arg1, arg2, arg3, arg4)
	rtkPlatform.GoSetMsgEventFunc(uint32(event), arg1, arg2, arg3, arg4)
}

//export SendXClipData
func SendXClipData(text, image, html, rtf string) {
	log.Printf("[%s] text:%d, image:%d, html:%d, rtf:%d \n\n", rtkMisc.GetFuncInfo(), len(text), len(image), len(html), len(rtf))
	rtkPlatform.GoCopyXClipData(text, image, html, rtf)
}

//export GetClientListEx
func GetClientListEx() *C.char {
	clientList := rtkUtils.GetClientListEx()
	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), clientList)
	return C.CString(clientList)
}

//export GetClientList
func GetClientList() *C.char {
	clientList := rtkUtils.GetClientList()
	log.Printf("GetClientList :[%s]", clientList)
	return C.CString(clientList)
}

//export SendAddrsFromPlatform
func SendAddrsFromPlatform(addrsList string) {
	parts := strings.Split(addrsList, "#")
	rtkUtils.GetAddrsFromPlatform(parts)
}

//export SendNetInterfaces
func SendNetInterfaces(name, mac string, mtu, index int, flag uint) {
	log.Printf("SendNetInterfaces [%s][%s][%d][%d][%d]", name, mac, mtu, index, flag)
	rtkUtils.SetNetInterfaces(name, index)
}

//export SendMultiFilesDropRequest
func SendMultiFilesDropRequest(multiFilesData string) int {
	return int(rtkPlatform.GoMultiFilesDropRequest(multiFilesData))
}

//export SetCancelFileTransfer
func SetCancelFileTransfer(ipPort, clientID string, timeStamp uint64) {
	log.Printf("[%s]  ID:[%s] IP:[%s]  timestamp[%d]", rtkMisc.GetFuncInfo(), clientID, ipPort, timeStamp)
	rtkPlatform.GoCancelFileTrans(ipPort, clientID, timeStamp)
}

//export SetNetWorkConnected
func SetNetWorkConnected(isConnect bool) {
	log.Printf("[%s] SetNetWorkConnected:[%v]", rtkMisc.GetFuncInfo(), isConnect)
	//rtkPlatform.SetNetWorkConnected(isConnect)
}

//export SetHostListenAddr
func SetHostListenAddr(listenHost string, listenPort int) {
	log.Printf("[%s] SetHostListAddr:[%s][%d]", rtkMisc.GetFuncInfo(), listenHost, listenPort)
	rtkPlatform.GoSetHostListenAddr(listenHost, listenPort)
}

//export SetDIASID
func SetDIASID(DiasID string) {
	log.Printf(" [%s]  DiasID:[%s]", rtkMisc.GetFuncInfo(), DiasID)
	rtkPlatform.GoGetMacAddressCallback(DiasID)
}

//export SetDetectPluginEvent
func SetDetectPluginEvent(isPlugin bool) {
	log.Printf(" [%s] isPlugin:[%+v]", rtkMisc.GetFuncInfo(), isPlugin)
	rtkPlatform.GoTriggerDetectPluginEvent(isPlugin)
}

//export GetVersion
func GetVersion() *C.char {
	return C.CString(rtkGlobal.ClientVersion)
}

//export GetBuildDate
func GetBuildDate() *C.char {
	return C.CString(rtkBuildConfig.BuildDate)
}

//export SetBrowseMdnsResult
func SetBrowseMdnsResult(instance, ip string, port int, productName, mName, timestamp, version string) {
	log.Printf("[%s], instacne:[%s], ip:[%s], port[%d], productName:[%s], mName:[%s], timestamp:[%s], verion:[%s]",
		rtkMisc.GetFuncInfo(), instance, ip, port, productName, mName, timestamp, version)
	rtkPlatform.GoBrowseMdnsResultCallback(instance, ip, port, productName, mName, timestamp, version)
}

//export SetConfirmDocumentsAccept
func SetConfirmDocumentsAccept(ifConfirm bool) {
	log.Printf("[%s], ifConfirm:[%+v]", rtkMisc.GetFuncInfo(), ifConfirm)
	rtkPlatform.SetConfirmDocumentsAccept(ifConfirm)
}

//export FreeCString
func FreeCString(p *C.char) {
	C.free(unsafe.Pointer(p))
}

//export BrowseLanServer
func BrowseLanServer() {
	log.Printf("[%s] triggered!", rtkMisc.GetFuncInfo())
	rtkPlatform.GoBrowseLanServer()
}

//export WorkerConnectLanServer
func WorkerConnectLanServer(instance string) {
	log.Printf("[%s]  instance:[%s]", rtkMisc.GetFuncInfo(), instance)
	rtkPlatform.GoConnectLanServer(instance)
}

//export SetupAppLink
func SetupAppLink(link string) {
	log.Printf("[%s] link:[%s]", rtkMisc.GetFuncInfo(), link)
	rtkPlatform.GoSetupAppLink(link)
}
