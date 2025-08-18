package connection

import (
	"context"
	"fmt"
	"log"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"time"

	ma "github.com/multiformats/go-multiaddr"
)

func ListenMultAddr() []ma.Multiaddr {
	addrs := []string{
		"/ip4/%s/tcp/%d",
		//"/ip4/%s/udp/%d/quic",
	}

	for i, a := range addrs {
		addrs[i] = fmt.Sprintf(a, rtkGlobal.ListenHost, rtkGlobal.ListenPort)
	}

	multAddr := make([]ma.Multiaddr, 0)
	for _, addrstr := range addrs {
		a, err := ma.NewMultiaddr(addrstr)
		if err != nil {
			continue
		}
		multAddr = append(multAddr, a)
	}

	return multAddr
}

func GetNetworkStatus() chan bool {
	return mdnsNoticeNetworkStatus
}

func WatchNetworkInfo(ctx context.Context) {
	lastIp := rtkGlobal.NodeInfo.IPAddr.PublicIP
	lastPort := rtkGlobal.NodeInfo.IPAddr.LocalPort
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	printNetError := true
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentIpList, ok := rtkMisc.GetNetworkIP([]string{})
			if !ok {
				if printNetError {
					log.Printf("[%s] GetNetworkIP error!", rtkMisc.GetFuncInfo())
					printNetError = false
				}
				continue
			}
			currentPort := GetListenPort()
			printNetError = true
			if (!rtkMisc.IsInTheList(lastIp, currentIpList) || currentPort != lastPort) && rtkMisc.IsNetworkConnected([]string{}) {
				log.Printf("[%s] NetworkInfo is change, new addr:[%+v]!", rtkMisc.GetFuncInfo(), currentIpList)
				log.Println("**************** Attention please, the host listen addr is switch! ********************\n\n")

				rtkGlobal.ListenPort = rtkGlobal.DefaultPort
				rtkGlobal.ListenHost = rtkMisc.DefaultIp //  libp2p will Assign a new IP address
				rtkPlatform.GoTriggerNetworkSwitch()
			}
		}
	}
}

func GetListenPort() string {
	nodeMutex.RLock()
	defer nodeMutex.RUnlock()
	if node != nil {
		return rtkUtils.GetLocalPort(node.Network().ListenAddresses())
	}
	return rtkGlobal.NodeInfo.IPAddr.LocalPort
}
