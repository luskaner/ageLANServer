package internal

import launcherCommon "launcherCommon"

const (
	ErrGameTimeoutStart = iota + launcherCommon.ErrLast
	ErrBattleServerTimeOutStart
	ErrFailedStopServer
	ErrFailedWaitForProcess
)
