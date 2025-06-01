package common

import (
	"log"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"
)

type SourcePortType string

const (
	SrcPortType_UNKNOWN  SourcePortType = "UNKNOWN_TYPE"
	SrcPortType_HDMI_1   SourcePortType = "HDMI1"
	SrcPortType_HDMI_2   SourcePortType = "HDMI2"
	SrcPortType_USBC_1   SourcePortType = "USBC1"
	SrcPortType_USBC_2   SourcePortType = "USBC2"
	SrcPortType_MIRACAST SourcePortType = "Miracast"
)

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
