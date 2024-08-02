package internal

import launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"

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
