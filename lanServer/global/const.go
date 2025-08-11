package global

const (
	LanServerVersion  = "2.2.13" // it must notify client and update client version and VersionReadme.txt  when intermediate version is update
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
