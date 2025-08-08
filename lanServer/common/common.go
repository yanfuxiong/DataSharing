package common

import (
	"log"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"
)

// sqlite table struct
type ClientInfoTb struct {
	Index      int
	ClientId   string
	Host       string
	IpAddr     string
	Source     int
	Port       int
	DeviceName string
	Platform   string
	Version    string
	Online     bool
	AuthStatus bool
	UpdateTime string
	CreateTime string
}

func (c *ClientInfoTb) Dump() {
	log.Printf("[ClientInfoTb] Index:%d, ClientId:%s, Host:%s, IpAddr:%s", c.Index, c.ClientId, c.Host, c.IpAddr)
	log.Printf("[ClientInfoTb] Source:%d, Port:%d, DeviceName:%s, Platform:%s, Version:%s", c.Source, c.Port, c.DeviceName, c.Platform, c.Version)
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
	SrcPortType_MIRACAST SourcePortType = "Miracast"
)

// TODO: This mapping is hardcode now. Need to consider the different PCB in the future
func GetClientSourcePortType(src, port int) string {
	srcPortType := SrcPortType_UNKNOWN
	if src == rtkGlobal.Src_HDMI && port == 0 {
		srcPortType = SrcPortType_HDMI_1
	} else if src == rtkGlobal.Src_HDMI && port == 1 {
		srcPortType = SrcPortType_HDMI_2
	} else if src == rtkGlobal.Src_DP && port == 0 {
		srcPortType = SrcPortType_USBC_1
	} else if src == rtkGlobal.Src_DP && port == 1 {
		srcPortType = SrcPortType_USBC_2
	} else if src == rtkGlobal.Src_STREAM && port == rtkGlobal.Port_subType_Miracast {
		srcPortType = SrcPortType_MIRACAST
	} else {
		log.Printf("[%s] source:[%d] port:[%d] unknown type!", rtkMisc.GetFuncInfo(), src, port)
	}

	return string(srcPortType)
}
