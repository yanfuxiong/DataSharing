package clientManager

import (
	"context"
	"fmt"
	"log"
	rtkCommon "rtk-cross-share/lanServer/common"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"
	"strconv"
	"sync"
	"time"
)

type SrcPortContextManager struct {
	mu     sync.Mutex
	ctxMap map[rtkMisc.SourcePort]context.CancelFunc
}

func NewSrcPortContextManager() *SrcPortContextManager {
	return &SrcPortContextManager{
		ctxMap: make(map[rtkMisc.SourcePort]context.CancelFunc),
	}
}

func (m *SrcPortContextManager) Build(parentCtx context.Context, srcPort rtkMisc.SourcePort) context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, ok := m.ctxMap[srcPort]; ok {
		cancel()
	}

	ctx, cancel := context.WithCancel(parentCtx)
	m.ctxMap[srcPort] = cancel

	return ctx
}

func (m *SrcPortContextManager) Cancel(srcPort rtkMisc.SourcePort) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, ok := m.ctxMap[srcPort]; ok {
		cancel()
		delete(m.ctxMap, srcPort)
	}
}

type SignalCheckingType int

const (
	SIG_CHK_OFFLINE_NOTI   SignalCheckingType = 0
	SIG_CHK_NO_CLIENT      SignalCheckingType = 1
	SIG_CHK_AUTHORIZED     SignalCheckingType = 2
	SIG_CHK_OVER_AUTH_TIME SignalCheckingType = 3
	SIG_CHK_OVER_SIG_TIME  SignalCheckingType = 4

	kMaxAuthTimeCnt   = 5
	kMaxSignalTimeCnt = 5

	kDefaultAndroidLink = "https://play.google.com/store/apps/details?id=com.realtek.crossshare"
	kDefaultiOSLink     = "https://apps.apple.com/app/crossshare/id6747169753"
)

var (
	ctxMgr            = NewSrcPortContextManager()
	chanSrcPortTiming = make(chan rtkCommon.SrcPortTiming, 5)
)

// =============================
// TimingData by source and port get event
// =============================
type NotifyGetTimingDataBySrcPortCallback func(source, port int) rtkCommon.TimingData

var notifyGetTimingDataBySrcPortCallback NotifyGetTimingDataBySrcPortCallback

func SetNotifyGetTimingDataBySrcPortCallback(cb NotifyGetTimingDataBySrcPortCallback) {
	notifyGetTimingDataBySrcPortCallback = cb
}

// call by InterfaceMgr
func UpdateSrcPortTiming(source, port, width, height, framerate int) rtkMisc.CrossShareErr {
	dbErr := rtkdbManager.UpsertTimingInfo(source, port, width, height, framerate)
	if dbErr != rtkMisc.SUCCESS {
		return dbErr
	}

	chanSrcPortTiming <- rtkCommon.SrcPortTiming{Source: source, Port: port, Width: width, Height: height, Framerate: framerate}

	return rtkMisc.SUCCESS
}

func HandleClientSignalChecking(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Cancel by context!", rtkMisc.GetFuncInfo())
			return
		case srcPortTiming := <-chanSrcPortTiming:
			subCtx := ctxMgr.Build(ctx, rtkMisc.SourcePort{Source: srcPortTiming.Source, Port: srcPortTiming.Port})

			if srcPortTiming.IsSingal() == false {
				break
			}

			rtkMisc.GoSafe(func() { handleClientSignalCheckingInternal(subCtx, srcPortTiming) })
		}
	}
}

func handleOfflineClientSignalChecking(clientIndex int) {
	clientInfo, err := rtkdbManager.QueryClientInfoByIndex(clientIndex)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s] Err(%d): Get Client info failed:(%d)", rtkMisc.GetFuncInfo(), err, clientIndex)
		return
	}

	timingData := notifyGetTimingDataBySrcPortCallback(clientInfo.Source, clientInfo.Port)
	chanSrcPortTiming <- rtkCommon.SrcPortTiming{Source: timingData.Source, Port: timingData.Port, Width: timingData.Width, Height: timingData.Height, Framerate: timingData.Framerate}
}

