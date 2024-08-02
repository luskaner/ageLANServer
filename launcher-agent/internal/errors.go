package internal

import launcherCommon "github.com/luskaner/aoe2DELanServer/launcherCommon"

const (
	ErrGameTimeoutStart = iota + launcherCommon.ErrLast
	ErrBattleServerTimeOutStart
	ErrFailedStopServer
	ErrFailedWaitForProcess
)
