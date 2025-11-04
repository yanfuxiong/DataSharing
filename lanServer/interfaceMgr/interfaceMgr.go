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
	UpdateClientInfoCb       func(clientInfo rtkCommon.ClientInfoTb)
	DisplayMonitorNameCb     func()
	GetDpSrcTypeCb           func(source, port int) rtkGlobal.DpSrcType
	GetTimingDataCb          func() []rtkCommon.TimingData
	GetTimingDataBySrcPortCb func(source, port int) rtkCommon.TimingData
	SendMsgEventCb           func(event int, arg1, arg2, arg3, arg4 string)
	InterfaceMgr             struct {
		mu                        sync.RWMutex
		mUpdateDeviceNameCb       UpdateDeviceNameCb
		mDragFileStartCb          DragFileStartCb
		mDragFileSrcInfo          DragFileSrcInfo
		mUpdateClientInfoCb       UpdateClientInfoCb
		mDisplayMonitorNameCb     DisplayMonitorNameCb
		mGetDpSrcTypeCb           GetDpSrcTypeCb
		mGetTimingDataCb          GetTimingDataCb
		mGetTimingDataBySrcPortCb GetTimingDataBySrcPortCb
		mSendMsgEventCb           SendMsgEventCb
	}
)

func GetInterfaceMgr() *InterfaceMgr {
	once.Do(func() {
		instance = &InterfaceMgr{}
		instance.initCallbackToClient()
	})
	return instance
}

func (mgr *InterfaceMgr) initCallbackToClient() {
	rtkdbManager.SetNotifyUpdateClientInfoCallback(mgr.TriggerUpdateClientInfo)
	rtkClientManager.SetNotifyGetTimingDataCallback(mgr.TriggerGetTimingData)
	rtkClientManager.SetNotifyGetTimingDataBySrcPortCallback(mgr.TriggerGetTimingDataBySrcPort)
	rtkClientManager.SetSendPlatformMsgEventCallback(mgr.TriggerSendMsgEvent)
}

func (mgr *InterfaceMgr) SetupCallbackFromServer(
	updateDeviceNameCb UpdateDeviceNameCb,
	dragFileStartCb DragFileStartCb,
	updateClientInfoCb UpdateClientInfoCb,
	displayMonitorNameCb DisplayMonitorNameCb,
	getDpSrcTypeCb GetDpSrcTypeCb,
	getTimingDataCb GetTimingDataCb,
	getTimingDataBySrcPortCb GetTimingDataBySrcPortCb,
	sendMsgEventCb SendMsgEventCb,
) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.mUpdateDeviceNameCb = updateDeviceNameCb
	mgr.mDragFileStartCb = dragFileStartCb
	mgr.mUpdateClientInfoCb = updateClientInfoCb
	mgr.mDisplayMonitorNameCb = displayMonitorNameCb
	mgr.mGetDpSrcTypeCb = getDpSrcTypeCb
	mgr.mGetTimingDataCb = getTimingDataCb
	mgr.mGetTimingDataBySrcPortCb = getTimingDataBySrcPortCb
	mgr.mSendMsgEventCb = sendMsgEventCb
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

func (mgr *InterfaceMgr) TriggerDisplayMonitorName() {
	mgr.mu.RLock()
	if mgr.mDisplayMonitorNameCb == nil {
		log.Printf("[%s][%s] Error: DisplayMonitorName callback is null", tag, rtkMisc.GetFuncInfo())
		return
	}
	mgr.mu.RUnlock()
	mgr.mDisplayMonitorNameCb()
}

func (mgr *InterfaceMgr) TriggerGetDpSrcTypeCb(source, port int) rtkGlobal.DpSrcType {
	mgr.mu.RLock()
	if mgr.mGetDpSrcTypeCb == nil {
		log.Printf("[%s][%s] Error: GetDpSrcType callback is null", tag, rtkMisc.GetFuncInfo())
		return rtkGlobal.DP_SRC_TYPE_NONE
	}
	mgr.mu.RUnlock()
	return mgr.mGetDpSrcTypeCb(source, port)
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

func (mgr *InterfaceMgr) TriggerGetTimingDataBySrcPort(source, port int) rtkCommon.TimingData {
	mgr.mu.RLock()
	if mgr.mGetTimingDataBySrcPortCb == nil {
		log.Printf("[%s][%s] Error: GetTimingDataBySrcPortCb callback is null", tag, rtkMisc.GetFuncInfo())
		return rtkCommon.TimingData{}
	}
	mgr.mu.RUnlock()

	return mgr.mGetTimingDataBySrcPortCb(source, port)
}

func (mgr *InterfaceMgr) TriggerSendMsgEvent(event int, arg1, arg2, arg3, arg4 string) {
	mgr.mu.RLock()
	if mgr.mSendMsgEventCb == nil {
		log.Printf("[%s][%s] Error: SendMsgEvent callback is null", tag, rtkMisc.GetFuncInfo())
		return
	}
	mgr.mu.RUnlock()

	mgr.mSendMsgEventCb(event, arg1, arg2, arg3, arg4)
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
	clientInfoTbList, err := rtkdbManager.QueryClientInfoBySrcPort(source, port)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: get client by (source,port):(%d,%d) failed: %d", tag, rtkMisc.GetFuncInfo(), source, port, err)
		return
	}

	for _, clientInfo := range clientInfoTbList {
		if (clientInfo.Online == true) || (clientInfo.AuthStatus == true) {
			if mgr.TriggerDragFileStart(source, port, horzSize, vertSize, posX, posY) {
				mgr.mDragFileSrcInfo = DragFileSrcInfo{clientInfo.Index, clientInfo.ClientId}
				return
			}
		}
	}

	log.Printf("[%s][%s] Error: not found valid client (source,port):(%d,%d)",
		tag, rtkMisc.GetFuncInfo(), source, port)
}

