package main

/*
#include <stdlib.h>

typedef struct CLIENT_INFO_DATA
{
    int index;
    char* clientId;
	char* host;
	char* ipAddr;
	int source;
	int port;
	char* deviceName;
	char* platform;
	int online;
	int authStatus;
	char* updateTime;
	char* createTime;
} CLIENT_INFO_DATA;

typedef struct TIMING_DATA
{
	int source;
	int port;
	int width;
	int height;
	int framerate;
	int displayMode;
	char* displayName;
} TIMING_DATA;

// Callback function type from C++ (AIDL)
typedef void (*UpdateDeviceNameCb)(int source, int port, const char* name);
typedef void (*DragFileStartCb)(int source, int port, int horzSize, int vertSize, int posX, int posY);
typedef void (*UpdateClientInfoCb)(const CLIENT_INFO_DATA clientInfo);
typedef void (*GetTimingDataCb)(TIMING_DATA** list, int* size);

// store callback pointer
static UpdateDeviceNameCb g_updateDeviceNameCb = 0;
static DragFileStartCb g_dragFileStartCb = 0;
static UpdateClientInfoCb g_updateClientInfoCb = 0;
static GetTimingDataCb g_getTimingDataCb = 0;

// function GO can call to invoke callback
static void setUpdateDeviceNameCb(UpdateDeviceNameCb cb) {g_updateDeviceNameCb = cb;}
static void onUpdateDeviceName(int source, int port, const char* name) {
	if (g_updateDeviceNameCb) {
		g_updateDeviceNameCb(source, port, name);
	}
}
static void setDragFileStartCb(DragFileStartCb cb) {g_dragFileStartCb = cb;}
static void onDragFileStart(int source, int port, int horzSize, int vertSize, int posX, int posY) {
	if (g_dragFileStartCb) {
		g_dragFileStartCb(source, port, horzSize, vertSize, posX, posY);
	}
}
static void setUpdateClientInfoCb(UpdateClientInfoCb cb) {g_updateClientInfoCb = cb;}
static void onUpdateClientInfoCb(const CLIENT_INFO_DATA clientInfo) {
	if (g_updateClientInfoCb) {
		g_updateClientInfoCb(clientInfo);
	}
}
static void setGetTimingDataCb(GetTimingDataCb cb) {g_getTimingDataCb = cb;}
static void onGetTimingDataCb(TIMING_DATA** list, int* size) {
	if (g_getTimingDataCb) {
		g_getTimingDataCb(list, size);
	}
}
*/
import "C"
import (
	rtkCommon "rtk-cross-share/lanServer/common"
	rtkIfaceMgr "rtk-cross-share/lanServer/interfaceMgr"
	"unsafe"
)

//export SetUpdateDeviceNameCb
func SetUpdateDeviceNameCb(cb C.UpdateDeviceNameCb) {
	C.setUpdateDeviceNameCb(cb)
}

//export SetDragFileStartCb
func SetDragFileStartCb(cb C.DragFileStartCb) {
	C.setDragFileStartCb(cb)
}

//export SetUpdateClientInfoCb
func SetUpdateClientInfoCb(cb C.UpdateClientInfoCb) {
	C.setUpdateClientInfoCb(cb)
}

//export SetGetTimingDataCb
func SetGetTimingDataCb(cb C.GetTimingDataCb) {
	C.setGetTimingDataCb(cb)
}

//export AuthDevice
func AuthDevice(cSource, cPort, cIndex C.int) C.int {
	source := int(cSource)
	port := int(cPort)
	index := int(cIndex)
	if rtkIfaceMgr.GetInterfaceMgr().AuthDevice(source, port, index) {
		return 1
	} else {
		return 0
	}
}

//export UpdateMousePos
func UpdateMousePos(cSource, cPort, cHorzSize, cVertSize, cPosX, cPosY C.int) {
	source := int(cSource)
	port := int(cPort)
	horzSize := int(cHorzSize)
	vertSize := int(cVertSize)
	posX := int(cPosX)
	posY := int(cPosY)
	rtkIfaceMgr.GetInterfaceMgr().UpdateMousePos(source, port, horzSize, vertSize, posX, posY)
}

