package interfaceMgr

import (
	"fmt"
	"log"
	rtkClientManager "rtk-cross-share/lanServer/clientManager"
	rtkCommon "rtk-cross-share/lanServer/common"
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
	UpdateClientInfoCb func(clientInfo rtkCommon.ClientInfoTb)
	GetTimingDataCb    func() []rtkCommon.TimingData
	InterfaceMgr       struct {
		mu                  sync.RWMutex
		mUpdateDeviceNameCb UpdateDeviceNameCb
		mDragFileStartCb    DragFileStartCb
		mDragFileSrcInfo    DragFileSrcInfo
		mUpdateClientInfoCb UpdateClientInfoCb
		mGetTimingDataCb    GetTimingDataCb
	}
)

func GetInterfaceMgr() *InterfaceMgr {
	once.Do(func() {
		instance = &InterfaceMgr{}
		instance.initDbCallback()
	})
	return instance
}

func (mgr *InterfaceMgr) initDbCallback() {
	rtkdbManager.SetNotifyUpdateClientInfoCallback(mgr.TriggerUpdateClientInfo)
	rtkClientManager.SetNotifyGetTimingDataCallback(mgr.TriggerGetTimingData)
}

func (mgr *InterfaceMgr) SetupCallback(
	updateDeviceNameCb UpdateDeviceNameCb,
	dragFileStartCb DragFileStartCb,
	updateClientInfoCb UpdateClientInfoCb,
	getTimingDataCb GetTimingDataCb,
) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.mUpdateDeviceNameCb = updateDeviceNameCb
	mgr.mDragFileStartCb = dragFileStartCb
	mgr.mUpdateClientInfoCb = updateClientInfoCb
	mgr.mGetTimingDataCb = getTimingDataCb
}

// Deprecated: Use UpdateClientInfodData
func (mgr *InterfaceMgr) TriggerUpdateDeviceName(source, port int, name string) bool {
	mgr.mu.RLock()
	if mgr.mUpdateDeviceNameCb == nil {
		log.Printf("[%s][%s] Error: UpdateDevice callback is null", tag, rtkMisc.GetFuncInfo())
		return false
	}
	mgr.mu.RUnlock()

	mgr.mUpdateDeviceNameCb(source, port, name)
	return true
}

func (mgr *InterfaceMgr) TriggerDragFileStart(source, port, horzSize, vertSize, posX, posY int) bool {
	mgr.mu.RLock()
	if mgr.mDragFileStartCb == nil {
		log.Printf("[%s][%s] Error: DragFileStart callback is null", tag, rtkMisc.GetFuncInfo())
		return false
	}
	mgr.mu.RUnlock()

	mgr.mDragFileStartCb(source, port, horzSize, vertSize, posX, posY)
	return true
}

func (mgr *InterfaceMgr) TriggerUpdateClientInfo(clientInfo rtkCommon.ClientInfoTb) {
	mgr.mu.RLock()
	if mgr.mUpdateClientInfoCb == nil {
		log.Printf("[%s][%s] Error: UpdateClientInfo callback is null", tag, rtkMisc.GetFuncInfo())
		return
	}
	mgr.mu.RUnlock()

	mgr.mUpdateClientInfoCb(clientInfo)
}

func (mgr *InterfaceMgr) TriggerGetTimingData() []rtkCommon.TimingData {
	mgr.mu.RLock()
	if mgr.mGetTimingDataCb == nil {
		log.Printf("[%s][%s] Error: GetTimingData callback is null", tag, rtkMisc.GetFuncInfo())
		return make([]rtkCommon.TimingData, 0)
	}
	mgr.mu.RUnlock()

	return mgr.mGetTimingDataCb()
}

func (mgr *InterfaceMgr) AuthDevice(source, port, index int) bool {
	authStatus := true
	errAuthAndSrcPort := rtkdbManager.UpdateAuthAndSrcPort(index, authStatus, source, port)
	if errAuthAndSrcPort != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: update auth status and source port failed: %s", tag, rtkMisc.GetFuncInfo(), rtkMisc.GetResponse(errAuthAndSrcPort).Msg)
		return false
	}

	clientInfoTb, err := rtkdbManager.QueryClientInfoByIndex(index)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: query device name failed: %d", tag, rtkMisc.GetFuncInfo(), err)
		return false
	}

	mgr.TriggerUpdateDeviceName(source, port, clientInfoTb.DeviceName)
	return true
}

