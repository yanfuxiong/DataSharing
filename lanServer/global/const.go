package global

const (
	LanServerVersion  = "2.2.18" // it must notify client and update client version and VersionReadme.txt  when intermediate version is update
	ClientBaseVersion = "2.3"

	LOG_PATH         = "/data/vendor/realtek/cross_share/"
	SOCKET_PATH_ROOT = "/mnt/vendor/tvdata/database/cross_share/"
	DB_PATH          = "/mnt/vendor/tvdata/database/cross_share/"
	DB_NAME          = "cross_share.db"

	Src_HDMI              = 8
	Src_DP                = 13
	Src_STREAM            = 12
	Port_max              = 4
	Port_subType_Miracast = 9
)

type DpSrcType int

const (
	DP_SRC_TYPE_NONE    DpSrcType = 0
	DP_SRC_TYPE_DP      DpSrcType = 1
	DP_SRC_TYPE_MINI_DP DpSrcType = 2
	DP_SRC_TYPE_USBC    DpSrcType = 3
)


type ClientEventMsgType int

const (
	CLIENT_EVENT_MSG_RESERVED   ClientEventMsgType = 0
	CLIENT_EVENT_MSG_SHOW_TOAST ClientEventMsgType = 1
	CLIENT_EVENT_MSG_OPEN_GUIDE ClientEventMsgType = 2
)