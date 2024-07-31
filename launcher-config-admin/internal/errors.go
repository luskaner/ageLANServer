package internal

import (
	launcherCommon "launcherCommon"
)

const (
	ErrLocalCertRemove = iota + launcherCommon.ErrLast
	ErrIpMapRemove
	ErrIpMapRemoveRevert
	ErrLocalCertAdd
	ErrLocalCertAddParse
	ErrIpMapAdd
	ErrIpMapAddRevert
)
