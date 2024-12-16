package global

import (
	rtkCommon "rtk-cross-share/common"
	"sync"
	"time"
)

var NodeInfo = rtkCommon.NodeInfo{
	IPAddr: rtkCommon.IPAddrInfo{
		LocalPort:  "",
		PublicPort: "",
		PublicIP:   "",
	},
	ID:         "",
	DeviceName: "",
}

var (
	RelayServerID     = "QmT4ZCzr1Jhnk2B81fgSsuu9t2qnexo8oP5b1m5eUcSxrg"
	RelayServerIPInfo = "/ip4/180.218.164.96/tcp/8878/p2p/"
	GuestList         []string
	MdnsClientList    []rtkCommon.ClientInfo
	MdnsListRWMutex   = sync.RWMutex{}

	RTT map[string]time.Duration = make(map[string]time.Duration)
)
