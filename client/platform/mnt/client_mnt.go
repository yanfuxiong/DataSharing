//go:build android

package libp2p_clipboard

import (
	"encoding/json"
	"log"
	"path/filepath"
	rtkBuildConfig "rtk-cross-share/client/buildConfig"
	rtkCmd "rtk-cross-share/client/cmd"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strings"
	"time"
)

type MultiFilesDropRequestInfo struct {
	TimeStamp int64
	Id        string
	Ip        string
	PathList  []string
}

type Callback interface {
	rtkPlatform.Callback
}

func SetupRootPath(rootPath string) {
	rtkPlatform.SetupRootPath(rootPath)
	log.Printf("[%s] rootPath:[%s]", rtkMisc.GetFuncInfo(), rootPath)
}

func WorkerConnectLanServer(instance string) {
	log.Printf("[%s]  instance:[%s]", rtkMisc.GetFuncInfo(), instance)
	rtkPlatform.GoConnectLanServer(instance)
}

func BrowseLanServer() {
	log.Printf("[%s]", rtkMisc.GetFuncInfo())
	rtkPlatform.GoBrowseLanServer()
}

func MainInit(cb Callback, deviceName, serverId, serverIpInfo, listentHost string, listentPort int) {
	rtkPlatform.SetCallback(cb)
	rtkPlatform.SetDeviceName(deviceName)

	rootPath := rtkPlatform.GetRootPath()
	if rootPath == "" || !rtkMisc.FolderExists(rootPath) {
		log.Fatalf("[%s] RootPath :[%s] is invalid!", rtkMisc.GetFuncInfo(), rootPath)
	}

	log.Printf("[%s] rootPath:[%s] device name:[%s] host:[%s] port:[%d]", rtkMisc.GetFuncInfo(), rootPath, deviceName, listentHost, listentPort)
	rtkCmd.MainInit(serverId, serverIpInfo, listentHost, listentPort)
}

func SetMsgEventFunc(event int, arg1, arg2, arg3, arg4 string) {
	log.Printf("[%s] event:[%d], arg1:%s, arg2:%s, arg3:%s, arg4:%s\n", rtkMisc.GetFuncInfo(), event, arg1, arg2, arg3, arg4)
	rtkPlatform.GoSetMsgEventFunc(uint32(event), arg1, arg2, arg3, arg4)
}

func SendXClipData(text, image, html, rtf string) {
	log.Printf("[%s] text:%d, image:%d, html:%d, rtf:%d \n\n", rtkMisc.GetFuncInfo(), len(text), len(image), len(html), len(rtf))
	rtkPlatform.GoCopyXClipData(text, image, html, rtf)
}

func GetClientList() string {
	clientList := rtkUtils.GetClientListEx()
	log.Printf("[%s] json Str:%s", rtkMisc.GetFuncInfo(), clientList)
	return clientList
}

func SendAddrsFromPlatform(addrsList string) {
	parts := strings.Split(addrsList, "#")
	rtkUtils.GetAddrsFromPlatform(parts)
}

func SendNetInterfaces(name string, index int) {
	log.Printf("[%s] SendNetInterfaces [%s][%d]", rtkMisc.GetFuncInfo(), name, index)
	rtkUtils.SetNetInterfaces(name, index)
}

func SendMultiFilesDropRequest(multiFilesData string) int {
	var multiFileInfo MultiFilesDropRequestInfo
	err := json.Unmarshal([]byte(multiFilesData), &multiFileInfo)
	if err != nil {
		log.Printf("[%s] Unmarshal[%s] err:%+v", rtkMisc.GetFuncInfo(), multiFilesData, err)
		return int(rtkCommon.SendFilesRequestParameterErr)
	}
	log.Printf("id:[%s] ip:[%s] len:[%d] timestamp:[%d]", multiFileInfo.Id, multiFileInfo.Ip, len(multiFileInfo.PathList), multiFileInfo.TimeStamp)

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

	timestamp := uint64(multiFileInfo.TimeStamp)
	if multiFileInfo.TimeStamp == 0 {
		timestamp = uint64(time.Now().UnixMilli())
	}
	log.Printf("[%s] ID[%s] IP:[%s] get file count:[%d] folder count:[%d], totalSize:[%d] totalDesc:[%s] timestamp:[%d]", rtkMisc.GetFuncInfo(), multiFileInfo.Id, multiFileInfo.Ip, len(fileList), len(folderList), totalSize, totalDesc, timestamp)
	return int(rtkPlatform.GoMultiFilesDropRequest(multiFileInfo.Id, &fileList, &folderList, totalSize, timestamp, totalDesc))
}

func CancelFileTrans(ip, id string, timestamp int64) {
	log.Printf("[%s]  ID:[%s] IP:[%s]  timestamp[%d]", rtkMisc.GetFuncInfo(), id, ip, timestamp)
	rtkPlatform.GoCancelFileTrans(ip, id, timestamp)
}

func SetHostListenAddr(listenHost string, listenPort int) {
	log.Printf("[%s] SetHostListAddr:[%s][%d]", rtkMisc.GetFuncInfo(), listenHost, listenPort)
	rtkPlatform.GoSetHostListenAddr(listenHost, listenPort)
}

func SetDIASID(DiasID string) {
	log.Printf(" [%s]  DiasID:[%s]", rtkMisc.GetFuncInfo(), DiasID)
	//rtkPlatform.GoGetMacAddress(DiasID)
}

func SetDetectPluginEvent(isPlugin bool, productName string) {
	log.Printf(" [%s] isPlugin:[%+v]  productName:[%s]", rtkMisc.GetFuncInfo(), isPlugin, productName)
	rtkPlatform.GoTriggerDetectPluginEvent(isPlugin, productName)
}

func SetConfirmDocumentsAccept(ifConfirm bool) {
	log.Printf("[%s], ifConfirm:[%+v]", rtkMisc.GetFuncInfo(), ifConfirm)
	rtkPlatform.SetConfirmDocumentsAccept(ifConfirm)
}

func GetVersion() string {
	return rtkPlatform.GoGetClientVersion()
}

func GetBuildDate() string {
	return rtkBuildConfig.BuildDate
}

func SetupAppLink(link string) {
	log.Printf("[%s] link:[%s]", rtkMisc.GetFuncInfo(), link)
	rtkPlatform.GoSetupAppLink(link)
}
