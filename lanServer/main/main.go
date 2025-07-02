package main

/*
#include <stdlib.h>

// Callback function type from C++ (AIDL)
typedef void (*UpdateDeviceNameCb)(int source, int port, const char* name);
typedef void (*DragFileStartCb)(int source, int port, int horzSize, int vertSize, int posX, int posY);

// store callback pointer
static UpdateDeviceNameCb g_updateDeviceNameCb = 0;
static DragFileStartCb g_dragFileStartCb = 0;

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
*/
import "C"
import (
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

//export Init
func Init() {
	rtkIfaceMgr.GetInterfaceMgr().SetupCallback(goUpdateDeviceNameCb, goDragFileStartCb)
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

func main() {
}
