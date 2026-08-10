package network

import (
	"log"
	"net"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"
	"sort"
	"strconv"
	"strings"
)

// get interfaces priority order from high to low :br0, eth0, eth1, eth2, eth3, eth4...., wlan0,wlan1,wlan2,wlan3,wlan4....
func GetValidInterfaceList() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("[%s]Failed to get network interfaces: %v", rtkMisc.GetFuncInfo(), err)
		return nil, err
	}

	ifaceNameWlan := make([]string, 0)
	ifaceNameEth := make([]string, 0)
	for _, iface := range interfaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					//log.Printf("[%s] name :%s, IP:%s ", rtkMisc.GetFuncInfo(), iface.Name, ipNet.IP.String())

					if strings.HasPrefix(iface.Name, "wlan") {
						ifaceNameWlan = append(ifaceNameWlan, iface.Name)
					} else if strings.HasPrefix(iface.Name, "eth") {
						ifaceNameEth = append(ifaceNameEth, iface.Name)
					}
				}
			}
		}
	}

	sort.Slice(ifaceNameWlan, func(i, j int) bool {
		//priority order: wlan0 > wlan1 > wlan2 > wlan4 ...
		return isIfaceNameSuffixSmaller("wlan", ifaceNameWlan[i], ifaceNameWlan[j])
	})

	sort.Slice(ifaceNameEth, func(i, j int) bool {
		//priority order: eth0 > eth1 > eth2 > eth4 ...
		return isIfaceNameSuffixSmaller("eth", ifaceNameEth[i], ifaceNameEth[j])
	})

	ifaceNameList := make([]string, 0)
	ifaceNameList = append(ifaceNameList, rtkGlobal.BridgeInterfaceName)
	ifaceNameList = append(ifaceNameList, ifaceNameEth...)
	ifaceNameList = append(ifaceNameList, ifaceNameWlan...)

	return ifaceNameList, nil
}

func isIfaceNameSuffixSmaller(prefix, ifaceI, ifaceJ string) bool {
	if !strings.HasPrefix(ifaceI, prefix) || !strings.HasPrefix(ifaceJ, prefix) {
		return false
	}
	numStrI := strings.TrimPrefix(ifaceI, prefix)
	numStrJ := strings.TrimPrefix(ifaceJ, prefix)
	if numStrI == "" || numStrJ == "" {
		return false
	}
	numI, err := strconv.Atoi(numStrI)
	if err != nil {
		return false
	}
	numJ, err := strconv.Atoi(numStrJ)
	if err != nil {
		return false
	}

	return numI < numJ
}
