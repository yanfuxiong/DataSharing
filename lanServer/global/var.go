package global

import rtkMisc "rtk-cross-share/misc"

var (
	ServerIPAddr      string
	ServerPort        int
	ServerMdnsId      string = ""
	ServerMonitorName string = "Unknown"
	ServerProductName string = ""
	Scenario          rtkMisc.ScenarioType
	Capability        int
)