func (mgr *InterfaceMgr) DragFileEnd(source, port int) {
	clientInfoTbList, err := rtkdbManager.QueryClientInfoBySrcPort(source, port)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: get client by (source,port):(%d,%d) failed: %d", tag, rtkMisc.GetFuncInfo(), source, port, err)
		return
	}

	for _, clientInfo := range clientInfoTbList {
		if (clientInfo.Online == true) && (clientInfo.AuthStatus == true) {
			rtkClientManager.SendDragFileEvent(mgr.mDragFileSrcInfo.id, clientInfo.ClientId, uint32(mgr.mDragFileSrcInfo.index))
			return
		}
	}

	log.Printf("[%s][%s] Error: not found valid client (source,port):(%d,%d)",
		tag, rtkMisc.GetFuncInfo(), source, port)
}

// Deprecated: Unused
func (mgr *InterfaceMgr) GetDiasId() string {
	return rtkGlobal.ServerMdnsId
}

// Deprecated: Use GetClientInfodData
func (mgr *InterfaceMgr) GetDeviceName(source, port int) string {
	clientInfoTbList, err := rtkdbManager.QueryClientInfoBySrcPort(source, port)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: query device name failed: %d", tag, rtkMisc.GetFuncInfo(), err)
		return ""
	}

	for _, clientInfo := range clientInfoTbList {
		if (clientInfo.Online == true) && (clientInfo.AuthStatus == true) {
			return clientInfo.DeviceName
		}
	}

	return ""
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
	clientInfoTbList, err := rtkdbManager.QueryClientInfoBySrcPort(source, port)
	if err != rtkMisc.SUCCESS && err != rtkMisc.ERR_DB_SQLITE_EMPTY_RESULT {
		log.Printf("[%s][%s] Error: query clientInfo data failed: %d", tag, rtkMisc.GetFuncInfo(), err)
		return rtkCommon.ClientInfoTb{}
	}

	for _, clientInfo := range clientInfoTbList {
		if (clientInfo.Online == true) && (clientInfo.AuthStatus == true) {
			return clientInfo
		}
	}

	return rtkCommon.ClientInfoTb{Source: source, Port: port}
}

func (mgr *InterfaceMgr) UpdateMonitorName(name string) {
	log.Printf("[%s][%s] MonitorName: %s", tag, rtkMisc.GetFuncInfo(), name)
	rtkGlobal.ServerMonitorName = name
}

func (mgr *InterfaceMgr) UpdateProductName(name string) {
	log.Printf("[%s][%s] ProductName: %s", tag, rtkMisc.GetFuncInfo(), name)
	rtkGlobal.ServerProductName = name
}

func (mgr *InterfaceMgr) UpdateMacAddr(macAddr string) {
	log.Printf("[%s][%s] MacAddr: %s", tag, rtkMisc.GetFuncInfo(), macAddr)
	rtkGlobal.ServerMdnsId = macAddr
}

func (mgr *InterfaceMgr) UpdateSrcPortTiming(source, port, width, height, framerate int) {
	log.Printf("[%s][%s] (Source,Port): (%d,%d), timing: %dx%d@%d", tag, rtkMisc.GetFuncInfo(), source, port, width, height, framerate)
	err := rtkClientManager.UpdateSrcPortTiming(source, port, width, height, framerate)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s][%s] Error: update source port timing failed: %s", tag, rtkMisc.GetFuncInfo(), rtkMisc.GetResponse(err).Msg)
	}
}

func (mgr *InterfaceMgr) EnableCrossShare(enable bool) {
	log.Printf("[%s][%s] Enable: %t", tag, rtkMisc.GetFuncInfo(), enable)
	// TODO: implement
}

