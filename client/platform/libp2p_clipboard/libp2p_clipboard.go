//go:build android

package libp2p_clipboard

import (
	"fmt"
	"log"
	rtkCmd "rtk-cross-share/cmd"
	rtkCommon "rtk-cross-share/common"
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
	rtkCmd.MainInit(cb, serverId, serverIpInfo, listentHost, listentPort)
}

func SetMainCallback(cb Callback) {
	rtkPlatform.SetCallback(cb)
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

func SendCopyFile(filePath, ipAddr string, fileSize int64) {
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

	rtkPlatform.GoFileDropRequest(ipAddr, fileInfo, int64(time.Now().Unix()))
	log.Printf("(SRC)Send file:[%s][%s], fileSize:%d", filePath, ipAddr, fileSize)
}

func IfClipboardPasteFile(fileName, ipAddr string, isReceive bool) {
	FilePath := rtkPlatform.GetReceiveFilePath()
	if fileName != "" {
		FilePath += fileName
	} else {
		FilePath += fmt.Sprintf("recevieFrom-%s.file", ipAddr)
	}

	if isReceive {
		rtkPlatform.GoFileDropResponse(ipAddr, rtkCommon.FILE_DROP_ACCEPT, FilePath)
		log.Printf("(DST) FilePath:[%s] from ipAddr:[%s], confirm receipt", FilePath, ipAddr)
	} else {
		rtkPlatform.GoFileDropResponse(ipAddr, rtkCommon.FILE_DROP_REJECT, "")
		log.Printf("(DST) FilePath:[%s] from ipAddr:[%s] reject", FilePath, ipAddr)
	}
}
