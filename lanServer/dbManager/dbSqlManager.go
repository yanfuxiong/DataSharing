package dbManager

import (
	"log"
	rtkCommon "rtk-cross-share/lanServer/common"
	rtkMisc "rtk-cross-share/misc"
)

// =============================
// ClientInfo update event
// =============================
type NotifyUpdateClientInfoCallback func(rtkCommon.ClientInfoTb)

var notifyUpdateClientInfoCallback NotifyUpdateClientInfoCallback

func SetNotifyUpdateClientInfoCallback(cb NotifyUpdateClientInfoCallback) {
	notifyUpdateClientInfoCallback = cb
}

func notifyUpdateClientInfo(pkIndexList []int) {
	for _, pkIndex := range pkIndexList {
		clientInfo, err := QueryClientInfoByIndex(pkIndex)
		if err != rtkMisc.SUCCESS {
			continue
		}

		if clientInfo.Online {
			clientInfoList := make([]rtkMisc.SourcePortInfo, 0)
			err = querySrcPortInfo(pkIndex, &clientInfoList)
			if err != rtkMisc.SUCCESS {
				continue
			}

			for _, srcPortInfo := range clientInfoList {
				clientInfo.Source = srcPortInfo.Source
				clientInfo.Port = srcPortInfo.Port
				clientInfo.UdpMousePort = srcPortInfo.UdpMousePort
				clientInfo.UdpKeyboardPort = srcPortInfo.UdpKeyboardPort

				notifyUpdateClientInfoCallback(clientInfo)
			}

		} else { // TODO: check it
			clientInfo.Source = 0
			clientInfo.Port = 0
			clientInfo.UdpMousePort = 0
			clientInfo.UdpKeyboardPort = 0
			notifyUpdateClientInfoCallback(clientInfo)
		}
	}
}

