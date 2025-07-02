package utils

import (
	"errors"
	"log"
	"net"
)

var g_addrs []string

func GetAddrsFromPlatform(addrsList []string) {
	g_addrs = make([]string, 0)
	for _, addr := range addrsList {
		if addr == "" {
			continue
		}
		g_addrs = append(g_addrs, addr)
		log.Printf("GetAddrsFromPlatform get addr: %+v", addr)
	}
	log.Printf("xyfGetAddrsFromPlatform get size: %+v\n", len(g_addrs))
}

func InterfaceFromJavaAddrs() ([]net.Addr, error) {
	/* addrArray is an addr array, like this:
	    netList := make([]string, 6)
	   netList[0] = "fe80::2a52:e7ad:75d9:3150/64"
	   netList[1] = "169.254.127.187/16"
	   netList[2] = "fe80::92b9:c4c:defl:459c/64"
	   netList[3] = "10.6.192.162/23"
	   netList[4] = "::1/128"
	   netList[5] = "127.0.0.1/8"
	   or netList = append(netList,"127.0.0.1/8")*/
	if len(g_addrs) <= 0 {
		return nil, errors.New("InterfaceFromJavaAddrs g_addrs is null")
	}

	addrs := make([]net.Addr, len(g_addrs))
	for i, val := range g_addrs {
		if val == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(val)
		if err != nil {
			log.Printf("%+v - %+v Error parsing CIDR: %v\n", i, val, err)
			return nil, err
		} else {
			addrs[i] = ipNet
		}
	}
	/*for i, addr := range addrs {
		log.Printf("InterfaceFromJavaAddrs %+v - %v : %v", i, addr.Network(), addr.String())
	}*/
	return addrs, nil
}

type Interface struct {
	Index        int
	MTU          int
	Name         string
	HardwareAddr []byte
	Flags        uint
}

var iFaces []net.Interface

func SetNetInterfaces(name string, index int) {
	iFaces = make([]net.Interface, 0)

	/*hw, err := net.ParseMAC(mac)
	if err != nil {
		log.Printf("ParseMAC [%s] error:%+v", err)
	}
	log.Printf("mac data[%+v]", hw)*/
	//hwMac, _ := net.ParseMAC("9a:84:2b:39:4c:f2")
	iFaces = append(iFaces, net.Interface{
		Index:        index,
		MTU:          1500,
		Name:         name,
		HardwareAddr: nil,
		Flags:        51,
	})
}

func GetNetInterfaces() []net.Interface {
	return iFaces
}
