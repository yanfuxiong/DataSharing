package common

type FileDropCmd string

const (
	FILE_DROP_REQUEST FileDropCmd = "FILE_DROP_REQUEST"
	FILE_DROP_ACCEPT  FileDropCmd = "FILE_DROP_ACCEPT"
	FILE_DROP_REJECT  FileDropCmd = "FILE_DROP_REJECT"
)

type ConnectMessage struct {
	Tag           string
	ObservedAddrs string
}

type SyncMessage struct {
	Tag string
}

type RegMessage struct {
	HOST  string
	GUEST string
}

type RegistMdnsMessage struct {
	Host     string
	Id       string
	Platform string
}

type RegResponseMessage struct {
	GUEST_LIST            []string
	GUEST_PUBLIC_TCP_IP   string
	GUEST_PUBLIC_TCP_PORT string
}

const (
	P2P_EVENT_SERVER_CONNEDTED    = 0
	P2P_EVENT_SERVER_CONNECT_FAIL = 1
	P2P_EVENT_CLIENT_CONNEDTED    = 2
	P2P_EVENT_CLIENT_CONNECT_FAIL = 3
)
