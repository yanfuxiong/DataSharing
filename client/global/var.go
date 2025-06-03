package global

import (
	rtkCommon "rtk-cross-share/client/common"
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"time"
)

var NodeInfo = rtkCommon.NodeInfo{
	IPAddr: rtkCommon.IPAddrInfo{
		LocalPort:  "",
		PublicPort: "",
		PublicIP:   "",
	},
	ID:             "",
	DeviceName:     "",
	Platform:       "",
	SourcePortType: "",
}

var (
	RelayServerID     = "QmT4ZCzr1Jhnk2B81fgSsuu9t2qnexo8oP5b1m5eUcSxrg"
	RelayServerIPInfo = "/ip4/180.218.164.96/tcp/8878/p2p/"
	ListenHost        string
	ListenPort        int
	GuestList         []string
	ClientInfoMap     = make(map[string]rtkMisc.ClientInfo)
	ClientListRWMutex = sync.RWMutex{}

	RTT map[string]time.Duration = make(map[string]time.Duration)
)
