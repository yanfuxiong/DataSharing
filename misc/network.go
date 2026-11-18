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

func IsNetworkConnected() bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		if printErrIface {
			log.Printf("[%s]Failed to get network interfaces: %v", GetFuncInfo(), err)
			printErrIface = false
		}
		return false
	}

	for _, iface := range interfaces {
		//log.Printf("[%s]  MTU[%d]  flag[%d]  HW[%+v] Index[%d]", iface.Name, iface.MTU, iface.Flags, iface.HardwareAddr, iface.Index)
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := GetValidAddrs(&iface)
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

// isIPActiveByUDP 通过 UDP Dial 验证 IP 是否属于本机
// 原理：Dial 一个不可达地址，从本地地址推断出用于该子网的本机 IP
func isIPActiveByUDP(ip net.IP) bool {
	// 构造一个目标地址，让内核选择路由
	var target string
	if ip.To4() != nil {
		target = ip.String() + ":9" // discard 端口
	} else {
		target = "[" + ip.String() + "]:9"
	}

	// 使用 UDP 不需要真实可达，只需内核做路由选择
	conn, err := net.Dial("udp", target)
	if err != nil {
		log.Printf("[isIPActiveByUDP] dial failed: %v", err)
		return false
	}
	defer conn.Close()

	// 获取本地地址
	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || localAddr == nil {
		return false
	}

	// 如果本地地址等于待验证 IP，说明是本机地址
	if localAddr.IP.Equal(ip) {
		log.Printf("[isIPActiveByUDP] UDP fallback matched: %s", ip.String())
		return true
	}

	log.Printf("[isIPActiveByUDP] local %s != target %s", localAddr.IP, ip)
	return false
}

func IsIPActive(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("[IsIPActive] net.Interfaces() failed: %v, fallback to UDP check", err)
		return isIPActiveByUDP(ip)
	}

	for _, iface := range interfaces {
		// Android 上不要严格过滤 FlagUp，某些有效接口标志位不完整
		// 只跳过明确是 loopback 的
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("[%s] iface %s Addrs() err: %v", GetFuncInfo(), iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			if netIP, _, _ := net.ParseCIDR(addr.String()); netIP.Equal(ip) {
				log.Printf("[%s] matched on interface %s: %s", GetFuncInfo(), iface.Name, ipStr)
				return true
			}
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.Equal(ip) {
					log.Printf("[%s] matched on interface %s: %s", GetFuncInfo(), iface.Name, ipStr)
					return true
				}
			case *net.IPAddr:
				if v.IP.Equal(ip) {
					log.Printf("[%s] matched on interface %s: %s", GetFuncInfo(), iface.Name, ipStr)
					return true
				}
			}
		}
	}
	return false
}