//export DragFileEnd
func DragFileEnd(cSource, cPort C.int) {
	source := int(cSource)
	port := int(cPort)
	rtkIfaceMgr.GetInterfaceMgr().DragFileEnd(source, port)
}

//export GetDiasId
func GetDiasId() *C.char {
	return C.CString(rtkIfaceMgr.GetInterfaceMgr().GetDiasId())
}

//export GetDeviceName
func GetDeviceName(cSource, cPort C.int) *C.char {
	source := int(cSource)
	port := int(cPort)
	return C.CString(rtkIfaceMgr.GetInterfaceMgr().GetDeviceName(source, port))
}

//export UpdateMiracastInfo
func UpdateMiracastInfo(cIp *C.char, cMacAddr *C.uchar, cName *C.char) {
	ip := C.GoString(cIp)
	macAddr := C.GoBytes(unsafe.Pointer(cMacAddr), 6)
	name := C.GoString(cName)
	rtkIfaceMgr.GetInterfaceMgr().UpdateMiracastInfo(ip, macAddr, name)
}

//export GetClientInfoData
func GetClientInfoData(cSource, cPort C.int) C.CLIENT_INFO_DATA {
	source := int(cSource)
	port := int(cPort)
	clientInfoData := rtkIfaceMgr.GetInterfaceMgr().GetClientInfodData(source, port)
	return goToCClientInfo(clientInfoData)
}

//export Init
func Init() {
	rtkIfaceMgr.GetInterfaceMgr().SetupCallback(goUpdateDeviceNameCb, goDragFileStartCb, goUpdateClientInfoCb, goGetTimingDataCb)
	MainInit()
}

func goUpdateDeviceNameCb(source, port int, name string) {
	cSource := C.int(source)
	cPort := C.int(port)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	C.onUpdateDeviceName(cSource, cPort, cName)
}

func goDragFileStartCb(source, port, horzSize, vertSize, posX, posY int) {
	cSource := C.int(source)
	cPort := C.int(port)
	cHorzSize := C.int(horzSize)
	cVertSize := C.int(vertSize)
	cPosX := C.int(posX)
	cPosY := C.int(posY)
	C.onDragFileStart(cSource, cPort, cHorzSize, cVertSize, cPosX, cPosY)
}

func goUpdateClientInfoCb(clientInfo rtkCommon.ClientInfoTb) {
	C.onUpdateClientInfoCb(goToCClientInfo(clientInfo))
	// TODO: Free CString
}

func goGetTimingDataCb() []rtkCommon.TimingData {
	var cList *C.TIMING_DATA
	var cSize C.int
	defer C.free(unsafe.Pointer(cList))
	C.onGetTimingDataCb(&cList, &cSize)

	size := int(cSize)
	cSlice := unsafe.Slice(cList, size)
	result := make([]rtkCommon.TimingData, size)
	for i := 0; i < size; i++ {
		result[i] = rtkCommon.TimingData{
			Source:      int(cSlice[i].source),
			Port:        int(cSlice[i].port),
			Width:       int(cSlice[i].width),
			Height:      int(cSlice[i].height),
			Framerate:   int(cSlice[i].framerate),
			DisplayMode: int(cSlice[i].displayMode),
			DisplayName: C.GoString(cSlice[i].displayName),
		}
		C.free(unsafe.Pointer(cSlice[i].displayName))
	}

	return result
}

func main() {
}

func goBoolToCInt(val bool) C.int {
	if val {
		return C.int(1)
	} else {
		return C.int(0)
	}
}

func goToCClientInfo(clientInfo rtkCommon.ClientInfoTb) C.CLIENT_INFO_DATA {
	return C.CLIENT_INFO_DATA{
		index:      C.int(clientInfo.Index),
		clientId:   C.CString(clientInfo.ClientId),
		host:       C.CString(clientInfo.Host),
		ipAddr:     C.CString(clientInfo.IpAddr),
		source:     C.int(clientInfo.Source),
		port:       C.int(clientInfo.Port),
		deviceName: C.CString(clientInfo.DeviceName),
		platform:   C.CString(clientInfo.Platform),
		online:     goBoolToCInt(clientInfo.Online),
		authStatus: goBoolToCInt(clientInfo.AuthStatus),
		updateTime: C.CString(clientInfo.UpdateTime),
		createTime: C.CString(clientInfo.CreateTime),
	}
}
