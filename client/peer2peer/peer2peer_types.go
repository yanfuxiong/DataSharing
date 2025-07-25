package peer2peer

import (
	rtkCommon "rtk-cross-share/client/common"
)

type Peer2PeerMessage struct {
	SourceID       string
	SourcePlatform string
	FmtType        rtkCommon.TransFmtType
	State          StateType
	Command        CommandType
	TimeStamp      uint64
	ExtData        interface{}
}

type StateType string

const (
	STATE_INIT  StateType = "STATE_INIT"
	STATE_INFO  StateType = "STATE_INFO"
	STATE_TRANS StateType = "STATE_TRANS"
	STATE_IO    StateType = "STATE_IO"
)

type CommandType string

const (
	COMM_INIT                        CommandType = "COMM_INIT"
	COMM_SRC                         CommandType = "COMM_SRC"
	COMM_DST                         CommandType = "COMM_DST"
	COMM_CANCEL_SRC                  CommandType = "COMM_CANCEL_SRC"
	COMM_CANCEL_DST                  CommandType = "COMM_CANCEL_DST"
	COMM_DISCONNECT                  CommandType = "COMM_DISCONNECT"
	COMM_CB_TRANSFER_SRC_INTERRUPT   CommandType = "COMM_CB_TRANSFER_SRC_INTERRUPT"
	COMM_CB_TRANSFER_DST_INTERRUPT   CommandType = "COMM_CB_TRANSFER_DST_INTERRUPT"
	COMM_FILE_TRANSFER_SRC_INTERRUPT CommandType = "COMM_FILE_TRANSFER_SRC_INTERRUPT"
	COMM_FILE_TRANSFER_DST_INTERRUPT CommandType = "COMM_FILE_TRANSFER_DST_INTERRUPT"
)

type DispatchCmd struct {
	FmtType rtkCommon.TransFmtType
	State   StateType
	Command CommandType
}

type EventResult struct {
	Cmd  DispatchCmd
	Data interface{}
}
