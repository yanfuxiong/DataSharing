package login

import (
	"context"
	rtkCommon "rtk-cross-share/client/common"
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
	lanServerInstance       string
	g_ProductName           string
	g_monitorName           string
	pSafeConnect            *safeConnect
	lanServerRunning        atomic.Bool
	disconnectAllClientFunc callbackDisconnectAllClientFunc
	cancelAllBusinessFunc   callbackCancelAllBusinessFunc
	mobileAuthData          rtkMisc.AuthDataInfo
	g_lookupByUnicast       bool
	initLanServerMutex      sync.Mutex
	displayInfoList         []rtkCommon.DisplayEventInfo //current plug in display info list

	// connect reliability
	heartBeatTicker     *HeartBeatTicker
	pingServerMtx       sync.Mutex
	pingServerErrCnt    int
	pingServerTimeStamp int64

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

type HeartBeatTicker struct {
	interval time.Duration
	proxyCh  chan time.Time
	cancel   context.CancelFunc
	mu       sync.Mutex
}

func NewHeartBeatTicker(interval time.Duration) *HeartBeatTicker {
	return &HeartBeatTicker{
		interval: interval,
		proxyCh:  make(chan time.Time),
	}
}

func (t *HeartBeatTicker) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.stopTicker()

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	rtkMisc.GoSafe(func() {
		tk := time.NewTicker(t.interval)
		defer tk.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case tickTime := <-tk.C:
				t.proxyCh <- tickTime
			}
		}
	})
}

func (t *HeartBeatTicker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopTicker()
}

func (t *HeartBeatTicker) stopTicker() {
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
}

func (t *HeartBeatTicker) C() <-chan time.Time {
	return t.proxyCh
}
