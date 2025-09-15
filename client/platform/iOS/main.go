//go:build ios
// +build ios

package main

/*
#include <stdlib.h>
#include <stdint.h>

typedef void (*CallbackMethodText)(char*);
typedef void (*CallbackMethodImage)(char* content);
typedef void (*EventCallback)(int event);
typedef void (*CallbackMethodFileConfirm)(char* id, char* platform, char* fileName, long long fileSize);
typedef void (*CallbackMethodFoundPeer)();
typedef void (*CallbackUpdateClientStatus)(char* clientJsonStr);
typedef void (*CallbackMethodFileListNotify)(char* ip, char* id, char* platform,unsigned int fileCnt, unsigned long long totalSize,unsigned long long timestamp, char* firstFileName, unsigned long long firstFileSize);
typedef void (*CallbackUpdateMultipleProgressBar)(char* ip,char* id, char* deviceName, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp);
typedef void (*CallbackFileError)(char* id, char* fileName, char* err);
typedef void (*CallbackMethodStartBrowseMdns)(char* instance, char* serviceType);
typedef void (*CallbackMethodStopBrowseMdns)();
typedef char* (*CallbackAuthData)();
typedef void (*CallbackSetDIASStatus)(unsigned int status);
typedef void (*CallbackSetMonitorName)(char* monitorName);
typedef void (*CallbackRequestUpdateClientVersion)(char* clientVer);
typedef void (*CallbackNotifyErrEvent)(char* id, unsigned int errCode, char* arg1, char* arg2, char* arg3, char* arg4);
typedef void (*CallbackNotifyBrowseResult)(char* monitorName, char* instance, char* ip, char* version, unsigned long long timestamp);

static CallbackMethodText gCallbackMethodText = 0;
static CallbackMethodImage gCallbackMethodImage = 0;
static EventCallback gEventCallback = 0;
static CallbackMethodFileConfirm gCallbackMethodFileConfirm = 0;
static CallbackMethodFoundPeer gCallbackMethodFoundPeer = 0;
static CallbackUpdateClientStatus gCallbackUpdateClientStatus = 0;
static CallbackMethodFileListNotify gCallbackMethodFileListNotify = 0;
static CallbackUpdateMultipleProgressBar gCallbackUpdateMultipleProgressBar = 0;
static CallbackFileError gCallbackFileError = 0;
static CallbackMethodStartBrowseMdns gCallbackMethodStartBrowseMdns = 0;
static CallbackMethodStopBrowseMdns gCallbackMethodStopBrowseMdns = 0;
static CallbackAuthData gCallbackAuthData = 0;
static CallbackSetDIASStatus gCallbackSetDIASStatus = 0;
static CallbackSetMonitorName gCallbackSetMonitorName = 0;
static CallbackRequestUpdateClientVersion gCallbackRequestUpdateClientVersion = 0;
static CallbackNotifyErrEvent gCallbackNotifyErrEvent = 0;
static CallbackNotifyBrowseResult gCallbackNotifyBrowseResult = 0;

static void setCallbackMethodText(CallbackMethodText cb) {gCallbackMethodText = cb;}
static void invokeCallbackMethodText(char* str) {
	if (gCallbackMethodText) {gCallbackMethodText(str);}
}
static void setCallbackMethodImage(CallbackMethodImage cb) {gCallbackMethodImage = cb;}
static void invokeCallbackMethodImage(char* str) {
	if (gCallbackMethodImage) {gCallbackMethodImage(str);}
}
static void setEventCallback(EventCallback cb) {gEventCallback = cb;}
static void invokeEventCallback(int event) {
	if (gEventCallback) {gEventCallback(event);}
}
static void setCallbackMethodFileConfirm(CallbackMethodFileConfirm cb) {gCallbackMethodFileConfirm = cb;}
static void invokeCallbackMethodFileConfirm(char* id, char* platform, char* fileName, long long fileSize) {
	if (gCallbackMethodFileConfirm) {gCallbackMethodFileConfirm(id, platform, fileName, fileSize);}
}
static void setCallbackUpdateClientStatus(CallbackUpdateClientStatus cb) {gCallbackUpdateClientStatus = cb;}
static void invokeCallbackUpdateClientStatus(char* clientJsonStr) {
	if (gCallbackUpdateClientStatus) {gCallbackUpdateClientStatus(clientJsonStr);}
}
static void setCallbackMethodFoundPeer(CallbackMethodFoundPeer cb) {gCallbackMethodFoundPeer = cb;}
static void invokeCallbackMethodFoundPeer() {
	if (gCallbackMethodFoundPeer) {gCallbackMethodFoundPeer();}
}
static void setCallbackMethodFileListNotify(CallbackMethodFileListNotify cb) {gCallbackMethodFileListNotify = cb;}
static void invokeCallbackMethodFileListNotify(char* ip, char* id, char* platform,unsigned int fileCnt, unsigned long long totalSize,unsigned long long timestamp, char* firstFileName, unsigned long long firstFileSize) {
	if (gCallbackMethodFileListNotify) {gCallbackMethodFileListNotify(ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize);}
}
static void setCallbackUpdateMultipleProgressBar(CallbackUpdateMultipleProgressBar cb) {gCallbackUpdateMultipleProgressBar = cb;}
static void invokeCallbackUpdateMultipleProgressBar(char* ip,char* id, char* deviceName, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp) {
	if (gCallbackUpdateMultipleProgressBar) {gCallbackUpdateMultipleProgressBar(ip,id, deviceName,currentfileName,recvFileCnt,totalFileCnt,currentFileSize,totalSize, recvSize, timestamp);}
}
static void setCallbackFileError(CallbackFileError cb) {gCallbackFileError = cb;}
static void invokeCallbackFileError(char* id, char* fileName, char* err) {
	if (gCallbackFileError) {gCallbackFileError(id, fileName, err);}
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
static char* invokeCallbackGetAuthData() {
	if (gCallbackAuthData) { return gCallbackAuthData();}
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
	"encoding/json"
	"log"
	"path/filepath"
	rtkBuildConfig "rtk-cross-share/client/buildConfig"
	rtkCmd "rtk-cross-share/client/cmd"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strings"
	"time"
	"unsafe"
)

func main() {
}

func init() {
	rtkPlatform.SetCallbackMethodText(GoTriggerCallbackMethodText)
	rtkPlatform.SetCallbackMethodImage(GoTriggerCallbackMethodImage)
	rtkPlatform.SetEventCallback(GoTriggerEventCallback)
	rtkPlatform.SetCallbackMethodFileConfirm(GoTriggerCallbackMethodFileConfirm)
	rtkPlatform.SetCallbackFileListNotify(GoTriggerCallbackMethodFileListNotify)
	rtkPlatform.SetCallbackUpdateClientStatus(GoTriggerCallbackUpdateClientStatus)
	rtkPlatform.SetCallbackMethodFoundPeer(GoTriggerCallbackMethodFoundPeer)
	rtkPlatform.SetCallbackUpdateMultipleProgressBar(GoTriggerCallbackUpdateMultipleProgressBar)
	rtkPlatform.SetCallbackFileError(GoTriggerCallbackFileError)
	rtkPlatform.SetCallbackMethodStartBrowseMdns(GoTriggerCallbackMethodStartBrowseMdns)
	rtkPlatform.SetCallbackMethodStopBrowseMdns(GoTriggerCallbackMethodStopBrowseMdns)
	rtkPlatform.SetCallbackGetAuthData(GoTriggerCallbackGetAuthData)
	rtkPlatform.SetCallbackDIASStatus(GoTriggerCallbackSetDIASStatus)
	rtkPlatform.SetCallbackMonitorName(GoTriggerCallbackSetMonitorName)
	rtkPlatform.SetCallbackRequestUpdateClientVersion(GoTriggerCallbackReqClientUpdateVer)
	rtkPlatform.SetCallbackNotifyErrEvent(GoTriggerCallbackNotifyErrEvent)
	rtkPlatform.SetCallbackNotifyBrowseResult(GoTriggerCallbackNotifyBrowseResult)

	rtkPlatform.SetConfirmDocumentsAccept(false)
}

type MultiFilesDropRequestInfo struct {
	Id       string
	Ip       string
	PathList []string
}

func GoTriggerCallbackMethodText(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.invokeCallbackMethodText(cstr)
}

func GoTriggerCallbackMethodImage(str string) {
	log.Printf("[%s] GoTriggerCallbackMethodImage", rtkMisc.GetFuncInfo())
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.invokeCallbackMethodImage(cstr)
}

func GoTriggerEventCallback(event int) {
	cevent := C.int(event)
	C.invokeEventCallback(cevent)
}

func GoTriggerCallbackMethodFileConfirm(id, platform, fileName string, fileSize int64) {
	cid := C.CString(id)
	cplatform := C.CString(platform)
	cfileName := C.CString(fileName)
	cfileSize := C.longlong(fileSize)
	defer C.free(unsafe.Pointer(cid))
	defer C.free(unsafe.Pointer(cplatform))
	defer C.free(unsafe.Pointer(cfileName))
	C.invokeCallbackMethodFileConfirm(cid, cplatform, cfileName, cfileSize)
}

func GoTriggerCallbackUpdateClientStatus(clientInfo string) {
	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), clientInfo)
	cClientInfo := C.CString(clientInfo)
	defer C.free(unsafe.Pointer(cClientInfo))

	C.invokeCallbackUpdateClientStatus(cClientInfo)
}

func GoTriggerCallbackMethodFoundPeer() {
	log.Printf("[%s] FoundPeer Trigger!", rtkMisc.GetFuncInfo())
	C.invokeCallbackMethodFoundPeer()
}

func GoTriggerCallbackMethodFileListNotify(ip, id, platform string, fileCnt uint32, totalSize uint64, timestamp uint64, firstFileName string, firstFileSize uint64) {
	cip := C.CString(ip)
	cid := C.CString(id)
	cplatform := C.CString(platform)
	cfirstFileName := C.CString(firstFileName)

	defer func() {
		C.free(unsafe.Pointer(cip))
		C.free(unsafe.Pointer(cid))
		C.free(unsafe.Pointer(cplatform))
		C.free(unsafe.Pointer(cfirstFileName))
	}()

	cFileCnt := C.uint(fileCnt)
	ctotalSize := C.ulonglong(totalSize)
	ctimeStamp := C.ulonglong(timestamp)
	cfirstFileSize := C.ulonglong(firstFileSize)
	C.invokeCallbackMethodFileListNotify(cip, cid, cplatform, cFileCnt, ctotalSize, ctimeStamp, cfirstFileName, cfirstFileSize)
}

func GoTriggerCallbackUpdateMultipleProgressBar(ip, id, deviceName, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
	cip := C.CString(ip)
	cid := C.CString(id)
	cdeviceName := C.CString(deviceName)
	ccurrentFileName := C.CString(currentFileName)

	defer func() {
		C.free(unsafe.Pointer(cip))
		C.free(unsafe.Pointer(cid))
		C.free(unsafe.Pointer(cdeviceName))
		C.free(unsafe.Pointer(ccurrentFileName))
	}()

	crecvFileCnt := C.uint(sentFileCnt)
	ctotalFileCnt := C.uint(totalFileCnt)

	ccurrentFileSize := C.ulonglong(currentFileSize)
	ctotalSize := C.ulonglong(totalSize)
	crecvSize := C.ulonglong(sentSize)
	ctimeStamp := C.ulonglong(timestamp)

	C.invokeCallbackUpdateMultipleProgressBar(cip, cid, cdeviceName, ccurrentFileName, crecvFileCnt, ctotalFileCnt, ccurrentFileSize, ctotalSize, crecvSize, ctimeStamp)
}

func GoTriggerCallbackFileError(id, filename, err string) {
	cid := C.CString(id)
	cfilename := C.CString(filename)
	cerr := C.CString(err)
	defer C.free(unsafe.Pointer(cid))
	defer C.free(unsafe.Pointer(cfilename))
	defer C.free(unsafe.Pointer(cerr))
	C.invokeCallbackFileError(cid, cfilename, cerr)
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

func GoTriggerCallbackGetAuthData() string {
	cAuthData := C.invokeCallbackGetAuthData()
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

//export SetCallbackMethodText
func SetCallbackMethodText(cb C.CallbackMethodText) {
	log.Printf("[%s] SetCallbackMethodText", rtkMisc.GetFuncInfo())
	C.setCallbackMethodText(cb)
}

//export SetCallbackMethodImage
func SetCallbackMethodImage(cb C.CallbackMethodImage) {
	log.Printf("[%s] SetCallbackMethodImage", rtkMisc.GetFuncInfo())
	C.setCallbackMethodImage(cb)
}

//export SetEventCallback
func SetEventCallback(cb C.EventCallback) {
	C.setEventCallback(cb)
}

//export SetCallbackMethodFileConfirm
func SetCallbackMethodFileConfirm(cb C.CallbackMethodFileConfirm) {
	C.setCallbackMethodFileConfirm(cb)
}

//export SetCallbackUpdateClientStatus
func SetCallbackUpdateClientStatus(cb C.CallbackUpdateClientStatus) {
	C.setCallbackUpdateClientStatus(cb)
}

//export SetCallbackMethodFoundPeer
func SetCallbackMethodFoundPeer(cb C.CallbackMethodFoundPeer) {
	C.setCallbackMethodFoundPeer(cb)
}

//export SetCallbackMethodFileListNotify
func SetCallbackMethodFileListNotify(cb C.CallbackMethodFileListNotify) {
	C.setCallbackMethodFileListNotify(cb)
}

//export SetCallbackUpdateMultipleProgressBar
func SetCallbackUpdateMultipleProgressBar(cb C.CallbackUpdateMultipleProgressBar) {
	C.setCallbackUpdateMultipleProgressBar(cb)
}

//export SetCallbackFileError
func SetCallbackFileError(cb C.CallbackFileError) {
	C.setCallbackFileError(cb)
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
func MainInit(deviceName, serverId, serverIpInfo, listenHost string, listenPort int) {
	rtkPlatform.SetDeviceName(deviceName)
	rootPath := rtkPlatform.GetRootPath()
	if rootPath == "" || !rtkMisc.FolderExists(rootPath) {
		log.Fatalf("[%s] RootPath :[%s] is invalid!", rtkMisc.GetFuncInfo(), rootPath)
	}

	log.Printf("[%s] device name:[%s] host:[%s] port:[%d]", rtkMisc.GetFuncInfo(), deviceName, listenHost, listenPort)
	rtkCmd.MainInit(serverId, serverIpInfo, listenHost, listenPort)
}

//export SetMsgEventFunc
func SetMsgEventFunc(event int, arg1, arg2, arg3, arg4 string) {
	log.Printf("[%s] event:[%d], arg1:%s, arg2:%s, arg3:%s, arg4:%s\n", rtkMisc.GetFuncInfo(), event, arg1, arg2, arg3, arg4)
	rtkPlatform.GoSetMsgEventFunc(uint32(event), arg1, arg2, arg3, arg4)
}

//export SendText
func SendText(s string) {
	rtkPlatform.SendMessage(s)
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

//export SendImage
func SendImage(content string) {
	if content == "" || len(content) == 0 {
		return
	}
	data := rtkUtils.Base64Decode(content)
	if data == nil {
		log.Printf("[%s] Image base64 decode error", rtkMisc.GetFuncInfo())
		return
	}

	format, width, height := rtkUtils.GetByteImageInfo(data)
	jpegData, err := rtkUtils.ImageToJpeg(format, data)
	if err != nil {
		return
	}
	if len(jpegData) == 0 {
		log.Printf("[CopyImage] Error: jpeg data is empty")
		return
	}
	log.Printf("SendImage:[%d][%d]", len(content), len(jpegData))

	imgHeader := rtkCommon.ImgHeader{
		Width:       int32(width),
		Height:      int32(height),
		Planes:      1,
		BitCount:    32,
		Compression: 0,
	}

	rtkPlatform.GoCopyImage(imgHeader, jpegData)
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
	var multiFileInfo MultiFilesDropRequestInfo
	err := json.Unmarshal([]byte(multiFilesData), &multiFileInfo)
	if err != nil {
		log.Printf("[%s] Unmarshal[%s] err:%+v", rtkMisc.GetFuncInfo(), multiFilesData, err)
		return int(rtkCommon.SendFilesRequestParameterErr)
	}
	log.Printf("id:[%s] ip:[%s] len:[%d] json:[%s]", multiFileInfo.Id, multiFileInfo.Ip, len(multiFileInfo.PathList), multiFilesData)

	fileList := make([]rtkCommon.FileInfo, 0)
	folderList := make([]string, 0)
	totalSize := uint64(0)

	for _, file := range multiFileInfo.PathList {
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
			log.Printf("[%s] get file or path:[%s] is not exist , so skit it!", rtkMisc.GetFuncInfo(), file)
		}
	}
	totalDesc := rtkMisc.FileSizeDesc(totalSize)
	timestamp := uint64(time.Now().UnixMilli())
	log.Printf("[%s] ID[%s] IP:[%s] get file count:[%d] folder count:[%d], totalSize:[%d] totalDesc:[%s] timestamp:[%d]", rtkMisc.GetFuncInfo(), multiFileInfo.Id, multiFileInfo.Ip, len(fileList), len(folderList), totalSize, totalDesc, timestamp)
	return int(rtkPlatform.GoMultiFilesDropRequest(multiFileInfo.Id, &fileList, &folderList, totalSize, timestamp, totalDesc))
}

//export SetCancelFileTransfer
func SetCancelFileTransfer(ipPort, clientID string, timeStamp uint64) {
	log.Printf("[%s]  ID:[%s] IP:[%s]  timestamp[%d]", rtkMisc.GetFuncInfo(), clientID, ipPort, timeStamp)
	rtkPlatform.GoCancelFileTrans(ipPort, clientID, timeStamp)
}

//export SetNetWorkConnected
func SetNetWorkConnected(isConnect bool) {
	log.Printf("[%s] SetNetWorkConnected:[%v]", rtkMisc.GetFuncInfo(), isConnect)
	rtkPlatform.SetNetWorkConnected(isConnect)
}

//export SetHostListenAddr
func SetHostListenAddr(listenHost string, listenPort int) {
	log.Printf("[%s] SetHostListAddr:[%s][%d]", rtkMisc.GetFuncInfo(), listenHost, listenPort)
	if listenHost == "" || listenHost == rtkMisc.DefaultIp || listenHost == rtkMisc.LoopBackIp || listenPort <= rtkGlobal.DefaultPort {
		return
	}
	if rtkGlobal.ListenHost != rtkMisc.DefaultIp &&
		rtkGlobal.ListenHost != "" &&
		rtkGlobal.ListenPort != rtkGlobal.DefaultPort &&
		(listenHost != rtkGlobal.ListenHost || listenPort != rtkGlobal.ListenPort) {

		log.Printf("[%s] The previous host Addr:[%s:%d], new host Addr:[%s:%d] ", rtkMisc.GetFuncInfo(), rtkGlobal.ListenHost, rtkGlobal.ListenPort, listenHost, listenPort)
		log.Println("**************** Attention please, the host listen addr is switch! ********************\n\n")
		rtkGlobal.ListenHost = listenHost
		rtkGlobal.ListenPort = listenPort
		rtkPlatform.GoTriggerNetworkSwitch()
	}
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

//export SetupRootPath
func SetupRootPath(rootPath string) {
	log.Printf("[%s] rootPath:[%s]", rtkMisc.GetFuncInfo(), rootPath)
	rtkPlatform.SetupRootPath(rootPath)
}

//export SetSrcAndPort
func SetSrcAndPort(source, port int) {
	log.Printf("[%s] , source:[%d], port:[%d]", rtkMisc.GetFuncInfo(), source, port)
	rtkPlatform.GoSetSrcAndPort(source, port)
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
