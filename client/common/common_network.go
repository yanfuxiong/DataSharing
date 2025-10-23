package common

import "time"

const (
	PingInterval     = 3500 * time.Millisecond
	PingTimeout      = 3000 * time.Millisecond
	PingTimeoutMilli = 3000
	PingErrMaxCnt    = 3
)
