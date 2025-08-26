package main

/*
#include <stdlib.h>

typedef struct CLIENT_INFO_DATA
{
    int index;
    char clientId[64];
	char host[64];
	char ipAddr[24];
	int source;
	int port;
	char deviceName[64];
	char platform[16];
	int online;
	int authStatus;
	char updateTime[32];
	char createTime[32];
} CLIENT_INFO_DATA;

typedef struct TIMING_DATA
{
	int source;
	int port;
	int width;
	int height;
	int framerate;
	int displayMode;
	char* displayName; // for Miracast
	char* deviceName; // for Miracast
} TIMING_DATA;

// Callback function type from C++ (AIDL)
typedef void (*UpdateDeviceNameCb)(int source, int port, const char* name);
typedef void (*DragFileStartCb)(int source, int port, int horzSize, int vertSize, int posX, int posY);
typedef void (*UpdateClientInfoCb)(const CLIENT_INFO_DATA clientInfo);
typedef void (*GetTimingDataCb)(TIMING_DATA** list, int* size);
typedef void (*DisplayMonitorNameCb)();

// store callback pointer
static UpdateDeviceNameCb g_updateDeviceNameCb = 0;
static DragFileStartCb g_dragFileStartCb = 0;
static UpdateClientInfoCb g_updateClientInfoCb = 0;
static GetTimingDataCb g_getTimingDataCb = 0;
static DisplayMonitorNameCb g_displayMonitorNameCb = 0;

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
static void setDisplayMonitorNameCb(DisplayMonitorNameCb cb) {g_displayMonitorNameCb = cb;}
static void onDisplayMonitorNameCb() {
	if (g_displayMonitorNameCb) {
		g_displayMonitorNameCb();
	}
}
*/
import "C"
import (
	"reflect"
	rtkCommon "rtk-cross-share/lanServer/common"
	rtkIfaceMgr "rtk-cross-share/lanServer/interfaceMgr"
	rtkMisc "rtk-cross-share/misc"
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

//export SetDisplayMonitorNameCb
func SetDisplayMonitorNameCb(cb C.DisplayMonitorNameCb) {
	C.setDisplayMonitorNameCb(cb)
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

//export UpdateMonitorName
func UpdateMonitorName(cName *C.char) {
	name := C.GoString(cName)
	rtkIfaceMgr.GetInterfaceMgr().UpdateMonitorName(name)
}

//export Init
func Init() {
	initFunc()
}

//export InitWithName
func InitWithName(cMonitorName *C.char) {
	UpdateMonitorName(cMonitorName)
	initFunc()
}

func initFunc() {
	rtkIfaceMgr.GetInterfaceMgr().SetupCallback(
		goUpdateDeviceNameCb,
		goDragFileStartCb,
		goUpdateClientInfoCb,
		goDisplayMonitorNameCb,
		goGetTimingDataCb)

	rtkMisc.GoSafe(func() {
		MainInit()
	})

	select {}
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
}

func goDisplayMonitorNameCb() {
	C.onDisplayMonitorNameCb()
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
			DeviceName:  C.GoString(cSlice[i].deviceName),
		}
		C.free(unsafe.Pointer(cSlice[i].displayName))
		C.free(unsafe.Pointer(cSlice[i].deviceName))
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

func copyStringToFixedArray(dst interface{}, src string) {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Array {
		return
	}

	arr := v.Elem()
	maxLen := arr.Len()

	b := []byte(src)
	length := len(b)
	if length >= maxLen {
		length = maxLen - 1 // for end '\0'
	}

	basePtr := arr.UnsafeAddr()

	for i := 0; i < length; i++ {
		ptr := (*C.char)(unsafe.Pointer(basePtr + uintptr(i)))
		*ptr = C.char(b[i])
	}

	// null terminator
	ptr := (*C.char)(unsafe.Pointer(basePtr + uintptr(length)))
	*ptr = 0
}

func goToCClientInfo(clientInfo rtkCommon.ClientInfoTb) C.CLIENT_INFO_DATA {
	var cData C.CLIENT_INFO_DATA
	cData.index = C.int(clientInfo.Index)
	copyStringToFixedArray(&cData.clientId, clientInfo.ClientId)
	copyStringToFixedArray(&cData.host, clientInfo.Host)
	copyStringToFixedArray(&cData.ipAddr, clientInfo.IpAddr)
	cData.source = C.int(clientInfo.Source)
	cData.port = C.int(clientInfo.Port)
	copyStringToFixedArray(&cData.deviceName, clientInfo.DeviceName)
	copyStringToFixedArray(&cData.platform, clientInfo.Platform)
	cData.online = goBoolToCInt(clientInfo.Online)
	cData.authStatus = goBoolToCInt(clientInfo.AuthStatus)
	copyStringToFixedArray(&cData.updateTime, clientInfo.UpdateTime)
	copyStringToFixedArray(&cData.createTime, clientInfo.CreateTime)
	return cData
}
