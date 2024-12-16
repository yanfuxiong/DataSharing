package common

type IPAddrInfo struct {
	LocalPort  string
	PublicPort string
	PublicIP   string
}

type NodeInfo struct {
	IPAddr     IPAddrInfo
	ID         string
	DeviceName string
}

type ClientInfo struct {
	ID         string
	IpAddr     string
	Platform   string
	DeviceName string
}
