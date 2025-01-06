package common

type SocketErr int

const (
	OK SocketErr = iota + 1
	ERR_TIMEOUT
	ERR_NETWORK
	ERR_JSON
	ERR_CANCEL
	ERR_CONNECTION
	ERR_EOF
	ERR_OTHER
)

type EventType int

const (
	EVENT_TYPE_OPEN_FILE_ERR = 0
	EVENT_TYPE_RECV_TIMEOUT  = 1
)
