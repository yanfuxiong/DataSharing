package login

import (
	"context"
	"fmt"
	"github.com/grandcat/zeroconf"
	"log"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"strconv"
	"strings"
	"time"
)

func BrowseInstance() rtkMisc.CrossShareErr {
	serverInstanceMap.Clear()
	if cancelBrowse != nil {
		cancelBrowse()
		cancelBrowse = nil
		time.Sleep(50 * time.Millisecond) // Delay 50ms between "stop browse server" and "start lookup server"
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancelBrowse = cancel

	resultChan := make(chan browseParam)

	var err rtkMisc.CrossShareErr
	if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformiOS {
		err = browseLanServeriOS(ctx, rtkMisc.LanServiceType, resultChan)
	} else if rtkGlobal.NodeInfo.Platform == rtkMisc.PlatformAndroid {
		err = browseLanServerAndroid(ctx, rtkMisc.LanServiceType, rtkMisc.LanServerDomain, resultChan)
	} else {
		err = browseLanServer(ctx, rtkMisc.LanServiceType, rtkMisc.LanServerDomain, resultChan)
	}

	rtkMisc.GoSafe(func() {
		for param := range resultChan {
			if len(param.instance) > 0 && len(param.ip) > 0 {
				serverInstanceMap.Store(param.instance, param)
			}
		}
	})
	return err
}

func stopBrowseInstance() {
	if cancelBrowse != nil {
		cancelBrowse()
		cancelBrowse = nil
	}
}

func browseLanServer(ctx context.Context, serviceType, domain string, resultChan chan<- browseParam) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(rtkUtils.GetNetInterfaces(), nil)
	if err != nil {
		log.Printf("[%s] Failed to initialize resolver:%+v", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_NETWORK_C2S_RESOLVER
	}

	getTextRecordMap := func(textRecord []string) map[string]string {
		txtMap := make(map[string]string)
		for _, txt := range textRecord {
			parts := strings.SplitN(txt, "=", 2)
			if len(parts) == 2 {
				txtMap[parts[0]] = parts[1]
			}
		}
		return txtMap
	}
	entries := make(chan *zeroconf.ServiceEntry)
	rtkMisc.GoSafe(func() {
		for entry := range entries {
			if len(entry.AddrIPv4) > 0 {
				lanServerIp := fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
				txtMap := getTextRecordMap(entry.Text)
				textRecordmonitorName := txtMap[rtkMisc.TextRecordKeyMonitorName]
				textRecordTimeStamp := txtMap[rtkMisc.TextRecordKeyTimestamp]
				textRecordKeyVersion := txtMap[rtkMisc.TextRecordKeyVersion]
				log.Printf("Browse get a Service, mName:[%s] instance:[%s] IP:[%s] ver:[%s] timestamp:[%s], use %d ms", textRecordmonitorName, entry.Instance, lanServerIp, textRecordKeyVersion, textRecordTimeStamp, time.Now().UnixMilli()-startTime)

				resultChan <- browseParam{entry.Instance, lanServerIp, textRecordmonitorName, textRecordKeyVersion, 0}
			}
		}
		log.Printf("Stop Browse service instances...")
		close(resultChan)
	})

	err = resolver.Browse(ctx, serviceType, domain, entries)
	if err != nil {
		log.Printf("[%s] Failed to browse:%+v", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_NETWORK_C2S_BROWSER
	}

	log.Printf("Start Browse service instances...")
	return rtkMisc.SUCCESS
}

func browseLanServerAndroid(ctx context.Context, serviceType, domain string, resultChan chan<- browseParam) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(rtkUtils.GetNetInterfaces(), nil)
	if err != nil {
		log.Printf("[%s] Failed to initialize resolver:%+v", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_NETWORK_C2S_RESOLVER
	}

	entries := make(chan *zeroconf.ServiceEntry)

	getTextRecordMap := func(textRecord []string) map[string]string {
		txtMap := make(map[string]string)
		for _, txt := range textRecord {
			parts := strings.SplitN(txt, "=", 2)
			if len(parts) == 2 {
				txtMap[parts[0]] = parts[1]
			}
		}
		return txtMap
	}
	rtkMisc.GoSafe(func() {
		for entry := range entries {
			if len(entry.AddrIPv4) > 0 {
				entryIp := entry.AddrIPv4[0].String()
				lanServerIp := fmt.Sprintf("%s:%d", entryIp, entry.Port)
				log.Printf("Browse get a Service:[%s] IP:[%s], use [%d] ms", entry.Instance, lanServerIp, time.Now().UnixMilli()-startTime)

				txtMap := getTextRecordMap(entry.Text)
				textRecordIp := txtMap[rtkMisc.TextRecordKeyIp]
				textRecordProductName := txtMap[rtkMisc.TextRecordKeyProductName]
				textRecordmName := txtMap[rtkMisc.TextRecordKeyMonitorName]
				textRecordTimeStamp := txtMap[rtkMisc.TextRecordKeyTimestamp]
				textRecordKeyVersion := txtMap[rtkMisc.TextRecordKeyVersion]
				if textRecordIp != entryIp {
					log.Printf("[%s] WARNING: Different IP. Entry:(%s); TextRecord:(%s)", rtkMisc.GetFuncInfo(), entryIp, textRecordIp)
					continue
				}

				if (textRecordProductName != "") && (g_ProductName != "") {
					if textRecordProductName != g_ProductName {
						log.Printf("[%s] WARNING: Different ProductName. Mobile:(%s); TextRecord:(%s)", rtkMisc.GetFuncInfo(), g_ProductName, textRecordProductName)
						continue
					}
				}
				stamp, err := strconv.Atoi(textRecordTimeStamp)
				if err != nil {
					log.Printf("[%s] WARNING: invalid timestamp:%s, err:%+v", rtkMisc.GetFuncInfo(), rtkMisc.TextRecordKeyTimestamp, err)
				}

				log.Printf("Found target Service, mName:[%s] instance:[%s] IP:[%s] ver:[%s] timestamp:[%s], use %d ms", textRecordmName, entry.Instance, lanServerIp, textRecordKeyVersion, textRecordTimeStamp, time.Now().UnixMilli()-startTime)

				if lanServerRunning.Load() {
					rtkPlatform.GoNotifyBrowseResult(textRecordmName, entry.Instance, lanServerIp, textRecordKeyVersion, int64(stamp))
				}

				resultChan <- browseParam{entry.Instance, lanServerIp, textRecordmName, textRecordKeyVersion, int64(stamp)}
			}
		}
		log.Printf("Stop Browse service instances")
		close(resultChan)
	})

	err = resolver.Browse(ctx, serviceType, domain, entries)
	if err != nil {
		log.Printf("[%s] Failed to browse:%+v", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_NETWORK_C2S_BROWSER
	}

	log.Printf("Start Browse service instances...")
	return rtkMisc.SUCCESS
}

func browseLanServeriOS(ctx context.Context, serviceType string, resultChan chan<- browseParam) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	log.Printf("Start Browse service instances...")
	rtkPlatform.SetGoBrowseMdnsResultCallback(func(instance, ip string, port int, productName, mName, timestamp, version string) {
		lanServerIp := fmt.Sprintf("%s:%d", ip, port)
		log.Printf("Browse get a Service:[%s] IP:[%s],use [%d] ms", instance, lanServerIp, time.Now().UnixMilli()-startTime)

		stamp, err := strconv.Atoi(timestamp)
		if err != nil {
			log.Printf("[%s] WARNING: invalid[%s]:%d. err:%s", rtkMisc.GetFuncInfo(), rtkMisc.TextRecordKeyTimestamp, stamp, err)
		}

		if lanServerRunning.Load() {
			rtkPlatform.GoNotifyBrowseResult(mName, instance, lanServerIp, version, int64(stamp))
		}
		resultChan <- browseParam{instance, lanServerIp, mName, version, int64(stamp)}
	})
	rtkPlatform.GoStartBrowseMdns("", serviceType)

	rtkMisc.GoSafe(func() {
		<-ctx.Done()
		log.Printf("Stop Browse service instances...")
		rtkPlatform.GoStopBrowseMdns()
		rtkPlatform.SetGoBrowseMdnsResultCallback(nil)
		close(resultChan)
	})

	return rtkMisc.SUCCESS
}

func lookupLanServer(ctx context.Context, instance, serviceType, domain string, bPrintErr bool) (string, rtkMisc.CrossShareErr) {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(rtkUtils.GetNetInterfaces(), nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return "", rtkMisc.ERR_NETWORK_C2S_RESOLVER
	}

	lanServerEntry := make(chan *zeroconf.ServiceEntry)
	if bPrintErr {
		log.Printf("Start Lookup service  by instance:%s  type:%s domain:%s", instance, serviceType, domain)
	}

	err = resolver.Lookup(ctx, instance, serviceType, domain, lanServerEntry, g_lookupByUnicast)
	g_lookupByUnicast = !g_lookupByUnicast // Retry with Unicast package every 2 times
	if err != nil {
		log.Println("Failed to Lookup:", err.Error())
		return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP
	}

	getTextRecordMap := func(textRecord []string) map[string]string {
		txtMap := make(map[string]string)
		for _, txt := range textRecord {
			parts := strings.SplitN(txt, "=", 2)
			if len(parts) == 2 {
				txtMap[parts[0]] = parts[1]
			}
		}
		return txtMap
	}
	select {
	case <-ctx.Done():
		return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP_TIMEOUT
	case entry, ok := <-lanServerEntry:
		if !ok {
			break
		}
		txtMap := getTextRecordMap(entry.Text)
		textRecordmonitorName := txtMap[rtkMisc.TextRecordKeyMonitorName]
		textRecordTimeStamp := txtMap[rtkMisc.TextRecordKeyTimestamp]
		textRecordKeyVersion := txtMap[rtkMisc.TextRecordKeyVersion]

		if entry.Instance != instance {
			log.Printf("Expect instance[%s], ignore instance [%s]", instance, entry.Instance)
		} else if len(entry.AddrIPv4) > 0 {
			lanServerIp := fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
			log.Printf("Lookup get Service, mName:[%s] instance:[%s] IP:[%s] ver:[%s] timestamp:[%s], use %d ms", textRecordmonitorName, entry.Instance, lanServerIp, textRecordKeyVersion, textRecordTimeStamp, time.Now().UnixMilli()-startTime)
			stamp, err := strconv.Atoi(textRecordTimeStamp)
			if err != nil {
				log.Printf("[%s] WARNING: invalid[%s]:%d. err:%s", rtkMisc.GetFuncInfo(), rtkMisc.TextRecordKeyTimestamp, stamp, err)
			}
			param := browseParam{entry.Instance, lanServerIp, textRecordmonitorName, textRecordKeyVersion, int64(stamp)}
			g_monitorName = textRecordmonitorName
			serverInstanceMap.Store(param.instance, param)
			return lanServerIp, rtkMisc.SUCCESS
		} else {
			log.Printf("ServiceInstanceName [%s] get AddrIPv4 is null", entry.ServiceInstanceName())
		}
	}
	return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP_INVALID
}

func lookupLanServeriOS(ctx context.Context, instance, serviceType string) (string, rtkMisc.CrossShareErr) {
	startTime := time.Now().UnixMilli()
	log.Printf("Start Lookup service  by name:%s  type:%s", instance, serviceType)
	lanServerEntry := make(chan browseParam)
	rtkPlatform.SetGoBrowseMdnsResultCallback(func(instance, ip string, port int, productName, mName, timestamp, version string) {
		lanServerIp := fmt.Sprintf("%s:%d", ip, port)
		stamp, err := strconv.Atoi(timestamp)
		if err != nil {
			log.Printf("[%s] WARNING: invalid[%s]:%d. err:%s", rtkMisc.GetFuncInfo(), rtkMisc.TextRecordKeyTimestamp, stamp, err)
		}
		lanServerEntry <- browseParam{instance, lanServerIp, mName, version, int64(stamp)}
	})
	rtkPlatform.GoStartBrowseMdns(instance, serviceType)

	select {
	case <-ctx.Done():
		log.Printf("Lookup Timeout, get no entries")
		rtkPlatform.GoStopBrowseMdns()
		rtkPlatform.SetGoBrowseMdnsResultCallback(nil)
		return "", rtkMisc.ERR_NETWORK_C2S_LOOKUP_TIMEOUT
	case val := <-lanServerEntry:
		log.Printf("Lookup get Service success, use [%d] ms", time.Now().UnixMilli()-startTime)
		return val.ip, rtkMisc.SUCCESS
	}
}
