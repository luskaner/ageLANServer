package internal

import launcherCommon "launcher-common"

const (
	ErrTimeoutStart = iota + launcherCommon.ErrLast
	ErrParseServerPid
	ErrFailedStopServer
	ErrFailedWaitForProcess
)
