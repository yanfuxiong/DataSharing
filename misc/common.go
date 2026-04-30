package misc

type C2SMsgType string

const (
	C2SMsg_INIT_CLIENT            C2SMsgType = "INIT_CLIENT"
	C2SMsg_RESET_CLIENT           C2SMsgType = "RESET_CLIENT"   // Deprecated: unused
	C2SMsg_AUTH_INDEX_MOBILE      C2SMsgType = "AUTH_VIA_INDEX" // Deprecated: unused
	C2SMsg_AUTH_DATA_INDEX_MOBILE C2SMsgType = "AUTH_VIA_DATA_INDEX"
	C2SMsg_REQ_CLIENT_LIST        C2SMsgType = "REQ_CLIENT_LIST"
	C2SMsg_CLIENT_HEARTBEAT       C2SMsgType = "CLIENT_HEARTBEAT"
	C2SMsg_DRAG_FILE_START        C2SMsgType = "DRAG_FILE_START"
	C2SMsg_DRAG_FILE_END          C2SMsgType = "REQ_CLIENT_DRAG_FILE"
	CS2Msg_PERIODIC_NOTIFY        C2SMsgType = "RECONN_CLIENT_LIST"
	CS2Msg_NOTIFY_CLIENT_VERSION  C2SMsgType = "NOTIFY_CLIENT_VERSION"
	CS2Msg_MESSAGE_EVENT          C2SMsgType = "MESSAGE_EVENT"
	CS2Msg_UPDATE_SRCPORT_INFO    C2SMsgType = "UPDATE_SRCPORT_INFO"
	CS2Msg_UPDATE_PLUG_EVENT      C2SMsgType = "UPDATE_PLUG_EVENT"
)

type ScenarioType int

const (
	ScenarioType_SingleView ScenarioType = iota + 0
	ScenarioType_ViewManager
	ScenarioType_MultiWindow
	ScenarioType_SingleView_PIP
)

type PlatformMsgEventReq struct {
	Event uint32
	Arg1  string
	Arg2  string
	Arg3  string
	Arg4  string
}

type ClientInfo struct {
	ID             string
	IpAddr         string
	Platform       string
	DeviceName     string
	SourcePortType string
	Version        string
}

type InitClientMessageReq struct {
	HOST          string
	ClientID      string
	Platform      string
	DeviceName    string
	IPAddr        string
	ClientVersion string
	AppStoreLink  string
}

type PlatformMsgEventResponse struct {
	Response
}

type InitClientMessageResponse struct {
	Response
	ClientIndex   uint32
	ClientVersion string
	Scenario      ScenarioType
	IsSupportFileDrag bool
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
	SourcePort
	AuthStatus bool
}

type NotifyClientVersionReq struct {
	ClientVersion string
}

type UpdateClientSrcPortInfoReq struct {
	ClientIndex        int
	SourcePortInfoList []SourcePortInfo
}

type UpdateClientSrcPortInfoResponse struct {
	SourcePortList []SourcePort
	Response
}

type DragFileStartResponse struct {
	Response
	SourcePort
}

type UpdatePlugEventReq struct {
	PlugEvent   bool
	ProductName string
}

type ReconnDirection int

const (
	RECONN_GREATER ReconnDirection = 0
	RECONN_LESS    ReconnDirection = 1
)

type PeriodicNotifyReq struct {
	ClientList    []ClientInfo
	ConnDirect    ReconnDirection
	ClientVersion string
	Scenario      ScenarioType
}

type C2SMessage struct {
	ClientID    string
	ClientIndex uint32
	MsgType     C2SMsgType
	TimeStamp   int64
	ExtData     interface{} //InitClientMessageReq InitClientMessageResponse GetClientListResponse ResetClientResponse ReconnClientListReq AuthDataIndexMobileReq
	// NotifyClientVersionReq PlatformMsgEventReq UpdateClientSrcPortInfoReq UpdateClientSrcPortInfoResponse    UpdatePlugEventReq
}

type SourcePort struct {
	Source int
	Port   int
}

type SourcePortInfo struct {
	SourcePort
	UdpMousePort    int
	UdpKeyboardPort int
}

type DragFileStartInfo struct {
	SourcePort
	HorzSize int
	VertSize int
	PosX     int
	PosY     int
}

type AuthDataInfo struct {
	Width       int
	Height      int
	Framerate   int
	Type        int // 0:Miracast, 1:USBC
	DisplayName string
	CenterX     uint32
	CenterY     uint32
}

const (
	DisplayModeMiracast = 0
	DisplayModeUsbC     = 1
)
