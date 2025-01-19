package connection

import (
	"fmt"
	ma "github.com/multiformats/go-multiaddr"
	"log"
	"net"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
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

func SetNetworkSwitchFlag() {
	networkSwitchSignalChan <- struct{}{}
}

func GetNetworkSwitchFlag() <-chan struct{} {
	return networkSwitchSignalChan
}
