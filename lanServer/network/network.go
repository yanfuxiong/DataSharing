package network

import (
	"errors"
	"log"
	"net"
	rtkMisc "rtk-cross-share/misc"
	"sort"
	"strconv"
	"strings"
)

// get interfaces priority order from high to low : eth0, eth1, eth2, eth3, eth4...., wlan0,wlan1,wlan2,wlan3,wlan4....
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
		return isIfaceNameSuffixBigger("wlan", ifaceNameWlan[i], ifaceNameWlan[j])
	})

	sort.Slice(ifaceNameEth, func(i, j int) bool {
		//priority order: eth0 > eth1 > eth2 > eth4 ...
		return isIfaceNameSuffixBigger("eth", ifaceNameEth[i], ifaceNameEth[j])
	})

	ifaceNameEth = append(ifaceNameEth, ifaceNameWlan...)
	if len(ifaceNameEth) == 0 {
		return nil, errors.New("GetValidInterfaceList is nil!")
	}

	return ifaceNameEth, nil
}

func isIfaceNameSuffixBigger(prefix, ethSmall, ethBig string) bool {
	if !strings.HasPrefix(ethBig, prefix) || !strings.HasPrefix(ethSmall, prefix) {
		return false
	}
	bigNumStr := strings.TrimPrefix(ethBig, prefix)
	smallNumStr := strings.TrimPrefix(ethSmall, prefix)
	if bigNumStr == "" || smallNumStr == "" {
		return false
	}
	bigNum, err := strconv.Atoi(bigNumStr)
	if err != nil {
		return false
	}
	smallNum, err := strconv.Atoi(smallNumStr)
	if err != nil {
		return false
	}

	return bigNum > smallNum
}
