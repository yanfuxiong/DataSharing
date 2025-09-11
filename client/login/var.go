package login

import (
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"sync/atomic"
	"time"
)

const (
	g_retryServerMaxCnt   = 2
	g_retryServerInterval = 200 * time.Millisecond
)

type CrossShareDiasStatus int

const (
	DIAS_Status_Wait_DiasMonitor CrossShareDiasStatus = iota + 1
	DIAS_Status_Connectting_DiasService
	DIAS_Status_Checking_Authorization //Windows
	DIAS_Status_Wait_screenCasting     // Android
	DIAS_Status_Authorization_Failed
	DIAS_Status_Wait_Other_Clients
	DIAS_Status_Get_Clients_Success
	DIAS_Status_Connected_DiasService_Failed
)

type callbackDisconnectAllClientFunc func()
type callbackCancelAllBusinessFunc func()

type browseParam struct {
	instance    string
	ip          string
	monitorName string
	ver         string
	timeStamp   int64
}

var (
	serverInstanceMap       sync.Map //KEY: instance
	cancelBrowse            func()
	lanServerAddr           string
	lanServerInstance       string
	g_ProductName           string
	g_monitorName           string
	pSafeConnect            *safeConnect
	heartBeatTicker         *time.Ticker
	isReconnectRunning      atomic.Bool
	lanServerRunning        atomic.Bool
	reconnectCancelFunc     func()
	disconnectAllClientFunc callbackDisconnectAllClientFunc
	cancelAllBusinessFunc   callbackCancelAllBusinessFunc
	mobileAuthData          rtkMisc.AuthDataInfo
	g_lookupByUnicast       bool
	initLanServerMutex      sync.Mutex

	// Used by connection package
	GetClientListFlag = make(chan []rtkMisc.ClientInfo)

	currentDiasStatus CrossShareDiasStatus
)

func SetDisconnectAllClientCallback(cb callbackDisconnectAllClientFunc) {
	disconnectAllClientFunc = cb
}

func SetCancelAllBusinessCallback(cb callbackCancelAllBusinessFunc) {
	cancelAllBusinessFunc = cb
}
