package internal

import (
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
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
