package connection

import (
	"context"
	"fmt"
	ma "github.com/multiformats/go-multiaddr"
	"log"
	"net"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"time"
)

var networkSwitchSignalChan = make(chan struct{})

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

func IsNetworkConnected() bool {
	if rtkPlatform.GetPlatform() == rtkGlobal.PlatformAndroid { // android network status must get from java
		return rtkPlatform.GetNetWorkConnected()
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("[%s]Failed to get network interfaces: %v", rtkUtils.GetFuncInfo(), err)
		return false
	}

	for _, iface := range interfaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Err: Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}

		// TODO: Multiple network cards this check is invalid, refine this flow later
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil || ipNet.IP.To16() != nil {
					return true
				}
			}
		}

	}

	return false
}

func GetNetworkStatus() chan bool {
	return mdnsNoticeNetworkStatus
}

func WatchNetworkConnected(ctx context.Context) {
	lastStatus := true

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
			if !IsNetworkConnected() {
				if lastStatus {
					mdnsNoticeNetworkStatus <- false
					CancelAllProcess <- struct{}{}
					CancelStreamPool()
					lastStatus = false
				}
				log.Printf("[%s] Network is unavailable!", rtkUtils.GetFuncInfo())
			} else {
				if !lastStatus {
					log.Printf("[%s] Network is reconnected!", rtkUtils.GetFuncInfo())
					lastStatus = true
					mdnsNoticeNetworkStatus <- true
				}
			}

		}
	}

}

func WatchNetworkInfo(ctx context.Context) {
	lastIp := rtkGlobal.NodeInfo.IPAddr.PublicIP
	lastPort := rtkGlobal.NodeInfo.IPAddr.LocalPort

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			currentIpList, ok := GetNetworkIP()
			if !ok {
				log.Printf("[%s] GetNetworkIP error!", rtkUtils.GetFuncInfo())
				continue
			}
			currentPort := GetListenPort()
			if (!rtkUtils.IsInTheList(lastIp, currentIpList) || currentPort != lastPort) && IsNetworkConnected() {
				log.Printf("[%s] NetworkInfo is change, new addr:[%+v]!", rtkUtils.GetFuncInfo(), currentIpList)
				log.Println("**************** Attention please, the host listen addr is switch! ********************\n\n")

				rtkGlobal.ListenPort = 0
				rtkGlobal.ListenHost = rtkGlobal.DefaultIp //  libp2p will Assign a new IP address
				SetNetworkSwitchFlag()
			}
		}
	}
}

func GetNetworkIP() ([]string, bool) {
	ipStrList := make([]string, 0)
	ipStrList = append(ipStrList, rtkGlobal.DefaultIp)

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Failed to get network interfaces: %v", err)
		return ipStrList, false
	}
	bCheckOk := false
	for _, iface := range interfaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Err: Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipStrList = append(ipStrList, ipNet.IP.String())
					bCheckOk = true
				}
			}
		}
	}

	return ipStrList, bCheckOk
}

func GetListenPort() string {
	return rtkUtils.GetLocalPort(node.Network().ListenAddresses())
}

func SetNetworkSwitchFlag() {
	networkSwitchSignalChan <- struct{}{}
}

func GetNetworkSwitchFlag() <-chan struct{} {
	return networkSwitchSignalChan
}
