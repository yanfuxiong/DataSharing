package misc

const (
	DefaultIp       = "0.0.0.0"
	LoopBackIp      = "127.0.0.1"
	LanServerName   = "GoZeroconfLanServer" // TODO: DIAS mac address
	LanServiceType  = "_rtkcs._tcp"
	LanServerDomain = "local"
	LanServerPort   = 42424

	LanServiceTypeForServer = "_rtkcsser._tcp" // Service for other server

	ClientHeartbeatInterval = 10 // second

	TextRecordKeyIp          = "ip"
	TextRecordKeyProductName = "productName"
	TextRecordKeyMonitorName = "mName"
	TextRecordKeyTimestamp   = "timestamp"
	TextRecordKeyVersion     = "ver"

	PlatformAndroid = "android"
	PlatformWindows = "windows"
	PlatformMac     = "macOs"
	PlatformiOS     = "iOS"
)
