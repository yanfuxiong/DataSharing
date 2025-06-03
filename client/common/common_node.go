package common

type IPAddrInfo struct {
	LocalPort  string
	PublicPort string
	PublicIP   string
}

type NodeInfo struct {
	IPAddr         IPAddrInfo
	ClientIndex    uint32
	ID             string
	DeviceName     string
	Platform       string
	SourcePortType string
}
