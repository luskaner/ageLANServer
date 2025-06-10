package internal

import launcherCommon "github.com/luskaner/ageLANServer/launcher-common"

const (
	ErrListen = iota + launcherCommon.ErrLast
	ErrDecode
	ErrNonExistingAction
	ErrConnectionClosing
	ErrCertAlreadyAdded
	ErrIpAlreadyMapped
	ErrCertInvalid
	ErrCDNAlreadyMapped
)
