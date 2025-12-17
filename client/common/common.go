package common

type FileDropCmd string

const (
	FILE_DROP_REQUEST FileDropCmd = "FILE_DROP_REQUEST"
	FILE_DROP_ACCEPT  FileDropCmd = "FILE_DROP_ACCEPT"
	FILE_DROP_REJECT  FileDropCmd = "FILE_DROP_REJECT"
)

type FilesDataRequestInfo struct {
	TimeStamp int64
	Id        string
	Ip        string
	PathList  []string
}

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

// Only new fields can be added, old fields cannot be modified
type RegistMdnsMessage struct {
	Host            string
	Id              string
	Platform        string
	DeviceName      string
	SourcePortType  string
	Version         string
	FileTransNodeID string
	UdpPort         string
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

type SendFilesRequestErrCode int

const (
	SendFilesRequestSuccess SendFilesRequestErrCode = iota + 1
	SendFilesRequestParameterErr
	SendFilesRequestInProgressBySrc
	SendFilesRequestInProgressByDst
	SendFilesRequestCallbackNotSet
	SendFilesRequestLengthOverRange
	SendFilesRequestSizeOverRange
	SendFilesRequestCacheOverRange
)

type CancelBusinessSource int

const (
	SourceCablePlugIn CancelBusinessSource = iota + 1000 //businessStart
	SourceCablePlugOut
	SourceNetworkSwitch
	SourceVerInvalid
)

const (
	UpperLevelBusinessCancel CancelBusinessSource = iota + 2000 // StartProcessForPeer
	LanServerBusinessCancel
	OldP2PBusinessCancel
	TcpNetworkCancel
	PeerDisconnectCancel
)

const (
	FileTransDone CancelBusinessSource = iota + 3000 // dealFilesCacheDataProcess
	FileTransSrcCancel
	FileTransSrcGuiCancel
	FileTransDstCancel
	FileTransDstGuiCancel
)