func (mgr *InterfaceMgr) UpdateMousePos(source, port, horzSize, vertSize, posX, posY int) {
	clientInfoTb, err := rtkdbManager.QueryClientInfoBySrcPort(source, port)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: get client by (source,port):(%d,%d) failed: %d", tag, rtkMisc.GetFuncInfo(), source, port, err)
		return
	}

	if (clientInfoTb.Online == false) || (clientInfoTb.AuthStatus == false) {
		log.Printf("[%s][%s] Error: not valid client (source,port):(%d,%d) online:%t, authStatus:%t",
			tag, rtkMisc.GetFuncInfo(), source, port, clientInfoTb.Online, clientInfoTb.AuthStatus)
		return
	}

	if mgr.TriggerDragFileStart(source, port, horzSize, vertSize, posX, posY) {
		mgr.mDragFileSrcInfo = DragFileSrcInfo{clientInfoTb.Index, clientInfoTb.ClientId}
	}
}

func (mgr *InterfaceMgr) DragFileEnd(source, port int) {
	clientInfoTb, err := rtkdbManager.QueryClientInfoBySrcPort(source, port)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: get client by (source,port):(%d,%d) failed: %d", tag, rtkMisc.GetFuncInfo(), source, port, err)
		return
	}

	if (clientInfoTb.Online == false) || (clientInfoTb.AuthStatus == false) {
		log.Printf("[%s][%s] Error: not valid client (source,port):(%d,%d) online:%t, authStatus:%t",
			tag, rtkMisc.GetFuncInfo(), source, port, clientInfoTb.Online, clientInfoTb.AuthStatus)
		return
	}

	rtkClientManager.SendDragFileEvent(mgr.mDragFileSrcInfo.id, clientInfoTb.ClientId, uint32(mgr.mDragFileSrcInfo.index))
}

func (mgr *InterfaceMgr) GetDiasId() string {
	return rtkGlobal.ServerMdnsId
}

// Deprecated: Use GetClientInfodData
func (mgr *InterfaceMgr) GetDeviceName(source, port int) string {
	clientInfoTb, err := rtkdbManager.QueryClientInfoBySrcPort(source, port)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: query device name failed: %d", tag, rtkMisc.GetFuncInfo(), err)
		return ""
	}

	return clientInfoTb.DeviceName
}

func (mgr *InterfaceMgr) UpdateMiracastInfo(ip string, macAddr []byte, name string) {
	if ip == "" || len(macAddr) != 6 {
		log.Printf("[%s][%s] Error: Invalid miracast info: IP=%s, macAddr len=%d, name=%s", tag, rtkMisc.GetFuncInfo(), ip, len(macAddr), name)
		return
	}

	// TODO:
	macAddrStr := fmt.Sprintf("%02X%02X%02X%02X%02X%02X", macAddr[0], macAddr[1], macAddr[2], macAddr[3], macAddr[4], macAddr[5])
	log.Printf("[%s][%s] TODO: Add miracast info to Database AuthTable. IP=%s, macAddr=%s, name=%s", tag, rtkMisc.GetFuncInfo(), ip, macAddrStr, name)
}

func (mgr *InterfaceMgr) GetClientInfodData(source, port int) rtkCommon.ClientInfoTb {
	clientInfoTb, err := rtkdbManager.QueryClientInfoBySrcPort(source, port)
	if err != rtkMisc.SUCCESS && err != rtkMisc.ERR_DB_SQLITE_EMPTY_RESULT {
		log.Printf("[%s][%s] Error: query clientInfo data failed: %d", tag, rtkMisc.GetFuncInfo(), err)
		return rtkCommon.ClientInfoTb{}
	}

	return clientInfoTb
}
