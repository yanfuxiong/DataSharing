package misc

type C2SMsgType string

const (
	C2SMsg_INIT_CLIENT            C2SMsgType = "INIT_CLIENT"
	C2SMsg_RESET_CLIENT           C2SMsgType = "RESET_CLIENT"   // Deprecated: unused
	C2SMsg_AUTH_INDEX_MOBILE      C2SMsgType = "AUTH_VIA_INDEX" // Deprecated: unused
	C2SMsg_AUTH_DATA_INDEX_MOBILE C2SMsgType = "AUTH_VIA_DATA_INDEX"
	C2SMsg_REQ_CLIENT_LIST        C2SMsgType = "REQ_CLIENT_LIST"
	C2SMsg_CLIENT_HEARTBEAT       C2SMsgType = "CLIENT_HEARTBEAT"
	C2SMsg_REQ_CLIENT_DRAG_FILE   C2SMsgType = "REQ_CLIENT_DRAG_FILE"
	CS2Msg_RECONN_CLIENT_LIST     C2SMsgType = "RECONN_CLIENT_LIST"
	CS2Msg_NOTIFY_CLIENT_VERSION  C2SMsgType = "NOTIFY_CLIENT_VERSION"
)

type ClientInfo struct {
	ID             string
	IpAddr         string
	Platform       string
	DeviceName     string
	SourcePortType string
}

type InitClientMessageReq struct {
	HOST          string
	ClientID      string
	Platform      string
	DeviceName    string
	IPAddr        string
	ClientVersion string
}

type InitClientMessageResponse struct {
	Response
	ClientIndex   uint32
	ClientVersion string
}

type ResetClientResponse struct {
	Response
}

type GetClientListResponse struct {
	Response
	ClientList []ClientInfo
}

type AuthIndexMobileReq struct {
	SourceAndPort SourcePort
}

type AuthDataIndexMobileReq struct {
	AuthData AuthDataInfo
}

type AuthDataIndexMobileResponse struct {
	Response
	AuthStatus bool
}

type AuthIndexMobileResponse struct {
	Response
	AuthStatus bool
}

type NotifyClientVersionReq struct {
	ClientVersion string
}

type ReconnDirection int

const (
	RECONN_GREATER ReconnDirection = 0
	RECONN_LESS    ReconnDirection = 1
)

type ReconnClientListReq struct {
	ClientList    []ClientInfo
	ConnDirect    ReconnDirection
	ClientVersion string
}

type C2SMessage struct {
	ClientID    string
	ClientIndex uint32
	MsgType     C2SMsgType
	TimeStamp   int64
	ExtData     interface{} //InitClientMessageReq InitClientMessageResponse GetClientListResponse ResetClientResponse ReconnClientListReq AuthDataIndexMobileReq NotifyClientVersionReq
}

type SourcePort struct {
	Source int
	Port   int
}

type AuthDataInfo struct {
	Width       int
	Height      int
	Framerate   int
	Type        int // 0:Miracast, 1:USBC
	DisplayName string
}

const (
	DisplayModeMiracast = 0
	DisplayModeUsbC     = 1
)
