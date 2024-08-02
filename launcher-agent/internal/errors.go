package internal

import launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"

const (
	ErrGameTimeoutStart = iota + launcherCommon.ErrLast
	ErrBattleServerTimeOutStart
	ErrFailedStopServer
	ErrFailedWaitForProcess
)
