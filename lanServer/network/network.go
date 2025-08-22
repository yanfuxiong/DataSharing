package network

import (
	"context"
	"log"
	rtkBuildConfig "rtk-cross-share/lanServer/buildConfig"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"

	"time"
)

var (
	SupInterfaces = []string{"wlan0", "eth0"} // e.g., "en0", "wlan0", "lo0", "eth0.100"
)

// var availableChanFlag = make(chan bool)
var networkSwitchSignalChan = make(chan struct{})

func WatchNetworkConnected(ctx context.Context) {
	lastStatus := true
	// TODO: measure the polling time (5s) is suitable or not
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !rtkMisc.IsNetworkConnected(SupInterfaces) {
				if lastStatus {
					lastStatus = false
					//availableChanFlag <- false
				}
				log.Printf("[%s] Network is unavailable!", rtkMisc.GetFuncInfo())
			} else {
				if !lastStatus {
					lastStatus = true
					//availableChanFlag <- true
					log.Printf("[%s] Network is reconnected!  try to login lan server...", rtkMisc.GetFuncInfo())
				}
			}

		}
	}

}

// Depreacted: Use listener.Accept error as network changed
func WatchNetworkInfo(ctx context.Context) {
	var lastIp string
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	printNetError := true
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentIpList, ok := rtkMisc.GetNetworkIP(SupInterfaces)
			if !ok {
				if printNetError {
					log.Printf("[%s] GetNetworkIP error!", rtkMisc.GetFuncInfo())
					printNetError = false
				}
				continue
			}

			lastIp = rtkGlobal.ServerIPAddr
			printNetError = true
			// FIXME: it will trigger few times even it already got new IP
			if lastIp != "" && !rtkMisc.IsInTheList(lastIp, currentIpList) && rtkMisc.IsNetworkConnected(SupInterfaces) {
				log.Println("==============================================================================")
				log.Printf("%s Network  old IP:[%s] new IP:[%+v]!", rtkBuildConfig.ServerName, lastIp, currentIpList)
				log.Printf("******** %s Network is Switch, cancel old business! ******** ", rtkBuildConfig.ServerName)
				// networkSwitchSignalChan <- struct{}{}
			}
		}
	}
}

// Deprecated: Unused
func GetNetworkSwitchFlag() <-chan struct{} {
	return networkSwitchSignalChan
}
