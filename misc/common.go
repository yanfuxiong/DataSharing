package misc

type C2SMsgType string

const (
	C2SMsg_INIT_CLIENT      C2SMsgType = "INIT_CLIENT"
	C2SMsg_RESET_CLIENT     C2SMsgType = "RESET_CLIENT"
	C2SMsg_REQ_CLIENT_LIST  C2SMsgType = "REQ_CLIENT_LIST"
	C2SMsg_CLIENT_HEARTBEAT C2SMsgType = "CLIENT_HEARTBEAT"
)

type ClientInfo struct {
	ID         string
	IpAddr     string
	Platform   string
	DeviceName string
}

type InitClientMessageReq struct {
	HOST       string
	ClientID   string
	Platform   string
	DeviceName string
	IPAddr     string
}

type InitClientMessageResponse struct {
	Code        CrossShareErr
	Msg         string
	ClientIndex int
}

type ResetClientResponse struct {
	Code CrossShareErr
	Msg  string
}

type GetClientListResponse struct {
	Code       CrossShareErr
	Msg        string
	ClientList []ClientInfo
}

type C2SMessage struct {
	ClientID    string
	ClientIndex int
	MsgType     C2SMsgType
	TimeStamp   int64
	ExtData     interface{} //InitClientMessageReq  InitClientMessageResponse  GetClientListResponse  ResetClientResponse
}