func handleClientSignalCheckingInternal(ctx context.Context, srcPortTiming rtkCommon.SrcPortTiming) {
	signalTime := time.Now()
	retryAuthTimeCnt := 0
	retrySignalTimeCnt := 0

	for {
		retClientInfoTb, retType := checkClientOfflineCond(srcPortTiming.Source, srcPortTiming.Port, signalTime)

		switch retType {
		case SIG_CHK_NO_CLIENT:
			return
		case SIG_CHK_AUTHORIZED:
			return
		case SIG_CHK_OVER_AUTH_TIME:
			retryAuthTimeCnt++
			log.Printf("[%s] Signal checking retry AuthTimeCnt:(%d)", rtkMisc.GetFuncInfo(), retryAuthTimeCnt)
		case SIG_CHK_OVER_SIG_TIME:
			retrySignalTimeCnt++
			log.Printf("[%s] Signal checking retry SignalTimeCnt:(%d)", rtkMisc.GetFuncInfo(), retrySignalTimeCnt)
		case SIG_CHK_OFFLINE_NOTI:
			androidLink, iosLink := getAppLink(retClientInfoTb)
			sendPlatformMsgEventCallback(
				int(rtkGlobal.CLIENT_EVENT_MSG_OPEN_GUIDE),
				strconv.Itoa(srcPortTiming.Source),
				strconv.Itoa(srcPortTiming.Port),
				androidLink,
				iosLink,
			)
			return
		}

		select {
		case <-ctx.Done():
			log.Printf("[%s] Cancel by context!", rtkMisc.GetFuncInfo())
			return
		case <-time.After(time.Second):
		}
	}
}

func checkClientOfflineCond(source, port int, signalTime time.Time) (rtkCommon.ClientInfoTb, SignalCheckingType) {
	// Get client by source and port
	clientInfoTbList, err := rtkdbManager.QueryClientInfoBySrcPort(source, port)
	if err != rtkMisc.SUCCESS {
		log.Printf("[%s] Ignore: detected timing but no client with (source,port)=(%d,%d)",
			rtkMisc.GetFuncInfo(), source, port)
		return rtkCommon.ClientInfoTb{}, SIG_CHK_NO_CLIENT
	}

	convertStr2Time := func(dateStr string) (time.Time, error) {
		if dateStr == "" {
			return time.Time{}, fmt.Errorf("addr is null!")
		}
		layout := "2006-01-02 15:04:05" // SQLite DATETIME format
		t, err := time.ParseInLocation(layout, dateStr, time.Local)
		return t, err
	}

	// Check client if online and authorized
	var latestTime time.Time
	var retClientInfoTb rtkCommon.ClientInfoTb
	for _, clientInfo := range clientInfoTbList {
		if (clientInfo.Online == true) && (clientInfo.AuthStatus == true) {
			log.Printf("[%s] Ignore: there is online client with (source,port)=(%d,%d)",
				rtkMisc.GetFuncInfo(), source, port)
			return clientInfo, SIG_CHK_AUTHORIZED
		}

		t, err := convertStr2Time(clientInfo.LastAuthTime)
		if err != nil {
			log.Printf("[%s] (%s) convert to time failed: %s", rtkMisc.GetFuncInfo(), clientInfo.LastAuthTime, err.Error())
			continue
		}

		if t.After(latestTime) {
			latestTime = t
			retClientInfoTb = clientInfo
		}
	}

	if time.Since(latestTime) < (kMaxAuthTimeCnt * time.Second) {
		return retClientInfoTb, SIG_CHK_OVER_AUTH_TIME
	}

	if time.Since(signalTime) < (kMaxSignalTimeCnt * time.Second) {
		return retClientInfoTb, SIG_CHK_OVER_SIG_TIME
	}

	return retClientInfoTb, SIG_CHK_OFFLINE_NOTI
}

func getAppLink(clientInfo rtkCommon.ClientInfoTb) (string, string) {
	getAppLinkByIdx := func(index int, platform string) string {
		link, err := rtkdbManager.QueryLinkInfoByIndex(index)
		if err != rtkMisc.SUCCESS || link == "" {
			return getPlatformAppLink(platform)
		}
		return link
	}

	var androidLink string
	var iosLink string
	if clientInfo.Platform == rtkMisc.PlatformAndroid {
		androidLink = getAppLinkByIdx(clientInfo.Index, clientInfo.Platform)
		iosLink = getPlatformAppLink(rtkMisc.PlatformiOS)
	} else if clientInfo.Platform == rtkMisc.PlatformiOS {
		androidLink = getPlatformAppLink(rtkMisc.PlatformAndroid)
		iosLink = getAppLinkByIdx(clientInfo.Index, clientInfo.Platform)
	} else {
		androidLink = getPlatformAppLink(rtkMisc.PlatformAndroid)
		iosLink = getPlatformAppLink(rtkMisc.PlatformiOS)
	}

	return androidLink, iosLink
}

func getPlatformAppLink(platform string) string {
	link, err := rtkdbManager.QueryLinkInfoByPlatform(platform)
	if err != rtkMisc.SUCCESS {
		if platform == rtkMisc.PlatformAndroid {
			return kDefaultAndroidLink
		} else if platform == rtkMisc.PlatformiOS {
			return kDefaultiOSLink
		} else {
			log.Printf("[%s] Unsupport get app link by (%s)", rtkMisc.GetFuncInfo(), platform)
			return ""
		}
	}

	return link
}
