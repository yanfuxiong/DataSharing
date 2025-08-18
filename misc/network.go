package misc

import (
	"fmt"
	"log"
	"net"
)

func GetValidAddrs(iface *net.Interface) ([]net.Addr, error) {
	if iface == nil {
		return nil, fmt.Errorf("[%s] Err: null interface", GetFuncInfo())
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	var retAddrs []net.Addr
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				retAddrs = append(retAddrs, addr)
			}
		}
	}

	if len(retAddrs) == 0 {
		return nil, fmt.Errorf("Err: Empty addrs (%s)", iface.Name)
	}

	return retAddrs, nil
}

func GetValidInterface(ifaceName string) (*net.Interface, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err == nil && iface != nil {
		return iface, nil
	}
	return nil, err
}

var printErrIface = true
var printErrIp = true

func IsNetworkConnected(forceInterfaces []string) bool {
	var interfaces []*net.Interface = make([]*net.Interface, 0)
	if len(forceInterfaces) > 0 {
		for _, ifaceName := range forceInterfaces {
			iface, err := GetValidInterface(ifaceName)
			if err != nil {
				continue
			}
			interfaces = append(interfaces, iface)
		}
	} else {
		tmpInterfaces, err := net.Interfaces()
		if err != nil {
			if printErrIface {
				log.Printf("[%s]Failed to get network interfaces: %v", GetFuncInfo(), err)
				printErrIface = false
			}
			return false
		}

		for i := range tmpInterfaces {
			interfaces = append(interfaces, &tmpInterfaces[i])
		}
	}

	for _, iface := range interfaces {
		//log.Printf("[%s]  MTU[%d]  flag[%d]  HW[%+v] Index[%d]", iface.Name, iface.MTU, iface.Flags, iface.HardwareAddr, iface.Index)
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := GetValidAddrs(iface)
		if err != nil {
			if printErrIp {
				log.Printf("Err: Failed to get addresses for interface %s: %v", iface.Name, err)
				printErrIp = false
			}
			continue
		}

		//TODO: Multiple network cards this check is invalid, refine this flow later
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil || ipNet.IP.To16() != nil {
					//log.Printf("ipNet.IP:[%s] ", ipNet.IP.String())
					printErrIface = true
					printErrIp = true
					return true
				}
			}
		}

	}

	return false
}

func GetNetworkIP(forceInterfaces []string) ([]string, bool) {
	ipStrList := make([]string, 0)
	ipStrList = append(ipStrList, DefaultIp)

	var interfaces []*net.Interface = make([]*net.Interface, 0)
	if len(forceInterfaces) > 0 {
		for _, ifaceName := range forceInterfaces {
			iface, err := GetValidInterface(ifaceName)
			if err != nil {
				continue
			}
			interfaces = append(interfaces, iface)
		}
	} else {
		tmpInterfaces, err := net.Interfaces()
		if err != nil {
			log.Printf("[%s]Failed to get network interfaces: %v", GetFuncInfo(), err)
			return ipStrList, false
		}

		for i := range tmpInterfaces {
			interfaces = append(interfaces, &tmpInterfaces[i])
		}
	}

	bCheckOk := false
	for _, iface := range interfaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, err := GetValidAddrs(iface)
		if err != nil {
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
