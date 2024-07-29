package internal

import launcherCommon "launcher-common"

const (
	ErrTimeoutStart = iota + launcherCommon.ErrLast
	ErrFailedStopServer
	ErrFailedWaitForProcess
)
