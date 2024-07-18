package internal

import launcherCommon "launcher-common"

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
