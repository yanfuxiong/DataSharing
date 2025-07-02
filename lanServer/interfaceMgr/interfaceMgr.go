package interfaceMgr

import (
	"fmt"
	"log"
	rtkClientManager "rtk-cross-share/lanServer/clientManager"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"
	"sync"
)

const (
	tag = "InterfaceMgr"
)

var (
	instance *InterfaceMgr
	once     sync.Once
)

type (
	UpdateDeviceNameCb func(source, port int, name string)
	DragFileStartCb    func(source, port, horzSize, vertSize, posX, posY int)
	DragFileSrcInfo    struct {
		index int
		id    string
	}
	InterfaceMgr struct {
		mu                  sync.RWMutex
		mUpdateDeviceNameCb UpdateDeviceNameCb
		mDragFileStartCb    DragFileStartCb
		mDragFileSrcInfo    DragFileSrcInfo
	}
)

func GetInterfaceMgr() *InterfaceMgr {
	once.Do(func() {
		instance = &InterfaceMgr{}
	})
	return instance
}

func (mgr *InterfaceMgr) SetupCallback(updateDeviceNameCb UpdateDeviceNameCb, dragFileStartCb DragFileStartCb) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.mUpdateDeviceNameCb = updateDeviceNameCb
	mgr.mDragFileStartCb = dragFileStartCb
}

func (mgr *InterfaceMgr) triggerUpdateDeviceName(source, port int, name string) bool {
	mgr.mu.RLock()
	if mgr.mUpdateDeviceNameCb == nil {
		log.Printf("[%s][%s] Error: UpdateDevice callback is null", tag, rtkMisc.GetFuncInfo())
		return false
	}
	mgr.mu.RUnlock()

	mgr.mUpdateDeviceNameCb(source, port, name)
	return true
}

func (mgr *InterfaceMgr) triggerDragFileStart(source, port, horzSize, vertSize, posX, posY int) bool {
	mgr.mu.RLock()
	if mgr.mDragFileStartCb == nil {
		log.Printf("[%s][%s] Error: DragFileStart callback is null", tag, rtkMisc.GetFuncInfo())
		return false
	}
	mgr.mu.RUnlock()

	mgr.mDragFileStartCb(source, port, horzSize, vertSize, posX, posY)
	return true
}

func (mgr *InterfaceMgr) AuthDevice(source, port, index int) bool {
	authStatus := true
	errAuthAndSrcPort := rtkdbManager.UpdateAuthAndSrcPort(index, authStatus, source, port)
	if errAuthAndSrcPort != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: update auth status and source port failed: %s", tag, rtkMisc.GetFuncInfo(), rtkMisc.GetResponse(errAuthAndSrcPort).Msg)
		return false
	}

	deviceName, errDeviceName := rtkdbManager.QueryDeviceName(index)
	if errDeviceName != nil {
		log.Printf("[%s][%s] Error: query device name failed: %s", tag, rtkMisc.GetFuncInfo(), errDeviceName.Error())
		return false
	}

	mgr.triggerUpdateDeviceName(source, port, deviceName)
	return true
}

func (mgr *InterfaceMgr) UpdateMousePos(source, port, horzSize, vertSize, posX, posY int) {
	clientIdx, clientId, err := rtkdbManager.QueryClientBySrcPort(source, port)
	if err != nil {
		log.Printf("[%s][%s] Error: get client by (source,port):(%d,%d) failed: %s", tag, rtkMisc.GetFuncInfo(), source, port, err.Error())
		return
	}

	if mgr.triggerDragFileStart(source, port, horzSize, vertSize, posX, posY) {
		mgr.mDragFileSrcInfo = DragFileSrcInfo{clientIdx, clientId}
	}
}

func (mgr *InterfaceMgr) DragFileEnd(source, port int) {
	_, dstClientId, err := rtkdbManager.QueryClientBySrcPort(source, port)
	if err != nil {
		log.Printf("[UnixSocket][%s] Error: get client by (source,port):(%d,%d) failed: %s", rtkMisc.GetFuncInfo(), source, port, err.Error())
		return
	}

	rtkClientManager.SendDragFileEvent(mgr.mDragFileSrcInfo.id, dstClientId, uint32(mgr.mDragFileSrcInfo.index))
}

func (mgr *InterfaceMgr) GetDiasId() string {
	return rtkGlobal.ServerMdnsId
}

func (mgr *InterfaceMgr) GetDeviceName(source, port int) string {
	deviceName, errDeviceName := rtkdbManager.QueryDeviceNameBySrcPort(source, port)
	if errDeviceName != nil {
		log.Printf("[%s][%s] Error: query device name failed: %s", tag, rtkMisc.GetFuncInfo(), errDeviceName.Error())
		return ""
	}

	return deviceName
}

func (mgr *InterfaceMgr) UpdateMiracastInfo(ip string, macAddr []byte, name string) {
	if ip == "" || len(macAddr) != 6 {
		log.Printf("[%s][%s] Error: Invalid miracast info: IP=%s, macAddr len=%d, name=%s", ip, len(macAddr), name)
		return
	}

	// TODO:
	macAddrStr := fmt.Sprintf("%02X%02X%02X%02X%02X%02X", macAddr[0], macAddr[1], macAddr[2], macAddr[3], macAddr[4], macAddr[5])
	log.Printf("[%s][%s] TODO: Add miracast info to Database AuthTable. IP=%s, macAddr=%s, name=%s", ip, macAddrStr, name)
}
