package login

import (
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"sync/atomic"
	"time"
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
)

var (
	serverInstanceMap sync.Map //KEY: instance
	cancelBrowse      func()
	lanServerAddr     string
	lanServerName     string
	pSafeConnect      *safeConnect

	heartBeatTicker *time.Ticker
	heartBeatFlag   = make(chan struct{}, 1)
	//stopLanServerRunFlag = make(chan struct{}, 1)
	isReconnectRunning  atomic.Bool
	reconnectCancelFunc func()

	// Used by connection package
	GetClientListFlag       = make(chan []rtkMisc.ClientInfo)
	DisconnectLanServerFlag = make(chan struct{})
	
	// Used by cmd package
	CurrentDiasStatus CrossShareDiasStatus
)