// =============================
// Query
// =============================
func QueryOnlineClientList(clientInfoList *[]rtkCommon.ClientInfoTb) rtkMisc.CrossShareErr {
	err := queryClientInfo(
		clientInfoList,
		[]SqlCond{SqlCondOnline, SqlCondAuthStatusIsTrue},
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	log.Printf("QueryOnlineClientList get len:%d", len(*clientInfoList))
	return rtkMisc.SUCCESS
}

func ClientQueryOnlineClientList(pkIndex int, clientInfoList *[]rtkCommon.ClientInfoTb) rtkMisc.CrossShareErr {
	err := clientQueryClientList(
		pkIndex,
		clientInfoList,
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	log.Printf("ClientQueryOnlineClientList get len:%d", len(*clientInfoList)) //Including the client itself
	return rtkMisc.SUCCESS
}

func QueryClientInfoByIndex(pkIndex int) (rtkCommon.ClientInfoTb, rtkMisc.CrossShareErr) {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, QueryClientInfoByIndex failed! Invalid Client Index", pkIndex)
		return rtkCommon.ClientInfoTb{}, rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	err := queryClientInfo(&clientInfoList,
		[]SqlCond{SqlCondPkIndex},
		pkIndex,
	)
	if err != rtkMisc.SUCCESS {
		return rtkCommon.ClientInfoTb{}, err
	}

	if len(clientInfoList) == 0 {
		return rtkCommon.ClientInfoTb{}, rtkMisc.ERR_DB_SQLITE_EMPTY_RESULT
	}

	if len(clientInfoList) > 1 {
		log.Printf("[%s] WARNING! The result count from database more than 1", rtkMisc.GetFuncInfo())
	}

	return clientInfoList[0], rtkMisc.SUCCESS
}

func QueryClientInfoBySrcPort(source, port int) ([]rtkCommon.ClientInfoTb, rtkMisc.CrossShareErr) {
	if source < 0 || port < 0 {
		log.Printf("[%s] Invalid (source,port)=(%d,%d)", rtkMisc.GetFuncInfo(), source, port)
		return []rtkCommon.ClientInfoTb{}, rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	err := queryClientInfo(
		&clientInfoList,
		[]SqlCond{SqlCondSource, SqlCondPort},
		source, port,
	)
	if err != rtkMisc.SUCCESS {
		return []rtkCommon.ClientInfoTb{}, err
	}

	if len(clientInfoList) == 0 {
		return []rtkCommon.ClientInfoTb{}, rtkMisc.ERR_DB_SQLITE_EMPTY_RESULT
	}

	if len(clientInfoList) > 1 {
		log.Printf("[%s] WARNING! The result count from database more than 1", rtkMisc.GetFuncInfo())
	}

	return clientInfoList, rtkMisc.SUCCESS
}

func QueryReconnList(clientInfoList *[]rtkCommon.ClientInfoTb) rtkMisc.CrossShareErr {
	err := queryClientInfo(
		clientInfoList,
		[]SqlCond{SqlCondOnline, SqlCondAuthStatusIsTrue, SqlCondGetClientListTrue, SqlCondLastUpdateTime},
		g_ReconnListInterval,
	)

	return err
}

func QueryMaxVersion() (string, rtkMisc.CrossShareErr) {
	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	err := QueryOnlineClientList(&clientInfoList)

	maxVer := string("")
	nMaxVerVal := int(0)
	for _, client := range clientInfoList {
		verVal := rtkMisc.GetVersionValue(client.Version)
		if nMaxVerVal < verVal {
			nMaxVerVal = verVal
			maxVer = rtkMisc.GetShortVersion(client.Version)
		}
	}

	return maxVer, err
}

func QueryLinkInfoByIndex(pkIndex int) (string, rtkMisc.CrossShareErr) {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, QueryLinkInfoByIndex failed! Invalid Client Index", pkIndex)
		return "", rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	return queryLinkInfo([]SqlCond{SqlCondClientIndex}, pkIndex)
}

func QueryLinkInfoByPlatform(platform string) (string, rtkMisc.CrossShareErr) {
	if platform == "" {
		log.Printf("platform is empty, QueryLinkInfoByPlatform failed! Invalid Client Index")
		return "", rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	return queryLinkInfo([]SqlCond{SqlCondPlatform, SqlCondLinkNotEmpty}, platform)
}

// =============================
// Update/Insert
// =============================
func UpsertClientInfo(clientId, host, ipAddr, deviceName, platform, version string) (int, rtkMisc.CrossShareErr) {
	pkIndex := 0
	err := upsertClientInfo(&pkIndex, clientId, host, ipAddr, deviceName, platform, version)
	if err != rtkMisc.SUCCESS {
		return 0, err
	}

	notifyUpdateClientInfo([]int{pkIndex})
	return pkIndex, err
}

func UpdateClientOffline(pkIndex int) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateClientOffline skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}

	pkIndexList := make([]int, 0)
	err := updateClientInfo(
		&pkIndexList,
		[]SqlCond{SqlCondOffline, SqlCondGetClientListFalse},
		[]SqlCond{SqlCondPkIndex},
		pkIndex,
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	authPkIndex := int(0)
	errUpsertAuthStatus := upsertAuthStatus(&authPkIndex, pkIndex, false)
	if errUpsertAuthStatus != rtkMisc.SUCCESS {
		return errUpsertAuthStatus
	}

	notifyUpdateClientInfo(pkIndexList)
	return rtkMisc.SUCCESS
}

func UpdateClientOnline(pkIndex int) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateClientOnline skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}

	pkIndexList := make([]int, 0)
	err := updateClientInfo(
		&pkIndexList,
		[]SqlCond{SqlCondOnline},
		[]SqlCond{SqlCondPkIndex},
		pkIndex,
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	notifyUpdateClientInfo(pkIndexList)
	return rtkMisc.SUCCESS
}

func UpdateAllClientOffline() rtkMisc.CrossShareErr {
	pkIndexList := make([]int, 0)
	err := updateClientInfo(
		&pkIndexList,
		[]SqlCond{SqlCondOffline, SqlCondGetClientListFalse},
		[]SqlCond{SqlCondOnline},
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	notifyUpdateClientInfo(pkIndexList)
	return rtkMisc.SUCCESS
}

func UpdateAuthAndSrcPort(pkIndex int, status bool, source, port int) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateClientSourceAndPort skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}
	if source < 0 || port < 0 {
		log.Printf("[%s] Invalid (source,port)=(%d,%d)", rtkMisc.GetFuncInfo(), source, port)
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	// Notify if source/port is changed
	preClientInfo, errQueryPre := QueryClientInfoByIndex(pkIndex)
	if errQueryPre == rtkMisc.SUCCESS {
		if preClientInfo.Source != source || preClientInfo.Port != port {
			emptyClientInfo := rtkCommon.ClientInfoTb{}
			emptyClientInfo.Source = preClientInfo.Source
			emptyClientInfo.Port = preClientInfo.Port
			notifyUpdateClientInfoCallback(emptyClientInfo)
		}
	}

	// Update AuthStatus
	authPkIndex := 0
	errUpsertAuthStatus := upsertAuthStatus(&authPkIndex, pkIndex, status)
	if errUpsertAuthStatus != rtkMisc.SUCCESS {
		return errUpsertAuthStatus
	}

	// Update Source, Port
	// Only update source and port if AuthStatus=1
	if status {
		pkIndexList := make([]int, 0)
		errUpdateSrcPort := updateClientInfo(
			&pkIndexList,
			[]SqlCond{SqlCondSource, SqlCondPort},
			[]SqlCond{SqlCondPkIndex},
			source, port, pkIndex,
		)
		if errUpdateSrcPort != rtkMisc.SUCCESS {
			return errUpdateSrcPort
		}
	}

	// Notify the lastest ClientInfoTb
	curClientInfo, errQueryCur := QueryClientInfoByIndex(pkIndex)
	if errQueryCur == rtkMisc.SUCCESS {
		notifyUpdateClientInfoCallback(curClientInfo)
	}

	return rtkMisc.SUCCESS
}

func UpdateSrcPortInfo(clientIndex, source, port, udpMousePort, udpKeyboardPort int) rtkMisc.CrossShareErr {
	errCode := upsertSrcPortInfo(source, port, clientIndex, udpMousePort, udpKeyboardPort)
	if errCode != rtkMisc.SUCCESS {
		return errCode
	}

	log.Printf("[%s] update srcPort success, source:%d, port:%d clientIndex:%d mousePort:%d kybrdPort:%d", rtkMisc.GetFuncInfo(), source, port, clientIndex, udpMousePort, udpKeyboardPort)

	clientInfo, err := QueryClientInfoByIndex(clientIndex)
	if err != rtkMisc.SUCCESS {
		return err
	}
	clientInfo.Source = source
	clientInfo.Port = port
	clientInfo.UdpMousePort = udpMousePort
	clientInfo.UdpKeyboardPort = udpKeyboardPort

	notifyUpdateClientInfoCallback(clientInfo)

	return rtkMisc.SUCCESS
}

func UpsertTimingInfo(source, port, width, height, framerate int) rtkMisc.CrossShareErr {
	if source < 0 || port < 0 {
		log.Printf("[%s] Invalid (source,port)=(%d,%d)", rtkMisc.GetFuncInfo(), source, port)
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	if width < 0 || height < 0 || framerate < 0 {
		log.Printf("[%s] Invalid timing=(%dx%d@%d)", rtkMisc.GetFuncInfo(), width, height, framerate)
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	return upsertTimingInfo(source, port, width, height, framerate)
}

func UpsertLinkInfo(pkIndex int, link string) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpsertLinkInfo failed! Invalid Client Index", pkIndex)
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	return upsertLinkInfo(pkIndex, link)
}
