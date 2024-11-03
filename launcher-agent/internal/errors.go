package internal

import launcherCommon "github.com/luskaner/ageLANServer/launcher-common"

const (
	ErrGameTimeoutStart = iota + launcherCommon.ErrLast
	ErrBattleServerTimeOutStart
	ErrFailedStopServer
	ErrFailedWaitForProcess
)
