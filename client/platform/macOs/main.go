//go:build darwin && !ios

package main

/*
#include <stdlib.h>
#include <stdint.h>

typedef void (*CallbackUpdateSystemInfo)(char* ipInfo, char* verInfo);
typedef void (*CallbackUpdateClientStatus)(char* clientJsonStr);
typedef void (*CallbackMethodFileListNotify)(char* ip, char* id, char* platform,unsigned int fileCnt, unsigned long long totalSize,unsigned long long timestamp, char* firstFileName, unsigned long long firstFileSize);
typedef void (*CallbackUpdateMultipleProgressBar)(char* ip,char* id, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp);
typedef void (*CallbackMethodStartBrowseMdns)(char* instance, char* serviceType);
typedef void (*CallbackMethodStopBrowseMdns)();
typedef void (*CallbackRequestSourceAndPort)();
typedef void (*CallbackAuthViaIndex)(unsigned int index);
typedef void (*CallbackSetDIASStatus)(unsigned int status);
typedef void (*CallbackSetMonitorName)(char* monitorName);
typedef void (*CallbackPasteXClipData)(char *text, char *image, char *html);
typedef void (*CallbackRequestUpdateClientVersion)(char* clientVer);
typedef void (*CallbackNotifyErrEvent)(char* id, unsigned int errCode, char* arg1, char* arg2, char* arg3, char* arg4);
typedef void (*CallbackNotifyBrowseResult)(char* monitorName, char* instance, char* ip, char* version, unsigned long long timestamp);

static CallbackUpdateSystemInfo gCallbackUpdateSystemInfo = 0;
static CallbackUpdateClientStatus gCallbackUpdateClientStatus = 0;
static CallbackMethodFileListNotify gCallbackMethodFileListNotify = 0;
static CallbackUpdateMultipleProgressBar gCallbackUpdateMultipleProgressBar = 0;
static CallbackMethodStartBrowseMdns gCallbackMethodStartBrowseMdns = 0;
static CallbackMethodStopBrowseMdns gCallbackMethodStopBrowseMdns = 0;
static CallbackRequestSourceAndPort gCallbackRequestSourceAndPort = 0;
static CallbackAuthViaIndex  gCallbackAuthViaIndex = 0;
static CallbackSetDIASStatus gCallbackSetDIASStatus = 0;
static CallbackSetMonitorName gCallbackSetMonitorName = 0;
static CallbackPasteXClipData gCallbackPasteXClipData = 0;
static CallbackRequestUpdateClientVersion gCallbackRequestUpdateClientVersion = 0;
static CallbackNotifyErrEvent gCallbackNotifyErrEvent = 0;
static CallbackNotifyBrowseResult gCallbackNotifyBrowseResult = 0;

static void setCallbackUpdateSystemInfo(CallbackUpdateSystemInfo cb) {gCallbackUpdateSystemInfo = cb;}
static void invokeCallbackUpdateSystemInfo(char* ipInfo, char* verInfo) {
	if (gCallbackUpdateSystemInfo) {gCallbackUpdateSystemInfo(ipInfo, verInfo);}
}
static void setCallbackUpdateClientStatus(CallbackUpdateClientStatus cb) {gCallbackUpdateClientStatus = cb;}
static void invokeCallbackUpdateClientStatus(char* clientJsonStr) {
	if (gCallbackUpdateClientStatus) {gCallbackUpdateClientStatus(clientJsonStr);}
}
static void setCallbackMethodFileListNotify(CallbackMethodFileListNotify cb) {gCallbackMethodFileListNotify = cb;}
static void invokeCallbackMethodFileListNotify(char* ip, char* id, char* platform,unsigned int fileCnt, unsigned long long totalSize,unsigned long long timestamp, char* firstFileName, unsigned long long firstFileSize) {
	if (gCallbackMethodFileListNotify) {gCallbackMethodFileListNotify(ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize);}
}
static void setCallbackUpdateMultipleProgressBar(CallbackUpdateMultipleProgressBar cb) {gCallbackUpdateMultipleProgressBar = cb;}
static void invokeCallbackUpdateMultipleProgressBar(char* ip,char* id, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp) {
	if (gCallbackUpdateMultipleProgressBar) {gCallbackUpdateMultipleProgressBar(ip,id, currentfileName,recvFileCnt,totalFileCnt,currentFileSize,totalSize, recvSize, timestamp);}
}
static void setCallbackMethodStartBrowseMdns(CallbackMethodStartBrowseMdns cb) {gCallbackMethodStartBrowseMdns = cb;}
static void invokeCallbackMethodStartBrowseMdns(char* instance, char* serviceType) {
	if (gCallbackMethodStartBrowseMdns) {gCallbackMethodStartBrowseMdns(instance, serviceType);}
}
static void setCallbackMethodStopBrowseMdns(CallbackMethodStopBrowseMdns cb) {gCallbackMethodStopBrowseMdns = cb;}
static void invokeCallbackMethodStopBrowseMdns() {
	if (gCallbackMethodStopBrowseMdns) {gCallbackMethodStopBrowseMdns();}
}
static void setCallbackRequestSourceAndPort(CallbackRequestSourceAndPort cb) {gCallbackRequestSourceAndPort = cb;}
static void invokeCallbackRequestSourceAndPort() {
	if (gCallbackRequestSourceAndPort) { gCallbackRequestSourceAndPort();}
}
static void setCallbackAuthViaIndex(CallbackAuthViaIndex cb) {gCallbackAuthViaIndex = cb;}
static void invokeCallbackAuthViaIndex(unsigned int index) {
	if (gCallbackAuthViaIndex) { gCallbackAuthViaIndex(index);}
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
static void invokeCallbackPasteXClipData(char *text, char *image, char *html) {
	if (gCallbackPasteXClipData) { gCallbackPasteXClipData(text, image, html);}
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
	"fmt"
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
	rtkPlatform.SetCallbackUpdateSystemInfo(GoTriggerCallbackUpdateSystemInfo)
	rtkPlatform.SetCallbackFileListNotify(GoTriggerCallbackMethodFileListNotify)
	rtkPlatform.SetCallbackUpdateClientStatus(GoTriggerCallbackUpdateClientStatus)
	rtkPlatform.SetCallbackUpdateMultipleProgressBar(GoTriggerCallbackUpdateMultipleProgressBar)
	rtkPlatform.SetCallbackMethodStartBrowseMdns(GoTriggerCallbackMethodStartBrowseMdns)
	rtkPlatform.SetCallbackMethodStopBrowseMdns(GoTriggerCallbackMethodStopBrowseMdns)
	rtkPlatform.SetCallbackDIASStatus(GoTriggerCallbackSetDIASStatus)
	rtkPlatform.SetCallbackAuthViaIndex(GoTriggerCallbackAuthViaIndex)
	rtkPlatform.SetCallbackRequestSourceAndPort(GoTriggerCallbackRequestSourceAndPort)
	rtkPlatform.SetCallbackMonitorName(GoTriggerCallbackSetMonitorName)
	rtkPlatform.SetCallbackPasteXClipData(GoTriggerCallbackPasteXClipData)
	rtkPlatform.SetCallbackRequestUpdateClientVersion(GoTriggerCallbackReqClientUpdateVer)
	rtkPlatform.SetCallbackNotifyErrEvent(GoTriggerCallbackNotifyErrEvent)

	rtkPlatform.SetConfirmDocumentsAccept(false)
}

type MultiFilesDropRequestInfo struct {
	Id       string
	Ip       string
	PathList []string
}

type MultiFilesDragRequestInfo struct {
	PathList []string
}

func GoTriggerCallbackUpdateSystemInfo(ipAddr, versionInfo string) {
	cIpAddr := C.CString(ipAddr)
	defer C.free(unsafe.Pointer(cIpAddr))

	cServiceVer := C.CString(versionInfo)
	defer C.free(unsafe.Pointer(cServiceVer))

	C.invokeCallbackUpdateSystemInfo(cIpAddr, cServiceVer)
}

func GoTriggerCallbackUpdateClientStatus(clientInfo string) {
	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), clientInfo)
	cClientInfo := C.CString(clientInfo)
	defer C.free(unsafe.Pointer(cClientInfo))

	C.invokeCallbackUpdateClientStatus(cClientInfo)
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

func GoTriggerCallbackUpdateMultipleProgressBar(ip, id, currentFileName string, sentFileCnt, totalFileCnt uint32, currentFileSize, totalSize, sentSize, timestamp uint64) {
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

	C.invokeCallbackUpdateMultipleProgressBar(cip, cid, ccurrentFileName, crecvFileCnt, ctotalFileCnt, ccurrentFileSize, ctotalSize, crecvSize, ctimeStamp)
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

func GoTriggerCallbackRequestSourceAndPort() {
	log.Printf("[%s] request source and port", rtkMisc.GetFuncInfo())
	C.invokeCallbackRequestSourceAndPort()
}

func GoTriggerCallbackAuthViaIndex(index uint32) {
	cIndex := C.uint(index)
	C.invokeCallbackAuthViaIndex(cIndex)
	log.Printf("[%s] index:%d", rtkMisc.GetFuncInfo(), index)
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

func GoTriggerCallbackPasteXClipData(text, image, html string) {
	cText := C.CString(text)
	cImage := C.CString(image)
	cHtml := C.CString(html)
	defer C.free(unsafe.Pointer(cText))
	defer C.free(unsafe.Pointer(cImage))
	defer C.free(unsafe.Pointer(cHtml))

	log.Printf("[%s] text len:%d , image len:%d, html:%d\n\n", rtkMisc.GetFuncInfo(), len(text), len(image), len(html))
	C.invokeCallbackPasteXClipData(cText, cImage, cHtml)
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

//export SetCallbackUpdateSystemInfo
func SetCallbackUpdateSystemInfo(cb C.CallbackUpdateSystemInfo) {
	C.setCallbackUpdateSystemInfo(cb)
}

//export SetCallbackUpdateClientStatus
func SetCallbackUpdateClientStatus(cb C.CallbackUpdateClientStatus) {
	C.setCallbackUpdateClientStatus(cb)
}

//export SetCallbackMethodFileListNotify
func SetCallbackMethodFileListNotify(cb C.CallbackMethodFileListNotify) {
	C.setCallbackMethodFileListNotify(cb)
}

//export SetCallbackUpdateMultipleProgressBar
func SetCallbackUpdateMultipleProgressBar(cb C.CallbackUpdateMultipleProgressBar) {
	C.setCallbackUpdateMultipleProgressBar(cb)
}

//export SetCallbackMethodStartBrowseMdns
func SetCallbackMethodStartBrowseMdns(cb C.CallbackMethodStartBrowseMdns) {
	C.setCallbackMethodStartBrowseMdns(cb)
}

//export SetCallbackMethodStopBrowseMdns
func SetCallbackMethodStopBrowseMdns(cb C.CallbackMethodStopBrowseMdns) {
	C.setCallbackMethodStopBrowseMdns(cb)
}

//export SetCallbackRequestSourceAndPort
func SetCallbackRequestSourceAndPort(cb C.CallbackRequestSourceAndPort) {
	log.Printf("[%s] SetCallbackRequestSourceAndPort", rtkMisc.GetFuncInfo())
	C.setCallbackRequestSourceAndPort(cb)
}

//export SetCallbackAuthViaIndex
func SetCallbackAuthViaIndex(cb C.CallbackAuthViaIndex) {
	log.Printf("[%s] SetCallbackAuthViaIndex", rtkMisc.GetFuncInfo())
	C.setCallbackAuthViaIndex(cb)
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
func MainInit(deviceName, rootPath, downloadPath, serverId, serverIpInfo, listenHost string, listenPort int) {
	rtkPlatform.SetDeviceName(deviceName)

	if rootPath == "" || !rtkMisc.FolderExists(rootPath) {
		log.Fatalf("[%s] RootPath :[%s] is invalid!", rtkMisc.GetFuncInfo(), rootPath)
	}
	rtkPlatform.SetupRootPath(rootPath)
	log.Printf("[%s] deviceName:[%s] rootPath:[%s] downloadPath:[%s] host:[%s] port:[%d]", rtkMisc.GetFuncInfo(), deviceName, rootPath, downloadPath, listenHost, listenPort)

	if downloadPath == "" || !rtkMisc.FolderExists(downloadPath) {
		log.Printf("[%s] downloadPath :[%s] is invalid, so use default path[%s]!", rtkMisc.GetFuncInfo(), downloadPath, rtkPlatform.GetDownloadPath())
	} else {
		rtkPlatform.GoUpdateDownloadPath(downloadPath)
	}

	rtkCmd.MainInit(serverId, serverIpInfo, listenHost, listenPort)
}

//export SetMsgEventFunc
func SetMsgEventFunc(event int, arg1, arg2, arg3, arg4 string) {
	log.Printf("[%s] event:[%d], arg1:%s, arg2:%s, arg3:%s, arg4:%s\n", rtkMisc.GetFuncInfo(), event, arg1, arg2, arg3, arg4)
	rtkPlatform.GoSetMsgEventFunc(uint32(event), arg1, arg2, arg3, arg4)
}

//export SendXClipData
func SendXClipData(text, image, html string) {
	log.Printf("[%s] text:%d, image:%d, html:%d\n\n", rtkMisc.GetFuncInfo(), len(text), len(image), len(html))

	imgData := []byte(nil)
	if image != "" {
		startTime := time.Now().UnixMilli()
		data := rtkUtils.Base64Decode(image)
		if data == nil {
			return
		}

		format, width, height := rtkUtils.GetByteImageInfo(data)
		jpegData, err := rtkUtils.ImageToJpeg(format, data)
		if err != nil {
			return
		}
		if len(jpegData) == 0 {
			log.Printf("[CopyXClip] Error: jpeg data is empty")
			return
		}

		imgData = jpegData
		log.Printf("image get jpg size:[%d](%d,%d),decode use:[%d]ms", len(imgData), width, height, time.Now().UnixMilli()-startTime)
	}

	rtkPlatform.GoCopyXClipData([]byte(text), imgData, []byte(html))
}

//export GetClientList
func GetClientList() *C.char {
	clientList := rtkUtils.GetClientListEx()
	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), clientList)
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
	nFileCnt := 0
	nFolderCnt := 0
	nPathSize := uint64(0)

	for _, file := range multiFileInfo.PathList {
		file = strings.ReplaceAll(file, "\\", "/")
		if rtkMisc.FolderExists(file) {
			nFileCnt = len(fileList)
			nFolderCnt = len(folderList)
			nPathSize = totalSize
			rtkUtils.WalkPath(file, &folderList, &fileList, &totalSize)
			log.Printf("[%s] walk a path:[%s], get [%d] files and [%d] folders, path total size:[%d]", rtkMisc.GetFuncInfo(), file, len(fileList)-nFileCnt, len(folderList)-nFolderCnt, totalSize-nPathSize)
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

//export RequestUpdateDownloadPath
func RequestUpdateDownloadPath(downloadPath string) {
	if downloadPath == "" || !rtkMisc.FolderExists(downloadPath) {
		log.Printf("[%s] get downloadPath [%s] is invalid!", rtkMisc.GetFuncInfo(), downloadPath)
		return
	}
	rtkPlatform.GoUpdateDownloadPath(downloadPath)
}

//export SetNetWorkConnected
func SetNetWorkConnected(isConnect bool) {
	log.Printf("[%s] SetNetWorkConnected:[%v]", rtkMisc.GetFuncInfo(), isConnect)
	//rtkPlatform.SetNetWorkConnected(isConnect)
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

//export SetMacAddress
func SetMacAddress(cMacAddress *C.char, length C.int) {
	log.Printf("SetMacAddress(%q, %d)\n", C.GoString(cMacAddress), length)
	if length != 6 {
		log.Printf("[%s] SetMacAddress failed, invalid MAC length:%d", rtkMisc.GetFuncInfo(), length)
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
func SetAuthStatusCode(authResult uint32) {
	log.Printf("[%s] authResult[%d]\n", rtkMisc.GetFuncInfo(), authResult)
	authStatus := uint8(authResult)
	rtkPlatform.GoSetAuthStatusCode(authStatus)
}

//export SetDIASSourceAndPort
func SetDIASSourceAndPort(cSource, cPort uint32) {
	source := uint8(cSource)
	port := uint8(cPort)
	log.Printf("[%s] diasSourceAndPortCallback (src,port): (%d,%d)", rtkMisc.GetFuncInfo(), source, port)
	rtkPlatform.GoSetDIASSourceAndPort(source, port)
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

//export SetDragFileListRequest
func SetDragFileListRequest(multiFilesData string, timeStamp uint64) {
	var multiFileInfo MultiFilesDragRequestInfo
	err := json.Unmarshal([]byte(multiFilesData), &multiFileInfo)
	if err != nil {
		log.Printf("[%s] Unmarshal[%s] err:%+v", rtkMisc.GetFuncInfo(), multiFilesData, err)
		return
	}
	log.Printf("len:[%d] json:[%s]", len(multiFileInfo.PathList), multiFilesData)

	fileList := make([]rtkCommon.FileInfo, 0)
	folderList := make([]string, 0)
	totalSize := uint64(0)
	nFileCnt := 0
	nFolderCnt := 0
	nPathSize := uint64(0)

	for _, file := range multiFileInfo.PathList {
		file = strings.ReplaceAll(file, "\\", "/")
		if rtkMisc.FolderExists(file) {
			nFileCnt = len(fileList)
			nFolderCnt = len(folderList)
			nPathSize = totalSize
			rtkUtils.WalkPath(file, &folderList, &fileList, &totalSize)
			log.Printf("[%s] walk a path:[%s], get [%d] files and [%d] folders, path total size:[%d]", rtkMisc.GetFuncInfo(), file, len(fileList)-nFileCnt, len(folderList)-nFolderCnt, totalSize-nPathSize)
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
	timestamp := uint64(timeStamp)
	totalDesc := rtkMisc.FileSizeDesc(totalSize)

	log.Printf("[%s] get file count:[%d] folder count:[%d], totalSize:[%d] totalDesc:[%s] timestamp:[%d]", rtkMisc.GetFuncInfo(), len(fileList), len(folderList), totalSize, totalDesc, timestamp)
	rtkPlatform.GoDragFileListRequest(&fileList, &folderList, totalSize, timestamp, totalDesc)
}

//export FreeCString
func FreeCString(p *C.char) {
	C.free(unsafe.Pointer(p))
}
