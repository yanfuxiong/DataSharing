//go:build android

package libp2p_clipboard

import (
	"log"
	rtkBuildConfig "rtk-cross-share/client/buildConfig"
	rtkCmd "rtk-cross-share/client/cmd"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strings"
)

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

func SendXClipData(text, image, html /*, rtf*/ string) {
	log.Printf("[%s] text:%d, image:%d, html:%d, rtf:%d \n\n", rtkMisc.GetFuncInfo(), len(text), len(image), len(html) /*, len(rtf)*/)
	rtkPlatform.GoCopyXClipData(text, image, html, "" /* []byte(rtf)*/)
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
	return int(rtkPlatform.GoMultiFilesDropRequest(multiFilesData))
}

func IfClipboardPasteFile(fileName, id string, isReceive bool) {
	/*FilePath := rtkPlatform.GetDownloadPath()
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
	}*/
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
	rtkPlatform.GoSetHostListenAddr(listenHost, listenPort)
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
