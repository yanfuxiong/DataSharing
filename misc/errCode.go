package misc

// TODO: client and lanServer ERR CODE  are Define here

type CrossShareErr int

const (
	SUCCESS CrossShareErr = iota + 1
	ERR_TIMEOUT
	ERR_NETWORK
	ERR_CANCEL
	ERR_INVALID
	ERR_DB
	ERR_RESET
	ERR_OTHER
)
