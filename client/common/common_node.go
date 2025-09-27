package common

import rtkMisc "rtk-cross-share/misc"

type IPAddrInfo struct {
	LocalPort  string
	PublicPort string
	UpdPort    string
	PublicIP   string
}

type NodeInfo struct {
	IPAddr          IPAddrInfo
	ClientIndex     uint32
	ID              string
	DeviceName      string
	Platform        string
	SourcePortType  string
	FileTransNodeID string
}

type ClientStatusInfo struct {
	TimeStamp int64
	Status    uint8 //1: online; 0:offline
	rtkMisc.ClientInfo
}

type ClientListInfo struct {
	TimeStamp  int64
	ID         string // self node ID
	IpAddr     string // self node IpAddr
	ClientList []rtkMisc.ClientInfo
}

type ClientInfoEx struct {
	rtkMisc.ClientInfo
	IsSupportXClip  bool
	FileTransNodeID string
	UpdPort         string
}
