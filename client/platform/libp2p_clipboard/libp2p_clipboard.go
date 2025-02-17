//go:build android

package libp2p_clipboard

import (
	"fmt"
	"log"
	rtkBuildConfig "rtk-cross-share/buildConfig"
	rtkCmd "rtk-cross-share/cmd"
	rtkCommon "rtk-cross-share/common"
	rtkConnection "rtk-cross-share/connection"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"strings"
	"time"
)

type Callback interface {
	rtkPlatform.Callback
}

// TODO: consider to replace int with long long type
func MainInit(cb Callback, serverId, serverIpInfo, listentHost string, listentPort int) {
	rtkPlatform.SetCallback(cb)
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

func SendAddrsFromJava(addrsList string) {
	parts := strings.Split(addrsList, "#")
	rtkUtils.GetAddrsFromJava(parts)
}

func SendNetInterfaces(name string, index int) {
	log.Printf("SendNetInterfaces [%s][%d]", name, index)
	rtkUtils.SetNetInterfaces(name, index)
}

func SendCopyFile(filePath, id string, fileSize int64) {
	if filePath == "" || len(filePath) == 0 || fileSize == 0 {
		log.Printf("filePath:[%s] or fileSizeLow:[%d] is null", filePath, fileSize)
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
	}

	rtkPlatform.GoFileDropRequest(id, fileInfo, int64(time.Now().Unix()))
	log.Printf("(SRC)Send file:[%s][%s], fileSize:%d", filePath, id, fileSize)
}

func IfClipboardPasteFile(fileName, id string, isReceive bool) {
	FilePath := rtkPlatform.GetReceiveFilePath()
	if fileName != "" {
		FilePath += fileName
	} else {
		FilePath += fmt.Sprintf("recevieFrom-%s.file", id)
	}

	if isReceive {
		rtkPlatform.GoFileDropResponse(id, rtkCommon.FILE_DROP_ACCEPT, FilePath)
		log.Printf("(DST) FilePath:[%s] from id:[%s], confirm receipt", FilePath, id)
	} else {
		rtkPlatform.GoFileDropResponse(id, rtkCommon.FILE_DROP_REJECT, "")
		log.Printf("(DST) FilePath:[%s] from id:[%s] reject", FilePath, id)
	}
}

func SetNetWorkConnected(isConnect bool) {
	log.Printf("[%s] SetNetWorkConnected:[%v]", rtkUtils.GetFuncInfo(), isConnect)
	rtkPlatform.SetNetWorkConnected(isConnect)
}

func SetHostListenAddr(listenHost string, listentPort int) {
	log.Printf("[%s] SetHostListAddr:[%s][%d]", rtkUtils.GetFuncInfo(), listenHost, listentPort)
	if listenHost != rtkGlobal.ListenHost || listentPort != rtkGlobal.ListenPort {
		log.Println("**************** Attention please, the host listen addr is switch! ********************\n\n")
		rtkGlobal.ListenHost = listenHost
		rtkGlobal.ListenPort = listentPort
		rtkConnection.SetNetworkSwitchFlag()
	}
}

func GetVersion() string {
	return rtkBuildConfig.Version
}

func GetBuildDate() string {
	return rtkBuildConfig.BuildDate
}
