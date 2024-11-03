package internal

import launcherCommon "github.com/luskaner/ageLANServer/launcher-common"

const (
	ErrListen = iota + launcherCommon.ErrLast
	ErrDecode
	ErrNonExistingAction
	ErrConnectionClosing
	ErrIpsInvalid
	ErrCertAlreadyAdded
	ErrIpsAlreadyMapped
	ErrCertInvalid
	ErrCDNAlreadyMapped
)
