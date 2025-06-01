package misc

import (
	"log"
	"net"
)

func IsNetworkConnected() bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("[%s]Failed to get network interfaces: %v", GetFuncInfo(), err)
		return false
	}

	//log.Printf("interfaces len %d", len(interfaces))

	for _, iface := range interfaces {
		//log.Printf("[%s]  MTU[%d]  flag[%d]  HW[%+v] Index[%d]", iface.Name, iface.MTU, iface.Flags, iface.HardwareAddr, iface.Index)
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Err: Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}

		//TODO: Multiple network cards this check is invalid, refine this flow later
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil || ipNet.IP.To16() != nil {
					//log.Printf("ipNet.IP:[%s] ", ipNet.IP.String())
					return true
				}
			}
		}

	}

	return false
}

func GetNetworkIP() ([]string, bool) {
	ipStrList := make([]string, 0)
	ipStrList = append(ipStrList, DefaultIp)

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
