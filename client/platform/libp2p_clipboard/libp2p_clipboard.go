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
	Id       string
	Ip       string
	PathList []string
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
	if !rtkPlatform.IsConnecting() {
		rtkPlatform.GoConnectLanServer(instance)
	} else {
		log.Printf("[%s] connecting in progress, skip it!", rtkMisc.GetFuncInfo())
	}
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
	rtkCmd.MainInit(serverId, serverIpInfo, listentHost, listentPort)
}

func SendMessage(s string) {
	rtkPlatform.SendMessage(s)
}

func GetClientList() string {
	clientList := rtkUtils.GetClientList()
	log.Printf("GetClientList :[%s]", clientList)
	return clientList
}

func SendImage(content string) {
	if content == "" || len(content) == 0 {
		return
	}
	data := rtkUtils.Base64Decode(content)
	if data == nil {
		return
	}

	w, h, size := rtkUtils.GetByteImageInfo(data)
	if size == 0 {
		log.Println("GetByteImageInfo err!")
		return
	}
	log.Printf("SendImage:[%d][%d][%d]", len(content), len(data), size)

	imgHeader := rtkCommon.ImgHeader{
		Width:       int32(w),
		Height:      int32(h),
		Planes:      1,
		BitCount:    uint16((size * 8) / (w * h)),
		Compression: 0,
	}
	// FIXME
	fileSize := rtkCommon.FileSize{
		SizeHigh: 0,
		SizeLow:  uint32(size),
	}
	rtkPlatform.GoCopyImage(fileSize, imgHeader, rtkUtils.ImageToBitmap(data))
}

func SendAddrsFromPlatform(addrsList string) {
	parts := strings.Split(addrsList, "#")
	rtkUtils.GetAddrsFromPlatform(parts)
}

func SendNetInterfaces(name string, index int) {
	log.Printf("[%s] SendNetInterfaces [%s][%d]", rtkMisc.GetFuncInfo(), name, index)
	rtkUtils.SetNetInterfaces(name, index)
}

func SendCopyFile(filePath, id string, fileSize int64) {
	if filePath == "" || len(filePath) == 0 || fileSize == 0 {
		log.Printf("filePath:[%s] or fileSize:[%d] is null", filePath, fileSize)
		return
	}
	if !rtkMisc.FileExists(filePath) {
		log.Printf("[%s] filePath:[%s] is not exists!", rtkMisc.GetFuncInfo(), filePath)
		return
	}

	low := uint32(fileSize & 0xFFFFFFFF)
	high := uint32(fileSize >> 32)

	var fileInfo = rtkCommon.FileInfo{
		FileSize_: rtkCommon.FileSize{
			SizeHigh: high,
			SizeLow:  low,
		},
		FilePath: filePath,
		FileName: filepath.Base(filePath),
	}

	rtkPlatform.GoFileDropRequest(id, fileInfo, uint64(time.Now().UnixMilli()))
	log.Printf("(SRC)Send file:[%s] to [%s], fileSize:%d", filePath, id, fileSize)
}

func SendMultiFilesDropRequest(multiFilesData string) int {
	var multiFileInfo MultiFilesDropRequestInfo
	err := json.Unmarshal([]byte(multiFilesData), &multiFileInfo)
	if err != nil {
		log.Printf("[%s] Unmarshal[%s] err:%+v", rtkMisc.GetFuncInfo(), multiFilesData, err)
		return int(rtkCommon.SendFilesRequestParameterErr)
	}
	log.Printf("id:[%s] ip:[%s] len:[%d]", multiFileInfo.Id, multiFileInfo.Ip, len(multiFileInfo.PathList))

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
