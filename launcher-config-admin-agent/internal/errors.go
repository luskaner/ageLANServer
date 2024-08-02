package internal

import launcherCommon "github.com/luskaner/aoe2DELanServer/launcherCommon"

const (
	ErrCreatePipe = iota + launcherCommon.ErrLast
	ErrDecode
	ErrNonExistingAction
	ErrConnectionClosing
	ErrIpsInvalid
	ErrCertAlreadyAdded
	ErrIpsAlreadyMapped
	ErrCertInvalid
)
