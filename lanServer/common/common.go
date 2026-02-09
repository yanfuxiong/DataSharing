package common

import (
	"log"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"
)

// sqlite table struct
type ClientInfoTb struct {
	Index           int
	ClientId        string
	Host            string
	IpAddr          string
	Source          int
	Port            int
	UdpMousePort    int
	UdpKeyboardPort int
	DeviceName      string
	Platform        string
	Version         string
	Online          bool
	AuthStatus      bool
	UpdateTime      string
	CreateTime      string
	LastAuthTime    string
}

func (c *ClientInfoTb) Dump() {
	log.Printf("[ClientInfoTb] Index:%d, ClientId:%s, Host:%s, IpAddr:%s, DeviceName:%s, Platform:%s", c.Index, c.ClientId, c.Host, c.IpAddr, c.DeviceName, c.Platform)
	log.Printf("[ClientInfoTb] Source:%d, Port:%d, UdpMousePort:%d, UdpKybrdPort:%d, Version:%s", c.Source, c.Port, c.UdpMousePort, c.UdpKeyboardPort, c.Version)
	log.Printf("[ClientInfoTb] Online:%t, AuthStatus:%t, UpdateTime:%s, CreateTime:%s", c.Online, c.AuthStatus, c.UpdateTime, c.CreateTime)
	log.Println()
}

type TimingData struct {
	Source      int
	Port        int
	Width       int
	Height      int
	Framerate   int
	DisplayMode int
	DisplayName string
	DeviceName  string
}

func (t *TimingData) Dump() {
	log.Printf("[TimingData] Source:%d, Port:%d, Width:%d, Height:%d", t.Source, t.Port, t.Width, t.Height)
	log.Printf("[TimingData] Framerate:%d, Type:%d, DisplayName:%s", t.Framerate, t.DisplayMode, t.DisplayName)
	log.Println()
}

type SourcePortType string

const (
	SrcPortType_UNKNOWN  SourcePortType = "UNKNOWN_TYPE"
	SrcPortType_HDMI_1   SourcePortType = "HDMI1"
	SrcPortType_HDMI_2   SourcePortType = "HDMI2"
	SrcPortType_USBC_1   SourcePortType = "USBC1"
	SrcPortType_USBC_2   SourcePortType = "USBC2"
	SrcPortType_DP_1     SourcePortType = "DP1"
	SrcPortType_DP_2     SourcePortType = "DP2"
	SrcPortType_MIRACAST SourcePortType = "Miracast"
)

var (
	MAX_PORT_HDMI = 2 // Now only support 2 port in HDMI
	MAX_PORT_DP   = 2 // Now only support 2 port in DP
	DpSrcTypeAry  = make([]SourcePortType, MAX_PORT_DP)
)

func IsSourceTypeUsbC(port int) bool {
	return DpSrcTypeAry[port] == SrcPortType_USBC_1 || DpSrcTypeAry[port] == SrcPortType_USBC_2
}

func GetUsbCPort() []int {
	usbcPortList := make([]int, 0)
	for port := range MAX_PORT_DP {
		if IsSourceTypeUsbC(port) {
			usbcPortList = append(usbcPortList, port)
		}
	}
	return usbcPortList
}

// TODO: This mapping is hardcode now. Need to consider the different PCB in the future
func GetClientSourcePortType(src, port int) string {
	srcPortType := SrcPortType_UNKNOWN
	switch src {
	case rtkGlobal.Src_HDMI:
		if port >= MAX_PORT_HDMI {
			log.Printf("[%s] Invalid port: %d", rtkMisc.GetFuncInfo(), port)
		} else if port == 0 {
			srcPortType = SrcPortType_HDMI_1
		} else if port == 1 {
			srcPortType = SrcPortType_HDMI_2
		}

	case rtkGlobal.Src_DP:
		if port >= MAX_PORT_DP {
			log.Printf("[%s] Invalid port: %d", rtkMisc.GetFuncInfo(), port)
		} else {
			srcPortType = DpSrcTypeAry[port]
		}

	case rtkGlobal.Src_STREAM:
		if port == rtkGlobal.Port_subType_Miracast {
			srcPortType = SrcPortType_MIRACAST
		}

	default:
		log.Printf("[%s] source:[%d] port:[%d] unknown type!", rtkMisc.GetFuncInfo(), src, port)
	}

	return string(srcPortType)
}

type SrcPortTiming struct {
	Source    int
	Port      int
	Width     int
	Height    int
	Framerate int
}

func (s *SrcPortTiming) IsSingal() bool {
	return s.Width > 0 && s.Height > 0 && s.Framerate > 0
}
