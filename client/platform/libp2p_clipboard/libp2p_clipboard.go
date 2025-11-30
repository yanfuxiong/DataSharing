//go:build android

package libp2p_clipboard

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

func WorkerConnectLanServer(instance string) {
	log.Printf("[%s]  instance:[%s]", rtkMisc.GetFuncInfo(), instance)
	rtkPlatform.GoConnectLanServer(instance)
}

func BrowseLanServer() {
	log.Printf("[%s]", rtkMisc.GetFuncInfo())
	rtkPlatform.GoBrowseLanServer()
}

func MainInit(cb Callback, rootPath, deviceName, serverId, serverIpInfo, listentHost string, listentPort int) {
	rtkPlatform.SetCallback(cb)
	rtkPlatform.SetDeviceName(deviceName)

	if rootPath == "" || !rtkMisc.FolderExists(rootPath) {
		log.Fatalf("[%s] RootPath :[%s] is invalid!", rtkMisc.GetFuncInfo(), rootPath)
	}
	rtkPlatform.SetupRootPath(rootPath)

	log.Printf("[%s] rootPath:[%s] device name:[%s] host:[%s] port:[%d]", rtkMisc.GetFuncInfo(), rootPath, deviceName, listentHost, listentPort)
	rtkCmd.MainInit(serverId, serverIpInfo, listentHost, listentPort)
}

func SetMsgEventFunc(event int, arg1, arg2, arg3, arg4 string) {
	log.Printf("[%s] event:[%d], arg1:%s, arg2:%s, arg3:%s, arg4:%s\n", rtkMisc.GetFuncInfo(), event, arg1, arg2, arg3, arg4)
	rtkPlatform.GoSetMsgEventFunc(uint32(event), arg1, arg2, arg3, arg4)
}

func SendXClipData(text, image, html, rtf string) {
	log.Printf("[%s] text:%d, image:%d, html:%d, rtf:%d \n\n", rtkMisc.GetFuncInfo(), len(text), len(image), len(html), len(rtf))

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

	rtkPlatform.GoCopyXClipData([]byte(text), imgData, []byte(html), []byte(rtf))
}

func GetClientListEx() string {
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

func IfClipboardPasteFile(fileName, id string, isReceive bool) {
	FilePath := rtkPlatform.GetDownloadPath()
	if fileName != "" {
		FilePath = filepath.Join(FilePath, fileName)
	} else {
		FilePath = filepath.Join(FilePath, fmt.Sprintf("recevieFrom-%s_%d", id, time.Now().UnixMilli()))
	}

	if isReceive {
		rtkPlatform.GoFileDropResponse(id, rtkCommon.FILE_DROP_ACCEPT, FilePath)
		log.Printf("(DST) FilePath:[%s] from id:[%s], confirm receipt", FilePath, id)
	} else {
		rtkPlatform.GoFileDropResponse(id, rtkCommon.FILE_DROP_REJECT, "")
		log.Printf("(DST) FilePath:[%s] from id:[%s] reject", FilePath, id)
	}
}

func CancelFileTrans(ip, id string, timestamp int64) {
	log.Printf("[%s]  ID:[%s] IP:[%s]  timestamp[%d]", rtkMisc.GetFuncInfo(), id, ip, timestamp)
	rtkPlatform.GoCancelFileTrans(ip, id, timestamp)
}

func SetNetWorkConnected(isConnect bool) {
	log.Printf("[%s] SetNetWorkConnected:[%v]", rtkMisc.GetFuncInfo(), isConnect)
	rtkPlatform.SetNetWorkConnected(isConnect)
}

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

func SetDIASID(DiasID string) {
	log.Printf(" [%s]  DiasID:[%s]", rtkMisc.GetFuncInfo(), DiasID)
	rtkPlatform.GoGetMacAddress(DiasID)
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
	return rtkGlobal.ClientVersion
}

func GetBuildDate() string {
	return rtkBuildConfig.BuildDate
}

func SetupAppLink(link string) {
	log.Printf("[%s] link:[%s]", rtkMisc.GetFuncInfo(), link)
	rtkPlatform.GoSetupAppLink(link)
}
